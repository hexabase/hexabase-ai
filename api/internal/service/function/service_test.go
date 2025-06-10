package function_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"os"

	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
	service "github.com/hexabase/hexabase-ai/api/internal/service/function"
)

// Mock implementations
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) CreateFunction(ctx context.Context, fn *function.FunctionDef) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *mockRepository) UpdateFunction(ctx context.Context, fn *function.FunctionDef) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *mockRepository) DeleteFunction(ctx context.Context, workspaceID, functionID string) error {
	args := m.Called(ctx, workspaceID, functionID)
	return args.Error(0)
}

func (m *mockRepository) GetFunction(ctx context.Context, workspaceID, functionID string) (*function.FunctionDef, error) {
	args := m.Called(ctx, workspaceID, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.FunctionDef), args.Error(1)
}

func (m *mockRepository) ListFunctions(ctx context.Context, workspaceID, projectID string) ([]*function.FunctionDef, error) {
	args := m.Called(ctx, workspaceID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionDef), args.Error(1)
}

func (m *mockRepository) CreateVersion(ctx context.Context, version *function.FunctionVersionDef) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *mockRepository) UpdateVersion(ctx context.Context, version *function.FunctionVersionDef) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *mockRepository) GetVersion(ctx context.Context, workspaceID, functionID, versionID string) (*function.FunctionVersionDef, error) {
	args := m.Called(ctx, workspaceID, functionID, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.FunctionVersionDef), args.Error(1)
}

func (m *mockRepository) ListVersions(ctx context.Context, workspaceID, functionID string) ([]*function.FunctionVersionDef, error) {
	args := m.Called(ctx, workspaceID, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionVersionDef), args.Error(1)
}

func (m *mockRepository) CreateTrigger(ctx context.Context, trigger *function.FunctionTrigger) error {
	args := m.Called(ctx, trigger)
	return args.Error(0)
}

func (m *mockRepository) UpdateTrigger(ctx context.Context, trigger *function.FunctionTrigger) error {
	args := m.Called(ctx, trigger)
	return args.Error(0)
}

func (m *mockRepository) DeleteTrigger(ctx context.Context, workspaceID, functionID, triggerID string) error {
	args := m.Called(ctx, workspaceID, functionID, triggerID)
	return args.Error(0)
}

func (m *mockRepository) ListTriggers(ctx context.Context, workspaceID, functionID string) ([]*function.FunctionTrigger, error) {
	args := m.Called(ctx, workspaceID, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionTrigger), args.Error(1)
}

func (m *mockRepository) CreateInvocation(ctx context.Context, invocation *function.InvocationStatus) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *mockRepository) UpdateInvocation(ctx context.Context, invocation *function.InvocationStatus) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *mockRepository) GetInvocation(ctx context.Context, workspaceID, invocationID string) (*function.InvocationStatus, error) {
	args := m.Called(ctx, workspaceID, invocationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.InvocationStatus), args.Error(1)
}

func (m *mockRepository) ListInvocations(ctx context.Context, workspaceID, functionID string, limit int) ([]*function.InvocationStatus, error) {
	args := m.Called(ctx, workspaceID, functionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.InvocationStatus), args.Error(1)
}

func (m *mockRepository) CreateEvent(ctx context.Context, event *function.FunctionAuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockRepository) ListEvents(ctx context.Context, workspaceID, functionID string, limit int) ([]*function.FunctionAuditEvent, error) {
	args := m.Called(ctx, workspaceID, functionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionAuditEvent), args.Error(1)
}

func (m *mockRepository) GetWorkspaceProviderConfig(ctx context.Context, workspaceID string) (*function.ProviderConfig, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.ProviderConfig), args.Error(1)
}

func (m *mockRepository) UpdateWorkspaceProviderConfig(ctx context.Context, workspaceID string, config *function.ProviderConfig) error {
	args := m.Called(ctx, workspaceID, config)
	return args.Error(0)
}

type mockProviderFactory struct {
	mock.Mock
}

func (m *mockProviderFactory) CreateProvider(ctx context.Context, config function.ProviderConfig) (function.Provider, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(function.Provider), args.Error(1)
}

