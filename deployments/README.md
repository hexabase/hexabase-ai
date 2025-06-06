# Deployments

This directory contains deployment configurations and operational resources for Hexabase KaaS.

## Directory Structure

```
deployments/
├── helm/               # Helm charts for Kubernetes deployment
├── gitops/            # GitOps configurations
│   ├── flux/          # Flux CD configurations
│   └── argocd/        # ArgoCD application definitions
├── policies/          # Security and governance policies
│   └── kyverno/       # Kyverno policy definitions
├── monitoring/        # Monitoring and observability
│   └── prometheus/    # Prometheus rules and ServiceMonitors
├── canary/           # Progressive delivery configurations
│   └── flagger/      # Flagger canary deployments
└── k8s/              # Raw Kubernetes manifests
```

## Quick Start

### Deploy with Helm
```bash
helm install hexabase-kaas ./helm/hexabase-kaas \
  --namespace hexabase-system \
  --create-namespace
```

### Deploy with GitOps

#### Using Flux
```bash
kubectl apply -f gitops/flux/gotk-sync.yaml
```

#### Using ArgoCD
```bash
kubectl apply -f gitops/argocd/application.yaml
```

## Components

### Helm Charts
The main deployment method for Hexabase KaaS. Includes:
- API server deployment
- UI deployment
- Database migrations
- RBAC configuration
- Service mesh integration

### GitOps
Automated deployment and synchronization:
- **Flux**: Lightweight GitOps with automatic image updates
- **ArgoCD**: Feature-rich GitOps with web UI

### Policies
Security and compliance enforcement:
- **Kyverno**: Policy-as-code for Kubernetes
  - Image signature verification
  - Resource quotas
  - Security standards

### Monitoring
Observability stack configuration:
- **Prometheus**: Metrics collection and alerting
- **ServiceMonitors**: Automatic service discovery
- **Alert rules**: Pre-configured alerts

### Progressive Delivery
Safe rollout strategies:
- **Flagger**: Automated canary deployments
- **Metrics-based promotion**: Automatic rollback on failures

## Environment-Specific Values

### Development
```bash
helm install hexabase-kaas ./helm/hexabase-kaas \
  -f ./helm/values-dev.yaml
```

### Staging
```bash
helm install hexabase-kaas ./helm/hexabase-kaas \
  -f ./helm/values-staging.yaml
```

### Production
```bash
helm install hexabase-kaas ./helm/hexabase-kaas \
  -f ./helm/values-production.yaml
```

## Security Considerations

1. **Image Signing**: All production images must be signed
2. **Policy Enforcement**: Kyverno policies are enforced in production
3. **Network Policies**: Strict network isolation between tenants
4. **RBAC**: Least-privilege access controls

## Monitoring

Deploy the monitoring stack:
```bash
kubectl apply -f monitoring/prometheus/
```

Access Grafana dashboards:
```bash
kubectl port-forward -n monitoring svc/grafana 3000:3000
```

## Troubleshooting

### Check Deployment Status
```bash
kubectl get deployments -n hexabase-system
kubectl get pods -n hexabase-system
```

### View Logs
```bash
kubectl logs -n hexabase-system deployment/hexabase-api
kubectl logs -n hexabase-system deployment/hexabase-ui
```

### Debug GitOps Sync
```bash
# Flux
flux get kustomizations
flux logs --follow

# ArgoCD
argocd app get hexabase-kaas
argocd app sync hexabase-kaas
```

## Related Documentation

- [Kubernetes Deployment Guide](../docs/operations/kubernetes-deployment.md)
- [CI/CD Architecture](../docs/architecture/cicd-architecture.md)
- [Security Architecture](../docs/architecture/security-architecture.md)