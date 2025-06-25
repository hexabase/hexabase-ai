package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	internalAuth "github.com/hexabase/hexabase-ai/api/internal/auth"
	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
)

// InternalAIOpsClaims defines the structure for the internal AIOps token.
type InternalAIOpsClaims struct {
	jwt.RegisteredClaims
	UserID            string   `json:"user_id"`
	OrgIDs            []string `json:"org_ids"`
	ActiveWorkspaceID string   `json:"active_workspace_id"`
	TokenType         string   `json:"token_type"`
}

type service struct {
	repo               domain.Repository
	oauthRepo          domain.OAuthRepository
	keyRepo            domain.KeyRepository
	tokenManager       *internalAuth.TokenManager
	tokenDomainService domain.TokenDomainService
	logger             *slog.Logger
	defaultTokenExpiry int // Default token expiry in seconds when claims don't have expiry info
}

// NewService creates a new auth service
func NewService(
	repo domain.Repository,
	oauthRepo domain.OAuthRepository,
	keyRepo domain.KeyRepository,
	tokenManager *internalAuth.TokenManager,
	tokenDomainService domain.TokenDomainService,
	logger *slog.Logger,
	defaultTokenExpiry int,
) domain.Service {
	return &service{
		repo:               repo,
		oauthRepo:          oauthRepo,
		keyRepo:            keyRepo,
		tokenManager:       tokenManager,
		tokenDomainService: tokenDomainService,
		logger:             logger,
		defaultTokenExpiry: defaultTokenExpiry,
	}
}

func (s *service) GetAuthURL(ctx context.Context, req *domain.LoginRequest) (string, string, error) {
	// Generate state
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Generate PKCE challenge if provided
	var codeChallenge string
	if req.CodeChallenge != "" {
		codeChallenge = req.CodeChallenge
	}

	// Store auth state
	authState := &domain.AuthState{
		State:        state,
		Provider:     req.Provider,
		RedirectURL:  req.RedirectURL,
		CodeVerifier: codeChallenge, // Store for later verification
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		CreatedAt:    time.Now(),
	}

	if err := s.repo.StoreAuthState(ctx, authState); err != nil {
		return "", "", fmt.Errorf("failed to store auth state: %w", err)
	}

	// Get auth URL with parameters
	params := map[string]string{}
	if codeChallenge != "" {
		params["code_challenge"] = codeChallenge
		params["code_challenge_method"] = req.CodeChallengeMethod
	}

	authURL, err := s.oauthRepo.GetAuthURL(req.Provider, state, params)
	if err != nil {
		return "", "", fmt.Errorf("failed to get auth URL: %w", err)
	}

	return authURL, state, nil
}

