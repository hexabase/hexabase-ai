# Function Service API Reference

## Overview

The Function Service provides a complete serverless function management system with support for multiple FaaS providers (Fission and Knative). This API allows you to create, deploy, version, and invoke functions with automatic scaling and high availability.

## Base URL

```
https://api.hexabase.ai/api/v1/workspaces/{workspaceId}/functions
```

## Authentication

All endpoints require authentication via Bearer token:

```http
Authorization: Bearer <your-token>
```

## Function Management

### Create Function

Creates a new serverless function in the specified project.

**POST** `/functions`

#### Request Body

```json
{
  "name": "my-function",
  "runtime": "python3.9",
  "handler": "main.handler",
  "source_code": "ZGVmIGhhbmRsZXIoY29udGV4dCk6CiAgICByZXR1cm4geyJzdGF0dXMiOiAyMDAsICJib2R5IjogIkhlbGxvIFdvcmxkIn0=",
  "environment": {
    "API_KEY": "secret-key",
    "POOL_SIZE": "3"
  },
  "resources": {
    "memory": "256Mi",
    "cpu": "100m"
  },
  "labels": {
    "team": "backend",
    "env": "production"
  }
}
```

#### Response

```json
{
  "id": "func-123456",
  "workspace_id": "ws-789",
  "project_id": "proj-456",
  "name": "my-function",
  "namespace": "proj-456",
  "runtime": "python3.9",
  "handler": "main.handler",
  "status": "ready",
  "active_version": "v1",
  "created_at": "2025-06-10T18:00:00Z",
  "updated_at": "2025-06-10T18:00:00Z"
}
```

### List Functions

Retrieves all functions in a project.

**GET** `/functions?project_id={projectId}`

#### Query Parameters

- `project_id` (required): The project ID to list functions for

#### Response

```json
{
  "functions": [
    {
      "id": "func-123456",
      "name": "my-function",
      "runtime": "python3.9",
      "status": "ready",
      "active_version": "v1",
      "created_at": "2025-06-10T18:00:00Z"
    }
  ]
}
```

### Get Function

Retrieves details of a specific function.

**GET** `/functions/{functionId}`

#### Response

```json
{
  "id": "func-123456",
  "workspace_id": "ws-789",
  "project_id": "proj-456",
  "name": "my-function",
  "namespace": "proj-456",
  "runtime": "python3.9",
  "handler": "main.handler",
  "status": "ready",
  "active_version": "v1",
  "labels": {
    "team": "backend",
    "env": "production"
  },
  "annotations": {
    "description": "Processes user data"
  },
  "created_at": "2025-06-10T18:00:00Z",
  "updated_at": "2025-06-10T18:00:00Z"
}
```

### Update Function

Updates an existing function's configuration.

**PUT** `/functions/{functionId}`

#### Request Body

```json
{
  "name": "updated-function",
  "handler": "main.new_handler",
  "environment": {
    "NEW_VAR": "value"
  },
  "resources": {
    "memory": "512Mi",
    "cpu": "200m"
  }
}
```

### Delete Function

Deletes a function and all its associated resources.

**DELETE** `/functions/{functionId}`

#### Response

```http
204 No Content
```

## Version Management

### Deploy Version

Creates and deploys a new version of a function.

**POST** `/functions/{functionId}/versions`

#### Request Body

```json
{
  "version": 2,
  "source_code": "ZGVmIGhhbmRsZXIoY29udGV4dCk6CiAgICByZXR1cm4geyJzdGF0dXMiOiAyMDAsICJib2R5IjogIlZlcnNpb24gMiJ9",
  "image": "myregistry/my-function:v2"
}
```

#### Response

```json
{
  "id": "ver-789",
  "workspace_id": "ws-789",
  "function_id": "func-123456",
  "function_name": "my-function",
  "version": 2,
  "build_status": "building",
  "created_at": "2025-06-10T18:05:00Z",
  "is_active": false
}
```

### List Versions

Retrieves all versions of a function.

**GET** `/functions/{functionId}/versions`

#### Response

```json
{
  "versions": [
    {
      "id": "ver-789",
      "version": 2,
      "build_status": "success",
      "created_at": "2025-06-10T18:05:00Z",
      "is_active": true
    },
    {
      "id": "ver-456",
      "version": 1,
      "build_status": "success",
      "created_at": "2025-06-10T18:00:00Z",
      "is_active": false
    }
  ]
}
```

