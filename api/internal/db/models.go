package db

import (
	"encoding/json"
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
	VClusterStatus          string    `gorm:"not null;default:'PENDING_CREATION';check:v_cluster_status IN ('PENDING_CREATION','CONFIGURING_HNC','RUNNING','UPDATING_PLAN','UPDATING_NODES','DELETING','ERROR','UNKNOWN','STOPPED','STARTING','STOPPING')" json:"vcluster_status"`
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
	ID                   string    `gorm:"primaryKey" json:"id"`
	WorkspaceID          string    `gorm:"not null;index" json:"workspace_id"`
	Name                 string    `gorm:"not null" json:"name"`
	Description          string    `json:"description"`
	ParentProjectID      *string   `json:"parent_project_id,omitempty"` // Self-referencing for HNC
	HNCanchorName        *string   `json:"hnc_anchor_name,omitempty"`
	NamespaceStatus      string    `gorm:"not null;default:'PENDING_CREATION';check:namespace_status IN ('PENDING_CREATION','ACTIVE','DELETING','ERROR')" json:"namespace_status"`
	KubernetesNamespace  *string   `json:"kubernetes_namespace,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	
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
	ID               string    `gorm:"primaryKey" json:"id"`
	WorkspaceID      *string   `json:"workspace_id,omitempty"` // Null for ClusterRoles
	ProjectID        *string   `json:"project_id,omitempty"`   // Null for workspace-scoped roles
	Name             string    `gorm:"not null" json:"name"`
	Description      string    `json:"description"`
	Rules            string    `gorm:"type:jsonb;not null" json:"rules"` // JSON array of Kubernetes RBAC rules
	Scope            string    `gorm:"not null;default:'namespace';check:scope IN ('namespace','cluster')" json:"scope"`
	K8sRoleName      *string   `json:"k8s_role_name,omitempty"`      // Kubernetes Role/ClusterRole name
	IsCustom         bool      `gorm:"default:true" json:"is_custom"`
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	
	// Associations
	Workspace    *Workspace     `json:"workspace,omitempty"`
	Project      *Project       `json:"project,omitempty"`
	RoleBindings []RoleBinding  `json:"role_bindings,omitempty"`
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

// RoleBinding represents the binding of roles to users or groups
type RoleBinding struct {
	ID                 string    `gorm:"primaryKey" json:"id"`
	WorkspaceID        string    `gorm:"not null;index" json:"workspace_id"`
	ProjectID          *string   `gorm:"index" json:"project_id,omitempty"` // Null for workspace-level bindings
	RoleID             string    `gorm:"not null;index" json:"role_id"`
	SubjectType        string    `gorm:"not null;check:subject_type IN ('User','Group')" json:"subject_type"`
	SubjectID          string    `gorm:"not null;index" json:"subject_id"`
	SubjectName        string    `gorm:"not null" json:"subject_name"` // User email or group name
	K8sRoleBindingName *string   `json:"k8s_rolebinding_name,omitempty"` // Kubernetes RoleBinding name
	IsActive           bool      `gorm:"default:true" json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	
	// Associations
	Workspace *Workspace `json:"workspace,omitempty"`
	Project   *Project   `json:"project,omitempty"`
	Role      Role       `json:"role,omitempty"`
	User      *User      `gorm:"foreignKey:SubjectID" json:"user,omitempty"`
	Group     *Group     `gorm:"foreignKey:SubjectID" json:"group,omitempty"`
}

func (rb *RoleBinding) BeforeCreate(tx *gorm.DB) error {
	if rb.ID == "" {
		rb.ID = "rb-" + uuid.New().String()
	}
	return nil
}

