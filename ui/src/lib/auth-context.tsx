"use client";

import React, { createContext, useContext, useEffect, useState, useCallback, useRef } from 'react';
import Cookies from 'js-cookie';
import { apiClient } from './api-client';

interface User {
  id: string;
  email: string;
  name: string;
  provider: string;
}

interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
}

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (provider: 'google' | 'github') => Promise<void>;
  logout: () => Promise<void>;
  token: string | null;
  refreshToken: () => Promise<void>;
}

// PKCE helper functions
function generateCodeVerifier(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode.apply(null, Array.from(array)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
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
  const [tokens, setTokens] = useState<AuthTokens | null>(null);
  const refreshTimerRef = useRef<NodeJS.Timeout | null>(null);

  const isAuthenticated = !!user && !!token;

  useEffect(() => {
    // Check for existing tokens in cookies
    const storedAccessToken = Cookies.get('hexabase_access_token');
    const storedRefreshToken = Cookies.get('hexabase_refresh_token');
    const storedExpiresAt = Cookies.get('hexabase_token_expires');
    
    if (storedAccessToken && storedRefreshToken && storedExpiresAt) {
      const expiresAt = parseInt(storedExpiresAt, 10);
      setTokens({
        accessToken: storedAccessToken,
        refreshToken: storedRefreshToken,
        expiresAt
      });
      setToken(storedAccessToken);
      
      // Check if token is expired
      if (Date.now() >= expiresAt - 60000) { // Refresh 1 minute before expiry
        refreshAccessToken(storedRefreshToken);
      } else {
        // Verify token and get user info
        fetchCurrentUser(storedAccessToken);
        // Schedule token refresh
        scheduleTokenRefresh(expiresAt);
      }
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
      // Token might be invalid, try to refresh
      if (tokens?.refreshToken) {
        await refreshAccessToken(tokens.refreshToken);
      } else {
        clearAuth();
      }
    } finally {
      setIsLoading(false);
    }
  };

  const clearAuth = () => {
    // Clear all auth data
    setUser(null);
    setToken(null);
    setTokens(null);
    Cookies.remove('hexabase_access_token');
    Cookies.remove('hexabase_refresh_token');
    Cookies.remove('hexabase_token_expires');
    Cookies.remove('hexabase_token'); // Legacy
    
    // Clear refresh timer
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current);
      refreshTimerRef.current = null;
    }
  };

  const scheduleTokenRefresh = (expiresAt: number) => {
    // Clear existing timer
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current);
    }
    
    // Schedule refresh 1 minute before expiry
    const refreshTime = expiresAt - Date.now() - 60000;
    if (refreshTime > 0) {
      refreshTimerRef.current = setTimeout(() => {
        if (tokens?.refreshToken) {
          refreshAccessToken(tokens.refreshToken);
        }
      }, refreshTime);
    }
  };

  const refreshAccessToken = async (refreshToken: string) => {
    try {
      const response = await apiClient.post('/auth/refresh', {
        refresh_token: refreshToken
      });
      
      const { access_token, refresh_token, expires_in } = response.data;
      const expiresAt = Date.now() + (expires_in * 1000);
      
      // Update tokens
      const newTokens: AuthTokens = {
        accessToken: access_token,
        refreshToken: refresh_token,
        expiresAt
      };
      
      setTokens(newTokens);
      setToken(access_token);
      
      // Store in cookies
      Cookies.set('hexabase_access_token', access_token, { expires: 7, secure: true, sameSite: 'strict' });
      Cookies.set('hexabase_refresh_token', refresh_token, { expires: 7, secure: true, sameSite: 'strict' });
      Cookies.set('hexabase_token_expires', expiresAt.toString(), { expires: 7, secure: true, sameSite: 'strict' });
      
      // Schedule next refresh
      scheduleTokenRefresh(expiresAt);
      
      // Fetch user info if needed
      if (!user) {
        await fetchCurrentUser(access_token);
      }
    } catch (error) {
      console.error('Failed to refresh token:', error);
      clearAuth();
      window.location.href = '/';
    }
  };

  const login = async (provider: 'google' | 'github') => {
    try {
      setIsLoading(true);
      
      // Generate PKCE parameters
      const codeVerifier = generateCodeVerifier();
      const codeChallenge = await generateCodeChallenge(codeVerifier);
      
      // Store code verifier for later use
      sessionStorage.setItem('pkce_code_verifier', codeVerifier);
      
      // Get OAuth URL from backend with PKCE
      const response = await apiClient.post(`/auth/login/${provider}`, {
        code_challenge: codeChallenge,
        code_challenge_method: 'S256'
      });
      
      const { auth_url, state } = response.data;
      
      // Store state for verification
      sessionStorage.setItem('oauth_state', state);
      
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
      
      // Call logout endpoint to revoke tokens
      if (token && tokens?.refreshToken) {
        await apiClient.post('/auth/logout', {
          refresh_token: tokens.refreshToken
        }, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
      }
      
      // Clear local state
      clearAuth();
      
      // Clear PKCE data
      sessionStorage.removeItem('pkce_code_verifier');
      sessionStorage.removeItem('oauth_state');
      
      // Redirect to login page
      window.location.href = '/';
    } catch (error) {
      console.error('Failed to logout:', error);
      // Even if logout fails, clear local state
      clearAuth();
      window.location.href = '/';
    } finally {
      setIsLoading(false);
    }
  };

  // Handle OAuth callback
  useEffect(() => {
    const handleCallback = async () => {
      const urlParams = new URLSearchParams(window.location.search);
      const code = urlParams.get('code');
      const state = urlParams.get('state');
      const error = urlParams.get('error');

      if (error) {
        console.error('OAuth error:', error);
        const errorDescription = urlParams.get('error_description');
        if (errorDescription) {
          console.error('Error description:', errorDescription);
        }
        // Clear auth state on error
        sessionStorage.removeItem('pkce_code_verifier');
        sessionStorage.removeItem('oauth_state');
        return;
      }

      if (code && state) {
        try {
          // Verify state
          const storedState = sessionStorage.getItem('oauth_state');
          if (state !== storedState) {
            console.error('State mismatch - possible CSRF attack');
            return;
          }
          
          // Get code verifier
          const codeVerifier = sessionStorage.getItem('pkce_code_verifier');
          if (!codeVerifier) {
            console.error('Code verifier not found');
            return;
          }
          
          // Exchange code for tokens
          const response = await apiClient.post('/auth/callback', {
            code,
            state,
            code_verifier: codeVerifier
          });
          
          const { access_token, refresh_token, expires_in, user: userData } = response.data;
          const expiresAt = Date.now() + (expires_in * 1000);
          
          // Store tokens
          const newTokens: AuthTokens = {
            accessToken: access_token,
            refreshToken: refresh_token,
            expiresAt
          };
          
          setTokens(newTokens);
          setToken(access_token);
          setUser(userData);
          
          // Store in secure cookies
          Cookies.set('hexabase_access_token', access_token, { expires: 7, secure: true, sameSite: 'strict' });
          Cookies.set('hexabase_refresh_token', refresh_token, { expires: 7, secure: true, sameSite: 'strict' });
          Cookies.set('hexabase_token_expires', expiresAt.toString(), { expires: 7, secure: true, sameSite: 'strict' });
          
          // Schedule token refresh
          scheduleTokenRefresh(expiresAt);
          
          // Clean up
          sessionStorage.removeItem('pkce_code_verifier');
          sessionStorage.removeItem('oauth_state');
          
          // Clean up URL
          const url = new URL(window.location.href);
          url.searchParams.delete('code');
          url.searchParams.delete('state');
          window.history.replaceState({}, '', url.pathname);
          
          setIsLoading(false);
        } catch (error) {
          console.error('Failed to exchange code for tokens:', error);
          sessionStorage.removeItem('pkce_code_verifier');
          sessionStorage.removeItem('oauth_state');
          setIsLoading(false);
        }
      } else {
        // No callback parameters, check for legacy token
        const legacyToken = urlParams.get('token');
        if (legacyToken) {
          // Handle legacy callback (fallback)
          console.warn('Using legacy token callback - consider updating to PKCE flow');
          Cookies.set('hexabase_access_token', legacyToken, { expires: 7, secure: true, sameSite: 'strict' });
          setToken(legacyToken);
          fetchCurrentUser(legacyToken);
          
          // Clean up URL
          const url = new URL(window.location.href);
          url.searchParams.delete('token');
          url.searchParams.delete('state');
          window.history.replaceState({}, '', url.pathname);
        }
      }
    };

    // Only run on client side
    if (typeof window !== 'undefined') {
      handleCallback();
    }
  }, []);

  const refreshToken = useCallback(async () => {
    if (tokens?.refreshToken) {
      await refreshAccessToken(tokens.refreshToken);
    }
  }, [tokens]);

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated,
    login,
    logout,
    token,
    refreshToken,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};