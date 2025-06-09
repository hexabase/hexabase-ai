package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Builder interface for building functions
type Builder interface {
	Build(ctx context.Context, opts BuildOptions) (string, error)
}

// BuildOptions contains options for building
type BuildOptions struct {
	Path       string
	Name       string
	Tag        string
	Runtime    string
	Handler    string
	BuildArgs  map[string]string
	Dockerfile string
	Platform   string
	Registry   string
	Push       bool
}

// New creates a new builder based on the type
func New(builderType string) (Builder, error) {
	switch builderType {
	case "pack", "":
		return &packBuilder{}, nil
	case "docker":
		return &dockerBuilder{}, nil
	case "ko":
		return &koBuilder{}, nil
	default:
		return nil, fmt.Errorf("unknown builder type: %s", builderType)
	}
}

// packBuilder implements Builder using Cloud Native Buildpacks
type packBuilder struct{}

// dockerBuilder implements Builder using Docker
type dockerBuilder struct{}

// koBuilder implements Builder using ko
type koBuilder struct{}

// Build builds using pack
func (b *packBuilder) Build(ctx context.Context, opts BuildOptions) (string, error) {
	image := getImageName(opts)
	
	args := []string{
		"build", image,
		"--path", opts.Path,
		"--builder", "paketobuildpacks/builder:base",
	}
	
	// Add build args
	for k, v := range opts.BuildArgs {
		args = append(args, "--env", fmt.Sprintf("%s=%s", k, v))
	}
	
	if opts.Push {
		args = append(args, "--publish")
	}
	
	cmd := exec.CommandContext(ctx, "pack", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pack build failed: %w", err)
	}
	
	return image, nil
}

// Build builds using docker
func (b *dockerBuilder) Build(ctx context.Context, opts BuildOptions) (string, error) {
	image := getImageName(opts)
	
	dockerfile := opts.Dockerfile
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}
	
	args := []string{
		"build",
		"-t", image,
		"-f", dockerfile,
	}
	
	// Add build args
	for k, v := range opts.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	
	if opts.Platform != "" {
		args = append(args, "--platform", opts.Platform)
	}
	
	args = append(args, opts.Path)
	
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = opts.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker build failed: %w", err)
	}
	
	if opts.Push {
		pushCmd := exec.CommandContext(ctx, "docker", "push", image)
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		
		if err := pushCmd.Run(); err != nil {
			return "", fmt.Errorf("docker push failed: %w", err)
		}
	}
	
	return image, nil
}

// Build builds using ko
func (b *koBuilder) Build(ctx context.Context, opts BuildOptions) (string, error) {
	if opts.Runtime != "go" {
		return "", fmt.Errorf("ko builder only supports Go runtime")
	}
	
	image := getImageName(opts)
	
	// Set KO_DOCKER_REPO environment variable
	os.Setenv("KO_DOCKER_REPO", opts.Registry)
	
	args := []string{
		"build",
		"--bare",
		"--tags", opts.Tag,
	}
	
	if opts.Push {
		args = append(args, "--push")
	} else {
		args = append(args, "--local")
	}
	
	if opts.Platform != "" {
		args = append(args, "--platform", opts.Platform)
	}
	
	args = append(args, opts.Path)
	
	cmd := exec.CommandContext(ctx, "ko", args...)
	cmd.Dir = opts.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ko build failed: %w", err)
	}
	
	return image, nil
}

// getImageName returns the full image name
func getImageName(opts BuildOptions) string {
	tag := opts.Tag
	if tag == "" {
		tag = "latest"
	}
	
	if opts.Registry != "" {
		return fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(opts.Registry, "/"), opts.Name, tag)
	}
	
	return fmt.Sprintf("%s:%s", opts.Name, tag)
}