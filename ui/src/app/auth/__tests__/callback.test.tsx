import { render, screen, waitFor, act } from '@testing-library/react';
import { AuthCallbackPage } from '@/app/auth/callback/page';
import { useRouter, useSearchParams } from 'next/navigation';
import { signIn, useSession } from 'next-auth/react';

// Mock dependencies
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
  useSearchParams: jest.fn(),
}));
jest.mock('next-auth/react', () => ({
  useSession: jest.fn(),
}));

describe('AuthCallbackPage', () => {
  const mockPush = jest.fn();
  const mockReplace = jest.fn();
  
  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue({ 
      push: mockPush, 
      replace: mockReplace 
    });
  });

  it('should display processing message during authentication', () => {
    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn((key) => {
        if (key === 'code') return 'auth-code-123';
        if (key === 'state') return 'state-123';
        return null;
      }),
    });

    (useSession as jest.Mock).mockReturnValue({
      status: 'loading',
      data: null,
    });

    render(<AuthCallbackPage />);

    expect(screen.getByText(/processing authentication/i)).toBeInTheDocument();
    expect(screen.getByTestId('auth-spinner')).toBeInTheDocument();
  });

  it('should handle successful OAuth callback', async () => {
    const mockSession = {
      user: { email: 'test@example.com', name: 'Test User' },
      expires: new Date(Date.now() + 3600000).toISOString(),
    };

    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn((key) => {
        if (key === 'code') return 'auth-code-123';
        if (key === 'state') return 'state-123';
        if (key === 'callbackUrl') return '/dashboard';
        return null;
      }),
    });

    (useSession as jest.Mock).mockReturnValue({
      status: 'authenticated',
      data: mockSession,
    });

    render(<AuthCallbackPage />);

    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith('/dashboard');
    });
  });

  it('should redirect to default URL when no callback URL provided', async () => {
    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn(() => null),
    });

    (useSession as jest.Mock).mockReturnValue({
      status: 'authenticated',
      data: { user: { email: 'test@example.com' } },
    });

    render(<AuthCallbackPage />);

    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith('/dashboard/organizations');
    });
  });

  it('should handle OAuth errors', async () => {
    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn((key) => {
        if (key === 'error') return 'access_denied';
        if (key === 'error_description') return 'User denied access';
        return null;
      }),
    });

    (useSession as jest.Mock).mockReturnValue({
      status: 'unauthenticated',
      data: null,
    });

    render(<AuthCallbackPage />);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
      expect(screen.getByText(/authentication failed/i)).toBeInTheDocument();
      expect(screen.getByText(/user denied access/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /try again/i })).toBeInTheDocument();
    });
  });

  it('should handle missing authorization code', async () => {
    jest.useFakeTimers();
    
    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn(() => null),
    });

    (useSession as jest.Mock).mockReturnValue({
      status: 'unauthenticated',
      data: null,
    });

    render(<AuthCallbackPage />);

    expect(screen.getByText(/invalid authentication request/i)).toBeInTheDocument();
    
    act(() => {
      jest.advanceTimersByTime(2000);
    });

    expect(mockPush).toHaveBeenCalledWith('/login?error=InvalidRequest');
    
    jest.useRealTimers();
  });

  it('should validate state parameter for CSRF protection', async () => {
    jest.useFakeTimers();
    
    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn((key) => {
        if (key === 'code') return 'auth-code-123';
        if (key === 'state') return 'invalid-state';
        return null;
      }),
    });

    // Mock sessionStorage
    const mockSessionStorage = {
      getItem: jest.fn().mockReturnValue('expected-state'),
      removeItem: jest.fn(),
    };
    Object.defineProperty(window, 'sessionStorage', {
      value: mockSessionStorage,
      writable: true,
    });

    (useSession as jest.Mock).mockReturnValue({
      status: 'unauthenticated',
      data: null,
    });

    render(<AuthCallbackPage />);

    expect(screen.getByText(/security validation failed/i)).toBeInTheDocument();
    
    act(() => {
      jest.advanceTimersByTime(2000);
    });

    expect(mockPush).toHaveBeenCalledWith('/login?error=InvalidState');
    
    jest.useRealTimers();
  });

  it('should handle first-time user setup', async () => {
    const mockSession = {
      user: { 
        email: 'newuser@example.com', 
        name: 'New User',
        isNewUser: true,
      },
      expires: new Date(Date.now() + 3600000).toISOString(),
    };

    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn((key) => {
        if (key === 'code') return 'auth-code-123';
        return null;
      }),
    });

    (useSession as jest.Mock).mockReturnValue({
      status: 'authenticated',
      data: mockSession,
    });

    render(<AuthCallbackPage />);

    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith('/onboarding');
    });
  });

  it('should respect organization invitation flow', async () => {
    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn((key) => {
        if (key === 'code') return 'auth-code-123';
        if (key === 'invitation') return 'invite-token-123';
        if (key === 'org') return 'org-456';
        return null;
      }),
    });

    (useSession as jest.Mock).mockReturnValue({
      status: 'authenticated',
      data: { user: { email: 'test@example.com' } },
    });

    render(<AuthCallbackPage />);

    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith(
        '/organizations/org-456/accept-invite?token=invite-token-123'
      );
    });
  });

  it('should show timeout error for long processing', async () => {
    jest.useFakeTimers();

    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn((key) => {
        if (key === 'code') return 'auth-code-123';
        return null;
      }),
    });

    (useSession as jest.Mock).mockReturnValue({
      status: 'loading',
      data: null,
    });

    render(<AuthCallbackPage />);

    // Fast-forward 30 seconds
    act(() => {
      jest.advanceTimersByTime(30000);
    });

    await waitFor(() => {
      expect(screen.getByText(/authentication is taking longer than expected/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /return to login/i })).toBeInTheDocument();
    });

    jest.useRealTimers();
  });

  it('should prevent clickjacking with X-Frame-Options', () => {
    render(<AuthCallbackPage />);
    
    // Component should set security headers
    const meta = document.querySelector('meta[http-equiv="X-Frame-Options"]');
    expect(meta).toHaveAttribute('content', 'DENY');
  });
});