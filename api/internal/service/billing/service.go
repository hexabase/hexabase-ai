package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/billing"
	"go.uber.org/zap"
)

type service struct {
	repo       billing.Repository
	stripeRepo billing.StripeRepository
	logger     *zap.Logger
}

// NewService creates a new billing service
func NewService(
	repo billing.Repository,
	stripeRepo billing.StripeRepository,
	logger *zap.Logger,
) billing.Service {
	return &service{
		repo:       repo,
		stripeRepo: stripeRepo,
		logger:     logger,
	}
}

func (s *service) CreateSubscription(ctx context.Context, orgID string, req *billing.CreateSubscriptionRequest) (*billing.Subscription, error) {
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
	sub := &billing.Subscription{
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

func (s *service) GetSubscription(ctx context.Context, subscriptionID string) (*billing.Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Get plan details
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		s.logger.Warn("failed to get plan details", zap.Error(err))
	} else {
		sub.Plan = plan
	}

	return sub, nil
}

func (s *service) GetOrganizationSubscription(ctx context.Context, orgID string) (*billing.Subscription, error) {
	sub, err := s.repo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Get plan details
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		s.logger.Warn("failed to get plan details", zap.Error(err))
	} else {
		sub.Plan = plan
	}

	return sub, nil
}

func (s *service) UpdateSubscription(ctx context.Context, subscriptionID string, req *billing.UpdateSubscriptionRequest) (*billing.Subscription, error) {
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

func (s *service) CancelSubscription(ctx context.Context, subscriptionID string, req *billing.CancelSubscriptionRequest) error {
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

func (s *service) ReactivateSubscription(ctx context.Context, subscriptionID string) (*billing.Subscription, error) {
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

func (s *service) GetPlan(ctx context.Context, planID string) (*billing.Plan, error) {
	return s.repo.GetPlan(ctx, planID)
}

func (s *service) ListPlans(ctx context.Context) ([]*billing.Plan, error) {
	return s.repo.ListPlans(ctx, true)
}

func (s *service) ComparePlans(ctx context.Context, currentPlanID, targetPlanID string) (*billing.PlanComparison, error) {
	currentPlan, err := s.repo.GetPlan(ctx, currentPlanID)
	if err != nil {
		return nil, fmt.Errorf("current plan not found: %w", err)
	}

	targetPlan, err := s.repo.GetPlan(ctx, targetPlanID)
	if err != nil {
		return nil, fmt.Errorf("target plan not found: %w", err)
	}

	comparison := &billing.PlanComparison{
		CurrentPlan: currentPlan,
		TargetPlan:  targetPlan,
		Changes:     make(map[string]billing.ComparisonItem),
		IsUpgrade:   targetPlan.Price > currentPlan.Price,
		PriceDiff:   targetPlan.Price - currentPlan.Price,
	}

	// Compare features
	for feature, currentLimit := range currentPlan.Features {
		targetLimit, exists := targetPlan.Features[feature]
		if !exists {
			comparison.Changes[feature] = billing.ComparisonItem{
				Feature: feature,
				Current: currentLimit,
				Target:  nil,
				Change:  "removed",
			}
		} else {
			change := "same"
			if targetLimit.(float64) > currentLimit.(float64) {
				change = "increase"
			} else if targetLimit.(float64) < currentLimit.(float64) {
				change = "decrease"
			}

			comparison.Changes[feature] = billing.ComparisonItem{
				Feature: feature,
				Current: currentLimit,
				Target:  targetLimit,
				Change:  change,
			}
		}
	}

	// Check for new features
	for feature, targetLimit := range targetPlan.Features {
		if _, exists := currentPlan.Features[feature]; !exists {
			comparison.Changes[feature] = billing.ComparisonItem{
				Feature: feature,
				Current: nil,
				Target:  targetLimit,
				Change:  "new",
			}
		}
	}

	return comparison, nil
}

func (s *service) AddPaymentMethod(ctx context.Context, orgID string, req *billing.AddPaymentMethodRequest) (*billing.PaymentMethod, error) {
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
	method := &billing.PaymentMethod{
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
			s.logger.Error("failed to set default payment method", zap.Error(err))
		}
	}

	return method, nil
}

func (s *service) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*billing.PaymentMethod, error) {
	return s.repo.GetPaymentMethod(ctx, paymentMethodID)
}

func (s *service) ListPaymentMethods(ctx context.Context, orgID string) ([]*billing.PaymentMethod, error) {
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

func (s *service) GetInvoice(ctx context.Context, invoiceID string) (*billing.Invoice, error) {
	invoice, err := s.repo.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Get line items
	lineItems, err := s.repo.GetInvoiceLineItems(ctx, invoiceID)
	if err != nil {
		s.logger.Warn("failed to get invoice line items", zap.Error(err))
	} else {
		invoice.LineItems = lineItems
	}

	return invoice, nil
}

func (s *service) ListInvoices(ctx context.Context, orgID string, filter billing.InvoiceFilter) ([]*billing.Invoice, int, error) {
	return s.repo.ListInvoices(ctx, filter)
}

func (s *service) GetUpcomingInvoice(ctx context.Context, orgID string) (*billing.Invoice, error) {
	sub, err := s.repo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("no active subscription found: %w", err)
	}

	// Calculate upcoming invoice
	upcoming := &billing.Invoice{
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
		s.logger.Warn("failed to calculate usage", zap.Error(err))
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

func (s *service) RecordUsage(ctx context.Context, usage *billing.UsageRecord) error {
	usage.ID = uuid.New().String()
	usage.RecordedAt = time.Now()

	if err := s.repo.CreateUsageRecord(ctx, usage); err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	return nil
}

func (s *service) GetCurrentUsage(ctx context.Context, orgID string) (*billing.CurrentUsage, error) {
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

	currentUsage := &billing.CurrentUsage{
		OrganizationID: orgID,
		PeriodStart:    sub.CurrentPeriodStart,
		PeriodEnd:      sub.CurrentPeriodEnd,
		Usage:          make(map[string]*billing.ResourceUsage),
	}

	// Calculate usage percentages
	for resource, used := range usage {
		limit, ok := plan.Features[resource].(float64)
		if !ok {
			continue
		}

		percentage := (used / limit) * 100
		if percentage > 100 {
			percentage = 100
		}

		currentUsage.Usage[resource] = &billing.ResourceUsage{
			Used:       used,
			Limit:      limit,
			Percentage: percentage,
		}
	}

	return currentUsage, nil
}

func (s *service) GetUsageHistory(ctx context.Context, orgID string, filter billing.UsageFilter) ([]*billing.UsageRecord, error) {
	return s.repo.GetUsageRecords(ctx, filter)
}

func (s *service) CalculateOverage(ctx context.Context, orgID string) (*billing.OverageReport, error) {
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

	report := &billing.OverageReport{
		OrganizationID: orgID,
		PeriodStart:    sub.CurrentPeriodStart,
		PeriodEnd:      sub.CurrentPeriodEnd,
		Overages:       make(map[string]*billing.OverageDetail),
		TotalCost:      0,
	}

	// Calculate overages
	for resource, used := range usage {
		limit, ok := plan.Features[resource].(float64)
		if !ok {
			continue
		}

		if used > limit {
			overage := used - limit
			rate := plan.UsageRates[resource]
			cost := overage * rate

			report.Overages[resource] = &billing.OverageDetail{
				ResourceType: resource,
				Used:         used,
				Limit:        limit,
				Overage:      overage,
				Rate:         rate,
				Cost:         cost,
			}

			report.TotalCost += cost
		}
	}

	return report, nil
}

func (s *service) GetBillingOverview(ctx context.Context, orgID string) (*billing.BillingOverview, error) {
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
		s.logger.Warn("failed to get upcoming invoice", zap.Error(err))
	}

	// Get recent invoices
	invoices, _, err := s.ListInvoices(ctx, orgID, billing.InvoiceFilter{
		Status:   "paid",
		PageSize: 5,
	})
	if err != nil {
		s.logger.Warn("failed to get recent invoices", zap.Error(err))
	}

	// Get payment methods
	paymentMethods, err := s.ListPaymentMethods(ctx, orgID)
	if err != nil {
		s.logger.Warn("failed to get payment methods", zap.Error(err))
	}

	overview := &billing.BillingOverview{
		Subscription:    sub,
		CurrentUsage:    currentUsage,
		UpcomingInvoice: upcomingInvoice,
		RecentInvoices:  invoices,
		PaymentMethods:  paymentMethods,
	}

	return overview, nil
}

func (s *service) GetBillingSettings(ctx context.Context, orgID string) (*billing.BillingSettings, error) {
	return s.repo.GetBillingSettings(ctx, orgID)
}

func (s *service) UpdateBillingSettings(ctx context.Context, orgID string, settings *billing.BillingSettings) error {
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
		s.logger.Info("subscription updated", zap.String("event_id", event.ID))

	case "customer.subscription.deleted":
		// Handle subscription cancellation
		s.logger.Info("subscription deleted", zap.String("event_id", event.ID))

	default:
		s.logger.Info("unhandled webhook event", zap.String("type", event.Type))
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
	if newPlan.Price < currentPlan.Price {
		// Check current usage against new plan limits
		usage, err := s.repo.SummarizeUsage(ctx, orgID, sub.CurrentPeriodStart, time.Now())
		if err != nil {
			return fmt.Errorf("failed to check usage: %w", err)
		}

		for resource, used := range usage {
			limit, ok := newPlan.Features[resource].(float64)
			if ok && used > limit {
				return fmt.Errorf("current %s usage (%.2f) exceeds new plan limit (%.2f)", resource, used, limit)
			}
		}
	}

	return nil
}

func (s *service) CheckUsageLimits(ctx context.Context, orgID string) (*billing.LimitCheckResult, error) {
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

	result := &billing.LimitCheckResult{
		WithinLimits: true,
		Violations:   []*billing.LimitViolation{},
		Usage:        usage,
		Limits:       make(map[string]float64),
	}

	// Check each resource
	for resource, limit := range plan.Features {
		limitValue, ok := limit.(float64)
		if !ok {
			continue
		}

		result.Limits[resource] = limitValue

		used := usage[resource]
		percentage := (used / limitValue) * 100

		if used > limitValue {
			result.WithinLimits = false
			result.Violations = append(result.Violations, &billing.LimitViolation{
				ResourceType: resource,
				Current:      used,
				Limit:        limitValue,
				Percentage:   percentage,
				Message:      fmt.Sprintf("%s usage exceeds limit", resource),
			})
		} else if percentage > 80 {
			// Warning for high usage
			result.Violations = append(result.Violations, &billing.LimitViolation{
				ResourceType: resource,
				Current:      used,
				Limit:        limitValue,
				Percentage:   percentage,
				Message:      fmt.Sprintf("%s usage is at %.0f%% of limit", resource, percentage),
			})
		}
	}

	return result, nil
}