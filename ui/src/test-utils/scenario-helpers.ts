import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { apiClient } from '@/lib/api-client';

/**
 * Scenario test helpers for common user flows
 */

export interface TestUser {
  id: string;
  email: string;
  name: string;
  password: string;
  organizations: string[];
}

export interface TestOrganization {
  id: string;
  name: string;
  owner_id: string;
  workspaces: string[];
}

export interface TestWorkspace {
  id: string;
  organization_id: string;
  name: string;
  plan: 'shared' | 'dedicated';
  projects: string[];
}

// Default test data
export const defaultTestUser: TestUser = {
  id: 'user-test-123',
  email: 'test@hexabase.ai',
  name: 'Test User',
  password: 'TestPassword123!',
  organizations: ['org-test-1'],
};

export const defaultTestOrg: TestOrganization = {
  id: 'org-test-1',
  name: 'Test Organization',
  owner_id: 'user-test-123',
  workspaces: ['ws-test-1'],
};

export const defaultTestWorkspace: TestWorkspace = {
  id: 'ws-test-1',
  organization_id: 'org-test-1',
  name: 'Test Workspace',
  plan: 'dedicated',
  projects: ['proj-test-1'],
};

/**
 * Setup function to initialize test environment
 */
export function setupScenarioTest() {
  const user = userEvent.setup();
  
  // Reset all mocks
  jest.clearAllMocks();
  
  // Setup default mock responses
  setupAuthMocks();
  setupOrganizationMocks();
  setupWorkspaceMocks();
  setupProjectMocks();
  setupApplicationMocks();
  setupMonitoringMocks();
  setupBackupMocks();
  
  return { user };
}

/**
 * Mock authentication flow
 */
export function setupAuthMocks(user: TestUser = defaultTestUser) {
  (apiClient.auth.login as jest.Mock).mockResolvedValue({
    access_token: 'mock-access-token',
    refresh_token: 'mock-refresh-token',
    user: {
      id: user.id,
      email: user.email,
      name: user.name,
    },
  });
  
  (apiClient.auth.me as jest.Mock).mockResolvedValue({
    user: {
      id: user.id,
      email: user.email,
      name: user.name,
      organizations: user.organizations,
    },
  });
  
  (apiClient.auth.logout as jest.Mock).mockResolvedValue({});
}

/**
 * Mock organization operations
 */
export function setupOrganizationMocks(orgs: TestOrganization[] = [defaultTestOrg]) {
  (apiClient.organizations.list as jest.Mock).mockResolvedValue({
    organizations: orgs.map(org => ({
      id: org.id,
      name: org.name,
      owner_id: org.owner_id,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    })),
    total: orgs.length,
  });
  
  (apiClient.organizations.create as jest.Mock).mockImplementation(async (data) => ({
    id: `org-${Date.now()}`,
    name: data.name,
    owner_id: defaultTestUser.id,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }));
}

/**
 * Mock workspace operations
 */
export function setupWorkspaceMocks(workspaces: TestWorkspace[] = [defaultTestWorkspace]) {
  (apiClient.workspaces.list as jest.Mock).mockResolvedValue({
    workspaces: workspaces.map(ws => ({
      id: ws.id,
      organization_id: ws.organization_id,
      name: ws.name,
      plan: ws.plan,
      status: 'active',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    })),
    total: workspaces.length,
  });
  
  (apiClient.workspaces.create as jest.Mock).mockImplementation(async (orgId, data) => ({
    id: `ws-${Date.now()}`,
    organization_id: orgId,
    name: data.name,
    plan: data.plan,
    status: 'creating',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }));
}

/**
 * Mock project operations
 */
export function setupProjectMocks() {
  (apiClient.projects.list as jest.Mock).mockResolvedValue({
    projects: [],
    total: 0,
  });
  
  (apiClient.projects.create as jest.Mock).mockImplementation(async (orgId, wsId, data) => ({
    id: `proj-${Date.now()}`,
    workspace_id: wsId,
    name: data.name,
    namespace: data.name.toLowerCase().replace(/\s+/g, '-'),
    status: 'active',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }));
}

/**
 * Mock application operations
 */
