# Fiscal Ops Guide

**Scope**: Production/staging operational checklist for fiscal shift cadence, Z-report capture, and Checkbox integration readiness.

---

## Current Situation (January 2026)

- **Sandbox validation**: Blocked (no sandbox API key yet).
- **Shift jobs registration**: Implemented; depends on config at startup.
- **Log verification**: Pending access to prod/staging logs.

---

## Prerequisites (Prod/Staging)

- `scheduler.enabled = true`
- `fiscal.checkbox.api_key` configured
- `fiscal.checkbox.sandbox = false` for prod; staging may use `true` if sandbox
- `fiscal.shift_open_cron` and `fiscal.shift_close_cron` set

## Expected Log Signals

On startup:
- "Scheduler initialized"
- Shift jobs registered if Checkbox client configured

If missing key:
- "Shift jobs not registered (checkbox client not configured)"

## Verification Checklist (When Log Access Available)

- [ ] Confirm `CHECKBOX_API_KEY` is set in prod/staging environment
- [ ] Confirm scheduler starts and registers `fiscal_shift_open` and `fiscal_shift_close`
- [ ] Confirm no log entry indicates missing checkbox client
- [ ] Confirm at least one register has `ShiftOpenedAt` after open window
- [ ] Confirm `LastZReportAt` updates after close window

## Operational Queries (Postgres)

```sql
-- Registers with open shifts after expected close window
SELECT id, active_shift_id, shift_opened_at, shift_closed_at
FROM cash_registers
WHERE active_shift_id <> ''
  AND (shift_closed_at IS NULL OR shift_closed_at < NOW() - INTERVAL '12 hours');

-- Registers missing Z-report after close window
SELECT id, last_z_report_id, last_z_report_at
FROM cash_registers
WHERE last_z_report_at IS NULL
   OR last_z_report_at < NOW() - INTERVAL '24 hours';
```

## Next Actions

1. Obtain sandbox API key and complete [internal/contexts/fiscal/SANDBOX_CHECKLIST.md](../../internal/contexts/fiscal/SANDBOX_CHECKLIST.md).
2. Verify prod/staging logs for job registration.
3. Enable alerting based on Z-report cadence signals.

---

**Owner**: Fiscal context
