package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateWorkspace(t *testing.T) {
	ctx := context.Background()

	t.Run("successful workspace creation", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		req := &domain.CreateWorkspaceRequest{
			Name:           "test-workspace",
			Description:    "Test workspace description",
			OrganizationID: "org-123",
			Plan:           domain.WorkspacePlanShared,
			PlanID:         "plan-shared-1",
			Settings: map[string]interface{}{
				"env": "test",
			},
		}

		mockRepo.On("CreateWorkspace", ctx, mock.AnythingOfType("*domain.Workspace")).Return(nil)
		mockRepo.On("CreateTask", ctx, mock.AnythingOfType("*domain.Task")).Return(nil)

		// Execute
		ws, task, err := service.CreateWorkspace(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, ws)
		assert.NotNil(t, task)
		assert.Equal(t, "test-workspace", ws.Name)
		assert.Equal(t, "org-123", ws.OrganizationID)
		assert.Equal(t, domain.WorkspacePlanShared, ws.Plan)
		assert.Equal(t, "creating", ws.Status)
		assert.Equal(t, "provision_vcluster", task.Type)
		assert.Equal(t, "pending", task.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("workspace name is required", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		req := &domain.CreateWorkspaceRequest{
			OrganizationID: "org-123",
			Plan:           domain.WorkspacePlanShared,
		}

		// Execute
		ws, task, err := service.CreateWorkspace(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, ws)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "workspace name is required")
	})

	t.Run("repository error during workspace creation", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		req := &domain.CreateWorkspaceRequest{
			Name:           "test-workspace",
			OrganizationID: "org-123",
			Plan:           domain.WorkspacePlanShared,
		}

		mockRepo.On("CreateWorkspace", ctx, mock.AnythingOfType("*domain.Workspace")).Return(errors.New("database error"))

		// Execute
		ws, task, err := service.CreateWorkspace(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, ws)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "failed to create workspace")

		mockRepo.AssertExpectations(t)
	})

	t.Run("task creation failure is logged but doesn't fail workspace creation", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		req := &domain.CreateWorkspaceRequest{
			Name:           "test-workspace",
			OrganizationID: "org-123",
			Plan:           domain.WorkspacePlanShared,
		}

		mockRepo.On("CreateWorkspace", ctx, mock.AnythingOfType("*domain.Workspace")).Return(nil)
		mockRepo.On("CreateTask", ctx, mock.AnythingOfType("*domain.Task")).Return(errors.New("task creation error"))

		// Execute
		ws, task, err := service.CreateWorkspace(ctx, req)

		// Assert
		assert.NoError(t, err) // Workspace creation should still succeed
		assert.NotNil(t, ws)
		assert.NotNil(t, task) // Task is still returned even if save failed

		mockRepo.AssertExpectations(t)
	})
}

