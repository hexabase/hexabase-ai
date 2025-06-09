package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/client"
	"github.com/spf13/cobra"
)

var (
	logsNamespace string
	logsFollow    bool
	logsTail      int
	logsSince     string
	logsContainer string
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [function-name]",
	Short: "View function logs",
	Long: `View logs from a deployed function.

This command retrieves logs from function containers and displays them
with timestamps and color coding for different log levels.

Examples:
  # View recent logs
  hks-func logs my-function

  # Follow logs in real-time
  hks-func logs my-function -f

  # View last 100 lines
  hks-func logs my-function --tail 100

  # View logs from last hour
  hks-func logs my-function --since 1h`,
	Args: cobra.ExactArgs(1),
	RunE: runLogs,
}

func init() {
	logsCmd.Flags().StringVarP(&logsNamespace, "namespace", "n", "default", "function namespace")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "follow log output")
	logsCmd.Flags().IntVar(&logsTail, "tail", 50, "number of recent lines to show")
	logsCmd.Flags().StringVar(&logsSince, "since", "", "show logs since duration (e.g. 10m, 1h)")
	logsCmd.Flags().StringVarP(&logsContainer, "container", "c", "", "specific container to get logs from")
}

func runLogs(cmd *cobra.Command, args []string) error {
	functionName := args[0]
	ctx := context.Background()

	// Create API client
	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Parse since duration
	var since *time.Duration
	if logsSince != "" {
		duration, err := time.ParseDuration(logsSince)
		if err != nil {
			return fmt.Errorf("invalid since duration: %w", err)
		}
		since = &duration
	}

	// Get logs options
	opts := client.LogOptions{
		Namespace: logsNamespace,
		Container: logsContainer,
		Follow:    logsFollow,
		Tail:      logsTail,
		Since:     since,
	}

	printInfo("Fetching logs for function '%s' in namespace '%s'", functionName, logsNamespace)

	// Get log stream
	logStream, err := apiClient.GetLogs(ctx, functionName, opts)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}
	defer logStream.Close()

	// Read and display logs
	return streamLogs(logStream)
}

func streamLogs(reader io.ReadCloser) error {
	// Create a scanner to read line by line
	buf := make([]byte, 1024)
	
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		
		if n > 0 {
			// Process and colorize log lines
			lines := string(buf[:n])
			colorizedLines := colorizeLogs(lines)
			fmt.Print(colorizedLines)
		}
	}
}

func colorizeLogs(logs string) string {
	// This is a simplified colorizer
	// In a real implementation, you would parse log lines more carefully
	
	lines := logs
	
	// Colorize timestamps (assuming ISO format)
	lines = color.New(color.FgHiBlack).Sprint(lines)
	
	// Colorize log levels
	// ERROR in red
	if contains := color.New(color.FgRed); contains != nil {
		// Would replace ERROR patterns
	}
	
	// WARN in yellow
	if contains := color.New(color.FgYellow); contains != nil {
		// Would replace WARN patterns
	}
	
	// INFO in cyan
	if contains := color.New(color.FgCyan); contains != nil {
		// Would replace INFO patterns
	}
	
	return lines
}

// LogEntry represents a parsed log entry
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Container string
	Pod       string
}