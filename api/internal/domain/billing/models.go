package billing

import (
	"time"
)

// Subscription represents a billing subscription
type Subscription struct {
	ID                   string    `json:"id"`
	OrganizationID       string    `json:"organization_id"`
	PlanID               string    `json:"plan_id"`
	StripeSubscriptionID string    `json:"stripe_subscription_id,omitempty"`
	Status               string    `json:"status"` // active, canceled, past_due, unpaid
	BillingCycle         string    `json:"billing_cycle"` // monthly, yearly
	CurrentPeriodStart   time.Time `json:"current_period_start"`
	CurrentPeriodEnd     time.Time `json:"current_period_end"`
	CancelAt             *time.Time `json:"cancel_at,omitempty"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty"`
	TrialEnd             *time.Time `json:"trial_end,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	Plan                 *Plan     `json:"plan,omitempty"`
}

// Plan represents a subscription plan
type Plan struct {
	ID                      string             `json:"id"`
	Name                    string             `json:"name"`
	Description             string             `json:"description"`
	Price                   float64            `json:"price"` // Current price based on billing cycle
	PriceMonthly            float64            `json:"price_monthly"`
	PriceYearly             float64            `json:"price_yearly"`
	Currency                string             `json:"currency"`
	Features                []string           `json:"features"`
	Limits                  *Limits            `json:"limits"`
	UsageRates              map[string]float64 `json:"usage_rates,omitempty"` // Overage rates per resource
	IsActive                bool               `json:"is_active"`
	IsPopular               bool               `json:"is_popular"`
	YearlyDiscountPercentage float64           `json:"yearly_discount_percentage"`
	StripePriceID           string             `json:"stripe_price_id,omitempty"`
	CreatedAt               time.Time          `json:"created_at"`
	UpdatedAt               time.Time          `json:"updated_at"`
}

// Limits represents plan resource limits
type Limits struct {
	Workspaces   int    `json:"workspaces"`
	Projects     int    `json:"projects"`
	Users        int    `json:"users"`
	CPUCores     int    `json:"cpu_cores"`
	MemoryGB     int    `json:"memory_gb"`
	StorageGB    int    `json:"storage_gb"`
	BandwidthGB  int    `json:"bandwidth_gb"`
	SupportLevel string `json:"support_level"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID                    string    `json:"id"`
	OrganizationID        string    `json:"organization_id"`
	StripePaymentMethodID string    `json:"stripe_payment_method_id,omitempty"`
	Type                  string    `json:"type"` // card, bank_account
	Brand                 string    `json:"brand,omitempty"` // visa, mastercard, etc
	Last4                 string    `json:"last4"`
	LastFour              string    `json:"last_four"` // Alias for backward compatibility
	ExpiryMonth           int       `json:"expiry_month,omitempty"`
	ExpiryYear            int       `json:"expiry_year,omitempty"`
	IsDefault             bool      `json:"is_default"`
	CreatedAt             time.Time `json:"created_at"`
}

// Invoice represents a billing invoice
type Invoice struct {
	ID             string         `json:"id"`
	OrganizationID string         `json:"organization_id"`
	SubscriptionID string         `json:"subscription_id"`
	Number         string         `json:"number"`
	Status         string         `json:"status"` // draft, open, paid, void, uncollectible
	Amount         float64        `json:"amount"`
	Currency       string         `json:"currency"`
	PeriodStart    time.Time      `json:"period_start"`
	PeriodEnd      time.Time      `json:"period_end"`
	DueDate        time.Time      `json:"due_date"`
	PaidAt         *time.Time     `json:"paid_at,omitempty"`
	LineItems      []*LineItem    `json:"line_items"`
	PaymentMethod  *PaymentMethod `json:"payment_method,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// LineItem represents an invoice line item
type LineItem struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
	Period      string  `json:"period,omitempty"`
}

// UsageRecord represents resource usage for billing
type UsageRecord struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	WorkspaceID    string    `json:"workspace_id,omitempty"`
	ResourceType   string    `json:"resource_type"` // cpu, memory, storage, bandwidth
	Quantity       float64   `json:"quantity"`
	Unit           string    `json:"unit"`
	Price          float64   `json:"price"`
	Amount         float64   `json:"amount"`
	Timestamp      time.Time `json:"timestamp"`
	RecordedAt     time.Time `json:"recorded_at"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
}

