package organization

import (
	"time"
)

// Organization represents a company or team
type Organization struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description,omitempty"`
	Website     string    `json:"website,omitempty"`
	Email       string    `json:"email,omitempty"`
	Status      string    `json:"status"` // active, suspended, deleted
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// OrganizationUser represents a user's membership in an organization
type OrganizationUser struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Role           string    `json:"role"` // owner, admin, member
	InvitedBy      string    `json:"invited_by,omitempty"`
	InvitedAt      time.Time `json:"invited_at"`
	JoinedAt       time.Time `json:"joined_at"`
	Status         string    `json:"status"` // pending, active, suspended
}

// CreateOrganizationRequest represents a request to create an organization
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty"`
	Email       string `json:"email,omitempty"`
}

// UpdateOrganizationRequest represents a request to update an organization
type UpdateOrganizationRequest struct {
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty"`
	Email       string `json:"email,omitempty"`
}

// InviteUserRequest represents a request to invite a user to an organization
type InviteUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member"`
}

// UpdateMemberRoleRequest represents a request to update a member's role
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

// OrganizationList represents a list of organizations
type OrganizationList struct {
	Organizations []*Organization `json:"organizations"`
	Total         int             `json:"total"`
	Page          int             `json:"page"`
	PageSize      int             `json:"page_size"`
}

// OrganizationMemberList represents a list of organization members
type OrganizationMemberList struct {
	Members  []*Member `json:"members"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

// Member represents a member with user details
type Member struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	JoinedAt    time.Time `json:"joined_at"`
	LastActive  time.Time `json:"last_active,omitempty"`
}

// OrganizationStats represents organization statistics
type OrganizationStats struct {
	OrganizationID   string    `json:"organization_id"`
	TotalMembers     int       `json:"total_members"`
	ActiveMembers    int       `json:"active_members"`
	TotalWorkspaces  int       `json:"total_workspaces"`
	ActiveWorkspaces int       `json:"active_workspaces"`
	TotalProjects    int       `json:"total_projects"`
	ResourceUsage    *Usage    `json:"resource_usage"`
	LastUpdated      time.Time `json:"last_updated"`
}

// Usage represents resource usage statistics
type Usage struct {
	CPU     float64 `json:"cpu_cores"`
	Memory  float64 `json:"memory_gb"`
	Storage float64 `json:"storage_gb"`
	Cost    float64 `json:"estimated_cost"`
}

// OrganizationFilter represents filter options for listing organizations
type OrganizationFilter struct {
	UserID    string
	Status    string
	Search    string
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// MemberFilter represents filter options for listing members
type MemberFilter struct {
	OrganizationID string
	Role           string
	Status         string
	Search         string
	Page           int
	PageSize       int
}

// Invitation represents a pending invitation
type Invitation struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	Token          string    `json:"token"`
	InvitedBy      string    `json:"invited_by"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
	Status         string    `json:"status"` // pending, accepted, expired
}