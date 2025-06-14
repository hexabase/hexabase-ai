import {
  OrganizationGenerator,
  WorkspaceGenerator,
  ProjectGenerator,
  ApplicationGenerator,
  UserGenerator,
  DeploymentGenerator,
  CronJobGenerator,
  FunctionGenerator,
  BackupGenerator,
  MetricsGenerator,
  AlertGenerator,
  PipelineGenerator,
} from '../generators';

import {
  Organization,
  Workspace,
  Project,
  Application,
  User,
  Deployment,
  CronJob,
  ServerlessFunction,
  Backup,
  BackupPolicy,
  BackupStorage,
  Alert,
  Pipeline,
} from '../generators';

export interface ScenarioData {
  organizations: Organization[];
  workspaces: Workspace[];
  projects: Project[];
  applications: Application[];
  users: User[];
  deployments: Deployment[];
  cronjobs: CronJob[];
  functions: ServerlessFunction[];
  backups: {
    storages: BackupStorage[];
    policies: BackupPolicy[];
    backups: Backup[];
  };
  monitoring: {
    alerts: Alert[];
    metrics: any[];
  };
  pipelines: Pipeline[];
}

export abstract class BaseScenario {
  protected orgGenerator = new OrganizationGenerator();
  protected workspaceGenerator = new WorkspaceGenerator();
  protected projectGenerator = new ProjectGenerator();
  protected appGenerator = new ApplicationGenerator();
  protected userGenerator = new UserGenerator();
  protected deploymentGenerator = new DeploymentGenerator();
  protected cronJobGenerator = new CronJobGenerator();
  protected functionGenerator = new FunctionGenerator();
  protected backupGenerator = new BackupGenerator();
  protected metricsGenerator = new MetricsGenerator();
  protected alertGenerator = new AlertGenerator();
  protected pipelineGenerator = new PipelineGenerator();
  
  protected data: ScenarioData = {
    organizations: [],
    workspaces: [],
    projects: [],
    applications: [],
    users: [],
    deployments: [],
    cronjobs: [],
    functions: [],
    backups: {
      storages: [],
      policies: [],
      backups: [],
    },
    monitoring: {
      alerts: [],
      metrics: [],
    },
    pipelines: [],
  };
  
  constructor(protected seed?: number) {
    if (seed) {
      this.setSeed(seed);
    }
  }
  
  abstract generate(): ScenarioData;
  
  protected setSeed(seed: number) {
    this.orgGenerator.resetSeed(seed);
    this.workspaceGenerator.resetSeed(seed);
    this.projectGenerator.resetSeed(seed);
    this.appGenerator.resetSeed(seed);
    this.userGenerator.resetSeed(seed);
    this.deploymentGenerator.resetSeed(seed);
    this.cronJobGenerator.resetSeed(seed);
    this.functionGenerator.resetSeed(seed);
    this.backupGenerator.resetSeed(seed);
    this.metricsGenerator.resetSeed(seed);
    this.alertGenerator.resetSeed(seed);
    this.pipelineGenerator.resetSeed(seed);
  }
  