func (m *mockProviderFactory) GetSupportedProviders() []function.ProviderType {
	args := m.Called()
	return args.Get(0).([]function.ProviderType)
}

type mockProvider struct {
	mock.Mock
}

func (m *mockProvider) CreateFunction(ctx context.Context, spec *function.FunctionSpec) (*function.FunctionDef, error) {
	args := m.Called(ctx, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.FunctionDef), args.Error(1)
}

func (m *mockProvider) UpdateFunction(ctx context.Context, name string, spec *function.FunctionSpec) (*function.FunctionDef, error) {
	args := m.Called(ctx, name, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.FunctionDef), args.Error(1)
}

func (m *mockProvider) DeleteFunction(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockProvider) GetFunction(ctx context.Context, name string) (*function.FunctionDef, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.FunctionDef), args.Error(1)
}

func (m *mockProvider) ListFunctions(ctx context.Context, namespace string) ([]*function.FunctionDef, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionDef), args.Error(1)
}

func (m *mockProvider) CreateVersion(ctx context.Context, functionName string, version *function.FunctionVersionDef) error {
	args := m.Called(ctx, functionName, version)
	return args.Error(0)
}

func (m *mockProvider) GetVersion(ctx context.Context, functionName, versionID string) (*function.FunctionVersionDef, error) {
	args := m.Called(ctx, functionName, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.FunctionVersionDef), args.Error(1)
}

func (m *mockProvider) ListVersions(ctx context.Context, functionName string) ([]*function.FunctionVersionDef, error) {
	args := m.Called(ctx, functionName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionVersionDef), args.Error(1)
}

func (m *mockProvider) SetActiveVersion(ctx context.Context, functionName, versionID string) error {
	args := m.Called(ctx, functionName, versionID)
	return args.Error(0)
}

func (m *mockProvider) CreateTrigger(ctx context.Context, functionName string, trigger *function.FunctionTrigger) error {
	args := m.Called(ctx, functionName, trigger)
	return args.Error(0)
}

func (m *mockProvider) UpdateTrigger(ctx context.Context, functionName, triggerName string, trigger *function.FunctionTrigger) error {
	args := m.Called(ctx, functionName, triggerName, trigger)
	return args.Error(0)
}

func (m *mockProvider) DeleteTrigger(ctx context.Context, functionName, triggerName string) error {
	args := m.Called(ctx, functionName, triggerName)
	return args.Error(0)
}

func (m *mockProvider) ListTriggers(ctx context.Context, functionName string) ([]*function.FunctionTrigger, error) {
	args := m.Called(ctx, functionName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionTrigger), args.Error(1)
}

func (m *mockProvider) InvokeFunction(ctx context.Context, functionName string, request *function.InvokeRequest) (*function.InvokeResponse, error) {
	args := m.Called(ctx, functionName, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.InvokeResponse), args.Error(1)
}

func (m *mockProvider) InvokeFunctionAsync(ctx context.Context, functionName string, request *function.InvokeRequest) (string, error) {
	args := m.Called(ctx, functionName, request)
	return args.String(0), args.Error(1)
}

func (m *mockProvider) GetInvocationStatus(ctx context.Context, invocationID string) (*function.InvocationStatus, error) {
	args := m.Called(ctx, invocationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.InvocationStatus), args.Error(1)
}

func (m *mockProvider) GetFunctionLogs(ctx context.Context, functionName string, opts *function.LogOptions) ([]*function.LogEntry, error) {
	args := m.Called(ctx, functionName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.LogEntry), args.Error(1)
}

func (m *mockProvider) GetFunctionMetrics(ctx context.Context, functionName string, opts *function.MetricOptions) (*function.Metrics, error) {
	args := m.Called(ctx, functionName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.Metrics), args.Error(1)
}

func (m *mockProvider) GetCapabilities() *function.Capabilities {
	args := m.Called()
	return args.Get(0).(*function.Capabilities)
}

