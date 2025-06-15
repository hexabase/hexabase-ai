import { BaseGenerator } from './base-generator';

export interface MetricPoint {
  timestamp: Date;
  value: number;
  labels?: Record<string, string>;
}

export interface MetricSeries {
  name: string;
  type: 'gauge' | 'counter' | 'histogram' | 'summary';
  unit?: string;
  description?: string;
  points: MetricPoint[];
  aggregations?: {
    min: number;
    max: number;
    avg: number;
    sum: number;
    count: number;
    p50?: number;
    p95?: number;
    p99?: number;
  };
}

export interface ResourceMetrics {
  cpu: {
    usage: MetricSeries;
    throttled?: MetricSeries;
    requests: number;
    limits: number;
  };
  memory: {
    usage: MetricSeries;
    workingSet?: MetricSeries;
    requests: number;
    limits: number;
  };
  disk?: {
    usage: MetricSeries;
    iops?: MetricSeries;
    throughput?: MetricSeries;
  };
  network?: {
    rxBytes: MetricSeries;
    txBytes: MetricSeries;
    rxPackets?: MetricSeries;
    txPackets?: MetricSeries;
    errors?: MetricSeries;
  };
}

export interface ApplicationMetrics {
  http?: {
    requestRate: MetricSeries;
    errorRate: MetricSeries;
    latency: MetricSeries;
    activeConnections?: MetricSeries;
  };
  business?: {
    [key: string]: MetricSeries;
  };
  custom?: {
    [key: string]: MetricSeries;
  };
}

export class MetricsGenerator extends BaseGenerator<MetricSeries> {
  generate(overrides?: Partial<MetricSeries>): MetricSeries {
    const name = overrides?.name || this.faker.helpers.arrayElement([
      'cpu_usage_percent',
      'memory_usage_bytes',
      'http_requests_total',
      'http_request_duration_seconds',
      'error_rate_percent',
    ]);
    
    const type = overrides?.type || this.inferMetricType(name);
    const unit = overrides?.unit || this.inferUnit(name);
    
    const series: MetricSeries = {
      name,
      type,
      unit,
      description: overrides?.description || this.generateDescription(name),
      points: overrides?.points || this.generateTimeSeries({
        duration: 3600, // 1 hour
        interval: 60,   // 1 minute
        pattern: this.inferPattern(name),
      }),
      ...overrides,
    };
    
    // Calculate aggregations
    series.aggregations = this.calculateAggregations(series.points, type);
    
    return series;
  }
  
  withTraits(traits: string[]): MetricSeries {
    const overrides: Partial<MetricSeries> = {};
    
    if (traits.includes('spike')) {
      overrides.points = this.generateTimeSeries({
        duration: 3600,
        interval: 60,
        pattern: 'spike',
        spikeAt: 0.5,
        spikeMultiplier: 5,
      });
    }
    
    if (traits.includes('gradualIncrease')) {
      overrides.points = this.generateTimeSeries({
        duration: 3600,
        interval: 60,
        pattern: 'linear',
        slope: 0.5,
      });
    }
    
    if (traits.includes('error')) {
      overrides.name = 'error_rate_percent';
      overrides.points = this.generateTimeSeries({
        duration: 3600,
        interval: 60,
        pattern: 'random',
        min: 0,
        max: 15,
        noise: 0.3,
      });
    }
    
    return this.generate(overrides);
  }
  
