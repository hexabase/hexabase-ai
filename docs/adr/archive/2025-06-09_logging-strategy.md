# ClickHouse for Centralized Logging

**Date**: 2024-12-16  
**Status**: Accepted  
**Deciders**: Architecture Team  
**Tags**: logging, observability, infrastructure

## Context

Our multi-tenant Kubernetes platform generates significant log volumes across multiple components:
- Control plane logs
- vCluster logs per tenant
- Application logs
- Audit logs
- Security events

We needed a scalable, cost-effective solution for log aggregation, storage, and analysis that could handle:
- High ingestion rates (10K+ logs/second)
- Long-term retention (90 days hot, 1 year cold)
- Fast queries across billions of log entries
- Multi-tenant isolation
- Compliance requirements

## Decision

We will use ClickHouse as our centralized logging backend with the following architecture:
- ClickHouse cluster for log storage and analytics
- Fluent Bit for log collection from Kubernetes
- Custom schema optimized for our use cases
- Materialized views for common queries
- TTL policies for automatic data lifecycle management

## Consequences

### Positive
- Extremely fast analytical queries on large datasets
- Cost-effective storage with compression ratios of 10:1 or better
- Native support for time-series data
- SQL interface familiar to developers
- Horizontal scalability
- Built-in replication and sharding

### Negative
- Additional operational complexity compared to managed solutions
- Requires ClickHouse expertise for optimization
- Not ideal for full-text search (may need Elasticsearch for specific use cases)
- Schema changes require careful planning

### Risks
- Data loss if replication not properly configured
- Performance degradation if queries not optimized
- Storage costs if retention policies not enforced

## Alternatives Considered

1. **Elasticsearch/OpenSearch**
   - Pros: Excellent full-text search, mature ecosystem
   - Cons: High resource consumption, expensive at scale
   - Rejected due to cost at our projected volumes

2. **Loki (Grafana)**
   - Pros: Kubernetes-native, good Grafana integration
   - Cons: Limited query capabilities, performance issues at scale
   - Rejected due to query limitations

3. **CloudWatch/Stackdriver**
   - Pros: Fully managed, no operations overhead
   - Cons: Vendor lock-in, very expensive at scale
   - Rejected due to cost and multi-cloud requirements

## References

- [ClickHouse Documentation](https://clickhouse.com/docs/)
- [Log Schema Design](../../../deployments/monitoring/clickhouse/schema.sql)
- [Fluent Bit Configuration](../../../deployments/monitoring/fluent-bit/)
- [Internal: Logging Architecture RFC]