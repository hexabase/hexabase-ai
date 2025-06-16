package auth

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/auth"
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

func TestPostgresRepository_CreateUser(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("successful user creation", func(t *testing.T) {
		user := &auth.User{
			ID:          uuid.New().String(),
			Email:       "test@example.com",
			DisplayName: "Test User",
			Provider:    "google",
			ExternalID:  "google-sub-123",
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "users"`).
			WithArgs(
				user.ID,
				user.ExternalID,
				user.Provider,
				user.Email,
				user.DisplayName,
				user.AvatarURL,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				sqlmock.AnyArg(), // last_login_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateUser(ctx, user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate user error", func(t *testing.T) {
		user := &auth.User{
			ID:          uuid.New().String(),
			Email:       "duplicate@example.com",
			DisplayName: "Duplicate User",
			Provider:    "google",
			ExternalID:  "google-sub-456",
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "users"`).
			WithArgs(
				user.ID,
				user.ExternalID,
				user.Provider,
				user.Email,
				user.DisplayName,
				user.AvatarURL,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(gorm.ErrDuplicatedKey)
		mock.ExpectRollback()

		err := repo.CreateUser(ctx, user)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetUser(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("user found", func(t *testing.T) {
		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "external_id", "provider", "email", "display_name", "avatar_url", 
			"created_at", "updated_at", "last_login_at",
		}).AddRow(
			userID, "google-sub-123", "google", "user@example.com", "Test User", "",
			now, now, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 ORDER BY "users"\."id" LIMIT \$2`).
			WithArgs(userID, 1).
			WillReturnRows(rows)

		user, err := repo.GetUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		userID := uuid.New().String()

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 ORDER BY "users"\."id" LIMIT \$2`).
			WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUser(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("user found", func(t *testing.T) {
		email := "user@example.com"
		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "external_id", "provider", "email", "display_name", "avatar_url",
			"created_at", "updated_at", "last_login_at",
		}).AddRow(
			userID, "google-sub-123", "google", email, "Test User", "",
			now, now, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 ORDER BY "users"\."id" LIMIT \$2`).
			WithArgs(email, 1).
			WillReturnRows(rows)

		user, err := repo.GetUserByEmail(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, userID, user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		email := "notfound@example.com"

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 ORDER BY "users"\."id" LIMIT \$2`).
			WithArgs(email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUserByEmail(ctx, email)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetUserByExternalID(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("user found by external ID", func(t *testing.T) {
		externalID := "google-sub-789"
		provider := "google"
		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "external_id", "provider", "email", "display_name", "avatar_url",
			"created_at", "updated_at", "last_login_at",
		}).AddRow(
			userID, externalID, provider, "user@example.com", "Test User", "",
			now, now, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE external_id = \$1 AND provider = \$2 ORDER BY "users"\."id" LIMIT \$3`).
			WithArgs(externalID, provider, 1).
			WillReturnRows(rows)

		user, err := repo.GetUserByExternalID(ctx, externalID, provider)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, externalID, user.ExternalID)
		assert.Equal(t, provider, user.Provider)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_UpdateUser(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("successful user update", func(t *testing.T) {
		user := &auth.User{
			ID:          uuid.New().String(),
			Email:       "updated@example.com",
			DisplayName: "Updated Name",
			Provider:    "google",
			ExternalID:  "google-sub-999",
		}

		mock.ExpectBegin()
		// GORM Save will update all fields
		mock.ExpectExec(`UPDATE "users" SET`).
			WithArgs(
				user.ExternalID,
				user.Provider,
				user.Email,
				user.DisplayName,
				user.AvatarURL,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				sqlmock.AnyArg(), // last_login_at
				user.ID,          // WHERE id = ?
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateUser(ctx, user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_UpdateLastLogin(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("update_last_login_time", func(t *testing.T) {
		userID := uuid.New().String()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "users" SET "last_login_at"=\$1,"updated_at"=\$2 WHERE id = \$3`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateLastLogin(ctx, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_CreateSession(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("create new session", func(t *testing.T) {
		session := &auth.Session{
			ID:           uuid.New().String(),
			UserID:       uuid.New().String(),
			RefreshToken: "refresh-token-123",
			DeviceID:     "device-123",
			IPAddress:    "192.168.1.100",
			UserAgent:    "Mozilla/5.0...",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "sessions"`).
			WithArgs(
				session.ID,
				session.UserID,
				session.RefreshToken,
				session.DeviceID,
				session.IPAddress,
				session.UserAgent,
				sqlmock.AnyArg(), // expires_at
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // last_used_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateSession(ctx, session)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetSession(t *testing.T) {
	ctx := context.Background()
	gormDB, mock := setupTestDB(t)
	repo := NewPostgresRepository(gormDB)

	t.Run("get session by ID", func(t *testing.T) {
		sessionID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "user_id", "refresh_token", "device_id", "ip_address",
			"user_agent", "expires_at", "created_at", "last_used_at",
		}).AddRow(
			sessionID, "user-123", "refresh-token-123", "device-123", "192.168.1.100",
			"Mozilla/5.0...", now.Add(24*time.Hour), now, now,
		)

		mock.ExpectQuery(`SELECT \* FROM "sessions" WHERE id = \$1 ORDER BY "sessions"\."id" LIMIT \$2`).
			WithArgs(sessionID, 1).
			WillReturnRows(rows)

		session, err := repo.GetSession(ctx, sessionID)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, sessionID, session.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}