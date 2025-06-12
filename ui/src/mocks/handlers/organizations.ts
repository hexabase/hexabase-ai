import { http, HttpResponse, delay } from 'msw';

// Mock organizations data
export const mockOrganizations = [
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
];

export const organizationHandlers = [
  // List organizations
  http.get('/api/v1/organizations', async () => {
    await delay(100);
    return HttpResponse.json({
      organizations: mockOrganizations,
      total: mockOrganizations.length,
    });
  }),

  // Get organization by ID
  http.get('/api/v1/organizations/:orgId', async ({ params }) => {
    await delay(100);
    const org = mockOrganizations.find(o => o.id === params.orgId);
    
    if (org) {
      return HttpResponse.json(org);
    }
    
    return HttpResponse.json(
      { error: 'Organization not found' },
      { status: 404 }
    );
  }),

  // Create organization
  http.post('/api/v1/organizations', async ({ request }) => {
    const body = await request.json() as any;
    await delay(100);
    
    const newOrg = {
      id: `org-${Date.now()}`,
      ...body,
      owner_id: 'user-123',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    
    mockOrganizations.push(newOrg);
    return HttpResponse.json(newOrg, { status: 201 });
  }),

  // Update organization
  http.put('/api/v1/organizations/:orgId', async ({ params, request }) => {
    const body = await request.json() as any;
    await delay(100);
    
    const index = mockOrganizations.findIndex(o => o.id === params.orgId);
    
    if (index !== -1) {
      mockOrganizations[index] = {
        ...mockOrganizations[index],
        ...body,
        updated_at: new Date().toISOString(),
      };
      return HttpResponse.json(mockOrganizations[index]);
    }
    
    return HttpResponse.json(
      { error: 'Organization not found' },
      { status: 404 }
    );
  }),

  // Delete organization
  http.delete('/api/v1/organizations/:orgId', async ({ params }) => {
    await delay(100);
    
    const index = mockOrganizations.findIndex(o => o.id === params.orgId);
    
    if (index !== -1) {
      mockOrganizations.splice(index, 1);
      return new HttpResponse(null, { status: 204 });
    }
    
    return HttpResponse.json(
      { error: 'Organization not found' },
      { status: 404 }
    );
  }),
];