import { http, HttpResponse, delay } from 'msw';

// Mock workspaces data
export const mockWorkspaces = [
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
  {
    id: 'ws-2',
    organization_id: 'org-1',
    name: 'Development',
    description: 'Development environment',
    plan: 'shared',
    status: 'active',
    kubernetes_namespace: 'dev-namespace',
    resource_limits: {
      cpu: '4',
      memory: '8Gi',
      storage: '100Gi',
    },
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
  },
];

export const workspaceHandlers = [
  // List workspaces
  http.get('/api/v1/organizations/:orgId/workspaces', async ({ params }) => {
    await delay(100);
    const orgWorkspaces = mockWorkspaces.filter(w => w.organization_id === params.orgId);
    
    return HttpResponse.json({
      workspaces: orgWorkspaces,
      total: orgWorkspaces.length,
    });
  }),

  // Get workspace by ID
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId', async ({ params }) => {
    await delay(100);
    const workspace = mockWorkspaces.find(
      w => w.id === params.workspaceId && w.organization_id === params.orgId
    );
    
    if (workspace) {
      return HttpResponse.json(workspace);
    }
    
    return HttpResponse.json(
      { error: 'Workspace not found' },
      { status: 404 }
    );
  }),

  // Create workspace
  http.post('/api/v1/organizations/:orgId/workspaces', async ({ params, request }) => {
    const body = await request.json() as any;
    await delay(200); // Longer delay for creation
    
    const newWorkspace = {
      id: `ws-${Date.now()}`,
      organization_id: params.orgId as string,
      ...body,
      status: 'creating',
      kubernetes_namespace: `ns-${Date.now()}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    
    mockWorkspaces.push(newWorkspace);
    
    // Simulate async workspace creation
    setTimeout(() => {
      const ws = mockWorkspaces.find(w => w.id === newWorkspace.id);
      if (ws) {
        ws.status = 'active';
      }
    }, 2000);
    
    return HttpResponse.json(newWorkspace, { status: 201 });
  }),

  // Update workspace
  http.put('/api/v1/organizations/:orgId/workspaces/:workspaceId', async ({ params, request }) => {
    const body = await request.json() as any;
    await delay(100);
    
    const index = mockWorkspaces.findIndex(
      w => w.id === params.workspaceId && w.organization_id === params.orgId
    );
    
    if (index !== -1) {
      mockWorkspaces[index] = {
        ...mockWorkspaces[index],
        ...body,
        updated_at: new Date().toISOString(),
      };
      return HttpResponse.json(mockWorkspaces[index]);
    }
    
    return HttpResponse.json(
      { error: 'Workspace not found' },
      { status: 404 }
    );
  }),

  // Delete workspace
  http.delete('/api/v1/organizations/:orgId/workspaces/:workspaceId', async ({ params }) => {
    await delay(100);
    
    const index = mockWorkspaces.findIndex(
      w => w.id === params.workspaceId && w.organization_id === params.orgId
    );
    
    if (index !== -1) {
      mockWorkspaces.splice(index, 1);
      return new HttpResponse(null, { status: 204 });
    }
    
    return HttpResponse.json(
      { error: 'Workspace not found' },
      { status: 404 }
    );
  }),

  // Get workspace metrics
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/metrics', async () => {
    await delay(100);
    
    return HttpResponse.json({
      metrics: {
        cpu: { used: 2.5, total: 4, unit: 'cores' },
        memory: { used: 5.2, total: 8, unit: 'Gi' },
        storage: { used: 45, total: 100, unit: 'Gi' },
        pods: { running: 12, total: 20 },
      },
    });
  }),
];