package backup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	"github.com/hexabase/hexabase-ai/api/internal/domain/workspace"
	"k8s.io/client-go/kubernetes"
)

// Service implements the backup service interface
type Service struct {
	repo           backup.Repository
	proxmoxRepo    backup.ProxmoxRepository
	appRepo        application.Repository
	workspaceRepo  workspace.Repository
	k8sClient      kubernetes.Interface
	executor       *BackupExecutor
	logger         *slog.Logger
}

// NewService creates a new backup service
func NewService(
	repo backup.Repository,
	proxmoxRepo backup.ProxmoxRepository,
	appRepo application.Repository,
	workspaceRepo workspace.Repository,
	k8sClient kubernetes.Interface,
	encryptionKey string,
) backup.Service {
	logger := slog.Default()
	
	// Create backup executor with default configuration
	executorConfig := &ExecutorConfig{
		TempDir:             "/tmp/hexabase-backups",
		EncryptionKey:       encryptionKey,
		DatabaseBackupImage: "hexabase/db-backup:latest",
		BackupJobTimeout:    30 * time.Minute,
		RetryAttempts:       3,
		MaxConcurrentBackups: 5,
	}
	
	executor := NewBackupExecutor(k8sClient, proxmoxRepo, logger, executorConfig)
	
	return &Service{
		repo:          repo,
		proxmoxRepo:   proxmoxRepo,
		appRepo:       appRepo,
		workspaceRepo: workspaceRepo,
		k8sClient:     k8sClient,
		executor:      executor,
		logger:        logger,
	}
}

