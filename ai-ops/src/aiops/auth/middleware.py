"""
JWT authentication middleware for FastAPI.
"""

import time
from typing import Optional

import structlog
from fastapi import Request, Response
from starlette.middleware.base import BaseHTTPMiddleware

from aiops.auth.jwt_validator import get_jwt_validator
from aiops.auth.models import AuthContext
from aiops.core.exceptions import AuthenticationError
from aiops.core.logging import add_user_context, add_workspace_context
from aiops.core.metrics import record_request

logger = structlog.get_logger(__name__)


class JWTAuthMiddleware(BaseHTTPMiddleware):
    """JWT authentication middleware."""
    
    # Paths that don't require authentication
    EXEMPT_PATHS = {
        "/health",
        "/health/",
        "/metrics",
        "/docs",
        "/redoc",
        "/openapi.json"
    }
    
    async def dispatch(self, request: Request, call_next) -> Response:
        """Process request with JWT authentication."""
        start_time = time.time()
        
        # Skip auth for exempt paths
        if self._is_exempt_path(request.url.path):
            response = await call_next(request)
            self._record_metrics(request, response, start_time, "public")
            return response
        
        try:
            # Extract and validate JWT token
            auth_context = await self._authenticate_request(request)
            
            # Add authentication context to request state
            request.state.auth = auth_context
            
            # Add context to logs
            add_user_context(auth_context.user_id)
            add_workspace_context(auth_context.workspace_id)
            
            # Process request
            response = await call_next(request)
            
            # Record metrics
            self._record_metrics(request, response, start_time, auth_context.workspace_id)
            
            return response
            
        except AuthenticationError as e:
            logger.warning("Authentication failed", error=str(e), path=request.url.path)
            
            # Create error response
            response = Response(
                content=f'{{"error": "AUTH_FAILED", "message": "{str(e)}"}}',
                status_code=401,
                media_type="application/json"
            )
            
            self._record_metrics(request, response, start_time, "unauthenticated")
            return response
        
        except Exception as e:
            logger.error("Middleware error", error=str(e), path=request.url.path)
            
            response = Response(
                content='{"error": "INTERNAL_ERROR", "message": "Authentication middleware error"}',
                status_code=500,
                media_type="application/json"
            )
            
            self._record_metrics(request, response, start_time, "error")
            return response
    
    def _is_exempt_path(self, path: str) -> bool:
        """Check if path is exempt from authentication."""
        return path in self.EXEMPT_PATHS
    
    async def _authenticate_request(self, request: Request) -> AuthContext:
        """Extract and validate JWT token from request."""
        # Extract token from Authorization header
        auth_header = request.headers.get("Authorization")
        if not auth_header:
            raise AuthenticationError("Missing Authorization header")
        
        if not auth_header.startswith("Bearer "):
            raise AuthenticationError("Invalid Authorization header format")
        
        token = auth_header[7:]  # Remove "Bearer " prefix
        
        # Validate token
        jwt_validator = get_jwt_validator()
        return jwt_validator.validate_token(token)
    
    def _record_metrics(
        self, 
        request: Request, 
        response: Response, 
        start_time: float, 
        workspace_id: str
    ) -> None:
        """Record request metrics."""
        duration = time.time() - start_time
        
        record_request(
            method=request.method,
            endpoint=self._normalize_endpoint(request.url.path),
            status_code=response.status_code,
            workspace_id=workspace_id,
            duration=duration
        )
    
    def _normalize_endpoint(self, path: str) -> str:
        """Normalize endpoint path for metrics (remove dynamic segments)."""
        # Replace UUIDs and other dynamic segments with placeholders
        import re
        
        # Replace workspace IDs
        path = re.sub(
            r'/workspaces/[0-9a-f-]{36}',
            '/workspaces/{workspace_id}',
            path
        )
        
        # Replace other UUIDs
        path = re.sub(
            r'/[0-9a-f-]{36}',
            '/{id}',
            path
        )
        
        return path


def get_auth_context(request: Request) -> AuthContext:
    """Get authentication context from request state."""
    if not hasattr(request.state, "auth"):
        raise AuthenticationError("No authentication context available")
    
    return request.state.auth