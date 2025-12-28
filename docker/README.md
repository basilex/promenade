# Docker Configuration

Docker setup for Promenade with environment-specific configurations.

## Structure

```
docker/
 docker-compose.dev.yml   # Development (local)
 docker-compose.test.yml  # Testing (CI/CD)
 docker-compose.prod.yml  # Production (deployment)
 Dockerfile               # Production image build
 init-db.sh              # Database initialization script
```

## Environment Files

### Development (`docker-compose.dev.yml`)

**Purpose**: Local development environment

**Services**:

- PostgreSQL 16 (port 5432) - Development database
- Redis 7 (port 6379) - Development cache/session store

**Usage**:

```bash
make docker-up      # Start containers
make docker-down    # Stop containers
make docker-logs    # View logs
```

**Database**: `promenade_dev`

---

### Testing (`docker-compose.test.yml`)

**Purpose**: Automated testing environment

**Services**:

- PostgreSQL 16 (port 5433) - Test database (different port to avoid conflicts)
- Redis 7 (port 6380) - Test cache/session store

**Usage**:

```bash
make test-db-start  # Start test containers
make test-db-stop   # Stop test containers
```

**Database**: `promenade_test`

---

### Production (`docker-compose.prod.yml`)

**Purpose**: Production deployment

**Services**:

- PostgreSQL 16 (port 5432) - Production database
- Redis 7 (port 6379) - Production cache/session store
- App (port 8080) - Promenade API

**Usage**:

```bash
make docker-build       # Build production image
make docker-prod-up     # Start production stack
make docker-prod-down   # Stop production stack
make docker-prod-logs   # View production logs
```

**Database**: `promenade_prod`

---

## Port Allocation

| Environment | PostgreSQL | Redis | App  |
| ----------- | ---------- | ----- | ---- |
| Development | 5432       | 6379  | 8081 |
| Testing     | 5433       | 6380  | -    |
| Production  | 5432       | 6379  | 8080 |

## Environment Variables

Production compose uses environment variables for security:

```bash
# Database
DB_USER=system
DB_PASSWORD=your-secure-password
DB_NAME=promenade_prod

# JWT
JWT_SECRET=your-jwt-secret

# Redis (optional)
REDIS_PASSWORD=your-redis-password
```

Set these in `.env.production` or export before running:

```bash
export JWT_SECRET="your-secret-key"
make docker-prod-up
```

## Volumes

Each environment has isolated volumes:

**Development**:

- `postgres_data` - Development database
- `redis_data` - Development cache

**Testing**:

- `postgres_test_data` - Test database
- `redis_test_data` - Test cache

**Production**:

- `postgres_prod_data` - Production database
- `redis_prod_data` - Production cache

## Networks

All services in each environment share a common network:

- `promenade-network` (bridge driver)

## Health Checks

All services have health checks:

**PostgreSQL**: `pg_isready` check every 10s
**Redis**: `redis-cli ping` check every 10s

Production app waits for healthy database and Redis before starting.

## Quick Start

### Development

```bash
# Start PostgreSQL + Redis
make docker-up

# Run migrations
make migrate

# Start API (separate terminal)
make run
```

### Testing

```bash
# Start test database
make test-db-start

# Run tests
make test

# Stop test database
make test-db-stop
```

### Production

```bash
# Build production image
make docker-build

# Start full production stack
make docker-prod-up

# Check logs
make docker-prod-logs

# Stop production stack
make docker-prod-down
```

## Troubleshooting

**Port conflicts**:

- Development and production use same ports (5432, 6379)
- Don't run them simultaneously on same machine
- Testing uses different ports (5433, 6380) to avoid conflicts

**Clean slate**:

```bash
# Development
make docker-clean  # Removes volumes

# Testing
make test-db-stop
docker volume rm promenade_postgres_test_data promenade_redis_test_data

# Production
docker-compose -f docker/docker-compose.prod.yml down -v
```

**Database not ready**:

- Wait for health check: `make docker-ps` should show "(healthy)"
- Check logs: `make docker-logs`
