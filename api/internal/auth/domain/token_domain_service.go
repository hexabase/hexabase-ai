package domain

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenDomainService implements business rules for token management
type tokenDomainService struct {
}

// NewTokenDomainService creates a new token domain service
func NewTokenDomainService() TokenDomainService {
	return &tokenDomainService{}
}

// RefreshToken generates new claims based on session and user data
// This method contains ONLY business logic, no infrastructure concerns
func (s *tokenDomainService) RefreshToken(ctx context.Context, session *Session, user *User) (*Claims, error) {
	// Apply business rules for refresh eligibility
	if err := s.ValidateRefreshEligibility(session); err != nil {
		return nil, err
	}

	// Generate new claims with business-appropriate values
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)), // 1 hour access token
		},
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.DisplayName,
		Provider:  user.Provider,
		SessionID: session.ID,
	}

	// Validate the generated claims meet business requirements
	if err := claims.ValidateBusinessRules(); err != nil {
		return nil, err
	}

	return claims, nil
}

// ValidateRefreshEligibility checks if a session is eligible for token refresh
func (s *tokenDomainService) ValidateRefreshEligibility(session *Session) error {
	// Check if session has expired
	if session.IsExpired() {
		return errors.New("session has expired")
	}

	// Check if session is revoked
	if session.Revoked {
		return errors.New("session is revoked")
	}

	// Note: Removed hardcoded 7-day limit - session validity is determined
	// by the session's own ExpiresAt field which is set by business policy
	// when the session is created (typically 30 days)
	
	return nil
}

// CreateSession creates a new session with proper business rules
func (s *tokenDomainService) CreateSession(sessionID, userID, refreshToken, deviceID, clientIP, userAgent string) (*Session, error) {
	now := time.Now()
	
	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		RefreshToken: refreshToken,
		DeviceID:     deviceID,
		IPAddress:    clientIP,
		UserAgent:    userAgent,
		ExpiresAt:    now.Add(30 * 24 * time.Hour), // 30 days refresh token expiry
		CreatedAt:    now,
		LastUsedAt:   now,
		Revoked:      false,
	}

	return session, nil
}

// ValidateTokenClaims validates token claims against business rules
func (s *tokenDomainService) ValidateTokenClaims(claims *Claims) error {
	if claims == nil {
		return errors.New("claims cannot be nil")
	}

	// Delegate to Claims validation
	return claims.ValidateBusinessRules()
}

// ShouldRefreshToken determines if a token should be refreshed based on business rules
func (s *tokenDomainService) ShouldRefreshToken(claims *Claims) bool {
	if claims == nil || claims.ExpiresAt == nil {
		return false
	}

	// Refresh if token expires within next 15 minutes
	refreshThreshold := time.Now().Add(15 * time.Minute)
	return claims.ExpiresAt.Before(refreshThreshold)
} 