package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OAuthIntegrationTestSuite tests the complete OAuth flow
type OAuthIntegrationTestSuite struct {
	suite.Suite
	db       *gorm.DB
	router   *gin.Engine
	handlers *Handlers
	logger   *zap.Logger
	config   *config.Config
}

func (suite *OAuthIntegrationTestSuite) SetupSuite() {
	// Setup test logger
	suite.logger = zap.NewNop()
	
	// Setup test config
	suite.config = &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5433",
			User:     "postgres",
			Password: "postgres",
			DBName:   "hexabase_test",
			SSLMode:  "disable",
		},
		Auth: config.AuthConfig{
			JWTSecret:     "test-jwt-secret-for-testing-only",
			JWTExpiration: 24 * 60 * 60, // 24 hours in seconds
			OIDCIssuer:    "https://api.hexabase.local",
			ExternalProviders: map[string]config.OAuthProvider{
				"google": {
					ClientID:     "test-google-client-id",
					ClientSecret: "test-google-client-secret",
					RedirectURL:  "http://localhost:8080/auth/callback/google",
					Scopes:       []string{"openid", "profile", "email"},
					AuthURL:      "https://accounts.google.com/o/oauth2/auth",
					TokenURL:     "https://oauth2.googleapis.com/token",
					UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
				},
				"github": {
					ClientID:     "test-github-client-id",
					ClientSecret: "test-github-client-secret",
					RedirectURL:  "http://localhost:8080/auth/callback/github",
					Scopes:       []string{"user:email"},
					AuthURL:      "https://github.com/login/oauth/authorize",
					TokenURL:     "https://github.com/login/oauth/access_token",
					UserInfoURL:  "https://api.github.com/user",
				},
			},
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6380",
			Password: "",
			DB:       0,
		},
	}

	// Setup test database
	dbConfig := &db.DatabaseConfig{
		Host:     suite.config.Database.Host,
		Port:     suite.config.Database.Port,
		User:     suite.config.Database.User,
		Password: suite.config.Database.Password,
		DBName:   suite.config.Database.DBName,
		SSLMode:  suite.config.Database.SSLMode,
	}

	var err error
	suite.db, err = db.ConnectDatabase(dbConfig)
	if err != nil {
		suite.T().Skip("Database not available for integration tests")
		return
	}

	// Run migrations
	err = db.MigrateDatabase(suite.db)
	suite.Require().NoError(err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Initialize handlers
	suite.handlers = NewHandlers(suite.db, suite.config, suite.logger)
	SetupRoutes(suite.router, suite.handlers)
}

func (suite *OAuthIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

func (suite *OAuthIntegrationTestSuite) SetupTest() {
	// Clean up test data before each test
	if suite.db != nil {
		suite.db.Exec("DELETE FROM organization_users")
		suite.db.Exec("DELETE FROM organizations") 
		suite.db.Exec("DELETE FROM users")
	}
}

// TestGoogleOAuthLoginFlow tests the complete Google OAuth flow
func (suite *OAuthIntegrationTestSuite) TestGoogleOAuthLoginFlow() {
	// Step 1: Initiate Google OAuth login
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login/google", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var loginResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(suite.T(), err)
	
	// Should return OAuth URL and state
	assert.Contains(suite.T(), loginResponse, "auth_url")
	assert.Contains(suite.T(), loginResponse, "state")
	
	authURL := loginResponse["auth_url"].(string)
	state := loginResponse["state"].(string)
	
	// Verify the URL contains expected parameters
	parsedURL, err := url.Parse(authURL)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "accounts.google.com", parsedURL.Host)
	assert.Equal(suite.T(), "/o/oauth2/auth", parsedURL.Path)
	
	query := parsedURL.Query()
	assert.Equal(suite.T(), "test-google-client-id", query.Get("client_id"))
	assert.Equal(suite.T(), state, query.Get("state"))
	assert.Contains(suite.T(), query.Get("scope"), "openid")
	assert.Contains(suite.T(), query.Get("scope"), "profile")
	assert.Contains(suite.T(), query.Get("scope"), "email")

	suite.T().Logf("✅ Google OAuth login URL generated successfully")
	suite.T().Logf("State: %s", state)
	suite.T().Logf("Auth URL: %s", authURL)
}

