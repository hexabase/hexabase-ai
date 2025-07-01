package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	internalAuth "github.com/hexabase/hexabase-ai/api/internal/auth"
	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
)

// TestRefreshTokenInvalidatesOldAccessToken tests that after refreshing tokens,
// the old access token should become invalid (issue #232)
func TestRefreshTokenInvalidatesOldAccessToken(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := new(mockRepository)
	mockKeyRepo := new(mockKeyRepository)
	mockTokenDomainService := new(mockTokenDomainService)

	// Create test RSA key pair
	testPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	testPublicKey := &testPrivateKey.PublicKey
	tokenManager := internalAuth.NewTokenManager(testPrivateKey, testPublicKey, "test-issuer", 15*time.Minute)

	// Create service
	svc := &service{
		repo:               mockRepo,
		keyRepo:            mockKeyRepo,
		tokenManager:       tokenManager,
		tokenDomainService: mockTokenDomainService,
		logger:             slog.Default(),
		defaultTokenExpiry: 900, // 15 minutes
	}

	// Create test user
	user := &domain.User{
		ID:          "user-123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Provider:    "github",
	}

	// Create initial session
	sessionID := uuid.New().String()
	hashedToken := "b8c8f5e6d4a7c2e9f1b3d6e8a5c7f9e2d4b6e8f1c3e5a7b9d2f4e6c8a1b3d5e7"
	salt := "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678"

	session := &domain.Session{
		ID:           sessionID,
		UserID:       user.ID,
		RefreshToken: hashedToken,
		Salt:         salt,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
	}

	// Step 1: Login and get initial tokens
	mockRepo.On("GetUserOrganizations", ctx, user.ID).Return([]string{"org-1"}, nil).Once()

	// Generate initial token pair
	initialTokenPair, err := svc.generateTokenPair(ctx, user, sessionID)
	assert.NoError(t, err)
	assert.NotEmpty(t, initialTokenPair.AccessToken)
	assert.NotEmpty(t, initialTokenPair.RefreshToken)

	// Get public key for token validation
	publicKeyDER, _ := x509.MarshalPKIXPublicKey(testPublicKey)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})
	mockKeyRepo.On("GetPublicKey").Return(publicKeyPEM, nil).Maybe()

	// Step 2: Verify initial access token is valid
	mockRepo.On("IsSessionBlocked", ctx, sessionID).Return(false, nil).Once()

	initialClaims, err := svc.ValidateAccessToken(ctx, initialTokenPair.AccessToken)
	assert.NoError(t, err)
	assert.NotNil(t, initialClaims)
	assert.Equal(t, user.ID, initialClaims.UserID)
	assert.Equal(t, sessionID, initialClaims.SessionID)

	// Step 3: Refresh tokens
	refreshToken := initialTokenPair.RefreshToken

	// Mock for refresh token flow
	mockRepo.On("IsRefreshTokenBlacklisted", ctx, refreshToken).Return(false, nil).Once()
	mockRepo.On("GetAllActiveSessions", ctx).Return([]*domain.Session{session}, nil).Once()
	mockRepo.On("VerifyToken", refreshToken, hashedToken, salt).Return(true).Once()
	mockRepo.On("GetUser", ctx, user.ID).Return(user, nil).Once()

	// Domain service returns claims with the original session ID
	domainClaims := &domain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.DisplayName,
		Provider:  user.Provider,
		SessionID: sessionID, // Original session ID from domain service
		OrgIDs:    []string{"org-1"},
	}

	mockTokenDomainService.On("RefreshToken", ctx, session, user).Return(domainClaims, nil).Once()
	mockRepo.On("BlacklistRefreshToken", ctx, refreshToken, session.ExpiresAt).Return(nil).Once()

	// Hash new refresh token
	mockRepo.On("HashToken", mock.AnythingOfType("string")).Return(
		"9876543210987654321098765432109876543210987654321098765432109876",
		"efghefghefghefghefghefghefghefghefghefghefghefghefghefghefghefgh",
		nil).Once()

	// Block the old session ID
	mockRepo.On("BlockSession", ctx, sessionID, mock.AnythingOfType("time.Time")).Return(nil).Once()

	// Create new session
	mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*domain.Session")).Return(nil).Once()

	// Delete old session
	mockRepo.On("DeleteSession", ctx, sessionID).Return(nil).Once()

	// Create security event
	mockRepo.On("CreateSecurityEvent", ctx, mock.AnythingOfType("*domain.SecurityEvent")).Return(nil).Once()

	// Perform refresh
	newTokenPair, err := svc.RefreshToken(ctx, refreshToken, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	assert.NotNil(t, newTokenPair)
	assert.NotEmpty(t, newTokenPair.AccessToken)
	assert.NotEmpty(t, newTokenPair.RefreshToken)
	assert.NotEqual(t, initialTokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, initialTokenPair.RefreshToken, newTokenPair.RefreshToken)

	// Step 4: Verify old access token is now invalid
	// The old session should be blocked, so this should fail
	mockRepo.On("IsSessionBlocked", ctx, sessionID).Return(true, nil).Once()

	oldTokenClaims, err := svc.ValidateAccessToken(ctx, initialTokenPair.AccessToken)
	assert.Error(t, err, "Old access token should be invalid after refresh")
	assert.Nil(t, oldTokenClaims)
	assert.Contains(t, err.Error(), "session has been invalidated")

	// Step 5: Verify new access token is valid
	// The new session ID is dynamically generated, so we use AnythingOfType
	mockRepo.On("IsSessionBlocked", ctx, mock.AnythingOfType("string")).Return(false, nil).Once()

	newTokenClaims, err := svc.ValidateAccessToken(ctx, newTokenPair.AccessToken)
	assert.NoError(t, err)
	assert.NotNil(t, newTokenClaims)
	assert.Equal(t, user.ID, newTokenClaims.UserID)
	// Verify that the new session ID is different from the old one
	assert.NotEqual(t, sessionID, newTokenClaims.SessionID)

	// Verify all expectations were met
	mockRepo.AssertExpectations(t)
	mockKeyRepo.AssertExpectations(t)
	mockTokenDomainService.AssertExpectations(t)
	mockKeyRepo.AssertCalled(t, "GetPublicKey")
}

