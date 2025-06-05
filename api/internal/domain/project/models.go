package project

import (
	"time"
)

// Project represents a project within a workspace
type Project struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	WorkspaceID   string               `json:"workspace_id"`
	WorkspaceName string               `json:"workspace_name,omitempty"`
	Status        string               `json:"status"` // active, inactive, archived
	NamespaceName string               `json:"namespace_name,omitempty"`
	ResourceQuotas *ResourceQuotas     `json:"resource_quotas,omitempty"`
	ResourceUsage  *ResourceUsage      `json:"resource_usage,omitempty"`
	Labels         map[string]string   `json:"labels,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	DeletedAt      *time.Time          `json:"deleted_at,omitempty"`
}

// ResourceQuotas represents resource quotas for a project
type ResourceQuotas struct {
	CPULimit     string `json:"cpu_limit"`
	MemoryLimit  string `json:"memory_limit"`
	StorageLimit string `json:"storage_limit"`
	PodLimit     string `json:"pod_limit,omitempty"`
}

// ResourceUsage represents current resource usage
type ResourceUsage struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Pods   int    `json:"pods"`
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name           string            `json:"name" binding:"required"`
	Description    string            `json:"description,omitempty"`
	WorkspaceID    string            `json:"workspace_id" binding:"required"`
	NamespaceName  string            `json:"namespace_name,omitempty"`
	ResourceQuotas *ResourceQuotas   `json:"resource_quotas,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Name           string            `json:"name,omitempty"`
	Description    string            `json:"description,omitempty"`
	ResourceQuotas *ResourceQuotas   `json:"resource_quotas,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Status         string            `json:"status,omitempty"`
}

// Namespace represents a Kubernetes namespace
type Namespace struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	ProjectID     string           `json:"project_id"`
	Description   string           `json:"description,omitempty"`
	Status        string           `json:"status"` // active, terminating
	ResourceQuota *ResourceQuota   `json:"resource_quota,omitempty"`
	ResourceUsage *NamespaceUsage  `json:"resource_usage,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// ResourceQuota represents Kubernetes resource quota
type ResourceQuota struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
	Pods    int    `json:"pods"`
}

// NamespaceUsage represents namespace resource usage
type NamespaceUsage struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
	Pods    int    `json:"pods"`
}

// CreateNamespaceRequest represents a request to create a namespace
type CreateNamespaceRequest struct {
	Name          string            `json:"name" binding:"required"`
	Description   string            `json:"description,omitempty"`
	ResourceQuota *ResourceQuota    `json:"resource_quota,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// ProjectMember represents a project member
type ProjectMember struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	UserID      string    `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	UserName    string    `json:"user_name"`
	Role        string    `json:"role"` // admin, developer, viewer
	AddedBy     string    `json:"added_by"`
	AddedAt     time.Time `json:"added_at"`
}

// AddProjectMemberRequest represents a request to add a project member
type AddProjectMemberRequest struct {
	UserEmail string `json:"user_email" binding:"required,email"`
	Role      string `json:"role" binding:"required,oneof=admin developer viewer"`
}

// UpdateMemberRoleRequest represents a request to update member role
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin developer viewer"`
}

// ProjectActivity represents a project activity log entry
type ProjectActivity struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	UserID      string                 `json:"user_id"`
	UserEmail   string                 `json:"user_email"`
	UserName    string                 `json:"user_name"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ProjectList represents a list of projects
type ProjectList struct {
	Projects []*Project `json:"projects"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// ProjectFilter represents filter options for listing projects
type ProjectFilter struct {
	WorkspaceID string
	Status      string
	Search      string
	Page        int
	PageSize    int
	SortBy      string
	SortOrder   string
}

// NamespaceList represents a list of namespaces
type NamespaceList struct {
	Namespaces []*Namespace `json:"namespaces"`
	Total      int          `json:"total"`
}

// MemberList represents a list of project members
type MemberList struct {
	Members []*ProjectMember `json:"members"`
	Total   int              `json:"total"`
}

// ActivityList represents a list of project activities
type ActivityList struct {
	Activities []*ProjectActivity `json:"activities"`
	Total      int                `json:"total"`
}