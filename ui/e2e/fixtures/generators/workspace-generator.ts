import { BaseGenerator, Builder } from './base-generator';

export interface Workspace {
  id: string;
  organizationId: string;
  name: string;
  slug: string;
  type: 'shared' | 'dedicated';
  plan: 'shared' | 'dedicated';
  region: string;
  vclusterName: string;
  status: 'provisioning' | 'active' | 'suspended' | 'terminating';
  resources: {
    cpu: string;
    memory: string;
    storage: string;
    nodes?: number;
  };
  quotas: {
    maxProjects: number;
    maxApplications: number;
    maxCPU: string;
    maxMemory: string;
    maxStorage: string;
  };
  features: string[];
  createdAt: Date;
  members: Array<{
    userId: string;
    email: string;
    role: 'owner' | 'admin' | 'developer' | 'viewer';
    joinedAt: Date;
  }>;
  projects?: number;
  kubeconfig?: string;
  backupEnabled?: boolean;
  monitoringEnabled?: boolean;
}

export class WorkspaceGenerator extends BaseGenerator<Workspace> {
  private regions = ['us-east-1', 'us-west-2', 'eu-west-1', 'eu-central-1', 'ap-southeast-1', 'ap-northeast-1'];
  
  private planFeatures = {
    shared: ['monitoring', 'logs', 'basic-support'],
    dedicated: ['monitoring', 'logs', 'backups', 'custom-domains', 'priority-support', 'sla', 'node-autoscaling'],
  };
  
  generate(overrides?: Partial<Workspace>): Workspace {
    const name = overrides?.name || `${this.faker.company.buzzPhrase()} Workspace`;
    const plan = overrides?.plan || this.faker.helpers.arrayElement(['shared', 'dedicated']);
    const region = overrides?.region || this.faker.helpers.arrayElement(this.regions);
    
    const workspace: Workspace = {
      id: this.generateId('ws'),
      organizationId: overrides?.organizationId || this.generateId('org'),
      name,
      slug: this.generateSlug(name),
      type: plan,
      plan,
      region,
      vclusterName: `vcluster-${this.generateSlug(name)}-${this.faker.string.alphanumeric(6)}`,
      status: 'active',
      resources: plan === 'dedicated' ? {
        cpu: '16',
        memory: '64Gi',
        storage: '500Gi',
        nodes: 3,
      } : {
        cpu: '4',
        memory: '16Gi',
        storage: '100Gi',
      },
      quotas: plan === 'dedicated' ? {
        maxProjects: 50,
        maxApplications: 200,
        maxCPU: '64',
        maxMemory: '256Gi',
        maxStorage: '2Ti',
      } : {
        maxProjects: 10,
        maxApplications: 50,
        maxCPU: '8',
        maxMemory: '32Gi',
        maxStorage: '200Gi',
      },
      features: this.planFeatures[plan],
      createdAt: this.faker.date.past({ years: 1 }),
      members: [],
      projects: this.faker.number.int({ min: 0, max: plan === 'dedicated' ? 30 : 8 }),
      backupEnabled: plan === 'dedicated',
      monitoringEnabled: true,
      ...overrides,
    };
    
    // Generate members if not provided
    if (workspace.members.length === 0) {
      const memberCount = this.faker.number.int({ min: 1, max: 10 });
      workspace.members = this.generateMembers(memberCount, workspace.createdAt);
    }
    
    // Generate kubeconfig for active workspaces
    if (workspace.status === 'active') {
      workspace.kubeconfig = this.generateKubeconfig(workspace);
    }
    
    return workspace;
  }
  
