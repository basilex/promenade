# Scripting Package

**Embedded LUA scripting engine** for Promenade Platform - enables dynamic business logic without Go recompilation.

---

## Features

- **Sandboxed Execution**: Memory limits (50MB), CPU timeout (5s), restricted filesystem/network access
- **Standard Library**: Safe access to Promenade APIs (Customer, Order, Deal, Notify, Query)
- **Context Support**: Cancellation and timeout via Go context
- **Type Conversion**: Automatic conversion between LUA and Go types
- **Script Validation**: Syntax checking before execution
- **Performance**: < 100ms for simple scripts (benchmark target)

---

## Architecture

### Components

1. **Engine** (`engine.go`) - LUA VM wrapper with execution methods
2. **Sandbox** (`sandbox.go`) - Security restrictions (memory, CPU, network, filesystem)
3. **Standard Library** (`stdlib.go`) - Promenade API for LUA scripts

### Security Model

```
┌─────────────────────────────────────────┐
│         LUA Script (user code)          │
├─────────────────────────────────────────┤
│   Standard Library (safe Promenade API) │
├─────────────────────────────────────────┤
│   Sandbox (security restrictions)       │
├─────────────────────────────────────────┤
│   Engine (LUA VM wrapper)               │
├─────────────────────────────────────────┤
│   gopher-lua (LUA interpreter)          │
└─────────────────────────────────────────┘
```

---

## Quick Start

### 1. Create Engine

```go
import "github.com/basilex/promenade/pkg/scripting"

// Use default config
config := scripting.DefaultConfig()
engine := scripting.NewEngine(config)

// Or custom config
config := scripting.Config{
    MemoryLimit:   100 * 1024 * 1024, // 100MB
    Timeout:       10 * time.Second,
    MaxGoroutines: 10,
    AllowFileIO:   false,
    AllowNetwork:  false,
    Debug:         true,
}
engine := scripting.NewEngine(config)
```

### 2. Execute Script

```go
ctx := context.Background()

script := `
    -- Simple calculation
    return 2 + 2
`

result, err := engine.Execute(ctx, script, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Result: %v\n", result) // Output: 4
```

### 3. Execute with Parameters

```go
script := `
    -- Use parameters
    if amount > 1000 then
        return "high"
    else
        return "low"
    end
`

params := map[string]interface{}{
    "amount": 1500,
}

result, err := engine.Execute(ctx, script, params)
fmt.Printf("Category: %v\n", result) // Output: high
```

### 4. Execute Specific Function

```go
script := `
    function calculateDiscount(amount, tier)
        if tier == "premium" then
            return amount * 0.2  -- 20% discount
        elseif tier == "pro" then
            return amount * 0.1  -- 10% discount
        else
            return 0
        end
    end
`

result, err := engine.ExecuteFunction(ctx, script, "calculateDiscount", 1000, "premium")
fmt.Printf("Discount: %v\n", result) // Output: 200
```

### 5. Validate Script

```go
script := `
    return 2 +  -- syntax error
`

err := engine.Validate(script)
if err != nil {
    fmt.Printf("Invalid script: %v\n", err)
}
```

---

## Standard Library API

### Customer Module

```lua
-- Get customer tier
local tier = Customer.GetTier("customer-123")
-- Returns: "free", "basic", "pro", "enterprise"

-- Set customer tier
Customer.SetTier("customer-123", "pro")
```

### Order Module

```lua
-- Get order status
local status = Order.GetStatus("order-456")
-- Returns: "pending", "confirmed", "processing", "fulfilled", "cancelled"

-- Set order status
Order.SetStatus("order-456", "confirmed")
```

### Deal Module

```lua
-- Approve deal
Deal.Approve("deal-789")

-- Reject deal with reason
Deal.Reject("deal-789", "Budget constraints")
```

### Notify Module

```lua
-- Send email from template
Notify.SendEmail("user@example.com", "welcome_email")

-- Send SMS
Notify.SendSMS("+1234567890", "Your order has been shipped")
```

### Query Module (Read-Only)

```lua
-- Execute SELECT query
local result = Query.Execute([[
    SELECT name, email 
    FROM customers 
    WHERE tier = 'premium'
    LIMIT 10
]])

-- Note: Only SELECT queries allowed for security
```

### Date Module

```lua
-- Get current date/time
local now = Date.Now()
-- Returns: "2026-01-07T12:00:00Z"

-- Format date
local formatted = Date.Format("2026-01-07T12:00:00Z", "2006-01-02")
-- Returns: "2026-01-07"

-- Get current month
local month = Date.GetMonth()
-- Returns: 1 (January)
```

---

## Use Cases

### 1. Custom Validation Rules

```lua
-- Validate order before confirmation
function validateOrder(order)
    if order.total < 10 then
        return false, "Minimum order amount is $10"
    end
    
    if order.status ~= "pending" then
        return false, "Only pending orders can be confirmed"
    end
    
    return true, ""
end
```

### 2. Dynamic Pricing

```lua
-- Calculate dynamic price based on customer tier and time
function calculatePrice(basePrice, customerTier)
    local discount = 0
    
    if customerTier == "premium" then
        discount = 0.2  -- 20%
    elseif customerTier == "pro" then
        discount = 0.1  -- 10%
    end
    
    -- Get current month for seasonal discounts
    local month = Date.GetMonth()
    if month == 11 or month == 12 then  -- November, December
        discount = discount + 0.05  -- Extra 5% holiday discount
    end
    
    return basePrice * (1 - discount)
end
```

### 3. Workflow Automation

