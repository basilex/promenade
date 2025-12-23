# Promenade Documentation Index

This directory contains comprehensive documentation for the Promenade application architecture, development workflows, and best practices.

---

## Start Here

### New to Promenade?

1. **[ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md)** - Visual architecture diagram and component overview
2. **[ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md)** - Quick reference guide for developers
3. **[README.md](../README.md)** - Main project README

### Architecture Review

- **[ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md)** - Complete architecture compliance audit (Core vs Modules)

---

## Core Concepts

### Architecture & Design

| Document                                             | Description                                | When to Read                        |
| ---------------------------------------------------- | ------------------------------------------ | ----------------------------------- |
| [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md) | Complete visual architecture with diagrams | Understanding system structure      |
| [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md)       | Architecture compliance report             | Verifying design principles         |
| [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md) | Quick reference for common patterns        | Daily development                   |
| [../internal/CORE.md](../internal/CORE.md)           | Core components documentation              | Understanding core responsibilities |

### Modules

| Document                                                       | Description                          | When to Read                    |
| -------------------------------------------------------------- | ------------------------------------ | ------------------------------- |
| [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md)                 | Complete guide to creating modules   | Building new modules            |
| [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.md)               | Module independence principles       | Understanding module boundaries |
| [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.md) | Configuration management for modules | Setting up module configs       |
| [../internal/modules/README.md](../internal/modules/README.md) | Module directory structure           | Quick module overview           |

---

## Technical Guides

### Database & Persistence

| Document                             | Description                      | When to Read              |
| ------------------------------------ | -------------------------------- | ------------------------- |
| [UUID_V7_GUIDE.md](UUID_V7_GUIDE.md) | Using time-ordered UUIDs         | Working with primary keys |
| [SOFT_DELETE.md](SOFT_DELETE.md)     | Soft delete patterns and gotchas | Implementing soft delete  |

### Infrastructure

| Document                                       | Description                     | When to Read                    |
| ---------------------------------------------- | ------------------------------- | ------------------------------- |
| [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.md) | Automated purge system design   | Implementing retention policies |
| [REDIS_BUS_TESTING.md](REDIS_BUS_TESTING.md)   | Testing with Redis event bus    | Testing event-driven features   |
| [LOGGING.md](LOGGING.md)                       | Structured logging with context | Adding logging to code          |

### Security & Licensing

| Document                                           | Description                              | When to Read                    |
| -------------------------------------------------- | ---------------------------------------- | ------------------------------- |
| [AUTH_SCHEMA.md](AUTH_SCHEMA.md)                   | Authentication schema and flows          | Understanding auth              |
| [AUTHORIZATION.md](AUTHORIZATION.md)               | RBAC permission system                   | Implementing authorization      |
| [CREDENTIALS.md](CREDENTIALS.md)                   | Credential management                    | Handling user credentials       |
| [LICENSE_ARCHITECTURE.md](LICENSE_ARCHITECTURE.md) | Module licensing system with HMAC-SHA256 | Implementing commercial modules |

---

## Testing

| Document                                               | Description               | When to Read                    |
| ------------------------------------------------------ | ------------------------- | ------------------------------- |
| [TESTING_GUIDE.md](TESTING_GUIDE.md)                   | Complete testing strategy | Writing tests                   |
| [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.md) | Test infrastructure setup | Setting up test environment     |
| [../test/README.md](../test/README.md)                 | Test directory structure  | Understanding test organization |

---

## Development Workflows

### Build & Deploy

| Document                                             | Description                   | When to Read               |
| ---------------------------------------------------- | ----------------------------- | -------------------------- |
| [MAKEFILE_ARCHITECTURE.md](MAKEFILE_ARCHITECTURE.md) | Makefile system documentation | Using make commands        |
| [../docker/README.md](../docker/README.md)           | Docker setup and deployment   | Containerizing application |

### Validation & Quality

| Document                       | Description               | When to Read            |
| ------------------------------ | ------------------------- | ----------------------- |
| [VALIDATION.md](VALIDATION.md) | Input validation patterns | Adding validation rules |

---

## API Documentation

### V1 API

- [v1/v1_docs.go](v1/v1_docs.go) - V1 API documentation
- [v1/v1_swagger.yaml](v1/v1_swagger.yaml) - V1 Swagger spec (YAML)
- [v1/v1_swagger.json](v1/v1_swagger.json) - V1 Swagger spec (JSON)

### V2 API

- [v2/v2_docs.go](v2/v2_docs.go) - V2 API documentation
- [v2/v2_swagger.yaml](v2/v2_swagger.yaml) - V2 Swagger spec (YAML)
- [v2/v2_swagger.json](v2/v2_swagger.json) - V2 Swagger spec (JSON)

---

## By Topic

### Core Architecture

- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md) - Visual overview
- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md) - Compliance review
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md) - Quick reference
- [../internal/CORE.md](../internal/CORE.md) - Core components

### Module System

- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md) - Development guide
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.md) - Independence principles
- [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.md) - Configuration
- [../internal/modules/README.md](../internal/modules/README.md) - Module index

