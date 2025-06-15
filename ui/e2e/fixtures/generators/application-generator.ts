import { BaseGenerator, Builder } from './base-generator';

export interface Application {
  id: string;
  projectId: string;
  name: string;
  type: 'stateless' | 'stateful' | 'cronjob' | 'function';
  status: 'deploying' | 'running' | 'error' | 'stopped' | 'updating';
  image: string;
  tag: string;
  replicas: number;
  port?: number;
  resources: {
    requests: {
      cpu: string;
      memory: string;
    };
    limits: {
      cpu: string;
      memory: string;
    };
  };
  env?: Record<string, string>;
  secrets?: string[];
  configMaps?: string[];
  volumes?: Array<{
    name: string;
    type: 'persistent' | 'configmap' | 'secret' | 'emptyDir';
    mountPath: string;
    size?: string;
  }>;
  networking?: {
    service: boolean;
    ingress: boolean;
    domain?: string;
    tls?: boolean;
    loadBalancer?: boolean;
  };
  healthCheck?: {
    type: 'http' | 'tcp' | 'exec';
    path?: string;
    port?: number;
    command?: string[];
    initialDelaySeconds: number;
    periodSeconds: number;
  };
  autoscaling?: {
    enabled: boolean;
    minReplicas: number;
    maxReplicas: number;
    targetCPU: number;
    targetMemory?: number;
  };
  deploymentStrategy?: {
    type: 'rolling' | 'recreate' | 'blue-green' | 'canary';
    maxSurge?: string;
    maxUnavailable?: string;
    canaryWeight?: number;
  };
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  createdAt: Date;
  updatedAt: Date;
  lastDeployedAt?: Date;
  version: string;
  previousVersions?: string[];
}

export class ApplicationGenerator extends BaseGenerator<Application> {
  private commonImages = [
    { name: 'nginx', defaultPort: 80 },
    { name: 'node', defaultPort: 3000 },
    { name: 'python', defaultPort: 8000 },
    { name: 'golang', defaultPort: 8080 },
    { name: 'postgres', defaultPort: 5432 },
    { name: 'mysql', defaultPort: 3306 },
    { name: 'redis', defaultPort: 6379 },
    { name: 'rabbitmq', defaultPort: 5672 },
    { name: 'elasticsearch', defaultPort: 9200 },
    { name: 'kibana', defaultPort: 5601 },
  ];
  
  generate(overrides?: Partial<Application>): Application {
    const name = overrides?.name || `${this.faker.hacker.noun()}-${this.faker.helpers.arrayElement(['api', 'service', 'app', 'worker'])}`;
    const type = overrides?.type || this.faker.helpers.arrayElement(['stateless', 'stateful']);
    const imageInfo = this.faker.helpers.arrayElement(this.commonImages);
    const image = overrides?.image || imageInfo.name;
    const tag = overrides?.tag || this.faker.helpers.arrayElement(['latest', '1.0.0', '2.1.3', 'stable', 'v3.2.1']);
    
    const app: Application = {
      id: this.generateId('app'),
      projectId: overrides?.projectId || this.generateId('proj'),
      name,
      type,
      status: 'running',
      image,
      tag,
      replicas: type === 'stateful' ? 1 : this.faker.number.int({ min: 1, max: 5 }),
      port: overrides?.port || imageInfo.defaultPort,
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
      env: this.generateEnvironmentVariables(image),
      secrets: this.faker.helpers.multiple(() => `${name}-secret-${this.faker.string.alphanumeric(6)}`, { count: { min: 0, max: 3 } }),
      configMaps: this.faker.helpers.multiple(() => `${name}-config-${this.faker.string.alphanumeric(6)}`, { count: { min: 0, max: 2 } }),
      networking: {
        service: true,
        ingress: type === 'stateless' && this.faker.datatype.boolean({ probability: 0.7 }),
        domain: type === 'stateless' && this.faker.datatype.boolean({ probability: 0.7 }) ? 
          `${name}.${this.faker.internet.domainName()}` : undefined,
        tls: true,
        loadBalancer: type === 'stateless' && this.faker.datatype.boolean({ probability: 0.3 }),
      },
      healthCheck: {
        type: imageInfo.defaultPort ? 'http' : 'tcp',
        path: imageInfo.defaultPort === 80 ? '/health' : undefined,
        port: imageInfo.defaultPort,
        initialDelaySeconds: 30,
        periodSeconds: 10,
      },
      labels: {
        app: name,
        version: tag,
        tier: this.faker.helpers.arrayElement(['frontend', 'backend', 'database', 'cache']),
      },
      annotations: {
        'deployment.kubernetes.io/revision': '1',
      },
      createdAt: this.faker.date.past({ years: 1 }),
      updatedAt: this.faker.date.recent({ days: 7 }),
      lastDeployedAt: this.faker.date.recent({ days: 30 }),
      version: `v${this.faker.number.int({ min: 1, max: 20 })}`,
      previousVersions: this.faker.helpers.multiple(
        () => `v${this.faker.number.int({ min: 1, max: 19 })}`, 
        { count: { min: 0, max: 5 } }
      ),
      ...overrides,
    };
    
    // Add volumes for stateful apps
    if (type === 'stateful') {
      app.volumes = [{
        name: `${name}-data`,
        type: 'persistent',
        mountPath: this.getDefaultMountPath(image),
        size: this.faker.helpers.arrayElement(['10Gi', '20Gi', '50Gi', '100Gi']),
      }];
    }
    
    // Add autoscaling for stateless apps
    if (type === 'stateless' && app.replicas > 1) {
      app.autoscaling = {
        enabled: this.faker.datatype.boolean({ probability: 0.6 }),
        minReplicas: 1,
        maxReplicas: app.replicas * 2,
        targetCPU: 70,
        targetMemory: 80,
      };
    }
    
    return app;
  }
  
