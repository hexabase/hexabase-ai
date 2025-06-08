"""
Authentication endpoints.
"""

import structlog
from fastapi import APIRouter, HTTPException

from aiops.auth.jwt_validator import get_jwt_validator
from aiops.auth.models import TokenValidationRequest, TokenValidationResponse
from aiops.core.exceptions import AuthenticationError

logger = structlog.get_logger(__name__)

router = APIRouter()


@router.post("/validate", response_model=TokenValidationResponse)
async def validate_token(request: TokenValidationRequest) -> TokenValidationResponse:
    """
    Validate JWT token and return authentication context.
    
    This endpoint is used by other services to validate AIOps tokens.
    """
    try:
        jwt_validator = get_jwt_validator()
        auth_context = jwt_validator.validate_token(request.token)
        
        return TokenValidationResponse(
            valid=True,
            auth_context=auth_context
        )
        
    except AuthenticationError as e:
        logger.warning("Token validation failed", error=str(e))
        return TokenValidationResponse(
            valid=False,
            error=str(e)
        )
    
    except Exception as e:
        logger.error("Token validation error", error=str(e))
        return TokenValidationResponse(
            valid=False,
            error="Internal validation error"
        )