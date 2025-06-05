package billing

import (
	"context"
	"time"
)

// Repository defines the data access interface for billing
type Repository interface {
	// Subscription operations
	CreateSubscription(ctx context.Context, subscription *Subscription) error
	GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error)
	GetOrganizationSubscription(ctx context.Context, orgID string) (*Subscription, error)
	UpdateSubscription(ctx context.Context, subscription *Subscription) error
	ListSubscriptions(ctx context.Context, filter SubscriptionFilter) ([]*Subscription, error)
	
	// Plan operations
	GetPlan(ctx context.Context, planID string) (*Plan, error)
	ListPlans(ctx context.Context, activeOnly bool) ([]*Plan, error)
	CreatePlan(ctx context.Context, plan *Plan) error
	UpdatePlan(ctx context.Context, plan *Plan) error
	
	// Payment method operations
	CreatePaymentMethod(ctx context.Context, method *PaymentMethod) error
	GetPaymentMethod(ctx context.Context, methodID string) (*PaymentMethod, error)
	ListPaymentMethods(ctx context.Context, orgID string) ([]*PaymentMethod, error)
	UpdatePaymentMethod(ctx context.Context, method *PaymentMethod) error
	DeletePaymentMethod(ctx context.Context, methodID string) error
	SetDefaultPaymentMethod(ctx context.Context, orgID, methodID string) error
	
	// Invoice operations
	CreateInvoice(ctx context.Context, invoice *Invoice) error
	GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error)
	ListInvoices(ctx context.Context, filter InvoiceFilter) ([]*Invoice, int, error)
	UpdateInvoice(ctx context.Context, invoice *Invoice) error
	GetInvoiceLineItems(ctx context.Context, invoiceID string) ([]*LineItem, error)
	
	// Usage operations
	CreateUsageRecord(ctx context.Context, record *UsageRecord) error
	BatchCreateUsageRecords(ctx context.Context, records []*UsageRecord) error
	GetUsageRecords(ctx context.Context, filter UsageFilter) ([]*UsageRecord, error)
	SummarizeUsage(ctx context.Context, orgID string, start, end time.Time) (map[string]float64, error)
	
	// Organization operations
	GetOrganization(ctx context.Context, orgID string) (*Organization, error)
	GetBillingSettings(ctx context.Context, orgID string) (*BillingSettings, error)
	UpdateBillingSettings(ctx context.Context, settings *BillingSettings) error
}

// StripeRepository defines the interface for Stripe payment processing
type StripeRepository interface {
	// Customer operations
	CreateCustomer(ctx context.Context, org *Organization) (string, error)
	UpdateCustomer(ctx context.Context, customerID string, org *Organization) error
	DeleteCustomer(ctx context.Context, customerID string) error
	
	// Subscription operations
	CreateStripeSubscription(ctx context.Context, customerID, priceID string) (*StripeSubscription, error)
	UpdateStripeSubscription(ctx context.Context, subscriptionID string, params map[string]interface{}) error
	CancelStripeSubscription(ctx context.Context, subscriptionID string, immediate bool) error
	
	// Payment method operations
	AttachPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) error
	SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error
	
	// Invoice operations
	CreateInvoice(ctx context.Context, customerID string) (*StripeInvoice, error)
	FinalizeInvoice(ctx context.Context, invoiceID string) error
	PayInvoice(ctx context.Context, invoiceID string) error
	VoidInvoice(ctx context.Context, invoiceID string) error
	
	// Webhook processing
	ConstructWebhookEvent(payload []byte, signature string) (*StripeEvent, error)
}

// SubscriptionFilter represents filter options for subscriptions
type SubscriptionFilter struct {
	Status string
	PlanID string
	Page   int
	Limit  int
}

// StripeSubscription represents a Stripe subscription
type StripeSubscription struct {
	ID                 string
	CustomerID         string
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAt           *time.Time
	Items              []StripeSubscriptionItem
}

// StripeSubscriptionItem represents a Stripe subscription item
type StripeSubscriptionItem struct {
	ID       string
	PriceID  string
	Quantity int64
}

// StripeInvoice represents a Stripe invoice
type StripeInvoice struct {
	ID            string
	Number        string
	CustomerID    string
	Amount        int64
	Currency      string
	Status        string
	PeriodStart   time.Time
	PeriodEnd     time.Time
	DueDate       time.Time
	Lines         []StripeLineItem
}

// StripeLineItem represents a Stripe invoice line item
type StripeLineItem struct {
	Description string
	Amount      int64
	Quantity    int64
}

// StripeEvent represents a Stripe webhook event
type StripeEvent struct {
	ID   string
	Type string
	Data map[string]interface{}
}