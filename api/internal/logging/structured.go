package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
)

// StructuredLogger wraps slog with ClickHouse integration and contextual information
type StructuredLogger struct {
	slogger   *slog.Logger
	clickHouse *ClickHouseLogger
	service   string
	component string
}

// ContextKey type for context keys
type ContextKey string

const (
	// Context keys for structured logging
	TraceIDKey     ContextKey = "trace_id"
	UserIDKey      ContextKey = "user_id"
	OrgIDKey       ContextKey = "org_id"
	WorkspaceIDKey ContextKey = "workspace_id"
	ComponentKey   ContextKey = "component"
	HTTPMethodKey  ContextKey = "http_method"
	HTTPPathKey    ContextKey = "http_path"
	HTTPStatusKey  ContextKey = "http_status"
	DurationKey    ContextKey = "duration_ms"
)

// LoggerConfig holds configuration for the structured logger
type LoggerConfig struct {
	Service         string
	Component       string
	Level           slog.Level
	AddSource       bool
	ClickHouseURL   string
	EnableClickHouse bool
}

// NewStructuredLogger creates a new structured logger with ClickHouse integration
func NewStructuredLogger(config *LoggerConfig, clickHouse *ClickHouseLogger) *StructuredLogger {
	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize attribute handling if needed
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   "timestamp",
					Value: a.Value,
				}
			}
			return a
		},
	}

	// Create a JSON handler for structured output
	handler := slog.NewJSONHandler(os.Stdout, opts)
	slogger := slog.New(handler)

	return &StructuredLogger{
		slogger:    slogger,
		clickHouse: clickHouse,
		service:    config.Service,
		component:  config.Component,
	}
}

// WithContext creates a new logger with context information
func (l *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	newLogger := &StructuredLogger{
		slogger:    l.slogger,
		clickHouse: l.clickHouse,
		service:    l.service,
		component:  getStringFromContext(ctx, ComponentKey, l.component),
	}
	return newLogger
}

// WithComponent creates a new logger with a specific component
func (l *StructuredLogger) WithComponent(component string) *StructuredLogger {
	return &StructuredLogger{
		slogger:    l.slogger,
		clickHouse: l.clickHouse,
		service:    l.service,
		component:  component,
	}
}

