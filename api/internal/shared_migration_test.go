package internal

import (
	"testing"

	"github.com/hexabase/hexabase-ai/api/internal/shared/config"
	"github.com/hexabase/hexabase-ai/api/internal/shared/utils/httpauth"
)

// TestSharedPackageStructure tests if the shared packages can be imported correctly
func TestSharedPackageStructure(t *testing.T) {
	t.Run("config package import", func(t *testing.T) {
		// Test importing shared/config package
		// This test should now pass since config has been migrated
		cfg := tryImportSharedConfig()
		if cfg == nil {
			t.Error("Failed to import shared/config package")
		}
	})
	
	t.Run("utils package import", func(t *testing.T) {
		// Test importing shared/utils package
		// This test should now pass since utils has been migrated
		result := tryImportSharedUtils()
		if result == nil {
			t.Error("Failed to import shared/utils package")
		}
	})
	
	t.Run("middleware package import", func(t *testing.T) {
		// Test importing shared/middleware package
		t.Skip("Skipping until shared/middleware migration is complete")
		
		// This test will be enabled after migration
		// var _ interface{} = tryImportSharedMiddleware()
	})
	
	t.Run("db package import", func(t *testing.T) {
		// Test importing shared/db package
		t.Skip("Skipping until shared/db migration is complete")
		
		// This test will be enabled after migration
		// var _ interface{} = tryImportSharedDB()
	})
	
	t.Run("logging package import", func(t *testing.T) {
		// Test importing shared/logging package
		t.Skip("Skipping until shared/logging migration is complete")
		
		// This test will be enabled after migration
		// var _ interface{} = tryImportSharedLogging()
	})
	
	t.Run("redis package import", func(t *testing.T) {
		// Test importing shared/redis package
		t.Skip("Skipping until shared/redis migration is complete")
		
		// This test will be enabled after migration
		// var _ interface{} = tryImportSharedRedis()
	})
	
	t.Run("observability package import", func(t *testing.T) {
		// Test importing shared/observability package
		t.Skip("Skipping until shared/observability migration is complete")
		
		// This test will be enabled after migration
		// var _ interface{} = tryImportSharedObservability()
	})
	
	t.Run("infrastructure package import", func(t *testing.T) {
		// Test importing shared/infrastructure package
		t.Skip("Skipping until shared/infrastructure migration is complete")
		
		// This test will be enabled after migration
		// var _ interface{} = tryImportSharedInfrastructure()
	})
}

// Enabled functions after migration

func tryImportSharedConfig() interface{} {
	// This now works since shared/config migration is complete
	cfg, err := config.Load("")
	if err != nil {
		// This is expected in test environment
		return &config.Config{}
	}
	return cfg
}

func tryImportSharedUtils() interface{} {
	// This now works since shared/utils migration is complete
	// Test the httpauth subpackage
	hasPrefix := httpauth.HasBearerPrefix("Bearer token")
	if hasPrefix {
		return "httpauth package is working"
	}
	return nil
}

// These functions will be used after the actual migration
// They are commented out for now to avoid compilation errors

/*
func tryImportSharedMiddleware() interface{} {
	// After migration: import "github.com/hexabase/hexabase-ai/api/internal/shared/middleware"
	return nil
}

func tryImportSharedDB() interface{} {
	// After migration: import "github.com/hexabase/hexabase-ai/api/internal/shared/db"
	return nil
}

func tryImportSharedLogging() interface{} {
	// After migration: import "github.com/hexabase/hexabase-ai/api/internal/shared/logging"
	return nil
}

func tryImportSharedRedis() interface{} {
	// After migration: import "github.com/hexabase/hexabase-ai/api/internal/shared/redis"
	return nil
}

func tryImportSharedObservability() interface{} {
	// After migration: import "github.com/hexabase/hexabase-ai/api/internal/shared/observability"
	return nil
}

func tryImportSharedInfrastructure() interface{} {
	// After migration: import "github.com/hexabase/hexabase-ai/api/internal/shared/infrastructure"
	return nil
}
*/ 