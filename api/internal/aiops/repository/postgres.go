package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/aiops/domain"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// PostgresRepository implements the AIOps repository interface using PostgreSQL
type PostgresRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *gorm.DB, logger *slog.Logger) domain.Repository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

// ChatSessionModel represents the database model for chat sessions
type ChatSessionModel struct {
	ID          string         `gorm:"primaryKey"`
	WorkspaceID string         `gorm:"index"`
	UserID      string         `gorm:"index"`
	Title       string
	Model       string
	Messages    json.RawMessage
	Context     pq.Int64Array  `gorm:"type:integer[]"`
	Metadata    json.RawMessage
	CreatedAt   time.Time      `gorm:"index"`
	UpdatedAt   time.Time      `gorm:"index"`
}

// TableName specifies the table name
func (ChatSessionModel) TableName() string {
	return "chat_sessions"
}

// ModelUsageModel represents the database model for model usage
type ModelUsageModel struct {
	ID                 string          `gorm:"primaryKey"`
	WorkspaceID        string          `gorm:"index"`
	UserID             string          `gorm:"index"`
	SessionID          string          `gorm:"index"`
	ModelName          string          `gorm:"index"`
	PromptTokens       int
	CompletionTokens   int
	TotalTokens        int
	RequestDurationMs  int64
	Timestamp          time.Time       `gorm:"index"`
	Metadata           json.RawMessage
}

// TableName specifies the table name
func (ModelUsageModel) TableName() string {
	return "model_usage"
}

// SaveChatSession saves or updates a chat session
func (r *PostgresRepository) SaveChatSession(ctx context.Context, session *domain.ChatSession) error {
	// Convert messages to JSON
	messagesJSON, err := json.Marshal(session.Messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}
	
	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(session.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	// Convert context to int64 array
	var contextArray pq.Int64Array
	if session.Context != nil {
		contextArray = make(pq.Int64Array, len(session.Context))
		for i, v := range session.Context {
			contextArray[i] = int64(v)
		}
	}
	
	model := ChatSessionModel{
		ID:          session.ID,
		WorkspaceID: session.WorkspaceID,
		UserID:      session.UserID,
		Title:       session.Title,
		Model:       session.Model,
		Messages:    messagesJSON,
		Context:     contextArray,
		Metadata:    metadataJSON,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
	}
	
	// Check if session exists
	var exists bool
	err = r.db.WithContext(ctx).
		Model(&ChatSessionModel{}).
		Where("id = ?", session.ID).
		Select("1").
		Limit(1).
		Scan(&exists).Error
	
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	
	if exists {
		// Update existing session
		result := r.db.WithContext(ctx).
			Model(&ChatSessionModel{}).
			Where("id = ?", session.ID).
			Updates(map[string]interface{}{
				"title":      model.Title,
				"messages":   model.Messages,
				"context":    model.Context,
				"metadata":   model.Metadata,
				"updated_at": model.UpdatedAt,
			})
		if result.Error != nil {
			return fmt.Errorf("failed to update chat session: %w", result.Error)
		}
	} else {
		// Create new session
		result := r.db.WithContext(ctx).Create(&model)
		if result.Error != nil {
			return fmt.Errorf("failed to create chat session: %w", result.Error)
		}
	}
	
	return nil
}

// GetChatSession retrieves a chat session by ID
func (r *PostgresRepository) GetChatSession(ctx context.Context, sessionID string) (*domain.ChatSession, error) {
	var model ChatSessionModel
	result := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("chat session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get chat session: %w", result.Error)
	}
	
	// Convert from model to domain
	session := &domain.ChatSession{
		ID:          model.ID,
		WorkspaceID: model.WorkspaceID,
		UserID:      model.UserID,
		Title:       model.Title,
		Model:       model.Model,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
	
	// Unmarshal messages
	if model.Messages != nil {
		err := json.Unmarshal(model.Messages, &session.Messages)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
		}
	}
	
	// Convert context
	if model.Context != nil {
		session.Context = make([]int, len(model.Context))
		for i, v := range model.Context {
			session.Context[i] = int(v)
		}
	}
	
	// Unmarshal metadata
	if model.Metadata != nil {
		err := json.Unmarshal(model.Metadata, &session.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	
	return session, nil
}

// ListChatSessions lists chat sessions for a workspace
func (r *PostgresRepository) ListChatSessions(ctx context.Context, workspaceID string, limit, offset int) ([]*domain.ChatSession, error) {
	var models []ChatSessionModel
	result := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list chat sessions: %w", result.Error)
	}
	
	sessions := make([]*domain.ChatSession, len(models))
	for i, model := range models {
		session := &domain.ChatSession{
			ID:          model.ID,
			WorkspaceID: model.WorkspaceID,
			UserID:      model.UserID,
			Title:       model.Title,
			Model:       model.Model,
			CreatedAt:   model.CreatedAt,
			UpdatedAt:   model.UpdatedAt,
		}
		
		// Unmarshal messages
		if model.Messages != nil {
			err := json.Unmarshal(model.Messages, &session.Messages)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal messages for session %s: %w", model.ID, err)
			}
		}
		
		// Convert context
		if model.Context != nil {
			session.Context = make([]int, len(model.Context))
			for j, v := range model.Context {
				session.Context[j] = int(v)
			}
		}
		
		// Unmarshal metadata
		if model.Metadata != nil {
			err := json.Unmarshal(model.Metadata, &session.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata for session %s: %w", model.ID, err)
			}
		}
		
		sessions[i] = session
	}
	
	return sessions, nil
}

