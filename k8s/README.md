# GOTRS Kubernetes Deployment

This directory contains comprehensive Kubernetes deployment manifests for GOTRS, a modern ticketing system. The deployment is production-ready with high availability, auto-scaling, monitoring, and security best practices.

## Architecture Overview

### Components
- **Backend**: Go application with JWT authentication, GraphQL API, and caching
- **Frontend**: Nginx reverse proxy with static asset serving
- **Database**: PostgreSQL with primary/replica setup
- **Cache**: Redis cluster with Sentinel for high availability
- **Monitoring**: Prometheus, Grafana, and exporters for metrics
- **Ingress**: TLS termination, load balancing, and security headers

### High Availability Features
- ✅ Multi-replica deployments with pod disruption budgets
- ✅ Database replication (PostgreSQL primary/replica)
- ✅ Redis Sentinel for automatic failover
- ✅ Horizontal Pod Autoscaler (HPA) with CPU, memory, and custom metrics
- ✅ Health checks and readiness probes
- ✅ Rolling updates with zero downtime

### Security Features
- ✅ Non-root containers (UID 1000)
- ✅ Read-only root filesystems where possible
- ✅ Network policies for service isolation
- ✅ RBAC with minimal required permissions
- ✅ Pod Security Standards (Restricted)
- ✅ TLS encryption with automatic certificate management
- ✅ Security headers and CORS configuration

## Quick Start

### Prerequisites
- Kubernetes cluster (v1.24+)
- kubectl configured to access your cluster
- Ingress controller (nginx recommended)
- StorageClass named `fast-ssd` for persistent volumes
- cert-manager for TLS certificates (optional)

### Deploy GOTRS

1. **Apply the manifests**:
```bash
# Apply in order
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secrets.yaml
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f gotrs-backend.yaml
kubectl apply -f nginx-frontend.yaml
kubectl apply -f ingress.yaml

# Optional: monitoring
kubectl apply -f monitoring.yaml
```

2. **Or use Kustomize for production**:
```bash
# Deploy using kustomize
kubectl apply -k .

# Or with specific overlays
kubectl apply -k overlays/production
```

3. **Verify deployment**:
```bash
# Check all pods are running
kubectl get pods -n gotrs

# Check services
kubectl get services -n gotrs

# Check ingress
kubectl get ingress -n gotrs
```

### Access the Application

- **Web UI**: https://gotrs.local
- **API**: https://api.gotrs.local
- **GraphQL**: https://gotrs.local/graphql
- **Monitoring**: Access Grafana through port-forward or ingress

## Configuration

### Environment Variables

Key configuration is managed through ConfigMaps and Secrets:

#### Database Configuration
- `DB_HOST`: PostgreSQL service endpoint
- `DB_PORT`: Database port (5432)
- `DB_NAME`: Database name (gotrs)
- `DB_USER/DB_PASSWORD`: Database credentials (from secrets)

#### Redis Configuration
- `REDIS_HOST`: Redis service endpoint  
- `REDIS_PORT`: Redis port (6379)
- `REDIS_PASSWORD`: Redis authentication (from secrets)

#### Application Configuration
- `APP_ENV`: Environment (production/staging/development)
- `LOG_LEVEL`: Logging level (info/debug/error)
- `JWT_SECRET`: JWT signing key (from secrets)

### Resource Requirements

#### Minimum Resources (Development)
- **Backend**: 500m CPU, 512Mi memory
- **Frontend**: 100m CPU, 64Mi memory  
- **PostgreSQL**: 500m CPU, 512Mi memory
- **Redis**: 250m CPU, 256Mi memory

#### Production Resources
- **Backend**: 2 CPU, 2Gi memory (auto-scaled 3-20 replicas)
- **Frontend**: 500m CPU, 256Mi memory (2-10 replicas)
- **PostgreSQL**: 2 CPU, 2Gi memory + 20Gi storage
- **Redis**: 1 CPU, 1Gi memory + 10Gi storage

### Storage

All persistent data uses PersistentVolumeClaims:
- **PostgreSQL**: 20Gi per replica (primary + replicas)
- **Redis**: 10Gi per instance
- **File uploads**: 100Gi shared storage (ReadWriteMany)

Default StorageClass: `fast-ssd` (configure based on your cluster)

## Scaling

### Horizontal Pod Autoscaler (HPA)

**Backend HPA**:
- Min replicas: 3
- Max replicas: 20
- CPU target: 70%
- Memory target: 80%
- Custom metrics: HTTP requests per second

**Frontend HPA**:
- Min replicas: 2  
- Max replicas: 10
- CPU target: 60%
- Custom metrics: Nginx requests

### Manual Scaling
```bash
# Scale backend
kubectl scale deployment gotrs-backend -n gotrs --replicas=5

# Scale frontend
kubectl scale deployment nginx-frontend -n gotrs --replicas=3
```

### Database Scaling
- **Read replicas**: Increase `postgres-replica` replicas
- **Vertical scaling**: Update resource requests/limits
- **Storage expansion**: Modify PVC size (if StorageClass supports expansion)

## Monitoring

### Metrics Collection

**Application Metrics** (`/metrics` endpoint):
- HTTP request duration and count
- Database connection pool stats
- Cache hit/miss rates
- Custom business metrics

**Infrastructure Metrics**:
- CPU, memory, disk usage
- Network I/O
- Kubernetes resource utilization
- PostgreSQL query performance
- Redis operations and memory

