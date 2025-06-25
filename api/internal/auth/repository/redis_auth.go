package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
	"github.com/hexabase/hexabase-ai/api/internal/shared/redis"
)

// redisAuthRepository handles all Redis cache operations for auth domain
type redisAuthRepository struct {
	client *redis.Client
}

// NewRedisAuthRepository creates a new Redis auth repository
func NewRedisAuthRepository(client *redis.Client) *redisAuthRepository {
	return &redisAuthRepository{
		client: client,
	}
}

// Auth state operations

func (r *redisAuthRepository) StoreAuthState(ctx context.Context, state *domain.AuthState) error {
	key := formatAuthStateKey(state.State)
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal auth state: %w", err)
	}

	ttl := time.Until(state.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("auth state already expired")
	}

	if err := r.client.SetWithTTL(ctx, key, string(data), ttl); err != nil {
		return fmt.Errorf("failed to store auth state in Redis: %w", err)
	}

	return nil
}

func (r *redisAuthRepository) GetAuthState(ctx context.Context, stateValue string) (*domain.AuthState, error) {
	key := formatAuthStateKey(stateValue)

	data, err := r.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("auth state not found: %w", err)
	}

	var state domain.AuthState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth state: %w", err)
	}

	return &state, nil
}

func (r *redisAuthRepository) DeleteAuthState(ctx context.Context, stateValue string) error {
	key := formatAuthStateKey(stateValue)

	if err := r.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete auth state from Redis: %w", err)
	}

	return nil
}

// Refresh token blacklist operations

func (r *redisAuthRepository) BlacklistRefreshToken(ctx context.Context, token string, expiresAt time.Time) error {
	key := formatRefreshTokenBlacklistKey(token)
	ttl := time.Until(expiresAt)

	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}

	if err := r.client.SetWithTTL(ctx, key, "blacklisted", ttl); err != nil {
		return fmt.Errorf("failed to blacklist refresh token in Redis: %w", err)
	}

	return nil
}

func (r *redisAuthRepository) IsRefreshTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := formatRefreshTokenBlacklistKey(token)

	exists, err := r.client.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check refresh token blacklist in Redis: %w", err)
	}

	return exists, nil
}

// Session blocklist operations

func (r *redisAuthRepository) BlockSession(ctx context.Context, sessionID string, expiresAt time.Time) error {
	key := formatSessionBlocklistKey(sessionID)
	ttl := time.Until(expiresAt)

	if ttl <= 0 {
		// Session already expired, no need to block
		return nil
	}

	if err := r.client.SetWithTTL(ctx, key, "blocked", ttl); err != nil {
		return fmt.Errorf("failed to block session in Redis: %w", err)
	}

	return nil
}

func (r *redisAuthRepository) IsSessionBlocked(ctx context.Context, sessionID string) (bool, error) {
	key := formatSessionBlocklistKey(sessionID)

	exists, err := r.client.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check session blocklist in Redis: %w", err)
	}

	return exists, nil
}

// Key formatting helpers

func formatAuthStateKey(state string) string {
	return "auth_state:" + state
}

func formatRefreshTokenBlacklistKey(token string) string {
	return "refresh_token_blacklist:" + token
}

func formatSessionBlocklistKey(sessionID string) string {
	return "session_blocklist:" + sessionID
}