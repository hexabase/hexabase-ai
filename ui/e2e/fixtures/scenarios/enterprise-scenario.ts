import { BaseScenario, ScenarioData } from './base-scenario';

/**
 * Enterprise Scenario
 * - Multiple organizations with enterprise plans
 * - Dedicated workspaces with multiple teams
 * - Complex application architecture
 * - Full backup and disaster recovery
 * - Advanced monitoring and compliance
 */
export class EnterpriseScenario extends BaseScenario {
  generate(): ScenarioData {
    // Create main enterprise organization
    const mainOrg = this.orgGenerator.generate({
      name: 'Global Enterprise Corp',
      plan: 'enterprise',
      members: 250,
      workspaces: 15,
      settings: {
        ssoEnabled: true,
        mfaRequired: true,
        ipWhitelist: ['10.0.0.0/8', '172.16.0.0/12'],
      },
    });
    
    // Create subsidiary organization
    const subsidiaryOrg = this.orgGenerator.generate({
      name: 'Enterprise Cloud Division',
      plan: 'enterprise',
      members: 75,
      workspaces: 5,
    });
    
    this.data.organizations.push(mainOrg, subsidiaryOrg);
    
    // Create diverse user base
    const userRoles = [
      { count: 2, role: 'owner' as const },
      { count: 5, role: 'admin' as const },
      { count: 20, role: 'developer' as const },
      { count: 10, role: 'viewer' as const },
    ];
    
    userRoles.forEach(({ count, role }) => {
      for (let i = 0; i < count; i++) {
        const user = this.userGenerator.withTraits(['enterprise']).generate({
          role,
          organizations: [{
            id: mainOrg.id,
            name: mainOrg.name,
            role: role === 'owner' ? 'owner' : 'member',
          }],
        });
        this.data.users.push(user);
      }
    });
    
    // Create dedicated workspaces for different environments
    const environments = [
      { name: 'Production US-East', region: 'us-east-1', purpose: 'production' },
      { name: 'Production EU-West', region: 'eu-west-1', purpose: 'production' },
      { name: 'Production APAC', region: 'ap-southeast-1', purpose: 'production' },
      { name: 'Staging Global', region: 'us-east-1', purpose: 'staging' },
      { name: 'Development', region: 'us-west-2', purpose: 'development' },
      { name: 'Data Analytics', region: 'us-east-1', purpose: 'analytics' },
    ];
    
    environments.forEach(env => {
      const workspace = this.workspaceGenerator.generate({
        name: env.name,
        organizationId: mainOrg.id,
        plan: 'dedicated',
        region: env.region,
        resources: {
          cpu: env.purpose === 'production' ? '64' : '32',
          memory: env.purpose === 'production' ? '256Gi' : '128Gi',
          storage: env.purpose === 'production' ? '2Ti' : '1Ti',
          nodes: env.purpose === 'production' ? 5 : 3,
        },
        backupEnabled: true,
        monitoringEnabled: true,
      });
      this.data.workspaces.push(workspace);
    });
    
    // Create projects for microservices architecture
    const prodWorkspace = this.data.workspaces[0];
    const projectTypes = [
      'Core Services',
      'Customer Portal',
      'Admin Dashboard',
      'Payment Processing',
      'Analytics Engine',
      'API Gateway',
      'Authentication',
      'Notification Service',
    ];
    
    projectTypes.forEach(projectName => {
      const project = this.projectGenerator.generate({
        name: projectName,
        workspaceId: prodWorkspace.id,
        environment: 'production',
        quotas: {
          maxApplications: 50,
          maxCPU: '32',
          maxMemory: '128Gi',
          maxStorage: '500Gi',
        },
        labels: {
          'compliance': 'pci-dss',
          'cost-center': 'it-infrastructure',
          'business-unit': projectName.toLowerCase().replace(/\s+/g, '-'),
        },
      });
      this.data.projects.push(project);
    });
    
    // Create diverse application portfolio
    const coreProject = this.data.projects[0];
    
    // Databases
    const databases = [
      { name: 'postgres-primary', image: 'postgres:14', size: '100Gi' },
      { name: 'postgres-replica', image: 'postgres:14', size: '100Gi' },
      { name: 'mongodb-cluster', image: 'mongodb:6', size: '200Gi' },
      { name: 'redis-cluster', image: 'redis:7', size: '50Gi' },
      { name: 'elasticsearch', image: 'elasticsearch:8', size: '500Gi' },
    ];
    
    databases.forEach(db => {
      const app = this.appGenerator.generate({
        name: db.name,
        projectId: coreProject.id,
        type: 'stateful',
        image: db.image,
        replicas: db.name.includes('cluster') ? 3 : 1,
        volumes: [{
          name: `${db.name}-data`,
          type: 'persistent',
          mountPath: '/data',
          size: db.size,
        }],
        resources: {
          requests: { cpu: '2', memory: '8Gi' },
          limits: { cpu: '4', memory: '16Gi' },
        },
      });
      this.data.applications.push(app);
    });
    
    // Microservices
    const microservices = [
      'user-service',
      'order-service',
      'inventory-service',
      'payment-service',
      'shipping-service',
      'recommendation-service',
      'search-service',
      'reporting-service',
    ];
    
    microservices.forEach(service => {
      const app = this.appGenerator.withTraits(['highAvailability']).generate({
        name: service,
        projectId: coreProject.id,
        type: 'stateless',
        image: 'company/services',
        tag: service,
        autoscaling: {
          enabled: true,
          minReplicas: 3,
          maxReplicas: 20,
          targetCPU: 70,
          targetMemory: 80,
        },
      });
      this.data.applications.push(app);
    });
    
    // Create complex deployments
    this.data.applications.forEach(app => {
      if (app.type === 'stateless') {
        // Blue-green deployment for stateless apps
        const deployment = this.deploymentGenerator.withTraits(['blueGreen']).generate({
          applicationId: app.id,
          status: 'succeeded',
        });
        this.data.deployments.push(deployment);
      } else {
        // Rolling deployment for stateful apps
        const deployment = this.deploymentGenerator.generate({
          applicationId: app.id,
          status: 'succeeded',
          strategy: {
            type: 'rolling',
            config: {
              maxSurge: '1',
              maxUnavailable: '0',
            },
          },
        });
        this.data.deployments.push(deployment);
      }
    });
    
    // Create backup infrastructure
    const backupStorage = this.backupGenerator.generateBackupStorage({
      workspaceId: prodWorkspace.id,
      name: 'primary-backup-storage',
      type: 'proxmox',
      capacity: {
        total: 10000, // 10TB
        used: 3500,
        available: 6500,
      },
      encryption: {
        enabled: true,
        algorithm: 'AES-256',
        keyManagement: 'customer',
      },
    });
    
    const drStorage = this.backupGenerator.generateBackupStorage({
      workspaceId: prodWorkspace.id,
      name: 'disaster-recovery-storage',
      type: 's3',
      config: {
        s3Bucket: 'enterprise-dr-backups',
        s3Region: 'us-west-2',
      },
      capacity: {
        total: 50000, // 50TB
        used: 15000,
        available: 35000,
      },
    });
    
    this.data.backups.storages.push(backupStorage, drStorage);
    
    // Create backup policies
    const policies = [
      {
        name: 'Database Full Backup',
        frequency: 'daily' as const,
        targets: [{
          type: 'database' as const,
          selector: { labels: { tier: 'database' } },
        }],
        backupType: 'full' as const,
      },
      {
        name: 'Application Incremental',
        frequency: 'hourly' as const,
        targets: [{
          type: 'application' as const,
          selector: { labels: { tier: 'application' } },
        }],
        backupType: 'incremental' as const,
      },
      {
        name: 'Disaster Recovery',
        frequency: 'weekly' as const,
        targets: [{
          type: 'namespace' as const,
          selector: { all: true },
        }],
        backupType: 'full' as const,
        storageId: drStorage.id,
      },
    ];
    
    policies.forEach(policyConfig => {
      const policy = this.backupGenerator.generateBackupPolicy({
        workspaceId: prodWorkspace.id,
        name: policyConfig.name,
        schedule: { frequency: policyConfig.frequency, timezone: 'UTC' },
        targets: policyConfig.targets,
        backupType: policyConfig.backupType,
        storageId: policyConfig.storageId || backupStorage.id,
      });
      this.data.backups.policies.push(policy);
      
      // Generate recent backups
      for (let i = 0; i < 5; i++) {
        const backup = this.backupGenerator.generate({
          policyId: policy.id,
          workspaceId: prodWorkspace.id,
          type: policyConfig.backupType,
          status: 'completed',
        });
        this.data.backups.backups.push(backup);
      }
    });
    
    // Create enterprise-grade monitoring
    const criticalAlerts = [
      'Database Connection Pool Exhausted',
      'API Gateway 5xx Errors',
      'Payment Service Timeout',
      'Disk Space Critical',
      'Certificate Expiry Warning',
      'Security Scan Failed',
      'Backup Job Failed',
    ];
    
    criticalAlerts.forEach(alertName => {
      const alert = this.alertGenerator.withTraits(['critical']).generate({
        name: alertName,
        status: alertName.includes('Failed') ? 'active' : 'resolved',
      });
      this.data.monitoring.alerts.push(alert);
    });
    
    // Create CronJobs for maintenance
    const maintenanceJobs = [
      {
        name: 'database-vacuum',
        schedule: '0 3 * * 0', // Weekly at 3 AM Sunday
        image: 'postgres:14',
        command: ['vacuumdb', '--all', '--analyze'],
      },
      {
        name: 'log-rotation',
        schedule: '0 0 * * *', // Daily at midnight
        image: 'alpine:latest',
        command: ['sh', '-c', 'find /logs -mtime +30 -delete'],
      },
      {
        name: 'security-scan',
        schedule: '0 2 * * *', // Daily at 2 AM
        image: 'aquasec/trivy:latest',
        command: ['trivy', 'image', '--severity', 'HIGH,CRITICAL'],
      },
      {
        name: 'certificate-renewal',
        schedule: '0 0 1 * *', // Monthly on 1st
        image: 'certbot/certbot:latest',
        command: ['certbot', 'renew'],
      },
    ];
    
    maintenanceJobs.forEach(jobConfig => {
      const job = this.cronJobGenerator.generate({
        name: jobConfig.name,
        projectId: coreProject.id,
        schedule: jobConfig.schedule,
        jobTemplate: {
          image: jobConfig.image,
          command: jobConfig.command,
          resources: {
            requests: { cpu: '500m', memory: '1Gi' },
            limits: { cpu: '2', memory: '4Gi' },
          },
        },
      });
      this.data.cronjobs.push(job);
    });
    
    // Create serverless functions for event processing
    const functions = [
      {
        name: 'order-processor',
        trigger: 'queue',
        runtime: 'nodejs18' as const,
      },
      {
        name: 'invoice-generator',
        trigger: 'event',
        runtime: 'python311' as const,
      },
      {
        name: 'fraud-detector',
        trigger: 'http',
        runtime: 'python311' as const,
      },
      {
        name: 'report-scheduler',
        trigger: 'schedule',
        runtime: 'go119' as const,
      },
    ];
    
    functions.forEach(funcConfig => {
      const func = this.functionGenerator.withTraits(['highTraffic']).generate({
        name: funcConfig.name,
        projectId: coreProject.id,
        runtime: funcConfig.runtime,
        memorySize: 1024,
        timeout: 300,
      });
      this.data.functions.push(func);
    });
    
    // Create CI/CD pipelines
    const pipelines = [
      {
        name: 'Microservices CI/CD',
        project: coreProject.id,
        traits: ['complex'],
      },
      {
        name: 'Infrastructure as Code',
        project: coreProject.id,
        traits: ['scheduled'],
      },
      {
        name: 'Security Compliance Check',
        project: coreProject.id,
        traits: ['scheduled'],
      },
    ];
    
    pipelines.forEach(pipelineConfig => {
      const pipeline = this.pipelineGenerator.withTraits(pipelineConfig.traits).generate({
        name: pipelineConfig.name,
        projectId: pipelineConfig.project,
        status: 'success',
      });
      this.data.pipelines.push(pipeline);
    });
    
    // Generate comprehensive metrics
    const metricsConfig = [
      { workloadType: 'web' as const, load: 'high' as const },
      { workloadType: 'database' as const, load: 'high' as const },
      { workloadType: 'batch' as const, load: 'medium' as const },
      { workloadType: 'cache' as const, load: 'high' as const },
    ];
    
    metricsConfig.forEach(config => {
      const resourceMetrics = this.metricsGenerator.generateResourceMetrics({
        duration: 604800, // 1 week
        interval: 300, // 5 minutes
        workloadType: config.workloadType,
      });
      
      const appMetrics = this.metricsGenerator.generateApplicationMetrics({
        duration: 604800,
        interval: 60,
        appType: 'api',
        load: config.load,
      });
      
      this.data.monitoring.metrics.push(resourceMetrics, appMetrics);
    });
    
    // Link all entities
    this.linkEntities();
    
    return this.data;
  }
}