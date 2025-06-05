package workspace

import (
	"time"
)

// Workspace represents a vCluster workspace
type Workspace struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OrganizationID string    `json:"organization_id"`
	PlanID         string    `json:"plan_id"`
	Status         string    `json:"status"` // provisioning, active, suspended, terminating, terminated
	VClusterName   string    `json:"vcluster_name"`
	Namespace      string    `json:"namespace"`
	KubeConfig     string    `json:"kubeconfig,omitempty"`
	APIEndpoint    string    `json:"api_endpoint,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// CreateWorkspaceRequest represents a request to create a workspace
type CreateWorkspaceRequest struct {
	Name           string            `json:"name" binding:"required"`
	OrganizationID string            `json:"organization_id" binding:"required"`
	PlanID         string            `json:"plan_id" binding:"required"`
	KubernetesVersion string         `json:"kubernetes_version,omitempty"`
	NodeCount      int               `json:"node_count,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// UpdateWorkspaceRequest represents a request to update a workspace
type UpdateWorkspaceRequest struct {
	Name      string            `json:"name,omitempty"`
	PlanID    string            `json:"plan_id,omitempty"`
	NodeCount int               `json:"node_count,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// WorkspaceList represents a list of workspaces
type WorkspaceList struct {
	Workspaces []*Workspace `json:"workspaces"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
}

// WorkspaceStatus represents the detailed status of a workspace
type WorkspaceStatus struct {
	WorkspaceID   string                 `json:"workspace_id"`
	Status        string                 `json:"status"`
	Healthy       bool                   `json:"healthy"`
	Message       string                 `json:"message,omitempty"`
	ClusterInfo   map[string]interface{} `json:"cluster_info,omitempty"`
	ResourceUsage *ResourceUsage         `json:"resource_usage,omitempty"`
	LastChecked   time.Time              `json:"last_checked"`
}

// ResourceUsage represents resource usage for a workspace
type ResourceUsage struct {
	CPU    ResourceMetric `json:"cpu"`
	Memory ResourceMetric `json:"memory"`
	Storage ResourceMetric `json:"storage"`
	Pods   PodMetric      `json:"pods"`
}

// ResourceMetric represents a resource metric
type ResourceMetric struct {
	Used      float64 `json:"used"`
	Requested float64 `json:"requested"`
	Limit     float64 `json:"limit"`
	Unit      string  `json:"unit"`
}

// PodMetric represents pod count metrics
type PodMetric struct {
	Running int `json:"running"`
	Pending int `json:"pending"`
	Failed  int `json:"failed"`
	Total   int `json:"total"`
}

// Task represents an async task for workspace operations
type Task struct {
	ID          string                 `json:"id"`
	WorkspaceID string                 `json:"workspace_id"`
	Type        string                 `json:"type"` // create, update, delete, backup, restore
	Status      string                 `json:"status"` // pending, running, completed, failed
	Progress    int                    `json:"progress"`
	Message     string                 `json:"message,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// WorkspaceFilter represents filter options for listing workspaces
type WorkspaceFilter struct {
	OrganizationID string
	Status         string
	PlanID         string
	Page           int
	PageSize       int
	SortBy         string
	SortOrder      string
}

// WorkspaceOperationRequest represents a request for workspace operations
type WorkspaceOperationRequest struct {
	Operation string                 `json:"operation" binding:"required,oneof=backup restore upgrade"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BackupRequest represents a backup request
type BackupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}

// RestoreRequest represents a restore request
type RestoreRequest struct {
	BackupID string `json:"backup_id" binding:"required"`
}

// UpgradeRequest represents an upgrade request
type UpgradeRequest struct {
	TargetVersion string `json:"target_version" binding:"required"`
	NodeCount     int    `json:"node_count,omitempty"`
}