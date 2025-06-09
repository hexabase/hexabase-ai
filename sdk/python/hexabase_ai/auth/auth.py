"""Authentication module for Hexabase AI SDK."""

import os
import time
import asyncio
from typing import Optional, Dict, Any
from datetime import datetime, timedelta
import httpx
import jwt
from jose import JWTError
import structlog

from ..exceptions import AuthenticationError, TokenExpiredError, NetworkError

logger = structlog.get_logger(__name__)


class TokenManager:
    """Manages access tokens and their lifecycle."""
    
    def __init__(self):
        self._access_token: Optional[str] = None
        self._token_expires_at: float = 0
        self._refresh_buffer: int = 300  # Refresh 5 minutes before expiry
        self._public_key_cache: Dict[str, str] = {}
        
    def store_token(self, token: str, expires_in: int) -> None:
        """Store access token with expiration time."""
        self._access_token = token
        self._token_expires_at = time.time() + expires_in
        logger.info("Token stored", expires_in=expires_in)
        
    def get_token(self) -> str:
        """Get current access token."""
        if not self.is_token_valid():
            raise TokenExpiredError("Token has expired or is not set")
        return self._access_token
        
    def is_token_valid(self) -> bool:
        """Check if token is valid and not expired."""
        if not self._access_token:
            return False
        return time.time() < self._token_expires_at
        
    def should_refresh(self) -> bool:
        """Check if token should be refreshed."""
        if not self._access_token:
            return True
        return time.time() > (self._token_expires_at - self._refresh_buffer)
        
    async def validate_jwt(self, token: str) -> Dict[str, Any]:
        """Validate JWT token and return decoded payload."""
        try:
            # For SDK, we trust the server's token
            # In production, fetch public key from server
            decoded = jwt.decode(token, options={"verify_signature": False})
            
            # Check expiration
            if "exp" in decoded:
                exp_time = datetime.fromtimestamp(decoded["exp"])
                if exp_time < datetime.utcnow():
                    raise TokenExpiredError("Token has expired")
                    
            return decoded
        except JWTError as e:
            raise AuthenticationError(f"Invalid JWT token: {str(e)}")
            
    async def fetch_public_key(self, key_id: str) -> str:
        """Fetch public key for JWT validation (placeholder)."""
        # In production, this would fetch from /.well-known/jwks.json
        if key_id not in self._public_key_cache:
            # Placeholder - would make actual HTTP request
            self._public_key_cache[key_id] = "public-key-placeholder"
        return self._public_key_cache[key_id]


class AuthProvider:
    """Handles authentication with Hexabase AI API."""
    
    def __init__(self, api_key: str, base_url: str):
        self.api_key = api_key
        self.base_url = base_url
        self._token_manager = TokenManager()
        self._refresh_lock = asyncio.Lock()
        
    async def authenticate(self) -> Dict[str, Any]:
        """Authenticate with API key and get access token."""
        url = f"{self.base_url}/auth/token"
        
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    url,
                    json={"api_key": self.api_key},
                    headers={"Content-Type": "application/json"}
                )
                
                if response.status_code == 401:
                    error_data = response.json()
                    raise AuthenticationError(
                        error_data.get("error", "Authentication failed"),
                        code=error_data.get("code", "AUTH_FAILED")
                    )
                    
                response.raise_for_status()
                token_data = response.json()
                
                # Store token
                self._token_manager.store_token(
                    token_data["access_token"],
                    token_data["expires_in"]
                )
                
                logger.info("Authentication successful")
                return token_data
                
        except httpx.HTTPError as e:
            raise NetworkError(f"Network error during authentication: {str(e)}")
            
    async def get_valid_token(self) -> str:
        """Get a valid access token, refreshing if necessary."""
        async with self._refresh_lock:
            if self._token_manager.should_refresh():
                logger.info("Token needs refresh")
                await self.authenticate()
                
        return self._token_manager.get_token()
        
    def get_token(self) -> str:
        """Get current token without refresh check."""
        return self._token_manager.get_token()
        
    async def get_auth_headers(self) -> Dict[str, str]:
        """Get authentication headers for API requests."""
        token = await self.get_valid_token()
        return {
            "Authorization": f"Bearer {token}",
            "X-API-Key": self.api_key
        }