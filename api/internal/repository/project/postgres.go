package project

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/project"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL project repository
func NewPostgresRepository(db *gorm.DB) project.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateProject(ctx context.Context, proj *project.Project) error {
	if err := r.db.WithContext(ctx).Create(proj).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetProject(ctx context.Context, projectID string) (*project.Project, error) {
	var proj project.Project
	if err := r.db.WithContext(ctx).Where("id = ?", projectID).First(&proj).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &proj, nil
}

func (r *postgresRepository) GetProjectByName(ctx context.Context, name string) (*project.Project, error) {
	var proj project.Project
	if err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&proj).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not an error, just not found
		}
		return nil, fmt.Errorf("failed to get project by name: %w", err)
	}
	return &proj, nil
}

func (r *postgresRepository) UpdateProject(ctx context.Context, proj *project.Project) error {
	if err := r.db.WithContext(ctx).Save(proj).Error; err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteProject(ctx context.Context, projectID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", projectID).Delete(&project.Project{}).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}


func (r *postgresRepository) GetChildProjects(ctx context.Context, parentID string) ([]*project.Project, error) {
	var projects []*project.Project
	if err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("name ASC").
		Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to get child projects: %w", err)
	}
	return projects, nil
}

func (r *postgresRepository) AddProjectMember(ctx context.Context, member *project.ProjectMember) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to add project member: %w", err)
	}
	return nil
}

func (r *postgresRepository) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&project.ProjectMember{}).Error; err != nil {
		return fmt.Errorf("failed to remove project member: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListProjectMembers(ctx context.Context, projectID string) ([]*project.ProjectMember, error) {
	var members []*project.ProjectMember
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to list project members: %w", err)
	}
	return members, nil
}

func (r *postgresRepository) CreateActivity(ctx context.Context, activity *project.Activity) error {
	if err := r.db.WithContext(ctx).Create(activity).Error; err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListActivities(ctx context.Context, filter project.ActivityFilter) ([]*project.Activity, error) {
	var activities []*project.Activity

	query := r.db.WithContext(ctx).Model(&project.Activity{})

	if filter.ProjectID != "" {
		query = query.Where("project_id = ?", filter.ProjectID)
	}


	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	if filter.PageSize > 0 {
		query = query.Limit(filter.PageSize)
	}
	
	if filter.Page > 0 {
		query = query.Offset((filter.Page - 1) * filter.PageSize)
	}

	if err := query.Order("created_at DESC").Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	return activities, nil
}

// AddMember adds a member to a project
func (r *postgresRepository) AddMember(ctx context.Context, member *project.ProjectMember) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

// GetMember gets a project member by project ID and user ID
func (r *postgresRepository) GetMember(ctx context.Context, projectID, userID string) (*project.ProjectMember, error) {
	var member project.ProjectMember
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("member not found")
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}
	return &member, nil
}

// GetMemberByID gets a project member by ID
func (r *postgresRepository) GetMemberByID(ctx context.Context, memberID string) (*project.ProjectMember, error) {
	var member project.ProjectMember
	if err := r.db.WithContext(ctx).Where("id = ?", memberID).First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("member not found")
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}
	return &member, nil
}

// ListMembers lists all members of a project
func (r *postgresRepository) ListMembers(ctx context.Context, projectID string) ([]*project.ProjectMember, error) {
	var members []*project.ProjectMember
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	return members, nil
}

// UpdateMember updates a project member
func (r *postgresRepository) UpdateMember(ctx context.Context, member *project.ProjectMember) error {
	if err := r.db.WithContext(ctx).Save(member).Error; err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}
	return nil
}

// RemoveMember removes a member from a project
func (r *postgresRepository) RemoveMember(ctx context.Context, memberID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", memberID).Delete(&project.ProjectMember{}).Error; err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

// CountMembers counts the number of members in a project
func (r *postgresRepository) CountMembers(ctx context.Context, projectID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&project.ProjectMember{}).
		Where("project_id = ?", projectID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count members: %w", err)
	}
	return int(count), nil
}

// GetUserByEmail gets a user by email
func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (*project.User, error) {
	var user project.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetLastActivity gets the last activity for a project
func (r *postgresRepository) GetLastActivity(ctx context.Context, projectID string) (*project.Activity, error) {
	var activity project.Activity
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last activity: %w", err)
	}
	return &activity, nil
}

