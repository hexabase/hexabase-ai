import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { ApplicationPage } from '../pages/ApplicationPage';
import { setupMockAPI, mockDeploymentProgress } from '../utils/mock-api';
import { testUsers, testWorkspaces } from '../fixtures/mock-data';

test.describe('Application Deployment and Management', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let workspacePage: WorkspacePage;
  let projectPage: ProjectPage;
  let applicationPage: ApplicationPage;

  test.beforeEach(async ({ page }) => {
    // Setup mock API and pages
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    workspacePage = new WorkspacePage(page);
    projectPage = new ProjectPage(page);
    applicationPage = new ApplicationPage(page);
    
    // Login and navigate to project
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    // Create and enter test project
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    await workspacePage.createProject('E2E App Test Project');
    await workspacePage.openProject('E2E App Test Project');
  });

  test('deploy stateless application with progressive status', async ({ page }) => {
    const appName = 'e2e-nginx-app';
    const appConfig = {
      name: appName,
      type: 'stateless' as const,
      image: 'nginx:1.21',
      replicas: 3,
      port: 80,
      env: {
        'NGINX_HOST': 'example.com',
        'NGINX_PORT': '80',
      },
    };

    // Setup progressive deployment mock
    const stopMock = await mockDeploymentProgress(page, `app-${Date.now()}`, 5000);

    // Deploy application
    await projectPage.deployApplication(appConfig);

    // Verify deployment notification
    await expect(page.getByText('Deployment started')).toBeVisible();

    // Monitor deployment progress
    const appCard = await projectPage.getApplicationCard(appName);
    
    // Check pending status
    await expect(appCard.getByTestId('app-status')).toContainText('pending');
    
    // Check provisioning
    await page.waitForTimeout(1500);
    await expect(appCard.getByTestId('app-status')).toContainText('provisioning');
    
    // Check deploying
    await page.waitForTimeout(1500);
    await expect(appCard.getByTestId('app-status')).toContainText('deploying');
    
    // Wait for running status
    await projectPage.waitForApplicationStatus(appName, 'running', 10000);
    
    // Verify application details
    await expect(appCard.getByTestId('app-type')).toContainText('Stateless');
    await expect(appCard.getByTestId('app-image')).toContainText('nginx:1.21');
    await expect(appCard.getByTestId('app-replicas')).toContainText('3');

    // Clean up mock
    stopMock();
  });

  test('deploy stateful application with persistent storage', async ({ page }) => {
    const appName = 'e2e-postgres-db';
    const appConfig = {
      name: appName,
      type: 'stateful' as const,
      image: 'postgres:14-alpine',
      port: 5432,
      storage: '20Gi',
      env: {
        'POSTGRES_USER': 'testuser',
        'POSTGRES_PASSWORD': 'testpass123',
        'POSTGRES_DB': 'e2e_test_db',
      },
    };

    // Deploy application
    await projectPage.deployApplication(appConfig);
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Verify stateful specific features
    const appCard = await projectPage.getApplicationCard(appName);
    await expect(appCard.getByTestId('app-type')).toContainText('Stateful');
    await expect(appCard.getByTestId('app-storage')).toContainText('20Gi');
    await expect(appCard.getByTestId('pvc-status')).toContainText('Bound');
  });

  test('enter application details page', async ({ page }) => {
    // Deploy a test application first
    const appName = 'e2e-detail-test-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'httpd:2.4',
      replicas: 2,
      port: 80,
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Click on application to view details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Wait for navigation to application detail page
    await page.waitForURL('**/applications/**');

    // Verify we're on the application page
    await expect(applicationPage.appHeader).toContainText(appName);
    await expect(applicationPage.statusBadge).toContainText('Running');

    // Verify tabs are visible
    await expect(applicationPage.overviewTab).toBeVisible();
    await expect(applicationPage.logsTab).toBeVisible();
    await expect(applicationPage.metricsTab).toBeVisible();
    await expect(applicationPage.configTab).toBeVisible();
    await expect(applicationPage.eventsTab).toBeVisible();
  });

  test('view and manage application pods', async ({ page }) => {
    // Deploy application
    const appName = 'e2e-pod-test-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 3,
      port: 80,
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Get pod information
    const pods = await applicationPage.getPods();
    
    // Verify 3 pods are running
    expect(pods).toHaveLength(3);
    pods.forEach(pod => {
      expect(pod.status).toBe('Running');
      expect(pod.restarts).toBe('0');
    });

    // Test pod restart
    const firstPodName = pods[0].name!;
    await applicationPage.restartPod(firstPodName);

    // Verify restart notification
    await expect(page.getByText('Pod restart initiated')).toBeVisible();

    // View pod logs
    await applicationPage.viewPodLogs(firstPodName);
    await expect(page.getByTestId('pod-logs-viewer')).toBeVisible();
    
    // Close logs viewer
    await page.getByTestId('close-logs-button').click();
  });

  test('scale application up and down', async ({ page }) => {
    // Deploy application with 1 replica
    const appName = 'e2e-scale-test-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 1,
      port: 80,
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Verify initial pod count
    let pods = await applicationPage.getPods();
    expect(pods).toHaveLength(1);

    // Scale up to 5 replicas
    await applicationPage.scaleApplication(5);
    await expect(page.getByText('Scaling application to 5 replicas')).toBeVisible();

    // Wait for scaling to complete
    await page.waitForTimeout(3000);

    // Verify new pod count
    pods = await applicationPage.getPods();
    expect(pods).toHaveLength(5);
    
    // Verify all pods are running
    pods.forEach(pod => {
      expect(pod.status).toBe('Running');
    });

    // Scale down to 2 replicas
    await applicationPage.scaleApplication(2);
    await page.waitForTimeout(3000);

    pods = await applicationPage.getPods();
    expect(pods).toHaveLength(2);
  });

  test('update application configuration', async ({ page }) => {
    // Deploy application
    const appName = 'e2e-update-test-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:1.20',
      replicas: 2,
      port: 80,
      env: {
        'VERSION': '1.0',
      },
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Update application
    await applicationPage.updateApplication({
      image: 'nginx:1.21',
      env: {
        'VERSION': '2.0',
        'NEW_VAR': 'new_value',
      },
      resources: {
        cpu: '500m',
        memory: '512Mi',
      },
    });

    // Verify update notification
    await expect(page.getByText('Application update initiated')).toBeVisible();

    // Wait for update to complete
    await applicationPage.waitForStatus('running', 30000);

    // Verify configuration updated
    await applicationPage.configTab.click();
    await expect(page.getByTestId('config-image')).toContainText('nginx:1.21');
    await expect(page.getByTestId('config-env-VERSION')).toContainText('2.0');
    await expect(page.getByTestId('config-env-NEW_VAR')).toContainText('new_value');
    await expect(page.getByTestId('config-cpu-request')).toContainText('500m');
    await expect(page.getByTestId('config-memory-request')).toContainText('512Mi');
  });

  test('view application logs with filters', async ({ page }) => {
    // Deploy application
    const appName = 'e2e-logs-test-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'busybox:latest',
      replicas: 2,
      port: 8080,
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // View logs with different options
    await applicationPage.viewLogs({
      since: '5m',
      tail: 100,
    });

    // Verify logs are displayed
    const logs = await applicationPage.getLogs();
    expect(logs).toBeTruthy();
    expect(logs?.length).toBeGreaterThan(0);

    // Test log search/filter
    const searchInput = page.getByTestId('log-search-input');
    await searchInput.fill('started');
    await page.getByTestId('search-logs-button').click();

    // Verify filtered logs
    await expect(page.getByTestId('logs-content')).toContainText('started');
  });

  test('view application metrics', async ({ page }) => {
    // Deploy application
    const appName = 'e2e-metrics-test-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // View metrics for different time ranges
    await applicationPage.viewMetrics('1h');

    // Get current metrics
    const metrics = await applicationPage.getMetrics();
    
    // Verify metrics are available
    expect(metrics.cpu).toBeTruthy();
    expect(metrics.memory).toBeTruthy();
    expect(metrics.requestRate).toBeTruthy();
    expect(metrics.errorRate).toBeTruthy();

    // Change time range
    await applicationPage.viewMetrics('24h');
    
    // Verify charts updated
    await expect(page.getByTestId('time-range-label')).toContainText('Last 24 hours');
  });

  test('view application events', async ({ page }) => {
    // Deploy application
    const appName = 'e2e-events-test-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 1,
      port: 80,
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Get events
    const events = await applicationPage.getEvents();
    
    // Verify deployment events exist
    expect(events.length).toBeGreaterThan(0);
    
    // Check for common events
    const pulledEvent = events.find(e => e.reason === 'Pulled');
    const createdEvent = events.find(e => e.reason === 'Created');
    const startedEvent = events.find(e => e.reason === 'Started');
    
    expect(pulledEvent).toBeTruthy();
    expect(createdEvent).toBeTruthy();
    expect(startedEvent).toBeTruthy();
  });

  test('restart application', async ({ page }) => {
    // Deploy application
    const appName = 'e2e-restart-test-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Get initial pod ages
    const initialPods = await applicationPage.getPods();
    
    // Restart application
    await applicationPage.restartApplication();
    
    // Verify restart notification
    await expect(page.getByText('Application restart initiated')).toBeVisible();
    
    // Wait for restart to complete
    await page.waitForTimeout(5000);
    
    // Verify new pods created
    const newPods = await applicationPage.getPods();
    expect(newPods).toHaveLength(2);
    
    // Verify all pods are fresh (age should be very recent)
    newPods.forEach(pod => {
      expect(pod.age).toMatch(/[0-9]+s|just now/i);
    });
  });

  test('export application configuration', async ({ page }) => {
    // Deploy application
    const appName = 'e2e-export-test-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
      env: {
        'APP_ENV': 'production',
      },
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Export configuration
    const filename = await applicationPage.exportConfiguration();
    
    // Verify export filename
    expect(filename).toContain(appName);
    expect(filename).toMatch(/\.(yaml|yml|json)$/);
  });

  test('delete application from details page', async ({ page }) => {
    // Deploy application
    const appName = 'e2e-delete-detail-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 1,
      port: 80,
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application details
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Delete application
    await applicationPage.deleteApplication();
    
    // Verify deletion notification
    await expect(page.getByText('Application deleted successfully')).toBeVisible();
    
    // Verify redirected back to project
    await page.waitForURL('**/projects/**');
    
    // Verify application no longer exists
    await expect(projectPage.getApplicationCard(appName)).not.toBeVisible();
  });

  test('handle application deployment failure', async ({ page }) => {
    // Override API to simulate deployment failure
    await page.route('**/api/organizations/*/workspaces/*/applications', async (route) => {
      if (route.request().method() === 'POST') {
        const body = await route.request().postDataJSON();
        
        // Return success for creation
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            id: `app-fail-${Date.now()}`,
            ...body,
            status: 'pending',
          }),
        });
      }
    });

    // Override status check to return failed
    await page.route('**/api/organizations/*/workspaces/*/applications/app-fail-*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: route.request().url().split('/').pop(),
          status: 'failed',
          error: 'Image pull failed: nginx:invalid-tag not found',
        }),
      });
    });

    // Try to deploy with invalid image
    const appName = 'e2e-fail-deploy-app';
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:invalid-tag',
      replicas: 1,
      port: 80,
    });

    // Wait for failed status
    await projectPage.waitForApplicationStatus(appName, 'failed');
    
    // Verify error message displayed
    const appCard = await projectPage.getApplicationCard(appName);
    await expect(appCard.getByTestId('app-error')).toContainText('Image pull failed');
    
    // Verify retry option available
    await expect(appCard.getByTestId('retry-deployment-button')).toBeVisible();
  });
});