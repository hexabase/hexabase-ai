# Development Guides

This directory contains guides for developers working on the Hexabase AI platform.

## 📚 Available Guides

### [Development Environment Setup](./dev-environment-setup.md)
Step-by-step guide to set up your local development environment, including:
- Prerequisites and required tools
- Quick setup with automated scripts
- Manual setup instructions
- IDE configuration
- Troubleshooting common issues

### [API Development Guide](./api-development-guide.md)
Best practices for developing the Go-based API, covering:
- Project structure and patterns
- Domain-driven design principles
- Error handling and validation
- Testing strategies
- Performance optimization

### [Frontend Development Guide](./frontend-development-guide.md)
Guide for Next.js frontend development, including:
- Component architecture
- State management with Zustand
- Data fetching with React Query
- UI components with shadcn/ui
- Testing with React Testing Library

### [Testing Guide](./testing-guide.md)
Comprehensive testing strategies:
- Unit testing patterns
- Integration testing
- End-to-end testing with Playwright
- Performance testing
- Test coverage requirements

### [Code Style Guide](./code-style-guide.md)
*Coming soon*

### [Git Workflow](./git-workflow.md)
*Coming soon*

## 🚀 Getting Started

If you're new to the project:

1. Start with [Development Environment Setup](./dev-environment-setup.md)
2. Review the guide specific to your area:
   - Backend: [API Development Guide](./api-development-guide.md)
   - Frontend: [Frontend Development Guide](./frontend-development-guide.md)
3. Understand our [Testing Guide](./testing-guide.md)
4. Check the main [Architecture Documentation](../../architecture/)

## 🛠️ Development Workflow

1. **Setup**: Follow the environment setup guide
2. **Branch**: Create a feature branch from `develop`
3. **Code**: Follow the relevant development guide
4. **Test**: Write and run tests as per testing guide
5. **Review**: Submit PR for code review
6. **Deploy**: Merge to `develop` triggers CI/CD

## 📝 Contributing

When contributing to these guides:

- Keep examples up-to-date with the codebase
- Test all code examples
- Include troubleshooting sections
- Link to relevant external documentation
- Update the README when adding new guides

## 🔗 Related Documentation

- [Architecture Overview](../../architecture/system-architecture.md)
- [API Reference](../../api-reference/README.md)
- [Deployment Guides](../deployment/)
- [Getting Started](../getting-started/)