package project

import (
	"context"
	"fmt"

	"github.com/hexabase/hexabase-kaas/api/internal/domain/project"
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

func (r *postgresRepository) GetProjectByName(ctx context.Context, workspaceID, name string) (*project.Project, error) {
	var proj project.Project
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND name = ?", workspaceID, name).
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

func (r *postgresRepository) ListProjects(ctx context.Context, workspaceID string) ([]*project.Project, error) {
	var projects []*project.Project
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
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

	if filter.WorkspaceID != "" {
		query = query.Where("workspace_id = ?", filter.WorkspaceID)
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