  generateResourceMetrics(config?: {
    duration?: number;
    interval?: number;
    resourceType?: 'pod' | 'node' | 'container';
    workloadType?: 'web' | 'batch' | 'database' | 'cache';
  }): ResourceMetrics {
    const duration = config?.duration || 3600;
    const interval = config?.interval || 60;
    const workloadType = config?.workloadType || 'web';
    
    const cpuPattern = this.getWorkloadPattern(workloadType, 'cpu');
    const memoryPattern = this.getWorkloadPattern(workloadType, 'memory');
    
    return {
      cpu: {
        usage: {
          name: 'cpu_usage_percent',
          type: 'gauge',
          unit: '%',
          description: 'CPU usage percentage',
          points: this.generateTimeSeries({
            duration,
            interval,
            pattern: cpuPattern.pattern,
            min: cpuPattern.min,
            max: cpuPattern.max,
            noise: 0.1,
          }),
          aggregations: undefined as any, // Will be calculated
        },
        requests: this.faker.helpers.arrayElement([0.1, 0.25, 0.5, 1, 2]),
        limits: this.faker.helpers.arrayElement([0.5, 1, 2, 4]),
      },
      memory: {
        usage: {
          name: 'memory_usage_bytes',
          type: 'gauge',
          unit: 'bytes',
          description: 'Memory usage in bytes',
          points: this.generateTimeSeries({
            duration,
            interval,
            pattern: memoryPattern.pattern,
            min: memoryPattern.min * 1024 * 1024 * 1024, // GB to bytes
            max: memoryPattern.max * 1024 * 1024 * 1024,
            noise: 0.05,
          }),
          aggregations: undefined as any,
        },
        requests: this.faker.helpers.arrayElement([128, 256, 512, 1024]) * 1024 * 1024,
        limits: this.faker.helpers.arrayElement([256, 512, 1024, 2048]) * 1024 * 1024,
      },
      disk: this.faker.datatype.boolean({ probability: 0.7 }) ? {
        usage: {
          name: 'disk_usage_bytes',
          type: 'gauge',
          unit: 'bytes',
          description: 'Disk usage in bytes',
          points: this.generateTimeSeries({
            duration,
            interval,
            pattern: 'linear',
            min: 1 * 1024 * 1024 * 1024,
            max: 10 * 1024 * 1024 * 1024,
            slope: 0.1,
          }),
          aggregations: undefined as any,
        },
      } : undefined,
      network: {
        rxBytes: {
          name: 'network_rx_bytes',
          type: 'counter',
          unit: 'bytes',
          description: 'Network received bytes',
          points: this.generateTimeSeries({
            duration,
            interval,
            pattern: 'sawtooth',
            min: 0,
            max: 100 * 1024 * 1024, // 100MB
          }),
          aggregations: undefined as any,
        },
        txBytes: {
          name: 'network_tx_bytes',
          type: 'counter',
          unit: 'bytes',
          description: 'Network transmitted bytes',
          points: this.generateTimeSeries({
            duration,
            interval,
            pattern: 'sawtooth',
            min: 0,
            max: 50 * 1024 * 1024, // 50MB
          }),
          aggregations: undefined as any,
        },
      },
    };
  }
  
  generateApplicationMetrics(config?: {
    duration?: number;
    interval?: number;
    appType?: 'api' | 'web' | 'worker' | 'streaming';
    load?: 'low' | 'medium' | 'high';
  }): ApplicationMetrics {
    const duration = config?.duration || 3600;
    const interval = config?.interval || 60;
    const appType = config?.appType || 'api';
    const load = config?.load || 'medium';
    
    const loadMultiplier = { low: 0.3, medium: 1, high: 3 }[load];
    
    const metrics: ApplicationMetrics = {};
    
    if (appType === 'api' || appType === 'web') {
      metrics.http = {
        requestRate: {
          name: 'http_requests_per_second',
          type: 'gauge',
          unit: 'req/s',
          description: 'HTTP request rate',
          points: this.generateTimeSeries({
            duration,
            interval,
            pattern: 'sinusoidal',
            min: 10 * loadMultiplier,
            max: 100 * loadMultiplier,
            period: 3600, // 1 hour cycle
            noise: 0.2,
          }),
          aggregations: undefined as any,
        },
        errorRate: {
          name: 'http_error_rate_percent',
          type: 'gauge',
          unit: '%',
          description: 'HTTP error rate percentage',
          points: this.generateTimeSeries({
            duration,
            interval,
            pattern: 'random',
            min: 0,
            max: load === 'high' ? 5 : 2,
            noise: 0.5,
          }),
          aggregations: undefined as any,
        },
        latency: {
          name: 'http_request_duration_ms',
          type: 'histogram',
          unit: 'ms',
          description: 'HTTP request latency',
          points: this.generateLatencyTimeSeries({
            duration,
            interval,
            baseLatency: appType === 'api' ? 50 : 200,
            loadMultiplier,
          }),
          aggregations: undefined as any,
        },
      };
    }
    
    // Add business metrics
    if (appType === 'api' || appType === 'web') {
      metrics.business = {
        user_signups: this.generate({
          name: 'user_signups_total',
          type: 'counter',
          points: this.generateTimeSeries({
            duration,
            interval,
            pattern: 'poisson',
            rate: 5 * loadMultiplier,
          }),
        }),
        orders_processed: this.generate({
          name: 'orders_processed_total',
          type: 'counter',
          points: this.generateTimeSeries({
            duration,
            interval,
            pattern: 'poisson',
            rate: 20 * loadMultiplier,
          }),
        }),
      };
    }
    
    // Calculate all aggregations
    this.calculateAllAggregations(metrics);
    
    return metrics;
  }
  