// CreateBackupStorage creates a new backup storage
func (s *Service) CreateBackupStorage(ctx context.Context, workspaceID string, req backup.CreateBackupStorageRequest) (*backup.BackupStorage, error) {
	// Verify workspace exists and is on Dedicated Plan
	ws, err := s.workspaceRepo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Plan != workspace.WorkspacePlanDedicated {
		return nil, errors.New("backup storage is only available for Dedicated Plan workspaces")
	}

	// Check if storage with same name already exists
	existing, _ := s.repo.GetBackupStorageByName(ctx, workspaceID, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("backup storage with name %s already exists", req.Name)
	}

	// Validate storage type
	if !req.Type.IsValid() {
		return nil, errors.New("invalid storage type")
	}

	// Create storage entity
	storage := &backup.BackupStorage{
		ID:               uuid.New().String(),
		WorkspaceID:      workspaceID,
		Name:             req.Name,
		Type:             req.Type,
		ProxmoxStorageID: req.ProxmoxStorageID,
		ProxmoxNodeID:    req.ProxmoxNodeID,
		CapacityGB:       req.CapacityGB,
		UsedGB:           0,
		Status:           backup.StorageStatusPending,
		ConnectionConfig: req.ConnectionConfig,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save to database
	if err := s.repo.CreateBackupStorage(ctx, storage); err != nil {
		return nil, fmt.Errorf("failed to create backup storage: %w", err)
	}

	// Create storage in Proxmox asynchronously
	go s.provisionStorage(context.Background(), storage)

	return storage, nil
}

// provisionStorage handles the actual storage provisioning in Proxmox
func (s *Service) provisionStorage(ctx context.Context, storage *backup.BackupStorage) {
	// Update status to creating
	storage.Status = backup.StorageStatusCreating
	if err := s.repo.UpdateBackupStorage(ctx, storage); err != nil {
		s.logger.Error("failed to update storage status", "error", err, "storageID", storage.ID)
		return
	}

	// Create Proxmox storage configuration
	config := backup.ProxmoxStorageConfig{
		Name:    fmt.Sprintf("backup-%s-%s", storage.WorkspaceID, storage.ID),
		Type:    string(storage.Type),
		Content: []string{"backup", "vztmpl"},
		Options: map[string]string{
			"nodes": storage.ProxmoxNodeID,
		},
	}

	// Add type-specific configuration
	switch storage.Type {
	case backup.StorageTypeNFS:
		if server, ok := storage.ConnectionConfig["server"].(string); ok {
			config.Server = server
		}
		if export, ok := storage.ConnectionConfig["export"].(string); ok {
			config.Export = export
		}
	case backup.StorageTypeLocal:
		config.Path = fmt.Sprintf("/var/lib/vz/backup-%s", storage.ID)
	}

	// Create storage in Proxmox
	storageID, err := s.proxmoxRepo.CreateStorage(ctx, storage.ProxmoxNodeID, config)
	if err != nil {
		storage.Status = backup.StorageStatusFailed
		storage.ErrorMessage = err.Error()
		s.repo.UpdateBackupStorage(ctx, storage)
		s.logger.Error("failed to create Proxmox storage", "error", err, "storageID", storage.ID)
		return
	}

	// Set storage quota
	if err := s.proxmoxRepo.SetStorageQuota(ctx, storage.ProxmoxNodeID, storageID, storage.CapacityGB); err != nil {
		s.logger.Warn("failed to set storage quota", "error", err, "storageID", storage.ID)
	}

	// Update storage with Proxmox ID and mark as active
	storage.ProxmoxStorageID = storageID
	storage.Status = backup.StorageStatusActive
	storage.ErrorMessage = ""
	if err := s.repo.UpdateBackupStorage(ctx, storage); err != nil {
		s.logger.Error("failed to update storage after provisioning", "error", err, "storageID", storage.ID)
	}
}

// GetBackupStorage retrieves a backup storage
func (s *Service) GetBackupStorage(ctx context.Context, storageID string) (*backup.BackupStorage, error) {
	return s.repo.GetBackupStorage(ctx, storageID)
}

// ListBackupStorages lists all backup storages for a workspace
func (s *Service) ListBackupStorages(ctx context.Context, workspaceID string) ([]backup.BackupStorage, error) {
	return s.repo.ListBackupStorages(ctx, workspaceID)
}

// UpdateBackupStorage updates a backup storage
func (s *Service) UpdateBackupStorage(ctx context.Context, storageID string, req backup.UpdateBackupStorageRequest) (*backup.BackupStorage, error) {
	storage, err := s.repo.GetBackupStorage(ctx, storageID)
	if err != nil {
		return nil, err
	}

	// Check if storage can be updated
	if storage.Status != backup.StorageStatusActive {
		return nil, fmt.Errorf("cannot update storage in status %s", storage.Status)
	}

	// Update fields
	if req.Name != "" {
		storage.Name = req.Name
	}
	if req.CapacityGB > 0 && req.CapacityGB != storage.CapacityGB {
		// Check if new capacity is sufficient
		if req.CapacityGB < storage.UsedGB {
			return nil, fmt.Errorf("new capacity (%dGB) is less than used space (%dGB)", req.CapacityGB, storage.UsedGB)
		}
		storage.CapacityGB = req.CapacityGB
		
		// Update quota in Proxmox
		if err := s.proxmoxRepo.SetStorageQuota(ctx, storage.ProxmoxNodeID, storage.ProxmoxStorageID, req.CapacityGB); err != nil {
			s.logger.Warn("failed to update storage quota", "error", err, "storageID", storage.ID)
		}
	}
	if req.ConnectionConfig != nil {
		storage.ConnectionConfig = req.ConnectionConfig
	}

	storage.UpdatedAt = time.Now()

	// Save changes
	if err := s.repo.UpdateBackupStorage(ctx, storage); err != nil {
		return nil, err
	}

	return storage, nil
}

// DeleteBackupStorage deletes a backup storage
func (s *Service) DeleteBackupStorage(ctx context.Context, storageID string) error {
	storage, err := s.repo.GetBackupStorage(ctx, storageID)
	if err != nil {
		return err
	}

	// Check if storage can be deleted
	if storage.Status == backup.StorageStatusDeleting {
		return errors.New("storage is already being deleted")
	}

	// Check if there are any active backup policies using this storage
	policies, err := s.repo.ListBackupPolicies(ctx, storage.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to check backup policies: %w", err)
	}

	for _, policy := range policies {
		if policy.StorageID == storageID {
			return fmt.Errorf("cannot delete storage: used by backup policy for application %s", policy.ApplicationID)
		}
	}

	// Update status to deleting
	storage.Status = backup.StorageStatusDeleting
	s.repo.UpdateBackupStorage(ctx, storage)

	// Delete from Proxmox
	if err := s.proxmoxRepo.DeleteStorage(ctx, storage.ProxmoxNodeID, storage.ProxmoxStorageID); err != nil {
		s.logger.Error("failed to delete Proxmox storage", "error", err, "storageID", storage.ID)
		// Continue with database deletion even if Proxmox deletion fails
	}

	// Delete from database
	return s.repo.DeleteBackupStorage(ctx, storageID)
}

// GetStorageUsage retrieves storage usage information
func (s *Service) GetStorageUsage(ctx context.Context, storageID string) (*backup.BackupStorageUsage, error) {
	return s.repo.GetStorageUsage(ctx, storageID)
}

// GetWorkspaceStorageUsage retrieves storage usage for all storages in a workspace
func (s *Service) GetWorkspaceStorageUsage(ctx context.Context, workspaceID string) ([]backup.BackupStorageUsage, error) {
	return s.repo.GetWorkspaceStorageUsage(ctx, workspaceID)
}

// CreateBackupPolicy creates a backup policy for an application
func (s *Service) CreateBackupPolicy(ctx context.Context, applicationID string, req backup.CreateBackupPolicyRequest) (*backup.BackupPolicy, error) {
	// Get application
	app, err := s.appRepo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	// Verify workspace is on Dedicated Plan
	ws, err := s.workspaceRepo.GetWorkspace(ctx, app.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Plan != workspace.WorkspacePlanDedicated {
		return nil, errors.New("backup policies are only available for Dedicated Plan workspaces")
	}

	// Check if policy already exists for this application
	existing, _ := s.repo.GetBackupPolicyByApplication(ctx, applicationID)
	if existing != nil {
		return nil, errors.New("backup policy already exists for this application")
	}

	// Verify storage exists and is active
	storage, err := s.repo.GetBackupStorage(ctx, req.StorageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup storage: %w", err)
	}

	if storage.WorkspaceID != app.WorkspaceID {
		return nil, errors.New("storage does not belong to the same workspace")
	}

	if storage.Status != backup.StorageStatusActive {
		return nil, fmt.Errorf("storage is not active (status: %s)", storage.Status)
	}

	// Validate backup type
	if !req.BackupType.IsValid() {
		return nil, errors.New("invalid backup type")
	}

	// Create policy
	policy := &backup.BackupPolicy{
		ID:                 uuid.New().String(),
		ApplicationID:      applicationID,
		StorageID:          req.StorageID,
		Enabled:            req.Enabled,
		Schedule:           req.Schedule,
		RetentionDays:      req.RetentionDays,
		BackupType:         req.BackupType,
		IncludeVolumes:     req.IncludeVolumes,
		IncludeDatabase:    req.IncludeDatabase,
		IncludeConfig:      req.IncludeConfig,
		CompressionEnabled: req.CompressionEnabled,
		EncryptionEnabled:  req.EncryptionEnabled,
		EncryptionKeyRef:   req.EncryptionKeyRef,
		PreBackupHook:      req.PreBackupHook,
		PostBackupHook:     req.PostBackupHook,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Save to database
	if err := s.repo.CreateBackupPolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to create backup policy: %w", err)
	}

	// Create CronJob for scheduled backups if enabled
	if policy.Enabled && policy.Schedule != "" {
		go s.createBackupCronJob(context.Background(), app, policy)
	}

	return policy, nil
}

// createBackupCronJob creates a CronJob for scheduled backups
func (s *Service) createBackupCronJob(ctx context.Context, app *application.Application, policy *backup.BackupPolicy) {
	// This will integrate with the CronJob feature to schedule backup executions
	// Implementation will be added when integrating with CronJob feature
	s.logger.Info("creating backup CronJob", "applicationID", app.ID, "policyID", policy.ID, "schedule", policy.Schedule)
}

// GetBackupPolicy retrieves a backup policy
func (s *Service) GetBackupPolicy(ctx context.Context, policyID string) (*backup.BackupPolicy, error) {
	return s.repo.GetBackupPolicy(ctx, policyID)
}

// GetBackupPolicyByApplication retrieves a backup policy by application ID
func (s *Service) GetBackupPolicyByApplication(ctx context.Context, applicationID string) (*backup.BackupPolicy, error) {
	policy, err := s.repo.GetBackupPolicyByApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, fmt.Errorf("no backup policy found for application")
	}
	return policy, nil
}

// UpdateBackupPolicy updates a backup policy
func (s *Service) UpdateBackupPolicy(ctx context.Context, policyID string, req backup.UpdateBackupPolicyRequest) (*backup.BackupPolicy, error) {
	policy, err := s.repo.GetBackupPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Enabled != nil {
		policy.Enabled = *req.Enabled
	}
	if req.Schedule != "" {
		policy.Schedule = req.Schedule
	}
	if req.RetentionDays != nil {
		policy.RetentionDays = *req.RetentionDays
	}
	if req.BackupType != "" && req.BackupType.IsValid() {
		policy.BackupType = req.BackupType
	}
	if req.IncludeVolumes != nil {
		policy.IncludeVolumes = *req.IncludeVolumes
	}
	if req.IncludeDatabase != nil {
		policy.IncludeDatabase = *req.IncludeDatabase
	}
	if req.IncludeConfig != nil {
		policy.IncludeConfig = *req.IncludeConfig
	}
	if req.CompressionEnabled != nil {
		policy.CompressionEnabled = *req.CompressionEnabled
	}
	if req.EncryptionEnabled != nil {
		policy.EncryptionEnabled = *req.EncryptionEnabled
	}
	if req.EncryptionKeyRef != "" {
		policy.EncryptionKeyRef = req.EncryptionKeyRef
	}
	if req.PreBackupHook != "" {
		policy.PreBackupHook = req.PreBackupHook
	}
	if req.PostBackupHook != "" {
		policy.PostBackupHook = req.PostBackupHook
	}

	policy.UpdatedAt = time.Now()

	// Save changes
	if err := s.repo.UpdateBackupPolicy(ctx, policy); err != nil {
		return nil, err
	}

	// Update CronJob if schedule changed
	// TODO: Implement CronJob update when integrating with CronJob feature

	return policy, nil
}

// DeleteBackupPolicy deletes a backup policy
func (s *Service) DeleteBackupPolicy(ctx context.Context, policyID string) error {
	_, err := s.repo.GetBackupPolicy(ctx, policyID)
	if err != nil {
		return err
	}

	// Delete associated CronJob
	// TODO: Implement CronJob deletion when integrating with CronJob feature

	// Delete from database
	return s.repo.DeleteBackupPolicy(ctx, policyID)
}

// TriggerManualBackup manually triggers a backup for an application
func (s *Service) TriggerManualBackup(ctx context.Context, applicationID string, req backup.TriggerBackupRequest) (*backup.BackupExecution, error) {
	// Get backup policy
	policy, err := s.repo.GetBackupPolicyByApplication(ctx, applicationID)
	if err != nil || policy == nil {
		return nil, errors.New("no backup policy found for application")
	}

	// Create backup execution
	execution := &backup.BackupExecution{
		ID:        uuid.New().String(),
		PolicyID:  policy.ID,
		Status:    backup.BackupExecutionStatusRunning,
		StartedAt: time.Now(),
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
	}

	// Override backup type if specified
	backupType := policy.BackupType
	if req.BackupType != "" && req.BackupType.IsValid() {
		backupType = req.BackupType
	}

	// Save execution record
	if err := s.repo.CreateBackupExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create backup execution: %w", err)
	}

	// Execute backup asynchronously
	go s.executeBackup(context.Background(), execution, policy, backupType)

	return execution, nil
}

