import { BaseScenario, ScenarioData } from './base-scenario';

/**
 * Startup Company Scenario
 * - 1 organization with professional plan
 * - 2 workspaces (production and staging)
 * - Small team (5 users)
 * - Microservices architecture
 * - Basic monitoring and CI/CD
 */
export class StartupScenario extends BaseScenario {
  generate(): ScenarioData {
    // Create organization
    const org = this.orgGenerator.generate({
      name: 'TechStartup Inc',
      plan: 'professional',
      members: 5,
      workspaces: 2,
    });
    this.data.organizations.push(org);
    
    // Create team members
    const users = [
      this.userGenerator.generate({
        email: 'founder@techstartup.com',
        firstName: 'Alex',
        lastName: 'Chen',
        role: 'owner',
      }),
      this.userGenerator.generate({
        email: 'cto@techstartup.com',
        firstName: 'Sarah',
        lastName: 'Johnson',
        role: 'admin',
      }),
      this.userGenerator.generate({
        email: 'backend@techstartup.com',
        firstName: 'Mike',
        lastName: 'Wilson',
        role: 'developer',
      }),
      this.userGenerator.generate({
        email: 'frontend@techstartup.com',
        firstName: 'Emma',
        lastName: 'Davis',
        role: 'developer',
      }),
      this.userGenerator.generate({
        email: 'devops@techstartup.com',
        firstName: 'Chris',
        lastName: 'Taylor',
        role: 'developer',
      }),
    ];
    this.data.users.push(...users);
    
    // Create workspaces
    const prodWorkspace = this.workspaceGenerator.generate({
      name: 'Production',
      organizationId: org.id,
      plan: 'shared',
      region: 'us-east-1',
      members: users.map((u, i) => ({
        userId: u.id,
        email: u.email,
        role: i === 0 ? 'owner' : i === 1 ? 'admin' : 'developer',
        joinedAt: new Date(),
      })),
    });
    
    const stagingWorkspace = this.workspaceGenerator.generate({
      name: 'Staging',
      organizationId: org.id,
      plan: 'shared',
      region: 'us-east-1',
      members: users.slice(1).map(u => ({
        userId: u.id,
        email: u.email,
        role: 'developer',
        joinedAt: new Date(),
      })),
    });
    
    this.data.workspaces.push(prodWorkspace, stagingWorkspace);
    
    // Create projects
    const apiProject = this.projectGenerator.generate({
      name: 'API Services',
      workspaceId: prodWorkspace.id,
      environment: 'production',
      gitIntegration: {
        provider: 'github',
        repository: 'techstartup/api',
        branch: 'main',
        autoSync: false,
      },
    });
    
    const webProject = this.projectGenerator.generate({
      name: 'Web Application',
      workspaceId: prodWorkspace.id,
      environment: 'production',
      gitIntegration: {
        provider: 'github',
        repository: 'techstartup/web-app',
        branch: 'main',
        autoSync: false,
      },
    });
    
    const stagingProject = this.projectGenerator.generate({
      name: 'Staging Environment',
      workspaceId: stagingWorkspace.id,
      environment: 'staging',
      gitIntegration: {
        provider: 'github',
        repository: 'techstartup/api',
        branch: 'develop',
        autoSync: true,
      },
    });
    
    this.data.projects.push(apiProject, webProject, stagingProject);
    
    // Create applications
    const apps = [
      // Production apps
      this.appGenerator.generate({
        name: 'api-gateway',
        projectId: apiProject.id,
        type: 'stateless',
        image: 'nginx',
        replicas: 2,
        port: 80,
        networking: {
          service: true,
          ingress: true,
          domain: 'api.techstartup.com',
          tls: true,
          loadBalancer: false,
        },
      }),
      this.appGenerator.generate({
        name: 'auth-service',
        projectId: apiProject.id,
        type: 'stateless',
        image: 'node',
        tag: '18-alpine',
        replicas: 2,
        port: 3000,
      }),
      this.appGenerator.generate({
        name: 'user-service',
        projectId: apiProject.id,
        type: 'stateless',
        image: 'node',
        tag: '18-alpine',
        replicas: 2,
        port: 3001,
      }),
      this.appGenerator.generate({
        name: 'postgres-db',
        projectId: apiProject.id,
        type: 'stateful',
        image: 'postgres',
        tag: '14',
        replicas: 1,
        port: 5432,
        volumes: [{
          name: 'postgres-data',
          type: 'persistent',
          mountPath: '/var/lib/postgresql/data',
          size: '20Gi',
        }],
      }),
      this.appGenerator.generate({
        name: 'redis-cache',
        projectId: apiProject.id,
        type: 'stateful',
        image: 'redis',
        tag: '7-alpine',
        replicas: 1,
        port: 6379,
      }),
      this.appGenerator.generate({
        name: 'web-frontend',
        projectId: webProject.id,
        type: 'stateless',
        image: 'nginx',
        replicas: 3,
        port: 80,
        networking: {
          service: true,
          ingress: true,
          domain: 'app.techstartup.com',
          tls: true,
          loadBalancer: true,
        },
      }),
    ];
    
    // Staging apps (fewer replicas)
    const stagingApps = [
      this.appGenerator.generate({
        name: 'api-gateway-staging',
        projectId: stagingProject.id,
        type: 'stateless',
        image: 'nginx',
        replicas: 1,
        port: 80,
      }),
      this.appGenerator.generate({
        name: 'auth-service-staging',
        projectId: stagingProject.id,
        type: 'stateless',
        image: 'node',
        tag: '18-alpine',
        replicas: 1,
        port: 3000,
      }),
    ];
    
    this.data.applications.push(...apps, ...stagingApps);
    
    // Create deployments
    apps.forEach(app => {
      const deployment = this.deploymentGenerator.generate({
        applicationId: app.id,
        status: 'succeeded',
        strategy: { type: 'rolling' },
      });
      this.data.deployments.push(deployment);
    });
    
    // Create CronJobs
    const backupJob = this.cronJobGenerator.generate({
      name: 'database-backup',
      projectId: apiProject.id,
      schedule: '0 2 * * *', // Daily at 2 AM
      enabled: true,
      jobTemplate: {
        image: 'postgres:14',
        command: ['pg_dump'],
        args: ['-h', 'postgres-db', '-U', 'postgres', '-d', 'app_db'],
      },
    });
    
    const cleanupJob = this.cronJobGenerator.generate({
      name: 'log-cleanup',
      projectId: apiProject.id,
      schedule: '0 0 * * 0', // Weekly
      enabled: true,
    });
    
    this.data.cronjobs.push(backupJob, cleanupJob);
    
    // Create serverless functions
    const webhookHandler = this.functionGenerator.generate({
      name: 'webhook-processor',
      projectId: apiProject.id,
      runtime: 'nodejs18',
      triggers: [{
        type: 'http',
        config: {
          method: 'POST',
          path: '/webhooks/github',
          auth: 'api-key',
        },
      }],
    });
    
    const reportGenerator = this.functionGenerator.generate({
      name: 'monthly-report',
      projectId: apiProject.id,
      runtime: 'python311',
      triggers: [{
        type: 'schedule',
        config: {
          schedule: '0 9 1 * *', // Monthly on 1st at 9 AM
          timezone: 'America/New_York',
        },
      }],
    });
    
    this.data.functions.push(webhookHandler, reportGenerator);
    
    // Create monitoring alerts
    const alerts = [
      this.alertGenerator.generate({
        name: 'API High Error Rate',
        severity: 'critical',
        status: 'active',
        source: {
          type: 'metric',
          metric: 'http_error_rate',
          labels: { service: 'api-gateway' },
        },
        condition: {
          type: 'threshold',
          operator: '>',
          value: 0.05,
          duration: '5m',
        },
      }),
      this.alertGenerator.generate({
        name: 'Database Connection Pool',
        severity: 'warning',
        status: 'resolved',
        source: {
          type: 'metric',
          metric: 'pg_connection_pool_usage',
        },
        condition: {
          type: 'threshold',
          operator: '>',
          value: 0.8,
          duration: '10m',
        },
      }),
    ];
    
    this.data.monitoring.alerts.push(...alerts);
    
    // Create CI/CD pipeline
    const pipeline = this.pipelineGenerator.generate({
      name: 'API Deployment Pipeline',
      projectId: apiProject.id,
      source: {
        type: 'github',
        repository: 'techstartup/api',
        branch: 'main',
      },
      trigger: {
        type: 'push',
        branches: ['main', 'develop'],
      },
      status: 'success',
    });
    
    this.data.pipelines.push(pipeline);
    
    // Generate metrics
    const cpuMetrics = this.metricsGenerator.generateResourceMetrics({
      duration: 86400, // 24 hours
      interval: 300, // 5 minutes
      workloadType: 'web',
    });
    
    const appMetrics = this.metricsGenerator.generateApplicationMetrics({
      duration: 86400,
      interval: 60,
      appType: 'api',
      load: 'medium',
    });
    
    this.data.monitoring.metrics.push(cpuMetrics, appMetrics);
    
    // Link all entities
    this.linkEntities();
    
    return this.data;
  }
}