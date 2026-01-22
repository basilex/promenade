# UI Metadata Context

**UI Metadata** - bounded context for dynamic form generation and UI configuration metadata.

---

## Overview

The UI Metadata context provides server-driven UI capabilities, allowing the backend to define and control form structures, validation rules, and field configurations. This enables dynamic UI generation without frontend code changes.

**Core Concept**: Backend defines what UI looks like → Frontend renders based on metadata

---

## Domain Model

### Aggregates

**FormMetadata** - Defines complete form structure

- Form ID, title, description
- Field definitions (type, validation, options)
- Layout configuration
- Submission rules

---

## Use Cases

### Form Metadata Management

- **GetFormMetadata** - Retrieve form definition by ID
- **ListForms** - Get available form definitions
- **ValidateSubmission** - Server-side validation based on metadata

**Example Forms**:

- Customer creation form
- Invoice entry form
- Product configuration form
- Dynamic search filters

---

## HTTP API

### Endpoints

```
GET    /api/v1/ui/forms                # List available forms
GET    /api/v1/ui/forms/:id            # Get form metadata
POST   /api/v1/ui/forms/:id/validate   # Validate form submission
```

### Example Response

```json
{
  "form_id": "customer_create",
  "title": "Create Customer",
  "fields": [
    {
      "id": "name",
      "type": "text",
      "label": "Customer Name",
      "required": true,
      "validation": {
        "min_length": 2,
        "max_length": 100
      }
    },
    {
      "id": "email",
      "type": "email",
      "label": "Email Address",
      "required": true,
      "validation": {
        "pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
      }
    },
    {
      "id": "status",
      "type": "select",
      "label": "Status",
      "options": [
        { "value": "lead", "label": "Lead" },
        { "value": "prospect", "label": "Prospect" },
        { "value": "customer", "label": "Customer" }
      ]
    }
  ]
}
```

---

## Architecture

### Directory Structure

```
ui/
 router.go                 # HTTP routes registration
 README.md                 # This file
 metadata/
     form/
         aggregate/
            form_metadata.go     # Form definition aggregate
         adapter/
            http/
                form_handler.go  # HTTP handlers
         usecase/
            form_usecase.go      # Business logic
         repository/
            form_repository.go   # Data access interface
         dto/
             form_dto.go          # Request/response DTOs
```

---

## Integration Points

### Consumed By

- **Frontend Applications** - Next.js, Flutter apps
- **Mobile Apps** - Dynamic form rendering
- **Admin Panels** - Configuration management

### Dependencies

- None (standalone context)

---

## Field Types

Supported field types for dynamic forms:

| Type       | Description                 | Validation                      |
| ---------- | --------------------------- | ------------------------------- |
| `text`     | Single-line text input      | min_length, max_length, pattern |
| `textarea` | Multi-line text input       | min_length, max_length          |
| `email`    | Email input with validation | pattern (email format)          |
| `number`   | Numeric input               | min, max, step                  |
| `select`   | Dropdown selection          | options (predefined list)       |
| `checkbox` | Boolean checkbox            | None                            |
| `radio`    | Radio button group          | options (predefined list)       |
| `date`     | Date picker                 | min_date, max_date              |
| `file`     | File upload                 | allowed_types, max_size         |

---

## Validation Rules

Server-side validation based on metadata:

```go
// Example: Validate customer name field
field := &FormField{
    ID: "name",
    Type: "text",
    Required: true,
    Validation: map[string]interface{}{
        "min_length": 2,
        "max_length": 100,
    },
}

// Validation logic
if value == "" && field.Required {
    return errors.New("field is required")
}
if len(value) < field.Validation["min_length"].(int) {
    return errors.New("value too short")
}
```

---

## Use Cases

### 1. Dynamic Customer Form

Backend defines form structure:

```yaml
form_id: customer_create
fields:
  - name: Company Name (text, required)
  - email: Email (email, required)
  - phone: Phone (text, optional)
  - country: Country (select, options from DB)
```

Frontend receives metadata and renders form automatically.

### 2. Search Filter Builder

Backend defines available filters:

```yaml
form_id: customer_search
fields:
  - status: Status (multi-select)
  - tier: Tier (select)
  - created_after: Created After (date)
  - created_before: Created Before (date)
```

Frontend builds dynamic search UI.

### 3. Invoice Entry

Backend defines invoice line fields:

```yaml
form_id: invoice_line
fields:
  - product: Product (autocomplete)
  - quantity: Quantity (number, min=1)
  - unit_price: Unit Price (number, min=0)
  - tax_code: Tax Code (select)
```

---

## Benefits

### 1. Flexibility

- Change form structure without frontend deployment
- A/B test different form layouts
- Customize forms per customer/tenant

### 2. Consistency

- Validation rules enforced server-side
- Single source of truth for form definitions
- Reduced frontend/backend validation drift

### 3. Rapid Development

- Add new forms by defining metadata
- No frontend code changes needed
- Faster feature iteration

---

## Future Enhancements

### Planned Features

1. **Conditional Fields** - Show/hide based on other field values
2. **Field Dependencies** - Cascade selection (country → state → city)
3. **Custom Validation Functions** - Lua script validation rules
4. **Multi-page Forms** - Wizard-style forms with steps
5. **Form Templates** - Reusable field groups
6. **Localization** - Multi-language form labels

### Example: Conditional Fields

```json
{
  "id": "payment_method",
  "type": "select",
  "options": ["card", "bank_transfer"],
  "conditional": {
    "card": {
      "show_fields": ["card_number", "cvv", "expiry"]
    },
    "bank_transfer": {
      "show_fields": ["bank_account", "swift_code"]
    }
  }
}
```

---

## Related Documentation

- [UI Metadata Guide](../../docs/guides/ui-metadata.md) - Detailed implementation guide
- [API Documentation](../../docs/swagger/) - OpenAPI specs
- [Frontend Integration](../../front/README.md) - Next.js/Flutter integration

---

## Testing

### Run Tests

```bash
# Unit tests
go test ./internal/contexts/ui/...

# Integration tests
make test-integration CONTEXT=ui

# All tests
make test
```

### Test Coverage

```bash
go test -cover ./internal/contexts/ui/...
```

---

## Quick Start

### 1. Get Form Metadata

```bash
curl http://localhost:8081/api/v1/ui/forms/customer_create
```

### 2. Validate Submission

```bash
curl -X POST http://localhost:8081/api/v1/ui/forms/customer_create/validate \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "email": "contact@acme.com"
  }'
```

### 3. Render Form (Frontend)

```typescript
// Next.js example
const metadata = await fetch('/api/v1/ui/forms/customer_create').then(r => r.json());

return (
  <DynamicForm
    metadata={metadata}
    onSubmit={handleSubmit}
  />
);
```

---

**Status**: Production-ready  
**Maintainer**: Promenade Team  
**Last Updated**: 2026-01-22
