package service

import (
	"context"
	"testing"
	"time"

	applicationDomain "github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/hexabase/hexabase-ai/api/internal/backup/domain"
	workspaceDomain "github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/kubernetes/fake"
)

// Mock implementations

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateBackupStorage(ctx context.Context, storage *domain.BackupStorage) error {
	args := m.Called(ctx, storage)
	return args.Error(0)
}

func (m *MockRepository) GetBackupStorage(ctx context.Context, storageID string) (*domain.BackupStorage, error) {
	args := m.Called(ctx, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupStorage), args.Error(1)
}

func (m *MockRepository) GetBackupStorageByName(ctx context.Context, workspaceID, name string) (*domain.BackupStorage, error) {
	args := m.Called(ctx, workspaceID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupStorage), args.Error(1)
}

func (m *MockRepository) ListBackupStorages(ctx context.Context, workspaceID string) ([]domain.BackupStorage, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]domain.BackupStorage), args.Error(1)
}

func (m *MockRepository) UpdateBackupStorage(ctx context.Context, storage *domain.BackupStorage) error {
	args := m.Called(ctx, storage)
	return args.Error(0)
}

func (m *MockRepository) DeleteBackupStorage(ctx context.Context, storageID string) error {
	args := m.Called(ctx, storageID)
	return args.Error(0)
}

func (m *MockRepository) UpdateStorageUsage(ctx context.Context, storageID string, usedGB int) error {
	args := m.Called(ctx, storageID, usedGB)
	return args.Error(0)
}

func (m *MockRepository) CreateBackupPolicy(ctx context.Context, policy *domain.BackupPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockRepository) GetBackupPolicy(ctx context.Context, policyID string) (*domain.BackupPolicy, error) {
	args := m.Called(ctx, policyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupPolicy), args.Error(1)
}

func (m *MockRepository) GetBackupPolicyByApplication(ctx context.Context, applicationID string) (*domain.BackupPolicy, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupPolicy), args.Error(1)
}

func (m *MockRepository) ListBackupPolicies(ctx context.Context, workspaceID string) ([]domain.BackupPolicy, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]domain.BackupPolicy), args.Error(1)
}

func (m *MockRepository) UpdateBackupPolicy(ctx context.Context, policy *domain.BackupPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockRepository) DeleteBackupPolicy(ctx context.Context, policyID string) error {
	args := m.Called(ctx, policyID)
	return args.Error(0)
}

func (m *MockRepository) ListEnabledPolicies(ctx context.Context) ([]domain.BackupPolicy, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.BackupPolicy), args.Error(1)
}

func (m *MockRepository) CreateBackupExecution(ctx context.Context, execution *domain.BackupExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockRepository) GetBackupExecution(ctx context.Context, executionID string) (*domain.BackupExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupExecution), args.Error(1)
}

func (m *MockRepository) ListBackupExecutions(ctx context.Context, policyID string, limit, offset int) ([]domain.BackupExecution, int, error) {
	args := m.Called(ctx, policyID, limit, offset)
	return args.Get(0).([]domain.BackupExecution), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateBackupExecution(ctx context.Context, execution *domain.BackupExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockRepository) GetLatestBackupExecution(ctx context.Context, policyID string) (*domain.BackupExecution, error) {
	args := m.Called(ctx, policyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupExecution), args.Error(1)
}

func (m *MockRepository) GetBackupExecutionsByApplication(ctx context.Context, applicationID string, limit, offset int) ([]domain.BackupExecution, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]domain.BackupExecution), args.Int(1), args.Error(2)
}

func (m *MockRepository) CleanupOldBackups(ctx context.Context, policyID string, retentionDays int) error {
	args := m.Called(ctx, policyID, retentionDays)
	return args.Error(0)
}

func (m *MockRepository) CreateBackupRestore(ctx context.Context, restore *domain.BackupRestore) error {
	args := m.Called(ctx, restore)
	return args.Error(0)
}

func (m *MockRepository) GetBackupRestore(ctx context.Context, restoreID string) (*domain.BackupRestore, error) {
	args := m.Called(ctx, restoreID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupRestore), args.Error(1)
}

