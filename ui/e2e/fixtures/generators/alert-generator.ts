import { BaseGenerator } from './base-generator';

export interface Alert {
  id: string;
  name: string;
  description?: string;
  severity: 'critical' | 'warning' | 'info';
  status: 'active' | 'resolved' | 'acknowledged' | 'silenced';
  source: {
    type: 'metric' | 'log' | 'synthetic' | 'custom';
    query?: string;
    metric?: string;
    labels?: Record<string, string>;
  };
  condition: {
    type: 'threshold' | 'rate' | 'absence' | 'pattern';
    operator: '>' | '<' | '>=' | '<=' | '==' | '!=';
    value: number | string;
    duration?: string; // e.g., "5m", "1h"
    aggregation?: 'avg' | 'sum' | 'min' | 'max' | 'count';
  };
  annotations?: {
    summary?: string;
    runbook?: string;
    dashboard?: string;
    [key: string]: string | undefined;
  };
  notifications: Array<{
    channel: 'email' | 'slack' | 'pagerduty' | 'webhook' | 'sms';
    config: {
      recipients?: string[];
      webhook?: string;
      slackChannel?: string;
      pagerdutyKey?: string;
    };
    cooldown?: number; // seconds
  }>;
  silence?: {
    id: string;
    startTime: Date;
    endTime: Date;
    reason: string;
    createdBy: string;
  };
  history: Array<{
    timestamp: Date;
    event: 'triggered' | 'resolved' | 'acknowledged' | 'escalated';
    value?: number | string;
    message?: string;
    actor?: string;
  }>;
  metadata: {
    workspace?: string;
    project?: string;
    application?: string;
    environment?: string;
    team?: string;
  };
  createdAt: Date;
  updatedAt: Date;
  triggeredAt?: Date;
  resolvedAt?: Date;
  acknowledgedAt?: Date;
  acknowledgedBy?: string;
}

export class AlertGenerator extends BaseGenerator<Alert> {
  private alertTemplates = [
    {
      name: 'High CPU Usage',
      source: { type: 'metric' as const, metric: 'cpu_usage_percent' },
      condition: { type: 'threshold' as const, operator: '>' as const, value: 80, duration: '5m' },
      severity: 'warning' as const,
    },
    {
      name: 'Memory Pressure',
      source: { type: 'metric' as const, metric: 'memory_usage_percent' },
      condition: { type: 'threshold' as const, operator: '>' as const, value: 90, duration: '10m' },
      severity: 'critical' as const,
    },
    {
      name: 'High Error Rate',
      source: { type: 'metric' as const, metric: 'http_error_rate' },
      condition: { type: 'rate' as const, operator: '>' as const, value: 0.05, duration: '5m' },
      severity: 'critical' as const,
    },
    {
      name: 'Service Down',
      source: { type: 'synthetic' as const, query: 'up{job="api"} == 0' },
      condition: { type: 'threshold' as const, operator: '==' as const, value: 0, duration: '1m' },
      severity: 'critical' as const,
    },
    {
      name: 'Disk Space Low',
      source: { type: 'metric' as const, metric: 'disk_free_percent' },
      condition: { type: 'threshold' as const, operator: '<' as const, value: 10, duration: '5m' },
      severity: 'warning' as const,
    },
    {
      name: 'No Data Received',
      source: { type: 'metric' as const, metric: 'data_ingestion_rate' },
      condition: { type: 'absence' as const, operator: '==' as const, value: 0, duration: '15m' },
      severity: 'warning' as const,
    },
  ];
  
