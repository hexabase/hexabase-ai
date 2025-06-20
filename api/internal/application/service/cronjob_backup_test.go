package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	monitoringDomain "github.com/hexabase/hexabase-ai/api/internal/monitoring/domain"
	projectDomain "github.com/hexabase/hexabase-ai/api/internal/project/domain"
)

// Extend existing mocks with additional methods if needed
// These should be in a shared mocks file, but for now we'll add the missing methods

// MockProjectService for project service
type MockProjectService struct {
	mock.Mock
}

func (m *MockProjectService) GetProject(ctx context.Context, projectID string) (*projectDomain.Project, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.Project), args.Error(1)
}

func (m *MockProjectService) CreateProject(ctx context.Context, req *projectDomain.CreateProjectRequest) (*projectDomain.Project, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.Project), args.Error(1)
}

func (m *MockProjectService) ListProjects(ctx context.Context, filter projectDomain.ProjectFilter) (*projectDomain.ProjectList, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.ProjectList), args.Error(1)
}

func (m *MockProjectService) UpdateProject(ctx context.Context, projectID string, req *projectDomain.UpdateProjectRequest) (*projectDomain.Project, error) {
	args := m.Called(ctx, projectID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.Project), args.Error(1)
}

func (m *MockProjectService) DeleteProject(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockProjectService) GetProjectStats(ctx context.Context, projectID string) (*projectDomain.ProjectStats, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.ProjectStats), args.Error(1)
}

func (m *MockProjectService) CreateSubProject(ctx context.Context, parentID string, req *projectDomain.CreateProjectRequest) (*projectDomain.Project, error) {
	args := m.Called(ctx, parentID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.Project), args.Error(1)
}

func (m *MockProjectService) GetProjectHierarchy(ctx context.Context, projectID string) (*projectDomain.ProjectHierarchy, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.ProjectHierarchy), args.Error(1)
}

func (m *MockProjectService) ApplyResourceQuota(ctx context.Context, projectID string, quota *projectDomain.ResourceQuota) error {
	args := m.Called(ctx, projectID, quota)
	return args.Error(0)
}

func (m *MockProjectService) GetResourceUsage(ctx context.Context, projectID string) (*projectDomain.ResourceUsage, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.ResourceUsage), args.Error(1)
}

func (m *MockProjectService) CreateNamespace(ctx context.Context, projectID string, req *projectDomain.CreateNamespaceRequest) (*projectDomain.Namespace, error) {
	args := m.Called(ctx, projectID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.Namespace), args.Error(1)
}

func (m *MockProjectService) GetNamespace(ctx context.Context, projectID, namespaceID string) (*projectDomain.Namespace, error) {
	args := m.Called(ctx, projectID, namespaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.Namespace), args.Error(1)
}

func (m *MockProjectService) ListNamespaces(ctx context.Context, projectID string) (*projectDomain.NamespaceList, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.NamespaceList), args.Error(1)
}

func (m *MockProjectService) UpdateNamespace(ctx context.Context, projectID, namespaceID string, req *projectDomain.CreateNamespaceRequest) (*projectDomain.Namespace, error) {
	args := m.Called(ctx, projectID, namespaceID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.Namespace), args.Error(1)
}

func (m *MockProjectService) DeleteNamespace(ctx context.Context, projectID, namespaceID string) error {
	args := m.Called(ctx, projectID, namespaceID)
	return args.Error(0)
}

func (m *MockProjectService) GetNamespaceUsage(ctx context.Context, projectID, namespaceID string) (*projectDomain.NamespaceUsage, error) {
	args := m.Called(ctx, projectID, namespaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.NamespaceUsage), args.Error(1)
}

func (m *MockProjectService) AddMember(ctx context.Context, projectID, adderID string, req *projectDomain.AddMemberRequest) (*projectDomain.ProjectMember, error) {
	args := m.Called(ctx, projectID, adderID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.ProjectMember), args.Error(1)
}

func (m *MockProjectService) GetMember(ctx context.Context, projectID, memberID string) (*projectDomain.ProjectMember, error) {
	args := m.Called(ctx, projectID, memberID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.ProjectMember), args.Error(1)
}

func (m *MockProjectService) ListMembers(ctx context.Context, projectID string) (*projectDomain.MemberList, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.MemberList), args.Error(1)
}

