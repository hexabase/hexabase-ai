# Authentication

Hexabase KaaS uses OAuth2/OIDC for authentication with JWT tokens for API access.

## Overview

The authentication flow follows these steps:

1. User initiates login with an OAuth provider
2. User is redirected to the provider's login page
3. After successful authentication, user is redirected back with an authorization code
4. The code is exchanged for Hexabase KaaS JWT tokens
5. JWT tokens are used for all subsequent API requests

## Supported Providers

### Google

- **Provider ID**: `google`
- **Required Scopes**: `openid`, `email`, `profile`
- **Configuration**: OAuth 2.0 client credentials required

### GitHub

- **Provider ID**: `github`
- **Required Scopes**: `user:email`, `read:org`
- **Configuration**: OAuth App credentials required

### Microsoft Azure AD

- **Provider ID**: `azure`
- **Required Scopes**: `openid`, `email`, `profile`
- **Configuration**: App registration in Azure AD required

### Custom OIDC Provider

- **Provider ID**: Custom identifier
- **Required Claims**: `sub`, `email`, `name`
- **Configuration**: OIDC discovery endpoint required

## Authentication Flows

### Authorization Code Flow (Web Applications)

This is the recommended flow for web applications.

#### 1. Initiate Login

Redirect user to:
```
https://api.hexabase.ai/auth/login/google?redirect_uri=https://app.hexabase.ai/auth/callback
```

#### 2. Handle Callback

After authentication, the user is redirected to:
```
https://app.hexabase.ai/auth/callback?code=AUTHORIZATION_CODE&state=STATE
```

#### 3. Exchange Code for Tokens

```http
POST /auth/callback/google
Content-Type: application/json

{
  "code": "AUTHORIZATION_CODE",
  "redirect_uri": "https://app.hexabase.ai/auth/callback"
}
```

Response:
```json
{
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "id": "user-123",
      "email": "user@example.com",
      "name": "John Doe",
      "picture": "https://..."
    }
  }
}
```

### PKCE Flow (Single Page Applications)

For enhanced security in SPAs, use the PKCE (Proof Key for Code Exchange) flow.

#### 1. Generate Code Verifier and Challenge

```javascript
// Generate code verifier
const codeVerifier = generateRandomString(128);

// Generate code challenge
const codeChallenge = await sha256(codeVerifier);
const codeChallengeBase64 = base64UrlEncode(codeChallenge);
```

#### 2. Initiate Login with PKCE

```
https://api.hexabase.ai/auth/login/google?
  redirect_uri=https://app.hexabase.ai/auth/callback&
  code_challenge=CODE_CHALLENGE&
  code_challenge_method=S256
```

#### 3. Exchange Code with Verifier

```http
POST /auth/callback/google
Content-Type: application/json

{
  "code": "AUTHORIZATION_CODE",
  "redirect_uri": "https://app.hexabase.ai/auth/callback",
  "code_verifier": "CODE_VERIFIER"
}
```

### Direct Token Exchange (Mobile/Desktop Apps)

For native applications that can securely handle OAuth flows.

```http
POST /auth/login/google
Content-Type: application/json

{
  "id_token": "GOOGLE_ID_TOKEN"
}
```

## JWT Tokens

### Token Structure

Hexabase KaaS issues JWT tokens with the following structure:

#### Access Token Claims

```json
{
  "sub": "user-123",
  "email": "user@example.com",
  "name": "John Doe",
  "picture": "https://...",
  "iss": "https://api.hexabase.ai",
  "aud": "hexabase-ai",
  "exp": 1705753200,
  "iat": 1705749600,
  "jti": "token-unique-id",
  "organizations": [
    {
      "id": "org-123",
      "role": "admin"
    }
  ],
  "fingerprint": "device-fingerprint-hash"
}
```

#### Refresh Token Claims

```json
{
  "sub": "user-123",
  "type": "refresh",
  "iss": "https://api.hexabase.ai",
  "aud": "hexabase-ai",
  "exp": 1708341600,
  "iat": 1705749600,
  "jti": "refresh-token-id",
  "family": "token-family-id"
}
```

### Token Lifetimes

- **Access Token**: 1 hour (3600 seconds)
- **Refresh Token**: 30 days
- **Session Maximum**: 90 days

### Token Usage

Include the access token in the Authorization header:

```http
GET /api/v1/organizations
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
```

### Token Refresh

When the access token expires, use the refresh token to get a new one:

