package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenHashRepository_HashToken(t *testing.T) {
	repo := NewTokenHashRepository()

	t.Run("Successfully generate hash and salt", func(t *testing.T) {
		token := "test-repository-token-123"

		hashedToken, salt, err := repo.HashToken(token)

		require.NoError(t, err)
		assert.NotEqual(t, token, hashedToken)
		assert.NotEmpty(t, salt)
		assert.Len(t, hashedToken, 64) // SHA-256 hash as hex = 64 chars
		assert.Len(t, salt, 64)         // 32-byte salt as hex = 64 chars
	})

	t.Run("Handle empty token at repository level (pure crypto)", func(t *testing.T) {
		hashedToken, salt, err := repo.HashToken("")

		require.NoError(t, err) // Repository doesn't validate business rules
		assert.NotEmpty(t, hashedToken)
		assert.NotEmpty(t, salt)
		assert.Len(t, hashedToken, 64) // SHA-256 hash as hex = 64 chars
		assert.Len(t, salt, 64)         // 32-byte salt as hex = 64 chars
	})

	t.Run("Generate unique salts for same token", func(t *testing.T) {
		token := "same-token-for-uniqueness-test"

		hash1, salt1, err1 := repo.HashToken(token)
		hash2, salt2, err2 := repo.HashToken(token)

		require.NoError(t, err1)
		require.NoError(t, err2)

		// Same token with different salts should produce different hashes
		assert.NotEqual(t, hash1, hash2)
		assert.NotEqual(t, salt1, salt2)
	})

	t.Run("Hash format validation", func(t *testing.T) {
		token := "format-validation-token"

		hashedToken, salt, err := repo.HashToken(token)
		require.NoError(t, err)

		// Verify hex encoding
		assert.Regexp(t, "^[a-f0-9]{64}$", hashedToken)
		assert.Regexp(t, "^[a-f0-9]{64}$", salt)
	})
}

func TestTokenHashRepository_VerifyToken(t *testing.T) {
	repo := NewTokenHashRepository()

	t.Run("Successfully verify matching token", func(t *testing.T) {
		plainToken := "verify-repository-test"

		hashedToken, salt, err := repo.HashToken(plainToken)
		require.NoError(t, err)

		isValid := repo.VerifyToken(plainToken, hashedToken, salt)
		assert.True(t, isValid)
	})

	t.Run("Reject mismatched token", func(t *testing.T) {
		originalToken := "original-token"
		wrongToken := "wrong-token"

		hashedToken, salt, err := repo.HashToken(originalToken)
		require.NoError(t, err)

		isValid := repo.VerifyToken(wrongToken, hashedToken, salt)
		assert.False(t, isValid)
	})

	t.Run("Handle invalid salt gracefully", func(t *testing.T) {
		plainToken := "test-token"
		hashedToken := "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234"
		invalidSalt := "invalid-hex-salt"

		isValid := repo.VerifyToken(plainToken, hashedToken, invalidSalt)
		assert.False(t, isValid)
	})

	t.Run("Constant-time comparison security", func(t *testing.T) {
		plainToken := "security-test-token"

		hashedToken, salt, err := repo.HashToken(plainToken)
		require.NoError(t, err)

		// Test with completely different token - should still use constant time
		wrongToken := "completely-different-token-with-different-length"
		isValid := repo.VerifyToken(wrongToken, hashedToken, salt)
		assert.False(t, isValid)
	})
}

func TestTokenHashRepository_CryptographicProperties(t *testing.T) {
	repo := NewTokenHashRepository()

	t.Run("Avalanche effect - small change produces very different hash", func(t *testing.T) {
		token1 := "avalanche-test-token-1"
		token2 := "avalanche-test-token-2" // Only last character differs

		// Generate hashes for both tokens
		hash1, _, err1 := repo.HashToken(token1)
		require.NoError(t, err1)

		hash2, _, err2 := repo.HashToken(token2)
		require.NoError(t, err2)

		// Even with similar tokens, hashes should be very different
		assert.NotEqual(t, hash1, hash2)

		// Count differing characters (should be many for good avalanche)
		differingChars := 0
		for i := range hash1 {
			if hash1[i] != hash2[i] {
				differingChars++
			}
		}

		// With good avalanche effect, at least 50% of characters should differ
		assert.Greater(t, differingChars, 32, "Poor avalanche effect detected")
	})

	t.Run("Salt entropy validation", func(t *testing.T) {
		// Generate multiple salts and verify they're different
		const numSalts = 100
		salts := make(map[string]bool)

		for i := 0; i < numSalts; i++ {
			_, salt, err := repo.HashToken("entropy-test-token")
			require.NoError(t, err)

			// Check for duplicates (should be extremely rare)
			assert.False(t, salts[salt], "Duplicate salt detected: %s", salt)
			salts[salt] = true
		}

		// Verify all salts are unique
		assert.Len(t, salts, numSalts)
	})

	t.Run("Deterministic verification", func(t *testing.T) {
		plainToken := "deterministic-test"

		hashedToken, salt, err := repo.HashToken(plainToken)
		require.NoError(t, err)

		// Verify multiple times - should always return same result
		for i := 0; i < 10; i++ {
			isValid := repo.VerifyToken(plainToken, hashedToken, salt)
			assert.True(t, isValid, "Verification failed on iteration %d", i)
		}
	})
}