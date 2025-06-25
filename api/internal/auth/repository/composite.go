package repository

import (
	"context"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
)

// CompositeRepository implements domain.Repository by delegating to specialized repositories
type CompositeRepository struct {
	dbRepo    *postgresRepository
	cacheRepo *redisAuthRepository
}

// NewCompositeRepository creates a new composite repository
func NewCompositeRepository(dbRepo *postgresRepository, cacheRepo *redisAuthRepository) domain.Repository {
	return &CompositeRepository{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
	}
}

// User operations (PostgreSQL)

func (r *CompositeRepository) CreateUser(ctx context.Context, user *domain.User) error {
	return r.dbRepo.CreateUser(ctx, user)
}

func (r *CompositeRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return r.dbRepo.GetUser(ctx, userID)
}

func (r *CompositeRepository) GetUserByExternalID(ctx context.Context, externalID, provider string) (*domain.User, error) {
	return r.dbRepo.GetUserByExternalID(ctx, externalID, provider)
}

func (r *CompositeRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.dbRepo.GetUserByEmail(ctx, email)
}

func (r *CompositeRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	return r.dbRepo.UpdateUser(ctx, user)
}

func (r *CompositeRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	return r.dbRepo.UpdateLastLogin(ctx, userID)
}

// Session operations (PostgreSQL)

func (r *CompositeRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	return r.dbRepo.CreateSession(ctx, session)
}

func (r *CompositeRepository) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	return r.dbRepo.GetSession(ctx, sessionID)
}

func (r *CompositeRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	return r.dbRepo.GetSessionByRefreshToken(ctx, refreshToken)
}

func (r *CompositeRepository) ListUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	return r.dbRepo.ListUserSessions(ctx, userID)
}

func (r *CompositeRepository) UpdateSession(ctx context.Context, session *domain.Session) error {
	return r.dbRepo.UpdateSession(ctx, session)
}

func (r *CompositeRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.dbRepo.DeleteSession(ctx, sessionID)
}

func (r *CompositeRepository) DeleteUserSessions(ctx context.Context, userID string, exceptSessionID string) error {
	return r.dbRepo.DeleteUserSessions(ctx, userID, exceptSessionID)
}

func (r *CompositeRepository) CleanupExpiredSessions(ctx context.Context, before time.Time) error {
	return r.dbRepo.CleanupExpiredSessions(ctx, before)
}

// Auth state operations (Redis)

func (r *CompositeRepository) StoreAuthState(ctx context.Context, state *domain.AuthState) error {
	return r.cacheRepo.StoreAuthState(ctx, state)
}

func (r *CompositeRepository) GetAuthState(ctx context.Context, stateValue string) (*domain.AuthState, error) {
	return r.cacheRepo.GetAuthState(ctx, stateValue)
}

func (r *CompositeRepository) DeleteAuthState(ctx context.Context, stateValue string) error {
	return r.cacheRepo.DeleteAuthState(ctx, stateValue)
}

// Refresh token operations (Redis)

func (r *CompositeRepository) BlacklistRefreshToken(ctx context.Context, token string, expiresAt time.Time) error {
	return r.cacheRepo.BlacklistRefreshToken(ctx, token, expiresAt)
}

func (r *CompositeRepository) IsRefreshTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	return r.cacheRepo.IsRefreshTokenBlacklisted(ctx, token)
}

// Session blocklist operations (Redis)

func (r *CompositeRepository) BlockSession(ctx context.Context, sessionID string, expiresAt time.Time) error {
	return r.cacheRepo.BlockSession(ctx, sessionID, expiresAt)
}

func (r *CompositeRepository) IsSessionBlocked(ctx context.Context, sessionID string) (bool, error) {
	return r.cacheRepo.IsSessionBlocked(ctx, sessionID)
}

// Security event operations (PostgreSQL)

func (r *CompositeRepository) CreateSecurityEvent(ctx context.Context, event *domain.SecurityEvent) error {
	return r.dbRepo.CreateSecurityEvent(ctx, event)
}

func (r *CompositeRepository) ListSecurityEvents(ctx context.Context, filter domain.SecurityLogFilter) ([]*domain.SecurityEvent, error) {
	return r.dbRepo.ListSecurityEvents(ctx, filter)
}

func (r *CompositeRepository) CleanupOldSecurityEvents(ctx context.Context, before time.Time) error {
	return r.dbRepo.CleanupOldSecurityEvents(ctx, before)
}

// User organization operations (PostgreSQL)

func (r *CompositeRepository) GetUserOrganizations(ctx context.Context, userID string) ([]string, error) {
	return r.dbRepo.GetUserOrganizations(ctx, userID)
}

func (r *CompositeRepository) GetUserWorkspaceGroups(ctx context.Context, userID, workspaceID string) ([]string, error) {
	return r.dbRepo.GetUserWorkspaceGroups(ctx, userID, workspaceID)
}