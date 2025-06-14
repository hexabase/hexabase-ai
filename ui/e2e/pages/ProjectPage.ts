import { Page, Locator } from '@playwright/test';

export class ProjectPage {
  readonly page: Page;
  readonly projectHeader: Locator;
  readonly applicationsTab: Locator;
  readonly functionsTab: Locator;
  readonly cronJobsTab: Locator;
  readonly settingsTab: Locator;
  readonly monitoringTab: Locator;
  readonly deployAppButton: Locator;
  readonly applicationsGrid: Locator;
  readonly resourceQuotaCard: Locator;
  readonly activityFeed: Locator;

  constructor(page: Page) {
    this.page = page;
    this.projectHeader = page.getByTestId('project-header');
    this.applicationsTab = page.getByRole('tab', { name: /applications/i });
    this.functionsTab = page.getByRole('tab', { name: /functions/i });
    this.cronJobsTab = page.getByRole('tab', { name: /cronjobs/i });
    this.settingsTab = page.getByRole('tab', { name: /settings/i });
    this.monitoringTab = page.getByRole('tab', { name: /monitoring/i });
    this.deployAppButton = page.getByTestId('deploy-application-button');
    this.applicationsGrid = page.getByTestId('applications-grid');
    this.resourceQuotaCard = page.getByTestId('resource-quota-card');
    this.activityFeed = page.getByTestId('activity-feed');
  }

  async deployApplication(config: {
    name: string;
    type: 'stateless' | 'stateful' | 'cronjob';
    image: string;
    replicas?: number;
    port?: number;
    storage?: string;
    schedule?: string;
    env?: Record<string, string>;
  }) {
    await this.deployAppButton.click();
    
    const dialog = this.page.getByRole('dialog');
    
    // Basic configuration
    await dialog.getByTestId('app-name-input').fill(config.name);
    await dialog.getByTestId('app-type-select').selectOption(config.type);
    await dialog.getByTestId('container-image-input').fill(config.image);
    
    // Type-specific configuration
    if (config.type === 'stateless' && config.replicas) {
      await dialog.getByTestId('replicas-input').fill(config.replicas.toString());
    }
    
    if (config.type === 'stateful' && config.storage) {
      await dialog.getByTestId('storage-size-input').fill(config.storage);
    }
    
    if (config.type === 'cronjob' && config.schedule) {
      await dialog.getByTestId('cron-schedule-input').fill(config.schedule);
    }
    
    if (config.port) {
      await dialog.getByTestId('container-port-input').fill(config.port.toString());
    }
    
    // Environment variables
    if (config.env) {
      const addEnvButton = dialog.getByTestId('add-env-button');
      
      for (const [key, value] of Object.entries(config.env)) {
        await addEnvButton.click();
        
        const envRows = dialog.locator('[data-testid^="env-row-"]');
        const lastRow = envRows.last();
        
        await lastRow.getByTestId('env-key-input').fill(key);
        await lastRow.getByTestId('env-value-input').fill(value);
      }
    }
    
    // Deploy
    await dialog.getByTestId('deploy-button').click();
    await dialog.waitFor({ state: 'hidden' });
  }

  async getApplicationCard(name: string) {
    return this.applicationsGrid.getByTestId(`app-card-${name}`);
  }

  async getApplicationStatus(name: string) {
    const card = await this.getApplicationCard(name);
    return card.getByTestId('app-status').textContent();
  }

  async waitForApplicationStatus(name: string, status: string, timeout: number = 60000) {
    await this.page.waitForFunction(
      ({ appName, expectedStatus }) => {
        const card = document.querySelector(`[data-testid="app-card-${appName}"]`);
        const statusElement = card?.querySelector('[data-testid="app-status"]');
        return statusElement?.textContent?.toLowerCase() === expectedStatus.toLowerCase();
      },
      { appName: name, expectedStatus: status },
      { timeout }
    );
  }

