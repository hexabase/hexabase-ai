package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db          *gorm.DB
	config      *config.Config
	logger      *zap.Logger
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *AuthHandler {
	authService, err := service.NewAuthService(db, cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize auth service", zap.Error(err))
	}

	return &AuthHandler{
		db:          db,
		config:      cfg,
		logger:      logger,
		authService: authService,
	}
}

// LoginProvider initiates OAuth flow with external provider
func (h *AuthHandler) LoginProvider(c *gin.Context) {
	provider := c.Param("provider")
	
	authURL, state, err := h.authService.GetAuthURL(provider)
	if err != nil {
		h.logger.Error("Failed to generate auth URL", 
			zap.Error(err),
			zap.String("provider", provider))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid provider or configuration error",
		})
		return
	}
	
	h.logger.Info("Login request", 
		zap.String("provider", provider),
		zap.String("state", state))
	
	c.JSON(http.StatusOK, gin.H{
		"provider": provider,
		"auth_url": authURL,
		"state":    state,
	})
}

// CallbackProvider handles OAuth callback from external provider
func (h *AuthHandler) CallbackProvider(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing authorization code",
		})
		return
	}
	
	user, token, err := h.authService.HandleCallback(c.Request.Context(), provider, code, state)
	if err != nil {
		h.logger.Error("OAuth callback failed", 
			zap.Error(err),
			zap.String("provider", provider))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication failed",
		})
		return
	}
	
	h.logger.Info("OAuth callback successful", 
		zap.String("provider", provider),
		zap.String("user_id", user.ID),
		zap.String("email", user.Email))
	
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
		},
	})
}

// Logout invalidates user session
func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: Implement session invalidation
	
	h.logger.Info("User logout")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// GetCurrentUser returns current user information
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not found in context",
		})
		return
	}
	
	email, _ := c.Get("user_email")
	name, _ := c.Get("user_name")
	
	// TODO: Fetch user's organizations from database
	
	h.logger.Info("Get current user", zap.Any("user_id", userID))
	
	c.JSON(http.StatusOK, gin.H{
		"id":    userID,
		"email": email,
		"name":  name,
		"organizations": []gin.H{
			// TODO: Populate from database
		},
	})
}

// AuthMiddleware validates JWT tokens and sets user context
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}
		
		// Extract Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}
		
		token := strings.TrimPrefix(authHeader, "Bearer ")
		
		// Validate JWT token
		claims, err := h.authService.ValidateToken(token)
		if err != nil {
			h.logger.Debug("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}
		
		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("org_ids", claims.OrgIDs)
		
		c.Next()
	}
}

// OIDCDiscovery returns OIDC discovery document
func (h *AuthHandler) OIDCDiscovery(c *gin.Context) {
	issuer := h.config.Auth.OIDCIssuer
	
	discovery := gin.H{
		"issuer":                 issuer,
		"authorization_endpoint": issuer + "/auth/authorize",
		"token_endpoint":         issuer + "/auth/token",
		"userinfo_endpoint":      issuer + "/auth/userinfo",
		"jwks_uri":              issuer + "/.well-known/jwks.json",
		"response_types_supported": []string{"code", "id_token", "token id_token"},
		"subject_types_supported":  []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported": []string{"openid", "profile", "email", "groups"},
		"claims_supported": []string{"sub", "iss", "aud", "exp", "iat", "auth_time", "groups"},
	}
	
	c.JSON(http.StatusOK, discovery)
}

// JWKS returns JSON Web Key Set for token verification
func (h *AuthHandler) JWKS(c *gin.Context) {
	jwksJSON, err := h.authService.GetJWKS()
	if err != nil {
		h.logger.Error("Failed to get JWKS", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve public keys",
		})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", jwksJSON)
}