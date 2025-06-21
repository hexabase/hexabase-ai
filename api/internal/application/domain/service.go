package domain

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

	// CronJob operations
	CreateCronJob(ctx context.Context, app *Application) error
	UpdateCronJobSchedule(ctx context.Context, applicationID, newSchedule string) error
	TriggerCronJob(ctx context.Context, req *TriggerCronJobRequest) (*CronJobExecution, error)
	GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]CronJobExecution, int, error)
	GetCronJobStatus(ctx context.Context, applicationID string) (*CronJobStatus, error)
	UpdateCronJobExecutionStatus(ctx context.Context, executionID string, status CronJobExecutionStatus) error

	// Function operations
	CreateFunction(ctx context.Context, workspaceID string, req CreateFunctionRequest) (*Application, error)
	DeployFunctionVersion(ctx context.Context, applicationID string, sourceCode string) (*FunctionVersion, error)
	GetFunctionVersions(ctx context.Context, applicationID string) ([]FunctionVersion, error)
	SetActiveFunctionVersion(ctx context.Context, applicationID, versionID string) error
	InvokeFunction(ctx context.Context, applicationID string, req InvokeFunctionRequest) (*InvokeFunctionResponse, error)
	GetFunctionInvocations(ctx context.Context, applicationID string, limit, offset int) ([]FunctionInvocation, int, error)
	GetFunctionEvents(ctx context.Context, applicationID string, limit int) ([]FunctionEvent, error)
	ProcessFunctionEvent(ctx context.Context, eventID string) error
}