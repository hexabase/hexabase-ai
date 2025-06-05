import { test, expect } from '@playwright/test';

test.describe('Workspace Management Demo', () => {
  test('demonstrate workspace functionality', async ({ page }) => {
    // Go to the home page
    await page.goto('http://localhost:3001');
    
    // Take screenshot of login page
    await page.screenshot({ 
      path: 'screenshots/01-login-page.png',
      fullPage: true 
    });
    
    // Mock successful authentication
    await page.evaluate(() => {
      // Set auth cookies and user data
      document.cookie = 'hexabase_access_token=test-token; path=/';
      document.cookie = 'hexabase_refresh_token=test-refresh; path=/';
      localStorage.setItem('hexabase_user', JSON.stringify({
        id: 'demo-user',
        email: 'demo@hexabase.com',
        name: 'Demo User',
        provider: 'google'
      }));
    });
    
    // Mock organizations API
    await page.route('**/api/v1/organizations/', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          organizations: [{
            id: 'org-123',
            name: 'Hexabase Demo Organization',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: new Date().toISOString()
          }],
          total: 1
        })
      });
    });
    
    // Navigate to dashboard
    await page.goto('http://localhost:3001/dashboard');
    await page.waitForTimeout(1000);
    
    // Take screenshot of organization list
    await page.screenshot({ 
      path: 'screenshots/02-organization-list.png',
      fullPage: true 
    });
    
    // Mock organization detail
    await page.route('**/api/v1/organizations/org-123', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'org-123',
          name: 'Hexabase Demo Organization',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: new Date().toISOString()
        })
      });
    });
    
    // Mock workspaces for the organization
    await page.route('**/api/v1/organizations/org-123/workspaces/', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          workspaces: [
            {
              id: 'ws-prod',
              name: 'Production Environment',
              plan_id: 'plan-pro',
              vcluster_status: 'RUNNING',
              vcluster_instance_name: 'vcluster-prod-001',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: new Date().toISOString()
            },
            {
              id: 'ws-dev',
              name: 'Development Environment',
              plan_id: 'plan-starter',
              vcluster_status: 'STOPPED',
              vcluster_instance_name: 'vcluster-dev-001',
              created_at: '2024-01-02T00:00:00Z',
              updated_at: new Date().toISOString()
            },
            {
              id: 'ws-staging',
              name: 'Staging Environment',
              plan_id: 'plan-pro',
              vcluster_status: 'RUNNING',
              vcluster_instance_name: 'vcluster-staging-001',
              created_at: '2024-01-03T00:00:00Z',
              updated_at: new Date().toISOString()
            }
          ],
          total: 3
        })
      });
    });
    
    // Mock empty projects
    await page.route('**/api/v1/organizations/org-123/projects/', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ projects: [], total: 0 })
      });
    });
    
    // Click on the organization
    await page.click('[data-testid="open-organization-org-123"]');
    await page.waitForTimeout(1000);
    
    // Take screenshot of organization dashboard with workspaces
    await page.screenshot({ 
      path: 'screenshots/03-organization-dashboard.png',
      fullPage: true 
    });
    
    // If workspaces tab exists, click it
    const workspacesTab = page.locator('button:has-text("Workspaces")');
    if (await workspacesTab.isVisible()) {
      await workspacesTab.click();
      await page.waitForTimeout(500);
      
      await page.screenshot({ 
        path: 'screenshots/04-workspaces-tab.png',
        fullPage: true 
      });
    }
    
    // Mock workspace detail and health
    await page.route('**/api/v1/organizations/org-123/workspaces/ws-prod', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'ws-prod',
          name: 'Production Environment',
          plan_id: 'plan-pro',
          vcluster_status: 'RUNNING',
          vcluster_instance_name: 'vcluster-prod-001',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: new Date().toISOString()
        })
      });
    });
    
    await page.route('**/api/v1/organizations/org-123/workspaces/ws-prod/vcluster/health', async route => {
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
            'cpu': '35.7%',
            'memory': '52.3%',
            'nodes': '3',
            'pods': '47'
          },
          last_checked: new Date().toISOString()
        })
      });
    });
    
    // Try to find and click on a workspace card
    const workspaceCard = page.locator('[data-testid="workspace-card-ws-prod"]').first();
    if (await workspaceCard.isVisible()) {
      await workspaceCard.click();
      await page.waitForTimeout(1000);
      
      await page.screenshot({ 
        path: 'screenshots/05-workspace-detail.png',
        fullPage: true 
      });
    }
    
    // Go back to workspaces
    const backButton = page.locator('button:has-text("Back")').first();
    if (await backButton.isVisible()) {
      await backButton.click();
      await page.waitForTimeout(500);
    }
    
    // Mock plans for create dialog
    await page.route('**/api/v1/plans', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          plans: [
            {
              id: 'plan-starter',
              name: 'Starter',
              description: 'Perfect for small projects and development',
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
              description: 'For production workloads with high availability',
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
              description: 'Unlimited resources with dedicated support',
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
    
    // Find and click create workspace button
    const createButton = page.locator('button:has-text("Create Workspace")').first();
    if (await createButton.isVisible()) {
      await createButton.click();
      await page.waitForTimeout(1000);
      
      // Fill in workspace name
      const nameInput = page.locator('input[name="name"]');
      if (await nameInput.isVisible()) {
        await nameInput.fill('New Demo Workspace');
      }
      
      await page.screenshot({ 
        path: 'screenshots/06-create-workspace-dialog.png',
        fullPage: true 
      });
    }
  });
});