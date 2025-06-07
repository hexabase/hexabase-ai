package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/workspace"
	"go.uber.org/zap"
)

// WorkspaceHandler handles workspace-related HTTP requests
type WorkspaceHandler struct {
	service workspace.Service
	logger  *zap.Logger
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(service workspace.Service, logger *zap.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{
		service: service,
		logger:  logger,
	}
}

// CreateWorkspace handles workspace creation
func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.GetString("user_id")

	var req workspace.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.OrganizationID = orgID
	req.CreatedBy = userID

	ws, task, err := h.service.CreateWorkspace(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("workspace created",
		zap.String("workspace_id", ws.ID),
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.String("task_id", task.ID))

	c.JSON(http.StatusCreated, gin.H{
		"workspace": ws,
		"task":      task,
	})
}

// GetWorkspace handles getting a workspace
func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	workspaceID := c.Param("wsId")

	ws, err := h.service.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get workspace", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	c.JSON(http.StatusOK, ws)
}

// ListWorkspaces handles listing workspaces
func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	orgID := c.Param("orgId")

	// Parse query parameters
	var filter workspace.WorkspaceFilter
	filter.Page = 1
	filter.PageSize = 20

	if page := c.Query("page"); page != "" {
		// Parse page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		// Parse page size
	}
	if status := c.Query("status"); status != "" {
		filter.Status = status
	}
	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	filter.OrganizationID = orgID
	result, err := h.service.ListWorkspaces(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list workspaces", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workspaces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspaces": result.Workspaces,
		"total":      result.Total,
		"page":       result.Page,
		"page_size":  result.PageSize,
	})
}

// UpdateWorkspace handles updating a workspace
func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.GetString("user_id")

	var req workspace.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.UpdatedBy = userID

	ws, err := h.service.UpdateWorkspace(c.Request.Context(), workspaceID, &req)
	if err != nil {
		h.logger.Error("failed to update workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("workspace updated",
		zap.String("workspace_id", workspaceID),
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, ws)
}

// DeleteWorkspace handles deleting a workspace
func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.GetString("user_id")

	task, err := h.service.DeleteWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to delete workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("workspace deletion initiated",
		zap.String("workspace_id", workspaceID),
		zap.String("user_id", userID),
		zap.String("task_id", task.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "workspace deletion initiated",
		"status":  "deleting",
		"task":    task,
	})
}

// GetKubeconfig handles getting workspace kubeconfig
func (h *WorkspaceHandler) GetKubeconfig(c *gin.Context) {
	workspaceID := c.Param("wsId")
	// userID := c.GetString("user_id") // Not used in the service method

	kubeconfig, err := h.service.GetKubeconfig(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get kubeconfig", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"kubeconfig": kubeconfig,
	})
}

// GetResourceUsage handles getting workspace resource usage
func (h *WorkspaceHandler) GetResourceUsage(c *gin.Context) {
	workspaceID := c.Param("wsId")

	usage, err := h.service.GetResourceUsage(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get resource usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// AddWorkspaceMember handles adding a member to workspace
func (h *WorkspaceHandler) AddWorkspaceMember(c *gin.Context) {
	workspaceID := c.Param("wsId")
	addedBy := c.GetString("user_id")

	var req workspace.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.AddedBy = addedBy

	member, err := h.service.AddWorkspaceMember(c.Request.Context(), workspaceID, &req)
	if err != nil {
		h.logger.Error("failed to add workspace member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "member added successfully",
		"member":  member,
	})
}

// RemoveWorkspaceMember handles removing a member from workspace
func (h *WorkspaceHandler) RemoveWorkspaceMember(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.Param("userId")

	err := h.service.RemoveWorkspaceMember(c.Request.Context(), workspaceID, userID)
	if err != nil {
		h.logger.Error("failed to remove workspace member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}

// ListWorkspaceMembers handles listing workspace members
func (h *WorkspaceHandler) ListWorkspaceMembers(c *gin.Context) {
	workspaceID := c.Param("wsId")

	members, err := h.service.ListWorkspaceMembers(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to list workspace members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   len(members),
	})
}