package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/node/domain"
	"github.com/hexabase/hexabase-ai/api/internal/node/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestPostgresRepository_CreateDedicatedNode(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("successful node creation", func(t *testing.T) {
		n := &domain.DedicatedNode{
			ID:          uuid.New().String(),
			WorkspaceID: "ws-123",
			Name:        "worker-node-1",
			Status:      domain.NodeStatusProvisioning,
			Specification: domain.NodeSpecification{
				Type:      "S-Type",
				CPUCores:  4,
				MemoryGB:  8,
				StorageGB: 100,
			},
			ProxmoxVMID: 1001,
			ProxmoxNode: "pve-node-1",
			IPAddress:   "10.0.1.100",
			Labels: map[string]string{
				"zone":        "us-east-1a",
				"environment": "production",
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "dedicated_nodes"`).
			WithArgs(
				n.ID,
				n.WorkspaceID,
				n.Name,
				n.Status,
				n.Specification.Type,
				n.Specification.CPUCores,
				n.Specification.MemoryGB,
				n.Specification.StorageGB,
				n.Specification.NetworkMbps,
				n.ProxmoxVMID,
				n.ProxmoxNode,
				n.IPAddress,
				n.K3sAgentVersion,
				n.SSHPublicKey,
				sqlmock.AnyArg(), // labels (JSON)
				sqlmock.AnyArg(), // taints (JSON)
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				n.DeletedAt,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateDedicatedNode(ctx, n)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetDedicatedNode(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("get node by ID", func(t *testing.T) {
		nodeID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "name", "status", 
			"spec_type", "spec_cpu_cores", "spec_memory_gb", "spec_storage_gb", "spec_network_mbps",
			"proxmox_vmid", "proxmox_node", "ip_address", "k3s_agent_version", "ssh_public_key",
			"labels", "taints", "created_at", "updated_at", "deleted_at",
		}).AddRow(
			nodeID, "ws-123", "master-1", "ready",
			"M-Type", 8, 16, 500, 1000,
			1000, "pve-node-1", "10.0.1.10", "v1.28.0", "ssh-rsa ...",
			`{"role":"master"}`, `[]`, now, now, nil,
		)

		mock.ExpectQuery(`SELECT \* FROM "dedicated_nodes" WHERE id = \$1 AND deleted_at IS NULL ORDER BY "dedicated_nodes"\."id" LIMIT \$2`).
			WithArgs(nodeID, 1).
			WillReturnRows(rows)

		n, err := repo.GetDedicatedNode(ctx, nodeID)
		assert.NoError(t, err)
		assert.NotNil(t, n)
		assert.Equal(t, nodeID, n.ID)
		assert.Equal(t, domain.NodeStatusReady, n.Status)
		assert.Equal(t, 8, n.Specification.CPUCores)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("node not found", func(t *testing.T) {
		nodeID := uuid.New().String()

		mock.ExpectQuery(`SELECT \* FROM "dedicated_nodes" WHERE id = \$1 AND deleted_at IS NULL ORDER BY "dedicated_nodes"\."id" LIMIT \$2`).
			WithArgs(nodeID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		n, err := repo.GetDedicatedNode(ctx, nodeID)
		assert.Error(t, err)
		assert.Nil(t, n)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_ListDedicatedNodes(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("list workspace nodes", func(t *testing.T) {
		workspaceID := "ws-456"
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "name", "status",
			"spec_type", "spec_cpu_cores", "spec_memory_gb", "spec_storage_gb", "spec_network_mbps",
			"proxmox_vmid", "proxmox_node", "ip_address", "k3s_agent_version", "ssh_public_key",
			"labels", "taints", "created_at", "updated_at", "deleted_at",
		}).
			AddRow(
				uuid.New().String(), workspaceID, "master-1", "ready",
				"M-Type", 8, 16, 500, 1000,
				1000, "pve-node-1", "10.0.1.10", "v1.28.0", "",
				`{}`, `[]`, now, now, nil,
			).
			AddRow(
				uuid.New().String(), workspaceID, "worker-1", "ready",
				"S-Type", 4, 8, 200, 500,
				1001, "pve-node-1", "10.0.1.11", "v1.28.0", "",
				`{}`, `[]`, now, now, nil,
			).
			AddRow(
				uuid.New().String(), workspaceID, "worker-2", "provisioning",
				"S-Type", 4, 8, 200, 500,
				1002, "pve-node-2", "10.0.1.12", "", "",
				`{}`, `[]`, now, now, nil,
			)

		mock.ExpectQuery(`SELECT \* FROM "dedicated_nodes" WHERE workspace_id = \$1 AND deleted_at IS NULL ORDER BY created_at DESC`).
			WithArgs(workspaceID).
			WillReturnRows(rows)

		nodes, err := repo.ListDedicatedNodes(ctx, workspaceID)
		assert.NoError(t, err)
		assert.Len(t, nodes, 3)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_UpdateDedicatedNode(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("update node status", func(t *testing.T) {
		n := &domain.DedicatedNode{
			ID:          uuid.New().String(),
			WorkspaceID: "ws-789",
			Name:        "updated-worker",
			Status:      domain.NodeStatusReady,
			Specification: domain.NodeSpecification{
				Type:      "L-Type",
				CPUCores:  16,
				MemoryGB:  32,
				StorageGB: 1000,
			},
			ProxmoxVMID: 2000,
			ProxmoxNode: "pve-node-2",
			IPAddress:   "10.0.2.100",
			Labels: map[string]string{
				"zone": "us-west-2b",
				"tier": "premium",
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "dedicated_nodes" SET`).
			WithArgs(
				n.WorkspaceID,
				n.Name,
				n.Status,
				n.Specification.Type,
				n.Specification.CPUCores,
				n.Specification.MemoryGB,
				n.Specification.StorageGB,
				n.Specification.NetworkMbps,
				n.ProxmoxVMID,
				n.ProxmoxNode,
				n.IPAddress,
				n.K3sAgentVersion,
				n.SSHPublicKey,
				sqlmock.AnyArg(), // labels
				sqlmock.AnyArg(), // taints
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				sqlmock.AnyArg(), // deleted_at
				n.ID,             // WHERE id = ?
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateDedicatedNode(ctx, n)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_DeleteDedicatedNode(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("soft delete node", func(t *testing.T) {
		nodeID := uuid.New().String()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "dedicated_nodes" SET "deleted_at"=NOW\(\),"updated_at"=\$1 WHERE id = \$2`).
			WithArgs(sqlmock.AnyArg(), nodeID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.DeleteDedicatedNode(ctx, nodeID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetDedicatedNodeByVMID(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("get node by proxmox VM ID", func(t *testing.T) {
		vmID := 2001
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "name", "status",
			"spec_type", "spec_cpu_cores", "spec_memory_gb", "spec_storage_gb", "spec_network_mbps",
			"proxmox_vmid", "proxmox_node", "ip_address", "k3s_agent_version", "ssh_public_key",
			"labels", "taints", "created_at", "updated_at", "deleted_at",
		}).AddRow(
			uuid.New().String(), "ws-999", "worker-special", "ready",
			"M-Type", 8, 16, 500, 1000,
			vmID, "pve-node-3", "10.0.3.50", "v1.28.0", "",
			`{}`, `[]`, now, now, nil,
		)

		mock.ExpectQuery(`SELECT \* FROM "dedicated_nodes" WHERE proxmox_vmid = \$1 AND deleted_at IS NULL ORDER BY "dedicated_nodes"\."id" LIMIT \$2`).
			WithArgs(vmID, 1).
			WillReturnRows(rows)

		n, err := repo.GetDedicatedNodeByVMID(ctx, vmID)
		assert.NoError(t, err)
		assert.NotNil(t, n)
		assert.Equal(t, vmID, n.ProxmoxVMID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_CreateNodeEvent(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("create node event", func(t *testing.T) {
		event := &domain.NodeEvent{
			ID:          uuid.New().String(),
			NodeID:      "node-123",
			WorkspaceID: "ws-123",
			Type:        domain.EventTypeStatusChange,
			Message:     "Node status changed from provisioning to ready",
			Details:     "Node initialization completed successfully",
			Timestamp:   time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "node_events"`).
			WithArgs(
				event.ID,
				event.NodeID,
				event.WorkspaceID,
				event.Type,
				event.Message,
				event.Details,
				sqlmock.AnyArg(), // timestamp
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateNodeEvent(ctx, event)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_ListNodeEvents(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("list node events", func(t *testing.T) {
		nodeID := "node-456"
		limit := 20
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "node_id", "workspace_id", "type", "message", "details", "timestamp",
		}).
			AddRow(uuid.New().String(), nodeID, "ws-456", "status_change", "Node started", "", now.Add(-10*time.Minute)).
			AddRow(uuid.New().String(), nodeID, "ws-456", "error", "Health check failed", "Connection timeout", now.Add(-5*time.Minute))

		mock.ExpectQuery(`SELECT \* FROM "node_events" WHERE node_id = \$1 ORDER BY timestamp DESC LIMIT \$2`).
			WithArgs(nodeID, limit).
			WillReturnRows(rows)

		events, err := repo.ListNodeEvents(ctx, nodeID, limit)
		assert.NoError(t, err)
		assert.Len(t, events, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetWorkspaceAllocation(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("get shared plan allocation", func(t *testing.T) {
		workspaceID := "ws-shared"
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "plan_type",
			"quota_cpu_limit", "quota_memory_limit", "quota_cpu_used", "quota_memory_used",
			"created_at", "updated_at",
		}).AddRow(
			uuid.New().String(), workspaceID, domain.PlanTypeShared,
			2.0, 4.0, 0.5, 1.0,
			now, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "workspace_node_allocations" WHERE workspace_id = \$1 ORDER BY "workspace_node_allocations"\."id" LIMIT \$2`).
			WithArgs(workspaceID, 1).
			WillReturnRows(rows)

		allocation, err := repo.GetWorkspaceAllocation(ctx, workspaceID)
		assert.NoError(t, err)
		assert.NotNil(t, allocation)
		assert.Equal(t, domain.PlanTypeShared, allocation.PlanType)
		assert.NotNil(t, allocation.SharedQuota)
		assert.Equal(t, 2.0, allocation.SharedQuota.CPULimit)
		assert.Equal(t, 0.5, allocation.SharedQuota.CPUUsed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_UpdateSharedQuotaUsage(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := repository.NewPostgresRepository(gormDB)

	t.Run("update shared plan usage", func(t *testing.T) {
		workspaceID := "ws-update"
		cpuDelta := 0.5
		memoryDelta := 1.0

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "workspace_node_allocations" SET "quota_cpu_used"=quota_cpu_used \+ \$1,"quota_memory_used"=quota_memory_used \+ \$2,"updated_at"=\$3 WHERE workspace_id = \$4 AND plan_type = \$5`).
			WithArgs(cpuDelta, memoryDelta, sqlmock.AnyArg(), workspaceID, domain.PlanTypeShared).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateSharedQuotaUsage(ctx, workspaceID, cpuDelta, memoryDelta)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}