package api

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateProjectRequest represents the request payload for creating a project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateProjectRequest represents the request payload for updating a project
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateRoleRequest represents the request payload for creating a role
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Rules       string `json:"rules" binding:"required"`
}

// UpdateRoleRequest represents the request payload for updating a role
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Rules       *string `json:"rules,omitempty"`
}

// CreateRoleAssignmentRequest represents the request payload for creating a role assignment
type CreateRoleAssignmentRequest struct {
	GroupID string `json:"group_id" binding:"required"`
	RoleID  string `json:"role_id" binding:"required"`
}

// ProjectHandler handles project-related endpoints
type ProjectHandler struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// CreateProject creates a new project (namespace)
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Check specific validation errors
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "project name is required"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		}
		return
	}

	// Validate project name
	if len(strings.TrimSpace(req.Name)) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project name must be at least 3 characters"})
		return
	}

	// Validate project name format (Kubernetes namespace naming rules)
	if !isValidKubernetesName(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project name must be lowercase alphanumeric with hyphens"})
		return
	}

	// Check if project name already exists in workspace
	var existingProj db.Project
	if err := h.db.First(&existingProj, "workspace_id = ? AND name = ?", wsID, req.Name).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "project with this name already exists"})
		return
	}

	// Check project limit per workspace
	var projectCount int64
	if err := h.db.Model(&db.Project{}).Where("workspace_id = ?", wsID).Count(&projectCount).Error; err != nil {
		h.logger.Error("Failed to count projects", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check project limit"})
		return
	}

	// Get workspace plan to check limits
	var plan db.Plan
	if err := h.db.First(&plan, "id = ?", workspace.PlanID).Error; err != nil {
		h.logger.Error("Failed to get workspace plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check workspace plan"})
		return
	}

	if plan.MaxProjectsPerWorkspace != nil && projectCount >= int64(*plan.MaxProjectsPerWorkspace) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace has reached maximum number of projects"})
		return
	}

	// Create project
	project := &db.Project{
		WorkspaceID:     wsID,
		Name:            strings.TrimSpace(req.Name),
		Description:     req.Description,
		NamespaceStatus: "PENDING_CREATION",
	}

	if err := h.db.Create(project).Error; err != nil {
		h.logger.Error("Failed to create project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create project"})
		return
	}

	h.logger.Info("Project created successfully", 
		zap.String("project_id", project.ID),
		zap.String("ws_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusCreated, project)
}

// ListProjects lists projects in workspace
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	var projects []db.Project
	if err := h.db.Where("workspace_id = ?", wsID).Find(&projects).Error; err != nil {
		h.logger.Error("Failed to list projects", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
		return
	}

	h.logger.Info("List projects successful", 
		zap.String("ws_id", wsID),
		zap.String("user_id", userID.(string)),
		zap.Int("count", len(projects)))

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"total":    len(projects),
	})
}

