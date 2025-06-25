package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// NOTE: Session blocklist tests are temporarily disabled
// Integration tests require Redis connection and would be better implemented with testcontainers

// TestSessionBlocklistRepository_KeyFormat tests Redis key formatting
func TestSessionBlocklistRepository_KeyFormat(t *testing.T) {
	// Test that we generate proper Redis keys
	sessionID := "session-123"
	expectedKey := "session_blocklist:session-123"

	key := formatSessionBlocklistKey(sessionID)
	assert.Equal(t, expectedKey, key)
}


