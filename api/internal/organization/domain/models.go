package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// Organization represents a company or team
type Organization struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	Name             string                 `json:"name"`
	DisplayName      string                 `json:"display_name"`
	Description      string                 `json:"description,omitempty"`
	Website          string                 `json:"website,omitempty"`
	Email            string                 `json:"email,omitempty"`
	Status           string                 `json:"status"` // active, suspended, deleted
	OwnerID          string                 `json:"owner_id,omitempty"`
	Settings         map[string]interface{} `json:"settings,omitempty" gorm:"-"`
	MemberCount      int                    `json:"member_count,omitempty" gorm:"-"`
	SubscriptionInfo *SubscriptionInfo      `json:"subscription_info,omitempty" gorm:"-"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	DeletedAt        *time.Time             `json:"deleted_at,omitempty"`
}

// TableName specifies the table name for GORM
func (Organization) TableName() string {
	return "organizations"
}

// OrganizationUser represents a user's membership in an organization
type OrganizationUser struct {
	OrganizationID string    `json:"organization_id" gorm:"primaryKey"`
	UserID         string    `json:"user_id" gorm:"primaryKey"`
	Email          string    `json:"email,omitempty"`
	Role           string    `json:"role"` // owner, admin, member
	InvitedBy      string    `json:"invited_by,omitempty"`
	InvitedAt      time.Time `json:"invited_at"`
	JoinedAt       time.Time `json:"joined_at"`
	Status         string    `json:"status"` // pending, active, suspended
}

// TableName specifies the table name for GORM
func (OrganizationUser) TableName() string {
	return "organization_users"
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
	DisplayName string                 `json:"display_name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Website     string                 `json:"website,omitempty"`
	Email       string                 `json:"email,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
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
	OwnerID   string
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

// Activity represents an activity log entry for an organization
type Activity struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Type           string    `json:"type"`
	Action         string    `json:"action"`
	ResourceType   string    `json:"resource_type"`
	ResourceID     string    `json:"resource_id"`
	Details        string    `json:"details,omitempty" gorm:"type:jsonb"`
	Timestamp      time.Time `json:"timestamp"`
}

// SetDetails serializes the provided data to JSON and sets it as the Details field
func (a *Activity) SetDetails(data interface{}) error {
	if data == nil {
		a.Details = ""
		return nil
	}
	
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal activity details: %w", err)
	}
	
	a.Details = string(jsonBytes)
	return nil
}

// GetDetails deserializes the Details field into the provided destination
func (a *Activity) GetDetails(dest interface{}) error {
	if a.Details == "" {
		return nil
	}
	
	if err := json.Unmarshal([]byte(a.Details), dest); err != nil {
		return fmt.Errorf("failed to unmarshal activity details: %w", err)
	}
	
	return nil
}

// GetDetailsAsMap returns the Details field as a map[string]interface{}
func (a *Activity) GetDetailsAsMap() (map[string]interface{}, error) {
	if a.Details == "" {
		return make(map[string]interface{}), nil
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(a.Details), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal activity details as map: %w", err)
	}
	
	return result, nil
}

// SetDetailsFromMap sets the Details field from a map[string]interface{}
func (a *Activity) SetDetailsFromMap(data map[string]interface{}) error {
	return a.SetDetails(data)
}

// ActivityFilter represents filter options for listing activities
type ActivityFilter struct {
	OrganizationID string
	UserID         string
	Type           string
	StartDate      *time.Time
	EndDate        *time.Time
	Limit          int
}

// WorkspaceInfo represents basic workspace information
type WorkspaceInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// AddMemberRequest represents a request to add a member
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin member"`
}

// CreateInvitationRequest represents a request to create an invitation
type CreateInvitationRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Role      string `json:"role" binding:"required,oneof=admin member"`
	InvitedBy string `json:"invited_by,omitempty"`
}

// SubscriptionInfo represents subscription information
type SubscriptionInfo struct {
	PlanID    string     `json:"plan_id"`
	PlanName  string     `json:"plan_name"`
	Status    string     `json:"status"`
	ExpiresAt time.Time  `json:"expires_at"`
}