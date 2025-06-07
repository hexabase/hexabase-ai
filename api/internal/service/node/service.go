package node

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/node"
)

// Service implements the node business logic
type Service struct {
	nodeRepo    node.Repository
	proxmoxRepo node.ProxmoxRepository
}

// NewService creates a new node service
func NewService(nodeRepo node.Repository, proxmoxRepo node.ProxmoxRepository) *Service {
	return &Service{
		nodeRepo:    nodeRepo,
		proxmoxRepo: proxmoxRepo,
	}
}

// GetAvailablePlans returns all available node plans
func (s *Service) GetAvailablePlans(ctx context.Context) ([]node.NodePlan, error) {
	return s.nodeRepo.GetNodePlans(ctx)
}

// GetPlanDetails returns details for a specific plan
func (s *Service) GetPlanDetails(ctx context.Context, planID string) (*node.NodePlan, error) {
	return s.nodeRepo.GetNodePlan(ctx, planID)
}

// ProvisionDedicatedNode provisions a new dedicated node for a workspace
func (s *Service) ProvisionDedicatedNode(ctx context.Context, workspaceID string, req node.ProvisionRequest) (*node.DedicatedNode, error) {
	// Validate request
	if err := s.validateProvisionRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get workspace allocation
	allocation, err := s.nodeRepo.GetWorkspaceAllocation(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace allocation: %w", err)
	}

	// Generate unique node ID
	nodeID := uuid.New().String()

	// Get node specification based on type
	spec, err := s.getNodeSpecification(req.NodeType)
	if err != nil {
		return nil, fmt.Errorf("invalid node type: %w", err)
	}

	// Create VM specification for Proxmox
	vmSpec := node.VMSpec{
		Name:          req.NodeName,
		NodeType:      req.NodeType,
		TemplateID:    s.getTemplateID(req.NodeType),
		TargetNode:    s.selectProxmoxNode(req.Region),
		CPUCores:      spec.CPUCores,
		MemoryMB:      spec.MemoryGB * 1024,
		DiskGB:        spec.StorageGB,
		NetworkBridge: "vmbr0",
		CloudInit: node.CloudInitConfig{
			SSHKeys: []string{req.SSHPublicKey},
			UserData: s.generateCloudInitUserData(workspaceID, nodeID),
		},
		Tags: []string{"hexabase", "workspace:" + workspaceID},
	}

	// Create dedicated node record
	dedicatedNode := &node.DedicatedNode{
		ID:            nodeID,
		WorkspaceID:   workspaceID,
		Name:          req.NodeName,
		Status:        node.NodeStatusProvisioning,
		Specification: *spec,
		SSHPublicKey:  req.SSHPublicKey,
		Labels:        req.Labels,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save node record first
	if err := s.nodeRepo.CreateDedicatedNode(ctx, dedicatedNode); err != nil {
		return nil, fmt.Errorf("failed to create node record: %w", err)
	}

	// Create event for provisioning start
	event := dedicatedNode.CreateEvent(node.EventTypeProvisioning, "Starting node provisioning")
	if err := s.nodeRepo.CreateNodeEvent(ctx, &event); err != nil {
		// Log but don't fail - events are not critical
	}

	// Create VM in Proxmox
	vmInfo, err := s.proxmoxRepo.CreateVM(ctx, vmSpec)
	if err != nil {
		// Update node status to failed
		dedicatedNode.Status = node.NodeStatusFailed
		s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode)
		
		// Create failure event
		failEvent := dedicatedNode.CreateEvent(node.EventTypeError, fmt.Sprintf("VM creation failed: %v", err))
		s.nodeRepo.CreateNodeEvent(ctx, &failEvent)
		
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Update node with VM information
	dedicatedNode.ProxmoxVMID = vmInfo.VMID
	dedicatedNode.ProxmoxNode = vmInfo.Node
	dedicatedNode.IPAddress = s.extractIPAddress(vmInfo)
	dedicatedNode.Status = node.NodeStatusReady
	dedicatedNode.UpdatedAt = time.Now()

	if err := s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode); err != nil {
		return nil, fmt.Errorf("failed to update node record: %w", err)
	}

	// Transition workspace to dedicated plan if this is the first node
	if allocation.PlanType == node.PlanTypeShared {
		allocation.PlanType = node.PlanTypeDedicated
		allocation.SharedQuota = nil // Clear shared quota
		allocation.UpdatedAt = time.Now()
		
		if err := s.nodeRepo.UpdateWorkspaceAllocation(ctx, allocation); err != nil {
			// Log but don't fail - billing transition can be handled separately
		}
	}

	// Create completion event
	completeEvent := dedicatedNode.CreateEvent(node.EventTypeStatusChange, "Node provisioning completed")
	s.nodeRepo.CreateNodeEvent(ctx, &completeEvent)

	return dedicatedNode, nil
}

