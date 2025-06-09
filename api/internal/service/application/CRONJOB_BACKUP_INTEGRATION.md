# CronJob with Backup Settings Integration

## Summary

This document outlines the TDD implementation for integrating CronJob applications with backup policies in the Hexabase AI platform.

## Implementation Status

### Completed

1. **Test-Driven Development Approach**
   - Created comprehensive test cases for CronJob-backup integration
   - Tests cover schedule validation, policy creation, and execution tracking
   - Mock implementations for testing without external dependencies

2. **Domain Model Updates**
   - Added `Metadata` field to `Application` and `CreateApplicationRequest` models
   - Updated service interfaces to include backup-related methods
   - Extended repository interfaces with CronJob execution tracking

3. **Core Implementation**
   - `CreateApplicationWithBackupPolicy`: Creates CronJob with associated backup policy
   - `TriggerCronJob`: Enhanced to create backup executions when policy exists
   - `UpdateCronJobExecutionStatus`: Updates backup status based on CronJob results
   - `validateBackupScheduleCompatibility`: Ensures backup runs after CronJob

### Integration Points

1. **CronJob Creation with Backup**
   ```go
   // Create CronJob that performs database backup
   app, err := svc.CreateApplicationWithBackupPolicy(ctx, 
       &application.CreateApplicationRequest{
           Name:         "db-backup",
           Type:         application.ApplicationTypeCronJob,
           CronSchedule: "0 2 * * *", // 2 AM daily
       },
       &backup.CreateBackupPolicyRequest{
           StorageID:     "storage-123",
           Schedule:      "0 4 * * *", // 4 AM daily (after backup)
           RetentionDays: 30,
       })
   ```

2. **Backup Execution Tracking**
   - CronJob executions are linked to backup executions via `CronJobExecutionID`
   - Backup status follows CronJob status (success/failure)
   - Metadata tracks relationship between CronJob and backup

3. **Schedule Validation**
   - Prevents backup policy schedule from running before CronJob
   - Ensures different execution times to avoid conflicts
   - Simple hour-based validation for cron expressions

## Architecture Decisions

1. **ExtendedService Pattern**
   - Base `Service` maintains core application logic
   - `ExtendedService` adds backup integration without modifying core
   - Clean separation of concerns

2. **Metadata Storage**
   - Backup policy ID stored in application metadata
   - Enables/disables backup via `backup_enabled` flag
   - Flexible for future extensions

3. **Loose Coupling**
   - Backup service handles its own execution tracking
   - CronJob service notifies backup service of status changes
   - No tight dependencies between services

## Testing Strategy

1. **Unit Tests**
   - Schedule validation logic
   - Mock-based integration tests
   - Error handling scenarios

2. **Integration Tests** (Future)
   - Full stack testing with real Kubernetes
   - Backup storage integration
   - End-to-end execution flow

## Future Enhancements

1. **Advanced Scheduling**
   - Support for complex cron expressions
   - Timezone-aware scheduling
   - Conflict resolution strategies

2. **Backup Hooks**
   - Pre/post backup scripts
   - Custom validation logic
   - Notification integration

3. **Monitoring Integration**
   - Backup metrics and alerts
   - Execution history dashboard
   - Storage usage tracking

## Usage Example

```go
// Create database backup CronJob with automatic backup archival
req := &application.CreateApplicationRequest{
    Name:      "postgres-backup",
    Type:      application.ApplicationTypeCronJob,
    ProjectID: "proj-123",
    Source: application.ApplicationSource{
        Type:  application.SourceTypeImage,
        Image: "postgres-backup:latest",
    },
    Config: application.ApplicationConfig{
        Environment: map[string]string{
            "DB_HOST": "postgres.default.svc",
            "DB_NAME": "production",
        },
    },
    CronSchedule: "0 2 * * *", // Daily at 2 AM
}

backupPolicy := &backup.CreateBackupPolicyRequest{
    StorageID:          "proxmox-storage-01",
    Schedule:           "0 3 * * *", // Archive at 3 AM
    RetentionDays:      30,
    BackupType:         backup.BackupTypeFull,
    CompressionEnabled: true,
    EncryptionEnabled:  true,
}

app, err := extendedService.CreateApplicationWithBackupPolicy(ctx, req, backupPolicy)
```

## Conclusion

The CronJob-backup integration provides a robust foundation for scheduled backup operations with proper lifecycle management and error handling. The TDD approach ensures reliability and maintainability.