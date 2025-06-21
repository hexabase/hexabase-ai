package domain

import (
	"errors"
	"fmt"
	"time"
)

// NodeStatus represents the current state of a node
type NodeStatus string

// Plan types
const (
	PlanTypeShared    = "shared"
	PlanTypeDedicated = "dedicated"
)

// Node statuses
const (
	NodeStatusProvisioning NodeStatus = "provisioning"
	NodeStatusReady        NodeStatus = "ready"
	NodeStatusStarting     NodeStatus = "starting"
	NodeStatusStopping     NodeStatus = "stopping"
	NodeStatusStopped      NodeStatus = "stopped"
	NodeStatusDeleting     NodeStatus = "deleting"
	NodeStatusFailed       NodeStatus = "failed"
	NodeStatusDeleted      NodeStatus = "deleted"
)

// Event types
const (
	EventTypeProvisioning = "provisioning"
	EventTypeStatusChange = "status_change"
	EventTypeError        = "error"
	EventTypeDeletion     = "deletion"
)

// Errors
var (
	ErrInvalidPlanType       = errors.New("invalid plan type")
	ErrInvalidResourceSpec   = errors.New("invalid resource specification")
	ErrNegativePrice         = errors.New("price cannot be negative")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrInsufficientResources = errors.New("insufficient resources")
	ErrNodeNotReady          = errors.New("node is not ready")
)

// ResourceSpec defines resource limits for a plan
type ResourceSpec struct {
	CPUCores    int `json:"cpu_cores"`
	MemoryGB    int `json:"memory_gb"`
	StorageGB   int `json:"storage_gb"`
	MaxPods     int `json:"max_pods"`
	MaxServices int `json:"max_services"`
}

// NodePlan represents a workspace compute plan
type NodePlan struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Type          string       `json:"type"` // shared or dedicated
	Resources     ResourceSpec `json:"resources"`
	PricePerMonth float64      `json:"price_per_month"`
	Description   string       `json:"description"`
	Features      []string     `json:"features"`
}

// Validate checks if the node plan is valid
func (p *NodePlan) Validate() error {
	if p.Type != PlanTypeShared && p.Type != PlanTypeDedicated {
		return ErrInvalidPlanType
	}

	if p.Resources.CPUCores <= 0 || p.Resources.MemoryGB <= 0 || p.Resources.StorageGB <= 0 {
		return ErrInvalidResourceSpec
	}

	if p.PricePerMonth < 0 {
		return ErrNegativePrice
	}

	return nil
}

// NodeSpecification defines the hardware specs for a dedicated node
type NodeSpecification struct {
	Type        string `json:"type"` // S-Type, M-Type, L-Type
	CPUCores    int    `json:"cpu_cores"`
	MemoryGB    int    `json:"memory_gb"`
	StorageGB   int    `json:"storage_gb"`
	NetworkMbps int    `json:"network_mbps"`
}

// Validate checks if the node specification is valid
func (s NodeSpecification) Validate() error {
	if s.CPUCores <= 0 {
		return errors.New("CPU cores must be greater than 0")
	}
	if s.MemoryGB < 4 {
		return errors.New("memory must be at least 4GB")
	}
	if s.StorageGB < 50 {
		return errors.New("storage must be at least 50GB")
	}
	return nil
}

// ProxmoxVMInfo contains Proxmox-specific VM information
type ProxmoxVMInfo struct {
	VMID       int    `json:"vmid"`
	Node       string `json:"node"`        // Proxmox node name
	Status     string `json:"status"`      // Proxmox VM status
	Uptime     int64  `json:"uptime"`      // Uptime in seconds
	MaxMem     int64  `json:"maxmem"`      // Max memory in bytes
	Mem        int64  `json:"mem"`         // Current memory usage in bytes
	MaxDisk    int64  `json:"maxdisk"`     // Max disk in bytes
	Disk       int64  `json:"disk"`        // Current disk usage in bytes
	NetIn      int64  `json:"netin"`       // Network in bytes
	NetOut     int64  `json:"netout"`      // Network out bytes
	DiskRead   int64  `json:"diskread"`    // Disk read bytes
	DiskWrite  int64  `json:"diskwrite"`   // Disk write bytes
	CPU        float64 `json:"cpu"`        // CPU usage percentage
}

// IsRunning checks if the VM is in running state
func (v ProxmoxVMInfo) IsRunning() bool {
	return v.Status == "running"
}

