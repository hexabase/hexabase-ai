# WebSocket API Reference

The Hexabase KaaS WebSocket API provides real-time updates for various events and operations.

## Connection

### WebSocket URL

```
Production: wss://api.hexabase.ai/ws
Staging: wss://api-staging.hexabase.ai/ws
Local: ws://api.localhost/ws
```

### Authentication

After establishing a WebSocket connection, you must authenticate by sending your JWT token:

```javascript
const ws = new WebSocket('wss://api.hexabase.ai/ws');

ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'your-jwt-token'
  }));
};
```

### Authentication Response

```json
{
  "type": "auth_result",
  "success": true,
  "user_id": "user-123",
  "message": "Authenticated successfully"
}
```

If authentication fails:

```json
{
  "type": "auth_result",
  "success": false,
  "error": "Invalid token"
}
```

## Message Format

All WebSocket messages follow this format:

### Client to Server

```json
{
  "type": "message_type",
  "id": "unique-message-id",
  "data": {
    // Message-specific data
  }
}
```

### Server to Client

```json
{
  "type": "message_type",
  "id": "unique-message-id",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    // Message-specific data
  }
}
```

## Subscriptions

### Subscribe to Events

You can subscribe to specific events or resources to receive real-time updates.

#### Subscribe to Workspace Events

```json
{
  "type": "subscribe",
  "id": "sub-123",
  "data": {
    "resource": "workspace",
    "workspace_id": "ws-123",
    "events": ["status_changed", "resource_updated", "alert_triggered"]
  }
}
```

#### Subscribe to Organization Events

```json
{
  "type": "subscribe",
  "id": "sub-124",
  "data": {
    "resource": "organization",
    "organization_id": "org-123",
    "events": ["member_added", "member_removed", "billing_updated"]
  }
}
```

#### Subscribe to Project Events

```json
{
  "type": "subscribe",
  "id": "sub-125",
  "data": {
    "resource": "project",
    "project_id": "proj-123",
    "events": ["created", "updated", "deleted", "resource_quota_exceeded"]
  }
}
```

### Subscription Response

```json
{
  "type": "subscribe_result",
  "id": "sub-123",
  "success": true,
  "subscription_id": "subscription-uuid"
}
```

### Unsubscribe

```json
{
  "type": "unsubscribe",
  "id": "unsub-123",
  "data": {
    "subscription_id": "subscription-uuid"
  }
}
```

## Event Types

### Workspace Events

#### Workspace Status Changed

```json
{
  "type": "workspace.status_changed",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "workspace_id": "ws-123",
    "previous_status": "provisioning",
    "new_status": "active",
    "message": "Workspace provisioning completed successfully"
  }
}
```

#### Workspace Resource Updated

```json
{
  "type": "workspace.resource_updated",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "workspace_id": "ws-123",
    "resources": {
      "cpu": {
        "used": "3.2",
        "limit": "10",
        "unit": "cores",
        "percentage": 32
      },
      "memory": {
        "used": "12288",
        "limit": "32768",
        "unit": "Mi",
        "percentage": 37.5
      }
    }
  }
}
```

#### Workspace Alert Triggered

```json
{
  "type": "workspace.alert_triggered",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "workspace_id": "ws-123",
    "alert": {
      "id": "alert-123",
      "type": "high_memory_usage",
      "severity": "warning",
      "title": "High Memory Usage",
      "description": "Memory usage has exceeded 80%",
      "resource": "pod/api-server-abc123",
      "threshold": 80,
      "value": 85.5
    }
  }
}
```

### Provisioning Events

#### Provisioning Progress

```json
{
  "type": "provisioning.progress",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "task_id": "task-123",
    "workspace_id": "ws-123",
    "stage": "creating_vcluster",
    "progress": 45,
    "message": "Creating vCluster instance..."
  }
}
```

#### Provisioning Completed

```json
{
  "type": "provisioning.completed",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "task_id": "task-123",
    "workspace_id": "ws-123",
    "duration_seconds": 120,
    "api_endpoint": "https://ws-123.api.hexabase.ai"
  }
}
```

#### Provisioning Failed

