# Monitoring Setup Guide

Comprehensive guide for setting up monitoring and observability for Hexabase AI platform.

## Overview

The Hexabase AI monitoring stack includes:
- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboards
- **Loki**: Log aggregation
- **Alertmanager**: Alert routing and management
- **ClickHouse**: Long-term log storage
- **OpenTelemetry**: Distributed tracing

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Applications  │────▶│   Prometheus    │────▶│     Grafana     │
│  (metrics/logs) │     │  (time-series)  │     │  (visualization)│
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                                                │
         │              ┌─────────────────┐               │
         └─────────────▶│      Loki       │───────────────┘
                        │  (log storage)  │
                        └─────────────────┘
                                 │
                        ┌─────────────────┐
                        │   ClickHouse    │
                        │ (long-term logs)│
                        └─────────────────┘
```

## Prerequisites

- Kubernetes cluster with Hexabase AI installed
- Helm 3.x
- kubectl configured
- Storage class for persistent volumes
- DNS configured for monitoring endpoints

## Installation

### 1. Create Monitoring Namespace

```bash
kubectl create namespace monitoring
```

### 2. Install Prometheus Stack

```bash
# Add Prometheus community Helm repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Create values file for Prometheus
cat > prometheus-values.yaml <<EOF
prometheus:
  prometheusSpec:
    retention: 30d
    retentionSize: 100GB
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: fast-ssd
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 200Gi
    
    # Resource limits
    resources:
      requests:
        cpu: 1000m
        memory: 2Gi
      limits:
        cpu: 2000m
        memory: 4Gi
    
    # Additional scrape configs
    additionalScrapeConfigs:
    - job_name: 'hexabase-api'
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - hexabase-system
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: \$1:\$2
        target_label: __address__

alertmanager:
  alertmanagerSpec:
    storage:
      volumeClaimTemplate:
        spec:
          storageClassName: fast-ssd
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 10Gi
  
  config:
    global:
      resolve_timeout: 5m
      slack_api_url: '$SLACK_WEBHOOK_URL'
    
    route:
      group_by: ['alertname', 'cluster', 'service']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 12h
      receiver: 'default-receiver'
      routes:
      - match:
          severity: critical
        receiver: 'critical-receiver'
        continue: true
      - match:
          severity: warning
        receiver: 'warning-receiver'
    
    receivers:
    - name: 'default-receiver'
      slack_configs:
      - channel: '#alerts'
        title: 'Hexabase Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}\n{{ end }}'
    
    - name: 'critical-receiver'
      slack_configs:
      - channel: '#critical-alerts'
        title: '🚨 CRITICAL: Hexabase Alert'
      pagerduty_configs:
      - service_key: '$PAGERDUTY_SERVICE_KEY'
    
    - name: 'warning-receiver'
      slack_configs:
      - channel: '#alerts'
        title: '⚠️ Warning: Hexabase Alert'

grafana:
  enabled: true
  adminPassword: '$GRAFANA_ADMIN_PASSWORD'
  
  persistence:
    enabled: true
    storageClassName: fast-ssd
    size: 50Gi
  
  ingress:
    enabled: true
    ingressClassName: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
    hosts:
    - monitoring.hexabase.ai
    tls:
    - secretName: grafana-tls
      hosts:
      - monitoring.hexabase.ai
  
  datasources:
    datasources.yaml:
      apiVersion: 1
      datasources:
      - name: Prometheus
        type: prometheus
        url: http://prometheus-kube-prometheus-prometheus:9090
        access: proxy
        isDefault: true
      - name: Loki
        type: loki
        url: http://loki:3100
        access: proxy
EOF

# Install Prometheus stack
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --values prometheus-values.yaml
```

### 3. Install Loki for Log Aggregation

```bash
# Add Grafana Helm repository
helm repo add grafana https://grafana.github.io/helm-charts

# Create Loki values
cat > loki-values.yaml <<EOF
loki:
  auth_enabled: false
  
  storage:
    type: filesystem
    filesystem:
      chunks_directory: /data/loki/chunks
      rules_directory: /data/loki/rules
  
  persistence:
    enabled: true
    storageClassName: fast-ssd
    size: 100Gi
  
  config:
    table_manager:
      retention_deletes_enabled: true
      retention_period: 168h  # 7 days
    
    limits_config:
      enforce_metric_name: false
      reject_old_samples: true
      reject_old_samples_max_age: 168h
      max_query_length: 0h
      max_streams_per_user: 10000
    
    ingester:
      chunk_idle_period: 30m
      max_chunk_age: 1h
      chunk_target_size: 1572864
      chunk_retain_period: 30s
      max_transfer_retries: 0

