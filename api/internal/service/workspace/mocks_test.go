package workspace

import (
	"context"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/workspace"
	"github.com/hexabase/hexabase-ai/api/internal/helm"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the workspace Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateWorkspace(ctx context.Context, ws *workspace.Workspace) error {
	args := m.Called(ctx, ws)
	return args.Error(0)
}

func (m *MockRepository) GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspace.Workspace), args.Error(1)
}

func (m *MockRepository) GetWorkspaceByNameAndOrg(ctx context.Context, name, orgID string) (*workspace.Workspace, error) {
	args := m.Called(ctx, name, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspace.Workspace), args.Error(1)
}

func (m *MockRepository) ListWorkspaces(ctx context.Context, filter workspace.WorkspaceFilter) ([]*workspace.Workspace, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*workspace.Workspace), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateWorkspace(ctx context.Context, ws *workspace.Workspace) error {
	args := m.Called(ctx, ws)
	return args.Error(0)
}

func (m *MockRepository) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockRepository) CreateTask(ctx context.Context, task *workspace.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockRepository) GetTask(ctx context.Context, taskID string) (*workspace.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspace.Task), args.Error(1)
}

func (m *MockRepository) ListTasks(ctx context.Context, workspaceID string) ([]*workspace.Task, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*workspace.Task), args.Error(1)
}

func (m *MockRepository) UpdateTask(ctx context.Context, task *workspace.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockRepository) GetPendingTasks(ctx context.Context, taskType string, limit int) ([]*workspace.Task, error) {
	args := m.Called(ctx, taskType, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*workspace.Task), args.Error(1)
}

func (m *MockRepository) SaveWorkspaceStatus(ctx context.Context, status *workspace.WorkspaceStatus) error {
	args := m.Called(ctx, status)
	return args.Error(0)
}

func (m *MockRepository) GetWorkspaceStatus(ctx context.Context, workspaceID string) (*workspace.WorkspaceStatus, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspace.WorkspaceStatus), args.Error(1)
}

func (m *MockRepository) SaveKubeconfig(ctx context.Context, workspaceID, kubeconfig string) error {
	args := m.Called(ctx, workspaceID, kubeconfig)
	return args.Error(0)
}

func (m *MockRepository) GetKubeconfig(ctx context.Context, workspaceID string) (string, error) {
	args := m.Called(ctx, workspaceID)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) CleanupExpiredTasks(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockRepository) CleanupDeletedWorkspaces(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockRepository) ListWorkspaceMembers(ctx context.Context, workspaceID string) ([]*workspace.WorkspaceMember, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*workspace.WorkspaceMember), args.Error(1)
}

func (m *MockRepository) AddWorkspaceMember(ctx context.Context, member *workspace.WorkspaceMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockRepository) RemoveWorkspaceMember(ctx context.Context, workspaceID, userID string) error {
	args := m.Called(ctx, workspaceID, userID)
	return args.Error(0)
}

func (m *MockRepository) CreateResourceUsage(ctx context.Context, usage *workspace.ResourceUsage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

// MockKubernetesRepository is a mock implementation of the KubernetesRepository interface
type MockKubernetesRepository struct {
	mock.Mock
}

func (m *MockKubernetesRepository) CreateVCluster(ctx context.Context, workspaceID, plan string) error {
	args := m.Called(ctx, workspaceID, plan)
	return args.Error(0)
}

func (m *MockKubernetesRepository) DeleteVCluster(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockKubernetesRepository) WaitForVClusterReady(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockKubernetesRepository) WaitForVClusterDeleted(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetVClusterStatus(ctx context.Context, workspaceID string) (string, error) {
	args := m.Called(ctx, workspaceID)
	return args.String(0), args.Error(1)
}

func (m *MockKubernetesRepository) GetVClusterInfo(ctx context.Context, workspaceID string) (*workspace.ClusterInfo, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspace.ClusterInfo), args.Error(1)
}

func (m *MockKubernetesRepository) ScaleVCluster(ctx context.Context, workspaceID string, replicas int) error {
	args := m.Called(ctx, workspaceID, replicas)
	return args.Error(0)
}

func (m *MockKubernetesRepository) ConfigureOIDC(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockKubernetesRepository) UpdateOIDCConfig(ctx context.Context, workspaceID string, config map[string]interface{}) error {
	args := m.Called(ctx, workspaceID, config)
	return args.Error(0)
}

func (m *MockKubernetesRepository) ApplyResourceQuotas(ctx context.Context, workspaceID, plan string) error {
	args := m.Called(ctx, workspaceID, plan)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetResourceMetrics(ctx context.Context, workspaceID string) (*workspace.ResourceUsage, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspace.ResourceUsage), args.Error(1)
}

func (m *MockKubernetesRepository) ListVClusterNodes(ctx context.Context, workspaceID string) ([]workspace.Node, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]workspace.Node), args.Error(1)
}

func (m *MockKubernetesRepository) ScaleVClusterDeployment(ctx context.Context, workspaceID, deploymentName string, replicas int) error {
	args := m.Called(ctx, workspaceID, deploymentName, replicas)
	return args.Error(0)
}

// MockAuthRepository is a mock implementation of the AuthRepository interface
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) GetUser(ctx context.Context, userID string) (*workspace.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workspace.User), args.Error(1)
}

func (m *MockAuthRepository) GenerateWorkspaceToken(ctx context.Context, userID, workspaceID string) (string, error) {
	args := m.Called(ctx, userID, workspaceID)
	return args.String(0), args.Error(1)
}

// MockHelmService is a mock implementation of the helm.Service interface
type MockHelmService struct {
	mock.Mock
}

func (m *MockHelmService) InstallOrUpgrade(releaseName, chartPath, namespace string, values map[string]interface{}) error {
	args := m.Called(releaseName, chartPath, namespace, values)
	return args.Error(0)
}

// Ensure interfaces are satisfied
var _ workspace.Repository = (*MockRepository)(nil)
var _ workspace.KubernetesRepository = (*MockKubernetesRepository)(nil)
var _ workspace.AuthRepository = (*MockAuthRepository)(nil)
var _ helm.Service = (*MockHelmService)(nil)