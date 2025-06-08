package node

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hexabase/hexabase-ai/api/internal/domain/node"
)

// PostgresRepository implements the node.Repository interface using PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetNodePlans returns all available node plans
func (r *PostgresRepository) GetNodePlans(ctx context.Context) ([]node.NodePlan, error) {
	// In a real implementation, these would come from the database
	// For now, return hardcoded plans
	plans := []node.NodePlan{
		{
			ID:   "shared-plan",
			Name: "Shared Plan",
			Type: node.PlanTypeShared,
			Resources: node.ResourceSpec{
				CPUCores:    2,
				MemoryGB:    4,
				StorageGB:   100,
				MaxPods:     50,
				MaxServices: 10,
			},
			PricePerMonth: 0,
			Description:   "Perfect for development and testing",
			Features: []string{
				"Shared infrastructure",
				"Auto-scaling",
				"99.5% SLA",
				"Community support",
			},
		},
		{
			ID:   "s-type-plan",
			Name: "S-Type Dedicated",
			Type: node.PlanTypeDedicated,
			Resources: node.ResourceSpec{
				CPUCores:    4,
				MemoryGB:    16,
				StorageGB:   200,
				MaxPods:     100,
				MaxServices: 50,
			},
			PricePerMonth: 99.99,
			Description:   "Small dedicated node for production workloads",
			Features: []string{
				"Dedicated VM",
				"4 vCPUs",
				"16GB RAM",
				"200GB SSD",
				"99.9% SLA",
				"Email support",
			},
		},
		{
			ID:   "m-type-plan",
			Name: "M-Type Dedicated",
			Type: node.PlanTypeDedicated,
			Resources: node.ResourceSpec{
				CPUCores:    8,
				MemoryGB:    32,
				StorageGB:   500,
				MaxPods:     200,
				MaxServices: 100,
			},
			PricePerMonth: 199.99,
			Description:   "Medium dedicated node for growing applications",
			Features: []string{
				"Dedicated VM",
				"8 vCPUs",
				"32GB RAM",
				"500GB SSD",
				"99.95% SLA",
				"Priority support",
			},
		},
		{
			ID:   "l-type-plan",
			Name: "L-Type Dedicated",
			Type: node.PlanTypeDedicated,
			Resources: node.ResourceSpec{
				CPUCores:    16,
				MemoryGB:    64,
				StorageGB:   1000,
				MaxPods:     500,
				MaxServices: 200,
			},
			PricePerMonth: 399.99,
			Description:   "Large dedicated node for enterprise workloads",
			Features: []string{
				"Dedicated VM",
				"16 vCPUs",
				"64GB RAM",
				"1TB SSD",
				"99.99% SLA",
				"24/7 phone support",
			},
		},
	}

	return plans, nil
}

// GetNodePlan returns a specific node plan
func (r *PostgresRepository) GetNodePlan(ctx context.Context, planID string) (*node.NodePlan, error) {
	plans, err := r.GetNodePlans(ctx)
	if err != nil {
		return nil, err
	}

	for _, plan := range plans {
		if plan.ID == planID {
			return &plan, nil
		}
	}

	return nil, fmt.Errorf("plan not found: %s", planID)
}

