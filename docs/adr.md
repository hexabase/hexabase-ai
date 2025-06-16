# ADR-008: Domain-Driven Design and Code Organization

**Date**: 2025-06-03  
**Status**: Implemented and Enforced  
**Authors**: Architecture Team

## 1. Background

Hexabase AI needed a clear, maintainable code organization strategy that would:
- Support team autonomy and parallel development
- Enforce separation of concerns
- Enable easy testing and mocking
- Prevent circular dependencies
- Scale with growing codebase complexity
- Support future microservices extraction

The team chose Domain-Driven Design (DDD) principles to organize the codebase.

## 2. Status

**Implemented and Enforced** - DDD structure is fully implemented with automated checks in CI/CD pipeline.

## 3. Other Options Considered

### Option A: Traditional MVC Layered Architecture
- Models, Views, Controllers separation
- Horizontal layer organization
- Service layer for business logic

### Option B: Hexagonal Architecture (Ports & Adapters)
- Core domain with ports
- Adapters for external systems
- Complete inversion of control

### Option C: Domain-Driven Design with Clean Architecture
- Domain models at the center
- Use cases/services layer
- Interface adapters
- Dependency rule enforcement

## 4. What Was Decided

We chose **Option C: Domain-Driven Design with Clean Architecture**:
- Clear domain boundaries (workspace, application, etc.)
- Repository pattern for data access
- Service layer for business logic
- Dependency injection throughout
- Strict dependency rules (domain → repository → service)

## 5. Why Did You Choose It?

- **Clarity**: Each domain is self-contained and understandable
- **Testability**: Easy to mock dependencies and test in isolation
- **Flexibility**: Can extract domains into microservices later
- **Maintainability**: Clear ownership and boundaries
- **Onboarding**: New developers quickly understand structure

## 6. Why Didn't You Choose the Other Options?

### Why not Traditional MVC?
- Leads to fat controllers
- Business logic scattered
- Difficult to test
- Poor domain modeling

### Why not Pure Hexagonal?
- Over-engineered for current needs
- Steep learning curve
- Too much abstraction initially

## 7. What Has Not Been Decided

- Criteria for extracting microservices
- Event sourcing implementation
- CQRS adoption timeline
- Domain event bus design

## 8. Considerations

### Directory Structure
```
internal/
├── domain/           # Business logic interfaces
│   ├── workspace/
│   │   ├── models.go      # Domain models
│   │   ├── repository.go  # Repository interface
│   │   └── service.go     # Service interface
│   └── application/
│       ├── models.go
│       ├── repository.go
│       └── service.go
├── repository/       # Data access implementations
│   ├── workspace/
│   │   ├── postgres.go    # PostgreSQL implementation
│   │   └── cache.go       # Redis caching layer
│   └── application/
│       └── postgres.go
├── service/         # Business logic implementations
│   ├── workspace/
│   │   └── service.go     # Service implementation
│   └── application/
│       └── service.go
└── api/
    └── handlers/    # HTTP handlers
        ├── workspace.go
        └── application.go
```

### Dependency Rules
```
┌─────────────┐
│   Handlers  │ ──depends on──┐
└─────────────┘                │
                               ▼
┌─────────────┐         ┌─────────────┐
│   Service   │ ◀───────│   Domain    │
└─────────────┘         └─────────────┘
       │                       ▲
       │                       │
       └──depends on───────────┘
               │
               ▼
       ┌─────────────┐
       │ Repository  │
       └─────────────┘
```

### Domain Model Example
```go
// domain/workspace/models.go
package workspace

type Workspace struct {
    ID              string
    OrganizationID  string
    Name           string
    Plan           Plan
    Status         Status
    CreatedAt      time.Time
}

type Plan string
const (
    PlanShared    Plan = "shared"
    PlanDedicated Plan = "dedicated"
)
```

### Repository Interface
```go
// domain/workspace/repository.go
package workspace

type Repository interface {
    Create(ctx context.Context, ws *Workspace) error
    GetByID(ctx context.Context, id string) (*Workspace, error)
    Update(ctx context.Context, ws *Workspace) error
    Delete(ctx context.Context, id string) error
}
```

### Service Implementation
```go
// service/workspace/service.go
package workspace

type service struct {
    repo      workspace.Repository
    k8sClient kubernetes.Interface
    logger    *zap.Logger
}

func (s *service) CreateWorkspace(ctx context.Context, req CreateRequest) (*workspace.Workspace, error) {
    // Business logic here
    ws := &workspace.Workspace{
        ID:             generateID(),
        OrganizationID: req.OrganizationID,
        Name:          req.Name,
        Plan:          req.Plan,
    }
    
    // Create vCluster
    if err := s.createVCluster(ctx, ws); err != nil {
        return nil, err
    }
    
    // Save to database
    if err := s.repo.Create(ctx, ws); err != nil {
        return nil, err
    }
    
    return ws, nil
}
```

### Testing Strategy
```go
// service/workspace/service_test.go
func TestCreateWorkspace(t *testing.T) {
    mockRepo := &mocks.MockRepository{}
    mockK8s := &mocks.MockKubernetesClient{}
    
    svc := NewService(mockRepo, mockK8s, zap.NewNop())
    
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    mockK8s.On("Create", mock.Anything).Return(nil)
    
    ws, err := svc.CreateWorkspace(context.Background(), CreateRequest{
        Name: "test-workspace",
        Plan: workspace.PlanShared,
    })
    
    assert.NoError(t, err)
    assert.NotEmpty(t, ws.ID)
}
```

### Enforcement
- Pre-commit hooks check import paths
- CI/CD validates dependency rules
- Architecture tests using go-arch-lint
- Regular architecture reviews

### Migration Path to Microservices
1. Identify high-traffic domains
2. Extract domain with its repository and service
3. Add gRPC/REST API layer
4. Deploy as separate service
5. Update clients to use new endpoint# Hexabase AI Architecture History