// executeBackup performs the actual backup operation
func (s *Service) executeBackup(ctx context.Context, execution *backup.BackupExecution, policy *backup.BackupPolicy, backupType backup.BackupType) {
	s.logger.Info("executing backup", "executionID", execution.ID, "policyID", policy.ID, "backupType", backupType)

	// Use the backup executor to perform the actual backup
	if err := s.executor.ExecuteBackup(ctx, execution, policy); err != nil {
		s.logger.Error("backup execution failed", "error", err, "executionID", execution.ID)
		
		// Update execution status to failed
		execution.Status = backup.BackupExecutionStatusFailed
		execution.ErrorMessage = err.Error()
		completedAt := time.Now()
		execution.CompletedAt = &completedAt
		
		if updateErr := s.repo.UpdateBackupExecution(ctx, execution); updateErr != nil {
			s.logger.Error("failed to update failed backup execution", "error", updateErr, "executionID", execution.ID)
		}
		return
	}

	// Update storage usage based on actual backup size
	if execution.CompressedSizeBytes > 0 {
		// Convert bytes to GB (rounded up)
		usedGB := int((execution.CompressedSizeBytes + (1024*1024*1024 - 1)) / (1024 * 1024 * 1024))
		
		// Get current storage usage
		storage, err := s.repo.GetBackupStorage(ctx, policy.StorageID)
		if err != nil {
			s.logger.Error("failed to get backup storage for usage update", "error", err, "storageID", policy.StorageID)
		} else {
			// Update storage usage
			newUsage := storage.UsedGB + usedGB
			if err := s.repo.UpdateStorageUsage(ctx, policy.StorageID, newUsage); err != nil {
				s.logger.Error("failed to update storage usage", "error", err, "storageID", policy.StorageID)
			}
		}
	}

	s.logger.Info("backup execution completed successfully", 
		"executionID", execution.ID, 
		"sizeBytes", execution.SizeBytes,
		"compressedSizeBytes", execution.CompressedSizeBytes,
		"backupPath", execution.BackupPath)
}

