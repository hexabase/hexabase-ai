package backup

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	"github.com/hexabase/hexabase-ai/api/internal/domain/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/kubernetes/fake"
)

// Mock implementations

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateBackupStorage(ctx context.Context, storage *backup.BackupStorage) error {
	args := m.Called(ctx, storage)
	return args.Error(0)
}

func (m *MockRepository) GetBackupStorage(ctx context.Context, storageID string) (*backup.BackupStorage, error) {
	args := m.Called(ctx, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupStorage), args.Error(1)
}

func (m *MockRepository) GetBackupStorageByName(ctx context.Context, workspaceID, name string) (*backup.BackupStorage, error) {
	args := m.Called(ctx, workspaceID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupStorage), args.Error(1)
}

func (m *MockRepository) ListBackupStorages(ctx context.Context, workspaceID string) ([]backup.BackupStorage, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]backup.BackupStorage), args.Error(1)
}

func (m *MockRepository) UpdateBackupStorage(ctx context.Context, storage *backup.BackupStorage) error {
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

func (m *MockRepository) CreateBackupPolicy(ctx context.Context, policy *backup.BackupPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockRepository) GetBackupPolicy(ctx context.Context, policyID string) (*backup.BackupPolicy, error) {
	args := m.Called(ctx, policyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupPolicy), args.Error(1)
}

func (m *MockRepository) GetBackupPolicyByApplication(ctx context.Context, applicationID string) (*backup.BackupPolicy, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupPolicy), args.Error(1)
}

func (m *MockRepository) ListBackupPolicies(ctx context.Context, workspaceID string) ([]backup.BackupPolicy, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]backup.BackupPolicy), args.Error(1)
}

func (m *MockRepository) UpdateBackupPolicy(ctx context.Context, policy *backup.BackupPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockRepository) DeleteBackupPolicy(ctx context.Context, policyID string) error {
	args := m.Called(ctx, policyID)
	return args.Error(0)
}

func (m *MockRepository) ListEnabledPolicies(ctx context.Context) ([]backup.BackupPolicy, error) {
	args := m.Called(ctx)
	return args.Get(0).([]backup.BackupPolicy), args.Error(1)
}

func (m *MockRepository) CreateBackupExecution(ctx context.Context, execution *backup.BackupExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockRepository) GetBackupExecution(ctx context.Context, executionID string) (*backup.BackupExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupExecution), args.Error(1)
}

func (m *MockRepository) ListBackupExecutions(ctx context.Context, policyID string, limit, offset int) ([]backup.BackupExecution, int, error) {
	args := m.Called(ctx, policyID, limit, offset)
	return args.Get(0).([]backup.BackupExecution), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateBackupExecution(ctx context.Context, execution *backup.BackupExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockRepository) GetLatestBackupExecution(ctx context.Context, policyID string) (*backup.BackupExecution, error) {
	args := m.Called(ctx, policyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupExecution), args.Error(1)
}

func (m *MockRepository) GetBackupExecutionsByApplication(ctx context.Context, applicationID string, limit, offset int) ([]backup.BackupExecution, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]backup.BackupExecution), args.Int(1), args.Error(2)
}

func (m *MockRepository) CleanupOldBackups(ctx context.Context, policyID string, retentionDays int) error {
	args := m.Called(ctx, policyID, retentionDays)
	return args.Error(0)
}

func (m *MockRepository) CreateBackupRestore(ctx context.Context, restore *backup.BackupRestore) error {
	args := m.Called(ctx, restore)
	return args.Error(0)
}

func (m *MockRepository) GetBackupRestore(ctx context.Context, restoreID string) (*backup.BackupRestore, error) {
	args := m.Called(ctx, restoreID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupRestore), args.Error(1)
}

func (m *MockRepository) ListBackupRestores(ctx context.Context, applicationID string, limit, offset int) ([]backup.BackupRestore, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]backup.BackupRestore), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateBackupRestore(ctx context.Context, restore *backup.BackupRestore) error {
	args := m.Called(ctx, restore)
	return args.Error(0)
}

func (m *MockRepository) GetStorageUsage(ctx context.Context, storageID string) (*backup.BackupStorageUsage, error) {
	args := m.Called(ctx, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupStorageUsage), args.Error(1)
}

