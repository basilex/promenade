# Production Deployment

Complete guide to deploying Promenade Platform in production environments with Docker, cloud hosting, monitoring, and scaling strategies.

---

## Quick Deployment Options

### Option 1: Docker Compose (Single Server)

**Best for**: Small teams, MVP stage, up to 1000 users

```bash
# 1. Clone repository
git clone https://github.com/basilex/promenade.git
cd promenade

# 2. Configure production environment
cp config/app.prod.yaml config/app.prod.local.yaml
vim config/app.prod.local.yaml  # Set DB_PASSWORD, JWT_SECRET, etc.

# 3. Start production stack
docker compose -f docker/docker-compose.prod.yml up -d

# 4. Run migrations
docker exec promenade_api /app/bin/migrate up

# 5. Verify health
curl http://localhost:8081/health
```

### Option 2: Kubernetes (Scalable)

**Best for**: Growth stage, 1000+ users, high availability

```bash
# 1. Apply Kubernetes manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/app.yaml

# 2. Check deployment
kubectl get pods -n promenade
kubectl logs -f deployment/promenade-api -n promenade
```

### Option 3: Cloud Platforms

**Best for**: Quick MVP, managed infrastructure

- **AWS**: ECS Fargate + RDS PostgreSQL + ElastiCache Redis
- **GCP**: Cloud Run + Cloud SQL + Memorystore
- **Azure**: Container Apps + PostgreSQL + Redis Cache

---

## Docker Deployment

### Production Docker Compose

**docker/docker-compose.prod.yml**:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: ${DB_USER:-promenade}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME:-promenade_prod}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U promenade"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  api:
    image: promenade/api:latest
    build:
      context: .
      dockerfile: docker/Dockerfile
    environment:
      ENVIRONMENT: production
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${DB_USER:-promenade}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME:-promenade_prod}
      REDIS_ADDR: redis:6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
    ports:
      - "8081:8081"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:
  redis_data:
```

### Build Production Image

**docker/Dockerfile** (Multi-stage build):
```dockerfile
# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/promenade ./cmd/api

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
WORKDIR /app

COPY --from=builder /app/bin/promenade .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD curl -f http://localhost:8081/health || exit 1

CMD ["./promenade"]
```

### Build and Push Image

```bash
# Build image
docker build -f docker/Dockerfile -t promenade/api:0.1.0 .
docker tag promenade/api:0.1.0 promenade/api:latest

# Push to registry (Docker Hub, ECR, GCR, ACR)
docker push promenade/api:0.1.0
docker push promenade/api:latest
```

---

## Environment Configuration

### Production Config File

**config/app.prod.yaml**:
```yaml
app:
  name: "Promenade Platform"
  environment: "production"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8081
  read_timeout: 30s
  write_timeout: 30s

database:
  postgres:
    host: "${DB_HOST}"
    port: 5432
    user: "${DB_USER}"
    password: "${DB_PASSWORD}"
    database: "${DB_NAME}"
    ssl_mode: "require"
    max_connections: 50
    max_idle_connections: 10
    connection_max_lifetime: 1h
  redis:
    addr: "${REDIS_ADDR}"
    password: "${REDIS_PASSWORD}"
    pool_size: 100
    max_retries: 3
    databases:
      revocation: 0
      bus: 1
      cache: 2
      sessions: 3

jwt:
  secret: "${JWT_SECRET}"
  access_token_duration: 15m
  refresh_token_duration: 168h  # 7 days
  issuer: "promenade-platform"

bus:
  adapter: "redis"  # Distributed event bus
  worker_pool_size: 50
  buffer_size: 10000
  retry_attempts: 5
  retry_delay: 2s
  retry_max_delay: 60s
  retry_multiplier: 2.0

cache:
  enabled: true
  adapter: "redis"
  prefix: "promenade:"
  ttl:
    reference: 86400  # 24 hours
    user: 1800        # 30 minutes
    session: 3600     # 1 hour