func TestGetWorkspace(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	mockAuth := new(MockAuthRepository)
	mockHelm := new(MockHelmService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

	t.Run("successful get workspace - active status", func(t *testing.T) {
		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:             workspaceID,
			Name:           "test-workspace",
			OrganizationID: "org-123",
			Status:         "active",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockK8s.On("GetVClusterStatus", ctx, workspaceID).Return("running", nil)

		// Execute
		result, err := service.GetWorkspace(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, workspaceID, result.ID)
		assert.NotNil(t, result.ClusterInfo)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("successful get workspace - non-active status", func(t *testing.T) {
		workspaceID := "ws-456"
		ws := &domain.Workspace{
			ID:             workspaceID,
			Name:           "test-workspace",
			OrganizationID: "org-123",
			Status:         "creating",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)

		// Execute
		result, err := service.GetWorkspace(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, workspaceID, result.ID)
		assert.Equal(t, "creating", result.Status)

		mockRepo.AssertExpectations(t)
		// mockK8s should not be called for non-active workspaces
	})

	t.Run("vcluster status error is logged but doesn't fail", func(t *testing.T) {
		workspaceID := "ws-789"
		ws := &domain.Workspace{
			ID:             workspaceID,
			Name:           "test-workspace",
			OrganizationID: "org-123",
			Status:         "active",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockK8s.On("GetVClusterStatus", ctx, workspaceID).Return("", errors.New("vcluster error"))

		// Execute
		result, err := service.GetWorkspace(ctx, workspaceID)

		// Assert
		assert.NoError(t, err) // Should not fail despite vcluster error
		assert.NotNil(t, result)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		workspaceID := "ws-not-found"

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Execute
		result, err := service.GetWorkspace(ctx, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get workspace")

		mockRepo.AssertExpectations(t)
	})
}

func TestListWorkspaces(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	mockAuth := new(MockAuthRepository)
	mockHelm := new(MockHelmService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

	t.Run("successful list workspaces", func(t *testing.T) {
		filter := domain.WorkspaceFilter{
			OrganizationID: "org-123",
			Page:           1,
			PageSize:       10,
		}

		workspaces := []*domain.Workspace{
			{
				ID:             "ws-1",
				Name:           "workspace-1",
				OrganizationID: "org-123",
			},
			{
				ID:             "ws-2",
				Name:           "workspace-2",
				OrganizationID: "org-123",
			},
		}

		mockRepo.On("ListWorkspaces", ctx, filter).Return(workspaces, 2, nil)

		// Execute
		result, err := service.ListWorkspaces(ctx, filter)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Workspaces, 2)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.PageSize)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		filter := domain.WorkspaceFilter{
			OrganizationID: "org-123",
		}

		mockRepo.On("ListWorkspaces", ctx, filter).Return(nil, 0, errors.New("database error"))

		// Execute
		result, err := service.ListWorkspaces(ctx, filter)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateWorkspace(t *testing.T) {
	ctx := context.Background()

	t.Run("successful workspace update", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		workspaceID := "ws-123"
		existingWs := &domain.Workspace{
			ID:             workspaceID,
			Name:           "old-name",
			Description:    "old-description",
			OrganizationID: "org-123",
			Settings: map[string]interface{}{
				"old": "value",
			},
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		req := &domain.UpdateWorkspaceRequest{
			Name:        "new-name",
			Description: "new-description",
			Settings: map[string]interface{}{
				"new": "value",
			},
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(existingWs, nil)
		mockRepo.On("UpdateWorkspace", ctx, mock.AnythingOfType("*domain.Workspace")).Return(nil)

		// Execute
		result, err := service.UpdateWorkspace(ctx, workspaceID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "new-name", result.Name)
		assert.Equal(t, "new-description", result.Description)
		assert.Equal(t, "value", result.Settings["new"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-not-found"
		req := &domain.UpdateWorkspaceRequest{
			Name: "new-name",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Execute
		result, err := service.UpdateWorkspace(ctx, workspaceID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get workspace")

		mockRepo.AssertExpectations(t)
	})

	t.Run("update repository error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"
		existingWs := &domain.Workspace{
			ID:             workspaceID,
			Name:           "old-name",
			OrganizationID: "org-123",
		}

		req := &domain.UpdateWorkspaceRequest{
			Name: "new-name",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(existingWs, nil)
		mockRepo.On("UpdateWorkspace", ctx, mock.AnythingOfType("*domain.Workspace")).Return(errors.New("update error"))

		// Execute
		result, err := service.UpdateWorkspace(ctx, workspaceID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to update workspace")

		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteWorkspace(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	mockAuth := new(MockAuthRepository)
	mockHelm := new(MockHelmService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

	t.Run("successful workspace deletion", func(t *testing.T) {
		workspaceID := "ws-123"
		existingWs := &domain.Workspace{
			ID:             workspaceID,
			Name:           "test-workspace",
			OrganizationID: "org-123",
			Status:         "active",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(existingWs, nil)
		mockRepo.On("UpdateWorkspace", ctx, mock.MatchedBy(func(ws *domain.Workspace) bool {
			return ws.Status == "deleting"
		})).Return(nil)
		mockRepo.On("CreateTask", ctx, mock.AnythingOfType("*domain.Task")).Return(nil)

		// Execute
		task, err := service.DeleteWorkspace(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, "delete_vcluster", task.Type)
		assert.Equal(t, "pending", task.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("workspace already being deleted", func(t *testing.T) {
		workspaceID := "ws-123"
		existingWs := &domain.Workspace{
			ID:             workspaceID,
			Name:           "test-workspace",
			OrganizationID: "org-123",
			Status:         "deleting",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(existingWs, nil)

		// Execute
		task, err := service.DeleteWorkspace(ctx, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "workspace is already being deleted")

		mockRepo.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		workspaceID := "ws-not-found"

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Execute
		task, err := service.DeleteWorkspace(ctx, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "failed to get workspace")

		mockRepo.AssertExpectations(t)
	})
}

func TestGetWorkspaceStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("successful status retrieval", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:             workspaceID,
			Name:           "test-workspace",
			OrganizationID: "org-123",
			Status:         "active",
		}

		resourceUsage := &domain.ResourceUsage{
			WorkspaceID: workspaceID,
			CPU: domain.ResourceMetric{
				Used:      0.5,
				Requested: 1.0,
				Limit:     2.0,
				Unit:      "cores",
			},
		}

		clusterInfo := &domain.ClusterInfo{
			Endpoint:  "https://example.com",
			APIServer: "api.example.com",
			Status:    "running",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockK8s.On("GetVClusterStatus", ctx, workspaceID).Return("running", nil)
		mockK8s.On("GetResourceMetrics", ctx, workspaceID).Return(resourceUsage, nil)
		mockK8s.On("GetVClusterInfo", ctx, workspaceID).Return(clusterInfo, nil)
		mockRepo.On("SaveWorkspaceStatus", ctx, mock.AnythingOfType("*domain.WorkspaceStatus")).Return(nil)

		// Execute
		status, err := service.GetWorkspaceStatus(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.Equal(t, workspaceID, status.WorkspaceID)
		assert.Equal(t, "active", status.Status)
		assert.True(t, status.Healthy)
		assert.Equal(t, "vCluster is running", status.Message)
		assert.NotNil(t, status.ResourceUsage)
		assert.NotNil(t, status.ClusterInfo)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-not-found"

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Execute
		status, err := service.GetWorkspaceStatus(ctx, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, status)
		assert.Contains(t, err.Error(), "workspace not found")

		mockRepo.AssertExpectations(t)
	})

	t.Run("vcluster errors are handled gracefully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:             workspaceID,
			Status:         "active",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockK8s.On("GetVClusterStatus", ctx, workspaceID).Return("unknown", errors.New("vcluster error"))
		mockK8s.On("GetResourceMetrics", ctx, workspaceID).Return(nil, errors.New("metrics error"))
		mockK8s.On("GetVClusterInfo", ctx, workspaceID).Return(nil, errors.New("info error"))
		mockRepo.On("SaveWorkspaceStatus", ctx, mock.AnythingOfType("*domain.WorkspaceStatus")).Return(nil)

		// Execute
		status, err := service.GetWorkspaceStatus(ctx, workspaceID)

		// Assert
		assert.NoError(t, err) // Should not fail despite errors
		assert.NotNil(t, status)
		assert.False(t, status.Healthy)
		assert.Equal(t, "vCluster is unknown", status.Message)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestExecuteOperation(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	mockAuth := new(MockAuthRepository)
	mockHelm := new(MockHelmService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

	t.Run("successful backup operation", func(t *testing.T) {
		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:   workspaceID,
			Name: "test-workspace",
		}

		req := &domain.WorkspaceOperationRequest{
			Operation: "backup",
			Metadata: map[string]interface{}{
				"backup_name": "test-backup",
			},
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockRepo.On("CreateTask", ctx, mock.AnythingOfType("*domain.Task")).Return(nil)

		// Execute
		task, err := service.ExecuteOperation(ctx, workspaceID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, "backup", task.Type)
		assert.Equal(t, "pending", task.Status)
		assert.Contains(t, task.Message, "Creating backup")

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful restore operation", func(t *testing.T) {
		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:   workspaceID,
			Name: "test-workspace",
		}

		req := &domain.WorkspaceOperationRequest{
			Operation: "restore",
			Metadata: map[string]interface{}{
				"backup_id": "backup-123",
			},
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockRepo.On("CreateTask", ctx, mock.AnythingOfType("*domain.Task")).Return(nil)

		// Execute
		task, err := service.ExecuteOperation(ctx, workspaceID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, "restore", task.Type)
		assert.Equal(t, "pending", task.Status)
		assert.Contains(t, task.Message, "Restoring workspace")
		assert.Equal(t, "backup-123", task.Payload["backup_id"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful upgrade operation", func(t *testing.T) {
		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:   workspaceID,
			Name: "test-workspace",
		}

		req := &domain.WorkspaceOperationRequest{
			Operation: "upgrade",
			Metadata: map[string]interface{}{
				"target_version": "v1.25.0",
			},
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockRepo.On("CreateTask", ctx, mock.AnythingOfType("*domain.Task")).Return(nil)

		// Execute
		task, err := service.ExecuteOperation(ctx, workspaceID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, "upgrade", task.Type)
		assert.Equal(t, "pending", task.Status)
		assert.Contains(t, task.Message, "Upgrading workspace")
		assert.Equal(t, "v1.25.0", task.Payload["target_version"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("unsupported operation", func(t *testing.T) {
		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:   workspaceID,
			Name: "test-workspace",
		}

		req := &domain.WorkspaceOperationRequest{
			Operation: "invalid-operation",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)

		// Execute
		task, err := service.ExecuteOperation(ctx, workspaceID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "unsupported operation")

		mockRepo.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		workspaceID := "ws-not-found"
		req := &domain.WorkspaceOperationRequest{
			Operation: "backup",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Execute
		task, err := service.ExecuteOperation(ctx, workspaceID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "workspace not found")

		mockRepo.AssertExpectations(t)
	})
}

func TestGetKubeconfig(t *testing.T) {
	ctx := context.Background()

	t.Run("successful kubeconfig retrieval from repository", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:     workspaceID,
			Status: "active",
		}
		expectedKubeconfig := "kubeconfig-content"

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockRepo.On("GetKubeconfig", ctx, workspaceID).Return(expectedKubeconfig, nil)

		// Execute
		kubeconfig, err := service.GetKubeconfig(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedKubeconfig, kubeconfig)

		mockRepo.AssertExpectations(t)
	})

	t.Run("kubeconfig retrieved from vcluster when not in repository", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:     workspaceID,
			Status: "active",
		}
		expectedKubeconfig := "vcluster-kubeconfig-content"
		clusterInfo := &domain.ClusterInfo{
			KubeConfig: expectedKubeconfig,
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockRepo.On("GetKubeconfig", ctx, workspaceID).Return("", errors.New("not found"))
		mockK8s.On("GetVClusterInfo", ctx, workspaceID).Return(clusterInfo, nil)
		mockRepo.On("SaveKubeconfig", ctx, workspaceID, expectedKubeconfig).Return(nil)

		// Execute
		kubeconfig, err := service.GetKubeconfig(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedKubeconfig, kubeconfig)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("workspace not active", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID:     workspaceID,
			Status: "creating",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)

		// Execute
		kubeconfig, err := service.GetKubeconfig(ctx, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, kubeconfig)
		assert.Contains(t, err.Error(), "workspace is not active")

		mockRepo.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-not-found"

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Execute
		kubeconfig, err := service.GetKubeconfig(ctx, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, kubeconfig)
		assert.Contains(t, err.Error(), "workspace not found")

		mockRepo.AssertExpectations(t)
	})
}

func TestAddWorkspaceMember(t *testing.T) {
	ctx := context.Background()

	t.Run("successful member addition", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID: workspaceID,
		}
		existingMembers := []*domain.WorkspaceMember{}

		req := &domain.AddMemberRequest{
			UserID:  "user-456",
			Role:    "editor",
			AddedBy: "user-123",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockRepo.On("ListWorkspaceMembers", ctx, workspaceID).Return(existingMembers, nil)
		mockRepo.On("AddWorkspaceMember", ctx, mock.AnythingOfType("*domain.WorkspaceMember")).Return(nil)
		mockK8s.On("UpdateOIDCConfig", ctx, workspaceID, mock.AnythingOfType("map[string]interface {}")).Return(nil)

		// Execute
		member, err := service.AddWorkspaceMember(ctx, workspaceID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, member)
		assert.Equal(t, "user-456", member.UserID)
		assert.Equal(t, "editor", member.Role)
		assert.Equal(t, "user-123", member.AddedBy)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("user already a member", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"
		ws := &domain.Workspace{
			ID: workspaceID,
		}
		existingMembers := []*domain.WorkspaceMember{
			{
				UserID: "user-456",
				Role:   "viewer",
			},
		}

		req := &domain.AddMemberRequest{
			UserID:  "user-456",
			Role:    "editor",
			AddedBy: "user-123",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockRepo.On("ListWorkspaceMembers", ctx, workspaceID).Return(existingMembers, nil)

		// Execute
		member, err := service.AddWorkspaceMember(ctx, workspaceID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, member)
		assert.Contains(t, err.Error(), "user is already a member")

		mockRepo.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-not-found"
		req := &domain.AddMemberRequest{
			UserID: "user-456",
			Role:   "editor",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Execute
		member, err := service.AddWorkspaceMember(ctx, workspaceID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, member)
		assert.Contains(t, err.Error(), "workspace not found")

		mockRepo.AssertExpectations(t)
	})
}

func TestRemoveWorkspaceMember(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	mockAuth := new(MockAuthRepository)
	mockHelm := new(MockHelmService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

	t.Run("successful member removal", func(t *testing.T) {
		workspaceID := "ws-123"
		userID := "user-456"
		existingMembers := []*domain.WorkspaceMember{
			{
				UserID: "user-456",
				Role:   "editor",
			},
			{
				UserID: "user-789",
				Role:   "viewer",
			},
		}

		mockRepo.On("ListWorkspaceMembers", ctx, workspaceID).Return(existingMembers, nil)
		mockRepo.On("RemoveWorkspaceMember", ctx, workspaceID, userID).Return(nil)
		mockK8s.On("UpdateOIDCConfig", ctx, workspaceID, mock.MatchedBy(func(config map[string]interface{}) bool {
			users := config["users"].([]string)
			return len(users) == 1 && users[0] == "user-789"
		})).Return(nil)

		// Execute
		err := service.RemoveWorkspaceMember(ctx, workspaceID, userID)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("user is not a member", func(t *testing.T) {
		workspaceID := "ws-123"
		userID := "user-not-member"
		existingMembers := []*domain.WorkspaceMember{
			{
				UserID: "user-456",
				Role:   "editor",
			},
		}

		mockRepo.On("ListWorkspaceMembers", ctx, workspaceID).Return(existingMembers, nil)

		// Execute
		err := service.RemoveWorkspaceMember(ctx, workspaceID, userID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user is not a member")

		mockRepo.AssertExpectations(t)
	})
}

func TestValidateWorkspaceAccess(t *testing.T) {
	ctx := context.Background()

	t.Run("user has access", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		userID := "user-123"
		workspaceID := "ws-456"
		members := []*domain.WorkspaceMember{
			{
				UserID: "user-123",
				Role:   "editor",
			},
		}

		mockRepo.On("ListWorkspaceMembers", ctx, workspaceID).Return(members, nil)

		// Execute
		err := service.ValidateWorkspaceAccess(ctx, userID, workspaceID)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user does not have access", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		userID := "user-123"
		workspaceID := "ws-456"
		members := []*domain.WorkspaceMember{
			{
				UserID: "user-789",
				Role:   "editor",
			},
		}

		mockRepo.On("ListWorkspaceMembers", ctx, workspaceID).Return(members, nil)

		// Execute
		err := service.ValidateWorkspaceAccess(ctx, userID, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user does not have access")

		mockRepo.AssertExpectations(t)
	})
}

func TestGetNodes(t *testing.T) {
	ctx := context.Background()

	t.Run("successful node listing", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		workspaceID := "ws-123"
		expectedNodes := []domain.Node{
			{
				Name:   "node-1",
				Status: "Ready",
				CPU:    "2",
				Memory: "4Gi",
				Pods:   10,
			},
			{
				Name:   "node-2",
				Status: "Ready",
				CPU:    "4",
				Memory: "8Gi",
				Pods:   15,
			},
		}

		mockK8s.On("ListVClusterNodes", ctx, workspaceID).Return(expectedNodes, nil)

		// Execute
		nodes, err := service.GetNodes(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, nodes, 2)
		assert.Equal(t, "node-1", nodes[0].Name)
		assert.Equal(t, "Ready", nodes[0].Status)

		mockK8s.AssertExpectations(t)
	})

	t.Run("kubernetes repository error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"

		mockK8s.On("ListVClusterNodes", ctx, workspaceID).Return(nil, errors.New("k8s error"))

		// Execute
		nodes, err := service.GetNodes(ctx, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, nodes)

		mockK8s.AssertExpectations(t)
	})
}

func TestScaleDeployment(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deployment scaling", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		workspaceID := "ws-123"
		deploymentName := "my-app"
		replicas := 5
		ws := &domain.Workspace{
			ID:     workspaceID,
			Status: "active",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockK8s.On("ScaleVClusterDeployment", ctx, workspaceID, deploymentName, replicas).Return(nil)

		// Execute
		err := service.ScaleDeployment(ctx, workspaceID, deploymentName, replicas)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("workspace not active", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"
		deploymentName := "my-app"
		replicas := 5
		ws := &domain.Workspace{
			ID:     workspaceID,
			Status: "creating",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)

		// Execute
		err := service.ScaleDeployment(ctx, workspaceID, deploymentName, replicas)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workspace is not active")

		mockRepo.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-not-found"
		deploymentName := "my-app"
		replicas := 5

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Execute
		err := service.ScaleDeployment(ctx, workspaceID, deploymentName, replicas)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspace")

		mockRepo.AssertExpectations(t)
	})

	t.Run("kubernetes error during scaling", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"
		deploymentName := "my-app"
		replicas := 5
		ws := &domain.Workspace{
			ID:     workspaceID,
			Status: "active",
		}

		mockRepo.On("GetWorkspace", ctx, workspaceID).Return(ws, nil)
		mockK8s.On("ScaleVClusterDeployment", ctx, workspaceID, deploymentName, replicas).Return(errors.New("scaling error"))

		// Execute
		err := service.ScaleDeployment(ctx, workspaceID, deploymentName, replicas)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scale deployment")

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestGetResourceUsage(t *testing.T) {
	ctx := context.Background()

	t.Run("successful resource usage retrieval", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		workspaceID := "ws-123"
		expectedUsage := &domain.ResourceUsage{
			CPU: domain.ResourceMetric{
				Used:      0.5,
				Requested: 1.0,
				Limit:     2.0,
				Unit:      "cores",
			},
			Memory: domain.ResourceMetric{
				Used:      512.0,
				Requested: 1024.0,
				Limit:     2048.0,
				Unit:      "MB",
			},
		}

		mockK8s.On("GetResourceMetrics", ctx, workspaceID).Return(expectedUsage, nil)
		mockRepo.On("CreateResourceUsage", ctx, mock.AnythingOfType("*domain.ResourceUsage")).Return(nil)

		// Execute
		usage, err := service.GetResourceUsage(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, usage)
		assert.Equal(t, workspaceID, usage.WorkspaceID)
		assert.Equal(t, 0.5, usage.CPU.Used)
		assert.Equal(t, 512.0, usage.Memory.Used)

		mockK8s.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("kubernetes error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"

		mockK8s.On("GetResourceMetrics", ctx, workspaceID).Return(nil, errors.New("metrics error"))

		// Execute
		usage, err := service.GetResourceUsage(ctx, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, usage)
		assert.Contains(t, err.Error(), "failed to get resource metrics")

		mockK8s.AssertExpectations(t)
	})
}

func TestGetTask(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	mockAuth := new(MockAuthRepository)
	mockHelm := new(MockHelmService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

	t.Run("successful task retrieval", func(t *testing.T) {
		taskID := "task-123"
		expectedTask := &domain.Task{
			ID:          taskID,
			WorkspaceID: "ws-123",
			Type:        "provision_vcluster",
			Status:      "completed",
		}

		mockRepo.On("GetTask", ctx, taskID).Return(expectedTask, nil)

		// Execute
		task, err := service.GetTask(ctx, taskID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, taskID, task.ID)
		assert.Equal(t, "provision_vcluster", task.Type)

		mockRepo.AssertExpectations(t)
	})

	t.Run("task not found", func(t *testing.T) {
		taskID := "task-not-found"

		mockRepo.On("GetTask", ctx, taskID).Return(nil, errors.New("not found"))

		// Execute
		task, err := service.GetTask(ctx, taskID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, task)

		mockRepo.AssertExpectations(t)
	})
}

func TestListTasks(t *testing.T) {
	ctx := context.Background()

	t.Run("successful task listing", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"
		expectedTasks := []*domain.Task{
			{
				ID:          "task-1",
				WorkspaceID: workspaceID,
				Type:        "provision_vcluster",
				Status:      "completed",
			},
			{
				ID:          "task-2",
				WorkspaceID: workspaceID,
				Type:        "backup",
				Status:      "pending",
			},
		}

		mockRepo.On("ListTasks", ctx, workspaceID).Return(expectedTasks, nil)

		// Execute
		tasks, err := service.ListTasks(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, "task-1", tasks[0].ID)
		assert.Equal(t, "task-2", tasks[1].ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		workspaceID := "ws-123"

		mockRepo.On("ListTasks", ctx, workspaceID).Return(nil, errors.New("database error"))

		// Execute
		tasks, err := service.ListTasks(ctx, workspaceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tasks)

		mockRepo.AssertExpectations(t)
	})
}

func TestListWorkspaceMembers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	mockAuth := new(MockAuthRepository)
	mockHelm := new(MockHelmService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

	t.Run("successful member listing", func(t *testing.T) {
		workspaceID := "ws-123"
		expectedMembers := []*domain.WorkspaceMember{
			{
				ID:          "member-1",
				WorkspaceID: workspaceID,
				UserID:      "user-1",
				Role:        "admin",
			},
			{
				ID:          "member-2",
				WorkspaceID: workspaceID,
				UserID:      "user-2",
				Role:        "editor",
			},
		}

		mockRepo.On("ListWorkspaceMembers", ctx, workspaceID).Return(expectedMembers, nil)

		// Execute
		members, err := service.ListWorkspaceMembers(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, members, 2)
		assert.Equal(t, "user-1", members[0].UserID)
		assert.Equal(t, "admin", members[0].Role)

		mockRepo.AssertExpectations(t)
	})
}

func TestProcessProvisioningTask(t *testing.T) {
	ctx := context.Background()

	t.Run("successful provisioning task processing", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		taskID := "task-123"
		task := &domain.Task{
			ID:          taskID,
			WorkspaceID: "ws-123",
			Type:        "create",
			Status:      "pending",
		}

		ws := &domain.Workspace{
			ID:   "ws-123",
			Plan: domain.WorkspacePlanShared,
		}

		mockRepo.On("GetTask", ctx, taskID).Return(task, nil)
		mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.Status == "running"
		})).Return(nil)
		mockRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)
		
		// VCluster provisioning expectations
		mockK8s.On("CreateVCluster", ctx, "ws-123", domain.WorkspacePlanShared).Return(nil)
		mockK8s.On("WaitForVClusterReady", ctx, "ws-123").Return(nil)
		mockK8s.On("ConfigureOIDC", ctx, "ws-123").Return(nil)
		mockK8s.On("ApplyResourceQuotas", ctx, "ws-123", domain.WorkspacePlanShared).Return(nil)
		
		// Helm deployment for shared plan
		mockHelm.On("InstallOrUpgrade", "hks-observability-agents", "./deployments/helm/hks-observability-agents", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(nil)
		
		mockRepo.On("UpdateWorkspace", ctx, mock.MatchedBy(func(ws *domain.Workspace) bool {
			return ws.Status == "active"
		})).Return(nil)
		
		mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.Status == "completed" && t.Progress == 100
		})).Return(nil)

		// Execute
		err := service.ProcessProvisioningTask(ctx, taskID)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
	})

	t.Run("task not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		taskID := "task-not-found"

		mockRepo.On("GetTask", ctx, taskID).Return(nil, errors.New("not found"))

		// Execute
		err := service.ProcessProvisioningTask(ctx, taskID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task not found")

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid task type", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		taskID := "task-123"
		task := &domain.Task{
			ID:   taskID,
			Type: "invalid-type",
		}

		mockRepo.On("GetTask", ctx, taskID).Return(task, nil)

		// Execute
		err := service.ProcessProvisioningTask(ctx, taskID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task type for provisioning")

		mockRepo.AssertExpectations(t)
	})

	t.Run("provisioning failure", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		taskID := "task-123"
		task := &domain.Task{
			ID:          taskID,
			WorkspaceID: "ws-123",
			Type:        "create",
			Status:      "pending",
		}

		ws := &domain.Workspace{
			ID:   "ws-123",
			Plan: domain.WorkspacePlanShared,
		}

		mockRepo.On("GetTask", ctx, taskID).Return(task, nil)
		mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.Status == "running"
		})).Return(nil)
		mockRepo.On("GetWorkspace", ctx, "ws-123").Return(ws, nil)
		mockK8s.On("CreateVCluster", ctx, "ws-123", domain.WorkspacePlanShared).Return(errors.New("vcluster creation failed"))
		
		// Failure update expectations
		mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.Status == "failed" && t.Error != ""
		})).Return(nil)

		// Execute
		err := service.ProcessProvisioningTask(ctx, taskID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vcluster creation failed")

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestProcessDeletionTask(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion task processing", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)
		taskID := "task-123"
		task := &domain.Task{
			ID:          taskID,
			WorkspaceID: "ws-123",
			Type:        "delete",
			Status:      "pending",
		}

		mockRepo.On("GetTask", ctx, taskID).Return(task, nil)
		mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.Status == "running"
		})).Return(nil)
		
		// VCluster deletion expectations
		mockK8s.On("DeleteVCluster", ctx, "ws-123").Return(nil)
		mockK8s.On("WaitForVClusterDeleted", ctx, "ws-123").Return(nil)
		mockRepo.On("DeleteWorkspace", ctx, "ws-123").Return(nil)
		
		mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.Status == "completed" && t.Progress == 100
		})).Return(nil)

		// Execute
		err := service.ProcessDeletionTask(ctx, taskID)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("task not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		taskID := "task-not-found"

		mockRepo.On("GetTask", ctx, taskID).Return(nil, errors.New("not found"))

		// Execute
		err := service.ProcessDeletionTask(ctx, taskID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task not found")

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid task type", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		taskID := "task-123"
		task := &domain.Task{
			ID:   taskID,
			Type: "invalid-type",
		}

		mockRepo.On("GetTask", ctx, taskID).Return(task, nil)

		// Execute
		err := service.ProcessDeletionTask(ctx, taskID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task type for deletion")

		mockRepo.AssertExpectations(t)
	})

	t.Run("deletion failure", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		mockAuth := new(MockAuthRepository)
		mockHelm := new(MockHelmService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		service := NewService(mockRepo, mockK8s, mockAuth, mockHelm, logger)

		taskID := "task-123"
		task := &domain.Task{
			ID:          taskID,
			WorkspaceID: "ws-123",
			Type:        "delete",
			Status:      "pending",
		}

		mockRepo.On("GetTask", ctx, taskID).Return(task, nil)
		mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.Status == "running"
		})).Return(nil)
		mockK8s.On("DeleteVCluster", ctx, "ws-123").Return(errors.New("deletion failed"))
		
		// Failure update expectations
		mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.Status == "failed" && t.Error != ""
		})).Return(nil)

		// Execute
		err := service.ProcessDeletionTask(ctx, taskID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deletion failed")

		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

// Note: ProcessTask is a private method in the service implementation
// We test the public interface methods that handle task processing instead