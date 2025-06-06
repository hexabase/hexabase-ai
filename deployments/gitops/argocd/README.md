# ArgoCD GitOps Configuration

This directory contains ArgoCD application configurations for GitOps-based deployments of Hexabase KaaS.

## Overview

ArgoCD is a declarative, GitOps continuous delivery tool for Kubernetes. This configuration enables automatic synchronization of Hexabase KaaS deployments with Git repositories.

## Configuration Files

### application.yaml
Defines the ArgoCD Application resource that manages Hexabase KaaS deployment with:
- Automatic synchronization from Git
- Self-healing capabilities
- Retry logic for transient failures
- Namespace creation

## Setup Instructions

1. **Install ArgoCD**:
   ```bash
   kubectl create namespace argocd
   kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   ```

2. **Access ArgoCD UI**:
   ```bash
   kubectl port-forward svc/argocd-server -n argocd 8080:443
   # Default admin password
   kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
   ```

3. **Apply Application Configuration**:
   ```bash
   kubectl apply -f application.yaml
   ```

## Application Settings

### Source Configuration
- `repoURL`: Git repository containing Kubernetes manifests
- `targetRevision`: Git reference to track (branch, tag, or commit)
- `path`: Path to manifests within the repository

### Destination Configuration
- `server`: Kubernetes API server URL
- `namespace`: Target namespace for deployment

### Sync Policy
- `automated`: Enable automatic synchronization
  - `prune`: Remove resources not in Git
  - `selfHeal`: Revert manual changes
  - `allowEmpty`: Prevent syncing empty manifests
- `syncOptions`: Additional sync behaviors
- `retry`: Automatic retry configuration

## Multiple Environments

Create separate applications for different environments:

```yaml
# staging-application.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: hexabase-kaas-staging
  namespace: argocd
spec:
  source:
    path: manifests/overlays/staging
  destination:
    namespace: hexabase-staging
  # ... other settings
```

## Monitoring

Monitor ArgoCD applications:
```bash
# CLI status
argocd app get hexabase-kaas
argocd app sync hexabase-kaas

# Watch application
argocd app wait hexabase-kaas

# View history
argocd app history hexabase-kaas
```

## Advanced Features

### Progressive Rollouts
Integrate with Flagger or Argo Rollouts:
```yaml
spec:
  source:
    plugin:
      name: argo-rollouts
      env:
        - name: ROLLOUT_STRATEGY
          value: canary
```

### Multi-Source Applications
Deploy from multiple repositories:
```yaml
spec:
  sources:
    - repoURL: https://github.com/hexabase/k8s-manifests
      path: base
    - repoURL: https://github.com/hexabase/k8s-config
      path: overlays/production
```

### Notifications
Configure Slack/email notifications:
```yaml
metadata:
  annotations:
    notifications.argoproj.io/subscribe.on-sync-succeeded.slack: deployment-notifications
```

## Troubleshooting

Common issues:
1. **Out of Sync**: Check for manual changes or drift
2. **Sync Failures**: Review application logs and events
3. **Permission Errors**: Verify RBAC settings
4. **Image Pull Errors**: Check registry credentials