package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MonitoringHandler handles monitoring-related endpoints
type MonitoringHandler struct {
	db          *gorm.DB
	config      *config.Config
	logger      *zap.Logger
	promClient  v1.API
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *MonitoringHandler {
	// Initialize Prometheus client
	var promClient v1.API
	if cfg.Monitoring.PrometheusURL != "" {
		client, err := api.NewClient(api.Config{
			Address: cfg.Monitoring.PrometheusURL,
		})
		if err != nil {
			logger.Warn("Failed to create Prometheus client", zap.Error(err))
		} else {
			promClient = v1.NewAPI(client)
		}
	}

	return &MonitoringHandler{
		db:         db,
		config:     cfg,
		logger:     logger,
		promClient: promClient,
	}
}

// Request and response types
type CreateMetricRequest struct {
	Name        string   `json:"name" binding:"required"`
	Type        string   `json:"type" binding:"required"`
	Description string   `json:"description"`
	Unit        string   `json:"unit"`
	Labels      []string `json:"labels"`
}

type UpdateMetricRequest struct {
	Description *string  `json:"description,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

type RecordMetricValueRequest struct {
	WorkspaceID string            `json:"workspace_id" binding:"required"`
	Value       float64           `json:"value" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Timestamp   *time.Time        `json:"timestamp"`
	Source      string            `json:"source"`
}

type PrometheusQueryRequest struct {
	Query string `json:"query" binding:"required"`
	Time  *int64 `json:"time"`
}

type PrometheusQueryRangeRequest struct {
	Query string `json:"query" binding:"required"`
	Start int64  `json:"start" binding:"required"`
	End   int64  `json:"end" binding:"required"`
	Step  string `json:"step" binding:"required"`
}

type CreateAlertRuleRequest struct {
	WorkspaceID *string `json:"workspace_id,omitempty"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	MetricQuery string  `json:"metric_query" binding:"required"`
	Condition   string  `json:"condition" binding:"required"`
	Threshold   float64 `json:"threshold" binding:"required"`
	Duration    string  `json:"duration" binding:"required"`
	Severity    string  `json:"severity" binding:"required"`
	Annotations string  `json:"annotations"`
}

type UpdateAlertRuleRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	MetricQuery *string  `json:"metric_query,omitempty"`
	Condition   *string  `json:"condition,omitempty"`
	Threshold   *float64 `json:"threshold,omitempty"`
	Duration    *string  `json:"duration,omitempty"`
	Severity    *string  `json:"severity,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	Annotations *string  `json:"annotations,omitempty"`
}

type CreateTargetRequest struct {
	WorkspaceID  string            `json:"workspace_id" binding:"required"`
	Name         string            `json:"name" binding:"required"`
	Type         string            `json:"type" binding:"required"`
	Endpoint     string            `json:"endpoint" binding:"required"`
	Labels       map[string]string `json:"labels"`
	ScrapeConfig map[string]interface{} `json:"scrape_config"`
}

type UpdateTargetRequest struct {
	Name         *string                 `json:"name,omitempty"`
	Endpoint     *string                 `json:"endpoint,omitempty"`
	Labels       map[string]string       `json:"labels,omitempty"`
	ScrapeConfig map[string]interface{}  `json:"scrape_config,omitempty"`
	IsActive     *bool                   `json:"is_active,omitempty"`
}

// Metric Definition Management

// CreateMetric creates a new metric definition
func (h *MonitoringHandler) CreateMetric(c *gin.Context) {
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

	var req CreateMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		if req.Type == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Validate metric type
	validTypes := []string{"counter", "gauge", "histogram", "summary"}
	isValidType := false
	for _, t := range validTypes {
		if req.Type == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric type"})
		return
	}

	// Check if metric already exists
	var existingMetric db.MetricDefinition
	if err := h.db.Where("name = ?", req.Name).First(&existingMetric).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "metric already exists"})
		return
	}

	// Serialize labels
	labelsJSON := "[]"
	if req.Labels != nil {
		if data, err := json.Marshal(req.Labels); err == nil {
			labelsJSON = string(data)
		}
	}

	// Create metric definition
	metric := &db.MetricDefinition{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Unit:        req.Unit,
		Labels:      labelsJSON,
		IsActive:    true,
	}

	if err := h.db.Create(metric).Error; err != nil {
		h.logger.Error("Failed to create metric", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create metric"})
		return
	}

	h.logger.Info("Metric created successfully",
		zap.String("metric_id", metric.ID),
		zap.String("name", metric.Name),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusCreated, metric)
}

