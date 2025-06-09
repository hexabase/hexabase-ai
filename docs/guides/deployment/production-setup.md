# Production Setup Guide

This guide covers deploying Hexabase AI platform in a production environment.

## Overview

A production Hexabase AI deployment consists of:
- Host K3s cluster for the control plane
- PostgreSQL database cluster
- Redis cluster for caching
- NATS for messaging
- Object storage (S3-compatible)
- Load balancers and ingress controllers
- Monitoring and logging infrastructure

## Prerequisites

### Infrastructure Requirements

#### Minimum Production Setup (Small)
- **Control Plane**: 3 nodes (4 CPU, 16GB RAM each)
- **Worker Nodes**: 3 nodes (8 CPU, 32GB RAM each)
- **Database**: 3 nodes (4 CPU, 16GB RAM, 500GB SSD each)
- **Storage**: 1TB S3-compatible object storage

#### Recommended Production Setup (Medium)
- **Control Plane**: 3 nodes (8 CPU, 32GB RAM each)
- **Worker Nodes**: 5+ nodes (16 CPU, 64GB RAM each)
- **Database**: 3 nodes (8 CPU, 32GB RAM, 1TB NVMe each)
- **Cache**: 3 Redis nodes (4 CPU, 16GB RAM each)
- **Storage**: 10TB S3-compatible object storage

### Software Requirements
- Ubuntu 22.04 LTS or RHEL 8+
- K3s v1.28+
- PostgreSQL 15+
- Redis 7+
- NATS 2.10+

## Installation Steps

### 1. Prepare Infrastructure

#### Set Up Nodes

```bash
# On all nodes
sudo apt-get update
sudo apt-get install -y curl wget git

# Configure kernel parameters
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF

sudo sysctl --system

# Disable swap
sudo swapoff -a
sudo sed -i '/ swap / s/^/#/' /etc/fstab
```

#### Configure Firewall

```bash
# Control plane nodes
sudo ufw allow 6443/tcp  # Kubernetes API
sudo ufw allow 2379:2380/tcp  # etcd
sudo ufw allow 10250/tcp  # Kubelet API
sudo ufw allow 10251/tcp  # kube-scheduler
sudo ufw allow 10252/tcp  # kube-controller-manager

# Worker nodes
sudo ufw allow 10250/tcp  # Kubelet API
sudo ufw allow 30000:32767/tcp  # NodePort Services

# All nodes
sudo ufw allow 8472/udp  # Flannel VXLAN
sudo ufw allow 51820/udp  # WireGuard
```

### 2. Install K3s Cluster

#### Install First Control Plane Node

```bash
# On first control plane node
curl -sfL https://get.k3s.io | sh -s - server \
  --cluster-init \
  --disable traefik \
  --disable servicelb \
  --write-kubeconfig-mode 644 \
  --node-taint CriticalAddonsOnly=true:NoExecute \
  --etcd-expose-metrics true
  
# Get node token
sudo cat /var/lib/rancher/k3s/server/node-token
```

#### Join Additional Control Plane Nodes

```bash
# On other control plane nodes
curl -sfL https://get.k3s.io | sh -s - server \
  --server https://<first-node-ip>:6443 \
  --token <node-token> \
  --disable traefik \
  --disable servicelb \
  --node-taint CriticalAddonsOnly=true:NoExecute
```

#### Join Worker Nodes

```bash
# On worker nodes
curl -sfL https://get.k3s.io | sh -s - agent \
  --server https://<control-plane-ip>:6443 \
  --token <node-token>
```

### 3. Install Core Components

#### Install Helm

```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

#### Install NGINX Ingress Controller

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer \
  --set controller.metrics.enabled=true \
  --set controller.podAnnotations."prometheus\.io/scrape"=true
```

#### Install Cert-Manager

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update

helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true \
  --set prometheus.enabled=true
```

### 4. Deploy PostgreSQL Cluster

#### Using CloudNativePG

```bash
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.22/releases/cnpg-1.22.0.yaml

# Create PostgreSQL cluster
cat <<EOF | kubectl apply -f -
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: hexabase-db
  namespace: hexabase-system
