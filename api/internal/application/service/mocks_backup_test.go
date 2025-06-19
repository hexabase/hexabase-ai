package service

import (
	"context"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	projectDomain "github.com/hexabase/hexabase-ai/api/internal/project/domain"
	"github.com/stretchr/testify/mock"
)

// MockApplicationRepository mocks the application repository interface
type MockApplicationRepository struct {
	mock.Mock
}

// CreateApplication creates a new application
func (m *MockApplicationRepository) CreateApplication(ctx context.Context, app *domain.Application) error {
	args := m.Called(ctx, app)
	// Set the ID if it's empty (simulating DB behavior)
	if app.ID == "" && app.Metadata != nil && app.Metadata["mock_app_id"] != "" {
		app.ID = app.Metadata["mock_app_id"]
	}
	return args.Error(0)
}

// Create is an alias for CreateApplication
func (m *MockApplicationRepository) Create(ctx context.Context, app *domain.Application) error {
	return m.CreateApplication(ctx, app)
}

// GetApplication gets an application by ID
func (m *MockApplicationRepository) GetApplication(ctx context.Context, id string) (*domain.Application, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Application), args.Error(1)
}

// GetByID is an alias for testing
func (m *MockApplicationRepository) GetByID(ctx context.Context, id string) (*domain.Application, error) {
	return m.GetApplication(ctx, id)
}

// UpdateApplication updates an application
func (m *MockApplicationRepository) UpdateApplication(ctx context.Context, app *domain.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

// Update is an alias for testing
func (m *MockApplicationRepository) Update(ctx context.Context, id string, app *domain.Application) error {
	return m.UpdateApplication(ctx, app)
}

// GetApplicationByName gets an application by name
func (m *MockApplicationRepository) GetApplicationByName(ctx context.Context, workspaceID, projectID, name string) (*domain.Application, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Application), args.Error(1)
}

// ListApplications lists applications
func (m *MockApplicationRepository) ListApplications(ctx context.Context, workspaceID, projectID string) ([]domain.Application, error) {
	args := m.Called(ctx, workspaceID, projectID)
	return args.Get(0).([]domain.Application), args.Error(1)
}

// DeleteApplication deletes an application
func (m *MockApplicationRepository) DeleteApplication(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// CreateEvent creates an event
func (m *MockApplicationRepository) CreateEvent(ctx context.Context, event *domain.ApplicationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// ListEvents lists events
func (m *MockApplicationRepository) ListEvents(ctx context.Context, applicationID string, limit int) ([]domain.ApplicationEvent, error) {
	args := m.Called(ctx, applicationID, limit)
	return args.Get(0).([]domain.ApplicationEvent), args.Error(1)
}

// GetApplicationsByNode gets applications by node
func (m *MockApplicationRepository) GetApplicationsByNode(ctx context.Context, nodeID string) ([]domain.Application, error) {
	args := m.Called(ctx, nodeID)
	return args.Get(0).([]domain.Application), args.Error(1)
}

// GetApplicationsByStatus gets applications by status
func (m *MockApplicationRepository) GetApplicationsByStatus(ctx context.Context, workspaceID string, status domain.ApplicationStatus) ([]domain.Application, error) {
	args := m.Called(ctx, workspaceID, status)
	return args.Get(0).([]domain.Application), args.Error(1)
}

// GetCronJobExecutions gets cronjob executions
func (m *MockApplicationRepository) GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]domain.CronJobExecution, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]domain.CronJobExecution), args.Int(1), args.Error(2)
}

// CreateCronJobExecution creates a cronjob execution
func (m *MockApplicationRepository) CreateCronJobExecution(ctx context.Context, execution *domain.CronJobExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

// GetCronJobExecution gets a cronjob execution by ID
func (m *MockApplicationRepository) GetCronJobExecution(ctx context.Context, executionID string) (*domain.CronJobExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CronJobExecution), args.Error(1)
}

// UpdateCronJobExecution updates a cronjob execution
func (m *MockApplicationRepository) UpdateCronJobExecution(ctx context.Context, executionID string, completedAt *time.Time, status domain.CronJobExecutionStatus, exitCode *int, logs string) error {
	args := m.Called(ctx, executionID, completedAt, status, exitCode, logs)
	return args.Error(0)
}

