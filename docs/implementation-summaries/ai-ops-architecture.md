# AIOps Architecture and Implementation Summary

This document provides a detailed overview of the Hexabase KaaS AIOps system, covering its architecture, technology stack, integration with the main control plane, and security model.

## 1. Architectural Vision

The AIOps system is designed as a distinct, Python-based subsystem that operates alongside the Go-based Hexabase Control Plane. This separation allows for leveraging the rich Python AI/ML ecosystem while maintaining a clear boundary between the core KaaS operations and the AI-driven intelligence layer.

**Core Principles:**

- **Decoupled Systems**: The AIOps system is a separate deployment, communicating with the Control Plane via internal, secured APIs.
- **Optimized Tech Stack**: Utilizes Python, FastAPI, LangChain/LlamaIndex, and Ollama for rapid development and access to state-of-the-art AI tooling.
- **Hierarchical Agents**: Employs a multi-layered agent architecture, from a user-facing chat agent to a central orchestrator and specialized worker agents, to efficiently manage tasks and analysis.
- **Secure by Design**: Inherits user permissions via a short-lived JWT model, with all actions ultimately authorized and executed by the Control Plane.
- **Extensible LLM Support**: Provides a flexible model for using both internally-hosted open-source LLMs and external commercial LLM APIs, with configuration available at both the organization and workspace levels.
- **Observable and Traceable**: All agent interactions are designed to be logged and traced, providing clear visibility into the system's reasoning and actions for debugging and analysis.

## 2. System Components and Deployment

```mermaid
graph TD
    subgraph "Hexabase Namespace"
        direction LR
        HKS_Control_Plane[HKS Control Plane (Go)]
        HKS_Service[hks-control-plane-svc]
        HKS_Control_Plane -- exposes --> HKS_Service
    end

    subgraph "AIOps Namespace"
        direction LR
        AIOps_System[AIOps API (Python/FastAPI)]
        AIOps_Service[ai-ops-svc]
        AIOps_System -- exposes --> AIOps_Service
    end

    subgraph "AIOps LLM Namespace"
        direction LR
        Ollama_DaemonSet[Ollama DaemonSet]
        Ollama_Service[ollama-svc]
        OLLAMA_NODE[GPU/CPU Node<br/>(label: node-role=private-llm)]
        Ollama_DaemonSet -- runs on --> OLLAMA_NODE
        Ollama_DaemonSet -- exposes --> Ollama_Service
    end

    HKS_Control_Plane -- "Internal API Call w/ JWT" --> AIOps_Service
    AIOps_System -- "Internal Ops API Call w/ JWT" --> HKS_Service
    AIOps_System -- "LLM Inference" --> Ollama_Service
```

### 2.1 Agent-based Architecture

The AIOps system employs a hierarchical, multi-agent architecture to manage user requests and interact with the HKS platform. This separation of concerns allows for specialized agents, better maintainability, and more complex reasoning capabilities.

```mermaid
graph TD
    subgraph "User Interaction Layer"
        User[HKS User]
        ChatClient[Chat Client <br/> (UI, Slack, Teams, etc.)]
        User -- Interacts with --> ChatClient
    end

    subgraph "AIOps System (Python)"
        direction LR
        UserChatAgent[UserChatAgent]
        OrchestrationAgent[Orchestration Agent]
        WorkerAgents[Worker Agents <br/> - Kubernetes Agent <br/> - Prometheus Agent <br/> - ClickHouse Agent <br/> - Storage Agent <br/> - etc.]

        UserChatAgent -- "User Query" --> OrchestrationAgent
        OrchestrationAgent -- "Sub-task" --> WorkerAgents
        WorkerAgents -- "Tool Output" --> OrchestrationAgent
        OrchestrationAgent -- "Synthesized Response" --> UserChatAgent
    end

    subgraph "HKS Control Plane (Go)"
        HKS_InternalAPI[HKS Internal Ops API]
        HKS_DB[HKS Database]
        HKS_InternalAPI -- Accesses --> HKS_DB
    end

    subgraph "LLM Services"
        Ollama[Private LLM (Ollama)]
        LLM_APIs[External LLM APIs <br/> (OpenAI, Anthropic, etc.)]
    end

    ChatClient -- "API Call w/ User Auth" --> HKS_Control_Plane
    HKS_Control_Plane -- "Generates JWT, forwards to" --> UserChatAgent
    UserChatAgent -- "Conversational Logic" --> LLM_APIs
    WorkerAgents -- "Executes Actions via" --> HKS_InternalAPI
    WorkerAgents -- "Inference (optional)" --> Ollama
```

