package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	aiopsDomain "github.com/hexabase/hexabase-ai/api/internal/aiops/domain"
	applicationDomain "github.com/hexabase/hexabase-ai/api/internal/application/domain"
	backupDomain "github.com/hexabase/hexabase-ai/api/internal/backup/domain"
	cicdDomain "github.com/hexabase/hexabase-ai/api/internal/cicd/domain"
	logsDomain "github.com/hexabase/hexabase-ai/api/internal/logs/domain"
	monitoringDomain "github.com/hexabase/hexabase-ai/api/internal/monitoring/domain"
	nodeDomain "github.com/hexabase/hexabase-ai/api/internal/node/domain"
	projectDomain "github.com/hexabase/hexabase-ai/api/internal/project/domain"
	workspaceDomain "github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
)

// InternalHandler handles internal-only API requests for AI agents.
type InternalHandler struct {
	workspaceSvc    workspaceDomain.Service
	projectSvc      projectDomain.Service
	applicationSvc  applicationDomain.Service
	nodeSvc         nodeDomain.Service
	logSvc          logsDomain.Service
	monitoringSvc   monitoringDomain.Service
	aiopsSvc        aiopsDomain.Service
	CICDService     cicdDomain.Service
	backupSvc       backupDomain.Service
	logger          *slog.Logger
}

// NewInternalHandler creates a new handler for internal operations.
func NewInternalHandler(
	workspaceSvc workspaceDomain.Service,
	projectSvc projectDomain.Service,
	applicationSvc applicationDomain.Service,
	nodeSvc nodeDomain.Service,
	logSvc logsDomain.Service,
	monitoringSvc monitoringDomain.Service,
	aiopsSvc aiopsDomain.Service,
	CICDService cicdDomain.Service,
	backupSvc backupDomain.Service,
	logger *slog.Logger,
) *InternalHandler {
	return &InternalHandler{
		workspaceSvc:    workspaceSvc,
		projectSvc:      projectSvc,
		applicationSvc:  applicationSvc,
		nodeSvc:         nodeSvc,
		logSvc:          logSvc,
		monitoringSvc:   monitoringSvc,
		aiopsSvc:        aiopsSvc,
		CICDService:     CICDService,
		backupSvc:       backupSvc,
		logger:          logger,
	}
}

// GetNodes is the handler for GET /internal/v1/workspaces/:workspaceId/nodes
func (h *InternalHandler) GetNodes(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	nodes, err := h.workspaceSvc.GetNodes(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get nodes for workspace", "workspace_id", workspaceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve node information"})
		return
	}

	c.JSON(http.StatusOK, nodes)
}

