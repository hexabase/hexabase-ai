# ADR-005: CI/CD Architecture with Provider Abstraction

**Date**: 2025-06-07  
**Status**: Implemented  
**Authors**: DevOps Team

## 1. Background

Hexabase AI needed a flexible CI/CD system that could:
- Support multiple CI/CD providers (Tekton, GitLab CI, GitHub Actions)
- Provide GitOps-based deployments
- Enable both push and pull-based deployment models
- Support various source code repositories
- Integrate with the existing Kubernetes infrastructure
- Provide secure credential management

The platform needed to accommodate different user preferences and existing CI/CD investments.

## 2. Status

**Implemented** - Provider abstraction layer with Tekton as the default provider is fully operational.

## 3. Other Options Considered

### Option A: Tekton-Only Solution
- Native Kubernetes CI/CD
- Pipeline as Code
- Good Kubernetes integration

### Option B: External CI/CD Integration Only
- Support only external systems
- Webhook-based triggers
- No built-in CI/CD

### Option C: Provider Abstraction with Default
- Support multiple providers
- Tekton as built-in option
- External provider integration
- Unified API surface

## 4. What Was Decided

We chose **Option C: Provider Abstraction with Default** implementing:
- Provider interface for CI/CD operations
- Tekton as the default built-in provider
- GitHub Actions and GitLab CI webhook integration
- Secure credential storage per workspace
- GitOps deployment via Flux CD
- Unified pipeline status tracking

## 5. Why Did You Choose It?

- **Flexibility**: Users can choose their preferred CI/CD system
- **No Lock-in**: Easy to switch between providers
- **Enterprise Ready**: Supports existing enterprise CI/CD
- **Cloud Native**: Tekton provides native Kubernetes pipelines
- **Security**: Credential isolation per workspace

## 6. Why Didn't You Choose the Other Options?

### Why not Tekton-Only?
- Forces users to learn new system
- No integration with existing pipelines
- Limited adoption outside Kubernetes

### Why not External Only?
- No built-in option for new users
- Complex webhook management
- Limited control over execution

## 7. What Has Not Been Decided

- Support for Jenkins integration
- Cross-workspace pipeline sharing
- Advanced pipeline composition
- Cost allocation for CI/CD resources

## 8. Considerations

### Architecture Design
```
┌─────────────────┐
│  CI/CD Service  │
└────────┬────────┘
         │
┌────────▼────────┐
│Provider Interface│
└────────┬────────┘
         │
┌────────┴────────┬──────────────┬─────────────┐
│                 │              │             │
▼                 ▼              ▼             ▼
Tekton         GitHub        GitLab      Future
Provider       Actions       CI          Providers
```

### Provider Interface
```go
type CICDProvider interface {
    CreatePipeline(ctx context.Context, spec PipelineSpec) (*Pipeline, error)
    TriggerPipeline(ctx context.Context, id string, params map[string]string) (*PipelineRun, error)
    GetPipelineRun(ctx context.Context, id string) (*PipelineRun, error)
    ListPipelineRuns(ctx context.Context, pipelineID string) ([]*PipelineRun, error)
    DeletePipeline(ctx context.Context, id string) error
}
```

### Credential Management
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cicd-credentials
  namespace: workspace-xxx
data:
  git-token: <base64>
  registry-password: <base64>
  deploy-key: <base64>
```

### Pipeline Templates
Common templates provided:
- **Build & Push**: Source → Build → Test → Push to Registry
- **GitOps Deploy**: Update manifests → Commit → Flux sync
- **Full Stack**: Build → Test → Deploy → Smoke test

### Webhook Integration
```go
// Webhook handler for external providers
func HandleWebhook(provider string, payload []byte) error {
    switch provider {
    case "github":
        return handleGitHubWebhook(payload)
    case "gitlab":
        return handleGitLabWebhook(payload)
    default:
        return ErrUnknownProvider
    }
}
```

### Security Considerations
- Webhook signature verification
- Network policies for pipeline pods
- RBAC for pipeline operations
- Credential rotation policies
- Image scanning integration

### Monitoring and Observability
- Pipeline execution metrics
- Success/failure rates
- Duration tracking
- Resource utilization
- Cost per pipeline run

### GitOps Integration
```yaml
# Flux HelmRelease for application deployment
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: user-app
  namespace: flux-system
spec:
  interval: 5m
  chart:
    spec:
      chart: ./charts/app
      sourceRef:
        kind: GitRepository
        name: user-repo
```

### Future Enhancements
- Pipeline marketplace
- Visual pipeline designer
- Advanced triggering rules
- Multi-cluster deployments
- Canary deployment strategies