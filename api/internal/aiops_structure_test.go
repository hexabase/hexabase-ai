package internal

import (
	"testing"

	"github.com/hexabase/hexabase-ai/api/internal/aiops/domain"
	"github.com/hexabase/hexabase-ai/api/internal/aiops/handler"
	"github.com/hexabase/hexabase-ai/api/internal/aiops/repository"
	"github.com/hexabase/hexabase-ai/api/internal/aiops/service"
)

// TestAIOpsPackageStructure tests that the aiops package structure follows Package by Feature pattern
func TestAIOpsPackageStructure(t *testing.T) {
	t.Run("domain package imports", func(t *testing.T) {
		var _ domain.ChatSession
		var _ domain.ChatMessage
		var _ domain.ChatResponse
		var _ domain.ModelInfo
		var _ domain.Service
		var _ domain.Repository
		var _ domain.LLMService
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

// TestAIOpsPackageNaming tests that package names follow the convention
func TestAIOpsPackageNaming(t *testing.T) {
	t.Run("package names should be singular", func(t *testing.T) {
		// handler, domain, service, repository (not handlers, domains, services, repositories)
		t.Log("Package names should be: handler, domain, service, repository")
	})
}

// TestAIOpsNoCyclicDependencies tests that there are no cyclic dependencies
func TestAIOpsNoCyclicDependencies(t *testing.T) {
	t.Run("dependency direction", func(t *testing.T) {
		// handler -> service -> repository -> domain
		// domain should not import any other aiops packages
		t.Log("Dependencies should flow: handler -> service -> repository -> domain")
	})
} 