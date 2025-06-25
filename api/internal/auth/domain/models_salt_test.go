package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionSaltSupport(t *testing.T) {
	t.Run("Session domain model includes salt field for refresh token storage", func(t *testing.T) {
		// Given
		session := Session{
			ID:           "sess-123",
			UserID:       "user-456",
			RefreshToken: "hashed-token-value",
			Salt:         "generated-salt-value",
			IPAddress:    "192.168.1.1",
			UserAgent:    "test-agent",
		}

		// When/Then - Salt field should be accessible
		assert.Equal(t, "generated-salt-value", session.Salt)
		assert.NotEmpty(t, session.Salt)
	})
}