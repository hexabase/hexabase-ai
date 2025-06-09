#!/bin/bash

# Knative Installation Script for Hexabase AI
# This script installs Knative on a K3s cluster

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
KNATIVE_VERSION="v1.13.0"
KOURIER_VERSION="v1.13.0"
CERT_MANAGER_VERSION="v1.14.0"

echo -e "${BLUE}==========================================="
echo "Knative Installation for Hexabase AI"
echo "==========================================="
echo -e "Knative Version: ${KNATIVE_VERSION}"
echo -e "Kourier Version: ${KOURIER_VERSION}${NC}"
echo ""

# Function to wait for deployment
wait_for_deployment() {
    local namespace=$1
    local deployment=$2
    local timeout=${3:-300}
    
    echo -n "Waiting for $deployment in $namespace to be ready..."
    kubectl wait --for=condition=available --timeout=${timeout}s deployment/$deployment -n $namespace
    echo -e " ${GREEN}✓${NC}"
}

# Function to wait for pods
wait_for_pods() {
    local namespace=$1
    local timeout=${2:-300}
    
    echo "Waiting for all pods in $namespace to be ready..."
    kubectl wait --for=condition=ready pods --all -n $namespace --timeout=${timeout}s || true
}

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}✗ kubectl not found. Please install kubectl first.${NC}"
    exit 1
fi

if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}✗ Cannot connect to Kubernetes cluster. Please check your kubeconfig.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Prerequisites checked${NC}"
echo ""

# Step 1: Install Knative Serving
echo -e "${BLUE}Step 1: Installing Knative Serving...${NC}"

echo "Installing Knative Serving CRDs..."
kubectl apply -f https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-crds.yaml

echo "Installing Knative Serving core components..."
kubectl apply -f https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-core.yaml

# Wait for Knative Serving to be ready
wait_for_deployment knative-serving controller
wait_for_deployment knative-serving webhook
wait_for_deployment knative-serving autoscaler
wait_for_deployment knative-serving activator

echo -e "${GREEN}✓ Knative Serving installed${NC}"
echo ""

# Step 2: Install Kourier
echo -e "${BLUE}Step 2: Installing Kourier networking layer...${NC}"

kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-${KOURIER_VERSION}/kourier.yaml

# Configure Knative to use Kourier
echo "Configuring Knative to use Kourier..."
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

wait_for_deployment kourier-system 3scale-kourier-control
wait_for_deployment kourier-system 3scale-kourier-gateway

echo -e "${GREEN}✓ Kourier installed and configured${NC}"
echo ""

# Step 3: Configure DNS
echo -e "${BLUE}Step 3: Configuring DNS...${NC}"

echo "Installing Magic DNS (nip.io) for development..."
kubectl apply -f https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-default-domain.yaml

echo -e "${GREEN}✓ DNS configured${NC}"
echo ""

# Step 4: Install Knative Eventing (Optional)
read -p "Do you want to install Knative Eventing? (y/N): " install_eventing
if [[ $install_eventing =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}Step 4: Installing Knative Eventing...${NC}"
    
    echo "Installing Knative Eventing CRDs..."
    kubectl apply -f https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/eventing-crds.yaml
    
    echo "Installing Knative Eventing core..."
    kubectl apply -f https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/eventing-core.yaml
    
    echo "Installing In-Memory Channel..."
    kubectl apply -f https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/in-memory-channel.yaml
    
    echo "Installing MT Channel Broker..."
    kubectl apply -f https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/mt-channel-broker.yaml
    
    wait_for_deployment knative-eventing eventing-controller
    wait_for_deployment knative-eventing eventing-webhook
    wait_for_deployment knative-eventing mt-broker-controller
    wait_for_deployment knative-eventing mt-broker-filter
    wait_for_deployment knative-eventing mt-broker-ingress
    
    echo -e "${GREEN}✓ Knative Eventing installed${NC}"
else
    echo -e "${YELLOW}Skipping Knative Eventing installation${NC}"
fi
echo ""

# Step 5: Install cert-manager for TLS
read -p "Do you want to install cert-manager for TLS support? (y/N): " install_certmanager
if [[ $install_certmanager =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}Step 5: Installing cert-manager...${NC}"
    
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml
    
    # Wait for cert-manager to be ready
    wait_for_deployment cert-manager cert-manager
    wait_for_deployment cert-manager cert-manager-webhook
    wait_for_deployment cert-manager cert-manager-cainjector
    
    echo -e "${GREEN}✓ cert-manager installed${NC}"
else
    echo -e "${YELLOW}Skipping cert-manager installation${NC}"
fi
echo ""

# Step 6: Apply custom configurations
echo -e "${BLUE}Step 6: Applying Hexabase AI custom configurations...${NC}"

if [ -f "knative-config.yaml" ]; then
    kubectl apply -f knative-config.yaml
    echo -e "${GREEN}✓ Custom configuration applied${NC}"
else
    echo -e "${YELLOW}⚠ knative-config.yaml not found, skipping custom configuration${NC}"
fi

if [ -f "knative-autoscaling.yaml" ]; then
    kubectl apply -f knative-autoscaling.yaml
    echo -e "${GREEN}✓ Autoscaling configuration applied${NC}"
else
    echo -e "${YELLOW}⚠ knative-autoscaling.yaml not found, skipping autoscaling configuration${NC}"
fi

echo ""

# Step 7: Verify installation
echo -e "${BLUE}Step 7: Verifying installation...${NC}"

if [ -f "verify-knative.sh" ]; then
    ./verify-knative.sh
else
    echo "Running basic verification..."
    kubectl get pods -n knative-serving
    kubectl get pods -n kourier-system
    
    if [[ $install_eventing =~ ^[Yy]$ ]]; then
        kubectl get pods -n knative-eventing
    fi
fi

echo ""
echo -e "${GREEN}==========================================="
echo "Installation Complete!"
echo "==========================================="
echo ""
echo "Next steps:"
echo "1. Check Kourier LoadBalancer status:"
echo "   kubectl get svc kourier -n kourier-system"
echo ""
echo "2. Deploy a test function:"
echo "   kubectl apply -f test-function.yaml"
echo ""
echo "3. Get the function URL:"
echo "   kubectl get ksvc test-function"
echo ""
echo "4. For production use:"
echo "   - Configure real DNS domain"
echo "   - Enable TLS with cert-manager"
echo "   - Apply resource limits and monitoring"
echo -e "===========================================${NC}"