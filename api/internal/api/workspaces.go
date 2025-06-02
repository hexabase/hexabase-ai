package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateWorkspaceRequest represents the request payload for creating a workspace
type CreateWorkspaceRequest struct {
	Name   string `json:"name" binding:"required"`
	PlanID string `json:"plan_id" binding:"required"`
}

// UpdateWorkspaceRequest represents the request payload for updating a workspace
type UpdateWorkspaceRequest struct {
	Name   *string `json:"name,omitempty"`
	PlanID *string `json:"plan_id,omitempty"`
}

// WorkspaceHandler handles workspace-related endpoints
type WorkspaceHandler struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// CreateWorkspace creates a new workspace (vCluster)
func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	orgID := c.Param("orgId")
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

	var req CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Check specific validation errors
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace name is required"})
		} else if req.PlanID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "plan_id is required"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		}
		return
	}

	// Validate workspace name
	if len(strings.TrimSpace(req.Name)) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace name must be at least 3 characters"})
		return
	}

	// Verify plan exists and is active
	var plan db.Plan
	if err := h.db.First(&plan, "id = ? AND is_active = ?", req.PlanID, true).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan selected"})
		} else {
			h.logger.Error("Failed to check plan", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	// Check if workspace name already exists in organization
	var existingWs db.Workspace
	if err := h.db.First(&existingWs, "organization_id = ? AND name = ?", orgID, req.Name).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "workspace with this name already exists"})
		return
	}

	// Create workspace
	workspace := &db.Workspace{
		OrganizationID: orgID,
		Name:           strings.TrimSpace(req.Name),
		PlanID:         req.PlanID,
		VClusterStatus: "PENDING_CREATION",
		VClusterConfig: "{}",
		DedicatedNodeConfig: "{}",
	}

	if err := h.db.Create(workspace).Error; err != nil {
		h.logger.Error("Failed to create workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace"})
		return
	}

	// Create vCluster provisioning task
	task := &db.VClusterProvisioningTask{
		WorkspaceID: workspace.ID,
		TaskType:    "CREATE",
		Status:      "PENDING",
		Payload:     "{}",
	}

	if err := h.db.Create(task).Error; err != nil {
		h.logger.Error("Failed to create provisioning task", zap.Error(err))
		// Don't fail the request, but log the error
	}

	h.logger.Info("Workspace created successfully", 
		zap.String("workspace_id", workspace.ID),
		zap.String("org_id", orgID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusCreated, workspace)
}

// hasOrgAccess checks if user has access to the organization
func (h *WorkspaceHandler) hasOrgAccess(userID, orgID string) bool {
	var membership db.OrganizationUser
	err := h.db.First(&membership, "user_id = ? AND organization_id = ?", userID, orgID).Error
	return err == nil
}

// ListWorkspaces lists workspaces in organization
func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	orgID := c.Param("orgId")
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

	var workspaces []db.Workspace
	if err := h.db.Where("organization_id = ?", orgID).Find(&workspaces).Error; err != nil {
		h.logger.Error("Failed to list workspaces", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workspaces"})
		return
	}

	h.logger.Info("List workspaces successful", 
		zap.String("org_id", orgID),
		zap.String("user_id", userID.(string)),
		zap.Int("count", len(workspaces)))

	c.JSON(http.StatusOK, gin.H{
		"workspaces": workspaces,
		"total":      len(workspaces),
	})
}

// GetWorkspace gets workspace details
func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
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

	h.logger.Info("Get workspace successful", 
		zap.String("org_id", orgID),
		zap.String("ws_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, workspace)
}

// UpdateWorkspace updates workspace configuration
func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
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

	var req UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Get existing workspace
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

	// Track what needs to be updated
	updates := make(map[string]interface{})

	// Update name if provided
	if req.Name != nil {
		if len(strings.TrimSpace(*req.Name)) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace name must be at least 3 characters"})
			return
		}
		updates["name"] = strings.TrimSpace(*req.Name)
	}

	// Update plan if provided
	if req.PlanID != nil {
		// Get current plan
		var currentPlan db.Plan
		if err := h.db.First(&currentPlan, "id = ?", workspace.PlanID).Error; err != nil {
			h.logger.Error("Failed to get current plan", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current plan"})
			return
		}

		// Get new plan
		var newPlan db.Plan
		if err := h.db.First(&newPlan, "id = ? AND is_active = ?", *req.PlanID, true).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan selected"})
			} else {
				h.logger.Error("Failed to check new plan", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}

		// Check if it's a downgrade (simplified check based on price)
		if newPlan.Price < currentPlan.Price {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot downgrade plan"})
			return
		}

		updates["plan_id"] = *req.PlanID
	}

	// Apply updates if any
	if len(updates) > 0 {
		if err := h.db.Model(&workspace).Updates(updates).Error; err != nil {
			h.logger.Error("Failed to update workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workspace"})
			return
		}

		// Reload workspace to get updated data
		if err := h.db.First(&workspace, "id = ?", wsID).Error; err != nil {
			h.logger.Error("Failed to reload workspace", zap.Error(err))
		}
	}

	h.logger.Info("Update workspace successful", 
		zap.String("org_id", orgID),
		zap.String("ws_id", wsID),
		zap.String("user_id", userID.(string)),
		zap.Any("updates", updates))

	c.JSON(http.StatusOK, workspace)
}

