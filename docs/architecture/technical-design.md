# Hexabase KaaS Control Plane Implementation Specification (Compact Edition)

## 1. System Overview

Hexabase KaaS is a multi-tenant Kubernetes as a Service platform built on K3s and vCluster. This specification defines the design guidelines for the control plane implemented in Go.

### Core Responsibilities
- **API Services**: RESTful API for Next.js UI
- **Authentication & Authorization**: External IdP integration and JWT session management
- **OIDC Provider**: Token issuance for kubectl access to each vCluster
- **vCluster Management**: Complete lifecycle management
- **Billing Processing**: Subscription management via Stripe integration
- **Async Processing**: NATS-based task processing

## 2. Architecture

### Component Structure
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Next.js UI │────▶│  API Server │────▶│ PostgreSQL  │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                    ┌──────┴──────┐
                    │             │
              ┌─────▼─────┐ ┌────▼────┐
              │   NATS    │ │  Redis  │
              └─────┬─────┘ └─────────┘
                    │
              ┌─────▼─────┐
              │  Workers  │
              └───────────┘
```

### External Integrations
- **Host K3s**: Host environment for vClusters
- **vCluster**: Per-tenant Kubernetes environments
- **External IdP**: Google/GitHub OIDC authentication
- **Stripe**: Billing and payment processing

## 3. Database Design

### Primary Tables
| Table | Purpose |
|-------|---------|
| users | User accounts (external IdP linked) |
| organizations | Billing and management units |
| plans | Subscription plan definitions |
| workspaces | vCluster instances |
| projects | Namespaces (HNC hierarchy support) |
| groups | Workspace user groups |
| roles | Custom/preset Roles |
| role_assignments | Group to Role mappings |

### Hierarchical Structure
- **Organization** → **Workspace** → **Project**
- **Group** (hierarchical) → **Role Assignment**

## 4. API Design

### Endpoint Structure
```
/auth
  POST   /login/{provider}     # Initiate external IdP auth
  GET    /callback/{provider}  # Auth callback
  POST   /logout              # Logout
  GET    /me                  # Current user info

/api/v1/organizations
  POST   /                    # Create organization
  GET    /{orgId}            # Organization details
  POST   /{orgId}/users      # Invite users
  
/api/v1/organizations/{orgId}/workspaces
  POST   /                    # Create workspace (async)
  GET    /{wsId}             # Workspace details
  GET    /{wsId}/kubeconfig  # Generate kubeconfig

/api/v1/workspaces/{wsId}
  /groups                    # Group management
  /projects                  # Project (Namespace) management
  /clusterroleassignments   # ClusterRole assignments

/api/v1/projects/{projectId}
  /roles                     # Custom Role management
  /roleassignments          # Role assignments

/webhooks/stripe            # Stripe webhook receiver
```

### Design Principles
- RESTful principles (resource-oriented URLs)
- JSON request/response format
- Versioning (/api/v1/)
- Idempotency guarantee
- Standard HTTP status codes

## 5. OIDC Provider Implementation

### Token Structure
```json
{
  "sub": "hxb-usr-xxxxx",
  "groups": ["WSAdmins", "WorkspaceMembers", "developers"],
  "iss": "https://api.hexabase.ai",
  "aud": "ws-xxxxx",
  "exp": 1234567890
}
```

### Endpoints
- `/.well-known/openid-configuration`: Discovery
- `/.well-known/jwks.json`: Public key distribution

## 6. vCluster Orchestration

### Provisioning Flow
1. Create vCluster (vcluster CLI)
2. Apply OIDC configuration
3. Install HNC
4. Configure ResourceQuota
5. Assign dedicated nodes (if applicable)

### Dedicated Node Management
```yaml
nodeSelector:
  hexabase.ai/node-pool: ws-xxxxx
tolerations:
- key: dedicated
  value: ws-xxxxx
  effect: NoSchedule
```

## 7. Asynchronous Processing

### NATS Topic Structure
```
vcluster.provisioning.*     # vCluster operations
vcluster.hnc.*             # HNC configuration
stripe.webhook.*           # Payment events
user.notification.*        # Notifications
system.maintenance.*       # Periodic tasks
```

### Worker Implementation
- Queue group based load balancing
- Exponential backoff
- Task progress tracking in DB

## 8. Security Measures

- Input validation (go-playground/validator)
- SQL injection prevention (ORM usage)
- Rate limiting (per IP/user)
- Secrets management (Kubernetes Secrets)
- Stripe signature verification
- Principle of least privilege

## 9. Go Package Structure

```
/cmd
  /api         # API server entry point
  /worker      # Worker entry point
/internal
  /api         # HTTP handlers, middleware
  /auth        # Authentication & authorization
  /billing     # Stripe integration
  /config      # Configuration management
  /db          # Models, repositories
  /k8s         # vCluster/HNC management
  /messaging   # NATS pub/sub
  /service     # Business logic
```

## 10. Testing Strategy

### Test Levels
1. **Unit Tests**: 80%+ coverage target
2. **Integration Tests**: API/DB/NATS integration
3. **E2E Tests**: Complete user scenarios

### Testing Tools
- testify: Assertions
- mockgen: Mock generation
- httptest: API testing
- testcontainers: DB integration testing

## 11. Future Extensibility

- Plugin architecture
- gRPC API addition
- Multi-cloud support
- Marketplace platform
- AI/ML workload optimization

## 12. Key Implementation Decisions

1. **Start with Monolith**: Logical module separation with future microservice decomposition in mind
2. **Interface-Driven Design**: For testability and extensibility
3. **Async-First**: Long-running operations handled asynchronously to maintain API responsiveness
4. **Security-First**: Multi-layered defense approach
5. **Cloud-Native**: Designed to run on Kubernetes

---

This specification provides comprehensive guidelines for building a production-ready KaaS platform. For detailed implementation, refer to individual module specifications.