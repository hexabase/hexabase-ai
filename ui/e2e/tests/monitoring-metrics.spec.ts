import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { ApplicationPage } from '../pages/ApplicationPage';
import { MonitoringPage } from '../pages/MonitoringPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers, testWorkspaces } from '../fixtures/mock-data';
import { generateAppName, expectNotification } from '../utils/test-helpers';
import { SMOKE_TAG } from '../utils/test-tags';

test.describe('Monitoring and Metrics', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let workspacePage: WorkspacePage;
  let projectPage: ProjectPage;
  let applicationPage: ApplicationPage;
  let monitoringPage: MonitoringPage;

  test.beforeEach(async ({ page }) => {
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    workspacePage = new WorkspacePage(page);
    projectPage = new ProjectPage(page);
    applicationPage = new ApplicationPage(page);
    monitoringPage = new MonitoringPage(page);
    
    // Login and navigate to project
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    await workspacePage.createProject('Monitoring Test Project');
    await workspacePage.openProject('Monitoring Test Project');
  });

  test(`view real-time metrics dashboard ${SMOKE_TAG}`, async ({ page }) => {
    // Deploy test application
    const appName = generateAppName();
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 3,
      port: 80,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Navigate to monitoring
    await projectPage.monitoringTab.click();
    
    // Verify monitoring dashboard components
    await expect(monitoringPage.metricsOverview).toBeVisible();
    await expect(monitoringPage.timeRangeSelector).toBeVisible();
    
    // Check key metrics cards
    await expect(monitoringPage.cpuUsageCard).toBeVisible();
    await expect(monitoringPage.memoryUsageCard).toBeVisible();
    await expect(monitoringPage.networkIOCard).toBeVisible();
    await expect(monitoringPage.diskUsageCard).toBeVisible();
    
    // Verify metric values are displayed
    const cpuValue = await monitoringPage.cpuUsageCard.getByTestId('metric-value').textContent();
    expect(cpuValue).toMatch(/\d+(\.\d+)?%/);
    
    const memoryValue = await monitoringPage.memoryUsageCard.getByTestId('metric-value').textContent();
    expect(memoryValue).toMatch(/\d+(\.\d+)?\s*(Mi|Gi)/);
  });

  test('change metrics time range', async ({ page }) => {
    await projectPage.monitoringTab.click();
    
    // Test different time ranges
    const timeRanges = ['1h', '6h', '24h', '7d', '30d'];
    
    for (const range of timeRanges) {
      await monitoringPage.selectTimeRange(range);
      
      // Verify time range is selected
      await expect(monitoringPage.timeRangeSelector).toContainText(range);
      
      // Wait for charts to update
      await page.waitForTimeout(500);
      
      // Verify charts are updated
      await expect(page.getByTestId('chart-loading')).not.toBeVisible();
      await expect(monitoringPage.cpuChart).toBeVisible();
      
      // Verify time range in chart axis
      const chartTimeLabel = page.getByTestId('chart-time-axis-label');
      if (range === '1h') {
        await expect(chartTimeLabel).toContainText(/minutes|mins/i);
      } else if (range.includes('d')) {
        await expect(chartTimeLabel).toContainText(/days/i);
      }
    }
  });

  test('view application-specific metrics', async ({ page }) => {
    // Deploy application
    const appName = generateAppName();
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    // Go to metrics tab
    await applicationPage.metricsTab.click();
    
    // Verify application metrics
    const metrics = await applicationPage.getMetrics();
    expect(metrics.cpu).toBeTruthy();
    expect(metrics.memory).toBeTruthy();
    expect(metrics.requestRate).toBeTruthy();
    expect(metrics.errorRate).toBeTruthy();
    
    // Check request latency histogram
    await expect(page.getByTestId('latency-histogram')).toBeVisible();
    
    // Check percentiles
    await expect(page.getByTestId('p50-latency')).toBeVisible();
    await expect(page.getByTestId('p95-latency')).toBeVisible();
    await expect(page.getByTestId('p99-latency')).toBeVisible();
  });

  test('configure custom metric alerts', async ({ page }) => {
    await projectPage.monitoringTab.click();
    
    // Open alerts configuration
    const alertsButton = page.getByTestId('configure-alerts-button');
    await alertsButton.click();
    
    const alertDialog = page.getByRole('dialog');
    await expect(alertDialog).toContainText('Configure Alerts');
    
    // Create CPU alert
    await alertDialog.getByTestId('add-alert-button').click();
    
    // Configure alert details
    await alertDialog.getByTestId('alert-name-input').fill('High CPU Usage Alert');
    await alertDialog.getByTestId('metric-select').selectOption('cpu_usage_percent');
    await alertDialog.getByTestId('condition-select').selectOption('greater_than');
    await alertDialog.getByTestId('threshold-input').fill('80');
    await alertDialog.getByTestId('duration-input').fill('5');
    await alertDialog.getByTestId('duration-unit-select').selectOption('minutes');
    
    // Configure notification channels
    await alertDialog.getByTestId('notify-email-checkbox').check();
    await alertDialog.getByTestId('email-recipients-input').fill('ops-team@example.com');
    
    await alertDialog.getByTestId('notify-slack-checkbox').check();
    await alertDialog.getByTestId('slack-channel-input').fill('#alerts');
    
    // Set severity
    await alertDialog.getByTestId('severity-select').selectOption('warning');
    
    // Save alert
    await alertDialog.getByTestId('save-alert-button').click();
    
    await expectNotification(page, 'Alert rule created successfully');
    
    // Verify alert in list
    const alertsList = page.getByTestId('alerts-list');
    await expect(alertsList).toContainText('High CPU Usage Alert');
    await expect(alertsList).toContainText('CPU > 80% for 5 minutes');
  });

  test('view Prometheus metrics and queries', async ({ page }) => {
    await projectPage.monitoringTab.click();
    
    // Open Prometheus query builder
    const prometheusTab = page.getByRole('tab', { name: /prometheus/i });
    await prometheusTab.click();
    
    // Verify query interface
    await expect(page.getByTestId('prometheus-query-input')).toBeVisible();
    await expect(page.getByTestId('query-builder-button')).toBeVisible();
    
    // Use query builder
    await page.getByTestId('query-builder-button').click();
    
    const queryBuilder = page.getByRole('dialog');
    await queryBuilder.getByTestId('metric-select').selectOption('container_cpu_usage_seconds_total');
    await queryBuilder.getByTestId('label-filter-add').click();
    await queryBuilder.getByTestId('label-key-0').fill('namespace');
    await queryBuilder.getByTestId('label-value-0').fill('monitoring-test-project');
    
    await queryBuilder.getByTestId('build-query-button').click();
    
    // Verify query is populated
    const queryInput = page.getByTestId('prometheus-query-input');
    await expect(queryInput).toHaveValue(/container_cpu_usage_seconds_total.*namespace="monitoring-test-project"/);
    
    // Execute query
    await page.getByTestId('execute-query-button').click();
    
    // Verify results
    await expect(page.getByTestId('query-results')).toBeVisible();
    await expect(page.getByTestId('results-chart')).toBeVisible();
    
    // Switch to table view
    await page.getByTestId('view-table-button').click();
    await expect(page.getByTestId('results-table')).toBeVisible();
  });

  test('access Grafana dashboards', async ({ page }) => {
    await projectPage.monitoringTab.click();
    
    // Open Grafana integration
    const grafanaTab = page.getByRole('tab', { name: /grafana/i });
    await grafanaTab.click();
    
    // Verify embedded Grafana or link
    const grafanaFrame = page.frameLocator('iframe[title="Grafana Dashboard"]');
    if (await grafanaFrame.locator('body').isVisible()) {
      // Embedded Grafana
      await expect(grafanaFrame.locator('.dashboard-header')).toBeVisible();
      await expect(grafanaFrame.locator('.panel-container')).toBeVisible();
    } else {
      // External Grafana link
      const grafanaLink = page.getByTestId('open-grafana-button');
      await expect(grafanaLink).toBeVisible();
      await expect(grafanaLink).toContainText('Open in Grafana');
      
      // Verify link has correct workspace context
      const href = await grafanaLink.getAttribute('href');
      expect(href).toContain('var-workspace=monitoring-test-project');
    }
  });

  test('monitor pod health and resource usage', async ({ page }) => {
    // Deploy application
    const appName = generateAppName();
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 3,
      port: 80,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    // Go to pods monitoring
    const podsTab = page.getByRole('tab', { name: /pods/i });
    await podsTab.click();
    
    // Get pod metrics
    const podMetrics = page.locator('[data-testid^="pod-metrics-"]');
    await expect(podMetrics).toHaveCount(3);
    
    // Check each pod's metrics
    for (let i = 0; i < 3; i++) {
      const pod = podMetrics.nth(i);
      
      // Verify metrics displayed
      await expect(pod.getByTestId('pod-cpu')).toBeVisible();
      await expect(pod.getByTestId('pod-memory')).toBeVisible();
      await expect(pod.getByTestId('pod-restarts')).toBeVisible();
      
      // Check health status
      const healthStatus = pod.getByTestId('pod-health-status');
      await expect(healthStatus).toBeVisible();
      await expect(healthStatus).toHaveClass(/healthy|warning|critical/);
      
      // Click for detailed metrics
      await pod.click();
      
      const detailDialog = page.getByRole('dialog');
      await expect(detailDialog).toContainText('Pod Metrics');
      
      // Verify detailed charts
      await expect(detailDialog.getByTestId('pod-cpu-chart')).toBeVisible();
      await expect(detailDialog.getByTestId('pod-memory-chart')).toBeVisible();
      await expect(detailDialog.getByTestId('pod-network-chart')).toBeVisible();
      
      await detailDialog.getByTestId('close-button').click();
    }
  });

  test('export metrics data', async ({ page }) => {
    await projectPage.monitoringTab.click();
    
    // Select time range
    await monitoringPage.selectTimeRange('24h');
    
    // Open export options
    const exportButton = page.getByTestId('export-metrics-button');
    await exportButton.click();
    
    const exportDialog = page.getByRole('dialog');
    
    // Select metrics to export
    await exportDialog.getByTestId('export-cpu-checkbox').check();
    await exportDialog.getByTestId('export-memory-checkbox').check();
    await exportDialog.getByTestId('export-network-checkbox').check();
    
    // Select format
    await exportDialog.getByTestId('export-format-select').selectOption('csv');
    
    // Configure granularity
    await exportDialog.getByTestId('granularity-select').selectOption('5m');
    
    // Export
    const downloadPromise = page.waitForEvent('download');
    await exportDialog.getByTestId('export-button').click();
    
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/metrics-export-.*\.csv/);
    
    // Verify export notification
    await expectNotification(page, 'Metrics exported successfully');
  });

  test('view logs with Loki integration', async ({ page }) => {
    // Deploy application
    const appName = generateAppName();
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 1,
      port: 80,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    // Go to logs tab
    await applicationPage.logsTab.click();
    
    // Verify Loki query interface
    await expect(page.getByTestId('loki-query-builder')).toBeVisible();
    
    // Use LogQL query
    const logqlInput = page.getByTestId('logql-query-input');
    await logqlInput.fill(`{app="${appName}"} |= "error"`);
    
    // Execute query
    await page.getByTestId('search-logs-button').click();
    
    // Verify log results
    await expect(page.getByTestId('log-entries')).toBeVisible();
    
    // Test log filtering
    await page.getByTestId('log-level-filter').selectOption('error');
    await page.getByTestId('apply-filters-button').click();
    
    // Verify filtered results
    const logEntries = page.locator('[data-testid^="log-entry-"]');
    for (const entry of await logEntries.all()) {
      await expect(entry).toContainText(/error|ERROR/i);
    }
    
    // Test log streaming
    const streamToggle = page.getByTestId('log-stream-toggle');
    await streamToggle.check();
    
    await expect(page.getByTestId('streaming-indicator')).toBeVisible();
    await expect(page.getByTestId('streaming-indicator')).toContainText('Live');
  });

  test('monitor cluster-wide resources', async ({ page }) => {
    // Navigate to workspace-level monitoring
    await page.getByTestId('breadcrumb-workspace').click();
    
    // Go to cluster monitoring
    const clusterMonitoringButton = page.getByTestId('cluster-monitoring-button');
    await clusterMonitoringButton.click();
    
    // Verify cluster overview
    await expect(page.getByTestId('cluster-nodes-card')).toBeVisible();
    await expect(page.getByTestId('cluster-cpu-card')).toBeVisible();
    await expect(page.getByTestId('cluster-memory-card')).toBeVisible();
    await expect(page.getByTestId('cluster-pods-card')).toBeVisible();
    
    // Check node details
    const nodesTab = page.getByRole('tab', { name: /nodes/i });
    await nodesTab.click();
    
    const nodesList = page.locator('[data-testid^="node-item-"]');
    await expect(nodesList).toHaveCount(3); // Based on test workspace
    
    // Click on a node for details
    await nodesList.first().click();
    
    const nodeDialog = page.getByRole('dialog');
    await expect(nodeDialog).toContainText('Node Details');
    
    // Verify node metrics
    await expect(nodeDialog.getByTestId('node-cpu-usage')).toBeVisible();
    await expect(nodeDialog.getByTestId('node-memory-usage')).toBeVisible();
    await expect(nodeDialog.getByTestId('node-disk-usage')).toBeVisible();
    await expect(nodeDialog.getByTestId('node-pod-count')).toBeVisible();
    
    // Check node conditions
    await expect(nodeDialog.getByTestId('node-conditions')).toBeVisible();
    await expect(nodeDialog.getByTestId('condition-Ready')).toContainText('True');
  });

  test('setup custom dashboards', async ({ page }) => {
    await projectPage.monitoringTab.click();
    
    // Create custom dashboard
    const customDashboardsTab = page.getByRole('tab', { name: /custom dashboards/i });
    await customDashboardsTab.click();
    
    const createDashboardButton = page.getByTestId('create-dashboard-button');
    await createDashboardButton.click();
    
    const dashboardDialog = page.getByRole('dialog');
    
    // Configure dashboard
    await dashboardDialog.getByTestId('dashboard-name-input').fill('Application Performance');
    await dashboardDialog.getByTestId('dashboard-description-input').fill('Key performance metrics for our applications');
    
    // Add widgets
    await dashboardDialog.getByTestId('add-widget-button').click();
    
    // Configure first widget - CPU usage
    const widgetConfig = page.getByTestId('widget-config-0');
    await widgetConfig.getByTestId('widget-type-select').selectOption('line-chart');
    await widgetConfig.getByTestId('widget-title-input').fill('CPU Usage Over Time');
    await widgetConfig.getByTestId('metric-select').selectOption('cpu_usage_percent');
    await widgetConfig.getByTestId('aggregation-select').selectOption('avg');
    
    // Add second widget - Request rate
    await dashboardDialog.getByTestId('add-widget-button').click();
    const widget2Config = page.getByTestId('widget-config-1');
    await widget2Config.getByTestId('widget-type-select').selectOption('gauge');
    await widget2Config.getByTestId('widget-title-input').fill('Current Request Rate');
    await widget2Config.getByTestId('metric-select').selectOption('http_requests_per_second');
    
    // Save dashboard
    await dashboardDialog.getByTestId('save-dashboard-button').click();
    
    await expectNotification(page, 'Dashboard created successfully');
    
    // Verify dashboard appears in list
    const dashboardsList = page.getByTestId('dashboards-list');
    await expect(dashboardsList).toContainText('Application Performance');
    
    // Open dashboard
    const dashboardCard = dashboardsList.getByText('Application Performance');
    await dashboardCard.click();
    
    // Verify widgets are displayed
    await expect(page.getByTestId('widget-cpu-usage-over-time')).toBeVisible();
    await expect(page.getByTestId('widget-current-request-rate')).toBeVisible();
  });

  test('monitor costs and resource optimization', async ({ page }) => {
    await projectPage.monitoringTab.click();
    
    // Navigate to cost monitoring
    const costTab = page.getByRole('tab', { name: /cost analysis/i });
    await costTab.click();
    
    // Verify cost overview
    await expect(page.getByTestId('total-cost-card')).toBeVisible();
    await expect(page.getByTestId('cost-trend-chart')).toBeVisible();
    
    // Check cost breakdown
    const costBreakdown = page.getByTestId('cost-breakdown');
    await expect(costBreakdown).toBeVisible();
    
    // Verify cost categories
    await expect(costBreakdown.getByTestId('compute-cost')).toBeVisible();
    await expect(costBreakdown.getByTestId('storage-cost')).toBeVisible();
    await expect(costBreakdown.getByTestId('network-cost')).toBeVisible();
    
    // Check optimization recommendations
    const optimizationTab = page.getByTestId('optimization-recommendations-tab');
    await optimizationTab.click();
    
    const recommendations = page.locator('[data-testid^="recommendation-"]');
    await expect(recommendations).toHaveCount(3);
    
    // Verify recommendation details
    const firstRecommendation = recommendations.first();
    await expect(firstRecommendation).toContainText(/Underutilized resources|Right-size|Cost saving/);
    
    // Apply recommendation
    const applyButton = firstRecommendation.getByTestId('apply-recommendation-button');
    await applyButton.click();
    
    const confirmDialog = page.getByRole('dialog');
    await expect(confirmDialog).toContainText('Apply Optimization');
    await confirmDialog.getByTestId('confirm-apply-button').click();
    
    await expectNotification(page, 'Optimization applied successfully');
  });
});