// DeleteChatSession deletes a chat session
func (r *PostgresRepository) DeleteChatSession(ctx context.Context, sessionID string) error {
	result := r.db.WithContext(ctx).Delete(&ChatSessionModel{}, "id = ?", sessionID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete chat session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("chat session not found: %s", sessionID)
	}
	return nil
}

// TrackModelUsage tracks model usage statistics
func (r *PostgresRepository) TrackModelUsage(ctx context.Context, usage *domain.ModelUsage) error {
	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(usage.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	model := ModelUsageModel{
		ID:                usage.ID,
		WorkspaceID:       usage.WorkspaceID,
		UserID:            usage.UserID,
		SessionID:         usage.SessionID,
		ModelName:         usage.ModelName,
		PromptTokens:      usage.PromptTokens,
		CompletionTokens:  usage.CompletionTokens,
		TotalTokens:       usage.TotalTokens,
		RequestDurationMs: usage.RequestDuration.Milliseconds(),
		Timestamp:         usage.Timestamp,
		Metadata:          metadataJSON,
	}
	
	result := r.db.WithContext(ctx).Create(&model)
	if result.Error != nil {
		return fmt.Errorf("failed to track model usage: %w", result.Error)
	}
	
	return nil
}

// GetModelUsageStats retrieves model usage statistics
func (r *PostgresRepository) GetModelUsageStats(ctx context.Context, workspaceID, modelName string, from, to time.Time) ([]*domain.ModelUsage, error) {
	var models []ModelUsageModel
	
	query := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)
	
	if modelName != "" {
		query = query.Where("model_name = ?", modelName)
	}
	
	query = query.Where("timestamp BETWEEN ? AND ?", from, to).Order("timestamp DESC")
	
	result := query.Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get model usage stats: %w", result.Error)
	}
	
	stats := make([]*domain.ModelUsage, len(models))
	for i, model := range models {
		usage := &domain.ModelUsage{
			ID:               model.ID,
			WorkspaceID:      model.WorkspaceID,
			UserID:           model.UserID,
			SessionID:        model.SessionID,
			ModelName:        model.ModelName,
			PromptTokens:     model.PromptTokens,
			CompletionTokens: model.CompletionTokens,
			TotalTokens:      model.TotalTokens,
			RequestDuration:  time.Duration(model.RequestDurationMs) * time.Millisecond,
			Timestamp:        model.Timestamp,
		}
		
		// Unmarshal metadata
		if model.Metadata != nil {
			err := json.Unmarshal(model.Metadata, &usage.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata for usage %s: %w", model.ID, err)
			}
		}
		
		stats[i] = usage
	}
	
	return stats, nil
}

