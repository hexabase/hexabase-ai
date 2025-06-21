package domain

import (
	"context"
)

// Service defines the interface for backup business logic
type Service interface {
	// BackupStorage operations
	CreateBackupStorage(ctx context.Context, workspaceID string, req CreateBackupStorageRequest) (*BackupStorage, error)
	GetBackupStorage(ctx context.Context, storageID string) (*BackupStorage, error)
	ListBackupStorages(ctx context.Context, workspaceID string) ([]BackupStorage, error)
	UpdateBackupStorage(ctx context.Context, storageID string, req UpdateBackupStorageRequest) (*BackupStorage, error)
	DeleteBackupStorage(ctx context.Context, storageID string) error
	GetStorageUsage(ctx context.Context, storageID string) (*BackupStorageUsage, error)
	GetWorkspaceStorageUsage(ctx context.Context, workspaceID string) ([]BackupStorageUsage, error)

	// BackupPolicy operations
	CreateBackupPolicy(ctx context.Context, applicationID string, req CreateBackupPolicyRequest) (*BackupPolicy, error)
	GetBackupPolicy(ctx context.Context, policyID string) (*BackupPolicy, error)
	GetBackupPolicyByApplication(ctx context.Context, applicationID string) (*BackupPolicy, error)
	UpdateBackupPolicy(ctx context.Context, policyID string, req UpdateBackupPolicyRequest) (*BackupPolicy, error)
	DeleteBackupPolicy(ctx context.Context, policyID string) error

	// Backup execution operations
	TriggerManualBackup(ctx context.Context, applicationID string, req TriggerBackupRequest) (*BackupExecution, error)
	GetBackupExecution(ctx context.Context, executionID string) (*BackupExecution, error)
	ListBackupExecutions(ctx context.Context, applicationID string, limit, offset int) ([]BackupExecution, int, error)
	GetLatestBackup(ctx context.Context, applicationID string) (*BackupExecution, error)
	ProcessScheduledBackup(ctx context.Context, policyID string) (*BackupExecution, error)
	CleanupOldBackups(ctx context.Context, policyID string) error

	// Restore operations
	RestoreBackup(ctx context.Context, backupExecutionID string, req RestoreBackupRequest) (*BackupRestore, error)
	GetBackupRestore(ctx context.Context, restoreID string) (*BackupRestore, error)
	ListBackupRestores(ctx context.Context, applicationID string, limit, offset int) ([]BackupRestore, int, error)
	ValidateBackup(ctx context.Context, backupExecutionID string) error

	// Backup content operations
	GetBackupManifest(ctx context.Context, backupExecutionID string) (map[string]interface{}, error)
	DownloadBackup(ctx context.Context, backupExecutionID string) (string, error) // Returns pre-signed URL
}