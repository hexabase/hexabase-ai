package service

import (
	"context"
	"fmt"
	"testing"

	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hexabase/hexabase-ai/api/internal/function/domain"
)

// Mock implementations
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) CreateFunction(ctx context.Context, fn *domain.FunctionDef) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *mockRepository) UpdateFunction(ctx context.Context, fn *domain.FunctionDef) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *mockRepository) DeleteFunction(ctx context.Context, workspaceID, functionID string) error {
	args := m.Called(ctx, workspaceID, functionID)
	return args.Error(0)
}

func (m *mockRepository) GetFunction(ctx context.Context, workspaceID, functionID string) (*domain.FunctionDef, error) {
	args := m.Called(ctx, workspaceID, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionDef), args.Error(1)
}

func (m *mockRepository) ListFunctions(ctx context.Context, workspaceID, projectID string) ([]*domain.FunctionDef, error) {
	args := m.Called(ctx, workspaceID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FunctionDef), args.Error(1)
}

func (m *mockRepository) CreateVersion(ctx context.Context, version *domain.FunctionVersionDef) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *mockRepository) UpdateVersion(ctx context.Context, version *domain.FunctionVersionDef) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *mockRepository) GetVersion(ctx context.Context, workspaceID, functionID, versionID string) (*domain.FunctionVersionDef, error) {
	args := m.Called(ctx, workspaceID, functionID, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionVersionDef), args.Error(1)
}

func (m *mockRepository) ListVersions(ctx context.Context, workspaceID, functionID string) ([]*domain.FunctionVersionDef, error) {
	args := m.Called(ctx, workspaceID, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FunctionVersionDef), args.Error(1)
}

func (m *mockRepository) CreateTrigger(ctx context.Context, trigger *domain.FunctionTrigger) error {
	args := m.Called(ctx, trigger)
	return args.Error(0)
}

func (m *mockRepository) UpdateTrigger(ctx context.Context, trigger *domain.FunctionTrigger) error {
	args := m.Called(ctx, trigger)
	return args.Error(0)
}

func (m *mockRepository) DeleteTrigger(ctx context.Context, workspaceID, functionID, triggerID string) error {
	args := m.Called(ctx, workspaceID, functionID, triggerID)
	return args.Error(0)
}

func (m *mockRepository) ListTriggers(ctx context.Context, workspaceID, functionID string) ([]*domain.FunctionTrigger, error) {
	args := m.Called(ctx, workspaceID, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FunctionTrigger), args.Error(1)
}

func (m *mockRepository) CreateInvocation(ctx context.Context, invocation *domain.InvocationStatus) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *mockRepository) UpdateInvocation(ctx context.Context, invocation *domain.InvocationStatus) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *mockRepository) GetInvocation(ctx context.Context, workspaceID, invocationID string) (*domain.InvocationStatus, error) {
	args := m.Called(ctx, workspaceID, invocationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InvocationStatus), args.Error(1)
}

func (m *mockRepository) ListInvocations(ctx context.Context, workspaceID, functionID string, limit int) ([]*domain.InvocationStatus, error) {
	args := m.Called(ctx, workspaceID, functionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InvocationStatus), args.Error(1)
}

func (m *mockRepository) CreateEvent(ctx context.Context, event *domain.FunctionAuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockRepository) ListEvents(ctx context.Context, workspaceID, functionID string, limit int) ([]*domain.FunctionAuditEvent, error) {
	args := m.Called(ctx, workspaceID, functionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FunctionAuditEvent), args.Error(1)
}

func (m *mockRepository) GetWorkspaceProviderConfig(ctx context.Context, workspaceID string) (*domain.ProviderConfig, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProviderConfig), args.Error(1)
}

func (m *mockRepository) UpdateWorkspaceProviderConfig(ctx context.Context, workspaceID string, config *domain.ProviderConfig) error {
	args := m.Called(ctx, workspaceID, config)
	return args.Error(0)
}

type mockProviderFactory struct {
	mock.Mock
}

