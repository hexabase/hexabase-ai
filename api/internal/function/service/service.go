package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/function/domain"
)

// Service implements the domain.Service interface
type Service struct {
	repo           domain.Repository
	providerFactory domain.ProviderFactory
	providers      map[string]domain.Provider // cache of initialized providers per workspace
	logger         *slog.Logger
}

// NewService creates a new function service instance
func NewService(
	repo domain.Repository,
	providerFactory domain.ProviderFactory,
	logger *slog.Logger,
) *Service {
	return &Service{
		repo:            repo,
		providerFactory: providerFactory,
		providers:       make(map[string]domain.Provider),
		logger:          logger,
	}
}

// getProvider returns the provider for a workspace, initializing it if needed
func (s *Service) getProvider(ctx context.Context, workspaceID string) (domain.Provider, error) {
	// Check cache first
	if provider, exists := s.providers[workspaceID]; exists {
		return provider, nil
	}

	// Get workspace provider configuration
	config, err := s.repo.GetWorkspaceProviderConfig(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace provider config: %w", err)
	}

	// Default to Fission if no config exists
	if config == nil {
		config = &domain.ProviderConfig{
			Type: domain.ProviderTypeFission,
			Config: map[string]interface{}{
				"endpoint": "http://controller.fission.svc.cluster.local",
			},
		}
	}

	// Create provider instance
	provider, err := s.providerFactory.CreateProvider(ctx, *config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Cache the provider
	s.providers[workspaceID] = provider
	return provider, nil
}

// CreateFunction creates a new function
func (s *Service) CreateFunction(ctx context.Context, workspaceID, projectID string, spec *domain.FunctionSpec) (*domain.FunctionDef, error) {
	s.logger.Info("creating function",
		"workspaceID", workspaceID,
		"projectID", projectID,
		"name", spec.Name)

	// Get provider for workspace
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Set namespace to project ID
	spec.Namespace = projectID

	// Create function in provider
	fn, err := provider.CreateFunction(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("provider failed to create function: %w", err)
	}

	// Store function metadata
	fn.WorkspaceID = workspaceID
	fn.ProjectID = projectID
	if err := s.repo.CreateFunction(ctx, fn); err != nil {
		// Try to clean up from provider
		_ = provider.DeleteFunction(ctx, spec.Name)
		return nil, fmt.Errorf("failed to store function metadata: %w", err)
	}

	// Record event
	event := &domain.FunctionAuditEvent{
		ID:          fmt.Sprintf("evt-%d", time.Now().Unix()),
		WorkspaceID: workspaceID,
		FunctionID:  fn.ID,
		Type:        "created",
		Description: fmt.Sprintf("Function %s created", fn.Name),
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	_ = s.repo.CreateEvent(ctx, event)

	return fn, nil
}

// UpdateFunction updates an existing function
func (s *Service) UpdateFunction(ctx context.Context, workspaceID, functionID string, spec *domain.FunctionSpec) (*domain.FunctionDef, error) {
	s.logger.Info("updating function",
		"workspaceID", workspaceID,
		"functionID", functionID)

	// Get existing function
	existing, err := s.repo.GetFunction(ctx, workspaceID, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Update in provider
	spec.Namespace = existing.ProjectID
	updated, err := provider.UpdateFunction(ctx, existing.Name, spec)
	if err != nil {
		return nil, fmt.Errorf("provider failed to update function: %w", err)
	}

	// Update metadata
	updated.ID = functionID
	updated.WorkspaceID = workspaceID
	updated.ProjectID = existing.ProjectID
	if err := s.repo.UpdateFunction(ctx, updated); err != nil {
		return nil, fmt.Errorf("failed to update function metadata: %w", err)
	}

	// Record event
	event := &domain.FunctionAuditEvent{
		ID:          fmt.Sprintf("evt-%d", time.Now().Unix()),
		WorkspaceID: workspaceID,
		FunctionID:  functionID,
		Type:        "updated",
		Description: fmt.Sprintf("Function %s updated", updated.Name),
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	_ = s.repo.CreateEvent(ctx, event)

	return updated, nil
}

// DeleteFunction deletes a function
func (s *Service) DeleteFunction(ctx context.Context, workspaceID, functionID string) error {
	s.logger.Info("deleting function",
		"workspaceID", workspaceID,
		"functionID", functionID)

	// Get function
	fn, err := s.repo.GetFunction(ctx, workspaceID, functionID)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Delete from provider
	if err := provider.DeleteFunction(ctx, fn.Name); err != nil {
		return fmt.Errorf("provider failed to delete function: %w", err)
	}

	// Delete metadata
	if err := s.repo.DeleteFunction(ctx, workspaceID, functionID); err != nil {
		return fmt.Errorf("failed to delete function metadata: %w", err)
	}

	// Record event
	event := &domain.FunctionAuditEvent{
		ID:          fmt.Sprintf("evt-%d", time.Now().Unix()),
		WorkspaceID: workspaceID,
		FunctionID:  functionID,
		Type:        "deleted",
		Description: fmt.Sprintf("Function %s deleted", fn.Name),
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	_ = s.repo.CreateEvent(ctx, event)

	return nil
}

// GetFunction retrieves a function by ID
func (s *Service) GetFunction(ctx context.Context, workspaceID, functionID string) (*domain.FunctionDef, error) {
	return s.repo.GetFunction(ctx, workspaceID, functionID)
}

// ListFunctions lists all functions in a project
func (s *Service) ListFunctions(ctx context.Context, workspaceID, projectID string) ([]*domain.FunctionDef, error) {
	return s.repo.ListFunctions(ctx, workspaceID, projectID)
}

// DeployVersion deploys a new version of a function
func (s *Service) DeployVersion(ctx context.Context, workspaceID, functionID string, version *domain.FunctionVersionDef) (*domain.FunctionVersionDef, error) {
	s.logger.Info("deploying function version",
		"workspaceID", workspaceID,
		"functionID", functionID)

	// Get function
	fn, err := s.repo.GetFunction(ctx, workspaceID, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Create version in provider
	version.FunctionName = fn.Name
	if err := provider.CreateVersion(ctx, fn.Name, version); err != nil {
		return nil, fmt.Errorf("provider failed to create version: %w", err)
	}

	// Store version metadata
	version.WorkspaceID = workspaceID
	version.FunctionID = functionID
	if err := s.repo.CreateVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to store version metadata: %w", err)
	}

	// Record event
	event := &domain.FunctionAuditEvent{
		ID:          fmt.Sprintf("evt-%d", time.Now().Unix()),
		WorkspaceID: workspaceID,
		FunctionID:  functionID,
		Type:        "deployed",
		Description: fmt.Sprintf("Version %s deployed", version.ID),
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	_ = s.repo.CreateEvent(ctx, event)

	return version, nil
}

// GetVersion retrieves a specific version
func (s *Service) GetVersion(ctx context.Context, workspaceID, functionID, versionID string) (*domain.FunctionVersionDef, error) {
	return s.repo.GetVersion(ctx, workspaceID, functionID, versionID)
}

// ListVersions lists all versions of a function
func (s *Service) ListVersions(ctx context.Context, workspaceID, functionID string) ([]*domain.FunctionVersionDef, error) {
	return s.repo.ListVersions(ctx, workspaceID, functionID)
}

// SetActiveVersion sets the active version of a function
func (s *Service) SetActiveVersion(ctx context.Context, workspaceID, functionID, versionID string) error {
	s.logger.Info("setting active version",
		"workspaceID", workspaceID,
		"functionID", functionID,
		"versionID", versionID)

	// Get function and version
	fn, err := s.repo.GetFunction(ctx, workspaceID, functionID)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}

	version, err := s.repo.GetVersion(ctx, workspaceID, functionID, versionID)
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Set active version in provider
	if err := provider.SetActiveVersion(ctx, fn.Name, version.ID); err != nil {
		return fmt.Errorf("provider failed to set active version: %w", err)
	}

	// Update function metadata
	fn.ActiveVersion = versionID
	fn.UpdatedAt = time.Now()
	if err := s.repo.UpdateFunction(ctx, fn); err != nil {
		return fmt.Errorf("failed to update function metadata: %w", err)
	}

	// Update version metadata
	version.IsActive = true
	if err := s.repo.UpdateVersion(ctx, version); err != nil {
		return fmt.Errorf("failed to update version metadata: %w", err)
	}

	// Record event
	event := &domain.FunctionAuditEvent{
		ID:          fmt.Sprintf("evt-%d", time.Now().Unix()),
		WorkspaceID: workspaceID,
		FunctionID:  functionID,
		Type:        "version_activated",
		Description: fmt.Sprintf("Version %s activated", versionID),
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	_ = s.repo.CreateEvent(ctx, event)

	return nil
}

// RollbackVersion rolls back to the previous version
func (s *Service) RollbackVersion(ctx context.Context, workspaceID, functionID string) error {
	// Get versions ordered by creation time
	versions, err := s.repo.ListVersions(ctx, workspaceID, functionID)
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	if len(versions) < 2 {
		return fmt.Errorf("no previous version to rollback to")
	}

	// Find current active and previous version
	var previousVersion *domain.FunctionVersionDef
	for i, v := range versions {
		if v.IsActive {
			if i > 0 {
				previousVersion = versions[i-1]
			}
			break
		}
	}

	if previousVersion == nil {
		return fmt.Errorf("no previous version found")
	}

	// Set previous version as active
	return s.SetActiveVersion(ctx, workspaceID, functionID, previousVersion.ID)
}

// CreateTrigger creates a new trigger for a function
func (s *Service) CreateTrigger(ctx context.Context, workspaceID, functionID string, trigger *domain.FunctionTrigger) (*domain.FunctionTrigger, error) {
	s.logger.Info("creating trigger",
		"workspaceID", workspaceID,
		"functionID", functionID,
		"triggerName", trigger.Name)

	// Get function
	fn, err := s.repo.GetFunction(ctx, workspaceID, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Create trigger in provider
	trigger.FunctionName = fn.Name
	if err := provider.CreateTrigger(ctx, fn.Name, trigger); err != nil {
		return nil, fmt.Errorf("provider failed to create trigger: %w", err)
	}

	// Store trigger metadata
	trigger.ID = fmt.Sprintf("trg-%s-%d", trigger.Name, time.Now().Unix())
	trigger.WorkspaceID = workspaceID
	trigger.FunctionID = functionID
	if err := s.repo.CreateTrigger(ctx, trigger); err != nil {
		// Try to clean up from provider
		_ = provider.DeleteTrigger(ctx, fn.Name, trigger.Name)
		return nil, fmt.Errorf("failed to store trigger metadata: %w", err)
	}

	return trigger, nil
}

// UpdateTrigger updates an existing trigger
func (s *Service) UpdateTrigger(ctx context.Context, workspaceID, functionID, triggerID string, trigger *domain.FunctionTrigger) (*domain.FunctionTrigger, error) {
	// Implementation similar to CreateTrigger but with update logic
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// DeleteTrigger deletes a trigger
func (s *Service) DeleteTrigger(ctx context.Context, workspaceID, functionID, triggerID string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

// ListTriggers lists all triggers for a function
func (s *Service) ListTriggers(ctx context.Context, workspaceID, functionID string) ([]*domain.FunctionTrigger, error) {
	return s.repo.ListTriggers(ctx, workspaceID, functionID)
}

// InvokeFunction invokes a function synchronously
func (s *Service) InvokeFunction(ctx context.Context, workspaceID, functionID string, request *domain.InvokeRequest) (*domain.InvokeResponse, error) {
	s.logger.Info("invoking function",
		"workspaceID", workspaceID,
		"functionID", functionID)

	// Get function
	fn, err := s.repo.GetFunction(ctx, workspaceID, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Invoke function
	response, err := provider.InvokeFunction(ctx, fn.Name, request)
	if err != nil {
		return nil, fmt.Errorf("provider failed to invoke function: %w", err)
	}

	// Record invocation
	invocation := &domain.InvocationStatus{
		InvocationID: response.InvocationID,
		WorkspaceID:  workspaceID,
		FunctionID:   functionID,
		Status:       "completed",
		StartedAt:    time.Now().Add(-response.Duration),
		CompletedAt:  &time.Time{},
		Result:       response,
	}
	*invocation.CompletedAt = time.Now()
	_ = s.repo.CreateInvocation(ctx, invocation)

	// Record event
	event := &domain.FunctionAuditEvent{
		ID:          fmt.Sprintf("evt-%d", time.Now().Unix()),
		WorkspaceID: workspaceID,
		FunctionID:  functionID,
		Type:        "invoked",
		Description: fmt.Sprintf("Function invoked with status %d", response.StatusCode),
		Metadata: map[string]string{
			"invocationID": response.InvocationID,
			"duration":     response.Duration.String(),
		},
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	_ = s.repo.CreateEvent(ctx, event)

	return response, nil
}

// InvokeFunctionAsync invokes a function asynchronously
func (s *Service) InvokeFunctionAsync(ctx context.Context, workspaceID, functionID string, request *domain.InvokeRequest) (string, error) {
	s.logger.Info("invoking function async",
		"workspaceID", workspaceID,
		"functionID", functionID)

	// Get function
	fn, err := s.repo.GetFunction(ctx, workspaceID, functionID)
	if err != nil {
		return "", fmt.Errorf("failed to get function: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return "", err
	}

	// Invoke function async
	invocationID, err := provider.InvokeFunctionAsync(ctx, fn.Name, request)
	if err != nil {
		return "", fmt.Errorf("provider failed to invoke function async: %w", err)
	}

	// Record invocation
	invocation := &domain.InvocationStatus{
		InvocationID: invocationID,
		WorkspaceID:  workspaceID,
		FunctionID:   functionID,
		Status:       "running",
		StartedAt:    time.Now(),
	}
	_ = s.repo.CreateInvocation(ctx, invocation)

	return invocationID, nil
}

// GetInvocationStatus gets the status of an async invocation
func (s *Service) GetInvocationStatus(ctx context.Context, workspaceID, invocationID string) (*domain.InvocationStatus, error) {
	// Get from repository first
	invocation, err := s.repo.GetInvocation(ctx, workspaceID, invocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invocation: %w", err)
	}

	// If still running, check with provider
	if invocation.Status == "running" {
		provider, err := s.getProvider(ctx, workspaceID)
		if err != nil {
			return nil, err
		}

		status, err := provider.GetInvocationStatus(ctx, invocationID)
		if err != nil {
			return nil, fmt.Errorf("provider failed to get invocation status: %w", err)
		}

		// Update if status changed
		if status.Status != invocation.Status {
			invocation.Status = status.Status
			invocation.CompletedAt = status.CompletedAt
			invocation.Result = status.Result
			invocation.Error = status.Error
			_ = s.repo.UpdateInvocation(ctx, invocation)
		}
	}

	return invocation, nil
}

// ListInvocations lists invocation history for a function
func (s *Service) ListInvocations(ctx context.Context, workspaceID, functionID string, limit int) ([]*domain.InvocationStatus, error) {
	return s.repo.ListInvocations(ctx, workspaceID, functionID, limit)
}

// GetFunctionLogs retrieves logs for a function
func (s *Service) GetFunctionLogs(ctx context.Context, workspaceID, functionID string, opts *domain.LogOptions) ([]*domain.LogEntry, error) {
	// Get function
	fn, err := s.repo.GetFunction(ctx, workspaceID, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Get logs from provider
	return provider.GetFunctionLogs(ctx, fn.Name, opts)
}

// GetFunctionMetrics retrieves metrics for a function
func (s *Service) GetFunctionMetrics(ctx context.Context, workspaceID, functionID string, opts *domain.MetricOptions) (*domain.Metrics, error) {
	// Get function
	fn, err := s.repo.GetFunction(ctx, workspaceID, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	// Get provider
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Get metrics from provider
	return provider.GetFunctionMetrics(ctx, fn.Name, opts)
}

// GetFunctionEvents retrieves events for a function
func (s *Service) GetFunctionEvents(ctx context.Context, workspaceID, functionID string, limit int) ([]*domain.FunctionAuditEvent, error) {
	return s.repo.ListEvents(ctx, workspaceID, functionID, limit)
}

// GetProviderCapabilities returns the capabilities of the workspace's provider
func (s *Service) GetProviderCapabilities(ctx context.Context, workspaceID string) (*domain.Capabilities, error) {
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return provider.GetCapabilities(), nil
}

// GetProviderHealth checks the health of the workspace's provider
func (s *Service) GetProviderHealth(ctx context.Context, workspaceID string) error {
	provider, err := s.getProvider(ctx, workspaceID)
	if err != nil {
		return err
	}

	return provider.HealthCheck(ctx)
}