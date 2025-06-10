package application

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	"github.com/hexabase/hexabase-ai/api/internal/domain/monitoring"
	"github.com/hexabase/hexabase-ai/api/internal/domain/project"
)

// Extend existing mocks with additional methods if needed
// These should be in a shared mocks file, but for now we'll add the missing methods

// MockProjectService for project service
type MockProjectService struct {
	mock.Mock
}

func (m *MockProjectService) GetProject(ctx context.Context, projectID string) (*project.Project, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *MockProjectService) CreateProject(ctx context.Context, req *project.CreateProjectRequest) (*project.Project, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *MockProjectService) ListProjects(ctx context.Context, filter project.ProjectFilter) (*project.ProjectList, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ProjectList), args.Error(1)
}

func (m *MockProjectService) UpdateProject(ctx context.Context, projectID string, req *project.UpdateProjectRequest) (*project.Project, error) {
	args := m.Called(ctx, projectID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *MockProjectService) DeleteProject(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockProjectService) GetProjectStats(ctx context.Context, projectID string) (*project.ProjectStats, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ProjectStats), args.Error(1)
}

func (m *MockProjectService) CreateSubProject(ctx context.Context, parentID string, req *project.CreateProjectRequest) (*project.Project, error) {
	args := m.Called(ctx, parentID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *MockProjectService) GetProjectHierarchy(ctx context.Context, projectID string) (*project.ProjectHierarchy, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ProjectHierarchy), args.Error(1)
}

func (m *MockProjectService) ApplyResourceQuota(ctx context.Context, projectID string, quota *project.ResourceQuota) error {
	args := m.Called(ctx, projectID, quota)
	return args.Error(0)
}

func (m *MockProjectService) GetResourceUsage(ctx context.Context, projectID string) (*project.ResourceUsage, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ResourceUsage), args.Error(1)
}

func (m *MockProjectService) CreateNamespace(ctx context.Context, projectID string, req *project.CreateNamespaceRequest) (*project.Namespace, error) {
	args := m.Called(ctx, projectID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Namespace), args.Error(1)
}

func (m *MockProjectService) GetNamespace(ctx context.Context, projectID, namespaceID string) (*project.Namespace, error) {
	args := m.Called(ctx, projectID, namespaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Namespace), args.Error(1)
}

func (m *MockProjectService) ListNamespaces(ctx context.Context, projectID string) (*project.NamespaceList, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.NamespaceList), args.Error(1)
}

func (m *MockProjectService) UpdateNamespace(ctx context.Context, projectID, namespaceID string, req *project.CreateNamespaceRequest) (*project.Namespace, error) {
	args := m.Called(ctx, projectID, namespaceID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Namespace), args.Error(1)
}

func (m *MockProjectService) DeleteNamespace(ctx context.Context, projectID, namespaceID string) error {
	args := m.Called(ctx, projectID, namespaceID)
	return args.Error(0)
}

func (m *MockProjectService) GetNamespaceUsage(ctx context.Context, projectID, namespaceID string) (*project.NamespaceUsage, error) {
	args := m.Called(ctx, projectID, namespaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.NamespaceUsage), args.Error(1)
}

func (m *MockProjectService) AddMember(ctx context.Context, projectID, adderID string, req *project.AddMemberRequest) (*project.ProjectMember, error) {
	args := m.Called(ctx, projectID, adderID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ProjectMember), args.Error(1)
}

func (m *MockProjectService) GetMember(ctx context.Context, projectID, memberID string) (*project.ProjectMember, error) {
	args := m.Called(ctx, projectID, memberID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ProjectMember), args.Error(1)
}

func (m *MockProjectService) ListMembers(ctx context.Context, projectID string) (*project.MemberList, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.MemberList), args.Error(1)
}

func (m *MockProjectService) UpdateMemberRole(ctx context.Context, projectID, memberID string, req *project.UpdateMemberRoleRequest) (*project.ProjectMember, error) {
	args := m.Called(ctx, projectID, memberID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ProjectMember), args.Error(1)
}

func (m *MockProjectService) RemoveMember(ctx context.Context, projectID, memberID, removerID string) error {
	args := m.Called(ctx, projectID, memberID, removerID)
	return args.Error(0)
}

func (m *MockProjectService) AddProjectMember(ctx context.Context, projectID string, req *project.AddMemberRequest) error {
	args := m.Called(ctx, projectID, req)
	return args.Error(0)
}

func (m *MockProjectService) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	args := m.Called(ctx, projectID, userID)
	return args.Error(0)
}

