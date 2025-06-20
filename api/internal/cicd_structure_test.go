package internal

import (
	"log/slog"
	"testing"

	cicdDomain "github.com/hexabase/hexabase-ai/api/internal/cicd/domain"
	cicdRepository "github.com/hexabase/hexabase-ai/api/internal/cicd/repository"
	cicdService "github.com/hexabase/hexabase-ai/api/internal/cicd/service"
)

// TestCICDPackageStructure テスト - cicd機能の新しいパッケージ構造で正しくimportできるかを確認
func TestCICDPackageStructure(t *testing.T) {
	t.Run("domain types should be accessible", func(t *testing.T) {
		// PipelineRun型の確認
		var pipelineRun cicdDomain.PipelineRun
		if pipelineRun.ID == "" {
			// 正常: PipelineRun型が適切にアクセス可能
		}
		
		// PipelineConfig型の確認  
		var config cicdDomain.PipelineConfig
		if config.Name == "" {
			// 正常: PipelineConfig型が適切にアクセス可能
		}
		
		// LogEntry型の確認
		var logEntry cicdDomain.LogEntry
		if logEntry.Message == "" {
			// 正常: LogEntry型が適切にアクセス可能
		}
		
		t.Log("✅ All domain types are accessible")
	})
	
	t.Run("repository interface should be accessible", func(t *testing.T) {
		// Repository interface型の確認
		var repo cicdDomain.Repository
		if repo == nil {
			// 正常: Repository interface型が適切にアクセス可能
		}
		
		t.Log("✅ Repository interface is accessible")
	})
	
	t.Run("service interface should be accessible", func(t *testing.T) {
		// Service interface型の確認
		var svc cicdDomain.Service
		if svc == nil {
			// 正常: Service interface型が適切にアクセス可能
		}
		
		t.Log("✅ Service interface is accessible")
	})
	
	t.Run("service implementation should be accessible", func(t *testing.T) {
		// Service実装の確認（実際のコンストラクタを呼び出してみる）
		var repo cicdDomain.Repository // nilでも型チェックは可能
		var factory cicdDomain.ProviderFactory
		var credManager cicdDomain.CredentialManager
		logger := slog.Default()
		svc := cicdService.NewService(repo, factory, credManager, logger)
		if svc == nil {
			t.Error("Service constructor should return valid service")
		}
		
		t.Log("✅ Service implementation is accessible")
	})
	
	t.Run("repository implementation should be accessible", func(t *testing.T) {
		// Repository実装の確認（実際のコンストラクタを呼び出してみる）
		repo := cicdRepository.NewPostgresRepository(nil)  // nilでも型チェックは可能
		if repo == nil {
			t.Error("Repository constructor should return valid repository")
		}
		
		t.Log("✅ Repository implementation is accessible")
	})
} 