  generate(overrides?: Partial<Alert>): Alert {
    const template = this.faker.helpers.arrayElement(this.alertTemplates);
    const status = overrides?.status || this.faker.helpers.weighted(
      ['active', 'resolved', 'acknowledged'],
      [0.3, 0.6, 0.1]
    );
    
    const alert: Alert = {
      id: this.generateId('alert'),
      name: overrides?.name || template.name,
      description: overrides?.description || `Alert for ${template.name.toLowerCase()} in production environment`,
      severity: overrides?.severity || template.severity,
      status,
      source: overrides?.source || {
        ...template.source,
        labels: {
          namespace: this.faker.helpers.arrayElement(['default', 'production', 'staging']),
          app: this.faker.hacker.noun(),
        },
      },
      condition: overrides?.condition || template.condition,
      annotations: {
        summary: `${template.name} has been triggered`,
        runbook: `https://runbooks.example.com/${template.name.toLowerCase().replace(/\s+/g, '-')}`,
        dashboard: `https://grafana.example.com/d/${this.faker.string.alphanumeric(10)}`,
      },
      notifications: this.generateNotifications(),
      history: [],
      metadata: {
        workspace: this.generateId('ws'),
        project: this.generateId('proj'),
        application: this.faker.hacker.noun(),
        environment: this.faker.helpers.arrayElement(['production', 'staging', 'development']),
        team: this.faker.helpers.arrayElement(['platform', 'backend', 'frontend', 'sre']),
      },
      createdAt: this.faker.date.past({ years: 1 }),
      updatedAt: this.faker.date.recent({ days: 7 }),
      ...overrides,
    };
    
    // Generate history based on status
    alert.history = this.generateAlertHistory(alert.status, alert.createdAt);
    
    // Set status timestamps
    const triggeredEvent = alert.history.find(h => h.event === 'triggered');
    if (triggeredEvent) {
      alert.triggeredAt = triggeredEvent.timestamp;
    }
    
    if (status === 'resolved') {
      const resolvedEvent = alert.history.find(h => h.event === 'resolved');
      if (resolvedEvent) {
        alert.resolvedAt = resolvedEvent.timestamp;
      }
    }
    
    if (status === 'acknowledged') {
      const ackEvent = alert.history.find(h => h.event === 'acknowledged');
      if (ackEvent) {
        alert.acknowledgedAt = ackEvent.timestamp;
        alert.acknowledgedBy = ackEvent.actor;
      }
    }
    
    // Add silence if status is silenced
    if (status === 'silenced' || this.faker.datatype.boolean({ probability: 0.1 })) {
      alert.silence = {
        id: this.generateId('silence'),
        startTime: this.faker.date.recent({ days: 1 }),
        endTime: this.faker.date.future({ years: 0.01 }), // Next few days
        reason: this.faker.helpers.arrayElement([
          'Scheduled maintenance',
          'Known issue - fix in progress',
          'False positive - tuning required',
          'Testing in progress',
        ]),
        createdBy: this.faker.internet.email(),
      };
      alert.status = 'silenced';
    }
    
    return alert;
  }
  
  withTraits(traits: string[]): Alert {
    const overrides: Partial<Alert> = {};
    
    if (traits.includes('critical')) {
      overrides.severity = 'critical';
      overrides.status = 'active';
      overrides.notifications = [
        {
          channel: 'pagerduty',
          config: { pagerdutyKey: 'service-key-123' },
          cooldown: 0,
        },
        {
          channel: 'slack',
          config: { slackChannel: '#incidents' },
          cooldown: 300,
        },
      ];
    }
    
    if (traits.includes('flapping')) {
      overrides.history = this.generateFlappingHistory();
    }
    
    if (traits.includes('longRunning')) {
      overrides.triggeredAt = this.faker.date.past({ years: 0.1 }); // Days/weeks ago
      overrides.status = 'active';
    }
    
    if (traits.includes('multiProject')) {
      overrides.source = {
        type: 'metric',
        metric: 'resource_usage',
        labels: {
          workspace: this.generateId('ws'),
          project: '*', // Wildcard
        },
      };
    }
    
    return this.generate(overrides);
  }
  
  generateAlertRule(overrides?: Partial<AlertRule>): AlertRule {
    const template = this.faker.helpers.arrayElement(this.alertTemplates);
    
    return {
      id: this.generateId('rule'),
      name: overrides?.name || `${template.name} Rule`,
      enabled: overrides?.enabled !== undefined ? overrides.enabled : true,
      expression: this.generatePrometheusQuery(template.source.metric || ''),
      duration: template.condition.duration || '5m',
      labels: {
        severity: template.severity,
        team: this.faker.helpers.arrayElement(['platform', 'backend', 'frontend']),
      },
      annotations: {
        summary: `{{ $labels.instance }} ${template.name}`,
        description: `{{ $labels.instance }} has {{ $value }} ${template.source.metric}`,
      },
      ...overrides,
    };
  }
  
