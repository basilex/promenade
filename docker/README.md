# Docker Configuration

Docker setup for Promenade with environment-specific PostgreSQL configurations.

## Structure

````
docker/
 # PostgreSQL configurations
 docker-compose.postgres.dev.yml   # Development
 docker-compose.postgres.test.yml  # Testing
 docker-compose.postgres.prod.yml  # Production

 # Build and initialization
 Dockerfile                         # Production image build
 init-db.postgres.sh               # PostgreSQL initialization

### Development (`docker-compose.postgres.dev.yml`)

**Purpose**: Local development environment with PostgreSQL

**Services**:

- PostgreSQL 16 (port 5432) - Development database
- Redis 7 (port 6379) - Development cache/session store

**Usage**:

```bash
make switch-postgres-dev
make docker-up           # Start PostgreSQL containers (automatically uses correct file)
make docker-down         # Stop containers
make docker-logs         # View logs
````

**Database**: `promenade_dev`

---

### Testing (`docker-compose.postgres.test.yml`)

**Purpose**: Automated testing environment

**Services**:

- PostgreSQL 16 (port 5433) - Test database (different port to avoid conflicts)
- Redis 7 (port 6380) - Test cache/session store

**Usage**:

```bash
make switch-postgres-test
make docker-up          # Uses postgres.test.yml automatically
```

**Database**: `promenade_test`

---

**Database**: `promenade_test`

---

### Production (`docker-compose.postgres.prod.yml`)

**Purpose**: Production deployment

**Services**:

- PostgreSQL 16 (port 5432) - Production database
- Redis 7 (port 6379) - Production cache/session store

**Usage**:

```bash
make switch-postgres-prod
make docker-up          # Uses postgres.prod.yml automatically
```

**Database**: `promenade_prod`

---

## Port Allocation

| Environment     | PostgreSQL | Redis | App  |
| --------------- | ---------- | ----- | ---- |
| **Development** | 5432       | 6379  | 8081 |
| **Testing**     | 5433       | 6380  | -    |
| **Production**  | 5432       | 6379  | 8080 |

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
docker-compose -f docker/docker-compose.postgres.prod.yml down -v
```

**Database not ready**:

- Wait for health check: `make docker-ps` should show "(healthy)"
- Check logs: `make docker-logs`
