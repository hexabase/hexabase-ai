# ADR-010: Proposal to Migrate to a "Package by Feature" Structure

**Date**: 2025-06-18
**Status**: Proposed
**Authors**: Platform Architecture Team

## 1. Background

### Context

This project currently employs a `Package by Layer` structure, where packages are divided by technical layers such as `internal/handlers` and `internal/service`. This is a standard structure seen in many web applications.

### Problem Statement

As the project has scaled, several development challenges stemming from this structure have become apparent.

- **Scattered Concerns:** When modifying a single feature (e.g., Workspace Management), the related code is scattered across multiple directories like `handlers`, `service`, `domain`, and `repository`. This requires developers to open and edit multiple files simultaneously, increasing their cognitive load.
- **Poor Readability/Discoverability:** It is difficult to grasp what business features the application offers by simply looking at the directory structure.
- **Risk of High Coupling:** The boundaries between features within the same layer are ambiguous, creating a structure that is prone to tight coupling, such as when one feature's component inadvertently references another's.

In the long term, these issues can lead to decreased development productivity and increased maintenance costs.

## 2. Status

Proposed

## 3. Other options considered

### Option A: Maintain the current "Package by Layer" structure

Continue using the current package structure and manage any issues through operational discipline.

### Option B: Migrate to a "Package by Feature" structure

Reorganize packages based on functional (domain) units.

## 4. What was decided - Proposal

This document proposes the adoption of **Option B: Migrate to a "Package by Feature" structure**.

This is strictly a proposal for discussion, not a final decision. The approach aims to solve the current challenges by grouping related code by feature.

## 5. Why did you choose it? - Rationale for Proposal

We propose migrating to `Package by Feature` because we believe this approach is the most effective solution to the problems outlined in the "Background" section.

- **Improved Cohesion:** Grouping all code related to a feature into a single package clarifies its relevance and makes it easier to identify the scope of impact when making changes.
- **Improved Readability:** The application's feature set becomes intuitively understandable by looking at the top-level directories. This is also expected to reduce the onboarding cost for new team members.
- **Promotes Loose Coupling:** Clear package boundaries enhance the independence of features and help prevent unintended dependencies.
- **Flexibility for Future Architectural Changes:** Because each feature package is physically isolated, its independence is significantly increased. This makes it easier to extract specific features as microservices if the project's load increases in the future. For example, if the "Authentication" feature requires scaling, the `internal/auth` directory could be migrated to a new service repository with relatively low cost. This provides a strong technical foundation for gradually transitioning from a monolithic architecture.

We believe these benefits will improve the overall health of the codebase and enhance development and maintenance efficiency now and in the future.

## 6. Why didn't you choose the other option - Why the Alternative Was Not Recommended

Option A: Maintaining the status quo has the short-term benefit of avoiding migration costs. However, it does not address the fundamental problems.

If we keep the current structure, the codebase will continue to grow in complexity, and the issues outlined in the "Background" will only become more severe. Continuing with ad-hoc solutions will accumulate technical debt and risk even higher refactoring costs in the future. From a long-term perspective, we judge that addressing the architecture now, while the problems are understood, is the most cost-effective choice.

## 7. What has not been decided - Undecided Matters

This proposal requires discussion and consensus. If it is decided to adopt this proposal, the following points will need to be decided next:

- **Migration Strategy:**
  - **Big Bang or Phased Migration:** Will the entire project be migrated at once, or will the new structure be applied to new features first, with existing features migrated gradually as resources permit?

## 8. Considerations - Considerations for Adoption

The following points must be considered when proceeding with this proposal:

- **Refactoring Cost:** The work will involve extensive code movement and updates to import paths, which will require a significant amount of effort.
- **Risk of Regression:** Large-scale refactoring can introduce unintended behavior changes. Thorough testing will be critical during the migration process.
- **Team's Learning Curve:** Time will be needed for the entire team to understand the intent of the new structure and its rules.
- **Impact on Concurrent Development:** During the refactoring process, there is a higher risk of conflicts with other ongoing development tasks. A clear branching strategy and close communication will be essential.
