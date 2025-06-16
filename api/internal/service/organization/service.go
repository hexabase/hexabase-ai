package organization

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/organization"
)

type service struct {
	repo       organization.Repository
	authRepo   organization.AuthRepository
	billingRepo organization.BillingRepository
	logger     *slog.Logger
}

// NewService creates a new organization service
func NewService(
	repo organization.Repository,
	authRepo organization.AuthRepository,
	billingRepo organization.BillingRepository,
	logger *slog.Logger,
) organization.Service {
	return &service{
		repo:        repo,
		authRepo:    authRepo,
		billingRepo: billingRepo,
		logger:      logger,
	}
}

func (s *service) CreateOrganization(ctx context.Context, userID string, req *organization.CreateOrganizationRequest) (*organization.Organization, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	// Create organization
	org := &organization.Organization{
		ID:          uuid.New().String(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Email:       req.Email,
		Website:     req.Website,
		Status:      "active",
		OwnerID:     userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if org.DisplayName == "" {
		org.DisplayName = org.Name
	}

	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add owner as admin member
	member := &organization.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         userID,
		Role:           "admin",
		JoinedAt:       time.Now(),
		Status:         "active",
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		s.logger.Error("failed to add owner as member", "error", err)
	}

	// Create billing customer
	if _, err := s.billingRepo.CreateCustomer(ctx, org); err != nil {
		s.logger.Error("failed to create billing customer", "error", err)
	}

	return org, nil
}

func (s *service) GetOrganization(ctx context.Context, orgID string) (*organization.Organization, error) {
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get member count
	filter := organization.MemberFilter{
		OrganizationID: orgID,
		PageSize:       1000, // Get all members for count
	}
	_, total, err := s.repo.ListMembers(ctx, filter)
	if err != nil {
		s.logger.Warn("failed to get member count", "error", err)
	} else {
		org.MemberCount = total
	}

	// Get subscription info
	subscription, err := s.billingRepo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		s.logger.Warn("failed to get subscription info", "error", err)
	} else {
		org.SubscriptionInfo = &organization.SubscriptionInfo{
			PlanID:    subscription.PlanID,
			PlanName:  subscription.PlanName,
			Status:    subscription.Status,
			ExpiresAt: subscription.CurrentPeriodEnd,
		}
	}

	return org, nil
}

