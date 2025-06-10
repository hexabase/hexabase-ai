package function

import (
	"context"
)

// Service defines the business logic for function management
type Service interface {
	// Function lifecycle management
	CreateFunction(ctx context.Context, workspaceID, projectID string, spec *FunctionSpec) (*FunctionDef, error)
	UpdateFunction(ctx context.Context, workspaceID, functionID string, spec *FunctionSpec) (*FunctionDef, error)
	DeleteFunction(ctx context.Context, workspaceID, functionID string) error
	GetFunction(ctx context.Context, workspaceID, functionID string) (*FunctionDef, error)
	ListFunctions(ctx context.Context, workspaceID, projectID string) ([]*FunctionDef, error)
	
	// Version management
	DeployVersion(ctx context.Context, workspaceID, functionID string, version *FunctionVersionDef) (*FunctionVersionDef, error)
	GetVersion(ctx context.Context, workspaceID, functionID, versionID string) (*FunctionVersionDef, error)
	ListVersions(ctx context.Context, workspaceID, functionID string) ([]*FunctionVersionDef, error)
	SetActiveVersion(ctx context.Context, workspaceID, functionID, versionID string) error
	RollbackVersion(ctx context.Context, workspaceID, functionID string) error
	
	// Trigger management
	CreateTrigger(ctx context.Context, workspaceID, functionID string, trigger *FunctionTrigger) (*FunctionTrigger, error)
	UpdateTrigger(ctx context.Context, workspaceID, functionID, triggerID string, trigger *FunctionTrigger) (*FunctionTrigger, error)
	DeleteTrigger(ctx context.Context, workspaceID, functionID, triggerID string) error
	ListTriggers(ctx context.Context, workspaceID, functionID string) ([]*FunctionTrigger, error)
	
	// Function invocation
	InvokeFunction(ctx context.Context, workspaceID, functionID string, request *InvokeRequest) (*InvokeResponse, error)
	InvokeFunctionAsync(ctx context.Context, workspaceID, functionID string, request *InvokeRequest) (string, error)
	GetInvocationStatus(ctx context.Context, workspaceID, invocationID string) (*InvocationStatus, error)
	ListInvocations(ctx context.Context, workspaceID, functionID string, limit int) ([]*InvocationStatus, error)
	
	// Function monitoring
	GetFunctionLogs(ctx context.Context, workspaceID, functionID string, opts *LogOptions) ([]*LogEntry, error)
	GetFunctionMetrics(ctx context.Context, workspaceID, functionID string, opts *MetricOptions) (*Metrics, error)
	GetFunctionEvents(ctx context.Context, workspaceID, functionID string, limit int) ([]*FunctionAuditEvent, error)
	
	// Provider management
	GetProviderCapabilities(ctx context.Context, workspaceID string) (*Capabilities, error)
	GetProviderHealth(ctx context.Context, workspaceID string) error
}

// Repository defines the data access layer for function management
type Repository interface {
	// Function metadata storage
	CreateFunction(ctx context.Context, function *FunctionDef) error
	UpdateFunction(ctx context.Context, function *FunctionDef) error
	DeleteFunction(ctx context.Context, workspaceID, functionID string) error
	GetFunction(ctx context.Context, workspaceID, functionID string) (*FunctionDef, error)
	ListFunctions(ctx context.Context, workspaceID, projectID string) ([]*FunctionDef, error)
	
	// Version metadata storage
	CreateVersion(ctx context.Context, version *FunctionVersionDef) error
	UpdateVersion(ctx context.Context, version *FunctionVersionDef) error
	GetVersion(ctx context.Context, workspaceID, functionID, versionID string) (*FunctionVersionDef, error)
	ListVersions(ctx context.Context, workspaceID, functionID string) ([]*FunctionVersionDef, error)
	
	// Trigger metadata storage
	CreateTrigger(ctx context.Context, trigger *FunctionTrigger) error
	UpdateTrigger(ctx context.Context, trigger *FunctionTrigger) error
	DeleteTrigger(ctx context.Context, workspaceID, functionID, triggerID string) error
	ListTriggers(ctx context.Context, workspaceID, functionID string) ([]*FunctionTrigger, error)
	
	// Invocation history storage
	CreateInvocation(ctx context.Context, invocation *InvocationStatus) error
	UpdateInvocation(ctx context.Context, invocation *InvocationStatus) error
	GetInvocation(ctx context.Context, workspaceID, invocationID string) (*InvocationStatus, error)
	ListInvocations(ctx context.Context, workspaceID, functionID string, limit int) ([]*InvocationStatus, error)
	
	// Event storage
	CreateEvent(ctx context.Context, event *FunctionAuditEvent) error
	ListEvents(ctx context.Context, workspaceID, functionID string, limit int) ([]*FunctionAuditEvent, error)
	
	// Workspace provider configuration
	GetWorkspaceProviderConfig(ctx context.Context, workspaceID string) (*ProviderConfig, error)
	UpdateWorkspaceProviderConfig(ctx context.Context, workspaceID string, config *ProviderConfig) error
}