**Last Updated**: 2025-06-12

## Overview

This document chronicles the architectural evolution of Hexabase AI from initial concept to current implementation.

## Problem Statement

Hexabase needed to evolve from a traditional application platform to a modern Kubernetes-as-a-Service platform while:
- Maintaining backward compatibility with existing Hexabase applications
- Supporting multi-tenant isolation for enterprise customers
- Enabling serverless and AI-powered operations
- Scaling cost-effectively to thousands of workspaces

## Architectural Evolution

### Phase 1: Foundation (June 1-3, 2025)

**Key Decisions:**
- Adopted vCluster for multi-tenant Kubernetes isolation
- Implemented OAuth2/OIDC with PKCE for security
- Established Domain-Driven Design principles

**Challenges Addressed:**
- Namespace-based isolation was insufficient for compliance
- Needed complete API isolation between tenants
- Required flexible resource allocation (shared vs dedicated)

### Phase 2: Core Services (June 6-7, 2025)

**Key Decisions:**
- Built Python-based AIOps service with sandbox security
- Created provider abstraction for CI/CD systems
- Integrated WebSocket for real-time updates

**Challenges Addressed:**
- Direct LLM integration lacked security boundaries
- Users wanted to keep existing CI/CD systems
- Real-time updates needed for better UX

### Phase 3: Operations & Scaling (June 8-9, 2025)

**Key Decisions:**
- Migrated from Knative to Fission (95% cold start improvement)
- Chose ClickHouse over Elasticsearch (70% cost reduction)
- Implemented hybrid backup strategy

**Challenges Addressed:**
- Knative cold starts (2-5s) were unacceptable
- ELK stack too expensive at scale
- Needed application-aware backups

### Phase 4: Optimization (June 10-11, 2025)

**Key Decisions:**
- Completed Fission provider implementation
- Refined provider abstraction patterns
- Enhanced performance monitoring

**Challenges Addressed:**
- Migration path for existing Knative users
- Need for provider flexibility
- Performance visibility requirements

## Key Architectural Patterns

### 1. Provider Abstraction
Emerged as a core pattern across multiple domains:
- **CI/CD**: Support for Tekton, GitHub Actions, GitLab CI
- **Functions**: Support for Knative, Fission, future providers
- **Storage**: Abstraction over Proxmox, cloud providers

### 2. Security-First Design
Consistent security approach:
- OAuth2/OIDC for all authentication
- Sandbox model for AI operations
- Zero-trust between components
- Comprehensive audit logging

### 3. Cost-Optimized Choices
Decisions driven by scalability costs:
- ClickHouse vs Elasticsearch (70% savings)
- Fission vs Knative (reduced compute time)
- vCluster vs separate clusters (operational efficiency)

### 4. Domain-Driven Organization
Clear boundaries and ownership:
- Separate domains for each business capability
- Repository pattern for data access
- Service layer for business logic
- Clean dependency rules

## Lessons Learned

### What Worked Well

1. **Provider Abstraction**: Allowed seamless migration between technologies
2. **vCluster Choice**: Provided perfect balance of isolation and efficiency
3. **TDD Approach**: Ensured quality and maintainability
4. **ClickHouse for Logs**: Massive cost savings with better performance

### Challenges Overcome

1. **Cold Start Performance**: Knative → Fission migration solved this
2. **Multi-tenancy Complexity**: vCluster simplified isolation
3. **AI Security**: Sandbox model provided necessary boundaries
4. **Cost at Scale**: Alternative technology choices reduced costs

### Technical Debt Addressed

1. **Monolith Extraction**: DDD prepared for future microservices
2. **Vendor Lock-in**: Provider abstractions prevent this
3. **Observability**: Comprehensive logging and monitoring added
4. **Security Gaps**: OAuth2/OIDC implementation closed these

## Future Architecture Direction

### Near Term (Q3-Q4 2025)
- Complete Knative deprecation
- Add WebAssembly function support
- Implement cross-region disaster recovery
- Enhanced AI capabilities with RAG

### Medium Term (2026)
- Extract high-traffic domains to microservices
- Edge function deployment
- GPU support for ML workloads
- Multi-cloud provider support

### Long Term Vision
- Fully distributed, edge-native platform
- AI-driven autonomous operations
- Seamless hybrid cloud deployment
- Industry-specific solution templates

## Architecture Principles

Through this evolution, key principles emerged:

1. **Avoid Lock-in**: Always provide abstraction layers
2. **Security First**: Never compromise on security
3. **Cost Awareness**: Consider TCO in all decisions
4. **Performance Matters**: Sub-second operations are the target
5. **Developer Experience**: Make the right thing easy to do

## Conclusion

Hexabase AI's architecture has evolved from a traditional platform to a modern, cloud-native Kubernetes-as-a-Service solution. The journey required difficult decisions, multiple pivots, and constant optimization, but resulted in a platform that is:

- **Secure**: Enterprise-grade security throughout
- **Scalable**: Proven to handle thousands of workspaces
- **Flexible**: Supports multiple providers and deployment models
- **Cost-effective**: 70% reduction in operational costs
- **Performant**: Sub-second operations with 50ms function cold starts

The architecture continues to evolve, but the foundation is solid and the principles are clear.# ADR-005: CI/CD Architecture with Provider Abstraction

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
- Canary deployment strategies# ADR Organization Migration Summary

**Date**: 2025-06-12  
**Completed By**: Architecture Team

## Summary

Reorganized 26 ADR files from `/docs/adr/records/` into 8 comprehensive, theme-based ADRs following a consistent format.

## Migration Details

### Original Files Consolidated

1. **Platform Architecture (ADR-001)**
   - Consolidated from implementation details and architecture overview files
   - Focused on vCluster decision and multi-tenancy approach

