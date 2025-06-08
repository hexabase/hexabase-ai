"""
Main FastAPI application for Hexabase AIOps.
"""

import os
from contextlib import asynccontextmanager
from typing import AsyncGenerator

import structlog
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from prometheus_client import make_asgi_app

from aiops.auth.middleware import JWTAuthMiddleware
from aiops.core.config import get_settings
from aiops.core.exceptions import AIOpsException
from aiops.core.logging import setup_logging
from aiops.core.metrics import setup_metrics
from aiops.routers import auth, chat, health, monitoring, operations

logger = structlog.get_logger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Application lifespan context manager."""
    settings = get_settings()
    
    # Setup logging
    setup_logging(settings)
    logger.info("AIOps service starting", version=settings.version)
    
    # Setup metrics
    setup_metrics()
    
    # Initialize external connections
    # TODO: Initialize ClickHouse, Redis, Ollama connections
    
    yield
    
    # Cleanup
    logger.info("AIOps service shutting down")


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    settings = get_settings()
    
    app = FastAPI(
        title="Hexabase AIOps",
        description="Intelligent Operations Platform for Hexabase KaaS",
        version=settings.version,
        docs_url="/docs" if settings.debug else None,
        redoc_url="/redoc" if settings.debug else None,
        lifespan=lifespan
    )
    
    # CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.allowed_origins,
        allow_credentials=True,
        allow_methods=["GET", "POST", "PUT", "DELETE"],
        allow_headers=["*"],
    )
    
    # JWT Authentication middleware (applied to all routes except health/metrics)
    app.add_middleware(JWTAuthMiddleware)
    
    # Exception handler
    @app.exception_handler(AIOpsException)
    async def aiops_exception_handler(request: Request, exc: AIOpsException) -> JSONResponse:
        logger.error("AIOps exception", error=str(exc), error_code=exc.error_code)
        return JSONResponse(
            status_code=exc.status_code,
            content={
                "error": exc.error_code,
                "message": str(exc),
                "details": exc.details
            }
        )
    
    # Include routers
    app.include_router(health.router, prefix="/health", tags=["health"])
    app.include_router(auth.router, prefix="/auth", tags=["auth"])
    app.include_router(chat.router, tags=["chat"])  # Chat router at root level for /v1/chat
    app.include_router(monitoring.router, prefix="/workspaces", tags=["monitoring"])
    app.include_router(operations.router, prefix="/workspaces", tags=["operations"])
    
    # Prometheus metrics endpoint
    metrics_app = make_asgi_app()
    app.mount("/metrics", metrics_app)
    
    return app


# Create the application instance
app = create_app()


if __name__ == "__main__":
    import uvicorn
    
    settings = get_settings()
    uvicorn.run(
        "aiops.main:app",
        host="0.0.0.0",
        port=8000,
        reload=settings.debug,
        log_level="info"
    )