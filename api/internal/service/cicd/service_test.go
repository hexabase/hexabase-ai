package cicd

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/cicd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
)

// Mock implementations

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreatePipeline(ctx context.Context, pipeline *cicd.Pipeline) error {
	args := m.Called(ctx, pipeline)
	return args.Error(0)
}

func (m *MockRepository) GetPipeline(ctx context.Context, id string) (*cicd.Pipeline, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.Pipeline), args.Error(1)
}

func (m *MockRepository) UpdatePipeline(ctx context.Context, pipeline *cicd.Pipeline) error {
	args := m.Called(ctx, pipeline)
	return args.Error(0)
}

func (m *MockRepository) DeletePipeline(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListPipelines(ctx context.Context, workspaceID, projectID string, limit, offset int) ([]*cicd.Pipeline, error) {
	args := m.Called(ctx, workspaceID, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cicd.Pipeline), args.Error(1)
}

func (m *MockRepository) CreatePipelineRun(ctx context.Context, run *cicd.PipelineRunRecord) error {
	args := m.Called(ctx, run)
	return args.Error(0)
}

func (m *MockRepository) GetPipelineRun(ctx context.Context, id string) (*cicd.PipelineRunRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.PipelineRunRecord), args.Error(1)
}

func (m *MockRepository) UpdatePipelineRun(ctx context.Context, run *cicd.PipelineRunRecord) error {
	args := m.Called(ctx, run)
	return args.Error(0)
}

func (m *MockRepository) ListTemplates(ctx context.Context, provider string) ([]*cicd.PipelineTemplate, error) {
	args := m.Called(ctx, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cicd.PipelineTemplate), args.Error(1)
}

func (m *MockRepository) GetTemplate(ctx context.Context, id string) (*cicd.PipelineTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.PipelineTemplate), args.Error(1)
}

func (m *MockRepository) GetProviderConfig(ctx context.Context, workspaceID string) (*cicd.WorkspaceProviderConfig, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.WorkspaceProviderConfig), args.Error(1)
}

func (m *MockRepository) SetProviderConfig(ctx context.Context, config *cicd.WorkspaceProviderConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockRepository) CreateTemplate(ctx context.Context, template *cicd.PipelineTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockRepository) UpdateTemplate(ctx context.Context, template *cicd.PipelineTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockRepository) DeleteTemplate(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetPipelineByRunID(ctx context.Context, runID string) (*cicd.Pipeline, error) {
	args := m.Called(ctx, runID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.Pipeline), args.Error(1)
}

func (m *MockRepository) ListPipelineRuns(ctx context.Context, workspaceID string, limit, offset int) ([]*cicd.PipelineRunRecord, error) {
	args := m.Called(ctx, workspaceID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cicd.PipelineRunRecord), args.Error(1)
}

type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) RunPipeline(ctx context.Context, config *cicd.PipelineConfig) (*cicd.PipelineRun, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.PipelineRun), args.Error(1)
}

func (m *MockProvider) GetStatus(ctx context.Context, workspaceID, runID string) (*cicd.PipelineRun, error) {
	args := m.Called(ctx, workspaceID, runID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.PipelineRun), args.Error(1)
}

func (m *MockProvider) CancelPipeline(ctx context.Context, workspaceID, runID string) error {
	args := m.Called(ctx, workspaceID, runID)
	return args.Error(0)
}

func (m *MockProvider) DeletePipeline(ctx context.Context, workspaceID, runID string) error {
	args := m.Called(ctx, workspaceID, runID)
	return args.Error(0)
}

func (m *MockProvider) GetLogs(ctx context.Context, runID, stage string) ([]cicd.LogEntry, error) {
	args := m.Called(ctx, runID, stage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]cicd.LogEntry), args.Error(1)
}

func (m *MockProvider) StreamLogs(ctx context.Context, runID, stage string) (io.ReadCloser, error) {
	args := m.Called(ctx, runID, stage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockProvider) ValidateConfig(ctx context.Context, config *cicd.PipelineConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockProvider) ListPipelines(ctx context.Context, workspaceID, projectID string) ([]*cicd.PipelineRun, error) {
	args := m.Called(ctx, workspaceID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cicd.PipelineRun), args.Error(1)
}

func (m *MockProvider) GetVersion() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProvider) GetTemplates(ctx context.Context) ([]*cicd.PipelineTemplate, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cicd.PipelineTemplate), args.Error(1)
}

