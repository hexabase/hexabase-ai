import { BaseGenerator } from './base-generator';

export interface CronJob {
  id: string;
  projectId: string;
  name: string;
  schedule: string;
  timezone?: string;
  enabled: boolean;
  status: 'active' | 'paused' | 'error';
  jobTemplate: {
    image: string;
    command?: string[];
    args?: string[];
    env?: Record<string, string>;
    resources?: {
      requests?: { cpu: string; memory: string };
      limits?: { cpu: string; memory: string };
    };
    timeout?: number;
    retries?: number;
  };
  lastExecution?: {
    id: string;
    startTime: Date;
    endTime?: Date;
    status: 'running' | 'succeeded' | 'failed';
    duration?: number;
    logs?: string;
    exitCode?: number;
  };
  nextExecution?: Date;
  history?: Array<{
    id: string;
    startTime: Date;
    endTime: Date;
    status: 'succeeded' | 'failed';
    duration: number;
    exitCode: number;
  }>;
  statistics?: {
    totalRuns: number;
    successfulRuns: number;
    failedRuns: number;
    averageDuration: number;
    lastSuccess?: Date;
    lastFailure?: Date;
  };
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;
}

export class CronJobGenerator extends BaseGenerator<CronJob> {
  private commonSchedules = [
    { cron: '0 * * * *', description: 'Every hour' },
    { cron: '0 0 * * *', description: 'Daily at midnight' },
    { cron: '0 2 * * *', description: 'Daily at 2 AM' },
    { cron: '0 0 * * 0', description: 'Weekly on Sunday' },
    { cron: '0 0 1 * *', description: 'Monthly on the 1st' },
    { cron: '*/5 * * * *', description: 'Every 5 minutes' },
    { cron: '*/15 * * * *', description: 'Every 15 minutes' },
    { cron: '0 */6 * * *', description: 'Every 6 hours' },
    { cron: '30 3 * * 1-5', description: 'Weekdays at 3:30 AM' },
  ];
  
  private jobTypes = [
    { 
      name: 'backup', 
      image: 'backup-tool:latest',
      command: ['/bin/sh', '-c'],
      args: ['backup.sh --type full --destination s3://backups'],
    },
    {
      name: 'cleanup',
      image: 'alpine:latest',
      command: ['/bin/sh', '-c'],
      args: ['find /tmp -type f -mtime +7 -delete'],
    },
    {
      name: 'report',
      image: 'python:3.9-slim',
      command: ['python'],
      args: ['generate_report.py', '--format', 'pdf'],
    },
    {
      name: 'sync',
      image: 'rsync:latest',
      command: ['rsync'],
      args: ['-avz', '/source/', '/destination/'],
    },
    {
      name: 'health-check',
      image: 'curlimages/curl:latest',
      command: ['curl'],
      args: ['-f', 'http://api/health'],
    },
  ];
  
