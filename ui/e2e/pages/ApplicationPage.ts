import { Page, Locator } from '@playwright/test';

export class ApplicationPage {
  readonly page: Page;
  readonly appHeader: Locator;
  readonly statusBadge: Locator;
  readonly overviewTab: Locator;
  readonly logsTab: Locator;
  readonly metricsTab: Locator;
  readonly configTab: Locator;
  readonly eventsTab: Locator;
  readonly podsTable: Locator;
  readonly restartButton: Locator;
  readonly scaleButton: Locator;
  readonly updateButton: Locator;
  readonly deleteButton: Locator;
  readonly healthStatus: Locator;

  constructor(page: Page) {
    this.page = page;
    this.appHeader = page.getByTestId('app-header');
    this.statusBadge = page.getByTestId('app-status-badge');
    this.overviewTab = page.getByRole('tab', { name: /overview/i });
    this.logsTab = page.getByRole('tab', { name: /logs/i });
    this.metricsTab = page.getByRole('tab', { name: /metrics/i });
    this.configTab = page.getByRole('tab', { name: /configuration/i });
    this.eventsTab = page.getByRole('tab', { name: /events/i });
    this.podsTable = page.getByTestId('pods-table');
    this.restartButton = page.getByTestId('restart-app-button');
    this.scaleButton = page.getByTestId('scale-app-button');
    this.updateButton = page.getByTestId('update-app-button');
    this.deleteButton = page.getByTestId('delete-app-button');
    this.healthStatus = page.getByTestId('health-status');
  }

  async getStatus() {
    return this.statusBadge.textContent();
  }

  async waitForStatus(status: string, timeout: number = 60000) {
    await this.page.waitForFunction(
      (expectedStatus) => {
        const badge = document.querySelector('[data-testid="app-status-badge"]');
        return badge?.textContent?.toLowerCase() === expectedStatus.toLowerCase();
      },
      status,
      { timeout }
    );
  }

  async restartApplication() {
    await this.restartButton.click();
    
    const confirmDialog = this.page.getByRole('dialog');
    await confirmDialog.getByTestId('confirm-restart-button').click();
    await confirmDialog.waitFor({ state: 'hidden' });
  }

  async scaleApplication(replicas: number) {
    await this.scaleButton.click();
    
    const dialog = this.page.getByRole('dialog');
    const replicasInput = dialog.getByTestId('replicas-input');
    await replicasInput.clear();
    await replicasInput.fill(replicas.toString());
    
    await dialog.getByTestId('apply-scale-button').click();
    await dialog.waitFor({ state: 'hidden' });
  }

  async updateApplication(config: {
    image?: string;
    env?: Record<string, string>;
    resources?: {
      cpu?: string;
      memory?: string;
    };
  }) {
    await this.updateButton.click();
    
    const dialog = this.page.getByRole('dialog');
    
    if (config.image) {
      const imageInput = dialog.getByTestId('container-image-input');
      await imageInput.clear();
      await imageInput.fill(config.image);
    }
    
    if (config.env) {
      // Navigate to environment tab
      await dialog.getByRole('tab', { name: /environment/i }).click();
      
      for (const [key, value] of Object.entries(config.env)) {
        const envRow = dialog.locator(`[data-testid="env-row-${key}"]`);
        if (await envRow.isVisible()) {
          // Update existing
          await envRow.getByTestId('env-value-input').fill(value);
        } else {
          // Add new
          await dialog.getByTestId('add-env-button').click();
          const newRow = dialog.locator('[data-testid^="env-row-"]').last();
          await newRow.getByTestId('env-key-input').fill(key);
          await newRow.getByTestId('env-value-input').fill(value);
        }
      }
    }
    
    if (config.resources) {
      // Navigate to resources tab
      await dialog.getByRole('tab', { name: /resources/i }).click();
      
      if (config.resources.cpu) {
        const cpuInput = dialog.getByTestId('cpu-request-input');
        await cpuInput.clear();
        await cpuInput.fill(config.resources.cpu);
      }
      
      if (config.resources.memory) {
        const memoryInput = dialog.getByTestId('memory-request-input');
        await memoryInput.clear();
        await memoryInput.fill(config.resources.memory);
      }
    }
    
    await dialog.getByTestId('update-button').click();
    await dialog.waitFor({ state: 'hidden' });
  }

