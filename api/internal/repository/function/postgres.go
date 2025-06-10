package function

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	
	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
)

// PostgresRepository implements the function.Repository interface using PostgreSQL
type PostgresRepository struct {
	db     *sql.DB
	config *ConfigRepository
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		config: NewConfigRepository(db),
	}
}

// CreateFunction creates a new function record
func (r *PostgresRepository) CreateFunction(ctx context.Context, fn *function.FunctionDef) error {
	query := `
		INSERT INTO functions (
			id, workspace_id, project_id, name, namespace, runtime, 
			handler, description, status, active_version, 
			labels, annotations, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	labelsJSON, _ := json.Marshal(fn.Labels)
	annotationsJSON, _ := json.Marshal(fn.Annotations)

	_, err := r.db.ExecContext(ctx, query,
		fn.ID, fn.WorkspaceID, fn.ProjectID, fn.Name, fn.Namespace,
		string(fn.Runtime), fn.Handler, "", string(fn.Status),
		fn.ActiveVersion, labelsJSON, annotationsJSON, fn.CreatedAt, fn.UpdatedAt,
	)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return fmt.Errorf("function already exists")
		}
		return fmt.Errorf("failed to create function: %w", err)
	}

	return nil
}

// UpdateFunction updates an existing function
func (r *PostgresRepository) UpdateFunction(ctx context.Context, fn *function.FunctionDef) error {
	query := `
		UPDATE functions SET
			name = $3, namespace = $4, runtime = $5, handler = $6,
			description = $7, status = $8, active_version = $9,
			labels = $10, annotations = $11, updated_at = $12
		WHERE workspace_id = $1 AND id = $2
	`

	labelsJSON, _ := json.Marshal(fn.Labels)
	annotationsJSON, _ := json.Marshal(fn.Annotations)

	result, err := r.db.ExecContext(ctx, query,
		fn.WorkspaceID, fn.ID, fn.Name, fn.Namespace,
		string(fn.Runtime), fn.Handler, "", string(fn.Status),
		fn.ActiveVersion, labelsJSON, annotationsJSON, fn.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("function not found")
	}

	return nil
}

// DeleteFunction deletes a function
func (r *PostgresRepository) DeleteFunction(ctx context.Context, workspaceID, functionID string) error {
	query := `DELETE FROM functions WHERE workspace_id = $1 AND id = $2`

	result, err := r.db.ExecContext(ctx, query, workspaceID, functionID)
	if err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("function not found")
	}

	return nil
}

// GetFunction retrieves a function by ID
func (r *PostgresRepository) GetFunction(ctx context.Context, workspaceID, functionID string) (*function.FunctionDef, error) {
	query := `
		SELECT id, workspace_id, project_id, name, namespace, runtime,
		       handler, description, status, active_version,
		       labels, annotations, created_at, updated_at
		FROM functions
		WHERE workspace_id = $1 AND id = $2
	`

	var fn function.FunctionDef
	var runtime, status string
	var labelsJSON, annotationsJSON []byte

	var description string
	err := r.db.QueryRowContext(ctx, query, workspaceID, functionID).Scan(
		&fn.ID, &fn.WorkspaceID, &fn.ProjectID, &fn.Name, &fn.Namespace,
		&runtime, &fn.Handler, &description, &status, &fn.ActiveVersion,
		&labelsJSON, &annotationsJSON, &fn.CreatedAt, &fn.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("function not found")
		}
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	fn.Runtime = function.Runtime(runtime)
	fn.Status = function.FunctionDefStatus(status)
	json.Unmarshal(labelsJSON, &fn.Labels)
	json.Unmarshal(annotationsJSON, &fn.Annotations)

	return &fn, nil
}

// ListFunctions lists all functions in a project
func (r *PostgresRepository) ListFunctions(ctx context.Context, workspaceID, projectID string) ([]*function.FunctionDef, error) {
	query := `
		SELECT id, workspace_id, project_id, name, namespace, runtime,
		       handler, description, status, active_version,
		       labels, annotations, created_at, updated_at
		FROM functions
		WHERE workspace_id = $1 AND project_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}
	defer rows.Close()

	var functions []*function.FunctionDef
	for rows.Next() {
		var fn function.FunctionDef
		var runtime, status string
		var labelsJSON, annotationsJSON []byte

		var description string
		err := rows.Scan(
			&fn.ID, &fn.WorkspaceID, &fn.ProjectID, &fn.Name, &fn.Namespace,
			&runtime, &fn.Handler, &description, &status, &fn.ActiveVersion,
			&labelsJSON, &annotationsJSON, &fn.CreatedAt, &fn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan function: %w", err)
		}

		fn.Runtime = function.Runtime(runtime)
		fn.Status = function.FunctionDefStatus(status)
		json.Unmarshal(labelsJSON, &fn.Labels)
		json.Unmarshal(annotationsJSON, &fn.Annotations)

		functions = append(functions, &fn)
	}

	return functions, nil
}

