"""Data models for Hexabase AI SDK."""

from datetime import datetime
from typing import Optional, Dict, Any, List
from pydantic import BaseModel, Field, validator


class FunctionConfig(BaseModel):
    """Configuration for a function deployment."""
    
    name: str = Field(..., description="Function name")
    runtime: str = Field(..., description="Runtime environment (e.g., python3.9)")
    handler: Optional[str] = Field(None, description="Function handler (e.g., main.handler)")
    memory_mb: int = Field(default=128, ge=128, le=3008, description="Memory allocation in MB")
    timeout_seconds: int = Field(default=30, ge=1, le=900, description="Execution timeout in seconds")
    environment: Dict[str, str] = Field(default_factory=dict, description="Environment variables")
    dependencies: List[str] = Field(default_factory=list, description="Package dependencies")
    
    @validator("runtime")
    def validate_runtime(cls, v):
        supported = ["python3.8", "python3.9", "python3.10", "python3.11", "nodejs16", "nodejs18"]
        if v not in supported:
            raise ValueError(f"Unsupported runtime: {v}. Supported: {', '.join(supported)}")
        return v
    
    @validator("handler")
    def validate_handler(cls, v, values):
        if values.get("runtime", "").startswith("python") and not v:
            raise ValueError("Handler is required for Python runtime")
        return v


class CleanupPolicy(BaseModel):
    """Auto-cleanup policy for functions."""
    
    ttl_hours: Optional[int] = Field(None, ge=1, description="Time-to-live in hours")
    max_executions: Optional[int] = Field(None, ge=1, description="Maximum number of executions")
    idle_hours: Optional[int] = Field(None, ge=1, description="Hours of inactivity before cleanup")
    
    @validator("ttl_hours", "max_executions", "idle_hours")
    def at_least_one_policy(cls, v, values):
        # Ensure at least one policy is set
        if not any([v, values.get("ttl_hours"), values.get("max_executions"), values.get("idle_hours")]):
            raise ValueError("At least one cleanup policy must be specified")
        return v


class FunctionDeployment(BaseModel):
    """Response from function deployment."""
    
    function_id: str
    name: str
    version: str
    endpoint: str
    created_at: str
    runtime: Optional[str] = None
    memory_mb: Optional[int] = None
    timeout_seconds: Optional[int] = None


class FunctionExecution(BaseModel):
    """Response from function execution."""
    
    execution_id: str
    status: str  # pending, running, completed, failed, timeout
    result: Optional[Any] = None
    error: Optional[str] = None
    duration_ms: Optional[int] = None
    billed_duration_ms: Optional[int] = None
    started_at: Optional[str] = None
    completed_at: Optional[str] = None


class Function(BaseModel):
    """Function metadata."""
    
    function_id: str
    name: str
    runtime: str
    version: str
    status: str  # active, inactive, deleting
    created_at: str
    updated_at: str
    last_executed_at: Optional[str] = None
    execution_count: int = 0
    endpoint: str
    memory_mb: int
    timeout_seconds: int