package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationDomain_CreateUser(t *testing.T) {
	t.Run("create user with valid data", func(t *testing.T) {
		user := &User{
			ID:          "user-123",
			ExternalID:  "google-456",
			Provider:    "google",
			Email:       "test@example.com",
			DisplayName: "Test User",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := user.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "user-123", user.ID)
		assert.Equal(t, "google-456", user.ExternalID)
		assert.Equal(t, "google", user.Provider)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.DisplayName)
	})
}

func TestAuthenticationDomain_CreateUser_ValidationError(t *testing.T) {
	tests := []struct {
		name        string
		user        *User
		expectedErr string
	}{
		{
			name: "missing email",
			user: &User{
				ID:          "user-123",
				Provider:    "google",
				DisplayName: "Test User",
			},
			expectedErr: "email is required",
		},
		{
			name: "missing provider",
			user: &User{
				ID:          "user-123",
				Email:       "test@example.com",
				DisplayName: "Test User",
			},
			expectedErr: "provider is required",
		},
		{
			name: "empty email",
			user: &User{
				ID:          "user-123",
				Provider:    "google",
				Email:       "",
				DisplayName: "Test User",
			},
			expectedErr: "email is required",
		},
		{
			name: "empty provider",
			user: &User{
				ID:          "user-123",
				Provider:    "",
				Email:       "test@example.com",
				DisplayName: "Test User",
			},
			expectedErr: "provider is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestAuthenticationDomain_CreateSession(t *testing.T) {
	t.Run("create session with proper expiration", func(t *testing.T) {
		now := time.Now()
		session := &Session{
			ID:           "session-123",
			UserID:       "user-456",
			RefreshToken: "refresh-token-789",
			DeviceID:     "device-abc",
			IPAddress:    "192.168.1.1",
			UserAgent:    "Mozilla/5.0",
			ExpiresAt:    now.Add(30 * 24 * time.Hour), // 30 days
			CreatedAt:    now,
			LastUsedAt:   now,
			Revoked:      false,
		}

		assert.Equal(t, "session-123", session.ID)
		assert.Equal(t, "user-456", session.UserID)
		assert.Equal(t, "refresh-token-789", session.RefreshToken)
		assert.False(t, session.IsExpired())
		assert.True(t, session.IsValid())
	})
}

func TestAuthenticationDomain_ValidateSession(t *testing.T) {
	t.Run("validate active session", func(t *testing.T) {
		now := time.Now()
		session := &Session{
			ID:         "session-123",
			UserID:     "user-456",
			ExpiresAt:  now.Add(1 * time.Hour),
			CreatedAt:  now,
			LastUsedAt: now,
			Revoked:    false,
		}

		assert.True(t, session.IsValid())
		assert.False(t, session.IsExpired())
	})
}

func TestAuthenticationDomain_ValidateSession_Expired(t *testing.T) {
	t.Run("reject expired session", func(t *testing.T) {
		now := time.Now()
		session := &Session{
			ID:         "session-123",
			UserID:     "user-456",
			ExpiresAt:  now.Add(-1 * time.Hour), // Expired 1 hour ago
			CreatedAt:  now.Add(-2 * time.Hour),
			LastUsedAt: now.Add(-1 * time.Hour),
			Revoked:    false,
		}

		assert.False(t, session.IsValid())
		assert.True(t, session.IsExpired())
	})
}

func TestAuthenticationDomain_RevokeSession(t *testing.T) {
	t.Run("revoke active session", func(t *testing.T) {
		now := time.Now()
		session := &Session{
			ID:         "session-123",
			UserID:     "user-456",
			ExpiresAt:  now.Add(1 * time.Hour),
			CreatedAt:  now,
			LastUsedAt: now,
			Revoked:    false,
		}

		// Session is initially valid
		assert.True(t, session.IsValid())

		// Revoke the session
		session.Revoked = true

		// Session should now be invalid
		assert.False(t, session.IsValid())
		assert.False(t, session.IsExpired()) // Still not expired, just revoked
	})
}

func TestAuthenticationDomain_RefreshTokenSelector(t *testing.T) {
	t.Run("create session with refresh token selector", func(t *testing.T) {
		now := time.Now()
		session := &Session{
			ID:                   "session-123",
			UserID:               "user-456",
			RefreshToken:         "hashed-verifier-789",
			RefreshTokenSelector: "selector-abc123",
			Salt:                 "salt-def456",
			DeviceID:             "device-xyz",
			IPAddress:            "192.168.1.1",
			UserAgent:            "Mozilla/5.0",
			ExpiresAt:            now.Add(30 * 24 * time.Hour), // 30 days
			CreatedAt:            now,
			LastUsedAt:           now,
			Revoked:              false,
		}

		// Test that RefreshTokenSelector field is properly set
		assert.Equal(t, "selector-abc123", session.RefreshTokenSelector)
		assert.Equal(t, "hashed-verifier-789", session.RefreshToken)
		assert.Equal(t, "salt-def456", session.Salt)
		assert.True(t, session.IsValid())
	})

	t.Run("validate refresh token selector format", func(t *testing.T) {
		tests := []struct {
			name     string
			selector string
			valid    bool
		}{
			{
				name:     "valid selector",
				selector: "abcd1234efgh5678",
				valid:    true,
			},
			{
				name:     "empty selector",
				selector: "",
				valid:    false,
			},
			{
				name:     "too short selector",
				selector: "short",
				valid:    false,
			},
			{
				name:     "valid long selector",
				selector: "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ01",
				valid:    true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				now := time.Now()
				session := &Session{
					ID:                   "session-123",
					UserID:               "user-456",
					RefreshToken:         "hashed-verifier",
					RefreshTokenSelector: tt.selector,
					Salt:                 "salt",
					ExpiresAt:            now.Add(1 * time.Hour),
					CreatedAt:            now,
					LastUsedAt:           now,
					Revoked:              false,
				}

				err := session.ValidateRefreshTokenSelector()
				if tt.valid {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}
