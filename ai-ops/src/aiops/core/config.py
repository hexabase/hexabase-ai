"""
Configuration management for AIOps service.
"""

import os
from functools import lru_cache
from typing import List

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings."""
    
    model_config = SettingsConfigDict(
        env_file=".env",
        env_ignore_empty=True,
        extra="ignore"
    )
    
    # Application
    version: str = "0.1.0"
    debug: bool = Field(default=False, description="Enable debug mode")
    allowed_origins: List[str] = Field(
        default=["http://localhost:3000"], 
        description="CORS allowed origins"
    )
    
    # JWT Configuration
    jwt_secret_key: str = Field(..., description="JWT secret key for token validation")
    jwt_algorithm: str = Field(default="RS256", description="JWT algorithm")
    jwt_issuer: str = Field(default="hexabase-control-plane", description="JWT issuer")
    jwt_audience: str = Field(default="hexabase-aiops", description="JWT audience")
    
    # Control Plane Integration
    hexabase_api_url: str = Field(
        default="http://hexabase-api:8080", 
        description="Hexabase API base URL"
    )
    hexabase_api_token: str = Field(..., description="Hexabase API authentication token")
    
    # LLM Configuration
    ollama_base_url: str = Field(
        default="http://ollama:11434", 
        description="Ollama server base URL"
    )
    ollama_model: str = Field(
        default="llama2:7b", 
        description="Default Ollama model"
    )
    openai_api_key: str = Field(
        default="", 
        description="OpenAI API key (fallback only)"
    )
    
    # Observability
    langfuse_public_key: str = Field(default="", description="Langfuse public key")
    langfuse_secret_key: str = Field(default="", description="Langfuse secret key")
    langfuse_host: str = Field(
        default="http://langfuse:3000", 
        description="Langfuse host URL"
    )
    
    # Data Sources
    clickhouse_url: str = Field(
        default="http://clickhouse:8123", 
        description="ClickHouse server URL"
    )
    clickhouse_user: str = Field(default="hexabase", description="ClickHouse username")
    clickhouse_password: str = Field(default="", description="ClickHouse password")
    clickhouse_database: str = Field(
        default="hexabase_logs", 
        description="ClickHouse database name"
    )
    
    prometheus_url: str = Field(
        default="http://prometheus-shared:9090", 
        description="Prometheus server URL"
    )
    
    # Redis Cache
    redis_url: str = Field(default="redis://redis:6379/0", description="Redis connection URL")
    
    # Security
    max_token_age_seconds: int = Field(
        default=3600, 
        description="Maximum JWT token age in seconds"
    )
    rate_limit_requests: int = Field(
        default=100, 
        description="Rate limit requests per minute"
    )
    
    # AI/ML Configuration
    max_context_length: int = Field(
        default=4096, 
        description="Maximum context length for LLM requests"
    )
    ai_analysis_timeout: int = Field(
        default=60, 
        description="AI analysis timeout in seconds"
    )
    remediation_timeout: int = Field(
        default=300, 
        description="Remediation action timeout in seconds"
    )
    
    # Monitoring
    metrics_collection_interval: int = Field(
        default=30, 
        description="Metrics collection interval in seconds"
    )
    alert_check_interval: int = Field(
        default=60, 
        description="Alert checking interval in seconds"
    )


@lru_cache()
def get_settings() -> Settings:
    """Get application settings (cached)."""
    return Settings()