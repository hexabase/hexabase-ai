package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"log/slog"

	"github.com/gorilla/mux"
	"github.com/hexabase/hexabase-ai/api/internal/aiops/domain"
)

// AIOpsService defines the interface for AIOps operations (preserved from original)
type AIOpsService interface {
	CreateChatSession(workspaceID, userID, model string) (*domain.ChatSession, error)
	GetChatSession(sessionID string) (*domain.ChatSession, error)
	ListChatSessions(workspaceID string, limit, offset int) ([]*domain.ChatSession, error)
	DeleteChatSession(sessionID string) error
	Chat(sessionID string, message string, context []int) (*domain.ChatResponse, error)
	StreamChat(sessionID string, message string, context []int) (<-chan *domain.ChatStreamResponse, error)
	GetAvailableModels() ([]*domain.ModelInfo, error)
	GetTokenUsage(workspaceID, model string, limit, offset int) ([]*domain.ModelUsage, error)
}

// Handler handles AIOps-related requests
type Handler struct {
	service AIOpsService
	logger  *slog.Logger
}

// NewHandler creates a new AIOps handler
func NewHandler(service AIOpsService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateChatSessionRequest represents the request to create a chat session
type CreateChatSessionRequest struct {
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
	Model       string `json:"model"`
}

// CreateChatSession handles POST /api/v1/aiops/sessions
func (h *Handler) CreateChatSession(w http.ResponseWriter, r *http.Request) {
	var req CreateChatSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.WorkspaceID == "" {
		h.respondWithError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}
	if req.UserID == "" {
		h.respondWithError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if req.Model == "" {
		h.respondWithError(w, http.StatusBadRequest, "model is required")
		return
	}

	session, err := h.service.CreateChatSession(req.WorkspaceID, req.UserID, req.Model)
	if err != nil {
		h.logger.Error("Failed to create chat session", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create chat session")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, session)
}

// GetChatSession handles GET /api/v1/aiops/sessions/{sessionId}
func (h *Handler) GetChatSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	session, err := h.service.GetChatSession(sessionID)
	if err != nil {
		h.logger.Error("Failed to get chat session", "error", err)
		h.respondWithError(w, http.StatusNotFound, "Chat session not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, session)
}

// ListChatSessionsResponse represents the response for listing chat sessions
type ListChatSessionsResponse struct {
	Sessions []*domain.ChatSession `json:"sessions"`
	Total    int                  `json:"total"`
}

// ListChatSessions handles GET /api/v1/aiops/sessions
func (h *Handler) ListChatSessions(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		h.respondWithError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}

	limit := 10
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	sessions, err := h.service.ListChatSessions(workspaceID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list chat sessions", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list chat sessions")
		return
	}

	response := ListChatSessionsResponse{
		Sessions: sessions,
		Total:    len(sessions), // In real implementation, this would be a separate count query
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// DeleteChatSession handles DELETE /api/v1/aiops/sessions/{sessionId}
func (h *Handler) DeleteChatSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	err := h.service.DeleteChatSession(sessionID)
	if err != nil {
		h.logger.Error("Failed to delete chat session", "error", err)
		h.respondWithError(w, http.StatusNotFound, "Chat session not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChatRequest represents a chat message request
type ChatRequest struct {
	Message string `json:"message"`
	Context []int  `json:"context,omitempty"`
}

// Chat handles POST /api/v1/aiops/sessions/{sessionId}/chat
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Message == "" {
		h.respondWithError(w, http.StatusBadRequest, "message is required")
		return
	}

	// Check if client accepts streaming
	if strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		h.streamChat(w, r, sessionID, req)
		return
	}

	// Regular non-streaming chat
	response, err := h.service.Chat(sessionID, req.Message, req.Context)
	if err != nil {
		h.logger.Error("Failed to process chat", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to process chat")
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// streamChat handles streaming chat responses
func (h *Handler) streamChat(w http.ResponseWriter, r *http.Request, sessionID string, req ChatRequest) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Create flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.respondWithError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// Start streaming
	stream, err := h.service.StreamChat(sessionID, req.Message, req.Context)
	if err != nil {
		h.logger.Error("Failed to start chat stream", "error", err)
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Stream responses
	for response := range stream {
		if response.Error != "" {
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", response.Error)
		} else if response.Done {
			// Send final message with context
			data, _ := json.Marshal(map[string]interface{}{
				"done":    true,
				"context": response.Context,
			})
			fmt.Fprintf(w, "event: done\ndata: %s\n\n", string(data))
		} else {
			// Send message content
			fmt.Fprintf(w, "data: %s\n\n", response.Message.Content)
		}
		flusher.Flush()
	}
}

// GetModelsResponse represents the response for getting available models
type GetModelsResponse struct {
	Models []*domain.ModelInfo `json:"models"`
}

// GetAvailableModels handles GET /api/v1/aiops/models
func (h *Handler) GetAvailableModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.GetAvailableModels()
	if err != nil {
		h.logger.Error("Failed to get available models", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get available models")
		return
	}

	response := GetModelsResponse{
		Models: models,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetTokenUsageResponse represents the response for token usage
type GetTokenUsageResponse struct {
	Usage []*domain.ModelUsage `json:"usage"`
	Total int                 `json:"total"`
}

// GetTokenUsage handles GET /api/v1/aiops/usage
func (h *Handler) GetTokenUsage(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		h.respondWithError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}

	model := r.URL.Query().Get("model")
	limit := 10
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	usage, err := h.service.GetTokenUsage(workspaceID, model, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get token usage", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get token usage")
		return
	}

	response := GetTokenUsageResponse{
		Usage: usage,
		Total: len(usage),
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// respondWithError sends an error response
func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}

// respondWithJSON sends a JSON response
func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}