// GetProjectResourceUsage gets resource usage for a project
func (r *postgresRepository) GetProjectResourceUsage(ctx context.Context, projectID string) (*project.ResourceUsage, error) {
	// This would typically aggregate usage from Kubernetes
	// For now, return a placeholder
	return &project.ResourceUsage{
		CPU:    "0",
		Memory: "0",
		Pods:   0,
	}, nil
}

// CreateNamespace creates a namespace record
func (r *postgresRepository) CreateNamespace(ctx context.Context, ns *project.Namespace) error {
	if err := r.db.WithContext(ctx).Create(ns).Error; err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}
	return nil
}

// GetNamespace gets a namespace by ID
func (r *postgresRepository) GetNamespace(ctx context.Context, namespaceID string) (*project.Namespace, error) {
	var ns project.Namespace
	if err := r.db.WithContext(ctx).Where("id = ?", namespaceID).First(&ns).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("namespace not found")
		}
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}
	return &ns, nil
}

// ListNamespaces lists namespaces for a project
func (r *postgresRepository) ListNamespaces(ctx context.Context, projectID string) ([]*project.Namespace, error) {
	var namespaces []*project.Namespace
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&namespaces).Error; err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	return namespaces, nil
}

// UpdateNamespace updates a namespace
func (r *postgresRepository) UpdateNamespace(ctx context.Context, ns *project.Namespace) error {
	if err := r.db.WithContext(ctx).Save(ns).Error; err != nil {
		return fmt.Errorf("failed to update namespace: %w", err)
	}
	return nil
}

// DeleteNamespace deletes a namespace record
func (r *postgresRepository) DeleteNamespace(ctx context.Context, namespaceID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", namespaceID).Delete(&project.Namespace{}).Error; err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}
	return nil
}

// ListProjects lists projects with filter
func (r *postgresRepository) ListProjects(ctx context.Context, filter project.ProjectFilter) ([]*project.Project, int, error) {
	var projects []*project.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&project.Project{})

	if filter.WorkspaceID != "" {
		query = query.Where("workspace_id = ?", filter.WorkspaceID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Count total results
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// Apply pagination
	if filter.PageSize > 0 {
		query = query.Limit(filter.PageSize)
		if filter.Page > 0 {
			query = query.Offset((filter.Page - 1) * filter.PageSize)
		}
	}

	// Apply sorting
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := filter.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	if err := query.Find(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, int(total), nil
}

// GetProjectByNameAndWorkspace gets a project by name and workspace ID
func (r *postgresRepository) GetProjectByNameAndWorkspace(ctx context.Context, name, workspaceID string) (*project.Project, error) {
	var proj project.Project
	if err := r.db.WithContext(ctx).
		Where("name = ? AND workspace_id = ?", name, workspaceID).
		First(&proj).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &proj, nil
}

// GetNamespaceResourceUsage gets resource usage for a namespace
func (r *postgresRepository) GetNamespaceResourceUsage(ctx context.Context, namespaceID string) (*project.NamespaceUsage, error) {
	// This would typically aggregate usage from Kubernetes
	// For now, return a placeholder
	return &project.NamespaceUsage{
		CPU:     "0",
		Memory:  "0",
		Storage: "0",
		Pods:    0,
	}, nil
}

// CleanupOldActivities cleans up old activities based on retention policy
func (r *postgresRepository) CleanupOldActivities(ctx context.Context, before time.Time) error {
	if err := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&project.Activity{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old activities: %w", err)
	}
	return nil
}

// CountProjects counts the total number of projects for a workspace
func (r *postgresRepository) CountProjects(ctx context.Context, workspaceID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&project.Project{}).
		Where("workspace_id = ?", workspaceID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count projects: %w", err)
	}
	return int(count), nil
}

// GetNamespaceByName gets a namespace by name and project ID
func (r *postgresRepository) GetNamespaceByName(ctx context.Context, projectID, name string) (*project.Namespace, error) {
	var ns project.Namespace
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID, name).
		First(&ns).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}
	return &ns, nil
}

// GetUser gets a user by ID
func (r *postgresRepository) GetUser(ctx context.Context, userID string) (*project.User, error) {
	var user project.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}