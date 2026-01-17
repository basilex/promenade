# Checkbox Sandbox Validation Checklist

**Purpose**: Validate the end-to-end fiscal flow against Checkbox sandbox before production enablement.

**Owner**: Fiscal context
**Scope**: Cash register, receipt print/cancel, shift open/close, Z-report confirmation

---

## Prerequisites

- [ ] Checkbox sandbox API key configured (`CHECKBOX_API_KEY`)
- [ ] Sandbox mode enabled (`CHECKBOX_SANDBOX=true`)
- [ ] Timeout set (`CHECKBOX_TIMEOUT`, e.g. `30s`)
- [ ] Fiscal routes available (`/api/v1/fiscal/...`)

## Credential Validation

- [ ] `ValidateCredentials` succeeds (Checkbox `/cashier/me`)
- [ ] Invalid credentials return clean error (no sensitive details)

## Cash Register Lifecycle (API)

- [ ] Create cash register (`POST /api/v1/fiscal/cash-registers`)
- [ ] Activate register (`POST /api/v1/fiscal/cash-registers/:id/activate`)
- [ ] Deactivate register (`POST /api/v1/fiscal/cash-registers/:id/deactivate`)
- [ ] Sync register (`POST /api/v1/fiscal/cash-registers/:id/sync`)

## Receipt Flow (API + Checkbox)

- [ ] Create receipt (`POST /api/v1/fiscal/receipts`)
- [ ] Print receipt (`POST /api/v1/fiscal/receipts/:id/print`) → Checkbox receipt created
- [ ] Receipt persisted with `ProviderReceiptID`, `FiscalURL`, `QRCode`
- [ ] Cancel receipt (`POST /api/v1/fiscal/receipts/:id/cancel`)
- [ ] Cancel rejection cases: already cancelled, missing reason

## Shift Automation (Scheduler + Checkbox)

- [ ] `fiscal_shift_open` job opens shifts for active registers
- [ ] `fiscal_shift_close` job closes shifts and returns Z-report
- [ ] `ActiveShiftID`, `ShiftOpenedAt`, `ShiftClosedAt` set correctly
- [ ] `LastZReportID`, `LastZReportAt` stored after close

## Edge Cases (Sandbox)

- [ ] Print already printed receipt → `BAD_REQUEST`
- [ ] Print cancelled receipt → `BAD_REQUEST`
- [ ] Cancel already cancelled receipt → `BAD_REQUEST`
- [ ] Shift close with no active shift is skipped safely
- [ ] Missing provider cash register ID is logged and skipped

## Evidence

- [ ] Capture sandbox receipt `fiscal_url`
- [ ] Capture QR code URL for verification
- [ ] Capture Z-report payload on shift close

---

**Status**: Blocked (sandbox API key missing)