// CreateVersion creates a new version record
func (r *PostgresRepository) CreateVersion(ctx context.Context, version *function.FunctionVersionDef) error {
	query := `
		INSERT INTO function_versions (
			id, workspace_id, function_id, function_name, version,
			runtime, handler, image, source_code, build_status,
			build_log, created_at, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		version.ID, version.WorkspaceID, version.FunctionID, version.FunctionName,
		version.Version, "", "", version.Image,
		version.SourceCode, string(version.BuildStatus), version.BuildLogs,
		version.CreatedAt, version.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	return nil
}

// UpdateVersion updates a version record
func (r *PostgresRepository) UpdateVersion(ctx context.Context, version *function.FunctionVersionDef) error {
	query := `
		UPDATE function_versions SET
			build_status = $4, build_log = $5, is_active = $6
		WHERE workspace_id = $1 AND function_id = $2 AND id = $3
	`

	result, err := r.db.ExecContext(ctx, query,
		version.WorkspaceID, version.FunctionID, version.ID,
		string(version.BuildStatus), version.BuildLogs, version.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update version: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("version not found")
	}

	return nil
}

// GetVersion retrieves a specific version
func (r *PostgresRepository) GetVersion(ctx context.Context, workspaceID, functionID, versionID string) (*function.FunctionVersionDef, error) {
	query := `
		SELECT id, workspace_id, function_id, function_name, version,
		       runtime, handler, image, source_code, build_status,
		       build_log, created_at, is_active
		FROM function_versions
		WHERE workspace_id = $1 AND function_id = $2 AND id = $3
	`

	var v function.FunctionVersionDef
	var runtime, buildStatus string

	var handler string
	err := r.db.QueryRowContext(ctx, query, workspaceID, functionID, versionID).Scan(
		&v.ID, &v.WorkspaceID, &v.FunctionID, &v.FunctionName, &v.Version,
		&runtime, &handler, &v.Image, &v.SourceCode, &buildStatus,
		&v.BuildLogs, &v.CreatedAt, &v.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("version not found")
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	v.BuildStatus = function.FunctionBuildStatus(buildStatus)

	return &v, nil
}

// ListVersions lists all versions of a function
func (r *PostgresRepository) ListVersions(ctx context.Context, workspaceID, functionID string) ([]*function.FunctionVersionDef, error) {
	query := `
		SELECT id, workspace_id, function_id, function_name, version,
		       runtime, handler, image, source_code, build_status,
		       build_log, created_at, is_active
		FROM function_versions
		WHERE workspace_id = $1 AND function_id = $2
		ORDER BY version DESC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}
	defer rows.Close()

	var versions []*function.FunctionVersionDef
	for rows.Next() {
		var v function.FunctionVersionDef
		var runtime, buildStatus string

		var handler string
		err := rows.Scan(
			&v.ID, &v.WorkspaceID, &v.FunctionID, &v.FunctionName, &v.Version,
			&runtime, &handler, &v.Image, &v.SourceCode, &buildStatus,
			&v.BuildLogs, &v.CreatedAt, &v.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}

		v.BuildStatus = function.FunctionBuildStatus(buildStatus)

		versions = append(versions, &v)
	}

	return versions, nil
}

