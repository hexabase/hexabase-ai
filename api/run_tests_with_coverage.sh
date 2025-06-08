#!/bin/bash

# Test runner script with coverage and organized output
# Creates timestamped test results in testresults directory

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_DIR="testresults"
UNIT_DIR="$RESULTS_DIR/unit"
COVERAGE_DIR="$RESULTS_DIR/coverage"
SUMMARY_DIR="$RESULTS_DIR/summary"
LOGS_DIR="$RESULTS_DIR/logs"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting comprehensive test run at $TIMESTAMP${NC}"
echo "Results will be saved to $RESULTS_DIR"

# Create timestamp subdirectories
mkdir -p "$UNIT_DIR/$TIMESTAMP"
mkdir -p "$COVERAGE_DIR/$TIMESTAMP"
mkdir -p "$LOGS_DIR/$TIMESTAMP"

# Function to run tests for a package
run_package_tests() {
    local pkg=$1
    local pkg_name=$(echo $pkg | sed 's/\//_/g' | sed 's/github.com_hexabase_hexabase-ai_api_//')
    
    echo -e "\n${YELLOW}Testing package: $pkg${NC}"
    
    # Run tests with verbose output and coverage
    go test -v -coverprofile="$COVERAGE_DIR/$TIMESTAMP/${pkg_name}.coverage" \
        -covermode=atomic \
        -json \
        "$pkg" > "$UNIT_DIR/$TIMESTAMP/${pkg_name}.json" 2> "$LOGS_DIR/$TIMESTAMP/${pkg_name}.log"
    
    local test_status=$?
    
    # Also save human-readable output
    go test -v -coverprofile="$COVERAGE_DIR/$TIMESTAMP/${pkg_name}.coverage" \
        "$pkg" > "$UNIT_DIR/$TIMESTAMP/${pkg_name}.txt" 2>&1
    
    return $test_status
}

# Get all packages with tests
PACKAGES=$(go list ./... | grep -v /vendor/)

# Initialize summary
SUMMARY_FILE="$SUMMARY_DIR/test_summary_$TIMESTAMP.md"
cat > "$SUMMARY_FILE" << EOF
# Test Summary Report
**Generated at:** $(date)
**Timestamp:** $TIMESTAMP

## Test Results by Package

| Package | Status | Tests | Coverage | Time |
|---------|--------|-------|----------|------|
EOF

# Track overall status
TOTAL_PACKAGES=0
PASSED_PACKAGES=0
FAILED_PACKAGES=0
FAILED_PACKAGE_LIST=""

# Run tests for each package
for pkg in $PACKAGES; do
    TOTAL_PACKAGES=$((TOTAL_PACKAGES + 1))
    
    if run_package_tests "$pkg"; then
        PASSED_PACKAGES=$((PASSED_PACKAGES + 1))
        status="✅ PASS"
    else
        FAILED_PACKAGES=$((FAILED_PACKAGES + 1))
        FAILED_PACKAGE_LIST="$FAILED_PACKAGE_LIST\n- $pkg"
        status="❌ FAIL"
    fi
    
    # Extract test summary from output
    pkg_name=$(echo $pkg | sed 's/\//_/g' | sed 's/github.com_hexabase_hexabase-ai_api_//')
    if [ -f "$UNIT_DIR/$TIMESTAMP/${pkg_name}.txt" ]; then
        # Extract test count and coverage
        test_info=$(grep -E "(PASS|FAIL)" "$UNIT_DIR/$TIMESTAMP/${pkg_name}.txt" | tail -1)
        coverage=$(grep -E "coverage:" "$UNIT_DIR/$TIMESTAMP/${pkg_name}.txt" | grep -oE "[0-9]+\.[0-9]+%" | head -1)
        test_time=$(echo "$test_info" | grep -oE "[0-9]+\.[0-9]+s" | head -1)
        test_count=$(grep -cE "(RUN|PASS|FAIL)" "$UNIT_DIR/$TIMESTAMP/${pkg_name}.txt" | head -1)
        
        [ -z "$coverage" ] && coverage="N/A"
        [ -z "$test_time" ] && test_time="N/A"
        [ -z "$test_count" ] && test_count="0"
        
        # Add to summary
        echo "| \`$pkg\` | $status | $test_count | $coverage | $test_time |" >> "$SUMMARY_FILE"
    fi
done

# Generate overall coverage report
echo -e "\n${YELLOW}Generating overall coverage report...${NC}"
go test -coverprofile="$COVERAGE_DIR/$TIMESTAMP/coverage.out" -covermode=atomic ./... > "$LOGS_DIR/$TIMESTAMP/coverage_generation.log" 2>&1

# Convert to HTML
go tool cover -html="$COVERAGE_DIR/$TIMESTAMP/coverage.out" -o "$COVERAGE_DIR/$TIMESTAMP/coverage.html" 2> "$LOGS_DIR/$TIMESTAMP/coverage_html.log"

# Generate coverage summary
go tool cover -func="$COVERAGE_DIR/$TIMESTAMP/coverage.out" > "$COVERAGE_DIR/$TIMESTAMP/coverage_summary.txt" 2> "$LOGS_DIR/$TIMESTAMP/coverage_summary.log"

# Extract total coverage
TOTAL_COVERAGE=$(tail -1 "$COVERAGE_DIR/$TIMESTAMP/coverage_summary.txt" | grep -oE "[0-9]+\.[0-9]+%" || echo "N/A")

# Complete summary
cat >> "$SUMMARY_FILE" << EOF

## Overall Statistics

- **Total Packages Tested:** $TOTAL_PACKAGES
- **Passed:** $PASSED_PACKAGES
- **Failed:** $FAILED_PACKAGES
- **Overall Coverage:** $TOTAL_COVERAGE

EOF

if [ $FAILED_PACKAGES -gt 0 ]; then
    cat >> "$SUMMARY_FILE" << EOF
## Failed Packages
$FAILED_PACKAGE_LIST

EOF
fi

# Add coverage details
cat >> "$SUMMARY_FILE" << EOF
## Coverage by Package

\`\`\`
$(cat "$COVERAGE_DIR/$TIMESTAMP/coverage_summary.txt")
\`\`\`

## File Locations

- **Unit Test Results:** \`$UNIT_DIR/$TIMESTAMP/\`
- **Coverage Reports:** \`$COVERAGE_DIR/$TIMESTAMP/\`
- **Test Logs:** \`$LOGS_DIR/$TIMESTAMP/\`
- **HTML Coverage Report:** \`$COVERAGE_DIR/$TIMESTAMP/coverage.html\`

EOF

# Create latest symlinks
ln -sf "$SUMMARY_FILE" "$SUMMARY_DIR/latest.md"
ln -sf "$COVERAGE_DIR/$TIMESTAMP/coverage.html" "$COVERAGE_DIR/latest.html"

# Display summary
echo -e "\n${GREEN}=== Test Run Complete ===${NC}"
echo -e "Total Packages: $TOTAL_PACKAGES"
echo -e "Passed: ${GREEN}$PASSED_PACKAGES${NC}"
echo -e "Failed: ${RED}$FAILED_PACKAGES${NC}"
echo -e "Coverage: ${YELLOW}$TOTAL_COVERAGE${NC}"
echo -e "\nDetailed summary: ${YELLOW}$SUMMARY_FILE${NC}"

# Exit with failure if any tests failed
if [ $FAILED_PACKAGES -gt 0 ]; then
    echo -e "\n${RED}Some tests failed. Please check the logs for details.${NC}"
    exit 1
else
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
fi