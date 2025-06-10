package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/monitoring"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestPostgresRepository_CreateAlert(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("successful alert creation", func(t *testing.T) {
		alert := &monitoring.Alert{
			ID:          uuid.New().String(),
			WorkspaceID: "ws-123",
			Type:        "cpu",
			Severity:    "warning",
			Title:       "High CPU Usage",
			Description: "CPU usage above 80%",
			Resource:    "node-1",
			Threshold:   80.0,
			Value:       85.5,
			Status:      "active",
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "alert_records"`).
			WithArgs(
				alert.ID,
				alert.WorkspaceID,
				alert.Type,
				alert.Severity,
				alert.Title,
				alert.Description,
				alert.Resource,
				alert.Threshold,
				alert.Value,
				alert.Status,
				sqlmock.AnyArg(), // created_at
				alert.ResolvedAt,
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateAlert(ctx, alert)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetAlert(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("get alert by ID", func(t *testing.T) {
		alertID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "type", "severity", "title", "description",
			"resource", "threshold", "value", "status", "created_at", "resolved_at", "updated_at",
		}).AddRow(
			alertID, "ws-123", "memory", "critical", "Memory Alert", "Memory usage above 90%",
			"pod-xyz", 90.0, 95.5, "active", now, nil, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "alert_records" WHERE id = \$1 ORDER BY "alert_records"\."id" LIMIT \$2`).
			WithArgs(alertID, 1).
			WillReturnRows(rows)

		alert, err := repo.GetAlert(ctx, alertID)
		assert.NoError(t, err)
		assert.NotNil(t, alert)
		assert.Equal(t, alertID, alert.ID)
		assert.Equal(t, "memory", alert.Type)
		assert.Equal(t, "critical", alert.Severity)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetAlerts(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("list workspace alerts with filter", func(t *testing.T) {
		workspaceID := "ws-456"
		now := time.Now()
		filter := monitoring.AlertFilter{
			Severity: "warning",
			Status:   "active",
			Limit:    10,
		}

		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "type", "severity", "title", "description",
			"resource", "threshold", "value", "status", "created_at", "resolved_at", "updated_at",
		}).
			AddRow(uuid.New().String(), workspaceID, "cpu", "warning", "CPU Alert", "High CPU", "node-1", 75.0, 80.0, "active", now, nil, now).
			AddRow(uuid.New().String(), workspaceID, "disk", "warning", "Disk Alert", "Low disk", "node-2", 85.0, 90.0, "active", now, nil, now)

		mock.ExpectQuery(`SELECT \* FROM "alert_records" WHERE workspace_id = \$1 AND severity = \$2 AND status = \$3 ORDER BY created_at DESC LIMIT \$4`).
			WithArgs(workspaceID, filter.Severity, filter.Status, filter.Limit).
			WillReturnRows(rows)

		alerts, err := repo.GetAlerts(ctx, workspaceID, filter)
		assert.NoError(t, err)
		assert.Len(t, alerts, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_UpdateAlert(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("update alert status", func(t *testing.T) {
		alert := &monitoring.Alert{
			ID:          uuid.New().String(),
			WorkspaceID: "ws-789",
			Type:        "cpu",
			Severity:    "error",
			Title:       "Updated Alert",
			Description: "Updated description",
			Threshold:   85.0,
			Value:       90.0,
			Status:      "resolved",
			ResolvedAt:  &[]time.Time{time.Now()}[0],
		}

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "alert_records" SET`).
			WithArgs(
				alert.WorkspaceID,
				alert.Type,
				alert.Severity,
				alert.Title,
				alert.Description,
				alert.Resource,
				alert.Threshold,
				alert.Value,
				alert.Status,
				sqlmock.AnyArg(), // created_at
				alert.ResolvedAt,
				sqlmock.AnyArg(), // updated_at
				alert.ID,         // WHERE id = ?
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateAlert(ctx, alert)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_DeleteAlert(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("delete alert", func(t *testing.T) {
		alertID := uuid.New().String()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "alert_records" WHERE id = \$1`).
			WithArgs(alertID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.DeleteAlert(ctx, alertID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_SaveMetrics(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("save multiple metrics", func(t *testing.T) {
		metrics := []*monitoring.MetricDataPoint{
			{
				ID:          uuid.New().String(),
				WorkspaceID: "ws-111",
				MetricName:  "cpu_usage",
				Value:       45.5,
				Labels: map[string]string{
					"node": "node-1",
					"pod":  "app-pod-1",
				},
				Timestamp: time.Now(),
			},
			{
				ID:          uuid.New().String(),
				WorkspaceID: "ws-111",
				MetricName:  "memory_usage",
				Value:       60.0,
				Labels: map[string]string{
					"node": "node-1",
					"pod":  "app-pod-1",
				},
				Timestamp: time.Now(),
			},
		}

		mock.ExpectBegin()
		// Expect two inserts
		for _, metric := range metrics {
			mock.ExpectExec(`INSERT INTO "metric_records"`).
				WithArgs(
					metric.ID,
					metric.WorkspaceID,
					metric.MetricName,
					metric.Value,
					sqlmock.AnyArg(), // labels (JSON)
					sqlmock.AnyArg(), // timestamp
					sqlmock.AnyArg(), // created_at
				).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectCommit()

		err := repo.SaveMetrics(ctx, metrics)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetMetrics(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("get metrics in time range", func(t *testing.T) {
		workspaceID := "ws-222"
		metricName := "cpu_usage"
		now := time.Now()
		startTime := now.Add(-1 * time.Hour)

		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "metric_name", "value", "labels", "timestamp", "created_at",
		}).
			AddRow(uuid.New().String(), workspaceID, metricName, 45.5, `{"node":"node-1"}`, now.Add(-30*time.Minute), now).
			AddRow(uuid.New().String(), workspaceID, metricName, 50.0, `{"node":"node-1"}`, now.Add(-15*time.Minute), now)

		mock.ExpectQuery(`SELECT \* FROM "metric_records" WHERE workspace_id = \$1 AND metric_name = \$2 AND timestamp >= \$3 AND timestamp <= \$4 ORDER BY timestamp ASC`).
			WithArgs(workspaceID, metricName, startTime, now).
			WillReturnRows(rows)

		metrics, err := repo.GetMetrics(ctx, workspaceID, metricName, startTime, now)
		assert.NoError(t, err)
		assert.Len(t, metrics, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_SaveHealthCheck(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("save cluster health", func(t *testing.T) {
		health := &monitoring.ClusterHealth{
			WorkspaceID: "ws-333",
			Healthy:     true,
			Components: map[string]monitoring.ComponentHealth{
				"api-server": {
					Name:   "api-server",
					Status: "healthy",
				},
				"etcd": {
					Name:   "etcd",
					Status: "healthy",
				},
			},
			LastChecked: time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "health_check_records"`).
			WithArgs(
				sqlmock.AnyArg(), // id
				health.WorkspaceID,
				health.Healthy,
				sqlmock.AnyArg(), // components (JSON)
				sqlmock.AnyArg(), // timestamp
				sqlmock.AnyArg(), // created_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.SaveHealthCheck(ctx, health)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetLatestHealthCheck(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("get latest health check", func(t *testing.T) {
		workspaceID := "ws-444"
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "healthy", "components", "timestamp", "created_at",
		}).AddRow(
			uuid.New().String(), workspaceID, true, 
			`{"api-server":{"name":"api-server","status":"healthy"},"etcd":{"name":"etcd","status":"healthy"}}`,
			now, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "health_check_records" WHERE workspace_id = \$1 ORDER BY timestamp DESC LIMIT \$2`).
			WithArgs(workspaceID, 1).
			WillReturnRows(rows)

		health, err := repo.GetLatestHealthCheck(ctx, workspaceID)
		assert.NoError(t, err)
		assert.NotNil(t, health)
		assert.True(t, health.Healthy)
		assert.Len(t, health.Components, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_DeleteOldMetrics(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("cleanup old metrics", func(t *testing.T) {
		before := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "metric_records" WHERE timestamp < \$1`).
			WithArgs(before).
			WillReturnResult(sqlmock.NewResult(0, 1000))
		mock.ExpectCommit()

		err := repo.DeleteOldMetrics(ctx, before)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}