package function

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the function configuration
type Config struct {
	Name        string            `yaml:"name"`
	Runtime     string            `yaml:"runtime"`
	Handler     string            `yaml:"handler"`
	Description string            `yaml:"description,omitempty"`
	Version     string            `yaml:"version"`
	Template    string            `yaml:"template,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Secrets     []string          `yaml:"secrets,omitempty"`
	Build       BuildConfig       `yaml:"build"`
	Deploy      DeployConfig      `yaml:"deploy"`
	Events      []EventConfig     `yaml:"events,omitempty"`
}

// BuildConfig contains build configuration
type BuildConfig struct {
	Builder    string            `yaml:"builder"`
	BuildArgs  map[string]string `yaml:"buildArgs,omitempty"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
}

// DeployConfig contains deployment configuration
type DeployConfig struct {
	Namespace   string              `yaml:"namespace"`
	Registry    string              `yaml:"registry,omitempty"`
	Autoscaling AutoscalingConfig   `yaml:"autoscaling"`
	Resources   ResourceConfig      `yaml:"resources"`
	Timeout     int                 `yaml:"timeout,omitempty"`
	Concurrency int                 `yaml:"concurrency,omitempty"`
	Annotations map[string]string   `yaml:"annotations,omitempty"`
	Labels      map[string]string   `yaml:"labels,omitempty"`
}

// AutoscalingConfig contains autoscaling configuration
type AutoscalingConfig struct {
	MinScale int    `yaml:"minScale"`
	MaxScale int    `yaml:"maxScale"`
	Target   int    `yaml:"target"`
	Metric   string `yaml:"metric"`
}

// ResourceConfig contains resource limits and requests
type ResourceConfig struct {
	Memory string `yaml:"memory"`
	CPU    string `yaml:"cpu"`
}

// EventConfig contains event trigger configuration
type EventConfig struct {
	Type   string            `yaml:"type"`
	Source string            `yaml:"source"`
	Filter map[string]string `yaml:"filter,omitempty"`
}

// LoadConfig loads function configuration from file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if config.Version == "" {
		config.Version = "0.0.1"
	}
	if config.Deploy.Namespace == "" {
		config.Deploy.Namespace = "default"
	}
	if config.Deploy.Timeout == 0 {
		config.Deploy.Timeout = 300
	}
	if config.Deploy.Autoscaling.Metric == "" {
		config.Deploy.Autoscaling.Metric = "concurrency"
	}
	if config.Deploy.Resources.Memory == "" {
		config.Deploy.Resources.Memory = "256Mi"
	}
	if config.Deploy.Resources.CPU == "" {
		config.Deploy.Resources.CPU = "100m"
	}

	return &config, nil
}

// Save saves the configuration to file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("function name is required")
	}

	if !isValidName(c.Name) {
		return fmt.Errorf("invalid function name: must contain only lowercase letters, numbers, and hyphens")
	}

	if c.Runtime == "" {
		return fmt.Errorf("runtime is required")
	}

	validRuntimes := []string{"node", "python", "go", "java", "dotnet", "ruby", "php", "rust", "custom"}
	if !contains(validRuntimes, c.Runtime) {
		return fmt.Errorf("invalid runtime: %s", c.Runtime)
	}

	if c.Handler == "" && c.Runtime != "custom" {
		return fmt.Errorf("handler is required")
	}

	if c.Deploy.Autoscaling.MinScale < 0 {
		return fmt.Errorf("minScale must be >= 0")
	}

	if c.Deploy.Autoscaling.MaxScale < c.Deploy.Autoscaling.MinScale {
		return fmt.Errorf("maxScale must be >= minScale")
	}

	validMetrics := []string{"concurrency", "rps", "cpu", "memory"}
	if !contains(validMetrics, c.Deploy.Autoscaling.Metric) {
		return fmt.Errorf("invalid autoscaling metric: %s", c.Deploy.Autoscaling.Metric)
	}

	return nil
}

// GetImageName returns the full image name including registry
func (c *Config) GetImageName() string {
	if c.Deploy.Registry != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(c.Deploy.Registry, "/"), c.Name)
	}
	return c.Name
}

// GetFullName returns the full function name including namespace
func (c *Config) GetFullName() string {
	return fmt.Sprintf("%s/%s", c.Deploy.Namespace, c.Name)
}

func isValidName(name string) bool {
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}