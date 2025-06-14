import { BaseGenerator } from './base-generator';

export interface BackupStorage {
  id: string;
  workspaceId: string;
  name: string;
  type: 'proxmox' | 's3' | 'azure' | 'gcs';
  status: 'active' | 'configuring' | 'error' | 'disconnected';
  config: {
    // Proxmox
    proxmoxUrl?: string;
    proxmoxNode?: string;
    datastoreId?: string;
    
    // S3
    s3Bucket?: string;
    s3Region?: string;
    s3Endpoint?: string;
    
    // Azure
    azureContainer?: string;
    azureAccount?: string;
    
    // GCS
    gcsBucket?: string;
    gcsProject?: string;
  };
  capacity: {
    total: number; // GB
    used: number;
    available: number;
  };
  encryption?: {
    enabled: boolean;
    algorithm: 'AES-256' | 'AES-128';
    keyManagement: 'managed' | 'customer';
  };
  retention?: {
    days: number;
    versions: number;
  };
  createdAt: Date;
  updatedAt: Date;
}

export interface BackupPolicy {
  id: string;
  workspaceId: string;
  name: string;
  description?: string;
  enabled: boolean;
  schedule: {
    frequency: 'hourly' | 'daily' | 'weekly' | 'monthly';
    time?: string; // HH:MM
    dayOfWeek?: number; // 0-6
    dayOfMonth?: number; // 1-31
    timezone: string;
  };
  targets: Array<{
    type: 'application' | 'database' | 'volume' | 'namespace';
    selector: {
      labels?: Record<string, string>;
      names?: string[];
      all?: boolean;
    };
  }>;
  storageId: string;
  backupType: 'full' | 'incremental' | 'differential';
  compression: boolean;
  verification: boolean;
  hooks?: {
    preBackup?: string[];
    postBackup?: string[];
  };
  notifications?: {
    email?: string[];
    slack?: string;
    webhook?: string;
  };
  lastExecution?: {
    startTime: Date;
    endTime?: Date;
    status: 'running' | 'succeeded' | 'failed' | 'partial';
    size?: number;
    itemsBackedUp?: number;
    errors?: string[];
  };
  nextExecution?: Date;
  statistics?: {
    totalBackups: number;
    successfulBackups: number;
    failedBackups: number;
    totalSize: number;
    averageDuration: number;
  };
  createdAt: Date;
  updatedAt: Date;
}

export interface Backup {
  id: string;
  policyId: string;
  workspaceId: string;
  name: string;
  type: 'full' | 'incremental' | 'differential';
  status: 'pending' | 'running' | 'completed' | 'failed' | 'expired';
  size: number; // bytes
  compressedSize?: number;
  items: Array<{
    type: 'application' | 'database' | 'volume';
    name: string;
    size: number;
    status: 'completed' | 'failed' | 'skipped';
    error?: string;
  }>;
  metadata: {
    workspaceVersion: string;
    applicationVersions: Record<string, string>;
    labels?: Record<string, string>;
  };
  encryption?: {
    algorithm: string;
    keyId: string;
  };
  verification?: {
    status: 'pending' | 'passed' | 'failed';
    checksum?: string;
    verifiedAt?: Date;
  };
  restore?: {
    available: boolean;
    lastTestedAt?: Date;
    estimatedDuration?: number;
  };
  expiresAt?: Date;
  createdAt: Date;
  completedAt?: Date;
  createdBy: string;
}

