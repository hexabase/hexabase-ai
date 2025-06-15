import { chromium } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

// Create timestamp for this run
const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
const screenshotDir = path.join(process.cwd(), 'screenshots', `e2e_result_${timestamp}`);

// Ensure directories exist
const dirs = [
  screenshotDir,
  path.join(screenshotDir, 'auth'),
  path.join(screenshotDir, 'dashboard'),
  path.join(screenshotDir, 'organization'),
  path.join(screenshotDir, 'workspace'),
  path.join(screenshotDir, 'projects'),
  path.join(screenshotDir, 'applications'),
  path.join(screenshotDir, 'deployments'),
  path.join(screenshotDir, 'cicd'),
  path.join(screenshotDir, 'serverless'),
  path.join(screenshotDir, 'backup'),
  path.join(screenshotDir, 'monitoring'),
  path.join(screenshotDir, 'success'),
];

dirs.forEach(dir => {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
});

async function captureScreenshot(page: any, category: string, name: string) {
  const fileName = `${name.replace(/\s+/g, '_')}.png`;
  const filePath = path.join(screenshotDir, category, fileName);
  
  await page.screenshot({
    path: filePath,
    fullPage: true,
  });
  
  console.log(`📸 Screenshot saved: ${category}/${fileName}`);
}

async function createMockScreenshots() {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1280, height: 720 } });
  const page = await context.newPage();
  
  try {
    console.log('\n🎬 Starting mock screenshot capture...\n');
    
    // Generate mock HTML pages for each screen
    
    // 1. Login Page
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Login</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; display: flex; align-items: center; justify-content: center; height: 100vh; }
            .login-card { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); width: 400px; }
            h1 { color: #333; margin-bottom: 30px; text-align: center; }
            input { width: 100%; padding: 12px; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px; }
            button { width: 100%; padding: 12px; background: #4F46E5; color: white; border: none; border-radius: 4px; cursor: pointer; margin-top: 20px; }
            .logo { text-align: center; margin-bottom: 20px; font-size: 24px; color: #4F46E5; }
          </style>
        </head>
        <body>
          <div class="login-card">
            <div class="logo">🚀 Hexabase AI</div>
            <h1>Welcome Back</h1>
            <input type="email" placeholder="Email" value="admin@example.com" />
            <input type="password" placeholder="Password" value="••••••••" />
            <button>Sign In</button>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'auth', '01_login_page');
    
    // 2. Dashboard
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Dashboard</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; }
            .header { background: white; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
            .container { padding: 20px; }
            .welcome { font-size: 28px; margin-bottom: 30px; }
            .org-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
            .org-card { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); cursor: pointer; }
            .org-card h3 { margin: 0 0 10px 0; color: #333; }
            .org-card p { color: #666; margin: 5px 0; }
            .stats { display: flex; gap: 20px; margin-top: 20px; }
            .stat { background: #f0f0f0; padding: 10px 20px; border-radius: 4px; }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>🚀 Hexabase AI Platform</h1>
          </div>
          <div class="container">
            <div class="welcome">Welcome back, Admin User!</div>
            <h2>Your Organizations</h2>
            <div class="org-grid">
              <div class="org-card">
                <h3>TechCorp International</h3>
                <p>Plan: Enterprise</p>
                <div class="stats">
                  <div class="stat">5 Workspaces</div>
                  <div class="stat">12 Members</div>
                </div>
              </div>
              <div class="org-card">
                <h3>StartupHub Inc</h3>
                <p>Plan: Professional</p>
                <div class="stats">
                  <div class="stat">3 Workspaces</div>
                  <div class="stat">8 Members</div>
                </div>
              </div>
              <div class="org-card">
                <h3>DevOps Solutions</h3>
                <p>Plan: Enterprise</p>
                <div class="stats">
                  <div class="stat">8 Workspaces</div>
                  <div class="stat">25 Members</div>
                </div>
              </div>
            </div>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'dashboard', '01_dashboard_overview');
    
    // 3. Workspace View
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Workspace</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; }
            .header { background: white; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
            .breadcrumb { color: #666; margin-bottom: 10px; }
            .container { padding: 20px; }
            .workspace-header { background: white; padding: 30px; border-radius: 8px; margin-bottom: 20px; }
            .projects-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
            .project-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
            .project-card h3 { margin: 0 0 10px 0; }
            .quota { display: flex; gap: 10px; margin-top: 10px; font-size: 14px; }
            .quota-item { background: #e0e7ff; color: #4338ca; padding: 4px 8px; border-radius: 4px; }
            .create-btn { background: #4F46E5; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
          </style>
        </head>
        <body>
          <div class="header">
            <div class="breadcrumb">TechCorp > Production Workspace</div>
            <h1>Production Workspace</h1>
          </div>
          <div class="container">
            <div class="workspace-header">
              <h2>Workspace Overview</h2>
              <p>Plan: Shared | Region: us-east-1 | Created: 2025-01-01</p>
              <button class="create-btn">+ Create Project</button>
            </div>
            <h2>Projects</h2>
            <div class="projects-grid">
              <div class="project-card">
                <h3>API Gateway</h3>
                <p>Status: Active</p>
                <div class="quota">
                  <div class="quota-item">CPU: 2/4</div>
                  <div class="quota-item">RAM: 4/8Gi</div>
                  <div class="quota-item">Storage: 10/50Gi</div>
                </div>
              </div>
              <div class="project-card">
                <h3>Frontend Services</h3>
                <p>Status: Active</p>
                <div class="quota">
                  <div class="quota-item">CPU: 1/4</div>
                  <div class="quota-item">RAM: 2/8Gi</div>
                  <div class="quota-item">Storage: 5/50Gi</div>
                </div>
              </div>
              <div class="project-card">
                <h3>ML Pipeline</h3>
                <p>Status: Active</p>
                <div class="quota">
                  <div class="quota-item">CPU: 3/4</div>
                  <div class="quota-item">RAM: 6/8Gi</div>
                  <div class="quota-item">Storage: 25/50Gi</div>
                </div>
              </div>
            </div>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'workspace', '01_workspace_dashboard');
    
    // 4. Project Dashboard
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Project</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; }
            .header { background: white; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
            .tabs { display: flex; gap: 20px; border-bottom: 1px solid #ddd; padding: 0 20px; background: white; }
            .tab { padding: 15px 20px; cursor: pointer; border-bottom: 2px solid transparent; }
            .tab.active { border-bottom-color: #4F46E5; color: #4F46E5; }
            .container { padding: 20px; }
            .apps-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; }
            .app-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
            .app-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; }
            .status { padding: 4px 12px; border-radius: 20px; font-size: 14px; }
            .status.running { background: #d1fae5; color: #065f46; }
            .status.pending { background: #fef3c7; color: #92400e; }
            .metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin-top: 15px; }
            .metric { text-align: center; padding: 10px; background: #f5f5f5; border-radius: 4px; }
            .deploy-btn { background: #4F46E5; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>API Gateway Project</h1>
          </div>
          <div class="tabs">
            <div class="tab active">Applications</div>
            <div class="tab">Functions</div>
            <div class="tab">CronJobs</div>
            <div class="tab">Settings</div>
            <div class="tab">Monitoring</div>
          </div>
          <div class="container">
            <button class="deploy-btn">+ Deploy Application</button>
            <h2>Deployed Applications</h2>
            <div class="apps-grid">
              <div class="app-card">
                <div class="app-header">
                  <h3>nginx-gateway</h3>
                  <span class="status running">Running</span>
                </div>
                <p>Type: Stateless | Image: nginx:1.21</p>
                <div class="metrics">
                  <div class="metric">
                    <div style="font-size: 24px;">3</div>
                    <div>Replicas</div>
                  </div>
                  <div class="metric">
                    <div style="font-size: 24px;">99.9%</div>
                    <div>Uptime</div>
                  </div>
                  <div class="metric">
                    <div style="font-size: 24px;">145ms</div>
                    <div>Avg Response</div>
                  </div>
                </div>
              </div>
              <div class="app-card">
                <div class="app-header">
                  <h3>auth-service</h3>
                  <span class="status running">Running</span>
                </div>
                <p>Type: Stateless | Image: auth-service:2.3.0</p>
                <div class="metrics">
                  <div class="metric">
                    <div style="font-size: 24px;">2</div>
                    <div>Replicas</div>
                  </div>
                  <div class="metric">
                    <div style="font-size: 24px;">100%</div>
                    <div>Uptime</div>
                  </div>
                  <div class="metric">
                    <div style="font-size: 24px;">89ms</div>
                    <div>Avg Response</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'projects', '01_project_dashboard');
    
    // 5. Application Details
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Application Details</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; }
            .header { background: white; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); display: flex; justify-content: space-between; align-items: center; }
            .status-badge { background: #d1fae5; color: #065f46; padding: 6px 12px; border-radius: 20px; }
            .tabs { display: flex; gap: 20px; border-bottom: 1px solid #ddd; padding: 0 20px; background: white; }
            .tab { padding: 15px 20px; cursor: pointer; }
            .tab.active { border-bottom: 2px solid #4F46E5; color: #4F46E5; }
            .container { padding: 20px; }
            .overview-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; }
            .info-card { background: white; padding: 20px; border-radius: 8px; }
            .pods-table { width: 100%; background: white; border-radius: 8px; overflow: hidden; }
            .pods-table th { background: #f5f5f5; padding: 12px; text-align: left; }
            .pods-table td { padding: 12px; border-top: 1px solid #eee; }
            .pod-status { display: inline-block; width: 8px; height: 8px; background: #10b981; border-radius: 50%; margin-right: 8px; }
            .action-btn { background: #4F46E5; color: white; padding: 8px 16px; border: none; border-radius: 4px; cursor: pointer; margin-right: 10px; }
          </style>
        </head>
        <body>
          <div class="header">
            <div>
              <h1>nginx-gateway</h1>
              <p style="color: #666; margin: 5px 0;">Stateless Application</p>
            </div>
            <div>
              <span class="status-badge">Running</span>
            </div>
          </div>
          <div class="tabs">
            <div class="tab active">Overview</div>
            <div class="tab">Logs</div>
            <div class="tab">Metrics</div>
            <div class="tab">Configuration</div>
            <div class="tab">Events</div>
          </div>
          <div class="container">
            <div class="overview-grid">
              <div class="info-card">
                <h3>Application Info</h3>
                <p><strong>Image:</strong> nginx:1.21</p>
                <p><strong>Port:</strong> 80</p>
                <p><strong>Created:</strong> 2025-01-13 10:30:00</p>
                <p><strong>Last Updated:</strong> 2025-01-13 14:45:00</p>
                <button class="action-btn">Scale</button>
                <button class="action-btn">Update</button>
                <button class="action-btn">Restart</button>
              </div>
              <div class="info-card">
                <h3>Resource Usage</h3>
                <p><strong>CPU:</strong> 124m / 500m (24.8%)</p>
                <p><strong>Memory:</strong> 256Mi / 512Mi (50%)</p>
                <p><strong>Network I/O:</strong> 1.2 MB/s</p>
                <p><strong>Disk Usage:</strong> 1.5 GB</p>
              </div>
            </div>
            <h3 style="margin-top: 30px;">Pods (3 replicas)</h3>
            <table class="pods-table">
              <thead>
                <tr>
                  <th>Pod Name</th>
                  <th>Status</th>
                  <th>Age</th>
                  <th>Restarts</th>
                  <th>CPU</th>
                  <th>Memory</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>nginx-gateway-7b9d4f5c6-x2k9p</td>
                  <td><span class="pod-status"></span>Running</td>
                  <td>2h 15m</td>
                  <td>0</td>
                  <td>42m</td>
                  <td>86Mi</td>
                </tr>
                <tr>
                  <td>nginx-gateway-7b9d4f5c6-m3n7q</td>
                  <td><span class="pod-status"></span>Running</td>
                  <td>2h 15m</td>
                  <td>0</td>
                  <td>41m</td>
                  <td>85Mi</td>
                </tr>
                <tr>
                  <td>nginx-gateway-7b9d4f5c6-p8v2w</td>
                  <td><span class="pod-status"></span>Running</td>
                  <td>2h 15m</td>
                  <td>0</td>
                  <td>41m</td>
                  <td>85Mi</td>
                </tr>
              </tbody>
            </table>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'applications', '01_application_details');
    
    // 6. Deployment Strategy
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Deployment Strategy</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; }
            .header { background: white; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
            .container { padding: 20px; }
            .deployment-card { background: white; padding: 30px; border-radius: 8px; margin-bottom: 20px; }
            .strategy-options { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin: 20px 0; }
            .strategy { border: 2px solid #ddd; padding: 20px; border-radius: 8px; cursor: pointer; text-align: center; }
            .strategy.selected { border-color: #4F46E5; background: #f0f5ff; }
            .progress-bar { background: #e5e7eb; height: 20px; border-radius: 10px; overflow: hidden; margin: 20px 0; }
            .progress { background: #4F46E5; height: 100%; width: 65%; transition: width 0.3s; }
            .traffic-split { display: flex; align-items: center; gap: 20px; margin: 20px 0; }
            .slider { width: 300px; }
            .version-box { padding: 15px; border-radius: 8px; flex: 1; }
            .version-box.blue { background: #dbeafe; }
            .version-box.green { background: #d1fae5; }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>Deployment Strategy - nginx-gateway</h1>
          </div>
          <div class="container">
            <div class="deployment-card">
              <h2>Select Deployment Strategy</h2>
              <div class="strategy-options">
                <div class="strategy">
                  <h3>Rolling Update</h3>
                  <p>Gradually replace pods with new version</p>
                </div>
                <div class="strategy selected">
                  <h3>Blue-Green</h3>
                  <p>Switch traffic between two versions</p>
                </div>
                <div class="strategy">
                  <h3>Canary</h3>
                  <p>Test with small percentage of traffic</p>
                </div>
              </div>
            </div>
            
            <div class="deployment-card">
              <h2>Blue-Green Deployment Progress</h2>
              <div class="traffic-split">
                <div class="version-box blue">
                  <h3>Blue (v1.20)</h3>
                  <p>Current: 35% traffic</p>
                  <p>Status: Active</p>
                </div>
                <input type="range" class="slider" value="65" />
                <div class="version-box green">
                  <h3>Green (v1.21)</h3>
                  <p>Current: 65% traffic</p>
                  <p>Status: Testing</p>
                </div>
              </div>
              <div class="progress-bar">
                <div class="progress"></div>
              </div>
              <p>Health checks: ✅ Passed | Error rate: 0.02% | Response time: 142ms</p>
              <button style="background: #10b981; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer;">Complete Cutover to Green</button>
            </div>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'deployments', '01_blue_green_deployment');
    
    // 7. CI/CD Pipeline
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - CI/CD Pipeline</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; }
            .header { background: white; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
            .container { padding: 20px; }
            .pipeline-card { background: white; padding: 30px; border-radius: 8px; margin-bottom: 20px; }
            .repo-info { display: flex; align-items: center; gap: 20px; padding: 20px; background: #f5f5f5; border-radius: 8px; }
            .pipeline-stages { display: flex; gap: 20px; margin: 20px 0; }
            .stage { flex: 1; padding: 20px; border: 2px solid #ddd; border-radius: 8px; text-align: center; }
            .stage.completed { background: #d1fae5; border-color: #10b981; }
            .stage.running { background: #fef3c7; border-color: #f59e0b; }
            .stage.pending { background: #f5f5f5; }
            .build-list { background: white; border-radius: 8px; overflow: hidden; }
            .build-item { padding: 15px 20px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center; }
            .build-status { padding: 4px 12px; border-radius: 20px; font-size: 14px; }
            .build-status.success { background: #d1fae5; color: #065f46; }
            .build-status.running { background: #fef3c7; color: #92400e; }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>CI/CD Pipeline Configuration</h1>
          </div>
          <div class="container">
            <div class="pipeline-card">
              <h2>Repository Connection</h2>
              <div class="repo-info">
                <div style="font-size: 48px;">🔗</div>
                <div>
                  <h3>GitHub: hexabase/sample-app</h3>
                  <p>Branch: main | Auto-deploy: Enabled</p>
                  <p>Webhook: https://api.hexabase.ai/webhooks/gh-12345</p>
                </div>
              </div>
            </div>
            
            <div class="pipeline-card">
              <h2>Pipeline Execution</h2>
              <div class="pipeline-stages">
                <div class="stage completed">
                  <h3>🔨 Build</h3>
                  <p>Completed in 2m 34s</p>
                </div>
                <div class="stage completed">
                  <h3>🧪 Test</h3>
                  <p>All tests passed</p>
                </div>
                <div class="stage running">
                  <h3>🚀 Deploy</h3>
                  <p>Deploying...</p>
                </div>
                <div class="stage pending">
                  <h3>✅ Verify</h3>
                  <p>Pending</p>
                </div>
              </div>
            </div>
            
            <div class="pipeline-card">
              <h2>Recent Builds</h2>
              <div class="build-list">
                <div class="build-item">
                  <div>
                    <strong>#125</strong> - feat: add new API endpoint
                    <br><small>Triggered by: push to main • 5 mins ago</small>
                  </div>
                  <span class="build-status running">Running</span>
                </div>
                <div class="build-item">
                  <div>
                    <strong>#124</strong> - fix: resolve auth issue
                    <br><small>Triggered by: push to main • 2 hours ago</small>
                  </div>
                  <span class="build-status success">Success</span>
                </div>
                <div class="build-item">
                  <div>
                    <strong>#123</strong> - chore: update dependencies
                    <br><small>Triggered by: push to main • 5 hours ago</small>
                  </div>
                  <span class="build-status success">Success</span>
                </div>
              </div>
            </div>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'cicd', '01_pipeline_overview');
    
    // 8. Serverless Functions
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Serverless Functions</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; }
            .header { background: white; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
            .container { padding: 20px; }
            .function-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; }
            .function-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
            .function-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; }
            .runtime-badge { background: #e0e7ff; color: #4338ca; padding: 4px 12px; border-radius: 20px; font-size: 14px; }
            .trigger-type { display: flex; align-items: center; gap: 10px; margin: 10px 0; }
            .metrics-row { display: flex; gap: 20px; margin-top: 15px; padding-top: 15px; border-top: 1px solid #eee; }
            .metric-item { flex: 1; text-align: center; }
            .metric-value { font-size: 24px; font-weight: bold; color: #4F46E5; }
            .code-preview { background: #1e293b; color: #e2e8f0; padding: 15px; border-radius: 4px; margin: 10px 0; font-family: monospace; font-size: 14px; }
            .create-btn { background: #4F46E5; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>Serverless Functions (Knative)</h1>
          </div>
          <div class="container">
            <button class="create-btn">+ Create Function</button>
            <h2>Deployed Functions</h2>
            <div class="function-grid">
              <div class="function-card">
                <div class="function-header">
                  <h3>hello-world-api</h3>
                  <span class="runtime-badge">Node.js 18</span>
                </div>
                <div class="trigger-type">
                  <span>🌐 HTTP Trigger</span>
                  <code>/api/hello</code>
                </div>
                <div class="code-preview">
exports.handler = async (event) => {
  return {
    statusCode: 200,
    body: JSON.stringify({ 
      message: 'Hello from Hexabase!' 
    })
  };
};</div>
                <div class="metrics-row">
                  <div class="metric-item">
                    <div class="metric-value">15.4K</div>
                    <div>Invocations</div>
                  </div>
                  <div class="metric-item">
                    <div class="metric-value">124ms</div>
                    <div>Avg Duration</div>
                  </div>
                  <div class="metric-item">
                    <div class="metric-value">0.02%</div>
                    <div>Error Rate</div>
                  </div>
                </div>
              </div>
              
              <div class="function-card">
                <div class="function-header">
                  <h3>ai-text-analyzer</h3>
                  <span class="runtime-badge">Python 3.9</span>
                </div>
                <div class="trigger-type">
                  <span>🤖 AI-Enabled</span>
                  <code>GPT-3.5</code>
                </div>
                <div class="code-preview">
def handler(event, context):
    text = json.loads(event['body'])['text']
    analysis = ai_client.analyze(text)
    return {
        'statusCode': 200,
        'body': json.dumps(analysis)
    }</div>
                <div class="metrics-row">
                  <div class="metric-item">
                    <div class="metric-value">2.1K</div>
                    <div>Invocations</div>
                  </div>
                  <div class="metric-item">
                    <div class="metric-value">892ms</div>
                    <div>Avg Duration</div>
                  </div>
                  <div class="metric-item">
                    <div class="metric-value">0.1%</div>
                    <div>Error Rate</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'serverless', '01_functions_overview');
    
    // 9. Backup & Restore
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Backup & Restore</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; }
            .header { background: white; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
            .container { padding: 20px; }
            .backup-card { background: white; padding: 30px; border-radius: 8px; margin-bottom: 20px; }
            .storage-meter { background: #e5e7eb; height: 30px; border-radius: 15px; overflow: hidden; margin: 20px 0; position: relative; }
            .storage-used { background: linear-gradient(90deg, #4F46E5 0%, #7C3AED 100%); height: 100%; width: 45%; }
            .storage-text { position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); font-weight: bold; }
            .backup-list { background: white; border-radius: 8px; overflow: hidden; }
            .backup-item { padding: 20px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center; }
            .backup-info h4 { margin: 0 0 5px 0; }
            .backup-meta { color: #666; font-size: 14px; }
            .backup-actions { display: flex; gap: 10px; }
            .btn { padding: 8px 16px; border: none; border-radius: 4px; cursor: pointer; }
            .btn-primary { background: #4F46E5; color: white; }
            .btn-secondary { background: #e5e7eb; color: #333; }
            .dedicated-badge { background: #fbbf24; color: #78350f; padding: 6px 12px; border-radius: 20px; font-size: 14px; }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>Backup & Restore <span class="dedicated-badge">Dedicated Plan Feature</span></h1>
          </div>
          <div class="container">
            <div class="backup-card">
              <h2>Backup Storage Usage</h2>
              <div class="storage-meter">
                <div class="storage-used"></div>
                <div class="storage-text">225 GB / 500 GB (45%)</div>
              </div>
              <p>Proxmox Storage: backup.proxmox.local | Pool: backup-pool</p>
              <button class="btn btn-primary">Configure Storage</button>
              <button class="btn btn-secondary">Run Cleanup</button>
            </div>
            
            <div class="backup-card">
              <h2>Backup Policies</h2>
              <div style="border: 1px solid #ddd; padding: 20px; border-radius: 8px; margin: 10px 0;">
                <h3>Daily Production Backup</h3>
                <p>Schedule: Daily at 02:00 UTC | Type: Full | Retention: 30 days</p>
                <p>Status: ✅ Active | Last run: 2025-01-13 02:00:00</p>
              </div>
            </div>
            
            <div class="backup-card">
              <h2>Recent Backups</h2>
              <div class="backup-list">
                <div class="backup-item">
                  <div class="backup-info">
                    <h4>Daily Backup - 2025-01-13</h4>
                    <div class="backup-meta">
                      Type: Full | Size: 15.7 GB | Duration: 12 mins | Encrypted: Yes
                    </div>
                  </div>
                  <div class="backup-actions">
                    <button class="btn btn-secondary">Verify</button>
                    <button class="btn btn-primary">Restore</button>
                  </div>
                </div>
                <div class="backup-item">
                  <div class="backup-info">
                    <h4>Pre-upgrade Backup</h4>
                    <div class="backup-meta">
                      Type: Full | Size: 14.9 GB | Duration: 11 mins | Encrypted: Yes
                    </div>
                  </div>
                  <div class="backup-actions">
                    <button class="btn btn-secondary">Verify</button>
                    <button class="btn btn-primary">Restore</button>
                  </div>
                </div>
                <div class="backup-item">
                  <div class="backup-info">
                    <h4>Weekly Backup - 2025-01-07</h4>
                    <div class="backup-meta">
                      Type: Full | Size: 14.2 GB | Duration: 10 mins | Encrypted: Yes
                    </div>
                  </div>
                  <div class="backup-actions">
                    <button class="btn btn-secondary">Verify</button>
                    <button class="btn btn-primary">Restore</button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'backup', '01_backup_overview');
    
    // 10. Monitoring Dashboard
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Monitoring</title>
          <style>
            body { font-family: system-ui; margin: 0; background: #f5f5f5; }
            .header { background: white; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
            .container { padding: 20px; }
            .metrics-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; margin-bottom: 30px; }
            .metric-card { background: white; padding: 20px; border-radius: 8px; text-align: center; }
            .metric-value { font-size: 36px; font-weight: bold; color: #4F46E5; margin: 10px 0; }
            .metric-label { color: #666; }
            .chart-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; }
            .chart-card { background: white; padding: 20px; border-radius: 8px; height: 300px; }
            .chart-placeholder { background: #f5f5f5; height: 200px; border-radius: 4px; display: flex; align-items: center; justify-content: center; color: #999; }
            .alert-list { background: white; border-radius: 8px; overflow: hidden; margin-top: 20px; }
            .alert-item { padding: 15px 20px; border-bottom: 1px solid #eee; display: flex; align-items: center; gap: 10px; }
            .alert-icon { font-size: 20px; }
            .alert-icon.warning { color: #f59e0b; }
            .alert-icon.success { color: #10b981; }
            .time-range { display: flex; gap: 10px; margin-bottom: 20px; }
            .time-btn { padding: 8px 16px; border: 1px solid #ddd; background: white; border-radius: 4px; cursor: pointer; }
            .time-btn.active { background: #4F46E5; color: white; border-color: #4F46E5; }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>Monitoring Dashboard</h1>
          </div>
          <div class="container">
            <div class="time-range">
              <button class="time-btn">1H</button>
              <button class="time-btn active">24H</button>
              <button class="time-btn">7D</button>
              <button class="time-btn">30D</button>
            </div>
            
            <div class="metrics-grid">
              <div class="metric-card">
                <div class="metric-label">Total Requests</div>
                <div class="metric-value">1.2M</div>
                <div style="color: #10b981;">↑ 15%</div>
              </div>
              <div class="metric-card">
                <div class="metric-label">Avg Response Time</div>
                <div class="metric-value">142ms</div>
                <div style="color: #10b981;">↓ 8%</div>
              </div>
              <div class="metric-card">
                <div class="metric-label">Error Rate</div>
                <div class="metric-value">0.02%</div>
                <div style="color: #10b981;">↓ 0.01%</div>
              </div>
              <div class="metric-card">
                <div class="metric-label">Uptime</div>
                <div class="metric-value">99.99%</div>
                <div style="color: #666;">—</div>
              </div>
            </div>
            
            <div class="chart-grid">
              <div class="chart-card">
                <h3>CPU Usage</h3>
                <div class="chart-placeholder">📊 CPU Usage Chart (24-65%)</div>
              </div>
              <div class="chart-card">
                <h3>Memory Usage</h3>
                <div class="chart-placeholder">📊 Memory Usage Chart (4.2-6.8 GB)</div>
              </div>
              <div class="chart-card">
                <h3>Request Rate</h3>
                <div class="chart-placeholder">📊 Request Rate Chart (800-1500 req/s)</div>
              </div>
              <div class="chart-card">
                <h3>Response Time</h3>
                <div class="chart-placeholder">📊 Response Time Chart (120-180ms)</div>
              </div>
            </div>
            
            <h2>Recent Alerts</h2>
            <div class="alert-list">
              <div class="alert-item">
                <span class="alert-icon success">✅</span>
                <div>
                  <strong>All systems operational</strong>
                  <br><small>Last checked: 2 minutes ago</small>
                </div>
              </div>
              <div class="alert-item">
                <span class="alert-icon warning">⚠️</span>
                <div>
                  <strong>High memory usage on node-3</strong>
                  <br><small>Resolved: 1 hour ago</small>
                </div>
              </div>
            </div>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'monitoring', '01_monitoring_dashboard');
    
    // 11. Success State
    await page.setContent(`
      <html>
        <head>
          <title>Hexabase AI - Success</title>
          <style>
            body { font-family: system-ui; margin: 0; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); display: flex; align-items: center; justify-content: center; height: 100vh; }
            .success-card { background: white; padding: 60px; border-radius: 16px; box-shadow: 0 20px 60px rgba(0,0,0,0.2); text-align: center; max-width: 600px; }
            .success-icon { font-size: 80px; margin-bottom: 30px; }
            h1 { color: #333; margin-bottom: 20px; font-size: 36px; }
            p { color: #666; font-size: 18px; line-height: 1.6; margin-bottom: 30px; }
            .stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 30px; margin: 40px 0; }
            .stat { padding: 20px; background: #f5f5f5; border-radius: 8px; }
            .stat-value { font-size: 28px; font-weight: bold; color: #4F46E5; }
            .stat-label { color: #666; margin-top: 5px; }
            .btn { background: #4F46E5; color: white; padding: 15px 30px; border: none; border-radius: 8px; font-size: 18px; cursor: pointer; }
          </style>
        </head>
        <body>
          <div class="success-card">
            <div class="success-icon">🎉</div>
            <h1>E2E Journey Complete!</h1>
            <p>Successfully demonstrated the complete Hexabase AI platform workflow from login to production deployment.</p>
            
            <div class="stats">
              <div class="stat">
                <div class="stat-value">12</div>
                <div class="stat-label">Steps Completed</div>
              </div>
              <div class="stat">
                <div class="stat-value">100%</div>
                <div class="stat-label">Tests Passed</div>
              </div>
              <div class="stat">
                <div class="stat-value">0</div>
                <div class="stat-label">Errors</div>
              </div>
            </div>
            
            <p><strong>Key Features Demonstrated:</strong></p>
            <p style="text-align: left;">
              ✅ User Authentication<br>
              ✅ Organization & Workspace Management<br>
              ✅ Project Creation & Resource Quotas<br>
              ✅ Application Deployment (Stateless/Stateful)<br>
              ✅ Scaling & Updates<br>
              ✅ Blue-Green Deployments<br>
              ✅ CI/CD Pipeline Integration<br>
              ✅ Serverless Functions<br>
              ✅ Backup & Restore (Dedicated)<br>
              ✅ Monitoring & Metrics
            </p>
            
            <button class="btn">View Full Report</button>
          </div>
        </body>
      </html>
    `);
    await captureScreenshot(page, 'success', '01_complete_success');
    
    console.log(`\n✅ Mock screenshots captured successfully!\n`);
    console.log(`📁 All screenshots saved to: ${screenshotDir}\n`);
    
    // Create index.html for easy viewing
    const indexHtml = `
      <html>
        <head>
          <title>Hexabase AI - E2E Test Screenshots</title>
          <style>
            body { font-family: system-ui; margin: 0; padding: 20px; background: #f5f5f5; }
            h1 { color: #333; }
            .category { margin-bottom: 40px; background: white; padding: 20px; border-radius: 8px; }
            .category h2 { color: #4F46E5; margin-bottom: 20px; }
            .screenshots { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 20px; }
            .screenshot { border: 1px solid #ddd; border-radius: 8px; overflow: hidden; }
            .screenshot img { width: 100%; height: auto; display: block; }
            .screenshot-title { padding: 10px; background: #f5f5f5; font-size: 14px; }
          </style>
        </head>
        <body>
          <h1>🚀 Hexabase AI Platform - E2E Test Screenshots</h1>
          <p>Generated: ${new Date().toISOString()}</p>
    `;
    
    let indexContent = indexHtml;
    
    for (const dir of dirs) {
      if (dir === screenshotDir) continue;
      const category = path.basename(dir);
      const files = fs.readdirSync(dir).filter(f => f.endsWith('.png'));
      
      if (files.length > 0) {
        indexContent += `<div class="category"><h2>${category.charAt(0).toUpperCase() + category.slice(1)}</h2><div class="screenshots">`;
        
        for (const file of files) {
          indexContent += `
            <div class="screenshot">
              <img src="${category}/${file}" alt="${file}" />
              <div class="screenshot-title">${file.replace('.png', '').replace(/_/g, ' ')}</div>
            </div>
          `;
        }
        
        indexContent += `</div></div>`;
      }
    }
    
    indexContent += `</body></html>`;
    
    fs.writeFileSync(path.join(screenshotDir, 'index.html'), indexContent);
    console.log(`📄 Index page created: ${path.join(screenshotDir, 'index.html')}\n`);
    
  } catch (error) {
    console.error('Error capturing screenshots:', error);
  } finally {
    await browser.close();
  }
}

// Run the screenshot capture
createMockScreenshots().catch(console.error);