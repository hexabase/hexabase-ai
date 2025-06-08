package aiops

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/aiops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
)

// Mock implementations for testing

type MockLLMService struct {
	mock.Mock
}

func (m *MockLLMService) Chat(ctx context.Context, req *aiops.ChatRequest) (*aiops.ChatResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aiops.ChatResponse), args.Error(1)
}

func (m *MockLLMService) StreamChat(ctx context.Context, req *aiops.ChatRequest) (<-chan *aiops.ChatStreamResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan *aiops.ChatStreamResponse), args.Error(1)
}

func (m *MockLLMService) ListModels(ctx context.Context) ([]*aiops.ModelInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aiops.ModelInfo), args.Error(1)
}

func (m *MockLLMService) PullModel(ctx context.Context, modelName string) error {
	args := m.Called(ctx, modelName)
	return args.Error(0)
}

func (m *MockLLMService) DeleteModel(ctx context.Context, modelName string) error {
	args := m.Called(ctx, modelName)
	return args.Error(0)
}

func (m *MockLLMService) IsHealthy(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockLLMService) GetModelInfo(ctx context.Context, modelName string) (*aiops.ModelInfo, error) {
	args := m.Called(ctx, modelName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aiops.ModelInfo), args.Error(1)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveChatSession(ctx context.Context, session *aiops.ChatSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) GetChatSession(ctx context.Context, sessionID string) (*aiops.ChatSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aiops.ChatSession), args.Error(1)
}

func (m *MockRepository) ListChatSessions(ctx context.Context, workspaceID string, limit, offset int) ([]*aiops.ChatSession, error) {
	args := m.Called(ctx, workspaceID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aiops.ChatSession), args.Error(1)
}

func (m *MockRepository) DeleteChatSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockRepository) TrackModelUsage(ctx context.Context, usage *aiops.ModelUsage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *MockRepository) GetModelUsageStats(ctx context.Context, workspaceID, modelName string, from, to time.Time) ([]*aiops.ModelUsage, error) {
	args := m.Called(ctx, workspaceID, modelName, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aiops.ModelUsage), args.Error(1)
}

// Test helper functions

func createTestService() (*Service, *MockLLMService, *MockRepository) {
	mockLLM := &MockLLMService{}
	mockRepo := &MockRepository{}
	logger := slog.Default()
	
	service := NewService(mockLLM, mockRepo, logger).(*Service)
	return service, mockLLM, mockRepo
}

