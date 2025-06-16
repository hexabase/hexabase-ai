package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/db"
	"github.com/hexabase/hexabase-ai/api/internal/domain/workspace"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL workspace repository
func NewPostgresRepository(db *gorm.DB) workspace.Repository {
	return &postgresRepository{db: db}
}

// toDTO converts domain workspace model to database model
func toDTO(domainWs *workspace.Workspace) *db.Workspace {
	dbWs := &db.Workspace{
		ID:             domainWs.ID,
		OrganizationID: domainWs.OrganizationID,
		Name:           domainWs.Name,
		PlanID:         domainWs.PlanID,
		CreatedAt:      domainWs.CreatedAt,
		UpdatedAt:      domainWs.UpdatedAt,
	}

	// Convert VClusterName
	if domainWs.VClusterName != "" {
		dbWs.VClusterInstanceName = &domainWs.VClusterName
	}

	// Convert Status
	dbWs.VClusterStatus = toDTOStatus(domainWs.Status)

	// Convert Settings and ClusterInfo to VClusterConfig JSON
	vclusterConfig := make(map[string]interface{})
	if domainWs.Settings != nil {
		for k, v := range domainWs.Settings {
			vclusterConfig[k] = v
		}
	}
	if domainWs.ClusterInfo != nil {
		for k, v := range domainWs.ClusterInfo {
			vclusterConfig[k] = v
		}
	}
	if domainWs.KubeConfig != "" {
		vclusterConfig["kubeconfig"] = domainWs.KubeConfig
	}
	if domainWs.APIEndpoint != "" {
		vclusterConfig["api_endpoint"] = domainWs.APIEndpoint
	}
	if domainWs.Namespace != "" {
		vclusterConfig["namespace"] = domainWs.Namespace
	}

	if len(vclusterConfig) > 0 {
		if configBytes, err := json.Marshal(vclusterConfig); err == nil {
			dbWs.VClusterConfig = string(configBytes)
		} else {
			dbWs.VClusterConfig = "{}"
		}
	} else {
		dbWs.VClusterConfig = "{}"
	}

	// Convert Metadata to DedicatedNodeConfig JSON
	if domainWs.Metadata != nil && len(domainWs.Metadata) > 0 {
		if metadataBytes, err := json.Marshal(domainWs.Metadata); err == nil {
			dbWs.DedicatedNodeConfig = string(metadataBytes)
		} else {
			dbWs.DedicatedNodeConfig = "{}"
		}
	} else {
		dbWs.DedicatedNodeConfig = "{}"
	}

	return dbWs
}

// toDomainModel converts database workspace model to domain model
func toDomainModel(dbWs *db.Workspace) (*workspace.Workspace, error) {
	domainWs := &workspace.Workspace{
		ID:             dbWs.ID,
		OrganizationID: dbWs.OrganizationID,
		Name:           dbWs.Name,
		PlanID:         dbWs.PlanID,
		CreatedAt:      dbWs.CreatedAt,
		UpdatedAt:      dbWs.UpdatedAt,
	}

	// Convert VClusterInstanceName
	if dbWs.VClusterInstanceName != nil {
		domainWs.VClusterName = *dbWs.VClusterInstanceName
	}

	// Convert Status
	domainWs.Status = toDomainModelStatus(dbWs.VClusterStatus)

	// Initialize Settings and ClusterInfo maps
	domainWs.Settings = make(map[string]interface{})
	domainWs.ClusterInfo = make(map[string]interface{})

	// Parse VClusterConfig JSON
	if dbWs.VClusterConfig != "" && dbWs.VClusterConfig != "{}" {
		var vclusterConfig map[string]interface{}
		if err := json.Unmarshal([]byte(dbWs.VClusterConfig), &vclusterConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal VClusterConfig: %w", err)
		}

		for k, v := range vclusterConfig {
			switch k {
			case "kubeconfig":
				if str, ok := v.(string); ok {
					domainWs.KubeConfig = str
				}
			case "api_endpoint":
				if str, ok := v.(string); ok {
					domainWs.APIEndpoint = str
				}
			case "namespace":
				if str, ok := v.(string); ok {
					domainWs.Namespace = str
				}
			case "nodes", "version", "capacity":
				// ClusterInfo fields
				domainWs.ClusterInfo[k] = v
			case "autoscaling", "replicas", "resources":
				// Settings fields
				domainWs.Settings[k] = v
			default:
				// Default to Settings for unknown fields
				domainWs.Settings[k] = v
			}
		}
	}

	// Parse DedicatedNodeConfig JSON to Metadata
	if dbWs.DedicatedNodeConfig != "" && dbWs.DedicatedNodeConfig != "{}" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(dbWs.DedicatedNodeConfig), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal DedicatedNodeConfig: %w", err)
		}
		domainWs.Metadata = metadata
	}

	return domainWs, nil
}

