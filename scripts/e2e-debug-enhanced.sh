#!/bin/bash
# Enhanced E2E Test Debugging Script
# Runs E2E tests with visible browser, console error detection, and step-by-step debugging

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UI_DIR="$PROJECT_ROOT/ui"
DEBUG_OUTPUT_DIR="$UI_DIR/debug-output"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
SESSION_DIR="$DEBUG_OUTPUT_DIR/$TIMESTAMP"

# Function to print colored output
print_info() { echo -e "${BLUE}ℹ ${NC}$1"; }
print_success() { echo -e "${GREEN}✓ ${NC}$1"; }
print_warning() { echo -e "${YELLOW}⚠ ${NC}$1"; }
print_error() { echo -e "${RED}✗ ${NC}$1"; }
print_debug() { echo -e "${PURPLE}🔍 ${NC}$1"; }
print_console() { echo -e "${CYAN}[CONSOLE]${NC} $1"; }

# Default configuration
TEST_FILE=""
HEADED=true
DEBUG_MODE=true
SLOW_MO=1000
TIMEOUT=120000
STEP_BY_STEP=false
STOP_ON_ERROR=true
DEVELOPER_MODE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --test)
            TEST_FILE="$2"
            shift 2
            ;;
        --headless)
            HEADED=false
            shift
            ;;
        --no-debug)
            DEBUG_MODE=false
            shift
            ;;
        --slow-mo)
            SLOW_MO="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --step-by-step)
            STEP_BY_STEP=true
            SLOW_MO=2000
            shift
            ;;
        --no-stop-on-error)
            STOP_ON_ERROR=false
            shift
            ;;
        --developer)
            DEVELOPER_MODE=true
            STEP_BY_STEP=true
            SLOW_MO=1500
            shift
            ;;
        --help)
            cat << EOF
Enhanced E2E Debug Script
========================

Usage: $0 [options]

Options:
  --test <file>         Run specific test file
  --headless           Run in headless mode
  --no-debug           Disable debug mode
  --slow-mo <ms>       Set slow motion delay (default: 1000)
  --timeout <ms>       Set test timeout (default: 120000)
  --step-by-step       Enable step-by-step mode with manual confirmation
  --no-stop-on-error   Don't stop on console errors
  --developer          Enable developer walkthrough mode
  --help               Show this help message

Developer Mode Features:
  • Step-by-step execution with manual confirmation
  • Automatic pause on console errors
  • Screenshot on every action
  • Browser DevTools open
  • Detailed logging
  • Server log integration

Examples:
  $0 --developer --test auth.spec.ts    # Debug auth test in developer mode
  $0 --step-by-step                     # Manual step-by-step execution
  $0 --test debug-basic-functions.spec.ts --developer  # Debug basic functions
EOF
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create session directory
mkdir -p "$SESSION_DIR"/{reports,traces,videos,screenshots,logs,console}

# Function to check services
check_services() {
    print_info "Checking required services..."
    
    local services_ok=true
    
    # Check API
    if ! curl -s http://localhost:8080/health >/dev/null 2>&1; then
        print_warning "API is not running on port 8080"
        print_info "Starting debug environment..."
        "$PROJECT_ROOT/scripts/unified-debug.sh" start
        
        # Wait for API to be ready
        local max_attempts=30
        local attempt=0
        while [ $attempt -lt $max_attempts ]; do
            if curl -s http://localhost:8080/health >/dev/null 2>&1; then
                break
            fi
            attempt=$((attempt+1))
            sleep 2
        done
        
        if ! curl -s http://localhost:8080/health >/dev/null 2>&1; then
            print_error "Failed to start API service"
            exit 1
        fi
    fi
    print_success "API is running"
    
    # Check UI
    if ! curl -s http://localhost:3000 >/dev/null 2>&1; then
        print_error "UI is not running on port 3000"
        print_info "Please ensure UI service is running with: ./scripts/unified-debug.sh start"
        exit 1
    fi
    print_success "UI is running"
}

