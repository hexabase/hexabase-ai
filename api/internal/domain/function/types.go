package function

import (
	"fmt"
	"time"
)

// FunctionSpec represents the specification for creating or updating a function
type FunctionSpec struct {
	Name        string                       `json:"name"`
	Namespace   string                       `json:"namespace"`
	Runtime     Runtime                      `json:"runtime"`
	Handler     string                       `json:"handler"`
	SourceCode  string                       `json:"source_code,omitempty"`
	SourceURL   string                       `json:"source_url,omitempty"`
	Image       string                       `json:"image,omitempty"`
	Environment map[string]string            `json:"environment,omitempty"`
	Secrets     map[string]string            `json:"secrets,omitempty"`
	Resources   FunctionResourceRequirements `json:"resources"`
	Timeout     int                          `json:"timeout"` // seconds
	Labels      map[string]string            `json:"labels,omitempty"`
	Annotations map[string]string            `json:"annotations,omitempty"`
}

// Runtime represents a function runtime environment
type Runtime string

const (
	RuntimeGo         Runtime = "go"
	RuntimePython     Runtime = "python"
	RuntimePython37   Runtime = "python37"
	RuntimePython38   Runtime = "python38"
	RuntimePython39   Runtime = "python39"
	RuntimeNode       Runtime = "node"
	RuntimeNode14     Runtime = "node14"
	RuntimeNode16     Runtime = "node16"
	RuntimeNode18     Runtime = "node18"
	RuntimeJava       Runtime = "java"
	RuntimeJava8      Runtime = "java8"
	RuntimeJava11     Runtime = "java11"
	RuntimeJava17     Runtime = "java17"
	RuntimeDotNet     Runtime = "dotnet"
	RuntimeDotNet31   Runtime = "dotnet31"
	RuntimeDotNet6    Runtime = "dotnet6"
	RuntimeRuby       Runtime = "ruby"
	RuntimePHP        Runtime = "php"
	RuntimeCustom     Runtime = "custom"
)

// IsValid checks if the runtime is valid
func (r Runtime) IsValid() bool {
	switch r {
	case RuntimeGo, RuntimePython, RuntimePython37, RuntimePython38, RuntimePython39,
		RuntimeNode, RuntimeNode14, RuntimeNode16, RuntimeNode18,
		RuntimeJava, RuntimeJava8, RuntimeJava11, RuntimeJava17,
		RuntimeDotNet, RuntimeDotNet31, RuntimeDotNet6,
		RuntimeRuby, RuntimePHP, RuntimeCustom:
		return true
	default:
		return false
	}
}

// GetBaseRuntime returns the base runtime without version
func (r Runtime) GetBaseRuntime() string {
	switch r {
	case RuntimePython37, RuntimePython38, RuntimePython39:
		return string(RuntimePython)
	case RuntimeNode14, RuntimeNode16, RuntimeNode18:
		return string(RuntimeNode)
	case RuntimeJava8, RuntimeJava11, RuntimeJava17:
		return string(RuntimeJava)
	case RuntimeDotNet31, RuntimeDotNet6:
		return string(RuntimeDotNet)
	default:
		return string(r)
	}
}

// FunctionResourceRequirements represents resource requirements for a function
type FunctionResourceRequirements struct {
	Memory string `json:"memory"` // e.g., "128Mi", "256Mi", "1Gi"
	CPU    string `json:"cpu"`    // e.g., "100m", "200m", "1"
}

// TriggerType represents the type of function trigger
type TriggerType string

const (
	TriggerHTTP          TriggerType = "http"
	TriggerSchedule      TriggerType = "schedule"
	TriggerEvent         TriggerType = "event"
	TriggerMessageQueue  TriggerType = "messagequeue"
	TriggerTimer         TriggerType = "timer"
)

// IsValid checks if the trigger type is valid
func (t TriggerType) IsValid() bool {
	switch t {
	case TriggerHTTP, TriggerSchedule, TriggerEvent, TriggerMessageQueue, TriggerTimer:
		return true
	default:
		return false
	}
}

// HTTPTriggerConfig represents configuration for HTTP triggers
type HTTPTriggerConfig struct {
	Method       string   `json:"method,omitempty"`       // GET, POST, etc. Empty means all methods
	Path         string   `json:"path,omitempty"`         // URL path pattern
	Host         string   `json:"host,omitempty"`         // Host header match
	AllowedHosts []string `json:"allowed_hosts,omitempty"` // Allowed hosts
}

// ScheduleTriggerConfig represents configuration for schedule triggers
type ScheduleTriggerConfig struct {
	Cron string `json:"cron"` // Cron expression
}

// EventTriggerConfig represents configuration for event triggers
type EventTriggerConfig struct {
	EventType   string            `json:"event_type"`
	EventSource string            `json:"event_source"`
	Filters     map[string]string `json:"filters,omitempty"`
}

// MessageQueueTriggerConfig represents configuration for message queue triggers
type MessageQueueTriggerConfig struct {
	QueueType       string `json:"queue_type"`       // e.g., "kafka", "rabbitmq", "nats"
	Topic           string `json:"topic"`            // Topic/Queue name
	Subscription    string `json:"subscription"`     // Subscription name (for pub/sub systems)
	ResponseTopic   string `json:"response_topic,omitempty"`   // Topic for responses
	ErrorTopic      string `json:"error_topic,omitempty"`      // Topic for errors
	MaxRetries      int    `json:"max_retries,omitempty"`      // Max retry attempts
	ContentType     string `json:"content_type,omitempty"`     // Expected content type
	PollingInterval string `json:"polling_interval,omitempty"` // For pull-based systems
}

