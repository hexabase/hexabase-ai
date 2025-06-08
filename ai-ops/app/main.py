import httpx
import os
from fastapi import FastAPI, Request, HTTPException, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel
import jwt
from typing import List

from .agents.orchestrator import OrchestratorAgent
from .llm_clients.ollama import OllamaClient

# --- Configuration ---
HKS_JWKS_URL = os.getenv("HKS_JWKS_URL", "http://api-service.hexabase.svc.cluster.local/.well-known/jwks.json")
OLLAMA_API_URL = os.getenv("OLLAMA_API_URL", "http://ollama-service.ai-ops-llm.svc.cluster.local:11434")
OLLAMA_MODEL_NAME = os.getenv("OLLAMA_MODEL_NAME", "llama3:8b") # Default to Llama 3 8B

# --- Globals ---
app = FastAPI(
    title="Hexabase AIOps Service",
    description="The AI/ML backend for Hexabase KaaS.",
    version="0.1.0",
)
security = HTTPBearer()
jwks_cache = {}

# Instantiate our agent with the REAL Ollama LLM client
llm_client = OllamaClient(api_url=OLLAMA_API_URL, model_name=OLLAMA_MODEL_NAME)
orchestrator_agent = OrchestratorAgent(llm_client=llm_client)

# --- Pydantic Models ---
class HealthResponse(BaseModel):
    status: str = "ok"

class ChatRequest(BaseModel):
    message: str

class ChatResponse(BaseModel):
    reply: str
    user_id: str
    workspace_id: str

class InternalTokenClaims(BaseModel):
    user_id: str
    org_ids: List[str]
    active_workspace_id: str
    token_type: str

# --- Helper Functions ---
async def get_public_key():
    """
    Fetches the public key from the Hexabase Control Plane's JWKS endpoint.
    Caches the key for performance.
    """
    if "public_key" in jwks_cache:
        return jwks_cache["public_key"]

    try:
        async with httpx.AsyncClient() as client:
            response = await client.get(HKS_JWKS_URL)
            response.raise_for_status()
            jwks = response.json()
            # Find the RSA key and construct a PEM public key
            key_data = jwks["keys"][0]
            public_key = jwt.algorithms.RSAAlgorithm.from_jwk(key_data)
            jwks_cache["public_key"] = public_key
            return public_key
    except Exception as e:
        # On startup or if the Go service is down, this will fail.
        # We should log this error properly.
        print(f"Error fetching JWKS: {e}")
        raise HTTPException(status_code=503, detail="Could not fetch validation keys. AIOps service is not ready.")

async def validate_internal_token(credentials: HTTPAuthorizationCredentials = Depends(security)) -> InternalTokenClaims:
    """
    A dependency that validates the internal JWT and returns its claims.
    """
    token = credentials.credentials
    try:
        public_key = await get_public_key()
        payload = jwt.decode(
            token,
            public_key,
            algorithms=["RS256"],
            audience="hexabase-aiops-service",
            issuer="hexabase-control-plane",
        )
        # Validate the specific token type
        if payload.get("token_type") != "internal-aiops-v1":
            raise HTTPException(status_code=403, detail="Invalid token type")

        return InternalTokenClaims(**payload)
    except jwt.ExpiredSignatureError:
        raise HTTPException(status_code=403, detail="Token has expired")
    except jwt.InvalidTokenError as e:
        print(f"Invalid token error: {e}")
        raise HTTPException(status_code=403, detail="Invalid token")
    except Exception as e:
        print(f"An unexpected error occurred during token validation: {e}")
        raise HTTPException(status_code=500, detail="Internal server error during token validation")


# --- API Endpoints ---
@app.get("/health", response_model=HealthResponse, tags=["Health"])
async def health_check():
    """
    Endpoint to verify that the service is running.
    """
    return HealthResponse()

@app.post("/v1/chat", response_model=ChatResponse, tags=["AIOps"])
async def chat_endpoint(
    request: ChatRequest,
    claims: InternalTokenClaims = Depends(validate_internal_token)
):
    """
    Main chat endpoint. It now uses the OrchestratorAgent to process the request.
    """
    # The orchestrator's predict method needs to be called to get the raw JSON string
    llm_json_response = llm_client.predict(
        system_prompt=orchestrator_agent._generate_system_prompt(),
        user_query=request.message
    )

    # In the real orchestrator, we would parse this response and execute a tool.
    # For now, let's just show that we got a response from the LLM.
    # This part needs to be updated to use the full `agent.run()` logic.
    # The `run` logic itself needs to be updated to handle the raw JSON object from the client.
    
    agent_reply = orchestrator_agent.run(
        user_query=request.message,
        workspace_id=claims.active_workspace_id
    )
    
    return ChatResponse(
        reply=agent_reply,
        user_id=claims.user_id,
        workspace_id=claims.active_workspace_id,
    )

# In a real application, you would import and include routers here
# from .api import chat, agents
#
# app.include_router(chat.router, prefix="/v1")
# app.include_router(agents.router, prefix="/v1/agents") 