2. **Authentication & Security (ADR-002)**
   - Merged from: OAuth implementation summary, security architecture docs
   - Comprehensive OAuth2/OIDC with PKCE documentation

3. **Function Service (ADR-003)**
   - Combined: Function service architecture, Fission migration guide, DI architecture
   - Documents evolution from Knative to Fission with provider abstraction

4. **AI Operations (ADR-004)**
   - Merged: AI agent architecture, AIOps architecture files
   - Complete AI/ML integration approach with security model

5. **CI/CD Architecture (ADR-005)**
   - Based on: CICD architecture document
   - Provider abstraction pattern for CI/CD systems

6. **Logging & Monitoring (ADR-006)**
   - From: Logging strategy document
   - ClickHouse decision and implementation details

7. **Backup & DR (ADR-007)**
   - Combined: Backup feature plan, CronJob backup integration
   - Hybrid backup approach documentation

8. **Domain-Driven Design (ADR-008)**
   - Extracted from: Structure guide and implementation patterns
   - Code organization and architectural principles

### Files Archived

The following files were moved to archive as they were redundant or outdated:
- `2025-06-01_implementation-details.md` (too verbose, content extracted)
- `2025-06-09_WORK-STATUS_OLD.md` (explicitly marked as old)
- Project management duplicates (consolidated into implementation status)
- Various work status and immediate next steps files (temporal documents)

### Key Improvements

1. **Consistent Format**: All ADRs now follow the 8-section template
2. **Clear Status**: Each ADR has implementation status clearly marked
3. **Technical Focus**: Removed project management content, focused on architecture
4. **Better Navigation**: Added comprehensive README with index
5. **Historical Context**: Preserved architectural evolution timeline

### Architecture Evolution Timeline

Based on the consolidated ADRs:

- **June 1-3**: Foundation (Platform, Security, DDD)
- **June 6-7**: Core Services (AIOps, CI/CD)  
- **June 8-9**: Operations (Functions, Logging, Backup)
- **June 10-11**: Optimization (Fission migration)

### Next Steps

1. Archive original files to `/docs/adr/archive/`
2. Update all code references to point to new ADR numbers
3. Establish review process for new ADRs
4. Create ADR template file for future use# ADR-004: AI Operations (AIOps) Architecture

**Date**: 2025-06-06  
**Status**: Implemented  
**Authors**: AI Platform Team

## 1. Background

Hexabase AI needed to integrate AI/ML capabilities to provide:
- Intelligent automation for DevOps tasks
- Natural language interface for platform operations
- Predictive analytics for resource optimization
- Automated troubleshooting and remediation
- Context-aware assistance for developers

The challenge was creating a secure, scalable AI operations layer that could access platform resources while maintaining strict security boundaries.

## 2. Status

**Implemented** - Python-based AI operations service with Ollama for local LLM and support for external providers (OpenAI, Anthropic) is deployed.

## 3. Other Options Considered

### Option A: Direct LLM Integration
- Direct API calls to LLM providers
- Simple request/response model
- No intermediate processing layer

### Option B: Agent Framework with LangChain
- LangChain for orchestration
- Multiple specialized agents
- Tool calling capabilities

### Option C: Custom Agent Architecture with Sandbox
- Python-based agent system
- Secure sandbox execution
- Multi-provider LLM support
- Context management system

## 4. What Was Decided

We chose **Option C: Custom Agent Architecture with Sandbox** featuring:
- Python-based AIOps service running in isolated containers
- Secure sandbox model for agent execution
- Support for local (Ollama) and external LLMs
- Context management with user workspace awareness
- Tool registry for platform operations
- Comprehensive audit logging

## 5. Why Did You Choose It?

- **Security**: Sandbox isolation prevents unauthorized access
- **Flexibility**: Support multiple LLM providers
- **Control**: Custom architecture allows fine-tuned permissions
- **Performance**: Local LLM option for sensitive data
- **Scalability**: Containerized architecture scales horizontally

## 6. Why Didn't You Choose the Other Options?

### Why not Direct LLM Integration?
- No security boundaries
- Limited context management
- No tool calling capabilities
- Poor observability

### Why not LangChain?
- Too opinionated for our use case
- Security concerns with arbitrary code execution
- Difficult to customize for Kubernetes operations
- Heavy dependency footprint

## 7. What Has Not Been Decided

- Support for fine-tuned models
- Multi-modal capabilities (image/video analysis)
- Distributed agent coordination
- Real-time learning from user interactions

## 8. Considerations

### Architecture Overview
```
┌──────────────────┐
│   API Gateway    │
└────────┬─────────┘
         │
┌────────▼─────────┐
│  AIOps Service   │
│   (Python)       │
└────────┬─────────┘
         │
┌────────┴─────────┐
│  Agent Manager   │
├──────────────────┤
│ Security Sandbox │
├──────────────────┤
│ Context Manager  │
└──────┬─────┬────┘
       │     │
┌──────▼─┐ ┌─▼──────────┐
│ Ollama │ │External LLM│
└────────┘ └────────────┘
```

### Security Model
```python
class SecuritySandbox:
    def __init__(self, user_context):
        self.user_id = user_context.user_id
        self.org_id = user_context.org_id
        self.permissions = self._load_permissions()
    
    def execute_tool(self, tool_name, params):
        if not self._check_permission(tool_name):
            raise PermissionDenied()
        
        # Execute in isolated environment
        with sandboxed_execution():
            return self.tool_registry[tool_name](**params)
```

### Tool Registry
- **Kubernetes Operations**: List/describe resources, get logs
- **Metric Analysis**: Query Prometheus, analyze trends
- **Code Generation**: Generate YAML, scripts, configurations
- **Troubleshooting**: Analyze errors, suggest fixes