export function setupApplicationMocks() {
  (apiClient.applications.list as jest.Mock).mockResolvedValue({
    data: {
      applications: [],
      total: 0,
    },
  });
  
  (apiClient.applications.create as jest.Mock).mockImplementation(async (orgId, wsId, projId, data) => ({
    data: {
      id: `app-${Date.now()}`,
      workspace_id: wsId,
      project_id: projId,
      name: data.name,
      type: data.type,
      status: 'pending',
      source_type: data.source_type,
      source_image: data.source_image,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
  }));
}

/**
 * Mock monitoring data
 */
export function setupMonitoringMocks() {
  (apiClient.monitoring.getWorkspaceMetrics as jest.Mock).mockResolvedValue({
    metrics: {
      cpu_usage: 35.5,
      memory_usage: 52.3,
      storage_usage: 25.7,
      network_ingress: 1024 * 1024 * 10, // 10MB
      network_egress: 1024 * 1024 * 25, // 25MB
    },
    timestamp: new Date().toISOString(),
  });
  
  (apiClient.monitoring.getActivityLogs as jest.Mock).mockResolvedValue({
    logs: [],
    total: 0,
  });
}

/**
 * Mock backup operations
 */
export function setupBackupMocks() {
  (apiClient.backup.listStorages as jest.Mock).mockResolvedValue({
    storages: [],
    total: 0,
  });
  
  (apiClient.backup.listPolicies as jest.Mock).mockResolvedValue({
    policies: [],
    total: 0,
  });
  
  (apiClient.backup.createStorage as jest.Mock).mockImplementation(async (wsId, data) => ({
    id: `bs-${Date.now()}`,
    workspace_id: wsId,
    name: data.name,
    type: data.type,
    status: 'active',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }));
}

/**
 * Helper to simulate user login
 */
export async function loginUser(user: any, email: string, password: string) {
  const emailInput = screen.getByLabelText(/email/i);
  const passwordInput = screen.getByLabelText(/password/i);
  const loginButton = screen.getByRole('button', { name: /sign in/i });
  
  await user.type(emailInput, email);
  await user.type(passwordInput, password);
  await user.click(loginButton);
  
  await waitFor(() => {
    expect(apiClient.auth.login).toHaveBeenCalledWith({ email, password });
  });
}

/**
 * Helper to create an organization
 */
export async function createOrganization(user: any, name: string) {
  const createButton = screen.getByRole('button', { name: /create organization/i });
  await user.click(createButton);
  
  const nameInput = screen.getByLabelText(/organization name/i);
  await user.type(nameInput, name);
  
  const submitButton = screen.getByRole('button', { name: /create/i });
  await user.click(submitButton);
  
  await waitFor(() => {
    expect(apiClient.organizations.create).toHaveBeenCalledWith({ name });
  });
}

/**
 * Helper to create a workspace
 */
export async function createWorkspace(user: any, orgId: string, name: string, plan: 'shared' | 'dedicated') {
  const createButton = screen.getByRole('button', { name: /create workspace/i });
  await user.click(createButton);
  
  const nameInput = screen.getByLabelText(/workspace name/i);
  await user.type(nameInput, name);
  
  const planSelect = screen.getByLabelText(/plan/i);
  await user.selectOptions(planSelect, plan);
  
  const submitButton = screen.getByRole('button', { name: /create/i });
  await user.click(submitButton);
  
  await waitFor(() => {
    expect(apiClient.workspaces.create).toHaveBeenCalledWith(orgId, { name, plan });
  });
}

/**
 * Helper to deploy an application
 */
export async function deployApplication(
  user: any,
  orgId: string,
  wsId: string,
  projId: string,
  appData: {
    name: string;
    type: string;
    source_type: string;
    source_image?: string;
  }
) {
  const deployButton = screen.getByRole('button', { name: /deploy application/i });
  await user.click(deployButton);
  
  const nameInput = screen.getByLabelText(/application name/i);
  await user.type(nameInput, appData.name);
  
  if (appData.source_image) {
    const imageInput = screen.getByLabelText(/image/i);
    await user.type(imageInput, appData.source_image);
  }
  
  const submitButton = screen.getByRole('button', { name: /deploy/i });
  await user.click(submitButton);
  
  await waitFor(() => {
    expect(apiClient.applications.create).toHaveBeenCalledWith(
      orgId, wsId, projId,
      expect.objectContaining(appData)
    );
  });
}

/**
 * Helper to check metrics
 */
export async function checkMetrics(expectedMetrics: {
  cpu?: number;
  memory?: number;
  storage?: number;
}) {
  await waitFor(() => {
    if (expectedMetrics.cpu !== undefined) {
      expect(screen.getByText(new RegExp(`${expectedMetrics.cpu}%`))).toBeInTheDocument();
    }
    if (expectedMetrics.memory !== undefined) {
      expect(screen.getByText(new RegExp(`${expectedMetrics.memory}%`))).toBeInTheDocument();
    }
    if (expectedMetrics.storage !== undefined) {
      expect(screen.getByText(new RegExp(`${expectedMetrics.storage}%`))).toBeInTheDocument();
    }
  });
}

/**
 * Helper to interact with AI chat
 */
export async function askAI(user: any, question: string) {
  const aiButton = screen.getByRole('button', { name: /ai assistant/i });
  await user.click(aiButton);
  
  const chatInput = screen.getByPlaceholderText(/ask ai/i);
  await user.type(chatInput, question);
  
  const sendButton = screen.getByRole('button', { name: /send/i });
  await user.click(sendButton);
  
  await waitFor(() => {
    expect(apiClient.aiops.chat).toHaveBeenCalledWith(
      expect.objectContaining({
        message: question,
      })
    );
  });
}