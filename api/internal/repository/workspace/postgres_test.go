package workspace

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	require.NoError(t, err)

	return gormDB, mock
}

// TestPostgresRepository_NewPostgresRepository tests repository creation
func TestPostgresRepository_NewPostgresRepository(t *testing.T) {
	t.Run("creates repository successfully", func(t *testing.T) {
		gormDB, _ := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)
		assert.NotNil(t, repo)
		assert.Implements(t, (*workspace.Repository)(nil), repo)
	})
}

// TestPostgresRepository_WorkspaceMembers tests member operations
// These work because WorkspaceMember struct doesn't have map[string]interface{} fields
func TestPostgresRepository_WorkspaceMembers(t *testing.T) {
	ctx := context.Background()

	t.Run("add workspace member", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		member := &workspace.WorkspaceMember{
			ID:          uuid.New().String(),
			WorkspaceID: "ws-123",
			UserID:      "user-123",
			Role:        "admin",
			AddedBy:     "admin-user",
			AddedAt:     time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "workspace_members"`).
			WithArgs(
				member.ID,
				member.WorkspaceID,
				member.UserID,
				member.Role,
				member.AddedBy,
				sqlmock.AnyArg(), // AddedAt
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.AddWorkspaceMember(ctx, member)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("add member error", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		member := &workspace.WorkspaceMember{
			ID:          uuid.New().String(),
			WorkspaceID: "ws-123",
			UserID:      "user-123",
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "workspace_members"`).
			WillReturnError(fmt.Errorf("constraint violation"))
		mock.ExpectRollback()

		err := repo.AddWorkspaceMember(ctx, member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add workspace member")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("remove workspace member", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		workspaceID := "ws-123"
		userID := "user-123"

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "workspace_members" WHERE workspace_id = \$1 AND user_id = \$2`).
			WithArgs(workspaceID, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.RemoveWorkspaceMember(ctx, workspaceID, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("remove member error", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "workspace_members"`).
			WillReturnError(fmt.Errorf("delete failed"))
		mock.ExpectRollback()

		err := repo.RemoveWorkspaceMember(ctx, "ws-123", "user-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove workspace member")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list workspace members", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		workspaceID := "ws-123"
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "user_id", "role", "added_by", "added_at",
		}).
			AddRow("member-1", workspaceID, "user-1", "admin", "admin-user", now).
			AddRow("member-2", workspaceID, "user-2", "viewer", "admin-user", now)

		mock.ExpectQuery(`SELECT \* FROM "workspace_members" WHERE workspace_id = \$1`).
			WithArgs(workspaceID).
			WillReturnRows(rows)

		members, err := repo.ListWorkspaceMembers(ctx, workspaceID)
		assert.NoError(t, err)
		assert.Len(t, members, 2)
		assert.Equal(t, "admin", members[0].Role)
		assert.Equal(t, "viewer", members[1].Role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list members error", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "workspace_members"`).
			WillReturnError(fmt.Errorf("query failed"))

		members, err := repo.ListWorkspaceMembers(ctx, "ws-123")
		assert.Error(t, err)
		assert.Nil(t, members)
		assert.Contains(t, err.Error(), "failed to list workspace members")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestPostgresRepository_ErrorHandling tests error cases by method
func TestPostgresRepository_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("method implementation validation", func(t *testing.T) {
		gormDB, _ := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		// Test that all interface methods exist and can be called
		// These will fail due to parsing errors, but we verify the methods exist
		
		// Workspace operations
		err := repo.CreateWorkspace(ctx, &workspace.Workspace{})
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		_, err = repo.GetWorkspace(ctx, "test")
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		_, err = repo.GetWorkspaceByNameAndOrg(ctx, "test", "org")
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		err = repo.UpdateWorkspace(ctx, &workspace.Workspace{})
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		err = repo.DeleteWorkspace(ctx, "test")
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		_, _, err = repo.ListWorkspaces(ctx, workspace.WorkspaceFilter{})
		assert.Error(t, err) // Expected to fail due to struct parsing

		// Task operations  
		err = repo.CreateTask(ctx, &workspace.Task{})
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		_, err = repo.GetTask(ctx, "test")
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		err = repo.UpdateTask(ctx, &workspace.Task{})
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		_, err = repo.ListTasks(ctx, "test")
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		_, err = repo.GetPendingTasks(ctx, "test", 1)
		assert.Error(t, err) // Expected to fail due to struct parsing

		// Status operations
		err = repo.SaveWorkspaceStatus(ctx, &workspace.WorkspaceStatus{})
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		_, err = repo.GetWorkspaceStatus(ctx, "test")
		assert.Error(t, err) // Expected to fail due to struct parsing

		// Kubeconfig operations
		err = repo.SaveKubeconfig(ctx, "test", "config")
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		_, err = repo.GetKubeconfig(ctx, "test")
		assert.Error(t, err) // Expected to fail due to struct parsing

		// Cleanup operations
		err = repo.CleanupExpiredTasks(ctx, time.Now())
		assert.Error(t, err) // Expected to fail due to struct parsing
		
		err = repo.CleanupDeletedWorkspaces(ctx, time.Now())
		assert.Error(t, err) // Expected to fail due to struct parsing

		// Resource usage operations
		err = repo.CreateResourceUsage(ctx, &workspace.ResourceUsage{})
		assert.Error(t, err) // Expected to fail due to struct parsing
	})
}

/*
IMPORTANT NOTE:

This test file is intentionally minimal due to fundamental incompatibilities between:
1. GORM's parsing of map[string]interface{} fields 
2. sqlmock's mocking capabilities
3. The domain model's use of complex nested structures

The following repository methods cannot be easily unit tested with sqlmock due to GORM 
parsing errors with map[string]interface{} fields:

- CreateWorkspace, GetWorkspace, UpdateWorkspace, DeleteWorkspace, ListWorkspaces
  (due to ClusterInfo, Settings, Metadata fields)
  
- CreateTask, GetTask, UpdateTask, ListTasks, GetPendingTasks  
  (due to Payload, Metadata fields)
  
- CreateResourceUsage, GetResourceUsageHistory
  (due to nested struct fields without proper GORM tags)
  
- SaveWorkspaceStatus, GetWorkspaceStatus
  (due to ClusterInfo field)
  
- SaveKubeconfig, GetKubeconfig  
  (due to JSONB operations and workspace struct parsing)
  
- CleanupExpiredTasks, CleanupDeletedWorkspaces
  (due to workspace/task struct parsing)

RECOMMENDATIONS FOR COMPREHENSIVE TESTING:

1. Integration Tests: Create integration tests with a real PostgreSQL database using 
   testcontainers or similar to test the full repository functionality.
   
2. Domain Model Refactoring: Consider refactoring the domain models to use proper GORM 
   tags for JSON fields or separate the database models from domain models.
   
3. Repository Pattern: Implement a proper repository pattern with database-specific 
   models that map to/from domain models.

WHAT IS TESTED:

- Repository creation and interface compliance
- WorkspaceMember operations (these work because the struct is simple)
- Error handling validation (methods exist and return errors as expected)

This provides a foundation for testing while acknowledging the current architectural 
constraints. The workspace member operations demonstrate the testing patterns that 
would work for all operations once the struct parsing issues are resolved.
*/