#!/bin/bash
# Hexabase KaaS Development Environment Setup Script
# This script automates the setup of a local development environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
K8S_CONTEXT="kind-hexabase-dev"
NAMESPACE="hexabase-dev"

echo -e "${GREEN}🚀 Hexabase KaaS Development Environment Setup${NC}"
echo -e "${GREEN}============================================${NC}"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print step
print_step() {
    echo -e "\n${YELLOW}▶ $1${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}✗ $1${NC}"
    exit 1
}

# Function to get docker compose command
get_docker_compose_cmd() {
    if command_exists "docker-compose"; then
        echo "docker-compose"
    elif docker compose version >/dev/null 2>&1; then
        echo "docker compose"
    else
        print_error "Neither 'docker-compose' nor 'docker compose' command found. Please install Docker Compose."
    fi
}

# Function to detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}

# Function to install dependency
install_dependency() {
    local dep=$1
    local os=$(detect_os)
    
    case $dep in
        "kind")
            echo "Installing kind..."
            if [[ "$os" == "macos" ]]; then
                if command_exists brew; then
                    brew install kind
                else
                    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-darwin-$(uname -m)
                    chmod +x ./kind
                    sudo mv ./kind /usr/local/bin/kind
                fi
            elif [[ "$os" == "linux" ]]; then
                curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
                chmod +x ./kind
                sudo mv ./kind /usr/local/bin/kind
            fi
            ;;
        "helm")
            echo "Installing helm..."
            if [[ "$os" == "macos" ]]; then
                if command_exists brew; then
                    brew install helm
                else
                    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
                fi
            elif [[ "$os" == "linux" ]]; then
                curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
            fi
            ;;
        "kubectl")
            echo "Installing kubectl..."
            if [[ "$os" == "macos" ]]; then
                if command_exists brew; then
                    brew install kubectl
                else
                    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/$(uname -m)/kubectl"
                    chmod +x kubectl
                    sudo mv kubectl /usr/local/bin/
                fi
            elif [[ "$os" == "linux" ]]; then
                curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
                chmod +x kubectl
                sudo mv kubectl /usr/local/bin/
            fi
            ;;
    esac
}

# Check prerequisites
print_step "Checking prerequisites..."

MISSING_DEPS=()
AUTO_INSTALLABLE=("kind" "helm" "kubectl")

if ! command_exists docker; then
    MISSING_DEPS+=("docker")
fi

if ! command_exists go; then
    MISSING_DEPS+=("go")
fi

if ! command_exists node; then
    MISSING_DEPS+=("node")
fi

if ! command_exists kubectl; then
    if [[ " ${AUTO_INSTALLABLE[@]} " =~ " kubectl " ]]; then
        echo -e "${YELLOW}kubectl not found. Installing...${NC}"
        install_dependency "kubectl"
        if command_exists kubectl; then
            print_success "kubectl installed successfully"
        else
            MISSING_DEPS+=("kubectl")
        fi
    else
        MISSING_DEPS+=("kubectl")
    fi
fi

if ! command_exists kind; then
    if [[ " ${AUTO_INSTALLABLE[@]} " =~ " kind " ]]; then
        echo -e "${YELLOW}kind not found. Installing...${NC}"
        install_dependency "kind"
        if command_exists kind; then
            print_success "kind installed successfully"
        else
            MISSING_DEPS+=("kind")
        fi
    else
        MISSING_DEPS+=("kind")
    fi
fi

if ! command_exists helm; then
    if [[ " ${AUTO_INSTALLABLE[@]} " =~ " helm " ]]; then
        echo -e "${YELLOW}helm not found. Installing...${NC}"
        install_dependency "helm"
        if command_exists helm; then
            print_success "helm installed successfully"
        else
            MISSING_DEPS+=("helm")
        fi
    else
        MISSING_DEPS+=("helm")
    fi
fi

if [ ${#MISSING_DEPS[@]} -ne 0 ]; then
    print_error "Missing required dependencies: ${MISSING_DEPS[*]}"
    echo -e "\n${YELLOW}Please install the following manually:${NC}"
    for dep in "${MISSING_DEPS[@]}"; do
        case $dep in
            "docker")
                echo "  • Docker: https://docs.docker.com/get-docker/"
                ;;
            "go")
                echo "  • Go (1.21+): https://golang.org/doc/install"
                ;;
            "node")
                echo "  • Node.js (18+): https://nodejs.org/"
                ;;
        esac
    done
    echo -e "\nAfter installing, run this script again."
    exit 1
fi

print_success "All prerequisites installed"

# Create kind cluster
print_step "Creating Kind cluster..."

if kind get clusters | grep -q "hexabase-dev"; then
    echo "Kind cluster 'hexabase-dev' already exists. Using existing cluster."
else
    cat <<EOF | kind create cluster --name hexabase-dev --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
- role: worker
- role: worker
EOF
    print_success "Kind cluster created"
fi

kubectl config use-context "kind-hexabase-dev"

# Install ingress controller
print_step "Installing NGINX Ingress Controller..."

if kubectl get ns ingress-nginx >/dev/null 2>&1; then
    echo "NGINX Ingress already installed"
else
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
    echo "Waiting for ingress controller to be ready..."
    kubectl wait --namespace ingress-nginx \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/component=controller \
        --timeout=90s
    print_success "NGINX Ingress installed"
fi

# Install cert-manager (for TLS)
print_step "Installing cert-manager..."

if kubectl get ns cert-manager >/dev/null 2>&1; then
    echo "cert-manager already installed"
else
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
    echo "Waiting for cert-manager to be ready..."
    kubectl wait --namespace cert-manager \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/component=webhook \
        --timeout=90s
    print_success "cert-manager installed"
