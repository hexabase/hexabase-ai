"""
Custom exceptions for AIOps service.
"""

from typing import Any, Dict, Optional


class AIOpsException(Exception):
    """Base exception for AIOps service."""
    
    def __init__(
        self,
        message: str,
        error_code: str,
        status_code: int = 500,
        details: Optional[Dict[str, Any]] = None
    ) -> None:
        super().__init__(message)
        self.message = message
        self.error_code = error_code
        self.status_code = status_code
        self.details = details or {}


class AuthenticationError(AIOpsException):
    """Authentication-related errors."""
    
    def __init__(self, message: str = "Authentication failed", details: Optional[Dict[str, Any]] = None) -> None:
        super().__init__(
            message=message,
            error_code="AUTH_FAILED",
            status_code=401,
            details=details
        )


class AuthorizationError(AIOpsException):
    """Authorization-related errors."""
    
    def __init__(self, message: str = "Access denied", details: Optional[Dict[str, Any]] = None) -> None:
        super().__init__(
            message=message,
            error_code="ACCESS_DENIED",
            status_code=403,
            details=details
        )


class ValidationError(AIOpsException):
    """Input validation errors."""
    
    def __init__(self, message: str, details: Optional[Dict[str, Any]] = None) -> None:
        super().__init__(
            message=message,
            error_code="VALIDATION_ERROR",
            status_code=400,
            details=details
        )


class ResourceNotFoundError(AIOpsException):
    """Resource not found errors."""
    
    def __init__(self, resource: str, identifier: str) -> None:
        super().__init__(
            message=f"{resource} not found: {identifier}",
            error_code="RESOURCE_NOT_FOUND",
            status_code=404,
            details={"resource": resource, "identifier": identifier}
        )


class ExternalServiceError(AIOpsException):
    """External service integration errors."""
    
    def __init__(self, service: str, message: str, details: Optional[Dict[str, Any]] = None) -> None:
        super().__init__(
            message=f"{service} error: {message}",
            error_code="EXTERNAL_SERVICE_ERROR",
            status_code=503,
            details={**(details or {}), "service": service}
        )


class AIAnalysisError(AIOpsException):
    """AI analysis-related errors."""
    
    def __init__(self, message: str, details: Optional[Dict[str, Any]] = None) -> None:
        super().__init__(
            message=message,
            error_code="AI_ANALYSIS_ERROR",
            status_code=500,
            details=details
        )


class RemediationError(AIOpsException):
    """Automated remediation errors."""
    
    def __init__(self, message: str, details: Optional[Dict[str, Any]] = None) -> None:
        super().__init__(
            message=message,
            error_code="REMEDIATION_ERROR",
            status_code=500,
            details=details
        )


class RateLimitError(AIOpsException):
    """Rate limiting errors."""
    
    def __init__(self, message: str = "Rate limit exceeded") -> None:
        super().__init__(
            message=message,
            error_code="RATE_LIMIT_EXCEEDED",
            status_code=429
        )