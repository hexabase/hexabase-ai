import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { ApplicationPage } from '../pages/ApplicationPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers, testWorkspaces } from '../fixtures/mock-data';
import { generateAppName, generateProjectName } from '../utils/test-helpers';
import { CRITICAL_TAG, FLAKY_TAG } from '../utils/test-tags';

test.describe('Error Handling and Edge Cases', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let workspacePage: WorkspacePage;
  let projectPage: ProjectPage;
  let applicationPage: ApplicationPage;

  test.beforeEach(async ({ page }) => {
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    workspacePage = new WorkspacePage(page);
    projectPage = new ProjectPage(page);
    applicationPage = new ApplicationPage(page);
  });

  test.describe('Network Failures', () => {
    test(`handle network timeout gracefully ${CRITICAL_TAG}`, async ({ page }) => {
      await loginPage.goto();
      
      // Simulate network timeout
      await page.route('**/api/auth/login', async (route) => {
        // Don't respond, causing a timeout
        await page.waitForTimeout(35000);
      });
      
      // Attempt login
      await loginPage.emailInput.fill(testUsers.admin.email);
      await loginPage.passwordInput.fill(testUsers.admin.password);
      
      // Set shorter timeout for test
      await page.setDefaultTimeout(5000);
      
      const loginPromise = loginPage.loginButton.click();
      
      // Should show timeout error
      await expect(loginPromise).rejects.toThrow();
      
      // Verify error message
      await expect(page.getByText(/request timed out|network error/i)).toBeVisible();
      
      // Verify retry option
      const retryButton = page.getByTestId('retry-button');
      await expect(retryButton).toBeVisible();
    });

    test('handle intermittent network failures', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      
      let requestCount = 0;
      
      // Simulate intermittent failures (every other request fails)
      await page.route('**/api/**', async (route) => {
        requestCount++;
        if (requestCount % 2 === 0) {
          await route.abort('failed');
        } else {
          await route.continue();
        }
      });
      
      // Try to create project
      const projectName = generateProjectName();
      
      // First attempt should fail
      await workspacePage.createProjectButton.click();
      const dialog = page.getByRole('dialog');
      await dialog.getByTestId('project-name-input').fill(projectName);
      await dialog.getByTestId('create-button').click();
      
      // Should show error
      await expect(page.getByText(/network error|failed to create/i)).toBeVisible();
      
      // Retry should succeed
      await page.getByTestId('retry-button').click();
      await expect(page.getByText(/project created|success/i)).toBeVisible();
    });

    test('handle complete network disconnection', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      
      // Simulate offline
      await page.context().setOffline(true);
      
      // Try to navigate
      await expect(dashboardPage.openWorkspace(testWorkspaces[0].name)).rejects.toThrow();
      
      // Verify offline indicator
      await expect(page.getByTestId('offline-indicator')).toBeVisible();
      await expect(page.getByText(/offline|no internet/i)).toBeVisible();
      
      // Restore connection
      await page.context().setOffline(false);
      
      // Verify reconnection
      await page.waitForTimeout(1000);
      await expect(page.getByTestId('offline-indicator')).not.toBeVisible();
    });
  });

  test.describe('Permission Errors', () => {
    test('handle unauthorized access to resources', async ({ page }) => {
      // Login as regular user
      await loginPage.goto();
      await loginPage.login(testUsers.developer.email, testUsers.developer.password);
      
      // Mock unauthorized response
      await page.route('**/api/organizations/*/workspaces/*/settings', async (route) => {
        await route.fulfill({
          status: 403,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'Forbidden',
            message: 'You do not have permission to access workspace settings'
          })
        });
      });
      
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      
      // Try to access settings
      const settingsButton = page.getByTestId('workspace-settings-button');
      await settingsButton.click();
      
      // Verify permission error
      await expect(page.getByText(/permission|forbidden|unauthorized/i)).toBeVisible();
      await expect(page.getByText('workspace settings')).toBeVisible();
    });

    test('handle role-based feature restrictions', async ({ page }) => {
      // Login as viewer
      await loginPage.goto();
      await loginPage.login(testUsers.viewer.email, testUsers.viewer.password);
      
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      await workspacePage.openProject('Test Project');
      
      // Verify read-only access
      const deployButton = page.getByTestId('deploy-application-button');
      await expect(deployButton).toBeDisabled();
      
      // Hover should show tooltip
      await deployButton.hover();
      await expect(page.getByText(/view.*only|no.*permission/i)).toBeVisible();
      
      // Verify edit actions are disabled
      const deleteButtons = page.locator('[data-testid*="delete"]');
      for (const button of await deleteButtons.all()) {
        await expect(button).toBeDisabled();
      }
    });

    test('handle expired session gracefully', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      
      // Mock session expiry
      await page.route('**/api/auth/session', async (route) => {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'Session expired',
            code: 'SESSION_EXPIRED'
          })
        });
      });
      
      // Trigger API call
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      
      // Should redirect to login
      await page.waitForURL('**/login');
      await expect(page.getByText(/session expired|please login/i)).toBeVisible();
      
      // Verify return URL is preserved
      const url = new URL(page.url());
      expect(url.searchParams.get('returnTo')).toContain('workspace');
    });
  });

  test.describe('Resource Quota Exceeded', () => {
    test('handle CPU quota exceeded', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      await workspacePage.openProject('Test Project');
      
      // Mock quota exceeded
      await page.route('**/api/organizations/*/workspaces/*/applications', async (route) => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 403,
            contentType: 'application/json',
            body: JSON.stringify({
              error: 'QuotaExceeded',
              message: 'CPU quota exceeded. Current: 8/8 cores',
              details: {
                resource: 'cpu',
                current: '8',
                limit: '8',
                requested: '2'
              }
            })
          });
        } else {
          await route.continue();
        }
      });
      
      // Try to deploy application
      await projectPage.deployApplication({
        name: generateAppName(),
        type: 'stateless',
        image: 'nginx:latest',
        replicas: 3,
        port: 80,
      });
      
      // Verify quota error
      const errorDialog = page.getByRole('dialog');
      await expect(errorDialog).toContainText('CPU quota exceeded');
      await expect(errorDialog).toContainText('Current: 8/8 cores');
      
      // Verify suggestions
      await expect(errorDialog).toContainText(/scale down|reduce replicas|upgrade plan/i);
      
      // Verify upgrade link
      const upgradeLink = errorDialog.getByTestId('upgrade-plan-link');
      await expect(upgradeLink).toBeVisible();
    });

    test('handle storage quota exceeded', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      await workspacePage.openProject('Test Project');
      
      // Mock storage quota exceeded
      await page.route('**/api/organizations/*/workspaces/*/applications', async (route) => {
        if (route.request().method() === 'POST') {
          const body = await route.request().postDataJSON();
          if (body.type === 'stateful') {
            await route.fulfill({
              status: 403,
              contentType: 'application/json',
              body: JSON.stringify({
                error: 'QuotaExceeded',
                message: 'Storage quota exceeded. Current: 95GB/100GB',
                details: {
                  resource: 'storage',
                  current: '95GB',
                  limit: '100GB',
                  requested: '20GB'
                }
              })
            });
          } else {
            await route.continue();
          }
        } else {
          await route.continue();
        }
      });
      
      // Try to deploy stateful app
      await projectPage.deployApplication({
        name: generateAppName(),
        type: 'stateful',
        image: 'postgres:14',
        port: 5432,
        storage: '20Gi',
      });
      
      // Verify storage quota error
      await expect(page.getByText('Storage quota exceeded')).toBeVisible();
      await expect(page.getByText('95GB/100GB')).toBeVisible();
      
      // Verify cleanup suggestion
      await expect(page.getByText(/clean up|delete unused/i)).toBeVisible();
    });
  });

  test.describe('Invalid Input Handling', () => {
    test('handle invalid application configuration', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      await workspacePage.openProject('Test Project');
      
      // Open deploy dialog
      const deployButton = page.getByTestId('deploy-application-button');
      await deployButton.click();
      
      const dialog = page.getByRole('dialog');
      
      // Test invalid app name
      await dialog.getByTestId('app-name-input').fill('Invalid Name!@#');
      await dialog.getByTestId('app-image-input').fill('nginx:latest');
      await dialog.getByTestId('deploy-button').click();
      
      await expect(dialog.getByText(/invalid.*name|alphanumeric/i)).toBeVisible();
      
      // Test invalid image
      await dialog.getByTestId('app-name-input').fill('valid-name');
      await dialog.getByTestId('app-image-input').fill('not a valid image');
      await dialog.getByTestId('deploy-button').click();
      
      await expect(dialog.getByText(/invalid.*image|format/i)).toBeVisible();
      
      // Test invalid port
      await dialog.getByTestId('app-image-input').fill('nginx:latest');
      await dialog.getByTestId('app-port-input').fill('99999');
      await dialog.getByTestId('deploy-button').click();
      
      await expect(dialog.getByText(/invalid.*port|1-65535/i)).toBeVisible();
      
      // Test invalid replicas
      await dialog.getByTestId('app-port-input').fill('80');
      await dialog.getByTestId('app-replicas-input').fill('0');
      await dialog.getByTestId('deploy-button').click();
      
      await expect(dialog.getByText(/at least 1|minimum.*1/i)).toBeVisible();
    });

    test('handle XSS attempts in user input', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      
      // Try XSS in project name
      const xssPayload = '<script>alert("XSS")</script>';
      
      await workspacePage.createProjectButton.click();
      const dialog = page.getByRole('dialog');
      await dialog.getByTestId('project-name-input').fill(xssPayload);
      await dialog.getByTestId('create-button').click();
      
      // Should sanitize or reject
      await expect(page.getByText(/invalid|not allowed/i)).toBeVisible();
      
      // Verify script not executed
      await page.waitForTimeout(1000);
      const alerts = await page.evaluate(() => window.alert);
      expect(alerts).toBeUndefined();
    });

    test('handle SQL injection attempts', async ({ page }) => {
      await loginPage.goto();
      
      // Try SQL injection in login
      const sqlPayload = "admin' OR '1'='1";
      await loginPage.emailInput.fill(sqlPayload);
      await loginPage.passwordInput.fill('password');
      await loginPage.loginButton.click();
      
      // Should show invalid credentials, not SQL error
      await expect(loginPage.errorMessage).toBeVisible();
      await expect(loginPage.errorMessage).toContainText(/invalid.*credentials|email.*password/i);
      await expect(loginPage.errorMessage).not.toContainText(/sql|query|syntax/i);
    });
  });

  test.describe('Concurrent Operations', () => {
    test(`handle concurrent deployments ${FLAKY_TAG}`, async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      await workspacePage.openProject('Test Project');
      
      // Deploy multiple apps simultaneously
      const deployments = [];
      for (let i = 0; i < 3; i++) {
        deployments.push(
          projectPage.deployApplication({
            name: `concurrent-app-${i}`,
            type: 'stateless',
            image: 'nginx:latest',
            replicas: 1,
            port: 80 + i,
          })
        );
      }
      
      // Wait for all deployments
      const results = await Promise.allSettled(deployments);
      
      // At least one should succeed
      const succeeded = results.filter(r => r.status === 'fulfilled').length;
      expect(succeeded).toBeGreaterThanOrEqual(1);
      
      // Check for race condition errors
      const failed = results.filter(r => r.status === 'rejected');
      for (const failure of failed) {
        // Should show meaningful error, not system error
        expect(failure.reason).not.toContain('undefined');
        expect(failure.reason).not.toContain('null');
      }
    });

    test('handle concurrent resource updates', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      await workspacePage.openProject('Test Project');
      
      // Deploy an application
      const appName = generateAppName();
      await projectPage.deployApplication({
        name: appName,
        type: 'stateless',
        image: 'nginx:latest',
        replicas: 2,
        port: 80,
      });
      
      await projectPage.waitForApplicationStatus(appName, 'running');
      
      // Open app in multiple tabs
      const appCard = await projectPage.getApplicationCard(appName);
      await appCard.click();
      
      const newTab = await page.context().newPage();
      await newTab.goto(page.url());
      
      // Try to scale in both tabs
      const scale1 = applicationPage.scaleApplication(3);
      
      const newAppPage = new ApplicationPage(newTab);
      const scale2 = newAppPage.scaleApplication(5);
      
      // Wait for both operations
      const [result1, result2] = await Promise.allSettled([scale1, scale2]);
      
      // One should succeed, one should get conflict
      if (result1.status === 'fulfilled') {
        expect(result2.status).toBe('rejected');
        // Verify conflict error on second tab
        await expect(newTab.getByText(/conflict|already.*modified/i)).toBeVisible();
      } else {
        expect(result2.status).toBe('fulfilled');
        // Verify conflict error on first tab
        await expect(page.getByText(/conflict|already.*modified/i)).toBeVisible();
      }
      
      await newTab.close();
    });
  });

  test.describe('Large Data Handling', () => {
    test('handle large log files gracefully', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      await workspacePage.openProject('Test Project');
      
      // Deploy app
      const appName = generateAppName();
      await projectPage.deployApplication({
        name: appName,
        type: 'stateless',
        image: 'nginx:latest',
        replicas: 1,
        port: 80,
      });
      
      await projectPage.waitForApplicationStatus(appName, 'running');
      
      // Mock large log response
      await page.route('**/api/organizations/*/workspaces/*/applications/*/logs', async (route) => {
        const largeLogs = Array(10000).fill('2025-01-14 12:00:00 INFO Large log entry with lots of text...\n').join('');
        await route.fulfill({
          status: 200,
          contentType: 'text/plain',
          body: largeLogs
        });
      });
      
      // View logs
      const appCard = await projectPage.getApplicationCard(appName);
      await appCard.click();
      await applicationPage.logsTab.click();
      
      // Should handle large logs without freezing
      await expect(page.getByTestId('logs-content')).toBeVisible({ timeout: 10000 });
      
      // Should show truncation warning
      await expect(page.getByText(/truncated|showing.*lines/i)).toBeVisible();
      
      // Should have download option
      await expect(page.getByTestId('download-logs-button')).toBeVisible();
    });

    test('handle pagination for large lists', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      
      // Mock large organization list
      await page.route('**/api/organizations', async (route) => {
        const organizations = Array(100).fill(null).map((_, i) => ({
          id: `org-${i}`,
          name: `Organization ${i}`,
          plan: i % 3 === 0 ? 'enterprise' : 'professional',
          members: Math.floor(Math.random() * 50) + 5,
        }));
        
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            organizations,
            total: 100,
            page: 1,
            pageSize: 20,
          })
        });
      });
      
      // Verify pagination controls
      await expect(page.getByTestId('pagination-controls')).toBeVisible();
      await expect(page.getByText('1-20 of 100')).toBeVisible();
      
      // Test next page
      const nextButton = page.getByTestId('pagination-next');
      await nextButton.click();
      
      // Verify page changed
      await expect(page.getByText('21-40 of 100')).toBeVisible();
      
      // Test page size change
      const pageSizeSelect = page.getByTestId('page-size-select');
      await pageSizeSelect.selectOption('50');
      
      // Verify more items shown
      await expect(page.getByText('1-50 of 100')).toBeVisible();
    });
  });

  test.describe('Browser Compatibility', () => {
    test('handle unsupported browser features', async ({ page, browserName }) => {
      if (browserName === 'webkit') {
        // Simulate missing feature
        await page.addInitScript(() => {
          // @ts-ignore
          delete window.ResizeObserver;
        });
      }
      
      await loginPage.goto();
      
      // Should show compatibility warning if feature missing
      if (browserName === 'webkit') {
        await expect(page.getByText(/browser.*not fully supported|some features/i)).toBeVisible();
      }
      
      // App should still be functional
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await loginPage.isLoggedIn();
    });

    test('handle browser storage quota exceeded', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      
      // Fill localStorage to simulate quota exceeded
      await page.evaluate(() => {
        try {
          const largeData = 'x'.repeat(1024 * 1024); // 1MB string
          for (let i = 0; i < 10; i++) {
            localStorage.setItem(`test-${i}`, largeData);
          }
        } catch (e) {
          // Quota exceeded
        }
      });
      
      // Try to save preferences
      await page.getByTestId('user-menu-button').click();
      await page.getByTestId('user-preferences').click();
      
      const themeToggle = page.getByTestId('theme-toggle');
      await themeToggle.click();
      
      // Should show storage error
      await expect(page.getByText(/storage.*full|quota.*exceeded/i)).toBeVisible();
      
      // Should offer to clear old data
      await expect(page.getByTestId('clear-storage-button')).toBeVisible();
    });
  });

  test.describe('Recovery and Resilience', () => {
    test('auto-save and recovery for forms', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      await workspacePage.openProject('Test Project');
      
      // Start creating application
      const deployButton = page.getByTestId('deploy-application-button');
      await deployButton.click();
      
      const dialog = page.getByRole('dialog');
      
      // Fill form partially
      await dialog.getByTestId('app-name-input').fill('test-recovery-app');
      await dialog.getByTestId('app-image-input').fill('nginx:latest');
      await dialog.getByTestId('app-replicas-input').fill('3');
      
      // Simulate page reload (crash/refresh)
      await page.reload();
      
      // Open deploy dialog again
      await deployButton.click();
      
      // Should show recovery option
      await expect(page.getByText(/recover.*unsaved|restore.*form/i)).toBeVisible();
      
      const recoverButton = page.getByTestId('recover-form-button');
      await recoverButton.click();
      
      // Verify form data restored
      await expect(dialog.getByTestId('app-name-input')).toHaveValue('test-recovery-app');
      await expect(dialog.getByTestId('app-image-input')).toHaveValue('nginx:latest');
      await expect(dialog.getByTestId('app-replicas-input')).toHaveValue('3');
    });

    test('handle partial data corruption', async ({ page }) => {
      await loginPage.goto();
      await loginPage.login(testUsers.admin.email, testUsers.admin.password);
      
      // Corrupt some cached data
      await page.evaluate(() => {
        localStorage.setItem('workspace-cache', '{invalid json');
        sessionStorage.setItem('user-preferences', 'corrupted');
      });
      
      // Navigate should handle corrupted cache
      await dashboardPage.openWorkspace(testWorkspaces[0].name);
      
      // Should not crash, should fetch fresh data
      await expect(workspacePage.workspaceHeader).toBeVisible();
      
      // Verify cache was cleared and rebuilt
      const cacheData = await page.evaluate(() => localStorage.getItem('workspace-cache'));
      expect(() => JSON.parse(cacheData!)).not.toThrow();
    });
  });
});