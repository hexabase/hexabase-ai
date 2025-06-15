// Remove incorrect import - ApiClient is not an exported type

// Mock API client factory
export const createMockApiClient = () => {
  return {
    // Auth methods
    login: jest.fn().mockResolvedValue({ token: 'mock-token' }),
    logout: jest.fn().mockResolvedValue(undefined),
    refreshToken: jest.fn().mockResolvedValue({ token: 'new-mock-token' }),
    
    // Organization methods
    getOrganizations: jest.fn().mockResolvedValue([]),
    createOrganization: jest.fn().mockResolvedValue({ id: 'org-1', name: 'Test Org' }),
    updateOrganization: jest.fn().mockResolvedValue({ id: 'org-1', name: 'Updated Org' }),
    deleteOrganization: jest.fn().mockResolvedValue(undefined),
    
    // Workspace methods
    getWorkspaces: jest.fn().mockResolvedValue([]),
    createWorkspace: jest.fn().mockResolvedValue({ id: 'ws-1', name: 'Test Workspace' }),
    updateWorkspace: jest.fn().mockResolvedValue({ id: 'ws-1', name: 'Updated Workspace' }),
    deleteWorkspace: jest.fn().mockResolvedValue(undefined),
    
    // Project methods
    getProjects: jest.fn().mockResolvedValue([]),
    createProject: jest.fn().mockResolvedValue({ id: 'proj-1', name: 'Test Project' }),
    updateProject: jest.fn().mockResolvedValue({ id: 'proj-1', name: 'Updated Project' }),
    deleteProject: jest.fn().mockResolvedValue(undefined),
    
    // Application methods
    getApplications: jest.fn().mockResolvedValue([]),
    createApplication: jest.fn().mockResolvedValue({ id: 'app-1', name: 'Test App' }),
    updateApplication: jest.fn().mockResolvedValue({ id: 'app-1', name: 'Updated App' }),
    deleteApplication: jest.fn().mockResolvedValue(undefined),
    deployApplication: jest.fn().mockResolvedValue({ status: 'deployed' }),
    
    // CronJob methods
    getCronJobs: jest.fn().mockResolvedValue([]),
    createCronJob: jest.fn().mockResolvedValue({ id: 'cj-1', name: 'Test CronJob' }),
    updateCronJob: jest.fn().mockResolvedValue({ id: 'cj-1', name: 'Updated CronJob' }),
    deleteCronJob: jest.fn().mockResolvedValue(undefined),
    getCronJobExecutions: jest.fn().mockResolvedValue([]),
    triggerCronJob: jest.fn().mockResolvedValue({ executionId: 'exec-1' }),
    
    // Function methods
    getFunctions: jest.fn().mockResolvedValue([]),
    createFunction: jest.fn().mockResolvedValue({ id: 'fn-1', name: 'Test Function' }),
    updateFunction: jest.fn().mockResolvedValue({ id: 'fn-1', name: 'Updated Function' }),
    deleteFunction: jest.fn().mockResolvedValue(undefined),
    deployFunction: jest.fn().mockResolvedValue({ version: 'v1' }),
    invokeFunction: jest.fn().mockResolvedValue({ result: 'success' }),
    getFunctionInvocations: jest.fn().mockResolvedValue([]),
    
    // Monitoring methods
    getMetrics: jest.fn().mockResolvedValue({}),
    getLogs: jest.fn().mockResolvedValue([]),
    getAlerts: jest.fn().mockResolvedValue([]),
    createAlert: jest.fn().mockResolvedValue({ id: 'alert-1' }),
    
    // Backup methods
    getBackupPolicies: jest.fn().mockResolvedValue([]),
    createBackupPolicy: jest.fn().mockResolvedValue({ id: 'bp-1' }),
    getBackupStorages: jest.fn().mockResolvedValue([]),
    createBackupStorage: jest.fn().mockResolvedValue({ id: 'bs-1' }),
    triggerBackup: jest.fn().mockResolvedValue({ backupId: 'backup-1' }),
    
    // WebSocket
    connectWebSocket: jest.fn(),
    disconnectWebSocket: jest.fn(),
    subscribeToUpdates: jest.fn(),
    unsubscribeFromUpdates: jest.fn(),
  } as any
}

// Mock data factories
export const mockOrganization = (overrides = {}) => ({
  id: 'org-123',
  name: 'Test Organization',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
  ...overrides,
})

export const mockWorkspace = (overrides = {}) => ({
  id: 'ws-123',
  organizationId: 'org-123',
  name: 'Test Workspace',
  plan: 'shared',
  status: 'active',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
  ...overrides,
})

export const mockProject = (overrides = {}) => ({
  id: 'proj-123',
  workspaceId: 'ws-123',
  name: 'Test Project',
  namespace: 'test-project',
  status: 'active',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
  ...overrides,
})

export const mockApplication = (overrides = {}) => ({
  id: 'app-123',
  projectId: 'proj-123',
  name: 'Test Application',
  type: 'deployment',
  status: 'running',
  replicas: 1,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
  ...overrides,
})

export const mockCronJob = (overrides = {}) => ({
  id: 'cj-123',
  applicationId: 'app-123',
  name: 'Test CronJob',
  schedule: '0 * * * *',
  enabled: true,
  lastExecution: '2024-01-01T00:00:00Z',
  nextExecution: '2024-01-01T01:00:00Z',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
  ...overrides,
})

export const mockFunction = (overrides = {}) => ({
  id: 'fn-123',
  projectId: 'proj-123',
  name: 'test-function',
  runtime: 'node18',
  handler: 'index.handler',
  status: 'deployed',
  version: 'v1',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
  ...overrides,
})

export const mockUser = (overrides = {}) => ({
  id: 'user-123',
  email: 'test@example.com',
  name: 'Test User',
  role: 'admin',
  ...overrides,
})