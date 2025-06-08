import unittest
from unittest.mock import MagicMock, patch

from app.agents.orchestrator import OrchestratorAgent
from app.agents.tools import GetKubernetesNodesTool, ScaleDeploymentTool

class TestOrchestratorAgent(unittest.TestCase):

    def test_run_selects_and_executes_correct_tool(self):
        # Arrange
        # 1. Mock the LLM client
        mock_llm_client = MagicMock()
        
        # 2. Program the mock LLM to return a specific JSON response when called.
        # This simulates the LLM choosing the 'get_kubernetes_nodes' tool.
        fake_llm_json_response = """
        {
            "tool_name": "get_kubernetes_nodes",
            "tool_input": {}
        }
        """
        mock_llm_client.predict.return_value = fake_llm_json_response

        # 3. Create the agent with the mocked LLM client
        agent = OrchestratorAgent(llm_client=mock_llm_client)
        
        # 4. Patch the actual tool's 'use' method so we can verify it was called.
        # We don't want to test the tool's logic here, just that the orchestrator calls it.
        with patch.object(GetKubernetesNodesTool, 'use', return_value="Mocked tool result") as mock_tool_use:
            
            # Act
            user_query = "How many nodes are in my cluster?"
            workspace_id = "ws-12345"
            result = agent.run(user_query=user_query, workspace_id=workspace_id)

            # Assert
            # 1. Assert that the LLM was called.
            mock_llm_client.predict.assert_called_once()
            
            # 2. Assert that the correct tool's 'use' method was called.
            mock_tool_use.assert_called_once()
            
            # 3. Assert that the final result is what the mocked tool returned.
            self.assertEqual(result, "Mocked tool result")

            # 4. (Optional) Check the input passed to the tool
            call_args, _ = mock_tool_use.call_args
            tool_input_arg = call_args[0]
            self.assertEqual(tool_input_arg.workspace_id, "ws-12345")

    def test_run_handles_malformed_llm_response(self):
        # Arrange
        # 1. Mock the LLM client to return a non-JSON string.
        mock_llm_client = MagicMock()
        mock_llm_client.predict.return_value = "This is not JSON."

        # 2. Create the agent with the mocked client.
        agent = OrchestratorAgent(llm_client=mock_llm_client)

        # Act
        user_query = "Some query"
        workspace_id = "ws-12345"
        result = agent.run(user_query=user_query, workspace_id=workspace_id)

        # Assert
        # 3. Assert that the agent returns a user-friendly error.
        self.assertIn("The AI model failed to produce a valid tool selection.", result)

    def test_run_handles_invalid_tool_name_from_llm(self):
        # Arrange
        # 1. Mock the LLM client to return a valid JSON but with a tool that doesn't exist.
        mock_llm_client = MagicMock()
        fake_llm_json_response = """
        {
            "tool_name": "make_coffee_tool",
            "tool_input": {}
        }
        """
        mock_llm_client.predict.return_value = fake_llm_json_response

        # 2. Create the agent.
        agent = OrchestratorAgent(llm_client=mock_llm_client)

        # Act
        result = agent.run(user_query="make me coffee", workspace_id="ws-12345")

        # Assert
        # 3. Assert that the agent returns an error about an invalid tool.
        self.assertIn("The AI selected an invalid tool ('make_coffee_tool')", result)

    def test_run_selects_and_executes_scale_tool(self):
        # Arrange
        mock_llm_client = MagicMock()
        
        # Simulate the LLM choosing the 'scale_deployment' tool with specific parameters.
        fake_llm_json_response = """
        {
            "tool_name": "scale_deployment",
            "tool_input": {
                "deployment_name": "my-web-api",
                "replicas": 3
            }
        }
        """
        mock_llm_client.predict.return_value = fake_llm_json_response
        agent = OrchestratorAgent(llm_client=mock_llm_client)
        
        # Patch the tool's 'use' method to verify it was called correctly.
        with patch.object(ScaleDeploymentTool, 'use', return_value="Mocked scale result") as mock_tool_use:
            
            # Act
            user_query = "scale my-web-api to 3 pods"
            workspace_id = "ws-12345"
            result = agent.run(user_query=user_query, workspace_id=workspace_id)

            # Assert
            mock_llm_client.predict.assert_called_once()
            mock_tool_use.assert_called_once()
            self.assertEqual(result, "Mocked scale result")

            # Check the input passed to the tool
            call_args, _ = mock_tool_use.call_args
            tool_input_arg = call_args[0]
            self.assertEqual(tool_input_arg.workspace_id, "ws-12345")
            self.assertEqual(tool_input_arg.deployment_name, "my-web-api")
            self.assertEqual(tool_input_arg.replicas, 3)

if __name__ == '__main__':
    unittest.main() 