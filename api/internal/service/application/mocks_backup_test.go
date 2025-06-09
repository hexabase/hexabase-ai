package application

import (
	"context"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	"github.com/stretchr/testify/mock"
)

// MockApplicationRepository mocks the application repository interface
type MockApplicationRepository struct {
	mock.Mock
}

// CreateApplication creates a new application
func (m *MockApplicationRepository) CreateApplication(ctx context.Context, app *application.Application) error {
	args := m.Called(ctx, app)
	// Set the ID if it's empty (simulating DB behavior)
	if app.ID == "" && app.Metadata != nil && app.Metadata["mock_app_id"] != "" {
		app.ID = app.Metadata["mock_app_id"]
	}
	return args.Error(0)
}

// Create is an alias for CreateApplication
func (m *MockApplicationRepository) Create(ctx context.Context, app *application.Application) error {
	return m.CreateApplication(ctx, app)
}

// GetApplication gets an application by ID
func (m *MockApplicationRepository) GetApplication(ctx context.Context, id string) (*application.Application, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.Application), args.Error(1)
}

// GetByID is an alias for testing
func (m *MockApplicationRepository) GetByID(ctx context.Context, id string) (*application.Application, error) {
	return m.GetApplication(ctx, id)
}

// UpdateApplication updates an application
func (m *MockApplicationRepository) UpdateApplication(ctx context.Context, app *application.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

// Update is an alias for testing
func (m *MockApplicationRepository) Update(ctx context.Context, id string, app *application.Application) error {
	return m.UpdateApplication(ctx, app)
}

// GetApplicationByName gets an application by name
func (m *MockApplicationRepository) GetApplicationByName(ctx context.Context, workspaceID, projectID, name string) (*application.Application, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.Application), args.Error(1)
}

// ListApplications lists applications
func (m *MockApplicationRepository) ListApplications(ctx context.Context, workspaceID, projectID string) ([]application.Application, error) {
	args := m.Called(ctx, workspaceID, projectID)
	return args.Get(0).([]application.Application), args.Error(1)
}

// DeleteApplication deletes an application
func (m *MockApplicationRepository) DeleteApplication(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// CreateEvent creates an event
func (m *MockApplicationRepository) CreateEvent(ctx context.Context, event *application.ApplicationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// ListEvents lists events
func (m *MockApplicationRepository) ListEvents(ctx context.Context, applicationID string, limit int) ([]application.ApplicationEvent, error) {
	args := m.Called(ctx, applicationID, limit)
	return args.Get(0).([]application.ApplicationEvent), args.Error(1)
}

// GetApplicationsByNode gets applications by node
func (m *MockApplicationRepository) GetApplicationsByNode(ctx context.Context, nodeID string) ([]application.Application, error) {
	args := m.Called(ctx, nodeID)
	return args.Get(0).([]application.Application), args.Error(1)
}

// GetApplicationsByStatus gets applications by status
func (m *MockApplicationRepository) GetApplicationsByStatus(ctx context.Context, workspaceID string, status application.ApplicationStatus) ([]application.Application, error) {
	args := m.Called(ctx, workspaceID, status)
	return args.Get(0).([]application.Application), args.Error(1)
}

// GetCronJobExecutions gets cronjob executions
func (m *MockApplicationRepository) GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]application.CronJobExecution, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]application.CronJobExecution), args.Int(1), args.Error(2)
}

// CreateCronJobExecution creates a cronjob execution
func (m *MockApplicationRepository) CreateCronJobExecution(ctx context.Context, execution *application.CronJobExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

// GetCronJobExecution gets a cronjob execution by ID
func (m *MockApplicationRepository) GetCronJobExecution(ctx context.Context, executionID string) (*application.CronJobExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.CronJobExecution), args.Error(1)
}

// UpdateCronJobExecution updates a cronjob execution
func (m *MockApplicationRepository) UpdateCronJobExecution(ctx context.Context, executionID string, completedAt *time.Time, status application.CronJobExecutionStatus, exitCode *int, logs string) error {
	args := m.Called(ctx, executionID, completedAt, status, exitCode, logs)
	return args.Error(0)
}

