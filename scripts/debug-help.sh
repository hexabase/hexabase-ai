#!/bin/bash
# Debug Help Script
# Shows available debugging commands and usage examples

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}🔧 Hexabase AI Enhanced Debugging Guide${NC}"
echo "=========================================="
echo

echo -e "${GREEN}Quick Start Commands:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${CYAN}make debug${NC}                    # Start unified debug environment"
echo -e "${CYAN}make debug-e2e-dev${NC}           # Run E2E tests in developer mode"
echo -e "${CYAN}make debug-basic${NC}              # Run debug basic functions test"
echo -e "${CYAN}make debug-logs${NC}               # Stream color-coded logs"
echo -e "${CYAN}make debug-stop${NC}               # Stop debug environment"
echo

echo -e "${GREEN}Debug Environment Management:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${PURPLE}./scripts/unified-debug.sh start${NC}      # Start all services"
echo -e "${PURPLE}./scripts/unified-debug.sh status${NC}     # Check service health"
echo -e "${PURPLE}./scripts/unified-debug.sh logs${NC}       # Stream all logs"
echo -e "${PURPLE}./scripts/unified-debug.sh logs api${NC}   # Stream API logs only"
echo -e "${PURPLE}./scripts/unified-debug.sh restart${NC}    # Restart all services"
echo -e "${PURPLE}./scripts/unified-debug.sh restart api${NC} # Restart API only"
echo

echo -e "${GREEN}Enhanced E2E Testing:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${YELLOW}Developer Mode (Recommended):${NC}"
echo -e "${PURPLE}./scripts/e2e-debug-enhanced.sh --developer${NC}"
echo "  • Visual browser with slow motion"
echo "  • Automatic pause on console errors"
echo "  • Screenshots on every action"
echo "  • Step-by-step execution"
echo

echo -e "${YELLOW}Specific Test Files:${NC}"
echo -e "${PURPLE}./scripts/e2e-debug-enhanced.sh --developer --test auth.spec.ts${NC}"
echo -e "${PURPLE}./scripts/e2e-debug-enhanced.sh --developer --test debug-basic-functions.spec.ts${NC}"
echo

echo -e "${YELLOW}Custom Options:${NC}"
echo -e "${PURPLE}./scripts/e2e-debug-enhanced.sh --step-by-step --slow-mo 2000${NC}"
echo -e "${PURPLE}./scripts/e2e-debug-enhanced.sh --no-stop-on-error --headless${NC}"
echo

echo -e "${GREEN}Service Debugging:${NC}"
echo "━━━━━━━━━━━━━━━━━━━"
echo -e "${YELLOW}API Debugging (Go + Delve):${NC}"
echo "  • Debug port: localhost:2345"
echo "  • VSCode: Use 'Debug API (Docker)' configuration"  
echo "  • Command line: dlv connect localhost:2345"
echo

echo -e "${YELLOW}UI Debugging (Node.js Inspector):${NC}"
echo "  • Debug port: localhost:9229"
echo "  • Chrome DevTools: chrome://inspect"
echo "  • VSCode: Use 'Debug UI (Next.js)' configuration"
echo

echo -e "${GREEN}Monitoring URLs:${NC}"
echo "━━━━━━━━━━━━━━━━━"
echo "  • API:        http://localhost:8080"
echo "  • UI:         http://localhost:3000"
echo "  • Jaeger:     http://localhost:16686"
echo "  • Prometheus: http://localhost:9090"
echo "  • Grafana:    http://localhost:3001"
echo

echo -e "${GREEN}Debug Session Analysis:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${PURPLE}# View latest test report${NC}"
echo "open ui/debug-output/latest/reports/html/index.html"
echo

echo -e "${PURPLE}# Check for console errors${NC}"
echo "grep 'Console error\\|Page error' ui/debug-output/latest/logs/test-output.log"
echo

echo -e "${PURPLE}# View server logs${NC}"
echo "less ui/debug-output/latest/logs/server-logs.log"
echo

echo -e "${PURPLE}# Show Playwright trace${NC}"
echo "npx playwright show-trace ui/debug-output/latest/traces/*.zip"
echo

echo -e "${GREEN}Common Workflows:${NC}"
echo "━━━━━━━━━━━━━━━━━━"
echo -e "${YELLOW}1. Basic Development Debugging:${NC}"
echo "   make debug                    # Start environment"
echo "   make debug-basic              # Test basic functions"
echo "   make debug-logs               # Monitor logs"
echo

echo -e "${YELLOW}2. Feature Development:${NC}"
echo "   make debug                    # Start environment"
echo "   # Develop your feature"
echo "   make debug-e2e-dev            # Test with visual browser"
echo "   # Fix any console errors found"
echo

echo -e "${YELLOW}3. Bug Investigation:${NC}"
echo "   make debug                    # Start environment"
echo "   ./scripts/e2e-debug-enhanced.sh --developer --step-by-step"
echo "   # Manual walkthrough with screenshots"
echo "   # Review debug artifacts in ui/debug-output/latest/"
echo

echo -e "${YELLOW}4. Console Error Debugging:${NC}"
echo "   make debug                    # Start environment"
echo "   ./scripts/e2e-debug-enhanced.sh --developer --test your-test.spec.ts"
echo "   # Test automatically stops on console errors"
echo "   # Check ui/debug-output/latest/console/ for error details"
echo

echo -e "${GREEN}Documentation:${NC}"
echo "━━━━━━━━━━━━━━━"
echo "  • Full guide: docs/DEBUGGING.md"
echo "  • Architecture: docs/architecture/"
echo "  • Project structure: STRUCTURE_GUIDE.md"
echo

echo -e "${GREEN}Troubleshooting:${NC}"
echo "━━━━━━━━━━━━━━━━"
echo -e "${YELLOW}Services not starting:${NC}"
echo "  make debug-status             # Check service health"
echo "  make debug-stop && make debug # Restart everything"
echo

echo -e "${YELLOW}Port conflicts:${NC}"
echo "  lsof -i :8080                 # Check what's using ports"
echo "  make debug-stop               # Stop services"
echo

echo -e "${YELLOW}Tests failing:${NC}"
echo "  make debug-status             # Ensure services are healthy"  
echo "  grep ERROR ui/debug-output/latest/logs/server-logs.log"
echo

echo -e "${BLUE}💡 Pro Tips:${NC}"
echo "━━━━━━━━━━━"
echo "• Use developer mode for visual debugging"
echo "• Check console errors first when tests fail"
echo "• Screenshots are saved for every test step"
echo "• Server logs are synchronized with test execution"
echo "• Use step-by-step mode for complex issues"
echo

echo -e "${GREEN}Need help? Check:${NC}"
echo "  • docs/DEBUGGING.md for detailed documentation"
echo "  • ui/debug-output/latest/ for latest test artifacts"
echo "  • make help for all available commands"