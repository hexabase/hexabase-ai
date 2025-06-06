# Tekton Pipeline Configuration

This directory contains Tekton pipeline definitions for the Hexabase KaaS project.

## Pipeline Overview

### pipeline.yaml
Main CI/CD pipeline that orchestrates the following tasks:

1. **git-clone**: Clones the source repository
2. **golang-test**: Runs Go unit and integration tests
3. **trivy-scan**: Performs security vulnerability scanning
4. **kaniko-build**: Builds container images without Docker daemon
5. **update-manifest**: Updates Kubernetes manifests with new image tags

## Parameters

The pipeline accepts the following parameters:
- `git-url`: Source repository URL
- `git-revision`: Git branch or tag to build (default: main)
- `image-registry`: Container registry URL (default: harbor.hexabase.ai)

## Workspaces

Required workspaces:
- `shared-workspace`: Shared storage for pipeline tasks
- `docker-credentials`: Docker registry credentials

## Prerequisites

Before using this pipeline, ensure you have:
1. Tekton Pipelines installed in your cluster
2. Required Tekton Catalog tasks installed:
   - git-clone
   - golang-test
   - trivy-scanner
   - kaniko
   - git-cli
3. PVC for shared workspace
4. Secret with Docker registry credentials

## Usage

Create a PipelineRun to execute the pipeline:

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: hexabase-pipeline-run
spec:
  pipelineRef:
    name: hexabase-ci-pipeline
  params:
    - name: git-url
      value: https://github.com/hexabase/hexabase-kaas
    - name: git-revision
      value: main
  workspaces:
    - name: shared-workspace
      persistentVolumeClaim:
        claimName: pipeline-pvc
    - name: docker-credentials
      secret:
        secretName: docker-registry-secret
```

## Customization

To customize this pipeline:
1. Add additional test tasks for different languages or frameworks
2. Include code quality checks (SonarQube, etc.)
3. Add deployment tasks for different environments
4. Configure notifications for pipeline status