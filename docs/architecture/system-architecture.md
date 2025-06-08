# Hexabase AI: Architecture Specification

## 1. Project Overview

### 1.1. Vision

This project aims to develop and provide an open-source, observable, and intuitive multi-tenant Kubernetes as a Service (KaaS) platform based on `K3s` and `vCluster`. While Kubernetes has become the de facto standard in modern application development, its high learning curve, complex operational management, and difficulty in securing resources for small teams and individual developers remain barriers to adoption. Hexabase AI is designed to solve these challenges.

Specifically, it provides the following value:

- **Ease of Deployment**: Based on `K3s`, a lightweight Kubernetes distribution, and utilizing `vCluster` virtualization technology, users are freed from the complexity of physical cluster management and can quickly start using isolated Kubernetes environments. This ease of use encourages developers and teams with limited Kubernetes expertise to embrace new technologies. Traditional Kubernetes cluster construction required extensive expertise in network configuration, security policy formulation, storage provisioning, and more. Hexabase AI automates and abstracts much of this, providing a ready-to-use state with just a few clicks.

- **Intuitive Operation**: Kubernetes' powerful features are abstracted through a sophisticated UI/UX, making them easily accessible to users without specialized knowledge. Resources can be managed using intuitive concepts such as Organization, Workspace, and Project. For example, common operations like application deployment, scaling, and monitoring can be performed from a graphical interface without directly editing YAML files. Error messages and logs are also displayed in an understandable way to support problem resolution.

- **Strong Tenant Isolation**: `vCluster` provides each tenant (Workspace) with dedicated API servers and control plane components, ensuring significantly higher security and independence than namespace-based isolation. This minimizes cross-tenant impact and allows resources to be used with confidence. For example, even if one tenant accidentally consumes excessive resources or causes security issues, other tenant environments are designed to remain unaffected. This is particularly crucial when hosting multiple customers or projects on the same physical infrastructure.

- **Cloud-Native Operations**: Equipped with a comprehensive monitoring stack using `Prometheus`, `Grafana`, and `Loki`, enabling real-time visibility into system health and tenant resource usage. Adopting a GitOps approach with `Flux` enables declarative configuration management and reproducible deployments. Policy management with `Kyverno` supports enhanced security compliance and governance. This ensures that infrastructure configuration changes, application deployments, and security policy applications are performed through version-controlled code, facilitating audit trails and rollbacks.

- **Open Source Transparency and Community**: By releasing this project as open source, we ensure technical transparency and actively welcome feedback and contributions from developers worldwide. We aim to build a reliable platform that grows with the community and can address more use cases. We also anticipate use in educational institutions and as a learning/validation platform for new cloud-native technologies. Open source code publication leads to early discovery and correction of security vulnerabilities. Additionally, incorporating diverse perspectives enables more innovative and practical feature development.

Hexabase AI envisions being a catalyst for delivering the power of Kubernetes to more people and accelerating innovation. Developers will be freed from infrastructure complexity and able to focus on the essential value creation of application development.

## 2. System Architecture

The Hexabase AI system architecture consists of the **Hexabase UI (Next.js)** that users directly interact with, the **Hexabase API (control plane, Go language)** that manages and controls the entire system, various supporting **middleware (PostgreSQL, Redis, NATS, etc.)**, and a new **AIOps System (Python)**. All these components are containerized and run on the operational foundation **Host K3s Cluster**. Per-tenant Kubernetes environments are virtually constructed within the Host K3s Cluster using **vCluster** technology, providing strong isolation and independence. This multi-layered architecture is designed with scalability, availability, maintainability, and intelligence in mind.

**Architectural Diagram:**

```mermaid
graph TD
    subgraph "User Interaction"
        direction LR
        User -- "Browser, Slack, etc." --> Frontend
        Frontend[Hexabase UI / Chat Client]
    end

    subgraph "Hexabase Control Plane (Go)"
        direction LR
        Frontend -- "REST/WebSocket" --> API_Server[API Server]
        API_Server -- "Publish Tasks" --> NATS[NATS Messaging]
        NATS -- "Consume Tasks" --> Workers[Async Workers]
    end

    subgraph "AIOps System (Python)"
        direction LR
        AIOps_API[AIOps API]
        AIOps_Orchestrator[Orchestrator]
        AIOps_Agents[Specialized Agents]
        Private_LLM[Private LLMs on Ollama]

        AIOps_API --> AIOps_Orchestrator
        AIOps_Orchestrator --> AIOps_Agents
        AIOps_Agents --> Private_LLM
        AIOps_Orchestrator -- "External LLM API" --> Internet
    end

    subgraph "Data & State"
        PostgreSQL
        Redis
        Central_Logging[Central Logging (ClickHouse)]
        LLMOps[LLMOps (Langfuse)]
    end

    subgraph "Host K3s Cluster"
        vClusters[vClusters per Tenant]
        Shared_Observability[Shared Observability Stack]
    end

    API_Server -- "Read/Write" --> PostgreSQL
    API_Server -- "Cache" --> Redis
    API_Server -- "Log" --> Central_Logging
    Workers -- "Read/Write" --> PostgreSQL

    API_Server -- "Internal JWT" --> AIOps_API
    AIOps_API -- "Log/Trace" --> LLMOps

    API_Server -- "Manage" --> vClusters
    Workers -- "Manage" --> vClusters

    style Frontend fill:#d4f0ff
    style AIOps_API fill:#e6ffc2
```

