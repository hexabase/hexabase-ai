# Kubernetes Deployment Guide

This guide provides detailed instructions for deploying Hexabase KaaS on Kubernetes or K3s clusters.

## Prerequisites

### Cluster Requirements

- Kubernetes v1.24+ or K3s v1.24+
- RBAC enabled
- Storage class for persistent volumes
- Ingress controller (nginx/traefik)
- cert-manager for TLS certificates

### Required Tools

```bash
# Helm 3 (required for both deployment methods)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# vcluster CLI
curl -L -o vcluster "https://github.com/loft-sh/vcluster/releases/latest/download/vcluster-linux-amd64"
chmod +x vcluster
sudo mv vcluster /usr/local/bin/
```

## Deployment Method 1: Helm Chart (Recommended)

The easiest and recommended way to deploy Hexabase KaaS is using our official Helm chart.

### 1. Add Hexabase Helm Repository

```bash
helm repo add hexabase https://charts.hexabase.ai
helm repo update
```

### 2. Configure Your Deployment

We provide pre-configured values files for different environments in the `deployments/helm/` directory:

- **Production**: [`values-production.yaml`](../../deployments/helm/values-production.yaml) - Self-hosted infrastructure with HA
- **Development**: [`values-local.yaml`](../../deployments/helm/values-local.yaml) - Local development setup

For production deployments with self-hosted infrastructure, use the production values file which includes:
- External PostgreSQL, Redis, and NATS configuration
- High availability settings
- Security policies
- Monitoring integration
- Backup configuration

You can copy and customize these files for your specific needs:

```bash
# Copy the production values file
cp deployments/helm/values-production.yaml my-values.yaml

# Edit with your specific configuration
vim my-values.yaml
```

Key configuration sections to update:
- `global.domain` - Your domain name
- Database connection details
- OAuth provider credentials
- Storage class names
- Monitoring endpoints

### 3. Install Hexabase KaaS

```bash
# Create namespace
kubectl create namespace hexabase-system

# Install with Helm
helm install hexabase-kaas hexabase/hexabase-kaas \
  --namespace hexabase-system \
  --values values.yaml \
  --wait
```

### 4. Verify Installation

```bash
# Check pod status
kubectl get pods -n hexabase-system

# Check ingress
kubectl get ingress -n hexabase-system

# Check services
kubectl get svc -n hexabase-system

# View logs
kubectl logs -n hexabase-system -l app.kubernetes.io/name=hexabase-api
```

### 5. Access the Application

Once the ingress is configured and DNS is set up:
- UI: https://app.hexabase.ai
- API: https://api.hexabase.ai/health

### Helm Chart Configuration Options

The provided values files in `deployments/helm/` contain comprehensive configuration examples:

- **Self-hosted infrastructure**: See [`values-production.yaml`](../../deployments/helm/values-production.yaml)
- **Cloud-managed services**: Modify the external database sections for RDS, Cloud SQL, ElastiCache, etc.
- **High Availability**: Production values include pod anti-affinity, autoscaling, and PDB configurations
- **Security**: Production values include security contexts, network policies, and RBAC settings

For detailed configuration options, run:
```bash
helm show values hexabase/hexabase-kaas > all-values.yaml
```

### Upgrading with Helm

```bash
# Update repo
helm repo update hexabase

# Check for updates
helm list -n hexabase-system

# Upgrade to new version
helm upgrade hexabase-kaas hexabase/hexabase-kaas \
  --namespace hexabase-system \
  --values values.yaml \
  --wait

# Rollback if needed
helm rollback hexabase-kaas 1 -n hexabase-system
```

### Uninstalling

```bash
helm uninstall hexabase-kaas -n hexabase-system
kubectl delete namespace hexabase-system
```

---

## Deployment Method 2: Manual Deployment (Alternative)

### 1. Create Namespace and Secrets

```bash
# Create namespace
kubectl create namespace hexabase-system

# Create database secret
kubectl create secret generic hexabase-db \
  --namespace hexabase-system \
  --from-literal=username=hexabase \
  --from-literal=password='<secure-password>' \
  --from-literal=database=hexabase_kaas

# Create Redis secret
kubectl create secret generic hexabase-redis \
  --namespace hexabase-system \
  --from-literal=password='<redis-password>'

# Create JWT keys
openssl genrsa -out jwt-private.pem 2048
openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem

kubectl create secret generic hexabase-jwt \
  --namespace hexabase-system \
  --from-file=private.pem=jwt-private.pem \
  --from-file=public.pem=jwt-public.pem

# Create OAuth secrets
kubectl create secret generic hexabase-oauth \
  --namespace hexabase-system \
  --from-literal=google-client-id='<client-id>' \
  --from-literal=google-client-secret='<client-secret>' \
  --from-literal=github-client-id='<client-id>' \
  --from-literal=github-client-secret='<client-secret>'
```

