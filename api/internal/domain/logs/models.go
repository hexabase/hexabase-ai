// Package logs defines the domain models and interfaces for logging.
package logs

import "time"

// LogEntry represents a single structured log record.
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	TraceID   string                 `json:"trace_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Source    string                 `json:"source,omitempty"` // e.g., "api-server", "ai-ops"
	Details   map[string]interface{} `json:"details,omitempty"`
}

// LogQuery represents the parameters for a log query.
type LogQuery struct {
	WorkspaceID string    `json:"workspace_id"`
	SearchTerm  string    `json:"search_term,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Level       string    `json:"level,omitempty"`
	Limit       int       `json:"limit,omitempty"`
} 