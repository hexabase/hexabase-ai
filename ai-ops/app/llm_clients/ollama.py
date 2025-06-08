import httpx
import json

class OllamaClient:
    """
    A client for interacting with an Ollama API server.
    """
    def __init__(self, api_url: str, model_name: str):
        self.api_url = api_url.rstrip('/')
        self.model_name = model_name
        self.http_client = httpx.Client(timeout=60.0)

    def predict(self, system_prompt: str, user_query: str) -> dict:
        """
        Sends a request to the Ollama /api/generate endpoint (non-streaming).

        Returns:
            A dictionary containing either the response or an error.
        """
        endpoint = f"{self.api_url}/api/generate"
        
        full_prompt = f"User Query: {user_query}"

        payload = {
            "model": self.model_name,
            "system": system_prompt,
            "prompt": full_prompt,
            "format": "json", # Instruct Ollama to output valid JSON
            "stream": False # We want a single response
        }
        
        try:
            response = self.http_client.post(endpoint, json=payload)
            response.raise_for_status()
            
            # The response from Ollama is a JSON object.
            # The actual LLM-generated JSON string is inside the 'response' key.
            # We return the whole body for the test to assert against.
            # A more robust client might parse this and just return the inner JSON.
            return response.json()

        except httpx.HTTPStatusError as e:
            print(f"HTTP error occurred: {e}")
            # Return consistent dict format with error details
            return {
                "error": f"HTTP error connecting to LLM: {e.response.status_code}",
                "status_code": e.response.status_code,
                "detail": str(e)
            }
        except httpx.RequestError as e:
            print(f"Request error occurred: {e}")
            return {
                "error": "Failed to connect to LLM service",
                "detail": str(e)
            }
        except Exception as e:
            print(f"An unexpected error occurred: {e}")
            return {
                "error": "An unexpected error occurred",
                "detail": str(e)
            }