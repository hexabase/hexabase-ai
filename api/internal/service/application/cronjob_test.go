package application

import (
	"context"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateCronJob(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name          string
		app           *application.Application
		expectedError error
		setupMocks    func(*MockRepository, *MockKubernetesRepository)
	}{
		{
			name: "successful cronjob creation",
			app: &application.Application{
				WorkspaceID:  "ws-123",
				ProjectID:    "proj-456",
				Name:         "daily-backup",
				Type:         application.ApplicationTypeCronJob,
				CronSchedule: "0 2 * * *",
				CronCommand:  []string{"/usr/bin/backup.sh"},
				CronArgs:     []string{"--full", "--compress"},
				Source: application.ApplicationSource{
					Type:  application.SourceTypeImage,
					Image: "backup-tool:latest",
				},
				Config: application.ApplicationConfig{
					Resources: application.ResourceRequests{
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
				repo.On("Create", ctx, mock.MatchedBy(func(app *application.Application) bool {
					return app.Name == "daily-backup" && app.Type == application.ApplicationTypeCronJob
				})).Return(nil).Run(func(args mock.Arguments) {
					app := args.Get(1).(*application.Application)
					app.ID = "app-789"
					app.Status = application.ApplicationStatusPending
				})
				
				// Mock Kubernetes CronJob creation
				k8sRepo.On("CreateCronJob", ctx, "ws-123", "proj-456", mock.MatchedBy(func(spec application.CronJobSpec) bool {
					return spec.Name == "daily-backup" &&
						spec.Schedule == "0 2 * * *" &&
						spec.Image == "backup-tool:latest" &&
						len(spec.Command) == 1 && spec.Command[0] == "/usr/bin/backup.sh" &&
						len(spec.Args) == 2
				})).Return(nil)
				
				// Mock status update
				repo.On("UpdateApplication", ctx, mock.MatchedBy(func(app *application.Application) bool {
					return app.ID == "app-789" && app.Status == application.ApplicationStatusRunning
				})).Return(nil)
			},
		},
		{
			name: "cronjob with template application",
			app: &application.Application{
				WorkspaceID:   "ws-123",
				ProjectID:     "proj-456",
				Name:          "scheduled-task",
				Type:          application.ApplicationTypeCronJob,
				CronSchedule:  "*/5 * * * *",
				TemplateAppID: "template-123",
			},
			expectedError: nil,
			setupMocks: func(repo *MockRepository, k8sRepo *MockKubernetesRepository) {
				// Template app will be fetched in repository layer
				repo.On("Create", ctx, mock.MatchedBy(func(app *application.Application) bool {
					return app.Name == "scheduled-task" && app.TemplateAppID == "template-123"
				})).Return(nil).Run(func(args mock.Arguments) {
					app := args.Get(1).(*application.Application)
					app.ID = "app-890"
					app.Status = application.ApplicationStatusPending
					// Simulate template app values being copied
					app.Source.Type = application.SourceTypeImage
					app.Source.Image = "template-image:v1"
					app.CronCommand = []string{"python", "script.py"}
				})
				
				// Mock Kubernetes CronJob creation
				k8sRepo.On("CreateCronJob", ctx, "ws-123", "proj-456", mock.MatchedBy(func(spec application.CronJobSpec) bool {
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
				app := &application.Application{
					ID:           "app-123",
					WorkspaceID:  "ws-123",
					ProjectID:    "proj-456",
					Name:         "daily-backup",
					Type:         application.ApplicationTypeCronJob,
					CronSchedule: "0 2 * * *",
					Status:       application.ApplicationStatusRunning,
				}
				repo.On("GetApplication", ctx, "app-123").Return(app, nil)
				
				// Mock update schedule in repository
				repo.On("UpdateCronSchedule", ctx, "app-123", "0 4 * * *").Return(nil)
				
				// Mock update CronJob in Kubernetes
				k8sRepo.On("UpdateCronJob", ctx, "ws-123", "proj-456", "daily-backup", 
					mock.MatchedBy(func(spec application.CronJobSpec) bool {
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
				app := &application.Application{
					ID:           "app-123",
					WorkspaceID:  "ws-123",
					ProjectID:    "proj-456",
					Name:         "daily-backup",
					Type:         application.ApplicationTypeCronJob,
					CronSchedule: "0 2 * * *",
					Status:       application.ApplicationStatusRunning,
				}
				repo.On("GetApplication", ctx, "app-123").Return(app, nil)
				
				// Mock trigger CronJob
				k8sRepo.On("TriggerCronJob", ctx, "ws-123", "proj-456", "daily-backup").Return(nil)
				
				// Mock create execution record
				repo.On("CreateCronJobExecution", ctx, mock.MatchedBy(func(exec *application.CronJobExecution) bool {
					return exec.ApplicationID == "app-123" &&
						exec.Status == application.CronJobExecutionStatusRunning
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
				k8sRepo: mockK8sRepo,
			}
			
			tt.setupMocks(mockRepo, mockK8sRepo)
			
			req := &application.TriggerCronJobRequest{
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
		expectedExecutions    []application.CronJobExecution
		expectedTotal         int
		expectedError         error
		setupMocks           func(*MockRepository)
	}{
		{
			name:          "successful get executions",
			applicationID: "app-123",
			limit:         10,
			offset:        0,
			expectedExecutions: []application.CronJobExecution{
				{
					ID:            "exec-1",
					ApplicationID: "app-123",
					JobName:       "daily-backup-1234567890",
					StartedAt:     now.Add(-1 * time.Hour),
					CompletedAt:   &now,
					Status:        application.CronJobExecutionStatusSucceeded,
					ExitCode:      intPtr(0),
					Logs:          "Backup completed successfully",
				},
				{
					ID:            "exec-2",
					ApplicationID: "app-123",
					JobName:       "daily-backup-1234567891",
					StartedAt:     now.Add(-2 * time.Hour),
					Status:        application.CronJobExecutionStatusRunning,
				},
			},
			expectedTotal: 2,
			expectedError: nil,
			setupMocks: func(repo *MockRepository) {
				repo.On("GetCronJobExecutions", ctx, "app-123", 10, 0).
					Return([]application.CronJobExecution{
						{
							ID:            "exec-1",
							ApplicationID: "app-123",
							JobName:       "daily-backup-1234567890",
							StartedAt:     now.Add(-1 * time.Hour),
							CompletedAt:   &now,
							Status:        application.CronJobExecutionStatusSucceeded,
							ExitCode:      intPtr(0),
							Logs:          "Backup completed successfully",
						},
						{
							ID:            "exec-2",
							ApplicationID: "app-123",
							JobName:       "daily-backup-1234567891",
							StartedAt:     now.Add(-2 * time.Hour),
							Status:        application.CronJobExecutionStatusRunning,
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