# Hexabase AI: Concept and Architecture

## 1. Project Overview

### Vision

Hexabase AI is an open-source, multi-tenant Kubernetes as a Service platform built on K3s and vCluster, designed to make Kubernetes accessible to developers of all skill levels.

### Core Values

- **Ease of Adoption**: Lightweight K3s base with vCluster virtualization for rapid deployment
- **Intuitive UX**: Abstract Kubernetes complexity through Organizations, Workspaces, and Projects
- **Strong Tenant Isolation**: vCluster provides dedicated API servers and control planes per tenant
- **Cloud-Native Operations**: Built-in Prometheus, Grafana, Loki monitoring; Flux GitOps; Kyverno policies
- **Open Source Transparency**: Community-driven development with full source code availability

### Existing Codebases

- **UI (Next.js)**: https://github.com/b-eee/hxb-next-webui
- **API (Go)**: https://github.com/b-eee/apicore

Both repositories require significant reimplementation based on this specification.

## 2. System Architecture

### Component Overview

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Hexabase UI    │────▶│  Hexabase API    │────▶│  Host K3s       │
│  (Next.js)      │     │  (Control Plane) │     │  Cluster        │
└─────────────────┘     └────────┬─────────┘     └────────┬────────┘
                                 │                        │
                        ┌────────┴────────┐               │
                        │                 │          ┌────▼────────┐
                  ┌─────▼─────┐     ┌─────▼─────┐    │  vClusters  │
                  │PostgreSQL │     │   Redis   │    │  (Tenants)  │
                  └───────────┘     └───────────┘    └─────────────┘
                                          │
                                    ┌─────▼─────┐
                                    │   NATS    │
                                    └───────────┘
```

### Data Flow

1. **User Operations**: Browser → UI → API requests with auth tokens
2. **API Processing**: Auth validation → Business logic → DB updates → Async tasks
3. **vCluster Orchestration**: API → client-go → Host K3s → vCluster lifecycle
4. **Async Processing**: API → NATS → Workers → Long-running operations
5. **State Persistence**: PostgreSQL for all entities, Redis for caching
6. **Monitoring**: Prometheus metrics → Loki logs → Grafana dashboards
7. **GitOps Deployment**: Git → Flux → Host K3s → Automated updates
8. **Policy Enforcement**: Kyverno admission controller → Policy validation

## 3. Core Concepts

| Hexabase Concept      | Kubernetes Equivalent | Scope     | Description                               |
| --------------------- | --------------------- | --------- | ----------------------------------------- |
| Organization          | N/A                   | Hexabase  | Billing and user management unit          |
| Workspace             | vCluster              | Host K3s  | Isolated Kubernetes environment           |
| Workspace Plan        | ResourceQuota/Nodes   | vCluster  | Resource limits and node allocation       |
| Organization User     | N/A                   | Hexabase  | Billing/admin personnel                   |
| Workspace Member      | OIDC Subject          | vCluster  | Technical users with kubectl access       |
| Workspace Group       | OIDC Claim            | vCluster  | Permission assignment unit (hierarchical) |
| Workspace ClusterRole | ClusterRole           | vCluster  | Preset workspace-wide permissions         |
| Project               | Namespace             | vCluster  | Resource isolation within workspace       |
| Project Role          | Role                  | Namespace | Custom permissions within project         |

## 4. User Flows

### 4.1 Signup and Organization Management

- **Authentication**: External IdP (Google/GitHub) via OIDC
- **Auto-provisioning**: Private Organization created on first signup
- **Organization Admin**: Manages billing (Stripe) and invites users

### 4.2 Workspace (vCluster) Management

- **Creation**: Select plan → Provision vCluster → Configure OIDC
- **Initial Setup**:
  - Auto-create ClusterRoles: `hexabase:workspace-admin`, `hexabase:workspace-viewer`
  - Create default groups: `WorkspaceMembers` → `WSAdmins`, `WSUsers`
  - Assign creator to `WSAdmins` group

### 4.3 Project (Namespace) Management

- **Creation**: UI request → Create namespace in vCluster
- **ResourceQuota**: Auto-apply based on workspace plan
- **Custom Roles**: Create project-scoped roles via UI

### 4.4 Permission Management

- **Assignment**: Groups → Roles/ClusterRoles via UI
- **Inheritance**: Recursive group membership resolution
- **OIDC Integration**: Flattened groups in token claims

## 5. Technology Stack

### Core Components

- **Frontend**: Next.js
- **Backend**: Go (Golang)
- **Database**: PostgreSQL (primary), Redis (cache)
- **Messaging**: NATS
- **Container Platform**: K3s + vCluster

### CI/CD & Operations

- **CI Pipeline**: Tekton (Kubernetes-native)
- **GitOps**: ArgoCD or Flux
- **Container Scanning**: Trivy
- **Runtime Security**: Falco
- **Policy Engine**: Kyverno

## 6. Installation (IaC)

### Helm Umbrella Chart

```yaml
apiVersion: v2
name: hexabase-ai
dependencies:
  - name: postgresql
    repository: https://charts.bitnami.com/bitnami
  - name: redis
    repository: https://charts.bitnami.com/bitnami
  - name: nats
    repository: https://nats-io.github.io/k8s/helm/charts/
```

### Quick Install

```bash
helm repo add hexabase https://hexabase.ai/charts
helm install hexabase-ai hexabase/hexabase-ai -f values.yaml
```

## 7. Key Features

### Multi-tenancy

- vCluster provides complete API server isolation
- Dedicated control plane components per tenant
- Optional dedicated nodes for premium plans

### Security

- External IdP authentication only
- Hexabase acts as OIDC provider for vClusters
- Kyverno policy enforcement
- Network isolation between tenants

### Scalability

- Horizontal scaling of control plane components
- Queue-based async processing
- Stateless API design
- Redis caching layer

### Observability

- Prometheus metrics collection
- Centralized logging with Loki
- Pre-built Grafana dashboards
- Real-time resource usage tracking

## 8. Summary

Hexabase AI democratizes Kubernetes access through intelligent abstractions, strong multi-tenancy, and enterprise-grade operations tooling. By leveraging K3s and vCluster, it provides a production-ready platform that scales from individual developers to large organizations, all while maintaining the flexibility and power of native Kubernetes.

The open-source nature ensures transparency, community-driven innovation, and the ability to customize for specific requirements. With simple Helm-based installation and comprehensive monitoring, Hexabase AI represents a new standard for accessible Kubernetes platforms.
