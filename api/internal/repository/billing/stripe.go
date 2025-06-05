package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-kaas/api/internal/domain/billing"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/invoice"
	"github.com/stripe/stripe-go/v74/paymentmethod"
	"github.com/stripe/stripe-go/v74/subscription"
	"github.com/stripe/stripe-go/v74/webhook"
)

type stripeRepository struct {
	apiKey        string
	webhookSecret string
}

// NewStripeRepository creates a new Stripe billing repository
func NewStripeRepository(apiKey, webhookSecret string) billing.StripeRepository {
	stripe.Key = apiKey
	return &stripeRepository{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
	}
}

func (r *stripeRepository) CreateCustomer(ctx context.Context, org *billing.Organization) (string, error) {
	params := &stripe.CustomerParams{
		Name:  stripe.String(org.DisplayName),
		Email: stripe.String(org.BillingEmail),
		Metadata: map[string]string{
			"organization_id": org.ID,
		},
	}

	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	return cust.ID, nil
}

func (r *stripeRepository) UpdateCustomer(ctx context.Context, customerID string, org *billing.Organization) error {
	params := &stripe.CustomerParams{
		Name:  stripe.String(org.DisplayName),
		Email: stripe.String(org.BillingEmail),
	}

	_, err := customer.Update(customerID, params)
	if err != nil {
		return fmt.Errorf("failed to update Stripe customer: %w", err)
	}

	return nil
}

func (r *stripeRepository) DeleteCustomer(ctx context.Context, customerID string) error {
	_, err := customer.Del(customerID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete Stripe customer: %w", err)
	}

	return nil
}

func (r *stripeRepository) CreateStripeSubscription(ctx context.Context, customerID, priceID string) (*billing.StripeSubscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		ExpandParams: stripe.ExpandParams{
			Expand: []*string{
				stripe.String("latest_invoice.payment_intent"),
			},
		},
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe subscription: %w", err)
	}

	return r.convertStripeSubscription(sub), nil
}

func (r *stripeRepository) UpdateStripeSubscription(ctx context.Context, subscriptionID string, params map[string]interface{}) error {
	updateParams := &stripe.SubscriptionParams{}

	// Handle different update scenarios
	if items, ok := params["items"].([]map[string]interface{}); ok {
		// Update subscription items (e.g., changing plan)
		for _, item := range items {
			if priceID, ok := item["price"].(string); ok {
				updateParams.Items = []*stripe.SubscriptionItemsParams{
					{
						Price: stripe.String(priceID),
					},
				}
			}
		}
	}

	if cancelAtPeriodEnd, ok := params["cancel_at_period_end"].(bool); ok {
		updateParams.CancelAtPeriodEnd = stripe.Bool(cancelAtPeriodEnd)
	}

	_, err := subscription.Update(subscriptionID, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update Stripe subscription: %w", err)
	}

	return nil
}

func (r *stripeRepository) CancelStripeSubscription(ctx context.Context, subscriptionID string, immediate bool) error {
	if immediate {
		_, err := subscription.Cancel(subscriptionID, nil)
		if err != nil {
			return fmt.Errorf("failed to cancel Stripe subscription: %w", err)
		}
	} else {
		// Cancel at period end
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		_, err := subscription.Update(subscriptionID, params)
		if err != nil {
			return fmt.Errorf("failed to schedule Stripe subscription cancellation: %w", err)
		}
	}

	return nil
}

func (r *stripeRepository) AttachPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}

	_, err := paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		return fmt.Errorf("failed to attach payment method: %w", err)
	}

	return nil
}

func (r *stripeRepository) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("failed to detach payment method: %w", err)
	}

	return nil
}

func (r *stripeRepository) SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}

	_, err := customer.Update(customerID, params)
	if err != nil {
		return fmt.Errorf("failed to set default payment method: %w", err)
	}

	return nil
}

func (r *stripeRepository) CreateInvoice(ctx context.Context, customerID string) (*billing.StripeInvoice, error) {
	params := &stripe.InvoiceParams{
		Customer: stripe.String(customerID),
	}

	inv, err := invoice.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe invoice: %w", err)
	}

	return r.convertStripeInvoice(inv), nil
}

func (r *stripeRepository) FinalizeInvoice(ctx context.Context, invoiceID string) error {
	_, err := invoice.FinalizeInvoice(invoiceID, nil)
	if err != nil {
		return fmt.Errorf("failed to finalize Stripe invoice: %w", err)
	}

	return nil
}

func (r *stripeRepository) PayInvoice(ctx context.Context, invoiceID string) error {
	_, err := invoice.Pay(invoiceID, nil)
	if err != nil {
		return fmt.Errorf("failed to pay Stripe invoice: %w", err)
	}

	return nil
}

func (r *stripeRepository) VoidInvoice(ctx context.Context, invoiceID string) error {
	_, err := invoice.VoidInvoice(invoiceID, nil)
	if err != nil {
		return fmt.Errorf("failed to void Stripe invoice: %w", err)
	}

	return nil
}

func (r *stripeRepository) ConstructWebhookEvent(payload []byte, signature string) (*billing.StripeEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, r.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to construct webhook event: %w", err)
	}

	// Convert event data to map
	data := make(map[string]interface{})
	if err := event.Data.Raw.Unmarshal(&data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	return &billing.StripeEvent{
		ID:   event.ID,
		Type: string(event.Type),
		Data: data,
	}, nil
}

// Helper functions

func (r *stripeRepository) convertStripeSubscription(sub *stripe.Subscription) *billing.StripeSubscription {
	stripeSub := &billing.StripeSubscription{
		ID:                 sub.ID,
		CustomerID:         sub.Customer.ID,
		Status:             string(sub.Status),
		CurrentPeriodStart: time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:   time.Unix(sub.CurrentPeriodEnd, 0),
		Items:              []billing.StripeSubscriptionItem{},
	}

	if sub.CancelAt > 0 {
		cancelAt := time.Unix(sub.CancelAt, 0)
		stripeSub.CancelAt = &cancelAt
	}

	for _, item := range sub.Items.Data {
		stripeSub.Items = append(stripeSub.Items, billing.StripeSubscriptionItem{
			ID:       item.ID,
			PriceID:  item.Price.ID,
			Quantity: item.Quantity,
		})
	}

	return stripeSub
}

func (r *stripeRepository) convertStripeInvoice(inv *stripe.Invoice) *billing.StripeInvoice {
	stripeInv := &billing.StripeInvoice{
		ID:          inv.ID,
		Number:      inv.Number,
		CustomerID:  inv.Customer.ID,
		Amount:      inv.Total,
		Currency:    string(inv.Currency),
		Status:      string(inv.Status),
		PeriodStart: time.Unix(inv.PeriodStart, 0),
		PeriodEnd:   time.Unix(inv.PeriodEnd, 0),
		DueDate:     time.Unix(inv.DueDate, 0),
		Lines:       []billing.StripeLineItem{},
	}

	for _, line := range inv.Lines.Data {
		stripeInv.Lines = append(stripeInv.Lines, billing.StripeLineItem{
			Description: line.Description,
			Amount:      line.Amount,
			Quantity:    line.Quantity,
		})
	}

	return stripeInv
}