func (m *mockProviderFactory) CreateProvider(ctx context.Context, config domain.ProviderConfig) (domain.Provider, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.Provider), args.Error(1)
}

func (m *mockProviderFactory) GetSupportedProviders() []domain.ProviderType {
	args := m.Called()
	return args.Get(0).([]domain.ProviderType)
}

type mockProvider struct {
	mock.Mock
}

func (m *mockProvider) CreateFunction(ctx context.Context, spec *domain.FunctionSpec) (*domain.FunctionDef, error) {
	args := m.Called(ctx, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionDef), args.Error(1)
}

func (m *mockProvider) UpdateFunction(ctx context.Context, name string, spec *domain.FunctionSpec) (*domain.FunctionDef, error) {
	args := m.Called(ctx, name, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionDef), args.Error(1)
}

func (m *mockProvider) DeleteFunction(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockProvider) GetFunction(ctx context.Context, name string) (*domain.FunctionDef, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionDef), args.Error(1)
}

func (m *mockProvider) ListFunctions(ctx context.Context, namespace string) ([]*domain.FunctionDef, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FunctionDef), args.Error(1)
}

func (m *mockProvider) CreateVersion(ctx context.Context, functionName string, version *domain.FunctionVersionDef) error {
	args := m.Called(ctx, functionName, version)
	return args.Error(0)
}

func (m *mockProvider) GetVersion(ctx context.Context, functionName, versionID string) (*domain.FunctionVersionDef, error) {
	args := m.Called(ctx, functionName, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionVersionDef), args.Error(1)
}

func (m *mockProvider) ListVersions(ctx context.Context, functionName string) ([]*domain.FunctionVersionDef, error) {
	args := m.Called(ctx, functionName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FunctionVersionDef), args.Error(1)
}

func (m *mockProvider) SetActiveVersion(ctx context.Context, functionName, versionID string) error {
	args := m.Called(ctx, functionName, versionID)
	return args.Error(0)
}

func (m *mockProvider) CreateTrigger(ctx context.Context, functionName string, trigger *domain.FunctionTrigger) error {
	args := m.Called(ctx, functionName, trigger)
	return args.Error(0)
}

func (m *mockProvider) UpdateTrigger(ctx context.Context, functionName, triggerName string, trigger *domain.FunctionTrigger) error {
	args := m.Called(ctx, functionName, triggerName, trigger)
	return args.Error(0)
}

func (m *mockProvider) DeleteTrigger(ctx context.Context, functionName, triggerName string) error {
	args := m.Called(ctx, functionName, triggerName)
	return args.Error(0)
}

func (m *mockProvider) ListTriggers(ctx context.Context, functionName string) ([]*domain.FunctionTrigger, error) {
	args := m.Called(ctx, functionName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FunctionTrigger), args.Error(1)
}

func (m *mockProvider) InvokeFunction(ctx context.Context, functionName string, request *domain.InvokeRequest) (*domain.InvokeResponse, error) {
	args := m.Called(ctx, functionName, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InvokeResponse), args.Error(1)
}

func (m *mockProvider) InvokeFunctionAsync(ctx context.Context, functionName string, request *domain.InvokeRequest) (string, error) {
	args := m.Called(ctx, functionName, request)
	return args.String(0), args.Error(1)
}

func (m *mockProvider) GetInvocationStatus(ctx context.Context, invocationID string) (*domain.InvocationStatus, error) {
	args := m.Called(ctx, invocationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InvocationStatus), args.Error(1)
}

func (m *mockProvider) GetFunctionURL(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}

func (m *mockProvider) GetFunctionLogs(ctx context.Context, functionName string, opts *domain.LogOptions) ([]*domain.LogEntry, error) {
	args := m.Called(ctx, functionName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LogEntry), args.Error(1)
}

func (m *mockProvider) GetFunctionMetrics(ctx context.Context, functionName string, opts *domain.MetricOptions) (*domain.Metrics, error) {
	args := m.Called(ctx, functionName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Metrics), args.Error(1)
}

