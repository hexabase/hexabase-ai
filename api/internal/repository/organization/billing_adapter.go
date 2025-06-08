package organization

import (
	"context"
	"fmt"
	"time"

	billingDomain "github.com/hexabase/hexabase-ai/api/internal/domain/billing"
	"github.com/hexabase/hexabase-ai/api/internal/domain/organization"
)

// BillingRepositoryAdapter adapts billing.StripeRepository to organization.BillingRepository
type BillingRepositoryAdapter struct {
	stripeRepo billingDomain.StripeRepository
}

// NewBillingRepositoryAdapter creates a new billing repository adapter for organization domain
func NewBillingRepositoryAdapter(stripeRepo billingDomain.StripeRepository) organization.BillingRepository {
	return &BillingRepositoryAdapter{stripeRepo: stripeRepo}
}

// CreateCustomer creates a Stripe customer for the organization
func (b *BillingRepositoryAdapter) CreateCustomer(ctx context.Context, org *organization.Organization) (string, error) {
	// Convert organization.Organization to billing.Organization
	billingOrg := &billingDomain.Organization{
		ID:          org.ID,
		Name:        org.Name,
		DisplayName: org.DisplayName,
		Email:       org.Email,
	}

	customerID, err := b.stripeRepo.CreateCustomer(ctx, billingOrg)
	if err != nil {
		return "", fmt.Errorf("failed to create customer: %w", err)
	}

	return customerID, nil
}

// DeleteCustomer deletes a Stripe customer
func (b *BillingRepositoryAdapter) DeleteCustomer(ctx context.Context, customerID string) error {
	err := b.stripeRepo.DeleteCustomer(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	return nil
}

// GetOrganizationSubscription gets the subscription for an organization
func (b *BillingRepositoryAdapter) GetOrganizationSubscription(ctx context.Context, orgID string) (*organization.Subscription, error) {
	// Note: This is a placeholder implementation
	// In a real implementation, we would need to:
	// 1. Get the organization's customer ID from the database
	// 2. Query Stripe for the subscription using the customer ID
	// 3. Convert the Stripe subscription to organization.Subscription
	
	// For now, return a placeholder implementation
	return &organization.Subscription{
		PlanID:           "free",
		PlanName:         "Free Plan",
		Status:           "active",
		CurrentPeriodEnd: time.Now().AddDate(1, 0, 0), // 1 year from now
	}, nil
}

// CancelSubscription cancels a subscription for an organization
func (b *BillingRepositoryAdapter) CancelSubscription(ctx context.Context, orgID string) error {
	// Note: This is a placeholder implementation
	// In a real implementation, we would need to:
	// 1. Get the organization's subscription ID from the database
	// 2. Cancel the subscription using Stripe API
	
	// For now, return success
	return nil
}