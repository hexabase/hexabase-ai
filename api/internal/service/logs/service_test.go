package logs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/logs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) QueryLogs(ctx context.Context, query logs.LogQuery) ([]*logs.LogEntry, error) {
	args := m.Called(ctx, query)
	if args.Get(0) != nil {
		return args.Get(0).([]*logs.LogEntry), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestService_QueryLogs(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo)

	t.Run("successful query with defaults", func(t *testing.T) {
		query := logs.LogQuery{
			WorkspaceID: "ws-123",
		}

		expectedLogs := []*logs.LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "Application started",
				Source:    "api-server",
			},
			{
				Timestamp: time.Now().Add(-5 * time.Minute),
				Level:     "warning",
				Message:   "High memory usage detected",
				Source:    "monitor",
			},
		}

		// Service will add defaults: limit=100, time range = last hour
		mockRepo.On("QueryLogs", ctx, mock.MatchedBy(func(q logs.LogQuery) bool {
			return q.WorkspaceID == "ws-123" && 
				   q.Limit == 100 && 
				   !q.StartTime.IsZero() &&
				   !q.EndTime.IsZero()
		})).Return(expectedLogs, nil)

		results, err := svc.QueryLogs(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "info", results[0].Level)
		assert.Equal(t, "warning", results[1].Level)

		mockRepo.AssertExpectations(t)
	})

	t.Run("query with custom parameters", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(-24 * time.Hour)
		
		query := logs.LogQuery{
			WorkspaceID: "ws-456",
			SearchTerm:  "error",
			Level:       "error",
			StartTime:   startTime,
			EndTime:     now,
			Limit:       50,
		}

		expectedLogs := []*logs.LogEntry{
			{
				Timestamp: now.Add(-10 * time.Minute),
				Level:     "error",
				Message:   "Database connection error",
				TraceID:   "trace-123",
			},
		}

		mockRepo.On("QueryLogs", ctx, query).Return(expectedLogs, nil)

		results, err := svc.QueryLogs(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "error", results[0].Level)
		assert.Contains(t, results[0].Message, "error")

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty workspace ID", func(t *testing.T) {
		query := logs.LogQuery{
			// Missing WorkspaceID
			SearchTerm: "test",
		}

		results, err := svc.QueryLogs(ctx, query)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "workspace ID is required")

		// Repository should not be called
		mockRepo.AssertNotCalled(t, "QueryLogs")
	})

	t.Run("invalid time range", func(t *testing.T) {
		now := time.Now()
		
		query := logs.LogQuery{
			WorkspaceID: "ws-789",
			StartTime:   now,
			EndTime:     now.Add(-1 * time.Hour), // End before start
		}

		results, err := svc.QueryLogs(ctx, query)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "invalid time range")

		mockRepo.AssertNotCalled(t, "QueryLogs")
	})

	t.Run("repository error", func(t *testing.T) {
		query := logs.LogQuery{
			WorkspaceID: "ws-error",
		}

		mockRepo.On("QueryLogs", ctx, mock.AnythingOfType("logs.LogQuery")).
			Return(nil, errors.New("database connection failed"))

		results, err := svc.QueryLogs(ctx, query)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "database connection failed")

		mockRepo.AssertExpectations(t)
	})

	t.Run("no results found", func(t *testing.T) {
		query := logs.LogQuery{
			WorkspaceID: "ws-empty",
			SearchTerm:  "nonexistent",
		}

		mockRepo.On("QueryLogs", ctx, mock.AnythingOfType("logs.LogQuery")).
			Return([]*logs.LogEntry{}, nil)

		results, err := svc.QueryLogs(ctx, query)
		assert.NoError(t, err)
		assert.Empty(t, results)

		mockRepo.AssertExpectations(t)
	})

	t.Run("logs with metadata", func(t *testing.T) {
		query := logs.LogQuery{
			WorkspaceID: "ws-metadata",
		}

		expectedLogs := []*logs.LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "User login successful",
				UserID:    "user-123",
				TraceID:   "trace-456",
				Source:    "auth-service",
				Details: map[string]interface{}{
					"ip_address": "192.168.1.100",
					"user_agent": "Mozilla/5.0",
					"session_id": "session-789",
				},
			},
		}

		mockRepo.On("QueryLogs", ctx, mock.AnythingOfType("logs.LogQuery")).
			Return(expectedLogs, nil)

		results, err := svc.QueryLogs(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		
		log := results[0]
		assert.Equal(t, "user-123", log.UserID)
		assert.Equal(t, "trace-456", log.TraceID)
		assert.NotNil(t, log.Details)
		assert.Equal(t, "192.168.1.100", log.Details["ip_address"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("limit exceeds maximum", func(t *testing.T) {
		query := logs.LogQuery{
			WorkspaceID: "ws-limit",
			Limit:       10000, // Very high limit
		}

		// Service should cap the limit to a reasonable value (e.g., 1000)
		mockRepo.On("QueryLogs", ctx, mock.MatchedBy(func(q logs.LogQuery) bool {
			return q.WorkspaceID == "ws-limit" && q.Limit <= 1000
		})).Return([]*logs.LogEntry{}, nil)

		results, err := svc.QueryLogs(ctx, query)
		assert.NoError(t, err)
		assert.Empty(t, results)

		mockRepo.AssertExpectations(t)
	})
}