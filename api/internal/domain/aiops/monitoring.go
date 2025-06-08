package aiops

import (
	"context"
	"time"
)

// LLMOpsMonitor defines the interface for LLM operations monitoring
type LLMOpsMonitor interface {
	// Trace management
	CreateTrace(ctx context.Context, trace *Trace) error
	CreateGeneration(ctx context.Context, generation *Generation) error
	CreateSpan(ctx context.Context, span *Span) error
	UpdateGeneration(ctx context.Context, generationID string, updates *GenerationUpdate) error
	
	// Scoring and feedback
	ScoreGeneration(ctx context.Context, generationID string, score *Score) error
	ScoreTrace(ctx context.Context, traceID string, score *Score) error
	
	// Analytics
	GetTraceMetrics(ctx context.Context, filter *MetricsFilter) (*TraceMetrics, error)
	GetModelMetrics(ctx context.Context, filter *MetricsFilter) (*LLMModelMetrics, error)
	GetUserMetrics(ctx context.Context, filter *MetricsFilter) (*UserMetrics, error)
	
	// Dataset management
	CreateDataset(ctx context.Context, dataset *Dataset) error
	AddToDataset(ctx context.Context, datasetID string, item *DatasetItem) error
	GetDataset(ctx context.Context, datasetID string) (*Dataset, error)
	
	// Health and diagnostics
	IsHealthy(ctx context.Context) bool
	Flush(ctx context.Context) error
}

// Trace represents a complete LLM interaction trace
type Trace struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	UserID      string            `json:"user_id"`
	SessionID   string            `json:"session_id"`
	Timestamp   time.Time         `json:"timestamp"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Input       any               `json:"input,omitempty"`
	Output      any               `json:"output,omitempty"`
	Release     string            `json:"release,omitempty"`
	Version     string            `json:"version,omitempty"`
}

// Generation represents a single LLM generation within a trace
type Generation struct {
	ID               string            `json:"id"`
	TraceID          string            `json:"trace_id"`
	ParentID         string            `json:"parent_id,omitempty"`
	Name             string            `json:"name"`
	Model            string            `json:"model"`
	ModelParameters  map[string]any    `json:"model_parameters,omitempty"`
	Input            any               `json:"input"`
	Output           any               `json:"output,omitempty"`
	StartTime        time.Time         `json:"start_time"`
	EndTime          *time.Time        `json:"end_time,omitempty"`
	CompletionTokens int               `json:"completion_tokens,omitempty"`
	PromptTokens     int               `json:"prompt_tokens,omitempty"`
	TotalTokens      int               `json:"total_tokens,omitempty"`
	StatusMessage    string            `json:"status_message,omitempty"`
	Level            ObservationLevel  `json:"level,omitempty"`
	Metadata         map[string]any    `json:"metadata,omitempty"`
}

// GenerationUpdate represents updates to a generation
type GenerationUpdate struct {
	Output           any               `json:"output,omitempty"`
	EndTime          *time.Time        `json:"end_time,omitempty"`
	CompletionTokens *int              `json:"completion_tokens,omitempty"`
	PromptTokens     *int              `json:"prompt_tokens,omitempty"`
	TotalTokens      *int              `json:"total_tokens,omitempty"`
	StatusMessage    *string           `json:"status_message,omitempty"`
	Metadata         map[string]any    `json:"metadata,omitempty"`
}

// Span represents a non-LLM operation within a trace
type Span struct {
	ID            string            `json:"id"`
	TraceID       string            `json:"trace_id"`
	ParentID      string            `json:"parent_id,omitempty"`
	Name          string            `json:"name"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       *time.Time        `json:"end_time,omitempty"`
	Input         any               `json:"input,omitempty"`
	Output        any               `json:"output,omitempty"`
	StatusMessage string            `json:"status_message,omitempty"`
	Level         ObservationLevel  `json:"level,omitempty"`
	Metadata      map[string]any    `json:"metadata,omitempty"`
}

// Score represents a score/feedback for a generation or trace
type Score struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Value        float64        `json:"value"`
	StringValue  string         `json:"string_value,omitempty"`
	Comment      string         `json:"comment,omitempty"`
	DataType     ScoreDataType  `json:"data_type"`
	Timestamp    time.Time      `json:"timestamp"`
	ObserverID   string         `json:"observer_id,omitempty"`
}

// ObservationLevel represents the level of detail for observations
type ObservationLevel string

const (
	ObservationLevelDebug   ObservationLevel = "debug"
	ObservationLevelDefault ObservationLevel = "default"
	ObservationLevelWarning ObservationLevel = "warning"
	ObservationLevelError   ObservationLevel = "error"
)

// ScoreDataType represents the data type of a score
type ScoreDataType string

const (
	ScoreDataTypeNumeric ScoreDataType = "numeric"
	ScoreDataTypeBoolean ScoreDataType = "boolean"
	ScoreDataTypeCategorical ScoreDataType = "categorical"
)

// MetricsFilter represents filters for metrics queries
type MetricsFilter struct {
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	UserID      string            `json:"user_id,omitempty"`
	SessionID   string            `json:"session_id,omitempty"`
	Model       string            `json:"model,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Release     string            `json:"release,omitempty"`
}

// TraceMetrics represents aggregated trace metrics
type TraceMetrics struct {
	TotalTraces      int               `json:"total_traces"`
	TotalGenerations int               `json:"total_generations"`
	TotalTokens      int               `json:"total_tokens"`
	AverageLatency   time.Duration     `json:"average_latency"`
	SuccessRate      float64           `json:"success_rate"`
	ErrorRate        float64           `json:"error_rate"`
	TokensPerTrace   float64           `json:"tokens_per_trace"`
	CostEstimate     float64           `json:"cost_estimate"`
	ScoreDistribution map[string]float64 `json:"score_distribution,omitempty"`
}

// LLMModelMetrics represents metrics specific to a model in monitoring
type LLMModelMetrics struct {
	ModelName        string            `json:"model_name"`
	TotalGenerations int               `json:"total_generations"`
	TotalTokens      int               `json:"total_tokens"`
	AverageLatency   time.Duration     `json:"average_latency"`
	TokensPerSecond  float64           `json:"tokens_per_second"`
	CostPerToken     float64           `json:"cost_per_token"`
	ErrorRate        float64           `json:"error_rate"`
	ModelVersions    []string          `json:"model_versions,omitempty"`
	ScoreStats       map[string]float64 `json:"score_stats,omitempty"`
}

// UserMetrics represents metrics specific to a user
type UserMetrics struct {
	UserID           string            `json:"user_id"`
	TotalTraces      int               `json:"total_traces"`
	TotalGenerations int               `json:"total_generations"`
	TotalTokens      int               `json:"total_tokens"`
	TotalCost        float64           `json:"total_cost"`
	ModelsUsed       []string          `json:"models_used"`
	AverageScores    map[string]float64 `json:"average_scores,omitempty"`
	ActivityByDay    map[string]int    `json:"activity_by_day,omitempty"`
}

// Dataset represents a dataset for evaluation or fine-tuning
type Dataset struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ItemCount   int            `json:"item_count"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// DatasetItem represents a single item in a dataset
type DatasetItem struct {
	ID              string         `json:"id"`
	Input           any            `json:"input"`
	ExpectedOutput  any            `json:"expected_output,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	SourceTraceID   string         `json:"source_trace_id,omitempty"`
	SourceObservationID string     `json:"source_observation_id,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}