package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/organization"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL organization repository
func NewPostgresRepository(db *gorm.DB) organization.Repository {
	return &postgresRepository{db: db}
}

// Organization operations

func (r *postgresRepository) CreateOrganization(ctx context.Context, org *organization.Organization) error {
	dbOrg := domainToDBOrganization(org)
	if err := r.db.WithContext(ctx).Create(dbOrg).Error; err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	// Update the domain model with any generated values
	org.CreatedAt = dbOrg.CreatedAt
	org.UpdatedAt = dbOrg.UpdatedAt
	return nil
}

func (r *postgresRepository) GetOrganization(ctx context.Context, orgID string) (*organization.Organization, error) {
	var dbOrg dbOrganization
	if err := r.db.WithContext(ctx).Where("id = ?", orgID).First(&dbOrg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return dbToDomainOrganization(&dbOrg), nil
}

func (r *postgresRepository) GetOrganizationByName(ctx context.Context, name string) (*organization.Organization, error) {
	var org organization.Organization
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization by name: %w", err)
	}
	return &org, nil
}

func (r *postgresRepository) ListOrganizations(ctx context.Context, filter organization.OrganizationFilter) ([]*organization.Organization, int, error) {
	var orgs []*organization.Organization
	var total int64

	query := r.db.WithContext(ctx).Model(&organization.Organization{})

	// Apply filters
	if filter.UserID != "" {
		query = query.Joins("JOIN organization_users ON organizations.id = organization_users.organization_id").
			Where("organization_users.user_id = ?", filter.UserID)
	}

	if filter.OwnerID != "" {
		query = query.Joins("JOIN organization_users ON organizations.id = organization_users.organization_id").
			Where("organization_users.user_id = ? AND organization_users.role = ?", filter.OwnerID, "owner")
	}

	if filter.Status != "" {
		query = query.Where("organizations.status = ?", filter.Status)
	}

	if filter.Search != "" {
		query = query.Where("organizations.name ILIKE ? OR organizations.display_name ILIKE ?", 
			"%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count organizations: %w", err)
	}

	// Apply pagination
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Apply sorting
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
		if filter.SortOrder != "" {
			orderBy += " " + filter.SortOrder
		}
	}

	if err := query.Order(orderBy).Find(&orgs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list organizations: %w", err)
	}

	return orgs, int(total), nil
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

// Member operations

func (r *postgresRepository) AddMember(ctx context.Context, member *organization.OrganizationUser) error {
	dbMember := domainToDBOrganizationUser(member)
	if err := r.db.WithContext(ctx).Create(dbMember).Error; err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	// Update the domain model with any generated values
	member.JoinedAt = dbMember.JoinedAt
	return nil
}

func (r *postgresRepository) GetMember(ctx context.Context, orgID, userID string) (*organization.OrganizationUser, error) {
	var member organization.OrganizationUser
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}
	return &member, nil
}

func (r *postgresRepository) ListMembers(ctx context.Context, filter organization.MemberFilter) ([]*organization.OrganizationUser, int, error) {
	var members []*organization.OrganizationUser
	var total int64

	query := r.db.WithContext(ctx).Model(&organization.OrganizationUser{})

	if filter.OrganizationID != "" {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}

	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Search != "" {
		// Join with users table to search by email/name
		query = query.Joins("JOIN users ON organization_users.user_id = users.id").
			Where("users.email ILIKE ? OR users.display_name ILIKE ?", 
				"%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count members: %w", err)
	}

	// Apply pagination
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	if err := query.Order("joined_at ASC").Find(&members).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list members: %w", err)
	}

	return members, int(total), nil
}

func (r *postgresRepository) UpdateMember(ctx context.Context, member *organization.OrganizationUser) error {
	if err := r.db.WithContext(ctx).Save(member).Error; err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}
	return nil
}

func (r *postgresRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&organization.OrganizationUser{}).Error; err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	if err := r.db.WithContext(ctx).
		Model(&organization.OrganizationUser{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("role", role).Error; err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}
	return nil
}

func (r *postgresRepository) CountMembers(ctx context.Context, orgID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&organization.OrganizationUser{}).
		Where("organization_id = ? AND status = ?", orgID, "active").
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count members: %w", err)
	}
	return int(count), nil
}

// User operations

