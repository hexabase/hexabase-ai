'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useSession } from 'next-auth/react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';

export function AuthCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { data: session, status } = useSession();
  const [error, setError] = useState<string | null>(null);
  const [isTimeout, setIsTimeout] = useState(false);

  const code = searchParams.get('code');
  const state = searchParams.get('state');
  const errorParam = searchParams.get('error');
  const errorDescription = searchParams.get('error_description');
  const callbackUrl = searchParams.get('callbackUrl') || '/dashboard/organizations';
  const invitation = searchParams.get('invitation');
  const orgId = searchParams.get('org');

  useEffect(() => {
    // Add X-Frame-Options meta tag
    const meta = document.createElement('meta');
    meta.httpEquiv = 'X-Frame-Options';
    meta.content = 'DENY';
    document.head.appendChild(meta);

    return () => {
      document.head.removeChild(meta);
    };
  }, []);

  useEffect(() => {
    // Handle errors
    if (errorParam) {
      setError(`Authentication failed: ${errorDescription || errorParam}`);
      return;
    }

    // Check for missing code
    if (!code && status === 'unauthenticated') {
      setError('Invalid authentication request - missing authorization code');
      setTimeout(() => {
        router.push('/login?error=InvalidRequest');
      }, 2000);
      return;
    }

    // Validate state for CSRF protection
    if (state) {
      const storedState = sessionStorage.getItem('oauth_state');
      if (storedState && state !== storedState) {
        setError('Security validation failed');
        setTimeout(() => {
          router.push('/login?error=InvalidState');
        }, 2000);
        return;
      }
      sessionStorage.removeItem('oauth_state');
    }

    // Handle successful authentication
    if (status === 'authenticated' && session) {
      // Check if this is a new user
      if (session.user?.isNewUser) {
        router.replace('/onboarding');
        return;
      }

      // Check for organization invitation
      if (invitation && orgId) {
        router.replace(`/organizations/${orgId}/accept-invite?token=${invitation}`);
        return;
      }

      // Redirect to callback URL
      router.replace(callbackUrl);
    }
  }, [status, session, code, state, errorParam, router, callbackUrl, invitation, orgId, errorDescription]);

  useEffect(() => {
    // Set timeout warning after 30 seconds
    const timer = setTimeout(() => {
      setIsTimeout(true);
    }, 30000);

    return () => clearTimeout(timer);
  }, []);

  if (error) {
    return (
      <div className="flex h-screen w-full items-center justify-center px-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Authentication Error</CardTitle>
          </CardHeader>
          <CardContent>
            <Alert variant="destructive" role="alert">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
            <Button 
              onClick={() => router.push('/login')}
              className="w-full mt-4"
            >
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex h-screen w-full items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Processing Authentication</CardTitle>
          <CardDescription>
            Please wait while we complete your sign-in...
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-4">
          <div 
            className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary" 
            data-testid="auth-spinner"
          />
          {isTimeout && (
            <div className="text-center">
              <p className="text-sm text-muted-foreground mb-4">
                Authentication is taking longer than expected.
              </p>
              <Button 
                variant="outline"
                onClick={() => router.push('/login')}
              >
                Return to Login
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default AuthCallbackPage;