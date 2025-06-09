package runner

import (
	"context"
	"fmt"
)

// Runner interface for running functions locally
type Runner interface {
	Run(ctx context.Context, opts RunOptions) error
}

// RunOptions contains options for running a function
type RunOptions struct {
	Path    string
	Port    int
	Env     map[string]string
	Handler string
	Watch   bool
}

// New creates a new runner for the given runtime
func New(runtime string) (Runner, error) {
	switch runtime {
	case "node", "python", "go", "java", "dotnet", "ruby", "php", "rust":
		return &genericRunner{runtime: runtime}, nil
	default:
		return nil, fmt.Errorf("unsupported runtime: %s", runtime)
	}
}

// genericRunner is a basic implementation of Runner
type genericRunner struct {
	runtime string
}

// Run starts the function locally
func (r *genericRunner) Run(ctx context.Context, opts RunOptions) error {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Start the appropriate runtime
	// 2. Set up environment variables
	// 3. Start HTTP server
	// 4. Handle hot reloading if watch is enabled
	return fmt.Errorf("local runner not yet implemented for %s", r.runtime)
}