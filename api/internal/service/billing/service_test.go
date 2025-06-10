package billing

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/billing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) CreateSubscription(ctx context.Context, subscription *billing.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *mockRepository) GetSubscription(ctx context.Context, subscriptionID string) (*billing.Subscription, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) != nil {
		return args.Get(0).(*billing.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetOrganizationSubscription(ctx context.Context, organizationID string) (*billing.Subscription, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) != nil {
		return args.Get(0).(*billing.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetOrganization(ctx context.Context, organizationID string) (*billing.Organization, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) != nil {
		return args.Get(0).(*billing.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) UpdateSubscription(ctx context.Context, subscription *billing.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *mockRepository) CreateInvoice(ctx context.Context, invoice *billing.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *mockRepository) GetInvoice(ctx context.Context, invoiceID string) (*billing.Invoice, error) {
	args := m.Called(ctx, invoiceID)
	if args.Get(0) != nil {
		return args.Get(0).(*billing.Invoice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetInvoiceLineItems(ctx context.Context, invoiceID string) ([]*billing.InvoiceLineItem, error) {
	args := m.Called(ctx, invoiceID)
	if args.Get(0) != nil {
		return args.Get(0).([]*billing.InvoiceLineItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) ListInvoices(ctx context.Context, filter billing.InvoiceFilter) ([]*billing.Invoice, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) != nil {
		return args.Get(0).([]*billing.Invoice), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *mockRepository) UpdateInvoice(ctx context.Context, invoice *billing.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *mockRepository) CreateUsageRecord(ctx context.Context, usage *billing.UsageRecord) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *mockRepository) GetUsageRecords(ctx context.Context, filter billing.UsageFilter) ([]*billing.UsageRecord, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) != nil {
		return args.Get(0).([]*billing.UsageRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) SummarizeUsage(ctx context.Context, organizationID string, startTime, endTime time.Time) (map[string]float64, error) {
	args := m.Called(ctx, organizationID, startTime, endTime)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]float64), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*billing.PaymentMethod, error) {
	args := m.Called(ctx, paymentMethodID)
	if args.Get(0) != nil {
		return args.Get(0).(*billing.PaymentMethod), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) ListPaymentMethods(ctx context.Context, organizationID string) ([]*billing.PaymentMethod, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) != nil {
		return args.Get(0).([]*billing.PaymentMethod), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) CreatePaymentMethod(ctx context.Context, paymentMethod *billing.PaymentMethod) error {
	args := m.Called(ctx, paymentMethod)
	return args.Error(0)
}

func (m *mockRepository) SetDefaultPaymentMethod(ctx context.Context, organizationID, paymentMethodID string) error {
	args := m.Called(ctx, organizationID, paymentMethodID)
	return args.Error(0)
}

func (m *mockRepository) DeletePaymentMethod(ctx context.Context, paymentMethodID string) error {
	args := m.Called(ctx, paymentMethodID)
	return args.Error(0)
}

func (m *mockRepository) GetPlan(ctx context.Context, planID string) (*billing.Plan, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) != nil {
		return args.Get(0).(*billing.Plan), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) ListPlans(ctx context.Context, activeOnly bool) ([]*billing.Plan, error) {
	args := m.Called(ctx, activeOnly)
	if args.Get(0) != nil {
		return args.Get(0).([]*billing.Plan), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetBillingSettings(ctx context.Context, organizationID string) (*billing.BillingSettings, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) != nil {
		return args.Get(0).(*billing.BillingSettings), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) UpdateBillingSettings(ctx context.Context, settings *billing.BillingSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

// Mock stripe repository
type mockStripeRepository struct {
	mock.Mock
}

func (m *mockStripeRepository) CreateStripeCustomer(ctx context.Context, org *billing.Organization) (string, error) {
	args := m.Called(ctx, org)
	return args.String(0), args.Error(1)
}

func (m *mockStripeRepository) CreateStripeSubscription(ctx context.Context, customerID, priceID string) (*billing.StripeSubscription, error) {
	args := m.Called(ctx, customerID, priceID)
	if args.Get(0) != nil {
		return args.Get(0).(*billing.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockStripeRepository) UpdateStripeSubscription(ctx context.Context, subscriptionID string, params map[string]interface{}) error {
	args := m.Called(ctx, subscriptionID, params)
	return args.Error(0)
}

func (m *mockStripeRepository) CancelStripeSubscription(ctx context.Context, subscriptionID string, immediate bool) error {
	args := m.Called(ctx, subscriptionID, immediate)
	return args.Error(0)
}

func (m *mockStripeRepository) AttachPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	args := m.Called(ctx, customerID, paymentMethodID)
	return args.Error(0)
}

func (m *mockStripeRepository) SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	args := m.Called(ctx, customerID, paymentMethodID)
	return args.Error(0)
}

func (m *mockStripeRepository) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	args := m.Called(ctx, paymentMethodID)
	return args.Error(0)
}

func (m *mockStripeRepository) ConstructWebhookEvent(payload []byte, signature string) (*billing.StripeEvent, error) {
	args := m.Called(payload, signature)
	if args.Get(0) != nil {
		return args.Get(0).(*billing.StripeEvent), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestService_CreateSubscription(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockStripe := new(mockStripeRepository)
	
	svc := &service{
		repo:       mockRepo,
		stripeRepo: mockStripe,
		logger:     slog.Default(),
	}

	t.Run("successful subscription creation", func(t *testing.T) {
		orgID := "org-123"
		req := &billing.CreateSubscriptionRequest{
			PlanID: "plan-premium",
		}
		
		org := &billing.Organization{
			ID:               orgID,
			StripeCustomerID: "cus_123",
		}
		
		plan := &billing.Plan{
			ID:            "plan-premium",
			StripePriceID: "price_123",
		}
		
		stripeSub := &billing.StripeSubscription{
			ID:                 "sub_123",
			Status:             "active",
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		}
		
		// Mock calls
		mockRepo.On("GetOrganization", ctx, orgID).Return(org, nil)
		mockRepo.On("GetPlan", ctx, "plan-premium").Return(plan, nil)
		mockStripe.On("CreateStripeSubscription", ctx, "cus_123", "price_123").Return(stripeSub, nil)
		mockRepo.On("CreateSubscription", ctx, mock.AnythingOfType("*billing.Subscription")).Return(nil)

		sub, err := svc.CreateSubscription(ctx, orgID, req)
		assert.NoError(t, err)
		assert.NotNil(t, sub)
		assert.Equal(t, orgID, sub.OrganizationID)
		assert.Equal(t, "plan-premium", sub.PlanID)
		assert.Equal(t, "active", sub.Status)
		assert.Equal(t, "sub_123", sub.StripeSubscriptionID)

		mockRepo.AssertExpectations(t)
		mockStripe.AssertExpectations(t)
	})

	t.Run("organization not found", func(t *testing.T) {
		orgID := "org-456"
		req := &billing.CreateSubscriptionRequest{
			PlanID: "plan-basic",
		}
		
		mockRepo.On("GetOrganization", ctx, orgID).Return(nil, errors.New("not found"))

		sub, err := svc.CreateSubscription(ctx, orgID, req)
		assert.Error(t, err)
		assert.Nil(t, sub)
		assert.Contains(t, err.Error(), "organization not found")

		mockRepo.AssertExpectations(t)
	})
}

func TestService_UpdateSubscription(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockStripe := new(mockStripeRepository)
	
	svc := &service{
		repo:       mockRepo,
		stripeRepo: mockStripe,
		logger:     slog.Default(),
	}

	t.Run("successful plan upgrade", func(t *testing.T) {
		subID := uuid.New().String()
		req := &billing.UpdateSubscriptionRequest{
			PlanID: "plan-enterprise",
		}
		
		existingSub := &billing.Subscription{
			ID:                   subID,
			OrganizationID:       "org-123",
			PlanID:               "plan-premium",
			Status:               "active",
			StripeSubscriptionID: "sub_123",
		}
		
		newPlan := &billing.Plan{
			ID:            "plan-enterprise",
			StripePriceID: "price_enterprise",
		}
		
		expectedParams := map[string]interface{}{
			"items": []map[string]interface{}{
				{"price": "price_enterprise"},
			},
		}
		
		mockRepo.On("GetSubscription", ctx, subID).Return(existingSub, nil)
		mockRepo.On("GetPlan", ctx, "plan-enterprise").Return(newPlan, nil)
		mockStripe.On("UpdateStripeSubscription", ctx, "sub_123", expectedParams).Return(nil)
		mockRepo.On("UpdateSubscription", ctx, mock.AnythingOfType("*billing.Subscription")).Return(nil)

		updatedSub, err := svc.UpdateSubscription(ctx, subID, req)
		assert.NoError(t, err)
		assert.NotNil(t, updatedSub)
		assert.Equal(t, "plan-enterprise", updatedSub.PlanID)

		mockRepo.AssertExpectations(t)
		mockStripe.AssertExpectations(t)
	})

	t.Run("subscription not found", func(t *testing.T) {
		subID := uuid.New().String()
		req := &billing.UpdateSubscriptionRequest{
			PlanID: "plan-basic",
		}
		
		mockRepo.On("GetSubscription", ctx, subID).Return(nil, errors.New("not found"))

		updatedSub, err := svc.UpdateSubscription(ctx, subID, req)
		assert.Error(t, err)
		assert.Nil(t, updatedSub)

		mockRepo.AssertExpectations(t)
	})
}

func TestService_CancelSubscription(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockStripe := new(mockStripeRepository)
	
	svc := &service{
		repo:       mockRepo,
		stripeRepo: mockStripe,
		logger:     slog.Default(),
	}

	t.Run("successful immediate cancellation", func(t *testing.T) {
		subID := uuid.New().String()
		req := &billing.CancelSubscriptionRequest{
			Immediate: true,
		}
		
		activeSub := &billing.Subscription{
			ID:                   subID,
			OrganizationID:       "org-789",
			PlanID:               "plan-premium",
			Status:               "active",
			StripeSubscriptionID: "sub_456",
		}
		
		mockRepo.On("GetSubscription", ctx, subID).Return(activeSub, nil)
		mockStripe.On("CancelStripeSubscription", ctx, "sub_456", true).Return(nil)
		mockRepo.On("UpdateSubscription", ctx, mock.AnythingOfType("*billing.Subscription")).Return(nil)

		err := svc.CancelSubscription(ctx, subID, req)
		assert.NoError(t, err)

		// Verify the subscription status was updated
		updateCall := mockRepo.Calls[1]
		updatedSub := updateCall.Arguments[1].(*billing.Subscription)
		assert.Equal(t, "canceled", updatedSub.Status)
		assert.NotNil(t, updatedSub.CanceledAt)

		mockRepo.AssertExpectations(t)
		mockStripe.AssertExpectations(t)
	})

	t.Run("cancel at period end", func(t *testing.T) {
		subID := uuid.New().String()
		req := &billing.CancelSubscriptionRequest{
			Immediate: false,
		}
		
		periodEnd := time.Now().AddDate(0, 1, 0)
		activeSub := &billing.Subscription{
			ID:                   subID,
			OrganizationID:       "org-999",
			PlanID:               "plan-basic",
			Status:               "active",
			StripeSubscriptionID: "sub_789",
			CurrentPeriodEnd:     periodEnd,
		}
		
		mockRepo.On("GetSubscription", ctx, subID).Return(activeSub, nil)
		mockStripe.On("CancelStripeSubscription", ctx, "sub_789", false).Return(nil)
		mockRepo.On("UpdateSubscription", ctx, mock.AnythingOfType("*billing.Subscription")).Return(nil)

		err := svc.CancelSubscription(ctx, subID, req)
		assert.NoError(t, err)

		// Verify the subscription status was updated
		updateCall := mockRepo.Calls[1]
		updatedSub := updateCall.Arguments[1].(*billing.Subscription)
		assert.Equal(t, "cancel_at_period_end", updatedSub.Status)
		assert.NotNil(t, updatedSub.CancelAt)
		assert.Equal(t, periodEnd, *updatedSub.CancelAt)

		mockRepo.AssertExpectations(t)
		mockStripe.AssertExpectations(t)
	})
}

func TestService_GetCurrentUsage(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	
	svc := &service{
		repo:   mockRepo,
		logger: slog.Default(),
	}

	t.Run("get usage for current period", func(t *testing.T) {
		orgID := "org-usage"
		now := time.Now()
		periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		periodEnd := periodStart.AddDate(0, 1, 0)
		
		sub := &billing.Subscription{
			ID:                 uuid.New().String(),
			OrganizationID:     orgID,
			PlanID:             "plan-premium",
			CurrentPeriodStart: periodStart,
			CurrentPeriodEnd:   periodEnd,
		}
		
		plan := &billing.Plan{
			ID: "plan-premium",
			Limits: &billing.PlanLimits{
				CPUCores:  8,
				MemoryGB:  16,
				StorageGB: 500,
			},
		}
		
		usageSummary := map[string]float64{
			"cpu_cores":  4.5,
			"memory_gb":  10.0,
			"storage_gb": 250.0,
		}
		
		mockRepo.On("GetOrganizationSubscription", ctx, orgID).Return(sub, nil)
		mockRepo.On("GetPlan", ctx, "plan-premium").Return(plan, nil)
		mockRepo.On("SummarizeUsage", ctx, orgID, periodStart, mock.AnythingOfType("time.Time")).
			Return(usageSummary, nil)

		usage, err := svc.GetCurrentUsage(ctx, orgID)
		assert.NoError(t, err)
		assert.NotNil(t, usage)
		assert.Equal(t, orgID, usage.OrganizationID)
		assert.Len(t, usage.ResourceUsage, 3)
		assert.Equal(t, 4.5, usage.ResourceUsage["cpu_cores"].Used)
		assert.Equal(t, 8.0, usage.ResourceUsage["cpu_cores"].Limit)

		mockRepo.AssertExpectations(t)
	})

	t.Run("no active subscription", func(t *testing.T) {
		orgID := "org-no-sub"
		
		mockRepo.On("GetOrganizationSubscription", ctx, orgID).Return(nil, errors.New("not found"))

		usage, err := svc.GetCurrentUsage(ctx, orgID)
		assert.Error(t, err)
		assert.Nil(t, usage)
		assert.Contains(t, err.Error(), "no active subscription")

		mockRepo.AssertExpectations(t)
	})
}

func TestService_RecordUsage(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	
	svc := &service{
		repo:   mockRepo,
		logger: slog.Default(),
	}

	t.Run("record CPU usage", func(t *testing.T) {
		usage := &billing.UsageRecord{
			OrganizationID: "org-record",
			ResourceType:   "cpu_cores",
			Quantity:       4.5,
			WorkspaceID:    strPtr("ws-123"),
			ApplicationID:  strPtr("app-456"),
		}
		
		mockRepo.On("CreateUsageRecord", ctx, mock.AnythingOfType("*billing.UsageRecord")).Return(nil)

		err := svc.RecordUsage(ctx, usage)
		assert.NoError(t, err)

		// Verify the usage record has an ID and timestamp
		createCall := mockRepo.Calls[0]
		createdUsage := createCall.Arguments[1].(*billing.UsageRecord)
		assert.NotEmpty(t, createdUsage.ID)
		assert.False(t, createdUsage.RecordedAt.IsZero())

		mockRepo.AssertExpectations(t)
	})
}

func TestService_ListInvoices(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	
	svc := &service{
		repo:   mockRepo,
		logger: slog.Default(),
	}

	t.Run("list organization invoices", func(t *testing.T) {
		orgID := "org-invoices"
		filter := billing.InvoiceFilter{
			Status:   "paid",
			PageSize: 20,
		}
		
		invoices := []*billing.Invoice{
			{
				ID:             uuid.New().String(),
				OrganizationID: orgID,
				Amount:         9900, // $99.00
				Status:         "paid",
			},
			{
				ID:             uuid.New().String(),
				OrganizationID: orgID,
				Amount:         9900,
				Status:         "paid",
			},
		}
		
		mockRepo.On("ListInvoices", ctx, filter).Return(invoices, 2, nil)

		result, total, err := svc.ListInvoices(ctx, orgID, filter)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, total)

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty invoice list", func(t *testing.T) {
		orgID := "org-no-invoices"
		filter := billing.InvoiceFilter{
			Status:   "pending",
			PageSize: 10,
		}
		
		mockRepo.On("ListInvoices", ctx, filter).Return([]*billing.Invoice{}, 0, nil)

		result, total, err := svc.ListInvoices(ctx, orgID, filter)
		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Equal(t, 0, total)

		mockRepo.AssertExpectations(t)
	})
}

// Helper function
func strPtr(s string) *string {
	return &s
}