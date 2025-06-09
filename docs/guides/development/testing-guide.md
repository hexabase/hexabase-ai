# Testing Guide

Comprehensive testing guide for Hexabase AI platform covering unit tests, integration tests, and end-to-end tests.

## Overview

Our testing strategy includes:
- **Unit Tests**: Test individual components/functions
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete user workflows
- **Performance Tests**: Ensure system meets performance requirements
- **Security Tests**: Validate security controls

Target: **80% code coverage** across all components

## Testing Stack

### Backend (Go)
- Testing framework: Built-in `testing` package
- Mocking: `testify/mock`
- Assertions: `testify/assert`
- HTTP testing: `httptest`
- Database testing: `sqlmock`

### Frontend (TypeScript/React)
- Test runner: Jest
- Component testing: React Testing Library
- E2E testing: Playwright
- Mocking: MSW (Mock Service Worker)

## Backend Testing

### Unit Tests

#### Service Layer Testing

```go
// internal/service/workspace/service_test.go
package workspace

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/hexabase/hexabase-ai/internal/domain/workspace"
    "github.com/hexabase/hexabase-ai/internal/domain/workspace/mocks"
)

func TestWorkspaceService_Create(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockWorkspaceRepository(t)
    mockK8s := mocks.NewMockKubernetesRepository(t)
    service := NewService(mockRepo, mockK8s, nil)
    
    ctx := context.Background()
    ws := &workspace.Workspace{
        Name: "Test Workspace",
        Plan: workspace.PlanStarter,
    }
    
    // Set expectations
    mockRepo.On("Create", ctx, mock.MatchedBy(func(w *workspace.Workspace) bool {
        return w.Name == "Test Workspace"
    })).Return(nil).Once()
    
    mockK8s.On("CreateVCluster", ctx, mock.AnythingOfType("*workspace.Workspace")).
        Return(nil).Once()
    
    // Act
    err := service.Create(ctx, ws)
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, ws.ID)
    assert.Equal(t, workspace.StatusProvisioning, ws.Status)
    
    mockRepo.AssertExpectations(t)
    mockK8s.AssertExpectations(t)
}

func TestWorkspaceService_Create_ValidationError(t *testing.T) {
    // Test validation failures
    testCases := []struct {
        name      string
        workspace *workspace.Workspace
        wantError string
    }{
        {
            name:      "empty name",
            workspace: &workspace.Workspace{Name: "", Plan: workspace.PlanStarter},
            wantError: "workspace name is required",
        },
        {
            name:      "invalid plan",
            workspace: &workspace.Workspace{Name: "Test", Plan: "invalid"},
            wantError: "invalid workspace plan",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            service := NewService(nil, nil, nil)
            err := service.Create(context.Background(), tc.workspace)
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tc.wantError)
        })
    }
}
```

#### Repository Layer Testing

```go
// internal/repository/workspace/postgres_test.go
package workspace

import (
    "context"
    "testing"
    "time"
    
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func TestPostgresRepository_GetByID(t *testing.T) {
    // Setup mock database
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()
    
    gormDB, err := gorm.Open(postgres.New(postgres.Config{
        Conn: db,
    }), &gorm.Config{})
    assert.NoError(t, err)
    
    repo := NewPostgresRepository(gormDB)
    
    // Expected query
    workspaceID := "ws-123"
    rows := sqlmock.NewRows([]string{"id", "name", "plan", "status", "created_at"}).
        AddRow(workspaceID, "Test Workspace", "starter", "active", time.Now())
    
    mock.ExpectQuery("SELECT (.+) FROM \"workspaces\" WHERE id = ?").
        WithArgs(workspaceID).
        WillReturnRows(rows)
    
    // Execute
    workspace, err := repo.GetByID(context.Background(), workspaceID)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, workspaceID, workspace.ID)
    assert.Equal(t, "Test Workspace", workspace.Name)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

#### HTTP Handler Testing

```go
// internal/api/handlers/workspaces_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestWorkspaceHandler_Create(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    mockService := mocks.NewMockWorkspaceService(t)
    handler := NewWorkspaceHandler(mockService)
    
    // Create request
    reqBody := map[string]interface{}{
        "name": "Test Workspace",
        "plan": "starter",
    }
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Mock expectations
    mockService.On("Create", mock.Anything, mock.MatchedBy(func(ws *workspace.Workspace) bool {
        return ws.Name == "Test Workspace" && ws.Plan == "starter"
    })).Return(nil).Once()
    
    // Execute
    w := httptest.NewRecorder()
    router := gin.New()
    router.POST("/api/v1/workspaces", handler.Create)
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "Test Workspace", response["data"].(map[string]interface{})["name"])
    
    mockService.AssertExpectations(t)
}
```

### Integration Tests

```go
// tests/integration/workspace_flow_test.go
package integration

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "github.com/hexabase/hexabase-ai/internal/testutil"
)

