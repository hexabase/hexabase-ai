package handlers

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/billing"
)

// BillingHandler handles billing-related HTTP requests
type BillingHandler struct {
	service billing.Service
	logger  *slog.Logger
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(service billing.Service, logger *slog.Logger) *BillingHandler {
	return &BillingHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSubscription handles subscription creation
func (h *BillingHandler) CreateSubscription(c *gin.Context) {
	orgID := c.Param("orgId")

	var req billing.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	sub, err := h.service.CreateSubscription(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to create subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// GetSubscription handles getting organization subscription
func (h *BillingHandler) GetSubscription(c *gin.Context) {
	orgID := c.Param("orgId")

	sub, err := h.service.GetOrganizationSubscription(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get subscription", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription handles subscription updates
func (h *BillingHandler) UpdateSubscription(c *gin.Context) {
	orgID := c.Param("orgId")

	var req billing.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Get subscription for organization
	currentSub, err := h.service.GetOrganizationSubscription(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get current subscription", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	sub, err := h.service.UpdateSubscription(c.Request.Context(), currentSub.ID, &req)
	if err != nil {
		h.logger.Error("failed to update subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// CancelSubscription handles subscription cancellation
func (h *BillingHandler) CancelSubscription(c *gin.Context) {
	orgID := c.Param("orgId")

	var req billing.CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Immediate = false // Default to cancel at period end
	}

	// Get subscription for organization
	currentSub, err := h.service.GetOrganizationSubscription(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get current subscription", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	err = h.service.CancelSubscription(c.Request.Context(), currentSub.ID, &req)
	if err != nil {
		h.logger.Error("failed to cancel subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription cancelled successfully"})
}

// ListPlans handles listing available plans
func (h *BillingHandler) ListPlans(c *gin.Context) {
	plans, err := h.service.ListPlans(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list plans", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list plans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
		"total": len(plans),
	})
}

// ComparePlans handles plan comparison
func (h *BillingHandler) ComparePlans(c *gin.Context) {
	currentPlanID := c.Query("current")
	targetPlanID := c.Query("target")

	if currentPlanID == "" || targetPlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current and target plan IDs are required"})
		return
	}

	comparison, err := h.service.ComparePlans(c.Request.Context(), currentPlanID, targetPlanID)
	if err != nil {
		h.logger.Error("failed to compare plans", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comparison)
}

// AddPaymentMethod handles adding a payment method
func (h *BillingHandler) AddPaymentMethod(c *gin.Context) {
	orgID := c.Param("orgId")

	var req billing.AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	method, err := h.service.AddPaymentMethod(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to add payment method", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, method)
}

// ListPaymentMethods handles listing payment methods
func (h *BillingHandler) ListPaymentMethods(c *gin.Context) {
	orgID := c.Param("orgId")

	methods, err := h.service.ListPaymentMethods(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to list payment methods", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_methods": methods,
		"total":           len(methods),
	})
}

// SetDefaultPaymentMethod handles setting default payment method
func (h *BillingHandler) SetDefaultPaymentMethod(c *gin.Context) {
	methodID := c.Param("methodId")

	err := h.service.SetDefaultPaymentMethod(c.Request.Context(), methodID)
	if err != nil {
		h.logger.Error("failed to set default payment method", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "default payment method updated"})
}

// RemovePaymentMethod handles removing a payment method
func (h *BillingHandler) RemovePaymentMethod(c *gin.Context) {
	methodID := c.Param("methodId")

	err := h.service.RemovePaymentMethod(c.Request.Context(), methodID)
	if err != nil {
		h.logger.Error("failed to remove payment method", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment method removed"})
}

// ListInvoices handles listing invoices
func (h *BillingHandler) ListInvoices(c *gin.Context) {
	orgID := c.Param("orgId")

	var filter billing.InvoiceFilter
	// Parse query parameters for filtering

	invoices, total, err := h.service.ListInvoices(c.Request.Context(), orgID, filter)
	if err != nil {
		h.logger.Error("failed to list invoices", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": invoices,
		"total":    total,
	})
}

// GetInvoice handles getting a specific invoice
func (h *BillingHandler) GetInvoice(c *gin.Context) {
	invoiceID := c.Param("invoiceId")

	invoice, err := h.service.GetInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		h.logger.Error("failed to get invoice", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// DownloadInvoice handles downloading invoice as PDF
func (h *BillingHandler) DownloadInvoice(c *gin.Context) {
	invoiceID := c.Param("invoiceId")

	pdfData, filename, err := h.service.DownloadInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		h.logger.Error("failed to download invoice", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// GetUpcomingInvoice handles getting upcoming invoice preview
func (h *BillingHandler) GetUpcomingInvoice(c *gin.Context) {
	orgID := c.Param("orgId")

	invoice, err := h.service.GetUpcomingInvoice(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get upcoming invoice", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// GetCurrentUsage handles getting current usage
func (h *BillingHandler) GetCurrentUsage(c *gin.Context) {
	orgID := c.Param("orgId")

	usage, err := h.service.GetCurrentUsage(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get current usage", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// GetBillingOverview handles getting billing overview
func (h *BillingHandler) GetBillingOverview(c *gin.Context) {
	orgID := c.Param("orgId")

	overview, err := h.service.GetBillingOverview(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get billing overview", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetBillingSettings handles getting billing settings
func (h *BillingHandler) GetBillingSettings(c *gin.Context) {
	orgID := c.Param("orgId")

	settings, err := h.service.GetBillingSettings(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get billing settings", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateBillingSettings handles updating billing settings
func (h *BillingHandler) UpdateBillingSettings(c *gin.Context) {
	orgID := c.Param("orgId")

	var settings billing.BillingSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := h.service.UpdateBillingSettings(c.Request.Context(), orgID, &settings)
	if err != nil {
		h.logger.Error("failed to update billing settings", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "billing settings updated"})
}

// HandleStripeWebhook handles Stripe webhook events
func (h *BillingHandler) HandleStripeWebhook(c *gin.Context) {
	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read webhook body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Get signature header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing signature"})
		return
	}

	// Process webhook
	err = h.service.ProcessStripeWebhook(c.Request.Context(), body, signature)
	if err != nil {
		h.logger.Error("failed to process webhook", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook processed"})
}