// GetProject gets project details
func (h *ProjectHandler) GetProject(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	projectID := c.Param("projectId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	var project db.Project
	if err := h.db.First(&project, "id = ? AND workspace_id = ?", projectID, wsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		} else {
			h.logger.Error("Failed to get project", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get project"})
		}
		return
	}

	h.logger.Info("Get project successful", 
		zap.String("project_id", projectID),
		zap.String("ws_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, project)
}

// UpdateProject updates project details
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	projectID := c.Param("projectId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Check if trying to update name (not allowed)
	if req.Name != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project name cannot be changed"})
		return
	}

	// Get existing project
	var project db.Project
	if err := h.db.First(&project, "id = ? AND workspace_id = ?", projectID, wsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		} else {
			h.logger.Error("Failed to get project", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get project"})
		}
		return
	}

	// Track what needs to be updated
	updates := make(map[string]interface{})

	// Update description if provided
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	// Apply updates if any
	if len(updates) > 0 {
		if err := h.db.Model(&project).Updates(updates).Error; err != nil {
			h.logger.Error("Failed to update project", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project"})
			return
		}

		// Reload project to get updated data
		if err := h.db.First(&project, "id = ?", projectID).Error; err != nil {
			h.logger.Error("Failed to reload project", zap.Error(err))
		}
	}

	h.logger.Info("Update project successful", 
		zap.String("project_id", projectID),
		zap.String("ws_id", wsID),
		zap.String("user_id", userID.(string)),
		zap.Any("updates", updates))

	c.JSON(http.StatusOK, project)
}

// DeleteProject deletes a project
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	projectID := c.Param("projectId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Get existing project
	var project db.Project
	if err := h.db.First(&project, "id = ? AND workspace_id = ?", projectID, wsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		} else {
			h.logger.Error("Failed to get project", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get project"})
		}
		return
	}

	// Check if project has any roles
	var roleCount int64
	if err := h.db.Model(&db.Role{}).Where("project_id = ?", projectID).Count(&roleCount).Error; err != nil {
		h.logger.Error("Failed to count roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check project roles"})
		return
	}

	if roleCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete project with existing roles"})
		return
	}

	// Mark project for deletion
	if err := h.db.Model(&project).Update("namespace_status", "DELETING").Error; err != nil {
		h.logger.Error("Failed to mark project for deletion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
		return
	}

	h.logger.Info("Delete project successful", 
		zap.String("project_id", projectID),
		zap.String("ws_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{
		"message": "project deletion initiated",
		"status":  "DELETING",
	})
}

// hasOrgAccess checks if user has access to the organization
func (h *ProjectHandler) hasOrgAccess(userID, orgID string) bool {
	var membership db.OrganizationUser
	err := h.db.First(&membership, "user_id = ? AND organization_id = ?", userID, orgID).Error
	return err == nil
}

// isValidKubernetesName validates if a name follows Kubernetes naming conventions
func isValidKubernetesName(name string) bool {
	// Kubernetes DNS-1123 subdomain names must:
	// - contain no more than 253 characters
	// - contain only lowercase alphanumeric characters, '-' or '.'
	// - start with an alphanumeric character
	// - end with an alphanumeric character
	if len(name) > 253 {
		return false
	}
	
	// For project names, we'll be more restrictive: lowercase letters, numbers, and hyphens only
	// Must start and end with alphanumeric
	matched, _ := regexp.MatchString(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, name)
	return matched
}

// CreateRole creates a custom role in project
func (h *ProjectHandler) CreateRole(c *gin.Context) {
	projectID := c.Param("projectId")
	h.logger.Info("Create role endpoint called", zap.String("project_id", projectID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// ListRoles lists roles in project
func (h *ProjectHandler) ListRoles(c *gin.Context) {
	projectID := c.Param("projectId")
	h.logger.Info("List roles endpoint called", zap.String("project_id", projectID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// GetRole gets role details
func (h *ProjectHandler) GetRole(c *gin.Context) {
	projectID := c.Param("projectId")
	roleID := c.Param("roleId")
	h.logger.Info("Get role endpoint called", 
		zap.String("project_id", projectID),
		zap.String("role_id", roleID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// UpdateRole updates role details
func (h *ProjectHandler) UpdateRole(c *gin.Context) {
	projectID := c.Param("projectId")
	roleID := c.Param("roleId")
	h.logger.Info("Update role endpoint called", 
		zap.String("project_id", projectID),
		zap.String("role_id", roleID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// DeleteRole deletes a role
func (h *ProjectHandler) DeleteRole(c *gin.Context) {
	projectID := c.Param("projectId")
	roleID := c.Param("roleId")
	h.logger.Info("Delete role endpoint called", 
		zap.String("project_id", projectID),
		zap.String("role_id", roleID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// CreateRoleAssignment creates role assignment
func (h *ProjectHandler) CreateRoleAssignment(c *gin.Context) {
	projectID := c.Param("projectId")
	h.logger.Info("Create role assignment endpoint called", zap.String("project_id", projectID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// ListRoleAssignments lists role assignments
func (h *ProjectHandler) ListRoleAssignments(c *gin.Context) {
	projectID := c.Param("projectId")
	h.logger.Info("List role assignments endpoint called", zap.String("project_id", projectID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// DeleteRoleAssignment deletes role assignment
func (h *ProjectHandler) DeleteRoleAssignment(c *gin.Context) {
	projectID := c.Param("projectId")
	assignmentID := c.Param("assignmentId")
	h.logger.Info("Delete role assignment endpoint called", 
		zap.String("project_id", projectID),
		zap.String("assignment_id", assignmentID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}