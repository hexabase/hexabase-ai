package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenHelpers(t *testing.T) {
	s := &service{}

	t.Run("parseRefreshToken should parse valid token", func(t *testing.T) {
		token := "selector123.verifier456"

		parts, err := s.parseRefreshToken(token)
		require.NoError(t, err)
		assert.Equal(t, "selector123", parts.Selector)
		assert.Equal(t, "verifier456", parts.Verifier)
	})

	t.Run("parseRefreshToken should reject invalid format", func(t *testing.T) {
		invalidTokens := []string{
			"invalid",
			"selector.verifier.extra",
			"",
			"selector.",
			".verifier",
		}

		for _, token := range invalidTokens {
			_, err := s.parseRefreshToken(token)
			assert.Error(t, err, "should reject token: %s", token)
			assert.Contains(t, err.Error(), "invalid refresh token format")
		}
	})

	t.Run("parseRefreshTokenWithLegacySupport should handle valid format", func(t *testing.T) {
		token := "selector123.verifier456"

		parts, err := s.parseRefreshToken(token)
		require.NoError(t, err)
		assert.Equal(t, "selector123", parts.Selector)
		assert.Equal(t, "verifier456", parts.Verifier)
	})

	t.Run("parseRefreshTokenWithLegacySupport should handle legacy format", func(t *testing.T) {
		legacyToken := "legacy-token-without-separator"

		parts, err := s.parseRefreshToken(legacyToken)
		require.Error(t, err)
		assert.Nil(t, parts, "should return nil parts for legacy format")
	})

	t.Run("buildRefreshToken should combine selector and verifier", func(t *testing.T) {
		selector := "selector123"
		verifier := "verifier456"

		token := s.buildRefreshToken(selector, verifier)
		assert.Equal(t, "selector123.verifier456", token)
	})

	t.Run("round trip should work correctly", func(t *testing.T) {
		originalSelector := "test-selector"
		originalVerifier := "test-verifier"

		// Build token
		token := s.buildRefreshToken(originalSelector, originalVerifier)

		// Parse token
		parts, err := s.parseRefreshToken(token)
		require.NoError(t, err)

		// Verify round trip
		assert.Equal(t, originalSelector, parts.Selector)
		assert.Equal(t, originalVerifier, parts.Verifier)
	})

	t.Run("constants should be used correctly", func(t *testing.T) {
		// Verify constants are defined correctly
		assert.Equal(t, ".", refreshTokenSeparator)
		assert.Equal(t, 2, refreshTokenExpectedParts)

		// Verify error message uses constant
		_, err := s.parseRefreshToken("invalid")
		assert.Contains(t, err.Error(), refreshTokenSeparator)
	})
}
