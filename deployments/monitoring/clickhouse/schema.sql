-- Central logging schema for Hexabase AI AIOps system
-- This schema follows the AIOps directive for high-speed log analysis

CREATE DATABASE IF NOT EXISTS hexabase_logs;

USE hexabase_logs;

-- Control Plane logs table
CREATE TABLE IF NOT EXISTS control_plane_logs (
    timestamp DateTime64(3) CODEC(Delta, ZSTD),
    trace_id String CODEC(ZSTD),
    level LowCardinality(String) CODEC(ZSTD),
    service LowCardinality(String) CODEC(ZSTD),
    component LowCardinality(String) CODEC(ZSTD),
    user_id String CODEC(ZSTD),
    org_id String CODEC(ZSTD),
    workspace_id String CODEC(ZSTD),
    message String CODEC(ZSTD),
    fields Map(String, String) CODEC(ZSTD),
    error_stack String CODEC(ZSTD),
    duration_ms UInt32 CODEC(ZSTD),
    http_method LowCardinality(String) CODEC(ZSTD),
    http_path String CODEC(ZSTD),
    http_status UInt16 CODEC(ZSTD),
    source_file String CODEC(ZSTD),
    source_line UInt32 CODEC(ZSTD)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, service, level)
TTL timestamp + INTERVAL 90 DAY DELETE
SETTINGS index_granularity = 8192;

-- AIOps system logs table
CREATE TABLE IF NOT EXISTS aiops_logs (
    timestamp DateTime64(3) CODEC(Delta, ZSTD),
    trace_id String CODEC(ZSTD),
    level LowCardinality(String) CODEC(ZSTD),
    agent_type LowCardinality(String) CODEC(ZSTD),
    agent_id String CODEC(ZSTD),
    user_id String CODEC(ZSTD),
    workspace_id String CODEC(ZSTD),
    session_id String CODEC(ZSTD),
    message String CODEC(ZSTD),
    llm_model LowCardinality(String) CODEC(ZSTD),
    prompt_tokens UInt32 CODEC(ZSTD),
    completion_tokens UInt32 CODEC(ZSTD),
    total_tokens UInt32 CODEC(ZSTD),
    latency_ms UInt32 CODEC(ZSTD),
    cost_usd Float32 CODEC(ZSTD),
    operation LowCardinality(String) CODEC(ZSTD),
    status LowCardinality(String) CODEC(ZSTD),
    fields Map(String, String) CODEC(ZSTD)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, agent_type, operation)
TTL timestamp + INTERVAL 30 DAY DELETE
SETTINGS index_granularity = 8192;

-- Pipeline execution logs table (from CI/CD)
CREATE TABLE IF NOT EXISTS pipeline_logs (
    timestamp DateTime64(3) CODEC(Delta, ZSTD),
    trace_id String CODEC(ZSTD),
    workspace_id String CODEC(ZSTD),
    project_id String CODEC(ZSTD),
    pipeline_id String CODEC(ZSTD),
    run_id String CODEC(ZSTD),
    stage String CODEC(ZSTD),
    task String CODEC(ZSTD),
    level LowCardinality(String) CODEC(ZSTD),
    message String CODEC(ZSTD),
    exit_code Nullable(Int32) CODEC(ZSTD),
    duration_ms Nullable(UInt32) CODEC(ZSTD),
    provider LowCardinality(String) CODEC(ZSTD),
    fields Map(String, String) CODEC(ZSTD)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, workspace_id, pipeline_id)
TTL timestamp + INTERVAL 180 DAY DELETE
SETTINGS index_granularity = 8192;