// CreateDedicatedNode creates a new dedicated node record
func (r *PostgresRepository) CreateDedicatedNode(ctx context.Context, node *node.DedicatedNode) error {
	if node.ID == "" {
		node.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(node).Error
}

// GetDedicatedNode retrieves a dedicated node by ID
func (r *PostgresRepository) GetDedicatedNode(ctx context.Context, nodeID string) (*node.DedicatedNode, error) {
	var dedicatedNode node.DedicatedNode
	err := r.db.WithContext(ctx).Where("id = ?", nodeID).First(&dedicatedNode).Error
	if err != nil {
		return nil, err
	}
	return &dedicatedNode, nil
}

// GetDedicatedNodeByVMID retrieves a dedicated node by Proxmox VM ID
func (r *PostgresRepository) GetDedicatedNodeByVMID(ctx context.Context, vmid int) (*node.DedicatedNode, error) {
	var dedicatedNode node.DedicatedNode
	err := r.db.WithContext(ctx).Where("proxmox_vmid = ?", vmid).First(&dedicatedNode).Error
	if err != nil {
		return nil, err
	}
	return &dedicatedNode, nil
}

// ListDedicatedNodes lists all dedicated nodes for a workspace
func (r *PostgresRepository) ListDedicatedNodes(ctx context.Context, workspaceID string) ([]node.DedicatedNode, error) {
	var nodes []node.DedicatedNode
	err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// UpdateDedicatedNode updates a dedicated node
func (r *PostgresRepository) UpdateDedicatedNode(ctx context.Context, dedicatedNode *node.DedicatedNode) error {
	return r.db.WithContext(ctx).Save(dedicatedNode).Error
}

// DeleteDedicatedNode soft deletes a dedicated node
func (r *PostgresRepository) DeleteDedicatedNode(ctx context.Context, nodeID string) error {
	return r.db.WithContext(ctx).
		Model(&node.DedicatedNode{}).
		Where("id = ?", nodeID).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

// CreateNodeEvent creates a new node event
func (r *PostgresRepository) CreateNodeEvent(ctx context.Context, event *node.NodeEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(event).Error
}

// ListNodeEvents lists events for a node
func (r *PostgresRepository) ListNodeEvents(ctx context.Context, nodeID string, limit int) ([]node.NodeEvent, error) {
	var events []node.NodeEvent
	query := r.db.WithContext(ctx).
		Where("node_id = ?", nodeID).
		Order("timestamp DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

// GetWorkspaceAllocation gets the allocation record for a workspace
func (r *PostgresRepository) GetWorkspaceAllocation(ctx context.Context, workspaceID string) (*node.WorkspaceNodeAllocation, error) {
	var allocation node.WorkspaceNodeAllocation
	err := r.db.WithContext(ctx).
		Preload("DedicatedNodes").
		Where("workspace_id = ?", workspaceID).
		First(&allocation).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create default shared allocation if not exists
		allocation = node.WorkspaceNodeAllocation{
			ID:          uuid.New().String(),
			WorkspaceID: workspaceID,
			PlanType:    node.PlanTypeShared,
			SharedQuota: &node.SharedQuota{
				CPULimit:    2,
				MemoryLimit: 4,
				CPUUsed:     0,
				MemoryUsed:  0,
			},
		}
		err = r.db.WithContext(ctx).Create(&allocation).Error
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	
	return &allocation, nil
}

// CreateWorkspaceAllocation creates a new workspace allocation
func (r *PostgresRepository) CreateWorkspaceAllocation(ctx context.Context, allocation *node.WorkspaceNodeAllocation) error {
	if allocation.ID == "" {
		allocation.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(allocation).Error
}

// UpdateWorkspaceAllocation updates a workspace allocation
func (r *PostgresRepository) UpdateWorkspaceAllocation(ctx context.Context, allocation *node.WorkspaceNodeAllocation) error {
	return r.db.WithContext(ctx).Save(allocation).Error
}

// UpdateSharedQuotaUsage updates the shared quota usage for a workspace
func (r *PostgresRepository) UpdateSharedQuotaUsage(ctx context.Context, workspaceID string, cpu, memory float64) error {
	return r.db.WithContext(ctx).
		Model(&node.WorkspaceNodeAllocation{}).
		Where("workspace_id = ?", workspaceID).
		Updates(map[string]interface{}{
			"quota_cpu_used":    cpu,
			"quota_memory_used": memory,
		}).Error
}

// GetNodeResourceUsage gets resource usage for a specific node
func (r *PostgresRepository) GetNodeResourceUsage(ctx context.Context, nodeID string) (*node.ResourceUsage, error) {
	// In a real implementation, this would query metrics from monitoring system
	// For now, return mock data
	dedicatedNode, err := r.GetDedicatedNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	return &node.ResourceUsage{
		NodeID:    nodeID,
		CPUCores:  float64(dedicatedNode.Specification.CPUCores) * 0.7,  // 70% usage
		MemoryGB:  float64(dedicatedNode.Specification.MemoryGB) * 0.6, // 60% usage
		StorageGB: float64(dedicatedNode.Specification.StorageGB) * 0.5, // 50% usage
		PodCount:  12,
		Timestamp: dedicatedNode.UpdatedAt,
	}, nil
}