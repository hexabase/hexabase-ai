import { test, expect } from '@playwright/test';

test.describe('Hexabase KaaS - Full Feature Showcase', () => {
  
  test.beforeEach(async ({ page, context }) => {
    // Set auth cookie through context for authenticated requests
    await context.addCookies([{
      name: 'hexabase_token',
      value: 'test-token-12345',
      domain: 'localhost',
      path: '/'
    }]);
  });

  test('should capture complete feature showcase', async ({ page }) => {
    // Navigate to showcase page
    await page.goto('/test-showcase');
    await page.waitForSelector('[data-testid="showcase-page"]');
    
    // Capture overview section
    await expect(page.locator('[data-testid="overview-section"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/01-overview.png',
      fullPage: true 
    });
    
    // Navigate to workspaces section
    await page.click('[data-testid="workspaces-button"]');
    await expect(page.locator('[data-testid="workspaces-section"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/02-workspaces.png',
      fullPage: true 
    });
    
    // Navigate to projects section
    await page.click('[data-testid="projects-button"]');
    await expect(page.locator('[data-testid="projects-section"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/03-projects.png',
      fullPage: true 
    });
    
    // Navigate to billing section
    await page.click('[data-testid="billing-button"]');
    await expect(page.locator('[data-testid="billing-section"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/04-billing.png',
      fullPage: true 
    });
    
    // Navigate to monitoring section
    await page.click('[data-testid="monitoring-button"]');
    await expect(page.locator('[data-testid="monitoring-section"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/05-monitoring.png',
      fullPage: true 
    });
  });

  test('should capture organization dashboard', async ({ page }) => {
    // Navigate to organization dashboard
    await page.goto('/dashboard/organizations/org1');
    await page.waitForSelector('[data-testid="organization-name"]');
    
    // Capture dashboard overview
    await page.screenshot({ 
      path: 'test-results/success/10-org-dashboard.png',
      fullPage: true 
    });
    
    // Navigate to workspaces tab
    await page.click('[data-testid="workspaces-tab"]');
    await expect(page.locator('[data-testid="workspaces-section"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/11-org-workspaces.png',
      fullPage: true 
    });
    
    // Navigate to projects tab
    await page.click('[data-testid="projects-tab"]');
    await expect(page.locator('[data-testid="projects-section"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/12-org-projects.png',
      fullPage: true 
    });
    
    // Navigate to billing tab
    await page.click('[data-testid="billing-tab"]');
    await expect(page.locator('[data-testid="billing-section"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/13-org-billing.png',
      fullPage: true 
    });
    
    // Navigate to monitoring tab
    await page.click('[data-testid="monitoring-tab"]');
    await expect(page.locator('[data-testid="monitoring-section"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/14-org-monitoring-link.png',
      fullPage: true 
    });
  });

  test('should capture project management features', async ({ page }) => {
    // Navigate to test projects page
    await page.goto('/test-projects');
    await page.waitForSelector('[data-testid="projects-test-page"]');
    
    // Capture project list
    await page.screenshot({ 
      path: 'test-results/success/20-project-list.png',
      fullPage: true 
    });
    
    // Navigate to project detail
    await page.click('[data-testid="project-card"]');
    await page.waitForSelector('[data-testid="project-detail-page"]');
    await page.screenshot({ 
      path: 'test-results/success/21-project-detail.png',
      fullPage: true 
    });
    
    // Open project settings modal
    await page.click('[data-testid="project-settings-button"]');
    await expect(page.locator('[data-testid="project-settings-modal"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/22-project-settings.png',
      fullPage: true 
    });
    
    // Close modal and open create namespace dialog
    await page.press('body', 'Escape');
    await page.click('[data-testid="create-namespace-button"]');
    await expect(page.locator('[data-testid="create-namespace-modal"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/23-create-namespace.png',
      fullPage: true 
    });
  });

  test('should capture billing and subscription features', async ({ page }) => {
    // Navigate to billing page
    await page.goto('/dashboard/organizations/org1/billing');
    await page.waitForSelector('[data-testid="billing-page"]');
    
    // Capture billing overview
    await page.screenshot({ 
      path: 'test-results/success/30-billing-overview.png',
      fullPage: true 
    });
    
    // Navigate to billing history
    await page.goto('/dashboard/organizations/org1/billing/history');
    await page.waitForSelector('[data-testid="billing-history"]');
    await page.screenshot({ 
      path: 'test-results/success/31-billing-history.png',
      fullPage: true 
    });
    
    // Navigate to usage analytics
    await page.goto('/dashboard/organizations/org1/billing/usage');
    await page.waitForSelector('[data-testid="usage-analytics"]');
    await page.screenshot({ 
      path: 'test-results/success/32-usage-analytics.png',
      fullPage: true 
    });
    
    // Navigate to billing settings
    await page.goto('/dashboard/organizations/org1/billing/settings');
    await page.waitForSelector('[data-testid="billing-settings"]');
    await page.screenshot({ 
      path: 'test-results/success/33-billing-settings.png',
      fullPage: true 
    });
  });

  test('should capture monitoring and observability features', async ({ page }) => {
    // Navigate to monitoring dashboard
    await page.goto('/dashboard/organizations/org1/monitoring');
    await page.waitForSelector('[data-testid="monitoring-dashboard"]');
    
    // Capture monitoring overview
    await page.screenshot({ 
      path: 'test-results/success/40-monitoring-dashboard.png',
      fullPage: true 
    });
    
    // Navigate to alerts page
    await page.goto('/dashboard/organizations/org1/monitoring/alerts');
    await page.waitForSelector('[data-testid="alerts-page"]');
    await page.screenshot({ 
      path: 'test-results/success/41-monitoring-alerts.png',
      fullPage: true 
    });
    
    // Navigate to logs viewer
    await page.goto('/dashboard/organizations/org1/monitoring/logs');
    await page.waitForSelector('[data-testid="logs-viewer"]');
    await page.screenshot({ 
      path: 'test-results/success/42-logs-viewer.png',
      fullPage: true 
    });
    
    // Navigate to insights page
    await page.goto('/dashboard/organizations/org1/monitoring/insights');
    await page.waitForSelector('[data-testid="performance-insights"]');
    await page.screenshot({ 
      path: 'test-results/success/43-performance-insights.png',
      fullPage: true 
    });
    
    // Navigate to monitoring settings
    await page.goto('/dashboard/organizations/org1/monitoring/settings');
    await page.waitForSelector('[data-testid="monitoring-settings"]');
    await page.screenshot({ 
      path: 'test-results/success/44-monitoring-settings.png',
      fullPage: true 
    });
  });

  test('should capture modal and dialog interactions', async ({ page }) => {
    // Navigate to monitoring settings for alert rule modal
    await page.goto('/dashboard/organizations/org1/monitoring/settings');
    await page.waitForSelector('[data-testid="monitoring-settings"]');
    
    // Open alert rule modal
    await page.click('[data-testid="add-alert-rule"]');
    await expect(page.locator('[data-testid="alert-rule-modal"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/50-alert-rule-modal.png',
      fullPage: true 
    });
    
    // Navigate to organization dashboard for workspace detail modal
    await page.goto('/dashboard/organizations/org1');
    await page.waitForSelector('[data-testid="organization-name"]');
    
    // Click monitoring tab and go to monitoring dashboard
    await page.click('[data-testid="monitoring-tab"]');
    await page.click('text=Go to Monitoring Dashboard');
    await page.waitForSelector('[data-testid="monitoring-dashboard"]');
    
    // Open workspace detail modal
    const workspaceCard = page.locator('[data-testid="workspace-card"]').first();
    await workspaceCard.click();
    await expect(page.locator('[data-testid="workspace-detail-modal"]')).toBeVisible();
    await page.screenshot({ 
      path: 'test-results/success/51-workspace-detail-modal.png',
      fullPage: true 
    });
  });

  test('should capture login page', async ({ page }) => {
    // Clear cookies to simulate unauthenticated state
    await page.context().clearCookies();
    
    // Navigate to dashboard (should redirect to login)
    await page.goto('/dashboard');
    await page.waitForSelector('[data-testid="login-page"]');
    
    // Capture login page
    await page.screenshot({ 
      path: 'test-results/success/00-login-page.png',
      fullPage: true 
    });
  });

  test('should capture organizations dashboard', async ({ page }) => {
    // Navigate to main dashboard
    await page.goto('/dashboard');
    await page.waitForSelector('[data-testid="dashboard-page"]');
    
    // Capture organizations list
    await page.screenshot({ 
      path: 'test-results/success/60-organizations-dashboard.png',
      fullPage: true 
    });
  });
});