logging:
  level: "info"
  format: "json"
  add_source: false
```

### Environment Variables

**Required secrets** (store in `.env.prod` or secrets manager):
```bash
# Database
export DB_HOST=your-db-host.rds.amazonaws.com
export DB_USER=promenade
export DB_PASSWORD=your-secure-password-here
export DB_NAME=promenade_prod

# Redis
export REDIS_ADDR=your-redis.cache.amazonaws.com:6379
export REDIS_PASSWORD=your-redis-password

# JWT
export JWT_SECRET=$(openssl rand -base64 32)

# Optional
export SENTRY_DSN=https://your-sentry-dsn
export SLACK_WEBHOOK=https://hooks.slack.com/services/YOUR/WEBHOOK
```

**Load secrets**:
```bash
# From file (development/staging)
source .env.prod

# From AWS Secrets Manager (production)
aws secretsmanager get-secret-value --secret-id promenade-prod | jq -r '.SecretString'

# From Google Secret Manager
gcloud secrets versions access latest --secret="promenade-prod"

# From Azure Key Vault
az keyvault secret show --vault-name promenade-vault --name prod-secrets
```

---

## Cloud Platform Deployment

### AWS ECS Fargate

**1. Infrastructure Setup**:
```bash
# Create VPC, subnets, security groups
terraform init
terraform plan -var-file=prod.tfvars
terraform apply

# Or use CloudFormation
aws cloudformation create-stack \
  --stack-name promenade-infrastructure \
  --template-body file://cloudformation/infrastructure.yaml
```

**2. Database (RDS PostgreSQL)**:
```bash
# Create database instance
aws rds create-db-instance \
  --db-instance-identifier promenade-prod \
  --db-instance-class db.t3.medium \
  --engine postgres \
  --engine-version 16.1 \
  --master-username promenade \
  --master-user-password $DB_PASSWORD \
  --allocated-storage 100 \
  --storage-type gp3 \
  --backup-retention-period 7 \
  --multi-az \
  --vpc-security-group-ids sg-xxxxxxxx
```

**3. Cache (ElastiCache Redis)**:
```bash
# Create Redis cluster
aws elasticache create-cache-cluster \
  --cache-cluster-id promenade-redis \
  --cache-node-type cache.t3.micro \
  --engine redis \
  --engine-version 7.0 \
  --num-cache-nodes 1 \
  --security-group-ids sg-xxxxxxxx
```

**4. Container (ECS Fargate)**:
```bash
# Create ECS cluster
aws ecs create-cluster --cluster-name promenade-prod

# Register task definition
aws ecs register-task-definition --cli-input-json file://ecs-task-definition.json

# Create service
aws ecs create-service \
  --cluster promenade-prod \
  --service-name promenade-api \
  --task-definition promenade-api:1 \
  --desired-count 2 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}"
```

**5. Load Balancer (ALB)**:
```bash
# Create target group
aws elbv2 create-target-group \
  --name promenade-api-tg \
  --protocol HTTP \
  --port 8081 \
  --vpc-id vpc-xxxxxxxx \
  --target-type ip \
  --health-check-path /health

# Create load balancer
aws elbv2 create-load-balancer \
  --name promenade-alb \
  --subnets subnet-xxx subnet-yyy \
  --security-groups sg-xxxxxxxx \
  --scheme internet-facing
```

### Google Cloud Platform (GCP)

**1. Cloud SQL (PostgreSQL)**:
```bash
# Create database instance
gcloud sql instances create promenade-prod \
  --database-version=POSTGRES_16 \
  --tier=db-custom-2-7680 \
  --region=us-central1 \
  --backup \
  --availability-type=REGIONAL

# Create database
gcloud sql databases create promenade_prod --instance=promenade-prod
```

**2. Memorystore (Redis)**:
```bash
# Create Redis instance
gcloud redis instances create promenade-redis \
  --size=1 \
  --region=us-central1 \
  --redis-version=redis_7_0 \
  --tier=standard
