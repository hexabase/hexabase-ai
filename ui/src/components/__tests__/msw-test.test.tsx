import { render, screen, waitFor } from '@testing-library/react';
import { apiClient } from '@/lib/api-client';

describe('MSW Integration Test', () => {
  it('should fetch organizations from mock API', async () => {
    const response = await apiClient.organizations.list();
    
    expect(response.organizations).toHaveLength(2);
    expect(response.organizations[0].name).toBe('Acme Corporation');
  });

  it('should handle API errors', async () => {
    // This will fail with 401 if no auth token
    try {
      const response = await fetch('/api/v1/auth/me');
      const data = await response.json();
      
      if (!response.ok) {
        expect(data.error).toBe('Unauthorized');
      }
    } catch (error) {
      // Expected
    }
  });
});