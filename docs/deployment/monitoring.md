# Monitoring & Observability

**Status**: **Phase 2** - Planned for Q3 2026  
**Stack**: Prometheus + Grafana + Jaeger (OpenTelemetry)

---

## Overview

Promenade observability stack (Three Pillars):

- **Logs**: Structured logging (JSON, zerolog)
- **Metrics**: Prometheus + Grafana dashboards
- **Traces**: OpenTelemetry + Jaeger (distributed tracing)

See: [Observability Strategy](../guides/observability-strategy.md)

---

## Architecture

**Observability Stack**:

| Component     | Purpose         | Endpoint   | Data Source                      |
| ------------- | --------------- | ---------- | -------------------------------- |
| Promenade API | Application     | `/metrics` | Prometheus metrics (HTTP)        |
| Promenade API | Logs            | Stdout     | JSON structured logs             |
| Promenade API | Traces          | Port 4317  | OpenTelemetry (OTLP)             |
| Prometheus    | Metrics storage | Port 9090  | Scrapes `/metrics`               |
| Loki          | Log aggregation | Port 3100  | Collects JSON logs               |
| Jaeger        | Trace storage   | Port 16686 | Receives OTLP traces             |
| Grafana       | Visualization   | Port 3000  | Queries Prometheus, Loki, Jaeger |

**Data Flow**:

1. API exposes `/metrics` → Prometheus scrapes every 15s
2. API writes JSON logs → Loki collects
3. API sends OTLP traces → Jaeger receives
4. Grafana queries all three datasources for dashboards

---

## Quick Start (Docker Compose)

### 1. Start Monitoring Stack

```bash
# docker/docker-compose.monitoring.yml (Phase 2)
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # Jaeger UI
      - "14268:14268"  # HTTP collector
      - "4317:4317"    # OTLP gRPC
    environment:
      - COLLECTOR_OTLP_ENABLED=true

volumes:
  prometheus_data:
  grafana_data:
```

### 2. Start Services

```bash
docker compose -f docker/docker-compose.monitoring.yml up -d

# Access UIs
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
# Jaeger: http://localhost:16686
```

---

## Prometheus Metrics

### Expose Metrics Endpoint

```go
// cmd/api/server.go
import "github.com/prometheus/client_golang/prometheus/promhttp"

func setupRouter(config *config.Config) *gin.Engine {
    router := gin.Default()

    // Prometheus metrics endpoint
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // API routes
    // ...

    return router
}
```

### Custom Metrics

```go
// pkg/metrics/metrics.go
import "github.com/prometheus/client_golang/prometheus"

var (
    // HTTP request duration
    HttpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path", "status"},
    )

    // Database query duration
    DbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
        },
        []string{"operation", "table"},
    )

    // Active connections
    ActiveConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "Number of active database connections",
        },
    )
)

func init() {
    prometheus.MustRegister(HttpRequestDuration)
    prometheus.MustRegister(DbQueryDuration)
    prometheus.MustRegister(ActiveConnections)
}
```

### Middleware

```go
// pkg/middleware/prometheus.go
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        c.Next()

        duration := time.Since(start).Seconds()
        status := fmt.Sprintf("%d", c.Writer.Status())

        metrics.HttpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Observe(duration)
    }
}
```

---

## Prometheus Configuration

### prometheus.yml

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: "promenade-api"
    static_configs:
      - targets: ["host.docker.internal:8080"]
    metrics_path: /metrics

  - job_name: "postgres"
    static_configs:
      - targets: ["postgres-exporter:9187"]

  - job_name: "redis"
    static_configs:
      - targets: ["redis-exporter:9121"]
```

---

## Grafana Dashboards

### API Performance Dashboard (JSON)

```json
{
  "dashboard": {
    "title": "Promenade API Performance",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(http_request_duration_seconds_count[5m])"
          }
        ]
      },
      {
        "title": "Response Time (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(http_request_duration_seconds_count{status=~\"5..\"}[5m])"
          }
        ]
      },
      {
        "title": "Database Query Duration (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m]))"
          }
        ]
      }
    ]
  }
}
```

---

## OpenTelemetry Tracing

### Enable Tracing (Phase 2)

```go
// pkg/tracing/tracing.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer(serviceName string, jaegerEndpoint string) error {
    exporter, err := otlptracegrpc.New(
        context.Background(),
        otlptracegrpc.WithEndpoint(jaegerEndpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithSampler(trace.AlwaysSample()),
    )

    otel.SetTracerProvider(tp)
    return nil
}
```

### Instrument HTTP Handlers

```go
// pkg/middleware/tracing.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func TracingMiddleware() gin.HandlerFunc {
    tracer := otel.Tracer("promenade-api")

    return func(c *gin.Context) {
        ctx, span := tracer.Start(c.Request.Context(), c.FullPath())
        defer span.End()

        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
        )

        c.Request = c.Request.WithContext(ctx)
        c.Next()

        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
        )
    }
}
```

---

## Alerting

### Prometheus Alerts (alerts.yml)

```yaml
# monitoring/alerts.yml
groups:
  - name: promenade-alerts
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: rate(http_request_duration_seconds_count{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }}% (threshold: 5%)"

      - alert: SlowResponses
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Slow API responses detected"
          description: "p95 response time is {{ $value }}s (threshold: 0.5s)"

      - alert: DatabaseConnectionPoolExhausted
        expr: active_connections > 180
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool nearly exhausted"
          description: "Active connections: {{ $value }} (limit: 200)"
```

---

## Log Aggregation (Loki)

### Loki Configuration

```yaml
# monitoring/loki.yml
auth_enabled: false

server:
  http_listen_port: 3100

ingester:
  lifecycler:
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1

schema_config:
  configs:
    - from: 2024-01-01
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/index
    cache_location: /loki/cache
  filesystem:
    directory: /loki/chunks
```

### Query Logs in Grafana

```
# Find errors
{job="promenade-api"} |= "error"

# Find slow queries
{job="promenade-api"} | json | duration > 1s

# Find customer-related logs
{job="promenade-api"} |= "customer" | json | customer_id="01932e8f-1234-7abc-9def-0123456789ab"
```

---

## Monitoring Checklist

### Application Metrics

- [ ] HTTP request rate, duration, errors
- [ ] Database query duration, connection pool
- [ ] Redis cache hit rate, latency
- [ ] Event bus publish rate, errors
- [ ] Saga execution duration, compensation rate

### Infrastructure Metrics

- [ ] CPU usage (API, PostgreSQL, Redis)
- [ ] Memory usage (RSS, heap)
- [ ] Disk I/O (database)
- [ ] Network I/O (API)

### Business Metrics

- [ ] Customer creation rate
- [ ] Order fulfillment time
- [ ] Invoice generation rate
- [ ] Payment success rate

---

## Related Documentation

- [Observability Strategy](../guides/observability-strategy.md) - Phase 2 plan
- [Profiling Guide](../guides/profiling.md) - CPU/memory profiling
- [Load Testing Guide](../../test/load/README.md) - Performance testing

---

## Next Steps

- [ ] **Phase 2A**: Implement Prometheus metrics
- [ ] **Phase 2B**: Create Grafana dashboards
- [ ] **Phase 2C**: Integrate OpenTelemetry + Jaeger
- [ ] **Phase 3**: Add alerting (PagerDuty, Slack)
