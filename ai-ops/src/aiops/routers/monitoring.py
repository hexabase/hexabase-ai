"""
Monitoring and alerting endpoints.
"""

from datetime import datetime
from typing import List, Optional
from uuid import UUID

import structlog
from fastapi import APIRouter, Depends, Request
from pydantic import BaseModel, Field

from aiops.auth.middleware import get_auth_context
from aiops.auth.models import AuthContext, Permission
from aiops.core.exceptions import ValidationError

logger = structlog.get_logger(__name__)

router = APIRouter()


class Alert(BaseModel):
    """Alert model."""
    id: str
    workspace_id: str
    severity: str = Field(..., pattern="^(info|warning|critical)$")
    title: str
    description: str
    created_at: datetime
    resolved_at: Optional[datetime] = None
    tags: List[str] = Field(default_factory=list)


class AlertsResponse(BaseModel):
    """Response for alerts endpoint."""
    alerts: List[Alert]
    total_count: int


class AnalysisRequest(BaseModel):
    """Request for AI analysis."""
    analysis_type: str = Field(..., pattern="^(anomaly|performance|security|capacity)$")
    time_range_minutes: int = Field(default=60, ge=5, le=1440)
    include_recommendations: bool = Field(default=True)


class Insight(BaseModel):
    """AI-generated insight."""
    type: str
    severity: str
    title: str
    description: str
    confidence: float = Field(..., ge=0.0, le=1.0)
    recommendations: List[str] = Field(default_factory=list)
    metrics: dict = Field(default_factory=dict)


class AnalysisResponse(BaseModel):
    """Response for AI analysis."""
    analysis_id: str
    workspace_id: str
    analysis_type: str
    status: str
    insights: List[Insight]
    created_at: datetime
    completed_at: Optional[datetime] = None


class InsightsResponse(BaseModel):
    """Response for insights endpoint."""
    insights: List[Insight]
    last_updated: datetime


@router.get("/{workspace_id}/alerts", response_model=AlertsResponse)
async def get_alerts(
    workspace_id: UUID,
    request: Request,
    severity: Optional[str] = None,
    limit: int = Field(default=100, ge=1, le=1000),
    offset: int = Field(default=0, ge=0)
) -> AlertsResponse:
    """
    Get active alerts for a workspace.
    
    Requires: aiops:read permission
    """
    auth = get_auth_context(request)
    auth.require_permission(Permission.READ)
    
    # Validate workspace access
    if str(workspace_id) != auth.workspace_id:
        raise ValidationError("Workspace ID mismatch")
    
    logger.info(
        "Fetching alerts",
        workspace_id=str(workspace_id),
        severity=severity,
        limit=limit,
        offset=offset
    )
    
    # TODO: Implement actual alert fetching from ClickHouse/Prometheus
    alerts = []
    
    return AlertsResponse(
        alerts=alerts,
        total_count=len(alerts)
    )


@router.post("/{workspace_id}/analyze", response_model=AnalysisResponse)
async def trigger_analysis(
    workspace_id: UUID,
    analysis_request: AnalysisRequest,
    request: Request
) -> AnalysisResponse:
    """
    Trigger AI analysis of workspace metrics.
    
    Requires: aiops:analyze permission
    """
    auth = get_auth_context(request)
    auth.require_permission(Permission.ANALYZE)
    
    # Validate workspace access
    if str(workspace_id) != auth.workspace_id:
        raise ValidationError("Workspace ID mismatch")
    
    logger.info(
        "Triggering AI analysis",
        workspace_id=str(workspace_id),
        analysis_type=analysis_request.analysis_type,
        time_range_minutes=analysis_request.time_range_minutes
    )
    
    # TODO: Implement actual AI analysis
    analysis_id = "analysis-001"
    
    return AnalysisResponse(
        analysis_id=analysis_id,
        workspace_id=str(workspace_id),
        analysis_type=analysis_request.analysis_type,
        status="pending",
        insights=[],
        created_at=datetime.utcnow()
    )


@router.get("/{workspace_id}/insights", response_model=InsightsResponse)
async def get_insights(
    workspace_id: UUID,
    request: Request,
    insight_type: Optional[str] = None,
    limit: int = Field(default=50, ge=1, le=200)
) -> InsightsResponse:
    """
    Get AI-generated insights for a workspace.
    
    Requires: aiops:read permission
    """
    auth = get_auth_context(request)
    auth.require_permission(Permission.READ)
    
    # Validate workspace access
    if str(workspace_id) != auth.workspace_id:
        raise ValidationError("Workspace ID mismatch")
    
    logger.info(
        "Fetching insights",
        workspace_id=str(workspace_id),
        insight_type=insight_type,
        limit=limit
    )
    
    # TODO: Implement actual insights fetching
    insights = []
    
    return InsightsResponse(
        insights=insights,
        last_updated=datetime.utcnow()
    )