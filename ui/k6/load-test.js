import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const apiErrors = new Counter('api_errors');
const apiSuccessRate = new Rate('api_success_rate');
const apiResponseTime = new Trend('api_response_time');

// Test configuration
export const options = {
  scenarios: {
    // Smoke test - verify system works under minimal load
    smoke: {
      executor: 'constant-vus',
      vus: 2,
      duration: '1m',
    },
    
    // Load test - assess performance under typical load
    load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },   // Ramp up
        { duration: '5m', target: 50 },   // Stay at 50 users
        { duration: '2m', target: 0 },    // Ramp down
      ],
      startTime: '2m', // Start after smoke test
    },
    
    // Stress test - find breaking point
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 100 },
        { duration: '3m', target: 100 },
        { duration: '2m', target: 200 },
        { duration: '3m', target: 200 },
        { duration: '2m', target: 300 },
        { duration: '3m', target: 300 },
        { duration: '5m', target: 0 },
      ],
      startTime: '12m', // Start after load test
    },
    
    // Spike test - sudden traffic increase
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 100 },  // Spike to 100 users
        { duration: '1m', target: 100 },   // Stay at 100
        { duration: '10s', target: 0 },    // Back to 0
      ],
      startTime: '30m', // Start after stress test
    },
  },
  
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% of requests under 500ms
    http_req_failed: ['rate<0.1'],                   // Error rate under 10%
    api_success_rate: ['rate>0.9'],                  // API success rate over 90%
    api_response_time: ['p(95)<400'],                // API 95th percentile under 400ms
  },
};

// Test data
const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const API_TOKEN = __ENV.API_TOKEN || 'test-token';

// Helper function to make API calls
function apiCall(endpoint, options = {}) {
  const url = `${BASE_URL}/api${endpoint}`;
  const params = {
    headers: {
      'Authorization': `Bearer ${API_TOKEN}`,
      'Content-Type': 'application/json',
    },
    ...options,
  };
  
  const response = http.request(options.method || 'GET', url, options.body || null, params);
  
  // Track metrics
  apiResponseTime.add(response.timings.duration);
  const success = check(response, {
    'status is 200-299': (r) => r.status >= 200 && r.status < 300,
  });
  
  apiSuccessRate.add(success);
  if (!success) {
    apiErrors.add(1);
  }
  
  return response;
}

// User scenarios
export default function () {
  const scenario = Math.random();
  
  if (scenario < 0.4) {
    // 40% - Dashboard viewing
    dashboardScenario();
  } else if (scenario < 0.7) {
    // 30% - Application management
    applicationScenario();
  } else if (scenario < 0.9) {
    // 20% - Monitoring
    monitoringScenario();
  } else {
    // 10% - Heavy operations
    heavyOperationsScenario();
  }
}

function dashboardScenario() {
  // Load dashboard
  let response = apiCall('/dashboard/stats');
  check(response, {
    'dashboard loads': (r) => r.status === 200,
  });
  
  sleep(2); // User views dashboard
  
  // Load recent activities
  response = apiCall('/activities?limit=10');
  check(response, {
    'activities load': (r) => r.status === 200,
  });
  
  sleep(1);
  
  // Check notifications
  response = apiCall('/notifications/unread');
  check(response, {
    'notifications load': (r) => r.status === 200,
  });
}

function applicationScenario() {
  // List applications
  let response = apiCall('/applications?page=1&limit=20');
  check(response, {
    'applications list': (r) => r.status === 200,
  });
  
  sleep(1);
  
  // View application details
  response = apiCall('/applications/app-123');
  check(response, {
    'application details': (r) => r.status === 200,
  });
  
  sleep(2);
  
  // Check application metrics
  response = apiCall('/applications/app-123/metrics?period=1h');
  check(response, {
    'application metrics': (r) => r.status === 200,
  });
  
  sleep(1);
  
  // Update application (10% chance)
  if (Math.random() < 0.1) {
    response = apiCall('/applications/app-123', {
      method: 'PATCH',
      body: JSON.stringify({
        replicas: Math.floor(Math.random() * 5) + 1,
      }),
    });
    check(response, {
      'application update': (r) => r.status === 200,
    });
  }
}

