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
# Application Configuration
APP_NAME=Promenade                          # Application display name
APP_URL=http://localhost:8080               # Base URL for links and callbacks

# Server Configuration
SERVER_HOST=0.0.0.0                         # Default: 0.0.0.0 (all interfaces)
SERVER_PORT=8080                            # API port
ENVIRONMENT=production                      # development|staging|production
SERVER_READ_TIMEOUT=15s                     # Optional: request read timeout
SERVER_WRITE_TIMEOUT=15s                    # Optional: response write timeout

# Database Configuration
DB_HOST=postgres                            # Service name in Docker network!
DB_PORT=5432                                # PostgreSQL port
DB_USER=system                              # Database user
DB_PASSWORD=passw0rd                        # Database password
DB_NAME=promenade_prod                      # Database name
DB_SSLMODE=disable                          # SSL mode (require in production)
DB_MAX_OPEN_CONNS=25                        # Optional: max open connections
DB_MAX_IDLE_CONNS=5                         # Optional: max idle connections
DB_CONN_MAX_LIFETIME=5m                     # Optional: connection lifetime

# JWT Configuration
JWT_SECRET=xTV/YnVTg4aoOiNLrLipZZMQfLwZDgDaEMBxzSz6l1s=  # Secret key (change!)
JWT_ACCESS_TTL_MINUTES=15                   # Access token TTL (15 minutes)
JWT_REFRESH_TTL_HOURS=168                   # Refresh token TTL (7 days)

# Event IBus Configuration
BUS_ADAPTER=memory                          # Adapter: "memory" or "redis"
BUS_WORKER_POOL_SIZE=10                     # Worker goroutines for event processing
BUS_BUFFER_SIZE=1000                        # Event queue buffer size
BUS_RETRY_ATTEMPTS=3                        # Retry failed event handlers
BUS_RETRY_DELAY=1s                          # Delay between retries
BUS_RETRY_MAX_DELAY=5s                      # Max retry delay (exponential backoff cap)
BUS_RETRY_MULTIPLIER=2.0                    # Exponential backoff multiplier

# Redis Configuration (when BUS_ADAPTER=redis)
REDIS_HOST=redis                            # Redis service name in Docker network
REDIS_PORT=6379                             # Redis port
REDIS_PASSWORD=                             # Redis password (empty if no auth)
REDIS_DB=0                                  # Redis database number (0-15)
REDIS_POOL_SIZE=10                          # Connection pool size

# Email Configuration (for notifications)
EMAIL_FROM_NAME=Promenade Team              # Sender name
EMAIL_FROM_ADDRESS=noreply@promenade.com    # Sender email

# Purge Configuration (Soft Delete Cleanup)
PURGE_ENABLED=true                          # Enable automatic purge
PURGE_SCHEDULE=0 2 * * *                    # Cron: daily at 2 AM
PURGE_DRY_RUN=false                         # Dry run mode (logs only)
PURGE_BATCH_SIZE=1000                       # Records per batch
PURGE_RETENTION_USER_POSTS=90               # Days to retain deleted posts
PURGE_RETENTION_POST_COMMENTS=30            # Days to retain deleted comments
```

**[!] Security Notes:**

- Change `JWT_SECRET` to strong random value: `openssl rand -base64 32`
- Change `DB_PASSWORD` to strong password
- Set `DB_SSLMODE=require` in production
- Set `BUS_ADAPTER=redis` for multi-instance deployments
- Never commit production secrets to git

**Configuration Priority** (highest to lowest):

1. Environment variables (docker-compose.yml)
2. `.env.{environment}.local` (e.g., `.env.production.local`)
3. `.env.{environment}` (e.g., `.env.production`)
4. `.env.local` (gitignored)
5. `.env` (default values)

See [config.go](../internal/infrastructure/config/config.go) for all available options and defaults.

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
