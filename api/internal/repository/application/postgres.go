package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/db"
	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/lib/pq"
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
	// For CronJob with template, fetch template app first
	if app.Type == application.ApplicationTypeCronJob && app.TemplateAppID != "" {
		templateApp, err := r.GetApplication(ctx, app.TemplateAppID)
		if err != nil {
			return fmt.Errorf("failed to get template application: %w", err)
		}
		
		// Copy source and config from template if not specified
		if app.Source.Type == "" {
			app.Source = templateApp.Source
		}
		if app.Config.Resources.CPURequest == "" {
			app.Config = templateApp.Config
		}
	}

	// Convert to DB model
	dbApp, err := r.domainToDBApp(app)
	if err != nil {
		return err
	}

	// Create in database
	if err := r.db.WithContext(ctx).Create(dbApp).Error; err != nil {
		return err
	}

	// Update app with generated values
	app.ID = dbApp.ID
	app.CreatedAt = dbApp.CreatedAt
	app.UpdatedAt = dbApp.UpdatedAt

	return nil
}

// GetApplication retrieves an application by ID
func (r *PostgresRepository) GetApplication(ctx context.Context, id string) (*application.Application, error) {
	var dbApp db.Application
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&dbApp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return r.dbToDomainApp(&dbApp)
}

// GetApplicationByName retrieves an application by name within a workspace and project
func (r *PostgresRepository) GetApplicationByName(ctx context.Context, workspaceID, projectID, name string) (*application.Application, error) {
	var dbApp db.Application
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND project_id = ? AND name = ?", workspaceID, projectID, name).
		First(&dbApp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.dbToDomainApp(&dbApp)
}

// ListApplications lists all applications in a workspace/project
func (r *PostgresRepository) ListApplications(ctx context.Context, workspaceID, projectID string) ([]application.Application, error) {
	var dbApps []db.Application
	query := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	err := query.Order("created_at DESC").Find(&dbApps).Error
	if err != nil {
		return nil, err
	}
	
	var apps []application.Application
	for _, dbApp := range dbApps {
		app, err := r.dbToDomainApp(&dbApp)
		if err != nil {
			return nil, err
		}
		apps = append(apps, *app)
	}
	return apps, nil
}

// UpdateApplication updates an application
func (r *PostgresRepository) UpdateApplication(ctx context.Context, app *application.Application) error {
	dbApp, err := r.domainToDBApp(app)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(dbApp).Error
}

// DeleteApplication deletes an application
func (r *PostgresRepository) DeleteApplication(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&db.Application{}).Error
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
	var dbApps []db.Application
	// Use JSON query for node selector
	err := r.db.WithContext(ctx).
		Where("config->'node_selector' @> ?", fmt.Sprintf(`{"node-id": "%s"}`, nodeID)).
		Find(&dbApps).Error
	if err != nil {
		return nil, err
	}
	
	var apps []application.Application
	for _, dbApp := range dbApps {
		app, err := r.dbToDomainApp(&dbApp)
		if err != nil {
			return nil, err
		}
		apps = append(apps, *app)
	}
	return apps, nil
}

// GetApplicationsByStatus retrieves applications by status
func (r *PostgresRepository) GetApplicationsByStatus(ctx context.Context, workspaceID string, status application.ApplicationStatus) ([]application.Application, error) {
	var dbApps []db.Application
	query := r.db.WithContext(ctx).Where("status = ?", string(status))
	if workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}
	err := query.Find(&dbApps).Error
	if err != nil {
		return nil, err
	}
	
	var apps []application.Application
	for _, dbApp := range dbApps {
		app, err := r.dbToDomainApp(&dbApp)
		if err != nil {
			return nil, err
		}
		apps = append(apps, *app)
	}
	return apps, nil
}

// Create is an alias for CreateApplication (for backward compatibility)
func (r *PostgresRepository) Create(ctx context.Context, app *application.Application) error {
	return r.CreateApplication(ctx, app)
}

