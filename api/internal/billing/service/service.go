package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/billing/domain"
)

type service struct {
	repo       domain.Repository
	stripeRepo domain.StripeRepository
	logger     *slog.Logger
}

// NewService creates a new billing service
func NewService(
	repo domain.Repository,
	stripeRepo domain.StripeRepository,
	logger *slog.Logger,
) domain.Service {
	return &service{
		repo:       repo,
		stripeRepo: stripeRepo,
		logger:     logger,
	}
}

func (s *service) CreateSubscription(ctx context.Context, orgID string, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	// Get organization
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Get plan
	plan, err := s.repo.GetPlan(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Create Stripe subscription
	stripesSub, err := s.stripeRepo.CreateStripeSubscription(ctx, org.StripeCustomerID, plan.StripePriceID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe subscription: %w", err)
	}

	// Create subscription record
	sub := &domain.Subscription{
		ID:                uuid.New().String(),
		OrganizationID:    orgID,
		PlanID:            req.PlanID,
		Status:            stripesSub.Status,
		CurrentPeriodStart: stripesSub.CurrentPeriodStart,
		CurrentPeriodEnd:   stripesSub.CurrentPeriodEnd,
		TrialEnd:          req.TrialEnd,
		StripeSubscriptionID: stripesSub.ID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

func (s *service) GetSubscription(ctx context.Context, subscriptionID string) (*domain.Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Get plan details
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		s.logger.Warn("failed to get plan details", "error", err)
	} else {
		sub.Plan = plan
	}

	return sub, nil
}

func (s *service) GetOrganizationSubscription(ctx context.Context, orgID string) (*domain.Subscription, error) {
	sub, err := s.repo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Get plan details
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		s.logger.Warn("failed to get plan details", "error", err)
	} else {
		sub.Plan = plan
	}

	return sub, nil
}

func (s *service) UpdateSubscription(ctx context.Context, subscriptionID string, req *domain.UpdateSubscriptionRequest) (*domain.Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Update Stripe subscription if plan changed
	if req.PlanID != "" && req.PlanID != sub.PlanID {
		plan, err := s.repo.GetPlan(ctx, req.PlanID)
		if err != nil {
			return nil, fmt.Errorf("plan not found: %w", err)
		}

		params := map[string]interface{}{
			"items": []map[string]interface{}{
				{"price": plan.StripePriceID},
			},
		}

		if err := s.stripeRepo.UpdateStripeSubscription(ctx, sub.StripeSubscriptionID, params); err != nil {
			return nil, fmt.Errorf("failed to update Stripe subscription: %w", err)
		}

		sub.PlanID = req.PlanID
	}

	sub.UpdatedAt = time.Now()

	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return sub, nil
}

func (s *service) CancelSubscription(ctx context.Context, subscriptionID string, req *domain.CancelSubscriptionRequest) error {
	sub, err := s.repo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Cancel in Stripe
	immediate := req.Immediate
	if err := s.stripeRepo.CancelStripeSubscription(ctx, sub.StripeSubscriptionID, immediate); err != nil {
		return fmt.Errorf("failed to cancel Stripe subscription: %w", err)
	}

	// Update subscription status
	if immediate {
		sub.Status = "canceled"
		sub.CanceledAt = &[]time.Time{time.Now()}[0]
	} else {
		sub.Status = "cancel_at_period_end"
		sub.CancelAt = &sub.CurrentPeriodEnd
	}

	sub.UpdatedAt = time.Now()

	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (s *service) ReactivateSubscription(ctx context.Context, subscriptionID string) (*domain.Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	if sub.Status != "cancel_at_period_end" {
		return nil, fmt.Errorf("can only reactivate subscriptions scheduled for cancellation")
	}

	// Reactivate in Stripe
	params := map[string]interface{}{
		"cancel_at_period_end": false,
	}

	if err := s.stripeRepo.UpdateStripeSubscription(ctx, sub.StripeSubscriptionID, params); err != nil {
		return nil, fmt.Errorf("failed to reactivate Stripe subscription: %w", err)
	}

	// Update subscription
	sub.Status = "active"
	sub.CancelAt = nil
	sub.UpdatedAt = time.Now()

	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return sub, nil
}

func (s *service) GetPlan(ctx context.Context, planID string) (*domain.Plan, error) {
	return s.repo.GetPlan(ctx, planID)
}

func (s *service) ListPlans(ctx context.Context) ([]*domain.Plan, error) {
	return s.repo.ListPlans(ctx, true)
}

func (s *service) ComparePlans(ctx context.Context, currentPlanID, targetPlanID string) (*domain.PlanComparison, error) {
	currentPlan, err := s.repo.GetPlan(ctx, currentPlanID)
	if err != nil {
		return nil, fmt.Errorf("current plan not found: %w", err)
	}

	targetPlan, err := s.repo.GetPlan(ctx, targetPlanID)
	if err != nil {
		return nil, fmt.Errorf("target plan not found: %w", err)
	}

	comparison := &domain.PlanComparison{
		CurrentPlan: currentPlan,
		TargetPlan:  targetPlan,
		Changes:     make(map[string]domain.ComparisonItem),
		IsUpgrade:   targetPlan.Price > currentPlan.Price,
		PriceDiff:   targetPlan.Price - currentPlan.Price,
	}

	// Compare limits
	if currentPlan.Limits != nil && targetPlan.Limits != nil {
		// Compare workspaces
		comparison.Changes["workspaces"] = domain.ComparisonItem{
			Feature: "workspaces",
			Current: currentPlan.Limits.Workspaces,
			Target:  targetPlan.Limits.Workspaces,
			Change:  getChangeType(currentPlan.Limits.Workspaces, targetPlan.Limits.Workspaces),
		}

		// Compare projects
		comparison.Changes["projects"] = domain.ComparisonItem{
			Feature: "projects",
			Current: currentPlan.Limits.Projects,
			Target:  targetPlan.Limits.Projects,
			Change:  getChangeType(currentPlan.Limits.Projects, targetPlan.Limits.Projects),
		}

		// Compare users
		comparison.Changes["users"] = domain.ComparisonItem{
			Feature: "users",
			Current: currentPlan.Limits.Users,
			Target:  targetPlan.Limits.Users,
			Change:  getChangeType(currentPlan.Limits.Users, targetPlan.Limits.Users),
		}

		// Compare CPU cores
		comparison.Changes["cpu_cores"] = domain.ComparisonItem{
			Feature: "cpu_cores",
			Current: currentPlan.Limits.CPUCores,
			Target:  targetPlan.Limits.CPUCores,
			Change:  getChangeType(currentPlan.Limits.CPUCores, targetPlan.Limits.CPUCores),
		}

		// Compare memory
		comparison.Changes["memory_gb"] = domain.ComparisonItem{
			Feature: "memory_gb",
			Current: currentPlan.Limits.MemoryGB,
			Target:  targetPlan.Limits.MemoryGB,
			Change:  getChangeType(currentPlan.Limits.MemoryGB, targetPlan.Limits.MemoryGB),
		}

		// Compare storage
		comparison.Changes["storage_gb"] = domain.ComparisonItem{
			Feature: "storage_gb",
			Current: currentPlan.Limits.StorageGB,
			Target:  targetPlan.Limits.StorageGB,
			Change:  getChangeType(currentPlan.Limits.StorageGB, targetPlan.Limits.StorageGB),
		}

		// Compare bandwidth
		comparison.Changes["bandwidth_gb"] = domain.ComparisonItem{
			Feature: "bandwidth_gb",
			Current: currentPlan.Limits.BandwidthGB,
			Target:  targetPlan.Limits.BandwidthGB,
			Change:  getChangeType(currentPlan.Limits.BandwidthGB, targetPlan.Limits.BandwidthGB),
		}

		// Compare support level
		comparison.Changes["support_level"] = domain.ComparisonItem{
			Feature: "support_level",
			Current: currentPlan.Limits.SupportLevel,
			Target:  targetPlan.Limits.SupportLevel,
			Change:  getChangeTypeString(currentPlan.Limits.SupportLevel, targetPlan.Limits.SupportLevel),
		}
	}

	return comparison, nil
}

func (s *service) AddPaymentMethod(ctx context.Context, orgID string, req *domain.AddPaymentMethodRequest) (*domain.PaymentMethod, error) {
	// Get organization
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Attach payment method to Stripe customer
	if err := s.stripeRepo.AttachPaymentMethod(ctx, org.StripeCustomerID, req.PaymentMethodID); err != nil {
		return nil, fmt.Errorf("failed to attach payment method: %w", err)
	}

	// Create payment method record
	method := &domain.PaymentMethod{
		ID:                 uuid.New().String(),
		OrganizationID:     orgID,
		StripePaymentMethodID: req.PaymentMethodID,
		Type:               req.Type,
		Last4:              req.Last4,
		Brand:              req.Brand,
		ExpiryMonth:        req.ExpiryMonth,
		ExpiryYear:         req.ExpiryYear,
		IsDefault:          req.SetAsDefault,
		CreatedAt:          time.Now(),
	}

	if err := s.repo.CreatePaymentMethod(ctx, method); err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	// Set as default if requested
	if req.SetAsDefault {
		if err := s.SetDefaultPaymentMethod(ctx, method.ID); err != nil {
			s.logger.Error("failed to set default payment method", "error", err)
		}
	}

	return method, nil
}

func (s *service) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*domain.PaymentMethod, error) {
	return s.repo.GetPaymentMethod(ctx, paymentMethodID)
}

func (s *service) ListPaymentMethods(ctx context.Context, orgID string) ([]*domain.PaymentMethod, error) {
	return s.repo.ListPaymentMethods(ctx, orgID)
}

func (s *service) SetDefaultPaymentMethod(ctx context.Context, paymentMethodID string) error {
	method, err := s.repo.GetPaymentMethod(ctx, paymentMethodID)
	if err != nil {
		return fmt.Errorf("payment method not found: %w", err)
	}

	// Get organization
	org, err := s.repo.GetOrganization(ctx, method.OrganizationID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	// Set default in Stripe
	if err := s.stripeRepo.SetDefaultPaymentMethod(ctx, org.StripeCustomerID, method.StripePaymentMethodID); err != nil {
		return fmt.Errorf("failed to set default payment method in Stripe: %w", err)
	}

	// Update in database
	if err := s.repo.SetDefaultPaymentMethod(ctx, method.OrganizationID, paymentMethodID); err != nil {
		return fmt.Errorf("failed to set default payment method: %w", err)
	}

	return nil
}

func (s *service) RemovePaymentMethod(ctx context.Context, paymentMethodID string) error {
	method, err := s.repo.GetPaymentMethod(ctx, paymentMethodID)
	if err != nil {
		return fmt.Errorf("payment method not found: %w", err)
	}

	// Check if it's the default method
	if method.IsDefault {
		// Check if there are other methods
		methods, err := s.repo.ListPaymentMethods(ctx, method.OrganizationID)
		if err != nil {
			return fmt.Errorf("failed to list payment methods: %w", err)
		}

		if len(methods) == 1 {
			return fmt.Errorf("cannot remove the only payment method")
		}
	}

	// Detach from Stripe
	if err := s.stripeRepo.DetachPaymentMethod(ctx, method.StripePaymentMethodID); err != nil {
		return fmt.Errorf("failed to detach payment method: %w", err)
	}

	// Delete from database
	if err := s.repo.DeletePaymentMethod(ctx, paymentMethodID); err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	return nil
}

func (s *service) GetInvoice(ctx context.Context, invoiceID string) (*domain.Invoice, error) {
	invoice, err := s.repo.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Get line items
	lineItems, err := s.repo.GetInvoiceLineItems(ctx, invoiceID)
	if err != nil {
		s.logger.Warn("failed to get invoice line items", "error", err)
	} else {
		invoice.LineItems = lineItems
	}

	return invoice, nil
}

func (s *service) ListInvoices(ctx context.Context, orgID string, filter domain.InvoiceFilter) ([]*domain.Invoice, int, error) {
	return s.repo.ListInvoices(ctx, filter)
}

func (s *service) GetUpcomingInvoice(ctx context.Context, orgID string) (*domain.Invoice, error) {
	sub, err := s.repo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("no active subscription found: %w", err)
	}

	// Calculate upcoming invoice
	upcoming := &domain.Invoice{
		ID:             "upcoming",
		OrganizationID: orgID,
		SubscriptionID: sub.ID,
		PeriodStart:    sub.CurrentPeriodEnd,
		PeriodEnd:      sub.CurrentPeriodEnd.AddDate(0, 1, 0), // Assuming monthly
		Status:         "draft",
		DueDate:        sub.CurrentPeriodEnd.AddDate(0, 0, 7), // 7 days after period end
	}

	// Get plan to calculate amount
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	upcoming.Amount = plan.Price
	upcoming.Currency = plan.Currency

	// Calculate usage-based charges
	usage, err := s.repo.SummarizeUsage(ctx, orgID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	if err != nil {
		s.logger.Warn("failed to calculate usage", "error", err)
	} else {
		// Add usage charges to amount
		for resource, amount := range usage {
			if rate, ok := plan.UsageRates[resource]; ok {
				upcoming.Amount += amount * rate
			}
		}
	}

	return upcoming, nil
}

func (s *service) DownloadInvoice(ctx context.Context, invoiceID string) ([]byte, string, error) {
	invoice, err := s.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, "", err
	}

	// Generate PDF (simplified - would use a proper PDF library)
	// For now, return a placeholder
	pdfContent := []byte(fmt.Sprintf("Invoice %s\nAmount: %.2f %s\n", invoice.Number, invoice.Amount, invoice.Currency))
	filename := fmt.Sprintf("invoice-%s.pdf", invoice.Number)

	return pdfContent, filename, nil
}

func (s *service) RecordUsage(ctx context.Context, usage *domain.UsageRecord) error {
	usage.ID = uuid.New().String()
	usage.RecordedAt = time.Now()

	if err := s.repo.CreateUsageRecord(ctx, usage); err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	return nil
}

func (s *service) GetCurrentUsage(ctx context.Context, orgID string) (*domain.CurrentUsage, error) {
	// Get current subscription
	sub, err := s.repo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("no active subscription found: %w", err)
	}

	// Get plan
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	// Get usage for current period
	usage, err := s.repo.SummarizeUsage(ctx, orgID, sub.CurrentPeriodStart, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	currentUsage := &domain.CurrentUsage{
		OrganizationID: orgID,
		PeriodStart:    sub.CurrentPeriodStart,
		PeriodEnd:      sub.CurrentPeriodEnd,
		Usage:          usage,
		ResourceUsage:  make(map[string]domain.Usage),
	}

	// Calculate usage based on plan limits
	if plan.Limits != nil {
		// CPU usage
		if cpuUsed, ok := usage["cpu_cores"]; ok {
			currentUsage.ResourceUsage["cpu_cores"] = domain.Usage{
				Used:  cpuUsed,
				Limit: float64(plan.Limits.CPUCores),
				Unit:  "cores",
			}
		}

		// Memory usage
		if memUsed, ok := usage["memory_gb"]; ok {
			currentUsage.ResourceUsage["memory_gb"] = domain.Usage{
				Used:  memUsed,
				Limit: float64(plan.Limits.MemoryGB),
				Unit:  "GB",
			}
		}

		// Storage usage
		if storageUsed, ok := usage["storage_gb"]; ok {
			currentUsage.ResourceUsage["storage_gb"] = domain.Usage{
				Used:  storageUsed,
				Limit: float64(plan.Limits.StorageGB),
				Unit:  "GB",
			}
		}

		// Workspace usage
		if wsUsed, ok := usage["workspaces"]; ok {
			currentUsage.ResourceUsage["workspaces"] = domain.Usage{
				Used:  wsUsed,
				Limit: float64(plan.Limits.Workspaces),
				Unit:  "count",
			}
		}
	}

	return currentUsage, nil
}

func (s *service) GetUsageHistory(ctx context.Context, orgID string, filter domain.UsageFilter) ([]*domain.UsageRecord, error) {
	return s.repo.GetUsageRecords(ctx, filter)
}

func (s *service) CalculateOverage(ctx context.Context, orgID string) (*domain.OverageReport, error) {
	// Get current subscription
	sub, err := s.repo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("no active subscription found: %w", err)
	}

	// Get plan
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	// Get usage for current period
	usage, err := s.repo.SummarizeUsage(ctx, orgID, sub.CurrentPeriodStart, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	report := &domain.OverageReport{
		OrganizationID: orgID,
		PeriodStart:    sub.CurrentPeriodStart,
		PeriodEnd:      sub.CurrentPeriodEnd,
		Overages:       make(map[string]*domain.OverageDetail),
		TotalCost:      0,
	}

	// Calculate overages based on plan limits
	if plan.Limits != nil && plan.UsageRates != nil {
		// Check CPU overage
		if cpuUsed, ok := usage["cpu_cores"]; ok {
			limit := float64(plan.Limits.CPUCores)
			if cpuUsed > limit {
				overage := cpuUsed - limit
				rate := plan.UsageRates["cpu_cores"]
				cost := overage * rate

				report.Overages["cpu_cores"] = &domain.OverageDetail{
					ResourceType: "cpu_cores",
					Used:         cpuUsed,
					Limit:        limit,
					Overage:      overage,
					Rate:         rate,
					Cost:         cost,
				}
				report.TotalCost += cost
			}
		}

		// Check Memory overage
		if memUsed, ok := usage["memory_gb"]; ok {
			limit := float64(plan.Limits.MemoryGB)
			if memUsed > limit {
				overage := memUsed - limit
				rate := plan.UsageRates["memory_gb"]
				cost := overage * rate

				report.Overages["memory_gb"] = &domain.OverageDetail{
					ResourceType: "memory_gb",
					Used:         memUsed,
					Limit:        limit,
					Overage:      overage,
					Rate:         rate,
					Cost:         cost,
				}
				report.TotalCost += cost
			}
		}

		// Check Storage overage
		if storageUsed, ok := usage["storage_gb"]; ok {
			limit := float64(plan.Limits.StorageGB)
			if storageUsed > limit {
				overage := storageUsed - limit
				rate := plan.UsageRates["storage_gb"]
				cost := overage * rate

				report.Overages["storage_gb"] = &domain.OverageDetail{
					ResourceType: "storage_gb",
					Used:         storageUsed,
					Limit:        limit,
					Overage:      overage,
					Rate:         rate,
					Cost:         cost,
				}
				report.TotalCost += cost
			}
		}
	}

	return report, nil
}

func (s *service) GetBillingOverview(ctx context.Context, orgID string) (*domain.BillingOverview, error) {
	// Get subscription
	sub, err := s.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Get current usage
	currentUsage, err := s.GetCurrentUsage(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Get upcoming invoice
	upcomingInvoice, err := s.GetUpcomingInvoice(ctx, orgID)
	if err != nil {
		s.logger.Warn("failed to get upcoming invoice", "error", err)
	}

	// Get recent invoices
	invoices, _, err := s.ListInvoices(ctx, orgID, domain.InvoiceFilter{
		Status:   "paid",
		PageSize: 5,
	})
	if err != nil {
		s.logger.Warn("failed to get recent invoices", "error", err)
	}

	// Get payment methods
	paymentMethods, err := s.ListPaymentMethods(ctx, orgID)
	if err != nil {
		s.logger.Warn("failed to get payment methods", "error", err)
	}

	overview := &domain.BillingOverview{
		Subscription:    sub,
		CurrentUsage:    currentUsage,
		UpcomingInvoice: upcomingInvoice,
		RecentInvoices:  invoices,
		PaymentMethods:  paymentMethods,
	}

	return overview, nil
}

func (s *service) GetBillingSettings(ctx context.Context, orgID string) (*domain.BillingSettings, error) {
	return s.repo.GetBillingSettings(ctx, orgID)
}

func (s *service) UpdateBillingSettings(ctx context.Context, orgID string, settings *domain.BillingSettings) error {
	settings.OrganizationID = orgID
	settings.UpdatedAt = time.Now()

	if err := s.repo.UpdateBillingSettings(ctx, settings); err != nil {
		return fmt.Errorf("failed to update billing settings: %w", err)
	}

	return nil
}

func (s *service) ProcessStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	event, err := s.stripeRepo.ConstructWebhookEvent(payload, signature)
	if err != nil {
		return fmt.Errorf("failed to verify webhook: %w", err)
	}

	switch event.Type {
	case "invoice.payment_succeeded":
		invoiceID := event.Data["id"].(string)
		return s.ProcessPaymentSuccess(ctx, invoiceID)

	case "invoice.payment_failed":
		invoiceID := event.Data["id"].(string)
		return s.ProcessPaymentFailure(ctx, invoiceID)

	case "customer.subscription.updated":
		// Handle subscription updates
		s.logger.Info("subscription updated", "event_id", event.ID)

	case "customer.subscription.deleted":
		// Handle subscription cancellation
		s.logger.Info("subscription deleted", "event_id", event.ID)

	default:
		s.logger.Info("unhandled webhook event", "type", event.Type)
	}

	return nil
}

func (s *service) ProcessPaymentSuccess(ctx context.Context, invoiceID string) error {
	// Update invoice status
	invoice, err := s.repo.GetInvoice(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	invoice.Status = "paid"
	invoice.PaidAt = &[]time.Time{time.Now()}[0]
	invoice.UpdatedAt = time.Now()

	if err := s.repo.UpdateInvoice(ctx, invoice); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	// TODO: Send payment success notification

	return nil
}

func (s *service) ProcessPaymentFailure(ctx context.Context, invoiceID string) error {
	// Update invoice status
	invoice, err := s.repo.GetInvoice(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	invoice.Status = "payment_failed"
	invoice.UpdatedAt = time.Now()

	if err := s.repo.UpdateInvoice(ctx, invoice); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	// TODO: Send payment failure notification
	// TODO: Check if subscription should be suspended

	return nil
}

func (s *service) ValidatePlanChange(ctx context.Context, orgID, newPlanID string) error {
	// Get current subscription
	sub, err := s.repo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		return fmt.Errorf("no active subscription found: %w", err)
	}

	// Get current and new plans
	currentPlan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("current plan not found: %w", err)
	}

	newPlan, err := s.repo.GetPlan(ctx, newPlanID)
	if err != nil {
		return fmt.Errorf("new plan not found: %w", err)
	}

	// Check if downgrade is allowed
	if newPlan.Price < currentPlan.Price && newPlan.Limits != nil {
		// Check current usage against new plan limits
		usage, err := s.repo.SummarizeUsage(ctx, orgID, sub.CurrentPeriodStart, time.Now())
		if err != nil {
			return fmt.Errorf("failed to check usage: %w", err)
		}

		// Check CPU usage
		if cpuUsed, ok := usage["cpu_cores"]; ok {
			if cpuUsed > float64(newPlan.Limits.CPUCores) {
				return fmt.Errorf("current CPU usage (%.2f cores) exceeds new plan limit (%d cores)", cpuUsed, newPlan.Limits.CPUCores)
			}
		}

		// Check Memory usage
		if memUsed, ok := usage["memory_gb"]; ok {
			if memUsed > float64(newPlan.Limits.MemoryGB) {
				return fmt.Errorf("current memory usage (%.2f GB) exceeds new plan limit (%d GB)", memUsed, newPlan.Limits.MemoryGB)
			}
		}

		// Check Storage usage
		if storageUsed, ok := usage["storage_gb"]; ok {
			if storageUsed > float64(newPlan.Limits.StorageGB) {
				return fmt.Errorf("current storage usage (%.2f GB) exceeds new plan limit (%d GB)", storageUsed, newPlan.Limits.StorageGB)
			}
		}

		// Check Workspace count
		if wsCount, ok := usage["workspaces"]; ok {
			if wsCount > float64(newPlan.Limits.Workspaces) {
				return fmt.Errorf("current workspace count (%.0f) exceeds new plan limit (%d)", wsCount, newPlan.Limits.Workspaces)
			}
		}
	}

	return nil
}

func (s *service) CheckUsageLimits(ctx context.Context, orgID string) (*domain.LimitCheckResult, error) {
	// Get current subscription
	sub, err := s.repo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("no active subscription found: %w", err)
	}

	// Get plan
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	// Get current usage
	usage, err := s.repo.SummarizeUsage(ctx, orgID, sub.CurrentPeriodStart, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	result := &domain.LimitCheckResult{
		WithinLimits: true,
		Violations:   []*domain.LimitViolation{},
		Usage:        usage,
		Limits:       make(map[string]float64),
	}

	// Check each resource limit
	if plan.Limits != nil {
		// Check CPU cores
		checkLimit(result, "cpu_cores", usage["cpu_cores"], float64(plan.Limits.CPUCores))
		
		// Check memory
		checkLimit(result, "memory_gb", usage["memory_gb"], float64(plan.Limits.MemoryGB))
		
		// Check storage
		checkLimit(result, "storage_gb", usage["storage_gb"], float64(plan.Limits.StorageGB))
		
		// Check workspaces
		checkLimit(result, "workspaces", usage["workspaces"], float64(plan.Limits.Workspaces))
		
		// Check projects
		checkLimit(result, "projects", usage["projects"], float64(plan.Limits.Projects))
		
		// Check users
		checkLimit(result, "users", usage["users"], float64(plan.Limits.Users))
		
		// Check bandwidth
		checkLimit(result, "bandwidth_gb", usage["bandwidth_gb"], float64(plan.Limits.BandwidthGB))
	}

	return result, nil
}

// getChangeType compares two integer values and returns the change type
func getChangeType(current, target int) string {
	if target > current {
		return "increase"
	} else if target < current {
		return "decrease"
	}
	return "same"
}

// getChangeTypeString compares two string values and returns the change type
func getChangeTypeString(current, target string) string {
	if current != target {
		return "change"
	}
	return "same"
}

// checkLimit checks if a resource usage exceeds its limit
func checkLimit(result *domain.LimitCheckResult, resource string, used, limit float64) {
	result.Limits[resource] = limit
	
	if limit <= 0 {
		return // No limit set
	}
	
	percentage := (used / limit) * 100
	
	if used > limit {
		result.WithinLimits = false
		result.Violations = append(result.Violations, &domain.LimitViolation{
			ResourceType: resource,
			Current:      used,
			Limit:        limit,
			Percentage:   percentage,
			Message:      fmt.Sprintf("%s usage exceeds limit", resource),
		})
	} else if percentage > 80 {
		// Warning for high usage
		result.Violations = append(result.Violations, &domain.LimitViolation{
			ResourceType: resource,
			Current:      used,
			Limit:        limit,
			Percentage:   percentage,
			Message:      fmt.Sprintf("%s usage is at %.0f%% of limit", resource, percentage),
		})
	}
}