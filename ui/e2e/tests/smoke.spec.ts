import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers, testWorkspaces } from '../fixtures/mock-data';
import { SMOKE_TAG, CRITICAL_TAG } from '../utils/test-tags';

test.describe('Smoke Tests - Critical User Paths', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let workspacePage: WorkspacePage;
  let projectPage: ProjectPage;

  test.beforeEach(async ({ page }) => {
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    workspacePage = new WorkspacePage(page);
    projectPage = new ProjectPage(page);
  });

  test(`user can login successfully ${SMOKE_TAG} ${CRITICAL_TAG}`, async ({ page }) => {
    await loginPage.goto();
    await expect(page).toHaveTitle(/Hexabase AI/);
    
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    await expect(dashboardPage.welcomeMessage).toBeVisible();
    await expect(dashboardPage.welcomeMessage).toContainText('Welcome');
  });

  test(`user can navigate to workspace ${SMOKE_TAG}`, async ({ page }) => {
    // Quick login
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    
    // Verify organizations are visible
    const orgCards = await dashboardPage.getOrganizationCards();
    expect(orgCards.length).toBeGreaterThan(0);
    
    // Enter workspace
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    await expect(workspacePage.workspaceHeader).toContainText(testWorkspaces[0].name);
    
    // Verify workspace is functional
    await expect(workspacePage.createProjectButton).toBeVisible();
  });

  test(`user can create and access project ${SMOKE_TAG}`, async ({ page }) => {
    // Quick setup
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    
    // Create project
    const projectName = `Smoke Test Project ${Date.now()}`;
    await workspacePage.createProject(projectName);
    
    // Verify project created
    const projectCard = await workspacePage.getProjectCard(projectName);
    await expect(projectCard).toBeVisible();
    
    // Enter project
    await workspacePage.openProject(projectName);
    await expect(projectPage.projectHeader).toContainText(projectName);
    
    // Verify project tabs are accessible
    await expect(projectPage.applicationsTab).toBeVisible();
    await expect(projectPage.functionsTab).toBeVisible();
    await expect(projectPage.cronJobsTab).toBeVisible();
  });

  test(`user can deploy application ${SMOKE_TAG} ${CRITICAL_TAG}`, async ({ page }) => {
    // Quick setup
    await loginPage.goto();
    await loginPage.login(testUsers.developer.email, testUsers.developer.password);
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    await workspacePage.createProject('Smoke Deploy Test');
    await workspacePage.openProject('Smoke Deploy Test');
    
    // Deploy simple application
    const appConfig = {
      name: 'smoke-test-app',
      type: 'stateless' as const,
      image: 'nginx:alpine',
      replicas: 1,
      port: 80,
    };
    
    await projectPage.deployApplication(appConfig);
    
    // Verify deployment started
    await expect(page.getByText('Deployment started')).toBeVisible();
    
    // Wait for running status (with shorter timeout for smoke test)
    await projectPage.waitForApplicationStatus(appConfig.name, 'running', 15000);
    
    // Verify application card shows correct status
    const appCard = await projectPage.getApplicationCard(appConfig.name);
    await expect(appCard.getByTestId('app-status')).toContainText('running');
  });

  test(`health check endpoints respond ${SMOKE_TAG}`, async ({ page }) => {
    // Test API health endpoint
    const apiHealth = await page.request.get('/api/health');
    expect(apiHealth.ok()).toBeTruthy();
    
    const healthData = await apiHealth.json();
    expect(healthData).toHaveProperty('status', 'healthy');
    
    // Test readiness endpoint
    const readiness = await page.request.get('/api/ready');
    expect(readiness.ok()).toBeTruthy();
  });

  test(`main navigation elements are accessible ${SMOKE_TAG}`, async ({ page }) => {
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    
    // Check main navigation
    await expect(page.getByTestId('nav-dashboard')).toBeVisible();
    await expect(page.getByTestId('nav-organizations')).toBeVisible();
    await expect(page.getByTestId('nav-workspaces')).toBeVisible();
    await expect(page.getByTestId('nav-settings')).toBeVisible();
    
    // Check user menu
    await page.getByTestId('user-menu-button').click();
    await expect(page.getByTestId('user-menu-profile')).toBeVisible();
    await expect(page.getByTestId('user-menu-logout')).toBeVisible();
  });

  test(`error pages display correctly ${SMOKE_TAG}`, async ({ page }) => {
    // Test 404 page
    await page.goto('/non-existent-page');
    await expect(page.getByText('404')).toBeVisible();
    await expect(page.getByText(/not found/i)).toBeVisible();
    
    // Test unauthorized access
    await page.goto('/dashboard');
    // Should redirect to login
    await expect(page).toHaveURL(/\/login/);
  });

  test(`user can logout successfully ${SMOKE_TAG}`, async ({ page }) => {
    // Login first
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    // Logout
    await loginPage.logout();
    
    // Verify redirected to login
    await expect(page).toHaveURL(/\/login/);
    await expect(loginPage.emailInput).toBeVisible();
    
    // Verify can't access protected routes
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });
});