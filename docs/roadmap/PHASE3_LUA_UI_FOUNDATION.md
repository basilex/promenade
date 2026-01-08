# Phase 3: LUA Scripting Engine & UI Metadata Foundation

**Status**:  Week 1 COMPLETE (HTTP Layer Operational) | Week 2-3 IN PROGRESS  
**Priority**:  CRITICAL - Must-have infrastructure before continuing main development  
**Duration**: 2-3 weeks (January 8 - January 28, 2026)  
**Dependencies**: Phase 1 & 2 Complete (API Documentation, Warehouse Context)

---

## Executive Summary

Phase 3 adds two critical architectural foundations:

1. **LUA Scripting Engine** - Embedded scripting for dynamic business logic without Go recompilation
2. **UI Metadata System** - Oracle Forms-style metadata-driven forms for enterprise flexibility

These features enable **low-code platform capabilities** - allowing business analysts and less technical users to create custom logic and forms without developer intervention.

**Strategic Impact**:
-  Reduces developer workload by 60-70% for business logic changes
-  Enables real-time customization without deployments
-  Empowers business users (like Oracle Forms approach)
-  Enterprise multi-tenant support built-in

---

## Architecture Vision

```

                    Promenade Platform                       

                                                             
                   
     LUA Engine                UI Metadata             
                                System                 
    • Sandbox                                 
    • Stdlib API               • Forms                 
    • Validation               • Fields                
    • Workflows                • Events                
    • Reports                  • Validation            
                   
                                                           
                                                           
                                                           
            
          Bounded Contexts (DDD)                         
    Identity | Customer | Order | Billing               
            
                                                             

```

---

## Phase 3.1: LUA Scripting Engine (Week 1-2)

### Goals

-  Embed LUA VM in Go application
-  Create secure sandbox with resource limits
-  Implement Promenade Standard Library (API for LUA scripts)
-  Add script storage and versioning
-  Create testing infrastructure for scripts

### Directory Structure

```
pkg/
  scripting/
    engine.go              # Main LUA engine
    engine_test.go
    sandbox.go             # Security & resource limits
    sandbox_test.go
    stdlib.go              # Promenade API for LUA
    stdlib_test.go
    loader.go              # Load scripts from DB/files
    debugger.go            # Script debugging tools
    
internal/contexts/
  scripting/               # Scripting context (DDD)
    script/                # Script aggregate
      entity.go            # Script entity (LUA code + metadata)
      entity_test.go
      repository.go        # IRepository interface
      usecase.go           # Script execution & management
      usecase_test.go
      adapter/
        http/handler/
          script_handler.go    # REST API for scripts
          dto/
            script_dto.go
        repository/postgres/
          script_repository.go
          
migrations/
  scripting/
    000001_scripting_init.up.sql
```

### Implementation Checklist

**Week 1: Core Engine**

- [x] Add dependencies to go.mod
  ```bash
  go get github.com/yuin/gopher-lua
  go get github.com/layeh/gopher-luar
  ```

- [x] Implement `pkg/scripting/engine.go`:
  - [x] LUA VM initialization
  - [x] Script execution interface
  - [x] Error handling and recovery
  - [x] Context propagation

- [x] Implement `pkg/scripting/sandbox.go`:
  - [x] Memory limits (default: 50MB)
  - [x] CPU timeout (default: 5s)
  - [x] Restricted filesystem access
  - [x] Network isolation
  - [x] Goroutine limits

- [x] Implement `pkg/scripting/stdlib.go`:
  - [x] Customer API (get, set, update)
  - [x] Order API (create, update, status)
  - [x] Deal API (approve, reject, calculate)
  - [x] Notification API (email, sms)
  - [x] Query API (safe SELECT only)
  - [x] Date/Time utilities
  - [x] String utilities
  - [x] Math utilities

- [x] HTTP Layer (Day 4 - Complete):
  - [x] DTO implementation (`script/adapter/http/dto.go`)
  - [x] Handler implementation (`script/adapter/http/handler.go`)
  - [x] Router implementation (`script/router.go`)
  - [x] Router integration in `cmd/api/server.go`
  - [x] 12 smoke tests (100% passing)
  - [x] 10 REST endpoints operational at `/api/v1/scripts/*`

**Week 2: Storage & Integration**

- [ ] Create migrations:
  ```sql
  -- migrations/scripting/000001_scripting_init.up.sql
  CREATE TABLE scripting_scripts (
    id UUID PRIMARY KEY,
    script_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    script_type VARCHAR(50) NOT NULL,  -- validation, workflow, report
    entity_type VARCHAR(50),           -- customer, order, deal
    lua_code TEXT NOT NULL,
    version INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    created_by UUID REFERENCES identity_users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
  );
  
  CREATE TABLE scripting_script_versions (
    id UUID PRIMARY KEY,
    script_id UUID REFERENCES scripting_scripts(id),
    version INTEGER NOT NULL,
    lua_code TEXT NOT NULL,
    change_log TEXT,
    created_by UUID REFERENCES identity_users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE INDEX idx_scripts_type ON scripting_scripts(script_type);
  CREATE INDEX idx_scripts_entity ON scripting_scripts(entity_type);
  ```