  generate(overrides?: Partial<CronJob>): CronJob {
    const jobType = this.faker.helpers.arrayElement(this.jobTypes);
    const schedule = this.faker.helpers.arrayElement(this.commonSchedules);
    const name = overrides?.name || `${jobType.name}-${this.faker.word.adjective()}`;
    
    const cronjob: CronJob = {
      id: this.generateId('cron'),
      projectId: overrides?.projectId || this.generateId('proj'),
      name,
      schedule: overrides?.schedule || schedule.cron,
      timezone: this.faker.helpers.arrayElement(['UTC', 'America/New_York', 'Europe/London', 'Asia/Tokyo']),
      enabled: this.faker.datatype.boolean({ probability: 0.8 }),
      status: 'active',
      jobTemplate: {
        image: jobType.image,
        command: jobType.command,
        args: jobType.args,
        env: this.generateEnvironment(jobType.name),
        resources: {
          requests: {
            cpu: this.faker.helpers.arrayElement(['100m', '250m', '500m']),
            memory: this.faker.helpers.arrayElement(['128Mi', '256Mi', '512Mi']),
          },
          limits: {
            cpu: this.faker.helpers.arrayElement(['500m', '1', '2']),
            memory: this.faker.helpers.arrayElement(['512Mi', '1Gi', '2Gi']),
          },
        },
        timeout: this.faker.number.int({ min: 300, max: 3600 }), // 5 min to 1 hour
        retries: this.faker.number.int({ min: 0, max: 3 }),
      },
      labels: {
        type: jobType.name,
        schedule: schedule.description.toLowerCase().replace(/\s+/g, '-'),
      },
      annotations: {
        description: `${schedule.description} ${jobType.name} job`,
      },
      createdAt: this.faker.date.past({ years: 1 }),
      updatedAt: this.faker.date.recent({ days: 7 }),
      createdBy: this.faker.internet.email(),
      ...overrides,
    };
    
    // Generate execution history
    if (cronjob.enabled) {
      cronjob.history = this.generateExecutionHistory(cronjob.schedule, cronjob.createdAt);
      cronjob.statistics = this.calculateStatistics(cronjob.history);
      
      // Set last execution
      if (cronjob.history.length > 0) {
        const lastExec = cronjob.history[0];
        cronjob.lastExecution = {
          ...lastExec,
          logs: this.generateLogs(jobType.name, lastExec.status),
        };
        
        // If still running
        if (this.faker.datatype.boolean({ probability: 0.1 })) {
          cronjob.lastExecution.status = 'running';
          cronjob.lastExecution.endTime = undefined;
          cronjob.lastExecution.duration = undefined;
        }
      }
      
      // Calculate next execution
      cronjob.nextExecution = this.calculateNextExecution(cronjob.schedule);
    }
    
    return cronjob;
  }
  
  withTraits(traits: string[]): CronJob {
    const overrides: Partial<CronJob> = {};
    
    if (traits.includes('frequent')) {
      overrides.schedule = '*/5 * * * *'; // Every 5 minutes
    }
    
    if (traits.includes('failing')) {
      overrides.status = 'error';
      overrides.lastExecution = {
        id: this.generateId('exec'),
        startTime: this.faker.date.recent({ days: 1 }),
        endTime: this.faker.date.recent({ days: 1 }),
        status: 'failed',
        duration: this.faker.number.int({ min: 1, max: 300 }),
        exitCode: this.faker.helpers.arrayElement([1, 2, 127, 255]),
        logs: 'Error: Connection timeout\nFailed to complete operation',
      };
    }
    
    if (traits.includes('paused')) {
      overrides.enabled = false;
      overrides.status = 'paused';
    }
    
    if (traits.includes('longRunning')) {
      overrides.jobTemplate = {
        image: 'data-processor:latest',
        command: ['python'],
        args: ['process_large_dataset.py'],
        resources: {
          requests: { cpu: '2', memory: '4Gi' },
          limits: { cpu: '4', memory: '8Gi' },
        },
        timeout: 14400, // 4 hours
      };
    }
    
    return this.generate(overrides);
  }
  
  private generateEnvironment(jobType: string): Record<string, string> {
    const common = {
      LOG_LEVEL: this.faker.helpers.arrayElement(['INFO', 'DEBUG', 'WARN']),
      NODE_ENV: 'production',
    };
    
    const specific: Record<string, Record<string, string>> = {
      backup: {
        S3_BUCKET: 'company-backups',
        RETENTION_DAYS: '30',
        COMPRESSION: 'gzip',
      },
      report: {
        REPORT_TYPE: this.faker.helpers.arrayElement(['daily', 'weekly', 'monthly']),
        EMAIL_RECIPIENTS: 'reports@company.com',
        INCLUDE_CHARTS: 'true',
      },
      sync: {
        SYNC_MODE: 'incremental',
        DELETE_ORPHANS: 'false',
        BANDWIDTH_LIMIT: '100M',
      },
    };
    
    return { ...common, ...(specific[jobType] || {}) };
  }
  
