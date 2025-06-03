package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BillingHandler handles billing-related endpoints
type BillingHandler struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *BillingHandler {
	return &BillingHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// CreateSubscriptionRequest represents the request to create a subscription
type CreateSubscriptionRequest struct {
	PlanID          string `json:"plan_id" binding:"required"`
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
}

// UpdateSubscriptionRequest represents the request to update a subscription
type UpdateSubscriptionRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}

// CancelSubscriptionRequest represents the request to cancel a subscription
type CancelSubscriptionRequest struct {
	Immediate bool `json:"immediate"`
}

// AddPaymentMethodRequest represents the request to add a payment method
type AddPaymentMethodRequest struct {
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
	SetDefault      bool   `json:"set_default"`
}

// ReportUsageRequest represents the request to report usage
type ReportUsageRequest struct {
	WorkspaceID string    `json:"workspace_id" binding:"required"`
	MetricType  string    `json:"metric_type" binding:"required"`
	Quantity    float64   `json:"quantity" binding:"required"`
	Timestamp   time.Time `json:"timestamp"`
}

// CreateSubscription creates a new subscription for an organization
func (h *BillingHandler) CreateSubscription(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Check for specific field validation errors
		if req.PlanID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "plan_id is required"})
			return
		}
		if req.PaymentMethodID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payment_method_id is required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Check if organization already has an active subscription
	var existingSubscription db.Subscription
	if err := h.db.Where("organization_id = ? AND status IN ('active', 'trialing')", orgID).First(&existingSubscription).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "organization already has an active subscription"})
		return
	}

	// Verify plan exists
	var plan db.Plan
	if err := h.db.First(&plan, "id = ? AND is_active = true", req.PlanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan"})
		} else {
			h.logger.Error("Failed to get plan", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan"})
		}
		return
	}

	// TODO: Create subscription in Stripe
	// For now, we'll create a mock subscription
	subscription := &db.Subscription{
		OrganizationID:       orgID,
		PlanID:               req.PlanID,
		StripeSubscriptionID: "sub_mock_" + time.Now().Format("20060102150405"),
		Status:               "active",
		CurrentPeriodStart:   time.Now(),
		CurrentPeriodEnd:     time.Now().Add(30 * 24 * time.Hour),
	}

	if err := h.db.Create(subscription).Error; err != nil {
		h.logger.Error("Failed to create subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	h.logger.Info("Subscription created successfully",
		zap.String("subscription_id", subscription.ID),
		zap.String("org_id", orgID),
		zap.String("plan_id", req.PlanID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription gets the current subscription for an organization
func (h *BillingHandler) GetSubscription(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var subscription db.Subscription
	if err := h.db.Where("organization_id = ? AND status IN ('active', 'trialing', 'past_due')", orgID).
		Preload("Plan").
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription found"})
		} else {
			h.logger.Error("Failed to get subscription", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		}
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// UpdateSubscription updates the subscription plan
func (h *BillingHandler) UpdateSubscription(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.PlanID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "plan_id is required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Get current subscription (order by created_at DESC to get latest)
	var subscription db.Subscription
	if err := h.db.Where("organization_id = ? AND status IN ('active', 'trialing')", orgID).Order("created_at DESC").First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription found"})
		} else {
			h.logger.Error("Failed to get subscription", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		}
		return
	}

	// Verify new plan exists
	var plan db.Plan
	if err := h.db.First(&plan, "id = ? AND is_active = true", req.PlanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan"})
		} else {
			h.logger.Error("Failed to get plan", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan"})
		}
		return
	}

	// TODO: Update subscription in Stripe
	// For now, we'll just update the plan ID
	subscription.PlanID = req.PlanID
	subscription.UpdatedAt = time.Now()

	if err := h.db.Save(&subscription).Error; err != nil {
		h.logger.Error("Failed to update subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}

	h.logger.Info("Subscription updated successfully",
		zap.String("subscription_id", subscription.ID),
		zap.String("org_id", orgID),
		zap.String("new_plan_id", req.PlanID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, subscription)
}

// CancelSubscription cancels a subscription
func (h *BillingHandler) CancelSubscription(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var req CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, default to end of period
		req.Immediate = false
	}

	// Get current subscription (order by created_at DESC to get latest)
	var subscription db.Subscription
	if err := h.db.Where("organization_id = ? AND status IN ('active', 'trialing')", orgID).Order("created_at DESC").First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription found"})
		} else {
			h.logger.Error("Failed to get subscription", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		}
		return
	}

	// TODO: Cancel subscription in Stripe
	// For now, we'll just update the status
	now := time.Now()
	if req.Immediate {
		subscription.Status = "canceled"
		subscription.CanceledAt = &now
	} else {
		subscription.CancelAtPeriodEnd = true
		subscription.CanceledAt = &now
	}

	if err := h.db.Save(&subscription).Error; err != nil {
		h.logger.Error("Failed to cancel subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel subscription"})
		return
	}

	h.logger.Info("Subscription canceled successfully",
		zap.String("subscription_id", subscription.ID),
		zap.String("org_id", orgID),
		zap.Bool("immediate", req.Immediate),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{
		"message":   "subscription canceled successfully",
		"immediate": req.Immediate,
	})
}

// ListPaymentMethods lists all payment methods for an organization
func (h *BillingHandler) ListPaymentMethods(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var paymentMethods []db.PaymentMethod
	if err := h.db.Where("organization_id = ?", orgID).Order("is_default DESC, created_at DESC").Find(&paymentMethods).Error; err != nil {
		h.logger.Error("Failed to list payment methods", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list payment methods"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_methods": paymentMethods,
		"total":           len(paymentMethods),
	})
}

// AddPaymentMethod adds a new payment method
func (h *BillingHandler) AddPaymentMethod(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var req AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.PaymentMethodID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payment_method_id is required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// TODO: Attach payment method to customer in Stripe
	// For now, we'll create a mock payment method
	paymentMethod := &db.PaymentMethod{
		OrganizationID:        orgID,
		StripePaymentMethodID: req.PaymentMethodID,
		Type:                  "card",
		Card: &db.CardDetails{
			Brand:    "visa",
			Last4:    "4242",
			ExpMonth: 12,
			ExpYear:  2025,
		},
		IsDefault: req.SetDefault,
	}

	// If setting as default, unset other defaults
	if req.SetDefault {
		h.db.Model(&db.PaymentMethod{}).Where("organization_id = ?", orgID).Update("is_default", false)
	}

	if err := h.db.Create(paymentMethod).Error; err != nil {
		h.logger.Error("Failed to add payment method", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add payment method"})
		return
	}

	h.logger.Info("Payment method added successfully",
		zap.String("payment_method_id", paymentMethod.ID),
		zap.String("org_id", orgID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusCreated, paymentMethod)
}

// RemovePaymentMethod removes a payment method
func (h *BillingHandler) RemovePaymentMethod(c *gin.Context) {
	orgID := c.Param("orgId")
	pmID := c.Param("pmId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Get payment method
	var paymentMethod db.PaymentMethod
	if err := h.db.Where("id = ? AND organization_id = ?", pmID, orgID).First(&paymentMethod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment method not found"})
		} else {
			h.logger.Error("Failed to get payment method", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment method"})
		}
		return
	}

	// TODO: Detach payment method from customer in Stripe

	if err := h.db.Delete(&paymentMethod).Error; err != nil {
		h.logger.Error("Failed to remove payment method", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove payment method"})
		return
	}

	h.logger.Info("Payment method removed successfully",
		zap.String("payment_method_id", pmID),
		zap.String("org_id", orgID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{"message": "payment method removed successfully"})
}

// SetDefaultPaymentMethod sets a payment method as default
func (h *BillingHandler) SetDefaultPaymentMethod(c *gin.Context) {
	orgID := c.Param("orgId")
	pmID := c.Param("pmId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Get payment method
	var paymentMethod db.PaymentMethod
	if err := h.db.Where("id = ? AND organization_id = ?", pmID, orgID).First(&paymentMethod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment method not found"})
		} else {
			h.logger.Error("Failed to get payment method", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment method"})
		}
		return
	}

	// Unset other defaults
	h.db.Model(&db.PaymentMethod{}).Where("organization_id = ?", orgID).Update("is_default", false)

	// Set this as default
	paymentMethod.IsDefault = true
	if err := h.db.Save(&paymentMethod).Error; err != nil {
		h.logger.Error("Failed to set default payment method", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set default payment method"})
		return
	}

	// TODO: Update default payment method in Stripe

	h.logger.Info("Default payment method set successfully",
		zap.String("payment_method_id", pmID),
		zap.String("org_id", orgID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, paymentMethod)
}

// ListInvoices lists all invoices for an organization
func (h *BillingHandler) ListInvoices(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var invoices []db.Invoice
	if err := h.db.Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Preload("Subscription").
		Find(&invoices).Error; err != nil {
		h.logger.Error("Failed to list invoices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list invoices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": invoices,
		"total":    len(invoices),
	})
}

// GetInvoice gets a specific invoice
func (h *BillingHandler) GetInvoice(c *gin.Context) {
	orgID := c.Param("orgId")
	invoiceID := c.Param("invoiceId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var invoice db.Invoice
	if err := h.db.Where("id = ? AND organization_id = ?", invoiceID, orgID).
		Preload("Subscription").
		First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		} else {
			h.logger.Error("Failed to get invoice", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get invoice"})
		}
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// DownloadInvoice downloads an invoice PDF
func (h *BillingHandler) DownloadInvoice(c *gin.Context) {
	orgID := c.Param("orgId")
	invoiceID := c.Param("invoiceId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var invoice db.Invoice
	if err := h.db.Where("id = ? AND organization_id = ?", invoiceID, orgID).First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		} else {
			h.logger.Error("Failed to get invoice", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get invoice"})
		}
		return
	}

	// TODO: Redirect to Stripe invoice PDF URL
	// For now, return the URL
	c.JSON(http.StatusOK, gin.H{
		"download_url": invoice.InvoicePDFURL,
	})
}

// GetUsage gets usage data for an organization
func (h *BillingHandler) GetUsage(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Get billing period from query params (default to current month)
	billingPeriod := c.DefaultQuery("period", time.Now().Format("2006-01"))

	var usageRecords []db.UsageRecord
	if err := h.db.Where("organization_id = ? AND billing_period = ?", orgID, billingPeriod).
		Preload("Workspace").
		Find(&usageRecords).Error; err != nil {
		h.logger.Error("Failed to get usage records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get usage records"})
		return
	}

	// Aggregate usage by metric type
	usage := make(map[string]float64)
	for _, record := range usageRecords {
		usage[record.MetricType] += record.Quantity
	}

	c.JSON(http.StatusOK, gin.H{
		"period": billingPeriod,
		"usage":  usage,
		"details": usageRecords,
	})
}

// ReportUsage reports usage for a workspace
func (h *BillingHandler) ReportUsage(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	var req ReportUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.WorkspaceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
			return
		}
		if req.MetricType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "metric_type is required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Validate metric type
	validMetrics := []string{"cpu_hours", "memory_gb_hours", "storage_gb_days", "network_gb", "api_calls"}
	isValid := false
	for _, metric := range validMetrics {
		if req.MetricType == metric {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric_type"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.Where("id = ? AND organization_id = ?", req.WorkspaceID, orgID).First(&workspace).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Set default timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Create usage record
	usageRecord := &db.UsageRecord{
		OrganizationID: orgID,
		WorkspaceID:    req.WorkspaceID,
		MetricType:     req.MetricType,
		Quantity:       req.Quantity,
		Unit:           getUnitForMetric(req.MetricType),
		Timestamp:      req.Timestamp,
	}

	if err := h.db.Create(usageRecord).Error; err != nil {
		h.logger.Error("Failed to create usage record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create usage record"})
		return
	}

	h.logger.Info("Usage reported successfully",
		zap.String("usage_id", usageRecord.ID),
		zap.String("org_id", orgID),
		zap.String("workspace_id", req.WorkspaceID),
		zap.String("metric_type", req.MetricType),
		zap.Float64("quantity", req.Quantity))

	c.JSON(http.StatusCreated, usageRecord)
}

// CreatePortalSession creates a Stripe billing portal session
func (h *BillingHandler) CreatePortalSession(c *gin.Context) {
	orgID := c.Param("orgId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Get organization
	var org db.Organization
	if err := h.db.First(&org, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		} else {
			h.logger.Error("Failed to get organization", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get organization"})
		}
		return
	}

	// TODO: Create billing portal session in Stripe
	// For now, return a mock URL
	portalURL := "https://billing.stripe.com/session/mock_" + time.Now().Format("20060102150405")

	h.logger.Info("Billing portal session created",
		zap.String("org_id", orgID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{
		"url": portalURL,
	})
}

// hasOrgAccess checks if user has access to organization
func (h *BillingHandler) hasOrgAccess(userID, orgID string) bool {
	var count int64
	h.db.Model(&db.OrganizationUser{}).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Count(&count)
	return count > 0
}

// getUnitForMetric returns the unit for a given metric type
func getUnitForMetric(metricType string) string {
	switch metricType {
	case "cpu_hours":
		return "hours"
	case "memory_gb_hours":
		return "GB-hours"
	case "storage_gb_days":
		return "GB-days"
	case "network_gb":
		return "GB"
	case "api_calls":
		return "calls"
	default:
		return "units"
	}
}

// HandleStripeWebhook is defined in webhooks.go