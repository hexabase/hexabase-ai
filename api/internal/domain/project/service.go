package project

import (
	"context"
	"time"
)

// Service defines the project business logic interface
type Service interface {
	// Project management
	CreateProject(ctx context.Context, req *CreateProjectRequest) (*Project, error)
	GetProject(ctx context.Context, projectID string) (*Project, error)
	ListProjects(ctx context.Context, filter ProjectFilter) (*ProjectList, error)
	UpdateProject(ctx context.Context, projectID string, req *UpdateProjectRequest) (*Project, error)
	DeleteProject(ctx context.Context, projectID string) error
	GetProjectStats(ctx context.Context, projectID string) (*ProjectStats, error)
	
	// Hierarchical project management
	CreateSubProject(ctx context.Context, parentID string, req *CreateProjectRequest) (*Project, error)
	GetProjectHierarchy(ctx context.Context, projectID string) (*ProjectHierarchy, error)
	
	// Resource management
	ApplyResourceQuota(ctx context.Context, projectID string, quota *ResourceQuota) error
	GetResourceUsage(ctx context.Context, projectID string) (*ResourceUsage, error)
	
	// Namespace management
	CreateNamespace(ctx context.Context, projectID string, req *CreateNamespaceRequest) (*Namespace, error)
	GetNamespace(ctx context.Context, projectID, namespaceID string) (*Namespace, error)
	ListNamespaces(ctx context.Context, projectID string) (*NamespaceList, error)
	UpdateNamespace(ctx context.Context, projectID, namespaceID string, req *CreateNamespaceRequest) (*Namespace, error)
	DeleteNamespace(ctx context.Context, projectID, namespaceID string) error
	GetNamespaceUsage(ctx context.Context, projectID, namespaceID string) (*NamespaceUsage, error)
	
	// Member management
	AddMember(ctx context.Context, projectID, adderID string, req *AddMemberRequest) (*ProjectMember, error)
	GetMember(ctx context.Context, projectID, memberID string) (*ProjectMember, error)
	ListMembers(ctx context.Context, projectID string) (*MemberList, error)
	UpdateMemberRole(ctx context.Context, projectID, memberID string, req *UpdateMemberRoleRequest) (*ProjectMember, error)
	RemoveMember(ctx context.Context, projectID, memberID, removerID string) error
	
	// Project member management (aliases for handler compatibility)
	AddProjectMember(ctx context.Context, projectID string, req *AddMemberRequest) error
	RemoveProjectMember(ctx context.Context, projectID, userID string) error
	ListProjectMembers(ctx context.Context, projectID string) ([]*ProjectMember, error)
	
	// Activity tracking
	ListActivities(ctx context.Context, projectID string, limit int) (*ActivityList, error)
	LogActivity(ctx context.Context, activity *ProjectActivity) error
	GetActivityLogs(ctx context.Context, projectID string, filter ActivityFilter) ([]*ProjectActivity, error)
	
	// Access control
	ValidateProjectAccess(ctx context.Context, userID, projectID string, requiredRole string) error
	GetUserProjectRole(ctx context.Context, userID, projectID string) (string, error)
}

// ProjectStats represents project statistics
type ProjectStats struct {
	ProjectID      string         `json:"project_id"`
	NamespaceCount int            `json:"namespace_count"`
	MemberCount    int            `json:"member_count"`
	ResourceUsage  *ResourceUsage `json:"resource_usage"`
	LastActivity   *time.Time     `json:"last_activity,omitempty"`
}