func (m *MockProvider) CreateFromTemplate(ctx context.Context, templateID string, params map[string]any) (*cicd.PipelineConfig, error) {
	args := m.Called(ctx, templateID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.PipelineConfig), args.Error(1)
}

type MockProviderFactory struct {
	mock.Mock
}

func (m *MockProviderFactory) CreateProvider(providerType string, config *cicd.ProviderConfig) (cicd.Provider, error) {
	args := m.Called(providerType, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(cicd.Provider), args.Error(1)
}

func (m *MockProviderFactory) ListProviders() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

type MockCredentialManager struct {
	mock.Mock
}

func (m *MockCredentialManager) StoreGitCredential(workspaceID string, cred *cicd.GitCredential) (*cicd.CredentialInfo, error) {
	args := m.Called(workspaceID, cred)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.CredentialInfo), args.Error(1)
}

func (m *MockCredentialManager) StoreRegistryCredential(workspaceID string, cred *cicd.RegistryCredential) (*cicd.CredentialInfo, error) {
	args := m.Called(workspaceID, cred)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.CredentialInfo), args.Error(1)
}

func (m *MockCredentialManager) CreateKubernetesSecret(workspaceID, secretName string, data map[string][]byte) error {
	args := m.Called(workspaceID, secretName, data)
	return args.Error(0)
}

func (m *MockCredentialManager) DeleteKubernetesSecret(workspaceID, secretName string) error {
	args := m.Called(workspaceID, secretName)
	return args.Error(0)
}

func (m *MockCredentialManager) GetCredentialRef(workspaceID, name string) (string, error) {
	args := m.Called(workspaceID, name)
	return args.String(0), args.Error(1)
}

func (m *MockCredentialManager) DeleteCredential(workspaceID, name string) error {
	args := m.Called(workspaceID, name)
	return args.Error(0)
}

func (m *MockCredentialManager) ListCredentials(workspaceID string) ([]*cicd.CredentialInfo, error) {
	args := m.Called(workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cicd.CredentialInfo), args.Error(1)
}

func (m *MockCredentialManager) GetGitCredential(workspaceID, name string) (*cicd.GitCredential, error) {
	args := m.Called(workspaceID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.GitCredential), args.Error(1)
}

func (m *MockCredentialManager) GetRegistryCredential(workspaceID, name string) (*cicd.RegistryCredential, error) {
	args := m.Called(workspaceID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cicd.RegistryCredential), args.Error(1)
}

func (m *MockCredentialManager) GetKubernetesSecret(workspaceID, secretName string) (map[string][]byte, error) {
	args := m.Called(workspaceID, secretName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string][]byte), args.Error(1)
}

// Test helper functions

func createTestService() (*Service, *MockRepository, *MockProviderFactory, *MockCredentialManager) {
	repo := &MockRepository{}
	factory := &MockProviderFactory{}
	credManager := &MockCredentialManager{}
	logger := slog.Default()

	service := NewService(repo, factory, credManager, logger).(*Service)
	return service, repo, factory, credManager
}

func createTestPipelineConfig() cicd.PipelineConfig {
	return cicd.PipelineConfig{
		Name:      "test-pipeline",
		ProjectID: "project-123",
		GitRepo: cicd.GitConfig{
			URL:    "https://github.com/test/repo.git",
			Branch: "main",
		},
		BuildConfig: &cicd.BuildConfig{
			Type:           cicd.BuildTypeDocker,
			DockerfilePath: "Dockerfile",
			BuildContext:   ".",
		},
		Metadata: map[string]any{"ENV": "test"},
	}
}

// Tests

