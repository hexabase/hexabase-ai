package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hexabase/kaas-api/internal/domain/monitoring"
	"gorm.io/gorm"
)

// postgresRepository implements monitoring.Repository using PostgreSQL
type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL monitoring repository
func NewPostgresRepository(db *gorm.DB) monitoring.Repository {
	return &postgresRepository{db: db}
}

// Database models
type metricRecord struct {
	ID          string    `gorm:"primaryKey"`
	WorkspaceID string    `gorm:"index"`
	MetricName  string    `gorm:"index"`
	Value       float64
	Labels      string // JSON string
	Timestamp   time.Time `gorm:"index"`
	CreatedAt   time.Time
}

type alertRecord struct {
	ID          string `gorm:"primaryKey"`
	WorkspaceID string `gorm:"index"`
	Type        string
	Severity    string `gorm:"index"`
	Title       string
	Description string
	Resource    string
	Threshold   float64
	Value       float64
	Status      string `gorm:"index"`
	CreatedAt   time.Time
	ResolvedAt  *time.Time
	UpdatedAt   time.Time
}

type healthCheckRecord struct {
	ID          string `gorm:"primaryKey"`
	WorkspaceID string `gorm:"uniqueIndex"`
	Healthy     bool
	Components  string // JSON string
	LastChecked time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SaveMetrics saves multiple metric data points
func (r *postgresRepository) SaveMetrics(ctx context.Context, metrics []*monitoring.MetricDataPoint) error {
	if len(metrics) == 0 {
		return nil
	}

	records := make([]metricRecord, len(metrics))
	for i, m := range metrics {
		labelsJSON, _ := json.Marshal(m.Labels)
		records[i] = metricRecord{
			ID:          m.ID,
			WorkspaceID: m.WorkspaceID,
			MetricName:  m.MetricName,
			Value:       m.Value,
			Labels:      string(labelsJSON),
			Timestamp:   m.Timestamp,
			CreatedAt:   time.Now(),
		}
	}

	return r.db.WithContext(ctx).Create(&records).Error
}

// GetMetrics retrieves metrics for a workspace within a time range
func (r *postgresRepository) GetMetrics(ctx context.Context, workspaceID string, metricName string, start, end time.Time) ([]*monitoring.MetricDataPoint, error) {
	var records []metricRecord

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND metric_name = ? AND timestamp BETWEEN ? AND ?", 
			workspaceID, metricName, start, end).
		Order("timestamp ASC").
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	metrics := make([]*monitoring.MetricDataPoint, len(records))
	for i, rec := range records {
		var labels map[string]string
		if rec.Labels != "" {
			json.Unmarshal([]byte(rec.Labels), &labels)
		}

		metrics[i] = &monitoring.MetricDataPoint{
			ID:          rec.ID,
			WorkspaceID: rec.WorkspaceID,
			MetricName:  rec.MetricName,
			Value:       rec.Value,
			Labels:      labels,
			Timestamp:   rec.Timestamp,
		}
	}

	return metrics, nil
}

// GetLatestMetrics retrieves the most recent metrics for given metric names
func (r *postgresRepository) GetLatestMetrics(ctx context.Context, workspaceID string, metricNames []string) (map[string]*monitoring.MetricDataPoint, error) {
	result := make(map[string]*monitoring.MetricDataPoint)

	for _, metricName := range metricNames {
		var record metricRecord
		err := r.db.WithContext(ctx).
			Where("workspace_id = ? AND metric_name = ?", workspaceID, metricName).
			Order("timestamp DESC").
			First(&record).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to get latest metric %s: %w", metricName, err)
		}

		if err != gorm.ErrRecordNotFound {
			var labels map[string]string
			if record.Labels != "" {
				json.Unmarshal([]byte(record.Labels), &labels)
			}

			result[metricName] = &monitoring.MetricDataPoint{
				ID:          record.ID,
				WorkspaceID: record.WorkspaceID,
				MetricName:  record.MetricName,
				Value:       record.Value,
				Labels:      labels,
				Timestamp:   record.Timestamp,
			}
		}
	}

	return result, nil
}

// DeleteOldMetrics removes metrics older than the specified time
func (r *postgresRepository) DeleteOldMetrics(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("timestamp < ?", before).
		Delete(&metricRecord{}).Error
}

