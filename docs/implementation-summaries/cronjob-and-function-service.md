# CronJob and Function Service Implementation Summary

This document details the architecture and implementation plan for two new major features in Hexabase KaaS: native CronJob management and a serverless platform, "HKS Functions," powered by Knative.

---

## 1. CronJob Management

This feature allows users to easily create and manage scheduled, recurring tasks within their projects using a standard, robust Kubernetes-native approach.

### 1.1. User Experience and Workflow

1.  **New "Application" Type**: In the UI, when creating a new application, the user can select a new type: **"CronJob"**.
2.  **Configuration UI**:
    - **Task Definition (Template-based)**: Instead of manually entering image details, the user can select an existing "Stateless" Application from the same Project via a dropdown. This action populates the CronJob template with the selected application's container image, environment variables, and resource requests. This provides an intuitive "run a task from this app" experience.
    - **Command Override**: Users can override the container's default command/arguments specifically for this job.
    - **Schedule Configuration**: A user-friendly UI will be provided for setting the schedule (e.g., presets for "hourly," "daily," "weekly") which translates to a standard cron expression. An advanced input field will also be available for users to enter a raw `cron` expression directly.
3.  **Management**: Users can view a list of their CronJobs, see the last execution time and result, view logs from past runs, and trigger a manual run.

### 1.2. Backend Implementation

- The HKS API server will receive the user's configuration and translate it into a standard Kubernetes `batch/v1.CronJob` resource manifest.
- This manifest will be applied to the tenant's vCluster. The `spec.jobTemplate` will contain the full pod specification derived from the selected application template and user overrides.
- The entire lifecycle (scheduling, job creation, pod execution, cleanup) is handled by the native Kubernetes CronJob controller within the vCluster, ensuring stability and reliability.

---

## 2. HKS Functions (Serverless Platform)

HKS Functions will provide a managed Function-as-a-Service (FaaS) experience, enabling developers to deploy and run event-driven code without managing underlying infrastructure. This feature will be built on top of **Knative**.

### 2.1. Core Architecture

1.  **Platform-Level Knative**: The **Knative Serving** and **Knative Functions (`kn func`)** components will be installed by platform administrators once on the **Host K3s Cluster**. This forms the serverless backbone for all tenants.
2.  **Developer-Centric Experience (DevEx)**: The primary interface for developers will be a dedicated CLI tool.
    - **`hks-func` CLI**: This will be a wrapper around the standard `kn func` CLI. It will automate HKS authentication and context configuration (e.g., fetching the correct Kubeconfig for the target project).
    - **Sample Workflow**:
      ```bash
      $ hks login
      $ hks project select my-serverless-project
      $ hks-func create -l node my-function
      # ... edit function code ...
      $ hks-func deploy
      ```
3.  **UI for Management**: The HKS UI will serve as a management and observability dashboard for deployed functions, allowing users to:
    - View a list of all functions within a project.
    - See function status, invocation endpoints (URLs), and resource consumption.
    - Access real-time logs and performance metrics.

### 2.2. Function Invocation Patterns

- **HTTP Trigger**: Knative automatically provides a public URL for every deployed function. This is the primary way to invoke functions.
- **Scheduled Trigger**: Functions can be invoked on a schedule by using the **CronJob feature** described above. The CronJob simply needs to run a container with `curl` or a similar tool to hit the function's HTTP endpoint at the scheduled time.

### 2.3. AI-Powered Dynamic Function Execution (Advanced)

This powerful capability allows AI agents to generate, deploy, and execute code on-the-fly in a secure sandbox.

**Execution Flow and Security Model:**

1.  **Code Generation**: An AIOps agent generates a piece of code to perform a specific task.
2.  **Dynamic Deploy Request**: The agent calls a secure **Internal Operations API** on the HKS Control Plane (e.g., `POST /internal/v1/operations/deploy-function`), passing the code and the short-lived internal JWT.
3.  **Secure Build & Deploy**:
    - The HKS backend receives the request and uses a secure, isolated in-cluster builder (e.g., **Kaniko**) to build a temporary container image from the provided code.
    - It then deploys this image as a new Knative Function to the user's vCluster. The function runs with a highly restricted, single-purpose Service Account.
4.  **Scoped Invocation**: The backend returns the function's internal URL to the agent. The agent invokes the function to get the result.
5.  **Automatic Cleanup**: After execution (or a timeout), the agent (or a garbage collector) calls another internal API (`delete-function`) to remove the temporary function and its associated resources.

**Developer Tooling:**

- An **HKS Internal SDK** (Python) will be developed to abstract this flow. The AI agent can simply call methods like `hks_sdk.functions.execute(code="...")`, and the SDK will handle the entire secure deploy-invoke-cleanup lifecycle.
- Detailed documentation will be provided, outlining the capabilities and limitations (e.g., available libraries, resource quotas) of this dynamic execution environment.
