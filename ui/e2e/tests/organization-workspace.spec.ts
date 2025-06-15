import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers, testOrganizations, testWorkspaces } from '../fixtures/mock-data';

test.describe('Organization and Workspace Management', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    // Setup mock API and pages
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    
    // Login before each test
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
  });

  test('displays existing organizations', async ({ page }) => {
    // Verify organizations are loaded
    await expect(dashboardPage.organizationSelector).toBeVisible();
    
    // Check first organization is selected by default
    await expect(dashboardPage.organizationSelector).toContainText(testOrganizations[0].name);
    
    // Open organization dropdown
    await dashboardPage.organizationSelector.click();
    
    // Verify all organizations are listed
    for (const org of testOrganizations) {
      await expect(page.getByRole('option', { name: org.name })).toBeVisible();
    }
  });

  test('create new organization', async ({ page }) => {
    const newOrgName = 'E2E New Organization';
    
    // Create organization
    await dashboardPage.createOrganization(newOrgName);
    
    // Verify organization was created
    await expect(dashboardPage.organizationSelector).toContainText(newOrgName);
    
    // Verify workspace grid is empty for new org
    await expect(page.getByText('No workspaces yet')).toBeVisible();
  });

  test('switch between organizations', async ({ page }) => {
    // Start with first organization
    await expect(dashboardPage.organizationSelector).toContainText(testOrganizations[0].name);
    
    // Switch to second organization
    await dashboardPage.selectOrganization(testOrganizations[1].name);
    
    // Verify switched
    await expect(dashboardPage.organizationSelector).toContainText(testOrganizations[1].name);
    
    // Verify workspace list updates (mocked to show same workspaces)
    await expect(dashboardPage.workspaceGrid).toBeVisible();
  });

  test('displays existing workspaces', async ({ page }) => {
    // Verify workspaces are displayed
    for (const workspace of testWorkspaces) {
      const card = await dashboardPage.getWorkspaceCard(workspace.name);
      await expect(card).toBeVisible();
      
      // Verify workspace details
      await expect(card.getByText(workspace.plan_id)).toBeVisible();
      await expect(card.getByTestId('workspace-status')).toContainText('active');
    }
  });

  test('create shared workspace', async ({ page }) => {
    const workspaceName = 'E2E Shared Test';
    
    // Create shared workspace
    await dashboardPage.createWorkspace(workspaceName, 'shared');
    
    // Verify workspace card appears
    const card = await dashboardPage.getWorkspaceCard(workspaceName);
    await expect(card).toBeVisible();
    await expect(card.getByText('shared')).toBeVisible();
    
    // Initially shows creating status
    await expect(card.getByTestId('workspace-status')).toContainText('creating');
    
    // Wait for it to become active (mocked to take 2 seconds)
    await dashboardPage.waitForWorkspaceActive(workspaceName);
    await expect(card.getByTestId('workspace-status')).toContainText('active');
  });

  test('create dedicated workspace with node selection', async ({ page }) => {
    const workspaceName = 'E2E Dedicated Test';
    
    // Start workspace creation
    await dashboardPage.createWorkspaceButton.click();
    
    const dialog = page.getByRole('dialog');
    await dialog.getByTestId('workspace-name-input').fill(workspaceName);
    await dialog.getByTestId('plan-dedicated').click();
    
    // Verify node pool selector appears for dedicated plan
    await expect(dialog.getByTestId('node-pool-select')).toBeVisible();
    
    // Select a node pool
    await dialog.getByTestId('node-pool-select').selectOption('dedicated-pool-1');
    
    // Verify resource preview
    await expect(dialog.getByText('16 CPU')).toBeVisible();
    await expect(dialog.getByText('32GB Memory')).toBeVisible();
    await expect(dialog.getByText('500GB Storage')).toBeVisible();
    
    // Create workspace
    await dialog.getByTestId('create-button').click();
    
    // Verify workspace created
    await dialog.waitFor({ state: 'hidden' });
    const card = await dashboardPage.getWorkspaceCard(workspaceName);
    await expect(card).toBeVisible();
    await expect(card.getByText('dedicated')).toBeVisible();
  });

  test('enter workspace', async ({ page }) => {
    const workspace = testWorkspaces[0];
    
    // Click enter on workspace
    await dashboardPage.openWorkspace(workspace.name);
    
    // Verify navigation
    await expect(page).toHaveURL(/\/workspaces\//);
    
    // Verify workspace context is loaded
    await expect(page.getByTestId('workspace-header')).toContainText(workspace.name);
    await expect(page.getByTestId('workspace-plan')).toContainText(workspace.plan_id);
  });

  test('shows workspace resource usage', async ({ page }) => {
    const workspace = testWorkspaces[0];
    const card = await dashboardPage.getWorkspaceCard(workspace.name);
    
    // Verify resource usage is displayed
    await expect(card.getByTestId('cpu-usage')).toBeVisible();
    await expect(card.getByTestId('memory-usage')).toBeVisible();
    await expect(card.getByTestId('storage-usage')).toBeVisible();
    
    // Hover for detailed view
    await card.hover();
    
    // Verify tooltip with detailed metrics
    await expect(page.getByTestId('resource-tooltip')).toBeVisible();
    await expect(page.getByTestId('resource-tooltip')).toContainText('CPU: 45.5%');
    await expect(page.getByTestId('resource-tooltip')).toContainText('Memory: 62.3%');
    await expect(page.getByTestId('resource-tooltip')).toContainText('Storage: 30.1%');
  });

  test('handles workspace creation errors', async ({ page }) => {
    // Override API to return error
    await page.route('**/api/organizations/*/workspaces', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'Workspace limit exceeded',
          }),
        });
      }
    });
    
    // Try to create workspace
    await dashboardPage.createWorkspace('Should Fail', 'shared');
    
    // Verify error message
    await expect(page.getByText('Workspace limit exceeded')).toBeVisible();
    
    // Dialog should remain open
    await expect(page.getByRole('dialog')).toBeVisible();
  });

  test('filters workspaces by status', async ({ page }) => {
    // Add filter controls
    const filterButton = page.getByTestId('workspace-filter-button');
    await filterButton.click();
    
    const filterMenu = page.getByTestId('workspace-filter-menu');
    
    // Filter by active workspaces
    await filterMenu.getByText('Active only').click();
    
    // Verify only active workspaces shown
    const cards = await dashboardPage.workspaceGrid.getByTestId(/workspace-card-/).all();
    for (const card of cards) {
      await expect(card.getByTestId('workspace-status')).toContainText('active');
    }
    
    // Clear filter
    await filterButton.click();
    await filterMenu.getByText('All workspaces').click();
  });
});