package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/auth"
)

// AIOpsProxyHandler forwards requests to the Python AIOps service.
type AIOpsProxyHandler struct {
	authSvc    auth.Service
	logger     *slog.Logger
	aiopsServiceURL *url.URL
}

// NewAIOpsProxyHandler creates a new AIOps proxy handler.
func NewAIOpsProxyHandler(authSvc auth.Service, logger *slog.Logger, aiopsServiceURL string) (*AIOpsProxyHandler, error) {
	parsedURL, err := url.Parse(aiopsServiceURL)
	if err != nil {
		return nil, err
	}
	return &AIOpsProxyHandler{
		authSvc:    authSvc,
		logger:     logger,
		aiopsServiceURL: parsedURL,
	}, nil
}

// ChatProxy handles proxying chat requests to the AIOps service.
func (h *AIOpsProxyHandler) ChatProxy(c *gin.Context) {
	// 1. Get user details from the authenticated context.
	userID := c.GetString("user_id")
	orgIDsInterface, _ := c.Get("org_ids")
	orgIDs, _ := orgIDsInterface.([]string)
	
	// For now, we'll assume the active workspace is passed as a header or query param.
	// In a real app, this might come from the user's session.
	activeWorkspaceID := c.Query("workspace_id")
	if activeWorkspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id query parameter is required"})
		return
	}

	// 2. Generate the internal token.
	internalToken, err := h.authSvc.GenerateInternalAIOpsToken(c.Request.Context(), userID, orgIDs, activeWorkspaceID)
	if err != nil {
		h.logger.Error("failed to generate internal AIOps token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process request"})
		return
	}

	// 3. Set up the reverse proxy.
	proxy := httputil.NewSingleHostReverseProxy(h.aiopsServiceURL)
	
	proxy.Director = func(req *http.Request) {
		// Rewrite the request to target the AIOps service.
		req.Host = h.aiopsServiceURL.Host
		req.URL.Scheme = h.aiopsServiceURL.Scheme
		req.URL.Host = h.aiopsServiceURL.Host
		req.URL.Path = "/v1/chat" // Assuming the python service has a /v1/chat endpoint

		// Set the internal token.
		req.Header.Set("Authorization", "Bearer "+internalToken)
		// Preserve the original body
		req.Body = c.Request.Body
	}

	proxy.ServeHTTP(c.Writer, c.Request)
} 