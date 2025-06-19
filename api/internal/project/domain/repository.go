package domain

import (
	"context"
	"time"
)

// Repository defines the data access interface for projects
type Repository interface {
	// Project operations
	CreateProject(ctx context.Context, project *Project) error
	GetProject(ctx context.Context, projectID string) (*Project, error)
	GetProjectByName(ctx context.Context, name string) (*Project, error)
	GetProjectByNameAndWorkspace(ctx context.Context, name, workspaceID string) (*Project, error)
	ListProjects(ctx context.Context, filter ProjectFilter) ([]*Project, int, error)
	UpdateProject(ctx context.Context, project *Project) error
	DeleteProject(ctx context.Context, projectID string) error
	CountProjects(ctx context.Context, workspaceID string) (int, error)
	
	// Namespace operations
	CreateNamespace(ctx context.Context, namespace *Namespace) error
	GetNamespace(ctx context.Context, namespaceID string) (*Namespace, error)
	GetNamespaceByName(ctx context.Context, projectID, name string) (*Namespace, error)
	ListNamespaces(ctx context.Context, projectID string) ([]*Namespace, error)
	UpdateNamespace(ctx context.Context, namespace *Namespace) error
	DeleteNamespace(ctx context.Context, namespaceID string) error
	
	// Member operations
	AddMember(ctx context.Context, member *ProjectMember) error
	GetMember(ctx context.Context, projectID, userID string) (*ProjectMember, error)
	GetMemberByID(ctx context.Context, memberID string) (*ProjectMember, error)
	ListMembers(ctx context.Context, projectID string) ([]*ProjectMember, error)
	UpdateMember(ctx context.Context, member *ProjectMember) error
	RemoveMember(ctx context.Context, memberID string) error
	CountMembers(ctx context.Context, projectID string) (int, error)
	
	// Activity operations
	CreateActivity(ctx context.Context, activity *ProjectActivity) error
	ListActivities(ctx context.Context, filter ActivityFilter) ([]*ProjectActivity, error)
	GetLastActivity(ctx context.Context, projectID string) (*ProjectActivity, error)
	CleanupOldActivities(ctx context.Context, before time.Time) error
	
	// Hierarchy operations
	GetChildProjects(ctx context.Context, parentID string) ([]*Project, error)
	
	// User operations (for member details)
	GetUser(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	
	// Resource usage operations
	GetProjectResourceUsage(ctx context.Context, projectID string) (*ResourceUsage, error)
	GetNamespaceResourceUsage(ctx context.Context, namespaceID string) (*NamespaceUsage, error)
}

// KubernetesRepository defines the interface for Kubernetes operations
type KubernetesRepository interface {
	// Namespace operations
	CreateNamespace(ctx context.Context, workspaceID, name string, labels map[string]string) error
	DeleteNamespace(ctx context.Context, workspaceID, name string) error
	GetNamespace(ctx context.Context, workspaceID, name string) (map[string]interface{}, error)
	ListNamespaces(ctx context.Context, workspaceID string) ([]string, error)
	
	// Resource quota operations
	CreateResourceQuota(ctx context.Context, workspaceID, namespace string, quota *ResourceQuota) error
	UpdateResourceQuota(ctx context.Context, workspaceID, namespace string, quota *ResourceQuota) error
	GetResourceQuota(ctx context.Context, workspaceID, namespace string) (*ResourceQuota, error)
	DeleteResourceQuota(ctx context.Context, workspaceID, namespace string) error
	ApplyResourceQuota(ctx context.Context, workspaceID, namespace string, quota *ResourceQuota) error
	
	// Resource usage operations
	GetNamespaceUsage(ctx context.Context, workspaceID, namespace string) (*NamespaceUsage, error)
	GetNamespaceResourceUsage(ctx context.Context, workspaceID, namespace string) (*ResourceUsage, error)
	
	// RBAC operations
	ApplyRBAC(ctx context.Context, workspaceID, namespace, userID, role string) error
	RemoveRBAC(ctx context.Context, workspaceID, namespace, userID string) error
	
	// HNC operations
	ConfigureHNC(ctx context.Context, workspaceID, parentNamespace, childNamespace string) error
}

// User represents a user in the system
type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}