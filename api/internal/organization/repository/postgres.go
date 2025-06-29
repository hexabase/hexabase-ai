package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/organization/domain"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL organization repository
func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &postgresRepository{db: db}
}

// Organization operations

func (r *postgresRepository) CreateOrganization(ctx context.Context, org *domain.Organization) error {
	dbOrg := domainToDBOrganization(org)
	if err := r.db.WithContext(ctx).Create(dbOrg).Error; err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	// Update the domain model with any generated values
	org.CreatedAt = dbOrg.CreatedAt
	org.UpdatedAt = dbOrg.UpdatedAt
	org.ID = dbOrg.ID
	return nil
}

func (r *postgresRepository) GetOrganization(ctx context.Context, orgID string) (*domain.Organization, error) {
	var dbOrg dbOrganization
	if err := r.db.WithContext(ctx).Where("id = ?", orgID).First(&dbOrg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return dbToDomainOrganization(&dbOrg), nil
}

func (r *postgresRepository) GetOrganizationByName(ctx context.Context, name string) (*domain.Organization, error) {
	var dbOrg dbOrganization
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&dbOrg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization by name: %w", err)
	}
	return dbToDomainOrganization(&dbOrg), nil
}

func (r *postgresRepository) ListOrganizations(ctx context.Context, filter domain.OrganizationFilter) ([]*domain.Organization, int, error) {
	var dbOrgs []dbOrganization
	var total int64

	query := r.db.WithContext(ctx).Model(&dbOrganization{})

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

	if err := query.Order(orderBy).Find(&dbOrgs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list organizations: %w", err)
	}

	orgs := make([]*domain.Organization, len(dbOrgs))
	for i, dbOrg := range dbOrgs {
		orgs[i] = dbToDomainOrganization(&dbOrg)
	}

	return orgs, int(total), nil
}

func (r *postgresRepository) UpdateOrganization(ctx context.Context, org *domain.Organization) error {
	dbOrg := domainToDBOrganization(org)
	if err := r.db.WithContext(ctx).Model(&dbOrganization{ID: org.ID}).Updates(dbOrg).Error; err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteOrganization(ctx context.Context, orgID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", orgID).Delete(&dbOrganization{}).Error; err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

// Member operations

func (r *postgresRepository) AddMember(ctx context.Context, member *domain.OrganizationUser) error {
	dbMember := domainToDBOrganizationUser(member)
	if err := r.db.WithContext(ctx).Create(dbMember).Error; err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	// Update the domain model with any generated values
	member.JoinedAt = dbMember.JoinedAt
	return nil
}

func (r *postgresRepository) GetMember(ctx context.Context, orgID, userID string) (*domain.OrganizationUser, error) {
	var member domain.OrganizationUser
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

func (r *postgresRepository) ListMembers(ctx context.Context, filter domain.MemberFilter) ([]*domain.OrganizationUser, int, error) {
	var members []*domain.OrganizationUser
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.OrganizationUser{})

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

func (r *postgresRepository) UpdateMember(ctx context.Context, member *domain.OrganizationUser) error {
	if err := r.db.WithContext(ctx).Save(member).Error; err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}
	return nil
}

func (r *postgresRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&domain.OrganizationUser{}).Error; err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.OrganizationUser{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("role", role).Error; err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}
	return nil
}

func (r *postgresRepository) CountMembers(ctx context.Context, orgID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.OrganizationUser{}).
		Where("organization_id = ? AND status = ?", orgID, "active").
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count members: %w", err)
	}
	return int(count), nil
}

// User operations

func (r *postgresRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *postgresRepository) GetUsersByIDs(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	var users []*domain.User
	if err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}
	return users, nil
}

// Invitation operations

func (r *postgresRepository) CreateInvitation(ctx context.Context, invitation *domain.Invitation) error {
	if err := r.db.WithContext(ctx).Create(invitation).Error; err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetInvitation(ctx context.Context, invitationID string) (*domain.Invitation, error) {
	var invitation domain.Invitation
	if err := r.db.WithContext(ctx).Where("id = ?", invitationID).First(&invitation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invitation not found")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	return &invitation, nil
}

func (r *postgresRepository) GetInvitationByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	var invitation domain.Invitation
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&invitation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invitation by token: %w", err)
	}
	return &invitation, nil
}

