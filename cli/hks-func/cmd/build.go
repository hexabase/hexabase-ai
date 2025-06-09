package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/builder"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/function"
	"github.com/spf13/cobra"
)

var (
	buildTag      string
	buildPush     bool
	buildPlatform string
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build function container image",
	Long: `Build a container image for your function.

This command builds a container image using the configured builder
(pack, docker, ko, etc.) and optionally pushes it to a registry.

The build process:
  1. Validates the function configuration
  2. Detects or uses specified builder
  3. Builds the container image
  4. Tags the image appropriately
  5. Optionally pushes to registry`,
	RunE: runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&buildTag, "tag", "t", "latest", "image tag")
	buildCmd.Flags().BoolVar(&buildPush, "push", false, "push image to registry after build")
	buildCmd.Flags().StringVar(&buildPlatform, "platform", "", "target platform (e.g., linux/amd64)")
}

func runBuild(cmd *cobra.Command, args []string) error {
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

	printInfo("Building function '%s'", config.Name)

	// Create spinner
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " Building container image..."
	s.Start()
	defer s.Stop()

	// Create builder
	b, err := builder.New(config.Build.Builder)
	if err != nil {
		return fmt.Errorf("failed to create builder: %w", err)
	}

	// Build options
	opts := builder.BuildOptions{
		Path:       workspaceRoot,
		Name:       config.Name,
		Tag:        buildTag,
		Runtime:    config.Runtime,
		Handler:    config.Handler,
		BuildArgs:  config.Build.BuildArgs,
		Dockerfile: config.Build.Dockerfile,
		Platform:   buildPlatform,
		Registry:   config.Deploy.Registry,
		Push:       buildPush,
	}

	// Build image
	image, err := b.Build(ctx, opts)
	if err != nil {
		s.Stop()
		return fmt.Errorf("build failed: %w", err)
	}

	s.Stop()
	
	if buildPush {
		printSuccess("Built and pushed image: %s", image)
	} else {
		printSuccess("Built image: %s", image)
	}

	// Show next steps
	fmt.Println("\nNext steps:")
	if !buildPush && config.Deploy.Registry != "" {
		fmt.Printf("  - Push image: hks-func build --push\n")
	}
	fmt.Printf("  - Deploy function: hks-func deploy --tag %s\n", buildTag)

	return nil
}