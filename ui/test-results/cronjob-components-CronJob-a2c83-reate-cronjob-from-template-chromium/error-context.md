# Test info

- Name: CronJob Management >> CronJob Creation >> should create cronjob from template
- Location: /Users/hi/src/hexabase-ai/ui/tests/cronjob-components.spec.ts:35:9

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:3000/login
Call log:
  - navigating to "http://localhost:3000/login", waiting until "load"

    at /Users/hi/src/hexabase-ai/ui/tests/cronjob-components.spec.ts:11:16
```

# Test source

```ts
   1 | import { test, expect } from '@playwright/test';
   2 |
   3 | test.describe('CronJob Management', () => {
   4 |   const orgId = 'org-123';
   5 |   const workspaceId = 'ws-123';
   6 |   const projectId = 'proj-123';
   7 |   const baseUrl = `/dashboard/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}`;
   8 |
   9 |   test.beforeEach(async ({ page }) => {
   10 |     // Mock authentication
>  11 |     await page.goto('/login');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:3000/login
   12 |     await page.fill('input[name="email"]', 'test@example.com');
   13 |     await page.fill('input[name="password"]', 'testpassword');
   14 |     await page.click('button[type="submit"]');
   15 |     
   16 |     // Navigate to project page
   17 |     await page.goto(baseUrl);
   18 |     await page.waitForLoadState('networkidle');
   19 |   });
   20 |
   21 |   test.describe('CronJob Creation', () => {
   22 |     test('should open create cronjob dialog', async ({ page }) => {
   23 |       // Click create application button
   24 |       await page.click('button:has-text("Create Application")');
   25 |       
   26 |       // Select CronJob type
   27 |       await page.click('[data-application-type="cronjob"]');
   28 |       
   29 |       // Verify cronjob-specific fields are shown
   30 |       await expect(page.locator('label:has-text("Schedule (Cron Expression)")')).toBeVisible();
   31 |       await expect(page.locator('label:has-text("Command")')).toBeVisible();
   32 |       await expect(page.locator('label:has-text("Arguments")')).toBeVisible();
   33 |     });
   34 |
   35 |     test('should create cronjob from template', async ({ page }) => {
   36 |       // Mock API responses
   37 |       await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications?type=stateless`, async route => {
   38 |         await route.fulfill({
   39 |           status: 200,
   40 |           contentType: 'application/json',
   41 |           body: JSON.stringify({
   42 |             applications: [
   43 |               {
   44 |                 id: 'app-template-1',
   45 |                 name: 'backup-tool',
   46 |                 type: 'stateless',
   47 |                 source_type: 'image',
   48 |                 source_image: 'backup-tool:latest',
   49 |                 status: 'running'
   50 |               }
   51 |             ],
   52 |             total: 1
   53 |           })
   54 |         });
   55 |       });
   56 |
   57 |       await page.route(`**/api/v1/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}/applications`, async route => {
   58 |         if (route.request().method() === 'POST') {
   59 |           await route.fulfill({
   60 |             status: 201,
   61 |             contentType: 'application/json',
   62 |             body: JSON.stringify({
   63 |               id: 'app-cronjob-1',
   64 |               name: 'backup-cronjob',
   65 |               type: 'cronjob',
   66 |               status: 'active',
   67 |               cron_schedule: '0 2 * * *',
   68 |               template_app_id: 'app-template-1'
   69 |             })
   70 |           });
   71 |         }
   72 |       });
   73 |
   74 |       // Open create dialog
   75 |       await page.click('button:has-text("Create Application")');
   76 |       await page.click('[data-application-type="cronjob"]');
   77 |       
   78 |       // Select template
   79 |       await page.click('input[name="use_template"]');
   80 |       await expect(page.locator('select[name="template_app_id"]')).toBeVisible();
   81 |       await page.selectOption('select[name="template_app_id"]', 'app-template-1');
   82 |       
   83 |       // Fill cronjob details
   84 |       await page.fill('input[name="name"]', 'backup-cronjob');
   85 |       await page.fill('input[name="cron_schedule"]', '0 2 * * *');
   86 |       await page.fill('input[name="command"]', '/bin/backup.sh');
   87 |       await page.fill('input[name="args"]', '--compress --incremental');
   88 |       
   89 |       // Submit form
   90 |       await page.click('button:has-text("Create CronJob")');
   91 |       
   92 |       // Verify success
   93 |       await expect(page.locator('text=CronJob created successfully')).toBeVisible();
   94 |     });
   95 |
   96 |     test('should validate cron expression', async ({ page }) => {
   97 |       // Open create dialog
   98 |       await page.click('button:has-text("Create Application")');
   99 |       await page.click('[data-application-type="cronjob"]');
  100 |       
  101 |       // Enter invalid cron expression
  102 |       await page.fill('input[name="cron_schedule"]', 'invalid cron');
  103 |       await page.click('button:has-text("Create CronJob")');
  104 |       
  105 |       // Verify error
  106 |       await expect(page.locator('text=Invalid cron expression')).toBeVisible();
  107 |       
  108 |       // Enter valid cron expression
  109 |       await page.fill('input[name="cron_schedule"]', '*/5 * * * *');
  110 |       await expect(page.locator('text=Every 5 minutes')).toBeVisible();
  111 |     });
```