// CreateTrigger creates a new trigger record
func (r *PostgresRepository) CreateTrigger(ctx context.Context, trigger *function.FunctionTrigger) error {
	query := `
		INSERT INTO function_triggers (
			id, workspace_id, function_id, name, type,
			function_name, enabled, config, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	configJSON, _ := json.Marshal(trigger.Config)

	_, err := r.db.ExecContext(ctx, query,
		trigger.ID, trigger.WorkspaceID, trigger.FunctionID, trigger.Name,
		string(trigger.Type), trigger.FunctionName, trigger.Enabled,
		configJSON, trigger.CreatedAt, trigger.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create trigger: %w", err)
	}

	return nil
}

// UpdateTrigger updates a trigger record
func (r *PostgresRepository) UpdateTrigger(ctx context.Context, trigger *function.FunctionTrigger) error {
	query := `
		UPDATE function_triggers SET
			name = $4, type = $5, enabled = $6, config = $7, updated_at = $8
		WHERE workspace_id = $1 AND function_id = $2 AND id = $3
	`

	configJSON, _ := json.Marshal(trigger.Config)

	result, err := r.db.ExecContext(ctx, query,
		trigger.WorkspaceID, trigger.FunctionID, trigger.ID,
		trigger.Name, string(trigger.Type), trigger.Enabled,
		configJSON, trigger.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update trigger: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("trigger not found")
	}

	return nil
}

// DeleteTrigger deletes a trigger
func (r *PostgresRepository) DeleteTrigger(ctx context.Context, workspaceID, functionID, triggerID string) error {
	query := `DELETE FROM function_triggers WHERE workspace_id = $1 AND function_id = $2 AND id = $3`

	result, err := r.db.ExecContext(ctx, query, workspaceID, functionID, triggerID)
	if err != nil {
		return fmt.Errorf("failed to delete trigger: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("trigger not found")
	}

	return nil
}

// ListTriggers lists all triggers for a function
func (r *PostgresRepository) ListTriggers(ctx context.Context, workspaceID, functionID string) ([]*function.FunctionTrigger, error) {
	query := `
		SELECT id, workspace_id, function_id, name, type,
		       function_name, enabled, config, created_at, updated_at
		FROM function_triggers
		WHERE workspace_id = $1 AND function_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list triggers: %w", err)
	}
	defer rows.Close()

	var triggers []*function.FunctionTrigger
	for rows.Next() {
		var t function.FunctionTrigger
		var triggerType string
		var configJSON []byte

		err := rows.Scan(
			&t.ID, &t.WorkspaceID, &t.FunctionID, &t.Name, &triggerType,
			&t.FunctionName, &t.Enabled, &configJSON, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trigger: %w", err)
		}

		t.Type = function.TriggerType(triggerType)
		json.Unmarshal(configJSON, &t.Config)

		triggers = append(triggers, &t)
	}

	return triggers, nil
}

