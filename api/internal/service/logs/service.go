package logs

import (
	"context"
	"log/slog"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/logs"
)

type service struct {
	repo   logs.Repository
	logger *slog.Logger
}

func NewLogService(repo logs.Repository, logger *slog.Logger) logs.Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) QueryLogs(ctx context.Context, query logs.LogQuery) ([]logs.LogEntry, error) {
	log := s.logger.With("workspace_id", query.WorkspaceID, "search_term", query.SearchTerm)
	log.Info("executing log query")
	
	// Apply default values
	if query.Limit <= 0 || query.Limit > 1000 {
		query.Limit = 100 // Default limit
	}
	if query.EndTime.IsZero() {
		query.EndTime = time.Now()
	}
	if query.StartTime.IsZero() {
		query.StartTime = query.EndTime.Add(-1 * time.Hour) // Default to last hour
	}

	results, err := s.repo.QueryLogs(ctx, query)
	if err != nil {
		log.Error("log query failed in repository", "error", err)
		return nil, err
	}

	log.Info("log query successful", "results_count", len(results))
	return results, nil
} 