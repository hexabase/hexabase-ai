"""
JWT token validation for AIOps security sandbox.
"""

import time
from typing import Optional

import jwt
import structlog
from jwt.exceptions import InvalidTokenError

from aiops.auth.models import AuthContext, JWTClaims, Permission, PlanType
from aiops.core.config import get_settings
from aiops.core.exceptions import AuthenticationError
from aiops.core.metrics import record_jwt_validation

logger = structlog.get_logger(__name__)


class JWTValidator:
    """JWT token validator with security sandbox enforcement."""
    
    def __init__(self) -> None:
        self.settings = get_settings()
    
    def validate_token(self, token: str) -> AuthContext:
        """
        Validate JWT token and return authentication context.
        
        Enforces the AIOps security sandbox:
        - Validates token signature and expiration
        - Ensures token is from trusted issuer (Hexabase Control Plane)
        - Verifies workspace isolation requirements
        - Checks token age limits
        """
        try:
            # Decode and validate JWT
            payload = jwt.decode(
                token,
                self.settings.jwt_secret_key,
                algorithms=[self.settings.jwt_algorithm],
                issuer=self.settings.jwt_issuer,
                audience=self.settings.jwt_audience,
                options={
                    "verify_exp": True,
                    "verify_iat": True,
                    "verify_iss": True,
                    "verify_aud": True,
                }
            )
            
            # Parse claims
            claims = JWTClaims(**payload)
            
            # Additional security checks
            self._validate_token_age(claims)
            self._validate_workspace_isolation(claims)
            
            # Create auth context
            auth_context = AuthContext(
                user_id=claims.sub,
                workspace_id=claims.workspace_id,
                permissions=claims.permissions,
                plan_type=claims.plan_type,
                org_id=claims.org_id
            )
            
            logger.info(
                "JWT validation successful",
                user_id=auth_context.user_id,
                workspace_id=auth_context.workspace_id,
                permissions=[p.value for p in auth_context.permissions],
                plan_type=auth_context.plan_type.value
            )
            
            record_jwt_validation("success")
            return auth_context
            
        except InvalidTokenError as e:
            logger.warning("JWT validation failed", error=str(e))
            record_jwt_validation("invalid_token")
            raise AuthenticationError(f"Invalid token: {str(e)}")
        
        except Exception as e:
            logger.error("JWT validation error", error=str(e))
            record_jwt_validation("error")
            raise AuthenticationError(f"Token validation failed: {str(e)}")
    
    def _validate_token_age(self, claims: JWTClaims) -> None:
        """Validate token age is within acceptable limits."""
        current_time = int(time.time())
        token_age = current_time - claims.iat
        
        if token_age > self.settings.max_token_age_seconds:
            raise AuthenticationError(
                "Token too old",
                details={"token_age": token_age, "max_age": self.settings.max_token_age_seconds}
            )
    
    def _validate_workspace_isolation(self, claims: JWTClaims) -> None:
        """Validate workspace isolation requirements."""
        if not claims.workspace_id:
            raise AuthenticationError(
                "Token missing workspace_id claim",
                details={"required_claim": "workspace_id"}
            )
        
        # Ensure workspace_id is a valid UUID format
        import uuid
        try:
            uuid.UUID(claims.workspace_id)
        except ValueError:
            raise AuthenticationError(
                "Invalid workspace_id format",
                details={"workspace_id": claims.workspace_id}
            )


# Global validator instance
_jwt_validator: Optional[JWTValidator] = None


def get_jwt_validator() -> JWTValidator:
    """Get JWT validator instance (singleton)."""
    global _jwt_validator
    if _jwt_validator is None:
        _jwt_validator = JWTValidator()
    return _jwt_validator