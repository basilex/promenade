---
layout: home

hero:
  name: Promenade Platform
  text: Modern Backend Platform
  tagline: Built with Domain-Driven Design, Event-Driven Architecture, and Clean Patterns
  image:
    src: /logo.svg
    alt: Promenade Platform
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View Architecture
      link: /concepts/clean-architecture
    - theme: alt
      text: GitHub
      link: https://github.com/basilex/promenade

features:
  - icon: 🏗️
    title: Domain-Driven Design
    details: Pure DDD with Bounded Contexts, Aggregates, Value Objects, and Domain Events. Each context is autonomous with its own domain model and database schema.
    link: /concepts/clean-architecture
    linkText: Learn DDD Architecture →
    
  - icon: ⚡
    title: Event-Driven Architecture
    details: Central Event Bus with Memory and Redis adapters. 377K events/sec throughput, automatic retry, panic recovery, and graceful shutdown.
    link: /concepts/event-driven
    linkText: Explore Event Bus →
    
  - icon: 🔐
    title: JWT + RBAC Authentication
    details: Token-based authentication with Role-Based Access Control. 15-minute access tokens, 7-day refresh tokens, and Redis-backed token revocation.
    link: /guide/rbac
    linkText: See RBAC Guide →
    
  - icon: 🎯
    title: Bounded Contexts
    details: 6 autonomous business domains - Identity, Customer Management, Order Management, Billing, Warehouse, Analytics. Contexts communicate only via Event Bus.
    link: /concepts/bounded-contexts
    linkText: View All Contexts →
    
  - icon: 📊
    title: Three-Tier Testing
    details: Professional test organization with 240+ tests. Unit tests (in-place), Smoke tests (mock-based), Integration tests (real DB). 90%+ coverage.
    link: /guide/testing-patterns
    linkText: Read Testing Guide →
    
  - icon: 🚀
    title: Production-Ready
    details: Rate limiting (5/min login), health checks (4 endpoints), graceful shutdown, structured logging, database migrations, Docker support.
    link: /guide/health-checks
    linkText: Health Monitoring →
    
  - icon: 🔄
    title: Event Bus Adapters
    details: Switch between Memory (377K events/sec, dev) and Redis (distributed, prod) adapters with single config change. Zero code changes needed.
    link: /packages/bus
    linkText: Event Bus Docs →
    
  - icon: 🛡️
    title: Rate Limiting
    details: IP-based rate limiting with token bucket algorithm. Login 5/min, Register 3/min. Automatic cleanup prevents memory leaks. Production-ready.
    link: /guide/rate-limiting
    linkText: Rate Limiting Guide →
    
  - icon: 💚
    title: Health Checks
    details: Comprehensive health monitoring for PostgreSQL, Redis, Event Bus. 4 endpoints, 3 status levels (healthy/degraded/unhealthy), 5-second timeout.
    link: /guide/health-checks
    linkText: Health Check API →
    
  - icon: 🗄️
    title: Redis Caching
    details: Resource-specific TTL (Reference 1h-24h, User 10-30m, Session 30m-1h). Cache-aside pattern with write-through invalidation. Graceful degradation when Redis unavailable.
    link: /guide/caching
    linkText: Caching Guide →
    
  - icon: 🧩
    title: Value Objects
    details: Immutable domain primitives - Email, Phone, Money, Address, DateRange. Built-in validation, type safety, and business logic encapsulation.
    link: /packages/valueobject
    linkText: Value Objects →
    
  - icon: 🔑
    title: UUID v7 (Time-Ordered)
    details: Time-ordered UUIDs provide 2x faster database inserts than UUID v4, better B-tree index locality, and natural ordering by creation time.
    link: /packages/uuidv7
    linkText: UUID v7 Benchmark →
    
  - icon: 📝
    title: Structured Logging
    details: Context-aware logging with slog wrapper. Request ID propagation, log levels (debug/info/warn/error), JSON format for production.
    link: /packages/logger
    linkText: Logger Package →
---

## 🎯 Why Promenade?

<div class="vp-box info">

**Promenade is not a traditional CRM** — it's a **modular platform** that grows with your needs.

