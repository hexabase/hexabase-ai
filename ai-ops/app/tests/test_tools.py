import pytest
from httpx import Response
from app.agents.tools import GetKubernetesNodesTool, GetKubernetesNodesInput, ScaleDeploymentTool, ScaleDeploymentInput
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