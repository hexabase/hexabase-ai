package application

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

	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
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
		app := &application.Application{
			ID:            "app-123",
			WorkspaceID:   "ws-123",
			ProjectID:     "proj-123",
			Name:          "test-cronjob",
			Type:          application.ApplicationTypeCronJob,
			Status:        application.ApplicationStatusPending,
			CronSchedule:  "0 */6 * * *", // Every 6 hours
			CronCommand:   []string{"/bin/backup.sh"},
			CronArgs:      []string{"--compress", "--incremental"},
			TemplateAppID: "app-template-123",
			Source: application.ApplicationSource{
				Type:  application.SourceTypeImage,
				Image: "backup-tool:latest",
			},
			Config: application.ApplicationConfig{
				Replicas: 1,
				Resources: application.ResourceRequests{
					CPURequest:    "100m",
					CPULimit:      "500m",
					MemoryRequest: "128Mi",
					MemoryLimit:   "512Mi",
				},
			},
		}

		// Expect template app lookup
		templateRows := sqlmock.NewRows([]string{
			"id", "workspace_id", "project_id", "name", "type", "status",
			"source_type", "source_image", "config", "endpoints",
			"cron_schedule", "template_app_id", "created_at", "updated_at",
		}).AddRow(
			"app-template-123", "ws-123", "proj-123", "backup-base", "stateless", "running",
			"image", "backup-tool:latest", []byte(`{}`), []byte(`[]`),
			nil, nil, time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT \* FROM "applications" WHERE id = \$1`).
			WithArgs("app-template-123").
			WillReturnRows(templateRows)

		// Expect CronJob creation
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "applications"`).
			WithArgs(
				"app-123", "ws-123", "proj-123", "test-cronjob", "cronjob", "pending",
				"image", "backup-tool:latest", "", "",
				sqlmock.AnyArg(), sqlmock.AnyArg(),
				"0 */6 * * *",
				pq.Array([]string{"/bin/backup.sh"}),
				pq.Array([]string{"--compress", "--incremental"}),
				"app-template-123",
				nil, nil,
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("app-123"))
		mock.ExpectCommit()

		err := repo.Create(ctx, app)
		assert.NoError(t, err)
		assert.Equal(t, "app-123", app.ID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Create CronJob without Template", func(t *testing.T) {
		app := &application.Application{
			WorkspaceID:  "ws-123",
			ProjectID:    "proj-123",
			Name:         "standalone-cronjob",
			Type:         application.ApplicationTypeCronJob,
			Status:       application.ApplicationStatusPending,
			CronSchedule: "0 0 * * *", // Daily at midnight
			Source: application.ApplicationSource{
				Type:  application.SourceTypeImage,
				Image: "alpine:latest",
			},
			Config: application.ApplicationConfig{
				Replicas: 1,
			},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "applications"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("app-generated-id"))
		mock.ExpectCommit()

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
		mock.ExpectQuery(`SELECT count\(\*\) FROM "cronjob_executions"`).
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

		mock.ExpectQuery(`SELECT \* FROM "cronjob_executions"`).
			WithArgs(appID, limit, offset).
			WillReturnRows(executionRows)

		executions, total, err := repo.GetCronJobExecutions(ctx, appID, limit, offset)
		assert.NoError(t, err)
		assert.Equal(t, 25, total)
		assert.Len(t, executions, 2)

		assert.Equal(t, "cje-1", executions[0].ID)
		assert.Equal(t, application.CronJobExecutionStatusSucceeded, executions[0].Status)
		assert.NotNil(t, executions[0].CompletedAt)
		assert.Equal(t, 0, *executions[0].ExitCode)

		assert.Equal(t, "cje-2", executions[1].ID)
		assert.Equal(t, application.CronJobExecutionStatusRunning, executions[1].Status)
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
		execution := &application.CronJobExecution{
			ApplicationID: "app-123",
			JobName:       "test-cronjob-28474952",
			StartedAt:     time.Now(),
			Status:        application.CronJobExecutionStatusRunning,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "cronjob_executions"`).
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
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("cje-generated"))
		mock.ExpectCommit()

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

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "cronjob_executions" SET`).
			WithArgs(
				&completedAt,
				string(application.CronJobExecutionStatusSucceeded),
				&exitCode,
				logs,
				sqlmock.AnyArg(), // UpdatedAt
				executionID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateCronJobExecution(ctx, executionID, &completedAt, 
			application.CronJobExecutionStatusSucceeded, &exitCode, logs)
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

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "applications" SET`).
			WithArgs(
				newSchedule,
				sqlmock.AnyArg(), // UpdatedAt
				appID,
				"cronjob",
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateCronSchedule(ctx, appID, newSchedule)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}