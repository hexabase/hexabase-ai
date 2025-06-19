package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks are now defined in mocks_test.go

// Test service
func TestCreateApplication(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	
	service := NewService(mockRepo, mockK8s)

	t.Run("successful stateless application creation", func(t *testing.T) {
		req := domain.CreateApplicationRequest{
			Name: "test-app",
			Type: domain.ApplicationTypeStateless,
			Source: domain.ApplicationSource{
				Type:  domain.SourceTypeImage,
				Image: "nginx:latest",
			},
			Config: domain.ApplicationConfig{
				Replicas: 3,
				Port:     80,
				EnvVars: map[string]string{
					"ENV": "production",
				},
				Resources: domain.ResourceRequests{
					CPURequest:    "100m",
					CPULimit:      "500m",
					MemoryRequest: "128Mi",
					MemoryLimit:   "512Mi",
				},
			},
			ProjectID: "proj-123",
		}

		// Mock expectations
		mockRepo.On("GetApplicationByName", ctx, "ws-123", "proj-123", "test-app").Return(nil, errors.New("not found"))
		mockRepo.On("CreateApplication", ctx, mock.AnythingOfType("*domain.Application")).Return(nil)
		mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.ApplicationEvent")).Return(nil)
		
		// Async deployment expectations
		mockK8s.On("CreateDeployment", mock.Anything, "ws-123", "proj-123", mock.MatchedBy(func(spec domain.DeploymentSpec) bool {
			return spec.Name == "test-app" && spec.Replicas == 3 && spec.Image == "nginx:latest"
		})).Return(nil).Maybe()
		
		mockK8s.On("CreateService", mock.Anything, "ws-123", "proj-123", mock.MatchedBy(func(spec domain.ServiceSpec) bool {
			return spec.Name == "test-app" && spec.Port == 80
		})).Return(nil).Maybe()
		
		mockK8s.On("GetServiceEndpoints", mock.Anything, "ws-123", "proj-123", "test-app").Return([]domain.Endpoint{}, nil).Maybe()
		
		mockRepo.On("UpdateApplication", mock.Anything, mock.AnythingOfType("*domain.Application")).Return(nil).Maybe()

		// Execute
		app, err := service.CreateApplication(ctx, "ws-123", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, app)
		assert.Equal(t, "test-app", app.Name)
		assert.Equal(t, domain.ApplicationTypeStateless, app.Type)
		// Don't check status as it's being modified in a goroutine
		
		// Wait a bit for async operations to complete
		time.Sleep(100 * time.Millisecond)
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("successful stateful application creation", func(t *testing.T) {
		req := domain.CreateApplicationRequest{
			Name: "postgres-db",
			Type: domain.ApplicationTypeStateful,
			Source: domain.ApplicationSource{
				Type:  domain.SourceTypeImage,
				Image: "postgres:14",
			},
			Config: domain.ApplicationConfig{
				Replicas: 1,
				Port:     5432,
				EnvVars: map[string]string{
					"POSTGRES_DB": "mydb",
				},
				Resources: domain.ResourceRequests{
					CPURequest:    "250m",
					CPULimit:      "1000m",
					MemoryRequest: "512Mi",
					MemoryLimit:   "2Gi",
				},
				Storage: &domain.StorageConfig{
					Size:         "10Gi",
					StorageClass: "standard",
					MountPath:    "/var/lib/postgresql/data",
				},
			},
			ProjectID: "proj-123",
		}

		// Mock expectations
		mockRepo.On("GetApplicationByName", ctx, "ws-123", "proj-123", "postgres-db").Return(nil, errors.New("not found"))
		mockRepo.On("CreateApplication", ctx, mock.AnythingOfType("*domain.Application")).Return(nil)
		mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.ApplicationEvent")).Return(nil)
		
		// Async deployment expectations
		mockK8s.On("CreatePVC", mock.Anything, "ws-123", "proj-123", mock.MatchedBy(func(spec domain.PVCSpec) bool {
			return spec.Name == "postgres-db-data" && spec.Size == "10Gi"
		})).Return(nil).Maybe()
		
		mockK8s.On("CreateStatefulSet", mock.Anything, "ws-123", "proj-123", mock.MatchedBy(func(spec domain.StatefulSetSpec) bool {
			return spec.Name == "postgres-db" && spec.Replicas == 1 && spec.Image == "postgres:14"
		})).Return(nil).Maybe()
		
		mockK8s.On("CreateService", mock.Anything, "ws-123", "proj-123", mock.MatchedBy(func(spec domain.ServiceSpec) bool {
			return spec.Name == "postgres-db" && spec.Port == 5432
		})).Return(nil).Maybe()
		
		mockK8s.On("GetServiceEndpoints", mock.Anything, "ws-123", "proj-123", "postgres-db").Return([]domain.Endpoint{}, nil).Maybe()
		
		mockRepo.On("UpdateApplication", mock.Anything, mock.AnythingOfType("*domain.Application")).Return(nil).Maybe()

		// Execute
		app, err := service.CreateApplication(ctx, "ws-123", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, app)
		assert.Equal(t, "postgres-db", app.Name)
		assert.Equal(t, domain.ApplicationTypeStateful, app.Type)
		// Don't check status as it's being modified in a goroutine
		
		// Wait a bit for async operations to complete
		time.Sleep(100 * time.Millisecond)
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("application name already exists", func(t *testing.T) {
		req := domain.CreateApplicationRequest{
			Name:      "existing-app",
			Type:      domain.ApplicationTypeStateless,
			Source: domain.ApplicationSource{
				Type:  domain.SourceTypeImage,
				Image: "nginx:latest",
			},
			ProjectID: "proj-123",
		}

		existingApp := &domain.Application{
			ID:   "app-existing",
			Name: "existing-app",
		}
		mockRepo.On("GetApplicationByName", ctx, "ws-123", "proj-123", "existing-app").Return(existingApp, nil)

		// Execute
		app, err := service.CreateApplication(ctx, "ws-123", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, app)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid application type", func(t *testing.T) {
		req := domain.CreateApplicationRequest{
			Name:      "invalid-app",
			Type:      domain.ApplicationType("invalid"),
			ProjectID: "proj-123",
		}

		// Execute
		app, err := service.CreateApplication(ctx, "ws-123", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, app)
		assert.Contains(t, err.Error(), "invalid application type")
	})

	t.Run("invalid source type", func(t *testing.T) {
		req := domain.CreateApplicationRequest{
			Name: "invalid-source",
			Type: domain.ApplicationTypeStateless,
			Source: domain.ApplicationSource{
				Type: domain.SourceType("invalid"),
			},
			ProjectID: "proj-123",
		}

		// Execute
		app, err := service.CreateApplication(ctx, "ws-123", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, app)
		assert.Contains(t, err.Error(), "invalid source type")
	})
}

func TestUpdateApplication(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	
	service := NewService(mockRepo, mockK8s)

	t.Run("successful update replicas", func(t *testing.T) {
		existingApp := &domain.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
			ProjectID:   "proj-123",
			Name:        "test-app",
			Type:        domain.ApplicationTypeStateless,
			Status:      domain.ApplicationStatusRunning,
			Config: domain.ApplicationConfig{
				Replicas: 3,
			},
		}

		replicas := 5
		req := domain.UpdateApplicationRequest{
			Replicas: &replicas,
		}

		mockRepo.On("GetApplication", ctx, "app-123").Return(existingApp, nil)
		mockRepo.On("UpdateApplication", ctx, mock.AnythingOfType("*domain.Application")).Return(nil)
		mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.ApplicationEvent")).Return(nil)
		
		// Async update expectations
		mockK8s.On("UpdateDeployment", mock.Anything, "ws-123", "proj-123", "test-app", mock.MatchedBy(func(spec domain.DeploymentSpec) bool {
			return spec.Replicas == 5
		})).Return(nil).Maybe()
		mockRepo.On("UpdateApplication", mock.Anything, mock.AnythingOfType("*domain.Application")).Return(nil).Maybe()
		mockRepo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*domain.ApplicationEvent")).Return(nil).Maybe()

		// Execute
		app, err := service.UpdateApplication(ctx, "app-123", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, app)
		assert.Equal(t, 5, app.Config.Replicas)
		// Don't check status as it's being modified in a goroutine
		
		// Wait a bit for async operations to complete
		time.Sleep(100 * time.Millisecond)
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("application not found", func(t *testing.T) {
		req := domain.UpdateApplicationRequest{}
		mockRepo.On("GetApplication", ctx, "app-not-found").Return(nil, errors.New("not found"))

		// Execute
		app, err := service.UpdateApplication(ctx, "app-not-found", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, app)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteApplication(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	
	service := NewService(mockRepo, mockK8s)

	t.Run("successful deletion of stateless app", func(t *testing.T) {
		app := &domain.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
			ProjectID:   "proj-123",
			Name:        "test-app",
			Type:        domain.ApplicationTypeStateless,
			Status:      domain.ApplicationStatusStopped,
		}

		mockRepo.On("GetApplication", ctx, "app-123").Return(app, nil)
		mockRepo.On("UpdateApplication", ctx, mock.AnythingOfType("*domain.Application")).Return(nil)
		mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.ApplicationEvent")).Return(nil)
		mockK8s.On("DeleteDeployment", ctx, "ws-123", "proj-123", "test-app").Return(nil).Maybe()
		mockK8s.On("DeleteService", ctx, "ws-123", "proj-123", "test-app").Return(nil).Maybe()
		mockK8s.On("DeleteIngress", ctx, "ws-123", "proj-123", "test-app").Return(nil).Maybe()
		mockRepo.On("DeleteApplication", ctx, "app-123").Return(nil)

		// Execute
		err := service.DeleteApplication(ctx, "app-123")

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("successful deletion of stateful app", func(t *testing.T) {
		app := &domain.Application{
			ID:          "app-456",
			WorkspaceID: "ws-123",
			ProjectID:   "proj-123",
			Name:        "postgres-db",
			Type:        domain.ApplicationTypeStateful,
			Status:      domain.ApplicationStatusStopped,
			Config: domain.ApplicationConfig{
				Storage: &domain.StorageConfig{
					Size: "10Gi",
				},
			},
		}

		mockRepo.On("GetApplication", ctx, "app-456").Return(app, nil)
		mockRepo.On("UpdateApplication", ctx, mock.AnythingOfType("*domain.Application")).Return(nil)
		mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.ApplicationEvent")).Return(nil)
		mockK8s.On("DeleteStatefulSet", ctx, "ws-123", "proj-123", "postgres-db").Return(nil).Maybe()
		mockK8s.On("DeleteService", ctx, "ws-123", "proj-123", "postgres-db").Return(nil).Maybe()
		mockK8s.On("DeletePVC", ctx, "ws-123", "proj-123", "postgres-db-data").Return(nil).Maybe()
		mockK8s.On("DeleteIngress", ctx, "ws-123", "proj-123", "postgres-db").Return(nil).Maybe()
		mockRepo.On("DeleteApplication", ctx, "app-456").Return(nil)

		// Execute
		err := service.DeleteApplication(ctx, "app-456")

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestListPods(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	
	service := NewService(mockRepo, mockK8s)

	t.Run("successful pod listing", func(t *testing.T) {
		app := &domain.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
			ProjectID:   "proj-123",
			Name:        "test-app",
		}

		pods := []domain.Pod{
			{
				Name:     "test-app-abc123",
				Status:   "Running",
				NodeName: "node-1",
				IP:       "10.0.0.1",
			},
			{
				Name:     "test-app-def456",
				Status:   "Running",
				NodeName: "node-2",
				IP:       "10.0.0.2",
			},
		}

		mockRepo.On("GetApplication", ctx, "app-123").Return(app, nil)
		mockK8s.On("ListPods", ctx, "ws-123", "proj-123", map[string]string{"app": "test-app"}).Return(pods, nil)

		// Execute
		result, err := service.ListPods(ctx, "app-123")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "test-app-abc123", result[0].Name)
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestGetPodLogs(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	
	service := NewService(mockRepo, mockK8s)

	t.Run("successful log retrieval", func(t *testing.T) {
		app := &domain.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
			ProjectID:   "proj-123",
			Name:        "test-app",
		}

		query := domain.LogQuery{
			ApplicationID: "app-123",
			PodName:       "test-app-abc123",
			Limit:         100,
		}

		logs := []domain.LogEntry{
			{
				Timestamp: time.Now(),
				PodName:   "test-app-abc123",
				Message:   "Application started",
			},
			{
				Timestamp: time.Now(),
				PodName:   "test-app-abc123",
				Message:   "Listening on port 80",
			},
		}

		mockRepo.On("GetApplication", ctx, "app-123").Return(app, nil)
		
		logOpts := domain.LogOptions{
			Limit: 100,
		}
		mockK8s.On("GetPodLogs", ctx, "ws-123", "proj-123", "test-app-abc123", "", logOpts).Return(logs, nil)

		// Execute
		result, err := service.GetPodLogs(ctx, query)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Application started", result[0].Message)
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestStreamPodLogs(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	
	service := NewService(mockRepo, mockK8s)

	t.Run("successful log streaming", func(t *testing.T) {
		app := &domain.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
			ProjectID:   "proj-123",
			Name:        "test-app",
		}

		query := domain.LogQuery{
			ApplicationID: "app-123",
			PodName:       "test-app-abc123",
			Follow:        true,
		}

		reader := io.NopCloser(strings.NewReader("log stream data"))

		mockRepo.On("GetApplication", ctx, "app-123").Return(app, nil)
		
		logOpts := domain.LogOptions{
			Follow: true,
		}
		mockK8s.On("StreamPodLogs", ctx, "ws-123", "proj-123", "test-app-abc123", "", logOpts).Return(reader, nil)

		// Execute
		result, err := service.StreamPodLogs(ctx, query)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestGetApplicationMetrics(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	
	service := NewService(mockRepo, mockK8s)

	t.Run("successful metrics retrieval", func(t *testing.T) {
		app := &domain.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
			ProjectID:   "proj-123",
			Name:        "test-app",
		}

		pods := []domain.Pod{
			{Name: "test-app-abc123"},
			{Name: "test-app-def456"},
		}

		podMetrics := []domain.PodMetrics{
			{
				PodName:     "test-app-abc123",
				CPUUsage:    0.5,
				MemoryUsage: 256.0,
				NetworkIn:   1.5,
				NetworkOut:  2.0,
			},
			{
				PodName:     "test-app-def456",
				CPUUsage:    0.3,
				MemoryUsage: 200.0,
				NetworkIn:   1.0,
				NetworkOut:  1.5,
			},
		}

		mockRepo.On("GetApplication", ctx, "app-123").Return(app, nil)
		mockK8s.On("ListPods", ctx, "ws-123", "proj-123", map[string]string{"app": "test-app"}).Return(pods, nil)
		mockK8s.On("GetPodMetrics", ctx, "ws-123", "proj-123", []string{"test-app-abc123", "test-app-def456"}).Return(podMetrics, nil)

		// Execute
		result, err := service.GetApplicationMetrics(ctx, "app-123")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "app-123", result.ApplicationID)
		assert.Len(t, result.PodMetrics, 2)
		assert.Equal(t, 0.8, result.AggregateUsage.TotalCPU)
		assert.Equal(t, 0.4, result.AggregateUsage.AverageCPU)
		assert.Equal(t, 456.0, result.AggregateUsage.TotalMemory)
		assert.Equal(t, 228.0, result.AggregateUsage.AverageMemory)
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestScaleApplication(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockK8s := new(MockKubernetesRepository)
	
	service := NewService(mockRepo, mockK8s)

	t.Run("successful scaling", func(t *testing.T) {
		app := &domain.Application{
			ID:          "app-123",
			WorkspaceID: "ws-123",
			ProjectID:   "proj-123",
			Name:        "test-app",
			Type:        domain.ApplicationTypeStateless,
			Status:      domain.ApplicationStatusRunning,
			Config: domain.ApplicationConfig{
				Replicas: 3,
			},
		}

		mockRepo.On("GetApplication", ctx, "app-123").Return(app, nil)
		mockRepo.On("UpdateApplication", ctx, mock.AnythingOfType("*domain.Application")).Return(nil)
		mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.ApplicationEvent")).Return(nil)
		
		// Async update expectations
		mockK8s.On("UpdateDeployment", mock.Anything, "ws-123", "proj-123", "test-app", mock.MatchedBy(func(spec domain.DeploymentSpec) bool {
			return spec.Replicas == 5
		})).Return(nil).Maybe()
		mockRepo.On("UpdateApplication", mock.Anything, mock.AnythingOfType("*domain.Application")).Return(nil).Maybe()
		mockRepo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*domain.ApplicationEvent")).Return(nil).Maybe()

		// Execute
		err := service.ScaleApplication(ctx, "app-123", 5)

		// Assert
		assert.NoError(t, err)
		
		// Wait a bit for async operations to complete
		time.Sleep(100 * time.Millisecond)
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("invalid replica count", func(t *testing.T) {
		// Execute
		err := service.ScaleApplication(ctx, "app-123", -1)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid replica count")
	})

	t.Run("scaling stateful set not allowed beyond 1", func(t *testing.T) {
		app := &domain.Application{
			ID:     "app-456",
			Type:   domain.ApplicationTypeStateful,
			Status: domain.ApplicationStatusRunning,
		}

		mockRepo.On("GetApplication", ctx, "app-456").Return(app, nil)

		// Execute
		err := service.ScaleApplication(ctx, "app-456", 3)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot scale beyond 1 replica")
		mockRepo.AssertExpectations(t)
	})
}