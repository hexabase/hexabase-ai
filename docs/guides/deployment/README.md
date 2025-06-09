# Deployment Guides

This directory contains guides for deploying and operating the Hexabase AI platform.

## 📚 Available Guides

### [Kubernetes Deployment](./kubernetes-deployment.md)
Basic deployment guide for Kubernetes environments, covering:
- Prerequisites
- Helm chart installation
- Configuration options
- Basic troubleshooting

### [Production Setup](./production-setup.md)
Comprehensive production deployment guide, including:
- Infrastructure requirements and sizing
- High availability configuration
- K3s cluster setup
- Core component installation
- Security hardening
- Post-installation procedures

### [Monitoring Setup](./monitoring-setup.md)
Complete monitoring and observability setup:
- Prometheus and Grafana installation
- Log aggregation with Loki
- ClickHouse for long-term storage
- Custom dashboards and alerts
- SLI/SLO monitoring

### [Backup & Recovery](./backup-recovery.md)
Backup and disaster recovery procedures:
- Backup strategy and components
- Database backup configuration
- Kubernetes resource backups
- Disaster recovery procedures
- Recovery testing automation

### [Deployment Policies](./deployment-policies.md)
Organizational policies and procedures for deployments:
- Deployment approval process
- Change management
- Rollback procedures
- Maintenance windows

## 🚀 Getting Started

If you're setting up Hexabase AI for the first time:

1. **Development/Testing**: Start with [Kubernetes Deployment](./kubernetes-deployment.md)
2. **Production**: Follow [Production Setup](./production-setup.md)
3. **Monitoring**: Implement [Monitoring Setup](./monitoring-setup.md)
4. **Backup**: Configure [Backup & Recovery](./backup-recovery.md)
5. **Policies**: Review [Deployment Policies](./deployment-policies.md)

## 📋 Deployment Checklist

### Pre-Deployment
- [ ] Infrastructure provisioned and validated
- [ ] Network connectivity verified
- [ ] Storage classes configured
- [ ] DNS entries prepared
- [ ] SSL certificates ready
- [ ] Secrets and credentials generated

### Deployment
- [ ] K3s/Kubernetes cluster installed
- [ ] Core components deployed
- [ ] Database cluster operational
- [ ] Ingress controllers configured
- [ ] Initial admin user created

### Post-Deployment
- [ ] Monitoring stack operational
- [ ] Backup schedules configured
- [ ] Security policies applied
- [ ] Documentation updated
- [ ] Team access configured

## 🏗️ Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Load Balancer │────▶│  Ingress Nginx  │────▶│   Hexabase API  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                          │
                        ┌─────────────────────────────────┴───────────┐
                        │                                             │
                ┌───────▼────────┐  ┌─────────────┐  ┌──────────────▼──┐
                │   PostgreSQL   │  │    Redis    │  │      NATS       │
                │    Cluster     │  │   Cluster   │  │   Message Bus   │
                └────────────────┘  └─────────────┘  └─────────────────┘
```

## 🔧 Common Operations

### Scaling

```bash
# Scale API replicas
kubectl scale deployment hexabase-api -n hexabase-system --replicas=5

# Scale worker nodes
# Add nodes to K3s cluster, they auto-join
```

### Upgrades

```bash
# Backup before upgrade
velero backup create pre-upgrade-$(date +%Y%m%d)

# Upgrade using Helm
helm upgrade hexabase ./charts/hexabase -n hexabase-system
```

### Troubleshooting

```bash
# Check pod status
kubectl get pods -n hexabase-system

# View logs
kubectl logs -n hexabase-system deployment/hexabase-api

# Access metrics
kubectl port-forward -n monitoring svc/grafana 3000:80
```

## 📊 Resource Requirements

### Minimum (Development)
- 3 nodes (2 CPU, 4GB RAM each)
- 100GB storage
- 10Mbps network

### Recommended (Production)
- Control Plane: 3 nodes (8 CPU, 32GB RAM each)
- Workers: 5+ nodes (16 CPU, 64GB RAM each)
- 1TB+ fast SSD storage
- 1Gbps+ network
- Dedicated database nodes

### Enterprise (Large Scale)
- Control Plane: 5 nodes (16 CPU, 64GB RAM each)
- Workers: 20+ nodes (32 CPU, 128GB RAM each)
- Multi-region deployment
- Dedicated infrastructure for each component

## 🔐 Security Considerations

1. **Network Security**
   - Private networks for internal communication
   - Firewall rules properly configured
   - Network policies enforced

2. **Access Control**
   - RBAC configured
   - Service accounts with minimal permissions
   - Regular credential rotation

3. **Data Protection**
   - Encryption at rest
   - Encryption in transit
   - Regular security scanning

4. **Compliance**
   - Audit logging enabled
   - Backup encryption
   - Data residency requirements met

## 📝 Maintenance

### Daily Tasks
- Monitor system health
- Check backup status
- Review error logs

### Weekly Tasks
- Apply security updates
- Review resource usage
- Test disaster recovery

### Monthly Tasks
- Update dependencies
- Capacity planning review
- Security audit

## 🆘 Support

- **Documentation**: Check guides in this directory
- **Community**: [Discord](https://discord.gg/hexabase)
- **Enterprise Support**: support@hexabase.ai
- **Emergency**: +1-xxx-xxx-xxxx (Enterprise only)

## 🔗 Related Documentation

- [Architecture Overview](../../architecture/system-architecture.md)
- [Development Setup](../development/dev-environment-setup.md)
- [API Reference](../../api-reference/README.md)
- [Roadmap](../../roadmap/README.md)