# Knative Installation Guide for Hexabase AI

This guide provides instructions for installing Knative on the host K3s cluster to support serverless function execution.

## Prerequisites

- K3s cluster running (host cluster)
- kubectl configured to access the cluster
- Helm 3.x installed
- Minimum cluster requirements:
  - 3 nodes with at least 4 CPU cores and 8GB RAM each
  - Storage class available for PVCs

## Architecture Overview

Knative will be installed on the host K3s cluster to provide:
- **Knative Serving**: For running serverless functions
- **Knative Eventing**: For event-driven architectures
- **Kourier**: As the networking layer (lightweight alternative to Istio)

## Installation Steps

### 1. Install Knative Serving

```bash
# Install Knative Serving CRDs
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.13.0/serving-crds.yaml

# Install Knative Serving core components
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.13.0/serving-core.yaml

# Verify installation
kubectl get pods -n knative-serving
```

### 2. Install Kourier Networking Layer

```bash
# Install Kourier
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.13.0/kourier.yaml

# Configure Knative to use Kourier
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

# Verify Kourier is running
kubectl get pods -n kourier-system
```

### 3. Configure DNS

For development/testing, use Magic DNS (nip.io):

```bash
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.13.0/serving-default-domain.yaml
```

For production, configure real DNS as described in the production section below.

### 4. Install Knative Eventing (Optional)

```bash
# Install Knative Eventing CRDs
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.13.0/eventing-crds.yaml

# Install Knative Eventing core
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.13.0/eventing-core.yaml

# Install In-Memory Channel (for development)
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.13.0/in-memory-channel.yaml

# Install MT Channel Broker
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.13.0/mt-channel-broker.yaml

# Verify installation
kubectl get pods -n knative-eventing
```

### 5. Configure Knative for Multi-tenancy

Apply our custom configuration for multi-tenant support:

```bash
kubectl apply -f knative-config.yaml
```

### 6. Install HPA (Horizontal Pod Autoscaler)

```bash
# Install metrics-server if not already installed
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Configure Knative autoscaling
kubectl apply -f knative-autoscaling.yaml
```

## Verification

Run the verification script:

```bash
./verify-knative.sh
```

Or manually verify:

```bash
# Check all Knative components are running
kubectl get pods -n knative-serving
kubectl get pods -n knative-eventing
kubectl get pods -n kourier-system

# Test with a sample function
kubectl apply -f test-function.yaml

# Get the function URL
kubectl get ksvc test-function
```

## Production Considerations

### 1. DNS Configuration

For production, configure a real domain:

```bash
kubectl edit cm config-domain -n knative-serving
```

Add your domain:
```yaml
data:
  your-domain.com: ""
```

### 2. TLS/SSL Configuration

Configure HTTPS with cert-manager:

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Configure Knative for HTTPS
kubectl apply -f knative-tls.yaml
```

### 3. Resource Limits

Configure appropriate resource limits in `knative-config.yaml`:

```yaml
data:
  container-concurrency: "100"
  container-concurrency-target-percentage: "70"
  stable-window: "60s"
  panic-window-percentage: "10.0"
  max-scale-up-rate: "1000.0"
  max-scale-down-rate: "2.0"
```

### 4. Monitoring Integration

Knative metrics are exposed for Prometheus scraping. Configure ServiceMonitors:

```bash
kubectl apply -f knative-monitoring.yaml
```

## Troubleshooting

### Common Issues

1. **Pods stuck in Pending state**
   - Check node resources: `kubectl top nodes`
   - Check PVC bindings: `kubectl get pvc -A`

2. **Services not accessible**
   - Check Kourier LoadBalancer: `kubectl get svc -n kourier-system`
   - Verify DNS configuration: `kubectl get cm config-domain -n knative-serving`

3. **Autoscaling not working**
   - Verify metrics-server: `kubectl get deployment metrics-server -n kube-system`
   - Check HPA status: `kubectl get hpa -A`

### Debug Commands

```bash
# Check Knative controller logs
kubectl logs -n knative-serving deployment/controller

# Check Kourier gateway logs
kubectl logs -n kourier-system deployment/3scale-kourier-gateway

# Describe a Knative service
kubectl describe ksvc <service-name>
```

## Uninstallation

To remove Knative from the cluster:

```bash
# Remove Knative Eventing
kubectl delete -f https://github.com/knative/eventing/releases/download/knative-v1.13.0/eventing-core.yaml
kubectl delete -f https://github.com/knative/eventing/releases/download/knative-v1.13.0/eventing-crds.yaml

# Remove Kourier
kubectl delete -f https://github.com/knative/net-kourier/releases/download/knative-v1.13.0/kourier.yaml

# Remove Knative Serving
kubectl delete -f https://github.com/knative/serving/releases/download/knative-v1.13.0/serving-core.yaml
kubectl delete -f https://github.com/knative/serving/releases/download/knative-v1.13.0/serving-crds.yaml
```

## Next Steps

After installing Knative, proceed with:
1. Deploy the hks-func CLI tool
2. Configure function runtimes
3. Set up the Internal Operations API
4. Test function deployment and execution