export class BackupGenerator extends BaseGenerator<Backup> {
  generate(overrides?: Partial<Backup>): Backup {
    const type = overrides?.type || this.faker.helpers.arrayElement(['full', 'incremental']);
    const status = overrides?.status || this.faker.helpers.weighted(
      ['completed', 'running', 'failed'],
      [0.8, 0.1, 0.1]
    );
    
    const backup: Backup = {
      id: this.generateId('bkp'),
      policyId: overrides?.policyId || this.generateId('bkp-policy'),
      workspaceId: overrides?.workspaceId || this.generateId('ws'),
      name: overrides?.name || `backup-${this.faker.date.recent().toISOString().split('T')[0]}-${this.faker.string.alphanumeric(6)}`,
      type,
      status,
      size: 0, // Will be calculated
      items: this.generateBackupItems(status),
      metadata: {
        workspaceVersion: `v${this.faker.system.semver()}`,
        applicationVersions: {},
        labels: {
          environment: this.faker.helpers.arrayElement(['production', 'staging', 'development']),
          automated: 'true',
        },
      },
      encryption: {
        algorithm: 'AES-256',
        keyId: `key-${this.faker.string.alphanumeric(16)}`,
      },
      restore: {
        available: status === 'completed',
        lastTestedAt: status === 'completed' && this.faker.datatype.boolean({ probability: 0.3 }) ?
          this.faker.date.recent({ days: 30 }) : undefined,
      },
      createdAt: this.faker.date.recent({ days: 7 }),
      createdBy: this.faker.helpers.arrayElement(['backup-policy', 'manual', 'system']),
      ...overrides,
    };
    
    // Calculate total size
    backup.size = backup.items.reduce((total, item) => total + item.size, 0);
    
    // Add compressed size if compression is likely
    if (this.faker.datatype.boolean({ probability: 0.8 })) {
      backup.compressedSize = Math.floor(backup.size * this.faker.number.float({ min: 0.3, max: 0.7 }));
    }
    
    // Add application versions to metadata
    backup.items
      .filter(item => item.type === 'application')
      .forEach(item => {
        backup.metadata.applicationVersions[item.name] = `v${this.faker.number.int({ min: 1, max: 100 })}`;
      });
    
    // Add completion time for completed backups
    if (status === 'completed' || status === 'failed') {
      const duration = this.faker.number.int({ min: 60, max: 3600 }); // 1 min to 1 hour
      backup.completedAt = new Date(backup.createdAt.getTime() + duration * 1000);
    }
    
    // Add verification for completed backups
    if (status === 'completed') {
      backup.verification = {
        status: this.faker.helpers.weighted(['passed', 'pending', 'failed'], [0.8, 0.15, 0.05]),
        checksum: `sha256:${this.faker.git.commitSha()}`,
        verifiedAt: this.faker.date.between({ from: backup.createdAt, to: new Date() }),
      };
      
      // Set expiration
      backup.expiresAt = new Date(backup.createdAt.getTime() + 30 * 24 * 60 * 60 * 1000); // 30 days
    }
    
    return backup;
  }
  
  withTraits(traits: string[]): Backup {
    const overrides: Partial<Backup> = {};
    
    if (traits.includes('large')) {
      overrides.items = this.generateBackupItems('completed', 20);
    }
    
    if (traits.includes('failed')) {
      overrides.status = 'failed';
      overrides.items = this.generateBackupItems('failed');
    }
    
    if (traits.includes('incremental')) {
      overrides.type = 'incremental';
      overrides.size = this.faker.number.int({ min: 100 * 1024 * 1024, max: 5 * 1024 * 1024 * 1024 });
    }
    
    if (traits.includes('verified')) {
      overrides.verification = {
        status: 'passed',
        checksum: `sha256:${this.faker.git.commitSha()}`,
        verifiedAt: new Date(),
      };
      overrides.restore = {
        available: true,
        lastTestedAt: new Date(),
        estimatedDuration: this.faker.number.int({ min: 300, max: 1800 }),
      };
    }
    
    return this.generate(overrides);
  }
  
  generateBackupStorage(overrides?: Partial<BackupStorage>): BackupStorage {
    const type = overrides?.type || this.faker.helpers.arrayElement(['proxmox', 's3', 'azure', 'gcs']);
    const totalCapacity = this.faker.number.int({ min: 100, max: 10000 }); // GB
    const used = this.faker.number.int({ min: 0, max: totalCapacity * 0.8 });
    
    const storage: BackupStorage = {
      id: this.generateId('bkp-storage'),
      workspaceId: overrides?.workspaceId || this.generateId('ws'),
      name: overrides?.name || `${type}-storage-${this.faker.location.city().toLowerCase()}`,
      type,
      status: 'active',
      config: this.generateStorageConfig(type),
      capacity: {
        total: totalCapacity,
        used,
        available: totalCapacity - used,
      },
      encryption: {
        enabled: true,
        algorithm: 'AES-256',
        keyManagement: this.faker.helpers.arrayElement(['managed', 'customer']),
      },
      retention: {
        days: this.faker.helpers.arrayElement([7, 14, 30, 90, 365]),
        versions: this.faker.helpers.arrayElement([3, 5, 10, -1]), // -1 means unlimited
      },
      createdAt: this.faker.date.past({ years: 1 }),
      updatedAt: this.faker.date.recent({ days: 30 }),
      ...overrides,
    };
    
    return storage;
  }
  