func (m *MockRepository) ListBackupRestores(ctx context.Context, applicationID string, limit, offset int) ([]domain.BackupRestore, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]domain.BackupRestore), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateBackupRestore(ctx context.Context, restore *domain.BackupRestore) error {
	args := m.Called(ctx, restore)
	return args.Error(0)
}

func (m *MockRepository) GetStorageUsage(ctx context.Context, storageID string) (*domain.BackupStorageUsage, error) {
	args := m.Called(ctx, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupStorageUsage), args.Error(1)
}

func (m *MockRepository) GetWorkspaceStorageUsage(ctx context.Context, workspaceID string) ([]domain.BackupStorageUsage, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]domain.BackupStorageUsage), args.Error(1)
}

type MockProxmoxRepository struct {
	mock.Mock
}

func (m *MockProxmoxRepository) CreateStorage(ctx context.Context, nodeID string, config domain.ProxmoxStorageConfig) (string, error) {
	args := m.Called(ctx, nodeID, config)
	return args.String(0), args.Error(1)
}

func (m *MockProxmoxRepository) DeleteStorage(ctx context.Context, nodeID, storageID string) error {
	args := m.Called(ctx, nodeID, storageID)
	return args.Error(0)
}

func (m *MockProxmoxRepository) GetStorageInfo(ctx context.Context, nodeID, storageID string) (*domain.ProxmoxStorageInfo, error) {
	args := m.Called(ctx, nodeID, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProxmoxStorageInfo), args.Error(1)
}

func (m *MockProxmoxRepository) ResizeStorage(ctx context.Context, nodeID, storageID string, newSizeGB int) error {
	args := m.Called(ctx, nodeID, storageID, newSizeGB)
	return args.Error(0)
}

func (m *MockProxmoxRepository) CreateBackupVolume(ctx context.Context, nodeID, storageID string, volumeSize int64) (string, error) {
	args := m.Called(ctx, nodeID, storageID, volumeSize)
	return args.String(0), args.Error(1)
}

func (m *MockProxmoxRepository) DeleteBackupVolume(ctx context.Context, nodeID, storageID, volumeID string) error {
	args := m.Called(ctx, nodeID, storageID, volumeID)
	return args.Error(0)
}

func (m *MockProxmoxRepository) GetVolumeInfo(ctx context.Context, nodeID, storageID, volumeID string) (*domain.ProxmoxVolumeInfo, error) {
	args := m.Called(ctx, nodeID, storageID, volumeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProxmoxVolumeInfo), args.Error(1)
}

func (m *MockProxmoxRepository) SetStorageQuota(ctx context.Context, nodeID, storageID string, quotaGB int) error {
	args := m.Called(ctx, nodeID, storageID, quotaGB)
	return args.Error(0)
}

func (m *MockProxmoxRepository) GetStorageQuota(ctx context.Context, nodeID, storageID string) (*domain.ProxmoxStorageQuota, error) {
	args := m.Called(ctx, nodeID, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProxmoxStorageQuota), args.Error(1)
}

type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) GetApplication(ctx context.Context, applicationID string) (*applicationDomain.Application, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*applicationDomain.Application), args.Error(1)
}

// Implement all required methods for applicationDomain.Repository interface
func (m *MockApplicationRepository) CreateApplication(ctx context.Context, app *applicationDomain.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetApplicationByName(ctx context.Context, workspaceID, projectID, name string) (*applicationDomain.Application, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*applicationDomain.Application), args.Error(1)
}

func (m *MockApplicationRepository) ListApplications(ctx context.Context, workspaceID, projectID string) ([]applicationDomain.Application, error) {
	args := m.Called(ctx, workspaceID, projectID)
	return args.Get(0).([]applicationDomain.Application), args.Error(1)
}

func (m *MockApplicationRepository) UpdateApplication(ctx context.Context, app *applicationDomain.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) DeleteApplication(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockApplicationRepository) CreateEvent(ctx context.Context, event *applicationDomain.ApplicationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockApplicationRepository) ListEvents(ctx context.Context, applicationID string, limit int) ([]applicationDomain.ApplicationEvent, error) {
	args := m.Called(ctx, applicationID, limit)
	return args.Get(0).([]applicationDomain.ApplicationEvent), args.Error(1)
}

