// Test data factory functions for consistent test data generation

let idCounter = 0

const generateId = (prefix: string) => {
  idCounter++
  return `${prefix}-${idCounter}`
}

export const resetIdCounter = () => {
  idCounter = 0
}

// Organization factory
export const createOrganization = (overrides = {}) => ({
  id: generateId('org'),
  name: 'Test Organization',
  slug: 'test-org',
  description: 'A test organization',
  billingEmail: 'billing@test.com',
  plan: 'starter',
  status: 'active',
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// Workspace factory
export const createWorkspace = (overrides = {}) => ({
  id: generateId('ws'),
  organizationId: 'org-1',
  name: 'Test Workspace',
  slug: 'test-workspace',
  plan: 'shared',
  status: 'active',
  resourceQuota: {
    cpu: '2',
    memory: '4Gi',
    storage: '10Gi',
  },
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// Project factory
export const createProject = (overrides = {}) => ({
  id: generateId('proj'),
  workspaceId: 'ws-1',
  name: 'Test Project',
  namespace: 'test-project',
  description: 'A test project',
  status: 'active',
  resourceQuota: {
    cpu: '1',
    memory: '2Gi',
    storage: '5Gi',
  },
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// Application factory
export const createApplication = (overrides = {}) => ({
  id: generateId('app'),
  projectId: 'proj-1',
  name: 'test-app',
  displayName: 'Test Application',
  type: 'deployment',
  image: 'nginx:latest',
  replicas: 1,
  status: 'running',
  resources: {
    requests: {
      cpu: '100m',
      memory: '128Mi',
    },
    limits: {
      cpu: '500m',
      memory: '512Mi',
    },
  },
  env: [],
  ports: [{ port: 80, protocol: 'TCP' }],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// CronJob factory
export const createCronJob = (overrides = {}) => ({
  id: generateId('cj'),
  applicationId: 'app-1',
  projectId: 'proj-1',
  name: 'test-cronjob',
  displayName: 'Test CronJob',
  schedule: '0 * * * *',
  enabled: true,
  command: ['echo', 'Hello World'],
  image: 'busybox:latest',
  lastExecution: null,
  nextExecution: new Date(Date.now() + 3600000).toISOString(),
  successfulJobsHistoryLimit: 3,
  failedJobsHistoryLimit: 1,
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// Function factory
export const createFunction = (overrides = {}) => ({
  id: generateId('fn'),
  projectId: 'proj-1',
  name: 'test-function',
  displayName: 'Test Function',
  runtime: 'node18',
  handler: 'index.handler',
  code: 'exports.handler = async (event) => { return { statusCode: 200, body: "Hello World" } }',
  env: {},
  timeout: 60,
  memory: 128,
  status: 'deployed',
  version: 'v1',
  triggers: [
    {
      type: 'http',
      path: '/test-function',
      methods: ['GET', 'POST'],
    },
  ],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// Execution factory (for CronJobs)
export const createCronJobExecution = (overrides = {}) => ({
  id: generateId('cje'),
  cronJobId: 'cj-1',
  status: 'succeeded',
  startTime: new Date(Date.now() - 300000).toISOString(),
  completionTime: new Date(Date.now() - 240000).toISOString(),
  duration: 60,
  logs: 'Hello World\\n',
  createdAt: new Date().toISOString(),
  ...overrides,
})

// Function invocation factory
export const createFunctionInvocation = (overrides = {}) => ({
  id: generateId('fi'),
  functionId: 'fn-1',
  version: 'v1',
  status: 'success',
  trigger: 'http',
  startTime: new Date(Date.now() - 1000).toISOString(),
  endTime: new Date().toISOString(),
  duration: 1000,
  statusCode: 200,
  request: {
    method: 'GET',
    path: '/test-function',
    headers: {},
  },
  response: {
    statusCode: 200,
    body: 'Hello World',
  },
  coldStart: false,
  billedDuration: 100,
  memoryUsed: 64,
  createdAt: new Date().toISOString(),
  ...overrides,
})

// Backup policy factory
export const createBackupPolicy = (overrides = {}) => ({
  id: generateId('bp'),
  workspaceId: 'ws-1',
  name: 'Daily Backup',
  type: 'full',
  schedule: '0 2 * * *',
  retention: {
    daily: 7,
    weekly: 4,
    monthly: 6,
  },
  targets: ['database', 'files'],
  enabled: true,
  lastExecution: null,
  nextExecution: new Date(Date.now() + 86400000).toISOString(),
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// Backup storage factory
export const createBackupStorage = (overrides = {}) => ({
  id: generateId('bs'),
  workspaceId: 'ws-1',
  name: 'Primary Backup Storage',
  type: 'proxmox',
  config: {
    host: 'proxmox.local',
    storage: 'backup-storage',
  },
  capacityGB: 1000,
  usedGB: 100,
  status: 'active',
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// Alert rule factory
export const createAlertRule = (overrides = {}) => ({
  id: generateId('alert'),
  workspaceId: 'ws-1',
  name: 'High CPU Usage',
  query: 'avg(cpu_usage) > 80',
  severity: 'warning',
  for: '5m',
  labels: {
    team: 'platform',
  },
  annotations: {
    summary: 'CPU usage is above 80%',
    description: 'The average CPU usage has been above 80% for 5 minutes',
  },
  enabled: true,
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// Metric data factory
export const createMetricData = (overrides = {}) => ({
  timestamp: new Date().toISOString(),
  value: Math.random() * 100,
  labels: {},
  ...overrides,
})

// User factory
export const createUser = (overrides = {}) => ({
  id: generateId('user'),
  email: 'test@example.com',
  name: 'Test User',
  avatar: 'https://example.com/avatar.jpg',
  role: 'admin',
  organizations: ['org-1'],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

// Session factory
export const createSession = (overrides = {}) => ({
  user: createUser(),
  accessToken: 'mock-access-token',
  refreshToken: 'mock-refresh-token',
  expiresAt: new Date(Date.now() + 3600000).toISOString(),
  ...overrides,
})