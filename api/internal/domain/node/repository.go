package node

import (
	"context"
	"time"
)

// Repository defines the interface for node data operations
type Repository interface {
	// Node plan operations
	GetNodePlans(ctx context.Context) ([]NodePlan, error)
	GetNodePlan(ctx context.Context, planID string) (*NodePlan, error)
	
	// Dedicated node operations
	CreateDedicatedNode(ctx context.Context, node *DedicatedNode) error
	GetDedicatedNode(ctx context.Context, nodeID string) (*DedicatedNode, error)
	GetDedicatedNodeByVMID(ctx context.Context, vmid int) (*DedicatedNode, error)
	ListDedicatedNodes(ctx context.Context, workspaceID string) ([]DedicatedNode, error)
	UpdateDedicatedNode(ctx context.Context, node *DedicatedNode) error
	DeleteDedicatedNode(ctx context.Context, nodeID string) error
	
	// Node event operations
	CreateNodeEvent(ctx context.Context, event *NodeEvent) error
	ListNodeEvents(ctx context.Context, nodeID string, limit int) ([]NodeEvent, error)
	
	// Workspace allocation operations
	GetWorkspaceAllocation(ctx context.Context, workspaceID string) (*WorkspaceNodeAllocation, error)
	CreateWorkspaceAllocation(ctx context.Context, allocation *WorkspaceNodeAllocation) error
	UpdateWorkspaceAllocation(ctx context.Context, allocation *WorkspaceNodeAllocation) error
	
	// Resource usage operations
	UpdateSharedQuotaUsage(ctx context.Context, workspaceID string, cpu, memory float64) error
	GetNodeResourceUsage(ctx context.Context, nodeID string) (*ResourceUsage, error)
}

// ProxmoxRepository defines the interface for Proxmox VM operations
type ProxmoxRepository interface {
	// VM lifecycle operations
	CreateVM(ctx context.Context, spec VMSpec) (*ProxmoxVMInfo, error)
	GetVM(ctx context.Context, vmid int) (*ProxmoxVMInfo, error)
	StartVM(ctx context.Context, vmid int) error
	StopVM(ctx context.Context, vmid int) error
	RebootVM(ctx context.Context, vmid int) error
	DeleteVM(ctx context.Context, vmid int) error
	
	// VM configuration
	UpdateVMConfig(ctx context.Context, vmid int, config VMConfig) error
	GetVMStatus(ctx context.Context, vmid int) (string, error)
	
	// Cloud-init operations
	SetCloudInitConfig(ctx context.Context, vmid int, config CloudInitConfig) error
	
	// Resource monitoring
	GetVMResourceUsage(ctx context.Context, vmid int) (*VMResourceUsage, error)
	
	// Template operations
	CloneTemplate(ctx context.Context, templateID int, name string) (int, error)
	ListTemplates(ctx context.Context) ([]VMTemplate, error)
}

// VMSpec defines the specification for creating a new VM
type VMSpec struct {
	Name          string            `json:"name"`
	NodeType      string            `json:"node_type"`      // S-Type, M-Type, L-Type
	TemplateID    int               `json:"template_id"`
	TargetNode    string            `json:"target_node"`    // Proxmox node to create VM on
	CPUCores      int               `json:"cpu_cores"`
	MemoryMB      int               `json:"memory_mb"`
	DiskGB        int               `json:"disk_gb"`
	NetworkBridge string            `json:"network_bridge"`
	VLAN          int               `json:"vlan,omitempty"`
	CloudInit     CloudInitConfig   `json:"cloud_init"`
	Tags          []string          `json:"tags"`
}

// VMConfig defines configuration updates for a VM
type VMConfig struct {
	CPUCores int    `json:"cpu_cores,omitempty"`
	MemoryMB int    `json:"memory_mb,omitempty"`
	Name     string `json:"name,omitempty"`
}

// CloudInitConfig defines cloud-init configuration
type CloudInitConfig struct {
	UserData    string   `json:"user_data"`
	MetaData    string   `json:"meta_data"`
	NetworkData string   `json:"network_data"`
	SSHKeys     []string `json:"ssh_keys"`
	IPAddress   string   `json:"ip_address,omitempty"`
	Gateway     string   `json:"gateway,omitempty"`
	Nameservers []string `json:"nameservers,omitempty"`
}

// VMResourceUsage represents current resource usage of a VM
type VMResourceUsage struct {
	CPUUsage    float64 `json:"cpu_usage"`    // Percentage
	MemoryUsage int64   `json:"memory_usage"` // Bytes
	DiskUsage   int64   `json:"disk_usage"`   // Bytes
	NetworkIn   int64   `json:"network_in"`   // Bytes/sec
	NetworkOut  int64   `json:"network_out"`  // Bytes/sec
}

// VMTemplate represents a Proxmox VM template
type VMTemplate struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OSType      string `json:"os_type"`
	Version     string `json:"version"`
}

// ResourceUsage represents resource usage for a node
type ResourceUsage struct {
	NodeID      string    `json:"node_id"`
	CPUCores    float64   `json:"cpu_cores"`
	MemoryGB    float64   `json:"memory_gb"`
	StorageGB   float64   `json:"storage_gb"`
	PodCount    int       `json:"pod_count"`
	Timestamp   time.Time `json:"timestamp"`
}

