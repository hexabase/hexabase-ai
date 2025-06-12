'use client';

import { useEffect, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { LoadingSpinner } from '@/components/ui/loading';

interface ProtectedRouteProps {
  children: ReactNode;
  requiredPermissions?: string[];
  requiredRole?: string;
  requireOrganization?: boolean;
  fallback?: ReactNode;
}

export function ProtectedRoute({
  children,
  requiredPermissions,
  requiredRole,
  requireOrganization = false,
  fallback,
}: ProtectedRouteProps) {
  const router = useRouter();
  const {
    isLoading,
    isAuthenticated,
    user,
    organizations,
    activeOrganization,
    sessionExpired,
    switchOrganization,
    hasAllPermissions,
    hasRole,
  } = useAuth();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      const currentPath = encodeURIComponent(window.location.pathname);
      router.push(`/login?callbackUrl=${currentPath}`);
    }
  }, [isLoading, isAuthenticated, router]);

  if (isLoading) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" data-testid="loading-spinner" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return null; // Will redirect in useEffect
  }

  // Check permissions
  if (requiredPermissions && !hasAllPermissions(requiredPermissions)) {
    if (fallback) {
      return <>{fallback}</>;
    }
    return (
      <div className="flex h-screen w-full items-center justify-center px-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Access Denied</CardTitle>
            <CardDescription>
              You don't have permission to access this page.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  // Check role
  if (requiredRole && !hasRole(requiredRole)) {
    if (fallback) {
      return <>{fallback}</>;
    }
    return (
      <div className="flex h-screen w-full items-center justify-center px-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Access Denied</CardTitle>
            <CardDescription>
              You need {requiredRole} role to access this page.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  // Check if organization is required but not selected
  if (requireOrganization && !activeOrganization && organizations.length > 0) {
    return (
      <div className="flex h-screen w-full items-center justify-center px-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Select Organization</CardTitle>
            <CardDescription>
              Please select an organization to continue.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-2">
              {organizations.map((org) => (
                <Button
                  key={org.id}
                  variant="outline"
                  onClick={() => switchOrganization(org.id)}
                  className="w-full justify-start"
                >
                  {org.name}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Handle session expiry
  if (sessionExpired) {
    return (
      <>
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>Session Expired</CardTitle>
              <CardDescription>
                Your session has expired. Please sign in again to continue.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button 
                onClick={() => router.push('/login')}
                className="w-full"
              >
                Sign in again
              </Button>
            </CardContent>
          </Card>
        </div>
        {children}
      </>
    );
  }

  return <>{children}</>;
}