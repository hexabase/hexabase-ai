package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `Manage hks-func CLI configuration including authentication and defaults.`,
}

// authCmd represents the config auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure authentication",
	Long: `Configure authentication for the Hexabase AI platform.

This command stores your API credentials securely for use with other commands.`,
	RunE: runAuth,
}

// setCmd represents the config set command
var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  - default.namespace: Default namespace for operations
  - default.registry: Default container registry
  - api.endpoint: API endpoint URL
  - api.timeout: API request timeout`,
	Args: cobra.ExactArgs(2),
	RunE: runSet,
}

// getCmd represents the config get command
var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

// viewCmd represents the config view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View all configuration",
	RunE:  runView,
}

func init() {
	configCmd.AddCommand(authCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(viewCmd)
}

func runAuth(cmd *cobra.Command, args []string) error {
	// Check if already authenticated
	existingToken := viper.GetString("auth.token")
	if existingToken != "" {
		confirm := false
		prompt := &survey.Confirm{
			Message: "You are already authenticated. Do you want to reconfigure?",
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirm); err != nil {
			return err
		}
		if !confirm {
			return nil
		}
	}

	// Prompt for authentication method
	authMethod := ""
	prompt := &survey.Select{
		Message: "Select authentication method:",
		Options: []string{"API Token", "OAuth"},
		Default: "API Token",
	}
	if err := survey.AskOne(prompt, &authMethod); err != nil {
		return err
	}

	if authMethod == "API Token" {
		// Prompt for API token
		token := ""
		tokenPrompt := &survey.Password{
			Message: "Enter your API token:",
		}
		if err := survey.AskOne(tokenPrompt, &token); err != nil {
			return err
		}

		// Prompt for API endpoint
		endpoint := viper.GetString("api.endpoint")
		if endpoint == "" {
			endpoint = "https://api.hexabase.ai"
		}
		endpointPrompt := &survey.Input{
			Message: "API endpoint:",
			Default: endpoint,
		}
		if err := survey.AskOne(endpointPrompt, &endpoint); err != nil {
			return err
		}

		// Save configuration
		viper.Set("auth.method", "token")
		viper.Set("auth.token", token)
		viper.Set("api.endpoint", endpoint)

		// Save to file
		configDir := filepath.Join(os.Getenv("HOME"), ".hks-func")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		configFile := filepath.Join(configDir, "config.yaml")
		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		printSuccess("Authentication configured successfully!")
		printInfo("Configuration saved to: %s", configFile)
	} else {
		// OAuth flow would be implemented here
		return fmt.Errorf("OAuth authentication not yet implemented")
	}

	return nil
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Validate key
	validKeys := []string{
		"default.namespace",
		"default.registry",
		"api.endpoint",
		"api.timeout",
		"build.platform",
	}

	isValid := false
	for _, k := range validKeys {
		if k == key {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid configuration key: %s", key)
	}

	// Set value
	viper.Set(key, value)

	// Save configuration
	if err := viper.WriteConfig(); err != nil {
		// Try to write to default location
		configDir := filepath.Join(os.Getenv("HOME"), ".hks-func")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		configFile := filepath.Join(configDir, "config.yaml")
		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
	}

	printSuccess("Configuration updated: %s = %s", key, value)
	return nil
}

func runGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := viper.Get(key)

	if value == nil {
		printWarning("Configuration key not found: %s", key)
		return nil
	}

	fmt.Println(value)
	return nil
}

func runView(cmd *cobra.Command, args []string) error {
	settings := viper.AllSettings()

	if len(settings) == 0 {
		printInfo("No configuration found")
		return nil
	}

	fmt.Println("Current configuration:")
	fmt.Println()
	
	// Display configuration in a formatted way
	displayConfig(settings, "")
	
	if configFile := viper.ConfigFileUsed(); configFile != "" {
		fmt.Println()
		printInfo("Configuration file: %s", configFile)
	}

	return nil
}

func displayConfig(config map[string]interface{}, prefix string) {
	for key, value := range config {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			displayConfig(v, fullKey)
		case string:
			// Mask sensitive values
			if key == "token" || key == "password" || key == "secret" {
				fmt.Printf("  %s: %s\n", fullKey, "***")
			} else {
				fmt.Printf("  %s: %s\n", fullKey, v)
			}
		default:
			fmt.Printf("  %s: %v\n", fullKey, v)
		}
	}
}