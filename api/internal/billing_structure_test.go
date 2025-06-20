package internal

import (
	"testing"

	billingDomain "github.com/hexabase/hexabase-ai/api/internal/billing/domain"
	billingHandler "github.com/hexabase/hexabase-ai/api/internal/billing/handler"
)

func TestBillingPackageStructure(t *testing.T) {
	// Domain layer types
	var subscription billingDomain.Subscription
	var invoice billingDomain.Invoice
	var paymentMethod billingDomain.PaymentMethod
	var usageRecord billingDomain.UsageRecord

	// Service interface
	var service billingDomain.Service

	// Repository interface  
	var repository billingDomain.Repository
	var stripeRepository billingDomain.StripeRepository

	// Service implementation (not exposed, uses interface)
	var serviceImpl billingDomain.Service

	// Repository implementations (not exposed, uses interface)
	var postgresRepo billingDomain.Repository
	var stripeRepo billingDomain.StripeRepository

	// Handler
	var handler *billingHandler.Handler

	// Test that all types are properly accessible
	_ = subscription
	_ = invoice
	_ = paymentMethod
	_ = usageRecord
	_ = service
	_ = repository
	_ = stripeRepository
	_ = serviceImpl
	_ = postgresRepo
	_ = stripeRepo
	_ = handler

	t.Log("Billing package structure is correctly organized")
} 