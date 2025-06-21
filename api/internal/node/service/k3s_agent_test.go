package service_test

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

	"github.com/hexabase/hexabase-ai/api/internal/node/domain"
	nodeService "github.com/hexabase/hexabase-ai/api/internal/node/service"
)

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
				mr.On("GetDedicatedNode", mock.Anything, "node-123").Return(&domain.DedicatedNode{
					ID:          "node-123",
					Name:        "test-node",
					WorkspaceID: "ws-123",
					Status:      domain.NodeStatusReady,
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
							"domain.hexabase.io/workspace": "ws-123",
							"domain.hexabase.io/node-id":   "node-123",
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
				mr.On("GetDedicatedNode", mock.Anything, "node-456").Return(&domain.DedicatedNode{
					ID:          "node-456",
					Name:        "test-node-2",
					WorkspaceID: "ws-456",
					Status:      domain.NodeStatusReady,
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
							"domain.hexabase.io/workspace": "ws-456",
							"domain.hexabase.io/node-id":   "node-456",
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
				mr.On("GetDedicatedNode", mock.Anything, "node-789").Return(&domain.DedicatedNode{
					ID:          "node-789",
					Name:        "test-node-3",
					WorkspaceID: "ws-789",
					Status:      domain.NodeStatusReady,
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
				mr.On("GetDedicatedNode", mock.Anything, "node-999").Return(&domain.DedicatedNode{
					ID:          "node-999",
					Name:        "test-node-4",
					WorkspaceID: "ws-999",
					Status:      domain.NodeStatusProvisioning,
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
				mr.On("GetDedicatedNode", mock.Anything, "node-stale").Return(&domain.DedicatedNode{
					ID:          "node-stale",
					Name:        "test-node-stale",
					WorkspaceID: "ws-stale",
					Status:      domain.NodeStatusReady,
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
							"domain.hexabase.io/workspace": "ws-stale",
							"domain.hexabase.io/node-id":   "node-stale",
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
			svc := nodeService.NewService(mockNodeRepo, mockProxmoxRepo)
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
		expectedConditions []domain.NodeCondition
		expectError        bool
	}{
		{
			name:   "get all node conditions",
			nodeID: "node-123",
			setupMocks: func(mr *MockNodeRepository, pr *MockProxmoxRepository) {
				mr.On("GetDedicatedNode", mock.Anything, "node-123").Return(&domain.DedicatedNode{
					ID:          "node-123",
					Name:        "test-node",
					WorkspaceID: "ws-123",
					Status:      domain.NodeStatusReady,
					ProxmoxVMID:        1001,
				}, nil)
			},
			setupK8s: func() *fake.Clientset {
				client := fake.NewSimpleClientset()
				k8sNode := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node",
						Labels: map[string]string{
							"domain.hexabase.io/workspace": "ws-123",
							"domain.hexabase.io/node-id":   "node-123",
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
			expectedConditions: []domain.NodeCondition{
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
				mr.On("GetDedicatedNode", mock.Anything, "node-notfound").Return(&domain.DedicatedNode{
					ID:          "node-notfound",
					Name:        "test-node-notfound",
					WorkspaceID: "ws-notfound",
					Status:      domain.NodeStatusReady,
					ProxmoxVMID:        1006,
				}, nil)
			},
			setupK8s: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			expectedConditions: []domain.NodeCondition{},
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
			svc := nodeService.NewService(mockNodeRepo, mockProxmoxRepo)
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