// VClusterProvisioningTask represents async tasks for vCluster operations
type VClusterProvisioningTask struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	WorkspaceID  string    `gorm:"not null;index" json:"workspace_id"`
	TaskType     string    `gorm:"not null;check:task_type IN ('CREATE','DELETE','UPDATE_PLAN','UPDATE_DEDICATED_NODES','SETUP_HNC','START','STOP','UPGRADE','BACKUP','RESTORE')" json:"task_type"`
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
	EventID     string     `gorm:"primaryKey" json:"event_id"` // Stripe Event ID
	EventType   string     `gorm:"not null;index" json:"event_type"`
	Data        string     `gorm:"type:jsonb;not null" json:"data"` // Full Stripe event object
	Status      string     `gorm:"not null;default:'PENDING';check:status IN ('PENDING','PROCESSED','FAILED')" json:"status"`
	ReceivedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"received_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// Subscription represents organization subscriptions
type Subscription struct {
	ID                   string    `gorm:"primaryKey" json:"id"`
	OrganizationID       string    `gorm:"not null;uniqueIndex" json:"organization_id"`
	PlanID               string    `gorm:"not null" json:"plan_id"`
	StripeSubscriptionID string    `gorm:"uniqueIndex" json:"stripe_subscription_id"`
	Status               string    `gorm:"not null;check:status IN ('active','canceled','past_due','trialing','incomplete','incomplete_expired','unpaid')" json:"status"`
	CurrentPeriodStart   time.Time `json:"current_period_start"`
	CurrentPeriodEnd     time.Time `json:"current_period_end"`
	CancelAtPeriodEnd    bool      `gorm:"default:false" json:"cancel_at_period_end"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty"`
	TrialEnd             *time.Time `json:"trial_end,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	
	// Associations
	Organization Organization `json:"organization,omitempty"`
	Plan         Plan         `json:"plan,omitempty"`
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = "sub-" + uuid.New().String()
	}
	return nil
}

// PaymentMethod represents stored payment methods
type PaymentMethod struct {
	ID                    string    `gorm:"primaryKey" json:"id"`
	OrganizationID        string    `gorm:"not null;index" json:"organization_id"`
	StripePaymentMethodID string    `gorm:"uniqueIndex" json:"stripe_payment_method_id"`
	Type                  string    `gorm:"not null" json:"type"` // card, bank_account, etc.
	CardJSON              string    `gorm:"column:card;type:text" json:"-"`
	Card                  *CardDetails `gorm:"-" json:"card,omitempty"`
	IsDefault             bool      `gorm:"default:false" json:"is_default"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	
	// Associations
	Organization Organization `json:"organization,omitempty"`
}

