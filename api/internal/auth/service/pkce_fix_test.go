package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"log/slog"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestPKCEVerificationRFC7636Compliance tests PKCE implementation compliance with RFC 7636
func TestPKCEVerificationRFC7636Compliance(t *testing.T) {
	ctx := context.Background()

	// Test vectors from RFC 7636 Appendix B
	// https://datatracker.ietf.org/doc/html/rfc7636#appendix-B
	testCases := []struct {
		name              string
		codeVerifier      string
		expectedChallenge string // RFC 7636 compliant challenge (Base64URL without padding)
		description       string
	}{
		{
			name:              "RFC 7636 Appendix B test vector",
			codeVerifier:      "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			expectedChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			description:       "Official test vector from RFC 7636",
		},
		{
			name:              "Custom test case 1",
			codeVerifier:      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789",
			expectedChallenge: "20v8vU2gzYWmDDw30_vYgFx38V_Gsf3-YU7gp8j9tMA",
			description:       "Alphanumeric code verifier",
		},
		{
			name:              "Minimum length verifier (43 chars)",
			codeVerifier:      "1234567890123456789012345678901234567890123",
			expectedChallenge: "WWHTYIjNclXxS69q1gerQ-eTlW5ab1YCpKTorurQ3zw",
			description:       "Minimum allowed length per RFC 7636",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			
			svc := &service{
				repo:   mockRepo,
				logger: slog.Default(),
			}

			// Verify our expected challenge calculation is correct
			h := sha256.New()
			h.Write([]byte(tc.codeVerifier))
			calculatedChallenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
			assert.Equal(t, tc.expectedChallenge, calculatedChallenge, 
				"Test case challenge calculation should match expected")

			// Test successful PKCE verification
			authState := &domain.AuthState{
				State:         "test-state",
				CodeChallenge: tc.expectedChallenge, // Store the RFC-compliant challenge
				ExpiresAt:     time.Now().Add(10 * time.Minute),
			}

			mockRepo.On("GetAuthState", ctx, "test-state").Return(authState, nil).Once()

			err := svc.VerifyPKCE(ctx, "test-state", tc.codeVerifier)
			assert.NoError(t, err, "PKCE verification should succeed with RFC-compliant implementation")

			// Test failed verification with incorrect verifier
			mockRepo.On("GetAuthState", ctx, "test-state").Return(authState, nil).Once()

			err = svc.VerifyPKCE(ctx, "test-state", "incorrect-verifier")
			assert.Error(t, err, "PKCE verification should fail with incorrect verifier")
			assert.Contains(t, err.Error(), "PKCE verification failed")

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPKCEEncodingDifferences demonstrates the differences between encoding methods
func TestPKCEEncodingDifferences(t *testing.T) {
	codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	hashValue := h.Sum(nil)
	
	// Different encoding methods
	standardBase64 := base64.StdEncoding.EncodeToString(hashValue)
	urlBase64WithPadding := base64.URLEncoding.EncodeToString(hashValue)
	urlBase64NoPadding := base64.RawURLEncoding.EncodeToString(hashValue)
	
	// Code Verifier: dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk
	// Standard Base64: <see assertion below>
	// URL Base64 (with padding): <see assertion below>
	// URL Base64 (no padding) - RFC 7636 compliant: <see assertion below>
	
	// RFC 7636 requires RawURLEncoding (no padding)
	assert.Equal(t, "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", urlBase64NoPadding)
	
	// Show the differences
	// Standard Base64 uses + and / characters (this specific example has +)
	assert.Contains(t, standardBase64, "+", "Standard Base64 uses + character")
	assert.Contains(t, urlBase64WithPadding, "=", "URL Base64 with padding contains =")
	assert.NotContains(t, urlBase64NoPadding, "=", "Raw URL Base64 has no padding")
	assert.NotContains(t, urlBase64NoPadding, "+", "Raw URL Base64 uses - instead of +")
	assert.NotContains(t, urlBase64NoPadding, "/", "Raw URL Base64 uses _ instead of /")
	
	// The key difference for RFC 7636 compliance
	assert.NotEqual(t, urlBase64WithPadding, urlBase64NoPadding, "Padded and unpadded versions differ")
}

// TestPKCEStateManagement tests the proper storage and retrieval of PKCE parameters
func TestPKCEStateManagement(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(mockRepository)
	mockOAuthRepo := new(mockOAuthRepository)

	svc := &service{
		repo:      mockRepo,
		oauthRepo: mockOAuthRepo,
		logger:    slog.Default(),
	}

	t.Run("PKCE code challenge stored in auth state", func(t *testing.T) {
		codeChallenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
		
		req := &domain.LoginRequest{
			Provider:            "google",
			CodeChallenge:       codeChallenge,
			CodeChallengeMethod: "S256",
		}

		var capturedAuthState *domain.AuthState
		mockRepo.On("StoreAuthState", ctx, mock.AnythingOfType("*domain.AuthState")).
			Run(func(args mock.Arguments) {
				capturedAuthState = args.Get(1).(*domain.AuthState)
			}).
			Return(nil)

		expectedParams := map[string]string{
			"code_challenge":        codeChallenge,
			"code_challenge_method": "S256",
		}
		mockOAuthRepo.On("GetAuthURL", "google", mock.AnythingOfType("string"), expectedParams).
			Return("https://accounts.google.com/o/oauth2/v2/auth", nil)

		_, _, err := svc.GetAuthURL(ctx, req)
		assert.NoError(t, err)

		// The CodeChallenge field stores the code challenge for later verification
		assert.Equal(t, codeChallenge, capturedAuthState.CodeChallenge, 
			"Code challenge should be stored in AuthState for later verification")

		mockRepo.AssertExpectations(t)
		mockOAuthRepo.AssertExpectations(t)
	})
}

// TestPKCESecurityProperties tests security properties of PKCE implementation
func TestPKCESecurityProperties(t *testing.T) {
	// Test that code challenge cannot be reversed to get code verifier
	codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	challenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	
	// SHA256 is a one-way function
	assert.NotEqual(t, codeVerifier, challenge, "Challenge should not equal verifier")
	assert.Equal(t, 43, len(challenge), "SHA256 Base64URL encoded should be 43 chars (256 bits / 6 bits per char, no padding)")
	
	// Different verifiers should produce different challenges
	h2 := sha256.New()
	h2.Write([]byte("different-verifier"))
	challenge2 := base64.RawURLEncoding.EncodeToString(h2.Sum(nil))
	
	assert.NotEqual(t, challenge, challenge2, "Different verifiers should produce different challenges")
}