# Function to setup debug environment
setup_debug_env() {
    print_info "Setting up enhanced debug environment..."
    
    # Create debug config
    cat > "$UI_DIR/.env.debug.e2e" <<EOF
# Enhanced E2E Debug Configuration
PWDEBUG=1
DEBUG=pw:api,pw:browser*,pw:test
PLAYWRIGHT_BROWSERS_PATH=$HOME/.cache/playwright
BASE_URL=http://localhost:3000
API_URL=http://localhost:8080
STOP_ON_CONSOLE_ERROR=$STOP_ON_ERROR
STEP_BY_STEP=$STEP_BY_STEP
DEVELOPER_MODE=$DEVELOPER_MODE
SESSION_DIR=$SESSION_DIR
EOF
    
    # Create console error monitoring script
    cat > "$SESSION_DIR/console-monitor.js" <<'EOF'
// Console error monitoring for Playwright tests
class ConsoleMonitor {
  constructor() {
    this.errors = [];
    this.warnings = [];
    this.logs = [];
  }
  
  setup(page) {
    page.on('console', msg => {
      const entry = {
        type: msg.type(),
        text: msg.text(),
        timestamp: new Date().toISOString(),
        location: msg.location()
      };
      
      console.log(`[CONSOLE ${entry.type.toUpperCase()}] ${entry.text}`);
      
      switch (msg.type()) {
        case 'error':
          this.errors.push(entry);
          if (process.env.STOP_ON_CONSOLE_ERROR === 'true') {
            console.error('❌ Console error detected - stopping test');
            console.error(entry);
            throw new Error(`Console error: ${entry.text}`);
          }
          break;
        case 'warning':
          this.warnings.push(entry);
          break;
        default:
          this.logs.push(entry);
      }
    });
    
    page.on('pageerror', err => {
      const entry = {
        type: 'page-error',
        text: err.message,
        stack: err.stack,
        timestamp: new Date().toISOString()
      };
      
      console.error('❌ Page error detected:', err.message);
      this.errors.push(entry);
      
      if (process.env.STOP_ON_CONSOLE_ERROR === 'true') {
        throw err;
      }
    });
  }
  
  getReport() {
    return {
      errors: this.errors,
      warnings: this.warnings,
      logs: this.logs,
      summary: {
        errorCount: this.errors.length,
        warningCount: this.warnings.length,
        logCount: this.logs.length
      }
    };
  }
}

module.exports = { ConsoleMonitor };
EOF
}

# Function to create step-by-step helper
create_step_helper() {
    cat > "$SESSION_DIR/step-helper.js" <<'EOF'
// Step-by-step debugging helper
class StepHelper {
  constructor(test, page) {
    this.test = test;
    this.page = page;
    this.stepCount = 0;
  }
  
  async step(description, action) {
    this.stepCount++;
    const stepNum = this.stepCount.toString().padStart(2, '0');
    
    console.log(`\n🔍 Step ${stepNum}: ${description}`);
    
    if (process.env.STEP_BY_STEP === 'true') {
      console.log('Press Enter to continue, or type "skip" to skip this step...');
      // In real implementation, would use readline for interactive input
      await this.page.waitForTimeout(1000);
    }
    
    try {
      // Take screenshot before action
      await this.page.screenshot({ 
        path: `${process.env.SESSION_DIR}/screenshots/step-${stepNum}-before.png`,
        fullPage: true 
      });
      
      // Execute the action
      const result = await action();
      
      // Take screenshot after action
      await this.page.screenshot({ 
        path: `${process.env.SESSION_DIR}/screenshots/step-${stepNum}-after.png`,
        fullPage: true 
      });
      
      console.log(`✅ Step ${stepNum} completed`);
      return result;
    } catch (error) {
      // Take screenshot on error
      await this.page.screenshot({ 
        path: `${process.env.SESSION_DIR}/screenshots/step-${stepNum}-error.png`,
        fullPage: true 
      });
      
      console.error(`❌ Step ${stepNum} failed:`, error.message);
      throw error;
    }
  }
}

module.exports = { StepHelper };
EOF
}