func (m *MockRepository) GetWorkspaceStorageUsage(ctx context.Context, workspaceID string) ([]backup.BackupStorageUsage, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]backup.BackupStorageUsage), args.Error(1)
}

type MockProxmoxRepository struct {
	mock.Mock
}

func (m *MockProxmoxRepository) CreateStorage(ctx context.Context, nodeID string, config backup.ProxmoxStorageConfig) (string, error) {
	args := m.Called(ctx, nodeID, config)
	return args.String(0), args.Error(1)
}

func (m *MockProxmoxRepository) DeleteStorage(ctx context.Context, nodeID, storageID string) error {
	args := m.Called(ctx, nodeID, storageID)
	return args.Error(0)
}

func (m *MockProxmoxRepository) GetStorageInfo(ctx context.Context, nodeID, storageID string) (*backup.ProxmoxStorageInfo, error) {
	args := m.Called(ctx, nodeID, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.ProxmoxStorageInfo), args.Error(1)
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

func (m *MockProxmoxRepository) GetVolumeInfo(ctx context.Context, nodeID, storageID, volumeID string) (*backup.ProxmoxVolumeInfo, error) {
	args := m.Called(ctx, nodeID, storageID, volumeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.ProxmoxVolumeInfo), args.Error(1)
}

func (m *MockProxmoxRepository) SetStorageQuota(ctx context.Context, nodeID, storageID string, quotaGB int) error {
	args := m.Called(ctx, nodeID, storageID, quotaGB)
	return args.Error(0)
}

func (m *MockProxmoxRepository) GetStorageQuota(ctx context.Context, nodeID, storageID string) (*backup.ProxmoxStorageQuota, error) {
	args := m.Called(ctx, nodeID, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.ProxmoxStorageQuota), args.Error(1)
}

type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) GetApplication(ctx context.Context, applicationID string) (*application.Application, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.Application), args.Error(1)
}

type MockWorkspaceRepository struct {
	mock.Mock
}

