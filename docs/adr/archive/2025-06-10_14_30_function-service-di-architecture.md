# Function Service DI Architecture with Fission as Default Provider

**Date**: 2025-06-10 14:30  
**Status**: Proposed  
**Deciders**: Architecture Team, Platform Team  
**Tags**: architecture, functions, serverless, dependency-injection, testing

## Context

The HKS platform currently assumes Knative as the Function-as-a-Service (FaaS) implementation. However, we've identified several limitations:

- Heavy resource footprint due to Istio/Kourier requirements
- Slow cold starts (2-5 seconds)
- Complex operational overhead
- Limited runtime support

Fission offers a more lightweight alternative with:

- Faster cold starts (50-200ms with pooling)
- Built-in builder service
- Wider language support
- Better resource efficiency

Additionally, we need flexibility to support different FaaS providers per workspace/project and ensure smooth migration paths.

## Decision

We will implement a Provider abstraction using Dependency Injection (DI) and Strategy pattern, with Fission as the default provider and Knative as an optional alternative.

### Core Design Principles

1. **Provider Interface**: Abstract all FaaS operations behind a clean interface
2. **Dependency Injection**: Use constructor injection for testability
3. **Test-Driven Development**: Write tests first, ensuring >90% coverage
4. **Factory Pattern**: Enable runtime provider selection
5. **Graceful Migration**: Support gradual migration from Knative to Fission

### Architecture Overview

```go
// Core abstraction
type Provider interface {
    CreateFunction(ctx context.Context, spec *FunctionSpec) (*Function, error)
    UpdateFunction(ctx context.Context, name string, spec *FunctionSpec) (*Function, error)
    DeleteFunction(ctx context.Context, name string) error
    GetFunction(ctx context.Context, name string) (*Function, error)
    InvokeFunction(ctx context.Context, name string, req *InvokeRequest) (*InvokeResponse, error)
    GetCapabilities() *Capabilities
}
```

### Implementation Plan

#### Phase 1: Core Abstractions (Week 1)

1. **Define Interfaces**

   - `Provider` interface
   - `Factory` interface
   - Common types (`FunctionSpec`, `Runtime`, `Trigger`)

2. **TDD Implementation**
   ```go
   // Start with tests
   func TestProviderInterface(t *testing.T) {
       mock := NewMockProvider()
       testCreateFunction(t, mock)
       testUpdateFunction(t, mock)
       testDeleteFunction(t, mock)
   }
   ```

#### Phase 2: Fission Provider (Week 2)

1. **Fission Client Wrapper**

   - Environment management
   - Package handling
   - Trigger creation

2. **Test Suite**
   ```go
   func TestFissionProvider(t *testing.T) {
       // Unit tests with mocked Fission client
       // Integration tests with test cluster
       // Performance benchmarks
   }
   ```

#### Phase 3: Knative Provider as Optional (Week 3)

1. **Knative Adapter**

   - Service management
   - Image building integration
   - Route handling

2. **Compatibility Tests**
   ```go
   func TestProviderCompatibility(t *testing.T) {
       providers := []Provider{
           NewFissionProvider(config),
           NewKnativeProvider(config),
       }
       for _, p := range providers {
           testProviderBehavior(t, p)
       }
   }
   ```

#### Phase 4: Service Layer Integration (Week 4)

1. **FunctionService with DI**

   ```go
   type FunctionService struct {
       providerFactory Factory
       db              Database
       auth            Authorizer
   }

   func NewFunctionService(deps Dependencies) *FunctionService {
       return &FunctionService{
           providerFactory: deps.ProviderFactory,
           db:              deps.Database,
           auth:            deps.Authorizer,
       }
   }
   ```

2. **Configuration Management**
   ```yaml
   function:
     defaultProvider: fission
     providers:
       fission:
         routerUrl: http://router.fission
         poolSize: 3
       knative:
         enabled: false # Optional, disabled by default
         domain: example.com
   ```

### Testing Strategy

1. **Unit Tests** (Target: 95% coverage)

   - Mock all external dependencies
   - Test each provider in isolation
   - Verify error handling

2. **Integration Tests**

   - Real Fission cluster tests
   - Optional Knative tests (if enabled)
   - End-to-end function lifecycle

3. **Contract Tests**

   - Ensure all providers meet interface contract
   - Verify consistent behavior across providers

4. **Performance Tests**
   ```go
   func BenchmarkColdStart(b *testing.B) {
       providers := getTestProviders()
       for name, provider := range providers {
           b.Run(name, func(b *testing.B) {
               benchmarkColdStart(b, provider)
           })
       }
   }
   ```

### Migration Path

1. **New Installations**: Fission by default
2. **Existing Installations**:
   - Keep Knative support via configuration
   - Provide migration tooling
   - Document migration process

## Consequences

### Positive

- **Flexibility**: Switch providers without code changes
- **Testability**: Easy to mock and test in isolation
- **Performance**: 10-40x faster cold starts with Fission
- **Cost**: Reduced resource usage (no service mesh required)
- **Developer Experience**: Built-in builder, wider language support
- **Future-proof**: Can add more providers (OpenFaaS, AWS Lambda)

### Negative

- **Complexity**: Additional abstraction layer
- **Maintenance**: Need to maintain multiple provider implementations
- **Feature Parity**: Some Knative-specific features need adaptation
- **Training**: Team needs to learn Fission operations

### Risks

- **Migration Errors**: Potential for data loss during migration
  - _Mitigation_: Comprehensive migration tooling and rollback procedures
- **Performance Regression**: Different scaling characteristics
  - _Mitigation_: Thorough benchmarking and gradual rollout
- **Operational Knowledge**: New debugging and monitoring patterns
  - _Mitigation_: Documentation and training programs

## Alternatives Considered

1. **Hard Switch to Fission**
   - Rejected: Too risky for existing users
2. **Maintain Both Separately**
   - Rejected: Code duplication and maintenance burden
3. **Build Custom FaaS**
   - Rejected: Significant engineering effort, reinventing the wheel

## Implementation Checklist

- [ ] Provider interface definition
- [ ] Mock provider implementation
- [ ] Comprehensive test suite setup
- [ ] Fission provider implementation
- [ ] Fission provider tests (unit, integration, performance)
- [ ] Knative provider adapter
- [ ] Provider factory implementation
- [ ] Service layer refactoring
- [ ] Configuration management
- [ ] Migration tooling
- [ ] Documentation updates
- [ ] Performance benchmarks
- [ ] Deployment guides
- [ ] Monitoring setup

## References

- [Fission Documentation](https://fission.io/docs/)
- [Knative Serving Docs](https://knative.dev/docs/serving/)
- [Go DI Best Practices](https://github.com/google/wire)
- [Testing Strategies for Microservices](https://martinfowler.com/articles/microservice-testing/)
- HKS Function Service Requirements (internal doc)
