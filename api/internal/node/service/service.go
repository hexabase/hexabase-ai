package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/node/domain"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Service implements the node business logic
type Service struct {
	nodeRepo    domain.Repository
	proxmoxRepo domain.ProxmoxRepository
	k8sClient   kubernetes.Interface
}

// NewService creates a new node service
func NewService(nodeRepo domain.Repository, proxmoxRepo domain.ProxmoxRepository) *Service {
	return &Service{
		nodeRepo:    nodeRepo,
		proxmoxRepo: proxmoxRepo,
	}
}

// SetK8sClient sets the Kubernetes client for the service
func (s *Service) SetK8sClient(client kubernetes.Interface) {
	s.k8sClient = client
}

// GetAvailablePlans returns all available node plans
func (s *Service) GetAvailablePlans(ctx context.Context) ([]domain.NodePlan, error) {
	return s.nodeRepo.GetNodePlans(ctx)
}

// GetPlanDetails returns details for a specific plan
func (s *Service) GetPlanDetails(ctx context.Context, planID string) (*domain.NodePlan, error) {
	return s.nodeRepo.GetNodePlan(ctx, planID)
}

// ProvisionDedicatedNode provisions a new dedicated node for a workspace
func (s *Service) ProvisionDedicatedNode(ctx context.Context, workspaceID string, req domain.ProvisionRequest) (*domain.DedicatedNode, error) {
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
	vmSpec := domain.VMSpec{
		Name:          req.NodeName,
		NodeType:      req.NodeType,
		TemplateID:    s.getTemplateID(req.NodeType),
		TargetNode:    s.selectProxmoxNode(req.Region),
		CPUCores:      spec.CPUCores,
		MemoryMB:      spec.MemoryGB * 1024,
		DiskGB:        spec.StorageGB,
		NetworkBridge: "vmbr0",
		CloudInit: domain.CloudInitConfig{
			SSHKeys: []string{req.SSHPublicKey},
			UserData: s.generateCloudInitUserData(workspaceID, nodeID),
		},
		Tags: []string{"hexabase", "workspace:" + workspaceID},
	}

	// Create dedicated node record
	dedicatedNode := &domain.DedicatedNode{
		ID:            nodeID,
		WorkspaceID:   workspaceID,
		Name:          req.NodeName,
		Status:        domain.NodeStatusProvisioning,
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
	event := dedicatedNode.CreateEvent(domain.EventTypeProvisioning, "Starting node provisioning")
	if err := s.nodeRepo.CreateNodeEvent(ctx, &event); err != nil {
		// Log but don't fail - events are not critical
	}

	// Create VM in Proxmox
	vmInfo, err := s.proxmoxRepo.CreateVM(ctx, vmSpec)
	if err != nil {
		// Update node status to failed
		dedicatedNode.Status = domain.NodeStatusFailed
		s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode)
		
		// Create failure event
		failEvent := dedicatedNode.CreateEvent(domain.EventTypeError, fmt.Sprintf("VM creation failed: %v", err))
		s.nodeRepo.CreateNodeEvent(ctx, &failEvent)
		
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Update node with VM information
	dedicatedNode.ProxmoxVMID = vmInfo.VMID
	dedicatedNode.ProxmoxNode = vmInfo.Node
	dedicatedNode.IPAddress = s.extractIPAddress(vmInfo)
	dedicatedNode.Status = domain.NodeStatusReady
	dedicatedNode.UpdatedAt = time.Now()

	if err := s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode); err != nil {
		return nil, fmt.Errorf("failed to update node record: %w", err)
	}

	// Transition workspace to dedicated plan if this is the first node
	if allocation.PlanType == domain.PlanTypeShared {
		allocation.PlanType = domain.PlanTypeDedicated
		allocation.SharedQuota = nil // Clear shared quota
		allocation.UpdatedAt = time.Now()
		
		if err := s.nodeRepo.UpdateWorkspaceAllocation(ctx, allocation); err != nil {
			// Log but don't fail - billing transition can be handled separately
		}
	}

	// Create completion event
	completeEvent := dedicatedNode.CreateEvent(domain.EventTypeStatusChange, "Node provisioning completed")
	s.nodeRepo.CreateNodeEvent(ctx, &completeEvent)

	return dedicatedNode, nil
}

// GetNode returns a dedicated node by ID
func (s *Service) GetNode(ctx context.Context, nodeID string) (*domain.DedicatedNode, error) {
	return s.nodeRepo.GetDedicatedNode(ctx, nodeID)
}

// ListNodes returns all dedicated nodes for a workspace
func (s *Service) ListNodes(ctx context.Context, workspaceID string) ([]domain.DedicatedNode, error) {
	return s.nodeRepo.ListDedicatedNodes(ctx, workspaceID)
}

