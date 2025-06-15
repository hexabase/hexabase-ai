import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
// Next.js uses its own routing, no need for BrowserRouter
import HomePage from '@/app/page';
import { AuthProvider } from '@/lib/auth-context';
import { mockApiClient } from '@/test-utils/mock-api-client';
import {
  loginUser,
  createOrganization,
  createWorkspace,
  createProject,
  deployApplication,
  triggerCronJob,
  verifyResourceStatus,
  expectNotification,
  expectNoErrors,
  generateProjectName,
  generateAppName,
} from './test-helpers';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
  }),
  usePathname: () => '/',
}));

describe('Basic User Journey - From Login to Resource Deployment', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    
    // Reset all mocks
    jest.clearAllMocks();
  });

  const renderApp = () => {
    return render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <HomePage />
        </AuthProvider>
      </QueryClientProvider>
    );
  };

  it('should complete full user journey from login to deploying multiple resources', async () => {
    renderApp();

    // Step 1: Login
    console.log('Step 1: Logging in...');
    await loginUser('test@hexabase.ai', 'Test123!');
    expect(mockApiClient.auth.login).toHaveBeenCalledWith({
      email: 'test@hexabase.ai',
      password: 'Test123!',
    });
    expectNoErrors();

    // Step 2: Create Organization
    console.log('Step 2: Creating organization...');
    const orgName = 'Test Organization';
    await createOrganization(orgName);
    expect(mockApiClient.organizations.create).toHaveBeenCalledWith({
      name: orgName,
    });
    await expectNotification(/organization created successfully/i);

    // Step 3: Create Dedicated Workspace
    console.log('Step 3: Creating dedicated workspace...');
    const workspaceName = 'Production Workspace';
    await createWorkspace(workspaceName, 'dedicated');
    expect(mockApiClient.workspaces.create).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        name: workspaceName,
        plan_id: 'dedicated',
      })
    );
    await expectNotification(/workspace created successfully/i);

    // Step 4: Navigate to workspace
    const workspaceCard = screen.getByText(workspaceName).closest('div');
    const enterButton = workspaceCard?.querySelector('button');
    if (enterButton) await userEvent.click(enterButton);

    // Step 5: Create Project with Resource Quotas
    console.log('Step 5: Creating project with quotas...');
    const projectName = generateProjectName();
    await createProject(projectName, {
      cpu: '4',
      memory: '8Gi',
      storage: '50Gi',
    });
    expect(mockApiClient.projects.create).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        name: projectName,
        resource_quota: {
          cpu: '4',
          memory: '8Gi',
          storage: '50Gi',
        },
      })
    );

    // Step 6: Navigate to project
    const projectCard = screen.getByText(projectName).closest('div');
    const projectEnterButton = projectCard?.querySelector('button');
    if (projectEnterButton) await userEvent.click(projectEnterButton);

    // Step 7: Deploy Stateless Application (nginx)
    console.log('Step 7: Deploying stateless application...');
    const nginxAppName = generateAppName();
    await deployApplication(nginxAppName, 'stateless', {
      image: 'nginx:latest',
      replicas: 3,
      port: 80,
    });
    expect(mockApiClient.applications.create).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        name: nginxAppName,
        type: 'stateless',
        source_type: 'image',
        source_image: 'nginx:latest',
        config: expect.objectContaining({
          replicas: 3,
          port: 80,
        }),
      })
    );
    await verifyResourceStatus(nginxAppName, 'running');

    // Step 8: Deploy Stateful Application (PostgreSQL)
    console.log('Step 8: Deploying stateful application...');
    const postgresAppName = generateAppName();
    await deployApplication(postgresAppName, 'stateful', {
      image: 'postgres:14',
      port: 5432,
      storage: '10Gi',
    });
    expect(mockApiClient.applications.create).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        name: postgresAppName,
        type: 'stateful',
        source_type: 'image',
        source_image: 'postgres:14',
        config: expect.objectContaining({
          port: 5432,
          storage_size: '10Gi',
        }),
      })
    );
    await verifyResourceStatus(postgresAppName, 'running');

    // Step 9: Create and Trigger CronJob
    console.log('Step 9: Creating CronJob...');
    const cronJobName = generateAppName();
    await deployApplication(cronJobName, 'cronjob', {
      image: 'busybox:latest',
      schedule: '0 2 * * *', // Daily at 2 AM
    });
    expect(mockApiClient.applications.create).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        name: cronJobName,
        type: 'cronjob',
        source_type: 'image',
        source_image: 'busybox:latest',
        cron_schedule: '0 2 * * *',
      })
    );
    await verifyResourceStatus(cronJobName, 'active');

    // Step 10: Manually trigger the CronJob
    console.log('Step 10: Triggering CronJob...');
    await triggerCronJob(cronJobName);
    expect(mockApiClient.applications.triggerCronJob).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.any(String)
    );
    await expectNotification(/cronjob triggered successfully/i);

    // Step 11: Verify all resources are healthy
    console.log('Step 11: Verifying all resources...');
    await verifyResourceStatus(nginxAppName, 'running');
    await verifyResourceStatus(postgresAppName, 'running');
    await verifyResourceStatus(cronJobName, 'active');

    // Step 12: Check resource usage
    const resourceUsageCard = screen.getByTestId('resource-usage');
    expect(resourceUsageCard).toBeInTheDocument();
    expect(resourceUsageCard).toHaveTextContent(/cpu.*used/i);
    expect(resourceUsageCard).toHaveTextContent(/memory.*used/i);
    expect(resourceUsageCard).toHaveTextContent(/storage.*used/i);

    console.log('✅ Basic user journey completed successfully!');
  });

  it('should handle errors gracefully during resource creation', async () => {
    renderApp();
    
    // Setup error response
    mockApiClient.applications.create.mockRejectedValueOnce(
      new Error('Insufficient resources')
    );

    await loginUser();
    await createOrganization('Test Org');
    await createWorkspace('Test Workspace');
    await createProject('Test Project');

    // Try to deploy application that will fail
    await expect(async () => {
      await deployApplication('failing-app', 'stateless', {
        image: 'nginx:latest',
        replicas: 100, // Too many replicas
      });
    }).rejects.toThrow();

    // Verify error notification
    await expectNotification(/insufficient resources/i);

    // Verify we can still continue with other operations
    await deployApplication('successful-app', 'stateless', {
      image: 'nginx:latest',
      replicas: 1,
    });
    await verifyResourceStatus('successful-app', 'running');
  });

  it('should support workspace switching', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Multi-Workspace Org');

    // Create multiple workspaces
    await createWorkspace('Development', 'shared');
    await createWorkspace('Staging', 'shared');
    await createWorkspace('Production', 'dedicated');

    // Verify all workspaces are listed
    expect(screen.getByText('Development')).toBeInTheDocument();
    expect(screen.getByText('Staging')).toBeInTheDocument();
    expect(screen.getByText('Production')).toBeInTheDocument();

    // Switch between workspaces
    const stagingCard = screen.getByText('Staging').closest('div');
    const enterStagingButton = stagingCard?.querySelector('button');
    if (enterStagingButton) await userEvent.click(enterStagingButton);

    // Verify we're in staging workspace
    expect(screen.getByTestId('current-workspace')).toHaveTextContent('Staging');

    // Create project in staging
    await createProject('Staging Project');
    
    // Go back to workspaces list
    const backButton = screen.getByRole('button', { name: /back/i });
    await userEvent.click(backButton);

    // Enter production workspace
    const prodCard = screen.getByText('Production').closest('div');
    const enterProdButton = prodCard?.querySelector('button');
    if (enterProdButton) await userEvent.click(enterProdButton);

    // Verify we're in production workspace
    expect(screen.getByTestId('current-workspace')).toHaveTextContent('Production');

    // Create project in production
    await createProject('Production Project');

    // Verify projects are isolated per workspace
    expect(screen.queryByText('Staging Project')).not.toBeInTheDocument();
    expect(screen.getByText('Production Project')).toBeInTheDocument();
  });
});