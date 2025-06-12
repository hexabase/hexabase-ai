# E2E Test Use Cases

This document outlines comprehensive end-to-end test scenarios for the Hexabase AI platform UI.

## Test Case Categories

1. [Authentication & Authorization](#authentication--authorization)
2. [Organization Management](#organization-management)
3. [Workspace Lifecycle](#workspace-lifecycle)
4. [Project Management](#project-management)
5. [Application Deployment](#application-deployment)
6. [CronJob Management](#cronjob-management)
7. [Function Management](#function-management)
8. [Monitoring & Observability](#monitoring--observability)
9. [Billing & Usage](#billing--usage)
10. [Real-time Updates](#real-time-updates)

## Authentication & Authorization

### TC-AUTH-001: OAuth Login Flow
**Priority**: Critical
**Prerequisites**: None

**Steps**:
1. Navigate to login page
2. Click "Login with GitHub/Google"
3. Complete OAuth provider authentication
4. Verify redirect to dashboard
5. Verify user profile is loaded

**Expected Results**:
- User is authenticated
- JWT token stored in secure storage
- User organizations are loaded
- Default organization is selected

### TC-AUTH-002: Session Persistence
**Priority**: High
**Prerequisites**: User logged in

**Steps**:
1. Login to application
2. Close browser tab
3. Open new tab and navigate to app
4. Verify automatic authentication

**Expected Results**:
- User remains logged in
- No re-authentication required
- Previous organization context maintained

### TC-AUTH-003: Multi-Organization Switching
**Priority**: High
**Prerequisites**: User with multiple organizations

**Steps**:
1. Login as user with 3+ organizations
2. Click organization switcher
3. Select different organization
4. Verify context switch
5. Navigate between organizations rapidly

**Expected Results**:
- Organization context switches correctly
- No data leakage between organizations
- UI updates reflect correct organization

## Organization Management

### TC-ORG-001: Create Organization Full Flow
**Priority**: Critical
**Prerequisites**: Authenticated user

**Steps**:
1. Click "Create Organization"
2. Enter organization name "Test Production Org"
3. Add description and billing email
4. Submit form
5. Verify organization created
6. Navigate to new organization

**Expected Results**:
- Organization created with unique ID
- User is organization admin
- Billing setup initiated
- Can access organization dashboard

### TC-ORG-002: Organization Settings Management
**Priority**: Medium
**Prerequisites**: Organization admin

**Steps**:
1. Navigate to organization settings
2. Update organization name
3. Change billing email
4. Add team members
5. Set organization policies
6. Save all changes

**Expected Results**:
- All settings persisted
- Team members receive invitations
- Audit log shows changes
- Billing email updated in Stripe

### TC-ORG-003: Organization Deletion with Dependencies
**Priority**: Medium
**Prerequisites**: Organization with workspaces

**Steps**:
1. Create organization with 2 workspaces
2. Deploy applications in workspaces
3. Attempt to delete organization
4. Confirm deletion warnings
5. Force delete with cascade

**Expected Results**:
- Warning about dependent resources
- Cascade deletion removes all resources
- Billing subscription cancelled
- Audit trail maintained

## Workspace Lifecycle

### TC-WS-001: Dedicated Workspace Provisioning
**Priority**: Critical
**Prerequisites**: Organization with billing

**Steps**:
1. Click "Create Workspace"
2. Enter name "Production Environment"
3. Select "Dedicated" plan
4. Configure resource limits (8 CPU, 16GB RAM)
5. Select data center region
6. Enable backup option
7. Submit and monitor provisioning

**Expected Results**:
- Proxmox VM created
- K3s cluster deployed
- vCluster provisioned
- Status updates in real-time
- Workspace active within 5 minutes

### TC-WS-002: Shared Workspace Quick Setup
**Priority**: High
**Prerequisites**: Organization exists

**Steps**:
1. Create workspace "Development"
2. Select "Shared" plan
3. Use default resource quotas
4. Submit creation

**Expected Results**:
- vCluster created immediately
- Resource quotas applied
- Namespace isolation verified
- Can deploy applications immediately

### TC-WS-003: Workspace Scaling
**Priority**: Medium
**Prerequisites**: Active dedicated workspace

**Steps**:
1. Navigate to workspace settings
2. Increase CPU from 4 to 8 cores
3. Increase memory from 8GB to 16GB
4. Add additional storage (100GB)
5. Apply changes
6. Monitor scaling operation

**Expected Results**:
- Proxmox VM resized
- No downtime for applications
- New limits reflected in UI
- Billing updated automatically

## Project Management

### TC-PROJ-001: Multi-Project Setup
**Priority**: High
**Prerequisites**: Active workspace

**Steps**:
1. Create project "frontend-app"
2. Set resource quotas (2 CPU, 4GB RAM)
3. Create project "backend-api"
4. Set different quotas (4 CPU, 8GB RAM)
5. Create project "shared-services"
6. Verify namespace isolation

**Expected Results**:
- Each project has isolated namespace
- Resource quotas enforced
- Can deploy to each project
- Projects listed in hierarchy

### TC-PROJ-002: Project Resource Management
**Priority**: Medium
**Prerequisites**: Project with applications

**Steps**:
1. Deploy 3 applications to project
2. Monitor resource usage
3. Approach resource quota limit
4. Attempt to deploy 4th application
5. Increase project quota
6. Retry deployment

**Expected Results**:
- Resource usage accurately tracked
- Deployment blocked when quota exceeded
- Clear error message shown
- Quota increase allows deployment

## Application Deployment

### TC-APP-001: Stateless Application Deployment
**Priority**: Critical
**Prerequisites**: Project exists

**Steps**:
1. Click "Deploy Application"
2. Select "Stateless" type
3. Enter name "web-frontend"
4. Choose container image "nginx:latest"
5. Set replicas to 3
6. Configure environment variables
7. Set resource limits
8. Deploy application
9. Monitor deployment progress

**Expected Results**:
- Kubernetes Deployment created
- 3 pods running
- Service exposed
- Health checks passing
- Application accessible

### TC-APP-002: Stateful Application with Storage
**Priority**: High
**Prerequisites**: Project exists

**Steps**:
1. Deploy stateful application
2. Select "PostgreSQL" template
3. Configure persistent volume (10GB)
4. Set backup policy
5. Deploy and wait for ready
6. Verify data persistence
7. Delete and recreate pod
8. Verify data retained

**Expected Results**:
- StatefulSet created
- PVC provisioned
- Data survives pod restart
- Backup policy active

### TC-APP-003: Rolling Update
**Priority**: High
**Prerequisites**: Running application

**Steps**:
1. Select running application
2. Click "Update"
3. Change image tag to new version
4. Modify environment variable
5. Apply rolling update
6. Monitor pod rollout
7. Verify zero downtime

**Expected Results**:
- Pods updated one by one
- No service interruption
- Old pods terminated after health check
- Rollback available if needed

## CronJob Management

### TC-CRON-001: Scheduled Backup Job
**Priority**: High
**Prerequisites**: Dedicated workspace with backup storage

**Steps**:
1. Create CronJob "daily-backup"
2. Set schedule "0 2 * * *" (2 AM daily)
3. Select backup template
4. Configure backup targets
5. Set retention policy (30 days)
6. Enable job
7. Manually trigger test run
8. Verify execution

**Expected Results**:
- CronJob created in Kubernetes
- Manual trigger creates Job
- Backup completes successfully
- Next run scheduled correctly
- Execution history recorded

### TC-CRON-002: Complex Schedule Management
**Priority**: Medium
**Prerequisites**: Project exists

**Steps**:
1. Create multiple CronJobs:
   - Hourly metrics collection
   - Daily reports (business hours only)
   - Weekly maintenance (Sunday 3 AM)
   - Monthly billing calculation
2. Verify schedule conflicts
3. Monitor execution overlaps
4. Check resource utilization

**Expected Results**:
- All schedules parsed correctly
- Schedule preview shows next 5 runs
- No resource conflicts
- Execution history maintained

### TC-CRON-003: Failed Job Handling
**Priority**: Medium
**Prerequisites**: CronJob exists

**Steps**:
1. Create CronJob with failing command
2. Let it execute automatically
3. View execution history
4. Check failure logs
5. Fix command
6. Verify next execution succeeds

**Expected Results**:
- Failed status recorded
- Error logs available
- Alerts sent (if configured)
- Retry policy followed
- Success after fix

## Function Management

### TC-FUNC-001: Serverless Function Deployment
**Priority**: High
**Prerequisites**: Project with Knative

**Steps**:
1. Click "Create Function"
2. Name function "image-processor"
3. Select runtime "Node.js 18"
4. Write/paste function code
5. Set memory limit (256MB)
6. Set timeout (30s)
7. Configure environment variables
8. Deploy function
9. Test with sample payload

**Expected Results**:
- Knative Service created
- Function scales to zero
- Cold start under 2 seconds
- Execution successful
- Response returned

### TC-FUNC-002: Function Versioning
**Priority**: Medium
**Prerequisites**: Deployed function

**Steps**:
1. Select existing function
2. Modify function code
3. Deploy as new version
4. Test new version
5. Compare performance
6. Rollback to previous version
7. Verify instant rollback

**Expected Results**:
- New version deployed (v2)
- Both versions accessible
- Traffic routing updated
- Rollback immediate
- No data loss

### TC-FUNC-003: Event-Driven Function
**Priority**: Medium
**Prerequisites**: Function deployed

**Steps**:
1. Create event-triggered function
2. Configure event source (S3, Queue)
3. Deploy function
4. Generate test event
5. Verify function execution
6. Check event processing

**Expected Results**:
- Event subscription active
- Function triggered by event
- Processing completed
- Result stored/forwarded
- Metrics recorded

## Monitoring & Observability

### TC-MON-001: Real-time Metrics Dashboard
**Priority**: High
**Prerequisites**: Applications running

**Steps**:
1. Navigate to monitoring dashboard
2. Select time range (last 1 hour)
3. View CPU/Memory metrics
4. Drill down to pod level
5. Create custom dashboard
6. Set up alerts

**Expected Results**:
- Metrics update every 30 seconds
- Historical data available
- Pod-level granularity
- Custom dashboards saved
- Alerts configured

### TC-MON-002: Log Aggregation
**Priority**: High
**Prerequisites**: Applications with logging

**Steps**:
1. Navigate to logs viewer
2. Select application
3. Filter by severity (ERROR)
4. Search for specific pattern
5. Export logs for time range
6. Set up log alerts

**Expected Results**:
- Logs collected from all pods
- Real-time log streaming
- Search functionality works
- Export includes all fields
- Alerts trigger on patterns

### TC-MON-003: Distributed Tracing
**Priority**: Medium
**Prerequisites**: Microservices deployed

**Steps**:
1. Generate traffic across services
2. View traces in UI
3. Identify slow service
4. Drill into span details
5. View service dependencies
6. Export trace data

**Expected Results**:
- Complete trace visible
- Service map accurate
- Latency breakdown shown
- Error traces highlighted
- Dependencies mapped

## Billing & Usage

### TC-BILL-001: Usage Tracking
**Priority**: High
**Prerequisites**: Active workspaces

**Steps**:
1. Navigate to billing dashboard
2. View current month usage
3. Check resource breakdown:
   - Compute hours
   - Storage GB/month
   - Network transfer
   - Function invocations
4. Compare with quotas
5. View cost projection

**Expected Results**:
- Real-time usage data
- Accurate cost calculation
- Clear quota indicators
- Projection for month-end
- Breakdown by workspace

### TC-BILL-002: Payment Method Management
**Priority**: Critical
**Prerequisites**: Organization admin

**Steps**:
1. Navigate to payment methods
2. Add new credit card
3. Set as default
4. Remove old card
5. Verify with test charge
6. Check invoice generation

**Expected Results**:
- Card validated with Stripe
- Set as default for org
- Old card removed
- Invoice uses new card
- Receipt emailed

## Real-time Updates

### TC-RT-001: Multi-User Collaboration
**Priority**: High
**Prerequisites**: Multiple users in organization

**Steps**:
1. User A and User B open same workspace
2. User A creates new project
3. Verify User B sees project immediately
4. User B deploys application
5. Verify User A sees deployment status
6. Both users modify different resources
7. Verify no conflicts

**Expected Results**:
- Updates appear within 2 seconds
- No page refresh required
- Conflict-free updates
- Status synchronized
- All users see same state

### TC-RT-002: Deployment Status Updates
**Priority**: High
**Prerequisites**: Application deploying

**Steps**:
1. Start application deployment
2. Open status in multiple tabs
3. Watch status progression:
   - Creating → Pending → Running
4. Simulate pod failure
5. Watch recovery status
6. Verify final state

**Expected Results**:
- Status updates in all tabs
- Pod events shown
- Failure detected immediately
- Recovery status accurate
- Final state consistent

## Complex End-to-End Scenarios

### TC-E2E-001: Complete Production Setup
**Priority**: Critical
**Time Estimate**: 30 minutes

**Steps**:
1. Create organization "Production Corp"
2. Set up billing with credit card
3. Create dedicated workspace "Production" (16 CPU, 32GB RAM)
4. Enable backup storage
5. Create projects:
   - frontend (public-facing)
   - backend (internal APIs)
   - databases (stateful services)
6. Deploy applications:
   - 3x frontend replicas
   - 2x backend services
   - 1x PostgreSQL with replication
   - 1x Redis cache
7. Configure monitoring and alerts
8. Set up CronJobs:
   - Database backup (daily)
   - Log cleanup (weekly)
   - Metrics aggregation (hourly)
9. Deploy serverless functions:
   - Image processing
   - PDF generation
   - Email notifications
10. Configure DNS and TLS
11. Run load test
12. Verify all components working

**Expected Results**:
- Complete production environment
- All services healthy
- Monitoring active
- Backups running
- Functions scaling
- Load test passes

### TC-E2E-002: Disaster Recovery
**Priority**: Critical
**Prerequisites**: TC-E2E-001 completed

**Steps**:
1. Simulate node failure
2. Verify pod rescheduling
3. Simulate database corruption
4. Initiate backup restore
5. Verify data integrity
6. Test failover procedures
7. Document recovery time

**Expected Results**:
- Automatic pod recovery
- Backup restore successful
- Data integrity maintained
- RTO under 15 minutes
- No data loss

## Performance Test Cases

### TC-PERF-001: UI Responsiveness
**Priority**: High

**Steps**:
1. Load organization with 50+ workspaces
2. Measure page load time
3. Navigate between workspaces rapidly
4. Open multiple projects
5. Deploy 10 applications simultaneously

**Expected Results**:
- Page load under 2 seconds
- Navigation under 500ms
- No UI freezing
- All deployments tracked

### TC-PERF-002: Concurrent Users
**Priority**: High

**Steps**:
1. Simulate 50 concurrent users
2. Each user performs different actions
3. Monitor API response times
4. Check WebSocket performance
5. Verify data consistency

**Expected Results**:
- API responses under 200ms (p95)
- WebSocket messages delivered
- No race conditions
- Data remains consistent