// GetNode returns a dedicated node by ID
func (s *Service) GetNode(ctx context.Context, nodeID string) (*node.DedicatedNode, error) {
	return s.nodeRepo.GetDedicatedNode(ctx, nodeID)
}

// ListNodes returns all dedicated nodes for a workspace
func (s *Service) ListNodes(ctx context.Context, workspaceID string) ([]node.DedicatedNode, error) {
	return s.nodeRepo.ListDedicatedNodes(ctx, workspaceID)
}

// StartNode starts a stopped node
func (s *Service) StartNode(ctx context.Context, nodeID string) error {
	return s.performNodeAction(ctx, nodeID, "start", func(n *node.DedicatedNode) error {
		return s.proxmoxRepo.StartVM(ctx, n.ProxmoxVMID)
	})
}

// StopNode stops a running node
func (s *Service) StopNode(ctx context.Context, nodeID string) error {
	return s.performNodeAction(ctx, nodeID, "stop", func(n *node.DedicatedNode) error {
		return s.proxmoxRepo.StopVM(ctx, n.ProxmoxVMID)
	})
}

// RebootNode reboots a node
func (s *Service) RebootNode(ctx context.Context, nodeID string) error {
	return s.performNodeAction(ctx, nodeID, "reboot", func(n *node.DedicatedNode) error {
		return s.proxmoxRepo.RebootVM(ctx, n.ProxmoxVMID)
	})
}

// DeleteNode deletes a dedicated node
func (s *Service) DeleteNode(ctx context.Context, nodeID string) error {
	dedicatedNode, err := s.nodeRepo.GetDedicatedNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Transition to deleting status
	if err := dedicatedNode.TransitionTo("delete"); err != nil {
		return fmt.Errorf("cannot delete node in current status: %w", err)
	}

	if err := s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	// Create deletion event
	event := dedicatedNode.CreateEvent(node.EventTypeDeletion, "Starting node deletion")
	s.nodeRepo.CreateNodeEvent(ctx, &event)

	// Delete VM from Proxmox
	if err := s.proxmoxRepo.DeleteVM(ctx, dedicatedNode.ProxmoxVMID); err != nil {
		// Update status to failed but continue with soft deletion
		dedicatedNode.Status = node.NodeStatusFailed
		s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode)
		
		failEvent := dedicatedNode.CreateEvent(node.EventTypeError, fmt.Sprintf("VM deletion failed: %v", err))
		s.nodeRepo.CreateNodeEvent(ctx, &failEvent)
		
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	// Mark as deleted in database
	if err := s.nodeRepo.DeleteDedicatedNode(ctx, nodeID); err != nil {
		return fmt.Errorf("failed to delete node record: %w", err)
	}

	return nil
}