promtail:
  enabled: true
  
  config:
    clients:
    - url: http://loki:3100/loki/api/v1/push
    
    positions:
      filename: /tmp/positions.yaml
    
    target_config:
      sync_period: 10s

    pipeline_stages:
    - regex:
        expression: '^(?P<namespace>\S+)\s+(?P<pod>\S+)\s+(?P<container>\S+)\s+(?P<level>\S+)\s+(?P<message>.*)$'
    - labels:
        namespace:
        pod:
        container:
        level:
    - timestamp:
        source: time
        format: RFC3339
    - output:
        source: message
EOF

# Install Loki
helm install loki grafana/loki-stack \
  --namespace monitoring \
  --values loki-values.yaml
```

### 4. Install ClickHouse for Long-term Storage

```bash
# Create ClickHouse configuration
cat > clickhouse-values.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: clickhouse-config
  namespace: monitoring
data:
  config.xml: |
    <clickhouse>
      <logger>
        <level>information</level>
        <log>/var/log/clickhouse-server/clickhouse-server.log</log>
        <errorlog>/var/log/clickhouse-server/clickhouse-server.err.log</errorlog>
        <size>1000M</size>
        <count>10</count>
      </logger>
      
      <max_connections>4096</max_connections>
      <keep_alive_timeout>3</keep_alive_timeout>
      <max_concurrent_queries>100</max_concurrent_queries>
      
      <profiles>
        <default>
          <max_memory_usage>10000000000</max_memory_usage>
          <load_balancing>random</load_balancing>
        </default>
      </profiles>
      
      <users>
        <default>
          <password_sha256_hex>$CLICKHOUSE_PASSWORD_SHA256</password_sha256_hex>
          <networks>
            <ip>::/0</ip>
          </networks>
          <profile>default</profile>
          <quota>default</quota>
        </default>
      </users>
    </clickhouse>
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: clickhouse
  namespace: monitoring
spec:
  serviceName: clickhouse
  replicas: 3
  selector:
    matchLabels:
      app: clickhouse
  template:
    metadata:
      labels:
        app: clickhouse
    spec:
      containers:
      - name: clickhouse
        image: clickhouse/clickhouse-server:23.8
        ports:
        - containerPort: 8123
          name: http
        - containerPort: 9000
          name: native
        volumeMounts:
        - name: data
          mountPath: /var/lib/clickhouse
        - name: config
          mountPath: /etc/clickhouse-server/config.d
        resources:
          requests:
            cpu: 2000m
            memory: 8Gi
          limits:
            cpu: 4000m
            memory: 16Gi
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 500Gi
EOF

kubectl apply -f clickhouse-values.yaml

# Create ClickHouse schema for logs
kubectl exec -n monitoring clickhouse-0 -- clickhouse-client --query "
CREATE DATABASE IF NOT EXISTS logs;

