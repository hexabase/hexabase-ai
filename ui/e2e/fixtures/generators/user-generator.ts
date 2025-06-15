import { BaseGenerator, Builder } from './base-generator';

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  fullName: string;
  avatar?: string;
  role: 'owner' | 'admin' | 'developer' | 'viewer';
  status: 'active' | 'inactive' | 'suspended' | 'pending';
  organizations: Array<{
    id: string;
    name: string;
    role: 'owner' | 'admin' | 'member';
  }>;
  workspaces: Array<{
    id: string;
    name: string;
    role: 'owner' | 'admin' | 'developer' | 'viewer';
  }>;
  preferences: {
    theme: 'light' | 'dark' | 'system';
    language: string;
    notifications: {
      email: boolean;
      slack: boolean;
      inApp: boolean;
    };
    defaultOrganization?: string;
    defaultWorkspace?: string;
  };
  authentication: {
    provider: 'email' | 'google' | 'github' | 'microsoft' | 'saml';
    mfaEnabled: boolean;
    lastLogin?: Date;
    passwordChangedAt?: Date;
  };
  permissions?: string[];
  apiKeys?: Array<{
    id: string;
    name: string;
    lastUsed?: Date;
    expiresAt?: Date;
  }>;
  createdAt: Date;
  updatedAt: Date;
  emailVerified: boolean;
  phoneNumber?: string;
  timezone?: string;
  country?: string;
}

export class UserGenerator extends BaseGenerator<User> {
  private rolePermissions = {
    owner: ['*'],
    admin: ['workspace:*', 'project:*', 'application:*', 'monitoring:view', 'billing:view'],
    developer: ['project:*', 'application:*', 'monitoring:view'],
    viewer: ['workspace:view', 'project:view', 'application:view', 'monitoring:view'],
  };
  
  generate(overrides?: Partial<User>): User {
    const firstName = overrides?.firstName || this.faker.person.firstName();
    const lastName = overrides?.lastName || this.faker.person.lastName();
    const role = overrides?.role || this.faker.helpers.arrayElement(['owner', 'admin', 'developer', 'viewer']);
    
    const user: User = {
      id: this.generateId('user'),
      email: overrides?.email || this.faker.internet.email({ firstName, lastName }).toLowerCase(),
      firstName,
      lastName,
      fullName: `${firstName} ${lastName}`,
      avatar: this.faker.image.avatar(),
      role,
      status: 'active',
      organizations: [],
      workspaces: [],
      preferences: {
        theme: this.faker.helpers.arrayElement(['light', 'dark', 'system']),
        language: this.faker.helpers.arrayElement(['en', 'es', 'fr', 'de', 'ja', 'zh']),
        notifications: {
          email: this.faker.datatype.boolean({ probability: 0.8 }),
          slack: this.faker.datatype.boolean({ probability: 0.5 }),
          inApp: true,
        },
      },
      authentication: {
        provider: this.faker.helpers.arrayElement(['email', 'google', 'github']),
        mfaEnabled: role === 'owner' || (role === 'admin' && this.faker.datatype.boolean({ probability: 0.7 })),
        lastLogin: this.faker.date.recent({ days: 7 }),
        passwordChangedAt: this.faker.date.past({ years: 0.5 }),
      },
      permissions: this.rolePermissions[role],
      createdAt: this.faker.date.past({ years: 2 }),
      updatedAt: this.faker.date.recent({ days: 30 }),
      emailVerified: true,
      phoneNumber: this.faker.phone.number(),
      timezone: this.faker.helpers.arrayElement(['America/New_York', 'Europe/London', 'Asia/Tokyo', 'Australia/Sydney']),
      country: this.faker.location.countryCode(),
      ...overrides,
    };
    
    // Generate organizations if not provided
    if (user.organizations.length === 0) {
      const orgCount = this.faker.number.int({ min: 1, max: 3 });
      user.organizations = this.generateOrganizationMemberships(orgCount, user.role);
      if (user.organizations.length > 0) {
        user.preferences.defaultOrganization = user.organizations[0].id;
      }
    }
    
    // Generate workspaces if not provided
    if (user.workspaces.length === 0) {
      const workspaceCount = this.faker.number.int({ min: 1, max: 5 });
      user.workspaces = this.generateWorkspaceMemberships(workspaceCount, user.role);
      if (user.workspaces.length > 0) {
        user.preferences.defaultWorkspace = user.workspaces[0].id;
      }
    }
    
    // Generate API keys for developers and admins
    if ((role === 'admin' || role === 'developer') && this.faker.datatype.boolean({ probability: 0.6 })) {
      user.apiKeys = this.generateApiKeys(this.faker.number.int({ min: 1, max: 3 }));
    }
    
    return user;
  }
  