// GetBackupExecution retrieves a backup execution
func (s *Service) GetBackupExecution(ctx context.Context, executionID string) (*backup.BackupExecution, error) {
	return s.repo.GetBackupExecution(ctx, executionID)
}

// ListBackupExecutions lists backup executions for an application
func (s *Service) ListBackupExecutions(ctx context.Context, applicationID string, limit, offset int) ([]backup.BackupExecution, int, error) {
	// Get policy first
	policy, err := s.repo.GetBackupPolicyByApplication(ctx, applicationID)
	if err != nil || policy == nil {
		return []backup.BackupExecution{}, 0, nil
	}

	return s.repo.ListBackupExecutions(ctx, policy.ID, limit, offset)
}

// GetLatestBackup retrieves the latest successful backup for an application
func (s *Service) GetLatestBackup(ctx context.Context, applicationID string) (*backup.BackupExecution, error) {
	// Get policy first
	policy, err := s.repo.GetBackupPolicyByApplication(ctx, applicationID)
	if err != nil || policy == nil {
		return nil, errors.New("no backup policy found for application")
	}

	return s.repo.GetLatestBackupExecution(ctx, policy.ID)
}

// ProcessScheduledBackup processes a scheduled backup execution
func (s *Service) ProcessScheduledBackup(ctx context.Context, policyID string) (*backup.BackupExecution, error) {
	// Get policy
	policy, err := s.repo.GetBackupPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	if !policy.Enabled {
		return nil, errors.New("backup policy is disabled")
	}

	// Create execution
	execution := &backup.BackupExecution{
		ID:        uuid.New().String(),
		PolicyID:  policyID,
		Status:    backup.BackupExecutionStatusRunning,
		StartedAt: time.Now(),
		Metadata: map[string]interface{}{
			"scheduled": true,
			"schedule":  policy.Schedule,
		},
		CreatedAt: time.Now(),
	}

	// Save execution record
	if err := s.repo.CreateBackupExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create backup execution: %w", err)
	}

	// Execute backup asynchronously
	go s.executeBackup(context.Background(), execution, policy, policy.BackupType)

	return execution, nil
}