// CardDetails represents credit card information
type CardDetails struct {
	Brand    string `json:"brand"`    // visa, mastercard, etc.
	Last4    string `json:"last4"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
}

func (pm *PaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == "" {
		pm.ID = "pm-" + uuid.New().String()
	}
	return pm.serializeCard()
}

func (pm *PaymentMethod) BeforeSave(tx *gorm.DB) error {
	return pm.serializeCard()
}

func (pm *PaymentMethod) AfterFind(tx *gorm.DB) error {
	return pm.deserializeCard()
}

// serializeCard converts Card struct to JSON string
func (pm *PaymentMethod) serializeCard() error {
	if pm.Card != nil {
		data, err := json.Marshal(pm.Card)
		if err != nil {
			return err
		}
		pm.CardJSON = string(data)
	}
	return nil
}

// deserializeCard converts JSON string to Card struct
func (pm *PaymentMethod) deserializeCard() error {
	if pm.CardJSON != "" {
		var card CardDetails
		if err := json.Unmarshal([]byte(pm.CardJSON), &card); err != nil {
			return err
		}
		pm.Card = &card
	}
	return nil
}

// Invoice represents billing invoices
type Invoice struct {
	ID                 string    `gorm:"primaryKey" json:"id"`
	OrganizationID     string    `gorm:"not null;index" json:"organization_id"`
	SubscriptionID     string    `gorm:"index" json:"subscription_id"`
	StripeInvoiceID    string    `gorm:"uniqueIndex" json:"stripe_invoice_id"`
	InvoiceNumber      string    `gorm:"index" json:"invoice_number"`
	Status             string    `gorm:"not null;check:status IN ('draft','open','paid','void','uncollectible')" json:"status"`
	AmountDue          int64     `json:"amount_due"`        // In cents
	AmountPaid         int64     `json:"amount_paid"`       // In cents
	Currency           string    `gorm:"not null" json:"currency"`
	BillingReason      string    `json:"billing_reason"`    // subscription_cycle, manual, etc.
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	PaidAt             *time.Time `json:"paid_at,omitempty"`
	HostedInvoiceURL   string    `json:"hosted_invoice_url"`
	InvoicePDFURL      string    `json:"invoice_pdf_url"`
	CreatedAt          time.Time `json:"created_at"`
	
	// Associations
	Organization Organization  `json:"organization,omitempty"`
	Subscription *Subscription `json:"subscription,omitempty"`
}

func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = "inv-" + uuid.New().String()
	}
	return nil
}

// UsageRecord represents resource usage for metering
type UsageRecord struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	OrganizationID string    `gorm:"not null;index" json:"organization_id"`
	WorkspaceID    string    `gorm:"not null;index" json:"workspace_id"`
	MetricType     string    `gorm:"not null;check:metric_type IN ('cpu_hours','memory_gb_hours','storage_gb_days','network_gb','api_calls')" json:"metric_type"`
	Quantity       float64   `gorm:"not null" json:"quantity"`
	Unit           string    `gorm:"not null" json:"unit"`
	Timestamp      time.Time `gorm:"not null;index" json:"timestamp"`
	BillingPeriod  string    `gorm:"index" json:"billing_period"` // YYYY-MM format
	Processed      bool      `gorm:"default:false" json:"processed"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	
	// Associations
	Organization Organization `json:"organization,omitempty"`
	Workspace    Workspace    `json:"workspace,omitempty"`
}

func (ur *UsageRecord) BeforeCreate(tx *gorm.DB) error {
	if ur.ID == "" {
		ur.ID = "usage-" + uuid.New().String()
	}
	if ur.BillingPeriod == "" {
		ur.BillingPeriod = ur.Timestamp.Format("2006-01")
	}
	return nil
}

// MetricDefinition represents a custom metric definition
type MetricDefinition struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex" json:"name"`
	Type        string    `gorm:"not null;check:type IN ('counter','gauge','histogram','summary')" json:"type"`
	Description string    `json:"description"`
	Unit        string    `json:"unit"`
	Labels      string    `gorm:"type:jsonb" json:"labels"` // JSON array of label names
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (md *MetricDefinition) BeforeCreate(tx *gorm.DB) error {
	if md.ID == "" {
		md.ID = "metric-" + uuid.New().String()
	}
	return nil
}