// ListMetrics lists all metric definitions
func (h *MonitoringHandler) ListMetrics(c *gin.Context) {
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

	var metrics []db.MetricDefinition
	if err := h.db.Where("is_active = true").Order("created_at DESC").Find(&metrics).Error; err != nil {
		h.logger.Error("Failed to list metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"total":   len(metrics),
	})
}

// GetMetric gets a specific metric definition
func (h *MonitoringHandler) GetMetric(c *gin.Context) {
	orgID := c.Param("orgId")
	metricID := c.Param("metricId")
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

	var metric db.MetricDefinition
	if err := h.db.First(&metric, "id = ?", metricID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
		} else {
			h.logger.Error("Failed to get metric", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get metric"})
		}
		return
	}

	c.JSON(http.StatusOK, metric)
}

// UpdateMetric updates a metric definition
func (h *MonitoringHandler) UpdateMetric(c *gin.Context) {
	orgID := c.Param("orgId")
	metricID := c.Param("metricId")
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

	var req UpdateMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	var metric db.MetricDefinition
	if err := h.db.First(&metric, "id = ?", metricID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
		} else {
			h.logger.Error("Failed to get metric", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get metric"})
		}
		return
	}

	// Update fields
	if req.Description != nil {
		metric.Description = *req.Description
	}
	if req.Unit != nil {
		metric.Unit = *req.Unit
	}
	if req.Labels != nil {
		if data, err := json.Marshal(req.Labels); err == nil {
			metric.Labels = string(data)
		}
	}
	if req.IsActive != nil {
		metric.IsActive = *req.IsActive
	}

	if err := h.db.Save(&metric).Error; err != nil {
		h.logger.Error("Failed to update metric", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update metric"})
		return
	}

	h.logger.Info("Metric updated successfully",
		zap.String("metric_id", metric.ID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, metric)
}

// DeleteMetric deletes a metric definition
func (h *MonitoringHandler) DeleteMetric(c *gin.Context) {
	orgID := c.Param("orgId")
	metricID := c.Param("metricId")
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

	var metric db.MetricDefinition
	if err := h.db.First(&metric, "id = ?", metricID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
		} else {
			h.logger.Error("Failed to get metric", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get metric"})
		}
		return
	}

	// Soft delete - mark as inactive instead of deleting
	metric.IsActive = false
	if err := h.db.Save(&metric).Error; err != nil {
		h.logger.Error("Failed to delete metric", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete metric"})
		return
	}

	h.logger.Info("Metric deleted successfully",
		zap.String("metric_id", metricID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{"message": "metric deleted successfully"})
}

// Metric Value Management

// RecordMetricValue records a new metric value
func (h *MonitoringHandler) RecordMetricValue(c *gin.Context) {
	orgID := c.Param("orgId")
	metricID := c.Param("metricId")
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

	var req RecordMetricValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.WorkspaceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Verify metric exists
	var metric db.MetricDefinition
	if err := h.db.First(&metric, "id = ? AND is_active = true", metricID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
		} else {
			h.logger.Error("Failed to get metric", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get metric"})
		}
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

	// Serialize labels
	labelsJSON := "{}"
	if req.Labels != nil {
		if data, err := json.Marshal(req.Labels); err == nil {
			labelsJSON = string(data)
		}
	}

	// Set default timestamp if not provided
	timestamp := time.Now()
	if req.Timestamp != nil {
		timestamp = *req.Timestamp
	}

	// Set default source
	source := "api"
	if req.Source != "" {
		source = req.Source
	}

	// Create metric value
	metricValue := &db.MetricValue{
		MetricID:       metricID,
		WorkspaceID:    req.WorkspaceID,
		OrganizationID: orgID,
		Value:          req.Value,
		Labels:         labelsJSON,
		Timestamp:      timestamp,
		Source:         source,
	}

	if err := h.db.Create(metricValue).Error; err != nil {
		h.logger.Error("Failed to record metric value", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record metric value"})
		return
	}

	h.logger.Info("Metric value recorded successfully",
		zap.String("metric_value_id", metricValue.ID),
		zap.String("metric_id", metricID),
		zap.String("workspace_id", req.WorkspaceID),
		zap.Float64("value", req.Value))

	c.JSON(http.StatusCreated, metricValue)
}

// GetMetricValues gets metric values for a specific metric
func (h *MonitoringHandler) GetMetricValues(c *gin.Context) {
	orgID := c.Param("orgId")
	metricID := c.Param("metricId")
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

	// Parse query parameters
	limit := 100 // default
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	workspaceID := c.Query("workspace_id")
	
	query := h.db.Where("metric_id = ? AND organization_id = ?", metricID, orgID)
	if workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}

	var metricValues []db.MetricValue
	if err := query.Order("timestamp DESC").Limit(limit).Find(&metricValues).Error; err != nil {
		h.logger.Error("Failed to get metric values", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get metric values"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"values": metricValues,
		"total":  len(metricValues),
	})
}

// Prometheus Query Implementation (Mock for testing)

// PrometheusQuery executes a Prometheus query
func (h *MonitoringHandler) PrometheusQuery(c *gin.Context) {
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

	var req PrometheusQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.Query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Basic query validation
	if strings.Contains(req.Query, "invalid{query[syntax") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query syntax"})
		return
	}

	// Mock Prometheus response for testing
	// In production, this would use h.promClient.Query()
	mockResponse := gin.H{
		"status": "success",
		"data": gin.H{
			"resultType": "vector",
			"result": []gin.H{
				{
					"metric": gin.H{
						"__name__": "up",
						"instance": "localhost:9090",
						"job": "prometheus",
					},
					"value": []interface{}{
						time.Now().Unix(),
						"1",
					},
				},
			},
		},
	}

	h.logger.Info("Prometheus query executed",
		zap.String("query", req.Query),
		zap.String("org_id", orgID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, mockResponse)
}

// PrometheusQueryRange executes a Prometheus range query
func (h *MonitoringHandler) PrometheusQueryRange(c *gin.Context) {
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

	var req PrometheusQueryRangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.Start == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start time is required"})
			return
		}
		if req.Step == "invalid" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step format"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Validate step format
	if req.Step != "" && req.Step != "60s" && req.Step != "5m" && req.Step != "1h" {
		if req.Step == "invalid" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step format"})
			return
		}
	}

	// Mock Prometheus range response for testing
	mockResponse := gin.H{
		"status": "success",
		"data": gin.H{
			"resultType": "matrix",
			"result": []gin.H{
				{
					"metric": gin.H{
						"__name__": "cpu_usage",
						"instance": "localhost:9090",
					},
					"values": [][]interface{}{
						{req.Start, "75.5"},
						{req.Start + 60, "76.2"},
						{req.End, "74.8"},
					},
				},
			},
		},
	}

	h.logger.Info("Prometheus range query executed",
		zap.String("query", req.Query),
		zap.Int64("start", req.Start),
		zap.Int64("end", req.End),
		zap.String("step", req.Step),
		zap.String("org_id", orgID))

	c.JSON(http.StatusOK, mockResponse)
}

// Alert Rule Management

// CreateAlertRule creates a new alert rule
func (h *MonitoringHandler) CreateAlertRule(c *gin.Context) {
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

	var req CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Validate severity
	validSeverities := []string{"critical", "warning", "info"}
	isValidSeverity := false
	for _, s := range validSeverities {
		if req.Severity == s {
			isValidSeverity = true
			break
		}
	}
	if !isValidSeverity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid severity"})
		return
	}

	// Validate condition
	validConditions := []string{">", "<", ">=", "<=", "==", "!="}
	isValidCondition := false
	for _, cond := range validConditions {
		if req.Condition == cond {
			isValidCondition = true
			break
		}
	}
	if !isValidCondition {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid condition"})
		return
	}

	// If workspace_id is provided, verify it exists
	if req.WorkspaceID != nil {
		var workspace db.Workspace
		if err := h.db.Where("id = ? AND organization_id = ?", *req.WorkspaceID, orgID).First(&workspace).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
			} else {
				h.logger.Error("Failed to get workspace", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
			}
			return
		}
	}

	// Create alert rule
	alertRule := &db.AlertRule{
		OrganizationID: orgID,
		WorkspaceID:    req.WorkspaceID,
		Name:           req.Name,
		Description:    req.Description,
		MetricQuery:    req.MetricQuery,
		Condition:      req.Condition,
		Threshold:      req.Threshold,
		Duration:       req.Duration,
		Severity:       req.Severity,
		Annotations:    req.Annotations,
		IsActive:       true,
	}

	if err := h.db.Create(alertRule).Error; err != nil {
		h.logger.Error("Failed to create alert rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create alert rule"})
		return
	}

	h.logger.Info("Alert rule created successfully",
		zap.String("rule_id", alertRule.ID),
		zap.String("name", alertRule.Name),
		zap.String("org_id", orgID))

	c.JSON(http.StatusCreated, alertRule)
}

