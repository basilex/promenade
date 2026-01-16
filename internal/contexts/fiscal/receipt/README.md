# Receipt Aggregate

Fiscal receipt management for the Fiscal context. This aggregate models receipt lifecycle, totals, and fiscal metadata returned by PRRO providers.

---

## Overview

The Receipt aggregate ties an order to a fiscal cash register and tracks the fiscal output (fiscal number, URL, QR code) once printed.

---

## Entity Fields

- **CashRegisterID**: Register that issues the receipt.
- **OrderID**: Order linked to the receipt.
- **PaymentType**: Payment method.
- **ReceiptType**: Receipt operation type.
- **Currency**: ISO currency code.
- **TotalAmount**: Total in minor units.
- **TaxAmount**: Total tax in minor units.
- **FiscalNumber**: Provider fiscal number once printed.
- **FiscalURL**: Provider receipt URL.
- **QRCode**: Base64 QR code data.
- **PrintedAt** / **CancelledAt**: Lifecycle timestamps.
- **CancellationReason**: Required when cancelled.
- **Lines**: Receipt line items stored as `jsonstore.Field[[]ReceiptLine]`.
- **CreatedBy** / **LastUpdatedBy**: Audit fields.
- **Status**: Receipt lifecycle state.

### Status Values

- `pending`
- `printed`
- `cancelled`

### Payment Types

- `cash`
- `card`
- `cashless`

### Receipt Types

- `sale`
- `return`
- `service_in`
- `service_out`

---

## Core Behaviors

- `NewReceipt(...)` validates input, calculates totals, and initializes the receipt.
- `MarkPrinted(...)` stores fiscal metadata and moves to `printed`.
- `Cancel(...)` requires a reason and moves to `cancelled`.
- `Validate()` ensures payment/receipt types and line items are valid.

---

## Receipt Lines

Each line item includes:

- `Name`
- `Quantity`
- `PriceCents`
- `TaxRate`
- Calculated `TotalCents` and `TaxAmountCents`

Totals are computed during creation and validation.

---

## Use Case Operations

The use case supports:

- Create, get, and list receipts
- Get by order ID
- Mark printed (set fiscal data)
- Cancel receipt
- Soft delete

---

## Validation Rules

- `CashRegisterID`, `OrderID`, `Currency`, and `CreatedBy` are required.
- `PaymentType` and `ReceiptType` must be valid enums.
- At least one line item is required; each line must have name, positive quantity, non-negative price, and tax rate $0–100$.
- `FiscalNumber` is required when marking as printed.
- Cancellation requires a reason.

---

## Related Documentation

- [Fiscal Context Overview](../README.md)
- [Cash Register Aggregate](../cashregister/README.md)