// UpdateCronSchedule updates cron schedule
func (m *MockApplicationRepository) UpdateCronSchedule(ctx context.Context, applicationID, schedule string) error {
	args := m.Called(ctx, applicationID, schedule)
	return args.Error(0)
}

// GetCronJobExecutionByID gets cronjob execution by ID
func (m *MockApplicationRepository) GetCronJobExecutionByID(ctx context.Context, executionID string) (*application.CronJobExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.CronJobExecution), args.Error(1)
}

// Function-related mock methods (stub implementations)
func (m *MockApplicationRepository) CreateFunctionVersion(ctx context.Context, version *application.FunctionVersion) error {
	return nil
}
func (m *MockApplicationRepository) GetFunctionVersion(ctx context.Context, versionID string) (*application.FunctionVersion, error) {
	return nil, nil
}
func (m *MockApplicationRepository) GetFunctionVersions(ctx context.Context, applicationID string) ([]application.FunctionVersion, error) {
	return nil, nil
}
func (m *MockApplicationRepository) GetActiveFunctionVersion(ctx context.Context, applicationID string) (*application.FunctionVersion, error) {
	return nil, nil
}
func (m *MockApplicationRepository) UpdateFunctionVersion(ctx context.Context, version *application.FunctionVersion) error {
	return nil
}
func (m *MockApplicationRepository) SetActiveFunctionVersion(ctx context.Context, applicationID, versionID string) error {
	return nil
}
func (m *MockApplicationRepository) CreateFunctionInvocation(ctx context.Context, invocation *application.FunctionInvocation) error {
	return nil
}
func (m *MockApplicationRepository) GetFunctionInvocation(ctx context.Context, invocationID string) (*application.FunctionInvocation, error) {
	return nil, nil
}
func (m *MockApplicationRepository) GetFunctionInvocations(ctx context.Context, applicationID string, limit, offset int) ([]application.FunctionInvocation, int, error) {
	return nil, 0, nil
}
func (m *MockApplicationRepository) UpdateFunctionInvocation(ctx context.Context, invocation *application.FunctionInvocation) error {
	return nil
}
func (m *MockApplicationRepository) CreateFunctionEvent(ctx context.Context, event *application.FunctionEvent) error {
	return nil
}
func (m *MockApplicationRepository) GetFunctionEvent(ctx context.Context, eventID string) (*application.FunctionEvent, error) {
	return nil, nil
}
func (m *MockApplicationRepository) GetPendingFunctionEvents(ctx context.Context, applicationID string, limit int) ([]application.FunctionEvent, error) {
	return nil, nil
}
func (m *MockApplicationRepository) UpdateFunctionEvent(ctx context.Context, event *application.FunctionEvent) error {
	return nil
}

// Note: MockKubernetesRepository is defined in mocks_test.go
// We'll extend it with additional methods if needed

// MockBackupService mocks the backup service
type MockBackupService struct {
	mock.Mock
}

func (m *MockBackupService) CreateBackupPolicy(ctx context.Context, applicationID string, req *backup.CreateBackupPolicyRequest) (*backup.BackupPolicy, error) {
	args := m.Called(ctx, applicationID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupPolicy), args.Error(1)
}

func (m *MockBackupService) CreateBackupExecution(ctx context.Context, exec *backup.BackupExecution) (*backup.BackupExecution, error) {
	args := m.Called(ctx, exec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupExecution), args.Error(1)
}

func (m *MockBackupService) GetBackupExecutionByCronJobID(ctx context.Context, cronJobID string) (*backup.BackupExecution, error) {
	args := m.Called(ctx, cronJobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupExecution), args.Error(1)
}

func (m *MockBackupService) UpdateBackupExecutionStatus(ctx context.Context, id string, status backup.BackupExecutionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockProjectRepository mocks the project repository
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id string) (*Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Project), args.Error(1)
}

// MockMonitoringService mocks the monitoring service
type MockMonitoringService struct {
	mock.Mock
}

// Project represents a project (simplified for testing)
type Project struct {
	ID        string
	Namespace string
}