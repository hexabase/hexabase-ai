package logs

import "context"

// Service defines the interface for log-related business logic.
type Service interface {
	// QueryLogs retrieves logs based on a set of query parameters.
	QueryLogs(ctx context.Context, query LogQuery) ([]LogEntry, error)
} 