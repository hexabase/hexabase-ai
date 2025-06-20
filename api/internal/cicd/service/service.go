package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/cicd/domain"
)

// Service implements the CI/CD service interface
type Service struct {
	repo              domain.Repository
	providerFactory   domain.ProviderFactory
	credentialManager domain.CredentialManager
	providers         map[string]domain.Provider
	logger            *slog.Logger
}

// NewService creates a new CI/CD service
func NewService(
	repo domain.Repository,
	providerFactory domain.ProviderFactory,
	credentialManager domain.CredentialManager,
	logger *slog.Logger,
) domain.Service {
	return &Service{
		repo:              repo,
		providerFactory:   providerFactory,
		credentialManager: credentialManager,
		providers:         make(map[string]domain.Provider),
		logger:            logger,
	}
}

// CreatePipeline creates a new pipeline
func (s *Service) CreatePipeline(ctx context.Context, workspaceID string, config domain.PipelineConfig) (*domain.PipelineRun, error) {
	// Set workspace ID
	config.WorkspaceID = workspaceID

	// Get provider for workspace
	provider, err := s.getProviderForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Validate configuration
	if err := provider.ValidateConfig(ctx, &config); err != nil {
		return nil, fmt.Errorf("invalid pipeline configuration: %w", err)
	}

	// Serialize config
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize config: %w", err)
	}

	// Create pipeline record
	pipeline := &domain.Pipeline{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		ProjectID:   config.ProjectID,
		Name:        config.Name,
		Provider:    provider.GetName(),
		Config:      string(configJSON),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   getUserID(ctx),
	}

	if err := s.repo.CreatePipeline(ctx, pipeline); err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	// Run the pipeline
	run, err := provider.RunPipeline(ctx, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to run pipeline: %w", err)
	}

	// Create run record
	runRecord := &domain.PipelineRunRecord{
		ID:         uuid.New().String(),
		PipelineID: pipeline.ID,
		RunID:      run.ID,
		Status:     string(run.Status),
		StartedAt:  run.StartedAt,
		CreatedBy:  getUserID(ctx),
	}

	if err := s.repo.CreatePipelineRun(ctx, runRecord); err != nil {
		s.logger.Error("failed to create pipeline run record", "error", err)
	}

	return run, nil
}

