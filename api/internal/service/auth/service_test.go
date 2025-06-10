package auth

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) CreateUser(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockRepository) GetUser(ctx context.Context, userID string) (*auth.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetUserByExternalID(ctx context.Context, externalID, provider string) (*auth.User, error) {
	args := m.Called(ctx, externalID, provider)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) UpdateUser(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockRepository) CreateSession(ctx context.Context, session *auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockRepository) GetSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.Session), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*auth.Session, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.Session), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) ListUserSessions(ctx context.Context, userID string) ([]*auth.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) != nil {
		return args.Get(0).([]*auth.Session), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) UpdateSession(ctx context.Context, session *auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockRepository) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockRepository) DeleteUserSessions(ctx context.Context, userID string, exceptSessionID string) error {
	args := m.Called(ctx, userID, exceptSessionID)
	return args.Error(0)
}

func (m *mockRepository) CleanupExpiredSessions(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *mockRepository) StoreAuthState(ctx context.Context, state *auth.AuthState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *mockRepository) GetAuthState(ctx context.Context, stateValue string) (*auth.AuthState, error) {
	args := m.Called(ctx, stateValue)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.AuthState), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) DeleteAuthState(ctx context.Context, stateValue string) error {
	args := m.Called(ctx, stateValue)
	return args.Error(0)
}

func (m *mockRepository) BlacklistRefreshToken(ctx context.Context, token string, expiresAt time.Time) error {
	args := m.Called(ctx, token, expiresAt)
	return args.Error(0)
}

func (m *mockRepository) IsRefreshTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *mockRepository) CreateSecurityEvent(ctx context.Context, event *auth.SecurityEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockRepository) ListSecurityEvents(ctx context.Context, filter auth.SecurityLogFilter) ([]*auth.SecurityEvent, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) != nil {
		return args.Get(0).([]*auth.SecurityEvent), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) CleanupOldSecurityEvents(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *mockRepository) GetUserOrganizations(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetUserWorkspaceGroups(ctx context.Context, userID, workspaceID string) ([]string, error) {
	args := m.Called(ctx, userID, workspaceID)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

// Mock OAuth repository
type mockOAuthRepository struct {
	mock.Mock
}

func (m *mockOAuthRepository) GetProviderConfig(provider string) (*auth.ProviderConfig, error) {
	args := m.Called(provider)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.ProviderConfig), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOAuthRepository) GetAuthURL(provider, state string, params map[string]string) (string, error) {
	args := m.Called(provider, state, params)
	return args.String(0), args.Error(1)
}

func (m *mockOAuthRepository) ExchangeCode(ctx context.Context, provider, code string) (*auth.OAuthToken, error) {
	args := m.Called(ctx, provider, code)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.OAuthToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOAuthRepository) GetUserInfo(ctx context.Context, provider string, token *auth.OAuthToken) (*auth.UserInfo, error) {
	args := m.Called(ctx, provider, token)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.UserInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOAuthRepository) RefreshOAuthToken(ctx context.Context, provider string, refreshToken string) (*auth.OAuthToken, error) {
	args := m.Called(ctx, provider, refreshToken)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.OAuthToken), args.Error(1)
	}
	return nil, args.Error(1)
}

// Mock key repository
type mockKeyRepository struct {
	mock.Mock
}

func (m *mockKeyRepository) GetPrivateKey() ([]byte, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockKeyRepository) GetPublicKey() ([]byte, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockKeyRepository) GetJWKS() ([]byte, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockKeyRepository) RotateKeys() error {
	args := m.Called()
	return args.Error(0)
}


