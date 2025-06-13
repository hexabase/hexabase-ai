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
  deployFunction,
  invokeFunction,
  verifyResourceStatus,
  expectNotification,
  expectNoErrors,
  generateProjectName,
  generateFunctionName,
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

// Mock code editor component (since Monaco Editor might not be used)
jest.mock('@/components/code-editor', () => ({
  CodeEditor: ({ value, onChange }: any) => (
    <textarea
      data-testid="code-editor"
      value={value}
      onChange={(e: any) => onChange(e.target.value)}
    />
  ),
}));

describe('Function Development and Deployment', () => {
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

  it('should write, deploy, and execute serverless functions', async () => {
    renderApp();

    // Initial setup
    await loginUser();
    await createOrganization('Function Development Org');
    await createWorkspace('Serverless Workspace', 'shared');
    const projectName = generateProjectName();
    await createProject(projectName);

    // Step 1: Navigate to functions tab
    console.log('Step 1: Navigating to functions...');
    const functionsTab = screen.getByRole('tab', { name: /functions/i });
    await userEvent.click(functionsTab);

    // Step 2: Create a new function
    console.log('Step 2: Creating new function...');
    const createFunctionButton = screen.getByRole('button', { name: /create function/i });
    await userEvent.click(createFunctionButton);

    const modal = screen.getByRole('dialog');
    
    // Fill in function details
    const nameInput = within(modal).getByLabelText(/function name/i);
    await userEvent.type(nameInput, 'image-processor');

    const descriptionInput = within(modal).getByLabelText(/description/i);
    await userEvent.type(descriptionInput, 'Processes and optimizes uploaded images');

    const runtimeSelect = within(modal).getByLabelText(/runtime/i);
    await userEvent.selectOptions(runtimeSelect, 'nodejs18');

    const memorySelect = within(modal).getByLabelText(/memory/i);
    await userEvent.selectOptions(memorySelect, '512');

    const timeoutInput = within(modal).getByLabelText(/timeout/i);
    await userEvent.clear(timeoutInput);
    await userEvent.type(timeoutInput, '60');

    // Add environment variables
    const addEnvButton = within(modal).getByRole('button', { name: /add variable/i });
    await userEvent.click(addEnvButton);

    const envKeyInput = within(modal).getByPlaceholderText(/key/i);
    await userEvent.type(envKeyInput, 'IMAGE_BUCKET');

    const envValueInput = within(modal).getByPlaceholderText(/value/i);
    await userEvent.type(envValueInput, 'processed-images');

    // Select triggers
    const httpTrigger = within(modal).getByLabelText(/http trigger/i);
    await userEvent.click(httpTrigger);

    const eventTrigger = within(modal).getByLabelText(/event trigger/i);
    await userEvent.click(eventTrigger);

    const createButton = within(modal).getByRole('button', { name: /create/i });
    await userEvent.click(createButton);

    await expectNotification(/function created/i);

    // Step 3: Write function code
    console.log('Step 3: Writing function code...');
    const functionRow = screen.getByText('image-processor').closest('tr');
    const editButton = within(functionRow!).getByRole('button', { name: /edit/i });
    await userEvent.click(editButton);

    // Wait for code editor to load
    await waitFor(() => {
      expect(screen.getByTestId('code-editor')).toBeInTheDocument();
    });

    const codeEditor = screen.getByTestId('code-editor') as HTMLTextAreaElement;
    
    // Write the function code
    const functionCode = `const sharp = require('sharp');
const AWS = require('aws-sdk');

const s3 = new AWS.S3();

exports.handler = async (event) => {
  console.log('Processing image:', event.imageUrl);
  
  try {
    // Download image from URL
    const response = await fetch(event.imageUrl);
    const buffer = await response.buffer();
    
    // Process image with sharp
    const processed = await sharp(buffer)
      .resize(800, 600, { fit: 'inside' })
      .jpeg({ quality: 80 })
      .toBuffer();
    
    // Upload to S3
    const key = \`processed/\${Date.now()}_\${event.fileName}\`;
    await s3.putObject({
      Bucket: process.env.IMAGE_BUCKET,
      Key: key,
      Body: processed,
      ContentType: 'image/jpeg'
    }).promise();
    
    return {
      statusCode: 200,
      body: JSON.stringify({
        message: 'Image processed successfully',
        url: \`https://\${process.env.IMAGE_BUCKET}.s3.amazonaws.com/\${key}\`
      })
    };
  } catch (error) {
    console.error('Error processing image:', error);
    return {
      statusCode: 500,
      body: JSON.stringify({
        error: 'Failed to process image'
      })
    };
  }
};`;

    await userEvent.clear(codeEditor);
    await userEvent.type(codeEditor, functionCode);

    // Save the code
    const saveButton = screen.getByRole('button', { name: /save/i });
    await userEvent.click(saveButton);

    await expectNotification(/code saved/i);

    // Step 4: Deploy the function
    console.log('Step 4: Deploying function...');
    const deployButton = screen.getByRole('button', { name: /deploy/i });
    await userEvent.click(deployButton);

    const deployModal = screen.getByRole('dialog');
    const versionInput = within(deployModal).getByLabelText(/version tag/i);
    await userEvent.type(versionInput, 'v1.0.0');

    const deployConfirmButton = within(deployModal).getByRole('button', { name: /deploy/i });
    await userEvent.click(deployConfirmButton);

    // Monitor deployment progress
    await waitFor(() => {
      const deploymentStatus = screen.getByTestId('deployment-status');
      expect(deploymentStatus).toHaveTextContent(/deploying/i);
    });

    // Wait for deployment to complete
    await waitFor(() => {
      const deploymentStatus = screen.getByTestId('deployment-status');
      expect(deploymentStatus).toHaveTextContent(/active/i);
    }, { timeout: 15000 });

    await expectNotification(/function deployed successfully/i);

    // Step 5: Test the function
    console.log('Step 5: Testing function execution...');
    const testButton = screen.getByRole('button', { name: /test function/i });
    await userEvent.click(testButton);

    const testModal = screen.getByRole('dialog');
    
    // Select test event
    const eventTypeSelect = within(testModal).getByLabelText(/event type/i);
    await userEvent.selectOptions(eventTypeSelect, 'http');

    // Provide test payload
    const payloadEditor = within(testModal).getByLabelText(/test payload/i);
    const testPayload = JSON.stringify({
      imageUrl: 'https://example.com/test-image.jpg',
      fileName: 'test-image.jpg'
    }, null, 2);
    
    await userEvent.clear(payloadEditor);
    await userEvent.type(payloadEditor, testPayload);

    const runTestButton = within(testModal).getByRole('button', { name: /run test/i });
    await userEvent.click(runTestButton);

    // Wait for test results
    await waitFor(() => {
      const resultSection = screen.getByTestId('test-results');
      expect(resultSection).toBeInTheDocument();
    });

    // Verify successful execution
    const resultStatus = screen.getByTestId('result-status');
    expect(resultStatus).toHaveTextContent(/200/i);

    const resultBody = screen.getByTestId('result-body');
    expect(resultBody).toHaveTextContent(/image processed successfully/i);

    // Check execution logs
    const logsTab = within(testModal).getByRole('tab', { name: /logs/i });
    await userEvent.click(logsTab);

    expect(screen.getByText(/processing image.*test-image\.jpg/i)).toBeInTheDocument();

    const closeTestButton = within(testModal).getByRole('button', { name: /close/i });
    await userEvent.click(closeTestButton);

    // Step 6: Set up scheduled execution
    console.log('Step 6: Setting up scheduled execution...');
    const scheduleButton = screen.getByRole('button', { name: /configure schedule/i });
    await userEvent.click(scheduleButton);

    const scheduleModal = screen.getByRole('dialog');
    
    const enableSchedule = within(scheduleModal).getByLabelText(/enable scheduled execution/i);
    await userEvent.click(enableSchedule);

    const cronInput = within(scheduleModal).getByLabelText(/cron expression/i);
    await userEvent.type(cronInput, '0 */4 * * *'); // Every 4 hours

    const schedulePayloadEditor = within(scheduleModal).getByLabelText(/default payload/i);
    const schedulePayload = JSON.stringify({
      source: 'scheduled',
      processAll: true
    }, null, 2);
    
    await userEvent.type(schedulePayloadEditor, schedulePayload);

    const saveScheduleButton = within(scheduleModal).getByRole('button', { name: /save/i });
    await userEvent.click(saveScheduleButton);

    await expectNotification(/schedule configured/i);

    // Step 7: View function metrics
    console.log('Step 7: Viewing function metrics...');
    const metricsButton = screen.getByRole('button', { name: /view metrics/i });
    await userEvent.click(metricsButton);

    await waitFor(() => {
      expect(screen.getByText(/invocations/i)).toBeInTheDocument();
      expect(screen.getByText(/average duration/i)).toBeInTheDocument();
      expect(screen.getByText(/error rate/i)).toBeInTheDocument();
    });

    console.log('✅ Function development and deployment completed successfully!');
  });

  it('should handle function versioning and rollback', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Versioning Test Org');
    await createWorkspace('Version Control', 'shared');
    await createProject('versioning-project');

    // Create and deploy initial function
    const functionName = 'data-transformer';
    await deployFunction(functionName, 'python39', `
def handler(event, context):
    # Version 1.0.0
    data = event.get('data', [])
    transformed = [item.upper() for item in data]
    return {
        'statusCode': 200,
        'body': {'result': transformed, 'version': '1.0.0'}
    }
`, 'handler');

    // Step 1: Deploy a new version
    console.log('Step 1: Deploying new version...');
    const functionsTab = screen.getByRole('tab', { name: /functions/i });
    await userEvent.click(functionsTab);

    const functionRow = screen.getByText(functionName).closest('tr');
    const editButton = within(functionRow!).getByRole('button', { name: /edit/i });
    await userEvent.click(editButton);

    // Update the code
    const codeEditor = screen.getByTestId('code-editor') as HTMLTextAreaElement;
    const updatedCode = `
def handler(event, context):
    # Version 2.0.0 - Added sorting
    data = event.get('data', [])
    transformed = sorted([item.upper() for item in data])
    return {
        'statusCode': 200,
        'body': {'result': transformed, 'version': '2.0.0'}
    }
`;
    
    await userEvent.clear(codeEditor);
    await userEvent.type(codeEditor, updatedCode);

    const saveButton = screen.getByRole('button', { name: /save/i });
    await userEvent.click(saveButton);

    // Deploy as new version
    const deployButton = screen.getByRole('button', { name: /deploy/i });
    await userEvent.click(deployButton);

    const deployModal = screen.getByRole('dialog');
    const versionInput = within(deployModal).getByLabelText(/version tag/i);
    await userEvent.type(versionInput, 'v2.0.0');

    const deployConfirmButton = within(deployModal).getByRole('button', { name: /deploy/i });
    await userEvent.click(deployConfirmButton);

    await waitFor(() => {
      expect(screen.getByTestId('deployment-status')).toHaveTextContent(/active/i);
    });

    // Step 2: View version history
    console.log('Step 2: Viewing version history...');
    const versionsButton = screen.getByRole('button', { name: /versions/i });
    await userEvent.click(versionsButton);

    const versionsModal = screen.getByRole('dialog');
    
    // Verify both versions are listed
    expect(within(versionsModal).getByText('v2.0.0')).toBeInTheDocument();
    expect(within(versionsModal).getByText(/active/i)).toBeInTheDocument();
    expect(within(versionsModal).getByText('v1.0.0')).toBeInTheDocument();

    // Step 3: Test the new version
    console.log('Step 3: Testing new version...');
    const testV2Button = within(versionsModal).getAllByRole('button', { name: /test/i })[0];
    await userEvent.click(testV2Button);

    const testPayload = JSON.stringify({
      data: ['zebra', 'alpha', 'beta']
    }, null, 2);

    const payloadEditor = screen.getByLabelText(/test payload/i);
    await userEvent.clear(payloadEditor);
    await userEvent.type(payloadEditor, testPayload);

    const runTestButton = screen.getByRole('button', { name: /run test/i });
    await userEvent.click(runTestButton);

    // Verify v2 behavior (sorted output)
    await waitFor(() => {
      const resultBody = screen.getByTestId('result-body');
      expect(resultBody).toHaveTextContent(/ALPHA.*BETA.*ZEBRA/);
      expect(resultBody).toHaveTextContent(/version.*2\.0\.0/);
    });

    const closeTestButton = screen.getByRole('button', { name: /close/i });
    await userEvent.click(closeTestButton);

    // Step 4: Simulate issue and rollback
    console.log('Step 4: Simulating rollback scenario...');
    
    // Mock function errors for v2
    mockApiClient.functions.invoke.mockImplementationOnce(async () => ({
      invocation_id: 'inv-error',
      function_id: 'func-1',
      status: 'error',
      trigger_type: 'http',
      error: 'Function timeout',
      duration_ms: 60000,
      started_at: new Date().toISOString(),
      completed_at: new Date(Date.now() + 60000).toISOString(),
    }));

    // Notice errors in monitoring
    const alertBadge = screen.getByTestId('error-alert');
    expect(alertBadge).toHaveTextContent(/1 error/i);

    // Initiate rollback
    const rollbackButton = within(versionsModal).getByRole('button', { name: /rollback/i });
    await userEvent.click(rollbackButton);

    const rollbackModal = screen.getByRole('dialog', { name: /confirm rollback/i });
    const rollbackToSelect = within(rollbackModal).getByLabelText(/rollback to version/i);
    await userEvent.selectOptions(rollbackToSelect, 'v1.0.0');

    const confirmRollbackButton = within(rollbackModal).getByRole('button', { name: /rollback/i });
    await userEvent.click(confirmRollbackButton);

    await expectNotification(/rollback initiated/i);

    // Wait for rollback to complete
    await waitFor(() => {
      const activeVersion = screen.getByTestId('active-version');
      expect(activeVersion).toHaveTextContent('v1.0.0');
    });

    await expectNotification(/rollback completed/i);

    console.log('✅ Function versioning and rollback tested successfully!');
  });

  it('should support multiple runtimes and dependencies', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Multi-Runtime Org');
    await createWorkspace('Runtime Testing', 'shared');
    await createProject('runtime-project');

    // Step 1: Create Node.js function with dependencies
    console.log('Step 1: Creating Node.js function with dependencies...');
    const functionsTab = screen.getByRole('tab', { name: /functions/i });
    await userEvent.click(functionsTab);

    const createButton = screen.getByRole('button', { name: /create function/i });
    await userEvent.click(createButton);

    const modal = screen.getByRole('dialog');
    
    const nameInput = within(modal).getByLabelText(/function name/i);
    await userEvent.type(nameInput, 'crypto-hasher');

    const runtimeSelect = within(modal).getByLabelText(/runtime/i);
    await userEvent.selectOptions(runtimeSelect, 'nodejs18');

    // Add dependencies
    const depsTab = within(modal).getByRole('tab', { name: /dependencies/i });
    await userEvent.click(depsTab);

    const addDepButton = within(modal).getByRole('button', { name: /add dependency/i });
    await userEvent.click(addDepButton);

    const depNameInput = within(modal).getByPlaceholderText(/package name/i);
    await userEvent.type(depNameInput, 'bcrypt');

    const depVersionInput = within(modal).getByPlaceholderText(/version/i);
    await userEvent.type(depVersionInput, '^5.1.0');

    // Add another dependency
    await userEvent.click(addDepButton);
    const dep2Inputs = within(modal).getAllByPlaceholderText(/package name/i);
    await userEvent.type(dep2Inputs[1], 'jsonwebtoken');

    const dep2VersionInputs = within(modal).getAllByPlaceholderText(/version/i);
    await userEvent.type(dep2VersionInputs[1], '^9.0.0');

    const createFuncButton = within(modal).getByRole('button', { name: /create/i });
    await userEvent.click(createFuncButton);

    // Step 2: Create Python function
    console.log('Step 2: Creating Python function...');
    await userEvent.click(createButton);

    const pyModal = screen.getByRole('dialog');
    
    const pyNameInput = within(pyModal).getByLabelText(/function name/i);
    await userEvent.type(pyNameInput, 'ml-predictor');

    const pyRuntimeSelect = within(pyModal).getByLabelText(/runtime/i);
    await userEvent.selectOptions(pyRuntimeSelect, 'python39');

    // Python dependencies
    const pyDepsTab = within(pyModal).getByRole('tab', { name: /dependencies/i });
    await userEvent.click(pyDepsTab);

    const pyAddDepButton = within(pyModal).getByRole('button', { name: /add dependency/i });
    await userEvent.click(pyAddDepButton);

    const pyDepInput = within(pyModal).getByPlaceholderText(/package name/i);
    await userEvent.type(pyDepInput, 'numpy==1.24.0');

    await userEvent.click(pyAddDepButton);
    const pyDep2Inputs = within(pyModal).getAllByPlaceholderText(/package name/i);
    await userEvent.type(pyDep2Inputs[1], 'scikit-learn==1.3.0');

    const createPyButton = within(pyModal).getByRole('button', { name: /create/i });
    await userEvent.click(createPyButton);

    // Step 3: Create Go function
    console.log('Step 3: Creating Go function...');
    await userEvent.click(createButton);

    const goModal = screen.getByRole('dialog');
    
    const goNameInput = within(goModal).getByLabelText(/function name/i);
    await userEvent.type(goNameInput, 'api-gateway');

    const goRuntimeSelect = within(goModal).getByLabelText(/runtime/i);
    await userEvent.selectOptions(goRuntimeSelect, 'go119');

    const createGoButton = within(goModal).getByRole('button', { name: /create/i });
    await userEvent.click(createGoButton);

    // Verify all functions are listed with correct runtimes
    await waitFor(() => {
      const nodeFunc = screen.getByText('crypto-hasher').closest('tr');
      expect(within(nodeFunc!).getByText(/node\.js 18/i)).toBeInTheDocument();
      
      const pyFunc = screen.getByText('ml-predictor').closest('tr');
      expect(within(pyFunc!).getByText(/python 3\.9/i)).toBeInTheDocument();
      
      const goFunc = screen.getByText('api-gateway').closest('tr');
      expect(within(goFunc!).getByText(/go 1\.19/i)).toBeInTheDocument();
    });

    // Step 4: Test runtime-specific features
    console.log('Step 4: Testing runtime-specific features...');
    
    // Test Node.js async/await support
    const nodeRow = screen.getByText('crypto-hasher').closest('tr');
    const nodeEditButton = within(nodeRow!).getByRole('button', { name: /edit/i });
    await userEvent.click(nodeEditButton);

    const nodeEditor = screen.getByTestId('code-editor') as HTMLTextAreaElement;
    const nodeCode = `
const bcrypt = require('bcrypt');
const jwt = require('jsonwebtoken');

exports.handler = async (event) => {
  const { password, action } = event;
  
  if (action === 'hash') {
    const hash = await bcrypt.hash(password, 10);
    return { statusCode: 200, body: { hash } };
  } else if (action === 'token') {
    const token = jwt.sign({ user: 'test' }, 'secret', { expiresIn: '1h' });
    return { statusCode: 200, body: { token } };
  }
};`;

    await userEvent.clear(nodeEditor);
    await userEvent.type(nodeEditor, nodeCode);

    const saveNodeButton = screen.getByRole('button', { name: /save/i });
    await userEvent.click(saveNodeButton);

    console.log('✅ Multiple runtime support tested successfully!');
  });

  it('should integrate functions with event sources', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Event Integration Org');
    await createWorkspace('Event Driven', 'shared');
    await createProject('events-project');

    // Step 1: Create event-driven function
    console.log('Step 1: Creating event-driven function...');
    const functionName = generateFunctionName();
    await deployFunction(functionName, 'nodejs18', `
exports.handler = async (event) => {
  console.log('Event received:', event);
  
  const { eventType, data } = event;
  
  switch (eventType) {
    case 'user.created':
      // Send welcome email
      console.log('Sending welcome email to:', data.email);
      break;
    case 'order.placed':
      // Process order
      console.log('Processing order:', data.orderId);
      break;
    case 'file.uploaded':
      // Process file
      console.log('Processing file:', data.fileName);
      break;
  }
  
  return {
    statusCode: 200,
    body: { processed: true, eventType }
  };
};`);

    // Step 2: Configure event triggers
    console.log('Step 2: Configuring event triggers...');
    const functionsTab = screen.getByRole('tab', { name: /functions/i });
    await userEvent.click(functionsTab);

    const functionRow = screen.getByText(functionName).closest('tr');
    const triggersButton = within(functionRow!).getByRole('button', { name: /triggers/i });
    await userEvent.click(triggersButton);

    const triggersModal = screen.getByRole('dialog');

    // Add S3 event trigger
    const addTriggerButton = within(triggersModal).getByRole('button', { name: /add trigger/i });
    await userEvent.click(addTriggerButton);

    const triggerTypeSelect = within(triggersModal).getByLabelText(/trigger type/i);
    await userEvent.selectOptions(triggerTypeSelect, 's3');

    const bucketInput = within(triggersModal).getByLabelText(/bucket name/i);
    await userEvent.type(bucketInput, 'uploads');

    const eventSelect = within(triggersModal).getByLabelText(/s3 event/i);
    await userEvent.selectOptions(eventSelect, 'ObjectCreated:*');

    const prefixInput = within(triggersModal).getByLabelText(/prefix filter/i);
    await userEvent.type(prefixInput, 'images/');

    const saveTriggerButton = within(triggersModal).getByRole('button', { name: /save trigger/i });
    await userEvent.click(saveTriggerButton);

    // Add message queue trigger
    await userEvent.click(addTriggerButton);

    const mqTypeSelect = within(triggersModal).getAllByLabelText(/trigger type/i)[1];
    await userEvent.selectOptions(mqTypeSelect, 'sqs');

    const queueInput = within(triggersModal).getByLabelText(/queue name/i);
    await userEvent.type(queueInput, 'order-events');

    const batchSizeInput = within(triggersModal).getByLabelText(/batch size/i);
    await userEvent.clear(batchSizeInput);
    await userEvent.type(batchSizeInput, '10');

    const saveMqButton = within(triggersModal).getAllByRole('button', { name: /save trigger/i })[1];
    await userEvent.click(saveMqButton);

    const closeTriggersButton = within(triggersModal).getByRole('button', { name: /close/i });
    await userEvent.click(closeTriggersButton);

    // Step 3: Test event processing
    console.log('Step 3: Testing event processing...');
    const testEventButton = screen.getByRole('button', { name: /test events/i });
    await userEvent.click(testEventButton);

    const eventModal = screen.getByRole('dialog');
    
    // Test S3 event
    const eventSourceSelect = within(eventModal).getByLabelText(/event source/i);
    await userEvent.selectOptions(eventSourceSelect, 's3');

    const s3EventPayload = JSON.stringify({
      Records: [{
        s3: {
          bucket: { name: 'uploads' },
          object: { key: 'images/photo.jpg' }
        }
      }]
    }, null, 2);

    const eventPayloadEditor = within(eventModal).getByLabelText(/event payload/i);
    await userEvent.type(eventPayloadEditor, s3EventPayload);

    const sendEventButton = within(eventModal).getByRole('button', { name: /send event/i });
    await userEvent.click(sendEventButton);

    // Verify event processing
    await waitFor(() => {
      const eventResult = screen.getByTestId('event-result');
      expect(eventResult).toHaveTextContent(/processed.*true/i);
    });

    // Check event logs
    const eventLogsTab = within(eventModal).getByRole('tab', { name: /event logs/i });
    await userEvent.click(eventLogsTab);

    expect(screen.getByText(/processing file.*photo\.jpg/i)).toBeInTheDocument();

    // Step 4: Monitor event metrics
    console.log('Step 4: Checking event metrics...');
    const metricsTab = within(eventModal).getByRole('tab', { name: /metrics/i });
    await userEvent.click(metricsTab);

    expect(screen.getByText(/events processed.*1/i)).toBeInTheDocument();
    expect(screen.getByText(/success rate.*100%/i)).toBeInTheDocument();
    expect(screen.getByText(/average latency/i)).toBeInTheDocument();

    console.log('✅ Function event integration tested successfully!');
  });

  it('should support function composition and chaining', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Function Chain Org');
    await createWorkspace('Workflow Testing', 'shared');
    await createProject('workflow-project');

    // Step 1: Create multiple functions for chaining
    console.log('Step 1: Creating functions for workflow...');
    
    // Function 1: Data validator
    await deployFunction('data-validator', 'nodejs18', `
exports.handler = async (event) => {
  const { data } = event;
  
  if (!data || !data.email || !data.name) {
    return {
      statusCode: 400,
      body: { error: 'Invalid data', valid: false }
    };
  }
  
  const emailRegex = /^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$/;
  if (!emailRegex.test(data.email)) {
    return {
      statusCode: 400,
      body: { error: 'Invalid email', valid: false }
    };
  }
  
  return {
    statusCode: 200,
    body: { valid: true, data }
  };
};`);

    // Function 2: Data enricher
    await deployFunction('data-enricher', 'nodejs18', `
exports.handler = async (event) => {
  const { data } = event;
  
  // Enrich data
  const enriched = {
    ...data,
    timestamp: new Date().toISOString(),
    id: 'user_' + Date.now(),
    preferences: {
      newsletter: true,
      notifications: true
    }
  };
  
  return {
    statusCode: 200,
    body: { enriched }
  };
};`);

    // Function 3: Data persister
    await deployFunction('data-persister', 'nodejs18', `
exports.handler = async (event) => {
  const { enriched } = event;
  
  // Simulate database save
  console.log('Saving to database:', enriched);
  
  return {
    statusCode: 200,
    body: {
      saved: true,
      id: enriched.id,
      message: 'User created successfully'
    }
  };
};`);

    // Step 2: Create workflow
    console.log('Step 2: Creating function workflow...');
    const workflowsTab = screen.getByRole('tab', { name: /workflows/i });
    await userEvent.click(workflowsTab);

    const createWorkflowButton = screen.getByRole('button', { name: /create workflow/i });
    await userEvent.click(createWorkflowButton);

    const workflowModal = screen.getByRole('dialog');
    
    const workflowNameInput = within(workflowModal).getByLabelText(/workflow name/i);
    await userEvent.type(workflowNameInput, 'user-registration');

    // Add steps
    const addStepButton = within(workflowModal).getByRole('button', { name: /add step/i });
    
    // Step 1: Validator
    await userEvent.click(addStepButton);
    const step1Select = within(workflowModal).getByLabelText(/function for step 1/i);
    await userEvent.selectOptions(step1Select, 'data-validator');

    // Step 2: Enricher
    await userEvent.click(addStepButton);
    const step2Select = within(workflowModal).getByLabelText(/function for step 2/i);
    await userEvent.selectOptions(step2Select, 'data-enricher');

    // Step 3: Persister
    await userEvent.click(addStepButton);
    const step3Select = within(workflowModal).getByLabelText(/function for step 3/i);
    await userEvent.selectOptions(step3Select, 'data-persister');

    // Configure error handling
    const errorHandlingTab = within(workflowModal).getByRole('tab', { name: /error handling/i });
    await userEvent.click(errorHandlingTab);

    const retryCheckbox = within(workflowModal).getByLabelText(/enable retry/i);
    await userEvent.click(retryCheckbox);

    const maxRetriesInput = within(workflowModal).getByLabelText(/max retries/i);
    await userEvent.type(maxRetriesInput, '3');

    const saveWorkflowButton = within(workflowModal).getByRole('button', { name: /save workflow/i });
    await userEvent.click(saveWorkflowButton);

    await expectNotification(/workflow created/i);

    // Step 3: Test the workflow
    console.log('Step 3: Testing workflow execution...');
    const workflowRow = screen.getByText('user-registration').closest('tr');
    const runButton = within(workflowRow!).getByRole('button', { name: /run/i });
    await userEvent.click(runButton);

    const runModal = screen.getByRole('dialog');
    
    const inputPayload = JSON.stringify({
      data: {
        name: 'John Doe',
        email: 'john@example.com'
      }
    }, null, 2);

    const inputEditor = within(runModal).getByLabelText(/input payload/i);
    await userEvent.type(inputEditor, inputPayload);

    const executeButton = within(runModal).getByRole('button', { name: /execute/i });
    await userEvent.click(executeButton);

    // Monitor workflow execution
    await waitFor(() => {
      const step1Status = screen.getByTestId('step-1-status');
      expect(step1Status).toHaveTextContent(/completed/i);
    });

    await waitFor(() => {
      const step2Status = screen.getByTestId('step-2-status');
      expect(step2Status).toHaveTextContent(/completed/i);
    });

    await waitFor(() => {
      const step3Status = screen.getByTestId('step-3-status');
      expect(step3Status).toHaveTextContent(/completed/i);
    });

    // Check final result
    const workflowResult = screen.getByTestId('workflow-result');
    expect(workflowResult).toHaveTextContent(/user created successfully/i);
    expect(workflowResult).toHaveTextContent(/user_\d+/);

    console.log('✅ Function composition and workflow tested successfully!');
  });
});