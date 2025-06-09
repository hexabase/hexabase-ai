"""Main client for Hexabase AI SDK."""

import os
import asyncio
from typing import Optional, Dict, Any, List
from contextlib import asynccontextmanager
import httpx
import structlog
from tenacity import (
    retry,
    stop_after_attempt,
    wait_exponential,
    retry_if_exception_type,
)

from .auth import AuthProvider
from .functions import FunctionManager
from .models import FunctionDeployment, FunctionExecution, Function, CleanupPolicy
from .exceptions import (
    HexabaseError,
    AuthenticationError,
    FunctionNotFoundError,
    FunctionExecutionError,
    NetworkError,
    ValidationError,
)

logger = structlog.get_logger(__name__)


class HexabaseClient:
    """Main client for interacting with Hexabase AI API."""
    
    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: str = "https://api.hexabase.io",
        timeout: float = 30.0,
        max_retries: int = 3,
    ):
        """Initialize Hexabase AI client.
        
        Args:
            api_key: API key for authentication. If not provided, will look for
                    HEXABASE_API_KEY environment variable.
            base_url: Base URL for the API.
            timeout: Request timeout in seconds.
            max_retries: Maximum number of retry attempts for failed requests.
        """
        # Get API key from environment if not provided
        self.api_key = api_key or os.environ.get("HEXABASE_API_KEY")
        if not self.api_key:
            raise ValueError(
                "API key is required. Provide it as parameter or set HEXABASE_API_KEY environment variable."
            )
            
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.max_retries = max_retries
        
        # Initialize components
        self._auth_provider = AuthProvider(self.api_key, self.base_url)
        self._function_manager = FunctionManager(self)
        self._http_client: Optional[httpx.AsyncClient] = None
        self._access_token: Optional[str] = None
        
        logger.info("Hexabase client initialized", base_url=self.base_url)
        
    @property
    def is_authenticated(self) -> bool:
        """Check if client is authenticated."""
        try:
            self._auth_provider.get_token()
            return True
        except Exception:
            return False
            
    async def authenticate(self) -> None:
        """Authenticate with Hexabase AI API."""
        await self._auth_provider.authenticate()
        
    async def __aenter__(self):
        """Async context manager entry."""
        await self.authenticate()
        return self
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self._cleanup()
        
    async def _cleanup(self) -> None:
        """Cleanup resources."""
        # Stop background cleanup if running
        if hasattr(self._function_manager, "cleanup_manager"):
            await self._function_manager.cleanup_manager.stop_background_cleanup()
            
        # Close HTTP client
        if self._http_client:
            await self._http_client.aclose()
            
    @retry(
        stop=stop_after_attempt(3),
        wait=wait_exponential(multiplier=1, min=4, max=10),
        retry=retry_if_exception_type(NetworkError),
    )
    async def _make_request(
        self,
        method: str,
        path: str,
        **kwargs
    ) -> Dict[str, Any]:
        """Make HTTP request to API with retry logic."""
        if not self._http_client:
            self._http_client = httpx.AsyncClient(timeout=self.timeout)
            
        # Get auth headers
        headers = await self._auth_provider.get_auth_headers()
        if "headers" in kwargs:
            kwargs["headers"].update(headers)
        else:
            kwargs["headers"] = headers
            
        url = f"{self.base_url}{path}"
        
        try:
            response = await self._http_client.request(method, url, **kwargs)
            
            # Handle specific error codes
            if response.status_code == 401:
                raise AuthenticationError("Authentication failed")
            elif response.status_code == 404:
                raise FunctionNotFoundError(f"Resource not found: {path}")
            elif response.status_code == 422:
                error_data = response.json()
                raise ValidationError(
                    error_data.get("message", "Validation failed"),
                    details=error_data.get("errors", {})
                )
            elif response.status_code >= 500:
                raise NetworkError(f"Server error: {response.status_code}")
                
            response.raise_for_status()
            return response.json()
            
        except httpx.TimeoutException:
            raise NetworkError(f"Request timed out: {method} {url}")
        except httpx.HTTPError as e:
            raise NetworkError(f"HTTP error: {str(e)}")
            
    # Function deployment and execution methods
    async def deploy_function(
        self,
        name: str,
        code: Optional[str] = None,
        file_path: Optional[str] = None,
        runtime: str = "python3.9",
        handler: Optional[str] = "handler",
        memory_mb: int = 128,
        timeout_seconds: int = 30,
        environment: Optional[Dict[str, str]] = None,
        dependencies: Optional[List[str]] = None,
        auto_cleanup: Optional[CleanupPolicy] = None,
        cleanup_interval: float = 3600,
    ) -> FunctionDeployment:
        """Deploy a new function to Hexabase AI.
        
        Args:
            name: Function name.
            code: Function code as string.
            file_path: Path to function code file.
            runtime: Runtime environment.
            handler: Function handler.
            memory_mb: Memory allocation in MB.
            timeout_seconds: Execution timeout.
            environment: Environment variables.
            dependencies: Package dependencies.
            auto_cleanup: Auto-cleanup policy.
            cleanup_interval: Cleanup check interval in seconds.
            
        Returns:
            FunctionDeployment object with deployment details.
        """
        deployment = await self._function_manager.deploy_function(
            name=name,
            code=code,
            file_path=file_path,
            runtime=runtime,
            handler=handler,
            memory_mb=memory_mb,
            timeout_seconds=timeout_seconds,
            environment=environment,
            dependencies=dependencies,
            auto_cleanup=auto_cleanup,
        )
        
        # Start background cleanup if policy provided
        if auto_cleanup:
            await self._function_manager.cleanup_manager.start_background_cleanup(
                cleanup_interval
            )
            
        return deployment
        
    async def execute_function(
        self,
        function_id: str,
        payload: Optional[Dict[str, Any]] = None,
        async_execution: bool = False,
        timeout: Optional[int] = None,
    ) -> FunctionExecution:
        """Execute a deployed function.
        
        Args:
            function_id: ID of the function to execute.
            payload: Input payload for the function.
            async_execution: Whether to execute asynchronously.
            timeout: Execution timeout override.
            
        Returns:
            FunctionExecution object with execution details.
        """
        return await self._function_manager.execute_function(
            function_id=function_id,
            payload=payload,
            async_execution=async_execution,
            timeout=timeout,
        )
        
    async def get_function(self, function_id: str) -> Dict[str, Any]:
        """Get function details."""
        return await self._make_request("GET", f"/functions/{function_id}")
        
    async def delete_function(self, function_id: str) -> Dict[str, Any]:
        """Delete a function."""
        return await self._make_request("DELETE", f"/functions/{function_id}")
        
    async def list_functions(
        self,
        limit: int = 100,
        offset: int = 0,
        name_filter: Optional[str] = None,
    ) -> List[Function]:
        """List deployed functions."""
        params = {"limit": limit, "offset": offset}
        if name_filter:
            params["name"] = name_filter
            
        response = await self._make_request("GET", "/functions", params=params)
        return [Function(**func) for func in response.get("functions", [])]
        
    async def get_execution_status(self, execution_id: str) -> FunctionExecution:
        """Get execution status."""
        return await self._function_manager.get_execution_status(execution_id)