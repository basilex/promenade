# AWS Deployment

**Status**: **Phase 2** - Planned for Q3 2026  
**Best for**: Managed cloud infrastructure, auto-scaling, multi-region

---

## Overview

AWS deployment options for Promenade:

- **AWS ECS** (Elastic Container Service) - Managed containers, simple setup
- **AWS EKS** (Elastic Kubernetes Service) - Managed Kubernetes, advanced features
- **AWS RDS** (PostgreSQL) - Managed database
- **AWS ElastiCache** (Redis) - Managed caching

---

## Architecture (ECS)

**AWS Components**:

| Component     | AWS Service                     | Purpose                                     |
| ------------- | ------------------------------- | ------------------------------------------- |
| Load Balancer | Application Load Balancer (ALB) | HTTPS/SSL termination, traffic distribution |
| API Servers   | ECS Tasks (Fargate)             | Application runtime (3+ tasks)              |
| Database      | RDS PostgreSQL 16               | Managed database with backups               |
| Cache         | ElastiCache Redis 7             | Managed caching layer                       |
| Networking    | VPC + Subnets                   | Network isolation                           |
| Secrets       | AWS Secrets Manager             | Credential management                       |

**Traffic Flow**:

1. Internet → ALB (HTTPS port 443)
2. ALB → ECS Tasks (port 8080)
3. ECS Tasks → RDS PostgreSQL (port 5432)
4. ECS Tasks → ElastiCache Redis (port 6379)

---

## Prerequisites

```bash
# Install AWS CLI
brew install awscli  # macOS
apt install awscli   # Ubuntu

# Configure credentials
aws configure
# AWS Access Key ID: ***********
# AWS Secret Access Key: ***********
# Default region: us-east-1
# Default output format: json

# Install Terraform (infrastructure as code)
brew install terraform  # macOS
apt install terraform   # Ubuntu
```

---

## Terraform Configuration (Phase 2)

### Project Structure

```
deploy/aws/
 main.tf
 variables.tf
 outputs.tf
 modules/
    ecs/
       main.tf
       variables.tf
       outputs.tf
    rds/
       main.tf
       variables.tf
       outputs.tf
    elasticache/
        main.tf
        variables.tf
        outputs.tf
```

### main.tf (Example)

```hcl
# deploy/aws/main.tf
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# VPC
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "promenade-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b", "us-east-1c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  enable_vpn_gateway = false
}

# RDS (PostgreSQL)
module "rds" {
  source = "./modules/rds"

  vpc_id              = module.vpc.vpc_id
  private_subnet_ids  = module.vpc.private_subnets
  instance_class      = "db.t3.medium"
  allocated_storage   = 100
  engine_version      = "16.1"
  database_name       = "promenade"
  master_username     = var.db_username
  master_password     = var.db_password
}

# ElastiCache (Redis)
module "elasticache" {
  source = "./modules/elasticache"

  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnets
  node_type          = "cache.t3.medium"
  num_cache_nodes    = 2
}

# ECS Cluster
module "ecs" {
  source = "./modules/ecs"

  cluster_name       = "promenade-cluster"
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnets
  public_subnet_ids  = module.vpc.public_subnets

  container_image    = "promenade/api:v0.1.0"
  container_port     = 8080
  desired_count      = 3

  db_host     = module.rds.endpoint
  redis_addr  = module.elasticache.endpoint

  environment_variables = {
    ENVIRONMENT = "production"
    LOG_LEVEL   = "info"
  }
}
```

---

## Deployment (Terraform)

```bash
# 1. Initialize Terraform
cd deploy/aws
terraform init

# 2. Plan deployment
terraform plan -out=tfplan

# 3. Apply deployment
terraform apply tfplan

# 4. Get outputs
terraform output api_url
terraform output db_endpoint
terraform output redis_endpoint
```

---

## ECS Task Definition (Example)

```json
{
  "family": "promenade-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "containerDefinitions": [
    {
      "name": "promenade-api",
      "image": "promenade/api:v0.1.0",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        { "name": "ENVIRONMENT", "value": "production" },
        { "name": "PORT", "value": "8080" },
        { "name": "LOG_LEVEL", "value": "info" }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789:secret:promenade/db-password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/promenade-api",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

---

## Auto-Scaling

```hcl
# Auto-scaling target
resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = 10
  min_capacity       = 3
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.api.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# Scale up policy
resource "aws_appautoscaling_policy" "scale_up" {
  name               = "promenade-scale-up"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = 70.0
  }
}
```

---

## Cost Estimation

### Small Deployment (Development)

| Service               | Type                      | Monthly Cost   |
| --------------------- | ------------------------- | -------------- |
| ECS Fargate (2 tasks) | 1vCPU, 2GB RAM            | $50            |
| RDS (PostgreSQL)      | db.t3.small               | $30            |
| ElastiCache (Redis)   | cache.t3.micro            | $15            |
| ALB                   | Application Load Balancer | $25            |
| **Total**             |                           | **$120/month** |

### Medium Deployment (Production)

| Service               | Type                      | Monthly Cost   |
| --------------------- | ------------------------- | -------------- |
| ECS Fargate (5 tasks) | 2vCPU, 4GB RAM            | $350           |
| RDS (PostgreSQL)      | db.t3.medium              | $120           |
| ElastiCache (Redis)   | cache.t3.medium           | $80            |
| ALB                   | Application Load Balancer | $25            |
| CloudWatch            | Logs + Metrics            | $30            |
| **Total**             |                           | **$605/month** |

---

## Monitoring (CloudWatch)

```bash
# View logs
aws logs tail /ecs/promenade-api --follow

# Get metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=promenade-api \
  --start-time 2026-01-22T00:00:00Z \
  --end-time 2026-01-22T23:59:59Z \
  --period 300 \
  --statistics Average
```

---

## CI/CD (GitHub Actions)

```yaml
# .github/workflows/deploy-aws.yml
name: Deploy to AWS ECS

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v1

      - name: Build and push image
        run: |
          docker build -t promenade-api .
          docker tag promenade-api:latest ${{ steps.login-ecr.outputs.registry }}/promenade-api:${{ github.sha }}
          docker push ${{ steps.login-ecr.outputs.registry }}/promenade-api:${{ github.sha }}

      - name: Deploy to ECS
        uses: aws-actions/amazon-ecs-deploy-task-definition@v1
        with:
          task-definition: task-definition.json
          service: promenade-api
          cluster: promenade-cluster
          wait-for-service-stability: true
```

---

## Next Steps

- [ ] **Phase 2A**: Create Terraform modules
- [ ] **Phase 2B**: Setup CI/CD pipeline (GitHub Actions → ECS)
- [ ] **Phase 2C**: Add CloudWatch dashboards
- [ ] **Phase 3**: Multi-region deployment (us-east-1, eu-west-1)

---

## Related Documentation

- [Docker Compose Guide](docker-compose.md) - Local development
- [Kubernetes Guide](kubernetes.md) - K8s deployment
- [Monitoring Guide](monitoring.md) - CloudWatch + Grafana
