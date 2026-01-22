# Scripting Context

**Scripting** - bounded context for managing and executing Lua scripts within the application, enabling custom business logic without code deployment.

---

## Overview

The Scripting context provides a sandboxed Lua execution environment for customer-defined automation, data transformations, validation rules, and custom business logic. Scripts are stored as database records and executed on-demand.

**Core Concept**: Customer writes Lua script → Promenade executes securely → Returns result

---

## Domain Model

### Aggregates

See [script/README.md](script/README.md) for detailed domain errors and patterns.

**Script** - Executable Lua code with metadata

- Script ID (UUIDv7)
- Name, description
- Lua source code
- Script type (validation, transformation, automation)
- Execution timeout
- Enabled/disabled status
- Version, timestamps

---

## Use Cases

### Script Lifecycle

- **CreateScript** - Store new Lua script
- **UpdateScript** - Modify existing script
- **GetScript** - Retrieve script by ID
- **ListScripts** - List all scripts (with filters)
- **DeleteScript** - Remove script
- **ExecuteScript** - Run script with input data

### Script Types

1. **Validation Scripts** - Custom validation logic

   ```lua
   -- Example: Validate customer credit limit
   function validate(customer)
       if customer.credit_limit > 100000 and customer.tier ~= "premium" then
           return false, "High credit limit requires premium tier"
       end
       return true, nil
   end
   ```

2. **Transformation Scripts** - Data mapping/enrichment

   ```lua
   -- Example: Normalize phone number
   function transform(phone)
       local cleaned = phone:gsub("[^0-9]", "")
       return "+1" .. cleaned
   end
   ```

3. **Automation Scripts** - Custom workflows
   ```lua
   -- Example: Auto-approve orders under threshold
   function automate(order)
       if order.total < 1000 and order.customer_tier == "premium" then
           return {action = "approve", reason = "Auto-approved"}
       end
       return {action = "review"}
   end
   ```

---

## HTTP API

### Endpoints

```
POST   /api/v1/scripting/scripts              # Create script
GET    /api/v1/scripting/scripts              # List scripts
GET    /api/v1/scripting/scripts/:id          # Get script
PUT    /api/v1/scripting/scripts/:id          # Update script
DELETE /api/v1/scripting/scripts/:id          # Delete script
POST   /api/v1/scripting/scripts/:id/execute  # Execute script
```

### Example: Create Script

```bash
curl -X POST http://localhost:8081/api/v1/scripting/scripts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "validate_credit_limit",
    "description": "Validate customer credit limit based on tier",
    "script_type": "validation",
    "source_code": "function validate(customer)\n  if customer.credit_limit > 100000 and customer.tier ~= \"premium\" then\n    return false, \"High credit limit requires premium tier\"\n  end\n  return true, nil\nend",
    "timeout_ms": 5000,
    "enabled": true
  }'
```

### Example: Execute Script

```bash
curl -X POST http://localhost:8081/api/v1/scripting/scripts/01933f76-8b4a-7890-abcd-1234567890ab/execute \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "customer": {
        "credit_limit": 150000,
        "tier": "standard"
      }
    }
  }'
```

Response:

```json
{
  "success": false,
  "result": null,
  "error": "High credit limit requires premium tier",
  "execution_time_ms": 12
}
```

---

## Architecture

### Directory Structure

```
scripting/
 router.go                 # HTTP routes registration
 README.md                 # This file
 script/
     README.md             # Domain errors documentation
     aggregate/
        script.go         # Script aggregate
     adapter/
        http/
           script_handler.go    # HTTP handlers
        repository/
            postgres/
                script_repository.go  # PostgreSQL implementation
     usecase/
        script_usecase.go        # Business logic + Lua execution
     repository/
        script_repository.go     # Data access interface
     dto/
        script_dto.go            # Request/response DTOs
     errors.go                    # Domain error constants
```

---

## Lua Sandbox

### Security Features

1. **Limited Standard Library** - Only safe modules exposed
   - `string`, `table`, `math` - Allowed
   - `io`, `os`, `debug` - Blocked

2. **Execution Timeout** - Default 5s, max 30s

   ```go
   ctx, cancel := context.WithTimeout(context.Background(), script.TimeoutMS)
   defer cancel()
   ```

