---
title: "Getting Started"
description: "Quick start guide for Promenade"
weight: 5
---

## Quick Start

Get up and running with Promenade in 5 minutes.

### Prerequisites

- **Go 1.25+**
- **Docker & Docker Compose**
- **Make**

### 1. Clone Repository

```bash
git clone https://github.com/basilex/promenade.git
cd promenade
```

### 2. Install Dependencies

```bash
make install
```

### 3. Start PostgreSQL and Redis

```bash
make docker-up
```

### 4. Run Migrations

```bash
make migrate
```

### 5. Start Application

```bash
make dev
```

Server starts on **http://localhost:8081**

### 6. Health Check

```bash
curl http://localhost:8081/api/v1/health
```

### 7. Swagger Documentation

Open http://localhost:8081/api/v1/docs/swagger/index.html

---

## Next Steps

- [Architecture Overview](/docs/architecture)
- [Module Development](/docs/module-development)
- [Database Schema](/docs/database-schema)