-- Kubernetes cluster events table
CREATE TABLE IF NOT EXISTS k8s_events (
    timestamp DateTime64(3) CODEC(Delta, ZSTD),
    namespace String CODEC(ZSTD),
    kind LowCardinality(String) CODEC(ZSTD),
    name String CODEC(ZSTD),
    reason LowCardinality(String) CODEC(ZSTD),
    message String CODEC(ZSTD),
    event_type LowCardinality(String) CODEC(ZSTD),
    source_component LowCardinality(String) CODEC(ZSTD),
    source_host String CODEC(ZSTD),
    count UInt32 CODEC(ZSTD),
    workspace_id String CODEC(ZSTD),
    labels Map(String, String) CODEC(ZSTD)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, namespace, kind)
TTL timestamp + INTERVAL 30 DAY DELETE
SETTINGS index_granularity = 8192;

-- Security events table
CREATE TABLE IF NOT EXISTS security_events (
    timestamp DateTime64(3) CODEC(Delta, ZSTD),
    event_type LowCardinality(String) CODEC(ZSTD),
    severity LowCardinality(String) CODEC(ZSTD),
    user_id String CODEC(ZSTD),
    org_id String CODEC(ZSTD),
    workspace_id String CODEC(ZSTD),
    source_ip String CODEC(ZSTD),
    user_agent String CODEC(ZSTD),
    resource String CODEC(ZSTD),
    action String CODEC(ZSTD),
    result LowCardinality(String) CODEC(ZSTD),
    message String CODEC(ZSTD),
    metadata Map(String, String) CODEC(ZSTD),
    risk_score Float32 CODEC(ZSTD)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, severity, event_type)
TTL timestamp + INTERVAL 365 DAY DELETE
SETTINGS index_granularity = 8192;

-- Performance metrics table
CREATE TABLE IF NOT EXISTS performance_metrics (
    timestamp DateTime64(3) CODEC(Delta, ZSTD),
    metric_name LowCardinality(String) CODEC(ZSTD),
    workspace_id String CODEC(ZSTD),
    service LowCardinality(String) CODEC(ZSTD),
    instance String CODEC(ZSTD),
    value Float64 CODEC(ZSTD),
    unit LowCardinality(String) CODEC(ZSTD),
    tags Map(String, String) CODEC(ZSTD)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, metric_name, workspace_id)
TTL timestamp + INTERVAL 90 DAY DELETE
SETTINGS index_granularity = 8192;

-- Create materialized views for common queries
CREATE MATERIALIZED VIEW IF NOT EXISTS error_summary
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (toStartOfHour(timestamp), service, level)
AS SELECT
    toStartOfHour(timestamp) as timestamp,
    service,
    level,
    count() as error_count
FROM control_plane_logs
WHERE level IN ('error', 'fatal')
GROUP BY toStartOfHour(timestamp), service, level;

-- Create materialized view for AI operations analysis
CREATE MATERIALIZED VIEW IF NOT EXISTS aiops_performance
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (toStartOfHour(timestamp), llm_model, operation)
AS SELECT
    toStartOfHour(timestamp) as timestamp,
    llm_model,
    operation,
    count() as request_count,
    avg(latency_ms) as avg_latency_ms,
    sum(total_tokens) as total_tokens,
    sum(cost_usd) as total_cost_usd
FROM aiops_logs
WHERE llm_model != ''
GROUP BY toStartOfHour(timestamp), llm_model, operation;

-- Index for fast trace lookups
CREATE INDEX IF NOT EXISTS idx_trace_id ON control_plane_logs (trace_id) TYPE bloom_filter(0.1) GRANULARITY 4096;
CREATE INDEX IF NOT EXISTS idx_aiops_trace_id ON aiops_logs (trace_id) TYPE bloom_filter(0.1) GRANULARITY 4096;

-- Index for user and workspace lookups
CREATE INDEX IF NOT EXISTS idx_workspace_id ON control_plane_logs (workspace_id) TYPE bloom_filter(0.1) GRANULARITY 4096;
CREATE INDEX IF NOT EXISTS idx_user_id ON control_plane_logs (user_id) TYPE bloom_filter(0.1) GRANULARITY 4096;