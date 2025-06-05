package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/kaas-api/internal/auth"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthService handles authentication business logic
type AuthService struct {
	db           *gorm.DB
	config       *config.Config
	logger       *zap.Logger
	oauthClient  *auth.OAuthClient
	tokenManager *auth.TokenManager
	keyManager   *auth.KeyManager
}

// NewAuthService creates a new authentication service
func NewAuthService(database *gorm.DB, cfg *config.Config, logger *zap.Logger) (*AuthService, error) {
	return NewAuthServiceWithRedis(database, cfg, logger, nil)
}

// NewAuthServiceWithRedis creates a new authentication service with Redis client
func NewAuthServiceWithRedis(database *gorm.DB, cfg *config.Config, logger *zap.Logger, redisClient auth.RedisClient) (*AuthService, error) {
	// Initialize OAuth client
	oauthClient := auth.NewOAuthClient(cfg, redisClient)

	// Initialize key manager
	keyManager, err := auth.NewKeyManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize key manager: %w", err)
	}

	// Initialize token manager
	tokenManager := auth.NewTokenManager(
		keyManager.GetPrivateKey(),
		keyManager.GetPublicKey(),
		cfg.Auth.OIDCIssuer,
		time.Duration(cfg.Auth.JWTExpiration)*time.Second,
	)

	return &AuthService{
		db:           database,
		config:       cfg,
		logger:       logger,
		oauthClient:  oauthClient,
		tokenManager: tokenManager,
		keyManager:   keyManager,
	}, nil
}

// GetAuthURL generates an OAuth authorization URL
func (s *AuthService) GetAuthURL(provider string) (string, string, error) {
	ctx := context.Background()
	state, err := s.oauthClient.GenerateAndStoreState(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	authURL, err := s.oauthClient.GetAuthURL(provider, state)
	if err != nil {
		return "", "", err
	}

	return authURL, state, nil
}

// GetAuthURLWithPKCE generates an OAuth authorization URL with PKCE
func (s *AuthService) GetAuthURLWithPKCE(provider, codeChallenge, challengeMethod string) (string, string, error) {
	ctx := context.Background()
	state, err := s.oauthClient.GenerateAndStoreState(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// For now, use the standard auth URL
	// TODO: Implement PKCE support in OAuth client
	authURL, err := s.oauthClient.GetAuthURL(provider, state)
	if err != nil {
		return "", "", err
	}

	return authURL, state, nil
}

// HandleCallback processes OAuth callback
func (s *AuthService) HandleCallback(ctx context.Context, provider, code, state string) (*db.User, string, error) {
	// Validate and consume state
	if err := s.oauthClient.ValidateAndConsumeState(ctx, state); err != nil {
		return nil, "", fmt.Errorf("invalid state: %w", err)
	}

	// Exchange code for token
	token, err := s.oauthClient.ExchangeCode(ctx, provider, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from provider
	userInfo, err := s.oauthClient.GetUserInfo(ctx, provider, token)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, isNewUser, err := s.findOrCreateUser(userInfo)
	if err != nil {
		return nil, "", fmt.Errorf("failed to process user: %w", err)
	}

	// If new user, create default organization
	if isNewUser {
		if err := s.createDefaultOrganization(user); err != nil {
			s.logger.Error("Failed to create default organization", 
				zap.Error(err),
				zap.String("user_id", user.ID))
			// Don't fail login if org creation fails
		}
	}

	// Get user's organizations
	orgIDs, err := s.getUserOrganizationIDs(user.ID)
	if err != nil {
		s.logger.Error("Failed to get user organizations", 
			zap.Error(err),
			zap.String("user_id", user.ID))
		orgIDs = []string{} // Continue with empty orgs
	}

	// Generate JWT token
	jwtToken, err := s.tokenManager.GenerateToken(user.ID, user.Email, user.DisplayName, orgIDs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, jwtToken, nil
}

// findOrCreateUser finds an existing user or creates a new one
func (s *AuthService) findOrCreateUser(userInfo *auth.UserInfo) (*db.User, bool, error) {
	var user db.User
	isNewUser := false

	// Try to find existing user
	err := s.db.Where("external_id = ? AND provider = ?", userInfo.ID, userInfo.Provider).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = db.User{
			ExternalID:  userInfo.ID,
			Provider:    userInfo.Provider,
			Email:       userInfo.Email,
			DisplayName: userInfo.Name,
		}

		if err := s.db.Create(&user).Error; err != nil {
			return nil, false, fmt.Errorf("failed to create user: %w", err)
		}
		isNewUser = true
	} else if err != nil {
		return nil, false, fmt.Errorf("failed to query user: %w", err)
	} else {
		// Update existing user info
		updates := map[string]interface{}{
			"email":        userInfo.Email,
			"display_name": userInfo.Name,
		}
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			s.logger.Warn("Failed to update user info", 
				zap.Error(err),
				zap.String("user_id", user.ID))
		}
	}

	return &user, isNewUser, nil
}

// createDefaultOrganization creates a default organization for new users
func (s *AuthService) createDefaultOrganization(user *db.User) error {
	// Create organization
	org := db.Organization{
		Name: fmt.Sprintf("%s's Organization", user.DisplayName),
	}

	if err := s.db.Create(&org).Error; err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	// Add user as admin
	orgUser := db.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           "admin",
	}

	if err := s.db.Create(&orgUser).Error; err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}

	s.logger.Info("Created default organization for new user",
		zap.String("user_id", user.ID),
		zap.String("org_id", org.ID))

	return nil
}

