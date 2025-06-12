import { render, screen, waitFor } from '@testing-library/react';
import { ProtectedRoute } from '@/components/protected-route';
import { useAuth } from '@/lib/auth-context';
import { useRouter } from 'next/navigation';

// Mock dependencies
jest.mock('@/lib/auth-context');
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

describe('ProtectedRoute', () => {
  const mockPush = jest.fn();
  const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue({ push: mockPush });
    // Reset the mock auth
    (mockUseAuth as any).mockReset?.();
  });

  it('should show loading spinner while authentication is loading', () => {
    mockUseAuth.mockReturnValue({
      isLoading: true,
      isAuthenticated: false,
      user: null,
      organizations: [],
      activeOrganization: null,
      login: jest.fn(),
      logout: jest.fn(),
      switchOrganization: jest.fn(),
      refreshProfile: jest.fn(),
      hasPermission: jest.fn(),
      hasAnyPermission: jest.fn(),
      hasAllPermissions: jest.fn(),
      hasRole: jest.fn(),
      sessionExpired: false,
    });

    render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>
    );

    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
  });

  it('should redirect to login when user is not authenticated', async () => {
    mockUseAuth.mockReturnValue({
      isLoading: false,
      isAuthenticated: false,
      user: null,
      organizations: [],
      activeOrganization: null,
      login: jest.fn(),
      logout: jest.fn(),
      switchOrganization: jest.fn(),
      refreshProfile: jest.fn(),
      hasPermission: jest.fn(),
      hasAnyPermission: jest.fn(),
      hasAllPermissions: jest.fn(),
      hasRole: jest.fn(),
      sessionExpired: false,
    });

    render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>
    );

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/login?callbackUrl=%2F');
    });
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
  });

  it('should render children when user is authenticated', () => {
    mockUseAuth.mockReturnValue({
      isLoading: false,
      isAuthenticated: true,
      user: { id: 'user-123', email: 'test@example.com', name: 'Test User' },
      organizations: [{ id: 'org-1', name: 'Org 1', role: 'admin' }],
      activeOrganization: { id: 'org-1', name: 'Org 1', role: 'admin' },
      login: jest.fn(),
      logout: jest.fn(),
      switchOrganization: jest.fn(),
      refreshProfile: jest.fn(),
      hasPermission: jest.fn(),
      hasAnyPermission: jest.fn(),
      hasAllPermissions: jest.fn(),
      hasRole: jest.fn(),
      sessionExpired: false,
    });

    render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>
    );

    expect(screen.getByText('Protected Content')).toBeInTheDocument();
    expect(mockPush).not.toHaveBeenCalled();
  });

  it('should check required permissions when specified', () => {
    const mockHasAllPermissions = jest.fn().mockReturnValue(false);
    
    mockUseAuth.mockReturnValue({
      isLoading: false,
      isAuthenticated: true,
      user: { id: 'user-123', email: 'test@example.com', name: 'Test User' },
      organizations: [],
      activeOrganization: null,
      login: jest.fn(),
      logout: jest.fn(),
      switchOrganization: jest.fn(),
      refreshProfile: jest.fn(),
      hasPermission: jest.fn(),
      hasAnyPermission: jest.fn(),
      hasAllPermissions: mockHasAllPermissions,
      hasRole: jest.fn(),
      sessionExpired: false,
    });

    render(
      <ProtectedRoute requiredPermissions={['workspace:create', 'project:delete']}>
        <div>Protected Content</div>
      </ProtectedRoute>
    );

    expect(mockHasAllPermissions).toHaveBeenCalledWith(['workspace:create', 'project:delete']);
    expect(screen.getByText(/you don't have permission to access this page/i)).toBeInTheDocument();
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
  });

  it('should check required role when specified', () => {
    const mockHasRole = jest.fn().mockReturnValue(false);
    
    mockUseAuth.mockReturnValue({
      isLoading: false,
      isAuthenticated: true,
      user: { id: 'user-123', email: 'test@example.com', name: 'Test User' },
      organizations: [],
      activeOrganization: { id: 'org-1', name: 'Org 1', role: 'member' },
      login: jest.fn(),
      logout: jest.fn(),
      switchOrganization: jest.fn(),
      refreshProfile: jest.fn(),
      hasPermission: jest.fn(),
      hasAnyPermission: jest.fn(),
      hasAllPermissions: jest.fn(),
      hasRole: mockHasRole,
      sessionExpired: false,
    });

    render(
      <ProtectedRoute requiredRole="admin">
        <div>Protected Content</div>
      </ProtectedRoute>
    );

    expect(mockHasRole).toHaveBeenCalledWith('admin');
    expect(screen.getByText(/you need admin role to access this page/i)).toBeInTheDocument();
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
  });

  it('should require organization selection when no active organization', () => {
    mockUseAuth.mockReturnValue({
      isLoading: false,
      isAuthenticated: true,
      user: { id: 'user-123', email: 'test@example.com', name: 'Test User' },
      organizations: [
        { id: 'org-1', name: 'Org 1', role: 'admin' },
        { id: 'org-2', name: 'Org 2', role: 'member' },
      ],
      activeOrganization: null,
      login: jest.fn(),
      logout: jest.fn(),
      switchOrganization: jest.fn(),
      refreshProfile: jest.fn(),
      hasPermission: jest.fn(),
      hasAnyPermission: jest.fn(),
      hasAllPermissions: jest.fn(),
      hasRole: jest.fn(),
      sessionExpired: false,
    });

    render(
      <ProtectedRoute requireOrganization>
        <div>Protected Content</div>
      </ProtectedRoute>
    );

    expect(screen.getByText(/please select an organization/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /org 1/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /org 2/i })).toBeInTheDocument();
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
  });

  it('should handle session expiry with modal', () => {
    mockUseAuth.mockReturnValue({
      isLoading: false,
      isAuthenticated: true,
      user: { id: 'user-123', email: 'test@example.com', name: 'Test User' },
      organizations: [],
      activeOrganization: null,
      login: jest.fn(),
      logout: jest.fn(),
      switchOrganization: jest.fn(),
      refreshProfile: jest.fn(),
      hasPermission: jest.fn(),
      hasAnyPermission: jest.fn(),
      hasAllPermissions: jest.fn(),
      hasRole: jest.fn(),
      sessionExpired: true,
    });

    render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>
    );

    expect(screen.getByText(/your session has expired/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in again/i })).toBeInTheDocument();
    expect(screen.getByText('Protected Content')).toBeInTheDocument(); // Still renders content behind modal
  });

  it('should preserve current path in login redirect', () => {
    // Mock window.location
    Object.defineProperty(window, 'location', {
      value: { pathname: '/organizations/org-123/workspaces' },
      writable: true,
    });

    mockUseAuth.mockReturnValue({
      isLoading: false,
      isAuthenticated: false,
      user: null,
      organizations: [],
      activeOrganization: null,
      login: jest.fn(),
      logout: jest.fn(),
      switchOrganization: jest.fn(),
      refreshProfile: jest.fn(),
      hasPermission: jest.fn(),
      hasAnyPermission: jest.fn(),
      hasAllPermissions: jest.fn(),
      hasRole: jest.fn(),
      sessionExpired: false,
    });

    render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>
    );

    expect(mockPush).toHaveBeenCalledWith(
      '/login?callbackUrl=%2Forganizations%2Forg-123%2Fworkspaces'
    );
  });

  it('should support custom fallback component', () => {
    mockUseAuth.mockReturnValue({
      isLoading: false,
      isAuthenticated: true,
      user: { id: 'user-123', email: 'test@example.com', name: 'Test User' },
      organizations: [],
      activeOrganization: null,
      login: jest.fn(),
      logout: jest.fn(),
      switchOrganization: jest.fn(),
      refreshProfile: jest.fn(),
      hasPermission: jest.fn(),
      hasAnyPermission: jest.fn(),
      hasAllPermissions: jest.fn().mockReturnValue(false),
      hasRole: jest.fn(),
      sessionExpired: false,
    });

    const CustomFallback = () => <div>Custom Access Denied Message</div>;

    render(
      <ProtectedRoute 
        requiredPermissions={['admin:all']} 
        fallback={<CustomFallback />}
      >
        <div>Protected Content</div>
      </ProtectedRoute>
    );

    expect(screen.getByText('Custom Access Denied Message')).toBeInTheDocument();
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
  });
});