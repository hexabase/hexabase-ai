package domain

import (
	"context"
	"time"
)

// Repository defines the data access interface for authentication
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, userID string) (*User, error)
	GetUserByExternalID(ctx context.Context, externalID, provider string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	UpdateLastLogin(ctx context.Context, userID string) error

	// Session operations
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	GetAllActiveSessions(ctx context.Context) ([]*Session, error)
	ListUserSessions(ctx context.Context, userID string) ([]*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteUserSessions(ctx context.Context, userID string, exceptSessionID string) error
	CleanupExpiredSessions(ctx context.Context, before time.Time) error

	// Auth state operations (Redis)
	StoreAuthState(ctx context.Context, state *AuthState) error
	GetAuthState(ctx context.Context, stateValue string) (*AuthState, error)
	DeleteAuthState(ctx context.Context, stateValue string) error

	// Refresh token operations (Redis)
	BlacklistRefreshToken(ctx context.Context, token string, expiresAt time.Time) error
	IsRefreshTokenBlacklisted(ctx context.Context, token string) (bool, error)

	// Security event operations
	CreateSecurityEvent(ctx context.Context, event *SecurityEvent) error
	ListSecurityEvents(ctx context.Context, filter SecurityLogFilter) ([]*SecurityEvent, error)
	CleanupOldSecurityEvents(ctx context.Context, before time.Time) error

	// User organization operations
	GetUserOrganizations(ctx context.Context, userID string) ([]string, error)
	GetUserWorkspaceGroups(ctx context.Context, userID, workspaceID string) ([]string, error)

	// Token hashing operations (infrastructure layer - pure crypto without business logic)
	HashToken(token string) (hashedToken string, salt string, err error)
	VerifyToken(plainToken, hashedToken, salt string) bool
}

// OAuthRepository defines the interface for OAuth operations
type OAuthRepository interface {
	// Provider configuration
	GetProviderConfig(provider string) (*ProviderConfig, error)

	// OAuth flow
	GetAuthURL(provider, state string, params map[string]string) (string, error)
	ExchangeCode(ctx context.Context, provider, code string) (*OAuthToken, error)
	GetUserInfo(ctx context.Context, provider string, token *OAuthToken) (*UserInfo, error)
	RefreshOAuthToken(ctx context.Context, provider string, refreshToken string) (*OAuthToken, error)
}

// KeyRepository defines the interface for key management
type KeyRepository interface {
	// RSA keys for JWT signing
	GetPrivateKey() ([]byte, error)
	GetPublicKey() ([]byte, error)
	GetJWKS() ([]byte, error)
	RotateKeys() error
}

// ProviderConfig represents OAuth provider configuration
type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

// OAuthToken represents OAuth tokens
type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	Scope        string
}