**Key Component Interactions and Data Flow:**

1. **User Operations and UI**: Users access the Hexabase UI through a web browser to perform operations such as creating Organizations, provisioning Workspaces (vClusters), managing Projects (Namespaces), inviting users, and setting permissions. The UI converts these operations into requests to the Hexabase API. The UI manages the user's authentication state and attaches authentication tokens to API requests. Real-time information updates (e.g., vCluster provisioning progress) will be implemented using technologies such as WebSocket or Server-Sent Events.

2. **API Request Processing**: The Hexabase API receives requests from the UI and first performs authentication and authorization processing. After confirming that the authenticated user has permission to perform the requested operation, it executes the business logic. This includes updating the PostgreSQL database state and issuing instructions to the vCluster orchestrator. Time-consuming processes (e.g., vCluster creation, large-scale configuration changes) are registered as tasks in the NATS message queue and delegated to asynchronous workers to maintain API server responsiveness. The API also strictly validates requests and returns appropriate error responses for invalid input.

3. **vCluster Orchestration**: The vCluster orchestrator interacts with the Host K3s cluster to manage the vCluster lifecycle (creation, configuration, deletion). Specifically, it uses `vcluster CLI` or Kubernetes API (`client-go`) to deploy vCluster Pods (typically as StatefulSets or Deployments), configure necessary network settings (Service, Ingress, etc.), and storage settings (PersistentVolumeClaim). It also handles applying OIDC settings to each vCluster, installing and configuring HNC (Hierarchical Namespace Controller), setting resource quotas according to tenant plans, and controlling Dedicated Node allocation (using Node Selectors and Taints/Tolerations). Additionally, this component executes configuration of Namespaces and RBAC (Role, RoleBinding, ClusterRole, ClusterRoleBinding) within vClusters based on user operations.

4. **Asynchronous Processing**: Asynchronous workers receive tasks from the NATS message queue and execute background processing such as vCluster provisioning, Stripe API integration (billing processing), HNC setup, and backup/restore processing (future feature). This allows the API server to return responses quickly without being blocked for long periods. Workers record processing progress in the database and will notify the API server or notification system of results through NATS upon completion or error.

5. **State Persistence**: The PostgreSQL database stores Organizations, Workspaces, Projects, Users, Groups, Roles, billing plans, subscription information, asynchronous task status, audit logs, and more. Transactions are used appropriately to maintain data consistency, and regular backup and restore strategies are planned. Schema changes are managed using migration tools (e.g., golang-migrate).

6. **Caching**: Redis caches user session information, frequently accessed configuration data, public keys (JWKS) required for OIDC token validation, rate limit counters, etc., reducing database load and improving system responsiveness and scalability. Cache expiration and invalidation strategies are also properly designed.

7. **Monitoring and Logging**:
   The monitoring architecture employs a hybrid model based on the tenant's plan.

   - **Shared Plan**: Each vCluster runs lightweight agents (`prometheus-agent`, `promtail`) that forward metrics and logs to a central, multi-tenant **Prometheus and Loki stack** on the host cluster. Tenant data is isolated using labels (`workspace_id`).
   - **Dedicated Plan**: A dedicated, fully independent observability stack (Prometheus, Grafana, Loki) can be deployed inside the tenant's vCluster for complete isolation.
   - **Central Logging**: All Hexabase control plane and AIOps system logs are aggregated into a central **ClickHouse** database for high-speed querying and analysis.

8. **GitOps Deployment**: Deployment and updates of the Hexabase control plane itself are managed through GitOps workflows using Flux. Infrastructure configuration (Kubernetes manifests, Helm Charts), application settings, security policies, etc., are all declaratively managed in Git repositories. Changes are made through Git commits and pull requests, and once approved, Flux automatically applies them to the Host K3s cluster. This improves deployment reproducibility, auditability, and reliability.