```

**3. Cloud Run**:
```bash
# Deploy service
gcloud run deploy promenade-api \
  --image gcr.io/your-project/promenade:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars ENVIRONMENT=production,DB_HOST=$DB_HOST \
  --set-secrets DB_PASSWORD=promenade-db-password:latest,JWT_SECRET=promenade-jwt:latest \
  --min-instances 1 \
  --max-instances 10 \
  --cpu 2 \
  --memory 4Gi
```

### Azure

**1. Azure Database for PostgreSQL**:
```bash
# Create resource group
az group create --name promenade-prod --location eastus

# Create database server
az postgres flexible-server create \
  --resource-group promenade-prod \
  --name promenade-db \
  --location eastus \
  --admin-user promenade \
  --admin-password $DB_PASSWORD \
  --sku-name Standard_D2s_v3 \
  --tier GeneralPurpose \
  --version 16
```

**2. Azure Cache for Redis**:
```bash
# Create Redis cache
az redis create \
  --resource-group promenade-prod \
  --name promenade-redis \
  --location eastus \
  --sku Standard \
  --vm-size c1
```

**3. Container Apps**:
```bash
# Create container app environment
az containerapp env create \
  --name promenade-env \
  --resource-group promenade-prod \
  --location eastus

# Deploy app
az containerapp create \
  --name promenade-api \
  --resource-group promenade-prod \
  --environment promenade-env \
  --image your-registry.azurecr.io/promenade:latest \
  --target-port 8081 \
  --ingress external \
  --min-replicas 1 \
  --max-replicas 5 \
  --cpu 1 \
  --memory 2Gi
```

---

## Database Migrations in Production

### Pre-Deployment Checklist

1. **Backup database** before migrations
2. **Test migrations** in staging environment
3. **Review migration SQL** for performance impact
4. **Plan rollback** strategy
5. **Schedule downtime** if needed (or use zero-downtime strategy)

### Running Migrations

**Option 1: Manual (via container)**:
```bash
# Connect to running API container
docker exec -it promenade_api bash

# Run migrations
./bin/migrate up

# Check status
./bin/migrate status
```

**Option 2: Automated (CI/CD)**:
```yaml
# .github/workflows/deploy.yml
- name: Run Migrations
  run: |
    docker run --rm \
      -e DB_HOST=${{ secrets.DB_HOST }} \
      -e DB_PASSWORD=${{ secrets.DB_PASSWORD }} \
      promenade/api:${{ github.sha }} \
      /app/bin/migrate up
```

**Option 3: Kubernetes Job**:
```yaml
# k8s/migration-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: promenade-migration
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: promenade/api:latest
        command: ["/app/bin/migrate", "up"]
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: promenade-secrets
              key: db-host
      restartPolicy: OnFailure
```

### Zero-Downtime Migrations

**Strategy**:
1. **Backward-compatible changes**: Add new columns (nullable), new tables
2. **Deploy new code**: App works with old and new schema
3. **Run migration**: Add columns, backfill data
4. **Remove old code**: Clean up deprecated fields

**Example**:
```sql
-- Migration 1: Add new column (nullable)
ALTER TABLE customer_customers ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

-- Deploy new code (reads/writes email_verified)

-- Migration 2: Backfill data
UPDATE customer_customers SET email_verified = TRUE WHERE email IS NOT NULL;

-- Migration 3: Make NOT NULL
ALTER TABLE customer_customers ALTER COLUMN email_verified SET NOT NULL;
```

---

## Monitoring & Observability

### Health Checks

**Kubernetes Probes**:
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8081
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health/db
    port: 8081
  initialDelaySeconds: 10
  periodSeconds: 5
```

**Load Balancer Health**:
- ALB: `/health` every 30s
- NGINX: `health_check` directive
- HAProxy: `option httpchk GET /health`

### Logging

