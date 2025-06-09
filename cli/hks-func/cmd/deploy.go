package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/builder"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/client"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/function"
	"github.com/spf13/cobra"
)

var (
	deployNamespace string
	deployTag       string
	deployDryRun    bool
	deployForce     bool
	deployNoTraffic bool
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy function to Hexabase AI",
	Long: `Deploy a function to the Hexabase AI platform.

This command:
  1. Builds the function container image
  2. Pushes the image to the registry
  3. Creates or updates the Knative service
  4. Optionally routes traffic to the new revision

The function must have a valid function.yaml configuration file.`,
	RunE: runDeploy,
}

func init() {
	deployCmd.Flags().StringVarP(&deployNamespace, "namespace", "n", "", "target namespace (default from function.yaml)")
	deployCmd.Flags().StringVarP(&deployTag, "tag", "t", "", "image tag (default: latest)")
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "print deployment manifest without applying")
	deployCmd.Flags().BoolVar(&deployForce, "force", false, "force deployment even if no changes detected")
	deployCmd.Flags().BoolVar(&deployNoTraffic, "no-traffic", false, "deploy without routing traffic to new revision")
}

func runDeploy(cmd *cobra.Command, args []string) error {
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

	// Override namespace if specified
	if deployNamespace != "" {
		config.Deploy.Namespace = deployNamespace
	}

	// Set default tag if not specified
	if deployTag == "" {
		deployTag = "latest"
	}

	printInfo("Deploying function '%s' to namespace '%s'", config.Name, config.Deploy.Namespace)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid function configuration: %w", err)
	}

	// Check authentication
	if err := checkAuth(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Build stage
	if !deployDryRun {
		if err := buildFunction(ctx, config, workspaceRoot, deployTag); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
	}

	// Deploy stage
	if err := deployFunction(ctx, config, deployTag); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	if !deployDryRun {
		printSuccess("Function deployed successfully!")
		
		// Get function URL
		apiClient, err := client.NewAPIClient()
		if err == nil {
			fn, err := apiClient.GetFunction(ctx, config.Name)
			if err == nil && fn.URL != "" {
				fmt.Println()
				printInfo("Function URL: %s", fn.URL)
			}
		}
		
		// Show next steps
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  - Test your function: hks-func invoke\n")
		fmt.Printf("  - View logs: hks-func logs\n")
		fmt.Printf("  - Check status: hks-func describe\n")
	}

	return nil
}

func buildFunction(ctx context.Context, config *function.Config, workspaceRoot, tag string) error {
	// Create spinner
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " Building function..."
	s.Start()
	defer s.Stop()

	// Create builder
	b, err := builder.New(config.Build.Builder)
	if err != nil {
		return err
	}

	// Build options
	opts := builder.BuildOptions{
		Path:       workspaceRoot,
		Name:       config.Name,
		Tag:        tag,
		Runtime:    config.Runtime,
		Handler:    config.Handler,
		BuildArgs:  config.Build.BuildArgs,
		Dockerfile: config.Build.Dockerfile,
	}

	// Build image
	image, err := b.Build(ctx, opts)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	s.Stop()
	printSuccess("Built image: %s", image)

	// Push image
	if config.Deploy.Registry != "" {
		s.Suffix = " Pushing image..."
		s.Start()

		// Build with push option
		opts.Push = true
		pushImage, err := b.Build(ctx, opts)
		if err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		s.Stop()
		printSuccess("Pushed image: %s", pushImage)
	}

	return nil
}


func checkAuth() error {
	// Check for authentication
	token := os.Getenv("HKS_API_TOKEN")
	if token == "" {
		// Try to read from config
		configPath := filepath.Join(os.Getenv("HOME"), ".hks-func", "config.yaml")
		if _, err := os.Stat(configPath); err != nil {
			return fmt.Errorf("not authenticated. Run 'hks-func config auth' to login")
		}
	}
	return nil
}