// GetPipeline retrieves a pipeline by ID
func (s *Service) GetPipeline(ctx context.Context, pipelineID string) (*domain.PipelineRun, error) {
	// Get pipeline run record
	runRecord, err := s.repo.GetPipelineRun(ctx, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	// Get pipeline
	pipeline, err := s.repo.GetPipeline(ctx, runRecord.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline configuration not found: %w", err)
	}

	// Get provider
	provider, err := s.getProviderForWorkspace(ctx, pipeline.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Get current status from provider
	return provider.GetStatus(ctx, pipeline.WorkspaceID, runRecord.RunID)
}

// ListPipelines lists pipelines for a workspace/project
func (s *Service) ListPipelines(ctx context.Context, workspaceID, projectID string, limit int) ([]*domain.PipelineRun, error) {
	// Get provider
	provider, err := s.getProviderForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// List from provider
	return provider.ListPipelines(ctx, workspaceID, projectID)
}

// CancelPipeline cancels a running pipeline
func (s *Service) CancelPipeline(ctx context.Context, pipelineID string) error {
	// Get pipeline run record
	runRecord, err := s.repo.GetPipelineRun(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("pipeline not found: %w", err)
	}

	// Get pipeline
	pipeline, err := s.repo.GetPipeline(ctx, runRecord.PipelineID)
	if err != nil {
		return fmt.Errorf("pipeline configuration not found: %w", err)
	}

	// Get provider
	provider, err := s.getProviderForWorkspace(ctx, pipeline.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Cancel the pipeline
	if err := provider.CancelPipeline(ctx, pipeline.WorkspaceID, runRecord.RunID); err != nil {
		return fmt.Errorf("failed to cancel pipeline: %w", err)
	}

	// Update run record
	runRecord.Status = string(domain.PipelineStatusCancelled)
	now := time.Now()
	runRecord.FinishedAt = &now
	
	return s.repo.UpdatePipelineRun(ctx, runRecord)
}

// DeletePipeline deletes a pipeline
func (s *Service) DeletePipeline(ctx context.Context, pipelineID string) error {
	// Get pipeline run record
	runRecord, err := s.repo.GetPipelineRun(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("pipeline not found: %w", err)
	}

	// Get pipeline
	pipeline, err := s.repo.GetPipeline(ctx, runRecord.PipelineID)
	if err != nil {
		return fmt.Errorf("pipeline configuration not found: %w", err)
	}

	// Get provider
	provider, err := s.getProviderForWorkspace(ctx, pipeline.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Delete from provider
	if err := provider.DeletePipeline(ctx, pipeline.WorkspaceID, runRecord.RunID); err != nil {
		s.logger.Error("failed to delete pipeline from provider", "error", err)
	}

	// Delete from database
	return s.repo.DeletePipeline(ctx, pipeline.ID)
}

// RetryPipeline retries a failed pipeline
func (s *Service) RetryPipeline(ctx context.Context, pipelineID string) (*domain.PipelineRun, error) {
	// Get pipeline run record
	runRecord, err := s.repo.GetPipelineRun(ctx, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	// Get pipeline
	pipeline, err := s.repo.GetPipeline(ctx, runRecord.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline configuration not found: %w", err)
	}

	// Deserialize config
	var config domain.PipelineConfig
	if err := json.Unmarshal([]byte(pipeline.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to deserialize config: %w", err)
	}

	// Create new pipeline run
	return s.CreatePipeline(ctx, pipeline.WorkspaceID, config)
}

// GetPipelineLogs retrieves logs for a pipeline
func (s *Service) GetPipelineLogs(ctx context.Context, pipelineID string, stage, task string) ([]domain.LogEntry, error) {
	// Get pipeline run record
	runRecord, err := s.repo.GetPipelineRun(ctx, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	// Get pipeline
	pipeline, err := s.repo.GetPipeline(ctx, runRecord.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline configuration not found: %w", err)
	}

	// Get provider
	provider, err := s.getProviderForWorkspace(ctx, pipeline.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Get logs
	return provider.GetLogs(ctx, pipeline.WorkspaceID, runRecord.RunID)
}

// StreamPipelineLogs streams logs for a pipeline
func (s *Service) StreamPipelineLogs(ctx context.Context, pipelineID string, stage, task string) (io.ReadCloser, error) {
	// Get pipeline run record
	runRecord, err := s.repo.GetPipelineRun(ctx, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	// Get pipeline
	pipeline, err := s.repo.GetPipeline(ctx, runRecord.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline configuration not found: %w", err)
	}

	// Get provider
	provider, err := s.getProviderForWorkspace(ctx, pipeline.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Stream logs
	return provider.StreamLogs(ctx, pipeline.WorkspaceID, runRecord.RunID)
}

// ListTemplates lists available pipeline templates
func (s *Service) ListTemplates(ctx context.Context, provider string) ([]*domain.PipelineTemplate, error) {
	// List from repository
	templates, err := s.repo.ListTemplates(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	// If provider is specified, also get provider-specific templates
	if provider != "" {
		p, err := s.getProvider(ctx, provider)
		if err == nil {
			providerTemplates, err := p.GetTemplates(ctx)
			if err == nil {
				templates = append(templates, providerTemplates...)
			}
		}
	}

	return templates, nil
}

// GetTemplate retrieves a template by ID
func (s *Service) GetTemplate(ctx context.Context, templateID string) (*domain.PipelineTemplate, error) {
	return s.repo.GetTemplate(ctx, templateID)
}

// CreatePipelineFromTemplate creates a pipeline from a template
func (s *Service) CreatePipelineFromTemplate(ctx context.Context, workspaceID, templateID string, params map[string]any) (*domain.PipelineRun, error) {
	// Get template
	template, err := s.repo.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, template.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Create config from template
	config, err := provider.CreateFromTemplate(ctx, templateID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create config from template: %w", err)
	}

	// Create pipeline
	return s.CreatePipeline(ctx, workspaceID, *config)
}

// CreateGitCredential stores Git credentials
func (s *Service) CreateGitCredential(ctx context.Context, workspaceID, name string, credential domain.GitCredential) error {
	// Store in Kubernetes
	credInfo, err := s.credentialManager.StoreGitCredential(workspaceID, &credential)
	if err != nil {
		return fmt.Errorf("failed to store git credential: %w", err)
	}

	// Store metadata in database (credInfo contains the metadata)
	_ = credInfo // metadata is already stored by credential manager
	
	return nil
}

// CreateRegistryCredential stores container registry credentials
func (s *Service) CreateRegistryCredential(ctx context.Context, workspaceID, name string, credential domain.RegistryCredential) error {
	// Store in Kubernetes
	credInfo, err := s.credentialManager.StoreRegistryCredential(workspaceID, &credential)
	if err != nil {
		return fmt.Errorf("failed to store registry credential: %w", err)
	}

	// Store metadata in database (credInfo contains the metadata)
	_ = credInfo // metadata is already stored by credential manager

	return nil
}

// ListCredentials lists available credentials
func (s *Service) ListCredentials(ctx context.Context, workspaceID string) ([]domain.CredentialInfo, error) {
	creds, err := s.credentialManager.ListCredentials(workspaceID)
	if err != nil {
		return nil, err
	}
	
	// Convert []*CredentialInfo to []CredentialInfo
	result := make([]domain.CredentialInfo, len(creds))
	for i, cred := range creds {
		result[i] = *cred
	}
	return result, nil
}

// DeleteCredential deletes a credential
func (s *Service) DeleteCredential(ctx context.Context, workspaceID, name string) error {
	return s.credentialManager.DeleteCredential(workspaceID, name)
}

// ListProviders lists available CI/CD providers
func (s *Service) ListProviders(ctx context.Context) ([]domain.ProviderInfo, error) {
	providers := s.providerFactory.ListProviders()
	
	infos := make([]domain.ProviderInfo, len(providers))
	for i, name := range providers {
		infos[i] = domain.ProviderInfo{
			Name:        name,
			DisplayName: s.getProviderDisplayName(name),
			Description: s.getProviderDescription(name),
			Features:    s.getProviderFeatures(name),
			Status:      s.getProviderStatus(name),
		}
	}
	
	return infos, nil
}

// GetProviderConfig retrieves provider configuration for a workspace
func (s *Service) GetProviderConfig(ctx context.Context, workspaceID string) (*domain.ProviderConfig, error) {
	config, err := s.repo.GetProviderConfig(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("provider config not found: %w", err)
	}

	// Deserialize config
	var providerConfig domain.ProviderConfig
	if err := json.Unmarshal([]byte(config.Config), &providerConfig); err != nil {
		return nil, fmt.Errorf("failed to deserialize config: %w", err)
	}

	return &providerConfig, nil
}

// SetProviderConfig sets provider configuration for a workspace
func (s *Service) SetProviderConfig(ctx context.Context, workspaceID string, config domain.ProviderConfig) error {
	// Validate provider type
	if !s.isValidProvider(config.Type) {
		return fmt.Errorf("invalid provider type: %s", config.Type)
	}

	// Serialize config
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Save to database
	workspaceConfig := &domain.WorkspaceProviderConfig{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Provider:    config.Type,
		Config:      string(configJSON),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return s.repo.SetProviderConfig(ctx, workspaceConfig)
}

// Helper methods

func (s *Service) getProviderForWorkspace(ctx context.Context, workspaceID string) (domain.Provider, error) {
	// Get provider config
	config, err := s.GetProviderConfig(ctx, workspaceID)
	if err != nil {
		// Default to Tekton if no config
		config = &domain.ProviderConfig{
			Type:     "tekton",
			Settings: map[string]any{},
		}
	}

	return s.getProvider(ctx, config.Type)
}

func (s *Service) getProvider(ctx context.Context, providerType string) (domain.Provider, error) {
	// Check cache
	if provider, ok := s.providers[providerType]; ok {
		return provider, nil
	}

	// Create provider
	provider, err := s.providerFactory.CreateProvider(providerType, &domain.ProviderConfig{
		Type:     providerType,
		Settings: map[string]any{},
	})
	if err != nil {
		return nil, err
	}

	// Cache provider
	s.providers[providerType] = provider

	return provider, nil
}

func (s *Service) isValidProvider(providerType string) bool {
	providers := s.providerFactory.ListProviders()
	for _, p := range providers {
		if p == providerType {
			return true
		}
	}
	return false
}

func (s *Service) getProviderDisplayName(name string) string {
	switch name {
	case "tekton":
		return "Tekton Pipelines"
	case "github-actions":
		return "GitHub Actions"
	case "gitlab-ci":
		return "GitLab CI"
	default:
		return name
	}
}

func (s *Service) getProviderDescription(name string) string {
	switch name {
	case "tekton":
		return "Cloud-native CI/CD pipelines running on Kubernetes"
	case "github-actions":
		return "GitHub's built-in CI/CD automation"
	case "gitlab-ci":
		return "GitLab's integrated CI/CD pipelines"
	default:
		return ""
	}
}

func (s *Service) getProviderFeatures(name string) []string {
	switch name {
	case "tekton":
		return []string{"kubernetes-native", "extensible", "reusable-tasks", "parallel-execution"}
	case "github-actions":
		return []string{"github-integration", "marketplace", "matrix-builds", "secrets-management"}
	case "gitlab-ci":
		return []string{"gitlab-integration", "auto-devops", "container-registry", "security-scanning"}
	default:
		return []string{}
	}
}

func (s *Service) getProviderStatus(name string) string {
	switch name {
	case "tekton":
		return "available"
	case "github-actions":
		return "beta"
	case "gitlab-ci":
		return "beta"
	default:
		return "deprecated"
	}
}

func getUserID(ctx context.Context) string {
	// Extract user ID from context
	// This is a placeholder - implement based on your auth system
	return "system"
}

// CICDCredential is a placeholder - should be imported from db package
type CICDCredential struct {
	ID          string
	WorkspaceID string
	Name        string
	Type        string
	SecretRef   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
}