**Centralized Logging** (ELK Stack, CloudWatch, Stackdriver):
```yaml
# Fluent Bit config (ship logs to Elasticsearch)
[OUTPUT]
    Name  es
    Match *
    Host  elasticsearch.prod.internal
    Port  9200
    Index promenade-logs
    Type  _doc
```

**Structured Logs** (JSON format in production):
```json
{
  "timestamp": "2025-12-29T10:15:30Z",
  "level": "error",
  "message": "Failed to create customer",
  "context": {
    "user_id": "01JGABC123",
    "request_id": "req-xyz",
    "error": "database connection timeout"
  }
}
```

### Metrics

**Prometheus Metrics** (future):
```go
// Instrument code with metrics
httpRequestsTotal.WithLabelValues("POST", "/customers").Inc()
httpRequestDuration.WithLabelValues("POST", "/customers").Observe(duration.Seconds())
```

**Grafana Dashboard**:
- Request rate (req/s)
- Response time (p50, p95, p99)
- Error rate (%)
- Database connections (active/idle)
- Redis operations (get/set rate)

### Error Tracking

**Sentry Integration** (future):
```go
import "github.com/getsentry/sentry-go"

sentry.Init(sentry.ClientOptions{
    Dsn: os.Getenv("SENTRY_DSN"),
    Environment: "production",
})

// Capture errors
sentry.CaptureException(err)
```

---

## Scaling Strategies

### Horizontal Scaling (Multiple Instances)

**Benefits**:
- Handle more traffic
- High availability (redundancy)
- Rolling deployments (zero downtime)

**Implementation**:
```bash
# Docker Swarm
docker service scale promenade_api=5

# Kubernetes
kubectl scale deployment promenade-api --replicas=5

# ECS
aws ecs update-service --service promenade-api --desired-count 5
```

**Auto-scaling**:
```yaml
# Kubernetes HPA (Horizontal Pod Autoscaler)
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: promenade-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: promenade-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Database Scaling

**Read Replicas**:
- Route read queries to replicas
- Write queries to primary
- Reduce load on primary database

**Connection Pooling**:
- Use PgBouncer for PostgreSQL
- Max connections: 100 (application pool)
- Database max connections: 200

**Query Optimization**:
- Add indexes for common queries
- Use EXPLAIN ANALYZE for slow queries
- Enable query caching (Redis)

### Redis Scaling

**Cluster Mode**:
- Sharding for large datasets
- High availability with replicas
- Automatic failover

**Configuration**:
```yaml
redis:
  mode: cluster
  master_replicas: 3
  shards: 3
  read_from_replicas: true
```

---

## Security Checklist

### Pre-Production Security

- [ ] **HTTPS Enforced**: TLS 1.3, valid SSL certificate
- [ ] **Secrets Rotation**: Regular rotation of JWT secret, DB password
- [ ] **Database Encryption**: Encryption at rest (RDS, Cloud SQL)
- [ ] **Network Isolation**: Private subnets for DB/Redis
- [ ] **Firewall Rules**: Whitelist only necessary IPs
- [ ] **Rate Limiting**: Protect authentication endpoints (5 req/min)
- [ ] **CORS Configuration**: Allow only trusted domains
- [ ] **SQL Injection Prevention**: Parameterized queries (sqlx)
- [ ] **XSS Protection**: Input validation, output escaping
- [ ] **CSRF Tokens**: For state-changing requests
- [ ] **Audit Logging**: Log sensitive operations (user creation, role changes)
- [ ] **Dependency Scanning**: `go list -m all | nancy` for vulnerabilities

### TLS/SSL Setup

**Let's Encrypt (Free Certificate)**:
```bash
# Install Certbot
sudo apt-get install certbot

# Generate certificate
sudo certbot certonly --standalone -d api.promenade.com

# Configure NGINX
server {
    listen 443 ssl;
    server_name api.promenade.com;
    
    ssl_certificate /etc/letsencrypt/live/api.promenade.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.promenade.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:8081;
    }
}
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

