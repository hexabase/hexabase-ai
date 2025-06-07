package application

import (
	"context"
	"io"
)

// Service defines the interface for application business logic
type Service interface {
	// Application lifecycle management
	CreateApplication(ctx context.Context, workspaceID string, req CreateApplicationRequest) (*Application, error)
	GetApplication(ctx context.Context, applicationID string) (*Application, error)
	ListApplications(ctx context.Context, workspaceID, projectID string) ([]Application, error)
	UpdateApplication(ctx context.Context, applicationID string, req UpdateApplicationRequest) (*Application, error)
	DeleteApplication(ctx context.Context, applicationID string) error

	// Application operations
	StartApplication(ctx context.Context, applicationID string) error
	StopApplication(ctx context.Context, applicationID string) error
	RestartApplication(ctx context.Context, applicationID string) error
	ScaleApplication(ctx context.Context, applicationID string, replicas int) error

	// Pod operations
	ListPods(ctx context.Context, applicationID string) ([]Pod, error)
	RestartPod(ctx context.Context, applicationID, podName string) error
	GetPodLogs(ctx context.Context, query LogQuery) ([]LogEntry, error)
	StreamPodLogs(ctx context.Context, query LogQuery) (io.ReadCloser, error)

	// Monitoring and metrics
	GetApplicationMetrics(ctx context.Context, applicationID string) (*ApplicationMetrics, error)
	GetApplicationEvents(ctx context.Context, applicationID string, limit int) ([]ApplicationEvent, error)

	// Network operations
	UpdateNetworkConfig(ctx context.Context, applicationID string, config NetworkConfig) error
	GetApplicationEndpoints(ctx context.Context, applicationID string) ([]Endpoint, error)

	// Node affinity operations (for dedicated nodes)
	UpdateNodeAffinity(ctx context.Context, applicationID string, nodeSelector map[string]string) error
	MigrateToNode(ctx context.Context, applicationID, targetNodeID string) error
}