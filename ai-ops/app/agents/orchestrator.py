import json
from .tools import TOOL_REGISTRY, GetKubernetesNodesInput, ScaleDeploymentInput, LogQueryInput

class OrchestratorAgent:
    def __init__(self, llm_client):
        self.llm_client = llm_client
        self.tools = {tool.name: tool for tool in TOOL_REGISTRY}

    def _generate_system_prompt(self) -> str:
        """Dynamically generates the system prompt based on available tools."""
        prompt = "You are the Hexabase AIOps Orchestrator. Your role is to understand a user's request and use the available tools to answer it. "\
                 "You must respond in a specific JSON format. Based on the user's query, select exactly one tool to use.\n\n" \
                 "Available Tools:\n"
        
        for tool in self.tools.values():
            prompt += f"- {tool.name}: {tool.description}\n"

        prompt += "\n" \
                  "Your response MUST be a single JSON object with two keys: 'tool_name' and 'tool_input'. "\
                  "'tool_name' must be a string matching one of the available tool names. "\
                  "'tool_input' must be a JSON object containing the parameters for that tool."

        return prompt

    def select_tool(self, user_query: str, workspace_id: str) -> tuple[str, dict]:
        """
        Uses an LLM to select the appropriate tool and generate its input based on the user's query.
        """
        system_prompt = self._generate_system_prompt()
        
        # Get response from LLM client
        llm_response = self.llm_client.predict(
            system_prompt=system_prompt,
            user_query=user_query
        )
        
        try:
            # Handle both dict responses (from real API) and string responses (from mocks)
            if isinstance(llm_response, dict):
                # Check for error in response
                if "error" in llm_response:
                    print(f"LLM error: {llm_response.get('error')}")
                    return "error", {
                        "message": "The AI model encountered an error",
                        "detail": llm_response.get('error')
                    }
                
                # Extract the actual response text if it's nested
                if "response" in llm_response:
                    response_text = llm_response["response"]
                    response_data = json.loads(response_text)
                else:
                    response_data = llm_response
            else:
                # Handle string responses (for backward compatibility)
                response_data = json.loads(llm_response)
            
            tool_name = response_data["tool_name"]
            tool_input = response_data["tool_input"]
            
            # Validate tool exists
            if tool_name not in self.tools:
                print(f"Invalid tool selected: {tool_name}")
                return "error", {
                    "message": f"The AI selected an invalid tool: '{tool_name}'",
                    "available_tools": list(self.tools.keys())
                }
            
            # Add workspace_id to every tool input for context
            tool_input["workspace_id"] = workspace_id
            
            return tool_name, tool_input
            
        except json.JSONDecodeError as e:
            print(f"Error decoding JSON response: {e}")
            return "error", {
                "message": "The AI model produced invalid JSON",
                "detail": str(e),
                "raw_response": str(llm_response)
            }
        except KeyError as e:
            print(f"Missing required field in response: {e}")
            return "error", {
                "message": f"The AI model response is missing required field: {e}",
                "detail": str(e)
            }
        except Exception as e:
            print(f"Unexpected error parsing LLM response: {e}")
            return "error", {
                "message": "An unexpected error occurred while processing the AI response",
                "detail": str(e),
                "type": type(e).__name__
            }

    def run(self, user_query: str, workspace_id: str) -> str:
        """
        Main execution loop for the orchestrator.
        1. Selects a tool.
        2. Executes the tool.
        3. Returns the result.
        """
        tool_name, tool_input_dict = self.select_tool(user_query, workspace_id)

        if tool_name not in self.tools:
            return f"Error: The AI selected an invalid tool ('{tool_name}')."

        tool = self.tools[tool_name]

        # Map tool names to their input model classes
        tool_input_models = {
            "get_kubernetes_nodes": GetKubernetesNodesInput,
            "scale_deployment": ScaleDeploymentInput,
            "query_logs": LogQueryInput
        }
        
        # Get the appropriate input model class
        input_model_class = tool_input_models.get(tool.name)
        if not input_model_class:
            return f"Error: Could not find input model for tool '{tool.name}'. Available tools: {list(tool_input_models.keys())}"
        
        try:
            # Create the input model instance
            input_model = input_model_class(**tool_input_dict)
        except Exception as e:
            return f"Error: Invalid input for tool '{tool.name}': {str(e)}"

        result = tool.use(input_model)
        return result 