func TestCreatePipeline(t *testing.T) {
	t.Run("successful pipeline creation", func(t *testing.T) {
		service, repo, factory, _ := createTestService()
		ctx := context.Background()
		workspaceID := "workspace-123"
		config := createTestPipelineConfig()

		mockProvider := &MockProvider{}
		mockProvider.On("GetName").Return("tekton")
		mockProvider.On("ValidateConfig", ctx, mock.AnythingOfType("*cicd.PipelineConfig")).Return(nil)
		
		runID := uuid.New().String()
		expectedRun := &cicd.PipelineRun{
			ID:        runID,
			Status:    cicd.PipelineStatusRunning,
			StartedAt: time.Now(),
		}
		mockProvider.On("RunPipeline", ctx, mock.AnythingOfType("*cicd.PipelineConfig")).Return(expectedRun, nil)

		factory.On("CreateProvider", "tekton", mock.AnythingOfType("*cicd.ProviderConfig")).Return(mockProvider, nil)
		
		repo.On("CreatePipeline", ctx, mock.AnythingOfType("*cicd.Pipeline")).Return(nil)
		repo.On("CreatePipelineRun", ctx, mock.AnythingOfType("*cicd.PipelineRunRecord")).Return(nil)
		repo.On("GetProviderConfig", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Act
		run, err := service.CreatePipeline(ctx, workspaceID, config)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, run)
		assert.Equal(t, runID, run.ID)
		assert.Equal(t, cicd.PipelineStatusRunning, run.Status)

		repo.AssertExpectations(t)
		factory.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})

	t.Run("validation failure", func(t *testing.T) {
		service, repo, factory, _ := createTestService()
		ctx := context.Background()
		workspaceID := "workspace-123"
		config := createTestPipelineConfig()

		mockProvider := &MockProvider{}
		mockProvider.On("GetName").Return("tekton")
		mockProvider.On("ValidateConfig", ctx, mock.AnythingOfType("*cicd.PipelineConfig")).Return(errors.New("invalid config"))

		factory.On("CreateProvider", "tekton", mock.AnythingOfType("*cicd.ProviderConfig")).Return(mockProvider, nil)
		repo.On("GetProviderConfig", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Act
		run, err := service.CreatePipeline(ctx, workspaceID, config)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, run)
		assert.Contains(t, err.Error(), "invalid pipeline configuration")
	})

	t.Run("provider creation failure", func(t *testing.T) {
		service, repo, factory, _ := createTestService()
		ctx := context.Background()
		workspaceID := "workspace-123"
		config := createTestPipelineConfig()

		factory.On("CreateProvider", "tekton", mock.AnythingOfType("*cicd.ProviderConfig")).Return(nil, errors.New("provider error"))
		repo.On("GetProviderConfig", ctx, workspaceID).Return(nil, errors.New("not found"))

		// Act
		run, err := service.CreatePipeline(ctx, workspaceID, config)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, run)
		assert.Contains(t, err.Error(), "failed to get provider")
	})
}

func TestGetPipeline(t *testing.T) {
	t.Run("successful get pipeline", func(t *testing.T) {
		service, repo, factory, _ := createTestService()
		ctx := context.Background()
		pipelineID := "pipeline-123"
		runID := "run-123"
		workspaceID := "workspace-123"

		runRecord := &cicd.PipelineRunRecord{
			ID:         pipelineID,
			PipelineID: "config-123",
			RunID:      runID,
			Status:     "running",
		}

		pipeline := &cicd.Pipeline{
			ID:          "config-123",
			WorkspaceID: workspaceID,
			Provider:    "tekton",
		}

		expectedRun := &cicd.PipelineRun{
			ID:     runID,
			Status: cicd.PipelineStatusSucceeded,
		}

		mockProvider := &MockProvider{}
		mockProvider.On("GetStatus", ctx, workspaceID, runID).Return(expectedRun, nil)

		repo.On("GetPipelineRun", ctx, pipelineID).Return(runRecord, nil)
		repo.On("GetPipeline", ctx, "config-123").Return(pipeline, nil)
		repo.On("GetProviderConfig", ctx, workspaceID).Return(nil, errors.New("not found"))
		factory.On("CreateProvider", "tekton", mock.AnythingOfType("*cicd.ProviderConfig")).Return(mockProvider, nil)

		// Act
		run, err := service.GetPipeline(ctx, pipelineID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, run)
		assert.Equal(t, runID, run.ID)
		assert.Equal(t, cicd.PipelineStatusSucceeded, run.Status)
	})

	t.Run("pipeline not found", func(t *testing.T) {
		service, repo, _, _ := createTestService()
		ctx := context.Background()
		pipelineID := "pipeline-123"

		repo.On("GetPipelineRun", ctx, pipelineID).Return(nil, errors.New("not found"))

		// Act
		run, err := service.GetPipeline(ctx, pipelineID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, run)
		assert.Contains(t, err.Error(), "pipeline not found")
	})
}