func (m *MockApplicationRepository) GetApplicationsByNode(ctx context.Context, nodeID string) ([]applicationDomain.Application, error) {
	args := m.Called(ctx, nodeID)
	return args.Get(0).([]applicationDomain.Application), args.Error(1)
}

func (m *MockApplicationRepository) GetApplicationsByStatus(ctx context.Context, workspaceID string, status applicationDomain.ApplicationStatus) ([]applicationDomain.Application, error) {
	args := m.Called(ctx, workspaceID, status)
	return args.Get(0).([]applicationDomain.Application), args.Error(1)
}

func (m *MockApplicationRepository) Create(ctx context.Context, app *applicationDomain.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]applicationDomain.CronJobExecution, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]applicationDomain.CronJobExecution), args.Int(1), args.Error(2)
}

func (m *MockApplicationRepository) CreateCronJobExecution(ctx context.Context, execution *applicationDomain.CronJobExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockApplicationRepository) UpdateCronJobExecution(ctx context.Context, executionID string, completedAt *time.Time, status applicationDomain.CronJobExecutionStatus, exitCode *int, logs string) error {
	args := m.Called(ctx, executionID, completedAt, status, exitCode, logs)
	return args.Error(0)
}

func (m *MockApplicationRepository) UpdateCronSchedule(ctx context.Context, applicationID, schedule string) error {
	args := m.Called(ctx, applicationID, schedule)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetCronJobExecution(ctx context.Context, executionID string) (*applicationDomain.CronJobExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*applicationDomain.CronJobExecution), args.Error(1)
}

func (m *MockApplicationRepository) GetCronJobExecutionByID(ctx context.Context, executionID string) (*applicationDomain.CronJobExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*applicationDomain.CronJobExecution), args.Error(1)
}

func (m *MockApplicationRepository) CreateFunctionVersion(ctx context.Context, version *applicationDomain.FunctionVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetFunctionVersion(ctx context.Context, versionID string) (*applicationDomain.FunctionVersion, error) {
	args := m.Called(ctx, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*applicationDomain.FunctionVersion), args.Error(1)
}

func (m *MockApplicationRepository) GetFunctionVersions(ctx context.Context, applicationID string) ([]applicationDomain.FunctionVersion, error) {
	args := m.Called(ctx, applicationID)
	return args.Get(0).([]applicationDomain.FunctionVersion), args.Error(1)
}

func (m *MockApplicationRepository) GetActiveFunctionVersion(ctx context.Context, applicationID string) (*applicationDomain.FunctionVersion, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*applicationDomain.FunctionVersion), args.Error(1)
}

func (m *MockApplicationRepository) UpdateFunctionVersion(ctx context.Context, version *applicationDomain.FunctionVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockApplicationRepository) SetActiveFunctionVersion(ctx context.Context, applicationID, versionID string) error {
	args := m.Called(ctx, applicationID, versionID)
	return args.Error(0)
}

func (m *MockApplicationRepository) CreateFunctionInvocation(ctx context.Context, invocation *applicationDomain.FunctionInvocation) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetFunctionInvocation(ctx context.Context, invocationID string) (*applicationDomain.FunctionInvocation, error) {
	args := m.Called(ctx, invocationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*applicationDomain.FunctionInvocation), args.Error(1)
}

func (m *MockApplicationRepository) GetFunctionInvocations(ctx context.Context, applicationID string, limit, offset int) ([]applicationDomain.FunctionInvocation, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]applicationDomain.FunctionInvocation), args.Int(1), args.Error(2)
}

func (m *MockApplicationRepository) UpdateFunctionInvocation(ctx context.Context, invocation *applicationDomain.FunctionInvocation) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *MockApplicationRepository) CreateFunctionEvent(ctx context.Context, event *applicationDomain.FunctionEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetFunctionEvent(ctx context.Context, eventID string) (*applicationDomain.FunctionEvent, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*applicationDomain.FunctionEvent), args.Error(1)
}

