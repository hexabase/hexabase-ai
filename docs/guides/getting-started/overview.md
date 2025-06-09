# Hexabase AI Overview

Welcome to Hexabase AI - the next-generation Kubernetes-as-a-Service platform with built-in AI capabilities.

## What is Hexabase AI?

Hexabase AI is a multi-tenant Kubernetes platform that simplifies application deployment and management while providing powerful AI-driven automation features. Built on K3s and vCluster technology, it offers isolated Kubernetes environments with enterprise-grade security and scalability.

## Key Features

### 🚀 Instant Kubernetes Environments
- **Workspaces**: Isolated vCluster environments provisioned in seconds
- **Projects**: Namespace-based resource organization
- **Auto-scaling**: Intelligent resource management

### 🤖 AI-Powered Operations
- **Smart Troubleshooting**: AI agents analyze and fix issues
- **Code Generation**: Generate Kubernetes manifests and configurations
- **Performance Optimization**: AI-driven resource recommendations

### 🔧 Developer-Friendly
- **Simple CLI**: Intuitive command-line interface
- **Web Dashboard**: Modern UI for visual management
- **API-First**: Complete REST and WebSocket APIs

### 🏗️ Enterprise Ready
- **Multi-tenancy**: Complete isolation between workspaces
- **RBAC**: Fine-grained access control
- **Compliance**: SOC2, HIPAA, GDPR ready
- **High Availability**: Built-in redundancy and failover

### 💼 Built-in Services
- **CronJobs**: Scheduled task management
- **Serverless Functions**: Event-driven compute with Knative
- **Backup & Restore**: Automated data protection
- **Monitoring**: Integrated Prometheus and Grafana

## Use Cases

### Development Teams
- Spin up isolated development environments
- Test applications in production-like settings
- Collaborate with built-in access controls

### DevOps Engineers
- Automate deployment pipelines
- Manage multiple environments from one place
- Monitor and optimize resource usage

### Enterprises
- Provide self-service Kubernetes to teams
- Maintain compliance and security standards
- Reduce infrastructure costs

### AI/ML Engineers
- Deploy ML models as serverless functions
- Schedule training jobs with CronJobs
- Use AI agents for automated operations

## How It Works

1. **Create Organization**: Set up your billing and team unit
2. **Provision Workspace**: Get an isolated Kubernetes environment
3. **Deploy Applications**: Use kubectl, UI, or API
4. **Monitor & Scale**: Built-in observability and auto-scaling
5. **Collaborate**: Invite team members with role-based access

## Platform Architecture

```
┌─────────────────────┐
│   User Interface    │
│  (Web UI / CLI)     │
└──────────┬──────────┘
           │
┌──────────▼──────────┐
│   Hexabase API      │
│  (Control Plane)    │
└──────────┬──────────┘
           │
┌──────────▼──────────┐     ┌─────────────────┐
│   Host K3s Cluster  │────▶│  vCluster       │
│                     │     │  (Workspace)    │
└─────────────────────┘     └─────────────────┘
           │
┌──────────▼──────────┐
│   Infrastructure    │
│ (Storage, Network)  │
└─────────────────────┘
```

## Pricing Plans

### Starter
- 1 workspace
- 2 CPU, 4GB RAM
- Community support
- Perfect for individuals

### Pro
- 5 workspaces
- 8 CPU, 32GB RAM
- Email support
- Great for small teams

### Enterprise
- Unlimited workspaces
- Custom resources
- 24/7 support
- SLA guarantees

## Getting Started

Ready to begin? Follow our [Quick Start Guide](./quick-start.md) to deploy your first application in minutes!

## Compare to Alternatives

| Feature | Hexabase AI | Traditional K8s | Other KaaS |
|---------|-------------|-----------------|------------|
| Setup Time | < 1 minute | Hours/Days | 10-30 minutes |
| Multi-tenancy | Built-in | Complex setup | Limited |
| AI Operations | ✅ | ❌ | ❌ |
| Cost | Pay-per-use | High fixed | Variable |
| Learning Curve | Low | High | Medium |

## Next Steps

1. [Understand Core Concepts](./concepts.md)
2. [Follow Quick Start Guide](./quick-start.md)
3. [Explore Features](../../architecture/system-architecture.md)
4. [Join Community](https://discord.gg/hexabase)

## Questions?

- **Sales**: sales@hexabase.ai
- **Support**: support@hexabase.ai
- **Documentation**: [docs.hexabase.ai](https://docs.hexabase.ai)
- **Status**: [status.hexabase.ai](https://status.hexabase.ai)

---

*Hexabase AI - Kubernetes Made Simple, Powered by AI* 🚀