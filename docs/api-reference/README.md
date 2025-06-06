# API Reference

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

See [API Changelog](./changelog.md) for version history and breaking changes.