func (m *MockApplicationRepository) GetPendingFunctionEvents(ctx context.Context, applicationID string, limit int) ([]applicationDomain.FunctionEvent, error) {
	args := m.Called(ctx, applicationID, limit)
	return args.Get(0).([]applicationDomain.FunctionEvent), args.Error(1)
}

func (m *MockApplicationRepository) UpdateFunctionEvent(ctx context.Context, event *applicationDomain.FunctionEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type MockWorkspaceRepository struct {
	mock.Mock
}

func (m *MockWorkspaceRepository) GetWorkspace(ctx context.Context, workspaceID string) (*workspaceDomain.Workspace, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspaceDomain.Workspace), args.Error(1)
}

// Implement all required methods for workspaceDomain.Repository interface
func (m *MockWorkspaceRepository) CreateWorkspace(ctx context.Context, ws *workspaceDomain.Workspace) error {
	args := m.Called(ctx, ws)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetWorkspaceByNameAndOrg(ctx context.Context, name, orgID string) (*workspaceDomain.Workspace, error) {
	args := m.Called(ctx, name, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspaceDomain.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) ListWorkspaces(ctx context.Context, filter workspaceDomain.WorkspaceFilter) ([]*workspaceDomain.Workspace, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*workspaceDomain.Workspace), args.Int(1), args.Error(2)
}

func (m *MockWorkspaceRepository) UpdateWorkspace(ctx context.Context, ws *workspaceDomain.Workspace) error {
	args := m.Called(ctx, ws)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) CreateTask(ctx context.Context, task *workspaceDomain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetTask(ctx context.Context, taskID string) (*workspaceDomain.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspaceDomain.Task), args.Error(1)
}

func (m *MockWorkspaceRepository) ListTasks(ctx context.Context, workspaceID string) ([]*workspaceDomain.Task, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]*workspaceDomain.Task), args.Error(1)
}

func (m *MockWorkspaceRepository) UpdateTask(ctx context.Context, task *workspaceDomain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetPendingTasks(ctx context.Context, taskType string, limit int) ([]*workspaceDomain.Task, error) {
	args := m.Called(ctx, taskType, limit)
	return args.Get(0).([]*workspaceDomain.Task), args.Error(1)
}

func (m *MockWorkspaceRepository) SaveWorkspaceStatus(ctx context.Context, status *workspaceDomain.WorkspaceStatus) error {
	args := m.Called(ctx, status)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetWorkspaceStatus(ctx context.Context, workspaceID string) (*workspaceDomain.WorkspaceStatus, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspaceDomain.WorkspaceStatus), args.Error(1)
}

func (m *MockWorkspaceRepository) SaveKubeconfig(ctx context.Context, workspaceID, kubeconfig string) error {
	args := m.Called(ctx, workspaceID, kubeconfig)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetKubeconfig(ctx context.Context, workspaceID string) (string, error) {
	args := m.Called(ctx, workspaceID)
	return args.String(0), args.Error(1)
}

func (m *MockWorkspaceRepository) CleanupExpiredTasks(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) CleanupDeletedWorkspaces(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) ListWorkspaceMembers(ctx context.Context, workspaceID string) ([]*workspaceDomain.WorkspaceMember, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]*workspaceDomain.WorkspaceMember), args.Error(1)
}

