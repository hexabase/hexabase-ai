# E2E Tests for Hexabase AI Platform

This directory contains end-to-end tests for the Hexabase AI platform using Playwright.

## 📁 Directory Structure

```
e2e/
├── tests/                    # Test specifications
│   ├── auth.spec.ts         # Authentication flows
│   ├── organization-workspace.spec.ts  # Multi-tenancy tests
│   ├── projects.spec.ts     # Project management
│   ├── applications.spec.ts # Application deployment
│   ├── deployments.spec.ts  # Deployment strategies
│   ├── cicd-pipeline.spec.ts # CI/CD integration
│   ├── backup-restore.spec.ts # Backup operations
│   ├── serverless-functions.spec.ts # Knative functions
│   └── smoke.spec.ts        # Critical path smoke tests
├── pages/                   # Page Object Models
│   ├── LoginPage.ts
│   ├── DashboardPage.ts
│   ├── WorkspacePage.ts
│   ├── ProjectPage.ts
│   ├── ApplicationPage.ts
│   └── MonitoringPage.ts
├── fixtures/                # Test data
│   └── mock-data.ts
├── utils/                   # Helper utilities
│   ├── mock-api.ts         # API mocking
│   ├── test-helpers.ts     # Common test utilities
│   ├── test-tags.ts        # Test categorization
│   └── screenshot-helper.ts # Screenshot utilities
└── README.md               # This file
```

## 🚀 Running Tests

### Local Development

```bash
# Install dependencies
npm install

# Install Playwright browsers
npx playwright install

# Run all tests
npm run test:e2e

# Run tests in UI mode (recommended for development)
npx playwright test --ui

# Run specific test file
npx playwright test e2e/tests/auth.spec.ts

# Run tests with specific tag
npx playwright test --grep "@smoke"

# Run tests in headed mode (see browser)
npx playwright test --headed
```

### Test Categories

Tests are tagged for different purposes:

- `@smoke` - Critical path tests that must always pass
- `@visual` - Visual regression tests
- `@critical` - Must pass before deployment
- `@flaky` - Known unstable tests
- `@slow` - Tests that take >30 seconds

### Running by Category

```bash
# Run only smoke tests
npx playwright test --grep "@smoke"

# Run critical tests
npx playwright test --grep "@critical"

# Exclude flaky tests
npx playwright test --grep-invert "@flaky"
```

## 🔧 Configuration

### Environment Variables

```bash
# Base URL for tests (defaults to http://localhost:3000)
BASE_URL=https://staging.hexabase.ai

# Test environment
TEST_ENV=staging

# API credentials (if needed)
TEST_USER_EMAIL=test@example.com
TEST_USER_PASSWORD=secure-password
```

### Playwright Configuration

See `playwright.config.ts` for full configuration. Key settings:

- **Browsers**: Chromium, Firefox, WebKit, Mobile Chrome/Safari
- **Parallelization**: Tests run in parallel by default
- **Retries**: 2 retries on CI, 0 locally
- **Timeouts**: 30s test timeout, 5s assertion timeout
- **Screenshots**: On failure
- **Videos**: On failure

## 📊 Test Reports

### HTML Report

After running tests, view the HTML report:

```bash
npx playwright show-report
```

### CI Reports

GitHub Actions automatically:
- Generates HTML reports
- Captures screenshots on failure
- Comments on PRs with results
- Sends Slack notifications for failures

## 🎯 Writing Tests

### Basic Test Structure

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test.describe('Feature Name', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
  });

  test('should do something', async ({ page }) => {
    // Arrange
    await loginPage.login('user@example.com', 'password');
    
    // Act
    await page.click('[data-testid="some-button"]');
    
    // Assert
    await expect(page.getByTestId('result')).toBeVisible();
  });
});
```

### Page Object Pattern

Always use Page Objects for better maintainability:

```typescript
// pages/ExamplePage.ts
export class ExamplePage {
  constructor(private page: Page) {}
  
  // Locators
  get submitButton() {
    return this.page.getByTestId('submit-button');
  }
  
  // Actions
  async submitForm(data: FormData) {
    await this.fillForm(data);
    await this.submitButton.click();
  }
}
```

### Best Practices

1. **Use data-testid attributes** for reliable element selection
2. **Keep tests independent** - each test should run in isolation
3. **Use meaningful test names** that describe what is being tested
4. **Follow AAA pattern** - Arrange, Act, Assert
5. **Mock external dependencies** to ensure consistent results
6. **Add appropriate tags** for test categorization
7. **Clean up after tests** - delete created resources

## 🐛 Debugging

### Debug Mode

```bash
# Run with debug logs
DEBUG=pw:api npx playwright test

# Pause at specific point
await page.pause();

# Use Inspector
npx playwright test --debug
```

### VS Code Integration

1. Install "Playwright Test for VS Code" extension
2. Run/debug tests directly from the editor
3. Set breakpoints in test code

## 🔄 CI/CD Integration

### GitHub Actions Workflows

1. **e2e-tests.yml** - Full test suite on PR/push
2. **e2e-smoke-tests.yml** - Quick smoke tests on deployment
3. **visual-regression.yml** - Visual comparison on UI changes

### Running in CI

Tests automatically run on:
- Pull requests to main/develop
- Pushes to main/develop
- Nightly schedule (2 AM UTC)
- Manual workflow dispatch

## 📸 Screenshots

### Capturing Screenshots

```typescript
// Manual screenshot
await page.screenshot({ 
  path: 'screenshot.png',
  fullPage: true 
});

// Use helper function
await captureSuccessScreenshot(page, 'test-name', 'step-name');
```

### Viewing Test Screenshots

Failed test screenshots are automatically:
1. Saved to `test-results/` directory
2. Uploaded as GitHub Actions artifacts
3. Available in HTML report

## 🤝 Contributing

1. Write tests for new features
2. Update Page Objects for UI changes
3. Add appropriate test tags
4. Ensure tests pass locally before PR
5. Update this README for significant changes

## 📚 Resources

- [Playwright Documentation](https://playwright.dev)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [API Reference](https://playwright.dev/docs/api/class-test)
- [Debugging Guide](https://playwright.dev/docs/debug)