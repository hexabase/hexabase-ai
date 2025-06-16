#!/bin/bash
# Unified Debug Environment Management Script
# Provides single command to start/restart/monitor both API and UI servers

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
COMPOSE_PROJECT_NAME="hexabase-debug"
LOG_DIR="$PROJECT_ROOT/logs/debug"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# State tracking
PID_FILE="$PROJECT_ROOT/.debug-pids"
STATE_FILE="$PROJECT_ROOT/.debug-state"

# Function to print colored output
print_info() { echo -e "${BLUE}ℹ ${NC}$1"; }
print_success() { echo -e "${GREEN}✓ ${NC}$1"; }
print_warning() { echo -e "${YELLOW}⚠ ${NC}$1"; }
print_error() { echo -e "${RED}✗ ${NC}$1"; }
print_debug() { echo -e "${PURPLE}🔍 ${NC}$1"; }
print_api() { echo -e "${CYAN}[API]${NC} $1"; }
print_ui() { echo -e "${PURPLE}[UI]${NC} $1"; }

# Function to get docker compose command
get_docker_compose_cmd() {
    if command -v docker-compose >/dev/null 2>&1; then
        echo "docker-compose"
    elif docker compose version >/dev/null 2>&1; then
        echo "docker compose"
    else
        print_error "Docker Compose not found"
        exit 1
    fi
}

# Function to check if services are running
check_services() {
    local api_running=false
    local ui_running=false
    
    if curl -s http://localhost:8080/health >/dev/null 2>&1; then
        api_running=true
    fi
    
    if curl -s http://localhost:3000 >/dev/null 2>&1; then
        ui_running=true
    fi
    
    echo "$api_running:$ui_running"
}

# Function to start services
start_services() {
    print_info "Starting Hexabase debug environment..."
    
    # Create log directory
    mkdir -p "$LOG_DIR"
    
    # Create debug environment file if not exists
    if [ ! -f "$PROJECT_ROOT/.env.debug" ]; then
        print_info "Creating debug environment configuration..."
        cp "$PROJECT_ROOT/.env.debug.example" "$PROJECT_ROOT/.env.debug"
    fi
    
    # Export environment variables
    export COMPOSE_PROJECT_NAME=$COMPOSE_PROJECT_NAME
    
    # Build and start services
    DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
    
    print_info "Building debug images..."
    $DOCKER_COMPOSE_CMD -f docker-compose.yml -f docker-compose.debug.yml --env-file .env.debug build
    
    print_info "Starting services..."
    $DOCKER_COMPOSE_CMD -f docker-compose.yml -f docker-compose.debug.yml --env-file .env.debug up -d
    
    # Save state
    echo "running" > "$STATE_FILE"
    
    # Wait for services
    print_info "Waiting for services to be ready..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        local status=$(check_services)
        IFS=':' read -r api_ok ui_ok <<< "$status"
        
        if [ "$api_ok" = "true" ] && [ "$ui_ok" = "true" ]; then
            print_success "All services are ready!"
            break
        fi
        
        attempt=$((attempt+1))
        sleep 2
        echo -n "."
    done
    echo
    
    # Show status
    show_status
}

# Function to stop services
stop_services() {
    print_info "Stopping services..."
    
    DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
    $DOCKER_COMPOSE_CMD -f docker-compose.yml -f docker-compose.debug.yml --env-file .env.debug down
    
    echo "stopped" > "$STATE_FILE"
    print_success "Services stopped"
}

# Function to restart services
restart_services() {
    print_info "Restarting services..."
    
    # Restart specific service or all
    local service=$1
    DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
    
    if [ -n "$service" ]; then
        print_info "Restarting $service..."
        $DOCKER_COMPOSE_CMD -f docker-compose.yml -f docker-compose.debug.yml --env-file .env.debug restart $service
    else
        print_info "Restarting all services..."
        $DOCKER_COMPOSE_CMD -f docker-compose.yml -f docker-compose.debug.yml --env-file .env.debug restart
    fi
    
    print_success "Services restarted"
}

