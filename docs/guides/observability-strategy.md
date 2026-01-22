# Observability Strategy

**Status**: Planned (Phase 2)  
**Priority**: MEDIUM  
**Target Date**: Q2 2026  
**Last Updated**: January 22, 2026

---

## Overview

Comprehensive observability strategy for Promenade Platform covering **logging**, **metrics**, and **distributed tracing**. This document defines our approach to production monitoring and debugging.

---

## Three Pillars of Observability

### 1. Logging ( Implemented)

**Current State**: **Production-ready**

**Implementation**: Structured logging with `log/slog`

```go
import (
    "log/slog"
    "github.com/basilex/promenade/pkg/logger"
)

func (uc *CustomerUseCase) CreateCustomer(ctx context.Context, email string) (*Customer, error) {
    log := logger.FromContext(ctx)

    log.Info("creating customer",
        slog.String("email", email),
        slog.String("operation", "create_customer"),
    )

    customer, err := uc.repo.Create(ctx, email)
    if err != nil {
        log.Error("failed to create customer",
            slog.Any("error", err),
            slog.String("email", email),
        )
        return nil, err
    }

    log.Info("customer created successfully",
        slog.String("customer_id", customer.ID.String()),
    )

    return customer, nil
}
```

**Features**:

- Structured JSON logs
- Context propagation
- Request ID tracking
- Log levels (Debug, Info, Warn, Error)
- Contextual fields (user_id, customer_id, etc.)

**Documentation**: [pkg/logger/README.md](../../pkg/logger/README.md)

---

### 2. Metrics ( Partial)

**Current State**: **Basic implementation**

**Existing Metrics**:

- Health check endpoints (`/health`, `/health/db`, `/health/redis`)
- HTTP status codes (via Gin framework)
- Database connection pool stats (available via introspection)

**Missing**:

- Business metrics (orders created, invoices generated, etc.)
- Performance metrics (request latency, DB query time)
- Resource metrics (memory usage, goroutines, CPU)
- Custom counters/gauges/histograms

**Planned**: Prometheus integration (Phase 2)

---

### 3. Distributed Tracing ( Not Implemented)

**Current State**: **Not implemented**

**Risk**: Difficult to debug cross-context operations in production

**Planned**: OpenTelemetry + Jaeger (Phase 2)

---

## Distributed Tracing (Planned)

### Why Tracing?

**Problem**: Complex request flow across multiple contexts

```
HTTP Request → Identity (Auth)
             → Customer Management (Get Customer)
             → Order Management (Create Order)
             → Warehouse (Reserve Inventory) ← EVENT BUS
             → Fiscal (Generate Receipt) ← EVENT BUS
             → Billing (Create Invoice)
             → Response
```

**Without Tracing**:

- Hard to see complete request flow
- Can't measure latency per service
- Difficult to find bottlenecks
- No cross-context correlation

**With Tracing**:

- Visual request flow diagram
- Latency breakdown per operation
- Error propagation tracking
- Cross-context correlation

---

## Planned Implementation: OpenTelemetry

### Technology Stack

