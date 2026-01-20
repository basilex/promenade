# Context Structure Refactoring Checklist

**Navigation**: [Home](../../README.md) > [Docs](../INDEX.md) > [Refactoring](README.md) > Structure Refactoring

---

## Overview

This document tracks the migration of all contexts to the **standardized structure** defined in [Context & Aggregate Naming Conventions](../guides/context-aggregate-naming-conventions.md).

**Goal**: Unified structure across ALL bounded contexts by end of Q1 2026.

---

## Refactoring Strategy

### Phase 1: New Contexts (Start Here)

-  **banking** - Already follows new standard (completed Jan 17, 2026)
- All new contexts MUST follow standard from day 1

### Phase 2: Active Development Contexts

- High-priority contexts with ongoing development
- Refactor during feature work (opportunistic)

### Phase 3: Stable Contexts

- Low churn, production-ready contexts
- Dedicated refactoring sprint

### Phase 4: Legacy/Shared Contexts

- Minimal changes unless needed
- Low priority

---

## Context Inventory & Status

###  Compliant Contexts

| Context     | Status       | Notes                          | Last Updated |
| ----------- | ------------ | ------------------------------ | ------------ |
| **banking** |  Compliant | New context following standard | Jan 17, 2026 |

###  In Progress

| Context | Status | Priority | Assignee | Notes |
| ------- | ------ | -------- | -------- | ----- |
| -       | -      | -        | -        | -     |

###  Needs Refactoring (High Priority)

| Context                       | Current State   | Issues                                                                                    | Priority  | Effort |
| ----------------------------- | --------------- | ----------------------------------------------------------------------------------------- | --------- | ------ |
| **customer-mgmt/analytics**   | Mixed structure | - Files in root + subdirs<br>- Inconsistent naming<br>- `usecase.go` + `usecase/` coexist |  High   | Medium |
| **customer-mgmt/company**     | Flat structure  | - `entity.go` in root<br>- No `/aggregate/` folder<br>- No separate use case files        |  High   | Medium |
| **customer-mgmt/deal**        | Flat structure  | - Same as company<br>- Mixed with adapter                                                 |  High   | Medium |
| **customer-mgmt/customer**    | Flat structure  | - Same pattern as company                                                                 |  Medium | Medium |
| **customer-mgmt/interaction** | Flat structure  | - Same pattern                                                                            |  Medium | Medium |

###  Lower Priority (Stable)

| Context                     | Current State              | Priority  | Effort         |
| --------------------------- | -------------------------- | --------- | -------------- |
| **identity/user**           | Flat structure             |  Medium | Low            |
| **identity/profile**        | Flat structure             |  Medium | Low            |
| **identity/contact**        | Flat structure             |  Medium | Low            |
| **identity/role**           | Flat structure             |  Medium | Low            |
| **identity/permission**     | Flat structure             |  Medium | Low            |
| **order-mgmt/order**        | Flat structure             |  Medium | Medium         |
| **order-mgmt/contract**     | Flat structure             |  Medium | Low            |
| **order-mgmt/fulfillment**  | Saga pattern               |  Low    | High (complex) |
| **billing/invoice**         | Flat structure             |  Medium | Low            |
| **billing/payment**         | Flat structure             |  Medium | Low            |
| **billing/subscription**    | Flat structure             |  Medium | Low            |
| **warehouse/product**       | Flat structure             |  Medium | Low            |
| **warehouse/inventory**     | Flat structure             |  Medium | Medium         |
| **warehouse/location**      | Flat structure             |  Medium | Low            |
| **warehouse/stockmovement** | Flat structure             |  Medium | Low            |
| **fiscal/receipt**          | Flat structure + providers |  Medium | Medium         |
| **fiscal/cashregister**     | Flat structure             |  Medium | Low            |
| **scripting/script**        | Flat structure             |  Low    | Low            |

###  Special Cases (Keep As-Is)

| Context              | Status       | Reason                   |
| -------------------- | ------------ | ------------------------ |
| **shared/country**   | Keep current | Reference data only      |
| **shared/currency**  | Keep current | Reference data only      |
| **shared/language**  | Keep current | Reference data only      |
| **shared/timezone**  | Keep current | Reference data only      |
| **ui/metadata/form** | Keep current | UI metadata (not domain) |

---

## Detailed Refactoring Plans

### 1. customer-mgmt/analytics ( High Priority)

**Current Structure**:

```
/analytics/
  read_model.go
  usecase.go                           # Old monolithic
  usecase_test.go
  README.md
  dto/sales_report_dto.go
  handler/sales_report_handler.go
  usecase/sales_report_usecase.go      # New separated
  integration/sales_report_event_handler.go
  adapter/http/handler.go
  adapter/http/dto.go
  adapter/repository/postgres/...
```

**Target Structure**:

```
/analytics/
  README.md
  errors.go

  /read_model/                         # Special: CQRS read models
    sales_report.go
    customer_stats.go

  /repository/
    sales_report_repository.go         # Interface

  /usecase/
    get_sales_report.go                # Query use case
    get_customer_stats.go

  /dto/
    sales_report_dto.go
    customer_stats_dto.go

  /adapter/
    /repository/postgres/
      base_repository.go
      sales_report_repository.go
    /http/
      sales_report_handler.go
      customer_stats_handler.go

  /integration/
    order_completed_handler.go         # Updates read model
    payment_received_handler.go
```

