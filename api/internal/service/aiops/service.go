package aiops

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/aiops"
)

// Service implements the AIOps service interface
type Service struct {
	llmService aiops.LLMService
	repository aiops.Repository
	logger     *slog.Logger
	startTime  time.Time
}

// NewService creates a new AIOps service
func NewService(llmService aiops.LLMService, repository aiops.Repository, logger *slog.Logger) aiops.Service {
	return &Service{
		llmService: llmService,
		repository: repository,
		logger:     logger,
		startTime:  time.Now(),
	}
}

// CreateChatSession creates a new chat session
func (s *Service) CreateChatSession(ctx context.Context, workspaceID, userID, title, model string) (*aiops.ChatSession, error) {
	session := &aiops.ChatSession{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Title:       title,
		Model:       model,
		Messages:    []aiops.ChatMessage{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]any),
	}
	
	err := s.repository.SaveChatSession(ctx, session)
	if err != nil {
		s.logger.Error("Failed to save chat session", "error", err, "session_id", session.ID)
		return nil, fmt.Errorf("failed to save chat session: %w", err)
	}
	
	s.logger.Info("Chat session created", 
		"session_id", session.ID, 
		"workspace_id", workspaceID,
		"user_id", userID,
		"model", model)
	
	return session, nil
}

// SendMessage sends a message and gets a response from the LLM
func (s *Service) SendMessage(ctx context.Context, sessionID string, message aiops.ChatMessage) (*aiops.ChatResponse, error) {
	// Get the chat session
	session, err := s.repository.GetChatSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to get chat session", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}
	
	// Prepare the chat request
	messages := append(session.Messages, message)
	chatRequest := &aiops.ChatRequest{
		Model:    session.Model,
		Messages: messages,
		Stream:   false,
	}
	
	// Get response from LLM
	startTime := time.Now()
	response, err := s.llmService.Chat(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to get LLM response", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}
	duration := time.Since(startTime)
	
	// Update session with new messages
	session.Messages = append(messages, response.Message)
	session.UpdatedAt = time.Now()
	if response.Context != nil {
		session.Context = response.Context
	}
	
	// Save updated session
	err = s.repository.SaveChatSession(ctx, session)
	if err != nil {
		s.logger.Error("Failed to update chat session", "error", err, "session_id", sessionID)
		// Don't return error here as we got the response successfully
	}
	
	// Track usage if available
	if response.Usage != nil {
		usage := &aiops.ModelUsage{
			ID:               uuid.New().String(),
			WorkspaceID:      session.WorkspaceID,
			UserID:           session.UserID,
			SessionID:        sessionID,
			ModelName:        session.Model,
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
			RequestDuration:  duration,
			Timestamp:        time.Now(),
			Metadata:         make(map[string]any),
		}
		
		err = s.repository.TrackModelUsage(ctx, usage)
		if err != nil {
			s.logger.Error("Failed to track model usage", "error", err, "session_id", sessionID)
			// Don't return error here as the main operation succeeded
		}
	}
	
	s.logger.Info("Message processed",
		"session_id", sessionID,
		"model", session.Model,
		"duration", duration,
		"prompt_tokens", response.Usage.PromptTokens,
		"completion_tokens", response.Usage.CompletionTokens)
	
	return response, nil
}

// StreamMessage sends a message and streams the response
func (s *Service) StreamMessage(ctx context.Context, sessionID string, message aiops.ChatMessage) (<-chan *aiops.ChatStreamResponse, error) {
	// Get the chat session
	session, err := s.repository.GetChatSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to get chat session", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}
	
	// Prepare the chat request
	messages := append(session.Messages, message)
	chatRequest := &aiops.ChatRequest{
		Model:    session.Model,
		Messages: messages,
		Stream:   true,
	}
	
	// Get streaming response from LLM
	responseChan, err := s.llmService.StreamChat(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to start streaming chat", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("failed to start streaming chat: %w", err)
	}
	
	s.logger.Info("Started streaming chat", "session_id", sessionID, "model", session.Model)
	return responseChan, nil
}

// GetChatSession retrieves a chat session by ID
func (s *Service) GetChatSession(ctx context.Context, sessionID string) (*aiops.ChatSession, error) {
	session, err := s.repository.GetChatSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to get chat session", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}
	return session, nil
}

// ListChatSessions lists chat sessions for a workspace
func (s *Service) ListChatSessions(ctx context.Context, workspaceID string, limit, offset int) ([]*aiops.ChatSession, error) {
	sessions, err := s.repository.ListChatSessions(ctx, workspaceID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list chat sessions", "error", err, "workspace_id", workspaceID)
		return nil, fmt.Errorf("failed to list chat sessions: %w", err)
	}
	return sessions, nil
}

// DeleteChatSession deletes a chat session
func (s *Service) DeleteChatSession(ctx context.Context, sessionID string) error {
	err := s.repository.DeleteChatSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to delete chat session", "error", err, "session_id", sessionID)
		return fmt.Errorf("failed to delete chat session: %w", err)
	}
	
	s.logger.Info("Chat session deleted", "session_id", sessionID)
	return nil
}

