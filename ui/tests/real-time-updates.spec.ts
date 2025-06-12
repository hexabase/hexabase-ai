import { test, expect, Page } from '@playwright/test';

test.describe('Real-time Updates', () => {
  let page1: Page;
  let page2: Page;

  test.beforeEach(async ({ browser }) => {
    // Create two browser contexts for different users
    const context1 = await browser.newContext();
    const context2 = await browser.newContext();
    
    page1 = await context1.newPage();
    page2 = await context2.newPage();
    
    // Login both users
    for (const page of [page1, page2]) {
      await page.goto('/login');
      await page.fill('[data-testid="email-input"]', 'user@example.com');
      await page.fill('[data-testid="password-input"]', 'password123');
      await page.click('[data-testid="login-button"]');
      await page.waitForURL('/dashboard');
    }
  });

  test.afterEach(async () => {
    await page1.close();
    await page2.close();
  });

  test('TC-RT-001: Multi-User Collaboration', async () => {
    // Both users navigate to same workspace
    const workspaceUrl = '/dashboard/organizations/org-1/workspaces/ws-1';
    await page1.goto(workspaceUrl);
    await page2.goto(workspaceUrl);
    
    // User A creates new project
    await page1.click('button:has-text("Create Project")');
    await page1.fill('[data-testid="project-name-input"]', 'Realtime Test Project');
    await page1.fill('[data-testid="project-description"]', 'Testing real-time sync');
    await page1.click('button:has-text("Create")');
    
    // User B should see the project immediately (within 2 seconds)
    await expect(page2.locator('[data-testid="project-card-Realtime Test Project"]')).toBeVisible({ timeout: 2000 });
    
    // User B deploys an application
    await page2.click('[data-testid="project-card-Realtime Test Project"]');
    await page2.click('button:has-text("Deploy Application")');
    await page2.click('[data-testid="app-type-stateless"]');
    await page2.fill('[data-testid="app-name-input"]', 'realtime-app');
    await page2.fill('[data-testid="container-image"]', 'nginx:latest');
    await page2.click('button:has-text("Deploy")');
    
    // User A should see deployment status immediately
    await page1.click('[data-testid="project-card-Realtime Test Project"]');
    await expect(page1.locator('[data-testid="app-card-realtime-app"]')).toBeVisible({ timeout: 2000 });
    await expect(page1.locator('[data-testid="app-status-realtime-app"]')).toContainText('Creating');
    
    // Both users modify different resources simultaneously
    // User A creates another project
    const createProject2 = page1.click('button:has-text("Create Project")').then(async () => {
      await page1.fill('[data-testid="project-name-input"]', 'Project A');
      await page1.click('button:has-text("Create")');
    });
    
    // User B creates another application
    const createApp2 = page2.click('button:has-text("Deploy Application")').then(async () => {
      await page2.click('[data-testid="app-type-stateless"]');
      await page2.fill('[data-testid="app-name-input"]', 'app-b');
      await page2.fill('[data-testid="container-image"]', 'redis:latest');
      await page2.click('button:has-text("Deploy")');
    });
    
    // Execute simultaneously
    await Promise.all([createProject2, createApp2]);
    
    // Verify no conflicts - both operations should succeed
    await expect(page1.locator('[data-testid="project-card-Project A"]')).toBeVisible({ timeout: 2000 });
    await expect(page2.locator('[data-testid="app-card-app-b"]')).toBeVisible({ timeout: 2000 });
    
    // Cross-verify: User A sees User B's app
    await expect(page1.locator('[data-testid="app-card-app-b"]')).toBeVisible({ timeout: 2000 });
    
    // Cross-verify: User B sees User A's project
    await page2.goto(workspaceUrl); // Navigate back to workspace
    await expect(page2.locator('[data-testid="project-card-Project A"]')).toBeVisible({ timeout: 2000 });
  });

  test('TC-RT-002: Deployment Status Updates', async () => {
    // Navigate to project with existing application
    const projectUrl = '/dashboard/organizations/org-1/workspaces/ws-1/projects/proj-1';
    await page1.goto(projectUrl);
    await page2.goto(projectUrl);
    
    // Start deployment in page1
    await page1.click('button:has-text("Deploy Application")');
    await page1.click('[data-testid="app-type-stateless"]');
    await page1.fill('[data-testid="app-name-input"]', 'status-test-app');
    await page1.fill('[data-testid="container-image"]', 'nginx:alpine');
    await page1.fill('[data-testid="replicas-input"]', '3');
    await page1.click('button:has-text("Deploy")');
    
    // Both pages should see initial status
    for (const page of [page1, page2]) {
      await expect(page.locator('[data-testid="app-card-status-test-app"]')).toBeVisible({ timeout: 2000 });
      await expect(page.locator('[data-testid="app-status-status-test-app"]')).toContainText('Creating');
    }
    
    // Watch status progression in both pages
    // Creating → Pending → Running
    for (const page of [page1, page2]) {
      await expect(page.locator('[data-testid="app-status-status-test-app"]')).toContainText('Pending', { timeout: 10000 });
      await expect(page.locator('[data-testid="pod-count-status-test-app"]')).toContainText('0/3');
    }
    
    // Pods starting
    for (const page of [page1, page2]) {
      await expect(page.locator('[data-testid="pod-count-status-test-app"]')).toContainText('1/3', { timeout: 15000 });
      await expect(page.locator('[data-testid="pod-count-status-test-app"]')).toContainText('2/3', { timeout: 15000 });
      await expect(page.locator('[data-testid="pod-count-status-test-app"]')).toContainText('3/3', { timeout: 15000 });
      await expect(page.locator('[data-testid="app-status-status-test-app"]')).toContainText('Running');
    }
    
    // Simulate pod failure in page1
    await page1.click('[data-testid="app-card-status-test-app"]');
    await page1.click('[data-testid="simulate-pod-failure"]');
    
    // Both pages should see failure immediately
    for (const page of [page1, page2]) {
      await expect(page.locator('[data-testid="pod-count-status-test-app"]')).toContainText('2/3', { timeout: 2000 });
      await expect(page.locator('[data-testid="pod-alert"]')).toContainText('Pod failure detected');
    }
    
    // Watch recovery in both pages
    for (const page of [page1, page2]) {
      await expect(page.locator('[data-testid="pod-status-recovery"]')).toContainText('Recovering...', { timeout: 5000 });
      await expect(page.locator('[data-testid="pod-count-status-test-app"]')).toContainText('3/3', { timeout: 30000 });
      await expect(page.locator('[data-testid="app-status-status-test-app"]')).toContainText('Running');
    }
  });

  test('WebSocket reconnection handling', async () => {
    // Navigate to workspace
    await page1.goto('/dashboard/organizations/org-1/workspaces/ws-1');
    
    // Verify WebSocket connected
    await expect(page1.locator('[data-testid="ws-status"]')).toContainText('Connected');
    
    // Simulate network interruption
    await page1.context().setOffline(true);
    
    // Should show disconnected status
    await expect(page1.locator('[data-testid="ws-status"]')).toContainText('Disconnected', { timeout: 5000 });
    
    // Restore connection
    await page1.context().setOffline(false);
    
    // Should automatically reconnect
    await expect(page1.locator('[data-testid="ws-status"]')).toContainText('Reconnecting...', { timeout: 5000 });
    await expect(page1.locator('[data-testid="ws-status"]')).toContainText('Connected', { timeout: 10000 });
    
    // Create project to verify connection works
    await page1.click('button:has-text("Create Project")');
    await page1.fill('[data-testid="project-name-input"]', 'After Reconnect');
    await page1.click('button:has-text("Create")');
    
    // Should work normally
    await expect(page1.locator('[data-testid="project-card-After Reconnect"]')).toBeVisible();
  });

  test('Concurrent resource updates', async () => {
    // Both users navigate to same application
    const appUrl = '/dashboard/organizations/org-1/workspaces/ws-1/projects/proj-1/applications/app-1';
    await page1.goto(appUrl);
    await page2.goto(appUrl);
    
    // Both users try to scale the application simultaneously
    const scale1 = page1.click('[data-testid="scale-app"]').then(async () => {
      await page1.fill('[data-testid="replicas-input"]', '5');
      await page1.click('button:has-text("Scale")');
    });
    
    const scale2 = page2.click('[data-testid="scale-app"]').then(async () => {
      await page2.fill('[data-testid="replicas-input"]', '10');
      await page2.click('button:has-text("Scale")');
    });
    
    // Execute simultaneously
    await Promise.all([scale1, scale2]);
    
    // One should succeed, one should get conflict error
    const page1Success = await page1.locator('[data-testid="scale-success"]').isVisible({ timeout: 2000 }).catch(() => false);
    const page2Success = await page2.locator('[data-testid="scale-success"]').isVisible({ timeout: 2000 }).catch(() => false);
    
    // Exactly one should succeed
    expect(page1Success !== page2Success).toBeTruthy();
    
    // The failed one should show conflict error
    const failedPage = page1Success ? page2 : page1;
    await expect(failedPage.locator('[data-testid="conflict-error"]')).toContainText('Resource was modified');
    
    // Both should eventually show the same replica count
    await expect(page1.locator('[data-testid="replica-count"]')).toHaveText(/^(5|10)$/);
    await expect(page2.locator('[data-testid="replica-count"]')).toHaveText(/^(5|10)$/);
    
    // Counts should match
    const count1 = await page1.locator('[data-testid="replica-count"]').textContent();
    const count2 = await page2.locator('[data-testid="replica-count"]').textContent();
    expect(count1).toBe(count2);
  });
});