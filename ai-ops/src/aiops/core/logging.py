"""
Structured logging configuration for AIOps service.
"""

import logging
import sys
from typing import Any, Dict

import structlog
from structlog.typing import Processor

from aiops.core.config import Settings


def setup_logging(settings: Settings) -> None:
    """Setup structured logging with ClickHouse integration."""
    
    # Configure standard library logging
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=logging.DEBUG if settings.debug else logging.INFO,
    )
    
    # Structlog processors
    shared_processors: list[Processor] = [
        structlog.contextvars.merge_contextvars,
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_log_level,
        structlog.stdlib.add_logger_name,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
    ]
    
    if settings.debug:
        # Pretty console output for development
        shared_processors.extend([
            structlog.dev.ConsoleRenderer(colors=True)
        ])
    else:
        # JSON output for production
        shared_processors.extend([
            structlog.processors.dict_tracebacks,
            structlog.processors.JSONRenderer()
        ])
    
    structlog.configure(
        processors=shared_processors,
        wrapper_class=structlog.stdlib.BoundLogger,
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True,
    )


def get_logger(name: str) -> structlog.stdlib.BoundLogger:
    """Get a structured logger instance."""
    return structlog.get_logger(name)


def add_workspace_context(workspace_id: str) -> None:
    """Add workspace context to all logs in current request."""
    structlog.contextvars.bind_contextvars(workspace_id=workspace_id)


def add_user_context(user_id: str) -> None:
    """Add user context to all logs in current request."""
    structlog.contextvars.bind_contextvars(user_id=user_id)


def add_operation_context(operation: str, **kwargs: Any) -> None:
    """Add operation context to all logs in current request."""
    structlog.contextvars.bind_contextvars(
        operation=operation,
        **kwargs
    )