// CreateSubscriptionRequest represents a request to create a subscription
type CreateSubscriptionRequest struct {
	PlanID          string     `json:"plan_id" binding:"required"`
	BillingCycle    string     `json:"billing_cycle" binding:"required,oneof=monthly yearly"`
	PaymentMethodID string     `json:"payment_method_id,omitempty"`
	TrialEnd        *time.Time `json:"trial_end,omitempty"`
}

// UpdateSubscriptionRequest represents a request to update a subscription
type UpdateSubscriptionRequest struct {
	PlanID       string `json:"plan_id,omitempty"`
	BillingCycle string `json:"billing_cycle,omitempty"`
}

// CancelSubscriptionRequest represents a request to cancel a subscription
type CancelSubscriptionRequest struct {
	Immediate bool   `json:"immediate"`
	Reason    string `json:"reason,omitempty"`
}

// AddPaymentMethodRequest represents a request to add a payment method
type AddPaymentMethodRequest struct {
	Token           string `json:"token" binding:"required"`
	PaymentMethodID string `json:"payment_method_id,omitempty"`
	Type            string `json:"type,omitempty"`
	Last4           string `json:"last4,omitempty"`
	Brand           string `json:"brand,omitempty"`
	ExpiryMonth     int    `json:"expiry_month,omitempty"`
	ExpiryYear      int    `json:"expiry_year,omitempty"`
	SetDefault      bool   `json:"set_default"`
	SetAsDefault    bool   `json:"set_as_default"` // Alias for backward compatibility
}

// BillingOverview represents billing overview for an organization
type BillingOverview struct {
	Organization      *Organization      `json:"organization"`
	Subscription      *Subscription      `json:"subscription"`
	CurrentPlan       *Plan              `json:"current_plan"`
	PaymentMethods    []*PaymentMethod   `json:"payment_methods"`
	CurrentUsage      *CurrentUsage      `json:"current_usage"`
	UpcomingInvoice   *Invoice           `json:"upcoming_invoice,omitempty"`
	RecentInvoices    []*Invoice         `json:"recent_invoices"`
}

// CurrentUsage represents current billing period usage
type CurrentUsage struct {
	OrganizationID string           `json:"organization_id"`
	PeriodStart    time.Time        `json:"period_start"`
	PeriodEnd      time.Time        `json:"period_end"`
	EstimatedCost  float64          `json:"estimated_cost"`
	ResourceUsage  map[string]Usage `json:"resource_usage"`
	Usage          map[string]float64 `json:"usage"` // Raw usage values
}

// Usage represents resource usage
type Usage struct {
	Used      float64 `json:"used"`
	Limit     float64 `json:"limit"`
	Unit      string  `json:"unit"`
	Cost      float64 `json:"cost"`
	Overage   float64 `json:"overage,omitempty"`
}

// Organization represents billing organization info
type Organization struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	DisplayName      string `json:"display_name,omitempty"`
	Email            string `json:"email"`
	BillingEmail     string `json:"billing_email,omitempty"`
	TaxID            string `json:"tax_id,omitempty"`
	Address          string `json:"address,omitempty"`
	StripeCustomerID string `json:"stripe_customer_id,omitempty"`
}

// BillingSettings represents billing settings
type BillingSettings struct {
	OrganizationID          string    `json:"organization_id"`
	BillingEmail            string    `json:"billing_email"`
	InvoicePrefix           string    `json:"invoice_prefix"`
	TaxExempt               bool      `json:"tax_exempt"`
	TaxExemptionCertificate string    `json:"tax_exemption_certificate,omitempty"`
	TaxID                   string    `json:"tax_id,omitempty"`
	PurchaseOrderNumber     string    `json:"purchase_order_number,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// ResourceUsage is an alias for Usage (for backward compatibility)
type ResourceUsage = Usage