// GetWorkspaceResourceUsage returns resource usage for a workspace
func (s *Service) GetWorkspaceResourceUsage(ctx context.Context, workspaceID string) (*node.WorkspaceResourceUsage, error) {
	allocation, err := s.nodeRepo.GetWorkspaceAllocation(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace allocation: %w", err)
	}

	usage := &node.WorkspaceResourceUsage{
		WorkspaceID: workspaceID,
		PlanType:    allocation.PlanType,
		Timestamp:   time.Now(),
	}

	if allocation.PlanType == node.PlanTypeShared {
		// Calculate shared resource usage
		if allocation.SharedQuota != nil {
			usage.SharedUsage = &node.SharedResourceUsage{
				CPUUsed:       allocation.SharedQuota.CPUUsed,
				CPULimit:      allocation.SharedQuota.CPULimit,
				MemoryUsedGB:  allocation.SharedQuota.MemoryUsed,
				MemoryLimitGB: allocation.SharedQuota.MemoryLimit,
			}
		}
	} else {
		// Calculate dedicated resource usage
		nodes, err := s.nodeRepo.ListDedicatedNodes(ctx, workspaceID)
		if err != nil {
			return nil, fmt.Errorf("failed to list dedicated nodes: %w", err)
		}

		dedicatedUsage := &node.DedicatedResourceUsage{
			TotalNodes: len(nodes),
			NodeUsage:  make([]node.NodeResourceSummary, 0, len(nodes)),
		}

		for _, n := range nodes {
			if n.CanScheduleWorkload() {
				dedicatedUsage.ActiveNodes++
			}

			dedicatedUsage.TotalCPUCores += n.Specification.CPUCores
			dedicatedUsage.TotalMemoryGB += n.Specification.MemoryGB
			dedicatedUsage.TotalStorageGB += n.Specification.StorageGB

			// Get current resource usage from Proxmox
			if vmUsage, err := s.proxmoxRepo.GetVMResourceUsage(ctx, n.ProxmoxVMID); err == nil {
				summary := node.NodeResourceSummary{
					NodeID:       n.ID,
					NodeName:     n.Name,
					CPUUsage:     vmUsage.CPUUsage,
					MemoryUsage:  float64(vmUsage.MemoryUsage) / (1024 * 1024 * 1024), // Convert to GB percentage
					StorageUsage: float64(vmUsage.DiskUsage) / (1024 * 1024 * 1024),   // Convert to GB percentage
					Status:       string(n.Status),
				}
				dedicatedUsage.NodeUsage = append(dedicatedUsage.NodeUsage, summary)
			}
		}

		usage.DedicatedUsage = dedicatedUsage
	}

	return usage, nil
}

// CanAllocateResources checks if workspace can allocate requested resources
func (s *Service) CanAllocateResources(ctx context.Context, workspaceID string, request node.ResourceRequest) (bool, error) {
	allocation, err := s.nodeRepo.GetWorkspaceAllocation(ctx, workspaceID)
	if err != nil {
		return false, fmt.Errorf("failed to get workspace allocation: %w", err)
	}

	return allocation.CanAllocate(request), nil
}

// GetNodeStatus returns detailed status information for a node
func (s *Service) GetNodeStatus(ctx context.Context, nodeID string) (*node.NodeStatusInfo, error) {
	dedicatedNode, err := s.nodeRepo.GetDedicatedNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Get Proxmox VM status
	proxmoxStatus, err := s.proxmoxRepo.GetVMStatus(ctx, dedicatedNode.ProxmoxVMID)
	if err != nil {
		proxmoxStatus = "unknown"
	}

	statusInfo := &node.NodeStatusInfo{
		NodeID:        nodeID,
		Status:        dedicatedNode.Status,
		ProxmoxStatus: proxmoxStatus,
		K3sStatus:     "unknown", // TODO: Implement K3s status check
		LastUpdated:   dedicatedNode.UpdatedAt,
		Conditions:    []node.NodeCondition{},
	}

	// Add basic conditions based on status
	if dedicatedNode.Status == node.NodeStatusReady && proxmoxStatus == "running" {
		statusInfo.Conditions = append(statusInfo.Conditions, node.NodeCondition{
			Type:    "Ready",
			Status:  "True",
			Reason:  "NodeReady",
			Message: "Node is ready to accept workloads",
			Since:   dedicatedNode.UpdatedAt,
		})
	}

	return statusInfo, nil
}

// GetNodeMetrics returns performance metrics for a node
func (s *Service) GetNodeMetrics(ctx context.Context, nodeID string) (*node.NodeMetrics, error) {
	dedicatedNode, err := s.nodeRepo.GetDedicatedNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	vmUsage, err := s.proxmoxRepo.GetVMResourceUsage(ctx, dedicatedNode.ProxmoxVMID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM resource usage: %w", err)
	}

	metrics := &node.NodeMetrics{
		NodeID:    nodeID,
		Timestamp: time.Now(),
		CPU: node.CPUMetrics{
			UsagePercent: vmUsage.CPUUsage,
		},
		Memory: node.MemoryMetrics{
			UsedBytes:    vmUsage.MemoryUsage,
			TotalBytes:   int64(dedicatedNode.Specification.MemoryGB) * 1024 * 1024 * 1024,
			UsagePercent: float64(vmUsage.MemoryUsage) / float64(dedicatedNode.Specification.MemoryGB*1024*1024*1024) * 100,
		},
		Disk: node.DiskMetrics{
			UsedBytes:    vmUsage.DiskUsage,
			TotalBytes:   int64(dedicatedNode.Specification.StorageGB) * 1024 * 1024 * 1024,
			UsagePercent: float64(vmUsage.DiskUsage) / float64(dedicatedNode.Specification.StorageGB*1024*1024*1024) * 100,
		},
		Network: node.NetworkMetrics{
			BytesIn:  vmUsage.NetworkIn,
			BytesOut: vmUsage.NetworkOut,
		},
	}

	metrics.Memory.AvailableBytes = metrics.Memory.TotalBytes - metrics.Memory.UsedBytes
	metrics.Disk.FreeBytes = metrics.Disk.TotalBytes - metrics.Disk.UsedBytes

	return metrics, nil
}

