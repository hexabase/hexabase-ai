package domain

import (
	"context"
	"io"
	"time"
)

// PipelineStatus represents the current status of a pipeline run
type PipelineStatus string

const (
	PipelineStatusPending   PipelineStatus = "pending"
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusSucceeded PipelineStatus = "succeeded"
	PipelineStatusFailed    PipelineStatus = "failed"
	PipelineStatusCancelled PipelineStatus = "cancelled"
)

// BuildType represents the type of build
type BuildType string

const (
	BuildTypeDocker    BuildType = "docker"
	BuildTypeBuildpack BuildType = "buildpack"
	BuildTypeCustom    BuildType = "custom"
)

// PipelineRun represents a CI/CD pipeline execution
type PipelineRun struct {
	ID          string         `json:"id"`
	WorkspaceID string         `json:"workspace_id"`
	ProjectID   string         `json:"project_id"`
	Name        string         `json:"name"`
	Status      PipelineStatus `json:"status"`
	StartedAt   time.Time      `json:"started_at"`
	FinishedAt  *time.Time     `json:"finished_at,omitempty"`
	Stages      []StageStatus  `json:"stages"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// StageStatus represents the status of a pipeline stage
type StageStatus struct {
	Name       string         `json:"name"`
	Status     PipelineStatus `json:"status"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	Tasks      []TaskStatus   `json:"tasks"`
}

// TaskStatus represents the status of a pipeline task
type TaskStatus struct {
	Name       string         `json:"name"`
	Status     PipelineStatus `json:"status"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	ExitCode   *int           `json:"exit_code,omitempty"`
	Message    string         `json:"message,omitempty"`
}

// LogEntry represents a log line from pipeline execution
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Stage     string    `json:"stage"`
	Task      string    `json:"task"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// GitConfig represents Git repository configuration
type GitConfig struct {
	URL        string `json:"url"`
	Branch     string `json:"branch"`
	CommitSHA  string `json:"commit_sha,omitempty"`
	SSHKeyRef  string `json:"ssh_key_ref,omitempty"`  // Reference to K8s secret
	TokenRef   string `json:"token_ref,omitempty"`    // Reference to K8s secret
	Submodules bool   `json:"submodules,omitempty"`
}

// BuildConfig represents build configuration
type BuildConfig struct {
	Type           BuildType         `json:"type"`
	DockerfilePath string            `json:"dockerfile_path,omitempty"`
	BuildContext   string            `json:"build_context,omitempty"`
	BuildArgs      map[string]string `json:"build_args,omitempty"`
	Target         string            `json:"target,omitempty"`
	CacheFrom      []string          `json:"cache_from,omitempty"`
}

// DeployConfig represents deployment configuration
type DeployConfig struct {
	TargetNamespace string            `json:"target_namespace"`
	ManifestPath    string            `json:"manifest_path,omitempty"`
	HelmChart       *HelmChartConfig  `json:"helm_chart,omitempty"`
	Kustomize       *KustomizeConfig  `json:"kustomize,omitempty"`
	Environment     map[string]string `json:"environment,omitempty"`
}

// HelmChartConfig represents Helm chart deployment configuration
type HelmChartConfig struct {
	Repository string            `json:"repository"`
	Chart      string            `json:"chart"`
	Version    string            `json:"version"`
	Values     map[string]any    `json:"values,omitempty"`
	ValuesFile string            `json:"values_file,omitempty"`
	Wait       bool              `json:"wait,omitempty"`
	Timeout    string            `json:"timeout,omitempty"`
}

// KustomizeConfig represents Kustomize deployment configuration
type KustomizeConfig struct {
	Path     string   `json:"path"`
	Overlays []string `json:"overlays,omitempty"`
}

// RegistryConfig represents container registry configuration
type RegistryConfig struct {
	URL         string `json:"url"`
	Namespace   string `json:"namespace"`
	CredRef     string `json:"cred_ref"` // Reference to K8s secret
	Insecure    bool   `json:"insecure,omitempty"`
}

// CredentialRefs represents references to credentials stored as K8s secrets
type CredentialRefs struct {
	GitSSHKey          string `json:"git_ssh_key,omitempty"`
	GitToken           string `json:"git_token,omitempty"`
	RegistryAuth       string `json:"registry_auth,omitempty"`
	KubeconfigSecret   string `json:"kubeconfig_secret,omitempty"`
}

// PipelineConfig represents the configuration for a pipeline run
type PipelineConfig struct {
	WorkspaceID    string          `json:"workspace_id"`
	ProjectID      string          `json:"project_id"`
	Name           string          `json:"name"`
	GitRepo        GitConfig       `json:"git_repo"`
	BuildConfig    *BuildConfig    `json:"build_config,omitempty"`
	DeployConfig   *DeployConfig   `json:"deploy_config,omitempty"`
	RegistryConfig *RegistryConfig `json:"registry_config,omitempty"`
	Credentials    CredentialRefs  `json:"credentials"`
	ServiceAccount string          `json:"service_account,omitempty"`
	Timeout        string          `json:"timeout,omitempty"`
	Metadata       map[string]any  `json:"metadata,omitempty"`
}

// PipelineTemplate represents a reusable pipeline template
type PipelineTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Provider    string                 `json:"provider"`
	Stages      []StageTemplate        `json:"stages"`
	Parameters  []ParameterDefinition  `json:"parameters"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// StageTemplate represents a stage in a pipeline template
type StageTemplate struct {
	Name      string         `json:"name"`
	Tasks     []TaskTemplate `json:"tasks"`
	DependsOn []string       `json:"depends_on,omitempty"`
	When      string         `json:"when,omitempty"` // Conditional execution
}

// TaskTemplate represents a task in a pipeline template
type TaskTemplate struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"` // e.g., "git-clone", "docker-build", "kubectl-apply"
	Parameters map[string]any    `json:"parameters"`
	Resources  ResourceRequests  `json:"resources,omitempty"`
	Timeout    string            `json:"timeout,omitempty"`
}