3. **Memory Limits** -  **TODO** (see CODE_REVIEW Issue 1.1)
   - No memory limits currently enforced
   - Future: Limit to 10MB per script execution

4. **CPU Limits** - Via context timeout
   - Prevents infinite loops
   - Kills long-running scripts

### Available Functions

Scripts can access:

```lua
-- String operations
string.upper(s)
string.lower(s)
string.sub(s, i, j)
string.gsub(s, pattern, replacement)

-- Table operations
table.insert(t, value)
table.remove(t, pos)
table.sort(t)

-- Math operations
math.floor(x)
math.ceil(x)
math.max(...)
math.min(...)

-- JSON encoding/decoding (custom)
json.encode(table)
json.decode(string)
```

---

## Integration Points

### Consumed By

- **Customer Management** - Validation rules
- **Order Management** - Auto-approval logic
- **Billing** - Custom pricing calculations
- **Fiscal** - Tax calculation overrides

### Event Publishing

Scripts can trigger events:

```lua
-- Example: Publish event from script
function process_order(order)
    if order.total > 10000 then
        emit_event("high_value_order", {order_id = order.id})
    end
end
```

### Dependencies

- **Event Bus** - For publishing events from scripts
- **PostgreSQL** - Script storage

---

## Use Cases

### 1. Dynamic Discount Rules

```lua
-- Script: calculate_discount
function calculate(order)
    local total = order.total
    local customer_tier = order.customer_tier

    if customer_tier == "premium" and total > 5000 then
        return 0.15  -- 15% discount
    elseif customer_tier == "standard" and total > 10000 then
        return 0.10  -- 10% discount
    end

    return 0  -- No discount
end
```

Backend execution:

```go
discount := scriptUseCase.Execute(ctx, scriptID, map[string]interface{}{
    "order": order,
})
```

### 2. Custom Data Validation

```lua
-- Script: validate_invoice
function validate(invoice)
    local errors = {}

    if invoice.total <= 0 then
        table.insert(errors, "Invoice total must be positive")
    end

    if #invoice.lines == 0 then
        table.insert(errors, "Invoice must have at least one line")
    end

    for i, line in ipairs(invoice.lines) do
        if line.quantity <= 0 then
            table.insert(errors, "Line " .. i .. ": quantity must be positive")
        end
    end

    if #errors > 0 then
        return false, table.concat(errors, "; ")
    end

    return true, nil
end
```

### 3. Data Enrichment

```lua
-- Script: enrich_customer
function enrich(customer)
    -- Add computed fields
    customer.full_name = customer.first_name .. " " .. customer.last_name
    customer.age_group = compute_age_group(customer.birth_date)
    customer.risk_score = compute_risk_score(customer)

    return customer
end

function compute_age_group(birth_date)
    local age = os.date("%Y") - string.sub(birth_date, 1, 4)
    if age < 25 then return "young" end
    if age < 45 then return "middle" end
    return "senior"
end
```

---

## Error Handling

See [script/README.md](script/README.md) for complete domain error catalog.

### Domain Errors (21 total)

```go
// Validation errors
ErrScriptNameRequired
ErrScriptSourceRequired
ErrScriptTypeInvalid
ErrScriptTimeoutInvalid

// Business logic errors
ErrScriptDisabled
ErrScriptNotFound
ErrScriptExecutionFailed
ErrScriptTimeout
ErrScriptSyntaxError

// Security errors
ErrScriptForbiddenFunction  // Attempted to use blocked function
ErrScriptMemoryExceeded     // Memory limit exceeded (future)
```

### HTTP Response Codes

| Domain Error               | HTTP Status | Exposed Message            |
| -------------------------- | ----------- | -------------------------- |
| `ErrScriptNameRequired`    | 400         | "script name required"     |
| `ErrScriptNotFound`        | 404         | "script not found"         |
| `ErrScriptExecutionFailed` | 500         | "script execution failed"  |
| `ErrScriptTimeout`         | 408         | "script execution timeout" |

---

## Testing

### Unit Tests

```bash
go test ./internal/contexts/scripting/script/...
```

### Integration Tests

```bash
make test-integration CONTEXT=scripting
```

### Test Coverage

```bash
go test -cover ./internal/contexts/scripting/...
```

### Example Test

