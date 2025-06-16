import { Page, Locator } from '@playwright/test';

export class WorkspacePage {
  readonly page: Page;
  readonly workspaceHeader: Locator;
  readonly workspacePlan: Locator;
  readonly projectsGrid: Locator;
  readonly createProjectButton: Locator;
  readonly resourceUsageCard: Locator;
  readonly settingsTab: Locator;
  readonly backupsTab: Locator;
  readonly membersTab: Locator;
  readonly breadcrumb: Locator;

  constructor(page: Page) {
    this.page = page;
    this.workspaceHeader = page.getByTestId('workspace-header');
    this.workspacePlan = page.getByTestId('workspace-plan');
    this.projectsGrid = page.getByTestId('projects-grid');
    this.createProjectButton = page.getByTestId('create-project-button');
    this.resourceUsageCard = page.getByTestId('resource-usage-card');
    this.settingsTab = page.getByRole('tab', { name: /settings/i });
    this.backupsTab = page.getByRole('tab', { name: /backups/i });
    this.membersTab = page.getByRole('tab', { name: /members/i });
    this.breadcrumb = page.getByTestId('breadcrumb');
  }

  async createProject(name: string, quotas?: { cpu: string; memory: string; storage: string }) {
    await this.createProjectButton.click();
    
    const dialog = this.page.getByRole('dialog');
    await dialog.getByTestId('project-name-input').fill(name);
    
    if (quotas) {
      // Expand advanced settings
      const advancedButton = dialog.getByText('Advanced Settings');
      if (await advancedButton.isVisible()) {
        await advancedButton.click();
      }
      
      await dialog.getByTestId('cpu-quota-input').fill(quotas.cpu);
      await dialog.getByTestId('memory-quota-input').fill(quotas.memory);
      await dialog.getByTestId('storage-quota-input').fill(quotas.storage);
    }
    
    await dialog.getByTestId('create-button').click();
    await dialog.waitFor({ state: 'hidden' });
  }

  async getProjectCard(name: string) {
    return this.projectsGrid.getByTestId(`project-card-${name}`);
  }

  async openProject(name: string) {
    const card = await this.getProjectCard(name);
    await card.getByTestId('enter-project-button').click();
    
    // Wait for navigation
    await this.page.waitForURL('**/projects/**');
  }

  async getResourceUsage() {
    const cpuUsage = await this.resourceUsageCard.getByTestId('cpu-usage').textContent();
    const memoryUsage = await this.resourceUsageCard.getByTestId('memory-usage').textContent();
    const storageUsage = await this.resourceUsageCard.getByTestId('storage-usage').textContent();
    
    return {
      cpu: cpuUsage,
      memory: memoryUsage,
      storage: storageUsage,
    };
  }

  async navigateToSettings() {
    await this.settingsTab.click();
    await this.page.waitForSelector('[data-testid="workspace-settings"]');
  }

  async navigateToBackups() {
    await this.backupsTab.click();
    await this.page.waitForSelector('[data-testid="backup-dashboard"]');
  }

  async navigateToMembers() {
    await this.membersTab.click();
    await this.page.waitForSelector('[data-testid="members-list"]');
  }

  async inviteMember(email: string, role: 'admin' | 'developer' | 'viewer') {
    await this.navigateToMembers();
    
    const inviteButton = this.page.getByTestId('invite-member-button');
    await inviteButton.click();
    
    const dialog = this.page.getByRole('dialog');
    await dialog.getByTestId('email-input').fill(email);
    await dialog.getByTestId('role-select').selectOption(role);
    await dialog.getByTestId('send-invite-button').click();
    
    await dialog.waitFor({ state: 'hidden' });
  }

  async deleteProject(name: string) {
    const card = await this.getProjectCard(name);
    const menuButton = card.getByTestId('project-menu-button');
    await menuButton.click();
    
    const deleteOption = this.page.getByRole('menuitem', { name: /delete/i });
    await deleteOption.click();
    
    // Confirm deletion
    const confirmDialog = this.page.getByRole('dialog');
    const confirmInput = confirmDialog.getByTestId('confirm-name-input');
    await confirmInput.fill(name);
    
    const deleteButton = confirmDialog.getByTestId('confirm-delete-button');
    await deleteButton.click();
    
    await confirmDialog.waitFor({ state: 'hidden' });
  }

  async waitForProjectStatus(name: string, status: string, timeout: number = 30000) {
    await this.page.waitForFunction(
      ({ projectName, expectedStatus }) => {
        const card = document.querySelector(`[data-testid="project-card-${projectName}"]`);
        const statusElement = card?.querySelector('[data-testid="project-status"]');
        return statusElement?.textContent?.toLowerCase() === expectedStatus.toLowerCase();
      },
      { projectName: name, expectedStatus: status },
      { timeout }
    );
  }
}