import axios, { AxiosError } from 'axios';
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
    // Use new token name, fallback to legacy
    const token = Cookies.get('hexabase_access_token') || Cookies.get('hexabase_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling and token refresh
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      const refreshToken = Cookies.get('hexabase_refresh_token');
      if (refreshToken) {
        try {
          // Try to refresh the token
          const response = await axios.post(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/auth/refresh`, {
            refresh_token: refreshToken
          });
          
          const { access_token, refresh_token: newRefreshToken, expires_in } = response.data;
          const expiresAt = Date.now() + (expires_in * 1000);
          
          // Update cookies
          Cookies.set('hexabase_access_token', access_token, { expires: 7, secure: true, sameSite: 'strict' });
          Cookies.set('hexabase_refresh_token', newRefreshToken, { expires: 7, secure: true, sameSite: 'strict' });
          Cookies.set('hexabase_token_expires', expiresAt.toString(), { expires: 7, secure: true, sameSite: 'strict' });
          
          // Update the authorization header and retry
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return apiClient(originalRequest);
        } catch (refreshError) {
          // Refresh failed, clear auth and redirect
          Cookies.remove('hexabase_access_token');
          Cookies.remove('hexabase_refresh_token');
          Cookies.remove('hexabase_token_expires');
          Cookies.remove('hexabase_token'); // Legacy
          window.location.href = '/';
        }
      } else {
        // No refresh token, clear auth and redirect
        Cookies.remove('hexabase_access_token');
        Cookies.remove('hexabase_token'); // Legacy
        window.location.href = '/';
      }
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


export interface BackupStorage {
  id: string;
  workspace_id: string;
  name: string;
  type: 'nfs' | 'ceph' | 'local' | 'proxmox';
  proxmox_storage_id?: string;
  proxmox_node_id?: string;
  capacity_gb: number;
  used_gb: number;
  status: 'provisioning' | 'active' | 'error' | 'deleting';
  connection_config?: Record<string, any>;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface BackupPolicy {
  id: string;
  application_id: string;
  storage_id: string;
  enabled: boolean;
  schedule: string;
  retention_days: number;
  backup_type: 'full' | 'incremental' | 'application';
  include_volumes: boolean;
  include_database: boolean;
  include_config: boolean;
  compression_enabled: boolean;
  encryption_enabled: boolean;
  encryption_key_ref?: string;
  pre_backup_hook?: string;
  post_backup_hook?: string;
  created_at: string;
  updated_at: string;
}

export interface BackupExecution {
  id: string;
  policy_id: string;
  cronjob_execution_id?: string;
  status: 'running' | 'succeeded' | 'failed' | 'cancelled';
  size_bytes: number;
  compressed_size_bytes: number;
  backup_path: string;
  backup_manifest?: {
    volumes?: string[];
    databases?: string[];
    config_maps?: string[];
  };
  started_at: string;
  completed_at?: string;
  error_message?: string;
  metadata?: Record<string, any>;
  created_at: string;
}

export interface BackupRestore {
  id: string;
  backup_execution_id: string;
  application_id: string;
  status: 'pending' | 'running' | 'succeeded' | 'failed';
  restore_type: 'in_place' | 'new_application' | 'selective';
  restore_options: {
    restore_volumes?: boolean;
    restore_database?: boolean;
    restore_config?: boolean;
    stop_application?: boolean;
  };
  new_application_id?: string;
  started_at?: string;
  completed_at?: string;
  error_message?: string;
  validation_results?: Record<string, any>;
  created_at: string;
}

export interface BackupStorageUsage {
  storage_id: string;
  total_gb: number;
  used_gb: number;
  available_gb: number;
  usage_percent: number;
  backup_count: number;
  oldest_backup?: string;
  latest_backup?: string;
}

export interface CreateBackupStorageRequest {
  name: string;
  type: 'nfs' | 'ceph' | 'local' | 'proxmox';
  capacity_gb: number;
  connection_config?: Record<string, any>;
}

export interface CreateBackupPolicyRequest {
  storage_id: string;
  schedule: string;
  retention_days: number;
  backup_type?: 'full' | 'incremental' | 'application';
  include_volumes?: boolean;
  include_database?: boolean;
  include_config?: boolean;
  compression_enabled?: boolean;
  encryption_enabled?: boolean;
  enabled?: boolean;
}

export interface RestoreBackupRequest {
  backup_execution_id: string;
  restore_type: 'in_place' | 'new_application' | 'selective';
  restore_options?: {
    restore_volumes?: boolean;
    restore_database?: boolean;
    restore_config?: boolean;
    stop_application?: boolean;
  };
  new_application_name?: string;
}

// Backup API functions
export const backupApi = {
  // Backup Storage operations
  listBackupStorages: async (orgId: string, workspaceId: string): Promise<{ data: BackupStorage[] }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages`);
    return response;
  },

  getBackupStorage: async (orgId: string, workspaceId: string, storageId: string): Promise<{ data: BackupStorage }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages/${storageId}`);
    return response;
  },

  createBackupStorage: async (orgId: string, workspaceId: string, data: CreateBackupStorageRequest): Promise<{ data: BackupStorage & { task_id?: string } }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages`, data);
    return response;
  },

  updateBackupStorage: async (orgId: string, workspaceId: string, storageId: string, data: Partial<CreateBackupStorageRequest>): Promise<{ data: BackupStorage }> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages/${storageId}`, data);
    return response;
  },

  deleteBackupStorage: async (orgId: string, workspaceId: string, storageId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages/${storageId}`);
  },

  // Backup Policy operations
  listBackupPolicies: async (orgId: string, workspaceId: string): Promise<{ data: { policies: (BackupPolicy & { application_name?: string })[]; total: number } }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-policies`);
    return response;
  },

  getBackupPolicy: async (orgId: string, workspaceId: string, applicationId: string): Promise<{ data: BackupPolicy }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${applicationId}/backup-policy`);
    return response;
  },

  createBackupPolicy: async (orgId: string, workspaceId: string, applicationId: string, data: CreateBackupPolicyRequest): Promise<{ data: BackupPolicy }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${applicationId}/backup-policies`, data);
    return response;
  },

  updateBackupPolicy: async (orgId: string, workspaceId: string, policyId: string, data: Partial<CreateBackupPolicyRequest>): Promise<{ data: BackupPolicy }> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-policies/${policyId}`, data);
    return response;
  },

  deleteBackupPolicy: async (orgId: string, workspaceId: string, policyId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-policies/${policyId}`);
  },

  executeBackupPolicy: async (orgId: string, workspaceId: string, policyId: string): Promise<{ data: { execution_id: string; status: string; message: string } }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-policies/${policyId}/execute`);
    return response;
  },

  // Backup Execution operations
  listBackupExecutions: async (orgId: string, workspaceId: string, policyId: string, params?: { page?: number; page_size?: number }): Promise<{ data: { executions: BackupExecution[]; total: number; page: number; page_size: number } }> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.page_size) searchParams.append('page_size', params.page_size.toString());
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-policies/${policyId}/executions?${searchParams}`);
    return response;
  },

  getBackupExecution: async (orgId: string, workspaceId: string, executionId: string): Promise<{ data: BackupExecution }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-executions/${executionId}`);
    return response;
  },

  validateBackup: async (orgId: string, workspaceId: string, executionId: string): Promise<{ data: { valid: boolean; integrity_check: string; backup_manifest: any } }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-executions/${executionId}/validate`);
    return response;
  },

  // Restore operations
  restoreBackup: async (orgId: string, workspaceId: string, applicationId: string, data: RestoreBackupRequest): Promise<{ data: { restore_id: string; status: string; message: string; task_id?: string } }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${applicationId}/restore`, data);
    return response;
  },

  getRestoreStatus: async (orgId: string, workspaceId: string, restoreId: string): Promise<{ data: BackupRestore }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-restores/${restoreId}`);
    return response;
  },

  listRestores: async (orgId: string, workspaceId: string, applicationId: string, params?: { page?: number; page_size?: number }): Promise<{ data: { restores: BackupRestore[]; total: number; page: number; page_size: number } }> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.page_size) searchParams.append('page_size', params.page_size.toString());
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${applicationId}/restores?${searchParams}`);
    return response;
  },

  // Storage usage operations
  getStorageUsage: async (orgId: string, workspaceId: string): Promise<{ data: BackupStorageUsage[] }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/backup-storages/usage`);
    return response;
  },
};


// Application Management Interfaces
export interface Application {
  id: string;
  workspace_id: string;
  project_id: string;
  name: string;
  type: 'stateless' | 'stateful' | 'cronjob' | 'function';
  status: 'pending' | 'running' | 'active' | 'suspended' | 'error' | 'terminating';
  source_type: 'image' | 'git' | 'buildpack';
  source_image?: string;
  source_git_url?: string;
  source_git_ref?: string;
  config?: any;
  endpoints?: any;
  // CronJob specific fields
  cron_schedule?: string;
  cron_command?: string[];
  cron_args?: string[];
  template_app_id?: string;
  last_execution_at?: string;
  next_execution_at?: string;
  // Function specific fields
  function_runtime?: string;
  function_handler?: string;
  function_timeout?: number;
  function_memory?: number;
  created_at: string;
  updated_at: string;
}

export interface CronJobExecution {
  id: string;
  application_id: string;
  job_name: string;
  started_at: string;
  completed_at?: string;
  status: 'running' | 'succeeded' | 'failed' | 'cancelled';
  exit_code?: number;
  logs?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateApplicationRequest {
  name: string;
  type: 'stateless' | 'stateful' | 'cronjob' | 'function';
  source_type: 'image' | 'git' | 'buildpack';
  source_image?: string;
  source_git_url?: string;
  source_git_ref?: string;
  config?: any;
  // CronJob specific
  cron_schedule?: string;
  cron_command?: string[];
  cron_args?: string[];
  template_app_id?: string;
  // Function specific
  function_runtime?: string;
  function_handler?: string;
  function_timeout?: number;
  function_memory?: number;
  status?: string;
}

// Applications API functions
export const applicationsApi = {
  // List applications
  list: async (orgId: string, workspaceId: string, projectId: string | null, params?: { 
    type?: string; 
    status?: string; 
    is_template?: boolean 
  }): Promise<{ data: { applications: Application[]; total: number } }> => {
    const searchParams = new URLSearchParams();
    if (params?.type) searchParams.append('type', params.type);
    if (params?.status) searchParams.append('status', params.status);
    if (params?.is_template !== undefined) searchParams.append('is_template', params.is_template.toString());
    
    let url = `/api/v1/organizations/${orgId}/workspaces/${workspaceId}`;
    if (projectId) {
      url += `/projects/${projectId}`;
    }
    url += `/applications?${searchParams}`;
    
    const response = await apiClient.get(url);
    return response;
  },

  // Get application by ID
  get: async (orgId: string, workspaceId: string, appId: string): Promise<{ data: Application }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${appId}`);
    return response;
  },

  // Create application
  create: async (orgId: string, workspaceId: string, projectId: string, data: CreateApplicationRequest): Promise<{ data: Application }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}/applications`, data);
    return response;
  },

  // Update application
  update: async (orgId: string, workspaceId: string, appId: string, data: Partial<CreateApplicationRequest>): Promise<{ data: Application }> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${appId}`, data);
    return response;
  },

  // Update application status
  updateStatus: async (orgId: string, workspaceId: string, appId: string, data: { status: string }): Promise<{ data: Application }> => {
    const response = await apiClient.patch(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${appId}`, data);
    return response;
  },

  // Delete application
  delete: async (orgId: string, workspaceId: string, projectId: string, appId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}/applications/${appId}`);
  },

  // CronJob specific operations
  getCronJobExecutions: async (orgId: string, workspaceId: string, appId: string, params?: { 
    page?: number; 
    page_size?: number 
  }): Promise<{ data: { executions: CronJobExecution[]; total: number; page: number; page_size: number } }> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.page_size) searchParams.append('page_size', params.page_size.toString());
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${appId}/executions?${searchParams}`);
    return response;
  },

  triggerCronJob: async (orgId: string, workspaceId: string, appId: string): Promise<{ data: { execution_id: string; job_name: string; status: string; message: string } }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${appId}/trigger`);
    return response;
  },

  updateCronSchedule: async (orgId: string, workspaceId: string, appId: string, data: { schedule: string }): Promise<{ data: { id: string; cron_schedule: string; next_execution_at?: string } }> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${appId}/schedule`, data);
    return response;
  },

  saveAsTemplate: async (orgId: string, workspaceId: string, appId: string, data: { template_name: string; template_description?: string }): Promise<{ data: { template_id: string; message: string } }> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/applications/${appId}/save-as-template`, data);
    return response;
  },
};

