package organization

import (
	"context"
	"time"
)

// Repository defines the data access interface for organizations
type Repository interface {
	// Organization operations
	CreateOrganization(ctx context.Context, org *Organization) error
	GetOrganization(ctx context.Context, orgID string) (*Organization, error)
	GetOrganizationByName(ctx context.Context, name string) (*Organization, error)
	ListOrganizations(ctx context.Context, filter OrganizationFilter) ([]*Organization, int, error)
	UpdateOrganization(ctx context.Context, org *Organization) error
	DeleteOrganization(ctx context.Context, orgID string) error
	
	// Member operations
	AddMember(ctx context.Context, member *OrganizationUser) error
	GetMember(ctx context.Context, orgID, userID string) (*OrganizationUser, error)
	ListMembers(ctx context.Context, filter MemberFilter) ([]*OrganizationUser, int, error)
	UpdateMember(ctx context.Context, member *OrganizationUser) error
	RemoveMember(ctx context.Context, orgID, userID string) error
	CountMembers(ctx context.Context, orgID string) (int, error)
	
	// User operations (for member details)
	GetUser(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUsersByIDs(ctx context.Context, userIDs []string) ([]*User, error)
	
	// Invitation operations
	CreateInvitation(ctx context.Context, invitation *Invitation) error
	GetInvitation(ctx context.Context, invitationID string) (*Invitation, error)
	GetInvitationByToken(ctx context.Context, token string) (*Invitation, error)
	ListInvitations(ctx context.Context, orgID string, status string) ([]*Invitation, error)
	UpdateInvitation(ctx context.Context, invitation *Invitation) error
	DeleteInvitation(ctx context.Context, invitationID string) error
	DeleteExpiredInvitations(ctx context.Context, before time.Time) error
	
	// Statistics operations
	GetOrganizationStats(ctx context.Context, orgID string) (*OrganizationStats, error)
	GetWorkspaceCount(ctx context.Context, orgID string) (total, active int, error)
	GetProjectCount(ctx context.Context, orgID string) (int, error)
	GetResourceUsage(ctx context.Context, orgID string) (*Usage, error)
}

// User represents a user in the system
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Provider    string    `json:"provider"`
	ExternalID  string    `json:"external_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}