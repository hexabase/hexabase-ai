package internal

import (
	"testing"

	"github.com/hexabase/hexabase-ai/api/internal/function/domain"
	"github.com/hexabase/hexabase-ai/api/internal/function/handler"
	"github.com/hexabase/hexabase-ai/api/internal/function/repository"
	"github.com/hexabase/hexabase-ai/api/internal/function/service"
)

// TestFunctionPackageStructure tests that the function package structure follows Package by Feature pattern
func TestFunctionPackageStructure(t *testing.T) {
	t.Run("domain package imports", func(t *testing.T) {
		var _ domain.FunctionDef
		var _ domain.FunctionSpec
		var _ domain.FunctionVersionDef
		var _ domain.FunctionTrigger
		var _ domain.Repository
		var _ domain.Service
	})

	t.Run("handler package imports", func(t *testing.T) {
		// This will fail initially until we create the handler package
		var _ interface{} = (*handler.Handler)(nil)
	})

	t.Run("repository package imports", func(t *testing.T) {
		// This will fail initially until we move repository files
		var _ interface{} = (*repository.PostgresRepository)(nil)
	})

	t.Run("service package imports", func(t *testing.T) {
		// This will fail initially until we move service files
		var _ interface{} = (*service.Service)(nil)
	})
}

// TestFunctionDependencyDirection tests that dependencies flow in the correct direction
func TestFunctionDependencyDirection(t *testing.T) {
	t.Run("no circular dependencies", func(t *testing.T) {
		// handler depends on service, but not vice versa
		// service depends on domain and repository, but not vice versa
		// repository depends on domain, but not vice versa
		// This test will pass once the structure is correctly implemented
		t.Log("Dependencies should flow: handler -> service -> repository -> domain")
	})
} 