### Prometheus Targets
- `gotrs-backend:9090` - Application metrics
- `nginx-frontend:9113` - Nginx metrics via nginx-exporter
- `postgres-exporter:9187` - PostgreSQL metrics
- `redis-exporter:9121` - Redis metrics

### Grafana Dashboards
Pre-configured dashboards for:
- Application overview
- Database performance
- Cache performance  
- Infrastructure metrics
- Business KPIs

### Alerts
Production-ready alerts for:
- High error rates (>10% 5xx responses)
- High response times (>2s 95th percentile)  
- Database connection pool exhaustion
- Memory/CPU usage thresholds
- Pod restart frequency

## Security

### Network Security
- **Network Policies**: Restrict inter-pod communication
- **Ingress Security**: TLS termination, rate limiting, security headers
- **Service Mesh Ready**: Compatible with Istio/Linkerd

### Pod Security
- **Security Context**: Non-root user (UID 1000)
- **Read-only root filesystem** where possible
- **Dropped capabilities**: ALL capabilities dropped
- **Pod Security Standards**: Enforced at namespace level

### Secrets Management
- All sensitive data in Kubernetes Secrets
- Base64 encoded secrets (replace with real values)
- Integration ready for external secret managers (HashiCorp Vault, AWS Secrets Manager)

### RBAC
- Minimal required permissions per service account
- Service-specific roles and role bindings
- No cluster-level permissions required

## Backup & Recovery

### Database Backup
```bash
# Manual backup
kubectl exec -n gotrs postgres-primary-0 -- pg_dump -U gotrs gotrs > backup.sql

# Restore
kubectl exec -i -n gotrs postgres-primary-0 -- psql -U gotrs gotrs < backup.sql
```

### Redis Backup
Redis persistence is configured with both RDB snapshots and AOF:
- RDB snapshots every 15 minutes (if changes)
- AOF for point-in-time recovery

### Persistent Volume Snapshots
If your StorageClass supports snapshots:
```bash
# Create volume snapshot
kubectl create volumesnapshot postgres-snapshot --pvc=postgres-storage-postgres-primary-0 -n gotrs
```

## Troubleshooting

### Common Issues

**Pods not starting**:
```bash
# Check pod status
kubectl describe pod <pod-name> -n gotrs

# Check logs
kubectl logs <pod-name> -n gotrs -f
```

**Database connection issues**:
```bash
# Check PostgreSQL logs
kubectl logs postgres-primary-0 -n gotrs

# Test connection
kubectl exec -it gotrs-backend-<id> -n gotrs -- nc -zv postgres-service 5432
```

**Storage issues**:
```bash
# Check PVC status
kubectl get pvc -n gotrs

# Check storage class
kubectl get storageclass
```

**Networking issues**:
```bash
# Test internal connectivity
kubectl exec -it gotrs-backend-<id> -n gotrs -- nslookup postgres-service

# Check network policies
kubectl get networkpolicy -n gotrs
```

### Performance Tuning

**Database Performance**:
- Adjust PostgreSQL config in `postgres-primary-config` ConfigMap
- Monitor slow queries via `pg_stat_statements`
- Consider read replica scaling for read-heavy workloads

**Cache Performance**:
- Monitor Redis memory usage and eviction
- Adjust cache TTL values in application config
- Scale Redis cluster for higher throughput

**Application Performance**:
- Monitor HPA metrics and adjust targets
- Profile application using pprof endpoints
- Optimize database queries based on slow query logs

## Production Checklist

Before deploying to production:

- [ ] Replace default secrets with strong, unique values
- [ ] Configure proper DNS records for ingress hosts  
- [ ] Set up external backup solution for databases
- [ ] Configure monitoring alerts and notification channels
- [ ] Test disaster recovery procedures
- [ ] Review and adjust resource limits based on load testing
- [ ] Enable and configure log aggregation (ELK, Loki, etc.)
- [ ] Set up CI/CD pipeline for automated deployments
- [ ] Configure external secrets management
- [ ] Review security policies and network restrictions
- [ ] Set up SSL certificates (Let's Encrypt or custom CA)
- [ ] Load test the application under expected traffic
- [ ] Document runbook procedures for operations team

## Files Structure

```
k8s/
├── README.md                    # This file
├── namespace.yaml               # Namespace and resource quotas
├── configmap.yaml              # Application configuration
├── secrets.yaml                # Sensitive configuration (replace values!)
├── postgres.yaml               # PostgreSQL primary/replica setup
├── redis.yaml                  # Redis cluster with Sentinel
├── gotrs-backend.yaml          # Go backend deployment with HPA
├── nginx-frontend.yaml         # Nginx frontend with metrics
├── ingress.yaml                # Ingress with TLS and security
├── monitoring.yaml             # Prometheus, Grafana, exporters
├── kustomization.yaml          # Kustomize configuration
├── patches/                    # Environment-specific patches
│   ├── production-resources.yaml
│   └── production-security.yaml
└── transformers/               # Kustomize transformers
    ├── resource-quotas.yaml
    └── pod-security.yaml
```

## Support

For deployment issues:
1. Check the troubleshooting section above
2. Review Kubernetes events: `kubectl get events -n gotrs --sort-by='.lastTimestamp'`
3. Check application logs for specific error messages
4. Verify all prerequisites are met (storage class, ingress controller, etc.)

For application-specific issues, refer to the main GOTRS documentation.