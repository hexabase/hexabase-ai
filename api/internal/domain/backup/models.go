package backup

import (
	"time"
)

// StorageType represents the type of backup storage
type StorageType string

const (
	StorageTypeNFS     StorageType = "nfs"
	StorageTypeCeph    StorageType = "ceph"
	StorageTypeLocal   StorageType = "local"
	StorageTypeProxmox StorageType = "proxmox"
)

// StorageStatus represents the status of backup storage
type StorageStatus string

const (
	StorageStatusPending  StorageStatus = "pending"
	StorageStatusCreating StorageStatus = "creating"
	StorageStatusActive   StorageStatus = "active"
	StorageStatusFailed   StorageStatus = "failed"
	StorageStatusDeleting StorageStatus = "deleting"
)

// BackupType represents the type of backup
type BackupType string

const (
	BackupTypeFull        BackupType = "full"
	BackupTypeIncremental BackupType = "incremental"
)

// BackupExecutionStatus represents the status of a backup execution
type BackupExecutionStatus string

const (
	BackupExecutionStatusRunning   BackupExecutionStatus = "running"
	BackupExecutionStatusSucceeded BackupExecutionStatus = "succeeded"
	BackupExecutionStatusFailed    BackupExecutionStatus = "failed"
	BackupExecutionStatusCancelled BackupExecutionStatus = "cancelled"
)

// RestoreStatus represents the status of a restore operation
type RestoreStatus string

const (
	RestoreStatusPending   RestoreStatus = "pending"
	RestoreStatusPreparing RestoreStatus = "preparing"
	RestoreStatusRestoring RestoreStatus = "restoring"
	RestoreStatusVerifying RestoreStatus = "verifying"
	RestoreStatusCompleted RestoreStatus = "completed"
	RestoreStatusFailed    RestoreStatus = "failed"
)

// RestoreType represents the type of restore operation
type RestoreType string

const (
	RestoreTypeFull      RestoreType = "full"
	RestoreTypeSelective RestoreType = "selective"
)

