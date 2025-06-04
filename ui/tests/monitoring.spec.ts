import { test, expect } from '@playwright/test';

test.describe('Monitoring & Observability', () => {
  
  test('should display cluster health overview', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/monitoring');
    
    // Wait for monitoring dashboard to load
    await page.waitForSelector('[data-testid="monitoring-dashboard"]');
    
    // Should see cluster health status
    await expect(page.locator('[data-testid="cluster-health"]')).toBeVisible();
    await expect(page.locator('[data-testid="health-status"]')).toContainText('Healthy');
    await expect(page.locator('[data-testid="cluster-uptime"]')).toBeVisible();
    
    // Should see resource utilization
    await expect(page.locator('[data-testid="cpu-utilization"]')).toBeVisible();
    await expect(page.locator('[data-testid="memory-utilization"]')).toBeVisible();
    await expect(page.locator('[data-testid="storage-utilization"]')).toBeVisible();
    await expect(page.locator('[data-testid="network-utilization"]')).toBeVisible();
    
    // Should see node status
    await expect(page.locator('[data-testid="nodes-overview"]')).toBeVisible();
    await expect(page.locator('[data-testid="total-nodes"]')).toBeVisible();
    await expect(page.locator('[data-testid="healthy-nodes"]')).toBeVisible();
  });

  test('should display real-time metrics charts', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/monitoring');
    await page.waitForSelector('[data-testid="monitoring-dashboard"]');
    
    // Should see metrics charts
    await expect(page.locator('[data-testid="cpu-usage-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="memory-usage-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="network-io-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="disk-io-chart"]')).toBeVisible();
    
    // Should have time range selector
    await expect(page.locator('[data-testid="time-range-selector"]')).toBeVisible();
    
    // Test time range selection
    await page.selectOption('[data-testid="time-range-selector"]', '1h');
    await page.waitForTimeout(500);
    await page.selectOption('[data-testid="time-range-selector"]', '24h');
    await page.waitForTimeout(500);
    await page.selectOption('[data-testid="time-range-selector"]', '7d');
  });

  test('should display workspace-level metrics', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/monitoring');
    await page.waitForSelector('[data-testid="monitoring-dashboard"]');
    
    // Should see workspace metrics
    await expect(page.locator('[data-testid="workspace-metrics"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspace-card"]').first()).toBeVisible();
    
    // Should show per-workspace details
    await expect(page.locator('[data-testid="workspace-name"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="workspace-cpu"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="workspace-memory"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="workspace-pods"]').first()).toBeVisible();
    
    // Click on workspace for detailed view
    await page.click('[data-testid="workspace-card"]');
    await expect(page.locator('[data-testid="workspace-detail-modal"]')).toBeVisible();
    await expect(page.locator('[data-testid="namespace-metrics"]')).toBeVisible();
  });

  test('should display alerts and incidents', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/monitoring/alerts');
    
    // Wait for alerts page
    await page.waitForSelector('[data-testid="alerts-dashboard"]');
    
    // Should see active alerts
    await expect(page.locator('[data-testid="active-alerts"]')).toBeVisible();
    await expect(page.locator('[data-testid="alert-item"]').first()).toBeVisible();
    
    // Should see alert details
    await expect(page.locator('[data-testid="alert-severity"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="alert-title"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="alert-time"]').first()).toBeVisible();
    
    // Should have alert filters
    await expect(page.locator('[data-testid="severity-filter"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspace-filter"]')).toBeVisible();
    
    // Test severity filter
    await page.selectOption('[data-testid="severity-filter"]', 'critical');
    await expect(page.locator('[data-testid="alert-severity"]').first()).toContainText('Critical');
    
    // Should see incident history
    await expect(page.locator('[data-testid="incident-history"]')).toBeVisible();
  });

  test('should display logs viewer', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/monitoring/logs');
    
    // Wait for logs viewer
    await page.waitForSelector('[data-testid="logs-viewer"]');
    
    // Should see log filters
    await expect(page.locator('[data-testid="workspace-selector"]')).toBeVisible();
    await expect(page.locator('[data-testid="namespace-selector"]')).toBeVisible();
    await expect(page.locator('[data-testid="pod-selector"]')).toBeVisible();
    await expect(page.locator('[data-testid="log-level-filter"]')).toBeVisible();
    
    // Should see log entries
    await expect(page.locator('[data-testid="log-entries"]')).toBeVisible();
    await expect(page.locator('[data-testid="log-entry"]').first()).toBeVisible();
    
    // Should have search functionality
    await expect(page.locator('[data-testid="log-search"]')).toBeVisible();
    await page.fill('[data-testid="log-search"]', 'error');
    
    // Should have export button
    await expect(page.locator('[data-testid="export-logs"]')).toBeVisible();
    
    // Test real-time streaming toggle
    await expect(page.locator('[data-testid="stream-logs"]')).toBeVisible();
    await page.click('[data-testid="stream-logs"]');
  });

  test('should display performance insights', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/monitoring/insights');
    
    // Wait for insights page
    await page.waitForSelector('[data-testid="performance-insights"]');
    
    // Should see recommendations
    await expect(page.locator('[data-testid="recommendations"]')).toBeVisible();
    await expect(page.locator('[data-testid="recommendation-card"]').first()).toBeVisible();
    
    // Should see cost optimization suggestions
    await expect(page.locator('[data-testid="cost-optimization"]')).toBeVisible();
    await expect(page.locator('[data-testid="potential-savings"]')).toBeVisible();
    
    // Should see performance bottlenecks
    await expect(page.locator('[data-testid="bottlenecks"]')).toBeVisible();
    await expect(page.locator('[data-testid="bottleneck-item"]').first()).toBeVisible();
    
    // Should see resource predictions
    await expect(page.locator('[data-testid="resource-predictions"]')).toBeVisible();
    await expect(page.locator('[data-testid="cpu-prediction-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="memory-prediction-chart"]')).toBeVisible();
  });

  test('should configure monitoring settings', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/monitoring/settings');
    
    // Wait for settings page
    await page.waitForSelector('[data-testid="monitoring-settings"]');
    
    // Should see alert configurations
    await expect(page.locator('[data-testid="alert-rules"]')).toBeVisible();
    await expect(page.locator('[data-testid="add-alert-rule"]')).toBeVisible();
    
    // Click add alert rule
    await page.click('[data-testid="add-alert-rule"]');
    await expect(page.locator('[data-testid="alert-rule-modal"]')).toBeVisible();
    
    // Should see metric thresholds
    await expect(page.locator('[data-testid="metric-thresholds"]')).toBeVisible();
    await expect(page.locator('[data-testid="cpu-threshold"]')).toBeVisible();
    await expect(page.locator('[data-testid="memory-threshold"]')).toBeVisible();
    
    // Should see notification channels
    await expect(page.locator('[data-testid="notification-channels"]')).toBeVisible();
    await expect(page.locator('[data-testid="email-channel"]')).toBeVisible();
    await expect(page.locator('[data-testid="slack-channel"]')).toBeVisible();
    await expect(page.locator('[data-testid="webhook-channel"]')).toBeVisible();
    
    // Should see data retention settings
    await expect(page.locator('[data-testid="data-retention"]')).toBeVisible();
    await expect(page.locator('[data-testid="metrics-retention"]')).toBeVisible();
    await expect(page.locator('[data-testid="logs-retention"]')).toBeVisible();
  });
});