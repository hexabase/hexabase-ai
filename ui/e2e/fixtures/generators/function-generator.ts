import { BaseGenerator } from './base-generator';

export interface ServerlessFunction {
  id: string;
  projectId: string;
  name: string;
  runtime: 'nodejs16' | 'nodejs18' | 'python39' | 'python311' | 'go119' | 'java11' | 'dotnet6';
  handler: string;
  code?: {
    source: string;
    size: number;
    checksum: string;
  };
  packageUrl?: string;
  environment?: Record<string, string>;
  timeout: number; // seconds
  memorySize: number; // MB
  triggers: Array<{
    type: 'http' | 'schedule' | 'event' | 'queue';
    config: {
      // HTTP trigger
      method?: string;
      path?: string;
      auth?: 'none' | 'api-key' | 'jwt';
      
      // Schedule trigger  
      schedule?: string;
      timezone?: string;
      
      // Event trigger
      eventType?: string;
      filter?: Record<string, any>;
      
      // Queue trigger
      queueName?: string;
      batchSize?: number;
    };
  }>;
  versions: Array<{
    version: string;
    createdAt: Date;
    createdBy: string;
    description?: string;
    active: boolean;
  }>;
  currentVersion: string;
  concurrency: {
    reserved?: number;
    max: number;
  };
  retries?: {
    attempts: number;
    backoff: 'exponential' | 'linear';
    maxDelay: number;
  };
  deadLetterQueue?: {
    enabled: boolean;
    maxRetries: number;
    queueName: string;
  };
  monitoring?: {
    tracingEnabled: boolean;
    logLevel: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
    customMetrics?: string[];
  };
  labels?: Record<string, string>;
  status: 'deploying' | 'active' | 'error' | 'updating';
  lastInvocation?: {
    timestamp: Date;
    duration: number;
    status: 'success' | 'error' | 'timeout';
    error?: string;
  };
  statistics?: {
    invocations: {
      total: number;
      successful: number;
      failed: number;
      throttled: number;
    };
    performance: {
      averageDuration: number;
      minDuration: number;
      maxDuration: number;
      p95Duration: number;
      p99Duration: number;
    };
    cost: {
      estimatedMonthly: number;
      lastMonth: number;
    };
  };
  createdAt: Date;
  updatedAt: Date;
}

export class FunctionGenerator extends BaseGenerator<ServerlessFunction> {
  private functionTemplates = [
    {
      name: 'api-handler',
      runtime: 'nodejs18' as const,
      handler: 'index.handler',
      trigger: { type: 'http' as const, method: 'POST', path: '/api/process' },
      code: `exports.handler = async (event) => {
  const body = JSON.parse(event.body);
  // Process request
  return {
    statusCode: 200,
    body: JSON.stringify({ success: true, id: Date.now() })
  };
};`,
    },
    {
      name: 'image-processor',
      runtime: 'python311' as const,
      handler: 'main.process_image',
      trigger: { type: 'event' as const, eventType: 'storage.object.created' },
      code: `import json
def process_image(event, context):
    # Process uploaded image
    return {"statusCode": 200, "body": json.dumps({"processed": True})}`,
    },
    {
      name: 'data-aggregator',
      runtime: 'go119' as const,
      handler: 'main',
      trigger: { type: 'schedule' as const, schedule: '0 */6 * * *' },
      code: `package main
import "fmt"
func main() {
    fmt.Println("Aggregating data...")
}`,
    },
    {
      name: 'notification-sender',
      runtime: 'nodejs18' as const,
      handler: 'notifications.send',
      trigger: { type: 'queue' as const, queueName: 'notifications', batchSize: 10 },
      code: `exports.send = async (messages) => {
  for (const msg of messages) {
    await sendEmail(msg);
  }
};`,
    },
  ];
  
