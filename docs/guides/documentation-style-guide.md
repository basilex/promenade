# Documentation Style Guide

**Professional documentation standards** for Promenade Platform - clear, consistent, and emoji-free.

---

## Core Principles

1. **Professional Tone**: Business-appropriate language for enterprise software
2. **Clarity First**: Technical accuracy over creative expression
3. **Consistency**: Uniform formatting and terminology across all documents
4. **No Emoji**: Text-based symbols only (see below)

---

## Emoji Policy

### PROHIBITED: Emoji Usage

**DO NOT use emoji** in any project documentation, code comments, or commit messages.

**Rationale**:
- **Professionalism**: Enterprise software documentation should maintain business standards
- **Accessibility**: Screen readers may not properly announce emoji
- **Clarity**: Text is always clearer than symbols
- **Consistency**: Emoji rendering varies across platforms
- **Searchability**: Text is easier to search and grep than emoji

**Prohibited Examples**:
```markdown
  Production (checkmark emoji)
  Planned (clipboard emoji)
  Ready to launch (rocket emoji)
  Core Concepts (target emoji)
  Architecture (building emoji)
  Documentation (books emoji)
```

### ALLOWED: Text-Based Status Indicators

**Use text** to indicate status:

```markdown
 Production      (Plain text "Production")
 In Progress     (Plain text "In Progress")
 Planned Q3'26   (Plain text with quarter)
 Deprecated      (Plain text "Deprecated")
 Experimental    (Plain text "Experimental")
```

**Example - Bounded Contexts Table**:

```markdown
| Context      | Status        |
|--------------|---------------|
| Identity     | Production    |
| Billing      | Production    |
| Warehouse    | Planned Q3'26 |
```

**NOT**:
```markdown
| Context      | Status           |
|--------------|------------------|
| Identity     |  Production    |
| Billing      |  Production    |
| Warehouse    |  Planned Q3'26 |
```

---

## Text Formatting Standards

### Headers

**Use standard markdown headers**:

```markdown
# Main Title (H1) - One per document
## Major Section (H2)
### Subsection (H3)
#### Detail Section (H4)
```

**NOT**:
```markdown
#  Main Title
##  Architecture
###  Documentation
```

### Lists

**Use standard list markers**:

```markdown
- Bullet point
- Another point
  - Nested point

1. Numbered item
2. Another item
   1. Nested item
```

**Checkboxes for task lists**:
```markdown
- [x] Completed task
- [ ] Pending task
```

### Emphasis

**Use markdown emphasis**:

```markdown
**Bold** for important terms
*Italic* for emphasis
`code` for technical terms
```

**NOT**:
```markdown
 Hot feature
 Fast performance
 Pro tip
```

### Status Indicators

**Production-ready alternatives**:

| Concept              | Use This          | NOT This     |
| -------------------- | ----------------- | ------------ |
| Completed            | Production        |  or      |
| In development       | In Progress       |  or      |
| Planned              | Planned Q3'26     |  or      |
| Warning              | **Warning:**      |            |
| Error                | **Error:**        |            |
| Success              | **Success:**      |            |
| Info                 | **Note:**         | ℹ           |
| Documentation        | Guide, README     |  or      |
| Performance          | Optimized         |  or      |
| Security             | Secure            |  or      |

---

## Section Headers

### Standard Section Prefixes

**Use text prefixes** for section categories:

```markdown
## Quick Start
## Architecture Overview
## API Reference
## Configuration Options
## Testing Guide
## Troubleshooting
## Best Practices
## Related Documentation
```

**NOT**:
```markdown
##  Quick Start
##  Architecture Overview
##  API Reference
##  Configuration Options
##  Testing Guide
##  Troubleshooting
##  Best Practices
##  Related Documentation
```

---

## Code Comments

### Go Code Comments

**Professional comments**:

```go
// CreateCustomer creates a new customer with validation.
// Returns error if email already exists.
func CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*Customer, error) {
    // Implementation
}
```

