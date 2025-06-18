package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenDomainService_RefreshToken(t *testing.T) {
	t.Run("successful token refresh with valid session", func(t *testing.T) {
		// Arrange
		session := &Session{
			ID:           "session-123",
			UserID:       "user-123",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now().Add(-1 * time.Hour),
			Revoked:      false,
		}

		user := &User{
			ID:          "user-123",
			Email:       "test@example.com",
			DisplayName: "Test User",
			Provider:    "google",
		}

		ctx := context.Background()
		service := NewTokenDomainService()

		// Act
		result, err := service.RefreshToken(ctx, session, user)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.ID, result.UserID)
		assert.Equal(t, user.Email, result.Email)
		assert.NotEmpty(t, result.SessionID)
	})

	t.Run("reject refresh for expired session", func(t *testing.T) {
		// Arrange
		session := &Session{
			ID:           "session-123",
			UserID:       "user-123",
			RefreshToken: "expired-refresh-token",
			ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
			CreatedAt:    time.Now().Add(-25 * time.Hour),
			Revoked:      false,
		}

		user := &User{
			ID:    "user-123",
			Email: "test@example.com",
		}

		ctx := context.Background()
		service := NewTokenDomainService()

		// Act
		result, err := service.RefreshToken(ctx, session, user)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "session has expired")
	})

	t.Run("reject refresh for revoked session", func(t *testing.T) {
		// Arrange
		session := &Session{
			ID:           "session-123",
			UserID:       "user-123",
			RefreshToken: "revoked-refresh-token",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now().Add(-1 * time.Hour),
			Revoked:      true, // Revoked
		}

		user := &User{
			ID:    "user-123",
			Email: "test@example.com",
		}

		ctx := context.Background()
		service := NewTokenDomainService()

		// Act
		result, err := service.RefreshToken(ctx, session, user)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "session is revoked")
	})
}

func TestTokenDomainService_ValidateRefreshEligibility(t *testing.T) {
	t.Run("allow refresh for valid session", func(t *testing.T) {
		session := &Session{
			ExpiresAt: time.Now().Add(12 * time.Hour),
			CreatedAt: time.Now().Add(-1 * time.Hour),
			Revoked:   false,
		}

		service := NewTokenDomainService()
		err := service.ValidateRefreshEligibility(session)
		assert.NoError(t, err)
	})

	t.Run("reject refresh for expired session", func(t *testing.T) {
		session := &Session{
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
			CreatedAt: time.Now().Add(-2 * time.Hour),
			Revoked:   false,
		}

		service := NewTokenDomainService()
		err := service.ValidateRefreshEligibility(session)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session has expired")
	})
} 