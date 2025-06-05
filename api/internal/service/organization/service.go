package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/organization"
	"go.uber.org/zap"
)

type service struct {
	repo       organization.Repository
	authRepo   organization.AuthRepository
	billingRepo organization.BillingRepository
	logger     *zap.Logger
}

// NewService creates a new organization service
func NewService(
	repo organization.Repository,
	authRepo organization.AuthRepository,
	billingRepo organization.BillingRepository,
	logger *zap.Logger,
) organization.Service {
	return &service{
		repo:        repo,
		authRepo:    authRepo,
		billingRepo: billingRepo,
		logger:      logger,
	}
}

func (s *service) CreateOrganization(ctx context.Context, req *organization.CreateOrganizationRequest) (*organization.Organization, error) {
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
		OwnerID:     req.OwnerID,
		Settings:    req.Settings,
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
	member := &organization.Member{
		OrganizationID: org.ID,
		UserID:         req.OwnerID,
		Role:           "admin",
		JoinedAt:       time.Now(),
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		s.logger.Error("failed to add owner as member", zap.Error(err))
	}

	// Create billing customer
	if err := s.billingRepo.CreateCustomer(ctx, org); err != nil {
		s.logger.Error("failed to create billing customer", zap.Error(err))
	}

	return org, nil
}

func (s *service) GetOrganization(ctx context.Context, orgID string) (*organization.Organization, error) {
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get member count
	members, err := s.repo.ListMembers(ctx, orgID)
	if err != nil {
		s.logger.Warn("failed to get member count", zap.Error(err))
	} else {
		org.MemberCount = len(members)
	}

	// Get subscription info
	subscription, err := s.billingRepo.GetOrganizationSubscription(ctx, orgID)
	if err != nil {
		s.logger.Warn("failed to get subscription info", zap.Error(err))
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

func (s *service) ListOrganizations(ctx context.Context, userID string) ([]*organization.Organization, error) {
	// Get organizations where user is a member
	orgIDs, err := s.authRepo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	if len(orgIDs) == 0 {
		return []*organization.Organization{}, nil
	}

	organizations := make([]*organization.Organization, 0, len(orgIDs))
	for _, orgID := range orgIDs {
		org, err := s.repo.GetOrganization(ctx, orgID)
		if err != nil {
			s.logger.Warn("failed to get organization", zap.String("org_id", orgID), zap.Error(err))
			continue
		}
		organizations = append(organizations, org)
	}

	return organizations, nil
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
		s.logger.Error("failed to cancel subscription", zap.Error(err))
	}

	// Delete billing customer
	if err := s.billingRepo.DeleteCustomer(ctx, orgID); err != nil {
		s.logger.Error("failed to delete billing customer", zap.Error(err))
	}

	// Delete organization
	if err := s.repo.DeleteOrganization(ctx, orgID); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

func (s *service) AddMember(ctx context.Context, orgID string, req *organization.AddMemberRequest) error {
	// Check if organization exists
	if _, err := s.repo.GetOrganization(ctx, orgID); err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	// Check if user is already a member
	members, err := s.repo.ListMembers(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to list members: %w", err)
	}

	for _, member := range members {
		if member.UserID == req.UserID {
			return fmt.Errorf("user is already a member")
		}
	}

	// Add member
	member := &organization.Member{
		OrganizationID: orgID,
		UserID:         req.UserID,
		Role:           req.Role,
		JoinedAt:       time.Now(),
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

func (s *service) RemoveMember(ctx context.Context, orgID, userID string) error {
	// Check if user is the owner
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if org.OwnerID == userID {
		return fmt.Errorf("cannot remove organization owner")
	}

	// Remove member
	if err := s.repo.RemoveMember(ctx, orgID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

func (s *service) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	// Validate role
	if role != "admin" && role != "member" {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Check if user is the owner
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if org.OwnerID == userID && role != "admin" {
		return fmt.Errorf("cannot change owner role from admin")
	}

	// Update role
	if err := s.repo.UpdateMemberRole(ctx, orgID, userID, role); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	return nil
}

func (s *service) ListMembers(ctx context.Context, orgID string) ([]*organization.Member, error) {
	return s.repo.ListMembers(ctx, orgID)
}

func (s *service) CreateInvitation(ctx context.Context, orgID string, req *organization.CreateInvitationRequest) (*organization.Invitation, error) {
	// Check if organization exists
	if _, err := s.repo.GetOrganization(ctx, orgID); err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Check if user is already a member
	members, err := s.repo.ListMembers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	for _, member := range members {
		if member.Email == req.Email {
			return nil, fmt.Errorf("user is already a member")
		}
	}

	// Check for existing invitation
	invitations, err := s.repo.ListInvitations(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}

	for _, inv := range invitations {
		if inv.Email == req.Email && inv.Status == "pending" {
			return nil, fmt.Errorf("invitation already exists")
		}
	}

	// Create invitation
	invitation := &organization.Invitation{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Email:          req.Email,
		Role:           req.Role,
		InvitedBy:      req.InvitedBy,
		Status:         "pending",
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:      time.Now(),
	}

	if err := s.repo.CreateInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// TODO: Send invitation email

	return invitation, nil
}

func (s *service) AcceptInvitation(ctx context.Context, invitationID, userID string) error {
	invitation, err := s.repo.GetInvitation(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != "pending" {
		return fmt.Errorf("invitation is not pending")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("invitation has expired")
	}

	// Update invitation status
	invitation.Status = "accepted"
	invitation.AcceptedAt = &[]time.Time{time.Now()}[0]

	if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Add member
	member := &organization.Member{
		OrganizationID: invitation.OrganizationID,
		UserID:         userID,
		Email:          invitation.Email,
		Role:           invitation.Role,
		JoinedAt:       time.Now(),
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

func (s *service) RevokeInvitation(ctx context.Context, invitationID string) error {
	invitation, err := s.repo.GetInvitation(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != "pending" {
		return fmt.Errorf("can only revoke pending invitations")
	}

	// Update invitation status
	invitation.Status = "revoked"

	if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	return nil
}

func (s *service) ListInvitations(ctx context.Context, orgID string) ([]*organization.Invitation, error) {
	return s.repo.ListInvitations(ctx, orgID)
}

func (s *service) LogActivity(ctx context.Context, activity *organization.Activity) error {
	activity.ID = uuid.New().String()
	activity.Timestamp = time.Now()

	if err := s.repo.CreateActivity(ctx, activity); err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}

func (s *service) GetActivityLogs(ctx context.Context, orgID string, filter organization.ActivityFilter) ([]*organization.Activity, error) {
	filter.OrganizationID = orgID
	return s.repo.ListActivities(ctx, filter)
}