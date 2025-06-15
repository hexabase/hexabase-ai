'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useSession, signOut } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import { setDevelopmentTokens, clearDevelopmentTokens } from './auth-token-handler';
import { organizationsApi } from './api-client';

export interface User {
  id: string;
  email: string;
  name?: string;
}

export interface Organization {
  id: string;
  name: string;
  display_name?: string;
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

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data: session, status } = useSession();
  const router = useRouter();
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [activeOrganization, setActiveOrganization] = useState<Organization | null>(null);
  const [isLoadingOrgs, setIsLoadingOrgs] = useState(false);
  const [sessionExpired, setSessionExpired] = useState(false);

  // Convert NextAuth session to our User type
  const user: User | null = session?.user ? {
    id: session.user.id || 'dev-user-1',
    email: session.user.email || '',
    name: session.user.name || undefined,
  } : null;

  const isAuthenticated = status === 'authenticated';
  const isLoading = status === 'loading';

  // Set up development tokens when authenticated
  useEffect(() => {
    if (process.env.NODE_ENV !== 'development') {
      return;
    }

    if (isAuthenticated && session?.user?.email === 'test@hexabase.com') {
      setDevelopmentTokens();
    } else if (status === 'unauthenticated') {
      clearDevelopmentTokens();
    }
  }, [isAuthenticated, status, session]);

  // Fetch organizations when authenticated
  useEffect(() => {
    if (isAuthenticated && !isLoadingOrgs) {
      // Add a small delay to ensure tokens are set
      const timer = setTimeout(() => {
        fetchOrganizations();
      }, 500);
      return () => clearTimeout(timer);
    }
  }, [isAuthenticated]);

  const fetchOrganizations = async () => {
    setIsLoadingOrgs(true);
    try {
      const response = await organizationsApi.list();
      const orgs = response.organizations || [];
      setOrganizations(orgs);
      
      // Set first organization as active if none selected
      if (orgs.length > 0 && !activeOrganization) {
        setActiveOrganization(orgs[0]);
      }
    } catch (error: any) {
      console.error('Failed to fetch organizations:', error);
      // Handle 401 errors by marking session as expired
      if (error?.response?.status === 401) {
        setSessionExpired(true);
        // Don't try to fetch again
        return;
      }
      // For development, if we get a network error, it might be because tokens aren't set yet
      if (process.env.NODE_ENV === 'development' && error?.code === 'ERR_NETWORK') {
        console.log('Network error in development, will retry...');
      }
    } finally {
      setIsLoadingOrgs(false);
    }
  };

  const switchOrganization = useCallback((orgId: string) => {
    const org = organizations.find(o => o.id === orgId);
    if (org) {
      setActiveOrganization(org);
    }
  }, [organizations]);

  const login = useCallback(async (email: string, password: string) => {
    // This is handled by NextAuth signIn in the actual login page
    throw new Error('Use NextAuth signIn instead');
  }, []);

  const logout = useCallback(async () => {
    if (process.env.NODE_ENV === 'development') {
      clearDevelopmentTokens();
    }
    await signOut({ redirect: false });
    router.push('/');
  }, [router]);

  const refreshProfile = useCallback(async () => {
    await fetchOrganizations();
  }, []);

  // Permission helpers (simplified for now)
  const hasPermission = useCallback((permission: string) => {
    // In development, grant all permissions
    if (process.env.NODE_ENV === 'development') return true;
    return false;
  }, []);

  const hasAnyPermission = useCallback((permissions: string[]) => {
    return permissions.some(p => hasPermission(p));
  }, [hasPermission]);

  const hasAllPermissions = useCallback((permissions: string[]) => {
    return permissions.every(p => hasPermission(p));
  }, [hasPermission]);

  const hasRole = useCallback((role: string) => {
    if (!activeOrganization) return false;
    return activeOrganization.role === role;
  }, [activeOrganization]);

  const value: AuthContextType = {
    user,
    organizations,
    activeOrganization,
    switchOrganization,
    isLoading: isLoading || isLoadingOrgs,
    login,
    logout,
    isAuthenticated,
    refreshProfile,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    hasRole,
    sessionExpired,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}