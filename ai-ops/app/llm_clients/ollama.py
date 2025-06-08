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

    def predict(self, system_prompt: str, user_query: str) -> str:
        """
        Sends a request to the Ollama /api/generate endpoint (non-streaming).

        Returns:
            The raw JSON string from the 'response' field of the last JSON object from Ollama.
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
            # In a real app, you might want to raise a custom exception.
            return f'{{"error": "HTTP error connecting to LLM: {e.get("status_code")}"}}'
        except Exception as e:
            print(f"An error occurred: {e}")
            return f'{{"error": "An unexpected error occurred: {str(e)}"}}' 