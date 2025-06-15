import { Page, Locator } from '@playwright/test';

export class MonitoringPage {
  readonly page: Page;
  readonly metricsOverview: Locator;
  readonly timeRangeSelector: Locator;
  readonly cpuUsageCard: Locator;
  readonly memoryUsageCard: Locator;
  readonly networkIOCard: Locator;
  readonly diskUsageCard: Locator;
  readonly cpuChart: Locator;
  readonly memoryChart: Locator;
  readonly alertsList: Locator;
  readonly grafanaFrame: Locator;

  constructor(page: Page) {
    this.page = page;
    this.metricsOverview = page.getByTestId('metrics-overview');
    this.timeRangeSelector = page.getByTestId('time-range-selector');
    this.cpuUsageCard = page.getByTestId('cpu-usage-card');
    this.memoryUsageCard = page.getByTestId('memory-usage-card');
    this.networkIOCard = page.getByTestId('network-io-card');
    this.diskUsageCard = page.getByTestId('disk-usage-card');
    this.cpuChart = page.getByTestId('cpu-chart');
    this.memoryChart = page.getByTestId('memory-chart');
    this.alertsList = page.getByTestId('alerts-list');
    this.grafanaFrame = page.frameLocator('iframe[title="Grafana Dashboard"]');
  }

  async selectTimeRange(range: string) {
    await this.timeRangeSelector.click();
    await this.page.getByTestId(`time-range-${range}`).click();
    // Wait for metrics to update
    await this.page.waitForTimeout(500);
  }

  async getMetricValue(metricCard: Locator): Promise<string> {
    const valueElement = metricCard.getByTestId('metric-value');
    return await valueElement.textContent() || '';
  }

  async getCPUUsage(): Promise<string> {
    return this.getMetricValue(this.cpuUsageCard);
  }

  async getMemoryUsage(): Promise<string> {
    return this.getMetricValue(this.memoryUsageCard);
  }

  async configureAlert(config: {
    name: string;
    metric: string;
    condition: string;
    threshold: string;
    duration: string;
  }) {
    const alertsButton = this.page.getByTestId('configure-alerts-button');
    await alertsButton.click();
    
    const dialog = this.page.getByRole('dialog');
    await dialog.getByTestId('add-alert-button').click();
    
    await dialog.getByTestId('alert-name-input').fill(config.name);
    await dialog.getByTestId('metric-select').selectOption(config.metric);
    await dialog.getByTestId('condition-select').selectOption(config.condition);
    await dialog.getByTestId('threshold-input').fill(config.threshold);
    await dialog.getByTestId('duration-input').fill(config.duration);
    
    await dialog.getByTestId('save-alert-button').click();
  }

  async executePrometheusQuery(query: string) {
    const queryInput = this.page.getByTestId('prometheus-query-input');
    await queryInput.fill(query);
    await this.page.getByTestId('execute-query-button').click();
    
    // Wait for results
    await this.page.waitForSelector('[data-testid="query-results"]');
  }

  async openGrafanaDashboard(dashboardName?: string) {
    const grafanaTab = this.page.getByRole('tab', { name: /grafana/i });
    await grafanaTab.click();
    
    if (dashboardName) {
      await this.page.getByTestId(`dashboard-${dashboardName}`).click();
    }
  }

  async exportMetrics(options: {
    metrics: string[];
    format: 'csv' | 'json';
    timeRange: string;
  }) {
    const exportButton = this.page.getByTestId('export-metrics-button');
    await exportButton.click();
    
    const dialog = this.page.getByRole('dialog');
    
    // Select metrics
    for (const metric of options.metrics) {
      await dialog.getByTestId(`export-${metric}-checkbox`).check();
    }
    
    // Select format
    await dialog.getByTestId('export-format-select').selectOption(options.format);
    
    // Export
    return dialog.getByTestId('export-button').click();
  }

  async getAlertCount(): Promise<number> {
    const alerts = this.alertsList.locator('[data-testid^="alert-item-"]');
    return alerts.count();
  }

  async viewClusterMetrics() {
    await this.page.getByTestId('cluster-monitoring-button').click();
    await this.page.waitForSelector('[data-testid="cluster-overview"]');
  }

  async createCustomDashboard(name: string, description: string) {
    const customDashboardsTab = this.page.getByRole('tab', { name: /custom dashboards/i });
    await customDashboardsTab.click();
    
    const createButton = this.page.getByTestId('create-dashboard-button');
    await createButton.click();
    
    const dialog = this.page.getByRole('dialog');
    await dialog.getByTestId('dashboard-name-input').fill(name);
    await dialog.getByTestId('dashboard-description-input').fill(description);
    
    return dialog;
  }

  async addDashboardWidget(dialog: Locator, widgetConfig: {
    type: string;
    title: string;
    metric: string;
    aggregation?: string;
  }) {
    await dialog.getByTestId('add-widget-button').click();
    
    const widgetIndex = await dialog.locator('[data-testid^="widget-config-"]').count() - 1;
    const widgetConfigSection = dialog.getByTestId(`widget-config-${widgetIndex}`);
    
    await widgetConfigSection.getByTestId('widget-type-select').selectOption(widgetConfig.type);
    await widgetConfigSection.getByTestId('widget-title-input').fill(widgetConfig.title);
    await widgetConfigSection.getByTestId('metric-select').selectOption(widgetConfig.metric);
    
    if (widgetConfig.aggregation) {
      await widgetConfigSection.getByTestId('aggregation-select').selectOption(widgetConfig.aggregation);
    }
  }
}