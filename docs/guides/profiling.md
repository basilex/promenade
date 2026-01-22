# Performance Profiling

**Status**:  **Phase 2** - Advanced tooling for performance optimization  
**Prerequisites**: Go profiling tools (`go tool pprof`), production access (for production profiling)

---

## Table of Contents

1. [Overview](#overview)
2. [Development Profiling](#development-profiling)
3. [Production Profiling](#production-profiling)
4. [Continuous Profiling](#continuous-profiling)
5. [Tooling & Visualization](#tooling--visualization)
6. [Common Patterns](#common-patterns)

---

## Overview

Promenade uses Go's built-in `pprof` profiling tools to identify performance bottlenecks.

### Profiling Types

| Type          | Use Case                                  | Overhead |
| ------------- | ----------------------------------------- | -------- |
| **CPU**       | Identify hot paths (high CPU usage)       | ~5%      |
| **Memory**    | Track allocations and memory leaks        | ~1%      |
| **Goroutine** | Debug goroutine leaks and blocking        | Minimal  |
| **Mutex**     | Identify lock contention                  | ~10%     |
| **Block**     | Track blocking operations (I/O, channels) | ~10%     |

---

## Development Profiling

### CPU Profiling (Unit Tests)

Profile CPU usage during tests:

```bash
# Run benchmark with CPU profile
go test -bench=BenchmarkCreateCustomer \
  -benchtime=10s \
  -cpuprofile=cpu.prof \
  ./test/benchmark/contexts/customer-mgmt/

# Analyze profile (interactive)
go tool pprof cpu.prof
```

**pprof commands**:

```
(pprof) top10          # Top 10 functions by CPU time
(pprof) list CreateCustomer  # Source code with annotations
(pprof) web            # Graphviz visualization (requires graphviz)
```

**Example output**:

```
Showing top 10 nodes out of 87
      flat  flat%   sum%        cum   cum%
     0.78s 15.03% 15.03%      1.23s 23.70%  github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/aggregate.(*Customer).Create
     0.52s 10.02% 25.05%      0.89s 17.15%  github.com/jmoiron/sqlx.(*DB).QueryContext
     0.41s  7.90% 32.95%      0.67s 12.91%  encoding/json.Marshal
```

---

### Memory Profiling (Benchmarks)

Track memory allocations:

```bash
# Run benchmark with memory profile
go test -bench=BenchmarkListCustomers \
  -benchtime=10s \
  -memprofile=mem.prof \
  ./test/benchmark/contexts/customer-mgmt/

# Analyze allocations
go tool pprof -alloc_space mem.prof
go tool pprof -inuse_space mem.prof  # Current memory usage
```

**pprof commands**:

```
(pprof) top10 -cum     # Top 10 by cumulative allocations
(pprof) list List      # Show allocations per line
(pprof) png > mem.png  # Export as image
```

---

### Goroutine Profiling (Integration Tests)

Debug goroutine leaks:

```bash
# Start application with pprof endpoint (see Production Profiling)
make dev

# Capture goroutine profile
curl http://localhost:8080/debug/pprof/goroutine > goroutine.prof

# Analyze
go tool pprof goroutine.prof
```

**Identify leaks**:

```
(pprof) top
Showing nodes accounting for 142 goroutines
      142   68.93% 68.93%      142 68.93%  runtime.gopark
       32   15.53% 84.47%       32 15.53%  net/http.(*conn).serve
       20    9.71% 94.17%       20  9.71%  github.com/basilex/promenade/pkg/scheduler.(*Scheduler).Start.func1
```

>  **Goroutine Leak Detection**: If goroutine count grows unbounded (e.g., 100→500→1000), you have a leak.

---

## Production Profiling

### Enable pprof Endpoints

Promenade exposes pprof endpoints in **development and test environments only** (disabled in production by default).

**Configuration**: [cmd/api/server.go](../../cmd/api/server.go)

```go
// pprof endpoints (only in dev/test)
if config.Environment != "production" {
    router.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
}
```

>  **Security**: pprof endpoints are **not exposed in production** to prevent information leakage.

---

### Production Profiling Strategy

#### Option 1: Temporary pprof Exposure (Risky)

Only enable pprof for **short debugging sessions** with IP whitelisting:

```go
// ONLY FOR EMERGENCY DEBUGGING
if config.Environment == "production" && config.EnablePprof {
    router.Use(middleware.IPWhitelist([]string{"10.0.0.0/8"})) // Internal IPs only
    router.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
}
```

**Capture profiles**:

```bash
# CPU profile (30 seconds)
curl http://prod-api:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Heap profile
curl http://prod-api:8080/debug/pprof/heap > heap.prof

# Goroutine profile
curl http://prod-api:8080/debug/pprof/goroutine > goroutine.prof

# Analyze locally
go tool pprof cpu.prof
```

---

#### Option 2: Continuous Profiling (Phase 2)

Use **Pyroscope** or **Datadog Continuous Profiler** for always-on profiling:

```go
import "github.com/pyroscope-io/client/pyroscope"

func main() {
    pyroscope.Start(pyroscope.Config{
        ApplicationName: "promenade-api",
        ServerAddress:   "http://pyroscope:4040",
        ProfileTypes: []pyroscope.ProfileType{
            pyroscope.ProfileCPU,
            pyroscope.ProfileAllocObjects,
            pyroscope.ProfileInuseObjects,
        },
    })
    defer pyroscope.Stop()
}
```

**Benefits**:

- **Historical comparison**: Compare profiles across deployments
- **No overhead**: ~0.5% CPU overhead
- **Flame graphs**: Built-in visualization

>  **Phase 2**: Pyroscope integration planned for Q3 2026 (see [observability-strategy.md](observability-strategy.md)).

---

## Continuous Profiling

### Automated Profiling (CI/CD)

Add profiling to CI pipeline to detect regressions:

```yaml
# .github/workflows/benchmark.yml
- name: Run Benchmarks with Profiling
  run: |
    go test -bench=. -benchtime=10s \
      -cpuprofile=cpu.prof \
      -memprofile=mem.prof \
      ./test/benchmark/contexts/...

    # Upload profiles as artifacts
    - uses: actions/upload-artifact@v3
      with:
        name: profiles
        path: |
          cpu.prof
          mem.prof
```

---

### Profiling Baselines

Maintain baseline profiles to compare against:

```bash
# Generate baseline
make test-benchmark
go test -bench=. -cpuprofile=baseline-cpu.prof ./test/benchmark/contexts/customer-mgmt/

# Compare after changes
go test -bench=. -cpuprofile=current-cpu.prof ./test/benchmark/contexts/customer-mgmt/
go tool pprof -base=baseline-cpu.prof current-cpu.prof

# Show differences
(pprof) top -diff_base
```

---

## Tooling & Visualization

### pprof Web UI

```bash
# Interactive web UI
go tool pprof -http=:8081 cpu.prof

# Open browser at http://localhost:8081
```

**Views**:

- **Top**: Table of top functions
- **Graph**: Call graph visualization
- **Flame Graph**: Hierarchical view of CPU time
- **Source**: Annotated source code

---

### External Tools

#### 1. **pprof** (Google)

```bash
# Install
go install github.com/google/pprof@latest

# Enhanced web UI
pprof -http=:8081 cpu.prof
```

#### 2. **Graphviz** (Call Graphs)

```bash
# Install
brew install graphviz  # macOS
apt install graphviz   # Ubuntu

# Generate PNG
go tool pprof -png cpu.prof > profile.png
```

#### 3. **Speedscope** (Flame Graphs)

```bash
# Install
npm install -g speedscope

# Open in browser
speedscope cpu.prof
```

---

## Common Patterns

### Pattern 1: Profile Before Optimization

```bash
# 1. Establish baseline
go test -bench=BenchmarkProcessOrder -benchtime=10s -cpuprofile=before.prof

# 2. Make optimization changes
# ... edit code ...

# 3. Compare
go test -bench=BenchmarkProcessOrder -benchtime=10s -cpuprofile=after.prof
go tool pprof -base=before.prof after.prof
```

---

### Pattern 2: Profile High-Throughput Endpoint

```bash
# Start server
make dev

# Generate load (hey tool)
hey -z 30s -c 50 http://localhost:8080/api/v1/customers

# Capture CPU profile during load
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze
go tool pprof -http=:8081 cpu.prof
```

---

### Pattern 3: Detect Memory Leaks

```bash
# 1. Baseline heap profile
curl http://localhost:8080/debug/pprof/heap > heap-before.prof

# 2. Run load test
hey -z 300s -c 100 http://localhost:8080/api/v1/orders

# 3. Compare heap after load
curl http://localhost:8080/debug/pprof/heap > heap-after.prof
go tool pprof -base=heap-before.prof heap-after.prof

# 4. Look for growing allocations
(pprof) top -cum
```

---

### Pattern 4: Benchmark-Driven Profiling

Run all benchmarks with profiling enabled:

```bash
# Profile all benchmarks
make test-benchmark-all
go test -bench=. -benchtime=10s \
  -cpuprofile=cpu.prof \
  -memprofile=mem.prof \
  ./test/benchmark/contexts/...

# Analyze CPU
go tool pprof -http=:8081 cpu.prof

# Analyze memory
go tool pprof -http=:8082 mem.prof
```

---

## Related Documentation

- [Benchmark README](../../test/benchmark/README.md) - Benchmarking strategy with baselines
- [Testing Patterns](testing-patterns.md) - Test suite structure
- [Observability Strategy](observability-strategy.md) - Tracing & metrics (Phase 2)

---

## Next Steps

- [ ] **Phase 2A**: Integrate Pyroscope for continuous profiling (Q3 2026)
- [ ] **Phase 2B**: Add automated profiling to CI/CD (detect regressions)
- [ ] **Phase 2C**: Create profiling runbook for production incidents
- [ ] **Phase 3**: Production pprof exposure with mutual TLS authentication