// UpdateCronSchedule updates cron schedule
func (m *MockApplicationRepository) UpdateCronSchedule(ctx context.Context, applicationID, schedule string) error {
	args := m.Called(ctx, applicationID, schedule)
	return args.Error(0)
}

// GetCronJobExecutionByID gets cronjob execution by ID
func (m *MockApplicationRepository) GetCronJobExecutionByID(ctx context.Context, executionID string) (*domain.CronJobExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CronJobExecution), args.Error(1)
}

// Function-related mock methods (stub implementations)
func (m *MockApplicationRepository) CreateFunctionVersion(ctx context.Context, version *domain.FunctionVersion) error {
	return nil
}
func (m *MockApplicationRepository) GetFunctionVersion(ctx context.Context, versionID string) (*domain.FunctionVersion, error) {
	return nil, nil
}
func (m *MockApplicationRepository) GetFunctionVersions(ctx context.Context, applicationID string) ([]domain.FunctionVersion, error) {
	return nil, nil
}
func (m *MockApplicationRepository) GetActiveFunctionVersion(ctx context.Context, applicationID string) (*domain.FunctionVersion, error) {
	return nil, nil
}
func (m *MockApplicationRepository) UpdateFunctionVersion(ctx context.Context, version *domain.FunctionVersion) error {
	return nil
}
func (m *MockApplicationRepository) SetActiveFunctionVersion(ctx context.Context, applicationID, versionID string) error {
	return nil
}
func (m *MockApplicationRepository) CreateFunctionInvocation(ctx context.Context, invocation *domain.FunctionInvocation) error {
	return nil
}
func (m *MockApplicationRepository) GetFunctionInvocation(ctx context.Context, invocationID string) (*domain.FunctionInvocation, error) {
	return nil, nil
}
func (m *MockApplicationRepository) GetFunctionInvocations(ctx context.Context, applicationID string, limit, offset int) ([]domain.FunctionInvocation, int, error) {
	return nil, 0, nil
}
func (m *MockApplicationRepository) UpdateFunctionInvocation(ctx context.Context, invocation *domain.FunctionInvocation) error {
	return nil
}
func (m *MockApplicationRepository) CreateFunctionEvent(ctx context.Context, event *domain.FunctionEvent) error {
	return nil
}
func (m *MockApplicationRepository) GetFunctionEvent(ctx context.Context, eventID string) (*domain.FunctionEvent, error) {
	return nil, nil
}
func (m *MockApplicationRepository) GetPendingFunctionEvents(ctx context.Context, applicationID string, limit int) ([]domain.FunctionEvent, error) {
	return nil, nil
}
func (m *MockApplicationRepository) UpdateFunctionEvent(ctx context.Context, event *domain.FunctionEvent) error {
	return nil
}

// Note: MockKubernetesRepository is defined in mocks_test.go
// We'll extend it with additional methods if needed

// MockBackupService mocks the backup service
type MockBackupService struct {
	mock.Mock
}

func (m *MockBackupService) CreateBackupPolicy(ctx context.Context, applicationID string, req backup.CreateBackupPolicyRequest) (*backup.BackupPolicy, error) {
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

func (m *MockProjectRepository) GetByID(ctx context.Context, id string) (*projectDomain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.Project), args.Error(1)
}

// MockMonitoringService mocks the monitoring service
type MockMonitoringService struct {
	mock.Mock
}

// Add all missing backup service methods
func (m *MockBackupService) CreateBackupStorage(ctx context.Context, workspaceID string, req backup.CreateBackupStorageRequest) (*backup.BackupStorage, error) {
	args := m.Called(ctx, workspaceID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupStorage), args.Error(1)
}

func (m *MockBackupService) GetBackupStorage(ctx context.Context, storageID string) (*backup.BackupStorage, error) {
	args := m.Called(ctx, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupStorage), args.Error(1)
}

func (m *MockBackupService) ListBackupStorages(ctx context.Context, workspaceID string) ([]backup.BackupStorage, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]backup.BackupStorage), args.Error(1)
}