// GetNodeEvents returns events for a node
func (s *Service) GetNodeEvents(ctx context.Context, nodeID string, limit int) ([]node.NodeEvent, error) {
	return s.nodeRepo.ListNodeEvents(ctx, nodeID, limit)
}

// GetNodeCosts calculates costs for nodes in a workspace
func (s *Service) GetNodeCosts(ctx context.Context, workspaceID string, period node.BillingPeriod) (*node.NodeCostReport, error) {
	nodes, err := s.nodeRepo.ListDedicatedNodes(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	report := &node.NodeCostReport{
		WorkspaceID: workspaceID,
		Period:      period,
		Currency:    "USD",
		NodeCosts:   make([]node.NodeCost, 0, len(nodes)),
	}

	for _, n := range nodes {
		cost := s.calculateNodeCost(n, period)
		report.NodeCosts = append(report.NodeCosts, cost)
		report.TotalCost += cost.TotalCost
	}

	return report, nil
}

// TransitionToSharedPlan transitions workspace to shared plan
func (s *Service) TransitionToSharedPlan(ctx context.Context, workspaceID string) error {
	allocation, err := s.nodeRepo.GetWorkspaceAllocation(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace allocation: %w", err)
	}

	if allocation.PlanType == node.PlanTypeShared {
		return nil // Already on shared plan
	}

	// Check if workspace has any active dedicated nodes
	nodes, err := s.nodeRepo.ListDedicatedNodes(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	activeNodes := 0
	for _, n := range nodes {
		if n.Status != node.NodeStatusDeleted {
			activeNodes++
		}
	}

	if activeNodes > 0 {
		return errors.New("cannot transition to shared plan while dedicated nodes exist")
	}

	// Transition to shared plan
	allocation.PlanType = node.PlanTypeShared
	allocation.SharedQuota = &node.SharedQuota{
		CPULimit:    2,
		MemoryLimit: 4,
		CPUUsed:     0,
		MemoryUsed:  0,
	}
	allocation.UpdatedAt = time.Now()

	return s.nodeRepo.UpdateWorkspaceAllocation(ctx, allocation)
}

// TransitionToDedicatedPlan transitions workspace to dedicated plan
func (s *Service) TransitionToDedicatedPlan(ctx context.Context, workspaceID string) error {
	allocation, err := s.nodeRepo.GetWorkspaceAllocation(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace allocation: %w", err)
	}

	if allocation.PlanType == node.PlanTypeDedicated {
		return nil // Already on dedicated plan
	}

	allocation.PlanType = node.PlanTypeDedicated
	allocation.SharedQuota = nil
	allocation.UpdatedAt = time.Now()

	return s.nodeRepo.UpdateWorkspaceAllocation(ctx, allocation)
}

// Helper methods

func (s *Service) validateProvisionRequest(req node.ProvisionRequest) error {
	if req.NodeName == "" {
		return errors.New("node name is required")
	}

	validTypes := map[string]bool{
		"S-Type": true,
		"M-Type": true,
		"L-Type": true,
	}

	if !validTypes[req.NodeType] {
		return errors.New("invalid node type")
	}

	return nil
}

func (s *Service) getNodeSpecification(nodeType string) (*node.NodeSpecification, error) {
	specs := map[string]node.NodeSpecification{
		"S-Type": {
			Type:        "S-Type",
			CPUCores:    4,
			MemoryGB:    16,
			StorageGB:   200,
			NetworkMbps: 1000,
		},
		"M-Type": {
			Type:        "M-Type",
			CPUCores:    8,
			MemoryGB:    32,
			StorageGB:   500,
			NetworkMbps: 2000,
		},
		"L-Type": {
			Type:        "L-Type",
			CPUCores:    16,
			MemoryGB:    64,
			StorageGB:   1000,
			NetworkMbps: 4000,
		},
	}

	spec, ok := specs[nodeType]
	if !ok {
		return nil, errors.New("unknown node type")
	}

	return &spec, nil
}

func (s *Service) getTemplateID(nodeType string) int {
	// Template IDs for different node types
	templates := map[string]int{
		"S-Type": 9000,
		"M-Type": 9001,
		"L-Type": 9002,
	}

	if id, ok := templates[nodeType]; ok {
		return id
	}

	return 9000 // Default template
}

func (s *Service) selectProxmoxNode(region string) string {
	// Simple node selection logic
	// In production, this would consider region, capacity, etc.
	return "pve-node1"
}

func (s *Service) generateCloudInitUserData(workspaceID, nodeID string) string {
	return fmt.Sprintf(`#cloud-config
users:
  - name: hexabase
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    lock_passwd: false

runcmd:
  - curl -sfL https://get.k3s.io | K3S_TOKEN=mytoken K3S_URL=https://k3s-server:6443 sh -s - agent
  - echo "workspace=%s node=%s" > /etc/hexabase/node.conf

final_message: "Hexabase node is ready"
`, workspaceID, nodeID)
}

func (s *Service) extractIPAddress(vmInfo *node.ProxmoxVMInfo) string {
	// In a real implementation, this would extract the IP from Proxmox
	// For now, return a placeholder
	return "10.0.0.100"
}

func (s *Service) performNodeAction(ctx context.Context, nodeID, action string, vmAction func(*node.DedicatedNode) error) error {
	dedicatedNode, err := s.nodeRepo.GetDedicatedNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Transition node state
	if err := dedicatedNode.TransitionTo(action); err != nil {
		return fmt.Errorf("cannot %s node in current status: %w", action, err)
	}

	// Update node status
	if err := s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	// Create event
	event := dedicatedNode.CreateEvent(node.EventTypeStatusChange, fmt.Sprintf("Node %s initiated", action))
	s.nodeRepo.CreateNodeEvent(ctx, &event)

	// Perform VM action
	if err := vmAction(dedicatedNode); err != nil {
		// Revert status on failure
		switch action {
		case "start":
			dedicatedNode.Status = node.NodeStatusStopped
		case "stop":
			dedicatedNode.Status = node.NodeStatusReady
		}
		s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode)
		
		failEvent := dedicatedNode.CreateEvent(node.EventTypeError, fmt.Sprintf("Node %s failed: %v", action, err))
		s.nodeRepo.CreateNodeEvent(ctx, &failEvent)
		
		return fmt.Errorf("VM %s failed: %w", action, err)
	}

	// Update to final status
	switch action {
	case "start":
		dedicatedNode.Status = node.NodeStatusReady
	case "stop":
		dedicatedNode.Status = node.NodeStatusStopped
	case "reboot":
		dedicatedNode.Status = node.NodeStatusReady
	}

	dedicatedNode.UpdatedAt = time.Now()
	if err := s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode); err != nil {
		return fmt.Errorf("failed to update final node status: %w", err)
	}

	// Create completion event
	completeEvent := dedicatedNode.CreateEvent(node.EventTypeStatusChange, fmt.Sprintf("Node %s completed", action))
	s.nodeRepo.CreateNodeEvent(ctx, &completeEvent)

	return nil
}

func (s *Service) calculateNodeCost(n node.DedicatedNode, period node.BillingPeriod) node.NodeCost {
	// Calculate hours the node was active during the period
	start := period.Start
	if n.CreatedAt.After(start) {
		start = n.CreatedAt
	}

	end := period.End
	if n.DeletedAt != nil && n.DeletedAt.Before(end) {
		end = *n.DeletedAt
	}

	duration := end.Sub(start)
	hoursActive := duration.Hours()

	// Get hourly rate based on node type
	hourlyRates := map[string]float64{
		"S-Type": 99.99 / (30 * 24),   // $99.99/month
		"M-Type": 199.99 / (30 * 24),  // $199.99/month
		"L-Type": 399.99 / (30 * 24),  // $399.99/month
	}

	hourlyRate := hourlyRates[n.Specification.Type]
	totalCost := hourlyRate * hoursActive

	return node.NodeCost{
		NodeID:      n.ID,
		NodeName:    n.Name,
		NodeType:    n.Specification.Type,
		HoursActive: hoursActive,
		HourlyRate:  hourlyRate,
		TotalCost:   totalCost,
		Status:      string(n.Status),
	}
}