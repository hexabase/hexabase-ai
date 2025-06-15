import { test, expect } from '@playwright/test';
import { ApplicationGenerator } from '../fixtures/generators/application-generator';
import { MetricsGenerator } from '../fixtures/generators/metrics-generator';
import { TestDataManager } from '../utils/test-data-manager';

/**
 * Performance and Load Tests
 * Tests the system's behavior under various load conditions
 */
test.describe('Performance and Load Tests', () => {
  test.setTimeout(300000); // 5 minutes for load tests
  
  let testData: TestDataManager;

  test.beforeEach(async ({ page }) => {
    testData = new TestDataManager(page);
    await testData.loadScenario('enterprise');
  });

  test.describe('Page Load Performance', () => {
    test('should load dashboard within performance budget', async ({ page }) => {
      const startTime = Date.now();
      
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      expect(loadTime).toBeLessThan(3000); // 3 second budget
      
      // Check Core Web Vitals
      const metrics = await page.evaluate(() => {
        return new Promise((resolve) => {
          let fcp = 0;
          let lcp = 0;
          
          new PerformanceObserver((list) => {
            const entries = list.getEntries();
            entries.forEach((entry) => {
              if (entry.name === 'first-contentful-paint') {
                fcp = entry.startTime;
              }
            });
          }).observe({ entryTypes: ['paint'] });
          
          new PerformanceObserver((list) => {
            const entries = list.getEntries();
            lcp = entries[entries.length - 1].startTime;
            resolve({ fcp, lcp });
          }).observe({ entryTypes: ['largest-contentful-paint'] });
          
          setTimeout(() => resolve({ fcp, lcp }), 2000);
        });
      });
      
      expect(metrics.fcp).toBeLessThan(1500); // FCP < 1.5s
      expect(metrics.lcp).toBeLessThan(2500); // LCP < 2.5s
    });

    test('should handle slow network gracefully', async ({ page }) => {
      // Simulate slow 3G
      await page.route('**/*', async (route) => {
        await new Promise(resolve => setTimeout(resolve, 500)); // 500ms delay
        await route.continue();
      });
      
      await page.goto('/dashboard');
      
      // Should show loading states
      await expect(page.getByTestId('loading-skeleton')).toBeVisible();
      
      // Should eventually load
      await expect(page.getByTestId('dashboard-content')).toBeVisible({
        timeout: 30000
      });
    });
  });

  test.describe('Data Loading Performance', () => {
    test('should efficiently load large application lists', async ({ page }) => {
      const appGen = new ApplicationGenerator();
      const apps = appGen.generateMany(100);
      
      // Mock API response with many apps
      await page.route('**/api/applications', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: apps, total: apps.length })
        });
      });
      
      const startTime = Date.now();
      await page.goto('/projects/test-project/applications');
      
      // Check that virtualization/pagination is working
      const visibleApps = await page.getByTestId('app-card').count();
      expect(visibleApps).toBeLessThan(50); // Should not render all 100 at once
      
      // Check render time
      const renderTime = Date.now() - startTime;
      expect(renderTime).toBeLessThan(2000);
      
      // Test scrolling performance
      await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
      await page.waitForTimeout(100);
      
      // More apps should be visible after scroll
      const visibleAppsAfterScroll = await page.getByTestId('app-card').count();
      expect(visibleAppsAfterScroll).toBeGreaterThan(visibleApps);
    });

    test('should handle concurrent data fetching efficiently', async ({ page }) => {
      const requests: string[] = [];
      
      // Track all API requests
      page.on('request', (request) => {
        if (request.url().includes('/api/')) {
          requests.push(request.url());
        }
      });
      
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');
      
      // Check for duplicate requests
      const uniqueRequests = new Set(requests);
      expect(uniqueRequests.size).toBe(requests.length); // No duplicates
      
      // Check that requests are batched/parallelized where possible
      const requestTimings = await page.evaluate(() => {
        return performance.getEntriesByType('resource')
          .filter(entry => entry.name.includes('/api/'))
          .map(entry => ({
            url: entry.name,
            start: entry.startTime,
            duration: entry.duration
          }));
      });
      
      // Verify parallel requests start at similar times
      const startTimes = requestTimings.map(t => t.start);
      const maxStartTimeDiff = Math.max(...startTimes) - Math.min(...startTimes);
      expect(maxStartTimeDiff).toBeLessThan(100); // Requests start within 100ms
    });
  });

  test.describe('UI Responsiveness Under Load', () => {
    test('should remain responsive with many real-time updates', async ({ page }) => {
      await page.goto('/monitoring');
      
      const metricsGen = new MetricsGenerator();
      let updateCount = 0;
      
      // Simulate WebSocket updates
      await page.evaluate(() => {
        window.mockWebSocket = {
          send: (data: string) => {},
          close: () => {}
        };
      });
      
      // Send many metric updates
      const updateInterval = setInterval(async () => {
        await page.evaluate((metrics) => {
          window.dispatchEvent(new CustomEvent('ws-message', {
            detail: { type: 'metrics-update', data: metrics }
          }));
        }, metricsGen.generateResourceMetrics({
          duration: 60,
          interval: 1,
          workloadType: 'web'
        }));
        
        updateCount++;
      }, 100); // 10 updates per second
      
      // Wait for 3 seconds of updates
      await page.waitForTimeout(3000);
      clearInterval(updateInterval);
      
      // UI should still be responsive
      const button = page.getByTestId('refresh-metrics');
      const clickStart = Date.now();
      await button.click();
      const clickDuration = Date.now() - clickStart;
      
      expect(clickDuration).toBeLessThan(100); // Click should be fast
      expect(updateCount).toBeGreaterThan(25); // Should handle many updates
    });

    test('should handle rapid user interactions', async ({ page }) => {
      await page.goto('/applications');
      
      // Rapidly toggle multiple filters
      const filterToggles = [
        'filter-stateless',
        'filter-stateful',
        'filter-cronjob',
        'filter-serverless'
      ];
      
      const startTime = Date.now();
      
      for (let i = 0; i < 20; i++) {
        const filter = filterToggles[i % filterToggles.length];
        await page.getByTestId(filter).click();
      }
      
      const totalTime = Date.now() - startTime;
      expect(totalTime).toBeLessThan(2000); // All clicks within 2s
      
      // UI should reflect final state correctly
      await expect(page.getByTestId('filtered-results')).toBeVisible();
    });
  });

  test.describe('Memory Management', () => {
    test('should not leak memory during long sessions', async ({ page }) => {
      // Get initial memory usage
      const getMemoryUsage = () => page.evaluate(() => {
        if ('memory' in performance) {
          return (performance as any).memory.usedJSHeapSize;
        }
        return 0;
      });
      
      const initialMemory = await getMemoryUsage();
      
      // Perform many operations
      for (let i = 0; i < 10; i++) {
        await page.goto('/dashboard');
        await page.goto('/applications');
        await page.goto('/monitoring');
        
        // Trigger some state changes
        await page.getByTestId('create-app-button').click();
        await page.getByTestId('close-dialog').click();
      }
      
      // Force garbage collection if available
      await page.evaluate(() => {
        if (window.gc) window.gc();
      });
      
      await page.waitForTimeout(1000);
      
      const finalMemory = await getMemoryUsage();
      const memoryIncrease = finalMemory - initialMemory;
      
      // Memory increase should be reasonable (< 50MB)
      expect(memoryIncrease).toBeLessThan(50 * 1024 * 1024);
    });
  });

  test.describe('Concurrent User Simulation', () => {
    test('should handle multiple concurrent deployments', async ({ browser }) => {
      const userCount = 5;
      const contexts = await Promise.all(
        Array.from({ length: userCount }, () => browser.newContext())
      );
      
      const deploymentPromises = contexts.map(async (context, index) => {
        const page = await context.newPage();
        const appGen = new ApplicationGenerator();
        
        await page.goto('/applications');
        
        // Each user deploys an app
        const app = appGen.generate({ name: `load-test-app-${index}` });
        
        const startTime = Date.now();
        await page.getByTestId('deploy-app-button').click();
        await page.getByTestId('app-name-input').fill(app.name);
        await page.getByTestId('app-image-input').fill(app.image);
        await page.getByTestId('deploy-button').click();
        
        // Wait for deployment to complete
        await expect(page.getByTestId(`app-status-${app.name}`)).toHaveText('Running', {
          timeout: 60000
        });
        
        const deployTime = Date.now() - startTime;
        
        await context.close();
        return deployTime;
      });
      
      const deployTimes = await Promise.all(deploymentPromises);
      
      // All deployments should complete
      expect(deployTimes.length).toBe(userCount);
      
      // Average deploy time should be reasonable even under load
      const avgDeployTime = deployTimes.reduce((a, b) => a + b, 0) / userCount;
      expect(avgDeployTime).toBeLessThan(30000); // 30s average
    });
  });

  test.describe('API Rate Limiting', () => {
    test('should handle rate limit responses gracefully', async ({ page }) => {
      let requestCount = 0;
      
      // Simulate rate limiting after 10 requests
      await page.route('**/api/**', async (route) => {
        requestCount++;
        
        if (requestCount > 10) {
          await route.fulfill({
            status: 429,
            headers: {
              'X-RateLimit-Limit': '10',
              'X-RateLimit-Remaining': '0',
              'X-RateLimit-Reset': String(Date.now() + 60000)
            },
            body: JSON.stringify({
              error: 'Rate limit exceeded',
              retryAfter: 60
            })
          });
        } else {
          await route.continue();
        }
      });
      
      await page.goto('/dashboard');
      
      // Trigger multiple API calls
      for (let i = 0; i < 15; i++) {
        await page.getByTestId('refresh-button').click();
        await page.waitForTimeout(100);
      }
      
      // Should show rate limit message
      await expect(page.getByTestId('rate-limit-warning')).toBeVisible();
      await expect(page.getByTestId('rate-limit-warning')).toContainText('Please wait');
      
      // Should not crash or throw errors
      const errors: Error[] = [];
      page.on('pageerror', (error) => errors.push(error));
      
      await page.waitForTimeout(1000);
      expect(errors.length).toBe(0);
    });
  });

  test.describe('Large File Operations', () => {
    test('should handle large log file streaming', async ({ page }) => {
      await page.goto('/applications/test-app');
      
      // Mock large log stream
      await page.route('**/api/applications/*/logs', async (route) => {
        const encoder = new TextEncoder();
        const logLines = Array.from({ length: 1000 }, (_, i) => 
          `[2024-01-01T00:00:${String(i % 60).padStart(2, '0')}] Log line ${i}: ${faker.lorem.sentence()}\n`
        );
        
        await route.fulfill({
          status: 200,
          headers: {
            'Content-Type': 'text/event-stream',
            'Cache-Control': 'no-cache'
          },
          body: encoder.encode(logLines.join(''))
        });
      });
      
      await page.getByTestId('view-logs-button').click();
      
      // Logs should stream efficiently
      await expect(page.getByTestId('log-viewer')).toBeVisible();
      
      // Should virtualize/paginate logs
      const visibleLogs = await page.getByTestId('log-line').count();
      expect(visibleLogs).toBeLessThan(100); // Not all 1000 lines rendered
      
      // Scrolling should be smooth
      await page.getByTestId('log-viewer').evaluate(el => {
        el.scrollTop = el.scrollHeight;
      });
      
      await page.waitForTimeout(500);
      
      // More logs visible after scroll
      const visibleLogsAfterScroll = await page.getByTestId('log-line').count();
      expect(visibleLogsAfterScroll).toBeGreaterThan(visibleLogs);
    });
  });

  test.describe('Resource Cleanup', () => {
    test('should clean up resources after navigation', async ({ page }) => {
      // Track active timers and listeners
      await page.evaluateOnNewDocument(() => {
        window.activeTimers = new Set();
        window.activeListeners = new Map();
        
        const originalSetInterval = window.setInterval;
        const originalClearInterval = window.clearInterval;
        const originalAddEventListener = window.addEventListener;
        const originalRemoveEventListener = window.removeEventListener;
        
        window.setInterval = function(...args) {
          const id = originalSetInterval.apply(window, args);
          window.activeTimers.add(id);
          return id;
        };
        
        window.clearInterval = function(id) {
          window.activeTimers.delete(id);
          return originalClearInterval.call(window, id);
        };
        
        window.addEventListener = function(type, listener, ...args) {
          if (!window.activeListeners.has(type)) {
            window.activeListeners.set(type, new Set());
          }
          window.activeListeners.get(type).add(listener);
          return originalAddEventListener.call(window, type, listener, ...args);
        };
        
        window.removeEventListener = function(type, listener, ...args) {
          window.activeListeners.get(type)?.delete(listener);
          return originalRemoveEventListener.call(window, type, listener, ...args);
        };
      });
      
      // Navigate to monitoring page (which has timers)
      await page.goto('/monitoring');
      await page.waitForTimeout(1000);
      
      const monitoringState = await page.evaluate(() => ({
        timers: window.activeTimers.size,
        listeners: Array.from(window.activeListeners.entries())
          .reduce((total, [_, set]) => total + set.size, 0)
      }));
      
      expect(monitoringState.timers).toBeGreaterThan(0);
      expect(monitoringState.listeners).toBeGreaterThan(0);
      
      // Navigate away
      await page.goto('/dashboard');
      await page.waitForTimeout(1000);
      
      // Check cleanup
      const dashboardState = await page.evaluate(() => ({
        timers: window.activeTimers.size,
        listeners: Array.from(window.activeListeners.entries())
          .reduce((total, [_, set]) => total + set.size, 0)
      }));
      
      // Should have fewer active resources
      expect(dashboardState.timers).toBeLessThan(monitoringState.timers);
      expect(dashboardState.listeners).toBeLessThan(monitoringState.listeners);
    });
  });
});

// Import faker for test data
const faker = {
  lorem: {
    sentence: () => 'Lorem ipsum dolor sit amet, consectetur adipiscing elit.'
  }
};