// DeleteWorkspace deletes a workspace
func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
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

	// Get existing workspace
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

	// Check if workspace has any projects
	var projectCount int64
	if err := h.db.Model(&db.Project{}).Where("workspace_id = ?", wsID).Count(&projectCount).Error; err != nil {
		h.logger.Error("Failed to count projects", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check workspace projects"})
		return
	}

	if projectCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete workspace with existing projects"})
		return
	}

	// Mark workspace for deletion
	if err := h.db.Model(&workspace).Update("v_cluster_status", "DELETING").Error; err != nil {
		h.logger.Error("Failed to mark workspace for deletion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workspace"})
		return
	}

	// Create deletion task
	task := &db.VClusterProvisioningTask{
		WorkspaceID: workspace.ID,
		TaskType:    "DELETE",
		Status:      "PENDING",
		Payload:     "{}",
	}

	if err := h.db.Create(task).Error; err != nil {
		h.logger.Error("Failed to create deletion task", zap.Error(err))
		// Don't fail the request, but log the error
	}

	h.logger.Info("Delete workspace successful", 
		zap.String("org_id", orgID),
		zap.String("ws_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{
		"message": "workspace deletion initiated",
		"status":  "DELETING",
	})
}

// GetKubeconfig generates kubeconfig for workspace
func (h *WorkspaceHandler) GetKubeconfig(c *gin.Context) {
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

	// Get workspace
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

	// Check if workspace is ready
	if workspace.VClusterStatus != "RUNNING" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace is not ready"})
		return
	}

	// Generate kubeconfig for vCluster
	kubeconfig := h.generateKubeconfig(&workspace, userID.(string))

	h.logger.Info("Get kubeconfig successful", 
		zap.String("org_id", orgID),
		zap.String("ws_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{
		"kubeconfig": kubeconfig,
		"workspace":  workspace.Name,
		"status":     workspace.VClusterStatus,
	})
}

// generateKubeconfig generates a kubeconfig for the given workspace and user
func (h *WorkspaceHandler) generateKubeconfig(workspace *db.Workspace, userID string) string {
	// In a real implementation, this would:
	// 1. Connect to the vCluster API
	// 2. Generate user certificates or service account tokens
	// 3. Create a proper kubeconfig with cluster info
	
	// For testing purposes, return a mock kubeconfig
	clusterName := workspace.Name
	if workspace.VClusterInstanceName != nil {
		clusterName = *workspace.VClusterInstanceName
	}

	return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://vcluster-%s.hexabase-workspaces.svc.cluster.local
    insecure-skip-tls-verify: true
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
users:
- name: %s
  user:
    token: vcluster-token-for-user-%s
`, clusterName, clusterName, clusterName, userID, clusterName, clusterName, userID, userID)
}

// CreateClusterRoleAssignment creates cluster role assignment
func (h *WorkspaceHandler) CreateClusterRoleAssignment(c *gin.Context) {
	wsID := c.Param("wsId")
	h.logger.Info("Create cluster role assignment endpoint called", zap.String("ws_id", wsID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// ListClusterRoleAssignments lists cluster role assignments
func (h *WorkspaceHandler) ListClusterRoleAssignments(c *gin.Context) {
	wsID := c.Param("wsId")
	h.logger.Info("List cluster role assignments endpoint called", zap.String("ws_id", wsID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// DeleteClusterRoleAssignment deletes cluster role assignment
func (h *WorkspaceHandler) DeleteClusterRoleAssignment(c *gin.Context) {
	wsID := c.Param("wsId")
	assignmentID := c.Param("assignmentId")
	h.logger.Info("Delete cluster role assignment endpoint called", 
		zap.String("ws_id", wsID),
		zap.String("assignment_id", assignmentID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}