// CreateAlert creates a new alert
func (r *postgresRepository) CreateAlert(ctx context.Context, alert *monitoring.Alert) error {
	record := alertRecord{
		ID:          alert.ID,
		WorkspaceID: alert.WorkspaceID,
		Type:        alert.Type,
		Severity:    alert.Severity,
		Title:       alert.Title,
		Description: alert.Description,
		Resource:    alert.Resource,
		Threshold:   alert.Threshold,
		Value:       alert.Value,
		Status:      alert.Status,
		CreatedAt:   alert.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	return r.db.WithContext(ctx).Create(&record).Error
}

// GetAlert retrieves a single alert by ID
func (r *postgresRepository) GetAlert(ctx context.Context, alertID string) (*monitoring.Alert, error) {
	var record alertRecord
	err := r.db.WithContext(ctx).Where("id = ?", alertID).First(&record).Error
	if err != nil {
		return nil, err
	}

	return r.recordToAlert(&record), nil
}

// GetAlerts retrieves alerts based on filter criteria
func (r *postgresRepository) GetAlerts(ctx context.Context, workspaceID string, filter monitoring.AlertFilter) ([]*monitoring.Alert, error) {
	query := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)

	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	var records []alertRecord
	err := query.Order("created_at DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	alerts := make([]*monitoring.Alert, len(records))
	for i, rec := range records {
		alerts[i] = r.recordToAlert(&rec)
	}

	return alerts, nil
}

// UpdateAlert updates an existing alert
func (r *postgresRepository) UpdateAlert(ctx context.Context, alert *monitoring.Alert) error {
	record := alertRecord{
		ID:          alert.ID,
		WorkspaceID: alert.WorkspaceID,
		Type:        alert.Type,
		Severity:    alert.Severity,
		Title:       alert.Title,
		Description: alert.Description,
		Resource:    alert.Resource,
		Threshold:   alert.Threshold,
		Value:       alert.Value,
		Status:      alert.Status,
		CreatedAt:   alert.CreatedAt,
		ResolvedAt:  alert.ResolvedAt,
		UpdatedAt:   time.Now(),
	}

	return r.db.WithContext(ctx).Save(&record).Error
}

// DeleteAlert removes an alert
func (r *postgresRepository) DeleteAlert(ctx context.Context, alertID string) error {
	return r.db.WithContext(ctx).Where("id = ?", alertID).Delete(&alertRecord{}).Error
}

// SaveHealthCheck saves or updates a health check result
func (r *postgresRepository) SaveHealthCheck(ctx context.Context, health *monitoring.ClusterHealth) error {
	componentsJSON, err := json.Marshal(health.Components)
	if err != nil {
		return fmt.Errorf("failed to marshal components: %w", err)
	}

	record := healthCheckRecord{
		ID:          fmt.Sprintf("%s-%d", health.WorkspaceID, time.Now().Unix()),
		WorkspaceID: health.WorkspaceID,
		Healthy:     health.Healthy,
		Components:  string(componentsJSON),
		LastChecked: health.LastChecked,
		UpdatedAt:   time.Now(),
	}

	// Upsert - update if exists, create if not
	return r.db.WithContext(ctx).
		Where("workspace_id = ?", health.WorkspaceID).
		Assign(record).
		FirstOrCreate(&record).Error
}

// GetLatestHealthCheck retrieves the most recent health check for a workspace
func (r *postgresRepository) GetLatestHealthCheck(ctx context.Context, workspaceID string) (*monitoring.ClusterHealth, error) {
	var record healthCheckRecord
	err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("last_checked DESC").
		First(&record).Error

	if err != nil {
		return nil, err
	}

	var components map[string]monitoring.ComponentHealth
	if err := json.Unmarshal([]byte(record.Components), &components); err != nil {
		return nil, fmt.Errorf("failed to unmarshal components: %w", err)
	}

	return &monitoring.ClusterHealth{
		WorkspaceID: record.WorkspaceID,
		Healthy:     record.Healthy,
		Components:  components,
		LastChecked: record.LastChecked,
	}, nil
}

// Helper method to convert record to domain model
func (r *postgresRepository) recordToAlert(rec *alertRecord) *monitoring.Alert {
	return &monitoring.Alert{
		ID:          rec.ID,
		WorkspaceID: rec.WorkspaceID,
		Type:        rec.Type,
		Severity:    rec.Severity,
		Title:       rec.Title,
		Description: rec.Description,
		Resource:    rec.Resource,
		Threshold:   rec.Threshold,
		Value:       rec.Value,
		Status:      rec.Status,
		CreatedAt:   rec.CreatedAt,
		ResolvedAt:  rec.ResolvedAt,
	}
}

// Ensure tables exist
func (r *postgresRepository) AutoMigrate() error {
	return r.db.AutoMigrate(
		&metricRecord{},
		&alertRecord{},
		&healthCheckRecord{},
	)
}