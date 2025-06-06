# Development Guide

This section contains everything developers need to contribute to Hexabase KaaS.

## In This Section

### [Development Environment Setup](./dev-environment-setup.md)
Complete guide to setting up your local development environment with all required tools and dependencies.

### [API Development Guide](./api-development-guide.md)
Learn how to develop and extend the Go-based API server, including:
- Project structure and architecture
- Adding new endpoints
- Working with the domain layer
- Database migrations
- Testing strategies

### [Frontend Development Guide](./frontend-development-guide.md)
Guide for working with the Next.js frontend application:
- Component architecture
- State management
- API integration
- UI/UX guidelines

### [Code Style Guide](./code-style-guide.md)
Coding standards and best practices for:
- Go code conventions
- TypeScript/React patterns
- Git commit messages
- Code review guidelines

### [Git Workflow](./git-workflow.md)
Our branching strategy and development workflow:
- Branch naming conventions
- Pull request process
- Code review requirements
- Release management

## Quick Start for Developers

1. **Set up your environment**: Follow the [Development Environment Setup](./dev-environment-setup.md)
2. **Understand the architecture**: Read the [System Architecture](../architecture/system-architecture.md)
3. **Choose your focus**:
   - Backend: Start with [API Development Guide](./api-development-guide.md)
   - Frontend: Begin with [Frontend Development Guide](./frontend-development-guide.md)
4. **Follow standards**: Review the [Code Style Guide](./code-style-guide.md)

## Development Tools

- **Go 1.24+** - Backend API development
- **Node.js 18+** - Frontend development
- **Docker** - Containerization
- **Kubernetes** - Local K3s/Kind cluster
- **PostgreSQL** - Database
- **Redis** - Caching layer

## Getting Help

- Check existing issues on GitHub
- Join our development discussions
- Review the [Troubleshooting Guide](../operations/troubleshooting.md)