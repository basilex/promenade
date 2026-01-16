# Cash Register Aggregate

Cash register management for the Fiscal context. This aggregate models a PRRO register, its lifecycle, and operational state.

---

## Overview

The Cash Register aggregate tracks registration data, operational status, and synchronization metadata. It follows the BaseAggregate pattern and uses UUID v7 identifiers.

---

## Entity Fields

- **OrganizationID**: Owning organization identifier.
- **FiscalNumber**: PRRO fiscal number assigned by the provider.
- **Model**: Cash register model identifier.
- **Status**: Operational status.
- **LicenseKey**: Provider license key (stored encrypted at rest in higher layers).
- **LastSyncAt**: Last synchronization timestamp.
- **LastUpdatedBy**: User who last modified the aggregate.

### Status Values

- `inactive`
- `active`
- `maintenance`
- `suspended`

---

## Core Behaviors

- `NewCashRegister(...)` validates required identifiers and fields.
- `Activate(...)` requires a license key and moves to `active`.
- `Deactivate(...)` returns to `inactive`.
- `Suspend(...)` forces `suspended`.
- `SetMaintenance(...)` switches to `maintenance`.
- `UpdateLastSync(...)` records a sync timestamp.
- `Validate()` enforces required fields.

---

## Use Case Operations

The use case orchestrates repository access and lifecycle transitions:

- Create, get, and list cash registers
- Activate and deactivate
- Sync last synchronization timestamp
- Update and soft-delete

---

## Repository Operations

The repository exposes common data access methods:

- `Create`
- `GetByID`
- `GetByFiscalNumber`
- `GetByLocation`
- `ListActive`
- `List` (with filters)
- `Update`
- `Delete` (soft delete)

---

## Validation Rules

- `OrganizationID` is required.
- `FiscalNumber` is required and unique.
- `Model` is required.
- `CreatedBy` is required for creation.

---

## Related Documentation

- [Fiscal Context Overview](../README.md)
- [Receipt Aggregate](../receipt/README.md)
