package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
)

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
		policy, err := s.backupService.CreateBackupPolicy(ctx, app.ID, backupPolicyReq)
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

	// Get project for namespace
	project, err := s.projectRepo.GetByID(ctx, app.ProjectID)
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
	
	if err := s.k8s.TriggerCronJob(ctx, project.Namespace, app.Name, app.ID); err != nil {
		return nil, fmt.Errorf("failed to trigger CronJob: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to trigger CronJob: %w", err)
	}

	// Store execution record
	if err := s.repo.CreateCronJobExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to store CronJob execution: %w", err)
	}

	// Check if backup is enabled
	if app.Metadata != nil && app.Metadata["backup_enabled"] == "true" {
		policyID := app.Metadata["backup_policy_id"]
		if policyID != "" {
			// Create backup execution linked to CronJob execution
			backupExec := &backup.BackupExecution{
				PolicyID:           policyID,
				CronJobExecutionID: execution.ID,
				Status:             backup.BackupExecutionStatusRunning,
				StartedAt:          time.Now(),
				Metadata: map[string]interface{}{
					"triggered_by": "cronjob",
					"cronjob_name": app.Name,
				},
			}

			_, err := s.backupService.CreateBackupExecution(ctx, backupExec)
			if err != nil {
				// Log error but don't fail the CronJob trigger
				s.logger.Error("failed to create backup execution",
					"error", err,
					"cronjob_execution_id", execution.ID,
					"policy_id", policyID)
			}
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

	// Check if there's an associated backup execution
	backupExec, err := s.backupService.GetBackupExecutionByCronJobID(ctx, executionID)
	if err != nil {
		// No backup execution found, which is fine
		return nil
	}

	// Update backup status based on CronJob status
	var backupStatus backup.BackupExecutionStatus
	switch status {
	case application.CronJobExecutionStatusSucceeded:
		backupStatus = backup.BackupExecutionStatusSucceeded
	case application.CronJobExecutionStatusFailed:
		backupStatus = backup.BackupExecutionStatusFailed
	default:
		// Keep current status
		return nil
	}

	if err := s.backupService.UpdateBackupExecutionStatus(ctx, backupExec.ID, backupStatus); err != nil {
		s.logger.Error("failed to update backup execution status",
			"error", err,
			"backup_execution_id", backupExec.ID,
			"new_status", backupStatus)
	}

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