func (m *MockProjectService) UpdateMemberRole(ctx context.Context, projectID, memberID string, req *projectDomain.UpdateMemberRoleRequest) (*projectDomain.ProjectMember, error) {
	args := m.Called(ctx, projectID, memberID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.ProjectMember), args.Error(1)
}

func (m *MockProjectService) RemoveMember(ctx context.Context, projectID, memberID, removerID string) error {
	args := m.Called(ctx, projectID, memberID, removerID)
	return args.Error(0)
}

func (m *MockProjectService) AddProjectMember(ctx context.Context, projectID string, req *projectDomain.AddMemberRequest) error {
	args := m.Called(ctx, projectID, req)
	return args.Error(0)
}

func (m *MockProjectService) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	args := m.Called(ctx, projectID, userID)
	return args.Error(0)
}

func (m *MockProjectService) ListProjectMembers(ctx context.Context, projectID string) ([]*projectDomain.ProjectMember, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*projectDomain.ProjectMember), args.Error(1)
}

func (m *MockProjectService) ListActivities(ctx context.Context, projectID string, limit int) (*projectDomain.ActivityList, error) {
	args := m.Called(ctx, projectID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*projectDomain.ActivityList), args.Error(1)
}

func (m *MockProjectService) LogActivity(ctx context.Context, activity *projectDomain.ProjectActivity) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

func (m *MockProjectService) GetActivityLogs(ctx context.Context, projectID string, filter projectDomain.ActivityFilter) ([]*projectDomain.ProjectActivity, error) {
	args := m.Called(ctx, projectID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*projectDomain.ProjectActivity), args.Error(1)
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

func (m *MockMonitoringService) GetWorkspaceMetrics(ctx context.Context, workspaceID string, opts monitoringDomain.QueryOptions) (*monitoringDomain.WorkspaceMetrics, error) {
	args := m.Called(ctx, workspaceID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*monitoringDomain.WorkspaceMetrics), args.Error(1)
}

func (m *MockMonitoringService) GetClusterHealth(ctx context.Context, workspaceID string) (*monitoringDomain.ClusterHealth, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*monitoringDomain.ClusterHealth), args.Error(1)
}

func (m *MockMonitoringService) GetResourceUsage(ctx context.Context, workspaceID string) (*monitoringDomain.ResourceUsage, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*monitoringDomain.ResourceUsage), args.Error(1)
}

func (m *MockMonitoringService) GetAlerts(ctx context.Context, workspaceID, severity string) ([]*monitoringDomain.Alert, error) {
	args := m.Called(ctx, workspaceID, severity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*monitoringDomain.Alert), args.Error(1)
}

