import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers } from '../fixtures/mock-data';

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Setup mock API responses
    await setupMockAPI(page);
  });

  test('successful login with valid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page);
    const dashboardPage = new DashboardPage(page);
    
    // Navigate to login page
    await loginPage.goto();
    
    // Verify login page is displayed
    await expect(loginPage.emailInput).toBeVisible();
    await expect(loginPage.passwordInput).toBeVisible();
    await expect(loginPage.loginButton).toBeVisible();
    
    // Perform login
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    
    // Verify successful login
    const isLoggedIn = await loginPage.isLoggedIn();
    expect(isLoggedIn).toBe(true);
    
    // Verify dashboard is displayed
    await expect(dashboardPage.organizationSelector).toBeVisible();
    await expect(dashboardPage.workspaceGrid).toBeVisible();
  });

  test('displays error with invalid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page);
    
    await loginPage.goto();
    
    // Try to login with invalid credentials
    await loginPage.login('invalid@email.com', 'wrongpassword');
    
    // Verify error message is displayed
    await loginPage.expectError('Invalid credentials');
    
    // Verify still on login page
    await expect(loginPage.emailInput).toBeVisible();
    expect(page.url()).not.toContain('/dashboard');
  });

  test('handles empty form submission', async ({ page }) => {
    const loginPage = new LoginPage(page);
    
    await loginPage.goto();
    
    // Click login without entering credentials
    await loginPage.loginButton.click();
    
    // Verify validation messages
    await expect(page.getByText('Email is required')).toBeVisible();
    await expect(page.getByText('Password is required')).toBeVisible();
  });

  test('logout flow', async ({ page }) => {
    const loginPage = new LoginPage(page);
    const dashboardPage = new DashboardPage(page);
    
    // Login first
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    // Perform logout
    await dashboardPage.logout();
    
    // Verify redirected to login page
    await expect(loginPage.emailInput).toBeVisible();
    expect(page.url()).toContain('/login');
  });

  test('persists session on page refresh', async ({ page }) => {
    const loginPage = new LoginPage(page);
    const dashboardPage = new DashboardPage(page);
    
    // Login
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    // Refresh page
    await page.reload();
    
    // Verify still logged in
    await expect(dashboardPage.organizationSelector).toBeVisible();
    expect(page.url()).toContain('/dashboard');
  });

  test('redirects to requested page after login', async ({ page }) => {
    const loginPage = new LoginPage(page);
    
    // Try to access protected page
    await page.goto('/dashboard/organizations/org-1/workspaces');
    
    // Should redirect to login
    await expect(loginPage.emailInput).toBeVisible();
    
    // Login
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    
    // Should redirect to originally requested page
    await page.waitForURL('**/dashboard/organizations/org-1/workspaces');
  });

  test('handles network errors gracefully', async ({ page }) => {
    const loginPage = new LoginPage(page);
    
    // Override API to simulate network error
    await page.route('**/api/auth/login', async (route) => {
      await route.abort('failed');
    });
    
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    
    // Should show network error message
    await expect(page.getByText(/network error|connection failed/i)).toBeVisible();
  });

  test('shows loading state during authentication', async ({ page }) => {
    const loginPage = new LoginPage(page);
    
    // Add delay to login response
    await page.route('**/api/auth/login', async (route) => {
      await page.waitForTimeout(1000); // Simulate slow network
      await route.continue();
    });
    
    await loginPage.goto();
    
    // Start login
    const loginPromise = loginPage.login(testUsers.admin.email, testUsers.admin.password);
    
    // Check for loading state
    await expect(loginPage.loginButton).toBeDisabled();
    await expect(page.getByTestId('loading-spinner')).toBeVisible();
    
    // Wait for login to complete
    await loginPromise;
  });
});