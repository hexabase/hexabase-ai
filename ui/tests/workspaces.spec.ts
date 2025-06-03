import { test, expect } from '@playwright/test';

test.describe('Workspace Management', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication for testing
    await page.goto('/');
    
    // Mock login flow - set auth token in localStorage or cookies
    await page.evaluate(() => {
      document.cookie = 'hexabase_token=test-token-12345; path=/';
    });
    
    // Navigate to dashboard page (assuming user is logged in)
    await page.goto('/dashboard');
  });

  test('should display workspace navigation in organization dashboard', async ({ page }) => {
    // Wait for organization to load
    await page.waitForSelector('[data-testid="open-organization-org1"]');
    
    // Click on first organization
    await page.click('[data-testid="open-organization-org1"]');
    
    // Should see workspace navigation
    await expect(page.locator('[data-testid="workspaces-tab"]')).toBeVisible();
    await expect(page.locator('text=Workspaces')).toBeVisible();
  });

  test('should display empty state when no workspaces exist', async ({ page }) => {
    // Navigate to organization
    await page.waitForSelector('[data-testid="open-organization-org1"]');
    await page.click('[data-testid="open-organization-org1"]');
    
    // Go to workspaces section
    await page.click('[data-testid="workspaces-nav"]');
    
    // Should show empty state
    await expect(page.locator('[data-testid="workspaces-empty-state"]')).toBeVisible();
    await expect(page.locator('text=No workspaces found')).toBeVisible();
    await expect(page.locator('[data-testid="create-workspace-button"]')).toBeVisible();
  });

  test('should open workspace creation modal', async ({ page }) => {
    // Navigate to workspaces section
    await page.waitForSelector('[data-testid="organization-card"]');
    await page.click('[data-testid="organization-card"]:first-child');
    await page.click('[data-testid="workspaces-nav"]');
    
    // Click create workspace button
    await page.click('[data-testid="create-workspace-button"]');
    
    // Modal should be visible
    await expect(page.locator('[data-testid="create-workspace-modal"]')).toBeVisible();
    await expect(page.locator('text=Create New Workspace')).toBeVisible();
    
    // Form fields should be present
    await expect(page.locator('[data-testid="workspace-name-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="plan-selection"]')).toBeVisible();
    await expect(page.locator('[data-testid="submit-workspace"]')).toBeVisible();
  });

  test('should validate workspace creation form', async ({ page }) => {
    // Open creation modal
    await page.waitForSelector('[data-testid="organization-card"]');
    await page.click('[data-testid="organization-card"]:first-child');
    await page.click('[data-testid="workspaces-nav"]');
    await page.click('[data-testid="create-workspace-button"]');
    
    // Try to submit empty form
    await page.click('[data-testid="submit-workspace"]');
    
    // Should show validation errors
    await expect(page.locator('[data-testid="name-error"]')).toBeVisible();
    await expect(page.locator('text=Workspace name is required')).toBeVisible();
  });

  test('should create new workspace with valid data', async ({ page }) => {
    // Mock the API response for workspace creation
    await page.route('**/api/v1/organizations/*/workspaces/', async route => {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'ws-test-123',
          name: 'Test Workspace',
          plan_id: 'plan-basic',
          vcluster_status: 'PENDING_CREATION',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        })
      });
    });

    // Mock plans API
    await page.route('**/api/v1/plans', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          plans: [
            {
              id: 'plan-basic',
              name: 'Basic Plan',
              description: 'Basic resources for development',
              price: 10.00,
              currency: 'usd'
            }
          ]
        })
      });
    });
    
    // Open creation modal and fill form
    await page.waitForSelector('[data-testid="organization-card"]');
    await page.click('[data-testid="organization-card"]:first-child');
    await page.click('[data-testid="workspaces-nav"]');
    await page.click('[data-testid="create-workspace-button"]');
    
    // Fill workspace details
    await page.fill('[data-testid="workspace-name-input"]', 'Test Workspace');
    await page.selectOption('[data-testid="plan-selection"]', 'plan-basic');
    
    // Submit form
    await page.click('[data-testid="submit-workspace"]');
    
    // Should show success message and close modal
    await expect(page.locator('[data-testid="success-message"]')).toBeVisible();
    await expect(page.locator('text=Workspace created successfully')).toBeVisible();
    
    // Modal should close
    await expect(page.locator('[data-testid="create-workspace-modal"]')).not.toBeVisible();
  });

  test('should display workspace list with status indicators', async ({ page }) => {
    // Mock workspaces API response
    await page.route('**/api/v1/organizations/*/workspaces/', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          workspaces: [
            {
              id: 'ws-1',
              name: 'Development Workspace',
              plan_id: 'plan-basic',
              vcluster_status: 'RUNNING',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString()
            },
            {
              id: 'ws-2', 
              name: 'Production Workspace',
              plan_id: 'plan-pro',
              vcluster_status: 'PENDING_CREATION',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString()
            }
          ],
          total: 2
        })
      });
    });
    
    // Navigate to workspaces
    await page.waitForSelector('[data-testid="organization-card"]');
    await page.click('[data-testid="organization-card"]:first-child');
    await page.click('[data-testid="workspaces-nav"]');
    
    // Should display workspace cards
    await expect(page.locator('[data-testid="workspace-card"]')).toHaveCount(2);
    
    // First workspace should show RUNNING status
    const runningWorkspace = page.locator('[data-testid="workspace-card"]').first();
    await expect(runningWorkspace.locator('text=Development Workspace')).toBeVisible();
    await expect(runningWorkspace.locator('[data-testid="status-running"]')).toBeVisible();
    
    // Second workspace should show PENDING status
    const pendingWorkspace = page.locator('[data-testid="workspace-card"]').nth(1);
    await expect(pendingWorkspace.locator('text=Production Workspace')).toBeVisible();
    await expect(pendingWorkspace.locator('[data-testid="status-pending"]')).toBeVisible();
  });

  test('should navigate to workspace detail page', async ({ page }) => {
    // Mock workspaces list
    await page.route('**/api/v1/organizations/*/workspaces/', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          workspaces: [{
            id: 'ws-1',
            name: 'Development Workspace',
            plan_id: 'plan-basic',
            vcluster_status: 'RUNNING',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          }],
          total: 1
        })
      });
    });

    // Mock workspace detail
    await page.route('**/api/v1/organizations/*/workspaces/ws-1', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'ws-1',
          name: 'Development Workspace',
          plan_id: 'plan-basic',
          vcluster_status: 'RUNNING',
          vcluster_config: '{"version": "0.15.0", "endpoint": "https://vcluster-ws-1.hexabase.com"}',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        })
      });
    });
    
    // Navigate to workspaces and click on workspace
    await page.waitForSelector('[data-testid="organization-card"]');
    await page.click('[data-testid="organization-card"]:first-child');
    await page.click('[data-testid="workspaces-nav"]');
    await page.click('[data-testid="workspace-card"]:first-child');
    
    // Should navigate to workspace detail page
    await expect(page.url()).toContain('/workspaces/ws-1');
    await expect(page.locator('[data-testid="workspace-detail"]')).toBeVisible();
    await expect(page.locator('text=Development Workspace')).toBeVisible();
  });

  test('should display vCluster health status', async ({ page }) => {
    // Mock workspace detail with health data
    await page.route('**/api/v1/organizations/*/workspaces/ws-1', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'ws-1',
          name: 'Development Workspace',
          vcluster_status: 'RUNNING',
          health: {
            healthy: true,
            components: {
              'api-server': 'healthy',
              'etcd': 'healthy',
              'controller-manager': 'healthy'
            },
            resource_usage: {
              cpu_usage: '45%',
              memory_usage: '2.1Gi/4Gi'
            }
          }
        })
      });
    });
    
    // Navigate to workspace detail
    await page.goto('/organizations/test-org/workspaces/ws-1');
    
    // Should display health information
    await expect(page.locator('[data-testid="health-status"]')).toBeVisible();
    await expect(page.locator('[data-testid="health-healthy"]')).toBeVisible();
    await expect(page.locator('text=API Server: healthy')).toBeVisible();
    await expect(page.locator('text=CPU: 45%')).toBeVisible();
  });

  test('should allow kubeconfig download', async ({ page }) => {
    // Mock kubeconfig API
    await page.route('**/api/v1/organizations/*/workspaces/*/kubeconfig', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          kubeconfig: 'apiVersion: v1\nkind: Config\n...',
          workspace: 'Development Workspace',
          status: 'RUNNING'
        })
      });
    });
    
    // Navigate to workspace detail
    await page.goto('/organizations/test-org/workspaces/ws-1');
    
    // Click download kubeconfig button
    const downloadPromise = page.waitForEvent('download');
    await page.click('[data-testid="download-kubeconfig"]');
    const download = await downloadPromise;
    
    // Verify download
    expect(download.suggestedFilename()).toBe('kubeconfig-ws-1.yaml');
  });

  test('should handle vCluster start/stop operations', async ({ page }) => {
    // Mock workspace with STOPPED status
    await page.route('**/api/v1/organizations/*/workspaces/ws-1', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'ws-1',
          name: 'Development Workspace',
          vcluster_status: 'STOPPED'
        })
      });
    });

    // Mock start operation
    await page.route('**/api/v1/organizations/*/workspaces/ws-1/vcluster/start', async route => {
      await route.fulfill({
        status: 202,
        contentType: 'application/json',
        body: JSON.stringify({
          task_id: 'task-123',
          status: 'start_initiated',
          message: 'vCluster start has been started'
        })
      });
    });
    
    // Navigate to workspace detail
    await page.goto('/organizations/test-org/workspaces/ws-1');
    
    // Should show start button for stopped workspace
    await expect(page.locator('[data-testid="start-vcluster"]')).toBeVisible();
    
    // Click start button
    await page.click('[data-testid="start-vcluster"]');
    
    // Should show success message
    await expect(page.locator('[data-testid="operation-success"]')).toBeVisible();
    await expect(page.locator('text=vCluster start initiated')).toBeVisible();
  });
});