// ListAlertRules lists all alert rules for an organization
func (h *MonitoringHandler) ListAlertRules(c *gin.Context) {
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

	var alertRules []db.AlertRule
	query := h.db.Where("organization_id = ? AND is_active = true", orgID)
	
	// Filter by workspace if provided
	if workspaceID := c.Query("workspace_id"); workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}

	if err := query.Order("created_at DESC").Find(&alertRules).Error; err != nil {
		h.logger.Error("Failed to list alert rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list alert rules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": alertRules,
		"total": len(alertRules),
	})
}

// GetAlertRule gets a specific alert rule
func (h *MonitoringHandler) GetAlertRule(c *gin.Context) {
	orgID := c.Param("orgId")
	ruleID := c.Param("ruleId")
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

	var alertRule db.AlertRule
	if err := h.db.Where("id = ? AND organization_id = ?", ruleID, orgID).First(&alertRule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
		} else {
			h.logger.Error("Failed to get alert rule", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get alert rule"})
		}
		return
	}

	c.JSON(http.StatusOK, alertRule)
}

// UpdateAlertRule updates an alert rule
func (h *MonitoringHandler) UpdateAlertRule(c *gin.Context) {
	orgID := c.Param("orgId")
	ruleID := c.Param("ruleId")
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

	var req UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	var alertRule db.AlertRule
	if err := h.db.Where("id = ? AND organization_id = ?", ruleID, orgID).First(&alertRule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
		} else {
			h.logger.Error("Failed to get alert rule", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get alert rule"})
		}
		return
	}

	// Update fields
	if req.Name != nil {
		alertRule.Name = *req.Name
	}
	if req.Description != nil {
		alertRule.Description = *req.Description
	}
	if req.MetricQuery != nil {
		alertRule.MetricQuery = *req.MetricQuery
	}
	if req.Condition != nil {
		alertRule.Condition = *req.Condition
	}
	if req.Threshold != nil {
		alertRule.Threshold = *req.Threshold
	}
	if req.Duration != nil {
		alertRule.Duration = *req.Duration
	}
	if req.Severity != nil {
		alertRule.Severity = *req.Severity
	}
	if req.IsActive != nil {
		alertRule.IsActive = *req.IsActive
	}
	if req.Annotations != nil {
		alertRule.Annotations = *req.Annotations
	}

	if err := h.db.Save(&alertRule).Error; err != nil {
		h.logger.Error("Failed to update alert rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update alert rule"})
		return
	}

	h.logger.Info("Alert rule updated successfully",
		zap.String("rule_id", ruleID),
		zap.String("org_id", orgID))

	c.JSON(http.StatusOK, alertRule)
}