**NOT**:
```go
//  CreateCustomer creates a new customer
//  Returns customer on success
//  Returns error if email exists
func CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*Customer, error) {
    // Implementation
}
```

### Commit Messages

**Standard git commit format**:

```bash
feat(api): Add Postman collection with environments

- Generate collection from OpenAPI spec
- Add 3 environment files (dev/staging/prod)
- Update documentation

Closes #123
```

**NOT**:
```bash
feat(api):  Add Postman collection 

-  Generate collection
-  Add environments
-  Update docs

 Closes #123
```

---

## Documentation File Names

**Use descriptive, lowercase, hyphenated names**:

```
api-documentation.md           Correct
testing-patterns.md            Correct
customer-management.md         Correct
rate-limiting.md               Correct
```

**NOT**:
```
API-Documentation.md           PascalCase
testing_patterns.md            snake_case
CustomerManagement.md          PascalCase
rate_limiting.md               snake_case
```

---

## README Structure

**Standard README template**:

```markdown
# Component Name

**Brief description** - what it does and why it matters.

---

## Quick Start

[Installation or usage instructions]

---

## Features

- Feature 1: Description
- Feature 2: Description
- Feature 3: Description

---

## Architecture

[Design explanation]

---

## API Reference

[Detailed API documentation]

---

## Configuration

[Configuration options]

---

## Testing

[How to run tests]

---

## Troubleshooting

[Common issues and solutions]

---

## Related Documentation

- [Link to guide 1](path/to/guide1.md)
- [Link to guide 2](path/to/guide2.md)

---

**Version**: 0.1.0  
**Status**: Production  
**Maintainer**: Promenade Team
```

---

## Status Terminology

### Standard Status Values

**Use these exact terms** for consistency:

| Status          | Meaning                                    | Use When                        |
| --------------- | ------------------------------------------ | ------------------------------- |
| **Production**  | Live, stable, fully tested                 | Feature complete, 90%+ coverage |
| **In Progress** | Active development                         | Currently being implemented     |
| **Planned**     | Scheduled for future development           | Roadmap item with quarter       |
| **Deprecated**  | No longer recommended, will be removed     | Legacy code, migration path set |
| **Experimental**| Testing, may change or be removed          | Alpha/beta features             |
| **Archived**    | No longer maintained                       | Obsolete code                   |

**Examples**:

```markdown
Status: Production (90+ tests, 95%+ coverage)
Status: In Progress (40% complete, expected Q1'26)
Status: Planned Q2'26 (awaiting Phase 3 completion)
Status: Deprecated (use NewAPI instead, removal Q4'26)
Status: Experimental (breaking changes possible)
Status: Archived (replaced by ModernComponent)
```

---

## Tables

### Clean Table Formatting

**Align columns** for readability:

```markdown
| Column 1        | Column 2      | Column 3    |
| --------------- | ------------- | ----------- |
| Value 1         | Value 2       | Value 3     |
| Longer value 1  | Short         | Medium val  |
```

**Use descriptive headers**:

```markdown
| Component       | Status     | Tests | Coverage |
| --------------- | ---------- | ----- | -------- |
| Event Bus       | Production | 67    | 100%     |
| JWT Auth        | Production | 27    | 87%      |
| Rate Limiting   | Production | 10    | 93%      |
```

---

## Links

### Internal Links

**Use relative paths**:

```markdown
See: [Testing Guide](../test/README.md)
Related: [API Documentation](docs/guides/api-documentation.md)
```

**NOT absolute paths**:
```markdown
See: [Testing Guide](/Users/user/project/test/README.md)
```

### External Links

**Use descriptive link text**:

```markdown
[OpenAPI Specification](https://swagger.io/specification/)
[Go Documentation](https://go.dev/doc/)
```

**NOT generic text**:
```markdown
[Click here](https://swagger.io/specification/)
[Link](https://go.dev/doc/)
```

---

## Code Blocks

### Language Specification

**Always specify language** for syntax highlighting:

```markdown
```go
func Example() {
    // Go code
}
```

```bash
make build
```

```yaml
config:
  value: example
