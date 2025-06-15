package db

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm/logger"
	"log/slog"
)

// customLogger wraps the default logger to filter out specific warnings
type customLogger struct {
	logger.Interface
	slogger *slog.Logger
}

// NewCustomLogger creates a custom logger that filters out column-related warnings
func NewCustomLogger(logLevel logger.LogLevel, slogger *slog.Logger) logger.Interface {
	return &customLogger{
		Interface: logger.Default.LogMode(logLevel),
		slogger:   slogger,
	}
}

// Error implements logger.Interface
func (l *customLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	// Filter out specific GORM errors that we expect in development
	if strings.Contains(msg, "failed to parse field") ||
		strings.Contains(msg, "column") && strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "unsupported data type") {
		// Log as debug instead of error
		l.slogger.Debug("GORM field mismatch (expected in dev)", "msg", msg)
		return
	}
	l.Interface.Error(ctx, msg, data...)
}

// Warn implements logger.Interface
func (l *customLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	// Filter out record not found warnings for dev organization
	if strings.Contains(msg, "record not found") && strings.Contains(msg, "dev-org-1") {
		// This is expected when first creating dev org
		l.slogger.Debug("Dev org not found (will be created)", "msg", msg)
		return
	}
	l.Interface.Warn(ctx, msg, data...)
}

// Info implements logger.Interface
func (l *customLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Interface.Info(ctx, msg, data...)
}

// Trace implements logger.Interface
func (l *customLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// Get SQL and rows affected
	sql, _ := fc()
	
	// Filter out queries that are expected to fail in development
	if err != nil && strings.Contains(sql, "INSERT INTO") {
		if strings.Contains(err.Error(), "column") && strings.Contains(err.Error(), "does not exist") {
			// Log as debug with shorter message
			l.slogger.Debug("Schema mismatch (non-critical)", 
				"table", extractTableName(sql),
				"error", err.Error(),
			)
			return
		}
	}
	
	l.Interface.Trace(ctx, begin, fc, err)
}

// extractTableName extracts table name from SQL
func extractTableName(sql string) string {
	parts := strings.Split(sql, " ")
	for i, part := range parts {
		if strings.ToUpper(part) == "INTO" && i+1 < len(parts) {
			return strings.Trim(parts[i+1], "\"")
		}
	}
	return "unknown"
}