```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

Response:
```json
{
  "data": {
    "access_token": "new-access-token",
    "refresh_token": "new-refresh-token",
    "expires_in": 3600
  }
}
```

## Security Features

### JWT Fingerprinting

To prevent token theft, Hexabase KaaS implements JWT fingerprinting:

1. A device fingerprint is generated during login
2. The fingerprint hash is embedded in the JWT
3. Each request validates the fingerprint matches
4. Tokens are bound to the originating device/browser

### Refresh Token Rotation

For enhanced security, refresh tokens are rotated on each use:

1. Each refresh token can only be used once
2. Using a refresh token invalidates it and issues a new one
3. Reuse of an old refresh token invalidates the entire token family
4. This prevents refresh token replay attacks

### Token Revocation

Tokens can be revoked in several ways:

#### Logout

```http
POST /auth/logout
Authorization: Bearer <access-token>
```

This revokes the current session and associated tokens.

#### Revoke All Sessions

```http
POST /auth/sessions/revoke-all
Authorization: Bearer <access-token>
```

This revokes all active sessions for the user.

#### Revoke Specific Session

```http
DELETE /auth/sessions/:sessionId
Authorization: Bearer <access-token>
```

### Rate Limiting

Authentication endpoints have specific rate limits:

- **Login attempts**: 5 per minute per IP
- **Token refresh**: 10 per minute per user
- **Failed attempts**: Exponential backoff after 3 failures

## Session Management

### List Active Sessions

```http
GET /auth/sessions
Authorization: Bearer <access-token>
```

Response:
```json
{
  "data": [
    {
      "id": "session-123",
      "device": "Chrome on Mac OS",
      "ip_address": "192.168.1.1",
      "location": "San Francisco, CA",
      "created_at": "2024-01-20T10:00:00Z",
      "last_active": "2024-01-20T15:30:00Z",
      "is_current": true
    }
  ]
}
```

### Security Events

View security-related events for your account:

```http
GET /auth/security-logs
Authorization: Bearer <access-token>
```

Response:
```json
{
  "data": [
    {
      "id": "evt-123",
      "type": "login_success",
      "timestamp": "2024-01-20T10:00:00Z",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "location": "San Francisco, CA",
      "provider": "google"
    },
    {
      "id": "evt-124",
      "type": "login_failed",
      "timestamp": "2024-01-20T09:55:00Z",
      "ip_address": "192.168.1.1",
      "reason": "invalid_credentials"
    }
  ]
}
```

## OIDC Discovery

Hexabase KaaS provides OIDC discovery endpoints for integration:

### Discovery Document

```http
GET /.well-known/openid-configuration
```

Response:
```json
{
  "issuer": "https://api.hexabase.ai",
  "authorization_endpoint": "https://api.hexabase.ai/auth/authorize",
  "token_endpoint": "https://api.hexabase.ai/auth/token",
  "userinfo_endpoint": "https://api.hexabase.ai/auth/userinfo",
  "jwks_uri": "https://api.hexabase.ai/.well-known/jwks.json",
  "response_types_supported": ["code", "token", "id_token"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid", "email", "profile"],
  "token_endpoint_auth_methods_supported": ["client_secret_post", "client_secret_basic"],
  "claims_supported": ["sub", "email", "name", "picture", "organizations"]
}
```

### JWKS Endpoint

```http
GET /.well-known/jwks.json
```

Response:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "key-1",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

## Multi-Factor Authentication (MFA)

### Enable MFA

```http
POST /auth/mfa/enable
Authorization: Bearer <access-token>
```

Response includes QR code for authenticator app setup.

### Verify MFA

```http
POST /auth/mfa/verify
Content-Type: application/json
Authorization: Bearer <access-token>

{
  "code": "123456"
}
```

### Login with MFA

After initial authentication, if MFA is enabled:

```http
POST /auth/mfa/challenge
Content-Type: application/json

{
  "session_token": "mfa-session-token",
  "code": "123456"
}
```

## API Keys (Service Accounts)

For automated systems and CI/CD pipelines:

### Create API Key

```http
POST /api/v1/organizations/:orgId/api-keys
Content-Type: application/json
Authorization: Bearer <access-token>

{
  "name": "CI/CD Pipeline",
  "scopes": ["workspaces:read", "workspaces:write"],
  "expires_at": "2025-01-01T00:00:00Z"
}
```

Response:
```json
{
  "data": {
    "id": "key-123",
    "name": "CI/CD Pipeline",
    "key": "hxb_live_1234567890abcdef",
    "created_at": "2024-01-20T10:00:00Z",
    "expires_at": "2025-01-01T00:00:00Z"
  }
}
```

### Use API Key

```http
GET /api/v1/workspaces
Authorization: Bearer hxb_live_1234567890abcdef
```

## Best Practices

1. **Token Storage**
   - Store tokens securely (httpOnly cookies or secure storage)
   - Never store tokens in localStorage for production
   - Clear tokens on logout

2. **Token Refresh**
   - Implement automatic token refresh before expiration
   - Handle refresh failures gracefully
   - Don't expose refresh tokens to client-side JavaScript

3. **PKCE Usage**
   - Always use PKCE for public clients (SPAs, mobile apps)
   - Generate cryptographically secure code verifiers
   - Never reuse code verifiers

4. **Session Security**
   - Regularly review active sessions
   - Revoke sessions from unknown devices
   - Enable MFA for sensitive accounts

5. **API Key Management**
   - Rotate API keys regularly
   - Use minimal required scopes
   - Monitor API key usage

## Error Codes

| Code | Description |
|------|-------------|
| `invalid_request` | Request is missing required parameters |
| `invalid_client` | Client authentication failed |
| `invalid_grant` | Authorization code or refresh token is invalid |
| `unauthorized_client` | Client is not authorized for this grant type |
| `unsupported_grant_type` | Grant type is not supported |
| `invalid_scope` | Requested scope is invalid or exceeds granted scope |
| `token_expired` | Access token has expired |
| `token_revoked` | Token has been revoked |
| `mfa_required` | Multi-factor authentication is required |
| `rate_limit_exceeded` | Too many authentication attempts |# WebSocket API Reference

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
8. **Heartbeat**: Respond to ping messages to keep connection alive# Function Service API Reference

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
8. **Enable logging** for debugging# API Reference

Complete API documentation for Hexabase KaaS.

## In This Section

### [REST API](./rest-api.md)
Complete REST API reference including:
- Authentication endpoints
- Organization management
- Workspace operations
- Project management
- Billing APIs
- Monitoring endpoints

### [WebSocket API](./websocket-api.md)
Real-time communication APIs:
- Connection management
- Event subscriptions
- Provisioning status updates
- Real-time metrics

### [Authentication](./authentication.md)
Authentication and authorization details:
- OAuth2/OIDC flows
- Token management
- API key authentication
- Permission model

### [Error Codes](./error-codes.md)
Comprehensive error code reference:
- HTTP status codes
- Application error codes
- Error response format
- Troubleshooting guide

## API Overview

### Base URL
```
Production: https://api.hexabase.ai
Staging: https://api-staging.hexabase.ai
```

### API Versioning
All API endpoints are versioned. The current version is `v1`.
```
https://api.hexabase.ai/api/v1/...
```

### Request Format
- **Content-Type**: `application/json`
- **Accept**: `application/json`
- **Authorization**: `Bearer <token>`

### Response Format
All responses follow a consistent format:

**Success Response:**
```json
{
  "data": {
    // Response data
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2024-01-20T10:30:00Z"
  }
}
```

**Error Response:**
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "The requested resource was not found",
    "details": {
      // Additional error context
    }
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2024-01-20T10:30:00Z"
  }
}
```

## Quick Start

### 1. Obtain Access Token
```bash
POST /auth/login/google
{
  "id_token": "google-id-token"
}
```

### 2. Make Authenticated Request
```bash
GET /api/v1/organizations
Authorization: Bearer <access-token>
```

### 3. Handle Responses
Check the HTTP status code and parse the JSON response accordingly.

## Rate Limiting

API requests are rate limited per user:
- **Default**: 1000 requests per hour
- **Authenticated**: 5000 requests per hour
- **Premium**: Custom limits

Rate limit headers:
```
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4999
X-RateLimit-Reset: 1642684800
```

## Pagination

List endpoints support pagination:
```
GET /api/v1/organizations?page=2&limit=20
```

Response includes pagination metadata:
```json
{
  "data": [...],
  "meta": {
    "pagination": {
      "page": 2,
      "limit": 20,
      "total": 150,
      "pages": 8
    }
  }
}
```

## Filtering and Sorting

Many endpoints support filtering and sorting:
```
GET /api/v1/workspaces?status=active&sort=-created_at
```

## WebSocket Connection

For real-time updates:
```javascript
const ws = new WebSocket('wss://api.hexabase.ai/ws');
ws.send(JSON.stringify({
  type: 'auth',
  token: 'bearer-token'
}));
```

## SDK Support

Official SDKs are available for:
- Go
- JavaScript/TypeScript
- Python (coming soon)
- Java (coming soon)

## API Changelog

See [API Changelog](./changelog.md) for version history and breaking changes.# Error Codes

This document provides a comprehensive reference for all error codes returned by the Hexabase KaaS API.

## Error Response Format

All API errors follow a consistent format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      // Additional context-specific information
    },
    "request_id": "req_123456"
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2024-01-20T10:00:00Z"
  }
}
```

## HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | OK - Request succeeded |
| 201 | Created - Resource created successfully |
| 204 | No Content - Request succeeded with no response body |
| 400 | Bad Request - Invalid request format or parameters |
| 401 | Unauthorized - Authentication required or failed |
| 403 | Forbidden - Authenticated but not authorized |
| 404 | Not Found - Resource does not exist |
| 409 | Conflict - Resource already exists or state conflict |
| 422 | Unprocessable Entity - Valid format but semantic errors |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |
| 502 | Bad Gateway - Upstream service error |
| 503 | Service Unavailable - Service temporarily unavailable |

## Error Code Categories

Error codes are organized into categories for easier identification:

- **AUTH_*** - Authentication and authorization errors
- **VALIDATION_*** - Input validation errors
- **RESOURCE_*** - Resource-related errors
- **BILLING_*** - Billing and subscription errors
- **WORKSPACE_*** - Workspace-specific errors
- **PROJECT_*** - Project-specific errors
- **QUOTA_*** - Resource quota errors
- **RATE_*** - Rate limiting errors
- **SYSTEM_*** - System and infrastructure errors

## Authentication Errors (AUTH_*)

### AUTH_REQUIRED
**HTTP Status**: 401  
**Description**: Authentication is required to access this resource  
**Example**:
```json
{
  "error": {
    "code": "AUTH_REQUIRED",
    "message": "Authentication required",
    "details": {
      "realm": "api"
    }
  }
}
```

### AUTH_INVALID_TOKEN
**HTTP Status**: 401  
**Description**: The provided JWT token is invalid  
**Example**:
```json
{
  "error": {
    "code": "AUTH_INVALID_TOKEN",
    "message": "Invalid authentication token",
    "details": {
      "reason": "malformed_token"
    }
  }
}
```

### AUTH_TOKEN_EXPIRED
**HTTP Status**: 401  
**Description**: The JWT token has expired  
**Example**:
```json
{
  "error": {
    "code": "AUTH_TOKEN_EXPIRED",
    "message": "Authentication token has expired",
    "details": {
      "expired_at": "2024-01-20T09:00:00Z"
    }
  }
}
```

### AUTH_INVALID_CREDENTIALS
**HTTP Status**: 401  
**Description**: Invalid login credentials provided  
**Example**:
```json
{
  "error": {
    "code": "AUTH_INVALID_CREDENTIALS",
    "message": "Invalid email or password",
    "details": {}
  }
}
```

### AUTH_PROVIDER_ERROR
**HTTP Status**: 502  
**Description**: OAuth provider returned an error  
**Example**:
```json
{
  "error": {
    "code": "AUTH_PROVIDER_ERROR",
    "message": "Authentication provider error",
    "details": {
      "provider": "google",
      "provider_error": "invalid_grant"
    }
  }
}
```

### AUTH_MFA_REQUIRED
**HTTP Status**: 403  
**Description**: Multi-factor authentication is required  
**Example**:
```json
{
  "error": {
    "code": "AUTH_MFA_REQUIRED",
    "message": "Multi-factor authentication required",
    "details": {
      "session_token": "mfa_session_123"
    }
  }
}
```

### AUTH_PERMISSION_DENIED
**HTTP Status**: 403  
**Description**: User lacks required permissions  
**Example**:
```json
{
  "error": {
    "code": "AUTH_PERMISSION_DENIED",
    "message": "Permission denied",
    "details": {
      "required_permission": "workspaces:write",
      "resource": "workspace:ws-123"
    }
  }
}
```

## Validation Errors (VALIDATION_*)

### VALIDATION_ERROR
**HTTP Status**: 400  
**Description**: General validation error  
**Example**:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "fields": {
        "name": "Name is required",
        "email": "Invalid email format"
      }
    }
  }
}
```

