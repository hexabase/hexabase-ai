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

    def select_tool(self, user_query: str, workspace_id: str) -> (str, dict):
        """
        Uses an LLM to select the appropriate tool and generate its input based on the user's query.
        """
        system_prompt = self._generate_system_prompt()
        
        # In a real scenario, the llm_client would make an API call to an LLM.
        # Here, we simulate the LLM's response for testing.
        llm_response_json = self.llm_client.predict(
            system_prompt=system_prompt,
            user_query=user_query
        )
        
        try:
            response_data = json.loads(llm_response_json)
            tool_name = response_data["tool_name"]
            tool_input = response_data["tool_input"]
            
            # Add workspace_id to every tool input for context
            tool_input["workspace_id"] = workspace_id
            
            return tool_name, tool_input
        except (json.JSONDecodeError, KeyError) as e:
            print(f"Error parsing LLM response: {e}")
            return "error", {"message": "The AI model failed to produce a valid tool selection."}

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

        # Find the correct Pydantic model for the tool's input
        if tool.name == "get_kubernetes_nodes":
            input_model = GetKubernetesNodesInput(**tool_input_dict)
        elif tool.name == "scale_deployment":
            input_model = ScaleDeploymentInput(**tool_input_dict)
        elif tool.name == "query_logs":
            input_model = LogQueryInput(**tool_input_dict)
        else:
            return f"Error: Could not find input model for tool '{tool.name}'"

        result = tool.use(input_model)
        return result 