func TestCancelPipeline(t *testing.T) {
	t.Run("successful cancel", func(t *testing.T) {
		service, repo, factory, _ := createTestService()
		ctx := context.Background()
		pipelineID := "pipeline-123"
		runID := "run-123"
		workspaceID := "workspace-123"

		runRecord := &cicd.PipelineRunRecord{
			ID:         pipelineID,
			PipelineID: "config-123",
			RunID:      runID,
			Status:     "running",
		}

		pipeline := &cicd.Pipeline{
			ID:          "config-123",
			WorkspaceID: workspaceID,
			Provider:    "tekton",
		}

		mockProvider := &MockProvider{}
		mockProvider.On("CancelPipeline", ctx, workspaceID, runID).Return(nil)

		repo.On("GetPipelineRun", ctx, pipelineID).Return(runRecord, nil)
		repo.On("GetPipeline", ctx, "config-123").Return(pipeline, nil)
		repo.On("GetProviderConfig", ctx, workspaceID).Return(nil, errors.New("not found"))
		repo.On("UpdatePipelineRun", ctx, mock.AnythingOfType("*cicd.PipelineRunRecord")).Return(nil)
		factory.On("CreateProvider", "tekton", mock.AnythingOfType("*cicd.ProviderConfig")).Return(mockProvider, nil)

		// Act
		err := service.CancelPipeline(ctx, pipelineID)

		// Assert
		assert.NoError(t, err)
		repo.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})
}

func TestListProviders(t *testing.T) {
	t.Run("successful list providers", func(t *testing.T) {
		service, _, factory, _ := createTestService()
		ctx := context.Background()

		providers := []string{"tekton", "github-actions", "gitlab-ci"}
		factory.On("ListProviders").Return(providers)

		// Act
		infos, err := service.ListProviders(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, infos, 3)
		assert.Equal(t, "tekton", infos[0].Name)
		assert.Equal(t, "Tekton Pipelines", infos[0].DisplayName)
		assert.Equal(t, "available", infos[0].Status)
		assert.Equal(t, "github-actions", infos[1].Name)
		assert.Equal(t, "beta", infos[1].Status)
	})
}

func TestCredentialManagement(t *testing.T) {
	t.Run("create git credential", func(t *testing.T) {
		service, _, _, credManager := createTestService()
		ctx := context.Background()
		workspaceID := "workspace-123"
		name := "github-ssh"
		credential := cicd.GitCredential{
			Type:   "ssh",
			SSHKey: "-----BEGIN RSA PRIVATE KEY-----",
		}

		expectedInfo := &cicd.CredentialInfo{
			Name:      name,
			Type:      "git-ssh",
			CreatedAt: time.Now(),
		}
		credManager.On("StoreGitCredential", workspaceID, &credential).Return(expectedInfo, nil)

		// Act
		err := service.CreateGitCredential(ctx, workspaceID, name, credential)

		// Assert
		assert.NoError(t, err)
		credManager.AssertExpectations(t)
	})

	t.Run("create registry credential", func(t *testing.T) {
		service, _, _, credManager := createTestService()
		ctx := context.Background()
		workspaceID := "workspace-123"
		name := "dockerhub"
		credential := cicd.RegistryCredential{
			Registry: "docker.io",
			Username: "user",
			Password: "pass",
		}

		expectedInfo := &cicd.CredentialInfo{
			Name:      name,
			Type:      "registry",
			CreatedAt: time.Now(),
		}
		credManager.On("StoreRegistryCredential", workspaceID, &credential).Return(expectedInfo, nil)

		// Act
		err := service.CreateRegistryCredential(ctx, workspaceID, name, credential)

		// Assert
		assert.NoError(t, err)
		credManager.AssertExpectations(t)
	})

	t.Run("list credentials", func(t *testing.T) {
		service, _, _, credManager := createTestService()
		ctx := context.Background()
		workspaceID := "workspace-123"

		expectedCreds := []*cicd.CredentialInfo{
			{
				Name:      "github-ssh",
				Type:      "git-ssh",
				CreatedAt: time.Now(),
			},
			{
				Name:      "dockerhub",
				Type:      "registry",
				CreatedAt: time.Now(),
			},
		}

		credManager.On("ListCredentials", workspaceID).Return(expectedCreds, nil)

		// Act
		creds, err := service.ListCredentials(ctx, workspaceID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, creds, 2)
		assert.Equal(t, "github-ssh", creds[0].Name)
		assert.Equal(t, "dockerhub", creds[1].Name)
	})
}