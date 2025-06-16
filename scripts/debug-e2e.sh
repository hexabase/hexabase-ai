#!/bin/bash
# E2E Test Debugging Script
# Runs E2E tests with enhanced debugging capabilities

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UI_DIR="$PROJECT_ROOT/ui"
DEBUG_OUTPUT_DIR="$UI_DIR/debug-output"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Function to print colored output
print_info() { echo -e "${BLUE}ℹ ${NC}$1"; }
print_success() { echo -e "${GREEN}✓ ${NC}$1"; }
print_warning() { echo -e "${YELLOW}⚠ ${NC}$1"; }
print_error() { echo -e "${RED}✗ ${NC}$1"; }
print_debug() { echo -e "${PURPLE}🔍 ${NC}$1"; }

# Parse arguments
TEST_FILE=""
HEADED=true
DEBUG_MODE=true
SLOW_MO=500
TIMEOUT=120000

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
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --test <file>    Run specific test file"
            echo "  --headless       Run in headless mode"
            echo "  --no-debug       Disable debug mode"
            echo "  --slow-mo <ms>   Set slow motion delay (default: 500)"
            echo "  --timeout <ms>   Set test timeout (default: 120000)"
            echo "  --help           Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create debug output directory
mkdir -p "$DEBUG_OUTPUT_DIR/$TIMESTAMP"

