package application

import (
	"time"
)

// ApplicationType represents the type of application (stateless or stateful)
type ApplicationType string

const (
	ApplicationTypeStateless ApplicationType = "stateless"
	ApplicationTypeStateful  ApplicationType = "stateful"
	ApplicationTypeCronJob   ApplicationType = "cronjob"
	ApplicationTypeFunction  ApplicationType = "function"
)

// ApplicationStatus represents the status of an application
type ApplicationStatus string

const (
	ApplicationStatusPending   ApplicationStatus = "pending"
	ApplicationStatusDeploying ApplicationStatus = "deploying"
	ApplicationStatusRunning   ApplicationStatus = "running"
	ApplicationStatusUpdating  ApplicationStatus = "updating"
	ApplicationStatusStopping  ApplicationStatus = "stopping"
	ApplicationStatusStopped   ApplicationStatus = "stopped"
	ApplicationStatusError     ApplicationStatus = "error"
	ApplicationStatusDeleting  ApplicationStatus = "deleting"
)

// SourceType represents the source type for the application
type SourceType string

const (
	SourceTypeImage SourceType = "image"
	SourceTypeGit   SourceType = "git"
)

// Application represents a deployed workload in a workspace
type Application struct {
	ID          string            `json:"id"`
	WorkspaceID string            `json:"workspace_id"`
	ProjectID   string            `json:"project_id"`
	Name        string            `json:"name"`
	Type        ApplicationType   `json:"type"`
	Status      ApplicationStatus `json:"status"`
	Source      ApplicationSource `json:"source"`
	Config      ApplicationConfig `json:"config"`
	Endpoints   []Endpoint        `json:"endpoints"`
	// CronJob specific fields
	CronSchedule    string     `json:"cron_schedule,omitempty"`
	CronCommand     []string   `json:"cron_command,omitempty"`
	CronArgs        []string   `json:"cron_args,omitempty"`
	TemplateAppID   string     `json:"template_app_id,omitempty"`
	LastExecutionAt *time.Time `json:"last_execution_at,omitempty"`
	NextExecutionAt *time.Time `json:"next_execution_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ApplicationSource represents the source configuration for an application
type ApplicationSource struct {
	Type     SourceType `json:"type"`
	Image    string     `json:"image,omitempty"`    // Container image for SourceTypeImage
	GitURL   string     `json:"git_url,omitempty"`  // Git repository URL for SourceTypeGit
	GitRef   string     `json:"git_ref,omitempty"`  // Git branch/tag/commit for SourceTypeGit
	Buildpack string    `json:"buildpack,omitempty"` // Buildpack for SourceTypeGit
}

// ApplicationConfig represents the configuration for an application
type ApplicationConfig struct {
	Replicas      int               `json:"replicas"`
	Port          int               `json:"port"`
	EnvVars       map[string]string `json:"env_vars"`
	Resources     ResourceRequests  `json:"resources"`
	NodeSelector  map[string]string `json:"node_selector,omitempty"`  // For dedicated node scheduling
	Storage       *StorageConfig    `json:"storage,omitempty"`        // For stateful apps
	NetworkConfig *NetworkConfig    `json:"network_config,omitempty"` // Optional network configuration
}

// ResourceRequests represents resource requests for an application
type ResourceRequests struct {
	CPURequest    string `json:"cpu_request"`    // e.g., "100m"
	CPULimit      string `json:"cpu_limit"`      // e.g., "500m"
	MemoryRequest string `json:"memory_request"` // e.g., "128Mi"
	MemoryLimit   string `json:"memory_limit"`   // e.g., "512Mi"
}

// StorageConfig represents storage configuration for stateful applications
type StorageConfig struct {
	Size         string `json:"size"`          // e.g., "10Gi"
	StorageClass string `json:"storage_class"` // Storage class name
	MountPath    string `json:"mount_path"`    // Mount path in container
}

// NetworkConfig represents network configuration for an application
type NetworkConfig struct {
	CreateIngress bool              `json:"create_ingress"`
	IngressPath   string            `json:"ingress_path,omitempty"`
	CustomDomain  string            `json:"custom_domain,omitempty"`
	TLSEnabled    bool              `json:"tls_enabled"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

// Endpoint represents a public endpoint for an application
type Endpoint struct {
	Type string `json:"type"` // "ingress" or "service"
	URL  string `json:"url"`
}

// Pod represents a running instance of an application
type Pod struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	NodeName  string    `json:"node_name"`
	IP        string    `json:"ip"`
	StartTime time.Time `json:"start_time"`
	Restarts  int       `json:"restarts"`
}

// ApplicationEvent represents an event in the application lifecycle
type ApplicationEvent struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"application_id"`
	Type          string    `json:"type"`
	Message       string    `json:"message"`
	Details       string    `json:"details,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// CreateApplicationRequest represents a request to create a new application
type CreateApplicationRequest struct {
	Name          string            `json:"name"`
	Type          ApplicationType   `json:"type"`
	Source        ApplicationSource `json:"source"`
	Config        ApplicationConfig `json:"config"`
	ProjectID     string            `json:"project_id"`
	NodePoolID    string            `json:"node_pool_id,omitempty"` // Optional target node pool
	// CronJob specific fields
	CronSchedule  string   `json:"cron_schedule,omitempty"`
	CronCommand   []string `json:"cron_command,omitempty"`
	CronArgs      []string `json:"cron_args,omitempty"`
	TemplateAppID string   `json:"template_app_id,omitempty"`
}

