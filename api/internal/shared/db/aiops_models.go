package db

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// ChatSession represents a chat session in the database
type ChatSession struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	WorkspaceID string         `gorm:"not null;index" json:"workspace_id"`
	UserID      string         `gorm:"not null;index" json:"user_id"`
	Title       string         `gorm:"not null" json:"title"`
	Model       string         `gorm:"not null" json:"model"`
	Messages    JSONB          `gorm:"type:jsonb;default:'[]'" json:"messages"`
	Context     pq.Int64Array  `gorm:"type:integer[]" json:"context,omitempty"`
	Metadata    JSONB          `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt   time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
}

// TableName specifies the table name
func (ChatSession) TableName() string {
	return "chat_sessions"
}

// ModelUsage represents model usage statistics in the database
type ModelUsage struct {
	ID                string    `gorm:"primaryKey" json:"id"`
	WorkspaceID       string    `gorm:"not null;index" json:"workspace_id"`
	UserID            string    `gorm:"not null;index" json:"user_id"`
	SessionID         *string   `gorm:"index" json:"session_id,omitempty"`
	ModelName         string    `gorm:"not null;index" json:"model_name"`
	PromptTokens      int       `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens  int       `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens       int       `gorm:"not null;default:0" json:"total_tokens"`
	RequestDurationMs int64     `gorm:"not null;default:0" json:"request_duration_ms"`
	Timestamp         time.Time `gorm:"not null;index" json:"timestamp"`
	Metadata          JSONB     `gorm:"type:jsonb;default:'{}'" json:"metadata"`
}

// TableName specifies the table name
func (ModelUsage) TableName() string {
	return "model_usage"
}

// JSONB is a wrapper for handling JSONB data type
type JSONB json.RawMessage

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB("null")
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*j = JSONB("null")
		} else {
			*j = JSONB(v)
		}
	case string:
		if v == "" {
			*j = JSONB("null")
		} else {
			*j = JSONB(v)
		}
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		*j = JSONB(data)
	}

	return nil
}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 || string(j) == "null" {
		return nil, nil
	}
	return []byte(j), nil
}

// MarshalJSON implements json.Marshaler
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return nil
	}
	*j = JSONB(data)
	return nil
}