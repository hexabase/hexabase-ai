package service_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
	"github.com/hexabase/hexabase-ai/api/internal/auth/repository"
	"github.com/hexabase/hexabase-ai/api/internal/auth/service"
	"github.com/hexabase/hexabase-ai/api/internal/shared/config"
	internalRedis "github.com/hexabase/hexabase-ai/api/internal/shared/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testUserID = "user-123"
)

// testContext returns a context for testing that's cancelled when the test ends
func testContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	return ctx
}

// mockSessionRepository is a mock for session repository
type mockSessionRepository struct {
	mock.Mock
}

func (m *mockSessionRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockSessionRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	user, ok := args.Get(0).(*domain.User)

	if !ok {
		return nil, args.Error(1)
	}

	return user, args.Error(1)
}

func (m *mockSessionRepository) GetUserByExternalID(
	ctx context.Context, externalID, provider string,
) (*domain.User, error) {
	args := m.Called(ctx, externalID, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	user, ok := args.Get(0).(*domain.User)

	if !ok {
		return nil, args.Error(1)
	}

	return user, args.Error(1)
}

func (m *mockSessionRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	user, ok := args.Get(0).(*domain.User)

	if !ok {
		return nil, args.Error(1)
	}

	return user, args.Error(1)
}

func (m *mockSessionRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockSessionRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockSessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	session, ok := args.Get(0).(*domain.Session)

	if !ok {
		return nil, args.Error(1)
	}

	return session, args.Error(1)
}

func (m *mockSessionRepository) GetSessionByRefreshTokenSelector(
	ctx context.Context, selector string,
) (*domain.Session, error) {
	args := m.Called(ctx, selector)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	session, ok := args.Get(0).(*domain.Session)

	if !ok {
		return nil, args.Error(1)
	}

	return session, args.Error(1)
}

func (m *mockSessionRepository) GetAllActiveSessions(ctx context.Context) ([]*domain.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	sessions, ok := args.Get(0).([]*domain.Session)

	if !ok {
		return nil, args.Error(1)
	}

	return sessions, args.Error(1)
}

func (m *mockSessionRepository) ListUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	sessions, ok := args.Get(0).([]*domain.Session)

	if !ok {
		return nil, args.Error(1)
	}

	return sessions, args.Error(1)
}

func (m *mockSessionRepository) UpdateSession(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockSessionRepository) DeleteUserSessions(ctx context.Context, userID string, exceptSessionID string) error {
	args := m.Called(ctx, userID, exceptSessionID)
	return args.Error(0)
}

func (m *mockSessionRepository) CleanupExpiredSessions(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *mockSessionRepository) StoreAuthState(ctx context.Context, state *domain.AuthState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *mockSessionRepository) GetAuthState(ctx context.Context, stateValue string) (*domain.AuthState, error) {
	args := m.Called(ctx, stateValue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	state, ok := args.Get(0).(*domain.AuthState)

	if !ok {
		return nil, args.Error(1)
	}

	return state, args.Error(1)
}

func (m *mockSessionRepository) DeleteAuthState(ctx context.Context, stateValue string) error {
	args := m.Called(ctx, stateValue)
	return args.Error(0)
}

func (m *mockSessionRepository) BlacklistRefreshToken(ctx context.Context, token string, expiresAt time.Time) error {
	args := m.Called(ctx, token, expiresAt)
	return args.Error(0)
}

func (m *mockSessionRepository) IsRefreshTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *mockSessionRepository) BlockSession(ctx context.Context, sessionID string, expiresAt time.Time) error {
	args := m.Called(ctx, sessionID, expiresAt)
	return args.Error(0)
}

func (m *mockSessionRepository) IsSessionBlocked(ctx context.Context, sessionID string) (bool, error) {
	args := m.Called(ctx, sessionID)
	return args.Bool(0), args.Error(1)
}

func (m *mockSessionRepository) CreateSecurityEvent(ctx context.Context, event *domain.SecurityEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockSessionRepository) ListSecurityEvents(
	ctx context.Context, filter domain.SecurityLogFilter,
) ([]*domain.SecurityEvent, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	events, ok := args.Get(0).([]*domain.SecurityEvent)

	if !ok {
		return nil, args.Error(1)
	}

	return events, args.Error(1)
}

func (m *mockSessionRepository) CleanupOldSecurityEvents(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *mockSessionRepository) GetUserOrganizations(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	strings, ok := args.Get(0).([]string)

	if !ok {
		return nil, args.Error(1)
	}

	return strings, args.Error(1)
}

