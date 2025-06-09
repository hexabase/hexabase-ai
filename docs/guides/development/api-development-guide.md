# API Development Guide

This guide covers best practices and patterns for developing the Hexabase AI API.

## Overview

The Hexabase AI API is built with:
- Go 1.24+
- Gin Web Framework
- GORM for database operations
- Wire for dependency injection
- OpenAPI/Swagger for documentation

## Project Structure

```
api/
├── cmd/
│   ├── api/          # API server entry point
│   └── worker/       # Background worker entry point
├── internal/
│   ├── api/          # HTTP handlers and routes
│   ├── domain/       # Business logic and models
│   ├── repository/   # Data access layer
│   ├── service/      # Service layer
│   └── infrastructure/ # Cross-cutting concerns
└── pkg/              # Public packages
```

## Development Patterns

### 1. Domain-Driven Design

We follow DDD principles with clear separation:
- **Domain Models**: Core business entities
- **Repositories**: Data access abstractions
- **Services**: Business logic implementation
- **Handlers**: HTTP request/response handling

### 2. Dependency Injection

Use Wire for compile-time dependency injection:

```go
// wire.go
var ServiceSet = wire.NewSet(
    NewUserService,
    wire.Bind(new(domain.UserService), new(*UserService)),
)
```

### 3. Error Handling

Consistent error handling across the API:

```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

func HandleError(c *gin.Context, err error) {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        c.JSON(apiErr.StatusCode(), apiErr)
        return
    }
    // Handle unknown errors
    c.JSON(500, APIError{Code: "INTERNAL_ERROR", Message: "Internal server error"})
}
```

### 4. Request Validation

Use binding tags for request validation:

```go
type CreateWorkspaceRequest struct {
    Name        string `json:"name" binding:"required,min=3,max=50"`
    Description string `json:"description" binding:"max=200"`
    Plan        string `json:"plan" binding:"required,oneof=starter pro enterprise"`
}
```

## API Conventions

### RESTful Endpoints

Follow REST conventions:
- `GET /api/v1/workspaces` - List resources
- `GET /api/v1/workspaces/:id` - Get single resource
- `POST /api/v1/workspaces` - Create resource
- `PUT /api/v1/workspaces/:id` - Update resource
- `DELETE /api/v1/workspaces/:id` - Delete resource

### Response Format

Consistent response structure:

```json
{
    "data": {
        "id": "123",
        "name": "My Workspace"
    },
    "meta": {
        "page": 1,
        "limit": 20,
        "total": 100
    }
}
```

### Status Codes

- `200 OK` - Successful GET, PUT
- `201 Created` - Successful POST
- `204 No Content` - Successful DELETE
- `400 Bad Request` - Validation errors
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict
- `500 Internal Server Error` - Server errors

## Testing

### Unit Tests

Write tests for all services:

```go
func TestUserService_Create(t *testing.T) {
    // Setup
    mockRepo := mocks.NewMockUserRepository(t)
    service := NewUserService(mockRepo)
    
    // Test
    user := &domain.User{Name: "Test User"}
    mockRepo.On("Create", mock.Anything, user).Return(nil)
    
    err := service.Create(context.Background(), user)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### Integration Tests

Test API endpoints:

```go
func TestCreateWorkspace(t *testing.T) {
    router := setupTestRouter()
    
    body := `{"name": "Test Workspace", "plan": "starter"}`
    req := httptest.NewRequest("POST", "/api/v1/workspaces", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
}
```

## Database Operations

### Migrations

Create migrations for schema changes:

```bash
migrate create -ext sql -dir migrations -seq create_users_table
```

### Transactions

Use transactions for data consistency:

```go
func (s *WorkspaceService) CreateWithProject(ctx context.Context, workspace *Workspace, project *Project) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(workspace).Error; err != nil {
            return err
        }
        project.WorkspaceID = workspace.ID
        return tx.Create(project).Error
    })
}
```

## Security

### Authentication

JWT-based authentication:

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }
        
        claims, err := ValidateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

### Authorization

Role-based access control:

```go
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        for _, role := range roles {
            if userRole == role {
                c.Next()
                return
            }
        }
        c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
    }
}
```

## Performance

### Database Queries

- Use eager loading to avoid N+1 queries
- Add indexes for frequently queried fields
- Use pagination for list endpoints
- Cache frequently accessed data

### Concurrent Operations

```go
func (s *Service) ProcessBatch(items []Item) error {
    errCh := make(chan error, len(items))
    var wg sync.WaitGroup
    
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            if err := s.ProcessItem(item); err != nil {
                errCh <- err
            }
        }(item)
    }
    
    wg.Wait()
    close(errCh)
    
    // Check for errors
    for err := range errCh {
        if err != nil {
            return err
        }
    }
    return nil
}
```

## Monitoring

### Structured Logging

```go
logger.Info("workspace created",
    zap.String("workspace_id", workspace.ID),
    zap.String("user_id", userID),
    zap.String("plan", workspace.Plan),
)
```

### Metrics

```go
var (
    apiRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "Total number of API requests",
        },
        []string{"method", "endpoint", "status"},
    )
)
```

## OpenAPI Documentation

Document all endpoints:

```go
// @Summary Create workspace
// @Description Create a new workspace
// @Tags workspaces
// @Accept json
// @Produce json
// @Param workspace body CreateWorkspaceRequest true "Workspace data"
// @Success 201 {object} WorkspaceResponse
// @Failure 400 {object} APIError
// @Router /api/v1/workspaces [post]
func CreateWorkspace(c *gin.Context) {
    // Implementation
}
```

## Best Practices

1. **Keep handlers thin** - Move business logic to services
2. **Use contexts** - Pass context through all layers
3. **Handle errors gracefully** - Return meaningful error messages
4. **Validate input** - Use struct tags and custom validators
5. **Log appropriately** - Info for normal operations, Error for failures
6. **Write tests** - Aim for >80% coverage
7. **Document APIs** - Keep OpenAPI specs up to date
8. **Use transactions** - Ensure data consistency
9. **Implement idempotency** - For critical operations
10. **Monitor performance** - Add metrics and traces

## Common Patterns

### Repository Pattern

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}
```

### Service Pattern

```go
type UserService struct {
    repo UserRepository
    logger *zap.Logger
}

func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Validation
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Business logic
    user := &User{
        Name: req.Name,
        Email: req.Email,
    }
    
    // Persistence
    if err := s.repo.Create(ctx, user); err != nil {
        s.logger.Error("failed to create user", zap.Error(err))
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

## Debugging Tips

1. Use `gin.DebugMode` for development
2. Enable GORM debug mode for SQL queries
3. Use pprof for performance profiling
4. Add request IDs for tracing
5. Use structured logging for easier debugging

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Wire Documentation](https://github.com/google/wire)