package repository

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// tokenHashRepository handles secure token hashing operations
// This is infrastructure layer code for auth-specific crypto operations
type tokenHashRepository struct{}

// NewTokenHashRepository creates a new instance of the token hashing repository
func NewTokenHashRepository() *tokenHashRepository {
	return &tokenHashRepository{}
}

// HashToken securely hashes a token with a unique salt using SHA-256
// Pure crypto implementation without business validation
func (r *tokenHashRepository) HashToken(token string) (hashedToken string, salt string, err error) {

	// Generate cryptographically secure 32-byte salt
	saltBytes := make([]byte, 32)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Create hash: SHA256(token + salt)
	hasher := sha256.New()
	hasher.Write([]byte(token))
	hasher.Write(saltBytes)
	hashBytes := hasher.Sum(nil)

	return hex.EncodeToString(hashBytes), hex.EncodeToString(saltBytes), nil
}

// VerifyToken verifies if a plain token matches the stored hash using the provided salt
// Pure crypto implementation without business validation
func (r *tokenHashRepository) VerifyToken(plainToken, hashedToken, salt string) bool {

	// Decode salt from hex
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return false
	}

	// Compute hash using same method as HashToken
	hasher := sha256.New()
	hasher.Write([]byte(plainToken))
	hasher.Write(saltBytes)
	computedHashBytes := hasher.Sum(nil)
	computedHash := hex.EncodeToString(computedHashBytes)

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(hashedToken)) == 1
}