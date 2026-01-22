# ADR-NNNN: [Title - Short Present Tense Phrase]

**Status**: [Proposed | Accepted | Rejected | Superseded | Deprecated]  
**Date**: YYYY-MM-DD  
**Deciders**: [Names or roles of decision makers]  
**Tags**: [Optional: e.g., database, performance, security]

---

## Context

**What is the issue we're seeing that is motivating this decision or change?**

Describe the forces at play:

- Technical constraints
- Business requirements
- Performance considerations
- Team skills
- Cost factors
- Time pressure

Example:

> We need globally unique identifiers for distributed entities (Customer, Order, Invoice).
> Current UUID v4 approach creates random IDs, leading to poor database index performance
> and difficult debugging (no timestamp information).

---

## Considered Options

**What alternatives did we consider?**

### Option 1: [Name]

- **Pros**: ...
- **Cons**: ...
- **Example**: ...

### Option 2: [Name]

- **Pros**: ...
- **Cons**: ...
- **Example**: ...

### Option 3: [Name]

- **Pros**: ...
- **Cons**: ...
- **Example**: ...

---

## Decision

**What is the change that we're proposing and/or doing?**

State the decision clearly:

> We will use **UUID v7 (time-ordered)** for all primary keys instead of UUID v4 (random).

Provide implementation details:

- How will this be implemented?
- What components are affected?
- What is the migration path (if applicable)?

Example:

```go
import "github.com/basilex/promenade/pkg/uuidv7"

func NewCustomer() *Customer {
    return &Customer{
        ID: uuidv7.New(),  // Time-sortable UUID
    }
}
```

---

## Consequences

**What becomes easier or more difficult to do because of this change?**

### Positive

-  Benefit 1
-  Benefit 2
-  Benefit 3

### Negative

-  Drawback 1
-  Drawback 2

### Neutral

- ℹ Consideration 1
- ℹ Consideration 2

---

## Implementation Notes

**How was/will this be rolled out?**

- **Phase 1**: ...
- **Phase 2**: ...
- **Rollback Plan**: ...

Example:

> All new entities use `uuidv7.New()`. Existing entities remain on UUID v4 (no migration needed).

---

## References

- [Link to RFC, blog post, or documentation]
- [Related ADRs]
- [Related code files]

Example:

- [RFC 9562 - UUID v7 Specification](https://www.rfc-editor.org/rfc/rfc9562.html)
- [pkg/uuidv7/README.md](../../pkg/uuidv7/README.md)

---

## Revisit Criteria

**Under what conditions should this decision be revisited?**

Example:

- If database write performance degrades by >20%
- If UUID v8 becomes widely adopted (better collision resistance)
- If distributed ID generation becomes a bottleneck

---

## Changelog

| Date       | Change                              | Author |
| ---------- | ----------------------------------- | ------ |
| YYYY-MM-DD | Initial version                     | Name   |
| YYYY-MM-DD | Updated consequences after 6 months | Name   |
