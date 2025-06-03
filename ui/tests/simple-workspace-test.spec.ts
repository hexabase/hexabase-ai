import { test, expect } from '@playwright/test';

test.describe('Workspace Components Test', () => {

  test('should display workspace list interface', async ({ page }) => {
    await page.goto('/test-workspace');
    
    // Wait for the page to load
    await page.waitForSelector('h1:has-text("Workspace Management Test")');
    
    // Check if the workspace list component is present
    await expect(page.locator('h2:has-text("Workspaces")')).toBeVisible();
    
    // Take a screenshot
    await page.screenshot({ 
      path: 'test-results/workspace-list-interface.png',
      fullPage: true 
    });
  });

  test('should display workspace detail interface', async ({ page }) => {
    await page.goto('/test-workspace/detail');
    
    // Wait for the page to load
    await page.waitForSelector('[data-testid="workspace-detail-page"]');
    
    // Check workspace details are visible
    await expect(page.locator('[data-testid="workspace-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspace-status"]')).toBeVisible();
    
    // Check health status cards
    await expect(page.locator('[data-testid="health-status-card"]')).toBeVisible();
    await expect(page.locator('[data-testid="nodes-card"]')).toBeVisible();
    await expect(page.locator('[data-testid="cpu-usage-card"]')).toBeVisible();
    await expect(page.locator('[data-testid="memory-usage-card"]')).toBeVisible();
    
    // Check workspace actions
    await expect(page.locator('[data-testid="download-kubeconfig"]')).toBeVisible();
    await expect(page.locator('[data-testid="stop-vcluster"]')).toBeVisible();
    
    // Take a screenshot
    await page.screenshot({ 
      path: 'test-results/workspace-detail-interface.png',
      fullPage: true 
    });
  });

  test('should handle workspace actions', async ({ page }) => {
    await page.goto('/test-workspace/detail');
    
    // Wait for the page to load
    await page.waitForSelector('[data-testid="workspace-detail-page"]');
    
    // Test download kubeconfig action
    await page.click('[data-testid="download-kubeconfig"]');
    // Should see toast notification
    await expect(page.locator('text=Download Started')).toBeVisible();
    
    // Test stop vCluster action
    await page.click('[data-testid="stop-vcluster"]');
    // Should see toast notification
    await expect(page.locator('text=vCluster stop')).toBeVisible();
    
    // Test refresh health action
    await page.click('[data-testid="refresh-health"]');
    
    // Take a screenshot after actions
    await page.screenshot({ 
      path: 'test-results/workspace-actions-completed.png',
      fullPage: true 
    });
  });
});