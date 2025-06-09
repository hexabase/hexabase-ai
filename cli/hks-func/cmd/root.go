package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	output  string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "hks-func",
	Short: "Hexabase AI Function CLI",
	Long: color.CyanString(`
╦ ╦╦╔═╔═╗   ╔═╗╦ ╦╔╗╔╔═╗
╠═╣╠╩╗╚═╗───╠╣ ║ ║║║║║  
╩ ╩╩ ╩╚═╝   ╚  ╚═╝╝╚╝╚═╝`) + `

Hexabase AI Function CLI - Deploy and manage serverless functions on Hexabase AI platform.

The hks-func CLI provides commands to:
  - Initialize new function projects
  - Test functions locally
  - Deploy functions to Knative
  - Manage function versions and traffic
  - Monitor function execution and logs`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hks-func.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "output format (text|json|yaml)")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(invokeCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".hks-func" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".hks-func")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("HKS_FUNC")

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// printSuccess prints a success message
func printSuccess(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, color.GreenString("✓ ")+format+"\n", a...)
}

// printError prints an error message
func printError(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, color.RedString("✗ ")+format+"\n", a...)
}

// printInfo prints an info message
func printInfo(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, color.CyanString("ℹ ")+format+"\n", a...)
}

// printWarning prints a warning message
func printWarning(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, color.YellowString("⚠ ")+format+"\n", a...)
}

// getWorkspaceRoot finds the workspace root by looking for function.yaml
func getWorkspaceRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for function.yaml in current directory and parent directories
	dir := cwd
	for {
		configPath := filepath.Join(dir, "function.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("not in a function directory (no function.yaml found)")
}