package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/auth"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	service auth.Service
	logger  *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(service auth.Service, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

// Login initiates OAuth flow with external provider
func (h *AuthHandler) Login(c *gin.Context) {
	provider := c.Param("provider")

	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use default if not provided
		req.Provider = provider
	}

	authURL, state, err := h.service.GetAuthURL(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to generate auth URL", 
			zap.Error(err),
			zap.String("provider", provider))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider or configuration error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": provider,
		"auth_url": authURL,
		"state":    state,
	})
}

// Callback handles OAuth callback from external provider
func (h *AuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")

	// Handle both query params (for redirect) and JSON body (for PKCE)
	code := c.Query("code")
	state := c.Query("state")
	var codeVerifier string

	// If not in query, check JSON body for PKCE flow
	if code == "" {
		var req auth.CallbackRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			code = req.Code
			state = req.State
			codeVerifier = req.CodeVerifier
		}
	}

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code or state"})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	req := &auth.CallbackRequest{
		Code:         code,
		State:        state,
		CodeVerifier: codeVerifier,
	}

	authResp, err := h.service.HandleCallback(c.Request.Context(), req, clientIP, userAgent)
	if err != nil {
		h.logger.Error("OAuth callback failed", 
			zap.Error(err),
			zap.String("provider", provider))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	h.logger.Info("OAuth login successful", 
		zap.String("provider", provider),
		zap.String("user_id", authResp.User.ID),
		zap.String("email", authResp.User.Email))

	c.JSON(http.StatusOK, gin.H{
		"access_token":  authResp.AccessToken,
		"refresh_token": authResp.RefreshToken,
		"token_type":    authResp.TokenType,
		"expires_in":    authResp.ExpiresIn,
		"user": gin.H{
			"id":           authResp.User.ID,
			"email":        authResp.User.Email,
			"display_name": authResp.User.DisplayName,
			"provider":     authResp.User.Provider,
		},
	})
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req auth.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	tokenPair, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken, clientIP, userAgent)
	if err != nil {
		h.logger.Error("token refresh failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
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

	userID := c.GetString("user_id")

	// Revoke refresh token if provided
	if req.RefreshToken != "" {
		if err := h.service.RevokeRefreshToken(c.Request.Context(), req.RefreshToken); err != nil {
			h.logger.Error("failed to revoke refresh token", zap.Error(err))
		}
	}

	// Log security event
	event := &auth.SecurityEvent{
		UserID:      userID,
		EventType:   "logout",
		Description: "User logged out",
		IPAddress:   c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
		Level:       "info",
	}

	if err := h.service.LogSecurityEvent(c.Request.Context(), event); err != nil {
		h.logger.Error("failed to log security event", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}

// GetCurrentUser returns current user information
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	user, err := h.service.GetCurrentUser(c.Request.Context(), token)
	if err != nil {
		h.logger.Error("failed to get current user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user information"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"provider":     user.Provider,
			"created_at":   user.CreatedAt,
			"last_login":   user.LastLoginAt,
		},
	})
}

// GetSessions returns active sessions for the current user
func (h *AuthHandler) GetSessions(c *gin.Context) {
	userID := c.GetString("user_id")

	sessions, err := h.service.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	userID := c.GetString("user_id")

	if err := h.service.RevokeSession(c.Request.Context(), userID, sessionID); err != nil {
		h.logger.Error("failed to revoke session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session revoked successfully"})
}

// RevokeAllSessions revokes all sessions except current
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.GetString("session_id") // Set by middleware

	if err := h.service.RevokeAllSessions(c.Request.Context(), userID, sessionID); err != nil {
		h.logger.Error("failed to revoke sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all other sessions revoked successfully"})
}

// GetSecurityLogs returns security logs for the current user
func (h *AuthHandler) GetSecurityLogs(c *gin.Context) {
	userID := c.GetString("user_id")

	filter := auth.SecurityLogFilter{
		UserID: userID,
		Limit:  50, // Last 50 events
	}

	logs, err := h.service.GetSecurityLogs(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to get security logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve security logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": len(logs),
	})
}

// OIDCDiscovery returns OIDC discovery document
func (h *AuthHandler) OIDCDiscovery(c *gin.Context) {
	config, err := h.service.GetOIDCConfiguration(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get OIDC configuration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// JWKS returns JSON Web Key Set for token verification
func (h *AuthHandler) JWKS(c *gin.Context) {
	jwks, err := h.service.GetJWKS(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get JWKS", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve public keys"})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", jwks)
}

// AuthMiddleware validates JWT tokens and sets user context
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Extract Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate JWT token
		claims, err := h.service.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			h.logger.Debug("token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.Subject)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("provider", claims.Provider)
		c.Set("org_ids", claims.OrgIDs)

		c.Next()
	}
}