// ListAvailableModels lists available LLM models
func (s *Service) ListAvailableModels(ctx context.Context) ([]*aiops.ModelInfo, error) {
	models, err := s.llmService.ListModels(ctx)
	if err != nil {
		s.logger.Error("Failed to list models", "error", err)
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	return models, nil
}

// EnsureModelAvailable ensures a model is available, pulling it if necessary
func (s *Service) EnsureModelAvailable(ctx context.Context, modelName string) error {
	// Check if model is already available
	_, err := s.llmService.GetModelInfo(ctx, modelName)
	if err == nil {
		// Model is already available
		return nil
	}
	
	// Model not found, try to pull it
	s.logger.Info("Pulling model", "model", modelName)
	err = s.llmService.PullModel(ctx, modelName)
	if err != nil {
		s.logger.Error("Failed to pull model", "error", err, "model", modelName)
		return fmt.Errorf("failed to pull model %s: %w", modelName, err)
	}
	
	s.logger.Info("Model pulled successfully", "model", modelName)
	return nil
}

// GetModelInfo gets information about a specific model
func (s *Service) GetModelInfo(ctx context.Context, modelName string) (*aiops.ModelInfo, error) {
	info, err := s.llmService.GetModelInfo(ctx, modelName)
	if err != nil {
		s.logger.Error("Failed to get model info", "error", err, "model", modelName)
		return nil, fmt.Errorf("failed to get model info: %w", err)
	}
	return info, nil
}

// GetUsageStats gets usage statistics for a workspace
func (s *Service) GetUsageStats(ctx context.Context, workspaceID string, from, to time.Time) (*aiops.UsageReport, error) {
	usage, err := s.repository.GetModelUsageStats(ctx, workspaceID, "", from, to)
	if err != nil {
		s.logger.Error("Failed to get usage stats", "error", err, "workspace_id", workspaceID)
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}
	
	// Aggregate usage data
	report := &aiops.UsageReport{
		WorkspaceID: workspaceID,
		Period: aiops.Period{
			From: from,
			To:   to,
		},
		ModelBreakdown: make(map[string]int),
		DailyUsage:     []aiops.DailyUsage{},
		TopUsers:       []aiops.UserUsage{},
	}
	
	totalTokens := 0
	totalMessages := len(usage)
	
	for _, u := range usage {
		totalTokens += u.TotalTokens
		report.ModelBreakdown[u.ModelName] += u.TotalTokens
	}
	
	report.TotalTokens = totalTokens
	report.TotalMessages = totalMessages
	
	return report, nil
}

// GetModelMetrics gets metrics for a specific model
func (s *Service) GetModelMetrics(ctx context.Context, modelName string, from, to time.Time) (*aiops.ModelMetrics, error) {
	// Implementation would analyze usage data for the specific model
	metrics := &aiops.ModelMetrics{
		ModelName: modelName,
		Period: aiops.Period{
			From: from,
			To:   to,
		},
		TotalRequests:   0,
		TotalTokens:     0,
		AverageLatency:  0,
		ErrorRate:       0,
		ThroughputRPS:   0,
		SuccessRate:     100.0,
	}
	
	return metrics, nil
}

// HealthCheck performs a health check of all AIOps services
func (s *Service) HealthCheck(ctx context.Context) *aiops.HealthStatus {
	services := make(map[string]aiops.ServiceHealth)
	
	// Check LLM service health
	llmHealthy := s.llmService.IsHealthy(ctx)
	llmStatus := aiops.StatusHealthy
	if !llmHealthy {
		llmStatus = aiops.StatusUnhealthy
	}
	
	services["llm"] = aiops.ServiceHealth{
		Status:      llmStatus,
		LastCheck:   time.Now(),
		ResponseTime: 0, // Would measure actual response time
	}
	
	// Determine overall status
	overallStatus := aiops.StatusHealthy
	for _, service := range services {
		if service.Status == aiops.StatusUnhealthy {
			overallStatus = aiops.StatusDegraded
			break
		}
	}
	
	return &aiops.HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Services:  services,
		Version:   "0.1.0",
		Uptime:    time.Since(s.startTime),
	}
}