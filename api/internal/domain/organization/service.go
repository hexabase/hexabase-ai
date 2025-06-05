package organization

import (
	"context"
)

// Service defines the organization business logic interface
type Service interface {
	// Organization management
	CreateOrganization(ctx context.Context, userID string, req *CreateOrganizationRequest) (*Organization, error)
	GetOrganization(ctx context.Context, orgID string) (*Organization, error)
	ListOrganizations(ctx context.Context, filter OrganizationFilter) (*OrganizationList, error)
	UpdateOrganization(ctx context.Context, orgID string, req *UpdateOrganizationRequest) (*Organization, error)
	DeleteOrganization(ctx context.Context, orgID string) error
	
	// Member management
	InviteUser(ctx context.Context, orgID, inviterID string, req *InviteUserRequest) (*Invitation, error)
	AcceptInvitation(ctx context.Context, token, userID string) (*OrganizationUser, error)
	ListMembers(ctx context.Context, filter MemberFilter) (*OrganizationMemberList, error)
	GetMember(ctx context.Context, orgID, userID string) (*Member, error)
	UpdateMemberRole(ctx context.Context, orgID, userID string, req *UpdateMemberRoleRequest) (*Member, error)
	RemoveMember(ctx context.Context, orgID, userID, removerID string) error
	
	// Statistics
	GetOrganizationStats(ctx context.Context, orgID string) (*OrganizationStats, error)
	
	// Access control
	ValidateOrganizationAccess(ctx context.Context, userID, orgID string, requiredRole string) error
	GetUserRole(ctx context.Context, userID, orgID string) (string, error)
	
	// Invitation management
	GetInvitation(ctx context.Context, invitationID string) (*Invitation, error)
	ListPendingInvitations(ctx context.Context, orgID string) ([]*Invitation, error)
	ResendInvitation(ctx context.Context, invitationID string) error
	CancelInvitation(ctx context.Context, invitationID string) error
	
	// Cleanup
	CleanupExpiredInvitations(ctx context.Context) error
}