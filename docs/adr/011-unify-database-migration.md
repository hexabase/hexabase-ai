# ADR-011: Unification of Database Migration

- **Date**: 2025-06-20
- **Status**: Proposed
- **Authors**: Platform Architecture Team

## 1. Background

### Context

The Hexabase KaaS product utilizes a migration function to manage changes to the database schema. As the application grows and features are added, the database schema continuously evolves.

### Problem Statement

Currently, the project's database migration strategy is a **mix** of two different methods:

1.  **GORM's `AutoMigrate` feature**: Automatically alters the schema based on Go model (struct) definitions when the application starts. **These changes are not version-controlled.**
2.  **`golang-migrate/migrate`**: Explicitly alters the schema using version-controlled SQL files (up/down).

This mixed approach has become a significant **technical debt** that threatens the stability and maintainability of the system, causing the following problems:

-   **Lack of Reliability and Schema Inconsistency**: The two methods can interfere with each other, creating a risk of schema discrepancies between development and production environments.
-   **Increased Management Complexity**: The Single Source of Truth for the schema definition is scattered across Go models and SQL files, complicating management.
-   **Breakdown of Coherent Versioning**: While `golang-migrate` performs strict versioning, untraceable changes made by `AutoMigrate` occur concurrently, **resulting in a breakdown of coherent schema version control.**
-   **Impossibility of Safe Rollbacks**: `AutoMigrate` lacks a safe rollback feature, making recovery difficult in the event of a problem.
-   **Confusion for New Developers**: It is difficult to guarantee a reproducible database state when new members set up their development environments.
-   **Incomplete Migration via Scripts**: Some of the current migration scripts only apply `up.sql` files, meaning that bidirectional (Up/Down) versioning is not fully functional in all cases.

## 2. Status

Proposed

## 3. Other options considered

### Option A: Unify on GORM `AutoMigrate`

-   **Summary:** Abolish `golang-migrate/migrate` and perform all schema changes using only `AutoMigrate`.
-   **Evaluation:**
    -   **Pros:** Schema definitions are unified with Go model definitions in the code.
    -   **Cons:** Unsuitable for production use. It cannot handle destructive changes and lacks versioning and rollback capabilities. This would worsen the current problems and is therefore not a viable option.

### Option B: Immediately Adopt `Atlas`

-   **Summary:** Migrate to `Atlas`, a declarative schema management tool. It has the capability to read GORM model definitions, detect differences from the current DB schema, and automatically generate migration SQL.
-   **Evaluation:**
    -   **Pros:** Can automatically generate migration SQL from GORM models, improving the development experience.
    -   **Cons:**
        -   Incurs the learning cost of a new tool.
        -   It offers a free-to-use OSS version and a paid cloud version with advanced management features (CI integration, approval workflows, etc.). Utilizing all features incurs costs.

### Option C: Cover the Mixed State with Development Rules

-   **Summary:** Keep the mixed toolset but establish a development rule that "all future changes must use `golang-migrate`."
-   **Evaluation:**
    -   **Pros:** Reduces short-term refactoring effort.
    -   **Cons:**
        -   High risk of human error. The underlying technical debt remains.
        -   The management methods of `AutoMigrate` and `golang-migrate` are fundamentally different, making it highly probable that unintended schema changes will occur even with rules in place.

## 4. Proposed Decision

This ADR proposes to **fully unify the database migration strategy on the SQL file-based method using `golang-migrate/migrate`**.

To achieve this, the following transition steps are recommended:

1.  **Create a Baseline**: Extract the table schema definitions currently managed only by `AutoMigrate` and create a new migration file (an up/down SQL pair) in the `golang-migrate` format.
2.  **Verification**: Apply all existing and new baseline migrations to a clean database to ensure that the current database schema can be reproduced perfectly.
3.  **Deprecate `AutoMigrate`**: After the verification is complete, entirely remove the `AutoMigrate` call from the application code (`api/cmd/api/main.go`).

## 5. Why this decision was made

This decision is essential to fundamentally resolve the issues with the current migration infrastructure and to build a more robust and reliable system.

-   **Reliability and Reproducibility:** Explicitly describing changes in SQL files ensures the same result regardless of who runs the migration or when.
-   **Strict Version Control:** All schema change history is recorded as version-controlled code, allowing for safe and easy migration to specific versions or rollbacks.
-   **Suitability for Production Operations:** It can handle complex data migrations and destructive changes, enabling the robust operations required in a production environment.
-   **Securing Future Options:** `go-migrate` is an industry standard and does not preclude a future migration to more advanced tools like `Atlas`.
-   **Fundamental Problem Resolution:** It fundamentally resolves all issues stemming from the mixed approach (inconsistency, management complexity, lack of traceability, etc.).

## 6. Why the Alternatives Were Not Recommended

-   **Option A (`AutoMigrate`)**: It is unsuitable for a product intended for production use due to its lack of robustness, version control, and rollback features, which would compromise the product's reliability.
-   **Option B (`Atlas`)**: While an excellent tool, we judged that our priority should be to resolve the current technical debt with a simple and reliable method first. Its future adoption remains a possibility.
-   **Option C (Mixed State + Rules)**: This only postpones the problem and is unacceptable as it would mean continuing to operate with the risk of accidents caused by human error.

## 7. Undecided Matters

None. The policy is proposed in this ADR. Specific tasks will be managed via GitHub Issues.

## 8. Considerations for Adoption

-   **Implementation Cost:** Development effort will be required to create and verify the baseline SQL files.
-   **Regression Risk:** There is a risk of inconsistencies with existing migrations during the transition. Testing migrations from a clean database is mandatory.
-   **Impact on Parallel Development:** During this refactoring, close communication and a clear branching strategy are essential to avoid conflicts with other development tasks involving schema changes.
-   **Team-wide Communication:** After the transition, the new development rule—that all schema changes must be made by creating a new SQL file with `go-migrate`—must be thoroughly communicated to the entire team.