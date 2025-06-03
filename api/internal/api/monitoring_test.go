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

// MonitoringTestSuite is the test suite for monitoring handlers
type MonitoringTestSuite struct {
	suite.Suite
	db          *gorm.DB
	handlers    *Handlers
	router      *gin.Engine
	authUser    *db.User
	authOrg     *db.Organization
	authToken   string
	testWorkspace *db.Workspace
}

// SetupSuite runs once before all tests
func (suite *MonitoringTestSuite) SetupSuite() {
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
		Monitoring: config.MonitoringConfig{
			PrometheusURL:   "http://localhost:9090",
			MetricsPort:     "2112",
			MetricsPath:     "/metrics",
			EnableMetrics:   true,
			EnableAlerts:    true,
			ScrapeInterval:  "15s",
			RetentionPeriod: "15d",
		},
	}

	// Auto-migrate all models
	err = suite.db.AutoMigrate(
		&db.User{},
		&db.Organization{},
		&db.OrganizationUser{},
		&db.Plan{},
		&db.Workspace{},
		&db.MetricDefinition{},
		&db.MetricValue{},
		&db.AlertRule{},
		&db.Alert{},
		&db.MonitoringTarget{},
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
	
	// Monitoring routes
	orgs := protected.Group("/organizations")
	monitoring := orgs.Group("/:orgId/monitoring")
	{
		// Metrics management
		monitoring.GET("/metrics", suite.handlers.Monitoring.ListMetrics)
		monitoring.POST("/metrics", suite.handlers.Monitoring.CreateMetric)
		monitoring.GET("/metrics/:metricId", suite.handlers.Monitoring.GetMetric)
		monitoring.PUT("/metrics/:metricId", suite.handlers.Monitoring.UpdateMetric)
		monitoring.DELETE("/metrics/:metricId", suite.handlers.Monitoring.DeleteMetric)
		
		// Metric values
		monitoring.GET("/metrics/:metricId/values", suite.handlers.Monitoring.GetMetricValues)
		monitoring.POST("/metrics/:metricId/values", suite.handlers.Monitoring.RecordMetricValue)
		
		// Prometheus queries
		monitoring.POST("/query", suite.handlers.Monitoring.PrometheusQuery)
		monitoring.POST("/query_range", suite.handlers.Monitoring.PrometheusQueryRange)
		
		// Alert rules
		monitoring.GET("/alerts/rules", suite.handlers.Monitoring.ListAlertRules)
		monitoring.POST("/alerts/rules", suite.handlers.Monitoring.CreateAlertRule)
		monitoring.GET("/alerts/rules/:ruleId", suite.handlers.Monitoring.GetAlertRule)
		monitoring.PUT("/alerts/rules/:ruleId", suite.handlers.Monitoring.UpdateAlertRule)
		monitoring.DELETE("/alerts/rules/:ruleId", suite.handlers.Monitoring.DeleteAlertRule)
		
		// Active alerts
		monitoring.GET("/alerts", suite.handlers.Monitoring.ListAlerts)
		monitoring.GET("/alerts/:alertId", suite.handlers.Monitoring.GetAlert)
		monitoring.POST("/alerts/:alertId/resolve", suite.handlers.Monitoring.ResolveAlert)
		
		// Monitoring targets
		monitoring.GET("/targets", suite.handlers.Monitoring.ListTargets)
		monitoring.POST("/targets", suite.handlers.Monitoring.CreateTarget)
		monitoring.GET("/targets/:targetId", suite.handlers.Monitoring.GetTarget)
		monitoring.PUT("/targets/:targetId", suite.handlers.Monitoring.UpdateTarget)
		monitoring.DELETE("/targets/:targetId", suite.handlers.Monitoring.DeleteTarget)
		
		// Workspace-specific monitoring
		workspaces := orgs.Group("/:orgId/workspaces/:wsId/monitoring")
		workspaces.GET("/metrics", suite.handlers.Monitoring.GetWorkspaceMetrics)
		workspaces.GET("/alerts", suite.handlers.Monitoring.GetWorkspaceAlerts)
		workspaces.GET("/targets", suite.handlers.Monitoring.GetWorkspaceTargets)
	}
}

// SetupTest runs before each test
func (suite *MonitoringTestSuite) SetupTest() {
	// Clean up database
	suite.db.Exec("DELETE FROM alerts")
	suite.db.Exec("DELETE FROM alert_rules")
	suite.db.Exec("DELETE FROM metric_values")
	suite.db.Exec("DELETE FROM metric_definitions")
	suite.db.Exec("DELETE FROM monitoring_targets")
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
		ID:   "test-org-1",
		Name: "Test Organization",
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

	// Create test plan
	plan := &db.Plan{
		ID:          "plan-test",
		Name:        "Test Plan",
		Description: "Test plan for monitoring",
		Price:       9.99,
		Currency:    "USD",
		StripePriceID: "price_test",
		IsActive:    true,
	}
	suite.Require().NoError(suite.db.Create(plan).Error)

	// Create test workspace
	suite.testWorkspace = &db.Workspace{
		ID:             "ws-test",
		OrganizationID: suite.authOrg.ID,
		Name:           "Test Workspace",
		PlanID:         "plan-test",
		VClusterStatus: "RUNNING",
	}
	suite.Require().NoError(suite.db.Create(suite.testWorkspace).Error)

	// Generate auth token
	suite.authToken = "Bearer test-token-" + suite.authUser.ID
}

