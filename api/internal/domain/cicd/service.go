package cicd

import (
	"context"
	"io"
)

// Service defines the business logic interface for CI/CD operations
type Service interface {
	// Pipeline operations
	CreatePipeline(ctx context.Context, workspaceID string, config PipelineConfig) (*PipelineRun, error)
	GetPipeline(ctx context.Context, pipelineID string) (*PipelineRun, error)
	ListPipelines(ctx context.Context, workspaceID, projectID string, limit int) ([]*PipelineRun, error)
	CancelPipeline(ctx context.Context, pipelineID string) error
	DeletePipeline(ctx context.Context, pipelineID string) error
	RetryPipeline(ctx context.Context, pipelineID string) (*PipelineRun, error)

	// Log operations
	GetPipelineLogs(ctx context.Context, pipelineID string, stage, task string) ([]LogEntry, error)
	StreamPipelineLogs(ctx context.Context, pipelineID string, stage, task string) (io.ReadCloser, error)

	// Template operations
	ListTemplates(ctx context.Context, provider string) ([]*PipelineTemplate, error)
	GetTemplate(ctx context.Context, templateID string) (*PipelineTemplate, error)
	CreatePipelineFromTemplate(ctx context.Context, workspaceID, templateID string, params map[string]any) (*PipelineRun, error)

	// Credential operations
	CreateGitCredential(ctx context.Context, workspaceID, name string, credential GitCredential) error
	CreateRegistryCredential(ctx context.Context, workspaceID, name string, credential RegistryCredential) error
	ListCredentials(ctx context.Context, workspaceID string) ([]CredentialInfo, error)
	DeleteCredential(ctx context.Context, workspaceID, name string) error

	// Provider operations
	ListProviders(ctx context.Context) ([]ProviderInfo, error)
	GetProviderConfig(ctx context.Context, workspaceID string) (*ProviderConfig, error)
	SetProviderConfig(ctx context.Context, workspaceID string, config ProviderConfig) error
}

// ProviderInfo represents information about an available CI/CD provider
type ProviderInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
	Status      string   `json:"status"` // "available", "beta", "deprecated"
}