import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { mockApiClient } from '@/test-utils/mock-api-client';

/**
 * Common test helpers for scenario tests
 */

// Delay utility for realistic user interactions
export const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

// Wait for element with retry
export async function waitForElement(
  selector: () => HTMLElement | null,
  options = { timeout: 5000 }
) {
  const startTime = Date.now();
  while (Date.now() - startTime < options.timeout) {
    const element = selector();
    if (element) return element;
    await delay(100);
  }
  throw new Error('Element not found within timeout');
}

// Login helper
export async function loginUser(email = 'test@example.com', password = 'password123') {
  // Fill login form
  const emailInput = screen.getByLabelText(/email/i);
  const passwordInput = screen.getByLabelText(/password/i);
  
  await userEvent.type(emailInput, email);
  await userEvent.type(passwordInput, password);
  
  const loginButton = screen.getByRole('button', { name: /sign in/i });
  await userEvent.click(loginButton);
  
  // Wait for redirect to dashboard
  await waitFor(() => {
    expect(screen.getByTestId('dashboard')).toBeInTheDocument();
  });
}

// Create organization helper
export async function createOrganization(name: string) {
  const createOrgButton = screen.getByRole('button', { name: /create organization/i });
  await userEvent.click(createOrgButton);
  
  const modal = screen.getByRole('dialog');
  const nameInput = within(modal).getByLabelText(/organization name/i);
  await userEvent.type(nameInput, name);
  
  const createButton = within(modal).getByRole('button', { name: /create/i });
  await userEvent.click(createButton);
  
  await waitFor(() => {
    expect(screen.getByText(name)).toBeInTheDocument();
  });
}

// Create workspace helper
export async function createWorkspace(name: string, plan: 'shared' | 'dedicated' = 'dedicated') {
  const createWsButton = screen.getByRole('button', { name: /create workspace/i });
  await userEvent.click(createWsButton);
  
  const modal = screen.getByRole('dialog');
  const nameInput = within(modal).getByLabelText(/workspace name/i);
  await userEvent.type(nameInput, name);
  
  // Select plan
  const planRadio = within(modal).getByLabelText(new RegExp(plan, 'i'));
  await userEvent.click(planRadio);
  
  if (plan === 'dedicated') {
    // Select node pool
    const nodeSelect = within(modal).getByLabelText(/node pool/i);
    await userEvent.selectOptions(nodeSelect, 'dedicated-pool-1');
  }
  
  const createButton = within(modal).getByRole('button', { name: /create/i });
  await userEvent.click(createButton);
  
  // Wait for workspace to be ready
  await waitFor(() => {
    expect(screen.getByText(name)).toBeInTheDocument();
    expect(screen.getByText(/active/i)).toBeInTheDocument();
  }, { timeout: 10000 });
}

// Create project helper
export async function createProject(name: string, quotas?: { cpu: string; memory: string; storage: string }) {
  const createProjButton = screen.getByRole('button', { name: /create project/i });
  await userEvent.click(createProjButton);
  
  const modal = screen.getByRole('dialog');
  const nameInput = within(modal).getByLabelText(/project name/i);
  await userEvent.type(nameInput, name);
  
  if (quotas) {
    const cpuInput = within(modal).getByLabelText(/cpu limit/i);
    const memoryInput = within(modal).getByLabelText(/memory limit/i);
    const storageInput = within(modal).getByLabelText(/storage limit/i);
    
    await userEvent.clear(cpuInput);
    await userEvent.type(cpuInput, quotas.cpu);
    
    await userEvent.clear(memoryInput);
    await userEvent.type(memoryInput, quotas.memory);
    
    await userEvent.clear(storageInput);
    await userEvent.type(storageInput, quotas.storage);
  }
  
  const createButton = within(modal).getByRole('button', { name: /create/i });
  await userEvent.click(createButton);
  
  await waitFor(() => {
    expect(screen.getByText(name)).toBeInTheDocument();
  });
}

