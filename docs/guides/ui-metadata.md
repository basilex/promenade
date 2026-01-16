# UI Metadata Guide

**Metadata-driven UI forms** for Promenade Platform. This guide describes the `FormDefinition` aggregate, storage format, API endpoints, and integration patterns.

---

## Overview

The UI Metadata system provides dynamic form definitions stored in the database. Forms are defined with JSON metadata and rendered by clients without redeploying the backend.

Key goals:
- Centralized form configuration
- Versioned changes
- Multi-tenant support
- LUA event hooks for validation and workflows

---

## FormDefinition Model

`FormDefinition` is the aggregate root for UI forms.

Fields:
- `form_id` (string): unique identifier (e.g., `customer_form`)
- `entity_type` (string): target entity (e.g., `customer`, `order`)
- `name` (string): display name
- `description` (string, optional)
- `layout` (JSON): layout metadata
- `fields` (JSON array): field definitions
- `validation` (JSON, optional): validation rules
- `events` (JSON, optional): LUA hooks
- `permissions` (JSON, optional): RBAC rules
- `i18n` (JSON, optional): localized labels
- `version` (int): aggregate version
- `is_active` (bool)
- `tenant_id` (UUID, optional)
- `created_by` (UUID, optional)

Storage:
- JSON metadata stored as TEXT for multi-database support.
- Accessed via `jsonstore.Field[T]` in Go.

---

## JSON Metadata Examples

### Layout
```json
{
  "type": "single_column",
  "sections": [
    {"title": "General", "fields": ["email", "name"]}
  ]
}
```

### Fields
```json
[
  {
    "name": "email",
    "label": "Email",
    "type": "text",
    "required": true
  },
  {
    "name": "tier",
    "label": "Tier",
    "type": "select",
    "options": ["free", "basic", "pro", "enterprise"]
  }
]
```

### Validation
```json
{
  "email": [
    {"rule": "email", "message": "Invalid email"}
  ],
  "tier": [
    {"rule": "required", "message": "Tier is required"}
  ]
}
```

### Events (LUA)
```json
{
  "onLoad": "function(form) return form end",
  "onSubmit": "function(form) Customer.SetTier(form.id, form.tier) end"
}
```

---

## API Endpoints

Base path: `/api/v1/ui/forms`

- `POST /ui/forms` - create form
- `GET /ui/forms` - list forms (filter by `entity_type`)
- `GET /ui/forms/:id` - get form by ID
- `PUT /ui/forms/:id` - update form
- `DELETE /ui/forms/:id` - soft delete form

Pagination:
- `page` (default: 1)
- `page_size` (default: 20)

---

## Request/Response Examples

### Create
```json
{
  "form_id": "customer_form",
  "entity_type": "customer",
  "name": "Customer Form",
  "description": "Default customer form",
  "layout": {"type": "single_column"},
  "fields": [{"name": "email", "type": "text"}],
  "is_active": true
}
```

### Update
```json
{
  "entity_type": "customer",
  "name": "Customer Form",
  "description": "Updated",
  "layout": {"type": "two_column"},
  "fields": [{"name": "email", "type": "text"}],
  "is_active": true
}
```

---

## Versioning

- `FormDefinition` uses `BaseAggregate.Version`.
- Each successful update increments version via `IncrementVersion()`.
- Historical versions can be stored in `ui_form_versions` for audit/rollback.

---

## Multi-tenant Support

- `tenant_id` scopes forms to a tenant.
- `NULL` means global/shared form.
- Clients should request tenant-specific forms first, then fall back to shared.

---

## Error Handling

- Validation errors: returned as `BAD_REQUEST` with details.
- System errors: returned as `INTERNAL_ERROR` with generic messages.

---

## Related Docs

- [Phase 3: LUA + UI Metadata](../roadmap/PHASE3_LUA_UI_FOUNDATION.md)
- [Scripting Package](../../pkg/scripting/README.md)
- [Testing Patterns](testing-patterns.md)
