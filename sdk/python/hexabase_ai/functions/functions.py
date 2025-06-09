"""Functions module for deploying and executing serverless functions."""

import asyncio
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List, Set
from pathlib import Path
import structlog

from ..models import (
    Function,
    FunctionConfig,
    FunctionDeployment,
    FunctionExecution,
    CleanupPolicy,
)
from ..exceptions import ValidationError, FunctionNotFoundError

logger = structlog.get_logger(__name__)


class AutoCleanupManager:
    """Manages automatic cleanup of functions based on policies."""
    
    def __init__(self, client):
        self.client = client
        self._registered_functions: Dict[str, CleanupPolicy] = {}
        self._cleanup_task: Optional[asyncio.Task] = None
        self._running = False
        
    def register_function(self, function_id: str, policy: CleanupPolicy) -> None:
        """Register a function for auto-cleanup."""
        self._registered_functions[function_id] = policy
        logger.info("Function registered for cleanup", 
                   function_id=function_id,
                   policy=policy.dict(exclude_none=True))
        
    def unregister_function(self, function_id: str) -> None:
        """Remove a function from auto-cleanup."""
        if function_id in self._registered_functions:
            del self._registered_functions[function_id]
            logger.info("Function unregistered from cleanup", function_id=function_id)
            
    async def cleanup_expired_functions(self) -> Set[str]:
        """Check and cleanup expired functions based on policies."""
        deleted_functions = set()
        
        for function_id, policy in list(self._registered_functions.items()):
            try:
                should_delete = await self._should_delete_function(function_id, policy)
                
                if should_delete:
                    logger.info("Deleting expired function", function_id=function_id)
                    await self.client.delete_function(function_id)
                    deleted_functions.add(function_id)
                    self.unregister_function(function_id)
                    
            except Exception as e:
                logger.error("Error during cleanup check", 
                           function_id=function_id, 
                           error=str(e))
                           
        return deleted_functions
        
    async def _should_delete_function(self, function_id: str, policy: CleanupPolicy) -> bool:
        """Check if a function should be deleted based on policy."""
        try:
            function_info = await self.client.get_function(function_id)
            
            # Check TTL
            if policy.ttl_hours:
                created_at = datetime.fromisoformat(function_info["created_at"].replace("Z", "+00:00"))
                if datetime.utcnow() > created_at + timedelta(hours=policy.ttl_hours):
                    logger.debug("Function exceeded TTL", function_id=function_id)
                    return True
                    
            # Check execution count
            if policy.max_executions and function_info.get("execution_count", 0) >= policy.max_executions:
                logger.debug("Function exceeded max executions", function_id=function_id)
                return True
                
            # Check idle time
            if policy.idle_hours and "last_executed_at" in function_info:
                last_exec = datetime.fromisoformat(function_info["last_executed_at"].replace("Z", "+00:00"))
                if datetime.utcnow() > last_exec + timedelta(hours=policy.idle_hours):
                    logger.debug("Function exceeded idle time", function_id=function_id)
                    return True
                    
        except FunctionNotFoundError:
            # Function already deleted
            self.unregister_function(function_id)
            return False
            
        return False
        
    async def start_background_cleanup(self, interval_seconds: int = 3600) -> None:
        """Start background cleanup task."""
        if self._cleanup_task and not self._cleanup_task.done():
            logger.warning("Cleanup task already running")
            return
            
        self._running = True
        self._cleanup_task = asyncio.create_task(self._cleanup_loop(interval_seconds))
        logger.info("Started background cleanup", interval_seconds=interval_seconds)
        
    async def stop_background_cleanup(self) -> None:
        """Stop background cleanup task."""
        self._running = False
        if self._cleanup_task:
            self._cleanup_task.cancel()
            try:
                await self._cleanup_task
            except asyncio.CancelledError:
                pass
        logger.info("Stopped background cleanup")
        
    async def _cleanup_loop(self, interval_seconds: int) -> None:
        """Background cleanup loop."""
        while self._running:
            try:
                await asyncio.sleep(interval_seconds)
                if self._running:
                    deleted = await self.cleanup_expired_functions()
                    if deleted:
                        logger.info("Cleanup completed", deleted_count=len(deleted))
            except Exception as e:
                logger.error("Error in cleanup loop", error=str(e))


class FunctionManager:
    """Manages function deployment and execution."""
    
    def __init__(self, client):
        self.client = client
        self.cleanup_manager = AutoCleanupManager(client)
        
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
    ) -> FunctionDeployment:
        """Deploy a new function."""
        # Validate inputs
        if not code and not file_path:
            raise ValidationError("Either code or file_path must be provided")
            
        # Read code from file if provided
        if file_path:
            path = Path(file_path)
            if not path.exists():
                raise ValidationError(f"File not found: {file_path}")
            code = path.read_text()
            
        # Create function config
        config = FunctionConfig(
            name=name,
            runtime=runtime,
            handler=handler,
            memory_mb=memory_mb,
            timeout_seconds=timeout_seconds,
            environment=environment or {},
            dependencies=dependencies or [],
        )
        
        # Deploy function
        deployment_data = {
            "name": config.name,
            "code": code,
            "runtime": config.runtime,
            "handler": config.handler,
            "memory_mb": config.memory_mb,
            "timeout_seconds": config.timeout_seconds,
            "environment": config.environment,
            "dependencies": config.dependencies,
        }
        
        response = await self.client._make_request(
            "POST",
            "/functions",
            json=deployment_data
        )
        
        deployment = FunctionDeployment(**response)
        
        # Register for auto-cleanup if policy provided
        if auto_cleanup:
            self.cleanup_manager.register_function(
                deployment.function_id,
                auto_cleanup
            )
            
        logger.info("Function deployed", 
                   function_id=deployment.function_id,
                   name=deployment.name)
                   
        return deployment
        
    async def execute_function(
        self,
        function_id: str,
        payload: Optional[Dict[str, Any]] = None,
        async_execution: bool = False,
        timeout: Optional[int] = None,
    ) -> FunctionExecution:
        """Execute a deployed function."""
        execution_data = {
            "payload": payload or {},
            "async": async_execution,
        }
        
        if timeout:
            execution_data["timeout"] = timeout
            
        response = await self.client._make_request(
            "POST",
            f"/functions/{function_id}/execute",
            json=execution_data
        )
        
        execution = FunctionExecution(**response)
        
        logger.info("Function executed",
                   function_id=function_id,
                   execution_id=execution.execution_id,
                   status=execution.status)
                   
        return execution
        
    async def get_execution_status(self, execution_id: str) -> FunctionExecution:
        """Get status of a function execution."""
        response = await self.client._make_request(
            "GET",
            f"/executions/{execution_id}"
        )
        
        return FunctionExecution(**response)