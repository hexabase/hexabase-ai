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
5. Update clients to use new endpoint