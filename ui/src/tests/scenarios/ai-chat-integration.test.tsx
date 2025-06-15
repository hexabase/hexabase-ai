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
  deployApplication,
  openAIChat,
  sendAIChatMessage,
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

describe('AI Chat Integration Scenarios', () => {
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
    
    // Setup AI chat responses
    mockApiClient.aiops.chat.mockImplementation(async (orgId, workspaceId, data) => {
      const message = data.message.toLowerCase();
      
      if (message.includes('optimize') && message.includes('nginx')) {
        return {
          message: `I'll help you optimize your nginx configuration. Based on your current setup, here are my recommendations:

1. **Enable Gzip Compression**: This will reduce bandwidth usage by 60-80%
2. **Configure Caching Headers**: Set appropriate cache-control headers for static assets
3. **Tune Worker Processes**: Set worker_processes to match your CPU cores (currently 4)
4. **Enable HTTP/2**: Modern protocol for better performance

Would you like me to apply these optimizations to your nginx configuration?`,
          suggestions: [
            'Enable gzip compression for text/html, text/css, application/javascript',
            'Set cache-control: max-age=31536000 for static assets',
            'Configure worker_processes auto',
            'Enable http2 in listen directive',
          ],
          actions: [
            {
              type: 'edit_config',
              label: 'Apply Optimizations',
              resource: 'app-nginx-1',
              changes: {
                'nginx.conf': {
                  gzip: 'on',
                  gzip_types: 'text/plain text/css application/json application/javascript',
                  worker_processes: 'auto',
                  http2: true,
                }
              }
            }
          ],
          context: {
            current_performance: {
              response_time_p95: 250,
              throughput: 1000,
              error_rate: 0.01,
            },
            expected_improvement: {
              response_time_p95: 150,
              throughput: 1500,
              error_rate: 0.01,
            }
          }
        };
      }
      
      if (message.includes('scale') && message.includes('database')) {
        return {
          message: `I notice your PostgreSQL database is experiencing high load. Here's my scaling recommendation:

**Current State:**
- CPU Usage: 85%
- Memory Usage: 92%
- Active Connections: 95/100
- Query Performance: Degrading

**Recommended Actions:**
1. **Immediate**: Increase connection pool to 200
2. **Short-term**: Scale to larger instance (4 CPU → 8 CPU, 8GB → 16GB RAM)
3. **Long-term**: Implement read replicas for read-heavy queries

I can help you implement these changes. Which would you like to start with?`,
          suggestions: [
            'Increase max_connections to 200',
            'Scale instance to 8 CPU / 16GB RAM',
            'Add read replica for analytics queries',
            'Enable connection pooling with pgBouncer',
          ],
          actions: [
            {
              type: 'scale_resource',
              label: 'Scale Database Now',
              resource: 'app-postgres-1',
              changes: {
                replicas: 1,
                resources: {
                  cpu: '8',
                  memory: '16Gi',
                }
              }
            },
            {
              type: 'create_replica',
              label: 'Add Read Replica',
              resource: 'app-postgres-1',
            }
          ]
        };
      }
      
      if (message.includes('debug') || message.includes('error')) {
        return {
          message: `I've analyzed your application logs and found the issue:

**Error Pattern Detected:**
\`\`\`
ConnectionRefusedError: [Errno 111] Connection refused
Location: api-service → redis:6379
Frequency: 127 occurrences in last hour
\`\`\`

**Root Cause:**
Your Redis instance crashed due to memory exhaustion. The maxmemory policy is set to 'noeviction' causing the process to terminate.

**Solution:**
1. Restart Redis with increased memory limit
2. Set maxmemory-policy to 'allkeys-lru'
3. Implement connection retry logic

I can fix this for you. Shall I proceed?`,
          suggestions: [
            'Increase Redis memory limit from 1GB to 2GB',
            'Change eviction policy to allkeys-lru',
            'Add exponential backoff for Redis connections',
            'Set up Redis Sentinel for high availability',
          ],
          actions: [
            {
              type: 'restart_service',
              label: 'Restart Redis',
              resource: 'app-redis-1',
            },
            {
              type: 'update_config',
              label: 'Fix Redis Config',
              resource: 'app-redis-1',
              changes: {
                maxmemory: '2gb',
                'maxmemory-policy': 'allkeys-lru',
              }
            }
          ],
          context: {
            error_analysis: {
              service: 'api-service',
              dependency: 'redis',
              error_count: 127,
              first_occurrence: '2024-01-01T10:00:00Z',
              impact: 'API endpoints returning 503',
            }
          }
        };
      }
      
      if (message.includes('create') && message.includes('function')) {
        return {
          message: `I'll help you create a serverless function. Based on your request, here's what I'll generate:

**Function: Image Thumbnail Generator**
- Runtime: Node.js 18
- Memory: 512MB
- Timeout: 30 seconds
- Triggers: HTTP endpoint + S3 events

The function will:
1. Accept image uploads via HTTP or S3 events
2. Generate thumbnails in multiple sizes (150x150, 300x300, 600x600)
3. Store processed images back to S3
4. Return URLs for all generated thumbnails

Would you like me to create this function with the code?`,
          code_snippet: `const sharp = require('sharp');
const AWS = require('aws-sdk');
const s3 = new AWS.S3();

exports.handler = async (event) => {
  const sizes = [150, 300, 600];
  const results = [];
  
  try {
    const imageBuffer = await getImageBuffer(event);
    
    for (const size of sizes) {
      const thumbnail = await sharp(imageBuffer)
        .resize(size, size, { fit: 'cover' })
        .jpeg({ quality: 85 })
        .toBuffer();
      
      const key = \`thumbnails/\${size}x\${size}/\${Date.now()}.jpg\`;
      await s3.putObject({
        Bucket: process.env.BUCKET_NAME,
        Key: key,
        Body: thumbnail,
        ContentType: 'image/jpeg'
      }).promise();
      
      results.push({
        size: \`\${size}x\${size}\`,
        url: \`https://\${process.env.BUCKET_NAME}.s3.amazonaws.com/\${key}\`
      });
    }
    
    return {
      statusCode: 200,
      body: JSON.stringify({ thumbnails: results })
    };
  } catch (error) {
    console.error('Error:', error);
    return {
      statusCode: 500,
      body: JSON.stringify({ error: 'Failed to process image' })
    };
  }
};

async function getImageBuffer(event) {
  // Handle both HTTP and S3 events
  if (event.body) {
    return Buffer.from(event.body, 'base64');
  } else if (event.Records?.[0]?.s3) {
    const record = event.Records[0].s3;
    const obj = await s3.getObject({
      Bucket: record.bucket.name,
      Key: record.object.key
    }).promise();
    return obj.Body;
  }
  throw new Error('No image data found');
}`,
          actions: [
            {
              type: 'create_function',
              label: 'Create Function',
              config: {
                name: 'image-thumbnail-generator',
                runtime: 'nodejs18',
                memory: 512,
                timeout: 30,
                handler: 'index.handler',
              }
            }
          ]
        };
      }
      
      // Default response
      return {
        message: "I'm here to help! I can assist with optimizing resources, debugging issues, scaling applications, and writing code. What would you like help with?",
        suggestions: [
          'Analyze resource usage and suggest optimizations',
          'Debug application errors',
          'Help with scaling decisions',
          'Generate code for common tasks',
        ]
      };
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

  it('should help optimize application performance through AI chat', async () => {
    renderApp();

    // Initial setup
    await loginUser();
    await createOrganization('AI Optimization Org');
    await createWorkspace('Production', 'dedicated');
    const projectName = generateProjectName();
    await createProject(projectName);

    // Deploy nginx application
    const nginxApp = generateAppName();
    await deployApplication(nginxApp, 'stateless', {
      image: 'nginx:latest',
      replicas: 3,
      port: 80,
    });

    // Step 1: Open AI Chat
    console.log('Step 1: Opening AI Chat...');
    await openAIChat();

    // Step 2: Ask for optimization suggestions
    console.log('Step 2: Asking for nginx optimization...');
    const response1 = await sendAIChatMessage(
      `Can you help me optimize my nginx configuration for better performance? The app name is ${nginxApp}`
    );

    // Verify AI response
    expect(response1).toContain('optimize your nginx configuration');
    expect(response1).toContain('Enable Gzip Compression');
    expect(response1).toContain('HTTP/2');

    // Step 3: Accept AI suggestions
    console.log('Step 3: Applying AI recommendations...');
    const applyButton = screen.getByRole('button', { name: /apply optimizations/i });
    await userEvent.click(applyButton);

    // Confirm changes
    const confirmModal = screen.getByRole('dialog');
    expect(within(confirmModal).getByText(/review changes/i)).toBeInTheDocument();
    expect(within(confirmModal).getByText(/gzip.*on/i)).toBeInTheDocument();
    expect(within(confirmModal).getByText(/worker_processes.*auto/i)).toBeInTheDocument();

    const confirmButton = within(confirmModal).getByRole('button', { name: /apply/i });
    await userEvent.click(confirmButton);

    await expectNotification(/configuration updated successfully/i);

    // Step 4: Verify improvements
    console.log('Step 4: Checking performance improvements...');
    await delay(2000); // Wait for metrics to update

    const metricsButton = screen.getByRole('button', { name: /view metrics/i });
    await userEvent.click(metricsButton);

    // Check improved metrics
    await waitFor(() => {
      const responseTime = screen.getByTestId('response-time-p95');
      expect(responseTime).toHaveTextContent(/150ms/); // Improved from 250ms
    });

    const throughput = screen.getByTestId('throughput');
    expect(throughput).toHaveTextContent(/1500/); // Improved from 1000

    console.log('✅ AI-assisted optimization completed successfully!');
  });

  it('should help debug and fix application errors', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Debug Test Org');
    await createWorkspace('Debugging Workspace', 'shared');
    await createProject('debug-project');

    // Deploy applications with issues
    const apiApp = generateAppName();
    await deployApplication(apiApp, 'stateless', {
      image: 'api-service:latest',
      replicas: 2,
      port: 3000,
    });

    const redisApp = generateAppName();
    await deployApplication(redisApp, 'stateful', {
      image: 'redis:7',
      port: 6379,
      storage: '1Gi',
    });

    // Simulate Redis crash
    mockApiClient.applications.updateStatus.mockResolvedValueOnce({
      data: { ...mockData.applications[0], status: 'error' }
    });

    // Step 1: Notice error state
    console.log('Step 1: Detecting application error...');
    const errorAlert = screen.getByRole('alert');
    expect(errorAlert).toHaveTextContent(/redis.*error/i);

    // Step 2: Ask AI for help
    console.log('Step 2: Asking AI to debug...');
    await openAIChat();
    
    const debugResponse = await sendAIChatMessage(
      'My API is returning 503 errors. Can you help me debug what\'s wrong?'
    );

    // Verify AI identified the issue
    expect(debugResponse).toContain('ConnectionRefusedError');
    expect(debugResponse).toContain('Redis instance crashed');
    expect(debugResponse).toContain('memory exhaustion');

    // Step 3: Apply AI fix
    console.log('Step 3: Applying AI-suggested fix...');
    const fixButton = screen.getByRole('button', { name: /fix redis config/i });
    await userEvent.click(fixButton);

    await expectNotification(/applying fix/i);

    // Also restart Redis
    const restartButton = screen.getByRole('button', { name: /restart redis/i });
    await userEvent.click(restartButton);

    // Wait for restart
    await waitFor(() => {
      expect(screen.getByText(/redis restarting/i)).toBeInTheDocument();
    });

    await delay(3000);

    await waitFor(() => {
      const redisStatus = screen.getByTestId(`${redisApp}-status`);
      expect(redisStatus).toHaveTextContent(/running/i);
    });

    // Step 4: Verify resolution
    console.log('Step 4: Verifying issue resolution...');
    const response2 = await sendAIChatMessage('Is the issue resolved now?');
    
    expect(mockApiClient.aiops.chat).toHaveBeenLastCalledWith(
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        message: 'Is the issue resolved now?'
      })
    );

    // Check error alert is gone
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();

    console.log('✅ AI-assisted debugging completed successfully!');
  });

  it('should generate code for serverless functions', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Code Generation Org');
    await createWorkspace('Development', 'shared');
    await createProject('codegen-project');

    // Step 1: Open AI Chat
    console.log('Step 1: Opening AI Chat for code generation...');
    await openAIChat();

    // Step 2: Request function generation
    console.log('Step 2: Requesting function code...');
    const codeResponse = await sendAIChatMessage(
      'Create a serverless function that generates thumbnails from uploaded images'
    );

    // Verify AI response includes code
    expect(codeResponse).toContain('Image Thumbnail Generator');
    expect(codeResponse).toContain('sharp');
    expect(codeResponse).toContain('exports.handler');

    // Step 3: Review generated code
    console.log('Step 3: Reviewing generated code...');
    const codePreview = screen.getByTestId('code-preview');
    expect(codePreview).toBeInTheDocument();
    expect(codePreview).toHaveTextContent(/const sharp = require/);
    expect(codePreview).toHaveTextContent(/resize.*fit.*cover/);

    // Step 4: Create function from AI code
    console.log('Step 4: Creating function from AI code...');
    const createFunctionButton = screen.getByRole('button', { name: /create function/i });
    await userEvent.click(createFunctionButton);

    // Verify function creation modal is pre-filled
    const modal = screen.getByRole('dialog');
    const nameInput = within(modal).getByLabelText(/function name/i) as HTMLInputElement;
    expect(nameInput.value).toBe('image-thumbnail-generator');

    const runtimeSelect = within(modal).getByLabelText(/runtime/i) as HTMLSelectElement;
    expect(runtimeSelect.value).toBe('nodejs18');

    const memorySelect = within(modal).getByLabelText(/memory/i) as HTMLSelectElement;
    expect(memorySelect.value).toBe('512');

    // Code should be pre-populated
    const codeEditor = within(modal).getByTestId('code-editor') as HTMLTextAreaElement;
    expect(codeEditor.value).toContain('const sharp = require');

    // Confirm creation
    const confirmCreateButton = within(modal).getByRole('button', { name: /create/i });
    await userEvent.click(confirmCreateButton);

    await expectNotification(/function created/i);

    // Step 5: Test the generated function
    console.log('Step 5: Testing generated function...');
    const functionsTab = screen.getByRole('tab', { name: /functions/i });
    await userEvent.click(functionsTab);

    const functionRow = screen.getByText('image-thumbnail-generator').closest('tr');
    const testButton = within(functionRow!).getByRole('button', { name: /test/i });
    await userEvent.click(testButton);

    // Provide test input
    const testModal = screen.getByRole('dialog');
    const payloadInput = within(testModal).getByLabelText(/test payload/i);
    
    const testPayload = JSON.stringify({
      body: 'base64ImageDataHere...',
      isBase64Encoded: true
    }, null, 2);
    
    await userEvent.clear(payloadInput);
    await userEvent.type(payloadInput, testPayload);

    const runTestButton = within(testModal).getByRole('button', { name: /run test/i });
    await userEvent.click(runTestButton);

    // Verify successful execution
    await waitFor(() => {
      const result = screen.getByTestId('test-result');
      expect(result).toHaveTextContent(/thumbnails/);
      expect(result).toHaveTextContent(/150x150/);
    });

    console.log('✅ AI code generation completed successfully!');
  });

  it('should provide intelligent scaling recommendations', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Scaling Test Org');
    await createWorkspace('Production', 'dedicated');
    await createProject('scaling-project');

    // Deploy database under load
    const dbApp = generateAppName();
    await deployApplication(dbApp, 'stateful', {
      image: 'postgres:14',
      port: 5432,
      storage: '20Gi',
    });

    // Simulate high load metrics
    mockApiClient.monitoring.getApplicationMetrics.mockResolvedValueOnce({
      metrics: [
        {
          timestamp: new Date().toISOString(),
          cpu_usage: 85,
          memory_usage: 92,
          disk_iops: 5000,
          connections: 95,
          queries_per_second: 1200,
        }
      ]
    });

    // Step 1: Ask AI about scaling
    console.log('Step 1: Asking AI about database scaling...');
    await openAIChat();
    
    const scalingResponse = await sendAIChatMessage(
      `My PostgreSQL database ${dbApp} seems to be under heavy load. Should I scale it?`
    );

    // Verify AI analysis
    expect(scalingResponse).toContain('high load');
    expect(scalingResponse).toContain('CPU Usage: 85%');
    expect(scalingResponse).toContain('Memory Usage: 92%');
    expect(scalingResponse).toContain('Scale to larger instance');

    // Step 2: Review scaling options
    console.log('Step 2: Reviewing scaling recommendations...');
    const scalingOptions = screen.getAllByTestId('scaling-option');
    expect(scalingOptions).toHaveLength(2); // Immediate scale + read replica

    // Check recommended specs
    const scaleOption = scalingOptions[0];
    expect(within(scaleOption).getByText(/8 cpu/i)).toBeInTheDocument();
    expect(within(scaleOption).getByText(/16gb ram/i)).toBeInTheDocument();

    // Step 3: Apply scaling
    console.log('Step 3: Applying scaling recommendation...');
    const scaleNowButton = within(scaleOption).getByRole('button', { name: /scale database now/i });
    await userEvent.click(scaleNowButton);

    // Confirm scaling
    const scaleModal = screen.getByRole('dialog');
    expect(within(scaleModal).getByText(/scale database/i)).toBeInTheDocument();
    expect(within(scaleModal).getByText(/4 cpu.*8 cpu/i)).toBeInTheDocument();
    expect(within(scaleModal).getByText(/8gb.*16gb/i)).toBeInTheDocument();

    const confirmScaleButton = within(scaleModal).getByRole('button', { name: /confirm/i });
    await userEvent.click(confirmScaleButton);

    await expectNotification(/scaling initiated/i);

    // Monitor scaling progress
    await waitFor(() => {
      const scalingStatus = screen.getByTestId('scaling-status');
      expect(scalingStatus).toHaveTextContent(/scaling in progress/i);
    });

    // Step 4: Add read replica as suggested
    console.log('Step 4: Adding read replica...');
    const replicaButton = screen.getByRole('button', { name: /add read replica/i });
    await userEvent.click(replicaButton);

    const replicaModal = screen.getByRole('dialog');
    const replicaNameInput = within(replicaModal).getByLabelText(/replica name/i);
    await userEvent.type(replicaNameInput, `${dbApp}-read-1`);

    const createReplicaButton = within(replicaModal).getByRole('button', { name: /create/i });
    await userEvent.click(createReplicaButton);

    await expectNotification(/read replica created/i);

    // Step 5: Verify improvements
    console.log('Step 5: Checking if scaling helped...');
    
    // Update metrics to show improvement
    mockApiClient.monitoring.getApplicationMetrics.mockResolvedValueOnce({
      metrics: [
        {
          timestamp: new Date().toISOString(),
          cpu_usage: 45,
          memory_usage: 55,
          disk_iops: 3000,
          connections: 50,
          queries_per_second: 1200,
        }
      ]
    });

    await delay(5000); // Wait for scaling to complete

    const checkResponse = await sendAIChatMessage('How is the database performing now?');
    
    expect(mockApiClient.aiops.chat).toHaveBeenCalled();

    console.log('✅ AI-assisted scaling completed successfully!');
  });

  it('should provide contextual help based on current view', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Context Test Org');
    await createWorkspace('Testing', 'shared');
    await createProject('context-project');

    // Step 1: Navigate to different sections and get contextual help
    console.log('Step 1: Testing contextual help in applications view...');
    
    // Deploy some applications first
    await deployApplication('web-app', 'stateless', {
      image: 'node:18',
      replicas: 2,
      port: 3000,
    });

    await deployApplication('worker', 'stateless', {
      image: 'worker:latest',
      replicas: 1,
    });

    // Open AI chat in applications context
    await openAIChat();
    
    const appContextResponse = await sendAIChatMessage('What can you tell me about my applications?');
    
    // AI should be aware of current context
    expect(mockApiClient.aiops.chat).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        context: expect.objectContaining({
          current_view: 'applications',
          visible_resources: expect.arrayContaining([
            expect.objectContaining({ name: 'web-app' }),
            expect.objectContaining({ name: 'worker' })
          ])
        })
      })
    );

    // Step 2: Navigate to monitoring and ask for help
    console.log('Step 2: Testing contextual help in monitoring view...');
    const monitoringTab = screen.getByRole('tab', { name: /monitoring/i });
    await userEvent.click(monitoringTab);

    const monitorResponse = await sendAIChatMessage('What metrics should I pay attention to?');
    
    // AI should provide monitoring-specific advice
    expect(mockApiClient.aiops.chat).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        context: expect.objectContaining({
          current_view: 'monitoring'
        })
      })
    );

    // Step 3: Test quick actions
    console.log('Step 3: Testing AI quick actions...');
    const quickActionsButton = screen.getByRole('button', { name: /ai actions/i });
    await userEvent.click(quickActionsButton);

    const actionsMenu = screen.getByRole('menu');
    expect(within(actionsMenu).getByText(/analyze performance/i)).toBeInTheDocument();
    expect(within(actionsMenu).getByText(/check for issues/i)).toBeInTheDocument();
    expect(within(actionsMenu).getByText(/suggest optimizations/i)).toBeInTheDocument();

    // Select analyze performance
    const analyzeButton = within(actionsMenu).getByText(/analyze performance/i);
    await userEvent.click(analyzeButton);

    // AI should automatically analyze without needing a prompt
    await waitFor(() => {
      const analysisResult = screen.getByTestId('ai-analysis');
      expect(analysisResult).toHaveTextContent(/performance analysis/i);
    });

    console.log('✅ Contextual AI help tested successfully!');
  });

  it('should learn from user interactions and improve suggestions', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Learning Test Org');
    await createWorkspace('ML Workspace', 'dedicated');
    await createProject('learning-project');

    // Step 1: Make several similar requests
    console.log('Step 1: Training AI with user preferences...');
    await openAIChat();

    // First deployment
    await sendAIChatMessage('Deploy a PostgreSQL database for my application');
    
    const deployButton1 = screen.getByRole('button', { name: /deploy postgres/i });
    await userEvent.click(deployButton1);

    // User modifies the suggested configuration
    const configModal1 = screen.getByRole('dialog');
    const memoryInput1 = within(configModal1).getByLabelText(/memory/i);
    await userEvent.clear(memoryInput1);
    await userEvent.type(memoryInput1, '4Gi'); // User prefers 4Gi instead of default

    const storageInput1 = within(configModal1).getByLabelText(/storage/i);
    await userEvent.clear(storageInput1);
    await userEvent.type(storageInput1, '100Gi'); // User prefers 100Gi

    const confirmButton1 = within(configModal1).getByRole('button', { name: /deploy/i });
    await userEvent.click(confirmButton1);

    // Second deployment - AI should learn preferences
    await sendAIChatMessage('I need another PostgreSQL database');

    // Step 2: Verify AI learned preferences
    console.log('Step 2: Verifying AI learned user preferences...');
    
    // Check that AI suggests the user's preferred settings
    const suggestion = screen.getByTestId('ai-suggestion');
    expect(suggestion).toHaveTextContent(/4gi memory/i); // AI learned preference
    expect(suggestion).toHaveTextContent(/100gi storage/i); // AI learned preference

    // Step 3: Test pattern recognition
    console.log('Step 3: Testing pattern recognition...');
    
    // Deploy several microservices with similar patterns
    await sendAIChatMessage('Deploy an API service with Node.js');
    await delay(1000);
    
    await sendAIChatMessage('Deploy another API service with Node.js');
    await delay(1000);

    // AI should recognize the pattern
    await sendAIChatMessage('I need one more service');
    
    const patternResponse = await mockApiClient.aiops.chat.mock.results.slice(-1)[0].value;
    expect(patternResponse.message).toContain('Node.js API service'); // AI inferred the pattern

    // Step 4: Test feedback learning
    console.log('Step 4: Testing feedback-based learning...');
    
    // Give feedback on suggestions
    const feedbackButton = screen.getByRole('button', { name: /👎/i }); // Thumbs down
    await userEvent.click(feedbackButton);

    const feedbackModal = screen.getByRole('dialog');
    const feedbackInput = within(feedbackModal).getByLabelText(/what could be better/i);
    await userEvent.type(feedbackInput, 'I prefer Python for API services, not Node.js');

    const submitFeedbackButton = within(feedbackModal).getByRole('button', { name: /submit/i });
    await userEvent.click(submitFeedbackButton);

    // Next suggestion should adapt
    await sendAIChatMessage('Create another API service');
    
    const adaptedResponse = await mockApiClient.aiops.chat.mock.results.slice(-1)[0].value;
    expect(adaptedResponse.message).toContain('Python'); // AI adapted based on feedback

    console.log('✅ AI learning and adaptation tested successfully!');
  });
});