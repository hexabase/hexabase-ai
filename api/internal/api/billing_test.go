package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// BillingTestSuite is the test suite for billing handlers
type BillingTestSuite struct {
	suite.Suite
	db         *gorm.DB
	handlers   *Handlers
	router     *gin.Engine
	authUser   *db.User
	authOrg    *db.Organization
	authToken  string
	testPlan   *db.Plan
}

// SetupSuite runs once before all tests
func (suite *BillingTestSuite) SetupSuite() {
	// Setup test database with SQLite in-memory
	var err error
	suite.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret:     "test-secret-key-for-testing-only",
			JWTExpiration: 3600,
			OIDCIssuer:    "https://test.example.com",
		},
		Stripe: config.StripeConfig{
			SecretKey:      "sk_test_123456789",
			WebhookSecret:  "whsec_test_123456789",
			PriceIDBasic:   "price_test_basic",
			PriceIDPro:     "price_test_pro",
			PriceIDEnterprise: "price_test_enterprise",
		},
	}

	// Auto-migrate all models
	err = suite.db.AutoMigrate(
		&db.User{},
		&db.Organization{},
		&db.OrganizationUser{},
		&db.Plan{},
		&db.Workspace{},
		&db.StripeEvent{},
		&db.Subscription{},
		&db.PaymentMethod{},
		&db.Invoice{},
		&db.UsageRecord{},
	)
	suite.Require().NoError(err)

	// Setup handlers and router
	logger, _ := zap.NewDevelopment()
	suite.handlers = NewHandlers(suite.db, &cfg, logger)
	
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Setup routes
	v1 := suite.router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(suite.testAuthMiddleware())
	
	// Billing routes
	orgs := protected.Group("/organizations")
	billing := orgs.Group("/:orgId/billing")
	{
		// Subscription management
		billing.GET("/subscription", suite.handlers.Billing.GetSubscription)
		billing.POST("/subscription", suite.handlers.Billing.CreateSubscription)
		billing.PUT("/subscription", suite.handlers.Billing.UpdateSubscription)
		billing.DELETE("/subscription", suite.handlers.Billing.CancelSubscription)
		
		// Payment methods
		billing.GET("/payment-methods", suite.handlers.Billing.ListPaymentMethods)
		billing.POST("/payment-methods", suite.handlers.Billing.AddPaymentMethod)
		billing.DELETE("/payment-methods/:pmId", suite.handlers.Billing.RemovePaymentMethod)
		billing.PUT("/payment-methods/:pmId/default", suite.handlers.Billing.SetDefaultPaymentMethod)
		
		// Invoices
		billing.GET("/invoices", suite.handlers.Billing.ListInvoices)
		billing.GET("/invoices/:invoiceId", suite.handlers.Billing.GetInvoice)
		billing.GET("/invoices/:invoiceId/download", suite.handlers.Billing.DownloadInvoice)
		
		// Usage and metering
		billing.GET("/usage", suite.handlers.Billing.GetUsage)
		billing.POST("/usage", suite.handlers.Billing.ReportUsage)
		
		// Billing portal
		billing.POST("/portal-session", suite.handlers.Billing.CreatePortalSession)
	}
	
	// Webhook route
	v1.POST("/webhooks/stripe", suite.handlers.Webhooks.HandleStripeWebhook)
}

