package aiops

import (
	"context"
	"time"
)

// Service defines the AIOps service interface
type Service interface {
	// Chat operations
	CreateChatSession(ctx context.Context, workspaceID, userID, title, model string) (*ChatSession, error)
	SendMessage(ctx context.Context, sessionID string, message ChatMessage) (*ChatResponse, error)
	StreamMessage(ctx context.Context, sessionID string, message ChatMessage) (<-chan *ChatStreamResponse, error)
	GetChatSession(ctx context.Context, sessionID string) (*ChatSession, error)
	ListChatSessions(ctx context.Context, workspaceID string, limit, offset int) ([]*ChatSession, error)
	DeleteChatSession(ctx context.Context, sessionID string) error
	
	// Model management
	ListAvailableModels(ctx context.Context) ([]*ModelInfo, error)
	EnsureModelAvailable(ctx context.Context, modelName string) error
	GetModelInfo(ctx context.Context, modelName string) (*ModelInfo, error)
	
	// Analytics and monitoring
	GetUsageStats(ctx context.Context, workspaceID string, from, to time.Time) (*UsageReport, error)
	GetModelMetrics(ctx context.Context, modelName string, from, to time.Time) (*ModelMetrics, error)
	
	// Health and diagnostics
	HealthCheck(ctx context.Context) *HealthStatus
}

// UsageReport represents usage statistics for a workspace
type UsageReport struct {
	WorkspaceID      string                 `json:"workspace_id"`
	Period          Period                 `json:"period"`
	TotalSessions   int                    `json:"total_sessions"`
	TotalMessages   int                    `json:"total_messages"`
	TotalTokens     int                    `json:"total_tokens"`
	ModelBreakdown  map[string]int         `json:"model_breakdown"`
	DailyUsage      []DailyUsage          `json:"daily_usage"`
	TopUsers        []UserUsage           `json:"top_users"`
	AverageSessionLength time.Duration     `json:"average_session_length"`
}

// Period represents a time period
type Period struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// DailyUsage represents usage for a specific day
type DailyUsage struct {
	Date     time.Time `json:"date"`
	Sessions int       `json:"sessions"`
	Messages int       `json:"messages"`
	Tokens   int       `json:"tokens"`
}

// UserUsage represents usage by a specific user
type UserUsage struct {
	UserID   string `json:"user_id"`
	Sessions int    `json:"sessions"`
	Messages int    `json:"messages"`
	Tokens   int    `json:"tokens"`
}

// ModelMetrics represents metrics for a specific model
type ModelMetrics struct {
	ModelName       string        `json:"model_name"`
	Period          Period        `json:"period"`
	TotalRequests   int           `json:"total_requests"`
	TotalTokens     int           `json:"total_tokens"`
	AverageLatency  time.Duration `json:"average_latency"`
	ErrorRate       float64       `json:"error_rate"`
	ThroughputRPS   float64       `json:"throughput_rps"`
	SuccessRate     float64       `json:"success_rate"`
}

// HealthStatus represents the health status of AIOps services
type HealthStatus struct {
	Status       string             `json:"status"`
	Timestamp    time.Time          `json:"timestamp"`
	Services     map[string]ServiceHealth `json:"services"`
	Version      string             `json:"version"`
	Uptime       time.Duration      `json:"uptime"`
}

// ServiceHealth represents health of individual services
type ServiceHealth struct {
	Status      string        `json:"status"`
	LastCheck   time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	Error       string        `json:"error,omitempty"`
}

// Service status constants
const (
	StatusHealthy   = "healthy"
	StatusDegraded  = "degraded"
	StatusUnhealthy = "unhealthy"
)