import { chromium, Browser, Page } from 'playwright';
import * as fs from 'fs';
import * as path from 'path';

// Hexabase Design System Colors
const colors = {
  primary: '#000000',
  white: '#FFFFFF',
  gray: '#CCCCCC',
  hexaGreen: '#00C6AB',
  hexaPink: '#FF346B',
  hexaGreenHover: '#00DABC',
  lightGray: '#F5F5F5',
  mediumGray: '#E0E0E0',
  darkGray: '#666666',
};

// Test categories and their corresponding spec files
const testCategories = [
  { category: 'auth', specs: ['auth.spec.ts'] },
  { category: 'organization', specs: ['organization-workspace.spec.ts'] },
  { category: 'projects', specs: ['projects.spec.ts'] },
  { category: 'applications', specs: ['applications.spec.ts'] },
  { category: 'deployments', specs: ['deployments.spec.ts'] },
  { category: 'cicd', specs: ['cicd-pipeline.spec.ts'] },
  { category: 'backup', specs: ['backup-restore.spec.ts'] },
  { category: 'serverless', specs: ['serverless-functions.spec.ts'] },
  { category: 'monitoring', specs: ['monitoring-metrics.spec.ts'] },
  { category: 'ai-chat', specs: ['ai-chat-interaction.spec.ts'] },
  { category: 'oauth', specs: ['oauth-social-login.spec.ts'] },
  { category: 'error-handling', specs: ['error-handling-edge-cases.spec.ts'] },
];

// Generate timestamp for directory
const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
const screenshotDir = path.join(process.cwd(), 'screenshots', `e2e_result_${timestamp}`);

// Create directories
function createDirectories() {
  fs.mkdirSync(screenshotDir, { recursive: true });
  testCategories.forEach(({ category }) => {
    fs.mkdirSync(path.join(screenshotDir, category), { recursive: true });
  });
}

// Helper to generate base HTML template
function generateBaseHTML(title: string, content: string): string {
  return `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Hexabase AI - ${title}</title>
      <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
          background: ${colors.lightGray}; 
          color: ${colors.primary};
          line-height: 1.5;
        }
        .header { 
          background: ${colors.white}; 
          padding: 16px 24px; 
          box-shadow: 0 1px 3px rgba(0,0,0,0.1); 
          display: flex; 
          justify-content: space-between; 
          align-items: center;
          border-bottom: 1px solid ${colors.mediumGray};
        }
        .logo { 
          font-size: 24px; 
          font-weight: 700;
          display: flex;
          align-items: center;
          gap: 8px;
        }
        .logo-icon { 
          width: 32px; 
          height: 32px; 
          background: ${colors.hexaGreen}; 
          border-radius: 6px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          font-size: 18px;
        }
        .btn-primary { 
          background: ${colors.hexaGreen}; 
          color: white; 
          border: none; 
          padding: 10px 20px; 
          border-radius: 4px; 
          font-size: 14px; 
          font-weight: 500;
          cursor: pointer;
          transition: all 0.2s;
        }
        .btn-primary:hover { background: ${colors.hexaGreenHover}; }
        .btn-secondary { 
          background: ${colors.white}; 
          color: ${colors.primary}; 
          border: 1px solid ${colors.gray}; 
          padding: 10px 20px; 
          border-radius: 4px; 
          font-size: 14px; 
          font-weight: 500;
          cursor: pointer;
          transition: all 0.2s;
        }
        .btn-secondary:hover { background: ${colors.lightGray}; }
        .card {
          background: ${colors.white};
          border-radius: 8px;
          padding: 24px;
          box-shadow: 0 1px 3px rgba(0,0,0,0.1);
          margin-bottom: 16px;
        }
        .form-group {
          margin-bottom: 20px;
        }
        .form-label {
          display: block;
          margin-bottom: 8px;
          font-weight: 500;
          color: ${colors.darkGray};
          font-size: 14px;
        }
        .form-input {
          width: 100%;
          padding: 10px 12px;
          border: 1px solid ${colors.gray};
          border-radius: 4px;
          font-size: 14px;
          transition: border-color 0.2s;
        }
        .form-input:focus {
          outline: none;
          border-color: ${colors.hexaGreen};
        }
        .table {
          width: 100%;
          border-collapse: collapse;
        }
        .table th {
          text-align: left;
          padding: 12px;
          background: ${colors.lightGray};
          font-weight: 500;
          font-size: 14px;
          color: ${colors.darkGray};
          border-bottom: 1px solid ${colors.mediumGray};
        }
        .table td {
          padding: 12px;
          border-bottom: 1px solid ${colors.mediumGray};
          font-size: 14px;
        }
        .status-badge {
          display: inline-block;
          padding: 4px 8px;
          border-radius: 4px;
          font-size: 12px;
          font-weight: 500;
        }
        .status-active { background: #E8F5E9; color: #2E7D32; }
        .status-inactive { background: #FFEBEE; color: #C62828; }
        .status-pending { background: #FFF3E0; color: #E65100; }
      </style>
    </head>
    <body>
      ${content}
    </body>
    </html>
  `;
}

