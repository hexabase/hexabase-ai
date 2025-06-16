import { BaseGenerator } from './base-generator';

export interface Deployment {
  id: string;
  applicationId: string;
  version: string;
  status: 'pending' | 'in-progress' | 'rolling' | 'succeeded' | 'failed' | 'rolled-back';
  strategy: {
    type: 'rolling' | 'recreate' | 'blue-green' | 'canary';
    config?: {
      maxSurge?: string;
      maxUnavailable?: string;
      canarySteps?: Array<{
        weight: number;
        duration: string;
        analysis?: {
          metrics: string[];
          threshold: number;
        };
      }>;
      blueGreenConfig?: {
        activeColor: 'blue' | 'green';
        prePromotionTests?: string[];
        autoPromotionEnabled?: boolean;
        scaleDownDelaySeconds?: number;
      };
    };
  };
  replicas: {
    desired: number;
    current: number;
    updated: number;
    available: number;
    ready: number;
  };
  image: {
    repository: string;
    tag: string;
    digest?: string;
    pullPolicy: 'Always' | 'IfNotPresent' | 'Never';
  };
  changes?: Array<{
    type: 'image' | 'config' | 'scale' | 'resources' | 'env';
    description: string;
    diff?: {
      old: any;
      new: any;
    };
  }>;
  rolloutHistory?: Array<{
    revision: number;
    deployedAt: Date;
    deployedBy: string;
    image: string;
    status: 'active' | 'superseded' | 'failed';
  }>;
  metrics?: {
    startTime: Date;
    duration?: number;
    podsReady: number;
    errorRate?: number;
    latency?: {
      p50: number;
      p95: number;
      p99: number;
    };
  };
  approvals?: Array<{
    stage: string;
    approver: string;
    approvedAt?: Date;
    comment?: string;
    required: boolean;
    status: 'pending' | 'approved' | 'rejected';
  }>;
  hooks?: {
    preDeployment?: Array<{ name: string; status: 'pending' | 'running' | 'succeeded' | 'failed' }>;
    postDeployment?: Array<{ name: string; status: 'pending' | 'running' | 'succeeded' | 'failed' }>;
    preRollback?: Array<{ name: string; status: 'pending' | 'running' | 'succeeded' | 'failed' }>;
  };
  createdAt: Date;
  updatedAt: Date;
  completedAt?: Date;
  createdBy: {
    id: string;
    email: string;
    type: 'user' | 'ci-pipeline' | 'auto-scaler' | 'gitops';
  };
}

export class DeploymentGenerator extends BaseGenerator<Deployment> {
  private imageRepositories = [
    'nginx',
    'node',
    'python',
    'golang',
    'redis',
    'postgres',
    'mysql',
    'rabbitmq',
    'registry.company.com/app',
    'gcr.io/project/service',
    'docker.io/company/microservice',
  ];
  
  generate(overrides?: Partial<Deployment>): Deployment {
    const repository = overrides?.image?.repository || this.faker.helpers.arrayElement(this.imageRepositories);
    const strategyType = overrides?.strategy?.type || this.faker.helpers.arrayElement(['rolling', 'recreate', 'blue-green', 'canary']);
    const status = overrides?.status || this.faker.helpers.arrayElement(['succeeded', 'in-progress', 'rolling']);
    const desiredReplicas = overrides?.replicas?.desired || this.faker.number.int({ min: 1, max: 10 });
    
    const deployment: Deployment = {
      id: this.generateId('deploy'),
      applicationId: overrides?.applicationId || this.generateId('app'),
      version: overrides?.version || `v${this.faker.number.int({ min: 1, max: 100 })}`,
      status,
      strategy: this.generateStrategy(strategyType),
      replicas: this.generateReplicaStatus(desiredReplicas, status),
      image: {
        repository,
        tag: overrides?.image?.tag || this.faker.helpers.arrayElement(['latest', '1.0.0', '2.1.3', 'stable', `build-${this.faker.number.int({ min: 100, max: 999 })}`]),
        digest: this.faker.datatype.boolean({ probability: 0.7 }) ? `sha256:${this.faker.string.hexadecimal({ length: 64, prefix: '' })}` : undefined,
        pullPolicy: 'IfNotPresent',
      },
      createdAt: this.faker.date.recent({ days: 7 }),
      updatedAt: new Date(),
      createdBy: this.generateCreator(),
      ...overrides,
    };
    
    // Add changes for non-initial deployments
    if (this.faker.datatype.boolean({ probability: 0.8 })) {
      deployment.changes = this.generateChanges();
    }
    
    // Add rollout history
    deployment.rolloutHistory = this.generateRolloutHistory(deployment.version);
    
    // Add metrics for in-progress or completed deployments
    if (status !== 'pending') {
      deployment.metrics = this.generateMetrics(deployment.createdAt, status);
    }
    
    // Add completed time for finished deployments
    if (status === 'succeeded' || status === 'failed' || status === 'rolled-back') {
      deployment.completedAt = this.faker.date.between({ 
        from: deployment.createdAt, 
        to: new Date() 
      });
    }
    
    // Add approvals for production deployments
    if (strategyType === 'blue-green' || strategyType === 'canary') {
      deployment.approvals = this.generateApprovals(status);
    }
    
    // Add hooks for complex deployments
    if (strategyType !== 'recreate') {
      deployment.hooks = this.generateHooks(status);
    }
    
    return deployment;
  }
  