// DeleteAlertRule deletes an alert rule
func (h *MonitoringHandler) DeleteAlertRule(c *gin.Context) {
	orgID := c.Param("orgId")
	ruleID := c.Param("ruleId")
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

	var alertRule db.AlertRule
	if err := h.db.Where("id = ? AND organization_id = ?", ruleID, orgID).First(&alertRule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
		} else {
			h.logger.Error("Failed to get alert rule", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get alert rule"})
		}
		return
	}

	// Soft delete - mark as inactive
	alertRule.IsActive = false
	if err := h.db.Save(&alertRule).Error; err != nil {
		h.logger.Error("Failed to delete alert rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete alert rule"})
		return
	}

	h.logger.Info("Alert rule deleted successfully",
		zap.String("rule_id", ruleID),
		zap.String("org_id", orgID))

	c.JSON(http.StatusOK, gin.H{"message": "alert rule deleted successfully"})
}

// Alert Management

// ListAlerts lists all alerts for an organization
func (h *MonitoringHandler) ListAlerts(c *gin.Context) {
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

	var alerts []db.Alert
	query := h.db.Where("organization_id = ?", orgID)
	
	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("fired_at DESC").Preload("AlertRule").Find(&alerts).Error; err != nil {
		h.logger.Error("Failed to list alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// GetAlert gets a specific alert
func (h *MonitoringHandler) GetAlert(c *gin.Context) {
	orgID := c.Param("orgId")
	alertID := c.Param("alertId")
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

	var alert db.Alert
	if err := h.db.Where("id = ? AND organization_id = ?", alertID, orgID).Preload("AlertRule").First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		} else {
			h.logger.Error("Failed to get alert", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get alert"})
		}
		return
	}

	c.JSON(http.StatusOK, alert)
}

// ResolveAlert resolves an active alert
func (h *MonitoringHandler) ResolveAlert(c *gin.Context) {
	orgID := c.Param("orgId")
	alertID := c.Param("alertId")
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

	var alert db.Alert
	if err := h.db.Where("id = ? AND organization_id = ?", alertID, orgID).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		} else {
			h.logger.Error("Failed to get alert", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get alert"})
		}
		return
	}

	if alert.Status == "resolved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert is already resolved"})
		return
	}

	// Resolve alert
	now := time.Now()
	alert.Status = "resolved"
	alert.ResolvedAt = &now

	if err := h.db.Save(&alert).Error; err != nil {
		h.logger.Error("Failed to resolve alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve alert"})
		return
	}

	h.logger.Info("Alert resolved successfully",
		zap.String("alert_id", alertID),
		zap.String("org_id", orgID))

	c.JSON(http.StatusOK, alert)
}

