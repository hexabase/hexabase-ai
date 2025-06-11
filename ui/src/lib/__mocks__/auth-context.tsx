import React, { createContext, useContext, ReactNode } from 'react';

export interface User {
  id: string;
  email: string;
  name?: string;
}

export interface Organization {
  id: string;
  name: string;
  role?: string;
  created_at?: string;
  updated_at?: string;
}

export interface AuthContextType {
  user: User | null;
  organizations: Organization[];
  activeOrganization: Organization | null;
  switchOrganization: (orgId: string) => void;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  isAuthenticated: boolean;
  refreshProfile: () => Promise<void>;
  hasPermission: (permission: string) => boolean;
  hasAnyPermission: (permissions: string[]) => boolean;
  hasAllPermissions: (permissions: string[]) => boolean;
  hasRole: (role: string) => boolean;
  sessionExpired: boolean;
}

const mockUser: User = {
  id: 'user-123',
  email: 'test@example.com',
  name: 'Test User',
};

const mockOrganizations: Organization[] = [
  { id: 'org-1', name: 'ACME Corp', role: 'admin', created_at: '2024-01-01', updated_at: '2024-01-01' },
  { id: 'org-2', name: 'Tech Startup', role: 'member', created_at: '2024-01-02', updated_at: '2024-01-02' },
];

let mockAuthContext: AuthContextType = {
  user: mockUser,
  organizations: mockOrganizations,
  activeOrganization: mockOrganizations[0],
  switchOrganization: jest.fn(),
  isLoading: false,
  login: jest.fn().mockResolvedValue(undefined),
  logout: jest.fn().mockResolvedValue(undefined),
  isAuthenticated: true,
  refreshProfile: jest.fn().mockResolvedValue(undefined),
  hasPermission: jest.fn().mockReturnValue(true),
  hasAnyPermission: jest.fn().mockReturnValue(true),
  hasAllPermissions: jest.fn().mockReturnValue(true),
  hasRole: jest.fn().mockReturnValue(true),
  sessionExpired: false,
};

const AuthContext = createContext<AuthContextType>(mockAuthContext);

export function AuthProvider({ children }: { children: ReactNode }) {
  return (
    <AuthContext.Provider value={mockAuthContext}>
      {children}
    </AuthContext.Provider>
  );
}

let customReturnValue: AuthContextType | null = null;

export function useAuth() {
  if (customReturnValue) {
    return customReturnValue;
  }
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// For test customization
export const mockAuth = mockAuthContext;

// Allow tests to override the return value
(useAuth as any).mockReturnValue = (value: AuthContextType) => {
  customReturnValue = value;
};

// Allow tests to reset the mock
(useAuth as any).mockReset = () => {
  customReturnValue = null;
  mockAuthContext = {
    user: mockUser,
    organizations: mockOrganizations,
    activeOrganization: mockOrganizations[0],
    switchOrganization: jest.fn(),
    isLoading: false,
    login: jest.fn().mockResolvedValue(undefined),
    logout: jest.fn().mockResolvedValue(undefined),
    isAuthenticated: true,
    refreshProfile: jest.fn().mockResolvedValue(undefined),
    hasPermission: jest.fn().mockReturnValue(true),
    hasAnyPermission: jest.fn().mockReturnValue(true),
    hasAllPermissions: jest.fn().mockReturnValue(true),
    hasRole: jest.fn().mockReturnValue(true),
    sessionExpired: false,
  };
};