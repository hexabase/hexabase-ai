/**
 * Comprehensive mock API client for testing
 * This provides realistic responses without needing actual API calls
 */

import { 
  Organization, 
  Workspace, 
  Project, 
  Application,
  CronJobExecution,
  FunctionConfig,
  FunctionInvocation,
  FunctionVersion,
  CreateOrganizationRequest,
  CreateWorkspaceRequest,
  CreateProjectRequest,
  CreateApplicationRequest,
  CreateFunctionRequest,
  DeployFunctionRequest,
} from '@/lib/api-client';

// Mock data stores
const mockData = {
  organizations: [
    {
      id: 'org-1',
      name: 'Acme Corporation',
      description: 'Enterprise software solutions',
      owner_id: 'user-123',
      billing_email: 'billing@acme.com',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 'org-2',
      name: 'Tech Startup',
      description: 'Innovative tech solutions',
      owner_id: 'user-123',
      billing_email: 'billing@techstartup.com',
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
    },
  ] as Organization[],
  
  workspaces: [
    {
      id: 'ws-1',
      organization_id: 'org-1',
      name: 'Production',
      description: 'Production environment',
      plan: 'dedicated',
      status: 'active',
      kubernetes_namespace: 'prod-namespace',
      resource_limits: {
        cpu: '16',
        memory: '32Gi',
        storage: '500Gi',
      },
      node_pool: 'dedicated-pool-1',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
  ] as Workspace[],
  
  projects: [
    {
      id: 'proj-1',
      workspace_id: 'ws-1',
      name: 'frontend-app',
      description: 'Main frontend application',
      namespace: 'frontend-namespace',
      resource_quotas: {
        'limits.cpu': '4',
        'limits.memory': '8Gi',
        'requests.storage': '50Gi',
      },
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
  ] as Project[],
  
  applications: [
    {
      id: 'app-1',
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: 'web-frontend',
      type: 'stateless',
      status: 'running',
      source_type: 'image',
      source_image: 'nginx:latest',
      config: {
        replicas: 3,
        port: 80,
      },
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 'cron-1',
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: 'Daily Backup',
      type: 'cronjob',
      status: 'active',
      source_type: 'image',
      source_image: 'backup:latest',
      cron_schedule: '0 2 * * *',
      last_execution_at: '2024-01-01T02:00:00Z',
      next_execution_at: '2024-01-02T02:00:00Z',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
  ] as Application[],
  
  cronJobExecutions: [
    {
      id: 'cje-1',
      application_id: 'cron-1',
      job_name: 'backup-job-12345',
      started_at: '2024-01-01T10:00:00Z',
      completed_at: '2024-01-01T10:05:00Z',
      status: 'succeeded',
      exit_code: 0,
      logs: 'Backup completed successfully',
      created_at: '2024-01-01T10:00:00Z',
      updated_at: '2024-01-01T10:05:00Z',
    },
  ] as CronJobExecution[],
  
  functions: [
    {
      id: 'func-1',
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: 'image-processor',
      description: 'Processes uploaded images',
      runtime: 'nodejs18',
      handler: 'index.handler',
      timeout: 30,
      memory: 256,
      environment_vars: {
        IMAGE_BUCKET: 'images',
      },
      triggers: ['http', 'event'],
      status: 'active',
      version: 'v1.2.0',
      last_deployed_at: '2024-01-01T10:00:00Z',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T10:00:00Z',
    },
  ] as FunctionConfig[],
  
  functionVersions: [
    {
      version: 'v1.2.0',
      deployed_at: '2024-01-01T10:00:00Z',
      deployed_by: 'user-123',
      status: 'active',
      size_bytes: 1048576,
    },
    {
      version: 'v1.1.0',
      deployed_at: '2023-12-01T10:00:00Z',
      deployed_by: 'user-123',
      status: 'inactive',
      size_bytes: 1024000,
    },
  ] as FunctionVersion[],
};

// Helper to simulate API delay
const delay = (ms: number = 100) => new Promise(resolve => setTimeout(resolve, ms));

// Mock API client implementation
export const createMockApiClient = () => ({
  auth: {
    login: jest.fn().mockImplementation(async () => {
      await delay();
      return {
        access_token: 'mock-jwt-token',
        refresh_token: 'mock-refresh-token',
        user: {
          id: 'user-123',
          email: 'test@example.com',
          name: 'Test User',
        },
      };
    }),
    me: jest.fn().mockResolvedValue({
      user: {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
        organizations: ['org-1', 'org-2'],
      },
    }),
    logout: jest.fn().mockResolvedValue({}),
    refreshToken: jest.fn().mockResolvedValue({
      access_token: 'mock-jwt-token-refreshed',
    }),
  },
  
  organizations: {
    list: jest.fn().mockImplementation(async () => {
      await delay();
      return {
        organizations: mockData.organizations,
        total: mockData.organizations.length,
      };
    }),
    get: jest.fn().mockImplementation(async (id: string) => {
      await delay();
      const org = mockData.organizations.find(o => o.id === id);
      if (!org) throw new Error('Organization not found');
      return org;
    }),
    create: jest.fn().mockImplementation(async (data: CreateOrganizationRequest) => {
      await delay();
      const newOrg = {
        id: `org-${Date.now()}`,
        ...data,
        owner_id: 'user-123',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      mockData.organizations.push(newOrg);
      return newOrg;
    }),
    update: jest.fn().mockImplementation(async (id: string, data: Partial<Organization>) => {
      await delay();
      const index = mockData.organizations.findIndex(o => o.id === id);
      if (index === -1) throw new Error('Organization not found');
      mockData.organizations[index] = { ...mockData.organizations[index], ...data };
      return mockData.organizations[index];
    }),
    delete: jest.fn().mockImplementation(async (id: string) => {
      await delay();
      const index = mockData.organizations.findIndex(o => o.id === id);
      if (index === -1) throw new Error('Organization not found');
      mockData.organizations.splice(index, 1);
    }),
  },
  
  workspaces: {
    list: jest.fn().mockImplementation(async (orgId: string) => {
      await delay();
      const workspaces = mockData.workspaces.filter(w => w.organization_id === orgId);
      return {
        workspaces,
        total: workspaces.length,
      };
    }),
    get: jest.fn().mockImplementation(async (orgId: string, id: string) => {
      await delay();
      const workspace = mockData.workspaces.find(w => w.id === id && w.organization_id === orgId);
      if (!workspace) throw new Error('Workspace not found');
      return workspace;
    }),
    create: jest.fn().mockImplementation(async (orgId: string, data: CreateWorkspaceRequest) => {
      await delay(200);
      const newWorkspace = {
        id: `ws-${Date.now()}`,
        organization_id: orgId,
        ...data,
        status: 'creating',
        kubernetes_namespace: `ns-${Date.now()}`,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      mockData.workspaces.push(newWorkspace);
      setTimeout(() => {
        newWorkspace.status = 'active';
      }, 1000);
      return newWorkspace;
    }),
    update: jest.fn().mockImplementation(async (orgId: string, id: string, data: Partial<Workspace>) => {
      await delay();
      const index = mockData.workspaces.findIndex(w => w.id === id && w.organization_id === orgId);
      if (index === -1) throw new Error('Workspace not found');
      mockData.workspaces[index] = { ...mockData.workspaces[index], ...data };
      return mockData.workspaces[index];
    }),
    delete: jest.fn().mockImplementation(async (orgId: string, id: string) => {
      await delay();
      const index = mockData.workspaces.findIndex(w => w.id === id && w.organization_id === orgId);
      if (index === -1) throw new Error('Workspace not found');
      mockData.workspaces.splice(index, 1);
    }),
  },
  
  projects: {
    list: jest.fn().mockImplementation(async (orgId: string, workspaceId: string) => {
      await delay();
      const projects = mockData.projects.filter(p => p.workspace_id === workspaceId);
      return {
        projects,
        total: projects.length,
      };
    }),
    get: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string) => {
      await delay();
      const project = mockData.projects.find(p => p.id === id && p.workspace_id === workspaceId);
      if (!project) throw new Error('Project not found');
      return project;
    }),
    create: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, data: CreateProjectRequest) => {
      await delay();
      const newProject = {
        id: `proj-${Date.now()}`,
        workspace_id: workspaceId,
        ...data,
        namespace: data.namespace || `${data.name}-namespace`,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      mockData.projects.push(newProject);
      return newProject;
    }),
    update: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string, data: Partial<Project>) => {
      await delay();
      const index = mockData.projects.findIndex(p => p.id === id && p.workspace_id === workspaceId);
      if (index === -1) throw new Error('Project not found');
      mockData.projects[index] = { ...mockData.projects[index], ...data };
      return mockData.projects[index];
    }),
    delete: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string) => {
      await delay();
      const index = mockData.projects.findIndex(p => p.id === id && p.workspace_id === workspaceId);
      if (index === -1) throw new Error('Project not found');
      mockData.projects.splice(index, 1);
    }),
  },
  
  applications: {
    list: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, projectId?: string, params?: any) => {
      await delay();
      let apps = [...mockData.applications];
      if (projectId) {
        apps = apps.filter(a => a.project_id === projectId);
      }
      if (params?.type) {
        apps = apps.filter(a => a.type === params.type);
      }
      if (params?.status) {
        apps = apps.filter(a => a.status === params.status);
      }
      return {
        applications: apps,
        total: apps.length,
      };
    }),
    get: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string) => {
      await delay();
      const app = mockData.applications.find(a => a.id === id);
      if (!app) throw new Error('Application not found');
      return app;
    }),
    create: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, projectId: string, data: CreateApplicationRequest) => {
      await delay(200);
      const newApp = {
        id: `app-${Date.now()}`,
        workspace_id: workspaceId,
        project_id: projectId,
        ...data,
        status: 'creating',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      mockData.applications.push(newApp);
      setTimeout(() => {
        newApp.status = 'running';
      }, 1000);
      return newApp;
    }),
    updateStatus: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string, data: { status: string }) => {
      await delay();
      const app = mockData.applications.find(a => a.id === id);
      if (!app) throw new Error('Application not found');
      app.status = data.status;
      return { data: app };
    }),
    delete: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, projectId: string, id: string) => {
      await delay();
      const index = mockData.applications.findIndex(a => a.id === id);
      if (index === -1) throw new Error('Application not found');
      mockData.applications.splice(index, 1);
    }),
    getCronJobExecutions: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, appId: string) => {
      await delay();
      const executions = mockData.cronJobExecutions.filter(e => e.application_id === appId);
      return {
        executions,
        total: executions.length,
      };
    }),
    triggerCronJob: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, appId: string) => {
      await delay();
      const app = mockData.applications.find(a => a.id === appId);
      if (!app || app.type !== 'cronjob') throw new Error('CronJob not found');
      
      const newExecution = {
        id: `cje-${Date.now()}`,
        application_id: appId,
        job_name: `manual-trigger-${Date.now()}`,
        started_at: new Date().toISOString(),
        status: 'running',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      mockData.cronJobExecutions.push(newExecution);
      
      return {
        execution_id: newExecution.id,
        message: 'CronJob triggered successfully',
      };
    }),
    updateCronSchedule: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, appId: string, data: { schedule: string }) => {
      await delay();
      const app = mockData.applications.find(a => a.id === appId);
      if (!app || app.type !== 'cronjob') throw new Error('CronJob not found');
      
      app.cron_schedule = data.schedule;
      app.next_execution_at = new Date(Date.now() + 86400000).toISOString();
      
      return {
        ...app,
        next_execution_at: app.next_execution_at,
      };
    }),
  },
  
  monitoring: {
    getWorkspaceMetrics: jest.fn().mockResolvedValue({
      metrics: {
        cpu_usage: 45.5,
        memory_usage: 62.3,
        storage_usage: 30.1,
        network_ingress: 1024 * 1024 * 100,
        network_egress: 1024 * 1024 * 200,
        pod_count: 15,
        container_count: 25,
        timestamp: '2024-01-01T00:00:00Z',
      }
    }),
    getResourceUsageHistory: jest.fn().mockResolvedValue({
      history: {
        cpu: [
          { timestamp: '2024-01-01T00:00:00Z', value: 40 },
          { timestamp: '2024-01-01T00:05:00Z', value: 45 },
          { timestamp: '2024-01-01T00:10:00Z', value: 45.5 },
        ],
        memory: [
          { timestamp: '2024-01-01T00:00:00Z', value: 60 },
          { timestamp: '2024-01-01T00:05:00Z', value: 61 },
          { timestamp: '2024-01-01T00:10:00Z', value: 62.3 },
        ],
      }
    }),
    getApplicationMetrics: jest.fn().mockResolvedValue({
      metrics: []
    }),
  },
  
  aiops: {
    chat: jest.fn().mockResolvedValue({
      message: 'Based on your current resource usage, everything looks healthy.',
      suggestions: [
        'Consider enabling autoscaling for better resource efficiency',
        'Your memory usage pattern suggests you could reduce allocation by 20%',
      ],
      context: {
        analyzed_metrics: {
          cpu_usage: 45,
          memory_usage: 60,
        },
      },
    }),
    getSuggestions: jest.fn().mockResolvedValue({
      suggestions: [
        {
          id: 'sug-1',
          type: 'optimization',
          title: 'Enable HPA for Frontend App',
          description: 'Your frontend app shows variable load patterns',
          priority: 'medium',
          estimated_savings: '~20% resource cost',
        },
        {
          id: 'sug-2',
          type: 'security',
          title: 'Update container images',
          description: '3 containers are running outdated images',
          priority: 'high',
        },
      ],
    }),
    analyzeMetrics: jest.fn().mockResolvedValue({
      analysis: {
        summary: 'Your infrastructure is running efficiently',
        findings: [
          {
            type: 'optimization',
            description: 'CPU usage is optimal at 45%',
            severity: 'info',
          },
          {
            type: 'warning',
            description: 'Memory usage trending upward',
            severity: 'warning',
          },
        ],
        recommendations: [
          'Consider implementing memory limits',
          'Enable monitoring alerts for memory > 80%',
        ],
      },
    }),
  },
  
  backup: {
    listStorages: jest.fn().mockResolvedValue({
      storages: [
        {
          id: 'bs-1',
          workspace_id: 'ws-123',
          name: 'Primary Backup Storage',
          type: 'proxmox',
          status: 'active',
          capacity_gb: 1000,
          used_gb: 250,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
    }),
    createStorage: jest.fn().mockResolvedValue({
      id: 'bs-new',
      workspace_id: 'ws-123',
      name: 'Secondary Storage',
      type: 'proxmox',
      status: 'active',
      capacity_gb: 500,
      used_gb: 0,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }),
    deleteStorage: jest.fn().mockResolvedValue({}),
    listPolicies: jest.fn().mockResolvedValue({
      policies: [
        {
          id: 'bp-1',
          workspace_id: 'ws-123',
          name: 'Daily Backup',
          storage_id: 'bs-1',
          schedule: '0 2 * * *',
          retention_days: 30,
          backup_type: 'full',
          enabled: true,
          last_execution: '2024-01-01T02:00:00Z',
          next_execution: '2024-01-02T02:00:00Z',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
    }),
    createPolicy: jest.fn().mockResolvedValue({
      id: 'bp-new',
      workspace_id: 'ws-123',
      name: 'Weekly Backup',
      storage_id: 'bs-1',
      schedule: '0 3 * * 0',
      retention_days: 90,
      backup_type: 'incremental',
      enabled: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }),
    updatePolicy: jest.fn().mockResolvedValue({}),
    deletePolicy: jest.fn().mockResolvedValue({}),
    listExecutions: jest.fn().mockResolvedValue({
      executions: [
        {
          id: 'be-1',
          policy_id: 'bp-1',
          status: 'completed',
          started_at: '2024-01-01T02:00:00Z',
          completed_at: '2024-01-01T02:30:00Z',
          size_bytes: 1024 * 1024 * 500,
          error: null,
        },
        {
          id: 'be-2',
          policy_id: 'bp-1',
          status: 'failed',
          started_at: '2024-01-02T02:00:00Z',
          completed_at: '2024-01-02T02:05:00Z',
          size_bytes: 0,
          error: 'Storage quota exceeded',
        },
      ],
      total: 2,
    }),
    triggerBackup: jest.fn().mockResolvedValue({
      execution_id: 'be-manual',
      status: 'running',
      message: 'Backup started successfully',
    }),
    getBackupDetails: jest.fn().mockResolvedValue({
      backup: {
        id: 'be-1',
        policy_name: 'Daily Backup',
        size_bytes: 1024 * 1024 * 500,
        created_at: '2024-01-01T02:30:00Z',
        includes: ['applications', 'configurations', 'volumes'],
      },
    }),
    restoreBackup: jest.fn().mockResolvedValue({
      restore_id: 'res-123',
      status: 'in_progress',
      message: 'Restore initiated',
    }),
  },
  
  functions: {
    list: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, projectId?: string, params?: any) => {
      await delay();
      let funcs = [...mockData.functions];
      if (params?.runtime && params.runtime !== 'all') {
        funcs = funcs.filter(f => f.runtime.includes(params.runtime));
      }
      return {
        functions: funcs,
        total: funcs.length,
      };
    }),
    get: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string) => {
      await delay();
      const func = mockData.functions.find(f => f.id === id);
      if (!func) throw new Error('Function not found');
      return func;
    }),
    create: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, projectId: string, data: CreateFunctionRequest) => {
      await delay(200);
      const newFunc = {
        id: `func-${Date.now()}`,
        workspace_id: workspaceId,
        project_id: projectId,
        ...data,
        status: 'deploying',
        version: 'v1.0.0',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      mockData.functions.push(newFunc);
      setTimeout(() => {
        newFunc.status = 'active';
      }, 1000);
      return newFunc;
    }),
    deploy: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string, data: DeployFunctionRequest) => {
      await delay();
      const func = mockData.functions.find(f => f.id === id);
      if (!func) throw new Error('Function not found');
      
      if (data.rollback_to) {
        func.version = data.rollback_to;
        func.status = 'active';
        return {
          version: data.rollback_to,
          status: 'active',
        };
      }
      
      const version = data.version || `v${Date.now()}`;
      func.version = version;
      func.status = 'deploying';
      setTimeout(() => {
        func.status = 'active';
      }, 1000);
      
      return {
        version,
        status: 'deploying',
      };
    }),
    invoke: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string, data: any) => {
      await delay(150);
      
      if (data.payload?.error) {
        return {
          invocation_id: `inv-${Date.now()}`,
          function_id: id,
          status: 'error',
          trigger_type: data.trigger_type || 'http',
          error: 'Function execution failed',
          duration_ms: 45,
          logs: 'Error: Invalid input',
          started_at: new Date().toISOString(),
          completed_at: new Date(Date.now() + 45).toISOString(),
        } as FunctionInvocation;
      }
      
      return {
        invocation_id: `inv-${Date.now()}`,
        function_id: id,
        status: 'success',
        trigger_type: data.trigger_type || 'http',
        payload: data.payload,
        output: { result: 'processed' },
        duration_ms: 150,
        logs: 'Function executed successfully',
        started_at: new Date().toISOString(),
        completed_at: new Date(Date.now() + 150).toISOString(),
      } as FunctionInvocation;
    }),
    getVersions: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string) => {
      await delay();
      return {
        versions: mockData.functionVersions,
      };
    }),
    getLogs: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string) => {
      await delay();
      return {
        logs: [
          '[2024-01-01T10:00:00Z] Function started',
          '[2024-01-01T10:00:01Z] Processing request...',
          '[2024-01-01T10:00:02Z] Request completed successfully',
        ],
        total: 3,
      };
    }),
    delete: jest.fn().mockImplementation(async (orgId: string, workspaceId: string, id: string) => {
      await delay();
      const index = mockData.functions.findIndex(f => f.id === id);
      if (index === -1) throw new Error('Function not found');
      mockData.functions.splice(index, 1);
    }),
  },
});

// Export a pre-configured instance
export const mockApiClient = createMockApiClient();