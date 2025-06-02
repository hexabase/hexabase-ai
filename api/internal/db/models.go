package db

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user account linked to external IdP
type User struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	ExternalID  string    `gorm:"not null;index" json:"external_id"`
	Provider    string    `gorm:"not null;index" json:"provider"`
	Email       string    `gorm:"unique;not null" json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate sets ID if not provided
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = "hxb-usr-" + uuid.New().String()
	}
	return nil
}

// Organization represents a billing and management unit
type Organization struct {
	ID                   string    `gorm:"primaryKey" json:"id"`
	Name                 string    `gorm:"not null" json:"name"`
	StripeCustomerID     *string   `gorm:"unique" json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string   `gorm:"unique" json:"stripe_subscription_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	
	// Associations
	Users      []OrganizationUser `json:"users,omitempty"`
	Workspaces []Workspace        `json:"workspaces,omitempty"`
}

func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = "org-" + uuid.New().String()
	}
	return nil
}

// OrganizationUser represents the many-to-many relationship between users and organizations
type OrganizationUser struct {
	OrganizationID string    `gorm:"primaryKey" json:"organization_id"`
	UserID         string    `gorm:"primaryKey" json:"user_id"`
	Role           string    `gorm:"not null;default:'member';check:role IN ('admin','member')" json:"role"`
	JoinedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	
	// Associations
	Organization Organization `json:"organization,omitempty"`
	User         User         `json:"user,omitempty"`
}

// Plan represents subscription plans with resource limits
type Plan struct {
	ID                        string  `gorm:"primaryKey" json:"id"`
	Name                      string  `gorm:"not null" json:"name"`
	Description               string  `json:"description"`
	Price                     float64 `gorm:"not null" json:"price"`
	Currency                  string  `gorm:"not null;size:3;check:length(currency) = 3" json:"currency"`
	StripePriceID            string  `gorm:"unique;not null" json:"stripe_price_id"`
	ResourceLimits           string  `gorm:"type:jsonb" json:"resource_limits"` // JSON string
	AllowsDedicatedNodes     bool    `gorm:"default:false" json:"allows_dedicated_nodes"`
	DefaultDedicatedNodeConfig string `gorm:"type:jsonb" json:"default_dedicated_node_config"` // JSON string
	MaxProjectsPerWorkspace  *int    `json:"max_projects_per_workspace,omitempty"`
	MaxMembersPerWorkspace   *int    `json:"max_members_per_workspace,omitempty"`
	IsActive                 bool    `gorm:"default:true" json:"is_active"`
	DisplayOrder             int     `gorm:"default:0" json:"display_order"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// Workspace represents a vCluster instance