  withTraits(traits: string[]): Application {
    const overrides: Partial<Application> = {};
    
    if (traits.includes('database')) {
      const dbType = this.faker.helpers.arrayElement(['postgres', 'mysql', 'mongodb']);
      overrides.type = 'stateful';
      overrides.image = dbType;
      overrides.tag = 'latest';
      overrides.replicas = 1;
      overrides.volumes = [{
        name: `${dbType}-data`,
        type: 'persistent',
        mountPath: this.getDefaultMountPath(dbType),
        size: '50Gi',
      }];
    }
    
    if (traits.includes('highAvailability')) {
      overrides.replicas = this.faker.number.int({ min: 3, max: 10 });
      overrides.autoscaling = {
        enabled: true,
        minReplicas: 3,
        maxReplicas: 10,
        targetCPU: 60,
        targetMemory: 70,
      };
      overrides.resources = {
        requests: { cpu: '500m', memory: '1Gi' },
        limits: { cpu: '2', memory: '4Gi' },
      };
    }
    
    if (traits.includes('microservice')) {
      overrides.type = 'stateless';
      overrides.resources = {
        requests: { cpu: '100m', memory: '128Mi' },
        limits: { cpu: '500m', memory: '512Mi' },
      };
      overrides.networking = {
        service: true,
        ingress: false,
        loadBalancer: false,
      };
    }
    
    if (traits.includes('publicFacing')) {
      overrides.networking = {
        service: true,
        ingress: true,
        domain: `${overrides.name || 'app'}.example.com`,
        tls: true,
        loadBalancer: true,
      };
    }
    
    if (traits.includes('canary')) {
      overrides.deploymentStrategy = {
        type: 'canary',
        canaryWeight: 10,
      };
    }
    
    return this.generate(overrides);
  }
  
  private generateEnvironmentVariables(image: string): Record<string, string> {
    const common = {
      NODE_ENV: 'production',
      LOG_LEVEL: this.faker.helpers.arrayElement(['debug', 'info', 'warn', 'error']),
    };
    
    const imageSpecific: Record<string, Record<string, string>> = {
      postgres: {
        POSTGRES_DB: this.faker.database.collation(),
        POSTGRES_USER: this.faker.internet.userName(),
      },
      mysql: {
        MYSQL_DATABASE: this.faker.database.collation(),
        MYSQL_USER: this.faker.internet.userName(),
      },
      redis: {
        REDIS_MAXMEMORY: '256mb',
        REDIS_MAXMEMORY_POLICY: 'allkeys-lru',
      },
      rabbitmq: {
        RABBITMQ_DEFAULT_USER: this.faker.internet.userName(),
        RABBITMQ_DEFAULT_VHOST: '/',
      },
    };
    
    return { ...common, ...(imageSpecific[image] || {}) };
  }
  
  private getDefaultMountPath(image: string): string {
    const mountPaths: Record<string, string> = {
      postgres: '/var/lib/postgresql/data',
      mysql: '/var/lib/mysql',
      mongodb: '/data/db',
      redis: '/data',
      elasticsearch: '/usr/share/elasticsearch/data',
    };
    
    return mountPaths[image] || '/data';
  }
}

export class ApplicationBuilder extends Builder<Application> {
  private generator = new ApplicationGenerator();
  
  withProject(projectId: string): this {
    return this.set('projectId', projectId);
  }
  
  withName(name: string): this {
    return this.set('name', name);
  }
  
  withImage(image: string, tag: string = 'latest'): this {
    return this.set('image', image).set('tag', tag);
  }
  
  withType(type: Application['type']): this {
    return this.set('type', type);
  }
  
  withReplicas(count: number): this {
    return this.set('replicas', count);
  }
  
  withPort(port: number): this {
    return this.set('port', port);
  }
  
  withResources(requests: { cpu: string; memory: string }, limits: { cpu: string; memory: string }): this {
    return this.set('resources', { requests, limits });
  }
  
  withEnvironment(env: Record<string, string>): this {
    return this.set('env', { ...this.data.env, ...env });
  }
  
  withVolume(name: string, type: 'persistent' | 'configmap' | 'secret', mountPath: string, size?: string): this {
    const volumes = this.data.volumes || [];
    volumes.push({ name, type, mountPath, size });
    return this.set('volumes', volumes);
  }
  
  withIngress(domain: string, tls: boolean = true): this {
    const networking = this.data.networking || { service: true, ingress: false };
    return this.set('networking', {
      ...networking,
      ingress: true,
      domain,
      tls,
    });
  }
  
  withAutoscaling(minReplicas: number, maxReplicas: number, targetCPU: number): this {
    return this.set('autoscaling', {
      enabled: true,
      minReplicas,
      maxReplicas,
      targetCPU,
    });
  }
  
  withDeploymentStrategy(strategy: Application['deploymentStrategy']): this {
    return this.set('deploymentStrategy', strategy);
  }
  
  build(): Application {
    return this.generator.generate(this.data);
  }
}