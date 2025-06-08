package node

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/hexabase/hexabase-ai/api/internal/domain/node"
)

// MockNodeRepository is a mock implementation of node.Repository
type MockNodeRepository struct {
	mock.Mock
}

func (m *MockNodeRepository) GetNodePlans(ctx context.Context) ([]node.NodePlan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]node.NodePlan), args.Error(1)
}

func (m *MockNodeRepository) GetNodePlan(ctx context.Context, planID string) (*node.NodePlan, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.NodePlan), args.Error(1)
}

func (m *MockNodeRepository) CreateDedicatedNode(ctx context.Context, n *node.DedicatedNode) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockNodeRepository) GetDedicatedNode(ctx context.Context, nodeID string) (*node.DedicatedNode, error) {
	args := m.Called(ctx, nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.DedicatedNode), args.Error(1)
}

func (m *MockNodeRepository) GetDedicatedNodeByVMID(ctx context.Context, vmid int) (*node.DedicatedNode, error) {
	args := m.Called(ctx, vmid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.DedicatedNode), args.Error(1)
}

func (m *MockNodeRepository) ListDedicatedNodes(ctx context.Context, workspaceID string) ([]node.DedicatedNode, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]node.DedicatedNode), args.Error(1)
}

func (m *MockNodeRepository) UpdateDedicatedNode(ctx context.Context, n *node.DedicatedNode) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockNodeRepository) DeleteDedicatedNode(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func (m *MockNodeRepository) CreateNodeEvent(ctx context.Context, event *node.NodeEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockNodeRepository) ListNodeEvents(ctx context.Context, nodeID string, limit int) ([]node.NodeEvent, error) {
	args := m.Called(ctx, nodeID, limit)
	return args.Get(0).([]node.NodeEvent), args.Error(1)
}

func (m *MockNodeRepository) GetWorkspaceAllocation(ctx context.Context, workspaceID string) (*node.WorkspaceNodeAllocation, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.WorkspaceNodeAllocation), args.Error(1)
}

func (m *MockNodeRepository) CreateWorkspaceAllocation(ctx context.Context, allocation *node.WorkspaceNodeAllocation) error {
	args := m.Called(ctx, allocation)
	return args.Error(0)
}

func (m *MockNodeRepository) UpdateWorkspaceAllocation(ctx context.Context, allocation *node.WorkspaceNodeAllocation) error {
	args := m.Called(ctx, allocation)
	return args.Error(0)
}

func (m *MockNodeRepository) UpdateSharedQuotaUsage(ctx context.Context, workspaceID string, cpu, memory float64) error {
	args := m.Called(ctx, workspaceID, cpu, memory)
	return args.Error(0)
}

func (m *MockNodeRepository) GetNodeResourceUsage(ctx context.Context, nodeID string) (*node.ResourceUsage, error) {
	args := m.Called(ctx, nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.ResourceUsage), args.Error(1)
}

// MockProxmoxRepository is a mock implementation of node.ProxmoxRepository
type MockProxmoxRepository struct {
	mock.Mock
}

func (m *MockProxmoxRepository) CreateVM(ctx context.Context, spec node.VMSpec) (*node.ProxmoxVMInfo, error) {
	args := m.Called(ctx, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.ProxmoxVMInfo), args.Error(1)
}

func (m *MockProxmoxRepository) GetVM(ctx context.Context, vmid int) (*node.ProxmoxVMInfo, error) {
	args := m.Called(ctx, vmid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.ProxmoxVMInfo), args.Error(1)
}

func (m *MockProxmoxRepository) StartVM(ctx context.Context, vmid int) error {
	args := m.Called(ctx, vmid)
	return args.Error(0)
}

func (m *MockProxmoxRepository) StopVM(ctx context.Context, vmid int) error {
	args := m.Called(ctx, vmid)
	return args.Error(0)
}

func (m *MockProxmoxRepository) RebootVM(ctx context.Context, vmid int) error {
	args := m.Called(ctx, vmid)
	return args.Error(0)
}

func (m *MockProxmoxRepository) DeleteVM(ctx context.Context, vmid int) error {
	args := m.Called(ctx, vmid)
	return args.Error(0)
}

func (m *MockProxmoxRepository) UpdateVMConfig(ctx context.Context, vmid int, config node.VMConfig) error {
	args := m.Called(ctx, vmid, config)
	return args.Error(0)
}

