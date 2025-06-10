package function_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
	"github.com/hexabase/hexabase-ai/api/internal/logging"
	funcmock "github.com/hexabase/hexabase-ai/api/internal/repository/function/mock"
	funcservice "github.com/hexabase/hexabase-ai/api/internal/service/function"
)

// MockRepository is a mock implementation of function.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateFunction(ctx context.Context, fn *function.FunctionDef) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *MockRepository) UpdateFunction(ctx context.Context, fn *function.FunctionDef) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *MockRepository) DeleteFunction(ctx context.Context, workspaceID, functionID string) error {
	args := m.Called(ctx, workspaceID, functionID)
	return args.Error(0)
}

func (m *MockRepository) GetFunction(ctx context.Context, workspaceID, functionID string) (*function.FunctionDef, error) {
	args := m.Called(ctx, workspaceID, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.FunctionDef), args.Error(1)
}

func (m *MockRepository) ListFunctions(ctx context.Context, workspaceID, projectID string) ([]*function.FunctionDef, error) {
	args := m.Called(ctx, workspaceID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionDef), args.Error(1)
}

func (m *MockRepository) CreateVersion(ctx context.Context, version *function.FunctionVersionDef) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockRepository) UpdateVersion(ctx context.Context, version *function.FunctionVersionDef) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockRepository) GetVersion(ctx context.Context, workspaceID, functionID, versionID string) (*function.FunctionVersionDef, error) {
	args := m.Called(ctx, workspaceID, functionID, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.FunctionVersionDef), args.Error(1)
}

func (m *MockRepository) ListVersions(ctx context.Context, workspaceID, functionID string) ([]*function.FunctionVersionDef, error) {
	args := m.Called(ctx, workspaceID, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionVersionDef), args.Error(1)
}

func (m *MockRepository) CreateTrigger(ctx context.Context, trigger *function.FunctionTrigger) error {
	args := m.Called(ctx, trigger)
	return args.Error(0)
}

func (m *MockRepository) UpdateTrigger(ctx context.Context, trigger *function.FunctionTrigger) error {
	args := m.Called(ctx, trigger)
	return args.Error(0)
}

func (m *MockRepository) DeleteTrigger(ctx context.Context, workspaceID, functionID, triggerID string) error {
	args := m.Called(ctx, workspaceID, functionID, triggerID)
	return args.Error(0)
}

func (m *MockRepository) ListTriggers(ctx context.Context, workspaceID, functionID string) ([]*function.FunctionTrigger, error) {
	args := m.Called(ctx, workspaceID, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionTrigger), args.Error(1)
}

func (m *MockRepository) CreateInvocation(ctx context.Context, invocation *function.InvocationStatus) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *MockRepository) UpdateInvocation(ctx context.Context, invocation *function.InvocationStatus) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *MockRepository) GetInvocation(ctx context.Context, workspaceID, invocationID string) (*function.InvocationStatus, error) {
	args := m.Called(ctx, workspaceID, invocationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.InvocationStatus), args.Error(1)
}

func (m *MockRepository) ListInvocations(ctx context.Context, workspaceID, functionID string, limit int) ([]*function.InvocationStatus, error) {
	args := m.Called(ctx, workspaceID, functionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.InvocationStatus), args.Error(1)
}

func (m *MockRepository) CreateEvent(ctx context.Context, event *function.FunctionAuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockRepository) ListEvents(ctx context.Context, workspaceID, functionID string, limit int) ([]*function.FunctionAuditEvent, error) {
	args := m.Called(ctx, workspaceID, functionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*function.FunctionAuditEvent), args.Error(1)
}

func (m *MockRepository) GetWorkspaceProviderConfig(ctx context.Context, workspaceID string) (*function.ProviderConfig, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.ProviderConfig), args.Error(1)
}

func (m *MockRepository) UpdateWorkspaceProviderConfig(ctx context.Context, workspaceID string, config *function.ProviderConfig) error {
	args := m.Called(ctx, workspaceID, config)
	return args.Error(0)
}

// MockProviderFactory is a mock implementation of function.ProviderFactory
type MockProviderFactory struct {
	mock.Mock
}

func (m *MockProviderFactory) CreateProvider(ctx context.Context, providerType function.ProviderType, config map[string]interface{}) (function.Provider, error) {
	args := m.Called(ctx, providerType, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(function.Provider), args.Error(1)
}

func (m *MockProviderFactory) GetAvailableProviders() []function.ProviderType {
	args := m.Called()
	return args.Get(0).([]function.ProviderType)
}

func (m *MockProviderFactory) ValidateProviderConfig(providerType function.ProviderType, config map[string]interface{}) error {
	args := m.Called(providerType, config)
	return args.Error(0)
}

func (m *MockProviderFactory) GetProviderCapabilities(providerType function.ProviderType) (*function.Capabilities, error) {
	args := m.Called(providerType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*function.Capabilities), args.Error(1)
}

func TestService_CreateFunction(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockFactory := new(MockProviderFactory)
	mockProvider := funcmock.NewFunctionProvider()
	logger := logging.NewLogger()

	service := funcservice.NewService(mockRepo, mockFactory, logger)

	workspaceID := "ws-123"
	projectID := "proj-456"
	spec := &function.FunctionSpec{
		Name:       "test-function",
		Runtime:    function.RuntimePython,
		Handler:    "main.handler",
		SourceCode: "def handler(): pass",
	}

	// Setup expectations
	mockRepo.On("GetWorkspaceProviderConfig", ctx, workspaceID).Return(nil, nil).Once()
	mockFactory.On("CreateProvider", ctx, function.ProviderTypeFission, mock.Anything).Return(mockProvider, nil).Once()
	mockRepo.On("CreateFunction", ctx, mock.AnythingOfType("*function.FunctionDef")).Return(nil).Once()
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*function.FunctionAuditEvent")).Return(nil).Once()

	// Execute
	result, err := service.CreateFunction(ctx, workspaceID, projectID, spec)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, spec.Name, result.Name)
	assert.Equal(t, projectID, result.Namespace)
	assert.Equal(t, workspaceID, result.WorkspaceID)
	assert.Equal(t, projectID, result.ProjectID)
	assert.NotEmpty(t, result.ID)
	assert.NotZero(t, result.CreatedAt)

	mockRepo.AssertExpectations(t)
	mockFactory.AssertExpectations(t)
}

