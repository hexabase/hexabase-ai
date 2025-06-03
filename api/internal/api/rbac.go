package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PolicyRule represents a Kubernetes RBAC policy rule
type PolicyRule struct {
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// RBACHandler handles role-based access control endpoints
type RBACHandler struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

// NewRBACHandler creates a new RBAC handler
func NewRBACHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *RBACHandler {
	return &RBACHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// Request and response types
type CreateRBACRoleRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Scope       string                 `json:"scope" binding:"required"` // "namespace" or "cluster"
	Rules       []PolicyRule    `json:"rules" binding:"required"`
	ProjectID   *string                `json:"project_id,omitempty"`
}

type UpdateRBACRoleRequest struct {
	Name        *string              `json:"name,omitempty"`
	Description *string              `json:"description,omitempty"`
	Rules       []PolicyRule  `json:"rules,omitempty"`
	IsActive    *bool                `json:"is_active,omitempty"`
}

type CreateRoleBindingRequest struct {
	RoleID      string  `json:"role_id"`
	ProjectID   *string `json:"project_id,omitempty"`
	SubjectType string  `json:"subject_type"` // "User" or "Group"
	SubjectID   string  `json:"subject_id"`
	SubjectName string  `json:"subject_name"`
}

type CheckPermissionsRequest struct {
	UserID    string  `json:"user_id" binding:"required"`
	APIGroup  string  `json:"api_group"`
	Resource  string  `json:"resource" binding:"required"`
	Verb      string  `json:"verb" binding:"required"`
	Namespace *string `json:"namespace,omitempty"`
}

type CheckPermissionsResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

// Role Management

// CreateRole creates a new role
func (h *RBACHandler) CreateRole(c *gin.Context) {
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

	var req CreateRBACRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		if req.Scope == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "scope is required"})
			return
		}
		if len(req.Rules) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rules are required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Validate scope
	if req.Scope != "namespace" && req.Scope != "cluster" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.Where("id = ? AND organization_id = ?", wsID, orgID).First(&workspace).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// If project_id is provided, verify it exists
	if req.ProjectID != nil {
		var project db.Project
		if err := h.db.Where("id = ? AND workspace_id = ?", *req.ProjectID, wsID).First(&project).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			} else {
				h.logger.Error("Failed to get project", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get project"})
			}
			return
		}
	}

	// Check if role already exists
	var existingRole db.Role
	whereClause := "name = ? AND workspace_id = ?"
	if req.ProjectID != nil {
		whereClause += " AND project_id = ?"
		if err := h.db.Where(whereClause, req.Name, wsID, *req.ProjectID).First(&existingRole).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "role already exists"})
			return
		}
	} else {
		if err := h.db.Where(whereClause+" AND project_id IS NULL", req.Name, wsID).First(&existingRole).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "role already exists"})
			return
		}
	}

	// Serialize rules
	rulesJSON, err := json.Marshal(req.Rules)
	if err != nil {
		h.logger.Error("Failed to serialize rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize rules"})
		return
	}

	// Create role
	role := &db.Role{
		WorkspaceID: &wsID,
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		Rules:       string(rulesJSON),
		Scope:       req.Scope,
		IsCustom:    true,
		IsActive:    true,
	}

	if err := h.db.Create(role).Error; err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create role"})
		return
	}

	h.logger.Info("Role created successfully",
		zap.String("role_id", role.ID),
		zap.String("name", role.Name),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusCreated, role)
}

// ListRoles lists all roles for a workspace
func (h *RBACHandler) ListRoles(c *gin.Context) {
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

	var roles []db.Role
	query := h.db.Where("workspace_id = ? AND is_active = true", wsID)

	// Filter by scope if provided
	if scope := c.Query("scope"); scope != "" {
		query = query.Where("scope = ?", scope)
	}

	if err := query.Order("created_at DESC").Find(&roles).Error; err != nil {
		h.logger.Error("Failed to list roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"total": len(roles),
	})
}

// GetRole gets a specific role
func (h *RBACHandler) GetRole(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	roleID := c.Param("roleId")
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

	var role db.Role
	if err := h.db.Where("id = ? AND workspace_id = ?", roleID, wsID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		} else {
			h.logger.Error("Failed to get role", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		}
		return
	}

	c.JSON(http.StatusOK, role)
}

