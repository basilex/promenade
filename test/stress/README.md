# Stress & Load Tests

Performance and load testing for Promenade API using **wrk** - a modern HTTP benchmarking tool.

## What Are Stress Tests?

Stress tests push the API to its limits to find:

- 🎯 **Maximum throughput** (requests per second)
- 🎯 **Latency under load** (p50, p95, p99)
- 🎯 **Breaking points** (when does it fail?)
- 🎯 **Resource usage** (CPU, memory, connections)

## Prerequisites

### Install wrk

**macOS:**

```bash
brew install wrk
```

**Linux (Ubuntu/Debian):**

```bash
sudo apt-get install wrk
```

**From source:**

```bash
git clone https://github.com/wg/wrk.git
cd wrk
make
sudo cp wrk /usr/local/bin/
```

Verify installation:

```bash
wrk --version
```

## Quick Start

```bash
# 1. Start API
make dev

# 2. Run stress tests (in another terminal)
cd test/stress
./stress_test.sh

# Or use Makefile
make stress-test
```

## Test Scenarios

### 1. Health Check Load Test

**Goal:** Baseline performance, minimal logic

```bash
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
```

**Expected Results:**

- RPS: 10,000+
- Latency (p99): < 50ms
- Zero errors

### 2. Authentication Load Test

**Goal:** Test auth flow under load

```bash
wrk -t4 -c100 -d30s -s scenarios/auth.lua http://localhost:8081
```

**Expected Results:**

- RPS: 500-1,000
- Latency (p99): < 200ms
- Error rate: < 1%

### 3. Posts CRUD Load Test

**Goal:** Database-heavy operations

```bash
wrk -t4 -c100 -d30s -s scenarios/posts.lua http://localhost:8081
```

**Expected Results:**

- RPS: 300-500
- Latency (p99): < 300ms
- Error rate: < 1%

### 4. Notifications Load Test

**Goal:** Test notification delivery and preferences under load

```bash
wrk -t4 -c100 -d30s -s scenarios/notifications.lua http://localhost:8081
```

**Expected Results:**

- RPS: 400-600
- Latency (p99): < 250ms
- Error rate: < 1%

### 5. Concurrent Users Simulation

**Goal:** Realistic user behavior

```bash
wrk -t8 -c200 -d60s -s scenarios/user_journey.lua http://localhost:8081
```

**Expected Results:**

- Concurrent users: 200
- Session duration: 60s
- Realistic think time between requests

## Load Levels

### Light Load (Warm-up)

```bash
wrk -t2 -c10 -d10s http://localhost:8081/api/v1/health
```

- 2 threads, 10 connections, 10 seconds
- Goal: ~1,000 RPS

### Medium Load

```bash
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
```

- 4 threads, 100 connections, 30 seconds
- Goal: ~10,000 RPS

### Heavy Load

```bash
wrk -t8 -c500 -d60s http://localhost:8081/api/v1/health
```

- 8 threads, 500 connections, 60 seconds
- Goal: Find limits

### Stress Test (Find Breaking Point)

```bash
wrk -t12 -c1000 -d120s http://localhost:8081/api/v1/health
```

- 12 threads, 1000 connections, 2 minutes
- Goal: Break the system, find bottlenecks

## wrk Parameters Explained

```bash
wrk -t4 -c100 -d30s -s script.lua http://localhost:8081/api/v1/endpoint
    ↑    ↑    ↑      ↑
    │    │    │      └─ Lua script for complex scenarios
    │    │    └─ Duration (10s, 30s, 1m, 2h)
    │    └─ Connections (concurrent requests)
    └─ Threads (CPU cores to use)
```

**Recommended:**

- **Threads (-t):** Number of CPU cores (2-8)
- **Connections (-c):** 10-1000 (start low, increase)
- **Duration (-d):** 10s-60s (longer for production tests)

## Lua Scripts

### Basic POST Request

```lua
-- scenarios/auth_login.lua
wrk.method = "POST"
wrk.body   = '{"email":"test@example.com","password":"password123"}'
wrk.headers["Content-Type"] = "application/json"
```

### Dynamic Requests with State

```lua
-- scenarios/posts_crud.lua
local counter = 0
local token = "Bearer YOUR_TOKEN_HERE"

request = function()
    counter = counter + 1
    local path = "/api/v1/posts"
    local body = string.format('{"title":"Post %d","content":"Test"}', counter)

    return wrk.format("POST", path, {
        ["Content-Type"] = "application/json",
        ["Authorization"] = token
    }, body)
end

response = function(status, headers, body)
    if status ~= 201 then
        print("Error: " .. status .. " - " .. body)
    end
end
```

## Interpreting Results

### Sample Output

```
Running 30s test @ http://localhost:8081/api/v1/health
  4 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.12ms    2.34ms  45.23ms   89.34%
    Req/Sec     4.89k   456.23    5.67k    76.45%
  586234 requests in 30.00s, 123.45MB read
Requests/sec:  19541.13
Transfer/sec:      4.11MB
```

**Key Metrics:**

- **Latency Avg:** Mean response time (lower is better)
- **Latency Stdev:** Consistency (lower is better)
- **Latency Max:** Worst case (watch for spikes)
- **Req/Sec:** Throughput per thread
- **Requests/sec:** Total throughput (RPS)
- **Transfer/sec:** Network bandwidth used

