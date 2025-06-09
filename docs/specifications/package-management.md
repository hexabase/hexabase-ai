# HKS Package Management

## 1. Overview

This document specifies the design for the HKS Package Management feature. This allows Workspace Admins to create a curated catalog of software packages (backed by Helm charts) that can be easily deployed as applications within any project in their workspace. The goal is to standardize application deployment, simplify configuration, and integrate seamlessly with both the HKS UI and the AIOps Chat Agent.

## 2. Core Concepts

### 2.1. Package

A `Package` is a template for an application that a Workspace Admin registers with the workspace. It consists of:

- **Metadata**: A user-friendly name, description, and icon.
- **Source**: A pointer to a Helm chart in a Git repository (URL, branch/tag, and path to the chart).
- **Configuration Schema**: A schema that defines the configurable values for the Helm chart. HKS will automatically parse the chart's `values.yaml` to generate a basic schema. Optionally, for a more user-friendly experience, a `questions.yaml` file can be included in the chart directory to define custom questions, variable types, validation rules, and grouping for the UI and AIOps wizard.

### 2.2. Application

An `Application` is an instance of a `Package` deployed into a specific `Project`. It represents a running Helm release and stores the specific values used for its deployment.

## 3. Architecture and Flows

### 3.1. Package Registration Flow

1.  A Workspace Admin navigates to the "Packages" section of their workspace.
2.  They choose to add a new package and provide the Git repository URL, the branch/tag/commit to use, and the path to the chart within the repository.
3.  The HKS Control Plane fetches the chart from the Git repository.
4.  The system looks for a `questions.yaml` file to build a user-friendly wizard. If not found, it parses the `values.yaml` file to generate a default form schema.
5.  The `Package` metadata and the generated schema are stored in the HKS database, associated with the workspace. The package is now available in the workspace's catalog.

### 3.2. UI-based Installation Flow

1.  A User with appropriate permissions navigates to a `Project` and chooses to create a new `Application`.
2.  They are presented with an option to "Install from Package".
3.  The UI displays the list of `Packages` available in the current workspace.
4.  Upon selecting a `Package`, the UI dynamically renders an input form based on the stored configuration schema.
5.  The user fills in the required values and submits the form.
6.  **(Pre-flight Check)** Before proceeding, the HKS Control Plane performs a pre-flight quota check. See section 3.4 for details.
7.  If the check passes, the HKS Control Plane uses its internal Helm client to execute a `helm install` command within the project's vCluster, using the selected chart and the user-provided values.
8.  If the check fails, the UI displays a clear error message to the user explaining which quota would be exceeded.

### 3.3. AIOps-based Installation Flow

The AIOps agent provides a conversational, wizard-like experience for package installation.

1.  **User**: "I want to deploy a monitoring stack in the 'dev' project."
2.  **UserChatAgent & OrchestrationAgent**: The query is routed to the `HelmAgent`.
3.  **HelmAgent**: The agent searches the workspace's `Package` catalog for items matching "monitoring stack" (e.g., a package for `kube-prometheus-stack`).
4.  **AIOps**: "I found a 'Prometheus' package. I can install it for you. I just need to ask a few questions. First, what do you want to name this installation?"
5.  The agent proceeds to ask questions based on the `Package`'s configuration schema (`questions.yaml` or `values.yaml`).
6.  Once all required values are collected, the `HelmAgent` initiates the installation, which includes the pre-flight quota check (see section 3.4).
7.  If the check is successful, the `HelmAgent` makes a request to the HKS Internal API to perform the installation. If it fails, the agent reports the specific quota error back to the user in the chat.

### 3.4. Pre-flight Quota Check

It is critical to ensure that a package can be deployed within a project's resource limits _before_ the installation begins. A partial installation due to a quota violation can leave the system in an inconsistent state. HKS will implement a server-side dry-run to validate resource requests against project quotas.

The process is as follows:

1.  **Template Generation**: The HKS backend runs `helm template` with the user-provided values to generate the final Kubernetes manifests for all resources in the chart.
2.  **Server-Side Dry-Run**: The system then sends these manifests to the target project's Kubernetes API server with a "server-side dry-run" flag (`kubectl apply --dry-run=server` equivalent).
3.  **Quota Admission Control**: The Kubernetes API server receives the resources and processes them through its admission controller chain, including the `ResourceQuota` controller. This controller calculates the cumulative resources requested by the chart and compares them against the currently available quota in the target namespace.
4.  **Validation Result**:
    - **Success**: If all resources fit within the project's quotas, the API server returns a success message without persisting any objects. The HKS backend then proceeds with the actual `helm install`.
    - **Failure**: If any quota would be exceeded, the API server rejects the dry-run request and returns a descriptive error (e.g., `Error: failed to create resource: pods "my-pod" is forbidden: exceeded quota: my-quota, requested: pods=1, used: pods=5, limited: pods=5`).
5.  **User Feedback**: This specific error message is captured and relayed back to the user, either in the UI or via the AIOps agent, allowing them to adjust their request or increase project quotas.

This approach leverages the native, authoritative validation logic of Kubernetes itself, providing the most accurate pre-installation check possible.

## 4. `questions.yaml` Schema (Recommended)

To provide the best user experience for both UI forms and the AIOps wizard, we recommend defining an optional `questions.yaml` file within the chart's directory. This provides more control over the user interaction than parsing `values.yaml` alone.

**Example `questions.yaml`:**

```yaml
questions:
  - group: "General Settings"
    questions:
      - variable: "replicaCount"
        label: "Number of Replicas"
        description: "How many pods do you want to run?"
        type: "int"
        default: 1
        required: true
  - group: "Database Settings"
    questions:
      - variable: "database.user"
        label: "Database Username"
        type: "string"
        required: true
      - variable: "database.password"
        label: "Database Password"
        type: "password"
        required: true
```

## 5. API and UI Considerations

- **API**: New endpoints under `/api/v1/workspaces/{ws_id}/packages` will be needed to manage the lifecycle of `Packages`. The application creation endpoint will be extended to support installation from a package.
- **UI**:
  - A new "Packages" page at the workspace level for managing the catalog.
  - The "New Application" dialog within a project will be updated to include an "Install from Package" option, which leads to the catalog browser and the dynamic configuration form.
