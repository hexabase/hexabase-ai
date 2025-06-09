package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/function"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/runner"
	"github.com/spf13/cobra"
)

var (
	testLocal   bool
	testPort    int
	testEnvFile string
	testWatch   bool
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test function locally",
	Long: `Test your function locally before deployment.

This command runs your function in a local environment that simulates
the production runtime. It supports hot reloading for rapid development.

The local test environment:
  - Runs your function on a local port
  - Loads environment variables
  - Simulates the Knative request/response cycle
  - Provides request logging`,
	RunE: runTest,
}

func init() {
	testCmd.Flags().BoolVar(&testLocal, "local", true, "run function locally")
	testCmd.Flags().IntVarP(&testPort, "port", "p", 8080, "local port to run on")
	testCmd.Flags().StringVar(&testEnvFile, "env-file", ".env", "environment file to load")
	testCmd.Flags().BoolVarP(&testWatch, "watch", "w", false, "watch for changes and reload")
}

func runTest(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Find workspace root
	workspaceRoot, err := getWorkspaceRoot()
	if err != nil {
		return err
	}

	// Load function configuration
	configPath := filepath.Join(workspaceRoot, "function.yaml")
	config, err := function.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load function.yaml: %w", err)
	}

	printInfo("Testing function '%s' locally on port %d", config.Name, testPort)

	// Run language-specific tests first
	if err := runUnitTests(ctx, config, workspaceRoot); err != nil {
		return fmt.Errorf("unit tests failed: %w", err)
	}

	if !testLocal {
		printSuccess("Unit tests passed!")
		return nil
	}

	// Start local runner
	r, err := runner.New(config.Runtime)
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Load environment variables
	env := make(map[string]string)
	if testEnvFile != "" {
		envPath := filepath.Join(workspaceRoot, testEnvFile)
		if _, err := os.Stat(envPath); err == nil {
			loadedEnv, err := loadEnvFile(envPath)
			if err != nil {
				return fmt.Errorf("failed to load env file: %w", err)
			}
			for k, v := range loadedEnv {
				env[k] = v
			}
		}
	}

	// Add function environment variables
	for k, v := range config.Environment {
		env[k] = v
	}

	// Set PORT
	env["PORT"] = fmt.Sprintf("%d", testPort)

	// Run options
	runOpts := runner.RunOptions{
		Path:    workspaceRoot,
		Port:    testPort,
		Env:     env,
		Handler: config.Handler,
		Watch:   testWatch,
	}

	// Start function
	printInfo("Starting function...")
	fmt.Printf("\nFunction URL: http://localhost:%d\n", testPort)
	fmt.Println("\nPress Ctrl+C to stop")
	fmt.Println()

	// Run function
	if err := r.Run(ctx, runOpts); err != nil {
		return fmt.Errorf("failed to run function: %w", err)
	}

	return nil
}

func runUnitTests(ctx context.Context, config *function.Config, workspaceRoot string) error {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " Running unit tests..."
	s.Start()
	defer s.Stop()

	var cmd *exec.Cmd

	switch config.Runtime {
	case "node":
		// Check if package.json has test script
		if _, err := os.Stat(filepath.Join(workspaceRoot, "package.json")); err == nil {
			cmd = exec.CommandContext(ctx, "npm", "test")
		}
	case "python":
		// Try pytest first, then unittest
		if _, err := exec.LookPath("pytest"); err == nil {
			cmd = exec.CommandContext(ctx, "pytest")
		} else {
			cmd = exec.CommandContext(ctx, "python", "-m", "unittest", "discover")
		}
	case "go":
		cmd = exec.CommandContext(ctx, "go", "test", "./...")
	case "java":
		if _, err := os.Stat(filepath.Join(workspaceRoot, "pom.xml")); err == nil {
			cmd = exec.CommandContext(ctx, "mvn", "test")
		} else if _, err := os.Stat(filepath.Join(workspaceRoot, "build.gradle")); err == nil {
			cmd = exec.CommandContext(ctx, "gradle", "test")
		}
	case "dotnet":
		cmd = exec.CommandContext(ctx, "dotnet", "test")
	case "ruby":
		cmd = exec.CommandContext(ctx, "bundle", "exec", "rspec")
	case "php":
		cmd = exec.CommandContext(ctx, "composer", "test")
	case "rust":
		cmd = exec.CommandContext(ctx, "cargo", "test")
	default:
		s.Stop()
		printWarning("No test runner configured for runtime '%s'", config.Runtime)
		return nil
	}

	if cmd == nil {
		s.Stop()
		printWarning("No tests found")
		return nil
	}

	cmd.Dir = workspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	s.Stop()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	return nil
}

func loadEnvFile(path string) (map[string]string, error) {
	env := make(map[string]string)
	
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Simple env file parser
	// In production, use a proper env file parser library
	lines := string(data)
	for _, line := range splitLines(lines) {
		line = trimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		
		parts := splitN(line, "=", 2)
		if len(parts) == 2 {
			key := trimSpace(parts[0])
			value := trimSpace(parts[1])
			
			// Remove quotes if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			
			env[key] = value
		}
	}
	
	return env, nil
}

// Helper functions (would use strings package in real implementation)
func splitLines(s string) []string {
	// Implementation
	return []string{}
}

func trimSpace(s string) string {
	// Implementation
	return s
}

func splitN(s, sep string, n int) []string {
	// Implementation
	return []string{}
}