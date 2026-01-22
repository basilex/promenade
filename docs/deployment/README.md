# Deployment Documentation

Deployment guides for **Promenade Platform** across different environments and cloud providers.

---

## Available Guides

| Guide                               | Target                    | Status         |
| ----------------------------------- | ------------------------- | -------------- |
| [Docker Compose](docker-compose.md) | Local development         |  Implemented |
| [Kubernetes](kubernetes.md)         | Production (K8s clusters) |  Phase 2     |
| [AWS](aws.md)                       | AWS ECS/EKS               |  Phase 2     |
| [Monitoring](monitoring.md)         | Prometheus + Grafana      |  Phase 2     |

---

## Quick Start

### Local Development

```bash
# 1. Configure workspace
make switch-postgres-dev

# 2. Start services (Docker Compose)
make dev

# API available at: http://localhost:8080
```

See: [Docker Compose Guide](docker-compose.md)

---

### Production Deployment (Phase 2)

**Kubernetes** (recommended for production):

- See [Kubernetes Guide](kubernetes.md)
- Helm charts in `deploy/kubernetes/`
- Auto-scaling, rolling updates, health checks

**AWS ECS** (managed containers):

- See [AWS Guide](aws.md)
- Terraform config in `deploy/aws/`
- RDS for PostgreSQL, ElastiCache for Redis

---

## Deployment Checklist

### Pre-Deployment

- [ ] Run full test suite: `make test-all`
- [ ] Run linter: `make lint`
- [ ] Check for vulnerabilities: `make security-scan` (Phase 2)
- [ ] Build Docker image: `make docker-build`
- [ ] Tag image: `docker tag promenade:latest promenade:v0.1.0`

### Post-Deployment

- [ ] Run database migrations: `make migrate`
- [ ] Seed reference data: `make seed-shared`
- [ ] Verify health endpoints: `curl http://api/health`
- [ ] Check logs: `kubectl logs -f deployment/promenade-api`
- [ ] Run smoke tests: `make test-smoke`

---

## Environment Variables

### Required

```bash
# Database
DATABASE_DRIVER=postgres
DB_HOST=postgres.example.com
DB_PORT=5432
DB_NAME=promenade
DB_USER=promenade
DB_PASSWORD=***********

# Redis
REDIS_ADDR=redis.example.com:6379
REDIS_PASSWORD=***********

# Application
ENVIRONMENT=production
PORT=8080
LOG_LEVEL=info
```

### Optional

```bash
# JWT
JWT_SECRET=***********
JWT_EXPIRATION=24h

# CORS
CORS_ORIGINS=https://app.example.com

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

---

## Related Documentation

- [Configuration Guide](../../config/README.md) - Config file structure
- [Migration Guide](../../migrations/README.md) - Database migrations
- [Docker README](../../docker/README.md) - Docker Compose setup
- [Observability Strategy](../guides/observability-strategy.md) - Monitoring & tracing

---

## Next Steps

- [ ] **Phase 2A**: Create Kubernetes Helm charts
- [ ] **Phase 2B**: Create AWS Terraform modules
- [ ] **Phase 2C**: Add Prometheus + Grafana dashboards
- [ ] **Phase 3**: CI/CD pipeline (GitHub Actions → AWS ECS)