// TimerTriggerConfig represents configuration for timer triggers
type TimerTriggerConfig struct {
	Interval string `json:"interval"` // e.g., "5m", "1h", "24h"
}

// ProviderError represents an error from a provider
type ProviderError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// IsNotFound returns true if the error indicates a resource was not found
func (e *ProviderError) IsNotFound() bool {
	return e.Code == "NOT_FOUND"
}

// IsAlreadyExists returns true if the error indicates a resource already exists
func (e *ProviderError) IsAlreadyExists() bool {
	return e.Code == "ALREADY_EXISTS"
}

// Common error codes
const (
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
	ErrCodeInvalidInput  = "INVALID_INPUT"
	ErrCodeTimeout       = "TIMEOUT"
	ErrCodeInternal      = "INTERNAL"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeQuotaExceeded = "QUOTA_EXCEEDED"
	ErrCodeNotSupported  = "NOT_SUPPORTED"
)

// NewProviderError creates a new provider error
func NewProviderError(code, message string) *ProviderError {
	return &ProviderError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *ProviderError) WithDetails(key string, value interface{}) *ProviderError {
	e.Details[key] = value
	return e
}

// InvocationMode represents how a function should be invoked
type InvocationMode string

const (
	InvocationModeSync  InvocationMode = "sync"
	InvocationModeAsync InvocationMode = "async"
)

// FunctionAuditEvent represents an audit event related to a function
type FunctionAuditEvent struct {
	ID           string            `json:"id"`
	WorkspaceID  string            `json:"workspace_id"`
	FunctionID   string            `json:"function_id"`
	Type         string            `json:"type"` // created, updated, deployed, invoked, error
	Description  string            `json:"description"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    string            `json:"created_at"`
}

// BuildOptions represents options for building a function
type BuildOptions struct {
	BuildCmd     string            `json:"build_cmd,omitempty"`
	BuildImage   string            `json:"build_image,omitempty"`
	BuildArgs    map[string]string `json:"build_args,omitempty"`
	Dependencies string            `json:"dependencies,omitempty"` // e.g., requirements.txt, package.json
}

// ScaleOptions represents autoscaling options for a function
type ScaleOptions struct {
	MinReplicas                    int    `json:"min_replicas"`
	MaxReplicas                    int    `json:"max_replicas"`
	TargetConcurrency              int    `json:"target_concurrency,omitempty"`
	ScaleDownDelay                 string `json:"scale_down_delay,omitempty"` // e.g., "5m"
	ScaleUpRate                    int    `json:"scale_up_rate,omitempty"`    // replicas per second
	MetricType                     string `json:"metric_type,omitempty"`      // e.g., "cpu", "memory", "concurrency"
	MetricTarget                   int    `json:"metric_target,omitempty"`    // target value for metric
}

// SecurityOptions represents security options for a function
type SecurityOptions struct {
	RunAsUser          int64             `json:"run_as_user,omitempty"`
	RunAsGroup         int64             `json:"run_as_group,omitempty"`
	ReadOnlyRootFS     bool              `json:"read_only_root_fs,omitempty"`
	AllowPrivilegeEsc  bool              `json:"allow_privilege_escalation,omitempty"`
	Capabilities       []string          `json:"capabilities,omitempty"`
	SeccompProfile     string            `json:"seccomp_profile,omitempty"`
	AppArmorProfile    string            `json:"app_armor_profile,omitempty"`
	SELinuxOptions     map[string]string `json:"se_linux_options,omitempty"`
}

// NetworkOptions represents network options for a function
type NetworkOptions struct {
	IngressEnabled     bool              `json:"ingress_enabled,omitempty"`
	IngressClass       string            `json:"ingress_class,omitempty"`
	IngressAnnotations map[string]string `json:"ingress_annotations,omitempty"`
	Domains            []string          `json:"domains,omitempty"`
	TLSEnabled         bool              `json:"tls_enabled,omitempty"`
	TLSSecretName      string            `json:"tls_secret_name,omitempty"`
}

// ObservabilityOptions represents observability options for a function
type ObservabilityOptions struct {
	TracingEnabled     bool              `json:"tracing_enabled,omitempty"`
	TracingEndpoint    string            `json:"tracing_endpoint,omitempty"`
	MetricsEnabled     bool              `json:"metrics_enabled,omitempty"`
	MetricsEndpoint    string            `json:"metrics_endpoint,omitempty"`
	LogLevel           string            `json:"log_level,omitempty"`
	CustomLabels       map[string]string `json:"custom_labels,omitempty"`
	CustomAnnotations  map[string]string `json:"custom_annotations,omitempty"`
}

// FunctionEvent represents an event that can trigger a function
type FunctionEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Time        time.Time              `json:"time"`
	Data        interface{}            `json:"data"`
	DataVersion string                 `json:"data_version,omitempty"`
	Subject     string                 `json:"subject,omitempty"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"`
}

// ProviderType represents the type of function provider
type ProviderType string

const (
	ProviderTypeMock    ProviderType = "mock"
	ProviderTypeKnative ProviderType = "knative"
	ProviderTypeFission ProviderType = "fission"
)

// ProviderConfig represents provider-specific configuration
type ProviderConfig struct {
	Type   ProviderType           `json:"type"`
	Config map[string]interface{} `json:"config"`
}