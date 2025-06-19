package domain_test

import (
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/node/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodePlan_Validation(t *testing.T) {
	tests := []struct {
		name    string
		plan    *domain.NodePlan
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid shared plan",
			plan: &domain.NodePlan{
				ID:   "shared-basic",
				Name: "Shared Basic",
				Type: domain.PlanTypeShared,
				Resources: domain.ResourceSpec{
					CPUCores:    2,
					MemoryGB:    4,
					StorageGB:   50,
					MaxPods:     50,
					MaxServices: 10,
				},
				PricePerMonth: 0,
			},
			wantErr: false,
		},
		{
			name: "valid dedicated plan",
			plan: &domain.NodePlan{
				ID:   "dedicated-s",
				Name: "Dedicated S-Type",
				Type: domain.PlanTypeDedicated,
				Resources: domain.ResourceSpec{
					CPUCores:    4,
					MemoryGB:    16,
					StorageGB:   200,
					MaxPods:     200,
					MaxServices: 50,
				},
				PricePerMonth: 99.99,
			},
			wantErr: false,
		},
		{
			name: "invalid plan type",
			plan: &domain.NodePlan{
				ID:   "invalid",
				Name: "Invalid Plan",
				Type: "invalid-type",
			},
			wantErr: true,
			errMsg:  "invalid plan type",
		},
		{
			name: "missing resources",
			plan: &domain.NodePlan{
				ID:   "no-resources",
				Name: "No Resources",
				Type: domain.PlanTypeShared,
			},
			wantErr: true,
			errMsg:  "invalid resource specification",
		},
		{
			name: "negative price for dedicated plan",
			plan: &domain.NodePlan{
				ID:   "negative-price",
				Name: "Negative Price",
				Type: domain.PlanTypeDedicated,
				Resources: domain.ResourceSpec{
					CPUCores:  4,
					MemoryGB:  16,
					StorageGB: 200,
				},
				PricePerMonth: -10,
			},
			wantErr: true,
			errMsg:  "price cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plan.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDedicatedNode_StateTransitions(t *testing.T) {
	tests := []struct {
		name         string
		currentState domain.NodeStatus
		action       string
		expectedNext domain.NodeStatus
		shouldError  bool
	}{
		{
			name:         "provision new node",
			currentState: "",
			action:       "provision",
			expectedNext: domain.NodeStatusProvisioning,
			shouldError:  false,
		},
		{
			name:         "provisioning to ready",
			currentState: domain.NodeStatusProvisioning,
			action:       "complete",
			expectedNext: domain.NodeStatusReady,
			shouldError:  false,
		},
		{
			name:         "ready to stopping",
			currentState: domain.NodeStatusReady,
			action:       "stop",
			expectedNext: domain.NodeStatusStopping,
			shouldError:  false,
		},
		{
			name:         "stopped to starting",
			currentState: domain.NodeStatusStopped,
			action:       "start",
			expectedNext: domain.NodeStatusStarting,
			shouldError:  false,
		},
		{
			name:         "invalid transition - stopped to ready",
			currentState: domain.NodeStatusStopped,
			action:       "complete",
			expectedNext: "",
			shouldError:  true,
		},
		{
			name:         "provisioning failed",
			currentState: domain.NodeStatusProvisioning,
			action:       "fail",
			expectedNext: domain.NodeStatusFailed,
			shouldError:  false,
		},
		{
			name:         "ready to deleting",
			currentState: domain.NodeStatusReady,
			action:       "delete",
			expectedNext: domain.NodeStatusDeleting,
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dedicatedNode := &domain.DedicatedNode{
				Status: tt.currentState,
			}

			err := dedicatedNode.TransitionTo(tt.action)
			if tt.shouldError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedNext, dedicatedNode.Status)
			}
		})
	}
}

