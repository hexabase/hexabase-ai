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