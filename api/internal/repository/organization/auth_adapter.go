package organization

import (
	"context"
	"fmt"

	authDomain "github.com/hexabase/hexabase-kaas/api/internal/domain/auth"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/organization"
)

// AuthRepositoryAdapter adapts auth.Repository to organization.AuthRepository
type AuthRepositoryAdapter struct {
	authRepo authDomain.Repository
}

// NewAuthRepositoryAdapter creates a new auth repository adapter for organization domain
func NewAuthRepositoryAdapter(authRepo authDomain.Repository) organization.AuthRepository {
	return &AuthRepositoryAdapter{authRepo: authRepo}
}

// GetUser adapts the auth.User to organization.User
func (a *AuthRepositoryAdapter) GetUser(ctx context.Context, userID string) (*organization.User, error) {
	authUser, err := a.authRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert auth.User to organization.User
	return &organization.User{
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

// GetUserByEmail adapts the auth.User to organization.User by email
func (a *AuthRepositoryAdapter) GetUserByEmail(ctx context.Context, email string) (*organization.User, error) {
	authUser, err := a.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Convert auth.User to organization.User
	return &organization.User{
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