# Function to check services
check_services() {
    print_info "Checking required services..."
    
    local services_ok=true
    
    # Check API
    if ! curl -s http://localhost:8080/health >/dev/null 2>&1; then
        print_warning "API is not running on port 8080"
        services_ok=false
    else
        print_success "API is running"
    fi
    
    # Check UI
    if ! curl -s http://localhost:3000 >/dev/null 2>&1; then
        print_warning "UI is not running on port 3000"
        services_ok=false
    else
        print_success "UI is running"
    fi
    
    if [ "$services_ok" = false ]; then
        print_warning "Some services are not running"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Function to setup debug environment
setup_debug_env() {
    print_info "Setting up debug environment..."
    
    # Create debug config
    cat > "$UI_DIR/.env.debug" <<EOF
# E2E Debug Configuration
PWDEBUG=1
DEBUG=pw:api,pw:browser*
PLAYWRIGHT_BROWSERS_PATH=$HOME/.cache/playwright
BASE_URL=http://localhost:3000
API_URL=http://localhost:8080
EOF
    
    # Create test report directory
    mkdir -p "$DEBUG_OUTPUT_DIR/$TIMESTAMP/reports"
    mkdir -p "$DEBUG_OUTPUT_DIR/$TIMESTAMP/traces"
    mkdir -p "$DEBUG_OUTPUT_DIR/$TIMESTAMP/videos"
    mkdir -p "$DEBUG_OUTPUT_DIR/$TIMESTAMP/screenshots"
    mkdir -p "$DEBUG_OUTPUT_DIR/$TIMESTAMP/logs"
}

# Function to run tests
run_tests() {
    cd "$UI_DIR"
    
    print_info "Running E2E tests in debug mode..."
    print_debug "Output directory: $DEBUG_OUTPUT_DIR/$TIMESTAMP"
    
    # Build test command
    local test_cmd="npx playwright test"
    
    if [ -n "$TEST_FILE" ]; then
        test_cmd="$test_cmd $TEST_FILE"
    fi
    
    test_cmd="$test_cmd --config=playwright.debug.config.ts"
    
    if [ "$HEADED" = true ]; then
        test_cmd="$test_cmd --headed"
    fi
    
    if [ "$DEBUG_MODE" = true ]; then
        test_cmd="$test_cmd --debug"
    fi
    
    # Set environment variables
    export PWDEBUG=1
    export DEBUG="pw:api,pw:browser*"
    export PLAYWRIGHT_SLOW_MO=$SLOW_MO
    export TEST_TIMEOUT=$TIMEOUT
    export PLAYWRIGHT_HTML_REPORT="$DEBUG_OUTPUT_DIR/$TIMESTAMP/reports/html"
    export PLAYWRIGHT_JUNIT_OUTPUT_FILE="$DEBUG_OUTPUT_DIR/$TIMESTAMP/reports/junit.xml"
    
    print_debug "Running command: $test_cmd"
    
    # Run tests and capture output
    if $test_cmd 2>&1 | tee "$DEBUG_OUTPUT_DIR/$TIMESTAMP/logs/test-output.log"; then
        print_success "Tests completed successfully"
    else
        print_error "Tests failed"
    fi
}

# Function to collect debug artifacts
collect_artifacts() {
    print_info "Collecting debug artifacts..."
    
    # Copy test results if they exist
    if [ -d "$UI_DIR/test-results" ]; then
        cp -r "$UI_DIR/test-results"/* "$DEBUG_OUTPUT_DIR/$TIMESTAMP/" 2>/dev/null || true
    fi
    
    # Generate summary report
    cat > "$DEBUG_OUTPUT_DIR/$TIMESTAMP/debug-summary.md" <<EOF
# E2E Debug Session Summary

**Date**: $(date)
**Mode**: $([ "$HEADED" = true ] && echo "Headed" || echo "Headless")
**Debug**: $([ "$DEBUG_MODE" = true ] && echo "Enabled" || echo "Disabled")
**Slow Motion**: ${SLOW_MO}ms
**Timeout**: ${TIMEOUT}ms
**Test File**: ${TEST_FILE:-"All tests"}

## Output Directory Structure

\`\`\`
$DEBUG_OUTPUT_DIR/$TIMESTAMP/
├── reports/         # Test reports
├── traces/          # Playwright traces
├── videos/          # Test videos
├── screenshots/     # Test screenshots
└── logs/            # Test logs
\`\`\`

## Viewing Results

1. **HTML Report**: Open \`reports/html/index.html\` in a browser
2. **Traces**: Use \`npx playwright show-trace <trace-file>\`
3. **Videos**: Check the \`videos/\` directory
4. **Screenshots**: Check the \`screenshots/\` directory

## Debugging Tips

1. Use the Playwright Inspector for step-by-step debugging
2. Set breakpoints in test code using \`await page.pause()\`
3. Use \`--debug\` flag to enable Playwright Inspector
4. Check browser DevTools console for errors
5. Review network activity in DevTools Network tab

EOF
    
    print_success "Debug artifacts collected in: $DEBUG_OUTPUT_DIR/$TIMESTAMP"
}

# Function to open results
open_results() {
    print_info "Opening debug results..."
    
    # Try to open HTML report
    local report_path="$DEBUG_OUTPUT_DIR/$TIMESTAMP/reports/html/index.html"
    if [ -f "$report_path" ]; then
        if command -v open >/dev/null 2>&1; then
            open "$report_path"
        elif command -v xdg-open >/dev/null 2>&1; then
            xdg-open "$report_path"
        else
            print_info "Report available at: file://$report_path"
        fi
    fi
    
    # Show summary
    echo
    print_success "Debug session complete!"
    echo
    echo "Results saved to: $DEBUG_OUTPUT_DIR/$TIMESTAMP"
    echo
    echo "Quick Commands:"
    echo "==============="
    echo "• View HTML report:  open $DEBUG_OUTPUT_DIR/$TIMESTAMP/reports/html/index.html"
    echo "• Show trace:        npx playwright show-trace $DEBUG_OUTPUT_DIR/$TIMESTAMP/traces/*.zip"
    echo "• View logs:         less $DEBUG_OUTPUT_DIR/$TIMESTAMP/logs/test-output.log"
    echo
    
    # Create symlink to latest debug session
    ln -sfn "$DEBUG_OUTPUT_DIR/$TIMESTAMP" "$DEBUG_OUTPUT_DIR/latest"
    print_info "Latest debug session linked to: $DEBUG_OUTPUT_DIR/latest"
}

# Main execution
main() {
    print_info "Starting E2E Debug Session"
    echo "=========================="
    
    check_services
    setup_debug_env
    run_tests
    collect_artifacts
    open_results
}

# Run main function
main "$@"