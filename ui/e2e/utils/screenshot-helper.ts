import { Page, test } from '@playwright/test';
import * as path from 'path';

const timestamp = new Date().toISOString().replace(/:/g, '-').replace(/\./g, '_').slice(0, 19);
const screenshotBaseDir = path.join(process.cwd(), 'screenshots', `e2e_result_${timestamp}`);

export async function captureSuccessScreenshot(
  page: Page, 
  testName: string, 
  stepName: string,
  testInfo?: any
) {
  const fileName = `${testName.replace(/\s+/g, '_')}-${stepName.replace(/\s+/g, '_')}.png`;
  const filePath = path.join(screenshotBaseDir, fileName);
  
  try {
    await page.screenshot({
      path: filePath,
      fullPage: true,
    });
    
    if (testInfo) {
      await testInfo.attach(stepName, {
        body: await page.screenshot({ fullPage: true }),
        contentType: 'image/png',
      });
    }
    
    console.log(`Screenshot saved: ${filePath}`);
  } catch (error) {
    console.error(`Failed to capture screenshot: ${error}`);
  }
}

// Hook to capture screenshot after each successful test
export function setupScreenshotCapture() {
  test.afterEach(async ({ page }, testInfo) => {
    if (testInfo.status === 'passed') {
      const testTitle = testInfo.title.replace(/\s+/g, '_');
      const fileName = `${testTitle}_success.png`;
      const filePath = path.join(screenshotBaseDir, 'success', fileName);
      
      await page.screenshot({
        path: filePath,
        fullPage: true,
      });
      
      await testInfo.attach('Success Screenshot', {
        body: await page.screenshot({ fullPage: true }),
        contentType: 'image/png',
      });
    }
  });
}