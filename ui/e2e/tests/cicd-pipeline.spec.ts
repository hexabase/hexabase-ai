import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { ApplicationPage } from '../pages/ApplicationPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers, testWorkspaces } from '../fixtures/mock-data';
import { generateAppName, expectNotification } from '../utils/test-helpers';

test.describe('CI/CD Pipeline Integration', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let workspacePage: WorkspacePage;
  let projectPage: ProjectPage;
  let applicationPage: ApplicationPage;

  test.beforeEach(async ({ page }) => {
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    workspacePage = new WorkspacePage(page);
    projectPage = new ProjectPage(page);
    applicationPage = new ApplicationPage(page);
    
    // Setup test environment
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    await workspacePage.createProject('E2E CI/CD Project');
    await workspacePage.openProject('E2E CI/CD Project');
  });

  test('setup GitHub repository integration', async ({ page }) => {
    // Navigate to project settings
    await projectPage.settingsTab.click();
    
    // Go to CI/CD settings
    const cicdTab = page.getByRole('tab', { name: /ci\/cd/i });
    await cicdTab.click();
    
    // Connect GitHub repository
    const connectRepoButton = page.getByTestId('connect-repository-button');
    await connectRepoButton.click();
    
    const repoDialog = page.getByRole('dialog');
    
    // Select GitHub as provider
    await repoDialog.getByTestId('provider-github').click();
    
    // Enter repository details
    await repoDialog.getByTestId('repo-url-input').fill('https://github.com/hexabase/sample-app');
    await repoDialog.getByTestId('default-branch-input').fill('main');
    
    // Configure access token
    await repoDialog.getByTestId('access-token-input').fill('ghp_test_token_12345');
    
    // Enable auto-deploy
    const autoDeployToggle = repoDialog.getByTestId('auto-deploy-toggle');
    await autoDeployToggle.check();
    
    // Save configuration
    await repoDialog.getByTestId('connect-button').click();
    
    // Verify connection successful
    await expectNotification(page, 'Repository connected successfully');
    
    // Verify webhook URL displayed
    const webhookUrl = page.getByTestId('webhook-url');
    await expect(webhookUrl).toBeVisible();
    await expect(webhookUrl).toContainText('https://api.hexabase.ai/webhooks/');
    
    // Verify repository info displayed
    await expect(page.getByTestId('connected-repo')).toContainText('hexabase/sample-app');
    await expect(page.getByTestId('repo-branch')).toContainText('main');
  });

  test('configure build pipeline', async ({ page }) => {
    const appName = generateAppName();
    
    // First connect a repository
    await projectPage.settingsTab.click();
    const cicdTab = page.getByRole('tab', { name: /ci\/cd/i });
    await cicdTab.click();
    
    // Mock repository already connected
    await page.route('**/api/organizations/*/workspaces/*/projects/*/cicd', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          repository: {
            url: 'https://github.com/hexabase/sample-app',
            branch: 'main',
            connected: true,
          },
        }),
      });
    });
    
    await page.reload();
    
    // Configure build pipeline
    const pipelineTab = page.getByTestId('pipeline-configuration-tab');
    await pipelineTab.click();
    
    // Add build stage
    const addStageButton = page.getByTestId('add-pipeline-stage-button');
    await addStageButton.click();
    
    const stageDialog = page.getByRole('dialog');
    await stageDialog.getByTestId('stage-name-input').fill('Build Application');
    await stageDialog.getByTestId('stage-type-select').selectOption('build');
    
    // Configure build settings
    await stageDialog.getByTestId('dockerfile-path-input').fill('./Dockerfile');
    await stageDialog.getByTestId('build-context-input').fill('.');
    await stageDialog.getByTestId('target-stage-input').fill('production');
    
    // Add build arguments
    const addArgButton = stageDialog.getByTestId('add-build-arg-button');
    await addArgButton.click();
    await stageDialog.getByTestId('arg-key-0').fill('NODE_ENV');
    await stageDialog.getByTestId('arg-value-0').fill('production');
    
    await stageDialog.getByTestId('save-stage-button').click();
    
    // Add test stage
    await addStageButton.click();
    const testDialog = page.getByRole('dialog');
    await testDialog.getByTestId('stage-name-input').fill('Run Tests');
    await testDialog.getByTestId('stage-type-select').selectOption('test');
    await testDialog.getByTestId('test-command-input').fill('npm test');
    await testDialog.getByTestId('coverage-threshold-input').fill('80');
    await testDialog.getByTestId('save-stage-button').click();
    
    // Add deploy stage
    await addStageButton.click();
    const deployDialog = page.getByRole('dialog');
    await deployDialog.getByTestId('stage-name-input').fill('Deploy to Cluster');
    await deployDialog.getByTestId('stage-type-select').selectOption('deploy');
    await deployDialog.getByTestId('deployment-strategy-select').selectOption('rolling');
    await deployDialog.getByTestId('save-stage-button').click();
    
    // Save pipeline
    const savePipelineButton = page.getByTestId('save-pipeline-button');
    await savePipelineButton.click();
    
    await expectNotification(page, 'Pipeline configuration saved');
    
    // Verify pipeline stages displayed
    const pipelineStages = page.locator('[data-testid^="pipeline-stage-"]');
    await expect(pipelineStages).toHaveCount(3);
  });

  test('trigger manual deployment', async ({ page }) => {
    const appName = generateAppName();
    
    // Create application linked to repository
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'registry.hexabase.ai/sample-app:latest',
      replicas: 2,
      port: 3000,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    // Go to CI/CD tab
    const cicdTab = page.getByRole('tab', { name: /ci\/cd/i });
    await cicdTab.click();
    
    // Trigger deployment
    const triggerDeployButton = page.getByTestId('trigger-deployment-button');
    await triggerDeployButton.click();
    
    const deployDialog = page.getByRole('dialog');
    
    // Select branch/tag
    await deployDialog.getByTestId('ref-type-select').selectOption('branch');
    await deployDialog.getByTestId('branch-select').selectOption('develop');
    
    // Add deployment message
    await deployDialog.getByTestId('deployment-message-input').fill('Manual deployment for testing');
    
    // Start deployment
    await deployDialog.getByTestId('start-deployment-button').click();
    
    // Monitor pipeline execution
    await expect(page.getByText('Pipeline started')).toBeVisible();
    
    // Check pipeline stages
    const buildStage = page.getByTestId('stage-build-status');
    await expect(buildStage).toContainText('running');
    
    // Mock build completion
    await page.waitForTimeout(3000);
    await expect(buildStage).toContainText('completed');
    
    const testStage = page.getByTestId('stage-test-status');
    await expect(testStage).toContainText('running');
    
    await page.waitForTimeout(2000);
    await expect(testStage).toContainText('completed');
    
    const deployStage = page.getByTestId('stage-deploy-status');
    await expect(deployStage).toContainText('running');
    
    await page.waitForTimeout(3000);
    await expect(deployStage).toContainText('completed');
    
    await expectNotification(page, 'Deployment completed successfully');
  });

  test('configure automatic deployments on push', async ({ page }) => {
    // Setup repository first
    await projectPage.settingsTab.click();
    const cicdTab = page.getByRole('tab', { name: /ci\/cd/i });
    await cicdTab.click();
    
    // Configure auto-deployment rules
    const autoDeployTab = page.getByTestId('auto-deploy-tab');
    await autoDeployTab.click();
    
    // Enable auto-deploy
    const enableAutoDeployToggle = page.getByTestId('enable-auto-deploy-toggle');
    await enableAutoDeployToggle.check();
    
    // Add deployment rule for main branch
    const addRuleButton = page.getByTestId('add-deploy-rule-button');
    await addRuleButton.click();
    
    const ruleDialog = page.getByRole('dialog');
    await ruleDialog.getByTestId('rule-name-input').fill('Deploy to Production');
    await ruleDialog.getByTestId('branch-pattern-input').fill('main');
    await ruleDialog.getByTestId('target-env-select').selectOption('production');
    
    // Configure deployment conditions
    await ruleDialog.getByTestId('require-tests-checkbox').check();
    await ruleDialog.getByTestId('require-approval-checkbox').uncheck();
    
    await ruleDialog.getByTestId('save-rule-button').click();
    
    // Add rule for develop branch
    await addRuleButton.click();
    const devRuleDialog = page.getByRole('dialog');
    await devRuleDialog.getByTestId('rule-name-input').fill('Deploy to Staging');
    await devRuleDialog.getByTestId('branch-pattern-input').fill('develop');
    await devRuleDialog.getByTestId('target-env-select').selectOption('staging');
    await devRuleDialog.getByTestId('save-rule-button').click();
    
    // Save auto-deploy configuration
    const saveAutoDeployButton = page.getByTestId('save-auto-deploy-button');
    await saveAutoDeployButton.click();
    
    await expectNotification(page, 'Auto-deployment configured');
    
    // Verify rules displayed
    const deployRules = page.locator('[data-testid^="deploy-rule-"]');
    await expect(deployRules).toHaveCount(2);
  });

  test('view deployment history and rollback', async ({ page }) => {
    const appName = generateAppName();
    
    // Create application with deployment history
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'registry.hexabase.ai/sample-app:v1.0.0',
      replicas: 2,
      port: 3000,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Mock deployment history
    await page.route('**/api/organizations/*/workspaces/*/applications/*/deployments', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          deployments: [
            {
              id: 'deploy-3',
              version: 'v1.2.0',
              status: 'active',
              deployed_at: new Date().toISOString(),
              deployed_by: 'CI/CD Pipeline',
              commit_sha: 'abc123',
              commit_message: 'feat: add new feature',
            },
            {
              id: 'deploy-2',
              version: 'v1.1.0',
              status: 'previous',
              deployed_at: new Date(Date.now() - 86400000).toISOString(),
              deployed_by: 'john@example.com',
              commit_sha: 'def456',
              commit_message: 'fix: bug fixes',
            },
            {
              id: 'deploy-1',
              version: 'v1.0.0',
              status: 'previous',
              deployed_at: new Date(Date.now() - 172800000).toISOString(),
              deployed_by: 'CI/CD Pipeline',
              commit_sha: 'ghi789',
              commit_message: 'initial release',
            },
          ],
        }),
      });
    });
    
    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    // Go to deployments tab
    const deploymentsTab = page.getByRole('tab', { name: /deployments/i });
    await deploymentsTab.click();
    
    // Verify deployment history
    const deploymentRows = page.locator('[data-testid^="deployment-row-"]');
    await expect(deploymentRows).toHaveCount(3);
    
    // Check current deployment
    const currentDeployment = deploymentRows.first();
    await expect(currentDeployment).toContainText('v1.2.0');
    await expect(currentDeployment).toContainText('Active');
    await expect(currentDeployment).toContainText('feat: add new feature');
    
    // Rollback to previous version
    const v1_1_deployment = deploymentRows.nth(1);
    const rollbackButton = v1_1_deployment.getByTestId('rollback-button');
    await rollbackButton.click();
    
    const rollbackDialog = page.getByRole('dialog');
    await expect(rollbackDialog).toContainText('Rollback to v1.1.0');
    await expect(rollbackDialog).toContainText('fix: bug fixes');
    
    // Confirm rollback
    await rollbackDialog.getByTestId('confirm-rollback-button').click();
    
    // Monitor rollback progress
    await expect(page.getByText('Rollback in progress')).toBeVisible();
    
    await page.waitForTimeout(3000);
    await expectNotification(page, 'Rollback completed successfully');
    
    // Verify new active version
    await expect(deploymentRows.nth(1)).toContainText('Active');
  });

  test('configure build notifications', async ({ page }) => {
    // Navigate to CI/CD settings
    await projectPage.settingsTab.click();
    const cicdTab = page.getByRole('tab', { name: /ci\/cd/i });
    await cicdTab.click();
    
    // Go to notifications tab
    const notificationsTab = page.getByTestId('notifications-tab');
    await notificationsTab.click();
    
    // Configure email notifications
    const emailSection = page.getByTestId('email-notifications-section');
    const enableEmailToggle = emailSection.getByTestId('enable-email-toggle');
    await enableEmailToggle.check();
    
    await emailSection.getByTestId('email-recipients-input').fill('team@example.com, dev@example.com');
    
    // Configure notification triggers
    await emailSection.getByTestId('notify-success-checkbox').check();
    await emailSection.getByTestId('notify-failure-checkbox').check();
    await emailSection.getByTestId('notify-started-checkbox').uncheck();
    
    // Configure Slack notifications
    const slackSection = page.getByTestId('slack-notifications-section');
    const enableSlackToggle = slackSection.getByTestId('enable-slack-toggle');
    await enableSlackToggle.check();
    
    await slackSection.getByTestId('slack-webhook-input').fill('https://hooks.slack.com/services/TEST/WEBHOOK');
    await slackSection.getByTestId('slack-channel-input').fill('#deployments');
    
    // Configure webhook notifications
    const webhookSection = page.getByTestId('webhook-notifications-section');
    const addWebhookButton = webhookSection.getByTestId('add-webhook-button');
    await addWebhookButton.click();
    
    const webhookDialog = page.getByRole('dialog');
    await webhookDialog.getByTestId('webhook-name-input').fill('Custom Integration');
    await webhookDialog.getByTestId('webhook-url-input').fill('https://api.example.com/deployments');
    await webhookDialog.getByTestId('webhook-secret-input').fill('webhook_secret_123');
    await webhookDialog.getByTestId('save-webhook-button').click();
    
    // Save notification settings
    const saveNotificationsButton = page.getByTestId('save-notifications-button');
    await saveNotificationsButton.click();
    
    await expectNotification(page, 'Notification settings saved');
  });

  test('view build logs and artifacts', async ({ page }) => {
    const appName = generateAppName();
    
    // Create application
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'registry.hexabase.ai/sample-app:latest',
      replicas: 1,
      port: 3000,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Navigate to application CI/CD
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    const cicdTab = page.getByRole('tab', { name: /ci\/cd/i });
    await cicdTab.click();
    
    // Go to builds tab
    const buildsTab = page.getByTestId('builds-tab');
    await buildsTab.click();
    
    // Mock build history
    await page.route('**/api/organizations/*/workspaces/*/applications/*/builds', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          builds: [
            {
              id: 'build-123',
              number: 123,
              status: 'success',
              branch: 'main',
              commit: 'abc123def',
              started_at: new Date(Date.now() - 3600000).toISOString(),
              finished_at: new Date(Date.now() - 3000000).toISOString(),
              duration: 600,
            },
          ],
        }),
      });
    });
    
    // View build details
    const buildRow = page.locator('[data-testid^="build-row-"]').first();
    await buildRow.click();
    
    // Check build logs
    const logsTab = page.getByRole('tab', { name: /logs/i });
    await logsTab.click();
    
    await expect(page.getByTestId('build-logs')).toBeVisible();
    await expect(page.getByTestId('build-logs')).toContainText('Building application');
    
    // Check artifacts
    const artifactsTab = page.getByRole('tab', { name: /artifacts/i });
    await artifactsTab.click();
    
    const artifactsList = page.getByTestId('artifacts-list');
    await expect(artifactsList).toBeVisible();
    
    // Download artifact
    const downloadButton = artifactsList.getByTestId('download-artifact-button').first();
    const downloadPromise = page.waitForEvent('download');
    await downloadButton.click();
    const download = await downloadPromise;
    
    expect(download.suggestedFilename()).toContain('build-123');
  });

  test('configure environment-specific deployments', async ({ page }) => {
    // Navigate to CI/CD settings
    await projectPage.settingsTab.click();
    const cicdTab = page.getByRole('tab', { name: /ci\/cd/i });
    await cicdTab.click();
    
    // Go to environments tab
    const environmentsTab = page.getByTestId('environments-tab');
    await environmentsTab.click();
    
    // Add development environment
    const addEnvButton = page.getByTestId('add-environment-button');
    await addEnvButton.click();
    
    const envDialog = page.getByRole('dialog');
    await envDialog.getByTestId('env-name-input').fill('development');
    await envDialog.getByTestId('env-url-input').fill('https://dev.example.com');
    
    // Configure environment variables
    await envDialog.getByTestId('add-env-var-button').click();
    await envDialog.getByTestId('env-var-key-0').fill('API_URL');
    await envDialog.getByTestId('env-var-value-0').fill('https://api-dev.example.com');
    
    await envDialog.getByTestId('add-env-var-button').click();
    await envDialog.getByTestId('env-var-key-1').fill('DEBUG');
    await envDialog.getByTestId('env-var-value-1').fill('true');
    
    await envDialog.getByTestId('save-environment-button').click();
    
    // Add production environment
    await addEnvButton.click();
    const prodDialog = page.getByRole('dialog');
    await prodDialog.getByTestId('env-name-input').fill('production');
    await prodDialog.getByTestId('env-url-input').fill('https://app.example.com');
    
    // Enable manual approval for production
    await prodDialog.getByTestId('require-approval-toggle').check();
    await prodDialog.getByTestId('approvers-input').fill('admin@example.com, lead@example.com');
    
    // Configure production variables
    await prodDialog.getByTestId('add-env-var-button').click();
    await prodDialog.getByTestId('env-var-key-0').fill('API_URL');
    await prodDialog.getByTestId('env-var-value-0').fill('https://api.example.com');
    
    await prodDialog.getByTestId('add-env-var-button').click();
    await prodDialog.getByTestId('env-var-key-1').fill('DEBUG');
    await prodDialog.getByTestId('env-var-value-1').fill('false');
    
    await prodDialog.getByTestId('save-environment-button').click();
    
    // Verify environments created
    const envList = page.locator('[data-testid^="environment-card-"]');
    await expect(envList).toHaveCount(2);
    
    // Verify production requires approval
    const prodCard = envList.filter({ hasText: 'production' });
    await expect(prodCard.getByTestId('approval-required-badge')).toBeVisible();
  });

  test('handle build failures and retry', async ({ page }) => {
    const appName = generateAppName();
    
    // Create application
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'registry.hexabase.ai/sample-app:latest',
      replicas: 1,
      port: 3000,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Navigate to application CI/CD
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    const cicdTab = page.getByRole('tab', { name: /ci\/cd/i });
    await cicdTab.click();
    
    // Mock build failure
    await page.route('**/api/organizations/*/workspaces/*/applications/*/builds/trigger', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          build_id: 'build-fail-123',
          status: 'running',
        }),
      });
    });
    
    // Trigger build
    const triggerBuildButton = page.getByTestId('trigger-build-button');
    await triggerBuildButton.click();
    
    // Mock build status updates
    await page.route('**/api/organizations/*/workspaces/*/applications/*/builds/build-fail-123', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'build-fail-123',
          status: 'failed',
          error: 'Tests failed: 5 tests failed out of 100',
          failed_stage: 'test',
        }),
      });
    });
    
    // Wait for build to fail
    await page.waitForTimeout(3000);
    await expect(page.getByText('Build failed')).toBeVisible();
    await expect(page.getByText('Tests failed: 5 tests failed')).toBeVisible();
    
    // View failed build details
    const viewDetailsButton = page.getByTestId('view-build-details-button');
    await viewDetailsButton.click();
    
    // Check failed stage
    await expect(page.getByTestId('stage-test-status')).toContainText('failed');
    
    // Retry build
    const retryButton = page.getByTestId('retry-build-button');
    await retryButton.click();
    
    const retryDialog = page.getByRole('dialog');
    await expect(retryDialog).toContainText('Retry failed build');
    
    // Option to skip failed tests (not recommended)
    const skipTestsCheckbox = retryDialog.getByTestId('skip-tests-checkbox');
    await expect(skipTestsCheckbox).toBeVisible();
    
    // Retry without skipping
    await retryDialog.getByTestId('confirm-retry-button').click();
    
    await expectNotification(page, 'Build retry started');
  });
});