type WorkspaceIntegrationSuite struct {
    suite.Suite
    testDB   *testutil.TestDatabase
    testK8s  *testutil.TestKubernetes
    app      *Application
}

func (s *WorkspaceIntegrationSuite) SetupSuite() {
    // Setup test database
    s.testDB = testutil.NewTestDatabase()
    s.testDB.Migrate()
    
    // Setup test Kubernetes
    s.testK8s = testutil.NewTestKubernetes()
    
    // Initialize application
    s.app = NewTestApplication(s.testDB.DB, s.testK8s.Client)
}

func (s *WorkspaceIntegrationSuite) TearDownSuite() {
    s.testDB.Cleanup()
    s.testK8s.Cleanup()
}

func (s *WorkspaceIntegrationSuite) TestCompleteWorkspaceLifecycle() {
    ctx := context.Background()
    
    // Create workspace
    ws := &workspace.Workspace{
        Name: "Integration Test",
        Plan: workspace.PlanPro,
    }
    err := s.app.WorkspaceService.Create(ctx, ws)
    s.NoError(err)
    s.NotEmpty(ws.ID)
    
    // Verify vCluster created
    vcluster, err := s.testK8s.GetVCluster(ws.ID)
    s.NoError(err)
    s.Equal("provisioning", vcluster.Status)
    
    // Simulate vCluster ready
    s.testK8s.UpdateVClusterStatus(ws.ID, "ready")
    
    // Create project in workspace
    project := &project.Project{
        WorkspaceID: ws.ID,
        Name:        "Test Project",
    }
    err = s.app.ProjectService.Create(ctx, project)
    s.NoError(err)
    
    // Verify namespace created in vCluster
    ns, err := s.testK8s.GetNamespace(ws.ID, project.ID)
    s.NoError(err)
    s.Equal(project.Name, ns.Labels["project-name"])
    
    // Delete workspace
    err = s.app.WorkspaceService.Delete(ctx, ws.ID)
    s.NoError(err)
    
    // Verify cleanup
    _, err = s.testK8s.GetVCluster(ws.ID)
    s.Error(err) // Should not exist
}

func TestWorkspaceIntegrationSuite(t *testing.T) {
    suite.Run(t, new(WorkspaceIntegrationSuite))
}
```

### Database Testing

```go
// internal/testutil/database.go
package testutil

import (
    "fmt"
    "testing"
    
    "github.com/ory/dockertest/v3"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type TestDatabase struct {
    DB       *gorm.DB
    pool     *dockertest.Pool
    resource *dockertest.Resource
}

func NewTestDatabase() *TestDatabase {
    pool, err := dockertest.NewPool("")
    if err != nil {
        panic(err)
    }
    
    resource, err := pool.Run("postgres", "14", []string{
        "POSTGRES_PASSWORD=test",
        "POSTGRES_DB=test",
    })
    if err != nil {
        panic(err)
    }
    
    dsn := fmt.Sprintf("host=localhost port=%s user=postgres password=test dbname=test sslmode=disable",
        resource.GetPort("5432/tcp"))
    
    var db *gorm.DB
    if err := pool.Retry(func() error {
        var err error
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
        return err
    }); err != nil {
        panic(err)
    }
    
    return &TestDatabase{
        DB:       db,
        pool:     pool,
        resource: resource,
    }
}

func (td *TestDatabase) Cleanup() {
    td.pool.Purge(td.resource)
}
```

## Frontend Testing

### Component Tests

```typescript
// components/__tests__/workspace-list.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { WorkspaceList } from '../workspace-list';
import { server } from '@/mocks/server';
import { rest } from 'msw';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
  },
});

const wrapper = ({ children }) => (
  <QueryClientProvider client={queryClient}>
    {children}
  </QueryClientProvider>
);

describe('WorkspaceList', () => {
  it('displays workspaces', async () => {
    render(<WorkspaceList />, { wrapper });
    
    // Wait for loading to complete
    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });
    
    // Check workspaces are displayed
    expect(screen.getByText('Production Workspace')).toBeInTheDocument();
    expect(screen.getByText('Development Workspace')).toBeInTheDocument();
  });
  
  it('handles empty state', async () => {
    server.use(
      rest.get('/api/v1/workspaces', (req, res, ctx) => {
        return res(ctx.json({ data: [] }));
      })
    );
    
    render(<WorkspaceList />, { wrapper });
    
    await waitFor(() => {
      expect(screen.getByText('No workspaces found')).toBeInTheDocument();
    });
  });
  
  it('handles errors gracefully', async () => {
    server.use(
      rest.get('/api/v1/workspaces', (req, res, ctx) => {
        return res(ctx.status(500), ctx.json({ error: 'Server error' }));
      })
    );
    
    render(<WorkspaceList />, { wrapper });
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load workspaces')).toBeInTheDocument();
    });
  });
});
```

### Hook Tests

```typescript
// hooks/__tests__/use-workspace.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { useWorkspace } from '../use-workspace';
import { wrapper } from '@/test-utils';