func (m *MockWorkspaceRepository) AddWorkspaceMember(ctx context.Context, member *workspaceDomain.WorkspaceMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) RemoveWorkspaceMember(ctx context.Context, workspaceID, userID string) error {
	args := m.Called(ctx, workspaceID, userID)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) CreateResourceUsage(ctx context.Context, usage *workspaceDomain.ResourceUsage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

// Tests

func TestCreateBackupStorage(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful storage creation for Dedicated Plan", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockProxmoxRepo := new(MockProxmoxRepository)
		mockAppRepo := new(MockApplicationRepository)
		mockWorkspaceRepo := new(MockWorkspaceRepository)

		fakeK8s := fake.NewSimpleClientset()
		service := NewService(mockRepo, mockProxmoxRepo, mockAppRepo, mockWorkspaceRepo, fakeK8s, "test-encryption-key")

		// Mock workspace on Dedicated Plan
		ws := &workspaceDomain.Workspace{
			ID:   "ws-123",
			Plan: workspaceDomain.WorkspacePlanDedicated,
		}
		mockWorkspaceRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)

		// No existing storage with same name
		mockRepo.On("GetBackupStorageByName", ctx, "ws-123", "test-storage").Return(nil, nil)

		// Create storage
		mockRepo.On("CreateBackupStorage", ctx, mock.MatchedBy(func(s *domain.BackupStorage) bool {
			return s.Name == "test-storage" &&
				s.WorkspaceID == "ws-123" &&
				s.Type == domain.StorageTypeNFS &&
				s.Status == domain.StorageStatusPending
		})).Return(nil)

		// Mock UpdateBackupStorage for the async provisioning
		mockRepo.On("UpdateBackupStorage", mock.Anything, mock.Anything).Return(nil).Maybe()

		// Mock the Proxmox operations for async provisioning
		mockProxmoxRepo.On("CreateStorage", mock.Anything, mock.Anything, mock.Anything).Return("nfs-storage-id", nil).Maybe()
		mockProxmoxRepo.On("SetStorageQuota", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		req := domain.CreateBackupStorageRequest{
			Name:             "test-storage",
			Type:             domain.StorageTypeNFS,
			ProxmoxStorageID: "nfs-storage",
			ProxmoxNodeID:    "node-1",
			CapacityGB:       100,
			ConnectionConfig: map[string]interface{}{
				"server": "192.168.1.100",
				"export": "/backup",
			},
		}

		storage, err := service.CreateBackupStorage(ctx, "ws-123", req)
		assert.NoError(t, err)
		assert.NotNil(t, storage)
		assert.Equal(t, "test-storage", storage.Name)
		// The status is set to creating when the provisioning starts
		assert.Contains(t, []domain.StorageStatus{domain.StorageStatusPending, domain.StorageStatusCreating}, storage.Status)

		mockWorkspaceRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Storage creation fails for non-Dedicated Plan", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockProxmoxRepo := new(MockProxmoxRepository)
		mockAppRepo := new(MockApplicationRepository)
		mockWorkspaceRepo := new(MockWorkspaceRepository)

		fakeK8s := fake.NewSimpleClientset()
		service := NewService(mockRepo, mockProxmoxRepo, mockAppRepo, mockWorkspaceRepo, fakeK8s, "test-encryption-key")

		// Mock workspace on Shared Plan
		ws := &workspaceDomain.Workspace{
			ID:   "ws-123",
			Plan: workspaceDomain.WorkspacePlanShared,
		}
		mockWorkspaceRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)

		req := domain.CreateBackupStorageRequest{
			Name:          "test-storage",
			Type:          domain.StorageTypeNFS,
			ProxmoxNodeID: "node-1",
			CapacityGB:    100,
		}

		storage, err := service.CreateBackupStorage(ctx, "ws-123", req)
		assert.Error(t, err)
		assert.Nil(t, storage)
		assert.Contains(t, err.Error(), "only available for Dedicated Plan")

		mockWorkspaceRepo.AssertExpectations(t)
	})
}