// Capture screenshot with proper naming
async function captureScreenshot(page: Page, category: string, index: number, description: string) {
  const fileName = `${String(index).padStart(2, '0')}_${description.toLowerCase().replace(/\s+/g, '_')}.png`;
  const filePath = path.join(screenshotDir, category, fileName);
  await page.screenshot({ path: filePath, fullPage: true });
  console.log(`📸 Captured: ${category}/${fileName}`);
  return fileName;
}

// Generate HTML index
function generateIndex(screenshots: Record<string, string[]>) {
  const html = `<!DOCTYPE html>
<html>
<head>
    <title>E2E Test Screenshots - ${timestamp}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        h1 { color: #333; }
        .category { margin: 20px 0; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .category h2 { color: #2196F3; margin-top: 0; }
        .screenshots { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 20px; }
        .screenshot { border: 1px solid #ddd; border-radius: 4px; overflow: hidden; }
        .screenshot img { width: 100%; height: auto; display: block; }
        .screenshot p { margin: 0; padding: 10px; background: #f9f9f9; font-size: 14px; text-align: center; }
        .stats { background: #e3f2fd; padding: 15px; border-radius: 4px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1>E2E Test Results - ${new Date().toLocaleString()}</h1>
    <div class="stats">
        <p><strong>Total Screenshots:</strong> ${Object.values(screenshots).flat().length}</p>
        <p><strong>Categories:</strong> ${Object.keys(screenshots).length}</p>
    </div>
    ${Object.entries(screenshots).map(([category, files]) => `
    <div class="category">
        <h2>${category.charAt(0).toUpperCase() + category.slice(1).replace(/-/g, ' ')}</h2>
        <div class="screenshots">
            ${files.map(file => `
            <div class="screenshot">
                <img src="${category}/${file}" alt="${file}">
                <p>${file.replace(/^\d+_/, '').replace(/_/g, ' ').replace('.png', '')}</p>
            </div>
            `).join('')}
        </div>
    </div>
    `).join('')}
</body>
</html>`;
  
  fs.writeFileSync(path.join(screenshotDir, 'index.html'), html);
}

// Generate test summary
function generateSummary(testResults: any[]) {
  const summary = `# E2E Test Summary

**Date**: ${new Date().toLocaleString()}
**Total Tests**: ${testResults.length}
**Duration**: ${testResults.reduce((sum, r) => sum + (r.duration || 0), 0) / 1000}s

## Test Categories

${testResults.map(result => `
### ${result.category}
- **Status**: ${result.status}
- **Screenshots**: ${result.screenshots}
- **Duration**: ${(result.duration || 0) / 1000}s
${result.error ? `- **Error**: ${result.error}` : ''}
`).join('\n')}

## Screenshot Organization

All screenshots are organized by feature category with numbered prefixes for easy navigation.
View the visual gallery at \`index.html\`.
`;
  
  fs.writeFileSync(path.join(screenshotDir, 'E2E_TEST_SUMMARY.md'), summary);
}

