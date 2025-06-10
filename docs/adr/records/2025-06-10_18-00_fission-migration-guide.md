# Migration Guide: Transitioning to Fission as Default FaaS Provider

**Date**: 2025-06-10  
**Status**: Implemented  
**Authors**: Hexabase AI Team

## Executive Summary

This guide provides step-by-step instructions for migrating existing workspaces from Knative to Fission as the Function-as-a-Service (FaaS) provider. Fission offers significant performance improvements with 50-200ms cold starts compared to Knative's 2-5s, making it ideal for latency-sensitive applications.

## Prerequisites

- Hexabase AI platform v2.0 or later
- Admin access to workspaces being migrated
- Fission v1.18.0+ installed in your cluster
- Backup of existing function configurations

## Benefits of Migration

### Performance Improvements
- **Cold Start**: 95% reduction (from 2-5s to 50-200ms)
- **Memory Usage**: 30-50% reduction in idle memory
- **Warm Pool**: Pre-warmed instances eliminate cold starts entirely

### Feature Enhancements
- Native time-based triggers (cron schedules)
- Built-in function versioning with atomic deployments
- Message queue triggers for event-driven architectures
- Simplified debugging with integrated logging

## Migration Strategies

### 1. New Workspaces (Default)

All new workspaces automatically use Fission as the default provider:

```yaml
# Default configuration for new workspaces
provider:
  type: fission
  config:
    endpoint: http://controller.fission.svc.cluster.local
    namespace: fission-function
```

### 2. Gradual Migration (Recommended)

Migrate functions one at a time with minimal downtime:

```bash
# Step 1: Enable Fission for workspace
hexabase workspace configure ws-123 --provider fission

# Step 2: Migrate individual functions
hexabase function migrate func-456 --from knative --to fission --preserve-triggers

# Step 3: Verify functionality
hexabase function test func-456 --compare-providers
```

### 3. Bulk Migration

For workspaces with many functions:

```bash
# Export all functions
hexabase function export --workspace ws-123 --format yaml > functions.yaml

# Update provider configuration
hexabase workspace configure ws-123 --provider fission

# Import with new provider
hexabase function import --workspace ws-123 --file functions.yaml --provider fission
```

## Step-by-Step Migration Process

### Phase 1: Assessment

1. **Inventory Functions**
   ```bash
   hexabase function list --workspace ws-123 --provider knative
   ```

2. **Check Compatibility**
   ```bash
   hexabase function analyze --workspace ws-123 --target-provider fission
   ```

3. **Identify Dependencies**
   - Custom runtimes
   - Specific Knative features
   - External integrations

### Phase 2: Preparation

1. **Backup Current State**
   ```bash
   hexabase backup create --workspace ws-123 --type functions
   ```

2. **Update Function Code** (if needed)
   ```python
   # Knative function
   def handler(request):
       return {"body": "Hello World"}
   
   # Fission function (compatible format)
   def main():
       return "Hello World"
   ```

3. **Configure Workspace**
   ```bash
   hexabase workspace configure ws-123 \
     --provider fission \
     --fission-endpoint http://controller.fission.svc.cluster.local
   ```

### Phase 3: Migration

1. **Create Function in Fission**
   ```bash
   hexabase function create \
     --workspace ws-123 \
     --name my-function \
     --runtime python \
     --handler main \
     --source function.py
   ```

2. **Configure Triggers**
   ```bash
   # HTTP trigger
   hexabase trigger create http \
     --function my-function \
     --method GET \
     --path /api/hello
   
   # Time trigger (Fission-specific)
   hexabase trigger create time \
     --function my-function \
     --cron "0 */5 * * *"
   ```

3. **Test Function**
   ```bash
   hexabase function invoke my-function --data '{"test": true}'
   ```

### Phase 4: Cutover

1. **Update DNS/Routes**
   ```bash
   hexabase route update \
     --from knative.my-function \
     --to fission.my-function
   ```

2. **Monitor Performance**
   ```bash
   hexabase function metrics my-function --duration 1h
   ```

