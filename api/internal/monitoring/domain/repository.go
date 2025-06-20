package domain

import (
	"context"
	"time"
)

// Repository defines the data access interface for monitoring
type Repository interface {
	// Metrics operations
	SaveMetrics(ctx context.Context, metrics []*MetricDataPoint) error
	GetMetrics(ctx context.Context, workspaceID string, metricName string, start, end time.Time) ([]*MetricDataPoint, error)
	GetLatestMetrics(ctx context.Context, workspaceID string, metricNames []string) (map[string]*MetricDataPoint, error)
	DeleteOldMetrics(ctx context.Context, before time.Time) error
	
	// Alerts operations
	CreateAlert(ctx context.Context, alert *Alert) error
	GetAlert(ctx context.Context, alertID string) (*Alert, error)
	GetAlerts(ctx context.Context, workspaceID string, filter AlertFilter) ([]*Alert, error)
	UpdateAlert(ctx context.Context, alert *Alert) error
	DeleteAlert(ctx context.Context, alertID string) error
	
	// Health check operations
	SaveHealthCheck(ctx context.Context, health *ClusterHealth) error
	GetLatestHealthCheck(ctx context.Context, workspaceID string) (*ClusterHealth, error)
}

// AlertFilter defines filtering options for alerts
type AlertFilter struct {
	Severity  string
	Status    string
	Type      string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
}