// CreateInvocation creates a new invocation record
func (r *PostgresRepository) CreateInvocation(ctx context.Context, invocation *function.InvocationStatus) error {
	query := `
		INSERT INTO function_invocations (
			invocation_id, workspace_id, function_id, status,
			started_at, completed_at, result, error
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	resultJSON, _ := json.Marshal(invocation.Result)

	_, err := r.db.ExecContext(ctx, query,
		invocation.InvocationID, invocation.WorkspaceID, invocation.FunctionID,
		invocation.Status, invocation.StartedAt, invocation.CompletedAt,
		resultJSON, invocation.Error,
	)

	if err != nil {
		return fmt.Errorf("failed to create invocation: %w", err)
	}

	return nil
}

// UpdateInvocation updates an invocation record
func (r *PostgresRepository) UpdateInvocation(ctx context.Context, invocation *function.InvocationStatus) error {
	query := `
		UPDATE function_invocations SET
			status = $3, completed_at = $4, result = $5, error = $6
		WHERE workspace_id = $1 AND invocation_id = $2
	`

	resultJSON, _ := json.Marshal(invocation.Result)

	result, err := r.db.ExecContext(ctx, query,
		invocation.WorkspaceID, invocation.InvocationID,
		invocation.Status, invocation.CompletedAt, resultJSON, invocation.Error,
	)

	if err != nil {
		return fmt.Errorf("failed to update invocation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("invocation not found")
	}

	return nil
}

// GetInvocation retrieves an invocation by ID
func (r *PostgresRepository) GetInvocation(ctx context.Context, workspaceID, invocationID string) (*function.InvocationStatus, error) {
	query := `
		SELECT invocation_id, workspace_id, function_id, status,
		       started_at, completed_at, result, error
		FROM function_invocations
		WHERE workspace_id = $1 AND invocation_id = $2
	`

	var inv function.InvocationStatus
	var resultJSON []byte
	var errorStr sql.NullString

	err := r.db.QueryRowContext(ctx, query, workspaceID, invocationID).Scan(
		&inv.InvocationID, &inv.WorkspaceID, &inv.FunctionID, &inv.Status,
		&inv.StartedAt, &inv.CompletedAt, &resultJSON, &errorStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invocation not found")
		}
		return nil, fmt.Errorf("failed to get invocation: %w", err)
	}

	if errorStr.Valid {
		inv.Error = errorStr.String
	}
	json.Unmarshal(resultJSON, &inv.Result)

	return &inv, nil
}

// ListInvocations lists invocations for a function
func (r *PostgresRepository) ListInvocations(ctx context.Context, workspaceID, functionID string, limit int) ([]*function.InvocationStatus, error) {
	query := `
		SELECT invocation_id, workspace_id, function_id, status,
		       started_at, completed_at, result, error
		FROM function_invocations
		WHERE workspace_id = $1 AND function_id = $2
		ORDER BY started_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID, functionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list invocations: %w", err)
	}
	defer rows.Close()

	var invocations []*function.InvocationStatus
	for rows.Next() {
		var inv function.InvocationStatus
		var resultJSON []byte
		var errorStr sql.NullString

		err := rows.Scan(
			&inv.InvocationID, &inv.WorkspaceID, &inv.FunctionID, &inv.Status,
			&inv.StartedAt, &inv.CompletedAt, &resultJSON, &errorStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invocation: %w", err)
		}

		if errorStr.Valid {
			inv.Error = errorStr.String
		}
		json.Unmarshal(resultJSON, &inv.Result)

		invocations = append(invocations, &inv)
	}

	return invocations, nil
}

// CreateEvent creates a new audit event
func (r *PostgresRepository) CreateEvent(ctx context.Context, event *function.FunctionAuditEvent) error {
	query := `
		INSERT INTO function_events (
			id, workspace_id, function_id, type, description,
			user_id, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	metadataJSON, _ := json.Marshal(event.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.WorkspaceID, event.FunctionID, event.Type,
		event.Description, "", metadataJSON, event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// ListEvents lists audit events for a function
func (r *PostgresRepository) ListEvents(ctx context.Context, workspaceID, functionID string, limit int) ([]*function.FunctionAuditEvent, error) {
	query := `
		SELECT id, workspace_id, function_id, type, description,
		       user_id, metadata, created_at
		FROM function_events
		WHERE workspace_id = $1 AND function_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID, functionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []*function.FunctionAuditEvent
	for rows.Next() {
		var e function.FunctionAuditEvent
		var metadataJSON []byte

		var userID string
		err := rows.Scan(
			&e.ID, &e.WorkspaceID, &e.FunctionID, &e.Type, &e.Description,
			&userID, &metadataJSON, &e.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		json.Unmarshal(metadataJSON, &e.Metadata)

		events = append(events, &e)
	}

	return events, nil
}

// GetWorkspaceProviderConfig retrieves the provider configuration for a workspace
func (r *PostgresRepository) GetWorkspaceProviderConfig(ctx context.Context, workspaceID string) (*function.ProviderConfig, error) {
	return r.config.GetWorkspaceProviderConfig(ctx, workspaceID)
}

// UpdateWorkspaceProviderConfig updates the provider configuration for a workspace
func (r *PostgresRepository) UpdateWorkspaceProviderConfig(ctx context.Context, workspaceID string, config *function.ProviderConfig) error {
	return r.config.UpdateWorkspaceProviderConfig(ctx, workspaceID, config)
}