// CleanupOldBackups removes old backups based on retention policy
func (s *Service) CleanupOldBackups(ctx context.Context, policyID string) error {
	policy, err := s.repo.GetBackupPolicy(ctx, policyID)
	if err != nil {
		return err
	}

	// Cleanup database records
	if err := s.repo.CleanupOldBackups(ctx, policyID, policy.RetentionDays); err != nil {
		return fmt.Errorf("failed to cleanup old backups: %w", err)
	}

	// TODO: Delete actual backup files from storage

	return nil
}

// RestoreBackup initiates a restore operation from a backup
func (s *Service) RestoreBackup(ctx context.Context, backupExecutionID string, req backup.RestoreBackupRequest) (*backup.BackupRestore, error) {
	// Get backup execution
	execution, err := s.repo.GetBackupExecution(ctx, backupExecutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup execution: %w", err)
	}

	if execution.Status != backup.BackupExecutionStatusSucceeded {
		return nil, errors.New("can only restore from successful backups")
	}

	// Get policy to determine application
	policy, err := s.repo.GetBackupPolicy(ctx, execution.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup policy: %w", err)
	}

	// Validate restore type
	if !req.RestoreType.IsValid() {
		return nil, errors.New("invalid restore type")
	}

	// Create restore record
	restore := &backup.BackupRestore{
		ID:                uuid.New().String(),
		BackupExecutionID: backupExecutionID,
		ApplicationID:     policy.ApplicationID,
		Status:            backup.RestoreStatusPending,
		RestoreType:       req.RestoreType,
		RestoreOptions:    req.RestoreOptions,
		CreatedAt:         time.Now(),
	}

	// If restoring to new application, set the ID
	if req.NewApplicationID != "" {
		restore.NewApplicationID = req.NewApplicationID
	}

	// Save restore record
	if err := s.repo.CreateBackupRestore(ctx, restore); err != nil {
		return nil, fmt.Errorf("failed to create restore record: %w", err)
	}

	// Execute restore asynchronously
	go s.executeRestore(context.Background(), restore, execution)

	return restore, nil
}