3. **Cleanup Old Resources**
   ```bash
   hexabase function delete my-function --provider knative
   ```

## API Changes

### Function Creation

**Before (Knative)**:
```go
spec := &function.FunctionSpec{
    Name:       "my-func",
    Runtime:    function.RuntimePython,
    Handler:    "handler", 
    SourceCode: sourceCode,
}
```

**After (Fission)**:
```go
spec := &function.FunctionSpec{
    Name:       "my-func",
    Runtime:    function.RuntimePython,
    Handler:    "main",  // Fission uses 'main' by convention
    SourceCode: sourceCode,
    Environment: map[string]string{
        "POOL_SIZE": "3",  // Warm pool configuration
    },
}
```

### Trigger Configuration

**HTTP Triggers** (Compatible):
```go
trigger := &function.FunctionTrigger{
    Type: function.TriggerHTTP,
    Config: map[string]string{
        "method": "GET",
        "path":   "/api/function",
    },
}
```

**Time Triggers** (New in Fission):
```go
trigger := &function.FunctionTrigger{
    Type: function.TriggerSchedule,
    Config: map[string]string{
        "cron": "*/5 * * * *",  // Every 5 minutes
    },
}
```

## Performance Tuning

### Warm Pool Configuration

Configure pre-warmed instances for zero cold start:

```yaml
# Per-function configuration
environment:
  POOL_SIZE: "3"           # Number of warm instances
  POOL_TYPE: "poolmgr"     # Use pool manager
  MIN_CPU: "100m"          # Minimum CPU per instance
  MIN_MEMORY: "128Mi"      # Minimum memory per instance
```

### Resource Optimization

```yaml
# Fission-optimized settings
resources:
  memory: "128Mi"  # Lower than Knative due to less overhead
  cpu: "100m"      # Efficient CPU usage
  
# Scaling configuration  
scaling:
  min_replicas: 1
  max_replicas: 100
  target_cpu: 50
```

## Monitoring and Troubleshooting

### Health Checks

```bash
# Check provider status
hexabase provider health --workspace ws-123

# Verify function deployment
hexabase function status my-function --detailed
```

### Common Issues

1. **Function Not Found**
   ```bash
   # Ensure namespace is correct
   hexabase function list --namespace fission-function
   ```

2. **Trigger Not Firing**
   ```bash
   # Check trigger configuration
   hexabase trigger describe my-trigger
   
   # View trigger logs
   hexabase logs --trigger my-trigger --tail 100
   ```

3. **Performance Issues**
   ```bash
   # Analyze cold starts
   hexabase function analyze-performance my-function
   
   # Adjust warm pool
   hexabase function configure my-function --pool-size 5
   ```

## Rollback Procedure

If issues arise during migration:

1. **Immediate Rollback**
   ```bash
   hexabase workspace configure ws-123 --provider knative
   ```

2. **Restore from Backup**
   ```bash
   hexabase backup restore --workspace ws-123 --backup-id bkp-789
   ```

3. **Verify Functionality**
   ```bash
   hexabase function test --all --workspace ws-123
   ```

## Best Practices

### 1. Test Thoroughly
- Run parallel deployments during transition
- Use feature flags for gradual rollout
- Monitor metrics closely

### 2. Optimize for Fission
- Use poolmgr executor for consistent performance
- Configure appropriate warm pool sizes
- Leverage Fission-specific features

### 3. Plan for Differences
- Update monitoring dashboards
- Retrain operations team
- Document provider-specific behaviors

## Support and Resources

- **Documentation**: https://docs.hexabase.ai/functions/fission
- **Migration Support**: support@hexabase.ai
- **Community Forum**: https://community.hexabase.ai/c/functions
- **Fission Documentation**: https://docs.fission.io

## Conclusion

Migrating to Fission provides significant performance benefits with minimal code changes. The improved cold start times and resource efficiency make it ideal for production workloads. Follow this guide carefully and leverage the provided tools for a smooth transition.