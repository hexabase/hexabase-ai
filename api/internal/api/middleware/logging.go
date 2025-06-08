// Package middleware provides Gin middleware functions.
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CtxLoggerKey = "logger"

// StructuredLogger returns a Gin middleware that logs requests using slog.
// It also injects a logger with contextual information into the request context.
func StructuredLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// Generate a unique trace ID for this request.
		traceID := uuid.New().String()
		
		// Create a new logger with request-specific attributes.
		requestLogger := logger.With(
			"trace_id", traceID,
			"http_method", c.Request.Method,
			"path", path,
		)
		
		// Add the logger to the context for use in handlers.
		c.Set(CtxLoggerKey, requestLogger)

		// Process the request.
		c.Next()

		// After the request is handled, log the final details.
		stop := time.Now()
		latency := stop.Sub(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		logAttrs := []slog.Attr{
			slog.Int("status_code", statusCode),
			slog.String("latency", latency.String()),
			slog.String("client_ip", clientIP),
			slog.Int("data_length", dataLength),
			slog.String("user_agent", clientUserAgent),
		}

		if rawQuery != "" {
			logAttrs = append(logAttrs, slog.String("query", rawQuery))
		}
		
		if len(c.Errors) > 0 {
			// Log as an error if there are any gin errors.
			logAttrs = append(logAttrs, slog.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()))
			requestLogger.LogAttrs(c.Request.Context(), slog.LevelError, "Request failed", logAttrs...)
		} else {
			// Log as info for successful requests.
			if statusCode >= 500 {
				requestLogger.LogAttrs(c.Request.Context(), slog.LevelError, "Server error", logAttrs...)
			} else if statusCode >= 400 {
				requestLogger.LogAttrs(c.Request.Context(), slog.LevelWarn, "Client error", logAttrs...)
			} else {
				requestLogger.LogAttrs(c.Request.Context(), slog.LevelInfo, "Request handled", logAttrs...)
			}
		}
	}
} 