func (s *service) HandleCallback(ctx context.Context, req *domain.CallbackRequest, clientIP, userAgent string) (*domain.AuthResponse, error) {
	// Verify state
	if err := s.VerifyAuthState(ctx, req.State, clientIP); err != nil {
		return nil, fmt.Errorf("invalid state: %w", err)
	}

	// Get auth state
	authState, err := s.repo.GetAuthState(ctx, req.State)
	if err != nil {
		return nil, fmt.Errorf("auth state not found: %w", err)
	}

	// Verify PKCE if provided
	if req.CodeVerifier != "" {
		if err := s.VerifyPKCE(ctx, req.State, req.CodeVerifier); err != nil {
			return nil, err
		}
	}

	// Exchange code for tokens
	oauthToken, err := s.oauthRepo.ExchangeCode(ctx, authState.Provider, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from provider
	userInfo, err := s.oauthRepo.GetUserInfo(ctx, authState.Provider, oauthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := s.repo.GetUserByExternalID(ctx, userInfo.ID, authState.Provider)
	if err != nil {
		// Create new user
		user = &domain.User{
			ID:          uuid.New().String(),
			ExternalID:  userInfo.ID,
			Provider:    authState.Provider,
			Email:       userInfo.Email,
			DisplayName: userInfo.Name,
			AvatarURL:   userInfo.Picture,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			LastLoginAt: time.Now(),
		}

		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Log security event
		s.logSecurityEvent(ctx, user.ID, "user_created", "New user created via OAuth", clientIP, userAgent, "info")
	} else {
		// Update last login
		if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
			s.logger.Error("failed to update last login", "error", err)
		}
	}

	// Generate session ID first
	sessionID := uuid.New().String()

	// Generate tokens with session ID
	tokenPair, err := s.generateTokenPair(ctx, user, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session with the pre-generated session ID
	session, err := s.tokenDomainService.CreateSession(sessionID, user.ID, tokenPair.RefreshToken, "", clientIP, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	// Save session to repository
	err = s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Clean up auth state
	if err := s.repo.DeleteAuthState(ctx, req.State); err != nil {
		s.logger.Error("failed to delete auth state", "error", err)
	}

	// Log security event
	s.logSecurityEvent(ctx, user.ID, "login_success", "Successful OAuth login", clientIP, userAgent, "info")

	return &domain.AuthResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken, clientIP, userAgent string) (*domain.TokenPair, error) {
	// Infrastructure concerns: Check if token is blacklisted
	blacklisted, err := s.repo.IsRefreshTokenBlacklisted(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if blacklisted {
		return nil, fmt.Errorf("refresh token is invalid")
	}

	// Infrastructure concerns: Get session by refresh token
	session, err := s.repo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("session has expired")
	}

	// Infrastructure concerns: Get user
	user, err := s.repo.GetUser(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Business logic: Apply domain rules through domain service
	newClaims, err := s.tokenDomainService.RefreshToken(ctx, session, user)
	if err != nil {
		return nil, fmt.Errorf("refresh validation failed: %w", err)
	}

	// Infrastructure concerns: Generate token pair using new claims
	tokenPair, err := s.generateTokenPairFromClaims(ctx, newClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Infrastructure concerns: Blacklist old refresh token
	if err := s.repo.BlacklistRefreshToken(ctx, refreshToken, session.ExpiresAt); err != nil {
		s.logger.Error("failed to blacklist old refresh token", "error", err)
	}

	// Infrastructure concerns: Update session with new refresh token
	session.RefreshToken = tokenPair.RefreshToken
	session.LastUsedAt = time.Now()
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Infrastructure concerns: Log security event
	s.logSecurityEvent(ctx, user.ID, "token_refreshed", "Access token refreshed", clientIP, userAgent, "info")

	return tokenPair, nil
}

func (s *service) CreateSession(ctx context.Context, userID, refreshToken, deviceID, clientIP, userAgent string) (*domain.Session, error) {
	session := &domain.Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		RefreshToken: refreshToken,
		DeviceID:     deviceID,
		IPAddress:    clientIP,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (s *service) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	return s.repo.GetSession(ctx, sessionID)
}

func (s *service) GetUserSessions(ctx context.Context, userID string) ([]*domain.SessionInfo, error) {
	sessions, err := s.repo.ListUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessionInfos := make([]*domain.SessionInfo, len(sessions))
	for i, session := range sessions {
		sessionInfos[i] = &domain.SessionInfo{
			ID:         session.ID,
			DeviceID:   session.DeviceID,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			CreatedAt:  session.CreatedAt,
			LastUsedAt: session.LastUsedAt,
			IsCurrent:  false, // TODO: Determine current session
		}

		// Parse user agent for better display
		if session.UserAgent != "" {
			// Simple parsing - in production use a proper UA parser
			if strings.Contains(session.UserAgent, "Chrome") {
				sessionInfos[i].Location = "Chrome Browser"
			} else if strings.Contains(session.UserAgent, "Firefox") {
				sessionInfos[i].Location = "Firefox Browser"
			} else if strings.Contains(session.UserAgent, "Safari") {
				sessionInfos[i].Location = "Safari Browser"
			}
		}
	}

	return sessionInfos, nil
}

func (s *service) RevokeSession(ctx context.Context, userID, sessionID string) error {
	// Verify session belongs to user
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Blacklist refresh token
	if err := s.repo.BlacklistRefreshToken(ctx, session.RefreshToken, session.ExpiresAt); err != nil {
		s.logger.Error("failed to blacklist refresh token", "error", err)
	}

	// Delete session
	if err := s.repo.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Log security event
	s.logSecurityEvent(ctx, userID, "session_revoked", fmt.Sprintf("Session %s revoked", sessionID), "", "", "info")

	return nil
}

func (s *service) RevokeAllSessions(ctx context.Context, userID string, exceptSessionID string) error {
	// Get all user sessions
	sessions, err := s.repo.ListUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Blacklist all refresh tokens except current
	for _, session := range sessions {
		if session.ID != exceptSessionID {
			if err := s.repo.BlacklistRefreshToken(ctx, session.RefreshToken, session.ExpiresAt); err != nil {
				s.logger.Error("failed to blacklist refresh token", "error", err)
			}
		}
	}

	// Delete all sessions except current
	if err := s.repo.DeleteUserSessions(ctx, userID, exceptSessionID); err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	// Log security event
	s.logSecurityEvent(ctx, userID, "all_sessions_revoked", "All sessions revoked except current", "", "", "warning")

	return nil
}

func (s *service) ValidateSession(ctx context.Context, sessionID, clientIP string) error {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Check if expired
	if session.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("session expired")
	}

	// Check IP change (optional security measure)
	if session.IPAddress != clientIP {
		s.logSecurityEvent(ctx, session.UserID, "session_ip_changed", 
			fmt.Sprintf("Session IP changed from %s to %s", session.IPAddress, clientIP), 
			clientIP, session.UserAgent, "warning")
	}

	// Update last used
	session.LastUsedAt = time.Now()
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		s.logger.Error("failed to update session last used", "error", err)
	}

	return nil
}

func (s *service) ValidateAccessToken(ctx context.Context, tokenString string) (*domain.Claims, error) {
	// Handle development tokens
	if strings.HasPrefix(tokenString, "dev_token_") {
		// For development mode, return mock claims
		return &domain.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "dev-user-1",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID:    "dev-user-1",
			Email:     "test@hexabase.com",
			Name:      "Test User",
			Provider:  "credentials",
			OrgIDs:    []string{"dev-org-1"}, // Include development organization
			SessionID: "dev-session-1",
		}, nil
	}

	// Get public key for production tokens
	publicKeyPEM, err := s.keyRepo.GetPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Parse public key
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	// Extract claims
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims")
	}

	claims := &domain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: mapClaims["sub"].(string),
			ExpiresAt: func() *jwt.NumericDate {
				if exp, ok := mapClaims["exp"].(float64); ok {
					return jwt.NewNumericDate(time.Unix(int64(exp), 0))
				}
				return nil
			}(),
			IssuedAt: func() *jwt.NumericDate {
				if iat, ok := mapClaims["iat"].(float64); ok {
					return jwt.NewNumericDate(time.Unix(int64(iat), 0))
				}
				return nil
			}(),
		},
		UserID:   mapClaims["sub"].(string),
		Email:    mapClaims["email"].(string),
		Name:     mapClaims["name"].(string),
		Provider: mapClaims["provider"].(string),
	}

	// Set SessionID if present
	if sessionID, ok := mapClaims["session_id"].(string); ok {
		claims.SessionID = sessionID
	} else {
		claims.SessionID = "legacy-session"
	}

	// Extract org IDs if present
	if orgIDs, ok := mapClaims["org_ids"].([]interface{}); ok {
		claims.OrgIDs = make([]string, len(orgIDs))
		for i, id := range orgIDs {
			claims.OrgIDs[i] = id.(string)
		}
	}

	return claims, nil
}

