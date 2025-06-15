# Test Data Generators Guide

This guide explains how to use the test data generators and scenarios for E2E testing in Hexabase AI.

## Overview

Our test data generators provide realistic, interconnected test data for E2E tests. They use the Faker.js library with custom logic to create comprehensive test scenarios.

## Generator Architecture

```
e2e/fixtures/
├── generators/
│   ├── base-generator.ts         # Base classes
│   ├── organization-generator.ts # Organization data
│   ├── workspace-generator.ts    # Workspace data
│   ├── project-generator.ts      # Project data
│   ├── application-generator.ts  # Application configs
│   ├── user-generator.ts         # User accounts
│   ├── deployment-generator.ts   # Deployment data
│   ├── cronjob-generator.ts      # Scheduled jobs
│   ├── function-generator.ts     # Serverless functions
│   ├── backup-generator.ts       # Backup configs
│   ├── metrics-generator.ts      # Time-series data
│   ├── alert-generator.ts        # Alert rules
│   └── pipeline-generator.ts     # CI/CD pipelines
└── scenarios/
    ├── base-scenario.ts          # Base scenario class
    ├── startup-scenario.ts       # Small company setup
    └── enterprise-scenario.ts    # Large org setup
```

## Basic Usage

### Using Individual Generators

```typescript
import { ApplicationGenerator } from '../fixtures/generators/application-generator';

// Create generator instance
const appGen = new ApplicationGenerator();

// Generate random application
const app = appGen.generate();

// Generate with overrides
const customApp = appGen.generate({
  name: 'my-api',
  type: 'stateless',
  replicas: 3
});

// Generate with seed for reproducibility
const seededGen = new ApplicationGenerator(12345);
const consistentApp = seededGen.generate();
```

### Using Builder Pattern

```typescript
import { ApplicationBuilder } from '../fixtures/generators/application-generator';

const app = new ApplicationBuilder()
  .withName('payment-service')
  .withImage('node', '18-alpine')
  .withType('stateless')
  .withReplicas(3)
  .withPort(3000)
  .withResources(
    { cpu: '500m', memory: '1Gi' },  // requests
    { cpu: '1', memory: '2Gi' }       // limits
  )
  .withEnvironment({
    NODE_ENV: 'production',
    LOG_LEVEL: 'info'
  })
  .withIngress('api.example.com', true)
  .withAutoscaling(2, 10, 70)
  .build();
```

### Using Traits

Traits provide pre-configured variations:

```typescript
// Generate high-traffic application
const haApp = appGen.withTraits(['highAvailability']).generate();

// Generate failing cronjob
const failingJob = cronJobGen.withTraits(['failing']).generate();

// Generate large enterprise organization
const bigOrg = orgGen.withTraits(['large']).generate();
```

## Available Generators

### OrganizationGenerator

```typescript
const org = orgGen.generate({
  name: 'ACME Corp',
  plan: 'enterprise',
  members: 100
});

// Available traits: 'large', 'new', 'trial'
const trialOrg = orgGen.withTraits(['trial']).generate();
```

### WorkspaceGenerator

```typescript
const workspace = workspaceGen.generate({
  name: 'Production',
  plan: 'dedicated',
  region: 'us-east-1',
  resources: {
    cpu: '32',
    memory: '128Gi',
    storage: '1Ti',
    nodes: 5
  }
});

// Available traits: 'new', 'large', 'suspended', 'full'
const largeWorkspace = workspaceGen.withTraits(['large']).generate();
```

### ApplicationGenerator

```typescript
const app = appGen.generate({
  name: 'web-frontend',
  type: 'stateless',
  image: 'nginx',
  replicas: 3
});

// Available traits: 'database', 'highAvailability', 'microservice', 'publicFacing', 'canary'
const database = appGen.withTraits(['database']).generate();
```

### DeploymentGenerator

```typescript
const deployment = deployGen.generate({
  status: 'succeeded',
  strategy: {
    type: 'canary',
    config: {
      canarySteps: [
        { weight: 10, duration: '5m' },
        { weight: 50, duration: '10m' },
        { weight: 100, duration: '0m' }
      ]
    }
  }
});

// Available traits: 'failed', 'canary', 'blueGreen', 'automated', 'rollback'
const canaryDeploy = deployGen.withTraits(['canary']).generate();
```

### MetricsGenerator

```typescript
// Generate resource metrics
const metrics = metricsGen.generateResourceMetrics({
  duration: 3600,     // 1 hour
  interval: 60,       // 1 minute
  workloadType: 'web',
  resourceType: 'pod'
});

// Generate application metrics
const appMetrics = metricsGen.generateApplicationMetrics({
  duration: 86400,    // 24 hours
  interval: 300,      // 5 minutes
  appType: 'api',
  load: 'high'
});
```