func (r *postgresRepository) GetUser(ctx context.Context, userID string) (*organization.User, error) {
	var user organization.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (*organization.User, error) {
	var user organization.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *postgresRepository) GetUsersByIDs(ctx context.Context, userIDs []string) ([]*organization.User, error) {
	var users []*organization.User
	if err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}
	return users, nil
}

// Invitation operations

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

func (r *postgresRepository) GetInvitationByToken(ctx context.Context, token string) (*organization.Invitation, error) {
	var invitation organization.Invitation
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&invitation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invitation by token: %w", err)
	}
	return &invitation, nil
}

func (r *postgresRepository) ListInvitations(ctx context.Context, orgID string, status string) ([]*organization.Invitation, error) {
	var invitations []*organization.Invitation
	
	query := r.db.WithContext(ctx).Where("organization_id = ?", orgID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Order("created_at DESC").Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	
	return invitations, nil
}

func (r *postgresRepository) UpdateInvitation(ctx context.Context, invitation *organization.Invitation) error {
	if err := r.db.WithContext(ctx).Save(invitation).Error; err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteInvitation(ctx context.Context, invitationID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", invitationID).Delete(&organization.Invitation{}).Error; err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteExpiredInvitations(ctx context.Context, before time.Time) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", before, "pending").
		Delete(&organization.Invitation{}).Error; err != nil {
		return fmt.Errorf("failed to delete expired invitations: %w", err)
	}
	return nil
}

// Statistics operations

func (r *postgresRepository) GetOrganizationStats(ctx context.Context, orgID string) (*organization.OrganizationStats, error) {
	stats := &organization.OrganizationStats{
		OrganizationID: orgID,
		LastUpdated:    time.Now(),
	}

	// Count total members
	var totalMembers int64
	if err := r.db.WithContext(ctx).
		Model(&organization.OrganizationUser{}).
		Where("organization_id = ?", orgID).
		Count(&totalMembers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total members: %w", err)
	}
	stats.TotalMembers = int(totalMembers)

	// Count active members
	var activeMembers int64
	if err := r.db.WithContext(ctx).
		Model(&organization.OrganizationUser{}).
		Where("organization_id = ? AND status = ?", orgID, "active").
		Count(&activeMembers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active members: %w", err)
	}
	stats.ActiveMembers = int(activeMembers)

	// Get workspace counts
	total, active, err := r.GetWorkspaceCount(ctx, orgID)
	if err != nil {
		return nil, err
	}
	stats.TotalWorkspaces = total
	stats.ActiveWorkspaces = active

	// Get project count
	projectCount, err := r.GetProjectCount(ctx, orgID)
	if err != nil {
		return nil, err
	}
	stats.TotalProjects = projectCount

	// Get resource usage
	usage, err := r.GetResourceUsage(ctx, orgID)
	if err != nil {
		return nil, err
	}
	stats.ResourceUsage = usage

	return stats, nil
}

func (r *postgresRepository) GetWorkspaceCount(ctx context.Context, orgID string) (total int, active int, err error) {
	// Count total workspaces
	var totalCount int64
	if err := r.db.WithContext(ctx).
		Table("workspaces").
		Where("organization_id = ?", orgID).
		Count(&totalCount).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to count total workspaces: %w", err)
	}

	// Count active workspaces
	var activeCount int64
	if err := r.db.WithContext(ctx).
		Table("workspaces").
		Where("organization_id = ? AND status = ?", orgID, "active").
		Count(&activeCount).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to count active workspaces: %w", err)
	}

	return int(totalCount), int(activeCount), nil
}

func (r *postgresRepository) GetProjectCount(ctx context.Context, orgID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Table("projects").
		Joins("JOIN workspaces ON projects.workspace_id = workspaces.id").
		Where("workspaces.organization_id = ?", orgID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count projects: %w", err)
	}
	return int(count), nil
}

func (r *postgresRepository) GetResourceUsage(ctx context.Context, orgID string) (*organization.Usage, error) {
	// This would aggregate resource usage across all workspaces
	// For now, return placeholder data
	usage := &organization.Usage{
		CPU:     0.0,
		Memory:  0.0,
		Storage: 0.0,
		Cost:    0.0,
	}

	// In a real implementation, you would query resource metrics
	// from the monitoring system or aggregate from workspace usage

	return usage, nil
}

// Activity operations (additional methods not in interface but used by implementation)

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

// Additional helper methods

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