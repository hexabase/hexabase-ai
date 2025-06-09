package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/client"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	listNamespace string
	listAll       bool
	listLimit     int
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List deployed functions",
	Long: `List all deployed functions in the specified namespace.

This command displays:
  - Function name
  - Status (Ready/Not Ready)
  - Created time
  - URL
  - Latest revision`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&listNamespace, "namespace", "n", "default", "namespace to list functions from")
	listCmd.Flags().BoolVarP(&listAll, "all-namespaces", "A", false, "list functions from all namespaces")
	listCmd.Flags().IntVar(&listLimit, "limit", 100, "maximum number of functions to list")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create API client
	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// List functions
	var functions []client.Function
	if listAll {
		functions, err = apiClient.ListAllFunctions(ctx, listLimit)
	} else {
		functions, err = apiClient.ListFunctions(ctx, listNamespace, listLimit)
	}
	
	if err != nil {
		return fmt.Errorf("failed to list functions: %w", err)
	}

	if len(functions) == 0 {
		if listAll {
			printInfo("No functions found in any namespace")
		} else {
			printInfo("No functions found in namespace '%s'", listNamespace)
		}
		return nil
	}

	// Display in table format
	if output == "json" {
		return outputJSON(functions)
	} else if output == "yaml" {
		return outputYAML(functions)
	}

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	headers := []string{"NAME", "STATUS", "AGE", "URL"}
	if listAll {
		headers = append([]string{"NAMESPACE"}, headers...)
	}
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	// Add rows
	for _, fn := range functions {
		age := formatAge(fn.Created)
		status := formatStatus(fn.Ready)
		url := fn.URL
		if url == "" {
			url = "-"
		}

		row := []string{fn.Name, status, age, url}
		if listAll {
			row = append([]string{fn.Namespace}, row...)
		}
		table.Append(row)
	}

	table.Render()

	// Show summary
	if verbose {
		fmt.Println()
		printInfo("Total functions: %d", len(functions))
	}

	return nil
}

func formatAge(createdAt time.Time) string {
	duration := time.Since(createdAt)
	
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
}

func formatStatus(ready bool) string {
	if ready {
		return "Ready"
	}
	return "Not Ready"
}

func outputJSON(data interface{}) error {
	// Implementation would marshal to JSON
	fmt.Println("JSON output not yet implemented")
	return nil
}

func outputYAML(data interface{}) error {
	// Implementation would marshal to YAML
	fmt.Println("YAML output not yet implemented")
	return nil
}