func (m *MockBackupService) UpdateBackupStorage(ctx context.Context, storageID string, req backup.UpdateBackupStorageRequest) (*backup.BackupStorage, error) {
	args := m.Called(ctx, storageID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupStorage), args.Error(1)
}

func (m *MockBackupService) DeleteBackupStorage(ctx context.Context, storageID string) error {
	args := m.Called(ctx, storageID)
	return args.Error(0)
}

func (m *MockBackupService) GetStorageUsage(ctx context.Context, storageID string) (*backup.BackupStorageUsage, error) {
	args := m.Called(ctx, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupStorageUsage), args.Error(1)
}

func (m *MockBackupService) GetWorkspaceStorageUsage(ctx context.Context, workspaceID string) ([]backup.BackupStorageUsage, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]backup.BackupStorageUsage), args.Error(1)
}

func (m *MockBackupService) GetBackupPolicy(ctx context.Context, policyID string) (*backup.BackupPolicy, error) {
	args := m.Called(ctx, policyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupPolicy), args.Error(1)
}

func (m *MockBackupService) GetBackupPolicyByApplication(ctx context.Context, applicationID string) (*backup.BackupPolicy, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupPolicy), args.Error(1)
}

func (m *MockBackupService) UpdateBackupPolicy(ctx context.Context, policyID string, req backup.UpdateBackupPolicyRequest) (*backup.BackupPolicy, error) {
	args := m.Called(ctx, policyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupPolicy), args.Error(1)
}

func (m *MockBackupService) DeleteBackupPolicy(ctx context.Context, policyID string) error {
	args := m.Called(ctx, policyID)
	return args.Error(0)
}

func (m *MockBackupService) TriggerManualBackup(ctx context.Context, applicationID string, req backup.TriggerBackupRequest) (*backup.BackupExecution, error) {
	args := m.Called(ctx, applicationID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupExecution), args.Error(1)
}

func (m *MockBackupService) GetBackupExecution(ctx context.Context, executionID string) (*backup.BackupExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupExecution), args.Error(1)
}

func (m *MockBackupService) ListBackupExecutions(ctx context.Context, applicationID string, limit, offset int) ([]backup.BackupExecution, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]backup.BackupExecution), args.Int(1), args.Error(2)
}

func (m *MockBackupService) GetLatestBackup(ctx context.Context, applicationID string) (*backup.BackupExecution, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupExecution), args.Error(1)
}

func (m *MockBackupService) ProcessScheduledBackup(ctx context.Context, policyID string) (*backup.BackupExecution, error) {
	args := m.Called(ctx, policyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupExecution), args.Error(1)
}

func (m *MockBackupService) RestoreBackup(ctx context.Context, backupExecutionID string, req backup.RestoreBackupRequest) (*backup.BackupRestore, error) {
	args := m.Called(ctx, backupExecutionID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupRestore), args.Error(1)
}

func (m *MockBackupService) GetBackupRestore(ctx context.Context, restoreID string) (*backup.BackupRestore, error) {
	args := m.Called(ctx, restoreID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupRestore), args.Error(1)
}

func (m *MockBackupService) ListBackupRestores(ctx context.Context, applicationID string, limit, offset int) ([]backup.BackupRestore, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]backup.BackupRestore), args.Int(1), args.Error(2)
}

func (m *MockBackupService) ValidateBackup(ctx context.Context, backupExecutionID string) error {
	args := m.Called(ctx, backupExecutionID)
	return args.Error(0)
}

func (m *MockBackupService) GetBackupManifest(ctx context.Context, backupExecutionID string) (map[string]interface{}, error) {
	args := m.Called(ctx, backupExecutionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockBackupService) DownloadBackup(ctx context.Context, backupExecutionID string) (string, error) {
	args := m.Called(ctx, backupExecutionID)
	return args.String(0), args.Error(1)
}

// Add monitoring service methods
func (m *MockMonitoringService) AcknowledgeAlert(ctx context.Context, alertID string, userID string) error {
	args := m.Called(ctx, alertID, userID)
	return args.Error(0)
}

// Note: Other monitoring service methods are already defined in cronjob_backup_test.go