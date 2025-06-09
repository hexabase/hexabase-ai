package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
)

// BackupHandler handles backup-related HTTP requests
type BackupHandler struct {
	backupService backup.Service
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(backupService backup.Service) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
	}
}

// CreateBackupStorage creates a new backup storage
// @Summary Create backup storage
// @Description Create a new backup storage for a workspace (Dedicated Plan only)
// @Tags backup
// @Accept json
// @Produce json
// @Param wsId path string true "Workspace ID"
// @Param request body backup.CreateBackupStorageRequest true "Backup storage configuration"
// @Success 201 {object} backup.BackupStorage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse "Only available for Dedicated Plan"
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{wsId}/backup/storages [post]
func (h *BackupHandler) CreateBackupStorage(c *gin.Context) {
	workspaceID := c.Param("wsId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace ID is required"})
		return
	}

	var req backup.CreateBackupStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	storage, err := h.backupService.CreateBackupStorage(c.Request.Context(), workspaceID, req)
	if err != nil {
		if err.Error() == "backup storage is only available for Dedicated Plan workspaces" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, storage)
}

// ListBackupStorages lists all backup storages for a workspace
// @Summary List backup storages
// @Description Get all backup storages for a workspace
// @Tags backup
// @Accept json
// @Produce json
// @Param wsId path string true "Workspace ID"
// @Success 200 {array} backup.BackupStorage
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{wsId}/backup/storages [get]
func (h *BackupHandler) ListBackupStorages(c *gin.Context) {
	workspaceID := c.Param("wsId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace ID is required"})
		return
	}

	storages, err := h.backupService.ListBackupStorages(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, storages)
}

// GetBackupStorage retrieves a specific backup storage
// @Summary Get backup storage
// @Description Get details of a specific backup storage
// @Tags backup
// @Accept json
// @Produce json
// @Param wsId path string true "Workspace ID"
// @Param storageId path string true "Storage ID"
// @Success 200 {object} backup.BackupStorage
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{wsId}/backup/storages/{storageId} [get]
func (h *BackupHandler) GetBackupStorage(c *gin.Context) {
	storageID := c.Param("storageId")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storage ID is required"})
		return
	}

	storage, err := h.backupService.GetBackupStorage(c.Request.Context(), storageID)
	if err != nil {
		if err.Error() == "backup storage not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, storage)
}

// UpdateBackupStorage updates a backup storage
// @Summary Update backup storage
// @Description Update an existing backup storage
// @Tags backup
// @Accept json
// @Produce json
// @Param wsId path string true "Workspace ID"
// @Param storageId path string true "Storage ID"
// @Param request body backup.UpdateBackupStorageRequest true "Update request"
// @Success 200 {object} backup.BackupStorage
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{wsId}/backup/storages/{storageId} [put]
func (h *BackupHandler) UpdateBackupStorage(c *gin.Context) {
	storageID := c.Param("storageId")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storage ID is required"})
		return
	}

	var req backup.UpdateBackupStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	storage, err := h.backupService.UpdateBackupStorage(c.Request.Context(), storageID, req)
	if err != nil {
		if err.Error() == "backup storage not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, storage)
}