func createTestChatSession() *aiops.ChatSession {
	return &aiops.ChatSession{
		ID:          uuid.New().String(),
		WorkspaceID: "workspace-123",
		UserID:      "user-123",
		Title:       "Test Chat",
		Model:       "llama2:7b",
		Messages:    []aiops.ChatMessage{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func createTestChatMessage() aiops.ChatMessage {
	return aiops.ChatMessage{
		Role:    "user",
		Content: "Hello, how can you help me with Kubernetes troubleshooting?",
	}
}

// Tests following TDD methodology

func TestCreateChatSession(t *testing.T) {
	t.Run("successful chat session creation", func(t *testing.T) {
		service, _, mockRepo := createTestService()
		ctx := context.Background()
		
		workspaceID := "workspace-123"
		userID := "user-123"
		title := "Kubernetes Troubleshooting"
		model := "llama2:7b"
		
		mockRepo.On("SaveChatSession", ctx, mock.AnythingOfType("*aiops.ChatSession")).Return(nil)
		
		// Act
		session, err := service.CreateChatSession(ctx, workspaceID, userID, title, model)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, workspaceID, session.WorkspaceID)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, title, session.Title)
		assert.Equal(t, model, session.Model)
		assert.NotEmpty(t, session.ID)
		assert.Empty(t, session.Messages)
		
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("repository save failure", func(t *testing.T) {
		service, _, mockRepo := createTestService()
		ctx := context.Background()
		
		mockRepo.On("SaveChatSession", ctx, mock.AnythingOfType("*aiops.ChatSession")).Return(errors.New("database error"))
		
		// Act
		session, err := service.CreateChatSession(ctx, "workspace-123", "user-123", "Test", "llama2:7b")
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "failed to save chat session")
	})
}

func TestSendMessage(t *testing.T) {
	t.Run("successful message send", func(t *testing.T) {
		service, mockLLM, mockRepo := createTestService()
		ctx := context.Background()
		
		session := createTestChatSession()
		message := createTestChatMessage()
		sessionID := session.ID
		
		// Expected LLM response
		expectedResponse := &aiops.ChatResponse{
			Model: "llama2:7b",
			Message: aiops.ChatMessage{
				Role:    "assistant",
				Content: "I can help you troubleshoot Kubernetes issues. What specific problem are you experiencing?",
			},
			Done:      true,
			CreatedAt: time.Now(),
			Usage: &aiops.UsageStats{
				PromptTokens:     15,
				CompletionTokens: 25,
				TotalTokens:      40,
			},
		}
		
		// Mock expectations
		mockRepo.On("GetChatSession", ctx, sessionID).Return(session, nil)
		mockLLM.On("Chat", ctx, mock.AnythingOfType("*aiops.ChatRequest")).Return(expectedResponse, nil)
		mockRepo.On("SaveChatSession", ctx, mock.AnythingOfType("*aiops.ChatSession")).Return(nil)
		mockRepo.On("TrackModelUsage", ctx, mock.AnythingOfType("*aiops.ModelUsage")).Return(nil)
		
		// Act
		response, err := service.SendMessage(ctx, sessionID, message)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "llama2:7b", response.Model)
		assert.Equal(t, "assistant", response.Message.Role)
		assert.True(t, response.Done)
		assert.NotNil(t, response.Usage)
		
		mockRepo.AssertExpectations(t)
		mockLLM.AssertExpectations(t)
	})
	
	t.Run("session not found", func(t *testing.T) {
		service, _, mockRepo := createTestService()
		ctx := context.Background()
		
		sessionID := "non-existent-session"
		message := createTestChatMessage()
		
		mockRepo.On("GetChatSession", ctx, sessionID).Return(nil, errors.New("session not found"))
		
		// Act
		response, err := service.SendMessage(ctx, sessionID, message)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get chat session")
	})
	
	t.Run("LLM service failure", func(t *testing.T) {
		service, mockLLM, mockRepo := createTestService()
		ctx := context.Background()
		
		session := createTestChatSession()
		message := createTestChatMessage()
		sessionID := session.ID
		
		mockRepo.On("GetChatSession", ctx, sessionID).Return(session, nil)
		mockLLM.On("Chat", ctx, mock.AnythingOfType("*aiops.ChatRequest")).Return(nil, errors.New("LLM service unavailable"))
		
		// Act
		response, err := service.SendMessage(ctx, sessionID, message)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get LLM response")
	})
}

func TestListAvailableModels(t *testing.T) {
	t.Run("successful model listing", func(t *testing.T) {
		service, mockLLM, _ := createTestService()
		ctx := context.Background()
		
		expectedModels := []*aiops.ModelInfo{
			{
				Name:       "llama2:7b",
				ModifiedAt: time.Now(),
				Size:       3800000000,
				Status:     aiops.ModelStatusAvailable,
			},
			{
				Name:       "codellama:13b",
				ModifiedAt: time.Now(),
				Size:       7300000000,
				Status:     aiops.ModelStatusAvailable,
			},
		}
		
		mockLLM.On("ListModels", ctx).Return(expectedModels, nil)
		
		// Act
		models, err := service.ListAvailableModels(ctx)
		
		// Assert
		assert.NoError(t, err)
		assert.Len(t, models, 2)
		assert.Equal(t, "llama2:7b", models[0].Name)
		assert.Equal(t, "codellama:13b", models[1].Name)
		
		mockLLM.AssertExpectations(t)
	})
	
	t.Run("LLM service failure", func(t *testing.T) {
		service, mockLLM, _ := createTestService()
		ctx := context.Background()
		
		mockLLM.On("ListModels", ctx).Return(nil, errors.New("connection failed"))
		
		// Act
		models, err := service.ListAvailableModels(ctx)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, models)
		assert.Contains(t, err.Error(), "failed to list models")
	})
}