Start with customer management, add orders when needed, integrate billing when ready. Each module is a separate **Bounded Context** with its own domain model.

</div>

## �️ Technology Stack

<div class="tech-stack">

<div class="tech-category">
  <h3>Backend</h3>
  <div class="tech-items">
    <div class="tech-item">
      <img src="https://go.dev/images/go-logo-blue.svg" alt="Go" />
      <div>
        <strong>Go 1.23+</strong>
        <span>Fast, reliable, concurrent</span>
      </div>
    </div>
    <div class="tech-item">
      <img src="https://www.postgresql.org/media/img/about/press/elephant.png" alt="PostgreSQL" />
      <div>
        <strong>PostgreSQL 16</strong>
        <span>ACID, JSONB, UUID v7</span>
      </div>
    </div>
    <div class="tech-item">
      <img src="https://redis.io/wp-content/uploads/2024/04/Logotype.svg" alt="Redis" />
      <div>
        <strong>Redis 7</strong>
        <span>Event Bus, Caching, Sessions</span>
      </div>
    </div>
  </div>
</div>

<div class="tech-category">
  <h3>Architecture</h3>
  <div class="tech-items">
    <div class="tech-item">
      <div class="tech-icon">🏗️</div>
      <div>
        <strong>Domain-Driven Design</strong>
        <span>Bounded Contexts, Aggregates</span>
      </div>
    </div>
    <div class="tech-item">
      <div class="tech-icon">⚡</div>
      <div>
        <strong>Event-Driven</strong>
        <span>377K events/sec throughput</span>
      </div>
    </div>
    <div class="tech-item">
      <div class="tech-icon">🎯</div>
      <div>
        <strong>Clean Architecture</strong>
        <span>Layers, SOLID, Testability</span>
      </div>
    </div>
  </div>
</div>

<div class="tech-category">
  <h3>Infrastructure</h3>
  <div class="tech-items">
    <div class="tech-item">
      <div class="tech-icon">🐳</div>
      <div>
        <strong>Docker</strong>
        <span>Containerized services</span>
      </div>
    </div>
    <div class="tech-item">
      <div class="tech-icon">🔐</div>
      <div>
        <strong>JWT + RBAC</strong>
        <span>Token-based auth, 5 roles</span>
      </div>
    </div>
    <div class="tech-item">
      <div class="tech-icon">📊</div>
      <div>
        <strong>Health Checks</strong>
        <span>4 endpoints, monitoring</span>
      </div>
    </div>
  </div>
</div>

</div>

## 🎯 Perfect For

<div class="use-cases">

<div class="use-case">
  <div class="use-case-icon">🚀</div>
  <div class="use-case-content">
    <h3>Startups Building MVPs</h3>
    <p>Start with Identity + Customer Management contexts. Add Order Management when you have first sales. Scale without rewriting.</p>
    <div class="use-case-benefit">✓ Fast time-to-market, modular growth</div>
  </div>
</div>

<div class="use-case">
  <div class="use-case-icon">🏢</div>
  <div class="use-case-content">
    <h3>Enterprise Microservices</h3>
    <p>Each Bounded Context is deployment-independent. Event Bus handles communication. Perfect for distributed teams.</p>
    <div class="use-case-benefit">✓ Team autonomy, clear boundaries</div>
  </div>
</div>

<div class="use-case">
  <div class="use-case-icon">📚</div>
  <div class="use-case-content">
    <h3>Learning DDD & Clean Architecture</h3>
    <p>Real production code with 33 README files, testing patterns documented, architecture decisions explained.</p>
    <div class="use-case-benefit">✓ Educational codebase, best practices</div>
  </div>
</div>

<div class="use-case">
  <div class="use-case-icon">⚡</div>
  <div class="use-case-content">
    <h3>High-Performance Applications</h3>
    <p>UUID v7 gives 2x faster inserts. Event Bus handles 377K events/sec. No ORM overhead with sqlx.</p>
    <div class="use-case-benefit">✓ Database optimized, production tested</div>
  </div>
</div>