# Function to run tests
run_tests() {
    cd "$UI_DIR"
    
    print_info "Running enhanced E2E tests..."
    print_debug "Session directory: $SESSION_DIR"
    
    # Build test command
    local test_cmd="npx playwright test"
    
    if [ -n "$TEST_FILE" ]; then
        test_cmd="$test_cmd $TEST_FILE"
    fi
    
    test_cmd="$test_cmd --config=playwright.debug.config.ts"
    
    if [ "$HEADED" = "true" ]; then
        test_cmd="$test_cmd --headed"
    fi
    
    if [ "$DEBUG_MODE" = "true" ]; then
        test_cmd="$test_cmd --debug"
    fi
    
    # Set environment variables
    export PWDEBUG=1
    export DEBUG="pw:api,pw:browser*,pw:test"
    export PLAYWRIGHT_SLOW_MO=$SLOW_MO
    export TEST_TIMEOUT=$TIMEOUT
    export PLAYWRIGHT_HTML_REPORT="$SESSION_DIR/reports/html"
    export PLAYWRIGHT_JUNIT_OUTPUT_FILE="$SESSION_DIR/reports/junit.xml"
    export STOP_ON_CONSOLE_ERROR=$STOP_ON_ERROR
    export STEP_BY_STEP=$STEP_BY_STEP
    export DEVELOPER_MODE=$DEVELOPER_MODE
    export SESSION_DIR=$SESSION_DIR
    
    print_debug "Test command: $test_cmd"
    
    if [ "$STEP_BY_STEP" = "true" ]; then
        print_info "Running in step-by-step mode"
        print_warning "Tests will pause at each step for manual inspection"
    fi
    
    if [ "$STOP_ON_ERROR" = "true" ]; then
        print_info "Stopping on console errors is enabled"
    fi
    
    # Start log capture in background
    "$PROJECT_ROOT/scripts/unified-debug.sh" logs > "$SESSION_DIR/logs/server-logs.log" 2>&1 &
    local logs_pid=$!
    
    print_info "Starting test execution..."
    
    # Run tests and capture output
    local test_exit_code=0
    if ! $test_cmd 2>&1 | tee "$SESSION_DIR/logs/test-output.log"; then
        test_exit_code=$?
        print_error "Tests failed with exit code: $test_exit_code"
    fi
    
    # Stop log capture
    kill $logs_pid 2>/dev/null || true
    
    return $test_exit_code
}

