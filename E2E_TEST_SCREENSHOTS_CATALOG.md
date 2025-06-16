# Hexabase AI E2E Test Screenshots Catalog

## Overview
This catalog contains all available E2E test screenshots from the Hexabase AI platform, captured during automated testing. The screenshots demonstrate the full user journey through various features and workflows.

## Test Execution Summary
- **Backend API**: Successfully started (fixed migration order issue)
- **Available Screenshots**: 30+ screens from previous successful test runs
- **Test Dates**: June 13-14, 2025

## Screenshot Gallery

### 1. Authentication & Login
- **Login Page**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/auth/01_login_page.png`
  - Shows the initial login screen with OAuth provider options

### 2. Dashboard & Overview
- **Main Dashboard**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/dashboard/01_dashboard_overview.png`
  - Displays the main dashboard after successful login
  - Shows workspace statistics, recent activities, and quick actions

### 3. Organization Management
- **Organization List**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/organization/01_organization_list.png`
- **Create Organization**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/organization/02_create_organization.png`
- **Organization Settings**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/organization/03_organization_settings.png`
- **Team Members**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/organization/04_team_members.png`
- **Billing Overview**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/organization/05_billing_overview.png`

### 4. Workspace Management
- **Workspace Dashboard**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/workspace/01_workspace_dashboard.png`
  - Shows workspace resources, vCluster status, and project hierarchy

### 5. Project Management
- **Project Dashboard**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/projects/01_project_dashboard.png`
  - Displays project namespaces, resource quotas, and deployments

### 6. Application Deployment
- **Application List**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/applications/01_application_list.png`
- **Deploy Application**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/applications/02_deploy_application.png`
- **Application Details**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/applications/01_application_details.png`
- **Application Logs**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/applications/04_application_logs.png`

### 7. CI/CD Pipeline
- **Pipeline Overview**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/cicd/01_pipeline_overview.png`
  - Shows GitHub Actions integration and build pipelines

### 8. Serverless Functions
- **Functions Overview**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/serverless/01_functions_overview.png`
  - Displays Knative-based serverless functions with versioning

### 9. Monitoring & Observability
- **Monitoring Dashboard**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/monitoring/01_monitoring_dashboard.png`
  - Shows Prometheus metrics, Grafana charts, and resource usage

### 10. Backup & Restore (Dedicated Plan)
- **Backup Overview**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/backup/01_backup_overview.png`
  - Displays backup configurations and Proxmox storage integration

### 11. AI Assistant Integration
- **AI Assistant Button**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/ai_chat/01_ai_assistant_button.png`
- **Chat Interface**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/ai_chat/02_chat_interface.png`
- **Code Generation**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/ai_chat/03_code_generation.png`
- **Troubleshooting Help**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/ai_chat/04_troubleshooting.png`
- **Deployment Assistance**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/ai_chat/05_deployment_help.png`

### 12. Advanced Deployments
- **Blue-Green Deployment**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/deployments/01_blue_green_deployment.png`
  - Shows zero-downtime deployment strategies

### 13. OAuth/SSO Integration
- **OAuth Providers**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/oauth/01_oauth_providers.png`
- **Google Consent**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/oauth/02_google_consent.png`
- **GitHub Authorize**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/oauth/03_github_authorize.png`
- **SSO Success**: `/ui/screenshots/e2e_result_2025-06-13T16-42-44/oauth/04_sso_success.png`

### 14. Success States
- **Complete Success**: `/ui/screenshots/e2e_result_2025-06-13T15-02-37/success/01_complete_success.png`
  - Shows successful deployment completion state

## Key Features Demonstrated

### Multi-Tenant Architecture
- vCluster-based workspace isolation
- Hierarchical namespace management with HNC
- Resource quotas and limits per workspace

### Application Types Supported
1. **Stateless Applications**: Standard Kubernetes Deployments
2. **Stateful Applications**: StatefulSets with persistent storage
3. **CronJobs**: Scheduled tasks with execution tracking
4. **Serverless Functions**: Knative-based functions with versioning
5. **AI Agents**: Python/Node.js functions with AI model access

### Platform Capabilities
- **OIDC Authentication**: External IdP integration
- **GitOps Workflows**: Tekton-based CI/CD pipelines
- **Observability**: Prometheus, Grafana, Loki integration
- **Security**: Kyverno policies, Trivy scanning, Falco monitoring
- **Backup System**: Proxmox-integrated backup for Dedicated Plans
- **AI Operations**: Integrated AI assistant with LangFuse tracking

## Viewing Screenshots

To view the screenshots locally:

```bash
# Navigate to UI directory
cd /Users/hi/src/hexabase-ai/ui

# Open screenshots directory
open screenshots/

# View specific test run
open screenshots/e2e_result_2025-06-13T15-02-37/
```

## Running New Tests

To capture new screenshots:

```bash
# Start backend API
cd api
./run-dev.sh

# In another terminal, start UI dev server
cd ui
npm run dev

# Run E2E tests with screenshots
npm run test:e2e
```

## Test Infrastructure Details

- **Framework**: Playwright with TypeScript
- **Mock API**: MSW (Mock Service Worker) for API mocking
- **Screenshot Config**: Automatic capture on test failure and specific checkpoints
- **Browser**: Chromium (headless mode for CI)

## Notes

- All screenshots use mock data to ensure consistent testing
- The UI implements Hexabase design system with consistent theming
- Screenshots demonstrate both Shared and Dedicated Plan features
- AI Assistant integration shows real-time code generation and troubleshooting