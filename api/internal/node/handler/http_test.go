package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	node "github.com/hexabase/hexabase-ai/api/internal/node/domain"
)

// MockNodeService mocks the node service interface
type MockNodeService struct {
	mock.Mock
}

func (m *MockNodeService) GetAvailablePlans(ctx context.Context) ([]node.NodePlan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]node.NodePlan), args.Error(1)
}

func (m *MockNodeService) GetPlanDetails(ctx context.Context, planID string) (*node.NodePlan, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.NodePlan), args.Error(1)
}

func (m *MockNodeService) ProvisionDedicatedNode(ctx context.Context, workspaceID string, req node.ProvisionRequest) (*node.DedicatedNode, error) {
	args := m.Called(ctx, workspaceID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.DedicatedNode), args.Error(1)
}

func (m *MockNodeService) GetNode(ctx context.Context, nodeID string) (*node.DedicatedNode, error) {
	args := m.Called(ctx, nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.DedicatedNode), args.Error(1)
}

func (m *MockNodeService) ListNodes(ctx context.Context, workspaceID string) ([]node.DedicatedNode, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]node.DedicatedNode), args.Error(1)
}

func (m *MockNodeService) StartNode(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func (m *MockNodeService) StopNode(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func (m *MockNodeService) RebootNode(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func (m *MockNodeService) DeleteNode(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func (m *MockNodeService) GetWorkspaceResourceUsage(ctx context.Context, workspaceID string) (*node.WorkspaceResourceUsage, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.WorkspaceResourceUsage), args.Error(1)
}

func (m *MockNodeService) CanAllocateResources(ctx context.Context, workspaceID string, req node.ResourceRequest) (bool, error) {
	args := m.Called(ctx, workspaceID, req)
	return args.Bool(0), args.Error(1)
}

func (m *MockNodeService) GetNodeStatus(ctx context.Context, nodeID string) (*node.NodeStatusInfo, error) {
	args := m.Called(ctx, nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.NodeStatusInfo), args.Error(1)
}

func (m *MockNodeService) GetNodeMetrics(ctx context.Context, nodeID string) (*node.NodeMetrics, error) {
	args := m.Called(ctx, nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.NodeMetrics), args.Error(1)
}

func (m *MockNodeService) GetNodeEvents(ctx context.Context, nodeID string, limit int) ([]node.NodeEvent, error) {
	args := m.Called(ctx, nodeID, limit)
	return args.Get(0).([]node.NodeEvent), args.Error(1)
}

func (m *MockNodeService) GetNodeCosts(ctx context.Context, workspaceID string, period node.BillingPeriod) (*node.NodeCostReport, error) {
	args := m.Called(ctx, workspaceID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.NodeCostReport), args.Error(1)
}

func (m *MockNodeService) TransitionToSharedPlan(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockNodeService) TransitionToDedicatedPlan(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func setupNodeHandler() (*Handler, *MockNodeService) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockService := new(MockNodeService)
	
	handler := NewHandler(mockService, logger)
	
	return handler, mockService
}

func setupGinContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	
	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Add user_id to context (simulating auth middleware)
	c.Set("user_id", "test-user-123")
	
	return c, w
}

func TestNodeHandler_GetAvailablePlans(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockNodeService)
		expectedStatus int
		expectedPlans  int
	}{
		{
			name: "successful plans retrieval",
			setupMock: func(mockService *MockNodeService) {
				plans := []node.NodePlan{
					{
						ID:   "shared-plan",
						Name: "Shared Plan",
						Type: node.PlanTypeShared,
						Resources: node.ResourceSpec{
							CPUCores:  2,
							MemoryGB:  4,
							StorageGB: 100,
						},
						PricePerMonth: 0,
					},
					{
						ID:   "s-type-plan",
						Name: "S-Type Dedicated",
						Type: node.PlanTypeDedicated,
						Resources: node.ResourceSpec{
							CPUCores:  4,
							MemoryGB:  16,
							StorageGB: 200,
						},
						PricePerMonth: 99.99,
					},
				}
				mockService.On("GetAvailablePlans", mock.Anything).Return(plans, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPlans:  2,
		},
		{
			name: "service error",
			setupMock: func(mockService *MockNodeService) {
				mockService.On("GetAvailablePlans", mock.Anything).Return([]node.NodePlan{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedPlans:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupNodeHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			c, w := setupGinContext("GET", "/plans", nil)

			handler.GetAvailablePlans(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				
				plans, ok := response["plans"].([]interface{})
				require.True(t, ok)
				assert.Len(t, plans, tt.expectedPlans)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestNodeHandler_ProvisionDedicatedNode(t *testing.T) {
	tests := []struct {
		name           string
		workspaceID    string
		requestBody    ProvisionRequest
		setupMock      func(*MockNodeService)
		expectedStatus int
		expectNodeID   string
	}{
		{
			name:        "successful node provisioning",
			workspaceID: "ws-123",
			requestBody: ProvisionRequest{
				NodeName:     "test-node",
				NodeType:     "S-Type",
				Region:       "us-west-1",
				SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2E...",
				Labels:       map[string]string{"env": "test"},
			},
			setupMock: func(mockService *MockNodeService) {
				expectedReq := node.ProvisionRequest{
					NodeName:     "test-node",
					NodeType:     "S-Type",
					Region:       "us-west-1",
					SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2E...",
					Labels:       map[string]string{"env": "test"},
				}
				mockNode := &node.DedicatedNode{
					ID:          "node-123",
					WorkspaceID: "ws-123",
					Name:        "test-node",
					Status:      node.NodeStatusProvisioning,
					Specification: node.NodeSpecification{
						Type:      "S-Type",
						CPUCores:  4,
						MemoryGB:  16,
						StorageGB: 200,
					},
					CreatedAt: time.Now(),
				}
				mockService.On("ProvisionDedicatedNode", mock.Anything, "ws-123", expectedReq).Return(mockNode, nil)
			},
			expectedStatus: http.StatusCreated,
			expectNodeID:   "node-123",
		},
		{
			name:        "invalid request payload",
			workspaceID: "ws-123",
			requestBody: ProvisionRequest{
				// Missing required fields
				NodeType: "S-Type",
			},
			setupMock:      func(mockService *MockNodeService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error",
			workspaceID: "ws-123",
			requestBody: ProvisionRequest{
				NodeName:     "test-node",
				NodeType:     "S-Type",
				SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2E...",
			},
			setupMock: func(mockService *MockNodeService) {
				mockService.On("ProvisionDedicatedNode", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("insufficient resources"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupNodeHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			c, w := setupGinContext("POST", "/workspaces/"+tt.workspaceID+"/nodes", tt.requestBody)
			c.Params = gin.Params{{Key: "wsId", Value: tt.workspaceID}}

			handler.ProvisionDedicatedNode(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				nodeData, ok := response["node"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, tt.expectNodeID, nodeData["id"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestNodeHandler_NodeLifecycle(t *testing.T) {
	tests := []struct {
		name           string
		nodeID         string
		operation      string
		handler        func(*Handler, *gin.Context)
		setupMock      func(*MockNodeService, string)
		expectedStatus int
	}{
		{
			name:      "start node successfully",
			nodeID:    "node-123",
			operation: "start",
			handler:   (*Handler).StartNode,
			setupMock: func(mockService *MockNodeService, nodeID string) {
				mockService.On("StartNode", mock.Anything, nodeID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "stop node successfully",
			nodeID:    "node-123",
			operation: "stop",
			handler:   (*Handler).StopNode,
			setupMock: func(mockService *MockNodeService, nodeID string) {
				mockService.On("StopNode", mock.Anything, nodeID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "reboot node successfully",
			nodeID:    "node-123",
			operation: "reboot",
			handler:   (*Handler).RebootNode,
			setupMock: func(mockService *MockNodeService, nodeID string) {
				mockService.On("RebootNode", mock.Anything, nodeID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "delete node successfully",
			nodeID:    "node-123",
			operation: "delete",
			handler:   (*Handler).DeleteNode,
			setupMock: func(mockService *MockNodeService, nodeID string) {
				mockService.On("DeleteNode", mock.Anything, nodeID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "node not found error",
			nodeID:    "node-nonexistent",
			operation: "start",
			handler:   (*Handler).StartNode,
			setupMock: func(mockService *MockNodeService, nodeID string) {
				mockService.On("StartNode", mock.Anything, nodeID).Return(errors.New("node not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupNodeHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService, tt.nodeID)
			}

			c, w := setupGinContext("POST", "/nodes/"+tt.nodeID+"/"+tt.operation, nil)
			c.Params = gin.Params{{Key: "nodeId", Value: tt.nodeID}}

			tt.handler(handler, c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				message := response["message"].(string)
				switch tt.operation {
				case "delete":
					assert.Contains(t, message, "deletion")
				default:
					assert.Contains(t, message, tt.operation)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestNodeHandler_GetWorkspaceResourceUsage(t *testing.T) {
	tests := []struct {
		name           string
		workspaceID    string
		setupMock      func(*MockNodeService)
		expectedStatus int
		expectPlanType string
	}{
		{
			name:        "shared plan resource usage",
			workspaceID: "ws-shared",
			setupMock: func(mockService *MockNodeService) {
				usage := &node.WorkspaceResourceUsage{
					WorkspaceID: "ws-shared",
					PlanType:    node.PlanTypeShared,
					SharedUsage: &node.SharedResourceUsage{
						CPUUsed:       2,
						CPULimit:      4,
						MemoryUsedGB:  4,
						MemoryLimitGB: 8,
					},
					Timestamp: time.Now(),
				}
				mockService.On("GetWorkspaceResourceUsage", mock.Anything, "ws-shared").Return(usage, nil)
			},
			expectedStatus: http.StatusOK,
			expectPlanType: node.PlanTypeShared,
		},
		{
			name:        "service error",
			workspaceID: "ws-error",
			setupMock: func(mockService *MockNodeService) {
				mockService.On("GetWorkspaceResourceUsage", mock.Anything, "ws-error").Return(nil, errors.New("workspace not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupNodeHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			c, w := setupGinContext("GET", "/workspaces/"+tt.workspaceID+"/usage", nil)
			c.Params = gin.Params{{Key: "wsId", Value: tt.workspaceID}}

			handler.GetWorkspaceResourceUsage(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				usage, ok := response["usage"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, tt.expectPlanType, usage["plan_type"])
			}

			mockService.AssertExpectations(t)
		})
	}
}