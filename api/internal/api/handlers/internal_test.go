//go:build unit

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/api/handlers"
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