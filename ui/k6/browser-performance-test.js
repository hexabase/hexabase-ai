import { chromium } from 'k6/experimental/browser';
import { check } from 'k6';
import { Trend } from 'k6/metrics';

// Custom metrics for Web Vitals
const firstContentfulPaint = new Trend('first_contentful_paint');
const largestContentfulPaint = new Trend('largest_contentful_paint');
const cumulativeLayoutShift = new Trend('cumulative_layout_shift');
const firstInputDelay = new Trend('first_input_delay');
const timeToInteractive = new Trend('time_to_interactive');
const totalBlockingTime = new Trend('total_blocking_time');

export const options = {
  scenarios: {
    browser: {
      executor: 'constant-vus',
      vus: 5,
      duration: '10m',
      options: {
        browser: {
          type: 'chromium',
        },
      },
    },
  },
  thresholds: {
    first_contentful_paint: ['p(75)<1800'],      // FCP < 1.8s for 75% of users
    largest_contentful_paint: ['p(75)<2500'],    // LCP < 2.5s for 75% of users
    cumulative_layout_shift: ['max<0.1'],        // CLS < 0.1
    first_input_delay: ['p(95)<100'],            // FID < 100ms for 95% of users
    time_to_interactive: ['p(90)<3500'],         // TTI < 3.5s for 90% of users
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

export default async function () {
  const browser = chromium.launch({
    headless: true,
    timeout: '60s',
  });
  
  const context = browser.newContext({
    viewport: { width: 1280, height: 720 },
    userAgent: 'k6-browser-test',
  });
  
  try {
    // Test different critical user journeys
    await testDashboardPerformance(context);
    await testApplicationDeployment(context);
    await testMonitoringDashboard(context);
    await testSearchPerformance(context);
  } finally {
    context.close();
    browser.close();
  }
}

async function testDashboardPerformance(context) {
  const page = context.newPage();
  
  try {
    // Measure Core Web Vitals
    await page.evaluateOnNewDocument(() => {
      window.webVitals = {
        FCP: 0,
        LCP: 0,
        CLS: 0,
        FID: 0,
        TTI: 0,
        TBT: 0,
      };
      
      // First Contentful Paint
      new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.name === 'first-contentful-paint') {
            window.webVitals.FCP = entry.startTime;
          }
        }
      }).observe({ entryTypes: ['paint'] });
      
      // Largest Contentful Paint
      new PerformanceObserver((list) => {
        const entries = list.getEntries();
        window.webVitals.LCP = entries[entries.length - 1].startTime;
      }).observe({ entryTypes: ['largest-contentful-paint'] });
      
      // Cumulative Layout Shift
      let clsValue = 0;
      new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (!entry.hadRecentInput) {
            clsValue += entry.value;
            window.webVitals.CLS = clsValue;
          }
        }
      }).observe({ entryTypes: ['layout-shift'] });
      
      // First Input Delay (simulated)
      ['click', 'keydown', 'mousedown', 'pointerdown', 'touchstart'].forEach(type => {
        window.addEventListener(type, (event) => {
          if (!window.webVitals.FID) {
            const delay = performance.now() - event.timeStamp;
            window.webVitals.FID = delay;
          }
        }, { once: true, passive: true });
      });
    });
    
    const startTime = Date.now();
    const response = await page.goto(`${BASE_URL}/dashboard`, {
      waitUntil: 'networkidle',
    });
    
    // Check response
    check(response.status(), {
      'dashboard loads successfully': (status) => status === 200,
    });
    
    // Wait for page to be interactive
    await page.waitForSelector('[data-testid="dashboard-content"]', {
      state: 'visible',
      timeout: 10000,
    });
    
    // Collect Web Vitals
    const vitals = await page.evaluate(() => window.webVitals);
    
    // Record metrics
    firstContentfulPaint.add(vitals.FCP);
    largestContentfulPaint.add(vitals.LCP);
    cumulativeLayoutShift.add(vitals.CLS);
    
    // Measure Time to Interactive
    const tti = await page.evaluate(() => {
      return new Promise((resolve) => {
        if ('PerformanceObserver' in window) {
          new PerformanceObserver((list) => {
            const entry = list.getEntries()[0];
            resolve(entry.startTime);
          }).observe({ entryTypes: ['measure'] });
          
          // Simulate TTI measurement
          setTimeout(() => {
            performance.measure('TTI', 'navigationStart');
            resolve(performance.now());
          }, 100);
        } else {
          resolve(performance.now());
        }
      });
    });
    timeToInteractive.add(tti);
    
    // Test interactivity
    const menuButton = page.locator('[data-testid="menu-button"]');
    if (await menuButton.isVisible()) {
      await menuButton.click();
      
      // Measure FID
      const clickTime = Date.now();
      await page.waitForSelector('[data-testid="menu-dropdown"]', {
        state: 'visible',
        timeout: 1000,
      });
      const responseTime = Date.now() - clickTime;
      firstInputDelay.add(responseTime);
    }
    
    // Measure JavaScript execution impact
    const longTaskCount = await page.evaluate(() => {
      return performance.getEntriesByType('longtask').length;
    });
    
    check(longTaskCount, {
      'minimal long tasks': (count) => count < 5,
    });
    
  } finally {
    page.close();
  }
}