### Set Active Version

Activates a specific version of a function.

**PUT** `/functions/{functionId}/versions/{versionId}/active`

#### Response

```json
{
  "message": "Version activated successfully"
}
```

### Rollback Version

Rolls back to the previous version.

**POST** `/functions/{functionId}/rollback`

#### Response

```json
{
  "message": "Rollback successful"
}
```

## Trigger Management

### Create Trigger

Creates a new trigger for a function.

**POST** `/functions/{functionId}/triggers`

#### Request Body (HTTP Trigger)

```json
{
  "name": "http-trigger",
  "type": "http",
  "enabled": true,
  "config": {
    "method": "GET",
    "path": "/api/hello"
  }
}
```

#### Request Body (Schedule Trigger)

```json
{
  "name": "cron-trigger",
  "type": "schedule",
  "enabled": true,
  "config": {
    "cron": "0 */5 * * *"
  }
}
```

#### Response

```json
{
  "id": "trg-123",
  "workspace_id": "ws-789",
  "function_id": "func-123456",
  "name": "http-trigger",
  "type": "http",
  "enabled": true,
  "config": {
    "method": "GET",
    "path": "/api/hello"
  },
  "created_at": "2025-06-10T18:10:00Z"
}
```

### List Triggers

Retrieves all triggers for a function.

**GET** `/functions/{functionId}/triggers`

### Update Trigger

Updates an existing trigger.

**PUT** `/functions/{functionId}/triggers/{triggerId}`

### Delete Trigger

Removes a trigger from a function.

**DELETE** `/functions/{functionId}/triggers/{triggerId}`

## Function Invocation

### Invoke Function (Synchronous)

Invokes a function and waits for the response.

**POST** `/functions/{functionId}/invoke`

#### Request Body

```json
{
  "method": "POST",
  "path": "/process",
  "headers": {
    "Content-Type": ["application/json"],
    "X-Custom-Header": ["value"]
  },
  "body": "eyJkYXRhIjogInRlc3QifQ==",
  "query": {
    "param1": ["value1"],
    "param2": ["value2"]
  }
}
```

#### Response

```json
{
  "status_code": 200,
  "headers": {
    "Content-Type": ["application/json"]
  },
  "body": "eyJyZXN1bHQiOiAic3VjY2VzcyJ9",
  "duration": 125,
  "cold_start": false,
  "invocation_id": "inv-456789"
}
```

### Invoke Function (Asynchronous)

Invokes a function without waiting for completion.

**POST** `/functions/{functionId}/invoke-async`

#### Request Body

Same as synchronous invocation.

#### Response

```json
{
  "invocation_id": "inv-456789"
}
```

### Get Invocation Status

Retrieves the status of an asynchronous invocation.

**GET** `/functions/invocations/{invocationId}`

#### Response

```json
{
  "invocation_id": "inv-456789",
  "workspace_id": "ws-789",
  "function_id": "func-123456",
  "status": "completed",
  "started_at": "2025-06-10T18:15:00Z",
  "completed_at": "2025-06-10T18:15:00.125Z",
  "result": {
    "status_code": 200,
    "body": "eyJyZXN1bHQiOiAic3VjY2VzcyJ9"
  }
}
```

### List Invocations

Retrieves invocation history for a function.

**GET** `/functions/{functionId}/invocations?limit=50`

## Monitoring

### Get Function Logs

Retrieves logs for a function.

**GET** `/functions/{functionId}/logs`

#### Query Parameters

- `since`: RFC3339 timestamp to start from
- `until`: RFC3339 timestamp to end at
- `limit`: Maximum number of log entries (default: 100)
- `follow`: Stream logs in real-time (boolean)
- `previous`: Show logs from previous version

#### Response

```json
{
  "logs": [
    {
      "timestamp": "2025-06-10T18:15:00.123Z",
      "level": "info",
      "message": "Processing request",
      "container": "my-function-v2-abc123",
      "pod": "my-function-v2-abc123-xyz789"
    }
  ]
}
```

### Get Function Metrics

Retrieves performance metrics for a function.

**GET** `/functions/{functionId}/metrics`

#### Query Parameters

- `start`: Start time (RFC3339)
- `end`: End time (RFC3339)
- `resolution`: Time resolution (1m, 5m, 1h)
- `metrics[]`: Specific metrics to retrieve

#### Response

