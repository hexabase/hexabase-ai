package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/hexabase/hexabase-ai/api/internal/project/domain"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	// Configure GORM to be more lenient with model validation
	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
		PrepareStmt:            false,
		DisableAutomaticPing:   true,
	})
	require.NoError(t, err)

	return gormDB, mock
}

// Test only methods that don't involve complex struct operations
func TestPostgresRepository_ResourceUsageMethods(t *testing.T) {
	gormDB, _ := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	t.Run("GetProjectResourceUsage returns placeholder data", func(t *testing.T) {
		projectID := "proj-123"

		usage, err := repo.GetProjectResourceUsage(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, usage)
		assert.Equal(t, "0", usage.CPU)
		assert.Equal(t, "0", usage.Memory)
		assert.Equal(t, 0, usage.Pods)
	})

	t.Run("GetNamespaceResourceUsage returns placeholder data", func(t *testing.T) {
		namespaceID := "ns-123"

		usage, err := repo.GetNamespaceResourceUsage(ctx, namespaceID)
		assert.NoError(t, err)
		assert.NotNil(t, usage)
		assert.Equal(t, "0", usage.CPU)
		assert.Equal(t, "0", usage.Memory)
		assert.Equal(t, "0", usage.Storage)
		assert.Equal(t, 0, usage.Pods)
	})
}

// Test repository creation
func TestNewPostgresRepository(t *testing.T) {
	gormDB, _ := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresRepository(gormDB)
	assert.NotNil(t, repo)
	
	// Verify it implements the interface
	var _ domain.Repository = repo
}

// Test delete operations that use simple SQL
func TestPostgresRepository_DeleteOperations(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresRepository(gormDB).(*postgresRepository)
	ctx := context.Background()

	t.Run("DeleteProject executes correct SQL", func(t *testing.T) {
		projectID := "proj-123"

		mock.ExpectExec(`DELETE FROM "projects" WHERE id = \$1`).
			WithArgs(projectID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteProject(ctx, projectID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteNamespace executes correct SQL", func(t *testing.T) {
		namespaceID := "ns-123"

		mock.ExpectExec(`DELETE FROM "namespaces" WHERE id = \$1`).
			WithArgs(namespaceID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteNamespace(ctx, namespaceID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RemoveMember executes correct SQL", func(t *testing.T) {
		memberID := "member-123"

		mock.ExpectExec(`DELETE FROM "project_members" WHERE id = \$1`).
			WithArgs(memberID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemoveMember(ctx, memberID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RemoveProjectMember executes correct SQL", func(t *testing.T) {
		projectID := "proj-123"
		userID := "user-123"

		mock.ExpectExec(`DELETE FROM "project_members" WHERE project_id = \$1 AND user_id = \$2`).
			WithArgs(projectID, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemoveProjectMember(ctx, projectID, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CleanupOldActivities executes correct SQL", func(t *testing.T) {
		before := time.Now().Add(-30 * 24 * time.Hour)

		mock.ExpectExec(`DELETE FROM "project_activities" WHERE created_at < \$1`).
			WithArgs(before).
			WillReturnResult(sqlmock.NewResult(0, 10))

		err := repo.CleanupOldActivities(ctx, before)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test count operations
func TestPostgresRepository_CountOperations(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	t.Run("CountProjects returns correct count", func(t *testing.T) {
		workspaceID := "ws-123"

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
		mock.ExpectQuery(`SELECT count\(\*\) FROM "projects" WHERE workspace_id = \$1`).
			WithArgs(workspaceID).
			WillReturnRows(countRows)

		count, err := repo.CountProjects(ctx, workspaceID)
		assert.NoError(t, err)
		assert.Equal(t, 10, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CountMembers returns correct count", func(t *testing.T) {
		projectID := "proj-123"

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery(`SELECT count\(\*\) FROM "project_members" WHERE project_id = \$1`).
			WithArgs(projectID).
			WillReturnRows(countRows)

		count, err := repo.CountMembers(ctx, projectID)
		assert.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Integration test template (commented out - requires actual database)
/*
func TestPostgresRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// and proper test data setup
	
	// Example:
	// db := setupTestDatabase(t)
	// defer cleanupTestDatabase(t, db)
	// 
	// repo := NewPostgresRepository(db)
	// 
	// Test actual CRUD operations with real data
}
*/

// Benchmark tests
func BenchmarkPostgresRepository_CountProjects(b *testing.B) {
	// For benchmarks, we'll use a simpler setup
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer mockDB.Close()

	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
		PrepareStmt:            false,
		DisableAutomaticPing:   true,
	})
	if err != nil {
		b.Fatal(err)
	}

	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()
	workspaceID := "ws-123"

	// Setup mock expectation that can be called multiple times
	for i := 0; i < b.N; i++ {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
		mock.ExpectQuery(`SELECT count\(\*\) FROM "projects" WHERE workspace_id = \$1`).
			WithArgs(workspaceID).
			WillReturnRows(countRows)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.CountProjects(ctx, workspaceID)
	}
}

// Example test to demonstrate usage
func ExampleNewPostgresRepository() {
	// This is an example of how to create and use the repository
	// In real usage, you would have a proper GORM database connection
	
	// db := yourGormDBConnection
	// repo := NewPostgresRepository(db)
	// 
	// // Example: Get project resource usage
	// usage, err := repo.GetProjectResourceUsage(context.Background(), "proj-123")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// 
	// fmt.Printf("CPU: %s, Memory: %s, Pods: %d\n", usage.CPU, usage.Memory, usage.Pods)
	// 
	// // Output: CPU: 0, Memory: 0, Pods: 0
}