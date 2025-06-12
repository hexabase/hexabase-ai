# Fission Provider Implementation Plan

**Date**: 2025-06-10 17:23  
**Status**: In Progress  
**Deciders**: Architecture Team, Platform Team  
**Tags**: functions, serverless, fission, knative, provider-pattern

## Context

Following the decision in [2025-06-10_14_30_function-service-di-architecture.md](./2025-06-10_14_30_function-service-di-architecture.md) to implement a provider abstraction for function services, this document outlines the detailed implementation plan for refactoring our function service to use Fission as the default FaaS provider while maintaining Knative as an optional alternative.

## Implementation Plan

### Phase 1: Provider Interface Design (Days 1-3)

#### 1.1 Core Provider Interface
Create the provider abstraction layer with clean interfaces:

```go
// api/internal/domain/function/provider.go
type Provider interface {
    // Lifecycle management
    CreateFunction(ctx context.Context, spec *FunctionSpec) (*Function, error)
    UpdateFunction(ctx context.Context, name string, spec *FunctionSpec) (*Function, error)
    DeleteFunction(ctx context.Context, name string) error
    GetFunction(ctx context.Context, name string) (*Function, error)
    ListFunctions(ctx context.Context, namespace string) ([]*Function, error)
    
    // Version management
    CreateVersion(ctx context.Context, functionName string, version *Version) error
    GetVersion(ctx context.Context, functionName, versionID string) (*Version, error)
    ListVersions(ctx context.Context, functionName string) ([]*Version, error)
    SetActiveVersion(ctx context.Context, functionName, versionID string) error
    
    // Invocation
    InvokeFunction(ctx context.Context, name string, req *InvokeRequest) (*InvokeResponse, error)
    GetFunctionURL(ctx context.Context, name string) (string, error)
    
    // Capabilities
    GetCapabilities() *Capabilities
}
```

#### 1.2 Common Types
Define types shared across all providers:

```go
// api/internal/domain/function/types.go
type FunctionSpec struct {
    Name        string
    Namespace   string
    Runtime     Runtime
    Handler     string
    SourceCode  string
    Environment map[string]string
    Resources   ResourceRequirements
    Triggers    []Trigger
}

type Runtime string
const (
    RuntimeGo         Runtime = "go"
    RuntimePython     Runtime = "python"
    RuntimeNode       Runtime = "node"
    RuntimeJava       Runtime = "java"
    RuntimeDotNet     Runtime = "dotnet"
)

type TriggerType string
const (
    TriggerHTTP     TriggerType = "http"
    TriggerSchedule TriggerType = "schedule"
    TriggerEvent    TriggerType = "event"
)
```

#### 1.3 Factory Pattern
Implement provider factory for runtime selection:

```go
// api/internal/domain/function/factory.go
type ProviderFactory interface {
    CreateProvider(config ProviderConfig) (Provider, error)
    GetSupportedProviders() []string
}
```

### Phase 2: Mock Provider Implementation (Day 4)

Create a comprehensive mock provider for testing:

```go
// api/internal/repository/function/mock/provider.go
type MockProvider struct {
    functions map[string]*function.Function
    versions  map[string][]*function.Version
    mu        sync.RWMutex
}
```

### Phase 3: Service Layer Refactoring (Days 5-7)

#### 3.1 Update Application Service
Refactor the existing service to use the provider interface:

```go
// api/internal/service/application/service.go
type Service struct {
    repo            application.Repository
    k8s             application.KubernetesRepository
    functionProvider function.Provider  // New field
    logger          *slog.Logger
}
```

#### 3.2 Extract Function Logic
Move function-specific logic to use the provider:
- Replace direct Knative calls with provider interface
- Update CreateFunction, DeployFunctionVersion, InvokeFunction methods
- Maintain backward compatibility

### Phase 4: Knative Provider Adapter (Days 8-9)

Create an adapter for the existing Knative implementation:

```go
// api/internal/repository/function/knative/provider.go
type KnativeProvider struct {
    k8sClient    kubernetes.Interface
    servingClient serving.Interface
    config       *KnativeConfig
}
```

### Phase 5: Fission Provider Implementation (Days 10-14)

#### 5.1 Fission Client
Implement Fission API client:

```go
// api/internal/repository/function/fission/client.go
type FissionClient struct {
    controller string
    router     string
    httpClient *http.Client
}
```

#### 5.2 Core Provider
Implement the provider interface for Fission:

```go
// api/internal/repository/function/fission/provider.go
type FissionProvider struct {
    client      *FissionClient
    namespace   string
    environments map[Runtime]*Environment
}
```

#### 5.3 Environment Management
Handle Fission runtime environments:

```go
// api/internal/repository/function/fission/environment.go
func (p *FissionProvider) ensureEnvironment(runtime Runtime) (*Environment, error) {
    // Create or get environment for runtime
}
```

