package db

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Application represents a deployed workload in the database
type Application struct {
	ID              string         `gorm:"primaryKey" json:"id"`
	WorkspaceID     string         `gorm:"not null;index" json:"workspace_id"`
	ProjectID       string         `gorm:"not null;index" json:"project_id"`
	Name            string         `gorm:"not null" json:"name"`
	Type            string         `gorm:"not null;default:'stateless'" json:"type"`
	Status          string         `gorm:"not null;default:'pending'" json:"status"`
	SourceType      string         `gorm:"not null" json:"source_type"`
	SourceImage     string         `json:"source_image,omitempty"`
	SourceGitURL    string         `json:"source_git_url,omitempty"`
	SourceGitRef    string         `json:"source_git_ref,omitempty"`
	Config          JSON           `gorm:"type:jsonb" json:"config"`
	Endpoints       JSON           `gorm:"type:jsonb" json:"endpoints"`
	// CronJob specific fields
	CronSchedule    *string        `json:"cron_schedule,omitempty"`
	CronCommand     pq.StringArray `gorm:"type:text[]" json:"cron_command,omitempty"`
	CronArgs        pq.StringArray `gorm:"type:text[]" json:"cron_args,omitempty"`
	TemplateAppID   *string        `gorm:"index" json:"template_app_id,omitempty"`
	LastExecutionAt *time.Time     `json:"last_execution_at,omitempty"`
	NextExecutionAt *time.Time     `json:"next_execution_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Function specific fields
	FunctionRuntime       *string `json:"function_runtime,omitempty"`
	FunctionHandler       *string `json:"function_handler,omitempty"`
	FunctionTimeout       *int    `json:"function_timeout,omitempty"`
	FunctionMemory        *int    `json:"function_memory,omitempty"`
	FunctionTriggerType   *string `json:"function_trigger_type,omitempty"`
	FunctionTriggerConfig JSON    `gorm:"type:jsonb" json:"function_trigger_config,omitempty"`
	FunctionEnvVars       JSON    `gorm:"type:jsonb" json:"function_env_vars,omitempty"`
	FunctionSecrets       JSON    `gorm:"type:jsonb" json:"function_secrets,omitempty"`

	// Associations
	Workspace        Workspace            `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project          Project              `gorm:"foreignKey:ProjectID" json:"-"`
	TemplateApp      *Application         `gorm:"foreignKey:TemplateAppID" json:"-"`
	Executions       []CronJobExecution   `gorm:"foreignKey:ApplicationID" json:"-"`
	FunctionVersions []FunctionVersion    `gorm:"foreignKey:ApplicationID" json:"-"`
	FunctionInvocations []FunctionInvocation `gorm:"foreignKey:ApplicationID" json:"-"`
	FunctionEvents   []FunctionEvent      `gorm:"foreignKey:ApplicationID" json:"-"`
}

// BeforeCreate sets ID if not provided
func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = "app-" + uuid.New().String()
	}
	return nil
}

// CronJobExecution represents a single execution of a CronJob
type CronJobExecution struct {
	ID            string     `gorm:"primaryKey" json:"id"`
	ApplicationID string     `gorm:"not null;index" json:"application_id"`
	JobName       string     `gorm:"not null" json:"job_name"`
	StartedAt     time.Time  `gorm:"not null" json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	Status        string     `gorm:"not null;default:'running'" json:"status"`
	ExitCode      *int       `json:"exit_code,omitempty"`
	Logs          string     `gorm:"type:text" json:"logs,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Association
	Application Application `gorm:"foreignKey:ApplicationID" json:"-"`
}

// BeforeCreate sets ID if not provided
func (c *CronJobExecution) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = "cje-" + uuid.New().String()
	}
	return nil
}

// JSON is a custom type for JSONB fields
type JSON json.RawMessage

// Value implements the driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Scan implements the sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = JSON("null")
		return nil
	}
	
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	
	*j = JSON(data)
	return nil
}

// MarshalJSON implements json.Marshaler
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSON) UnmarshalJSON(data []byte) error {
	*j = JSON(data)
	return nil
}