**Tracing Framework**: [OpenTelemetry](https://opentelemetry.io/)  
**Backend**: [Jaeger](https://www.jaegertracing.io/) or [Tempo](https://grafana.com/oss/tempo/)  
**Exporter**: OTLP (OpenTelemetry Protocol)

**Rationale**:

- OpenTelemetry is vendor-neutral standard
- Works with Jaeger, Tempo, DataDog, New Relic
- Go SDK mature and well-documented
- Automatic instrumentation available

---

## Implementation Plan (Phase 2)

### Step 1: Add Dependencies

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/trace
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace
go get go.opentelemetry.io/otel/sdk/trace
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
go get go.opentelemetry.io/contrib/instrumentation/github.com/jmoiron/sqlx/otelsqlx
```

### Step 2: Initialize Tracer

```go
// pkg/telemetry/tracer.go
package telemetry

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func InitTracer(ctx context.Context, serviceName, jaegerEndpoint string) (*sdktrace.TracerProvider, error) {
    // Create OTLP exporter
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(jaegerEndpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    // Create resource with service name
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(serviceName),
            semconv.ServiceVersion("0.1.0"),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Create tracer provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample 100% in dev, adjust for prod
    )

    // Set global tracer provider
    otel.SetTracerProvider(tp)

    return tp, nil
}
```

### Step 3: Instrument HTTP Server

```go
// cmd/api/main.go
import (
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
    // Initialize tracer
    tp, err := telemetry.InitTracer(ctx, "promenade-api", "localhost:4317")
    if err != nil {
        log.Fatal("Failed to initialize tracer", err)
    }
    defer tp.Shutdown(context.Background())

    // Create Gin router
    router := gin.New()

    // Add OpenTelemetry middleware (automatic HTTP tracing)
    router.Use(otelgin.Middleware("promenade-api"))

    // Register routes...
}
```

### Step 4: Instrument Use Cases

```go
// internal/contexts/customer-mgmt/customer/usecase/customer_usecase.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("customer-mgmt")

func (uc *CustomerUseCase) CreateCustomer(ctx context.Context, email string) (*Customer, error) {
    // Start span
    ctx, span := tracer.Start(ctx, "CustomerUseCase.CreateCustomer")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("customer.email", email),
        attribute.String("operation", "create"),
    )

    // Business logic
    customer, err := uc.repo.Create(ctx, aggregate.NewCustomer(email))
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "failed to create customer")
        return nil, err
    }

    // Add result attributes
    span.SetAttributes(
        attribute.String("customer.id", customer.ID.String()),
    )

    span.SetStatus(codes.Ok, "customer created successfully")
    return customer, nil
}
```

### Step 5: Instrument Repositories

```go
// Auto-instrumentation for sqlx
import "go.opentelemetry.io/contrib/instrumentation/github.com/jmoiron/sqlx/otelsqlx"

func NewCustomerRepository(db *sqlx.DB) *CustomerRepository {
    // Wrap DB with tracing
    tracedDB := otelsqlx.Wrap(db, otelsqlx.WithTracerProvider(otel.GetTracerProvider()))

    return &CustomerRepository{db: tracedDB}
}
```

### Step 6: Instrument Event Bus

```go
// pkg/bus/bus.go
func (b *EventBus) Publish(ctx context.Context, topic string, event Event) error {
    ctx, span := tracer.Start(ctx, "EventBus.Publish")
    defer span.End()

    span.SetAttributes(
        attribute.String("event.topic", topic),
        attribute.String("event.type", event.Type),
        attribute.String("event.id", event.ID),
    )

    // Publish event...
}

func (b *EventBus) Subscribe(topic string, handler Handler) {
    wrappedHandler := func(ctx context.Context, event Event) error {
        ctx, span := tracer.Start(ctx, "EventBus.Handler")
        defer span.End()

        span.SetAttributes(
            attribute.String("event.topic", topic),
            attribute.String("event.type", event.Type),
        )

        return handler(ctx, event)
    }

    b.subscribe(topic, wrappedHandler)
}
```

---

## Trace Visualization

### Jaeger UI Example

**Request Flow**:

```
POST /api/v1/orders
 [100ms] HTTP Request (gin)
   [10ms] AuthMiddleware (jwt validation)
   [80ms] OrderHandler.Create
     [5ms] CustomerUseCase.GetCustomer
       [3ms] SELECT FROM customers (sqlx)
     [60ms] OrderUseCase.CreateOrder
       [8ms] INSERT INTO orders (sqlx)
       [40ms] EventBus.Publish (order.created)
         [15ms] WarehouseHandler (reserve inventory)
           [12ms] UPDATE inventory (sqlx)
         [20ms] FiscalHandler (generate receipt)
            [18ms] Checkbox API call
       [5ms] UPDATE order.status
     [7ms] Response serialization
   [3ms] Logging
 [2ms] HTTP Response