  generate(overrides?: Partial<ServerlessFunction>): ServerlessFunction {
    const template = this.faker.helpers.arrayElement(this.functionTemplates);
    const name = overrides?.name || `${template.name}-${this.faker.word.adjective()}`;
    
    const func: ServerlessFunction = {
      id: this.generateId('func'),
      projectId: overrides?.projectId || this.generateId('proj'),
      name,
      runtime: overrides?.runtime || template.runtime,
      handler: overrides?.handler || template.handler,
      code: {
        source: template.code,
        size: Buffer.from(template.code).length,
        checksum: this.faker.git.commitSha().substring(0, 16),
      },
      environment: this.generateEnvironment(template.name),
      timeout: this.faker.helpers.arrayElement([30, 60, 180, 300]),
      memorySize: this.faker.helpers.arrayElement([128, 256, 512, 1024]),
      triggers: [this.generateTrigger(template.trigger)],
      versions: [],
      currentVersion: 'v1',
      concurrency: {
        max: this.faker.number.int({ min: 100, max: 1000 }),
        reserved: this.faker.datatype.boolean({ probability: 0.3 }) ? 
          this.faker.number.int({ min: 1, max: 10 }) : undefined,
      },
      status: 'active',
      labels: {
        team: this.faker.helpers.arrayElement(['backend', 'frontend', 'data']),
        environment: this.faker.helpers.arrayElement(['dev', 'staging', 'prod']),
      },
      createdAt: this.faker.date.past({ years: 1 }),
      updatedAt: this.faker.date.recent({ days: 7 }),
      ...overrides,
    };
    
    // Generate version history
    func.versions = this.generateVersionHistory(func.currentVersion, func.createdAt);
    
    // Add retry configuration for non-HTTP triggers
    if (func.triggers[0].type !== 'http') {
      func.retries = {
        attempts: 3,
        backoff: 'exponential',
        maxDelay: 60,
      };
    }
    
    // Add dead letter queue for queue triggers
    if (func.triggers[0].type === 'queue') {
      func.deadLetterQueue = {
        enabled: true,
        maxRetries: 5,
        queueName: `${func.name}-dlq`,
      };
    }
    
    // Add monitoring
    func.monitoring = {
      tracingEnabled: this.faker.datatype.boolean({ probability: 0.7 }),
      logLevel: this.faker.helpers.arrayElement(['INFO', 'DEBUG', 'WARN']),
      customMetrics: this.faker.datatype.boolean({ probability: 0.5 }) ?
        ['custom.processing_time', 'custom.items_processed'] : undefined,
    };
    
    // Generate invocation statistics
    if (func.status === 'active') {
      func.lastInvocation = this.generateLastInvocation();
      func.statistics = this.generateStatistics();
    }
    
    return func;
  }
  
  withTraits(traits: string[]): ServerlessFunction {
    const overrides: Partial<ServerlessFunction> = {};
    
    if (traits.includes('highTraffic')) {
      overrides.concurrency = {
        max: 5000,
        reserved: 50,
      };
      overrides.statistics = {
        invocations: {
          total: this.faker.number.int({ min: 100000, max: 1000000 }),
          successful: this.faker.number.int({ min: 95000, max: 995000 }),
          failed: this.faker.number.int({ min: 100, max: 5000 }),
          throttled: this.faker.number.int({ min: 0, max: 1000 }),
        },
        performance: {
          averageDuration: 45,
          minDuration: 10,
          maxDuration: 2000,
          p95Duration: 100,
          p99Duration: 500,
        },
        cost: {
          estimatedMonthly: this.faker.number.float({ min: 100, max: 1000, fractionDigits: 2 }),
          lastMonth: this.faker.number.float({ min: 80, max: 900, fractionDigits: 2 }),
        },
      };
    }
    
    if (traits.includes('errorProne')) {
      overrides.status = 'error';
      overrides.lastInvocation = {
        timestamp: this.faker.date.recent({ days: 1 }),
        duration: this.faker.number.int({ min: 100, max: 30000 }),
        status: 'error',
        error: this.faker.helpers.arrayElement([
          'TypeError: Cannot read property of undefined',
          'TimeoutError: Function timed out after 300 seconds',
          'MemoryError: Function ran out of memory',
          'PermissionError: Access denied to resource',
        ]),
      };
    }
    
    if (traits.includes('multiTrigger')) {
      overrides.triggers = [
        { type: 'http', config: { method: 'GET', path: '/api/status', auth: 'none' } },
        { type: 'http', config: { method: 'POST', path: '/api/process', auth: 'api-key' } },
        { type: 'schedule', config: { schedule: '0 0 * * *', timezone: 'UTC' } },
      ];
    }
    
    if (traits.includes('dataProcessing')) {
      overrides.runtime = 'python311';
      overrides.handler = 'processor.main';
      overrides.memorySize = 3008; // Maximum Lambda memory
      overrides.timeout = 900; // 15 minutes
      overrides.environment = {
        ...overrides.environment,
        BATCH_SIZE: '1000',
        PARALLEL_WORKERS: '10',
        OUTPUT_FORMAT: 'parquet',
      };
    }
    
    return this.generate(overrides);
  }
  
