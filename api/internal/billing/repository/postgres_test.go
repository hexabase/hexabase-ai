package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/billing/domain"
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

func TestPostgresRepository_CreateSubscription(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("successful subscription creation", func(t *testing.T) {
		sub := &domain.Subscription{
			ID:             uuid.New().String(),
			OrganizationID: "org-123",
			PlanID:         "plan-premium",
			Status:         "active",
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "subscriptions"`).
			WithArgs(
				sub.ID,
				sub.OrganizationID,
				sub.PlanID,
				sub.Status,
				sub.CurrentPeriodStart,
				sub.CurrentPeriodEnd,
				sub.StripeSubscriptionID,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateSubscription(ctx, sub)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetSubscription(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("get subscription by ID", func(t *testing.T) {
		subID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "organization_id", "plan_id", "status", "start_date", "end_date",
			"current_period_start", "current_period_end", "stripe_subscription_id",
			"created_at", "updated_at",
		}).AddRow(
			subID, "org-123", "plan-premium", "active", now, nil,
			now, now.AddDate(0, 1, 0), "sub_stripe_123",
			now, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE id = \$1`).
			WithArgs(subID).
			WillReturnRows(rows)

		sub, err := repo.GetSubscription(ctx, subID)
		assert.NoError(t, err)
		assert.NotNil(t, sub)
		assert.Equal(t, subID, sub.ID)
		assert.Equal(t, "active", sub.Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetOrganizationSubscription(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("get active subscription for organization", func(t *testing.T) {
		orgID := "org-456"
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "organization_id", "plan_id", "status", "start_date", "end_date",
			"current_period_start", "current_period_end", "stripe_subscription_id",
			"created_at", "updated_at",
		}).AddRow(
			uuid.New().String(), orgID, "plan-basic", "active", now, nil,
			now, now.AddDate(0, 1, 0), nil,
			now, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE organization_id = \$1 AND status = \$2`).
			WithArgs(orgID, "active").
			WillReturnRows(rows)

		sub, err := repo.GetOrganizationSubscription(ctx, orgID)
		assert.NoError(t, err)
		assert.NotNil(t, sub)
		assert.Equal(t, orgID, sub.OrganizationID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_UpdateSubscription(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("update subscription status", func(t *testing.T) {
		sub := &domain.Subscription{
			ID:             uuid.New().String(),
			OrganizationID: "org-789",
			PlanID:         "plan-enterprise",
			Status:         "canceled",
			CanceledAt:     timePtr(time.Now()),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "subscriptions" SET`).
			WithArgs(
				sub.OrganizationID,
				sub.PlanID,
				sub.Status,
				sub.CanceledAt,
				sqlmock.AnyArg(), // current_period_start
				sqlmock.AnyArg(), // current_period_end
				sqlmock.AnyArg(), // stripe_subscription_id
				sqlmock.AnyArg(), // updated_at
				sub.ID,           // WHERE id = ?
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateSubscription(ctx, sub)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_CreateInvoice(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("create invoice", func(t *testing.T) {
		invoice := &domain.Invoice{
			ID:               uuid.New().String(),
			SubscriptionID:   "sub-123",
			OrganizationID:   "org-123",
			Amount:           9900, // $99.00
			Currency:         "USD",
			Status:           "pending",
			PeriodStart:      time.Now(),
			PeriodEnd:        time.Now().AddDate(0, 1, 0),
			DueDate:          time.Now().AddDate(0, 0, 7),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "invoices"`).
			WithArgs(
				invoice.ID,
				invoice.SubscriptionID,
				invoice.OrganizationID,
				invoice.Amount,
				invoice.Currency,
				invoice.Status,
				invoice.PeriodStart,
				invoice.PeriodEnd,
				invoice.DueDate,
				invoice.PaidAt,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateInvoice(ctx, invoice)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetInvoice(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("get invoice by ID", func(t *testing.T) {
		invoiceID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "subscription_id", "organization_id", "amount", "currency", "status",
			"period_start", "period_end", "due_date", "paid_at", "stripe_invoice_id",
			"created_at", "updated_at",
		}).AddRow(
			invoiceID, "sub-123", "org-123", 9900, "USD", "paid",
			now, now.AddDate(0, 1, 0), now.AddDate(0, 0, 7), &now, "inv_stripe_123",
			now, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE id = \$1`).
			WithArgs(invoiceID).
			WillReturnRows(rows)

		invoice, err := repo.GetInvoice(ctx, invoiceID)
		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, invoiceID, invoice.ID)
		assert.Equal(t, "paid", invoice.Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_ListOrganizationInvoices(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("list invoices for organization", func(t *testing.T) {
		orgID := "org-999"
		limit := 20
		now := time.Now()

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery(`SELECT count\(\*\) FROM "invoices" WHERE organization_id = \$1`).
			WithArgs(orgID).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{
			"id", "subscription_id", "organization_id", "amount", "currency", "status",
			"period_start", "period_end", "due_date", "paid_at", "stripe_invoice_id",
			"created_at", "updated_at",
		}).
			AddRow(uuid.New().String(), "sub-1", orgID, 9900, "USD", "paid", now, now.AddDate(0, 1, 0), now.AddDate(0, 0, 7), &now, nil, now, now).
			AddRow(uuid.New().String(), "sub-1", orgID, 9900, "USD", "pending", now.AddDate(0, -1, 0), now, now, nil, nil, now, now)

		mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE organization_id = \$1 ORDER BY created_at DESC LIMIT \$2`).
			WithArgs(orgID, limit).
			WillReturnRows(rows)

		filter := domain.InvoiceFilter{
			PageSize: limit,
		}
		invoices, total, err := repo.ListInvoices(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, invoices, 2)
		assert.Equal(t, 5, total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_CreateUsageRecord(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("create usage record", func(t *testing.T) {
		usage := &domain.UsageRecord{
			ID:             uuid.New().String(),
			OrganizationID: "org-555",
			ResourceType:   "cpu_cores",
			Quantity:       1000, // 1000 CPU hours
			Unit:           "cores",
			Timestamp:      time.Now(),
			WorkspaceID:    "ws-123",
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "usage_records"`).
			WithArgs(
				usage.ID,
				usage.OrganizationID,
				usage.ResourceType,
				usage.Quantity,
				usage.Unit,
				usage.Timestamp,
				usage.WorkspaceID,
				sqlmock.AnyArg(), // created_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateUsageRecord(ctx, usage)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetOrganizationUsage(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("get usage summary", func(t *testing.T) {
		orgID := "org-777"
		startTime := time.Now().AddDate(0, -1, 0)
		endTime := time.Now()

		rows := sqlmock.NewRows([]string{
			"resource_type", "total_quantity",
		}).
			AddRow("cpu_cores", 5000.0).
			AddRow("memory_gb", 10000.0).
			AddRow("storage_gb", 500.0)

		mock.ExpectQuery(`SELECT resource_type, SUM\(quantity\) as total_quantity FROM "usage_records"`).
			WithArgs(orgID, startTime, endTime).
			WillReturnRows(rows)

		usage, err := repo.SummarizeUsage(ctx, orgID, startTime, endTime)
		assert.NoError(t, err)
		assert.Len(t, usage, 3)
		assert.Equal(t, float64(5000), usage["cpu_cores"])
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}