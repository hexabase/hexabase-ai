package domain

import (
	"context"
)

// Service defines the authentication business logic interface
type Service interface {
	// OAuth authentication
	GetAuthURL(ctx context.Context, req *LoginRequest) (string, string, error)
	HandleCallback(ctx context.Context, req *CallbackRequest, clientIP, userAgent string) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken, clientIP, userAgent string) (*TokenPair, error)
	
	// Session management
	CreateSession(ctx context.Context, userID, refreshToken, deviceID, clientIP, userAgent string) (*Session, error)
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	GetUserSessions(ctx context.Context, userID string) ([]*SessionInfo, error)
	RevokeSession(ctx context.Context, userID, sessionID string) error
	RevokeAllSessions(ctx context.Context, userID string, exceptSessionID string) error
	ValidateSession(ctx context.Context, sessionID, clientIP string) error
	
	// Token management
	ValidateAccessToken(ctx context.Context, token string) (*Claims, error)
	GenerateWorkspaceToken(ctx context.Context, userID, workspaceID string) (string, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	
	// User management
	GetUser(ctx context.Context, userID string) (*User, error)
	GetCurrentUser(ctx context.Context, token string) (*User, error)
	UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) (*User, error)
	
	// Security
	LogSecurityEvent(ctx context.Context, event *SecurityEvent) error
	GetSecurityLogs(ctx context.Context, filter SecurityLogFilter) ([]*SecurityEvent, error)
	
	// JWKS and OIDC
	GetJWKS(ctx context.Context) ([]byte, error)
	GetOIDCConfiguration(ctx context.Context) (map[string]interface{}, error)
	
	// State management
	StoreAuthState(ctx context.Context, state *AuthState) error
	VerifyAuthState(ctx context.Context, state, clientIP string) error
	
	// PKCE
	VerifyPKCE(ctx context.Context, state, codeVerifier string) error

	// Internal communication
	GenerateInternalAIOpsToken(ctx context.Context, userID string, orgIDs []string, activeWorkspaceID string) (string, error)
}