CREATE TABLE IF NOT EXISTS logs.hexabase (
    timestamp DateTime64(3),
    level String,
    namespace String,
    pod String,
    container String,
    message String,
    trace_id String,
    span_id String,
    user_id String,
    workspace_id String,
    project_id String,
    method String,
    path String,
    status_code UInt16,
    duration_ms UInt32,
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1,
    INDEX idx_trace_id trace_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_workspace workspace_id TYPE bloom_filter GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (namespace, pod, timestamp)
TTL timestamp + INTERVAL 90 DAY;
"
```

### 5. Configure Log Forwarding to ClickHouse

```yaml
# fluent-bit-clickhouse.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: monitoring
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush         5
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
    
    [INPUT]
        Name              tail
        Path              /var/log/containers/*hexabase*.log
        Parser            docker
        Tag               kube.*
        Refresh_Interval  5
        Mem_Buf_Limit     50MB
        Skip_Long_Lines   On
    
    [FILTER]
        Name                kubernetes
        Match               kube.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Merge_Log           On
        K8S-Logging.Parser  On
        K8S-Logging.Exclude On
    
    [OUTPUT]
        Name          http
        Match         *
        Host          clickhouse.monitoring.svc.cluster.local
        Port          8123
        URI           /
        Format        json_lines
        Header        X-ClickHouse-Database logs
        Header        X-ClickHouse-Table hexabase
        Header        X-ClickHouse-Format JSONEachRow
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluent-bit
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: fluent-bit
  template:
    metadata:
      labels:
        app: fluent-bit
    spec:
      serviceAccountName: fluent-bit
      containers:
      - name: fluent-bit
        image: fluent/fluent-bit:2.1
        volumeMounts:
        - name: config
          mountPath: /fluent-bit/etc/
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: fluent-bit-config
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
```

### 6. Set Up OpenTelemetry for Tracing

```bash
# Install OpenTelemetry Collector
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts

cat > otel-values.yaml <<EOF
mode: deployment

config:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318
  
  processors:
    batch:
      timeout: 1s
      send_batch_size: 1024
    
    memory_limiter:
      check_interval: 1s
      limit_mib: 2048
      spike_limit_mib: 512
  
  exporters:
    prometheus:
      endpoint: 0.0.0.0:8889
    
    jaeger:
      endpoint: jaeger-collector.monitoring.svc.cluster.local:14250
      tls:
        insecure: true
    
    logging:
      loglevel: info
  
  service:
    pipelines:
      traces:
        receivers: [otlp]
        processors: [memory_limiter, batch]
        exporters: [jaeger, logging]
      
      metrics:
        receivers: [otlp]
        processors: [memory_limiter, batch]
        exporters: [prometheus]

service:
  type: ClusterIP
  ports:
    otlp-grpc:
      port: 4317
    otlp-http:
      port: 4318
    metrics:
      port: 8889

resources:
  limits:
    cpu: 1000m
    memory: 2Gi
  requests:
    cpu: 200m
    memory: 400Mi
EOF

helm install opentelemetry-collector open-telemetry/opentelemetry-collector \
  --namespace monitoring \
  --values otel-values.yaml
```

## Grafana Dashboards

### 1. Import Hexabase Dashboards

```bash
# Download Hexabase dashboards
curl -O https://raw.githubusercontent.com/hexabase/monitoring/main/dashboards/api-performance.json
curl -O https://raw.githubusercontent.com/hexabase/monitoring/main/dashboards/workspace-usage.json
curl -O https://raw.githubusercontent.com/hexabase/monitoring/main/dashboards/resource-utilization.json

# Import via Grafana API
GRAFANA_URL="https://monitoring.hexabase.ai"
GRAFANA_API_KEY="your-api-key"

for dashboard in *.json; do
  curl -X POST \
    -H "Authorization: Bearer $GRAFANA_API_KEY" \
    -H "Content-Type: application/json" \
    -d @$dashboard \
    "$GRAFANA_URL/api/dashboards/db"
done
```

### 2. Key Dashboards to Create

#### API Performance Dashboard
- Request rate by endpoint
- Response time percentiles (p50, p95, p99)
- Error rate by status code
- Active connections
- Request size distribution

#### Workspace Usage Dashboard
- Active workspaces
- Resource usage per workspace
- vCluster provisioning times
- Workspace creation/deletion trends
- Cost allocation by workspace

#### Infrastructure Dashboard
- Node CPU/Memory usage
- Pod distribution across nodes
- Storage utilization
- Network traffic
- Certificate expiration

## Alert Rules

### 1. Critical Alerts

```yaml
# critical-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: hexabase-critical
  namespace: monitoring
spec:
  groups:
  - name: critical
    interval: 30s
    rules:
    - alert: APIDown
      expr: up{job="hexabase-api"} == 0
      for: 1m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Hexabase API is down"
        description: "{{ $labels.instance }} API endpoint has been down for more than 1 minute."
    
    - alert: DatabaseDown
      expr: pg_up == 0
      for: 1m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "PostgreSQL database is down"
        description: "PostgreSQL instance {{ $labels.instance }} is down."
    
    - alert: HighErrorRate
      expr: |
        sum(rate(http_requests_total{status=~"5.."}[5m])) 
        / 
        sum(rate(http_requests_total[5m])) > 0.05
      for: 5m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "High API error rate"
        description: "Error rate is above 5% for the last 5 minutes."
    
    - alert: CertificateExpiringSoon
      expr: certmanager_certificate_expiration_timestamp_seconds - time() < 7 * 24 * 3600
      for: 1h
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Certificate expiring soon"
        description: "Certificate {{ $labels.name }} in namespace {{ $labels.namespace }} expires in less than 7 days."
```

### 2. Warning Alerts

```yaml
# warning-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: hexabase-warnings
  namespace: monitoring
spec:
  groups:
  - name: warnings
    interval: 1m
    rules:
    - alert: HighMemoryUsage
      expr: |
        (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) 
        / node_memory_MemTotal_bytes > 0.85
      for: 10m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "High memory usage on node"
        description: "Node {{ $labels.instance }} memory usage is above 85%."
    
    - alert: HighDiskUsage
      expr: |
        (node_filesystem_size_bytes - node_filesystem_avail_bytes) 
        / node_filesystem_size_bytes > 0.80
      for: 10m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "High disk usage"
        description: "Disk usage on {{ $labels.instance }} is above 80%."
    
    - alert: SlowAPIResponse
      expr: |
        histogram_quantile(0.95, 
          sum(rate(http_request_duration_seconds_bucket[5m])) 
          by (le, endpoint)
        ) > 1
      for: 10m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "Slow API response times"
        description: "95th percentile response time for {{ $labels.endpoint }} is above 1 second."
    
    - alert: PodCrashLooping
      expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
      for: 5m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "Pod is crash looping"
        description: "Pod {{ $labels.namespace }}/{{ $labels.pod }} is crash looping."
```

## Custom Metrics

### 1. Application Metrics

```go
// internal/observability/metrics.go
package observability

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // API metrics
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "hexabase_api_request_duration_seconds",
            Help: "API request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )
    
    ActiveWorkspaces = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "hexabase_active_workspaces",
            Help: "Number of active workspaces",
        },
    )
    
    WorkspaceResources = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "hexabase_workspace_resources",
            Help: "Resource usage by workspace",
        },
        []string{"workspace_id", "resource_type"},
    )
    
    // Business metrics
    WorkspacesCreated = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hexabase_workspaces_created_total",
            Help: "Total number of workspaces created",
        },
        []string{"plan"},
    )
    
    APICallsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hexabase_api_calls_total",
            Help: "Total number of API calls",
        },
        []string{"workspace_id", "endpoint"},
    )
)
```

### 2. SLI/SLO Monitoring

```yaml
# slo-rules.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: hexabase-slo
  namespace: monitoring
spec:
  groups:
  - name: slo
    interval: 30s
    rules:
    # API Availability SLO: 99.9%
    - record: slo:api_availability:ratio
      expr: |
        sum(rate(http_requests_total{status!~"5.."}[5m]))
        /
        sum(rate(http_requests_total[5m]))
    
    - alert: APIAvailabilitySLO
      expr: slo:api_availability:ratio < 0.999
      for: 5m
      labels:
        severity: critical
        slo: true
      annotations:
        summary: "API availability SLO breach"
        description: "API availability is {{ $value | humanizePercentage }}, below 99.9% SLO"
    
    # Latency SLO: 95% of requests < 500ms
    - record: slo:api_latency:ratio
      expr: |
        histogram_quantile(0.95,
          sum(rate(http_request_duration_seconds_bucket{le="0.5"}[5m]))
          by (le)
        )
    
    - alert: APILatencySLO
      expr: slo:api_latency:ratio < 0.95
      for: 5m
      labels:
        severity: warning
        slo: true
      annotations:
        summary: "API latency SLO breach"
        description: "95th percentile latency SLO breach"
```

## Log Analysis Queries

### ClickHouse Queries

```sql
-- Top errors by workspace
SELECT 
    workspace_id,
    level,
    COUNT(*) as error_count,
    groupArray(message)[1:5] as sample_messages
FROM logs.hexabase
WHERE level = 'ERROR'
    AND timestamp > now() - INTERVAL 1 HOUR
GROUP BY workspace_id, level
ORDER BY error_count DESC
LIMIT 10;

-- API performance by endpoint
SELECT 
    path,
    quantile(0.5)(duration_ms) as p50,
    quantile(0.95)(duration_ms) as p95,
    quantile(0.99)(duration_ms) as p99,
    COUNT(*) as requests
FROM logs.hexabase
WHERE timestamp > now() - INTERVAL 1 HOUR
    AND status_code < 500
GROUP BY path
ORDER BY requests DESC;

-- User activity timeline
SELECT 
    toStartOfMinute(timestamp) as minute,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(*) as total_requests
FROM logs.hexabase
WHERE timestamp > now() - INTERVAL 1 DAY
GROUP BY minute
ORDER BY minute;
```

### Loki LogQL Queries

```logql
# Error logs from API pods
{namespace="hexabase-system", container="api"} |= "ERROR"

# Slow requests (>1s)
{namespace="hexabase-system"} 
  | json 
  | duration_ms > 1000
  | line_format "{{.timestamp}} {{.path}} {{.duration_ms}}ms"

# Authentication failures
{namespace="hexabase-system"} 
  |= "authentication failed"
  | json
  | line_format "{{.timestamp}} user={{.user_email}} ip={{.client_ip}}"

# Workspace provisioning timeline
{namespace="hexabase-system"} 
  |~ "workspace.*provisioning|vcluster.*created"
  | json
  | line_format "{{.timestamp}} {{.workspace_id}} {{.message}}"
```

## Maintenance

### 1. Retention Policies

```bash
# Configure Prometheus retention
kubectl patch prometheus kube-prometheus-stack-prometheus \
  -n monitoring \
  --type merge \
  -p '{"spec":{"retention":"30d","retentionSize":"100GB"}}'

# Configure Loki retention
kubectl patch configmap loki \
  -n monitoring \
  --type merge \
  -p '{"data":{"loki.yaml":"table_manager:\n  retention_period: 168h"}}'
```

### 2. Backup Monitoring Data

```bash
# Backup Prometheus data
kubectl exec -n monitoring prometheus-kube-prometheus-prometheus-0 -- \
  tar czf /tmp/prometheus-backup.tar.gz /prometheus

kubectl cp monitoring/prometheus-kube-prometheus-prometheus-0:/tmp/prometheus-backup.tar.gz \
  ./prometheus-backup-$(date +%Y%m%d).tar.gz

# Backup Grafana dashboards
kubectl exec -n monitoring deployment/kube-prometheus-stack-grafana -- \
  grafana-cli admin export-dashboard --dir=/tmp/dashboards

kubectl cp monitoring/deployment/kube-prometheus-stack-grafana:/tmp/dashboards \
  ./grafana-dashboards-$(date +%Y%m%d)
```

### 3. Performance Tuning

```yaml
# Optimize Prometheus performance
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-optimization
  namespace: monitoring
data:
  prometheus.yaml: |
    global:
      scrape_interval: 30s      # Reduce frequency
      evaluation_interval: 30s
      external_labels:
        cluster: 'production'
        region: 'us-east-1'
    
    # Optimize TSDB
    storage:
      tsdb:
        out_of_order_time_window: 30m
        min_block_duration: 2h
        max_block_duration: 48h
        retention.size: 100GB
```

## Troubleshooting

### Common Issues

**High Prometheus memory usage**
```bash
# Check memory usage
kubectl top pod -n monitoring -l app.kubernetes.io/name=prometheus

# Reduce metric cardinality
kubectl exec -n monitoring prometheus-kube-prometheus-prometheus-0 -- \
  promtool tsdb analyze /prometheus

# Identify high cardinality metrics
curl -s http://localhost:9090/api/v1/label/__name__/values | \
  jq -r '.data[]' | \
  xargs -I {} curl -s "http://localhost:9090/api/v1/query?query=count(count+by(__name__)({__name__=\"{}\"}))" | \
  jq -r '.data.result[0].value[1] // 0' | \
  sort -nr | head -20
```

**Grafana not loading dashboards**
```bash
# Check datasources
kubectl exec -n monitoring deployment/kube-prometheus-stack-grafana -- \
  grafana-cli admin data-sources list

# Restart Grafana
kubectl rollout restart deployment/kube-prometheus-stack-grafana -n monitoring
```

**Missing logs in Loki**
```bash
# Check Promtail status
kubectl logs -n monitoring daemonset/loki-promtail --tail=100

# Verify log parsing
kubectl exec -n monitoring daemonset/loki-promtail -- \
  promtail --dry-run --config.file=/etc/promtail/config.yml
```

## Security Best Practices

1. **Enable authentication** for all monitoring endpoints
2. **Use TLS** for all monitoring traffic
3. **Implement RBAC** for Grafana users
4. **Rotate credentials** regularly
5. **Audit access** to monitoring systems
6. **Encrypt backups** of monitoring data
7. **Restrict network access** to monitoring endpoints

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [ClickHouse Documentation](https://clickhouse.com/docs/)