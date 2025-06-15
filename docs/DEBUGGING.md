# Enhanced Debugging Guide

This guide covers the enhanced debugging capabilities for Hexabase AI, providing unified Docker-based debugging with visual E2E testing and console error detection.

## Quick Start

### 1. Start Debug Environment
```bash
# Start all services with debug capabilities
make debug

# Or use the unified script directly
./scripts/unified-debug.sh start
```

### 2. Run Enhanced E2E Tests
```bash
# Run tests in developer mode with visual browser
make debug-e2e-dev

# Run specific basic functions test
make debug-basic

# Run with custom options
./scripts/e2e-debug-enhanced.sh --developer --test auth.spec.ts
```

### 3. Monitor and Debug
```bash
# Stream all service logs with color coding
make debug-logs

# Check debug environment status
make debug-status

# Stop debug environment
make debug-stop
```

## Debug Environment Features

### Unified Debug Script (`unified-debug.sh`)

The unified debug script provides a single command interface for managing the entire debug environment:

**Key Features:**
- **Single Command Management**: Start, stop, restart all services
- **Hot Reload**: Source code mounted into Docker containers
- **Color-Coded Logging**: Different colors for API, UI, and error logs
- **Health Monitoring**: Automatic service health checks
- **Port Management**: Automatic port conflict detection

**Usage:**
```bash
./scripts/unified-debug.sh [command] [options]

Commands:
  start               Start all debug services
  stop                Stop all debug services  
  restart [service]   Restart all or specific service
  status              Show service status
  logs [service]      Show logs (all or specific service)
  attach <service>    Show debugger connection info
```

**Services Available:**
- `api` - Go API server with Delve debugger (port 2345)
- `ui` - Next.js UI with Node.js debugger (port 9229)
- `postgres` - PostgreSQL with query logging
- `redis` - Redis with debug logging
- `nats` - NATS with verbose logging

### Enhanced E2E Testing (`e2e-debug-enhanced.sh`)

Advanced E2E testing with visual debugging and error detection:

**Developer Mode Features:**
- **Visual Browser**: Tests run in headed mode with slow motion
- **Console Error Detection**: Automatically pauses on browser console errors
- **Step-by-Step Execution**: Manual confirmation between test steps
- **Screenshot Capture**: Screenshots before/after each action
- **Server Log Integration**: Synchronized logging with E2E test execution
- **Debug Artifacts**: Comprehensive debug output collection

**Usage:**
```bash
./scripts/e2e-debug-enhanced.sh [options]

Options:
  --test <file>         Run specific test file
  --developer           Enable developer walkthrough mode
  --step-by-step        Manual step-by-step execution
  --slow-mo <ms>        Set slow motion delay (default: 1000ms)
  --no-stop-on-error    Don't pause on console errors
  --headless            Run without visible browser
```

**Examples:**
```bash
# Developer walkthrough of auth tests
./scripts/e2e-debug-enhanced.sh --developer --test auth.spec.ts

# Step-by-step execution with manual confirmation
./scripts/e2e-debug-enhanced.sh --step-by-step

# Debug basic functions with all features enabled
./scripts/e2e-debug-enhanced.sh --developer --test debug-basic-functions.spec.ts
```

## Debug Test Suite

### Basic Functions Test (`debug-basic-functions.spec.ts`)

A specialized test suite designed for debugging core functionality:

**Test Coverage:**
1. **Login Flow** - Complete authentication walkthrough
2. **Navigation** - Menu and navigation testing  
3. **Form Validation** - Error handling and validation
4. **API Integration** - Network request monitoring

**Features:**
- Console error detection and reporting
- Visual element verification
- API request/response monitoring
- Detailed logging at each step
- Developer account integration

**Run the test:**
```bash
make debug-basic

# Or with custom options
./scripts/e2e-debug-enhanced.sh --developer --test debug-basic-functions.spec.ts
```

## Debug Helper Utilities

### DebugHelper Class

The `DebugHelper` class provides enhanced debugging capabilities for E2E tests:

```typescript
import { DebugHelper } from '../utils/debug-helpers';

test('My debug test', async ({ page }, testInfo) => {
  const debugHelper = new DebugHelper(page, testInfo);
  
  // Step-by-step execution with screenshots
  await debugHelper.step('Navigate to login', async () => {
    await debugHelper.goto('http://localhost:3000');
  });
  
  // Enhanced clicking with debugging
  await debugHelper.click('[data-testid="login-button"]');
  
  // Form filling with validation
  await debugHelper.fill('[data-testid="email"]', 'user@example.com');
  
  // Generate debug report
  const report = debugHelper.generateDebugReport();
});
```

**Features:**
- Automatic screenshot capture
- Console error monitoring
- Page state logging
- Network request tracking
- Debug artifact generation

## Service Debugging

### API Debugging (Go with Delve)

**Connect with IDE:**
1. Start debug environment: `make debug`
2. In VSCode: Use "Debug API (Docker)" launch configuration
3. Or connect manually: `dlv connect localhost:2345`

**Available Debug Ports:**
- API Debugger: `localhost:2345`
- Worker Debugger: `localhost:2346`

