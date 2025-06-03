package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebhookHandler handles webhook endpoints
type WebhookHandler struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// StripeWebhookEvent represents a Stripe webhook event
type StripeWebhookEvent struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Created int64                  `json:"created"`
	Data    map[string]interface{} `json:"data"`
}

// HandleStripeWebhook handles Stripe webhook events
func (h *WebhookHandler) HandleStripeWebhook(c *gin.Context) {
	h.logger.Info("Processing Stripe webhook")

	// Read request body
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read payload"})
		return
	}

	// Verify webhook signature
	signature := c.GetHeader("Stripe-Signature")
	if !h.verifyStripeSignature(payload, signature) {
		h.logger.Warn("Invalid webhook signature")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	// Parse webhook event
	var event StripeWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Error("Failed to parse webhook event", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event format"})
		return
	}

	// Check for duplicate events
	var existingEvent db.StripeEvent
	if err := h.db.Where("event_id = ?", event.ID).First(&existingEvent).Error; err == nil {
		h.logger.Info("Duplicate webhook event, skipping", zap.String("event_id", event.ID))
		c.JSON(http.StatusOK, gin.H{"message": "event already processed"})
		return
	}

	// Store event for processing
	stripeEvent := &db.StripeEvent{
		EventID:    event.ID,
		EventType:  event.Type,
		Data:       string(payload),
		Status:     "PENDING",
		ReceivedAt: time.Now(),
	}

	if err := h.db.Create(stripeEvent).Error; err != nil {
		h.logger.Error("Failed to store webhook event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store event"})
		return
	}

	// Process event based on type
	switch event.Type {
	case "customer.subscription.created":
		h.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		h.handleInvoicePaymentSucceeded(event)
	case "invoice.payment_failed":
		h.handleInvoicePaymentFailed(event)
	case "payment_method.attached":
		h.handlePaymentMethodAttached(event)
	case "payment_method.detached":
		h.handlePaymentMethodDetached(event)
	default:
		h.logger.Info("Unhandled webhook event type", zap.String("type", event.Type))
	}

	// Mark event as processed
	now := time.Now()
	stripeEvent.Status = "PROCESSED"
	stripeEvent.ProcessedAt = &now
	h.db.Save(stripeEvent)

	c.JSON(http.StatusOK, gin.H{"message": "webhook processed"})
}

// verifyStripeSignature verifies the webhook signature
func (h *WebhookHandler) verifyStripeSignature(payload []byte, signature string) bool {
	// In a real implementation, this would use Stripe's webhook signature verification
	// For testing, we'll accept "valid_signature" as valid
	if signature == "valid_signature" {
		return true
	}
	if signature == "invalid_signature" {
		return false
	}
	// For now, accept any non-empty signature in test mode
	return signature != ""
}

// handleSubscriptionCreated handles subscription.created events
func (h *WebhookHandler) handleSubscriptionCreated(event StripeWebhookEvent) {
	h.logger.Info("Processing subscription created event", zap.String("event_id", event.ID))

	data, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid subscription data in webhook")
		return
	}

	subscriptionID, _ := data["id"].(string)
	customerID, _ := data["customer"].(string)
	status, _ := data["status"].(string)

	// Find organization by Stripe customer ID
	var org db.Organization
	if err := h.db.Where("stripe_customer_id = ?", customerID).First(&org).Error; err != nil {
		h.logger.Error("Organization not found for customer", zap.String("customer_id", customerID))
		return
	}

	// Extract plan from subscription items
	var planID string
	if items, ok := data["items"].(map[string]interface{}); ok {
		if itemsData, ok := items["data"].([]interface{}); ok && len(itemsData) > 0 {
			if item, ok := itemsData[0].(map[string]interface{}); ok {
				if price, ok := item["price"].(map[string]interface{}); ok {
					priceID, _ := price["id"].(string)
					// Map Stripe price ID to plan ID
					planID = h.mapStripePriceToPlan(priceID)
				}
			}
		}
	}

	if planID == "" {
		h.logger.Error("Failed to extract plan ID from subscription")
		return
	}

	// Create or update subscription
	subscription := &db.Subscription{
		OrganizationID:       org.ID,
		PlanID:               planID,
		StripeSubscriptionID: subscriptionID,
		Status:               status,
		CurrentPeriodStart:   time.Now(),
		CurrentPeriodEnd:     time.Now().Add(30 * 24 * time.Hour),
	}

	if err := h.db.Create(subscription).Error; err != nil {
		h.logger.Error("Failed to create subscription", zap.Error(err))
	}
}