// UpdateApplicationRequest represents a request to update an application
type UpdateApplicationRequest struct {
	Replicas      *int               `json:"replicas,omitempty"`
	ImageVersion  string             `json:"image_version,omitempty"`
	EnvVars       map[string]string  `json:"env_vars,omitempty"`
	Resources     *ResourceRequests  `json:"resources,omitempty"`
	NetworkConfig *NetworkConfig     `json:"network_config,omitempty"`
}

// ApplicationMetrics represents metrics for an application
type ApplicationMetrics struct {
	ApplicationID string                 `json:"application_id"`
	Timestamp     time.Time              `json:"timestamp"`
	PodMetrics    []PodMetrics           `json:"pod_metrics"`
	AggregateUsage AggregateResourceUsage `json:"aggregate_usage"`
}

// PodMetrics represents metrics for a single pod
type PodMetrics struct {
	PodName      string  `json:"pod_name"`
	CPUUsage     float64 `json:"cpu_usage"`      // in cores
	MemoryUsage  float64 `json:"memory_usage"`   // in MB
	NetworkIn    float64 `json:"network_in"`     // in MB/s
	NetworkOut   float64 `json:"network_out"`    // in MB/s
}

// AggregateResourceUsage represents aggregated resource usage across all pods
type AggregateResourceUsage struct {
	TotalCPU     float64 `json:"total_cpu"`
	TotalMemory  float64 `json:"total_memory"`
	AverageCPU   float64 `json:"average_cpu"`
	AverageMemory float64 `json:"average_memory"`
}

// LogEntry represents a log entry from an application
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	PodName   string    `json:"pod_name"`
	Container string    `json:"container"`
	Level     string    `json:"level,omitempty"`
	Message   string    `json:"message"`
}

// LogQuery represents parameters for querying application logs
type LogQuery struct {
	ApplicationID string    `json:"application_id"`
	PodName       string    `json:"pod_name,omitempty"`
	Container     string    `json:"container,omitempty"`
	Since         time.Time `json:"since,omitempty"`
	Until         time.Time `json:"until,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Follow        bool      `json:"follow,omitempty"` // For real-time streaming
}

// ValidateApplicationType checks if the application type is valid
func (a ApplicationType) IsValid() bool {
	switch a {
	case ApplicationTypeStateless, ApplicationTypeStateful, ApplicationTypeCronJob, ApplicationTypeFunction:
		return true
	default:
		return false
	}
}

// ValidateSourceType checks if the source type is valid
func (s SourceType) IsValid() bool {
	switch s {
	case SourceTypeImage, SourceTypeGit:
		return true
	default:
		return false
	}
}

// CanTransition checks if the application can transition to the target status
func (s ApplicationStatus) CanTransition(target ApplicationStatus) bool {
	transitions := map[ApplicationStatus][]ApplicationStatus{
		ApplicationStatusPending:   {ApplicationStatusDeploying, ApplicationStatusError},
		ApplicationStatusDeploying: {ApplicationStatusRunning, ApplicationStatusError},
		ApplicationStatusRunning:   {ApplicationStatusUpdating, ApplicationStatusStopping, ApplicationStatusError},
		ApplicationStatusUpdating:  {ApplicationStatusRunning, ApplicationStatusError},
		ApplicationStatusStopping:  {ApplicationStatusStopped, ApplicationStatusError},
		ApplicationStatusStopped:   {ApplicationStatusDeploying, ApplicationStatusDeleting},
		ApplicationStatusError:     {ApplicationStatusDeploying, ApplicationStatusDeleting},
		ApplicationStatusDeleting:  {}, // Terminal state
	}

	allowed, exists := transitions[s]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// CronJobExecutionStatus represents the status of a CronJob execution
type CronJobExecutionStatus string

const (
	CronJobExecutionStatusRunning   CronJobExecutionStatus = "running"
	CronJobExecutionStatusSucceeded CronJobExecutionStatus = "succeeded"
	CronJobExecutionStatusFailed    CronJobExecutionStatus = "failed"
)

// CronJobExecution represents a single execution of a CronJob
type CronJobExecution struct {
	ID            string                 `json:"id"`
	ApplicationID string                 `json:"application_id"`
	JobName       string                 `json:"job_name"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Status        CronJobExecutionStatus `json:"status"`
	ExitCode      *int                   `json:"exit_code,omitempty"`
	Logs          string                 `json:"logs,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// CronJobExecutionList represents a list of CronJob executions
type CronJobExecutionList struct {
	Executions []CronJobExecution `json:"executions"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
}

// TriggerCronJobRequest represents a request to manually trigger a CronJob
type TriggerCronJobRequest struct {
	ApplicationID string `json:"application_id"`
}

// UpdateCronScheduleRequest represents a request to update a CronJob's schedule
type UpdateCronScheduleRequest struct {
	Schedule string `json:"schedule"`
}