package cmd

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/client"
	"github.com/spf13/cobra"
)

var (
	deleteNamespace string
	deleteForce     bool
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [function-name]",
	Short: "Delete a deployed function",
	Long: `Delete a deployed function from the platform.

This command removes the function and all its associated resources,
including revisions and configurations. This action cannot be undone.`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteNamespace, "namespace", "n", "default", "function namespace")
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
	functionName := args[0]
	ctx := context.Background()

	// Confirm deletion
	if !deleteForce {
		confirm := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Are you sure you want to delete function '%s' in namespace '%s'?", functionName, deleteNamespace),
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirm); err != nil {
			return err
		}
		if !confirm {
			printInfo("Deletion cancelled")
			return nil
		}
	}

	// Create API client
	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	printInfo("Deleting function '%s'...", functionName)

	// Delete function
	if err := apiClient.DeleteFunction(ctx, functionName, deleteNamespace); err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}

	printSuccess("Function '%s' deleted successfully", functionName)
	return nil
}