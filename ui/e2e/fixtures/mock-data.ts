/**
 * Mock data fixtures for E2E tests
 */

export const testUsers = {
  admin: {
    id: 'user-e2e-admin',
    email: 'admin@e2e.test',
    password: 'Test123!',
    name: 'E2E Admin',
    role: 'admin',
  },
  developer: {
    id: 'user-e2e-dev',
    email: 'dev@e2e.test',
    password: 'Test123!',
    name: 'E2E Developer',
    role: 'developer',
  },
};

export const testOrganizations = [
  {
    id: 'org-e2e-1',
    name: 'E2E Test Organization',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    role: 'owner',
  },
  {
    id: 'org-e2e-2',
    name: 'E2E Secondary Org',
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
    role: 'member',
  },
];

export const testWorkspaces = [
  {
    id: 'ws-e2e-shared',
    name: 'E2E Shared Workspace',
    plan_id: 'shared',
    vcluster_status: 'active',
    vcluster_config: JSON.stringify({
      cpu: '4',
      memory: '8Gi',
      storage: '100Gi',
    }),
    vcluster_instance_name: 'e2e-shared-vcluster',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 'ws-e2e-dedicated',
    name: 'E2E Dedicated Workspace',
    plan_id: 'dedicated',
    vcluster_status: 'active',
    vcluster_config: JSON.stringify({
      cpu: '16',
      memory: '32Gi',
      storage: '500Gi',
    }),
    vcluster_instance_name: 'e2e-dedicated-vcluster',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

export const testProjects = [
  {
    id: 'proj-e2e-1',
    workspace_id: 'ws-e2e-shared',
    name: 'e2e-frontend',
    description: 'E2E test frontend project',
    namespace: 'e2e-frontend-ns',
    status: 'active',
    resource_quota: {
      cpu: '2',
      memory: '4Gi',
      storage: '20Gi',
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

export const testApplications = [
  {
    id: 'app-e2e-nginx',
    workspace_id: 'ws-e2e-shared',
    project_id: 'proj-e2e-1',
    name: 'e2e-nginx',
    type: 'stateless',
    status: 'running',
    source_type: 'image',
    source_image: 'nginx:latest',
    config: {
      replicas: 2,
      port: 80,
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 'app-e2e-postgres',
    workspace_id: 'ws-e2e-shared',
    project_id: 'proj-e2e-1',
    name: 'e2e-postgres',
    type: 'stateful',
    status: 'running',
    source_type: 'image',
    source_image: 'postgres:14',
    config: {
      port: 5432,
      storage_size: '10Gi',
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

export const deploymentStages = {
  pending: {
    status: 'pending',
    message: 'Deployment queued',
  },
  provisioning: {
    status: 'provisioning',
    message: 'Provisioning resources',
    progress: 25,
  },
  deploying: {
    status: 'deploying',
    message: 'Deploying application',
    progress: 50,
  },
  configuring: {
    status: 'configuring',
    message: 'Configuring services',
    progress: 75,
  },
  running: {
    status: 'running',
    message: 'Application is running',
    progress: 100,
  },
};

export const testMetrics = {
  workspace: {
    cpu_usage: 45.5,
    memory_usage: 62.3,
    storage_usage: 30.1,
    network_ingress: 1024 * 1024 * 100,
    network_egress: 1024 * 1024 * 200,
    pod_count: 15,
    container_count: 25,
  },
  application: {
    cpu_usage: 25.0,
    memory_usage: 512 * 1024 * 1024, // 512MB
    request_rate: 150,
    error_rate: 0.5,
    response_time_p50: 45,
    response_time_p95: 120,
    response_time_p99: 250,
  },
};