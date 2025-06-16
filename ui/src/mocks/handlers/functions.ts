import { http, HttpResponse, delay } from 'msw';

// Mock functions data
export const mockFunctions = [
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
  {
    id: 'func-2',
    workspace_id: 'ws-1',
    project_id: 'proj-1',
    name: 'data-transformer',
    description: 'Transforms data between formats',
    runtime: 'python39',
    handler: 'main.handler',
    timeout: 60,
    memory: 512,
    environment_vars: {},
    triggers: ['schedule'],
    status: 'updating',
    version: 'v2.0.0',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

// Mock function versions
export const mockFunctionVersions: Record<string, Array<{
  version: string;
  deployed_at: string;
  deployed_by: string;
  status: string;
  size_bytes: number;
}>> = {
  'func-1': [
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
  ],
  'func-2': [
    {
      version: 'v2.0.0',
      deployed_at: '2024-01-01T00:00:00Z',
      deployed_by: 'user-123',
      status: 'active',
      size_bytes: 2097152,
    },
  ],
};

export const functionHandlers = [
  // List functions
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects/:projectId/functions', async ({ request }) => {
    await delay(100);
    
    const url = new URL(request.url);
    const runtime = url.searchParams.get('runtime');
    
    let funcs = [...mockFunctions];
    
    if (runtime && runtime !== 'all') {
      funcs = funcs.filter(f => f.runtime.includes(runtime));
    }
    
    return HttpResponse.json({
      functions: funcs,
      total: funcs.length,
    });
  }),

  // Get function by ID
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/functions/:functionId', async ({ params }) => {
    await delay(100);
    const func = mockFunctions.find(f => f.id === params.functionId);
    
    if (func) {
      return HttpResponse.json(func);
    }
    
    return HttpResponse.json(
      { error: 'Function not found' },
      { status: 404 }
    );
  }),

  // Create function
  http.post('/api/v1/organizations/:orgId/workspaces/:workspaceId/projects/:projectId/functions', async ({ request }) => {
    const formData = await request.formData();
    await delay(200);
    
    const newFunc = {
      id: `func-${Date.now()}`,
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: formData.get('name') as string,
      description: formData.get('description') as string,
      runtime: formData.get('runtime') as string,
      handler: formData.get('handler') as string,
      timeout: parseInt(formData.get('timeout') as string) || 30,
      memory: parseInt(formData.get('memory') as string) || 256,
      environment_vars: JSON.parse(formData.get('environment_vars') as string || '{}'),
      triggers: JSON.parse(formData.get('triggers') as string || '["http"]'),
      status: 'deploying',
      version: 'v1.0.0',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    
    mockFunctions.push(newFunc);
    
    // Simulate deployment
    setTimeout(() => {
      const func = mockFunctions.find(f => f.id === newFunc.id);
      if (func) {
        func.status = 'active';
        func.last_deployed_at = new Date().toISOString();
      }
    }, 2000);
    
    return HttpResponse.json(newFunc, { status: 201 });
  }),

  // Deploy function
  http.post('/api/v1/organizations/:orgId/workspaces/:workspaceId/functions/:functionId/deploy', async ({ params, request }) => {
    const formData = await request.formData();
    await delay(100);
    
    const func = mockFunctions.find(f => f.id === params.functionId);
    
    if (func) {
      const rollbackTo = formData.get('rollback_to') as string;
      
      if (rollbackTo) {
        // Rollback to previous version
        func.version = rollbackTo;
        func.status = 'active';
        func.last_deployed_at = new Date().toISOString();
        
        return HttpResponse.json({
          version: rollbackTo,
          status: 'active',
        });
      } else {
        // Deploy new version
        const version = formData.get('version') as string || `v${Date.now()}`;
        func.version = version;
        func.status = 'deploying';
        
        // Update environment vars if provided
        const envVars = formData.get('environment_vars') as string;
        if (envVars) {
          func.environment_vars = JSON.parse(envVars);
        }
        
        // Simulate deployment
        setTimeout(() => {
          func.status = 'active';
          func.last_deployed_at = new Date().toISOString();
        }, 1500);
        
        return HttpResponse.json({
          version,
          status: 'deploying',
        });
      }
    }
    
    return HttpResponse.json(
      { error: 'Function not found' },
      { status: 404 }
    );
  }),

  // Get function versions
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/functions/:functionId/versions', async ({ params }) => {
    await delay(100);
    
    const versions = mockFunctionVersions[params.functionId as string] || [];
    
    return HttpResponse.json({ versions });
  }),

  // Invoke function
  http.post('/api/v1/organizations/:orgId/workspaces/:workspaceId/functions/:functionId/invoke', async ({ params, request }) => {
    const body = await request.json() as any;
    await delay(150);
    
    const func = mockFunctions.find(f => f.id === params.functionId);
    
    if (func) {
      const invocationId = `inv-${Date.now()}`;
      
      // Simulate different responses based on payload
      if (body.payload?.error) {
        return HttpResponse.json({
          invocation_id: invocationId,
          function_id: params.functionId,
          status: 'error',
          trigger_type: body.trigger_type || 'http',
          error: 'Function execution failed',
          duration_ms: 45,
          logs: 'Error: Invalid input\n  at handler (index.js:10:5)',
          started_at: new Date().toISOString(),
          completed_at: new Date(Date.now() + 45).toISOString(),
        });
      }
      
      return HttpResponse.json({
        invocation_id: invocationId,
        function_id: params.functionId,
        status: 'success',
        trigger_type: body.trigger_type || 'http',
        payload: body.payload,
        output: { result: 'processed', timestamp: Date.now() },
        duration_ms: 150,
        logs: 'Function started\nProcessing input...\nCompleted successfully',
        started_at: new Date().toISOString(),
        completed_at: new Date(Date.now() + 150).toISOString(),
      });
    }
    
    return HttpResponse.json(
      { error: 'Function not found' },
      { status: 404 }
    );
  }),

  // Get function logs
  http.get('/api/v1/organizations/:orgId/workspaces/:workspaceId/functions/:functionId/logs', async ({ params, request }) => {
    await delay(100);
    
    const func = mockFunctions.find(f => f.id === params.functionId);
    
    if (func) {
      const logs = [
        `[2024-01-01T10:00:00Z] Function ${func.name} started`,
        '[2024-01-01T10:00:01Z] Processing request...',
        '[2024-01-01T10:00:02Z] Request completed successfully',
        '[2024-01-01T10:05:00Z] Function invoked via HTTP trigger',
        '[2024-01-01T10:05:01Z] Processed image: test.jpg',
        '[2024-01-01T10:05:02Z] Upload complete',
      ];
      
      return HttpResponse.json({
        logs,
        total: logs.length,
      });
    }
    
    return HttpResponse.json(
      { error: 'Function not found' },
      { status: 404 }
    );
  }),

  // Delete function
  http.delete('/api/v1/organizations/:orgId/workspaces/:workspaceId/functions/:functionId', async ({ params }) => {
    await delay(100);
    
    const index = mockFunctions.findIndex(f => f.id === params.functionId);
    
    if (index !== -1) {
      mockFunctions.splice(index, 1);
      return new HttpResponse(null, { status: 204 });
    }
    
    return HttpResponse.json(
      { error: 'Function not found' },
      { status: 404 }
    );
  }),
];