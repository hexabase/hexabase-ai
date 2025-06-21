package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// InitializeLogging sets up the structured logging system with ClickHouse integration
func InitializeLogging() (*StructuredLogger, *ClickHouseLogger, error) {
	// Parse log level from environment
	logLevel := parseLogLevel(getEnv("LOG_LEVEL", "info"))

	// Create ClickHouse logger if enabled
	var clickHouseLogger *ClickHouseLogger
	if getEnv("CLICKHOUSE_ENABLED", "false") == "true" {
		clickHouseConfig := &ClickHouseConfig{
			Host:     getEnv("CLICKHOUSE_HOST", "clickhouse"),
			Port:     parseInt(getEnv("CLICKHOUSE_PORT", "9000")),
			Database: getEnv("CLICKHOUSE_DATABASE", "hexabase_logs"),
			Username: getEnv("CLICKHOUSE_USERNAME", "hexabase"),
			Password: getEnv("CLICKHOUSE_PASSWORD", ""),
			TLS:      getEnv("CLICKHOUSE_TLS", "false") == "true",
		}

		var err error
		clickHouseLogger, err = NewClickHouseLogger(clickHouseConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize ClickHouse logger: %w", err)
		}
	}

	// Create structured logger
	loggerConfig := &LoggerConfig{
		Service:          getEnv("SERVICE_NAME", "hexabase-api"),
		Component:        getEnv("COMPONENT_NAME", "main"),
		Level:            logLevel,
		AddSource:        getEnv("LOG_ADD_SOURCE", "true") == "true",
		EnableClickHouse: clickHouseLogger != nil,
	}

	structuredLogger := NewStructuredLogger(loggerConfig, clickHouseLogger)

	return structuredLogger, clickHouseLogger, nil
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// parseInt parses string to int with default value
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}