func (m *mockProvider) GetCapabilities() *domain.Capabilities {
	args := m.Called()
	return args.Get(0).(*domain.Capabilities)
}

func (m *mockProvider) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Tests
func TestService_CreateFunction(t *testing.T) {
	mockRepo := &mockRepository{}
	mockProviderFactory := &mockProviderFactory{}
	mockProv := &mockProvider{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(mockRepo, mockProviderFactory, logger)

	ctx := context.Background()
	workspaceID := "test-workspace"
	projectID := "test-project"

	spec := &domain.FunctionSpec{
		Name:       "test-function",
		Namespace:  "test-ns",
		Runtime:    domain.RuntimePython,
		Handler:    "main.handler",
		SourceCode: "def handler(): pass",
	}

	// Mock provider config
	providerConfig := &domain.ProviderConfig{
		Type: domain.ProviderTypeMock,
		Config: map[string]interface{}{
			"endpoint": "http://mock",
		},
	}

	// Mock repository GetWorkspaceProviderConfig call
	mockRepo.On("GetWorkspaceProviderConfig", ctx, workspaceID).Return(providerConfig, nil)

	// Mock provider creation
	mockProviderFactory.On("CreateProvider", ctx, *providerConfig).Return(mockProv, nil)

	// Mock function creation in provider
	expectedFn := &domain.FunctionDef{
		Name:      spec.Name,
		Namespace: spec.Namespace,
		Runtime:   spec.Runtime,
		Handler:   spec.Handler,
		Status:    domain.FunctionDefStatusReady,
	}
	mockProv.On("CreateFunction", ctx, spec).Return(expectedFn, nil)

	// Mock repository save
	mockRepo.On("CreateFunction", ctx, mock.AnythingOfType("*domain.FunctionDef")).Return(nil)

	// Mock audit event creation
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.FunctionAuditEvent")).Return(nil)

	// Test
	result, err := service.CreateFunction(ctx, workspaceID, projectID, spec)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, spec.Name, result.Name)
	assert.Equal(t, spec.Runtime, result.Runtime)

	mockRepo.AssertExpectations(t)
	mockProviderFactory.AssertExpectations(t)
	mockProv.AssertExpectations(t)
}

