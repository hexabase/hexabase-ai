# ADR-011: Proposal to Migrate Identity Broker Functionality to Libraries

- **Date**: 2025-06-18
- **Status**: Proposed
- **Authors**: Platform Architecture Team

## 1. Background

### Context

The current Hexabase kaas product acts as an "**Identity Broker**," processing authentication results (tokens) from external Identity Providers (IdPs) and managing its own access tokens, refresh tokens, and login sessions based on these results.

### Problem Statement

Continuously developing authentication-related functionalities in-house leads to potential security concerns, challenges in keeping up with standard specifications, and increased operational and maintenance costs. Specifically, there is a known bug in the existing refresh token functionality, which poses security risks and stability issues.

**This ADR addresses more than just internal code improvements. It is a strategic investment to replace the core authentication functionality, moving from an unstable custom implementation to a robust, globally standardized foundation.** This initiative will directly contribute to improving service reliability and reducing development and operational costs in the future.

## 2. Status

Proposed

## 3. Other options considered

### Option A: Utilize Fosite (Go-language OAuth 2.0/OpenID Connect Authorization Server SDK)

* **Overview:** A comprehensive SDK for building OAuth 2.0/OpenID Connect authorization servers in Go. It provides robust protocol-level functionalities.
* **Evaluation:**
    * **Pros:** Expected to implement OAuth/OIDC best practices, contributing to fixing refresh token bugs. Provides full control.
    * **Cons:** Primarily intended for building "authorization servers," which is overkill for the current "Identity Broker" role (validating external tokens and issuing custom tokens). Requires implementing a full suite of authorization server endpoints, leading to high modification costs to align with existing flows. High learning curve.

### Option B: Utilize Ory Hydra (Go-based OAuth 2.0/OpenID Connect Authorization Server Service)

* **Overview:** Provides OAuth 2.0/OpenID Connect authorization server functionality as a microservice. Designed to delegate user authentication to external services.
* **Evaluation:**
    * **Pros:** High robustness as an authorization server, and easier to clarify responsibilities as a decoupled microservice.
    * **Cons:** Similar to Fosite, it's an "authorization server" and overkill for the current Identity Broker role. Requires deploying and operating an independent service, increasing introduction costs and operational burden on the existing system. Potential challenges in integrating with custom token management logic.

### Option C: Utilize Authlete (OAuth 2.0/OpenID Connect BaaS)

* **Overview:** Provides OAuth 2.0/OpenID Connect protocol processing as an API via a Backend as a Service.
* **Evaluation:**
    * **Pros:** Eliminates the need for in-house OAuth/OIDC expertise, enabling rapid and reliable construction of a secure authentication foundation. Significantly reduces operational burden. Can fundamentally resolve refresh token bugs.
    * **Cons:** **Not free.** Incurs external service dependency and usage-based charges, requiring cost consideration.

## 4. What was decided - Proposal

This document proposes to replace the Identity Broker functionality of the Kaas product through **optimization of existing Go libraries and custom implementation**. This approach involves **rebuilding and strengthening** the current authentication functionality using proven open-source libraries and best practices.

Specifically, this involves leveraging the following set of libraries and integrating them to **reconstruct and improve** the existing code:

* **JWT Generation and Validation:** `github.com/golang-jwt/jwt/v5`
* **OpenID Connect Client and ID Token Validation:** `github.com/coreos/go-oidc` (which includes `golang.org/x/oauth2`)
* **Session Management:** `github.com/alexedwards/scs` or `github.com/gorilla/sessions`
* **Database:** For persistent storage and state management of custom access tokens, refresh tokens, and session information.

## 5. Why this proposal is necessary (Rationale for Proposal)

This proposal is essential for resolving the challenges faced by the current authentication infrastructure and building a more robust and reliable system.

* **Optimal Role Fit:** The proposed set of libraries is **specifically tailored to the Identity Broker role** (validating external tokens and converting them into custom tokens) that the Kaas product fulfills, without unnecessary features. This enables efficient and precise improvements.
* **Cost-Effectiveness:** All selected libraries are open source and incur no additional license fees or service charges. This allows for building a **high-quality, secure foundation while meeting the requirement of being free**.
* **Ease of Integration with Existing System:** Being Go-language libraries, they can be directly integrated into the existing Go application, eliminating the need for major system architecture changes or deployment of new independent services (like Hydra). This ensures a **smooth transition within the development team's familiar environment**.
* **Direct Problem Resolution and Strategic Investment:** The existing refresh token bug can be fundamentally resolved by using `golang-jwt/jwt` for robust token generation and by implementing **refresh token rotation** with database utilization. Furthermore, external IdP ID token validation will become **strictly compliant with OpenID Connect standards** via `go-oidc`, enhancing security. This is not merely technical debt repayment but a **strategic investment** contributing to improved system reliability and reduced future operational costs.

## 6. Why the Alternative Was Not Recommended

* **Option A (Fosite) and Option B (Ory Hydra):**
    * These tools are primarily designed for building "authorization servers," which are overkill for the current Identity Broker role of the Kaas product. Especially, Hydra requires operation as an independent service, which would **introduce new operational burdens and complexity to the existing system**, making it unsuitable for current requirements.
* **Option C (Authlete):**
    * While this service offers highly robust functionality, it **does not meet the crucial requirement of being free.** Usage-based charges apply, and dependence on an external service is not acceptable.

## 7. Undecided Matters

This proposal requires discussion and consensus. If it is decided to adopt this proposal, the following points will need to be decided next:

* **Details of Refresh Token Persistence:** Which database (existing or dedicated) will be used, and what schema will be employed to store refresh tokens? Also, the specific implementation of refresh token rotation (e.g., one-time usable tokens or time-limited rotation).
* **Details of Session Storage:** Whether to use in-memory storage or an external KVS (like Redis).
* **Error Handling and Logging Strategy:** Detailed error handling for authentication and authorization processes, and the logging level required for security audits.
* **Scope of Existing Custom Logic Reconstruction:** How much of the existing authentication-related code will be replaced by libraries, and how much will be reimplemented in-house.

## 8. Considerations for Adoption

The following points must be considered when proceeding with this proposal:

* **Implementation Cost:**
    * The introduction of new libraries and the **reconstruction** of existing code will require a significant amount of development effort, including code migration and updating import paths.
* **Regression Risk:**
    * Large-scale **codebase changes** can introduce unintended behavioral changes. Thorough testing will be critical during the migration process.
* **Team's Learning Curve:**
    * Time will be needed for the entire team to understand the new libraries and implementation patterns.
* **Impact on Concurrent Development:**
    * During the **code changes**, there is a higher risk of conflicts with other ongoing development tasks. A clear branching strategy and close communication will be essential.
* **Security Enhancement:**
    * **JWT Key Management:** Private keys used for signing custom tokens must be managed securely (e.g., using environment variables, secret management services).
    * **Refresh Token Rotation:** Implementing robust logic to issue a new refresh token and invalidate the old one each time a refresh token is used will significantly reduce the risk of refresh token theft.
    * **Input Validation:** Strict validation must be performed on all inputs, including callback data from external IdPs, to protect the system from malicious data or attacks.
    * **Strict HTTPS Usage:** HTTPS must be used for all communications.
* **Performance and Scalability:**
    * Token validation and database access will occur frequently, so careful design of indexing and consideration of caching strategies (if necessary) are required to avoid performance bottlenecks.
    * Database load balancing should be considered in anticipation of increased refresh token usage.
* **Migration Strategy:** A smooth migration plan from the existing custom implementation to the new library-based system should be devised, minimizing service downtime.