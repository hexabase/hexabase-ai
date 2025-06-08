package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/node"
)

// NodeHandler handles node-related HTTP requests
type NodeHandler struct {
	service node.Service
	logger  *slog.Logger
}

// NewNodeHandler creates a new node handler
func NewNodeHandler(service node.Service, logger *slog.Logger) *NodeHandler {
	return &NodeHandler{
		service: service,
		logger:  logger,
	}
}

// ProvisionRequest represents a request to provision a new dedicated node
type ProvisionRequest struct {
	NodeName     string            `json:"node_name" binding:"required"`
	NodeType     string            `json:"node_type" binding:"required"`
	Region       string            `json:"region"`
	SSHPublicKey string            `json:"ssh_public_key" binding:"required"`
	Labels       map[string]string `json:"labels"`
}

// GetAvailablePlans returns all available node plans
func (h *NodeHandler) GetAvailablePlans(c *gin.Context) {
	plans, err := h.service.GetAvailablePlans(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get available plans", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve plans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// GetPlanDetails returns details for a specific plan
func (h *NodeHandler) GetPlanDetails(c *gin.Context) {
	planID := c.Param("planId")

	plan, err := h.service.GetPlanDetails(c.Request.Context(), planID)
	if err != nil {
		h.logger.Error("failed to get plan details", slog.String("error", err.Error()), slog.String("plan_id", planID))
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

// ProvisionDedicatedNode provisions a new dedicated node for a workspace
func (h *NodeHandler) ProvisionDedicatedNode(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.GetString("user_id")

	var req ProvisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Convert to domain request
	domainReq := node.ProvisionRequest{
		NodeName:     req.NodeName,
		NodeType:     req.NodeType,
		Region:       req.Region,
		SSHPublicKey: req.SSHPublicKey,
		Labels:       req.Labels,
	}

	dedicatedNode, err := h.service.ProvisionDedicatedNode(c.Request.Context(), workspaceID, domainReq)
	if err != nil {
		h.logger.Error("failed to provision dedicated node",
			slog.String("error", err.Error()),
			slog.String("workspace_id", workspaceID),
			slog.String("user_id", userID),
			slog.String("node_name", req.NodeName),
			slog.String("node_type", req.NodeType))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("dedicated node provisioned",
		slog.String("node_id", dedicatedNode.ID),
		slog.String("workspace_id", workspaceID),
		slog.String("user_id", userID),
		slog.String("node_name", req.NodeName),
		slog.String("node_type", req.NodeType))

	c.JSON(http.StatusCreated, gin.H{
		"node":    dedicatedNode,
		"message": "Node provisioning initiated successfully",
	})
}

// GetNode returns a dedicated node by ID
func (h *NodeHandler) GetNode(c *gin.Context) {
	nodeID := c.Param("nodeId")

	dedicatedNode, err := h.service.GetNode(c.Request.Context(), nodeID)
	if err != nil {
		h.logger.Error("failed to get node", slog.String("error", err.Error()), slog.String("node_id", nodeID))
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"node": dedicatedNode})
}

// ListNodes returns all dedicated nodes for a workspace
func (h *NodeHandler) ListNodes(c *gin.Context) {
	workspaceID := c.Param("wsId")

	nodes, err := h.service.ListNodes(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to list nodes", slog.String("error", err.Error()), slog.String("workspace_id", workspaceID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve nodes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

// StartNode starts a stopped node
func (h *NodeHandler) StartNode(c *gin.Context) {
	nodeID := c.Param("nodeId")
	userID := c.GetString("user_id")

	err := h.service.StartNode(c.Request.Context(), nodeID)
	if err != nil {
		h.logger.Error("failed to start node",
			slog.String("error", err.Error()),
			slog.String("node_id", nodeID),
			slog.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("node start initiated",
		slog.String("node_id", nodeID),
		slog.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{"message": "Node start initiated successfully"})
}

// StopNode stops a running node
func (h *NodeHandler) StopNode(c *gin.Context) {
	nodeID := c.Param("nodeId")
	userID := c.GetString("user_id")

	err := h.service.StopNode(c.Request.Context(), nodeID)
	if err != nil {
		h.logger.Error("failed to stop node",
			slog.String("error", err.Error()),
			slog.String("node_id", nodeID),
			slog.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("node stop initiated",
		slog.String("node_id", nodeID),
		slog.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{"message": "Node stop initiated successfully"})
}

// RebootNode reboots a node
func (h *NodeHandler) RebootNode(c *gin.Context) {
	nodeID := c.Param("nodeId")
	userID := c.GetString("user_id")

	err := h.service.RebootNode(c.Request.Context(), nodeID)
	if err != nil {
		h.logger.Error("failed to reboot node",
			slog.String("error", err.Error()),
			slog.String("node_id", nodeID),
			slog.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("node reboot initiated",
		slog.String("node_id", nodeID),
		slog.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{"message": "Node reboot initiated successfully"})
}

// DeleteNode deletes a dedicated node
func (h *NodeHandler) DeleteNode(c *gin.Context) {
	nodeID := c.Param("nodeId")
	userID := c.GetString("user_id")

	err := h.service.DeleteNode(c.Request.Context(), nodeID)
	if err != nil {
		h.logger.Error("failed to delete node",
			slog.String("error", err.Error()),
			slog.String("node_id", nodeID),
			slog.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("node deletion initiated",
		slog.String("node_id", nodeID),
		slog.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{"message": "Node deletion initiated successfully"})
}

// GetWorkspaceResourceUsage returns resource usage for a workspace
func (h *NodeHandler) GetWorkspaceResourceUsage(c *gin.Context) {
	workspaceID := c.Param("wsId")

	usage, err := h.service.GetWorkspaceResourceUsage(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get workspace resource usage",
			slog.String("error", err.Error()),
			slog.String("workspace_id", workspaceID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve resource usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"usage": usage})
}

// GetNodeStatus returns detailed status information for a node
func (h *NodeHandler) GetNodeStatus(c *gin.Context) {
	nodeID := c.Param("nodeId")

	status, err := h.service.GetNodeStatus(c.Request.Context(), nodeID)
	if err != nil {
		h.logger.Error("failed to get node status", slog.String("error", err.Error()), slog.String("node_id", nodeID))
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// GetNodeMetrics returns performance metrics for a node
func (h *NodeHandler) GetNodeMetrics(c *gin.Context) {
	nodeID := c.Param("nodeId")

	metrics, err := h.service.GetNodeMetrics(c.Request.Context(), nodeID)
	if err != nil {
		h.logger.Error("failed to get node metrics", slog.String("error", err.Error()), slog.String("node_id", nodeID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

// GetNodeEvents returns events for a node
func (h *NodeHandler) GetNodeEvents(c *gin.Context) {
	nodeID := c.Param("nodeId")

	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000 // Cap at 1000 events
	}

	events, err := h.service.GetNodeEvents(c.Request.Context(), nodeID, limit)
	if err != nil {
		h.logger.Error("failed to get node events", slog.String("error", err.Error()), slog.String("node_id", nodeID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"limit":  limit,
	})
}

// GetNodeCosts calculates costs for nodes in a workspace
func (h *NodeHandler) GetNodeCosts(c *gin.Context) {
	workspaceID := c.Param("wsId")

	// Parse query parameters for billing period
	startStr := c.Query("start")
	endStr := c.Query("end")

	var period node.BillingPeriod
	if startStr != "" && endStr != "" {
		var err error
		period.Start, err = node.ParseTime(startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start time format"})
			return
		}
		period.End, err = node.ParseTime(endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end time format"})
			return
		}
	} else {
		// Default to current month
		period = node.CurrentMonthPeriod()
	}

	report, err := h.service.GetNodeCosts(c.Request.Context(), workspaceID, period)
	if err != nil {
		h.logger.Error("failed to get node costs",
			slog.String("error", err.Error()),
			slog.String("workspace_id", workspaceID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate costs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

// CanAllocateResources checks if workspace can allocate requested resources
func (h *NodeHandler) CanAllocateResources(c *gin.Context) {
	workspaceID := c.Param("wsId")

	var req node.ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	canAllocate, err := h.service.CanAllocateResources(c.Request.Context(), workspaceID, req)
	if err != nil {
		h.logger.Error("failed to check resource allocation",
			slog.String("error", err.Error()),
			slog.String("workspace_id", workspaceID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check allocation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"can_allocate": canAllocate,
		"requested":    req,
	})
}

// TransitionToSharedPlan transitions workspace to shared plan
func (h *NodeHandler) TransitionToSharedPlan(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.GetString("user_id")

	err := h.service.TransitionToSharedPlan(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to transition to shared plan",
			slog.String("error", err.Error()),
			slog.String("workspace_id", workspaceID),
			slog.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("workspace transitioned to shared plan",
		slog.String("workspace_id", workspaceID),
		slog.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{"message": "Workspace transitioned to shared plan successfully"})
}

// TransitionToDedicatedPlan transitions workspace to dedicated plan
func (h *NodeHandler) TransitionToDedicatedPlan(c *gin.Context) {
	workspaceID := c.Param("wsId")
	userID := c.GetString("user_id")

	err := h.service.TransitionToDedicatedPlan(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to transition to dedicated plan",
			slog.String("error", err.Error()),
			slog.String("workspace_id", workspaceID),
			slog.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("workspace transitioned to dedicated plan",
		slog.String("workspace_id", workspaceID),
		slog.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{"message": "Workspace transitioned to dedicated plan successfully"})
}