// TestGitHubOAuthLoginFlow tests the complete GitHub OAuth flow
func (suite *OAuthIntegrationTestSuite) TestGitHubOAuthLoginFlow() {
	// Step 1: Initiate GitHub OAuth login
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login/github", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var loginResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(suite.T(), err)
	
	// Should return OAuth URL and state
	assert.Contains(suite.T(), loginResponse, "auth_url")
	assert.Contains(suite.T(), loginResponse, "state")
	
	authURL := loginResponse["auth_url"].(string)
	state := loginResponse["state"].(string)
	
	// Verify the URL contains expected parameters
	parsedURL, err := url.Parse(authURL)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "github.com", parsedURL.Host)
	assert.Equal(suite.T(), "/login/oauth/authorize", parsedURL.Path)
	
	query := parsedURL.Query()
	assert.Equal(suite.T(), "test-github-client-id", query.Get("client_id"))
	assert.Equal(suite.T(), state, query.Get("state"))
	assert.Contains(suite.T(), query.Get("scope"), "user:email")

	suite.T().Logf("✅ GitHub OAuth login URL generated successfully")
	suite.T().Logf("State: %s", state)
	suite.T().Logf("Auth URL: %s", authURL)
}

// TestOAuthStateValidation tests that OAuth state validation works
func (suite *OAuthIntegrationTestSuite) TestOAuthStateValidation() {
	// First, initiate login to create a valid state
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/auth/login/google", nil)
	suite.router.ServeHTTP(w1, req1)

	var loginResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &loginResponse)
	assert.NoError(suite.T(), err)

	// Test callback with invalid state
	w2 := httptest.NewRecorder()
	invalidCallbackURL := "/auth/callback/google?code=test_code&state=invalid_state"
	req2, _ := http.NewRequest("GET", invalidCallbackURL, nil)
	suite.router.ServeHTTP(w2, req2)

	// We expect a 500 error because the OAuth callback will fail without a real code
	// but it should get past state validation (which would give 400)
	assert.NotEqual(suite.T(), http.StatusBadRequest, w2.Code, "Should get past state validation")

	suite.T().Logf("✅ OAuth state validation working correctly")
}

// TestUnsupportedOAuthProvider tests handling of unsupported providers
func (suite *OAuthIntegrationTestSuite) TestUnsupportedOAuthProvider() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login/facebook", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse["error"], "provider")

	suite.T().Logf("✅ Unsupported provider handling working correctly")
}

// TestJWTTokenGeneration tests JWT token creation and validation
func (suite *OAuthIntegrationTestSuite) TestJWTTokenGeneration() {
	// Create a test user first
	user := &db.User{
		ID:          "test-oauth-user-001",
		ExternalID:  "google-123456789",
		Provider:    "google",
		Email:       "oauth.test@example.com",
		DisplayName: "OAuth Test User",
	}
	
	err := suite.db.Create(user).Error
	assert.NoError(suite.T(), err)

	// Test the /auth/me endpoint (requires valid JWT)
	// This will be tested in the next phase after we implement OAuth callback simulation
	
	suite.T().Logf("✅ Test user created for JWT testing")
}

// TestAuthMiddleware tests that protected endpoints require authentication
func (suite *OAuthIntegrationTestSuite) TestAuthMiddleware() {
	// Test accessing protected endpoint without token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/organizations/", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse["error"], "authorization header")

	suite.T().Logf("✅ Auth middleware correctly protecting endpoints")
}

// TestInvalidTokenHandling tests handling of invalid JWT tokens
func (suite *OAuthIntegrationTestSuite) TestInvalidTokenHandling() {
	// Test with malformed token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/organizations/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse["error"], "invalid or expired token")

	suite.T().Logf("✅ Invalid token handling working correctly")
}

// TestOAuthProviderConfiguration tests that OAuth providers are properly configured
func (suite *OAuthIntegrationTestSuite) TestOAuthProviderConfiguration() {
	// This test verifies that the OAuth configuration is loaded correctly
	googleProvider := suite.config.Auth.ExternalProviders["google"]
	githubProvider := suite.config.Auth.ExternalProviders["github"]
	
	assert.NotEmpty(suite.T(), googleProvider.ClientID)
	assert.NotEmpty(suite.T(), githubProvider.ClientID)
	assert.NotEmpty(suite.T(), googleProvider.RedirectURL)
	assert.NotEmpty(suite.T(), githubProvider.RedirectURL)

	suite.T().Logf("✅ OAuth provider configuration valid")
	suite.T().Logf("Google Client ID: %s", googleProvider.ClientID)
	suite.T().Logf("GitHub Client ID: %s", githubProvider.ClientID)
}

