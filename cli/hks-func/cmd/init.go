package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/function"
	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/templates"
	"github.com/spf13/cobra"
)

var (
	functionRuntime string
	template        string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new function project",
	Long: `Initialize a new function project with the specified functionRuntime and template.

This command creates a new function project directory with:
  - function.yaml configuration file
  - Source code template for the selected functionRuntime
  - README with deployment instructions
  - .gitignore file
  - Test files (if applicable)

Examples:
  # Initialize a Python function
  hks-func init my-function --functionRuntime python

  # Initialize a Node.js function with HTTP template
  hks-func init my-api --functionRuntime node --template http

  # Interactive initialization
  hks-func init my-function`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&functionRuntime, "functionRuntime", "r", "", "function functionRuntime (node|python|go|java|dotnet|ruby|php|rust)")
	initCmd.Flags().StringVarP(&template, "template", "t", "http", "function template (http|event|scheduled|stream)")
}

func runInit(cmd *cobra.Command, args []string) error {
	var functionName string

	// Get function name from args or prompt
	if len(args) > 0 {
		functionName = args[0]
	} else {
		prompt := &survey.Input{
			Message: "Function name:",
			Default: "my-function",
		}
		if err := survey.AskOne(prompt, &functionName); err != nil {
			return err
		}
	}

	// Validate function name
	if !isValidFunctionName(functionName) {
		return fmt.Errorf("invalid function name: must contain only lowercase letters, numbers, and hyphens")
	}

	// Get functionRuntime if not specified
	if functionRuntime == "" {
		prompt := &survey.Select{
			Message: "Select functionRuntime:",
			Options: []string{
				"node",
				"python",
				"go",
				"java",
				"dotnet",
				"ruby",
				"php",
				"rust",
			},
			Default: "node",
		}
		if err := survey.AskOne(prompt, &functionRuntime); err != nil {
			return err
		}
	}

	// Get template if not specified
	if template == "" {
		prompt := &survey.Select{
			Message: "Select template:",
			Options: []string{
				"http",
				"event",
				"scheduled",
				"stream",
			},
			Default: "http",
		}
		if err := survey.AskOne(prompt, &template); err != nil {
			return err
		}
	}

	// Create function directory
	functionDir := filepath.Join(".", functionName)
	if err := os.MkdirAll(functionDir, 0755); err != nil {
		return fmt.Errorf("failed to create function directory: %w", err)
	}

	printInfo("Initializing function '%s' with functionRuntime '%s' and template '%s'", functionName, functionRuntime, template)

	// Create function configuration
	config := &function.Config{
		Name:        functionName,
		Runtime:     functionRuntime,
		Handler:     getDefaultHandler(functionRuntime),
		Description: fmt.Sprintf("%s function created with hks-func", functionName),
		Version:     "0.0.1",
		Template:    template,
		Environment: map[string]string{},
		Build: function.BuildConfig{
			Builder: getDefaultBuilder(functionRuntime),
		},
		Deploy: function.DeployConfig{
			Namespace:   "default",
			Autoscaling: getDefaultAutoscaling(template),
			Resources:   getDefaultResources(functionRuntime),
		},
	}

	// Save function.yaml
	configPath := filepath.Join(functionDir, "function.yaml")
	if err := config.Save(configPath); err != nil {
		return fmt.Errorf("failed to save function.yaml: %w", err)
	}
	printSuccess("Created function.yaml")

	// Generate template files
	tmpl, err := templates.GetTemplate(functionRuntime, template)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Write source files
	for filename, content := range tmpl.Files {
		filePath := filepath.Join(functionDir, filename)
		
		// Create directory if needed
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Process template
		processedContent := strings.ReplaceAll(content, "{{.FunctionName}}", functionName)
		processedContent = strings.ReplaceAll(processedContent, "{{.Description}}", config.Description)

		if err := os.WriteFile(filePath, []byte(processedContent), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filename, err)
		}
		printSuccess("Created %s", filename)
	}

	// Create .gitignore
	gitignorePath := filepath.Join(functionDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(tmpl.GitIgnore), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	printSuccess("Created .gitignore")

	// Create README.md
	readmePath := filepath.Join(functionDir, "README.md")
	readme := generateReadme(functionName, functionRuntime, template)
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	printSuccess("Created README.md")

	// Print next steps
	fmt.Println()
	printSuccess("Function initialized successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. cd %s\n", functionName)
	fmt.Println("  2. Edit the function code")
	fmt.Println("  3. hks-func test        # Test locally")
	fmt.Println("  4. hks-func deploy      # Deploy to Hexabase AI")

	return nil
}

