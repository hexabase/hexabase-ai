"""Test cases for authentication integration."""

import pytest
from unittest.mock import Mock, patch, AsyncMock
from datetime import datetime, timedelta
import jwt
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.backends import default_backend

from hexabase_ai.auth import TokenManager, AuthProvider
from hexabase_ai.exceptions import AuthenticationError, TokenExpiredError


class TestTokenManager:
    """Test cases for TokenManager."""

    @pytest.fixture
    def token_manager(self):
        """Create a token manager instance."""
        return TokenManager()

    @pytest.fixture
    def mock_rsa_keys(self):
        """Generate mock RSA keys for testing."""
        private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=2048,
            backend=default_backend()
        )
        public_key = private_key.public_key()
        
        return {
            "private": private_key,
            "public": public_key,
            "public_pem": public_key.public_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PublicFormat.SubjectPublicKeyInfo
            ).decode('utf-8')
        }

    def test_store_and_retrieve_token(self, token_manager):
        """Test storing and retrieving access token."""
        token = "test-access-token"
        expires_in = 3600
        
        token_manager.store_token(token, expires_in)
        
        assert token_manager.get_token() == token
        assert token_manager.is_token_valid() is True

    def test_token_expiration(self, token_manager):
        """Test token expiration handling."""
        token = "test-access-token"
        expires_in = 1  # 1 second
        
        token_manager.store_token(token, expires_in)
        
        # Token should be valid immediately
        assert token_manager.is_token_valid() is True
        
        # Mock time passing
        with patch('time.time', return_value=token_manager._token_expires_at + 1):
            assert token_manager.is_token_valid() is False
            
            with pytest.raises(TokenExpiredError):
                token_manager.get_token()

    def test_refresh_token_before_expiry(self, token_manager):
        """Test that token is considered expired before actual expiry for refresh."""
        token = "test-access-token"
        expires_in = 3600  # 1 hour
        
        token_manager.store_token(token, expires_in)
        
        # Mock time to 5 minutes before expiry
        refresh_time = token_manager._token_expires_at - 300
        with patch('time.time', return_value=refresh_time):
            assert token_manager.should_refresh() is True

    @pytest.mark.asyncio
    async def test_validate_jwt_token(self, token_manager, mock_rsa_keys):
        """Test JWT token validation."""
        # Create a valid JWT token
        payload = {
            "sub": "user-123",
            "exp": datetime.utcnow() + timedelta(hours=1),
            "iat": datetime.utcnow(),
            "scope": "functions:execute"
        }
        
        token = jwt.encode(
            payload,
            mock_rsa_keys["private"],
            algorithm="RS256"
        )
        
        # Mock fetching public key
        with patch.object(token_manager, "fetch_public_key") as mock_fetch:
            mock_fetch.return_value = mock_rsa_keys["public_pem"]
            
            decoded = await token_manager.validate_jwt(token)
            
            assert decoded["sub"] == "user-123"
            assert "scope" in decoded

    @pytest.mark.asyncio
    async def test_validate_expired_jwt(self, token_manager, mock_rsa_keys):
        """Test validation of expired JWT token."""
        # Create an expired JWT token
        payload = {
            "sub": "user-123",
            "exp": datetime.utcnow() - timedelta(hours=1),
            "iat": datetime.utcnow() - timedelta(hours=2)
        }
        
        token = jwt.encode(
            payload,
            mock_rsa_keys["private"],
            algorithm="RS256"
        )
        
        with patch.object(token_manager, "fetch_public_key") as mock_fetch:
            mock_fetch.return_value = mock_rsa_keys["public_pem"]
            
            with pytest.raises(TokenExpiredError, match="Token has expired"):
                await token_manager.validate_jwt(token)


class TestAuthProvider:
    """Test cases for AuthProvider."""

    @pytest.fixture
    def auth_provider(self):
        """Create an auth provider instance."""
        return AuthProvider(
            api_key="test-api-key",
            base_url="https://api.hexabase.io"
        )

    @pytest.mark.asyncio
    async def test_authenticate_with_api_key(self, auth_provider):
        """Test authentication with API key."""
        with patch("httpx.AsyncClient") as mock_client:
            mock_response = Mock()
            mock_response.status_code = 200
            mock_response.json.return_value = {
                "access_token": "jwt-token",
                "expires_in": 3600,
                "token_type": "Bearer"
            }
            
            mock_client.return_value.__aenter__.return_value.post.return_value = mock_response
            
            token_data = await auth_provider.authenticate()
            
            assert token_data["access_token"] == "jwt-token"
            assert token_data["expires_in"] == 3600

    @pytest.mark.asyncio
    async def test_authenticate_with_invalid_key(self, auth_provider):
        """Test authentication with invalid API key."""
        with patch("httpx.AsyncClient") as mock_client:
            mock_response = Mock()
            mock_response.status_code = 401
            mock_response.json.return_value = {
                "error": "Invalid API key",
                "code": "INVALID_API_KEY"
            }
            
            mock_client.return_value.__aenter__.return_value.post.return_value = mock_response
            
            with pytest.raises(AuthenticationError, match="Invalid API key"):
                await auth_provider.authenticate()

    @pytest.mark.asyncio
    async def test_auto_refresh_token(self, auth_provider):
        """Test automatic token refresh."""
        # Initial authentication
        with patch("httpx.AsyncClient") as mock_client:
            mock_response = Mock()
            mock_response.status_code = 200
            mock_response.json.return_value = {
                "access_token": "initial-token",
                "expires_in": 300  # 5 minutes
            }
            
            mock_client.return_value.__aenter__.return_value.post.return_value = mock_response
            
            await auth_provider.authenticate()
            assert auth_provider.get_token() == "initial-token"
        
        # Mock time passing and refresh
        with patch('time.time', return_value=auth_provider._token_manager._token_expires_at - 100):
            with patch("httpx.AsyncClient") as mock_client:
                mock_response = Mock()
                mock_response.status_code = 200
                mock_response.json.return_value = {
                    "access_token": "refreshed-token",
                    "expires_in": 3600
                }
                
                mock_client.return_value.__aenter__.return_value.post.return_value = mock_response
                
                # Should trigger refresh
                token = await auth_provider.get_valid_token()
                assert token == "refreshed-token"

    @pytest.mark.asyncio
    async def test_concurrent_token_refresh(self, auth_provider):
        """Test that concurrent token refresh requests don't cause multiple API calls."""
        refresh_count = 0
        
        async def mock_authenticate():
            nonlocal refresh_count
            refresh_count += 1
            return {
                "access_token": f"token-{refresh_count}",
                "expires_in": 3600
            }
        
        auth_provider.authenticate = mock_authenticate
        
        # Force token to need refresh
        auth_provider._token_manager._token_expires_at = 0
        
        # Make concurrent requests
        tasks = [auth_provider.get_valid_token() for _ in range(10)]
        tokens = await asyncio.gather(*tasks)
        
        # All should get the same token (only one refresh)
        assert refresh_count == 1
        assert all(token == "token-1" for token in tokens)

    @pytest.mark.asyncio
    async def test_auth_header_injection(self, auth_provider):
        """Test automatic auth header injection in requests."""
        auth_provider._token_manager.store_token("test-token", 3600)
        
        headers = await auth_provider.get_auth_headers()
        
        assert headers["Authorization"] == "Bearer test-token"
        assert headers["X-API-Key"] == "test-api-key"