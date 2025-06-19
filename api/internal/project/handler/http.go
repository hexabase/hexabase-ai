package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/project/domain"
)

// Handler handles project-related HTTP requests
type Handler struct {
	service domain.Service
	logger  *slog.Logger
}

// NewHandler creates a new project handler
func NewHandler(service domain.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateProject handles project creation
func (h *Handler) CreateProject(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.GetString("user_id")

	var req domain.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.WorkspaceID = workspaceID
	req.CreatedBy = userID

	proj, err := h.service.CreateProject(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create project", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("project created",
		"project_id", proj.ID,
		"workspace_id", workspaceID,
		"user_id", userID)

	c.JSON(http.StatusCreated, proj)
}

// GetProject handles getting a project
func (h *Handler) GetProject(c *gin.Context) {
	projectID := c.Param("projectId")

	proj, err := h.service.GetProject(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get project", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, proj)
}

// ListProjects handles listing projects in a workspace
func (h *Handler) ListProjects(c *gin.Context) {
	workspaceID := c.Param("wsId")

	// Parse query parameters for filtering
	filter := domain.ProjectFilter{
		WorkspaceID: workspaceID,
		Status:      c.Query("status"),
		Search:      c.Query("search"),
		SortBy:      c.DefaultQuery("sort_by", "created_at"),
		SortOrder:   c.DefaultQuery("sort_order", "desc"),
	}
	
	// Parse pagination
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	filter.Page = page
	
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	filter.PageSize = pageSize

	result, err := h.service.ListProjects(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list projects", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateProject handles updating a project
func (h *Handler) UpdateProject(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.GetString("user_id")

	var req domain.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.UpdatedBy = userID

	proj, err := h.service.UpdateProject(c.Request.Context(), projectID, &req)
	if err != nil {
		h.logger.Error("failed to update project", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("project updated",
		"project_id", projectID,
		"user_id", userID)

	c.JSON(http.StatusOK, proj)
}

// DeleteProject handles deleting a project
func (h *Handler) DeleteProject(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.GetString("user_id")

	err := h.service.DeleteProject(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to delete project", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("project deleted",
		"project_id", projectID,
		"user_id", userID)

	c.JSON(http.StatusOK, gin.H{"message": "project deleted successfully"})
}

// CreateSubProject handles creating a sub-project
func (h *Handler) CreateSubProject(c *gin.Context) {
	parentID := c.Param("projectId")
	userID := c.GetString("user_id")

	var req domain.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.CreatedBy = userID

	proj, err := h.service.CreateSubProject(c.Request.Context(), parentID, &req)
	if err != nil {
		h.logger.Error("failed to create sub-project", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, proj)
}

// GetProjectHierarchy handles getting project hierarchy
func (h *Handler) GetProjectHierarchy(c *gin.Context) {
	projectID := c.Param("projectId")

	hierarchy, err := h.service.GetProjectHierarchy(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get project hierarchy", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, hierarchy)
}

// ApplyResourceQuota handles applying resource quota to a project
func (h *Handler) ApplyResourceQuota(c *gin.Context) {
	projectID := c.Param("projectId")

	var quota domain.ResourceQuota
	if err := c.ShouldBindJSON(&quota); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := h.service.ApplyResourceQuota(c.Request.Context(), projectID, &quota)
	if err != nil {
		h.logger.Error("failed to apply resource quota", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("resource quota applied",
		"project_id", projectID,
		"quota", quota)

	c.JSON(http.StatusOK, gin.H{"message": "resource quota applied successfully"})
}

// GetResourceUsage handles getting resource usage for a project
func (h *Handler) GetResourceUsage(c *gin.Context) {
	projectID := c.Param("projectId")

	usage, err := h.service.GetResourceUsage(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get resource usage", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// AddProjectMember handles adding a member to a project
func (h *Handler) AddProjectMember(c *gin.Context) {
	projectID := c.Param("projectId")

	var req domain.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := h.service.AddProjectMember(c.Request.Context(), projectID, &req)
	if err != nil {
		h.logger.Error("failed to add project member", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added successfully"})
}

// RemoveProjectMember handles removing a member from a project
func (h *Handler) RemoveProjectMember(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.Param("userId")

	err := h.service.RemoveProjectMember(c.Request.Context(), projectID, userID)
	if err != nil {
		h.logger.Error("failed to remove project member", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}

// ListProjectMembers handles listing project members
func (h *Handler) ListProjectMembers(c *gin.Context) {
	projectID := c.Param("projectId")

	members, err := h.service.ListProjectMembers(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list project members", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// GetActivityLogs handles getting activity logs for a project
func (h *Handler) GetActivityLogs(c *gin.Context) {
	projectID := c.Param("projectId")

	// Create default filter
	filter := domain.ActivityFilter{
		ProjectID: projectID,
		PageSize: 50, // Default limit
	}

	logs, err := h.service.GetActivityLogs(c.Request.Context(), projectID, filter)
	if err != nil {
		h.logger.Error("failed to get activity logs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}