func (m *MockWorkspaceRepository) GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspace.Workspace), args.Error(1)
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
		ws := &workspace.Workspace{
			ID:   "ws-123",
			Plan: workspace.WorkspacePlanDedicated,
		}
		mockWorkspaceRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)

		// No existing storage with same name
		mockRepo.On("GetBackupStorageByName", ctx, "ws-123", "test-storage").Return(nil, nil)

		// Create storage
		mockRepo.On("CreateBackupStorage", ctx, mock.MatchedBy(func(s *backup.BackupStorage) bool {
			return s.Name == "test-storage" &&
				s.WorkspaceID == "ws-123" &&
				s.Type == backup.StorageTypeNFS &&
				s.Status == backup.StorageStatusPending
		})).Return(nil)

		req := backup.CreateBackupStorageRequest{
			Name:             "test-storage",
			Type:             backup.StorageTypeNFS,
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
		assert.Equal(t, backup.StorageStatusPending, storage.Status)

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
		ws := &workspace.Workspace{
			ID:   "ws-123",
			Plan: workspace.WorkspacePlanShared,
		}
		mockWorkspaceRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)

		req := backup.CreateBackupStorageRequest{
			Name:          "test-storage",
			Type:          backup.StorageTypeNFS,
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
		app := &application.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
			Name:        "test-app",
		}
		mockAppRepo.On("GetApplication", ctx, "app-123").Return(app, nil)

		// Mock workspace on Dedicated Plan
		ws := &workspace.Workspace{
			ID:   "ws-123",
			Plan: workspace.WorkspacePlanDedicated,
		}
		mockWorkspaceRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)

		// No existing policy
		mockRepo.On("GetBackupPolicyByApplication", ctx, "app-123").Return(nil, nil)

		// Mock storage
		storage := &backup.BackupStorage{
			ID:          "storage-123",
			WorkspaceID: "ws-123",
			Status:      backup.StorageStatusActive,
		}
		mockRepo.On("GetBackupStorage", ctx, "storage-123").Return(storage, nil)

		// Create policy
		mockRepo.On("CreateBackupPolicy", ctx, mock.MatchedBy(func(p *backup.BackupPolicy) bool {
			return p.ApplicationID == "app-123" &&
				p.StorageID == "storage-123" &&
				p.Schedule == "0 2 * * *" &&
				p.Enabled == true
		})).Return(nil)

		req := backup.CreateBackupPolicyRequest{
			StorageID:          "storage-123",
			Enabled:            true,
			Schedule:           "0 2 * * *",
			RetentionDays:      30,
			BackupType:         backup.BackupTypeFull,
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
		app := &application.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
		}
		mockAppRepo.On("GetApplication", ctx, "app-123").Return(app, nil)

		// Mock workspace on Dedicated Plan
		ws := &workspace.Workspace{
			ID:   "ws-123",
			Plan: workspace.WorkspacePlanDedicated,
		}
		mockWorkspaceRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)

		// No existing policy
		mockRepo.On("GetBackupPolicyByApplication", ctx, "app-123").Return(nil, nil)

		// Mock storage in pending state
		storage := &backup.BackupStorage{
			ID:          "storage-123",
			WorkspaceID: "ws-123",
			Status:      backup.StorageStatusPending,
		}
		mockRepo.On("GetBackupStorage", ctx, "storage-123").Return(storage, nil)

		req := backup.CreateBackupPolicyRequest{
			StorageID:     "storage-123",
			Schedule:      "0 2 * * *",
			BackupType:    backup.BackupTypeFull,
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
		policy := &backup.BackupPolicy{
			ID:            "policy-123",
			ApplicationID: "app-123",
			BackupType:    backup.BackupTypeFull,
		}
		mockRepo.On("GetBackupPolicyByApplication", ctx, "app-123").Return(policy, nil)

		// Create execution
		mockRepo.On("CreateBackupExecution", ctx, mock.MatchedBy(func(e *backup.BackupExecution) bool {
			return e.PolicyID == "policy-123" &&
				e.Status == backup.BackupExecutionStatusRunning
		})).Return(nil)

		req := backup.TriggerBackupRequest{
			ApplicationID: "app-123",
			Metadata: map[string]interface{}{
				"triggered_by": "user",
			},
		}

		execution, err := service.TriggerManualBackup(ctx, "app-123", req)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Equal(t, "policy-123", execution.PolicyID)
		assert.Equal(t, backup.BackupExecutionStatusRunning, execution.Status)

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

		req := backup.TriggerBackupRequest{
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
		execution := &backup.BackupExecution{
			ID:       "exec-123",
			PolicyID: "policy-123",
			Status:   backup.BackupExecutionStatusSucceeded,
		}
		mockRepo.On("GetBackupExecution", ctx, "exec-123").Return(execution, nil)

		// Mock policy
		policy := &backup.BackupPolicy{
			ID:            "policy-123",
			ApplicationID: "app-123",
		}
		mockRepo.On("GetBackupPolicy", ctx, "policy-123").Return(policy, nil)

		// Create restore
		mockRepo.On("CreateBackupRestore", ctx, mock.MatchedBy(func(r *backup.BackupRestore) bool {
			return r.BackupExecutionID == "exec-123" &&
				r.ApplicationID == "app-123" &&
				r.Status == backup.RestoreStatusPending &&
				r.RestoreType == backup.RestoreTypeFull
		})).Return(nil)

		req := backup.RestoreBackupRequest{
			RestoreType: backup.RestoreTypeFull,
			RestoreOptions: map[string]interface{}{
				"verify": true,
			},
		}

		restore, err := service.RestoreBackup(ctx, "exec-123", req)
		assert.NoError(t, err)
		assert.NotNil(t, restore)
		assert.Equal(t, "exec-123", restore.BackupExecutionID)
		assert.Equal(t, "app-123", restore.ApplicationID)
		assert.Equal(t, backup.RestoreStatusPending, restore.Status)

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
		execution := &backup.BackupExecution{
			ID:       "exec-123",
			PolicyID: "policy-123",
			Status:   backup.BackupExecutionStatusFailed,
		}
		mockRepo.On("GetBackupExecution", ctx, "exec-123").Return(execution, nil)

		req := backup.RestoreBackupRequest{
			RestoreType: backup.RestoreTypeFull,
		}

		restore, err := service.RestoreBackup(ctx, "exec-123", req)
		assert.Error(t, err)
		assert.Nil(t, restore)
		assert.Contains(t, err.Error(), "can only restore from successful backups")

		mockRepo.AssertExpectations(t)
	})
}