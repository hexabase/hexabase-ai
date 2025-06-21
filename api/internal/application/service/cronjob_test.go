package service

import (
	"context"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateCronJob(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name          string
		app           *domain.Application
		expectedError error
		setupMocks    func(*MockRepository, *MockKubernetesRepository)
	}{
		{
			name: "successful cronjob creation",
			app: &domain.Application{
				WorkspaceID:  "ws-123",
				ProjectID:    "proj-456",
				Name:         "daily-backup",
				Type:         domain.ApplicationTypeCronJob,
				CronSchedule: "0 2 * * *",
				CronCommand:  []string{"/usr/bin/backup.sh"},
				CronArgs:     []string{"--full", "--compress"},
				Source: domain.ApplicationSource{
					Type:  domain.SourceTypeImage,
					Image: "backup-tool:latest",
				},
				Config: domain.ApplicationConfig{
					Resources: domain.ResourceRequests{
						CPURequest:    "100m",
						MemoryRequest: "256Mi",
						CPULimit:      "500m",
						MemoryLimit:   "1Gi",
					},
					EnvVars: map[string]string{
						"BACKUP_PATH": "/data",
					},
				},
			},
			expectedError: nil,
			setupMocks: func(repo *MockRepository, k8sRepo *MockKubernetesRepository) {
				// Mock repository create
				repo.On("Create", ctx, mock.MatchedBy(func(app *domain.Application) bool {
					return app.Name == "daily-backup" && app.Type == domain.ApplicationTypeCronJob
				})).Return(nil).Run(func(args mock.Arguments) {
					app := args.Get(1).(*domain.Application)
					app.ID = "app-789"
					app.Status = domain.ApplicationStatusPending
				})
				
				// Mock Kubernetes CronJob creation
				k8sRepo.On("CreateCronJob", ctx, "ws-123", "proj-456", mock.MatchedBy(func(spec domain.CronJobSpec) bool {
					return spec.Name == "daily-backup" &&
						spec.Schedule == "0 2 * * *" &&
						spec.Image == "backup-tool:latest" &&
						len(spec.Command) == 1 && spec.Command[0] == "/usr/bin/backup.sh" &&
						len(spec.Args) == 2
				})).Return(nil)
				
				// Mock status update
				repo.On("UpdateApplication", ctx, mock.MatchedBy(func(app *domain.Application) bool {
					return app.ID == "app-789" && app.Status == domain.ApplicationStatusRunning
				})).Return(nil)
			},
		},
		{
			name: "cronjob with template application",
			app: &domain.Application{
				WorkspaceID:   "ws-123",
				ProjectID:     "proj-456",
				Name:          "scheduled-task",
				Type:          domain.ApplicationTypeCronJob,
				CronSchedule:  "*/5 * * * *",
				TemplateAppID: "template-123",
			},
			expectedError: nil,
			setupMocks: func(repo *MockRepository, k8sRepo *MockKubernetesRepository) {
				// Template app will be fetched in repository layer
				repo.On("Create", ctx, mock.MatchedBy(func(app *domain.Application) bool {
					return app.Name == "scheduled-task" && app.TemplateAppID == "template-123"
				})).Return(nil).Run(func(args mock.Arguments) {
					app := args.Get(1).(*domain.Application)
					app.ID = "app-890"
					app.Status = domain.ApplicationStatusPending
					// Simulate template app values being copied
					app.Source.Type = domain.SourceTypeImage
					app.Source.Image = "template-image:v1"
					app.CronCommand = []string{"python", "script.py"}
				})
				
				// Mock Kubernetes CronJob creation
				k8sRepo.On("CreateCronJob", ctx, "ws-123", "proj-456", mock.MatchedBy(func(spec domain.CronJobSpec) bool {
					return spec.Name == "scheduled-task" &&
						spec.Schedule == "*/5 * * * *" &&
						spec.Image == "template-image:v1"
				})).Return(nil)
				
				// Mock status update
				repo.On("UpdateApplication", ctx, mock.Anything).Return(nil)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockK8sRepo := new(MockKubernetesRepository)
			
			service := &Service{
				repo:    mockRepo,
			k8s:     mockK8sRepo,
				k8sRepo: mockK8sRepo,
			}
			
			tt.setupMocks(mockRepo, mockK8sRepo)
			
			err := service.CreateCronJob(ctx, tt.app)
			
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			
			mockRepo.AssertExpectations(t)
			mockK8sRepo.AssertExpectations(t)
		})
	}
}