# Function to show logs with color coding
show_logs() {
    local service=$1
    local follow=${2:-true}
    
    DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
    
    if [ -n "$service" ]; then
        # Show logs for specific service
        if [ "$follow" = "true" ]; then
            $DOCKER_COMPOSE_CMD -f docker-compose.yml -f docker-compose.debug.yml logs -f $service
        else
            $DOCKER_COMPOSE_CMD -f docker-compose.yml -f docker-compose.debug.yml logs --tail=100 $service
        fi
    else
        # Show logs for all services with color coding
        print_info "Streaming logs from all services..."
        print_info "Press Ctrl+C to stop"
        echo
        
        # Start log streaming in background with color coding
        $DOCKER_COMPOSE_CMD -f docker-compose.yml -f docker-compose.debug.yml logs -f | while IFS= read -r line; do
            if [[ $line == *"api"* ]] || [[ $line == *"api_1"* ]] || [[ $line == *"api-1"* ]]; then
                print_api "$line"
            elif [[ $line == *"ui"* ]] || [[ $line == *"ui_1"* ]] || [[ $line == *"ui-1"* ]]; then
                print_ui "$line"
            elif [[ $line == *"ERROR"* ]] || [[ $line == *"error"* ]]; then
                echo -e "${RED}$line${NC}"
            elif [[ $line == *"WARN"* ]] || [[ $line == *"warning"* ]]; then
                echo -e "${YELLOW}$line${NC}"
            elif [[ $line == *"DEBUG"* ]] || [[ $line == *"debug"* ]]; then
                echo -e "${PURPLE}$line${NC}"
            else
                echo "$line"
            fi
        done
    fi
}

# Function to show status
show_status() {
    print_info "Service Status:"
    echo "==============="
    
    DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
    
    # Check container status
    $DOCKER_COMPOSE_CMD -f docker-compose.yml -f docker-compose.debug.yml ps
    
    echo
    print_info "Service Health:"
    echo "==============="
    
    local status=$(check_services)
    IFS=':' read -r api_ok ui_ok <<< "$status"
    
    if [ "$api_ok" = "true" ]; then
        print_success "API: Running at http://localhost:8080"
    else
        print_error "API: Not responding"
    fi
    
    if [ "$ui_ok" = "true" ]; then
        print_success "UI: Running at http://localhost:3000"
    else
        print_error "UI: Not responding"
    fi
    
    echo
    print_info "Debug Endpoints:"
    echo "================"
    echo "• API Debugger:    localhost:2345"
    echo "• UI Debugger:     localhost:9229"
    echo "• Jaeger UI:       http://localhost:16686"
    echo "• Prometheus:      http://localhost:9090"
    echo "• Grafana:         http://localhost:3001"
}

# Function to attach to service
attach_to_service() {
    local service=$1
    
    DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
    
    case $service in
        api)
            print_info "Attaching to API debugger..."
            echo "Run this in your debugger: dlv connect localhost:2345"
            ;;
        ui)
            print_info "Attaching to UI debugger..."
            echo "Chrome DevTools URL: chrome://inspect"
            echo "Or use VSCode with 'Debug UI (Next.js)' configuration"
            ;;
        *)
            print_error "Unknown service: $service"
            echo "Available services: api, ui"
            ;;
    esac
}

# Function to show help
show_help() {
    echo "Hexabase Unified Debug Script"
    echo "============================="
    echo
    echo "Usage: $0 [command] [options]"
    echo
    echo "Commands:"
    echo "  start               Start all debug services"
    echo "  stop                Stop all debug services"
    echo "  restart [service]   Restart all services or specific service"
    echo "  status              Show service status"
    echo "  logs [service]      Show logs (all services or specific)"
    echo "  attach <service>    Show how to attach debugger to service"
    echo "  help                Show this help message"
    echo
    echo "Services:"
    echo "  api                 API server with Delve debugger"
    echo "  ui                  UI server with Node.js debugger"
    echo "  postgres            PostgreSQL database"
    echo "  redis               Redis cache"
    echo "  nats                NATS message queue"
    echo
    echo "Examples:"
    echo "  $0 start            # Start all services"
    echo "  $0 logs             # Show all logs"
    echo "  $0 logs api         # Show only API logs"
    echo "  $0 restart api      # Restart only API service"
    echo "  $0 attach api       # Show how to debug API"
}

# Main command handling
case "${1:-start}" in
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        restart_services "$2"
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "$2"
        ;;
    attach)
        if [ -z "$2" ]; then
            print_error "Please specify a service to attach to"
            exit 1
        fi
        attach_to_service "$2"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac