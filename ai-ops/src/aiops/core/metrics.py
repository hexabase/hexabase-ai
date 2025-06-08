"""
Prometheus metrics for AIOps service.
"""

from prometheus_client import Counter, Histogram, Gauge, start_http_server
from typing import Dict, Any

# Request metrics
REQUEST_COUNT = Counter(
    'aiops_requests_total',
    'Total number of requests',
    ['method', 'endpoint', 'status_code', 'workspace_id']
)

REQUEST_DURATION = Histogram(
    'aiops_request_duration_seconds',
    'Request duration in seconds',
    ['method', 'endpoint', 'workspace_id']
)

# AI Analysis metrics
AI_ANALYSIS_COUNT = Counter(
    'aiops_ai_analysis_total',
    'Total number of AI analysis requests',
    ['workspace_id', 'analysis_type', 'status']
)

AI_ANALYSIS_DURATION = Histogram(
    'aiops_ai_analysis_duration_seconds',
    'AI analysis duration in seconds',
    ['workspace_id', 'analysis_type']
)

AI_TOKENS_USED = Counter(
    'aiops_ai_tokens_total',
    'Total number of AI tokens used',
    ['workspace_id', 'model', 'token_type']
)

# Remediation metrics
REMEDIATION_COUNT = Counter(
    'aiops_remediation_total',
    'Total number of remediation actions',
    ['workspace_id', 'action_type', 'status']
)

REMEDIATION_DURATION = Histogram(
    'aiops_remediation_duration_seconds',
    'Remediation action duration in seconds',
    ['workspace_id', 'action_type']
)

# Data source metrics
CLICKHOUSE_QUERIES = Counter(
    'aiops_clickhouse_queries_total',
    'Total number of ClickHouse queries',
    ['query_type', 'status']
)

CLICKHOUSE_QUERY_DURATION = Histogram(
    'aiops_clickhouse_query_duration_seconds',
    'ClickHouse query duration in seconds',
    ['query_type']
)

PROMETHEUS_QUERIES = Counter(
    'aiops_prometheus_queries_total',
    'Total number of Prometheus queries',
    ['query_type', 'status']
)

# System metrics
ACTIVE_WORKSPACES = Gauge(
    'aiops_active_workspaces',
    'Number of active workspaces being monitored'
)

ACTIVE_ALERTS = Gauge(
    'aiops_active_alerts_total',
    'Number of active alerts',
    ['workspace_id', 'severity']
)

# External service metrics
EXTERNAL_SERVICE_REQUESTS = Counter(
    'aiops_external_service_requests_total',
    'Total requests to external services',
    ['service', 'endpoint', 'status_code']
)

EXTERNAL_SERVICE_DURATION = Histogram(
    'aiops_external_service_duration_seconds',
    'External service request duration',
    ['service', 'endpoint']
)

# JWT and auth metrics
JWT_VALIDATIONS = Counter(
    'aiops_jwt_validations_total',
    'Total number of JWT token validations',
    ['status']
)

PERMISSION_CHECKS = Counter(
    'aiops_permission_checks_total',
    'Total number of permission checks',
    ['permission', 'workspace_id', 'result']
)


def setup_metrics() -> None:
    """Initialize metrics collection."""
    # Metrics are automatically registered when imported
    pass


def record_request(method: str, endpoint: str, status_code: int, workspace_id: str, duration: float) -> None:
    """Record HTTP request metrics."""
    REQUEST_COUNT.labels(
        method=method,
        endpoint=endpoint,
        status_code=str(status_code),
        workspace_id=workspace_id
    ).inc()
    
    REQUEST_DURATION.labels(
        method=method,
        endpoint=endpoint,
        workspace_id=workspace_id
    ).observe(duration)


def record_ai_analysis(workspace_id: str, analysis_type: str, status: str, duration: float) -> None:
    """Record AI analysis metrics."""
    AI_ANALYSIS_COUNT.labels(
        workspace_id=workspace_id,
        analysis_type=analysis_type,
        status=status
    ).inc()
    
    AI_ANALYSIS_DURATION.labels(
        workspace_id=workspace_id,
        analysis_type=analysis_type
    ).observe(duration)


def record_ai_tokens(workspace_id: str, model: str, prompt_tokens: int, completion_tokens: int) -> None:
    """Record AI token usage."""
    AI_TOKENS_USED.labels(
        workspace_id=workspace_id,
        model=model,
        token_type="prompt"
    ).inc(prompt_tokens)
    
    AI_TOKENS_USED.labels(
        workspace_id=workspace_id,
        model=model,
        token_type="completion"
    ).inc(completion_tokens)


def record_remediation(workspace_id: str, action_type: str, status: str, duration: float) -> None:
    """Record remediation action metrics."""
    REMEDIATION_COUNT.labels(
        workspace_id=workspace_id,
        action_type=action_type,
        status=status
    ).inc()
    
    REMEDIATION_DURATION.labels(
        workspace_id=workspace_id,
        action_type=action_type
    ).observe(duration)


def record_clickhouse_query(query_type: str, status: str, duration: float) -> None:
    """Record ClickHouse query metrics."""
    CLICKHOUSE_QUERIES.labels(
        query_type=query_type,
        status=status
    ).inc()
    
    CLICKHOUSE_QUERY_DURATION.labels(
        query_type=query_type
    ).observe(duration)


def record_external_service_request(service: str, endpoint: str, status_code: int, duration: float) -> None:
    """Record external service request metrics."""
    EXTERNAL_SERVICE_REQUESTS.labels(
        service=service,
        endpoint=endpoint,
        status_code=str(status_code)
    ).inc()
    
    EXTERNAL_SERVICE_DURATION.labels(
        service=service,
        endpoint=endpoint
    ).observe(duration)


def update_active_workspaces(count: int) -> None:
    """Update active workspaces gauge."""
    ACTIVE_WORKSPACES.set(count)


def update_active_alerts(workspace_id: str, severity: str, count: int) -> None:
    """Update active alerts gauge."""
    ACTIVE_ALERTS.labels(
        workspace_id=workspace_id,
        severity=severity
    ).set(count)


def record_jwt_validation(status: str) -> None:
    """Record JWT validation metrics."""
    JWT_VALIDATIONS.labels(status=status).inc()


def record_permission_check(permission: str, workspace_id: str, result: str) -> None:
    """Record permission check metrics."""
    PERMISSION_CHECKS.labels(
        permission=permission,
        workspace_id=workspace_id,
        result=result
    ).inc()