package organization

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/organization"
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

// NOTE: The current implementation has a mismatch between domain models and database models.
// The domain.Organization struct contains fields like Settings (map[string]interface{})
// that GORM cannot handle directly. This causes the repository methods to fail.
// These tests demonstrate the issue and test what we can.

func TestPostgresRepository_DeleteOrganization(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "organizations" WHERE id = \?`).
			WithArgs(orgID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteOrganization(ctx, orgID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete non-existent organization", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "organizations" WHERE id = \?`).
			WithArgs(orgID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.DeleteOrganization(ctx, orgID)
		assert.NoError(t, err) // No error even if nothing was deleted
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_MemberOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("add member successfully", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		member := &organization.OrganizationUser{
			ID:             uuid.New().String(),
			OrganizationID: uuid.New().String(),
			UserID:         uuid.New().String(),
			Email:          "member@example.com",
			Role:           "member",
			Status:         "active",
			JoinedAt:       time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "organization_users"`).
			WithArgs(
				member.ID,
				member.OrganizationID,
				member.UserID,
				member.Email,
				member.Role,
				member.InvitedBy,
				sqlmock.AnyArg(), // invited_at
				sqlmock.AnyArg(), // joined_at
				member.Status,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.AddMember(ctx, member)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update member role", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		userID := uuid.New().String()
		newRole := "admin"

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "organization_users" SET "role"=\? WHERE organization_id = \? AND user_id = \?`).
			WithArgs(newRole, orgID, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateMemberRole(ctx, orgID, userID, newRole)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("remove member", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		userID := uuid.New().String()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "organization_users" WHERE organization_id = \? AND user_id = \?`).
			WithArgs(orgID, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.RemoveMember(ctx, orgID, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count members", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery(`SELECT count\(\*\) FROM "organization_users" WHERE organization_id = \? AND status = \?`).
			WithArgs(orgID, "active").
			WillReturnRows(countRows)

		count, err := repo.CountMembers(ctx, orgID)
		assert.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update member", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		member := &organization.OrganizationUser{
			ID:             uuid.New().String(),
			OrganizationID: uuid.New().String(),
			UserID:         uuid.New().String(),
			Status:         "suspended",
		}

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "organization_users" SET`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateMember(ctx, member)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_InvitationOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("create invitation", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		invitation := &organization.Invitation{
			ID:             uuid.New().String(),
			OrganizationID: uuid.New().String(),
			Email:          "invitee@example.com",
			Role:           "member",
			Token:          uuid.New().String(),
			InvitedBy:      uuid.New().String(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			CreatedAt:      time.Now(),
			Status:         "pending",
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "invitations"`).
			WithArgs(
				invitation.ID,
				invitation.OrganizationID,
				invitation.Email,
				invitation.Role,
				invitation.Token,
				invitation.InvitedBy,
				sqlmock.AnyArg(), // expires_at
				sqlmock.AnyArg(), // created_at
				nil,              // accepted_at
				invitation.Status,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateInvitation(ctx, invitation)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update invitation", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		now := time.Now()
		invitation := &organization.Invitation{
			ID:         uuid.New().String(),
			Status:     "accepted",
			AcceptedAt: &now,
		}

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "invitations" SET`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateInvitation(ctx, invitation)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete invitation", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		invitationID := uuid.New().String()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "invitations" WHERE id = \?`).
			WithArgs(invitationID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteInvitation(ctx, invitationID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete expired invitations", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		before := time.Now()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "invitations" WHERE expires_at < \? AND status = \?`).
			WithArgs(before, "pending").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		err := repo.DeleteExpiredInvitations(ctx, before)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_ActivityOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("create activity", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		activity := &organization.Activity{
			ID:             uuid.New().String(),
			OrganizationID: uuid.New().String(),
			UserID:         uuid.New().String(),
			Type:           "member",
			Action:         "added",
			ResourceType:   "organization_user",
			ResourceID:     uuid.New().String(),
			Timestamp:      time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "activities"`).
			WithArgs(
				activity.ID,
				activity.OrganizationID,
				activity.UserID,
				activity.Type,
				activity.Action,
				activity.ResourceType,
				activity.ResourceID,
				sqlmock.AnyArg(), // details
				sqlmock.AnyArg(), // timestamp
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateActivity(ctx, activity)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list activities with filters", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		userID := uuid.New().String()
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()
		filter := organization.ActivityFilter{
			OrganizationID: orgID,
			UserID:         userID,
			Type:           "member",
			StartDate:      &startDate,
			EndDate:        &endDate,
			Limit:          10,
		}

		rows := sqlmock.NewRows([]string{
			"id", "organization_id", "user_id", "type", "action",
			"resource_type", "resource_id", "details", "timestamp",
		}).
			AddRow(uuid.New().String(), orgID, userID, "member", "added",
				"organization_user", uuid.New().String(), nil, time.Now()).
			AddRow(uuid.New().String(), orgID, userID, "member", "role_updated",
				"organization_user", uuid.New().String(), nil, time.Now())

		mock.ExpectQuery(`SELECT \* FROM "activities" WHERE organization_id = \? AND user_id = \? AND type = \? AND timestamp >= \? AND timestamp <= \?`).
			WithArgs(orgID, userID, "member", sqlmock.AnyArg(), sqlmock.AnyArg(), 10).
			WillReturnRows(rows)

		activities, err := repo.ListActivities(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, activities, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_StatisticsOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("get organization stats", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		// Count total members
		mock.ExpectQuery(`SELECT count\(\*\) FROM "organization_users" WHERE organization_id = \?`).
			WithArgs(orgID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		// Count active members
		mock.ExpectQuery(`SELECT count\(\*\) FROM "organization_users" WHERE organization_id = \? AND status = \?`).
			WithArgs(orgID, "active").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))

		// Count total workspaces
		mock.ExpectQuery(`SELECT count\(\*\) FROM "workspaces" WHERE organization_id = \?`).
			WithArgs(orgID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		// Count active workspaces
		mock.ExpectQuery(`SELECT count\(\*\) FROM "workspaces" WHERE organization_id = \? AND status = \?`).
			WithArgs(orgID, "active").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))

		// Count projects
		mock.ExpectQuery(`SELECT count\(\*\) FROM "projects" JOIN workspaces ON projects.workspace_id = workspaces.id WHERE workspaces.organization_id = \?`).
			WithArgs(orgID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))

		stats, err := repo.GetOrganizationStats(ctx, orgID)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, orgID, stats.OrganizationID)
		assert.Equal(t, 10, stats.TotalMembers)
		assert.Equal(t, 8, stats.ActiveMembers)
		assert.Equal(t, 5, stats.TotalWorkspaces)
		assert.Equal(t, 4, stats.ActiveWorkspaces)
		assert.Equal(t, 12, stats.TotalProjects)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get workspace count", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		// Total count
		mock.ExpectQuery(`SELECT count\(\*\) FROM "workspaces" WHERE organization_id = \?`).
			WithArgs(orgID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		// Active count
		mock.ExpectQuery(`SELECT count\(\*\) FROM "workspaces" WHERE organization_id = \? AND status = \?`).
			WithArgs(orgID, "active").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))

		total, active, err := repo.GetWorkspaceCount(ctx, orgID)
		assert.NoError(t, err)
		assert.Equal(t, 10, total)
		assert.Equal(t, 8, active)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get project count", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		rows := sqlmock.NewRows([]string{"count"}).AddRow(15)
		mock.ExpectQuery(`SELECT count\(\*\) FROM "projects" JOIN workspaces ON projects.workspace_id = workspaces.id WHERE workspaces.organization_id = \?`).
			WithArgs(orgID).
			WillReturnRows(rows)

		count, err := repo.GetProjectCount(ctx, orgID)
		assert.NoError(t, err)
		assert.Equal(t, 15, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get resource usage", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		// Currently returns placeholder data
		usage, err := repo.GetResourceUsage(ctx, orgID)
		assert.NoError(t, err)
		assert.NotNil(t, usage)
		assert.Equal(t, float64(0), usage.CPU)
		assert.Equal(t, float64(0), usage.Memory)
		assert.Equal(t, float64(0), usage.Storage)
		assert.Equal(t, float64(0), usage.Cost)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list workspaces", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		rows := sqlmock.NewRows([]string{"id", "name", "status"}).
			AddRow(uuid.New().String(), "workspace-1", "active").
			AddRow(uuid.New().String(), "workspace-2", "active").
			AddRow(uuid.New().String(), "workspace-3", "suspended")

		mock.ExpectQuery(`SELECT id, name, status FROM "workspaces" WHERE organization_id = \?`).
			WithArgs(orgID).
			WillReturnRows(rows)

		workspaces, err := repo.ListWorkspaces(ctx, orgID)
		assert.NoError(t, err)
		assert.Len(t, workspaces, 3)
		assert.Equal(t, "workspace-1", workspaces[0].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Edge cases and error scenarios
func TestPostgresRepository_ErrorScenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("database connection error on delete", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "organizations"`).
			WithArgs(orgID).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.DeleteOrganization(ctx, orgID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete organization")
	})

	t.Run("member not found error", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		userID := uuid.New().String()

		mock.ExpectQuery(`SELECT \* FROM "organization_users" WHERE organization_id = \? AND user_id = \?`).
			WithArgs(orgID, userID).
			WillReturnError(gorm.ErrRecordNotFound)

		member, err := repo.GetMember(ctx, orgID, userID)
		assert.NoError(t, err) // Returns nil, no error
		assert.Nil(t, member)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invitation not found error", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		invitationID := uuid.New().String()

		mock.ExpectQuery(`SELECT \* FROM "invitations" WHERE id = \?`).
			WithArgs(invitationID).
			WillReturnError(gorm.ErrRecordNotFound)

		invitation, err := repo.GetInvitation(ctx, invitationID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invitation not found")
		assert.Nil(t, invitation)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test complex queries
func TestPostgresRepository_ComplexQueries(t *testing.T) {
	ctx := context.Background()

	t.Run("list invitations with filters", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		status := "pending"
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "organization_id", "email", "role", "token",
			"invited_by", "expires_at", "created_at", "accepted_at", "status",
		}).
			AddRow(uuid.New().String(), orgID, "user1@example.com", "member", uuid.New().String(),
				uuid.New().String(), now.Add(24*time.Hour), now, nil, status).
			AddRow(uuid.New().String(), orgID, "user2@example.com", "admin", uuid.New().String(),
				uuid.New().String(), now.Add(24*time.Hour), now, nil, status)

		mock.ExpectQuery(`SELECT \* FROM "invitations" WHERE organization_id = \? AND status = \? ORDER BY created_at DESC`).
			WithArgs(orgID, status).
			WillReturnRows(rows)

		invitations, err := repo.ListInvitations(ctx, orgID, status)
		assert.NoError(t, err)
		assert.Len(t, invitations, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list invitations without status filter", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "organization_id", "email", "role", "token",
			"invited_by", "expires_at", "created_at", "accepted_at", "status",
		}).
			AddRow(uuid.New().String(), orgID, "user1@example.com", "member", uuid.New().String(),
				uuid.New().String(), now.Add(24*time.Hour), now, nil, "pending").
			AddRow(uuid.New().String(), orgID, "user2@example.com", "admin", uuid.New().String(),
				uuid.New().String(), now.Add(24*time.Hour), now.Add(-1*time.Hour), &now, "accepted")

		mock.ExpectQuery(`SELECT \* FROM "invitations" WHERE organization_id = \? ORDER BY created_at DESC`).
			WithArgs(orgID).
			WillReturnRows(rows)

		invitations, err := repo.ListInvitations(ctx, orgID, "")
		assert.NoError(t, err)
		assert.Len(t, invitations, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test User operations
func TestPostgresRepository_UserOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("get user by ID", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "email", "display_name", "provider", "external_id",
			"created_at", "updated_at", "last_login_at",
		}).AddRow(
			userID, "user@example.com", "Test User", "google", "google-123",
			now, now, &now,
		)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \?`).
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "user@example.com", user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get user by email", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		email := "user@example.com"
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "email", "display_name", "provider", "external_id",
			"created_at", "updated_at", "last_login_at",
		}).AddRow(
			uuid.New().String(), email, "Test User", "google", "google-123",
			now, now, nil,
		)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \?`).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.GetUserByEmail(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get users by multiple IDs", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		userID1 := uuid.New().String()
		userID2 := uuid.New().String()
		userIDs := []string{userID1, userID2}
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "email", "display_name", "provider", "external_id",
			"created_at", "updated_at", "last_login_at",
		}).
			AddRow(userID1, "user1@example.com", "User 1", "google", "google-1", now, now, nil).
			AddRow(userID2, "user2@example.com", "User 2", "github", "github-2", now, now, nil)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE id IN \(\?,\?\)`).
			WithArgs(userID1, userID2).
			WillReturnRows(rows)

		users, err := repo.GetUsersByIDs(ctx, userIDs)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		userID := uuid.New().String()

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \?`).
			WithArgs(userID).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUser(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user by email not found returns nil", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		email := "nonexistent@example.com"

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \?`).
			WithArgs(email).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUserByEmail(ctx, email)
		assert.NoError(t, err) // Returns nil, no error
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test GetMember method
func TestPostgresRepository_GetMember(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get member", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "organization_id", "user_id", "email", "role",
			"invited_by", "invited_at", "joined_at", "status",
		}).AddRow(
			uuid.New().String(), orgID, userID, "member@example.com", "admin",
			uuid.New().String(), now, now, "active",
		)

		mock.ExpectQuery(`SELECT \* FROM "organization_users" WHERE organization_id = \? AND user_id = \?`).
			WithArgs(orgID, userID).
			WillReturnRows(rows)

		member, err := repo.GetMember(ctx, orgID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, member)
		assert.Equal(t, userID, member.UserID)
		assert.Equal(t, orgID, member.OrganizationID)
		assert.Equal(t, "admin", member.Role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test GetInvitation methods
func TestPostgresRepository_GetInvitation(t *testing.T) {
	ctx := context.Background()

	t.Run("get invitation by ID", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		invitationID := uuid.New().String()
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)

		rows := sqlmock.NewRows([]string{
			"id", "organization_id", "email", "role", "token",
			"invited_by", "expires_at", "created_at", "accepted_at", "status",
		}).AddRow(
			invitationID, uuid.New().String(), "invitee@example.com", "member", uuid.New().String(),
			uuid.New().String(), expiresAt, now, nil, "pending",
		)

		mock.ExpectQuery(`SELECT \* FROM "invitations" WHERE id = \?`).
			WithArgs(invitationID).
			WillReturnRows(rows)

		invitation, err := repo.GetInvitation(ctx, invitationID)
		assert.NoError(t, err)
		assert.NotNil(t, invitation)
		assert.Equal(t, invitationID, invitation.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get invitation by token", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		token := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "organization_id", "email", "role", "token",
			"invited_by", "expires_at", "created_at", "accepted_at", "status",
		}).AddRow(
			uuid.New().String(), uuid.New().String(), "invitee@example.com", "member", token,
			uuid.New().String(), now.Add(24*time.Hour), now, nil, "pending",
		)

		mock.ExpectQuery(`SELECT \* FROM "invitations" WHERE token = \?`).
			WithArgs(token).
			WillReturnRows(rows)

		invitation, err := repo.GetInvitationByToken(ctx, token)
		assert.NoError(t, err)
		assert.NotNil(t, invitation)
		assert.Equal(t, token, invitation.Token)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invitation by token not found returns nil", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		token := uuid.New().String()

		mock.ExpectQuery(`SELECT \* FROM "invitations" WHERE token = \?`).
			WithArgs(token).
			WillReturnError(gorm.ErrRecordNotFound)

		invitation, err := repo.GetInvitationByToken(ctx, token)
		assert.NoError(t, err) // Returns nil, no error
		assert.Nil(t, invitation)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test coverage for methods that would fail due to model mismatch
func TestPostgresRepository_ModelMismatchIssues(t *testing.T) {
	ctx := context.Background()

	t.Run("create organization fails due to model mismatch", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		org := &organization.Organization{
			ID:          uuid.New().String(),
			Name:        "test-org",
			DisplayName: "Test Organization",
			Status:      "active",
			Settings:    map[string]interface{}{"key": "value"}, // This causes issues
		}

		// GORM will fail to parse the Settings field
		mock.ExpectBegin()
		mock.ExpectRollback()

		err := repo.CreateOrganization(ctx, org)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create organization")
	})

	t.Run("update organization with incompatible fields", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		org := &organization.Organization{
			ID:               uuid.New().String(),
			Name:             "test-org",
			SubscriptionInfo: &organization.SubscriptionInfo{PlanID: "premium"}, // This causes issues
		}

		mock.ExpectBegin()
		mock.ExpectRollback()

		err := repo.UpdateOrganization(ctx, org)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update organization")
	})

	// NOTE: GetOrganization, GetOrganizationByName, and ListOrganizations
	// would also fail due to the model mismatch when GORM tries to scan
	// the results into the domain model. The repository implementation
	// needs to be updated to handle the mapping between DB and domain models.
}

// Test ListMembers and ListOrganizations with pagination
func TestPostgresRepository_PaginationAndFiltering(t *testing.T) {
	ctx := context.Background()

	t.Run("list members with pagination", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		filter := organization.MemberFilter{
			OrganizationID: orgID,
			Page:           2,
			PageSize:       5,
		}

		// Count query
		mock.ExpectQuery(`SELECT count\(\*\) FROM "organization_users" WHERE organization_id = \?`).
			WithArgs(orgID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))

		// Data query with offset
		now := time.Now()
		dataRows := sqlmock.NewRows([]string{
			"id", "organization_id", "user_id", "email", "role",
			"invited_by", "invited_at", "joined_at", "status",
		})
		
		// Add 5 rows for page 2
		for i := 0; i < 5; i++ {
			dataRows.AddRow(
				uuid.New().String(), orgID, uuid.New().String(),
				fmt.Sprintf("user%d@example.com", i+6), "member", "", now, now, "active",
			)
		}

		mock.ExpectQuery(`SELECT \* FROM "organization_users" WHERE organization_id = \?`).
			WithArgs(orgID, 5, 5). // offset=5, limit=5
			WillReturnRows(dataRows)

		members, total, err := repo.ListMembers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 12, total)
		assert.Len(t, members, 5)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list members with search filter", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		filter := organization.MemberFilter{
			OrganizationID: orgID,
			Search:         "john",
			Page:           1,
			PageSize:       10,
		}

		// Count query with join
		mock.ExpectQuery(`SELECT count\(\*\) FROM "organization_users" JOIN users`).
			WithArgs(orgID, "%john%", "%john%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Data query with join
		now := time.Now()
		dataRows := sqlmock.NewRows([]string{
			"id", "organization_id", "user_id", "email", "role",
			"invited_by", "invited_at", "joined_at", "status",
		}).
			AddRow(uuid.New().String(), orgID, uuid.New().String(), "john.doe@example.com", "member", "", now, now, "active").
			AddRow(uuid.New().String(), orgID, uuid.New().String(), "johnny@example.com", "admin", "", now, now, "active")

		mock.ExpectQuery(`SELECT .* FROM "organization_users" JOIN users`).
			WithArgs(orgID, "%john%", "%john%", 10).
			WillReturnRows(dataRows)

		members, total, err := repo.ListMembers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, members, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}