### 2. Install vCluster Operator

```bash
# Add vcluster helm repo
helm repo add loft-sh https://charts.loft.sh
helm repo update

# Install vcluster operator
helm upgrade --install vcluster-operator loft-sh/vcluster-k8s \
  --namespace vcluster-system \
  --create-namespace \
  --set operator.enabled=true
```

### 3. Deploy PostgreSQL (or use external)

For production, use managed PostgreSQL (RDS, Cloud SQL). For testing:

```yaml
# postgres.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: hexabase-system
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: hexabase-system
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14
        env:
        - name: POSTGRES_DB
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: database
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: hexabase-system
spec:
  ports:
  - port: 5432
  selector:
    app: postgres
```

### 4. Deploy Redis

```yaml
# redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: hexabase-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command:
        - redis-server
        - --requirepass
        - $(REDIS_PASSWORD)
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-redis
              key: password
        ports:
        - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: hexabase-system
spec:
  ports:
  - port: 6379
  selector:
    app: redis
```

### 5. Deploy NATS (Message Queue)

```yaml
# nats.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nats
  namespace: hexabase-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nats
  template:
    metadata:
      labels:
        app: nats
    spec:
      containers:
      - name: nats
        image: nats:2.9-alpine
        command:
        - nats-server
        - --js
        - --sd
        - /data
        ports:
        - containerPort: 4222
        - containerPort: 8222
        volumeMounts:
        - name: nats-storage
          mountPath: /data
      volumes:
      - name: nats-storage
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: nats
  namespace: hexabase-system
spec:
  ports:
  - name: client
    port: 4222
  - name: monitoring
    port: 8222
  selector:
    app: nats
```

### 6. Deploy Hexabase KaaS API

```yaml
# hexabase-api.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hexabase-config
  namespace: hexabase-system
data:
  config.yaml: |
    server:
      port: 8080
      mode: production
    
    database:
      host: postgres
      port: 5432
      sslmode: require
    
    redis:
      host: redis
      port: 6379
    
    nats:
      url: nats://nats:4222
    
    auth:
      jwt:
        issuer: https://api.hexabase.ai
        audience: hexabase-kaas
      oauth:
        redirect_base_url: https://app.hexabase.ai
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hexabase-api
  namespace: hexabase-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hexabase-api
  template:
    metadata:
      labels:
        app: hexabase-api
    spec:
      serviceAccountName: hexabase-api
      containers:
      - name: api
        image: hexabase/hexabase-kaas-api:latest
        env:
        - name: CONFIG_PATH
          value: /config/config.yaml
        - name: DB_USERNAME
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: password
        - name: DB_NAME
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: database
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-redis
              key: password
        - name: GOOGLE_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: hexabase-oauth
              key: google-client-id
        - name: GOOGLE_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: hexabase-oauth
              key: google-client-secret
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /config
        - name: jwt-keys
          mountPath: /keys
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: hexabase-config
      - name: jwt-keys
        secret:
          secretName: hexabase-jwt
---
apiVersion: v1
kind: Service
metadata:
  name: hexabase-api
  namespace: hexabase-system
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: hexabase-api
```

### 7. Deploy Frontend

```yaml
# hexabase-ui.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hexabase-ui
  namespace: hexabase-system
spec:
  replicas: 2
  selector:
    matchLabels:
      app: hexabase-ui
  template:
    metadata:
      labels:
        app: hexabase-ui
    spec:
      containers:
      - name: ui
        image: hexabase/hexabase-kaas-ui:latest
        env:
        - name: NEXT_PUBLIC_API_URL
          value: https://api.hexabase.ai
        - name: NEXT_PUBLIC_WS_URL
          value: wss://api.hexabase.ai
        ports:
        - containerPort: 3000
---
apiVersion: v1
kind: Service
metadata:
  name: hexabase-ui
  namespace: hexabase-system
spec:
  ports:
  - port: 80
    targetPort: 3000
  selector:
    app: hexabase-ui
```

### 8. Configure Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hexabase-api
  namespace: hexabase-system
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.hexabase.ai
    secretName: hexabase-api-tls
  rules:
  - host: api.hexabase.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hexabase-api
            port:
              number: 80
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hexabase-ui
  namespace: hexabase-system
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - app.hexabase.ai
    secretName: hexabase-ui-tls
  rules:
  - host: app.hexabase.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hexabase-ui
            port:
              number: 80