// executeRestore performs the actual restore operation
func (s *Service) executeRestore(ctx context.Context, restore *backup.BackupRestore, execution *backup.BackupExecution) {
	// Update status to restoring
	now := time.Now()
	restore.Status = backup.RestoreStatusRestoring
	restore.StartedAt = &now
	if err := s.repo.UpdateBackupRestore(ctx, restore); err != nil {
		s.logger.Error("failed to update restore status", "error", err, "restoreID", restore.ID)
		return
	}

	s.logger.Info("executing restore", "restoreID", restore.ID, "backupID", execution.ID, "restoreType", restore.RestoreType)

	// Get the backup policy to access storage information
	policy, err := s.repo.GetBackupPolicy(ctx, execution.PolicyID)
	if err != nil {
		s.logger.Error("failed to get backup policy for restore", "error", err, "policyID", execution.PolicyID)
		restore.Status = backup.RestoreStatusFailed
		restore.ErrorMessage = fmt.Sprintf("failed to get backup policy: %v", err)
		completedAt := time.Now()
		restore.CompletedAt = &completedAt
		s.repo.UpdateBackupRestore(ctx, restore)
		return
	}

	// Use the backup executor to perform the actual restore
	if err := s.executor.RestoreBackup(ctx, restore, execution, policy); err != nil {
		s.logger.Error("restore execution failed", "error", err, "restoreID", restore.ID)
		
		// Update restore status to failed
		restore.Status = backup.RestoreStatusFailed
		restore.ErrorMessage = err.Error()
		completedAt := time.Now()
		restore.CompletedAt = &completedAt
		
		if updateErr := s.repo.UpdateBackupRestore(ctx, restore); updateErr != nil {
			s.logger.Error("failed to update failed restore", "error", updateErr, "restoreID", restore.ID)
		}
		return
	}

	// Restore completed successfully
	s.logger.Info("restore execution completed successfully", 
		"restoreID", restore.ID, 
		"backupID", execution.ID,
		"restoreType", restore.RestoreType)
}

// GetBackupRestore retrieves a restore operation
func (s *Service) GetBackupRestore(ctx context.Context, restoreID string) (*backup.BackupRestore, error) {
	return s.repo.GetBackupRestore(ctx, restoreID)
}

// ListBackupRestores lists restore operations for an application
func (s *Service) ListBackupRestores(ctx context.Context, applicationID string, limit, offset int) ([]backup.BackupRestore, int, error) {
	return s.repo.ListBackupRestores(ctx, applicationID, limit, offset)
}

// ValidateBackup validates that a backup is restorable
func (s *Service) ValidateBackup(ctx context.Context, backupExecutionID string) error {
	execution, err := s.repo.GetBackupExecution(ctx, backupExecutionID)
	if err != nil {
		return err
	}

	if execution.Status != backup.BackupExecutionStatusSucceeded {
		return errors.New("backup did not complete successfully")
	}

	if execution.BackupPath == "" {
		return errors.New("backup path is missing")
	}

	// TODO: Verify backup file exists in storage
	// TODO: Verify backup integrity (checksum, etc.)

	return nil
}

// GetBackupManifest retrieves the manifest of backed up resources
func (s *Service) GetBackupManifest(ctx context.Context, backupExecutionID string) (map[string]interface{}, error) {
	execution, err := s.repo.GetBackupExecution(ctx, backupExecutionID)
	if err != nil {
		return nil, err
	}

	if execution.BackupManifest == nil {
		return map[string]interface{}{}, nil
	}

	return execution.BackupManifest, nil
}

// DownloadBackup generates a pre-signed URL for downloading a backup
func (s *Service) DownloadBackup(ctx context.Context, backupExecutionID string) (string, error) {
	execution, err := s.repo.GetBackupExecution(ctx, backupExecutionID)
	if err != nil {
		return "", err
	}

	if execution.Status != backup.BackupExecutionStatusSucceeded {
		return "", errors.New("backup did not complete successfully")
	}

	if execution.BackupPath == "" {
		return "", errors.New("backup path is missing")
	}

	// TODO: Generate pre-signed URL for backup download
	// This would integrate with the storage backend (S3, Proxmox, etc.)
	
	// For now, return a placeholder URL
	downloadURL := fmt.Sprintf("https://backup.hexabase.io/download/%s/%s", execution.PolicyID, execution.ID)
	
	return downloadURL, nil
}