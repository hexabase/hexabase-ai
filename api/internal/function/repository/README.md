# Function Provider Implementation

This package implements a provider abstraction pattern for serverless function platforms, supporting both Knative and Fission.

## Structure

```
function/
├── provider.go          # Provider interface definition
├── types.go            # Common types
├── capabilities.go     # Provider capabilities
├── factory.go          # Provider factory
├── mock/               # Mock provider for testing
│   ├── provider.go
│   └── provider_test.go
├── knative/            # Knative provider (TODO)
│   └── provider.go
└── fission/            # Fission provider (TODO)
    └── provider.go
```

## Provider Interface

The `Provider` interface defines standard operations for serverless platforms:

- **Lifecycle Management**: Create, Get, Update, Delete, List functions
- **Version Management**: Create, Get, List versions, set active version
- **Invocation**: Invoke functions and retrieve logs
- **Triggers**: Create, Update, Delete, List triggers (HTTP, Schedule, Event)
- **Capabilities**: Query provider capabilities

## Implemented Providers

### Mock Provider (✓ Complete)
- Full implementation for testing
- 84.1% test coverage
- Thread-safe with configurable failures and delays

### Knative Provider (TODO)
- Will wrap existing Knative integration
- Container-based functions
- HTTP triggers via Knative Serving

### Fission Provider (TODO)
- Source-based functions
- Built-in runtime support
- Native cron triggers

## Usage

```go
// Create provider factory
factory := function.NewProviderFactory(kubeClient, dynamicClient)

// Create a provider instance
provider, err := factory.CreateProvider(ctx, function.ProviderTypeFission, map[string]interface{}{
    "endpoint": "http://fission-router.fission.svc.cluster.local",
})

// Use the provider
spec := &function.FunctionSpec{
    Namespace: "my-namespace",
    Name:      "hello-world",
    Runtime:   "python",
    Handler:   "main.handler",
    Source:    sourceCode,
}

fn, err := provider.CreateFunction(ctx, spec)
```

## Testing

Run tests:
```bash
# Unit tests
go test ./internal/repository/function/mock -v

# Contract tests (all providers must pass)
go test ./internal/domain/function -v

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Next Steps

1. Refactor application service to use provider interface
2. Implement Knative provider adapter
3. Implement Fission provider
4. Add integration tests
5. Update deployment configurations