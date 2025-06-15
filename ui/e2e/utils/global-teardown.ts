import { FullConfig } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Global teardown function for E2E tests
 * Runs once after all tests
 */
async function globalTeardown(config: FullConfig) {
  console.log('\n🏁 Starting E2E test global teardown...');
  
  // Update session info with end time
  const sessionInfoPath = path.join(config.rootDir, 'test-results', 'session-info.json');
  if (fs.existsSync(sessionInfoPath)) {
    try {
      const sessionInfo = JSON.parse(fs.readFileSync(sessionInfoPath, 'utf-8'));
      sessionInfo.endTime = new Date().toISOString();
      sessionInfo.duration = new Date(sessionInfo.endTime).getTime() - new Date(sessionInfo.startTime).getTime();
      
      fs.writeFileSync(sessionInfoPath, JSON.stringify(sessionInfo, null, 2));
      console.log(`⏱️  Total test duration: ${(sessionInfo.duration / 1000).toFixed(2)}s`);
    } catch (error) {
      console.error('Error updating session info:', error);
    }
  }
  
  // Generate summary report
  const summaryPath = path.join(config.rootDir, 'test-results', 'summary.md');
  const summary = `# E2E Test Summary

**Date**: ${new Date().toLocaleString()}
**Environment**: ${process.env.NODE_ENV || 'test'}
**Base URL**: ${config.use?.baseURL || 'http://localhost:3000'}

## Test Results

Check the following locations for detailed results:
- HTML Report: \`test-results/html/index.html\`
- Videos: \`test-results/videos/\`
- Screenshots: \`test-results/screenshots/\`
- Traces: \`test-results/traces/\`

## Debug Information

To debug failed tests:
1. Open the HTML report to see detailed failure information
2. Use \`npx playwright show-trace <trace-file>\` to replay test execution
3. Check videos and screenshots for visual debugging

## Rerun Failed Tests

To rerun only failed tests:
\`\`\`bash
npx playwright test --last-failed
\`\`\`

To debug a specific test:
\`\`\`bash
npx playwright test <test-file> --debug
\`\`\`
`;
  
  fs.writeFileSync(summaryPath, summary);
  console.log('📄 Test summary written to:', summaryPath);
  
  // Clean up old test results (keep last 5 runs)
  try {
    const testResultsDir = path.join(config.rootDir, 'test-results');
    const files = fs.readdirSync(testResultsDir)
      .filter(f => f.startsWith('artifacts-') || f.startsWith('html-report-'))
      .map(f => ({
        name: f,
        path: path.join(testResultsDir, f),
        time: fs.statSync(path.join(testResultsDir, f)).mtime.getTime()
      }))
      .sort((a, b) => b.time - a.time);
    
    // Remove old results (keep last 5)
    const toRemove = files.slice(5);
    for (const file of toRemove) {
      fs.rmSync(file.path, { recursive: true, force: true });
      console.log('🗑️  Cleaned up old results:', file.name);
    }
  } catch (error) {
    console.error('Error cleaning up old results:', error);
  }
  
  console.log('✅ Global teardown completed');
}

export default globalTeardown;