  withTraits(traits: string[]): Deployment {
    const overrides: Partial<Deployment> = {};
    
    if (traits.includes('failed')) {
      overrides.status = 'failed';
      overrides.metrics = {
        startTime: this.faker.date.recent({ days: 1 }),
        duration: this.faker.number.int({ min: 60, max: 600 }),
        podsReady: 0,
        errorRate: this.faker.number.float({ min: 0.1, max: 1.0, fractionDigits: 2 }),
      };
    }
    
    if (traits.includes('canary')) {
      overrides.strategy = {
        type: 'canary',
        config: {
          canarySteps: [
            { weight: 10, duration: '5m', analysis: { metrics: ['error-rate', 'latency'], threshold: 0.01 } },
            { weight: 25, duration: '10m', analysis: { metrics: ['error-rate', 'latency'], threshold: 0.01 } },
            { weight: 50, duration: '15m', analysis: { metrics: ['error-rate', 'latency'], threshold: 0.01 } },
            { weight: 100, duration: '0m' },
          ],
        },
      };
    }
    
    if (traits.includes('blueGreen')) {
      overrides.strategy = {
        type: 'blue-green',
        config: {
          blueGreenConfig: {
            activeColor: this.faker.helpers.arrayElement(['blue', 'green']),
            prePromotionTests: ['smoke-tests', 'integration-tests'],
            autoPromotionEnabled: false,
            scaleDownDelaySeconds: 300,
          },
        },
      };
    }
    
    if (traits.includes('automated')) {
      overrides.createdBy = {
        id: 'gitops-controller',
        email: 'gitops@system',
        type: 'gitops',
      };
      overrides.approvals = [{
        stage: 'auto-approval',
        approver: 'gitops-controller',
        approvedAt: new Date(),
        comment: 'Automated deployment via GitOps',
        required: false,
        status: 'approved',
      }];
    }
    
    if (traits.includes('rollback')) {
      overrides.status = 'rolled-back';
      overrides.changes = [{
        type: 'image',
        description: 'Rollback to previous version due to increased error rate',
        diff: {
          old: 'app:v2.0.0',
          new: 'app:v1.9.5',
        },
      }];
    }
    
    return this.generate(overrides);
  }
  
  private generateStrategy(type: Deployment['strategy']['type']): Deployment['strategy'] {
    const strategy: Deployment['strategy'] = { type };
    
    switch (type) {
      case 'rolling':
        strategy.config = {
          maxSurge: this.faker.helpers.arrayElement(['25%', '50%', '1', '2']),
          maxUnavailable: this.faker.helpers.arrayElement(['25%', '0%', '1', '0']),
        };
        break;
        
      case 'canary':
        strategy.config = {
          canarySteps: [
            { weight: 20, duration: '10m' },
            { weight: 50, duration: '10m' },
            { weight: 100, duration: '0m' },
          ],
        };
        break;
        
      case 'blue-green':
        strategy.config = {
          blueGreenConfig: {
            activeColor: 'blue',
            autoPromotionEnabled: this.faker.datatype.boolean(),
            scaleDownDelaySeconds: 300,
          },
        };
        break;
    }
    
    return strategy;
  }
  
  private generateReplicaStatus(desired: number, status: string): Deployment['replicas'] {
    if (status === 'pending') {
      return { desired, current: 0, updated: 0, available: 0, ready: 0 };
    }
    
    if (status === 'failed') {
      const ready = this.faker.number.int({ min: 0, max: Math.floor(desired / 2) });
      return { desired, current: desired, updated: ready, available: ready, ready };
    }
    
    if (status === 'in-progress' || status === 'rolling') {
      const progress = this.faker.number.float({ min: 0.1, max: 0.9 });
      const ready = Math.floor(desired * progress);
      return { 
        desired, 
        current: desired, 
        updated: ready + this.faker.number.int({ min: 0, max: 2 }), 
        available: ready, 
        ready 
      };
    }
    
    // succeeded
    return { desired, current: desired, updated: desired, available: desired, ready: desired };
  }
  
