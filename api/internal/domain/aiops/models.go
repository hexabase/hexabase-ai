package aiops

import (
	"context"
	"time"
)

// LLMService defines the interface for LLM operations
type LLMService interface {
	// Chat operations
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	StreamChat(ctx context.Context, req *ChatRequest) (<-chan *ChatStreamResponse, error)
	
	// Model management
	ListModels(ctx context.Context) ([]*ModelInfo, error)
	PullModel(ctx context.Context, modelName string) error
	DeleteModel(ctx context.Context, modelName string) error
	
	// Health and diagnostics
	IsHealthy(ctx context.Context) bool
	GetModelInfo(ctx context.Context, modelName string) (*ModelInfo, error)
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Context     []int         `json:"context,omitempty"`
	Options     map[string]any `json:"options,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Model     string     `json:"model"`
	Message   ChatMessage `json:"message"`
	Done      bool       `json:"done"`
	Context   []int      `json:"context,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	
	// Usage statistics
	Usage *UsageStats `json:"usage,omitempty"`
}

// ChatStreamResponse represents a streaming chat response chunk
type ChatStreamResponse struct {
	Model     string     `json:"model"`
	Message   ChatMessage `json:"message"`
	Done      bool       `json:"done"`
	Context   []int      `json:"context,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	
	// Error information for streaming
	Error string `json:"error,omitempty"`
}

// ModelInfo represents information about an LLM model
type ModelInfo struct {
	Name      string            `json:"name"`
	ModifiedAt time.Time        `json:"modified_at"`
	Size      int64            `json:"size"`
	Digest    string           `json:"digest"`
	Details   *ModelDetails    `json:"details,omitempty"`
	Tags      []string         `json:"tags,omitempty"`
	Status    ModelStatus      `json:"status"`
	Metadata  map[string]any   `json:"metadata,omitempty"`
}

// ModelDetails contains detailed model information
type ModelDetails struct {
	Format    string   `json:"format"`
	Family    string   `json:"family"`
	Families  []string `json:"families"`
	Parameter string   `json:"parameter"`
	Quantization string `json:"quantization"`
}

// ModelStatus represents the status of a model
type ModelStatus string

const (
	ModelStatusAvailable   ModelStatus = "available"
	ModelStatusDownloading ModelStatus = "downloading"
	ModelStatusError       ModelStatus = "error"
	ModelStatusNotFound    ModelStatus = "not_found"
)

// UsageStats represents token usage statistics
type UsageStats struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OllamaProvider represents configuration for Ollama LLM provider
type OllamaProvider struct {
	BaseURL    string            `json:"base_url"`
	Timeout    time.Duration     `json:"timeout"`
	Headers    map[string]string `json:"headers,omitempty"`
	MaxRetries int              `json:"max_retries"`
}

// Repository defines data persistence for AIOps
type Repository interface {
	// Chat history operations
	SaveChatSession(ctx context.Context, session *ChatSession) error
	GetChatSession(ctx context.Context, sessionID string) (*ChatSession, error)
	ListChatSessions(ctx context.Context, workspaceID string, limit, offset int) ([]*ChatSession, error)
	DeleteChatSession(ctx context.Context, sessionID string) error
	
	// Model tracking
	TrackModelUsage(ctx context.Context, usage *ModelUsage) error
	GetModelUsageStats(ctx context.Context, workspaceID, modelName string, from, to time.Time) ([]*ModelUsage, error)
}

// ChatSession represents a chat conversation session
type ChatSession struct {
	ID          string        `json:"id"`
	WorkspaceID string        `json:"workspace_id"`
	UserID      string        `json:"user_id"`
	Title       string        `json:"title"`
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Context     []int         `json:"context,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ModelUsage represents model usage tracking
type ModelUsage struct {
	ID              string    `json:"id"`
	WorkspaceID     string    `json:"workspace_id"`
	UserID          string    `json:"user_id"`
	SessionID       string    `json:"session_id"`
	ModelName       string    `json:"model_name"`
	PromptTokens    int       `json:"prompt_tokens"`
	CompletionTokens int      `json:"completion_tokens"`
	TotalTokens     int       `json:"total_tokens"`
	RequestDuration time.Duration `json:"request_duration"`
	Timestamp       time.Time `json:"timestamp"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}