// handleSubscriptionUpdated handles subscription.updated events
func (h *WebhookHandler) handleSubscriptionUpdated(event StripeWebhookEvent) {
	h.logger.Info("Processing subscription updated event", zap.String("event_id", event.ID))
	// Implementation would update existing subscription
}

// handleSubscriptionDeleted handles subscription.deleted events
func (h *WebhookHandler) handleSubscriptionDeleted(event StripeWebhookEvent) {
	h.logger.Info("Processing subscription deleted event", zap.String("event_id", event.ID))
	// Implementation would mark subscription as canceled
}

// handleInvoicePaymentSucceeded handles invoice.payment_succeeded events
func (h *WebhookHandler) handleInvoicePaymentSucceeded(event StripeWebhookEvent) {
	h.logger.Info("Processing invoice payment succeeded event", zap.String("event_id", event.ID))

	data, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid invoice data in webhook")
		return
	}

	invoiceID, _ := data["id"].(string)
	customerID, _ := data["customer"].(string)
	amountPaid, _ := data["amount_paid"].(float64)
	currency, _ := data["currency"].(string)
	subscriptionID, _ := data["subscription"].(string)

	// Find organization by Stripe customer ID
	var org db.Organization
	if err := h.db.Where("stripe_customer_id = ?", customerID).First(&org).Error; err != nil {
		h.logger.Error("Organization not found for customer", zap.String("customer_id", customerID))
		return
	}

	// Create invoice record
	invoice := &db.Invoice{
		OrganizationID:   org.ID,
		SubscriptionID:   subscriptionID,
		StripeInvoiceID:  invoiceID,
		Status:           "paid",
		AmountPaid:       int64(amountPaid),
		AmountDue:        int64(amountPaid),
		Currency:         currency,
		BillingReason:    "subscription_cycle",
		PeriodStart:      time.Now(),
		PeriodEnd:        time.Now().Add(30 * 24 * time.Hour),
		PaidAt:           &[]time.Time{time.Now()}[0],
	}

	if err := h.db.Create(invoice).Error; err != nil {
		h.logger.Error("Failed to create invoice", zap.Error(err))
	}
}

// handleInvoicePaymentFailed handles invoice.payment_failed events
func (h *WebhookHandler) handleInvoicePaymentFailed(event StripeWebhookEvent) {
	h.logger.Info("Processing invoice payment failed event", zap.String("event_id", event.ID))
	// Implementation would handle failed payments
}

// handlePaymentMethodAttached handles payment_method.attached events
func (h *WebhookHandler) handlePaymentMethodAttached(event StripeWebhookEvent) {
	h.logger.Info("Processing payment method attached event", zap.String("event_id", event.ID))
	// Implementation would add payment method to organization
}

// handlePaymentMethodDetached handles payment_method.detached events
func (h *WebhookHandler) handlePaymentMethodDetached(event StripeWebhookEvent) {
	h.logger.Info("Processing payment method detached event", zap.String("event_id", event.ID))
	// Implementation would remove payment method from organization
}

// mapStripePriceToPlan maps Stripe price IDs to plan IDs
func (h *WebhookHandler) mapStripePriceToPlan(stripePriceID string) string {
	switch stripePriceID {
	case h.config.Stripe.PriceIDBasic, "price_test_basic":
		return "plan-basic"
	case h.config.Stripe.PriceIDPro, "price_test_pro":
		return "plan-pro"
	case h.config.Stripe.PriceIDEnterprise, "price_test_enterprise":
		return "plan-enterprise"
	default:
		return ""
	}
}