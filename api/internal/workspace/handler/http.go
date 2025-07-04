package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
)

// Handler handles workspace-related HTTP requests
type Handler struct {
	service domain.Service
	logger  *slog.Logger
}

// NewHandler creates a new workspace handler
func NewHandler(service domain.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateWorkspace handles workspace creation
func (h *Handler) CreateWorkspace(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.GetString("user_id")

	var req domain.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.OrganizationID = orgID
	req.CreatedBy = userID

	ws, task, err := h.service.CreateWorkspace(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create workspace", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("workspace created",
		"workspace_id", ws.ID,
		"org_id", orgID,
		"user_id", userID,
		"task_id", task.ID)

	c.JSON(http.StatusCreated, gin.H{
		"workspace": ws,
		"task":      task,
	})
}

// GetWorkspace handles getting a workspace
func (h *Handler) GetWorkspace(c *gin.Context) {
	workspaceID := c.Param("wsId")

	ws, err := h.service.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get workspace", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	c.JSON(http.StatusOK, ws)
}

// ListWorkspaces handles listing workspaces
func (h *Handler) ListWorkspaces(c *gin.Context) {
	orgID := c.Param("orgId")

	// Parse query parameters
	var filter domain.WorkspaceFilter
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
		h.logger.Error("failed to list workspaces", "error", err)
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
func (h *Handler) UpdateWorkspace(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.GetString("user_id")

	var req domain.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.UpdatedBy = userID

	ws, err := h.service.UpdateWorkspace(c.Request.Context(), workspaceID, &req)
	if err != nil {
		h.logger.Error("failed to update workspace", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("workspace updated",
		"workspace_id", workspaceID,
		"user_id", userID)

	c.JSON(http.StatusOK, ws)
}

// DeleteWorkspace handles deleting a workspace
func (h *Handler) DeleteWorkspace(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.GetString("user_id")

	task, err := h.service.DeleteWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to delete workspace", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("workspace deletion initiated",
		"workspace_id", workspaceID,
		"user_id", userID,
		"task_id", task.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "workspace deletion initiated",
		"status":  "deleting",
		"task":    task,
	})
}

// GetKubeconfig handles getting workspace kubeconfig
func (h *Handler) GetKubeconfig(c *gin.Context) {
	workspaceID := c.Param("wsId")
	// userID := c.GetString("user_id") // Not used in the service method

	kubeconfig, err := h.service.GetKubeconfig(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get kubeconfig", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"kubeconfig": kubeconfig,
	})
}

// GetResourceUsage handles getting workspace resource usage
func (h *Handler) GetResourceUsage(c *gin.Context) {
	workspaceID := c.Param("wsId")

	usage, err := h.service.GetResourceUsage(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get resource usage", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// AddWorkspaceMember handles adding a member to workspace
func (h *Handler) AddWorkspaceMember(c *gin.Context) {
	workspaceID := c.Param("wsId")
	addedBy := c.GetString("user_id")

	var req domain.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.AddedBy = addedBy

	member, err := h.service.AddWorkspaceMember(c.Request.Context(), workspaceID, &req)
	if err != nil {
		h.logger.Error("failed to add workspace member", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "member added successfully",
		"member":  member,
	})
}

// RemoveWorkspaceMember handles removing a member from workspace
func (h *Handler) RemoveWorkspaceMember(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.Param("userId")

	err := h.service.RemoveWorkspaceMember(c.Request.Context(), workspaceID, userID)
	if err != nil {
		h.logger.Error("failed to remove workspace member", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}

// ListWorkspaceMembers handles listing workspace members
func (h *Handler) ListWorkspaceMembers(c *gin.Context) {
	workspaceID := c.Param("wsId")

	members, err := h.service.ListWorkspaceMembers(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to list workspace members", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   len(members),
	})
}