func isValidFunctionName(name string) bool {
	if name == "" {
		return false
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}
	return true
}

func getDefaultHandler(functionRuntime string) string {
	switch functionRuntime {
	case "node":
		return "index.handler"
	case "python":
		return "main.handler"
	case "go":
		return "main"
	case "java":
		return "com.example.Handler"
	case "dotnet":
		return "Function::Handler"
	case "ruby":
		return "handler.rb"
	case "php":
		return "index.php"
	case "rust":
		return "handler"
	default:
		return "handler"
	}
}

func getDefaultBuilder(functionRuntime string) string {
	switch functionRuntime {
	case "node":
		return "pack"
	case "python":
		return "pack"
	case "go":
		return "ko"
	case "java":
		return "pack"
	case "dotnet":
		return "pack"
	case "ruby":
		return "pack"
	case "php":
		return "pack"
	case "rust":
		return "docker"
	default:
		return "pack"
	}
}

func getDefaultAutoscaling(template string) function.AutoscalingConfig {
	switch template {
	case "http":
		return function.AutoscalingConfig{
			MinScale: 0,
			MaxScale: 100,
			Target:   100,
			Metric:   "concurrency",
		}
	case "event":
		return function.AutoscalingConfig{
			MinScale: 0,
			MaxScale: 50,
			Target:   10,
			Metric:   "rps",
		}
	case "scheduled":
		return function.AutoscalingConfig{
			MinScale: 0,
			MaxScale: 1,
			Target:   1,
			Metric:   "concurrency",
		}
	case "stream":
		return function.AutoscalingConfig{
			MinScale: 1,
			MaxScale: 10,
			Target:   100,
			Metric:   "concurrency",
		}
	default:
		return function.AutoscalingConfig{
			MinScale: 0,
			MaxScale: 100,
			Target:   100,
			Metric:   "concurrency",
		}
	}
}

func getDefaultResources(functionRuntime string) function.ResourceConfig {
	switch functionRuntime {
	case "go", "rust":
		return function.ResourceConfig{
			Memory: "128Mi",
			CPU:    "100m",
		}
	case "java", "dotnet":
		return function.ResourceConfig{
			Memory: "512Mi",
			CPU:    "500m",
		}
	default:
		return function.ResourceConfig{
			Memory: "256Mi",
			CPU:    "100m",
		}
	}
}

func generateReadme(name, functionRuntime, template string) string {
	return fmt.Sprintf(`# %s

A %s function created with hks-func CLI.

## Development

### Prerequisites
- hks-func CLI installed
- %s functionRuntime installed
- Docker (optional, for local testing)

### Local Development

1. Install dependencies:
   %s

2. Run tests:
   ` + "`hks-func test`" + `

3. Test locally:
   ` + "`hks-func test --local`" + `

## Deployment

Deploy to Hexabase AI:
` + "```bash" + `
hks-func deploy
` + "```" + `

Deploy with custom settings:
` + "```bash" + `
hks-func deploy --namespace my-namespace --tag v1.0.0
` + "```" + `

## Configuration

Edit ` + "`function.yaml`" + ` to configure:
- Runtime settings
- Environment variables
- Resource limits
- Autoscaling parameters

## Function URLs

After deployment, your function will be available at:
- Development: https://%s-{namespace}.fn.dev.hexabase.ai
- Production: https://%s-{namespace}.fn.hexabase.ai

## Monitoring

View logs:
` + "```bash" + `
hks-func logs
` + "```" + `

View metrics:
` + "```bash" + `
hks-func describe
` + "```" + `
`, name, functionRuntime, functionRuntime, getDependencyInstallCommand(functionRuntime), name, name)
}

func getDependencyInstallCommand(functionRuntime string) string {
	switch functionRuntime {
	case "node":
		return "`npm install`"
	case "python":
		return "`pip install -r requirements.txt`"
	case "go":
		return "`go mod download`"
	case "java":
		return "`mvn install` or `gradle build`"
	case "dotnet":
		return "`dotnet restore`"
	case "ruby":
		return "`bundle install`"
	case "php":
		return "`composer install`"
	case "rust":
		return "`cargo build`"
	default:
		return "See functionRuntime documentation"
	}
}