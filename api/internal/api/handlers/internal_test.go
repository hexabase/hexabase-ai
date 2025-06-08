//go:build unit

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/api/handlers"
	"github.com/hexabase/hexabase-ai/api/internal/domain/logs"
	"github.com/hexabase/hexabase-ai/api/internal/domain/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorkspaceService for internal handler tests.
type MockWorkspaceService struct {
	mock.Mock
}

// We only need to mock the method we're using in the handler.
func (m *MockWorkspaceService) GetNodes(ctx context.Context, workspaceID string) ([]workspace.Node, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]workspace.Node), args.Error(1)
}

// We need to add the new method to our mock service.
func (m *MockWorkspaceService) ScaleDeployment(ctx context.Context, workspaceID, deploymentName string, replicas int) error {
	args := m.Called(ctx, workspaceID, deploymentName, replicas)
	return args.Error(0)
}

// MockLogService for internal handler tests.
type MockLogService struct {
	mock.Mock
}

func (m *MockLogService) QueryLogs(ctx context.Context, query logs.LogQuery) ([]logs.LogEntry, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]logs.LogEntry), args.Error(1)
}

func TestInternalHandler_GetNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return a list of nodes for a given workspace", func(t *testing.T) {
		// Arrange
		mockWorkspaceSvc := new(MockWorkspaceService)
		
		// For this test, we don't need a real logger.
		internalHandler := handlers.NewInternalHandler(mockWorkspaceSvc, nil)

		workspaceID := "ws-12345"
		expectedNodes := []workspace.Node{
			{Name: "node-1", Status: "Ready", CPU: "500m", Memory: "2Gi"},
			{Name: "node-2", Status: "Ready", CPU: "500m", Memory: "2Gi"},
		}

		mockWorkspaceSvc.On("GetNodes", mock.Anything, workspaceID).Return(expectedNodes, nil)

		router := gin.Default()
		router.GET("/internal/v1/workspaces/:workspaceId/nodes", internalHandler.GetNodes)

		req, _ := http.NewRequest(http.MethodGet, "/internal/v1/workspaces/"+workspaceID+"/nodes", nil)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody []workspace.Node
		err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Equal(t, expectedNodes, responseBody)

		mockWorkspaceSvc.AssertExpectations(t)
	})
}

func TestInternalHandler_ScaleDeployment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should scale a deployment successfully", func(t *testing.T) {
		// Arrange
		mockWorkspaceSvc := new(MockWorkspaceService)
		internalHandler := handlers.NewInternalHandler(mockWorkspaceSvc, nil)

		workspaceID := "ws-12345"
		deploymentName := "my-web-api"
		
		scaleRequest := struct {
			Replicas int `json:"replicas"`
		}{
			Replicas: 3,
		}

		mockWorkspaceSvc.On("ScaleDeployment", mock.Anything, workspaceID, deploymentName, scaleRequest.Replicas).Return(nil)

		router := gin.Default()
		router.POST("/internal/v1/workspaces/:workspaceId/deployments/:deploymentName/scale", internalHandler.ScaleDeployment)

		body, _ := json.Marshal(scaleRequest)
		req, _ := http.NewRequest(http.MethodPost, "/internal/v1/workspaces/"+workspaceID+"/deployments/"+deploymentName+"/scale", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		mockWorkspaceSvc.AssertExpectations(t)
	})
}

func TestInternalHandler_QueryLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return logs based on a query", func(t *testing.T) {
		// Arrange
		mockLogSvc := new(MockLogService)
		// We're not testing the workspace service here, so it can be nil.
		internalHandler := handlers.NewInternalHandler(nil, mockLogSvc, nil)

		workspaceID := "ws-12345"
		now := time.Now()
		query := logs.LogQuery{
			WorkspaceID: workspaceID,
			SearchTerm:  "error",
			StartTime:   now.Add(-1 * time.Hour),
			EndTime:     now,
			Limit:       100,
		}
		
		expectedLogs := []logs.LogEntry{
			{Timestamp: now, Level: "error", Message: "Something went wrong"},
		}

		// Use mock.AnythingOfType because comparing time objects in tests can be flaky.
		mockLogSvc.On("QueryLogs", mock.Anything, mock.AnythingOfType("logs.LogQuery")).Return(expectedLogs, nil)

		router := gin.Default()
		router.POST("/internal/v1/logs/query", internalHandler.QueryLogs)

		body, _ := json.Marshal(query)
		req, _ := http.NewRequest(http.MethodPost, "/internal/v1/logs/query", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody []logs.LogEntry
		err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Len(t, responseBody, 1)
		assert.Equal(t, "Something went wrong", responseBody[0].Message)

		mockLogSvc.AssertExpectations(t)
	})
} 