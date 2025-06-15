import { BaseGenerator, Builder } from './base-generator';

export interface Organization {
  id: string;
  name: string;
  slug: string;
  plan: 'free' | 'professional' | 'enterprise';
  owner: {
    id: string;
    email: string;
    name: string;
  };
  members: number;
  workspaces: number;
  createdAt: Date;
  features: string[];
  billing?: {
    customerId: string;
    subscriptionId: string;
    nextBillingDate: Date;
    amount: number;
  };
  settings?: {
    ssoEnabled: boolean;
    mfaRequired: boolean;
    ipWhitelist?: string[];
  };
}

export class OrganizationGenerator extends BaseGenerator<Organization> {
  private planFeatures = {
    free: ['basic-monitoring', 'shared-resources', '1-workspace'],
    professional: ['advanced-monitoring', 'dedicated-resources', '10-workspaces', 'backup-restore', 'sla'],
    enterprise: ['full-monitoring', 'dedicated-nodes', 'unlimited-workspaces', 'backup-restore', 'sla', 'custom-domain', 'audit-logs', 'sso']
  };
  
  generate(overrides?: Partial<Organization>): Organization {
    const name = overrides?.name || this.faker.company.name();
    const plan = overrides?.plan || this.faker.helpers.arrayElement(['free', 'professional', 'enterprise']);
    
    const org: Organization = {
      id: this.generateId('org'),
      name,
      slug: this.generateSlug(name),
      plan,
      owner: {
        id: this.generateId('user'),
        email: this.faker.internet.email(),
        name: this.faker.person.fullName(),
      },
      members: this.faker.number.int({ min: 1, max: plan === 'enterprise' ? 100 : 20 }),
      workspaces: this.faker.number.int({ min: 1, max: plan === 'enterprise' ? 50 : 10 }),
      createdAt: this.faker.date.past({ years: 2 }),
      features: this.planFeatures[plan],
      ...overrides
    };
    
    // Add billing for paid plans
    if (plan !== 'free') {
      org.billing = {
        customerId: `cus_${this.faker.string.alphanumeric(14)}`,
        subscriptionId: `sub_${this.faker.string.alphanumeric(14)}`,
        nextBillingDate: this.faker.date.future({ years: 0.1 }),
        amount: plan === 'professional' ? 299 : 999,
      };
    }
    
    // Add enterprise settings
    if (plan === 'enterprise') {
      org.settings = {
        ssoEnabled: this.faker.datatype.boolean(),
        mfaRequired: true,
        ipWhitelist: this.faker.datatype.boolean() ? 
          this.faker.helpers.multiple(() => this.faker.internet.ip(), { count: { min: 1, max: 5 } }) : 
          undefined,
      };
    }
    
    return org;
  }
  
  withTraits(traits: string[]): Organization {
    const overrides: Partial<Organization> = {};
    
    if (traits.includes('large')) {
      overrides.members = this.faker.number.int({ min: 50, max: 200 });
      overrides.workspaces = this.faker.number.int({ min: 20, max: 100 });
    }
    
    if (traits.includes('new')) {
      overrides.createdAt = this.faker.date.recent({ days: 7 });
      overrides.members = 1;
      overrides.workspaces = 1;
    }
    
    if (traits.includes('trial')) {
      overrides.plan = 'professional';
      overrides.billing = {
        customerId: `cus_${this.faker.string.alphanumeric(14)}`,
        subscriptionId: `sub_${this.faker.string.alphanumeric(14)}`,
        nextBillingDate: this.faker.date.future({ years: 0.05 }), // ~2 weeks
        amount: 0,
      };
    }
    
    return this.generate(overrides);
  }
}

export class OrganizationBuilder extends Builder<Organization> {
  private generator = new OrganizationGenerator();
  
  withName(name: string): this {
    return this.set('name', name).set('slug', this.generator['generateSlug'](name));
  }
  
  withPlan(plan: Organization['plan']): this {
    return this.set('plan', plan);
  }
  
  withOwner(email: string, name: string): this {
    return this.set('owner', {
      id: `user-${this.generator['faker'].string.alphanumeric(8)}`,
      email,
      name,
    });
  }
  
  withMembers(count: number): this {
    return this.set('members', count);
  }
  
  withBilling(customerId: string, amount: number): this {
    return this.set('billing', {
      customerId,
      subscriptionId: `sub_${this.generator['faker'].string.alphanumeric(14)}`,
      nextBillingDate: this.generator['faker'].date.future({ years: 0.1 }),
      amount,
    });
  }
  
  withEnterpriseSettings(ssoEnabled: boolean, mfaRequired: boolean): this {
    return this.set('settings', {
      ssoEnabled,
      mfaRequired,
    });
  }
  
  build(): Organization {
    return this.generator.generate(this.data);
  }
}