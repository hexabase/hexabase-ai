import axios from 'axios';
import Cookies from 'js-cookie';

// Create axios instance with base configuration
export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = Cookies.get('hexabase_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      Cookies.remove('hexabase_token');
      window.location.href = '/';
    }
    return Promise.reject(error);
  }
);

// Organization API functions
export interface Organization {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
  role?: string;
}

export interface CreateOrganizationRequest {
  name: string;
}

export interface UpdateOrganizationRequest {
  name: string;
}

export const organizationsApi = {
  // List organizations
  list: async (): Promise<{ organizations: Organization[]; total: number }> => {
    const response = await apiClient.get('/api/v1/organizations/');
    return response.data;
  },

  // Get organization by ID
  get: async (id: string): Promise<Organization> => {
    const response = await apiClient.get(`/api/v1/organizations/${id}`);
    return response.data;
  },

  // Create organization
  create: async (data: CreateOrganizationRequest): Promise<Organization> => {
    const response = await apiClient.post('/api/v1/organizations/', data);
    return response.data;
  },

  // Update organization
  update: async (id: string, data: UpdateOrganizationRequest): Promise<Organization> => {
    const response = await apiClient.put(`/api/v1/organizations/${id}`, data);
    return response.data;
  },

  // Delete organization
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${id}`);
  },
};

// Workspace types
export interface Workspace {
  id: string;
  name: string;
  plan_id: string;
  vcluster_status: string;
  vcluster_config?: string;
  vcluster_instance_name?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateWorkspaceRequest {
  name: string;
  plan_id: string;
}

export interface Plan {
  id: string;
  name: string;
  description: string;
  price: number;
  currency: string;
  resource_limits?: string;
}

export interface VClusterHealth {
  healthy: boolean;
  components: Record<string, string>;
  resource_usage: Record<string, string>;
  last_checked: string;
}

export interface VClusterStatus {
  status: string;
  workspace: string;
  cluster_info: Record<string, unknown>;
  health?: VClusterHealth;
}

// Workspaces API functions
export const workspacesApi = {
  // List workspaces for organization
  list: async (orgId: string): Promise<{ workspaces: Workspace[]; total: number }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/`);
    return response.data;
  },

  // Get workspace by ID
  get: async (orgId: string, wsId: string): Promise<Workspace> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${wsId}`);
    return response.data;
  },

  // Create workspace
  create: async (orgId: string, data: CreateWorkspaceRequest): Promise<Workspace> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/`, data);
    return response.data;
  },

  // Update workspace
  update: async (orgId: string, wsId: string, data: Partial<CreateWorkspaceRequest>): Promise<Workspace> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/workspaces/${wsId}`, data);
    return response.data;
  },

  // Delete workspace
  delete: async (orgId: string, wsId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/workspaces/${wsId}`);
  },

  // Get kubeconfig
  getKubeconfig: async (orgId: string, wsId: string): Promise<{ kubeconfig: string; workspace: string; status: string }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${wsId}/kubeconfig`);
    return response.data;
  },
};

// VCluster API functions
export const vclusterApi = {
  // Get vCluster status
  getStatus: async (orgId: string, wsId: string): Promise<VClusterStatus> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${wsId}/vcluster/status`);
    return response.data;
  },

  // Start vCluster
  start: async (orgId: string, wsId: string): Promise<{ task_id: string; status: string; message: string }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${wsId}/vcluster/start`);
    return response.data;
  },

  // Stop vCluster
  stop: async (orgId: string, wsId: string): Promise<{ task_id: string; status: string; message: string }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${wsId}/vcluster/stop`);
    return response.data;
  },

  // Get vCluster health
  getHealth: async (orgId: string, wsId: string): Promise<VClusterHealth> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${wsId}/vcluster/health`);
    return response.data;
  },

  // Provision vCluster
  provision: async (orgId: string, wsId: string, config: Record<string, unknown>): Promise<{ task_id: string; status: string; message: string }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${wsId}/vcluster/provision`, config);
    return response.data;
  },

  // Upgrade vCluster
  upgrade: async (orgId: string, wsId: string, config: { target_version: string; strategy?: string }): Promise<{ task_id: string; status: string; message: string }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${wsId}/vcluster/upgrade`, config);
    return response.data;
  },

  // Backup vCluster
  backup: async (orgId: string, wsId: string, config: { backup_name: string; retention?: string }): Promise<{ task_id: string; status: string; message: string }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${wsId}/vcluster/backup`, config);
    return response.data;
  },

  // Restore vCluster
  restore: async (orgId: string, wsId: string, config: { backup_name: string; strategy?: string }): Promise<{ task_id: string; status: string; message: string }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${wsId}/vcluster/restore`, config);
    return response.data;
  },
};

// Plans API functions
export const plansApi = {
  // List available plans
  list: async (): Promise<{ plans: Plan[]; total: number }> => {
    const response = await apiClient.get('/api/v1/plans');
    return response.data;
  },

  // Get plan by ID
  get: async (planId: string): Promise<Plan> => {
    const response = await apiClient.get(`/api/v1/plans/${planId}`);
    return response.data;
  },
};

// Auth API functions
export const authApi = {
  // Get current user
  me: async () => {
    const response = await apiClient.get('/auth/me');
    return response.data;
  },

  // Initiate OAuth login
  login: async (provider: 'google' | 'github') => {
    const response = await apiClient.post(`/auth/login/${provider}`);
    return response.data;
  },

  // Logout
  logout: async () => {
    const response = await apiClient.post('/auth/logout');
    return response.data;
  },
};