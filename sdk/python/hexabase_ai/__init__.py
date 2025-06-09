"""Hexabase AI Python SDK for dynamic function execution.

This SDK provides a simple interface to deploy and execute serverless functions
on the Hexabase AI platform with built-in authentication and auto-cleanup mechanisms.
"""

from .client import HexabaseClient
from .exceptions import (
    HexabaseError,
    AuthenticationError,
    FunctionNotFoundError,
    FunctionExecutionError,
    ValidationError,
    NetworkError,
    TokenExpiredError,
)
from .functions import (
    Function,
    FunctionDeployment,
    FunctionExecution,
    CleanupPolicy,
    AutoCleanupManager,
)

__version__ = "0.1.0"
__all__ = [
    "HexabaseClient",
    "HexabaseError",
    "AuthenticationError",
    "FunctionNotFoundError",
    "FunctionExecutionError",
    "ValidationError",
    "NetworkError",
    "TokenExpiredError",
    "Function",
    "FunctionDeployment",
    "FunctionExecution",
    "CleanupPolicy",
    "AutoCleanupManager",
]