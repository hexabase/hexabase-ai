package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/cicd/domain"
	"gorm.io/gorm"
)

// PostgresRepository implements the CI/CD repository interface using PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &PostgresRepository{db: db}
}

// CreatePipeline creates a new pipeline
func (r *PostgresRepository) CreatePipeline(ctx context.Context, pipeline *domain.Pipeline) error {
	return r.db.WithContext(ctx).Create(pipeline).Error
}

// GetPipeline retrieves a pipeline by ID
func (r *PostgresRepository) GetPipeline(ctx context.Context, id string) (*domain.Pipeline, error) {
	var pipeline domain.Pipeline
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&pipeline).Error
	if err != nil {
		return nil, err
	}
	return &pipeline, nil
}

// GetPipelineByRunID retrieves a pipeline by run ID
func (r *PostgresRepository) GetPipelineByRunID(ctx context.Context, runID string) (*domain.Pipeline, error) {
	var run domain.PipelineRunRecord
	err := r.db.WithContext(ctx).Where("run_id = ?", runID).First(&run).Error
	if err != nil {
		return nil, err
	}

	return r.GetPipeline(ctx, run.PipelineID)
}

// ListPipelines lists pipelines for a workspace/project
func (r *PostgresRepository) ListPipelines(ctx context.Context, workspaceID, projectID string, limit, offset int) ([]*domain.Pipeline, error) {
	var pipelines []*domain.Pipeline
	
	query := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&pipelines).Error
	if err != nil {
		return nil, err
	}
	
	return pipelines, nil
}

// UpdatePipeline updates a pipeline
func (r *PostgresRepository) UpdatePipeline(ctx context.Context, pipeline *domain.Pipeline) error {
	return r.db.WithContext(ctx).Save(pipeline).Error
}

// DeletePipeline deletes a pipeline
func (r *PostgresRepository) DeletePipeline(ctx context.Context, id string) error {
	// Delete associated runs first
	if err := r.db.WithContext(ctx).Where("pipeline_id = ?", id).Delete(&domain.PipelineRunRecord{}).Error; err != nil {
		return err
	}
	
	// Delete the pipeline
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Pipeline{}).Error
}

// CreatePipelineRun creates a new pipeline run record
func (r *PostgresRepository) CreatePipelineRun(ctx context.Context, run *domain.PipelineRunRecord) error {
	return r.db.WithContext(ctx).Create(run).Error
}

// GetPipelineRun retrieves a pipeline run by ID
func (r *PostgresRepository) GetPipelineRun(ctx context.Context, runID string) (*domain.PipelineRunRecord, error) {
	var run domain.PipelineRunRecord
	// First try to find by ID (which is what the handler passes)
	err := r.db.WithContext(ctx).Where("id = ?", runID).First(&run).Error
	if err != nil {
		// If not found, try by run_id (provider-specific ID)
		err = r.db.WithContext(ctx).Where("run_id = ?", runID).First(&run).Error
		if err != nil {
			return nil, err
		}
	}
	return &run, nil
}

// UpdatePipelineRun updates a pipeline run
func (r *PostgresRepository) UpdatePipelineRun(ctx context.Context, run *domain.PipelineRunRecord) error {
	return r.db.WithContext(ctx).Save(run).Error
}

// ListPipelineRuns lists runs for a pipeline
func (r *PostgresRepository) ListPipelineRuns(ctx context.Context, pipelineID string, limit, offset int) ([]*domain.PipelineRunRecord, error) {
	var runs []*domain.PipelineRunRecord
	
	err := r.db.WithContext(ctx).
		Where("pipeline_id = ?", pipelineID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&runs).Error
		
	if err != nil {
		return nil, err
	}
	
	return runs, nil
}