```

### 9. Create RBAC Resources

```yaml
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: hexabase-api
  namespace: hexabase-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: hexabase-api
rules:
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list", "create", "delete"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["get", "list"]
- apiGroups: ["vcluster.loft.sh"]
  resources: ["vclusters"]
  verbs: ["get", "list", "create", "update", "delete"]
- apiGroups: [""]
  resources: ["secrets", "configmaps"]
  verbs: ["get", "list", "create", "update", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: hexabase-api
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: hexabase-api
subjects:
- kind: ServiceAccount
  name: hexabase-api
  namespace: hexabase-system
```

### 10. Apply All Resources

```bash
# Apply all configurations
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f nats.yaml
kubectl apply -f rbac.yaml
kubectl apply -f hexabase-api.yaml
kubectl apply -f hexabase-ui.yaml
kubectl apply -f ingress.yaml

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=hexabase-api -n hexabase-system --timeout=300s
kubectl wait --for=condition=ready pod -l app=hexabase-ui -n hexabase-system --timeout=300s
```

### 11. Run Database Migrations

```bash
# Get API pod name
API_POD=$(kubectl get pod -n hexabase-system -l app=hexabase-api -o jsonpath='{.items[0].metadata.name}')

# Run migrations
kubectl exec -n hexabase-system $API_POD -- hexabase-migrate up
```

## Post-Deployment Steps

### 1. Verify Installation

```bash
# Check pod status
kubectl get pods -n hexabase-system

# Check logs
kubectl logs -n hexabase-system -l app=hexabase-api

# Test API health
curl https://api.hexabase.ai/health
```

### 2. Configure DNS

Point your domains to the ingress controller's external IP:
```bash
kubectl get ingress -n hexabase-system
```

### 3. Initial Admin Setup

Access the UI at `https://app.hexabase.ai` and complete initial setup.

## Helm Chart Deployment (Alternative)

For easier deployment, use the Hexabase KaaS Helm chart:

```bash
# Add Hexabase helm repo
helm repo add hexabase https://charts.hexabase.ai
helm repo update

# Install with custom values
helm install hexabase-kaas hexabase/hexabase-kaas \
  --namespace hexabase-system \
  --create-namespace \
  --values values.yaml
```

For example configurations, see the values files in [`deployments/helm/`](../../deployments/helm/).

## Deployment Method Comparison

| Aspect | Helm Chart | Manual Deployment |
|--------|------------|-------------------|
| **Ease of Use** | ⭐⭐⭐⭐⭐ Simple one-command deployment | ⭐⭐ Complex, multiple steps |
| **Maintainability** | ⭐⭐⭐⭐⭐ Easy upgrades and rollbacks | ⭐⭐ Manual tracking required |
| **Configuration** | ⭐⭐⭐⭐⭐ Centralized values.yaml | ⭐⭐ Scattered across files |
| **Customization** | ⭐⭐⭐⭐ Extensive options | ⭐⭐⭐⭐⭐ Full control |
| **Production Ready** | ⭐⭐⭐⭐⭐ Best practices built-in | ⭐⭐⭐ Requires expertise |
| **Time to Deploy** | ~5 minutes | ~30-60 minutes |

**Recommendation**: Use the Helm chart for all deployments unless you have specific requirements that necessitate manual deployment.

## Troubleshooting

### Common Helm Issues

**Chart not found:**
```bash
helm search repo hexabase
# If empty, re-add the repo:
helm repo add hexabase https://charts.hexabase.ai
helm repo update
```

**Values validation errors:**
```bash
# Validate your values file
helm lint hexabase/hexabase-kaas -f values.yaml

# See all available options
helm show values hexabase/hexabase-kaas
```

**Installation timeouts:**
```bash
# Increase timeout
helm install hexabase-kaas hexabase/hexabase-kaas \
  --timeout 10m \
  --wait
```

### Getting Help

1. **Check Helm release status:**
   ```bash
   helm status hexabase-kaas -n hexabase-system
   ```

2. **View rendered manifests:**
   ```bash
   helm get manifest hexabase-kaas -n hexabase-system
   ```

3. **Enable debug output:**
   ```bash
   helm install hexabase-kaas hexabase/hexabase-kaas \
     --debug \
     --dry-run
   ```

## Next Steps

- Set up [Monitoring & Observability](./monitoring-setup.md)
- Configure [Backup & Recovery](./backup-recovery.md)
- Review [Production Setup](./production-setup.md) for hardening
- Explore [Helm Chart Advanced Configuration](https://charts.hexabase.ai/docs)