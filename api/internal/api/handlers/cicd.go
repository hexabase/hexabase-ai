package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/cicd"
	"log/slog"
)

// CICDHandler handles CI/CD related requests
type CICDHandler struct {
	service cicd.Service
	logger  *slog.Logger
}

// NewCICDHandler creates a new CI/CD handler
func NewCICDHandler(service cicd.Service, logger *slog.Logger) *CICDHandler {
	return &CICDHandler{
		service: service,
		logger:  logger,
	}
}

// CreatePipeline handles POST /api/v1/workspaces/{workspaceId}/pipelines
func (h *CICDHandler) CreatePipeline(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	var config cicd.PipelineConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("userID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Create pipeline
	run, err := h.service.CreatePipeline(c.Request.Context(), workspaceID, config)
	if err != nil {
		h.logger.Error("failed to create pipeline", "error", err, "workspaceId", workspaceID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pipeline"})
		return
	}

	c.JSON(http.StatusCreated, run)
}

// GetPipeline handles GET /api/v1/pipelines/{pipelineId}
func (h *CICDHandler) GetPipeline(c *gin.Context) {
	pipelineID := c.Param("pipelineId")

	run, err := h.service.GetPipeline(c.Request.Context(), pipelineID)
	if err != nil {
		h.logger.Error("failed to get pipeline", "error", err, "pipelineId", pipelineID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
		return
	}

	c.JSON(http.StatusOK, run)
}

// ListPipelines handles GET /api/v1/workspaces/{workspaceId}/pipelines
func (h *CICDHandler) ListPipelines(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	projectID := c.Query("projectId")
	
	// Parse limit
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	pipelines, err := h.service.ListPipelines(c.Request.Context(), workspaceID, projectID, limit)
	if err != nil {
		h.logger.Error("failed to list pipelines", "error", err, "workspaceId", workspaceID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pipelines"})
		return
	}

	c.JSON(http.StatusOK, pipelines)
}

// CancelPipeline handles POST /api/v1/pipelines/{pipelineId}/cancel
func (h *CICDHandler) CancelPipeline(c *gin.Context) {
	pipelineID := c.Param("pipelineId")

	if err := h.service.CancelPipeline(c.Request.Context(), pipelineID); err != nil {
		h.logger.Error("failed to cancel pipeline", "error", err, "pipelineId", pipelineID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel pipeline"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pipeline cancelled"})
}

// DeletePipeline handles DELETE /api/v1/pipelines/{pipelineId}
func (h *CICDHandler) DeletePipeline(c *gin.Context) {
	pipelineID := c.Param("pipelineId")

	if err := h.service.DeletePipeline(c.Request.Context(), pipelineID); err != nil {
		h.logger.Error("failed to delete pipeline", "error", err, "pipelineId", pipelineID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pipeline"})
		return
	}

	c.Status(http.StatusNoContent)
}

// RetryPipeline handles POST /api/v1/pipelines/{pipelineId}/retry
func (h *CICDHandler) RetryPipeline(c *gin.Context) {
	pipelineID := c.Param("pipelineId")

	run, err := h.service.RetryPipeline(c.Request.Context(), pipelineID)
	if err != nil {
		h.logger.Error("failed to retry pipeline", "error", err, "pipelineId", pipelineID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry pipeline"})
		return
	}

	c.JSON(http.StatusCreated, run)
}

// GetPipelineLogs handles GET /api/v1/pipelines/{pipelineId}/logs
func (h *CICDHandler) GetPipelineLogs(c *gin.Context) {
	pipelineID := c.Param("pipelineId")
	stage := c.Query("stage")
	task := c.Query("task")

	logs, err := h.service.GetPipelineLogs(c.Request.Context(), pipelineID, stage, task)
	if err != nil {
		h.logger.Error("failed to get pipeline logs", "error", err, "pipelineId", pipelineID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pipeline logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// StreamPipelineLogs handles GET /api/v1/pipelines/{pipelineId}/logs/stream
func (h *CICDHandler) StreamPipelineLogs(c *gin.Context) {
	pipelineID := c.Param("pipelineId")
	stage := c.Query("stage")
	task := c.Query("task")

	// Set headers for streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	logStream, err := h.service.StreamPipelineLogs(c.Request.Context(), pipelineID, stage, task)
	if err != nil {
		h.logger.Error("failed to stream pipeline logs", "error", err, "pipelineId", pipelineID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream pipeline logs"})
		return
	}
	defer logStream.Close()

	// Stream logs to client
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := logStream.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				h.logger.Error("failed to write log stream", "error", writeErr)
				return false
			}
		}
		return err == nil
	})
}

// ListTemplates handles GET /api/v1/pipelines/templates
func (h *CICDHandler) ListTemplates(c *gin.Context) {
	provider := c.Query("provider")

	templates, err := h.service.ListTemplates(c.Request.Context(), provider)
	if err != nil {
		h.logger.Error("failed to list templates", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// GetTemplate handles GET /api/v1/pipelines/templates/{templateId}
func (h *CICDHandler) GetTemplate(c *gin.Context) {
	templateID := c.Param("templateId")

	template, err := h.service.GetTemplate(c.Request.Context(), templateID)
	if err != nil {
		h.logger.Error("failed to get template", "error", err, "templateId", templateID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// CreatePipelineFromTemplate handles POST /api/v1/workspaces/{workspaceId}/pipelines/from-template
func (h *CICDHandler) CreatePipelineFromTemplate(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	var req struct {
		TemplateID string         `json:"templateId"`
		Parameters map[string]any `json:"parameters"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	run, err := h.service.CreatePipelineFromTemplate(c.Request.Context(), workspaceID, req.TemplateID, req.Parameters)
	if err != nil {
		h.logger.Error("failed to create pipeline from template", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pipeline from template"})
		return
	}

	c.JSON(http.StatusCreated, run)
}

// CreateGitCredential handles POST /api/v1/workspaces/{workspaceId}/credentials/git
func (h *CICDHandler) CreateGitCredential(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	var req struct {
		Name       string               `json:"name"`
		Credential cicd.GitCredential   `json:"credential"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.CreateGitCredential(c.Request.Context(), workspaceID, req.Name, req.Credential); err != nil {
		h.logger.Error("failed to create git credential", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create git credential"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Git credential created"})
}

// CreateRegistryCredential handles POST /api/v1/workspaces/{workspaceId}/credentials/registry
func (h *CICDHandler) CreateRegistryCredential(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	var req struct {
		Name       string                    `json:"name"`
		Credential cicd.RegistryCredential   `json:"credential"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.CreateRegistryCredential(c.Request.Context(), workspaceID, req.Name, req.Credential); err != nil {
		h.logger.Error("failed to create registry credential", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create registry credential"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registry credential created"})
}

// ListCredentials handles GET /api/v1/workspaces/{workspaceId}/credentials
func (h *CICDHandler) ListCredentials(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	credentials, err := h.service.ListCredentials(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to list credentials", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list credentials"})
		return
	}

	c.JSON(http.StatusOK, credentials)
}

// DeleteCredential handles DELETE /api/v1/workspaces/{workspaceId}/credentials/{credentialName}
func (h *CICDHandler) DeleteCredential(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	credentialName := c.Param("credentialName")

	if err := h.service.DeleteCredential(c.Request.Context(), workspaceID, credentialName); err != nil {
		h.logger.Error("failed to delete credential", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete credential"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListProviders handles GET /api/v1/providers
func (h *CICDHandler) ListProviders(c *gin.Context) {
	providers, err := h.service.ListProviders(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list providers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list providers"})
		return
	}

	c.JSON(http.StatusOK, providers)
}

// GetProviderConfig handles GET /api/v1/workspaces/{workspaceId}/provider-config
func (h *CICDHandler) GetProviderConfig(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	config, err := h.service.GetProviderConfig(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get provider config", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// SetProviderConfig handles PUT /api/v1/workspaces/{workspaceId}/provider-config
func (h *CICDHandler) SetProviderConfig(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	var config cicd.ProviderConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.SetProviderConfig(c.Request.Context(), workspaceID, config); err != nil {
		h.logger.Error("failed to set provider config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set provider config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider config updated"})
}