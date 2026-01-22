# Architecture Decision Records (ADR)

This directory contains **Architecture Decision Records** - documentation of significant architectural decisions made during Promenade development.

---

## What is an ADR?

An **Architecture Decision Record (ADR)** is a document that captures an important architectural decision made along with its context and consequences.

**Purpose**:

- Document the **"why"** behind architectural choices
- Preserve context for future developers
- Enable informed decision-making (avoid repeating past mistakes)
- Track evolution of architectural thinking

---

## Format

Each ADR follows this structure:

```markdown
# ADR-NNNN: Title

**Status**: [Accepted | Rejected | Superseded | Deprecated]
**Date**: YYYY-MM-DD
**Deciders**: Name(s)

## Context

What is the issue we're seeing that is motivating this decision or change?

## Decision

What is the change that we're proposing and/or doing?

## Consequences

What becomes easier or more difficult to do because of this change?

### Positive

- ...

### Negative

- ...

### Neutral

- ...
```

---

## Naming Convention

ADRs are numbered sequentially:

```
adr-0001-use-uuid-v7.md
adr-0002-use-sqlx-over-gorm.md
adr-0003-migrate-to-pgx-stdlib.md
```

---

## Current ADRs

| ADR                                              | Title                                    | Status   | Date       |
| ------------------------------------------------ | ---------------------------------------- | -------- | ---------- |
| [0001](adr-0001-use-uuid-v7.md)                  | Use UUID v7 for Primary Keys             | Accepted | 2026-01-15 |
| [0002](adr-0002-use-sqlx-over-gorm.md)           | Use sqlx Instead of GORM                 | Accepted | 2026-01-15 |
| [0003](adr-0003-migrate-to-pgx-stdlib.md)        | Migrate from lib/pq to pgx/v5/stdlib     | Accepted | 2026-01-22 |
| [0004](adr-0003-use-redis-for-caching.md)        | Use Redis for Caching Layer              | Accepted | 2026-01-15 |
| [0005](adr-0004-event-bus-architecture.md)       | In-Memory Event Bus for Bounded Contexts | Accepted | 2026-01-18 |
| [0006](adr-0005-saga-pattern-for-fulfillment.md) | Saga Pattern for Order Fulfillment       | Accepted | 2026-01-20 |

---

## How to Create a New ADR

1. **Find next number**:

   ```bash
   ls docs/adr/*.md | tail -1  # Check last ADR number
   ```

2. **Copy template**:

   ```bash
   cp docs/adr/adr-template.md docs/adr/adr-NNNN-title.md
   ```

3. **Fill in details**:
   - Context: Why this decision is needed
   - Decision: What was decided
   - Consequences: Positive, negative, neutral impacts

4. **Update README**: Add entry to table above

5. **Commit**:
   ```bash
   git add docs/adr/adr-NNNN-*.md
   git commit -m "docs: add ADR-NNNN for [decision]"
   ```

---

## References

- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) by Michael Nygard
- [ADR GitHub Organization](https://adr.github.io/)
- [When Should I Write an Architecture Decision Record](https://engineering.atspotify.com/2020/04/when-should-i-write-an-architecture-decision-record/)

---

## Related Documentation

- [docs/guides/](../guides/) - Technical guides
- [docs/concepts/](../concepts/) - Architectural concepts
- [CHANGELOG.md](../../CHANGELOG.md) - Project changelog
