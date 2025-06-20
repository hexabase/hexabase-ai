package internal

import (
	"log/slog"
	"testing"

	logsDomain "github.com/hexabase/hexabase-ai/api/internal/logs/domain"
	logsRepository "github.com/hexabase/hexabase-ai/api/internal/logs/repository"
	logsService "github.com/hexabase/hexabase-ai/api/internal/logs/service"
)

// TestLogsPackageStructure テスト - logs機能の新しいパッケージ構造で正しくimportできるかを確認
func TestLogsPackageStructure(t *testing.T) {
	t.Run("domain types should be accessible", func(t *testing.T) {
		// LogEntry型の確認
		var logEntry logsDomain.LogEntry
		if logEntry.Timestamp.IsZero() {
			// 正常: LogEntry型が適切にアクセス可能
		}
		
		// AuditLog型の確認  
		var auditLog logsDomain.AuditLog
		if auditLog.ID == "" {
			// 正常: AuditLog型が適切にアクセス可能
		}
		
		// LogQuery型の確認
		var query logsDomain.LogQuery
		if query.WorkspaceID == "" {
			// 正常: LogQuery型が適切にアクセス可能
		}
		
		t.Log("✅ All domain types are accessible")
	})
	
	t.Run("repository interface should be accessible", func(t *testing.T) {
		// Repository interface型の確認
		var repo logsDomain.Repository
		if repo == nil {
			// 正常: Repository interface型が適切にアクセス可能
		}
		
		t.Log("✅ Repository interface is accessible")
	})
	
	t.Run("service interface should be accessible", func(t *testing.T) {
		// Service interface型の確認
		var svc logsDomain.Service
		if svc == nil {
			// 正常: Service interface型が適切にアクセス可能
		}
		
		t.Log("✅ Service interface is accessible")
	})
	
	t.Run("service implementation should be accessible", func(t *testing.T) {
		// Service実装の確認（実際のコンストラクタを呼び出してみる）
		// これはmockリポジトリで構築できるはず
		var repo logsDomain.Repository // nilでも型チェックは可能
		logger := slog.Default()
		svc := logsService.NewLogService(repo, logger)
		if svc == nil {
			t.Error("Service constructor should return valid service")
		}
		
		t.Log("✅ Service implementation is accessible")
	})
	
	t.Run("repository implementation should be accessible", func(t *testing.T) {
		// Repository実装の確認（実際のコンストラクタを呼び出してみる）
		// ClickHouseConnとloggerを渡す必要がある
		logger := slog.Default()
		repo := logsRepository.NewClickHouseRepository(nil, logger)  // nilでも型チェックは可能
		if repo == nil {
			t.Error("Repository constructor should return valid repository")
		}
		
		t.Log("✅ Repository implementation is accessible")
	})
} 