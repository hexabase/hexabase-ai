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
	function, err := apiClient.GetFunction(ctx, functionName, describeNamespace)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}

	// Display function information
	fmt.Printf("Name:      %s\n", function.Name)
	fmt.Printf("Namespace: %s\n", function.Namespace)
	fmt.Printf("Status:    %s\n", getStatusString(function))
	fmt.Printf("URL:       %s\n", function.URL)
	fmt.Printf("Created:   %s\n", function.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:   %s\n", function.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Runtime information
	fmt.Println("\nRuntime:")
	fmt.Printf("  Image:   %s\n", function.Image)
	fmt.Printf("  Timeout: %ds\n", function.Timeout)
	if function.Concurrency > 0 {
		fmt.Printf("  Concurrency: %d\n", function.Concurrency)
	}

	// Resources
	if function.Resources != nil {
		fmt.Println("\nResources:")
		if function.Resources.Requests != nil {
			fmt.Println("  Requests:")
			if function.Resources.Requests.CPU != "" {
				fmt.Printf("    CPU:    %s\n", function.Resources.Requests.CPU)
			}
			if function.Resources.Requests.Memory != "" {
				fmt.Printf("    Memory: %s\n", function.Resources.Requests.Memory)
			}
		}
		if function.Resources.Limits != nil {
			fmt.Println("  Limits:")
			if function.Resources.Limits.CPU != "" {
				fmt.Printf("    CPU:    %s\n", function.Resources.Limits.CPU)
			}
			if function.Resources.Limits.Memory != "" {
				fmt.Printf("    Memory: %s\n", function.Resources.Limits.Memory)
			}
		}
	}

	// Autoscaling
	if function.Autoscaling != nil {
		fmt.Println("\nAutoscaling:")
		fmt.Printf("  Min Scale: %d\n", function.Autoscaling.MinScale)
		fmt.Printf("  Max Scale: %d\n", function.Autoscaling.MaxScale)
		fmt.Printf("  Metric:    %s\n", function.Autoscaling.Metric)
		fmt.Printf("  Target:    %d\n", function.Autoscaling.Target)
	}

	// Environment variables
	if len(function.Environment) > 0 {
		fmt.Println("\nEnvironment Variables:")
		for k, v := range function.Environment {
			// Mask sensitive values
			displayValue := v
			if isSensitiveKey(k) {
				displayValue = "***"
			}
			fmt.Printf("  %s: %s\n", k, displayValue)
		}
	}

	// Labels
	if len(function.Labels) > 0 {
		fmt.Println("\nLabels:")
		for k, v := range function.Labels {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	// Annotations
	if verbose && len(function.Annotations) > 0 {
		fmt.Println("\nAnnotations:")
		for k, v := range function.Annotations {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	// Revisions
	if len(function.Revisions) > 0 {
		fmt.Println("\nRevisions:")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"NAME", "READY", "TRAFFIC", "CREATED"})
		table.SetBorder(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		
		for _, rev := range function.Revisions {
			traffic := "-"
			if rev.TrafficPercent > 0 {
				traffic = fmt.Sprintf("%d%%", rev.TrafficPercent)
			}
			ready := "No"
			if rev.Ready {
				ready = "Yes"
			}
			table.Append([]string{
				rev.Name,
				ready,
				traffic,
				rev.CreatedAt.Format("2006-01-02 15:04"),
			})
		}
		table.Render()
	}

	// Traffic distribution
	if len(function.Traffic) > 0 {
		fmt.Println("\nTraffic Distribution:")
		for _, t := range function.Traffic {
			if t.Tag != "" {
				fmt.Printf("  %s (%s): %d%%\n", t.RevisionName, t.Tag, t.Percent)
			} else {
				fmt.Printf("  %s: %d%%\n", t.RevisionName, t.Percent)
			}
		}
	}

	// Events (if any)
	if len(function.Events) > 0 {
		fmt.Println("\nEvent Triggers:")
		for _, e := range function.Events {
			fmt.Printf("  Type: %s, Source: %s\n", e.Type, e.Source)
		}
	}

	return nil
}

func getStatusString(function *client.FunctionDetails) string {
	if function.Ready {
		return "Ready"
	}
	if function.Error != "" {
		return fmt.Sprintf("Error: %s", function.Error)
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