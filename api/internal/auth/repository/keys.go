package repository

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"

	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
)

type keyRepository struct {
	mu         sync.RWMutex
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
}

// NewKeyRepository creates a new key repository
func NewKeyRepository() (domain.KeyRepository, error) {
	r := &keyRepository{}
	if err := r.generateKeys(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *keyRepository) GetPrivateKey() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.privateKey == nil {
		return nil, fmt.Errorf("private key not initialized")
	}

	// Encode private key to PEM
	privKeyBytes := x509.MarshalPKCS1PrivateKey(r.privateKey)
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return privKeyPEM, nil
}

func (r *keyRepository) GetPublicKey() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.publicKey == nil {
		return nil, fmt.Errorf("public key not initialized")
	}

	// Encode public key to PEM
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(r.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return pubKeyPEM, nil
}

func (r *keyRepository) GetJWKS() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.publicKey == nil {
		return nil, fmt.Errorf("public key not initialized")
	}

	// Create JWK
	n := base64.RawURLEncoding.EncodeToString(r.publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(r.publicKey.E)).Bytes())

	jwk := map[string]interface{}{
		"kty": "RSA",
		"use": "sig",
		"kid": r.keyID,
		"alg": "RS256",
		"n":   n,
		"e":   e,
	}

	// Create JWKS
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{jwk},
	}

	return json.Marshal(jwks)
}

func (r *keyRepository) RotateKeys() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.generateKeys()
}

func (r *keyRepository) generateKeys() error {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Generate key ID
	keyIDBytes := make([]byte, 16)
	if _, err := rand.Read(keyIDBytes); err != nil {
		return fmt.Errorf("failed to generate key ID: %w", err)
	}

	r.privateKey = privateKey
	r.publicKey = &privateKey.PublicKey
	r.keyID = base64.RawURLEncoding.EncodeToString(keyIDBytes)

	return nil
}