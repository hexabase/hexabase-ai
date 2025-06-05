package organization

import (
	"context"
	"fmt"

	"github.com/hexabase/hexabase-kaas/api/internal/domain/organization"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL organization repository
func NewPostgresRepository(db *gorm.DB) organization.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateOrganization(ctx context.Context, org *organization.Organization) error {
	if err := r.db.WithContext(ctx).Create(org).Error; err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetOrganization(ctx context.Context, orgID string) (*organization.Organization, error) {
	var org organization.Organization
	if err := r.db.WithContext(ctx).Where("id = ?", orgID).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return &org, nil
}

func (r *postgresRepository) UpdateOrganization(ctx context.Context, org *organization.Organization) error {
	if err := r.db.WithContext(ctx).Save(org).Error; err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteOrganization(ctx context.Context, orgID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", orgID).Delete(&organization.Organization{}).Error; err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListOrganizations(ctx context.Context, filter organization.OrganizationFilter) ([]*organization.Organization, error) {
	var orgs []*organization.Organization

	query := r.db.WithContext(ctx).Model(&organization.Organization{})

	if filter.OwnerID != "" {
		query = query.Where("owner_id = ?", filter.OwnerID)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", 
			"%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if err := query.Order("created_at DESC").Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	return orgs, nil
}

func (r *postgresRepository) AddMember(ctx context.Context, member *organization.Member) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

func (r *postgresRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&organization.Member{}).Error; err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	if err := r.db.WithContext(ctx).
		Model(&organization.Member{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("role", role).Error; err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListMembers(ctx context.Context, orgID string) ([]*organization.Member, error) {
	var members []*organization.Member
	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("joined_at ASC").
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	return members, nil
}

func (r *postgresRepository) CreateInvitation(ctx context.Context, invitation *organization.Invitation) error {
	if err := r.db.WithContext(ctx).Create(invitation).Error; err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetInvitation(ctx context.Context, invitationID string) (*organization.Invitation, error) {
	var invitation organization.Invitation
	if err := r.db.WithContext(ctx).Where("id = ?", invitationID).First(&invitation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invitation not found")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	return &invitation, nil
}

func (r *postgresRepository) UpdateInvitation(ctx context.Context, invitation *organization.Invitation) error {
	if err := r.db.WithContext(ctx).Save(invitation).Error; err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListInvitations(ctx context.Context, orgID string) ([]*organization.Invitation, error) {
	var invitations []*organization.Invitation
	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	return invitations, nil
}

func (r *postgresRepository) CreateActivity(ctx context.Context, activity *organization.Activity) error {
	if err := r.db.WithContext(ctx).Create(activity).Error; err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListActivities(ctx context.Context, filter organization.ActivityFilter) ([]*organization.Activity, error) {
	var activities []*organization.Activity

	query := r.db.WithContext(ctx).Model(&organization.Activity{})

	if filter.OrganizationID != "" {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.StartDate != nil {
		query = query.Where("timestamp >= ?", filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("timestamp <= ?", filter.EndDate)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if err := query.Order("timestamp DESC").Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	return activities, nil
}

func (r *postgresRepository) ListWorkspaces(ctx context.Context, orgID string) ([]*organization.WorkspaceInfo, error) {
	var workspaces []*organization.WorkspaceInfo
	
	// Query workspaces table
	if err := r.db.WithContext(ctx).
		Table("workspaces").
		Select("id, name, status").
		Where("organization_id = ?", orgID).
		Scan(&workspaces).Error; err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	return workspaces, nil
}