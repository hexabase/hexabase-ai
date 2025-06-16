import { Page, Route } from '@playwright/test';
import {
  testUsers,
  testOrganizations,
  testWorkspaces,
  testProjects,
  testApplications,
  deploymentStages,
  testMetrics,
} from '../fixtures/mock-data';

/**
 * Sets up mock API responses for E2E tests
 */
export async function setupMockAPI(page: Page) {
  // Auth endpoints
  await page.route('**/api/auth/login', async (route) => {
    const request = route.request();
    const body = await request.postDataJSON();
    
    const user = Object.values(testUsers).find(
      u => u.email === body.email && u.password === body.password
    );
    
    if (user) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          access_token: `mock-jwt-${user.id}`,
          refresh_token: `mock-refresh-${user.id}`,
          user: {
            id: user.id,
            email: user.email,
            name: user.name,
          },
        }),
      });
    } else {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Invalid credentials',
        }),
      });
    }
  });

  await page.route('**/api/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        user: testUsers.admin,
      }),
    });
  });

  // Organizations
  await page.route('**/api/organizations', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          organizations: testOrganizations,
          total: testOrganizations.length,
        }),
      });
    } else if (route.request().method() === 'POST') {
      const body = await route.request().postDataJSON();
      const newOrg = {
        id: `org-e2e-${Date.now()}`,
        ...body,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        role: 'owner',
      };
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(newOrg),
      });
    }
  });

  // Workspaces
  await page.route('**/api/organizations/*/workspaces', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          workspaces: testWorkspaces,
          total: testWorkspaces.length,
        }),
      });
    } else if (route.request().method() === 'POST') {
      const body = await route.request().postDataJSON();
      const newWorkspace = {
        id: `ws-e2e-${Date.now()}`,
        ...body,
        vcluster_status: 'creating',
        vcluster_instance_name: `vcluster-${Date.now()}`,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      
      // Simulate async vCluster creation
      setTimeout(() => {
        newWorkspace.vcluster_status = 'active';
      }, 2000);
      
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(newWorkspace),
      });
    }
  });

  // Projects
  await page.route('**/api/organizations/*/workspaces/*/projects', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          projects: testProjects,
          total: testProjects.length,
        }),
      });
    } else if (route.request().method() === 'POST') {
      const body = await route.request().postDataJSON();
      const newProject = {
        id: `proj-e2e-${Date.now()}`,
        workspace_id: route.request().url().match(/workspaces\/([^\/]+)/)?.[1],
        ...body,
        namespace: body.namespace || `${body.name}-ns`,
        status: 'creating',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(newProject),
      });
    }
  });

  // Applications
  await page.route('**/api/organizations/*/workspaces/*/projects/*/applications', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          applications: testApplications,
          total: testApplications.length,
        }),
      });
    }
  });

  // Create application with deployment simulation
  await page.route('**/api/organizations/*/workspaces/*/applications', async (route) => {
    if (route.request().method() === 'POST') {
      const body = await route.request().postDataJSON();
      const appId = `app-e2e-${Date.now()}`;
      
      const newApp = {
        id: appId,
        workspace_id: route.request().url().match(/workspaces\/([^\/]+)/)?.[1],
        ...body,
        status: 'pending',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(newApp),
      });
    }
  });

  // Monitoring metrics
  await page.route('**/api/organizations/*/workspaces/*/metrics', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        metrics: testMetrics.workspace,
        timestamp: new Date().toISOString(),
      }),
    });
  });

  // Application metrics
  await page.route('**/api/organizations/*/workspaces/*/applications/*/metrics', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        metrics: [
          {
            ...testMetrics.application,
            timestamp: new Date().toISOString(),
          },
        ],
      }),
    });
  });
}

/**
 * Simulates progressive deployment stages
 */
export async function mockDeploymentProgress(
  page: Page,
  appId: string,
  duration: number = 5000
) {
  const stages = Object.values(deploymentStages);
  const stageDelay = duration / stages.length;
  
  let currentStage = 0;
  
  await page.route(`**/api/organizations/*/workspaces/*/applications/${appId}`, async (route) => {
    const stage = stages[Math.min(currentStage, stages.length - 1)];
    
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: appId,
        ...stage,
      }),
    });
  });
  
  // Progress through stages
  const interval = setInterval(() => {
    currentStage++;
    if (currentStage >= stages.length) {
      clearInterval(interval);
    }
  }, stageDelay);
  
  return () => clearInterval(interval);
}

/**
 * Mocks an API error response
 */
export async function mockAPIError(
  page: Page,
  urlPattern: string,
  error: { status: number; message: string }
) {
  await page.route(urlPattern, async (route) => {
    await route.fulfill({
      status: error.status,
      contentType: 'application/json',
      body: JSON.stringify({
        error: error.message,
      }),
    });
  });
}