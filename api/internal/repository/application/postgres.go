package application

import (
	"context"
	"fmt"

	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"gorm.io/gorm"
)

// PostgresRepository implements the application repository interface using PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *gorm.DB) application.Repository {
	return &PostgresRepository{db: db}
}

// CreateApplication creates a new application
func (r *PostgresRepository) CreateApplication(ctx context.Context, app *application.Application) error {
	return r.db.WithContext(ctx).Create(app).Error
}

// GetApplication retrieves an application by ID
func (r *PostgresRepository) GetApplication(ctx context.Context, id string) (*application.Application, error) {
	var app application.Application
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// GetApplicationByName retrieves an application by name within a workspace and project
func (r *PostgresRepository) GetApplicationByName(ctx context.Context, workspaceID, projectID, name string) (*application.Application, error) {
	var app application.Application
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND project_id = ? AND name = ?", workspaceID, projectID, name).
		First(&app).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

// ListApplications lists all applications in a workspace/project
func (r *PostgresRepository) ListApplications(ctx context.Context, workspaceID, projectID string) ([]application.Application, error) {
	var apps []application.Application
	query := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	err := query.Order("created_at DESC").Find(&apps).Error
	return apps, err
}

// UpdateApplication updates an application
func (r *PostgresRepository) UpdateApplication(ctx context.Context, app *application.Application) error {
	return r.db.WithContext(ctx).Save(app).Error
}

// DeleteApplication deletes an application
func (r *PostgresRepository) DeleteApplication(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&application.Application{}).Error
}

// CreateEvent creates a new application event
func (r *PostgresRepository) CreateEvent(ctx context.Context, event *application.ApplicationEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// ListEvents lists events for an application
func (r *PostgresRepository) ListEvents(ctx context.Context, applicationID string, limit int) ([]application.ApplicationEvent, error) {
	var events []application.ApplicationEvent
	query := r.db.WithContext(ctx).Where("application_id = ?", applicationID).Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&events).Error
	return events, err
}

// GetApplicationsByNode retrieves applications running on a specific node
func (r *PostgresRepository) GetApplicationsByNode(ctx context.Context, nodeID string) ([]application.Application, error) {
	var apps []application.Application
	// Use JSON query for node selector
	err := r.db.WithContext(ctx).
		Where("config->'node_selector' @> ?", fmt.Sprintf(`{"node-id": "%s"}`, nodeID)).
		Find(&apps).Error
	return apps, err
}

// GetApplicationsByStatus retrieves applications by status
func (r *PostgresRepository) GetApplicationsByStatus(ctx context.Context, workspaceID string, status application.ApplicationStatus) ([]application.Application, error) {
	var apps []application.Application
	query := r.db.WithContext(ctx).Where("status = ?", status)
	if workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}
	err := query.Find(&apps).Error
	return apps, err
}