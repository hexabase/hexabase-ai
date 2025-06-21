package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/hexabase/hexabase-ai/api/internal/shared/db"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// PostgresRepository implements the application repository interface using PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &PostgresRepository{db: db}
}

// CreateApplication creates a new application
func (r *PostgresRepository) CreateApplication(ctx context.Context, app *domain.Application) error {
	// For CronJob with template, fetch template app first
	if app.Type == domain.ApplicationTypeCronJob && app.TemplateAppID != "" {
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
func (r *PostgresRepository) GetApplication(ctx context.Context, id string) (*domain.Application, error) {
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
func (r *PostgresRepository) GetApplicationByName(ctx context.Context, workspaceID, projectID, name string) (*domain.Application, error) {
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
func (r *PostgresRepository) ListApplications(ctx context.Context, workspaceID, projectID string) ([]domain.Application, error) {
	var dbApps []db.Application
	query := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	err := query.Order("created_at DESC").Find(&dbApps).Error
	if err != nil {
		return nil, err
	}
	
	var apps []domain.Application
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
func (r *PostgresRepository) UpdateApplication(ctx context.Context, app *domain.Application) error {
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
func (r *PostgresRepository) CreateEvent(ctx context.Context, event *domain.ApplicationEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// ListEvents lists events for an application
func (r *PostgresRepository) ListEvents(ctx context.Context, applicationID string, limit int) ([]domain.ApplicationEvent, error) {
	var events []domain.ApplicationEvent
	query := r.db.WithContext(ctx).Where("application_id = ?", applicationID).Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&events).Error
	return events, err
}

// GetApplicationsByNode retrieves applications running on a specific node
func (r *PostgresRepository) GetApplicationsByNode(ctx context.Context, nodeID string) ([]domain.Application, error) {
	var dbApps []db.Application
	// Use JSON query for node selector
	err := r.db.WithContext(ctx).
		Where("config->'node_selector' @> ?", fmt.Sprintf(`{"node-id": "%s"}`, nodeID)).
		Find(&dbApps).Error
	if err != nil {
		return nil, err
	}
	
	var apps []domain.Application
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
func (r *PostgresRepository) GetApplicationsByStatus(ctx context.Context, workspaceID string, status domain.ApplicationStatus) ([]domain.Application, error) {
	var dbApps []db.Application
	query := r.db.WithContext(ctx).Where("status = ?", string(status))
	if workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}
	err := query.Find(&dbApps).Error
	if err != nil {
		return nil, err
	}
	
	var apps []domain.Application
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
func (r *PostgresRepository) Create(ctx context.Context, app *domain.Application) error {
	return r.CreateApplication(ctx, app)
}

// GetCronJobExecutions retrieves executions for a CronJob application
func (r *PostgresRepository) GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]domain.CronJobExecution, int, error) {
	var executions []domain.CronJobExecution
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
		exec := domain.CronJobExecution{
			ID:            dbExec.ID,
			ApplicationID: dbExec.ApplicationID,
			JobName:       dbExec.JobName,
			StartedAt:     dbExec.StartedAt,
			CompletedAt:   dbExec.CompletedAt,
			Status:        domain.CronJobExecutionStatus(dbExec.Status),
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
func (r *PostgresRepository) CreateCronJobExecution(ctx context.Context, execution *domain.CronJobExecution) error {
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
func (r *PostgresRepository) UpdateCronJobExecution(ctx context.Context, executionID string, completedAt *time.Time, status domain.CronJobExecutionStatus, exitCode *int, logs string) error {
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
func (r *PostgresRepository) GetCronJobExecutionByID(ctx context.Context, executionID string) (*domain.CronJobExecution, error) {
	var dbExec db.CronJobExecution
	if err := r.db.WithContext(ctx).Where("id = ?", executionID).First(&dbExec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	exec := &domain.CronJobExecution{
		ID:            dbExec.ID,
		ApplicationID: dbExec.ApplicationID,
		JobName:       dbExec.JobName,
		StartedAt:     dbExec.StartedAt,
		CompletedAt:   dbExec.CompletedAt,
		Status:        domain.CronJobExecutionStatus(dbExec.Status),
		ExitCode:      dbExec.ExitCode,
		Logs:          dbExec.Logs,
		CreatedAt:     dbExec.CreatedAt,
		UpdatedAt:     dbExec.UpdatedAt,
	}

	return exec, nil
}

// Helper method to convert DB Application to Domain Application
func (r *PostgresRepository) dbToDomainApp(dbApp *db.Application) (*domain.Application, error) {
	// Parse config JSON
	var config domain.ApplicationConfig
	if dbApp.Config != nil {
		if err := json.Unmarshal(dbApp.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Parse endpoints JSON
	var endpoints []domain.Endpoint
	if dbApp.Endpoints != nil {
		if err := json.Unmarshal(dbApp.Endpoints, &endpoints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal endpoints: %w", err)
		}
	}

	app := &domain.Application{
		ID:              dbApp.ID,
		WorkspaceID:     dbApp.WorkspaceID,
		ProjectID:       dbApp.ProjectID,
		Name:            dbApp.Name,
		Type:            domain.ApplicationType(dbApp.Type),
		Status:          domain.ApplicationStatus(dbApp.Status),
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

	// Function specific fields
	if dbApp.FunctionRuntime != nil {
		app.FunctionRuntime = domain.FunctionRuntime(*dbApp.FunctionRuntime)
	}
	if dbApp.FunctionHandler != nil {
		app.FunctionHandler = *dbApp.FunctionHandler
	}
	if dbApp.FunctionTimeout != nil {
		app.FunctionTimeout = *dbApp.FunctionTimeout
	}
	if dbApp.FunctionMemory != nil {
		app.FunctionMemory = *dbApp.FunctionMemory
	}
	if dbApp.FunctionTriggerType != nil {
		app.FunctionTriggerType = domain.FunctionTriggerType(*dbApp.FunctionTriggerType)
	}
	if dbApp.FunctionTriggerConfig != nil {
		var triggerConfig map[string]interface{}
		if err := json.Unmarshal(dbApp.FunctionTriggerConfig, &triggerConfig); err == nil {
			app.FunctionTriggerConfig = triggerConfig
		}
	}
	if dbApp.FunctionEnvVars != nil {
		var envVars map[string]string
		if err := json.Unmarshal(dbApp.FunctionEnvVars, &envVars); err == nil {
			app.FunctionEnvVars = envVars
		}
	}
	if dbApp.FunctionSecrets != nil {
		var secrets map[string]string
		if err := json.Unmarshal(dbApp.FunctionSecrets, &secrets); err == nil {
			app.FunctionSecrets = secrets
		}
	}

	// Set source
	app.Source.Type = domain.SourceType(dbApp.SourceType)
	app.Source.Image = dbApp.SourceImage
	app.Source.GitURL = dbApp.SourceGitURL
	app.Source.GitRef = dbApp.SourceGitRef

	return app, nil
}

// Helper method to convert Domain Application to DB Application
func (r *PostgresRepository) domainToDBApp(app *domain.Application) (*db.Application, error) {
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

	// Function specific fields
	if app.Type == domain.ApplicationTypeFunction {
		runtime := string(app.FunctionRuntime)
		dbApp.FunctionRuntime = &runtime
		dbApp.FunctionHandler = &app.FunctionHandler
		dbApp.FunctionTimeout = &app.FunctionTimeout
		dbApp.FunctionMemory = &app.FunctionMemory
		triggerType := string(app.FunctionTriggerType)
		dbApp.FunctionTriggerType = &triggerType

		if app.FunctionTriggerConfig != nil {
			triggerConfigJSON, _ := json.Marshal(app.FunctionTriggerConfig)
			dbApp.FunctionTriggerConfig = db.JSON(triggerConfigJSON)
		}
		if app.FunctionEnvVars != nil {
			envVarsJSON, _ := json.Marshal(app.FunctionEnvVars)
			dbApp.FunctionEnvVars = db.JSON(envVarsJSON)
		}
		if app.FunctionSecrets != nil {
			secretsJSON, _ := json.Marshal(app.FunctionSecrets)
			dbApp.FunctionSecrets = db.JSON(secretsJSON)
		}
	}

	return dbApp, nil
}

// CreateFunctionVersion creates a new function version
func (r *PostgresRepository) CreateFunctionVersion(ctx context.Context, version *domain.FunctionVersion) error {
	dbVersion := &db.FunctionVersion{
		ID:            version.ID,
		ApplicationID: version.ApplicationID,
		VersionNumber: version.VersionNumber,
		SourceCode:    version.SourceCode,
		SourceType:    string(version.SourceType),
		SourceURL:     version.SourceURL,
		BuildLogs:     version.BuildLogs,
		BuildStatus:   string(version.BuildStatus),
		ImageURI:      version.ImageURI,
		IsActive:      version.IsActive,
		DeployedAt:    version.DeployedAt,
		CreatedAt:     version.CreatedAt,
		UpdatedAt:     version.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Create(dbVersion).Error; err != nil {
		return err
	}

	version.ID = dbVersion.ID
	version.CreatedAt = dbVersion.CreatedAt
	version.UpdatedAt = dbVersion.UpdatedAt
	return nil
}

// GetFunctionVersion retrieves a function version by ID
func (r *PostgresRepository) GetFunctionVersion(ctx context.Context, versionID string) (*domain.FunctionVersion, error) {
	var dbVersion db.FunctionVersion
	if err := r.db.WithContext(ctx).Where("id = ?", versionID).First(&dbVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomainFunctionVersion(&dbVersion), nil
}

// GetFunctionVersions retrieves all versions for an application
func (r *PostgresRepository) GetFunctionVersions(ctx context.Context, applicationID string) ([]domain.FunctionVersion, error) {
	var dbVersions []db.FunctionVersion
	if err := r.db.WithContext(ctx).
		Where("application_id = ?", applicationID).
		Order("version_number DESC").
		Find(&dbVersions).Error; err != nil {
		return nil, err
	}

	var versions []domain.FunctionVersion
	for _, dbVersion := range dbVersions {
		versions = append(versions, *r.dbToDomainFunctionVersion(&dbVersion))
	}
	return versions, nil
}

// GetActiveFunctionVersion retrieves the active function version for an application
func (r *PostgresRepository) GetActiveFunctionVersion(ctx context.Context, applicationID string) (*domain.FunctionVersion, error) {
	var dbVersion db.FunctionVersion
	if err := r.db.WithContext(ctx).
		Where("application_id = ? AND is_active = ?", applicationID, true).
		First(&dbVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomainFunctionVersion(&dbVersion), nil
}

// UpdateFunctionVersion updates a function version
func (r *PostgresRepository) UpdateFunctionVersion(ctx context.Context, version *domain.FunctionVersion) error {
	dbVersion := &db.FunctionVersion{
		ID:          version.ID,
		BuildLogs:   version.BuildLogs,
		BuildStatus: string(version.BuildStatus),
		ImageURI:    version.ImageURI,
		IsActive:    version.IsActive,
		DeployedAt:  version.DeployedAt,
		UpdatedAt:   time.Now(),
	}

	return r.db.WithContext(ctx).Model(&db.FunctionVersion{}).
		Where("id = ?", version.ID).
		Updates(dbVersion).Error
}

// SetActiveFunctionVersion sets a specific version as active and deactivates others
func (r *PostgresRepository) SetActiveFunctionVersion(ctx context.Context, applicationID, versionID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deactivate all versions for this application
		if err := tx.Model(&db.FunctionVersion{}).
			Where("application_id = ?", applicationID).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// Activate the specified version
		now := time.Now()
		if err := tx.Model(&db.FunctionVersion{}).
			Where("id = ? AND application_id = ?", versionID, applicationID).
			Updates(map[string]interface{}{
				"is_active":   true,
				"deployed_at": &now,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateFunctionInvocation creates a new function invocation record
func (r *PostgresRepository) CreateFunctionInvocation(ctx context.Context, invocation *domain.FunctionInvocation) error {
	dbInvocation := r.domainToDBFunctionInvocation(invocation)

	if err := r.db.WithContext(ctx).Create(dbInvocation).Error; err != nil {
		return err
	}

	invocation.ID = dbInvocation.ID
	invocation.CreatedAt = dbInvocation.CreatedAt
	return nil
}

// GetFunctionInvocation retrieves a function invocation by ID
func (r *PostgresRepository) GetFunctionInvocation(ctx context.Context, invocationID string) (*domain.FunctionInvocation, error) {
	var dbInvocation db.FunctionInvocation
	if err := r.db.WithContext(ctx).Where("id = ?", invocationID).First(&dbInvocation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomainFunctionInvocation(&dbInvocation)
}

// GetFunctionInvocations retrieves invocations for an application
func (r *PostgresRepository) GetFunctionInvocations(ctx context.Context, applicationID string, limit, offset int) ([]domain.FunctionInvocation, int, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&db.FunctionInvocation{}).
		Where("application_id = ?", applicationID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var dbInvocations []db.FunctionInvocation
	if err := r.db.WithContext(ctx).
		Where("application_id = ?", applicationID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbInvocations).Error; err != nil {
		return nil, 0, err
	}

	var invocations []domain.FunctionInvocation
	for _, dbInvocation := range dbInvocations {
		inv, err := r.dbToDomainFunctionInvocation(&dbInvocation)
		if err != nil {
			return nil, 0, err
		}
		invocations = append(invocations, *inv)
	}

	return invocations, int(total), nil
}

// UpdateFunctionInvocation updates a function invocation
func (r *PostgresRepository) UpdateFunctionInvocation(ctx context.Context, invocation *domain.FunctionInvocation) error {
	updates := map[string]interface{}{
		"response_status":  invocation.ResponseStatus,
		"response_headers": r.marshalJSONField(invocation.ResponseHeaders),
		"response_body":    invocation.ResponseBody,
		"error_message":    invocation.ErrorMessage,
		"duration_ms":      invocation.DurationMs,
		"memory_used":      invocation.MemoryUsed,
		"completed_at":     invocation.CompletedAt,
	}

	return r.db.WithContext(ctx).
		Model(&db.FunctionInvocation{}).
		Where("id = ?", invocation.ID).
		Updates(updates).Error
}

// CreateFunctionEvent creates a new function event
func (r *PostgresRepository) CreateFunctionEvent(ctx context.Context, event *domain.FunctionEvent) error {
	dbEvent, err := r.domainToDBFunctionEvent(event)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Create(dbEvent).Error; err != nil {
		return err
	}

	event.ID = dbEvent.ID
	event.CreatedAt = dbEvent.CreatedAt
	return nil
}

// GetFunctionEvent retrieves a function event by ID
func (r *PostgresRepository) GetFunctionEvent(ctx context.Context, eventID string) (*domain.FunctionEvent, error) {
	var dbEvent db.FunctionEvent
	if err := r.db.WithContext(ctx).Where("id = ?", eventID).First(&dbEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomainFunctionEvent(&dbEvent)
}

// GetPendingFunctionEvents retrieves pending events for an application
func (r *PostgresRepository) GetPendingFunctionEvents(ctx context.Context, applicationID string, limit int) ([]domain.FunctionEvent, error) {
	var dbEvents []db.FunctionEvent
	query := r.db.WithContext(ctx).
		Where("application_id = ? AND processing_status IN ?", applicationID, []string{"pending", "retry"}).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&dbEvents).Error; err != nil {
		return nil, err
	}

	var events []domain.FunctionEvent
	for _, dbEvent := range dbEvents {
		event, err := r.dbToDomainFunctionEvent(&dbEvent)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}

	return events, nil
}

// UpdateFunctionEvent updates a function event
func (r *PostgresRepository) UpdateFunctionEvent(ctx context.Context, event *domain.FunctionEvent) error {
	updates := map[string]interface{}{
		"processing_status": event.ProcessingStatus,
		"retry_count":       event.RetryCount,
		"invocation_id":     event.InvocationID,
		"error_message":     event.ErrorMessage,
		"processed_at":      event.ProcessedAt,
	}

	return r.db.WithContext(ctx).
		Model(&db.FunctionEvent{}).
		Where("id = ?", event.ID).
		Updates(updates).Error
}

// Helper methods for Function domain/db conversion
func (r *PostgresRepository) dbToDomainFunctionVersion(dbVersion *db.FunctionVersion) *domain.FunctionVersion {
	return &domain.FunctionVersion{
		ID:            dbVersion.ID,
		ApplicationID: dbVersion.ApplicationID,
		VersionNumber: dbVersion.VersionNumber,
		SourceCode:    dbVersion.SourceCode,
		SourceType:    domain.FunctionSourceType(dbVersion.SourceType),
		SourceURL:     dbVersion.SourceURL,
		BuildLogs:     dbVersion.BuildLogs,
		BuildStatus:   domain.FunctionBuildStatus(dbVersion.BuildStatus),
		ImageURI:      dbVersion.ImageURI,
		IsActive:      dbVersion.IsActive,
		DeployedAt:    dbVersion.DeployedAt,
		CreatedAt:     dbVersion.CreatedAt,
		UpdatedAt:     dbVersion.UpdatedAt,
	}
}

func (r *PostgresRepository) domainToDBFunctionInvocation(inv *domain.FunctionInvocation) *db.FunctionInvocation {
	return &db.FunctionInvocation{
		ID:              inv.ID,
		ApplicationID:   inv.ApplicationID,
		VersionID:       inv.VersionID,
		InvocationID:    inv.InvocationID,
		TriggerSource:   inv.TriggerSource,
		RequestMethod:   inv.RequestMethod,
		RequestPath:     inv.RequestPath,
		RequestHeaders:  r.marshalJSONField(inv.RequestHeaders),
		RequestBody:     inv.RequestBody,
		ResponseStatus:  inv.ResponseStatus,
		ResponseHeaders: r.marshalJSONField(inv.ResponseHeaders),
		ResponseBody:    inv.ResponseBody,
		ErrorMessage:    inv.ErrorMessage,
		DurationMs:      inv.DurationMs,
		ColdStart:       inv.ColdStart,
		MemoryUsed:      inv.MemoryUsed,
		StartedAt:       inv.StartedAt,
		CompletedAt:     inv.CompletedAt,
		CreatedAt:       inv.CreatedAt,
	}
}

func (r *PostgresRepository) dbToDomainFunctionInvocation(dbInv *db.FunctionInvocation) (*domain.FunctionInvocation, error) {
	inv := &domain.FunctionInvocation{
		ID:             dbInv.ID,
		ApplicationID:  dbInv.ApplicationID,
		VersionID:      dbInv.VersionID,
		InvocationID:   dbInv.InvocationID,
		TriggerSource:  dbInv.TriggerSource,
		RequestMethod:  dbInv.RequestMethod,
		RequestPath:    dbInv.RequestPath,
		RequestBody:    dbInv.RequestBody,
		ResponseStatus: dbInv.ResponseStatus,
		ResponseBody:   dbInv.ResponseBody,
		ErrorMessage:   dbInv.ErrorMessage,
		DurationMs:     dbInv.DurationMs,
		ColdStart:      dbInv.ColdStart,
		MemoryUsed:     dbInv.MemoryUsed,
		StartedAt:      dbInv.StartedAt,
		CompletedAt:    dbInv.CompletedAt,
		CreatedAt:      dbInv.CreatedAt,
	}

	// Unmarshal headers
	if dbInv.RequestHeaders != nil {
		var headers map[string][]string
		if err := json.Unmarshal([]byte(dbInv.RequestHeaders), &headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request headers: %w", err)
		}
		inv.RequestHeaders = headers
	}

	if dbInv.ResponseHeaders != nil {
		var headers map[string][]string
		if err := json.Unmarshal([]byte(dbInv.ResponseHeaders), &headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response headers: %w", err)
		}
		inv.ResponseHeaders = headers
	}

	return inv, nil
}

func (r *PostgresRepository) domainToDBFunctionEvent(event *domain.FunctionEvent) (*db.FunctionEvent, error) {
	eventData, err := json.Marshal(event.EventData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	return &db.FunctionEvent{
		ID:               event.ID,
		ApplicationID:    event.ApplicationID,
		EventType:        event.EventType,
		EventSource:      event.EventSource,
		EventData:        db.JSON(eventData),
		ProcessingStatus: event.ProcessingStatus,
		RetryCount:       event.RetryCount,
		MaxRetries:       event.MaxRetries,
		InvocationID:     event.InvocationID,
		ErrorMessage:     event.ErrorMessage,
		CreatedAt:        event.CreatedAt,
		ProcessedAt:      event.ProcessedAt,
	}, nil
}

func (r *PostgresRepository) dbToDomainFunctionEvent(dbEvent *db.FunctionEvent) (*domain.FunctionEvent, error) {
	var eventData map[string]interface{}
	if dbEvent.EventData != nil {
		if err := json.Unmarshal([]byte(dbEvent.EventData), &eventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}
	}

	return &domain.FunctionEvent{
		ID:               dbEvent.ID,
		ApplicationID:    dbEvent.ApplicationID,
		EventType:        dbEvent.EventType,
		EventSource:      dbEvent.EventSource,
		EventData:        eventData,
		ProcessingStatus: dbEvent.ProcessingStatus,
		RetryCount:       dbEvent.RetryCount,
		MaxRetries:       dbEvent.MaxRetries,
		InvocationID:     dbEvent.InvocationID,
		ErrorMessage:     dbEvent.ErrorMessage,
		CreatedAt:        dbEvent.CreatedAt,
		ProcessedAt:      dbEvent.ProcessedAt,
	}, nil
}

func (r *PostgresRepository) marshalJSONField(data interface{}) db.JSON {
	if data == nil {
		return nil
	}
	jsonData, _ := json.Marshal(data)
	return db.JSON(jsonData)
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

// GetCronJobExecution retrieves a CronJob execution by ID
func (r *PostgresRepository) GetCronJobExecution(ctx context.Context, executionID string) (*domain.CronJobExecution, error) {
	return r.GetCronJobExecutionByID(ctx, executionID)
}