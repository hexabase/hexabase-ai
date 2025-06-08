# Hexabase AIOps

Intelligent Operations Platform for Hexabase Kubernetes as a Service.

## Overview

The AIOps service provides AI-powered operations capabilities for Hexabase KaaS, including:

- **Intelligent Monitoring**: AI-driven anomaly detection and alerting
- **Predictive Analytics**: Resource usage forecasting and capacity planning  
- **Automated Remediation**: Smart incident response and auto-healing
- **Security Intelligence**: Threat detection and compliance monitoring
- **Performance Optimization**: AI-based resource optimization recommendations

## Architecture

### Security Sandbox

The AIOps service operates within a JWT-based security sandbox with:

- **Stateless Permission Model**: All operations validated against JWT claims
- **Workspace Isolation**: Strict tenant boundaries enforced at API level
- **Limited Scope Tokens**: Time-bounded access tokens with minimal permissions
- **Audit Trail**: Complete logging of all AIOps actions via ClickHouse

### Components

- **Core Engine**: FastAPI-based service with async operations
- **LLM Integration**: Ollama for private model hosting + OpenAI API fallback
- **Observability**: Langfuse for LLMOps monitoring and tracing
- **Data Pipeline**: Real-time metrics ingestion from Prometheus/ClickHouse
- **Security**: JWT validation, workspace isolation, audit logging

## Development

### Prerequisites

- Python 3.11+
- Poetry for dependency management
- Docker for local development
- Access to Hexabase Control Plane API

### Setup

```bash
# Install dependencies
poetry install

# Set up pre-commit hooks
poetry run pre-commit install

# Start development server
poetry run uvicorn aiops.main:app --reload --host 0.0.0.0 --port 8000
```

### Environment Variables

```bash
# JWT Configuration
JWT_SECRET_KEY=your-secret-key
JWT_ALGORITHM=RS256
JWT_ISSUER=hexabase-control-plane

# Control Plane Integration
HEXABASE_API_URL=http://hexabase-api:8080
HEXABASE_API_TOKEN=your-api-token

# LLM Configuration
OLLAMA_BASE_URL=http://ollama:11434
OPENAI_API_KEY=your-openai-key  # Fallback only

# Observability
LANGFUSE_PUBLIC_KEY=your-langfuse-key
LANGFUSE_SECRET_KEY=your-langfuse-secret
LANGFUSE_HOST=http://langfuse:3000

# Data Sources
CLICKHOUSE_URL=http://clickhouse:8123
CLICKHOUSE_USER=hexabase
CLICKHOUSE_PASSWORD=your-password
PROMETHEUS_URL=http://prometheus-shared:9090

# Redis Cache
REDIS_URL=redis://redis:6379/0
```

### Testing

```bash
# Run tests
poetry run pytest

# Run with coverage
poetry run pytest --cov=aiops --cov-report=html

# Run security checks
poetry run bandit -r src/aiops/
```

## API Endpoints

### Authentication
- `POST /auth/validate` - Validate JWT token and extract permissions

### Monitoring
- `GET /workspaces/{workspace_id}/alerts` - Get active alerts for workspace
- `POST /workspaces/{workspace_id}/analyze` - Trigger AI analysis of workspace metrics
- `GET /workspaces/{workspace_id}/insights` - Get AI-generated insights

### Operations
- `POST /workspaces/{workspace_id}/remediate` - Execute automated remediation
- `GET /workspaces/{workspace_id}/recommendations` - Get optimization recommendations

### System
- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus metrics endpoint

## Security Model

### JWT Claims Required

```json
{
  "sub": "user-id",
  "workspace_id": "workspace-uuid",
  "permissions": ["aiops:read", "aiops:analyze", "aiops:remediate"],
  "plan_type": "shared|dedicated",
  "exp": 1234567890,
  "iss": "hexabase-control-plane"
}
```

### Permission Levels

- **aiops:read**: View alerts, insights, and recommendations
- **aiops:analyze**: Trigger AI analysis and generate insights  
- **aiops:remediate**: Execute automated remediation actions
- **aiops:admin**: Full administrative access (internal operations only)

## Deployment

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hexabase-aiops
  namespace: hexabase-control-plane
spec:
  replicas: 2
  selector:
    matchLabels:
      app: hexabase-aiops
  template:
    metadata:
      labels:
        app: hexabase-aiops
    spec:
      containers:
      - name: aiops
        image: hexabase/aiops:latest
        ports:
        - containerPort: 8000
        env:
        - name: JWT_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: aiops-secrets
              key: jwt-secret
        # ... other env vars
```

### Docker

```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY pyproject.toml poetry.lock ./
RUN pip install poetry && poetry config virtualenvs.create false && poetry install --only=main

COPY src/ ./src/
EXPOSE 8000

CMD ["uvicorn", "aiops.main:app", "--host", "0.0.0.0", "--port", "8000"]
```