  /**
   * Link related entities together
   */
  protected linkEntities() {
    // Link users to organizations
    this.data.users.forEach((user, index) => {
      if (user.organizations.length === 0 && this.data.organizations.length > 0) {
        const org = this.data.organizations[index % this.data.organizations.length];
        user.organizations.push({
          id: org.id,
          name: org.name,
          role: index === 0 ? 'owner' : 'member',
        });
      }
    });
    
    // Link workspaces to organizations
    this.data.workspaces.forEach((workspace, index) => {
      if (!workspace.organizationId && this.data.organizations.length > 0) {
        workspace.organizationId = this.data.organizations[
          index % this.data.organizations.length
        ].id;
      }
    });
    
    // Link projects to workspaces
    this.data.projects.forEach((project, index) => {
      if (!project.workspaceId && this.data.workspaces.length > 0) {
        project.workspaceId = this.data.workspaces[
          index % this.data.workspaces.length
        ].id;
      }
    });
    
    // Link applications to projects
    this.data.applications.forEach((app, index) => {
      if (!app.projectId && this.data.projects.length > 0) {
        app.projectId = this.data.projects[
          index % this.data.projects.length
        ].id;
      }
    });
    
    // Link deployments to applications
    this.data.deployments.forEach((deployment, index) => {
      if (!deployment.applicationId && this.data.applications.length > 0) {
        deployment.applicationId = this.data.applications[
          index % this.data.applications.length
        ].id;
      }
    });
    
    // Link cronjobs to projects
    this.data.cronjobs.forEach((job, index) => {
      if (!job.projectId && this.data.projects.length > 0) {
        job.projectId = this.data.projects[
          index % this.data.projects.length
        ].id;
      }
    });
    
    // Link functions to projects
    this.data.functions.forEach((func, index) => {
      if (!func.projectId && this.data.projects.length > 0) {
        func.projectId = this.data.projects[
          index % this.data.projects.length
        ].id;
      }
    });
    
    // Link backups to workspaces
    this.data.backups.storages.forEach((storage, index) => {
      if (!storage.workspaceId && this.data.workspaces.length > 0) {
        storage.workspaceId = this.data.workspaces[
          index % this.data.workspaces.length
        ].id;
      }
    });
    
    this.data.backups.policies.forEach((policy, index) => {
      if (!policy.workspaceId && this.data.workspaces.length > 0) {
        policy.workspaceId = this.data.workspaces[
          index % this.data.workspaces.length
        ].id;
      }
      if (!policy.storageId && this.data.backups.storages.length > 0) {
        policy.storageId = this.data.backups.storages[
          index % this.data.backups.storages.length
        ].id;
      }
    });
    
    this.data.backups.backups.forEach((backup, index) => {
      if (!backup.workspaceId && this.data.workspaces.length > 0) {
        backup.workspaceId = this.data.workspaces[
          index % this.data.workspaces.length
        ].id;
      }
      if (!backup.policyId && this.data.backups.policies.length > 0) {
        backup.policyId = this.data.backups.policies[
          index % this.data.backups.policies.length
        ].id;
      }
    });
    
    // Link pipelines to projects
    this.data.pipelines.forEach((pipeline, index) => {
      if (!pipeline.projectId && this.data.projects.length > 0) {
        pipeline.projectId = this.data.projects[
          index % this.data.projects.length
        ].id;
      }
    });
  }
  
  /**
   * Get data subset for a specific workspace
   */
  getWorkspaceData(workspaceId: string): Partial<ScenarioData> {
    const workspace = this.data.workspaces.find(w => w.id === workspaceId);
    if (!workspace) return {};
    
    const projects = this.data.projects.filter(p => p.workspaceId === workspaceId);
    const projectIds = projects.map(p => p.id);
    
    return {
      workspaces: [workspace],
      projects,
      applications: this.data.applications.filter(a => projectIds.includes(a.projectId)),
      cronjobs: this.data.cronjobs.filter(j => projectIds.includes(j.projectId)),
      functions: this.data.functions.filter(f => projectIds.includes(f.projectId)),
      pipelines: this.data.pipelines.filter(p => projectIds.includes(p.projectId)),
    };
  }
  
  /**
   * Export scenario data as JSON
   */
  toJSON(): string {
    return JSON.stringify(this.data, null, 2);
  }
  
  /**
   * Export scenario data for API mocking
   */
  toMockAPI() {
    return {
      '/api/organizations': this.data.organizations,
      '/api/workspaces': this.data.workspaces,
      '/api/projects': this.data.projects,
      '/api/applications': this.data.applications,
      '/api/users': this.data.users,
      '/api/deployments': this.data.deployments,
      '/api/cronjobs': this.data.cronjobs,
      '/api/functions': this.data.functions,
      '/api/backups/storages': this.data.backups.storages,
      '/api/backups/policies': this.data.backups.policies,
      '/api/backups': this.data.backups.backups,
      '/api/alerts': this.data.monitoring.alerts,
      '/api/pipelines': this.data.pipelines,
    };
  }
}