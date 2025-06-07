package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hexabase/hexabase-ai/api/internal/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockSecureRedisClient for testing enhanced session management
type MockSecureRedisClient struct {
	mock.Mock
	storage map[string]interface{}
}

func NewMockSecureRedisClient() *MockSecureRedisClient {
	return &MockSecureRedisClient{
		storage: make(map[string]interface{}),
	}
}

func (m *MockSecureRedisClient) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	if args.Error(0) == nil {
		m.storage[key] = value
	}
	return args.Error(0)
}

func (m *MockSecureRedisClient) GetDel(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	if args.Error(1) == nil {
		if val, ok := m.storage[key]; ok {
			delete(m.storage, key)
			return val.(string), nil
		}
	}
	return args.String(0), args.Error(1)
}

func (m *MockSecureRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	if args.Error(1) == nil {
		if val, ok := m.storage[key]; ok {
			return val.(string), nil
		}
	}
	return args.String(0), args.Error(1)
}

func (m *MockSecureRedisClient) Delete(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	if args.Error(0) == nil {
		for _, key := range keys {
			delete(m.storage, key)
		}
	}
	return args.Error(0)
}

func (m *MockSecureRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

// OAuthSecurityTestSuite tests enhanced OAuth security features
type OAuthSecurityTestSuite struct {
	suite.Suite
	client      *SecureOAuthClient
	jwtManager  *EnhancedJWTManager
	redisClient *MockSecureRedisClient
	ctx         context.Context
}

func (s *OAuthSecurityTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.redisClient = NewMockSecureRedisClient()
	
	// Setup mock responses
	s.redisClient.On("SetWithTTL", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.redisClient.On("GetDel", mock.Anything, mock.Anything).Return("valid", nil)
	s.redisClient.On("Get", mock.Anything, mock.Anything).Return("", nil)
	s.redisClient.On("Delete", mock.Anything, mock.Anything).Return(nil)
	s.redisClient.On("Exists", mock.Anything, mock.Anything).Return(int64(0), nil)
	
	// Initialize JWT manager with test keys
	privateKey, publicKey := generateTestRSAKeys()
	s.jwtManager = NewEnhancedJWTManager(privateKey, publicKey, s.redisClient)
	
	// Initialize secure OAuth client with test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			ExternalProviders: map[string]config.OAuthProvider{
				"google": {
					ClientID:     "test-google-client",
					ClientSecret: "test-google-secret",
					RedirectURL:  "http://localhost:8080/auth/callback/google",
					Scopes:       []string{"email", "profile"},
				},
				"github": {
					ClientID:     "test-github-client",
					ClientSecret: "test-github-secret",
					RedirectURL:  "http://localhost:8080/auth/callback/github",
					Scopes:       []string{"user:email"},
				},
				"gitlab": {
					ClientID:     "test-gitlab-client",
					ClientSecret: "test-gitlab-secret",
					RedirectURL:  "http://localhost:8080/auth/callback/gitlab",
					Scopes:       []string{"read_user"},
				},
			},
		},
	}
	s.client = NewSecureOAuthClient(cfg, s.redisClient, privateKey, publicKey)
}

