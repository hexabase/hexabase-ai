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

	// Associations
	Workspace    Workspace            `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project      Project              `gorm:"foreignKey:ProjectID" json:"-"`
	TemplateApp  *Application         `gorm:"foreignKey:TemplateAppID" json:"-"`
	Executions   []CronJobExecution   `gorm:"foreignKey:ApplicationID" json:"-"`
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