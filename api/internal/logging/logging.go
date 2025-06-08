// Package logging provides standardized logging configuration for the application.
package logging

import (
	"log/slog"
	"os"
)

// NewLogger creates and returns a new structured logger (slog).
// It is configured to write JSON logs to standard output.
func NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		// AddSource: true, // Uncomment for development to see source file and line
		Level: slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return logger
}