  private generateNotifications(): Alert['notifications'] {
    const channels: Alert['notifications'] = [];
    
    // Always add email
    channels.push({
      channel: 'email',
      config: {
        recipients: this.faker.helpers.multiple(
          () => this.faker.internet.email(),
          { count: { min: 1, max: 3 } }
        ),
      },
      cooldown: 3600, // 1 hour
    });
    
    // Add Slack for most alerts
    if (this.faker.datatype.boolean({ probability: 0.7 })) {
      channels.push({
        channel: 'slack',
        config: {
          slackChannel: this.faker.helpers.arrayElement([
            '#alerts', '#ops-alerts', '#team-notifications', '#incidents'
          ]),
        },
        cooldown: 1800, // 30 minutes
      });
    }
    
    // Add PagerDuty for critical alerts
    if (this.faker.datatype.boolean({ probability: 0.3 })) {
      channels.push({
        channel: 'pagerduty',
        config: {
          pagerdutyKey: `service-${this.faker.string.alphanumeric(12)}`,
        },
        cooldown: 0, // No cooldown for PagerDuty
      });
    }
    
    return channels;
  }
  
  private generateAlertHistory(currentStatus: string, since: Date): Alert['history'] {
    const history: Alert['history'] = [];
    const now = new Date();
    
    // Always start with a trigger event
    const triggerTime = this.faker.date.between({ from: since, to: now });
    history.push({
      timestamp: triggerTime,
      event: 'triggered',
      value: this.faker.number.float({ min: 81, max: 99, fractionDigits: 1 }),
      message: 'Alert condition met',
    });
    
    // Add intermediate events based on current status
    if (currentStatus === 'acknowledged' || currentStatus === 'resolved') {
      const ackTime = this.faker.date.between({ from: triggerTime, to: now });
      history.push({
        timestamp: ackTime,
        event: 'acknowledged',
        actor: this.faker.internet.email(),
        message: this.faker.helpers.arrayElement([
          'Looking into this',
          'Known issue, working on fix',
          'Escalated to engineering',
        ]),
      });
    }
    
    if (currentStatus === 'resolved') {
      const resolveTime = this.faker.date.between({ 
        from: history[history.length - 1].timestamp, 
        to: now 
      });
      history.push({
        timestamp: resolveTime,
        event: 'resolved',
        value: this.faker.number.float({ min: 10, max: 79, fractionDigits: 1 }),
        message: 'Alert condition cleared',
        actor: this.faker.helpers.arrayElement([undefined, this.faker.internet.email()]),
      });
    }
    
    // Sort by timestamp descending
    return history.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
  }
  
  private generateFlappingHistory(): Alert['history'] {
    const history: Alert['history'] = [];
    const events = ['triggered', 'resolved', 'triggered', 'resolved', 'triggered'];
    let currentTime = this.faker.date.recent({ days: 1 });
    
    for (const event of events) {
      history.push({
        timestamp: new Date(currentTime),
        event: event as any,
        value: event === 'triggered' ? 
          this.faker.number.float({ min: 81, max: 99 }) : 
          this.faker.number.float({ min: 60, max: 79 }),
        message: event === 'triggered' ? 'Threshold exceeded' : 'Returned to normal',
      });
      
      currentTime = new Date(currentTime.getTime() + this.faker.number.int({ min: 300000, max: 1800000 })); // 5-30 min
    }
    
    return history;
  }
  
  private generatePrometheusQuery(metric: string): string {
    const queries = {
      cpu_usage_percent: 'avg(rate(container_cpu_usage_seconds_total[5m])) by (pod) * 100',
      memory_usage_percent: '(container_memory_working_set_bytes / container_spec_memory_limit_bytes) * 100',
      http_error_rate: 'sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))',
      disk_free_percent: '(node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100',
    };
    
    return queries[metric] || `${metric}{job="prometheus"}`;
  }
}

interface AlertRule {
  id: string;
  name: string;
  enabled: boolean;
  expression: string;
  duration: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
}