"""
Authentication and authorization models.
"""

from datetime import datetime
from enum import Enum
from typing import List, Optional

from pydantic import BaseModel, Field


class PlanType(str, Enum):
    """Workspace plan types."""
    SHARED = "shared"
    DEDICATED = "dedicated"


class Permission(str, Enum):
    """AIOps permissions."""
    READ = "aiops:read"
    ANALYZE = "aiops:analyze" 
    REMEDIATE = "aiops:remediate"
    ADMIN = "aiops:admin"


class JWTClaims(BaseModel):
    """JWT token claims."""
    sub: str = Field(..., description="User ID (subject)")
    workspace_id: str = Field(..., description="Workspace UUID")
    permissions: List[Permission] = Field(default_factory=list, description="Granted permissions")
    plan_type: PlanType = Field(..., description="Workspace plan type")
    exp: int = Field(..., description="Expiration timestamp")
    iat: int = Field(..., description="Issued at timestamp")
    iss: str = Field(..., description="Token issuer")
    aud: str = Field(..., description="Token audience")
    
    # Optional claims
    org_id: Optional[str] = Field(None, description="Organization ID")
    user_email: Optional[str] = Field(None, description="User email")
    user_name: Optional[str] = Field(None, description="User display name")


class AuthContext(BaseModel):
    """Authentication context for a request."""
    user_id: str
    workspace_id: str
    permissions: List[Permission]
    plan_type: PlanType
    org_id: Optional[str] = None
    
    def has_permission(self, permission: Permission) -> bool:
        """Check if user has a specific permission."""
        return permission in self.permissions or Permission.ADMIN in self.permissions
    
    def require_permission(self, permission: Permission) -> None:
        """Raise exception if user doesn't have permission."""
        if not self.has_permission(permission):
            from aiops.core.exceptions import AuthorizationError
            raise AuthorizationError(
                f"Permission required: {permission.value}",
                details={"required_permission": permission.value, "user_permissions": [p.value for p in self.permissions]}
            )


class TokenValidationRequest(BaseModel):
    """Request to validate JWT token."""
    token: str = Field(..., description="JWT token to validate")


class TokenValidationResponse(BaseModel):
    """Response from token validation."""
    valid: bool = Field(..., description="Whether token is valid")
    auth_context: Optional[AuthContext] = Field(None, description="Authentication context if valid")
    error: Optional[str] = Field(None, description="Error message if invalid")