package auth_test

import (
	"testing"

	// Test importing auth domain
	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"

	// Test importing auth service
	"github.com/hexabase/hexabase-ai/api/internal/auth/service"

	// Test importing auth repository
	"github.com/hexabase/hexabase-ai/api/internal/auth/repository"

	// Test importing auth handler
	"github.com/hexabase/hexabase-ai/api/internal/auth/handler"
)

// TestPackageStructure tests that all auth packages can be imported correctly
func TestPackageStructure(t *testing.T) {
	t.Run("domain package import", func(t *testing.T) {
		// Test that domain interfaces can be referenced
		var _ domain.Service
		var _ domain.Repository
	})
	
	t.Run("service package import", func(t *testing.T) {
		// Test that service can be referenced
		_ = service.NewService
	})
	
	t.Run("repository package import", func(t *testing.T) {
		// Test that repository can be referenced
		_ = repository.NewPostgresRepository
	})
	
	t.Run("handler package import", func(t *testing.T) {
		// Test that handler can be referenced
		_ = handler.NewHandler
	})
}

// TestCircularDependency tests that there are no circular dependencies
func TestCircularDependency(t *testing.T) {
	// This test passes if the imports above work without circular dependency errors
	t.Log("No circular dependencies detected in auth package structure")
} 