func (s *service) ListOrganizations(ctx context.Context, filter organization.OrganizationFilter) (*organization.OrganizationList, error) {
	if filter.UserID != "" {
		// Handle development user specially
		if filter.UserID == "dev-user-1" {
			// Check if dev organization exists
			devOrg, err := s.repo.GetOrganization(ctx, "dev-org-1")
			if err != nil {
				// Create development organization if it doesn't exist
				devOrg = &organization.Organization{
					ID:          "dev-org-1",
					Name:        "development",
					DisplayName: "Development Organization",
					Description: "Default organization for development",
					Status:      "active",
					OwnerID:     "dev-user-1",
					Settings:    nil, // Don't set Settings to avoid GORM issues
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := s.repo.CreateOrganization(ctx, devOrg); err != nil {
					s.logger.Warn("failed to create dev organization", "error", err)
				}
				
				// Add dev user as member
				member := &organization.OrganizationUser{
					OrganizationID: devOrg.ID,
					UserID:         "dev-user-1",
					Role:           "admin",
					JoinedAt:       time.Now(),
					InvitedAt:      time.Now(),
					Status:         "active",
				}
				if err := s.repo.AddMember(ctx, member); err != nil {
					s.logger.Warn("failed to add dev user as member", "error", err)
				}
			}
			
			return &organization.OrganizationList{
				Organizations: []*organization.Organization{devOrg},
				Total:         1,
				Page:          filter.Page,
				PageSize:      filter.PageSize,
			}, nil
		}

		// Get organizations where user is a member
		orgIDs, err := s.authRepo.GetUserOrganizations(ctx, filter.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user organizations: %w", err)
		}

		if len(orgIDs) == 0 {
			return &organization.OrganizationList{
				Organizations: []*organization.Organization{},
				Total:         0,
				Page:          filter.Page,
				PageSize:      filter.PageSize,
			}, nil
		}

		organizations := make([]*organization.Organization, 0, len(orgIDs))
		for _, orgID := range orgIDs {
			org, err := s.repo.GetOrganization(ctx, orgID)
			if err != nil {
				s.logger.Warn("failed to get organization", "org_id", orgID, "error", err)
				continue
			}
			organizations = append(organizations, org)
		}

		return &organization.OrganizationList{
			Organizations: organizations,
			Total:         len(organizations),
			Page:          filter.Page,
			PageSize:      filter.PageSize,
		}, nil
	}

	// Use repository to list with filters
	orgs, total, err := s.repo.ListOrganizations(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	return &organization.OrganizationList{
		Organizations: orgs,
		Total:         total,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
	}, nil
}

func (s *service) UpdateOrganization(ctx context.Context, orgID string, req *organization.UpdateOrganizationRequest) (*organization.Organization, error) {
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Update fields
	if req.DisplayName != "" {
		org.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		org.Description = req.Description
	}
	if req.Settings != nil {
		org.Settings = req.Settings
	}

	org.UpdatedAt = time.Now()

	if err := s.repo.UpdateOrganization(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return org, nil
}

func (s *service) DeleteOrganization(ctx context.Context, orgID string) error {
	// Check if organization has active workspaces
	workspaces, err := s.repo.ListWorkspaces(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to check workspaces: %w", err)
	}

	if len(workspaces) > 0 {
		return fmt.Errorf("cannot delete organization with active workspaces")
	}

	// Cancel subscription
	if err := s.billingRepo.CancelSubscription(ctx, orgID); err != nil {
		s.logger.Error("failed to cancel subscription", "error", err)
	}

	// Delete billing customer
	if err := s.billingRepo.DeleteCustomer(ctx, orgID); err != nil {
		s.logger.Error("failed to delete billing customer", "error", err)
	}

	// Delete organization
	if err := s.repo.DeleteOrganization(ctx, orgID); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}


func (s *service) RemoveMember(ctx context.Context, orgID, userID, removerID string) error {
	// Check if user is the owner
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if org.OwnerID == userID {
		return fmt.Errorf("cannot remove organization owner")
	}

	// Remove member
	// Get member details before removal for logging
	member, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	if err := s.repo.RemoveMember(ctx, orgID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Log activity with structured details
	s.logActivityWithDetails(ctx, orgID, removerID, "member", "removed", "organization_user", userID, map[string]interface{}{
		"removed_role": member.Role,
		"removed_by":   removerID,
	})

	return nil
}

func (s *service) UpdateMemberRole(ctx context.Context, orgID, userID, updatedBy string, req *organization.UpdateMemberRoleRequest) (*organization.Member, error) {
	// Validate role
	if req.Role != "admin" && req.Role != "member" {
		return nil, fmt.Errorf("invalid role: %s", req.Role)
	}

	// Check if user is the owner
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if org.OwnerID == userID && req.Role != "admin" {
		return nil, fmt.Errorf("cannot change owner role from admin")
	}

	// Update role
	// Get current member to log the change
	currentMember, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current member: %w", err)
	}

	if err := s.repo.UpdateMemberRole(ctx, orgID, userID, req.Role); err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	// Log activity with structured details
	s.logActivityWithDetails(ctx, orgID, updatedBy, "member", "role_updated", "organization_user", userID, map[string]interface{}{
		"old_role": currentMember.Role,
		"new_role": req.Role,
	})

	// Get updated member
	return s.GetMember(ctx, orgID, userID)
}

func (s *service) ListMembers(ctx context.Context, filter organization.MemberFilter) (*organization.OrganizationMemberList, error) {
	orgUsers, total, err := s.repo.ListMembers(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	// Convert OrganizationUser to Member
	members := make([]*organization.Member, 0, len(orgUsers))
	for _, ou := range orgUsers {
		// Get user details
		user, err := s.authRepo.GetUser(ctx, ou.UserID)
		if err != nil {
			s.logger.Warn("failed to get user details", "user_id", ou.UserID, "error", err)
			continue
		}
		
		member := &organization.Member{
			ID:          fmt.Sprintf("%s-%s", ou.OrganizationID, ou.UserID), // Composite ID
			UserID:      ou.UserID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        ou.Role,
			Status:      ou.Status,
			JoinedAt:    ou.JoinedAt,
		}
		members = append(members, member)
	}
	
	return &organization.OrganizationMemberList{
		Members:  members,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}


func (s *service) AcceptInvitation(ctx context.Context, token, userID string) (*organization.OrganizationUser, error) {
	// Get invitation by token
	invitation, err := s.repo.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != "pending" {
		return nil, fmt.Errorf("invitation is not pending")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invitation has expired")
	}

	// Update invitation status
	invitation.Status = "accepted"
	invitation.AcceptedAt = &[]time.Time{time.Now()}[0]

	if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	// Add member
	member := &organization.OrganizationUser{
		OrganizationID: invitation.OrganizationID,
		UserID:         userID,
		Email:          invitation.Email,
		Role:           invitation.Role,
		JoinedAt:       time.Now(),
		Status:         "active",
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Log activity with structured details
	s.logActivityWithDetails(ctx, invitation.OrganizationID, userID, "member", "joined", "organization_user", userID, map[string]interface{}{
		"role":        invitation.Role,
		"invited_by":  invitation.InvitedBy,
		"via_invitation": true,
	})

	return member, nil
}




func (s *service) LogActivity(ctx context.Context, activity *organization.Activity) error {
	activity.ID = uuid.New().String()
	activity.Timestamp = time.Now()

	if err := s.repo.CreateActivity(ctx, activity); err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}

// logActivityWithDetails is a helper method for logging activities with structured details
func (s *service) logActivityWithDetails(ctx context.Context, orgID, userID, activityType, action, resourceType, resourceID string, details map[string]interface{}) {
	activity := &organization.Activity{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		UserID:         userID,
		Type:           activityType,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		Timestamp:      time.Now(),
	}

	// Use helper method to set details
	if err := activity.SetDetailsFromMap(details); err != nil {
		s.logger.Error("failed to set activity details", "error", err)
		return
	}

	if err := s.repo.CreateActivity(ctx, activity); err != nil {
		s.logger.Error("failed to log activity", "error", err)
	}
}

func (s *service) GetActivityLogs(ctx context.Context, orgID string, filter organization.ActivityFilter) ([]*organization.Activity, error) {
	filter.OrganizationID = orgID
	return s.repo.ListActivities(ctx, filter)
}

// InviteUser sends an invitation to a user to join the organization
func (s *service) InviteUser(ctx context.Context, orgID, inviterID string, req *organization.InviteUserRequest) (*organization.Invitation, error) {
	// Check if organization exists
	if _, err := s.repo.GetOrganization(ctx, orgID); err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Check if user is already a member by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil && user != nil {
		// User exists, check if already a member
		filter := organization.MemberFilter{
			OrganizationID: orgID,
			PageSize:       1000,
		}
		members, _, err := s.repo.ListMembers(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to list members: %w", err)
		}

		for _, member := range members {
			if member.UserID == user.ID {
				return nil, fmt.Errorf("user is already a member")
			}
		}
	}

	// Check for existing invitation
	invitations, err := s.repo.ListInvitations(ctx, orgID, "pending")
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}

	for _, inv := range invitations {
		if inv.Email == req.Email {
			return nil, fmt.Errorf("invitation already exists")
		}
	}

	// Create invitation
	invitation := &organization.Invitation{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Email:          req.Email,
		Role:           req.Role,
		InvitedBy:      inviterID,
		Status:         "pending",
		Token:          uuid.New().String(),
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:      time.Now(),
	}

	if err := s.repo.CreateInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// TODO: Send invitation email

	return invitation, nil
}

// GetMember gets a specific member of an organization
func (s *service) GetMember(ctx context.Context, orgID, userID string) (*organization.Member, error) {
	member, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	// Get user details
	user, err := s.authRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user details: %w", err)
	}

	return &organization.Member{
		ID:          fmt.Sprintf("%s-%s", member.OrganizationID, member.UserID), // Composite ID
		UserID:      member.UserID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        member.Role,
		Status:      member.Status,
		JoinedAt:    member.JoinedAt,
	}, nil
}

// GetOrganizationStats gets statistics for an organization
func (s *service) GetOrganizationStats(ctx context.Context, orgID string) (*organization.OrganizationStats, error) {
	return s.repo.GetOrganizationStats(ctx, orgID)
}

// ValidateOrganizationAccess validates if a user has access to an organization with a specific role
func (s *service) ValidateOrganizationAccess(ctx context.Context, userID, orgID string, requiredRole string) error {
	member, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return fmt.Errorf("access denied: user is not a member")
	}

	if member.Status != "active" {
		return fmt.Errorf("access denied: member status is %s", member.Status)
	}

	// Check role hierarchy
	if requiredRole == "admin" && member.Role != "admin" {
		return fmt.Errorf("access denied: admin role required")
	}

	return nil
}

// GetUserRole gets the role of a user in an organization
func (s *service) GetUserRole(ctx context.Context, userID, orgID string) (string, error) {
	member, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get member: %w", err)
	}
	return member.Role, nil
}

// GetInvitation gets an invitation by ID
func (s *service) GetInvitation(ctx context.Context, invitationID string) (*organization.Invitation, error) {
	return s.repo.GetInvitation(ctx, invitationID)
}

// ListPendingInvitations lists all pending invitations for an organization
func (s *service) ListPendingInvitations(ctx context.Context, orgID string) ([]*organization.Invitation, error) {
	return s.repo.ListInvitations(ctx, orgID, "pending")
}

// ResendInvitation resends an invitation
func (s *service) ResendInvitation(ctx context.Context, invitationID string) error {
	invitation, err := s.repo.GetInvitation(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != "pending" {
		return fmt.Errorf("can only resend pending invitations")
	}

	// Update expiration
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days
	if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// TODO: Send invitation email

	return nil
}

// CancelInvitation cancels a pending invitation
func (s *service) CancelInvitation(ctx context.Context, invitationID string) error {
	invitation, err := s.repo.GetInvitation(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != "pending" {
		return fmt.Errorf("can only cancel pending invitations")
	}

	// Update invitation status
	invitation.Status = "canceled"

	if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	return nil
}

// CleanupExpiredInvitations removes expired invitations
func (s *service) CleanupExpiredInvitations(ctx context.Context) error {
	return s.repo.DeleteExpiredInvitations(ctx, time.Now())
}