spec:
  instances: 3
  primaryUpdateStrategy: unsupervised
  
  postgresql:
    parameters:
      max_connections: "200"
      shared_buffers: "4GB"
      effective_cache_size: "12GB"
      maintenance_work_mem: "1GB"
      checkpoint_completion_target: "0.9"
      wal_buffers: "16MB"
      default_statistics_target: "100"
      random_page_cost: "1.1"
      effective_io_concurrency: "200"
      work_mem: "20MB"
      min_wal_size: "1GB"
      max_wal_size: "4GB"
  
  bootstrap:
    initdb:
      database: hexabase
      owner: hexabase
      secret:
        name: hexabase-db-auth
  
  storage:
    size: 1Ti
    storageClass: fast-ssd
  
  monitoring:
    enabled: true
  
  backup:
    retentionPolicy: "30d"
    barmanObjectStore:
      destinationPath: "s3://hexabase-backups/postgres"
      s3Credentials:
        accessKeyId:
          name: s3-credentials
          key: ACCESS_KEY_ID
        secretAccessKey:
          name: s3-credentials
          key: SECRET_ACCESS_KEY
EOF
```

### 5. Deploy Redis Cluster

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami

helm install redis bitnami/redis \
  --namespace hexabase-system \
  --set auth.enabled=true \
  --set auth.existingSecret=redis-auth \
  --set auth.existingSecretPasswordKey=password \
  --set sentinel.enabled=true \
  --set sentinel.masterSet=hexabase \
  --set replica.replicaCount=3 \
  --set master.persistence.size=50Gi \
  --set replica.persistence.size=50Gi \
  --set metrics.enabled=true
```

### 6. Deploy NATS

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/

helm install nats nats/nats \
  --namespace hexabase-system \
  --set nats.jetstream.enabled=true \
  --set nats.jetstream.memStorage.size=2Gi \
  --set nats.jetstream.fileStorage.size=50Gi \
  --set cluster.enabled=true \
  --set cluster.replicas=3 \
  --set natsbox.enabled=true \
  --set metrics.enabled=true
```

### 7. Deploy Hexabase Control Plane

#### Create Namespace and Secrets

```bash
kubectl create namespace hexabase-system

# Create database secret
kubectl create secret generic hexabase-db-auth \
  --namespace hexabase-system \
  --from-literal=username=hexabase \
  --from-literal=password=$(openssl rand -base64 32)

# Create JWT keys
openssl genrsa -out jwt-private.pem 4096
openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem

kubectl create secret generic jwt-keys \
  --namespace hexabase-system \
  --from-file=private.pem=jwt-private.pem \
  --from-file=public.pem=jwt-public.pem

# Create OAuth secrets
kubectl create secret generic oauth-providers \
  --namespace hexabase-system \
  --from-literal=google-client-id=$GOOGLE_CLIENT_ID \
  --from-literal=google-client-secret=$GOOGLE_CLIENT_SECRET \
  --from-literal=github-client-id=$GITHUB_CLIENT_ID \
  --from-literal=github-client-secret=$GITHUB_CLIENT_SECRET
```

#### Deploy Hexabase API

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
      host: hexabase-db-rw
      port: 5432
      name: hexabase
      sslMode: require
      maxOpenConns: 100
      maxIdleConns: 10
      connMaxLifetime: 1h
    
    redis:
      addr: redis-master:6379
      db: 0
      poolSize: 100
    
    nats:
      url: nats://nats:4222
      streamName: hexabase
    
    auth:
      jwtExpiry: 24h
      refreshExpiry: 168h
    
    kubernetes:
      inCluster: true
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
        image: hexabase/api:latest
        ports:
        - containerPort: 8080
        env:
        - name: CONFIG_PATH
          value: /etc/hexabase/config.yaml
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-db-auth
              key: password
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-auth
              key: password
        volumeMounts:
        - name: config
          mountPath: /etc/hexabase
        - name: jwt-keys
          mountPath: /etc/hexabase/keys
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: hexabase-config
      - name: jwt-keys
        secret:
          secretName: jwt-keys
---
apiVersion: v1
kind: Service
metadata:
  name: hexabase-api
  namespace: hexabase-system
spec:
  selector:
    app: hexabase-api
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hexabase-api
  namespace: hexabase-system
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.hexabase.ai
    secretName: api-tls
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
```

Apply the configuration:

```bash
kubectl apply -f hexabase-api.yaml
```

### 8. Configure Monitoring

#### Install Prometheus Stack

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.retention=30d \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=100Gi \
  --set alertmanager.alertmanagerSpec.storage.volumeClaimTemplate.spec.resources.requests.storage=10Gi
