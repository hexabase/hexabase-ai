# Hexabase.AI - KaaS Platform with AIOps -

An open-source, multi-tenant Kubernetes as a Service platform built on K3s and vCluster.

## 🚀 Overview

Hexabase AI provides a user-friendly abstraction layer over Kubernetes, enabling developers to deploy and manage applications without dealing with Kubernetes complexity directly. It offers isolated Kubernetes environments with enterprise-grade security, monitoring, and resource quota management capabilities.

## 📚 Documentation

### Quick Links

- **[Getting Started Guide](./docs/getting-started/README.md)** - Platform overview and quick start
- **[Development Setup](./docs/development/dev-environment-setup.md)** - Set up your local environment
- **[Project Structure Guide](./STRUCTURE_GUIDE.md)** - Code organization and conventions
- **[Deployment Guide](./docs/operations/kubernetes-deployment.md)** - Deploy to production
- **[API Reference](./docs/api-reference/README.md)** - Complete API documentation
- **[CI/CD Architecture](./docs/architecture/cicd-architecture.md)** - CI/CD pipelines and GitOps
- **[CI/CD Configurations](./ci/README.md)** - Pipeline configurations for different platforms

### Documentation Structure

```
docs/
├── getting-started/        # Introduction and concepts
├── architecture/          # System design and architecture
├── development/           # Developer guides
├── operations/           # Deployment and operations
├── api-reference/        # API documentation
├── testing/              # Testing guides ONLY (no results)
├── implementation-summaries/ # Implementation notes
└── project-management/   # Project status and roadmap

api/
├── testresults/          # ALL test results and reports
│   ├── coverage/        # Coverage data
│   ├── unit/            # Unit test results
│   └── coverage-reports/ # Test coverage reports

ci/                       # CI/CD configurations
├── github-actions/       # GitHub Actions workflows
├── gitlab-ci/           # GitLab CI pipelines
└── tekton/              # Tekton pipeline definitions

deployments/
├── gitops/              # GitOps configurations
│   ├── flux/           # Flux CD configurations
│   └── argocd/         # ArgoCD applications
├── policies/            # Security policies
│   └── kyverno/        # Kyverno policies
├── monitoring/          # Monitoring configurations
│   └── prometheus/     # Prometheus rules
└── canary/             # Progressive delivery
    └── flagger/        # Flagger configurations
```

## 🎯 Key Features

### Multi-Tenant Kubernetes

- **vCluster Isolation**: Each workspace gets its own virtual Kubernetes cluster
- **Resource Quotas**: Fine-grained resource limits and quotas
- **Network Policies**: Secure tenant isolation at the network level

### Enterprise Security

- **OAuth2/OIDC**: Support for Google, GitHub, Azure AD, and custom providers
- **PKCE Flow**: Enhanced security for public clients
- **JWT Fingerprinting**: Token binding to prevent replay attacks
- **Audit Logging**: Comprehensive security event tracking

### Developer Experience

- **Simple Abstractions**: Organizations → Workspaces → Projects
- **Self-Service**: Create and manage Kubernetes environments via UI/API
- **Real-time Updates**: WebSocket notifications for provisioning status

### Operations & Monitoring

- **Prometheus Metrics**: Built-in metrics collection
- **Grafana Dashboards**: Pre-configured visualization
- **Health Checks**: Automated cluster health monitoring
- **Alert Management**: Configurable alerting rules

### Billing & Subscriptions

- **Stripe Integration**: Automated billing and invoicing
- **Usage Tracking**: Resource consumption monitoring
- **Flexible Plans**: Multiple subscription tiers

## 🏗️ Architecture

### Core Components

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Next.js UI    │────▶│    Go API       │────▶│   PostgreSQL    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ├──▶ Redis (Cache & Sessions)
                               ├──▶ NATS (Message Queue)
                               ├──▶ ClickHouse (Logs)
                               └──▶ Kubernetes API
                                         │
                                   ┌─────┴─────┐
                                   │ vClusters │
                                   └───────────┘

┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Python AIOps   │────▶│     Ollama      │────▶│    Langfuse     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

### Technology Stack

