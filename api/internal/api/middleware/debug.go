package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// responseWriter wraps gin.ResponseWriter to capture response
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// DebugAuth logs authentication attempts for debugging
func DebugAuth(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.Mode() != gin.DebugMode {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			logger.Debug("Auth attempt",
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"has_auth", authHeader != "",
				"token_prefix", getTokenPrefix(token),
			)
		}

		c.Next()
	}
}

// DebugRequest logs detailed request/response information
func DebugRequest(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.Mode() != gin.DebugMode {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Log request headers
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			headers[key] = strings.Join(values, ", ")
		}

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(requestBody))
		}

		// Log request
		logger.Debug("HTTP Request",
			"method", c.Request.Method,
			"path", path,
			"query", raw,
			"headers", headers,
			"body", string(requestBody),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Wrap response writer to capture response
		w := &responseWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log response
		logger.Debug("HTTP Response",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"latency_ms", float64(latency.Nanoseconds())/1e6,
			"body_size", w.body.Len(),
			"response_body", truncateBody(w.body.String(), 1000),
			"errors", c.Errors.String(),
		)

		// Log slow queries
		if latency > 100*time.Millisecond {
			logger.Warn("Slow request detected",
				"method", c.Request.Method,
				"path", path,
				"latency", latency.String(),
			)
		}
	}
}

// DebugDatabase logs database queries and execution time
func DebugDatabase(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.Mode() != gin.DebugMode {
			c.Next()
			return
		}

		// Store logger in context for database hooks
		c.Set("debug_logger", logger)
		c.Next()
	}
}

// DebugPanic recovers from panics and logs detailed stack traces
func DebugPanic(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"client_ip", c.ClientIP(),
					"stack_trace", string(debug.Stack()),
				)

				// Return error response
				c.JSON(500, gin.H{
					"error": "Internal server error",
					"debug": fmt.Sprintf("%v", err),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// DebugMetrics collects and logs request metrics
func DebugMetrics(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.Mode() != gin.DebugMode {
			c.Next()
			return
		}

		// Collect metrics
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Log metrics
		logger.Debug("Request metrics",
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"status", c.Writer.Status(),
			"duration_ms", float64(duration.Nanoseconds())/1e6,
			"body_size", c.Writer.Size(),
			"errors", len(c.Errors),
		)
	}
}

// Helper functions
func getTokenPrefix(token string) string {
	if len(token) > 20 {
		return token[:20] + "..."
	}
	return token
}

func truncateBody(body string, maxLen int) string {
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "... (truncated)"
}

// DebugJSON pretty prints JSON payloads for debugging
func DebugJSON(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err)
	}
	return string(b)
}