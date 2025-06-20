package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/aiops/domain"
)

// GinHandler wraps the AIOps handler for Gin framework
type GinHandler struct {
	service domain.Service
	logger  *slog.Logger
}

// NewGinHandler creates a new AIOps Gin handler
func NewGinHandler(service domain.Service, logger *slog.Logger) *GinHandler {
	return &GinHandler{
		service: service,
		logger:  logger,
	}
}

// CreateChatSession handles POST /api/v1/aiops/sessions
func (h *GinHandler) CreateChatSession(c *gin.Context) {
	var req CreateChatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to decode request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.WorkspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}
	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	ctx := c.Request.Context()
	title := "New Chat Session" // Default title
	session, err := h.service.CreateChatSession(ctx, req.WorkspaceID, req.UserID, title, req.Model)
	if err != nil {
		h.logger.Error("Failed to create chat session", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat session"})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetChatSession handles GET /api/v1/aiops/sessions/:sessionId
func (h *GinHandler) GetChatSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	ctx := c.Request.Context()

	session, err := h.service.GetChatSession(ctx, sessionID)
	if err != nil {
		h.logger.Error("Failed to get chat session", slog.Any("error", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ListChatSessions handles GET /api/v1/aiops/sessions
func (h *GinHandler) ListChatSessions(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ctx := c.Request.Context()
	sessions, err := h.service.ListChatSessions(ctx, workspaceID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list chat sessions", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list chat sessions"})
		return
	}

	response := ListChatSessionsResponse{
		Sessions: sessions,
		Total:    len(sessions), // In real implementation, this would be a separate count query
	}

	c.JSON(http.StatusOK, response)
}

// DeleteChatSession handles DELETE /api/v1/aiops/sessions/:sessionId
func (h *GinHandler) DeleteChatSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	ctx := c.Request.Context()

	err := h.service.DeleteChatSession(ctx, sessionID)
	if err != nil {
		h.logger.Error("Failed to delete chat session", slog.Any("error", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat session not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Chat handles POST /api/v1/aiops/sessions/:sessionId/chat
func (h *GinHandler) Chat(c *gin.Context) {
	sessionID := c.Param("sessionId")
	ctx := c.Request.Context()

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to decode request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	// Create chat message
	message := domain.ChatMessage{
		Role:    "user",
		Content: req.Message,
	}

	// Check if client accepts streaming
	if c.GetHeader("Accept") == "text/event-stream" {
		h.streamChat(c, sessionID, message)
		return
	}

	// Regular non-streaming chat
	response, err := h.service.SendMessage(ctx, sessionID, message)
	if err != nil {
		h.logger.Error("Failed to process chat", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process chat"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// streamChat handles streaming chat responses
func (h *GinHandler) streamChat(c *gin.Context, sessionID string, message domain.ChatMessage) {
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx := c.Request.Context()
	// Start streaming
	stream, err := h.service.StreamMessage(ctx, sessionID, message)
	if err != nil {
		h.logger.Error("Failed to start chat stream", slog.Any("error", err))
		c.SSEvent("error", err.Error())
		c.Writer.Flush()
		return
	}

	// Stream responses
	for response := range stream {
		if response.Error != "" {
			c.SSEvent("error", response.Error)
		} else if response.Done {
			// Send final message with context
			c.SSEvent("done", gin.H{
				"done":    true,
				"context": response.Context,
			})
		} else {
			// Send message content
			c.SSEvent("message", response.Message.Content)
		}
		c.Writer.Flush()
	}
}

// GetAvailableModels handles GET /api/v1/aiops/models
func (h *GinHandler) GetAvailableModels(c *gin.Context) {
	ctx := c.Request.Context()
	models, err := h.service.ListAvailableModels(ctx)
	if err != nil {
		h.logger.Error("Failed to get available models", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get available models"})
		return
	}

	response := GetModelsResponse{
		Models: models,
	}

	c.JSON(http.StatusOK, response)
}

// GetTokenUsage handles GET /api/v1/aiops/usage
func (h *GinHandler) GetTokenUsage(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	// Parse time range from query params
	from := time.Now().AddDate(0, -1, 0) // Default to last month
	to := time.Now()
	
	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = parsed
		}
	}
	
	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = parsed
		}
	}

	ctx := c.Request.Context()
	usageReport, err := h.service.GetUsageStats(ctx, workspaceID, from, to)
	if err != nil {
		h.logger.Error("Failed to get usage stats", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage stats"})
		return
	}

	c.JSON(http.StatusOK, usageReport)
}