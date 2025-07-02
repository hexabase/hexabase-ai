# ADR-012: API Schema-Driven Development

- Date: 2025-06-24
- Status: Accepted
- Authors: Platform Architecture Team

## 1. Background

### Context

The Hexabase KaaS product currently utilizes an AI-assisted process to generate API implementation code and basic Markdown-based API reference documentation. This approach excels in rapid prototyping and initial feature development.

### Problem Statement

As the product evolves, we plan to develop client-side applications (CLI, Frontend UI), create SDKs for external developers, and publish official API documentation. The current development process presents several challenges for this future growth:

- Lack of a Single Source of Truth: The API's "contract" is implicitly defined within the Go code and manually maintained in Markdown documents. This makes it difficult to ensure consistency and accuracy.
- Inefficient Client Development: Without a machine-readable schema, client-side developers must manually write API clients and data models, which is time-consuming and prone to error.
- Maintenance Overhead: Any changes to the API require manual updates to both the implementation and the documentation, increasing the risk of them becoming out of sync.
- Inability to Leverage Tooling: We cannot utilize the rich ecosystem of tools for automated testing, validation, and documentation generation that rely on standardized API schemas.

## 2. Status

Accepted

## 3. Other options considered

### Option A: Continue with the current AI-assisted manual process

- Summary: Maintain the existing workflow of generating code and manually creating/updating Markdown documentation as needed.
- Evaluation:
  - Pros: No changes to the current development process are required. Fast for initial drafts.
  - Cons: This does not solve any of the problems stated above. It will accumulate technical debt and significantly slow down future client-side and external-facing development.

### Option B: Adopt a strict "Code-First" approach

- Summary: Annotate the Go handler code and DTOs, and then automatically generate the OpenAPI schema from the code.
- Evaluation:
  - Pros: Familiar workflow for backend developers. The code remains the single source of truth.
  - Cons: API design can become inconsistent as it's driven by individual implementations. It also hinders parallel development, as client teams must wait for the backend implementation to be complete before a stable schema is available.

### Option C: Adopt a strict "Schema-First" approach

- Summary: Design the complete OpenAPI schema first, then generate server-side code stubs from it. The backend team implements the business logic against these generated interfaces.
- Evaluation:
  - Pros: Enforces API design consistency and facilitates parallel development.
  - Cons: Can be rigid. Implementation-time discoveries (e.g., performance issues, technical constraints) can lead to significant rework of the schema, causing delays. Past experience shows this can become a development bottleneck.

## 4. Proposed Decision

This ADR proposes the adoption of an "**Iterative API Schema-Driven Development**" process. This hybrid approach treats the OpenAPI schema as the single source of truth while incorporating agile feedback from implementation.

### 4.1. Development Process Policy

- Principle: For new API development, we will first design a draft schema and then establish a phase for verifying its technical feasibility through implementation. This verification will be done without waiting for the entire feature to be complete.
- Schema Review: If a schema change is required based on implementation challenges, a review will be conducted to validate its appropriateness. Team consensus will be formed through this review to finalize the schema.
- Detailed Documentation: The concrete procedures for this process (e.g., the scope of a Pull Request, review contents) will be detailed in a separate API Development Guideline, not in this ADR.

### 4.2. Migration Plan for Existing Codebase

For the **existing codebase**, we propose a three-step migration plan to transition to this new process:

1.  **Extract & Visualize:** Extract the current API behavior from the existing Go code into a draft openapi.yaml file.
2.  **Review & Agree:** Review this draft schema as a team to define and agree upon a consistent, high-quality "target schema."
3.  **Refactor & Test:** Refactor the existing code and add tests to ensure it correctly implements the agreed-upon target schema.

## 5. Why this decision was made

This iterative approach was chosen because it combines the benefits of both schema-first and code-first development while mitigating their respective drawbacks.

- Design Consistency: The schema remains the central "contract," ensuring API-wide consistency, which is a key benefit of the schema-first approach.
- Agility and Flexibility: It allows for rapid feedback from implementation to be incorporated into the design, avoiding the rigidity and potential for major rework found in a strict schema-first process.
- Enables Parallel Development: Client teams can begin development against the initial draft schema, without waiting for the final backend implementation.
- Reduces Technical Debt: It establishes a clear, maintainable, and scalable foundation for all future API development, resolving the issues outlined in the problem statement.

## 6. Why the Alternatives Were Not Recommended

- Option A (Current Process): Fails to address the core problems of scalability and maintainability, making it unsuitable for the project's long-term goals.
- Option B (Code-First): While familiar, it compromises the ability to enforce consistent API design upfront and creates dependencies that slow down parallel workstreams.
- Option C (Strict Schema-First): Poses a high risk of creating a design bottleneck and can lead to inefficient development cycles due to its rigidity. Our past experiences have shown this to be a potential issue.

## 7. Undecided Matters

- The detailed procedures for the process defined in 4.1.
- Tool Selection Finalization: Conduct PoCs to make a final decision on tools for linting and code generation.
- CI/CD Pipeline Detailed Implementation: Construction of the specific pipeline to automate schema linting and code generation validation.

These will be handled as separate tasks.

## 8. Considerations for Adoption

- OpenAPI Schema Maintainability:
  - Increase readability and maintainability by splitting the schema into multiple files by feature using $ref.
  - Enforce the DRY principle and maintain consistency by reusing common components (e.g., error responses).
- Initial Refactoring Effort: The 3-step transition plan for the existing codebase will require a dedicated development effort. This should be prioritized to establish a solid foundation before significant new feature development begins.