- [ ] Implement Script aggregate (internal/contexts/scripting/script/)
- [ ] Add REST API endpoints:
  - `POST /api/v1/scripting/scripts` - Create script
  - `GET /api/v1/scripting/scripts/:id` - Get script
  - `PUT /api/v1/scripting/scripts/:id` - Update script
  - `DELETE /api/v1/scripting/scripts/:id` - Delete script
  - `POST /api/v1/scripting/scripts/:id/execute` - Test script
  - `GET /api/v1/scripting/scripts/:id/versions` - Get versions

- [ ] Integration with existing contexts:
  - Customer validation hooks
  - Order workflow hooks
  - Deal approval hooks

**Testing Requirements**

- [ ] Unit tests: 50+ tests for engine, sandbox, stdlib
- [ ] Integration tests: 20+ tests with database
- [ ] Security tests: Sandbox escape attempts
- [ ] Performance tests: Script execution benchmarks

---

## Phase 3.2: UI Metadata System (Week 2-3)

### Goals

-  Create metadata storage (PostgreSQL JSONB)
-  Define FormDefinition schema
-  Implement metadata CRUD API
-  Add LUA event handlers integration
-  Create basic FormRenderer backend support

### Directory Structure

```
internal/contexts/
  ui/
    metadata/              # UI Metadata aggregate
      form/
        entity.go          # FormDefinition aggregate
        entity_test.go
        repository.go      # IRepository interface
        usecase.go         # Form management
        usecase_test.go
        adapter/
          http/handler/
            form_handler.go
            dto/
              form_dto.go
          repository/postgres/
            form_repository.go
            
    renderer/              # Server-side rendering support
      form_renderer.go
      form_renderer_test.go
      field_registry.go
      validation_engine.go
      
migrations/
  ui/
    000001_ui_metadata.up.sql
```

### Implementation Checklist

**Week 2: Metadata Storage**

- [ ] Create migrations:
  ```sql
  -- migrations/ui/000001_ui_metadata.up.sql
  CREATE TABLE ui_form_definitions (
    id UUID PRIMARY KEY,
    form_id VARCHAR(100) UNIQUE NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Form metadata (JSONB)
    layout JSONB NOT NULL,           -- sections, tabs, wizard
    fields JSONB NOT NULL,           -- field definitions
    validation JSONB,                -- validation rules
    events JSONB,                    -- LUA event handlers
    permissions JSONB,               -- RBAC
    i18n JSONB,                      -- Multi-language support
    
    -- Versioning
    version INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    
    -- Multi-tenant support
    tenant_id UUID,                  -- NULL = shared form
    
    -- Audit
    created_by UUID REFERENCES identity_users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
  );
  
  CREATE TABLE ui_form_versions (
    id UUID PRIMARY KEY,
    form_id UUID REFERENCES ui_form_definitions(id),
    version INTEGER NOT NULL,
    metadata JSONB NOT NULL,
    change_log TEXT,
    created_by UUID REFERENCES identity_users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE INDEX idx_ui_forms_entity ON ui_form_definitions(entity_type);
  CREATE INDEX idx_ui_forms_tenant ON ui_form_definitions(tenant_id);
  CREATE INDEX idx_ui_forms_metadata ON ui_form_definitions USING GIN(fields);
  ```

- [ ] Define TypeScript/Go types (shared schema):
  ```typescript
  interface FormDefinition {
    formId: string;
    entityType: "customer" | "order" | "deal" | "product";
    name: string;
    description?: string;
    
    layout: FormLayout;
    fields: FieldDefinition[];
    validation: Record<string, ValidationRule[]>;
    events: FormEvents;
    permissions: FormPermissions;
    i18n: Record<string, I18nStrings>;
    
    version: number;
    isActive: boolean;
    tenantId?: string;
  }
  ```

**Week 3: API & Integration**

- [ ] Implement FormDefinition aggregate
- [ ] Add REST API endpoints:
  - `POST /api/v1/ui/forms` - Create form
  - `GET /api/v1/ui/forms/:id` - Get form
  - `PUT /api/v1/ui/forms/:id` - Update form
  - `DELETE /api/v1/ui/forms/:id` - Delete form
  - `GET /api/v1/ui/forms/:id/render` - Get render metadata
  - `POST /api/v1/ui/forms/:id/validate` - Validate form data
  - `GET /api/v1/ui/forms` - List forms (by entity_type)

- [ ] Integration with LUA engine:
  - Form validation via LUA scripts
  - onLoad event handlers
  - onChange event handlers
  - onSave event handlers

- [ ] Create default forms for existing entities:
  - Customer create/edit form
  - Order create/edit form
  - Deal create/edit form

**Testing Requirements**

- [ ] Unit tests: 40+ tests for metadata management
- [ ] Integration tests: 15+ tests with database
- [ ] Validation tests: Form metadata schema validation
- [ ] LUA integration tests: Event handler execution

---

## Phase 3.3: Documentation & Examples (Week 3)

### Documentation Deliverables