### VALIDATION_FIELD_REQUIRED
**HTTP Status**: 400  
**Description**: Required field is missing  
**Example**:
```json
{
  "error": {
    "code": "VALIDATION_FIELD_REQUIRED",
    "message": "Required field missing",
    "details": {
      "field": "organization_id"
    }
  }
}
```

### VALIDATION_FIELD_INVALID
**HTTP Status**: 400  
**Description**: Field value is invalid  
**Example**:
```json
{
  "error": {
    "code": "VALIDATION_FIELD_INVALID",
    "message": "Invalid field value",
    "details": {
      "field": "email",
      "value": "not-an-email",
      "reason": "must be a valid email address"
    }
  }
}
```

### VALIDATION_FIELD_TOO_LONG
**HTTP Status**: 400  
**Description**: Field value exceeds maximum length  
**Example**:
```json
{
  "error": {
    "code": "VALIDATION_FIELD_TOO_LONG",
    "message": "Field value too long",
    "details": {
      "field": "name",
      "max_length": 255,
      "actual_length": 300
    }
  }
}
```

## Resource Errors (RESOURCE_*)

### RESOURCE_NOT_FOUND
**HTTP Status**: 404  
**Description**: Requested resource does not exist  
**Example**:
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Resource not found",
    "details": {
      "resource_type": "workspace",
      "resource_id": "ws-123"
    }
  }
}
```

### RESOURCE_ALREADY_EXISTS
**HTTP Status**: 409  
**Description**: Resource with same identifier already exists  
**Example**:
```json
{
  "error": {
    "code": "RESOURCE_ALREADY_EXISTS",
    "message": "Resource already exists",
    "details": {
      "resource_type": "project",
      "field": "namespace",
      "value": "frontend"
    }
  }
}
```

### RESOURCE_IN_USE
**HTTP Status**: 409  
**Description**: Resource cannot be deleted because it's in use  
**Example**:
```json
{
  "error": {
    "code": "RESOURCE_IN_USE",
    "message": "Resource is in use and cannot be deleted",
    "details": {
      "resource_type": "organization",
      "resource_id": "org-123",
      "used_by": "3 active workspaces"
    }
  }
}
```

### RESOURCE_LIMIT_EXCEEDED
**HTTP Status**: 422  
**Description**: Resource limit for the account has been exceeded  
**Example**:
```json
{
  "error": {
    "code": "RESOURCE_LIMIT_EXCEEDED",
    "message": "Resource limit exceeded",
    "details": {
      "resource_type": "workspace",
      "limit": 5,
      "current": 5,
      "plan": "standard"
    }
  }
}
```

## Billing Errors (BILLING_*)

### BILLING_PAYMENT_FAILED
**HTTP Status**: 402  
**Description**: Payment processing failed  
**Example**:
```json
{
  "error": {
    "code": "BILLING_PAYMENT_FAILED",
    "message": "Payment failed",
    "details": {
      "reason": "insufficient_funds",
      "last4": "4242",
      "amount": 9900,
      "currency": "usd"
    }
  }
}
```

### BILLING_SUBSCRIPTION_INACTIVE
**HTTP Status**: 403  
**Description**: Subscription is not active  
**Example**:
```json
{
  "error": {
    "code": "BILLING_SUBSCRIPTION_INACTIVE",
    "message": "Subscription is not active",
    "details": {
      "status": "canceled",
      "ended_at": "2024-01-15T00:00:00Z"
    }
  }
}
```

### BILLING_PLAN_LIMIT_EXCEEDED
**HTTP Status**: 403  
**Description**: Action exceeds plan limits  
**Example**:
```json
{
  "error": {
    "code": "BILLING_PLAN_LIMIT_EXCEEDED",
    "message": "Plan limit exceeded",
    "details": {
      "plan": "free",
      "limit_type": "workspaces",
      "limit": 1,
      "requested": 2
    }
  }
}
```

### BILLING_INVALID_PAYMENT_METHOD
**HTTP Status**: 400  
**Description**: Payment method is invalid or expired  
**Example**:
```json
{
  "error": {
    "code": "BILLING_INVALID_PAYMENT_METHOD",
    "message": "Invalid payment method",
    "details": {
      "reason": "card_expired",
      "exp_month": 12,
      "exp_year": 2023
    }
  }
}
```

## Workspace Errors (WORKSPACE_*)

### WORKSPACE_PROVISIONING_FAILED
**HTTP Status**: 500  
**Description**: Workspace provisioning failed  
**Example**:
```json
{
  "error": {
    "code": "WORKSPACE_PROVISIONING_FAILED",
    "message": "Failed to provision workspace",
    "details": {
      "workspace_id": "ws-123",
      "stage": "vcluster_creation",
      "reason": "insufficient_resources"
    }
  }
}
```

### WORKSPACE_NOT_READY
**HTTP Status**: 503  
**Description**: Workspace is not ready for operations  
**Example**:
```json
{
  "error": {
    "code": "WORKSPACE_NOT_READY",
    "message": "Workspace is not ready",
    "details": {
      "workspace_id": "ws-123",
      "status": "provisioning",
      "retry_after": 30
    }
  }
}
```

### WORKSPACE_SUSPENDED
**HTTP Status**: 403  
**Description**: Workspace has been suspended  
**Example**:
```json
{
  "error": {
    "code": "WORKSPACE_SUSPENDED",
    "message": "Workspace is suspended",
    "details": {
      "workspace_id": "ws-123",
      "reason": "payment_overdue",
      "suspended_at": "2024-01-15T00:00:00Z"
    }
  }
}
```

## Project Errors (PROJECT_*)

### PROJECT_QUOTA_EXCEEDED
**HTTP Status**: 422  
**Description**: Project resource quota exceeded  
**Example**:
```json
{
  "error": {
    "code": "PROJECT_QUOTA_EXCEEDED",
    "message": "Project resource quota exceeded",
    "details": {
      "project_id": "proj-123",
      "resource": "memory",
      "requested": "5Gi",
      "limit": "4Gi",
      "current_usage": "3.8Gi"
    }
  }
}
```

### PROJECT_HIERARCHY_DEPTH_EXCEEDED
**HTTP Status**: 422  
**Description**: Maximum project hierarchy depth exceeded  
**Example**:
```json
{
  "error": {
    "code": "PROJECT_HIERARCHY_DEPTH_EXCEEDED",
    "message": "Maximum project hierarchy depth exceeded",
    "details": {
      "max_depth": 5,
      "current_depth": 5
    }
  }
}
```

## Quota Errors (QUOTA_*)

### QUOTA_CPU_EXCEEDED
**HTTP Status**: 422  
**Description**: CPU quota exceeded  
**Example**:
```json
{
  "error": {
    "code": "QUOTA_CPU_EXCEEDED",
    "message": "CPU quota exceeded",
    "details": {
      "requested": "5",
      "available": "2",
      "limit": "10",
      "current_usage": "8"
    }
  }
}
```

### QUOTA_MEMORY_EXCEEDED
**HTTP Status**: 422  
**Description**: Memory quota exceeded  
**Example**:
```json
{
  "error": {
    "code": "QUOTA_MEMORY_EXCEEDED",
    "message": "Memory quota exceeded",
    "details": {
      "requested": "16Gi",
      "available": "8Gi",
      "limit": "32Gi",
      "current_usage": "24Gi"
    }
  }
}
```

### QUOTA_STORAGE_EXCEEDED
**HTTP Status**: 422  
**Description**: Storage quota exceeded  
**Example**:
```json
{
  "error": {
    "code": "QUOTA_STORAGE_EXCEEDED",
    "message": "Storage quota exceeded",
    "details": {
      "requested": "50Gi",
      "available": "20Gi",
      "limit": "100Gi",
      "current_usage": "80Gi"
    }
  }
}
```

## Rate Limiting Errors (RATE_*)

### RATE_LIMIT_EXCEEDED
**HTTP Status**: 429  
**Description**: API rate limit exceeded  
**Example**:
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded",
    "details": {
      "limit": 5000,
      "window": "1h",
      "retry_after": 1234
    }
  }
}
```

