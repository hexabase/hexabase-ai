import { test, expect } from '@playwright/test';
import { TestDataManager } from '../utils/test-data-manager';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { ApplicationBuilder } from '../fixtures/generators/application-generator';

test.describe('Data-Driven Tests with Generated Fixtures', () => {
  let testData: TestDataManager;
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let workspacePage: WorkspacePage;
  let projectPage: ProjectPage;
  
  test.beforeEach(async ({ page }) => {
    testData = new TestDataManager(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    workspacePage = new WorkspacePage(page);
    projectPage = new ProjectPage(page);
  });
  
  test('startup scenario - verify complete environment', async ({ page }) => {
    // Load startup scenario with consistent seed for reproducibility
    const scenario = await testData.loadScenario('startup', 12345);
    
    // Login as admin
    const adminCreds = testData.getAdminCredentials();
    await loginPage.goto();
    await loginPage.login(adminCreds.email, adminCreds.password);
    await loginPage.isLoggedIn();
    
    // Verify organization
    const org = scenario.organizations[0];
    await expect(page.getByTestId('org-selector')).toContainText(org.name);
    await expect(page.getByTestId('org-plan-badge')).toContainText('Professional');
    
    // Verify workspaces
    const workspaces = testData.getWorkspaces();
    expect(workspaces).toHaveLength(2);
    
    for (const workspace of workspaces) {
      const workspaceCard = page.getByTestId(`workspace-card-${workspace.id}`);
      await expect(workspaceCard).toBeVisible();
      await expect(workspaceCard).toContainText(workspace.name);
      await expect(workspaceCard).toContainText(workspace.region);
    }
    
    // Navigate to production workspace
    const prodWorkspace = testData.findWorkspaceByName('Production');
    await dashboardPage.openWorkspace(prodWorkspace!.name);
    
    // Verify projects
    const projects = scenario.projects.filter(p => p.workspaceId === prodWorkspace!.id);
    for (const project of projects) {
      await expect(page.getByText(project.name)).toBeVisible();
    }
    
    // Open API Services project
    await workspacePage.openProject('API Services');
    
    // Verify applications
    const apiProject = testData.findProjectByName('API Services');
    const apps = scenario.applications.filter(a => a.projectId === apiProject!.id);
    
    for (const app of apps) {
      const appCard = await projectPage.getApplicationCard(app.name);
      await expect(appCard).toBeVisible();
      
      // Verify app status
      const statusBadge = appCard.getByTestId('app-status');
      await expect(statusBadge).toHaveClass(/running/);
      
      // Verify app type icon
      if (app.type === 'stateful') {
        await expect(appCard.getByTestId('stateful-icon')).toBeVisible();
      }
    }
    
    // Verify monitoring alerts
    await projectPage.monitoringTab.click();
    
    const alerts = scenario.monitoring.alerts;
    const activeAlerts = alerts.filter(a => a.status === 'active');
    const alertCount = page.getByTestId('active-alerts-count');
    await expect(alertCount).toContainText(activeAlerts.length.toString());
  });
  
  test('enterprise scenario - complex operations', async ({ page }) => {
    // Load enterprise scenario
    await testData.loadScenario('enterprise', 67890);
    
    const adminCreds = testData.getAdminCredentials();
    await loginPage.goto();
    await loginPage.login(adminCreds.email, adminCreds.password);
    
    // Verify multiple organizations
    const orgs = testData.getOrganizations();
    await page.getByTestId('org-switcher').click();
    
    const orgDropdown = page.getByTestId('org-dropdown');
    for (const org of orgs) {
      await expect(orgDropdown).toContainText(org.name);
    }
    await page.keyboard.press('Escape');
    
    // Navigate to production workspace
    const prodWorkspace = testData.findWorkspaceByName('Production US-East');
    await dashboardPage.openWorkspace(prodWorkspace!.name);
    
    // Verify dedicated resources
    await expect(page.getByTestId('workspace-plan')).toContainText('Dedicated');
    await expect(page.getByTestId('workspace-nodes')).toContainText('5 nodes');
    
    // Check backup status
    const backupPolicies = testData.getScenario().backups.policies;
    const activeBackups = backupPolicies.filter(p => p.enabled);
    
    await page.getByTestId('backup-status').click();
    await expect(page.getByTestId('active-policies')).toContainText(`${activeBackups.length} active`);
    
    // Verify high availability applications
    await workspacePage.openProject('Core Services');
    
    const haApps = testData.getApplications().filter(a => 
      a.autoscaling?.enabled && a.autoscaling.minReplicas >= 3
    );
    
    for (const app of haApps.slice(0, 3)) { // Check first 3
      const appCard = await projectPage.getApplicationCard(app.name);
      await expect(appCard.getByTestId('ha-badge')).toBeVisible();
      await expect(appCard.getByTestId('replicas')).toContainText(/\d+\/\d+/);
    }
  });
  
  test('dynamic application deployment with custom generator', async ({ page }) => {
    // Load scenario
    await testData.loadScenario('startup', 11111);
    
    // Login
    const devCreds = testData.getDeveloperCredentials();
    await loginPage.goto();
    await loginPage.login(devCreds.email, devCreds.password);
    
    // Navigate to project
    const workspace = testData.getWorkspaces()[0];
    const project = testData.getProjects()[0];
    
    await dashboardPage.openWorkspace(workspace.name);
    await workspacePage.openProject(project.name);
    
    // Create custom application using builder
    const appBuilder = new ApplicationBuilder();
    const customApp = appBuilder
      .withName('ml-inference-api')
      .withImage('tensorflow/serving', '2.14.0')
      .withType('stateless')
      .withReplicas(3)
      .withPort(8501)
      .withResources(
        { cpu: '2', memory: '4Gi' },
        { cpu: '4', memory: '8Gi' }
      )
      .withEnvironment({
        MODEL_NAME: 'sentiment_analysis',
        MODEL_VERSION: 'v2',
        ENABLE_BATCHING: 'true',
      })
      .withIngress('ml-api.example.com', true)
      .withAutoscaling(2, 10, 60)
      .build();
    
    // Deploy the application
    await projectPage.deployApplication({
      name: customApp.name,
      type: customApp.type,
      image: `${customApp.image}:${customApp.tag}`,
      replicas: customApp.replicas,
      port: customApp.port,
    });
    
    // Verify deployment
    await projectPage.waitForApplicationStatus(customApp.name, 'running');
    
    const appCard = await projectPage.getApplicationCard(customApp.name);
    await expect(appCard).toBeVisible();
    
    // Verify autoscaling is configured
    await appCard.click();
    const appDetails = page.getByTestId('app-details');
    await expect(appDetails).toContainText('Autoscaling: Enabled');
    await expect(appDetails).toContainText('Min: 2, Max: 10');
  });
  
  test('scenario comparison - startup vs enterprise', async ({ page }) => {
    // Load both scenarios with same seed for fair comparison
    const seed = 99999;
    
    // First check startup
    const startupScenario = await testData.loadScenario('startup', seed);
    const startupMetrics = {
      totalApps: startupScenario.applications.length,
      totalResources: startupScenario.workspaces.reduce((sum, w) => 
        sum + parseInt(w.resources.cpu), 0
      ),
      backupPolicies: startupScenario.backups.policies.length,
      activeAlerts: startupScenario.monitoring.alerts.filter(a => a.status === 'active').length,
    };
    
    // Then check enterprise
    const enterpriseScenario = await testData.loadScenario('enterprise', seed);
    const enterpriseMetrics = {
      totalApps: enterpriseScenario.applications.length,
      totalResources: enterpriseScenario.workspaces.reduce((sum, w) => 
        sum + parseInt(w.resources.cpu), 0
      ),
      backupPolicies: enterpriseScenario.backups.policies.length,
      activeAlerts: enterpriseScenario.monitoring.alerts.filter(a => a.status === 'active').length,
    };
    
    // Verify enterprise has significantly more resources
    expect(enterpriseMetrics.totalApps).toBeGreaterThan(startupMetrics.totalApps * 3);
    expect(enterpriseMetrics.totalResources).toBeGreaterThan(startupMetrics.totalResources * 5);
    expect(enterpriseMetrics.backupPolicies).toBeGreaterThan(startupMetrics.backupPolicies);
    
    // Log comparison for debugging
    console.log('Scenario Comparison:', {
      startup: startupMetrics,
      enterprise: enterpriseMetrics,
      ratios: {
        apps: (enterpriseMetrics.totalApps / startupMetrics.totalApps).toFixed(2),
        resources: (enterpriseMetrics.totalResources / startupMetrics.totalResources).toFixed(2),
      },
    });
  });
  
  test('simulate progressive deployment with generated data', async ({ page }) => {
    await testData.loadScenario('startup');
    
    const adminCreds = testData.getAdminCredentials();
    await loginPage.goto();
    await loginPage.login(adminCreds.email, adminCreds.password);
    
    // Get an application to update
    const workspace = testData.getWorkspaces()[0];
    const project = testData.getProjects()[0];
    const app = testData.getApplications().find(a => a.type === 'stateless')!;
    
    await dashboardPage.openWorkspace(workspace.name);
    await workspacePage.openProject(project.name);
    
    // Start canary deployment
    const appCard = await projectPage.getApplicationCard(app.name);
    await appCard.click();
    
    await page.getByTestId('deploy-new-version').click();
    
    const deployDialog = page.getByRole('dialog');
    await deployDialog.getByTestId('deployment-strategy').selectOption('canary');
    await deployDialog.getByTestId('new-image-tag').fill('v2.0.0');
    await deployDialog.getByTestId('canary-weight').fill('10');
    await deployDialog.getByTestId('deploy-button').click();
    
    // Monitor canary progress
    await expect(page.getByTestId('deployment-status')).toContainText('Canary: 10%');
    
    // Simulate metrics check
    await page.getByTestId('canary-metrics').click();
    await expect(page.getByTestId('error-rate')).toContainText(/0\.\d+%/);
    await expect(page.getByTestId('latency-p95')).toBeVisible();
    
    // Promote canary
    await page.getByTestId('promote-canary').click();
    await page.getByTestId('confirm-promotion').click();
    
    await expect(page.getByTestId('deployment-status')).toContainText('Rolling out: 100%');
  });
});