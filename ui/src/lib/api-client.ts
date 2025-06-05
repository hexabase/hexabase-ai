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

// Task Management Interfaces

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

// Task Management Interfaces

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
};

// Task Management Interfaces

// Project types
export interface Project {
  id: string;
  name: string;
  description?: string;
  workspace_id: string;
  workspace_name?: string;
  status: 'active' | 'inactive' | 'archived';
  namespace_count?: number;
  namespace_name?: string;
  resource_quotas?: {
    cpu_limit: string;
    memory_limit: string;
    storage_limit: string;
    pod_limit?: string;
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
  workspace_id: string;
  namespace_name?: string;
  resource_quotas?: {
    cpu_limit: string;
    memory_limit: string;
    storage_limit: string;
    pod_limit?: string;
  };
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
  create: async (orgId: string, data: CreateProjectRequest): Promise<Project> => {
    const response = await apiClient.post(`/api/v1/organizations/${orgId}/projects/`, data);
    return response.data;
  },

  // Update project
  update: async (orgId: string, projectId: string, data: Partial<CreateProjectRequest>): Promise<Project> => {
    const response = await apiClient.put(`/api/v1/organizations/${orgId}/projects/${projectId}`, data);
    return response.data;
  },

  // Delete project
  delete: async (orgId: string, projectId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/organizations/${orgId}/projects/${projectId}`);
  },

  // Get project statistics
  getStats: async (orgId: string, projectId: string): Promise<{ namespaces: number; pods: number; cpu_usage: string; memory_usage: string }> => {
    const response = await apiClient.get(`/api/v1/organizations/${orgId}/projects/${projectId}/stats`);
    return response.data;
  },
};

// Task Management Interfaces

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

// Task Management Interfaces

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

// Task Management Interfaces

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

// Task Management Interfaces
