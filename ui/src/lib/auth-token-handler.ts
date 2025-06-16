import { setCookie, deleteCookie } from 'cookies-next';

export const AUTH_COOKIE_NAME = 'hexabase_access_token';
export const REFRESH_COOKIE_NAME = 'hexabase_refresh_token';

export function setDevelopmentTokens(userId: string = 'dev-user-1') {
  if (process.env.NODE_ENV !== 'development') {
    return;
  }

  // Generate development tokens
  const accessToken = `dev_token_${userId}_${Date.now()}`;
  const refreshToken = `dev_refresh_${userId}_${Date.now()}`;

  // Set cookies with proper options
  const cookieOptions = {
    path: '/',
    sameSite: 'lax' as const,
    secure: false, // false for localhost development
    httpOnly: false, // Allow JavaScript access in development
    maxAge: 60 * 60 * 24 * 7, // 7 days
  };

  setCookie(AUTH_COOKIE_NAME, accessToken, cookieOptions);
  setCookie(REFRESH_COOKIE_NAME, refreshToken, cookieOptions);

  // Also set in localStorage as backup for development
  if (typeof window !== 'undefined') {
    localStorage.setItem('dev_access_token', accessToken);
    localStorage.setItem('dev_refresh_token', refreshToken);
  }

  return { accessToken, refreshToken };
}

export function getDevelopmentTokens() {
  if (process.env.NODE_ENV !== 'development') {
    return null;
  }

  // Try localStorage first (more reliable in development)
  if (typeof window !== 'undefined') {
    const accessToken = localStorage.getItem('dev_access_token');
    const refreshToken = localStorage.getItem('dev_refresh_token');
    if (accessToken && refreshToken) {
      return { accessToken, refreshToken };
    }
  }

  return null;
}

export function clearDevelopmentTokens() {
  deleteCookie(AUTH_COOKIE_NAME, { path: '/' });
  deleteCookie(REFRESH_COOKIE_NAME, { path: '/' });
  
  if (typeof window !== 'undefined') {
    localStorage.removeItem('dev_access_token');
    localStorage.removeItem('dev_refresh_token');
  }
}

export function isValidDevelopmentToken(token: string): boolean {
  return token.startsWith('dev_token_') || token.startsWith('dev_refresh_');
}