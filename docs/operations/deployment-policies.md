# Deployment Policies

This document defines the deployment policies, procedures, and requirements for each environment in the Hexabase KaaS platform.

## Environment Overview

| Environment | Purpose | URL | Deployment Method | Approval Required |
|-------------|---------|-----|-------------------|-------------------|
| **Local** | Developer testing | `*.localhost` | Manual/Automated | No |
| **Staging** | Pre-production testing | `*.staging.hexabase.ai` | Automated CI/CD | No |
| **Production** | Live service | `*.hexabase.ai` | Automated CI/CD | Yes |

## Local Environment Policies

### Purpose
- Developer testing and debugging
- Feature development
- Unit and integration testing

### Requirements
- **Infrastructure**: Kind/Minikube cluster or Docker Compose
- **Resources**: Minimal (1-2 GB RAM, 2 CPU cores)
- **Data**: Test data only, no production data allowed
- **Secrets**: Development secrets only

### Deployment Process
```bash
# Automated setup
make setup
make dev

# OR Manual deployment
helm install hexabase-kaas ./deployments/helm/hexabase-kaas \
  --namespace hexabase-dev \
  --values deployments/helm/values-local.yaml
```

### Policies
1. **No production data** - Never use real customer data
2. **Ephemeral** - Can be destroyed and recreated at any time
3. **Self-contained** - All dependencies run locally
4. **Fast iteration** - Hot reloading enabled
5. **Simplified security** - Basic auth, self-signed certificates

### Configuration
- Internal databases (PostgreSQL, Redis, NATS)
- No persistence for Redis
- Simplified monitoring
- Debug logging enabled
- Mock external services allowed

## Staging Environment Policies

### Purpose
- Integration testing
- Performance testing
- User acceptance testing (UAT)
- Demo environment

### Requirements
- **Infrastructure**: Dedicated Kubernetes cluster
- **Resources**: 50% of production capacity
- **Data**: Anonymized production-like data
- **Secrets**: Staging-specific secrets (never share with production)

### Deployment Process

#### Automatic Deployment
- Triggered on merge to `develop` branch
- Automated via GitHub Actions/GitLab CI
- No manual approval required

```yaml
# .github/workflows/deploy-staging.yml
on:
  push:
    branches: [develop]
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Staging
        run: |
          helm upgrade --install hexabase-kaas \
            hexabase/hexabase-kaas \
            --namespace hexabase-staging \
            --values deployments/helm/values-staging.yaml \
            --wait
```

### Policies

1. **Automated Testing**
   - All tests must pass before deployment
   - Minimum 70% code coverage
   - Security scanning required
   - Performance benchmarks must be met

2. **Data Management**
   - Production data must be anonymized
   - PII must be scrubbed
   - Retention: 30 days
   - Weekly database refresh from production (anonymized)

3. **Access Control**
   - Basic auth on all endpoints
   - VPN required for direct access
   - Limited to development team and QA

4. **Rollback Policy**
   - Automatic rollback on deployment failure
   - Keep last 5 releases
   - 1-click rollback capability

5. **Monitoring**
   - Same monitoring stack as production
   - Alerts sent to #staging-alerts Slack channel
   - Lower alert thresholds than production

### Configuration
```yaml
# Key staging configurations
replicas:
  api: 2      # Reduced from production
  ui: 1       # Single instance
resources:
  reduced: true  # 50% of production
security:
  basic_auth: enabled
  debug_endpoints: enabled
monitoring:
  retention: 7d  # Shorter retention
backups:
  frequency: weekly
```

### Testing Requirements
- [ ] Unit tests pass (100%)
- [ ] Integration tests pass (100%)
- [ ] E2E tests pass (critical paths)
- [ ] Security scan clean
- [ ] Performance within 20% of targets

## Production Environment Policies

### Purpose
- Live customer service
- Revenue-generating operations
- SLA commitments

### Requirements
- **Infrastructure**: High-availability Kubernetes cluster
- **Resources**: Full capacity with auto-scaling
- **Data**: Real customer data with encryption
- **Secrets**: Production secrets in HashiCorp Vault or Kubernetes Secrets

### Deployment Process

#### Approval Workflow
1. **Release Creation**
   - Tag release in Git: `v1.2.3`
   - Create release notes
   - Generate changelog

2. **Approval Required**
   - Engineering Manager approval
   - DevOps team review
   - Security team sign-off (for security updates)

3. **Deployment Window**
   - Preferred: Tuesday-Thursday, 10 AM - 3 PM PST
   - Emergency fixes: Anytime with incident declaration

4. **Deployment Steps**
```yaml
# Production deployment via CI/CD
on:
  release:
    types: [published]

jobs:
  deploy-production:
    environment: production
    runs-on: ubuntu-latest
    steps:
      - name: Require Approval
        uses: trstringer/manual-approval@v1
        with:
          approvers: engineering-managers,devops-team
          
      - name: Deploy to Production
        run: |
          helm upgrade --install hexabase-kaas \
            hexabase/hexabase-kaas \
            --namespace hexabase-system \
            --values deployments/helm/values-production.yaml \
            --atomic \
            --timeout 10m
```

### Policies

