# GitLab CI Configuration

This directory contains GitLab CI/CD pipeline configuration for the Hexabase KaaS project.

## Pipeline Configuration

### .gitlab-ci.yml
Main pipeline configuration file that defines the following stages:

#### Test Stage
- **unit**: Runs Go unit tests with race detection and coverage reporting
- **integration**: Runs integration tests with PostgreSQL and Redis services

#### Security Stage
- **scan**: Performs filesystem vulnerability scanning using Trivy

#### Build Stage
- **images**: Builds multi-platform Docker images using buildx and pushes to GitLab Container Registry

#### Deploy Stage
- **staging**: Automatically deploys to staging on develop branch
- **production**: Manual deployment to production on tags

## Variables

The pipeline uses the following variables:
- `DOCKER_DRIVER`: Docker storage driver (overlay2)
- `DOCKER_TLS_CERTDIR`: TLS certificate directory for Docker
- `REGISTRY`: Container registry URL (default: registry.gitlab.com)
- `IMAGE_TAG`: Image tag based on commit SHA

## Required CI/CD Variables

Configure these in your GitLab project settings:
- `CI_REGISTRY_USER`: Registry username (automatically provided)
- `CI_REGISTRY_PASSWORD`: Registry password (automatically provided)
- `KUBECONFIG_STAGING`: Kubernetes config for staging cluster
- `KUBECONFIG_PRODUCTION`: Kubernetes config for production cluster

## Usage

This pipeline is automatically triggered:
- On every push to any branch (test and security stages)
- On develop branch (staging deployment)
- On tags (production deployment with manual approval)

## Customization

To adapt this pipeline for your environment:
1. Update registry URLs and image naming conventions
2. Modify Helm chart paths and values files
3. Configure additional test suites or security scanners
4. Adjust deployment environments and URLs