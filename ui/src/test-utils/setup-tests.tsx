import React from 'react';
import { render as rtlRender } from '@testing-library/react';
import { AuthProvider } from '@/lib/auth-context';
import { Toaster } from '@/components/ui/toaster';

// Mock auth user for tests
export const mockAuthUser = {
  id: 'user-123',
  email: 'test@example.com',
  name: 'Test User',
  picture: 'https://example.com/avatar.jpg',
  organizations: ['org-1', 'org-2'],
};

// Custom render function that includes providers
export function renderWithProviders(
  ui: React.ReactElement,
  {
    authValue = {
      user: mockAuthUser,
      loading: false,
      login: jest.fn(),
      logout: jest.fn(),
      refreshToken: jest.fn(),
    },
    ...renderOptions
  } = {}
) {
  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <AuthProvider>
        {children}
        <Toaster />
      </AuthProvider>
    );
  }

  return rtlRender(ui, { wrapper: Wrapper, ...renderOptions });
}

// Re-export everything from RTL
export * from '@testing-library/react';

// Override render method
export { renderWithProviders as render };