  withTraits(traits: string[]): User {
    const overrides: Partial<User> = {};
    
    if (traits.includes('new')) {
      overrides.status = 'pending';
      overrides.createdAt = new Date();
      overrides.emailVerified = false;
      overrides.organizations = [];
      overrides.workspaces = [];
    }
    
    if (traits.includes('suspended')) {
      overrides.status = 'suspended';
      overrides.authentication = {
        provider: 'email',
        mfaEnabled: false,
        lastLogin: this.faker.date.past({ years: 0.5 }),
      };
    }
    
    if (traits.includes('enterprise')) {
      overrides.role = 'owner';
      overrides.authentication = {
        provider: 'saml',
        mfaEnabled: true,
        lastLogin: this.faker.date.recent({ days: 1 }),
      };
      overrides.organizations = [{
        id: this.generateId('org'),
        name: `${this.faker.company.name()} Enterprise`,
        role: 'owner',
      }];
    }
    
    if (traits.includes('apiUser')) {
      overrides.apiKeys = this.generateApiKeys(5);
      overrides.authentication = {
        provider: 'email',
        mfaEnabled: true,
        lastLogin: undefined, // API users might not login via UI
      };
    }
    
    return this.generate(overrides);
  }
  
  private generateOrganizationMemberships(count: number, userRole: User['role']): User['organizations'] {
    const orgs = [];
    
    for (let i = 0; i < count; i++) {
      const orgRole = i === 0 && userRole === 'owner' ? 'owner' : 
        this.faker.helpers.arrayElement(['admin', 'member']);
      
      orgs.push({
        id: this.generateId('org'),
        name: this.faker.company.name(),
        role: orgRole,
      });
    }
    
    return orgs;
  }
  
  private generateWorkspaceMemberships(count: number, userRole: User['role']): User['workspaces'] {
    const workspaces = [];
    
    for (let i = 0; i < count; i++) {
      workspaces.push({
        id: this.generateId('ws'),
        name: `${this.faker.hacker.adjective()} ${this.faker.hacker.noun()}`,
        role: userRole,
      });
    }
    
    return workspaces;
  }
  
  private generateApiKeys(count: number): User['apiKeys'] {
    const keys = [];
    
    for (let i = 0; i < count; i++) {
      keys.push({
        id: this.generateId('key'),
        name: this.faker.helpers.arrayElement(['CI/CD Pipeline', 'Monitoring', 'Backup Script', 'Dev Environment']),
        lastUsed: this.faker.datatype.boolean({ probability: 0.8 }) ? 
          this.faker.date.recent({ days: 30 }) : undefined,
        expiresAt: this.faker.datatype.boolean({ probability: 0.5 }) ? 
          this.faker.date.future({ years: 1 }) : undefined,
      });
    }
    
    return keys;
  }
}

export class UserBuilder extends Builder<User> {
  private generator = new UserGenerator();
  
  withEmail(email: string): this {
    return this.set('email', email);
  }
  
  withName(firstName: string, lastName: string): this {
    return this.set('firstName', firstName)
      .set('lastName', lastName)
      .set('fullName', `${firstName} ${lastName}`);
  }
  
  withRole(role: User['role']): this {
    return this.set('role', role);
  }
  
  withStatus(status: User['status']): this {
    return this.set('status', status);
  }
  
  withOrganization(id: string, name: string, role: 'owner' | 'admin' | 'member'): this {
    const orgs = this.data.organizations || [];
    orgs.push({ id, name, role });
    return this.set('organizations', orgs);
  }
  
  withWorkspace(id: string, name: string, role: User['role']): this {
    const workspaces = this.data.workspaces || [];
    workspaces.push({ id, name, role });
    return this.set('workspaces', workspaces);
  }
  
  withAuthentication(provider: User['authentication']['provider'], mfaEnabled: boolean): this {
    return this.set('authentication', {
      provider,
      mfaEnabled,
      lastLogin: new Date(),
    });
  }
  
  withPreferences(preferences: Partial<User['preferences']>): this {
    const current = this.data.preferences || {
      theme: 'system',
      language: 'en',
      notifications: { email: true, slack: false, inApp: true },
    };
    return this.set('preferences', { ...current, ...preferences });
  }
  
  withApiKey(name: string, expiresAt?: Date): this {
    const keys = this.data.apiKeys || [];
    keys.push({
      id: `key-${this.generator['faker'].string.alphanumeric(8)}`,
      name,
      expiresAt,
    });
    return this.set('apiKeys', keys);
  }
  
  build(): User {
    return this.generator.generate(this.data);
  }
}