- **UserChatAgent**: The primary point of contact for the end-user.

  - Responsible for managing the conversation flow, maintaining session state, and providing a user-friendly experience.
  - To handle nuanced human conversation, this agent is designed to use powerful external, commercial LLMs (e.g., GPT-4, Claude 3).
  - It forwards the user's core intent to the Orchestration Agent.

- **Orchestration Agent**: The central "brain" or router of the AIOps system.

  - It receives a task from the `UserChatAgent`, breaks it down into smaller, executable steps, and dispatches those steps to the appropriate `Worker Agent(s)`.
  - It synthesizes the results from the workers into a coherent final answer for the `UserChatAgent`.

- **Worker Agents**: A collection of specialized, tool-using agents.
  - Each worker is an expert on a specific domain (e.g., interacting with the Kubernetes API, querying Prometheus, analyzing logs in ClickHouse, monitoring storage).
  - They execute concrete tasks using predefined tools and APIs. These agents may use smaller, local LLMs for simple data processing but often do not require an LLM for their core function.
  - All actions that modify the HKS state are performed by making secure calls to the HKS Internal Operations API, never directly.

This structure allows for future integration with various chat clients (Slack, Teams, Discord) and even as a backend for tools like Cursor, as the `UserChatAgent` abstracts the interaction logic.

- **HKS Control Plane (Go)**: The existing main application.
- **AIOps System (Python)**: A new deployment in a separate `ai-ops` namespace. It consists of:
  - **API Server**: A FastAPI application that serves as the entry point for the HKS Control Plane.
  - **Orchestrators & Agents**: Implemented in Python using frameworks like LlamaIndex or LangChain.
- **Private LLM Server (Ollama)**: Deployed as a `DaemonSet` onto dedicated nodes (labeled `node-role: private-llm`) in an `ai-ops-llm` namespace. This ensures LLM workloads are isolated.

## 3. LLM Configuration and Management

The AIOps system supports a flexible approach to LLM usage, accommodating both private, self-hosted models for internal tasks and powerful commercial models for user-facing interactions.

### 3.1 LLM Providers

- **Private LLMs (Ollama)**: We use Ollama to simplify the deployment and management of open-source LLMs (e.g., Llama 3, Phi-3). These are used for internal tasks like routing, data extraction, or simple analysis where data residency is critical. The setup involves:

  1.  **Provisioning Nodes**: Designating Kubernetes nodes with the label `node-role: private-llm`.
  2.  **Deploying Ollama**: Using a `DaemonSet` with a `nodeSelector` for `node-role: private-llm`.
  3.  **Exposing Service**: Creating a `Service` (`ollama-service`) as a stable internal endpoint.
  4.  **Pre-pulling Models**: Using an `initContainer` or a `Job` to pull required models into Ollama.
  5.  **Integration**: The Python code points to `http://ollama-service.ai-ops-llm.svc.cluster.local` for inference.

- **External LLMs**: For the `UserChatAgent`, which requires advanced conversational abilities, the system will integrate with external commercial LLM providers (e.g., OpenAI, Google, Anthropic). API keys and model preferences are managed securely.

### 3.2 Configuration Hierarchy

To provide flexibility, LLM settings can be configured at two levels:

