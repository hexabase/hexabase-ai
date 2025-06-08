"""
Health check endpoints.
"""

from datetime import datetime
from typing import Dict, Any

import structlog
from fastapi import APIRouter
from pydantic import BaseModel

from aiops.core.config import get_settings

logger = structlog.get_logger(__name__)

router = APIRouter()


class HealthResponse(BaseModel):
    """Health check response."""
    status: str
    timestamp: datetime
    version: str
    services: Dict[str, str]


@router.get("/", response_model=HealthResponse)
async def health_check() -> HealthResponse:
    """Basic health check endpoint."""
    settings = get_settings()
    
    # TODO: Add actual service health checks
    services = {
        "clickhouse": "unknown",
        "prometheus": "unknown", 
        "redis": "unknown",
        "ollama": "unknown",
        "langfuse": "unknown"
    }
    
    return HealthResponse(
        status="healthy",
        timestamp=datetime.utcnow(),
        version=settings.version,
        services=services
    )


@router.get("/ready")
async def readiness_check() -> Dict[str, Any]:
    """Kubernetes readiness probe endpoint."""
    # TODO: Add readiness checks for dependencies
    return {"status": "ready"}


@router.get("/live")
async def liveness_check() -> Dict[str, Any]:
    """Kubernetes liveness probe endpoint."""
    return {"status": "alive"}