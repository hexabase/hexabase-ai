package db

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NodePlan represents available node plans
type NodePlan struct {
	ID            string  `gorm:"primaryKey" json:"id"`
	Name          string  `gorm:"not null" json:"name"`
	Type          string  `gorm:"not null;index" json:"type"` // shared or dedicated
	Description   string  `gorm:"type:text" json:"description"`
	PricePerMonth float64 `gorm:"not null" json:"price_per_month"`
	
	// Resource specifications stored as JSON
	ResourcesJSON string       `gorm:"column:resources;type:jsonb" json:"-"`
	Resources     ResourceSpec `gorm:"-" json:"resources"`
	
	// Features stored as JSON array
	FeaturesJSON string   `gorm:"column:features;type:jsonb" json:"-"`
	Features     []string `gorm:"-" json:"features"`
	
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResourceSpec represents resource limits for a plan
type ResourceSpec struct {
	CPUCores    int `json:"cpu_cores"`
	MemoryGB    int `json:"memory_gb"`
	StorageGB   int `json:"storage_gb"`
	MaxPods     int `json:"max_pods"`
	MaxServices int `json:"max_services"`
}

// DedicatedNode represents a dedicated VM node for a workspace
type DedicatedNode struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"workspace_id"`
	Name        string    `gorm:"not null" json:"name"`
	Status      string    `gorm:"not null;index" json:"status"`
	
	// Node specification stored as JSON
	SpecificationJSON string             `gorm:"column:specification;type:jsonb" json:"-"`
	Specification     NodeSpecification  `gorm:"-" json:"specification"`
	
	// Proxmox VM information
	ProxmoxVMID int    `gorm:"uniqueIndex" json:"proxmox_vmid"`
	ProxmoxNode string `json:"proxmox_node"`
	IPAddress   string `json:"ip_address"`
	
	// K3s agent information
	K3sAgentVersion string `json:"k3s_agent_version"`
	SSHPublicKey    string `gorm:"type:text" json:"ssh_public_key"`
	
	// Labels and taints stored as JSON
	LabelsJSON string            `gorm:"column:labels;type:jsonb" json:"-"`
	Labels     map[string]string `gorm:"-" json:"labels"`
	
	TaintsJSON string   `gorm:"column:taints;type:jsonb" json:"-"`
	Taints     []string `gorm:"-" json:"taints"`
	
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	
	// Association
	Workspace Workspace `json:"workspace,omitempty"`
}

// NodeSpecification defines the hardware specs for a dedicated node
type NodeSpecification struct {
	Type        string `json:"type"` // S-Type, M-Type, L-Type
	CPUCores    int    `json:"cpu_cores"`
	MemoryGB    int    `json:"memory_gb"`
	StorageGB   int    `json:"storage_gb"`
	NetworkMbps int    `json:"network_mbps"`
}

// NodeEvent represents an event in the node lifecycle
type NodeEvent struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	NodeID      string    `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"node_id"`
	WorkspaceID string    `gorm:"not null;index" json:"workspace_id"`
	Type        string    `gorm:"not null;index" json:"type"`
	Message     string    `gorm:"not null" json:"message"`
	Details     string    `gorm:"type:text" json:"details"`
	Timestamp   time.Time `gorm:"index" json:"timestamp"`
	
	// Association
	Node DedicatedNode `json:"node,omitempty"`
}

// WorkspaceNodeAllocation tracks node allocation for a workspace
type WorkspaceNodeAllocation struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE" json:"workspace_id"`
	PlanType    string    `gorm:"not null" json:"plan_type"` // shared or dedicated
	
	// Shared quota stored as JSON (only for shared plan)
	SharedQuotaJSON string       `gorm:"column:shared_quota;type:jsonb" json:"-"`
	SharedQuota     *SharedQuota `gorm:"-" json:"shared_quota,omitempty"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Associations
	Workspace      Workspace       `json:"workspace,omitempty"`
	DedicatedNodes []DedicatedNode `gorm:"foreignKey:WorkspaceID;references:WorkspaceID" json:"dedicated_nodes,omitempty"`
}

// SharedQuota represents resource quota for shared plan workspaces
type SharedQuota struct {
	CPULimit    float64 `json:"cpu_limit"`
	MemoryLimit float64 `json:"memory_limit"`
	CPUUsed     float64 `json:"cpu_used"`
	MemoryUsed  float64 `json:"memory_used"`
}

// BeforeCreate hooks for UUID generation
func (np *NodePlan) BeforeCreate(tx *gorm.DB) error {
	if np.ID == "" {
		np.ID = "plan-" + uuid.New().String()
	}
	return np.serializeFields()
}

func (dn *DedicatedNode) BeforeCreate(tx *gorm.DB) error {
	if dn.ID == "" {
		dn.ID = "node-" + uuid.New().String()
	}
	return dn.serializeFields()
}

func (ne *NodeEvent) BeforeCreate(tx *gorm.DB) error {
	if ne.ID == "" {
		ne.ID = "event-" + uuid.New().String()
	}
	return nil
}

func (wna *WorkspaceNodeAllocation) BeforeCreate(tx *gorm.DB) error {
	if wna.ID == "" {
		wna.ID = "alloc-" + uuid.New().String()
	}
	return wna.serializeFields()
}

// Serialization/deserialization methods for JSON fields
func (np *NodePlan) BeforeSave(tx *gorm.DB) error {
	return np.serializeFields()
}

func (np *NodePlan) AfterFind(tx *gorm.DB) error {
	return np.deserializeFields()
}

func (np *NodePlan) serializeFields() error {
	// Serialize Resources
	if data, err := json.Marshal(np.Resources); err != nil {
		return err
	} else {
		np.ResourcesJSON = string(data)
	}
	
	// Serialize Features
	if data, err := json.Marshal(np.Features); err != nil {
		return err
	} else {
		np.FeaturesJSON = string(data)
	}
	
	return nil
}

func (np *NodePlan) deserializeFields() error {
	// Deserialize Resources
	if np.ResourcesJSON != "" {
		if err := json.Unmarshal([]byte(np.ResourcesJSON), &np.Resources); err != nil {
			return err
		}
	}
	
	// Deserialize Features
	if np.FeaturesJSON != "" {
		if err := json.Unmarshal([]byte(np.FeaturesJSON), &np.Features); err != nil {
			return err
		}
	}
	
	return nil
}

func (dn *DedicatedNode) BeforeSave(tx *gorm.DB) error {
	return dn.serializeFields()
}

func (dn *DedicatedNode) AfterFind(tx *gorm.DB) error {
	return dn.deserializeFields()
}

func (dn *DedicatedNode) serializeFields() error {
	// Serialize Specification
	if data, err := json.Marshal(dn.Specification); err != nil {
		return err
	} else {
		dn.SpecificationJSON = string(data)
	}
	
	// Serialize Labels
	if data, err := json.Marshal(dn.Labels); err != nil {
		return err
	} else {
		dn.LabelsJSON = string(data)
	}
	
	// Serialize Taints
	if data, err := json.Marshal(dn.Taints); err != nil {
		return err
	} else {
		dn.TaintsJSON = string(data)
	}
	
	return nil
}

func (dn *DedicatedNode) deserializeFields() error {
	// Deserialize Specification
	if dn.SpecificationJSON != "" {
		if err := json.Unmarshal([]byte(dn.SpecificationJSON), &dn.Specification); err != nil {
			return err
		}
	}
	
	// Deserialize Labels
	if dn.LabelsJSON != "" {
		if err := json.Unmarshal([]byte(dn.LabelsJSON), &dn.Labels); err != nil {
			return err
		}
	}
	
	// Deserialize Taints
	if dn.TaintsJSON != "" {
		if err := json.Unmarshal([]byte(dn.TaintsJSON), &dn.Taints); err != nil {
			return err
		}
	}
	
	return nil
}

func (wna *WorkspaceNodeAllocation) BeforeSave(tx *gorm.DB) error {
	return wna.serializeFields()
}

func (wna *WorkspaceNodeAllocation) AfterFind(tx *gorm.DB) error {
	return wna.deserializeFields()
}

func (wna *WorkspaceNodeAllocation) serializeFields() error {
	// Serialize SharedQuota
	if wna.SharedQuota != nil {
		if data, err := json.Marshal(wna.SharedQuota); err != nil {
			return err
		} else {
			wna.SharedQuotaJSON = string(data)
		}
	}
	
	return nil
}

func (wna *WorkspaceNodeAllocation) deserializeFields() error {
	// Deserialize SharedQuota
	if wna.SharedQuotaJSON != "" {
		wna.SharedQuota = &SharedQuota{}
		if err := json.Unmarshal([]byte(wna.SharedQuotaJSON), wna.SharedQuota); err != nil {
			return err
		}
	}
	
	return nil
}