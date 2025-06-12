# ADR-007: Backup and Disaster Recovery Architecture

**Date**: 2025-06-09  
**Status**: Partially Implemented  
**Authors**: Infrastructure Team

## 1. Background

Hexabase AI needed a comprehensive backup and disaster recovery solution for:
- User application data (persistent volumes)
- Platform configuration and state
- Database backups (PostgreSQL, Redis)
- Disaster recovery with RTO < 4 hours, RPO < 1 hour
- Compliance with data retention requirements
- Cost-effective storage for long-term retention

The solution needed to support both shared and dedicated plan customers with different SLAs.

## 2. Status

**Partially Implemented** - Backup features for dedicated plans are implemented. Shared plan backups and full DR are in progress.

## 3. Other Options Considered

### Option A: Velero-based Kubernetes Backup
- Native Kubernetes backup tool
- Snapshot-based backups
- Cloud provider integration

### Option B: Storage-level Snapshots
- Direct storage snapshots
- Proxmox backup integration
- Volume-level consistency

### Option C: Hybrid Application-aware Backup
- Combination of Velero and storage snapshots
- Application-specific backup strategies
- Tiered storage approach

## 4. What Was Decided

We chose **Option C: Hybrid Application-aware Backup** with:
- Velero for Kubernetes resource backup
- Proxmox snapshots for dedicated plan storage
- Application-aware backups for databases
- CronJob integration for scheduled backups
- S3-compatible object storage for long-term retention

## 5. Why Did You Choose It?

- **Flexibility**: Different strategies for different data types
- **Consistency**: Application-aware backups ensure data integrity
- **Cost-effective**: Tiered storage reduces long-term costs
- **Performance**: Minimal impact on running workloads
- **Compliance**: Meets retention and recovery requirements

## 6. Why Didn't You Choose the Other Options?

### Why not Velero-only?
- Limited support for external storage systems
- No application-level consistency
- Slower recovery for large volumes

### Why not Storage Snapshots only?
- No Kubernetes resource backup
- Platform-specific limitations
- Difficult cross-platform recovery

## 7. What Has Not Been Decided

- Cross-region disaster recovery implementation
- Backup encryption key management strategy
- Automated DR testing procedures
- Backup cost allocation model

## 8. Considerations

### Backup Architecture
```
┌─────────────────┐
│   Workspaces    │
└────────┬────────┘
         │
┌────────┴────────┐
│  Backup Service │
├─────────────────┤
│ Backup Policies │
├─────────────────┤
│ Storage Manager │
└───┬─────────┬───┘
    │         │
┌───▼───┐ ┌───▼────┐
│Velero │ │Proxmox │
│       │ │Backup  │
└───┬───┘ └───┬────┘
    │         │
┌───▼─────────▼───┐
│ S3 Object Store │
└─────────────────┘
```

### Backup Types and Strategies

| Backup Type | Method | Frequency | Retention |
|-------------|--------|-----------|-----------|
| Platform Config | Velero | Daily | 30 days |
| Application PVs | Snapshot | Hourly | 7 days |
| PostgreSQL | pg_dump | 4 hours | 30 days |
| Redis | RDB snapshot | Daily | 7 days |
| Full DR | All above | Weekly | 90 days |

### Backup Policy Configuration
```go
type BackupPolicy struct {
    ID          string
    WorkspaceID string
    Schedule    string // Cron expression
    Retention   time.Duration
    Type        BackupType
    Targets     []BackupTarget
    Encryption  EncryptionConfig
}

type BackupTarget struct {
    Type     string // "pvc", "database", "all"
    Selector map[string]string
}
```

### Recovery Procedures

#### Application Recovery
1. Restore Kubernetes resources via Velero
2. Restore persistent volumes from snapshots
3. Verify application health
4. Update DNS/routing

#### Database Recovery
```bash
# PostgreSQL point-in-time recovery
pg_restore -h localhost -U postgres \
  --clean --create -d postgres \
  backup_2025_06_09_1200.dump

# Redis recovery
redis-cli --rdb /backup/dump.rdb
```

### Storage Tiers
- **Hot**: Last 7 days - Fast SSD storage
- **Warm**: 8-30 days - Standard storage
- **Cold**: 31-365 days - Archive storage

### Encryption
- AES-256 encryption at rest
- Per-workspace encryption keys
- Key rotation every 90 days
- Hardware security module (HSM) for key storage

### Cost Model
| Plan | Backup Frequency | Retention | Monthly Cost |
|------|-----------------|-----------|--------------|
| Shared | Daily | 7 days | Included |
| Dedicated | Hourly | 30 days | $50/TB |
| Enterprise | 15 min | 365 days | Custom |

### Monitoring and Alerting
- Backup job success/failure alerts
- Storage capacity warnings
- Recovery time tracking
- Backup size trends

### Compliance Considerations
- GDPR right to deletion
- Data residency requirements
- Audit trail of all backup operations
- Encryption key escrow for compliance

### Future Enhancements
- Continuous data protection (CDP)
- Cross-region replication
- Automated DR drills
- AI-powered backup optimization