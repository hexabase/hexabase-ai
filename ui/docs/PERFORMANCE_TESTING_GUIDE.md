# Performance Testing Guide

This guide covers performance and load testing strategies for the Hexabase AI platform.

## Overview

Performance testing ensures the platform can handle expected user loads while maintaining acceptable response times and resource usage.

## Testing Types

### 1. Frontend Performance Tests (Playwright)

Located in `e2e/tests/performance-load.spec.ts`

#### Page Load Performance
- Measures Core Web Vitals (FCP, LCP, CLS)
- Tests under different network conditions
- Verifies performance budgets

```bash
npm run test:e2e performance-load.spec.ts -- --grep "Page Load Performance"
```

#### UI Responsiveness
- Tests interaction performance under load
- Measures response to rapid user actions
- Monitors real-time update handling

#### Memory Management
- Detects memory leaks during long sessions
- Tracks resource cleanup
- Monitors garbage collection impact

### 2. API Load Testing (k6)

Located in `k6/load-test.js`

#### Test Scenarios

1. **Smoke Test** (1 minute, 2 users)
   - Verify basic functionality
   - Baseline performance metrics

2. **Load Test** (9 minutes, 50 users)
   - Normal expected load
   - Sustained performance verification

3. **Stress Test** (20 minutes, up to 300 users)
   - Find breaking points
   - Performance degradation patterns

4. **Spike Test** (2 minutes, 100 users sudden)
   - Sudden traffic increases
   - Recovery behavior

#### Running k6 Tests

```bash
# Install k6
brew install k6  # macOS
# or download from https://k6.io/docs/getting-started/installation/

# Run all scenarios
k6 run k6/load-test.js

# Run specific scenario
k6 run k6/load-test.js --env SCENARIO=load

# With custom URL and token
k6 run k6/load-test.js --env BASE_URL=https://staging.example.com --env API_TOKEN=your-token

# Generate HTML report
k6 run k6/load-test.js --out html=report.html
```

### 3. Browser Performance Testing (k6 Browser)

Located in `k6/browser-performance-test.js`

Measures real browser metrics:
- First Contentful Paint (FCP)
- Largest Contentful Paint (LCP)
- Cumulative Layout Shift (CLS)
- First Input Delay (FID)
- Time to Interactive (TTI)

```bash
# Run browser performance tests
k6 run k6/browser-performance-test.js

# With specific browser
k6 run k6/browser-performance-test.js --env BROWSER=firefox
```

## Performance Targets

### Frontend Metrics

| Metric | Good | Needs Improvement | Poor |
|--------|------|-------------------|------|
| FCP | < 1.8s | 1.8s - 3s | > 3s |
| LCP | < 2.5s | 2.5s - 4s | > 4s |
| CLS | < 0.1 | 0.1 - 0.25 | > 0.25 |
| FID | < 100ms | 100ms - 300ms | > 300ms |
| TTI | < 3.5s | 3.5s - 7.5s | > 7.5s |

### API Response Times

| Endpoint Type | Target p95 | Maximum |
|--------------|------------|---------|
| Read operations | < 200ms | 500ms |
| Write operations | < 500ms | 1000ms |
| Dashboard stats | < 300ms | 800ms |
| Real-time metrics | < 100ms | 300ms |

### Load Capacity

| Resource | Expected | Target | Maximum |
|----------|----------|--------|---------|
| Concurrent users | 100 | 500 | 1000 |
| Requests/second | 200 | 1000 | 2000 |
| WebSocket connections | 50 | 200 | 500 |

## Performance Optimization Checklist

### Frontend Optimizations

- [ ] Code splitting for routes
- [ ] Lazy loading for components
- [ ] Image optimization (WebP, responsive)
- [ ] Bundle size analysis
- [ ] Tree shaking unused code
- [ ] Service worker caching
- [ ] Prefetching critical resources

### Backend Optimizations

- [ ] Database query optimization
- [ ] Redis caching strategy
- [ ] Connection pooling
- [ ] Response compression
- [ ] Rate limiting
- [ ] Horizontal scaling ready
- [ ] CDN for static assets

### Monitoring Setup

- [ ] Real User Monitoring (RUM)
- [ ] Application Performance Monitoring (APM)
- [ ] Custom performance metrics
- [ ] Alert thresholds configured
- [ ] Performance dashboards

## Debugging Performance Issues

### Frontend Debugging

1. **Chrome DevTools Performance Tab**
   ```javascript
   // Record performance trace
   performance.mark('myFeature-start');
   // ... feature code ...
   performance.mark('myFeature-end');
   performance.measure('myFeature', 'myFeature-start', 'myFeature-end');
   ```

2. **React DevTools Profiler**
   - Identify unnecessary re-renders
   - Find expensive components
   - Optimize React.memo usage

3. **Bundle Analysis**
   ```bash
   npm run build -- --analyze
   ```

### Backend Debugging

1. **Database Query Analysis**
   ```sql
   EXPLAIN ANALYZE SELECT ...;
   ```

2. **API Response Time Logging**
   ```javascript
   app.use((req, res, next) => {
     const start = Date.now();
     res.on('finish', () => {
       const duration = Date.now() - start;
       console.log(`${req.method} ${req.url} - ${duration}ms`);
     });
     next();
   });
   ```

3. **Memory Profiling**
   ```bash
   node --inspect app.js
   # Open chrome://inspect
   ```

## Continuous Performance Testing

### CI Integration

```yaml
# .github/workflows/performance.yml
name: Performance Tests
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  workflow_dispatch:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run k6 Load Test
        uses: k6io/action@v0.1
        with:
          filename: k6/load-test.js
          
      - name: Upload Results
        uses: actions/upload-artifact@v3
        with:
          name: k6-results
          path: load-test-results.json
```

### Performance Regression Detection

1. **Lighthouse CI**
   ```bash
   npm install -g @lhci/cli
   lhci autorun
   ```

2. **Custom Metrics Tracking**
   ```javascript
   // Track custom business metrics
   performance.mark('checkout-flow-start');
   // ... checkout process ...
   performance.mark('checkout-flow-end');
   performance.measure('checkout-flow', 'checkout-flow-start', 'checkout-flow-end');
   
   // Send to analytics
   analytics.track('performance', {
     metric: 'checkout-flow',
     duration: performance.getEntriesByName('checkout-flow')[0].duration
   });
   ```

## Best Practices

### 1. Test Early and Often
- Include performance tests in PR checks
- Run nightly performance regression tests
- Monitor production performance continuously

### 2. Realistic Test Data
- Use production-like data volumes
- Test with realistic user behaviors
- Include geographic distribution

### 3. Incremental Improvements
- Set achievable targets
- Focus on user-impacting metrics
- Celebrate small wins

### 4. Documentation
- Document performance budgets
- Track optimization history
- Share learnings with team

## Troubleshooting

### Common Issues

1. **Flaky Performance Tests**
   - Use percentiles instead of averages
   - Increase sample size
   - Account for system warm-up

2. **Environment Differences**
   - Use consistent hardware
   - Control background processes
   - Match production configurations

3. **Unrealistic Load Patterns**
   - Analyze actual user behavior
   - Use production access logs
   - Implement think time in tests

### Getting Help

- Performance testing channel: #performance-testing
- k6 documentation: https://k6.io/docs/
- Playwright performance: https://playwright.dev/docs/test-advanced#performance-testing