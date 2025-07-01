package repository

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/organization/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
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
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "organizations" WHERE id = $1`)).
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
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "organizations" WHERE id = $1`)).
			WithArgs(orgID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.DeleteOrganization(ctx, orgID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_MemberOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("add member successfully", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		member := &domain.OrganizationUser{
			OrganizationID: uuid.New().String(),
			UserID:         uuid.New().String(),
			Role:           "member",
			JoinedAt:       time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "organization_users" ("organization_id","user_id","role","joined_at") VALUES ($1,$2,$3,$4) RETURNING "joined_at"`)).
			WithArgs(
				member.OrganizationID,
				member.UserID,
				member.Role,
				sqlmock.AnyArg(), // joined_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"joined_at"}).AddRow(member.JoinedAt))
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
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organization_users" SET "role"=$1 WHERE organization_id = $2 AND user_id = $3`)).
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
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "organization_users" WHERE organization_id = $1 AND user_id = $2`)).
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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organization_users" WHERE organization_id = $1 AND status = $2`)).
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

		member := &domain.OrganizationUser{
			OrganizationID: uuid.New().String(),
			UserID:         uuid.New().String(),
			Email:          "test@example.com",
			Role:           "member",
			InvitedBy:      "admin",
			InvitedAt:      time.Now(),
			JoinedAt:       time.Now(),
			Status:         "suspended",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organization_users" SET "email"=$1,"role"=$2,"invited_by"=$3,"invited_at"=$4,"joined_at"=$5,"status"=$6 WHERE "organization_id" = $7 AND "user_id" = $8`)).
			WithArgs(
				member.Email,
				member.Role,
				member.InvitedBy,
				sqlmock.AnyArg(), // invited_at
				sqlmock.AnyArg(), // joined_at
				member.Status,
				member.OrganizationID,
				member.UserID,
			).
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

		invitation := &domain.Invitation{
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
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "invitations" ("id","organization_id","email","role","token","invited_by","expires_at","created_at","accepted_at","status") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`)).
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

		invitation := &domain.Invitation{
			ID:             uuid.New().String(),
			OrganizationID: uuid.New().String(),
			Email:          "invitee@example.com",
			Role:           "admin",
			Token:          uuid.New().String(),
			InvitedBy:      uuid.New().String(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			CreatedAt:      time.Now(),
			Status:         "accepted",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "invitations" SET "organization_id"=$1,"email"=$2,"role"=$3,"token"=$4,"invited_by"=$5,"expires_at"=$6,"created_at"=$7,"accepted_at"=$8,"status"=$9 WHERE "id" = $10`)).
			WithArgs(
				invitation.OrganizationID,
				invitation.Email,
				invitation.Role,
				invitation.Token,
				invitation.InvitedBy,
				sqlmock.AnyArg(), // expires_at
				sqlmock.AnyArg(), // created_at
				invitation.AcceptedAt,
				invitation.Status,
				invitation.ID,
			).
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
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "invitations" WHERE id = $1`)).
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
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "invitations" WHERE expires_at < $1 AND status = $2`)).
			WithArgs(sqlmock.AnyArg(), "pending").
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

		activity := &domain.Activity{
			ID:             uuid.New().String(),
			OrganizationID: uuid.New().String(),
			UserID:         uuid.New().String(),
			Type:           "member",
			Action:         "added",
			ResourceType:   "organization_user",
			ResourceID:     uuid.New().String(),
			Timestamp:      time.Now(),
		}

		// Use helper method to set details
		detailsErr := activity.SetDetailsFromMap(map[string]interface{}{
			"role": "member",
		})
		require.NoError(t, detailsErr)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "activities" ("id","organization_id","user_id","type","action","resource_type","resource_id","details","timestamp") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`)).
			WithArgs(
				activity.ID,
				activity.OrganizationID,
				activity.UserID,
				activity.Type,
				activity.Action,
				activity.ResourceType,
				activity.ResourceID,
				`{"role":"member"}`,
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

		filter := domain.ActivityFilter{
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
				"organization_user", uuid.New().String(), `{"role": "member"}`, time.Now()).
			AddRow(uuid.New().String(), orgID, userID, "member", "role_updated",
				"organization_user", uuid.New().String(), `{"old_role": "member", "new_role": "admin"}`, time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "activities" WHERE organization_id = $1 AND user_id = $2 AND type = $3 AND timestamp >= $4 AND timestamp <= $5 ORDER BY timestamp DESC LIMIT $6`)).
			WithArgs(orgID, userID, "member", sqlmock.AnyArg(), sqlmock.AnyArg(), 10).
			WillReturnRows(rows)

		activities, err := repo.ListActivities(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, activities, 2)

		// Verify details can be parsed using helper methods
		firstActivityDetails, err := activities[0].GetDetailsAsMap()
		assert.NoError(t, err)
		assert.Equal(t, "member", firstActivityDetails["role"])

		secondActivityDetails, err := activities[1].GetDetailsAsMap()
		assert.NoError(t, err)
		assert.Equal(t, "member", secondActivityDetails["old_role"])
		assert.Equal(t, "admin", secondActivityDetails["new_role"])

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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organization_users" WHERE organization_id = $1`)).
			WithArgs(orgID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		// Count active members
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organization_users" WHERE organization_id = $1 AND status = $2`)).
			WithArgs(orgID, "active").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))

		// Count total workspaces
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "workspaces" WHERE organization_id = $1`)).
			WithArgs(orgID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		// Count active workspaces
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "workspaces" WHERE organization_id = $1 AND v_cluster_status = $2`)).
			WithArgs(orgID, "RUNNING").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))

		// Count projects
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "projects" JOIN workspaces ON projects.workspace_id = workspaces.id WHERE workspaces.organization_id = $1`)).
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

	t.Run("get workspace count with v_cluster_status", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		// Total count
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "workspaces" WHERE organization_id = $1`)).
			WithArgs(orgID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		// Active count with correct column name
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "workspaces" WHERE organization_id = $1 AND v_cluster_status = $2`)).
			WithArgs(orgID, "RUNNING").
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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "projects" JOIN workspaces ON projects.workspace_id = workspaces.id WHERE workspaces.organization_id = $1`)).
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

	t.Run("list workspaces with v_cluster_status", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		rows := sqlmock.NewRows([]string{"id", "name", "v_cluster_status"}).
			AddRow(uuid.New().String(), "workspace-1", "RUNNING").
			AddRow(uuid.New().String(), "workspace-2", "RUNNING").
			AddRow(uuid.New().String(), "workspace-3", "ERROR")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, v_cluster_status FROM "workspaces" WHERE organization_id = $1`)).
			WithArgs(orgID).
			WillReturnRows(rows)

		workspaces, err := repo.ListWorkspaces(ctx, orgID)
		assert.NoError(t, err)
		assert.Len(t, workspaces, 3)
		assert.Equal(t, "workspace-1", workspaces[0].Name)

		assert.Equal(t, "active", workspaces[0].Status)
		assert.Equal(t, "error", workspaces[2].Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count workspaces with v_cluster_status", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()

		totalRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "workspaces" WHERE organization_id = $1`)).
			WithArgs(orgID).
			WillReturnRows(totalRows)

		activeRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "workspaces" WHERE organization_id = $1 AND v_cluster_status = $2`)).
			WithArgs(orgID, "RUNNING").
			WillReturnRows(activeRows)

		total, active, err := repo.GetWorkspaceCount(ctx, orgID)
		assert.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Equal(t, 3, active)
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
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "organizations"`)).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organization_users" WHERE organization_id = $1 AND user_id = $2 ORDER BY "organization_users"."organization_id" LIMIT $3`)).
			WithArgs(orgID, userID, 1).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invitations" WHERE id = $1 ORDER BY "invitations"."id" LIMIT $2`)).
			WithArgs(invitationID, 1).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invitations" WHERE organization_id = $1 AND status = $2 ORDER BY created_at DESC`)).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invitations" WHERE organization_id = $1 ORDER BY created_at DESC`)).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`)).
			WithArgs(userID, 1).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
			WithArgs(email, 1).
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
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "email", "display_name", "provider", "external_id",
			"created_at", "updated_at", "last_login_at",
		}).
			AddRow(userID1, "user1@example.com", "User 1", "google", "google-1", now, now, nil).
			AddRow(userID2, "user2@example.com", "User 2", "github", "github-2", now, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id IN ($1,$2)`)).
			WithArgs(userID1, userID2).
			WillReturnRows(rows)

		users, err := repo.GetUsersByIDs(ctx, []string{userID1, userID2})
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		userID := uuid.New().String()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`)).
			WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUser(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user by email not found returns nil", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		email := "nonexistent@example.com"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
			WithArgs(email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUserByEmail(ctx, email)
		assert.NoError(t, err) // Returns nil, no error for GetUserByEmail
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
			"organization_id", "user_id", "email", "role",
			"invited_by", "invited_at", "joined_at", "status",
		}).AddRow(
			orgID, userID, "member@example.com", "admin",
			uuid.New().String(), now, now, "active",
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organization_users" WHERE organization_id = $1 AND user_id = $2 ORDER BY "organization_users"."organization_id" LIMIT $3`)).
			WithArgs(orgID, userID, 1).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invitations" WHERE id = $1 ORDER BY "invitations"."id" LIMIT $2`)).
			WithArgs(invitationID, 1).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invitations" WHERE token = $1 ORDER BY "invitations"."id" LIMIT $2`)).
			WithArgs(token, 1).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invitations" WHERE token = $1 ORDER BY "invitations"."id" LIMIT $2`)).
			WithArgs(token, 1).
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

		org := &domain.Organization{
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

		org := &domain.Organization{
			ID:               uuid.New().String(),
			Name:             "test-org",
			SubscriptionInfo: &domain.SubscriptionInfo{PlanID: "premium"}, // This causes issues
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

// Test UpdateOrganization with RowsAffected check
// Test UpdateOrganization with RowsAffected check
// Note: These tests use SQLite which has limited RETURNING clause support compared to PostgreSQL.
// In production PostgreSQL, clause.Returning{} will populate the original struct with all updated fields.
func TestPostgresRepository_UpdateOrganization_RowsAffected(t *testing.T) {
	ctx := context.Background()

	t.Run("returns ErrOrganizationNotFound when organization does not exist", func(t *testing.T) {
		// Use in-memory SQLite for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Auto-migrate the schema
		err = db.AutoMigrate(&dbOrganization{})
		require.NoError(t, err)

		repo := NewPostgresRepository(db)

		org := &domain.Organization{
			ID:          "non-existent-org",
			DisplayName: "Updated Name",
			UpdatedAt:   time.Now(),
		}

		err = repo.UpdateOrganization(ctx, org)

		// This should return ErrOrganizationNotFound
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOrganizationNotFound)
	})

	t.Run("successfully updates when organization exists", func(t *testing.T) {
		// Use in-memory SQLite for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Auto-migrate the schema
		err = db.AutoMigrate(&dbOrganization{})
		require.NoError(t, err)

		repo := NewPostgresRepository(db)

		// First create an organization
		existingOrg := &domain.Organization{
			ID:          "existing-org",
			Name:        "Test Org",
			DisplayName: "Original Name",
			Status:      "active",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err = repo.CreateOrganization(ctx, existingOrg)
		require.NoError(t, err)

		// Now update it
		updateOrg := &domain.Organization{
			ID:          "existing-org",
			DisplayName: "Updated Name",
			UpdatedAt:   time.Now(),
		}

		err = repo.UpdateOrganization(ctx, updateOrg)

		assert.NoError(t, err)

		// Note: In SQLite, RETURNING clause support is limited, so we may not get all original fields
		// In production PostgreSQL, updatedOrg would contain all fields including Name and Status
		// For now, we verify the update worked by checking a separate query
		retrieved, err := repo.GetOrganization(ctx, "existing-org")
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.DisplayName)
		assert.Equal(t, "Updated Name", retrieved.DisplayName)
	})
}

// Test ListMembers and ListOrganizations with pagination
func TestPostgresRepository_PaginationAndFiltering(t *testing.T) {
	ctx := context.Background()

	t.Run("list members with pagination", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		filter := domain.MemberFilter{
			OrganizationID: orgID,
			Page:           2,
			PageSize:       5,
		}

		// Count query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organization_users" WHERE organization_id = $1`)).
			WithArgs(orgID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))

		now := time.Now()
		dataRows := sqlmock.NewRows([]string{
			"organization_id", "user_id", "email", "role",
			"invited_by", "invited_at", "joined_at", "status",
		})

		// Add 5 rows for page 2
		for i := 0; i < 5; i++ {
			dataRows.AddRow(
				orgID, uuid.New().String(),
				fmt.Sprintf("user%d@example.com", i+6), "member", "", now, now, "active",
			)
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organization_users" WHERE organization_id = $1 ORDER BY joined_at ASC LIMIT $2 OFFSET $3`)).
			WithArgs(orgID, 5, 5).
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
		filter := domain.MemberFilter{
			OrganizationID: orgID,
			Search:         "john",
			Page:           1,
			PageSize:       10,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organization_users" JOIN users ON organization_users.user_id = users.id WHERE organization_id = $1 AND (users.email ILIKE $2 OR users.display_name ILIKE $3)`)).
			WithArgs(orgID, "%john%", "%john%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Data query with search
		now := time.Now()
		dataRows := sqlmock.NewRows([]string{
			"organization_id", "user_id", "email", "role",
			"invited_by", "invited_at", "joined_at", "status",
		}).
			AddRow(orgID, uuid.New().String(), "john.doe@example.com", "member", "", now, now, "active").
			AddRow(orgID, uuid.New().String(), "johnny@example.com", "admin", "", now, now, "active")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "organization_users"."organization_id","organization_users"."user_id","organization_users"."email","organization_users"."role","organization_users"."invited_by","organization_users"."invited_at","organization_users"."joined_at","organization_users"."status" FROM "organization_users" JOIN users ON organization_users.user_id = users.id WHERE organization_id = $1 AND (users.email ILIKE $2 OR users.display_name ILIKE $3) ORDER BY joined_at ASC LIMIT $4`)).
			WithArgs(orgID, "%john%", "%john%", 10).
			WillReturnRows(dataRows)

		members, total, err := repo.ListMembers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, members, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create and get organization", func(t *testing.T) {
		ctx := context.Background()
		gormDB, mock := setupTestDB(t)
		repo := NewPostgresRepository(gormDB)

		orgID := uuid.New().String()
		now := time.Now().UTC()
		deletedAt := now.Add(24 * time.Hour)
		org := &domain.Organization{
			ID:          orgID,
			Name:        "testorg",
			DisplayName: "Test Org",
			Description: "desc",
			Website:     "https://example.com",
			Email:       "test@example.com",
			Status:      "active",
			OwnerID:     uuid.New().String(),
			CreatedAt:   now,
			UpdatedAt:   now,
			DeletedAt:   &deletedAt,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO \"organizations\" (\"id\",\"name\",\"display_name\",\"description\",\"website\",\"email\",\"status\",\"owner_id\",\"stripe_customer_id\",\"stripe_subscription_id\",\"created_at\",\"updated_at\",\"deleted_at\") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)")).
			WithArgs(
				org.ID, org.Name, org.DisplayName, org.Description, org.Website, org.Email, org.Status, org.OwnerID, nil, nil, org.CreatedAt, org.UpdatedAt, org.DeletedAt,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateOrganization(ctx, org)
		assert.NoError(t, err)

		// GetOrganization
		rows := sqlmock.NewRows([]string{
			"id", "name", "display_name", "description", "website", "email", "status", "owner_id", "stripe_customer_id", "stripe_subscription_id", "created_at", "updated_at", "deleted_at",
		}).AddRow(
			org.ID, org.Name, org.DisplayName, org.Description, org.Website, org.Email, org.Status, org.OwnerID, nil, nil, org.CreatedAt, org.UpdatedAt, org.DeletedAt,
		)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"organizations\" WHERE id = $1 ORDER BY \"organizations\".\"id\" LIMIT $2")).
			WithArgs(orgID, 1).
			WillReturnRows(rows)

		got, err := repo.GetOrganization(ctx, orgID)
		assert.NoError(t, err)
		assert.Equal(t, org.ID, got.ID)
		assert.Equal(t, org.Name, got.Name)
		assert.Equal(t, org.DisplayName, got.DisplayName)
		assert.Equal(t, org.Description, got.Description)
		assert.Equal(t, org.Website, got.Website)
		assert.Equal(t, org.Email, got.Email)
		assert.Equal(t, org.Status, got.Status)
		assert.Equal(t, org.OwnerID, got.OwnerID)
		assert.WithinDuration(t, org.CreatedAt, got.CreatedAt, time.Second)
		assert.WithinDuration(t, org.UpdatedAt, got.UpdatedAt, time.Second)
		assert.WithinDuration(t, *org.DeletedAt, *got.DeletedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
