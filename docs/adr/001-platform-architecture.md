# ADR-001: Multi-tenant Kubernetes Platform Architecture

**Date**: 2025-06-01  
**Status**: Implemented  
**Authors**: Platform Architecture Team

## 1. Background

Hexabase AI needed to build a Kubernetes-as-a-Service platform that provides isolated environments for multiple tenants while maintaining cost efficiency and operational simplicity. The platform needed to support various workload types including stateless applications, stateful services, scheduled jobs, and serverless functions.

The key challenges were:
- Complete isolation between tenants for security and compliance
- Cost-effective resource utilization
- Support for both shared and dedicated infrastructure
- Easy scaling and management
- Integration with existing Hexabase ecosystem

## 2. Status

**Implemented** - The core platform architecture using K3s as the host cluster and vCluster for tenant isolation has been deployed and is operational.

## 3. Other Options Considered

### Option A: Namespace-based Multi-tenancy
- Use Kubernetes namespaces for tenant isolation
- Network policies for inter-tenant communication control
- ResourceQuotas and LimitRanges for resource management

### Option B: Multiple K8s Clusters
- Deploy separate Kubernetes clusters for each tenant
- Use cluster federation for management
- Direct infrastructure provisioning

### Option C: vCluster-based Virtual Clusters
- Single host K3s cluster
- vCluster for each tenant workspace
- Shared control plane with isolated data planes

## 4. What Was Decided

We chose **Option C: vCluster-based Virtual Clusters** with the following architecture:
- K3s as the lightweight host cluster
- vCluster for complete Kubernetes API isolation per tenant
- Two plans: Shared (multiple vClusters per node) and Dedicated (exclusive nodes)
- Integration with Proxmox for dedicated node provisioning

## 5. Why Did You Choose It?

- **Complete API Isolation**: Each tenant gets their own Kubernetes API server, preventing any cross-tenant API access
- **Cost Efficiency**: Multiple vClusters can run on shared infrastructure
- **Flexibility**: Easy to migrate tenants between shared and dedicated plans
- **Operational Simplicity**: Single host cluster to manage
- **Native Kubernetes**: Tenants get full Kubernetes API compatibility

## 6. Why Didn't You Choose the Other Options?

### Why not Namespace-based Multi-tenancy?
- Insufficient isolation for compliance requirements
- Shared API server creates security risks
- Complex RBAC management
- Limited ability to customize per-tenant

### Why not Multiple K8s Clusters?
- High operational overhead
- Expensive infrastructure requirements
- Complex networking between clusters
- Difficult to manage at scale

## 7. What Has Not Been Decided

- Long-term strategy for cross-region deployment
- Disaster recovery approach for vClusters
- Migration path for very large tenants (>100 nodes)
- Integration with other cloud providers beyond Proxmox

## 8. Considerations

### Security Considerations
- Each vCluster runs with restricted permissions
- Network policies enforce traffic isolation
- Regular security audits of the vCluster implementation

### Performance Considerations
- Monitor vCluster control plane overhead
- Optimize scheduling for shared infrastructure
- Regular capacity planning reviews

### Operational Considerations
- Automated vCluster lifecycle management
- Monitoring and alerting strategy
- Backup and restore procedures

### Future Considerations
- Evaluate newer isolation technologies as they emerge
- Consider contributing improvements back to vCluster project
- Plan for potential migration to CNCF sandbox projects