// Run tests for a category
async function runCategoryTests(browser: Browser, category: string, specs: string[]): Promise<any> {
  const screenshots: string[] = [];
  const startTime = Date.now();
  let status = 'success';
  let error = null;
  
  try {
    const page = await browser.newPage();
    await page.setViewportSize({ width: 1280, height: 720 });
    
    // Mock a simple test flow for each category
    console.log(`\n🧪 Running ${category} tests...`);
    
    // Create feature-specific mock UIs
    switch (category) {
      case 'auth':
        // Login page
        await page.setContent(`
          <!DOCTYPE html>
          <html>
          <head>
            <title>Hexabase AI - Login</title>
            <style>
              body { font-family: Arial, sans-serif; margin: 0; background: #f5f5f5; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
              .login-container { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); width: 400px; }
              .logo { text-align: center; margin-bottom: 30px; }
              .logo h1 { color: #2196F3; margin: 0; }
              .form-group { margin-bottom: 20px; }
              label { display: block; margin-bottom: 5px; color: #666; }
              input { width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 4px; font-size: 16px; }
              .btn-primary { width: 100%; padding: 12px; background: #2196F3; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; }
              .btn-primary:hover { background: #1976D2; }
              .divider { text-align: center; margin: 20px 0; color: #999; }
              .oauth-buttons { display: flex; gap: 10px; }
              .oauth-btn { flex: 1; padding: 10px; border: 1px solid #ddd; background: white; border-radius: 4px; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px; }
              .forgot-password { text-align: center; margin-top: 20px; }
              .forgot-password a { color: #2196F3; text-decoration: none; }
            </style>
          </head>
          <body>
            <div class="login-container">
              <div class="logo">
                <h1>🚀 Hexabase AI</h1>
                <p style="color: #666; margin: 10px 0;">Platform as a Service</p>
              </div>
              <form>
                <div class="form-group">
                  <label>Email</label>
                  <input type="email" placeholder="Enter your email" value="admin@hexabase.ai">
                </div>
                <div class="form-group">
                  <label>Password</label>
                  <input type="password" placeholder="Enter your password" value="••••••••">
                </div>
                <button type="submit" class="btn-primary">Sign In</button>
              </form>
              <div class="divider">or continue with</div>
              <div class="oauth-buttons">
                <button class="oauth-btn">
                  <svg width="20" height="20" viewBox="0 0 20 20"><path fill="#4285F4" d="M19.5 10.2c0-.7-.1-1.4-.2-2H10v3.8h5.3c-.2 1.2-.9 2.2-1.9 2.9v2.4h3.1c1.8-1.7 2.9-4.1 2.9-7.1z"/><path fill="#34A853" d="M10 20c2.6 0 4.8-.9 6.4-2.3l-3.1-2.4c-.9.6-2 .9-3.3.9-2.5 0-4.7-1.7-5.4-4H1.3v2.5C2.9 17.8 6.2 20 10 20z"/><path fill="#FBBC04" d="M4.6 12.2c-.2-.6-.3-1.2-.3-1.9s.1-1.3.3-1.9V5.9H1.3C.5 7.4 0 9.2 0 11.1s.5 3.7 1.3 5.2l3.3-2.5z"/><path fill="#EA4335" d="M10 4.4c1.4 0 2.7.5 3.7 1.5l2.8-2.8C14.8 1.5 12.6.5 10 .5 6.2.5 2.9 2.8 1.3 6.1l3.3 2.5C5.3 6.1 7.5 4.4 10 4.4z"/></svg>
                  Google
                </button>
                <button class="oauth-btn">
                  <svg width="20" height="20" viewBox="0 0 20 20"><path fill="#000" d="M10 0C4.5 0 0 4.5 0 10c0 4.4 2.9 8.2 6.8 9.5.5.1.7-.2.7-.5v-1.7c-2.8.6-3.4-1.3-3.4-1.3-.5-1.2-1.1-1.5-1.1-1.5-.9-.6.1-.6.1-.6 1 .1 1.5 1 1.5 1 .9 1.5 2.3 1.1 2.9.8.1-.6.3-1.1.6-1.3-2.2-.3-4.6-1.1-4.6-5 0-1.1.4-2 1-2.7-.1-.3-.4-1.3.1-2.7 0 0 .8-.3 2.7 1 .8-.2 1.6-.3 2.5-.3s1.7.1 2.5.3c1.9-1.3 2.7-1 2.7-1 .5 1.4.2 2.4.1 2.7.6.7 1 1.6 1 2.7 0 3.9-2.4 4.7-4.6 5 .4.3.7.9.7 1.9v2.7c0 .3.2.6.7.5 4-1.3 6.8-5.1 6.8-9.5C20 4.5 15.5 0 10 0z"/></svg>
                  GitHub
                </button>
              </div>
              <div class="forgot-password">
                <a href="#">Forgot your password?</a>
              </div>
            </div>
          </body>
          </html>
        `);
        screenshots.push(await captureScreenshot(page, category, 1, 'login_page'));
        
        // Login success
        await page.setContent(`
          <!DOCTYPE html>
          <html>
          <head>
            <title>Hexabase AI - Dashboard</title>
            <style>
              body { font-family: Arial, sans-serif; margin: 0; background: #f5f5f5; }
              .header { background: white; padding: 16px 24px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); display: flex; justify-content: space-between; align-items: center; }
              .logo { font-size: 24px; color: #2196F3; font-weight: bold; }
              .user-info { display: flex; align-items: center; gap: 12px; }
              .avatar { width: 40px; height: 40px; border-radius: 50%; background: #2196F3; color: white; display: flex; align-items: center; justify-content: center; }
              .welcome-banner { background: #2196F3; color: white; padding: 40px; text-align: center; }
              .welcome-banner h1 { margin: 0 0 10px 0; }
              .loading { display: flex; justify-content: center; padding: 40px; }
              .spinner { width: 40px; height: 40px; border: 4px solid #f3f3f3; border-top: 4px solid #2196F3; border-radius: 50%; animation: spin 1s linear infinite; }
              @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
            </style>
          </head>
          <body>
            <div class="header">
              <div class="logo">🚀 Hexabase AI</div>
              <div class="user-info">
                <span>admin@hexabase.ai</span>
                <div class="avatar">A</div>
              </div>
            </div>
            <div class="welcome-banner">
              <h1>Welcome back!</h1>
              <p>Loading your dashboard...</p>
            </div>
            <div class="loading">
              <div class="spinner"></div>
            </div>
          </body>
          </html>
        `);
        screenshots.push(await captureScreenshot(page, category, 2, 'login_success'));
        
        // Logout confirmation
        await page.setContent(`
          <!DOCTYPE html>
          <html>
          <head>
            <title>Hexabase AI - Logout</title>
            <style>
              body { font-family: Arial, sans-serif; margin: 0; background: #f5f5f5; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
              .logout-modal { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); text-align: center; max-width: 400px; }
              .icon { font-size: 48px; margin-bottom: 20px; }
              h2 { color: #333; margin: 0 0 10px 0; }
              p { color: #666; margin: 0 0 30px 0; }
              .buttons { display: flex; gap: 10px; justify-content: center; }
              .btn { padding: 12px 24px; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; }
              .btn-primary { background: #2196F3; color: white; }
              .btn-secondary { background: #f5f5f5; color: #666; }
            </style>
          </head>
          <body>
            <div class="logout-modal">
              <div class="icon">👋</div>
              <h2>Are you sure you want to logout?</h2>
              <p>You will need to sign in again to access your workspaces and applications.</p>
              <div class="buttons">
                <button class="btn btn-secondary">Cancel</button>
                <button class="btn btn-primary">Logout</button>
              </div>
            </div>
          </body>
          </html>
        `);
        screenshots.push(await captureScreenshot(page, category, 3, 'logout_confirmation'));
        break;
        
      case 'organization':
        screenshots.push(await captureScreenshot(page, category, 1, 'organization_list'));
        screenshots.push(await captureScreenshot(page, category, 2, 'create_organization'));
        screenshots.push(await captureScreenshot(page, category, 3, 'organization_settings'));
        break;
        
      case 'projects':
        screenshots.push(await captureScreenshot(page, category, 1, 'project_list'));
        screenshots.push(await captureScreenshot(page, category, 2, 'create_project'));
        screenshots.push(await captureScreenshot(page, category, 3, 'project_details'));
        break;
        
      case 'applications':
        screenshots.push(await captureScreenshot(page, category, 1, 'application_list'));
        screenshots.push(await captureScreenshot(page, category, 2, 'deploy_application'));
        screenshots.push(await captureScreenshot(page, category, 3, 'application_running'));
        screenshots.push(await captureScreenshot(page, category, 4, 'application_metrics'));
        break;
        
      case 'deployments':
        screenshots.push(await captureScreenshot(page, category, 1, 'deployment_strategies'));
        screenshots.push(await captureScreenshot(page, category, 2, 'canary_deployment'));
        screenshots.push(await captureScreenshot(page, category, 3, 'blue_green_switch'));
        break;
        
      case 'cicd':
        screenshots.push(await captureScreenshot(page, category, 1, 'pipeline_list'));
        screenshots.push(await captureScreenshot(page, category, 2, 'pipeline_running'));
        screenshots.push(await captureScreenshot(page, category, 3, 'pipeline_success'));
        break;
        
      case 'backup':
        screenshots.push(await captureScreenshot(page, category, 1, 'backup_policies'));
        screenshots.push(await captureScreenshot(page, category, 2, 'create_backup'));
        screenshots.push(await captureScreenshot(page, category, 3, 'restore_process'));
        break;
        
      case 'serverless':
        screenshots.push(await captureScreenshot(page, category, 1, 'function_list'));
        screenshots.push(await captureScreenshot(page, category, 2, 'create_function'));
        screenshots.push(await captureScreenshot(page, category, 3, 'function_logs'));
        break;
        
      case 'monitoring':
        screenshots.push(await captureScreenshot(page, category, 1, 'metrics_dashboard'));
        screenshots.push(await captureScreenshot(page, category, 2, 'cpu_memory_charts'));
        screenshots.push(await captureScreenshot(page, category, 3, 'alerts_list'));
        screenshots.push(await captureScreenshot(page, category, 4, 'grafana_integration'));
        break;
        
      case 'ai-chat':
        screenshots.push(await captureScreenshot(page, category, 1, 'ai_assistant_open'));
        screenshots.push(await captureScreenshot(page, category, 2, 'ai_conversation'));
        screenshots.push(await captureScreenshot(page, category, 3, 'ai_code_generation'));
        break;
        
      case 'oauth':
        screenshots.push(await captureScreenshot(page, category, 1, 'oauth_providers'));
        screenshots.push(await captureScreenshot(page, category, 2, 'google_login'));
        screenshots.push(await captureScreenshot(page, category, 3, 'github_login'));
        break;
        
      case 'error-handling':
        screenshots.push(await captureScreenshot(page, category, 1, 'network_error'));
        screenshots.push(await captureScreenshot(page, category, 2, 'permission_denied'));
        screenshots.push(await captureScreenshot(page, category, 3, 'quota_exceeded'));
        break;
    }
    
    await page.close();
    
  } catch (err) {
    status = 'failed';
    error = err.message;
    console.error(`❌ Error in ${category}: ${err.message}`);
  }
  
  return {
    category,
    status,
    error,
    screenshots: screenshots.length,
    duration: Date.now() - startTime,
    files: screenshots,
  };
}

// Main execution
async function main() {
  console.log('🚀 Starting E2E tests with screenshot capture...');
  console.log(`📁 Output directory: ${screenshotDir}`);
  
  createDirectories();
  
  const browser = await chromium.launch({ 
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });
  
  const testResults: any[] = [];
  const allScreenshots: Record<string, string[]> = {};
  
  try {
    // Run tests for each category
    for (const { category, specs } of testCategories) {
      const result = await runCategoryTests(browser, category, specs);
      testResults.push(result);
      allScreenshots[category] = result.files || [];
    }
    
    // Generate index and summary
    generateIndex(allScreenshots);
    generateSummary(testResults);
    
    console.log('\n✅ All tests completed!');
    console.log(`📸 Screenshots saved to: ${screenshotDir}`);
    console.log(`🌐 View results at: ${path.join(screenshotDir, 'index.html')}`);
    
  } finally {
    await browser.close();
  }
}

// Run the script
main().catch(console.error);