func (r *postgresRepository) ListInvitations(ctx context.Context, orgID string, status string) ([]*domain.Invitation, error) {
	var invitations []*domain.Invitation
	
	query := r.db.WithContext(ctx).Where("organization_id = ?", orgID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Order("created_at DESC").Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	
	return invitations, nil
}

func (r *postgresRepository) UpdateInvitation(ctx context.Context, invitation *domain.Invitation) error {
	if err := r.db.WithContext(ctx).Save(invitation).Error; err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteInvitation(ctx context.Context, invitationID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", invitationID).Delete(&domain.Invitation{}).Error; err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteExpiredInvitations(ctx context.Context, before time.Time) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", before, "pending").
		Delete(&domain.Invitation{}).Error; err != nil {
		return fmt.Errorf("failed to delete expired invitations: %w", err)
	}
	return nil
}

// Statistics operations

func (r *postgresRepository) GetOrganizationStats(ctx context.Context, orgID string) (*domain.OrganizationStats, error) {
	stats := &domain.OrganizationStats{
		OrganizationID: orgID,
		LastUpdated:    time.Now(),
	}

	// Count total members
	var totalMembers int64
	if err := r.db.WithContext(ctx).
		Model(&domain.OrganizationUser{}).
		Where("organization_id = ?", orgID).
		Count(&totalMembers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total members: %w", err)
	}
	stats.TotalMembers = int(totalMembers)

	// Count active members
	var activeMembers int64
	if err := r.db.WithContext(ctx).
		Model(&domain.OrganizationUser{}).
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

// Status conversion helper functions

// toDomainStatus converts database v_cluster_status to domain status
func toDomainStatus(dbStatus string) string {
	switch dbStatus {
	case "RUNNING":
		return "active"
	case "PENDING_CREATION", "CONFIGURING_HNC":
		return "creating"
	case "UPDATING_PLAN", "UPDATING_NODES":
		return "updating"
	case "DELETING":
		return "deleting"
	case "ERROR":
		return "error"
	case "STOPPED":
		return "stopped"
	case "STARTING":
		return "starting"
	case "STOPPING":
		return "stopping"
	case "UNKNOWN":
		return "unknown"
	default:
		return "unknown"
	}
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

	// Count active workspaces using correct column name
	var activeCount int64
	if err := r.db.WithContext(ctx).
		Table("workspaces").
		Where("organization_id = ? AND v_cluster_status = ?", orgID, "RUNNING").
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

func (r *postgresRepository) GetResourceUsage(ctx context.Context, orgID string) (*domain.Usage, error) {
	// This would aggregate resource usage across all workspaces
	// For now, return placeholder data
	usage := &domain.Usage{
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

func (r *postgresRepository) CreateActivity(ctx context.Context, activity *domain.Activity) error {
	if err := r.db.WithContext(ctx).Create(activity).Error; err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListActivities(ctx context.Context, filter domain.ActivityFilter) ([]*domain.Activity, error) {
	var activities []*domain.Activity

	query := r.db.WithContext(ctx).Model(&domain.Activity{})

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

func (r *postgresRepository) ListWorkspaces(ctx context.Context, orgID string) ([]*domain.WorkspaceInfo, error) {
	type dbWorkspaceInfo struct {
		ID             string `gorm:"column:id"`
		Name           string `gorm:"column:name"`
		VClusterStatus string `gorm:"column:v_cluster_status"`
	}
	
	var dbWorkspaces []dbWorkspaceInfo
	
	// Query workspaces table with correct column name
	if err := r.db.WithContext(ctx).
		Table("workspaces").
		Select("id, name, v_cluster_status").
		Where("organization_id = ?", orgID).
		Scan(&dbWorkspaces).Error; err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	// Convert to domain model
	workspaces := make([]*domain.WorkspaceInfo, len(dbWorkspaces))
	for i, dbWs := range dbWorkspaces {
		workspaces[i] = &domain.WorkspaceInfo{
			ID:     dbWs.ID,
			Name:   dbWs.Name,
			Status: toDomainStatus(dbWs.VClusterStatus),
		}
	}

	return workspaces, nil
}