// ResourceRequests represents resource requirements for a task
type ResourceRequests struct {
	CPU    string `json:"cpu,omitempty"`    // e.g., "100m"
	Memory string `json:"memory,omitempty"` // e.g., "256Mi"
	Disk   string `json:"disk,omitempty"`   // e.g., "1Gi"
}

// ParameterDefinition represents a parameter in a pipeline template
type ParameterDefinition struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // string, number, boolean
	Description  string `json:"description"`
	Default      any    `json:"default,omitempty"`
	Required     bool   `json:"required"`
	AllowedValues []any  `json:"allowed_values,omitempty"`
}

// SecretReference represents a reference to a Kubernetes secret
type SecretReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Key       string `json:"key,omitempty"`
}

// ProviderConfig represents provider-specific configuration
type ProviderConfig struct {
	Type     string         `json:"type"` // e.g., "tekton", "github-actions", "gitlab-ci"
	Settings map[string]any `json:"settings"`
}

// GitCredential represents Git authentication credentials
type GitCredential struct {
	Type        string `json:"type"` // "ssh-key" or "token"
	SSHKey      string `json:"ssh_key,omitempty"`      // Private SSH key
	Token       string `json:"token,omitempty"`        // Git token (GitHub, GitLab, etc.)
	Username    string `json:"username,omitempty"`     // Username for token auth
	Passphrase  string `json:"passphrase,omitempty"`   // SSH key passphrase
}

// RegistryCredential represents container registry authentication credentials
type RegistryCredential struct {
	Registry string `json:"registry"`           // Registry URL
	Username string `json:"username"`           // Registry username
	Password string `json:"password"`           // Registry password/token
	Email    string `json:"email,omitempty"`    // Registry email
}

// CredentialInfo represents metadata about stored credentials
type CredentialInfo struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "git" or "registry"
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Provider interface defines the contract for CI/CD providers
type Provider interface {
	// Configuration operations
	ValidateConfig(ctx context.Context, config *PipelineConfig) error
	
	// Pipeline operations
	RunPipeline(ctx context.Context, config *PipelineConfig) (*PipelineRun, error)
	GetStatus(ctx context.Context, workspaceID, runID string) (*PipelineRun, error)
	ListPipelines(ctx context.Context, workspaceID, projectID string) ([]*PipelineRun, error)
	CancelPipeline(ctx context.Context, workspaceID, runID string) error
	DeletePipeline(ctx context.Context, workspaceID, runID string) error
	
	// Log operations
	GetLogs(ctx context.Context, workspaceID, runID string) ([]LogEntry, error)
	StreamLogs(ctx context.Context, workspaceID, runID string) (io.ReadCloser, error)
	
	// Template operations
	GetTemplates(ctx context.Context) ([]*PipelineTemplate, error)
	CreateFromTemplate(ctx context.Context, templateID string, params map[string]any) (*PipelineConfig, error)
	
	// Provider information
	GetName() string
	GetVersion() string
	IsHealthy() bool
}

// ProviderFactory interface defines the contract for creating providers
type ProviderFactory interface {
	CreateProvider(providerType string, config *ProviderConfig) (Provider, error)
	ListProviders() []string
}

// CredentialManager interface defines the contract for managing credentials
type CredentialManager interface {
	// Git credentials
	StoreGitCredential(workspaceID string, cred *GitCredential) (*CredentialInfo, error)
	GetGitCredential(workspaceID, credentialID string) (*GitCredential, error)
	
	// Registry credentials
	StoreRegistryCredential(workspaceID string, cred *RegistryCredential) (*CredentialInfo, error)
	GetRegistryCredential(workspaceID, credentialID string) (*RegistryCredential, error)
	
	// General credential operations
	ListCredentials(workspaceID string) ([]*CredentialInfo, error)
	DeleteCredential(workspaceID, credentialID string) error
	
	// Kubernetes secret management
	CreateKubernetesSecret(workspaceID, secretName string, data map[string][]byte) error
	GetKubernetesSecret(workspaceID, secretName string) (map[string][]byte, error)
	DeleteKubernetesSecret(workspaceID, secretName string) error
}