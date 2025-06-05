package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-kaas/api/internal/domain/workspace"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL workspace repository
func NewPostgresRepository(db *gorm.DB) workspace.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateWorkspace(ctx context.Context, ws *workspace.Workspace) error {
	if err := r.db.WithContext(ctx).Create(ws).Error; err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error) {
	var ws workspace.Workspace
	if err := r.db.WithContext(ctx).Where("id = ?", workspaceID).First(&ws).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("workspace not found")
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	return &ws, nil
}

func (r *postgresRepository) GetWorkspaceByNameAndOrg(ctx context.Context, name, orgID string) (*workspace.Workspace, error) {
	var ws workspace.Workspace
	if err := r.db.WithContext(ctx).
		Where("name = ? AND organization_id = ?", name, orgID).
		First(&ws).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workspace by name and org: %w", err)
	}
	return &ws, nil
}

func (r *postgresRepository) UpdateWorkspace(ctx context.Context, ws *workspace.Workspace) error {
	if err := r.db.WithContext(ctx).Save(ws).Error; err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", workspaceID).Delete(&workspace.Workspace{}).Error; err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListWorkspaces(ctx context.Context, filter workspace.WorkspaceFilter) ([]*workspace.Workspace, int, error) {
	var workspaces []*workspace.Workspace
	var total int64

	query := r.db.WithContext(ctx).Model(&workspace.Workspace{})

	if filter.OrganizationID != "" {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", 
			"%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count workspaces: %w", err)
	}

	// Apply pagination
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Apply sorting
	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortOrder == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	if err := query.Find(&workspaces).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list workspaces: %w", err)
	}

	return workspaces, int(total), nil
}

func (r *postgresRepository) AddWorkspaceMember(ctx context.Context, member *workspace.WorkspaceMember) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to add workspace member: %w", err)
	}
	return nil
}

func (r *postgresRepository) RemoveWorkspaceMember(ctx context.Context, workspaceID, userID string) error {
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Delete(&workspace.WorkspaceMember{}).Error; err != nil {
		return fmt.Errorf("failed to remove workspace member: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListWorkspaceMembers(ctx context.Context, workspaceID string) ([]*workspace.WorkspaceMember, error) {
	var members []*workspace.WorkspaceMember
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to list workspace members: %w", err)
	}
	return members, nil
}

func (r *postgresRepository) CreateTask(ctx context.Context, task *workspace.Task) error {
	if err := r.db.WithContext(ctx).Create(task).Error; err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetTask(ctx context.Context, taskID string) (*workspace.Task, error) {
	var task workspace.Task
	if err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

func (r *postgresRepository) UpdateTask(ctx context.Context, task *workspace.Task) error {
	if err := r.db.WithContext(ctx).Save(task).Error; err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListTasks(ctx context.Context, workspaceID string) ([]*workspace.Task, error) {
	var tasks []*workspace.Task
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

func (r *postgresRepository) GetPendingTasks(ctx context.Context, taskType string, limit int) ([]*workspace.Task, error) {
	var tasks []*workspace.Task
	query := r.db.WithContext(ctx).Where("status = ?", "pending")
	
	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Order("created_at ASC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}
	return tasks, nil
}

func (r *postgresRepository) CreateResourceUsage(ctx context.Context, usage *workspace.ResourceUsage) error {
	if err := r.db.WithContext(ctx).Create(usage).Error; err != nil {
		return fmt.Errorf("failed to create resource usage: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetResourceUsageHistory(ctx context.Context, workspaceID string, limit int) ([]*workspace.ResourceUsage, error) {
	var usages []*workspace.ResourceUsage
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&usages).Error; err != nil {
		return nil, fmt.Errorf("failed to get resource usage history: %w", err)
	}
	return usages, nil
}

// CleanupExpiredTasks removes expired tasks
func (r *postgresRepository) CleanupExpiredTasks(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("updated_at < ? AND status IN ?", before, []string{"completed", "failed"}).
		Delete(&workspace.Task{}).Error
}

// CleanupDeletedWorkspaces removes deleted workspaces
func (r *postgresRepository) CleanupDeletedWorkspaces(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("deleted_at IS NOT NULL AND deleted_at < ?", before).
		Delete(&workspace.Workspace{}).Error
}

// SaveWorkspaceStatus saves the workspace status
func (r *postgresRepository) SaveWorkspaceStatus(ctx context.Context, status *workspace.WorkspaceStatus) error {
	return r.db.WithContext(ctx).Save(status).Error
}

// GetWorkspaceStatus retrieves the workspace status
func (r *postgresRepository) GetWorkspaceStatus(ctx context.Context, workspaceID string) (*workspace.WorkspaceStatus, error) {
	var status workspace.WorkspaceStatus
	if err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).First(&status).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workspace status: %w", err)
	}
	return &status, nil
}

// SaveKubeconfig saves the kubeconfig for a workspace
func (r *postgresRepository) SaveKubeconfig(ctx context.Context, workspaceID, kubeconfig string) error {
	// In a real implementation, this would be stored securely, potentially encrypted
	// For now, we'll store it in the workspace metadata
	return r.db.WithContext(ctx).
		Model(&workspace.Workspace{}).
		Where("id = ?", workspaceID).
		Update("metadata", gorm.Expr("jsonb_set(COALESCE(metadata, '{}'::jsonb), '{kubeconfig}', ?)", kubeconfig)).
		Error
}

// GetKubeconfig retrieves the kubeconfig for a workspace
func (r *postgresRepository) GetKubeconfig(ctx context.Context, workspaceID string) (string, error) {
	var result struct {
		Kubeconfig string
	}
	
	if err := r.db.WithContext(ctx).
		Model(&workspace.Workspace{}).
		Where("id = ?", workspaceID).
		Select("metadata->>'kubeconfig' as kubeconfig").
		Scan(&result).Error; err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}
	
	return result.Kubeconfig, nil
}