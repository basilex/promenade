# Getting Started

Welcome to **Promenade Platform** - a modern backend platform built with Domain-Driven Design, Event-Driven Architecture, and Clean Patterns.

## What is Promenade?

Promenade is **not a traditional CRM** - it's a **modular platform** that grows with your needs:

- Start with **Customer Management**
- Add **Order Processing** when needed
- Integrate **Billing** when ready
- Extend with **custom contexts**

## Prerequisites

- **Go 1.23+** - [Install Go](https://go.dev/doc/install)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Make** - Build automation tool

## Installation

### 1. Clone Repository

```bash
git clone https://github.com/basilex/promenade.git
cd promenade
```

### 2. Start Dependencies

```bash
# Start PostgreSQL (localhost:5432) + Redis (localhost:6379)
make docker-up
```

### 3. Run Migrations

```bash
# Migrations run automatically on app startup
# Or manually:
make migrate
```

### 4. Start Application

```bash
# Development mode (Docker + migrations + API)
make dev

# Or build and run separately
make build
./bin/promenade
```

Server starts on **http://localhost:8081**

## Quick Verification

### Health Check

```bash
curl http://localhost:8081/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-12-29T12:00:00Z"
}
```

### API Authentication

Register a new user:

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

Login to get JWT tokens:

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

Use access token:

```bash
curl -X GET http://localhost:8081/api/v1/identity/users/me \
  -H "Authorization: Bearer <access_token>"
```

## Project Structure

```
promenade/
├── cmd/api/              # Application entry point
├── internal/
│   ├── contexts/         # Bounded Contexts (DDD)
│   │   ├── identity/     # User, Contact, Profile
│   │   ├── shared/       # Reference data
│   │   └── customer-mgmt/# Customer, Company, Deal
│   └── infrastructure/   # Config, Database
├── pkg/                  # Shared packages
│   ├── bus/             # Event Bus
│   ├── jwt/             # Authentication
│   └── uuidv7/          # Time-ordered UUIDs
├── migrations/          # Database migrations
└── test/                # Three-tier testing
```

## Available Commands

```bash
# Development
make dev               # Full dev environment
make build             # Build binary
make run              # Build and run

# Testing
make test              # All tests (~40s)
make test-unit         # Unit tests (~5s)
make test-smoke        # Smoke tests (~0.3s)
make test-integration  # Integration tests (~5s)

# Database
make migrate           # Run all migrations
make db-reset          # Fresh database

# Code Quality
make fmt               # Format code
make lint              # Run linters
```

## Next Steps

- [Quick Start Tutorial](/guide/quick-start) - Build your first aggregate
- [Architecture Guide](/guide/architecture) - Understand DDD concepts
- [Testing Strategy](/guide/testing) - Write professional tests
- [API Reference](/guide/api-reference) - Explore all endpoints

## Configuration

Edit `config/app.dev.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8081

database:
  postgres:
    host: "localhost"
    port: 5432
    database: "promenade_dev"

jwt:
  secret: "your-secret-key-at-least-32-characters"
  access_token_duration: 15m
  refresh_token_duration: 168h  # 7 days

bus:
  adapter: "memory"  # or "redis" for production
```

## Troubleshooting

### Port Already in Use

```bash
# Stop existing PostgreSQL
make docker-down

# Start fresh
make docker-up
```

### Database Connection Failed

Check PostgreSQL is running:
```bash
make docker-ps
```

### Migration Errors

Reset database:
```bash
make db-reset
make migrate
```

## Support

- **Documentation**: [GitHub Docs](https://github.com/basilex/promenade/tree/dev/docs)
- **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Email**: alexander.vasilenko@gmail.com
