"""
Tests for the chat router.
"""

import pytest
from fastapi.testclient import TestClient
from unittest.mock import Mock, patch

from aiops.main import app
from aiops.auth.models import AuthContext, Permission


@pytest.fixture
def client():
    """Create a test client."""
    return TestClient(app)


@pytest.fixture
def mock_auth_context():
    """Create a mock auth context."""
    return AuthContext(
        user_id="test-user-123",
        org_ids=["org-123"],
        workspace_id="ws-123",
        permissions=[Permission.READ, Permission.WRITE]
    )


def test_chat_endpoint(client, mock_auth_context):
    """Test the /v1/chat endpoint."""
    with patch("aiops.auth.middleware.get_auth_context") as mock_get_auth:
        mock_get_auth.return_value = mock_auth_context
        
        # Prepare request
        request_data = {
            "messages": [
                {"role": "user", "content": "Hello, how can you help me?"}
            ],
            "model": "llama3.2",
            "temperature": 0.7
        }
        
        # Make request
        response = client.post("/v1/chat", json=request_data)
        
        # Assert response
        assert response.status_code == 200
        data = response.json()
        assert "message" in data
        assert "session_id" in data
        assert data["message"]["role"] == "assistant"
        assert len(data["session_id"]) > 0


def test_chat_with_session_id(client, mock_auth_context):
    """Test chat with existing session ID."""
    with patch("aiops.auth.middleware.get_auth_context") as mock_get_auth:
        mock_get_auth.return_value = mock_auth_context
        
        session_id = "existing-session-123"
        request_data = {
            "messages": [
                {"role": "user", "content": "Continue our conversation"}
            ],
            "session_id": session_id
        }
        
        response = client.post("/v1/chat", json=request_data)
        
        assert response.status_code == 200
        data = response.json()
        assert data["session_id"] == session_id


def test_get_chat_history(client, mock_auth_context):
    """Test getting chat history."""
    with patch("aiops.auth.middleware.get_auth_context") as mock_get_auth:
        mock_get_auth.return_value = mock_auth_context
        
        session_id = "test-session-123"
        response = client.get(f"/v1/chat/sessions/{session_id}/history")
        
        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)


def test_delete_chat_session(client, mock_auth_context):
    """Test deleting a chat session."""
    with patch("aiops.auth.middleware.get_auth_context") as mock_get_auth:
        mock_get_auth.return_value = mock_auth_context
        
        session_id = "test-session-123"
        response = client.delete(f"/v1/chat/sessions/{session_id}")
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "deleted"
        assert data["session_id"] == session_id


def test_chat_without_permission(client):
    """Test chat endpoint without proper permissions."""
    with patch("aiops.auth.middleware.get_auth_context") as mock_get_auth:
        # Create auth context without READ permission
        auth_context = AuthContext(
            user_id="test-user-123",
            org_ids=["org-123"],
            workspace_id="ws-123",
            permissions=[]  # No permissions
        )
        mock_get_auth.return_value = auth_context
        
        request_data = {
            "messages": [
                {"role": "user", "content": "Hello"}
            ]
        }
        
        response = client.post("/v1/chat", json=request_data)
        
        # Should fail due to missing permission
        assert response.status_code == 403