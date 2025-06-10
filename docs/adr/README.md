# Architecture Decision Records (ADRs)

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences. ADRs help us:

- Document why decisions were made
- Track the evolution of our architecture
- Onboard new team members effectively
- Avoid repeating past mistakes
- Understand the trade-offs we've accepted

## ADR Format

We use the following format for our ADRs:

```markdown
# Title

**Date**: YYYY-MM-DD HH:MI
**Status**: Proposed | Accepted | Deprecated | Superseded  
**Deciders**: List of people involved  
**Tags**: architecture, security, performance, etc.

## Context

What is the issue that we're seeing that is motivating this decision or change?

## Decision

What is the change that we're proposing and/or doing?

## Consequences

### Positive

- What becomes easier or more possible?

### Negative

- What becomes harder or less possible?

### Risks

- What risks are we accepting?

## Alternatives Considered

What other options did we consider? Why didn't we choose them?

## References

Links to related documents, discussions, or external resources.
```

## ADR Lifecycle

1. **Proposed**: The ADR is a proposal under discussion
2. **Accepted**: The decision has been made and is being implemented
3. **Deprecated**: The decision is no longer relevant or has been reversed
4. **Superseded**: The decision has been replaced by a newer ADR

## Current ADRs

### 🟢 Active Decisions

- [2024-12-16 Logging Strategy](./records/2024-12-16_logging-strategy.md) - ClickHouse for centralized logging
- [2024-12-17 Package Management](./records/2024-12-17_package-management.md) - Node.js package management architecture
- [2025-06-09 AI-Ops Architecture](./records/2025-06-09_aiops-architecture.md) - Integration with LLMs and AI agents

### 📊 Implementation Status

For current implementation progress and work status, see:

- [WORK-STATUS.md](./WORK-STATUS.md) - Live tracking of current sprint and priorities

## How to Create a New ADR

1. Copy the template from above
2. Create a new file in `records/` with format: `YYYY-MM-DD_HH_MM_short-title.md`
3. Fill out all sections
4. Submit PR for review
5. Update this README with the new ADR

## Rules and Guidelines

1. **One Decision Per ADR**: Each ADR should document a single architectural decision
2. **Immutable Once Accepted**: Don't edit accepted ADRs; create new ones to supersede
3. **Include Context**: Always explain why the decision was needed
4. **Document Trade-offs**: Be honest about what we're giving up
5. **Link Related ADRs**: Reference previous decisions that this relates to
6. **Keep It Concise**: Aim for 1-2 pages maximum

## Search ADRs

To find relevant ADRs:

```bash
# Search by content
grep -r "pattern" docs/adr/records/

# List by date
ls -la docs/adr/records/ | sort

# Find by status
grep -l "Status: Deprecated" docs/adr/records/*.md
```

## Questions?

If you have questions about ADRs or need help creating one, please:

- Check existing ADRs for examples
- Ask in the #architecture Slack channel
- Create a GitHub discussion

---

_The ADR process helps us make better decisions by forcing us to think through the implications and document our reasoning._