// DeleteBackupStorage deletes a backup storage
// @Summary Delete backup storage
// @Description Delete a backup storage
// @Tags backup
// @Accept json
// @Produce json
// @Param wsId path string true "Workspace ID"
// @Param storageId path string true "Storage ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{wsId}/backup/storages/{storageId} [delete]
func (h *BackupHandler) DeleteBackupStorage(c *gin.Context) {
	storageID := c.Param("storageId")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storage ID is required"})
		return
	}

	err := h.backupService.DeleteBackupStorage(c.Request.Context(), storageID)
	if err != nil {
		if err.Error() == "backup storage not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetStorageUsage retrieves storage usage information
// @Summary Get storage usage
// @Description Get usage statistics for a backup storage
// @Tags backup
// @Accept json
// @Produce json
// @Param wsId path string true "Workspace ID"
// @Param storageId path string true "Storage ID"
// @Success 200 {object} backup.BackupStorageUsage
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{wsId}/backup/storages/{storageId}/usage [get]
func (h *BackupHandler) GetStorageUsage(c *gin.Context) {
	storageID := c.Param("storageId")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storage ID is required"})
		return
	}

	usage, err := h.backupService.GetStorageUsage(c.Request.Context(), storageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// CreateBackupPolicy creates a backup policy for an application
// @Summary Create backup policy
// @Description Create a backup policy for an application (Dedicated Plan only)
// @Tags backup
// @Accept json
// @Produce json
// @Param appId path string true "Application ID"
// @Param request body backup.CreateBackupPolicyRequest true "Backup policy configuration"
// @Success 201 {object} backup.BackupPolicy
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse "Only available for Dedicated Plan"
// @Failure 500 {object} ErrorResponse
// @Router /applications/{appId}/backup/policy [post]
func (h *BackupHandler) CreateBackupPolicy(c *gin.Context) {
	applicationID := c.Param("appId")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var req backup.CreateBackupPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	policy, err := h.backupService.CreateBackupPolicy(c.Request.Context(), applicationID, req)
	if err != nil {
		if err.Error() == "backup policies are only available for Dedicated Plan workspaces" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, policy)
}

// GetBackupPolicy retrieves the backup policy for an application
// @Summary Get backup policy
// @Description Get the backup policy for an application
// @Tags backup
// @Accept json
// @Produce json
// @Param appId path string true "Application ID"
// @Success 200 {object} backup.BackupPolicy
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /applications/{appId}/backup/policy [get]
func (h *BackupHandler) GetBackupPolicy(c *gin.Context) {
	applicationID := c.Param("appId")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	policy, err := h.backupService.GetBackupPolicyByApplication(c.Request.Context(), applicationID)
	if err != nil {
		if err.Error() == "no backup policy found for application" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// UpdateBackupPolicy updates a backup policy
// @Summary Update backup policy
// @Description Update an existing backup policy
// @Tags backup
// @Accept json
// @Produce json
// @Param appId path string true "Application ID"
// @Param policyId path string true "Policy ID"
// @Param request body backup.UpdateBackupPolicyRequest true "Update request"
// @Success 200 {object} backup.BackupPolicy
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /applications/{appId}/backup/policy/{policyId} [put]
func (h *BackupHandler) UpdateBackupPolicy(c *gin.Context) {
	policyID := c.Param("policyId")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy ID is required"})
		return
	}

	var req backup.UpdateBackupPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	policy, err := h.backupService.UpdateBackupPolicy(c.Request.Context(), policyID, req)
	if err != nil {
		if err.Error() == "backup policy not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// DeleteBackupPolicy deletes a backup policy
// @Summary Delete backup policy
// @Description Delete a backup policy
// @Tags backup
// @Accept json
// @Produce json
// @Param appId path string true "Application ID"
// @Param policyId path string true "Policy ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /applications/{appId}/backup/policy/{policyId} [delete]
func (h *BackupHandler) DeleteBackupPolicy(c *gin.Context) {
	policyID := c.Param("policyId")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy ID is required"})
		return
	}

	err := h.backupService.DeleteBackupPolicy(c.Request.Context(), policyID)
	if err != nil {
		if err.Error() == "backup policy not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// TriggerManualBackup manually triggers a backup
// @Summary Trigger manual backup
// @Description Manually trigger a backup for an application
// @Tags backup
// @Accept json
// @Produce json
// @Param appId path string true "Application ID"
// @Param request body backup.TriggerBackupRequest true "Backup trigger request"
// @Success 202 {object} backup.BackupExecution
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse "No backup policy found"
// @Failure 500 {object} ErrorResponse
// @Router /applications/{appId}/backup/trigger [post]
func (h *BackupHandler) TriggerManualBackup(c *gin.Context) {
	applicationID := c.Param("appId")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var req backup.TriggerBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	req.ApplicationID = applicationID
	execution, err := h.backupService.TriggerManualBackup(c.Request.Context(), applicationID, req)
	if err != nil {
		if err.Error() == "no backup policy found for application" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, execution)
}

// ListBackupExecutions lists backup executions for an application
// @Summary List backup executions
// @Description Get all backup executions for an application
// @Tags backup
// @Accept json
// @Produce json
// @Param appId path string true "Application ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{executions=[]backup.BackupExecution,total=int}
// @Failure 500 {object} ErrorResponse
// @Router /applications/{appId}/backup/executions [get]
func (h *BackupHandler) ListBackupExecutions(c *gin.Context) {
	applicationID := c.Param("appId")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	executions, total, err := h.backupService.ListBackupExecutions(c.Request.Context(), applicationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"total":      total,
	})
}

// GetBackupExecution retrieves a specific backup execution
// @Summary Get backup execution
// @Description Get details of a specific backup execution
// @Tags backup
// @Accept json
// @Produce json
// @Param appId path string true "Application ID"
// @Param executionId path string true "Execution ID"
// @Success 200 {object} backup.BackupExecution
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /applications/{appId}/backup/executions/{executionId} [get]
func (h *BackupHandler) GetBackupExecution(c *gin.Context) {
	executionID := c.Param("executionId")
	if executionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "execution ID is required"})
		return
	}

	execution, err := h.backupService.GetBackupExecution(c.Request.Context(), executionID)
	if err != nil {
		if err.Error() == "backup execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// RestoreBackup initiates a restore from a backup
// @Summary Restore backup
// @Description Initiate a restore operation from a backup
// @Tags backup
// @Accept json
// @Produce json
// @Param executionId path string true "Backup Execution ID"
// @Param request body backup.RestoreBackupRequest true "Restore request"
// @Success 202 {object} backup.BackupRestore
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /backup/executions/{executionId}/restore [post]
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	executionID := c.Param("executionId")
	if executionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "execution ID is required"})
		return
	}

	var req backup.RestoreBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	restore, err := h.backupService.RestoreBackup(c.Request.Context(), executionID, req)
	if err != nil {
		if err.Error() == "backup execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, restore)
}

// GetBackupRestore retrieves a restore operation status
// @Summary Get restore status
// @Description Get the status of a restore operation
// @Tags backup
// @Accept json
// @Produce json
// @Param restoreId path string true "Restore ID"
// @Success 200 {object} backup.BackupRestore
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /backup/restores/{restoreId} [get]
func (h *BackupHandler) GetBackupRestore(c *gin.Context) {
	restoreID := c.Param("restoreId")
	if restoreID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "restore ID is required"})
		return
	}

	restore, err := h.backupService.GetBackupRestore(c.Request.Context(), restoreID)
	if err != nil {
		if err.Error() == "backup restore not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, restore)
}

