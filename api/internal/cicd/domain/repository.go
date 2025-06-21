package domain

import (
	"context"
	"time"
)

// Repository defines the interface for CI/CD data persistence
type Repository interface {
	// Pipeline operations
	CreatePipeline(ctx context.Context, pipeline *Pipeline) error
	GetPipeline(ctx context.Context, id string) (*Pipeline, error)
	GetPipelineByRunID(ctx context.Context, runID string) (*Pipeline, error)
	ListPipelines(ctx context.Context, workspaceID, projectID string, limit, offset int) ([]*Pipeline, error)
	UpdatePipeline(ctx context.Context, pipeline *Pipeline) error
	DeletePipeline(ctx context.Context, id string) error

	// Pipeline run operations
	CreatePipelineRun(ctx context.Context, run *PipelineRunRecord) error
	GetPipelineRun(ctx context.Context, runID string) (*PipelineRunRecord, error)
	UpdatePipelineRun(ctx context.Context, run *PipelineRunRecord) error
	ListPipelineRuns(ctx context.Context, pipelineID string, limit, offset int) ([]*PipelineRunRecord, error)

	// Template operations
	CreateTemplate(ctx context.Context, template *PipelineTemplate) error
	GetTemplate(ctx context.Context, id string) (*PipelineTemplate, error)
	ListTemplates(ctx context.Context, provider string) ([]*PipelineTemplate, error)
	UpdateTemplate(ctx context.Context, template *PipelineTemplate) error
	DeleteTemplate(ctx context.Context, id string) error

	// Provider configuration
	GetProviderConfig(ctx context.Context, workspaceID string) (*WorkspaceProviderConfig, error)
	SetProviderConfig(ctx context.Context, config *WorkspaceProviderConfig) error
}

// Pipeline represents a CI/CD pipeline configuration
type Pipeline struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	Config      string    `json:"config"` // JSON serialized PipelineConfig
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`
}

// PipelineRunRecord represents a pipeline run in the database
type PipelineRunRecord struct {
	ID         string    `json:"id"`
	PipelineID string    `json:"pipeline_id"`
	RunID      string    `json:"run_id"` // Provider-specific run ID
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Metadata   string    `json:"metadata"` // JSON serialized metadata
	CreatedBy  string    `json:"created_by"`
}

// WorkspaceProviderConfig represents provider configuration for a workspace
type WorkspaceProviderConfig struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Provider    string    `json:"provider"`
	Config      string    `json:"config"` // JSON serialized ProviderConfig
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}