# Docker Compose Deployment

**Status**:  **Implemented** - Production-ready for single-node deployments  
**Best for**: Development, staging, small-scale production (<1000 users)

---

## Overview

Docker Compose setup for Promenade with:

- **API server** (Gin, Go 1.24)
- **PostgreSQL 16** (database)
- **Redis 7** (caching, rate limiting)

---

## Quick Start

### Development

```bash
# 1. Configure workspace
make switch-postgres-dev

# 2. Start all services
make dev

# API: http://localhost:8080
# PostgreSQL: localhost:5433
# Redis: localhost:6379
```

### Test

```bash
# 1. Configure workspace
make switch-postgres-test

# 2. Start test services
make test-db-start

# 3. Run tests
make test-integration
```

### Production

```bash
# 1. Configure workspace
make switch-postgres-prod

# 2. Start production services
make docker-up

# 3. Run migrations
make migrate

# 4. Start API
make run
```

---

## Docker Compose Files

### Development: `docker/docker-compose.postgres.dev.yml`

```yaml
version: "3.8"

services:
  postgres:
    image: postgres:16-alpine
    container_name: promenade-postgres-dev
    environment:
      POSTGRES_USER: system
      POSTGRES_PASSWORD: passw0rd
      POSTGRES_DB: promenade_dev
    ports:
      - "5433:5432"
    volumes:
      - postgres_data_dev:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U system"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: promenade-redis-dev
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  postgres_data_dev:
```

---

### Test: `docker/docker-compose.postgres.test.yml`

```yaml
version: "3.8"

services:
  postgres:
    image: postgres:16-alpine
    container_name: promenade-postgres-test
    environment:
      POSTGRES_USER: system
      POSTGRES_PASSWORD: passw0rd
      POSTGRES_DB: promenade_test
    ports:
      - "5433:5432"
    tmpfs: # In-memory storage for faster tests
      - /var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U system"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: promenade-redis-test
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
```

---

### Production: `docker/docker-compose.postgres.prod.yml`

```yaml
version: "3.8"

services:
  postgres:
    image: postgres:16-alpine
    container_name: promenade-postgres-prod
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data_prod:/var/lib/postgresql/data
      - ./backups:/backups # Backup directory
    command: >
      postgres
      -c shared_buffers=256MB
      -c max_connections=200
      -c effective_cache_size=1GB
      -c maintenance_work_mem=64MB
      -c checkpoint_completion_target=0.9
      -c wal_buffers=16MB
      -c default_statistics_target=100
      -c random_page_cost=1.1
      -c effective_io_concurrency=200
      -c work_mem=1MB
      -c min_wal_size=1GB
      -c max_wal_size=4GB
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: promenade-redis-prod
    command: >
      redis-server
      --maxmemory 2gb
      --maxmemory-policy allkeys-lru
      --save 900 1
      --save 300 10
      --save 60 10000
      --appendonly yes
    ports:
      - "6379:6379"
    volumes:
      - redis_data_prod:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  postgres_data_prod:
  redis_data_prod:
```

---

## Configuration

### Environment Variables

Create `.env` file:

```bash
# Database
DB_HOST=localhost
DB_PORT=5433
DB_USER=system
DB_PASSWORD=passw0rd
DB_NAME=promenade_dev

# Redis
REDIS_ADDR=localhost:6379

# Application
ENVIRONMENT=development
PORT=8080
LOG_LEVEL=debug
```

Load in application:

```go
import "github.com/joho/godotenv"

func main() {
    godotenv.Load()  // Load .env file
    // ...
}
```

---

## Management Commands

### Start Services

```bash
# Development
make docker-up

# Or manually
docker compose -f docker/docker-compose.postgres.dev.yml up -d
```

### Stop Services

```bash
make docker-down

# Or manually
docker compose -f docker/docker-compose.postgres.dev.yml down
```

### View Logs

```bash
# All services
make docker-logs

# Specific service
docker logs -f promenade-postgres-dev
docker logs -f promenade-redis-dev
```

### Check Status

```bash
make docker-ps

# Or manually
docker ps
```

---

## Database Backups

### Manual Backup

```bash
# Backup to file
docker exec promenade-postgres-prod pg_dump -U system promenade > backup-$(date +%Y%m%d-%H%M%S).sql

# Backup with compression
docker exec promenade-postgres-prod pg_dump -U system promenade | gzip > backup-$(date +%Y%m%d-%H%M%S).sql.gz
```

### Restore Backup

```bash
# Restore from file
docker exec -i promenade-postgres-prod psql -U system promenade < backup-20260122-103000.sql

# Restore from compressed file
gunzip -c backup-20260122-103000.sql.gz | docker exec -i promenade-postgres-prod psql -U system promenade
```

### Automated Backups (cron)

```bash
# /etc/cron.daily/promenade-backup
#!/bin/bash
BACKUP_DIR=/backups/promenade
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

docker exec promenade-postgres-prod pg_dump -U system promenade | gzip > $BACKUP_DIR/backup-$TIMESTAMP.sql.gz

# Keep only last 7 days
find $BACKUP_DIR -name "backup-*.sql.gz" -mtime +7 -delete
```

---

## Scaling Considerations

### When to Use Docker Compose

 **Good for**:

- Development environments
- Staging servers
- Small production deployments (<1000 users)
- Single-node setups

 **Not suitable for**:

- Multi-node clusters (use Kubernetes)
- High availability (no automatic failover)
- Auto-scaling (use Kubernetes HPA)
- Large-scale production (>10K users)

### Migration Path

**Phase 1** (Current): Docker Compose  
**Phase 2** (Q3 2026): Kubernetes + Helm charts  
**Phase 3** (Q4 2026): Multi-region deployment

---

## Troubleshooting

### PostgreSQL Connection Refused

```bash
# Check if running
docker ps | grep postgres

# Check health
docker inspect --format='{{.State.Health.Status}}' promenade-postgres-dev

# View logs
docker logs promenade-postgres-dev
```

### Redis Connection Timeout

```bash
# Test connection
docker exec promenade-redis-dev redis-cli ping

# Check memory usage
docker exec promenade-redis-dev redis-cli info memory
```

### Port Conflicts

```bash
# Check if port in use
lsof -i :5433
lsof -i :6379

# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
```

---

## Related Documentation

- [Docker README](../../docker/README.md) - Dockerfile and init scripts
- [Kubernetes Guide](kubernetes.md) - Production deployment (Phase 2)
- [AWS Guide](aws.md) - Cloud deployment (Phase 2)