```go
func TestExecuteScript_Validation(t *testing.T) {
    script := &aggregate.Script{
        SourceCode: `
            function validate(value)
                if value < 0 then
                    return false, "value must be non-negative"
                end
                return true, nil
            end
        `,
    }

    result, err := executor.Execute(ctx, script, map[string]interface{}{
        "value": -10,
    })

    assert.False(t, result.Success)
    assert.Equal(t, "value must be non-negative", result.Error)
}
```

---

## Performance

### Benchmarks

- **Script Parsing**: ~2ms (cached after first parse)
- **Simple Execution**: ~1-5ms
- **Complex Scripts**: ~10-50ms
- **Timeout Overhead**: ~1ms

### Optimization

1. **Script Caching** - Parse once, execute many times
2. **Pool Lua States** - Reuse VM instances
3. **Timeout Monitoring** - Context-based cancellation

---

## Database Schema

### Table: `scripting_scripts`

```sql
CREATE TABLE scripting_scripts (
    id            UUID PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    description   TEXT,
    script_type   VARCHAR(50) NOT NULL,  -- validation, transformation, automation
    source_code   TEXT NOT NULL,
    timeout_ms    INTEGER DEFAULT 5000,
    enabled       BOOLEAN DEFAULT true,
    version       INTEGER DEFAULT 1,
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_scripting_scripts_type ON scripting_scripts(script_type);
CREATE INDEX idx_scripting_scripts_enabled ON scripting_scripts(enabled);
```

---

## Security Considerations

###  Implemented

1. **Sandboxed Execution** - No access to OS, filesystem, network
2. **Timeout Enforcement** - Prevents infinite loops
3. **Limited API** - Only safe Lua standard library functions
4. **Input Validation** - Validate script source before execution

###  TODO

1. **Memory Limits** - Issue 1.1 in CODE_REVIEW (Phase 1)
2. **Rate Limiting** - Per-user script execution limits
3. **Audit Logging** - Log all script executions
4. **Code Review** - Admin approval for production scripts

---

## Known Limitations

1. **No Memory Limits** -  Lua VM can consume unlimited memory (Issue 1.1)
2. **No Concurrency Control** - Scripts execute sequentially
3. **No Import System** - Cannot import other scripts
4. **No Debugging** - No step-through debugger
5. **Single Lua Version** - Tied to `gopher-lua` implementation

---

## Future Enhancements

### Planned Features

1. **Script Library** - Reusable function modules
2. **Version Control** - Script versioning with rollback
3. **Testing Framework** - Unit tests for scripts
4. **IDE Integration** - Syntax highlighting, autocomplete
5. **Debugging Tools** - Step-through debugger
6. **Performance Metrics** - Execution time tracking
7. **Script Marketplace** - Share scripts between tenants

### Example: Script Library

```lua
-- Import shared library
local utils = require("utils")

function validate(customer)
    if not utils.validate_email(customer.email) then
        return false, "invalid email"
    end
    return true, nil
end
```

---

## Related Documentation

- [Script Domain Errors](script/README.md) - Domain error catalog
- [Lua Scripting Guide](../../docs/guides/lua-scripting.md) - Writing secure scripts
- [CODE_REVIEW Issue 1.1](../../docs/CODE_REVIEW_2026-01-22.md#L50) - Lua sandbox memory limits
- [Scripting API Reference](../../docs/swagger/) - OpenAPI specs

---

## Migration

### Create Scripts Table

```bash
make migrate-module MODULE=scripting
```

### Rollback

```bash
make migrate-rollback MODULE=scripting STEPS=1
```

---

## Quick Start

### 1. Create a Validation Script

```bash
curl -X POST http://localhost:8081/api/v1/scripting/scripts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "validate_positive",
    "script_type": "validation",
    "source_code": "function validate(n) return n > 0, nil end",
    "enabled": true
  }'
```

### 2. Execute Script

```bash
curl -X POST http://localhost:8081/api/v1/scripting/scripts/{id}/execute \
  -H "Content-Type: application/json" \
  -d '{"input": {"n": 42}}'
```

### 3. List All Scripts

```bash
curl http://localhost:8081/api/v1/scripting/scripts
```

---

**Status**: Production-ready (with memory limit TODO)  
**Maintainer**: Promenade Team  
**Last Updated**: 2026-01-22  
**Security Note**:  Memory limits not enforced (see Issue 1.1)
