# ADR-009: Secure and Scalable Multi-Tenant Logging Architecture

**Date**: 2025-06-15
**Tags**: logging, security, observability, multi-tenancy

## 1. Background

As a multi-tenant platform, HKS must provide robust logging and auditing capabilities for different audiences (Platform SREs, Organization Admins, Users). Our current structured logs are a good technical foundation but lack the tenancy and user context (`organization_id`, `user_id`) required to build secure, role-based access controls. This poses a significant security risk, as there is no mechanism to prevent one tenant from potentially accessing another's logs. We need a formal architecture to guarantee strict data isolation.

## 2. Status

**Proposed**

## 3. Other Options Considered

### Option A: Frontend-based Filtering

- The backend API provides a broad set of logs, and the frontend UI is responsible for filtering them based on the user's role.

### Option B: Per-Tenant Data Stores

- Provision a completely separate logging stack (e.g., a dedicated Loki and ClickHouse instance) for each tenant organization.

### Option C: Centralized Storage with Backend-Enforced Access Control

- Use a shared, centralized logging backend (Loki/ClickHouse) for all tenants. All client access is brokered through a secure HKS API gateway that enriches logs and strictly enforces access control rules based on the authenticated user's identity.

## 4. What Was Decided

We will implement **Option C: Centralized Storage with Backend-Enforced Access Control**.

The core components of this decision are:

1.  **Three-Tiered Logging**: Logs are categorized into `System Logs` (for SREs), `Audit Logs` (for users/admins), and `Workload Logs` (for application owners), each with its own storage and access rules.
2.  **Mandatory Log Enrichment**: The API backend will implement a logging middleware to automatically inject `organization_id`, `workspace_id`, and `user_id` into every structured log generated within an authenticated user's request context.
3.  **Secure API Proxy**: The UI and other clients will **never** access logging data stores directly. All requests will be proxied through the HKS API, which acts as a secure gatekeeper.
4.  **Backend-Enforced Scoping**:
    - For **Audit Logs**, the API will inject non-negotiable `WHERE` clauses into ClickHouse queries based on the user's JWT claims.
    - For **Workload Logs**, the API will perform a Kubernetes `SubjectAccessReview` to validate the user's pod-level permissions before streaming any log data.

## 5. Why Did You Choose It?

This approach provides the best balance of security, scalability, and usability:

- **Security**: It establishes a single, secure gateway for all log access, eliminating the possibility of client-side bypass and ensuring that access control logic is centralized and consistently enforced.
- **Scalability**: It leverages shared infrastructure, which is more cost-effective and operationally manageable than provisioning a separate stack for every tenant.
- **Usability**: It enables the platform to provide rich, role-aware observability features within the UI, powered by a secure and flexible API.

## 6. Why Didn't You Choose the Other Option?

- **Option A (Frontend Filtering)** was rejected because it is fundamentally insecure. Relying on the client to enforce security is a critical anti-pattern; a malicious user could easily bypass the UI and query the API directly to access unauthorized data.
- **Option B (Per-Tenant Stacks)** was rejected due to excessive operational complexity and cost. Managing thousands of separate logging instances is not feasible and would not scale economically.

## 7. What Has Not Been Decided

- The precise log retention policies (e.g., 7 vs. 30 vs. 90 days) for each log category.
- The specific design and content of the default Grafana dashboards for each user role.
- The exact schema for the `details_json` field in the audit log table.

## 8. Considerations

- **Performance**: The logging enrichment middleware must be highly performant to avoid adding significant latency to API requests.
- **Implementation**: Requires disciplined implementation of the secure API endpoints and careful construction of the database/Loki queries to prevent injection or scoping bugs.
- **AIOps**: The AIOps agent must also have its actions logged through this system, correctly impersonating the user on whose behalf it is acting. The `initiated_by` field will be used to differentiate user vs. agent actions.