### Context Management
- User workspace and project context
- Historical conversation memory
- Resource access patterns
- Performance baselines

### LLM Provider Configuration
```python
providers = {
    "ollama": {
        "endpoint": "http://ollama:11434",
        "model": "mistral",
        "timeout": 30
    },
    "openai": {
        "endpoint": "https://api.openai.com/v1",
        "model": "gpt-4-turbo",
        "api_key": "${OPENAI_API_KEY}"
    }
}
```

### Performance Considerations
- Response streaming for better UX
- Caching for repeated queries
- Async execution for long-running tasks
- Connection pooling for LLM providers

### Compliance and Audit
- All AI interactions logged
- PII detection and masking
- Model decision explanations
- Usage tracking and quotas

### Future Enhancements
- RAG (Retrieval Augmented Generation) for documentation
- Multi-agent collaboration
- Automated workflow generation
- Continuous learning from outcomes# ADR-006: Logging and Monitoring Architecture

**Date**: 2025-06-09  
**Status**: Implemented  
**Authors**: Platform Observability Team

## 1. Background

Hexabase AI required a comprehensive observability solution that could:
- Handle logs from thousands of containers across multiple clusters
- Provide real-time metrics and alerting
- Support long-term storage for compliance
- Enable efficient debugging and troubleshooting
- Scale cost-effectively with platform growth
- Provide tenant isolation for logs and metrics

Traditional solutions like ELK stack were evaluated but found too expensive at scale.

## 2. Status

**Implemented** - ClickHouse-based logging with Prometheus/Grafana for metrics is fully deployed.

## 3. Other Options Considered

### Option A: ELK Stack (Elasticsearch, Logstash, Kibana)
- Industry standard logging solution
- Rich query capabilities
- Mature ecosystem

### Option B: Loki + Prometheus + Grafana
- Lightweight log aggregation
- Good Kubernetes integration
- Cost-effective

### Option C: ClickHouse + Prometheus + Grafana
- Columnar storage for logs
- Excellent compression ratios
- Fast analytical queries
- Very cost-effective at scale

## 4. What Was Decided

We chose **Option C: ClickHouse + Prometheus + Grafana** with:
- ClickHouse for log storage and analytics
- Prometheus for metrics collection
- Grafana for visualization
- Vector for log shipping
- Custom query API for tenant isolation

## 5. Why Did You Choose It?

- **Cost**: 70% reduction compared to Elasticsearch at scale
- **Performance**: Sub-second queries on billions of log lines
- **Compression**: 10:1 compression ratios typical
- **SQL Interface**: Familiar query language
- **Scalability**: Linear scaling with data volume

## 6. Why Didn't You Choose the Other Options?

### Why not ELK Stack?
- High infrastructure costs
- Complex scaling requirements
- Java heap management issues
- Expensive licensing for enterprise features

### Why not Loki?
- Limited query capabilities
- No full-text search
- Less mature than alternatives
- Performance issues at scale

## 7. What Has Not Been Decided

- Log retention policies beyond 90 days
- Real-time streaming analytics
- Machine learning on logs
- Cross-region log replication

## 8. Considerations

### Architecture Overview
```
┌──────────────┐
│ Applications │
└──────┬───────┘
       │
┌──────▼───────┐
│   Vector     │ (Log Shipper)
└──────┬───────┘
       │
┌──────▼───────┐
│  ClickHouse  │ (Log Storage)
├──────────────┤
│  Prometheus  │ (Metrics)
├──────────────┤
│   Grafana    │ (Visualization)
└──────────────┘
```

### ClickHouse Schema
```sql
CREATE TABLE logs (
    timestamp DateTime64(3),
    workspace_id String,
    project_id String,
    application_id String,
    container_name String,
    log_level LowCardinality(String),
    message String,
    metadata String,
    INDEX idx_message message TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 4
) ENGINE = MergeTree()
PARTITION BY toDate(timestamp)
ORDER BY (workspace_id, timestamp)
TTL timestamp + INTERVAL 90 DAY;
```

### Vector Configuration
```toml
[sources.kubernetes_logs]
type = "kubernetes_logs"

[transforms.parse]
type = "remap"
inputs = ["kubernetes_logs"]
source = '''
.workspace_id = .kubernetes.labels."hexabase.ai/workspace"
.project_id = .kubernetes.labels."hexabase.ai/project"
.application_id = .kubernetes.labels."hexabase.ai/application"
'''

[sinks.clickhouse]
type = "clickhouse"
inputs = ["parse"]
endpoint = "http://clickhouse:8123"
table = "logs"
batch.max_events = 10000
batch.timeout_secs = 10
```

### Query Performance
| Query Type | Data Size | Response Time |
|------------|-----------|---------------|
| Exact match | 1TB | <100ms |
| Wildcard | 1TB | <500ms |
| Aggregation | 1TB | <2s |
| Full scan | 1TB | <10s |

### Tenant Isolation
```go
// Query builder with tenant isolation
func BuildLogQuery(workspace string, query string) string {
    return fmt.Sprintf(`
        SELECT timestamp, log_level, message
        FROM logs
        WHERE workspace_id = '%s'
        AND message LIKE '%%%s%%'
        ORDER BY timestamp DESC
        LIMIT 1000
    `, workspace, query)
}
```

### Metrics Architecture
- **Prometheus**: 15s scrape interval
- **Retention**: 30 days local, 1 year in Thanos
- **Cardinality**: Strict limits per workspace
- **Alerting**: AlertManager with PagerDuty integration

### Cost Analysis
| Component | Monthly Cost (1000 workspaces) |
|-----------|-------------------------------|
| ClickHouse | $500 (3 nodes) |
| Prometheus | $300 (2 nodes) |
| Storage | $200 (10TB) |
| **Total** | **$1,000** |

*Elasticsearch equivalent: ~$3,500*

