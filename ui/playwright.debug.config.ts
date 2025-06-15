import { defineConfig, devices } from '@playwright/test';
import baseConfig from './playwright.config';

/**
 * Debug configuration for Playwright tests
 * Optimized for visual debugging and troubleshooting
 */
export default defineConfig({
  ...baseConfig,
  
  /* Run tests in headed mode by default */
  use: {
    ...baseConfig.use,
    /* Base URL to use in actions */
    baseURL: process.env.BASE_URL || 'http://localhost:3000',
    
    /* Run in headed mode for debugging */
    headless: false,
    
    /* Slow down actions by 500ms for visual debugging */
    launchOptions: {
      slowMo: 500,
    },
    
    /* Capture screenshot on every action */
    screenshot: {
      mode: 'on',
      fullPage: true,
    },
    
    /* Always record video */
    video: {
      mode: 'on',
      size: { width: 1280, height: 720 }
    },
    
    /* Always capture trace */
    trace: 'on',
    
    /* Extended timeouts for debugging */
    actionTimeout: 60000,
    navigationTimeout: 60000,
    
    /* Viewport size */
    viewport: { width: 1280, height: 720 },
    
    /* Accept downloads */
    acceptDownloads: true,
    
    /* Ignore HTTPS errors */
    ignoreHTTPSErrors: true,
    
    /* Permissions */
    permissions: ['geolocation', 'notifications'],
    
    /* Locale */
    locale: 'en-US',
    
    /* Timezone */
    timezoneId: 'America/New_York',
    
    /* Color scheme */
    colorScheme: 'light',
  },

  /* Test timeout */
  timeout: 120 * 1000,
  
  /* Expect timeout */
  expect: {
    timeout: 30 * 1000,
  },

  /* Run only one test at a time for debugging */
  workers: 1,
  fullyParallel: false,

  /* Fail fast on first failure */
  maxFailures: 1,

  /* Reporter configuration for debugging */
  reporter: [
    ['list', { printSteps: true }],
    ['html', { open: 'on-failure' }],
  ],

  /* Only use Chromium for consistent debugging */
  projects: [
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        /* Enable devtools */
        launchOptions: {
          devtools: true,
          slowMo: 500,
        },
      },
    },
  ],

  /* Global setup/teardown */
  globalSetup: require.resolve('./e2e/utils/global-setup'),
  globalTeardown: require.resolve('./e2e/utils/global-teardown'),
});