```json
{
  "type": "provisioning.failed",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "task_id": "task-123",
    "workspace_id": "ws-123",
    "error": "Failed to allocate resources",
    "stage": "resource_allocation",
    "can_retry": true
  }
}
```

### Project Events

#### Project Created

```json
{
  "type": "project.created",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "project": {
      "id": "proj-123",
      "name": "New Project",
      "namespace": "new-project",
      "workspace_id": "ws-123",
      "created_by": "user-123"
    }
  }
}
```

#### Project Resource Quota Exceeded

```json
{
  "type": "project.resource_quota_exceeded",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "project_id": "proj-123",
    "namespace": "frontend",
    "resource": "memory",
    "requested": "5Gi",
    "limit": "4Gi",
    "current_usage": "3.8Gi"
  }
}
```

### Organization Events

#### Member Added

```json
{
  "type": "organization.member_added",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "organization_id": "org-123",
    "member": {
      "user_id": "user-456",
      "email": "newuser@example.com",
      "role": "member",
      "added_by": "user-123"
    }
  }
}
```

#### Billing Updated

```json
{
  "type": "organization.billing_updated",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "organization_id": "org-123",
    "billing": {
      "previous_plan": "standard",
      "new_plan": "pro",
      "effective_date": "2024-01-20T10:00:00Z"
    }
  }
}
```

### Monitoring Events

#### Metrics Update

```json
{
  "type": "monitoring.metrics_update",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "workspace_id": "ws-123",
    "metrics": {
      "cpu": {
        "instant": 2.5,
        "average_5m": 2.3,
        "average_1h": 2.1
      },
      "memory": {
        "instant": 8192,
        "average_5m": 8000,
        "average_1h": 7800
      },
      "network": {
        "ingress_rate": "1.2MB/s",
        "egress_rate": "0.8MB/s"
      }
    }
  }
}
```

#### Health Status Change

```json
{
  "type": "monitoring.health_changed",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "workspace_id": "ws-123",
    "component": "api_server",
    "previous_status": "healthy",
    "new_status": "unhealthy",
    "reason": "API server not responding to health checks"
  }
}
```

## Commands

### Get Workspace Status

Request current status of a workspace:

```json
{
  "type": "get_workspace_status",
  "id": "cmd-123",
  "data": {
    "workspace_id": "ws-123"
  }
}
```

Response:

```json
{
  "type": "workspace_status",
  "id": "cmd-123",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "workspace_id": "ws-123",
    "status": "active",
    "health": "healthy",
    "nodes": {
      "ready": 3,
      "total": 3
    },
    "pods": {
      "running": 45,
      "pending": 2,
      "failed": 0
    }
  }
}
```

### Get Real-time Logs

Stream logs from a specific resource:

```json
{
  "type": "stream_logs",
  "id": "cmd-124",
  "data": {
    "workspace_id": "ws-123",
    "namespace": "frontend",
    "pod": "api-server-abc123",
    "container": "api",
    "follow": true,
    "tail_lines": 100
  }
}
```

Log entries will be streamed as:

```json
{
  "type": "log_entry",
  "timestamp": "2024-01-20T10:00:00Z",
  "data": {
    "stream_id": "cmd-124",
    "timestamp": "2024-01-20T10:00:00Z",
    "level": "info",
    "message": "Request processed successfully",
    "pod": "api-server-abc123",
    "container": "api"
  }
}
```

Stop streaming:

```json
{
  "type": "stop_stream",
  "id": "cmd-125",
  "data": {
    "stream_id": "cmd-124"
  }
}
```

## Connection Management

### Ping/Pong

The server sends ping messages every 30 seconds to keep the connection alive:

```json
{
  "type": "ping",
  "timestamp": "2024-01-20T10:00:00Z"
}
```

Clients should respond with:

```json
{
  "type": "pong"
}
```

### Reconnection

If the connection is lost, clients should:

1. Implement exponential backoff for reconnection attempts
2. Re-authenticate after reconnecting
3. Re-subscribe to previously subscribed events

### Connection Limits

- Maximum message size: 1MB
- Maximum subscriptions per connection: 100
- Idle timeout: 5 minutes (kept alive by ping/pong)

## Error Handling