**.github/workflows/deploy.yml**:
```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      - name: Run tests
        run: make test
      
      - name: Run linter
        run: make lint

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker image
        run: docker build -f docker/Dockerfile -t promenade/api:${{ github.sha }} .
      
      - name: Push to registry
        run: |
          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker push promenade/api:${{ github.sha }}
          docker tag promenade/api:${{ github.sha }} promenade/api:latest
          docker push promenade/api:latest

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to ECS
        run: |
          aws ecs update-service \
            --cluster promenade-prod \
            --service promenade-api \
            --force-new-deployment
```

---

## Backup & Disaster Recovery

### Database Backups

**Automated Backups**:
```bash
# RDS PostgreSQL (AWS)
aws rds modify-db-instance \
  --db-instance-identifier promenade-prod \
  --backup-retention-period 30 \
  --preferred-backup-window "03:00-04:00"

# Cloud SQL (GCP)
gcloud sql instances patch promenade-prod \
  --backup-start-time=03:00
```

**Manual Backup**:
```bash
# pg_dump backup
pg_dump -h your-db-host.com -U promenade promenade_prod > backup_$(date +%Y%m%d).sql

# Upload to S3
aws s3 cp backup_$(date +%Y%m%d).sql s3://promenade-backups/
```

**Restore**:
```bash
# Download backup
aws s3 cp s3://promenade-backups/backup_20250129.sql .

# Restore database
psql -h your-db-host.com -U promenade promenade_prod < backup_20250129.sql
```

### Point-in-Time Recovery

**Enable PITR** (RDS, Cloud SQL):
- Restore to any second within retention period
- Useful for accidental data deletion
- Retention: 7-35 days

---

## Cost Optimization

### Resource Sizing

**Start Small**:
- API: 2 vCPU, 4GB RAM (1-2 instances)
- DB: db.t3.medium (2 vCPU, 4GB RAM)
- Redis: cache.t3.micro (1GB RAM)
- **Est. Cost**: $200-300/month (AWS)

**Scale as Needed**:
- Monitor CPU/Memory usage
- Upgrade when consistently > 70%
- Use auto-scaling to save costs

### Cost-Saving Tips

1. **Reserved Instances**: 30-60% savings (1-year commitment)
2. **Spot Instances**: 70% savings for non-critical workloads
3. **S3 Intelligent Tiering**: Auto-move cold data to cheaper storage
4. **CloudFront CDN**: Cache static assets, reduce bandwidth
5. **Lambda Functions**: Use for infrequent tasks (migrations, cron jobs)

---

## Troubleshooting

### Common Issues

**1. Database Connection Failures**
```bash
# Check database reachability
psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# Check security group rules (AWS)
aws ec2 describe-security-groups --group-ids sg-xxxxxxxx

# Check logs
docker logs promenade_api | grep "database"
```

**2. Redis Connection Errors**
```bash
# Test Redis connection
redis-cli -h $REDIS_ADDR -a $REDIS_PASSWORD ping

# Check Redis memory usage
redis-cli -h $REDIS_ADDR -a $REDIS_PASSWORD INFO memory
```

**3. High CPU/Memory Usage**
```bash
# Check container stats
docker stats promenade_api

# Kubernetes
kubectl top pods -n promenade
```

**4. Slow API Responses**
```bash
# Enable query logging (PostgreSQL)
ALTER DATABASE promenade_prod SET log_min_duration_statement = 1000;  # Log queries > 1s

# Check slow queries
SELECT query, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;
```

---

## Related Documentation

- [Getting Started](getting-started.md) - Local development setup
- [Development Workflow](development-workflow.md) - Git workflow, testing
- [Health Checks](health-checks.md) - Health monitoring
- [Configuration Reference](../reference/configuration.md) - Config options
- [API Reference](../reference/api-reference.md) - REST API docs

---

**Questions?** [Open a discussion](https://github.com/basilex/promenade/discussions)

**Need help deploying?** Email: alexander.vasilenko@gmail.com
