package billing

import (
	"context"
	"time"
)

// Service defines the billing business logic interface
type Service interface {
	// Subscription management
	CreateSubscription(ctx context.Context, orgID string, req *CreateSubscriptionRequest) (*Subscription, error)
	GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error)
	GetOrganizationSubscription(ctx context.Context, orgID string) (*Subscription, error)
	UpdateSubscription(ctx context.Context, subscriptionID string, req *UpdateSubscriptionRequest) (*Subscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string, req *CancelSubscriptionRequest) error
	ReactivateSubscription(ctx context.Context, subscriptionID string) (*Subscription, error)
	
	// Plan management
	GetPlan(ctx context.Context, planID string) (*Plan, error)
	ListPlans(ctx context.Context) ([]*Plan, error)
	ComparePlans(ctx context.Context, currentPlanID, targetPlanID string) (*PlanComparison, error)
	
	// Payment method management
	AddPaymentMethod(ctx context.Context, orgID string, req *AddPaymentMethodRequest) (*PaymentMethod, error)
	GetPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethod, error)
	ListPaymentMethods(ctx context.Context, orgID string) ([]*PaymentMethod, error)
	SetDefaultPaymentMethod(ctx context.Context, paymentMethodID string) error
	RemovePaymentMethod(ctx context.Context, paymentMethodID string) error
	
	// Invoice management
	GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error)
	ListInvoices(ctx context.Context, orgID string, filter InvoiceFilter) ([]*Invoice, int, error)
	GetUpcomingInvoice(ctx context.Context, orgID string) (*Invoice, error)
	DownloadInvoice(ctx context.Context, invoiceID string) ([]byte, string, error)
	
	// Usage tracking
	RecordUsage(ctx context.Context, usage *UsageRecord) error
	GetCurrentUsage(ctx context.Context, orgID string) (*CurrentUsage, error)
	GetUsageHistory(ctx context.Context, orgID string, filter UsageFilter) ([]*UsageRecord, error)
	CalculateOverage(ctx context.Context, orgID string) (*OverageReport, error)
	
	// Billing overview
	GetBillingOverview(ctx context.Context, orgID string) (*BillingOverview, error)
	GetBillingSettings(ctx context.Context, orgID string) (*BillingSettings, error)
	UpdateBillingSettings(ctx context.Context, orgID string, settings *BillingSettings) error
	
	// Webhooks and processing
	ProcessStripeWebhook(ctx context.Context, payload []byte, signature string) error
	ProcessPaymentSuccess(ctx context.Context, invoiceID string) error
	ProcessPaymentFailure(ctx context.Context, invoiceID string) error
	
	// Validation
	ValidatePlanChange(ctx context.Context, orgID, newPlanID string) error
	CheckUsageLimits(ctx context.Context, orgID string) (*LimitCheckResult, error)
}

// PlanComparison represents a comparison between two plans
type PlanComparison struct {
	CurrentPlan *Plan                    `json:"current_plan"`
	TargetPlan  *Plan                    `json:"target_plan"`
	Changes     map[string]ComparisonItem `json:"changes"`
	IsUpgrade   bool                     `json:"is_upgrade"`
	PriceDiff   float64                  `json:"price_difference"`
}

// ComparisonItem represents a single comparison item
type ComparisonItem struct {
	Feature string      `json:"feature"`
	Current interface{} `json:"current"`
	Target  interface{} `json:"target"`
	Change  string      `json:"change"` // increase, decrease, same, new, removed
}

// InvoiceFilter represents filter options for invoices
type InvoiceFilter struct {
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PageSize  int
}

// UsageFilter represents filter options for usage records
type UsageFilter struct {
	ResourceType string
	WorkspaceID  string
	StartDate    time.Time
	EndDate      time.Time
}

// OverageReport represents resource overage information
type OverageReport struct {
	OrganizationID string                     `json:"organization_id"`
	PeriodStart    time.Time                  `json:"period_start"`
	PeriodEnd      time.Time                  `json:"period_end"`
	Overages       map[string]*OverageDetail  `json:"overages"`
	TotalCost      float64                    `json:"total_cost"`
}

// OverageDetail represents overage details for a resource
type OverageDetail struct {
	ResourceType string  `json:"resource_type"`
	Used         float64 `json:"used"`
	Limit        float64 `json:"limit"`
	Overage      float64 `json:"overage"`
	Rate         float64 `json:"rate"`
	Cost         float64 `json:"cost"`
}

// LimitCheckResult represents the result of limit checking
type LimitCheckResult struct {
	WithinLimits bool                      `json:"within_limits"`
	Violations   []*LimitViolation         `json:"violations,omitempty"`
	Usage        map[string]float64        `json:"usage"`
	Limits       map[string]float64        `json:"limits"`
}

// LimitViolation represents a limit violation
type LimitViolation struct {
	ResourceType string  `json:"resource_type"`
	Current      float64 `json:"current"`
	Limit        float64 `json:"limit"`
	Percentage   float64 `json:"percentage"`
	Message      string  `json:"message"`
}