9. **Policy Application**: Kyverno operates as a Kubernetes Admission Controller, enforcing security and operational policies on the Host K3s cluster and within each vCluster (if configurable). For example, policies such as "all Namespaces must have an `owner` label," "prohibit launching privileged containers," or "block image pulls from untrusted registries" can be defined to maintain compliance. Policies should also be managed through GitOps.

10. **Serverless Backbone**: **Knative** is installed on the host cluster to provide the underlying infrastructure for the HKS Functions (FaaS) offering. It manages the entire lifecycle of serverless containers, including scaling to zero.

11. **AIOps System Interaction**: The AIOps system operates as a separate Python-based service. The Hexabase API server communicates with it via internal, RESTful APIs, passing a short-lived, scoped JWT for secure, context-aware operations. The AIOps system analyzes data from the observability stack and its own agents, and can request operational changes (e.g., scaling a deployment) by calling back to a secured internal API on the Hexabase control plane, which performs the final authorization and execution.

This architecture aims to realize a scalable, resilient, intelligent, and operationally friendly KaaS platform. By clarifying the division of responsibilities among components and utilizing standardized technologies and open-source products, we enhance development efficiency and system reliability.

## 3. Core Concepts and Entity Mapping

Hexabase AI provides unique abstracted concepts to allow users to use the service without being aware of Kubernetes complexity. These concepts are internally mapped to standard Kubernetes resources and features. Understanding this mapping is crucial for grasping system behavior and using it effectively.

| Hexabase Concept      | Kubernetes Equivalent                              | Scope              | Notes                                                                                |
| --------------------- | -------------------------------------------------- | ------------------ | ------------------------------------------------------------------------------------ |
| Organization          | (None)                                             | Hexabase           | Unit for billing, invoicing, and organizational user management. Business logic.     |
| Workspace             | vCluster                                           | Host K3s Cluster   | Strong tenant isolation boundary.                                                    |
| Workspace Plan        | ResourceQuota / Node Configuration                 | vCluster / Host    | Defines resource limits.                                                             |
| Organization User     | (None)                                             | Hexabase           | Organization administrators and billing managers.                                    |
| Workspace Member      | User (OIDC Subject)                                | vCluster           | Technical personnel operating vCluster. Authenticated via OIDC.                      |
| Workspace Group       | Group (OIDC Claim)                                 | vCluster           | Unit for permission assignment. Hierarchy resolved by Hexabase.                      |
| Workspace ClusterRole | ClusterRole                                        | vCluster           | Preset permissions spanning entire Workspace (e.g., Admin, Viewer).                  |
| Project               | Namespace                                          | vCluster           | Resource isolation unit within Workspace.                                            |
| Project Role          | Role                                               | vCluster Namespace | Custom permissions that users can create within a Project.                           |
| **CronJob**           | `batch/v1.CronJob`                                 | vCluster Namespace | A scheduled task, configured via the UI but maps to a native CronJob resource.       |
| **Function**          | Knative Service (`serving.knative.dev/v1.Service`) | vCluster Namespace | A serverless function deployed via the `hks-func` CLI or dynamically by an AI agent. |

# 4. Functional Specifications and User Flows

## 4.1. Signup and Organization Management

- **New User Registration**  
  Sign up via OpenID Connect with external IdPs (Google, GitHub, etc.). User is created in Hexabase DB.

- **Organization Creation**  
  Upon initial signup, a private Organization is automatically created for the user. The user becomes the first Organization User of this Org.

- **Organization Management**  
  Organization Users can manage billing information (Stripe integration) and invite other Organization Users.  
  \*Note: This permission does not allow direct manipulation of resources within subordinate Workspaces (vClusters).

## 4.2. Workspace (vCluster) Management

- **Creation**  
  Organization Users select a Plan (resource limits) to create a new Workspace.

- **Provisioning**  
  The Hexabase control plane provisions a vCluster on the Host cluster and configures itself as a trusted OIDC provider.

- **Initial Setup (within vCluster)**

  - Create preset ClusterRoles:  
    Automatically create two ClusterRoles: `hexabase:workspace-admin` and `hexabase:workspace-viewer`.  
    \*Note: Custom ClusterRole creation by users is prohibited.
  - Create default ClusterRoleBinding:  
    Automatically create a ClusterRoleBinding that binds the `hexabase:workspace-admin` ClusterRole to the `WSAdmins` group.

- **Initial Setup (within Hexabase DB)**
  - Create default groups:  
    Create three groups in a hierarchical structure: `WorkspaceMembers` (top level), `WSAdmins`, and `WSUsers`.
  - By assigning the Workspace creator to the `WSAdmins` group, they become the vCluster administrator.

