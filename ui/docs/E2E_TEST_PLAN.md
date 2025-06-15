# E2E Test Plan with Playwright

## Overview
This document outlines a comprehensive E2E testing strategy using Playwright with mock data. The tests will cover real user workflows while maintaining test isolation and reliability.

## Test Infrastructure Setup

### 1. Mock Server Strategy
- Use Playwright's route interception to mock API responses
- Create a mock data layer that simulates the backend API
- Implement realistic delays and error scenarios
- Support state transitions (e.g., deployment progress)

### 2. Test Data Management
```typescript
// test-data/fixtures.ts
export const testOrganization = {
  id: 'test-org-1',
  name: 'E2E Test Organization',
  // ... other fields
};

export const testWorkspace = {
  id: 'test-ws-1',
  name: 'E2E Test Workspace',
  plan_id: 'dedicated',
  vcluster_status: 'active',
  // ... other fields
};
```

### 3. Page Object Model (POM)
Create page objects for maintainable tests:
```typescript
// e2e/pages/LoginPage.ts
export class LoginPage {
  constructor(private page: Page) {}
  
  async login(email: string, password: string) {
    await this.page.fill('[data-testid="email-input"]', email);
    await this.page.fill('[data-testid="password-input"]', password);
    await this.page.click('[data-testid="login-button"]');
  }
}
```

## Test Suites

### Phase 1: Core User Flows (Week 1)

#### 1.1 Authentication Flow
```typescript
// e2e/tests/auth.spec.ts
test.describe('Authentication', () => {
  test('successful login flow', async ({ page }) => {
    // Test login with valid credentials
    // Verify redirect to dashboard
    // Check user session
  });
  
  test('handles authentication errors', async ({ page }) => {
    // Test invalid credentials
    // Test network errors
    // Verify error messages
  });
  
  test('logout flow', async ({ page }) => {
    // Test logout
    // Verify session cleanup
    // Redirect to login
  });
});
```

#### 1.2 Organization & Workspace Management
```typescript
// e2e/tests/organization.spec.ts
test.describe('Organization Management', () => {
  test('create new organization', async ({ page }) => {
    // Fill organization form
    // Submit and verify creation
    // Check navigation
  });
  
  test('create workspace with different plans', async ({ page }) => {
    // Test shared plan creation
    // Test dedicated plan with node selection
    // Verify vCluster provisioning simulation
  });
  
  test('switch between workspaces', async ({ page }) => {
    // Create multiple workspaces
    // Test workspace switching
    // Verify context changes
  });
});
```

#### 1.3 Project & Resource Management
```typescript
// e2e/tests/projects.spec.ts
test.describe('Project Management', () => {
  test('create project with quotas', async ({ page }) => {
    // Create project
    // Set resource quotas
    // Verify namespace creation
  });
  
  test('deploy applications', async ({ page }) => {
    // Deploy stateless app
    // Deploy stateful app with storage
    // Monitor deployment progress
    // Verify running status
  });
});
```

### Phase 2: Advanced Features (Week 2)

#### 2.1 CI/CD Pipeline
```typescript
// e2e/tests/cicd.spec.ts
test.describe('CI/CD Integration', () => {
  test('setup GitHub integration', async ({ page }) => {
    // Configure repository
    // Set build commands
    // Configure deployment strategy
  });
  
  test('trigger and monitor deployment', async ({ page }) => {
    // Trigger deployment
    // Watch build progress
    // Monitor rolling update
    // Verify deployment success
  });
  
  test('rollback deployment', async ({ page }) => {
    // Simulate failed deployment
    // Initiate rollback
    // Select previous version
    // Verify rollback completion
  });
});
```

#### 2.2 Backup & Restore (Dedicated Plan)
```typescript
// e2e/tests/backup.spec.ts
test.describe('Backup Management', () => {
  test.beforeEach(async ({ page }) => {
    // Ensure dedicated workspace context
  });
  
  test('configure backup storage', async ({ page }) => {
    // Add Proxmox storage
    // Configure S3 storage
    // Verify storage status
  });
  
  test('create and execute backup policy', async ({ page }) => {
    // Create backup policy
    // Set schedule and retention
    // Trigger manual backup
    // Monitor execution
  });
  
  test('restore from backup', async ({ page }) => {
    // Select backup
    // Configure restore options
    // Execute restore
    // Verify data integrity
  });
});
```

#### 2.3 Serverless Functions
```typescript
// e2e/tests/functions.spec.ts
test.describe('Function Management', () => {
  test('create and deploy function', async ({ page }) => {
    // Create function
    // Write code in editor
    // Deploy function
    // Test execution
  });
  
  test('function versioning', async ({ page }) => {
    // Deploy new version
    // View version history
    // Rollback to previous version
  });
  
  test('configure event triggers', async ({ page }) => {
    // Add HTTP trigger
    // Configure S3 event
    // Test event processing
  });
});
```

### Phase 3: Monitoring & AI Features (Week 3)

#### 3.1 Monitoring & Alerts
```typescript
// e2e/tests/monitoring.spec.ts
test.describe('Monitoring', () => {
  test('view resource metrics', async ({ page }) => {
    // Navigate to monitoring
    // Check CPU/Memory graphs
    // Verify metric updates
  });
  
  test('configure alerts', async ({ page }) => {
    // Create alert rule
    // Set thresholds
    // Test alert triggering
  });
});
```