### RATE_LIMIT_AUTH_EXCEEDED
**HTTP Status**: 429  
**Description**: Authentication rate limit exceeded  
**Example**:
```json
{
  "error": {
    "code": "RATE_LIMIT_AUTH_EXCEEDED",
    "message": "Too many authentication attempts",
    "details": {
      "limit": 5,
      "window": "1m",
      "retry_after": 45
    }
  }
}
```

## System Errors (SYSTEM_*)

### SYSTEM_INTERNAL_ERROR
**HTTP Status**: 500  
**Description**: Internal server error  
**Example**:
```json
{
  "error": {
    "code": "SYSTEM_INTERNAL_ERROR",
    "message": "An internal error occurred",
    "details": {
      "request_id": "req_123456"
    }
  }
}
```

### SYSTEM_SERVICE_UNAVAILABLE
**HTTP Status**: 503  
**Description**: Service temporarily unavailable  
**Example**:
```json
{
  "error": {
    "code": "SYSTEM_SERVICE_UNAVAILABLE",
    "message": "Service temporarily unavailable",
    "details": {
      "service": "kubernetes_api",
      "retry_after": 30
    }
  }
}
```

### SYSTEM_MAINTENANCE
**HTTP Status**: 503  
**Description**: System under maintenance  
**Example**:
```json
{
  "error": {
    "code": "SYSTEM_MAINTENANCE",
    "message": "System under maintenance",
    "details": {
      "maintenance_end": "2024-01-20T12:00:00Z",
      "affected_services": ["workspace_provisioning"]
    }
  }
}
```