// UpdateRole updates a role
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	roleID := c.Param("roleId")
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

	var req UpdateRBACRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	var role db.Role
	if err := h.db.Where("id = ? AND workspace_id = ?", roleID, wsID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		} else {
			h.logger.Error("Failed to get role", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		}
		return
	}

	// Update fields
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Rules != nil {
		rulesJSON, err := json.Marshal(req.Rules)
		if err != nil {
			h.logger.Error("Failed to serialize rules", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize rules"})
			return
		}
		role.Rules = string(rulesJSON)
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	if err := h.db.Save(&role).Error; err != nil {
		h.logger.Error("Failed to update role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	h.logger.Info("Role updated successfully",
		zap.String("role_id", roleID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, role)
}

// DeleteRole deletes a role
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	roleID := c.Param("roleId")
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

	var role db.Role
	if err := h.db.Where("id = ? AND workspace_id = ?", roleID, wsID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		} else {
			h.logger.Error("Failed to get role", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		}
		return
	}

	// Soft delete - mark as inactive
	role.IsActive = false
	if err := h.db.Save(&role).Error; err != nil {
		h.logger.Error("Failed to delete role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete role"})
		return
	}

	h.logger.Info("Role deleted successfully",
		zap.String("role_id", roleID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{"message": "role deleted successfully"})
}

// Role Binding Management

// CreateRoleBinding creates a new role binding
func (h *RBACHandler) CreateRoleBinding(c *gin.Context) {
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

	var req CreateRoleBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Manual validation for better error messages
	if req.RoleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role_id is required"})
		return
	}
	if req.SubjectType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject_type is required"})
		return
	}

	// Validate subject type
	if req.SubjectType != "User" && req.SubjectType != "Group" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject_type"})
		return
	}

	// Verify role exists and belongs to workspace
	var role db.Role
	if err := h.db.Where("id = ? AND workspace_id = ? AND is_active = true", req.RoleID, wsID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		} else {
			h.logger.Error("Failed to get role", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		}
		return
	}

	// If project_id is provided, verify it exists
	if req.ProjectID != nil {
		var project db.Project
		if err := h.db.Where("id = ? AND workspace_id = ?", *req.ProjectID, wsID).First(&project).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			} else {
				h.logger.Error("Failed to get project", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get project"})
			}
			return
		}
	}

	// Check if role binding already exists
	var existingBinding db.RoleBinding
	whereClause := "workspace_id = ? AND role_id = ? AND subject_type = ? AND subject_id = ?"
	args := []interface{}{wsID, req.RoleID, req.SubjectType, req.SubjectID}
	
	if req.ProjectID != nil {
		whereClause += " AND project_id = ?"
		args = append(args, *req.ProjectID)
	} else {
		whereClause += " AND project_id IS NULL"
	}

	if err := h.db.Where(whereClause, args...).First(&existingBinding).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role binding already exists"})
		return
	}

	// Create role binding
	binding := &db.RoleBinding{
		WorkspaceID: wsID,
		ProjectID:   req.ProjectID,
		RoleID:      req.RoleID,
		SubjectType: req.SubjectType,
		SubjectID:   req.SubjectID,
		SubjectName: req.SubjectName,
		IsActive:    true,
	}

	if err := h.db.Create(binding).Error; err != nil {
		h.logger.Error("Failed to create role binding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create role binding"})
		return
	}

	h.logger.Info("Role binding created successfully",
		zap.String("binding_id", binding.ID),
		zap.String("role_id", req.RoleID),
		zap.String("subject_type", req.SubjectType),
		zap.String("subject_id", req.SubjectID),
		zap.String("workspace_id", wsID))

	c.JSON(http.StatusCreated, binding)
}

// ListRoleBindings lists all role bindings for a workspace
func (h *RBACHandler) ListRoleBindings(c *gin.Context) {
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

	var bindings []db.RoleBinding
	query := h.db.Where("workspace_id = ? AND is_active = true", wsID)

	// Filter by subject type if provided
	if subjectType := c.Query("subject_type"); subjectType != "" {
		query = query.Where("subject_type = ?", subjectType)
	}

	// Filter by project if provided
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	if err := query.Order("created_at DESC").Preload("Role").Find(&bindings).Error; err != nil {
		h.logger.Error("Failed to list role bindings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list role bindings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role_bindings": bindings,
		"total":         len(bindings),
	})
}

// GetRoleBinding gets a specific role binding
func (h *RBACHandler) GetRoleBinding(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	bindingID := c.Param("bindingId")
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

	var binding db.RoleBinding
	if err := h.db.Where("id = ? AND workspace_id = ?", bindingID, wsID).Preload("Role").First(&binding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role binding not found"})
		} else {
			h.logger.Error("Failed to get role binding", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role binding"})
		}
		return
	}

	c.JSON(http.StatusOK, binding)
}

// DeleteRoleBinding deletes a role binding
func (h *RBACHandler) DeleteRoleBinding(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	bindingID := c.Param("bindingId")
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

	var binding db.RoleBinding
	if err := h.db.Where("id = ? AND workspace_id = ?", bindingID, wsID).First(&binding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role binding not found"})
		} else {
			h.logger.Error("Failed to get role binding", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role binding"})
		}
		return
	}

	// Soft delete - mark as inactive
	binding.IsActive = false
	if err := h.db.Save(&binding).Error; err != nil {
		h.logger.Error("Failed to delete role binding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete role binding"})
		return
	}

	h.logger.Info("Role binding deleted successfully",
		zap.String("binding_id", bindingID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{"message": "role binding deleted successfully"})
}

// Permission Checking

