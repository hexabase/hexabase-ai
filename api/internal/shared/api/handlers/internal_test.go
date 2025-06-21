//go:build unit

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	logsDomain "github.com/hexabase/hexabase-ai/api/internal/logs/domain"
	"github.com/hexabase/hexabase-ai/api/internal/shared/api/handlers"
	workspaceDomain "github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorkspaceService for internal handler tests.
type MockWorkspaceService struct {
	mock.Mock
}

// Implement only the methods we're testing
func (m *MockWorkspaceService) GetNodes(ctx context.Context, workspaceID string) ([]workspaceDomain.Node, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]workspaceDomain.Node), args.Error(1)
}

func (m *MockWorkspaceService) ScaleDeployment(ctx context.Context, workspaceID, deploymentName string, replicas int) error {
	args := m.Called(ctx, workspaceID, deploymentName, replicas)
	return args.Error(0)
}

// Implement other interface methods as stubs (not used in our tests)
func (m *MockWorkspaceService) CreateWorkspace(ctx context.Context, req *workspaceDomain.CreateWorkspaceRequest) (*workspaceDomain.Workspace, *workspaceDomain.Task, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*workspaceDomain.Workspace), args.Get(1).(*workspaceDomain.Task), args.Error(2)
}

func (m *MockWorkspaceService) GetWorkspace(ctx context.Context, workspaceID string) (*workspaceDomain.Workspace, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).(*workspaceDomain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) ListWorkspaces(ctx context.Context, filter workspaceDomain.WorkspaceFilter) (*workspaceDomain.WorkspaceList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*workspaceDomain.WorkspaceList), args.Error(1)
}

func (m *MockWorkspaceService) UpdateWorkspace(ctx context.Context, workspaceID string, req *workspaceDomain.UpdateWorkspaceRequest) (*workspaceDomain.Workspace, error) {
	args := m.Called(ctx, workspaceID, req)
	return args.Get(0).(*workspaceDomain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) DeleteWorkspace(ctx context.Context, workspaceID string) (*workspaceDomain.Task, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).(*workspaceDomain.Task), args.Error(1)
}

func (m *MockWorkspaceService) GetWorkspaceStatus(ctx context.Context, workspaceID string) (*workspaceDomain.WorkspaceStatus, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).(*workspaceDomain.WorkspaceStatus), args.Error(1)
}

func (m *MockWorkspaceService) GetKubeconfig(ctx context.Context, workspaceID string) (string, error) {
	args := m.Called(ctx, workspaceID)
	return args.String(0), args.Error(1)
}

func (m *MockWorkspaceService) ExecuteOperation(ctx context.Context, workspaceID string, req *workspaceDomain.WorkspaceOperationRequest) (*workspaceDomain.Task, error) {
	args := m.Called(ctx, workspaceID, req)
	return args.Get(0).(*workspaceDomain.Task), args.Error(1)
}

func (m *MockWorkspaceService) GetTask(ctx context.Context, taskID string) (*workspaceDomain.Task, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).(*workspaceDomain.Task), args.Error(1)
}

func (m *MockWorkspaceService) ListTasks(ctx context.Context, workspaceID string) ([]*workspaceDomain.Task, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]*workspaceDomain.Task), args.Error(1)
}

func (m *MockWorkspaceService) ProcessProvisioningTask(ctx context.Context, taskID string) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockWorkspaceService) ProcessDeletionTask(ctx context.Context, taskID string) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockWorkspaceService) ValidateWorkspaceAccess(ctx context.Context, userID, workspaceID string) error {
	args := m.Called(ctx, userID, workspaceID)
	return args.Error(0)
}

func (m *MockWorkspaceService) GetResourceUsage(ctx context.Context, workspaceID string) (*workspaceDomain.ResourceUsage, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).(*workspaceDomain.ResourceUsage), args.Error(1)
}

func (m *MockWorkspaceService) AddWorkspaceMember(ctx context.Context, workspaceID string, req *workspaceDomain.AddMemberRequest) (*workspaceDomain.WorkspaceMember, error) {
	args := m.Called(ctx, workspaceID, req)
	return args.Get(0).(*workspaceDomain.WorkspaceMember), args.Error(1)
}

func (m *MockWorkspaceService) ListWorkspaceMembers(ctx context.Context, workspaceID string) ([]*workspaceDomain.WorkspaceMember, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]*workspaceDomain.WorkspaceMember), args.Error(1)
}

func (m *MockWorkspaceService) RemoveWorkspaceMember(ctx context.Context, workspaceID, userID string) error {
	args := m.Called(ctx, workspaceID, userID)
	return args.Error(0)
}

// MockLogService for internal handler tests.
type MockLogService struct {
	mock.Mock
}

func (m *MockLogService) QueryLogs(ctx context.Context, query logsDomain.LogQuery) ([]logsDomain.LogEntry, error) {
	args := m.Called(ctx, query)
	if args.Get(0) != nil {
		return args.Get(0).([]logsDomain.LogEntry), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestInternalHandler_GetNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return a list of nodes for a given workspace", func(t *testing.T) {
		// Arrange
		mockWorkspaceSvc := new(MockWorkspaceService)
		
		// For this test, we don't need a real logger.
		logger := slog.Default()
		internalHandler := handlers.NewInternalHandler(
			mockWorkspaceSvc, nil, nil, nil, nil, nil, nil, nil, nil, logger,
		)

		workspaceID := "ws-12345"
		expectedNodes := []workspaceDomain.Node{
			{Name: "node-1", Status: "Ready", CPU: "500m", Memory: "2Gi", Pods: 5},
			{Name: "node-2", Status: "Ready", CPU: "500m", Memory: "2Gi", Pods: 3},
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

		var responseBody []workspaceDomain.Node
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
		logger := slog.Default()
		internalHandler := handlers.NewInternalHandler(
			mockWorkspaceSvc, nil, nil, nil, nil, nil, nil, nil, nil, logger,
		)

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
		logger := slog.Default()
		internalHandler := handlers.NewInternalHandler(
			nil, nil, nil, nil, mockLogSvc, nil, nil, nil, nil, logger,
		)

		workspaceID := "ws-12345"
		now := time.Now()
		query := logsDomain.LogQuery{
			WorkspaceID: workspaceID,
			SearchTerm:  "error",
			StartTime:   now.Add(-1 * time.Hour),
			EndTime:     now,
			Limit:       100,
		}
		
		expectedLogs := []logsDomain.LogEntry{
			{Timestamp: now, Level: "error", Message: "Something went wrong"},
		}

		// Use mock.AnythingOfType because comparing time objects in tests can be flaky.
		mockLogSvc.On("QueryLogs", mock.Anything, mock.AnythingOfType("domain.LogQuery")).Return(expectedLogs, nil)

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

		var responseBody []logsDomain.LogEntry
		err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Len(t, responseBody, 1)
		assert.Equal(t, "Something went wrong", responseBody[0].Message)

		mockLogSvc.AssertExpectations(t)
	})
} 