// CreateTemplate creates a new pipeline template
func (r *PostgresRepository) CreateTemplate(ctx context.Context, template *domain.PipelineTemplate) error {
	// Serialize stages and parameters to JSON
	stagesJSON, err := json.Marshal(template.Stages)
	if err != nil {
		return fmt.Errorf("failed to marshal stages: %w", err)
	}
	
	paramsJSON, err := json.Marshal(template.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}
	
	// Create database record
	record := &PipelineTemplateRecord{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		Provider:    template.Provider,
		Stages:      string(stagesJSON),
		Parameters:  string(paramsJSON),
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
	
	return r.db.WithContext(ctx).Create(record).Error
}

// GetTemplate retrieves a template by ID
func (r *PostgresRepository) GetTemplate(ctx context.Context, id string) (*domain.PipelineTemplate, error) {
	var record PipelineTemplateRecord
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	
	// Deserialize stages and parameters
	var stages []domain.StageTemplate
	if err := json.Unmarshal([]byte(record.Stages), &stages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stages: %w", err)
	}
	
	var parameters []domain.ParameterDefinition
	if err := json.Unmarshal([]byte(record.Parameters), &parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}
	
	template := &domain.PipelineTemplate{
		ID:          record.ID,
		Name:        record.Name,
		Description: record.Description,
		Provider:    record.Provider,
		Stages:      stages,
		Parameters:  parameters,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
	
	return template, nil
}

// ListTemplates lists templates for a provider
func (r *PostgresRepository) ListTemplates(ctx context.Context, provider string) ([]*domain.PipelineTemplate, error) {
	var records []PipelineTemplateRecord
	
	query := r.db.WithContext(ctx)
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	
	err := query.Order("name ASC").Find(&records).Error
	if err != nil {
		return nil, err
	}
	
	templates := make([]*domain.PipelineTemplate, len(records))
	for i, record := range records {
		// Deserialize stages and parameters
		var stages []domain.StageTemplate
		if err := json.Unmarshal([]byte(record.Stages), &stages); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stages: %w", err)
		}
		
		var parameters []domain.ParameterDefinition
		if err := json.Unmarshal([]byte(record.Parameters), &parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}
		
		templates[i] = &domain.PipelineTemplate{
			ID:          record.ID,
			Name:        record.Name,
			Description: record.Description,
			Provider:    record.Provider,
			Stages:      stages,
			Parameters:  parameters,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
		}
	}
	
	return templates, nil
}

// UpdateTemplate updates a template
func (r *PostgresRepository) UpdateTemplate(ctx context.Context, template *domain.PipelineTemplate) error {
	// Serialize stages and parameters to JSON
	stagesJSON, err := json.Marshal(template.Stages)
	if err != nil {
		return fmt.Errorf("failed to marshal stages: %w", err)
	}
	
	paramsJSON, err := json.Marshal(template.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}
	
	// Update database record
	record := &PipelineTemplateRecord{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		Provider:    template.Provider,
		Stages:      string(stagesJSON),
		Parameters:  string(paramsJSON),
		UpdatedAt:   template.UpdatedAt,
	}
	
	return r.db.WithContext(ctx).Save(record).Error
}

// DeleteTemplate deletes a template
func (r *PostgresRepository) DeleteTemplate(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&PipelineTemplateRecord{}).Error
}

// GetProviderConfig retrieves provider configuration for a workspace
func (r *PostgresRepository) GetProviderConfig(ctx context.Context, workspaceID string) (*domain.WorkspaceProviderConfig, error) {
	var config domain.WorkspaceProviderConfig
	err := r.db.WithContext(ctx).Where("workspace_id = ? AND is_active = ?", workspaceID, true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// SetProviderConfig sets provider configuration for a workspace
func (r *PostgresRepository) SetProviderConfig(ctx context.Context, config *domain.WorkspaceProviderConfig) error {
	// Deactivate existing configs
	if err := r.db.WithContext(ctx).
		Model(&domain.WorkspaceProviderConfig{}).
		Where("workspace_id = ?", config.WorkspaceID).
		Update("is_active", false).Error; err != nil {
		return err
	}
	
	// Create or update the new config
	config.IsActive = true
	return r.db.WithContext(ctx).Save(config).Error
}

// PipelineTemplateRecord represents a pipeline template in the database
type PipelineTemplateRecord struct {
	ID          string    `gorm:"primaryKey"`
	Name        string    `gorm:"not null"`
	Description string
	Provider    string    `gorm:"not null;index"`
	Stages      string    `gorm:"type:text"` // JSON serialized stages
	Parameters  string    `gorm:"type:text"` // JSON serialized parameters
	CreatedAt   time.Time
	UpdatedAt   time.Time
}