1. **Change Management**
   - All changes require PR review (2 approvers minimum)
   - Changes must be tested in staging first
   - Database migrations require separate approval
   - Infrastructure changes require CAB approval

2. **Security Requirements**
   - TLS 1.2+ for all communications
   - Encryption at rest for all data
   - WAF enabled
   - DDoS protection active
   - Regular security audits
   - Penetration testing quarterly

3. **High Availability**
   - Minimum 3 replicas for all services
   - Multi-AZ deployment
   - Database replication
   - Redis Sentinel for HA
   - 99.9% uptime SLA

4. **Backup and Recovery**
   - Daily automated backups
   - Point-in-time recovery capability
   - Backup retention: 30 days daily, 12 months monthly
   - Disaster recovery plan tested quarterly
   - RTO: 4 hours, RPO: 1 hour

5. **Monitoring and Alerting**
   - 24/7 monitoring
   - PagerDuty integration
   - Escalation policies defined
   - Runbooks for all alerts
   - SLO/SLI tracking

6. **Rollback Procedures**
   - Blue-green deployment preferred
   - Instant rollback capability
   - Database migration rollback scripts
   - Maximum rollback time: 15 minutes

### Configuration
```yaml
# Production configurations
replicas:
  api: 3-10      # Auto-scaling enabled
  ui: 2-5        # Auto-scaling enabled
resources:
  guaranteed: true  # Resource guarantees
security:
  waf: enabled
  ddos_protection: enabled
  rate_limiting: strict
monitoring:
  retention: 90d
  high_resolution: true
backups:
  frequency: daily
  replication: cross_region
```

### Pre-deployment Checklist
- [ ] All tests pass in staging
- [ ] Security scan completed
- [ ] Performance benchmarks met
- [ ] Release notes prepared
- [ ] Rollback plan documented
- [ ] Customer communication sent (if needed)
- [ ] On-call engineer assigned
- [ ] Monitoring dashboards ready

### Post-deployment Checklist
- [ ] Health checks passing
- [ ] Metrics within normal range
- [ ] No error rate increase
- [ ] Customer reports verified
- [ ] Performance metrics stable
- [ ] Security scans clean

## Emergency Procedures

### Hotfix Process
1. Declare incident in PagerDuty
2. Create hotfix branch from production
3. Minimal testing in staging
4. Emergency approval (1 senior engineer)
5. Deploy with increased monitoring
6. Post-mortem within 48 hours

### Rollback Triggers
- Error rate > 5%
- Response time > 2x baseline
- Health checks failing
- Customer-impacting bugs
- Security vulnerabilities

## Compliance and Auditing

### Audit Requirements
- All deployments logged with:
  - Who deployed
  - What was deployed
  - When it was deployed
  - Approval trail
  - Configuration changes

### Compliance Checks
- SOC 2 compliance
- GDPR requirements
- HIPAA (if applicable)
- PCI DSS (for payment processing)

### Regular Reviews
- Monthly deployment metrics review
- Quarterly policy review
- Annual security audit
- Continuous compliance monitoring

## Version Management

### Versioning Strategy
- Semantic versioning (MAJOR.MINOR.PATCH)
- Breaking changes require major version bump
- New features require minor version bump
- Bug fixes require patch version bump

### Version Support
- Latest version: Full support
- Previous minor version: Security updates only
- Older versions: Best effort support
- EOL notice: 6 months advance

## Communication

### Deployment Notifications

#### Staging
- Slack: #deployments-staging
- No customer communication

#### Production
- Slack: #deployments-production
- Status page update for major changes
- Email notification for breaking changes
- In-app notification for new features

### Incident Communication
- Status page update within 5 minutes
- Customer email within 30 minutes
- Post-mortem published within 5 days

## Metrics and KPIs

### Deployment Metrics
- Deployment frequency
- Lead time for changes
- Mean time to recovery (MTTR)
- Change failure rate

### Targets
- **Local**: Unlimited deployments
- **Staging**: 10+ deployments per day
- **Production**: 2-5 deployments per week
- **MTTR**: < 30 minutes
- **Change failure rate**: < 5%

## Tools and Technologies

### CI/CD Pipeline
- **Source Control**: Git (GitHub/GitLab)
- **CI/CD**: GitHub Actions / GitLab CI / Tekton
- **Container Registry**: Harbor / DockerHub / ECR
- **Helm Repository**: ChartMuseum / Harbor
- **Secret Management**: HashiCorp Vault / Sealed Secrets

### Deployment Tools
- **Orchestration**: Kubernetes
- **Package Manager**: Helm
- **GitOps**: Flux / ArgoCD
- **Service Mesh**: Istio (optional)

### Monitoring Stack
- **Metrics**: Prometheus + Grafana
- **Logs**: Loki + Grafana
- **Traces**: Jaeger / Tempo
- **Alerts**: AlertManager + PagerDuty

## Policy Enforcement

### Automated Checks
- Pre-commit hooks
- CI/CD pipeline gates
- Admission controllers (OPA/Kyverno)
- Security scanning (Trivy)
- Policy as Code

### Manual Reviews
- Code review requirements
- Architecture review for major changes
- Security review for sensitive changes
- Performance review for critical paths