# REST API Reference

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