package logging

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggingMiddleware creates a Gin middleware for structured logging
func LoggingMiddleware(logger *StructuredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate trace ID if not present
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
			c.Header("X-Trace-ID", traceID)
		}

		// Add contextual information to request context
		ctx := c.Request.Context()
		ctx = WithTraceID(ctx, traceID)

		// Extract user information from headers or JWT
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			ctx = WithUserID(ctx, userID)
		}
		if orgID := c.GetHeader("X-Org-ID"); orgID != "" {
			ctx = WithOrgID(ctx, orgID)
		}
		if workspaceID := c.GetHeader("X-Workspace-ID"); workspaceID != "" {
			ctx = WithWorkspaceID(ctx, workspaceID)
		}

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log the request
		logger.LogHTTP(
			ctx,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			slog.String("remote_addr", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Int64("response_size", int64(c.Writer.Size())),
			slog.String("referer", c.Request.Referer()),
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error(ctx, "Request error", err.Err,
					slog.Int("error_type", int(err.Type)),
					slog.Any("error_meta", err.Meta),
				)
			}
		}
	}
}

// RecoveryMiddleware creates a Gin middleware for panic recovery with logging
func RecoveryMiddleware(logger *StructuredLogger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		ctx := c.Request.Context()
		
		logger.Error(ctx, "Panic recovered",
			nil,
			slog.Any("panic", recovered),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("remote_addr", c.ClientIP()),
		)

		// Log to security events as well
		logger.LogSecurity(ctx, "panic", "high", "Application panic recovered",
			slog.Any("panic_value", recovered),
			slog.String("endpoint", c.Request.Method+" "+c.Request.URL.Path),
		)

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"trace_id": ctx.Value(TraceIDKey),
		})
	})
}

// AuthLoggingMiddleware logs authentication events
func AuthLoggingMiddleware(logger *StructuredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Log authentication attempt
		logger.Info(ctx, "Authentication attempt",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("remote_addr", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)

		c.Next()

		// Log authentication result
		status := c.Writer.Status()
		if status == http.StatusUnauthorized {
			logger.LogSecurity(ctx, "auth_failure", "medium", "Authentication failed",
				slog.String("endpoint", c.Request.Method+" "+c.Request.URL.Path),
				slog.String("reason", "unauthorized"),
			)
		} else if status == http.StatusForbidden {
			logger.LogSecurity(ctx, "auth_failure", "medium", "Authorization failed",
				slog.String("endpoint", c.Request.Method+" "+c.Request.URL.Path),
				slog.String("reason", "forbidden"),
			)
		} else if status >= 200 && status < 300 {
			logger.LogSecurity(ctx, "auth_success", "low", "Authentication successful",
				slog.String("endpoint", c.Request.Method+" "+c.Request.URL.Path),
			)
		}
	}
}

// AuditMiddleware logs audit events for sensitive operations
func AuditMiddleware(logger *StructuredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Determine if this is an audit-worthy operation
		method := c.Request.Method
		isAuditWorthy := method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH"

		if isAuditWorthy {
			logger.LogSecurity(ctx, "audit", "low", "Sensitive operation performed",
				slog.String("method", method),
				slog.String("path", c.Request.URL.Path),
				slog.String("remote_addr", c.ClientIP()),
				slog.String("user_agent", c.Request.UserAgent()),
			)
		}

		c.Next()
	}
}

// PerformanceMiddleware logs performance metrics
func PerformanceMiddleware(logger *StructuredLogger, clickHouse *ClickHouseLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		
		// Log slow requests
		if duration > time.Second {
			ctx := c.Request.Context()
			logger.Warn(ctx, "Slow request detected",
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.Int64("duration_ms", duration.Milliseconds()),
				slog.Int("status", c.Writer.Status()),
			)
		}

		// Log performance metrics to ClickHouse (async)
		if clickHouse != nil {
			go func() {
				ctx := context.Background()
				
				// You could implement performance metrics logging here
				// This is a placeholder for performance metric collection
				_ = ctx
			}()
		}
	}
}

// SecurityHeadersMiddleware adds security headers and logs security events
func SecurityHeadersMiddleware(logger *StructuredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Check for security headers
		if c.GetHeader("X-Forwarded-Proto") == "http" {
			logger.LogSecurity(ctx, "insecure_request", "medium", "HTTP request to HTTPS endpoint",
				slog.String("proto", "http"),
				slog.String("path", c.Request.URL.Path),
			)
		}

		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// RateLimitLoggingMiddleware logs rate limiting events
func RateLimitLoggingMiddleware(logger *StructuredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log if rate limited
		if c.Writer.Status() == http.StatusTooManyRequests {
			ctx := c.Request.Context()
			logger.LogSecurity(ctx, "rate_limit", "medium", "Rate limit exceeded",
				slog.String("remote_addr", c.ClientIP()),
				slog.String("path", c.Request.URL.Path),
				slog.String("user_agent", c.Request.UserAgent()),
			)
		}
	}
}