func (m *MockProxmoxRepository) GetVMStatus(ctx context.Context, vmid int) (string, error) {
	args := m.Called(ctx, vmid)
	return args.String(0), args.Error(1)
}

func (m *MockProxmoxRepository) SetCloudInitConfig(ctx context.Context, vmid int, config node.CloudInitConfig) error {
	args := m.Called(ctx, vmid, config)
	return args.Error(0)
}

func (m *MockProxmoxRepository) GetVMResourceUsage(ctx context.Context, vmid int) (*node.VMResourceUsage, error) {
	args := m.Called(ctx, vmid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*node.VMResourceUsage), args.Error(1)
}

func (m *MockProxmoxRepository) CloneTemplate(ctx context.Context, templateID int, name string) (int, error) {
	args := m.Called(ctx, templateID, name)
	return args.Int(0), args.Error(1)
}

func (m *MockProxmoxRepository) ListTemplates(ctx context.Context) ([]node.VMTemplate, error) {
	args := m.Called(ctx)
	return args.Get(0).([]node.VMTemplate), args.Error(1)
}

func TestService_CheckK3sAgentStatus(t *testing.T) {
	tests := []struct {
		name           string
		nodeID         string
		setupMocks     func(*MockNodeRepository, *MockProxmoxRepository)
		setupK8s       func() *fake.Clientset
		expectedStatus string
		expectError    bool
		errorContains  string
	}{
		{
			name:   "agent ready - node found in cluster",
			nodeID: "node-123",
			setupMocks: func(mr *MockNodeRepository, pr *MockProxmoxRepository) {
				mr.On("GetDedicatedNode", mock.Anything, "node-123").Return(&node.DedicatedNode{
					ID:          "node-123",
					Name:        "test-node",
					WorkspaceID: "ws-123",
					Status:      node.NodeStatusReady,
					ProxmoxVMID: 1001,
				}, nil)
			},
			setupK8s: func() *fake.Clientset {
				client := fake.NewSimpleClientset()
				// Create a node that matches our dedicated node
				k8sNode := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node",
						Labels: map[string]string{
							"node.hexabase.io/workspace": "ws-123",
							"node.hexabase.io/node-id":   "node-123",
						},
					},
					Status: corev1.NodeStatus{
						Conditions: []corev1.NodeCondition{
							{
								Type:               corev1.NodeReady,
								Status:             corev1.ConditionTrue,
								LastHeartbeatTime:  metav1.NewTime(time.Now()),
								LastTransitionTime: metav1.NewTime(time.Now()),
								Reason:             "KubeletReady",
								Message:            "kubelet is posting ready status",
							},
						},
						NodeInfo: corev1.NodeSystemInfo{
							KubeletVersion: "v1.27.3+k3s1",
						},
					},
				}
				client.CoreV1().Nodes().Create(context.Background(), k8sNode, metav1.CreateOptions{})
				return client
			},
			expectedStatus: "ready",
			expectError:    false,
		},
		{
			name:   "agent not ready - node conditions not met",
			nodeID: "node-456",
			setupMocks: func(mr *MockNodeRepository, pr *MockProxmoxRepository) {
				mr.On("GetDedicatedNode", mock.Anything, "node-456").Return(&node.DedicatedNode{
					ID:          "node-456",
					Name:        "test-node-2",
					WorkspaceID: "ws-456",
					Status:      node.NodeStatusReady,
					ProxmoxVMID:        1002,
				}, nil)
			},
			setupK8s: func() *fake.Clientset {
				client := fake.NewSimpleClientset()
				// Create a node that's not ready
				k8sNode := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-2",
						Labels: map[string]string{
							"node.hexabase.io/workspace": "ws-456",
							"node.hexabase.io/node-id":   "node-456",
						},
					},
					Status: corev1.NodeStatus{
						Conditions: []corev1.NodeCondition{
							{
								Type:               corev1.NodeReady,
								Status:             corev1.ConditionFalse,
								LastHeartbeatTime:  metav1.NewTime(time.Now()),
								LastTransitionTime: metav1.NewTime(time.Now()),
								Reason:             "KubeletNotReady",
								Message:            "kubelet is not ready",
							},
						},
					},
				}
				client.CoreV1().Nodes().Create(context.Background(), k8sNode, metav1.CreateOptions{})
				return client
			},
			expectedStatus: "not_ready",
			expectError:    false,
		},
		{
			name:   "agent not found - node not in cluster",
			nodeID: "node-789",
			setupMocks: func(mr *MockNodeRepository, pr *MockProxmoxRepository) {
				mr.On("GetDedicatedNode", mock.Anything, "node-789").Return(&node.DedicatedNode{
					ID:          "node-789",
					Name:        "test-node-3",
					WorkspaceID: "ws-789",
					Status:      node.NodeStatusReady,
					ProxmoxVMID:        1003,
				}, nil)
			},
			setupK8s: func() *fake.Clientset {
				// Empty client with no nodes
				return fake.NewSimpleClientset()
			},
			expectedStatus: "not_found",
			expectError:    false,
		},
		{
			name:   "node status provisioning - expected not ready",
			nodeID: "node-999",
			setupMocks: func(mr *MockNodeRepository, pr *MockProxmoxRepository) {
				mr.On("GetDedicatedNode", mock.Anything, "node-999").Return(&node.DedicatedNode{
					ID:          "node-999",
					Name:        "test-node-4",
					WorkspaceID: "ws-999",
					Status:      node.NodeStatusProvisioning,
					ProxmoxVMID:        1004,
				}, nil)
			},
			setupK8s: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			expectedStatus: "provisioning",
			expectError:    false,
		},
		{
			name:   "database error",
			nodeID: "node-error",
			setupMocks: func(mr *MockNodeRepository, pr *MockProxmoxRepository) {
				mr.On("GetDedicatedNode", mock.Anything, "node-error").Return(nil, errors.New("database error"))
			},
			setupK8s: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			expectedStatus: "",
			expectError:    true,
			errorContains:  "database error",
		},
		{
			name:   "node has old heartbeat - stale",
			nodeID: "node-stale",
			setupMocks: func(mr *MockNodeRepository, pr *MockProxmoxRepository) {
				mr.On("GetDedicatedNode", mock.Anything, "node-stale").Return(&node.DedicatedNode{
					ID:          "node-stale",
					Name:        "test-node-stale",
					WorkspaceID: "ws-stale",
					Status:      node.NodeStatusReady,
					ProxmoxVMID:        1005,
				}, nil)
			},
			setupK8s: func() *fake.Clientset {
				client := fake.NewSimpleClientset()
				// Create a node with old heartbeat
				k8sNode := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-stale",
						Labels: map[string]string{
							"node.hexabase.io/workspace": "ws-stale",
							"node.hexabase.io/node-id":   "node-stale",
						},
					},
					Status: corev1.NodeStatus{
						Conditions: []corev1.NodeCondition{
							{
								Type:               corev1.NodeReady,
								Status:             corev1.ConditionTrue,
								LastHeartbeatTime:  metav1.NewTime(time.Now().Add(-10 * time.Minute)),
								LastTransitionTime: metav1.NewTime(time.Now().Add(-10 * time.Minute)),
								Reason:             "KubeletReady",
								Message:            "kubelet is posting ready status",
							},
						},
						NodeInfo: corev1.NodeSystemInfo{
							KubeletVersion: "v1.27.3+k3s1",
						},
					},
				}
				client.CoreV1().Nodes().Create(context.Background(), k8sNode, metav1.CreateOptions{})
				return client
			},
			expectedStatus: "stale",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockNodeRepo := new(MockNodeRepository)
			mockProxmoxRepo := new(MockProxmoxRepository)
			
			if tt.setupMocks != nil {
				tt.setupMocks(mockNodeRepo, mockProxmoxRepo)
			}

			// Create service with fake k8s client
			k8sClient := tt.setupK8s()
			svc := NewService(mockNodeRepo, mockProxmoxRepo)
			svc.SetK8sClient(k8sClient)

			// Call method
			status, err := svc.CheckK3sAgentStatus(context.Background(), tt.nodeID)

			// Assert results
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}

			// Verify mock calls
			mockNodeRepo.AssertExpectations(t)
			mockProxmoxRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetK3sAgentConditions(t *testing.T) {
	tests := []struct {
		name               string
		nodeID             string
		setupMocks         func(*MockNodeRepository, *MockProxmoxRepository)
		setupK8s           func() *fake.Clientset
		expectedConditions []node.NodeCondition
		expectError        bool
	}{
		{
			name:   "get all node conditions",
			nodeID: "node-123",
			setupMocks: func(mr *MockNodeRepository, pr *MockProxmoxRepository) {
				mr.On("GetDedicatedNode", mock.Anything, "node-123").Return(&node.DedicatedNode{
					ID:          "node-123",
					Name:        "test-node",
					WorkspaceID: "ws-123",
					Status:      node.NodeStatusReady,
					ProxmoxVMID:        1001,
				}, nil)
			},
			setupK8s: func() *fake.Clientset {
				client := fake.NewSimpleClientset()
				k8sNode := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node",
						Labels: map[string]string{
							"node.hexabase.io/workspace": "ws-123",
							"node.hexabase.io/node-id":   "node-123",
						},
					},
					Status: corev1.NodeStatus{
						Conditions: []corev1.NodeCondition{
							{
								Type:               corev1.NodeReady,
								Status:             corev1.ConditionTrue,
								LastHeartbeatTime:  metav1.NewTime(time.Now()),
								LastTransitionTime: metav1.NewTime(time.Now()),
								Reason:             "KubeletReady",
								Message:            "kubelet is posting ready status",
							},
							{
								Type:               corev1.NodeMemoryPressure,
								Status:             corev1.ConditionFalse,
								LastHeartbeatTime:  metav1.NewTime(time.Now()),
								LastTransitionTime: metav1.NewTime(time.Now()),
								Reason:             "KubeletHasSufficientMemory",
								Message:            "kubelet has sufficient memory available",
							},
							{
								Type:               corev1.NodeDiskPressure,
								Status:             corev1.ConditionFalse,
								LastHeartbeatTime:  metav1.NewTime(time.Now()),
								LastTransitionTime: metav1.NewTime(time.Now()),
								Reason:             "KubeletHasNoDiskPressure",
								Message:            "kubelet has no disk pressure",
							},
						},
					},
				}
				client.CoreV1().Nodes().Create(context.Background(), k8sNode, metav1.CreateOptions{})
				return client
			},
			expectedConditions: []node.NodeCondition{
				{
					Type:    "Ready",
					Status:  "True",
					Reason:  "KubeletReady",
					Message: "kubelet is posting ready status",
				},
				{
					Type:    "MemoryPressure",
					Status:  "False",
					Reason:  "KubeletHasSufficientMemory",
					Message: "kubelet has sufficient memory available",
				},
				{
					Type:    "DiskPressure",
					Status:  "False",
					Reason:  "KubeletHasNoDiskPressure",
					Message: "kubelet has no disk pressure",
				},
			},
			expectError: false,
		},
		{
			name:   "node not found returns empty conditions",
			nodeID: "node-notfound",
			setupMocks: func(mr *MockNodeRepository, pr *MockProxmoxRepository) {
				mr.On("GetDedicatedNode", mock.Anything, "node-notfound").Return(&node.DedicatedNode{
					ID:          "node-notfound",
					Name:        "test-node-notfound",
					WorkspaceID: "ws-notfound",
					Status:      node.NodeStatusReady,
					ProxmoxVMID:        1006,
				}, nil)
			},
			setupK8s: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			expectedConditions: []node.NodeCondition{},
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockNodeRepo := new(MockNodeRepository)
			mockProxmoxRepo := new(MockProxmoxRepository)
			
			if tt.setupMocks != nil {
				tt.setupMocks(mockNodeRepo, mockProxmoxRepo)
			}

			// Create service with fake k8s client
			k8sClient := tt.setupK8s()
			svc := NewService(mockNodeRepo, mockProxmoxRepo)
			svc.SetK8sClient(k8sClient)

			// Call method
			conditions, err := svc.GetK3sAgentConditions(context.Background(), tt.nodeID)

			// Assert results
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, conditions, len(tt.expectedConditions))
				
				// Compare conditions (ignoring timestamps)
				for i, expected := range tt.expectedConditions {
					assert.Equal(t, expected.Type, conditions[i].Type)
					assert.Equal(t, expected.Status, conditions[i].Status)
					assert.Equal(t, expected.Reason, conditions[i].Reason)
					assert.Equal(t, expected.Message, conditions[i].Message)
				}
			}

			// Verify mock calls
			mockNodeRepo.AssertExpectations(t)
			mockProxmoxRepo.AssertExpectations(t)
		})
	}
}