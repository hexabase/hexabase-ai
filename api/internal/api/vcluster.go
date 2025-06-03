package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// VClusterHandler handles vCluster lifecycle operations
type VClusterHandler struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

// NewVClusterHandler creates a new vCluster handler
func NewVClusterHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *VClusterHandler {
	return &VClusterHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// Request/Response types
type ProvisionVClusterRequest struct {
	Version   string                 `json:"version,omitempty"`
	Resources map[string]interface{} `json:"resources,omitempty"`
	Features  []string               `json:"features,omitempty"`
	Values    map[string]interface{} `json:"values,omitempty"`
}

type VClusterStatusResponse struct {
	Status      string                 `json:"status"`
	Workspace   string                 `json:"workspace"`
	ClusterInfo map[string]interface{} `json:"cluster_info"`
	Health      *VClusterHealth        `json:"health,omitempty"`
}

type VClusterHealth struct {
	Healthy       bool                   `json:"healthy"`
	Components    map[string]string      `json:"components"`
	ResourceUsage map[string]interface{} `json:"resource_usage"`
	LastChecked   time.Time              `json:"last_checked"`
}

type UpgradeVClusterRequest struct {
	TargetVersion string `json:"target_version" binding:"required"`
	Strategy      string `json:"strategy,omitempty"` // rolling, replace
}

type BackupVClusterRequest struct {
	BackupName string `json:"backup_name" binding:"required"`
	Retention  string `json:"retention,omitempty"` // 30d, 90d, etc.
}

type RestoreVClusterRequest struct {
	BackupName string `json:"backup_name" binding:"required"`
	Strategy   string `json:"strategy,omitempty"` // replace, merge
}

// ProvisionVCluster provisions a new vCluster instance
func (h *VClusterHandler) ProvisionVCluster(c *gin.Context) {
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

	var req ProvisionVClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Get workspace
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

	// Check if vCluster is already running or being provisioned
	if workspace.VClusterStatus == "RUNNING" || workspace.VClusterStatus == "CONFIGURING_HNC" {
		c.JSON(http.StatusConflict, gin.H{"error": "vCluster is already provisioned"})
		return
	}

	// Create provisioning task
	payload, _ := json.Marshal(req)
	task := &db.VClusterProvisioningTask{
		WorkspaceID: wsID,
		TaskType:    "CREATE",
		Status:      "PENDING",
		Payload:     string(payload),
	}

	if err := h.db.Create(task).Error; err != nil {
		h.logger.Error("Failed to create provisioning task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create provisioning task"})
		return
	}

	// Update workspace status
	if err := h.db.Model(&workspace).Update("v_cluster_status", "PENDING_CREATION").Error; err != nil {
		h.logger.Error("Failed to update workspace status", zap.Error(err))
	}

	// Start provisioning process asynchronously
	go h.provisionVClusterAsync(task)

	h.logger.Info("VCluster provisioning initiated",
		zap.String("workspace_id", wsID),
		zap.String("task_id", task.ID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": task.ID,
		"status":  "provisioning_initiated",
		"message": "vCluster provisioning has been started",
	})
}

// GetVClusterStatus gets the current status of a vCluster
func (h *VClusterHandler) GetVClusterStatus(c *gin.Context) {
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
	if err := h.db.Where("id = ? AND organization_id = ?", wsID, orgID).First(&workspace).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Parse cluster info
	var clusterInfo map[string]interface{}
	if err := json.Unmarshal([]byte(workspace.VClusterConfig), &clusterInfo); err != nil {
		clusterInfo = make(map[string]interface{})
	}

	response := VClusterStatusResponse{
		Status:      workspace.VClusterStatus,
		Workspace:   workspace.Name,
		ClusterInfo: clusterInfo,
	}

	// Add health info for running clusters
	if workspace.VClusterStatus == "RUNNING" {
		health, err := h.getVClusterHealth(wsID)
		if err != nil {
			h.logger.Error("Failed to get vCluster health", zap.Error(err))
		} else {
			response.Health = health
		}
	}

	c.JSON(http.StatusOK, response)
}

// StartVCluster starts a stopped vCluster
func (h *VClusterHandler) StartVCluster(c *gin.Context) {
	h.performVClusterAction(c, "START", "STOPPED", "STARTING")
}

// StopVCluster stops a running vCluster
func (h *VClusterHandler) StopVCluster(c *gin.Context) {
	h.performVClusterAction(c, "STOP", "RUNNING", "STOPPING")
}

