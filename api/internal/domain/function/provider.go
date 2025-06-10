package function

import (
	"context"
	"time"
)

// Provider defines the interface for function-as-a-service providers
type Provider interface {
	// Lifecycle management
	CreateFunction(ctx context.Context, spec *FunctionSpec) (*FunctionDef, error)
	UpdateFunction(ctx context.Context, name string, spec *FunctionSpec) (*FunctionDef, error)
	DeleteFunction(ctx context.Context, name string) error
	GetFunction(ctx context.Context, name string) (*FunctionDef, error)
	ListFunctions(ctx context.Context, namespace string) ([]*FunctionDef, error)

	// Version management
	CreateVersion(ctx context.Context, functionName string, version *FunctionVersionDef) error
	GetVersion(ctx context.Context, functionName, versionID string) (*FunctionVersionDef, error)
	ListVersions(ctx context.Context, functionName string) ([]*FunctionVersionDef, error)
	SetActiveVersion(ctx context.Context, functionName, versionID string) error

	// Trigger management
	CreateTrigger(ctx context.Context, functionName string, trigger *FunctionTrigger) error
	UpdateTrigger(ctx context.Context, functionName, triggerName string, trigger *FunctionTrigger) error
	DeleteTrigger(ctx context.Context, functionName, triggerName string) error
	ListTriggers(ctx context.Context, functionName string) ([]*FunctionTrigger, error)

	// Invocation
	InvokeFunction(ctx context.Context, name string, req *InvokeRequest) (*InvokeResponse, error)
	InvokeFunctionAsync(ctx context.Context, name string, req *InvokeRequest) (string, error) // Returns invocation ID
	GetInvocationStatus(ctx context.Context, invocationID string) (*InvocationStatus, error)
	GetFunctionURL(ctx context.Context, name string) (string, error)

	// Logs and metrics
	GetFunctionLogs(ctx context.Context, name string, opts *LogOptions) ([]*LogEntry, error)
	GetFunctionMetrics(ctx context.Context, name string, opts *MetricOptions) (*Metrics, error)

	// Provider information
	GetCapabilities() *Capabilities
	HealthCheck(ctx context.Context) error
}

// FunctionDef represents a serverless function definition
type FunctionDef struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	WorkspaceID  string            `json:"workspace_id"`
	ProjectID    string            `json:"project_id"`
	Runtime      Runtime           `json:"runtime"`
	Handler      string            `json:"handler"`
	ActiveVersion string           `json:"active_version,omitempty"`
	Status       FunctionDefStatus `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

// FunctionDefStatus represents the status of a function
type FunctionDefStatus string

const (
	FunctionDefStatusPending   FunctionDefStatus = "pending"
	FunctionDefStatusBuilding  FunctionDefStatus = "building"
	FunctionDefStatusReady     FunctionDefStatus = "ready"
	FunctionDefStatusError     FunctionDefStatus = "error"
	FunctionDefStatusDeleting  FunctionDefStatus = "deleting"
)

// FunctionVersionDef represents a specific version of a function
type FunctionVersionDef struct {
	ID          string         `json:"id"`
	WorkspaceID string         `json:"workspace_id"`
	FunctionID  string         `json:"function_id"`
	FunctionName string        `json:"function_name"`
	Version     int            `json:"version"`
	SourceCode  string         `json:"source_code,omitempty"`
	Image       string         `json:"image,omitempty"`
	BuildStatus FunctionBuildStatus `json:"build_status"`
	BuildLogs   string         `json:"build_logs,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	IsActive    bool           `json:"is_active"`
}

// FunctionBuildStatus represents the build status of a function version
type FunctionBuildStatus string

const (
	FunctionBuildStatusPending   FunctionBuildStatus = "pending"
	FunctionBuildStatusBuilding  FunctionBuildStatus = "building"
	FunctionBuildStatusSuccess   FunctionBuildStatus = "success"
	FunctionBuildStatusFailed    FunctionBuildStatus = "failed"
)

// FunctionTrigger represents a function trigger
type FunctionTrigger struct {
	ID           string            `json:"id"`
	WorkspaceID  string            `json:"workspace_id"`
	FunctionID   string            `json:"function_id"`
	Name         string            `json:"name"`
	Type         TriggerType       `json:"type"`
	FunctionName string            `json:"function_name"`
	Enabled      bool              `json:"enabled"`
	Config       map[string]string `json:"config"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// InvokeRequest represents a function invocation request
type InvokeRequest struct {
	Method      string              `json:"method,omitempty"`
	Path        string              `json:"path,omitempty"`
	Headers     map[string][]string `json:"headers,omitempty"`
	Body        []byte              `json:"body,omitempty"`
	QueryParams map[string][]string `json:"query_params,omitempty"`
}

// InvokeResponse represents a function invocation response
type InvokeResponse struct {
	StatusCode  int                 `json:"status_code"`
	Headers     map[string][]string `json:"headers"`
	Body        []byte              `json:"body"`
	Duration    time.Duration       `json:"duration"`
	ColdStart   bool                `json:"cold_start"`
	InvocationID string             `json:"invocation_id"`
}

// InvocationStatus represents the status of an async invocation
type InvocationStatus struct {
	InvocationID string        `json:"invocation_id"`
	WorkspaceID  string        `json:"workspace_id"`
	FunctionID   string        `json:"function_id"`
	Status       string        `json:"status"`
	StartedAt    time.Time     `json:"started_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	Result       *InvokeResponse `json:"result,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// LogOptions represents options for retrieving function logs
type LogOptions struct {
	Since      *time.Time `json:"since,omitempty"`
	Until      *time.Time `json:"until,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Follow     bool       `json:"follow,omitempty"`
	Container  string     `json:"container,omitempty"`
	Previous   bool       `json:"previous,omitempty"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Container string    `json:"container,omitempty"`
	Pod       string    `json:"pod,omitempty"`
}

// MetricOptions represents options for retrieving function metrics
type MetricOptions struct {
	StartTime  time.Time  `json:"start_time"`
	EndTime    time.Time  `json:"end_time"`
	Resolution string     `json:"resolution,omitempty"` // e.g., "1m", "5m", "1h"
	Metrics    []string   `json:"metrics,omitempty"`    // specific metrics to retrieve
}

// Metrics represents function metrics
type Metrics struct {
	Invocations   int64                  `json:"invocations"`
	Errors        int64                  `json:"errors"`
	Duration      MetricStats            `json:"duration"`
	ColdStarts    int64                  `json:"cold_starts"`
	Concurrency   MetricStats            `json:"concurrency"`
	Memory        MetricStats            `json:"memory,omitempty"`
	CPU           MetricStats            `json:"cpu,omitempty"`
	CustomMetrics map[string]interface{} `json:"custom_metrics,omitempty"`
}

// MetricStats represents statistical data for a metric
type MetricStats struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Avg    float64 `json:"avg"`
	P50    float64 `json:"p50"`
	P95    float64 `json:"p95"`
	P99    float64 `json:"p99"`
}