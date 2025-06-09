"""Test cases for function deployment and execution."""

import pytest
from unittest.mock import Mock, patch, AsyncMock
from pathlib import Path
import tempfile
import json

from hexabase_ai import HexabaseClient
from hexabase_ai.functions import Function, FunctionDeployment, FunctionExecution
from hexabase_ai.exceptions import (
    FunctionNotFoundError,
    FunctionExecutionError,
    ValidationError,
)


class TestFunctionDeployment:
    """Test cases for function deployment."""

    @pytest.fixture
    def client(self):
        """Create authenticated test client."""
        client = HexabaseClient(api_key="test-key")
        client._access_token = "test-token"
        return client

    @pytest.fixture
    def sample_function(self):
        """Create a sample function for testing."""
        return """
def handler(event, context):
    name = event.get('name', 'World')
    return {
        'statusCode': 200,
        'body': f'Hello, {name}!'
    }
"""

    @pytest.mark.asyncio
    async def test_deploy_function_from_string(self, client):
        """Test deploying a function from string."""
        function_code = "def handler(event, context): return {'status': 'ok'}"
        
        with patch.object(client, "_make_request") as mock_request:
            mock_request.return_value = {
                "function_id": "func-123",
                "name": "test-function",
                "version": "v1",
                "endpoint": "https://api.hexabase.io/functions/func-123"
            }
            
            deployment = await client.deploy_function(
                name="test-function",
                code=function_code,
                runtime="python3.9",
                handler="handler"
            )
            
            assert deployment.function_id == "func-123"
            assert deployment.name == "test-function"
            assert deployment.version == "v1"
            assert deployment.endpoint == "https://api.hexabase.io/functions/func-123"

    @pytest.mark.asyncio
    async def test_deploy_function_from_file(self, client, sample_function):
        """Test deploying a function from file."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
            f.write(sample_function)
            f.flush()
            
            try:
                with patch.object(client, "_make_request") as mock_request:
                    mock_request.return_value = {
                        "function_id": "func-456",
                        "name": "file-function",
                        "version": "v1",
                        "endpoint": "https://api.hexabase.io/functions/func-456"
                    }
                    
                    deployment = await client.deploy_function(
                        name="file-function",
                        file_path=f.name,
                        runtime="python3.9"
                    )
                    
                    assert deployment.function_id == "func-456"
                    assert deployment.name == "file-function"
            finally:
                Path(f.name).unlink()

    @pytest.mark.asyncio
    async def test_deploy_function_with_dependencies(self, client):
        """Test deploying a function with dependencies."""
        function_code = "def handler(event, context): return {'status': 'ok'}"
        requirements = ["requests==2.28.0", "pandas==1.5.0"]
        
        with patch.object(client, "_make_request") as mock_request:
            mock_request.return_value = {
                "function_id": "func-789",
                "name": "deps-function",
                "version": "v1",
                "endpoint": "https://api.hexabase.io/functions/func-789"
            }
            
            deployment = await client.deploy_function(
                name="deps-function",
                code=function_code,
                runtime="python3.9",
                dependencies=requirements
            )
            
            # Verify the request included dependencies
            call_args = mock_request.call_args
            assert call_args[1]["json"]["dependencies"] == requirements

    @pytest.mark.asyncio
    async def test_deploy_function_validation_errors(self, client):
        """Test function deployment validation."""
        # Test missing code
        with pytest.raises(ValidationError, match="Either code or file_path must be provided"):
            await client.deploy_function(name="test", runtime="python3.9")
        
        # Test invalid runtime
        with pytest.raises(ValidationError, match="Unsupported runtime"):
            await client.deploy_function(
                name="test",
                code="def handler(): pass",
                runtime="invalid-runtime"
            )
        
        # Test missing handler for Python runtime
        with pytest.raises(ValidationError, match="Handler is required for Python runtime"):
            await client.deploy_function(
                name="test",
                code="def handler(): pass",
                runtime="python3.9",
                handler=None
            )


class TestFunctionExecution:
    """Test cases for function execution."""

    @pytest.fixture
    def client(self):
        """Create authenticated test client."""
        client = HexabaseClient(api_key="test-key")
        client._access_token = "test-token"
        return client

    @pytest.fixture
    def deployed_function(self):
        """Create a mock deployed function."""
        return FunctionDeployment(
            function_id="func-123",
            name="test-function",
            version="v1",
            endpoint="https://api.hexabase.io/functions/func-123",
            created_at="2025-06-10T00:00:00Z"
        )

    @pytest.mark.asyncio
    async def test_execute_function_success(self, client, deployed_function):
        """Test successful function execution."""
        with patch.object(client, "_make_request") as mock_request:
            mock_request.return_value = {
                "execution_id": "exec-123",
                "status": "completed",
                "result": {"message": "Hello, World!"},
                "duration_ms": 150,
                "billed_duration_ms": 200
            }
            
            result = await client.execute_function(
                function_id="func-123",
                payload={"name": "World"}
            )
            
            assert result.execution_id == "exec-123"
            assert result.status == "completed"
            assert result.result == {"message": "Hello, World!"}
            assert result.duration_ms == 150

    @pytest.mark.asyncio
    async def test_execute_function_async(self, client):
        """Test asynchronous function execution."""
        with patch.object(client, "_make_request") as mock_request:
            mock_request.return_value = {
                "execution_id": "exec-456",
                "status": "pending",
                "message": "Function execution started"
            }
            
            result = await client.execute_function(
                function_id="func-123",
                payload={"data": "test"},
                async_execution=True
            )
            
            assert result.execution_id == "exec-456"
            assert result.status == "pending"

    @pytest.mark.asyncio
    async def test_execute_function_not_found(self, client):
        """Test executing non-existent function."""
        with patch.object(client, "_make_request") as mock_request:
            mock_request.side_effect = FunctionNotFoundError("Function not found")
            
            with pytest.raises(FunctionNotFoundError):
                await client.execute_function(
                    function_id="non-existent",
                    payload={}
                )

    @pytest.mark.asyncio
    async def test_execute_function_timeout(self, client):
        """Test function execution timeout."""
        with patch.object(client, "_make_request") as mock_request:
            mock_request.return_value = {
                "execution_id": "exec-789",
                "status": "timeout",
                "error": "Function execution timed out after 30 seconds"
            }
            
            result = await client.execute_function(
                function_id="func-123",
                payload={},
                timeout=30
            )
            
            assert result.status == "timeout"
            assert "timed out" in result.error

    @pytest.mark.asyncio
    async def test_get_execution_status(self, client):
        """Test getting execution status."""
        with patch.object(client, "_make_request") as mock_request:
            mock_request.return_value = {
                "execution_id": "exec-123",
                "status": "completed",
                "result": {"data": "processed"},
                "started_at": "2025-06-10T00:00:00Z",
                "completed_at": "2025-06-10T00:00:05Z"
            }
            
            status = await client.get_execution_status("exec-123")
            
            assert status.execution_id == "exec-123"
            assert status.status == "completed"
            assert status.result == {"data": "processed"}