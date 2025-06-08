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
	applicationSvc application.Service
}

// NewApplicationHandler creates a new ApplicationHandler
func NewApplicationHandler(applicationSvc application.Service) *ApplicationHandler {
	return &ApplicationHandler{
		applicationSvc: applicationSvc,
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

	// Validate source type
	if !req.Source.Type.IsValid() {
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
	app, err := h.applicationSvc.CreateApplication(c.Request.Context(), workspaceID, req)
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

	app, err := h.applicationSvc.GetApplication(c.Request.Context(), appID)
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

	apps, err := h.applicationSvc.ListApplications(c.Request.Context(), workspaceID, projectID)
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

	app, err := h.applicationSvc.UpdateApplication(c.Request.Context(), appID, req)
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

	err := h.applicationSvc.DeleteApplication(c.Request.Context(), appID)
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

	err := h.applicationSvc.StartApplication(c.Request.Context(), appID)
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

	err := h.applicationSvc.StopApplication(c.Request.Context(), appID)
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

	err := h.applicationSvc.RestartApplication(c.Request.Context(), appID)
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

	err := h.applicationSvc.ScaleApplication(c.Request.Context(), appID, req.Replicas)
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

	pods, err := h.applicationSvc.ListPods(c.Request.Context(), appID)
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

	err := h.applicationSvc.RestartPod(c.Request.Context(), appID, podName)
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

	logs, err := h.applicationSvc.GetPodLogs(c.Request.Context(), query)
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

	stream, err := h.applicationSvc.StreamPodLogs(c.Request.Context(), query)
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

	metrics, err := h.applicationSvc.GetApplicationMetrics(c.Request.Context(), appID)
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

	events, err := h.applicationSvc.GetApplicationEvents(c.Request.Context(), appID, limit)
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

	err := h.applicationSvc.UpdateNetworkConfig(c.Request.Context(), appID, config)
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

	endpoints, err := h.applicationSvc.GetApplicationEndpoints(c.Request.Context(), appID)
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

	err := h.applicationSvc.UpdateNodeAffinity(c.Request.Context(), appID, req.NodeSelector)
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

	err := h.applicationSvc.MigrateToNode(c.Request.Context(), appID, req.TargetNodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application migration started"})
}