  withTraits(traits: string[]): Workspace {
    const overrides: Partial<Workspace> = {};
    
    if (traits.includes('new')) {
      overrides.status = 'provisioning';
      overrides.createdAt = new Date();
      overrides.projects = 0;
    }
    
    if (traits.includes('large')) {
      overrides.plan = 'dedicated';
      overrides.type = 'dedicated';
      overrides.resources = {
        cpu: '32',
        memory: '128Gi',
        storage: '1Ti',
        nodes: 5,
      };
      overrides.projects = this.faker.number.int({ min: 20, max: 50 });
    }
    
    if (traits.includes('suspended')) {
      overrides.status = 'suspended';
    }
    
    if (traits.includes('full')) {
      const plan = overrides.plan || 'shared';
      const quotas = plan === 'dedicated' ? {
        maxProjects: 50,
        maxApplications: 200,
        maxCPU: '64',
        maxMemory: '256Gi',
        maxStorage: '2Ti',
      } : {
        maxProjects: 10,
        maxApplications: 50,
        maxCPU: '8',
        maxMemory: '32Gi',
        maxStorage: '200Gi',
      };
      
      overrides.projects = quotas.maxProjects;
      overrides.quotas = quotas;
    }
    
    return this.generate(overrides);
  }
  
  private generateMembers(count: number, workspaceCreatedAt: Date): Workspace['members'] {
    const roles: Array<'owner' | 'admin' | 'developer' | 'viewer'> = ['owner', 'admin', 'developer', 'viewer'];
    const members: Workspace['members'] = [];
    
    // Always have at least one owner
    members.push({
      userId: this.generateId('user'),
      email: this.faker.internet.email(),
      role: 'owner',
      joinedAt: workspaceCreatedAt,
    });
    
    // Generate additional members
    for (let i = 1; i < count; i++) {
      members.push({
        userId: this.generateId('user'),
        email: this.faker.internet.email(),
        role: this.faker.helpers.arrayElement(roles.slice(1)), // No additional owners
        joinedAt: this.generateDateInRange(workspaceCreatedAt, new Date()),
      });
    }
    
    return members;
  }
  
  private generateKubeconfig(workspace: Workspace): string {
    return Buffer.from(`
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://${workspace.vclusterName}.${workspace.region}.hexabase.io:443
    certificate-authority-data: ${this.faker.string.alphanumeric(64)}
  name: ${workspace.vclusterName}
contexts:
- context:
    cluster: ${workspace.vclusterName}
    user: ${workspace.vclusterName}-user
  name: ${workspace.vclusterName}
current-context: ${workspace.vclusterName}
users:
- name: ${workspace.vclusterName}-user
  user:
    token: ${this.faker.string.alphanumeric(128)}
`).toString('base64');
  }
}

export class WorkspaceBuilder extends Builder<Workspace> {
  private generator = new WorkspaceGenerator();
  
  withOrganization(organizationId: string): this {
    return this.set('organizationId', organizationId);
  }
  
  withName(name: string): this {
    return this.set('name', name).set('slug', this.generator['generateSlug'](name));
  }
  
  withPlan(plan: 'shared' | 'dedicated'): this {
    return this.set('plan', plan).set('type', plan);
  }
  
  withRegion(region: string): this {
    return this.set('region', region);
  }
  
  withStatus(status: Workspace['status']): this {
    return this.set('status', status);
  }
  
  withResources(cpu: string, memory: string, storage: string, nodes?: number): this {
    const resources: Workspace['resources'] = { cpu, memory, storage };
    if (nodes) resources.nodes = nodes;
    return this.set('resources', resources);
  }
  
  withQuotas(quotas: Workspace['quotas']): this {
    return this.set('quotas', quotas);
  }
  
  withMembers(members: Workspace['members']): this {
    return this.set('members', members);
  }
  
  addMember(email: string, role: 'owner' | 'admin' | 'developer' | 'viewer'): this {
    const currentMembers = this.data.members || [];
    currentMembers.push({
      userId: `user-${this.generator['faker'].string.alphanumeric(8)}`,
      email,
      role,
      joinedAt: new Date(),
    });
    return this.set('members', currentMembers);
  }
  
  withBackup(enabled: boolean): this {
    return this.set('backupEnabled', enabled);
  }
  
  build(): Workspace {
    return this.generator.generate(this.data);
  }
}