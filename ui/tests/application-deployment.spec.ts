import { test, expect } from '@playwright/test';

test.describe('Application Deployment', () => {
  test.beforeEach(async ({ page }) => {
    // Setup: Login and navigate to a project
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'admin@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    await page.waitForURL('/dashboard');
    
    // Navigate to workspace project
    await page.click('[data-testid="org-card-org-1"]');
    await page.click('[data-testid="workspace-card-production"]');
    await page.click('[data-testid="project-card-frontend"]');
  });

  test('TC-APP-001: Stateless Application Deployment', async ({ page }) => {
    // Start deployment
    await page.click('button:has-text("Deploy Application")');
    
    // Select application type
    await page.click('[data-testid="app-type-stateless"]');
    
    // Fill application details
    await page.fill('[data-testid="app-name-input"]', 'web-frontend');
    await page.fill('[data-testid="app-description"]', 'Main web application frontend');
    
    // Configure container
    await page.fill('[data-testid="container-image"]', 'nginx:latest');
    await page.fill('[data-testid="container-port"]', '80');
    
    // Set replicas
    await page.fill('[data-testid="replicas-input"]', '3');
    
    // Add environment variables
    await page.click('button:has-text("Add Environment Variable")');
    await page.fill('[data-testid="env-key-0"]', 'API_URL');
    await page.fill('[data-testid="env-value-0"]', 'https://api.example.com');
    
    // Set resource limits
    await page.fill('[data-testid="cpu-request"]', '100m');
    await page.fill('[data-testid="cpu-limit"]', '500m');
    await page.fill('[data-testid="memory-request"]', '128Mi');
    await page.fill('[data-testid="memory-limit"]', '512Mi');
    
    // Deploy
    await page.click('button:has-text("Deploy")');
    
    // Monitor deployment progress
    await expect(page.locator('[data-testid="deployment-status"]')).toContainText('Creating');
    
    // Real-time pod status updates
    await expect(page.locator('[data-testid="pod-count"]')).toContainText('0/3 pods ready');
    await expect(page.locator('[data-testid="pod-count"]')).toContainText('1/3 pods ready', { timeout: 30000 });
    await expect(page.locator('[data-testid="pod-count"]')).toContainText('2/3 pods ready', { timeout: 30000 });
    await expect(page.locator('[data-testid="pod-count"]')).toContainText('3/3 pods ready', { timeout: 30000 });
    
    // Deployment should be successful
    await expect(page.locator('[data-testid="deployment-status"]')).toContainText('Running', { timeout: 60000 });
    
    // Verify service endpoint
    await expect(page.locator('[data-testid="service-endpoint"]')).toContainText('web-frontend.frontend.svc.cluster.local');
    
    // Check health status
    await expect(page.locator('[data-testid="health-check-status"]')).toContainText('Healthy');
  });

  test('TC-APP-002: Stateful Application with Storage', async ({ page }) => {
    // Deploy PostgreSQL
    await page.click('button:has-text("Deploy Application")');
    
    // Select stateful type
    await page.click('[data-testid="app-type-stateful"]');
    
    // Select template
    await page.click('[data-testid="template-postgresql"]');
    
    // Customize deployment
    await page.fill('[data-testid="app-name-input"]', 'main-database');
    
    // Configure storage
    await page.fill('[data-testid="storage-size"]', '10');
    await page.selectOption('[data-testid="storage-class"]', 'fast-ssd');
    
    // Set backup policy
    await page.check('[data-testid="enable-backup"]');
    await page.selectOption('[data-testid="backup-schedule"]', 'daily');
    await page.fill('[data-testid="backup-retention"]', '30');
    
    // Configure resources
    await page.fill('[data-testid="cpu-limit"]', '2');
    await page.fill('[data-testid="memory-limit"]', '4Gi');
    
    // Set password
    await page.fill('[data-testid="db-password"]', 'SecurePass123!');
    
    // Deploy
    await page.click('button:has-text("Deploy")');
    
    // Monitor StatefulSet creation
    await expect(page.locator('[data-testid="statefulset-status"]')).toContainText('Creating');
    
    // PVC should be created
    await expect(page.locator('[data-testid="pvc-status"]')).toContainText('Pending', { timeout: 10000 });
    await expect(page.locator('[data-testid="pvc-status"]')).toContainText('Bound', { timeout: 30000 });
    
    // Pod should start
    await expect(page.locator('[data-testid="pod-0-status"]')).toContainText('Running', { timeout: 60000 });
    
    // Test data persistence
    await page.click('[data-testid="test-connection"]');
    await expect(page.locator('[data-testid="connection-test-result"]')).toContainText('Connection successful');
    
    // Simulate pod restart
    await page.click('[data-testid="restart-pod-0"]');
    await expect(page.locator('[data-testid="pod-0-status"]')).toContainText('Terminating');
    await expect(page.locator('[data-testid="pod-0-status"]')).toContainText('Running', { timeout: 60000 });
    
    // Verify data persisted
    await page.click('[data-testid="test-connection"]');
    await expect(page.locator('[data-testid="connection-test-result"]')).toContainText('Connection successful');
    await expect(page.locator('[data-testid="data-check"]')).toContainText('Data intact');
  });

  test('TC-APP-003: Rolling Update', async ({ page }) => {
    // Navigate to existing application
    await page.click('[data-testid="app-card-web-frontend"]');
    
    // Current version info
    await expect(page.locator('[data-testid="current-image"]')).toContainText('nginx:1.20');
    await expect(page.locator('[data-testid="pod-count"]')).toContainText('3/3 pods ready');
    
    // Start update
    await page.click('button:has-text("Update")');
    
    // Change image version
    await page.fill('[data-testid="new-image"]', 'nginx:1.21');
    
    // Add new environment variable
    await page.click('button:has-text("Add Environment Variable")');
    await page.fill('[data-testid="env-key-new"]', 'FEATURE_FLAG');
    await page.fill('[data-testid="env-value-new"]', 'enabled');
    
    // Configure rolling update strategy
    await page.fill('[data-testid="max-unavailable"]', '1');
    await page.fill('[data-testid="max-surge"]', '1');
    
    // Apply update
    await page.click('button:has-text("Apply Rolling Update")');
    
    // Monitor rolling update progress
    await expect(page.locator('[data-testid="update-status"]')).toContainText('Updating');
    
    // Should see pods updating one by one
    await expect(page.locator('[data-testid="pod-web-frontend-0"]')).toContainText('1.20');
    await expect(page.locator('[data-testid="pod-web-frontend-1"]')).toContainText('1.20');
    await expect(page.locator('[data-testid="pod-web-frontend-2"]')).toContainText('1.20');
    
    // First pod updates
    await expect(page.locator('[data-testid="pod-web-frontend-0"]')).toContainText('Terminating', { timeout: 10000 });
    await expect(page.locator('[data-testid="pod-web-frontend-0"]')).toContainText('1.21', { timeout: 30000 });
    
    // Verify service remains available
    await page.click('[data-testid="check-availability"]');
    await expect(page.locator('[data-testid="availability-status"]')).toContainText('Service available');
    
    // Second pod updates
    await expect(page.locator('[data-testid="pod-web-frontend-1"]')).toContainText('1.21', { timeout: 30000 });
    
    // Third pod updates
    await expect(page.locator('[data-testid="pod-web-frontend-2"]')).toContainText('1.21', { timeout: 30000 });
    
    // Update complete
    await expect(page.locator('[data-testid="update-status"]')).toContainText('Update completed');
    await expect(page.locator('[data-testid="current-image"]')).toContainText('nginx:1.21');
    
    // Rollback available
    await expect(page.locator('button:has-text("Rollback")')).toBeEnabled();
  });

  test('Application deletion with dependencies', async ({ page }) => {
    // Navigate to application with dependencies
    await page.click('[data-testid="app-card-backend-api"]');
    
    // Attempt deletion
    await page.click('button:has-text("Delete Application")');
    
    // Should show dependencies
    await expect(page.locator('[data-testid="dependency-warning"]')).toContainText('This application has dependencies');
    await expect(page.locator('[data-testid="dependent-service-1"]')).toContainText('frontend-app (depends on API)');
    
    // Force delete option
    await page.check('[data-testid="force-delete"]');
    
    // Confirm deletion
    await page.fill('[data-testid="confirm-app-name"]', 'backend-api');
    await page.click('button:has-text("Delete Permanently")');
    
    // Monitor deletion
    await expect(page.locator('[data-testid="deletion-status"]')).toContainText('Deleting');
    
    // Should clean up resources
    await expect(page.locator('[data-testid="cleanup-pods"]')).toContainText('Terminating pods...');
    await expect(page.locator('[data-testid="cleanup-service"]')).toContainText('Removing service...');
    await expect(page.locator('[data-testid="cleanup-configmap"]')).toContainText('Cleaning config...');
    
    // Deletion complete
    await expect(page).toHaveURL(/.*\/projects\/.*$/);
    await expect(page.locator('[data-testid="app-card-backend-api"]')).not.toBeVisible();
  });
});