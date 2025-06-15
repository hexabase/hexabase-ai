import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { ApplicationPage } from '../pages/ApplicationPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers, testOrganizations, testWorkspaces } from '../fixtures/mock-data';
import * as fs from 'fs';
import * as path from 'path';

// Create screenshot directory
const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
const screenshotDir = path.join(process.cwd(), 'screenshots', `e2e_result_${timestamp}`);

// Ensure directories exist
const ensureDir = (dir: string) => {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
};

ensureDir(screenshotDir);

// Helper to capture screenshots
async function captureScreenshot(page: any, category: string, name: string) {
  ensureDir(path.join(screenshotDir, category));
  const fileName = `${name.replace(/\s+/g, '_')}.png`;
  const filePath = path.join(screenshotDir, category, fileName);
  
  await page.screenshot({
    path: filePath,
    fullPage: true,
  });
  
  console.log(`📸 Screenshot saved: ${filePath}`);
}

test.describe('E2E Screenshot Demo', () => {
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

  test('complete user journey with screenshots', async ({ page }) => {
    // 1. Login Flow
    await loginPage.goto();
    await captureScreenshot(page, 'auth', '01_login_page');
    
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    await captureScreenshot(page, 'auth', '02_login_success');

    // 2. Dashboard View
    await expect(dashboardPage.welcomeMessage).toContainText('Welcome');
    await captureScreenshot(page, 'dashboard', '01_dashboard_overview');
    
    // Show organizations
    const orgCards = await dashboardPage.getOrganizationCards();
    expect(orgCards.length).toBeGreaterThan(0);
    await captureScreenshot(page, 'dashboard', '02_organizations_list');

    // 3. Enter Organization
    await dashboardPage.selectOrganization(testOrganizations[0].name);
    await expect(page.getByText(testOrganizations[0].name)).toBeVisible();
    await captureScreenshot(page, 'organization', '01_organization_selected');

    // Show workspaces
    const workspaceCards = await dashboardPage.getWorkspaceCards();
    expect(workspaceCards.length).toBeGreaterThan(0);
    await captureScreenshot(page, 'organization', '02_workspaces_list');

    // 4. Enter Workspace
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    await expect(workspacePage.workspaceHeader).toContainText(testWorkspaces[0].name);
    await captureScreenshot(page, 'workspace', '01_workspace_dashboard');

    // 5. Create Project
    const projectName = 'Demo E2E Project';
    await workspacePage.createProject(projectName);
    await captureScreenshot(page, 'projects', '01_project_created');
    
    await workspacePage.openProject(projectName);
    await expect(projectPage.projectHeader).toContainText(projectName);
    await captureScreenshot(page, 'projects', '02_project_dashboard');

    // 6. Deploy Application
    const appConfig = {
      name: 'demo-nginx-app',
      type: 'stateless' as const,
      image: 'nginx:latest',
      replicas: 3,
      port: 80,
      env: {
        'APP_ENV': 'production',
        'VERSION': '1.0.0',
      },
    };

    await projectPage.deployApplication(appConfig);
    await captureScreenshot(page, 'applications', '01_deployment_started');
    
    // Wait for deployment
    await projectPage.waitForApplicationStatus(appConfig.name, 'running');
    await captureScreenshot(page, 'applications', '02_deployment_complete');

    // 7. View Application Details
    const appCard = await projectPage.getApplicationCard(appConfig.name);
    await appCard.click();
    
    await expect(applicationPage.appHeader).toContainText(appConfig.name);
    await expect(applicationPage.statusBadge).toContainText('Running');
    await captureScreenshot(page, 'applications', '03_application_details');

    // View pods
    const pods = await applicationPage.getPods();
    expect(pods).toHaveLength(3);
    await captureScreenshot(page, 'applications', '04_pods_list');

    // View logs
    await applicationPage.logsTab.click();
    await page.waitForTimeout(1000);
    await captureScreenshot(page, 'applications', '05_application_logs');

    // View metrics
    await applicationPage.metricsTab.click();
    await page.waitForTimeout(1000);
    await captureScreenshot(page, 'applications', '06_application_metrics');

    // 8. Scale Application
    await applicationPage.overviewTab.click();
    await applicationPage.scaleApplication(5);
    await page.waitForTimeout(2000);
    await captureScreenshot(page, 'applications', '07_scaled_to_5_replicas');

    // 9. Update Application
    await applicationPage.updateApplication({
      image: 'nginx:1.21',
      env: {
        'VERSION': '2.0.0',
      },
    });
    await page.waitForTimeout(2000);
    await captureScreenshot(page, 'deployments', '01_rolling_update_progress');
    
    await applicationPage.waitForStatus('running');
    await captureScreenshot(page, 'deployments', '02_update_complete');

    // 10. Final Success
    await captureScreenshot(page, 'success', 'complete_e2e_journey_success');
    
    console.log(`\n✅ All screenshots saved to: ${screenshotDir}\n`);
  });

  test('CI/CD pipeline configuration with screenshots', async ({ page }) => {
    // Login and navigate
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    await workspacePage.createProject('CI/CD Demo Project');
    await workspacePage.openProject('CI/CD Demo Project');
    
    // Navigate to CI/CD settings
    await projectPage.settingsTab.click();
    await captureScreenshot(page, 'cicd', '01_project_settings');
    
    const cicdTab = page.getByRole('tab', { name: /ci\/cd/i });
    await cicdTab.click();
    await captureScreenshot(page, 'cicd', '02_cicd_settings');
    
    // Connect repository
    const connectRepoButton = page.getByTestId('connect-repository-button');
    await connectRepoButton.click();
    await page.waitForTimeout(500);
    await captureScreenshot(page, 'cicd', '03_connect_repository_dialog');
    
    const repoDialog = page.getByRole('dialog');
    await repoDialog.getByTestId('provider-github').click();
    await repoDialog.getByTestId('repo-url-input').fill('https://github.com/hexabase/sample-app');
    await repoDialog.getByTestId('default-branch-input').fill('main');
    await captureScreenshot(page, 'cicd', '04_repository_configured');
    
    // Close dialog for demo
    await page.keyboard.press('Escape');
    
    console.log(`\n✅ CI/CD screenshots saved to: ${screenshotDir}/cicd\n`);
  });

  test('serverless function creation with screenshots', async ({ page }) => {
    // Login and navigate
    await loginPage.goto();
    await loginPage.login(testUsers.developer.email, testUsers.developer.password);
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    await workspacePage.createProject('Serverless Demo');
    await workspacePage.openProject('Serverless Demo');
    
    // Navigate to functions
    await projectPage.functionsTab.click();
    await captureScreenshot(page, 'serverless', '01_functions_list_empty');
    
    // Create function
    const createFunctionButton = page.getByTestId('create-function-button');
    await createFunctionButton.click();
    await page.waitForTimeout(500);
    await captureScreenshot(page, 'serverless', '02_create_function_dialog');
    
    const functionDialog = page.getByRole('dialog');
    await functionDialog.getByTestId('function-name-input').fill('hello-api');
    await functionDialog.getByTestId('runtime-select').selectOption('nodejs18');
    await functionDialog.getByTestId('trigger-type-http').click();
    await captureScreenshot(page, 'serverless', '03_function_configured');
    
    // Add code
    const codeEditor = functionDialog.getByTestId('function-code-editor');
    await codeEditor.click();
    await codeEditor.fill(`exports.handler = async (event) => {
  return {
    statusCode: 200,
    body: JSON.stringify({ message: 'Hello from Hexabase!' })
  };
};`);
    await captureScreenshot(page, 'serverless', '04_function_code_added');
    
    console.log(`\n✅ Serverless function screenshots saved to: ${screenshotDir}/serverless\n`);
  });

  test('backup configuration for dedicated workspace', async ({ page }) => {
    // Login and navigate
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    
    // Mock dedicated workspace
    await page.route('**/api/organizations/*/workspaces/dedicated-*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'ws-dedicated-123',
          name: 'Production Dedicated',
          plan: 'dedicated',
          features: {
            backup: true,
            nodes: 3,
          },
        }),
      });
    });
    
    await dashboardPage.openWorkspace('Production Dedicated');
    await captureScreenshot(page, 'backup', '01_dedicated_workspace');
    
    // Go to backup settings
    const settingsButton = page.getByTestId('workspace-settings-button');
    await settingsButton.click();
    await captureScreenshot(page, 'backup', '02_workspace_settings');
    
    const backupTab = page.getByRole('tab', { name: /backup/i });
    await backupTab.click();
    await captureScreenshot(page, 'backup', '03_backup_settings');
    
    // Show backup configuration
    await expect(page.getByTestId('backup-feature-badge')).toContainText('Dedicated Plan Feature');
    await captureScreenshot(page, 'backup', '04_backup_feature_available');
    
    console.log(`\n✅ Backup screenshots saved to: ${screenshotDir}/backup\n`);
  });
});

// Summary report
test.afterAll(async () => {
  console.log(`
📸 Screenshot Summary
====================
All E2E test screenshots have been saved to:
${screenshotDir}

Directory structure:
- /auth          - Login and authentication flows
- /dashboard     - Dashboard and navigation
- /organization  - Organization management
- /workspace     - Workspace operations
- /projects      - Project creation and management
- /applications  - Application deployment and scaling
- /deployments   - Deployment strategies
- /cicd          - CI/CD pipeline configuration
- /serverless    - Serverless function creation
- /backup        - Backup and restore features
- /success       - Final success states

Total screenshots captured: Check the directory for all images.
`);
});