func (m *MockProjectService) ListProjectMembers(ctx context.Context, projectID string) ([]*project.ProjectMember, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*project.ProjectMember), args.Error(1)
}

func (m *MockProjectService) ListActivities(ctx context.Context, projectID string, limit int) (*project.ActivityList, error) {
	args := m.Called(ctx, projectID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.ActivityList), args.Error(1)
}

func (m *MockProjectService) LogActivity(ctx context.Context, activity *project.ProjectActivity) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

func (m *MockProjectService) GetActivityLogs(ctx context.Context, projectID string, filter project.ActivityFilter) ([]*project.ProjectActivity, error) {
	args := m.Called(ctx, projectID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*project.ProjectActivity), args.Error(1)
}

func (m *MockProjectService) ValidateProjectAccess(ctx context.Context, userID, projectID string, requiredRole string) error {
	args := m.Called(ctx, userID, projectID, requiredRole)
	return args.Error(0)
}

func (m *MockProjectService) GetUserProjectRole(ctx context.Context, userID, projectID string) (string, error) {
	args := m.Called(ctx, userID, projectID)
	return args.String(0), args.Error(1)
}

// Fix MockBackupService to match the interface
func (m *MockBackupService) CleanupOldBackups(ctx context.Context, policyID string) error {
	args := m.Called(ctx, policyID)
	return args.Error(0)
}

// Add missing monitoring service methods
func (m *MockMonitoringService) CollectMetrics(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockMonitoringService) GetWorkspaceMetrics(ctx context.Context, workspaceID string, opts monitoring.QueryOptions) (*monitoring.WorkspaceMetrics, error) {
	args := m.Called(ctx, workspaceID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*monitoring.WorkspaceMetrics), args.Error(1)
}

func (m *MockMonitoringService) GetClusterHealth(ctx context.Context, workspaceID string) (*monitoring.ClusterHealth, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*monitoring.ClusterHealth), args.Error(1)
}

func (m *MockMonitoringService) GetResourceUsage(ctx context.Context, workspaceID string) (*monitoring.ResourceUsage, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*monitoring.ResourceUsage), args.Error(1)
}

func (m *MockMonitoringService) GetAlerts(ctx context.Context, workspaceID, severity string) ([]*monitoring.Alert, error) {
	args := m.Called(ctx, workspaceID, severity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*monitoring.Alert), args.Error(1)
}

func (m *MockMonitoringService) CreateAlert(ctx context.Context, alert *monitoring.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockMonitoringService) ResolveAlert(ctx context.Context, alertID string) error {
	args := m.Called(ctx, alertID)
	return args.Error(0)
}

// Remove duplicate Project type - use the actual domain type

