package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/client"
	"github.com/spf13/cobra"
)

var (
	invokeNamespace string
	invokeData      string
	invokeFile      string
	invokeMethod    string
	invokePath      string
	invokeHeaders   []string
	invokeAsync     bool
)

// invokeCmd represents the invoke command
var invokeCmd = &cobra.Command{
	Use:   "invoke [function-name]",
	Short: "Invoke a deployed function",
	Long: `Invoke a deployed function with optional data.

This command sends an HTTP request to your function and displays the response.
You can provide input data as JSON, from a file, or from stdin.

Examples:
  # Simple invocation
  hks-func invoke my-function

  # With JSON data
  hks-func invoke my-function --data '{"name": "world"}'

  # From file
  hks-func invoke my-function --file input.json

  # With custom method and path
  hks-func invoke my-function --method POST --path /api/users

  # With headers
  hks-func invoke my-function --header "X-Custom: value"

  # Pipe data from stdin
  echo '{"test": true}' | hks-func invoke my-function`,
	Args: cobra.ExactArgs(1),
	RunE: runInvoke,
}

func init() {
	invokeCmd.Flags().StringVarP(&invokeNamespace, "namespace", "n", "default", "function namespace")
	invokeCmd.Flags().StringVarP(&invokeData, "data", "d", "", "JSON data to send")
	invokeCmd.Flags().StringVarP(&invokeFile, "file", "f", "", "file containing data to send")
	invokeCmd.Flags().StringVarP(&invokeMethod, "method", "m", "POST", "HTTP method")
	invokeCmd.Flags().StringVarP(&invokePath, "path", "p", "/", "request path")
	invokeCmd.Flags().StringArrayVarP(&invokeHeaders, "header", "H", []string{}, "custom headers (format: 'Key: Value')")
	invokeCmd.Flags().BoolVar(&invokeAsync, "async", false, "invoke asynchronously (fire and forget)")
}

func runInvoke(cmd *cobra.Command, args []string) error {
	functionName := args[0]
	ctx := context.Background()

	// Create API client
	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Prepare request data
	var requestData []byte
	
	// Priority: --data flag, --file flag, stdin
	if invokeData != "" {
		requestData = []byte(invokeData)
	} else if invokeFile != "" {
		data, err := os.ReadFile(invokeFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		requestData = data
	} else {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			requestData = data
		}
	}

	// Validate JSON if data provided
	if len(requestData) > 0 {
		var jsonData interface{}
		if err := json.Unmarshal(requestData, &jsonData); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
	}

	// Parse headers
	headers := make(map[string]string)
	for _, h := range invokeHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid header format: %s (expected 'Key: Value')", h)
		}
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	// Set default content type if not specified
	if _, ok := headers["Content-Type"]; !ok && len(requestData) > 0 {
		headers["Content-Type"] = "application/json"
	}

	// Create invocation request
	invokeReq := client.InvokeRequest{
		Namespace: invokeNamespace,
		Method:    invokeMethod,
		Path:      invokePath,
		Headers:   headers,
		Body:      requestData,
		Async:     invokeAsync,
	}

	printInfo("Invoking function '%s'...", functionName)

	// Invoke function
	response, err := apiClient.InvokeFunction(ctx, functionName, invokeReq)
	if err != nil {
		return fmt.Errorf("failed to invoke function: %w", err)
	}

	// Display response
	if invokeAsync {
		printSuccess("Function invoked asynchronously")
		return nil
	}

	// Show response status
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		printSuccess("Function responded with status %d", response.StatusCode)
	} else {
		printWarning("Function responded with status %d", response.StatusCode)
	}

	// Show response headers if verbose
	if verbose && len(response.Headers) > 0 {
		fmt.Println("\nResponse Headers:")
		for k, v := range response.Headers {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	// Show response body
	if len(response.Body) > 0 {
		fmt.Println("\nResponse Body:")
		
		// Try to pretty print JSON
		var jsonResponse interface{}
		if err := json.Unmarshal(response.Body, &jsonResponse); err == nil {
			prettyJSON, err := json.MarshalIndent(jsonResponse, "", "  ")
			if err == nil {
				fmt.Println(string(prettyJSON))
				return nil
			}
		}
		
		// Otherwise print as string
		fmt.Println(string(response.Body))
	}

	// Show timing if verbose
	if verbose && response.Duration > 0 {
		fmt.Printf("\nResponse Time: %v\n", response.Duration)
	}

	return nil
}