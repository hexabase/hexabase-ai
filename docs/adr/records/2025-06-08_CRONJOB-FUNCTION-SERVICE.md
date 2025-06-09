# CronJob and Function Service Implementation Summary

## Overview
This document summarizes the implementation of CronJob and Function Service features in Hexabase AI platform, enabling users to build small AI agents and applications with scheduled tasks and serverless functions.

## CronJob Implementation (Completed)

### Database Layer
1. **Schema Updates** (`005_add_cronjob_support.up.sql`):
   - Added `type` column to applications table with values: 'stateless', 'cronjob', 'function'
   - Added CronJob-specific columns: `cron_schedule`, `cron_command`, `cron_args`, `template_app_id`
   - Added execution tracking columns: `last_execution_at`, `next_execution_at`
   - Created `cronjob_executions` table for execution history

2. **Domain Models** (`internal/domain/application/models.go`):
   - Extended `ApplicationType` enum with `CronJob` and `Function` types
   - Added CronJob fields to `Application` struct
   - Created `CronJobExecution` model for tracking job runs
   - Added `CronJobStatus` for Kubernetes status reporting

3. **Repository Layer**:
   - **PostgreSQL** (`internal/repository/application/postgres.go`):
     - Implemented CronJob CRUD operations
     - Added execution history management
     - Support for template-based CronJobs
   - **Kubernetes** (`internal/repository/application/kubernetes.go`):
     - Implemented Kubernetes CronJob resource management
     - Added manual trigger functionality
     - Status monitoring integration

4. **Service Layer** (`internal/service/application/service.go`):
   - `CreateCronJob`: Creates CronJob with validation
   - `UpdateCronJobSchedule`: Updates cron schedule
   - `TriggerCronJob`: Manual job triggering
   - `GetCronJobExecutions`: Paginated execution history
   - `GetCronJobStatus`: Real-time Kubernetes status

5. **API Layer** (`internal/api/handlers/applications.go`):
   - REST endpoints for CronJob management
   - Schedule update endpoint: `PUT /applications/:appId/schedule`
   - Manual trigger endpoint: `POST /applications/:appId/trigger`
   - Execution history endpoint: `GET /applications/:appId/executions`
   - Status endpoint: `GET /applications/:appId/cronjob-status`

### Key Features
- **Template-based CronJobs**: Reuse existing application configurations
- **Manual Triggering**: On-demand job execution
- **Execution History**: Track all job runs with logs and exit codes
- **Schedule Management**: Dynamic schedule updates without recreating jobs
- **Integration with Backup Settings**: CronJobs can be used for automated backups

## Function Service Implementation Plan

### Architecture
1. **Knative Installation**: Serverless runtime on K3s cluster
2. **hks-func CLI**: Developer tool for function management
3. **Function Registry**: Store and version function code
4. **Runtime Support**: Python, Node.js, Go

### Planned Features
1. **Function Types**:
   - HTTP-triggered functions
   - Event-driven functions
   - Scheduled functions (using CronJob integration)

2. **Development Workflow**:
   - Local development with hks-func CLI
   - Hot reload and debugging
   - Automatic deployment on push

3. **AI Agent Support**:
   - Python SDK for LLM integration
   - Built-in connectors for popular AI services
   - State management for conversational agents

4. **Internal Operations API**:
   - Secure endpoints for AI agents
   - Rate limiting and authentication
   - Audit logging for compliance

### Implementation Phases

#### Phase 1: Knative Setup (Next)
- Install Knative Serving on host K3s cluster
- Configure autoscaling and cold start optimization
- Set up internal registry for function images

#### Phase 2: CLI Development
- Create hks-func CLI in Go
- Commands: init, build, deploy, logs, delete
- Local development server

#### Phase 3: Function Management API
- REST API for function lifecycle
- Version management
- Environment variable configuration
- Secret management integration

#### Phase 4: Python SDK
- SDK for AI agent development
- LLM provider abstractions
- State management utilities
- Example templates

#### Phase 5: Internal Operations API
- Secure API for agent operations
- WebSocket support for real-time features
- Metrics and monitoring integration

## Use Cases

### CronJob Use Cases
1. **Scheduled Backups**: Regular database or file backups
2. **Report Generation**: Periodic business intelligence reports
3. **Data Synchronization**: ETL jobs between systems
4. **Cleanup Tasks**: Remove old logs, temporary files
5. **Health Checks**: Regular system health monitoring

### Function Service Use Cases
1. **AI Chatbots**: Conversational interfaces for applications
2. **Webhook Handlers**: Process GitHub, Stripe, Slack events
3. **Data Processing**: Transform and analyze data on demand
4. **API Gateways**: Custom API endpoints with business logic
5. **Automation Scripts**: Replace traditional cron with event-driven functions

## Testing Strategy

### Unit Tests
- Repository layer tests with mocks
- Service layer business logic tests
- API handler tests with gin test context

### Integration Tests
- CronJob creation and execution flow
- Kubernetes resource lifecycle
- Database transaction handling

### E2E Tests
- Complete CronJob workflow from UI to execution
- Function deployment and invocation
- Performance benchmarks

## Security Considerations

1. **Resource Limits**: Enforce CPU/memory limits on CronJobs and Functions
2. **Network Policies**: Isolate function execution environments
3. **Secret Management**: Integrate with Kubernetes secrets
4. **Audit Logging**: Track all operations for compliance
5. **RBAC**: Fine-grained permissions for CronJob/Function management

## Monitoring and Observability

1. **Metrics**:
   - Job execution duration
   - Success/failure rates
   - Resource consumption
   - Function invocation counts

2. **Logging**:
   - Centralized logging with ClickHouse
   - Structured logs with correlation IDs
   - Real-time log streaming

3. **Alerting**:
   - Failed job notifications
   - Resource threshold alerts
   - SLA monitoring

## Next Steps

1. Complete Knative installation on K3s cluster
2. Begin hks-func CLI development
3. Design Function registry schema
4. Create Python SDK prototype
5. Document API specifications

## References
- [Knative Documentation](https://knative.dev/docs/)
- [Kubernetes CronJob API](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
- [OpenFaaS Architecture](https://docs.openfaas.com/) (for comparison)