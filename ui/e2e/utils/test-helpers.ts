import { Page } from '@playwright/test';

/**
 * Utility functions for E2E tests
 */

/**
 * Generate unique names for test resources
 */
export function generateResourceName(prefix: string): string {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 1000);
  return `${prefix}-${timestamp}-${random}`;
}

export function generateProjectName(): string {
  return generateResourceName('e2e-project');
}

export function generateAppName(): string {
  return generateResourceName('e2e-app');
}

export function generateWorkspaceName(): string {
  return generateResourceName('e2e-workspace');
}

export function generateOrgName(): string {
  return generateResourceName('e2e-org');
}

/**
 * Wait for element with retry logic
 */
export async function waitForElementWithRetry(
  page: Page,
  selector: string,
  options: {
    timeout?: number;
    retries?: number;
    retryDelay?: number;
  } = {}
) {
  const { timeout = 30000, retries = 3, retryDelay = 1000 } = options;
  
  for (let i = 0; i < retries; i++) {
    try {
      await page.waitForSelector(selector, { timeout: timeout / retries });
      return page.locator(selector);
    } catch (error) {
      if (i === retries - 1) throw error;
      await page.waitForTimeout(retryDelay);
    }
  }
}

/**
 * Take screenshot with descriptive name
 */
export async function takeScreenshot(page: Page, name: string) {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  await page.screenshot({
    path: `test-results/screenshots/${name}-${timestamp}.png`,
    fullPage: true,
  });
}

/**
 * Mock successful API response
 */
export async function mockSuccessResponse(
  page: Page,
  urlPattern: string,
  responseData: any,
  delay: number = 100
) {
  await page.route(urlPattern, async (route) => {
    await page.waitForTimeout(delay); // Simulate network delay
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(responseData),
    });
  });
}

/**
 * Mock API error response
 */
export async function mockErrorResponse(
  page: Page,
  urlPattern: string,
  error: { status: number; message: string },
  delay: number = 100
) {
  await page.route(urlPattern, async (route) => {
    await page.waitForTimeout(delay);
    await route.fulfill({
      status: error.status,
      contentType: 'application/json',
      body: JSON.stringify({ error: error.message }),
    });
  });
}

/**
 * Generate random resource configuration
 */
export function generateRandomConfig() {
  const images = [
    'nginx:latest',
    'nginx:1.21',
    'httpd:2.4',
    'node:16-alpine',
    'python:3.9-slim',
    'redis:6-alpine',
    'postgres:14-alpine',
    'mysql:8',
    'mongo:5',
  ];

  const envVars = {
    APP_ENV: ['development', 'staging', 'production'],
    LOG_LEVEL: ['debug', 'info', 'warn', 'error'],
    FEATURE_FLAG: ['enabled', 'disabled'],
  };

  return {
    image: images[Math.floor(Math.random() * images.length)],
    replicas: Math.floor(Math.random() * 5) + 1,
    port: [80, 3000, 5000, 8080, 9000][Math.floor(Math.random() * 5)],
    cpu: ['100m', '250m', '500m', '1'][Math.floor(Math.random() * 4)],
    memory: ['128Mi', '256Mi', '512Mi', '1Gi'][Math.floor(Math.random() * 4)],
    env: Object.entries(envVars).reduce((acc, [key, values]) => {
      acc[key] = values[Math.floor(Math.random() * values.length)];
      return acc;
    }, {} as Record<string, string>),
  };
}

/**
 * Wait for notification to appear and disappear
 */
export async function expectNotification(page: Page, text: string | RegExp) {
  // Wait for notification to appear
  const notification = page.getByRole('alert').filter({ hasText: text });
  await notification.waitFor({ state: 'visible', timeout: 5000 });
  
  // Wait for it to disappear (auto-dismiss)
  await notification.waitFor({ state: 'hidden', timeout: 10000 });
}

/**
 * Fill form with delays to simulate human typing
 */
export async function fillFormWithDelay(
  page: Page,
  inputs: Array<{ selector: string; value: string }>,
  delay: number = 100
) {
  for (const { selector, value } of inputs) {
    const input = page.locator(selector);
    await input.click();
    await input.fill(''); // Clear first
    
    // Type character by character with delay
    for (const char of value) {
      await input.type(char, { delay: delay / 10 });
    }
    
    await page.waitForTimeout(delay);
  }
}

/**
 * Check if element has expected CSS class
 */
export async function hasClass(locator: any, className: string): Promise<boolean> {
  const classes = await locator.getAttribute('class');
  return classes?.includes(className) || false;
}

/**
 * Scroll element into view
 */
export async function scrollIntoView(page: Page, selector: string) {
  await page.evaluate((sel) => {
    const element = document.querySelector(sel);
    element?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }, selector);
  
  // Wait for scroll to complete
  await page.waitForTimeout(500);
}

/**
 * Wait for loading indicators to disappear
 */
export async function waitForLoadingComplete(page: Page) {
  // Wait for common loading indicators
  const loadingSelectors = [
    '[data-testid="loading-spinner"]',
    '[data-testid="loading-skeleton"]',
    '.loading',
    '.spinner',
    '[aria-busy="true"]',
  ];
  
  for (const selector of loadingSelectors) {
    const elements = await page.$$(selector);
    if (elements.length > 0) {
      await page.waitForSelector(selector, { state: 'hidden', timeout: 30000 });
    }
  }
}

/**
 * Retry an action until it succeeds
 */
export async function retryAction<T>(
  action: () => Promise<T>,
  options: {
    retries?: number;
    delay?: number;
    timeout?: number;
  } = {}
): Promise<T> {
  const { retries = 3, delay = 1000, timeout = 30000 } = options;
  const startTime = Date.now();
  
  for (let i = 0; i < retries; i++) {
    try {
      return await action();
    } catch (error) {
      if (Date.now() - startTime > timeout) {
        throw new Error(`Action timed out after ${timeout}ms`);
      }
      
      if (i === retries - 1) throw error;
      
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
  
  throw new Error('Retry action failed');
}

/**
 * Format bytes to human readable format
 */
export function formatBytes(bytes: number): string {
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  if (bytes === 0) return '0 Bytes';
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
}

/**
 * Parse resource string (e.g., "100m", "1Gi") to number
 */
export function parseResourceValue(value: string): number {
  const match = value.match(/^(\d+(?:\.\d+)?)\s*([a-zA-Z]+)?$/);
  if (!match) return 0;
  
  const num = parseFloat(match[1]);
  const unit = match[2]?.toLowerCase();
  
  const multipliers: Record<string, number> = {
    m: 0.001, // millicores
    k: 1000,
    ki: 1024,
    m: 1000000,
    mi: 1024 * 1024,
    g: 1000000000,
    gi: 1024 * 1024 * 1024,
  };
  
  return num * (multipliers[unit || ''] || 1);
}

/**
 * Compare semantic versions
 */
export function compareVersions(v1: string, v2: string): number {
  const parts1 = v1.split('.').map(Number);
  const parts2 = v2.split('.').map(Number);
  
  for (let i = 0; i < Math.max(parts1.length, parts2.length); i++) {
    const p1 = parts1[i] || 0;
    const p2 = parts2[i] || 0;
    
    if (p1 < p2) return -1;
    if (p1 > p2) return 1;
  }
  
  return 0;
}