### UI Debugging (Node.js Inspector)

**Connect with Chrome DevTools:**
1. Start debug environment: `make debug`
2. Open Chrome and go to: `chrome://inspect`
3. Click "Open dedicated DevTools for Node"

**Available Debug Port:**
- UI Debugger: `localhost:9229`

### Database Debugging

**PostgreSQL Query Logging:**
All SQL queries are logged with timing information when running in debug mode.

**Access Database:**
```bash
# Database shell
make db-shell

# Or connect directly
docker compose -f docker-compose.yml -f docker-compose.debug.yml exec postgres psql -U postgres -d hexabase
```

## Monitoring and Observability

### Debug Dashboard URLs

When debug environment is running, access these URLs:

- **API**: http://localhost:8080
- **UI**: http://localhost:3000  
- **Jaeger Tracing**: http://localhost:16686
- **Prometheus Metrics**: http://localhost:9090
- **Grafana Dashboard**: http://localhost:3001

### Log Management

**Stream All Logs:**
```bash
make debug-logs
```

**Service-Specific Logs:**
```bash
./scripts/unified-debug.sh logs api    # API logs only
./scripts/unified-debug.sh logs ui     # UI logs only
```

**Log Features:**
- Color-coded output by service
- Error highlighting
- Real-time streaming
- Structured log format

## Debug Session Management

### Session Artifacts

Each debug session creates a timestamped directory with:

```
debug-output/YYYYMMDD_HHMMSS/
├── reports/           # HTML and JUnit test reports
├── traces/            # Playwright traces  
├── videos/            # Test execution videos
├── screenshots/       # Step-by-step screenshots
├── logs/              # Test and server logs
├── console/           # Console error details
└── analysis-report.md # Automated analysis
```

### Session Analysis

**View Latest Session:**
```bash
# Latest session is symlinked for easy access
open debug-output/latest/reports/html/index.html

# Check for console errors
grep "Console error" debug-output/latest/logs/test-output.log

# View server logs
less debug-output/latest/logs/server-logs.log
```

**Analyze Test Traces:**
```bash
npx playwright show-trace debug-output/latest/traces/*.zip
```

## Best Practices

### 1. Developer Walkthrough Process

1. **Start Environment**: `make debug`
2. **Run Basic Functions**: `make debug-basic`  
3. **Check for Issues**: Review console errors and screenshots
4. **Iterate**: Fix issues and re-run tests
5. **Validate**: Run full test suite when issues are resolved

### 2. Console Error Investigation

When console errors are detected:

1. **Review Error Details**: Check `debug-output/latest/console/` directory
2. **Check Screenshots**: Look at before/after screenshots  
3. **Server Logs**: Check corresponding server logs for backend errors
4. **Network Tab**: Use browser DevTools to inspect network requests

### 3. Step-by-Step Debugging

For complex issues:

1. **Enable Developer Mode**: `--developer` flag
2. **Use Step-by-Step**: Manual confirmation between steps
3. **Inspect Elements**: Browser DevTools open automatically
4. **Check State**: Page state logged at each step

### 4. Performance Debugging

Monitor performance issues:

1. **Slow Motion**: Adjust `--slow-mo` for better visibility
2. **Network Monitoring**: Check API request/response times
3. **Resource Usage**: Monitor Docker container resources
4. **Traces**: Use Playwright traces for detailed timing

## Troubleshooting

### Common Issues

**Port Conflicts:**
```bash
# Check what's using a port
lsof -i :8080

# Stop debug environment and restart
make debug-stop && make debug
```

**Service Not Starting:**
```bash
# Check service status
make debug-status

# View service logs
./scripts/unified-debug.sh logs [service]

# Restart specific service
./scripts/unified-debug.sh restart [service]
```

**Console Errors Not Detected:**
- Ensure `--stop-on-error` is enabled
- Check browser console manually
- Verify debug helper is properly initialized

**Tests Timing Out:**
- Increase timeout with `--timeout` option
- Check if services are healthy with `make debug-status`
- Review server logs for backend issues

### Debug Environment Reset

If you encounter persistent issues:

```bash
# Stop everything
make debug-stop

# Clean Docker resources  
docker system prune -f

# Restart debug environment
make debug
```

## Advanced Configuration

### Environment Variables

Debug behavior can be customized with environment variables:

```bash
# Enable step-by-step mode
export STEP_BY_STEP=true

# Disable stopping on console errors
export STOP_ON_CONSOLE_ERROR=false

# Set custom slow motion delay
export PLAYWRIGHT_SLOW_MO=2000

# Enable developer mode
export DEVELOPER_MODE=true
```

### Custom Test Data

For debugging with specific data:

```typescript
// In your test file
import { testUsers } from '../fixtures/mock-data';

// Use developer account for consistent testing
const developerUser = testUsers.developer || {
  email: 'dev@hexabase.com', 
  password: 'dev123456'
};
```

This debugging setup provides comprehensive tools for developing and troubleshooting Hexabase AI with visual feedback, detailed logging, and automated error detection.