### SYSTEM_DATABASE_ERROR
**HTTP Status**: 500  
**Description**: Database operation failed  
**Example**:
```json
{
  "error": {
    "code": "SYSTEM_DATABASE_ERROR",
    "message": "Database operation failed",
    "details": {
      "operation": "insert",
      "table": "workspaces"
    }
  }
}
```

## WebSocket Errors

### WS_AUTH_REQUIRED
**Description**: WebSocket authentication required  
**Example**:
```json
{
  "type": "error",
  "error": {
    "code": "WS_AUTH_REQUIRED",
    "message": "Authentication required before subscribing"
  }
}
```

### WS_SUBSCRIPTION_LIMIT_EXCEEDED
**Description**: WebSocket subscription limit exceeded  
**Example**:
```json
{
  "type": "error",
  "error": {
    "code": "WS_SUBSCRIPTION_LIMIT_EXCEEDED",
    "message": "Maximum number of subscriptions exceeded",
    "details": {
      "limit": 100,
      "current": 100
    }
  }
}
```

### WS_MESSAGE_TOO_LARGE
**Description**: WebSocket message size limit exceeded  
**Example**:
```json
{
  "type": "error",
  "error": {
    "code": "WS_MESSAGE_TOO_LARGE",
    "message": "Message size exceeds limit",
    "details": {
      "max_size": 1048576,
      "actual_size": 2097152
    }
  }
}
```

## Error Handling Best Practices

### Client-Side Error Handling

```javascript
async function makeApiRequest(url, options) {
  try {
    const response = await fetch(url, options);
    
    if (!response.ok) {
      const error = await response.json();
      
      switch (error.error.code) {
        case 'AUTH_TOKEN_EXPIRED':
          // Refresh token and retry
          await refreshToken();
          return makeApiRequest(url, options);
          
        case 'RATE_LIMIT_EXCEEDED':
          // Wait and retry
          const retryAfter = error.error.details.retry_after;
          await sleep(retryAfter * 1000);
          return makeApiRequest(url, options);
          
        case 'RESOURCE_NOT_FOUND':
          // Handle 404
          throw new NotFoundError(error.error.message);
          
        default:
          // Handle other errors
          throw new ApiError(error.error);
      }
    }
    
    return response.json();
  } catch (error) {
    // Handle network errors
    console.error('API request failed:', error);
    throw error;
  }
}
```

### Retry Strategy

For transient errors, implement exponential backoff:

```javascript
async function retryWithBackoff(fn, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (error.code === 'SYSTEM_SERVICE_UNAVAILABLE' && i < maxRetries - 1) {
        const delay = Math.pow(2, i) * 1000; // Exponential backoff
        await sleep(delay);
        continue;
      }
      throw error;
    }
  }
}
```

### Error Logging

Always log the request ID for troubleshooting:

```javascript
function logError(error) {
  console.error('API Error:', {
    code: error.code,
    message: error.message,
    request_id: error.request_id,
    details: error.details
  });
}
```

## Common Error Scenarios

### Scenario: Expired Token

```
1. Client makes request with expired token
2. Server returns AUTH_TOKEN_EXPIRED
3. Client uses refresh token to get new access token
4. Client retries original request with new token
```

### Scenario: Resource Creation Conflict

```
1. Client attempts to create resource
2. Server returns RESOURCE_ALREADY_EXISTS
3. Client can either:
   - Use existing resource
   - Update existing resource
   - Choose different identifier
```

### Scenario: Quota Exceeded

```
1. Client requests resource allocation
2. Server returns QUOTA_CPU_EXCEEDED
3. Client can either:
   - Request smaller allocation
   - Upgrade plan for more resources
   - Free up existing resources
```

## Getting Help

If you encounter an error not documented here or need assistance:

