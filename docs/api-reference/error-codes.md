# Error Codes

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
4. For critical issues, use the emergency support channel