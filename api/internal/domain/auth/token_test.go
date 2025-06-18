package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaims_ValidateRequired(t *testing.T) {
	t.Run("ensure required claims are present", func(t *testing.T) {
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-123",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID:    "user-123",
			Email:     "test@example.com",
			Name:      "Test User",
			Provider:  "google",
			SessionID: "session-456",
		}

		err := claims.ValidateBusinessRules()
		assert.NoError(t, err)
	})
}

func TestClaims_ValidateRequired_MissingFields(t *testing.T) {
	tests := []struct {
		name        string
		claims      *Claims
		expectedErr string
	}{
		{
			name: "missing user_id",
			claims: &Claims{
				Email:     "test@example.com",
				Name:      "Test User",
				SessionID: "session-456",
			},
			expectedErr: "user_id claim is required",
		},
		{
			name: "missing email",
			claims: &Claims{
				UserID:    "user-123",
				Name:      "Test User", 
				SessionID: "session-456",
			},
			expectedErr: "email claim is required",
		},
		{
			name: "missing session_id",
			claims: &Claims{
				UserID: "user-123",
				Email:  "test@example.com",
				Name:   "Test User",
			},
			expectedErr: "session_id claim is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claims.ValidateBusinessRules()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestClaims_ValidateExpiration(t *testing.T) {
	t.Run("check token expiration logic", func(t *testing.T) {
		now := time.Now()
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-123",
				ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
			UserID:    "user-123",
			Email:     "test@example.com",
			Name:      "Test User",
			SessionID: "session-456",
		}

		err := claims.ValidateBusinessRules()
		assert.NoError(t, err)
	})

	t.Run("reject token with excessive expiration", func(t *testing.T) {
		now := time.Now()
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-123",
				ExpiresAt: jwt.NewNumericDate(now.Add(25 * time.Hour)), // Exceeds 24 hour limit
				IssuedAt:  jwt.NewNumericDate(now),
			},
			UserID:    "user-123",
			Email:     "test@example.com",
			Name:      "Test User",
			SessionID: "session-456",
		}

		err := claims.ValidateBusinessRules()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token expiry exceeds maximum allowed duration")
	})
}

func TestClaims_ValidateAudience(t *testing.T) {
	t.Run("verify audience claim validation", func(t *testing.T) {
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:  "user-123",
				Audience: jwt.ClaimStrings{"hexabase-api"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID:    "user-123",
			Email:     "test@example.com",
			Name:      "Test User",
			SessionID: "session-456",
		}

		err := claims.ValidateBusinessRules()
		assert.NoError(t, err)
		assert.Contains(t, claims.Audience, "hexabase-api")
	})
}

func TestClaims_ValidateIssuer(t *testing.T) {
	t.Run("verify issuer claim validation", func(t *testing.T) {
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:  "user-123",
				Issuer:   "https://api.hexabase-kaas.io",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID:    "user-123",
			Email:     "test@example.com",
			Name:      "Test User",
			SessionID: "session-456",
		}

		err := claims.ValidateBusinessRules()
		assert.NoError(t, err)
		assert.Equal(t, "https://api.hexabase-kaas.io", claims.Issuer)
	})
} 