  generateBackupPolicy(overrides?: Partial<BackupPolicy>): BackupPolicy {
    const frequency = overrides?.schedule?.frequency || 
      this.faker.helpers.arrayElement(['daily', 'weekly', 'hourly']);
    
    const policy: BackupPolicy = {
      id: this.generateId('bkp-policy'),
      workspaceId: overrides?.workspaceId || this.generateId('ws'),
      name: overrides?.name || `${frequency}-backup-${this.faker.word.noun()}`,
      description: overrides?.description || `Automated ${frequency} backup policy`,
      enabled: overrides?.enabled !== undefined ? overrides.enabled : true,
      schedule: this.generateSchedule(frequency),
      targets: overrides?.targets || this.generateTargets(),
      storageId: overrides?.storageId || this.generateId('bkp-storage'),
      backupType: overrides?.backupType || 
        frequency === 'hourly' ? 'incremental' : 'full',
      compression: true,
      verification: true,
      notifications: {
        email: [this.faker.internet.email()],
        slack: this.faker.datatype.boolean({ probability: 0.5 }) ? 
          '#backup-notifications' : undefined,
      },
      createdAt: this.faker.date.past({ years: 1 }),
      updatedAt: this.faker.date.recent({ days: 7 }),
      ...overrides,
    };
    
    // Generate execution history
    if (policy.enabled) {
      policy.lastExecution = this.generateLastExecution();
      policy.nextExecution = this.calculateNextExecution(policy.schedule);
      policy.statistics = this.generatePolicyStatistics();
    }
    
    // Add hooks for database backups
    const hasDatabaseTarget = policy.targets.some(t => t.type === 'database');
    if (hasDatabaseTarget) {
      policy.hooks = {
        preBackup: ['flush-logs', 'create-snapshot'],
        postBackup: ['verify-backup', 'cleanup-snapshots'],
      };
    }
    
    return policy;
  }
  
  private generateBackupItems(status: string, count?: number): Backup['items'] {
    const itemCount = count || this.faker.number.int({ min: 3, max: 10 });
    const items = [];
    
    for (let i = 0; i < itemCount; i++) {
      const type = this.faker.helpers.arrayElement(['application', 'database', 'volume']);
      const itemStatus = status === 'failed' && i === 0 ? 'failed' : 
        status === 'running' && i > itemCount / 2 ? 'completed' : 'completed';
      
      const item = {
        type,
        name: this.generateItemName(type),
        size: this.generateItemSize(type),
        status: itemStatus,
        error: itemStatus === 'failed' ? this.faker.helpers.arrayElement([
          'Connection timeout',
          'Insufficient permissions',
          'Resource locked',
          'Disk space error',
        ]) : undefined,
      };
      
      items.push(item);
    }
    
    return items;
  }
  
  private generateItemName(type: string): string {
    switch (type) {
      case 'application':
        return `${this.faker.hacker.noun()}-${this.faker.helpers.arrayElement(['api', 'web', 'worker'])}`;
      case 'database':
        return `${this.faker.helpers.arrayElement(['postgres', 'mysql', 'mongodb'])}-${this.faker.database.collation()}`;
      case 'volume':
        return `pvc-${this.faker.string.uuid().substring(0, 8)}`;
      default:
        return `item-${this.faker.string.alphanumeric(8)}`;
    }
  }
  
  private generateItemSize(type: string): number {
    switch (type) {
      case 'application':
        return this.faker.number.int({ min: 50 * 1024 * 1024, max: 2 * 1024 * 1024 * 1024 }); // 50MB - 2GB
      case 'database':
        return this.faker.number.int({ min: 100 * 1024 * 1024, max: 50 * 1024 * 1024 * 1024 }); // 100MB - 50GB
      case 'volume':
        return this.faker.number.int({ min: 1 * 1024 * 1024 * 1024, max: 100 * 1024 * 1024 * 1024 }); // 1GB - 100GB
      default:
        return this.faker.number.int({ min: 1024 * 1024, max: 1024 * 1024 * 1024 });
    }
  }
  
  private generateStorageConfig(type: string): BackupStorage['config'] {
    switch (type) {
      case 'proxmox':
        return {
          proxmoxUrl: `https://proxmox-${this.faker.number.int({ min: 1, max: 5 })}.example.com:8006`,
          proxmoxNode: `node-${this.faker.number.int({ min: 1, max: 10 })}`,
          datastoreId: `backup-store-${this.faker.number.int({ min: 1, max: 3 })}`,
        };
        
      case 's3':
        return {
          s3Bucket: `backup-${this.faker.company.name().toLowerCase().replace(/\s+/g, '-')}`,
          s3Region: this.faker.helpers.arrayElement(['us-east-1', 'us-west-2', 'eu-west-1']),
          s3Endpoint: this.faker.datatype.boolean({ probability: 0.3 }) ? 
            `https://s3.${this.faker.internet.domainName()}` : undefined,
        };
        
      case 'azure':
        return {
          azureContainer: 'backups',
          azureAccount: `storage${this.faker.string.alphanumeric(8)}`,
        };
        
      case 'gcs':
        return {
          gcsBucket: `backup-${this.faker.company.name().toLowerCase().replace(/\s+/g, '-')}`,
          gcsProject: `project-${this.faker.string.alphanumeric(8)}`,
        };
        
      default:
        return {};
    }
  }
  