// ListBackupRestores lists restore operations for an application
// @Summary List restore operations
// @Description Get all restore operations for an application
// @Tags backup
// @Accept json
// @Produce json
// @Param appId path string true "Application ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{restores=[]backup.BackupRestore,total=int}
// @Failure 500 {object} ErrorResponse
// @Router /applications/{appId}/backup/restores [get]
func (h *BackupHandler) ListBackupRestores(c *gin.Context) {
	applicationID := c.Param("appId")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	restores, total, err := h.backupService.ListBackupRestores(c.Request.Context(), applicationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"restores": restores,
		"total":    total,
	})
}

// GetLatestBackup retrieves the latest successful backup for an application
// @Summary Get latest backup
// @Description Get the most recent successful backup for an application
// @Tags backup
// @Accept json
// @Produce json
// @Param appId path string true "Application ID"
// @Success 200 {object} backup.BackupExecution
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /applications/{appId}/backup/latest [get]
func (h *BackupHandler) GetLatestBackup(c *gin.Context) {
	applicationID := c.Param("appId")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	execution, err := h.backupService.GetLatestBackup(c.Request.Context(), applicationID)
	if err != nil {
		if err.Error() == "no backup policy found for application" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// GetBackupManifest retrieves the manifest of backed up resources
// @Summary Get backup manifest
// @Description Get the manifest of resources included in a backup
// @Tags backup
// @Accept json
// @Produce json
// @Param backupId path string true "Backup Execution ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /backup/executions/{backupId}/manifest [get]
func (h *BackupHandler) GetBackupManifest(c *gin.Context) {
	backupID := c.Param("backupId")
	if backupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backup ID is required"})
		return
	}

	manifest, err := h.backupService.GetBackupManifest(c.Request.Context(), backupID)
	if err != nil {
		if err.Error() == "backup execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, manifest)
}

// GetWorkspaceStorageUsage retrieves storage usage for all storages in a workspace
// @Summary Get workspace storage usage
// @Description Get storage usage statistics for all backup storages in a workspace
// @Tags backup
// @Accept json
// @Produce json
// @Param wsId path string true "Workspace ID"
// @Success 200 {array} backup.BackupStorageUsage
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{wsId}/backup/storage-usage [get]
func (h *BackupHandler) GetWorkspaceStorageUsage(c *gin.Context) {
	workspaceID := c.Param("wsId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace ID is required"})
		return
	}

	usageList, err := h.backupService.GetWorkspaceStorageUsage(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usageList)
}

// DownloadBackup generates a download link for a backup
// @Summary Download backup
// @Description Get a pre-signed URL to download a backup
// @Tags backup
// @Accept json
// @Produce json
// @Param executionId path string true "Backup Execution ID"
// @Success 200 {object} object{downloadUrl=string,expiresAt=string}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /backup/executions/{executionId}/download [get]
func (h *BackupHandler) DownloadBackup(c *gin.Context) {
	executionID := c.Param("executionId")
	if executionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "execution ID is required"})
		return
	}

	downloadURL, err := h.backupService.DownloadBackup(c.Request.Context(), executionID)
	if err != nil {
		if err.Error() == "backup execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// URL expires in 1 hour
	expiresAt := time.Now().Add(time.Hour)

	c.JSON(http.StatusOK, gin.H{
		"downloadUrl": downloadURL,
		"expiresAt":   expiresAt.Format(time.RFC3339),
	})
}