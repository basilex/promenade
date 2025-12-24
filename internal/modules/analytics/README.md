# Analytics Module

**Status:** Free
**Version:** 1.0.0

Analytics, metrics collection, custom reports, and dashboards for Promenade.

---

## Features

### Metrics Collection

- Real-time metric collection (counters, gauges, histograms)
- Automatic aggregation (hourly, daily, weekly)
- User-specific and system-wide metrics
- Custom tags and metadata
- Efficient JSONB storage

### Reports

- Pre-built report templates (user activity, content performance, growth trends)
- Custom report builder
- Multiple export formats (JSON, CSV, PDF, XLSX)
- Scheduled reports with email delivery
- Historical data analysis

### Dashboards

- Custom dashboard builder
- Drag-and-drop widget system
- Multiple visualization types (charts, tables, counters, trends)
- Public and private dashboards
- Real-time updates

---

## Database Schema

All tables use `analytics_` prefix:

| Table                         | Purpose                         |
| ----------------------------- | ------------------------------- |
| `analytics_metrics`           | Raw metric data                 |
| `analytics_metric_aggregates` | Pre-aggregated metrics          |
| `analytics_reports`           | Generated reports               |
| `analytics_report_schedules`  | Scheduled report configurations |
| `analytics_dashboards`        | User dashboards                 |
| `analytics_dashboard_widgets` | Dashboard widgets               |

---

## API Endpoints

### Metrics

```
GET    /api/v1/analytics/metrics              # List metrics
POST   /api/v1/analytics/metrics              # Record metric
GET    /api/v1/analytics/metrics/summary      # Get summary
GET    /api/v1/analytics/metrics/aggregates   # Get aggregated data
```

### Reports

```
GET    /api/v1/analytics/reports              # List reports
POST   /api/v1/analytics/reports              # Create report
GET    /api/v1/analytics/reports/:id          # Get report
DELETE /api/v1/analytics/reports/:id          # Delete report
POST   /api/v1/analytics/reports/:id/export   # Export report
```

### Dashboards

```
GET    /api/v1/analytics/dashboards           # List dashboards
POST   /api/v1/analytics/dashboards           # Create dashboard
GET    /api/v1/analytics/dashboards/:id       # Get dashboard
PUT    /api/v1/analytics/dashboards/:id       # Update dashboard
DELETE /api/v1/analytics/dashboards/:id       # Delete dashboard
POST   /api/v1/analytics/dashboards/:id/widgets # Add widget
```

---

## Permissions

| Permission         | Description               |
| ------------------ | ------------------------- |
| `analytics:view`   | View analytics dashboards |
| `analytics:create` | Create custom reports     |
| `analytics:export` | Export analytics data     |
| `analytics:manage` | Manage analytics settings |

---

## Event Subscriptions

The module listens to events from other modules to collect metrics:

| Event             | Metric Collected         |
| ----------------- | ------------------------ |
| `user.registered` | User growth metrics      |
| `post.created`    | Content creation metrics |
| `comment.created` | Engagement metrics       |
| `post.liked`      | Interaction metrics      |

---

## Configuration

### Metrics

```yaml
analytics:
  metrics:
    enabled: true
    aggregation_interval: "1h" # Aggregate every hour
    retention_days: 90 # Keep raw data for 90 days
```

### Reports

```yaml
analytics:
  reports:
    max_per_user: 10 # Max reports per user
    max_export_rows: 10000 # Max rows in exports
    allowed_formats:
      - "json"
      - "csv"
      - "pdf"
```

### Dashboards

```yaml
analytics:
  dashboards:
    max_per_user: 5 # Max dashboards per user
    max_widgets: 20 # Max widgets per dashboard
```

---

## Monetization Strategy

**Current:** Commercial module (license required)

**Future:** When Audit Log module is released:

- Analytics → Free (basic tier)
- Audit Log → Commercial (enterprise compliance)

This allows broader adoption while monetizing enterprise features.

---

## Testing

Run module tests:

```bash
go test ./internal/modules/analytics/...
```

Run license tests:

```bash
go test ./internal/modules/analytics/license/...
```

---

## Dependencies

**Internal:**

- `pkg/module` - Module system interface
- `pkg/bus` - Event bus for metric collection
- `pkg/uuidv7` - UUID generation

**External:**

- `github.com/jmoiron/sqlx` - Database access
- `github.com/gin-gonic/gin` - HTTP routing
- `github.com/spf13/viper` - Configuration

---

## Migration

Apply migrations:

```bash
make migrate-module MODULE=analytics
```

Rollback migrations:

```bash
make migrate-rollback MODULE=analytics STEPS=1
```

---

## Troubleshooting

### License Issues

**Error:** "license key not provided"

- Set `ANALYTICS_LICENSE_KEY` environment variable

**Error:** "invalid license signature"

- Verify `LICENSE_SECRET` matches the key used during generation
- Check license key format

**Error:** "license has expired"

- License expired beyond grace period
- Obtain new license or extend existing one

### Module Not Loading

1. Check `config/modules.yaml`:

   ```yaml
   modules:
     enabled:
       - analytics
   ```

2. Verify license (if required):
   Testing

Run module tests:

````bash
go test ./internal/modules/analyticsModule Not Loading

1. Check `config/modules.yaml`:

   ```yaml
   modules:
     enabled:
       - analytics
````

## 2

**Author:** Promenade Team
**License:** MIT (Free)