// StartNode starts a stopped node
func (s *Service) StartNode(ctx context.Context, nodeID string) error {
	return s.performNodeAction(ctx, nodeID, "start", func(n *domain.DedicatedNode) error {
		return s.proxmoxRepo.StartVM(ctx, n.ProxmoxVMID)
	})
}

// StopNode stops a running node
func (s *Service) StopNode(ctx context.Context, nodeID string) error {
	return s.performNodeAction(ctx, nodeID, "stop", func(n *domain.DedicatedNode) error {
		return s.proxmoxRepo.StopVM(ctx, n.ProxmoxVMID)
	})
}

// RebootNode reboots a node
func (s *Service) RebootNode(ctx context.Context, nodeID string) error {
	return s.performNodeAction(ctx, nodeID, "reboot", func(n *domain.DedicatedNode) error {
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
	event := dedicatedNode.CreateEvent(domain.EventTypeDeletion, "Starting node deletion")
	s.nodeRepo.CreateNodeEvent(ctx, &event)

	// Delete VM from Proxmox
	if err := s.proxmoxRepo.DeleteVM(ctx, dedicatedNode.ProxmoxVMID); err != nil {
		// Update status to failed but continue with soft deletion
		dedicatedNode.Status = domain.NodeStatusFailed
		s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode)
		
		failEvent := dedicatedNode.CreateEvent(domain.EventTypeError, fmt.Sprintf("VM deletion failed: %v", err))
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
func (s *Service) GetWorkspaceResourceUsage(ctx context.Context, workspaceID string) (*domain.WorkspaceResourceUsage, error) {
	allocation, err := s.nodeRepo.GetWorkspaceAllocation(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace allocation: %w", err)
	}

	usage := &domain.WorkspaceResourceUsage{
		WorkspaceID: workspaceID,
		PlanType:    allocation.PlanType,
		Timestamp:   time.Now(),
	}

	if allocation.PlanType == domain.PlanTypeShared {
		// Calculate shared resource usage
		if allocation.SharedQuota != nil {
			usage.SharedUsage = &domain.SharedResourceUsage{
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

		dedicatedUsage := &domain.DedicatedResourceUsage{
			TotalNodes: len(nodes),
			NodeUsage:  make([]domain.NodeResourceSummary, 0, len(nodes)),
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
				summary := domain.NodeResourceSummary{
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
func (s *Service) CanAllocateResources(ctx context.Context, workspaceID string, request domain.ResourceRequest) (bool, error) {
	allocation, err := s.nodeRepo.GetWorkspaceAllocation(ctx, workspaceID)
	if err != nil {
		return false, fmt.Errorf("failed to get workspace allocation: %w", err)
	}

	return allocation.CanAllocate(request), nil
}

// GetNodeStatus returns detailed status information for a node
func (s *Service) GetNodeStatus(ctx context.Context, nodeID string) (*domain.NodeStatusInfo, error) {
	dedicatedNode, err := s.nodeRepo.GetDedicatedNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Get Proxmox VM status
	proxmoxStatus, err := s.proxmoxRepo.GetVMStatus(ctx, dedicatedNode.ProxmoxVMID)
	if err != nil {
		proxmoxStatus = "unknown"
	}

	// Get K3s agent status
	k3sStatus := "unknown"
	if s.k8sClient != nil {
		if status, err := s.CheckK3sAgentStatus(ctx, nodeID); err == nil {
			k3sStatus = status
		}
	}

	statusInfo := &domain.NodeStatusInfo{
		NodeID:        nodeID,
		Status:        dedicatedNode.Status,
		ProxmoxStatus: proxmoxStatus,
		K3sStatus:     k3sStatus,
		LastUpdated:   dedicatedNode.UpdatedAt,
		Conditions:    []domain.NodeCondition{},
	}

	// Get K3s conditions if available
	if s.k8sClient != nil {
		if conditions, err := s.GetK3sAgentConditions(ctx, nodeID); err == nil {
			statusInfo.Conditions = conditions
		}
	}

	// Add basic conditions if no K3s conditions are available
	if len(statusInfo.Conditions) == 0 && dedicatedNode.Status == domain.NodeStatusReady && proxmoxStatus == "running" {
		statusInfo.Conditions = append(statusInfo.Conditions, domain.NodeCondition{
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
func (s *Service) GetNodeMetrics(ctx context.Context, nodeID string) (*domain.NodeMetrics, error) {
	dedicatedNode, err := s.nodeRepo.GetDedicatedNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	vmUsage, err := s.proxmoxRepo.GetVMResourceUsage(ctx, dedicatedNode.ProxmoxVMID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM resource usage: %w", err)
	}

	metrics := &domain.NodeMetrics{
		NodeID:    nodeID,
		Timestamp: time.Now(),
		CPU: domain.CPUMetrics{
			UsagePercent: vmUsage.CPUUsage,
		},
		Memory: domain.MemoryMetrics{
			UsedBytes:    vmUsage.MemoryUsage,
			TotalBytes:   int64(dedicatedNode.Specification.MemoryGB) * 1024 * 1024 * 1024,
			UsagePercent: float64(vmUsage.MemoryUsage) / float64(dedicatedNode.Specification.MemoryGB*1024*1024*1024) * 100,
		},
		Disk: domain.DiskMetrics{
			UsedBytes:    vmUsage.DiskUsage,
			TotalBytes:   int64(dedicatedNode.Specification.StorageGB) * 1024 * 1024 * 1024,
			UsagePercent: float64(vmUsage.DiskUsage) / float64(dedicatedNode.Specification.StorageGB*1024*1024*1024) * 100,
		},
		Network: domain.NetworkMetrics{
			BytesIn:  vmUsage.NetworkIn,
			BytesOut: vmUsage.NetworkOut,
		},
	}

	metrics.Memory.AvailableBytes = metrics.Memory.TotalBytes - metrics.Memory.UsedBytes
	metrics.Disk.FreeBytes = metrics.Disk.TotalBytes - metrics.Disk.UsedBytes

	return metrics, nil
}

// GetNodeEvents returns events for a node
func (s *Service) GetNodeEvents(ctx context.Context, nodeID string, limit int) ([]domain.NodeEvent, error) {
	return s.nodeRepo.ListNodeEvents(ctx, nodeID, limit)
}

// GetNodeCosts calculates costs for nodes in a workspace
func (s *Service) GetNodeCosts(ctx context.Context, workspaceID string, period domain.BillingPeriod) (*domain.NodeCostReport, error) {
	nodes, err := s.nodeRepo.ListDedicatedNodes(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	report := &domain.NodeCostReport{
		WorkspaceID: workspaceID,
		Period:      period,
		Currency:    "USD",
		NodeCosts:   make([]domain.NodeCost, 0, len(nodes)),
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

	if allocation.PlanType == domain.PlanTypeShared {
		return nil // Already on shared plan
	}

	// Check if workspace has any active dedicated nodes
	nodes, err := s.nodeRepo.ListDedicatedNodes(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	activeNodes := 0
	for _, n := range nodes {
		if n.Status != domain.NodeStatusDeleted {
			activeNodes++
		}
	}

	if activeNodes > 0 {
		return errors.New("cannot transition to shared plan while dedicated nodes exist")
	}

	// Transition to shared plan
	allocation.PlanType = domain.PlanTypeShared
	allocation.SharedQuota = &domain.SharedQuota{
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

	if allocation.PlanType == domain.PlanTypeDedicated {
		return nil // Already on dedicated plan
	}

	allocation.PlanType = domain.PlanTypeDedicated
	allocation.SharedQuota = nil
	allocation.UpdatedAt = time.Now()

	return s.nodeRepo.UpdateWorkspaceAllocation(ctx, allocation)
}

// Helper methods

func (s *Service) validateProvisionRequest(req domain.ProvisionRequest) error {
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

func (s *Service) getNodeSpecification(nodeType string) (*domain.NodeSpecification, error) {
	specs := map[string]domain.NodeSpecification{
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
  - echo "workspace=%s node=%s" > /etc/hexabase/domain.conf

final_message: "Hexabase node is ready"
`, workspaceID, nodeID)
}

func (s *Service) extractIPAddress(vmInfo *domain.ProxmoxVMInfo) string {
	// In a real implementation, this would extract the IP from Proxmox
	// For now, return a placeholder
	return "10.0.0.100"
}

func (s *Service) performNodeAction(ctx context.Context, nodeID, action string, vmAction func(*domain.DedicatedNode) error) error {
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
	event := dedicatedNode.CreateEvent(domain.EventTypeStatusChange, fmt.Sprintf("Node %s initiated", action))
	s.nodeRepo.CreateNodeEvent(ctx, &event)

	// Perform VM action
	if err := vmAction(dedicatedNode); err != nil {
		// Revert status on failure
		switch action {
		case "start":
			dedicatedNode.Status = domain.NodeStatusStopped
		case "stop":
			dedicatedNode.Status = domain.NodeStatusReady
		}
		s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode)
		
		failEvent := dedicatedNode.CreateEvent(domain.EventTypeError, fmt.Sprintf("Node %s failed: %v", action, err))
		s.nodeRepo.CreateNodeEvent(ctx, &failEvent)
		
		return fmt.Errorf("VM %s failed: %w", action, err)
	}

	// Update to final status
	switch action {
	case "start":
		dedicatedNode.Status = domain.NodeStatusReady
	case "stop":
		dedicatedNode.Status = domain.NodeStatusStopped
	case "reboot":
		dedicatedNode.Status = domain.NodeStatusReady
	}

	dedicatedNode.UpdatedAt = time.Now()
	if err := s.nodeRepo.UpdateDedicatedNode(ctx, dedicatedNode); err != nil {
		return fmt.Errorf("failed to update final node status: %w", err)
	}

	// Create completion event
	completeEvent := dedicatedNode.CreateEvent(domain.EventTypeStatusChange, fmt.Sprintf("Node %s completed", action))
	s.nodeRepo.CreateNodeEvent(ctx, &completeEvent)

	return nil
}

func (s *Service) calculateNodeCost(n domain.DedicatedNode, period domain.BillingPeriod) domain.NodeCost {
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

	return domain.NodeCost{
		NodeID:      n.ID,
		NodeName:    n.Name,
		NodeType:    n.Specification.Type,
		HoursActive: hoursActive,
		HourlyRate:  hourlyRate,
		TotalCost:   totalCost,
		Status:      string(n.Status),
	}
}

// CheckK3sAgentStatus checks the status of K3s agent on a dedicated node
func (s *Service) CheckK3sAgentStatus(ctx context.Context, nodeID string) (string, error) {
	if s.k8sClient == nil {
		return "unknown", errors.New("kubernetes client not configured")
	}

	// Get the dedicated node
	dedicatedNode, err := s.nodeRepo.GetDedicatedNode(ctx, nodeID)
	if err != nil {
		return "", fmt.Errorf("failed to get dedicated node: %w", err)
	}

	// If node is not ready or provisioning, return appropriate status
	switch dedicatedNode.Status {
	case domain.NodeStatusProvisioning:
		return "provisioning", nil
	case domain.NodeStatusStopped, domain.NodeStatusFailed:
		return "stopped", nil
	}

	// List all nodes in the cluster
	nodes, err := s.k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("domain.hexabase.io/node-id=%s", nodeID),
	})
	if err != nil {
		return "unknown", fmt.Errorf("failed to list k8s nodes: %w", err)
	}

	// Check if node exists
	if len(nodes.Items) == 0 {
		// Try to find by name
		k8sNode, err := s.k8sClient.CoreV1().Nodes().Get(ctx, dedicatedNode.Name, metav1.GetOptions{})
		if err != nil {
			return "not_found", nil
		}
		nodes.Items = []corev1.Node{*k8sNode}
	}

	// Check node conditions
	for _, k8sNode := range nodes.Items {
		// Check if heartbeat is recent (within 5 minutes)
		for _, condition := range k8sNode.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				heartbeatAge := time.Since(condition.LastHeartbeatTime.Time)
				if heartbeatAge > 5*time.Minute {
					return "stale", nil
				}
				
				if condition.Status == corev1.ConditionTrue {
					return "ready", nil
				} else {
					return "not_ready", nil
				}
			}
		}
	}

	return "unknown", nil
}

// GetK3sAgentConditions returns the Kubernetes node conditions for a dedicated node
func (s *Service) GetK3sAgentConditions(ctx context.Context, nodeID string) ([]domain.NodeCondition, error) {
	if s.k8sClient == nil {
		return nil, errors.New("kubernetes client not configured")
	}

	// Get the dedicated node
	dedicatedNode, err := s.nodeRepo.GetDedicatedNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dedicated node: %w", err)
	}

	// List nodes with matching label
	nodes, err := s.k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("domain.hexabase.io/node-id=%s", nodeID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list k8s nodes: %w", err)
	}

	// If not found by label, try by name
	if len(nodes.Items) == 0 {
		k8sNode, err := s.k8sClient.CoreV1().Nodes().Get(ctx, dedicatedNode.Name, metav1.GetOptions{})
		if err == nil {
			nodes.Items = []corev1.Node{*k8sNode}
		} else {
			// Node not found, return empty conditions
			return []domain.NodeCondition{}, nil
		}
	}

	// Convert Kubernetes conditions to our domain model
	var conditions []domain.NodeCondition
	for _, k8sNode := range nodes.Items {
		for _, cond := range k8sNode.Status.Conditions {
			conditions = append(conditions, domain.NodeCondition{
				Type:    string(cond.Type),
				Status:  string(cond.Status),
				Reason:  cond.Reason,
				Message: cond.Message,
				Since:   cond.LastTransitionTime.Time,
			})
		}
		break // Only process the first node
	}

	return conditions, nil
}