```

#### Install Loki Stack

```bash
helm repo add grafana https://grafana.github.io/helm-charts

helm install loki grafana/loki-stack \
  --namespace monitoring \
  --set loki.persistence.enabled=true \
  --set loki.persistence.size=100Gi \
  --set promtail.enabled=true
```

### 9. Configure Backup

#### Set Up Velero

```bash
# Install Velero CLI
wget https://github.com/vmware-tanzu/velero/releases/download/v1.13.0/velero-v1.13.0-linux-amd64.tar.gz
tar -xvf velero-v1.13.0-linux-amd64.tar.gz
sudo mv velero-v1.13.0-linux-amd64/velero /usr/local/bin/

# Install Velero in cluster
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.9.0 \
  --bucket hexabase-velero-backups \
  --secret-file ./credentials-velero \
  --backup-location-config region=us-east-1 \
  --snapshot-location-config region=us-east-1
```

#### Create Backup Schedule

```bash
# Daily backup of all namespaces
velero schedule create daily-backup \
  --schedule="0 2 * * *" \
  --include-namespaces hexabase-system,hexabase-workspaces \
  --ttl 720h
```

### 10. Security Hardening

#### Network Policies

```yaml
# default-deny-all.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: hexabase-system
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
# allow-api-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-ingress
  namespace: hexabase-system
spec:
  podSelector:
    matchLabels:
      app: hexabase-api
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
```

#### Pod Security Standards

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: hexabase-system
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

## Post-Installation

### 1. Verify Installation

```bash
# Check all pods are running
kubectl get pods -n hexabase-system

# Check API health
curl https://api.hexabase.ai/health

# Check database connection
kubectl exec -n hexabase-system hexabase-db-1 -- psql -U hexabase -c "SELECT version();"
```

### 2. Configure DNS

Point your domain to the load balancer:

```bash
# Get load balancer IP
kubectl get svc -n ingress-nginx ingress-nginx-controller

# Configure DNS A records
api.hexabase.ai -> <LB_IP>
app.hexabase.ai -> <LB_IP>
*.workspaces.hexabase.ai -> <LB_IP>
```

### 3. Initial Admin Setup

```bash
# Create admin user
kubectl exec -n hexabase-system deploy/hexabase-api -- \
  hexabase-cli user create \
    --email admin@hexabase.ai \
    --role super_admin
```

### 4. Configure Observability

Access Grafana:
```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
# Default credentials: admin/prom-operator
```

Import Hexabase dashboards:
- API Performance Dashboard
- Workspace Usage Dashboard
- Resource Utilization Dashboard

## Maintenance

### Regular Tasks

1. **Daily**
   - Check backup completion
   - Review error logs
   - Monitor resource usage

2. **Weekly**
   - Review security alerts
   - Check for updates
   - Analyze performance metrics

3. **Monthly**
   - Rotate secrets
   - Update dependencies
   - Capacity planning review

### Upgrade Process

```bash
# 1. Backup current state
velero backup create pre-upgrade-$(date +%Y%m%d)

# 2. Update Helm values
helm upgrade hexabase-api ./charts/hexabase-api \
  --namespace hexabase-system \
  --values production-values.yaml

# 3. Verify upgrade
kubectl rollout status deployment/hexabase-api -n hexabase-system
```

## Troubleshooting

### Common Issues

**API pods not starting**
```bash
kubectl describe pod -n hexabase-system <pod-name>
kubectl logs -n hexabase-system <pod-name> --previous
```

**Database connection issues**
```bash
# Check database cluster status
kubectl get cluster -n hexabase-system
kubectl describe cluster hexabase-db -n hexabase-system
```

**High memory usage**
```bash
# Check resource usage
kubectl top nodes
kubectl top pods -n hexabase-system

# Adjust resource limits if needed
```

## Security Checklist

- [ ] All secrets stored in Kubernetes secrets
- [ ] Network policies configured
- [ ] Pod security standards enforced
- [ ] Regular security scanning enabled
- [ ] Audit logging configured
- [ ] Backup encryption enabled
- [ ] TLS certificates valid and auto-renewing
- [ ] RBAC properly configured
- [ ] Resource quotas set
- [ ] Monitoring alerts configured

## Support

For production support:
- Email: support@hexabase.ai
- Enterprise Support Portal: https://support.hexabase.ai
- 24/7 Hotline: +1-xxx-xxx-xxxx (Enterprise only)