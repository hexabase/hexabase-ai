## **Hexabase KaaS: Addendum Specification for VM and Application Management**

- **Document Version:** 1.0 (Addendum)
- **Date:** June 7, 2025
- **Applies to:** Hexabase KaaS: Architecture Specification v2.0

### **1. Introduction**

This document serves as an addendum to the core Hexabase KaaS Architecture Specification. It details the new functional capabilities and underlying architectural enhancements required to empower users with direct management of Virtual Machines (Nodes) and a simplified, UI-driven workflow for basic container operations (Applications).

These features fulfill the vision of providing a seamless scalability path, allowing users to transition from a limited, shared trial environment to a robust, production-grade infrastructure with dedicated resources.

### **2. Core Concept Enhancements**

To support the new features, the following concepts are introduced and mapped to their underlying infrastructure equivalents. This table should be considered an extension of the primary Core Concepts mapping table.

| Hexabase Concept        | Kubernetes / Infrastructure Equivalent                | Scope                       | Notes                                                                                               |
| :---------------------- | :---------------------------------------------------- | :-------------------------- | :-------------------------------------------------------------------------------------------------- |
| **Shared Node Plan**    | `ResourceQuota` on Multi-Tenant Nodes                 | vCluster                    | Default trial plan on shared infrastructure with strict resource limits. Users do not manage Nodes. |
| **Dedicated Node Plan** | **Dedicated Proxmox VM** + `Node Taints/Tolerations`  | Host K3s Cluster / vCluster | Paid plan with user-provisioned, isolated VMs providing guaranteed resources and performance.       |
| **Application**         | `Deployment` or `StatefulSet` + `Service` + `Ingress` | vCluster                    | A user-deployed workload, managed as a single logical unit in the UI for simplified operations.     |

### **3. New and Enhanced Functional Specifications & User Flows**

#### 3.1. Workspace Lifecycle & Scaling Path

The user journey within a Workspace is enhanced to provide a clear upgrade path:

1.  **Initial State (Shared Node Plan):** A new Workspace is automatically provisioned on the Shared Node Plan. The user operates within a strict `ResourceQuota` on shared, multi-tenant infrastructure. This allows for immediate, low-cost experimentation.
2.  **Trigger for Upgrade:** As application needs grow, the UI will proactively notify the user when they approach their resource quota limits. This notification, along with prominent UI elements, will guide them to the Node Management section.
3.  **Transition to Dedicated Plan:** The user initiates the "Add New Dedicated Node" wizard. The successful creation of the _first_ dedicated node automatically transitions the Workspace's billing and operational model to the **Dedicated Node Plan**.

#### 3.2. Feature: Dedicated Node Management

This new feature set, accessible from within a Workspace, allows users to provision and manage the lifecycle of their dedicated compute resources.

##### 3.2.1. Node Overview Page

- **Purpose:** Provides a centralized dashboard for a Workspace's compute resources.
- **Specification:** The UI is conditional based on the Workspace plan.
  - **On Shared Plan:** It displays quota usage graphs (CPU, Memory, Pods) and features a prominent Call-to-Action (CTA) to launch the "Add New Dedicated Node" wizard, explicitly framing it as an upgrade.
  - **On Dedicated Plan:** It displays aggregate resource usage across all dedicated nodes and provides entry points to manage them.

##### 3.2.2. Dedicated Node List Page

- **Purpose:** To list and manage all dedicated nodes associated with the Workspace.
- **Specification:**
  - **UI:** A table view with columns for `Node Name`, `Status` (combined Proxmox/K3s status), `Specifications` (vCPU, RAM, Disk), `IP Address`, and `Created At`.
  - **Actions:** Each node has a menu for `Start`, `Stop`, `Reboot`, and `Delete` operations. These actions trigger orchestrated workflows in the backend that interact with both the Proxmox and K3s APIs.

##### 3.2.3. 'Add New Dedicated Node' Wizard

- **Purpose:** A guided, multi-step process for provisioning new Proxmox VMs and adding them to the user's vCluster.
- **Specification:**
  - **Step 1: Plan Selection & Billing Confirmation:**
    - Presents a clear notice of the plan upgrade and associated costs.
    - Users select from a list of predefined instance types (e.g., `S-Type`, `M-Type`) with transparent monthly pricing.
    - Displays a dynamic summary of the new estimated monthly cost.
    - Requires explicit user consent via a checkbox before proceeding.
  - **Step 2: Node Configuration:**
    - Users provide a `Node Name` prefix and select a `Region` (if applicable).
    - Optionally, they can add a public `SSH Key` for direct VM access.
  - **Step 3: Provisioning:**
    - The UI displays a real-time progress screen, providing feedback as the backend orchestrates the multi-stage provisioning process (`Creating VM in Proxmox...`, `Installing K3s agent...`, `Joining node to cluster...`).

#### 3.3. Feature: Application Lifecycle Management

This new feature set provides an intuitive, UI-driven interface for deploying and managing containerized workloads ("Applications").

##### 3.3.1. Application List View

- **Purpose:** A central dashboard within a Project to view all deployed applications.
- **Specification:** A table lists applications with their `Name`, `Status`, `Type` (Stateless/Stateful), and public `Endpoints`. A "+ Create New Application" button initiates the deployment wizard.

##### 3.3.2. "Create New Application" Wizard

- **Purpose:** A guided flow to deploy containers without requiring knowledge of Kubernetes YAML.
- **Specification:**
  - Users are guided through logical steps: Application Type, Source (Image or Git), Configuration (Replicas, Ports), Networking (optional Ingress creation), Storage (for Stateful types), and Environment Variables.
  - **New Scheduling Step:** For Workspaces on the Dedicated Node Plan, an additional step allows the user to select a target node pool. The UI translates this selection into the appropriate `nodeSelector` or `tolerations` in the underlying Kubernetes manifest, ensuring the application is scheduled onto the desired dedicated hardware.

##### 3.3.3. Application Detail View

- **Purpose:** A drill-down dashboard for a single application.
- **Specification:** A tabbed interface provides access to:
  - **Overview:** Key metrics and resource links.
  - **Instances (Pods):** A list of running pods with status and actions (`View Logs`, `Restart`).
  - **Logs:** A real-time, aggregated log stream.
  - **Settings:** An interface to modify application parameters like image version or replica count.

### **4. Architectural and Technology Stack Additions**

#### 4.1. Architectural Enhancements

To support these features, the **vCluster Orchestrator** component's responsibilities are expanded:

- **Amended Specification (from Section 2.3):**
  > The orchestrator ... handles applying OIDC settings, installing HNC, setting resource quotas, and controlling **Dedicated Node allocation**. This involves interacting with an underlying virtualization platform API (e.g., **Proxmox VE**) to provision or de-provision VMs, which are then configured using tools like `cloud-init` to securely install the K3s agent and join the specific Workspace's vCluster. The allocation is managed using Node Selectors and Taints/Tolerations to ensure workloads are scheduled on the correct dedicated resources.

#### 4.2. Technology Stack Additions

The following component is added to the official technology stack.

##### 4.2.1. Virtualization Layer

- **Technology**: Proxmox VE
- **Reason**: A robust, open-source virtualization platform with a comprehensive API. It allows HKS to programmatically manage the entire lifecycle of dedicated VMs (Nodes) for tenants, providing the strong resource isolation required for the Dedicated Node Plan.