// DedicatedNode represents a dedicated VM node for a workspace
type DedicatedNode struct {
	ID              string            `json:"id" gorm:"type:uuid;primary_key"`
	WorkspaceID     string            `json:"workspace_id" gorm:"type:uuid;index"`
	Name            string            `json:"name"`
	Status          NodeStatus        `json:"status"`
	Specification   NodeSpecification `json:"specification" gorm:"embedded;embeddedPrefix:spec_"`
	ProxmoxVMID     int               `json:"proxmox_vmid" gorm:"uniqueIndex"`
	ProxmoxNode     string            `json:"proxmox_node"`
	IPAddress       string            `json:"ip_address"`
	K3sAgentVersion string            `json:"k3s_agent_version"`
	SSHPublicKey    string            `json:"ssh_public_key" gorm:"type:text"`
	Labels          map[string]string `json:"labels" gorm:"serializer:json"`
	Taints          []string          `json:"taints" gorm:"serializer:json"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	DeletedAt       *time.Time        `json:"deleted_at,omitempty" gorm:"index"`
}

// CanScheduleWorkload checks if the node can accept new workloads
func (n *DedicatedNode) CanScheduleWorkload() bool {
	return n.Status == NodeStatusReady
}

// TransitionTo transitions the node to a new state based on action
func (n *DedicatedNode) TransitionTo(action string) error {
	transitions := map[NodeStatus]map[string]NodeStatus{
		"": {
			"provision": NodeStatusProvisioning,
		},
		NodeStatusProvisioning: {
			"complete": NodeStatusReady,
			"fail":     NodeStatusFailed,
		},
		NodeStatusReady: {
			"stop":   NodeStatusStopping,
			"delete": NodeStatusDeleting,
		},
		NodeStatusStopping: {
			"complete": NodeStatusStopped,
			"fail":     NodeStatusReady,
		},
		NodeStatusStopped: {
			"start":  NodeStatusStarting,
			"delete": NodeStatusDeleting,
		},
		NodeStatusStarting: {
			"complete": NodeStatusReady,
			"fail":     NodeStatusStopped,
		},
		NodeStatusDeleting: {
			"complete": NodeStatusDeleted,
			"fail":     NodeStatusFailed,
		},
	}

	validTransitions, ok := transitions[n.Status]
	if !ok {
		return fmt.Errorf("%w: no transitions from status %s", ErrInvalidStateTransition, n.Status)
	}

	newStatus, ok := validTransitions[action]
	if !ok {
		return fmt.Errorf("%w: cannot %s from status %s", ErrInvalidStateTransition, action, n.Status)
	}

	n.Status = newStatus
	n.UpdatedAt = time.Now()
	return nil
}

// CreateEvent creates a new node event
func (n *DedicatedNode) CreateEvent(eventType, message string) NodeEvent {
	return NodeEvent{
		NodeID:      n.ID,
		WorkspaceID: n.WorkspaceID,
		Type:        eventType,
		Message:     message,
		Timestamp:   time.Now(),
	}
}

// CalculateUsageCost calculates the cost of the node up to the given time
func (n *DedicatedNode) CalculateUsageCost(until time.Time) float64 {
	if n.Status == NodeStatusDeleted || n.DeletedAt != nil {
		return 0
	}

	// Get hourly rate based on node type
	hourlyRate := getHourlyRate(n.Specification.Type)
	
	// Calculate hours used
	duration := until.Sub(n.CreatedAt)
	hours := duration.Hours()
	
	return hourlyRate * hours
}

// getHourlyRate returns the hourly rate for a node type
func getHourlyRate(nodeType string) float64 {
	rates := map[string]float64{
		"S-Type": 99.99 / (30 * 24),  // $99.99/month
		"M-Type": 199.99 / (30 * 24), // $199.99/month
		"L-Type": 399.99 / (30 * 24), // $399.99/month
	}
	
	if rate, ok := rates[nodeType]; ok {
		return rate
	}
	return 0
}

// NodeEvent represents an event in the node lifecycle
type NodeEvent struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key"`
	NodeID      string    `json:"node_id" gorm:"type:uuid;index"`
	WorkspaceID string    `json:"workspace_id" gorm:"type:uuid;index"`
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Details     string    `json:"details" gorm:"type:text"`
	Timestamp   time.Time `json:"timestamp"`
}

// ResourceRequest represents a request for resources
type ResourceRequest struct {
	CPU    float64 `json:"cpu"`    // CPU cores
	Memory float64 `json:"memory"` // Memory in GB
}

// SharedQuota represents resource quota for shared plan workspaces
type SharedQuota struct {
	CPULimit    float64 `json:"cpu_limit"`
	MemoryLimit float64 `json:"memory_limit"`
	CPUUsed     float64 `json:"cpu_used"`
	MemoryUsed  float64 `json:"memory_used"`
}

// WorkspaceNodeAllocation tracks node allocation for a workspace
type WorkspaceNodeAllocation struct {
	ID            string       `json:"id" gorm:"type:uuid;primary_key"`
	WorkspaceID   string       `json:"workspace_id" gorm:"type:uuid;uniqueIndex"`
	PlanType      string       `json:"plan_type"`
	SharedQuota   *SharedQuota `json:"shared_quota,omitempty" gorm:"embedded;embeddedPrefix:quota_"`
	DedicatedNodes []DedicatedNode `json:"dedicated_nodes,omitempty" gorm:"foreignKey:WorkspaceID"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// CanAllocate checks if the workspace can allocate the requested resources
func (w *WorkspaceNodeAllocation) CanAllocate(requested ResourceRequest) bool {
	if w.PlanType == PlanTypeShared && w.SharedQuota != nil {
		remainingCPU := w.SharedQuota.CPULimit - w.SharedQuota.CPUUsed
		remainingMemory := w.SharedQuota.MemoryLimit - w.SharedQuota.MemoryUsed
		
		return requested.CPU <= remainingCPU && requested.Memory <= remainingMemory
	}
	
	// For dedicated plans, check if any node can handle the request
	if w.PlanType == PlanTypeDedicated {
		for _, node := range w.DedicatedNodes {
			if node.CanScheduleWorkload() {
				return true
			}
		}
	}
	
	return false
}

// NodePool represents a group of nodes
type NodePool struct {
	Name  string          `json:"name"`
	Nodes []DedicatedNode `json:"nodes"`
}

// GetAvailableNodes returns nodes that can accept workloads
func (p *NodePool) GetAvailableNodes() []DedicatedNode {
	var available []DedicatedNode
	for _, node := range p.Nodes {
		if node.CanScheduleWorkload() {
			available = append(available, node)
		}
	}
	return available
}

// ParseTime parses a time string in RFC3339 format
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

// CurrentMonthPeriod returns a billing period for the current month
func CurrentMonthPeriod() BillingPeriod {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	
	return BillingPeriod{
		Start: start,
		End:   end,
	}
}