package repository

import (
	"context"
	"fmt"
	"time"

	authDomain "github.com/hexabase/hexabase-ai/api/internal/auth/domain"
	"github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
)

// AuthRepositoryAdapter adapts auth.Repository to workspace.AuthRepository
type AuthRepositoryAdapter struct {
	authRepo authDomain.Repository
}

// NewAuthRepositoryAdapter creates a new adapter
func NewAuthRepositoryAdapter(authRepo authDomain.Repository) domain.AuthRepository {
	return &AuthRepositoryAdapter{authRepo: authRepo}
}

// GetUser adapts the auth.User to workspace.User
func (a *AuthRepositoryAdapter) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	authUser, err := a.authRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert auth.User to workspace.User
	return &domain.User{
		ID:    authUser.ID,
		Email: authUser.Email,
		Name:  authUser.Email, // Use email as name for now
	}, nil
}

// GenerateWorkspaceToken generates a token for accessing a workspace
func (a *AuthRepositoryAdapter) GenerateWorkspaceToken(ctx context.Context, userID, workspaceID string) (string, error) {
	// This would typically generate a JWT token with workspace-specific claims
	// For now, return a placeholder token
	// In a real implementation, this would use the JWT service
	token := fmt.Sprintf("workspace-token-%s-%s-%d", userID, workspaceID, time.Now().Unix())
	return token, nil
}