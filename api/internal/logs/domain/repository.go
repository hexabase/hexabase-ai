package domain

import "context"

// Repository defines the interface for the log persistence layer.
type Repository interface {
	QueryLogs(ctx context.Context, query LogQuery) ([]LogEntry, error)
} 