// Functions API Interfaces
export interface FunctionConfig {
  id: string;
  workspace_id: string;
  project_id: string;
  name: string;
  description?: string;
  runtime: string; // 'nodejs18' | 'python39' | 'go119' etc
  handler: string;
  timeout: number; // in seconds
  memory: number; // in MB
  environment_vars?: Record<string, string>;
  triggers: string[]; // 'http' | 'event' | 'schedule'
  status: 'active' | 'updating' | 'error' | 'inactive';
  version: string;
  last_deployed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface FunctionVersion {
  version: string;
  deployed_at: string;
  deployed_by: string;
  status: 'active' | 'inactive';
  size_bytes?: number;
}

export interface FunctionInvocation {
  invocation_id: string;
  function_id: string;
  status: 'success' | 'error' | 'timeout';
  trigger_type: string;
  payload?: any;
  output?: any;
  error?: string;
  duration_ms?: number;
  logs?: string;
  started_at: string;
  completed_at?: string;
}

export interface CreateFunctionRequest {
  name: string;
  description?: string;
  runtime: string;
  handler: string;
  timeout?: number;
  memory?: number;
  environment_vars?: Record<string, string>;
  triggers?: string[];
  source?: File;
}

export interface DeployFunctionRequest {
  version?: string;
  source?: File;
  environment_vars?: Record<string, string>;
  rollback_to?: string;
}

// Functions API
export const functionsApi = {
  // List functions
  list: async (
    orgId: string, 
    workspaceId: string, 
    projectId?: string,
    params?: { 
      runtime?: string;
      status?: string;
      trigger?: string;
    }
  ): Promise<{ data: { functions: FunctionConfig[]; total: number } }> => {
    const searchParams = new URLSearchParams();
    if (params?.runtime) searchParams.append('runtime', params.runtime);
    if (params?.status) searchParams.append('status', params.status);
    if (params?.trigger) searchParams.append('trigger', params.trigger);
    
    let url = `/api/v1/organizations/${orgId}/workspaces/${workspaceId}`;
    if (projectId) {
      url += `/projects/${projectId}`;
    }
    url += `/functions?${searchParams}`;
    
    const response = await apiClient.get(url);
    return response;
  },

  // Get function by ID
  get: async (orgId: string, workspaceId: string, functionId: string): Promise<{ data: FunctionConfig }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/functions/${functionId}`);
    return response;
  },

  // Create function
  create: async (orgId: string, workspaceId: string, projectId: string, data: CreateFunctionRequest): Promise<{ data: FunctionConfig }> => {
    const formData = new FormData();
    Object.entries(data).forEach(([key, value]) => {
      if (key === 'source' && value instanceof File) {
        formData.append(key, value);
      } else if (key === 'environment_vars' || key === 'triggers') {
        formData.append(key, JSON.stringify(value));
      } else if (value !== undefined) {
        formData.append(key, String(value));
      }
    });

    const response = await apiClient.post(
      `/api/v1/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}/functions`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
    return response;
  },

  // Update function
  update: async (orgId: string, workspaceId: string, functionId: string, data: Partial<CreateFunctionRequest>): Promise<{ data: FunctionConfig }> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/functions/${functionId}`, data);
    return response;
  },

  // Deploy function
  deploy: async (orgId: string, workspaceId: string, functionId: string, data: DeployFunctionRequest): Promise<{ data: { version: string; status: string } }> => {
    const formData = new FormData();
    Object.entries(data).forEach(([key, value]) => {
      if (key === 'source' && value instanceof File) {
        formData.append(key, value);
      } else if (key === 'environment_vars') {
        formData.append(key, JSON.stringify(value));
      } else if (value !== undefined) {
        formData.append(key, String(value));
      }
    });

    const response = await apiClient.post(
      `/api/v1/organizations/${orgId}/workspaces/${workspaceId}/functions/${functionId}/deploy`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
    return response;
  },

  // Delete function
  delete: async (orgId: string, workspaceId: string, functionId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/functions/${functionId}`);
  },

  // Invoke function
  invoke: async (orgId: string, workspaceId: string, functionId: string, data: { trigger_type?: string; payload?: any; http_method?: string; headers?: Record<string, string> }): Promise<{ data: FunctionInvocation }> => {
    const response = await apiClient.post(
      `/api/v1/organizations/${orgId}/workspaces/${workspaceId}/functions/${functionId}/invoke`,
      data
    );
    return response;
  },

  // Get function logs
  getLogs: async (
    orgId: string, 
    workspaceId: string, 
    functionId: string,
    params?: { 
      start_time?: string;
      end_time?: string;
      limit?: number;
      invocation_id?: string;
    }
  ): Promise<{ data: { logs: string[]; total: number } }> => {
    const searchParams = new URLSearchParams();
    if (params?.start_time) searchParams.append('start_time', params.start_time);
    if (params?.end_time) searchParams.append('end_time', params.end_time);
    if (params?.limit) searchParams.append('limit', params.limit.toString());
    if (params?.invocation_id) searchParams.append('invocation_id', params.invocation_id);
    
    const response = await apiClient.get(
      `/api/v1/organizations/${orgId}/workspaces/${workspaceId}/functions/${functionId}/logs?${searchParams}`
    );
    return response;
  },

  // Get function versions
  getVersions: async (orgId: string, workspaceId: string, functionId: string): Promise<{ data: { versions: FunctionVersion[] } }> => {
    const response = await apiClient.get(
      `/api/v1/organizations/${orgId}/workspaces/${workspaceId}/functions/${functionId}/versions`
    );
    return response;
  },

  // Get function metrics
  getMetrics: async (
    orgId: string, 
    workspaceId: string, 
    functionId: string,
    params?: { 
      metric?: string;
      start_time?: string;
      end_time?: string;
      interval?: string;
    }
  ): Promise<{ data: { metrics: any } }> => {
    const searchParams = new URLSearchParams();
    if (params?.metric) searchParams.append('metric', params.metric);
    if (params?.start_time) searchParams.append('start_time', params.start_time);
    if (params?.end_time) searchParams.append('end_time', params.end_time);
    if (params?.interval) searchParams.append('interval', params.interval);
    
    const response = await apiClient.get(
      `/api/v1/organizations/${orgId}/workspaces/${workspaceId}/functions/${functionId}/metrics?${searchParams}`
    );
    return response;
  },
};


// Task Management Interfaces
export interface Task {
  id: string;
  type: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  progress?: number;
  message?: string;
  workspace_id?: string;
  organization_id?: string;
  result?: Record<string, any>;
  error?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

// Task API functions
export const taskApi = {
  // Get task by ID
  get: async (taskId: string): Promise<Task> => {
    const response = await apiClient.get(`/api/v1/tasks/${taskId}`);
    return response.data;
  },

  // List tasks
  list: async (params?: { 
    workspace_id?: string; 
    organization_id?: string; 
    status?: string;
    type?: string;
    limit?: number;
  }): Promise<{ tasks: Task[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.workspace_id) searchParams.append('workspace_id', params.workspace_id);
    if (params?.organization_id) searchParams.append('organization_id', params.organization_id);
    if (params?.status) searchParams.append('status', params.status);
    if (params?.type) searchParams.append('type', params.type);
    if (params?.limit) searchParams.append('limit', params.limit.toString());
    
    const response = await apiClient.get(`/api/v1/tasks?${searchParams}`);
    return response.data;
  },

  // Cancel task
  cancel: async (taskId: string): Promise<void> => {
    await apiClient.post(`/api/v1/tasks/${taskId}/cancel`);
  },

  // Retry failed task
  retry: async (taskId: string): Promise<Task> => {
    const response = await apiClient.post(`/api/v1/tasks/${taskId}/retry`);
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

  // Initiate OAuth login with PKCE
  login: async (provider: 'google' | 'github', pkceParams?: { code_challenge: string; code_challenge_method: string }) => {
    const response = await apiClient.post(`/auth/login/${provider}`, pkceParams);
    return response.data;
  },
  
  // Exchange authorization code for tokens
  callback: async (params: { code: string; state: string; code_verifier: string }) => {
    const response = await apiClient.post('/auth/callback', params);
    return response.data;
  },
  
  // Refresh access token
  refresh: async (refreshToken: string) => {
    const response = await apiClient.post('/auth/refresh', { refresh_token: refreshToken });
    return response.data;
  },

  // Logout with token revocation
  logout: async (refreshToken?: string) => {
    const response = await apiClient.post('/auth/logout', refreshToken ? { refresh_token: refreshToken } : {});
    return response.data;
  },

  // Get user profile
  getProfile: async (): Promise<{ data: any }> => {
    const response = await apiClient.get('/api/v1/auth/profile');
    return response;
  },

  // Update user profile
  updateProfile: async (data: any): Promise<{ data: any }> => {
    const response = await apiClient.put('/api/v1/auth/profile', data);
    return response;
  },
};

// Project types
export interface Project {
  id: string;
  name: string;
  description?: string;
  workspace_id: string;
  workspace_name?: string;
  namespace: string;
  status: 'active' | 'creating' | 'error' | 'suspended' | 'inactive' | 'archived';
  namespace_count?: number;
  resource_quota?: {
    cpu: string;
    memory: string;
    storage: string;
    pods?: string;
  };
  resources?: {
    deployments: number;
    services: number;
    pods: number;
    configmaps: number;
    secrets: number;
  };
  resource_usage?: {
    cpu: string;
    memory: string;
    pods: number;
  };
  created_at: string;
  updated_at: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  namespace?: string;
  resource_quota?: {
    cpu: string;
    memory: string;
    storage: string;
    pods?: string;
  };
}

export interface ProjectMember {
  id: string;
  project_id: string;
  user_id: string;
  user_email: string;
  user_name: string;
  role: 'admin' | 'developer' | 'viewer';
  added_at: string;
  added_by: string;
}

export interface AddProjectMemberRequest {
  user_email: string;
  role: 'admin' | 'developer' | 'viewer';
}

export interface ProjectActivity {
  id: string;
  project_id: string;
  type: 'member_added' | 'member_removed' | 'member_role_changed' | 
        'namespace_created' | 'namespace_deleted' | 'quota_changed' |
        'project_created' | 'project_updated' | 'project_archived';
  description: string;
  user_id: string;
  user_email: string;
  user_name: string;
  metadata?: Record<string, any>;
  created_at: string;
}

export interface Namespace {
  id: string;
  name: string;
  description?: string;
  project_id: string;
  status: 'active' | 'inactive';
  resource_quota: {
    cpu: string;
    memory: string;
    pods: number;
  };
  resource_usage: {
    cpu: string;
    memory: string;
    pods: number;
  };
  created_at: string;
  updated_at: string;
}

export interface CreateNamespaceRequest {
  name: string;
  description?: string;
  resource_quota?: {
    cpu: string;
    memory: string;
    pods: number;
  };
}

// Projects API functions
export const projectsApi = {
  // List projects for organization
  list: async (orgId: string, params?: { workspace_id?: string; status?: string; search?: string }): Promise<{ projects: Project[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.workspace_id) searchParams.append('workspace_id', params.workspace_id);
    if (params?.status) searchParams.append('status', params.status);
    if (params?.search) searchParams.append('search', params.search);
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/projects/?${searchParams}`);
    return response.data;
  },

  // Get project by ID
  get: async (orgId: string, projectId: string): Promise<Project> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/projects/${projectId}`);
    return response.data;
  },

  // Create new project
  create: async (orgId: string, workspaceId: string, data: CreateProjectRequest): Promise<Project> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/projects/`, data);
    return response.data;
  },

  // Update project
  update: async (orgId: string, workspaceId: string, projectId: string, data: Partial<CreateProjectRequest>): Promise<Project> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}`, data);
    return response.data;
  },

  // Delete project
  delete: async (orgId: string, workspaceId: string, projectId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}`);
  },

  // Get project statistics
  getStats: async (orgId: string, projectId: string): Promise<{ namespaces: number; pods: number; cpu_usage: string; memory_usage: string }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/projects/${projectId}/stats`);
    return response.data;
  },
};

// Namespaces API functions
export const namespacesApi = {
  // List namespaces for project
  list: async (orgId: string, projectId: string): Promise<{ namespaces: Namespace[]; total: number }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/projects/${projectId}/namespaces/`);
    return response.data;
  },

  // Get namespace by ID
  get: async (orgId: string, projectId: string, namespaceId: string): Promise<Namespace> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/projects/${projectId}/namespaces/${namespaceId}`);
    return response.data;
  },

  // Create new namespace
  create: async (orgId: string, projectId: string, data: CreateNamespaceRequest): Promise<Namespace> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/projects/${projectId}/namespaces/`, data);
    return response.data;
  },

  // Update namespace
  update: async (orgId: string, projectId: string, namespaceId: string, data: Partial<CreateNamespaceRequest>): Promise<Namespace> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/projects/${projectId}/namespaces/${namespaceId}`, data);
    return response.data;
  },

  // Delete namespace
  delete: async (orgId: string, projectId: string, namespaceId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/projects/${projectId}/namespaces/${namespaceId}`);
  },

  // Get namespace resource usage
  getUsage: async (orgId: string, projectId: string, namespaceId: string): Promise<{ cpu: string; memory: string; pods: number }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/projects/${projectId}/namespaces/${namespaceId}/usage`);
    return response.data;
  },
};


// Project Members API
export const projectMembersApi = {
  // List project members
  list: async (orgId: string, projectId: string): Promise<{ members: ProjectMember[]; total: number }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/projects/${projectId}/members/`);
    return response.data;
  },

  // Add member to project
  add: async (orgId: string, projectId: string, data: AddProjectMemberRequest): Promise<ProjectMember> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/projects/${projectId}/members/`, data);
    return response.data;
  },

  // Update member role
  updateRole: async (orgId: string, projectId: string, memberId: string, role: 'admin' | 'developer' | 'viewer'): Promise<ProjectMember> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/projects/${projectId}/members/${memberId}`, { role });
    return response.data;
  },

  // Remove member from project
  remove: async (orgId: string, projectId: string, memberId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/projects/${projectId}/members/${memberId}`);
  },
};


// Project Activity API
export const projectActivityApi = {
  // Get project activity timeline
  list: async (orgId: string, projectId: string, limit: number = 50): Promise<{ activities: ProjectActivity[]; total: number }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/projects/${projectId}/activity/`, {
      params: { limit }
    });
    return response.data;
  },
};

