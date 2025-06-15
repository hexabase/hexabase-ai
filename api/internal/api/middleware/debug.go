package middleware

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
)

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

func getTokenPrefix(token string) string {
	if len(token) > 20 {
		return token[:20] + "..."
	}
	return token
}