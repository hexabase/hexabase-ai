#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "==========================================="
echo "Knative Installation Verification Script"
echo "==========================================="
echo ""

# Function to check if namespace exists
check_namespace() {
    local namespace=$1
    if kubectl get namespace "$namespace" &> /dev/null; then
        echo -e "${GREEN}✓${NC} Namespace $namespace exists"
        return 0
    else
        echo -e "${RED}✗${NC} Namespace $namespace does not exist"
        return 1
    fi
}

# Function to check pod status
check_pods() {
    local namespace=$1
    echo ""
    echo "Checking pods in $namespace..."
    
    local pods=$(kubectl get pods -n "$namespace" --no-headers 2>/dev/null)
    if [ -z "$pods" ]; then
        echo -e "${YELLOW}⚠${NC} No pods found in $namespace"
        return 1
    fi
    
    local all_ready=true
    while IFS= read -r line; do
        local pod_name=$(echo "$line" | awk '{print $1}')
        local ready=$(echo "$line" | awk '{print $2}')
        local status=$(echo "$line" | awk '{print $3}')
        
        if [[ "$status" == "Running" ]] || [[ "$status" == "Completed" ]]; then
            echo -e "${GREEN}✓${NC} Pod $pod_name is $status ($ready)"
        else
            echo -e "${RED}✗${NC} Pod $pod_name is $status ($ready)"
            all_ready=false
        fi
    done <<< "$pods"
    
    if $all_ready; then
        return 0
    else
        return 1
    fi
}

# Function to check CRDs
check_crds() {
    echo ""
    echo "Checking Knative CRDs..."
    
    local crds=(
        "services.serving.knative.dev"
        "configurations.serving.knative.dev"
        "revisions.serving.knative.dev"
        "routes.serving.knative.dev"
        "brokers.eventing.knative.dev"
        "triggers.eventing.knative.dev"
        "eventtypes.eventing.knative.dev"
    )
    
    for crd in "${crds[@]}"; do
        if kubectl get crd "$crd" &> /dev/null; then
            echo -e "${GREEN}✓${NC} CRD $crd is installed"
        else
            echo -e "${YELLOW}⚠${NC} CRD $crd is not installed (might be optional)"
        fi
    done
}

# Function to check services
check_services() {
    local namespace=$1
    echo ""
    echo "Checking services in $namespace..."
    
    local services=$(kubectl get svc -n "$namespace" --no-headers | awk '{print $1}')
    if [ -z "$services" ]; then
        echo -e "${YELLOW}⚠${NC} No services found in $namespace"
        return 1
    fi
    
    echo "$services" | while read -r svc; do
        echo -e "${GREEN}✓${NC} Service $svc is available"
    done
}

# Function to check Kourier LoadBalancer
check_kourier_lb() {
    echo ""
    echo "Checking Kourier LoadBalancer..."
    
    local lb_ip=$(kubectl get svc kourier -n kourier-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
    local lb_hostname=$(kubectl get svc kourier -n kourier-system -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)
    
    if [ -n "$lb_ip" ]; then
        echo -e "${GREEN}✓${NC} Kourier LoadBalancer IP: $lb_ip"
    elif [ -n "$lb_hostname" ]; then
        echo -e "${GREEN}✓${NC} Kourier LoadBalancer Hostname: $lb_hostname"
    else
        echo -e "${YELLOW}⚠${NC} Kourier LoadBalancer IP/Hostname not assigned yet"
        echo "  You may need to configure your cloud provider or use a tunnel service"
    fi
}

# Function to check domain configuration
check_domain() {
    echo ""
    echo "Checking domain configuration..."
    
    local domain=$(kubectl get cm config-domain -n knative-serving -o jsonpath='{.data}' 2>/dev/null)
    if [ -n "$domain" ]; then
        echo -e "${GREEN}✓${NC} Domain configuration found:"
        echo "$domain" | sed 's/^/  /'
    else
        echo -e "${RED}✗${NC} Domain configuration not found"
    fi
}

# Function to test sample deployment
test_deployment() {
    echo ""
    echo "Testing sample deployment..."
    
    # Check if test-function exists
    if kubectl get ksvc test-function &> /dev/null; then
        echo -e "${GREEN}✓${NC} Test function already deployed"
        
        # Get URL
        local url=$(kubectl get ksvc test-function -o jsonpath='{.status.url}' 2>/dev/null)
        if [ -n "$url" ]; then
            echo -e "${GREEN}✓${NC} Test function URL: $url"
            
            # Try to curl the endpoint
            if command -v curl &> /dev/null; then
                echo "  Testing HTTP request..."
                if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "200\|301\|302"; then
                    echo -e "${GREEN}✓${NC} Test function is responding"
                else
                    echo -e "${YELLOW}⚠${NC} Test function not responding (might need DNS setup)"
                fi
            fi
        fi
    else
        echo -e "${YELLOW}⚠${NC} Test function not deployed. Deploy with:"
        echo "  kubectl apply -f test-function.yaml"
    fi
}

# Main verification
echo "1. Checking Namespaces..."
echo "--------------------------"
check_namespace "knative-serving" || exit 1
check_namespace "knative-eventing"
check_namespace "kourier-system"

echo ""
echo "2. Checking Knative Serving Components..."
echo "-----------------------------------------"
check_pods "knative-serving"
check_services "knative-serving"

echo ""
echo "3. Checking Knative Eventing Components..."
echo "------------------------------------------"
if check_namespace "knative-eventing" &> /dev/null; then
    check_pods "knative-eventing"
    check_services "knative-eventing"
else
    echo -e "${YELLOW}⚠${NC} Knative Eventing not installed (optional)"
fi

echo ""
echo "4. Checking Kourier Networking..."
echo "---------------------------------"
check_pods "kourier-system"
check_kourier_lb

echo ""
echo "5. Checking CRDs..."
echo "-------------------"
check_crds

echo ""
echo "6. Checking Configuration..."
echo "----------------------------"
check_domain

echo ""
echo "7. Testing Deployment..."
echo "------------------------"
test_deployment

echo ""
echo "==========================================="
echo "Verification Summary"
echo "==========================================="

# Summary
if kubectl get pods -n knative-serving --no-headers | grep -qv "Running\|Completed"; then
    echo -e "${RED}✗${NC} Some Knative components are not ready"
    echo "  Run 'kubectl get pods -n knative-serving' for details"
else
    echo -e "${GREEN}✓${NC} Knative Serving is operational"
fi

if kubectl get namespace knative-eventing &> /dev/null; then
    if kubectl get pods -n knative-eventing --no-headers | grep -qv "Running\|Completed"; then
        echo -e "${YELLOW}⚠${NC} Some Eventing components are not ready"
    else
        echo -e "${GREEN}✓${NC} Knative Eventing is operational"
    fi
fi

echo ""
echo "Next steps:"
echo "1. If using cloud provider, ensure LoadBalancer is configured"
echo "2. Configure DNS for your domain"
echo "3. Deploy test functions with 'kubectl apply -f test-function.yaml'"
echo "4. Install hks-func CLI tool for easier function management"