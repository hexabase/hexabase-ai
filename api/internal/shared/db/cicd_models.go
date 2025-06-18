package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Pipeline represents a CI/CD pipeline configuration in the database
type Pipeline struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"not null;index" json:"workspace_id"`
	ProjectID   string    `gorm:"index" json:"project_id"`
	Name        string    `gorm:"not null" json:"name"`
	Provider    string    `gorm:"not null" json:"provider"`
	Config      string    `gorm:"type:text" json:"config"` // JSON serialized PipelineConfig
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`

	// Associations
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"workspace,omitempty"`
	Project   *Project   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Runs      []PipelineRun `gorm:"foreignKey:PipelineID" json:"runs,omitempty"`
}

// BeforeCreate hook to generate UUID
func (p *Pipeline) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PipelineRun represents a pipeline run in the database
type PipelineRun struct {
	ID         string     `gorm:"primaryKey" json:"id"`
	PipelineID string     `gorm:"not null;index" json:"pipeline_id"`
	RunID      string     `gorm:"not null;uniqueIndex" json:"run_id"` // Provider-specific run ID
	Status     string     `gorm:"not null;index" json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Metadata   string     `gorm:"type:text" json:"metadata"` // JSON serialized metadata
	CreatedBy  string     `json:"created_by"`

	// Associations
	Pipeline *Pipeline `gorm:"foreignKey:PipelineID;constraint:OnDelete:CASCADE" json:"pipeline,omitempty"`
}

// BeforeCreate hook to generate UUID
func (pr *PipelineRun) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == "" {
		pr.ID = uuid.New().String()
	}
	return nil
}

// PipelineTemplate represents a reusable pipeline template
type PipelineTemplate struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex" json:"name"`
	Description string    `json:"description"`
	Provider    string    `gorm:"not null;index" json:"provider"`
	Stages      string    `gorm:"type:text" json:"stages"`     // JSON serialized stages
	Parameters  string    `gorm:"type:text" json:"parameters"` // JSON serialized parameters
	IsPublic    bool      `gorm:"default:false" json:"is_public"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate hook to generate UUID
func (pt *PipelineTemplate) BeforeCreate(tx *gorm.DB) error {
	if pt.ID == "" {
		pt.ID = uuid.New().String()
	}
	return nil
}

// WorkspaceProviderConfig represents CI/CD provider configuration for a workspace
type WorkspaceProviderConfig struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"not null;index" json:"workspace_id"`
	Provider    string    `gorm:"not null" json:"provider"`
	Config      string    `gorm:"type:text" json:"config"` // JSON serialized ProviderConfig
	IsActive    bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Associations
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"workspace,omitempty"`
}

// BeforeCreate hook to generate UUID
func (wpc *WorkspaceProviderConfig) BeforeCreate(tx *gorm.DB) error {
	if wpc.ID == "" {
		wpc.ID = uuid.New().String()
	}
	return nil
}

// CICDCredential represents stored CI/CD credentials metadata (actual secrets are in K8s)
type CICDCredential struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"not null;index" json:"workspace_id"`
	Name        string    `gorm:"not null" json:"name"`
	Type        string    `gorm:"not null" json:"type"` // git-ssh, git-token, registry
	SecretRef   string    `gorm:"not null" json:"secret_ref"` // K8s secret name
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`

	// Associations
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"workspace,omitempty"`
}

// BeforeCreate hook to generate UUID
func (cc *CICDCredential) BeforeCreate(tx *gorm.DB) error {
	if cc.ID == "" {
		cc.ID = uuid.New().String()
	}
	return nil
}

// Ensure unique constraint on workspace + name
func init() {
	// This will be used when creating indexes
	_ = &CICDCredential{}
}