// GetCronJobExecutions retrieves executions for a CronJob application
func (r *PostgresRepository) GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]application.CronJobExecution, int, error) {
	var executions []application.CronJobExecution
	var total int64

	// Count total executions
	if err := r.db.WithContext(ctx).
		Model(&db.CronJobExecution{}).
		Where("application_id = ?", applicationID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated executions
	var dbExecutions []db.CronJobExecution
	if err := r.db.WithContext(ctx).
		Where("application_id = ?", applicationID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbExecutions).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain models
	for _, dbExec := range dbExecutions {
		exec := application.CronJobExecution{
			ID:            dbExec.ID,
			ApplicationID: dbExec.ApplicationID,
			JobName:       dbExec.JobName,
			StartedAt:     dbExec.StartedAt,
			CompletedAt:   dbExec.CompletedAt,
			Status:        application.CronJobExecutionStatus(dbExec.Status),
			ExitCode:      dbExec.ExitCode,
			Logs:          dbExec.Logs,
			CreatedAt:     dbExec.CreatedAt,
			UpdatedAt:     dbExec.UpdatedAt,
		}
		executions = append(executions, exec)
	}

	return executions, int(total), nil
}

// CreateCronJobExecution creates a new CronJob execution record
func (r *PostgresRepository) CreateCronJobExecution(ctx context.Context, execution *application.CronJobExecution) error {
	dbExec := &db.CronJobExecution{
		ID:            execution.ID,
		ApplicationID: execution.ApplicationID,
		JobName:       execution.JobName,
		StartedAt:     execution.StartedAt,
		CompletedAt:   execution.CompletedAt,
		Status:        string(execution.Status),
		ExitCode:      execution.ExitCode,
		Logs:          execution.Logs,
		CreatedAt:     execution.CreatedAt,
		UpdatedAt:     execution.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Create(dbExec).Error; err != nil {
		return err
	}

	execution.ID = dbExec.ID
	execution.CreatedAt = dbExec.CreatedAt
	execution.UpdatedAt = dbExec.UpdatedAt
	return nil
}

// UpdateCronJobExecution updates a CronJob execution record
func (r *PostgresRepository) UpdateCronJobExecution(ctx context.Context, executionID string, completedAt *time.Time, status application.CronJobExecutionStatus, exitCode *int, logs string) error {
	updates := map[string]interface{}{
		"completed_at": completedAt,
		"status":       string(status),
		"exit_code":    exitCode,
		"logs":         logs,
		"updated_at":   time.Now(),
	}

	return r.db.WithContext(ctx).
		Model(&db.CronJobExecution{}).
		Where("id = ?", executionID).
		Updates(updates).Error
}

// UpdateCronSchedule updates the cron schedule for a CronJob application
func (r *PostgresRepository) UpdateCronSchedule(ctx context.Context, applicationID, schedule string) error {
	return r.db.WithContext(ctx).
		Model(&db.Application{}).
		Where("id = ? AND type = ?", applicationID, "cronjob").
		Update("cron_schedule", schedule).Error
}

// GetCronJobExecutionByID retrieves a single CronJob execution by ID
func (r *PostgresRepository) GetCronJobExecutionByID(ctx context.Context, executionID string) (*application.CronJobExecution, error) {
	var dbExec db.CronJobExecution
	if err := r.db.WithContext(ctx).Where("id = ?", executionID).First(&dbExec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	exec := &application.CronJobExecution{
		ID:            dbExec.ID,
		ApplicationID: dbExec.ApplicationID,
		JobName:       dbExec.JobName,
		StartedAt:     dbExec.StartedAt,
		CompletedAt:   dbExec.CompletedAt,
		Status:        application.CronJobExecutionStatus(dbExec.Status),
		ExitCode:      dbExec.ExitCode,
		Logs:          dbExec.Logs,
		CreatedAt:     dbExec.CreatedAt,
		UpdatedAt:     dbExec.UpdatedAt,
	}

	return exec, nil
}

// Helper method to convert DB Application to Domain Application
func (r *PostgresRepository) dbToDomainApp(dbApp *db.Application) (*application.Application, error) {
	// Parse config JSON
	var config application.ApplicationConfig
	if dbApp.Config != nil {
		if err := json.Unmarshal([]byte(dbApp.Config), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Parse endpoints JSON
	var endpoints []application.Endpoint
	if dbApp.Endpoints != nil {
		if err := json.Unmarshal([]byte(dbApp.Endpoints), &endpoints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal endpoints: %w", err)
		}
	}

	app := &application.Application{
		ID:              dbApp.ID,
		WorkspaceID:     dbApp.WorkspaceID,
		ProjectID:       dbApp.ProjectID,
		Name:            dbApp.Name,
		Type:            application.ApplicationType(dbApp.Type),
		Status:          application.ApplicationStatus(dbApp.Status),
		Config:          config,
		Endpoints:       endpoints,
		CronSchedule:    ptrToString(dbApp.CronSchedule),
		CronCommand:     dbApp.CronCommand,
		CronArgs:        dbApp.CronArgs,
		TemplateAppID:   ptrToString(dbApp.TemplateAppID),
		LastExecutionAt: dbApp.LastExecutionAt,
		NextExecutionAt: dbApp.NextExecutionAt,
		CreatedAt:       dbApp.CreatedAt,
		UpdatedAt:       dbApp.UpdatedAt,
	}

	// Set source
	app.Source.Type = application.SourceType(dbApp.SourceType)
	app.Source.Image = dbApp.SourceImage
	app.Source.GitURL = dbApp.SourceGitURL
	app.Source.GitRef = dbApp.SourceGitRef

	return app, nil
}

// Helper method to convert Domain Application to DB Application
func (r *PostgresRepository) domainToDBApp(app *application.Application) (*db.Application, error) {
	// Marshal config to JSON
	configJSON, err := json.Marshal(app.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Marshal endpoints to JSON
	endpointsJSON, err := json.Marshal(app.Endpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal endpoints: %w", err)
	}

	dbApp := &db.Application{
		ID:              app.ID,
		WorkspaceID:     app.WorkspaceID,
		ProjectID:       app.ProjectID,
		Name:            app.Name,
		Type:            string(app.Type),
		Status:          string(app.Status),
		SourceType:      string(app.Source.Type),
		SourceImage:     app.Source.Image,
		SourceGitURL:    app.Source.GitURL,
		SourceGitRef:    app.Source.GitRef,
		Config:          db.JSON(configJSON),
		Endpoints:       db.JSON(endpointsJSON),
		CronSchedule:    stringToPtr(app.CronSchedule),
		CronCommand:     pq.StringArray(app.CronCommand),
		CronArgs:        pq.StringArray(app.CronArgs),
		TemplateAppID:   stringToPtr(app.TemplateAppID),
		LastExecutionAt: app.LastExecutionAt,
		NextExecutionAt: app.NextExecutionAt,
		CreatedAt:       app.CreatedAt,
		UpdatedAt:       app.UpdatedAt,
	}

	return dbApp, nil
}

// Helper functions
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}