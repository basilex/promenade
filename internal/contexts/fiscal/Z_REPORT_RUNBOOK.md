# Daily Z-Report Cadence & Monitoring

**Purpose**: Ensure fiscal shifts close daily and Z-reports are captured reliably.

**Owner**: Fiscal context
**Scope**: Scheduler jobs + operational monitoring

---

## Cadence (Default)

- **Shift open**: `0 9 * * *` (09:00 local)
- **Shift close**: `0 23 * * *` (23:00 local)

These defaults are defined in the shift scheduler integration and can be overridden via configuration when wiring the jobs.

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
