# Relationship between Hexabase KaaS and K3s

This document details the relationship between the Hexabase KaaS platform and the underlying K3s cluster, clarifying the role of each component, and the scope of resources and nodes that users can manage.

## 1. Role Model: Orchestrator and Execution Environment

The relationship between Hexabase KaaS and K3s can be understood as a separation between an "Orchestrator" (Hexabase KaaS) and an "Execution Environment" (K3s).

- **Hexabase KaaS (Orchestrator)**

  - Acts as the central **control plane** for users.
  - Provides a user-friendly UI (Next.js) and a management API (Go).
  - Abstracts complex Kubernetes concepts into `Organizations`, `Workspaces`, and `Projects`.
  - Manages the entire lifecycle of tenant environments (vClusters), including provisioning, configuration, user access (OIDC), and resource allocation.
  - Handles multi-tenant concerns such as billing, user management, and hierarchical permissions.
  - It does **not** run user workloads directly. It orchestrates the environment where workloads will run.

- **Host K3s Cluster (Execution Environment)**
  - Acts as the foundational **host cluster** where all components run.
  - Runs the Hexabase KaaS control plane components (API server, workers, etc.).
  - Hosts the virtualized tenant environments (`vClusters`). A `Workspace` in Hexabase KaaS maps directly to a `vCluster` instance running on the host K3s.
  - Manages the physical (or virtual) infrastructure, including nodes, networking, and storage.
  - User-deployed applications and pods run inside the `vClusters`, not directly on the host K3s cluster.

**In summary:** Users interact with the Hexabase KaaS platform. KaaS then translates user actions into technical operations on the host K3s cluster to manage the isolated `vCluster` environments provided to tenants.

## 2. Scope of Manageable K3s Resources

Tenant resource management is strictly confined within the boundaries of their assigned **Workspace (vCluster)**. This ensures strong isolation and prevents tenants from impacting each other or the host cluster.

| Scope                     | Manageable by Tenant? | Details                                                                                                                                                                     |
| ------------------------- | :-------------------: | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Workspace (vCluster)**  |          Yes          | Tenants have near-full administrative access _within_ their vCluster. They can manage most standard Kubernetes resources inside it.                                         |
| ↳ **Project (Namespace)** |          Yes          | Tenants can create, manage, and isolate resources within `Projects`, which correspond to Kubernetes `Namespaces` inside their vCluster.                                     |
| ↳ **Standard Resources**  |          Yes          | `Deployments`, `Services`, `Pods`, `ConfigMaps`, `Secrets`, `Ingresses`, etc., within their Projects.                                                                       |
| ↳ **RBAC (Role-Based)**   |        Limited        | - Can create `Roles` and `RoleBindings` within a Project (Namespace).<br>- Cannot create `ClusterRoles`; must use pre-defined ones (`workspace-admin`, `workspace-viewer`). |
| **Host K3s Cluster**      |          No           | Tenants have no direct access to the host K3s cluster's API server or its resources.                                                                                        |
| ↳ **Host Resources**      |          No           | `Nodes`, host-level `PersistentVolumes`, `StorageClasses`, host `Namespaces`, and host cluster configurations are all managed by the platform and are inaccessible.         |
| ↳ **CRDs on Host**        |          No           | Installation of new Custom Resource Definitions (CRDs) on the host cluster is a platform-level administrative task.                                                         |

## 3. Scope of Manageable Nodes

Node management is a critical aspect of security and resource isolation. In Hexabase KaaS, direct node access is abstracted away from tenants, but resource allocation can be influenced by the selected plan.

- **Default Behavior (Shared Nodes)**

  - By default, all vClusters and their workloads run on a shared pool of nodes within the host K3s cluster.
  - Tenants cannot choose which node their pods run on. The K3s scheduler, managed by the host, handles pod placement.
  - Tenants have no `ssh` access or `kubectl debug` permissions for any nodes.

- **Dedicated Nodes (Premium Plans)**
  - Hexabase KaaS supports assigning **dedicated nodes** to specific Workspaces, typically offered under premium subscription plans.
  - This provides a higher level of isolation for performance and security.
  - The assignment is managed by the Hexabase KaaS control plane using Kubernetes taints and tolerations, and node selectors.
    - A specific `taint` (e.g., `dedicated=ws-xxxxx:NoSchedule`) is applied to the node.
    - The corresponding `toleration` and a `nodeSelector` (e.g., `hexabase.ai/node-pool: ws-xxxxx`) are automatically added to the pods within the tenant's vCluster.
  - Even with dedicated nodes, tenants **do not gain direct administrative access** to the node itself. The management remains the responsibility of the platform administrators. The benefit is guaranteed resource allocation and workload isolation at the hardware level.

This model allows Hexabase KaaS to provide the flexibility of dedicated resources while maintaining a secure, abstracted, and centrally managed infrastructure.