// Test metric definition management
func (suite *MonitoringTestSuite) TestCreateMetric() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful metric creation",
			payload: map[string]interface{}{
				"name":        "cpu_usage",
				"type":        "gauge",
				"description": "CPU usage percentage",
				"unit":        "percent",
				"labels":      []string{"pod", "namespace"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid metric type",
			payload: map[string]interface{}{
				"name":        "invalid_metric",
				"type":        "invalid_type",
				"description": "Invalid metric type",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid metric type",
		},
		{
			name:           "missing name",
			payload:        map[string]interface{}{"type": "gauge"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "duplicate metric name",
			payload: map[string]interface{}{
				"name": "cpu_usage",
				"type": "gauge",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "metric already exists",
		},
	}

	for i, tt := range tests {
		suite.Run(tt.name, func() {
			// For duplicate test, create the metric first
			if i == 3 {
				metric := &db.MetricDefinition{
					Name: "cpu_usage",
					Type: "gauge",
				}
				suite.db.Create(metric)
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/monitoring/metrics", suite.authOrg.ID), bytes.NewBuffer(body))
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

func (suite *MonitoringTestSuite) TestListMetrics() {
	// Create test metrics
	metrics := []db.MetricDefinition{
		{Name: "cpu_usage", Type: "gauge", Description: "CPU usage"},
		{Name: "memory_usage", Type: "gauge", Description: "Memory usage"},
		{Name: "request_count", Type: "counter", Description: "Request count"},
	}
	for _, metric := range metrics {
		suite.Require().NoError(suite.db.Create(&metric).Error)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/monitoring/metrics", suite.authOrg.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response struct {
		Metrics []db.MetricDefinition `json:"metrics"`
		Total   int                   `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(3, response.Total)
	suite.Len(response.Metrics, 3)
}

// Test metric value recording
func (suite *MonitoringTestSuite) TestRecordMetricValue() {
	// Create test metric
	metric := &db.MetricDefinition{
		ID:   "metric-test",
		Name: "cpu_usage",
		Type: "gauge",
	}
	suite.Require().NoError(suite.db.Create(metric).Error)

	tests := []struct {
		name           string
		metricID       string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful value recording",
			metricID: "metric-test",
			payload: map[string]interface{}{
				"value":       75.5,
				"workspace_id": suite.testWorkspace.ID,
				"labels":      map[string]string{"pod": "test-pod"},
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "missing workspace_id",
			metricID: "metric-test",
			payload: map[string]interface{}{
				"value": 75.5,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace_id is required",
		},
		{
			name:     "metric not found",
			metricID: "metric-nonexistent",
			payload: map[string]interface{}{
				"value":       75.5,
				"workspace_id": suite.testWorkspace.ID,
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "metric not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/monitoring/metrics/%s/values", suite.authOrg.ID, tt.metricID), bytes.NewBuffer(body))
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

// Test Prometheus query endpoints
func (suite *MonitoringTestSuite) TestPrometheusQuery() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful prometheus query",
			payload: map[string]interface{}{
				"query": "up",
				"time":  time.Now().Unix(),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful query with evaluation time",
			payload: map[string]interface{}{
				"query": "cpu_usage{workspace=\"test\"}",
				"time":  time.Now().Unix(),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing query",
			payload:        map[string]interface{}{"time": time.Now().Unix()},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "query is required",
		},
		{
			name: "invalid query syntax",
			payload: map[string]interface{}{
				"query": "invalid{query[syntax",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid query syntax",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/monitoring/query", suite.authOrg.ID), bytes.NewBuffer(body))
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

func (suite *MonitoringTestSuite) TestPrometheusQueryRange() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful range query",
			payload: map[string]interface{}{
				"query": "cpu_usage",
				"start": time.Now().Add(-1 * time.Hour).Unix(),
				"end":   time.Now().Unix(),
				"step":  "60s",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing start time",
			payload: map[string]interface{}{
				"query": "cpu_usage",
				"end":   time.Now().Unix(),
				"step":  "60s",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "start time is required",
		},
		{
			name: "invalid step",
			payload: map[string]interface{}{
				"query": "cpu_usage",
				"start": time.Now().Add(-1 * time.Hour).Unix(),
				"end":   time.Now().Unix(),
				"step":  "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid step format",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/monitoring/query_range", suite.authOrg.ID), bytes.NewBuffer(body))
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

// Test alert rule management
func (suite *MonitoringTestSuite) TestCreateAlertRule() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful alert rule creation",
			payload: map[string]interface{}{
				"name":         "High CPU Usage",
				"description":  "Alert when CPU usage exceeds 80%",
				"metric_query": "cpu_usage > 80",
				"condition":    ">",
				"threshold":    80.0,
				"duration":     "5m",
				"severity":     "warning",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "workspace-specific alert rule",
			payload: map[string]interface{}{
				"name":         "Workspace CPU Alert",
				"workspace_id": suite.testWorkspace.ID,
				"metric_query": "cpu_usage{workspace=\"test\"} > 90",
				"condition":    ">",
				"threshold":    90.0,
				"duration":     "2m",
				"severity":     "critical",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid severity",
			payload: map[string]interface{}{
				"name":         "Invalid Severity",
				"metric_query": "cpu_usage > 80",
				"condition":    ">",
				"threshold":    80.0,
				"duration":     "5m",
				"severity":     "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid severity",
		},
		{
			name:           "missing name",
			payload:        map[string]interface{}{"metric_query": "cpu_usage > 80"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/monitoring/alerts/rules", suite.authOrg.ID), bytes.NewBuffer(body))
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

func (suite *MonitoringTestSuite) TestListAlerts() {
	// Create test alert rule
	rule := &db.AlertRule{
		ID:             "rule-test",
		OrganizationID: suite.authOrg.ID,
		Name:           "Test Rule",
		MetricQuery:    "cpu_usage > 80",
		Condition:      ">",
		Threshold:      80.0,
		Duration:       "5m",
		Severity:       "warning",
	}
	suite.Require().NoError(suite.db.Create(rule).Error)

	// Create test alerts
	alerts := []db.Alert{
		{
			ID:             "alert-1",
			AlertRuleID:    "rule-test",
			OrganizationID: suite.authOrg.ID,
			Status:         "firing",
			Value:          85.5,
			FiredAt:        time.Now(),
		},
		{
			ID:             "alert-2",
			AlertRuleID:    "rule-test",
			OrganizationID: suite.authOrg.ID,
			Status:         "resolved",
			Value:          82.0,
			FiredAt:        time.Now().Add(-1 * time.Hour),
			ResolvedAt:     &[]time.Time{time.Now().Add(-30 * time.Minute)}[0],
		},
	}
	for _, alert := range alerts {
		suite.Require().NoError(suite.db.Create(&alert).Error)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/monitoring/alerts", suite.authOrg.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response struct {
		Alerts []db.Alert `json:"alerts"`
		Total  int        `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(2, response.Total)
	suite.Len(response.Alerts, 2)
}

// Test monitoring targets
func (suite *MonitoringTestSuite) TestCreateTarget() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful target creation",
			payload: map[string]interface{}{
				"name":         "Test vCluster",
				"type":         "vcluster",
				"workspace_id": suite.testWorkspace.ID,
				"endpoint":     "http://test-vcluster:8080/metrics",
				"labels":       map[string]string{"cluster": "test"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "pod target creation",
			payload: map[string]interface{}{
				"name":         "Test Pod",
				"type":         "pod",
				"workspace_id": suite.testWorkspace.ID,
				"endpoint":     "http://test-pod:9090/metrics",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid target type",
			payload: map[string]interface{}{
				"name":         "Invalid Target",
				"type":         "invalid_type",
				"workspace_id": suite.testWorkspace.ID,
				"endpoint":     "http://test:8080/metrics",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid target type",
		},
		{
			name:           "missing workspace_id",
			payload:        map[string]interface{}{"name": "Test", "type": "pod"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace_id is required",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/monitoring/targets", suite.authOrg.ID), bytes.NewBuffer(body))
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

// Test workspace-specific endpoints
func (suite *MonitoringTestSuite) TestGetWorkspaceMetrics() {
	// Create test metric and values
	metric := &db.MetricDefinition{
		ID:   "metric-test",
		Name: "cpu_usage",
		Type: "gauge",
	}
	suite.Require().NoError(suite.db.Create(metric).Error)

	metricValue := &db.MetricValue{
		MetricID:       "metric-test",
		WorkspaceID:    suite.testWorkspace.ID,
		OrganizationID: suite.authOrg.ID,
		Value:          75.5,
		Timestamp:      time.Now(),
	}
	suite.Require().NoError(suite.db.Create(metricValue).Error)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/monitoring/metrics", suite.authOrg.ID, suite.testWorkspace.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response struct {
		Metrics []db.MetricValue `json:"metrics"`
		Total   int              `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(1, response.Total)
}

// testAuthMiddleware creates a mock auth middleware for testing
func (suite *MonitoringTestSuite) testAuthMiddleware() gin.HandlerFunc {
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

// TestMonitoringSuite runs the test suite
func TestMonitoringSuite(t *testing.T) {
	suite.Run(t, new(MonitoringTestSuite))
}