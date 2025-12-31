#!/usr/bin/env python3
"""Fix features blocks - convert YAML to HTML"""

with open('index.md', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Find and replace Technical Foundation features block
in_tech_features = False
tech_start = None
tech_end = None

for i, line in enumerate(lines):
    if line.strip() == '## ⚙️ Technical Foundation':
        # Start looking for features: block after this
        for j in range(i, min(i + 10, len(lines))):
            if lines[j].strip() == 'features:':
                tech_start = j
                in_tech_features = True
                break
    
    if in_tech_features and line.startswith('---'):
        tech_end = i
        break

if tech_start and tech_end:
    print(f"Found Technical Foundation features block: lines {tech_start+1}-{tech_end+1}")
    
    # Replace with HTML
    html_features = [
        '<div class="vp-features">\n',
        '  <div class="feature">\n',
        '    <div class="icon">🏗️</div>\n',
        '    <h3>Domain-Driven Design</h3>\n',
        '    <p>Pure DDD with Bounded Contexts, Aggregates, Value Objects, and Domain Events. Each context is autonomous with its own domain model and database schema.</p>\n',
        '    <a href="/concepts/clean-architecture" class="link">Learn DDD Architecture →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">⚡</div>\n',
        '    <h3>Event-Driven Architecture</h3>\n',
        '    <p>Central Event Bus with Memory and Redis adapters. 377K events/sec throughput, automatic retry, panic recovery, and graceful shutdown.</p>\n',
        '    <a href="/concepts/event-driven" class="link">Explore Event Bus →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">🔐</div>\n',
        '    <h3>JWT + RBAC Authentication</h3>\n',
        '    <p>Token-based authentication with Role-Based Access Control. 15-minute access tokens, 7-day refresh tokens, and Redis-backed token revocation.</p>\n',
        '    <a href="/guide/rbac" class="link">See RBAC Guide →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">🎯</div>\n',
        '    <h3>Bounded Contexts</h3>\n',
        '    <p>6 autonomous business domains - Identity, Customer Management, Order Management, Billing, Warehouse, Analytics. Contexts communicate only via Event Bus.</p>\n',
        '    <a href="/concepts/bounded-contexts" class="link">View All Contexts →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">📊</div>\n',
        '    <h3>Three-Tier Testing</h3>\n',
        '    <p>Professional test organization with 240+ tests. Unit tests (in-place), Smoke tests (mock-based), Integration tests (real DB). 90%+ coverage.</p>\n',
        '    <a href="/guide/testing-patterns" class="link">Read Testing Guide →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">🚀</div>\n',
        '    <h3>Production-Ready</h3>\n',
        '    <p>Rate limiting (5/min login), health checks (4 endpoints), graceful shutdown, structured logging, database migrations, Docker support.</p>\n',
        '    <a href="/guide/health-checks" class="link">Health Monitoring →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">🔄</div>\n',
        '    <h3>Event Bus Adapters</h3>\n',
        '    <p>Switch between Memory (377K events/sec, dev) and Redis (distributed, prod) adapters with single config change. Zero code changes needed.</p>\n',
        '    <a href="/packages/bus" class="link">Event Bus Docs →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">🛡️</div>\n',
        '    <h3>Rate Limiting</h3>\n',
        '    <p>IP-based rate limiting with token bucket algorithm. Login 5/min, Register 3/min. Automatic cleanup prevents memory leaks. Production-ready.</p>\n',
        '    <a href="/guide/rate-limiting" class="link">Rate Limiting Guide →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">💚</div>\n',
        '    <h3>Health Checks</h3>\n',
        '    <p>Comprehensive health monitoring for PostgreSQL, Redis, Event Bus. 4 endpoints, 3 status levels (healthy/degraded/unhealthy), 5-second timeout.</p>\n',
        '    <a href="/guide/health-checks" class="link">Health Check API →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">🗄️</div>\n',
        '    <h3>Redis Caching</h3>\n',
        '    <p>Resource-specific TTL (Reference 1h-24h, User 10-30m, Session 30m-1h). Cache-aside pattern with write-through invalidation. Graceful degradation when Redis unavailable.</p>\n',
        '    <a href="/guide/caching" class="link">Caching Guide →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">🧩</div>\n',
        '    <h3>Value Objects</h3>\n',
        '    <p>Immutable domain primitives - Email, Phone, Money, Address, DateRange. Built-in validation, type safety, and business logic encapsulation.</p>\n',
        '    <a href="/packages/valueobject" class="link">Value Objects →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">🔑</div>\n',
        '    <h3>UUID v7 (Time-Ordered)</h3>\n',
        '    <p>Time-ordered UUIDs provide 2x faster database inserts than UUID v4, better B-tree index locality, and natural ordering by creation time.</p>\n',
        '    <a href="/packages/uuidv7" class="link">UUID v7 Benchmark →</a>\n',
        '  </div>\n',
        '\n',
        '  <div class="feature">\n',
        '    <div class="icon">📝</div>\n',
        '    <h3>Structured Logging</h3>\n',
        '    <p>Context-aware logging with slog wrapper. Request ID propagation, log levels (debug/info/warn/error), JSON format for production.</p>\n',
        '    <a href="/packages/logger" class="link">Logger Package →</a>\n',
        '  </div>\n',
        '</div>\n',
        '\n',
    ]
    
    # Reconstruct file
    new_lines = lines[:tech_start] + html_features + lines[tech_end:]
    
    with open('index.md', 'w', encoding='utf-8') as f:
        f.writelines(new_lines)
    
    print(f"✅ Replaced Technical Foundation features block")
    print(f"   Old: {tech_end - tech_start} lines")
    print(f"   New: {len(html_features)} lines")
else:
    print("❌ Could not find Technical Foundation features block")
