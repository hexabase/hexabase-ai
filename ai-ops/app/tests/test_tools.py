import pytest
from httpx import Response
from app.agents.tools import (
    GetKubernetesNodesTool, GetKubernetesNodesInput, 
    ScaleDeploymentTool, ScaleDeploymentInput,
    LogQueryingTool, LogQueryInput
)
from app.config import settings

def test_get_kubernetes_nodes_tool_success(httpx_mock):
    # Arrange
    # 1. Mock the internal Go API endpoint.
    workspace_id = "ws-12345"
    go_api_endpoint = f"{settings.HKS_INTERNAL_API_URL}/internal/v1/workspaces/{workspace_id}/nodes"
    
    mock_response_data = [
        {"name": "node-1", "status": "Ready", "cpu": "500m", "memory": "2Gi", "pods": 10},
        {"name": "node-2", "status": "Ready", "cpu": "600m", "memory": "3Gi", "pods": 15},
        {"name": "node-3", "status": "NotReady", "cpu": "0m", "memory": "0Gi", "pods": 0},
    ]
    
    httpx_mock.add_response(
        url=go_api_endpoint,
        method="GET",
        json=mock_response_data,
        status_code=200,
    )

    # 2. Instantiate the tool.
    tool = GetKubernetesNodesTool()
    tool_input = GetKubernetesNodesInput(workspace_id=workspace_id)

    # Act
    # 3. Call the tool's use method.
    result = tool.use(tool_input)

    # Assert
    # 4. Verify the result is a human-readable summary.
    assert "Found 3 nodes for workspace ws-12345." in result
    assert "node-1 is Ready" in result
    assert "node-3 is NotReady" in result
    
    # 5. Verify that the correct API call was made.
    request = httpx_mock.get_request()
    assert request is not None
    assert request.method == "GET"

def test_get_kubernetes_nodes_tool_api_error(httpx_mock):
    # Arrange
    # 1. Mock the internal Go API endpoint to return a 500 server error.
    workspace_id = "ws-500-error"
    go_api_endpoint = f"{settings.HKS_INTERNAL_API_URL}/internal/v1/workspaces/{workspace_id}/nodes"
    
    httpx_mock.add_response(
        url=go_api_endpoint,
        method="GET",
        status_code=500,
        json={"error": "internal server error in HKS API"}
    )

    # 2. Instantiate the tool.
    tool = GetKubernetesNodesTool()
    tool_input = GetKubernetesNodesInput(workspace_id=workspace_id)

    # Act
    # 3. Call the tool's use method.
    result = tool.use(tool_input)

    # Assert
    # 4. Verify the result is a user-friendly error message.
    assert "Error: Could not retrieve node data." in result
    assert "API returned a 500 status" in result 

def test_scale_deployment_tool_success(httpx_mock):
    # Arrange
    # 1. Mock the internal Go API endpoint for scaling.
    workspace_id = "ws-123"
    deployment_name = "my-api"
    replicas = 5
    go_api_endpoint = f"{settings.HKS_INTERNAL_API_URL}/internal/v1/workspaces/{workspace_id}/deployments/{deployment_name}/scale"
    
    httpx_mock.add_response(
        url=go_api_endpoint,
        method="POST",
        status_code=200,
        json={"message": "deployment scaled successfully"}
    )

    # 2. Instantiate the tool.
    tool = ScaleDeploymentTool()
    tool_input = ScaleDeploymentInput(workspace_id=workspace_id, deployment_name=deployment_name, replicas=replicas)

    # Act
    # 3. Call the tool's use method.
    result = tool.use(tool_input)

    # Assert
    # 4. Verify the result is a success message.
    assert "Successfully scaled deployment 'my-api' to 5 replicas" in result
    
    # 5. Verify the API call was made with the correct body.
    request = httpx_mock.get_request()
    assert request is not None
    assert request.json() == {"replicas": 5}


def test_log_querying_tool_success(httpx_mock):
    # Arrange
    # 1. Mock the internal Go API endpoint for log querying.
    workspace_id = "ws-123"
    query = "error 500"
    time_range = "1h"
    go_api_endpoint = f"{settings.HKS_INTERNAL_API_URL}/internal/v1/workspaces/{workspace_id}/logs/query"
    
    mock_response_data = {
        "logs": [
            {
                "timestamp": "2025-06-09T10:00:00Z",
                "level": "ERROR",
                "message": "Internal server error 500",
                "source": "api-gateway"
            },
            {
                "timestamp": "2025-06-09T10:05:00Z",
                "level": "ERROR", 
                "message": "Database connection error 500",
                "source": "backend-service"
            }
        ],
        "total_count": 2
    }
    
    httpx_mock.add_response(
        url=go_api_endpoint,
        method="POST",
        json=mock_response_data,
        status_code=200,
    )

    # 2. Instantiate the tool.
    tool = LogQueryingTool()
    tool_input = LogQueryInput(workspace_id=workspace_id, query=query, time_range=time_range)

    # Act
    # 3. Call the tool's use method.
    result = tool.use(tool_input)

    # Assert
    # 4. Verify the result contains log summary.
    assert "Found 2 log entries" in result
    assert "error 500" in result
    assert "api-gateway" in result
    assert "backend-service" in result
    
    # 5. Verify the API call was made with correct parameters.
    request = httpx_mock.get_request()
    assert request is not None
    assert request.json() == {"query": "error 500", "time_range": "1h"}


def test_log_querying_tool_no_results(httpx_mock):
    # Arrange
    # 1. Mock the internal Go API endpoint to return empty results.
    workspace_id = "ws-123"
    query = "non-existent-error"
    time_range = "1h"
    go_api_endpoint = f"{settings.HKS_INTERNAL_API_URL}/internal/v1/workspaces/{workspace_id}/logs/query"
    
    mock_response_data = {
        "logs": [],
        "total_count": 0
    }
    
    httpx_mock.add_response(
        url=go_api_endpoint,
        method="POST",
        json=mock_response_data,
        status_code=200,
    )

    # 2. Instantiate the tool.
    tool = LogQueryingTool()
    tool_input = LogQueryInput(workspace_id=workspace_id, query=query, time_range=time_range)

    # Act
    # 3. Call the tool's use method.
    result = tool.use(tool_input)

    # Assert
    # 4. Verify the result indicates no logs found.
    assert "No log entries found" in result
    assert "non-existent-error" in result


def test_log_querying_tool_api_error(httpx_mock):
    # Arrange
    # 1. Mock the internal Go API endpoint to return an error.
    workspace_id = "ws-error"
    query = "test"
    time_range = "1h"
    go_api_endpoint = f"{settings.HKS_INTERNAL_API_URL}/internal/v1/workspaces/{workspace_id}/logs/query"
    
    httpx_mock.add_response(
        url=go_api_endpoint,
        method="POST",
        status_code=403,
        json={"error": "unauthorized access to logs"}
    )

    # 2. Instantiate the tool.
    tool = LogQueryingTool()
    tool_input = LogQueryInput(workspace_id=workspace_id, query=query, time_range=time_range)

    # Act
    # 3. Call the tool's use method.
    result = tool.use(tool_input)

    # Assert
    # 4. Verify the result contains error message.
    assert "Error: Could not query logs" in result
    assert "403" in result 