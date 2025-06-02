package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/api"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthTestRouter(t *testing.T) (*gin.Engine, *api.AuthHandler) {
	// Setup test database (using SQLite for simplicity)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			OIDCIssuer:    "https://api.hexabase.test",
			JWTExpiration: 3600,
			ExternalProviders: map[string]config.OAuthProvider{
				"google": {
					ClientID:     "test-google-client",
					ClientSecret: "test-google-secret",
					RedirectURL:  "https://api.hexabase.test/auth/callback/google",
					Scopes:       []string{"openid", "email", "profile"},
				},
				"github": {
					ClientID:     "test-github-client",
					ClientSecret: "test-github-secret",
					RedirectURL:  "https://api.hexabase.test/auth/callback/github",
					Scopes:       []string{"user:email", "read:user"},
				},
			},
		},
	}

	logger := zap.NewNop()

	// Create handler
	authHandler := api.NewAuthHandler(db, cfg, logger)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	authGroup := router.Group("/auth")
	{
		authGroup.GET("/login/:provider", authHandler.LoginProvider)
		authGroup.GET("/callback/:provider", authHandler.CallbackProvider)
		authGroup.POST("/logout", authHandler.AuthMiddleware(), authHandler.Logout)
		authGroup.GET("/me", authHandler.AuthMiddleware(), authHandler.GetCurrentUser)
		authGroup.GET("/.well-known/openid-configuration", authHandler.OIDCDiscovery)
		authGroup.GET("/.well-known/jwks.json", authHandler.JWKS)
	}

	return router, authHandler
}

func TestAuthHandler_LoginProvider(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	// Test Google login
	req := httptest.NewRequest("GET", "/auth/login/google", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "google", response["provider"])
	assert.Contains(t, response, "auth_url")
	assert.Contains(t, response, "state")
	assert.Contains(t, response["auth_url"], "accounts.google.com")

	// Test GitHub login
	req = httptest.NewRequest("GET", "/auth/login/github", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "github", response["provider"])
	assert.Contains(t, response["auth_url"], "github.com")

	// Test invalid provider
	req = httptest.NewRequest("GET", "/auth/login/invalid", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_CallbackProvider_MissingCode(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	// Test callback without code
	req := httptest.NewRequest("GET", "/auth/callback/google?state=test-state", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Missing authorization code")
}

func TestAuthHandler_Logout(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	// Test logout without auth (should fail)
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetCurrentUser_Unauthorized(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	// Test without authorization header
	req := httptest.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "missing authorization header")
}

func TestAuthHandler_AuthMiddleware_InvalidToken(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	// Test with invalid authorization header format
	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid authorization header format")

	// Test with invalid token
	req = httptest.NewRequest("GET", "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid or expired token")
}

func TestAuthHandler_OIDCDiscovery(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	req := httptest.NewRequest("GET", "/auth/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var discovery map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &discovery)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.hexabase.test", discovery["issuer"])
	assert.Contains(t, discovery, "authorization_endpoint")
	assert.Contains(t, discovery, "token_endpoint")
	assert.Contains(t, discovery, "userinfo_endpoint")
	assert.Contains(t, discovery, "jwks_uri")
	assert.Contains(t, discovery, "response_types_supported")
	assert.Contains(t, discovery, "subject_types_supported")
	assert.Contains(t, discovery, "id_token_signing_alg_values_supported")
	assert.Contains(t, discovery, "scopes_supported")
	assert.Contains(t, discovery, "claims_supported")
}

func TestAuthHandler_JWKS(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	req := httptest.NewRequest("GET", "/auth/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var jwks map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &jwks)
	assert.NoError(t, err)
	assert.Contains(t, jwks, "keys")
	
	// Verify we actually have keys in the response
	keys, ok := jwks["keys"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, keys, 1)
	
	// Verify the key structure
	key, ok := keys[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "RSA", key["kty"])
	assert.Equal(t, "sig", key["use"])
	assert.Equal(t, "RS256", key["alg"])
	assert.NotEmpty(t, key["kid"])
	assert.NotEmpty(t, key["n"])
	assert.NotEmpty(t, key["e"])
}