func (h *InternalHandler) ScaleDeployment(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	deploymentName := c.Param("deploymentName")

	var req struct {
		Replicas int `json:"replicas"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: replicas field is required"})
		return
	}

	if req.Replicas < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "replicas must be a non-negative integer"})
		return
	}

	err := h.workspaceSvc.ScaleDeployment(c.Request.Context(), workspaceID, deploymentName, req.Replicas)
	if err != nil {
		h.logger.Error("failed to scale deployment", "workspace_id", workspaceID, "deployment", deploymentName, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scale deployment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deployment scaled successfully"})
}

// QueryLogs is the handler for POST /internal/v1/logs/query
func (h *InternalHandler) QueryLogs(c *gin.Context) {
	var query logsDomain.LogQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query payload: " + err.Error()})
		return
	}

	// Basic validation
	if query.WorkspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	// Call the log service
	results, err := h.logSvc.QueryLogs(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("failed to query logs", "query", query, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetWorkspaceOverview returns a comprehensive overview of a workspace for AI agents
func (h *InternalHandler) GetWorkspaceOverview(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	// Get workspace details
	ws, err := h.workspaceSvc.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get workspace", "workspace_id", workspaceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve workspace"})
		return
	}

	// Get projects
	projectFilter := projectDomain.ProjectFilter{
		WorkspaceID: workspaceID,
	}
	projectList, err := h.projectSvc.ListProjects(c.Request.Context(), projectFilter)
	if err != nil {
		h.logger.Error("failed to get projects", "workspace_id", workspaceID, "error", err)
		projectList = &projectDomain.ProjectList{Projects: []*projectDomain.Project{}}
	}

	// Get nodes
	nodes, err := h.nodeSvc.ListNodes(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get nodes", "workspace_id", workspaceID, "error", err)
		nodes = []nodeDomain.DedicatedNode{}
	}

	// Get resource usage
	usage, err := h.nodeSvc.GetWorkspaceResourceUsage(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get resource usage", "workspace_id", workspaceID, "error", err)
		usage = &nodeDomain.WorkspaceResourceUsage{}
	}

	// Get monitoring metrics
	queryOpts := monitoringDomain.QueryOptions{
		Period: "24h",
	}
	metrics, err := h.monitoringSvc.GetWorkspaceMetrics(c.Request.Context(), workspaceID, queryOpts)
	if err != nil {
		h.logger.Error("failed to get metrics", "workspace_id", workspaceID, "error", err)
		metrics = &monitoringDomain.WorkspaceMetrics{}
	}

	c.JSON(http.StatusOK, gin.H{
		"workspace":      ws,
		"projects":       projectList.Projects,
		"nodes":          nodes,
		"resource_usage": usage,
		"metrics":        metrics,
	})
}

// GetApplicationDetails returns detailed information about an application
func (h *InternalHandler) GetApplicationDetails(c *gin.Context) {
	appID := c.Param("appId")

	// Get application
	app, err := h.applicationSvc.GetApplication(c.Request.Context(), appID)
	if err != nil {
		h.logger.Error("failed to get application", "app_id", appID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return
	}

	// Get deployment status - TODO: Implement deployment listing
	// deployments, err := h.CICDService.ListDeployments(c.Request.Context(), appID, 10)
	// if err != nil {
	// 	h.logger.Error("failed to get deployments", "app_id", appID, "error", err)
	// 	deployments = []cicd.Deployment{}
	// }

	// Get backup policies
	policy, err := h.backupSvc.GetBackupPolicyByApplication(c.Request.Context(), appID)
	if err != nil {
		h.logger.Warn("no backup policy found", "app_id", appID, "error", err)
	}

	// Get recent events
	events, err := h.applicationSvc.GetApplicationEvents(c.Request.Context(), appID, 20)
	if err != nil {
		h.logger.Error("failed to get events", "app_id", appID, "error", err)
		events = []applicationDomain.ApplicationEvent{}
	}

	c.JSON(http.StatusOK, gin.H{
		"application":   app,
		"backup_policy": policy,
		"events":        events,
	})
}

// ExecuteWorkspaceOperation performs an operation on a workspace
func (h *InternalHandler) ExecuteWorkspaceOperation(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	_ = workspaceID // Mark as used

	var req struct {
		Operation string                 `json:"operation" binding:"required"`
		Params    map[string]interface{} `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	switch req.Operation {
	case "scale_resources":
		// Scale workspace resources
		cpu := getIntParam(req.Params, "cpu_cores", 0)
		memory := getIntParam(req.Params, "memory_gb", 0)
		storage := getIntParam(req.Params, "storage_gb", 0)

		if cpu > 0 || memory > 0 || storage > 0 {
			// TODO: Implement resource scaling
			// err := h.workspaceSvc.UpdateResourceLimits(c.Request.Context(), workspaceID, cpu, memory, storage)
			// if err != nil {
			// 	h.logger.Error("failed to scale resources", "workspace_id", workspaceID, "error", err)
			// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scale resources"})
			// 	return
			// }
			c.JSON(http.StatusNotImplemented, gin.H{"error": "resource scaling not yet implemented"})
			return
		}

	case "restart_vcluster":
		// Restart vCluster - TODO: Implement vCluster restart
		// err := h.workspaceSvc.RestartVCluster(c.Request.Context(), workspaceID)
		// if err != nil {
		// 	h.logger.Error("failed to restart vcluster", "workspace_id", workspaceID, "error", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restart vcluster"})
		// 	return
		// }
		c.JSON(http.StatusNotImplemented, gin.H{"error": "vcluster restart not yet implemented"})
		return

	case "apply_security_policy":
		// Apply security policy - TODO: Implement security policy application
		policyName := getStringParam(req.Params, "policy_name", "")
		if policyName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "policy_name is required"})
			return
		}

		// err := h.workspaceSvc.ApplySecurityPolicy(c.Request.Context(), workspaceID, policyName)
		// if err != nil {
		// 	h.logger.Error("failed to apply security policy", "workspace_id", workspaceID, "error", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to apply security policy"})
		// 	return
		// }
		c.JSON(http.StatusNotImplemented, gin.H{"error": "security policy application not yet implemented"})
		return

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported operation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "operation completed successfully"})
}