### What's Good?

| Metric              | Good     | Warning  | Critical |
| ------------------- | -------- | -------- | -------- |
| **Health endpoint** | >10k RPS | <5k RPS  | <1k RPS  |
| **Auth endpoints**  | >500 RPS | <200 RPS | <100 RPS |
| **CRUD endpoints**  | >300 RPS | <100 RPS | <50 RPS  |
| **p99 Latency**     | <100ms   | <500ms   | >1s      |
| **Error rate**      | 0%       | <1%      | >5%      |

### Red Flags 🚨

- **High Stdev:** Inconsistent performance (investigate)
- **Max >> Avg:** Occasional very slow requests (caching issue?)
- **Errors:** Database connection pool exhausted?
- **Linear degradation:** Not scaling with connections

## Monitoring During Tests

### Terminal 1: Run API

```bash
make dev
```

### Terminal 2: Monitor Resources

```bash
# Watch CPU/Memory
watch -n 1 'ps aux | grep promenade | grep -v grep'

# Or use htop
htop -p $(pgrep promenade)
```

### Terminal 3: Monitor Database

```bash
# PostgreSQL connections
watch -n 1 'psql -U promenade -d promenade_dev -c "SELECT count(*) FROM pg_stat_activity;"'

# Database load
watch -n 1 'psql -U promenade -d promenade_dev -c "SELECT * FROM pg_stat_database WHERE datname = '\''promenade_dev'\'';"'
```

### Terminal 4: Run wrk

```bash
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
```

## Optimization Tips

### If RPS is Low

1. **Check database connection pool:**

   ```yaml
   # config/app.dev.yaml
   database:
     max_open_conns: 100 # Increase
     max_idle_conns: 25 # Increase
   ```

2. **Enable HTTP/2:**

   ```go
   // May need HTTP/2 support
   ```

3. **Add indexes:**

   ```sql
   CREATE INDEX idx_posts_user_id ON user_posts(user_id);
   ```

4. **Use caching:**
   - Redis for session cache
   - In-memory cache for reference data

### If Latency is High

1. **Profile the code:**

   ```bash
   go tool pprof http://localhost:8081/debug/pprof/profile
   ```

2. **Check slow queries:**

   ```sql
   SELECT query, mean_exec_time, calls
   FROM pg_stat_statements
   ORDER BY mean_exec_time DESC
   LIMIT 10;
   ```

3. **Optimize JSON serialization:**
   - Use `json.Marshal` efficiently
   - Avoid reflection where possible

### If Errors Occur

1. **Database connection limit reached:**

   ```
   Error: pq: sorry, too many clients already
   ```

   → Increase `max_connections` in PostgreSQL

2. **Context deadline exceeded:**
   → Increase timeout in config

3. **Out of memory:**
   → Reduce connection pool or add more RAM

## Advanced Scenarios

### Ramp-Up Test

```bash
# Start light
wrk -t2 -c10 -d10s http://localhost:8081/api/v1/health

# Increase gradually
wrk -t4 -c50 -d20s http://localhost:8081/api/v1/health
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
wrk -t8 -c500 -d60s http://localhost:8081/api/v1/health
wrk -t12 -c1000 -d120s http://localhost:8081/api/v1/health
```

### Sustained Load Test

```bash
# Run for 10 minutes at medium load
wrk -t4 -c100 -d600s http://localhost:8081/api/v1/health
```

Goal: Verify stability over time (no memory leaks, connection leaks)

### Spike Test

```bash
# Normal load
wrk -t4 -c100 -d30s &

# Wait 10s, then spike
sleep 10 && wrk -t12 -c1000 -d20s
```

Goal: Test recovery from sudden traffic increase

## CI/CD Integration

### Makefile Target

```makefile
stress-test: ## Run basic stress test
	@echo "Running stress tests..."
	@wrk -t4 -c100 -d10s http://localhost:8081/api/v1/health
```

### GitHub Actions

```yaml
- name: Run stress tests
  run: |
    make dev &
    sleep 5
    wrk -t4 -c50 -d10s http://localhost:8081/api/v1/health
```

**Note:** Keep CI stress tests light (short duration, low connections)

## Troubleshooting

### wrk: command not found

```bash
# Install wrk (see Prerequisites)
brew install wrk  # macOS
```

### Connection refused

```bash
# Check API is running
curl http://localhost:8081/api/v1/health
```

### wrk crashes

```bash
# Too many connections? Reduce -c parameter
wrk -t4 -c50 -d30s http://localhost:8081/api/v1/health
```

### Inconsistent results

```bash
# Run multiple times and average
for i in {1..5}; do
  wrk -t4 -c100 -d10s http://localhost:8081/api/v1/health
  sleep 5
done
```

## Next Steps

After stress testing:

1. **Profile bottlenecks** - Use pprof
2. **Optimize queries** - Check pg_stat_statements
3. **Add caching** - Redis for hot data
4. **Scale horizontally** - Load balancer + multiple instances
5. **Database tuning** - Connection pool, indexes, vacuum

## Resources

- [wrk GitHub](https://github.com/wg/wrk)
- [wrk Lua Scripting](https://github.com/wg/wrk/blob/master/SCRIPTING)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Go Performance Tips](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)

---

**Created:** December 25, 2025
**Maintained by:** Promenade DevOps Team