  async scaleApplication(name: string, replicas: number) {
    const card = await this.getApplicationCard(name);
    const menuButton = card.getByTestId('app-menu-button');
    await menuButton.click();
    
    const scaleOption = this.page.getByRole('menuitem', { name: /scale/i });
    await scaleOption.click();
    
    const dialog = this.page.getByRole('dialog');
    await dialog.getByTestId('replicas-input').fill(replicas.toString());
    await dialog.getByTestId('scale-button').click();
    
    await dialog.waitFor({ state: 'hidden' });
  }

  async deleteApplication(name: string) {
    const card = await this.getApplicationCard(name);
    const menuButton = card.getByTestId('app-menu-button');
    await menuButton.click();
    
    const deleteOption = this.page.getByRole('menuitem', { name: /delete/i });
    await deleteOption.click();
    
    // Confirm deletion
    const confirmDialog = this.page.getByRole('dialog');
    await confirmDialog.getByTestId('confirm-delete-button').click();
    
    await confirmDialog.waitFor({ state: 'hidden' });
  }

  async viewApplicationLogs(name: string) {
    const card = await this.getApplicationCard(name);
    const logsButton = card.getByTestId('view-logs-button');
    await logsButton.click();
    
    // Wait for logs modal
    await this.page.waitForSelector('[data-testid="logs-viewer"]');
  }

  async viewApplicationMetrics(name: string) {
    const card = await this.getApplicationCard(name);
    const metricsButton = card.getByTestId('view-metrics-button');
    await metricsButton.click();
    
    // Wait for metrics view
    await this.page.waitForSelector('[data-testid="metrics-dashboard"]');
  }

  async navigateToFunctions() {
    await this.functionsTab.click();
    await this.page.waitForSelector('[data-testid="functions-list"]');
  }

  async navigateToCronJobs() {
    await this.cronJobsTab.click();
    await this.page.waitForSelector('[data-testid="cronjobs-list"]');
  }

  async navigateToMonitoring() {
    await this.monitoringTab.click();
    await this.page.waitForSelector('[data-testid="monitoring-dashboard"]');
  }

  async getResourceQuota() {
    const cpuLimit = await this.resourceQuotaCard.getByTestId('cpu-limit').textContent();
    const memoryLimit = await this.resourceQuotaCard.getByTestId('memory-limit').textContent();
    const storageLimit = await this.resourceQuotaCard.getByTestId('storage-limit').textContent();
    
    const cpuUsed = await this.resourceQuotaCard.getByTestId('cpu-used').textContent();
    const memoryUsed = await this.resourceQuotaCard.getByTestId('memory-used').textContent();
    const storageUsed = await this.resourceQuotaCard.getByTestId('storage-used').textContent();
    
    return {
      limits: { cpu: cpuLimit, memory: memoryLimit, storage: storageLimit },
      used: { cpu: cpuUsed, memory: memoryUsed, storage: storageUsed },
    };
  }

  async updateResourceQuota(quotas: { cpu?: string; memory?: string; storage?: string }) {
    await this.settingsTab.click();
    
    const editQuotaButton = this.page.getByTestId('edit-quota-button');
    await editQuotaButton.click();
    
    const dialog = this.page.getByRole('dialog');
    
    if (quotas.cpu) {
      await dialog.getByTestId('cpu-quota-input').fill(quotas.cpu);
    }
    if (quotas.memory) {
      await dialog.getByTestId('memory-quota-input').fill(quotas.memory);
    }
    if (quotas.storage) {
      await dialog.getByTestId('storage-quota-input').fill(quotas.storage);
    }
    
    await dialog.getByTestId('save-button').click();
    await dialog.waitFor({ state: 'hidden' });
  }

  async getRecentActivity() {
    const activities = await this.activityFeed.locator('[data-testid^="activity-item-"]').all();
    const activityData = [];
    
    for (const activity of activities) {
      const type = await activity.getByTestId('activity-type').textContent();
      const message = await activity.getByTestId('activity-message').textContent();
      const timestamp = await activity.getByTestId('activity-timestamp').textContent();
      
      activityData.push({ type, message, timestamp });
    }
    
    return activityData;
  }
}