#### 3.2 AI Chat Integration
```typescript
// e2e/tests/ai-chat.spec.ts
test.describe('AI Assistant', () => {
  test('optimization suggestions', async ({ page }) => {
    // Open AI chat
    // Ask for optimization help
    // Review suggestions
    // Apply recommendations
  });
  
  test('debugging assistance', async ({ page }) => {
    // Simulate application error
    // Ask AI for help
    // Follow AI guidance
    // Verify issue resolution
  });
  
  test('code generation', async ({ page }) => {
    // Request function code
    // Review generated code
    // Deploy from AI suggestion
  });
});
```

### Phase 4: Edge Cases & Error Handling (Week 4)

#### 4.1 Error Scenarios
```typescript
// e2e/tests/error-handling.spec.ts
test.describe('Error Handling', () => {
  test('handles API failures gracefully', async ({ page }) => {
    // Simulate API errors
    // Verify error messages
    // Test retry mechanisms
  });
  
  test('handles resource limits', async ({ page }) => {
    // Exceed quotas
    // Verify error handling
    // Test limit warnings
  });
});
```

#### 4.2 Performance & Load
```typescript
// e2e/tests/performance.spec.ts
test.describe('Performance', () => {
  test('handles large resource lists', async ({ page }) => {
    // Mock 100+ applications
    // Test pagination
    // Verify scroll performance
  });
  
  test('concurrent operations', async ({ page }) => {
    // Deploy multiple apps
    // Test UI responsiveness
    // Verify status updates
  });
});
```

## Implementation Guidelines

### 1. Test Structure
```
e2e/
├── fixtures/
│   ├── mock-data.ts
│   ├── api-mocks.ts
│   └── test-users.ts
├── pages/
│   ├── LoginPage.ts
│   ├── DashboardPage.ts
│   ├── WorkspacePage.ts
│   └── ...
├── utils/
│   ├── mock-server.ts
│   ├── test-helpers.ts
│   └── wait-utils.ts
└── tests/
    ├── auth.spec.ts
    ├── organization.spec.ts
    ├── projects.spec.ts
    └── ...
```

### 2. Mock API Setup
```typescript
// e2e/utils/mock-server.ts
export async function setupMockAPI(page: Page) {
  // Mock login
  await page.route('**/api/auth/login', async route => {
    await route.fulfill({
      status: 200,
      body: JSON.stringify({
        access_token: 'mock-token',
        user: testUser
      })
    });
  });
  
  // Mock organizations list
  await page.route('**/api/organizations', async route => {
    await route.fulfill({
      status: 200,
      body: JSON.stringify({
        organizations: [testOrganization],
        total: 1
      })
    });
  });
  
  // Add more mocks...
}
```

### 3. Test Utilities
```typescript
// e2e/utils/test-helpers.ts
export async function waitForDeployment(page: Page, appName: string) {
  // Wait for pending status
  await page.waitForSelector(`[data-testid="${appName}-status"]:has-text("pending")`);
  
  // Simulate deployment progress
  await page.waitForTimeout(1000);
  
  // Wait for running status
  await page.waitForSelector(`[data-testid="${appName}-status"]:has-text("running")`);
}

export async function mockProgressiveResponse(page: Page, url: string, stages: any[]) {
  let stageIndex = 0;
  await page.route(url, async route => {
    const response = stages[Math.min(stageIndex++, stages.length - 1)];
    await route.fulfill({
      status: 200,
      body: JSON.stringify(response)
    });
  });
}
```

### 4. Configuration
```typescript
// playwright.config.ts
export default defineConfig({
  testDir: './e2e/tests',
  timeout: 30000,
  retries: 2,
  use: {
    baseURL: 'http://localhost:3000',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'mobile',
      use: { ...devices['iPhone 12'] },
    },
  ],
});
```

## Testing Best Practices

### 1. Test Isolation
- Each test should be independent
- Use fresh test data for each test
- Clean up resources after tests
- Don't rely on test execution order

### 2. Reliable Selectors
- Use data-testid attributes
- Avoid brittle CSS selectors
- Use role-based queries when possible
- Implement proper wait strategies

### 3. Mock Data Realism
- Use realistic delays for async operations
- Simulate actual API response structures
- Include error scenarios
- Support state transitions

### 4. Debugging Support
- Take screenshots on failure
- Record videos for complex flows
- Use Playwright trace viewer
- Add descriptive test names

### 5. Performance Considerations
- Run tests in parallel when possible
- Use page.waitForLoadState() appropriately
- Minimize unnecessary waits
- Cache static mock responses

## Continuous Integration

### GitHub Actions Setup
```yaml
name: E2E Tests
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm ci
      - run: npx playwright install
      - run: npm run test:e2e
      - uses: actions/upload-artifact@v3
        if: failure()
        with:
          name: test-results
          path: test-results/
```

## Success Metrics

1. **Coverage**: 90%+ of critical user paths
2. **Reliability**: <5% flaky test rate
3. **Performance**: Tests complete in <10 minutes
4. **Maintainability**: <2 hours to update tests for UI changes

## Timeline

- **Week 1**: Core flows (auth, org, workspace, projects)
- **Week 2**: Advanced features (CI/CD, backup, functions)
- **Week 3**: Monitoring, AI, and integrations
- **Week 4**: Error handling, performance, and polish

## Next Steps

1. Set up Playwright project structure
2. Implement mock API layer
3. Create page objects for main UI components
4. Write first test suite (authentication)
5. Establish CI/CD pipeline
6. Document test writing guidelines for team