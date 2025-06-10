package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
)

// ApplicationHandler handles application-related requests
type ApplicationHandler struct {
	appService application.Service
}

// NewApplicationHandler creates a new ApplicationHandler
func NewApplicationHandler(appService application.Service) *ApplicationHandler {
	return &ApplicationHandler{
		appService: appService,
	}
}

// CreateApplication creates a new application
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req application.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Validate application type
	if !req.Type.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application type"})
		return
	}

	// Validate source type (only if not using template)
	if req.TemplateAppID == "" && !req.Source.Type.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source type"})
		return
	}

	// Get workspace ID from URL
	workspaceID := c.Param("wsId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace ID is required"})
		return
	}

	// Create the application
	app, err := h.appService.CreateApplication(c.Request.Context(), workspaceID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, app)
}

// GetApplication retrieves an application by ID
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	app, err := h.appService.GetApplication(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, app)
}

// ListApplications lists applications in a workspace
func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	workspaceID := c.Param("wsId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace ID is required"})
		return
	}

	projectID := c.Query("project_id")

	apps, err := h.appService.ListApplications(c.Request.Context(), workspaceID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apps)
}

// UpdateApplication updates an application
func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var req application.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	app, err := h.appService.UpdateApplication(c.Request.Context(), appID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, app)
}

// DeleteApplication deletes an application
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	err := h.appService.DeleteApplication(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// StartApplication starts an application
func (h *ApplicationHandler) StartApplication(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	err := h.appService.StartApplication(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application started"})
}

// StopApplication stops an application
func (h *ApplicationHandler) StopApplication(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	err := h.appService.StopApplication(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application stopped"})
}

// RestartApplication restarts an application
func (h *ApplicationHandler) RestartApplication(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	err := h.appService.RestartApplication(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application restarted"})
}

// ScaleApplication scales an application
func (h *ApplicationHandler) ScaleApplication(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var req struct {
		Replicas int `json:"replicas"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	err := h.appService.ScaleApplication(c.Request.Context(), appID, req.Replicas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application scaled"})
}

// ListPods lists pods for an application
func (h *ApplicationHandler) ListPods(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	pods, err := h.appService.ListPods(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pods)
}

// RestartPod restarts a specific pod
func (h *ApplicationHandler) RestartPod(c *gin.Context) {
	appID := c.Param("appId")
	podName := c.Param("podName")
	if appID == "" || podName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID and pod name are required"})
		return
	}

	err := h.appService.RestartPod(c.Request.Context(), appID, podName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pod restarted"})
}

// GetPodLogs gets logs for an application
func (h *ApplicationHandler) GetPodLogs(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	query := application.LogQuery{
		ApplicationID: appID,
		PodName:       c.Query("pod"),
		Container:     c.Query("container"),
		Limit:         100, // Default limit
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			query.Limit = limit
		}
	}

	// Parse since
	if sinceStr := c.Query("since"); sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			query.Since = since
		}
	}

	logs, err := h.appService.GetPodLogs(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// StreamPodLogs streams logs for an application
func (h *ApplicationHandler) StreamPodLogs(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	query := application.LogQuery{
		ApplicationID: appID,
		PodName:       c.Query("pod"),
		Container:     c.Query("container"),
		Follow:        true,
	}

	stream, err := h.appService.StreamPodLogs(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// Stream the logs
	// TODO: Implement proper SSE streaming
	c.String(http.StatusOK, "Log streaming not fully implemented")
}

// GetApplicationMetrics gets metrics for an application
func (h *ApplicationHandler) GetApplicationMetrics(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	metrics, err := h.appService.GetApplicationMetrics(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetApplicationEvents gets events for an application
func (h *ApplicationHandler) GetApplicationEvents(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	limit := 50 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	events, err := h.appService.GetApplicationEvents(c.Request.Context(), appID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// UpdateNetworkConfig updates network configuration for an application
func (h *ApplicationHandler) UpdateNetworkConfig(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var config application.NetworkConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	err := h.appService.UpdateNetworkConfig(c.Request.Context(), appID, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "network configuration updated"})
}

// GetApplicationEndpoints gets endpoints for an application
func (h *ApplicationHandler) GetApplicationEndpoints(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	endpoints, err := h.appService.GetApplicationEndpoints(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, endpoints)
}

// UpdateNodeAffinity updates node affinity for an application
func (h *ApplicationHandler) UpdateNodeAffinity(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var req struct {
		NodeSelector map[string]string `json:"node_selector"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	err := h.appService.UpdateNodeAffinity(c.Request.Context(), appID, req.NodeSelector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "node affinity updated"})
}

// MigrateToNode migrates an application to a specific node
func (h *ApplicationHandler) MigrateToNode(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var req struct {
		TargetNodeID string `json:"target_node_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	err := h.appService.MigrateToNode(c.Request.Context(), appID, req.TargetNodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application migration started"})
}

// UpdateCronJobSchedule updates the schedule of a CronJob
func (h *ApplicationHandler) UpdateCronJobSchedule(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var req application.UpdateCronScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	err := h.appService.UpdateCronJobSchedule(c.Request.Context(), appID, req.Schedule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cron schedule updated"})
}

// TriggerCronJob manually triggers a CronJob
func (h *ApplicationHandler) TriggerCronJob(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	req := &application.TriggerCronJobRequest{
		ApplicationID: appID,
	}

	execution, err := h.appService.TriggerCronJob(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// GetCronJobExecutions retrieves executions for a CronJob
func (h *ApplicationHandler) GetCronJobExecutions(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	// Parse pagination parameters
	page := 1
	perPage := 10
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	offset := (page - 1) * perPage

	executions, total, err := h.appService.GetCronJobExecutions(c.Request.Context(), appID, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := application.CronJobExecutionList{
		Executions: executions,
		Total:      total,
		Page:       page,
		PageSize:   perPage,
	}

	c.JSON(http.StatusOK, response)
}

// GetCronJobStatus retrieves the current status of a CronJob
func (h *ApplicationHandler) GetCronJobStatus(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	status, err := h.appService.GetCronJobStatus(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// CreateFunction creates a new serverless function
func (h *ApplicationHandler) CreateFunction(c *gin.Context) {
	workspaceID := c.Param("wsId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace ID is required"})
		return
	}

	var req application.CreateFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	app, err := h.appService.CreateFunction(c.Request.Context(), workspaceID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, app)
}

// DeployFunctionVersion creates and deploys a new version of a function
func (h *ApplicationHandler) DeployFunctionVersion(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var req struct {
		SourceCode string `json:"source_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.SourceCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source code is required"})
		return
	}

	version, err := h.appService.DeployFunctionVersion(c.Request.Context(), appID, req.SourceCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, version)
}

// GetFunctionVersions retrieves all versions of a function
func (h *ApplicationHandler) GetFunctionVersions(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	versions, err := h.appService.GetFunctionVersions(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, versions)
}

// SetActiveFunctionVersion sets the active version for a function
func (h *ApplicationHandler) SetActiveFunctionVersion(c *gin.Context) {
	appID := c.Param("appId")
	versionID := c.Param("versionId")
	if appID == "" || versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID and version ID are required"})
		return
	}

	err := h.appService.SetActiveFunctionVersion(c.Request.Context(), appID, versionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "active version updated"})
}

