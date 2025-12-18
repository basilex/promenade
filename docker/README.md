# Docker Setup for Promenade

This document describes how to use Docker to run Promenade API.

## Quick Start

```bash
# Build image and start all services
make docker-run

# Or step by step:
make docker-build VERSION=0.1.0 ENV=dev
make docker-up
```

## Services

Docker Compose runs the following services:

1. **PostgreSQL** (port 5432)

   - Database with healthcheck
   - Persistent volume: `postgres_data`
   - **Auto-creates databases:** On first startup, automatically creates:
     - `promenade_prod` (production, default)
     - `promenade_dev` (development)
     - `promenade_test` (testing)
   - Init script: `docker/init-db.sh`

2. **Redis** (port 6379)

   - Cache (ready for future use)
   - Persistent volume: `redis_data`

3. **Migrate**

   - Applies migrations to `promenade_prod` on startup
   - Runs once, then exits
   - For dev database, run manually: `make migrate-up`

4. **App** (port 8080)
   - Promenade API service
   - Depends on postgres and migrate
   - Automatically restarts on failure

## Building the Image

```bash
# With version (recommended)
make docker-build VERSION=0.1.0 ENV=dev
# Creates images: promenade:0.1.0-dev and promenade:latest

# For production
make docker-build VERSION=0.1.0 ENV=prod
# Creates images: promenade:0.1.0-prod and promenade:latest

# Default (without parameters)
make docker-build
# Creates: promenade:0.1.0-dev and promenade:latest
```

## Container Management

```bash
# Start all services
make docker-up

# Stop all services
make docker-down

# Rebuild and start
make docker-run

# Restart containers
make docker-restart

# View logs
make docker-logs

# Show running containers
make docker-ps

# Clean slate (remove containers and volumes)
make docker-clean
```

## ? Verifying Everything Works

After starting, check service availability:

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Swagger UI v1
open http://localhost:8080/api/v1/docs/swagger/index.html

# Swagger UI v2
open http://localhost:8080/api/v2/docs/swagger/index.html
```

## Network Architecture

All services run in a single Docker network `promenade-network`:

- **App** connects to **Postgres** by service name `postgres` (not `localhost`)
- **Migrate** connects to **Postgres** by service name `postgres`
- Ports are exposed to host for development convenience

## Environment Variables

The application uses the following environment variables (see `docker-compose.yml`):

```yaml
# Server
ENVIRONMENT=production
SERVER_PORT=8080

# Database
DB_HOST=postgres          # Service name in Docker network!
DB_PORT=5432
DB_USER=system
DB_PASSWORD=passw0rd
DB_NAME=promenade_prod
DB_SSLMODE=disable

# JWT
JWT_SECRET=xTV/YnVTg4aoOiNLrLipZZMQfLwZDgDaEMBxzSz6l1s=
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168

# Event Bus
BUS_WORKER_POOL_SIZE=10
BUS_BUFFER_SIZE=1000
BUS_RETRY_ATTEMPTS=3
BUS_RETRY_DELAY=1s
```

[!] **Important**: Change `JWT_SECRET` and DB passwords in production!

## Database Initialization

### Automatic Database Creation

PostgreSQL container uses an init script (`docker/init-db.sh`) that automatically creates three databases on **first startup**:

```sql
promenade_prod  -- Production database (migrations applied by migrate service)
promenade_dev   -- Development database (for local `make dev`)
promenade_test  -- Test database (for integration tests)
```

**How it works:**

1. Docker mounts `init-db.sh` to `/docker-entrypoint-initdb.d/`
2. PostgreSQL executes scripts in this directory once during first container initialization
3. Databases are created only if they don't exist
4. Subsequent starts skip init scripts (data persists in `postgres_data` volume)

**Fresh Start:**

```bash
# Remove all containers and volumes
make docker-clean

# Restart - databases will be recreated
make docker-up
```

### Manual Database Operations

For development work with `promenade_dev`:

```bash
# Apply migrations to dev database
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status
```

## [!] Troubleshooting

### Issue: Database "promenade_dev" does not exist

**Cause**: You ran `make clean` or `make docker-clean` which removed the volumes.

**Solution**:

```bash
# Restart Docker services (databases will be auto-created)
make docker-up

# Apply migrations to dev database
make migrate-up

# Start app
make dev
```

### Issue: Connection refused to database

**Solution**: Ensure that in `docker-compose.yml` the application uses `DB_HOST=postgres` (service name), not `localhost`.

### Issue: Migrations didn't apply

```bash
# Check migrate container logs
docker logs promenade_migrate

# Apply migrations manually
docker-compose run --rm migrate \
  -path=/migrations \
  -database="postgresql://system:passw0rd@postgres:5432/promenade_prod?sslmode=disable" \
  up
```

### Issue: Application won't start

```bash
# Check logs
docker logs promenade_app

# Or in real-time
docker logs -f promenade_app

# Enter container for debugging
docker exec -it promenade_app sh
```

### Clean and Recreate

```bash
# Stop and remove containers + volumes
make docker-clean

# Or manually:
docker-compose -f docker/docker-compose.yml down -v

# Remove images
docker rmi promenade:latest promenade:0.1.0-dev

# Rebuild everything
make docker-build && make docker-up
```

## -> Monitoring

```bash
# Container status
docker-compose -f docker/docker-compose.yml ps

# Resource usage
docker stats promenade_app promenade_postgres

# All service logs
docker-compose -f docker/docker-compose.yml logs -f

# Application logs only
docker logs -f promenade_app
```

## Production Deployment

For production:

1. **Change secrets** in `docker-compose.yml` or use Docker secrets/env files
2. **Setup reverse proxy** (Nginx, Traefik) in front of the application
3. **Use external DB** instead of container (AWS RDS, Azure Database, etc.)
4. **Configure logging** to centralized system (ELK, Loki)
5. **Add monitoring** (Prometheus, Grafana)

Example for production:

```bash
# Build production image
make docker-build VERSION=1.0.0 ENV=prod

# Run with production configuration
COMPOSE_FILE=docker-compose.prod.yml make docker-up
```

## -> Additional Information

- [Main README](../README.md)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