  async deleteApplication() {
    await this.deleteButton.click();
    
    const confirmDialog = this.page.getByRole('dialog');
    const confirmCheckbox = confirmDialog.getByTestId('confirm-delete-checkbox');
    await confirmCheckbox.check();
    
    await confirmDialog.getByTestId('confirm-delete-button').click();
    await confirmDialog.waitFor({ state: 'hidden' });
  }

  async viewLogs(options?: { container?: string; since?: string; tail?: number }) {
    await this.logsTab.click();
    
    if (options?.container) {
      const containerSelect = this.page.getByTestId('container-select');
      await containerSelect.selectOption(options.container);
    }
    
    if (options?.since) {
      const sinceSelect = this.page.getByTestId('logs-since-select');
      await sinceSelect.selectOption(options.since);
    }
    
    if (options?.tail) {
      const tailInput = this.page.getByTestId('logs-tail-input');
      await tailInput.fill(options.tail.toString());
    }
    
    // Wait for logs to load
    await this.page.waitForSelector('[data-testid="logs-content"]');
  }

  async getLogs() {
    const logsContent = this.page.getByTestId('logs-content');
    return logsContent.textContent();
  }

  async viewMetrics(timeRange?: '1h' | '6h' | '24h' | '7d') {
    await this.metricsTab.click();
    
    if (timeRange) {
      const timeRangeSelect = this.page.getByTestId('time-range-select');
      await timeRangeSelect.selectOption(timeRange);
    }
    
    // Wait for metrics to load
    await this.page.waitForSelector('[data-testid="metrics-charts"]');
  }

  async getMetrics() {
    const cpuUsage = await this.page.getByTestId('cpu-usage-value').textContent();
    const memoryUsage = await this.page.getByTestId('memory-usage-value').textContent();
    const requestRate = await this.page.getByTestId('request-rate-value').textContent();
    const errorRate = await this.page.getByTestId('error-rate-value').textContent();
    
    return {
      cpu: cpuUsage,
      memory: memoryUsage,
      requestRate,
      errorRate,
    };
  }

  async getPods() {
    const pods = await this.podsTable.locator('tbody tr').all();
    const podData = [];
    
    for (const pod of pods) {
      const name = await pod.getByTestId('pod-name').textContent();
      const status = await pod.getByTestId('pod-status').textContent();
      const restarts = await pod.getByTestId('pod-restarts').textContent();
      const age = await pod.getByTestId('pod-age').textContent();
      
      podData.push({ name, status, restarts, age });
    }
    
    return podData;
  }

  async restartPod(podName: string) {
    const podRow = this.podsTable.locator(`tr:has-text("${podName}")`);
    const restartButton = podRow.getByTestId('restart-pod-button');
    await restartButton.click();
    
    const confirmDialog = this.page.getByRole('dialog');
    await confirmDialog.getByTestId('confirm-restart-button').click();
    await confirmDialog.waitFor({ state: 'hidden' });
  }

  async viewPodLogs(podName: string) {
    const podRow = this.podsTable.locator(`tr:has-text("${podName}")`);
    const logsButton = podRow.getByTestId('view-pod-logs-button');
    await logsButton.click();
    
    // Wait for logs viewer
    await this.page.waitForSelector('[data-testid="pod-logs-viewer"]');
  }

  async getEvents() {
    await this.eventsTab.click();
    
    const events = await this.page.locator('[data-testid^="event-item-"]').all();
    const eventData = [];
    
    for (const event of events) {
      const type = await event.getByTestId('event-type').textContent();
      const reason = await event.getByTestId('event-reason').textContent();
      const message = await event.getByTestId('event-message').textContent();
      const timestamp = await event.getByTestId('event-timestamp').textContent();
      
      eventData.push({ type, reason, message, timestamp });
    }
    
    return eventData;
  }

  async getHealthStatus() {
    const status = await this.healthStatus.getByTestId('health-status-text').textContent();
    const checks = await this.healthStatus.locator('[data-testid^="health-check-"]').all();
    const healthChecks = [];
    
    for (const check of checks) {
      const name = await check.getByTestId('check-name').textContent();
      const status = await check.getByTestId('check-status').textContent();
      healthChecks.push({ name, status });
    }
    
    return { status, checks: healthChecks };
  }

  async exportConfiguration() {
    await this.configTab.click();
    
    const exportButton = this.page.getByTestId('export-config-button');
    await exportButton.click();
    
    // Handle download
    const download = await this.page.waitForEvent('download');
    return download.suggestedFilename();
  }
}