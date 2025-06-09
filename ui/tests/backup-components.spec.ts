import { test, expect } from '@playwright/test';

test.describe('Backup Management', () => {
  const orgId = 'org-123';
  const workspaceId = 'ws-123';
  const baseUrl = `/dashboard/organizations/${orgId}/workspaces/${workspaceId}`;

  test.beforeEach(async ({ page }) => {
    // Mock authentication
    await page.goto('/login');
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', 'testpassword');
    await page.click('button[type="submit"]');
    
    // Navigate to workspace page
    await page.goto(baseUrl);
    await page.waitForLoadState('networkidle');
  });

  test.describe('Backup Storage Management', () => {
    test('should display backup storage section for dedicated plan workspaces', async ({ page }) => {
      // Mock API response for workspace with dedicated plan
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            id: workspaceId,
            name: 'Test Workspace',
            plan: 'dedicated',
            status: 'active'
          })
        });
      });

      await page.reload();
      
      // Check if backup section is visible
      await expect(page.locator('text=Backup Storage')).toBeVisible();
      await expect(page.locator('button:has-text("Create Backup Storage")')).toBeVisible();
    });

    test('should not display backup storage for shared plan workspaces', async ({ page }) => {
      // Mock API response for workspace with shared plan
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            id: workspaceId,
            name: 'Test Workspace',
            plan: 'shared',
            status: 'active'
          })
        });
      });

      await page.reload();
      
      // Check that backup section is not visible
      await expect(page.locator('text=Backup Storage')).not.toBeVisible();
    });

    test('should create new backup storage', async ({ page }) => {
      // Mock successful backup storage creation
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages`, async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              id: 'bs-123',
              name: 'production-backups',
              type: 'proxmox',
              capacity_gb: 100,
              status: 'provisioning'
            })
          });
        }
      });

      // Click create button
      await page.click('button:has-text("Create Backup Storage")');
      
      // Fill form
      await page.fill('input[name="name"]', 'production-backups');
      await page.selectOption('select[name="type"]', 'proxmox');
      await page.fill('input[name="capacity_gb"]', '100');
      
      // Submit form
      await page.click('button:has-text("Create")');
      
      // Verify success message
      await expect(page.locator('text=Backup storage created successfully')).toBeVisible();
    });

    test('should list existing backup storages', async ({ page }) => {
      // Mock API response for backup storages list
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages`, async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([
              {
                id: 'bs-1',
                name: 'primary-backup',
                type: 'nfs',
                capacity_gb: 500,
                used_gb: 150,
                status: 'active'
              },
              {
                id: 'bs-2',
                name: 'secondary-backup',
                type: 'proxmox',
                capacity_gb: 1000,
                used_gb: 200,
                status: 'active'
              }
            ])
          });
        }
      });

      await page.reload();
      
      // Verify backup storages are displayed
      await expect(page.locator('text=primary-backup')).toBeVisible();
      await expect(page.locator('text=NFS')).toBeVisible();
      await expect(page.locator('text=150 GB / 500 GB')).toBeVisible();
      
      await expect(page.locator('text=secondary-backup')).toBeVisible();
      await expect(page.locator('text=Proxmox')).toBeVisible();
      await expect(page.locator('text=200 GB / 1000 GB')).toBeVisible();
    });
  });

  test.describe('Backup Policy Management', () => {
    test('should create backup policy for application', async ({ page }) => {
      const appId = 'app-123';
      
      // Navigate to application page
      await page.goto(`${baseUrl}/applications/${appId}`);
      
      // Mock API response for backup policy creation
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${appId}/backup-policies`, async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              id: 'bp-123',
              application_id: appId,
              storage_id: 'bs-1',
              schedule: '0 2 * * *',
              retention_days: 30,
              enabled: true
            })
          });
        }
      });

      // Click create backup policy button
      await page.click('button:has-text("Create Backup Policy")');
      
      // Fill policy form
      await page.selectOption('select[name="storage_id"]', 'bs-1');
      await page.fill('input[name="schedule"]', '0 2 * * *');
      await page.fill('input[name="retention_days"]', '30');
      await page.check('input[name="include_volumes"]');
      await page.check('input[name="include_database"]');
      await page.check('input[name="compression_enabled"]');
      await page.check('input[name="encryption_enabled"]');
      
      // Submit form
      await page.click('button:has-text("Create Policy")');
      
      // Verify success
      await expect(page.locator('text=Backup policy created successfully')).toBeVisible();
    });

    test('should display backup execution history', async ({ page }) => {
      const appId = 'app-123';
      const policyId = 'bp-123';
      
      // Mock API response for backup executions
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-policies/${policyId}/executions`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            executions: [
              {
                id: 'be-1',
                policy_id: policyId,
                status: 'succeeded',
                started_at: '2024-01-09T02:00:00Z',
                completed_at: '2024-01-09T02:15:00Z',
                size_bytes: 1073741824,
                compressed_size_bytes: 536870912
              },
              {
                id: 'be-2',
                policy_id: policyId,
                status: 'failed',
                started_at: '2024-01-08T02:00:00Z',
                completed_at: '2024-01-08T02:05:00Z',
                error_message: 'Storage quota exceeded'
              }
            ],
            total: 2,
            page: 1,
            page_size: 10
          })
        });
      });

      // Navigate to backup history
      await page.goto(`${baseUrl}/applications/${appId}/backups`);
      
      // Verify execution history is displayed
      await expect(page.locator('text=Succeeded')).toBeVisible();
      await expect(page.locator('text=1 GB → 512 MB')).toBeVisible();
      await expect(page.locator('text=15 minutes')).toBeVisible();
      
      await expect(page.locator('text=Failed')).toBeVisible();
      await expect(page.locator('text=Storage quota exceeded')).toBeVisible();
    });

    test('should trigger manual backup', async ({ page }) => {
      const appId = 'app-123';
      const policyId = 'bp-123';
      
      // Mock API response for manual backup trigger
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-policies/${policyId}/execute`, async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 202,
            contentType: 'application/json',
            body: JSON.stringify({
              execution_id: 'be-new',
              status: 'running',
              message: 'Backup started'
            })
          });
        }
      });

      // Navigate to backup page
      await page.goto(`${baseUrl}/applications/${appId}/backups`);
      
      // Click manual backup button
      await page.click('button:has-text("Backup Now")');
      
      // Confirm backup
      await page.click('button:has-text("Start Backup")');
      
      // Verify backup started
      await expect(page.locator('text=Backup started successfully')).toBeVisible();
      await expect(page.locator('text=Running')).toBeVisible();
    });
  });

  test.describe('Backup Restore Operations', () => {
    test('should restore from backup', async ({ page }) => {
      const appId = 'app-123';
      const executionId = 'be-1';
      
      // Mock API response for restore operation
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${appId}/restore`, async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 202,
            contentType: 'application/json',
            body: JSON.stringify({
              restore_id: 'br-123',
              status: 'running',
              message: 'Restore started'
            })
          });
        }
      });

      // Navigate to backup history
      await page.goto(`${baseUrl}/applications/${appId}/backups`);
      
      // Click restore button on a backup
      await page.click(`button[data-execution-id="${executionId}"]:has-text("Restore")`);
      
      // Configure restore options
      await page.selectOption('select[name="restore_type"]', 'in_place');
      await page.check('input[name="restore_volumes"]');
      await page.check('input[name="restore_database"]');
      
      // Confirm restore
      await page.click('button:has-text("Start Restore")');
      
      // Verify restore started
      await expect(page.locator('text=Restore started successfully')).toBeVisible();
    });

    test('should validate backup before restore', async ({ page }) => {
      const appId = 'app-123';
      const executionId = 'be-1';
      
      // Mock API response for backup validation
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-executions/${executionId}/validate`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            valid: true,
            integrity_check: 'passed',
            backup_manifest: {
              volumes: ['pvc-data', 'pvc-logs'],
              databases: ['postgres'],
              config_maps: ['app-config']
            }
          })
        });
      });

      // Navigate to backup details
      await page.goto(`${baseUrl}/applications/${appId}/backups/${executionId}`);
      
      // Click validate button
      await page.click('button:has-text("Validate Backup")');
      
      // Verify validation results
      await expect(page.locator('text=Backup is valid')).toBeVisible();
      await expect(page.locator('text=Integrity check passed')).toBeVisible();
      await expect(page.locator('text=pvc-data')).toBeVisible();
      await expect(page.locator('text=postgres')).toBeVisible();
    });
  });

  test.describe('Storage Usage Monitoring', () => {
    test('should display storage usage dashboard', async ({ page }) => {
      // Mock API response for storage usage
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages/usage`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              storage_id: 'bs-1',
              total_gb: 500,
              used_gb: 350,
              available_gb: 150,
              usage_percent: 70,
              backup_count: 45,
              oldest_backup: '2023-12-01T00:00:00Z',
              latest_backup: '2024-01-09T02:00:00Z'
            }
          ])
        });
      });

      // Navigate to backup dashboard
      await page.goto(`${baseUrl}/backups`);
      
      // Verify usage information
      await expect(page.locator('text=70% Used')).toBeVisible();
      await expect(page.locator('text=350 GB / 500 GB')).toBeVisible();
      await expect(page.locator('text=45 backups')).toBeVisible();
      await expect(page.locator('text=Oldest: Dec 1, 2023')).toBeVisible();
    });

    test('should alert on high storage usage', async ({ page }) => {
      // Mock API response with high usage
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages/usage`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              storage_id: 'bs-1',
              total_gb: 500,
              used_gb: 480,
              available_gb: 20,
              usage_percent: 96,
              backup_count: 120
            }
          ])
        });
      });

      await page.goto(`${baseUrl}/backups`);
      
      // Verify high usage warning
      await expect(page.locator('[role="alert"]:has-text("Storage usage critical")')).toBeVisible();
      await expect(page.locator('text=96% Used')).toHaveClass(/text-red-600/);
    });
  });
});