### Phase 6: Provider Factory Implementation (Day 15)

Implement concrete factory:

```go
// api/internal/repository/function/factory.go
type DefaultProviderFactory struct {
    knativeConfig *KnativeConfig
    fissionConfig *FissionConfig
}

func (f *DefaultProviderFactory) CreateProvider(config ProviderConfig) (Provider, error) {
    switch config.Type {
    case "fission":
        return fission.NewProvider(f.fissionConfig)
    case "knative":
        return knative.NewProvider(f.knativeConfig)
    case "mock":
        return mock.NewProvider()
    default:
        return nil, fmt.Errorf("unsupported provider: %s", config.Type)
    }
}
```

### Phase 7: Configuration Management (Day 16)

#### 7.1 Update Configuration
Add function provider configuration:

```yaml
function:
  defaultProvider: fission
  providers:
    fission:
      controllerUrl: http://controller.fission
      routerUrl: http://router.fission
      builderUrl: http://builder.fission
      storageUrl: http://storagesvc.fission
      poolSize: 3
      specializedPoolSize: 1
    knative:
      enabled: false
      domain: example.com
```

#### 7.2 Database Migration
Add provider selection to workspaces:

```sql
-- api/internal/db/migrations/XXX_add_function_provider.up.sql
ALTER TABLE workspaces 
ADD COLUMN function_provider VARCHAR(50) DEFAULT 'fission';

ALTER TABLE applications
ADD COLUMN function_provider_metadata JSONB;
```

### Phase 8: Wire Integration (Day 17)

Update dependency injection:

```go
// api/internal/infrastructure/wire/wire.go
func InitializeFunctionProvider(config *config.Config) (function.Provider, error) {
    factory := function.NewProviderFactory(config)
    return factory.CreateProvider(config.Function.DefaultProvider)
}
```

### Phase 9: Testing & Benchmarks (Days 18-19)

#### 9.1 Contract Tests
Ensure all providers meet the same contract:

```go
// api/internal/repository/function/provider_contract_test.go
func TestProviderContract(t *testing.T) {
    providers := []function.Provider{
        mock.NewProvider(),
        // Add real providers in integration tests
    }
    
    for _, provider := range providers {
        t.Run(provider.GetCapabilities().Name, func(t *testing.T) {
            testCreateFunction(t, provider)
            testUpdateFunction(t, provider)
            testInvokeFunction(t, provider)
            // ... more tests
        })
    }
}
```

#### 9.2 Performance Benchmarks
Compare provider performance:

```go
// api/internal/repository/function/benchmark_test.go
func BenchmarkColdStart(b *testing.B) {
    providers := getTestProviders()
    for name, provider := range providers {
        b.Run(name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                // Benchmark cold start
            }
        })
    }
}
```

### Phase 10: Documentation & Migration (Day 20)

#### 10.1 Update Architecture Docs
- Update function-service-architecture.md
- Create migration guide
- Document provider capabilities

#### 10.2 Migration Tools
Create tools for migrating existing functions:

```go
// cmd/migrate-functions/main.go
func migrateKnativeToFission(workspace string) error {
    // Migration logic
}
```

## Testing Strategy

### Unit Tests
- Mock provider: 95% coverage
- Each provider implementation: 90% coverage
- Factory and configuration: 100% coverage

### Integration Tests
- Test with real Fission cluster
- Test with real Knative cluster (if available)
- End-to-end function lifecycle

### Performance Tests
- Cold start benchmarks
- Concurrent invocation tests
- Resource usage comparison

## Migration Path

### New Installations
- Use Fission by default
- Knative optional via configuration

### Existing Installations
1. Add provider configuration
2. Run migration tool
3. Verify functionality
4. Switch provider at workspace level
5. Monitor and rollback if needed

## Success Criteria

1. ✅ All existing function tests pass with both providers
2. ✅ Fission cold start < 200ms with pooling
3. ✅ Resource usage reduced by 50%+ compared to Knative
4. ✅ Zero-downtime provider switching
5. ✅ No breaking API changes

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Migration failures | High | Comprehensive backup and rollback procedures |
| Performance regression | Medium | Thorough benchmarking before migration |
| Provider incompatibility | Medium | Contract tests ensure compatibility |
| Operational complexity | Low | Clear documentation and training |

## Timeline

- **Week 1**: Provider interface and mock implementation
- **Week 2**: Service refactoring and Knative adapter
- **Week 3**: Fission provider implementation
- **Week 4**: Integration, testing, and documentation

Total estimated time: 20 working days

## References

- [Fission Architecture](https://fission.io/docs/architecture/)
- [Knative Serving API](https://knative.dev/docs/serving/spec/knative-api-specification-1.0/)
- [Go Interface Best Practices](https://go.dev/doc/effective_go#interfaces)
- [Provider Pattern in Go](https://refactoring.guru/design-patterns/strategy/go/example)