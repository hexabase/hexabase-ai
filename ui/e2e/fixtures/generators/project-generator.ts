import { BaseGenerator, Builder } from './base-generator';

export interface Project {
  id: string;
  workspaceId: string;
  name: string;
  namespace: string;
  description?: string;
  status: 'active' | 'inactive' | 'terminating';
  resources: {
    applications: number;
    cronjobs: number;
    functions: number;
    secrets: number;
    configmaps: number;
  };
  quotas: {
    maxApplications: number;
    maxCPU: string;
    maxMemory: string;
    maxStorage: string;
  };
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  createdAt: Date;
  updatedAt: Date;
  owner: {
    userId: string;
    email: string;
  };
  environment?: 'development' | 'staging' | 'production';
  gitIntegration?: {
    provider: 'github' | 'gitlab' | 'bitbucket';
    repository: string;
    branch: string;
    autoSync: boolean;
  };
}

export class ProjectGenerator extends BaseGenerator<Project> {
  generate(overrides?: Partial<Project>): Partial<Project> {
    const name = overrides?.name || `${this.faker.hacker.adjective()}-${this.faker.hacker.noun()}`;
    const namespace = overrides?.namespace || this.generateSlug(name);
    const environment = overrides?.environment || this.faker.helpers.arrayElement(['development', 'staging', 'production']);
    
    const project: Project = {
      id: this.generateId('proj'),
      workspaceId: overrides?.workspaceId || this.generateId('ws'),
      name,
      namespace,
      description: overrides?.description || this.faker.lorem.sentence(),
      status: 'active',
      resources: {
        applications: this.faker.number.int({ min: 0, max: 10 }),
        cronjobs: this.faker.number.int({ min: 0, max: 5 }),
        functions: this.faker.number.int({ min: 0, max: 8 }),
        secrets: this.faker.number.int({ min: 0, max: 20 }),
        configmaps: this.faker.number.int({ min: 0, max: 15 }),
      },
      quotas: {
        maxApplications: 20,
        maxCPU: '16',
        maxMemory: '32Gi',
        maxStorage: '100Gi',
      },
      labels: {
        environment,
        team: this.faker.helpers.arrayElement(['backend', 'frontend', 'devops', 'data']),
        'cost-center': `cc-${this.faker.number.int({ min: 1000, max: 9999 })}`,
      },
      annotations: {
        'created-by': 'e2e-test',
        'managed-by': 'hexabase-ai',
      },
      createdAt: this.faker.date.past({ years: 1 }),
      updatedAt: this.faker.date.recent({ days: 7 }),
      owner: {
        userId: this.generateId('user'),
        email: this.faker.internet.email(),
      },
      environment,
      ...overrides,
    };
    
    // Add git integration for some projects
    if (this.faker.datatype.boolean({ probability: 0.6 })) {
      project.gitIntegration = {
        provider: this.faker.helpers.arrayElement(['github', 'gitlab', 'bitbucket']),
        repository: `${this.faker.internet.userName()}/${namespace}`,
        branch: environment === 'production' ? 'main' : environment,
        autoSync: environment !== 'production',
      };
    }
    
    return project;
  }
  
  withTraits(traits: string[]): Project {
    const overrides: Partial<Project> = {};
    
    if (traits.includes('empty')) {
      overrides.resources = {
        applications: 0,
        cronjobs: 0,
        functions: 0,
        secrets: 0,
        configmaps: 0,
      };
    }
    
    if (traits.includes('busy')) {
      overrides.resources = {
        applications: this.faker.number.int({ min: 15, max: 20 }),
        cronjobs: this.faker.number.int({ min: 8, max: 15 }),
        functions: this.faker.number.int({ min: 10, max: 20 }),
        secrets: this.faker.number.int({ min: 30, max: 50 }),
        configmaps: this.faker.number.int({ min: 20, max: 40 }),
      };
    }
    
    if (traits.includes('production')) {
      overrides.environment = 'production';
      overrides.labels = {
        ...overrides.labels,
        environment: 'production',
        'high-availability': 'true',
        'backup-enabled': 'true',
      };
    }
    
    if (traits.includes('gitops')) {
      overrides.gitIntegration = {
        provider: 'github',
        repository: `${this.faker.internet.userName()}/${overrides.namespace || 'project'}`,
        branch: 'main',
        autoSync: true,
      };
      overrides.annotations = {
        ...overrides.annotations,
        'argocd.argoproj.io/sync-policy': 'automated',
      };
    }
    
    return this.generate(overrides) as Project;
  }
}

export class ProjectBuilder extends Builder<Project> {
  private generator = new ProjectGenerator();
  
  withWorkspace(workspaceId: string): this {
    return this.set('workspaceId', workspaceId);
  }
  
  withName(name: string): this {
    return this.set('name', name).set('namespace', this.generator['generateSlug'](name));
  }
  
  withDescription(description: string): this {
    return this.set('description', description);
  }
  
  withEnvironment(environment: 'development' | 'staging' | 'production'): this {
    return this.set('environment', environment);
  }
  
  withOwner(userId: string, email: string): this {
    return this.set('owner', { userId, email });
  }
  
  withQuotas(quotas: Partial<Project['quotas']>): this {
    const currentQuotas = this.data.quotas || {
      maxApplications: 20,
      maxCPU: '16',
      maxMemory: '32Gi',
      maxStorage: '100Gi',
    };
    return this.set('quotas', { ...currentQuotas, ...quotas });
  }
  
  withGitIntegration(provider: 'github' | 'gitlab' | 'bitbucket', repository: string, branch: string = 'main'): this {
    return this.set('gitIntegration', {
      provider,
      repository,
      branch,
      autoSync: branch !== 'main',
    });
  }
  
  withLabels(labels: Record<string, string>): this {
    return this.set('labels', { ...this.data.labels, ...labels });
  }
  
  withAnnotations(annotations: Record<string, string>): this {
    return this.set('annotations', { ...this.data.annotations, ...annotations });
  }
  
  build(): Project {
    return this.generator.generate(this.data) as Project;
  }
}