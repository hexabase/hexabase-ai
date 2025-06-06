# Flux GitOps Configuration

This directory contains Flux CD configurations for GitOps-based deployments of Hexabase KaaS.

## Overview

Flux CD enables GitOps workflows by continuously reconciling the cluster state with Git repositories. This configuration sets up automatic deployment of Hexabase KaaS components.

## Configuration Files

### gotk-sync.yaml
Main Flux synchronization configuration that includes:
- **GitRepository**: Defines the source Git repository containing Kubernetes manifests
- **Kustomization**: Configures how Flux applies the manifests to the cluster

## Setup Instructions

1. **Install Flux CLI**:
   ```bash
   curl -s https://fluxcd.io/install.sh | sudo bash
   ```

2. **Bootstrap Flux**:
   ```bash
   flux bootstrap github \
     --owner=hexabase \
     --repository=k8s-manifests \
     --branch=main \
     --path=./clusters/production \
     --personal
   ```

3. **Apply this configuration**:
   ```bash
   kubectl apply -f gotk-sync.yaml
   ```

## Configuration Options

### GitRepository Settings
- `interval`: How often to check for updates (default: 1m)
- `ref.branch`: Git branch to track (default: main)
- `url`: Repository URL containing manifests

### Kustomization Settings
- `interval`: Reconciliation interval (default: 10m)
- `path`: Path to manifests in the repository
- `prune`: Remove resources not in Git (default: true)
- `validation`: Client-side validation before apply

## Variable Substitution

The configuration supports variable substitution from:
- **ConfigMap**: `cluster-config` for environment-specific values
- **Secret**: `cluster-secrets` for sensitive data

Create these resources before deploying:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-config
  namespace: flux-system
data:
  cluster_name: production
  region: us-east-1
---
apiVersion: v1
kind: Secret
metadata:
  name: cluster-secrets
  namespace: flux-system
stringData:
  database_url: postgresql://...
  redis_url: redis://...
```

## Monitoring

Monitor Flux operations:
```bash
# Check Flux components
flux check

# View GitRepository status
flux get sources git

# View Kustomization status
flux get kustomizations

# Watch logs
flux logs --follow
```

## Troubleshooting

Common issues and solutions:
1. **Repository access**: Ensure Flux has deploy keys or PAT configured
2. **Image pull errors**: Check imagePullSecrets are properly configured
3. **Resource conflicts**: Review prune settings and resource ownership
4. **Validation failures**: Check manifest syntax and API versions