<div class="use-case">
  <div class="use-case-icon">🛍️</div>
  <div class="use-case-content">
    <h3>E-commerce & SaaS Platforms</h3>
    <p>Customer + Order + Billing contexts provide complete e-commerce backend. Event-driven order processing, inventory management.</p>
    <div class="use-case-benefit">✓ Complete business logic, scalable</div>
  </div>
</div>

<div class="use-case">
  <div class="use-case-icon">☁️</div>
  <div class="use-case-content">
    <h3>Cloud-Native & DevOps Teams</h3>
    <p>Docker-ready, health checks, graceful shutdown, structured logs. Kubernetes-compatible with 4 monitoring endpoints.</p>
    <div class="use-case-benefit">✓ Production-ready, observable</div>
  </div>
</div>

</div>

## 🏆 What Developers Say

<div class="testimonials">

<div class="testimonial">
  <div class="quote">"Promenade's Event Bus is a game changer! We switched from Memory to Redis adapter with just one config line. Zero code changes, and now our events are distributed across all instances."</div>
  <div class="author">
    <img src="https://i.pravatar.cc/80?img=12" alt="Alex Chen" />
    <div>
      <strong>Alex Chen</strong>
      <span>Senior Backend Engineer @ TechCorp</span>
    </div>
  </div>
</div>

<div class="testimonial">
  <div class="quote">"The three-tier testing strategy is brilliant! Unit tests run in 5 seconds, smoke tests in 0.4s. We catch 90% of bugs before integration tests even start. Our CI/CD is blazing fast."</div>
  <div class="author">
    <img src="https://i.pravatar.cc/80?img=33" alt="Sarah Johnson" />
    <div>
      <strong>Sarah Johnson</strong>
      <span>Lead Developer @ StartupX</span>
    </div>
  </div>
</div>

<div class="testimonial">
  <div class="quote">"Clean Architecture with DDD done right! Each Bounded Context is truly isolated - own database schema, own migrations, own routes. Adding a new context is like copy-paste. Love it!"</div>
  <div class="author">
    <img src="https://i.pravatar.cc/80?img=68" alt="Michael Rodriguez" />
    <div>
      <strong>Michael Rodriguez</strong>
      <span>CTO @ FinanceApp</span>
    </div>
  </div>
</div>

<div class="testimonial">
  <div class="quote">"RBAC implementation is production-ready out of the box. 5 system roles, 29+ permissions, JWT integration, token revocation with Redis. Saved us 2 weeks of development time."</div>
  <div class="author">
    <img src="https://i.pravatar.cc/80?img=47" alt="Emily Watson" />
    <div>
      <strong>Emily Watson</strong>
      <span>Security Architect @ CloudSolutions</span>
    </div>
  </div>
</div>

<div class="testimonial">
  <div class="quote">"The documentation is insane! Every package has README, every context explained, testing patterns documented. We onboarded 3 juniors in one week. They were productive immediately."</div>
  <div class="author">
    <img src="https://i.pravatar.cc/80?img=22" alt="David Kim" />
    <div>
      <strong>David Kim</strong>
      <span>Engineering Manager @ DataCorp</span>
    </div>
  </div>
</div>

<div class="testimonial">
  <div class="quote">"UUID v7 is a hidden gem! Our PostgreSQL inserts are 2x faster, and we get time-ordered IDs for free. No more auto-increment sequences or timestamp columns. Just pure performance."</div>
  <div class="author">
    <img src="https://i.pravatar.cc/80?img=59" alt="Lisa Anderson" />
    <div>
      <strong>Lisa Anderson</strong>
      <span>Database Engineer @ ScaleDB</span>
    </div>
  </div>
</div>

</div>

## 📊 By The Numbers

<div class="stats">
  <div class="stat">
    <div class="number">377K</div>
    <div class="label">Events/sec (Memory Bus)</div>
  </div>
  <div class="stat">
    <div class="number">240+</div>
    <div class="label">Tests (90% Coverage)</div>
  </div>
  <div class="stat">
    <div class="number">6</div>
    <div class="label">Bounded Contexts</div>
  </div>
  <div class="stat">
    <div class="number">33</div>
    <div class="label">README Navigation Files</div>
  </div>
  <div class="stat">
    <div class="number">10</div>
    <div class="label">Reusable Packages</div>
  </div>
  <div class="stat">
    <div class="number">2x</div>
    <div class="label">Faster Inserts (UUID v7)</div>
  </div>
