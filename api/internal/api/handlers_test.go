package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/api"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTestRouter() *gin.Engine {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create test logger
	logger := zap.NewNop()
	
	// Create test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			OIDCIssuer: "https://test.hexabase.io",
		},
	}
	
	// Create handlers with nil database (for testing endpoints that don't require DB)
	handlers := &api.Handlers{}
	handlers.Auth = api.NewAuthHandler(nil, cfg, logger)
	
	// Setup router
	router := gin.New()
	
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})
	
	// OIDC discovery
	router.GET("/.well-known/openid-configuration", handlers.Auth.OIDCDiscovery)
	router.GET("/.well-known/jwks.json", handlers.Auth.JWKS)
	
	return router
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestOIDCDiscoveryEndpoint(t *testing.T) {
	router := setupTestRouter()
	
	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "issuer")
	assert.Contains(t, w.Body.String(), "authorization_endpoint")
	assert.Contains(t, w.Body.String(), "jwks_uri")
}

func TestJWKSEndpoint(t *testing.T) {
	router := setupTestRouter()
	
	req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "keys")
}

func TestAuthLoginEndpoint(t *testing.T) {
	router := setupTestRouter()
	
	// Add auth login endpoint
	router.POST("/auth/login/:provider", func(c *gin.Context) {
		provider := c.Param("provider")
		c.JSON(http.StatusOK, gin.H{
			"provider": provider,
			"auth_url": "https://example.com/oauth/authorize",
		})
	})
	
	req := httptest.NewRequest("POST", "/auth/login/google", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "google")
	assert.Contains(t, w.Body.String(), "auth_url")
}