```

**Insights**:

- Total request time: 100ms
- Bottleneck: Checkbox API (18ms)
- Database queries efficient (<10ms)
- Event bus adds overhead (40ms)

---

## Metrics Implementation (Phase 2)

### Prometheus Integration

```go
// pkg/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP metrics
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "promenade_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "promenade_http_request_duration_seconds",
            Help:    "HTTP request latency",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    // Business metrics
    ordersCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "promenade_orders_created_total",
            Help: "Total number of orders created",
        },
    )

    invoicesGenerated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "promenade_invoices_generated_total",
            Help: "Total number of invoices generated",
        },
    )

    // Database metrics
    dbConnectionsActive = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "promenade_db_connections_active",
            Help: "Number of active database connections",
        },
    )

    dbQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "promenade_db_query_duration_seconds",
            Help:    "Database query latency",
            Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
        },
        []string{"query_type"},
    )
)

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, endpoint string, status int, duration float64) {
    httpRequestsTotal.WithLabelValues(method, endpoint, fmt.Sprint(status)).Inc()
    httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// RecordOrderCreated increments order counter
func RecordOrderCreated() {
    ordersCreated.Inc()
}
```

### Expose Metrics Endpoint

```go
// cmd/api/server.go
import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) SetupRoutes() {
    // Prometheus metrics endpoint
    s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
```

### Grafana Dashboard

**Metrics to visualize**:

- Request rate (requests/sec)
- Error rate (4xx, 5xx)
- Latency (p50, p95, p99)
- Database query time
- Event bus throughput
- Active connections
- Memory usage
- Goroutine count

---

## Observability Best Practices

### DO

 **Use structured logging** - JSON format, consistent fields  
 **Propagate context** - Pass `ctx` through all layers  
 **Log at appropriate levels** - Debug → Info → Warn → Error  
 **Include correlation IDs** - Track requests across services  
 **Measure critical paths** - Instrument hot code paths  
 **Set SLOs** - Define acceptable latency/error rates  
 **Alert on symptoms** - High error rate, slow response time

### DON'T

 **Don't log sensitive data** - PII, passwords, API keys  
 **Don't over-instrument** - Too many spans = noise  
 **Don't sample too aggressively** - Balance cost vs visibility  
 **Don't forget error handling** - Log errors with context  
 **Don't use magic numbers** - Name metrics clearly  
 **Don't alert on causes** - Alert on user impact

---

## Cost Considerations

### Tracing Costs

**Storage**: ~1KB per trace (compressed)  
**Ingestion**: 1000 traces/sec = ~1GB/day = ~30GB/month  
**Retention**: 7 days (hot), 30 days (cold)

**Sampling Strategy**:

- **Development**: 100% sampling
- **Staging**: 50% sampling
- **Production**: 10% sampling (adjust based on traffic)
- **Always**: Sample errors at 100%

### Metrics Costs

**Storage**: ~1MB per metric/day (1min resolution)  
**Cardinality**: Keep label values <1000 per metric  
**Retention**: 30 days (Prometheus), 1 year (long-term storage)

---

## Migration Path

### Phase 1 (Current - Q1 2026)

 Structured logging (`log/slog`)  
 Request ID tracking  
 Health check endpoints  
 Context propagation

### Phase 2 (Q2 2026)

 OpenTelemetry integration  
 Jaeger deployment  
 Prometheus metrics  
 Grafana dashboards  
 Alert rules (PagerDuty/Opsgenie)

### Phase 3 (Q3 2026)

 Log aggregation (Loki/ELK)  
 Distributed tracing across all contexts  
 Custom business metrics dashboards  
 SLO/SLA monitoring  
 Automated incident response

---

## Configuration

### Environment Variables (Phase 2)

```yaml
# config/app.postgres-prod.yaml
observability:
  tracing:
    enabled: true
    jaeger_endpoint: "jaeger:4317"
    sampling_rate: 0.1 # 10% sampling in production

  metrics:
    enabled: true
    prometheus_port: 9090

  logging:
    level: "info" # debug, info, warn, error
    format: "json"
```

---

## Related Documentation

- [Logger Package](../../pkg/logger/README.md) - Current logging implementation
- [Health Checks](health-checks.md) - Health monitoring endpoints
- [Testing Patterns](testing-patterns.md) - Testing observability code
- [Security Patterns](security-patterns.md) - Don't log sensitive data

---

## External Resources

- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [The Three Pillars of Observability](https://www.oreilly.com/library/view/distributed-systems-observability/9781492033431/ch04.html)

---

**Status**: Planned (Phase 2)  
**Target Date**: Q2 2026  
**Owner**: Platform Team  
**Maintainer**: Promenade Team