```
```

### Command Examples

**Show command and expected output**:

```markdown
**Run tests**:
```bash
make test
```

**Output**:
```
ok  github.com/basilex/promenade/pkg/bus  0.123s
ok  github.com/basilex/promenade/pkg/jwt  0.456s
```
```

---

## Enforcement

### Automated Emoji Cleaner

**Professional automation tool** for emoji removal and policy enforcement.

**Tool**: `scripts/clean-emojies.py` (v2.0.0)

**Features**:
- Scans all documentation and code files (.md, .go, .yaml, .yml, .json, .toml)
- Three operating modes: dry-run (default), apply, check (CI)
- Detailed reporting with line numbers and emoji counts
- Safe by default (dry-run shows changes without modifying)
- CI integration for preventing new emoji commits

**Usage**:

```bash
# Dry-run: Show what would change (safe)
make clean-emoji
# OR
python3 scripts/clean-emojies.py

# Actually remove emoji from files
make clean-emoji-apply
# OR
python3 scripts/clean-emojies.py --apply

# CI mode: Exit 1 if emoji found (for pre-commit hooks)
make check-emoji
# OR
python3 scripts/clean-emojies.py --check

# Show help
python3 scripts/clean-emojies.py --help
```

**Integration with CI/CD**:

Add to `.github/workflows/ci.yml`:

```yaml
- name: Check for emoji violations
  run: make check-emoji
```

**Statistics** (as of January 5, 2026):
- Processed: 463 files
- Files with emoji detected: 106
- Total emoji found: 2,926
- Coverage: docs/, website/, pkg/, internal/, test/, code files

### Pre-Commit Checks

**Add to `.git/hooks/pre-commit`** (optional, manual approach):

```bash
#!/bin/bash

# Check for emoji in staged files using automated tool
if ! python3 scripts/clean-emojies.py --check; then
    echo ""
    echo "Error: Emoji detected in codebase"
    echo "Run 'make clean-emoji-apply' to fix automatically"
    echo "See: docs/guides/documentation-style-guide.md"
    exit 1
fi
```

**OR use simple pattern matching**:

```bash
#!/bin/bash

# Check for emoji in staged files
EMOJI_PATTERN="[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]"

if git diff --cached --name-only | xargs grep -P "$EMOJI_PATTERN" > /dev/null 2>&1; then
    echo "Error: Emoji detected in staged files"
    echo "Please remove emoji and use text instead"
    echo "See: docs/guides/documentation-style-guide.md"
    exit 1
fi
```

### Code Review Checklist

**Reviewers should verify**:

- [ ] No emoji in documentation
- [ ] No emoji in code comments
- [ ] No emoji in commit messages
- [ ] Status indicators use text only
- [ ] Headers follow standard format
- [ ] Tables are properly formatted
- [ ] Links use relative paths
- [ ] Code blocks specify language

---

## Examples

### BEFORE (with emoji)

```markdown
#  Core Features

##  Quick Start

 Production-ready  
 Planned features  
 High performance  
 Pro tip: Use caching

###  Architecture

-  Event Bus (377K events/sec)
-  JWT Authentication
-  API Versioning (planned)
```

### AFTER (professional)

```markdown
# Core Features

## Quick Start

**Status**: Production-ready  
**Planned**: API versioning (Q1'26)  
**Performance**: 377K events/sec  
**Note**: Enable caching for optimal performance

### Architecture

- Event Bus: Production (377K events/sec)
- JWT Authentication: Production
- API Versioning: Planned Q1'26
```

---

## Related Documentation

- [README.md](../../README.md) - Main project documentation
- [Documentation Index](../INDEX.md) - All documentation overview
- [Contributing Guide](../CONTRIBUTING.md) - Contribution guidelines
- [Testing Guide](../../test/README.md) - Testing standards

---

**Version**: 1.0.0  
**Effective Date**: January 5, 2026  
**Status**: Official Policy  
**Maintainer**: Promenade Team