// SetupTest runs before each test
func (suite *BillingTestSuite) SetupTest() {
	// Clean up database
	suite.db.Exec("DELETE FROM usage_records")
	suite.db.Exec("DELETE FROM invoices")
	suite.db.Exec("DELETE FROM payment_methods")
	suite.db.Exec("DELETE FROM subscriptions")
	suite.db.Exec("DELETE FROM stripe_events")
	suite.db.Exec("DELETE FROM workspaces")
	suite.db.Exec("DELETE FROM organization_users")
	suite.db.Exec("DELETE FROM organizations")
	suite.db.Exec("DELETE FROM users")
	suite.db.Exec("DELETE FROM plans")

	// Create test user
	suite.authUser = &db.User{
		ID:          "test-user-1",
		ExternalID:  "google-123456",
		Provider:    "google",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	suite.Require().NoError(suite.db.Create(suite.authUser).Error)

	// Create test organization
	suite.authOrg = &db.Organization{
		ID:               "test-org-1",
		Name:             "Test Organization",
		StripeCustomerID: stringPtr("cus_test123"),
	}
	suite.Require().NoError(suite.db.Create(suite.authOrg).Error)

	// Link user to organization as admin
	orgUser := &db.OrganizationUser{
		OrganizationID: suite.authOrg.ID,
		UserID:         suite.authUser.ID,
		Role:           "admin",
		JoinedAt:       time.Now(),
	}
	suite.Require().NoError(suite.db.Create(orgUser).Error)

	// Create test plans
	plans := []db.Plan{
		{
			ID:            "plan-basic",
			Name:          "Basic",
			Description:   "Basic plan with limited resources",
			Price:         9.99,
			Currency:      "USD",
			StripePriceID: "price_test_basic",
			ResourceLimits: `{
				"cpu": "4",
				"memory": "8Gi",
				"storage": "100Gi",
				"workspaces": "1"
			}`,
			MaxProjectsPerWorkspace: intPtr(5),
			MaxMembersPerWorkspace:  intPtr(10),
			IsActive:                true,
		},
		{
			ID:            "plan-pro",
			Name:          "Pro",
			Description:   "Professional plan with more resources",
			Price:         49.99,
			Currency:      "USD",
			StripePriceID: "price_test_pro",
			ResourceLimits: `{
				"cpu": "16",
				"memory": "32Gi",
				"storage": "500Gi",
				"workspaces": "5"
			}`,
			MaxProjectsPerWorkspace: intPtr(20),
			MaxMembersPerWorkspace:  intPtr(50),
			IsActive:                true,
		},
		{
			ID:            "plan-enterprise",
			Name:          "Enterprise",
			Description:   "Enterprise plan with unlimited resources",
			Price:         299.99,
			Currency:      "USD",
			StripePriceID: "price_test_enterprise",
			ResourceLimits: `{
				"cpu": "unlimited",
				"memory": "unlimited",
				"storage": "unlimited",
				"workspaces": "unlimited"
			}`,
			MaxProjectsPerWorkspace: intPtr(999),
			MaxMembersPerWorkspace:  intPtr(999),
			IsActive:                true,
		},
	}
	for _, plan := range plans {
		suite.Require().NoError(suite.db.Create(&plan).Error)
	}

	suite.testPlan = &plans[0] // Basic plan

	// Generate auth token
	suite.authToken = "Bearer test-token-" + suite.authUser.ID
}

// Test subscription management
func (suite *BillingTestSuite) TestCreateSubscription() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
		setup          func()
	}{
		{
			name: "successful subscription creation",
			payload: map[string]interface{}{
				"plan_id":           "plan-pro",
				"payment_method_id": "pm_test_123",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "organization already has active subscription",
			payload: map[string]interface{}{
				"plan_id":           "plan-pro",
				"payment_method_id": "pm_test_123",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "organization already has an active subscription",
			setup: func() {
				// First clean any existing subscriptions
				suite.db.Where("organization_id = ?", suite.authOrg.ID).Delete(&db.Subscription{})
				
				subscription := &db.Subscription{
					ID:                     "sub-existing",
					OrganizationID:         suite.authOrg.ID,
					PlanID:                 "plan-basic",
					StripeSubscriptionID:   "sub_test_existing",
					Status:                 "active",
					CurrentPeriodStart:     time.Now(),
					CurrentPeriodEnd:       time.Now().Add(30 * 24 * time.Hour),
				}
				suite.Require().NoError(suite.db.Create(subscription).Error)
			},
		},
		{
			name:           "missing plan_id",
			payload:        map[string]interface{}{
				"payment_method_id": "pm_test_123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "plan_id is required",
		},
		{
			name: "invalid plan_id",
			payload: map[string]interface{}{
				"plan_id":           "plan-invalid",
				"payment_method_id": "pm_test_123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid plan",
		},
		{
			name: "missing payment method",
			payload: map[string]interface{}{
				"plan_id": "plan-pro",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "payment_method_id is required",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Clean up any existing subscriptions before each test
			suite.db.Where("organization_id = ?", suite.authOrg.ID).Delete(&db.Subscription{})
			
			if tt.setup != nil {
				tt.setup()
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/billing/subscription", suite.authOrg.ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusCreated {
				var subscription db.Subscription
				err := json.Unmarshal(w.Body.Bytes(), &subscription)
				suite.NoError(err)
				suite.NotEmpty(subscription.ID)
				suite.Equal("plan-pro", subscription.PlanID)
			}
		})
	}
}

func (suite *BillingTestSuite) TestGetSubscription() {
	// Create test subscription
	subscription := &db.Subscription{
		ID:                     "sub-test",
		OrganizationID:         suite.authOrg.ID,
		PlanID:                 suite.testPlan.ID,
		StripeSubscriptionID:   "sub_test_123",
		Status:                 "active",
		CurrentPeriodStart:     time.Now(),
		CurrentPeriodEnd:       time.Now().Add(30 * 24 * time.Hour),
	}
	suite.Require().NoError(suite.db.Create(subscription).Error)

	tests := []struct {
		name           string
		orgID          string
		expectedStatus int
		expectedError  string
		setup          func()
	}{
		{
			name:           "successful get subscription",
			orgID:          suite.authOrg.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no active subscription",
			orgID:          suite.authOrg.ID,
			expectedStatus: http.StatusNotFound,
			expectedError:  "no active subscription found",
			setup: func() {
				// Delete the subscription
				suite.db.Delete(&subscription)
			},
		},
		{
			name:           "organization not found",
			orgID:          "org-nonexistent",
			expectedStatus: http.StatusForbidden,
			expectedError:  "not authorized",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/billing/subscription", tt.orgID), nil)
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			}
		})
	}
}

func (suite *BillingTestSuite) TestUpdateSubscription() {
	// Create test subscription
	subscription := &db.Subscription{
		ID:                     "sub-test",
		OrganizationID:         suite.authOrg.ID,
		PlanID:                 "plan-basic",
		StripeSubscriptionID:   "sub_test_123",
		Status:                 "active",
		CurrentPeriodStart:     time.Now(),
		CurrentPeriodEnd:       time.Now().Add(30 * 24 * time.Hour),
	}
	suite.Require().NoError(suite.db.Create(subscription).Error)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful plan upgrade",
			payload: map[string]interface{}{
				"plan_id": "plan-pro",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful plan downgrade",
			payload: map[string]interface{}{
				"plan_id": "plan-basic",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid plan_id",
			payload: map[string]interface{}{
				"plan_id": "plan-invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid plan",
		},
		{
			name:           "missing plan_id",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "plan_id is required",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/organizations/%s/billing/subscription", suite.authOrg.ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			}
		})
	}
}

func (suite *BillingTestSuite) TestCancelSubscription() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful immediate cancellation",
			payload: map[string]interface{}{
				"immediate": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful end of period cancellation",
			payload: map[string]interface{}{
				"immediate": false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "default to end of period",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusOK,
		},
	}

	for i, tt := range tests {
		suite.Run(tt.name, func() {
			// Clean up any existing subscriptions first
			suite.db.Where("organization_id = ?", suite.authOrg.ID).Delete(&db.Subscription{})
			
			// Create fresh subscription for each test
			subscription := &db.Subscription{
				ID:                     fmt.Sprintf("sub-test-%d", i),
				OrganizationID:         suite.authOrg.ID,
				PlanID:                 suite.testPlan.ID,
				StripeSubscriptionID:   fmt.Sprintf("sub_test_%d", i),
				Status:                 "active",
				CurrentPeriodStart:     time.Now(),
				CurrentPeriodEnd:       time.Now().Add(30 * 24 * time.Hour),
			}
			suite.Require().NoError(suite.db.Create(subscription).Error)

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/organizations/%s/billing/subscription", suite.authOrg.ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			}
		})
	}
}