// BackupStorage represents a backup storage configuration
type BackupStorage struct {
	ID               string                 `json:"id"`
	WorkspaceID      string                 `json:"workspace_id"`
	Name             string                 `json:"name"`
	Type             StorageType            `json:"type"`
	ProxmoxStorageID string                 `json:"proxmox_storage_id"`
	ProxmoxNodeID    string                 `json:"proxmox_node_id"`
	CapacityGB       int                    `json:"capacity_gb"`
	UsedGB           int                    `json:"used_gb"`
	Status           StorageStatus          `json:"status"`
	ConnectionConfig map[string]interface{} `json:"connection_config,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// BackupPolicy represents a backup policy for an application
type BackupPolicy struct {
	ID                 string     `json:"id"`
	ApplicationID      string     `json:"application_id"`
	StorageID          string     `json:"storage_id"`
	Enabled            bool       `json:"enabled"`
	Schedule           string     `json:"schedule"` // Cron expression
	RetentionDays      int        `json:"retention_days"`
	BackupType         BackupType `json:"backup_type"`
	IncludeVolumes     bool       `json:"include_volumes"`
	IncludeDatabase    bool       `json:"include_database"`
	IncludeConfig      bool       `json:"include_config"`
	CompressionEnabled bool       `json:"compression_enabled"`
	EncryptionEnabled  bool       `json:"encryption_enabled"`
	EncryptionKeyRef   string     `json:"encryption_key_ref,omitempty"`
	PreBackupHook      string     `json:"pre_backup_hook,omitempty"`
	PostBackupHook     string     `json:"post_backup_hook,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// BackupExecution represents a single backup execution
type BackupExecution struct {
	ID                 string                 `json:"id"`
	PolicyID           string                 `json:"policy_id"`
	CronJobExecutionID string                 `json:"cronjob_execution_id,omitempty"`
	Status             BackupExecutionStatus  `json:"status"`
	SizeBytes          int64                  `json:"size_bytes,omitempty"`
	CompressedSizeBytes int64                 `json:"compressed_size_bytes,omitempty"`
	BackupPath         string                 `json:"backup_path,omitempty"`
	BackupManifest     map[string]interface{} `json:"backup_manifest,omitempty"`
	StartedAt          time.Time              `json:"started_at"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}

// BackupRestore represents a restore operation from a backup
type BackupRestore struct {
	ID                string                 `json:"id"`
	BackupExecutionID string                 `json:"backup_execution_id"`
	ApplicationID     string                 `json:"application_id"`
	Status            RestoreStatus          `json:"status"`
	RestoreType       RestoreType            `json:"restore_type"`
	RestoreOptions    map[string]interface{} `json:"restore_options,omitempty"`
	NewApplicationID  string                 `json:"new_application_id,omitempty"`
	StartedAt         *time.Time             `json:"started_at,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
	ValidationResults map[string]interface{} `json:"validation_results,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// CreateBackupStorageRequest represents a request to create backup storage
type CreateBackupStorageRequest struct {
	Name             string                 `json:"name"`
	Type             StorageType            `json:"type"`
	ProxmoxStorageID string                 `json:"proxmox_storage_id"`
	ProxmoxNodeID    string                 `json:"proxmox_node_id"`
	CapacityGB       int                    `json:"capacity_gb"`
	ConnectionConfig map[string]interface{} `json:"connection_config,omitempty"`
}

// UpdateBackupStorageRequest represents a request to update backup storage
type UpdateBackupStorageRequest struct {
	Name             string                 `json:"name,omitempty"`
	CapacityGB       int                    `json:"capacity_gb,omitempty"`
	ConnectionConfig map[string]interface{} `json:"connection_config,omitempty"`
}

// CreateBackupPolicyRequest represents a request to create a backup policy
type CreateBackupPolicyRequest struct {
	StorageID          string     `json:"storage_id"`
	Enabled            bool       `json:"enabled"`
	Schedule           string     `json:"schedule"`
	RetentionDays      int        `json:"retention_days"`
	BackupType         BackupType `json:"backup_type"`
	IncludeVolumes     bool       `json:"include_volumes"`
	IncludeDatabase    bool       `json:"include_database"`
	IncludeConfig      bool       `json:"include_config"`
	CompressionEnabled bool       `json:"compression_enabled"`
	EncryptionEnabled  bool       `json:"encryption_enabled"`
	EncryptionKeyRef   string     `json:"encryption_key_ref,omitempty"`
	PreBackupHook      string     `json:"pre_backup_hook,omitempty"`
	PostBackupHook     string     `json:"post_backup_hook,omitempty"`
}

// UpdateBackupPolicyRequest represents a request to update a backup policy
type UpdateBackupPolicyRequest struct {
	Enabled            *bool      `json:"enabled,omitempty"`
	Schedule           string     `json:"schedule,omitempty"`
	RetentionDays      *int       `json:"retention_days,omitempty"`
	BackupType         BackupType `json:"backup_type,omitempty"`
	IncludeVolumes     *bool      `json:"include_volumes,omitempty"`
	IncludeDatabase    *bool      `json:"include_database,omitempty"`
	IncludeConfig      *bool      `json:"include_config,omitempty"`
	CompressionEnabled *bool      `json:"compression_enabled,omitempty"`
	EncryptionEnabled  *bool      `json:"encryption_enabled,omitempty"`
	EncryptionKeyRef   string     `json:"encryption_key_ref,omitempty"`
	PreBackupHook      string     `json:"pre_backup_hook,omitempty"`
	PostBackupHook     string     `json:"post_backup_hook,omitempty"`
}

// TriggerBackupRequest represents a request to manually trigger a backup
type TriggerBackupRequest struct {
	ApplicationID string                 `json:"application_id"`
	BackupType    BackupType             `json:"backup_type,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// RestoreBackupRequest represents a request to restore from a backup
type RestoreBackupRequest struct {
	RestoreType      RestoreType            `json:"restore_type"`
	RestoreOptions   map[string]interface{} `json:"restore_options,omitempty"`
	TargetNamespace  string                 `json:"target_namespace,omitempty"`
	NewApplicationID string                 `json:"new_application_id,omitempty"`
}

// BackupStorageUsage represents storage usage information
type BackupStorageUsage struct {
	StorageID      string  `json:"storage_id"`
	TotalGB        int     `json:"total_gb"`
	UsedGB         int     `json:"used_gb"`
	AvailableGB    int     `json:"available_gb"`
	UsagePercent   float64 `json:"usage_percent"`
	BackupCount    int     `json:"backup_count"`
	OldestBackup   *time.Time `json:"oldest_backup,omitempty"`
	LatestBackup   *time.Time `json:"latest_backup,omitempty"`
}

// BackupExecutionList represents a paginated list of backup executions
type BackupExecutionList struct {
	Executions []BackupExecution `json:"executions"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

// BackupRestoreList represents a paginated list of restore operations
type BackupRestoreList struct {
	Restores []BackupRestore `json:"restores"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// IsValid checks if the storage type is valid
func (t StorageType) IsValid() bool {
	switch t {
	case StorageTypeNFS, StorageTypeCeph, StorageTypeLocal, StorageTypeProxmox:
		return true
	default:
		return false
	}
}

// IsValid checks if the backup type is valid
func (t BackupType) IsValid() bool {
	switch t {
	case BackupTypeFull, BackupTypeIncremental:
		return true
	default:
		return false
	}
}

// IsValid checks if the restore type is valid
func (t RestoreType) IsValid() bool {
	switch t {
	case RestoreTypeFull, RestoreTypeSelective:
		return true
	default:
		return false
	}
}

// CanTransition checks if the storage can transition to the target status
func (s StorageStatus) CanTransition(target StorageStatus) bool {
	transitions := map[StorageStatus][]StorageStatus{
		StorageStatusPending:  {StorageStatusCreating, StorageStatusFailed},
		StorageStatusCreating: {StorageStatusActive, StorageStatusFailed},
		StorageStatusActive:   {StorageStatusDeleting, StorageStatusFailed},
		StorageStatusFailed:   {StorageStatusDeleting},
		StorageStatusDeleting: {}, // Terminal state
	}

	allowed, exists := transitions[s]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// CanTransition checks if the restore can transition to the target status
func (s RestoreStatus) CanTransition(target RestoreStatus) bool {
	transitions := map[RestoreStatus][]RestoreStatus{
		RestoreStatusPending:   {RestoreStatusPreparing, RestoreStatusFailed},
		RestoreStatusPreparing: {RestoreStatusRestoring, RestoreStatusFailed},
		RestoreStatusRestoring: {RestoreStatusVerifying, RestoreStatusFailed},
		RestoreStatusVerifying: {RestoreStatusCompleted, RestoreStatusFailed},
		RestoreStatusCompleted: {}, // Terminal state
		RestoreStatusFailed:    {}, // Terminal state
	}

	allowed, exists := transitions[s]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}