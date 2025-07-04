package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/monitoring/domain"
)

// Handler handles monitoring-related HTTP requests
type Handler struct {
	service domain.Service
	logger  *slog.Logger
}

// NewHandler creates a new monitoring handler
func NewHandler(service domain.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetMetrics handles GET /api/v1/workspaces/:workspace_id/metrics
func (h *Handler) GetMetrics(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	period := c.DefaultQuery("period", "1h")

	// Validate workspace access
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Create query options
	opts := domain.QueryOptions{
		Period: period,
	}

	// Get metrics from service
	metrics, err := h.service.GetWorkspaceMetrics(c.Request.Context(), workspaceID, opts)
	if err != nil {
		h.logger.Error("Failed to get workspace metrics",
			"workspace_id", workspaceID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetClusterHealth handles GET /api/v1/workspaces/:workspace_id/health
func (h *Handler) GetClusterHealth(c *gin.Context) {
	workspaceID := c.Param("workspace_id")

	health, err := h.service.GetClusterHealth(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("Failed to get cluster health",
			"workspace_id", workspaceID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check cluster health"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetResourceUsage handles GET /api/v1/workspaces/:workspace_id/resources
func (h *Handler) GetResourceUsage(c *gin.Context) {
	workspaceID := c.Param("workspace_id")

	usage, err := h.service.GetResourceUsage(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("Failed to get resource usage",
			"workspace_id", workspaceID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve resource usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// GetAlerts handles GET /api/v1/workspaces/:workspace_id/alerts
func (h *Handler) GetAlerts(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	severity := c.Query("severity")

	alerts, err := h.service.GetAlerts(c.Request.Context(), workspaceID, severity)
	if err != nil {
		h.logger.Error("Failed to get alerts",
			"workspace_id", workspaceID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// CreateAlert handles POST /api/v1/workspaces/:workspace_id/alerts
func (h *Handler) CreateAlert(c *gin.Context) {
	workspaceID := c.Param("workspace_id")

	var req struct {
		Type        string  `json:"type" binding:"required"`
		Severity    string  `json:"severity" binding:"required,oneof=critical warning info"`
		Title       string  `json:"title" binding:"required"`
		Description string  `json:"description"`
		Resource    string  `json:"resource"`
		Threshold   float64 `json:"threshold"`
		Value       float64 `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert := &domain.Alert{
		WorkspaceID: workspaceID,
		Type:        req.Type,
		Severity:    req.Severity,
		Title:       req.Title,
		Description: req.Description,
		Resource:    req.Resource,
		Threshold:   req.Threshold,
		Value:       req.Value,
	}

	if err := h.service.CreateAlert(c.Request.Context(), alert); err != nil {
		h.logger.Error("Failed to create alert",
			"workspace_id", workspaceID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create alert"})
		return
	}

	c.JSON(http.StatusCreated, alert)
}

// AcknowledgeAlert handles PUT /api/v1/alerts/:alert_id/acknowledge
func (h *Handler) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("alert_id")
	userID := c.GetString("user_id")

	if err := h.service.AcknowledgeAlert(c.Request.Context(), alertID, userID); err != nil {
		h.logger.Error("Failed to acknowledge alert",
			"alert_id", alertID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to acknowledge alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert acknowledged"})
}

// ResolveAlert handles PUT /api/v1/alerts/:alert_id/resolve
func (h *Handler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("alert_id")

	if err := h.service.ResolveAlert(c.Request.Context(), alertID); err != nil {
		h.logger.Error("Failed to resolve alert",
			"alert_id", alertID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert resolved"})
}

// GetWorkspaceOverview handles GET /api/v1/organizations/:org_id/monitoring/overview
func (h *Handler) GetWorkspaceOverview(c *gin.Context) {
	orgID := c.Param("org_id")
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	// This would typically aggregate data across all workspaces in the organization
	// For now, return a structured response
	c.JSON(http.StatusOK, gin.H{
		"organization_id": orgID,
		"summary": gin.H{
			"total_workspaces": 5,
			"healthy_workspaces": 4,
			"alerts_count": gin.H{
				"critical": 1,
				"warning": 3,
				"info": 5,
			},
			"resource_usage": gin.H{
				"cpu_percentage": 65,
				"memory_percentage": 42,
				"storage_percentage": 38,
			},
		},
		"workspaces": []gin.H{
			// Would be populated with actual workspace data
		},
		"recent_alerts": []gin.H{
			// Would be populated with recent alerts
		},
		"limit": limit,
	})
}