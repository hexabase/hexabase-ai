# Test info

- Name: Workspace Management Demo >> demonstrate workspace functionality
- Location: /Users/hi/src/hexabase-kaas/ui/tests/workspace-demo.spec.ts:4:7

# Error details

```
Error: page.click: Test timeout of 30000ms exceeded.
Call log:
  - waiting for locator('[data-testid="open-organization-org-123"]')

    at /Users/hi/src/hexabase-kaas/ui/tests/workspace-demo.spec.ts:118:16
```

# Page snapshot

```yaml
- alert
- button "Open Next.js Dev Tools":
  - img
- button "Open issues overlay": 1 Issue
- navigation:
  - button "previous" [disabled]:
    - img "previous"
  - text: 1/1
  - button "next" [disabled]:
    - img "next"
- img
- img
- text: Next.js 15.3.3 Turbopack
- img
- dialog "Build Error":
  - text: Build Error
  - button "Copy Stack Trace":
    - img
  - button "No related documentation found" [disabled]:
    - img
  - link "Learn more about enabling Node.js inspector for server code with Chrome DevTools":
    - /url: https://nextjs.org/docs/app/building-your-application/configuring/debugging#server-side-code
    - img
  - paragraph: Ecmascript file had an error
  - img
  - text: ./src/lib/api-client.ts (1269:14)
  - button "Open in editor":
    - img
  - text: "Ecmascript file had an error 1267 | 1268 | // Task API functions > 1269 | export const taskApi = { | ^^^^^^^ 1270 | // Get task by ID 1271 | get: async (taskId: string): Promise<Task> => { 1272 | const response = await apiClient.get(`/api/v1/tasks/${taskId}`); the name `taskApi` is defined multiple times"
- contentinfo:
  - paragraph: This error occurred during the build process and can only be dismissed by fixing the error.
```

# Test source