func TestService_CreateCronJobWithBackupPolicy(t *testing.T) {
	tests := []struct {
		name              string
		req               *application.CreateApplicationRequest
		backupPolicyReq   *backup.CreateBackupPolicyRequest
		setupMocks        func(*MockApplicationRepository, *MockKubernetesRepository, *MockBackupService)
		wantErr           bool
		errMessage        string
	}{
		{
			name: "create cronjob with backup policy - success",
			req: &application.CreateApplicationRequest{
				Name:      "backup-cronjob",
				Type:      application.ApplicationTypeCronJob,
				ProjectID: "proj-123",
				Source: application.ApplicationSource{
					Type:  application.SourceTypeImage,
					Image: "backup-tool:latest",
				},
				Config: application.ApplicationConfig{
					Environment: map[string]string{
						"BACKUP_TARGET": "database",
					},
				},
				CronSchedule: "0 2 * * *", // Daily at 2 AM
			},
			backupPolicyReq: &backup.CreateBackupPolicyRequest{
				StorageID:          "storage-123",
				Enabled:            true,
				Schedule:           "0 3 * * *", // Daily at 3 AM (after backup job)
				RetentionDays:      30,
				BackupType:         backup.BackupTypeFull,
				IncludeVolumes:     true,
				IncludeDatabase:    true,
				IncludeConfig:      true,
				CompressionEnabled: true,
				EncryptionEnabled:  true,
			},
			setupMocks: func(appRepo *MockApplicationRepository, k8sRepo *MockKubernetesRepository, backupSvc *MockBackupService) {
				// Create application
				appRepo.On("Create", mock.Anything, mock.MatchedBy(func(app *application.Application) bool {
					return app.Name == "backup-cronjob" && 
						   app.Type == application.ApplicationTypeCronJob &&
						   app.Metadata["backup_enabled"] == "true"
				})).Return(nil)

				// Create CronJob in Kubernetes
				k8sRepo.On("CreateCronJob", mock.Anything, mock.Anything).Return(nil)

				// Create backup policy
				backupSvc.On("CreateBackupPolicy", mock.Anything, "app-123", mock.MatchedBy(func(req backup.CreateBackupPolicyRequest) bool {
					return req.StorageID == "storage-123" &&
						   req.Schedule == "0 3 * * *" &&
						   req.RetentionDays == 30
				})).Return(&backup.BackupPolicy{
					ID:            "policy-123",
					ApplicationID: "app-123",
					StorageID:     "storage-123",
					Schedule:      "0 3 * * *",
				}, nil)

				// Update application with backup policy ID
				appRepo.On("Update", mock.Anything, "app-123", mock.MatchedBy(func(app *application.Application) bool {
					return app.Metadata["backup_policy_id"] == "policy-123"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "create cronjob with backup policy - validate schedule compatibility",
			req: &application.CreateApplicationRequest{
				Name:         "backup-cronjob",
				Type:         application.ApplicationTypeCronJob,
				ProjectID:    "proj-123",
				CronSchedule: "0 2 * * *", // Daily at 2 AM
			},
			backupPolicyReq: &backup.CreateBackupPolicyRequest{
				StorageID:     "storage-123",
				Schedule:      "0 1 * * *", // Daily at 1 AM (before cronjob)
				RetentionDays: 30,
			},
			setupMocks: func(appRepo *MockApplicationRepository, k8sRepo *MockKubernetesRepository, backupSvc *MockBackupService) {
				// Should not be called due to validation error
			},
			wantErr:    true,
			errMessage: "backup policy schedule must run after cronjob schedule",
		},
		{
			name: "create cronjob with backup hooks",
			req: &application.CreateApplicationRequest{
				Name:         "db-backup-cronjob",
				Type:         application.ApplicationTypeCronJob,
				ProjectID:    "proj-123",
				CronSchedule: "0 2 * * *",
			},
			backupPolicyReq: &backup.CreateBackupPolicyRequest{
				StorageID:      "storage-123",
				Schedule:       "0 4 * * *",
				RetentionDays:  30,
				PreBackupHook:  "kubectl exec -n {namespace} {pod} -- /scripts/pre-backup.sh",
				PostBackupHook: "kubectl exec -n {namespace} {pod} -- /scripts/post-backup.sh",
			},
			setupMocks: func(appRepo *MockApplicationRepository, k8sRepo *MockKubernetesRepository, backupSvc *MockBackupService) {
				appRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				k8sRepo.On("CreateCronJob", mock.Anything, mock.Anything).Return(nil)
				
				backupSvc.On("CreateBackupPolicy", mock.Anything, "app-123", mock.MatchedBy(func(req backup.CreateBackupPolicyRequest) bool {
					return req.PreBackupHook != "" && req.PostBackupHook != ""
				})).Return(&backup.BackupPolicy{
					ID:             "policy-123",
					PreBackupHook:  "kubectl exec -n {namespace} {pod} -- /scripts/pre-backup.sh",
					PostBackupHook: "kubectl exec -n {namespace} {pod} -- /scripts/post-backup.sh",
				}, nil)
				
				appRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			appRepo := new(MockApplicationRepository)
			k8sRepo := new(MockKubernetesRepository)
			projectService := new(MockProjectService)
			backupSvc := new(MockBackupService)
			monitoringSvc := new(MockMonitoringService)

			// Configure project mock
			projectService.On("GetProject", mock.Anything, "proj-123").Return(&project.Project{
				ID:            "proj-123",
				NamespaceName: "test-namespace",
			}, nil)

			// Setup specific test mocks
			if tt.setupMocks != nil {
				tt.setupMocks(appRepo, k8sRepo, backupSvc)
			}

			// Create base service
			baseService := &Service{
				repo:    appRepo,
				k8s:     k8sRepo,
				k8sRepo: k8sRepo,
				logger:  slog.Default(),
			}
			
			// Create extended service with correct field names
			svc := &ExtendedService{
				Service:           baseService,
				projectService:    projectService,
				backupService:     backupSvc,
				monitoringService: monitoringSvc,
			}

			// Set application ID for mocking
			if tt.req != nil {
				tt.req.Metadata = map[string]string{"mock_app_id": "app-123"}
			}

			// Execute
			ctx := context.Background()
			app, err := svc.CreateApplicationWithBackupPolicy(ctx, tt.req, tt.backupPolicyReq)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app)
				assert.Equal(t, "true", app.Metadata["backup_enabled"])
			}

			// Verify mock expectations
			appRepo.AssertExpectations(t)
			k8sRepo.AssertExpectations(t)
			backupSvc.AssertExpectations(t)
		})
	}
}

func TestService_TriggerCronJobWithBackup(t *testing.T) {
	tests := []struct {
		name       string
		appID      string
		setupMocks func(*MockApplicationRepository, *MockKubernetesRepository, *MockBackupService)
		wantErr    bool
	}{
		{
			name:  "trigger cronjob and create backup execution",
			appID: "app-123",
			setupMocks: func(appRepo *MockApplicationRepository, k8sRepo *MockKubernetesRepository, backupSvc *MockBackupService) {
				// Get application with backup policy
				appRepo.On("GetByID", mock.Anything, "app-123").Return(&application.Application{
					ID:   "app-123",
					Name: "backup-cronjob",
					Type: application.ApplicationTypeCronJob,
					Metadata: map[string]string{
						"backup_enabled":    "true",
						"backup_policy_id": "policy-123",
					},
					Status: application.ApplicationStatusRunning,
				}, nil)

				// Trigger CronJob
				k8sRepo.On("TriggerCronJob", mock.Anything, "test-namespace", "backup-cronjob").Return(&application.CronJobExecution{
					ID:            "cje-123",
					ApplicationID: "app-123",
					JobName:       "backup-cronjob-manual-123",
					StartedAt:     time.Now(),
					Status:        application.CronJobExecutionStatusRunning,
				}, nil)

				// Create backup execution linked to cronjob execution
				backupSvc.On("CreateBackupExecution", mock.Anything, mock.MatchedBy(func(exec *backup.BackupExecution) bool {
					return exec.PolicyID == "policy-123" &&
						   exec.CronJobExecutionID == "cje-123" &&
						   exec.Status == backup.BackupExecutionStatusRunning
				})).Return(&backup.BackupExecution{
					ID:                 "be-123",
					PolicyID:           "policy-123",
					CronJobExecutionID: "cje-123",
					Status:             backup.BackupExecutionStatusRunning,
				}, nil)

				// Store cronjob execution
				appRepo.On("CreateCronJobExecution", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "trigger cronjob without backup policy",
			appID: "app-456",
			setupMocks: func(appRepo *MockApplicationRepository, k8sRepo *MockKubernetesRepository, backupSvc *MockBackupService) {
				// Get application without backup policy
				appRepo.On("GetByID", mock.Anything, "app-456").Return(&application.Application{
					ID:       "app-456",
					Name:     "regular-cronjob",
					Type:     application.ApplicationTypeCronJob,
					Metadata: map[string]string{},
					Status:   application.ApplicationStatusRunning,
				}, nil)

				// Trigger CronJob normally
				k8sRepo.On("TriggerCronJob", mock.Anything, mock.Anything, mock.Anything).Return(&application.CronJobExecution{
					ID:        "cje-456",
					StartedAt: time.Now(),
					Status:    application.CronJobExecutionStatusRunning,
				}, nil)

				// Store cronjob execution
				appRepo.On("CreateCronJobExecution", mock.Anything, mock.Anything).Return(nil)

				// No backup service calls expected
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			appRepo := new(MockApplicationRepository)
			k8sRepo := new(MockKubernetesRepository)
			projectService := new(MockProjectService)
			backupSvc := new(MockBackupService)

			// Configure project mock
			projectService.On("GetProject", mock.Anything, mock.Anything).Return(&project.Project{
				ID:        "proj-123",
				NamespaceName: "test-namespace",
			}, nil)

			// Setup specific test mocks
			if tt.setupMocks != nil {
				tt.setupMocks(appRepo, k8sRepo, backupSvc)
			}

			// Create base service
			baseService := &Service{
				repo:    appRepo,
				k8s:     k8sRepo,
				k8sRepo: k8sRepo,
				logger:  slog.Default(),
			}
			
			// Create extended service
			svc := &ExtendedService{
				Service:        baseService,
				projectService: projectService,
				backupService:  backupSvc,
			}

			// Execute
			ctx := context.Background()
			execution, err := svc.TriggerCronJob(ctx, &application.TriggerCronJobRequest{
				ApplicationID: tt.appID,
			})

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, execution)
			}

			// Verify mock expectations
			appRepo.AssertExpectations(t)
			k8sRepo.AssertExpectations(t)
			backupSvc.AssertExpectations(t)
		})
	}
}

func TestService_UpdateCronJobExecutionWithBackupStatus(t *testing.T) {
	tests := []struct {
		name       string
		execID     string
		jobStatus  application.CronJobExecutionStatus
		setupMocks func(*MockApplicationRepository, *MockBackupService)
		wantErr    bool
	}{
		{
			name:      "update cronjob execution success - trigger backup success",
			execID:    "cje-123",
			jobStatus: application.CronJobExecutionStatusSucceeded,
			setupMocks: func(appRepo *MockApplicationRepository, backupSvc *MockBackupService) {
				// Get execution with backup
				appRepo.On("GetCronJobExecution", mock.Anything, "cje-123").Return(&application.CronJobExecution{
					ID:            "cje-123",
					ApplicationID: "app-123",
					Status:        application.CronJobExecutionStatusRunning,
				}, nil)

				// Get backup execution
				backupSvc.On("GetBackupExecutionByCronJobID", mock.Anything, "cje-123").Return(&backup.BackupExecution{
					ID:                 "be-123",
					CronJobExecutionID: "cje-123",
					Status:             backup.BackupExecutionStatusRunning,
				}, nil)

				// Update backup status to succeeded
				backupSvc.On("UpdateBackupExecutionStatus", mock.Anything, "be-123", backup.BackupExecutionStatusSucceeded).Return(nil)

				// Update cronjob execution
				appRepo.On("UpdateCronJobExecution", mock.Anything, "cje-123", mock.MatchedBy(func(exec *application.CronJobExecution) bool {
					return exec.Status == application.CronJobExecutionStatusSucceeded
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "update cronjob execution failed - mark backup as failed",
			execID:    "cje-456",
			jobStatus: application.CronJobExecutionStatusFailed,
			setupMocks: func(appRepo *MockApplicationRepository, backupSvc *MockBackupService) {
				// Get execution
				appRepo.On("GetCronJobExecution", mock.Anything, "cje-456").Return(&application.CronJobExecution{
					ID:     "cje-456",
					Status: application.CronJobExecutionStatusRunning,
				}, nil)

				// Get backup execution
				backupSvc.On("GetBackupExecutionByCronJobID", mock.Anything, "cje-456").Return(&backup.BackupExecution{
					ID:     "be-456",
					Status: backup.BackupExecutionStatusRunning,
				}, nil)

				// Update backup status to failed
				backupSvc.On("UpdateBackupExecutionStatus", mock.Anything, "be-456", backup.BackupExecutionStatusFailed).Return(nil)

				// Update cronjob execution
				appRepo.On("UpdateCronJobExecution", mock.Anything, "cje-456", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			appRepo := new(MockApplicationRepository)
			backupSvc := new(MockBackupService)

			// Setup specific test mocks
			if tt.setupMocks != nil {
				tt.setupMocks(appRepo, backupSvc)
			}

			// Create base service
			baseService := &Service{
				repo:   appRepo,
				logger: slog.Default(),
			}
			
			// Create extended service
			svc := &ExtendedService{
				Service:       baseService,
				backupService: backupSvc,
			}

			// Execute
			ctx := context.Background()
			err := svc.UpdateCronJobExecutionStatus(ctx, tt.execID, tt.jobStatus)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			appRepo.AssertExpectations(t)
			backupSvc.AssertExpectations(t)
		})
	}
}

func TestService_ValidateBackupScheduleCompatibility(t *testing.T) {
	tests := []struct {
		name           string
		cronSchedule   string
		backupSchedule string
		wantErr        bool
		errMessage     string
	}{
		{
			name:           "compatible schedules - backup after cronjob",
			cronSchedule:   "0 2 * * *", // 2 AM
			backupSchedule: "0 4 * * *", // 4 AM
			wantErr:        false,
		},
		{
			name:           "incompatible schedules - backup before cronjob",
			cronSchedule:   "0 4 * * *", // 4 AM
			backupSchedule: "0 2 * * *", // 2 AM
			wantErr:        true,
			errMessage:     "backup schedule must run after cronjob",
		},
		{
			name:           "same schedule time",
			cronSchedule:   "0 2 * * *",
			backupSchedule: "0 2 * * *",
			wantErr:        true,
			errMessage:     "backup schedule must have different time",
		},
		{
			name:           "weekly schedules",
			cronSchedule:   "0 2 * * 0", // Sunday 2 AM
			backupSchedule: "0 4 * * 0", // Sunday 4 AM
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &ExtendedService{Service: &Service{logger: slog.Default()}}

			err := svc.validateBackupScheduleCompatibility(tt.cronSchedule, tt.backupSchedule)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}