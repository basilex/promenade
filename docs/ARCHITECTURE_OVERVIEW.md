# Promenade Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         APPLICATION LAYER                            │
│                         (cmd/api/main.go)                           │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ├─── Initialize Infrastructure
                           ├─── Load Core Configuration
                           ├─── Initialize Module System
                           └─── Start HTTP Server

┌─────────────────────────────────────────────────────────────────────┐
│                         CORE INFRASTRUCTURE                          │
│                    (internal/infrastructure/)                        │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │  Database    │  │  Event Bus   │  │  Scheduler   │             │
│  │  Connection  │  │  (Memory/    │  │  (Cron)      │             │
│  │  (PostgreSQL)│  │   Redis)     │  │              │             │
│  └──────────────┘  └──────────────┘  └──────────────┘             │
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │  Config      │  │  Logger      │  │  Email       │             │
│  │  (YAML)      │  │  (slog)      │  │  Service     │             │
│  └──────────────┘  └──────────────┘  └──────────────┘             │
└─────────────────────────────────────────────────────────────────────┘
                           │
                           │ provides infrastructure to
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         CORE DOMAIN                                  │
│                    (internal/domain/entity)                          │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐       │
│  │  SECURITY & AUTH (always enabled)                        │       │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐              │       │
│  │  │   User   │  │ Session  │  │   Role   │              │       │
│  │  │          │  │  (JWT)   │  │          │              │       │
│  │  └──────────┘  └──────────┘  └──────────┘              │       │
│  │                                                          │       │
│  │  ┌──────────┐                                           │       │
│  │  │Permission│  (RBAC: resource:action)                  │       │
│  │  └──────────┘                                           │       │
│  └─────────────────────────────────────────────────────────┘       │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────┐       │
│  │  REFERENCE DATA (справочники - stable, shared)          │       │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐              │       │
│  │  │ Country  │  │ Currency │  │ Timezone │              │       │
│  │  │(195)     │  │ (170)    │  │ (500+)   │              │       │
│  │  └──────────┘  └──────────┘  └──────────┘              │       │
│  │                                                          │       │
│  │  ┌──────────┐  ┌──────────┐                            │       │
│  │  │ Language │  │Retention │                            │       │
│  │  │ (180)    │  │ Policy   │                            │       │
│  │  └──────────┘  └──────────┘                            │       │
│  └─────────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────┘
                           │
                           │ uses repositories
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         CORE USE CASES                               │
│                      (internal/usecase/)                             │
├─────────────────────────────────────────────────────────────────────┤
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐       │
│  │ Auth UseCase   │  │ Role UseCase   │  │Permission UC   │       │
│  │ - Register     │  │ - CRUD roles   │  │ - CRUD perms   │       │
│  │ - Login        │  │ - Assign       │  │ - Check access │       │
│  │ - Logout       │  │                │  │                │       │
│  └────────────────┘  └────────────────┘  └────────────────┘       │
│                                                                      │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐       │
│  │Country UseCase │  │Currency UC     │  │ Purge UseCase  │       │
│  │ - List         │  │ - List         │  │ - Orchestrate  │       │
│  │ - Get by code  │  │ - Get by code  │  │ - Schedule     │       │
│  └────────────────┘  └────────────────┘  └────────────────┘       │
└─────────────────────────────────────────────────────────────────────┘
                           │
                           │ exposes via HTTP
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         CORE API ROUTES                              │
│                  (internal/adapter/http/v1/)                         │
├─────────────────────────────────────────────────────────────────────┤
│  /api/v1/                                                           │
│    ├── /health           Health checks                           │
│    ├── /auth/*           Login, register, logout                 │
│    ├── /countries/*      Reference data                          │
│    ├── /currencies/*     Reference data                          │
│    ├── /roles/*          RBAC management                         │
│    ├── /permissions/*    RBAC management                         │
│    └── /admin/*          Purge, system management                │
└─────────────────────────────────────────────────────────────────────┘


═══════════════════════════════════════════════════════════════════════
                         MODULE SYSTEM
═══════════════════════════════════════════════════════════════════════

┌─────────────────────────────────────────────────────────────────────┐
│                      MODULE REGISTRY & MANAGER                       │
│                         (pkg/module/)                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌────────────────────────────────────────────────────────┐        │
│  │  DefaultRegistry (global)                              │        │
│  │  ┌──────────────────────────────────────────────┐     │        │
│  │  │  - Register(module)                          │     │        │
│  │  │  - GetEnabled(config)                        │     │        │
│  │  │  - InitializeAll() // dependency order       │     │        │
│  │  │  - StartAll() // background workers          │     │        │
│  │  │  - StopAll() // graceful shutdown            │     │        │
│  │  └──────────────────────────────────────────────┘     │        │
│  └────────────────────────────────────────────────────────┘        │
│                                                                      │
│  Provides to modules:                                               │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │
│  │     DB     │  │  EventBus  │  │    JWT     │  │   Config   │  │
│  │ (shared)   │  │  (shared)  │  │  (shared)  │  │  (shared)  │  │
│  └────────────┘  └────────────┘  └────────────┘  └────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                           │
                           │ manages modules
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         MODULE: POSTS                                │
│                  (internal/modules/posts/)                           │
├─────────────────────────────────────────────────────────────────────┤
│  Status:  Enabled                                                 │
│  License: Free                                                      │
│                                                                      │
│  ┌──────────────────────────────────────────────────────┐          │
│  │  Domain Entities                                      │          │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │          │
│  │  │   Post   │  │ Comment  │  │   Like   │           │          │
│  │  │          │  │(threaded)│  │          │           │          │
│  │  └──────────┘  └──────────┘  └──────────┘           │          │
│  └──────────────────────────────────────────────────────┘          │
│                                                                      │
│  ┌──────────────────────────────────────────────────────┐          │
│  │  Use Cases                                            │          │
│  │  - CreatePost, UpdatePost, DeletePost                │          │
│  │  - CreateComment, ReplyToComment                     │          │
│  │  - LikePost, LikeComment                             │          │
│  └──────────────────────────────────────────────────────┘          │
│                                                                      │
│  ┌──────────────────────────────────────────────────────┐          │
│  │  Configuration (config/config.dev.yaml)              │          │
│  │  - max_content_length: 10000                         │          │
│  │  - comments.max_depth: 10                            │          │
│  │  - purge.user_posts.retention_days: 90               │          │
│  │  - purge.post_comments.retention_days: 30            │          │
│  └──────────────────────────────────────────────────────┘          │
│                                                                      │
│  ┌──────────────────────────────────────────────────────┐          │
│  │  Purge Handlers (registered to purge.DefaultRegistry)│          │
│  │  - PostPurgeHandler                                  │          │
│  │  - CommentPurgeHandler                               │          │
│  └──────────────────────────────────────────────────────┘          │
│                                                                      │
│  Routes: /api/v1/posts/*, /api/v1/comments/*, /api/v1/likes/*     │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                        MODULE: PROFILES                              │
│                  (internal/modules/profiles/)                        │
├─────────────────────────────────────────────────────────────────────┤
│  Status:  Enabled                                                 │
│  License: Free                                                      │
│                                                                      │
│  ┌──────────────────────────────────────────────────────┐          │
│  │  Domain Entities                                      │          │
│  │  ┌──────────┐  ┌──────────┐                          │          │
│  │  │  Profile │  │ Contact  │                          │          │
│  │  │          │  │          │                          │          │
│  │  └──────────┘  └──────────┘                          │          │
│  └──────────────────────────────────────────────────────┘          │
│                                                                      │
│  ┌──────────────────────────────────────────────────────┐          │
│  │  Use Cases                                            │          │
│  │  - CreateProfile, UpdateProfile                      │          │
│  │  - AddContact, VerifyContact                         │          │
│  │  - SetPrimaryContact                                 │          │
│  └──────────────────────────────────────────────────────┘          │
│                                                                      │
│  ┌──────────────────────────────────────────────────────┐          │
│  │  Configuration (config/config.dev.yaml)              │          │
│  │  - profiles.max_per_user: 1                          │          │
│  │  - contacts.max_per_user: 5                          │          │
│  │  - contacts.verification_required: true              │          │
│  └──────────────────────────────────────────────────────┘          │
│                                                                      │
│  Routes: /api/v1/profiles/*, /api/v1/contacts/*                    │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                      MODULE: WAREHOUSE                            │
│                  (internal/modules/warehouse/)                       │
├─────────────────────────────────────────────────────────────────────┤
│  Status:   Disabled (requires license)                           │
│  License: Commercial (license key validation required)             │
│                                                                      │
│  ┌──────────────────────────────────────────────────────┐          │
│  │  Features (when licensed)                             │          │
│  │  - Inventory management                              │          │
│  │  - Stock tracking                                    │          │
│  │  - Barcode scanning                                  │          │
│  │  - Warehouse locations                               │          │
│  └──────────────────────────────────────────────────────┘          │
│                                                                      │
│  Routes: /api/v1/warehouse/* (when enabled)                        │
└─────────────────────────────────────────────────────────────────────┘


═══════════════════════════════════════════════════════════════════════
                    INTER-MODULE COMMUNICATION
═══════════════════════════════════════════════════════════════════════

┌─────────────────────────────────────────────────────────────────────┐
│                         EVENT BUS (pkg/bus/)                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Adapters:                                                          │
│  ┌──────────────────┐              ┌──────────────────┐           │
│  │  Memory Adapter  │              │  Redis Adapter   │           │
│  │  (dev/test)      │              │  (production)    │           │
│  │  - Fast          │              │  - Distributed   │           │
│  │  - Zero deps     │              │  - Persistent    │           │
│  └──────────────────┘              └──────────────────┘           │
│                                                                      │
│  Event Flow:                                                        │
│  ┌────────────┐      publish      ┌────────────┐                  │
│  │  Module A  │ ─────────────────>│  EventBus  │                  │
│  └────────────┘                    └─────┬──────┘                  │
│                                           │ fan-out                │
│                      ┌───────────────────┼───────────────────┐    │
│                      ▼                    ▼                   ▼    │
│                 ┌────────┐          ┌────────┐          ┌────────┐│
│                 │Module B│          │Module C│          │Module D││
│                 │(async) │          │(async) │          │(async) ││
│                 └────────┘          └────────┘          └────────┘│
│                                                                      │
│  Examples:                                                          │
│  - user.registered → send welcome email (async)                    │
│  - post.created → update user stats (async)                        │
│  - purge.completed → log to audit (async)                          │
└─────────────────────────────────────────────────────────────────────┘


═══════════════════════════════════════════════════════════════════════
                    CONFIGURATION ARCHITECTURE
═══════════════════════════════════════════════════════════════════════

config/
├── app.dev.yaml            Core infrastructure (dev)
├── app.test.yaml           Core infrastructure (test)
├── app.prod.yaml           Core infrastructure (prod)
└── modules.yaml            Which modules to load

internal/modules/posts/config/
├── config.dev.yaml         Posts module settings (dev)
├── config.test.yaml        Posts module settings (test)
└── config.prod.yaml        Posts module settings (prod)

internal/modules/profiles/config/
├── config.dev.yaml         Profiles module settings (dev)
├── config.test.yaml        Profiles module settings (test)
└── config.prod.yaml        Profiles module settings (prod)

.env.example                Environment variable overrides (optional)


═══════════════════════════════════════════════════════════════════════
                    MODULE LIFECYCLE
═══════════════════════════════════════════════════════════════════════

1. AUTO-REGISTRATION (via init())
   ┌──────────────────────────────────────────┐
   │ package posts                            │
   │                                          │
   │ func init() {                            │
   │     module.DefaultRegistry.Register(     │
   │         New()                            │
   │     )                                    │
   │ }                                        │
   └──────────────────────────────────────────┘

2. DISCOVERY & FILTERING
   - Read config/modules.yaml
   - Filter enabled modules
   - Resolve dependencies (topological sort)

3. INITIALIZATION (in dependency order)
   For each module:
   ┌──────────────────────────────────────────┐
   │ Initialize(ctx, core)                    │
   │   - Load module config                   │
   │   - Setup repositories                   │
   │   - Setup use cases                      │
   │   - Setup handlers                       │
   │   - Register purge handlers              │
   │   - Register retention policies          │
   │   - Register permissions                 │
   └──────────────────────────────────────────┘

4. ROUTE REGISTRATION
   For each module:
   ┌──────────────────────────────────────────┐
   │ RegisterRoutes(router)                   │
   │   - Mount module routes                  │
   │   - Apply middleware                     │
   └──────────────────────────────────────────┘

5. EVENT SUBSCRIPTION
   For each module:
   ┌──────────────────────────────────────────┐
   │ RegisterEventHandlers(eventBus)          │
   │   - Subscribe to events                  │
   │   - Setup async handlers                 │
   └──────────────────────────────────────────┘

6. START (background workers)
   For each module:
   ┌──────────────────────────────────────────┐
   │ Start(ctx)                               │
   │   - Start cron jobs                      │
   │   - Start background workers             │
   └──────────────────────────────────────────┘

7. RUNTIME
   - Modules process requests
   - Publish/subscribe to events
   - Execute scheduled tasks

8. SHUTDOWN (on SIGTERM/SIGINT)
   For each module (reverse order):
   ┌──────────────────────────────────────────┐
   │ Stop(ctx)                                │
   │   - Stop workers gracefully              │
   │   - Close connections                    │
   │   - Cleanup resources                    │
   └──────────────────────────────────────────┘


═══════════════════════════════════════════════════════════════════════
                    KEY DESIGN PRINCIPLES
═══════════════════════════════════════════════════════════════════════

1. CORE = INFRASTRUCTURE + REFERENCE DATA
    Core provides services (DB, EventBus, Config, JWT)
    Core contains stable reference data (countries, currencies)
    Core manages security (auth, RBAC)
    Core does NOT contain business logic

2. MODULES = BUSINESS LOGIC
    Modules are self-contained vertical slices
    Modules own their entities, use cases, adapters
    Modules register handlers, policies, permissions
    Modules can be enabled/disabled via config
    Modules do NOT import from internal/domain or internal/usecase
    Modules do NOT depend on each other directly (use events)

3. PLUGIN ARCHITECTURE
    Dynamic loading via module registry
    Dependency resolution (topological sort)
    Lifecycle management (init → start → stop)
    Auto-registration via init()

4. CONFIGURATION AUTONOMY
    Core: config/app.*.yaml (infrastructure only)
    Modules: internal/modules/{name}/config/config.*.yaml
    Each module loads its own config
    Environment-specific configs (dev, test, prod)

5. LOOSE COUPLING
    Event-driven communication (pub/sub)
    Registry pattern (handlers, policies, permissions)
    Interface-based dependencies
    No direct module-to-module imports

6. LICENSING SUPPORT
    Per-module license keys
    License validation in Initialize()
    Graceful degradation if license invalid


═══════════════════════════════════════════════════════════════════════
                    BENEFITS
═══════════════════════════════════════════════════════════════════════

 MODULARITY
   - Add new modules without touching core
   - Remove modules without breaking others
   - Test modules independently

 SCALABILITY
   - Commercial modules (warehouse, fleet, finance)
   - License-based feature enablement
   - Easy to add new verticals

 MAINTAINABILITY
   - Clear boundaries (core vs modules)
   - Single responsibility (each module owns its domain)
   - Configuration clarity (no monolithic config)

 TESTABILITY
   - Unit test modules in isolation
   - Integration test with real/mock core
   - Smoke test critical flows

 DEPLOYABILITY
   - Enable only needed modules per deployment
   - A/B testing of new modules
   - Gradual rollout of features


═══════════════════════════════════════════════════════════════════════
                    CURRENT STATUS
═══════════════════════════════════════════════════════════════════════

Core:
 Infrastructure (DB, EventBus, Scheduler, Config, Logger)
 Security (Auth, RBAC, JWT, Sessions)
 Reference Data (Countries, Currencies, Timezones, Languages)
 Module Management (Registry, Lifecycle, Config Loader)
 Purge Orchestration (Registry-based, no entity knowledge)

Modules:
 Posts (posts + comments + likes) - Fully independent
 Profiles (profiles + contacts) - Fully independent
 Warehouse (inventory management) - Commercial, disabled

Architecture Compliance:
 Core contains ONLY infrastructure + reference data
 Modules are FULLY independent (no core imports)
 Configuration is AUTONOMOUS (each module owns config)
 Licensing is SUPPORTED (ready for commercial modules)

Recent Improvements:
 Purge system refactored (core = orchestrator, modules = workers)
 Profiles/Contacts migrated to module (removed from core)
 15,000+ lines of code removed from core
 Complete module independence achieved
```