// toDTOStatus converts domain status to database status
func toDTOStatus(domainStatus string) string {
	switch domainStatus {
	case "creating", "provisioning":
		return "PENDING_CREATION"
	case "active", "running":
		return "RUNNING"
	case "updating":
		return "UPDATING_PLAN"
	case "deleting":
		return "DELETING"
	case "error", "failed":
		return "ERROR"
	case "stopped":
		return "STOPPED"
	case "starting":
		return "STARTING"
	case "stopping":
		return "STOPPING"
	default:
		return "UNKNOWN"
	}
}

// toDomainModelStatus converts database status to domain status
func toDomainModelStatus(dbStatus string) string {
	switch dbStatus {
	case "PENDING_CREATION", "CONFIGURING_HNC":
		return "creating"
	case "RUNNING":
		return "active"
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

func (r *postgresRepository) CreateWorkspace(ctx context.Context, ws *workspace.Workspace) error {
	dbWs := toDTO(ws)
	if err := r.db.WithContext(ctx).Create(dbWs).Error; err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error) {
	var dbWs db.Workspace
	if err := r.db.WithContext(ctx).Where("id = ?", workspaceID).First(&dbWs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("workspace not found")
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	
	domainWs, err := toDomainModel(&dbWs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert database model to domain: %w", err)
	}
	
	return domainWs, nil
}

func (r *postgresRepository) GetWorkspaceByNameAndOrg(ctx context.Context, name, orgID string) (*workspace.Workspace, error) {
	var dbWs db.Workspace
	if err := r.db.WithContext(ctx).
		Where("name = ? AND organization_id = ?", name, orgID).
		First(&dbWs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workspace by name and org: %w", err)
	}
	
	domainWs, err := toDomainModel(&dbWs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert database model to domain: %w", err)
	}
	
	return domainWs, nil
}

func (r *postgresRepository) UpdateWorkspace(ctx context.Context, ws *workspace.Workspace) error {
	dbWs := toDTO(ws)
	if err := r.db.WithContext(ctx).Save(dbWs).Error; err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", workspaceID).Delete(&db.Workspace{}).Error; err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListWorkspaces(ctx context.Context, filter workspace.WorkspaceFilter) ([]*workspace.Workspace, int, error) {
	var dbWorkspaces []db.Workspace
	var total int64

	query := r.db.WithContext(ctx).Model(&db.Workspace{})

	if filter.OrganizationID != "" {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}

	if filter.Status != "" {
		dbStatus := toDTOStatus(filter.Status)
		query = query.Where("v_cluster_status = ?", dbStatus)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Search+"%")
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

	if err := query.Find(&dbWorkspaces).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list workspaces: %w", err)
	}

	// Convert database models to domain models
	domainWorkspaces := make([]*workspace.Workspace, len(dbWorkspaces))
	for i, dbWs := range dbWorkspaces {
		domainWs, err := toDomainModel(&dbWs)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert workspace %s: %w", dbWs.ID, err)
		}
		domainWorkspaces[i] = domainWs
	}

	return domainWorkspaces, int(total), nil
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
		Delete(&db.Workspace{}).Error
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
	// For now, we'll store it in the workspace vcluster_config
	return r.db.WithContext(ctx).
		Model(&db.Workspace{}).
		Where("id = ?", workspaceID).
		Update("vcluster_config", gorm.Expr("jsonb_set(COALESCE(vcluster_config, '{}'::jsonb), '{kubeconfig}', ?)", kubeconfig)).
		Error
}

// GetKubeconfig retrieves the kubeconfig for a workspace
func (r *postgresRepository) GetKubeconfig(ctx context.Context, workspaceID string) (string, error) {
	var result struct {
		Kubeconfig string
	}
	
	if err := r.db.WithContext(ctx).
		Model(&db.Workspace{}).
		Where("id = ?", workspaceID).
		Select("vcluster_config->>'kubeconfig' as kubeconfig").
		Scan(&result).Error; err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}
	
	return result.Kubeconfig, nil
}