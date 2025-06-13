import { defineConfig, devices } from '@playwright/test';
import baseConfig from './playwright.config';

const timestamp = new Date().toISOString().replace(/:/g, '-').replace(/\./g, '_').slice(0, 19);
const screenshotDir = `screenshots/e2e_result_${timestamp}`;

/**
 * Special configuration for screenshot capture run
 */
export default defineConfig({
  ...baseConfig,
  
  /* Use only Chromium for screenshot consistency */
  projects: [
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        /* Capture screenshot after each test */
        screenshot: {
          mode: 'only-on-failure',
          fullPage: true,
        },
      },
    },
  ],

  /* Custom reporter configuration */
  reporter: [
    ['list'],
    ['html', { outputFolder: `${screenshotDir}/html-report` }],
    ['json', { outputFile: `${screenshotDir}/results.json` }],
  ],

  /* Output directory for artifacts */
  outputDir: screenshotDir,

  /* Override use settings for screenshot capture */
  use: {
    ...baseConfig.use,
    /* Take screenshot after each test step */
    screenshot: {
      mode: 'only-on-failure',
      fullPage: true,
    },
    /* Record video for all tests */
    video: 'on',
    /* Slow down actions to capture better screenshots */
    actionTimeout: 15000,
    navigationTimeout: 30000,
  },

  /* Reduce parallel execution for better screenshot capture */
  workers: 1,
  fullyParallel: false,
});