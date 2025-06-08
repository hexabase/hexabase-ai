package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/workspace"
)

// InternalHandler handles internal-only API requests.
type InternalHandler struct {
	workspaceSvc workspace.Service
	logger       *slog.Logger
}

// NewInternalHandler creates a new handler for internal operations.
func NewInternalHandler(workspaceSvc workspace.Service, logger *slog.Logger) *InternalHandler {
	return &InternalHandler{
		workspaceSvc: workspaceSvc,
		logger:       logger,
	}
}

// GetNodes is the handler for GET /internal/v1/workspaces/:workspaceId/nodes
func (h *InternalHandler) GetNodes(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	nodes, err := h.workspaceSvc.GetNodes(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get nodes for workspace", "workspace_id", workspaceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve node information"})
		return
	}

	c.JSON(http.StatusOK, nodes)
}

func (h *InternalHandler) ScaleDeployment(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	deploymentName := c.Param("deploymentName")

	var req struct {
		Replicas int `json:"replicas"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: replicas field is required"})
		return
	}

	if req.Replicas < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "replicas must be a non-negative integer"})
		return
	}

	err := h.workspaceSvc.ScaleDeployment(c.Request.Context(), workspaceID, deploymentName, req.Replicas)
	if err != nil {
		h.logger.Error("failed to scale deployment", "workspace_id", workspaceID, "deployment", deploymentName, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scale deployment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deployment scaled successfully"})
} 