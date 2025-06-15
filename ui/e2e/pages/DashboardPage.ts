import { Page, Locator } from '@playwright/test';

export class DashboardPage {
  readonly page: Page;
  readonly organizationSelector: Locator;
  readonly createOrgButton: Locator;
  readonly workspaceGrid: Locator;
  readonly createWorkspaceButton: Locator;
  readonly userMenu: Locator;
  readonly logoutButton: Locator;

  constructor(page: Page) {
    this.page = page;
    this.organizationSelector = page.getByTestId('organization-selector');
    this.createOrgButton = page.getByTestId('create-organization-button');
    this.workspaceGrid = page.getByTestId('workspace-grid');
    this.createWorkspaceButton = page.getByTestId('create-workspace-button');
    this.userMenu = page.getByTestId('user-menu');
    this.logoutButton = page.getByTestId('logout-button');
  }

  async selectOrganization(orgName: string) {
    await this.organizationSelector.click();
    await this.page.getByRole('option', { name: orgName }).click();
  }

  async createOrganization(name: string) {
    await this.createOrgButton.click();
    
    const dialog = this.page.getByRole('dialog');
    await dialog.getByTestId('org-name-input').fill(name);
    await dialog.getByTestId('create-button').click();
    
    // Wait for dialog to close
    await dialog.waitFor({ state: 'hidden' });
    
    // Wait for new org to appear
    await this.page.waitForSelector(`text=${name}`);
  }

  async getWorkspaceCard(name: string) {
    return this.workspaceGrid.getByTestId(`workspace-card-${name}`);
  }

  async openWorkspace(name: string) {
    const card = await this.getWorkspaceCard(name);
    await card.getByTestId('enter-workspace-button').click();
    
    // Wait for navigation
    await this.page.waitForURL('**/workspaces/**');
  }

  async createWorkspace(name: string, plan: 'shared' | 'dedicated', nodePool?: string) {
    await this.createWorkspaceButton.click();
    
    const dialog = this.page.getByRole('dialog');
    await dialog.getByTestId('workspace-name-input').fill(name);
    await dialog.getByTestId(`plan-${plan}`).click();
    
    if (plan === 'dedicated' && nodePool) {
      await dialog.getByTestId('node-pool-select').selectOption(nodePool);
    }
    
    await dialog.getByTestId('create-button').click();
    
    // Wait for workspace creation
    await dialog.waitFor({ state: 'hidden' });
    await this.page.waitForSelector(`[data-testid="workspace-card-${name}"]`);
  }

  async logout() {
    await this.userMenu.click();
    await this.logoutButton.click();
    
    // Wait for redirect to login
    await this.page.waitForURL('**/login', { timeout: 5000 });
  }

  async getWorkspaceStatus(name: string) {
    const card = await this.getWorkspaceCard(name);
    const status = await card.getByTestId('workspace-status').textContent();
    return status?.toLowerCase();
  }

  async waitForWorkspaceActive(name: string, timeout: number = 30000) {
    await this.page.waitForFunction(
      async (workspaceName) => {
        const card = document.querySelector(`[data-testid="workspace-card-${workspaceName}"]`);
        const status = card?.querySelector('[data-testid="workspace-status"]')?.textContent;
        return status?.toLowerCase() === 'active';
      },
      name,
      { timeout }
    );
  }
}