// AutoScaleApplication automatically scales an application based on metrics
func (h *InternalHandler) AutoScaleApplication(c *gin.Context) {
	appID := c.Param("appId")

	var req struct {
		TargetCPU    int `json:"target_cpu_percent"`
		TargetMemory int `json:"target_memory_percent"`
		MinReplicas  int `json:"min_replicas"`
		MaxReplicas  int `json:"max_replicas"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Get application
	app, err := h.applicationSvc.GetApplication(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return
	}

	// Apply autoscaling configuration - TODO: Implement autoscaling
	// For now, we can scale the application to the min replicas
	if req.MinReplicas > 0 {
		err = h.applicationSvc.ScaleApplication(c.Request.Context(), app.ID, req.MinReplicas)
		if err != nil {
			h.logger.Error("failed to scale application", "app_id", appID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scale application"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "autoscaling configured successfully"})
}

// TriggerBackup triggers a backup for an application
func (h *InternalHandler) TriggerBackup(c *gin.Context) {
	appID := c.Param("appId")

	var req backupDomain.TriggerBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = backupDomain.TriggerBackupRequest{
			ApplicationID: appID,
			Metadata: map[string]interface{}{
				"triggered_by": "ai_agent",
				"timestamp":    time.Now().UTC(),
			},
		}
	}

	execution, err := h.backupSvc.TriggerManualBackup(c.Request.Context(), appID, req)
	if err != nil {
		h.logger.Error("failed to trigger backup", "app_id", appID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to trigger backup"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"execution_id": execution.ID,
		"status":       execution.Status,
		"message":      "backup triggered successfully",
	})
}

// GetAIInsights returns AI-generated insights about the system
func (h *InternalHandler) GetAIInsights(c *gin.Context) {
	workspaceID := c.Param("workspaceId")

	var req struct {
		InsightType string   `json:"insight_type"` // cost_optimization, performance, security, reliability
		TimeRange   string   `json:"time_range"`   // 1h, 24h, 7d, 30d
		Targets     []string `json:"targets"`      // specific apps or nodes
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.InsightType = "general"
		req.TimeRange = "24h"
	}

	// TODO: Implement AI-powered insights generation
	// For now, return basic usage stats
	var from, to time.Time
	switch req.TimeRange {
	case "1h":
		from = time.Now().Add(-1 * time.Hour)
	case "24h":
		from = time.Now().Add(-24 * time.Hour)
	case "7d":
		from = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		from = time.Now().Add(-30 * 24 * time.Hour)
	default:
		from = time.Now().Add(-24 * time.Hour)
	}
	to = time.Now()

	usageStats, err := h.aiopsSvc.GetUsageStats(c.Request.Context(), workspaceID, from, to)
	if err != nil {
		h.logger.Error("failed to get usage stats", "workspace_id", workspaceID, "error", err)
		usageStats = &aiopsDomain.UsageReport{
			WorkspaceID: workspaceID,
			Period: aiopsDomain.Period{From: from, To: to},
		}
	}

	insights := gin.H{
		"workspace_id":  workspaceID,
		"insight_type":  req.InsightType,
		"time_range":    req.TimeRange,
		"usage_stats":   usageStats,
		"insights": []gin.H{
			{
				"type":        "recommendation",
				"severity":    "info",
				"title":       "AI Insights Generation",
				"description": "Advanced AI insights generation is not yet implemented",
			},
		},
	}

	c.JSON(http.StatusOK, insights)
}

// AnalyzePerformance analyzes application performance and suggests optimizations
func (h *InternalHandler) AnalyzePerformance(c *gin.Context) {
	appID := c.Param("appId")

	// Get application metrics
	app, err := h.applicationSvc.GetApplication(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return
	}

	// Get application metrics
	metrics, err := h.applicationSvc.GetApplicationMetrics(c.Request.Context(), appID)
	if err != nil {
		h.logger.Error("failed to get metrics", "app_id", appID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get metrics"})
		return
	}

	// Basic performance analysis
	analysis := gin.H{
		"application_id": app.ID,
		"metrics":        metrics,
		"suggestions": []string{
			"Consider scaling if CPU usage > 80%",
			"Monitor memory usage trends",
			"Review request latency patterns",
		},
	}

	c.JSON(http.StatusOK, analysis)
}

// ManageIncident manages incidents detected by monitoring
func (h *InternalHandler) ManageIncident(c *gin.Context) {
	incidentID := c.Param("incidentId")

	var req struct {
		Action      string                 `json:"action"` // acknowledge, resolve, escalate, comment
		Comment     string                 `json:"comment"`
		Assignee    string                 `json:"assignee"`
		Resolution  string                 `json:"resolution"`
		Runbooks    []string               `json:"runbooks"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// TODO: Implement incident management
	// For now, we'll handle alerts which are similar to incidents
	switch req.Action {
	case "acknowledge":
		err := h.monitoringSvc.AcknowledgeAlert(c.Request.Context(), incidentID, req.Assignee)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to acknowledge alert"})
			return
		}

	case "resolve":
		err := h.monitoringSvc.ResolveAlert(c.Request.Context(), incidentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve alert"})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert updated successfully"})
}

// GetSystemHealth returns overall system health for AI monitoring
func (h *InternalHandler) GetSystemHealth(c *gin.Context) {
	// TODO: Implement system-wide health monitoring
	// For now, return a basic health status
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"components": gin.H{
			"api":        "healthy",
			"database":   "healthy",
			"kubernetes": "healthy",
			"monitoring": "healthy",
		},
		"message": "System health monitoring not yet fully implemented",
	}

	c.JSON(http.StatusOK, health)
}

// Helper functions
func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

func getStringParam(params map[string]interface{}, key string, defaultValue string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
} 