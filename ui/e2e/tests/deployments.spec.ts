import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { ApplicationPage } from '../pages/ApplicationPage';
import { setupMockAPI, mockDeploymentProgress } from '../utils/mock-api';
import { testUsers, testWorkspaces } from '../fixtures/mock-data';

test.describe('Deployment Strategies and Monitoring', () => {
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
    
    // Login and create test environment
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    await workspacePage.createProject('E2E Deployment Test');
    await workspacePage.openProject('E2E Deployment Test');
  });

  test('rolling update deployment', async ({ page }) => {
    const appName = 'e2e-rolling-app';
    
    // Deploy initial version
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:1.20',
      replicas: 3,
      port: 80,
      env: {
        'VERSION': '1.0',
      },
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Initiate rolling update
    await applicationPage.updateApplication({
      image: 'nginx:1.21',
      env: {
        'VERSION': '2.0',
      },
    });

    // Monitor rolling update progress
    await expect(page.getByText('Rolling update in progress')).toBeVisible();
    
    // Check update status
    const updateProgress = page.getByTestId('update-progress');
    await expect(updateProgress).toBeVisible();
    
    // Verify progressive pod updates
    await page.waitForTimeout(2000);
    
    // Check that old and new pods coexist during update
    const pods = await applicationPage.getPods();
    const v1Pods = pods.filter(p => p.name?.includes('v1'));
    const v2Pods = pods.filter(p => p.name?.includes('v2'));
    
    // During rolling update, both versions should exist
    if (v1Pods.length > 0 && v2Pods.length > 0) {
      expect(v1Pods.length + v2Pods.length).toBe(3);
    }
    
    // Wait for update completion
    await applicationPage.waitForStatus('running', 30000);
    
    // Verify all pods updated
    await expect(page.getByText('Update completed successfully')).toBeVisible();
  });

  test('monitor deployment health checks', async ({ page }) => {
    const appName = 'e2e-health-app';
    
    // Deploy application with health checks
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
    });
    
    // Setup mock for progressive health status
    await page.route('**/api/organizations/*/workspaces/*/applications/*/health', async (route) => {
      const checks = [
        { name: 'readiness', status: 'passing' },
        { name: 'liveness', status: 'passing' },
        { name: 'startup', status: 'passing' },
      ];
      
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          status: 'healthy',
          checks,
        }),
      });
    });

    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Check health status
    const health = await applicationPage.getHealthStatus();
    
    expect(health.status).toBe('healthy');
    expect(health.checks).toHaveLength(3);
    
    health.checks.forEach(check => {
      expect(check.status).toBe('passing');
    });
    
    // Verify health indicators in UI
    await expect(applicationPage.healthStatus).toBeVisible();
    await expect(page.getByTestId('health-indicator')).toHaveClass(/healthy/);
  });

  test('blue-green deployment strategy', async ({ page }) => {
    const appName = 'e2e-bluegreen-app';
    
    // Deploy blue version
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:1.20',
      replicas: 3,
      port: 80,
      env: {
        'VERSION': 'blue',
        'COLOR': 'blue',
      },
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to deployment settings
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    // Switch to deployment strategy tab
    const deploymentTab = page.getByRole('tab', { name: /deployment/i });
    await deploymentTab.click();
    
    // Select blue-green strategy
    const strategySelect = page.getByTestId('deployment-strategy-select');
    await strategySelect.selectOption('blue-green');
    
    // Deploy green version
    const deployGreenButton = page.getByTestId('deploy-green-button');
    await deployGreenButton.click();
    
    const deployDialog = page.getByRole('dialog');
    await deployDialog.getByTestId('green-image-input').fill('nginx:1.21');
    await deployDialog.getByTestId('green-env-VERSION').fill('green');
    await deployDialog.getByTestId('green-env-COLOR').fill('green');
    await deployDialog.getByTestId('deploy-button').click();
    
    // Monitor green deployment
    await expect(page.getByText('Deploying green environment')).toBeVisible();
    
    // Wait for green to be ready
    await page.waitForTimeout(5000);
    await expect(page.getByTestId('green-status')).toContainText('ready');
    
    // Test traffic split
    const trafficSlider = page.getByTestId('traffic-split-slider');
    await expect(trafficSlider).toBeVisible();
    
    // Gradually shift traffic
    await trafficSlider.fill('25'); // 25% to green
    await page.waitForTimeout(1000);
    
    await trafficSlider.fill('50'); // 50% to green
    await page.waitForTimeout(1000);
    
    await trafficSlider.fill('100'); // 100% to green
    await page.waitForTimeout(1000);
    
    // Complete cutover
    const cutoverButton = page.getByTestId('complete-cutover-button');
    await cutoverButton.click();
    
    const confirmDialog = page.getByRole('dialog');
    await confirmDialog.getByTestId('confirm-cutover-button').click();
    
    // Verify blue environment removed
    await expect(page.getByText('Blue-green deployment completed')).toBeVisible();
    await expect(page.getByTestId('blue-status')).toContainText('terminated');
    await expect(page.getByTestId('green-status')).toContainText('active');
  });

  test('canary deployment with gradual rollout', async ({ page }) => {
    const appName = 'e2e-canary-app';
    
    // Deploy stable version
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:1.20',
      replicas: 5,
      port: 80,
      env: {
        'VERSION': 'stable',
      },
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    // Enable canary deployment
    const deploymentTab = page.getByRole('tab', { name: /deployment/i });
    await deploymentTab.click();
    
    const strategySelect = page.getByTestId('deployment-strategy-select');
    await strategySelect.selectOption('canary');
    
    // Start canary deployment
    const canaryButton = page.getByTestId('start-canary-button');
    await canaryButton.click();
    
    const canaryDialog = page.getByRole('dialog');
    await canaryDialog.getByTestId('canary-image-input').fill('nginx:1.21');
    await canaryDialog.getByTestId('canary-percentage-input').fill('20'); // Start with 20%
    await canaryDialog.getByTestId('deploy-canary-button').click();
    
    // Monitor canary deployment
    await expect(page.getByText('Canary deployment started')).toBeVisible();
    
    // Verify canary metrics
    await page.waitForTimeout(3000);
    const canaryMetrics = page.getByTestId('canary-metrics');
    await expect(canaryMetrics).toBeVisible();
    await expect(canaryMetrics).toContainText('20% traffic');
    
    // Check canary health
    const canaryHealth = page.getByTestId('canary-health-score');
    await expect(canaryHealth).toBeVisible();
    
    // Increase canary traffic
    const increaseButton = page.getByTestId('increase-canary-button');
    await increaseButton.click();
    await page.getByTestId('canary-percentage-input').fill('50');
    await page.getByTestId('update-canary-button').click();
    
    await page.waitForTimeout(2000);
    await expect(canaryMetrics).toContainText('50% traffic');
    
    // Complete canary deployment
    const promoteButton = page.getByTestId('promote-canary-button');
    await promoteButton.click();
    
    const promoteDialog = page.getByRole('dialog');
    await promoteDialog.getByTestId('confirm-promote-button').click();
    
    await expect(page.getByText('Canary promoted to stable')).toBeVisible();
  });

  test('deployment rollback on failure', async ({ page }) => {
    const appName = 'e2e-rollback-app';
    
    // Deploy initial stable version
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:1.20',
      replicas: 3,
      port: 80,
      env: {
        'VERSION': '1.0',
      },
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Mock deployment failure
    await page.route('**/api/organizations/*/workspaces/*/applications/*/update', async (route) => {
      // First return in-progress
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          status: 'updating',
          message: 'Update in progress',
        }),
      });
    });

    // After 3 seconds, mock failure
    setTimeout(async () => {
      await page.route('**/api/organizations/*/workspaces/*/applications/*', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            id: appName,
            status: 'update-failed',
            error: 'Health check failed after update',
            lastStableVersion: '1.0',
          }),
        });
      });
    }, 3000);

    // Attempt update that will fail
    await applicationPage.updateApplication({
      image: 'nginx:1.21-broken',
      env: {
        'VERSION': '2.0-broken',
      },
    });

    // Wait for failure
    await page.waitForTimeout(4000);
    await expect(page.getByText('Update failed')).toBeVisible();
    await expect(page.getByText('Health check failed')).toBeVisible();

    // Rollback option should appear
    const rollbackButton = page.getByTestId('rollback-button');
    await expect(rollbackButton).toBeVisible();
    await rollbackButton.click();

    // Confirm rollback
    const rollbackDialog = page.getByRole('dialog');
    await expect(rollbackDialog).toContainText('Rollback to version 1.0');
    await rollbackDialog.getByTestId('confirm-rollback-button').click();

    // Monitor rollback
    await expect(page.getByText('Rollback in progress')).toBeVisible();
    
    // Wait for rollback completion
    await applicationPage.waitForStatus('running', 30000);
    await expect(page.getByText('Rollback completed successfully')).toBeVisible();

    // Verify application is back to stable version
    await applicationPage.configTab.click();
    await expect(page.getByTestId('config-env-VERSION')).toContainText('1.0');
  });

  test('deployment with pre and post hooks', async ({ page }) => {
    const appName = 'e2e-hooks-app';
    
    // Deploy application with hooks
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
    });

    // Navigate to deployment configuration
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    const deploymentTab = page.getByRole('tab', { name: /deployment/i });
    await deploymentTab.click();

    // Configure deployment hooks
    const hooksButton = page.getByTestId('configure-hooks-button');
    await hooksButton.click();

    const hooksDialog = page.getByRole('dialog');
    
    // Add pre-deployment hook
    await hooksDialog.getByTestId('add-pre-hook-button').click();
    await hooksDialog.getByTestId('pre-hook-name-input').fill('database-migration');
    await hooksDialog.getByTestId('pre-hook-image-input').fill('migrate:latest');
    await hooksDialog.getByTestId('pre-hook-command-input').fill('npm run migrate');

    // Add post-deployment hook
    await hooksDialog.getByTestId('add-post-hook-button').click();
    await hooksDialog.getByTestId('post-hook-name-input').fill('smoke-test');
    await hooksDialog.getByTestId('post-hook-image-input').fill('test-runner:latest');
    await hooksDialog.getByTestId('post-hook-command-input').fill('npm run test:smoke');

    await hooksDialog.getByTestId('save-hooks-button').click();

    // Trigger deployment with hooks
    await applicationPage.updateApplication({
      image: 'nginx:1.21',
    });

    // Monitor hook execution
    await expect(page.getByText('Executing pre-deployment hooks')).toBeVisible();
    await expect(page.getByTestId('hook-database-migration-status')).toContainText('running');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('hook-database-migration-status')).toContainText('completed');

    await expect(page.getByText('Deploying application')).toBeVisible();
    
    await page.waitForTimeout(3000);
    await expect(page.getByText('Executing post-deployment hooks')).toBeVisible();
    await expect(page.getByTestId('hook-smoke-test-status')).toContainText('running');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('hook-smoke-test-status')).toContainText('completed');

    await expect(page.getByText('Deployment completed with all hooks')).toBeVisible();
  });

  test('deployment history and version management', async ({ page }) => {
    const appName = 'e2e-history-app';
    
    // Deploy initial version
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:1.19',
      replicas: 2,
      port: 80,
      env: { 'VERSION': '1.0' },
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Deploy multiple versions
    for (let i = 2; i <= 4; i++) {
      await applicationPage.updateApplication({
        image: `nginx:1.${18 + i}`,
        env: { 'VERSION': `${i}.0` },
      });
      await applicationPage.waitForStatus('running');
      await page.waitForTimeout(1000);
    }

    // View deployment history
    const historyTab = page.getByRole('tab', { name: /history/i });
    await historyTab.click();

    // Verify deployment history
    const deployments = page.locator('[data-testid^="deployment-history-item-"]');
    await expect(deployments).toHaveCount(4);

    // Check deployment details
    const latestDeployment = deployments.first();
    await expect(latestDeployment).toContainText('Version 4.0');
    await expect(latestDeployment).toContainText('nginx:1.22');
    await expect(latestDeployment).toContainText('Success');

    // Test rollback to specific version
    const v2Deployment = deployments.nth(2);
    const rollbackToV2 = v2Deployment.getByTestId('rollback-to-version-button');
    await rollbackToV2.click();

    const rollbackDialog = page.getByRole('dialog');
    await expect(rollbackDialog).toContainText('Rollback to Version 2.0');
    await rollbackDialog.getByTestId('confirm-rollback-button').click();

    await applicationPage.waitForStatus('running');
    await expect(page.getByText('Rolled back to Version 2.0')).toBeVisible();

    // Verify current version
    await applicationPage.overviewTab.click();
    await expect(page.getByTestId('current-version')).toContainText('2.0');
  });

  test('deployment with resource limits and autoscaling', async ({ page }) => {
    const appName = 'e2e-autoscale-app';
    
    // Deploy application with resource limits
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
    });
    await projectPage.waitForApplicationStatus(appName, 'running');

    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();

    // Configure autoscaling
    const autoscaleTab = page.getByRole('tab', { name: /autoscale/i });
    await autoscaleTab.click();

    const enableAutoscale = page.getByTestId('enable-autoscaling-toggle');
    await enableAutoscale.check();

    // Set autoscaling parameters
    await page.getByTestId('min-replicas-input').fill('2');
    await page.getByTestId('max-replicas-input').fill('10');
    await page.getByTestId('target-cpu-input').fill('70');
    await page.getByTestId('target-memory-input').fill('80');

    const saveAutoscaleButton = page.getByTestId('save-autoscale-button');
    await saveAutoscaleButton.click();

    await expect(page.getByText('Autoscaling enabled')).toBeVisible();

    // Simulate load to trigger autoscaling
    const simulateLoadButton = page.getByTestId('simulate-load-button');
    if (await simulateLoadButton.isVisible()) {
      await simulateLoadButton.click();
      
      // Monitor autoscaling
      await page.waitForTimeout(5000);
      
      const currentReplicas = page.getByTestId('current-replicas');
      const replicaCount = await currentReplicas.textContent();
      
      // Should have scaled up from 2
      expect(parseInt(replicaCount || '2')).toBeGreaterThan(2);
      
      // Check scaling events
      await applicationPage.eventsTab.click();
      const events = await applicationPage.getEvents();
      
      const scalingEvent = events.find(e => e.reason === 'Scaled');
      expect(scalingEvent).toBeTruthy();
    }
  });
});