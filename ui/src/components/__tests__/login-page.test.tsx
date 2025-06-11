import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { LoginPage } from '@/components/login-page';
import { useRouter, useSearchParams } from 'next/navigation';
import { signIn, useSession } from 'next-auth/react';

// Mock next-auth
jest.mock('next-auth/react', () => ({
  signIn: jest.fn(),
  useSession: jest.fn(() => ({ data: null, status: 'unauthenticated' })),
}));

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
  useSearchParams: jest.fn(() => ({
    get: jest.fn((key) => {
      if (key === 'error') return null;
      if (key === 'callbackUrl') return '/dashboard';
      return null;
    }),
  })),
}));

describe('LoginPage', () => {
  const mockPush = jest.fn();
  
  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue({ push: mockPush });
  });

  it('should render login page with OAuth providers', () => {
    render(<LoginPage />);
    
    expect(screen.getByRole('heading', { name: /sign in to hexabase ai/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /continue with google/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /continue with github/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /continue with microsoft/i })).toBeInTheDocument();
  });

  it('should display error message when error query param is present', () => {
    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn((key) => {
        if (key === 'error') return 'OAuthSignin';
        return null;
      }),
    });

    render(<LoginPage />);
    
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/error signing in/i)).toBeInTheDocument();
  });

  it('should call signIn with Google provider when Google button is clicked', async () => {
    render(<LoginPage />);
    
    const googleButton = screen.getByRole('button', { name: /continue with google/i });
    fireEvent.click(googleButton);

    await waitFor(() => {
      expect(signIn).toHaveBeenCalledWith('google', {
        callbackUrl: '/dashboard',
      });
    });
  });

  it('should call signIn with GitHub provider when GitHub button is clicked', async () => {
    render(<LoginPage />);
    
    const githubButton = screen.getByRole('button', { name: /continue with github/i });
    fireEvent.click(githubButton);

    await waitFor(() => {
      expect(signIn).toHaveBeenCalledWith('github', {
        callbackUrl: '/dashboard',
      });
    });
  });

  it('should call signIn with Azure AD provider when Microsoft button is clicked', async () => {
    render(<LoginPage />);
    
    const microsoftButton = screen.getByRole('button', { name: /continue with microsoft/i });
    fireEvent.click(microsoftButton);

    await waitFor(() => {
      expect(signIn).toHaveBeenCalledWith('azure-ad', {
        callbackUrl: '/dashboard',
      });
    });
  });

  it('should show loading state while signing in', async () => {
    render(<LoginPage />);
    
    const googleButton = screen.getByRole('button', { name: /continue with google/i });
    fireEvent.click(googleButton);

    expect(googleButton).toBeDisabled();
    expect(screen.getByTestId('google-loading-spinner')).toBeInTheDocument();
  });

  it('should redirect authenticated users to dashboard', () => {
    (useSession as jest.Mock).mockReturnValue({ 
      data: { user: { email: 'test@example.com' } }, 
      status: 'authenticated' 
    });

    render(<LoginPage />);

    expect(mockPush).toHaveBeenCalledWith('/dashboard');
  });

  it('should use custom callback URL from query params', async () => {
    (useSearchParams as jest.Mock).mockReturnValue({
      get: jest.fn((key) => {
        if (key === 'callbackUrl') return '/organizations';
        return null;
      }),
    });

    render(<LoginPage />);
    
    const googleButton = screen.getByRole('button', { name: /continue with google/i });
    fireEvent.click(googleButton);

    await waitFor(() => {
      expect(signIn).toHaveBeenCalledWith('google', {
        callbackUrl: '/organizations',
      });
    });
  });

  it('should display organization branding if configured', () => {
    render(<LoginPage orgBranding={{ logo: '/org-logo.png', name: 'ACME Corp' }} />);
    
    expect(screen.getByAltText('ACME Corp')).toBeInTheDocument();
    expect(screen.getByText(/sign in to acme corp/i)).toBeInTheDocument();
  });

  it('should show terms and privacy policy links', () => {
    render(<LoginPage />);
    
    expect(screen.getByRole('link', { name: /terms of service/i })).toHaveAttribute('href', '/terms');
    expect(screen.getByRole('link', { name: /privacy policy/i })).toHaveAttribute('href', '/privacy');
  });
});