func (s *service) GenerateWorkspaceToken(ctx context.Context, userID, workspaceID string) (string, error) {
	// Verify user exists
	_, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Get user's workspace groups
	groups, err := s.repo.GetUserWorkspaceGroups(ctx, userID, workspaceID)
	if err != nil {
		return "", fmt.Errorf("failed to get user groups: %w", err)
	}

	// Create workspace claims
	claims := &domain.WorkspaceClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Subject:     userID,
		WorkspaceID: workspaceID,
		Groups:      groups,
	}

	// Get private key
	privateKeyPEM, err := s.keyRepo.GetPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	// Parse private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// generateTokenPairFromClaims creates a token pair from pre-generated claims
func (s *service) generateTokenPairFromClaims(ctx context.Context, claims *domain.Claims) (*domain.TokenPair, error) {
	// Use TokenManager to sign claims
	accessToken, err := s.tokenManager.SignClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Calculate expiration time from claims
	var expiresIn int
	if claims.ExpiresAt != nil && claims.IssuedAt != nil {
		expiresIn = int(claims.ExpiresAt.Sub(claims.IssuedAt.Time).Seconds())
	} else {
		// Use configurable default instead of hardcoded value
		// TODO: Consider making this configurable per user/organization/session type
		expiresIn = s.defaultTokenExpiry
		s.logger.Warn("using default token expiry due to missing claims timestamps", 
			"default_expiry_seconds", expiresIn,
			"user_id", claims.UserID,
		)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

// generateRefreshToken generates a new refresh token
func (s *service) generateRefreshToken() (string, error) {
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(refreshTokenBytes), nil
}

func (s *service) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	// Get session to find expiry
	session, err := s.repo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Blacklist token
	if err := s.repo.BlacklistRefreshToken(ctx, refreshToken, session.ExpiresAt); err != nil {
		return fmt.Errorf("failed to blacklist refresh token: %w", err)
	}

	return nil
}

func (s *service) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.GetUser(ctx, userID)
}