// MetricValue represents a recorded metric value
type MetricValue struct {
	ID              string    `gorm:"primaryKey" json:"id"`
	MetricID        string    `gorm:"not null;index" json:"metric_id"`
	WorkspaceID     string    `gorm:"not null;index" json:"workspace_id"`
	OrganizationID  string    `gorm:"not null;index" json:"organization_id"`
	Value           float64   `gorm:"not null" json:"value"`
	Labels          string    `gorm:"type:jsonb" json:"labels"` // JSON object with label values
	Timestamp       time.Time `gorm:"not null;index" json:"timestamp"`
	ScrapedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"scraped_at"`
	Source          string    `gorm:"index" json:"source"` // e.g., "prometheus", "manual", "api"
	CreatedAt       time.Time `json:"created_at"`

	// Associations
	Metric       MetricDefinition `json:"metric,omitempty"`
	Workspace    Workspace        `json:"workspace,omitempty"`
	Organization Organization     `json:"organization,omitempty"`
}

func (mv *MetricValue) BeforeCreate(tx *gorm.DB) error {
	if mv.ID == "" {
		mv.ID = "metric-val-" + uuid.New().String()
	}
	return nil
}

// AlertRule represents an alerting rule configuration
type AlertRule struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	OrganizationID string    `gorm:"not null;index" json:"organization_id"`
	WorkspaceID    *string   `gorm:"index" json:"workspace_id,omitempty"` // null for org-level alerts
	Name           string    `gorm:"not null" json:"name"`
	Description    string    `json:"description"`
	MetricQuery    string    `gorm:"not null" json:"metric_query"` // PromQL query
	Condition      string    `gorm:"not null;check:condition IN ('>', '<', '>=', '<=', '==', '!=')" json:"condition"`
	Threshold      float64   `gorm:"not null" json:"threshold"`
	Duration       string    `gorm:"not null" json:"duration"` // e.g., "5m", "1h"
	Severity       string    `gorm:"not null;check:severity IN ('critical','warning','info')" json:"severity"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	Annotations    string    `gorm:"type:jsonb" json:"annotations"` // JSON object with additional info
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Associations
	Organization Organization `json:"organization,omitempty"`
	Workspace    *Workspace   `json:"workspace,omitempty"`
	Alerts       []Alert      `json:"alerts,omitempty"`
}

func (ar *AlertRule) BeforeCreate(tx *gorm.DB) error {
	if ar.ID == "" {
		ar.ID = "alert-rule-" + uuid.New().String()
	}
	return nil
}

// Alert represents an active or resolved alert
type Alert struct {
	ID             string     `gorm:"primaryKey" json:"id"`
	AlertRuleID    string     `gorm:"not null;index" json:"alert_rule_id"`
	OrganizationID string     `gorm:"not null;index" json:"organization_id"`
	WorkspaceID    *string    `gorm:"index" json:"workspace_id,omitempty"`
	Status         string     `gorm:"not null;check:status IN ('firing','resolved')" json:"status"`
	Value          float64    `json:"value"`           // The value that triggered the alert
	FiredAt        time.Time  `gorm:"not null" json:"fired_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	Labels         string     `gorm:"type:jsonb" json:"labels"`      // JSON object with label values
	Annotations    string     `gorm:"type:jsonb" json:"annotations"` // JSON object with alert details
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Associations
	AlertRule    AlertRule    `json:"alert_rule,omitempty"`
	Organization Organization `json:"organization,omitempty"`
	Workspace    *Workspace   `json:"workspace,omitempty"`
}

func (a *Alert) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = "alert-" + uuid.New().String()
	}
	return nil
}

// MonitoringTarget represents a monitoring target (vCluster, pod, etc.)
type MonitoringTarget struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	OrganizationID string    `gorm:"not null;index" json:"organization_id"`
	WorkspaceID    string    `gorm:"not null;index" json:"workspace_id"`
	Name           string    `gorm:"not null" json:"name"`
	Type           string    `gorm:"not null;check:type IN ('vcluster','pod','service','node')" json:"type"`
	Endpoint       string    `gorm:"not null" json:"endpoint"` // Metrics endpoint URL
	Labels         string    `gorm:"type:jsonb" json:"labels"` // JSON object with target labels
	ScrapeConfig   string    `gorm:"type:jsonb" json:"scrape_config"` // JSON object with scrape configuration
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	LastScrapedAt  *time.Time `json:"last_scraped_at,omitempty"`
	ScrapeDuration *float64   `json:"scrape_duration,omitempty"` // Last scrape duration in seconds
	ScrapeError    *string    `json:"scrape_error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Associations
	Organization Organization `json:"organization,omitempty"`
	Workspace    Workspace    `json:"workspace,omitempty"`
}

func (mt *MonitoringTarget) BeforeCreate(tx *gorm.DB) error {
	if mt.ID == "" {
		mt.ID = "target-" + uuid.New().String()
	}
	return nil
}