fi

# Start local services with Docker Compose
print_step "Starting infrastructure services..."

cd "$PROJECT_ROOT"

# Create docker-compose.override.yml for development (if not exists)
if [ -f "docker-compose.override.yml" ]; then
    print_success "docker-compose.override.yml already exists, skipping creation"
else
    cat > docker-compose.override.yml <<EOF
version: '3.8'

services:
  postgres:
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: devpassword
      POSTGRES_USER: hexabase
      POSTGRES_DB: hexabase_kaas

  redis:
    ports:
      - "6379:6379"
    command: redis-server --requirepass devpassword

  nats:
    image: nats:2.9-alpine
    ports:
      - "4222:4222"
      - "8222:8222"
    command: "-js -sd /data"
    volumes:
      - nats-data:/data

volumes:
  nats-data:
EOF
    print_success "docker-compose.override.yml created"
fi

DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
$DOCKER_COMPOSE_CMD up -d postgres redis nats
print_success "Infrastructure services started"

# Wait for services to be ready
print_step "Waiting for services to be ready..."
sleep 10

# Run database migrations
print_step "Running database migrations..."

# Create .env file for API
cat > "$PROJECT_ROOT/api/.env" <<EOF
# Database
DATABASE_URL=postgres://hexabase:devpassword@localhost:5432/hexabase_kaas?sslmode=disable

# Redis
REDIS_URL=redis://:devpassword@localhost:6379

# NATS
NATS_URL=nats://localhost:4222

# JWT Keys
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem

# OAuth (development)
GOOGLE_CLIENT_ID=dev-google-client-id
GOOGLE_CLIENT_SECRET=dev-google-client-secret

# Stripe (test keys)
STRIPE_API_KEY=sk_test_dummy_key
STRIPE_WEBHOOK_SECRET=whsec_test_dummy_secret

# Kubernetes
KUBECONFIG=$HOME/.kube/config
KUBE_CONTEXT=$K8S_CONTEXT
EOF

# Generate JWT keys
print_step "Generating JWT keys..."
mkdir -p "$PROJECT_ROOT/api/keys"
if [ ! -f "$PROJECT_ROOT/api/keys/private.pem" ]; then
    openssl genrsa -out "$PROJECT_ROOT/api/keys/private.pem" 2048
    openssl rsa -in "$PROJECT_ROOT/api/keys/private.pem" -pubout -out "$PROJECT_ROOT/api/keys/public.pem"
    print_success "JWT keys generated"
else
    echo "JWT keys already exist"
fi

# Create UI .env file
cat > "$PROJECT_ROOT/ui/.env.local" <<EOF
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
EOF

# Install vCluster
print_step "Installing vCluster..."

if ! command_exists vcluster; then
    curl -L -o vcluster "https://github.com/loft-sh/vcluster/releases/latest/download/vcluster-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')"
    chmod +x vcluster
    sudo mv vcluster /usr/local/bin/
    print_success "vCluster CLI installed"
else
    echo "vCluster CLI already installed"
fi

# Deploy Hexabase KaaS to Kind
print_step "Deploying Hexabase KaaS to Kind cluster..."

# Add Hexabase helm repo (using local chart for now)
if [ -d "$PROJECT_ROOT/deployments/helm/hexabase-kaas" ]; then
    helm upgrade --install hexabase-kaas "$PROJECT_ROOT/deployments/helm/hexabase-kaas" \
        --namespace "$NAMESPACE" \
        --create-namespace \
        --values "$PROJECT_ROOT/deployments/helm/values-local.yaml" \
        --wait
else
    echo "Note: Helm chart not found locally. Skipping Kubernetes deployment."
    echo "You can deploy manually later using:"
    echo "  helm install hexabase-kaas ./deployments/helm/hexabase-kaas -f deployments/helm/values-local.yaml"
fi

# Add entries to /etc/hosts
print_step "Updating /etc/hosts..."

if grep -q "api.localhost" /etc/hosts && grep -q "app.localhost" /etc/hosts; then
    echo "/etc/hosts already configured"
else
    echo "Adding entries to /etc/hosts (requires sudo)..."
    sudo tee -a /etc/hosts > /dev/null <<EOF

# Hexabase KaaS Development
127.0.0.1 api.localhost app.localhost
EOF
    print_success "/etc/hosts updated"
fi

# Print summary
echo -e "\n${GREEN}✨ Development environment setup complete!${NC}"
echo -e "\n${YELLOW}Services running:${NC}"
echo "  • PostgreSQL: localhost:5432 (user: hexabase, pass: devpassword)"
echo "  • Redis: localhost:6379 (pass: devpassword)"
echo "  • NATS: localhost:4222"
echo "  • Kind cluster: hexabase-dev"

echo -e "\n${YELLOW}Next steps:${NC}"
echo "1. Start the API server:"
echo "   cd api && go run cmd/api/main.go"
echo ""
echo "2. Start the UI development server:"
echo "   cd ui && npm install && npm run dev"
echo ""
echo "3. Access the application:"
echo "   • API: http://api.localhost"
echo "   • UI: http://app.localhost"

echo -e "\n${YELLOW}Useful commands:${NC}"
echo "  • View logs: ${DOCKER_COMPOSE_CMD:-docker compose} logs -f"
echo "  • Stop services: ${DOCKER_COMPOSE_CMD:-docker compose} down"
echo "  • Delete Kind cluster: kind delete cluster --name hexabase-dev"
echo "  • Connect to cluster: kubectl config use-context kind-hexabase-dev"

echo -e "\n${GREEN}Happy coding! 🚀${NC}"