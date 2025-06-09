package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/client"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var describeNamespace string

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe [function-name]",
	Short: "Show detailed information about a function",
	Long: `Display detailed information about a deployed function.

This includes:
  - Basic information (name, namespace, status)
  - Runtime configuration
  - Resource limits and requests
  - Autoscaling configuration
  - Environment variables
  - Recent revisions
  - Traffic distribution`,
	Args: cobra.ExactArgs(1),
	RunE: runDescribe,
}

func init() {
	describeCmd.Flags().StringVarP(&describeNamespace, "namespace", "n", "default", "function namespace")
}

func runDescribe(cmd *cobra.Command, args []string) error {
	functionName := args[0]
	ctx := context.Background()

	// Create API client
	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Get function details
	function, err := apiClient.GetFunction(ctx, functionName)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}

	// Convert to FunctionDetails for display
	details := &client.FunctionDetails{
		FunctionInfo: *function,
	}

	// Display function information
	fmt.Printf("Name:      %s\n", details.Name)
	fmt.Printf("Namespace: %s\n", details.Namespace)
	fmt.Printf("Status:    %s\n", getStatusString(details))
	fmt.Printf("URL:       %s\n", details.URL)
	fmt.Printf("Created:   %s\n", details.Created.Format("2006-01-02 15:04:05"))

	// Runtime information
	fmt.Println("\nRuntime:")
	fmt.Printf("  Image:   %s\n", function.Image)
	// TODO: Add timeout, concurrency, and resources when available in FunctionInfo
	// fmt.Printf("  Timeout: %ds\n", function.Timeout)
	// if function.Concurrency > 0 {
	// 	fmt.Printf("  Concurrency: %d\n", function.Concurrency)
	// }

	// TODO: Add autoscaling details when available
	// if function.Autoscaling != nil {
	// 	fmt.Println("\nAutoscaling:")
	// 	fmt.Printf("  Min Scale: %d\n", function.Autoscaling.MinScale)
	// 	fmt.Printf("  Max Scale: %d\n", function.Autoscaling.MaxScale)
	// }

	// TODO: Add environment variables when available
	// if len(function.Environment) > 0 {
	// 	fmt.Println("\nEnvironment Variables:")
	// 	for k, v := range function.Environment {
	// 		fmt.Printf("  %s: %s\n", k, v)
	// 	}
	// }

	// Labels
	if len(details.Labels) > 0 {
		fmt.Println("\nLabels:")
		for k, v := range details.Labels {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	// Annotations
	if verbose && len(details.Annotations) > 0 {
		fmt.Println("\nAnnotations:")
		for k, v := range details.Annotations {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	// TODO: Add revisions when available
	// if len(details.Revisions) > 0 {
		fmt.Println("\nRevisions:")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"NAME", "READY", "TRAFFIC", "CREATED"})
		table.SetBorder(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		
	// 	for _, rev := range details.Revisions {
	// 		traffic := "-"
	// 		if rev.TrafficPercent > 0 {
	// 			traffic = fmt.Sprintf("%d%%", rev.TrafficPercent)
	// 		}
	// 		ready := "No"
	// 		if rev.Ready {
	// 			ready = "Yes"
	// 		}
	// 		table.Append([]string{
	// 			rev.Name,
	// 			ready,
	// 			traffic,
	// 			rev.CreatedAt.Format("2006-01-02 15:04"),
	// 		})
	// 	}
	// 	table.Render()
	// }

	// TODO: Add traffic distribution when available
	// if len(function.Traffic) > 0 {
	// 	fmt.Println("\nTraffic Distribution:")
	// 	for _, t := range function.Traffic {
	// 		if t.Tag != "" {
	// 			fmt.Printf("  %s (%s): %d%%\n", t.RevisionName, t.Tag, t.Percent)
	// 		} else {
	// 			fmt.Printf("  %s: %d%%\n", t.RevisionName, t.Percent)
	// 		}
	// 	}
	// }

	// Events (if any)
	if len(details.Events) > 0 {
		fmt.Println("\nEvent Triggers:")
		for _, e := range details.Events {
			fmt.Printf("  Type: %s, Source: %s\n", e.Type, e.Source)
		}
	}

	return nil
}

func getStatusString(details *client.FunctionDetails) string {
	if details.Ready {
		return "Ready"
	}
	if details.Error != "" {
		return fmt.Sprintf("Error: %s", details.Error)
	}
	return "Not Ready"
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	sensitivePatterns := []string{
		"password",
		"secret",
		"key",
		"token",
		"credential",
		"api_key",
		"private",
	}
	
	for _, pattern := range sensitivePatterns {
		if strings.Contains(key, pattern) {
			return true
		}
	}
	return false
}