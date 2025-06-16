import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers, testOrganizations, testWorkspaces } from '../fixtures/mock-data';

test.describe('Project Management', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let workspacePage: WorkspacePage;
  let projectPage: ProjectPage;

  test.beforeEach(async ({ page }) => {
    // Setup mock API and pages
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    workspacePage = new WorkspacePage(page);
    projectPage = new ProjectPage(page);
    
    // Login and navigate to workspace
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    // Enter first workspace
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
  });

  test('create project with default settings', async ({ page }) => {
    const projectName = 'E2E Test Project';
    
    // Create project
    await workspacePage.createProject(projectName);
    
    // Verify project appears in list
    const projectCard = await workspacePage.getProjectCard(projectName);
    await expect(projectCard).toBeVisible();
    
    // Verify project status
    await expect(projectCard.getByTestId('project-status')).toContainText('active');
    
    // Verify default resource quotas are applied
    await expect(projectCard.getByTestId('cpu-quota')).toContainText('2');
    await expect(projectCard.getByTestId('memory-quota')).toContainText('4Gi');
    await expect(projectCard.getByTestId('storage-quota')).toContainText('20Gi');
  });

  test('create project with custom resource quotas', async ({ page }) => {
    const projectName = 'E2E High Performance Project';
    const customQuotas = {
      cpu: '8',
      memory: '16Gi',
      storage: '100Gi',
    };
    
    // Create project with custom quotas
    await workspacePage.createProject(projectName, customQuotas);
    
    // Verify project is created
    const projectCard = await workspacePage.getProjectCard(projectName);
    await expect(projectCard).toBeVisible();
    
    // Verify custom quotas are applied
    await expect(projectCard.getByTestId('cpu-quota')).toContainText(customQuotas.cpu);
    await expect(projectCard.getByTestId('memory-quota')).toContainText(customQuotas.memory);
    await expect(projectCard.getByTestId('storage-quota')).toContainText(customQuotas.storage);
  });

  test('enter project and verify dashboard', async ({ page }) => {
    const projectName = 'E2E Dashboard Test';
    
    // Create and enter project
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
    
    // Verify we're in the project
    await expect(projectPage.projectHeader).toContainText(projectName);
    
    // Verify tabs are visible
    await expect(projectPage.applicationsTab).toBeVisible();
    await expect(projectPage.functionsTab).toBeVisible();
    await expect(projectPage.cronJobsTab).toBeVisible();
    await expect(projectPage.settingsTab).toBeVisible();
    await expect(projectPage.monitoringTab).toBeVisible();
    
    // Verify empty state
    await expect(page.getByText('No applications deployed yet')).toBeVisible();
    
    // Verify resource quota card
    await expect(projectPage.resourceQuotaCard).toBeVisible();
    const quota = await projectPage.getResourceQuota();
    expect(quota.limits.cpu).toBeTruthy();
    expect(quota.limits.memory).toBeTruthy();
    expect(quota.limits.storage).toBeTruthy();
  });

  test('update project resource quotas', async ({ page }) => {
    const projectName = 'E2E Quota Update Test';
    
    // Create project
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
    
    // Update quotas
    const newQuotas = {
      cpu: '4',
      memory: '8Gi',
      storage: '50Gi',
    };
    
    await projectPage.updateResourceQuota(newQuotas);
    
    // Verify success notification
    await expect(page.getByText('Resource quotas updated successfully')).toBeVisible();
    
    // Verify new quotas are applied
    const updatedQuota = await projectPage.getResourceQuota();
    expect(updatedQuota.limits.cpu).toContain(newQuotas.cpu);
    expect(updatedQuota.limits.memory).toContain(newQuotas.memory);
    expect(updatedQuota.limits.storage).toContain(newQuotas.storage);
  });

  test('deploy multiple applications in project', async ({ page }) => {
    const projectName = 'E2E Multi-App Project';
    
    // Create and enter project
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
    
    // Deploy stateless application
    const webAppConfig = {
      name: 'e2e-web-app',
      type: 'stateless' as const,
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
      env: {
        'API_URL': 'https://api.example.com',
        'DEBUG': 'false',
      },
    };
    
    await projectPage.deployApplication(webAppConfig);
    
    // Verify deployment started
    await expect(page.getByText('Deployment started')).toBeVisible();
    
    // Wait for deployment to complete
    await projectPage.waitForApplicationStatus(webAppConfig.name, 'running');
    
    // Deploy stateful application
    const dbAppConfig = {
      name: 'e2e-database',
      type: 'stateful' as const,
      image: 'postgres:14',
      port: 5432,
      storage: '10Gi',
      env: {
        'POSTGRES_USER': 'admin',
        'POSTGRES_PASSWORD': 'secure123',
        'POSTGRES_DB': 'e2e_test',
      },
    };
    
    await projectPage.deployApplication(dbAppConfig);
    await projectPage.waitForApplicationStatus(dbAppConfig.name, 'running');
    
    // Deploy CronJob
    const cronJobConfig = {
      name: 'e2e-backup-job',
      type: 'cronjob' as const,
      image: 'busybox:latest',
      schedule: '0 2 * * *',
    };
    
    await projectPage.deployApplication(cronJobConfig);
    await projectPage.waitForApplicationStatus(cronJobConfig.name, 'active');
    
    // Verify all applications are listed
    const webAppCard = await projectPage.getApplicationCard(webAppConfig.name);
    const dbAppCard = await projectPage.getApplicationCard(dbAppConfig.name);
    const cronJobCard = await projectPage.getApplicationCard(cronJobConfig.name);
    
    await expect(webAppCard).toBeVisible();
    await expect(dbAppCard).toBeVisible();
    await expect(cronJobCard).toBeVisible();
    
    // Verify resource usage updated
    const usage = await projectPage.getResourceQuota();
    expect(usage.used.cpu).not.toBe('0');
    expect(usage.used.memory).not.toBe('0');
    expect(usage.used.storage).not.toBe('0');
  });

  test('scale stateless application', async ({ page }) => {
    const projectName = 'E2E Scaling Project';
    const appName = 'e2e-scalable-app';
    
    // Create project and deploy app
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
    
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 1,
      port: 80,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Scale up
    await projectPage.scaleApplication(appName, 3);
    
    // Verify scaling notification
    await expect(page.getByText('Scaling application')).toBeVisible();
    
    // Wait for scaling to complete
    await page.waitForTimeout(2000);
    
    // Verify replicas updated
    const appCard = await projectPage.getApplicationCard(appName);
    await expect(appCard.getByTestId('app-replicas')).toContainText('3');
    
    // Scale down
    await projectPage.scaleApplication(appName, 1);
    await page.waitForTimeout(2000);
    await expect(appCard.getByTestId('app-replicas')).toContainText('1');
  });

  test('view application logs and metrics', async ({ page }) => {
    const projectName = 'E2E Monitoring Project';
    const appName = 'e2e-monitored-app';
    
    // Create project and deploy app
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
    
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // View logs
    await projectPage.viewApplicationLogs(appName);
    
    // Verify logs viewer opened
    await expect(page.getByTestId('logs-viewer')).toBeVisible();
    await expect(page.getByTestId('logs-content')).toContainText('started');
    
    // Close logs
    await page.getByTestId('close-logs-button').click();
    
    // View metrics
    await projectPage.viewApplicationMetrics(appName);
    
    // Verify metrics dashboard opened
    await expect(page.getByTestId('metrics-dashboard')).toBeVisible();
    await expect(page.getByTestId('cpu-chart')).toBeVisible();
    await expect(page.getByTestId('memory-chart')).toBeVisible();
    await expect(page.getByTestId('network-chart')).toBeVisible();
  });

  test('delete application', async ({ page }) => {
    const projectName = 'E2E Delete Test Project';
    const appName = 'e2e-deletable-app';
    
    // Create project and deploy app
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
    
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 1,
      port: 80,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Delete application
    await projectPage.deleteApplication(appName);
    
    // Verify deletion confirmation
    await expect(page.getByText('Application deleted successfully')).toBeVisible();
    
    // Verify application removed from list
    await expect(projectPage.getApplicationCard(appName)).not.toBeVisible();
    
    // Verify resource usage updated
    const usage = await projectPage.getResourceQuota();
    expect(usage.used.cpu).toBe('0');
    expect(usage.used.memory).toBe('0');
  });

  test('view project activity feed', async ({ page }) => {
    const projectName = 'E2E Activity Project';
    
    // Create project
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
    
    // Deploy an application to generate activity
    await projectPage.deployApplication({
      name: 'e2e-activity-app',
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 1,
      port: 80,
    });
    
    // Check activity feed
    const activities = await projectPage.getRecentActivity();
    
    // Verify activities recorded
    expect(activities.length).toBeGreaterThan(0);
    
    // Verify activity types
    const deployActivity = activities.find(a => a.type === 'deployment');
    expect(deployActivity).toBeTruthy();
    expect(deployActivity?.message).toContain('deployed');
  });

  test('handle quota exceeded error', async ({ page }) => {
    const projectName = 'E2E Quota Test';
    
    // Create project with small quotas
    await workspacePage.createProject(projectName, {
      cpu: '1',
      memory: '1Gi',
      storage: '5Gi',
    });
    await workspacePage.openProject(projectName);
    
    // Override API to simulate quota exceeded
    await page.route('**/api/organizations/*/workspaces/*/applications', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 403,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'Resource quota exceeded: CPU limit reached',
          }),
        });
      }
    });
    
    // Try to deploy application
    await projectPage.deployApplication({
      name: 'e2e-too-large-app',
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 10,
      port: 80,
    });
    
    // Verify error message
    await expect(page.getByText('Resource quota exceeded')).toBeVisible();
    await expect(page.getByText('CPU limit reached')).toBeVisible();
  });

  test('delete project', async ({ page }) => {
    const projectName = 'E2E Project to Delete';
    
    // Create project
    await workspacePage.createProject(projectName);
    
    // Delete project
    await workspacePage.deleteProject(projectName);
    
    // Verify deletion confirmation
    await expect(page.getByText('Project deleted successfully')).toBeVisible();
    
    // Verify project removed from list
    await expect(workspacePage.getProjectCard(projectName)).not.toBeVisible();
  });
});