func TestService_ProviderCaching(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	repo := new(mockRepository)
	factory := new(mockProviderFactory)
	provider := new(mockProvider)
	
	// Setup: provider config and factory should only be called once
	repo.On("GetWorkspaceProviderConfig", ctx, "ws-123").Return(nil, nil).Once()
	factory.On("CreateProvider", ctx, mock.Anything).Return(provider, nil).Once()
	
	// Multiple function operations
	provider.On("CreateFunction", ctx, mock.AnythingOfType("*domain.FunctionSpec")).Return(&domain.FunctionDef{
		ID:   "func-1",
		Name: "func-1",
	}, nil).Times(3)
	
	repo.On("CreateFunction", ctx, mock.AnythingOfType("*domain.FunctionDef")).Return(nil).Times(3)
	repo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.FunctionAuditEvent")).Return(nil).Times(3)
	
	svc := NewService(repo, factory, logger)
	
	// Create multiple functions - provider should be cached
	for i := 0; i < 3; i++ {
		spec := &domain.FunctionSpec{
			Name:    fmt.Sprintf("func-%d", i),
			Runtime: domain.RuntimePython,
			Handler: "main.handler",
		}
		_, err := svc.CreateFunction(ctx, "ws-123", "proj-456", spec)
		assert.NoError(t, err)
	}
	
	// Verify provider was only created once
	repo.AssertExpectations(t)
	factory.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestService_RollbackVersion(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	repo := new(mockRepository)
	factory := new(mockProviderFactory)
	provider := new(mockProvider)
	
	// Setup versions - ordered with v2 first, then v3 (active)
	// RollbackVersion looks for the previous version in the array
	versions := []*domain.FunctionVersionDef{
		{ID: "v2", Version: 2, IsActive: false},
		{ID: "v3", Version: 3, IsActive: true},
		{ID: "v1", Version: 1, IsActive: false},
	}
	
	// RollbackVersion calls ListVersions to find previous version
	repo.On("ListVersions", ctx, "ws-123", "func-123").Return(versions, nil)
	
	// Then it calls SetActiveVersion which needs these mocks:
	repo.On("GetFunction", ctx, "ws-123", "func-123").Return(&domain.FunctionDef{
		ID:            "func-123",
		Name:          "test-func",
		ActiveVersion: "v3",
	}, nil)
	repo.On("GetVersion", ctx, "ws-123", "func-123", "v2").Return(versions[0], nil)
	
	// Provider setup for SetActiveVersion
	repo.On("GetWorkspaceProviderConfig", ctx, "ws-123").Return(&domain.ProviderConfig{
		Type: domain.ProviderTypeMock,
	}, nil)
	factory.On("CreateProvider", ctx, mock.Anything).Return(provider, nil)
	
	// SetActiveVersion will be called with the new version ID (v2)
	provider.On("SetActiveVersion", ctx, "test-func", "v2").Return(nil)
	
	// Update metadata in SetActiveVersion
	repo.On("UpdateFunction", ctx, mock.AnythingOfType("*domain.FunctionDef")).Return(nil)
	repo.On("UpdateVersion", ctx, mock.AnythingOfType("*domain.FunctionVersionDef")).Return(nil)
	repo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.FunctionAuditEvent")).Return(nil)
	
	svc := NewService(repo, factory, logger)
	
	err := svc.RollbackVersion(ctx, "ws-123", "func-123")
	assert.NoError(t, err)
	
	repo.AssertExpectations(t)
	factory.AssertExpectations(t)
	provider.AssertExpectations(t)
}

// Test cold start performance difference
func TestService_ColdStartPerformance(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	// Test Fission provider (fast cold start)
	fissionRepo := new(mockRepository)
	fissionFactory := new(mockProviderFactory)
	fissionProvider := new(mockProvider)
	
	fissionRepo.On("GetWorkspaceProviderConfig", ctx, "ws-fission").Return(&domain.ProviderConfig{
		Type: domain.ProviderTypeFission,
		Config: map[string]interface{}{"endpoint": "http://controller.fission"},
	}, nil)
	fissionFactory.On("CreateProvider", ctx, mock.Anything).Return(fissionProvider, nil)
	fissionProvider.On("GetCapabilities").Return(&domain.Capabilities{
		TypicalColdStartMs: 100,
	})
	
	fissionSvc := NewService(fissionRepo, fissionFactory, logger)
	fissionCaps, _ := fissionSvc.GetProviderCapabilities(ctx, "ws-fission")
	
	// Test Knative provider (slower cold start)
	knativeRepo := new(mockRepository)
	knativeFactory := new(mockProviderFactory)
	knativeProvider := new(mockProvider)
	
	knativeRepo.On("GetWorkspaceProviderConfig", ctx, "ws-knative").Return(&domain.ProviderConfig{
		Type: domain.ProviderTypeKnative,
	}, nil)
	knativeFactory.On("CreateProvider", ctx, mock.Anything).Return(knativeProvider, nil)
	knativeProvider.On("GetCapabilities").Return(&domain.Capabilities{
		TypicalColdStartMs: 2000,
	})
	
	knativeSvc := NewService(knativeRepo, knativeFactory, logger)
	knativeCaps, _ := knativeSvc.GetProviderCapabilities(ctx, "ws-knative")
	
	// Assert Fission has significantly faster cold starts
	assert.Less(t, fissionCaps.TypicalColdStartMs, knativeCaps.TypicalColdStartMs)
	assert.Less(t, fissionCaps.TypicalColdStartMs, 500) // Fission should be under 500ms
	assert.Greater(t, knativeCaps.TypicalColdStartMs, 1000) // Knative typically over 1s
}