// Billing & Subscription Interfaces
export interface Subscription {
  id: string;
  organization_id: string;
  plan_id: string;
  plan_name: string;
  status: 'active' | 'canceled' | 'past_due' | 'unpaid';
  billing_cycle: 'monthly' | 'yearly';
  current_period_start: string;
  current_period_end: string;
  price_per_month: number;
  price_per_year: number;
  features: string[];
  limits: {
    workspaces: number;
    storage_gb: number;
    bandwidth_gb: number;
    support_level: string;
  };
}

export interface SubscriptionPlan {
  id: string;
  name: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  yearly_discount_percentage: number;
  features: string[];
  limits: {
    workspaces: number;
    storage_gb: number;
    bandwidth_gb: number;
    support_level: string;
  };
  popular: boolean;
}

export interface PaymentMethod {
  id: string;
  type: 'card' | 'bank_account';
  last_four: string;
  brand?: string;
  expiry_month?: number;
  expiry_year?: number;
  is_default: boolean;
  created_at: string;
}

export interface Invoice {
  id: string;
  organization_id: string;
  subscription_id: string;
  amount_due: number;
  amount_paid: number;
  currency: string;
  status: 'draft' | 'open' | 'paid' | 'void' | 'uncollectible';
  period_start: string;
  period_end: string;
  due_date: string;
  invoice_pdf?: string;
  created_at: string;
  line_items: InvoiceLineItem[];
}