**Migration Steps**:

1. [ ] Create `/read_model/` folder
2. [ ] Move `read_model.go` → `/read_model/sales_report.go`
3. [ ] Create `/repository/sales_report_repository.go` interface
4. [ ] Move `/usecase/sales_report_usecase.go` → `/usecase/get_sales_report.go`
5. [ ] Delete old `usecase.go` (monolithic)
6. [ ] Move `/handler/sales_report_handler.go` → `/adapter/http/sales_report_handler.go`
7. [ ] Delete old `adapter/http/handler.go` and `adapter/http/dto.go`
8. [ ] Move `/integration/sales_report_event_handler.go` → `/integration/order_completed_handler.go`
9. [ ] Update imports in all files
10. [ ] Update bootstrap.go
11. [ ] Run tests
12. [ ] Delete old empty directories

**Breaking Changes**: None (internal only)

---

### 2. customer-mgmt/company ( High Priority)

**Current Structure**:

```
/company/
  entity.go
  entity_test.go
  usecase.go
  usecase_test.go
  repository.go                        # Interface
  errors.go
  adapter/http/handler.go
  adapter/http/dto.go
  adapter/repository/postgres/base_repository.go
  adapter/repository/postgres/company_repository.go
```

**Target Structure**:

```
/company/
  README.md
  errors.go

  /aggregate/
    company.go
    company_test.go

  /repository/
    company_repository.go              # Interface only

  /usecase/
    create_company.go
    update_company.go
    get_company.go
    list_companies.go
    delete_company.go
    # Split monolithic usecase.go into separate files

  /dto/
    company_dto.go

  /adapter/
    /repository/postgres/
      base_repository.go
      company_repository.go            # Implementation
    /http/
      company_handler.go
```

**Migration Steps**:

1. [ ] Create `/aggregate/` folder
2. [ ] Move `entity.go` → `/aggregate/company.go`
3. [ ] Move `entity_test.go` → `/aggregate/company_test.go`
4. [ ] Create `/repository/` folder
5. [ ] Move `repository.go` → `/repository/company_repository.go`
6. [ ] Create `/usecase/` folder
7. [ ] Analyze `usecase.go` and split into:
   - `create_company.go` (CreateCompanyUseCase)
   - `update_company.go` (UpdateCompanyUseCase)
   - `get_company.go` (GetCompanyUseCase)
   - `list_companies.go` (ListCompaniesUseCase)
   - `delete_company.go` (DeleteCompanyUseCase)
8. [ ] Split `usecase_test.go` into separate test files
9. [ ] Create `/dto/` folder
10. [ ] Move `adapter/http/dto.go` → `/dto/company_dto.go`
11. [ ] Move `adapter/http/handler.go` → `/adapter/http/company_handler.go`
12. [ ] Update imports in all files
13. [ ] Update bootstrap.go (register all use cases)
14. [ ] Run tests
15. [ ] Delete old files

**Breaking Changes**: None (internal only)

---

### 3. customer-mgmt/deal ( High Priority)

**Same pattern as company** - follow identical migration steps.

---

### 4. customer-mgmt/customer ( Medium Priority)

**Same pattern as company** - follow identical migration steps.

---

### 5. identity/user ( Medium Priority)

**Current Structure**: Flat (same as company)

**Migration**: Follow company pattern (create aggregate/, repository/, usecase/, dto/)

---

## Generic Migration Template

For all **flat structure** contexts (company, deal, customer, user, etc.):

```bash
# 1. Create new structure
mkdir -p {context}/aggregate
mkdir -p {context}/repository
mkdir -p {context}/usecase
mkdir -p {context}/dto

# 2. Move files
mv {context}/entity.go {context}/aggregate/{name}.go
mv {context}/entity_test.go {context}/aggregate/{name}_test.go
mv {context}/repository.go {context}/repository/{name}_repository.go
mv {context}/adapter/http/dto.go {context}/dto/{name}_dto.go
mv {context}/adapter/http/handler.go {context}/adapter/http/{name}_handler.go

# 3. Split monolithic usecase.go
# (manual - analyze methods and create separate files)

# 4. Update imports
# (manual or script)

# 5. Update bootstrap.go
# (manual - register all new use cases)

# 6. Test
make test-unit
make test-smoke

# 7. Clean up
rm {context}/usecase.go
rm {context}/usecase_test.go
```

---

## Testing Strategy

After each refactoring:

1. **Unit Tests**: `make test-unit`
2. **Smoke Tests**: `make test-smoke`
3. **Integration Tests**: `make test-integration`
4. **Compilation**: `make build`

**Rule**: All tests must pass before committing.

---

## Import Update Script

Create script to update imports automatically:

