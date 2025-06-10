package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
)

// FunctionRequest represents a function creation/update request
type FunctionRequest struct {
	Name        string                              `json:"name"`
	Runtime     string                              `json:"runtime"`
	Handler     string                              `json:"handler"`
	SourceCode  string                              `json:"source_code"`
	Environment map[string]string                   `json:"environment,omitempty"`
	Resources   *function.FunctionResourceRequirements `json:"resources,omitempty"`
	Labels      map[string]string                   `json:"labels,omitempty"`
	Annotations map[string]string                   `json:"annotations,omitempty"`
}

// VersionRequest represents a version deployment request
type VersionRequest struct {
	ID         string `json:"id,omitempty"`
	Version    int    `json:"version"`
	SourceCode string `json:"source_code"`
	Image      string `json:"image,omitempty"`
}

// TriggerRequest represents a trigger creation/update request
type TriggerRequest struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Enabled bool              `json:"enabled"`
	Config  map[string]string `json:"config,omitempty"`
}

// InvokeRequest represents a function invocation request
type InvokeRequest struct {
	Method  string              `json:"method,omitempty"`
	Path    string              `json:"path,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    []byte              `json:"body,omitempty"`
	Query   map[string][]string `json:"query,omitempty"`
}

// CreateFunctionGin handles POST /api/v1/workspaces/:wsId/functions
func (h *FunctionHandler) CreateFunctionGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	var req FunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	spec := &function.FunctionSpec{
		Name:        req.Name,
		Runtime:     function.Runtime(req.Runtime),
		Handler:     req.Handler,
		SourceCode:  req.SourceCode,
		Environment: req.Environment,
		Resources:   *req.Resources,
		Labels:      req.Labels,
		Annotations: req.Annotations,
	}

	fn, err := h.service.CreateFunction(c.Request.Context(), workspaceID, projectID, spec)
	if err != nil {
		h.logger.Error("Failed to create function", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create function"})
		return
	}

	c.JSON(http.StatusCreated, fn)
}

// ListFunctionsGin handles GET /api/v1/workspaces/:wsId/functions
func (h *FunctionHandler) ListFunctionsGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	functions, err := h.service.ListFunctions(c.Request.Context(), workspaceID, projectID)
	if err != nil {
		h.logger.Error("Failed to list functions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list functions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"functions": functions})
}

// GetFunctionGin handles GET /api/v1/workspaces/:wsId/functions/:functionId
func (h *FunctionHandler) GetFunctionGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	fn, err := h.service.GetFunction(c.Request.Context(), workspaceID, functionID)
	if err != nil {
		h.logger.Error("Failed to get function", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		return
	}

	c.JSON(http.StatusOK, fn)
}

// UpdateFunctionGin handles PUT /api/v1/workspaces/:wsId/functions/:functionId
func (h *FunctionHandler) UpdateFunctionGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	var req FunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	spec := &function.FunctionSpec{
		Name:        req.Name,
		Runtime:     function.Runtime(req.Runtime),
		Handler:     req.Handler,
		SourceCode:  req.SourceCode,
		Environment: req.Environment,
		Resources:   *req.Resources,
		Labels:      req.Labels,
		Annotations: req.Annotations,
	}

	fn, err := h.service.UpdateFunction(c.Request.Context(), workspaceID, functionID, spec)
	if err != nil {
		h.logger.Error("Failed to update function", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update function"})
		return
	}

	c.JSON(http.StatusOK, fn)
}

// DeleteFunctionGin handles DELETE /api/v1/workspaces/:wsId/functions/:functionId
func (h *FunctionHandler) DeleteFunctionGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	err := h.service.DeleteFunction(c.Request.Context(), workspaceID, functionID)
	if err != nil {
		h.logger.Error("Failed to delete function", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete function"})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeployVersionGin handles POST /api/v1/workspaces/:wsId/functions/:functionId/versions
func (h *FunctionHandler) DeployVersionGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	var req VersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	version := &function.FunctionVersionDef{
		ID:         req.ID,
		Version:    req.Version,
		SourceCode: req.SourceCode,
		Image:      req.Image,
	}

	result, err := h.service.DeployVersion(c.Request.Context(), workspaceID, functionID, version)
	if err != nil {
		h.logger.Error("Failed to deploy version", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deploy version"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// ListVersionsGin handles GET /api/v1/workspaces/:wsId/functions/:functionId/versions
func (h *FunctionHandler) ListVersionsGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	versions, err := h.service.ListVersions(c.Request.Context(), workspaceID, functionID)
	if err != nil {
		h.logger.Error("Failed to list versions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list versions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

// GetVersionGin handles GET /api/v1/workspaces/:wsId/functions/:functionId/versions/:versionId
func (h *FunctionHandler) GetVersionGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")
	versionID := c.Param("versionId")

	version, err := h.service.GetVersion(c.Request.Context(), workspaceID, functionID, versionID)
	if err != nil {
		h.logger.Error("Failed to get version", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	c.JSON(http.StatusOK, version)
}

// SetActiveVersionGin handles PUT /api/v1/workspaces/:wsId/functions/:functionId/versions/:versionId/active
func (h *FunctionHandler) SetActiveVersionGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")
	versionID := c.Param("versionId")

	err := h.service.SetActiveVersion(c.Request.Context(), workspaceID, functionID, versionID)
	if err != nil {
		h.logger.Error("Failed to set active version", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set active version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Version activated successfully"})
}

// RollbackVersionGin handles POST /api/v1/workspaces/:wsId/functions/:functionId/rollback
func (h *FunctionHandler) RollbackVersionGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	err := h.service.RollbackVersion(c.Request.Context(), workspaceID, functionID)
	if err != nil {
		h.logger.Error("Failed to rollback version", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rollback version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rollback successful"})
}

// CreateTriggerGin handles POST /api/v1/workspaces/:wsId/functions/:functionId/triggers
func (h *FunctionHandler) CreateTriggerGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	var req TriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	trigger := &function.FunctionTrigger{
		Name:    req.Name,
		Type:    function.TriggerType(req.Type),
		Config:  req.Config,
		Enabled: req.Enabled,
	}

	result, err := h.service.CreateTrigger(c.Request.Context(), workspaceID, functionID, trigger)
	if err != nil {
		h.logger.Error("Failed to create trigger", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trigger"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// ListTriggersGin handles GET /api/v1/workspaces/:wsId/functions/:functionId/triggers
func (h *FunctionHandler) ListTriggersGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	triggers, err := h.service.ListTriggers(c.Request.Context(), workspaceID, functionID)
	if err != nil {
		h.logger.Error("Failed to list triggers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list triggers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"triggers": triggers})
}

// UpdateTriggerGin handles PUT /api/v1/workspaces/:wsId/functions/:functionId/triggers/:triggerId
func (h *FunctionHandler) UpdateTriggerGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")
	triggerID := c.Param("triggerId")

	var req TriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	trigger := &function.FunctionTrigger{
		ID:      triggerID,
		Name:    req.Name,
		Type:    function.TriggerType(req.Type),
		Config:  req.Config,
		Enabled: req.Enabled,
	}

	result, err := h.service.UpdateTrigger(c.Request.Context(), workspaceID, functionID, triggerID, trigger)
	if err != nil {
		h.logger.Error("Failed to update trigger", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trigger"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteTriggerGin handles DELETE /api/v1/workspaces/:wsId/functions/:functionId/triggers/:triggerId
func (h *FunctionHandler) DeleteTriggerGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")
	triggerID := c.Param("triggerId")

	err := h.service.DeleteTrigger(c.Request.Context(), workspaceID, functionID, triggerID)
	if err != nil {
		h.logger.Error("Failed to delete trigger", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete trigger"})
		return
	}

	c.Status(http.StatusNoContent)
}

// InvokeFunctionGin handles POST /api/v1/workspaces/:wsId/functions/:functionId/invoke
func (h *FunctionHandler) InvokeFunctionGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	var req InvokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	invokeReq := &function.InvokeRequest{
		Method:  req.Method,
		Path:    req.Path,
		Headers: req.Headers,
		Body:    req.Body,
		QueryParams: req.Query,
	}

	response, err := h.service.InvokeFunction(c.Request.Context(), workspaceID, functionID, invokeReq)
	if err != nil {
		h.logger.Error("Failed to invoke function", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invoke function"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// InvokeFunctionAsyncGin handles POST /api/v1/workspaces/:wsId/functions/:functionId/invoke-async
func (h *FunctionHandler) InvokeFunctionAsyncGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	var req InvokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	invokeReq := &function.InvokeRequest{
		Method:  req.Method,
		Path:    req.Path,
		Headers: req.Headers,
		Body:    req.Body,
		QueryParams: req.Query,
	}

	invocationID, err := h.service.InvokeFunctionAsync(c.Request.Context(), workspaceID, functionID, invokeReq)
	if err != nil {
		h.logger.Error("Failed to invoke function async", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invoke function async"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"invocation_id": invocationID})
}

// GetInvocationStatusGin handles GET /api/v1/workspaces/:wsId/functions/invocations/:invocationId
func (h *FunctionHandler) GetInvocationStatusGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	invocationID := c.Param("invocationId")

	status, err := h.service.GetInvocationStatus(c.Request.Context(), workspaceID, invocationID)
	if err != nil {
		h.logger.Error("Failed to get invocation status", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Invocation not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ListInvocationsGin handles GET /api/v1/workspaces/:wsId/functions/:functionId/invocations
func (h *FunctionHandler) ListInvocationsGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	invocations, err := h.service.ListInvocations(c.Request.Context(), workspaceID, functionID, limit)
	if err != nil {
		h.logger.Error("Failed to list invocations", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list invocations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invocations": invocations})
}

// GetFunctionLogsGin handles GET /api/v1/workspaces/:wsId/functions/:functionId/logs
func (h *FunctionHandler) GetFunctionLogsGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	var since, until *time.Time
	if s := c.Query("since"); s != "" {
		if parsed, err := time.Parse(time.RFC3339, s); err == nil {
			since = &parsed
		}
	}
	if u := c.Query("until"); u != "" {
		if parsed, err := time.Parse(time.RFC3339, u); err == nil {
			until = &parsed
		}
	}

	opts := &function.LogOptions{
		Since:      since,
		Until:      until,
		Limit:      100,
		Follow:     c.Query("follow") == "true",
		Previous:   c.Query("previous") == "true",
	}

	if t := c.Query("limit"); t != "" {
		if parsed, err := strconv.Atoi(t); err == nil && parsed > 0 {
			opts.Limit = parsed
		}
	}

	logs, err := h.service.GetFunctionLogs(c.Request.Context(), workspaceID, functionID, opts)
	if err != nil {
		h.logger.Error("Failed to get function logs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get function logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// GetFunctionMetricsGin handles GET /api/v1/workspaces/:wsId/functions/:functionId/metrics
func (h *FunctionHandler) GetFunctionMetricsGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	startTime := time.Now().Add(-1 * time.Hour) // Default to last hour
	endTime := time.Now()
	
	if s := c.Query("start"); s != "" {
		if parsed, err := time.Parse(time.RFC3339, s); err == nil {
			startTime = parsed
		}
	}
	if e := c.Query("end"); e != "" {
		if parsed, err := time.Parse(time.RFC3339, e); err == nil {
			endTime = parsed
		}
	}

	opts := &function.MetricOptions{
		StartTime:  startTime,
		EndTime:    endTime,
		Resolution: c.Query("resolution"),
		Metrics:    c.QueryArray("metrics"),
	}

	metrics, err := h.service.GetFunctionMetrics(c.Request.Context(), workspaceID, functionID, opts)
	if err != nil {
		h.logger.Error("Failed to get function metrics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get function metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetFunctionEventsGin handles GET /api/v1/workspaces/:wsId/functions/:functionId/events
func (h *FunctionHandler) GetFunctionEventsGin(c *gin.Context) {
	workspaceID := c.Param("wsId")
	functionID := c.Param("functionId")

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	events, err := h.service.GetFunctionEvents(c.Request.Context(), workspaceID, functionID, limit)
	if err != nil {
		h.logger.Error("Failed to get function events", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get function events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

// GetProviderCapabilitiesGin handles GET /api/v1/workspaces/:wsId/functions/provider/capabilities
func (h *FunctionHandler) GetProviderCapabilitiesGin(c *gin.Context) {
	workspaceID := c.Param("wsId")

	capabilities, err := h.service.GetProviderCapabilities(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("Failed to get provider capabilities", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider capabilities"})
		return
	}

	c.JSON(http.StatusOK, capabilities)
}

// GetProviderHealthGin handles GET /api/v1/workspaces/:wsId/functions/provider/health
func (h *FunctionHandler) GetProviderHealthGin(c *gin.Context) {
	workspaceID := c.Param("wsId")

	err := h.service.GetProviderHealth(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("Failed to get provider health", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Provider unhealthy", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}