// Test payment methods
func (suite *BillingTestSuite) TestAddPaymentMethod() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful add payment method",
			payload: map[string]interface{}{
				"payment_method_id": "pm_test_new_1",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing payment_method_id",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "payment_method_id is required",
		},
		{
			name: "set as default",
			payload: map[string]interface{}{
				"payment_method_id": "pm_test_new_2",
				"set_default":       true,
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/billing/payment-methods", suite.authOrg.ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			}
		})
	}
}

func (suite *BillingTestSuite) TestListPaymentMethods() {
	// Create test payment methods
	paymentMethods := []db.PaymentMethod{
		{
			ID:                      "pm-1",
			OrganizationID:          suite.authOrg.ID,
			StripePaymentMethodID:   "pm_test_1",
			Type:                    "card",
			Card: &db.CardDetails{
				Brand:    "visa",
				Last4:    "4242",
				ExpMonth: 12,
				ExpYear:  2025,
			},
			IsDefault: true,
		},
		{
			ID:                      "pm-2",
			OrganizationID:          suite.authOrg.ID,
			StripePaymentMethodID:   "pm_test_2",
			Type:                    "card",
			Card: &db.CardDetails{
				Brand:    "mastercard",
				Last4:    "5555",
				ExpMonth: 6,
				ExpYear:  2026,
			},
			IsDefault: false,
		},
	}
	for _, pm := range paymentMethods {
		suite.Require().NoError(suite.db.Create(&pm).Error)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/billing/payment-methods", suite.authOrg.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response struct {
		PaymentMethods []db.PaymentMethod `json:"payment_methods"`
		Total          int                `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(2, response.Total)
	suite.Len(response.PaymentMethods, 2)
}

// Test usage and metering
func (suite *BillingTestSuite) TestReportUsage() {
	// Create workspace for usage reporting
	workspace := &db.Workspace{
		ID:             "ws-test",
		OrganizationID: suite.authOrg.ID,
		Name:           "Test Workspace",
		PlanID:         suite.testPlan.ID,
		VClusterStatus: "RUNNING",
	}
	suite.Require().NoError(suite.db.Create(workspace).Error)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful usage report",
			payload: map[string]interface{}{
				"workspace_id": "ws-test",
				"metric_type":  "cpu_hours",
				"quantity":     10.5,
				"timestamp":    time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing workspace_id",
			payload: map[string]interface{}{
				"metric_type": "cpu_hours",
				"quantity":    10.5,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace_id is required",
		},
		{
			name: "invalid metric_type",
			payload: map[string]interface{}{
				"workspace_id": "ws-test",
				"metric_type":  "invalid_metric",
				"quantity":     10.5,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid metric_type",
		},
		{
			name: "workspace not found",
			payload: map[string]interface{}{
				"workspace_id": "ws-nonexistent",
				"metric_type":  "cpu_hours",
				"quantity":     10.5,
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "workspace not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/billing/usage", suite.authOrg.ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			}
		})
	}
}

// Test webhook handling
func (suite *BillingTestSuite) TestStripeWebhook() {
	tests := []struct {
		name           string
		event          interface{}
		signature      string
		expectedStatus int
		expectedError  string
		setup          func()
	}{
		{
			name: "successful subscription created event",
			event: map[string]interface{}{
				"id":      "evt_test_123",
				"type":    "customer.subscription.created",
				"created": time.Now().Unix(),
				"data": map[string]interface{}{
					"object": map[string]interface{}{
						"id":       "sub_test_123",
						"customer": *suite.authOrg.StripeCustomerID,
						"status":   "active",
						"items": map[string]interface{}{
							"data": []map[string]interface{}{
								{
									"price": map[string]interface{}{
										"id": "price_test_pro",
									},
								},
							},
						},
					},
				},
			},
			signature:      "valid_signature",
			expectedStatus: http.StatusOK,
		},
		{
			name: "invoice payment succeeded",
			event: map[string]interface{}{
				"id":      "evt_test_456",
				"type":    "invoice.payment_succeeded",
				"created": time.Now().Unix(),
				"data": map[string]interface{}{
					"object": map[string]interface{}{
						"id":                "in_test_123",
						"customer":          *suite.authOrg.StripeCustomerID,
						"amount_paid":       4999,
						"currency":          "usd",
						"billing_reason":    "subscription_cycle",
						"subscription":      "sub_test_123",
						"period_start":      time.Now().Unix(),
						"period_end":        time.Now().Add(30 * 24 * time.Hour).Unix(),
					},
				},
			},
			signature:      "valid_signature",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid signature",
			event:          map[string]interface{}{},
			signature:      "invalid_signature",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid signature",
		},
		{
			name: "duplicate event",
			event: map[string]interface{}{
				"id":   "evt_duplicate",
				"type": "customer.subscription.created",
			},
			signature:      "valid_signature",
			expectedStatus: http.StatusOK, // Should return OK even for duplicates
			setup: func() {
				// Create existing event
				event := &db.StripeEvent{
					EventID:     "evt_duplicate",
					EventType:   "customer.subscription.created",
					ProcessedAt: timePtr(time.Now()),
				}
				suite.db.Create(event)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			body, _ := json.Marshal(tt.event)
			req := httptest.NewRequest("POST", "/api/v1/webhooks/stripe", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Stripe-Signature", tt.signature)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			}
		})
	}
}

// testAuthMiddleware creates a mock auth middleware for testing
func (suite *BillingTestSuite) testAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Simple mock authentication - extract user ID from token
		if authHeader == "Bearer test-token-"+suite.authUser.ID {
			// Get user's organizations
			var orgMemberships []db.OrganizationUser
			suite.db.Where("user_id = ?", suite.authUser.ID).Find(&orgMemberships)
			
			var orgIDs []string
			for _, membership := range orgMemberships {
				orgIDs = append(orgIDs, membership.OrganizationID)
			}
			
			c.Set("user_id", suite.authUser.ID)
			c.Set("user_email", suite.authUser.Email)
			c.Set("user_name", suite.authUser.DisplayName)
			c.Set("org_ids", orgIDs)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function for time pointers (stringPtr and intPtr already exist in other test files)
func timePtr(t time.Time) *time.Time {
	return &t
}

// TestBillingSuite runs the test suite
func TestBillingSuite(t *testing.T) {
	suite.Run(t, new(BillingTestSuite))
}