### Security Considerations
- TLS encryption for all connections
- Query injection prevention
- Rate limiting per workspace
- Audit logging of all queries

### Future Enhancements
- Log anomaly detection
- Automatic pattern extraction
- Correlation with traces
- Cost attribution per workspace# Architecture Decision Records (ADRs) - Organized

This directory contains the organized and consolidated Architecture Decision Records for Hexabase AI. Each ADR follows a consistent format and covers a specific technical theme.

## ADR Index

| ADR                                               | Title                                                 | Status                     | Date       |
| ------------------------------------------------- | ----------------------------------------------------- | -------------------------- | ---------- |
| [ADR-001](001-platform-architecture.md)           | Multi-tenant Kubernetes Platform Architecture         | Implemented                | 2025-06-01 |
| [ADR-002](002-authentication-security.md)         | OAuth2/OIDC Authentication and Security Architecture  | Implemented                | 2025-06-02 |
| [ADR-003](003-function-service-architecture.md)   | Function as a Service (FaaS) Architecture             | Implemented with Migration | 2025-06-08 |
| [ADR-004](004-aiops-architecture.md)              | AI Operations (AIOps) Architecture                    | Implemented                | 2025-06-06 |
| [ADR-005](005-cicd-architecture.md)               | CI/CD Architecture with Provider Abstraction          | Implemented                | 2025-06-07 |
| [ADR-006](006-logging-monitoring-architecture.md) | Logging and Monitoring Architecture                   | Implemented                | 2025-06-09 |
| [ADR-007](007-backup-disaster-recovery.md)        | Backup and Disaster Recovery Architecture             | Partially Implemented      | 2025-06-09 |
| [ADR-008](008-domain-driven-design.md)            | Domain-Driven Design and Code Organization            | Implemented and Enforced   | 2025-06-03 |
| [ADR-009](009-secure-logging-architecture.md)     | Secure and Scalable Multi-Tenant Logging Architecture | Proposed                   | 2025-06-15 |

## ADR Format

Each ADR follows this structure:

1. **Background** - Context and problem statement
2. **Status** - Current implementation status
3. **Other options considered** - Alternative solutions evaluated
4. **What was decided** - The chosen solution
5. **Why did you choose it?** - Rationale for the decision
6. **Why didn't you choose the other option** - Reasons for rejecting alternatives
7. **What has not been decided** - Open questions and future decisions
8. **Considerations** - Implementation details, security, performance, etc.

## Architecture Evolution Timeline

### June 1-3, 2025: Foundation

- Platform architecture with vCluster
- Security architecture with OAuth2/OIDC
- Domain-driven design adoption

### June 6-7, 2025: Core Services

- AI Operations architecture
- CI/CD provider abstraction

### June 8-9, 2025: Operations & Scaling

- Function service with Fission migration
- Logging with ClickHouse
- Backup and disaster recovery

### June 10-11, 2025: Optimization

- Fission provider implementation
- Performance improvements

## Key Architectural Principles

1. **Provider Abstraction** - Avoid vendor lock-in with clean interfaces
2. **Security First** - Zero-trust, encryption, audit trails
3. **Performance Focused** - Optimize for sub-second operations
4. **Cost Effective** - Choose solutions that scale economically
5. **Domain Driven** - Clear boundaries and ownership
6. **Test Driven** - 90% coverage requirement

## Using These ADRs

- For new features, check relevant ADRs for patterns and decisions
- When proposing changes, create new ADRs following the format
- Update ADR status when implementations change
- Link to ADRs in code comments for context
# ADR-002: OAuth2/OIDC Authentication and Security Architecture

**Date**: 2025-06-02  
**Status**: Implemented  
**Authors**: Security Team

## 1. Background

Hexabase AI required a robust authentication and authorization system that could:
- Support multiple identity providers (GitHub, Google, Microsoft, etc.)
- Enable Single Sign-On (SSO) across all platform services
- Provide secure API access for programmatic clients
- Support multi-factor authentication
- Scale to thousands of concurrent users
- Maintain audit trails for compliance

The platform handles sensitive customer workloads and needed enterprise-grade security.

## 2. Status

**Implemented** - OAuth2/OIDC authentication with PKCE, JWT fingerprinting, and comprehensive audit logging is fully deployed.

## 3. Other Options Considered

### Option A: Basic JWT with Local User Database
- Local user management with bcrypt passwords
- Simple JWT tokens for session management
- Custom RBAC implementation

### Option B: SAML 2.0 Integration
- Enterprise SAML SSO
- XML-based assertions
- Session management via SAML

### Option C: OAuth2/OIDC with Enhanced Security
- OAuth2 with PKCE flow
- JWT with fingerprinting
- Integration with external IdPs
- Redis-based session management

## 4. What Was Decided

We chose **Option C: OAuth2/OIDC with Enhanced Security** featuring:
- OAuth2 authorization code flow with PKCE
- JWT tokens with browser fingerprinting
- Support for multiple OIDC providers
- Redis session storage with 24-hour expiry
- Comprehensive audit logging
- Rate limiting and DDoS protection

## 5. Why Did You Choose It?

- **Industry Standard**: OAuth2/OIDC is widely supported and understood
- **Security**: PKCE prevents authorization code interception attacks
- **Flexibility**: Easy to add new identity providers
- **User Experience**: Seamless SSO with existing accounts
- **Auditability**: Comprehensive logging for compliance requirements

## 6. Why Didn't You Choose the Other Options?

### Why not Basic JWT?
- No SSO capability
- Password management overhead
- Less secure than delegated authentication
- No built-in MFA support

### Why not SAML 2.0?
- Complex XML processing
- Poor mobile/SPA support
- Heavier protocol overhead
- Less developer-friendly

## 7. What Has Not Been Decided