// InvokeFunction invokes a function synchronously
func (h *ApplicationHandler) InvokeFunction(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	var req application.InvokeFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Set defaults if not provided
	if req.Method == "" {
		req.Method = "POST"
	}
	if req.Path == "" {
		req.Path = "/"
	}

	resp, err := h.appService.InvokeFunction(c.Request.Context(), appID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetFunctionInvocations retrieves invocation history for a function
func (h *ApplicationHandler) GetFunctionInvocations(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	// Parse pagination parameters
	page := 1
	perPage := 20
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	offset := (page - 1) * perPage

	invocations, total, err := h.appService.GetFunctionInvocations(c.Request.Context(), appID, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := application.FunctionInvocationList{
		Invocations: invocations,
		Total:       total,
		Page:        page,
		PageSize:    perPage,
	}

	c.JSON(http.StatusOK, response)
}

// GetFunctionEvents retrieves pending events for a function
func (h *ApplicationHandler) GetFunctionEvents(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application ID is required"})
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	events, err := h.appService.GetFunctionEvents(c.Request.Context(), appID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// ProcessFunctionEvent processes a function event
func (h *ApplicationHandler) ProcessFunctionEvent(c *gin.Context) {
	eventID := c.Param("eventId")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event ID is required"})
		return
	}

	err := h.appService.ProcessFunctionEvent(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event processed"})
}