// Deploy application helper
export async function deployApplication(
  name: string,
  type: 'stateless' | 'stateful' | 'cronjob',
  config: {
    image: string;
    replicas?: number;
    port?: number;
    storage?: string;
    schedule?: string;
  }
) {
  const deployButton = screen.getByRole('button', { name: /deploy application/i });
  await userEvent.click(deployButton);
  
  const modal = screen.getByRole('dialog');
  
  // Basic info
  const nameInput = within(modal).getByLabelText(/application name/i);
  await userEvent.type(nameInput, name);
  
  const typeSelect = within(modal).getByLabelText(/application type/i);
  await userEvent.selectOptions(typeSelect, type);
  
  const imageInput = within(modal).getByLabelText(/container image/i);
  await userEvent.type(imageInput, config.image);
  
  // Type-specific config
  if (type === 'stateless' && config.replicas) {
    const replicasInput = within(modal).getByLabelText(/replicas/i);
    await userEvent.clear(replicasInput);
    await userEvent.type(replicasInput, config.replicas.toString());
  }
  
  if (type === 'stateful' && config.storage) {
    const storageInput = within(modal).getByLabelText(/storage size/i);
    await userEvent.type(storageInput, config.storage);
  }
  
  if (type === 'cronjob' && config.schedule) {
    const scheduleInput = within(modal).getByLabelText(/schedule/i);
    await userEvent.type(scheduleInput, config.schedule);
  }
  
  if (config.port) {
    const portInput = within(modal).getByLabelText(/port/i);
    await userEvent.type(portInput, config.port.toString());
  }
  
  const deployButton = within(modal).getByRole('button', { name: /deploy/i });
  await userEvent.click(deployButton);
  
  // Wait for deployment to complete
  await waitFor(() => {
    expect(screen.getByText(name)).toBeInTheDocument();
    expect(screen.getByText(/running|active/i)).toBeInTheDocument();
  }, { timeout: 15000 });
}

// Trigger CronJob helper
export async function triggerCronJob(name: string) {
  const cronJobRow = screen.getByText(name).closest('tr');
  if (!cronJobRow) throw new Error(`CronJob ${name} not found`);
  
  const triggerButton = within(cronJobRow).getByRole('button', { name: /trigger/i });
  await userEvent.click(triggerButton);
  
  await waitFor(() => {
    expect(screen.getByText(/triggered successfully/i)).toBeInTheDocument();
  });
}

// Setup backup storage helper
export async function setupBackupStorage(name: string, type: 'proxmox' | 's3' = 'proxmox') {
  const backupTab = screen.getByRole('tab', { name: /backup/i });
  await userEvent.click(backupTab);
  
  const addStorageButton = screen.getByRole('button', { name: /add storage/i });
  await userEvent.click(addStorageButton);
  
  const modal = screen.getByRole('dialog');
  
  const nameInput = within(modal).getByLabelText(/storage name/i);
  await userEvent.type(nameInput, name);
  
  const typeSelect = within(modal).getByLabelText(/storage type/i);
  await userEvent.selectOptions(typeSelect, type);
  
  if (type === 'proxmox') {
    const hostInput = within(modal).getByLabelText(/proxmox host/i);
    await userEvent.type(hostInput, 'proxmox.local');
    
    const storageInput = within(modal).getByLabelText(/storage name/i);
    await userEvent.type(storageInput, 'backup-storage');
  }
  
  const createButton = within(modal).getByRole('button', { name: /create/i });
  await userEvent.click(createButton);
  
  await waitFor(() => {
    expect(screen.getByText(name)).toBeInTheDocument();
  });
}

// Create backup policy helper
export async function createBackupPolicy(
  appName: string,
  schedule: string,
  retention: number
) {
  const policyTab = screen.getByRole('tab', { name: /policies/i });
  await userEvent.click(policyTab);
  
  const createPolicyButton = screen.getByRole('button', { name: /create policy/i });
  await userEvent.click(createPolicyButton);
  
  const modal = screen.getByRole('dialog');
  
  const appSelect = within(modal).getByLabelText(/application/i);
  await userEvent.selectOptions(appSelect, appName);
  
  const scheduleInput = within(modal).getByLabelText(/schedule/i);
  await userEvent.type(scheduleInput, schedule);
  
  const retentionInput = within(modal).getByLabelText(/retention days/i);
  await userEvent.clear(retentionInput);
  await userEvent.type(retentionInput, retention.toString());
  
  const createButton = within(modal).getByRole('button', { name: /create/i });
  await userEvent.click(createButton);
  
  await waitFor(() => {
    expect(screen.getByText(appName)).toBeInTheDocument();
  });
}

// Deploy function helper
export async function deployFunction(
  name: string,
  runtime: string,
  code: string,
  handler = 'index.handler'
) {
  const functionsTab = screen.getByRole('tab', { name: /functions/i });
  await userEvent.click(functionsTab);
  
  const deployButton = screen.getByRole('button', { name: /deploy function/i });
  await userEvent.click(deployButton);
  
  const modal = screen.getByRole('dialog');
  
  const nameInput = within(modal).getByLabelText(/function name/i);
  await userEvent.type(nameInput, name);
  
  const runtimeSelect = within(modal).getByLabelText(/runtime/i);
  await userEvent.selectOptions(runtimeSelect, runtime);
  
  const handlerInput = within(modal).getByLabelText(/handler/i);
  await userEvent.clear(handlerInput);
  await userEvent.type(handlerInput, handler);
  
  const codeEditor = within(modal).getByLabelText(/code/i);
  await userEvent.type(codeEditor, code);
  
  const deployButton2 = within(modal).getByRole('button', { name: /deploy/i });
  await userEvent.click(deployButton2);
  
  await waitFor(() => {
    expect(screen.getByText(name)).toBeInTheDocument();
    expect(screen.getByText(/active/i)).toBeInTheDocument();
  }, { timeout: 10000 });
}