func TestService_GetAuthURL(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockOAuthRepo := new(mockOAuthRepository)
	mockKeyRepo := new(mockKeyRepository)
	
	svc := &service{
		repo:      mockRepo,
		oauthRepo: mockOAuthRepo,
		keyRepo:   mockKeyRepo,
		logger:    slog.Default(),
	}

	t.Run("successful get auth URL", func(t *testing.T) {
		req := &auth.LoginRequest{
			Provider: "google",
		}

		expectedURL := "https://accounts.google.com/o/oauth2/v2/auth?state=random-state-123"

		// Mock repository calls
		mockRepo.On("StoreAuthState", ctx, mock.AnythingOfType("*auth.AuthState")).Return(nil)
		mockOAuthRepo.On("GetAuthURL", "google", mock.AnythingOfType("string"), mock.Anything).Return(expectedURL, nil)

		url, state, err := svc.GetAuthURL(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
		assert.NotEmpty(t, state)

		mockRepo.AssertExpectations(t)
		mockOAuthRepo.AssertExpectations(t)
	})

	t.Run("invalid provider", func(t *testing.T) {
		req := &auth.LoginRequest{
			Provider: "invalid-provider",
		}

		mockRepo.On("StoreAuthState", ctx, mock.AnythingOfType("*auth.AuthState")).Return(nil)
		mockOAuthRepo.On("GetAuthURL", "invalid-provider", mock.AnythingOfType("string"), mock.Anything).
			Return("", errors.New("unsupported provider"))

		url, state, err := svc.GetAuthURL(ctx, req)
		assert.Error(t, err)
		assert.Empty(t, url)
		assert.Empty(t, state)
	})
}

func TestService_HandleCallback(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockOAuthRepo := new(mockOAuthRepository)
	mockKeyRepo := new(mockKeyRepository)
	
	svc := &service{
		repo:       mockRepo,
		oauthRepo:  mockOAuthRepo,
		keyRepo:    mockKeyRepo,
		logger:     slog.Default(),
	}

	t.Run("successful callback - new user", func(t *testing.T) {
		req := &auth.CallbackRequest{
			Code:  "auth-code-123",
			State: "valid-state-123",
		}
		clientIP := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		authState := &auth.AuthState{
			State:    "valid-state-123",
			Provider: "google",
		}

		oauthToken := &auth.OAuthToken{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-123",
		}

		userInfo := &auth.UserInfo{
			ID:       "google-123",
			Email:    "user@example.com",
			Name:     "Test User",
			Provider: "google",
		}

		// Mock private key for token generation
		privateKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF32M0w5md/Y2c2Shj5cHOtIVYC1Y
/c2Id85e5iJGPO85VaFfpUW0DwvnkLanf2rgQvSCFUvz8L6CLLxOZjG5MwJwmhE/
ZFqEuHqIbpJfDFhkKjBLJmkEDDW2KFvquTe9+CZmI3O6eSspJWRQUfACLRCOFRHc
Vz7RppLnzXvQqgMKiCxvUBhJRaAQJlCvDafI8MrDwAzdllbHFQhVVv6sKNnStzHZ
EqGKN3f0GanVfJUw8Ys1wyYmMWNYbJbIiJGXLYSWwElaUltFapKhKjC7OvHLfJ5P
4HUR1ON6lxYl3ciTisg0VkHAcDaGLmhK7YqUgwIDAQABAoIBAD2wzqIms6LHJIoG
SLoSJc/lQItrFMnLBHBF3aP4J9zHO3KJmHbr1h+v9OdJq0Uv3k4DRCx55xtDC0++
LGDNGKxMg5qFN1FYuF+ALmwDyx5tuxCOZ8p32Rv6iZ9efgh9InVyKFOPiUMX8Sjn
S+x0S6LSXB7HIUcMXih37vc9qLI1K3BcNdAGQm1xNZKepIDkztMULdVcQPuGvr6T
MsJ7vANmBLwHTP1UXNdG8nxo9AQs/RCEgIFxQmKQSpHw3bqBR4eHMrn5aUPrmQKG
Lg4OGgP0yWZhGAqm2x7VHvTznHlvDZBAyycf3x+gDh4hFq0nqDZiB9lNsa5Ooqwz
4sLvovECgYEA6eVMd2C4ffwHFhrBCBx5e6logOXqxkpE2MDTrWL7jHJGgdb3bCXU
Qx0BhbG2PU+gaPe9cQSz6gCLDt7x/zJq6Qb5TXp6rZbbQiLnKYQQbb5Z7ZV6zpJk
w1gnqrDi8a1eexMdpvr+pHK7IjYaTPRvFsyM4j9Vo1GlMD96ZBaGiI8CgYEA5Z2X
z9f7f0G7bqZOvcJVCPL5WANpyPFmCx1Gx0Sg6pNOdIIJJtUYwXIXgF9Vf+i2BbO5
byaNMixLFNIUZvFb5shc8HrnQ0qVLuEpyHh0SOIYU3jE4yZECZxFnkd5ZTQqNdun
VtLNGDGXU8L2cCF0rGcLd3aVQJL5RR0i3DodYv0CgYAyX7xGIxGZECIR8WwcJyRF
aHE8LY+6qWKJIMH0rM8N0XvsuWI7u3kkcOEJYWdq0V5OjdVaobFdxP9x7NIhvGeT
pQ3k1yvvvMCGNiGLq2N7i5qiVLU4vD5q9rCh1HQZDa12lXEjUDQbRuFMPmh+7F+o
DuWDq3iikCBnIVnlXM4FGQKBgH4HZyMDQfBLupzcO0kC6F1P7bc1j/itIvCUoTHv
rFTMZRsKcDvE4xvxm6aTqrv+Q8I3AeoXYGNTm2S9F8KGQNRs/4s8fngTanhUJoWJ
4GhHldgqGF7LHMsBG+wQxWgGRANwcy/8mm5n6e6SlJQWdmDpfrB6HrGGP9Z3i0K4
TqMdAoGBAMvQm5j8yZSrPa8TMLxUKMjQVQ7CkGIDTZaIQrcVMU5PKNtryp7LQJkl
a1rAOo4S1l7kZX8b6HBRZLNX+zqLAQ3q+qu/PWhYGz3FvFocKqJSFl6bPdHVWHGD
IwvA1RJNp2TVgqetD1QYn7BdJCz/LoYUJ4cUF6j4BFxWGQJBfwuo
-----END RSA PRIVATE KEY-----`)

		// Mock the flow
		mockRepo.On("GetAuthState", ctx, "valid-state-123").Return(authState, nil)
		mockRepo.On("DeleteAuthState", ctx, "valid-state-123").Return(nil)
		mockOAuthRepo.On("ExchangeCode", ctx, "google", "auth-code-123").Return(oauthToken, nil)
		mockOAuthRepo.On("GetUserInfo", ctx, "google", oauthToken).Return(userInfo, nil)
		
		// User doesn't exist yet
		mockRepo.On("GetUserByExternalID", ctx, "google-123", "google").Return(nil, errors.New("not found"))
		mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*auth.User")).Return(nil)
		mockRepo.On("UpdateLastLogin", ctx, mock.AnythingOfType("string")).Return(nil)
		mockRepo.On("CreateSecurityEvent", ctx, mock.AnythingOfType("*auth.SecurityEvent")).Return(nil).Times(2)
		
		// Generate tokens
		mockRepo.On("GetUserOrganizations", ctx, mock.AnythingOfType("string")).Return([]string{}, nil)
		mockKeyRepo.On("GetPrivateKey").Return(privateKey, nil)
		
		// Create session
		mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*auth.Session")).Return(nil)

		response, err := svc.HandleCallback(ctx, req, clientIP, userAgent)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.User)
		assert.Equal(t, "user@example.com", response.User.Email)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)

		mockRepo.AssertExpectations(t)
		mockOAuthRepo.AssertExpectations(t)
		mockKeyRepo.AssertExpectations(t)
	})

	t.Run("invalid state", func(t *testing.T) {
		req := &auth.CallbackRequest{
			Code:  "auth-code-789",
			State: "invalid-state",
		}
		clientIP := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		mockRepo.On("GetAuthState", ctx, "invalid-state").Return(nil, errors.New("not found"))

		response, err := svc.HandleCallback(ctx, req, clientIP, userAgent)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid state")

		mockRepo.AssertExpectations(t)
	})
}

func TestService_RefreshToken(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockKeyRepo := new(mockKeyRepository)
	
	svc := &service{
		repo:       mockRepo,
		keyRepo:    mockKeyRepo,
		logger:     slog.Default(),
	}

	// Mock private key for token generation
	privateKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF32M0w5md/Y2c2Shj5cHOtIVYC1Y
/c2Id85e5iJGPO85VaFfpUW0DwvnkLanf2rgQvSCFUvz8L6CLLxOZjG5MwJwmhE/
ZFqEuHqIbpJfDFhkKjBLJmkEDDW2KFvquTe9+CZmI3O6eSspJWRQUfACLRCOFRHc
Vz7RppLnzXvQqgMKiCxvUBhJRaAQJlCvDafI8MrDwAzdllbHFQhVVv6sKNnStzHZ
EqGKN3f0GanVfJUw8Ys1wyYmMWNYbJbIiJGXLYSWwElaUltFapKhKjC7OvHLfJ5P
4HUR1ON6lxYl3ciTisg0VkHAcDaGLmhK7YqUgwIDAQABAoIBAD2wzqIms6LHJIoG
SLoSJc/lQItrFMnLBHBF3aP4J9zHO3KJmHbr1h+v9OdJq0Uv3k4DRCx55xtDC0++
LGDNGKxMg5qFN1FYuF+ALmwDyx5tuxCOZ8p32Rv6iZ9efgh9InVyKFOPiUMX8Sjn
S+x0S6LSXB7HIUcMXih37vc9qLI1K3BcNdAGQm1xNZKepIDkztMULdVcQPuGvr6T
MsJ7vANmBLwHTP1UXNdG8nxo9AQs/RCEgIFxQmKQSpHw3bqBR4eHMrn5aUPrmQKG
Lg4OGgP0yWZhGAqm2x7VHvTznHlvDZBAyycf3x+gDh4hFq0nqDZiB9lNsa5Ooqwz
4sLvovECgYEA6eVMd2C4ffwHFhrBCBx5e6logOXqxkpE2MDTrWL7jHJGgdb3bCXU
Qx0BhbG2PU+gaPe9cQSz6gCLDt7x/zJq6Qb5TXp6rZbbQiLnKYQQbb5Z7ZV6zpJk
w1gnqrDi8a1eexMdpvr+pHK7IjYaTPRvFsyM4j9Vo1GlMD96ZBaGiI8CgYEA5Z2X
z9f7f0G7bqZOvcJVCPL5WANpyPFmCx1Gx0Sg6pNOdIIJJtUYwXIXgF9Vf+i2BbO5
byaNMixLFNIUZvFb5shc8HrnQ0qVLuEpyHh0SOIYU3jE4yZECZxFnkd5ZTQqNdun
VtLNGDGXU8L2cCF0rGcLd3aVQJL5RR0i3DodYv0CgYAyX7xGIxGZECIR8WwcJyRF
aHE8LY+6qWKJIMH0rM8N0XvsuWI7u3kkcOEJYWdq0V5OjdVaobFdxP9x7NIhvGeT
pQ3k1yvvvMCGNiGLq2N7i5qiVLU4vD5q9rCh1HQZDa12lXEjUDQbRuFMPmh+7F+o
DuWDq3iikCBnIVnlXM4FGQKBgH4HZyMDQfBLupzcO0kC6F1P7bc1j/itIvCUoTHv
rFTMZRsKcDvE4xvxm6aTqrv+Q8I3AeoXYGNTm2S9F8KGQNRs/4s8fngTanhUJoWJ
4GhHldgqGF7LHMsBG+wQxWgGRANwcy/8mm5n6e6SlJQWdmDpfrB6HrGGP9Z3i0K4
TqMdAoGBAMvQm5j8yZSrPa8TMLxUKMjQVQ7CkGIDTZaIQrcVMU5PKNtryp7LQJkl
a1rAOo4S1l7kZX8b6HBRZLNX+zqLAQ3q+qu/PWhYGz3FvFocKqJSFl6bPdHVWHGD
IwvA1RJNp2TVgqetD1QYn7BdJCz/LoYUJ4cUF6j4BFxWGQJBfwuo
-----END RSA PRIVATE KEY-----`)

	t.Run("successful refresh", func(t *testing.T) {
		refreshToken := "valid-refresh-token"
		clientIP := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		user := &auth.User{
			ID:          "user-123",
			Email:       "user@example.com",
			DisplayName: "Test User",
			Provider:    "google",
		}

		session := &auth.Session{
			ID:           uuid.New().String(),
			UserID:       "user-123",
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		// Mock repository calls
		mockRepo.On("IsRefreshTokenBlacklisted", ctx, refreshToken).Return(false, nil)
		mockRepo.On("GetSessionByRefreshToken", ctx, refreshToken).Return(session, nil)
		mockRepo.On("GetUser", ctx, "user-123").Return(user, nil)
		mockRepo.On("GetUserOrganizations", ctx, "user-123").Return([]string{"org-1"}, nil)
		mockKeyRepo.On("GetPrivateKey").Return(privateKey, nil)
		mockRepo.On("BlacklistRefreshToken", ctx, refreshToken, session.ExpiresAt).Return(nil)
		mockRepo.On("UpdateSession", ctx, mock.AnythingOfType("*auth.Session")).Return(nil)
		mockRepo.On("CreateSecurityEvent", ctx, mock.AnythingOfType("*auth.SecurityEvent")).Return(nil)

		response, err := svc.RefreshToken(ctx, refreshToken, clientIP, userAgent)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Equal(t, 900, response.ExpiresIn)

		mockRepo.AssertExpectations(t)
		mockKeyRepo.AssertExpectations(t)
	})

	t.Run("blacklisted token", func(t *testing.T) {
		refreshToken := "blacklisted-token"
		clientIP := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		mockRepo.On("IsRefreshTokenBlacklisted", ctx, refreshToken).Return(true, nil)

		response, err := svc.RefreshToken(ctx, refreshToken, clientIP, userAgent)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid")

		mockRepo.AssertExpectations(t)
	})

	t.Run("expired session", func(t *testing.T) {
		refreshToken := "expired-token"
		clientIP := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		session := &auth.Session{
			ID:           uuid.New().String(),
			UserID:       "user-789",
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
		}

		mockRepo.On("IsRefreshTokenBlacklisted", ctx, refreshToken).Return(false, nil)
		mockRepo.On("GetSessionByRefreshToken", ctx, refreshToken).Return(session, nil)

		response, err := svc.RefreshToken(ctx, refreshToken, clientIP, userAgent)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "expired")

		mockRepo.AssertExpectations(t)
	})
}

func TestService_RevokeSession(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	
	svc := &service{
		repo:   mockRepo,
		logger: slog.Default(),
	}

	t.Run("successful revoke", func(t *testing.T) {
		userID := "user-123"
		sessionID := "session-123"
		refreshToken := "refresh-token-123"

		session := &auth.Session{
			ID:           sessionID,
			UserID:       userID,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		// Mock repository calls
		mockRepo.On("GetSession", ctx, sessionID).Return(session, nil)
		mockRepo.On("BlacklistRefreshToken", ctx, refreshToken, session.ExpiresAt).Return(nil)
		mockRepo.On("DeleteSession", ctx, sessionID).Return(nil)
		mockRepo.On("CreateSecurityEvent", ctx, mock.AnythingOfType("*auth.SecurityEvent")).Return(nil)

		err := svc.RevokeSession(ctx, userID, sessionID)
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		userID := "user-456"
		sessionID := "non-existent"

		mockRepo.On("GetSession", ctx, sessionID).Return(nil, errors.New("not found"))

		err := svc.RevokeSession(ctx, userID, sessionID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		mockRepo.AssertExpectations(t)
	})

	t.Run("user mismatch", func(t *testing.T) {
		userID := "user-789"
		sessionID := "session-789"

		session := &auth.Session{
			ID:     sessionID,
			UserID: "different-user",
		}

		mockRepo.On("GetSession", ctx, sessionID).Return(session, nil)

		err := svc.RevokeSession(ctx, userID, sessionID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")

		mockRepo.AssertExpectations(t)
	})
}