## 4.3. Project (Namespace) Management

- **Creation**  
  Workspace Members (WSAdmins, etc., users with permissions) create new Projects within a Workspace.

- **Namespace Creation**  
  The Hexabase control plane creates corresponding Namespaces within the vCluster.

- **ResourceQuota Application**  
  Automatically create default ResourceQuota objects defined in the Workspace Plan in the Namespace.

- **Custom Role Creation**  
  Custom Roles valid within a Project (Namespace) can be created and edited from the UI.

## 4.4. Permission Management and Inheritance

- **Permission Assignment**  
  Assign Project Roles or preset ClusterRoles to Workspace Groups through the UI.  
  Hexabase creates and deletes RoleBindings and ClusterRoleBindings within the vCluster.

- **Permission Inheritance Resolution**
  - When a user accesses a vCluster, the OIDC provider performs the following:
    1. Recursively retrieve the user's groups and parent groups from the DB.
    2. Include a flattened group list in the `groups` claim of the OIDC token.
    3. The vCluster API server performs native RBAC authorization based on this information.

---

# 5. Technology Stack and Infrastructure

## 5.1. Applications

- **Frontend**: Next.js
- **Backend**: Go (Golang)

## 5.2. Data Stores

- **Primary DB**: PostgreSQL
- **Cache**: Redis

## 5.3. Messaging and Asynchronous Processing

- **Message Queue**: NATS

## 5.4. CI/CD (Continuous Integration/Delivery)

- **Pipeline Engine**: Tekton

  - **Reason**: Enables building Kubernetes-native declarative pipelines. Automates container builds, tests, and security scans.

- **Deployment (GitOps)**: ArgoCD or Flux
  - **Reason**:  
    Treats Git repositories as the single source of truth and declaratively manages cluster state.  
    ArgoCD has a powerful UI, while Flux excels in simplicity and extensibility. Choose based on project preferences.

## 5.5. Security and Policy Management

- **Container Vulnerability Scanning**: Trivy

  - **Role**:  
    Integrated into CI pipelines (Tekton) to scan for known vulnerabilities (CVE) in OS packages and language libraries during container image builds. Can also detect IaC misconfigurations.

- **Runtime Security Auditing**: Falco

  - **Role**:  
    Runtime threat detection tool (CNCF graduated project). Monitors system calls at the kernel level to detect and alert on events such as "unexpected shell launches within containers" or "access to sensitive files" in real-time.

- **Policy Management Engine**: Kyverno
  - **Kyverno**:  
    Low learning curve as policies can be written as Kubernetes resources (YAML), enabling intuitive management of policies like "prohibit Pod creation without specific labels" or "block use of untrusted image registries."

---

# 6. Installation and Deployment (IaC)

This project adopts **Helm** as Infrastructure as Code (IaC) to achieve "easy installation."

## 6.1. Helm Umbrella Chart

Provides a Helm Umbrella Chart that enables deployment of all Hexabase components and dependent middleware with a single command.

```yaml
apiVersion: v2
name: hexabase-ai
description: A Helm chart for deploying the Hexabase AI Control Plane
version: 0.1.0
appVersion: "0.1.0"

dependencies:
  # Define official/community Helm Chart dependencies
  - name: postgresql
    version: "14.x.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: postgresql.enabled # Can be disabled if needed
  - name: redis
    version: "18.x.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: redis.enabled
  - name: nats
    version: "1.x.x"
    repository: "https://nats-io.github.io/k8s/helm/charts/"
    condition: nats.enabled
```

**Template Examples within Chart (`templates/`)**:

- Hexabase API (Go) Deployment / Service
- Hexabase UI (Next.js) Deployment / Service
- Secrets for DB connection information (auto-generated on initial install)
- ConfigMaps for managing various settings

## 6.2. Installation Flow

End users can deploy Hexabase AI following these steps after preparing a K3s cluster:

### Add Helm Repository

```bash
helm repo add hexabase https://<your-chart-repository-url>
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### Edit Configuration File (values.yaml) (Optional):

- Edit items requiring customization such as domain names and resource allocations.

### Install with Helm:

```bash
helm install hexabase-ai hexabase/hexabase-ai -f values.yaml
```

This single command sets up the entire Hexabase control plane on the K3s cluster along with dependent components like PostgreSQL, Redis, and NATS.

# 7. Conclusion

This specification is a conceptual design blueprint for Hexabase AI based on modern technology stacks and cloud-native best practices. By incorporating simple deployment with Helm, efficient CI/CD with Tekton and GitOps, robust security with Trivy and Falco, and flexible policy management with Kyverno, we build a strong foundation for an open-source project that can be trusted by users worldwide and grow with the community.