## Test Scenarios

### Using Pre-built Scenarios

```typescript
import { TestDataManager } from '../utils/test-data-manager';

test('startup company scenario', async ({ page }) => {
  const testData = new TestDataManager(page);
  
  // Load startup scenario
  const scenario = await testData.loadScenario('startup', 12345);
  
  // Access generated data
  const orgs = testData.getOrganizations();
  const workspaces = testData.getWorkspaces();
  const projects = testData.getProjects();
  
  // Find specific entities
  const prodWorkspace = testData.findWorkspaceByName('Production');
  const apiProject = testData.findProjectByName('API Services');
  
  // Get test credentials
  const admin = testData.getAdminCredentials();
  await loginPage.login(admin.email, admin.password);
});
```

### Available Scenarios

#### Startup Scenario
- 1 professional organization
- 2 workspaces (production, staging)
- 5 team members
- Basic microservices setup
- Simple CI/CD pipeline

#### Enterprise Scenario
- Multiple enterprise organizations
- 6 dedicated workspaces across regions
- 37+ team members with various roles
- Complex microservices architecture
- Full backup and disaster recovery
- Advanced monitoring and alerting

### Creating Custom Scenarios

```typescript
import { BaseScenario } from '../fixtures/scenarios/base-scenario';

export class CustomScenario extends BaseScenario {
  generate(): ScenarioData {
    // Create organization
    const org = this.orgGenerator.generate({
      name: 'Custom Corp',
      plan: 'professional'
    });
    this.data.organizations.push(org);
    
    // Create workspaces
    for (let i = 0; i < 3; i++) {
      const workspace = this.workspaceGenerator.generate({
        name: `Environment ${i}`,
        organizationId: org.id,
        plan: 'shared'
      });
      this.data.workspaces.push(workspace);
    }
    
    // Link entities
    this.linkEntities();
    
    return this.data;
  }
}
```

## Advanced Patterns

### Generating Related Data

```typescript
// Generate organization with full hierarchy
const org = orgGen.generate();
const workspace = workspaceGen.generate({ organizationId: org.id });
const project = projectGen.generate({ workspaceId: workspace.id });
const app = appGen.generate({ projectId: project.id });
const deployment = deployGen.generate({ applicationId: app.id });
```

### Batch Generation

```typescript
// Generate multiple items
const apps = appGen.generateMany(10, {
  type: 'stateless',
  projectId: project.id
});

// Generate with variations
const users = Array.from({ length: 5 }, (_, i) => 
  userGen.generate({
    role: i === 0 ? 'owner' : 'developer',
    organizationId: org.id
  })
);
```

### Time-based Data

```typescript
// Generate metrics for the last week
const weeklyMetrics = metricsGen.generateResourceMetrics({
  duration: 7 * 24 * 60 * 60, // 7 days in seconds
  interval: 300,              // 5 minute intervals
  workloadType: 'web'
});

// Generate deployment history
const deployments = Array.from({ length: 10 }, (_, i) => 
  deployGen.generate({
    createdAt: new Date(Date.now() - i * 24 * 60 * 60 * 1000), // Past 10 days
    status: i === 0 ? 'running' : 'succeeded'
  })
);
```

### Conditional Generation

```typescript
// Generate data based on plan
const workspace = workspaceGen.generate();
if (workspace.plan === 'dedicated') {
  // Add backup configuration for dedicated plans
  const backupPolicy = backupGen.generateBackupPolicy({
    workspaceId: workspace.id,
    enabled: true,
    schedule: { frequency: 'daily' }
  });
  
  // Generate recent backups
  const backups = backupGen.generateMany(5, {
    policyId: backupPolicy.id,
    status: 'completed'
  });
}
```

## Testing Different States

### Success States

```typescript
const successfulDeployment = deployGen.generate({
  status: 'succeeded',
  metrics: {
    errorRate: 0.001,
    latency: { p50: 50, p95: 100, p99: 200 }
  }
});
```

### Error States

```typescript
const failedApp = appGen.generate({
  status: 'error',
  error: 'CrashLoopBackOff'
});

const alerting = alertGen.withTraits(['critical']).generate({
  status: 'active',
  severity: 'critical'
});
```

### Edge Cases

```typescript
// Resource limits
const maxedWorkspace = workspaceGen.withTraits(['full']).generate();

// Large datasets
const manyApps = appGen.generateMany(100);

// Complex relationships
const interconnected = new EnterpriseScenario().generate();
```