func TestCreateBackupPolicy(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful policy creation", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockProxmoxRepo := new(MockProxmoxRepository)
		mockAppRepo := new(MockApplicationRepository)
		mockWorkspaceRepo := new(MockWorkspaceRepository)

		fakeK8s := fake.NewSimpleClientset()
		service := NewService(mockRepo, mockProxmoxRepo, mockAppRepo, mockWorkspaceRepo, fakeK8s, "test-encryption-key")

		// Mock application
		app := &applicationDomain.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
			Name:        "test-app",
		}
		mockAppRepo.On("GetApplication", ctx, "app-123").Return(app, nil)

		// Mock workspace on Dedicated Plan
		ws := &workspaceDomain.Workspace{
			ID:   "ws-123",
			Plan: workspaceDomain.WorkspacePlanDedicated,
		}
		mockWorkspaceRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)

		// No existing policy
		mockRepo.On("GetBackupPolicyByApplication", ctx, "app-123").Return(nil, nil)

		// Mock storage
		storage := &domain.BackupStorage{
			ID:          "storage-123",
			WorkspaceID: "ws-123",
			Status:      domain.StorageStatusActive,
		}
		mockRepo.On("GetBackupStorage", ctx, "storage-123").Return(storage, nil)

		// Create policy
		mockRepo.On("CreateBackupPolicy", ctx, mock.MatchedBy(func(p *domain.BackupPolicy) bool {
			return p.ApplicationID == "app-123" &&
				p.StorageID == "storage-123" &&
				p.Schedule == "0 2 * * *" &&
				p.Enabled == true
		})).Return(nil)

		req := domain.CreateBackupPolicyRequest{
			StorageID:          "storage-123",
			Enabled:            true,
			Schedule:           "0 2 * * *",
			RetentionDays:      30,
			BackupType:         domain.BackupTypeFull,
			IncludeVolumes:     true,
			IncludeDatabase:    true,
			IncludeConfig:      true,
			CompressionEnabled: true,
			EncryptionEnabled:  true,
		}

		policy, err := service.CreateBackupPolicy(ctx, "app-123", req)
		assert.NoError(t, err)
		assert.NotNil(t, policy)
		assert.Equal(t, "app-123", policy.ApplicationID)
		assert.Equal(t, "storage-123", policy.StorageID)
		assert.Equal(t, "0 2 * * *", policy.Schedule)

		mockAppRepo.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Policy creation fails if storage is not active", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockProxmoxRepo := new(MockProxmoxRepository)
		mockAppRepo := new(MockApplicationRepository)
		mockWorkspaceRepo := new(MockWorkspaceRepository)

		fakeK8s := fake.NewSimpleClientset()
		service := NewService(mockRepo, mockProxmoxRepo, mockAppRepo, mockWorkspaceRepo, fakeK8s, "test-encryption-key")

		// Mock application
		app := &applicationDomain.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
		}
		mockAppRepo.On("GetApplication", ctx, "app-123").Return(app, nil)

		// Mock workspace on Dedicated Plan
		ws := &workspaceDomain.Workspace{
			ID:   "ws-123",
			Plan: workspaceDomain.WorkspacePlanDedicated,
		}
		mockWorkspaceRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)

		// No existing policy
		mockRepo.On("GetBackupPolicyByApplication", ctx, "app-123").Return(nil, nil)

		// Mock storage in pending state
		storage := &domain.BackupStorage{
			ID:          "storage-123",
			WorkspaceID: "ws-123",
			Status:      domain.StorageStatusPending,
		}
		mockRepo.On("GetBackupStorage", ctx, "storage-123").Return(storage, nil)

		req := domain.CreateBackupPolicyRequest{
			StorageID:     "storage-123",
			Schedule:      "0 2 * * *",
			BackupType:    domain.BackupTypeFull,
			RetentionDays: 30,
		}

		policy, err := service.CreateBackupPolicy(ctx, "app-123", req)
		assert.Error(t, err)
		assert.Nil(t, policy)
		assert.Contains(t, err.Error(), "storage is not active")

		mockAppRepo.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestTriggerManualBackup(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful manual backup trigger", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockProxmoxRepo := new(MockProxmoxRepository)
		mockAppRepo := new(MockApplicationRepository)
		mockWorkspaceRepo := new(MockWorkspaceRepository)

		fakeK8s := fake.NewSimpleClientset()
		service := NewService(mockRepo, mockProxmoxRepo, mockAppRepo, mockWorkspaceRepo, fakeK8s, "test-encryption-key")

		// Mock policy
		policy := &domain.BackupPolicy{
			ID:            "policy-123",
			ApplicationID: "app-123",
			BackupType:    domain.BackupTypeFull,
		}
		mockRepo.On("GetBackupPolicyByApplication", ctx, "app-123").Return(policy, nil)

		// Create execution
		mockRepo.On("CreateBackupExecution", ctx, mock.MatchedBy(func(e *domain.BackupExecution) bool {
			return e.PolicyID == "policy-123" &&
				e.Status == domain.BackupExecutionStatusRunning
		})).Return(nil)

		req := domain.TriggerBackupRequest{
			ApplicationID: "app-123",
			Metadata: map[string]interface{}{
				"triggered_by": "user",
			},
		}

		execution, err := service.TriggerManualBackup(ctx, "app-123", req)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Equal(t, "policy-123", execution.PolicyID)
		assert.Equal(t, domain.BackupExecutionStatusRunning, execution.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Manual backup fails without policy", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockProxmoxRepo := new(MockProxmoxRepository)
		mockAppRepo := new(MockApplicationRepository)
		mockWorkspaceRepo := new(MockWorkspaceRepository)

		fakeK8s := fake.NewSimpleClientset()
		service := NewService(mockRepo, mockProxmoxRepo, mockAppRepo, mockWorkspaceRepo, fakeK8s, "test-encryption-key")

		// No policy found
		mockRepo.On("GetBackupPolicyByApplication", ctx, "app-123").Return(nil, nil)

		req := domain.TriggerBackupRequest{
			ApplicationID: "app-123",
		}

		execution, err := service.TriggerManualBackup(ctx, "app-123", req)
		assert.Error(t, err)
		assert.Nil(t, execution)
		assert.Contains(t, err.Error(), "no backup policy found")

		mockRepo.AssertExpectations(t)
	})
}