func TestService_UpdateCronJobSchedule(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name          string
		applicationID string
		newSchedule   string
		expectedError error
		setupMocks    func(*MockRepository, *MockKubernetesRepository)
	}{
		{
			name:          "successful schedule update",
			applicationID: "app-123",
			newSchedule:   "0 4 * * *",
			expectedError: nil,
			setupMocks: func(repo *MockRepository, k8sRepo *MockKubernetesRepository) {
				// Mock get application
				app := &domain.Application{
					ID:           "app-123",
					WorkspaceID:  "ws-123",
					ProjectID:    "proj-456",
					Name:         "daily-backup",
					Type:         domain.ApplicationTypeCronJob,
					CronSchedule: "0 2 * * *",
					Status:       domain.ApplicationStatusRunning,
				}
				repo.On("GetApplication", ctx, "app-123").Return(app, nil)
				
				// Mock update schedule in repository
				repo.On("UpdateCronSchedule", ctx, "app-123", "0 4 * * *").Return(nil)
				
				// Mock update CronJob in Kubernetes
				k8sRepo.On("UpdateCronJob", ctx, "ws-123", "proj-456", "daily-backup", 
					mock.MatchedBy(func(spec domain.CronJobSpec) bool {
						return spec.Schedule == "0 4 * * *"
					})).Return(nil)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockK8sRepo := new(MockKubernetesRepository)
			
			service := &Service{
				repo:    mockRepo,
			k8s:     mockK8sRepo,
				k8sRepo: mockK8sRepo,
			}
			
			tt.setupMocks(mockRepo, mockK8sRepo)
			
			err := service.UpdateCronJobSchedule(ctx, tt.applicationID, tt.newSchedule)
			
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			
			mockRepo.AssertExpectations(t)
			mockK8sRepo.AssertExpectations(t)
		})
	}
}

func TestService_TriggerCronJob(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name          string
		applicationID string
		expectedError error
		setupMocks    func(*MockRepository, *MockKubernetesRepository)
	}{
		{
			name:          "successful manual trigger",
			applicationID: "app-123",
			expectedError: nil,
			setupMocks: func(repo *MockRepository, k8sRepo *MockKubernetesRepository) {
				// Mock get application
				app := &domain.Application{
					ID:           "app-123",
					WorkspaceID:  "ws-123",
					ProjectID:    "proj-456",
					Name:         "daily-backup",
					Type:         domain.ApplicationTypeCronJob,
					CronSchedule: "0 2 * * *",
					Status:       domain.ApplicationStatusRunning,
				}
				repo.On("GetApplication", ctx, "app-123").Return(app, nil)
				
				// Mock trigger CronJob
				k8sRepo.On("TriggerCronJob", ctx, "ws-123", "proj-456", "daily-backup").Return(nil)
				
				// Mock create execution record
				repo.On("CreateCronJobExecution", ctx, mock.MatchedBy(func(exec *domain.CronJobExecution) bool {
					return exec.ApplicationID == "app-123" &&
						exec.Status == domain.CronJobExecutionStatusRunning
				})).Return(nil)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockK8sRepo := new(MockKubernetesRepository)
			
			service := &Service{
				repo:    mockRepo,
			k8s:     mockK8sRepo,
				k8sRepo: mockK8sRepo,
			}
			
			tt.setupMocks(mockRepo, mockK8sRepo)
			
			req := &domain.TriggerCronJobRequest{
				ApplicationID: tt.applicationID,
			}
			_, err := service.TriggerCronJob(ctx, req)
			
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			
			mockRepo.AssertExpectations(t)
			mockK8sRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetCronJobExecutions(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	
	tests := []struct {
		name                  string
		applicationID         string
		limit                 int
		offset                int
		expectedExecutions    []domain.CronJobExecution
		expectedTotal         int
		expectedError         error
		setupMocks           func(*MockRepository)
	}{
		{
			name:          "successful get executions",
			applicationID: "app-123",
			limit:         10,
			offset:        0,
			expectedExecutions: []domain.CronJobExecution{
				{
					ID:            "exec-1",
					ApplicationID: "app-123",
					JobName:       "daily-backup-1234567890",
					StartedAt:     now.Add(-1 * time.Hour),
					CompletedAt:   &now,
					Status:        domain.CronJobExecutionStatusSucceeded,
					ExitCode:      intPtr(0),
					Logs:          "Backup completed successfully",
				},
				{
					ID:            "exec-2",
					ApplicationID: "app-123",
					JobName:       "daily-backup-1234567891",
					StartedAt:     now.Add(-2 * time.Hour),
					Status:        domain.CronJobExecutionStatusRunning,
				},
			},
			expectedTotal: 2,
			expectedError: nil,
			setupMocks: func(repo *MockRepository) {
				// Mock get application
				app := &domain.Application{
					ID:   "app-123",
					Type: domain.ApplicationTypeCronJob,
				}
				repo.On("GetApplication", ctx, "app-123").Return(app, nil)
				
				repo.On("GetCronJobExecutions", ctx, "app-123", 10, 0).
					Return([]domain.CronJobExecution{
						{
							ID:            "exec-1",
							ApplicationID: "app-123",
							JobName:       "daily-backup-1234567890",
							StartedAt:     now.Add(-1 * time.Hour),
							CompletedAt:   &now,
							Status:        domain.CronJobExecutionStatusSucceeded,
							ExitCode:      intPtr(0),
							Logs:          "Backup completed successfully",
						},
						{
							ID:            "exec-2",
							ApplicationID: "app-123",
							JobName:       "daily-backup-1234567891",
							StartedAt:     now.Add(-2 * time.Hour),
							Status:        domain.CronJobExecutionStatusRunning,
						},
					}, 2, nil)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			
			service := &Service{
				repo: mockRepo,
			}
			
			tt.setupMocks(mockRepo)
			
			executions, total, err := service.GetCronJobExecutions(ctx, tt.applicationID, tt.limit, tt.offset)
			
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedExecutions, executions)
				assert.Equal(t, tt.expectedTotal, total)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}