describe('useWorkspace', () => {
  it('fetches workspace data', async () => {
    const { result } = renderHook(() => useWorkspace('ws-123'), { wrapper });
    
    // Initially loading
    expect(result.current.isLoading).toBe(true);
    expect(result.current.data).toBeUndefined();
    
    // Wait for data
    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });
    
    // Check data
    expect(result.current.data).toEqual({
      id: 'ws-123',
      name: 'Test Workspace',
      plan: 'pro',
    });
  });
  
  it('handles workspace not found', async () => {
    const { result } = renderHook(() => useWorkspace('invalid'), { wrapper });
    
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    
    expect(result.current.error?.message).toBe('Workspace not found');
  });
});
```

### E2E Tests with Playwright

```typescript
// tests/e2e/workspace-management.spec.ts
import { test, expect } from '@playwright/test';
import { mockAPI } from './helpers/mock-api';

test.describe('Workspace Management', () => {
  test.beforeEach(async ({ page }) => {
    await mockAPI(page);
    await page.goto('/login');
    await page.fill('[name=email]', 'test@example.com');
    await page.fill('[name=password]', 'password');
    await page.click('button[type=submit]');
    await page.waitForURL('/dashboard');
  });
  
  test('create new workspace', async ({ page }) => {
    // Navigate to workspaces
    await page.click('nav >> text=Workspaces');
    await page.waitForURL('/workspaces');
    
    // Click create button
    await page.click('button:has-text("Create Workspace")');
    
    // Fill form
    await page.fill('[name=name]', 'E2E Test Workspace');
    await page.fill('[name=description]', 'Created by Playwright test');
    await page.selectOption('[name=plan]', 'pro');
    
    // Submit
    await page.click('button[type=submit]');
    
    // Verify creation
    await expect(page.locator('text=E2E Test Workspace')).toBeVisible();
    await expect(page.locator('text=Pro Plan')).toBeVisible();
    
    // Take screenshot for visual regression
    await page.screenshot({ path: 'tests/screenshots/workspace-created.png' });
  });
  
  test('delete workspace', async ({ page }) => {
    await page.goto('/workspaces');
    
    // Find workspace card
    const workspaceCard = page.locator('[data-testid=workspace-card]', {
      hasText: 'Test Workspace'
    });
    
    // Open menu
    await workspaceCard.locator('button[aria-label="Options"]').click();
    
    // Click delete
    await page.click('text=Delete Workspace');
    
    // Confirm in dialog
    await page.click('dialog >> button:has-text("Delete")');
    
    // Verify deletion
    await expect(workspaceCard).not.toBeVisible();
  });
});
```

## Test Data Management

### Test Fixtures

```go
// internal/testutil/fixtures/workspace.go
package fixtures

import (
    "github.com/hexabase/hexabase-ai/internal/domain/workspace"
)

func NewWorkspace(opts ...func(*workspace.Workspace)) *workspace.Workspace {
    ws := &workspace.Workspace{
        ID:     "ws-test-123",
        Name:   "Test Workspace",
        Plan:   workspace.PlanStarter,
        Status: workspace.StatusActive,
    }
    
    for _, opt := range opts {
        opt(ws)
    }
    
    return ws
}

func WithPlan(plan string) func(*workspace.Workspace) {
    return func(ws *workspace.Workspace) {
        ws.Plan = plan
    }
}

func WithStatus(status string) func(*workspace.Workspace) {
    return func(ws *workspace.Workspace) {
        ws.Status = status
    }
}
```

### Test Factories

```typescript
// test-utils/factories.ts
import { Factory } from 'fishery';
import { faker } from '@faker-js/faker';
import { Workspace, Project, User } from '@/types';

export const workspaceFactory = Factory.define<Workspace>(() => ({
  id: faker.string.uuid(),
  name: faker.company.name(),
  description: faker.company.catchPhrase(),
  plan: faker.helpers.arrayElement(['starter', 'pro', 'enterprise']),
  status: 'active',
  createdAt: faker.date.past().toISOString(),
  updatedAt: faker.date.recent().toISOString(),
}));

export const projectFactory = Factory.define<Project>(() => ({
  id: faker.string.uuid(),
  workspaceId: faker.string.uuid(),
  name: faker.commerce.productName(),
  description: faker.commerce.productDescription(),
  createdAt: faker.date.past().toISOString(),
  updatedAt: faker.date.recent().toISOString(),
}));

