/**
 * Test tagging utilities for categorizing and filtering tests
 */

/**
 * Tag test as a smoke test - critical path that should always pass
 * Usage: test('login flow @smoke', async ({ page }) => { ... })
 */
export const SMOKE_TAG = '@smoke';

/**
 * Tag test for visual regression testing
 * Usage: test('dashboard layout @visual', async ({ page }) => { ... })
 */
export const VISUAL_TAG = '@visual';

/**
 * Tag test as critical - must pass before deployment
 * Usage: test('payment processing @critical', async ({ page }) => { ... })
 */
export const CRITICAL_TAG = '@critical';

/**
 * Tag test as flaky - known to be unstable
 * Usage: test('websocket connection @flaky', async ({ page }) => { ... })
 */
export const FLAKY_TAG = '@flaky';

/**
 * Tag test as slow - takes longer than 30s
 * Usage: test('full backup restore @slow', async ({ page }) => { ... })
 */
export const SLOW_TAG = '@slow';

/**
 * Tag test to run only in specific environments
 */
export const ENV_TAGS = {
  PRODUCTION_ONLY: '@production-only',
  STAGING_ONLY: '@staging-only',
  DEV_ONLY: '@dev-only',
};

/**
 * Helper to check if a test should run in current environment
 */
export function shouldRunInEnvironment(testTitle: string): boolean {
  const env = process.env.TEST_ENV || 'development';
  
  if (testTitle.includes(ENV_TAGS.PRODUCTION_ONLY) && env !== 'production') {
    return false;
  }
  
  if (testTitle.includes(ENV_TAGS.STAGING_ONLY) && env !== 'staging') {
    return false;
  }
  
  if (testTitle.includes(ENV_TAGS.DEV_ONLY) && env !== 'development') {
    return false;
  }
  
  return true;
}

/**
 * Get all tags from a test title
 */
export function getTestTags(testTitle: string): string[] {
  const tagPattern = /@[\w-]+/g;
  return testTitle.match(tagPattern) || [];
}

/**
 * Check if test has a specific tag
 */
export function hasTag(testTitle: string, tag: string): boolean {
  return testTitle.includes(tag);
}