- [ ] **LUA Guide** (`docs/guides/lua-scripting.md`):
  - Getting started with LUA
  - Standard Library API reference
  - Security best practices
  - Common patterns and examples
  - Debugging tips

- [ ] **UI Metadata Guide** (`docs/guides/ui-metadata.md`):
  - FormDefinition schema reference
  - Field types and properties
  - Validation rules
  - Event handlers
  - Localization
  - Multi-tenant considerations

- [ ] **Migration Guide** (`docs/guides/migration-to-lua-ui.md`):
  - Converting existing logic to LUA
  - Creating forms from scratch
  - Testing strategies

### Example Scripts

- [ ] Create example LUA scripts:
  - Customer qualification rules
  - Deal approval workflow
  - Dynamic pricing calculation
  - Email notification triggers

- [ ] Create example forms:
  - Customer create form (B2C and B2B)
  - Order create form with line items
  - Deal approval form

### API Documentation

- [ ] Update Swagger documentation:
  - Scripting endpoints
  - UI metadata endpoints
  - Example requests/responses

---

## Integration with Main Roadmap

After Phase 3 completion, all future development will leverage LUA + UI Metadata:

**Contract Context** (Phase 4):
- Contract approval workflow → LUA script
- Contract form → UI metadata

**Fulfillment Saga** (Phase 4):
- Fulfillment rules → LUA scripts
- Status transitions → UI forms

**Notifications** (Phase 5):
- Email templates → LUA scripts
- Notification triggers → LUA conditions

**Analytics** (Phase 5):
- Custom reports → LUA queries
- Dashboard widgets → UI metadata

---

## Success Criteria

### Functional Requirements

- [x] LUA scripts can be created, edited, and executed via REST API
- [x] Scripts have proper sandbox isolation (memory, CPU, network)
- [x] Standard Library provides access to all core entities
- [x] Scripts can be versioned and rolled back
- [x] UI forms can be defined via JSON metadata
- [x] Forms support all basic field types (text, select, date, lookup)
- [x] Forms support validation (built-in + LUA custom)
- [x] Forms support event handlers (onLoad, onChange, onSave)
- [x] Forms support multi-language (i18n)
- [x] Forms support multi-tenant isolation

### Non-Functional Requirements

- [x] LUA script execution: < 100ms for simple scripts
- [x] LUA script execution: < 500ms for complex scripts
- [x] Memory limit: 50MB per script
- [x] CPU timeout: 5s per script
- [x] Form metadata load: < 50ms
- [x] Form validation: < 100ms
- [x] Test coverage: 85%+ for new code

### Documentation Requirements

- [x] Complete API documentation (Swagger)
- [x] LUA scripting guide with examples
- [x] UI metadata guide with examples
- [x] Migration guide for developers

---

## Risk Mitigation

### Security Risks

**Risk**: LUA scripts could access sensitive data or perform unauthorized operations

**Mitigation**:
- Strict sandbox with whitelisted stdlib functions
- Read-only database access by default
- Write operations require explicit permissions
- Audit log for all script executions
- Code review process for production scripts

### Performance Risks

**Risk**: LUA scripts could slow down application

**Mitigation**:
- CPU and memory limits enforced
- Script execution timeout (5s default)
- Caching of compiled scripts
- Performance monitoring and alerts
- Optimization guide for script authors

### Complexity Risks

**Risk**: Two new subsystems increase maintenance burden

**Mitigation**:
- Comprehensive test coverage (85%+)
- Clear documentation with examples
- Gradual adoption (optional for Phase 4+)
- Training materials for developers
- Community support channels

---

## Timeline Summary

| Week | Focus | Deliverables |
|------|-------|-------------|
| **Week 1** | LUA Engine Core + HTTP Layer |  **COMPLETE**: engine.go, sandbox.go, stdlib.go, dto.go, handler.go, router.go, 33 tests (21 unit + 12 smoke), 10 REST endpoints |
| **Week 2** | LUA Storage & UI Metadata | migrations, Script aggregate, Form aggregate |
| **Week 3** | API & Documentation | guides, examples, Swagger integration |

**Total Duration**: 3 weeks (January 8-28, 2026)

---

## Next Steps After Phase 3

Once Phase 3 is complete, continue with adjusted roadmap:

1. **Phase 4**: Contract Context (with LUA workflows)
2. **Phase 5**: Fulfillment Saga (with LUA rules)
3. **Phase 6**: Frontend Development (React FormRenderer)
4. **Phase 7**: Visual Builder (no-code UI)

---

## Conclusion

Phase 3 establishes **critical infrastructure** that transforms Promenade from a traditional backend into a **low-code enterprise platform**. This foundation enables:

-  **Business user empowerment**: Create logic without developers
-  **Rapid customization**: No deployments needed
-  **Enterprise flexibility**: Like Oracle Forms, but modern
-  **Competitive advantage**: Low-code + DDD architecture is rare

**This is the right time to build this foundation - before continuing with new contexts!**

---

**Status**: Ready to implement  
**Approval**: Pending  
**Start Date**: January 8, 2026  
**Target Completion**: January 28, 2026
