# Kubernetes Deployment

**Status**: **Phase 2** - Planned for Q3 2026  
**Best for**: Production, multi-node clusters, auto-scaling

---

## Overview

Kubernetes deployment for Promenade with:

- **Helm charts** for easy deployment
- **Horizontal Pod Autoscaling** (HPA)
- **Rolling updates** (zero-downtime deployments)
- **Health checks** (liveness, readiness probes)
- **Secrets management** (K8s secrets, Vault)

---

## Prerequisites

```bash
# Install kubectl
brew install kubectl  # macOS
apt install kubectl   # Ubuntu

# Install Helm
brew install helm  # macOS
apt install helm   # Ubuntu

# Configure kubectl context
kubectl config use-context production-cluster
```

---

## Deployment Architecture

**Components**:

| Component     | Type           | Count                | Purpose                        |
| ------------- | -------------- | -------------------- | ------------------------------ |
| API Server    | Deployment/Pod | 3+ (auto-scaled)     | Application servers            |
| Load Balancer | Service        | 1                    | Traffic distribution (ALB/NLB) |
| PostgreSQL    | StatefulSet    | 1 primary + replicas | Database persistence           |
| Redis         | StatefulSet    | 1 primary + replicas | Caching layer                  |

**Traffic Flow**:

1. External traffic → Load Balancer (Service)
2. Load Balancer → API Server Pods (round-robin)
3. API Server → PostgreSQL (primary)
4. API Server → Redis (cache layer)

---

## Helm Chart Structure (Planned)

```
deploy/kubernetes/promenade/
 Chart.yaml
 values.yaml
 templates/
    deployment.yaml
    service.yaml
    ingress.yaml
    configmap.yaml
    secret.yaml
    hpa.yaml  (Horizontal Pod Autoscaler)
    pdb.yaml  (Pod Disruption Budget)
```

---

## Installation (Phase 2)

```bash
# 1. Add Helm repo (Phase 2)
helm repo add promenade https://charts.promenade.io
helm repo update

# 2. Install Promenade
helm install promenade promenade/promenade \
  --namespace promenade \
  --create-namespace \
  --values values-production.yaml

# 3. Check status
kubectl get pods -n promenade
kubectl get services -n promenade

# 4. Access API
kubectl port-forward service/promenade-api 8080:80 -n promenade
```

---

## Configuration (values.yaml)

```yaml
# Replica count (auto-scaling)
replicaCount: 3

image:
  repository: promenade/api
  tag: v0.1.0
  pullPolicy: IfNotPresent

service:
  type: LoadBalancer
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: api.promenade.io
      paths:
        - path: /
          pathType: Prefix

resources:
  limits:
    cpu: 2000m
    memory: 2Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

postgresql:
  enabled: true
  auth:
    username: promenade
    password: ***********
    database: promenade
  primary:
    persistence:
      size: 50Gi
    resources:
      limits:
        cpu: 4000m
        memory: 8Gi

redis:
  enabled: true
  master:
    persistence:
      size: 2Gi
    resources:
      limits:
        cpu: 1000m
        memory: 2Gi
```

---

## Health Checks

### Liveness Probe

```yaml
# deployment.yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 2
```

---

## Scaling

### Manual Scaling

```bash
# Scale to 5 replicas
kubectl scale deployment/promenade-api --replicas=5 -n promenade

# Check status
kubectl get pods -n promenade
```

### Auto-Scaling (HPA)

```bash
# HPA automatically scales based on CPU/memory
kubectl get hpa -n promenade

# Example output:
# NAME            REFERENCE                  TARGETS   MINPODS   MAXPODS   REPLICAS
# promenade-api   Deployment/promenade-api   45%/70%   3         10        5
```

---

## Rolling Updates

```bash
# Update image
kubectl set image deployment/promenade-api promenade-api=promenade/api:v0.2.0 -n promenade

# Check rollout status
kubectl rollout status deployment/promenade-api -n promenade

# Rollback if needed
kubectl rollout undo deployment/promenade-api -n promenade
```

---

## Secrets Management

### Kubernetes Secrets

```bash
# Create secret
kubectl create secret generic promenade-db \
  --from-literal=username=promenade \
  --from-literal=password=********** \
  -n promenade

# Use in deployment
env:
  - name: DB_USER
    valueFrom:
      secretKeyRef:
        name: promenade-db
        key: username
```

### External Secrets (Vault)

```yaml
# Phase 3: Vault integration
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: promenade-secrets
spec:
  secretStoreRef:
    name: vault-backend
  target:
    name: promenade-db
  data:
    - secretKey: username
      remoteRef:
        key: secret/promenade/db
        property: username
```

---

## Monitoring (Phase 2)

```bash
# Install Prometheus
helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring

# Install Grafana
helm install grafana grafana/grafana -n monitoring

# Access Grafana
kubectl port-forward service/grafana 3000:80 -n monitoring
```

See: [Monitoring Guide](monitoring.md)

---

## Next Steps

- [ ] **Phase 2A**: Create Helm charts
- [ ] **Phase 2B**: Add HPA configuration
- [ ] **Phase 2C**: Setup Ingress (NGINX/Traefik)
- [ ] **Phase 3**: Multi-region deployment (geo-replication)

---

## Related Documentation

- [Docker Compose Guide](docker-compose.md) - Local development
- [AWS Guide](aws.md) - AWS EKS deployment
- [Monitoring Guide](monitoring.md) - Prometheus + Grafana