func (m *MockMonitoringService) CreateAlert(ctx context.Context, alert *monitoringDomain.Alert) error {
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
		req               *domain.CreateApplicationRequest
		backupPolicyReq   *backup.CreateBackupPolicyRequest
		setupMocks        func(*MockApplicationRepository, *MockKubernetesRepository, *MockBackupService)
		wantErr           bool
		errMessage        string
	}{
		{
			name: "create cronjob with backup policy - success",
			req: &domain.CreateApplicationRequest{
				Name:      "backup-cronjob",
				Type:      domain.ApplicationTypeCronJob,
				ProjectID: "proj-123",
				Source: domain.ApplicationSource{
					Type:  domain.SourceTypeImage,
					Image: "backup-tool:latest",
				},
				Config: domain.ApplicationConfig{
					Environment: map[string]string{
						"BACKUP_TARGET": "database",
					},
				},
				CronSchedule: "0 2 * * *",
				CronCommand:  []string{"/bin/backup.sh"},
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
				// Check if application already exists (should return nil, indicating not found)
				appRepo.On("GetApplicationByName", mock.Anything, "workspace-123", "proj-123", "backup-cronjob").Return(nil, nil)
				
				// Create application
				appRepo.On("CreateApplication", mock.Anything, mock.AnythingOfType("*domain.Application")).Return(nil)

				// Create CronJob in Kubernetes
				k8sRepo.On("CreateCronJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				// Create backup policy
				backupSvc.On("CreateBackupPolicy", mock.Anything, mock.AnythingOfType("string"), mock.MatchedBy(func(req backup.CreateBackupPolicyRequest) bool {
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
				appRepo.On("UpdateApplication", mock.Anything, mock.AnythingOfType("*domain.Application")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "create cronjob with backup policy - validate schedule compatibility",
			req: &domain.CreateApplicationRequest{
				Name:         "backup-cronjob",
				Type:         domain.ApplicationTypeCronJob,
				ProjectID:    "proj-123",
				CronSchedule: "0 2 * * *",
				CronCommand:  []string{"/bin/backup.sh"},
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
			errMessage: "backup schedule must run after cronjob schedule",
		},
		{
			name: "create cronjob with backup hooks",
			req: &domain.CreateApplicationRequest{
				Name:         "db-backup-cronjob",
				Type:         domain.ApplicationTypeCronJob,
				ProjectID:    "proj-123",
				Source: domain.ApplicationSource{
					Type:  domain.SourceTypeImage,
					Image: "db-backup:latest",
				},
				CronSchedule: "0 2 * * *",
				CronCommand:  []string{"/bin/backup.sh"},
			},
			backupPolicyReq: &backup.CreateBackupPolicyRequest{
				StorageID:      "storage-123",
				Schedule:       "0 4 * * *",
				RetentionDays:  30,
				PreBackupHook:  "kubectl exec -n {namespace} {pod} -- /scripts/pre-backup.sh",
				PostBackupHook: "kubectl exec -n {namespace} {pod} -- /scripts/post-backup.sh",
			},
			setupMocks: func(appRepo *MockApplicationRepository, k8sRepo *MockKubernetesRepository, backupSvc *MockBackupService) {
				// Check if application already exists
				appRepo.On("GetApplicationByName", mock.Anything, "workspace-123", "proj-123", "db-backup-cronjob").Return(nil, nil)
				
				appRepo.On("CreateApplication", mock.Anything, mock.AnythingOfType("*domain.Application")).Return(nil)
				k8sRepo.On("CreateCronJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				
				backupSvc.On("CreateBackupPolicy", mock.Anything, mock.AnythingOfType("string"), mock.MatchedBy(func(req backup.CreateBackupPolicyRequest) bool {
					return req.PreBackupHook != "" && req.PostBackupHook != ""
				})).Return(&backup.BackupPolicy{
					ID:             "policy-123",
					PreBackupHook:  "kubectl exec -n {namespace} {pod} -- /scripts/pre-backup.sh",
					PostBackupHook: "kubectl exec -n {namespace} {pod} -- /scripts/post-backup.sh",
				}, nil)
				
				appRepo.On("UpdateApplication", mock.Anything, mock.AnythingOfType("*domain.Application")).Return(nil)
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
			projectService.On("GetProject", mock.Anything, "proj-123").Return(&projectDomain.Project{
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
			app, err := svc.CreateApplicationWithBackupPolicy(ctx, "workspace-123", tt.req, tt.backupPolicyReq)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app)
				if tt.backupPolicyReq != nil {
					assert.NotNil(t, app.Metadata)
					assert.Equal(t, "true", app.Metadata["backup_enabled"])
				}
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
				appRepo.On("GetApplication", mock.Anything, "app-123").Return(&domain.Application{
					ID:   "app-123",
					Name: "backup-cronjob",
					Type: domain.ApplicationTypeCronJob,
					Metadata: map[string]string{
						"backup_enabled":    "true",
						"backup_policy_id": "policy-123",
					},
					Status: domain.ApplicationStatusRunning,
				}, nil)

				// Trigger CronJob
				k8sRepo.On("TriggerCronJob", mock.Anything, mock.Anything, mock.Anything, "backup-cronjob").Return(nil)

				// Create backup execution linked to cronjob execution
				backupSvc.On("TriggerManualBackup", mock.Anything, mock.Anything, mock.MatchedBy(func(req backup.TriggerBackupRequest) bool {
					return req.ApplicationID == "app-123" &&
						   req.BackupType == backup.BackupTypeFull
				})).Return(&backup.BackupExecution{
					ID:                 "be-123",
					
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
				appRepo.On("GetApplication", mock.Anything, "app-456").Return(&domain.Application{
					ID:       "app-456",
					Name:     "regular-cronjob",
					Type:     domain.ApplicationTypeCronJob,
					Metadata: map[string]string{},
					Status:   domain.ApplicationStatusRunning,
				}, nil)

				// Trigger CronJob normally
				k8sRepo.On("TriggerCronJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
			projectService.On("GetProject", mock.Anything, mock.Anything).Return(&projectDomain.Project{
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
			execution, err := svc.TriggerCronJob(ctx, "workspace-123", &domain.TriggerCronJobRequest{
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
