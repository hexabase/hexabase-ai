import React from 'react';
import { render, screen, within, waitFor } from '@testing-library/react';
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
  setupCICD,
  monitorDeployment,
  verifyResourceStatus,
  expectNotification,
  expectNoErrors,
  generateProjectName,
  generateAppName,
  delay,
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

describe('CI/CD and Deployment Scenario', () => {
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
    
    // Setup mock responses for CI/CD operations
    mockApiClient.applications.create.mockResolvedValue({
      id: 'app-cicd-1',
      name: 'test-app',
      status: 'pending',
    });
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

  it('should setup CI/CD pipeline and deploy application through GitHub', async () => {
    renderApp();

    // Initial setup
    await loginUser();
    await createOrganization('DevOps Organization');
    await createWorkspace('CI/CD Workspace', 'dedicated');
    const projectName = generateProjectName();
    await createProject(projectName);

    // Step 1: Navigate to project settings
    console.log('Step 1: Navigating to CI/CD settings...');
    const projectCard = screen.getByText(projectName).closest('div');
    const enterButton = projectCard?.querySelector('button');
    if (enterButton) await userEvent.click(enterButton);

    // Step 2: Setup CI/CD with GitHub
    console.log('Step 2: Setting up GitHub integration...');
    await setupCICD(
      'https://github.com/hexabase/sample-app.git',
      'main',
      'npm run build'
    );
    
    // Verify webhook was created
    expect(mockApiClient.applications.create).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        source_type: 'git',
        source_git_url: 'https://github.com/hexabase/sample-app.git',
        source_git_ref: 'main',
      })
    );

    // Step 3: Configure build pipeline
    console.log('Step 3: Configuring build pipeline...');
    const pipelineTab = screen.getByRole('tab', { name: /pipeline/i });
    await userEvent.click(pipelineTab);

    // Add build step
    const addStepButton = screen.getByRole('button', { name: /add step/i });
    await userEvent.click(addStepButton);

    const stepModal = screen.getByRole('dialog');
    const stepNameInput = within(stepModal).getByLabelText(/step name/i);
    await userEvent.type(stepNameInput, 'Build Application');

    const stepTypeSelect = within(stepModal).getByLabelText(/step type/i);
    await userEvent.selectOptions(stepTypeSelect, 'build');

    const dockerfileInput = within(stepModal).getByLabelText(/dockerfile/i);
    await userEvent.type(dockerfileInput, './Dockerfile');

    const saveStepButton = within(stepModal).getByRole('button', { name: /save/i });
    await userEvent.click(saveStepButton);

    // Add test step
    await userEvent.click(addStepButton);
    const testStepModal = screen.getByRole('dialog');
    const testNameInput = within(testStepModal).getByLabelText(/step name/i);
    await userEvent.type(testNameInput, 'Run Tests');

    const testCommandInput = within(testStepModal).getByLabelText(/command/i);
    await userEvent.type(testCommandInput, 'npm test');

    const saveTestButton = within(testStepModal).getByRole('button', { name: /save/i });
    await userEvent.click(saveTestButton);

    // Step 4: Configure deployment settings
    console.log('Step 4: Configuring deployment...');
    const deployTab = screen.getByRole('tab', { name: /deployment/i });
    await userEvent.click(deployTab);

    const strategySelect = screen.getByLabelText(/deployment strategy/i);
    await userEvent.selectOptions(strategySelect, 'rolling');

    const replicasInput = screen.getByLabelText(/replicas/i);
    await userEvent.clear(replicasInput);
    await userEvent.type(replicasInput, '3');

    const healthCheckInput = screen.getByLabelText(/health check path/i);
    await userEvent.type(healthCheckInput, '/health');

    const saveDeployButton = screen.getByRole('button', { name: /save deployment config/i });
    await userEvent.click(saveDeployButton);

    // Step 5: Trigger deployment via git push simulation
    console.log('Step 5: Simulating git push to trigger deployment...');
    const triggerButton = screen.getByRole('button', { name: /trigger deployment/i });
    await userEvent.click(triggerButton);

    // Confirm deployment
    const confirmModal = screen.getByRole('dialog');
    const confirmButton = within(confirmModal).getByRole('button', { name: /deploy/i });
    await userEvent.click(confirmButton);

    await expectNotification(/deployment triggered/i);

    // Step 6: Monitor deployment progress
    console.log('Step 6: Monitoring deployment...');
    const deploymentName = 'test-app-deployment-1';
    
    // Check build stage
    await waitFor(() => {
      const buildStage = screen.getByTestId('stage-build');
      expect(buildStage).toHaveTextContent(/in progress/i);
    });

    // Simulate build completion
    await delay(2000);
    await waitFor(() => {
      const buildStage = screen.getByTestId('stage-build');
      expect(buildStage).toHaveTextContent(/completed/i);
    });

    // Check test stage
    await waitFor(() => {
      const testStage = screen.getByTestId('stage-test');
      expect(testStage).toHaveTextContent(/in progress/i);
    });

    // Simulate test completion
    await delay(1000);
    await waitFor(() => {
      const testStage = screen.getByTestId('stage-test');
      expect(testStage).toHaveTextContent(/completed/i);
    });

    // Check deployment stage
    await waitFor(() => {
      const deployStage = screen.getByTestId('stage-deploy');
      expect(deployStage).toHaveTextContent(/in progress/i);
    });

    // Monitor rolling update progress
    const progressBar = screen.getByRole('progressbar');
    expect(progressBar).toBeInTheDocument();

    // Wait for deployment completion
    await monitorDeployment(deploymentName);
    await expectNotification(/deployment completed successfully/i);

    // Step 7: Verify deployed application
    console.log('Step 7: Verifying deployed application...');
    const appsTab = screen.getByRole('tab', { name: /applications/i });
    await userEvent.click(appsTab);

    await verifyResourceStatus('test-app', 'running');

    // Check deployment metrics
    const metricsButton = screen.getByRole('button', { name: /view metrics/i });
    await userEvent.click(metricsButton);

    expect(screen.getByText(/deployment frequency/i)).toBeInTheDocument();
    expect(screen.getByText(/lead time/i)).toBeInTheDocument();
    expect(screen.getByText(/mttr/i)).toBeInTheDocument();

    console.log('✅ CI/CD deployment scenario completed successfully!');
  });

  it('should handle deployment rollback', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Rollback Test Org');
    await createWorkspace('Production', 'dedicated');
    await createProject('production-project');

    // Deploy initial version
    const appName = generateAppName();
    await setupCICD(
      'https://github.com/hexabase/sample-app.git',
      'main',
      'npm run build'
    );

    // Trigger deployment
    const triggerButton = screen.getByRole('button', { name: /trigger deployment/i });
    await userEvent.click(triggerButton);

    const confirmButton = screen.getByRole('button', { name: /deploy/i });
    await userEvent.click(confirmButton);

    await monitorDeployment(`${appName}-deployment-1`);

    // Simulate failed deployment of new version
    console.log('Simulating failed deployment...');
    mockApiClient.applications.updateStatus.mockRejectedValueOnce(
      new Error('Health check failed')
    );

    // Trigger another deployment
    await userEvent.click(triggerButton);
    await userEvent.click(confirmButton);

    // Wait for failure
    await waitFor(() => {
      const deployStage = screen.getByTestId('stage-deploy');
      expect(deployStage).toHaveTextContent(/failed/i);
    });

    await expectNotification(/deployment failed/i);

    // Initiate rollback
    console.log('Initiating rollback...');
    const rollbackButton = screen.getByRole('button', { name: /rollback/i });
    await userEvent.click(rollbackButton);

    // Select previous version
    const versionModal = screen.getByRole('dialog');
    const versionSelect = within(versionModal).getByLabelText(/select version/i);
    await userEvent.selectOptions(versionSelect, 'v1.0.0');

    const confirmRollbackButton = within(versionModal).getByRole('button', { 
      name: /confirm rollback/i 
    });
    await userEvent.click(confirmRollbackButton);

    // Monitor rollback
    await monitorDeployment(`${appName}-rollback-1`);
    await expectNotification(/rollback completed successfully/i);

    // Verify application is running previous version
    await verifyResourceStatus(appName, 'running');
    expect(screen.getByText(/version: v1.0.0/i)).toBeInTheDocument();
  });

  it('should support blue-green deployment strategy', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Blue-Green Org');
    await createWorkspace('Production', 'dedicated');
    await createProject('blue-green-project');

    // Setup CI/CD with blue-green strategy
    await setupCICD(
      'https://github.com/hexabase/sample-app.git',
      'main',
      'npm run build'
    );

    // Configure blue-green deployment
    const deployTab = screen.getByRole('tab', { name: /deployment/i });
    await userEvent.click(deployTab);

    const strategySelect = screen.getByLabelText(/deployment strategy/i);
    await userEvent.selectOptions(strategySelect, 'blue-green');

    const trafficSplitInput = screen.getByLabelText(/traffic split duration/i);
    await userEvent.type(trafficSplitInput, '300'); // 5 minutes

    const saveButton = screen.getByRole('button', { name: /save/i });
    await userEvent.click(saveButton);

    // Trigger deployment
    const triggerButton = screen.getByRole('button', { name: /trigger deployment/i });
    await userEvent.click(triggerButton);

    const confirmButton = screen.getByRole('button', { name: /deploy/i });
    await userEvent.click(confirmButton);

    // Monitor blue-green deployment stages
    console.log('Monitoring blue-green deployment...');
    
    // Stage 1: Deploy to green environment
    await waitFor(() => {
      const greenEnv = screen.getByTestId('env-green');
      expect(greenEnv).toHaveTextContent(/deploying/i);
    });

    await delay(2000);
    
    await waitFor(() => {
      const greenEnv = screen.getByTestId('env-green');
      expect(greenEnv).toHaveTextContent(/ready/i);
    });

    // Stage 2: Traffic split
    await waitFor(() => {
      const trafficStatus = screen.getByTestId('traffic-status');
      expect(trafficStatus).toHaveTextContent(/splitting traffic/i);
    });

    // Monitor traffic percentage
    const trafficSlider = screen.getByRole('slider', { name: /traffic to green/i });
    expect(trafficSlider).toHaveAttribute('aria-valuenow', '50');

    // Stage 3: Full cutover
    const cutoverButton = screen.getByRole('button', { name: /complete cutover/i });
    await userEvent.click(cutoverButton);

    await waitFor(() => {
      const blueEnv = screen.getByTestId('env-blue');
      expect(blueEnv).toHaveTextContent(/inactive/i);
      
      const greenEnv = screen.getByTestId('env-green');
      expect(greenEnv).toHaveTextContent(/active/i);
    });

    await expectNotification(/blue-green deployment completed/i);
    
    console.log('✅ Blue-green deployment completed successfully!');
  });

  it('should integrate with external CI systems', async () => {
    renderApp();

    await loginUser();
    await createOrganization('External CI Org');
    await createWorkspace('Integration Workspace', 'shared');
    await createProject('external-ci-project');

    // Navigate to CI/CD settings
    const settingsTab = screen.getByRole('tab', { name: /settings/i });
    await userEvent.click(settingsTab);

    const cicdTab = screen.getByRole('tab', { name: /ci\/cd/i });
    await userEvent.click(cicdTab);

    // Select external CI
    const ciTypeSelect = screen.getByLabelText(/ci provider/i);
    await userEvent.selectOptions(ciTypeSelect, 'github-actions');

    // Configure GitHub Actions
    const tokenInput = screen.getByLabelText(/github token/i);
    await userEvent.type(tokenInput, 'ghp_test_token_12345');

    const workflowInput = screen.getByLabelText(/workflow file/i);
    await userEvent.type(workflowInput, '.github/workflows/deploy.yml');

    // Configure webhook
    const webhookUrlDisplay = screen.getByTestId('webhook-url');
    expect(webhookUrlDisplay).toHaveTextContent(/https:\/\/api\.hexabase\.ai\/webhooks/i);

    const saveButton = screen.getByRole('button', { name: /save configuration/i });
    await userEvent.click(saveButton);

    await expectNotification(/github actions integrated successfully/i);

    // Verify webhook endpoint is active
    const testWebhookButton = screen.getByRole('button', { name: /test webhook/i });
    await userEvent.click(testWebhookButton);

    await expectNotification(/webhook test successful/i);

    console.log('✅ External CI integration completed!');
  });
});