func TestService_InvokeFunction(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockFactory := new(MockProviderFactory)
	mockProvider := funcmock.NewFunctionProvider()
	logger := logging.NewLogger()

	service := funcservice.NewService(mockRepo, mockFactory, logger)

	workspaceID := "ws-123"
	functionID := "func-789"
	functionName := "test-function"

	// Create function in mock provider first
	spec := &function.FunctionSpec{
		Name:       functionName,
		Namespace:  "test-ns",
		Runtime:    function.RuntimePython,
		Handler:    "main.handler",
		SourceCode: "def handler(): pass",
	}
	_, err := mockProvider.CreateFunction(ctx, spec)
	require.NoError(t, err)

	// Setup expectations
	fn := &function.FunctionDef{
		ID:          functionID,
		Name:        functionName,
		WorkspaceID: workspaceID,
	}
	mockRepo.On("GetFunction", ctx, workspaceID, functionID).Return(fn, nil).Once()
	mockRepo.On("GetWorkspaceProviderConfig", ctx, workspaceID).Return(nil, nil).Once()
	mockFactory.On("CreateProvider", ctx, function.ProviderTypeFission, mock.Anything).Return(mockProvider, nil).Once()
	mockRepo.On("CreateInvocation", ctx, mock.AnythingOfType("*function.InvocationStatus")).Return(nil).Once()
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*function.FunctionAuditEvent")).Return(nil).Once()

	// Execute
	request := &function.InvokeRequest{
		Method: "POST",
		Path:   "/test",
		Body:   []byte(`{"test": true}`),
	}
	response, err := service.InvokeFunction(ctx, workspaceID, functionID, request)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)
	assert.NotEmpty(t, response.InvocationID)
	assert.Greater(t, response.Duration, time.Duration(0))

	mockRepo.AssertExpectations(t)
	mockFactory.AssertExpectations(t)
}

func TestService_SetActiveVersion(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockFactory := new(MockProviderFactory)
	mockProvider := funcmock.NewFunctionProvider()
	logger := logging.NewLogger()

	service := funcservice.NewService(mockRepo, mockFactory, logger)

	workspaceID := "ws-123"
	functionID := "func-789"
	functionName := "test-function"
	versionID := "v2"

	// Create function and version in mock provider
	spec := &function.FunctionSpec{
		Name:       functionName,
		Namespace:  "test-ns",
		Runtime:    function.RuntimePython,
		Handler:    "main.handler",
		SourceCode: "def handler(): pass",
	}
	fn, err := mockProvider.CreateFunction(ctx, spec)
	require.NoError(t, err)

	version := &function.FunctionVersionDef{
		FunctionName: functionName,
		SourceCode:   "def handler(): return 'v2'",
	}
	err = mockProvider.CreateVersion(ctx, functionName, version)
	require.NoError(t, err)

	// Get the created version ID
	versions, err := mockProvider.ListVersions(ctx, functionName)
	require.NoError(t, err)
	require.Len(t, versions, 2)
	actualVersionID := versions[1].ID

	// Setup expectations
	fnDef := &function.FunctionDef{
		ID:            functionID,
		Name:          functionName,
		WorkspaceID:   workspaceID,
		ActiveVersion: fn.ActiveVersion,
	}
	versionDef := &function.FunctionVersionDef{
		ID:          actualVersionID,
		WorkspaceID: workspaceID,
		FunctionID:  functionID,
		Version:     2,
	}

	mockRepo.On("GetFunction", ctx, workspaceID, functionID).Return(fnDef, nil).Once()
	mockRepo.On("GetVersion", ctx, workspaceID, functionID, versionID).Return(versionDef, nil).Once()
	mockRepo.On("GetWorkspaceProviderConfig", ctx, workspaceID).Return(nil, nil).Once()
	mockFactory.On("CreateProvider", ctx, function.ProviderTypeFission, mock.Anything).Return(mockProvider, nil).Once()
	mockRepo.On("UpdateFunction", ctx, mock.AnythingOfType("*function.FunctionDef")).Return(nil).Once()
	mockRepo.On("UpdateVersion", ctx, mock.AnythingOfType("*function.FunctionVersionDef")).Return(nil).Once()
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*function.FunctionAuditEvent")).Return(nil).Once()

	// Execute
	err = service.SetActiveVersion(ctx, workspaceID, functionID, versionID)

	// Assert
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockFactory.AssertExpectations(t)
}