  private generateChanges(): Deployment['changes'] {
    const changeTypes: Array<Deployment['changes'][0]['type']> = ['image', 'config', 'scale', 'resources', 'env'];
    const changeCount = this.faker.number.int({ min: 1, max: 3 });
    
    return this.faker.helpers.multiple(() => {
      const type = this.faker.helpers.arrayElement(changeTypes);
      
      switch (type) {
        case 'image':
          return {
            type,
            description: 'Updated application image',
            diff: {
              old: `app:v${this.faker.number.int({ min: 1, max: 99 })}`,
              new: `app:v${this.faker.number.int({ min: 100, max: 200 })}`,
            },
          };
          
        case 'config':
          return {
            type,
            description: 'Updated configuration',
            diff: {
              old: { logLevel: 'info' },
              new: { logLevel: 'debug' },
            },
          };
          
        case 'scale':
          return {
            type,
            description: 'Scaled replicas',
            diff: {
              old: this.faker.number.int({ min: 1, max: 5 }),
              new: this.faker.number.int({ min: 3, max: 10 }),
            },
          };
          
        case 'resources':
          return {
            type,
            description: 'Updated resource limits',
            diff: {
              old: { cpu: '500m', memory: '512Mi' },
              new: { cpu: '1000m', memory: '1Gi' },
            },
          };
          
        case 'env':
          return {
            type,
            description: 'Updated environment variables',
            diff: {
              old: { FEATURE_FLAG: 'false' },
              new: { FEATURE_FLAG: 'true' },
            },
          };
      }
    }, { count: changeCount });
  }
  
  private generateRolloutHistory(currentVersion: string): Deployment['rolloutHistory'] {
    const historyCount = this.faker.number.int({ min: 0, max: 5 });
    const history = [];
    
    for (let i = 0; i < historyCount; i++) {
      const revision = parseInt(currentVersion.replace('v', '')) - i - 1;
      history.push({
        revision,
        deployedAt: this.faker.date.past({ years: 0.1 }),
        deployedBy: this.faker.internet.email(),
        image: `app:v${revision}`,
        status: i === 0 ? 'superseded' : this.faker.helpers.arrayElement(['superseded', 'failed']),
      });
    }
    
    return history;
  }
  
  private generateMetrics(startTime: Date, status: string): Deployment['metrics'] {
    const metrics: Deployment['metrics'] = {
      startTime,
      podsReady: status === 'succeeded' ? this.faker.number.int({ min: 1, max: 10 }) : 0,
    };
    
    if (status !== 'pending') {
      metrics.duration = this.faker.number.int({ min: 30, max: 600 });
      metrics.errorRate = status === 'failed' ? 
        this.faker.number.float({ min: 0.05, max: 0.5, fractionDigits: 2 }) :
        this.faker.number.float({ min: 0, max: 0.01, fractionDigits: 3 });
      
      metrics.latency = {
        p50: this.faker.number.int({ min: 10, max: 100 }),
        p95: this.faker.number.int({ min: 100, max: 500 }),
        p99: this.faker.number.int({ min: 200, max: 1000 }),
      };
    }
    
    return metrics;
  }
  
  private generateApprovals(status: string): Deployment['approvals'] {
    const stages = ['dev-lead', 'qa-sign-off', 'production-approval'];
    
    return stages.map((stage, index) => {
      const isApproved = status === 'succeeded' || (status === 'in-progress' && index < 2);
      
      return {
        stage,
        approver: isApproved ? this.faker.internet.email() : '',
        approvedAt: isApproved ? this.faker.date.recent({ days: 1 }) : undefined,
        comment: isApproved ? this.faker.lorem.sentence() : undefined,
        required: true,
        status: isApproved ? 'approved' : 'pending',
      };
    });
  }
  
  private generateHooks(status: string): Deployment['hooks'] {
    const hookStatus = status === 'succeeded' ? 'succeeded' : 
      status === 'failed' ? 'failed' : 
      this.faker.helpers.arrayElement(['running', 'succeeded']);
    
    return {
      preDeployment: [
        { name: 'database-migration', status: hookStatus },
        { name: 'cache-warm-up', status: hookStatus },
      ],
      postDeployment: [
        { name: 'smoke-tests', status: status === 'succeeded' ? 'succeeded' : 'running' },
        { name: 'monitoring-setup', status: status === 'succeeded' ? 'succeeded' : 'pending' },
      ],
    };
  }
  
  private generateCreator(): Deployment['createdBy'] {
    const type = this.faker.helpers.arrayElement(['user', 'ci-pipeline', 'auto-scaler', 'gitops']);
    
    switch (type) {
      case 'user':
        return {
          id: this.generateId('user'),
          email: this.faker.internet.email(),
          type,
        };
        
      case 'ci-pipeline':
        return {
          id: `github-actions-${this.faker.number.int({ min: 1000, max: 9999 })}`,
          email: 'ci@github.com',
          type,
        };
        
      case 'auto-scaler':
        return {
          id: 'hpa-controller',
          email: 'autoscaler@system',
          type,
        };
        
      case 'gitops':
        return {
          id: 'argocd',
          email: 'gitops@system',
          type,
        };
    }
  }
}