func TestDedicatedNode_CanScheduleWorkload(t *testing.T) {
	tests := []struct {
		name     string
		node     *domain.DedicatedNode
		expected bool
	}{
		{
			name: "ready node can schedule",
			node: &domain.DedicatedNode{
				Status: domain.NodeStatusReady,
			},
			expected: true,
		},
		{
			name: "provisioning node cannot schedule",
			node: &domain.DedicatedNode{
				Status: domain.NodeStatusProvisioning,
			},
			expected: false,
		},
		{
			name: "stopped node cannot schedule",
			node: &domain.DedicatedNode{
				Status: domain.NodeStatusStopped,
			},
			expected: false,
		},
		{
			name: "failed node cannot schedule",
			node: &domain.DedicatedNode{
				Status: domain.NodeStatusFailed,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.CanScheduleWorkload()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNodeSpecification_Validation(t *testing.T) {
	tests := []struct {
		name    string
		spec    domain.NodeSpecification
		wantErr bool
	}{
		{
			name: "valid S-Type spec",
			spec: domain.NodeSpecification{
				Type:      "S-Type",
				CPUCores:  4,
				MemoryGB:  16,
				StorageGB: 200,
				NetworkMbps: 1000,
			},
			wantErr: false,
		},
		{
			name: "valid M-Type spec",
			spec: domain.NodeSpecification{
				Type:      "M-Type",
				CPUCores:  8,
				MemoryGB:  32,
				StorageGB: 500,
				NetworkMbps: 2000,
			},
			wantErr: false,
		},
		{
			name: "invalid - zero CPU",
			spec: domain.NodeSpecification{
				Type:      "S-Type",
				CPUCores:  0,
				MemoryGB:  16,
				StorageGB: 200,
			},
			wantErr: true,
		},
		{
			name: "invalid - insufficient memory",
			spec: domain.NodeSpecification{
				Type:      "S-Type",
				CPUCores:  4,
				MemoryGB:  2, // Too low
				StorageGB: 200,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProxmoxVMInfo_IsRunning(t *testing.T) {
	tests := []struct {
		name     string
		vmInfo   domain.ProxmoxVMInfo
		expected bool
	}{
		{
			name: "running VM",
			vmInfo: domain.ProxmoxVMInfo{
				VMID:   100,
				Status: "running",
			},
			expected: true,
		},
		{
			name: "stopped VM",
			vmInfo: domain.ProxmoxVMInfo{
				VMID:   101,
				Status: "stopped",
			},
			expected: false,
		},
		{
			name: "paused VM",
			vmInfo: domain.ProxmoxVMInfo{
				VMID:   102,
				Status: "paused",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.vmInfo.IsRunning()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNodeEvent_Creation(t *testing.T) {
	nodeObj := &domain.DedicatedNode{
		ID:          "node-123",
		WorkspaceID: "ws-456",
		Name:        "dedicated-node-1",
	}

	event := nodeObj.CreateEvent(domain.EventTypeProvisioning, "Starting VM provisioning")
	
	assert.Equal(t, "node-123", event.NodeID)
	assert.Equal(t, "ws-456", event.WorkspaceID)
	assert.Equal(t, domain.EventTypeProvisioning, event.Type)
	assert.Equal(t, "Starting VM provisioning", event.Message)
	assert.WithinDuration(t, time.Now(), event.Timestamp, time.Second)
}

func TestWorkspaceNodeAllocation_QuotaCheck(t *testing.T) {
	allocation := &domain.WorkspaceNodeAllocation{
		WorkspaceID: "ws-123",
		PlanType:    domain.PlanTypeShared,
		SharedQuota: &domain.SharedQuota{
			CPULimit:    4,
			MemoryLimit: 8,
			CPUUsed:     2,
			MemoryUsed:  4,
		},
	}

	tests := []struct {
		name      string
		requested domain.ResourceRequest
		expected  bool
	}{
		{
			name: "within quota",
			requested: domain.ResourceRequest{
				CPU:    1,
				Memory: 2,
			},
			expected: true,
		},
		{
			name: "exceeds CPU quota",
			requested: domain.ResourceRequest{
				CPU:    3,
				Memory: 2,
			},
			expected: false,
		},
		{
			name: "exceeds memory quota",
			requested: domain.ResourceRequest{
				CPU:    1,
				Memory: 5,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := allocation.CanAllocate(tt.requested)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDedicatedNode_BillingCalculation(t *testing.T) {
	dedicatedNode := &domain.DedicatedNode{
		ID: "node-123",
		Specification: domain.NodeSpecification{
			Type: "S-Type",
		},
		CreatedAt: time.Now().Add(-24 * time.Hour), // Created 1 day ago
		Status:    domain.NodeStatusReady,
	}

	// Assuming S-Type costs $99.99/month
	hourlyRate := 99.99 / (30 * 24) // Simplified monthly to hourly
	expectedCost := hourlyRate * 24  // 24 hours

	cost := dedicatedNode.CalculateUsageCost(time.Now())
	assert.InDelta(t, expectedCost, cost, 0.01)
}

func TestNodePool_Selection(t *testing.T) {
	pool := &domain.NodePool{
		Name: "production-pool",
		Nodes: []domain.DedicatedNode{
			{ID: "node-1", Status: domain.NodeStatusReady},
			{ID: "node-2", Status: domain.NodeStatusReady},
			{ID: "node-3", Status: domain.NodeStatusStopped},
		},
	}

	available := pool.GetAvailableNodes()
	assert.Len(t, available, 2)
	assert.Equal(t, "node-1", available[0].ID)
	assert.Equal(t, "node-2", available[1].ID)
}