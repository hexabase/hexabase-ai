import { test, expect } from '@playwright/test';
import { DebugHelper } from '../utils/debug-helpers';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { testUsers } from '../fixtures/mock-data';

/**
 * Debug Basic Functions Test Suite
 * 
 * This test suite is specifically designed for debugging and developer walkthrough.
 * It tests core functionality with enhanced logging, console error detection,
 * and step-by-step execution capabilities.
 * 
 * Run with: ./scripts/e2e-debug-enhanced.sh --developer --test debug-basic-functions.spec.ts
 */

test.describe('Debug: Basic Functions', () => {
  let debugHelper: DebugHelper;

  test.beforeEach(async ({ page }, testInfo) => {
    // Initialize debug helper
    debugHelper = new DebugHelper(page, testInfo);
    
    console.log('\n🧪 Starting basic functions debug test');
    console.log(`📁 Test: ${testInfo.title}`);
    console.log(`🔧 Debug mode: ${process.env.DEVELOPER_MODE === 'true' ? 'Developer' : 'Standard'}`);
    console.log(`⏱️ Timeout: ${testInfo.timeout}ms`);
  });

  test.afterEach(async ({ page }, testInfo) => {
    // Generate debug report
    const report = debugHelper.generateDebugReport();
    console.log('\n📊 Debug Report Summary:');
    console.log(`  • Test duration: ${report.test.duration}ms`);
    console.log(`  • Steps executed: ${report.test.steps}`);
    console.log(`  • Console errors: ${report.console.errors}`);
    console.log(`  • Console warnings: ${report.console.warnings}`);

    // Check for console errors
    if (debugHelper.hasConsoleErrors()) {
      const errorSummary = debugHelper.getConsoleErrorSummary();
      console.error(`\n❌ Found ${errorSummary.count} console errors:`);
      errorSummary.errors.forEach((error, index) => {
        console.error(`  ${index + 1}. [${error.type}] ${error.text}`);
      });
      
      // In developer mode, don't fail the test on console errors for investigation
      if (process.env.DEVELOPER_MODE !== 'true') {
        throw new Error(`Test completed with ${errorSummary.count} console errors`);
      }
    } else {
      console.log('✅ No console errors detected');
    }
  });

  test('Login Flow - Developer Walkthrough', async ({ page }) => {
    const loginPage = new LoginPage(page);
    const dashboardPage = new DashboardPage(page);

    await debugHelper.step('Navigate to application home page', async () => {
      await debugHelper.goto('http://localhost:3000');
    });

    await debugHelper.step('Verify login page elements are visible', async () => {
      await debugHelper.waitForSelector('[data-testid="email-input"]');
      await debugHelper.waitForSelector('[data-testid="password-input"]');
      await debugHelper.waitForSelector('[data-testid="login-button"]');
      
      // Check for any missing elements
      const missingElements = [];
      try {
        await expect(page.locator('[data-testid="email-input"]')).toBeVisible();
      } catch {
        missingElements.push('email-input');
      }
      
      try {
        await expect(page.locator('[data-testid="password-input"]')).toBeVisible();
      } catch {
        missingElements.push('password-input');
      }
      
      try {
        await expect(page.locator('[data-testid="login-button"]')).toBeVisible();
      } catch {
        missingElements.push('login-button');
      }

      if (missingElements.length > 0) {
        console.warn(`⚠️ Missing elements: ${missingElements.join(', ')}`);
      }
    });

    await debugHelper.step('Enter developer account credentials', async () => {
      await debugHelper.fill('[data-testid="email-input"]', testUsers.developer?.email || 'dev@hexabase.com');
      await debugHelper.fill('[data-testid="password-input"]', testUsers.developer?.password || 'dev123456');
    });

    await debugHelper.step('Submit login form', async () => {
      await debugHelper.click('[data-testid="login-button"]');
    });

    await debugHelper.step('Wait for authentication to complete', async () => {
      // Wait for either dashboard or error message
      try {
        await page.waitForURL('**/dashboard**', { timeout: 10000 });
        console.log('✅ Successfully redirected to dashboard');
      } catch (error) {
        console.log('⚠️ Not redirected to dashboard, checking for error messages...');
        
        // Check for error messages
        const errorMessage = page.locator('[data-testid="error-message"]');
        if (await errorMessage.isVisible()) {
          const errorText = await errorMessage.textContent();
          console.error(`❌ Login error: ${errorText}`);
        }
        
        // Check for validation errors
        const validationErrors = page.locator('.error, .alert-error, [role="alert"]');
        const errorCount = await validationErrors.count();
        if (errorCount > 0) {
          console.log(`Found ${errorCount} validation errors:`);
          for (let i = 0; i < errorCount; i++) {
            const error = validationErrors.nth(i);
            const text = await error.textContent();
            console.log(`  • ${text}`);
          }
        }
        
        throw error;
      }
    });

    await debugHelper.step('Verify dashboard elements are loaded', async () => {
      // Check for main dashboard components
      const dashboardElements = [
        '[data-testid="organization-selector"]',
        '[data-testid="workspace-grid"]',
        '[data-testid="user-menu"]',
        '[data-testid="main-navigation"]'
      ];

      const loadedElements = [];
      const missingElements = [];

      for (const selector of dashboardElements) {
        try {
          await expect(page.locator(selector)).toBeVisible({ timeout: 5000 });
          loadedElements.push(selector);
        } catch {
          missingElements.push(selector);
        }
      }

      console.log(`✅ Loaded elements: ${loadedElements.length}/${dashboardElements.length}`);
      if (missingElements.length > 0) {
        console.warn(`⚠️ Missing elements: ${missingElements.join(', ')}`);
      }
    });
  });

  test('Navigation and Menu Testing', async ({ page }) => {
    // Login first
    await debugHelper.step('Perform login for navigation test', async () => {
      await debugHelper.goto('http://localhost:3000');
      await debugHelper.fill('[data-testid="email-input"]', testUsers.developer?.email || 'dev@hexabase.com');
      await debugHelper.fill('[data-testid="password-input"]', testUsers.developer?.password || 'dev123456');
      await debugHelper.click('[data-testid="login-button"]');
      await page.waitForURL('**/dashboard**', { timeout: 10000 });
    });

    await debugHelper.step('Test main navigation menu', async () => {
      const navigationItems = [
        '[data-testid="nav-dashboard"]',
        '[data-testid="nav-workspaces"]',
        '[data-testid="nav-projects"]',
        '[data-testid="nav-applications"]',
        '[data-testid="nav-settings"]'
      ];

      for (const navItem of navigationItems) {
        try {
          await debugHelper.waitForSelector(navItem, { timeout: 3000 });
          console.log(`✅ Navigation item found: ${navItem}`);
        } catch {
          console.warn(`⚠️ Navigation item missing: ${navItem}`);
        }
      }
    });

    await debugHelper.step('Test user menu functionality', async () => {
      await debugHelper.click('[data-testid="user-menu"]');
      
      // Check for user menu items
      const userMenuItems = [
        '[data-testid="user-profile"]',
        '[data-testid="user-settings"]',
        '[data-testid="logout-button"]'
      ];

      for (const menuItem of userMenuItems) {
        try {
          await debugHelper.waitForSelector(menuItem, { timeout: 3000 });
          console.log(`✅ User menu item found: ${menuItem}`);
        } catch {
          console.warn(`⚠️ User menu item missing: ${menuItem}`);
        }
      }
    });
  });

  test('Form Validation and Error Handling', async ({ page }) => {
    await debugHelper.step('Navigate to login page for validation testing', async () => {
      await debugHelper.goto('http://localhost:3000');
    });

    await debugHelper.step('Test empty form submission', async () => {
      await debugHelper.click('[data-testid="login-button"]');
      
      // Check for validation messages
      const validationMessages = [
        '[data-testid="email-error"]',
        '[data-testid="password-error"]',
        '.field-error',
        '.validation-error'
      ];

      let foundValidations = 0;
      for (const selector of validationMessages) {
        try {
          await debugHelper.waitForSelector(selector, { timeout: 2000 });
          const message = await page.locator(selector).textContent();
          console.log(`✅ Validation message: ${message}`);
          foundValidations++;
        } catch {
          // Validation message not found with this selector
        }
      }

      if (foundValidations === 0) {
        console.warn('⚠️ No validation messages found for empty form submission');
      }
    });

    await debugHelper.step('Test invalid email format', async () => {
      await debugHelper.fill('[data-testid="email-input"]', 'invalid-email');
      await debugHelper.fill('[data-testid="password-input"]', 'somepassword');
      await debugHelper.click('[data-testid="login-button"]');
      
      // Look for email format validation
      const emailValidation = page.locator('[data-testid="email-error"], .email-error');
      try {
        await expect(emailValidation).toBeVisible({ timeout: 3000 });
        const message = await emailValidation.textContent();
        console.log(`✅ Email validation: ${message}`);
      } catch {
        console.warn('⚠️ No email format validation found');
      }
    });

    await debugHelper.step('Test invalid credentials', async () => {
      await debugHelper.fill('[data-testid="email-input"]', 'test@invalid.com');
      await debugHelper.fill('[data-testid="password-input"]', 'wrongpassword');
      await debugHelper.click('[data-testid="login-button"]');
      
      // Check for authentication error
      try {
        const errorMessage = page.locator('[data-testid="auth-error"], .auth-error, .login-error');
        await expect(errorMessage).toBeVisible({ timeout: 5000 });
        const message = await errorMessage.textContent();
        console.log(`✅ Auth error message: ${message}`);
      } catch {
        console.warn('⚠️ No authentication error message found');
      }
    });
  });

  test('API Integration Check', async ({ page }) => {
    let apiRequests: any[] = [];
    let apiErrors: any[] = [];

    // Monitor network requests
    page.on('request', request => {
      if (request.url().includes('/api/') || request.url().includes('localhost:8080')) {
        apiRequests.push({
          method: request.method(),
          url: request.url(),
          timestamp: new Date().toISOString()
        });
      }
    });

    page.on('requestfailed', request => {
      if (request.url().includes('/api/') || request.url().includes('localhost:8080')) {
        apiErrors.push({
          method: request.method(),
          url: request.url(),
          error: request.failure()?.errorText,
          timestamp: new Date().toISOString()
        });
      }
    });

    await debugHelper.step('Navigate and trigger API calls', async () => {
      await debugHelper.goto('http://localhost:3000');
    });

    await debugHelper.step('Perform login to test API integration', async () => {
      await debugHelper.fill('[data-testid="email-input"]', testUsers.developer?.email || 'dev@hexabase.com');
      await debugHelper.fill('[data-testid="password-input"]', testUsers.developer?.password || 'dev123456');
      await debugHelper.click('[data-testid="login-button"]');
      
      // Wait a bit for API calls to complete
      await page.waitForTimeout(2000);
    });

    await debugHelper.step('Analyze API requests and responses', async () => {
      console.log(`📡 Total API requests: ${apiRequests.length}`);
      apiRequests.forEach((req, index) => {
        console.log(`  ${index + 1}. ${req.method} ${req.url}`);
      });

      if (apiErrors.length > 0) {
        console.error(`❌ API errors found: ${apiErrors.length}`);
        apiErrors.forEach((error, index) => {
          console.error(`  ${index + 1}. ${error.method} ${error.url} - ${error.error}`);
        });
      } else {
        console.log('✅ No API errors detected');
      }

      // Check for expected API endpoints
      const expectedEndpoints = ['/api/auth/login', '/api/user/profile', '/api/health'];
      const foundEndpoints = apiRequests.filter(req => 
        expectedEndpoints.some(endpoint => req.url.includes(endpoint))
      );

      console.log(`🎯 Expected endpoints found: ${foundEndpoints.length}/${expectedEndpoints.length}`);
    });
  });
});