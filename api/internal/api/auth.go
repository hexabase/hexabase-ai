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
	
	// Parse PKCE parameters if provided
	var loginReq struct {
		CodeChallenge       string `json:"code_challenge"`
		CodeChallengeMethod string `json:"code_challenge_method"`
	}
	_ = c.ShouldBindJSON(&loginReq)
	
	// Get client IP for security tracking
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	
	authURL, state, err := h.authService.GetAuthURLWithPKCE(provider, loginReq.CodeChallenge, loginReq.CodeChallengeMethod)
	if err != nil {
		h.logger.Error("Failed to generate auth URL", 
			zap.Error(err),
			zap.String("provider", provider))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid provider or configuration error",
		})
		return
	}
	
	// Store state with additional security context
	if err := h.authService.StoreAuthState(state, clientIP, userAgent); err != nil {
		h.logger.Error("Failed to store auth state", zap.Error(err))
	}
	
	h.logger.Info("Login request", 
		zap.String("provider", provider),
		zap.String("state", state),
		zap.String("client_ip", clientIP))
	
	c.JSON(http.StatusOK, gin.H{
		"provider": provider,
		"auth_url": authURL,
		"state":    state,
	})
}

// CallbackProvider handles OAuth callback from external provider
func (h *AuthHandler) CallbackProvider(c *gin.Context) {
	provider := c.Param("provider")
	
	// Handle both query params (for redirect) and JSON body (for PKCE)
	code := c.Query("code")
	state := c.Query("state")
	
	// If not in query, check JSON body for PKCE flow
	if code == "" {
		var callbackReq struct {
			Code         string `json:"code"`
			State        string `json:"state"`
			CodeVerifier string `json:"code_verifier"`
		}
		if err := c.ShouldBindJSON(&callbackReq); err == nil {
			code = callbackReq.Code
			state = callbackReq.State
			
			// Verify PKCE if provided
			if callbackReq.CodeVerifier != "" {
				if err := h.authService.VerifyPKCE(state, callbackReq.CodeVerifier); err != nil {
					h.logger.Error("PKCE verification failed", zap.Error(err))
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid PKCE verification",
					})
					return
				}
			}
		}
	}
	
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing authorization code",
		})
		return
	}
	
	// Get client context for security
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	
	// Verify state and get auth context
	if err := h.authService.VerifyAuthState(state, clientIP); err != nil {
		h.logger.Error("State verification failed", 
			zap.Error(err),
			zap.String("state", state),
			zap.String("client_ip", clientIP))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid or expired state",
		})
		return
	}
	
	user, tokenPair, err := h.authService.HandleCallbackWithTokenPair(c.Request.Context(), provider, code, state, clientIP, userAgent)
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
		zap.String("email", user.Email),
		zap.String("client_ip", clientIP))
	
	// Log security event
	h.authService.LogSecurityEvent(user.ID, "login", "OAuth login successful", clientIP, "info")
	
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"token_type":    tokenPair.TokenType,
		"expires_in":    tokenPair.ExpiresIn,
		"user": gin.H{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"provider":     provider,
		},
	})
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	
	// Get client context
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	
	tokenPair, err := h.authService.RefreshTokenPair(req.RefreshToken, clientIP, userAgent)
	if err != nil {
		h.logger.Error("Token refresh failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"token_type":    tokenPair.TokenType,
		"expires_in":    tokenPair.ExpiresIn,
	})
}

// Logout invalidates user session
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)
	
	// Get user from context
	userID, _ := c.Get("user_id")
	
	// Revoke refresh token if provided
	if req.RefreshToken != "" {
		if err := h.authService.RevokeRefreshToken(req.RefreshToken); err != nil {
			h.logger.Error("Failed to revoke refresh token", zap.Error(err))
		}
	}
	
	// Log security event
	clientIP := c.ClientIP()
	if userIDStr, ok := userID.(string); ok {
		h.authService.LogSecurityEvent(userIDStr, "logout", "User logged out", clientIP, "info")
	}
	
	h.logger.Info("User logout", zap.Any("user_id", userID))
	
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
	provider, _ := c.Get("provider")
	
	// TODO: Fetch user's organizations from database
	
	h.logger.Info("Get current user", zap.Any("user_id", userID))
	
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       userID,
			"email":    email,
			"name":     name,
			"provider": provider,
		},
		"organizations": []gin.H{
			// TODO: Populate from database
		},
	})
}

// GetSessions returns active sessions for the current user
func (h *AuthHandler) GetSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)
	
	sessions, err := h.authService.GetUserSessions(userIDStr)
	if err != nil {
		h.logger.Error("Failed to get sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve sessions",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
	})
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)
	
	if err := h.authService.RevokeSession(userIDStr, sessionID); err != nil {
		h.logger.Error("Failed to revoke session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revoke session",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Session revoked successfully",
	})
}

// RevokeAllSessions revokes all sessions except current
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)
	
	// Get current session ID from token
	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	
	if err := h.authService.RevokeAllSessionsExcept(userIDStr, token); err != nil {
		h.logger.Error("Failed to revoke sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revoke sessions",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "All other sessions revoked successfully",
	})
}

// GetSecurityLogs returns security logs for the current user
func (h *AuthHandler) GetSecurityLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)
	
	logs, err := h.authService.GetSecurityLogs(userIDStr, 50) // Last 50 events
	if err != nil {
		h.logger.Error("Failed to get security logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve security logs",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
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