- Support for WebAuthn/FIDO2 passwordless authentication
- Integration with enterprise Active Directory
- Advanced threat detection mechanisms
- Biometric authentication support

## 8. Considerations

### Security Considerations
- Regular rotation of signing keys
- Monitoring for abnormal authentication patterns
- Protection against token replay attacks
- Secure storage of OAuth client secrets

### Performance Considerations
- Redis clustering for session storage scale
- JWT validation caching strategies
- Optimize OIDC discovery endpoint calls

### Compliance Considerations
- GDPR compliance for user data
- SOC2 audit trail requirements
- Right to deletion implementation
- Data residency requirements

### Implementation Details

```go
// JWT with fingerprinting implementation
type EnhancedClaims struct {
    jwt.StandardClaims
    Fingerprint string `json:"fingerprint"`
    UserID      string `json:"user_id"`
    OrgID       string `json:"org_id"`
}

// PKCE verification
func VerifyPKCE(codeVerifier, codeChallenge string) bool {
    hash := sha256.Sum256([]byte(codeVerifier))
    computed := base64.RawURLEncoding.EncodeToString(hash[:])
    return computed == codeChallenge
}
```

### Future Enhancements
- Zero-trust network architecture integration
- Risk-based authentication
- Continuous authentication mechanisms
- Decentralized identity support# ADR-001: Multi-tenant Kubernetes Platform Architecture

**Date**: 2025-06-01  
**Status**: Implemented  
**Authors**: Platform Architecture Team

## 1. Background

Hexabase AI needed to build a Kubernetes-as-a-Service platform that provides isolated environments for multiple tenants while maintaining cost efficiency and operational simplicity. The platform needed to support various workload types including stateless applications, stateful services, scheduled jobs, and serverless functions.

The key challenges were:
- Complete isolation between tenants for security and compliance
- Cost-effective resource utilization
- Support for both shared and dedicated infrastructure
- Easy scaling and management
- Integration with existing Hexabase ecosystem

## 2. Status

**Implemented** - The core platform architecture using K3s as the host cluster and vCluster for tenant isolation has been deployed and is operational.

## 3. Other Options Considered

### Option A: Namespace-based Multi-tenancy
- Use Kubernetes namespaces for tenant isolation
- Network policies for inter-tenant communication control
- ResourceQuotas and LimitRanges for resource management

### Option B: Multiple K8s Clusters
- Deploy separate Kubernetes clusters for each tenant
- Use cluster federation for management
- Direct infrastructure provisioning

### Option C: vCluster-based Virtual Clusters
- Single host K3s cluster
- vCluster for each tenant workspace
- Shared control plane with isolated data planes

## 4. What Was Decided

We chose **Option C: vCluster-based Virtual Clusters** with the following architecture:
- K3s as the lightweight host cluster
- vCluster for complete Kubernetes API isolation per tenant
- Two plans: Shared (multiple vClusters per node) and Dedicated (exclusive nodes)
- Integration with Proxmox for dedicated node provisioning

## 5. Why Did You Choose It?

- **Complete API Isolation**: Each tenant gets their own Kubernetes API server, preventing any cross-tenant API access
- **Cost Efficiency**: Multiple vClusters can run on shared infrastructure
- **Flexibility**: Easy to migrate tenants between shared and dedicated plans
- **Operational Simplicity**: Single host cluster to manage
- **Native Kubernetes**: Tenants get full Kubernetes API compatibility

## 6. Why Didn't You Choose the Other Options?

### Why not Namespace-based Multi-tenancy?
- Insufficient isolation for compliance requirements
- Shared API server creates security risks
- Complex RBAC management
- Limited ability to customize per-tenant

### Why not Multiple K8s Clusters?
- High operational overhead
- Expensive infrastructure requirements
- Complex networking between clusters
- Difficult to manage at scale

## 7. What Has Not Been Decided

- Long-term strategy for cross-region deployment
- Disaster recovery approach for vClusters
- Migration path for very large tenants (>100 nodes)
- Integration with other cloud providers beyond Proxmox

## 8. Considerations

### Security Considerations
- Each vCluster runs with restricted permissions
- Network policies enforce traffic isolation
- Regular security audits of the vCluster implementation

### Performance Considerations
- Monitor vCluster control plane overhead
- Optimize scheduling for shared infrastructure
- Regular capacity planning reviews

### Operational Considerations
- Automated vCluster lifecycle management
- Monitoring and alerting strategy
- Backup and restore procedures

### Future Considerations
- Evaluate newer isolation technologies as they emerge
- Consider contributing improvements back to vCluster project
- Plan for potential migration to CNCF sandbox projects# ADR-009: Secure and Scalable Multi-Tenant Logging Architecture

**Date**: 2025-06-15
**Tags**: logging, security, observability, multi-tenancy

## 1. Background

As a multi-tenant platform, HKS must provide robust logging and auditing capabilities for different audiences (Platform SREs, Organization Admins, Users). Our current structured logs are a good technical foundation but lack the tenancy and user context (`organization_id`, `user_id`) required to build secure, role-based access controls. This poses a significant security risk, as there is no mechanism to prevent one tenant from potentially accessing another's logs. We need a formal architecture to guarantee strict data isolation.

## 2. Status

**Proposed**

## 3. Other Options Considered

### Option A: Frontend-based Filtering

- The backend API provides a broad set of logs, and the frontend UI is responsible for filtering them based on the user's role.

### Option B: Per-Tenant Data Stores

- Provision a completely separate logging stack (e.g., a dedicated Loki and ClickHouse instance) for each tenant organization.

### Option C: Centralized Storage with Backend-Enforced Access Control

- Use a shared, centralized logging backend (Loki/ClickHouse) for all tenants. All client access is brokered through a secure HKS API gateway that enriches logs and strictly enforces access control rules based on the authenticated user's identity.

## 4. What Was Decided