// CheckPermissions checks if a user has specific permissions
func (h *RBACHandler) CheckPermissions(c *gin.Context) {
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

	var req CheckPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Get all role bindings for the user
	var userBindings []db.RoleBinding
	if err := h.db.Where("workspace_id = ? AND subject_type = ? AND subject_id = ? AND is_active = true", 
		wsID, "User", req.UserID).Preload("Role").Find(&userBindings).Error; err != nil {
		h.logger.Error("Failed to get user role bindings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}

	// Get group memberships for the user
	var groupMemberships []db.GroupMembership
	if err := h.db.Where("user_id = ?", req.UserID).Find(&groupMemberships).Error; err != nil {
		h.logger.Error("Failed to get user group memberships", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}

	// Get role bindings for user's groups
	var groupBindings []db.RoleBinding
	if len(groupMemberships) > 0 {
		var groupIDs []string
		for _, membership := range groupMemberships {
			groupIDs = append(groupIDs, membership.GroupID)
		}
		
		if err := h.db.Where("workspace_id = ? AND subject_type = ? AND subject_id IN ? AND is_active = true", 
			wsID, "Group", groupIDs).Preload("Role").Find(&groupBindings).Error; err != nil {
			h.logger.Error("Failed to get group role bindings", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
			return
		}
	}

	// Combine all bindings
	allBindings := append(userBindings, groupBindings...)

	// Check permissions against all roles
	allowed := false
	reason := "No matching permissions found"

	for _, binding := range allBindings {
		var rules []PolicyRule
		if err := json.Unmarshal([]byte(binding.Role.Rules), &rules); err != nil {
			h.logger.Error("Failed to parse role rules", zap.Error(err))
			continue
		}

		if h.checkRulePermissions(rules, req.APIGroup, req.Resource, req.Verb) {
			allowed = true
			reason = fmt.Sprintf("Allowed by role: %s", binding.Role.Name)
			break
		}
	}

	response := CheckPermissionsResponse{
		Allowed: allowed,
		Reason:  reason,
	}

	c.JSON(http.StatusOK, response)
}

// Project-specific RBAC endpoints

// ListProjectRoles lists roles for a specific project
func (h *RBACHandler) ListProjectRoles(c *gin.Context) {
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

	var roles []db.Role
	if err := h.db.Where("workspace_id = ? AND project_id = ? AND is_active = true", wsID, projectID).
		Order("created_at DESC").Find(&roles).Error; err != nil {
		h.logger.Error("Failed to list project roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list project roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"total": len(roles),
	})
}

// CreateProjectRole creates a role for a specific project
func (h *RBACHandler) CreateProjectRole(c *gin.Context) {
	projectID := c.Param("projectId")
	
	// Set project_id in the request and call CreateRole
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	
	body["project_id"] = projectID
	
	// Re-bind the modified body
	c.Request.Body = jsonToReadCloser(body)
	
	h.CreateRole(c)
}

// ListProjectRoleBindings lists role bindings for a specific project
func (h *RBACHandler) ListProjectRoleBindings(c *gin.Context) {
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

	var bindings []db.RoleBinding
	if err := h.db.Where("workspace_id = ? AND project_id = ? AND is_active = true", wsID, projectID).
		Order("created_at DESC").Preload("Role").Find(&bindings).Error; err != nil {
		h.logger.Error("Failed to list project role bindings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list project role bindings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role_bindings": bindings,
		"total":         len(bindings),
	})
}

// CreateProjectRoleBinding creates a role binding for a specific project
func (h *RBACHandler) CreateProjectRoleBinding(c *gin.Context) {
	projectID := c.Param("projectId")
	
	// Set project_id in the request and call CreateRoleBinding
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	
	body["project_id"] = projectID
	
	// Re-bind the modified body
	c.Request.Body = jsonToReadCloser(body)
	
	h.CreateRoleBinding(c)
}

// Helper functions

// hasOrgAccess checks if user has access to organization
func (h *RBACHandler) hasOrgAccess(userID, orgID string) bool {
	var count int64
	h.db.Model(&db.OrganizationUser{}).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Count(&count)
	return count > 0
}

// checkRulePermissions checks if the given rules allow the specified action
func (h *RBACHandler) checkRulePermissions(rules []PolicyRule, apiGroup, resource, verb string) bool {
	for _, rule := range rules {
		// Check API groups
		if !h.matchesRule(rule.APIGroups, apiGroup) {
			continue
		}

		// Check resources
		if !h.matchesRule(rule.Resources, resource) {
			continue
		}

		// Check verbs
		if !h.matchesRule(rule.Verbs, verb) {
			continue
		}

		return true
	}
	return false
}

// matchesRule checks if a value matches any item in a rule list (supports wildcards)
func (h *RBACHandler) matchesRule(ruleList []string, value string) bool {
	for _, ruleItem := range ruleList {
		if ruleItem == "*" || ruleItem == value {
			return true
		}
	}
	return false
}

// jsonToReadCloser converts a map to a ReadCloser for re-binding
func jsonToReadCloser(data map[string]interface{}) io.ReadCloser {
	jsonBytes, _ := json.Marshal(data)
	return ioutil.NopCloser(bytes.NewReader(jsonBytes))
}