// generateTestRSAKeys generates RSA key pair for testing
func generateTestRSAKeys() (privateKey, publicKey interface{}) {
	// Use the same keys from existing tests
	privateKeyPEM := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0Z0VS5JJcds3xfNnvW0qgJt75j/i6xO1Zc3lR1eQZjzrP0a7
2RKKSXDbY1vPiPJxRBYJzPCBbgUGbhKKq8LvTFDKPWNBL9vC6hGRJxKGmPnV5bDg
wQcJ7io4pzWRRNxVPKmJquVhqwmOu0D4jiR8ggNcJj7lUFhXCbU1xPbpf/kXDk7B
LgAXdEPvrkvLN6Lzq7Ej0Y7sTcdQF3uWL4qwZ4OXRGaXTS0I8hP8BMwy6D4SGmqF
e7cCAwEAAQKCAQEAyHwCYPAMW7r1h2n2ypOmaMr6xnfZI8TrFJmOPdJIMSz1OFAD
+7Qht7V8YLnNgJhm5BpIzKjdZ7kLQMJKC5zK2JLQZ5L9S/ALCsLjs4gUNCQkU0nE
0GGN4D8EYQazgyJ0XyYLQslzMEGfZHATJBjd9jW5KGd6atQvBkDDKOn2lvFVMOGy
YB+rmuQDyIPSvTjrRGp0W1aaDTjTlEo9lWLUcj3gb2gAB6CUBxKBquKzaq3f5biG
YmbdMnCcBIdCkD9noTtNKL8TLPjNORG8b8A9G6cGheRGnIjBYLDpFGLcABHKuJoy
JQKBgQDhZnnZnz1qk1xBinvI8hu2FHRMx6ZqMXAGKMVQvJFQn8K/T9p6z8rxm7wR
0B+7ND1k8SSpwg9HiwqAIQKBgQDsR3v3BnN2urQ/I7kCOF1CTF4ePfFvvu5PXUNQ
2qT1qQvLDQdnUdMKHXRo5g==
-----END RSA PRIVATE KEY-----`

	publicKeyPEM := `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA0Z0VS5JJcds3xfNnvW0qgJt75j/i6xO1Zc3lR1eQZjzrP0a7
2RKKSXDbY1vPiPJxRBYJzPCBbgUGbhKKq8LvTFDKPWNBL9vC6hGRJxKGmPnV5bDg
wQcJ7io4pzWRRNxVPKmJquVhqwmOu0D4jiR8ggNcJj7lUFhXCbU1xPbpf/kXDk7B
LgAXdEPvrkvLN6Lzq7Ej0Y7sTcdQF3uWL4qwZ4OXRGaXTS0I8hP8BMwy6D4SGmqF
e7cCAwEAAQ==
-----END RSA PUBLIC KEY-----`

	priv, _ := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	pub, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
	
	return priv, pub
}

// Test PKCE flow implementation
func (s *OAuthSecurityTestSuite) TestPKCEFlow() {
	// Test code verifier generation
	verifier, err := GenerateCodeVerifier()
	s.NoError(err)
	s.Len(verifier, 128) // Should be 128 characters base64url
	
	// Test code challenge generation
	challenge := GenerateCodeChallenge(verifier)
	s.NotEmpty(challenge)
	s.NotEqual(verifier, challenge) // Should be SHA256 hashed
	
	// Verify PKCE parameters in auth URL
	authURL, err := s.client.GetAuthURLWithPKCE("google", "test-state", verifier)
	s.NoError(err)
	s.Contains(authURL, "code_challenge=")
	s.Contains(authURL, "code_challenge_method=S256")
}

// Test enhanced JWT with refresh tokens
func (s *OAuthSecurityTestSuite) TestJWTRefreshTokenFlow() {
	userInfo := &UserInfo{
		ID:       "123",
		Email:    "test@example.com",
		Provider: "google",
	}
	
	// Generate access and refresh tokens
	tokens, err := s.jwtManager.GenerateTokenPair(userInfo, "device-123", "127.0.0.1")
	s.NoError(err)
	s.NotEmpty(tokens.AccessToken)
	s.NotEmpty(tokens.RefreshToken)
	s.NotEqual(tokens.AccessToken, tokens.RefreshToken)
	
	// Verify access token has shorter expiry
	accessClaims, err := s.jwtManager.ValidateAccessToken(tokens.AccessToken)
	s.NoError(err)
	s.WithinDuration(time.Now().Add(15*time.Minute), accessClaims.ExpiresAt.Time, time.Minute)
	
	// Verify refresh token has longer expiry
	refreshClaims, err := s.jwtManager.ValidateRefreshToken(tokens.RefreshToken)
	s.NoError(err)
	s.WithinDuration(time.Now().Add(7*24*time.Hour), refreshClaims.ExpiresAt.Time, time.Hour)
	
	// Test token refresh
	newTokens, err := s.jwtManager.RefreshTokens(tokens.RefreshToken, "device-123", "127.0.0.1")
	s.NoError(err)
	s.NotEqual(tokens.AccessToken, newTokens.AccessToken)
}

// Test JWT fingerprinting for enhanced security
func (s *OAuthSecurityTestSuite) TestJWTFingerprinting() {
	userInfo := &UserInfo{
		ID:       "123",
		Email:    "test@example.com",
		Provider: "google",
	}
	
	deviceID := "device-123"
	ipAddress := "192.168.1.1"
	
	// Generate token with fingerprint
	tokens, err := s.jwtManager.GenerateTokenPair(userInfo, deviceID, ipAddress)
	s.NoError(err)
	
	// Validate token with correct fingerprint
	claims, err := s.jwtManager.ValidateWithFingerprint(tokens.AccessToken, deviceID, ipAddress)
	s.NoError(err)
	s.Equal(userInfo.ID, claims.UserID)
	
	// Validate token with wrong device ID should fail
	_, err = s.jwtManager.ValidateWithFingerprint(tokens.AccessToken, "wrong-device", ipAddress)
	s.Error(err)
	s.Contains(err.Error(), "fingerprint mismatch")
	
	// Validate token with wrong IP should fail
	_, err = s.jwtManager.ValidateWithFingerprint(tokens.AccessToken, deviceID, "10.0.0.1")
	s.Error(err)
	s.Contains(err.Error(), "fingerprint mismatch")
}

// Test session security with device tracking
func (s *OAuthSecurityTestSuite) TestSecureSessionManagement() {
	userInfo := &UserInfo{
		ID:       "123",
		Email:    "test@example.com",
		Provider: "google",
	}
	
	sessionManager := NewSessionManager(s.redisClient)
	
	// Create secure session
	session, err := sessionManager.CreateSession(s.ctx, userInfo, "device-123", "192.168.1.1", "Mozilla/5.0")
	s.NoError(err)
	s.NotEmpty(session.ID)
	s.Equal(userInfo.ID, session.UserID)
	s.Equal("device-123", session.DeviceID)
	
	// Test session retrieval
	retrieved, err := sessionManager.GetSession(s.ctx, session.ID)
	s.NoError(err)
	s.Equal(session.ID, retrieved.ID)
	
	// Test session timeout
	err = sessionManager.TouchSession(s.ctx, session.ID)
	s.NoError(err)
	
	// Test concurrent session limiting
	sessions, err := sessionManager.GetUserSessions(s.ctx, userInfo.ID)
	s.NoError(err)
	s.Len(sessions, 1)
	
	// Create multiple sessions and verify limit
	for i := 0; i < 5; i++ {
		_, err = sessionManager.CreateSession(s.ctx, userInfo, fmt.Sprintf("device-%d", i), "192.168.1.1", "Mozilla/5.0")
		s.NoError(err)
	}
	
	// Should enforce max concurrent sessions (e.g., 3)
	sessions, err = sessionManager.GetUserSessions(s.ctx, userInfo.ID)
	s.NoError(err)
	s.LessOrEqual(len(sessions), 3)
}

// Test rate limiting implementation
func (s *OAuthSecurityTestSuite) TestRateLimiting() {
	limiter := NewRateLimiter(s.redisClient, 10, time.Minute) // 10 requests per minute
	
	// Test within limit
	for i := 0; i < 10; i++ {
		allowed, err := limiter.Allow(s.ctx, "user-123", "login")
		s.NoError(err)
		s.True(allowed)
	}
	
	// Test exceeding limit
	allowed, err := limiter.Allow(s.ctx, "user-123", "login")
	s.NoError(err)
	s.False(allowed)
	
	// Test different key doesn't affect limit
	allowed, err = limiter.Allow(s.ctx, "user-456", "login")
	s.NoError(err)
	s.True(allowed)
}

// Test security headers middleware
func (s *OAuthSecurityTestSuite) TestSecurityHeaders() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	secureHandler := SecurityHeadersMiddleware(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	
	secureHandler.ServeHTTP(recorder, req)
	
	// Verify security headers
	s.Equal("max-age=31536000; includeSubDomains", recorder.Header().Get("Strict-Transport-Security"))
	s.Equal("nosniff", recorder.Header().Get("X-Content-Type-Options"))
	s.Equal("DENY", recorder.Header().Get("X-Frame-Options"))
	s.Equal("1; mode=block", recorder.Header().Get("X-XSS-Protection"))
	s.Contains(recorder.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

// Test CORS configuration
func (s *OAuthSecurityTestSuite) TestCORSConfiguration() {
	allowedOrigins := []string{"https://app.hexabase.com", "https://staging.hexabase.com"}
	corsHandler := ConfigureCORS(allowedOrigins)
	
	handler := corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	
	// Test allowed origin
	req := httptest.NewRequest("OPTIONS", "/api/auth", nil)
	req.Header.Set("Origin", "https://app.hexabase.com")
	recorder := httptest.NewRecorder()
	
	handler.ServeHTTP(recorder, req)
	
	s.Equal("https://app.hexabase.com", recorder.Header().Get("Access-Control-Allow-Origin"))
	s.Equal("true", recorder.Header().Get("Access-Control-Allow-Credentials"))
	s.Contains(recorder.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	
	// Test disallowed origin
	req2 := httptest.NewRequest("OPTIONS", "/api/auth", nil)
	req2.Header.Set("Origin", "https://evil.com")
	recorder2 := httptest.NewRecorder()
	
	handler.ServeHTTP(recorder2, req2)
	
	s.Empty(recorder2.Header().Get("Access-Control-Allow-Origin"))
}

// Test audit logging
func (s *OAuthSecurityTestSuite) TestAuditLogging() {
	logger := NewAuditLogger(s.redisClient)
	
	// Log authentication event
	event := AuditEvent{
		Type:      "auth.login",
		UserID:    "123",
		IP:        "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Provider:  "google",
		Success:   true,
		Timestamp: time.Now(),
	}
	
	err := logger.Log(s.ctx, event)
	s.NoError(err)
	
	// Retrieve audit logs
	logs, err := logger.GetUserLogs(s.ctx, "123", 10)
	s.NoError(err)
	s.Len(logs, 1)
	s.Equal("auth.login", logs[0].Type)
}

// Test token revocation
func (s *OAuthSecurityTestSuite) TestTokenRevocation() {
	userInfo := &UserInfo{
		ID:       "123",
		Email:    "test@example.com",
		Provider: "google",
	}
	
	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(userInfo, "device-123", "127.0.0.1")
	s.NoError(err)
	
	// Token should be valid initially
	_, err = s.jwtManager.ValidateAccessToken(tokens.AccessToken)
	s.NoError(err)
	
	// Revoke token
	err = s.jwtManager.RevokeToken(s.ctx, tokens.AccessToken)
	s.NoError(err)
	
	// Token should now be invalid
	_, err = s.jwtManager.ValidateAccessToken(tokens.AccessToken)
	s.Error(err)
	s.Contains(err.Error(), "token revoked")
}

// Test multi-provider OAuth handling
func (s *OAuthSecurityTestSuite) TestMultiProviderOAuth() {
	providers := []string{"google", "github", "gitlab"}
	
	for _, provider := range providers {
		// Test auth URL generation
		authURL, err := s.client.GetAuthURL(provider, "test-state")
		s.NoError(err)
		s.Contains(authURL, provider)
		
		// Test provider-specific validation
		err = s.client.ValidateProvider(provider)
		s.NoError(err)
	}
	
	// Test invalid provider
	_, err := s.client.GetAuthURL("invalid-provider", "test-state")
	s.Error(err)
}

// Test session hijacking prevention
func (s *OAuthSecurityTestSuite) TestSessionHijackingPrevention() {
	userInfo := &UserInfo{
		ID:       "123",
		Email:    "test@example.com",
		Provider: "google",
	}
	
	sessionManager := NewSessionManager(s.redisClient)
	
	// Create session from one IP
	session, err := sessionManager.CreateSession(s.ctx, userInfo, "device-123", "192.168.1.1", "Mozilla/5.0")
	s.NoError(err)
	
	// Try to use session from different IP
	err = sessionManager.ValidateSession(s.ctx, session.ID, "10.0.0.1", "Mozilla/5.0")
	s.Error(err)
	s.Contains(err.Error(), "IP mismatch")
	
	// Try to use session with different user agent
	err = sessionManager.ValidateSession(s.ctx, session.ID, "192.168.1.1", "Chrome/96.0")
	s.Error(err)
	s.Contains(err.Error(), "user agent mismatch")
}

func TestOAuthSecurityTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthSecurityTestSuite))
}

// GenerateCodeVerifier generates a random code verifier for PKCE
func GenerateCodeVerifier() (string, error) {
	b := make([]byte, 96) // Will encode to 128 chars in base64url
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}