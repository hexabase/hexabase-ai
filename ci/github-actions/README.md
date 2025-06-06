# GitHub Actions CI/CD Configuration

This directory contains GitHub Actions workflow configurations for the Hexabase KaaS project.

## Workflows

### ci-cd.yml
Main CI/CD pipeline that runs on every push and pull request. It includes:
- **Test Stage**: Runs Go unit tests with race detection and coverage reporting
- **Security Stage**: Performs vulnerability scanning with Trivy and static analysis with Gosec
- **Build Stage**: Builds multi-platform Docker images and pushes to GitHub Container Registry
- **Deploy Staging**: Automatically deploys to staging environment on develop branch merges
- **Deploy Production**: Deploys to production on release creation (requires manual approval)

### supply-chain.yml
Supply chain security workflow that runs on version tags. It includes:
- **SBOM Generation**: Creates Software Bill of Materials using Syft
- **Container Signing**: Signs container images using Cosign
- **SBOM Attestation**: Attaches SBOM to container images for verification

## Usage

These workflows are automatically triggered based on their configured events:
- `ci-cd.yml`: Push to main/develop, pull requests, and releases
- `supply-chain.yml`: Version tags (v*)

## Environment Variables

Required secrets and variables:
- `GITHUB_TOKEN`: Automatically provided by GitHub Actions
- `REGISTRY`: Container registry URL (default: ghcr.io)
- `IMAGE_NAME`: Image name (default: repository name)

## Customization

To customize these workflows for your environment:
1. Update the registry URLs in environment variables
2. Modify deployment contexts for your Kubernetes clusters
3. Adjust build platforms based on your requirements
4. Configure additional security scanning tools as needed