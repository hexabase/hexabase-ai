package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hexabase/hexabase-ai/api/internal/aiops/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAIOpsService is a mock implementation of AIOpsService
type MockAIOpsService struct {
	mock.Mock
}

func (m *MockAIOpsService) CreateChatSession(workspaceID, userID, model string) (*domain.ChatSession, error) {
	args := m.Called(workspaceID, userID, model)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ChatSession), args.Error(1)
}

func (m *MockAIOpsService) GetChatSession(sessionID string) (*domain.ChatSession, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ChatSession), args.Error(1)
}

func (m *MockAIOpsService) ListChatSessions(workspaceID string, limit, offset int) ([]*domain.ChatSession, error) {
	args := m.Called(workspaceID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ChatSession), args.Error(1)
}

func (m *MockAIOpsService) DeleteChatSession(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockAIOpsService) Chat(sessionID string, message string, context []int) (*domain.ChatResponse, error) {
	args := m.Called(sessionID, message, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ChatResponse), args.Error(1)
}

func (m *MockAIOpsService) StreamChat(sessionID string, message string, context []int) (<-chan *domain.ChatStreamResponse, error) {
	args := m.Called(sessionID, message, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan *domain.ChatStreamResponse), args.Error(1)
}

func (m *MockAIOpsService) GetAvailableModels() ([]*domain.ModelInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ModelInfo), args.Error(1)
}

func (m *MockAIOpsService) GetTokenUsage(workspaceID, model string, limit, offset int) ([]*domain.ModelUsage, error) {
	args := m.Called(workspaceID, model, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ModelUsage), args.Error(1)
}

func TestAIOpsHandler_CreateChatSession(t *testing.T) {
	t.Run("successful session creation", func(t *testing.T) {
		mockService := new(MockAIOpsService)
		handler := NewHandler(mockService, slog.Default())

		expectedSession := &domain.ChatSession{
			ID:          uuid.New().String(),
			WorkspaceID: "workspace-123",
			UserID:      "user-123",
			Model:       "llama2:7b",
		}

		mockService.On("CreateChatSession", "workspace-123", "user-123", "llama2:7b").
			Return(expectedSession, nil)

		reqBody := CreateChatSessionRequest{
			WorkspaceID: "workspace-123",
			UserID:      "user-123",
			Model:       "llama2:7b",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/aiops/sessions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.CreateChatSession(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response domain.ChatSession
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedSession.ID, response.ID)
		assert.Equal(t, expectedSession.WorkspaceID, response.WorkspaceID)

		mockService.AssertExpectations(t)
	})

	t.Run("missing required fields", func(t *testing.T) {
		mockService := new(MockAIOpsService)
		handler := NewHandler(mockService, slog.Default())

		reqBody := CreateChatSessionRequest{
			WorkspaceID: "",
			UserID:      "user-123",
			Model:       "llama2:7b",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/aiops/sessions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.CreateChatSession(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response.Error, "workspace_id is required")
	})
}

func TestAIOpsHandler_Chat(t *testing.T) {
	t.Run("successful chat", func(t *testing.T) {
		mockService := new(MockAIOpsService)
		handler := NewHandler(mockService, slog.Default())

		sessionID := uuid.New().String()
		expectedResponse := &domain.ChatResponse{
			Model: "llama2:7b",
			Message: domain.ChatMessage{
				Role:    "assistant",
				Content: "Hello! How can I help you today?",
			},
			Done: true,
			Usage: &domain.UsageStats{
				PromptTokens:     10,
				CompletionTokens: 8,
				TotalTokens:      18,
			},
		}

		mockService.On("Chat", sessionID, "Hello", []int(nil)).
			Return(expectedResponse, nil)

		reqBody := ChatRequest{
			Message: "Hello",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/aiops/sessions/"+sessionID+"/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		
		// Set up mux vars to simulate router behavior
		req = mux.SetURLVars(req, map[string]string{"sessionId": sessionID})

		rr := httptest.NewRecorder()
		handler.Chat(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response domain.ChatResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Message.Content, response.Message.Content)
		assert.Equal(t, expectedResponse.Usage.TotalTokens, response.Usage.TotalTokens)

		mockService.AssertExpectations(t)
	})

	t.Run("empty message", func(t *testing.T) {
		mockService := new(MockAIOpsService)
		handler := NewHandler(mockService, slog.Default())

		sessionID := uuid.New().String()
		reqBody := ChatRequest{
			Message: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/aiops/sessions/"+sessionID+"/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		
		// Set up mux vars to simulate router behavior
		req = mux.SetURLVars(req, map[string]string{"sessionId": sessionID})

		rr := httptest.NewRecorder()
		handler.Chat(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response.Error, "message is required")
	})
}

func TestAIOpsHandler_ListChatSessions(t *testing.T) {
	t.Run("successful listing", func(t *testing.T) {
		mockService := new(MockAIOpsService)
		handler := NewHandler(mockService, slog.Default())

		expectedSessions := []*domain.ChatSession{
			{
				ID:          uuid.New().String(),
				WorkspaceID: "workspace-123",
				Title:       "Chat 1",
			},
			{
				ID:          uuid.New().String(),
				WorkspaceID: "workspace-123",
				Title:       "Chat 2",
			},
		}

		mockService.On("ListChatSessions", "workspace-123", 10, 0).
			Return(expectedSessions, nil)

		req := httptest.NewRequest("GET", "/api/v1/aiops/sessions?workspace_id=workspace-123&limit=10&offset=0", nil)

		rr := httptest.NewRecorder()
		handler.ListChatSessions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response ListChatSessionsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Sessions, 2)
		assert.Equal(t, expectedSessions[0].ID, response.Sessions[0].ID)

		mockService.AssertExpectations(t)
	})

	t.Run("missing workspace_id", func(t *testing.T) {
		mockService := new(MockAIOpsService)
		handler := NewHandler(mockService, slog.Default())

		req := httptest.NewRequest("GET", "/api/v1/aiops/sessions", nil)

		rr := httptest.NewRecorder()
		handler.ListChatSessions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response.Error, "workspace_id is required")
	})
}

func TestAIOpsHandler_GetAvailableModels(t *testing.T) {
	t.Run("successful model listing", func(t *testing.T) {
		mockService := new(MockAIOpsService)
		handler := NewHandler(mockService, slog.Default())

		expectedModels := []*domain.ModelInfo{
			{
				Name:     "llama2:7b",
				Status:   domain.ModelStatusAvailable,
				Size:     3826793472,
			},
			{
				Name:     "codellama:13b",
				Status:   domain.ModelStatusAvailable,
				Size:     7365834752,
			},
		}

		mockService.On("GetAvailableModels").Return(expectedModels, nil)

		req := httptest.NewRequest("GET", "/api/v1/aiops/models", nil)

		rr := httptest.NewRecorder()
		handler.GetAvailableModels(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response GetModelsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Models, 2)
		assert.Equal(t, expectedModels[0].Name, response.Models[0].Name)

		mockService.AssertExpectations(t)
	})
}