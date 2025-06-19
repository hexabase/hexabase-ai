package domain

import (
	"context"
	"time"
)

// Repository defines the data access interface for workspaces
type Repository interface {
	// Workspace operations
	CreateWorkspace(ctx context.Context, workspace *Workspace) error
	GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
	GetWorkspaceByNameAndOrg(ctx context.Context, name, orgID string) (*Workspace, error)
	ListWorkspaces(ctx context.Context, filter WorkspaceFilter) ([]*Workspace, int, error)
	UpdateWorkspace(ctx context.Context, workspace *Workspace) error
	DeleteWorkspace(ctx context.Context, workspaceID string) error
	
	// Task operations
	CreateTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, taskID string) (*Task, error)
	ListTasks(ctx context.Context, workspaceID string) ([]*Task, error)
	UpdateTask(ctx context.Context, task *Task) error
	GetPendingTasks(ctx context.Context, taskType string, limit int) ([]*Task, error)
	
	// Status operations
	SaveWorkspaceStatus(ctx context.Context, status *WorkspaceStatus) error
	GetWorkspaceStatus(ctx context.Context, workspaceID string) (*WorkspaceStatus, error)
	
	// Kubeconfig operations
	SaveKubeconfig(ctx context.Context, workspaceID, kubeconfig string) error
	GetKubeconfig(ctx context.Context, workspaceID string) (string, error)
	
	// Cleanup operations
	CleanupExpiredTasks(ctx context.Context, before time.Time) error
	CleanupDeletedWorkspaces(ctx context.Context, before time.Time) error
	
	// Member operations
	ListWorkspaceMembers(ctx context.Context, workspaceID string) ([]*WorkspaceMember, error)
	AddWorkspaceMember(ctx context.Context, member *WorkspaceMember) error
	RemoveWorkspaceMember(ctx context.Context, workspaceID, userID string) error
	
	// Resource usage operations
	CreateResourceUsage(ctx context.Context, usage *ResourceUsage) error
}

// VClusterRepository defines the interface for vCluster operations
type VClusterRepository interface {
	// CreateVCluster creates a new vCluster
	CreateVCluster(ctx context.Context, name, namespace string, values map[string]interface{}) error
	
	// DeleteVCluster deletes a vCluster
	DeleteVCluster(ctx context.Context, name, namespace string) error
	
	// GetVClusterStatus gets the status of a vCluster
	GetVClusterStatus(ctx context.Context, name, namespace string) (map[string]interface{}, error)
	
	// GetVClusterKubeconfig retrieves the kubeconfig for a vCluster
	GetVClusterKubeconfig(ctx context.Context, name, namespace string) (string, error)
	
	// ListVClusters lists all vClusters
	ListVClusters(ctx context.Context) ([]VClusterInfo, error)
	
	// UpgradeVCluster upgrades a vCluster to a new version
	UpgradeVCluster(ctx context.Context, name, namespace, version string) error
}

// VClusterInfo represents basic vCluster information
type VClusterInfo struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	Created   time.Time              `json:"created"`
	Labels    map[string]string      `json:"labels"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// KubernetesRepository defines the interface for Kubernetes operations
type KubernetesRepository interface {
	CreateVCluster(ctx context.Context, workspaceID, plan string) error
	DeleteVCluster(ctx context.Context, workspaceID string) error
	WaitForVClusterReady(ctx context.Context, workspaceID string) error
	WaitForVClusterDeleted(ctx context.Context, workspaceID string) error
	GetVClusterStatus(ctx context.Context, workspaceID string) (string, error)
	GetVClusterInfo(ctx context.Context, workspaceID string) (*ClusterInfo, error)
	ScaleVCluster(ctx context.Context, workspaceID string, replicas int) error
	ConfigureOIDC(ctx context.Context, workspaceID string) error
	UpdateOIDCConfig(ctx context.Context, workspaceID string, config map[string]interface{}) error
	ApplyResourceQuotas(ctx context.Context, workspaceID, plan string) error
	GetResourceMetrics(ctx context.Context, workspaceID string) (*ResourceUsage, error)
	ListVClusterNodes(ctx context.Context, workspaceID string) ([]Node, error)
	ScaleVClusterDeployment(ctx context.Context, workspaceID, deploymentName string, replicas int) error
}

// AuthRepository defines an adapter for auth-related operations needed by the workspace domain
type AuthRepository interface {
	// GetUser retrieves user information
	GetUser(ctx context.Context, userID string) (*User, error)
	
	// GenerateWorkspaceToken generates a token for workspace access
	GenerateWorkspaceToken(ctx context.Context, userID, workspaceID string) (string, error)
}