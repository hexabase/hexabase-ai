# ADR-003: Function as a Service (FaaS) Architecture

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