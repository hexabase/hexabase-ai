# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hexabase AI is a multi-tenant Kubernetes as a Service platform built on K3s and vCluster. The project consists of multiple repositories:

- **Current Repository**: Documentation and architecture specifications
- **UI Repository**: https://github.com/b-eee/hxb-next-webui (Next.js)
- **API Repository**: https://github.com/b-eee/apicore (Go)

## Architecture Context

### Core Components
- **Control Plane**: Go-based API server handling tenant management, OIDC auth, and vCluster orchestration
- **UI**: Next.js frontend with state management (Zustand/Recoil/Redux Toolkit) and data fetching (SWR/React Query)
- **Infrastructure**: Host K3s cluster with vCluster for tenant isolation
- **Data Layer**: PostgreSQL (primary), Redis (cache), NATS (message queue)
- **Monitoring**: Prometheus, Grafana, Loki stack
- **Security**: Kyverno policies, Trivy scanning, Falco runtime monitoring

### Key Concepts Mapping
| Hexabase Concept | Kubernetes Equivalent | Description |
|-----------------|---------------------|-------------|
| Organization | (none) | Billing and organization management unit |
| Workspace | vCluster | Tenant isolation boundary |
| Project | Namespace | Resource isolation within Workspace |
| Workspace Member | OIDC Subject | Technical user with vCluster access |
| Workspace Group | OIDC Claim | Permission assignment unit |

## Development Commands

### For Go API Development
```bash
# When working on the API repository
go mod download          # Download dependencies
go test ./...           # Run all tests
go run cmd/api/main.go  # Run API server locally
go build -o hexabase-api cmd/api/main.go  # Build binary
```

### For Next.js UI Development
```bash
# When working on the UI repository
npm install             # Install dependencies
npm run dev            # Run development server
npm run build          # Build for production
npm test               # Run tests
npm run lint           # Run linting
```

### For Kubernetes/Helm Development
```bash
# Deploy Hexabase AI using Helm
helm install hexabase-ai ./charts/hexabase-ai -f values.yaml

# Common kubectl commands for debugging
kubectl get vcluster -A                    # List all vClusters
kubectl get pods -n hexabase-control-plane # List control plane pods
kubectl logs -n hexabase-control-plane <pod-name> # View pod logs
```

## Implementation Guidelines

### API Development (Go)
- Use structured logging (Logrus/Zap) for all log output
- Implement comprehensive error handling with context
- Follow repository pattern for database access with GORM
- Use client-go for Kubernetes API interactions
- Implement metrics collection with Prometheus client library
- Configuration management with Viper

### UI Development (Next.js)
- Component-based architecture with clear separation of concerns
- Type-safe API client generation from OpenAPI specs
- Real-time updates via WebSocket/SSE for vCluster provisioning status
- Implement proper loading states and error boundaries
- Consider accessibility (a11y) and internationalization (i18n)

### vCluster Management
- Use vcluster CLI or Kubernetes API for lifecycle management
- Apply OIDC configuration for each vCluster
- Set ResourceQuotas based on Workspace Plan
- Configure HNC (Hierarchical Namespace Controller) for project hierarchy
- Implement proper cleanup on deletion

### Async Processing
- Use NATS for task queuing (vCluster provisioning, Stripe billing)
- Implement retry logic with exponential backoff
- Store task status in PostgreSQL
- Notify completion/failure via NATS

## Security Considerations
- Never expose or log secrets/API keys
- All external IdP auth via OIDC
- Enforce Kyverno policies for compliance
- Regular vulnerability scanning with Trivy
- Runtime threat detection with Falco