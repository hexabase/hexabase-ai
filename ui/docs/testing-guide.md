# UI Testing Guide

This guide covers testing approaches for the Hexabase AI UI, including unit tests and end-to-end (E2E) tests.

## Table of Contents

1. [Testing Stack](#testing-stack)
2. [Unit Testing](#unit-testing)
3. [E2E Testing](#e2e-testing)
4. [Mock API Setup](#mock-api-setup)
5. [Coverage Goals](#coverage-goals)
6. [Best Practices](#best-practices)

## Testing Stack

- **Unit Testing**: Jest + React Testing Library
- **E2E Testing**: Playwright
- **Mock API**: MSW (Mock Service Worker) + Custom Mock Client
- **Test Utilities**: Custom test utilities and factories

## Unit Testing

### Running Unit Tests

```bash
# Run all unit tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm test -- --coverage

# Run specific test file
npm test -- organization-list.test.tsx
```

### Writing Unit Tests

#### 1. Component Testing Pattern

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { OrganizationList } from '@/components/organizations/organization-list';
import { mockApiClient } from '@/test-utils/mock-api-client';

// Mock dependencies
jest.mock('@/lib/api-client', () => ({
  apiClient: mockApiClient
}));

describe('OrganizationList', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should display list of organizations', async () => {
    render(<OrganizationList />);
    
    await waitFor(() => {
      expect(screen.getByText('Acme Corporation')).toBeInTheDocument();
    });
  });
});
```

#### 2. Testing User Interactions

```typescript
it('should create new organization', async () => {
  render(<OrganizationList />);
  
  // Open dialog
  fireEvent.click(screen.getByRole('button', { name: /create organization/i }));
  
  // Fill form
  fireEvent.change(screen.getByLabelText(/organization name/i), {
    target: { value: 'New Org' }
  });
  
  // Submit
  fireEvent.click(screen.getByRole('button', { name: /create/i }));
  
  // Verify API call
  await waitFor(() => {
    expect(mockApiClient.organizations.create).toHaveBeenCalledWith({
      name: 'New Org'
    });
  });
});
```

#### 3. Testing Loading States

```typescript
it('should show loading skeleton', () => {
  // Mock API to never resolve
  mockApiClient.organizations.list.mockImplementation(
    () => new Promise(() => {})
  );
  
  render(<OrganizationList />);
  
  expect(screen.getByTestId('organizations-skeleton')).toBeInTheDocument();
});
```

#### 4. Testing Error States

```typescript
it('should handle API errors', async () => {
  mockApiClient.organizations.list.mockRejectedValue(
    new Error('Network error')
  );
  
  render(<OrganizationList />);
  
  await waitFor(() => {
    expect(screen.getByText(/failed to load/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
  });
});
```

### Testing Hooks

```typescript
import { renderHook, act } from '@testing-library/react';
import { useWebSocket } from '@/hooks/use-websocket';

describe('useWebSocket', () => {
  it('should connect to websocket', () => {
    const { result } = renderHook(() => useWebSocket('/api/ws'));
    
    expect(result.current.isConnected).toBe(false);
    
    act(() => {
      result.current.connect();
    });
    
    expect(result.current.isConnected).toBe(true);
  });
});
```

### Testing Context Providers

```typescript
import { render, screen } from '@testing-library/react';
import { AuthProvider, useAuth } from '@/lib/auth-context';

const TestComponent = () => {
  const { user } = useAuth();
  return <div>{user?.email || 'Not logged in'}</div>;
};

describe('AuthContext', () => {
  it('should provide user data', () => {
    render(
      <AuthProvider initialUser={{ email: 'test@example.com' }}>
        <TestComponent />
      </AuthProvider>
    );
    
    expect(screen.getByText('test@example.com')).toBeInTheDocument();
  });
});
```

## E2E Testing

### Setup Playwright

```bash
# Install Playwright
npm install -D @playwright/test

# Install browsers
npx playwright install

# Run E2E tests
npm run test:e2e

# Run in UI mode
npm run test:e2e:ui

# Run specific test
npm run test:e2e -- workspace-flow.spec.ts
```

### Writing E2E Tests

#### 1. Basic E2E Test Structure

```typescript
import { test, expect } from '@playwright/test';

test.describe('Organization Management', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to app
    await page.goto('http://localhost:3000');
    
    // Login if needed
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="password-input"]', 'password');
    await page.click('[data-testid="login-button"]');
    
    // Wait for dashboard
    await page.waitForURL('**/dashboard');
  });
  
  test('should create new organization', async ({ page }) => {
    // Click create button
    await page.click('button:has-text("Create Organization")');
    
    // Fill form
    await page.fill('[data-testid="org-name-input"]', 'Test Organization');
    
    // Submit
    await page.click('button:has-text("Create")');
    
    // Verify success
    await expect(page.locator('text=Test Organization')).toBeVisible();
  });
});
```

#### 2. Testing Complex Workflows

```typescript
test('complete workspace setup flow', async ({ page }) => {
  // Create organization
  await page.click('button:has-text("Create Organization")');
  await page.fill('[data-testid="org-name-input"]', 'Test Org');
  await page.click('button:has-text("Create")');
  
  // Navigate to organization
  await page.click('text=Test Org');
  
  // Create workspace
  await page.click('button:has-text("Create Workspace")');
  await page.fill('[data-testid="workspace-name-input"]', 'Production');
  await page.selectOption('[data-testid="plan-select"]', 'dedicated');
  await page.click('button:has-text("Create")');
  
  // Wait for provisioning
  await expect(page.locator('text=active')).toBeVisible({ timeout: 30000 });
  
  // Create project
  await page.click('text=Production');
  await page.click('button:has-text("Create Project")');
  await page.fill('[data-testid="project-name-input"]', 'frontend-app');
  await page.click('button:has-text("Create")');
  
  // Verify complete setup
  await expect(page.locator('text=frontend-app')).toBeVisible();
});
```

#### 3. Testing Real-time Updates

```typescript
test('should show real-time status updates', async ({ page, context }) => {
  // Open two tabs
  const page2 = await context.newPage();
  
  // Navigate both to same workspace
  await page.goto('http://localhost:3000/dashboard/organizations/org-1/workspaces');
  await page2.goto('http://localhost:3000/dashboard/organizations/org-1/workspaces');
  
  // Create workspace in first tab
  await page.click('button:has-text("Create Workspace")');
  await page.fill('[data-testid="workspace-name-input"]', 'Realtime Test');
  await page.click('button:has-text("Create")');
  
  // Verify appears in second tab
  await expect(page2.locator('text=Realtime Test')).toBeVisible();
  
  // Verify status updates in both
  await expect(page.locator('text=creating')).toBeVisible();
  await expect(page2.locator('text=creating')).toBeVisible();
  
  await expect(page.locator('text=active')).toBeVisible({ timeout: 10000 });
  await expect(page2.locator('text=active')).toBeVisible({ timeout: 10000 });
});
```

## Mock API Setup

### Using Mock Service Worker (MSW)

```typescript
// src/mocks/handlers/organizations.ts
import { http, HttpResponse } from 'msw';

export const organizationHandlers = [
  http.get('/api/v1/organizations', () => {
    return HttpResponse.json({
      organizations: [
        { id: 'org-1', name: 'Acme Corp' }
      ],
      total: 1
    });
  }),
  
  http.post('/api/v1/organizations', async ({ request }) => {
    const body = await request.json();
    return HttpResponse.json({
      id: `org-${Date.now()}`,
      name: body.name,
      created_at: new Date().toISOString()
    });
  })
];
```

### Using Custom Mock Client

```typescript
// In tests
import { mockApiClient } from '@/test-utils/mock-api-client';

beforeEach(() => {
  // Reset mock data
  mockApiClient.organizations.list.mockResolvedValue({
    data: {
      organizations: mockOrganizations,
      total: mockOrganizations.length
    }
  });
});
```

## Coverage Goals

### Target Coverage Metrics

- **Statements**: 90%+
- **Branches**: 85%+
- **Functions**: 90%+
- **Lines**: 90%+

### Checking Coverage

```bash
# Generate coverage report
npm test -- --coverage

# View HTML report
open coverage/lcov-report/index.html

# Check specific component coverage
npm test -- --coverage --collectCoverageFrom='src/components/organizations/**'
```

### Coverage Configuration

```javascript
// jest.config.ts
export default {
  coverageThreshold: {
    global: {
      statements: 90,
      branches: 85,
      functions: 90,
      lines: 90
    }
  },
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/*.stories.tsx',
    '!src/test-utils/**'
  ]
};
```

## Best Practices

### 1. Test Organization

```
src/
├── components/
│   ├── organizations/
│   │   ├── organization-list.tsx
│   │   └── __tests__/
│   │       └── organization-list.test.tsx
├── tests/           # E2E tests
│   ├── auth.spec.ts
│   └── organization-flow.spec.ts
└── test-utils/      # Shared test utilities
    ├── index.tsx
    └── mock-api-client.ts
```

### 2. Test Data Factories

```typescript
// src/test-utils/factories.ts
export const createMockOrganization = (overrides = {}) => ({
  id: `org-${Date.now()}`,
  name: 'Test Organization',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  ...overrides
});
```

### 3. Custom Render Function

```typescript
// src/test-utils/index.tsx
import { render } from '@testing-library/react';
import { AuthProvider } from '@/lib/auth-context';

export const renderWithProviders = (ui: React.ReactElement, options = {}) => {
  return render(
    <AuthProvider>
      {ui}
    </AuthProvider>,
    options
  );
};
```

### 4. Async Testing Patterns

```typescript
// Wait for element
await waitFor(() => {
  expect(screen.getByText('Success')).toBeInTheDocument();
});

// Wait for element removal
await waitFor(() => {
  expect(screen.queryByText('Loading')).not.toBeInTheDocument();
});

// Multiple assertions
await waitFor(() => {
  expect(mockApiClient.organizations.create).toHaveBeenCalled();
  expect(screen.getByText('Created')).toBeInTheDocument();
});
```

### 5. Debugging Tests

```typescript
// Debug DOM
screen.debug();

// Debug specific element
screen.debug(screen.getByRole('button'));

// Log accessible roles
screen.logTestingPlaygroundURL();

// Use testing-library queries
const button = screen.getByRole('button', { name: /submit/i });
```

### 6. Performance Testing

```typescript
test('should render large lists efficiently', async () => {
  const largeDataset = Array.from({ length: 1000 }, (_, i) => 
    createMockOrganization({ id: `org-${i}`, name: `Org ${i}` })
  );
  
  const startTime = performance.now();
  render(<OrganizationList data={largeDataset} />);
  const endTime = performance.now();
  
  expect(endTime - startTime).toBeLessThan(1000); // Under 1 second
});
```

## Common Testing Scenarios

### 1. Form Validation

```typescript
test('should validate required fields', async () => {
  render(<CreateOrganizationDialog />);
  
  // Submit without filling
  fireEvent.click(screen.getByRole('button', { name: /create/i }));
  
  // Check validation message
  await waitFor(() => {
    expect(screen.getByText(/name is required/i)).toBeInTheDocument();
  });
});
```

### 2. Permission-based UI

```typescript
test('should hide admin features for non-admin users', () => {
  mockApiClient.auth.me.mockResolvedValue({
    user: { role: 'member' }
  });
  
  render(<OrganizationSettings />);
  
  expect(screen.queryByRole('button', { name: /delete/i })).not.toBeInTheDocument();
});
```

### 3. Optimistic Updates

```typescript
test('should show optimistic update', async () => {
  render(<WorkspaceList />);
  
  // Change status
  fireEvent.click(screen.getByRole('switch', { name: /enable/i }));
  
  // Should update immediately
  expect(screen.getByText('active')).toBeInTheDocument();
  
  // Verify API call
  await waitFor(() => {
    expect(mockApiClient.workspaces.update).toHaveBeenCalled();
  });
});
```

## Troubleshooting

### Common Issues

1. **Act warnings**: Wrap state updates in `act()` or use `waitFor()`
2. **Query failures**: Use `screen.debug()` to inspect DOM
3. **Async issues**: Increase timeout in `waitFor({ timeout: 5000 })`
4. **Mock not working**: Check import paths and jest.mock placement

### Resources

- [React Testing Library Docs](https://testing-library.com/docs/react-testing-library/intro/)
- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [Playwright Documentation](https://playwright.dev/docs/intro)
- [MSW Documentation](https://mswjs.io/docs/)