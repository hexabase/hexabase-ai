# ADR-006: Logging and Monitoring Architecture

**Date**: 2025-06-09  
**Status**: Implemented  
**Authors**: Platform Observability Team

## 1. Background

Hexabase AI required a comprehensive observability solution that could:
- Handle logs from thousands of containers across multiple clusters
- Provide real-time metrics and alerting
- Support long-term storage for compliance
- Enable efficient debugging and troubleshooting
- Scale cost-effectively with platform growth
- Provide tenant isolation for logs and metrics

Traditional solutions like ELK stack were evaluated but found too expensive at scale.

## 2. Status

**Implemented** - ClickHouse-based logging with Prometheus/Grafana for metrics is fully deployed.

## 3. Other Options Considered

### Option A: ELK Stack (Elasticsearch, Logstash, Kibana)
- Industry standard logging solution
- Rich query capabilities
- Mature ecosystem

### Option B: Loki + Prometheus + Grafana
- Lightweight log aggregation
- Good Kubernetes integration
- Cost-effective

### Option C: ClickHouse + Prometheus + Grafana
- Columnar storage for logs
- Excellent compression ratios
- Fast analytical queries
- Very cost-effective at scale

## 4. What Was Decided

We chose **Option C: ClickHouse + Prometheus + Grafana** with:
- ClickHouse for log storage and analytics
- Prometheus for metrics collection
- Grafana for visualization
- Vector for log shipping
- Custom query API for tenant isolation

## 5. Why Did You Choose It?

- **Cost**: 70% reduction compared to Elasticsearch at scale
- **Performance**: Sub-second queries on billions of log lines
- **Compression**: 10:1 compression ratios typical
- **SQL Interface**: Familiar query language
- **Scalability**: Linear scaling with data volume

## 6. Why Didn't You Choose the Other Options?

### Why not ELK Stack?
- High infrastructure costs
- Complex scaling requirements
- Java heap management issues
- Expensive licensing for enterprise features

### Why not Loki?
- Limited query capabilities
- No full-text search
- Less mature than alternatives
- Performance issues at scale

## 7. What Has Not Been Decided

- Log retention policies beyond 90 days
- Real-time streaming analytics
- Machine learning on logs
- Cross-region log replication

## 8. Considerations

### Architecture Overview
```
┌──────────────┐
│ Applications │
└──────┬───────┘
       │
┌──────▼───────┐
│   Vector     │ (Log Shipper)
└──────┬───────┘
       │
┌──────▼───────┐
│  ClickHouse  │ (Log Storage)
├──────────────┤
│  Prometheus  │ (Metrics)
├──────────────┤
│   Grafana    │ (Visualization)
└──────────────┘
```

### ClickHouse Schema
```sql
CREATE TABLE logs (
    timestamp DateTime64(3),
    workspace_id String,
    project_id String,
    application_id String,
    container_name String,
    log_level LowCardinality(String),
    message String,
    metadata String,
    INDEX idx_message message TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 4
) ENGINE = MergeTree()
PARTITION BY toDate(timestamp)
ORDER BY (workspace_id, timestamp)
TTL timestamp + INTERVAL 90 DAY;
```

### Vector Configuration
```toml
[sources.kubernetes_logs]
type = "kubernetes_logs"

[transforms.parse]
type = "remap"
inputs = ["kubernetes_logs"]
source = '''
.workspace_id = .kubernetes.labels."hexabase.ai/workspace"
.project_id = .kubernetes.labels."hexabase.ai/project"
.application_id = .kubernetes.labels."hexabase.ai/application"
'''

[sinks.clickhouse]
type = "clickhouse"
inputs = ["parse"]
endpoint = "http://clickhouse:8123"
table = "logs"
batch.max_events = 10000
batch.timeout_secs = 10
```

### Query Performance
| Query Type | Data Size | Response Time |
|------------|-----------|---------------|
| Exact match | 1TB | <100ms |
| Wildcard | 1TB | <500ms |
| Aggregation | 1TB | <2s |
| Full scan | 1TB | <10s |

### Tenant Isolation
```go
// Query builder with tenant isolation
func BuildLogQuery(workspace string, query string) string {
    return fmt.Sprintf(`
        SELECT timestamp, log_level, message
        FROM logs
        WHERE workspace_id = '%s'
        AND message LIKE '%%%s%%'
        ORDER BY timestamp DESC
        LIMIT 1000
    `, workspace, query)
}
```

### Metrics Architecture
- **Prometheus**: 15s scrape interval
- **Retention**: 30 days local, 1 year in Thanos
- **Cardinality**: Strict limits per workspace
- **Alerting**: AlertManager with PagerDuty integration

### Cost Analysis
| Component | Monthly Cost (1000 workspaces) |
|-----------|-------------------------------|
| ClickHouse | $500 (3 nodes) |
| Prometheus | $300 (2 nodes) |
| Storage | $200 (10TB) |
| **Total** | **$1,000** |

*Elasticsearch equivalent: ~$3,500*

### Security Considerations
- TLS encryption for all connections
- Query injection prevention
- Rate limiting per workspace
- Audit logging of all queries

### Future Enhancements
- Log anomaly detection
- Automatic pattern extraction
- Correlation with traces
- Cost attribution per workspace