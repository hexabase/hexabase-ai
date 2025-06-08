import httpx
from abc import ABC, abstractmethod
from pydantic import BaseModel
from app.config import settings

class ToolInput(BaseModel):
    """Base model for tool inputs."""
    pass

class Tool(ABC):
    """Interface for a tool that can be called by an agent."""
    
    name: str
    description: str

    @abstractmethod
    def use(self, input_data: ToolInput) -> str:
        """Executes the tool and returns a string result."""
        pass

class GetKubernetesNodesInput(ToolInput):
    """Input model for the GetKubernetesNodes tool."""
    workspace_id: str

class GetKubernetesNodesTool(Tool):
    """A tool to get information about Kubernetes nodes."""
    name = "get_kubernetes_nodes"
    description = "Fetches a list of Kubernetes nodes and their status for a given workspace. Use this to answer questions about cluster nodes, their health, or count."

    def use(self, input_data: GetKubernetesNodesInput) -> str:
        """
        Calls the internal HKS API to get node info for the workspace and formats it.
        """
        endpoint = f"{settings.HKS_INTERNAL_API_URL}/internal/v1/workspaces/{input_data.workspace_id}/nodes"
        
        try:
            with httpx.Client() as client:
                response = client.get(endpoint)
                response.raise_for_status()
                nodes = response.json()
            
            if not nodes:
                return f"No nodes found for workspace {input_data.workspace_id}."

            summary = f"Found {len(nodes)} nodes for workspace {input_data.workspace_id}.\n"
            for node in nodes:
                summary += f"- Node '{node.get('name')}' is {node.get('status')}.\n"
            
            return summary.strip()

        except httpx.HTTPStatusError as e:
            return f"Error: Could not retrieve node data. The API returned a {e.response.status_code} status."
        except Exception as e:
            return f"Error: An unexpected error occurred while fetching node data: {str(e)}"

class ScaleDeploymentInput(ToolInput):
    """Input model for the ScaleDeploymentTool."""
    workspace_id: str
    deployment_name: str
    replicas: int

class ScaleDeploymentTool(Tool):
    """A tool to scale a Kubernetes deployment to a specific number of replicas."""
    name = "scale_deployment"
    description = "Scales a specified deployment in a given workspace to the desired number of replicas. Use this for requests like 'scale my-app to 3 pods' or 'set replicas for web-api to 5'."

    def use(self, input_data: ScaleDeploymentInput) -> str:
        """
        Calls the internal HKS API to scale a deployment.
        """
        endpoint = f"{settings.HKS_INTERNAL_API_URL}/internal/v1/workspaces/{input_data.workspace_id}/deployments/{input_data.deployment_name}/scale"
        payload = {"replicas": input_data.replicas}

        try:
            with httpx.Client() as client:
                response = client.post(endpoint, json=payload)
                response.raise_for_status()
            
            return f"Successfully scaled deployment '{input_data.deployment_name}' to {input_data.replicas} replicas."

        except httpx.HTTPStatusError as e:
            error_detail = e.response.json().get("error", "unknown error")
            return f"Error: Could not scale deployment. The API returned a {e.response.status_code} status: {error_detail}"
        except Exception as e:
            return f"Error: An unexpected error occurred while scaling deployment: {str(e)}"

class LogQueryInput(ToolInput):
    """Input model for the LogQueryingTool."""
    workspace_id: str
    search_term: str = None
    level: str = None
    # In a real tool, we would handle time parsing more robustly
    start_time: str = None
    end_time: str = None
    limit: int = 100

class LogQueryingTool(Tool):
    """A tool to query logs for a given workspace. Use this to answer questions about errors, specific pod logs, or events within a time range."""
    name = "query_logs"
    description = "Searches and retrieves logs from a workspace based on search terms, log level, and time range."

    def use(self, input_data: LogQueryInput) -> str:
        """
        Calls the internal HKS API to query logs.
        """
        endpoint = f"{settings.HKS_INTERNAL_API_URL}/internal/v1/logs/query"
        payload = input_data.dict(exclude_none=True)

        try:
            with httpx.Client() as client:
                response = client.post(endpoint, json=payload)
                response.raise_for_status()
                logs = response.json()

            if not logs:
                return f"No logs found for workspace {input_data.workspace_id} with the given criteria."
            
            summary = f"Found {len(logs)} log entries:\n"
            for log in logs:
                summary += f"- [{log.get('timestamp')}] [{log.get('level')}] {log.get('message')}\n"
            
            return summary.strip()

        except httpx.HTTPStatusError as e:
            error_detail = e.response.json().get("error", "unknown error")
            return f"Error: Could not query logs. The API returned a {e.response.status_code} status: {error_detail}"
        except Exception as e:
            return f"Error: An unexpected error occurred while querying logs: {str(e)}"

# A registry of available tools for the orchestrator to use.
TOOL_REGISTRY = [
    GetKubernetesNodesTool(),
    ScaleDeploymentTool(),
    LogQueryingTool(),
] 