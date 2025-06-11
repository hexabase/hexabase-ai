import { http, HttpResponse, delay } from 'msw';

// Mock projects data
export const mockProjects = [
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
  {
    id: 'proj-2',
    workspace_id: 'ws-1',
    name: 'backend-api',
    description: 'Backend API services',
    namespace: 'backend-namespace',
    resource_quotas: {
      'limits.cpu': '8',
      'limits.memory': '16Gi',
      'requests.storage': '100Gi',
    },
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
  },
];

export const projectHandlers = [
  // List projects
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects', async ({ params }) => {
    await delay(100);
    const projects = mockProjects.filter(p => p.workspace_id === params.workspaceId);
    
    return HttpResponse.json({
      projects,
      total: projects.length,
    });
  }),

  // Get project by ID
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects/:projectId', async ({ params }) => {
    await delay(100);
    const project = mockProjects.find(
      p => p.id === params.projectId && p.workspace_id === params.workspaceId
    );
    
    if (project) {
      return HttpResponse.json(project);
    }
    
    return HttpResponse.json(
      { error: 'Project not found' },
      { status: 404 }
    );
  }),

  // Create project
  http.post('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects', async ({ params, request }) => {
    const body = await request.json() as any;
    await delay(100);
    
    const newProject = {
      id: `proj-${Date.now()}`,
      workspace_id: params.workspaceId as string,
      ...body,
      namespace: body.namespace || `${body.name}-namespace`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    
    mockProjects.push(newProject);
    return HttpResponse.json(newProject, { status: 201 });
  }),

  // Update project
  http.put('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects/:projectId', async ({ params, request }) => {
    const body = await request.json() as any;
    await delay(100);
    
    const index = mockProjects.findIndex(
      p => p.id === params.projectId && p.workspace_id === params.workspaceId
    );
    
    if (index !== -1) {
      mockProjects[index] = {
        ...mockProjects[index],
        ...body,
        updated_at: new Date().toISOString(),
      };
      return HttpResponse.json(mockProjects[index]);
    }
    
    return HttpResponse.json(
      { error: 'Project not found' },
      { status: 404 }
    );
  }),

  // Delete project
  http.delete('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects/:projectId', async ({ params }) => {
    await delay(100);
    
    const index = mockProjects.findIndex(
      p => p.id === params.projectId && p.workspace_id === params.workspaceId
    );
    
    if (index !== -1) {
      mockProjects.splice(index, 1);
      return new HttpResponse(null, { status: 204 });
    }
    
    return HttpResponse.json(
      { error: 'Project not found' },
      { status: 404 }
    );
  }),
];