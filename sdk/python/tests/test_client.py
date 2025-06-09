"""Test cases for Hexabase AI SDK client."""

import pytest
from unittest.mock import Mock, patch, AsyncMock
import httpx
from hexabase_ai import HexabaseClient
from hexabase_ai.exceptions import (
    AuthenticationError,
    FunctionNotFoundError,
    FunctionExecutionError,
    NetworkError,
)


class TestHexabaseClient:
    """Test cases for HexabaseClient."""

    @pytest.fixture
    def client(self):
        """Create a test client instance."""
        return HexabaseClient(
            api_key="test-api-key",
            base_url="https://api.hexabase.io"
        )

    @pytest.fixture
    def mock_httpx_client(self):
        """Mock httpx AsyncClient."""
        with patch("hexabase_ai.client.httpx.AsyncClient") as mock:
            yield mock

    def test_client_initialization(self):
        """Test client initialization with various parameters."""
        # Test with API key
        client = HexabaseClient(api_key="test-key")
        assert client.api_key == "test-key"
        assert client.base_url == "https://api.hexabase.io"
        
        # Test with custom base URL
        client = HexabaseClient(
            api_key="test-key",
            base_url="https://custom.hexabase.io"
        )
        assert client.base_url == "https://custom.hexabase.io"
        
        # Test with environment variables
        with patch.dict("os.environ", {"HEXABASE_API_KEY": "env-key"}):
            client = HexabaseClient()
            assert client.api_key == "env-key"

    def test_client_without_api_key_raises_error(self):
        """Test that client raises error without API key."""
        with pytest.raises(ValueError, match="API key is required"):
            HexabaseClient(api_key=None)

    @pytest.mark.asyncio
    async def test_authenticate_success(self, client, mock_httpx_client):
        """Test successful authentication."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "access_token": "test-token",
            "expires_in": 3600
        }
        
        mock_httpx_client.return_value.__aenter__.return_value.post.return_value = mock_response
        
        await client.authenticate()
        assert client._access_token == "test-token"
        assert client.is_authenticated is True

    @pytest.mark.asyncio
    async def test_authenticate_failure(self, client, mock_httpx_client):
        """Test authentication failure."""
        mock_response = Mock()
        mock_response.status_code = 401
        mock_response.json.return_value = {"error": "Invalid API key"}
        
        mock_httpx_client.return_value.__aenter__.return_value.post.return_value = mock_response
        
        with pytest.raises(AuthenticationError, match="Authentication failed"):
            await client.authenticate()

    @pytest.mark.asyncio
    async def test_auto_cleanup_on_exit(self, client):
        """Test that client cleans up resources on exit."""
        cleanup_called = False
        
        async def mock_cleanup():
            nonlocal cleanup_called
            cleanup_called = True
        
        client._cleanup = mock_cleanup
        
        async with client:
            pass
        
        assert cleanup_called is True