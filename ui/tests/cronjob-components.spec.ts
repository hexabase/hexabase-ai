import { test, expect } from '@playwright/test';

test.describe('CronJob Management', () => {
  const orgId = 'org-123';
  const workspaceId = 'ws-123';
  const projectId = 'proj-123';
  const baseUrl = `/dashboard/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}`;

  test.beforeEach(async ({ page }) => {
    // Mock authentication
    await page.goto('/login');
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', 'testpassword');
    await page.click('button[type="submit"]');
    
    // Navigate to project page
    await page.goto(baseUrl);
    await page.waitForLoadState('networkidle');
  });

  test.describe('CronJob Creation', () => {
    test('should open create cronjob dialog', async ({ page }) => {
      // Click create application button
      await page.click('button:has-text("Create Application")');
      
      // Select CronJob type
      await page.click('[data-application-type="cronjob"]');
      
      // Verify cronjob-specific fields are shown
      await expect(page.locator('label:has-text("Schedule (Cron Expression)")')).toBeVisible();
      await expect(page.locator('label:has-text("Command")')).toBeVisible();
      await expect(page.locator('label:has-text("Arguments")')).toBeVisible();
    });

    test('should create cronjob from template', async ({ page }) => {
      // Mock API responses
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications?type=stateless`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            applications: [
              {
                id: 'app-template-1',
                name: 'backup-tool',
                type: 'stateless',
                source_type: 'image',
                source_image: 'backup-tool:latest',
                status: 'running'
              }
            ],
            total: 1
          })
        });
      });

      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}/applications`, async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              id: 'app-cronjob-1',
              name: 'backup-cronjob',
              type: 'cronjob',
              status: 'active',
              cron_schedule: '0 2 * * *',
              template_app_id: 'app-template-1'
            })
          });
        }
      });

      // Open create dialog
      await page.click('button:has-text("Create Application")');
      await page.click('[data-application-type="cronjob"]');
      
      // Select template
      await page.click('input[name="use_template"]');
      await expect(page.locator('select[name="template_app_id"]')).toBeVisible();
      await page.selectOption('select[name="template_app_id"]', 'app-template-1');
      
      // Fill cronjob details
      await page.fill('input[name="name"]', 'backup-cronjob');
      await page.fill('input[name="cron_schedule"]', '0 2 * * *');
      await page.fill('input[name="command"]', '/bin/backup.sh');
      await page.fill('input[name="args"]', '--compress --incremental');
      
      // Submit form
      await page.click('button:has-text("Create CronJob")');
      
      // Verify success
      await expect(page.locator('text=CronJob created successfully')).toBeVisible();
    });

    test('should validate cron expression', async ({ page }) => {
      // Open create dialog
      await page.click('button:has-text("Create Application")');
      await page.click('[data-application-type="cronjob"]');
      
      // Enter invalid cron expression
      await page.fill('input[name="cron_schedule"]', 'invalid cron');
      await page.click('button:has-text("Create CronJob")');
      
      // Verify error
      await expect(page.locator('text=Invalid cron expression')).toBeVisible();
      
      // Enter valid cron expression
      await page.fill('input[name="cron_schedule"]', '*/5 * * * *');
      await expect(page.locator('text=Every 5 minutes')).toBeVisible();
    });
  });

  test.describe('CronJob List View', () => {
    test('should display cronjobs list', async ({ page }) => {
      // Mock API response
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}/applications?type=cronjob`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            applications: [
              {
                id: 'app-cron-1',
                name: 'daily-backup',
                type: 'cronjob',
                status: 'active',
                cron_schedule: '0 2 * * *',
                last_execution_at: '2024-01-09T02:00:00Z',
                next_execution_at: '2024-01-10T02:00:00Z'
              },
              {
                id: 'app-cron-2',
                name: 'hourly-sync',
                type: 'cronjob',
                status: 'suspended',
                cron_schedule: '0 * * * *',
                last_execution_at: '2024-01-09T12:00:00Z',
                next_execution_at: null
              }
            ],
            total: 2
          })
        });
      });

      // Navigate to cronjobs tab
      await page.click('a[href*="cronjobs"]');
      
      // Verify cronjobs are displayed
      await expect(page.locator('text=daily-backup')).toBeVisible();
      await expect(page.locator('text=0 2 * * *')).toBeVisible();
      await expect(page.locator('text=Daily at 2:00 AM')).toBeVisible();
      await expect(page.locator('text=Active')).toBeVisible();
      
      await expect(page.locator('text=hourly-sync')).toBeVisible();
      await expect(page.locator('text=0 * * * *')).toBeVisible();
      await expect(page.locator('text=Suspended')).toBeVisible();
    });

    test('should toggle cronjob status', async ({ page }) => {
      const cronJobId = 'app-cron-1';
      
      // Mock API responses
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${cronJobId}`, async route => {
        if (route.request().method() === 'PATCH') {
          const body = await route.request().postDataJSON();
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              id: cronJobId,
              name: 'daily-backup',
              type: 'cronjob',
              status: body.status
            })
          });
        }
      });

      // Navigate to cronjob details
      await page.goto(`${baseUrl}/applications/${cronJobId}`);
      
      // Toggle status
      await page.click('[data-testid="status-toggle"]');
      
      // Verify status changed
      await expect(page.locator('text=CronJob suspended')).toBeVisible();
      
      // Toggle back
      await page.click('[data-testid="status-toggle"]');
      await expect(page.locator('text=CronJob activated')).toBeVisible();
    });
  });

  test.describe('CronJob Execution History', () => {
    test('should display execution history', async ({ page }) => {
      const cronJobId = 'app-cron-1';
      
      // Mock API response
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${cronJobId}/executions`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            executions: [
              {
                id: 'cje-1',
                application_id: cronJobId,
                job_name: 'daily-backup-28474950',
                started_at: '2024-01-09T02:00:00Z',
                completed_at: '2024-01-09T02:15:00Z',
                status: 'succeeded',
                exit_code: 0,
                logs: 'Backup completed successfully\nFiles: 1024\nSize: 2.5GB'
              },
              {
                id: 'cje-2',
                application_id: cronJobId,
                job_name: 'daily-backup-28474951',
                started_at: '2024-01-08T02:00:00Z',
                completed_at: '2024-01-08T02:05:00Z',
                status: 'failed',
                exit_code: 1,
                logs: 'Error: Storage quota exceeded'
              }
            ],
            total: 2,
            page: 1,
            page_size: 10
          })
        });
      });

      // Navigate to cronjob details
      await page.goto(`${baseUrl}/applications/${cronJobId}`);
      
      // Click on executions tab
      await page.click('button:has-text("Execution History")');
      
      // Verify executions are displayed
      await expect(page.locator('text=daily-backup-28474950')).toBeVisible();
      await expect(page.locator('text=Succeeded')).toBeVisible();
      await expect(page.locator('text=15 minutes')).toBeVisible();
      
      await expect(page.locator('text=daily-backup-28474951')).toBeVisible();
      await expect(page.locator('text=Failed')).toBeVisible();
      await expect(page.locator('text=Exit code: 1')).toBeVisible();
    });

    test('should view execution logs', async ({ page }) => {
      const cronJobId = 'app-cron-1';
      const executionId = 'cje-1';
      
      // Navigate to cronjob details
      await page.goto(`${baseUrl}/applications/${cronJobId}`);
      
      // Click on executions tab
      await page.click('button:has-text("Execution History")');
      
      // Click on view logs button
      await page.click(`[data-execution-id="${executionId}"] button:has-text("View Logs")`);
      
      // Verify logs dialog is shown
      await expect(page.locator('[role="dialog"]')).toBeVisible();
      await expect(page.locator('text=Execution Logs')).toBeVisible();
      await expect(page.locator('text=Backup completed successfully')).toBeVisible();
      await expect(page.locator('text=Files: 1024')).toBeVisible();
    });

    test('should manually trigger cronjob', async ({ page }) => {
      const cronJobId = 'app-cron-1';
      
      // Mock API response
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${cronJobId}/trigger`, async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 202,
            contentType: 'application/json',
            body: JSON.stringify({
              execution_id: 'cje-new',
              job_name: 'daily-backup-manual',
              status: 'running',
              message: 'CronJob triggered successfully'
            })
          });
        }
      });

      // Navigate to cronjob details
      await page.goto(`${baseUrl}/applications/${cronJobId}`);
      
      // Click trigger button
      await page.click('button:has-text("Trigger Now")');
      
      // Confirm trigger
      await page.click('button:has-text("Yes, trigger")');
      
      // Verify success message
      await expect(page.locator('text=CronJob triggered successfully')).toBeVisible();
      await expect(page.locator('text=Running')).toBeVisible();
    });
  });

  test.describe('CronJob Schedule Management', () => {
    test('should update cronjob schedule', async ({ page }) => {
      const cronJobId = 'app-cron-1';
      
      // Mock API response
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${cronJobId}/schedule`, async route => {
        if (route.request().method() === 'PUT') {
          const body = await route.request().postDataJSON();
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              id: cronJobId,
              cron_schedule: body.schedule,
              next_execution_at: '2024-01-10T00:00:00Z'
            })
          });
        }
      });

      // Navigate to cronjob details
      await page.goto(`${baseUrl}/applications/${cronJobId}`);
      
      // Click edit schedule button
      await page.click('button:has-text("Edit Schedule")');
      
      // Update schedule
      await page.fill('input[name="cron_schedule"]', '0 0 * * *');
      await expect(page.locator('text=Daily at midnight')).toBeVisible();
      
      // Save changes
      await page.click('button:has-text("Update Schedule")');
      
      // Verify success
      await expect(page.locator('text=Schedule updated successfully')).toBeVisible();
      await expect(page.locator('text=0 0 * * *')).toBeVisible();
    });

    test('should show schedule preview', async ({ page }) => {
      // Navigate to create cronjob
      await page.click('button:has-text("Create Application")');
      await page.click('[data-application-type="cronjob"]');
      
      // Test various cron expressions
      const schedules = [
        { expr: '*/5 * * * *', preview: 'Every 5 minutes' },
        { expr: '0 * * * *', preview: 'Every hour' },
        { expr: '0 0 * * *', preview: 'Daily at midnight' },
        { expr: '0 0 * * 0', preview: 'Weekly on Sunday' },
        { expr: '0 0 1 * *', preview: 'Monthly on the 1st' },
      ];

      for (const schedule of schedules) {
        await page.fill('input[name="cron_schedule"]', schedule.expr);
        await expect(page.locator(`text=${schedule.preview}`)).toBeVisible();
      }
    });
  });

  test.describe('CronJob Template Management', () => {
    test('should create cronjob template', async ({ page }) => {
      const cronJobId = 'app-cron-1';
      
      // Mock API response
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${cronJobId}/save-as-template`, async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              template_id: 'app-template-new',
              message: 'Template created successfully'
            })
          });
        }
      });

      // Navigate to cronjob details
      await page.goto(`${baseUrl}/applications/${cronJobId}`);
      
      // Click save as template
      await page.click('button:has-text("Save as Template")');
      
      // Fill template details
      await page.fill('input[name="template_name"]', 'Daily Backup Template');
      await page.fill('textarea[name="template_description"]', 'Template for daily backup cronjobs');
      
      // Save template
      await page.click('button:has-text("Create Template")');
      
      // Verify success
      await expect(page.locator('text=Template created successfully')).toBeVisible();
    });

    test('should list cronjob templates', async ({ page }) => {
      // Mock API response
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications?type=cronjob&is_template=true`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            applications: [
              {
                id: 'app-template-1',
                name: 'backup-template',
                type: 'cronjob',
                cron_schedule: '0 2 * * *',
                description: 'Daily backup template'
              },
              {
                id: 'app-template-2',
                name: 'sync-template',
                type: 'cronjob',
                cron_schedule: '*/30 * * * *',
                description: 'Data sync template'
              }
            ],
            total: 2
          })
        });
      });

      // Open create dialog
      await page.click('button:has-text("Create Application")');
      await page.click('[data-application-type="cronjob"]');
      await page.click('input[name="use_template"]');
      
      // Verify templates are listed
      await expect(page.locator('text=backup-template')).toBeVisible();
      await expect(page.locator('text=Daily backup template')).toBeVisible();
      await expect(page.locator('text=sync-template')).toBeVisible();
      await expect(page.locator('text=Data sync template')).toBeVisible();
    });
  });

  test.describe('CronJob Integration', () => {
    test('should link cronjob with backup policy', async ({ page }) => {
      const cronJobId = 'app-cron-backup';
      
      // Mock API responses
      await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-policies`, async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            policies: [
              {
                id: 'bp-1',
                application_id: 'app-db-1',
                schedule: '0 2 * * *',
                retention_days: 30
              }
            ],
            total: 1
          })
        });
      });

      // Navigate to cronjob creation with backup integration
      await page.click('button:has-text("Create Application")');
      await page.click('[data-application-type="cronjob"]');
      
      // Enable backup integration
      await page.click('input[name="enable_backup_integration"]');
      
      // Select backup policy
      await expect(page.locator('select[name="backup_policy_id"]')).toBeVisible();
      await page.selectOption('select[name="backup_policy_id"]', 'bp-1');
      
      // Verify schedule is synced
      await expect(page.locator('input[name="cron_schedule"][value="0 2 * * *"]')).toBeVisible();
      await expect(page.locator('text=Schedule synced with backup policy')).toBeVisible();
    });
  });
});