```json
{
  "invocations": 1523,
  "errors": 12,
  "duration": {
    "min": 45,
    "max": 2100,
    "avg": 125,
    "p50": 95,
    "p95": 180,
    "p99": 450
  },
  "cold_starts": 23,
  "concurrency": {
    "min": 0,
    "max": 15,
    "avg": 3.2
  }
}
```

### Get Function Events

Retrieves audit events for a function.

**GET** `/functions/{functionId}/events?limit=100`

#### Response

```json
{
  "events": [
    {
      "id": "evt-123",
      "type": "deployed",
      "description": "Version v2 deployed",
      "metadata": {
        "version": "v2",
        "deployed_by": "user-123"
      },
      "created_at": "2025-06-10T18:05:00Z"
    }
  ]
}
```

## Provider Management

### Get Provider Capabilities

Retrieves the capabilities of the workspace's FaaS provider.

**GET** `/functions/provider/capabilities`

#### Response

```json
{
  "name": "fission",
  "version": "1.18.0",
  "description": "Fission lightweight serverless platform",
  "supports_versioning": true,
  "supported_runtimes": ["python3.9", "nodejs18", "go1.21", "java11"],
  "supported_trigger_types": ["http", "schedule", "event", "messagequeue"],
  "supports_async": true,
  "supports_logs": true,
  "supports_metrics": true,
  "supports_environment_vars": true,
  "supports_custom_images": true,
  "max_memory_mb": 4096,
  "max_timeout_secs": 300,
  "max_payload_size_mb": 50,
  "typical_cold_start_ms": 100,
  "supports_scale_to_zero": true,
  "supports_auto_scaling": true,
  "supports_https": true,
  "supports_warm_pool": true
}
```

### Check Provider Health

Verifies the health of the FaaS provider.

**GET** `/functions/provider/health`

#### Response (Healthy)

```json
{
  "status": "healthy"
}
```

#### Response (Unhealthy)

```json
{
  "error": "Provider unhealthy",
  "details": "Controller endpoint unreachable: connection timeout"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context"
  }
}
```

### Common Error Codes

- `FUNCTION_NOT_FOUND`: Function does not exist
- `VERSION_NOT_FOUND`: Version does not exist
- `TRIGGER_NOT_FOUND`: Trigger does not exist
- `INVALID_RUNTIME`: Runtime not supported
- `BUILD_FAILED`: Function build failed
- `INVOCATION_FAILED`: Function invocation failed
- `PROVIDER_ERROR`: FaaS provider error
- `QUOTA_EXCEEDED`: Resource quota exceeded

## Rate Limits

- Function creation: 100 per hour per workspace
- Function invocation: 10,000 per minute per function
- Log retrieval: 1,000 requests per hour per workspace

## Webhooks

Configure webhooks to receive notifications about function events:

```json
{
  "url": "https://your-domain.com/webhooks/functions",
  "events": ["function.created", "function.deployed", "function.failed"],
  "secret": "webhook-secret"
}
```

## SDK Examples

### Python

```python
from hexabase import FunctionClient

client = FunctionClient(workspace_id="ws-789", api_key="your-key")

# Create function
function = client.create_function(
    name="my-function",
    runtime="python3.9",
    handler="main.handler",
    source_code="""
def handler(context):
    return {"status": 200, "body": "Hello World"}
"""
)

# Invoke function
result = client.invoke_function(
    function_id=function.id,
    data={"message": "test"}
)
print(result.body)
```

### JavaScript

```javascript
const { FunctionClient } = require('@hexabase/sdk');

const client = new FunctionClient({
  workspaceId: 'ws-789',
  apiKey: 'your-key'
});

// Create function
const fn = await client.createFunction({
  name: 'my-function',
  runtime: 'nodejs18',
  handler: 'index.handler',
  sourceCode: `
exports.handler = async (event) => {
  return {
    statusCode: 200,
    body: JSON.stringify({ message: 'Hello World' })
  };
};
`
});

// Invoke function
const result = await client.invokeFunction(fn.id, {
  body: { message: 'test' }
});
console.log(result.body);
```

## Best Practices

1. **Use versioning** for production functions
2. **Configure warm pools** for latency-sensitive functions
3. **Set appropriate resource limits** to prevent overuse
4. **Monitor metrics** regularly
5. **Use asynchronous invocation** for long-running tasks
6. **Implement proper error handling** in your functions
7. **Use environment variables** for configuration
8. **Enable logging** for debugging