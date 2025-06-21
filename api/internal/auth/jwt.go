package auth

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
)

// TokenManager handles JWT token generation and validation
type TokenManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	expiration time.Duration
}

// Claims represents JWT claims for Hexabase
type Claims struct {
	jwt.RegisteredClaims
	UserID    string   `json:"user_id"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	OrgIDs    []string `json:"org_ids,omitempty"`
}

// WorkspaceClaims represents JWT claims for vCluster access
type WorkspaceClaims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"user_id"`
	WorkspaceID string   `json:"workspace_id"`
	Groups      []string `json:"groups"`
}

// NewTokenManager creates a new token manager
func NewTokenManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, issuer string, expiration time.Duration) *TokenManager {
	return &TokenManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
		expiration: expiration,
	}
}

// GenerateToken generates a JWT token for a user
func (tm *TokenManager) GenerateToken(userID, email, name string, orgIDs []string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tm.issuer,
			Subject:   userID,
			Audience:  []string{"hexabase-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.expiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		UserID: userID,
		Email:  email,
		Name:   name,
		OrgIDs: orgIDs,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(tm.privateKey)
}

// GenerateWorkspaceToken generates a JWT token for vCluster access
func (tm *TokenManager) GenerateWorkspaceToken(userID, workspaceID string, groups []string) (string, error) {
	now := time.Now()
	claims := WorkspaceClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tm.issuer,
			Subject:   userID,
			Audience:  []string{fmt.Sprintf("workspace-%s", workspaceID)},
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // 24 hour expiration for kubectl
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		UserID:      userID,
		WorkspaceID: workspaceID,
		Groups:      groups,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(tm.privateKey)
}

// ValidateToken validates a JWT token
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateWorkspaceToken validates a workspace JWT token
func (tm *TokenManager) ValidateWorkspaceToken(tokenString string) (*WorkspaceClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &WorkspaceClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*WorkspaceClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetPublicKey returns the public key in PEM format
func (tm *TokenManager) GetPublicKey() *rsa.PublicKey {
	return tm.publicKey
}

// SignClaims signs claims to create a JWT token
// This method is pure infrastructure - no business logic
func (tm *TokenManager) SignClaims(claims *domain.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(tm.privateKey)
}