package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/client"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/deployer"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/function"
)

func deployFunction(ctx context.Context, config *function.Config, tag string) error {
	// Create API client
	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Create deployer
	d := deployer.New(apiClient.KnativeClient, config.Deploy.Namespace)

	// Build image name
	image := fmt.Sprintf("%s:%s", config.GetImageName(), tag)

	if deployDryRun {
		// Just print what would be deployed
		fmt.Println("Would deploy:")
		fmt.Printf("  Name: %s\n", config.Name)
		fmt.Printf("  Namespace: %s\n", config.Deploy.Namespace)
		fmt.Printf("  Image: %s\n", image)
		fmt.Printf("  Runtime: %s\n", config.Runtime)
		fmt.Printf("  Handler: %s\n", config.Handler)
		fmt.Printf("  Autoscaling:\n")
		fmt.Printf("    Min Scale: %d\n", config.Deploy.Autoscaling.MinScale)
		fmt.Printf("    Max Scale: %d\n", config.Deploy.Autoscaling.MaxScale)
		fmt.Printf("  Resources:\n")
		fmt.Printf("    CPU: %s\n", config.Deploy.Resources.CPU)
		fmt.Printf("    Memory: %s\n", config.Deploy.Resources.Memory)
		return nil
	}

	// Create spinner
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " Deploying to Knative..."
	s.Start()
	defer s.Stop()

	// Deploy
	if err := d.Deploy(ctx, config, image); err != nil {
		return err
	}

	return nil
}