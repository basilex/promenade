# Load Testing

**Status**:  **Phase 2** - Load testing infrastructure for performance validation  
**Tools**: k6, vegeta, hey

---

## Table of Contents

1. [Overview](#overview)
2. [Tools](#tools)
3. [Test Scenarios](#test-scenarios)
4. [Running Load Tests](#running-load-tests)
5. [Analyzing Results](#analyzing-results)
6. [Performance Targets](#performance-targets)

---

## Overview

Load testing validates that Promenade API can handle expected production traffic:

- **Target**: 10,000 transactions per second (TPS)
- **Response Time**: p95 < 200ms, p99 < 500ms
- **Concurrency**: 1,000 concurrent users

**When to Load Test**:

- Before production deployment
- After major refactoring
- When adding new contexts
- To establish performance baselines

---

## Tools

### 1. k6 (Recommended)

**Best for**: Complex scenarios, JavaScript scripting, CI/CD integration

**Install**:

```bash
# macOS
brew install k6

# Ubuntu
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-update && sudo apt-get install k6

# Docker
docker pull grafana/k6:latest
```

### 2. vegeta

**Best for**: Simple HTTP load testing, quick benchmarks

**Install**:

```bash
# macOS
brew install vegeta

# Go
go install github.com/tsenart/vegeta@latest
```

### 3. hey

**Best for**: Quick one-liners, exploratory testing

**Install**:

```bash
# macOS
brew install hey

# Go
go install github.com/rakyll/hey@latest
```

---

## Test Scenarios

### Scenario 1: Create Customer (k6)

**File**: `k6/create-customer.js`

```javascript
import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  stages: [
    { duration: "30s", target: 50 }, // Ramp-up to 50 users
    { duration: "1m", target: 100 }, // Ramp-up to 100 users
    { duration: "3m", target: 100 }, // Stay at 100 users
    { duration: "30s", target: 0 }, // Ramp-down to 0 users
  ],
  thresholds: {
    http_req_duration: ["p(95)<200", "p(99)<500"], // 95% < 200ms, 99% < 500ms
    http_req_failed: ["rate<0.01"], // < 1% errors
  },
};

export default function () {
  const url = "http://localhost:8080/api/v1/customers";

  const payload = JSON.stringify({
    name: `Customer ${__VU}-${__ITER}`,
    email: `customer-${__VU}-${__ITER}@example.com`,
    status: "active",
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(url, payload, params);

  check(res, {
    "status is 201": (r) => r.status === 201,
    "response has id": (r) => JSON.parse(r.body).data.id !== undefined,
  });

  sleep(1); // 1 second pause between iterations
}
```

**Run**:

```bash
k6 run k6/create-customer.js
```

---

### Scenario 2: List Customers (k6)

**File**: `k6/list-customers.js`

```javascript
import http from "k6/http";
import { check } from "k6";

export let options = {
  vus: 100, // 100 virtual users
  duration: "30s", // 30 seconds
  thresholds: {
    http_req_duration: ["p(95)<100", "p(99)<200"],
  },
};

export default function () {
  const res = http.get("http://localhost:8080/api/v1/customers?page=1&page_size=20");

  check(res, {
    "status is 200": (r) => r.status === 200,
    "has pagination": (r) => JSON.parse(r.body).data.pagination !== undefined,
  });
}
```

**Run**:

```bash
k6 run k6/list-customers.js
```

---

### Scenario 3: Order Creation Workflow (k6)

**File**: `k6/order-workflow.js`

```javascript
import http from "k6/http";
import { check, group } from "k6";

export let options = {
  vus: 50,
  duration: "2m",
  thresholds: {
    http_req_duration: ["p(95)<500"],
  },
};

const BASE_URL = "http://localhost:8080/api/v1";

export default function () {
  let customerID, orderID;

  // Step 1: Create Customer
  group("Create Customer", function () {
    const res = http.post(
      `${BASE_URL}/customers`,
      JSON.stringify({
        name: `Customer ${__VU}-${__ITER}`,
        email: `customer-${__VU}-${__ITER}@example.com`,
        status: "active",
      }),
      {
        headers: { "Content-Type": "application/json" },
      },
    );

    check(res, { "customer created": (r) => r.status === 201 });
    customerID = JSON.parse(res.body).data.id;
  });

  // Step 2: Create Order
  group("Create Order", function () {
    const res = http.post(
      `${BASE_URL}/orders`,
      JSON.stringify({
        customer_id: customerID,
        currency_code: "USD",
        items: [
          {
            product_id: "01932e9a-5678-7abc-9def-0123456789cd",
            quantity: 5,
            unit_price: 99.99,
          },
        ],
      }),
      {
        headers: { "Content-Type": "application/json" },
      },
    );

    check(res, { "order created": (r) => r.status === 201 });
    orderID = JSON.parse(res.body).data.id;
  });

  // Step 3: Confirm Order (trigger saga)
  group("Confirm Order", function () {
    const res = http.post(`${BASE_URL}/orders/${orderID}/confirm`);
    check(res, { "order confirmed": (r) => r.status === 200 });
  });
}
```

**Run**:

```bash
k6 run k6/order-workflow.js
```

---

### Scenario 4: Simple HTTP Load (vegeta)

**File**: `vegeta/targets.txt`

```
GET http://localhost:8080/api/v1/customers

GET http://localhost:8080/api/v1/orders

POST http://localhost:8080/api/v1/customers
Content-Type: application/json
@vegeta/create-customer.json
```

**File**: `vegeta/create-customer.json`

```json
{
  "name": "Load Test Customer",
  "email": "loadtest@example.com",
  "status": "active"
}
```

**Run**:

```bash
# 100 requests per second for 30 seconds
vegeta attack -targets=vegeta/targets.txt -rate=100 -duration=30s | vegeta report
```

---

### Scenario 5: Quick Load Test (hey)

**Command**:

```bash
# 10,000 requests with 50 concurrent workers
hey -n 10000 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","status":"active"}' \
  http://localhost:8080/api/v1/customers
```

---

## Running Load Tests

### Prerequisites

1. **Start development server**:

   ```bash
   make dev
   ```

2. **Seed test data** (optional):

   ```bash
   make seed
   ```

3. **Monitor resources**:

   ```bash
   # Terminal 1: API logs
   docker logs -f promenade-api

   # Terminal 2: Database metrics
   docker stats promenade-postgres-dev

   # Terminal 3: Redis metrics
   docker stats promenade-redis-dev
   ```

### Run k6 Tests

```bash
# Single scenario
k6 run k6/create-customer.js

# With custom VUs and duration
k6 run --vus 200 --duration 5m k6/list-customers.js

# Output to JSON
k6 run --out json=results.json k6/order-workflow.js

# Run in Docker
docker run -i grafana/k6 run - <k6/create-customer.js
```

### Run vegeta Tests

```bash
# Generate load
vegeta attack -targets=vegeta/targets.txt -rate=100 -duration=30s > results.bin

# Generate report
vegeta report results.bin

# Plot results
vegeta plot results.bin > plot.html
```

### Run hey Tests

```bash
# Quick test
hey -n 1000 -c 10 http://localhost:8080/api/v1/customers

# Detailed report
hey -n 10000 -c 50 -m GET http://localhost:8080/api/v1/customers > hey-report.txt
```

---

## Analyzing Results

### k6 Output

```
     execution: local
        script: k6/create-customer.js
        output: -

     scenarios: (100.00%) 1 scenario, 100 max VUs, 5m30s max duration

     data_received..................: 1.2 MB  10 kB/s
     data_sent......................: 980 kB  8.2 kB/s
     http_req_blocked...............: avg=1.23ms   min=2µs      med=7µs      max=123ms   p(90)=15µs    p(95)=22µs
     http_req_connecting............: avg=1.15ms   min=0s       med=0s       max=119ms   p(90)=0s      p(95)=0s
     http_req_duration..............: avg=85.34ms  min=23.45ms  med=78.12ms  max=456ms   p(90)=145ms   p(95)=187ms
       { expected_response:true }...: avg=85.34ms  min=23.45ms  med=78.12ms  max=456ms   p(90)=145ms   p(95)=187ms
     http_req_failed................: 0.00%    0         10000
     http_req_receiving.............: avg=124µs    min=18µs     med=89µs     max=2.45ms  p(90)=234µs   p(95)=345µs
     http_req_sending...............: avg=45µs     min=8µs      med=32µs     max=1.12ms  p(90)=78µs    p(95)=112µs
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s       max=0s      p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=85.17ms  min=23.41ms  med=77.98ms  max=455ms   p(90)=144ms   p(95)=186ms
     http_reqs......................: 10000   83.33/s
     iteration_duration.............: avg=1.19s    min=1.02s    med=1.18s    max=1.56s   p(90)=1.24s   p(95)=1.28s
     iterations.....................: 10000   83.33/s
     vus............................: 100     min=0      max=100
```

**Key Metrics**:

- `http_req_duration`: 85ms average, 187ms p95  (target: <200ms)
- `http_req_failed`: 0% errors 
- `http_reqs`: 83 TPS (target: 10,000 TPS  needs scaling)

---

### vegeta Output

```
Requests      [total, rate, throughput]  3000, 100.03, 100.01
Duration      [total, attack, wait]      29.99s, 29.98s, 12.34ms
Latencies     [mean, 50, 95, 99, max]    95.67ms, 87.23ms, 178.45ms, 234.56ms, 456.78ms
Bytes In      [total, mean]              1234567, 411.52
Bytes Out     [total, mean]              987654, 329.22
Success       [ratio]                    99.97%
Status Codes  [code:count]               200:2999  201:1
```

---

## Performance Targets

### Response Times

| Endpoint                   | p50    | p95    | p99    | Max    |
| -------------------------- | ------ | ------ | ------ | ------ |
| `POST /customers`          | <50ms  | <150ms | <300ms | <500ms |
| `GET /customers`           | <30ms  | <100ms | <200ms | <300ms |
| `POST /orders`             | <100ms | <300ms | <600ms | <1s    |
| `POST /orders/:id/confirm` | <200ms | <500ms | <1s    | <2s    |

### Throughput

| Scenario        | Target TPS | Current | Status                      |
| --------------- | ---------- | ------- | --------------------------- |
| Create Customer | 1,000      | 83      |  Needs optimization       |
| List Customers  | 5,000      | 450     |  Needs caching            |
| Order Workflow  | 500        | 35      |  Saga optimization needed |

### Resource Limits

| Resource          | Dev     | Prod    |
| ----------------- | ------- | ------- |
| API CPU           | 2 cores | 8 cores |
| API Memory        | 1GB     | 4GB     |
| PostgreSQL CPU    | 2 cores | 8 cores |
| PostgreSQL Memory | 2GB     | 16GB    |
| Redis Memory      | 256MB   | 2GB     |

---

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/load-test.yml
name: Load Test

on:
  push:
    branches: [main, dev]

jobs:
  load-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: passw0rd
        ports:
          - 5432:5432
      redis:
        image: redis:7
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v3

      - name: Install k6
        run: |
          sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update && sudo apt-get install k6

      - name: Build and Run API
        run: |
          make build
          ./bin/promenade &
          sleep 5

      - name: Run Load Test
        run: k6 run test/load/k6/create-customer.js
```

---

## Related Documentation

- [Benchmark README](../benchmark/README.md) - Unit benchmarks
- [Profiling Guide](../../docs/guides/profiling.md) - CPU/memory profiling
- [Observability Strategy](../../docs/guides/observability-strategy.md) - Metrics & tracing

---

## Next Steps

- [ ] **Phase 2A**: Establish baselines for all contexts (customer-mgmt, order-mgmt, billing, etc.)
- [ ] **Phase 2B**: Add k6 scenarios for fulfillment saga (order workflow)
- [ ] **Phase 2C**: Integrate load testing in CI/CD (regression detection)
- [ ] **Phase 3**: Production load testing (staging environment)
