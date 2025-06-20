package domain

import (
	"context"
)

// Service defines the monitoring business logic interface
type Service interface {
	// GetWorkspaceMetrics retrieves aggregated metrics for a workspace
	GetWorkspaceMetrics(ctx context.Context, workspaceID string, opts QueryOptions) (*WorkspaceMetrics, error)
	
	// GetClusterHealth checks and returns the health status of a vCluster
	GetClusterHealth(ctx context.Context, workspaceID string) (*ClusterHealth, error)
	
	// GetResourceUsage returns current resource usage for a workspace
	GetResourceUsage(ctx context.Context, workspaceID string) (*ResourceUsage, error)
	
	// GetAlerts retrieves alerts for a workspace
	GetAlerts(ctx context.Context, workspaceID string, severity string) ([]*Alert, error)
	
	// CreateAlert creates a new monitoring alert
	CreateAlert(ctx context.Context, alert *Alert) error
	
	// AcknowledgeAlert marks an alert as acknowledged
	AcknowledgeAlert(ctx context.Context, alertID string, userID string) error
	
	// ResolveAlert marks an alert as resolved
	ResolveAlert(ctx context.Context, alertID string) error
	
	// CollectMetrics collects and stores current metrics for a workspace
	CollectMetrics(ctx context.Context, workspaceID string) error
}