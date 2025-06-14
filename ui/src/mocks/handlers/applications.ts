import { http, HttpResponse, delay } from 'msw';
import type { Application, CronJobExecution, CreateApplicationRequest } from '@/lib/api-client';

// Mock applications data
export const mockApplications: Application[] = [
  {
    id: 'app-1',
    workspace_id: 'ws-1',
    project_id: 'proj-1',
    name: 'web-frontend',
    type: 'stateless',
    status: 'running' as const,
    source_type: 'image',
    source_image: 'nginx:latest',
    config: {
      replicas: 3,
      port: 80,
      resources: {
        requests: { cpu: '100m', memory: '128Mi' },
        limits: { cpu: '500m', memory: '512Mi' },
      },
      environment_vars: {
        API_URL: 'https://api.example.com',
      },
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 'app-2',
    workspace_id: 'ws-1',
    project_id: 'proj-1',
    name: 'database',
    type: 'stateful',
    status: 'running' as const,
    source_type: 'image',
    source_image: 'postgres:14',
    config: {
      replicas: 1,
      port: 5432,
      storage_size: '10Gi',
      storage_class: 'fast-ssd',
    },
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
  },
  {
    id: 'cron-1',
    workspace_id: 'ws-1',
    project_id: 'proj-1',
    name: 'Daily Backup',
    type: 'cronjob',
    status: 'active' as const,
    source_type: 'image',
    source_image: 'backup:latest',
    cron_schedule: '0 2 * * *',
    last_execution_at: '2024-01-01T02:00:00Z',
    next_execution_at: '2024-01-02T02:00:00Z',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

// Mock CronJob executions
export const mockCronJobExecutions: CronJobExecution[] = [
  {
    id: 'cje-1',
    application_id: 'cron-1',
    job_name: 'daily-backup-12345',
    started_at: '2024-01-01T02:00:00Z',
    completed_at: '2024-01-01T02:05:00Z',
    status: 'succeeded' as const,
    exit_code: 0,
    logs: 'Backup completed successfully\n10GB backed up',
    created_at: '2024-01-01T02:00:00Z',
    updated_at: '2024-01-01T02:05:00Z',
  },
  {
    id: 'cje-2',
    application_id: 'cron-1',
    job_name: 'daily-backup-12346',
    started_at: '2023-12-31T02:00:00Z',
    completed_at: '2023-12-31T02:03:00Z',
    status: 'failed' as const,
    exit_code: 1,
    logs: 'Error: Connection timeout\nFailed to connect to backup server',
    created_at: '2023-12-31T02:00:00Z',
    updated_at: '2023-12-31T02:03:00Z',
  },
];

export const applicationHandlers = [
  // List applications
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects/:projectId/applications', async ({ request }) => {
    await delay(100);
    
    const url = new URL(request.url);
    const type = url.searchParams.get('type');
    const status = url.searchParams.get('status');
    
    let apps = [...mockApplications];
    
    if (type) {
      apps = apps.filter(app => app.type === type);
    }
    
    if (status) {
      apps = apps.filter(app => app.status === status);
    }
    
    return HttpResponse.json({
      applications: apps,
      total: apps.length,
    });
  }),

  // Get application by ID
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/applications/:appId', async ({ params }) => {
    await delay(100);
    const app = mockApplications.find(a => a.id === params.appId);
    
    if (app) {
      return HttpResponse.json(app);
    }
    
    return HttpResponse.json(
      { error: 'Application not found' },
      { status: 404 }
    );
  }),

  // Create application
  http.post('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects/:projectId/applications', async ({ request }) => {
    const body = await request.json() as CreateApplicationRequest;
    await delay(200);
    
    const newApp: Application = {
      id: `app-${Date.now()}`,
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      ...body,
      status: 'pending' as const,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    
    mockApplications.push(newApp);
    
    // Simulate async deployment
    setTimeout(() => {
      const app = mockApplications.find(a => a.id === newApp.id);
      if (app) {
        app.status = 'running' as const;
      }
    }, 2000);
    
    return HttpResponse.json(newApp, { status: 201 });
  }),

  // Update application status
  http.patch('/api/v1/organizations/:orgId/workspaces/:workspaceId/applications/:appId', async ({ params, request }) => {
    const body = await request.json() as { status: Application['status'] };
    await delay(100);
    
    const app = mockApplications.find(a => a.id === params.appId);
    
    if (app) {
      app.status = body.status;
      app.updated_at = new Date().toISOString();
      return HttpResponse.json(app);
    }
    
    return HttpResponse.json(
      { error: 'Application not found' },
      { status: 404 }
    );
  }),

  // Delete application
  http.delete('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects/:projectId/applications/:appId', async ({ params }) => {
    await delay(100);
    
    const index = mockApplications.findIndex(a => a.id === params.appId);
    
    if (index !== -1) {
      mockApplications.splice(index, 1);
      return new HttpResponse(null, { status: 204 });
    }
    
    return HttpResponse.json(
      { error: 'Application not found' },
      { status: 404 }
    );
  }),

  // CronJob specific endpoints
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/applications/:appId/cronjob/executions', async ({ params }) => {
    await delay(100);
    
    const executions = mockCronJobExecutions.filter(e => e.application_id === params.appId);
    
    return HttpResponse.json({
      executions,
      total: executions.length,
    });
  }),

  http.post('/api/v1/organizations/:orgId/workspaces/:workspaceId/applications/:appId/cronjob/trigger', async ({ params }) => {
    await delay(100);
    
    const app = mockApplications.find(a => a.id === params.appId);
    
    if (app && app.type === 'cronjob') {
      const newExecution = {
        id: `cje-${Date.now()}`,
        application_id: params.appId as string,
        job_name: `manual-trigger-${Date.now()}`,
        started_at: new Date().toISOString(),
        status: 'running' as const,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      
      mockCronJobExecutions.push(newExecution);
      
      return HttpResponse.json({
        execution_id: newExecution.id,
        message: 'CronJob triggered successfully',
      });
    }
    
    return HttpResponse.json(
      { error: 'Application not found or not a CronJob' },
      { status: 400 }
    );
  }),

  http.put('/api/v1/organizations/:orgId/workspaces/:workspaceId/applications/:appId/cronjob/schedule', async ({ params, request }) => {
    const body = await request.json() as { schedule: string };
    await delay(100);
    
    const app = mockApplications.find(a => a.id === params.appId);
    
    if (app && app.type === 'cronjob') {
      app.cron_schedule = body.schedule;
      app.next_execution_at = new Date(Date.now() + 86400000).toISOString(); // Next day
      app.updated_at = new Date().toISOString();
      
      return HttpResponse.json({
        ...app,
        next_execution_at: app.next_execution_at,
      });
    }
    
    return HttpResponse.json(
      { error: 'Application not found or not a CronJob' },
      { status: 400 }
    );
  }),
];