We will implement **Option C: Centralized Storage with Backend-Enforced Access Control**.

The core components of this decision are:

1.  **Three-Tiered Logging**: Logs are categorized into `System Logs` (for SREs), `Audit Logs` (for users/admins), and `Workload Logs` (for application owners), each with its own storage and access rules.
2.  **Mandatory Log Enrichment**: The API backend will implement a logging middleware to automatically inject `organization_id`, `workspace_id`, and `user_id` into every structured log generated within an authenticated user's request context.
3.  **Secure API Proxy**: The UI and other clients will **never** access logging data stores directly. All requests will be proxied through the HKS API, which acts as a secure gatekeeper.
4.  **Backend-Enforced Scoping**:
    - For **Audit Logs**, the API will inject non-negotiable `WHERE` clauses into ClickHouse queries based on the user's JWT claims.
    - For **Workload Logs**, the API will perform a Kubernetes `SubjectAccessReview` to validate the user's pod-level permissions before streaming any log data.

## 5. Why Did You Choose It?

This approach provides the best balance of security, scalability, and usability:

- **Security**: It establishes a single, secure gateway for all log access, eliminating the possibility of client-side bypass and ensuring that access control logic is centralized and consistently enforced.
- **Scalability**: It leverages shared infrastructure, which is more cost-effective and operationally manageable than provisioning a separate stack for every tenant.
- **Usability**: It enables the platform to provide rich, role-aware observability features within the UI, powered by a secure and flexible API.

## 6. Why Didn't You Choose the Other Option?

- **Option A (Frontend Filtering)** was rejected because it is fundamentally insecure. Relying on the client to enforce security is a critical anti-pattern; a malicious user could easily bypass the UI and query the API directly to access unauthorized data.
- **Option B (Per-Tenant Stacks)** was rejected due to excessive operational complexity and cost. Managing thousands of separate logging instances is not feasible and would not scale economically.

## 7. What Has Not Been Decided

- The precise log retention policies (e.g., 7 vs. 30 vs. 90 days) for each log category.
- The specific design and content of the default Grafana dashboards for each user role.
- The exact schema for the `details_json` field in the audit log table.

## 8. Considerations

- **Performance**: The logging enrichment middleware must be highly performant to avoid adding significant latency to API requests.
- **Implementation**: Requires disciplined implementation of the secure API endpoints and careful construction of the database/Loki queries to prevent injection or scoping bugs.
- **AIOps**: The AIOps agent must also have its actions logged through this system, correctly impersonating the user on whose behalf it is acting. The `initiated_by` field will be used to differentiate user vs. agent actions.
# ADR-007: Backup and Disaster Recovery Architecture

**Date**: 2025-06-09  
**Status**: Partially Implemented  
**Authors**: Infrastructure Team

## 1. Background

Hexabase AI needed a comprehensive backup and disaster recovery solution for:
- User application data (persistent volumes)
- Platform configuration and state
- Database backups (PostgreSQL, Redis)
- Disaster recovery with RTO < 4 hours, RPO < 1 hour
- Compliance with data retention requirements
- Cost-effective storage for long-term retention

The solution needed to support both shared and dedicated plan customers with different SLAs.

## 2. Status

**Partially Implemented** - Backup features for dedicated plans are implemented. Shared plan backups and full DR are in progress.

## 3. Other Options Considered

### Option A: Velero-based Kubernetes Backup
- Native Kubernetes backup tool
- Snapshot-based backups
- Cloud provider integration

### Option B: Storage-level Snapshots
- Direct storage snapshots
- Proxmox backup integration
- Volume-level consistency

### Option C: Hybrid Application-aware Backup
- Combination of Velero and storage snapshots
- Application-specific backup strategies
- Tiered storage approach

## 4. What Was Decided

We chose **Option C: Hybrid Application-aware Backup** with:
- Velero for Kubernetes resource backup
- Proxmox snapshots for dedicated plan storage
- Application-aware backups for databases
- CronJob integration for scheduled backups
- S3-compatible object storage for long-term retention

## 5. Why Did You Choose It?

- **Flexibility**: Different strategies for different data types
- **Consistency**: Application-aware backups ensure data integrity
- **Cost-effective**: Tiered storage reduces long-term costs
- **Performance**: Minimal impact on running workloads
- **Compliance**: Meets retention and recovery requirements

## 6. Why Didn't You Choose the Other Options?

### Why not Velero-only?
- Limited support for external storage systems
- No application-level consistency
- Slower recovery for large volumes

### Why not Storage Snapshots only?
- No Kubernetes resource backup
- Platform-specific limitations
- Difficult cross-platform recovery

## 7. What Has Not Been Decided

- Cross-region disaster recovery implementation
- Backup encryption key management strategy
- Automated DR testing procedures
- Backup cost allocation model

## 8. Considerations

### Backup Architecture
```
┌─────────────────┐
│   Workspaces    │
└────────┬────────┘
         │
┌────────┴────────┐
│  Backup Service │
├─────────────────┤
│ Backup Policies │
├─────────────────┤
│ Storage Manager │
└───┬─────────┬───┘
    │         │
┌───▼───┐ ┌───▼────┐
│Velero │ │Proxmox │
│       │ │Backup  │
└───┬───┘ └───┬────┘
    │         │
┌───▼─────────▼───┐
│ S3 Object Store │
└─────────────────┘
```

### Backup Types and Strategies

| Backup Type | Method | Frequency | Retention |
|-------------|--------|-----------|-----------|
| Platform Config | Velero | Daily | 30 days |
| Application PVs | Snapshot | Hourly | 7 days |
| PostgreSQL | pg_dump | 4 hours | 30 days |
| Redis | RDB snapshot | Daily | 7 days |
| Full DR | All above | Weekly | 90 days |