func TestEnsureModelAvailable(t *testing.T) {
	t.Run("model already available", func(t *testing.T) {
		service, mockLLM, _ := createTestService()
		ctx := context.Background()
		
		modelName := "llama2:7b"
		modelInfo := &aiops.ModelInfo{
			Name:   modelName,
			Status: aiops.ModelStatusAvailable,
		}
		
		mockLLM.On("GetModelInfo", ctx, modelName).Return(modelInfo, nil)
		
		// Act
		err := service.EnsureModelAvailable(ctx, modelName)
		
		// Assert
		assert.NoError(t, err)
		mockLLM.AssertExpectations(t)
	})
	
	t.Run("model needs to be pulled", func(t *testing.T) {
		service, mockLLM, _ := createTestService()
		ctx := context.Background()
		
		modelName := "llama2:7b"
		
		// First call returns not found, then after pull it's available
		mockLLM.On("GetModelInfo", ctx, modelName).Return(nil, errors.New("model not found")).Once()
		mockLLM.On("PullModel", ctx, modelName).Return(nil)
		
		// Act
		err := service.EnsureModelAvailable(ctx, modelName)
		
		// Assert
		assert.NoError(t, err)
		mockLLM.AssertExpectations(t)
	})
	
	t.Run("model pull failure", func(t *testing.T) {
		service, mockLLM, _ := createTestService()
		ctx := context.Background()
		
		modelName := "llama2:7b"
		
		mockLLM.On("GetModelInfo", ctx, modelName).Return(nil, errors.New("model not found"))
		mockLLM.On("PullModel", ctx, modelName).Return(errors.New("pull failed"))
		
		// Act
		err := service.EnsureModelAvailable(ctx, modelName)
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to pull model")
	})
}

func TestHealthCheck(t *testing.T) {
	t.Run("all services healthy", func(t *testing.T) {
		service, mockLLM, _ := createTestService()
		ctx := context.Background()
		
		mockLLM.On("IsHealthy", ctx).Return(true)
		
		// Act
		status := service.HealthCheck(ctx)
		
		// Assert
		assert.NotNil(t, status)
		assert.Equal(t, aiops.StatusHealthy, status.Status)
		assert.Contains(t, status.Services, "llm")
		assert.Equal(t, aiops.StatusHealthy, status.Services["llm"].Status)
		
		mockLLM.AssertExpectations(t)
	})
	
	t.Run("LLM service unhealthy", func(t *testing.T) {
		service, mockLLM, _ := createTestService()
		ctx := context.Background()
		
		mockLLM.On("IsHealthy", ctx).Return(false)
		
		// Act
		status := service.HealthCheck(ctx)
		
		// Assert
		assert.NotNil(t, status)
		assert.Equal(t, aiops.StatusDegraded, status.Status)
		assert.Contains(t, status.Services, "llm")
		assert.Equal(t, aiops.StatusUnhealthy, status.Services["llm"].Status)
	})
}

func TestGetUsageStats(t *testing.T) {
	t.Run("successful usage stats retrieval", func(t *testing.T) {
		service, _, mockRepo := createTestService()
		ctx := context.Background()
		
		workspaceID := "workspace-123"
		from := time.Now().AddDate(0, 0, -7)
		to := time.Now()
		
		mockUsage := []*aiops.ModelUsage{
			{
				ID:               "usage-1",
				WorkspaceID:      workspaceID,
				UserID:           "user-1",
				ModelName:        "llama2:7b",
				PromptTokens:     50,
				CompletionTokens: 75,
				TotalTokens:      125,
				Timestamp:        time.Now().AddDate(0, 0, -1),
			},
			{
				ID:               "usage-2",
				WorkspaceID:      workspaceID,
				UserID:           "user-2",
				ModelName:        "codellama:13b",
				PromptTokens:     100,
				CompletionTokens: 150,
				TotalTokens:      250,
				Timestamp:        time.Now(),
			},
		}
		
		mockRepo.On("GetModelUsageStats", ctx, workspaceID, "", from, to).Return(mockUsage, nil)
		
		// Act
		report, err := service.GetUsageStats(ctx, workspaceID, from, to)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, workspaceID, report.WorkspaceID)
		assert.Equal(t, 2, report.TotalMessages)
		assert.Equal(t, 375, report.TotalTokens)
		assert.Len(t, report.ModelBreakdown, 2)
		assert.Equal(t, 125, report.ModelBreakdown["llama2:7b"])
		assert.Equal(t, 250, report.ModelBreakdown["codellama:13b"])
		
		mockRepo.AssertExpectations(t)
	})
}