export interface InvoiceLineItem {
  id: string;
  description: string;
  amount: number;
  quantity: number;
  unit_price: number;
  period_start: string;
  period_end: string;
}

export interface UsageMetrics {
  organization_id: string;
  period_start: string;
  period_end: string;
  workspaces_count: number;
  workspaces_limit: number;
  storage_used_gb: number;
  storage_limit_gb: number;
  bandwidth_used_gb: number;
  bandwidth_limit_gb: number;
  overage_charges: {
    storage: number;
    bandwidth: number;
    total: number;
  };
}

export interface BillingForecast {
  organization_id: string;
  projected_amount: number;
  projected_period_end: string;
  usage_trend: 'increasing' | 'stable' | 'decreasing';
  recommendations: string[];
}

// Billing API functions
export const billingApi = {
  // Get current subscription
  getSubscription: async (orgId: string): Promise<Subscription> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/billing/subscription`);
    return response.data;
  },

  // Get available subscription plans
  getPlans: async (): Promise<{ plans: SubscriptionPlan[]; total: number }> => {
    const response = await apiClient.get('/api/v1/billing/plans');
    return response.data;
  },

  // Upgrade/downgrade subscription
  updateSubscription: async (orgId: string, data: { plan_id: string; billing_cycle: 'monthly' | 'yearly' }): Promise<Subscription> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/billing/subscription`, data);
    return response.data;
  },

  // Cancel subscription
  cancelSubscription: async (orgId: string): Promise<{ message: string; effective_date: string }> => {
    const response = await apiClient.delete(`/api/v1/organizations/${orgId}/billing/subscription`);
    return response.data;
  },

  // Get payment methods
  getPaymentMethods: async (orgId: string): Promise<{ payment_methods: PaymentMethod[]; total: number }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/billing/payment-methods`);
    return response.data;
  },

  // Add payment method
  addPaymentMethod: async (orgId: string, data: { token: string; is_default?: boolean }): Promise<PaymentMethod> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/billing/payment-methods`, data);
    return response.data;
  },

  // Update payment method
  updatePaymentMethod: async (orgId: string, methodId: string, data: { is_default?: boolean }): Promise<PaymentMethod> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/billing/payment-methods/${methodId}`, data);
    return response.data;
  },

  // Delete payment method
  deletePaymentMethod: async (orgId: string, methodId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/billing/payment-methods/${methodId}`);
  },

  // Get invoices
  getInvoices: async (orgId: string, params?: { status?: string; limit?: number; starting_after?: string }): Promise<{ invoices: Invoice[]; total: number; has_more: boolean }> => {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.append('status', params.status);
    if (params?.limit) searchParams.append('limit', params.limit.toString());
    if (params?.starting_after) searchParams.append('starting_after', params.starting_after);
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/billing/invoices?${searchParams}`);
    return response.data;
  },

  // Get invoice by ID
  getInvoice: async (orgId: string, invoiceId: string): Promise<Invoice> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/billing/invoices/${invoiceId}`);
    return response.data;
  },

  // Download invoice PDF
  downloadInvoicePdf: async (orgId: string, invoiceId: string): Promise<Blob> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/billing/invoices/${invoiceId}/pdf`, {
      responseType: 'blob',
    });
    return response.data;
  },

  // Get usage metrics
  getUsageMetrics: async (orgId: string, params?: { period?: string }): Promise<UsageMetrics> => {
    const searchParams = new URLSearchParams();
    if (params?.period) searchParams.append('period', params.period);
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/billing/usage?${searchParams}`);
    return response.data;
  },

  // Get billing forecast
  getBillingForecast: async (orgId: string): Promise<BillingForecast> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/billing/forecast`);
    return response.data;
  },

  // Update billing settings
  updateBillingSettings: async (orgId: string, data: { usage_threshold?: number; email_notifications?: boolean; slack_webhook?: string }): Promise<{ message: string }> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/billing/settings`, data);
    return response.data;
  },

  // Get billing settings
  getBillingSettings: async (orgId: string): Promise<{ usage_threshold: number; email_notifications: boolean; slack_webhook?: string }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/billing/settings`);
    return response.data;
  },
};

// Monitoring & Observability Interfaces
export interface ClusterHealth {
  status: 'healthy' | 'degraded' | 'critical';
  uptime_seconds: number;
  last_check: string;
  nodes_total: number;
  nodes_healthy: number;
  nodes_unhealthy: number;
  services_total: number;
  services_healthy: number;
}

export interface ResourceMetrics {
  timestamp: string;
  cpu: {
    usage_percentage: number;
    cores_used: number;
    cores_total: number;
  };
  memory: {
    usage_percentage: number;
    used_gb: number;
    total_gb: number;
  };
  storage: {
    usage_percentage: number;
    used_gb: number;
    total_gb: number;
  };
  network: {
    ingress_mbps: number;
    egress_mbps: number;
  };
}

export interface WorkspaceMetrics {
  workspace_id: string;
  workspace_name: string;
  cpu_usage: number;
  memory_usage: number;
  storage_usage: number;
  pod_count: number;
  namespace_count: number;
  status: 'healthy' | 'warning' | 'critical';
  metrics_history?: ResourceMetrics[];
}

export interface Alert {
  id: string;
  severity: 'info' | 'warning' | 'error' | 'critical';
  title: string;
  description: string;
  workspace_id?: string;
  workspace_name?: string;
  resource_type: string;
  resource_name: string;
  metric_name: string;
  current_value: number;
  threshold_value: number;
  triggered_at: string;
  resolved_at?: string;
  status: 'active' | 'resolved' | 'acknowledged';
}

export interface LogEntry {
  timestamp: string;
  level: 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  workspace_id: string;
  namespace: string;
  pod: string;
  container: string;
  message: string;
  metadata?: Record<string, any>;
}

export interface PerformanceInsight {
  id: string;
  type: 'optimization' | 'bottleneck' | 'cost_saving';
  severity: 'low' | 'medium' | 'high';
  title: string;
  description: string;
  recommendation: string;
  potential_savings?: number;
  affected_resources: string[];
  created_at: string;
}

export interface AlertRule {
  id: string;
  name: string;
  description?: string;
  metric: string;
  condition: 'above' | 'below' | 'equals';
  threshold: number;
  duration_minutes: number;
  severity: 'info' | 'warning' | 'error' | 'critical';
  enabled: boolean;
  notification_channels: string[];
  created_at: string;
  updated_at: string;
}

// Monitoring API functions
export const monitoringApi = {
  // Get cluster health
  getClusterHealth: async (orgId: string): Promise<ClusterHealth> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/monitoring/health`);
    return response.data;
  },

  // Get resource metrics
  getResourceMetrics: async (orgId: string, params?: { time_range?: string; interval?: string }): Promise<ResourceMetrics[]> => {
    const searchParams = new URLSearchParams();
    if (params?.time_range) searchParams.append('time_range', params.time_range);
    if (params?.interval) searchParams.append('interval', params.interval);
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/monitoring/metrics?${searchParams}`);
    return response.data;
  },

  // Get workspace metrics
  getWorkspaceMetrics: async (orgId: string): Promise<{ workspaces: WorkspaceMetrics[]; total: number }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/monitoring/workspaces`);
    return response.data;
  },

  // Get workspace details
  getWorkspaceDetails: async (orgId: string, workspaceId: string, params?: { time_range?: string }): Promise<WorkspaceMetrics> => {
    const searchParams = new URLSearchParams();
    if (params?.time_range) searchParams.append('time_range', params.time_range);
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/monitoring/workspaces/${workspaceId}?${searchParams}`);
    return response.data;
  },

  // Get alerts
  getAlerts: async (orgId: string, params?: { status?: string; severity?: string; workspace_id?: string }): Promise<{ alerts: Alert[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.append('status', params.status);
    if (params?.severity) searchParams.append('severity', params.severity);
    if (params?.workspace_id) searchParams.append('workspace_id', params.workspace_id);
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/monitoring/alerts?${searchParams}`);
    return response.data;
  },

  // Acknowledge alert
  acknowledgeAlert: async (orgId: string, alertId: string): Promise<Alert> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/monitoring/alerts/${alertId}/acknowledge`);
    return response.data;
  },

  // Get logs
  getLogs: async (orgId: string, params?: { 
    workspace_id?: string; 
    namespace?: string; 
    pod?: string; 
    level?: string; 
    search?: string; 
    start_time?: string; 
    end_time?: string;
    limit?: number;
  }): Promise<{ logs: LogEntry[]; total: number; has_more: boolean }> => {
    const searchParams = new URLSearchParams();
    if (params?.workspace_id) searchParams.append('workspace_id', params.workspace_id);
    if (params?.namespace) searchParams.append('namespace', params.namespace);
    if (params?.pod) searchParams.append('pod', params.pod);
    if (params?.level) searchParams.append('level', params.level);
    if (params?.search) searchParams.append('search', params.search);
    if (params?.start_time) searchParams.append('start_time', params.start_time);
    if (params?.end_time) searchParams.append('end_time', params.end_time);
    if (params?.limit) searchParams.append('limit', params.limit.toString());
    
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/monitoring/logs?${searchParams}`);
    return response.data;
  },

  // Stream logs (WebSocket endpoint)
  streamLogs: (orgId: string, params?: { workspace_id?: string; namespace?: string; pod?: string }) => {
    const searchParams = new URLSearchParams();
    if (params?.workspace_id) searchParams.append('workspace_id', params.workspace_id);
    if (params?.namespace) searchParams.append('namespace', params.namespace);
    if (params?.pod) searchParams.append('pod', params.pod);
    
    const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/api/v1/organizations/${orgId}/monitoring/logs/stream?${searchParams}`;
    return new WebSocket(wsUrl);
  },

  // Get performance insights
  getInsights: async (orgId: string): Promise<{ insights: PerformanceInsight[]; total: number }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/monitoring/insights`);
    return response.data;
  },

  // Get alert rules
  getAlertRules: async (orgId: string): Promise<{ rules: AlertRule[]; total: number }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/monitoring/alert-rules`);
    return response.data;
  },

  // Create alert rule
  createAlertRule: async (orgId: string, data: Omit<AlertRule, 'id' | 'created_at' | 'updated_at'>): Promise<AlertRule> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/monitoring/alert-rules`, data);
    return response.data;
  },

  // Update alert rule
  updateAlertRule: async (orgId: string, ruleId: string, data: Partial<Omit<AlertRule, 'id' | 'created_at' | 'updated_at'>>): Promise<AlertRule> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/monitoring/alert-rules/${ruleId}`, data);
    return response.data;
  },

  // Delete alert rule
  deleteAlertRule: async (orgId: string, ruleId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/monitoring/alert-rules/${ruleId}`);
  },

  // Export logs
  exportLogs: async (orgId: string, params: { format: 'csv' | 'json'; start_time: string; end_time: string }): Promise<Blob> => {
    const searchParams = new URLSearchParams(params);
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/monitoring/logs/export?${searchParams}`, {
      responseType: 'blob',
    });
    return response.data;
  },
};


// Unified API client - for backwards compatibility
(apiClient as any).auth = authApi;
(apiClient as any).organizations = organizationsApi;
(apiClient as any).workspaces = workspacesApi;
(apiClient as any).backup = backupApi;
(apiClient as any).applications = applicationsApi;
(apiClient as any).functions = functionsApi;
(apiClient as any).billing = billingApi;
(apiClient as any).monitoring = monitoringApi;