### Backup Policy Configuration
```go
type BackupPolicy struct {
    ID          string
    WorkspaceID string
    Schedule    string // Cron expression
    Retention   time.Duration
    Type        BackupType
    Targets     []BackupTarget
    Encryption  EncryptionConfig
}

type BackupTarget struct {
    Type     string // "pvc", "database", "all"
    Selector map[string]string
}
```

### Recovery Procedures

#### Application Recovery
1. Restore Kubernetes resources via Velero
2. Restore persistent volumes from snapshots
3. Verify application health
4. Update DNS/routing

#### Database Recovery
```bash
# PostgreSQL point-in-time recovery
pg_restore -h localhost -U postgres \
  --clean --create -d postgres \
  backup_2025_06_09_1200.dump

# Redis recovery
redis-cli --rdb /backup/dump.rdb
```

### Storage Tiers
- **Hot**: Last 7 days - Fast SSD storage
- **Warm**: 8-30 days - Standard storage
- **Cold**: 31-365 days - Archive storage

### Encryption
- AES-256 encryption at rest
- Per-workspace encryption keys
- Key rotation every 90 days
- Hardware security module (HSM) for key storage

### Cost Model
| Plan | Backup Frequency | Retention | Monthly Cost |
|------|-----------------|-----------|--------------|
| Shared | Daily | 7 days | Included |
| Dedicated | Hourly | 30 days | $50/TB |
| Enterprise | 15 min | 365 days | Custom |

### Monitoring and Alerting
- Backup job success/failure alerts
- Storage capacity warnings
- Recovery time tracking
- Backup size trends

### Compliance Considerations
- GDPR right to deletion
- Data residency requirements
- Audit trail of all backup operations
- Encryption key escrow for compliance

### Future Enhancements
- Continuous data protection (CDP)
- Cross-region replication
- Automated DR drills
- AI-powered backup optimization# ADR-003: Function as a Service (FaaS) Architecture

**Date**: 2025-06-08  
**Status**: Implemented with Migration in Progress  
**Authors**: Platform Team

## 1. Background

Hexabase AI needed to provide serverless function capabilities to users for:
- Event-driven compute workloads
- API endpoints without managing servers
- Scheduled function execution
- Auto-scaling based on demand
- Multi-language support (Node.js, Python, Go)
- Integration with the broader platform

Initial implementation used Knative but performance issues led to a provider abstraction layer and migration to Fission.

## 2. Status

**Implemented with Migration in Progress** - Provider abstraction is complete. Fission is the default provider. Knative support maintained for backward compatibility.

## 3. Other Options Considered

### Option A: Knative Only
- Industry standard Kubernetes serverless
- Good ecosystem support
- Built on Istio service mesh

### Option B: OpenFaaS
- Simple function deployment
- Good community support
- Multiple language templates

### Option C: Fission
- Fast cold starts (50-200ms)
- Simple architecture
- Good Kubernetes integration

### Option D: Provider Abstraction Layer
- Support multiple FaaS backends
- Easy migration between providers
- Future flexibility

## 4. What Was Decided

We chose **Option D: Provider Abstraction Layer** with:
- Clean provider interface for function operations
- Fission as the default provider (95% faster cold starts)
- Knative support for compatibility
- Dependency injection for provider selection
- Unified function management API

## 5. Why Did You Choose It?

- **Performance**: Fission provides 50-200ms cold starts vs 2-5s for Knative
- **Flexibility**: Can switch providers without changing application code
- **Future-proof**: Easy to add new providers as they emerge
- **User Experience**: Faster function execution improves user satisfaction
- **Cost Efficiency**: Reduced compute time means lower costs

## 6. Why Didn't You Choose the Other Options?

### Why not Knative Only?
- Unacceptable cold start performance (2-5 seconds)
- Complex Istio dependency
- Higher resource overhead

### Why not OpenFaaS?
- Less mature than other options
- Limited enterprise features
- Smaller ecosystem

### Why not Fission Only?
- Lock-in to single provider
- No migration path for existing Knative users
- Limited flexibility for future needs

## 7. What Has Not Been Decided

- Long-term Knative deprecation timeline
- Support for WebAssembly functions
- Edge function deployment strategy
- GPU-accelerated function support

## 8. Considerations

### Performance Metrics
| Provider | Cold Start | Warm Start | Memory Overhead |
|----------|------------|------------|-----------------|
| Knative  | 2-5s       | 100-200ms  | 512MB          |
| Fission  | 50-200ms   | 10-50ms    | 128MB          |

### Migration Considerations
```go
// Provider interface enabling seamless migration
type FunctionProvider interface {
    CreateFunction(ctx context.Context, req CreateFunctionRequest) (*Function, error)
    InvokeFunction(ctx context.Context, name string, data []byte) ([]byte, error)
    DeleteFunction(ctx context.Context, name string) error
    ListFunctions(ctx context.Context) ([]*Function, error)
}
```

### Implementation Architecture
```
┌─────────────────┐
│ Function Service │
└────────┬────────┘
         │
    ┌────▼─────┐
    │ Provider │
    │ Interface│
    └─────┬────┘
          │
    ┌─────┴──────┬─────────────┐
    │            │             │
┌───▼──┐    ┌───▼───┐    ┌────▼───┐
│Fission│    │Knative│    │Future  │
└───────┘    └───────┘    │Provider│
                          └────────┘
```

### Security Considerations
- Function isolation via gVisor
- Network policies for function communication
- Secret injection mechanisms
- Resource limits enforcement

### Operational Considerations
- Automated function deployment pipelines
- Monitoring and tracing integration
- Log aggregation strategies
- Auto-scaling policies

### Future Roadmap
1. Complete Knative to Fission migration (Q3 2025)
2. Add WebAssembly support (Q4 2025)
3. Implement edge function deployment (Q1 2026)
4. GPU function support for ML workloads (Q2 2026)