#!/bin/bash
# Hexabase KaaS Development Environment Cleanup Script
# This script removes all development environment resources

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${YELLOW}🧹 Hexabase KaaS Development Environment Cleanup${NC}"
echo -e "${YELLOW}===============================================${NC}"
echo -e "${RED}⚠️  This will remove all local development resources!${NC}"
echo -n "Are you sure you want to continue? (y/N): "
read -r response

if [[ ! "$response" =~ ^[Yy]$ ]]; then
    echo "Cleanup cancelled."
    exit 0
fi

# Stop Docker Compose services
echo -e "\n${YELLOW}▶ Stopping Docker Compose services...${NC}"
cd "$PROJECT_ROOT"
docker-compose down -v
echo -e "${GREEN}✓ Docker services stopped${NC}"

# Delete Kind cluster
echo -e "\n${YELLOW}▶ Deleting Kind cluster...${NC}"
if kind get clusters | grep -q "hexabase-dev"; then
    kind delete cluster --name hexabase-dev
    echo -e "${GREEN}✓ Kind cluster deleted${NC}"
else
    echo "Kind cluster 'hexabase-dev' not found"
fi

# Remove generated files
echo -e "\n${YELLOW}▶ Removing generated files...${NC}"

# Remove .env files
rm -f "$PROJECT_ROOT/api/.env"
rm -f "$PROJECT_ROOT/ui/.env.local"

# Remove JWT keys
rm -rf "$PROJECT_ROOT/api/keys"

# Remove docker-compose override
rm -f "$PROJECT_ROOT/docker-compose.override.yml"

echo -e "${GREEN}✓ Generated files removed${NC}"

# Clean up /etc/hosts (optional)
echo -e "\n${YELLOW}▶ Clean up /etc/hosts entries?${NC}"
echo -n "Remove Hexabase development entries from /etc/hosts? (y/N): "
read -r response

if [[ "$response" =~ ^[Yy]$ ]]; then
    echo "Removing entries from /etc/hosts (requires sudo)..."
    sudo sed -i.bak '/# Hexabase KaaS Development/,+2d' /etc/hosts
    echo -e "${GREEN}✓ /etc/hosts cleaned${NC}"
fi

echo -e "\n${GREEN}✨ Cleanup complete!${NC}"
echo -e "\nTo set up the development environment again, run:"
echo -e "  ${YELLOW}./scripts/dev-setup.sh${NC}"