// DestroyVCluster destroys a vCluster instance
func (h *VClusterHandler) DestroyVCluster(c *gin.Context) {
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
	if err := h.db.Where("id = ? AND organization_id = ?", wsID, orgID).First(&workspace).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Check if already being deleted
	if workspace.VClusterStatus == "DELETING" {
		c.JSON(http.StatusConflict, gin.H{"error": "vCluster is already being deleted"})
		return
	}

	// Create deletion task
	task := &db.VClusterProvisioningTask{
		WorkspaceID: wsID,
		TaskType:    "DELETE",
		Status:      "PENDING",
		Payload:     "{}",
	}

	if err := h.db.Create(task).Error; err != nil {
		h.logger.Error("Failed to create deletion task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create deletion task"})
		return
	}

	// Update workspace status
	if err := h.db.Model(&workspace).Update("v_cluster_status", "DELETING").Error; err != nil {
		h.logger.Error("Failed to update workspace status", zap.Error(err))
	}

	// Start deletion process asynchronously
	go h.destroyVClusterAsync(task)

	h.logger.Info("VCluster destruction initiated",
		zap.String("workspace_id", wsID),
		zap.String("task_id", task.ID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": task.ID,
		"status":  "deletion_initiated",
		"message": "vCluster destruction has been started",
	})
}

// GetVClusterHealth returns health information for a vCluster
func (h *VClusterHandler) GetVClusterHealth(c *gin.Context) {
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
	if err := h.db.Where("id = ? AND organization_id = ?", wsID, orgID).First(&workspace).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Check if vCluster is running
	if workspace.VClusterStatus != "RUNNING" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":  "vCluster is not running",
			"status": workspace.VClusterStatus,
		})
		return
	}

	health, err := h.getVClusterHealth(wsID)
	if err != nil {
		h.logger.Error("Failed to get vCluster health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get health information"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// UpgradeVCluster upgrades a vCluster to a new version
func (h *VClusterHandler) UpgradeVCluster(c *gin.Context) {
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

	var req UpgradeVClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Get workspace
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

	// Check if vCluster is running
	if workspace.VClusterStatus != "RUNNING" {
		c.JSON(http.StatusConflict, gin.H{"error": "vCluster must be running to upgrade"})
		return
	}

	// Create upgrade task
	payload, _ := json.Marshal(req)
	task := &db.VClusterProvisioningTask{
		WorkspaceID: wsID,
		TaskType:    "UPGRADE",
		Status:      "PENDING",
		Payload:     string(payload),
	}

	if err := h.db.Create(task).Error; err != nil {
		h.logger.Error("Failed to create upgrade task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upgrade task"})
		return
	}

	// Start upgrade process asynchronously
	go h.upgradeVClusterAsync(task)

	h.logger.Info("VCluster upgrade initiated",
		zap.String("workspace_id", wsID),
		zap.String("target_version", req.TargetVersion),
		zap.String("task_id", task.ID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": task.ID,
		"status":  "upgrade_initiated",
		"message": fmt.Sprintf("vCluster upgrade to %s has been started", req.TargetVersion),
	})
}

// GetVClusterLogs retrieves logs from vCluster components
func (h *VClusterHandler) GetVClusterLogs(c *gin.Context) {
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

	// Get query parameters
	lines := c.DefaultQuery("lines", "100")
	component := c.DefaultQuery("component", "vcluster")

	logs, err := h.getVClusterLogs(wsID, component, lines)
	if err != nil {
		h.logger.Error("Failed to get vCluster logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"component": component,
		"lines":     lines,
	})
}

// BackupVCluster creates a backup of the vCluster
func (h *VClusterHandler) BackupVCluster(c *gin.Context) {
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

	var req BackupVClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Get workspace
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

	// Check if vCluster is running
	if workspace.VClusterStatus != "RUNNING" {
		c.JSON(http.StatusConflict, gin.H{"error": "vCluster must be running to backup"})
		return
	}

	// Create backup task
	payload, _ := json.Marshal(req)
	task := &db.VClusterProvisioningTask{
		WorkspaceID: wsID,
		TaskType:    "BACKUP",
		Status:      "PENDING",
		Payload:     string(payload),
	}

	if err := h.db.Create(task).Error; err != nil {
		h.logger.Error("Failed to create backup task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create backup task"})
		return
	}

	// Start backup process asynchronously
	go h.backupVClusterAsync(task)

	h.logger.Info("VCluster backup initiated",
		zap.String("workspace_id", wsID),
		zap.String("backup_name", req.BackupName),
		zap.String("task_id", task.ID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": task.ID,
		"status":  "backup_initiated",
		"message": fmt.Sprintf("vCluster backup '%s' has been started", req.BackupName),
	})
}