  private generateTrigger(template: any): ServerlessFunction['triggers'][0] {
    const trigger: ServerlessFunction['triggers'][0] = {
      type: template.type,
      config: {},
    };
    
    switch (template.type) {
      case 'http':
        trigger.config = {
          method: template.method || 'GET',
          path: template.path || '/api/function',
          auth: this.faker.helpers.arrayElement(['none', 'api-key', 'jwt']),
        };
        break;
        
      case 'schedule':
        trigger.config = {
          schedule: template.schedule || '0 0 * * *',
          timezone: 'UTC',
        };
        break;
        
      case 'event':
        trigger.config = {
          eventType: template.eventType || 'custom.event',
          filter: {
            source: 'application',
            type: this.faker.helpers.arrayElement(['created', 'updated', 'deleted']),
          },
        };
        break;
        
      case 'queue':
        trigger.config = {
          queueName: template.queueName || 'default-queue',
          batchSize: template.batchSize || 10,
        };
        break;
    }
    
    return trigger;
  }
  
  private generateEnvironment(functionType: string): Record<string, string> {
    const common = {
      NODE_ENV: 'production',
      LOG_LEVEL: 'INFO',
      REGION: this.faker.helpers.arrayElement(['us-east-1', 'eu-west-1', 'ap-southeast-1']),
    };
    
    const specific: Record<string, Record<string, string>> = {
      'api-handler': {
        API_KEY: this.faker.string.alphanumeric(32),
        DATABASE_URL: 'postgresql://user:pass@db:5432/app',
        CACHE_TTL: '3600',
      },
      'image-processor': {
        S3_BUCKET: 'processed-images',
        MAX_IMAGE_SIZE: '10485760', // 10MB
        SUPPORTED_FORMATS: 'jpg,png,webp',
      },
      'data-aggregator': {
        AGGREGATION_INTERVAL: '6h',
        OUTPUT_TABLE: 'aggregated_metrics',
        RETENTION_DAYS: '90',
      },
      'notification-sender': {
        SMTP_HOST: 'smtp.example.com',
        FROM_EMAIL: 'noreply@example.com',
        TEMPLATE_BUCKET: 'email-templates',
      },
    };
    
    return { ...common, ...(specific[functionType] || {}) };
  }
  
  private generateVersionHistory(currentVersion: string, since: Date): ServerlessFunction['versions'] {
    const versions = [];
    const versionCount = this.faker.number.int({ min: 1, max: 10 });
    const currentVersionNum = parseInt(currentVersion.replace('v', ''));
    
    for (let i = 0; i < versionCount; i++) {
      const versionNum = currentVersionNum - i;
      if (versionNum < 1) break;
      
      versions.push({
        version: `v${versionNum}`,
        createdAt: this.faker.date.between({ from: since, to: new Date() }),
        createdBy: this.faker.internet.email(),
        description: i === 0 ? 'Current version' : 
          this.faker.helpers.arrayElement([
            'Fixed memory leak issue',
            'Added new API endpoint',
            'Performance improvements',
            'Updated dependencies',
            'Bug fixes and improvements',
          ]),
        active: i === 0,
      });
    }
    
    return versions.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
  }
  
  private generateLastInvocation(): ServerlessFunction['lastInvocation'] {
    const status = this.faker.helpers.weighted(
      ['success', 'error', 'timeout'],
      [0.9, 0.08, 0.02]
    );
    
    return {
      timestamp: this.faker.date.recent({ days: 1 }),
      duration: status === 'timeout' ? 300000 : 
        this.faker.number.int({ min: 10, max: status === 'error' ? 5000 : 1000 }),
      status,
      error: status === 'error' ? this.faker.helpers.arrayElement([
        'Connection refused',
        'Invalid input format',
        'Rate limit exceeded',
        'Internal server error',
      ]) : undefined,
    };
  }
  
  private generateStatistics(): ServerlessFunction['statistics'] {
    const total = this.faker.number.int({ min: 1000, max: 100000 });
    const failureRate = this.faker.number.float({ min: 0.001, max: 0.05 });
    const failed = Math.floor(total * failureRate);
    const throttled = Math.floor(total * 0.001);
    
    return {
      invocations: {
        total,
        successful: total - failed - throttled,
        failed,
        throttled,
      },
      performance: {
        averageDuration: this.faker.number.int({ min: 50, max: 500 }),
        minDuration: this.faker.number.int({ min: 10, max: 50 }),
        maxDuration: this.faker.number.int({ min: 1000, max: 5000 }),
        p95Duration: this.faker.number.int({ min: 200, max: 1000 }),
        p99Duration: this.faker.number.int({ min: 500, max: 2000 }),
      },
      cost: {
        estimatedMonthly: this.faker.number.float({ min: 10, max: 500, fractionDigits: 2 }),
        lastMonth: this.faker.number.float({ min: 8, max: 480, fractionDigits: 2 }),
      },
    };
  }
}