func (s *service) GetCurrentUser(ctx context.Context, token string) (*domain.User, error) {
	// Validate token and get claims
	claims, err := s.ValidateAccessToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get user
	user, err := s.repo.GetUser(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (s *service) UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) (*domain.User, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Apply updates
	if displayName, ok := updates["display_name"].(string); ok {
		user.DisplayName = displayName
	}
	if avatarURL, ok := updates["avatar_url"].(string); ok {
		user.AvatarURL = avatarURL
	}

	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *service) LogSecurityEvent(ctx context.Context, event *domain.SecurityEvent) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()

	if err := s.repo.CreateSecurityEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to log security event: %w", err)
	}

	return nil
}

func (s *service) GetSecurityLogs(ctx context.Context, filter domain.SecurityLogFilter) ([]*domain.SecurityEvent, error) {
	return s.repo.ListSecurityEvents(ctx, filter)
}

func (s *service) GetJWKS(ctx context.Context) ([]byte, error) {
	return s.keyRepo.GetJWKS()
}

func (s *service) GetOIDCConfiguration(ctx context.Context) (map[string]interface{}, error) {
	// Return OIDC discovery document
	config := map[string]interface{}{
		"issuer":                 "https://api.hexabase-kaas.io",
		"authorization_endpoint": "https://api.hexabase-kaas.io/auth/authorize",
		"token_endpoint":         "https://api.hexabase-kaas.io/auth/token",
		"userinfo_endpoint":      "https://api.hexabase-kaas.io/auth/userinfo",
		"jwks_uri":               "https://api.hexabase-kaas.io/.well-known/jwks.json",
		"response_types_supported": []string{"code"},
		"subject_types_supported":  []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported": []string{"openid", "profile", "email"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic"},
		"claims_supported": []string{
			"sub", "email", "name", "picture", "provider", "org_ids",
		},
	}

	return config, nil
}

func (s *service) StoreAuthState(ctx context.Context, state *domain.AuthState) error {
	return s.repo.StoreAuthState(ctx, state)
}

func (s *service) VerifyAuthState(ctx context.Context, state, clientIP string) error {
	authState, err := s.repo.GetAuthState(ctx, state)
	if err != nil {
		return fmt.Errorf("auth state not found: %w", err)
	}

	// Check expiry
	if authState.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("auth state expired")
	}

	// Optionally check IP
	// if authState.ClientIP != clientIP {
	// 	return fmt.Errorf("client IP mismatch")
	// }

	return nil
}

func (s *service) VerifyPKCE(ctx context.Context, state, codeVerifier string) error {
	authState, err := s.repo.GetAuthState(ctx, state)
	if err != nil {
		return fmt.Errorf("auth state not found: %w", err)
	}

	if authState.CodeVerifier == "" {
		return nil // PKCE not required
	}

	// Verify code verifier matches stored challenge
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	computedChallenge := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if computedChallenge != authState.CodeVerifier {
		return fmt.Errorf("PKCE verification failed")
	}

	return nil
}

// Helper functions

func (s *service) generateTokenPair(ctx context.Context, user *domain.User, sessionID string) (*domain.TokenPair, error) {
	// Get user's organizations
	orgIDs, err := s.repo.GetUserOrganizations(ctx, user.ID)
	if err != nil {
		s.logger.Warn("failed to get user organizations", "error", err)
		orgIDs = []string{}
	}

	// Create domain claims
	now := time.Now()
	claims := &domain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			Issuer:    "https://api.hexabase-kaas.io",
			Audience:  []string{"hexabase-api"},
		},
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.DisplayName,
		Provider:  user.Provider,
		OrgIDs:    orgIDs,
		SessionID: sessionID,
	}

	// Use common token pair generation logic
	return s.generateTokenPairFromClaims(ctx, claims)
}

func (s *service) logSecurityEvent(ctx context.Context, userID, eventType, description, ipAddress, userAgent, level string) {
	event := &domain.SecurityEvent{
		ID:          uuid.New().String(),
		UserID:      userID,
		EventType:   eventType,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Level:       level,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateSecurityEvent(ctx, event); err != nil {
		s.logger.Error("failed to log security event", "error", err)
	}
}

func (s *service) GenerateInternalAIOpsToken(ctx context.Context, userID string, orgIDs []string, activeWorkspaceID string) (string, error) {
	claims := &InternalAIOpsClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Second)), // Very short-lived
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "hexabase-control-plane",
			Audience:  jwt.ClaimStrings{"hexabase-aiops-service"},
		},
		UserID:            userID,
		OrgIDs:            orgIDs,
		ActiveWorkspaceID: activeWorkspaceID,
		TokenType:         "internal-aiops-v1",
	}

	privateKeyPEM, err := s.keyRepo.GetPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key for internal token: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key for internal token: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign internal token: %w", err)
	}

	return tokenString, nil
}