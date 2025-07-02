# ADR-013: Transition to Single-Organization Scoped JWT Token Authentication Flow

- **Date**: 2025-07-01
- **Status**: Proposed
- **Authors**: Platform Architecture Team

## 1. Background

Hexabase KaaS is a multi-tenant product, and as a Kubernetes as a Service (KaaS) provider, it needs to accurately and securely manage resources, billing, authorization, and rate limiting on an organization-by-organization basis.

**This ADR aims to clarify and resolve current issues within the existing authentication design documented in ADR-002: Authentication and Security.**

The current authentication design includes multiple organization IDs (`OrgIDs []string`) in the JWT Claims. This design makes it difficult to clearly determine which organization a single API request pertains to. This leads to complex logic for organization-unit rate limiting, fine-grained authorization (RBAC), and audit logging, raising concerns about accumulating technical debt.

Furthermore, while an alternative solution of including the organization context in the API path was considered to address this issue, concerns arose regarding backward compatibility problems due to API path changes, and a desire to avoid designing APIs with future modifications as a prerequisite.

## 2. Status

Proposed

## 3. Other options considered

1. **Extracting Organization ID from API Path**:
   - **Overview**: The rate limit middleware would extract the `orgId` from the API path, such as `/api/v1/organizations/:orgId/resource`, for organization-scoped API paths.
   - **Evaluation**: This pattern is common in existing API paths, ensuring consistency. However, it presented challenges: **it might necessitate API path changes (leading to backward compatibility issues), and it would be difficult to apply to common endpoints that do not include `orgId` in their path (e.g., `/auth/me`)**.
2. **Adding "Active Organization ID" to JWT Claims**:
   - **Overview**: A field like `ActiveOrgID string` would be added to JWT Claims, embedding the ID of the organization the user is currently operating on.
   - **Evaluation**: This would avoid API path changes and maintain statelessness by including the context in the JWT. However, **the coexistence of "a list of all belonging organization IDs" and "the active organization ID" in the token's Claims could still lead to complex token interpretation**. Additionally, the need to re-issue tokens every time the active organization is switched would increase the complexity of the client and the authentication flow, similar to the challenges faced by the chosen solution's final iteration.

## 4. What we propose to decide

We propose to modify the user authentication flow as follows to **always issue JWT tokens with a single organization scope**:

1. **Initial Login**:
   - The user submits authentication credentials (e.g., email/password), which are then verified by the authentication service.
   - **During login, the user explicitly specifies the `orgId` they wish to operate on as a parameter.** (e.g., by adding an `orgId` input field to the login form, or by using a URL like `/login?orgId=xxx`)
2. **Issuance of Production Access Token (Scoped Token Issuance)**:
   - The authentication service verifies that the `orgId` specified during login is valid and that the user belongs to (or is newly creating) that organization.
   - Upon successful verification, the authentication service issues a "production access token" with a standard expiration, which **contains only the specified single `orgId` in its Claims**. This token will also include the user's role within that specific organization.
3. **Subsequent API Calls**:
   - All subsequent API calls will use this single-organization-scoped access token.
4. **Organization Switching**:
   - If a user wishes to access a different organization, their current session will be effectively terminated. They will then need to restart the login process (or go through a "switch organization" flow in the UI) by specifying the new `orgId`. This will result in the issuance of a new access token containing the new `orgId`.

We propose that this decision will entail removing `OrgIDs []string` from JWT Claims and replacing it with a single `OrgID string`.

## 5. Why did you choose it?

- **Architectural Clarity**: The organization context of a request is uniquely defined by the JWT, simplifying all backend layers (especially authentication, authorization, rate limiting, and auditing). This makes system understanding, development, and maintenance significantly easier.
- **Technical Debt Reduction**: By enforcing a "single-organization scope" at the core of the authentication flow, we can prevent future complex context resolution logic and authorization vulnerabilities arising from multi-organization handling. This is crucial for a startup with limited resources.
- **Consistency with IaaS/PaaS Principles**: For a multi-tenant IaaS/PaaS, tenant (organization) isolation is a paramount requirement. This approach implements it at the foundation of authentication, strengthening the basis of security and isolation.
- **Maintenance of Backward Compatibility**: Since there is no need to change API paths, we can maintain existing API design principles and minimize the risk of future compatibility issues.
- **Opportunity with No Existing Users**: The product is currently under development with no existing users, making this an excellent opportunity to introduce a fundamental change to the authentication flow without impacting existing users.

### 6. Why didn't you choose the other option

1. **Extracting Organization ID from API Path**:
   - This approach might necessitate API path changes in some cases, and **backward compatibility issues** during such changes were a major concern. Modifying API paths after release can significantly impact clients in the future.
   - It was difficult to apply this approach to generic API endpoints (e.g., `/auth/me`) that do not include `orgId` in their path.
2. **Adding "Active Organization ID" to JWT Claims**:
   - This would lead to the Claims containing both "a list of all belonging organizations" and "the active organization ID", potentially **increasing the complexity of the Claims themselves and leading to ambiguity in interpretation**. The context provided by the token should be as clear as possible.
   - It would not allow us to fully reap the main benefit of this decision: unifying to a single-organization scope.

### 7. What has not been decided

- **Handling of `orgId` during initial signup**: How will `orgId` be determined and registered when a new user signs up (e.g., user input, system auto-generation)?
- **Details of UX improvements for multi-organization users**: How will a "switch organization" button be implemented in the UI to provide a seamless experience, abstracting the internal re-authentication process (e.g., managing cookies or local storage, redirection methods during re-authentication)?
- **Behavior when attempting to log in with a non-existent or unauthorized `orgId`**: What kind of errors will be returned, and how will the UI guide the user?

### 8. Considerations

- **Implementation Effort**: Modifying the authentication flow will involve extensive changes across both the backend (API) and clients (UI, CLI). Specifically, significant changes will be required in `jwt.go`, `domain/models.go`, `service.go` within `api/internal/auth/` and their respective handlers.
- **Usability**: Users will need to input an organization ID during login. New users will need to know which `orgId` to use, and users belonging to multiple organizations will need to specify which `orgId` they intend to access. This will require careful UI/UX design. We believe that introducing aliases in the future can lower this barrier.
- **Documentation**: New authentication flow, JWT Claims structure, and UI/API specifications for organization switching need to be clearly documented for developers.

### 9. Note

- This ADR is a outlining a fundamental shift in the authentication model to enhance clarity, security, and scalability for Hexabase KaaS as a multi-tenant platform. This change is considered a strategic investment given the current stage of product development.
- **This proposal directly builds upon and modifies the existing authentication design documented in [ADR-002: Authentication and Security](./001-platform-architecture.md).**