func (m *mockProvider) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Tests
func TestService_CreateFunction(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	tests := []struct {
		name          string
		workspaceID   string
		projectID     string
		spec          *function.FunctionSpec
		setupMocks    func(*mockRepository, *mockProviderFactory, *mockProvider)
		expectedError bool
	}{
		{
			name:        "successful creation with Fission provider",
			workspaceID: "ws-123",
			projectID:   "proj-456",
			spec: &function.FunctionSpec{
				Name:       "test-func",
				Runtime:    function.RuntimePython,
				Handler:    "main.handler",
				SourceCode: "def handler(): pass",
			},
			setupMocks: func(repo *mockRepository, factory *mockProviderFactory, provider *mockProvider) {
				// No existing provider config - should default to Fission
				repo.On("GetWorkspaceProviderConfig", ctx, "ws-123").Return(nil, nil)
				
				// Factory creates Fission provider
				factory.On("CreateProvider", ctx, function.ProviderConfig{
					Type: function.ProviderTypeFission,
					Config: map[string]interface{}{
						"endpoint": "http://controller.fission.svc.cluster.local",
					},
				}).Return(provider, nil)
				
				// Provider creates function
				provider.On("CreateFunction", ctx, mock.AnythingOfType("*function.FunctionSpec")).Return(&function.FunctionDef{
					ID:          "func-789",
					Name:        "test-func",
					WorkspaceID: "ws-123",
					ProjectID:   "proj-456",
					Runtime:     function.RuntimePython,
					Handler:     "main.handler",
					Status:      function.FunctionDefStatusReady,
				}, nil)
				
				// Repository stores metadata
				repo.On("CreateFunction", ctx, mock.AnythingOfType("*function.FunctionDef")).Return(nil)
				
				// Repository records event
				repo.On("CreateEvent", ctx, mock.AnythingOfType("*function.FunctionAuditEvent")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "workspace with Knative provider",
			workspaceID: "ws-knative",
			projectID:   "proj-789",
			spec: &function.FunctionSpec{
				Name:    "knative-func",
				Runtime: function.RuntimeNode,
				Handler: "index.handler",
			},
			setupMocks: func(repo *mockRepository, factory *mockProviderFactory, provider *mockProvider) {
				// Workspace configured for Knative
				repo.On("GetWorkspaceProviderConfig", ctx, "ws-knative").Return(&function.ProviderConfig{
					Type: function.ProviderTypeKnative,
					Config: map[string]interface{}{
						"namespace": "knative-serving",
					},
				}, nil)
				
				factory.On("CreateProvider", ctx, function.ProviderConfig{
					Type: function.ProviderTypeKnative,
					Config: map[string]interface{}{
						"namespace": "knative-serving",
					},
				}).Return(provider, nil)
				
				provider.On("CreateFunction", ctx, mock.AnythingOfType("*function.FunctionSpec")).Return(&function.FunctionDef{
					ID:      "func-knative",
					Name:    "knative-func",
					Runtime: function.RuntimeNode,
					Handler: "index.handler",
				}, nil)
				
				repo.On("CreateFunction", ctx, mock.AnythingOfType("*function.FunctionDef")).Return(nil)
				repo.On("CreateEvent", ctx, mock.AnythingOfType("*function.FunctionAuditEvent")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "provider creation fails",
			workspaceID: "ws-fail",
			projectID:   "proj-fail",
			spec: &function.FunctionSpec{
				Name: "fail-func",
			},
			setupMocks: func(repo *mockRepository, factory *mockProviderFactory, provider *mockProvider) {
				repo.On("GetWorkspaceProviderConfig", ctx, "ws-fail").Return(nil, nil)
				factory.On("CreateProvider", ctx, mock.Anything).Return(nil, errors.New("provider unavailable"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			factory := new(mockProviderFactory)
			provider := new(mockProvider)
			
			tt.setupMocks(repo, factory, provider)
			
			svc := service.NewService(repo, factory, logger)
			
			result, err := svc.CreateFunction(ctx, tt.workspaceID, tt.projectID, tt.spec)
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.workspaceID, result.WorkspaceID)
				assert.Equal(t, tt.projectID, result.ProjectID)
			}
			
			repo.AssertExpectations(t)
			factory.AssertExpectations(t)
			provider.AssertExpectations(t)
		})
	}
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
	provider.On("CreateFunction", ctx, mock.AnythingOfType("*function.FunctionSpec")).Return(&function.FunctionDef{
		ID:   "func-1",
		Name: "func-1",
	}, nil).Times(3)
	
	repo.On("CreateFunction", ctx, mock.AnythingOfType("*function.FunctionDef")).Return(nil).Times(3)
	repo.On("CreateEvent", ctx, mock.AnythingOfType("*function.FunctionAuditEvent")).Return(nil).Times(3)
	
	svc := service.NewService(repo, factory, logger)
	
	// Create multiple functions - provider should be cached
	for i := 0; i < 3; i++ {
		spec := &function.FunctionSpec{
			Name:    fmt.Sprintf("func-%d", i),
			Runtime: function.RuntimePython,
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
	
	// Setup versions
	versions := []*function.FunctionVersionDef{
		{ID: "v3", Version: 3, IsActive: true},
		{ID: "v2", Version: 2, IsActive: false},
		{ID: "v1", Version: 1, IsActive: false},
	}
	
	repo.On("ListVersions", ctx, "ws-123", "func-123").Return(versions, nil)
	repo.On("GetFunction", ctx, "ws-123", "func-123").Return(&function.FunctionDef{
		ID:            "func-123",
		Name:          "test-func",
		ActiveVersion: "v3",
	}, nil)
	repo.On("GetVersion", ctx, "ws-123", "func-123", "v2").Return(versions[1], nil)
	
	// Provider setup
	repo.On("GetWorkspaceProviderConfig", ctx, "ws-123").Return(nil, nil)
	factory.On("CreateProvider", ctx, mock.Anything).Return(provider, nil)
	provider.On("SetActiveVersion", ctx, "test-func", "v2").Return(nil)
	
	// Update metadata
	repo.On("UpdateFunction", ctx, mock.AnythingOfType("*function.FunctionDef")).Return(nil)
	repo.On("UpdateVersion", ctx, mock.AnythingOfType("*function.FunctionVersionDef")).Return(nil)
	repo.On("CreateEvent", ctx, mock.AnythingOfType("*function.FunctionAuditEvent")).Return(nil)
	
	svc := service.NewService(repo, factory, logger)
	
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
	
	fissionRepo.On("GetWorkspaceProviderConfig", ctx, "ws-fission").Return(&function.ProviderConfig{
		Type: function.ProviderTypeFission,
		Config: map[string]interface{}{"endpoint": "http://controller.fission"},
	}, nil)
	fissionFactory.On("CreateProvider", ctx, mock.Anything).Return(fissionProvider, nil)
	fissionProvider.On("GetCapabilities").Return(&function.Capabilities{
		TypicalColdStartMs: 100,
	})
	
	fissionSvc := service.NewService(fissionRepo, fissionFactory, logger)
	fissionCaps, _ := fissionSvc.GetProviderCapabilities(ctx, "ws-fission")
	
	// Test Knative provider (slower cold start)
	knativeRepo := new(mockRepository)
	knativeFactory := new(mockProviderFactory)
	knativeProvider := new(mockProvider)
	
	knativeRepo.On("GetWorkspaceProviderConfig", ctx, "ws-knative").Return(&function.ProviderConfig{
		Type: function.ProviderTypeKnative,
	}, nil)
	knativeFactory.On("CreateProvider", ctx, mock.Anything).Return(knativeProvider, nil)
	knativeProvider.On("GetCapabilities").Return(&function.Capabilities{
		TypicalColdStartMs: 2000,
	})
	
	knativeSvc := service.NewService(knativeRepo, knativeFactory, logger)
	knativeCaps, _ := knativeSvc.GetProviderCapabilities(ctx, "ws-knative")
	
	// Assert Fission has significantly faster cold starts
	assert.Less(t, fissionCaps.TypicalColdStartMs, knativeCaps.TypicalColdStartMs)
	assert.Less(t, fissionCaps.TypicalColdStartMs, 500) // Fission should be under 500ms
	assert.Greater(t, knativeCaps.TypicalColdStartMs, 1000) // Knative typically over 1s
}