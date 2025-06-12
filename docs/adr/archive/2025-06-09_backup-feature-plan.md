# Backup Feature Implementation Plan

## Overview
The Backup feature will be implemented under the Application section, allowing workspace administrators to create and manage backup storage for their applications. This feature is only available for workspaces on the "Dedicated Plan" and integrates with Proxmox API for storage management.

## Architecture

### Key Components
1. **Backup Storage Management**: Create and manage backup storage volumes using Proxmox API
2. **Quota Management**: Set storage quotas at the workspace level
3. **Backup Policies**: Define backup schedules and retention policies
4. **Backup Execution**: Integrate with CronJob for scheduled backups
5. **Restore Operations**: Ability to restore from backups

### Access Control
- Only available for workspaces on "Dedicated Plan"
- Workspace admin role required to create/manage backup storage
- Application owners can configure backup policies for their apps

## Database Schema

### New Tables
```sql
-- Backup storage configuration
CREATE TABLE backup_storages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('nfs', 'ceph', 'local')),
    proxmox_storage_id VARCHAR(255) NOT NULL,
    capacity_gb INTEGER NOT NULL,
    used_gb INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending',
    connection_config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workspace_id, name)
);

-- Backup policies
CREATE TABLE backup_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    storage_id UUID NOT NULL REFERENCES backup_storages(id),
    enabled BOOLEAN DEFAULT true,
    schedule VARCHAR(100) NOT NULL, -- Cron expression
    retention_days INTEGER DEFAULT 30,
    backup_type VARCHAR(50) DEFAULT 'full' CHECK (backup_type IN ('full', 'incremental')),
    include_volumes BOOLEAN DEFAULT true,
    include_database BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(application_id)
);

-- Backup executions
CREATE TABLE backup_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES backup_policies(id) ON DELETE CASCADE,
    cronjob_execution_id UUID REFERENCES cronjob_executions(id),
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    size_bytes BIGINT,
    backup_path TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Backup restore operations
CREATE TABLE backup_restores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    backup_execution_id UUID NOT NULL REFERENCES backup_executions(id),
    application_id UUID NOT NULL REFERENCES applications(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    restore_type VARCHAR(50) NOT NULL CHECK (restore_type IN ('full', 'selective')),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Implementation Steps

### Phase 1: Backup Storage Management
1. Create domain models for BackupStorage
2. Implement Proxmox repository for storage operations
3. Add storage quota management
4. Create service layer for backup storage
5. Implement API endpoints for storage CRUD
6. Add workspace plan validation

### Phase 2: Backup Policy Configuration
1. Create domain models for BackupPolicy
2. Implement repository layer for policies
3. Create service layer for policy management
4. Add API endpoints for policy CRUD
5. Integrate with CronJob for scheduled execution

### Phase 3: Backup Execution
1. Create backup execution job template
2. Implement backup execution service
3. Add volume snapshot integration
4. Implement database backup logic
5. Store backups in configured storage
6. Track execution status and metrics

### Phase 4: Restore Operations
1. Create restore service
2. Implement restore API endpoints
3. Add selective restore capabilities
4. Integrate with application lifecycle

### Phase 5: UI Components
1. Backup storage management UI
2. Backup policy configuration UI
3. Backup history and monitoring
4. Restore operation UI
5. Storage usage visualization

## API Endpoints

### Backup Storage
- `GET /api/v1/workspaces/{workspaceId}/backup-storages` - List backup storages
- `POST /api/v1/workspaces/{workspaceId}/backup-storages` - Create backup storage
- `GET /api/v1/workspaces/{workspaceId}/backup-storages/{storageId}` - Get storage details
- `PUT /api/v1/workspaces/{workspaceId}/backup-storages/{storageId}` - Update storage
- `DELETE /api/v1/workspaces/{workspaceId}/backup-storages/{storageId}` - Delete storage

### Backup Policies
- `GET /api/v1/applications/{appId}/backup-policy` - Get backup policy
- `POST /api/v1/applications/{appId}/backup-policy` - Create/update policy
- `DELETE /api/v1/applications/{appId}/backup-policy` - Delete policy

### Backup Operations
- `GET /api/v1/applications/{appId}/backups` - List backups
- `POST /api/v1/applications/{appId}/backups` - Trigger manual backup
- `GET /api/v1/applications/{appId}/backups/{backupId}` - Get backup details
- `POST /api/v1/applications/{appId}/backups/{backupId}/restore` - Restore from backup

## Integration Points

### With CronJob Feature
- Use CronJob infrastructure for scheduled backups
- Create specialized CronJob templates for backup operations
- Track execution through CronJob execution records

### With Proxmox API
- Create storage volumes
- Manage storage quotas
- Monitor storage usage
- Handle storage lifecycle

### With Kubernetes
- Volume snapshots for persistent volumes
- Job execution for backup operations
- ConfigMap/Secret backup

## Security Considerations
- Encrypt backups at rest
- Secure storage credentials in Kubernetes secrets
- Audit trail for backup/restore operations
- Role-based access control
- Network isolation for backup storage

## Monitoring and Metrics
- Backup success/failure rates
- Storage usage trends
- Backup duration metrics
- Restore operation success rates
- Alert on backup failures

## Testing Strategy
- Unit tests for backup service logic
- Integration tests with mock Proxmox API
- End-to-end tests for backup/restore cycle
- Performance tests for large backups
- Disaster recovery testing