### Error Message Format

```json
{
  "type": "error",
  "id": "message-id-if-applicable",
  "timestamp": "2024-01-20T10:00:00Z",
  "error": {
    "code": "SUBSCRIPTION_LIMIT_EXCEEDED",
    "message": "Maximum number of subscriptions (100) exceeded",
    "details": {
      "current_subscriptions": 100,
      "requested_subscription": "workspace:ws-456"
    }
  }
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `AUTH_REQUIRED` | Authentication required before subscribing |
| `INVALID_TOKEN` | JWT token is invalid or expired |
| `PERMISSION_DENIED` | User lacks permission for requested resource |
| `RESOURCE_NOT_FOUND` | Requested resource does not exist |
| `SUBSCRIPTION_LIMIT_EXCEEDED` | Too many active subscriptions |
| `INVALID_MESSAGE_FORMAT` | Message format is invalid |
| `RATE_LIMIT_EXCEEDED` | Too many messages sent |
| `MESSAGE_TOO_LARGE` | Message exceeds size limit |

## Client Libraries

### JavaScript/TypeScript Example

```typescript
class HexabaseWebSocket {
  private ws: WebSocket;
  private subscriptions: Map<string, string> = new Map();
  private messageHandlers: Map<string, Function> = new Map();

  constructor(private token: string) {
    this.connect();
  }

  private connect() {
    this.ws = new WebSocket('wss://api.hexabase.ai/ws');
    
    this.ws.onopen = () => {
      this.authenticate();
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      // Implement reconnection logic
      setTimeout(() => this.connect(), 5000);
    };
  }

  private authenticate() {
    this.send({
      type: 'auth',
      token: this.token
    });
  }

  private handleMessage(message: any) {
    switch (message.type) {
      case 'auth_result':
        if (message.success) {
          this.resubscribe();
        }
        break;
      case 'ping':
        this.send({ type: 'pong' });
        break;
      default:
        const handler = this.messageHandlers.get(message.type);
        if (handler) {
          handler(message.data);
        }
    }
  }

  subscribe(resource: string, id: string, events: string[]): Promise<string> {
    return new Promise((resolve, reject) => {
      const messageId = `sub-${Date.now()}`;
      
      this.send({
        type: 'subscribe',
        id: messageId,
        data: {
          resource,
          [`${resource}_id`]: id,
          events
        }
      });

      // Handle subscription response
      const handler = (message: any) => {
        if (message.id === messageId) {
          if (message.success) {
            this.subscriptions.set(message.subscription_id, messageId);
            resolve(message.subscription_id);
          } else {
            reject(new Error(message.error));
          }
        }
      };

      this.messageHandlers.set('subscribe_result', handler);
    });
  }

  on(event: string, handler: Function) {
    this.messageHandlers.set(event, handler);
  }

  private send(message: any) {
    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  private resubscribe() {
    // Re-subscribe to all previous subscriptions after reconnection
    this.subscriptions.forEach((messageId, subscriptionId) => {
      // Implement resubscription logic
    });
  }

  close() {
    this.ws.close();
  }
}

// Usage
const ws = new HexabaseWebSocket('your-jwt-token');

// Subscribe to workspace events
ws.subscribe('workspace', 'ws-123', ['status_changed', 'alert_triggered'])
  .then(subscriptionId => {
    console.log('Subscribed:', subscriptionId);
  });

// Handle events
ws.on('workspace.status_changed', (data) => {
  console.log('Workspace status changed:', data);
});

ws.on('workspace.alert_triggered', (data) => {
  console.log('Alert triggered:', data);
});
```

## Best Practices

1. **Authentication**: Always authenticate immediately after connecting
2. **Error Handling**: Implement proper error handling for all message types
3. **Reconnection**: Implement automatic reconnection with exponential backoff
4. **Subscriptions**: Track subscriptions for re-subscribing after reconnection
5. **Message IDs**: Use unique message IDs for request/response correlation
6. **Rate Limiting**: Implement client-side rate limiting to avoid exceeding limits
7. **Cleanup**: Unsubscribe from events when no longer needed
8. **Heartbeat**: Respond to ping messages to keep connection alive