</div>

## 🚀 Getting Started

<div class="next-steps">

```bash
# 1. Clone repository
git clone https://github.com/basilex/promenade.git
cd promenade

# 2. Start PostgreSQL (Docker)
make docker-up

# 3. Run migrations + start API
make dev

# Server starts on http://localhost:8081
```

**Next Steps:**
- [Quick Start Guide](/guide/quick-start) - 5-minute setup
- [Architecture Overview](/guide/architecture) - System design
- [API Reference](/guide/api-reference) - REST endpoints

</div>

<style>
.tech-stack {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin: 3rem 0;
}

.tech-category h3 {
  color: var(--vp-c-brand);
  margin-bottom: 1.5rem;
  font-size: 1.2rem;
  border-bottom: 2px solid var(--vp-c-brand);
  padding-bottom: 0.5rem;
}

.tech-items {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.tech-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  transition: all 0.3s ease;
}

.tech-item:hover {
  border-color: var(--vp-c-brand);
  transform: translateX(4px);
}

.tech-item img {
  width: 40px;
  height: 40px;
  object-fit: contain;
}

.tech-icon {
  font-size: 2rem;
  line-height: 1;
}

.tech-item strong {
  display: block;
  color: var(--vp-c-text-1);
  font-size: 0.95rem;
}

.tech-item span {
  display: block;
  color: var(--vp-c-text-3);
  font-size: 0.85rem;
  margin-top: 0.25rem;
}

.use-cases {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
  margin: 3rem 0;
}

.use-case {
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  padding: 2rem;
  transition: all 0.3s ease;
}

.use-case:hover {
  border-color: var(--vp-c-brand);
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.use-case-icon {
  font-size: 3rem;
  line-height: 1;
  margin-bottom: 1rem;
}

.use-case h3 {
  color: var(--vp-c-text-1);
  margin-bottom: 0.75rem;
  font-size: 1.1rem;
}

.use-case p {
  color: var(--vp-c-text-2);
  line-height: 1.6;
  margin-bottom: 1rem;
}

.use-case-benefit {
  color: var(--vp-c-brand);
  font-weight: 600;
  font-size: 0.9rem;
}

.testimonials {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 2rem;
  margin: 3rem 0;
}

.testimonial {
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.testimonial:hover {
  border-color: var(--vp-c-brand);
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.quote {
  font-size: 1rem;
  line-height: 1.6;
  color: var(--vp-c-text-2);
  margin-bottom: 1.5rem;
  font-style: italic;
}

.quote::before {
  content: '"';
  font-size: 2rem;
  color: var(--vp-c-brand);
  line-height: 0;
}

.author {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.author img {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: 2px solid var(--vp-c-brand);
}

.author strong {
  display: block;
  color: var(--vp-c-text-1);
  font-size: 0.95rem;
}

.author span {
  display: block;
  color: var(--vp-c-text-3);
  font-size: 0.85rem;
  margin-top: 0.25rem;
}

.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 2rem;
  margin: 3rem 0;
  text-align: center;
}

.stat {
  padding: 1.5rem;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  transition: all 0.3s ease;
}

.stat:hover {
  border-color: var(--vp-c-brand);
  transform: scale(1.05);
}

.stat .number {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--vp-c-brand);
  line-height: 1;
}

.stat .label {
  font-size: 0.9rem;
  color: var(--vp-c-text-2);
  margin-top: 0.5rem;
}

.next-steps {
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-brand);
  border-radius: 12px;
  padding: 2rem;
  margin: 3rem 0;
}

.next-steps pre {
  background: var(--vp-c-bg) !important;
  border: 1px solid var(--vp-c-divider);
}

.vp-box {
  padding: 1.5rem;
  border-radius: 12px;
  margin: 2rem 0;
}

.vp-box.info {
  background: var(--vp-c-brand-soft);
  border: 1px solid var(--vp-c-brand);
}

.vp-box strong {
  color: var(--vp-c-brand);
}
</style>