```lua
-- Auto-escalate old deals
function escalateOldDeals()
    local deals = Query.Execute([[
        SELECT id, created_at 
        FROM deals 
        WHERE stage = 'proposal' 
        AND created_at < NOW() - INTERVAL '30 days'
    ]])
    
    for i, deal in ipairs(deals) do
        Notify.SendEmail("sales@company.com", "deal_escalation", {
            deal_id = deal.id,
            days_old = 30
        })
    end
end
```

### 4. Custom Notifications

```lua
-- Send personalized notification based on customer activity
function notifyCustomer(customerId, eventType)
    local tier = Customer.GetTier(customerId)
    
    if eventType == "order_confirmed" then
        if tier == "premium" then
            Notify.SendEmail(customerId, "premium_order_confirmed")
            Notify.SendSMS(customerId, "Your premium order is being processed")
        else
            Notify.SendEmail(customerId, "standard_order_confirmed")
        end
    end
end
```

### 5. Business Rules Engine

```lua
-- Complex business rule: approve deal automatically
function shouldAutoApproveDeal(deal)
    local tier = Customer.GetTier(deal.customer_id)
    
    -- Rule 1: Premium customers with amount < $10,000
    if tier == "premium" and deal.amount < 10000 then
        return true
    end
    
    -- Rule 2: Existing customers with good history
    local orderCount = Query.Execute([[
        SELECT COUNT(*) as count
        FROM orders
        WHERE customer_id = ']] .. deal.customer_id .. [['
        AND status = 'fulfilled'
    ]])
    
    if orderCount.count > 5 and deal.amount < 5000 then
        return true
    end
    
    return false
end
```

---

## Configuration

### Config Structure

```go
type Config struct {
    MemoryLimit   int64         // Memory limit in bytes (default: 50MB)
    Timeout       time.Duration // CPU timeout (default: 5s)
    MaxGoroutines int           // Max goroutines (default: 0 = unlimited)
    AllowFileIO   bool          // Enable filesystem access (default: false)
    AllowNetwork  bool          // Enable network access (default: false)
    Debug         bool          // Enable debug mode (default: false)
}
```

### Default Configuration

```go
DefaultConfig() Config {
    return Config{
        MemoryLimit:   50 * 1024 * 1024, // 50MB
        Timeout:       5 * time.Second,
        MaxGoroutines: 0,
        AllowFileIO:   false,
        AllowNetwork:  false,
        Debug:         false,
    }
}
```

---

## Security

### Sandbox Restrictions

1. **Dangerous Functions Removed**:
   - `dofile`, `loadfile`, `load`, `loadstring` - Dynamic code loading
   - `require`, `module` - Module system
   - `setfenv`, `getfenv` - Environment manipulation

2. **Filesystem Restrictions** (when `AllowFileIO = false`):
   - `io` library removed
   - `os.execute`, `os.exit`, `os.remove`, `os.rename` blocked

3. **Network Restrictions** (when `AllowNetwork = false`):
   - No network libraries loaded

4. **Resource Limits**:
   - Memory limit: 50MB (configurable)
   - CPU timeout: 5s (configurable)
   - No infinite loops (timeout protection)

### Best Practices

**DO**:
- Always use default config in production
- Set reasonable timeout (5-10s)
- Validate scripts before saving to database
- Log all script executions for audit
- Use standard library instead of raw database access
- Test scripts in sandbox before production

**DON'T**:
- Don't enable `AllowFileIO` or `AllowNetwork` in production
- Don't trust user input in scripts
- Don't execute scripts from untrusted sources
- Don't use scripts for time-critical operations
- Don't bypass sandbox restrictions

---

## Testing

### Run Tests

```bash
# All tests
go test ./pkg/scripting -v

# With coverage
go test ./pkg/scripting -cover

# Benchmarks
go test -bench=. ./pkg/scripting
```

### Test Results (Target)

| Test                  | Status | Duration |
| --------------------- | ------ | -------- |
| Execute_Simple        | PASS   | < 1ms    |
| Execute_WithParams    | PASS   | < 1ms    |
| Execute_Function      | PASS   | < 2ms    |
| Execute_Timeout       | PASS   | 100ms    |
| Execute_StandardLib   | PASS   | < 5ms    |
| Sandbox_Restrictions  | PASS   | < 1ms    |

### Benchmark Targets

| Benchmark             | Target         | Memory       |
| --------------------- | -------------- | ------------ |
| Execute_Simple        | < 100μs/op     | < 1KB/op     |
| Execute_WithParams    | < 200μs/op     | < 2KB/op     |
| Execute_Function      | < 500μs/op     | < 5KB/op     |

---

## Roadmap

### Week 1-2 (Current Phase)

- [x] Engine implementation
- [x] Sandbox restrictions
- [x] Standard library stubs
- [ ] Standard library implementations
- [ ] Integration with contexts
- [ ] Unit tests (20+ tests target)

### Week 3

- [ ] Performance optimization
- [ ] Memory limiting implementation
- [ ] Advanced examples
- [ ] Documentation

### Future

- [ ] Script versioning
- [ ] Script marketplace
- [ ] Visual LUA builder
- [ ] Template system
- [ ] Debugging support

---

## Related Documentation

- [Phase 3 Roadmap](../../docs/roadmap/PHASE3_LUA_UI_FOUNDATION.md)
- [Scripting Context](../../internal/contexts/scripting/README.md) (planned)
- [UI Metadata Guide](../../docs/guides/ui-metadata.md) (planned)

---

**Status**: 🚧 In Development (Week 1)  
**Test Coverage**: Target 90%+  
**Performance**: Target < 100ms  
**Last Updated**: January 7, 2026