// FunctionVersion represents a version of a function in the database
type FunctionVersion struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	ApplicationID string    `gorm:"not null;index" json:"application_id"`
	VersionNumber int       `gorm:"not null" json:"version_number"`
	SourceCode    string    `gorm:"type:text" json:"source_code,omitempty"`
	SourceType    string    `gorm:"not null" json:"source_type"`
	SourceURL     string    `json:"source_url,omitempty"`
	BuildLogs     string    `gorm:"type:text" json:"build_logs,omitempty"`
	BuildStatus   string    `gorm:"not null;default:'pending'" json:"build_status"`
	ImageURI      string    `json:"image_uri,omitempty"`
	IsActive      bool      `gorm:"default:false" json:"is_active"`
	DeployedAt    *time.Time `json:"deployed_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Association
	Application Application `gorm:"foreignKey:ApplicationID" json:"-"`
}

// BeforeCreate sets ID if not provided
func (f *FunctionVersion) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = "fv-" + uuid.New().String()
	}
	return nil
}

// FunctionInvocation represents a single function invocation in the database
type FunctionInvocation struct {
	ID              string     `gorm:"primaryKey" json:"id"`
	ApplicationID   string     `gorm:"not null;index" json:"application_id"`
	VersionID       string     `gorm:"index" json:"version_id,omitempty"`
	InvocationID    string     `gorm:"not null;uniqueIndex" json:"invocation_id"`
	TriggerSource   string     `gorm:"not null" json:"trigger_source"`
	RequestMethod   string     `json:"request_method,omitempty"`
	RequestPath     string     `gorm:"type:text" json:"request_path,omitempty"`
	RequestHeaders  JSON       `gorm:"type:jsonb" json:"request_headers,omitempty"`
	RequestBody     string     `gorm:"type:text" json:"request_body,omitempty"`
	ResponseStatus  int        `json:"response_status,omitempty"`
	ResponseHeaders JSON       `gorm:"type:jsonb" json:"response_headers,omitempty"`
	ResponseBody    string     `gorm:"type:text" json:"response_body,omitempty"`
	ErrorMessage    string     `gorm:"type:text" json:"error_message,omitempty"`
	DurationMs      int        `json:"duration_ms,omitempty"`
	ColdStart       bool       `gorm:"default:false" json:"cold_start"`
	MemoryUsed      int        `json:"memory_used,omitempty"`
	StartedAt       time.Time  `gorm:"not null" json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`

	// Associations
	Application Application      `gorm:"foreignKey:ApplicationID" json:"-"`
	Version     *FunctionVersion `gorm:"foreignKey:VersionID" json:"-"`
}

// BeforeCreate sets ID if not provided
func (f *FunctionInvocation) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = "fi-" + uuid.New().String()
	}
	return nil
}

// FunctionEvent represents an event for event-driven functions in the database
type FunctionEvent struct {
	ID               string     `gorm:"primaryKey" json:"id"`
	ApplicationID    string     `gorm:"not null;index" json:"application_id"`
	EventType        string     `gorm:"not null;index" json:"event_type"`
	EventSource      string     `gorm:"not null" json:"event_source"`
	EventData        JSON       `gorm:"type:jsonb;not null" json:"event_data"`
	ProcessingStatus string     `gorm:"not null;default:'pending';index" json:"processing_status"`
	RetryCount       int        `gorm:"default:0" json:"retry_count"`
	MaxRetries       int        `gorm:"default:3" json:"max_retries"`
	InvocationID     string     `gorm:"index" json:"invocation_id,omitempty"`
	ErrorMessage     string     `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt        time.Time  `gorm:"index" json:"created_at"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`

	// Associations
	Application Application         `gorm:"foreignKey:ApplicationID" json:"-"`
	Invocation  *FunctionInvocation `gorm:"foreignKey:InvocationID" json:"-"`
}

// BeforeCreate sets ID if not provided
func (f *FunctionEvent) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = "fe-" + uuid.New().String()
	}
	return nil
}