```ts
   18 |       document.cookie = 'hexabase_refresh_token=test-refresh; path=/';
   19 |       localStorage.setItem('hexabase_user', JSON.stringify({
   20 |         id: 'demo-user',
   21 |         email: 'demo@hexabase.com',
   22 |         name: 'Demo User',
   23 |         provider: 'google'
   24 |       }));
   25 |     });
   26 |     
   27 |     // Mock organizations API
   28 |     await page.route('**/api/v1/organizations/', async route => {
   29 |       await route.fulfill({
   30 |         status: 200,
   31 |         contentType: 'application/json',
   32 |         body: JSON.stringify({
   33 |           organizations: [{
   34 |             id: 'org-123',
   35 |             name: 'Hexabase Demo Organization',
   36 |             created_at: '2024-01-01T00:00:00Z',
   37 |             updated_at: new Date().toISOString()
   38 |           }],
   39 |           total: 1
   40 |         })
   41 |       });
   42 |     });
   43 |     
   44 |     // Navigate to dashboard
   45 |     await page.goto('http://localhost:3001/dashboard');
   46 |     await page.waitForTimeout(1000);
   47 |     
   48 |     // Take screenshot of organization list
   49 |     await page.screenshot({ 
   50 |       path: 'screenshots/02-organization-list.png',
   51 |       fullPage: true 
   52 |     });
   53 |     
   54 |     // Mock organization detail
   55 |     await page.route('**/api/v1/organizations/org-123', async route => {
   56 |       await route.fulfill({
   57 |         status: 200,
   58 |         contentType: 'application/json',
   59 |         body: JSON.stringify({
   60 |           id: 'org-123',
   61 |           name: 'Hexabase Demo Organization',
   62 |           created_at: '2024-01-01T00:00:00Z',
   63 |           updated_at: new Date().toISOString()
   64 |         })
   65 |       });
   66 |     });
   67 |     
   68 |     // Mock workspaces for the organization
   69 |     await page.route('**/api/v1/organizations/org-123/workspaces/', async route => {
   70 |       await route.fulfill({
   71 |         status: 200,
   72 |         contentType: 'application/json',
   73 |         body: JSON.stringify({
   74 |           workspaces: [
   75 |             {
   76 |               id: 'ws-prod',
   77 |               name: 'Production Environment',
   78 |               plan_id: 'plan-pro',
   79 |               vcluster_status: 'RUNNING',
   80 |               vcluster_instance_name: 'vcluster-prod-001',
   81 |               created_at: '2024-01-01T00:00:00Z',
   82 |               updated_at: new Date().toISOString()
   83 |             },
   84 |             {
   85 |               id: 'ws-dev',
   86 |               name: 'Development Environment',
   87 |               plan_id: 'plan-starter',
   88 |               vcluster_status: 'STOPPED',
   89 |               vcluster_instance_name: 'vcluster-dev-001',
   90 |               created_at: '2024-01-02T00:00:00Z',
   91 |               updated_at: new Date().toISOString()
   92 |             },
   93 |             {
   94 |               id: 'ws-staging',
   95 |               name: 'Staging Environment',
   96 |               plan_id: 'plan-pro',
   97 |               vcluster_status: 'RUNNING',
   98 |               vcluster_instance_name: 'vcluster-staging-001',
   99 |               created_at: '2024-01-03T00:00:00Z',
  100 |               updated_at: new Date().toISOString()
  101 |             }
  102 |           ],
  103 |           total: 3
  104 |         })
  105 |       });
  106 |     });
  107 |     
  108 |     // Mock empty projects
  109 |     await page.route('**/api/v1/organizations/org-123/projects/', async route => {
  110 |       await route.fulfill({
  111 |         status: 200,
  112 |         contentType: 'application/json',
  113 |         body: JSON.stringify({ projects: [], total: 0 })
  114 |       });
  115 |     });
  116 |     
  117 |     // Click on the organization
> 118 |     await page.click('[data-testid="open-organization-org-123"]');
      |                ^ Error: page.click: Test timeout of 30000ms exceeded.
  119 |     await page.waitForTimeout(1000);
  120 |     
  121 |     // Take screenshot of organization dashboard with workspaces
  122 |     await page.screenshot({ 
  123 |       path: 'screenshots/03-organization-dashboard.png',
  124 |       fullPage: true 
  125 |     });
  126 |     
  127 |     // If workspaces tab exists, click it
  128 |     const workspacesTab = page.locator('button:has-text("Workspaces")');
  129 |     if (await workspacesTab.isVisible()) {
  130 |       await workspacesTab.click();
  131 |       await page.waitForTimeout(500);
  132 |       
  133 |       await page.screenshot({ 
  134 |         path: 'screenshots/04-workspaces-tab.png',
  135 |         fullPage: true 
  136 |       });
  137 |     }
  138 |     
  139 |     // Mock workspace detail and health
  140 |     await page.route('**/api/v1/organizations/org-123/workspaces/ws-prod', async route => {
  141 |       await route.fulfill({
  142 |         status: 200,
  143 |         contentType: 'application/json',
  144 |         body: JSON.stringify({
  145 |           id: 'ws-prod',
  146 |           name: 'Production Environment',
  147 |           plan_id: 'plan-pro',
  148 |           vcluster_status: 'RUNNING',
  149 |           vcluster_instance_name: 'vcluster-prod-001',
  150 |           created_at: '2024-01-01T00:00:00Z',
  151 |           updated_at: new Date().toISOString()
  152 |         })
  153 |       });
  154 |     });
  155 |     
  156 |     await page.route('**/api/v1/organizations/org-123/workspaces/ws-prod/vcluster/health', async route => {
  157 |       await route.fulfill({
  158 |         status: 200,
  159 |         contentType: 'application/json',
  160 |         body: JSON.stringify({
  161 |           healthy: true,
  162 |           components: {
  163 |             'api-server': 'healthy',
  164 |             'etcd': 'healthy',
  165 |             'scheduler': 'healthy',
  166 |             'controller-manager': 'healthy'
  167 |           },
  168 |           resource_usage: {
  169 |             'cpu': '35.7%',
  170 |             'memory': '52.3%',
  171 |             'nodes': '3',
  172 |             'pods': '47'
  173 |           },
  174 |           last_checked: new Date().toISOString()
  175 |         })
  176 |       });
  177 |     });
  178 |     
  179 |     // Try to find and click on a workspace card
  180 |     const workspaceCard = page.locator('[data-testid="workspace-card-ws-prod"]').first();
  181 |     if (await workspaceCard.isVisible()) {
  182 |       await workspaceCard.click();
  183 |       await page.waitForTimeout(1000);
  184 |       
  185 |       await page.screenshot({ 
  186 |         path: 'screenshots/05-workspace-detail.png',
  187 |         fullPage: true 
  188 |       });
  189 |     }
  190 |     
  191 |     // Go back to workspaces
  192 |     const backButton = page.locator('button:has-text("Back")').first();
  193 |     if (await backButton.isVisible()) {
  194 |       await backButton.click();
  195 |       await page.waitForTimeout(500);
  196 |     }
  197 |     
  198 |     // Mock plans for create dialog
  199 |     await page.route('**/api/v1/plans', async route => {
  200 |       await route.fulfill({
  201 |         status: 200,
  202 |         contentType: 'application/json',
  203 |         body: JSON.stringify({
  204 |           plans: [
  205 |             {
  206 |               id: 'plan-starter',
  207 |               name: 'Starter',
  208 |               description: 'Perfect for small projects and development',
  209 |               price: 0,
  210 |               currency: 'USD',
  211 |               resource_limits: JSON.stringify({
  212 |                 cpu: '2 cores',
  213 |                 memory: '4 GB',
  214 |                 storage: '10 GB'
  215 |               })
  216 |             },
  217 |             {
  218 |               id: 'plan-pro',
```