func (m *mockSessionRepository) GetUserWorkspaceGroups(
	ctx context.Context, userID, workspaceID string,
) ([]string, error) {
	args := m.Called(ctx, userID, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	strings, ok := args.Get(0).([]string)

	if !ok {
		return nil, args.Error(1)
	}

	return strings, args.Error(1)
}

func (m *mockSessionRepository) HashToken(token string) (hashedToken string, salt string, err error) {
	args := m.Called(token)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockSessionRepository) VerifyToken(plainToken, hashedToken, salt string) bool {
	args := m.Called(plainToken, hashedToken, salt)
	return args.Bool(0)
}

// Test setup function
//
//nolint:ireturn // Test helper returning interface
func setupSessionManagerTest(t *testing.T) (
	*miniredis.Miniredis, domain.SessionLimiterRepository, *mockSessionRepository,
) {
	t.Helper()

	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create wrapped Redis client for internal usage
	logger := slog.Default()
	wrappedClient, err := internalRedis.NewClient(&config.RedisConfig{
		Host: mr.Host(),
		Port: mr.Port(),
		DB:   0,
	}, logger)
	require.NoError(t, err)

	// Create SessionLimiterRepository
	limiterRepo := repository.NewSessionLimiterRepository(wrappedClient)

	// Create mock repository
	mockRepo := new(mockSessionRepository)

	// Cleanup
	t.Cleanup(func() {
		_ = client.Close()
		_ = wrappedClient.Close()
		mr.Close()
	})

	return mr, limiterRepo, mockRepo
}

func TestSessionManager_CreateSession_UnderLimit(t *testing.T) {
	t.Parallel()
	// Test: When concurrent session count is less than 3, new session is created successfully

	ctx := testContext(t)
	userID := testUserID
	sessionID := "session-new"

	// Setup
	mr, limiterRepo, mockRepo := setupSessionManagerTest(t)

	// Create SessionManager
	sessionManager := service.NewSessionManager(mockRepo, limiterRepo)

	// Precondition: 2 sessions already exist
	_, err2 := mr.SAdd("user_sessions:"+testUserID, "session-1", "session-2")
	require.NoError(t, err2)

	// Execute
	err := sessionManager.CreateSession(ctx, userID, sessionID)

	// Verify
	require.NoError(t, err)

	// Verify that new session was added to Redis
	members, err := mr.SMembers("user_sessions:" + testUserID)
	require.NoError(t, err)
	assert.Len(t, members, 3)
	assert.Contains(t, members, sessionID)
}

func TestSessionManager_CreateSession_AtLimit_ReturnsError(t *testing.T) {
	t.Parallel()
	// Test: When concurrent session count is 3 or more, returns ErrTooManySessions

	ctx := testContext(t)
	userID := testUserID
	newSessionID := "session-new"

	// Setup
	mr, limiterRepo, mockRepo := setupSessionManagerTest(t)

	// Create SessionManager
	sessionManager := service.NewSessionManager(mockRepo, limiterRepo)

	// Precondition: 3 sessions already exist (at limit)
	_, err2 := mr.SAdd("user_sessions:"+testUserID, "session-1", "session-2", "session-3")
	require.NoError(t, err2)

	// Execute
	err := sessionManager.CreateSession(ctx, userID, newSessionID)

	// Verify
	require.Error(t, err)
	assert.Equal(t, domain.ErrTooManySessions, err)

	// Check Redis state - should remain unchanged
	members, err := mr.SMembers("user_sessions:" + testUserID)
	require.NoError(t, err)
	assert.Len(t, members, 3)                    // Still 3
	assert.NotContains(t, members, newSessionID) // New session NOT added

	// Verify no mock calls were made (no DB operations)
	mockRepo.AssertExpectations(t)
}

func TestSessionManager_DeleteSession(t *testing.T) {
	t.Parallel()
	// Test: Session removal works correctly

	ctx := testContext(t)
	userID := testUserID
	sessionID := "session-to-remove"

	// Setup
	mr, limiterRepo, mockRepo := setupSessionManagerTest(t)

	// Create SessionManager
	sessionManager := service.NewSessionManager(mockRepo, limiterRepo)

	// Precondition: Session exists
	_, err2 := mr.SAdd("user_sessions:"+testUserID, sessionID, "session-2")
	require.NoError(t, err2)

	// Execute
	err := sessionManager.DeleteSession(ctx, userID, sessionID)

	// Verify
	require.NoError(t, err)

	// Verify removal from Redis
	members, err := mr.SMembers("user_sessions:" + testUserID)
	require.NoError(t, err)
	assert.NotContains(t, members, sessionID)
}

func TestSessionManager_GetActiveSessionCount(t *testing.T) {
	t.Parallel()
	// Test: Getting active session count works correctly

	ctx := testContext(t)
	userID := testUserID

	// Setup
	mr, limiterRepo, mockRepo := setupSessionManagerTest(t)

	// Create SessionManager
	sessionManager := service.NewSessionManager(mockRepo, limiterRepo)

	// Precondition: 2 sessions exist
	_, err2 := mr.SAdd("user_sessions:"+testUserID, "session-1", "session-2")
	require.NoError(t, err2)

	// Execute
	count, err := sessionManager.GetActiveSessionCount(ctx, userID)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}
