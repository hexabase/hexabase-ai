package workspace

import (
	"context"
	"time"
)

// Workspace represents a vCluster workspace
type Workspace struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	OrganizationID string                 `json:"organization_id"`
	Plan           string                 `json:"plan"`
	PlanID         string                 `json:"plan_id"`
	Status         string                 `json:"status"` // provisioning, active, suspended, terminating, terminated
	VClusterName   string                 `json:"vcluster_name"`
	Namespace      string                 `json:"namespace"`
	KubeConfig     string                 `json:"kubeconfig,omitempty"`
	APIEndpoint    string                 `json:"api_endpoint,omitempty"`
	ClusterInfo    map[string]interface{} `json:"cluster_info,omitempty"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
}

// CreateWorkspaceRequest represents a request to create a workspace
type CreateWorkspaceRequest struct {
	Name              string                 `json:"name" binding:"required"`
	Description       string                 `json:"description,omitempty"`
	OrganizationID    string                 `json:"organization_id" binding:"required"`
	Plan              string                 `json:"plan" binding:"required"`
	PlanID            string                 `json:"plan_id" binding:"required"`
	KubernetesVersion string                 `json:"kubernetes_version,omitempty"`
	NodeCount         int                    `json:"node_count,omitempty"`
	Labels            map[string]string      `json:"labels,omitempty"`
	Settings          map[string]interface{} `json:"settings,omitempty"`
	CreatedBy         string                 `json:"created_by,omitempty"`
}

// UpdateWorkspaceRequest represents a request to update a workspace
type UpdateWorkspaceRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	PlanID      string                 `json:"plan_id,omitempty"`
	NodeCount   int                    `json:"node_count,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	UpdatedBy   string                 `json:"updated_by,omitempty"`
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
	WorkspaceID string         `json:"workspace_id"`
	CPU         ResourceMetric `json:"cpu"`
	Memory      ResourceMetric `json:"memory"`
	Storage     ResourceMetric `json:"storage"`
	Pods        PodMetric      `json:"pods"`
	Timestamp   time.Time      `json:"timestamp"`
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
	Payload     map[string]interface{} `json:"payload,omitempty"`
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
	Search         string
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

// WorkspaceMember represents a member of a workspace
type WorkspaceMember struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	AddedBy     string    `json:"added_by"`
	AddedAt     time.Time `json:"added_at"`
}

// AddMemberRequest represents a request to add a member to a workspace
type AddMemberRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	Role    string `json:"role" binding:"required,oneof=viewer editor admin"`
	AddedBy string `json:"added_by,omitempty"`
}

// ClusterInfo represents vCluster information
type ClusterInfo struct {
	Endpoint   string `json:"endpoint"`
	APIServer  string `json:"api_server"`
	KubeConfig string `json:"kubeconfig"`
	Status     string `json:"status"`
}

// Define repository interfaces to be used by other packages
type KubernetesRepository interface {
	CreateVCluster(ctx context.Context, workspaceID string, plan string) error
	DeleteVCluster(ctx context.Context, workspaceID string) error
	GetVClusterStatus(ctx context.Context, workspaceID string) (string, error)
	ScaleVCluster(ctx context.Context, workspaceID string, nodeCount int) error
	GetVClusterInfo(ctx context.Context, workspaceID string) (*ClusterInfo, error)
	GetResourceMetrics(ctx context.Context, workspaceID string) (*ResourceUsage, error)
	
	// Additional operations
	UpdateOIDCConfig(ctx context.Context, workspaceID string, config map[string]interface{}) error
	WaitForVClusterReady(ctx context.Context, workspaceID string) error
	WaitForVClusterDeleted(ctx context.Context, workspaceID string) error
	ConfigureOIDC(ctx context.Context, workspaceID string) error
	ApplyResourceQuotas(ctx context.Context, workspaceID, plan string) error
	GetVClusterKubeconfig(ctx context.Context, workspaceID string) (string, error)
}

type AuthRepository interface {
	GetUser(ctx context.Context, userID string) (*User, error)
	GenerateWorkspaceToken(ctx context.Context, userID, workspaceID string) (string, error)
}

// User represents a minimal user for workspace context
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}