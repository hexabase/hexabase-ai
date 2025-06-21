package repository

import (
	"context"
	"fmt"

	authDomain "github.com/hexabase/hexabase-ai/api/internal/auth/domain"
	"github.com/hexabase/hexabase-ai/api/internal/organization/domain"
)

// AuthRepositoryAdapter adapts auth.Repository to domain.AuthRepository
type AuthRepositoryAdapter struct {
	authRepo authDomain.Repository
}

// NewAuthRepositoryAdapter creates a new auth repository adapter for organization domain
func NewAuthRepositoryAdapter(authRepo authDomain.Repository) domain.AuthRepository {
	return &AuthRepositoryAdapter{authRepo: authRepo}
}

// GetUser adapts the auth.User to domain.User
func (a *AuthRepositoryAdapter) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	authUser, err := a.authRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert auth.User to domain.User
	return &domain.User{
		ID:          authUser.ID,
		Email:       authUser.Email,
		DisplayName: authUser.DisplayName,
		Provider:    authUser.Provider,
		ExternalID:  authUser.ExternalID,
		CreatedAt:   authUser.CreatedAt,
		UpdatedAt:   authUser.UpdatedAt,
		LastLoginAt: &authUser.LastLoginAt,
	}, nil
}

// GetUserByEmail adapts the auth.User to domain.User by email
func (a *AuthRepositoryAdapter) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	authUser, err := a.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Convert auth.User to domain.User
	return &domain.User{
		ID:          authUser.ID,
		Email:       authUser.Email,
		DisplayName: authUser.DisplayName,
		Provider:    authUser.Provider,
		ExternalID:  authUser.ExternalID,
		CreatedAt:   authUser.CreatedAt,
		UpdatedAt:   authUser.UpdatedAt,
		LastLoginAt: &authUser.LastLoginAt,
	}, nil
}

// GetUserOrganizations gets the organization IDs for a user
func (a *AuthRepositoryAdapter) GetUserOrganizations(ctx context.Context, userID string) ([]string, error) {
	orgIDs, err := a.authRepo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}
	return orgIDs, nil
}