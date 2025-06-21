package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BackupStorage represents the backup_storages table
type BackupStorage struct {
	ID               string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkspaceID      string         `gorm:"type:uuid;not null" json:"workspace_id"`
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	Type             string         `gorm:"type:varchar(50);not null" json:"type"`
	ProxmoxStorageID string         `gorm:"type:varchar(255);not null" json:"proxmox_storage_id"`
	ProxmoxNodeID    string         `gorm:"type:varchar(255);not null" json:"proxmox_node_id"`
	CapacityGB       int            `gorm:"not null" json:"capacity_gb"`
	UsedGB           int            `gorm:"default:0" json:"used_gb"`
	Status           string         `gorm:"type:varchar(50);default:'pending'" json:"status"`
	ConnectionConfig JSON           `gorm:"type:jsonb" json:"connection_config,omitempty"`
	ErrorMessage     string         `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	Workspace        *Workspace     `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
}

// TableName specifies the table name for BackupStorage
func (BackupStorage) TableName() string {
	return "backup_storages"
}

// BeforeCreate hook for BackupStorage
func (b *BackupStorage) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// BackupPolicy represents the backup_policies table
type BackupPolicy struct {
	ID                 string          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ApplicationID      string          `gorm:"type:uuid;not null" json:"application_id"`
	StorageID          string          `gorm:"type:uuid;not null" json:"storage_id"`
	Enabled            bool            `gorm:"default:true" json:"enabled"`
	Schedule           string          `gorm:"type:varchar(100);not null" json:"schedule"`
	RetentionDays      int             `gorm:"default:30" json:"retention_days"`
	BackupType         string          `gorm:"type:varchar(50);default:'full'" json:"backup_type"`
	IncludeVolumes     bool            `gorm:"default:true" json:"include_volumes"`
	IncludeDatabase    bool            `gorm:"default:true" json:"include_database"`
	IncludeConfig      bool            `gorm:"default:true" json:"include_config"`
	CompressionEnabled bool            `gorm:"default:true" json:"compression_enabled"`
	EncryptionEnabled  bool            `gorm:"default:true" json:"encryption_enabled"`
	EncryptionKeyRef   string          `gorm:"type:varchar(255)" json:"encryption_key_ref,omitempty"`
	PreBackupHook      string          `gorm:"type:text" json:"pre_backup_hook,omitempty"`
	PostBackupHook     string          `gorm:"type:text" json:"post_backup_hook,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	Application        *Application    `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
	Storage            *BackupStorage  `gorm:"foreignKey:StorageID" json:"storage,omitempty"`
}

// TableName specifies the table name for BackupPolicy
func (BackupPolicy) TableName() string {
	return "backup_policies"
}

// BeforeCreate hook for BackupPolicy
func (b *BackupPolicy) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// BackupExecution represents the backup_executions table
type BackupExecution struct {
	ID                  string              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PolicyID            string              `gorm:"type:uuid;not null" json:"policy_id"`
	CronJobExecutionID  *string             `gorm:"type:uuid" json:"cronjob_execution_id,omitempty"`
	Status              string              `gorm:"type:varchar(50);not null;default:'running'" json:"status"`
	SizeBytes           int64               `json:"size_bytes,omitempty"`
	CompressedSizeBytes int64               `json:"compressed_size_bytes,omitempty"`
	BackupPath          string              `gorm:"type:text" json:"backup_path,omitempty"`
	BackupManifest      JSON                `gorm:"type:jsonb" json:"backup_manifest,omitempty"`
	StartedAt           time.Time           `gorm:"not null" json:"started_at"`
	CompletedAt         *time.Time          `json:"completed_at,omitempty"`
	ErrorMessage        string              `gorm:"type:text" json:"error_message,omitempty"`
	Metadata            JSON                `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt           time.Time           `json:"created_at"`
	Policy              *BackupPolicy       `gorm:"foreignKey:PolicyID" json:"policy,omitempty"`
	CronJobExecution    *CronJobExecution   `gorm:"foreignKey:CronJobExecutionID" json:"cronjob_execution,omitempty"`
}

// TableName specifies the table name for BackupExecution
func (BackupExecution) TableName() string {
	return "backup_executions"
}

// BeforeCreate hook for BackupExecution
func (b *BackupExecution) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	if b.StartedAt.IsZero() {
		b.StartedAt = time.Now()
	}
	return nil
}

// BackupRestore represents the backup_restores table
type BackupRestore struct {
	ID                string           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BackupExecutionID string           `gorm:"type:uuid;not null" json:"backup_execution_id"`
	ApplicationID     string           `gorm:"type:uuid;not null" json:"application_id"`
	Status            string           `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	RestoreType       string           `gorm:"type:varchar(50);not null" json:"restore_type"`
	RestoreOptions    JSON             `gorm:"type:jsonb" json:"restore_options,omitempty"`
	NewApplicationID  *string          `gorm:"type:uuid" json:"new_application_id,omitempty"`
	StartedAt         *time.Time       `json:"started_at,omitempty"`
	CompletedAt       *time.Time       `json:"completed_at,omitempty"`
	ErrorMessage      string           `gorm:"type:text" json:"error_message,omitempty"`
	ValidationResults JSON             `gorm:"type:jsonb" json:"validation_results,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	BackupExecution   *BackupExecution `gorm:"foreignKey:BackupExecutionID" json:"backup_execution,omitempty"`
	Application       *Application     `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
	NewApplication    *Application     `gorm:"foreignKey:NewApplicationID" json:"new_application,omitempty"`
}

// TableName specifies the table name for BackupRestore
func (BackupRestore) TableName() string {
	return "backup_restores"
}

// BeforeCreate hook for BackupRestore
func (b *BackupRestore) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