  private generateTimeSeries(config: {
    duration: number; // seconds
    interval: number; // seconds
    pattern: 'constant' | 'linear' | 'sinusoidal' | 'sawtooth' | 'spike' | 'random' | 'poisson';
    min?: number;
    max?: number;
    noise?: number;
    slope?: number;
    period?: number;
    spikeAt?: number;
    spikeMultiplier?: number;
    rate?: number;
  }): MetricPoint[] {
    const points: MetricPoint[] = [];
    const count = Math.floor(config.duration / config.interval);
    const startTime = this.faker.date.recent({ days: 1 });
    
    const min = config.min || 0;
    const max = config.max || 100;
    const range = max - min;
    
    for (let i = 0; i < count; i++) {
      const timestamp = new Date(startTime.getTime() + i * config.interval * 1000);
      let value: number;
      
      switch (config.pattern) {
        case 'constant':
          value = (min + max) / 2;
          break;
          
        case 'linear':
          value = min + (range * i / count) * (config.slope || 1);
          break;
          
        case 'sinusoidal':
          const period = config.period || config.duration;
          value = min + range * (0.5 + 0.5 * Math.sin(2 * Math.PI * i * config.interval / period));
          break;
          
        case 'sawtooth':
          value = min + range * ((i % 10) / 10);
          break;
          
        case 'spike':
          const spikePoint = Math.floor(count * (config.spikeAt || 0.5));
          if (Math.abs(i - spikePoint) < 3) {
            value = max * (config.spikeMultiplier || 2);
          } else {
            value = min + range * 0.3;
          }
          break;
          
        case 'random':
          value = this.faker.number.float({ min, max, fractionDigits: 2 });
          break;
          
        case 'poisson':
          value = this.generatePoissonValue(config.rate || 10);
          break;
          
        default:
          value = min + range / 2;
      }
      
      // Add noise
      if (config.noise && config.pattern !== 'random') {
        const noise = (Math.random() - 0.5) * 2 * config.noise * range;
        value = Math.max(min, Math.min(max, value + noise));
      }
      
      points.push({
        timestamp,
        value: parseFloat(value.toFixed(2)),
      });
    }
    
    return points;
  }
  
  private generateLatencyTimeSeries(config: {
    duration: number;
    interval: number;
    baseLatency: number;
    loadMultiplier: number;
  }): MetricPoint[] {
    const points: MetricPoint[] = [];
    const count = Math.floor(config.duration / config.interval);
    const startTime = this.faker.date.recent({ days: 1 });
    
    for (let i = 0; i < count; i++) {
      const timestamp = new Date(startTime.getTime() + i * config.interval * 1000);
      
      // Generate realistic latency distribution
      const p50 = config.baseLatency * (1 + 0.2 * config.loadMultiplier);
      const p95 = p50 * 2.5;
      const p99 = p50 * 4;
      
      // Simulate percentile-based values
      const rand = Math.random();
      let value: number;
      
      if (rand < 0.5) {
        value = this.faker.number.float({ min: config.baseLatency * 0.5, max: p50 });
      } else if (rand < 0.95) {
        value = this.faker.number.float({ min: p50, max: p95 });
      } else {
        value = this.faker.number.float({ min: p95, max: p99 });
      }
      
      points.push({
        timestamp,
        value: parseFloat(value.toFixed(2)),
        labels: {
          percentile: rand < 0.5 ? 'p50' : rand < 0.95 ? 'p95' : 'p99',
        },
      });
    }
    
    return points;
  }
  