// RestoreVCluster restores a vCluster from backup
func (h *VClusterHandler) RestoreVCluster(c *gin.Context) {
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

	var req RestoreVClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Create restore task
	payload, _ := json.Marshal(req)
	task := &db.VClusterProvisioningTask{
		WorkspaceID: wsID,
		TaskType:    "RESTORE",
		Status:      "PENDING",
		Payload:     string(payload),
	}

	if err := h.db.Create(task).Error; err != nil {
		h.logger.Error("Failed to create restore task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create restore task"})
		return
	}

	// Start restore process asynchronously
	go h.restoreVClusterAsync(task)

	h.logger.Info("VCluster restore initiated",
		zap.String("workspace_id", wsID),
		zap.String("backup_name", req.BackupName),
		zap.String("task_id", task.ID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": task.ID,
		"status":  "restore_initiated",
		"message": fmt.Sprintf("vCluster restore from '%s' has been started", req.BackupName),
	})
}

// Task Management

// ListTasks lists vCluster provisioning tasks
func (h *VClusterHandler) ListTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Get query parameters
	status := c.Query("status")
	taskType := c.Query("task_type")
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	// Build query
	query := h.db.Model(&db.VClusterProvisioningTask{}).
		Joins("JOIN workspaces ON workspaces.id = v_cluster_provisioning_tasks.workspace_id").
		Joins("JOIN organizations ON organizations.id = workspaces.organization_id").
		Joins("JOIN organization_users ON organization_users.organization_id = organizations.id").
		Where("organization_users.user_id = ?", userID)

	if status != "" {
		query = query.Where("v_cluster_provisioning_tasks.status = ?", status)
	}
	if taskType != "" {
		query = query.Where("v_cluster_provisioning_tasks.task_type = ?", taskType)
	}

	var tasks []db.VClusterProvisioningTask
	if err := query.Order("v_cluster_provisioning_tasks.created_at DESC").
		Limit(limitInt).Offset(offsetInt).
		Preload("Workspace").Find(&tasks).Error; err != nil {
		h.logger.Error("Failed to list tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// GetTask gets a specific task
func (h *VClusterHandler) GetTask(c *gin.Context) {
	taskID := c.Param("taskId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var task db.VClusterProvisioningTask
	err := h.db.Joins("JOIN workspaces ON workspaces.id = v_cluster_provisioning_tasks.workspace_id").
		Joins("JOIN organizations ON organizations.id = workspaces.organization_id").
		Joins("JOIN organization_users ON organization_users.organization_id = organizations.id").
		Where("v_cluster_provisioning_tasks.id = ? AND organization_users.user_id = ?", taskID, userID).
		Preload("Workspace").First(&task).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		} else {
			h.logger.Error("Failed to get task", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

// RetryTask retries a failed task
func (h *VClusterHandler) RetryTask(c *gin.Context) {
	taskID := c.Param("taskId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var task db.VClusterProvisioningTask
	err := h.db.Joins("JOIN workspaces ON workspaces.id = v_cluster_provisioning_tasks.workspace_id").
		Joins("JOIN organizations ON organizations.id = workspaces.organization_id").
		Joins("JOIN organization_users ON organization_users.organization_id = organizations.id").
		Where("v_cluster_provisioning_tasks.id = ? AND organization_users.user_id = ?", taskID, userID).
		First(&task).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		} else {
			h.logger.Error("Failed to get task", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
		}
		return
	}

	if task.Status != "FAILED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only failed tasks can be retried"})
		return
	}

	// Reset task status
	task.Status = "PENDING"
	task.ErrorMessage = nil
	if err := h.db.Save(&task).Error; err != nil {
		h.logger.Error("Failed to update task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retry task"})
		return
	}

	// Restart task processing
	switch task.TaskType {
	case "CREATE":
		go h.provisionVClusterAsync(&task)
	case "DELETE":
		go h.destroyVClusterAsync(&task)
	case "UPGRADE":
		go h.upgradeVClusterAsync(&task)
	case "BACKUP":
		go h.backupVClusterAsync(&task)
	case "RESTORE":
		go h.restoreVClusterAsync(&task)
	case "START":
		go h.startVClusterAsync(&task)
	case "STOP":
		go h.stopVClusterAsync(&task)
	}

	h.logger.Info("Task retried",
		zap.String("task_id", taskID),
		zap.String("task_type", task.TaskType),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{
		"message": "task has been retried",
		"task":    task,
	})
}

// Helper methods

// hasOrgAccess checks if user has access to organization
func (h *VClusterHandler) hasOrgAccess(userID, orgID string) bool {
	var count int64
	h.db.Model(&db.OrganizationUser{}).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Count(&count)
	return count > 0
}

// performVClusterAction performs a generic vCluster action (start/stop)
func (h *VClusterHandler) performVClusterAction(c *gin.Context, taskType, requiredStatus, intermediateStatus string) {
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
	if err := h.db.Where("id = ? AND organization_id = ?", wsID, orgID).First(&workspace).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Check current status
	if requiredStatus != "" && workspace.VClusterStatus != requiredStatus {
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("vCluster must be in %s state", requiredStatus),
		})
		return
	}

	// Create task
	task := &db.VClusterProvisioningTask{
		WorkspaceID: wsID,
		TaskType:    taskType,
		Status:      "PENDING",
		Payload:     "{}",
	}

	if err := h.db.Create(task).Error; err != nil {
		h.logger.Error("Failed to create task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	// Update workspace status
	if intermediateStatus != "" {
		if err := h.db.Model(&workspace).Update("v_cluster_status", intermediateStatus).Error; err != nil {
			h.logger.Error("Failed to update workspace status", zap.Error(err))
		}
	}

	// Start task processing
	if taskType == "START" {
		go h.startVClusterAsync(task)
	} else if taskType == "STOP" {
		go h.stopVClusterAsync(task)
	}

	h.logger.Info("VCluster action initiated",
		zap.String("action", taskType),
		zap.String("workspace_id", wsID),
		zap.String("task_id", task.ID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": task.ID,
		"status":  fmt.Sprintf("%s_initiated", strings.ToLower(taskType)),
		"message": fmt.Sprintf("vCluster %s has been started", strings.ToLower(taskType)),
	})
}

// Async processing methods (simplified for testing)

func (h *VClusterHandler) provisionVClusterAsync(task *db.VClusterProvisioningTask) {
	h.processTaskAsync(task, func() error {
		// Update task status
		h.db.Model(task).Update("status", "RUNNING")

		// Simulate vCluster provisioning
		h.logger.Info("Starting vCluster provisioning", zap.String("workspace_id", task.WorkspaceID))

		// In a real implementation, this would:
		// 1. Create Kubernetes namespace
		// 2. Install vCluster helm chart
		// 3. Wait for vCluster to be ready
		// 4. Configure OIDC
		// 5. Set up resource quotas
		// 6. Configure HNC (Hierarchical Namespace Controller)

		// For testing, simulate work with sleep
		time.Sleep(2 * time.Second)

		// Update workspace status
		updates := map[string]interface{}{
			"v_cluster_status":        "RUNNING",
			"v_cluster_instance_name": fmt.Sprintf("vcluster-%s", task.WorkspaceID),
			"v_cluster_config": `{
				"version": "0.15.0",
				"endpoint": "https://vcluster-` + task.WorkspaceID + `.hexabase-workspaces.svc.cluster.local",
				"status": "ready"
			}`,
		}

		if err := h.db.Model(&db.Workspace{}).Where("id = ?", task.WorkspaceID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update workspace: %w", err)
		}

		h.logger.Info("VCluster provisioning completed", zap.String("workspace_id", task.WorkspaceID))
		return nil
	})
}

func (h *VClusterHandler) destroyVClusterAsync(task *db.VClusterProvisioningTask) {
	h.processTaskAsync(task, func() error {
		h.db.Model(task).Update("status", "RUNNING")

		h.logger.Info("Starting vCluster destruction", zap.String("workspace_id", task.WorkspaceID))

		// Simulate destruction work
		time.Sleep(1 * time.Second)

		// Update workspace status
		updates := map[string]interface{}{
			"v_cluster_status":        "PENDING_CREATION",
			"v_cluster_instance_name": nil,
			"v_cluster_config":        "{}",
		}

		if err := h.db.Model(&db.Workspace{}).Where("id = ?", task.WorkspaceID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update workspace: %w", err)
		}

		h.logger.Info("VCluster destruction completed", zap.String("workspace_id", task.WorkspaceID))
		return nil
	})
}

func (h *VClusterHandler) startVClusterAsync(task *db.VClusterProvisioningTask) {
	h.processTaskAsync(task, func() error {
		h.db.Model(task).Update("status", "RUNNING")
		time.Sleep(1 * time.Second)
		h.db.Model(&db.Workspace{}).Where("id = ?", task.WorkspaceID).Update("v_cluster_status", "RUNNING")
		return nil
	})
}

func (h *VClusterHandler) stopVClusterAsync(task *db.VClusterProvisioningTask) {
	h.processTaskAsync(task, func() error {
		h.db.Model(task).Update("status", "RUNNING")
		time.Sleep(1 * time.Second)
		h.db.Model(&db.Workspace{}).Where("id = ?", task.WorkspaceID).Update("v_cluster_status", "STOPPED")
		return nil
	})
}

func (h *VClusterHandler) upgradeVClusterAsync(task *db.VClusterProvisioningTask) {
	h.processTaskAsync(task, func() error {
		h.db.Model(task).Update("status", "RUNNING")
		time.Sleep(2 * time.Second)
		// Parse upgrade request and update cluster config
		var req UpgradeVClusterRequest
		json.Unmarshal([]byte(task.Payload), &req)

		config := fmt.Sprintf(`{
			"version": "%s",
			"endpoint": "https://vcluster-%s.hexabase-workspaces.svc.cluster.local",
			"status": "ready"
		}`, req.TargetVersion, task.WorkspaceID)

		h.db.Model(&db.Workspace{}).Where("id = ?", task.WorkspaceID).Update("v_cluster_config", config)
		return nil
	})
}

func (h *VClusterHandler) backupVClusterAsync(task *db.VClusterProvisioningTask) {
	h.processTaskAsync(task, func() error {
		h.db.Model(task).Update("status", "RUNNING")
		time.Sleep(2 * time.Second)
		// Simulate backup creation
		h.logger.Info("VCluster backup completed", zap.String("workspace_id", task.WorkspaceID))
		return nil
	})
}

func (h *VClusterHandler) restoreVClusterAsync(task *db.VClusterProvisioningTask) {
	h.processTaskAsync(task, func() error {
		h.db.Model(task).Update("status", "RUNNING")
		time.Sleep(3 * time.Second)
		// Simulate restore process
		h.logger.Info("VCluster restore completed", zap.String("workspace_id", task.WorkspaceID))
		return nil
	})
}

func (h *VClusterHandler) processTaskAsync(task *db.VClusterProvisioningTask, processFunc func() error) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Task processing panicked", 
				zap.String("task_id", task.ID),
				zap.Any("panic", r))
			
			errorMsg := fmt.Sprintf("Task processing panicked: %v", r)
			h.db.Model(task).Updates(map[string]interface{}{
				"status":        "FAILED",
				"error_message": errorMsg,
			})
		}
	}()

	if err := processFunc(); err != nil {
		h.logger.Error("Task processing failed",
			zap.String("task_id", task.ID),
			zap.Error(err))

		h.db.Model(task).Updates(map[string]interface{}{
			"status":        "FAILED",
			"error_message": err.Error(),
		})
		return
	}

	// Mark task as completed
	h.db.Model(task).Update("status", "COMPLETED")
}

