package domain

import (
	"context"
)

// Repository defines the interface for backup data persistence
type Repository interface {
	// BackupStorage operations
	CreateBackupStorage(ctx context.Context, storage *BackupStorage) error
	GetBackupStorage(ctx context.Context, storageID string) (*BackupStorage, error)
	GetBackupStorageByName(ctx context.Context, workspaceID, name string) (*BackupStorage, error)
	ListBackupStorages(ctx context.Context, workspaceID string) ([]BackupStorage, error)
	UpdateBackupStorage(ctx context.Context, storage *BackupStorage) error
	DeleteBackupStorage(ctx context.Context, storageID string) error
	UpdateStorageUsage(ctx context.Context, storageID string, usedGB int) error

	// BackupPolicy operations
	CreateBackupPolicy(ctx context.Context, policy *BackupPolicy) error
	GetBackupPolicy(ctx context.Context, policyID string) (*BackupPolicy, error)
	GetBackupPolicyByApplication(ctx context.Context, applicationID string) (*BackupPolicy, error)
	ListBackupPolicies(ctx context.Context, workspaceID string) ([]BackupPolicy, error)
	UpdateBackupPolicy(ctx context.Context, policy *BackupPolicy) error
	DeleteBackupPolicy(ctx context.Context, policyID string) error
	ListEnabledPolicies(ctx context.Context) ([]BackupPolicy, error)

	// BackupExecution operations
	CreateBackupExecution(ctx context.Context, execution *BackupExecution) error
	GetBackupExecution(ctx context.Context, executionID string) (*BackupExecution, error)
	ListBackupExecutions(ctx context.Context, policyID string, limit, offset int) ([]BackupExecution, int, error)
	UpdateBackupExecution(ctx context.Context, execution *BackupExecution) error
	GetLatestBackupExecution(ctx context.Context, policyID string) (*BackupExecution, error)
	GetBackupExecutionsByApplication(ctx context.Context, applicationID string, limit, offset int) ([]BackupExecution, int, error)
	CleanupOldBackups(ctx context.Context, policyID string, retentionDays int) error

	// BackupRestore operations
	CreateBackupRestore(ctx context.Context, restore *BackupRestore) error
	GetBackupRestore(ctx context.Context, restoreID string) (*BackupRestore, error)
	ListBackupRestores(ctx context.Context, applicationID string, limit, offset int) ([]BackupRestore, int, error)
	UpdateBackupRestore(ctx context.Context, restore *BackupRestore) error

	// Storage usage operations
	GetStorageUsage(ctx context.Context, storageID string) (*BackupStorageUsage, error)
	GetWorkspaceStorageUsage(ctx context.Context, workspaceID string) ([]BackupStorageUsage, error)
}

// ProxmoxRepository defines the interface for Proxmox storage operations
type ProxmoxRepository interface {
	// Storage management
	CreateStorage(ctx context.Context, nodeID string, config ProxmoxStorageConfig) (string, error)
	DeleteStorage(ctx context.Context, nodeID, storageID string) error
	GetStorageInfo(ctx context.Context, nodeID, storageID string) (*ProxmoxStorageInfo, error)
	ResizeStorage(ctx context.Context, nodeID, storageID string, newSizeGB int) error

	// Volume management for backups
	CreateBackupVolume(ctx context.Context, nodeID, storageID string, volumeSize int64) (string, error)
	DeleteBackupVolume(ctx context.Context, nodeID, storageID, volumeID string) error
	GetVolumeInfo(ctx context.Context, nodeID, storageID, volumeID string) (*ProxmoxVolumeInfo, error)

	// Storage quota management
	SetStorageQuota(ctx context.Context, nodeID, storageID string, quotaGB int) error
	GetStorageQuota(ctx context.Context, nodeID, storageID string) (*ProxmoxStorageQuota, error)
}

// ProxmoxStorageConfig represents configuration for Proxmox storage
type ProxmoxStorageConfig struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"` // dir, nfs, ceph, etc.
	Content   []string               `json:"content"` // ["backup", "vztmpl", "iso"]
	Path      string                 `json:"path,omitempty"`
	Server    string                 `json:"server,omitempty"` // For NFS
	Export    string                 `json:"export,omitempty"` // For NFS
	Pool      string                 `json:"pool,omitempty"`   // For Ceph
	Username  string                 `json:"username,omitempty"` // For Ceph
	KeyRing   string                 `json:"keyring,omitempty"`  // For Ceph
	Options   map[string]string      `json:"options,omitempty"`
}

// ProxmoxStorageInfo represents information about a Proxmox storage
type ProxmoxStorageInfo struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Total      int64   `json:"total"`      // Total space in bytes
	Used       int64   `json:"used"`       // Used space in bytes
	Available  int64   `json:"available"`  // Available space in bytes
	Active     bool    `json:"active"`
	Enabled    bool    `json:"enabled"`
	Shared     bool    `json:"shared"`
	Content    []string `json:"content"`
}

// ProxmoxVolumeInfo represents information about a storage volume
type ProxmoxVolumeInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`      // Size in bytes
	Used      int64     `json:"used"`      // Used space in bytes
	Format    string    `json:"format"`
	Path      string    `json:"path"`
}

// ProxmoxStorageQuota represents storage quota information
type ProxmoxStorageQuota struct {
	StorageID string `json:"storage_id"`
	QuotaGB   int    `json:"quota_gb"`
	UsedGB    int    `json:"used_gb"`
	Available int    `json:"available_gb"`
}