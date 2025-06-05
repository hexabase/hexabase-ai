package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-kaas/api/internal/domain/billing"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL billing repository
func NewPostgresRepository(db *gorm.DB) billing.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateSubscription(ctx context.Context, subscription *billing.Subscription) error {
	if err := r.db.WithContext(ctx).Create(subscription).Error; err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetSubscription(ctx context.Context, subscriptionID string) (*billing.Subscription, error) {
	var subscription billing.Subscription
	if err := r.db.WithContext(ctx).Where("id = ?", subscriptionID).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &subscription, nil
}

func (r *postgresRepository) GetOrganizationSubscription(ctx context.Context, orgID string) (*billing.Subscription, error) {
	var subscription billing.Subscription
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status IN ?", orgID, []string{"active", "trialing", "cancel_at_period_end"}).
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active subscription not found")
		}
		return nil, fmt.Errorf("failed to get organization subscription: %w", err)
	}
	return &subscription, nil
}

func (r *postgresRepository) UpdateSubscription(ctx context.Context, subscription *billing.Subscription) error {
	if err := r.db.WithContext(ctx).Save(subscription).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListSubscriptions(ctx context.Context, filter billing.SubscriptionFilter) ([]*billing.Subscription, error) {
	var subscriptions []*billing.Subscription

	query := r.db.WithContext(ctx).Model(&billing.Subscription{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.PlanID != "" {
		query = query.Where("plan_id = ?", filter.PlanID)
	}

	// Apply pagination
	if filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	if err := query.Order("created_at DESC").Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (r *postgresRepository) GetPlan(ctx context.Context, planID string) (*billing.Plan, error) {
	var plan billing.Plan
	if err := r.db.WithContext(ctx).Where("id = ?", planID).First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plan not found")
		}
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	return &plan, nil
}

func (r *postgresRepository) ListPlans(ctx context.Context, activeOnly bool) ([]*billing.Plan, error) {
	var plans []*billing.Plan

	query := r.db.WithContext(ctx).Model(&billing.Plan{})

	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("price ASC").Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	return plans, nil
}

func (r *postgresRepository) CreatePlan(ctx context.Context, plan *billing.Plan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdatePlan(ctx context.Context, plan *billing.Plan) error {
	if err := r.db.WithContext(ctx).Save(plan).Error; err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}
	return nil
}

func (r *postgresRepository) CreatePaymentMethod(ctx context.Context, method *billing.PaymentMethod) error {
	if err := r.db.WithContext(ctx).Create(method).Error; err != nil {
		return fmt.Errorf("failed to create payment method: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetPaymentMethod(ctx context.Context, methodID string) (*billing.PaymentMethod, error) {
	var method billing.PaymentMethod
	if err := r.db.WithContext(ctx).Where("id = ?", methodID).First(&method).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment method not found")
		}
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	return &method, nil
}

func (r *postgresRepository) ListPaymentMethods(ctx context.Context, orgID string) ([]*billing.PaymentMethod, error) {
	var methods []*billing.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("is_default DESC, created_at DESC").
		Find(&methods).Error; err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}
	return methods, nil
}

func (r *postgresRepository) UpdatePaymentMethod(ctx context.Context, method *billing.PaymentMethod) error {
	if err := r.db.WithContext(ctx).Save(method).Error; err != nil {
		return fmt.Errorf("failed to update payment method: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeletePaymentMethod(ctx context.Context, methodID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", methodID).Delete(&billing.PaymentMethod{}).Error; err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}
	return nil
}

func (r *postgresRepository) SetDefaultPaymentMethod(ctx context.Context, orgID, methodID string) error {
	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()

	// Unset current default
	if err := tx.Model(&billing.PaymentMethod{}).
		Where("organization_id = ? AND is_default = ?", orgID, true).
		Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unset default payment method: %w", err)
	}

	// Set new default
	if err := tx.Model(&billing.PaymentMethod{}).
		Where("id = ?", methodID).
		Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to set default payment method: %w", err)
	}

	return tx.Commit().Error
}

func (r *postgresRepository) CreateInvoice(ctx context.Context, invoice *billing.Invoice) error {
	if err := r.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetInvoice(ctx context.Context, invoiceID string) (*billing.Invoice, error) {
	var invoice billing.Invoice
	if err := r.db.WithContext(ctx).Where("id = ?", invoiceID).First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

func (r *postgresRepository) ListInvoices(ctx context.Context, filter billing.InvoiceFilter) ([]*billing.Invoice, int, error) {
	var invoices []*billing.Invoice
	var total int64

	query := r.db.WithContext(ctx).Model(&billing.Invoice{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Apply pagination
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	if err := query.Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}

	return invoices, int(total), nil
}

func (r *postgresRepository) UpdateInvoice(ctx context.Context, invoice *billing.Invoice) error {
	if err := r.db.WithContext(ctx).Save(invoice).Error; err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetInvoiceLineItems(ctx context.Context, invoiceID string) ([]*billing.LineItem, error) {
	var items []*billing.LineItem
	if err := r.db.WithContext(ctx).
		Where("invoice_id = ?", invoiceID).
		Order("created_at ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoice line items: %w", err)
	}
	return items, nil
}

func (r *postgresRepository) CreateUsageRecord(ctx context.Context, record *billing.UsageRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to create usage record: %w", err)
	}
	return nil
}

func (r *postgresRepository) BatchCreateUsageRecords(ctx context.Context, records []*billing.UsageRecord) error {
	if err := r.db.WithContext(ctx).CreateInBatches(records, 100).Error; err != nil {
		return fmt.Errorf("failed to batch create usage records: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetUsageRecords(ctx context.Context, filter billing.UsageFilter) ([]*billing.UsageRecord, error) {
	var records []*billing.UsageRecord

	query := r.db.WithContext(ctx).Model(&billing.UsageRecord{})

	if filter.ResourceType != "" {
		query = query.Where("resource_type = ?", filter.ResourceType)
	}

	if filter.WorkspaceID != "" {
		query = query.Where("workspace_id = ?", filter.WorkspaceID)
	}

	query = query.Where("recorded_at >= ? AND recorded_at <= ?", filter.StartDate, filter.EndDate)

	if err := query.Order("recorded_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	return records, nil
}

func (r *postgresRepository) SummarizeUsage(ctx context.Context, orgID string, start, end time.Time) (map[string]float64, error) {
	type UsageSummary struct {
		ResourceType string
		Total        float64
	}

	var summaries []UsageSummary

	if err := r.db.WithContext(ctx).
		Table("usage_records").
		Select("resource_type, SUM(quantity) as total").
		Where("organization_id = ? AND recorded_at >= ? AND recorded_at <= ?", orgID, start, end).
		Group("resource_type").
		Scan(&summaries).Error; err != nil {
		return nil, fmt.Errorf("failed to summarize usage: %w", err)
	}

	usage := make(map[string]float64)
	for _, summary := range summaries {
		usage[summary.ResourceType] = summary.Total
	}

	return usage, nil
}

func (r *postgresRepository) GetOrganization(ctx context.Context, orgID string) (*billing.Organization, error) {
	var org billing.Organization
	if err := r.db.WithContext(ctx).
		Table("organizations").
		Where("id = ?", orgID).
		First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return &org, nil
}

func (r *postgresRepository) GetBillingSettings(ctx context.Context, orgID string) (*billing.BillingSettings, error) {
	var settings billing.BillingSettings
	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default settings
			return &billing.BillingSettings{
				OrganizationID:       orgID,
				BillingEmail:         "",
				InvoicePrefix:        "",
				TaxExempt:            false,
				TaxID:                "",
				PurchaseOrderNumber:  "",
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get billing settings: %w", err)
	}
	return &settings, nil
}

func (r *postgresRepository) UpdateBillingSettings(ctx context.Context, settings *billing.BillingSettings) error {
	// Check if settings exist
	var existing billing.BillingSettings
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", settings.OrganizationID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new settings
		if err := r.db.WithContext(ctx).Create(settings).Error; err != nil {
			return fmt.Errorf("failed to create billing settings: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing settings: %w", err)
	} else {
		// Update existing settings
		if err := r.db.WithContext(ctx).Save(settings).Error; err != nil {
			return fmt.Errorf("failed to update billing settings: %w", err)
		}
	}

	return nil
}