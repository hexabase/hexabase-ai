import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { ApplicationPage } from '../pages/ApplicationPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers, testWorkspaces } from '../fixtures/mock-data';
import { generateAppName, expectNotification } from '../utils/test-helpers';
import { SMOKE_TAG, CRITICAL_TAG } from '../utils/test-tags';

test.describe('AI Chat Interactions', () => {
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
    
    // Login and setup
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
  });

  test(`open AI assistant chat ${SMOKE_TAG}`, async ({ page }) => {
    // Click AI assistant button (floating or in header)
    const aiAssistantButton = page.getByTestId('ai-assistant-button');
    await expect(aiAssistantButton).toBeVisible();
    await aiAssistantButton.click();
    
    // Verify chat window opens
    const chatWindow = page.getByTestId('ai-chat-window');
    await expect(chatWindow).toBeVisible();
    
    // Verify initial UI elements
    await expect(page.getByTestId('ai-chat-header')).toContainText('AI Assistant');
    await expect(page.getByTestId('chat-input')).toBeVisible();
    await expect(page.getByTestId('send-message-button')).toBeVisible();
    
    // Verify welcome message
    const welcomeMessage = page.getByTestId('ai-welcome-message');
    await expect(welcomeMessage).toBeVisible();
    await expect(welcomeMessage).toContainText(/Hello|Hi|Welcome/i);
    await expect(welcomeMessage).toContainText(/How can I help/i);
  });

  test('basic AI conversation flow', async ({ page }) => {
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    const chatInput = page.getByTestId('chat-input');
    const sendButton = page.getByTestId('send-message-button');
    
    // Send first message
    await chatInput.fill('How do I deploy a new application?');
    await sendButton.click();
    
    // Verify user message appears
    const userMessage = page.locator('[data-testid="chat-message-user"]').last();
    await expect(userMessage).toBeVisible();
    await expect(userMessage).toContainText('How do I deploy a new application?');
    
    // Wait for AI response
    const aiTypingIndicator = page.getByTestId('ai-typing-indicator');
    await expect(aiTypingIndicator).toBeVisible();
    
    // Mock AI response
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'To deploy a new application:\n1. Navigate to your project\n2. Click "Deploy Application"\n3. Fill in the application details\n4. Click "Deploy"',
          metadata: {
            model: 'gpt-4',
            tokens_used: 45,
            sources: ['docs/deployment.md']
          }
        })
      });
    });
    
    // Verify AI response
    await expect(aiTypingIndicator).not.toBeVisible({ timeout: 10000 });
    const aiMessage = page.locator('[data-testid="chat-message-ai"]').last();
    await expect(aiMessage).toBeVisible();
    await expect(aiMessage).toContainText('To deploy a new application');
    await expect(aiMessage).toContainText('Navigate to your project');
    
    // Verify response metadata
    const responseMetadata = aiMessage.getByTestId('response-metadata');
    await expect(responseMetadata).toContainText('gpt-4');
    await expect(responseMetadata).toContainText('45 tokens');
  });

  test('AI-assisted application deployment', async ({ page }) => {
    // Navigate to project
    await workspacePage.createProject('AI Assistant Test');
    await workspacePage.openProject('AI Assistant Test');
    
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    // Ask AI to help deploy an app
    const chatInput = page.getByTestId('chat-input');
    await chatInput.fill('Help me deploy a PostgreSQL database');
    await page.getByTestId('send-message-button').click();
    
    // Mock AI response with action buttons
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'I can help you deploy PostgreSQL. Here\'s a recommended configuration:',
          actions: [
            {
              type: 'deploy_application',
              label: 'Deploy PostgreSQL',
              config: {
                name: 'postgres-db',
                type: 'stateful',
                image: 'postgres:14',
                port: 5432,
                storage: '20Gi',
                env: {
                  POSTGRES_DB: 'myapp',
                  POSTGRES_USER: 'admin'
                }
              }
            }
          ]
        })
      });
    });
    
    // Wait for AI response with action
    await page.waitForTimeout(1000);
    const actionButton = page.getByTestId('ai-action-deploy-application');
    await expect(actionButton).toBeVisible();
    await expect(actionButton).toContainText('Deploy PostgreSQL');
    
    // Click action button
    await actionButton.click();
    
    // Verify deployment dialog opens with pre-filled values
    const deployDialog = page.getByRole('dialog');
    await expect(deployDialog).toBeVisible();
    await expect(deployDialog.getByTestId('app-name-input')).toHaveValue('postgres-db');
    await expect(deployDialog.getByTestId('app-type-select')).toHaveValue('stateful');
    await expect(deployDialog.getByTestId('app-image-input')).toHaveValue('postgres:14');
    await expect(deployDialog.getByTestId('storage-size-input')).toHaveValue('20Gi');
    
    // Complete deployment
    await deployDialog.getByTestId('env-value-POSTGRES_PASSWORD').fill('secure-password-123');
    await deployDialog.getByTestId('deploy-button').click();
    
    // Verify deployment started
    await expectNotification(page, 'Deployment started');
    
    // Verify AI acknowledges deployment
    await expect(page.getByTestId('ai-chat-window')).toContainText('PostgreSQL deployment initiated');
  });

  test('AI code generation and explanation', async ({ page }) => {
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    // Ask for code generation
    const chatInput = page.getByTestId('chat-input');
    await chatInput.fill('Generate a Kubernetes manifest for a Node.js application');
    await page.getByTestId('send-message-button').click();
    
    // Mock AI response with code
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Here\'s a Kubernetes manifest for your Node.js application:',
          code: {
            language: 'yaml',
            content: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nodejs-app
  labels:
    app: nodejs-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nodejs-app
  template:
    metadata:
      labels:
        app: nodejs-app
    spec:
      containers:
      - name: nodejs
        image: node:16-alpine
        ports:
        - containerPort: 3000
        env:
        - name: NODE_ENV
          value: "production"
---
apiVersion: v1
kind: Service
metadata:
  name: nodejs-service
spec:
  selector:
    app: nodejs-app
  ports:
  - port: 80
    targetPort: 3000
  type: LoadBalancer`
          },
          actions: [
            {
              type: 'copy_code',
              label: 'Copy to clipboard'
            },
            {
              type: 'apply_manifest',
              label: 'Apply to cluster'
            }
          ]
        })
      });
    });
    
    // Verify code block display
    await page.waitForTimeout(1000);
    const codeBlock = page.getByTestId('ai-code-block');
    await expect(codeBlock).toBeVisible();
    await expect(codeBlock).toContainText('apiVersion: apps/v1');
    await expect(codeBlock).toContainText('kind: Deployment');
    
    // Verify syntax highlighting
    await expect(codeBlock.locator('.language-yaml')).toBeVisible();
    
    // Test copy code button
    const copyButton = page.getByTestId('copy-code-button');
    await copyButton.click();
    await expectNotification(page, 'Code copied to clipboard');
    
    // Test apply manifest action
    const applyButton = page.getByTestId('ai-action-apply-manifest');
    await applyButton.click();
    
    // Verify confirmation dialog
    const confirmDialog = page.getByRole('dialog');
    await expect(confirmDialog).toContainText('Apply Kubernetes Manifest');
    await confirmDialog.getByTestId('cancel-button').click();
  });

  test('AI troubleshooting assistance', async ({ page }) => {
    // Create and deploy a failing application
    await workspacePage.createProject('Troubleshooting Test');
    await workspacePage.openProject('Troubleshooting Test');
    
    const appName = generateAppName();
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 1,
      port: 80,
    });
    
    // Mock application in error state
    await page.route(`**/api/organizations/*/workspaces/*/applications/${appName}`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: appName,
          status: 'error',
          error: 'CrashLoopBackOff',
          message: 'Container repeatedly crashing after startup'
        })
      });
    });
    
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    // Ask for help with error
    const chatInput = page.getByTestId('chat-input');
    await chatInput.fill(`My application ${appName} is showing CrashLoopBackOff error. Can you help?`);
    await page.getByTestId('send-message-button').click();
    
    // Mock AI diagnostic response
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'I can see your application is experiencing a CrashLoopBackOff error. Let me help you diagnose this:',
          diagnostic: {
            issue: 'CrashLoopBackOff',
            possibleCauses: [
              'Application crashes immediately after starting',
              'Missing environment variables',
              'Insufficient resources',
              'Configuration errors'
            ],
            suggestedActions: [
              'Check application logs',
              'Verify environment variables',
              'Increase resource limits',
              'Review container configuration'
            ]
          },
          actions: [
            {
              type: 'view_logs',
              label: 'View Application Logs',
              target: appName
            },
            {
              type: 'describe_pod',
              label: 'Get Pod Details',
              target: appName
            },
            {
              type: 'edit_config',
              label: 'Edit Configuration',
              target: appName
            }
          ]
        })
      });
    });
    
    // Verify diagnostic information
    await page.waitForTimeout(1000);
    const diagnosticCard = page.getByTestId('ai-diagnostic-card');
    await expect(diagnosticCard).toBeVisible();
    await expect(diagnosticCard).toContainText('CrashLoopBackOff');
    await expect(diagnosticCard).toContainText('Possible Causes');
    await expect(diagnosticCard).toContainText('Missing environment variables');
    
    // Click view logs action
    const viewLogsButton = page.getByTestId('ai-action-view-logs');
    await viewLogsButton.click();
    
    // Verify logs viewer opens
    await expect(page.getByTestId('logs-viewer')).toBeVisible();
  });

  test('AI model selection and configuration', async ({ page }) => {
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    // Open settings
    const settingsButton = page.getByTestId('ai-chat-settings-button');
    await settingsButton.click();
    
    const settingsDialog = page.getByRole('dialog');
    await expect(settingsDialog).toContainText('AI Assistant Settings');
    
    // Verify available models
    const modelSelect = settingsDialog.getByTestId('ai-model-select');
    await expect(modelSelect).toBeVisible();
    
    // Check available options
    const options = await modelSelect.locator('option').all();
    const modelNames = await Promise.all(options.map(opt => opt.textContent()));
    expect(modelNames).toContain('GPT-4');
    expect(modelNames).toContain('GPT-3.5 Turbo');
    expect(modelNames).toContain('Claude 3');
    
    // Change model
    await modelSelect.selectOption('claude-3');
    
    // Configure temperature
    const temperatureSlider = settingsDialog.getByTestId('temperature-slider');
    await temperatureSlider.fill('0.7');
    
    // Configure max tokens
    const maxTokensInput = settingsDialog.getByTestId('max-tokens-input');
    await maxTokensInput.fill('2000');
    
    // Enable features
    await settingsDialog.getByTestId('enable-code-execution-checkbox').check();
    await settingsDialog.getByTestId('enable-web-search-checkbox').check();
    
    // Save settings
    await settingsDialog.getByTestId('save-settings-button').click();
    await expectNotification(page, 'AI settings updated');
    
    // Verify settings applied
    await page.getByTestId('ai-model-indicator').click();
    await expect(page.getByTestId('current-model')).toContainText('Claude 3');
  });

  test('AI conversation history and context', async ({ page }) => {
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    // Have a multi-turn conversation
    const chatInput = page.getByTestId('chat-input');
    
    // First question
    await chatInput.fill('What are the resource limits for my current project?');
    await page.getByTestId('send-message-button').click();
    await page.waitForTimeout(1000);
    
    // Second question (context-dependent)
    await chatInput.fill('Can you increase the CPU limit to 4 cores?');
    await page.getByTestId('send-message-button').click();
    await page.waitForTimeout(1000);
    
    // Third question (reference previous)
    await chatInput.fill('What about memory? Should I increase that too?');
    await page.getByTestId('send-message-button').click();
    await page.waitForTimeout(1000);
    
    // Verify conversation maintains context
    const messages = page.locator('[data-testid^="chat-message-"]');
    await expect(messages).toHaveCount(7); // 1 welcome + 3 user + 3 AI
    
    // Open history
    const historyButton = page.getByTestId('chat-history-button');
    await historyButton.click();
    
    const historyPanel = page.getByTestId('chat-history-panel');
    await expect(historyPanel).toBeVisible();
    
    // Verify conversation sessions
    const sessions = historyPanel.locator('[data-testid^="chat-session-"]');
    await expect(sessions.first()).toContainText('Current Session');
    await expect(sessions.first()).toContainText('resource limits');
    
    // Test loading previous session
    if (await sessions.count() > 1) {
      await sessions.nth(1).click();
      await expect(page.getByTestId('chat-messages')).toBeVisible();
      await expect(page.getByTestId('session-restored-banner')).toBeVisible();
    }
    
    // Test clear conversation
    const clearButton = page.getByTestId('clear-conversation-button');
    await clearButton.click();
    
    const confirmDialog = page.getByRole('dialog');
    await confirmDialog.getByTestId('confirm-clear-button').click();
    
    // Verify conversation cleared
    await expect(page.getByTestId('ai-welcome-message')).toBeVisible();
    const remainingMessages = page.locator('[data-testid^="chat-message-"]');
    await expect(remainingMessages).toHaveCount(1); // Only welcome message
  });

  test('AI integration with LangFuse tracking', async ({ page }) => {
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    // Send a message
    const chatInput = page.getByTestId('chat-input');
    await chatInput.fill('Explain Kubernetes deployments');
    await page.getByTestId('send-message-button').click();
    
    // Mock response with LangFuse tracking
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Kubernetes Deployments are...',
          metadata: {
            model: 'gpt-4',
            tokens_used: 150,
            langfuse: {
              trace_id: 'trace-123-456',
              session_id: 'session-789',
              latency_ms: 1234,
              cost_usd: 0.003
            }
          }
        })
      });
    });
    
    await page.waitForTimeout(1000);
    
    // Open metrics/tracking view
    const metricsButton = page.getByTestId('ai-metrics-button');
    await metricsButton.click();
    
    const metricsPanel = page.getByTestId('ai-metrics-panel');
    await expect(metricsPanel).toBeVisible();
    
    // Verify LangFuse tracking data
    await expect(metricsPanel).toContainText('Total Tokens: 150');
    await expect(metricsPanel).toContainText('Latency: 1.23s');
    await expect(metricsPanel).toContainText('Cost: $0.003');
    
    // Verify trace link
    const traceLink = metricsPanel.getByTestId('langfuse-trace-link');
    await expect(traceLink).toBeVisible();
    await expect(traceLink).toHaveAttribute('href', /trace-123-456/);
    
    // Check session metrics
    await expect(metricsPanel.getByTestId('session-messages')).toContainText('2'); // 1 user + 1 AI
    await expect(metricsPanel.getByTestId('session-tokens')).toContainText('150');
  });

  test('AI-powered search and documentation', async ({ page }) => {
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    // Test documentation search
    const chatInput = page.getByTestId('chat-input');
    await chatInput.fill('Search docs for backup procedures');
    await page.getByTestId('send-message-button').click();
    
    // Mock search results
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'I found the following documentation about backup procedures:',
          search_results: [
            {
              title: 'Backup and Restore Guide',
              url: '/docs/backup-restore',
              snippet: 'Learn how to configure automated backups...',
              relevance: 0.95
            },
            {
              title: 'Disaster Recovery',
              url: '/docs/disaster-recovery',
              snippet: 'Best practices for disaster recovery...',
              relevance: 0.87
            }
          ],
          actions: [
            {
              type: 'open_docs',
              label: 'Open Backup Guide',
              url: '/docs/backup-restore'
            }
          ]
        })
      });
    });
    
    await page.waitForTimeout(1000);
    
    // Verify search results display
    const searchResults = page.getByTestId('ai-search-results');
    await expect(searchResults).toBeVisible();
    await expect(searchResults).toContainText('Backup and Restore Guide');
    await expect(searchResults).toContainText('95% relevant');
    
    // Test opening documentation
    const openDocsButton = page.getByTestId('ai-action-open-docs');
    await openDocsButton.click();
    
    // Verify docs opened in new tab or panel
    const docsPanel = page.getByTestId('docs-panel');
    if (await docsPanel.isVisible()) {
      await expect(docsPanel).toContainText('Backup and Restore Guide');
    }
  });

  test('AI chat with file and image support', async ({ page }) => {
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    // Test file upload
    const fileInput = page.getByTestId('chat-file-input');
    await fileInput.setInputFiles({
      name: 'deployment.yaml',
      mimeType: 'text/yaml',
      buffer: Buffer.from(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 3`)
    });
    
    // Verify file preview
    const filePreview = page.getByTestId('file-preview');
    await expect(filePreview).toBeVisible();
    await expect(filePreview).toContainText('deployment.yaml');
    
    // Send message with file
    const chatInput = page.getByTestId('chat-input');
    await chatInput.fill('Can you review this deployment file?');
    await page.getByTestId('send-message-button').click();
    
    // Mock AI response analyzing file
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'I\'ve analyzed your deployment file. Here are my observations:\n\n1. The deployment is set to 3 replicas\n2. Missing container specification\n3. No resource limits defined\n\nWould you like me to help complete this configuration?',
          file_analysis: {
            filename: 'deployment.yaml',
            issues: ['Missing container spec', 'No resource limits'],
            suggestions: ['Add container configuration', 'Define resource requests/limits']
          }
        })
      });
    });
    
    await page.waitForTimeout(1000);
    
    // Verify file analysis
    const analysisCard = page.getByTestId('file-analysis-card');
    await expect(analysisCard).toBeVisible();
    await expect(analysisCard).toContainText('deployment.yaml');
    await expect(analysisCard).toContainText('Missing container spec');
    
    // Test image upload
    const imageButton = page.getByTestId('chat-image-button');
    await imageButton.click();
    
    const imageInput = page.getByTestId('image-file-input');
    await imageInput.setInputFiles({
      name: 'error-screenshot.png',
      mimeType: 'image/png',
      buffer: Buffer.from('fake-image-data')
    });
    
    // Send with image
    await chatInput.fill('What does this error mean?');
    await page.getByTestId('send-message-button').click();
    
    // Verify image in chat
    const messageWithImage = page.locator('[data-testid="chat-message-user"]').last();
    await expect(messageWithImage.getByTestId('message-image')).toBeVisible();
  });

  test('AI shortcuts and quick actions', async ({ page }) => {
    // Test keyboard shortcut to open AI
    await page.keyboard.press('Control+Shift+A'); // or Cmd+Shift+A on Mac
    
    const chatWindow = page.getByTestId('ai-chat-window');
    await expect(chatWindow).toBeVisible();
    
    // Test quick action commands
    const chatInput = page.getByTestId('chat-input');
    
    // Test slash commands
    await chatInput.fill('/');
    const commandMenu = page.getByTestId('ai-command-menu');
    await expect(commandMenu).toBeVisible();
    
    // Verify available commands
    const commands = commandMenu.locator('[data-testid^="command-"]');
    await expect(commands).toHaveCount(5);
    await expect(commandMenu).toContainText('/deploy');
    await expect(commandMenu).toContainText('/troubleshoot');
    await expect(commandMenu).toContainText('/docs');
    await expect(commandMenu).toContainText('/status');
    await expect(commandMenu).toContainText('/help');
    
    // Select deploy command
    await commandMenu.getByTestId('command-deploy').click();
    
    // Verify command template inserted
    await expect(chatInput).toHaveValue('/deploy ');
    
    // Complete command
    await chatInput.fill('/deploy nginx with 3 replicas');
    await page.getByTestId('send-message-button').click();
    
    // Mock command processing
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'I\'ll help you deploy nginx with 3 replicas.',
          command: 'deploy',
          actions: [
            {
              type: 'deploy_application',
              label: 'Deploy nginx',
              config: {
                name: 'nginx-app',
                image: 'nginx:latest',
                replicas: 3
              }
            }
          ]
        })
      });
    });
    
    // Test escape to close
    await page.keyboard.press('Escape');
    await expect(chatWindow).not.toBeVisible();
  });

  test('AI error handling and fallbacks ${CRITICAL_TAG}', async ({ page }) => {
    // Open AI chat
    await page.getByTestId('ai-assistant-button').click();
    
    // Test rate limiting
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 429,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Rate limit exceeded',
          retry_after: 60
        })
      });
    });
    
    const chatInput = page.getByTestId('chat-input');
    await chatInput.fill('Help me deploy an app');
    await page.getByTestId('send-message-button').click();
    
    // Verify rate limit message
    const errorMessage = page.getByTestId('ai-error-message');
    await expect(errorMessage).toBeVisible();
    await expect(errorMessage).toContainText('Rate limit exceeded');
    await expect(errorMessage).toContainText('Please try again in 60 seconds');
    
    // Test AI service unavailable
    await page.route('**/api/ai/chat', async (route) => {
      await route.fulfill({
        status: 503,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'AI service temporarily unavailable'
        })
      });
    });
    
    await page.waitForTimeout(1000);
    await chatInput.fill('Another question');
    await page.getByTestId('send-message-button').click();
    
    // Verify fallback to documentation
    await expect(errorMessage).toContainText('AI service temporarily unavailable');
    const fallbackSuggestions = page.getByTestId('fallback-suggestions');
    await expect(fallbackSuggestions).toBeVisible();
    await expect(fallbackSuggestions).toContainText('Browse Documentation');
    await expect(fallbackSuggestions).toContainText('Contact Support');
    
    // Test timeout handling
    await page.route('**/api/ai/chat', async (route) => {
      // Simulate timeout by not responding
      await page.waitForTimeout(35000);
      await route.fulfill({
        status: 504,
        body: 'Gateway Timeout'
      });
    });
    
    await chatInput.fill('Test timeout');
    await page.getByTestId('send-message-button').click();
    
    // Verify timeout message (after 30s)
    await expect(errorMessage).toContainText('Request timed out', { timeout: 35000 });
    await expect(page.getByTestId('retry-button')).toBeVisible();
  });
});