type Workspace struct {
	ID                       string    `gorm:"primaryKey" json:"id"`
	OrganizationID          string    `gorm:"not null;index" json:"organization_id"`
	Name                     string    `gorm:"not null" json:"name"`
	PlanID                   string    `gorm:"not null" json:"plan_id"`
	VClusterInstanceName     *string   `gorm:"unique" json:"vcluster_instance_name,omitempty"`
	VClusterStatus          string    `gorm:"not null;default:'PENDING_CREATION';check:v_cluster_status IN ('PENDING_CREATION','CONFIGURING_HNC','RUNNING','UPDATING_PLAN','UPDATING_NODES','DELETING','ERROR','UNKNOWN')" json:"vcluster_status"`
	VClusterConfig          string    `gorm:"type:jsonb" json:"vcluster_config"`
	DedicatedNodeConfig     string    `gorm:"type:jsonb" json:"dedicated_node_config"`
	StripeSubscriptionItemID *string   `gorm:"unique" json:"stripe_subscription_item_id,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	
	// Associations
	Organization Organization `json:"organization,omitempty"`
	Plan         Plan         `json:"plan,omitempty"`
	Projects     []Project    `json:"projects,omitempty"`
	Groups       []Group      `json:"groups,omitempty"`
}

func (w *Workspace) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = "ws-" + uuid.New().String()
	}
	return nil
}

// Project represents a Namespace within a vCluster
type Project struct {
	ID               string    `gorm:"primaryKey" json:"id"`
	WorkspaceID      string    `gorm:"not null;index" json:"workspace_id"`
	Name             string    `gorm:"not null" json:"name"`
	ParentProjectID  *string   `json:"parent_project_id,omitempty"` // Self-referencing for HNC
	HNCanchorName    *string   `json:"hnc_anchor_name,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	
	// Associations
	Workspace      Workspace  `json:"workspace,omitempty"`
	ParentProject  *Project   `gorm:"foreignKey:ParentProjectID" json:"parent_project,omitempty"`
	ChildProjects  []Project  `gorm:"foreignKey:ParentProjectID" json:"child_projects,omitempty"`
	Roles          []Role     `json:"roles,omitempty"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = "prj-" + uuid.New().String()
	}
	return nil
}

// Group represents hierarchical user groups within workspaces
type Group struct {
	ID              string    `gorm:"primaryKey" json:"id"`
	WorkspaceID     string    `gorm:"not null;index" json:"workspace_id"`
	Name            string    `gorm:"not null" json:"name"`
	ParentGroupID   *string   `json:"parent_group_id,omitempty"` // Self-referencing
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	
	// Associations
	Workspace     Workspace      `json:"workspace,omitempty"`
	ParentGroup   *Group         `gorm:"foreignKey:ParentGroupID" json:"parent_group,omitempty"`
	ChildGroups   []Group        `gorm:"foreignKey:ParentGroupID" json:"child_groups,omitempty"`
	Memberships   []GroupMembership `json:"memberships,omitempty"`
}

func (g *Group) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = "grp-" + uuid.New().String()
	}
	return nil
}

// GroupMembership represents user membership in groups
type GroupMembership struct {
	GroupID   string    `gorm:"primaryKey" json:"group_id"`
	UserID    string    `gorm:"primaryKey" json:"user_id"`
	JoinedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	
	// Associations
	Group Group `json:"group,omitempty"`
	User  User  `json:"user,omitempty"`
}

// Role represents custom or preset roles
type Role struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	ProjectID   *string   `json:"project_id,omitempty"` // Null for ClusterRoles
	Name        string    `gorm:"not null" json:"name"`
	Rules       string    `gorm:"type:jsonb;not null" json:"rules"` // JSON array of RBAC rules
	IsCustom    bool      `gorm:"default:true" json:"is_custom"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Associations
	Project *Project `json:"project,omitempty"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = "role-" + uuid.New().String()
	}
	return nil
}

// RoleAssignment represents role assignments to groups
type RoleAssignment struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	GroupID   string    `gorm:"not null;index" json:"group_id"`
	RoleID    string    `gorm:"not null;index" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	
	// Associations
	Group Group `json:"group,omitempty"`
	Role  Role  `json:"role,omitempty"`
}

func (ra *RoleAssignment) BeforeCreate(tx *gorm.DB) error {
	if ra.ID == "" {
		ra.ID = "ra-" + uuid.New().String()
	}
	return nil
}

// VClusterProvisioningTask represents async tasks for vCluster operations
type VClusterProvisioningTask struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	WorkspaceID  string    `gorm:"not null;index" json:"workspace_id"`
	TaskType     string    `gorm:"not null;check:task_type IN ('CREATE','DELETE','UPDATE_PLAN','UPDATE_DEDICATED_NODES','SETUP_HNC')" json:"task_type"`
	Status       string    `gorm:"not null;default:'PENDING';check:status IN ('PENDING','RUNNING','COMPLETED','FAILED')" json:"status"`
	Payload      string    `gorm:"type:jsonb" json:"payload"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// Associations
	Workspace Workspace `json:"workspace,omitempty"`
}

func (vt *VClusterProvisioningTask) BeforeCreate(tx *gorm.DB) error {
	if vt.ID == "" {
		vt.ID = "task-" + uuid.New().String()
	}
	return nil
}

// StripeEvent represents Stripe webhook events
type StripeEvent struct {
	ID          string    `gorm:"primaryKey" json:"id"` // Stripe Event ID
	EventType   string    `gorm:"not null;index" json:"event_type"`
	Data        string    `gorm:"type:jsonb;not null" json:"data"` // Full Stripe event object
	Status      string    `gorm:"not null;default:'PENDING';check:status IN ('PENDING','PROCESSED','FAILED')" json:"status"`
	ReceivedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"received_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}