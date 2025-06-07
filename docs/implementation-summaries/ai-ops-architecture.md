# AIOps Architecture and Implementation Summary

This document provides a detailed overview of the Hexabase KaaS AIOps system, covering its architecture, technology stack, integration with the main control plane, and security model.

## 1. Architectural Vision

The AIOps system is designed as a distinct, Python-based subsystem that operates alongside the Go-based Hexabase Control Plane. This separation allows for leveraging the rich Python AI/ML ecosystem while maintaining a clear boundary between the core KaaS operations and the AI-driven intelligence layer.

**Core Principles:**

- **Decoupled Systems**: The AIOps system is a separate deployment, communicating with the Control Plane via internal, secured APIs.
- **Optimized Tech Stack**: Utilizes Python, FastAPI, LangChain/LlamaIndex, and Ollama for rapid development and access to state-of-the-art AI tooling.
- **Hierarchical Agents**: Employs a multi-layered agent architecture, from a central orchestrator to specialized agents, to efficiently manage tasks and analysis.
- **Secure by Design**: Inherits user permissions via a short-lived JWT model, with all actions ultimately authorized and executed by the Control Plane.

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

- **HKS Control Plane (Go)**: The existing main application.
- **AIOps System (Python)**: A new deployment in a separate `ai-ops` namespace. It consists of:
  - **API Server**: A FastAPI application that serves as the entry point for the HKS Control Plane.
  - **Orchestrators & Agents**: Implemented in Python using frameworks like LlamaIndex or LangChain.
- **Private LLM Server (Ollama)**: Deployed as a `DaemonSet` onto dedicated nodes (labeled `node-role: private-llm`) in an `ai-ops-llm` namespace. This ensures LLM workloads are isolated.

## 3. Private LLM Setup with Ollama

We will use Ollama to simplify the deployment and management of open-source LLMs (e.g., Llama 3, Phi-3).

**Setup Steps:**

1.  **Provision Node(s)**: Create one or more Kubernetes nodes with GPUs (recommended) or powerful CPUs. Apply the label `node-role: private-llm` to these nodes.
2.  **Deploy Ollama**: Create a Kubernetes manifest for an `Ollama` `DaemonSet` that includes a `nodeSelector` for `node-role: private-llm`. This ensures Ollama runs on all designated nodes.
3.  **Expose Service**: Create a `Service` named `ollama-service` that targets the Ollama pods, providing a stable internal endpoint for the AIOps system.
4.  **Pre-pull Models**: Use an `initContainer` in the AIOps deployment or a one-off Kubernetes `Job` to pre-pull the required models into Ollama (e.g., `ollama pull llama3:8b`). This avoids cold-start delays.
5.  **Integration**: The Python AIOps code will use the LangChain/LlamaIndex Ollama integration, pointing to `http://ollama-service.ai-ops-llm.svc.cluster.local` to perform inference.

## 4. Security Model: AIOps Sandbox

The security model is critical and is based on user impersonation via short-lived, scoped tokens.

**Flow:**

1.  A user sends a chat message to the HKS Control Plane.
2.  The Control Plane authenticates the user and generates a short-lived **Internal JWT**. This JWT contains the user's ID, their roles, and the scope of their request (e.g., `workspace_id`).
3.  The Control Plane calls the AIOps API, passing the user's request and the Internal JWT in an `Authorization` header.
4.  The AIOps system receives the request. Its agents can use the information in the JWT for context (e.g., to know which workspace to query), but they cannot perform actions directly.
5.  To execute an action (e.g., scale a deployment), the AIOps agent makes a call back to a specific, non-public **Internal Operations API** on the HKS Control Plane (e.g., `POST /internal/v1/operations/scale`).
6.  This request **must** include the original Internal JWT.
7.  The HKS Control Plane receives the request. It **re-validates** the JWT and performs a **final authorization check**: "Does this user (`sub` from JWT) have permission to perform this action on this resource?"
8.  If authorized, the Control Plane executes the operation using its own privileged service account. If not, it returns a permission error.

This flow ensures that the AIOps system is fully sandboxed. It acts as an intelligent "advisor" that can request actions, but the Control Plane remains the sole, authoritative "executor," enforcing all security and RBAC policies.

## 5. Development and Repository Structure

Initially, the AIOps system will be developed in a subdirectory of the main repository to facilitate close integration.

```
/hexabase-ai
├── api/           # Go HKS Control Plane
├── ui/            # Next.js UI
└── ai-ops/        # Python AIOps System
    ├── app/
    │   ├── agents/
    │   ├── main.py
    ├── Dockerfile
    └── requirements.txt
```

This structure supports the potential for future separation into a dedicated repository by simply moving the `ai-ops` directory, as the communication is already designed to be via internal APIs.