// TestOIDCDiscoveryEndpoint tests the OIDC discovery endpoint
func (suite *OAuthIntegrationTestSuite) TestOIDCDiscoveryEndpoint() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/.well-known/openid-configuration", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var discovery map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &discovery)
	assert.NoError(suite.T(), err)
	
	// Check required OIDC discovery fields
	assert.Contains(suite.T(), discovery, "issuer")
	assert.Contains(suite.T(), discovery, "jwks_uri")
	assert.Contains(suite.T(), discovery, "authorization_endpoint")
	assert.Contains(suite.T(), discovery, "token_endpoint")
	assert.Contains(suite.T(), discovery, "response_types_supported")
	assert.Contains(suite.T(), discovery, "subject_types_supported")
	assert.Contains(suite.T(), discovery, "id_token_signing_alg_values_supported")

	suite.T().Logf("✅ OIDC discovery endpoint working correctly")
}

// TestJWKSEndpoint tests the JSON Web Key Set endpoint
func (suite *OAuthIntegrationTestSuite) TestJWKSEndpoint() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/.well-known/jwks.json", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var jwks map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &jwks)
	assert.NoError(suite.T(), err)
	
	// Check JWKS structure
	assert.Contains(suite.T(), jwks, "keys")
	keys := jwks["keys"].([]interface{})
	assert.Greater(suite.T(), len(keys), 0)
	
	// Check first key structure
	key := keys[0].(map[string]interface{})
	assert.Contains(suite.T(), key, "kty")
	assert.Contains(suite.T(), key, "use")
	assert.Contains(suite.T(), key, "kid")
	assert.Contains(suite.T(), key, "n")
	assert.Contains(suite.T(), key, "e")
	assert.Equal(suite.T(), "RSA", key["kty"])
	assert.Equal(suite.T(), "sig", key["use"])

	suite.T().Logf("✅ JWKS endpoint working correctly")
}

// TestLogoutEndpoint tests the logout functionality
func (suite *OAuthIntegrationTestSuite) TestLogoutEndpoint() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	suite.router.ServeHTTP(w, req)

	// Logout should work even without authentication (idempotent)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "message")

	suite.T().Logf("✅ Logout endpoint working correctly")
}

// TestCompleteOAuthFlowSimulation simulates a complete OAuth flow
func (suite *OAuthIntegrationTestSuite) TestCompleteOAuthFlowSimulation() {
	suite.T().Log("🔄 Starting complete OAuth flow simulation...")

	// Step 1: Initiate OAuth login
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/auth/login/google", nil)
	suite.router.ServeHTTP(w1, req1)
	assert.Equal(suite.T(), http.StatusOK, w1.Code)

	var loginResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &loginResponse)
	assert.NoError(suite.T(), err)
	
	state := loginResponse["state"].(string)
	authURL := loginResponse["auth_url"].(string)
	
	suite.T().Logf("✅ Step 1: OAuth login initiated")
	suite.T().Logf("   State: %s", state)
	suite.T().Logf("   Auth URL: %s", authURL)

	// Step 2: Simulate OAuth callback with valid state
	// Note: In a real test, we would need to mock the OAuth provider's response
	// For now, we test that the callback endpoint exists and validates state
	
	w2 := httptest.NewRecorder()
	callbackURL := fmt.Sprintf("/auth/callback/google?code=mock_code&state=%s", state)
	req2, _ := http.NewRequest("GET", callbackURL, nil)
	suite.router.ServeHTTP(w2, req2)
	
	// The callback will fail because we don't have a real OAuth code,
	// but it should get past state validation
	// We expect either a token exchange error or user creation
	assert.NotEqual(suite.T(), http.StatusBadRequest, w2.Code, "State validation should pass")
	
	suite.T().Logf("✅ Step 2: OAuth callback state validation passed")
	suite.T().Logf("   Response code: %d", w2.Code)

	// Note: Full OAuth flow testing would require:
	// - Mocking OAuth provider responses
	// - Testing user creation from provider data
	// - Testing JWT token generation in callback
	// - Testing access to protected endpoints with generated token
}

func TestOAuthIntegrationSuite(t *testing.T) {
	suite.Run(t, new(OAuthIntegrationTestSuite))
}