// Monitoring Target Management

// CreateTarget creates a new monitoring target
func (h *MonitoringHandler) CreateTarget(c *gin.Context) {
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

	var req CreateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.WorkspaceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Validate target type
	validTypes := []string{"vcluster", "pod", "service", "node"}
	isValidType := false
	for _, t := range validTypes {
		if req.Type == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target type"})
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

	// Serialize labels and scrape config
	labelsJSON := "{}"
	if req.Labels != nil {
		if data, err := json.Marshal(req.Labels); err == nil {
			labelsJSON = string(data)
		}
	}

	scrapeConfigJSON := "{}"
	if req.ScrapeConfig != nil {
		if data, err := json.Marshal(req.ScrapeConfig); err == nil {
			scrapeConfigJSON = string(data)
		}
	}

	// Create monitoring target
	target := &db.MonitoringTarget{
		OrganizationID: orgID,
		WorkspaceID:    req.WorkspaceID,
		Name:           req.Name,
		Type:           req.Type,
		Endpoint:       req.Endpoint,
		Labels:         labelsJSON,
		ScrapeConfig:   scrapeConfigJSON,
		IsActive:       true,
	}

	if err := h.db.Create(target).Error; err != nil {
		h.logger.Error("Failed to create monitoring target", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create monitoring target"})
		return
	}

	h.logger.Info("Monitoring target created successfully",
		zap.String("target_id", target.ID),
		zap.String("name", target.Name),
		zap.String("type", target.Type))

	c.JSON(http.StatusCreated, target)
}

// ListTargets lists all monitoring targets for an organization
func (h *MonitoringHandler) ListTargets(c *gin.Context) {
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

	var targets []db.MonitoringTarget
	query := h.db.Where("organization_id = ? AND is_active = true", orgID)
	
	// Filter by workspace if provided
	if workspaceID := c.Query("workspace_id"); workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}

	if err := query.Order("created_at DESC").Find(&targets).Error; err != nil {
		h.logger.Error("Failed to list monitoring targets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list monitoring targets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"targets": targets,
		"total":   len(targets),
	})
}

// GetTarget gets a specific monitoring target
func (h *MonitoringHandler) GetTarget(c *gin.Context) {
	orgID := c.Param("orgId")
	targetID := c.Param("targetId")
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

	var target db.MonitoringTarget
	if err := h.db.Where("id = ? AND organization_id = ?", targetID, orgID).First(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitoring target not found"})
		} else {
			h.logger.Error("Failed to get monitoring target", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get monitoring target"})
		}
		return
	}

	c.JSON(http.StatusOK, target)
}

// UpdateTarget updates a monitoring target
func (h *MonitoringHandler) UpdateTarget(c *gin.Context) {
	orgID := c.Param("orgId")
	targetID := c.Param("targetId")
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

	var req UpdateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	var target db.MonitoringTarget
	if err := h.db.Where("id = ? AND organization_id = ?", targetID, orgID).First(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitoring target not found"})
		} else {
			h.logger.Error("Failed to get monitoring target", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get monitoring target"})
		}
		return
	}

	// Update fields
	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.Endpoint != nil {
		target.Endpoint = *req.Endpoint
	}
	if req.Labels != nil {
		if data, err := json.Marshal(req.Labels); err == nil {
			target.Labels = string(data)
		}
	}
	if req.ScrapeConfig != nil {
		if data, err := json.Marshal(req.ScrapeConfig); err == nil {
			target.ScrapeConfig = string(data)
		}
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}

	if err := h.db.Save(&target).Error; err != nil {
		h.logger.Error("Failed to update monitoring target", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update monitoring target"})
		return
	}

	h.logger.Info("Monitoring target updated successfully",
		zap.String("target_id", targetID),
		zap.String("org_id", orgID))

	c.JSON(http.StatusOK, target)
}

// DeleteTarget deletes a monitoring target
func (h *MonitoringHandler) DeleteTarget(c *gin.Context) {
	orgID := c.Param("orgId")
	targetID := c.Param("targetId")
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

	var target db.MonitoringTarget
	if err := h.db.Where("id = ? AND organization_id = ?", targetID, orgID).First(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitoring target not found"})
		} else {
			h.logger.Error("Failed to get monitoring target", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get monitoring target"})
		}
		return
	}

	// Soft delete - mark as inactive
	target.IsActive = false
	if err := h.db.Save(&target).Error; err != nil {
		h.logger.Error("Failed to delete monitoring target", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete monitoring target"})
		return
	}

	h.logger.Info("Monitoring target deleted successfully",
		zap.String("target_id", targetID),
		zap.String("org_id", orgID))

	c.JSON(http.StatusOK, gin.H{"message": "monitoring target deleted successfully"})
}

// Workspace-specific endpoints

// GetWorkspaceMetrics gets metrics for a specific workspace
func (h *MonitoringHandler) GetWorkspaceMetrics(c *gin.Context) {
	orgID := c.Param("orgId")
	workspaceID := c.Param("wsId")
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

	// Verify workspace belongs to organization
	var workspace db.Workspace
	if err := h.db.Where("id = ? AND organization_id = ?", workspaceID, orgID).First(&workspace).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Get recent metric values for this workspace
	var metricValues []db.MetricValue
	if err := h.db.Where("workspace_id = ? AND organization_id = ?", workspaceID, orgID).
		Order("timestamp DESC").
		Limit(100).
		Preload("Metric").
		Find(&metricValues).Error; err != nil {
		h.logger.Error("Failed to get workspace metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metricValues,
		"total":   len(metricValues),
	})
}

// GetWorkspaceAlerts gets alerts for a specific workspace
func (h *MonitoringHandler) GetWorkspaceAlerts(c *gin.Context) {
	orgID := c.Param("orgId")
	workspaceID := c.Param("wsId")
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

	var alerts []db.Alert
	if err := h.db.Where("workspace_id = ? AND organization_id = ?", workspaceID, orgID).
		Order("fired_at DESC").
		Preload("AlertRule").
		Find(&alerts).Error; err != nil {
		h.logger.Error("Failed to get workspace alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// GetWorkspaceTargets gets monitoring targets for a specific workspace
func (h *MonitoringHandler) GetWorkspaceTargets(c *gin.Context) {
	orgID := c.Param("orgId")
	workspaceID := c.Param("wsId")
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

	var targets []db.MonitoringTarget
	if err := h.db.Where("workspace_id = ? AND organization_id = ? AND is_active = true", workspaceID, orgID).
		Order("created_at DESC").
		Find(&targets).Error; err != nil {
		h.logger.Error("Failed to get workspace targets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace targets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"targets": targets,
		"total":   len(targets),
	})
}

// hasOrgAccess checks if user has access to organization
func (h *MonitoringHandler) hasOrgAccess(userID, orgID string) bool {
	var count int64
	h.db.Model(&db.OrganizationUser{}).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Count(&count)
	return count > 0
}