func TestRestoreBackup(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful restore initiation", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockProxmoxRepo := new(MockProxmoxRepository)
		mockAppRepo := new(MockApplicationRepository)
		mockWorkspaceRepo := new(MockWorkspaceRepository)

		fakeK8s := fake.NewSimpleClientset()
		service := NewService(mockRepo, mockProxmoxRepo, mockAppRepo, mockWorkspaceRepo, fakeK8s, "test-encryption-key")

		// Mock backup execution
		execution := &domain.BackupExecution{
			ID:       "exec-123",
			PolicyID: "policy-123",
			Status:   domain.BackupExecutionStatusSucceeded,
		}
		mockRepo.On("GetBackupExecution", ctx, "exec-123").Return(execution, nil)

		// Mock policy
		policy := &domain.BackupPolicy{
			ID:            "policy-123",
			ApplicationID: "app-123",
		}
		mockRepo.On("GetBackupPolicy", ctx, "policy-123").Return(policy, nil)

		// Create restore
		mockRepo.On("CreateBackupRestore", ctx, mock.MatchedBy(func(r *domain.BackupRestore) bool {
			return r.BackupExecutionID == "exec-123" &&
				r.ApplicationID == "app-123" &&
				r.Status == domain.RestoreStatusPending &&
				r.RestoreType == domain.RestoreTypeFull
		})).Return(nil)

		req := domain.RestoreBackupRequest{
			RestoreType: domain.RestoreTypeFull,
			RestoreOptions: map[string]interface{}{
				"verify": true,
			},
		}

		restore, err := service.RestoreBackup(ctx, "exec-123", req)
		assert.NoError(t, err)
		assert.NotNil(t, restore)
		assert.Equal(t, "exec-123", restore.BackupExecutionID)
		assert.Equal(t, "app-123", restore.ApplicationID)
		assert.Equal(t, domain.RestoreStatusPending, restore.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Restore fails for unsuccessful backup", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockProxmoxRepo := new(MockProxmoxRepository)
		mockAppRepo := new(MockApplicationRepository)
		mockWorkspaceRepo := new(MockWorkspaceRepository)

		fakeK8s := fake.NewSimpleClientset()
		service := NewService(mockRepo, mockProxmoxRepo, mockAppRepo, mockWorkspaceRepo, fakeK8s, "test-encryption-key")

		// Mock failed backup execution
		execution := &domain.BackupExecution{
			ID:       "exec-123",
			PolicyID: "policy-123",
			Status:   domain.BackupExecutionStatusFailed,
		}
		mockRepo.On("GetBackupExecution", ctx, "exec-123").Return(execution, nil)

		req := domain.RestoreBackupRequest{
			RestoreType: domain.RestoreTypeFull,
		}

		restore, err := service.RestoreBackup(ctx, "exec-123", req)
		assert.Error(t, err)
		assert.Nil(t, restore)
		assert.Contains(t, err.Error(), "can only restore from successful backups")

		mockRepo.AssertExpectations(t)
	})
}