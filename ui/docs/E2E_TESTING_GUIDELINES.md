# E2E Testing Guidelines for Hexabase AI

This guide provides best practices and standards for writing E2E tests for the Hexabase AI platform.

## Table of Contents

1. [Overview](#overview)
2. [Test Structure](#test-structure)
3. [Writing Tests](#writing-tests)
4. [Page Object Model](#page-object-model)
5. [Test Data Management](#test-data-management)
6. [Best Practices](#best-practices)
7. [Common Patterns](#common-patterns)
8. [Debugging Tests](#debugging-tests)
9. [CI/CD Integration](#cicd-integration)

## Overview

Our E2E tests use Playwright with TypeScript to ensure the Hexabase AI platform works correctly from a user's perspective. Tests should be:

- **Reliable**: No flaky tests
- **Maintainable**: Easy to update when UI changes
- **Fast**: Run in parallel when possible
- **Comprehensive**: Cover critical user journeys

## Test Structure

### Directory Organization

```
e2e/
├── tests/              # Test specifications
├── pages/              # Page Object Models
├── fixtures/           # Test data and scenarios
├── utils/              # Helper functions
└── screenshots/        # Test results
```

### Test File Naming

- Test files: `feature-name.spec.ts`
- Page objects: `FeatureNamePage.ts`
- Use kebab-case for test files
- Use PascalCase for Page Object classes

## Writing Tests

### Basic Test Structure

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers } from '../fixtures/mock-data';

test.describe('Feature Name', () => {
  let loginPage: LoginPage;
  
  test.beforeEach(async ({ page }) => {
    // Setup mock API if needed
    await setupMockAPI(page);
    
    // Initialize page objects
    loginPage = new LoginPage(page);
  });
  
  test('should perform expected action', async ({ page }) => {
    // Arrange
    await loginPage.goto();
    
    // Act
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    
    // Assert
    await expect(page).toHaveURL('/dashboard');
    await loginPage.isLoggedIn();
  });
});
```

### Test Descriptions

Use clear, descriptive test names that explain what is being tested:

```typescript
// ✅ Good
test('should display error when login credentials are invalid', async () => {});
test('should create new organization with professional plan', async () => {});

// ❌ Bad
test('test login', async () => {});
test('org test', async () => {});
```

### Using Test Tags

Tag tests for better organization and selective execution:

```typescript
import { SMOKE_TAG, CRITICAL_TAG, FLAKY_TAG } from '../utils/test-tags';

test(`critical user journey ${CRITICAL_TAG}`, async () => {});
test(`basic smoke test ${SMOKE_TAG}`, async () => {});
test(`known flaky test ${FLAKY_TAG}`, async () => {});
```

## Page Object Model

### Creating Page Objects

Each page or component should have its own Page Object:

```typescript
import { Page, Locator } from '@playwright/test';

export class ApplicationPage {
  readonly page: Page;
  readonly deployButton: Locator;
  readonly appNameInput: Locator;
  readonly statusBadge: Locator;
  
  constructor(page: Page) {
    this.page = page;
    this.deployButton = page.getByTestId('deploy-application-button');
    this.appNameInput = page.getByTestId('app-name-input');
    this.statusBadge = page.getByTestId('app-status');
  }
  
  async deployApplication(config: DeploymentConfig) {
    await this.deployButton.click();
    
    const dialog = this.page.getByRole('dialog');
    await dialog.getByTestId('app-name-input').fill(config.name);
    await dialog.getByTestId('app-image-input').fill(config.image);
    await dialog.getByTestId('deploy-button').click();
  }
  
  async waitForStatus(appName: string, status: string) {
    const appCard = this.page.getByTestId(`app-card-${appName}`);
    await expect(appCard.getByTestId('status')).toHaveText(status);
  }
}
```

### Locator Best Practices

1. **Use data-testid attributes** for reliable element selection:
   ```typescript
   page.getByTestId('submit-button')
   ```

2. **Use semantic locators** when appropriate:
   ```typescript
   page.getByRole('button', { name: 'Submit' })
   page.getByLabel('Email address')
   page.getByPlaceholder('Enter your email')
   ```

3. **Avoid CSS/XPath selectors** unless absolutely necessary

4. **Chain locators** for scoped searches:
   ```typescript
   const dialog = page.getByRole('dialog');
   await dialog.getByTestId('confirm-button').click();
   ```

## Test Data Management

### Using Test Data Generators

```typescript
import { ApplicationGenerator } from '../fixtures/generators/application-generator';

test('deploy custom application', async ({ page }) => {
  const appGenerator = new ApplicationGenerator();
  
  // Generate random app
  const app = appGenerator.generate();
  
  // Generate app with specific traits
  const haApp = appGenerator.withTraits(['highAvailability']).generate();
  
  // Use builder for precise control
  const customApp = new ApplicationBuilder()
    .withName('my-api')
    .withImage('node', '18')
    .withReplicas(3)
    .withAutoscaling(1, 10, 70)
    .build();
});
```

### Using Test Scenarios

```typescript
import { TestDataManager } from '../utils/test-data-manager';

test('enterprise scenario', async ({ page }) => {
  const testData = new TestDataManager(page);
  
  // Load pre-built scenario
  await testData.loadScenario('enterprise', 12345); // with seed for reproducibility
  
  // Get test credentials
  const admin = testData.getAdminCredentials();
  
  // Find specific entities
  const workspace = testData.findWorkspaceByName('Production');
  const apps = testData.getApplications();
});
```

## Best Practices

### 1. Keep Tests Independent

Each test should be able to run in isolation:

```typescript
// ✅ Good - Self-contained test
test('should create and delete project', async ({ page }) => {
  const projectName = `test-project-${Date.now()}`;
  await projectPage.createProject(projectName);
  await projectPage.deleteProject(projectName);
});

// ❌ Bad - Depends on previous test state
test('should delete the project', async ({ page }) => {
  await projectPage.deleteProject('test-project'); // Assumes project exists
});
```

### 2. Use Explicit Waits

Wait for specific conditions rather than arbitrary timeouts:

```typescript
// ✅ Good
await page.waitForSelector('[data-testid="loading"]', { state: 'hidden' });
await expect(page.getByTestId('status')).toHaveText('Ready');

// ❌ Bad
await page.waitForTimeout(5000);
```

### 3. Handle Async Operations

Always handle loading states and async operations:

```typescript
test('should deploy application', async ({ page }) => {
  await projectPage.deployApplication(appConfig);
  
  // Wait for deployment to start
  await expect(page.getByTestId('deployment-status')).toHaveText('Deploying');
  
  // Wait for completion (with timeout)
  await expect(page.getByTestId('deployment-status')).toHaveText('Running', {
    timeout: 60000 // 1 minute for deployment
  });
});
```

### 4. Clean Up After Tests

Always clean up resources created during tests:

```typescript
test('should manage workspace lifecycle', async ({ page }) => {
  const workspaceName = `test-ws-${Date.now()}`;
  
  try {
    // Create workspace
    await workspacePage.createWorkspace(workspaceName);
    
    // Test workspace features
    await workspacePage.configureQuotas(workspaceName, { cpu: '8', memory: '16Gi' });
    
  } finally {
    // Always clean up
    await workspacePage.deleteWorkspace(workspaceName);
  }
});
```

### 5. Use Meaningful Assertions

Write assertions that clearly express the expected behavior:

```typescript
// ✅ Good - Clear assertions
await expect(page.getByTestId('error-message')).toContainText('Invalid credentials');
await expect(page.getByTestId('user-count')).toHaveText('5 users');

// ❌ Bad - Unclear assertions
await expect(page.locator('.msg')).toBeVisible();
await expect(page.locator('div')).toHaveText('5');
```

## Common Patterns

### Testing Form Validation

```typescript
test('should validate application deployment form', async ({ page }) => {
  await projectPage.deployButton.click();
  
  const dialog = page.getByRole('dialog');
  
  // Submit empty form
  await dialog.getByTestId('deploy-button').click();
  
  // Check validation messages
  await expect(dialog.getByTestId('name-error')).toContainText('Name is required');
  await expect(dialog.getByTestId('image-error')).toContainText('Image is required');
  
  // Fill invalid data
  await dialog.getByTestId('app-name-input').fill('Invalid Name!');
  await dialog.getByTestId('app-port-input').fill('99999');
  
  // Check format validation
  await expect(dialog.getByTestId('name-error')).toContainText('Only alphanumeric');
  await expect(dialog.getByTestId('port-error')).toContainText('Port must be 1-65535');
});
```

### Testing Error Scenarios

```typescript
test('should handle network errors gracefully', async ({ page }) => {
  // Simulate network failure
  await page.route('**/api/**', route => route.abort());
  
  // Attempt action
  await projectPage.createProject('test-project');
  
  // Verify error handling
  await expect(page.getByTestId('error-banner')).toBeVisible();
  await expect(page.getByTestId('retry-button')).toBeVisible();
  
  // Restore network and retry
  await page.unroute('**/api/**');
  await page.getByTestId('retry-button').click();
  
  // Verify recovery
  await expect(page.getByTestId('success-message')).toBeVisible();
});
```

### Testing Real-time Updates

```typescript
test('should show real-time deployment progress', async ({ page }) => {
  // Start deployment
  await applicationPage.deployApplication(appConfig);
  
  // Monitor progress updates
  const progressBar = page.getByTestId('deployment-progress');
  
  // Initial state
  await expect(progressBar).toHaveAttribute('aria-valuenow', '0');
  
  // Progress updates
  await expect(progressBar).toHaveAttribute('aria-valuenow', '50', {
    timeout: 30000
  });
  
  // Completion
  await expect(progressBar).toHaveAttribute('aria-valuenow', '100', {
    timeout: 60000
  });
  
  // Status change
  await expect(page.getByTestId('app-status')).toHaveText('Running');
});
```

### Testing Complex Workflows

```typescript
test('should complete end-to-end application deployment', async ({ page }) => {
  // Step 1: Create project
  const projectName = `e2e-project-${Date.now()}`;
  await workspacePage.createProject(projectName);
  await workspacePage.openProject(projectName);
  
  // Step 2: Deploy application
  const appName = `nginx-app-${Date.now()}`;
  await projectPage.deployApplication({
    name: appName,
    image: 'nginx:latest',
    replicas: 2,
  });
  
  // Step 3: Wait for deployment
  await projectPage.waitForApplicationStatus(appName, 'running');
  
  // Step 4: Configure scaling
  await projectPage.openApplication(appName);
  await applicationPage.configureAutoscaling({
    enabled: true,
    minReplicas: 2,
    maxReplicas: 10,
    targetCPU: 70,
  });
  
  // Step 5: Verify metrics appear
  await applicationPage.metricsTab.click();
  await expect(page.getByTestId('cpu-chart')).toBeVisible();
  await expect(page.getByTestId('memory-chart')).toBeVisible();
});
```

## Debugging Tests

### Using Playwright Inspector

```bash
# Run with inspector
npx playwright test --debug

# Run specific test with inspector
npx playwright test auth.spec.ts --debug
```

### Capturing Screenshots on Failure

```typescript
test('should handle errors', async ({ page }, testInfo) => {
  try {
    await someAction();
  } catch (error) {
    // Capture screenshot on failure
    await testInfo.attach('failure-screenshot', {
      body: await page.screenshot(),
      contentType: 'image/png'
    });
    throw error;
  }
});
```

### Using Page Pause

```typescript
test('debug interaction', async ({ page }) => {
  await page.goto('/');
  await page.pause(); // Opens inspector and pauses execution
  await page.click('button');
});
```

### Verbose Logging

```typescript
test('with logging', async ({ page }) => {
  // Enable verbose logging
  await page.on('console', msg => console.log('PAGE LOG:', msg.text()));
  await page.on('pageerror', err => console.log('PAGE ERROR:', err));
  
  // Your test actions
  await page.goto('/dashboard');
});
```

## CI/CD Integration

### Running Tests in CI

```yaml
# .github/workflows/e2e-tests.yml
- name: Run E2E tests
  run: |
    npm ci
    npx playwright install --with-deps
    npm run test:e2e
  env:
    CI: true
```

### Test Parallelization

Configure parallel execution in `playwright.config.ts`:

```typescript
export default defineConfig({
  workers: process.env.CI ? 4 : undefined,
  fullyParallel: true,
  
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
  ],
});
```

### Test Reporting

```typescript
export default defineConfig({
  reporter: [
    ['html', { open: 'never' }],
    ['junit', { outputFile: 'test-results/junit.xml' }],
    ['json', { outputFile: 'test-results/results.json' }],
  ],
});
```

## Test Execution Commands

```bash
# Run all tests
npm run test:e2e

# Run specific test file
npm run test:e2e auth.spec.ts

# Run tests with specific tag
npm run test:e2e -- --grep "@smoke"

# Run tests in headed mode
npm run test:e2e -- --headed

# Run tests in specific browser
npm run test:e2e -- --project=firefox

# Run tests with trace
npm run test:e2e -- --trace on

# Update screenshots
npm run test:e2e -- --update-snapshots
```

## Troubleshooting

### Common Issues

1. **Flaky Tests**
   - Add explicit waits for elements
   - Check for race conditions
   - Ensure proper test isolation

2. **Timeout Errors**
   - Increase timeout for slow operations
   - Check network conditions
   - Verify mock API responses

3. **Element Not Found**
   - Verify data-testid attributes
   - Check element visibility
   - Ensure page is fully loaded

### Getting Help

- Check test logs in `test-results/`
- Review traces in `test-results/trace.zip`
- Use `--debug` flag for interactive debugging
- Consult Playwright documentation

## Contributing

When adding new tests:

1. Follow the established patterns
2. Add Page Objects for new features
3. Update test data generators
4. Document complex test scenarios
5. Ensure tests pass locally before pushing
6. Add appropriate test tags

## Resources

- [Playwright Documentation](https://playwright.dev)
- [Testing Best Practices](https://testingjavascript.com)
- [Page Object Model Pattern](https://martinfowler.com/bliki/PageObject.html)
- Internal Slack: #testing-help