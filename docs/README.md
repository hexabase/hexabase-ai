# Hexabase AI Documentation

Welcome to the Hexabase AI documentation. This guide will help you navigate our documentation structure and find the information you need.

## 📚 Documentation Structure

### [Architecture](./architecture/)
System design documents and architectural decisions
- [System Architecture](./architecture/system-architecture.md) - High-level system design
- [Technical Design](./architecture/technical-design.md) - Detailed technical architecture
- [Security Architecture](./architecture/security-architecture.md) - Security model and practices
- [AI-Ops Architecture](./architecture/ai-ops-architecture.md) - AI integration design

### [Specifications](./specifications/)
Detailed feature specifications and requirements
- [Storage and Backup Service](./specifications/storage-and-backup-service.md)
- [CronJob and Function Service](./specifications/cronjob-and-function-service.md)
- [Logging and Auditing](./specifications/logging-and-auditing.md)
- [Package Management](./specifications/package-management.md)

### [Guides](./guides/)
How-to guides and tutorials

#### [Development](./guides/development/)
For platform developers
- [Development Environment Setup](./guides/development/dev-environment-setup.md)
- [API Development Guide](./guides/development/api-development-guide.md)
- [Frontend Development Guide](./guides/development/frontend-development-guide.md)
- [Testing Guide](./guides/development/testing-guide.md)

#### [Deployment](./guides/deployment/)
For infrastructure and operations teams
- [Kubernetes Deployment](./guides/deployment/kubernetes-deployment.md)
- [Production Setup](./guides/deployment/production-setup.md)
- [Monitoring Setup](./guides/deployment/monitoring-setup.md)
- [Backup & Recovery](./guides/deployment/backup-recovery.md)

#### [Getting Started](./guides/getting-started/)
For new users
- [Overview](./guides/getting-started/README.md)
- [Core Concepts](./guides/getting-started/concepts.md)
- [Quick Start](./guides/getting-started/quick-start.md)

### [Architecture Decision Records (ADRs)](./adr/)
Track architectural decisions and their rationale
- [What is an ADR?](./adr/README.md)
- [Current Work Status](./adr/WORK-STATUS.md) - Live implementation progress
- [Decision Records](./adr/records/) - Historical decisions

### [API Reference](./api-reference/)
Complete API documentation
- [REST API](./api-reference/rest-api.md)
- [WebSocket API](./api-reference/websocket-api.md)
- [Authentication](./api-reference/authentication.md)
- [Error Codes](./api-reference/error-codes.md)

### [Roadmap](./roadmap/)
Project planning and future development
- [2025 Roadmap](./roadmap/ROADMAP_2025.md) - Current year planning
- [Project Overview](./roadmap/README.md) - Sprint status and metrics
- [Work Logs](./roadmap/work-logs/) - Historical implementation details

### [Testing](./testing/)
Testing strategies and coverage
- [Testing Guide](./testing/testing-guide.md)
- [OAuth Testing](./testing/oauth-testing.md)
- [Coverage Reports](./testing/coverage-reports/)

## 🚀 Quick Start by Role

### For Developers
1. [Set up development environment](./guides/development/dev-environment-setup.md)
2. [Understand the architecture](./architecture/system-architecture.md)
3. [Review API documentation](./api-reference/rest-api.md)
4. [Check current work status](./adr/WORK-STATUS.md)

### For DevOps/Operations
1. [Review deployment options](./guides/deployment/kubernetes-deployment.md)
2. [Set up monitoring](./guides/deployment/monitoring-setup.md)
3. [Configure backups](./guides/deployment/backup-recovery.md)
4. [Security considerations](./architecture/security-architecture.md)

### For Users
1. [Understand core concepts](./guides/getting-started/concepts.md)
2. [Follow quick start guide](./guides/getting-started/quick-start.md)
3. [Explore API capabilities](./api-reference/README.md)

## 📊 Project Status

- **Current Phase**: Phase 1 - Core Platform & AI Agent Foundation
- **Sprint**: Jun 9-23, 2025
- **Focus**: CronJob management, serverless functions, documentation

For detailed status, see [Work Status](./adr/WORK-STATUS.md)

## 🔍 Finding Information

### Search Documentation
```bash
# Search for a specific term
grep -r "search term" docs/

# Find recently updated docs
find docs -name "*.md" -mtime -7

# List all specification documents
ls -la docs/specifications/
```

### Navigation Tips
- Each directory has its own README with an overview
- Use relative links to navigate between documents
- Check ADRs for understanding why decisions were made
- Review work logs for implementation history

## 🤝 Contributing to Documentation

1. Follow the existing structure
2. Write clear, concise content
3. Include code examples where helpful
4. Update index files when adding new docs
5. Check for broken links

See [Contributing Guidelines](../CONTRIBUTING.md) for more details.

## 📞 Getting Help

- **GitHub Issues**: [Report bugs or request features](https://github.com/hexabase/hexabase-ai/issues)
- **Discussions**: [Ask questions and share ideas](https://github.com/hexabase/hexabase-ai/discussions)
- **Documentation Issues**: [Help improve these docs](https://github.com/hexabase/hexabase-ai/issues/new?labels=documentation)

## 🔧 Additional Resources

- [Claude AI Integration](../CLAUDE.md) - Working with Claude Code assistant
- [API Implementation](./roadmap/work-logs/) - Detailed implementation notes
- [Change Log](../CHANGELOG.md) - Version history

---

*Documentation last updated: June 9, 2025*