// Invoke function helper
export async function invokeFunction(name: string, payload: any) {
  const functionRow = screen.getByText(name).closest('tr');
  if (!functionRow) throw new Error(`Function ${name} not found`);
  
  const invokeButton = within(functionRow).getByRole('button', { name: /invoke/i });
  await userEvent.click(invokeButton);
  
  const modal = screen.getByRole('dialog');
  
  const payloadInput = within(modal).getByLabelText(/payload/i);
  await userEvent.type(payloadInput, JSON.stringify(payload));
  
  const invokeButton2 = within(modal).getByRole('button', { name: /invoke/i });
  await userEvent.click(invokeButton2);
  
  await waitFor(() => {
    expect(screen.getByText(/result/i)).toBeInTheDocument();
  });
  
  return screen.getByTestId('function-result').textContent;
}

// AI Chat helper
export async function openAIChat() {
  const aiChatButton = screen.getByRole('button', { name: /ai chat/i });
  await userEvent.click(aiChatButton);
  
  await waitFor(() => {
    expect(screen.getByTestId('ai-chat-panel')).toBeInTheDocument();
  });
}

// Send AI chat message helper
export async function sendAIChatMessage(message: string) {
  const chatInput = screen.getByPlaceholderText(/ask me anything/i);
  await userEvent.type(chatInput, message);
  
  const sendButton = screen.getByRole('button', { name: /send/i });
  await userEvent.click(sendButton);
  
  // Wait for response
  await waitFor(() => {
    const messages = screen.getAllByTestId('chat-message');
    expect(messages.length).toBeGreaterThan(1);
  }, { timeout: 10000 });
  
  const messages = screen.getAllByTestId('chat-message');
  return messages[messages.length - 1].textContent;
}

// Setup CI/CD pipeline helper
export async function setupCICD(
  repoUrl: string,
  branch: string,
  buildCommand: string
) {
  const settingsTab = screen.getByRole('tab', { name: /settings/i });
  await userEvent.click(settingsTab);
  
  const cicdTab = screen.getByRole('tab', { name: /ci\/cd/i });
  await userEvent.click(cicdTab);
  
  const connectButton = screen.getByRole('button', { name: /connect repository/i });
  await userEvent.click(connectButton);
  
  const modal = screen.getByRole('dialog');
  
  const repoInput = within(modal).getByLabelText(/repository url/i);
  await userEvent.type(repoInput, repoUrl);
  
  const branchInput = within(modal).getByLabelText(/branch/i);
  await userEvent.type(branchInput, branch);
  
  const buildInput = within(modal).getByLabelText(/build command/i);
  await userEvent.type(buildInput, buildCommand);
  
  const connectButton2 = within(modal).getByRole('button', { name: /connect/i });
  await userEvent.click(connectButton2);
  
  await waitFor(() => {
    expect(screen.getByText(/connected/i)).toBeInTheDocument();
  });
}

// Monitor deployment helper
export async function monitorDeployment(deploymentName: string) {
  const deploymentRow = screen.getByText(deploymentName).closest('tr');
  if (!deploymentRow) throw new Error(`Deployment ${deploymentName} not found`);
  
  // Wait for deployment to complete
  await waitFor(() => {
    const status = within(deploymentRow).getByTestId('deployment-status');
    expect(status).toHaveTextContent(/success|completed/i);
  }, { timeout: 30000 });
}

// Verify resource status helper
export async function verifyResourceStatus(resourceName: string, expectedStatus: string) {
  const resourceRow = screen.getByText(resourceName).closest('tr');
  if (!resourceRow) throw new Error(`Resource ${resourceName} not found`);
  
  const status = within(resourceRow).getByTestId('resource-status');
  expect(status).toHaveTextContent(new RegExp(expectedStatus, 'i'));
}

// Mock data generators
export const generateProjectName = () => `test-project-${Date.now()}`;
export const generateAppName = () => `test-app-${Date.now()}`;
export const generateFunctionName = () => `test-func-${Date.now()}`;

// Assertion helpers
export const expectNotification = async (message: string | RegExp) => {
  await waitFor(() => {
    const notification = screen.getByRole('alert');
    expect(notification).toHaveTextContent(message);
  });
};

export const expectNoErrors = () => {
  const errors = screen.queryAllByRole('alert', { name: /error/i });
  expect(errors).toHaveLength(0);
};