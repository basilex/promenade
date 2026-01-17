# Daily Z-Report Cadence & Monitoring

**Purpose**: Ensure fiscal shifts close daily and Z-reports are captured reliably.

**Owner**: Fiscal context
**Scope**: Scheduler jobs + operational monitoring

---

## Cadence (Default)

- **Shift open**: `0 9 * * *` (09:00 local)
- **Shift close**: `0 23 * * *` (23:00 local)

These defaults are defined in the shift scheduler integration and can be overridden via configuration when wiring the jobs.

## Job Registration (Prod/Staging)

Registration happens on app startup when:
- `scheduler.enabled = true`
- `fiscal.checkbox.api_key` is configured

Expected log signals:
- "Scheduler initialized"
- **If configured**: shift jobs registered with `fiscal_shift_open` and `fiscal_shift_close`
- **If missing key**: "Shift jobs not registered (checkbox client not configured)"

Checklist:
- [ ] `CHECKBOX_API_KEY` set in prod/staging
- [ ] `fiscal.checkbox.sandbox = false` in prod; staging may be `true` if sandbox
- [ ] `scheduler.enabled = true`

## Jobs

- `fiscal_shift_open`
- `fiscal_shift_close`

## Expected Outcomes

- Active registers have `ActiveShiftID` after open
- Closed shifts store `LastZReportID` + `LastZReportAt`
- No open shift remains after the close window

## Monitoring Signals

- Scheduler job execution records (success/failure)
- Logs:
  - "Failed to open shift"
  - "Failed to close shift"
  - "Failed to persist shift close"
- Cash register state:
  - `ActiveShiftID` still set after close window
  - `LastZReportAt` missing or stale

## Validation Checklist (Ops)

- [ ] Scheduler has `fiscal_shift_open` and `fiscal_shift_close` jobs registered
- [ ] At least one register shows `ShiftOpenedAt` after open window
- [ ] `LastZReportAt` updated after close window
- [ ] No register has `ActiveShiftID` set after close window

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

## Alerting Triggers (Suggested)

- No successful `fiscal_shift_close` execution by $T_{close}+1h$
- Any register with `ActiveShiftID` set after close window
- Missing `LastZReportID` after close job success

## Manual Recovery (If Needed)

1. Verify job execution results in scheduler logs.
2. Check cash register state for active shift and last Z-report.
3. Re-run close job manually via scheduler tooling (admin-only).
4. If Checkbox API failure, retry after cooldown and record incident.

---

**Status**: Draft (operational validation pending)