// Info logs an info message
func (l *StructuredLogger) Info(ctx context.Context, msg string, fields ...slog.Attr) {
	l.log(ctx, slog.LevelInfo, msg, fields...)
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(ctx context.Context, msg string, fields ...slog.Attr) {
	l.log(ctx, slog.LevelDebug, msg, fields...)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(ctx context.Context, msg string, fields ...slog.Attr) {
	l.log(ctx, slog.LevelWarn, msg, fields...)
}

// Error logs an error message
func (l *StructuredLogger) Error(ctx context.Context, msg string, err error, fields ...slog.Attr) {
	if err != nil {
		fields = append(fields, slog.String("error", err.Error()))
	}
	l.log(ctx, slog.LevelError, msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(ctx context.Context, msg string, err error, fields ...slog.Attr) {
	if err != nil {
		fields = append(fields, slog.String("error", err.Error()))
	}
	l.log(ctx, slog.Level(12), msg, fields...) // Custom level for fatal
	os.Exit(1)
}

// LogHTTP logs HTTP request/response information
func (l *StructuredLogger) LogHTTP(ctx context.Context, method, path string, status int, duration time.Duration, fields ...slog.Attr) {
	httpFields := []slog.Attr{
		slog.String("http_method", method),
		slog.String("http_path", path),
		slog.Int("http_status", status),
		slog.Int64("duration_ms", duration.Milliseconds()),
	}
	httpFields = append(httpFields, fields...)
	
	level := slog.LevelInfo
	if status >= 400 {
		level = slog.LevelWarn
	}
	if status >= 500 {
		level = slog.LevelError
	}

	l.log(ctx, level, fmt.Sprintf("%s %s", method, path), httpFields...)
}

// LogSecurity logs security-related events
func (l *StructuredLogger) LogSecurity(ctx context.Context, eventType string, severity string, message string, fields ...slog.Attr) {
	securityFields := []slog.Attr{
		slog.String("event_type", eventType),
		slog.String("severity", severity),
		slog.String("category", "security"),
	}
	securityFields = append(securityFields, fields...)

	level := slog.LevelInfo
	switch severity {
	case "low":
		level = slog.LevelInfo
	case "medium":
		level = slog.LevelWarn
	case "high", "critical":
		level = slog.LevelError
	}

	l.log(ctx, level, message, securityFields...)

	// Also log to ClickHouse security events table
	if l.clickHouse != nil {
		securityEntry := &SecurityLogEntry{
			Timestamp:   time.Now(),
			EventType:   eventType,
			Severity:    severity,
			UserID:      getStringFromContext(ctx, UserIDKey, ""),
			OrgID:       getStringFromContext(ctx, OrgIDKey, ""),
			WorkspaceID: getStringFromContext(ctx, WorkspaceIDKey, ""),
			Message:     message,
			Metadata:    convertAttrsToMap(fields),
		}

		// Calculate risk score based on severity
		switch severity {
		case "low":
			securityEntry.RiskScore = 1.0
		case "medium":
			securityEntry.RiskScore = 5.0
		case "high":
			securityEntry.RiskScore = 8.0
		case "critical":
			securityEntry.RiskScore = 10.0
		}

		if err := l.clickHouse.LogSecurity(ctx, securityEntry); err != nil {
			l.slogger.Error("Failed to log security event to ClickHouse", "error", err)
		}
	}
}

// log is the internal logging method
func (l *StructuredLogger) log(ctx context.Context, level slog.Level, msg string, fields ...slog.Attr) {
	// Add contextual information
	contextFields := l.extractContextFields(ctx)
	allFields := append(contextFields, fields...)

	// Add source information
	if pc, file, line, ok := runtime.Caller(2); ok {
		allFields = append(allFields,
			slog.String("source_file", file),
			slog.Int("source_line", line),
			slog.String("source_func", runtime.FuncForPC(pc).Name()),
		)
	}

	// Log to slog
	l.slogger.LogAttrs(ctx, level, msg, allFields...)

	// Log to ClickHouse if enabled
	if l.clickHouse != nil {
		entry := &LogEntry{
			Timestamp:   time.Now(),
			TraceID:     getStringFromContext(ctx, TraceIDKey, ""),
			Level:       level.String(),
			Service:     l.service,
			Component:   l.component,
			UserID:      getStringFromContext(ctx, UserIDKey, ""),
			OrgID:       getStringFromContext(ctx, OrgIDKey, ""),
			WorkspaceID: getStringFromContext(ctx, WorkspaceIDKey, ""),
			Message:     msg,
			Fields:      convertAttrsToMap(allFields),
			HTTPMethod:  getStringFromContext(ctx, HTTPMethodKey, ""),
			HTTPPath:    getStringFromContext(ctx, HTTPPathKey, ""),
			HTTPStatus:  getUint16FromContext(ctx, HTTPStatusKey, 0),
			DurationMS:  getUint32FromContext(ctx, DurationKey, 0),
		}

		// Add source information
		if _, file, line, ok := runtime.Caller(2); ok {
			entry.SourceFile = file
			entry.SourceLine = uint32(line)
		}

		// Log asynchronously to avoid blocking
		go func() {
			if err := l.clickHouse.LogControlPlane(context.Background(), entry); err != nil {
				l.slogger.Error("Failed to log to ClickHouse", "error", err)
			}
		}()
	}
}

// extractContextFields extracts logging fields from context
func (l *StructuredLogger) extractContextFields(ctx context.Context) []slog.Attr {
	var fields []slog.Attr

	if traceID := getStringFromContext(ctx, TraceIDKey, ""); traceID != "" {
		fields = append(fields, slog.String("trace_id", traceID))
	}
	if userID := getStringFromContext(ctx, UserIDKey, ""); userID != "" {
		fields = append(fields, slog.String("user_id", userID))
	}
	if orgID := getStringFromContext(ctx, OrgIDKey, ""); orgID != "" {
		fields = append(fields, slog.String("org_id", orgID))
	}
	if workspaceID := getStringFromContext(ctx, WorkspaceIDKey, ""); workspaceID != "" {
		fields = append(fields, slog.String("workspace_id", workspaceID))
	}

	// Add service and component
	fields = append(fields,
		slog.String("service", l.service),
		slog.String("component", l.component),
	)

	return fields
}

// Helper functions for context extraction
func getStringFromContext(ctx context.Context, key ContextKey, defaultValue string) string {
	if value := ctx.Value(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getUint16FromContext(ctx context.Context, key ContextKey, defaultValue uint16) uint16 {
	if value := ctx.Value(key); value != nil {
		if val, ok := value.(uint16); ok {
			return val
		}
		if val, ok := value.(int); ok {
			return uint16(val)
		}
	}
	return defaultValue
}

func getUint32FromContext(ctx context.Context, key ContextKey, defaultValue uint32) uint32 {
	if value := ctx.Value(key); value != nil {
		if val, ok := value.(uint32); ok {
			return val
		}
		if val, ok := value.(int); ok {
			return uint32(val)
		}
		if val, ok := value.(int64); ok {
			return uint32(val)
		}
	}
	return defaultValue
}

func convertAttrsToMap(attrs []slog.Attr) map[string]string {
	result := make(map[string]string)
	for _, attr := range attrs {
		result[attr.Key] = attr.Value.String()
	}
	return result
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		traceID = uuid.New().String()
	}
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithOrgID adds an organization ID to the context
func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, OrgIDKey, orgID)
}

// WithWorkspaceID adds a workspace ID to the context
func WithWorkspaceID(ctx context.Context, workspaceID string) context.Context {
	return context.WithValue(ctx, WorkspaceIDKey, workspaceID)
}

// WithHTTPInfo adds HTTP request information to the context
func WithHTTPInfo(ctx context.Context, method, path string, status int, duration time.Duration) context.Context {
	ctx = context.WithValue(ctx, HTTPMethodKey, method)
	ctx = context.WithValue(ctx, HTTPPathKey, path)
	ctx = context.WithValue(ctx, HTTPStatusKey, status)
	ctx = context.WithValue(ctx, DurationKey, uint32(duration.Milliseconds()))
	return ctx
}