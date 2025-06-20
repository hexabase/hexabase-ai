package repository

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/aiops/domain"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	
	return db, mock
}

func TestPostgresRepository_SaveChatSession(t *testing.T) {
	t.Run("successful save new session", func(t *testing.T) {
		db, mock := setupTestDB(t)
		repo := NewPostgresRepository(db, slog.Default())
		ctx := context.Background()
		
		session := &domain.ChatSession{
			ID:          uuid.New().String(),
			WorkspaceID: "workspace-123",
			UserID:      "user-123",
			Title:       "Test Chat",
			Model:       "llama2:7b",
			Messages: []domain.ChatMessage{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		// Mock checking for existing record (GORM uses SELECT 1)
		mock.ExpectQuery(`SELECT 1 FROM "chat_sessions" WHERE id = \$1 LIMIT \$2`).
			WithArgs(session.ID, 1).
			WillReturnError(gorm.ErrRecordNotFound)
		
		// Mock the insert
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "chat_sessions"`).
			WithArgs(
				session.ID,
				session.WorkspaceID,
				session.UserID,
				session.Title,
				session.Model,
				sqlmock.AnyArg(), // messages JSON
				sqlmock.AnyArg(), // context
				sqlmock.AnyArg(), // metadata
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		
		// Act
		err := repo.SaveChatSession(ctx, session)
		
		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	
	t.Run("successful update existing session", func(t *testing.T) {
		db, mock := setupTestDB(t)
		repo := NewPostgresRepository(db, slog.Default())
		ctx := context.Background()
		
		session := &domain.ChatSession{
			ID:          uuid.New().String(),
			WorkspaceID: "workspace-123",
			UserID:      "user-123",
			Title:       "Updated Chat",
			Model:       "llama2:7b",
			Messages: []domain.ChatMessage{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
				{Role: "user", Content: "How are you?"},
				{Role: "assistant", Content: "I'm doing well, thanks!"},
			},
			UpdatedAt: time.Now(),
		}
		
		// Mock checking for existing record - return any value to indicate it exists
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT 1 FROM "chat_sessions" WHERE id = \$1 LIMIT \$2`).
			WithArgs(session.ID, 1).
			WillReturnRows(rows)
		
		// Mock the update
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "chat_sessions" SET`).
			WithArgs(
				sqlmock.AnyArg(), // title
				sqlmock.AnyArg(), // messages JSON
				sqlmock.AnyArg(), // context
				sqlmock.AnyArg(), // metadata
				sqlmock.AnyArg(), // updated_at
				session.ID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		
		// Act
		err := repo.SaveChatSession(ctx, session)
		
		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetChatSession(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		db, mock := setupTestDB(t)
		repo := NewPostgresRepository(db, slog.Default())
		ctx := context.Background()
		
		sessionID := uuid.New().String()
		expectedMessages := []byte(`[{"role":"user","content":"Hello"},{"role":"assistant","content":"Hi!"}]`)
		
		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "user_id", "title", "model",
			"messages", "context", "metadata", "created_at", "updated_at",
		}).AddRow(
			sessionID, "workspace-123", "user-123", "Test Chat", "llama2:7b",
			expectedMessages, pq.Array([]int64{}), []byte("{}"), time.Now(), time.Now(),
		)
		
		mock.ExpectQuery(`SELECT \* FROM "chat_sessions" WHERE id = \$1 ORDER BY "chat_sessions"\."id" LIMIT \$2`).
			WithArgs(sessionID, 1).
			WillReturnRows(rows)
		
		// Act
		session, err := repo.GetChatSession(ctx, sessionID)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, sessionID, session.ID)
		assert.Equal(t, "workspace-123", session.WorkspaceID)
		assert.Len(t, session.Messages, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	
	t.Run("session not found", func(t *testing.T) {
		db, mock := setupTestDB(t)
		repo := NewPostgresRepository(db, slog.Default())
		ctx := context.Background()
		
		sessionID := uuid.New().String()
		
		mock.ExpectQuery(`SELECT \* FROM "chat_sessions" WHERE id = \$1 ORDER BY "chat_sessions"\."id" LIMIT \$2`).
			WithArgs(sessionID, 1).
			WillReturnError(gorm.ErrRecordNotFound)
		
		// Act
		session, err := repo.GetChatSession(ctx, sessionID)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_ListChatSessions(t *testing.T) {
	t.Run("successful listing with pagination", func(t *testing.T) {
		db, mock := setupTestDB(t)
		repo := NewPostgresRepository(db, slog.Default())
		ctx := context.Background()
		
		workspaceID := "workspace-123"
		limit := 10
		offset := 0
		
		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "user_id", "title", "model",
			"messages", "context", "metadata", "created_at", "updated_at",
		}).
			AddRow(
				"session-1", workspaceID, "user-1", "Chat 1", "llama2:7b",
				[]byte("[]"), pq.Array([]int64{}), []byte("{}"), time.Now(), time.Now(),
			).
			AddRow(
				"session-2", workspaceID, "user-2", "Chat 2", "codellama:13b",
				[]byte("[]"), pq.Array([]int64{}), []byte("{}"), time.Now(), time.Now(),
			)
		
		mock.ExpectQuery(`SELECT \* FROM "chat_sessions" WHERE workspace_id = \$1 ORDER BY updated_at DESC LIMIT \$2`).
			WithArgs(workspaceID, limit).
			WillReturnRows(rows)
		
		// Act
		sessions, err := repo.ListChatSessions(ctx, workspaceID, limit, offset)
		
		// Assert
		assert.NoError(t, err)
		assert.Len(t, sessions, 2)
		assert.Equal(t, "session-1", sessions[0].ID)
		assert.Equal(t, "session-2", sessions[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_DeleteChatSession(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		db, mock := setupTestDB(t)
		repo := NewPostgresRepository(db, slog.Default())
		ctx := context.Background()
		
		sessionID := uuid.New().String()
		
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "chat_sessions" WHERE id = \$1`).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		
		// Act
		err := repo.DeleteChatSession(ctx, sessionID)
		
		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_TrackModelUsage(t *testing.T) {
	t.Run("successful usage tracking", func(t *testing.T) {
		db, mock := setupTestDB(t)
		repo := NewPostgresRepository(db, slog.Default())
		ctx := context.Background()
		
		usage := &domain.ModelUsage{
			ID:               uuid.New().String(),
			WorkspaceID:      "workspace-123",
			UserID:           "user-123",
			SessionID:        "session-456",
			ModelName:        "llama2:7b",
			PromptTokens:     50,
			CompletionTokens: 75,
			TotalTokens:      125,
			RequestDuration:  time.Duration(1500) * time.Millisecond,
			Timestamp:        time.Now(),
		}
		
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "model_usage"`).
			WithArgs(
				usage.ID,
				usage.WorkspaceID,
				usage.UserID,
				usage.SessionID,
				usage.ModelName,
				usage.PromptTokens,
				usage.CompletionTokens,
				usage.TotalTokens,
				usage.RequestDuration.Milliseconds(),
				sqlmock.AnyArg(), // timestamp
				sqlmock.AnyArg(), // metadata
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		
		// Act
		err := repo.TrackModelUsage(ctx, usage)
		
		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetModelUsageStats(t *testing.T) {
	t.Run("successful stats retrieval", func(t *testing.T) {
		db, mock := setupTestDB(t)
		repo := NewPostgresRepository(db, slog.Default())
		ctx := context.Background()
		
		workspaceID := "workspace-123"
		modelName := "llama2:7b"
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now()
		
		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "user_id", "session_id", "model_name",
			"prompt_tokens", "completion_tokens", "total_tokens", 
			"request_duration_ms", "timestamp", "metadata",
		}).
			AddRow(
				"usage-1", workspaceID, "user-1", "session-1", modelName,
				50, 75, 125, 1500, time.Now().Add(-1*time.Hour), []byte("{}"),
			).
			AddRow(
				"usage-2", workspaceID, "user-2", "session-2", modelName,
				100, 150, 250, 2000, time.Now(), []byte("{}"),
			)
		
		mock.ExpectQuery(`SELECT \* FROM "model_usage" WHERE workspace_id = \$1 AND model_name = \$2 AND \(timestamp BETWEEN \$3 AND \$4\) ORDER BY timestamp DESC`).
			WithArgs(workspaceID, modelName, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)
		
		// Act
		stats, err := repo.GetModelUsageStats(ctx, workspaceID, modelName, from, to)
		
		// Assert
		assert.NoError(t, err)
		assert.Len(t, stats, 2)
		assert.Equal(t, "usage-1", stats[0].ID)
		assert.Equal(t, 125, stats[0].TotalTokens)
		assert.Equal(t, "usage-2", stats[1].ID)
		assert.Equal(t, 250, stats[1].TotalTokens)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	
	t.Run("retrieve all models when model name is empty", func(t *testing.T) {
		db, mock := setupTestDB(t)
		repo := NewPostgresRepository(db, slog.Default())
		ctx := context.Background()
		
		workspaceID := "workspace-123"
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now()
		
		rows := sqlmock.NewRows([]string{
			"id", "workspace_id", "user_id", "session_id", "model_name",
			"prompt_tokens", "completion_tokens", "total_tokens",
			"request_duration_ms", "timestamp", "metadata",
		}).
			AddRow(
				"usage-1", workspaceID, "user-1", "session-1", "llama2:7b",
				50, 75, 125, 1500, time.Now().Add(-1*time.Hour), []byte("{}"),
			).
			AddRow(
				"usage-2", workspaceID, "user-2", "session-2", "codellama:13b",
				100, 150, 250, 2000, time.Now(), []byte("{}"),
			)
		
		mock.ExpectQuery(`SELECT \* FROM "model_usage" WHERE workspace_id = \$1 AND \(timestamp BETWEEN \$2 AND \$3\) ORDER BY timestamp DESC`).
			WithArgs(workspaceID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)
		
		// Act
		stats, err := repo.GetModelUsageStats(ctx, workspaceID, "", from, to)
		
		// Assert
		assert.NoError(t, err)
		assert.Len(t, stats, 2)
		assert.Equal(t, "llama2:7b", stats[0].ModelName)
		assert.Equal(t, "codellama:13b", stats[1].ModelName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}