- **Backend**: Go 1.24+, Gin, GORM, Wire
- **Frontend**: Next.js 14+, TypeScript, Tailwind CSS
- **Infrastructure**: Kubernetes, vCluster, Helm
- **Database**: PostgreSQL 14+
- **Cache**: Redis 6+
- **Message Queue**: NATS JetStream
- **Monitoring**: Prometheus, Grafana, Loki
- **Logging Stack**: ClickHouse (structured logs), Logrus/Zap (application logging)
- **LLMOps Stack**: Ollama (LLM serving), Langfuse (LLM observability), OpenAI API

## 🚦 Getting Started

### Prerequisites

- Go 1.24+
- Node.js 18+
- Docker & Docker Compose
- Kubernetes cluster (or K3s)
- PostgreSQL 14+
- Redis 6+

### Quick Start

#### Option 1: Automated Setup (Recommended)

```bash
# Clone and setup everything with one command
git clone https://github.com/hexabase/hexabase-kaas.git
cd hexabase-kaas
make setup

# Start development
make dev  # Runs both API and UI in tmux
# OR run separately:
make dev-api  # Terminal 1
make dev-ui   # Terminal 2
```

#### Option 2: Manual Setup

```bash
# Clone the repository
git clone https://github.com/hexabase/hexabase-kaas.git
cd hexabase-kaas

# Start infrastructure
docker-compose up -d

# Run the API
cd api && go run cmd/api/main.go

# Run the UI (new terminal)
cd ui && npm install && npm run dev
```

Access the application at:

- **UI**: http://app.localhost
- **API**: http://api.localhost

For detailed setup instructions, see the [Development Environment Setup](./docs/development/dev-environment-setup.md).

## 🧪 Testing

```bash
# API tests
cd api
go test ./...

# UI tests
cd ui
npm test
npm run test:e2e
```

See the [Testing Guide](./docs/testing/testing-guide.md) for comprehensive testing strategies.

## 🚀 Deployment

### Quick Deployment with Helm (Recommended)

```bash
# Add Hexabase Helm repository
helm repo add hexabase https://charts.hexabase.ai
helm repo update

# Install Hexabase KaaS with production values
helm install hexabase-kaas hexabase/hexabase-kaas \
  --namespace hexabase-system \
  --create-namespace \
  --values deployments/helm/values-production.yaml
```

For detailed deployment options, see the [Kubernetes Deployment Guide](./docs/operations/kubernetes-deployment.md).

## 🔄 CI/CD & GitOps

Hexabase KaaS supports multiple CI/CD platforms and GitOps workflows:

### CI/CD Platforms

- **[GitHub Actions](./ci/github-actions/)** - Native GitHub integration
- **[GitLab CI](./ci/gitlab-ci/)** - GitLab pipeline support
- **[Tekton](./ci/tekton/)** - Cloud-native Kubernetes pipelines

### GitOps Tools

- **[Flux](./deployments/gitops/flux/)** - Automated Git-to-Kubernetes sync
- **[ArgoCD](./deployments/gitops/argocd/)** - Declarative GitOps with UI

### Security & Policies

- **[Kyverno Policies](./deployments/policies/kyverno/)** - Policy enforcement
- **[Supply Chain Security](./ci/github-actions/supply-chain.yml)** - SBOM and signing

For architecture details, see the [CI/CD Architecture Guide](./docs/architecture/cicd-architecture.md).

## 🤝 Contributing

We welcome contributions! Please read our:
- [Contributing Guidelines](./CONTRIBUTING.md) - How to contribute
- [Code of Conduct](./CODE_OF_CONDUCT.md) - Community standards
- [Project Structure Guide](./STRUCTURE_GUIDE.md) - Code organization rules
- [AI Assistant Guide](./CLAUDE.md) - For AI-assisted development

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📋 Project Status

See [Project Management](./docs/project-management/README.md) for current development status and roadmap.

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](./LICENSE) file for details.

## 🙏 Acknowledgments

- [vCluster](https://www.vcluster.com/) for virtual Kubernetes clusters
- [K3s](https://k3s.io/) for lightweight Kubernetes
- All our contributors and supporters

## 📞 Support

- **Documentation**: [Full Documentation](./docs/README.md)
- **Issues**: [GitHub Issues](https://github.com/hexabase/hexabase-kaas/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hexabase/hexabase-kaas/discussions)

---

Built with ❤️ by the Hexabase team