  private generateSchedule(frequency: string): BackupPolicy['schedule'] {
    const schedule: BackupPolicy['schedule'] = {
      frequency: frequency as any,
      timezone: this.faker.helpers.arrayElement(['UTC', 'America/New_York', 'Europe/London']),
    };
    
    switch (frequency) {
      case 'hourly':
        // No additional config needed
        break;
        
      case 'daily':
        schedule.time = `${this.faker.number.int({ min: 0, max: 23 })}:00`;
        break;
        
      case 'weekly':
        schedule.time = `${this.faker.number.int({ min: 0, max: 23 })}:00`;
        schedule.dayOfWeek = this.faker.number.int({ min: 0, max: 6 });
        break;
        
      case 'monthly':
        schedule.time = `${this.faker.number.int({ min: 0, max: 23 })}:00`;
        schedule.dayOfMonth = this.faker.number.int({ min: 1, max: 28 });
        break;
    }
    
    return schedule;
  }
  
  private generateTargets(): BackupPolicy['targets'] {
    const targetCount = this.faker.number.int({ min: 1, max: 3 });
    const targets = [];
    
    for (let i = 0; i < targetCount; i++) {
      const type = this.faker.helpers.arrayElement(['application', 'database', 'volume', 'namespace']);
      
      targets.push({
        type,
        selector: this.faker.helpers.arrayElement([
          { labels: { tier: 'production', backup: 'enabled' } },
          { names: this.faker.helpers.multiple(() => this.faker.hacker.noun(), { count: 3 }) },
          { all: true },
        ]),
      });
    }
    
    return targets;
  }
  
  private generateLastExecution(): BackupPolicy['lastExecution'] {
    const status = this.faker.helpers.weighted(
      ['succeeded', 'running', 'failed', 'partial'],
      [0.7, 0.1, 0.1, 0.1]
    );
    
    const startTime = this.faker.date.recent({ days: 1 });
    const duration = status === 'running' ? undefined : 
      this.faker.number.int({ min: 300, max: 7200 }); // 5 min to 2 hours
    
    return {
      startTime,
      endTime: duration ? new Date(startTime.getTime() + duration * 1000) : undefined,
      status,
      size: status !== 'failed' ? 
        this.faker.number.int({ min: 1024 * 1024 * 1024, max: 100 * 1024 * 1024 * 1024 }) : undefined,
      itemsBackedUp: status !== 'failed' ? 
        this.faker.number.int({ min: 5, max: 50 }) : 0,
      errors: status === 'failed' || status === 'partial' ? 
        this.faker.helpers.multiple(() => this.faker.hacker.phrase(), { count: { min: 1, max: 3 } }) : undefined,
    };
  }
  
  private calculateNextExecution(schedule: BackupPolicy['schedule']): Date {
    const now = new Date();
    
    switch (schedule.frequency) {
      case 'hourly':
        return new Date(now.getTime() + 60 * 60 * 1000);
        
      case 'daily':
        const tomorrow = new Date(now);
        tomorrow.setDate(tomorrow.getDate() + 1);
        if (schedule.time) {
          const [hour] = schedule.time.split(':');
          tomorrow.setHours(parseInt(hour), 0, 0, 0);
        }
        return tomorrow;
        
      case 'weekly':
        const nextWeek = new Date(now);
        nextWeek.setDate(nextWeek.getDate() + 7);
        return nextWeek;
        
      case 'monthly':
        const nextMonth = new Date(now);
        nextMonth.setMonth(nextMonth.getMonth() + 1);
        if (schedule.dayOfMonth) {
          nextMonth.setDate(schedule.dayOfMonth);
        }
        return nextMonth;
        
      default:
        return new Date(now.getTime() + 24 * 60 * 60 * 1000);
    }
  }
  
  private generatePolicyStatistics(): BackupPolicy['statistics'] {
    const total = this.faker.number.int({ min: 10, max: 1000 });
    const failureRate = this.faker.number.float({ min: 0, max: 0.15 });
    const failed = Math.floor(total * failureRate);
    
    return {
      totalBackups: total,
      successfulBackups: total - failed,
      failedBackups: failed,
      totalSize: this.faker.number.int({ 
        min: 10 * 1024 * 1024 * 1024, 
        max: 10 * 1024 * 1024 * 1024 * 1024 
      }), // 10GB - 10TB
      averageDuration: this.faker.number.int({ min: 300, max: 3600 }), // 5 min to 1 hour
    };
  }
}