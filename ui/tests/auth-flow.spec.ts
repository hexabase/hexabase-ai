import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test('TC-AUTH-001: OAuth Login Flow', async ({ page }) => {
    // Navigate to login page
    await page.goto('/');
    
    // Should redirect to login page
    await expect(page).toHaveURL('/login');
    
    // Click OAuth login button
    await page.click('button:has-text("Login with GitHub")');
    
    // Mock OAuth provider response
    // In real test, would handle OAuth provider login
    await page.route('**/api/auth/callback', async route => {
      await route.fulfill({
        status: 302,
        headers: {
          'Location': '/dashboard'
        }
      });
    });
    
    // Verify redirect to dashboard
    await expect(page).toHaveURL('/dashboard');
    
    // Verify user profile loaded
    await expect(page.locator('[data-testid="user-menu"]')).toContainText('test@example.com');
    
    // Verify organizations loaded
    await expect(page.locator('[data-testid="org-switcher"]')).toBeVisible();
  });

  test('TC-AUTH-002: Session Persistence', async ({ page, context }) => {
    // Login first
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    
    // Wait for dashboard
    await expect(page).toHaveURL('/dashboard');
    
    // Get cookies
    const cookies = await context.cookies();
    
    // Close and reopen page
    await page.close();
    const newPage = await context.newPage();
    
    // Navigate directly to dashboard
    await newPage.goto('/dashboard');
    
    // Should not redirect to login
    await expect(newPage).toHaveURL('/dashboard');
    
    // User should still be logged in
    await expect(newPage.locator('[data-testid="user-menu"]')).toContainText('test@example.com');
  });

  test('TC-AUTH-003: Multi-Organization Switching', async ({ page }) => {
    // Login as user with multiple organizations
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'multi-org@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    
    // Wait for dashboard
    await expect(page).toHaveURL('/dashboard');
    
    // Open organization switcher
    await page.click('[data-testid="org-switcher"]');
    
    // Verify multiple organizations
    await expect(page.locator('[data-testid="org-option"]')).toHaveCount(3);
    
    // Current org should be highlighted
    await expect(page.locator('[data-testid="org-option-active"]')).toContainText('Acme Corp');
    
    // Switch to different organization
    await page.click('[data-testid="org-option"]:has-text("Tech Startup")');
    
    // Verify URL updated
    await expect(page).toHaveURL(/.*\/organizations\/org-2/);
    
    // Verify context switched
    await expect(page.locator('[data-testid="org-name"]')).toContainText('Tech Startup');
    
    // Switch rapidly between organizations
    await page.click('[data-testid="org-switcher"]');
    await page.click('[data-testid="org-option"]:has-text("Enterprise Inc")');
    await expect(page).toHaveURL(/.*\/organizations\/org-3/);
    
    await page.click('[data-testid="org-switcher"]');
    await page.click('[data-testid="org-option"]:has-text("Acme Corp")');
    await expect(page).toHaveURL(/.*\/organizations\/org-1/);
    
    // Verify no data leakage (workspaces should be different)
    const workspaces = await page.locator('[data-testid="workspace-card"]').count();
    expect(workspaces).toBeGreaterThan(0);
  });

  test('Logout and session cleanup', async ({ page, context }) => {
    // Login first
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    
    // Wait for dashboard
    await expect(page).toHaveURL('/dashboard');
    
    // Open user menu and logout
    await page.click('[data-testid="user-menu"]');
    await page.click('[data-testid="logout-button"]');
    
    // Should redirect to login
    await expect(page).toHaveURL('/login');
    
    // Try to access protected route
    await page.goto('/dashboard');
    
    // Should redirect back to login
    await expect(page).toHaveURL('/login');
    
    // Verify session cookies cleared
    const cookies = await context.cookies();
    const sessionCookie = cookies.find(c => c.name === 'session');
    expect(sessionCookie).toBeUndefined();
  });
});