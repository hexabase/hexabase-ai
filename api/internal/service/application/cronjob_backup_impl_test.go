package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
)

// Simple mock for backup policy creation
type mockBackupService struct {
	mock.Mock
}

func (m *mockBackupService) CreateBackupPolicy(ctx context.Context, applicationID string, req *backup.CreateBackupPolicyRequest) (*backup.BackupPolicy, error) {
	args := m.Called(ctx, applicationID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.BackupPolicy), args.Error(1)
}

// Simple mock for project repository
type mockProjectRepository struct {
	mock.Mock
}

func (m *mockProjectRepository) Get(ctx context.Context, id string) (*Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Project), args.Error(1)
}

func TestValidateBackupScheduleCompatibility(t *testing.T) {
	svc := &ExtendedService{}

	tests := []struct {
		name           string
		cronSchedule   string
		backupSchedule string
		expectError    bool
		errorMessage   string
	}{
		{
			name:           "valid - backup after cronjob",
			cronSchedule:   "0 2 * * *", // 2 AM
			backupSchedule: "0 4 * * *", // 4 AM
			expectError:    false,
		},
		{
			name:           "invalid - backup before cronjob",
			cronSchedule:   "0 4 * * *", // 4 AM
			backupSchedule: "0 2 * * *", // 2 AM
			expectError:    true,
			errorMessage:   "backup schedule must run after cronjob",
		},
		{
			name:           "invalid - same time",
			cronSchedule:   "0 2 * * *",
			backupSchedule: "0 2 * * *",
			expectError:    true,
			errorMessage:   "different time",
		},
		{
			name:           "valid - weekly schedules",
			cronSchedule:   "0 2 * * 0", // Sunday 2 AM
			backupSchedule: "0 4 * * 0", // Sunday 4 AM
			expectError:    false,
		},
		{
			name:           "invalid format - missing fields",
			cronSchedule:   "0 2",
			backupSchedule: "0 4 * * *",
			expectError:    true,
			errorMessage:   "invalid cron expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateBackupScheduleCompatibility(tt.cronSchedule, tt.backupSchedule)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCronJobBackupIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("create cronjob with backup policy succeeds", func(t *testing.T) {
		// Setup
		appRepo := new(MockRepository)
		k8sRepo := new(MockKubernetesRepository)
		backupSvc := &mockBackupService{}
		projectRepo := &mockProjectRepository{}

		baseService := NewService(appRepo, k8sRepo)
		extSvc := &ExtendedService{
			Service:       baseService.(*Service),
			projectRepo:   projectRepo,
			backupService: backupSvc,
		}

		// Test data
		req := &application.CreateApplicationRequest{
			Name:      "backup-cronjob",
			Type:      application.ApplicationTypeCronJob,
			ProjectID: "proj-123",
			Source: application.ApplicationSource{
				Type:  application.SourceTypeImage,
				Image: "backup-tool:latest",
			},
			Config: application.ApplicationConfig{
				Replicas: 1,
			},
			CronSchedule: "0 2 * * *",
			Metadata: map[string]string{
				"test": "true",
			},
		}

		backupPolicyReq := &backup.CreateBackupPolicyRequest{
			StorageID:     "storage-123",
			Schedule:      "0 4 * * *",
			RetentionDays: 30,
		}

		// Mock expectations
		appRepo.On("CreateApplication", mock.Anything, mock.MatchedBy(func(app *application.Application) bool {
			return app.Name == "backup-cronjob" && 
				   app.Type == application.ApplicationTypeCronJob &&
				   app.Metadata["backup_enabled"] == "true"
		})).Return(nil).Once()

		appRepo.On("GetApplicationByName", mock.Anything, "", "proj-123", "backup-cronjob").
			Return(nil, errors.New("not found")).Once()

		k8sRepo.On("CreateCronJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Once()

		backupSvc.On("CreateBackupPolicy", mock.Anything, mock.Anything, backupPolicyReq).
			Return(&backup.BackupPolicy{
				ID:        "policy-123",
				StorageID: "storage-123",
			}, nil).Once()

		appRepo.On("UpdateApplication", mock.Anything, mock.MatchedBy(func(app *application.Application) bool {
			return app.Metadata["backup_policy_id"] == "policy-123"
		})).Return(nil).Once()

		// Execute
		app, err := extSvc.CreateApplicationWithBackupPolicy(ctx, req, backupPolicyReq)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, app)
		assert.Equal(t, "true", app.Metadata["backup_enabled"])
		assert.Equal(t, "policy-123", app.Metadata["backup_policy_id"])

		// Verify mock expectations
		appRepo.AssertExpectations(t)
		k8sRepo.AssertExpectations(t)
		backupSvc.AssertExpectations(t)
	})

	t.Run("schedule validation prevents incompatible times", func(t *testing.T) {
		// Setup
		appRepo := new(MockRepository)
		k8sRepo := new(MockKubernetesRepository)
		backupSvc := &mockBackupService{}

		baseService := NewService(appRepo, k8sRepo)
		extSvc := &ExtendedService{
			Service:       baseService.(*Service),
			backupService: backupSvc,
		}

		// Test data - backup scheduled before cronjob
		req := &application.CreateApplicationRequest{
			Name:         "backup-cronjob",
			Type:         application.ApplicationTypeCronJob,
			ProjectID:    "proj-123",
			CronSchedule: "0 4 * * *", // 4 AM
		}

		backupPolicyReq := &backup.CreateBackupPolicyRequest{
			StorageID:     "storage-123",
			Schedule:      "0 2 * * *", // 2 AM - before cronjob
			RetentionDays: 30,
		}

		// Execute
		_, err := extSvc.CreateApplicationWithBackupPolicy(ctx, req, backupPolicyReq)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "backup schedule must run after cronjob")
	})
}