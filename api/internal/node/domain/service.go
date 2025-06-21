package domain

import (
	"context"
	"time"
)

// Service defines the business logic interface for node management
type Service interface {
	// Plan operations
	GetAvailablePlans(ctx context.Context) ([]NodePlan, error)
	GetPlanDetails(ctx context.Context, planID string) (*NodePlan, error)
	
	// Node provisioning operations
	ProvisionDedicatedNode(ctx context.Context, workspaceID string, spec ProvisionRequest) (*DedicatedNode, error)
	GetNode(ctx context.Context, nodeID string) (*DedicatedNode, error)
	ListNodes(ctx context.Context, workspaceID string) ([]DedicatedNode, error)
	
	// Node lifecycle operations
	StartNode(ctx context.Context, nodeID string) error
	StopNode(ctx context.Context, nodeID string) error
	RebootNode(ctx context.Context, nodeID string) error
	DeleteNode(ctx context.Context, nodeID string) error
	
	// Resource management
	GetWorkspaceResourceUsage(ctx context.Context, workspaceID string) (*WorkspaceResourceUsage, error)
	CanAllocateResources(ctx context.Context, workspaceID string, request ResourceRequest) (bool, error)
	
	// Node monitoring
	GetNodeStatus(ctx context.Context, nodeID string) (*NodeStatusInfo, error)
	GetNodeMetrics(ctx context.Context, nodeID string) (*NodeMetrics, error)
	GetNodeEvents(ctx context.Context, nodeID string, limit int) ([]NodeEvent, error)
	
	// Billing operations
	GetNodeCosts(ctx context.Context, workspaceID string, period BillingPeriod) (*NodeCostReport, error)
	
	// Plan transitions
	TransitionToSharedPlan(ctx context.Context, workspaceID string) error
	TransitionToDedicatedPlan(ctx context.Context, workspaceID string) error
}

// ProvisionRequest contains the parameters for provisioning a new dedicated node
type ProvisionRequest struct {
	NodeName      string `json:"node_name"`
	NodeType      string `json:"node_type"`      // S-Type, M-Type, L-Type
	Region        string `json:"region,omitempty"`
	SSHPublicKey  string `json:"ssh_public_key,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// WorkspaceResourceUsage represents the resource usage for a workspace
type WorkspaceResourceUsage struct {
	WorkspaceID   string                 `json:"workspace_id"`
	PlanType      string                 `json:"plan_type"`
	SharedUsage   *SharedResourceUsage   `json:"shared_usage,omitempty"`
	DedicatedUsage *DedicatedResourceUsage `json:"dedicated_usage,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// SharedResourceUsage represents resource usage on shared infrastructure
type SharedResourceUsage struct {
	CPUUsed       float64 `json:"cpu_used"`
	CPULimit      float64 `json:"cpu_limit"`
	MemoryUsedGB  float64 `json:"memory_used_gb"`
	MemoryLimitGB float64 `json:"memory_limit_gb"`
	PodCount      int     `json:"pod_count"`
	PodLimit      int     `json:"pod_limit"`
}

// DedicatedResourceUsage represents resource usage on dedicated nodes
type DedicatedResourceUsage struct {
	TotalNodes     int                    `json:"total_nodes"`
	ActiveNodes    int                    `json:"active_nodes"`
	TotalCPUCores  int                    `json:"total_cpu_cores"`
	TotalMemoryGB  int                    `json:"total_memory_gb"`
	TotalStorageGB int                    `json:"total_storage_gb"`
	NodeUsage      []NodeResourceSummary  `json:"node_usage"`
}

// NodeResourceSummary represents resource usage for a single node
type NodeResourceSummary struct {
	NodeID        string  `json:"node_id"`
	NodeName      string  `json:"node_name"`
	CPUUsage      float64 `json:"cpu_usage"`      // Percentage
	MemoryUsage   float64 `json:"memory_usage"`   // Percentage
	StorageUsage  float64 `json:"storage_usage"`  // Percentage
	PodCount      int     `json:"pod_count"`
	Status        string  `json:"status"`
}

// NodeStatusInfo contains detailed status information for a node
type NodeStatusInfo struct {
	NodeID        string            `json:"node_id"`
	Status        NodeStatus        `json:"status"`
	ProxmoxStatus string            `json:"proxmox_status"`
	K3sStatus     string            `json:"k3s_status"`
	Conditions    []NodeCondition   `json:"conditions"`
	LastUpdated   time.Time         `json:"last_updated"`
}

// NodeCondition represents a condition affecting the node
type NodeCondition struct {
	Type    string    `json:"type"`    // Ready, MemoryPressure, DiskPressure, etc.
	Status  string    `json:"status"`  // True, False, Unknown
	Reason  string    `json:"reason"`
	Message string    `json:"message"`
	Since   time.Time `json:"since"`
}

// NodeMetrics contains performance metrics for a node
type NodeMetrics struct {
	NodeID    string              `json:"node_id"`
	Timestamp time.Time           `json:"timestamp"`
	CPU       CPUMetrics          `json:"cpu"`
	Memory    MemoryMetrics       `json:"memory"`
	Disk      DiskMetrics         `json:"disk"`
	Network   NetworkMetrics      `json:"network"`
}

// CPUMetrics represents CPU usage metrics
type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	LoadAverage  []float64 `json:"load_average"` // 1m, 5m, 15m
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	TotalBytes     int64   `json:"total_bytes"`
	UsedBytes      int64   `json:"used_bytes"`
	AvailableBytes int64   `json:"available_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
}

// DiskMetrics represents disk usage metrics
type DiskMetrics struct {
	TotalBytes   int64   `json:"total_bytes"`
	UsedBytes    int64   `json:"used_bytes"`
	FreeBytes    int64   `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	IOPSRead     int64   `json:"iops_read"`
	IOPSWrite    int64   `json:"iops_write"`
}

// NetworkMetrics represents network usage metrics
type NetworkMetrics struct {
	BytesIn     int64   `json:"bytes_in"`
	BytesOut    int64   `json:"bytes_out"`
	PacketsIn   int64   `json:"packets_in"`
	PacketsOut  int64   `json:"packets_out"`
	ErrorsIn    int64   `json:"errors_in"`
	ErrorsOut   int64   `json:"errors_out"`
}

// BillingPeriod represents a billing time period
type BillingPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// NodeCostReport contains cost information for nodes in a workspace
type NodeCostReport struct {
	WorkspaceID  string         `json:"workspace_id"`
	Period       BillingPeriod  `json:"period"`
	TotalCost    float64        `json:"total_cost"`
	Currency     string         `json:"currency"`
	NodeCosts    []NodeCost     `json:"node_costs"`
	SharedCost   *float64       `json:"shared_cost,omitempty"`
}

// NodeCost represents the cost for a single node
type NodeCost struct {
	NodeID       string    `json:"node_id"`
	NodeName     string    `json:"node_name"`
	NodeType     string    `json:"node_type"`
	HoursActive  float64   `json:"hours_active"`
	HourlyRate   float64   `json:"hourly_rate"`
	TotalCost    float64   `json:"total_cost"`
	Status       string    `json:"status"`
}