package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/project"
	"go.uber.org/zap"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	service project.Service
	logger  *zap.Logger
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(service project.Service, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		service: service,
		logger:  logger,
	}
}

// CreateProject handles project creation
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.GetString("user_id")

	var req project.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.WorkspaceID = workspaceID
	req.CreatedBy = userID

	proj, err := h.service.CreateProject(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("project created",
		zap.String("project_id", proj.ID),
		zap.String("workspace_id", workspaceID),
		zap.String("user_id", userID))

	c.JSON(http.StatusCreated, proj)
}

// GetProject handles getting a project
func (h *ProjectHandler) GetProject(c *gin.Context) {
	projectID := c.Param("projectId")

	proj, err := h.service.GetProject(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get project", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, proj)
}

// ListProjects handles listing projects in a workspace
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	workspaceID := c.Param("wsId")

	projects, err := h.service.ListProjects(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to list projects", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"total":    len(projects),
	})
}

// UpdateProject handles updating a project
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.GetString("user_id")

	var req project.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.UpdatedBy = userID

	proj, err := h.service.UpdateProject(c.Request.Context(), projectID, &req)
	if err != nil {
		h.logger.Error("failed to update project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("project updated",
		zap.String("project_id", projectID),
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, proj)
}

// DeleteProject handles deleting a project
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.GetString("user_id")

	err := h.service.DeleteProject(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to delete project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("project deleted",
		zap.String("project_id", projectID),
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{"message": "project deleted successfully"})
}

// CreateSubProject handles creating a sub-project
func (h *ProjectHandler) CreateSubProject(c *gin.Context) {
	parentID := c.Param("projectId")
	userID := c.GetString("user_id")

	var req project.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.CreatedBy = userID

	proj, err := h.service.CreateSubProject(c.Request.Context(), parentID, &req)
	if err != nil {
		h.logger.Error("failed to create sub-project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, proj)
}

// GetProjectHierarchy handles getting project hierarchy
func (h *ProjectHandler) GetProjectHierarchy(c *gin.Context) {
	projectID := c.Param("projectId")

	hierarchy, err := h.service.GetProjectHierarchy(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get project hierarchy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, hierarchy)
}

// ApplyResourceQuota handles applying resource quota to a project
func (h *ProjectHandler) ApplyResourceQuota(c *gin.Context) {
	projectID := c.Param("projectId")

	var quota project.ResourceQuota
	if err := c.ShouldBindJSON(&quota); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := h.service.ApplyResourceQuota(c.Request.Context(), projectID, &quota)
	if err != nil {
		h.logger.Error("failed to apply resource quota", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "resource quota applied successfully"})
}

// GetResourceUsage handles getting project resource usage
func (h *ProjectHandler) GetResourceUsage(c *gin.Context) {
	projectID := c.Param("projectId")

	usage, err := h.service.GetResourceUsage(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get resource usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// AddProjectMember handles adding a member to project
func (h *ProjectHandler) AddProjectMember(c *gin.Context) {
	projectID := c.Param("projectId")
	addedBy := c.GetString("user_id")

	var req project.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.AddedBy = addedBy

	err := h.service.AddProjectMember(c.Request.Context(), projectID, &req)
	if err != nil {
		h.logger.Error("failed to add project member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added successfully"})
}

// RemoveProjectMember handles removing a member from project
func (h *ProjectHandler) RemoveProjectMember(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.Param("userId")

	err := h.service.RemoveProjectMember(c.Request.Context(), projectID, userID)
	if err != nil {
		h.logger.Error("failed to remove project member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}

// ListProjectMembers handles listing project members
func (h *ProjectHandler) ListProjectMembers(c *gin.Context) {
	projectID := c.Param("projectId")

	members, err := h.service.ListProjectMembers(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list project members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   len(members),
	})
}

// GetActivityLogs handles getting project activity logs
func (h *ProjectHandler) GetActivityLogs(c *gin.Context) {
	projectID := c.Param("projectId")

	var filter project.ActivityFilter
	// Parse query parameters for filtering

	logs, err := h.service.GetActivityLogs(c.Request.Context(), projectID, filter)
	if err != nil {
		h.logger.Error("failed to get activity logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": logs,
		"total":      len(logs),
	})
}