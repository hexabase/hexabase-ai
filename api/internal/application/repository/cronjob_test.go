package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return gormDB, mock
}

func TestPostgresRepository_CreateCronJob(t *testing.T) {
	gormDB, mock := setupTestDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	t.Run("Create CronJob with Template", func(t *testing.T) {
		app := &domain.Application{
			ID:            "app-123",
			WorkspaceID:   "ws-123",
			ProjectID:     "proj-123",
			Name:          "test-cronjob",
			Type:          domain.ApplicationTypeCronJob,
			Status:        domain.ApplicationStatusPending,
			CronSchedule:  "0 */6 * * *", // Every 6 hours
			CronCommand:   []string{"/bin/backup.sh"},
			CronArgs:      []string{"--compress", "--incremental"},
			TemplateAppID: "app-template-123",
			Source: domain.ApplicationSource{
				Type:  domain.SourceTypeImage,
				Image: "backup-tool:latest",
			},
			Config: domain.ApplicationConfig{
				Replicas: 1,
				Resources: domain.ResourceRequests{
					CPURequest:    "100m",
					CPULimit:      "500m",
					MemoryRequest: "128Mi",
					MemoryLimit:   "512Mi",
				},
			},
		}

		// Expect template app lookup
		configJSON := `{"replicas":1,"port":0,"resources":{"cpu_request":"100m","memory_request":"256Mi"}}`
		endpointsJSON := `[]`
		
		templateRows := sqlmock.NewRows([]string{
			"id", "workspace_id", "project_id", "name", "type", "status",
			"source_type", "source_image", "source_git_url", "source_git_ref",
			"config", "endpoints", 
			"cron_schedule", "cron_command", "cron_args", "template_app_id",
			"last_execution_at", "next_execution_at", "created_at", "updated_at",
			"function_runtime", "function_handler", "function_timeout", "function_memory",
			"function_trigger_type", "function_trigger_config", "function_env_vars", "function_secrets",
		}).AddRow(
			"app-template-123", "ws-123", "proj-123", "backup-base", "stateless", "running",
			"image", "backup-tool:latest", "", "",
			configJSON, endpointsJSON,
			nil, nil, nil, nil,
			nil, nil, time.Now(), time.Now(),
			nil, nil, nil, nil,
			nil, nil, nil, nil,
		)

		mock.ExpectQuery(`SELECT \* FROM "applications" WHERE id = \$1 ORDER BY "applications"\."id" LIMIT \$2`).
			WithArgs("app-template-123", 1).
			WillReturnRows(templateRows)

		// Expect CronJob creation
		mock.ExpectExec(`INSERT INTO "applications"`).
			WithArgs(
				sqlmock.AnyArg(), // id
				"ws-123", "proj-123", "test-cronjob", "cronjob", "pending",
				"image", "backup-tool:latest", "", "",
				sqlmock.AnyArg(), sqlmock.AnyArg(),
				"0 */6 * * *",
				pq.Array([]string{"/bin/backup.sh"}),
				pq.Array([]string{"--compress", "--incremental"}),
				"app-template-123",
				nil, nil,
				sqlmock.AnyArg(), sqlmock.AnyArg(),
				nil, nil, nil, nil, nil, nil, nil, nil, // function fields
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, app)
		assert.NoError(t, err)
		assert.Equal(t, "app-123", app.ID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Create CronJob without Template", func(t *testing.T) {
		app := &domain.Application{
			WorkspaceID:  "ws-123",
			ProjectID:    "proj-123",
			Name:         "standalone-cronjob",
			Type:         domain.ApplicationTypeCronJob,
			Status:       domain.ApplicationStatusPending,
			CronSchedule: "0 0 * * *", // Daily at midnight
			Source: domain.ApplicationSource{
				Type:  domain.SourceTypeImage,
				Image: "alpine:latest",
			},
			Config: domain.ApplicationConfig{
				Replicas: 1,
			},
		}

		mock.ExpectExec(`INSERT INTO "applications"`).
			WithArgs(
				sqlmock.AnyArg(), // id
				"ws-123", "proj-123", "standalone-cronjob", "cronjob", "pending",
				"image", "alpine:latest", "", "",
				sqlmock.AnyArg(), sqlmock.AnyArg(),
				"0 0 * * *",
				nil, nil, nil, // no command/args/template
				nil, nil,
				sqlmock.AnyArg(), sqlmock.AnyArg(),
				nil, nil, nil, nil, nil, nil, nil, nil, // function fields
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, app)
		assert.NoError(t, err)
		assert.NotEmpty(t, app.ID)
	})
}

func TestPostgresRepository_GetCronJobExecutions(t *testing.T) {
	gormDB, mock := setupTestDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	t.Run("List CronJob Executions", func(t *testing.T) {
		appID := "app-123"
		limit := 10
		offset := 0

		// Mock count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
		mock.ExpectQuery(`SELECT count\(\*\) FROM "cron_job_executions" WHERE application_id = \$1`).
			WithArgs(appID).
			WillReturnRows(countRows)

		// Mock executions query
		now := time.Now()
		completedAt := now.Add(-5 * time.Minute)
		exitCode := 0

		executionRows := sqlmock.NewRows([]string{
			"id", "application_id", "job_name", "started_at", "completed_at",
			"status", "exit_code", "logs", "created_at", "updated_at",
		}).
			AddRow(
				"cje-1", appID, "test-cronjob-28474950", now.Add(-10*time.Minute), &completedAt,
				"succeeded", &exitCode, "Backup completed successfully", now.Add(-10*time.Minute), now.Add(-5*time.Minute),
			).
			AddRow(
				"cje-2", appID, "test-cronjob-28474951", now.Add(-5*time.Minute), nil,
				"running", nil, "", now.Add(-5*time.Minute), now.Add(-5*time.Minute),
			)

		mock.ExpectQuery(`SELECT \* FROM "cron_job_executions" WHERE application_id = \$1 ORDER BY started_at DESC LIMIT \$2`).
			WithArgs(appID, limit).
			WillReturnRows(executionRows)

		executions, total, err := repo.GetCronJobExecutions(ctx, appID, limit, offset)
		assert.NoError(t, err)
		assert.Equal(t, 25, total)
		assert.Len(t, executions, 2)

		assert.Equal(t, "cje-1", executions[0].ID)
		assert.Equal(t, domain.CronJobExecutionStatusSucceeded, executions[0].Status)
		assert.NotNil(t, executions[0].CompletedAt)
		assert.Equal(t, 0, *executions[0].ExitCode)

		assert.Equal(t, "cje-2", executions[1].ID)
		assert.Equal(t, domain.CronJobExecutionStatusRunning, executions[1].Status)
		assert.Nil(t, executions[1].CompletedAt)
		
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_CreateCronJobExecution(t *testing.T) {
	gormDB, mock := setupTestDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	t.Run("Create New Execution", func(t *testing.T) {
		execution := &domain.CronJobExecution{
			ApplicationID: "app-123",
			JobName:       "test-cronjob-28474952",
			StartedAt:     time.Now(),
			Status:        domain.CronJobExecutionStatusRunning,
		}

		mock.ExpectExec(`INSERT INTO "cron_job_executions"`).
			WithArgs(
				sqlmock.AnyArg(), // ID
				execution.ApplicationID,
				execution.JobName,
				sqlmock.AnyArg(), // StartedAt
				nil,              // CompletedAt
				string(execution.Status),
				nil, // ExitCode
				"",  // Logs
				sqlmock.AnyArg(), // CreatedAt
				sqlmock.AnyArg(), // UpdatedAt
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateCronJobExecution(ctx, execution)
		assert.NoError(t, err)
		assert.NotEmpty(t, execution.ID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_UpdateCronJobExecution(t *testing.T) {
	gormDB, mock := setupTestDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	t.Run("Update Execution Status", func(t *testing.T) {
		executionID := "cje-123"
		completedAt := time.Now()
		exitCode := 0
		logs := "Job completed successfully\nFiles backed up: 1024"

		// GORM generates UPDATE with fields in alphabetical order
		mock.ExpectExec(`UPDATE "cron_job_executions" SET`).
			WithArgs(
				&completedAt,         // completed_at
				&exitCode,            // exit_code
				logs,                 // logs
				string(domain.CronJobExecutionStatusSucceeded), // status
				sqlmock.AnyArg(),     // updated_at
				executionID,          // WHERE id = ?
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateCronJobExecution(ctx, executionID, &completedAt, 
			domain.CronJobExecutionStatusSucceeded, &exitCode, logs)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_UpdateCronSchedule(t *testing.T) {
	gormDB, mock := setupTestDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	t.Run("Update CronJob Schedule", func(t *testing.T) {
		appID := "app-123"
		newSchedule := "0 */3 * * *" // Every 3 hours

		// No transaction used in the implementation
		mock.ExpectExec(`UPDATE "applications" SET`).
			WithArgs(
				newSchedule,
				sqlmock.AnyArg(), // UpdatedAt
				appID,
				"cronjob",
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateCronSchedule(ctx, appID, newSchedule)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}