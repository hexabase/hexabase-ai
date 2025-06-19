package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	"github.com/hexabase/hexabase-ai/api/internal/domain/monitoring"
	projectDomain "github.com/hexabase/hexabase-ai/api/internal/project/domain"
)

// Update to use project.Service instead of direct repository
type ExtendedService struct {
	*Service
	projectService    projectDomain.Service  // Change from Repository to Service
	backupService     backup.Service
	monitoringService monitoring.Service
}

// CreateApplicationWithBackupPolicy creates a CronJob application with an associated backup policy
func (s *ExtendedService) CreateApplicationWithBackupPolicy(
	ctx context.Context,
	req *application.CreateApplicationRequest,
	backupPolicyReq *backup.CreateBackupPolicyRequest,
) (*application.Application, error) {
	// Validate CronJob type
	if req.Type != application.ApplicationTypeCronJob {
		return nil, fmt.Errorf("backup policy can only be created for CronJob applications")
	}

	// Validate schedule compatibility if backup policy is provided
	if backupPolicyReq != nil && req.CronSchedule != "" && backupPolicyReq.Schedule != "" {
		if err := s.validateBackupScheduleCompatibility(req.CronSchedule, backupPolicyReq.Schedule); err != nil {
			return nil, err
		}
	}

	// Initialize metadata if not exists
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}

	// Mark that backup is enabled
	if backupPolicyReq != nil {
		req.Metadata["backup_enabled"] = "true"
	}

	// Create the application
	app, err := s.CreateApplication(ctx, req.ProjectID, *req)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// Create backup policy if requested
	if backupPolicyReq != nil {
		// Pass the value, not pointer
		policy, err := s.backupService.CreateBackupPolicy(ctx, app.ID, *backupPolicyReq)
		if err != nil {
			// Rollback application creation
			_ = s.DeleteApplication(ctx, app.ID)
			return nil, fmt.Errorf("failed to create backup policy: %w", err)
		}

		// Update application metadata with backup policy ID
		app.Metadata["backup_policy_id"] = policy.ID
		if err := s.repo.UpdateApplication(ctx, app); err != nil {
			return nil, fmt.Errorf("failed to update application with backup policy: %w", err)
		}
	}

	return app, nil
}

// TriggerCronJob triggers a CronJob and creates backup execution if configured
func (s *ExtendedService) TriggerCronJob(ctx context.Context, req *application.TriggerCronJobRequest) (*application.CronJobExecution, error) {
	// Get application
	app, err := s.repo.GetApplication(ctx, req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	if app.Type != application.ApplicationTypeCronJob {
		return nil, fmt.Errorf("application is not a CronJob")
	}

	if app.Status != application.ApplicationStatusRunning {
		return nil, fmt.Errorf("CronJob is not in running state")
	}

	// Get project using service instead of repository
	project, err := s.projectService.GetProject(ctx, app.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Trigger the CronJob in Kubernetes
	execution := &application.CronJobExecution{
		ID:            "cje-" + uuid.New().String(),
		ApplicationID: app.ID,
		JobName:       app.Name + "-manual-" + time.Now().Format("20060102150405"),
		StartedAt:     time.Now(),
		Status:        application.CronJobExecutionStatusRunning,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	if err := s.k8s.TriggerCronJob(ctx, app.WorkspaceID, project.ID, app.Name); err != nil {
		return nil, fmt.Errorf("failed to trigger CronJob: %w", err)
	}

	// Store execution record
	if err := s.repo.CreateCronJobExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to store CronJob execution: %w", err)
	}

	// Check if backup is enabled
	if app.Metadata != nil && app.Metadata["backup_enabled"] == "true" {
		// Trigger backup through the backup service
		backupReq := backup.TriggerBackupRequest{
			ApplicationID: app.ID,
			BackupType:    backup.BackupTypeFull,
			Metadata: map[string]interface{}{
				"triggered_by":        "cronjob",
				"cronjob_name":        app.Name,
				"cronjob_execution_id": execution.ID,
			},
		}

		_, err := s.backupService.TriggerManualBackup(ctx, app.ID, backupReq)
		if err != nil {
			// Log error but don't fail the CronJob trigger
			s.logger.Error("failed to trigger backup",
				"error", err,
				"cronjob_execution_id", execution.ID,
				"application_id", app.ID)
		}
	}

	return execution, nil
}

// UpdateCronJobExecutionStatus updates CronJob execution status and related backup status
func (s *ExtendedService) UpdateCronJobExecutionStatus(
	ctx context.Context,
	executionID string,
	status application.CronJobExecutionStatus,
) error {
	// Get the execution
	execution, err := s.repo.GetCronJobExecution(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get CronJob execution: %w", err)
	}

	// Update execution status
	execution.Status = status
	execution.CompletedAt = timePtr(time.Now())
	
	if err := s.repo.UpdateCronJobExecution(ctx, executionID, execution.CompletedAt, status, nil, ""); err != nil {
		return fmt.Errorf("failed to update CronJob execution: %w", err)
	}

	// Note: Backup execution status is managed internally by the backup service
	// when triggered through TriggerManualBackup. The backup service will handle
	// its own status updates based on the actual backup operation results.

	return nil
}

// validateBackupScheduleCompatibility validates that backup schedule runs after CronJob
func (s *ExtendedService) validateBackupScheduleCompatibility(cronSchedule, backupSchedule string) error {
	// Simple validation: extract hour from cron expression
	// Format: "minute hour day month weekday"
	cronParts := strings.Fields(cronSchedule)
	backupParts := strings.Fields(backupSchedule)

	if len(cronParts) < 2 || len(backupParts) < 2 {
		return fmt.Errorf("invalid cron expression format")
	}

	// Compare hours (simple comparison for this example)
	cronHour := cronParts[1]
	backupHour := backupParts[1]

	// If both are specific hours (not wildcards)
	if cronHour != "*" && backupHour != "*" {
		var cronH, backupH int
		_, err1 := fmt.Sscanf(cronHour, "%d", &cronH)
		_, err2 := fmt.Sscanf(backupHour, "%d", &backupH)
		
		if err1 == nil && err2 == nil {
			if backupH <= cronH {
				return fmt.Errorf("backup schedule must run after cronjob schedule")
			}
		}
	}

	// Check if schedules are identical
	if cronSchedule == backupSchedule {
		return fmt.Errorf("backup schedule must have different time than cronjob")
	}

	return nil
}

// Helper function to get time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}