  private generatePoissonValue(lambda: number): number {
    let L = Math.exp(-lambda);
    let p = 1.0;
    let k = 0;
    
    do {
      k++;
      p *= Math.random();
    } while (p > L);
    
    return k - 1;
  }
  
  private calculateAggregations(points: MetricPoint[], type: MetricSeries['type']): MetricSeries['aggregations'] {
    if (points.length === 0) return undefined;
    
    const values = points.map(p => p.value);
    const sorted = [...values].sort((a, b) => a - b);
    
    const aggregations = {
      min: Math.min(...values),
      max: Math.max(...values),
      avg: values.reduce((a, b) => a + b, 0) / values.length,
      sum: values.reduce((a, b) => a + b, 0),
      count: values.length,
    };
    
    if (type === 'histogram' || type === 'summary') {
      aggregations.p50 = sorted[Math.floor(sorted.length * 0.5)];
      aggregations.p95 = sorted[Math.floor(sorted.length * 0.95)];
      aggregations.p99 = sorted[Math.floor(sorted.length * 0.99)];
    }
    
    return aggregations;
  }
  
  private calculateAllAggregations(metrics: ApplicationMetrics | ResourceMetrics) {
    const processMetric = (metric: any) => {
      if (metric && metric.points && Array.isArray(metric.points)) {
        metric.aggregations = this.calculateAggregations(metric.points, metric.type);
      }
    };
    
    Object.values(metrics).forEach(category => {
      if (category && typeof category === 'object') {
        Object.values(category).forEach(metric => {
          processMetric(metric);
        });
      }
    });
  }
  
  private inferMetricType(name: string): MetricSeries['type'] {
    if (name.includes('_total') || name.includes('_count')) return 'counter';
    if (name.includes('_duration') || name.includes('_latency')) return 'histogram';
    if (name.includes('_summary')) return 'summary';
    return 'gauge';
  }
  
  private inferUnit(name: string): string | undefined {
    if (name.includes('_bytes')) return 'bytes';
    if (name.includes('_seconds')) return 's';
    if (name.includes('_milliseconds') || name.includes('_ms')) return 'ms';
    if (name.includes('_percent')) return '%';
    if (name.includes('_ratio')) return 'ratio';
    if (name.includes('_per_second')) return '/s';
    return undefined;
  }
  
  private inferPattern(name: string): 'constant' | 'linear' | 'sinusoidal' | 'random' {
    if (name.includes('cpu') || name.includes('request')) return 'sinusoidal';
    if (name.includes('memory') || name.includes('disk')) return 'linear';
    if (name.includes('error')) return 'random';
    return 'constant';
  }
  
  private generateDescription(name: string): string {
    const descriptions: Record<string, string> = {
      cpu_usage_percent: 'CPU usage as a percentage of allocated resources',
      memory_usage_bytes: 'Memory usage in bytes',
      http_requests_total: 'Total number of HTTP requests',
      http_request_duration_seconds: 'HTTP request duration in seconds',
      error_rate_percent: 'Percentage of requests that resulted in errors',
    };
    
    return descriptions[name] || `Metric: ${name}`;
  }
  
  private getWorkloadPattern(workloadType: string, resource: 'cpu' | 'memory') {
    const patterns = {
      web: {
        cpu: { pattern: 'sinusoidal' as const, min: 10, max: 70 },
        memory: { pattern: 'constant' as const, min: 0.5, max: 1.5 },
      },
      batch: {
        cpu: { pattern: 'spike' as const, min: 5, max: 95 },
        memory: { pattern: 'sawtooth' as const, min: 0.2, max: 4 },
      },
      database: {
        cpu: { pattern: 'random' as const, min: 20, max: 60 },
        memory: { pattern: 'linear' as const, min: 2, max: 8 },
      },
      cache: {
        cpu: { pattern: 'constant' as const, min: 5, max: 15 },
        memory: { pattern: 'linear' as const, min: 1, max: 4 },
      },
    };
    
    return patterns[workloadType][resource];
  }
}