## Best Practices

### 1. Use Realistic Data

```typescript
// ✅ Good - Realistic application names
const app = appGen.generate(); // Generates "payment-api", "user-service", etc.

// ❌ Bad - Hardcoded unrealistic data
const app = { name: 'test-app-1', image: 'test' };
```

### 2. Maintain Consistency

```typescript
// Use seeds for reproducible tests
const gen = new ApplicationGenerator(12345);
const app1 = gen.generate(); // Always generates the same data
const app2 = gen.generate(); // Different but predictable
```

### 3. Test Data Cleanup

```typescript
test('with cleanup', async ({ page }) => {
  const testData = new TestDataManager(page);
  const scenario = await testData.loadScenario('startup');
  
  try {
    // Run your tests
  } finally {
    // Data is automatically cleaned up when test ends
    // But you can manually cleanup if needed
    await testData.cleanup();
  }
});
```

### 4. Avoid Over-Generation

```typescript
// ✅ Good - Generate only what you need
const app = appGen.generate({ name: 'test-app' });

// ❌ Bad - Generating unnecessary data
const fullScenario = new EnterpriseScenario().generate(); // Too much for simple test
```

## Debugging Generated Data

### Inspect Generated Data

```typescript
const app = appGen.generate();
console.log('Generated app:', JSON.stringify(app, null, 2));
```

### Export Scenario Data

```typescript
const scenario = new StartupScenario().generate();
fs.writeFileSync('test-data.json', JSON.stringify(scenario, null, 2));
```

### Validate Relationships

```typescript
const scenario = testData.getScenario();

// Check all apps belong to valid projects
scenario.applications.forEach(app => {
  const project = scenario.projects.find(p => p.id === app.projectId);
  expect(project).toBeDefined();
});
```

## Common Use Cases

### Login Testing

```typescript
test('login with different user roles', async ({ page }) => {
  const users = [
    userGen.generate({ role: 'owner' }),
    userGen.generate({ role: 'admin' }),
    userGen.generate({ role: 'developer' }),
    userGen.generate({ role: 'viewer' })
  ];
  
  for (const user of users) {
    await loginPage.login(user.email, 'Test123!');
    await verifyPermissions(user.role);
    await loginPage.logout();
  }
});
```

### Load Testing

```typescript
test('handle many applications', async ({ page }) => {
  const apps = appGen.generateMany(50, {
    projectId: project.id,
    type: 'stateless'
  });
  
  // Test pagination
  await projectPage.openProject(project.name);
  await expect(page.getByTestId('app-count')).toHaveText('50 applications');
  await expect(page.getByTestId('pagination')).toBeVisible();
});
```

### Monitoring Testing

```typescript
test('alert on high CPU usage', async ({ page }) => {
  // Generate metrics with spike
  const metrics = metricsGen.withTraits(['spike']).generate({
    name: 'cpu_usage_percent',
    spikeAt: 0.5,
    spikeMultiplier: 5
  });
  
  // Generate corresponding alert
  const alert = alertGen.generate({
    source: { metric: 'cpu_usage_percent' },
    condition: { type: 'threshold', operator: '>', value: 80 },
    status: 'active'
  });
  
  await monitoringPage.verifyAlert(alert);
});
```

## Extending Generators

### Adding New Traits

```typescript
// In application-generator.ts
withTraits(traits: string[]): Application {
  const overrides: Partial<Application> = {};
  
  if (traits.includes('myCustomTrait')) {
    overrides.replicas = 10;
    overrides.resources = {
      requests: { cpu: '2', memory: '4Gi' },
      limits: { cpu: '4', memory: '8Gi' }
    };
  }
  
  return this.generate(overrides);
}
```

### Creating New Generators

```typescript
import { BaseGenerator } from './base-generator';

export class CustomResourceGenerator extends BaseGenerator<CustomResource> {
  generate(overrides?: Partial<CustomResource>): CustomResource {
    return {
      id: this.generateId('custom'),
      name: this.faker.commerce.productName(),
      type: this.faker.helpers.arrayElement(['type1', 'type2']),
      createdAt: this.faker.date.past(),
      ...overrides
    };
  }
  
  withTraits(traits: string[]): CustomResource {
    // Implement trait logic
    return this.generate();
  }
}
```

## Summary

Test data generators provide:
- Realistic, interconnected test data
- Reproducible test scenarios
- Flexible data generation
- Type-safe test fixtures
- Easy maintenance and updates

Use them to create comprehensive test scenarios that closely mirror production environments while maintaining test reliability and speed.