1.  **Organization Level (Default)**: A default LLM configuration (e.g., for the `UserChatAgent`) is set for the entire HKS organization. This configuration is managed via environment variables in the AIOps system's deployment.
2.  **Workspace Level (Override)**: Workspace Admins can override the default LLM settings for their specific workspace. This allows them to choose a different model or provide their own API key. This requires:
    - An API endpoint in the HKS Control Plane to store and retrieve workspace-specific LLM settings.
    - A corresponding UI for Workspace Admins to manage these settings.
    - The AIOps system will first check for a workspace-specific configuration and fall back to the organization-level default if none is found.

## 4. Security Model: AIOps Sandbox and Session Management

The security model is critical and is based on user impersonation via short-lived, scoped tokens. The AIOps system acts as a sandboxed advisor, with the HKS Control Plane as the sole enforcer of permissions.

### 4.1 Authorization Flow

1.  A user initiates a chat session via a client (HKS UI, Slack, etc.).
2.  The HKS Control Plane authenticates the user and generates a short-lived **Internal JWT**. This JWT contains the user's ID, their roles, and the scope of their request (e.g., `workspace_id`).
3.  The Control Plane calls the `UserChatAgent` in the AIOps system, passing the user's request and the Internal JWT.
4.  The AIOps system's agents (`UserChatAgent`, `Orchestrator`, `Workers`) pass this JWT internally for context and subsequent API calls. The agents themselves have no inherent privileges.
5.  To execute an action (e.g., scale a deployment), a `Worker Agent` makes a call back to a specific, non-public **Internal Operations API** on the HKS Control Plane (e.g., `POST /internal/v1/operations/scale`).
6.  This request **must** include the original Internal JWT.
7.  The HKS Control Plane receives the request. It **re-validates** the JWT and performs a **final authorization check**: "Does this user (`sub` from JWT) have permission to perform this action on this resource, according to the _latest_ data in our database?"
8.  If authorized, the Control Plane executes the operation using its own privileged service account. If not, it returns a permission error.

This flow ensures that the AIOps system is fully sandboxed. It can request actions, but the Control Plane remains the sole, authoritative "executor," enforcing all security and RBAC policies at the moment of execution.

### 4.2 Session Management

User sessions with the `UserChatAgent` are stateful but must adapt to changes in user permissions.

- **Session Timeout**: Sessions will have a defined idle timeout, after which the `UserChatAgent` will effectively "log out". The next user interaction will trigger a new authentication flow with the HKS Control Plane.
- **Permission Change Detection**: HKS user permissions can change. To ensure the AIOps system never operates on stale permissions, the session must be re-validated. On each request that requires an action, the final authorization check by the Control Plane (Step 7 above) implicitly handles this. If a user's permissions were revoked, the action will fail. The AIOps system should interpret this as a potential permission change and can prompt the user to re-authenticate to "resume" the session with updated credentials.

Thorough security test cases for this entire flow, especially covering permission changes and token validation, will be a critical part of the development process.

## 5. LLMOps and Observability

To ensure the AIOps system is transparent, debuggable, and continuously improving, its interactions are tracked in two primary ways: AI Tracing for development and Audit Logging for user-visible actions.

- **AI Tracing with Langfuse**: We will integrate the AIOps system with [Langfuse](https://langfuse.com/). This is a developer-focused tool that captures the entire internal reasoning lifecycle of a request:
  - The initial prompt from the `UserChatAgent`.
  - The reasoning and task breakdown from the `Orchestration Agent`.
  - The specific tools called and results returned by the `Worker Agents`.
  - The final synthesized response.
- **Benefits**: This detailed tracing provides invaluable data for debugging complex agent behaviors, analyzing performance, evaluating LLM quality, and creating datasets for future fine-tuning. The Langfuse SDK will be integrated directly into the Python AIOps application.
- **Audit Logging**: All definitive actions taken by an agent on behalf of a user (e.g., modifying a Kubernetes resource) are logged in the central HKS Audit Log system (ClickHouse). This provides a compliant, user-visible record of all changes. For more details, see the main `Logging and Auditing Architecture` document.

## 6. Development and Repository Structure

Initially, the AIOps system will be developed in a subdirectory of the main repository to facilitate close integration.

```

```