  private generateExecutionHistory(schedule: string, since: Date): CronJob['history'] {
    const history = [];
    const executions = this.calculatePastExecutions(schedule, since, 20);
    
    for (const execTime of executions) {
      const succeeded = this.faker.datatype.boolean({ probability: 0.85 });
      const duration = this.faker.number.int({ min: 10, max: 600 });
      
      history.push({
        id: this.generateId('exec'),
        startTime: execTime,
        endTime: new Date(execTime.getTime() + duration * 1000),
        status: succeeded ? 'succeeded' : 'failed',
        duration,
        exitCode: succeeded ? 0 : this.faker.helpers.arrayElement([1, 2, 127]),
      });
    }
    
    return history;
  }
  
  private calculateStatistics(history: CronJob['history']): CronJob['statistics'] {
    if (!history || history.length === 0) {
      return {
        totalRuns: 0,
        successfulRuns: 0,
        failedRuns: 0,
        averageDuration: 0,
      };
    }
    
    const successful = history.filter(h => h.status === 'succeeded');
    const failed = history.filter(h => h.status === 'failed');
    const totalDuration = history.reduce((sum, h) => sum + h.duration, 0);
    
    return {
      totalRuns: history.length,
      successfulRuns: successful.length,
      failedRuns: failed.length,
      averageDuration: Math.round(totalDuration / history.length),
      lastSuccess: successful.length > 0 ? successful[0].endTime : undefined,
      lastFailure: failed.length > 0 ? failed[0].endTime : undefined,
    };
  }
  
  private calculatePastExecutions(cronExpression: string, since: Date, limit: number): Date[] {
    // Simplified - in real implementation would use a cron parser
    const executions: Date[] = [];
    const now = new Date();
    const interval = this.getIntervalFromCron(cronExpression);
    
    let current = new Date(now.getTime() - interval);
    while (current > since && executions.length < limit) {
      executions.push(new Date(current));
      current = new Date(current.getTime() - interval);
    }
    
    return executions;
  }
  
  private calculateNextExecution(cronExpression: string): Date {
    // Simplified - in real implementation would use a cron parser
    const interval = this.getIntervalFromCron(cronExpression);
    return new Date(Date.now() + interval);
  }
  
  private getIntervalFromCron(cron: string): number {
    // Very simplified mapping
    if (cron.includes('*/5 * * * *')) return 5 * 60 * 1000;
    if (cron.includes('*/15 * * * *')) return 15 * 60 * 1000;
    if (cron.includes('0 * * * *')) return 60 * 60 * 1000;
    if (cron.includes('0 0 * * *')) return 24 * 60 * 60 * 1000;
    if (cron.includes('0 0 * * 0')) return 7 * 24 * 60 * 60 * 1000;
    return 60 * 60 * 1000; // Default to hourly
  }
  
  private generateLogs(jobType: string, status: 'succeeded' | 'failed' | 'running'): string {
    const logs = [];
    
    // Start logs
    logs.push(`[${new Date().toISOString()}] Starting ${jobType} job`);
    logs.push(`[${new Date().toISOString()}] Initializing environment`);
    
    // Job-specific logs
    switch (jobType) {
      case 'backup':
        logs.push('Connecting to database...');
        logs.push('Creating backup snapshot...');
        logs.push('Compressing data (gzip)...');
        logs.push('Uploading to S3...');
        break;
        
      case 'report':
        logs.push('Fetching data from analytics API...');
        logs.push('Processing 10,234 records...');
        logs.push('Generating charts...');
        logs.push('Creating PDF report...');
        break;
        
      case 'cleanup':
        logs.push('Scanning directories...');
        logs.push('Found 156 files older than 7 days');
        logs.push('Removing old files...');
        break;
    }
    
    // Status-specific logs
    if (status === 'failed') {
      logs.push('ERROR: ' + this.faker.helpers.arrayElement([
        'Connection timeout',
        'Permission denied',
        'Resource not found',
        'Out of memory',
      ]));
      logs.push('Job failed with exit code 1');
    } else if (status === 'succeeded') {
      logs.push('Job completed successfully');
    }
    
    return logs.join('\n');
  }
}