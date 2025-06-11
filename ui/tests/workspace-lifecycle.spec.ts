import { test, expect } from '@playwright/test';

test.describe('Workspace Lifecycle', () => {
  test.beforeEach(async ({ page }) => {
    // Login and navigate to organization
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'admin@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    await page.waitForURL('/dashboard');
    
    // Navigate to workspaces
    await page.click('[data-testid="org-card-org-1"]');
    await page.click('a:has-text("Workspaces")');
  });

  test('TC-WS-001: Dedicated Workspace Provisioning', async ({ page }) => {
    // Start workspace creation
    await page.click('button:has-text("Create Workspace")');
    
    // Fill workspace details
    await page.fill('[data-testid="workspace-name-input"]', 'Production Environment');
    await page.fill('[data-testid="workspace-description-input"]', 'Main production workspace');
    
    // Select dedicated plan
    await page.click('[data-testid="plan-dedicated"]');
    
    // Configure resources
    await page.fill('[data-testid="cpu-input"]', '8');
    await page.fill('[data-testid="memory-input"]', '16');
    await page.fill('[data-testid="storage-input"]', '200');
    
    // Select region
    await page.selectOption('[data-testid="region-select"]', 'us-east-1');
    
    // Enable backup
    await page.check('[data-testid="enable-backup-checkbox"]');
    
    // Submit
    await page.click('button:has-text("Create Workspace")');
    
    // Monitor provisioning status
    await expect(page.locator('[data-testid="workspace-status"]')).toContainText('creating');
    
    // Wait for real-time status updates
    await expect(page.locator('[data-testid="provisioning-step-vm"]')).toContainText('Creating VM...', { timeout: 10000 });
    await expect(page.locator('[data-testid="provisioning-step-vm"]')).toContainText('✓', { timeout: 60000 });
    
    await expect(page.locator('[data-testid="provisioning-step-k3s"]')).toContainText('Installing K3s...', { timeout: 10000 });
    await expect(page.locator('[data-testid="provisioning-step-k3s"]')).toContainText('✓', { timeout: 60000 });
    
    await expect(page.locator('[data-testid="provisioning-step-vcluster"]')).toContainText('Creating vCluster...', { timeout: 10000 });
    await expect(page.locator('[data-testid="provisioning-step-vcluster"]')).toContainText('✓', { timeout: 60000 });
    
    // Verify workspace is active
    await expect(page.locator('[data-testid="workspace-status"]')).toContainText('active', { timeout: 300000 });
    
    // Verify resource allocation
    await page.click('[data-testid="workspace-card-Production Environment"]');
    await expect(page.locator('[data-testid="resource-cpu"]')).toContainText('8 cores');
    await expect(page.locator('[data-testid="resource-memory"]')).toContainText('16 GB');
    await expect(page.locator('[data-testid="resource-storage"]')).toContainText('200 GB');
  });

  test('TC-WS-002: Shared Workspace Quick Setup', async ({ page }) => {
    // Create shared workspace
    await page.click('button:has-text("Create Workspace")');
    
    // Fill basic details
    await page.fill('[data-testid="workspace-name-input"]', 'Development');
    await page.fill('[data-testid="workspace-description-input"]', 'Development environment');
    
    // Select shared plan (default)
    await page.click('[data-testid="plan-shared"]');
    
    // Use default quotas
    await expect(page.locator('[data-testid="quota-info"]')).toContainText('2 CPU, 4GB RAM');
    
    // Submit
    await page.click('button:has-text("Create Workspace")');
    
    // Should be created quickly
    await expect(page.locator('[data-testid="workspace-status"]')).toContainText('active', { timeout: 30000 });
    
    // Verify can access immediately
    await page.click('[data-testid="workspace-card-Development"]');
    await expect(page).toHaveURL(/.*\/workspaces\/ws-.*$/);
    
    // Can create project immediately
    await page.click('button:has-text("Create Project")');
    await expect(page.locator('[data-testid="create-project-dialog"]')).toBeVisible();
  });

  test('TC-WS-003: Workspace Scaling', async ({ page }) => {
    // Navigate to existing workspace
    await page.click('[data-testid="workspace-card-production"]');
    
    // Go to settings
    await page.click('a:has-text("Settings")');
    
    // Click scale resources
    await page.click('button:has-text("Scale Resources")');
    
    // Current resources should be shown
    await expect(page.locator('[data-testid="current-cpu"]')).toContainText('4 cores');
    await expect(page.locator('[data-testid="current-memory"]')).toContainText('8 GB');
    
    // Increase resources
    await page.fill('[data-testid="new-cpu-input"]', '8');
    await page.fill('[data-testid="new-memory-input"]', '16');
    await page.fill('[data-testid="additional-storage-input"]', '100');
    
    // Show cost preview
    await expect(page.locator('[data-testid="cost-preview"]')).toContainText('Additional: $50/month');
    
    // Confirm scaling
    await page.click('button:has-text("Scale Workspace")');
    
    // Confirm in dialog
    await page.click('button:has-text("Confirm Scaling")');
    
    // Monitor scaling operation
    await expect(page.locator('[data-testid="scaling-status"]')).toContainText('Scaling in progress...');
    
    // Should complete without downtime
    await expect(page.locator('[data-testid="scaling-status"]')).toContainText('Scaling completed', { timeout: 180000 });
    
    // Verify new resources
    await expect(page.locator('[data-testid="resource-cpu"]')).toContainText('8 cores');
    await expect(page.locator('[data-testid="resource-memory"]')).toContainText('16 GB');
    
    // Applications should remain running
    await page.click('a:has-text("Applications")');
    await expect(page.locator('[data-testid="app-status"]').first()).toContainText('running');
  });

  test('Workspace deletion with confirmation', async ({ page }) => {
    // Create a test workspace first
    await page.click('button:has-text("Create Workspace")');
    await page.fill('[data-testid="workspace-name-input"]', 'To Delete');
    await page.click('[data-testid="plan-shared"]');
    await page.click('button:has-text("Create Workspace")');
    await expect(page.locator('[data-testid="workspace-card-To Delete"]')).toBeVisible({ timeout: 30000 });
    
    // Navigate to workspace
    await page.click('[data-testid="workspace-card-To Delete"]');
    await page.click('a:has-text("Settings")');
    
    // Click delete
    await page.click('button:has-text("Delete Workspace")');
    
    // Should show warning
    await expect(page.locator('[data-testid="delete-warning"]')).toContainText('This action cannot be undone');
    
    // Type workspace name to confirm
    await page.fill('[data-testid="confirm-workspace-name"]', 'To Delete');
    
    // Delete button should be enabled
    await page.click('button:has-text("Delete Permanently")');
    
    // Should redirect to workspaces list
    await expect(page).toHaveURL(/.*\/workspaces$/);
    
    // Workspace should be gone
    await expect(page.locator('[data-testid="workspace-card-To Delete"]')).not.toBeVisible();
  });
});