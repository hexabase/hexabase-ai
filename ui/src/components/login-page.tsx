'use client';

import { useState, useEffect } from 'react';
import { signIn, useSession } from 'next-auth/react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Icons } from '@/components/ui/icons';
import { Alert, AlertDescription } from '@/components/ui/alert';
import Link from 'next/link';

interface LoginPageProps {
  orgBranding?: {
    logo: string;
    name: string;
  };
}

const errorMessages: Record<string, string> = {
  OAuthSignin: 'Error constructing an authorization URL',
  OAuthCallback: 'Error handling the response from OAuth provider',
  OAuthCreateAccount: 'Could not create OAuth provider user',
  EmailCreateAccount: 'Could not create email provider user',
  Callback: 'Error in OAuth callback handler route',
  OAuthAccountNotLinked: 'Account is already linked with another provider',
  SessionRequired: 'Please sign in to continue',
  default: 'An error occurred during sign in',
};

export function LoginPage({ orgBranding }: LoginPageProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { data: session, status } = useSession();
  const [isLoading, setIsLoading] = useState<string | null>(null);

  const error = searchParams.get('error');
  const callbackUrl = searchParams.get('callbackUrl') || '/dashboard';

  useEffect(() => {
    if (status === 'authenticated' && session) {
      router.push(callbackUrl);
    }
  }, [status, session, router, callbackUrl]);

  const handleSignIn = async (provider: string) => {
    try {
      setIsLoading(provider);
      await signIn(provider, {
        callbackUrl,
      });
    } catch (error) {
      console.error('Sign in error:', error);
      setIsLoading(null);
    }
  };

  if (status === 'loading') {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="flex h-screen w-full items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          {orgBranding && (
            <div className="flex justify-center mb-4">
              <img src={orgBranding.logo} alt={orgBranding.name} className="h-12" />
            </div>
          )}
          <CardTitle className="text-2xl text-center">
            Sign in to {orgBranding?.name || 'Hexabase AI'}
          </CardTitle>
          <CardDescription className="text-center">
            Choose your preferred sign-in method
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4">
          {error && (
            <Alert variant="destructive" role="alert">
              <AlertDescription>
                Error signing in: {errorMessages[error] || errorMessages.default}
              </AlertDescription>
            </Alert>
          )}
          
          <Button
            variant="outline"
            onClick={() => handleSignIn('google')}
            disabled={isLoading !== null}
            className="w-full"
          >
            {isLoading === 'google' ? (
              <div 
                className="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-current" 
                data-testid="google-loading-spinner" 
              />
            ) : (
              <Icons.google className="mr-2 h-4 w-4" />
            )}
            Continue with Google
          </Button>

          <Button
            variant="outline"
            onClick={() => handleSignIn('github')}
            disabled={isLoading !== null}
            className="w-full"
          >
            {isLoading === 'github' ? (
              <div 
                className="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-current" 
                data-testid="github-loading-spinner" 
              />
            ) : (
              <Icons.gitHub className="mr-2 h-4 w-4" />
            )}
            Continue with GitHub
          </Button>

          <Button
            variant="outline"
            onClick={() => handleSignIn('azure-ad')}
            disabled={isLoading !== null}
            className="w-full"
          >
            {isLoading === 'azure-ad' ? (
              <div 
                className="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-current" 
                data-testid="microsoft-loading-spinner" 
              />
            ) : (
              <Icons.microsoft className="mr-2 h-4 w-4" />
            )}
            Continue with Microsoft
          </Button>
        </CardContent>
        <CardFooter>
          <p className="text-sm text-muted-foreground text-center w-full">
            By continuing, you agree to our{' '}
            <Link href="/terms" className="underline underline-offset-4 hover:text-primary">
              Terms of Service
            </Link>{' '}
            and{' '}
            <Link href="/privacy" className="underline underline-offset-4 hover:text-primary">
              Privacy Policy
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}