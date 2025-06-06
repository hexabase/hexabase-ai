# Hexabase KaaS Helm Deployment

This directory contains Helm values files for deploying Hexabase KaaS in different environments.

## Files

- `values-production.yaml` - Production-ready configuration with self-hosted infrastructure
- `values-staging.yaml` - Staging environment configuration with reduced scale
- `values-local.yaml` - Simplified configuration for local development

## Production Deployment

The production values file assumes you have the following self-hosted infrastructure:

### Prerequisites

1. **PostgreSQL Cluster** (v14+)
   - High availability setup (primary + replicas)
   - Service: `postgres.database.svc.cluster.local`
   - Secret: `hexabase-postgresql` with keys `username` and `password`

2. **Redis Sentinel Cluster** (v6+)
   - 3 Redis instances with Sentinel for HA
   - Services: `redis-sentinel-{1,2,3}.cache.svc.cluster.local`
   - Secret: `hexabase-redis` with key `password`

3. **NATS Cluster** (v2.9+)
   - 3-node cluster with JetStream enabled
   - Services: `nats-{1,2,3}.messaging.svc.cluster.local`
   - Secret: `hexabase-nats` with keys `username` and `password`

4. **Monitoring Stack**
   - Prometheus: `prometheus.monitoring.svc.cluster.local`
   - Grafana: `grafana.monitoring.svc.cluster.local`
   - AlertManager: `alertmanager.monitoring.svc.cluster.local`

5. **Storage**
   - S3-compatible storage for backups
   - StorageClass: `fast-ssd` for persistent volumes

### Required Secrets

Create these secrets before deployment:

```bash
# PostgreSQL credentials
kubectl create secret generic hexabase-postgresql \
  --namespace hexabase-system \
  --from-literal=username=hexabase \
  --from-literal=password='your-secure-password'

# Redis password
kubectl create secret generic hexabase-redis \
  --namespace hexabase-system \
  --from-literal=password='your-redis-password'

# NATS credentials
kubectl create secret generic hexabase-nats \
  --namespace hexabase-system \
  --from-literal=username=hexabase \
  --from-literal=password='your-nats-password'

# OAuth providers
kubectl create secret generic hexabase-oauth-google \
  --namespace hexabase-system \
  --from-literal=client-id='your-google-client-id' \
  --from-literal=client-secret='your-google-client-secret'

kubectl create secret generic hexabase-oauth-github \
  --namespace hexabase-system \
  --from-literal=client-id='your-github-client-id' \
  --from-literal=client-secret='your-github-client-secret'

# Stripe
kubectl create secret generic hexabase-stripe \
  --namespace hexabase-system \
  --from-literal=api-key='sk_live_...' \
  --from-literal=webhook-secret='whsec_...'

# S3 backup credentials
kubectl create secret generic hexabase-backup-s3 \
  --namespace hexabase-system \
  --from-literal=access-key-id='your-access-key' \
  --from-literal=secret-access-key='your-secret-key'
```

### Deployment Commands

```bash
# Add Helm repository
helm repo add hexabase https://charts.hexabase.ai
helm repo update

# Deploy to production
helm install hexabase-kaas hexabase/hexabase-kaas \
  --namespace hexabase-system \
  --create-namespace \
  --values values-production.yaml \
  --wait

# Verify deployment
kubectl get pods -n hexabase-system
helm status hexabase-kaas -n hexabase-system
```

## Staging Deployment

The staging configuration provides a production-like environment with reduced scale:

### Key Differences from Production:
- Reduced replicas (2 API, 1 UI)
- Single database instance (no HA)
- Basic Redis setup (no Sentinel)
- Lower resource requests/limits
- Less frequent backups (weekly)
- Basic auth on ingress
- Let's Encrypt staging certificates

### Deploy to Staging:
```bash
helm install hexabase-kaas hexabase/hexabase-kaas \
  --namespace hexabase-staging \
  --create-namespace \
  --values values-staging.yaml
```

## Development Deployment

For local development with kind or minikube:

```bash
# Use the automated setup script (recommended)
./scripts/dev-setup.sh

# Or deploy manually with development values
helm install hexabase-kaas hexabase/hexabase-kaas \
  --namespace hexabase-dev \
  --create-namespace \
  --values values-local.yaml
```

## Customization

You can override any values by:

1. Modifying the values files
2. Using `--set` flags
3. Using multiple values files

Example:
```bash
helm install hexabase-kaas hexabase/hexabase-kaas \
  --values values-production.yaml \
  --values custom-overrides.yaml \
  --set api.replicas=5
```

## Upgrading

```bash
# Check for updates
helm repo update
helm search repo hexabase/hexabase-kaas --versions

# Upgrade
helm upgrade hexabase-kaas hexabase/hexabase-kaas \
  --namespace hexabase-system \
  --values values-production.yaml \
  --wait

# Rollback if needed
helm rollback hexabase-kaas -n hexabase-system
```

## Monitoring

After deployment, access:
- Grafana dashboards: Pre-configured dashboards for Hexabase KaaS
- Prometheus metrics: Available at `/metrics` endpoint
- Custom alerts: Configured in AlertManager

## Backup

Backups are automatically configured to run daily at 2 AM and stored in S3-compatible storage.

To manually trigger a backup:
```bash
kubectl create job --from=cronjob/hexabase-backup hexabase-backup-manual -n hexabase-system
```