1. Note the `request_id` from the error response
2. Check our [Status Page](https://status.hexabase.ai) for any ongoing issues
3. Contact support with the request ID and error details
4. For critical issues, use the emergency support channel# REST API Reference

This document provides a complete reference for the Hexabase KaaS REST API.

## Base URL

```
Production: https://api.hexabase.ai
Staging: https://api-staging.hexabase.ai
Local: http://api.localhost
```

## API Version

All endpoints are prefixed with `/api/v1`

## Authentication

All API requests (except auth endpoints) require a valid JWT token in the Authorization header:

```
Authorization: Bearer <jwt-token>
```

## Common Headers

```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <jwt-token>
```

## Response Format

### Success Response

```json
{
  "data": {
    // Response data
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2024-01-20T10:30:00Z"
  }
}
```

### Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "field": "name",
      "reason": "required"
    }
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2024-01-20T10:30:00Z"
  }
}
```

## Endpoints

### Authentication

#### Login with OAuth Provider

```http
POST /auth/login/:provider
```

**Parameters:**
- `provider` (path) - OAuth provider name (`google`, `github`, `azure`)

**Request Body:**
```json
{
  "id_token": "provider-id-token",
  "code": "authorization-code",
  "redirect_uri": "https://app.hexabase.ai/auth/callback"
}
```

**Response:**
```json
{
  "data": {
    "access_token": "jwt-access-token",
    "refresh_token": "jwt-refresh-token",
    "expires_in": 3600,
    "user": {
      "id": "user-123",
      "email": "user@example.com",
      "name": "John Doe",
      "picture": "https://..."
    }
  }
}
```

#### OAuth Callback

```http
GET /auth/callback/:provider
POST /auth/callback/:provider
```

**Query Parameters:**
- `code` - Authorization code
- `state` - OAuth state parameter

#### Refresh Token

```http
POST /auth/refresh
```

**Request Body:**
```json
{
  "refresh_token": "jwt-refresh-token"
}
```

**Response:**
```json
{
  "data": {
    "access_token": "new-jwt-access-token",
    "refresh_token": "new-jwt-refresh-token",
    "expires_in": 3600
  }
}
```

#### Logout

```http
POST /auth/logout
```

**Headers:**
- `Authorization: Bearer <token>` (required)

**Response:**
```json
{
  "data": {
    "message": "Logged out successfully"
  }
}
```

#### Get Current User

```http
GET /auth/me
```

**Response:**
```json
{
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "name": "John Doe",
    "picture": "https://...",
    "provider": "google",
    "created_at": "2024-01-01T00:00:00Z",
    "last_login": "2024-01-20T10:00:00Z"
  }
}
```

### Organizations

#### List Organizations

```http
GET /api/v1/organizations
```

**Query Parameters:**
- `page` (integer) - Page number (default: 1)
- `limit` (integer) - Items per page (default: 20, max: 100)
- `search` (string) - Search by name

**Response:**
```json
{
  "data": [
    {
      "id": "org-123",
      "name": "My Organization",
      "slug": "my-org",
      "owner_id": "user-123",
      "created_at": "2024-01-01T00:00:00Z",
      "member_count": 5,
      "workspace_count": 3
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "pages": 5
    }
  }
}
```

#### Create Organization

```http
POST /api/v1/organizations
```

**Request Body:**
```json
{
  "name": "My Organization",
  "description": "Organization description"
}
```

**Response:**
```json
{
  "data": {
    "id": "org-123",
    "name": "My Organization",
    "slug": "my-org",
    "description": "Organization description",
    "owner_id": "user-123",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Get Organization

```http
GET /api/v1/organizations/:orgId
```

**Response:**
```json
{
  "data": {
    "id": "org-123",
    "name": "My Organization",
    "slug": "my-org",
    "description": "Organization description",
    "owner_id": "user-123",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-20T00:00:00Z",
    "member_count": 5,
    "workspace_count": 3,
    "billing": {
      "plan": "pro",
      "status": "active"
    }
  }
}
```

#### Update Organization

```http
PUT /api/v1/organizations/:orgId
```

**Request Body:**
```json
{
  "name": "Updated Organization Name",
  "description": "Updated description"
}
```

#### Delete Organization

```http
DELETE /api/v1/organizations/:orgId
```

### Organization Members

#### List Organization Members

```http
GET /api/v1/organizations/:orgId/members
```

**Response:**
```json
{
  "data": [
    {
      "id": "member-123",
      "user_id": "user-123",
      "organization_id": "org-123",
      "role": "admin",
      "joined_at": "2024-01-01T00:00:00Z",
      "user": {
        "id": "user-123",
        "email": "user@example.com",
        "name": "John Doe",
        "picture": "https://..."
      }
    }
  ]
}
```

#### Remove Organization Member

```http
DELETE /api/v1/organizations/:orgId/members/:userId
```

#### Update Member Role

```http
PUT /api/v1/organizations/:orgId/members/:userId/role
```

**Request Body:**
```json
{
  "role": "admin"
}
```

**Valid Roles:**
- `owner` - Full control
- `admin` - Administrative access
- `member` - Regular member

### Organization Invitations

#### Invite User

```http
POST /api/v1/organizations/:orgId/invitations
```

**Request Body:**
```json
{
  "email": "newuser@example.com",
  "role": "member"
}
```

**Response:**
```json
{
  "data": {
    "id": "invite-123",
    "organization_id": "org-123",
    "email": "newuser@example.com",
    "role": "member",
    "token": "invitation-token",
    "expires_at": "2024-01-27T00:00:00Z",
    "created_at": "2024-01-20T00:00:00Z"
  }
}
```

#### List Pending Invitations

```http
GET /api/v1/organizations/:orgId/invitations
```

#### Accept Invitation

```http
POST /api/v1/organizations/invitations/:token/accept
```

#### Cancel Invitation

```http
DELETE /api/v1/organizations/invitations/:invitationId
```

### Workspaces

#### List Workspaces

```http
GET /api/v1/organizations/:orgId/workspaces
```

**Query Parameters:**
- `page` (integer) - Page number
- `limit` (integer) - Items per page
- `status` (string) - Filter by status (`active`, `provisioning`, `suspended`)

**Response:**
```json
{
  "data": [
    {
      "id": "ws-123",
      "name": "Development Workspace",
      "slug": "dev-workspace",
      "organization_id": "org-123",
      "status": "active",
      "plan": "standard",
      "created_at": "2024-01-01T00:00:00Z",
      "resource_usage": {
        "cpu": "2.5 cores",
        "memory": "8Gi",
        "storage": "50Gi"
      }
    }
  ]
}
```

#### Create Workspace

```http
POST /api/v1/organizations/:orgId/workspaces
```

**Request Body:**
```json
{
  "name": "Development Workspace",
  "plan": "standard",
  "kubernetes_version": "1.28"
}
```

**Response:**
```json
{
  "data": {
    "id": "ws-123",
    "name": "Development Workspace",
    "slug": "dev-workspace",
    "organization_id": "org-123",
    "status": "provisioning",
    "plan": "standard",
    "kubernetes_version": "1.28",
    "created_at": "2024-01-20T00:00:00Z",
    "provisioning_task_id": "task-123"
  }
}
```

#### Get Workspace

```http
GET /api/v1/organizations/:orgId/workspaces/:wsId
```

**Response:**
```json
{
  "data": {
    "id": "ws-123",
    "name": "Development Workspace",
    "slug": "dev-workspace",
    "organization_id": "org-123",
    "status": "active",
    "plan": "standard",
    "kubernetes_version": "1.28",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-20T00:00:00Z",
    "vcluster": {
      "name": "vcluster-ws-123",
      "namespace": "ws-123",
      "api_endpoint": "https://ws-123.api.hexabase.ai",
      "version": "0.19.0"
    },
    "resource_limits": {
      "cpu": "10 cores",
      "memory": "32Gi",
      "storage": "100Gi"
    },
    "resource_usage": {
      "cpu": "2.5 cores",
      "memory": "8Gi",
      "storage": "50Gi"
    }
  }
}
```

#### Update Workspace

```http
PUT /api/v1/organizations/:orgId/workspaces/:wsId
```

**Request Body:**
```json
{
  "name": "Updated Workspace Name",
  "plan": "pro"
}
```

#### Delete Workspace

```http
DELETE /api/v1/organizations/:orgId/workspaces/:wsId
```

#### Get Kubeconfig

```http
GET /api/v1/organizations/:orgId/workspaces/:wsId/kubeconfig
```

**Response:**
```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://ws-123.api.hexabase.ai
    certificate-authority-data: ...
  name: hexabase-ws-123
contexts:
- context:
    cluster: hexabase-ws-123
    user: hexabase-user
  name: hexabase-ws-123
current-context: hexabase-ws-123
users:
- name: hexabase-user
  user:
    token: ...
```

#### Get Resource Usage

```http
GET /api/v1/organizations/:orgId/workspaces/:wsId/resource-usage
```

**Response:**
```json
{
  "data": {
    "cpu": {
      "used": "2.5",
      "limit": "10",
      "unit": "cores",
      "percentage": 25
    },
    "memory": {
      "used": "8192",
      "limit": "32768",
      "unit": "Mi",
      "percentage": 25
    },
    "storage": {
      "used": "51200",
      "limit": "102400",
      "unit": "Mi",
      "percentage": 50
    },
    "pods": {
      "used": 15,
      "limit": 100,
      "percentage": 15
    }
  }
}
```

### Projects

#### List Projects

```http
GET /api/v1/workspaces/:wsId/projects
```

**Query Parameters:**
- `page` (integer) - Page number
- `limit` (integer) - Items per page
- `parent_id` (string) - Filter by parent project

**Response:**
```json
{
  "data": [
    {
      "id": "proj-123",
      "name": "Frontend Project",
      "namespace": "frontend",
      "workspace_id": "ws-123",
      "parent_id": null,
      "created_at": "2024-01-01T00:00:00Z",
      "resource_quota": {
        "cpu": "2 cores",
        "memory": "4Gi",
        "storage": "10Gi"
      }
    }
  ]
}
```

#### Create Project

```http
POST /api/v1/workspaces/:wsId/projects
```

**Request Body:**
```json
{
  "name": "Frontend Project",
  "parent_id": null,
  "resource_quota": {
    "cpu": "2",
    "memory": "4Gi",
    "storage": "10Gi"
  }
}
```

#### Get Project

```http
GET /api/v1/workspaces/:wsId/projects/:projectId
```

#### Update Project

```http
PUT /api/v1/workspaces/:wsId/projects/:projectId
```

**Request Body:**
```json
{
  "name": "Updated Project Name",
  "resource_quota": {
    "cpu": "4",
    "memory": "8Gi",
    "storage": "20Gi"
  }
}
```

#### Delete Project

```http
DELETE /api/v1/workspaces/:wsId/projects/:projectId
```

#### Create Sub-Project

```http
POST /api/v1/workspaces/:wsId/projects/:projectId/subprojects
```

**Request Body:**
```json
{
  "name": "Sub-Project Name"
}
```

#### Get Project Hierarchy

```http
GET /api/v1/workspaces/:wsId/projects/:projectId/hierarchy
```

**Response:**
```json
{
  "data": {
    "id": "proj-123",
    "name": "Parent Project",
    "namespace": "parent",
    "children": [
      {
        "id": "proj-456",
        "name": "Child Project 1",
        "namespace": "parent-child1",
        "children": []
      },
      {
        "id": "proj-789",
        "name": "Child Project 2",
        "namespace": "parent-child2",
        "children": []
      }
    ]
  }
}
```

### Billing

#### Get Subscription

```http
GET /api/v1/organizations/:orgId/billing/subscription
```

**Response:**
```json
{
  "data": {
    "id": "sub_123",
    "organization_id": "org-123",
    "plan": "pro",
    "status": "active",
    "current_period_start": "2024-01-01T00:00:00Z",
    "current_period_end": "2024-02-01T00:00:00Z",
    "cancel_at_period_end": false,
    "items": [
      {
        "id": "si_123",
        "price_id": "price_123",
        "quantity": 5,
        "description": "Pro Plan - Per Workspace"
      }
    ]
  }
}
```

#### Create Subscription

```http
POST /api/v1/organizations/:orgId/billing/subscription
```

**Request Body:**
```json
{
  "price_id": "price_pro_monthly",
  "quantity": 5,
  "payment_method_id": "pm_123"
}
```

#### Update Subscription

```http
PUT /api/v1/organizations/:orgId/billing/subscription
```

**Request Body:**
```json
{
  "items": [
    {
      "price_id": "price_pro_monthly",
      "quantity": 10
    }
  ]
}
```

#### Cancel Subscription

```http
DELETE /api/v1/organizations/:orgId/billing/subscription
```

**Query Parameters:**
- `immediately` (boolean) - Cancel immediately or at period end

### Payment Methods

#### List Payment Methods

```http
GET /api/v1/organizations/:orgId/billing/payment-methods
```

**Response:**
```json
{
  "data": [
    {
      "id": "pm_123",
      "type": "card",
      "card": {
        "brand": "visa",
        "last4": "4242",
        "exp_month": 12,
        "exp_year": 2025
      },
      "is_default": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### Add Payment Method

```http
POST /api/v1/organizations/:orgId/billing/payment-methods
```

**Request Body:**
```json
{
  "payment_method_id": "pm_123"
}
```

#### Set Default Payment Method

```http
PUT /api/v1/organizations/:orgId/billing/payment-methods/:methodId/default
```

#### Remove Payment Method

```http
DELETE /api/v1/organizations/:orgId/billing/payment-methods/:methodId
```

### Invoices

#### List Invoices

```http
GET /api/v1/organizations/:orgId/billing/invoices
```

**Query Parameters:**
- `page` (integer) - Page number
- `limit` (integer) - Items per page
- `status` (string) - Filter by status

**Response:**
```json
{
  "data": [
    {
      "id": "in_123",
      "number": "INV-2024-001",
      "status": "paid",
      "amount_due": 10000,
      "amount_paid": 10000,
      "currency": "usd",
      "created_at": "2024-01-01T00:00:00Z",
      "paid_at": "2024-01-05T00:00:00Z",
      "period_start": "2024-01-01T00:00:00Z",
      "period_end": "2024-02-01T00:00:00Z"
    }
  ]
}
```

#### Get Invoice

```http
GET /api/v1/organizations/:orgId/billing/invoices/:invoiceId
```

#### Download Invoice

```http
GET /api/v1/organizations/:orgId/billing/invoices/:invoiceId/download
```

**Response:** PDF file

#### Get Upcoming Invoice

```http
GET /api/v1/organizations/:orgId/billing/invoices/upcoming
```

### Usage

#### Get Current Usage

```http
GET /api/v1/organizations/:orgId/billing/usage/current
```

**Response:**
```json
{
  "data": {
    "period_start": "2024-01-01T00:00:00Z",
    "period_end": "2024-02-01T00:00:00Z",
    "workspaces": {
      "active": 5,
      "limit": 10
    },
    "compute_hours": {
      "used": 3600,
      "included": 5000,
      "overage": 0
    },
    "storage_gb_hours": {
      "used": 72000,
      "included": 100000,
      "overage": 0
    }
  }
}
```

### Monitoring

#### Get Workspace Metrics

```http
GET /api/v1/workspaces/:wsId/monitoring/metrics
```

**Query Parameters:**
- `period` (string) - Time period (`1h`, `6h`, `24h`, `7d`, `30d`)
- `metric` (string) - Specific metric to retrieve

**Response:**
```json
{
  "data": {
    "cpu": {
      "series": [
        {
          "timestamp": "2024-01-20T10:00:00Z",
          "value": 2.5
        }
      ]
    },
    "memory": {
      "series": [
        {
          "timestamp": "2024-01-20T10:00:00Z",
          "value": 8192
        }
      ]
    }
  }
}
```

#### Get Cluster Health

```http
GET /api/v1/workspaces/:wsId/monitoring/health
```

**Response:**
```json
{
  "data": {
    "status": "healthy",
    "checks": {
      "api_server": {
        "status": "healthy",
        "message": "API server is responsive"
      },
      "etcd": {
        "status": "healthy",
        "message": "etcd cluster is healthy"
      },
      "scheduler": {
        "status": "healthy",
        "message": "Scheduler is working"
      },
      "controller_manager": {
        "status": "healthy",
        "message": "Controller manager is running"
      }
    },
    "nodes": {
      "ready": 3,
      "total": 3
    }
  }
}
```

#### Get Alerts

```http
GET /api/v1/workspaces/:wsId/monitoring/alerts
```

**Query Parameters:**
- `severity` (string) - Filter by severity (`critical`, `warning`, `info`)
- `status` (string) - Filter by status (`firing`, `resolved`)

**Response:**
```json
{
  "data": [
    {
      "id": "alert-123",
      "workspace_id": "ws-123",
      "type": "high_cpu_usage",
      "severity": "warning",
      "status": "firing",
      "title": "High CPU Usage",
      "description": "CPU usage has been above 80% for 10 minutes",
      "resource": "deployment/api-server",
      "threshold": 80,
      "value": 85.5,
      "started_at": "2024-01-20T10:00:00Z"
    }
  ]
}
```

#### Create Alert

```http
POST /api/v1/workspaces/:wsId/monitoring/alerts
```

**Request Body:**
```json
{
  "type": "custom",
  "severity": "warning",
  "title": "Custom Alert",
  "description": "Alert description",
  "resource": "deployment/my-app",
  "threshold": 90,
  "value": 95
}
```

#### Acknowledge Alert

```http
PUT /api/v1/workspaces/:wsId/monitoring/alerts/:alertId/acknowledge
```

#### Resolve Alert

```http
PUT /api/v1/workspaces/:wsId/monitoring/alerts/:alertId/resolve
```

## Rate Limiting

API requests are rate limited based on your subscription plan:

| Plan | Requests per Hour |
|------|-------------------|
| Free | 1,000 |
| Standard | 5,000 |
| Pro | 10,000 |
| Enterprise | Unlimited |

Rate limit information is included in response headers:

```http
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4999
X-RateLimit-Reset: 1705749600
```

## Pagination

List endpoints support pagination with the following parameters:

- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

Pagination metadata is included in the response:

```json
{
  "meta": {
    "pagination": {
      "page": 2,
      "limit": 20,
      "total": 150,
      "pages": 8
    }
  }
}
```

## Filtering and Sorting

Many list endpoints support filtering and sorting:

- Use query parameters for filtering (e.g., `?status=active`)
- Use `sort` parameter with field name (prefix with `-` for descending)
  - Example: `?sort=-created_at` (newest first)

## Webhook Events

Hexabase KaaS can send webhook notifications for various events. Configure webhooks in your organization settings.

### Event Types

- `workspace.created`
- `workspace.deleted`
- `workspace.status_changed`
- `project.created`
- `project.deleted`
- `billing.payment_succeeded`
- `billing.payment_failed`
- `alert.triggered`
- `alert.resolved`

### Webhook Payload

```json
{
  "id": "evt_123",
  "type": "workspace.created",
  "created_at": "2024-01-20T10:00:00Z",
  "data": {
    // Event-specific data
  }
}
```

### Webhook Security

All webhooks include a signature in the `X-Hexabase-Signature` header for verification.