```bash
#!/bin/bash
# scripts/update-context-imports.sh

CONTEXT=$1
OLD_PATH="internal/contexts/${CONTEXT}"

# Update aggregate imports
find . -type f -name "*.go" -exec sed -i '' \
  "s|${OLD_PATH}/entity|${OLD_PATH}/aggregate|g" {} \;

# Update repository imports
find . -type f -name "*.go" -exec sed -i '' \
  "s|${OLD_PATH}/repository\.go|${OLD_PATH}/repository|g" {} \;

# Update usecase imports
find . -type f -name "*.go" -exec sed -i '' \
  "s|${OLD_PATH}/usecase\.go|${OLD_PATH}/usecase|g" {} \;

# Update dto imports
find . -type f -name "*.go" -exec sed -i '' \
  "s|${OLD_PATH}/adapter/http|${OLD_PATH}/dto|g" {} \;

echo " Updated imports for ${CONTEXT}"
```

---

## Rollout Schedule

### Week 1-2 (Jan 20 - Jan 31, 2026)

- [ ] Refactor **customer-mgmt/analytics** (high priority, active development)
- [ ] Refactor **customer-mgmt/company** (foundation for others)

### Week 3-4 (Feb 3 - Feb 14, 2026)

- [ ] Refactor **customer-mgmt/deal** (similar to company)
- [ ] Refactor **customer-mgmt/customer** (similar to company)
- [ ] Refactor **customer-mgmt/interaction**

### Week 5-6 (Feb 17 - Feb 28, 2026)

- [ ] Refactor **identity/user**
- [ ] Refactor **identity/profile**
- [ ] Refactor **identity/contact**
- [ ] Refactor **identity/role**
- [ ] Refactor **identity/permission**

### Week 7-8 (Mar 3 - Mar 14, 2026)

- [ ] Refactor **order-mgmt/order**
- [ ] Refactor **order-mgmt/contract**
- [ ] Refactor **billing/invoice**
- [ ] Refactor **billing/payment**
- [ ] Refactor **billing/subscription**

### Week 9-10 (Mar 17 - Mar 28, 2026)

- [ ] Refactor **warehouse/product**
- [ ] Refactor **warehouse/inventory**
- [ ] Refactor **warehouse/location**
- [ ] Refactor **warehouse/stockmovement**

### Week 11-12 (Mar 31 - Apr 11, 2026)

- [ ] Refactor **fiscal/receipt**
- [ ] Refactor **fiscal/cashregister**
- [ ] Refactor **scripting/script**
- [ ] Final cleanup & documentation

---

## Bootstrap.go Changes

Each refactoring requires updating `cmd/api/bootstrap.go`:

**Before (company example)**:

```go
// Company
companyRepo := company_postgres.NewCompanyRepository(db)
companyUseCase := company.NewUseCase(companyRepo, eventBus)
companyHandler := company_http.NewHandler(companyUseCase)
```

**After**:

```go
// Company
companyRepo := postgres.NewCompanyRepository(db)

// Use cases
createCompanyUC := usecase.NewCreateCompanyUseCase(companyRepo, eventBus)
updateCompanyUC := usecase.NewUpdateCompanyUseCase(companyRepo, eventBus)
getCompanyUC := usecase.NewGetCompanyUseCase(companyRepo)
listCompaniesUC := usecase.NewListCompaniesUseCase(companyRepo)
deleteCompanyUC := usecase.NewDeleteCompanyUseCase(companyRepo, eventBus)

// Handler (inject all use cases)
companyHandler := http.NewCompanyHandler(
    createCompanyUC,
    updateCompanyUC,
    getCompanyUC,
    listCompaniesUC,
    deleteCompanyUC,
)
```

---

## Migration Validation Checklist

For each refactored context:

- [ ] All files follow snake_case naming
- [ ] All structs follow PascalCase naming
- [ ] `/aggregate/` contains domain entities
- [ ] `/repository/` contains interfaces only
- [ ] `/usecase/` contains separate use case files
- [ ] `/dto/` contains request/response DTOs
- [ ] `/adapter/repository/postgres/` contains implementations
- [ ] `/adapter/http/` contains HTTP handlers
- [ ] No files in context root (except README.md, errors.go)
- [ ] All imports updated
- [ ] bootstrap.go updated
- [ ] All tests pass
- [ ] No lint errors
- [ ] Documentation updated

---

## Metrics & Progress Tracking

| Phase            | Contexts | Completed | In Progress | Remaining | % Complete |
| ---------------- | -------- | --------- | ----------- | --------- | ---------- |
| Phase 1 (New)    | 1        | 1         | 0           | 0         | 100%       |
| Phase 2 (High)   | 5        | 0         | 0           | 5         | 0%         |
| Phase 3 (Medium) | 20       | 0         | 0           | 20        | 0%         |
| Phase 4 (Low)    | 3        | 0         | 0           | 3         | 0%         |
| **Total**        | **29**   | **1**     | **0**       | **28**    | **3%**     |

---

## References

- [Context & Aggregate Naming Conventions](../guides/context-aggregate-naming-conventions.md)
- [Bounded Contexts Strategy](../concepts/bounded-contexts.md)
- [Clean Architecture](../concepts/clean-architecture.md)

---

**Last Updated**: January 17, 2026  
**Status**:  In Progress (Phase 1 Complete)
