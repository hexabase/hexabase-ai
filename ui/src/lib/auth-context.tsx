'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useSession, signIn, signOut } from 'next-auth/react';
import { apiClient } from '@/lib/api-client';

interface User {
  id?: string;
  email: string;
  name?: string;
  image?: string;
  bio?: string;
  permissions?: string[];
  isNewUser?: boolean;
}

interface Organization {
  id: string;
  name: string;
  role: string;
}

interface AuthContextType {
  isLoading: boolean;
  isAuthenticated: boolean;
  user: User | null;
  organizations: Organization[];
  activeOrganization: Organization | null;
  sessionExpired: boolean;
  login: (provider: string, callbackUrl?: string) => Promise<void>;
  logout: () => Promise<void>;
  switchOrganization: (orgId: string) => void;
  refreshProfile: () => Promise<void>;
  hasPermission: (permission: string) => boolean;
  hasAnyPermission: (permissions: string[]) => boolean;
  hasAllPermissions: (permissions: string[]) => boolean;
  hasRole: (role: string) => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const { data: session, status } = useSession();
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [activeOrganization, setActiveOrganization] = useState<Organization | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [sessionExpired, setSessionExpired] = useState(false);

  useEffect(() => {
    if (session?.user) {
      setUser(session.user as User);
      
      // Check session expiry
      if (session.expires) {
        const expiryDate = new Date(session.expires);
        if (expiryDate < new Date()) {
          setSessionExpired(true);
        }
      }
      
      // Load organizations
      loadOrganizations();
    } else {
      setUser(null);
      setOrganizations([]);
      setActiveOrganization(null);
    }
  }, [session]);

  const loadOrganizations = async () => {
    try {
      const response = await apiClient.organizations.list();
      const orgs = response.data || [];
      setOrganizations(orgs);
      
      // Set first organization as active if none selected
      if (orgs.length > 0 && !activeOrganization) {
        setActiveOrganization(orgs[0]);
      }
    } catch (error) {
      console.error('Failed to load organizations:', error);
    }
  };

  const login = async (provider: string, callbackUrl: string = '/dashboard') => {
    await signIn(provider, { callbackUrl });
  };

  const logout = async () => {
    await signOut({ callbackUrl: '/login' });
  };

  const switchOrganization = (orgId: string) => {
    const org = organizations.find(o => o.id === orgId);
    if (org) {
      setActiveOrganization(org);
    }
  };

  const refreshProfile = async () => {
    try {
      const response = await apiClient.auth.getProfile();
      if (response.data) {
        setUser(response.data);
      }
    } catch (error) {
      console.error('Failed to refresh profile:', error);
    }
  };

  const hasPermission = (permission: string): boolean => {
    return user?.permissions?.includes(permission) || false;
  };

  const hasAnyPermission = (permissions: string[]): boolean => {
    return permissions.some(permission => hasPermission(permission));
  };

  const hasAllPermissions = (permissions: string[]): boolean => {
    return permissions.every(permission => hasPermission(permission));
  };

  const hasRole = (role: string): boolean => {
    return activeOrganization?.role === role;
  };


  const value: AuthContextType = {
    isLoading: status === 'loading',
    isAuthenticated: status === 'authenticated' && !sessionExpired,
    user,
    organizations,
    activeOrganization,
    sessionExpired,
    login,
    logout,
    switchOrganization,
    refreshProfile,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    hasRole,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}