// getUserOrganizationIDs gets all organization IDs for a user
func (s *AuthService) getUserOrganizationIDs(userID string) ([]string, error) {
	var orgUsers []db.OrganizationUser
	err := s.db.Where("user_id = ?", userID).Find(&orgUsers).Error
	if err != nil {
		return nil, err
	}

	orgIDs := make([]string, len(orgUsers))
	for i, ou := range orgUsers {
		orgIDs[i] = ou.OrganizationID
	}

	return orgIDs, nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
	return s.tokenManager.ValidateToken(tokenString)
}

// GetJWKS returns the JSON Web Key Set
func (s *AuthService) GetJWKS() ([]byte, error) {
	return s.keyManager.GetJWKSJSON()
}

// GenerateWorkspaceToken generates a token for vCluster access
func (s *AuthService) GenerateWorkspaceToken(userID, workspaceID string) (string, error) {
	// Get user's groups for the workspace
	groups, err := s.getUserWorkspaceGroups(userID, workspaceID)
	if err != nil {
		return "", fmt.Errorf("failed to get user groups: %w", err)
	}

	return s.tokenManager.GenerateWorkspaceToken(userID, workspaceID, groups)
}

// getUserWorkspaceGroups gets all groups (with hierarchy) for a user in a workspace
func (s *AuthService) getUserWorkspaceGroups(userID, workspaceID string) ([]string, error) {
	// TODO: Implement proper group hierarchy resolution
	// For now, return default groups
	return []string{"WorkspaceMembers"}, nil
}

// StoreAuthState stores authentication state data
func (s *AuthService) StoreAuthState(state string, clientIP, userAgent string) error {
	// TODO: Implement state storage with Redis
	// Will store: {"client_ip": clientIP, "user_agent": userAgent}
	return nil
}

// VerifyAuthState verifies and retrieves authentication state data
func (s *AuthService) VerifyAuthState(state string, clientIP string) error {
	// TODO: Implement state verification with Redis
	// Should verify that the state exists and matches the client IP
	return nil
}

// VerifyPKCE verifies PKCE parameters
func (s *AuthService) VerifyPKCE(state, codeVerifier string) error {
	// TODO: Implement PKCE verification
	// Should retrieve the code challenge from the state and verify against the verifier
	return nil
}

// HandleCallbackWithTokenPair processes OAuth callback and returns token pair
func (s *AuthService) HandleCallbackWithTokenPair(ctx context.Context, provider, code, state, clientIP, userAgent string) (*db.User, *auth.TokenPair, error) {
	user, accessToken, err := s.HandleCallback(ctx, provider, code, state)
	if err != nil {
		return nil, nil, err
	}

	// Create token pair (for now, use same token for both)
	// TODO: Implement proper refresh token generation
	tokenPair := &auth.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: accessToken, // Temporary: use same token
		TokenType:    "Bearer",
		ExpiresIn:    int(s.config.Auth.JWTExpiration),
	}

	return user, tokenPair, nil
}

// RefreshTokenPair refreshes an access token using a refresh token
func (s *AuthService) RefreshTokenPair(refreshToken, clientIP, userAgent string) (*auth.TokenPair, error) {
	// Validate the refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new access token
	newToken, err := s.tokenManager.GenerateToken(claims.Subject, claims.Email, claims.Name, claims.OrgIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token: %w", err)
	}

	// Create token pair
	tokenPair := &auth.TokenPair{
		AccessToken:  newToken,
		RefreshToken: refreshToken, // Keep same refresh token
		TokenType:    "Bearer",
		ExpiresIn:    int(s.config.Auth.JWTExpiration),
	}

	return tokenPair, nil
}

// RevokeRefreshToken revokes a refresh token
func (s *AuthService) RevokeRefreshToken(refreshToken string) error {
	// TODO: Implement token revocation with Redis blacklist
	return nil
}

// GetUserSessions retrieves active sessions for a user
func (s *AuthService) GetUserSessions(userID string) ([]map[string]interface{}, error) {
	// TODO: Implement session management
	return []map[string]interface{}{}, nil
}

// LogSecurityEvent logs a security-related event
func (s *AuthService) LogSecurityEvent(userID, eventType, description, clientIP, level string) {
	metadata := map[string]interface{}{
		"client_ip": clientIP,
		"level": level,
	}
	s.logger.Info("Security event",
		zap.String("user_id", userID),
		zap.String("event_type", eventType),
		zap.String("description", description),
		zap.Any("metadata", metadata),
	)
	// TODO: Implement persistent security event logging
}

// RevokeSession revokes a specific session for a user
func (s *AuthService) RevokeSession(userID, sessionID string) error {
	// TODO: Implement session revocation with Redis
	s.logger.Info("Revoking session",
		zap.String("user_id", userID),
		zap.String("session_id", sessionID),
	)
	return nil
}

// RevokeAllSessionsExcept revokes all sessions except the current one
func (s *AuthService) RevokeAllSessionsExcept(userID, currentToken string) error {
	// TODO: Implement bulk session revocation with Redis
	s.logger.Info("Revoking all sessions except current",
		zap.String("user_id", userID),
	)
	return nil
}

// GetSecurityLogs retrieves security logs for a user
func (s *AuthService) GetSecurityLogs(userID string, limit int) ([]map[string]interface{}, error) {
	// TODO: Implement security log retrieval from storage
	s.logger.Info("Getting security logs",
		zap.String("user_id", userID),
		zap.Int("limit", limit),
	)
	return []map[string]interface{}{}, nil
}