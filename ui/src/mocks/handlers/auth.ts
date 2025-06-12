import { http, HttpResponse, delay } from 'msw';

// Mock user data
export const mockUser = {
  id: 'user-123',
  email: 'test@example.com',
  name: 'Test User',
  picture: 'https://example.com/avatar.jpg',
  organizations: ['org-1', 'org-2'],
};

export const authHandlers = [
  // Login endpoint
  http.post('/api/v1/auth/login', async ({ request }) => {
    const body = await request.json() as { email: string; password: string };
    
    await delay(100); // Simulate network delay
    
    if (body.email === 'test@example.com' && body.password === 'password') {
      return HttpResponse.json({
        access_token: 'mock-jwt-token',
        refresh_token: 'mock-refresh-token',
        user: mockUser,
      });
    }
    
    return HttpResponse.json(
      { error: 'Invalid credentials' },
      { status: 401 }
    );
  }),

  // Get current user
  http.get('/api/v1/auth/me', async ({ request }) => {
    const authHeader = request.headers.get('Authorization');
    
    if (authHeader === 'Bearer mock-jwt-token') {
      return HttpResponse.json({ user: mockUser });
    }
    
    return HttpResponse.json(
      { error: 'Unauthorized' },
      { status: 401 }
    );
  }),

  // Refresh token
  http.post('/api/v1/auth/refresh', async ({ request }) => {
    const body = await request.json() as { refresh_token: string };
    
    if (body.refresh_token === 'mock-refresh-token') {
      return HttpResponse.json({
        access_token: 'mock-jwt-token-refreshed',
        refresh_token: 'mock-refresh-token-new',
      });
    }
    
    return HttpResponse.json(
      { error: 'Invalid refresh token' },
      { status: 401 }
    );
  }),

  // Logout
  http.post('/api/v1/auth/logout', async () => {
    await delay(100);
    return HttpResponse.json({ message: 'Logged out successfully' });
  }),

  // OAuth endpoints
  http.get('/api/v1/auth/oauth/google', () => {
    return HttpResponse.json({
      auth_url: 'https://accounts.google.com/oauth/authorize?client_id=mock',
    });
  }),

  http.get('/api/v1/auth/oauth/callback', async ({ request }) => {
    const url = new URL(request.url);
    const code = url.searchParams.get('code');
    
    if (code === 'valid-oauth-code') {
      return HttpResponse.json({
        access_token: 'mock-oauth-token',
        refresh_token: 'mock-oauth-refresh',
        user: mockUser,
      });
    }
    
    return HttpResponse.json(
      { error: 'Invalid OAuth code' },
      { status: 400 }
    );
  }),
];