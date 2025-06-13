import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers, testWorkspaces } from '../fixtures/mock-data';
import { generateProjectName, expectNotification } from '../utils/test-helpers';

test.describe('Serverless Functions (Knative)', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let workspacePage: WorkspacePage;
  let projectPage: ProjectPage;

  test.beforeEach(async ({ page }) => {
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    workspacePage = new WorkspacePage(page);
    projectPage = new ProjectPage(page);
    
    // Login and setup
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    // Create project for functions
    await dashboardPage.openWorkspace(testWorkspaces[0].name);
    const projectName = generateProjectName();
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
  });

  test('create HTTP triggered function', async ({ page }) => {
    // Navigate to functions tab
    await projectPage.functionsTab.click();
    
    // Create new function
    const createFunctionButton = page.getByTestId('create-function-button');
    await createFunctionButton.click();
    
    const functionDialog = page.getByRole('dialog');
    
    // Configure function basics
    await functionDialog.getByTestId('function-name-input').fill('hello-world-api');
    await functionDialog.getByTestId('function-description-input').fill('Simple HTTP API function');
    
    // Select runtime
    await functionDialog.getByTestId('runtime-select').selectOption('nodejs18');
    
    // Select trigger type
    await functionDialog.getByTestId('trigger-type-http').click();
    
    // Configure HTTP trigger
    await functionDialog.getByTestId('http-path-input').fill('/api/hello');
    await functionDialog.getByTestId('http-method-select').selectOption('GET,POST');
    
    // Add function code
    const codeEditor = functionDialog.getByTestId('function-code-editor');
    await codeEditor.click();
    await codeEditor.fill(`exports.handler = async (event, context) => {
  const name = event.queryStringParameters?.name || 'World';
  
  return {
    statusCode: 200,
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      message: \`Hello, \${name}!\`,
      timestamp: new Date().toISOString(),
      version: '1.0.0'
    })
  };
};`);
    
    // Configure resources
    await functionDialog.getByTestId('function-memory-select').selectOption('256');
    await functionDialog.getByTestId('function-timeout-input').fill('30');
    
    // Add environment variables
    await functionDialog.getByTestId('add-env-var-button').click();
    await functionDialog.getByTestId('env-key-0').fill('NODE_ENV');
    await functionDialog.getByTestId('env-value-0').fill('production');
    
    // Deploy function
    await functionDialog.getByTestId('deploy-function-button').click();
    
    // Monitor deployment
    await expect(page.getByText('Function deployment started')).toBeVisible();
    
    // Wait for deployment stages
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('deployment-stage')).toContainText('Building function');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('deployment-stage')).toContainText('Creating Knative service');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('deployment-stage')).toContainText('Configuring routes');
    
    await expectNotification(page, 'Function deployed successfully');
    
    // Verify function in list
    const functionCard = page.getByTestId('function-hello-world-api');
    await expect(functionCard).toBeVisible();
    await expect(functionCard.getByTestId('function-status')).toContainText('Ready');
    await expect(functionCard.getByTestId('function-runtime')).toContainText('Node.js 18');
    await expect(functionCard.getByTestId('function-trigger')).toContainText('HTTP');
    
    // Verify endpoint URL
    const endpointUrl = functionCard.getByTestId('function-endpoint');
    await expect(endpointUrl).toBeVisible();
    await expect(endpointUrl).toContainText('/api/hello');
  });

  test('test function invocation', async ({ page }) => {
    // Create a simple function first
    await projectPage.functionsTab.click();
    const createButton = page.getByTestId('create-function-button');
    await createButton.click();
    
    const dialog = page.getByRole('dialog');
    await dialog.getByTestId('function-name-input').fill('test-function');
    await dialog.getByTestId('runtime-select').selectOption('nodejs18');
    await dialog.getByTestId('trigger-type-http').click();
    await dialog.getByTestId('http-path-input').fill('/test');
    
    const codeEditor = dialog.getByTestId('function-code-editor');
    await codeEditor.fill(`exports.handler = async (event) => {
  return {
    statusCode: 200,
    body: JSON.stringify({ result: 'success', input: event.body })
  };
};`);
    
    await dialog.getByTestId('deploy-function-button').click();
    await page.waitForTimeout(5000);
    
    // Open function details
    const functionCard = page.getByTestId('function-test-function');
    await functionCard.click();
    
    // Go to test tab
    const testTab = page.getByRole('tab', { name: /test/i });
    await testTab.click();
    
    // Configure test invocation
    const testPayload = page.getByTestId('test-payload-editor');
    await testPayload.fill(JSON.stringify({
      name: 'Test User',
      action: 'greet'
    }, null, 2));
    
    // Add headers
    await page.getByTestId('add-header-button').click();
    await page.getByTestId('header-key-0').fill('Content-Type');
    await page.getByTestId('header-value-0').fill('application/json');
    
    // Invoke function
    const invokeButton = page.getByTestId('invoke-function-button');
    await invokeButton.click();
    
    // Check response
    await expect(page.getByTestId('invocation-status')).toContainText('200');
    
    const responseBody = page.getByTestId('response-body');
    await expect(responseBody).toBeVisible();
    await expect(responseBody).toContainText('success');
    
    // Check execution details
    const executionTime = page.getByTestId('execution-time');
    await expect(executionTime).toBeVisible();
    await expect(executionTime).toContainText('ms');
    
    // Check logs
    const logsTab = page.getByRole('tab', { name: /logs/i });
    await logsTab.click();
    
    const functionLogs = page.getByTestId('function-logs');
    await expect(functionLogs).toBeVisible();
  });

  test('create event-triggered function', async ({ page }) => {
    await projectPage.functionsTab.click();
    
    const createButton = page.getByTestId('create-function-button');
    await createButton.click();
    
    const dialog = page.getByRole('dialog');
    
    // Configure function
    await dialog.getByTestId('function-name-input').fill('event-processor');
    await dialog.getByTestId('runtime-select').selectOption('python39');
    await dialog.getByTestId('trigger-type-event').click();
    
    // Configure event trigger
    await dialog.getByTestId('event-source-select').selectOption('custom');
    await dialog.getByTestId('event-type-input').fill('user.created');
    await dialog.getByTestId('event-filter-input').fill('$.data.type == "premium"');
    
    // Add Python code
    const codeEditor = dialog.getByTestId('function-code-editor');
    await codeEditor.fill(`import json
import logging

logger = logging.getLogger()

def handler(event, context):
    logger.info(f"Processing event: {event}")
    
    try:
        # Parse CloudEvent
        event_data = json.loads(event.get('body', '{}'))
        user_data = event_data.get('data', {})
        
        # Process premium user creation
        result = {
            'user_id': user_data.get('id'),
            'status': 'processed',
            'actions': [
                'welcome_email_sent',
                'premium_features_enabled',
                'onboarding_scheduled'
            ]
        }
        
        return {
            'statusCode': 200,
            'body': json.dumps(result)
        }
    except Exception as e:
        logger.error(f"Error processing event: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps({'error': str(e)})
        }
`);
    
    // Configure resources for event processing
    await dialog.getByTestId('function-memory-select').selectOption('512');
    await dialog.getByTestId('function-timeout-input').fill('60');
    await dialog.getByTestId('concurrent-executions-input').fill('10');
    
    await dialog.getByTestId('deploy-function-button').click();
    
    await page.waitForTimeout(5000);
    await expectNotification(page, 'Function deployed successfully');
    
    // Verify event subscription created
    const functionCard = page.getByTestId('function-event-processor');
    await expect(functionCard.getByTestId('function-trigger')).toContainText('Event: user.created');
  });

  test('create scheduled function (CronJob integration)', async ({ page }) => {
    await projectPage.functionsTab.click();
    
    const createButton = page.getByTestId('create-function-button');
    await createButton.click();
    
    const dialog = page.getByRole('dialog');
    
    // Configure scheduled function
    await dialog.getByTestId('function-name-input').fill('daily-report-generator');
    await dialog.getByTestId('runtime-select').selectOption('python39');
    await dialog.getByTestId('trigger-type-schedule').click();
    
    // Configure schedule
    await dialog.getByTestId('schedule-input').fill('0 8 * * *');
    await dialog.getByTestId('schedule-helper').selectOption('daily-8am');
    await dialog.getByTestId('timezone-select').selectOption('America/New_York');
    
    // Add function code
    const codeEditor = dialog.getByTestId('function-code-editor');
    await codeEditor.fill(`import json
from datetime import datetime, timedelta
import boto3  # For S3 upload

def handler(event, context):
    print(f"Generating daily report at {datetime.now()}")
    
    # Generate report data
    report_data = {
        'date': datetime.now().strftime('%Y-%m-%d'),
        'metrics': {
            'total_users': 1542,
            'active_users': 892,
            'new_signups': 45,
            'revenue': 12450.00
        },
        'generated_at': datetime.now().isoformat()
    }
    
    # Save report (mock)
    report_name = f"daily-report-{report_data['date']}.json"
    
    return {
        'statusCode': 200,
        'body': json.dumps({
            'report': report_name,
            'status': 'generated',
            'next_run': (datetime.now() + timedelta(days=1)).strftime('%Y-%m-%d 08:00:00')
        })
    }
`);
    
    await dialog.getByTestId('deploy-function-button').click();
    
    await page.waitForTimeout(5000);
    await expectNotification(page, 'Scheduled function deployed successfully');
    
    // Verify CronJob created
    await projectPage.cronJobsTab.click();
    
    const cronJobCard = page.getByTestId('cronjob-daily-report-generator');
    await expect(cronJobCard).toBeVisible();
    await expect(cronJobCard).toContainText('0 8 * * *');
    await expect(cronJobCard).toContainText('Function Trigger');
  });

  test('version management and rollback', async ({ page }) => {
    // Create initial function
    await projectPage.functionsTab.click();
    
    const createButton = page.getByTestId('create-function-button');
    await createButton.click();
    
    const dialog = page.getByRole('dialog');
    await dialog.getByTestId('function-name-input').fill('versioned-api');
    await dialog.getByTestId('runtime-select').selectOption('nodejs18');
    await dialog.getByTestId('trigger-type-http').click();
    await dialog.getByTestId('http-path-input').fill('/api/data');
    
    // Initial version
    const codeEditor = dialog.getByTestId('function-code-editor');
    await codeEditor.fill(`exports.handler = async (event) => {
  return {
    statusCode: 200,
    body: JSON.stringify({ version: '1.0.0', data: 'initial' })
  };
};`);
    
    await dialog.getByTestId('deploy-function-button').click();
    await page.waitForTimeout(5000);
    
    // Open function details
    const functionCard = page.getByTestId('function-versioned-api');
    await functionCard.click();
    
    // Deploy new version
    const updateButton = page.getByTestId('update-function-button');
    await updateButton.click();
    
    const updateDialog = page.getByRole('dialog');
    const updateEditor = updateDialog.getByTestId('function-code-editor');
    await updateEditor.clear();
    await updateEditor.fill(`exports.handler = async (event) => {
  return {
    statusCode: 200,
    body: JSON.stringify({ version: '2.0.0', data: 'updated', features: ['new'] })
  };
};`);
    
    await updateDialog.getByTestId('version-description-input').fill('Added new features');
    await updateDialog.getByTestId('deploy-update-button').click();
    
    await page.waitForTimeout(5000);
    await expectNotification(page, 'Function updated to version 2');
    
    // Check versions tab
    const versionsTab = page.getByRole('tab', { name: /versions/i });
    await versionsTab.click();
    
    const versionsList = page.getByTestId('versions-list');
    const versions = versionsList.locator('[data-testid^="version-item-"]');
    await expect(versions).toHaveCount(2);
    
    // Verify current version
    const v2Item = versions.first();
    await expect(v2Item).toContainText('v2');
    await expect(v2Item).toContainText('Active');
    await expect(v2Item).toContainText('Added new features');
    
    // Test traffic splitting
    const trafficSplitButton = page.getByTestId('configure-traffic-split-button');
    await trafficSplitButton.click();
    
    const trafficDialog = page.getByRole('dialog');
    await trafficDialog.getByTestId('enable-traffic-split-toggle').check();
    await trafficDialog.getByTestId('v1-traffic-slider').fill('20');
    await trafficDialog.getByTestId('v2-traffic-slider').fill('80');
    await trafficDialog.getByTestId('save-traffic-split-button').click();
    
    await expectNotification(page, 'Traffic split configured');
    
    // Rollback to v1
    const v1Item = versions.nth(1);
    const rollbackButton = v1Item.getByTestId('rollback-button');
    await rollbackButton.click();
    
    const rollbackDialog = page.getByRole('dialog');
    await rollbackDialog.getByTestId('confirm-rollback-button').click();
    
    await expectNotification(page, 'Function rolled back to version 1');
  });

  test('function with AI model integration', async ({ page }) => {
    await projectPage.functionsTab.click();
    
    const createButton = page.getByTestId('create-function-button');
    await createButton.click();
    
    const dialog = page.getByRole('dialog');
    
    // Configure AI function
    await dialog.getByTestId('function-name-input').fill('ai-text-analyzer');
    await dialog.getByTestId('runtime-select').selectOption('python39');
    await dialog.getByTestId('trigger-type-http').click();
    await dialog.getByTestId('http-path-input').fill('/api/analyze');
    
    // Enable AI features
    await dialog.getByTestId('enable-ai-toggle').check();
    await dialog.getByTestId('ai-model-select').selectOption('gpt-3.5-turbo');
    await dialog.getByTestId('ai-provider-select').selectOption('openai');
    
    // Add AI-enabled code
    const codeEditor = dialog.getByTestId('function-code-editor');
    await codeEditor.fill(`import json
import os
from openai import OpenAI

# AI client is injected by the platform
client = OpenAI(api_key=os.environ.get('OPENAI_API_KEY'))

def handler(event, context):
    try:
        # Parse request
        body = json.loads(event.get('body', '{}'))
        text = body.get('text', '')
        
        # Analyze text with AI
        response = client.chat.completions.create(
            model="gpt-3.5-turbo",
            messages=[
                {"role": "system", "content": "You are a text analysis assistant. Analyze the sentiment and key topics."},
                {"role": "user", "content": text}
            ],
            temperature=0.3,
            max_tokens=500
        )
        
        analysis = response.choices[0].message.content
        
        return {
            'statusCode': 200,
            'body': json.stringify({
                'analysis': analysis,
                'model': 'gpt-3.5-turbo',
                'tokens_used': response.usage.total_tokens
            })
        }
    except Exception as e:
        return {
            'statusCode': 500,
            'body': json.stringify({'error': str(e)})
        }
`);
    
    // Configure AI resources
    await dialog.getByTestId('function-memory-select').selectOption('1024');
    await dialog.getByTestId('ai-timeout-input').fill('120');
    await dialog.getByTestId('ai-max-tokens-input').fill('1000');
    
    await dialog.getByTestId('deploy-function-button').click();
    
    await page.waitForTimeout(5000);
    await expectNotification(page, 'AI function deployed successfully');
    
    // Verify AI badge
    const functionCard = page.getByTestId('function-ai-text-analyzer');
    await expect(functionCard.getByTestId('ai-enabled-badge')).toBeVisible();
    await expect(functionCard.getByTestId('ai-model-badge')).toContainText('GPT-3.5');
  });

  test('function monitoring and metrics', async ({ page }) => {
    // Assume function exists
    await projectPage.functionsTab.click();
    
    // Mock existing function
    await page.route('**/api/organizations/*/workspaces/*/projects/*/functions', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          functions: [{
            id: 'func-123',
            name: 'metrics-test-function',
            runtime: 'nodejs18',
            status: 'ready',
            invocations_24h: 15420,
            avg_duration_ms: 124,
            error_rate: 0.02,
          }],
        }),
      });
    });
    
    await page.reload();
    
    // Open function metrics
    const functionCard = page.getByTestId('function-metrics-test-function');
    await functionCard.click();
    
    const metricsTab = page.getByRole('tab', { name: /metrics/i });
    await metricsTab.click();
    
    // Verify metrics dashboard
    const metricsPanel = page.getByTestId('function-metrics-panel');
    await expect(metricsPanel).toBeVisible();
    
    // Check invocation metrics
    await expect(metricsPanel.getByTestId('total-invocations')).toContainText('15,420');
    await expect(metricsPanel.getByTestId('avg-duration')).toContainText('124ms');
    await expect(metricsPanel.getByTestId('error-rate')).toContainText('2%');
    
    // Check cold start metrics
    await expect(metricsPanel.getByTestId('cold-starts')).toBeVisible();
    await expect(metricsPanel.getByTestId('cold-start-duration')).toBeVisible();
    
    // Configure alerts
    const configureAlertsButton = page.getByTestId('configure-function-alerts-button');
    await configureAlertsButton.click();
    
    const alertDialog = page.getByRole('dialog');
    
    // Add error rate alert
    await alertDialog.getByTestId('add-alert-button').click();
    await alertDialog.getByTestId('alert-metric-select').selectOption('error_rate');
    await alertDialog.getByTestId('alert-operator-select').selectOption('greater_than');
    await alertDialog.getByTestId('alert-threshold-input').fill('5');
    await alertDialog.getByTestId('alert-duration-input').fill('5');
    
    // Add latency alert
    await alertDialog.getByTestId('add-alert-button').click();
    await alertDialog.getByTestId('alert-metric-select-1').selectOption('avg_duration');
    await alertDialog.getByTestId('alert-operator-select-1').selectOption('greater_than');
    await alertDialog.getByTestId('alert-threshold-input-1').fill('1000');
    
    await alertDialog.getByTestId('save-alerts-button').click();
    
    await expectNotification(page, 'Function alerts configured');
  });

  test('function with custom dependencies', async ({ page }) => {
    await projectPage.functionsTab.click();
    
    const createButton = page.getByTestId('create-function-button');
    await createButton.click();
    
    const dialog = page.getByRole('dialog');
    
    // Configure function with dependencies
    await dialog.getByTestId('function-name-input').fill('data-processor');
    await dialog.getByTestId('runtime-select').selectOption('python39');
    await dialog.getByTestId('trigger-type-http').click();
    
    // Add dependencies
    const depsTab = dialog.getByRole('tab', { name: /dependencies/i });
    await depsTab.click();
    
    const requirementsEditor = dialog.getByTestId('requirements-editor');
    await requirementsEditor.fill(`pandas==1.5.3
numpy==1.24.3
scikit-learn==1.2.2
requests==2.28.2
boto3==1.26.137`);
    
    // Add custom build commands
    await dialog.getByTestId('enable-build-commands-toggle').check();
    const buildCommandsEditor = dialog.getByTestId('build-commands-editor');
    await buildCommandsEditor.fill(`pip install --upgrade pip
apt-get update && apt-get install -y libgomp1`);
    
    // Back to code tab
    const codeTab = dialog.getByRole('tab', { name: /code/i });
    await codeTab.click();
    
    const codeEditor = dialog.getByTestId('function-code-editor');
    await codeEditor.fill(`import pandas as pd
import numpy as np
from sklearn.preprocessing import StandardScaler
import json

def handler(event, context):
    try:
        # Parse input data
        body = json.loads(event.get('body', '{}'))
        data = body.get('data', [])
        
        # Create DataFrame
        df = pd.DataFrame(data)
        
        # Process data
        scaler = StandardScaler()
        numeric_columns = df.select_dtypes(include=[np.number]).columns
        df[numeric_columns] = scaler.fit_transform(df[numeric_columns])
        
        # Return processed data
        result = {
            'processed_rows': len(df),
            'columns': df.columns.tolist(),
            'summary': df.describe().to_dict()
        }
        
        return {
            'statusCode': 200,
            'body': json.dumps(result)
        }
    except Exception as e:
        return {
            'statusCode': 500,
            'body': json.dumps({'error': str(e)})
        }
`);
    
    await dialog.getByTestId('deploy-function-button').click();
    
    // Monitor build process
    await expect(page.getByText('Building function with dependencies')).toBeVisible();
    
    await page.waitForTimeout(3000);
    await expect(page.getByTestId('build-stage')).toContainText('Installing dependencies');
    
    await page.waitForTimeout(3000);
    await expect(page.getByTestId('build-stage')).toContainText('Running build commands');
    
    await page.waitForTimeout(3000);
    await expectNotification(page, 'Function deployed with custom dependencies');
  });

  test('delete function with cleanup', async ({ page }) => {
    // Create a function first
    await projectPage.functionsTab.click();
    
    const createButton = page.getByTestId('create-function-button');
    await createButton.click();
    
    const dialog = page.getByRole('dialog');
    await dialog.getByTestId('function-name-input').fill('temp-function');
    await dialog.getByTestId('runtime-select').selectOption('nodejs18');
    await dialog.getByTestId('trigger-type-http').click();
    await dialog.getByTestId('http-path-input').fill('/temp');
    
    const codeEditor = dialog.getByTestId('function-code-editor');
    await codeEditor.fill(`exports.handler = async () => ({ statusCode: 200, body: 'temp' });`);
    
    await dialog.getByTestId('deploy-function-button').click();
    await page.waitForTimeout(5000);
    
    // Open function details
    const functionCard = page.getByTestId('function-temp-function');
    await functionCard.click();
    
    // Delete function
    const deleteButton = page.getByTestId('delete-function-button');
    await deleteButton.click();
    
    const deleteDialog = page.getByRole('dialog');
    await expect(deleteDialog).toContainText('Delete Function');
    await expect(deleteDialog).toContainText('This will permanently delete the function and all its versions');
    
    // Check cleanup options
    await expect(deleteDialog.getByTestId('cleanup-logs-checkbox')).toBeChecked();
    await expect(deleteDialog.getByTestId('cleanup-metrics-checkbox')).toBeChecked();
    
    // Confirm deletion
    await deleteDialog.getByTestId('confirm-delete-input').fill('temp-function');
    await deleteDialog.getByTestId('confirm-delete-button').click();
    
    // Monitor deletion
    await expect(page.getByText('Deleting function')).toBeVisible();
    
    await page.waitForTimeout(2000);
    await expectNotification(page, 'Function deleted successfully');
    
    // Verify redirect to functions list
    await expect(page).toHaveURL(/.*\/projects\/.*\/functions/);
    
    // Verify function removed
    await expect(page.getByTestId('function-temp-function')).not.toBeVisible();
  });
});