// getVClusterHealth retrieves health information for a vCluster
func (h *VClusterHandler) getVClusterHealth(workspaceID string) (*VClusterHealth, error) {
	// In a real implementation, this would check:
	// 1. vCluster API server health
	// 2. Component status (etcd, controller-manager, scheduler)
	// 3. Resource usage (CPU, memory, storage)
	// 4. Node status

	// For testing, return mock health data
	return &VClusterHealth{
		Healthy: true,
		Components: map[string]string{
			"api-server":          "healthy",
			"etcd":                "healthy",
			"controller-manager":  "healthy",
			"scheduler":           "healthy",
		},
		ResourceUsage: map[string]interface{}{
			"cpu_usage":    "45%",
			"memory_usage": "2.1Gi/4Gi",
			"storage_usage": "3.2Gi/10Gi",
		},
		LastChecked: time.Now(),
	}, nil
}

// getVClusterLogs retrieves logs from vCluster components
func (h *VClusterHandler) getVClusterLogs(workspaceID, component, lines string) ([]string, error) {
	// In a real implementation, this would use kubectl logs or Kubernetes API
	// to retrieve actual logs from vCluster components

	// For testing, return mock logs
	return []string{
		fmt.Sprintf("[%s] VCluster %s started successfully", time.Now().Format(time.RFC3339), component),
		fmt.Sprintf("[%s] All components are healthy", time.Now().Add(-1*time.Minute).Format(time.RFC3339)),
		fmt.Sprintf("[%s] Resource usage within limits", time.Now().Add(-2*time.Minute).Format(time.RFC3339)),
	}, nil
}