"""Test cases for auto-cleanup functionality."""

import pytest
from unittest.mock import Mock, patch, AsyncMock
import asyncio
from datetime import datetime, timedelta

from hexabase_ai import HexabaseClient
from hexabase_ai.functions import AutoCleanupManager, CleanupPolicy


class TestAutoCleanup:
    """Test cases for auto-cleanup functionality."""

    @pytest.fixture
    def client(self):
        """Create authenticated test client."""
        client = HexabaseClient(api_key="test-key")
        client._access_token = "test-token"
        return client

    @pytest.fixture
    def cleanup_manager(self, client):
        """Create a cleanup manager instance."""
        return AutoCleanupManager(client)

    @pytest.mark.asyncio
    async def test_register_function_for_cleanup(self, cleanup_manager):
        """Test registering a function for auto-cleanup."""
        function_id = "func-123"
        policy = CleanupPolicy(
            ttl_hours=24,
            max_executions=100,
            idle_hours=6
        )
        
        cleanup_manager.register_function(function_id, policy)
        
        assert function_id in cleanup_manager._registered_functions
        assert cleanup_manager._registered_functions[function_id] == policy

    @pytest.mark.asyncio
    async def test_cleanup_by_ttl(self, client, cleanup_manager):
        """Test cleanup based on TTL (time-to-live)."""
        function_id = "func-ttl"
        policy = CleanupPolicy(ttl_hours=1)
        
        # Mock function metadata
        with patch.object(client, "get_function") as mock_get:
            mock_get.return_value = {
                "function_id": function_id,
                "created_at": (datetime.utcnow() - timedelta(hours=2)).isoformat()
            }
            
            with patch.object(client, "delete_function") as mock_delete:
                mock_delete.return_value = {"deleted": True}
                
                # Register and run cleanup
                cleanup_manager.register_function(function_id, policy)
                deleted = await cleanup_manager.cleanup_expired_functions()
                
                assert function_id in deleted
                mock_delete.assert_called_once_with(function_id)

    @pytest.mark.asyncio
    async def test_cleanup_by_execution_count(self, client, cleanup_manager):
        """Test cleanup based on execution count."""
        function_id = "func-exec"
        policy = CleanupPolicy(max_executions=5)
        
        # Mock function with execution count
        with patch.object(client, "get_function") as mock_get:
            mock_get.return_value = {
                "function_id": function_id,
                "execution_count": 10
            }
            
            with patch.object(client, "delete_function") as mock_delete:
                mock_delete.return_value = {"deleted": True}
                
                cleanup_manager.register_function(function_id, policy)
                deleted = await cleanup_manager.cleanup_expired_functions()
                
                assert function_id in deleted

    @pytest.mark.asyncio
    async def test_cleanup_by_idle_time(self, client, cleanup_manager):
        """Test cleanup based on idle time."""
        function_id = "func-idle"
        policy = CleanupPolicy(idle_hours=2)
        
        # Mock function with last execution time
        with patch.object(client, "get_function") as mock_get:
            mock_get.return_value = {
                "function_id": function_id,
                "last_executed_at": (datetime.utcnow() - timedelta(hours=3)).isoformat()
            }
            
            with patch.object(client, "delete_function") as mock_delete:
                mock_delete.return_value = {"deleted": True}
                
                cleanup_manager.register_function(function_id, policy)
                deleted = await cleanup_manager.cleanup_expired_functions()
                
                assert function_id in deleted

    @pytest.mark.asyncio
    async def test_cleanup_with_multiple_conditions(self, client, cleanup_manager):
        """Test cleanup with multiple conditions (any condition triggers cleanup)."""
        function_id = "func-multi"
        policy = CleanupPolicy(
            ttl_hours=24,
            max_executions=100,
            idle_hours=6
        )
        
        # Function meets idle condition but not others
        with patch.object(client, "get_function") as mock_get:
            mock_get.return_value = {
                "function_id": function_id,
                "created_at": datetime.utcnow().isoformat(),  # Created recently
                "execution_count": 50,  # Below max
                "last_executed_at": (datetime.utcnow() - timedelta(hours=7)).isoformat()  # Idle
            }
            
            with patch.object(client, "delete_function") as mock_delete:
                mock_delete.return_value = {"deleted": True}
                
                cleanup_manager.register_function(function_id, policy)
                deleted = await cleanup_manager.cleanup_expired_functions()
                
                assert function_id in deleted

    @pytest.mark.asyncio
    async def test_auto_cleanup_background_task(self, client):
        """Test automatic cleanup running in background."""
        cleanup_interval = 0.1  # 100ms for testing
        
        with patch.object(client, "get_function") as mock_get:
            mock_get.return_value = {
                "function_id": "func-auto",
                "created_at": (datetime.utcnow() - timedelta(hours=2)).isoformat()
            }
            
            with patch.object(client, "delete_function") as mock_delete:
                mock_delete.return_value = {"deleted": True}
                
                # Deploy function with auto-cleanup
                with patch.object(client, "_make_request") as mock_request:
                    mock_request.return_value = {
                        "function_id": "func-auto",
                        "name": "auto-cleanup-func"
                    }
                    
                    deployment = await client.deploy_function(
                        name="auto-cleanup-func",
                        code="def handler(): pass",
                        runtime="python3.9",
                        auto_cleanup=CleanupPolicy(ttl_hours=1),
                        cleanup_interval=cleanup_interval
                    )
                    
                    # Wait for cleanup to run
                    await asyncio.sleep(0.2)
                    
                    # Verify function was checked for cleanup
                    assert mock_get.called

    @pytest.mark.asyncio
    async def test_cleanup_error_handling(self, client, cleanup_manager):
        """Test cleanup error handling."""
        function_id = "func-error"
        policy = CleanupPolicy(ttl_hours=1)
        
        with patch.object(client, "get_function") as mock_get:
            mock_get.return_value = {
                "function_id": function_id,
                "created_at": (datetime.utcnow() - timedelta(hours=2)).isoformat()
            }
            
            with patch.object(client, "delete_function") as mock_delete:
                mock_delete.side_effect = Exception("Delete failed")
                
                cleanup_manager.register_function(function_id, policy)
                
                # Should not raise exception
                deleted = await cleanup_manager.cleanup_expired_functions()
                
                # Function should not be in deleted list due to error
                assert function_id not in deleted

    @pytest.mark.asyncio
    async def test_cleanup_preserves_active_functions(self, client, cleanup_manager):
        """Test that cleanup preserves active functions."""
        active_function = "func-active"
        policy = CleanupPolicy(
            ttl_hours=24,
            max_executions=100,
            idle_hours=6
        )
        
        with patch.object(client, "get_function") as mock_get:
            mock_get.return_value = {
                "function_id": active_function,
                "created_at": datetime.utcnow().isoformat(),
                "execution_count": 50,
                "last_executed_at": datetime.utcnow().isoformat()
            }
            
            with patch.object(client, "delete_function") as mock_delete:
                cleanup_manager.register_function(active_function, policy)
                deleted = await cleanup_manager.cleanup_expired_functions()
                
                assert active_function not in deleted
                mock_delete.assert_not_called()