test.describe('Workspace Management - Error Handling', () => {
  test('should handle API errors gracefully', async ({ page }) => {
    // Mock API error
    await page.route('**/api/v1/organizations/*/workspaces/', async route => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Internal server error'
        })
      });
    });
    
    // Navigate to workspaces
    await page.goto('/organizations/test-org/workspaces');
    
    // Should display error message
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('text=Failed to load workspaces')).toBeVisible();
  });

  test('should handle network errors', async ({ page }) => {
    // Mock network failure
    await page.route('**/api/v1/organizations/*/workspaces/', async route => {
      await route.abort('failed');
    });
    
    // Navigate to workspaces
    await page.goto('/organizations/test-org/workspaces');
    
    // Should display network error
    await expect(page.locator('[data-testid="network-error"]')).toBeVisible();
    await expect(page.locator('text=Unable to connect to server')).toBeVisible();
  });
});

test.describe('Workspace Management - Mobile', () => {
  test('should be responsive on mobile devices', async ({ page, isMobile }) => {
    if (!isMobile) {
      await page.setViewportSize({ width: 375, height: 812 }); // iPhone X size
    }
    
    // Mock workspaces
    await page.route('**/api/v1/organizations/*/workspaces/', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          workspaces: [{
            id: 'ws-1',
            name: 'Mobile Test Workspace',
            vcluster_status: 'RUNNING'
          }],
          total: 1
        })
      });
    });
    
    // Navigate to workspaces
    await page.goto('/organizations/test-org/workspaces');
    
    // Should display mobile-friendly layout
    await expect(page.locator('[data-testid="mobile-workspace-list"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspace-card"]')).toBeVisible();
  });
});