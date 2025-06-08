import pytest
from httpx import Response
from app.llm_clients.ollama import OllamaClient
import httpx
import json

def test_ollama_client_predict(httpx_mock):
    # Arrange
    # 1. Mock the HTTPX client to intercept outgoing requests to Ollama.
    #    Return a canned JSON response that mimics Ollama's streaming format.
    #    Note: We test the non-streaming case for simplicity here.
    ollama_api_endpoint = "http://ollama-service.ai-ops-llm.svc.cluster.local:11434/api/generate"
    mock_response_content = '{"response": "{\\"tool_name\\": \\"get_kubernetes_nodes\\", \\"tool_input\\": {}}", "done": true}'
    
    httpx_mock.add_response(
        url=ollama_api_endpoint,
        method="POST",
        json=mock_response_content, # httpx handles JSON encoding
        status_code=200,
    )

    # 2. Instantiate the client we are testing.
    client = OllamaClient(
        api_url="http://ollama-service.ai-ops-llm.svc.cluster.local:11434",
        model_name="test-model"
    )

    # Act
    # 3. Call the method we want to test.
    result_json = client.predict(
        system_prompt="You are a helpful assistant.",
        user_query="How many nodes are there?"
    )

    # Assert
    # 4. Verify the result.
    assert result_json == '{"response": "{\\"tool_name\\": \\"get_kubernetes_nodes\\", \\"tool_input\\": {}}", "done": true}'
    
    # 5. Verify that the correct request was made.
    request = httpx_mock.get_request()
    assert request is not None
    assert request.method == "POST"
    
    request_data = request.json()
    assert request_data["model"] == "test-model"
    assert "You are a helpful assistant." in request_data["system"]
    assert "How many nodes are there?" in request_data["prompt"]
    assert request_data["format"] == "json"
    assert request_data["stream"] is False

def test_ollama_client_predict_success(httpx_mock):
    # Arrange
    ollama_api_endpoint = "http://ollama-service.ai-ops-llm.svc.cluster.local:11434/api/generate"
    mock_response_content = {"response": "{\"tool_name\": \"get_kubernetes_nodes\", \"tool_input\": {}}", "done": True}
    
    httpx_mock.add_response(
        url=ollama_api_endpoint,
        method="POST",
        json=mock_response_content,
        status_code=200,
    )

    client = OllamaClient(
        api_url="http://ollama-service.ai-ops-llm.svc.cluster.local:11434",
        model_name="test-model"
    )

    # Act
    result = client.predict(
        system_prompt="You are a helpful assistant.",
        user_query="How many nodes are there?"
    )

    # Assert
    assert result == mock_response_content
    request = httpx_mock.get_request()
    assert request is not None
    request_data = request.json()
    assert request_data["model"] == "test-model"

def test_ollama_client_predict_api_error(httpx_mock):
    # Arrange
    ollama_api_endpoint = "http://ollama-service.ai-ops-llm.svc.cluster.local:11434/api/generate"
    httpx_mock.add_response(
        url=ollama_api_endpoint,
        method="POST",
        status_code=500,
    )

    client = OllamaClient(api_url="http://ollama-service.ai-ops-llm.svc.cluster.local:11434", model_name="test-model")

    # Act
    result_str = client.predict(system_prompt="Doesn't matter", user_query="Doesn't matter")
    result_json = json.loads(result_str)
    
    # Assert
    assert "error" in result_json
    assert "HTTP error connecting to LLM" in result_json["error"]

def test_ollama_client_predict_network_error(httpx_mock):
    # Arrange
    ollama_api_endpoint = "http://ollama-service.ai-ops-llm.svc.cluster.local:11434/api/generate"
    httpx_mock.add_exception(httpx.ConnectError("Connection refused"))

    client = OllamaClient(api_url="http://ollama-service.ai-ops-llm.svc.cluster.local:11434", model_name="test-model")

    # Act
    result_str = client.predict(system_prompt="Doesn't matter", user_query="Doesn't matter")
    result_json = json.loads(result_str)

    # Assert
    assert "error" in result_json
    assert "An unexpected error occurred" in result_json["error"] 