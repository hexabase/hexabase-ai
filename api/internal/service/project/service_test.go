package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
)

// Mock implementations

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateProject(ctx context.Context, proj *project.Project) error {
	args := m.Called(ctx, proj)
	return args.Error(0)
}

func (m *MockRepository) GetProject(ctx context.Context, projectID string) (*project.Project, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *MockRepository) GetProjectByName(ctx context.Context, name string) (*project.Project, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *MockRepository) GetProjectByNameAndWorkspace(ctx context.Context, name, workspaceID string) (*project.Project, error) {
	args := m.Called(ctx, name, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *MockRepository) ListProjects(ctx context.Context, filter project.ProjectFilter) ([]*project.Project, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*project.Project), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateProject(ctx context.Context, proj *project.Project) error {
	args := m.Called(ctx, proj)
	return args.Error(0)
}

func (m *MockRepository) DeleteProject(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockRepository) CountProjects(ctx context.Context, workspaceID string) (int, error) {
	args := m.Called(ctx, workspaceID)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) CreateNamespace(ctx context.Context, namespace *project.Namespace) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *MockRepository) GetNamespace(ctx context.Context, namespaceID string) (*project.Namespace, error) {
	args := m.Called(ctx, namespaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Namespace), args.Error(1)
}

func (m *MockRepository) GetNamespaceByName(ctx context.Context, projectID, name string) (*project.Namespace, error) {
	args := m.Called(ctx, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Namespace), args.Error(1)
}

func (m *MockRepository) ListNamespaces(ctx context.Context, projectID string) ([]*project.Namespace, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*project.Namespace), args.Error(1)
}

func (m *MockRepository) UpdateNamespace(ctx context.Context, namespace *project.Namespace) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *MockRepository) DeleteNamespace(ctx context.Context, namespaceID string) error {
	args := m.Called(ctx, namespaceID)
	return args.Error(0)
}

func (m *MockRepository) AddMember(ctx context.Context, member *project.ProjectMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockRepository) GetMember(ctx context.Context, projectID, userID string) (*project.ProjectMember, error) {
	args := m.Called(ctx, projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ProjectMember), args.Error(1)
}

func (m *MockRepository) GetMemberByID(ctx context.Context, memberID string) (*project.ProjectMember, error) {
	args := m.Called(ctx, memberID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ProjectMember), args.Error(1)
}

func (m *MockRepository) ListMembers(ctx context.Context, projectID string) ([]*project.ProjectMember, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*project.ProjectMember), args.Error(1)
}

func (m *MockRepository) UpdateMember(ctx context.Context, member *project.ProjectMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockRepository) RemoveMember(ctx context.Context, memberID string) error {
	args := m.Called(ctx, memberID)
	return args.Error(0)
}

func (m *MockRepository) CountMembers(ctx context.Context, projectID string) (int, error) {
	args := m.Called(ctx, projectID)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) CreateActivity(ctx context.Context, activity *project.ProjectActivity) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

func (m *MockRepository) ListActivities(ctx context.Context, filter project.ActivityFilter) ([]*project.ProjectActivity, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*project.ProjectActivity), args.Error(1)
}

func (m *MockRepository) GetLastActivity(ctx context.Context, projectID string) (*project.ProjectActivity, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ProjectActivity), args.Error(1)
}

func (m *MockRepository) CleanupOldActivities(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockRepository) GetChildProjects(ctx context.Context, parentID string) ([]*project.Project, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]*project.Project), args.Error(1)
}

func (m *MockRepository) GetUser(ctx context.Context, userID string) (*project.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*project.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.User), args.Error(1)
}

func (m *MockRepository) GetProjectResourceUsage(ctx context.Context, projectID string) (*project.ResourceUsage, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ResourceUsage), args.Error(1)
}

func (m *MockRepository) GetNamespaceResourceUsage(ctx context.Context, namespaceID string) (*project.NamespaceUsage, error) {
	args := m.Called(ctx, namespaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.NamespaceUsage), args.Error(1)
}

type MockKubernetesRepository struct {
	mock.Mock
}

func (m *MockKubernetesRepository) CreateNamespace(ctx context.Context, workspaceID, name string, labels map[string]string) error {
	args := m.Called(ctx, workspaceID, name, labels)
	return args.Error(0)
}

