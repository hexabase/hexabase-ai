import { test, expect } from '@playwright/test';

test.describe('Workspace UI Screenshots', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication
    await page.goto('http://localhost:3001');
    await page.evaluate(() => {
      document.cookie = 'hexabase_access_token=mock-token; path=/';
      localStorage.setItem('hexabase_user', JSON.stringify({
        id: 'test-user-123',
        email: 'test@hexabase.com',
        name: 'Test User',
        provider: 'test'
      }));
    });
  });

  test('capture workspace listing page', async ({ page }) => {
    // Mock API responses
    await page.route('**/api/v1/organizations/*/workspaces/', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          workspaces: [
            {
              id: 'ws-1',
              name: 'Production Workspace',
              plan_id: 'plan-pro',
              vcluster_status: 'RUNNING',
              vcluster_instance_name: 'vcluster-prod-001',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: new Date().toISOString()
            },
            {
              id: 'ws-2',
              name: 'Development Workspace',
              plan_id: 'plan-dev',
              vcluster_status: 'STOPPED',
              vcluster_instance_name: 'vcluster-dev-001',
              created_at: '2024-01-02T00:00:00Z',
              updated_at: new Date().toISOString()
            },
            {
              id: 'ws-3',
              name: 'Staging Environment',
              plan_id: 'plan-starter',
              vcluster_status: 'PENDING_CREATION',
              created_at: '2024-01-03T00:00:00Z',
              updated_at: new Date().toISOString()
            }
          ],
          total: 3
        })
      });
    });

    // Navigate to workspaces page
    await page.goto('http://localhost:3001/dashboard/organizations/org-123/workspaces');
    
    // Wait for content to load
    await page.waitForSelector('[data-testid="workspace-list"]', { timeout: 10000 });
    
    // Take screenshot
    await page.screenshot({ 
      path: 'screenshots/workspace-listing.png',
      fullPage: true 
    });
  });

  test('capture workspace creation dialog', async ({ page }) => {
    // Mock plans API
    await page.route('**/api/v1/plans', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          plans: [
            {
              id: 'plan-starter',
              name: 'Starter',
              description: 'Perfect for small projects',
              price: 0,
              currency: 'USD',
              resource_limits: JSON.stringify({
                cpu: '2 cores',
                memory: '4 GB',
                storage: '10 GB'
              })
            },
            {
              id: 'plan-pro',
              name: 'Professional',
              description: 'For production workloads',
              price: 99,
              currency: 'USD',
              resource_limits: JSON.stringify({
                cpu: '8 cores',
                memory: '16 GB',
                storage: '100 GB'
              })
            },
            {
              id: 'plan-enterprise',
              name: 'Enterprise',
              description: 'Unlimited resources with SLA',
              price: 499,
              currency: 'USD',
              resource_limits: JSON.stringify({
                cpu: 'Unlimited',
                memory: 'Unlimited',
                storage: 'Unlimited'
              })
            }
          ],
          total: 3
        })
      });
    });

    await page.goto('http://localhost:3001/dashboard/organizations/org-123/workspaces');
    
    // Open create dialog
    await page.click('button:has-text("Create Workspace")');
    
    // Wait for dialog
    await page.waitForSelector('[role="dialog"]');
    
    // Fill in workspace name
    await page.fill('input[name="name"]', 'My New Workspace');
    
    // Take screenshot
    await page.screenshot({ 
      path: 'screenshots/workspace-create-dialog.png',
      fullPage: true 
    });
  });

  test('capture workspace detail page', async ({ page }) => {
    // Mock workspace detail API
    await page.route('**/api/v1/organizations/*/workspaces/ws-1', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'ws-1',
          name: 'Production Workspace',
          plan_id: 'plan-pro',
          vcluster_status: 'RUNNING',
          vcluster_instance_name: 'vcluster-prod-001',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: new Date().toISOString()
        })
      });
    });

    // Mock vCluster health API
    await page.route('**/api/v1/organizations/*/workspaces/ws-1/vcluster/health', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          healthy: true,
          components: {
            'api-server': 'healthy',
            'etcd': 'healthy',
            'scheduler': 'healthy',
            'controller-manager': 'healthy'
          },
          resource_usage: {
            'cpu': '23.5%',
            'memory': '67.2%',
            'nodes': '3',
            'pods': '42'
          },
          last_checked: new Date().toISOString()
        })
      });
    });

    await page.goto('http://localhost:3001/dashboard/organizations/org-123/workspaces/ws-1');
    
    // Wait for content to load
    await page.waitForSelector('[data-testid="workspace-detail-page"]', { timeout: 10000 });
    
    // Take screenshot
    await page.screenshot({ 
      path: 'screenshots/workspace-detail.png',
      fullPage: true 
    });
  });
});