function monitoringScenario() {
  // Load monitoring dashboard
  let response = apiCall('/monitoring/overview');
  check(response, {
    'monitoring overview': (r) => r.status === 200,
  });
  
  sleep(1);
  
  // Get resource metrics
  response = apiCall('/metrics/resources?interval=5m&duration=1h');
  check(response, {
    'resource metrics': (r) => r.status === 200,
  });
  
  sleep(2);
  
  // Check alerts
  response = apiCall('/alerts/active');
  check(response, {
    'active alerts': (r) => r.status === 200,
  });
  
  // Simulate real-time updates (WebSocket would be used in real app)
  for (let i = 0; i < 5; i++) {
    sleep(2);
    response = apiCall('/metrics/resources/latest');
    check(response, {
      'real-time metrics': (r) => r.status === 200,
    });
  }
}

function heavyOperationsScenario() {
  // Deploy new application
  let response = apiCall('/applications', {
    method: 'POST',
    body: JSON.stringify({
      name: `load-test-app-${Date.now()}`,
      image: 'nginx:latest',
      replicas: 2,
      resources: {
        requests: { cpu: '100m', memory: '128Mi' },
        limits: { cpu: '200m', memory: '256Mi' },
      },
    }),
  });
  
  const deployed = check(response, {
    'application deployed': (r) => r.status === 201,
  });
  
  if (deployed) {
    const appId = JSON.parse(response.body).id;
    sleep(5);
    
    // Check deployment status
    response = apiCall(`/applications/${appId}/deployment/status`);
    check(response, {
      'deployment status': (r) => r.status === 200,
    });
    
    sleep(10);
    
    // Clean up - delete application
    response = apiCall(`/applications/${appId}`, {
      method: 'DELETE',
    });
    check(response, {
      'application deleted': (r) => r.status === 204,
    });
  }
}

// Lifecycle hooks
export function setup() {
  // Verify system is ready
  const response = http.get(`${BASE_URL}/health`);
  check(response, {
    'system is healthy': (r) => r.status === 200,
  });
  
  if (response.status !== 200) {
    throw new Error('System is not healthy, aborting test');
  }
  
  return { startTime: Date.now() };
}

export function teardown(data) {
  // Could perform cleanup here
  console.log(`Test completed in ${Date.now() - data.startTime}ms`);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'load-test-results.json': JSON.stringify(data),
    'load-test-results.html': htmlReport(data),
  };
}

// Generate HTML report
function htmlReport(data) {
  return `
<!DOCTYPE html>
<html>
<head>
    <title>Load Test Results - ${new Date().toISOString()}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        .metric { margin: 10px 0; padding: 10px; background: #f5f5f5; }
        .passed { color: green; }
        .failed { color: red; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
    </style>
</head>
<body>
    <h1>Hexabase AI Load Test Results</h1>
    <p>Test completed at: ${new Date().toISOString()}</p>
    
    <h2>Summary</h2>
    <div class="metric">
        <strong>Total Requests:</strong> ${data.metrics.http_reqs.values.count}
    </div>
    <div class="metric">
        <strong>Failed Requests:</strong> ${data.metrics.http_req_failed.values.passes} 
        (${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%)
    </div>
    <div class="metric">
        <strong>Average Response Time:</strong> ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms
    </div>
    <div class="metric">
        <strong>95th Percentile:</strong> ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms
    </div>
    
    <h2>Thresholds</h2>
    <table>
        <tr>
            <th>Metric</th>
            <th>Threshold</th>
            <th>Result</th>
            <th>Status</th>
        </tr>
        ${Object.entries(data.metrics).map(([name, metric]) => {
            if (metric.thresholds) {
                return Object.entries(metric.thresholds).map(([threshold, passed]) => `
                    <tr>
                        <td>${name}</td>
                        <td>${threshold}</td>
                        <td>${metric.values[threshold.match(/p\((\d+)\)/)?.[1] || 'rate']}</td>
                        <td class="${passed ? 'passed' : 'failed'}">${passed ? 'PASSED' : 'FAILED'}</td>
                    </tr>
                `).join('');
            }
            return '';
        }).join('')}
    </table>
    
    <h2>Scenarios</h2>
    ${Object.entries(options.scenarios).map(([name, scenario]) => `
        <div class="metric">
            <h3>${name}</h3>
            <p>${JSON.stringify(scenario, null, 2)}</p>
        </div>
    `).join('')}
</body>
</html>
  `;
}

function textSummary(data, options) {
  // Simple text summary
  return `
Load Test Summary
=================
Total Requests: ${data.metrics.http_reqs.values.count}
Failed Requests: ${data.metrics.http_req_failed.values.passes} (${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%)
Avg Response Time: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms
95th Percentile: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms
99th Percentile: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms
  `;
}