export const userFactory = Factory.define<User>(() => ({
  id: faker.string.uuid(),
  email: faker.internet.email(),
  name: faker.person.fullName(),
  role: faker.helpers.arrayElement(['admin', 'developer', 'viewer']),
  createdAt: faker.date.past().toISOString(),
}));
```

## Performance Testing

### Load Testing with k6

```javascript
// tests/performance/workspace-api.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up
    { duration: '1m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 100 }, // Spike to 100
    { duration: '1m', target: 100 },  // Stay at 100
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    errors: ['rate<0.1'],             // Error rate under 10%
  },
};

const BASE_URL = 'https://api.hexabase.ai';

export function setup() {
  // Login and get token
  const loginRes = http.post(`${BASE_URL}/auth/login`, {
    email: 'loadtest@example.com',
    password: 'loadtest123',
  });
  
  return { token: loginRes.json('token') };
}

export default function (data) {
  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };
  
  // List workspaces
  const listRes = http.get(`${BASE_URL}/api/v1/workspaces`, { headers });
  check(listRes, {
    'list status is 200': (r) => r.status === 200,
    'list duration < 500ms': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);
  
  // Create workspace
  const createRes = http.post(
    `${BASE_URL}/api/v1/workspaces`,
    JSON.stringify({
      name: `Load Test ${Date.now()}`,
      plan: 'starter',
    }),
    { headers }
  );
  
  check(createRes, {
    'create status is 201': (r) => r.status === 201,
    'create duration < 1000ms': (r) => r.timings.duration < 1000,
  }) || errorRate.add(1);
  
  sleep(1);
}
```

## Security Testing

### Authentication Tests

```go
func TestAuthMiddleware_ValidToken(t *testing.T) {
    // Create valid JWT
    token, err := auth.GenerateToken("user-123", "admin")
    assert.NoError(t, err)
    
    // Create request with token
    req := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    // Test middleware
    w := httptest.NewRecorder()
    router := gin.New()
    router.Use(AuthMiddleware())
    router.GET("/api/v1/workspaces", func(c *gin.Context) {
        userID, _ := c.Get("user_id")
        c.JSON(200, gin.H{"user_id": userID})
    })
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "user-123")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
    testCases := []struct {
        name   string
        token  string
        status int
    }{
        {"missing token", "", 401},
        {"invalid format", "invalid", 401},
        {"expired token", generateExpiredToken(), 401},
        {"wrong signature", "Bearer wrong.token.here", 401},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
            if tc.token != "" {
                req.Header.Set("Authorization", tc.token)
            }
            
            w := httptest.NewRecorder()
            router := gin.New()
            router.Use(AuthMiddleware())
            router.GET("/api/v1/workspaces", func(c *gin.Context) {
                c.JSON(200, gin.H{})
            })
            
            router.ServeHTTP(w, req)
            assert.Equal(t, tc.status, w.Code)
        })
    }
}
```

## Test Coverage

### Running Coverage

```bash
# Backend coverage
cd api
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Frontend coverage
cd ui
npm run test:coverage
```

### Coverage Requirements

- Overall: 80% minimum
- Critical paths: 90% minimum
- New code: 85% minimum

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run tests
        run: |
          cd api
          go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./api/coverage.out
  
  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: |
          cd ui
          npm ci
      
      - name: Run tests
        run: |
          cd ui
          npm run test:ci
      
      - name: Run E2E tests
        run: |
          cd ui
          npx playwright install
          npm run test:e2e
```

## Best Practices

1. **Test Naming**: Use descriptive test names that explain what is being tested
2. **Test Independence**: Each test should be independent and not rely on others
3. **Mock External Dependencies**: Use mocks for databases, APIs, and external services
4. **Test Data**: Use factories and fixtures for consistent test data
5. **Assertions**: Make specific assertions, avoid generic checks
6. **Coverage**: Aim for high coverage but focus on critical paths
7. **Performance**: Keep tests fast, parallelize where possible
8. **Maintenance**: Refactor tests alongside code changes
9. **Documentation**: Document complex test scenarios
10. **CI/CD**: Run tests automatically on every commit

## Troubleshooting

### Common Issues

1. **Flaky Tests**
   - Add proper waits and retries
   - Ensure proper test isolation
   - Mock time-dependent operations

2. **Slow Tests**
   - Use test database containers
   - Parallelize test execution
   - Mock expensive operations

3. **Coverage Gaps**
   - Focus on untested critical paths
   - Add tests for error scenarios
   - Test edge cases

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Playwright Documentation](https://playwright.dev/)
- [k6 Documentation](https://k6.io/docs/)