func (m *MockKubernetesRepository) DeleteNamespace(ctx context.Context, workspaceID, name string) error {
	args := m.Called(ctx, workspaceID, name)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetNamespace(ctx context.Context, workspaceID, name string) (map[string]interface{}, error) {
	args := m.Called(ctx, workspaceID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockKubernetesRepository) ListNamespaces(ctx context.Context, workspaceID string) ([]string, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockKubernetesRepository) CreateResourceQuota(ctx context.Context, workspaceID, namespace string, quota *project.ResourceQuota) error {
	args := m.Called(ctx, workspaceID, namespace, quota)
	return args.Error(0)
}

func (m *MockKubernetesRepository) UpdateResourceQuota(ctx context.Context, workspaceID, namespace string, quota *project.ResourceQuota) error {
	args := m.Called(ctx, workspaceID, namespace, quota)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetResourceQuota(ctx context.Context, workspaceID, namespace string) (*project.ResourceQuota, error) {
	args := m.Called(ctx, workspaceID, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ResourceQuota), args.Error(1)
}

func (m *MockKubernetesRepository) DeleteResourceQuota(ctx context.Context, workspaceID, namespace string) error {
	args := m.Called(ctx, workspaceID, namespace)
	return args.Error(0)
}

func (m *MockKubernetesRepository) ApplyResourceQuota(ctx context.Context, workspaceID, namespace string, quota *project.ResourceQuota) error {
	args := m.Called(ctx, workspaceID, namespace, quota)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetNamespaceUsage(ctx context.Context, workspaceID, namespace string) (*project.NamespaceUsage, error) {
	args := m.Called(ctx, workspaceID, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.NamespaceUsage), args.Error(1)
}

func (m *MockKubernetesRepository) GetNamespaceResourceUsage(ctx context.Context, workspaceID, namespace string) (*project.ResourceUsage, error) {
	args := m.Called(ctx, workspaceID, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ResourceUsage), args.Error(1)
}

func (m *MockKubernetesRepository) ApplyRBAC(ctx context.Context, workspaceID, namespace, userID, role string) error {
	args := m.Called(ctx, workspaceID, namespace, userID, role)
	return args.Error(0)
}

func (m *MockKubernetesRepository) RemoveRBAC(ctx context.Context, workspaceID, namespace, userID string) error {
	args := m.Called(ctx, workspaceID, namespace, userID)
	return args.Error(0)
}

func (m *MockKubernetesRepository) ConfigureHNC(ctx context.Context, workspaceID, parentNamespace, childNamespace string) error {
	args := m.Called(ctx, workspaceID, parentNamespace, childNamespace)
	return args.Error(0)
}

// Helper function to create a test logger
func createTestLogger() *slog.Logger {
	return slog.Default()
}

// Tests

func TestCreateProject(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful project creation", func(t *testing.T) {
		req := &project.CreateProjectRequest{
			Name:        "test-project",
			DisplayName: "Test Project",
			Description: "A test project",
			WorkspaceID: "ws-123",
		}

		// Mock expectations
		mockRepo.On("GetProjectByNameAndWorkspace", ctx, "test-project", "ws-123").Return(nil, errors.New("not found"))
		mockRepo.On("CreateProject", ctx, mock.MatchedBy(func(p *project.Project) bool {
			return p.Name == "test-project" &&
				p.DisplayName == "Test Project" &&
				p.Description == "A test project" &&
				p.WorkspaceID == "ws-123"
		})).Return(nil)
		
		mockK8sRepo.On("CreateNamespace", ctx, "ws-123", "test-project", mock.MatchedBy(func(labels map[string]string) bool {
			return labels["hexabase.ai/workspace-id"] == "ws-123" &&
				labels["hexabase.ai/managed"] == "true" &&
				labels["hexabase.ai/project-id"] != ""
		})).Return(nil)
		mockRepo.On("UpdateProject", ctx, mock.AnythingOfType("*project.Project")).Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		proj, err := service.CreateProject(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, proj)
		assert.Equal(t, "test-project", proj.Name)
		assert.Equal(t, "Test Project", proj.DisplayName)
		assert.Equal(t, "ws-123", proj.WorkspaceID)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("project name already exists", func(t *testing.T) {
		req := &project.CreateProjectRequest{
			Name:        "existing-project",
			WorkspaceID: "ws-123",
		}

		existingProject := &project.Project{
			ID:   "proj-existing",
			Name: "existing-project",
		}

		mockRepo.On("GetProjectByNameAndWorkspace", ctx, "existing-project", "ws-123").Return(existingProject, nil)

		// Execute
		proj, err := service.CreateProject(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, proj)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid project name", func(t *testing.T) {
		req := &project.CreateProjectRequest{
			Name:        "Invalid-Name!",
			WorkspaceID: "ws-123",
		}

		// Execute
		proj, err := service.CreateProject(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, proj)
		assert.Contains(t, err.Error(), "invalid project name")
	})

	t.Run("empty project name", func(t *testing.T) {
		req := &project.CreateProjectRequest{
			Name:        "",
			WorkspaceID: "ws-123",
		}

		// Execute
		proj, err := service.CreateProject(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, proj)
		assert.Contains(t, err.Error(), "project name is required")
	})

	t.Run("namespace creation fails but project still created", func(t *testing.T) {
		req := &project.CreateProjectRequest{
			Name:        "test-project-no-ns",
			WorkspaceID: "ws-123",
		}

		mockRepo.On("GetProjectByNameAndWorkspace", ctx, "test-project-no-ns", "ws-123").Return(nil, errors.New("not found"))
		mockRepo.On("CreateProject", ctx, mock.AnythingOfType("*project.Project")).Return(nil)
		mockK8sRepo.On("CreateNamespace", ctx, "ws-123", "test-project-no-ns", mock.Anything).Return(errors.New("namespace creation failed"))
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		proj, err := service.CreateProject(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, proj)
		assert.Equal(t, "test-project-no-ns", proj.Name)
		assert.Empty(t, proj.NamespaceName) // Namespace not set due to failure
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})
}

func TestGetProject(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful project retrieval with namespace status", func(t *testing.T) {
		proj := &project.Project{
			ID:            "proj-123",
			Name:          "test-project",
			WorkspaceID:   "ws-123",
			NamespaceName: "test-project",
		}

		namespaceInfo := map[string]interface{}{
			"status": "active",
		}

		usage := &project.ResourceUsage{
			CPU:    "250m",
			Memory: "512Mi",
			Pods:   5,
		}

		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepo.On("GetNamespace", ctx, "ws-123", "test-project").Return(namespaceInfo, nil)
		mockK8sRepo.On("GetNamespaceResourceUsage", ctx, "ws-123", "test-project").Return(usage, nil)

		// Execute
		result, err := service.GetProject(ctx, "proj-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "proj-123", result.ID)
		assert.Equal(t, "active", result.Status)
		assert.Equal(t, usage, result.ResourceUsage)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("project not found", func(t *testing.T) {
		mockRepo.On("GetProject", ctx, "proj-not-found").Return(nil, errors.New("not found"))

		// Execute
		result, err := service.GetProject(ctx, "proj-not-found")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get project")
		mockRepo.AssertExpectations(t)
	})

	t.Run("namespace status retrieval fails", func(t *testing.T) {
		// Create new mock instances for this test
		mockRepoLocal := new(MockRepository)
		mockK8sRepoLocal := new(MockKubernetesRepository)
		serviceLocal := NewService(mockRepoLocal, mockK8sRepoLocal, logger)
		
		proj := &project.Project{
			ID:            "proj-123",
			Name:          "test-project",
			WorkspaceID:   "ws-123",
			NamespaceName: "test-project",
		}

		mockRepoLocal.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepoLocal.On("GetNamespace", ctx, "ws-123", "test-project").Return(nil, errors.New("namespace not found"))
		mockK8sRepoLocal.On("GetNamespaceResourceUsage", ctx, "ws-123", "test-project").Return(nil, errors.New("usage not found"))

		// Execute
		result, err := serviceLocal.GetProject(ctx, "proj-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "proj-123", result.ID)
		assert.Empty(t, result.Status) // Status not set due to failure
		assert.Nil(t, result.ResourceUsage) // Usage not set due to failure
		mockRepoLocal.AssertExpectations(t)
		mockK8sRepoLocal.AssertExpectations(t)
	})
}

func TestListProjects(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful project listing", func(t *testing.T) {
		filter := project.ProjectFilter{
			WorkspaceID: "ws-123",
			Page:        1,
			PageSize:    10,
		}

		projects := []*project.Project{
			{
				ID:            "proj-1",
				Name:          "project-1",
				WorkspaceID:   "ws-123",
				NamespaceName: "project-1",
			},
			{
				ID:            "proj-2",
				Name:          "project-2",
				WorkspaceID:   "ws-123",
				NamespaceName: "project-2",
			},
		}

		mockRepo.On("ListProjects", ctx, filter).Return(projects, 2, nil)
		mockK8sRepo.On("GetNamespace", ctx, "ws-123", "project-1").Return(map[string]interface{}{"status": "active"}, nil)
		mockK8sRepo.On("GetNamespace", ctx, "ws-123", "project-2").Return(map[string]interface{}{"status": "terminating"}, nil)

		// Execute
		result, err := service.ListProjects(ctx, filter)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Projects, 2)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, "active", result.Projects[0].Status)
		assert.Equal(t, "terminating", result.Projects[1].Status)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("empty project list", func(t *testing.T) {
		// Create new mock instances for this test
		mockRepoLocal := new(MockRepository)
		mockK8sRepoLocal := new(MockKubernetesRepository)
		serviceLocal := NewService(mockRepoLocal, mockK8sRepoLocal, logger)
		
		filter := project.ProjectFilter{
			WorkspaceID: "ws-123",
			Page:        1,
			PageSize:    10,
		}

		mockRepoLocal.On("ListProjects", ctx, filter).Return([]*project.Project{}, 0, nil)

		// Execute
		result, err := serviceLocal.ListProjects(ctx, filter)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Projects, 0)
		assert.Equal(t, 0, result.Total)
		mockRepoLocal.AssertExpectations(t)
	})
}

func TestUpdateProject(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful project update", func(t *testing.T) {
		existingProject := &project.Project{
			ID:          "proj-123",
			Name:        "test-project",
			DisplayName: "Old Name",
			Description: "Old description",
		}

		req := &project.UpdateProjectRequest{
			DisplayName: "New Name",
			Description: "New description",
			Settings: map[string]interface{}{
				"feature": "enabled",
			},
		}

		mockRepo.On("GetProject", ctx, "proj-123").Return(existingProject, nil)
		mockRepo.On("UpdateProject", ctx, mock.MatchedBy(func(p *project.Project) bool {
			return p.DisplayName == "New Name" &&
				p.Description == "New description" &&
				p.Settings["feature"] == "enabled"
		})).Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		result, err := service.UpdateProject(ctx, "proj-123", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New Name", result.DisplayName)
		assert.Equal(t, "New description", result.Description)
		mockRepo.AssertExpectations(t)
	})

	t.Run("project not found", func(t *testing.T) {
		req := &project.UpdateProjectRequest{
			DisplayName: "New Name",
		}

		mockRepo.On("GetProject", ctx, "proj-not-found").Return(nil, errors.New("not found"))

		// Execute
		result, err := service.UpdateProject(ctx, "proj-not-found", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get project")
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteProject(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful project deletion", func(t *testing.T) {
		proj := &project.Project{
			ID:          "proj-123",
			Name:        "test-project",
			WorkspaceID: "ws-123",
		}

		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepo.On("DeleteNamespace", ctx, "ws-123", "test-project").Return(nil)
		mockRepo.On("DeleteProject", ctx, "proj-123").Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		err := service.DeleteProject(ctx, "proj-123")

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("namespace deletion fails but project still deleted", func(t *testing.T) {
		proj := &project.Project{
			ID:          "proj-123",
			Name:        "test-project",
			WorkspaceID: "ws-123",
		}

		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepo.On("DeleteNamespace", ctx, "ws-123", "test-project").Return(errors.New("namespace deletion failed"))
		mockRepo.On("DeleteProject", ctx, "proj-123").Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		err := service.DeleteProject(ctx, "proj-123")

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})
}

func TestCreateSubProject(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful sub-project creation", func(t *testing.T) {
		parentProject := &project.Project{
			ID:          "parent-123",
			Name:        "parent-project",
			WorkspaceID: "ws-123",
		}

		req := &project.CreateProjectRequest{
			Name:        "sub-project",
			DisplayName: "Sub Project",
		}

		// Mock parent project retrieval
		mockRepo.On("GetProject", ctx, "parent-123").Return(parentProject, nil)
		
		// Mock sub-project creation (same as CreateProject)
		mockRepo.On("GetProjectByNameAndWorkspace", ctx, "sub-project", "ws-123").Return(nil, errors.New("not found"))
		mockRepo.On("CreateProject", ctx, mock.AnythingOfType("*project.Project")).Return(nil)
		mockK8sRepo.On("CreateNamespace", ctx, "ws-123", "sub-project", mock.Anything).Return(nil)
		mockRepo.On("UpdateProject", ctx, mock.AnythingOfType("*project.Project")).Return(nil).Times(2) // Once for namespace, once for parent relationship
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)
		
		// Mock HNC configuration
		mockK8sRepo.On("ConfigureHNC", ctx, "ws-123", "parent-project", "sub-project").Return(nil)

		// Execute
		result, err := service.CreateSubProject(ctx, "parent-123", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "sub-project", result.Name)
		assert.NotNil(t, result.ParentID)
		assert.Equal(t, "parent-123", *result.ParentID)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("parent project not found", func(t *testing.T) {
		req := &project.CreateProjectRequest{
			Name: "sub-project",
		}

		mockRepo.On("GetProject", ctx, "parent-not-found").Return(nil, errors.New("not found"))

		// Execute
		result, err := service.CreateSubProject(ctx, "parent-not-found", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "parent project not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestGetProjectHierarchy(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful hierarchy retrieval", func(t *testing.T) {
		rootProject := &project.Project{
			ID:   "root-123",
			Name: "root-project",
		}

		childProjects := []*project.Project{
			{
				ID:   "child-1",
				Name: "child-project-1",
			},
			{
				ID:   "child-2",
				Name: "child-project-2",
			},
		}

		grandchildProjects := []*project.Project{
			{
				ID:   "grandchild-1",
				Name: "grandchild-project-1",
			},
		}

		mockRepo.On("GetProject", ctx, "root-123").Return(rootProject, nil)
		mockRepo.On("GetChildProjects", ctx, "root-123").Return(childProjects, nil)
		mockRepo.On("GetChildProjects", ctx, "child-1").Return(grandchildProjects, nil)
		mockRepo.On("GetChildProjects", ctx, "child-2").Return([]*project.Project{}, nil)
		mockRepo.On("GetChildProjects", ctx, "grandchild-1").Return([]*project.Project{}, nil)

		// Execute
		result, err := service.GetProjectHierarchy(ctx, "root-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "root-123", result.Project.ID)
		assert.Len(t, result.Children, 2)
		assert.Len(t, result.Children[0].Children, 1)
		assert.Len(t, result.Children[1].Children, 0)
		mockRepo.AssertExpectations(t)
	})
}

func TestApplyResourceQuota(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful resource quota application", func(t *testing.T) {
		proj := &project.Project{
			ID:          "proj-123",
			Name:        "test-project",
			WorkspaceID: "ws-123",
		}

		quota := &project.ResourceQuota{
			CPU:     "4",
			Memory:  "8Gi",
			Storage: "100Gi",
			Pods:    50,
		}

		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepo.On("ApplyResourceQuota", ctx, "ws-123", "test-project", quota).Return(nil)
		mockRepo.On("UpdateProject", ctx, mock.MatchedBy(func(p *project.Project) bool {
			return p.Settings["resource_quota"] != nil
		})).Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		err := service.ApplyResourceQuota(ctx, "proj-123", quota)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})
}

func TestGetResourceUsage(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful resource usage retrieval", func(t *testing.T) {
		proj := &project.Project{
			ID:            "proj-123",
			Name:          "test-project",
			WorkspaceID:   "ws-123",
			NamespaceName: "test-namespace",
		}

		usage := &project.ResourceUsage{
			CPU:    "500m",
			Memory: "1Gi",
			Pods:   10,
		}

		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepo.On("GetNamespaceResourceUsage", ctx, "ws-123", "test-namespace").Return(usage, nil)

		// Execute
		result, err := service.GetResourceUsage(ctx, "proj-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "500m", result.CPU)
		assert.Equal(t, "1Gi", result.Memory)
		assert.Equal(t, 10, result.Pods)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})
}

func TestNamespaceManagement(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("create namespace successfully", func(t *testing.T) {
		proj := &project.Project{
			ID:          "proj-123",
			Name:        "test-project",
			WorkspaceID: "ws-123",
		}

		req := &project.CreateNamespaceRequest{
			Name:        "test-namespace",
			Description: "Test namespace",
			ResourceQuota: &project.ResourceQuota{
				CPU:    "2",
				Memory: "4Gi",
			},
		}

		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		
		labels := map[string]string{
			"project-id":   "proj-123",
			"workspace-id": "ws-123",
		}
		mockK8sRepo.On("CreateNamespace", ctx, "ws-123", "test-namespace", labels).Return(nil)
		mockK8sRepo.On("CreateResourceQuota", ctx, "ws-123", "test-namespace", req.ResourceQuota).Return(nil)
		mockRepo.On("CreateNamespace", ctx, mock.MatchedBy(func(ns *project.Namespace) bool {
			return ns.Name == "test-namespace" &&
				ns.ProjectID == "proj-123" &&
				ns.Description == "Test namespace"
		})).Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		result, err := service.CreateNamespace(ctx, "proj-123", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-namespace", result.Name)
		assert.Equal(t, "proj-123", result.ProjectID)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("list namespaces", func(t *testing.T) {
		namespaces := []*project.Namespace{
			{
				ID:        "ns-1",
				Name:      "namespace-1",
				ProjectID: "proj-123",
			},
			{
				ID:        "ns-2",
				Name:      "namespace-2",
				ProjectID: "proj-123",
			},
		}

		mockRepo.On("ListNamespaces", ctx, "proj-123").Return(namespaces, nil)

		// Execute
		result, err := service.ListNamespaces(ctx, "proj-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Namespaces, 2)
		assert.Equal(t, 2, result.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete namespace", func(t *testing.T) {
		namespace := &project.Namespace{
			ID:        "ns-123",
			Name:      "test-namespace",
			ProjectID: "proj-123",
		}

		proj := &project.Project{
			ID:          "proj-123",
			WorkspaceID: "ws-123",
		}

		mockRepo.On("GetNamespace", ctx, "ns-123").Return(namespace, nil)
		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepo.On("DeleteNamespace", ctx, "ws-123", "test-namespace").Return(nil)
		mockRepo.On("DeleteNamespace", ctx, "ns-123").Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		err := service.DeleteNamespace(ctx, "proj-123", "ns-123")

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})
}

func TestMemberManagement(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("add member successfully", func(t *testing.T) {
		proj := &project.Project{
			ID:            "proj-123",
			WorkspaceID:   "ws-123",
			NamespaceName: "test-project",
		}

		user := &project.User{
			ID:          "user-456",
			Email:       "user@example.com",
			DisplayName: "Test User",
		}

		req := &project.AddMemberRequest{
			UserEmail: "user@example.com",
			Role:      "developer",
		}

		existingMembers := []*project.ProjectMember{}

		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockRepo.On("GetUserByEmail", ctx, "user@example.com").Return(user, nil)
		mockRepo.On("ListMembers", ctx, "proj-123").Return(existingMembers, nil)
		mockRepo.On("AddMember", ctx, mock.MatchedBy(func(m *project.ProjectMember) bool {
			return m.ProjectID == "proj-123" &&
				m.UserID == "user-456" &&
				m.Role == "developer"
		})).Return(nil)
		mockK8sRepo.On("ApplyRBAC", ctx, "ws-123", "test-project", "user-456", "developer").Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		result, err := service.AddMember(ctx, "proj-123", "adder-123", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "user-456", result.UserID)
		assert.Equal(t, "developer", result.Role)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("add member - user already exists", func(t *testing.T) {
		// Create new mock instances for this test
		mockRepoLocal := new(MockRepository)
		mockK8sRepoLocal := new(MockKubernetesRepository)
		serviceLocal := NewService(mockRepoLocal, mockK8sRepoLocal, logger)
		
		proj := &project.Project{
			ID: "proj-123",
		}

		user := &project.User{
			ID:    "user-456",
			Email: "user@example.com",
		}

		req := &project.AddMemberRequest{
			UserEmail: "user@example.com",
			Role:      "developer",
		}

		existingMembers := []*project.ProjectMember{
			{
				UserID: "user-456",
			},
		}

		mockRepoLocal.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockRepoLocal.On("GetUserByEmail", ctx, "user@example.com").Return(user, nil)
		mockRepoLocal.On("ListMembers", ctx, "proj-123").Return(existingMembers, nil)

		// Execute
		result, err := serviceLocal.AddMember(ctx, "proj-123", "adder-123", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already a project member")
		mockRepoLocal.AssertExpectations(t)
	})

	t.Run("remove member successfully", func(t *testing.T) {
		member := &project.ProjectMember{
			ID:        "member-123",
			ProjectID: "proj-123",
			UserID:    "user-456",
			UserEmail: "user@example.com",
		}

		proj := &project.Project{
			ID:            "proj-123",
			WorkspaceID:   "ws-123",
			NamespaceName: "test-project",
		}

		mockRepo.On("GetMemberByID", ctx, "member-123").Return(member, nil)
		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockRepo.On("RemoveMember", ctx, "member-123").Return(nil)
		mockK8sRepo.On("RemoveRBAC", ctx, "ws-123", "test-project", "user-456").Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		err := service.RemoveMember(ctx, "proj-123", "member-123", "remover-123")

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("update member role", func(t *testing.T) {
		member := &project.ProjectMember{
			ID:        "member-123",
			ProjectID: "proj-123",
			UserID:    "user-456",
			Role:      "viewer",
		}

		proj := &project.Project{
			ID:            "proj-123",
			WorkspaceID:   "ws-123",
			NamespaceName: "test-project",
		}

		req := &project.UpdateMemberRoleRequest{
			Role: "admin",
		}

		mockRepo.On("GetMemberByID", ctx, "member-123").Return(member, nil)
		mockRepo.On("UpdateMember", ctx, mock.MatchedBy(func(m *project.ProjectMember) bool {
			return m.Role == "admin"
		})).Return(nil)
		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepo.On("ApplyRBAC", ctx, "ws-123", "test-project", "user-456", "admin").Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		result, err := service.UpdateMemberRole(ctx, "proj-123", "member-123", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "admin", result.Role)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})
}

func TestActivityManagement(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("list activities", func(t *testing.T) {
		activities := []*project.ProjectActivity{
			{
				ID:        "act-1",
				ProjectID: "proj-123",
				Type:      "project_created",
			},
			{
				ID:        "act-2",
				ProjectID: "proj-123",
				Type:      "member_added",
			},
		}

		mockRepo.On("ListActivities", ctx, mock.MatchedBy(func(filter project.ActivityFilter) bool {
			return filter.ProjectID == "proj-123" && filter.PageSize == 10
		})).Return(activities, nil)

		// Execute
		result, err := service.ListActivities(ctx, "proj-123", 10)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Activities, 2)
		assert.Equal(t, 2, result.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("log activity", func(t *testing.T) {
		activity := &project.ProjectActivity{
			ProjectID:   "proj-123",
			Type:        "custom_event",
			Description: "Custom event occurred",
		}

		mockRepo.On("CreateActivity", ctx, mock.MatchedBy(func(a *project.ProjectActivity) bool {
			return a.ProjectID == "proj-123" &&
				a.Type == "custom_event" &&
				a.ID != "" &&
				!a.CreatedAt.IsZero()
		})).Return(nil)

		// Execute
		err := service.LogActivity(ctx, activity)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestProjectStats(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("get project stats", func(t *testing.T) {
		proj := &project.Project{
			ID: "proj-123",
			Settings: map[string]interface{}{
				"namespaces": []interface{}{"ns-1", "ns-2", "ns-3"},
			},
		}

		usage := &project.ResourceUsage{
			CPU:    "2",
			Memory: "4Gi",
			Pods:   20,
		}

		lastActivity := &project.ProjectActivity{
			CreatedAt: time.Now().Add(-1 * time.Hour),
		}

		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockRepo.On("CountMembers", ctx, "proj-123").Return(5, nil)
		mockRepo.On("GetProjectResourceUsage", ctx, "proj-123").Return(usage, nil)
		mockRepo.On("GetLastActivity", ctx, "proj-123").Return(lastActivity, nil)

		// Execute
		result, err := service.GetProjectStats(ctx, "proj-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "proj-123", result.ProjectID)
		assert.Equal(t, 3, result.NamespaceCount)
		assert.Equal(t, 5, result.MemberCount)
		assert.Equal(t, usage, result.ResourceUsage)
		assert.NotNil(t, result.LastActivity)
		mockRepo.AssertExpectations(t)
	})
}

func TestAccessControl(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("validate project access - allowed", func(t *testing.T) {
		member := &project.ProjectMember{
			UserID: "user-123",
			Role:   "admin",
		}

		mockRepo.On("GetMember", ctx, "proj-123", "user-123").Return(member, nil)

		// Execute
		err := service.ValidateProjectAccess(ctx, "user-123", "proj-123", "developer")

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validate project access - denied (not member)", func(t *testing.T) {
		// Create new mock instances for this test
		mockRepoLocal := new(MockRepository)
		mockK8sRepoLocal := new(MockKubernetesRepository)
		serviceLocal := NewService(mockRepoLocal, mockK8sRepoLocal, logger)
		
		mockRepoLocal.On("GetMember", ctx, "proj-123", "user-123").Return(nil, errors.New("not found"))

		// Execute
		err := serviceLocal.ValidateProjectAccess(ctx, "user-123", "proj-123", "viewer")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user is not a project member")
		mockRepoLocal.AssertExpectations(t)
	})

	t.Run("validate project access - denied (insufficient role)", func(t *testing.T) {
		// Create new mock instances for this test
		mockRepoLocal := new(MockRepository)
		mockK8sRepoLocal := new(MockKubernetesRepository)
		serviceLocal := NewService(mockRepoLocal, mockK8sRepoLocal, logger)
		
		member := &project.ProjectMember{
			UserID: "user-123",
			Role:   "viewer",
		}

		mockRepoLocal.On("GetMember", ctx, "proj-123", "user-123").Return(member, nil)

		// Execute
		err := serviceLocal.ValidateProjectAccess(ctx, "user-123", "proj-123", "admin")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient permissions")
		mockRepoLocal.AssertExpectations(t)
	})

	t.Run("get user project role", func(t *testing.T) {
		// Create new mock instances for this test
		mockRepoLocal := new(MockRepository)
		mockK8sRepoLocal := new(MockKubernetesRepository)
		serviceLocal := NewService(mockRepoLocal, mockK8sRepoLocal, logger)
		
		member := &project.ProjectMember{
			UserID: "user-123",
			Role:   "developer",
		}

		mockRepoLocal.On("GetMember", ctx, "proj-123", "user-123").Return(member, nil)

		// Execute
		role, err := serviceLocal.GetUserProjectRole(ctx, "user-123", "proj-123")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "developer", role)
		mockRepoLocal.AssertExpectations(t)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("isValidProjectName", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected bool
		}{
			{"valid lowercase", "myproject", true},
			{"valid with hyphen", "my-project", true},
			{"valid with numbers", "project123", true},
			{"valid complex", "my-project-123", true},
			{"invalid uppercase", "MyProject", false},
			{"invalid special char", "my_project", false},
			{"invalid starts with hyphen", "-project", false},
			{"invalid ends with hyphen", "project-", false},
			{"invalid too long", "this-is-a-very-long-project-name-that-exceeds-the-maximum-allowed-length", false},
			{"invalid empty", "", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := isValidProjectName(tc.input)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("hasRequiredRole", func(t *testing.T) {
		testCases := []struct {
			name         string
			userRole     string
			requiredRole string
			expected     bool
		}{
			{"admin has admin access", "admin", "admin", true},
			{"admin has developer access", "admin", "developer", true},
			{"admin has viewer access", "admin", "viewer", true},
			{"developer has developer access", "developer", "developer", true},
			{"developer has viewer access", "developer", "viewer", true},
			{"developer lacks admin access", "developer", "admin", false},
			{"viewer has viewer access", "viewer", "viewer", true},
			{"viewer lacks developer access", "viewer", "developer", false},
			{"viewer lacks admin access", "viewer", "admin", false},
			{"invalid user role", "invalid", "viewer", false},
			{"invalid required role", "admin", "invalid", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := hasRequiredRole(tc.userRole, tc.requiredRole)
				assert.Equal(t, tc.expected, result)
			})
		}
	})
}

// Test edge cases and error scenarios
func TestEdgeCases(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("create project with nil settings", func(t *testing.T) {
		req := &project.CreateProjectRequest{
			Name:        "test-project",
			WorkspaceID: "ws-123",
			Settings:    nil,
		}

		mockRepo.On("GetProjectByNameAndWorkspace", ctx, "test-project", "ws-123").Return(nil, errors.New("not found"))
		mockRepo.On("CreateProject", ctx, mock.AnythingOfType("*project.Project")).Return(nil)
		mockK8sRepo.On("CreateNamespace", ctx, "ws-123", "test-project", mock.Anything).Return(nil)
		mockRepo.On("UpdateProject", ctx, mock.AnythingOfType("*project.Project")).Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		proj, err := service.CreateProject(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, proj)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update namespace with wrong project ID", func(t *testing.T) {
		namespace := &project.Namespace{
			ID:        "ns-123",
			ProjectID: "proj-456",
		}

		req := &project.CreateNamespaceRequest{
			Description: "Updated description",
		}

		mockRepo.On("GetNamespace", ctx, "ns-123").Return(namespace, nil)

		// Execute
		result, err := service.UpdateNamespace(ctx, "proj-123", "ns-123", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "namespace does not belong to project")
		mockRepo.AssertExpectations(t)
	})

	t.Run("get member with wrong project ID", func(t *testing.T) {
		member := &project.ProjectMember{
			ID:        "member-123",
			ProjectID: "proj-456",
		}

		mockRepo.On("GetMemberByID", ctx, "member-123").Return(member, nil)

		// Execute
		result, err := service.GetMember(ctx, "proj-123", "member-123")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "member does not belong to project")
		mockRepo.AssertExpectations(t)
	})
}

// Additional tests for uncovered methods
func TestGetNamespace(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful namespace retrieval", func(t *testing.T) {
		namespace := &project.Namespace{
			ID:        "ns-123",
			Name:      "test-namespace",
			ProjectID: "proj-123",
		}

		proj := &project.Project{
			ID:          "proj-123",
			WorkspaceID: "ws-123",
		}

		usage := &project.NamespaceUsage{
			CPU:     "100m",
			Memory:  "256Mi",
			Storage: "1Gi",
			Pods:    3,
		}

		mockRepo.On("GetNamespace", ctx, "ns-123").Return(namespace, nil)
		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepo.On("GetNamespaceUsage", ctx, "ws-123", "test-namespace").Return(usage, nil)

		// Execute
		result, err := service.GetNamespace(ctx, "proj-123", "ns-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ns-123", result.ID)
		assert.Equal(t, usage, result.ResourceUsage)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("namespace belongs to different project", func(t *testing.T) {
		// Create new mock instances for this test
		mockRepoLocal := new(MockRepository)
		mockK8sRepoLocal := new(MockKubernetesRepository)
		serviceLocal := NewService(mockRepoLocal, mockK8sRepoLocal, logger)
		
		namespace := &project.Namespace{
			ID:        "ns-123",
			ProjectID: "proj-456",
		}

		mockRepoLocal.On("GetNamespace", ctx, "ns-123").Return(namespace, nil)

		// Execute
		result, err := serviceLocal.GetNamespace(ctx, "proj-123", "ns-123")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "namespace does not belong to project")
		mockRepoLocal.AssertExpectations(t)
	})
}

func TestGetNamespaceUsage(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful usage retrieval", func(t *testing.T) {
		namespace := &project.Namespace{
			ID:        "ns-123",
			Name:      "test-namespace",
			ProjectID: "proj-123",
		}

		proj := &project.Project{
			ID:          "proj-123",
			WorkspaceID: "ws-123",
		}

		usage := &project.NamespaceUsage{
			CPU:     "200m",
			Memory:  "512Mi",
			Storage: "5Gi",
			Pods:    7,
		}

		mockRepo.On("GetNamespace", ctx, "ns-123").Return(namespace, nil)
		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockK8sRepo.On("GetNamespaceUsage", ctx, "ws-123", "test-namespace").Return(usage, nil)

		// Execute
		result, err := service.GetNamespaceUsage(ctx, "proj-123", "ns-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, usage, result)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})
}

func TestListMembers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful member listing", func(t *testing.T) {
		members := []*project.ProjectMember{
			{
				ID:        "member-1",
				ProjectID: "proj-123",
				UserID:    "user-1",
				Role:      "admin",
			},
			{
				ID:        "member-2",
				ProjectID: "proj-123",
				UserID:    "user-2",
				Role:      "developer",
			},
		}

		mockRepo.On("ListMembers", ctx, "proj-123").Return(members, nil)

		// Execute
		result, err := service.ListMembers(ctx, "proj-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Members, 2)
		assert.Equal(t, 2, result.Total)
		mockRepo.AssertExpectations(t)
	})
}

func TestProjectMemberAliases(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("AddProjectMember", func(t *testing.T) {
		proj := &project.Project{
			ID:            "proj-123",
			WorkspaceID:   "ws-123",
			NamespaceName: "test-project",
		}

		user := &project.User{
			ID:          "user-456",
			Email:       "newuser@example.com",
			DisplayName: "New User",
		}

		req := &project.AddMemberRequest{
			UserEmail: "newuser@example.com",
			Role:      "viewer",
		}

		mockRepo.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockRepo.On("GetUserByEmail", ctx, "newuser@example.com").Return(user, nil)
		mockRepo.On("ListMembers", ctx, "proj-123").Return([]*project.ProjectMember{}, nil)
		mockRepo.On("AddMember", ctx, mock.AnythingOfType("*project.ProjectMember")).Return(nil)
		mockK8sRepo.On("ApplyRBAC", ctx, "ws-123", "test-project", "user-456", "viewer").Return(nil)
		mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		err := service.AddProjectMember(ctx, "proj-123", req)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockK8sRepo.AssertExpectations(t)
	})

	t.Run("RemoveProjectMember", func(t *testing.T) {
		// Create new mock instances for this test
		mockRepoLocal := new(MockRepository)
		mockK8sRepoLocal := new(MockKubernetesRepository)
		serviceLocal := NewService(mockRepoLocal, mockK8sRepoLocal, logger)
		
		proj := &project.Project{
			ID:          "proj-123",
			Name:        "test-project",
			WorkspaceID: "ws-123",
		}

		member := &project.ProjectMember{
			ID:        "member-123",
			ProjectID: "proj-123",
			UserID:    "user-456",
		}

		mockRepoLocal.On("GetProject", ctx, "proj-123").Return(proj, nil)
		mockRepoLocal.On("GetMember", ctx, "proj-123", "user-456").Return(member, nil)
		mockRepoLocal.On("RemoveMember", ctx, "member-123").Return(nil)
		mockK8sRepoLocal.On("RemoveRBAC", ctx, "ws-123", "test-project", "user-456").Return(nil)
		mockRepoLocal.On("CreateActivity", ctx, mock.AnythingOfType("*project.ProjectActivity")).Return(nil)

		// Execute
		err := serviceLocal.RemoveProjectMember(ctx, "proj-123", "user-456")

		// Assert
		assert.NoError(t, err)
		mockRepoLocal.AssertExpectations(t)
		mockK8sRepoLocal.AssertExpectations(t)
	})

	t.Run("ListProjectMembers", func(t *testing.T) {
		// Create new mock instances for this test
		mockRepoLocal := new(MockRepository)
		mockK8sRepoLocal := new(MockKubernetesRepository)
		serviceLocal := NewService(mockRepoLocal, mockK8sRepoLocal, logger)
		
		members := []*project.ProjectMember{
			{
				ID:     "member-1",
				UserID: "user-1",
			},
		}

		mockRepoLocal.On("ListMembers", ctx, "proj-123").Return(members, nil)

		// Execute
		result, err := serviceLocal.ListProjectMembers(ctx, "proj-123")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepoLocal.AssertExpectations(t)
	})
}

func TestGetActivityLogs(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8sRepo := new(MockKubernetesRepository)
	logger := createTestLogger()
	
	service := NewService(mockRepo, mockK8sRepo, logger)

	t.Run("successful activity logs retrieval", func(t *testing.T) {
		activities := []*project.ProjectActivity{
			{
				ID:        "act-1",
				ProjectID: "proj-123",
				Type:      "member_added",
			},
			{
				ID:        "act-2",
				ProjectID: "proj-123",
				Type:      "project_updated",
			},
		}

		filter := project.ActivityFilter{
			Type: "member_added",
		}

		mockRepo.On("ListActivities", ctx, mock.MatchedBy(func(f project.ActivityFilter) bool {
			return f.ProjectID == "proj-123" && f.Type == "member_added"
		})).Return(activities, nil)

		// Execute
		result, err := service.GetActivityLogs(ctx, "proj-123", filter)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})
}