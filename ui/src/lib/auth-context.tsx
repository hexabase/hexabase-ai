"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import Cookies from 'js-cookie';
import { apiClient } from './api-client';

interface User {
  id: string;
  email: string;
  name: string;
  provider: string;
}

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (provider: 'google' | 'github') => Promise<void>;
  logout: () => Promise<void>;
  token: string | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [token, setToken] = useState<string | null>(null);

  const isAuthenticated = !!user && !!token;

  useEffect(() => {
    // Check for existing token in cookies
    const storedToken = Cookies.get('hexabase_token');
    if (storedToken) {
      setToken(storedToken);
      // Verify token and get user info
      fetchCurrentUser(storedToken);
    } else {
      setIsLoading(false);
    }
  }, []);

  const fetchCurrentUser = async (authToken: string) => {
    try {
      const response = await apiClient.get('/auth/me', {
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
      });
      setUser(response.data.user);
    } catch (error) {
      console.error('Failed to fetch current user:', error);
      // Token might be invalid, clear it
      Cookies.remove('hexabase_token');
      setToken(null);
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (provider: 'google' | 'github') => {
    try {
      setIsLoading(true);
      
      // Get OAuth URL from backend
      const response = await apiClient.post(`/auth/login/${provider}`);
      const { auth_url } = response.data;
      
      // Redirect to OAuth provider
      window.location.href = auth_url;
    } catch (error) {
      console.error(`Failed to initiate ${provider} login:`, error);
      setIsLoading(false);
    }
  };

  const logout = async () => {
    try {
      setIsLoading(true);
      
      // Call logout endpoint
      if (token) {
        await apiClient.post('/auth/logout', {}, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
      }
      
      // Clear local state
      setUser(null);
      setToken(null);
      Cookies.remove('hexabase_token');
      
      // Redirect to login page
      window.location.href = '/';
    } catch (error) {
      console.error('Failed to logout:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle OAuth callback
  useEffect(() => {
    const handleCallback = () => {
      const urlParams = new URLSearchParams(window.location.search);
      const token = urlParams.get('token');
      const error = urlParams.get('error');

      if (error) {
        console.error('OAuth error:', error);
        // You might want to show a toast notification here
        return;
      }

      if (token) {
        // Store token in cookie
        Cookies.set('hexabase_token', token, { expires: 7 }); // 7 days
        setToken(token);
        
        // Fetch user info
        fetchCurrentUser(token);
        
        // Clean up URL
        const url = new URL(window.location.href);
        url.searchParams.delete('token');
        url.searchParams.delete('state');
        window.history.replaceState({}, '', url.pathname);
      }
    };

    // Only run on client side
    if (typeof window !== 'undefined') {
      handleCallback();
    }
  }, []);

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated,
    login,
    logout,
    token,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};