async function testApplicationDeployment(context) {
  const page = context.newPage();
  
  try {
    await page.goto(`${BASE_URL}/applications`, {
      waitUntil: 'networkidle',
    });
    
    // Measure deployment form performance
    const deployButton = page.locator('[data-testid="deploy-app-button"]');
    await deployButton.click();
    
    const dialogAppearTime = await page.evaluate(() => {
      const start = performance.now();
      return new Promise((resolve) => {
        const observer = new MutationObserver(() => {
          const dialog = document.querySelector('[role="dialog"]');
          if (dialog) {
            resolve(performance.now() - start);
            observer.disconnect();
          }
        });
        observer.observe(document.body, { childList: true, subtree: true });
      });
    });
    
    check(dialogAppearTime, {
      'deploy dialog appears quickly': (time) => time < 300,
    });
    
    // Test form responsiveness
    const nameInput = page.locator('[data-testid="app-name-input"]');
    const typeStartTime = Date.now();
    await nameInput.type('performance-test-app', { delay: 50 });
    const typeEndTime = Date.now();
    
    check(typeEndTime - typeStartTime, {
      'form input is responsive': (time) => time < 2000,
    });
    
    // Close dialog
    await page.locator('[data-testid="close-dialog"]').click();
    
  } finally {
    page.close();
  }
}

async function testMonitoringDashboard(context) {
  const page = context.newPage();
  
  try {
    await page.goto(`${BASE_URL}/monitoring`, {
      waitUntil: 'networkidle',
    });
    
    // Measure chart rendering performance
    const chartRenderTime = await page.evaluate(() => {
      const start = performance.now();
      return new Promise((resolve) => {
        const checkCharts = () => {
          const charts = document.querySelectorAll('[data-testid^="chart-"]');
          if (charts.length > 0) {
            resolve(performance.now() - start);
          } else {
            requestAnimationFrame(checkCharts);
          }
        };
        checkCharts();
      });
    });
    
    check(chartRenderTime, {
      'charts render within 2 seconds': (time) => time < 2000,
    });
    
    // Test real-time updates performance
    const initialMemory = await page.evaluate(() => {
      if ('memory' in performance) {
        return (performance as any).memory.usedJSHeapSize;
      }
      return 0;
    });
    
    // Simulate 30 seconds of real-time updates
    await page.waitForTimeout(30000);
    
    const finalMemory = await page.evaluate(() => {
      if ('memory' in performance) {
        return (performance as any).memory.usedJSHeapSize;
      }
      return 0;
    });
    
    const memoryIncrease = finalMemory - initialMemory;
    check(memoryIncrease, {
      'no significant memory leak': (increase) => increase < 10 * 1024 * 1024, // < 10MB
    });
    
  } finally {
    page.close();
  }
}

async function testSearchPerformance(context) {
  const page = context.newPage();
  
  try {
    await page.goto(`${BASE_URL}/applications`, {
      waitUntil: 'networkidle',
    });
    
    const searchInput = page.locator('[data-testid="search-input"]');
    
    // Measure search responsiveness
    const searchQueries = ['nginx', 'api', 'frontend', 'backend', 'database'];
    
    for (const query of searchQueries) {
      await searchInput.clear();
      
      const searchStartTime = Date.now();
      await searchInput.type(query, { delay: 100 });
      
      // Wait for search results
      await page.waitForFunction(
        () => {
          const results = document.querySelector('[data-testid="search-results"]');
          return results && !results.classList.contains('loading');
        },
        { timeout: 2000 }
      );
      
      const searchEndTime = Date.now();
      const searchTime = searchEndTime - searchStartTime;
      
      check(searchTime, {
        'search completes within 1 second': (time) => time < 1000,
      });
      
      // Check result rendering performance
      const resultCount = await page.locator('[data-testid^="search-result-"]').count();
      check(resultCount, {
        'search results render': (count) => count > 0,
      });
    }
    
  } finally {
    page.close();
  }
}

export function handleSummary(data) {
  const summary = {
    'Web Vitals Summary': {
      'First Contentful Paint (FCP)': {
        'p50': `${data.metrics.first_contentful_paint.values['p(50)']?.toFixed(0)}ms`,
        'p75': `${data.metrics.first_contentful_paint.values['p(75)']?.toFixed(0)}ms`,
        'p90': `${data.metrics.first_contentful_paint.values['p(90)']?.toFixed(0)}ms`,
      },
      'Largest Contentful Paint (LCP)': {
        'p50': `${data.metrics.largest_contentful_paint.values['p(50)']?.toFixed(0)}ms`,
        'p75': `${data.metrics.largest_contentful_paint.values['p(75)']?.toFixed(0)}ms`,
        'p90': `${data.metrics.largest_contentful_paint.values['p(90)']?.toFixed(0)}ms`,
      },
      'Cumulative Layout Shift (CLS)': {
        'max': data.metrics.cumulative_layout_shift.values.max?.toFixed(3),
        'avg': data.metrics.cumulative_layout_shift.values.avg?.toFixed(3),
      },
      'First Input Delay (FID)': {
        'p75': `${data.metrics.first_input_delay.values['p(75)']?.toFixed(0)}ms`,
        'p95': `${data.metrics.first_input_delay.values['p(95)']?.toFixed(0)}ms`,
      },
      'Time to Interactive (TTI)': {
        'p50': `${data.metrics.time_to_interactive.values['p(50)']?.toFixed(0)}ms`,
        'p90': `${data.metrics.time_to_interactive.values['p(90)']?.toFixed(0)}ms`,
      },
    },
  };
  
  return {
    'stdout': JSON.stringify(summary, null, 2),
    'browser-performance-results.json': JSON.stringify(data),
  };
}