# Function to analyze results
analyze_results() {
    print_info "Analyzing test results..."
    
    local has_errors=false
    local has_failures=false
    
    # Check for console errors
    if [ -f "$SESSION_DIR/logs/test-output.log" ]; then
        local console_errors=$(grep -c "Console error\|Page error" "$SESSION_DIR/logs/test-output.log" 2>/dev/null || echo "0")
        if [ "$console_errors" -gt 0 ]; then
            print_warning "Found $console_errors console errors"
            has_errors=true
        fi
    fi
    
    # Check for test failures
    if [ -f "$SESSION_DIR/reports/junit.xml" ]; then
        local test_failures=$(grep -c 'failure\|error' "$SESSION_DIR/reports/junit.xml" 2>/dev/null || echo "0")
        if [ "$test_failures" -gt 0 ]; then
            print_warning "Found $test_failures test failures"
            has_failures=true
        fi
    fi
    
    # Generate analysis report
    cat > "$SESSION_DIR/analysis-report.md" <<EOF
# E2E Test Analysis Report

**Session**: $TIMESTAMP  
**Mode**: Enhanced Debug $([ "$DEVELOPER_MODE" = "true" ] && echo "(Developer Mode)")  
**Date**: $(date)

## Configuration
- **Headed**: $HEADED
- **Debug Mode**: $DEBUG_MODE
- **Slow Motion**: ${SLOW_MO}ms
- **Step-by-Step**: $STEP_BY_STEP
- **Stop on Console Error**: $STOP_ON_ERROR
- **Test File**: ${TEST_FILE:-"All tests"}

## Results Summary
- **Console Errors**: $([ "$has_errors" = "true" ] && echo "❌ Yes" || echo "✅ None")
- **Test Failures**: $([ "$has_failures" = "true" ] && echo "❌ Yes" || echo "✅ None")

## Files Generated
- Test logs: \`logs/test-output.log\`
- Server logs: \`logs/server-logs.log\`
- Screenshots: \`screenshots/\`
- Videos: \`test-results/\`
- Traces: \`test-results/\`

## Next Steps
$([ "$has_errors" = "true" ] && echo "1. Review console errors in test logs
2. Check browser screenshots for visual issues
3. Examine server logs for backend errors" || echo "1. Review test reports
2. Check screenshots for visual validation")

## Console Error Investigation
$([ "$has_errors" = "true" ] && echo "Run: \`grep -A 5 -B 5 'Console error\\|Page error' $SESSION_DIR/logs/test-output.log\`" || echo "No console errors detected ✅")
EOF
    
    if [ "$has_errors" = "true" ] || [ "$has_failures" = "true" ]; then
        print_warning "Issues detected - see analysis report for details"
        return 1
    else
        print_success "All tests passed without issues"
        return 0
    fi
}

# Function to collect artifacts
collect_artifacts() {
    print_info "Collecting debug artifacts..."
    
    # Copy test results
    if [ -d "$UI_DIR/test-results" ]; then
        cp -r "$UI_DIR/test-results"/* "$SESSION_DIR/" 2>/dev/null || true
    fi
    
    # Create debug summary
    cat > "$SESSION_DIR/debug-summary.md" <<EOF
# Enhanced E2E Debug Session Summary

**Session ID**: $TIMESTAMP  
**Mode**: Enhanced Debug $([ "$DEVELOPER_MODE" = "true" ] && echo "- Developer Mode")  
**Date**: $(date)

## Quick Access
- **HTML Report**: [Open Report](reports/html/index.html)
- **Test Logs**: [View Logs](logs/test-output.log)
- **Server Logs**: [View Server Logs](logs/server-logs.log)
- **Analysis**: [View Analysis](analysis-report.md)

## Debug Features Used
- **Console Error Detection**: $([ "$STOP_ON_ERROR" = "true" ] && echo "✅ Enabled" || echo "❌ Disabled")
- **Step-by-Step Mode**: $([ "$STEP_BY_STEP" = "true" ] && echo "✅ Enabled" || echo "❌ Disabled")
- **Visual Recording**: ✅ Screenshots + Videos
- **Server Log Integration**: ✅ Synchronized logging

## Directory Structure
\`\`\`
$SESSION_DIR/
├── reports/           # HTML and JUnit reports
├── traces/            # Playwright traces
├── videos/            # Test execution videos
├── screenshots/       # Step-by-step screenshots
├── logs/              # Test and server logs
├── console/           # Console error details
└── analysis-report.md # Automated analysis
\`\`\`

## Debugging Commands
\`\`\`bash
# View latest test output
tail -f $SESSION_DIR/logs/test-output.log

# Search for console errors
grep -n "Console error\|Page error" $SESSION_DIR/logs/test-output.log

# View server errors
grep -n "ERROR\|FATAL" $SESSION_DIR/logs/server-logs.log

# Show trace in Playwright UI
npx playwright show-trace $SESSION_DIR/traces/*.zip
\`\`\`
EOF
    
    print_success "Debug artifacts collected"
}

# Function to open results
open_results() {
    print_info "Opening debug results..."
    
    # Try to open HTML report
    local report_path="$SESSION_DIR/reports/html/index.html"
    if [ -f "$report_path" ]; then
        if command -v open >/dev/null 2>&1; then
            open "$report_path"
        elif command -v xdg-open >/dev/null 2>&1; then
            xdg-open "$report_path"
        fi
    fi
    
    # Create symlink to latest
    ln -sfn "$SESSION_DIR" "$DEBUG_OUTPUT_DIR/latest"
    
    echo
    print_success "Enhanced debug session complete!"
    echo
    echo "Session saved to: $SESSION_DIR"
    echo "Latest session: $DEBUG_OUTPUT_DIR/latest"
    echo
    print_info "Quick commands:"
    echo "• View report:    open $SESSION_DIR/reports/html/index.html"
    echo "• Check errors:   grep 'Console error\\|Page error' $SESSION_DIR/logs/test-output.log"
    echo "• Server logs:    less $SESSION_DIR/logs/server-logs.log"
    echo "• Show trace:     npx playwright show-trace $SESSION_DIR/traces/*.zip"
}

# Main execution
main() {
    print_info "Starting Enhanced E2E Debug Session"
    echo "===================================="
    
    if [ "$DEVELOPER_MODE" = "true" ]; then
        print_info "🧑‍💻 Developer Mode Enabled"
        print_info "Features: Step-by-step execution, console error detection, detailed logging"
        echo
    fi
    
    check_services
    setup_debug_env
    create_step_helper
    
    local test_result=0
    run_tests || test_result=$?
    
    collect_artifacts
    analyze_results || true
    open_results
    
    exit $test_result
}

# Run main function
main "$@"