### Data Management

- [UUID_V7_GUIDE.md](UUID_V7_GUIDE.md) - Primary keys
- [SOFT_DELETE.md](SOFT_DELETE.md) - Soft delete patterns
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.md) - Automated purge

### Security & Auth

- [AUTH_SCHEMA.md](AUTH_SCHEMA.md) - Authentication
- [AUTHORIZATION.md](AUTHORIZATION.md) - RBAC system
- [CREDENTIALS.md](CREDENTIALS.md) - Credential handling

### Testing & Quality

- [TESTING_GUIDE.md](TESTING_GUIDE.md) - Testing strategy
- [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.md) - Test setup
- [VALIDATION.md](VALIDATION.md) - Input validation
- [REDIS_BUS_TESTING.md](REDIS_BUS_TESTING.md) - Event bus testing

### Infrastructure

- [LOGGING.md](LOGGING.md) - Structured logging
- [MAKEFILE_ARCHITECTURE.md](MAKEFILE_ARCHITECTURE.md) - Build system
- [../docker/README.md](../docker/README.md) - Docker setup

---

## Learning Paths

### Path 1: Understanding the System (New Developer)

1. [README.md](../README.md) - Project overview
2. [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md) - System architecture
3. [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md) - Common patterns
4. [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md) - Building features
5. [TESTING_GUIDE.md](TESTING_GUIDE.md) - Testing your code

### Path 2: Building a New Module

1. [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md) - Module creation guide
2. [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.md) - Design principles
3. [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.md) - Configuration
4. [UUID_V7_GUIDE.md](UUID_V7_GUIDE.md) - Primary keys
5. [SOFT_DELETE.md](SOFT_DELETE.md) - If using soft delete
6. [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.md) - If implementing purge

### Path 3: Architecture Review (Technical Lead)

1. [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md) - Current state analysis
2. [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md) - Visual diagrams
3. [../internal/CORE.md](../internal/CORE.md) - Core boundaries
4. [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.md) - Module isolation

### Path 4: Security Implementation

1. [AUTH_SCHEMA.md](AUTH_SCHEMA.md) - Authentication flows
2. [AUTHORIZATION.md](AUTHORIZATION.md) - RBAC permissions
3. [CREDENTIALS.md](CREDENTIALS.md) - Credential security

### Path 5: Testing & Quality

1. [TESTING_GUIDE.md](TESTING_GUIDE.md) - Testing strategy
2. [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.md) - Test setup
3. [VALIDATION.md](VALIDATION.md) - Input validation
4. [REDIS_BUS_TESTING.md](REDIS_BUS_TESTING.md) - Event testing

---

## Quick Lookups

### How do I...

**...create a new module?**
→ [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md)

**...add configuration to my module?**
→ [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.md)

**...implement soft delete?**
→ [SOFT_DELETE.md](SOFT_DELETE.md)

**...add retention policies?**
→ [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.md)

**...use UUIDs correctly?**
→ [UUID_V7_GUIDE.md](UUID_V7_GUIDE.md)

**...implement authentication?**
→ [AUTH_SCHEMA.md](AUTH_SCHEMA.md)

**...add RBAC permissions?**
→ [AUTHORIZATION.md](AUTHORIZATION.md)

**...write tests?**
→ [TESTING_GUIDE.md](TESTING_GUIDE.md)

**...add logging?**
→ [LOGGING.md](LOGGING.md)

**...validate input?**
→ [VALIDATION.md](VALIDATION.md)

**...understand the architecture?**
→ [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md)

**...check if my code follows principles?**
→ [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md)

**...get a quick reference?**
→ [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md)

---

## Recent Updates

### December 22, 2025

- Created comprehensive architecture documentation:
- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md) - Complete compliance review
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md) - Visual diagrams
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md) - Quick reference
- Updated purge system documentation:
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.md) - Registry-based approach
- Verified module independence (15,000+ lines removed from core)

---

## 🤝 Contributing

When adding new documentation:

1. **Choose the right type:**

   - `ARCHITECTURE_*.md` - Architecture and design patterns
   - `MODULE_*.md` - Module system documentation
   - `*_GUIDE.md` - How-to guides and tutorials
   - `*_SCHEMA.md` - Data schemas and structures
   - `README.md` - Directory overviews

2. **Update this index:**

   - Add to relevant section
   - Update "Recent Updates"
   - Add to "Quick Lookups" if applicable

3. **Cross-reference:**

   - Link to related documents
   - Update related docs with links back

4. **Keep it current:**
   - Update when architecture changes
   - Archive outdated docs with `DEPRECATED_` prefix

---

## 📞 Support

- **Questions about architecture?** → Read [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md)
- **Questions about modules?** → Read [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md)
- **Questions about testing?** → Read [TESTING_GUIDE.md](TESTING_GUIDE.md)
- **Other questions?** → Check this index or main [README.md](../README.md)

---

**Documentation Version:** 2.0  
**Last Updated:** December 22, 2025  
**Maintained by:** Promenade Development Team
