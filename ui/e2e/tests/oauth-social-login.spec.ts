import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { setupMockAPI } from '../utils/mock-api';
import { SMOKE_TAG, CRITICAL_TAG } from '../utils/test-tags';

test.describe('OAuth and Social Login', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    
    await loginPage.goto();
  });

  test(`Google OAuth login flow ${SMOKE_TAG} ${CRITICAL_TAG}`, async ({ page, context }) => {
    // Mock Google OAuth endpoints
    await context.route('https://accounts.google.com/o/oauth2/v2/auth*', async (route) => {
      const url = new URL(route.request().url());
      const redirectUri = url.searchParams.get('redirect_uri');
      const state = url.searchParams.get('state');
      
      // Simulate Google OAuth consent screen
      await route.fulfill({
        status: 302,
        headers: {
          'Location': `${redirectUri}?code=mock-google-auth-code&state=${state}`
        }
      });
    });

    // Mock token exchange
    await page.route('**/api/auth/callback/google*', async (route) => {
      await route.fulfill({
        status: 302,
        headers: {
          'Location': '/dashboard',
          'Set-Cookie': 'session=mock-google-session; Path=/; HttpOnly; Secure'
        }
      });
    });

    // Click Google login button
    const googleLoginButton = page.getByTestId('login-with-google');
    await expect(googleLoginButton).toBeVisible();
    await expect(googleLoginButton).toContainText('Continue with Google');
    
    // Start OAuth flow
    const [popup] = await Promise.all([
      context.waitForEvent('page'),
      googleLoginButton.click()
    ]);

    // Verify popup opened with correct OAuth URL
    await expect(popup.url()).toContain('accounts.google.com');
    
    // Mock successful authentication in popup
    await popup.close();
    
    // Verify redirect to dashboard
    await page.waitForURL('**/dashboard');
    await expect(dashboardPage.welcomeMessage).toBeVisible();
    
    // Verify user info from Google OAuth
    await expect(page.getByTestId('user-avatar')).toBeVisible();
    await expect(page.getByTestId('user-email')).toContainText('@gmail.com');
  });

  test(`GitHub OAuth login flow ${SMOKE_TAG} ${CRITICAL_TAG}`, async ({ page, context }) => {
    // Mock GitHub OAuth endpoints
    await context.route('https://github.com/login/oauth/authorize*', async (route) => {
      const url = new URL(route.request().url());
      const redirectUri = url.searchParams.get('redirect_uri');
      const state = url.searchParams.get('state');
      
      // Simulate GitHub OAuth consent
      await route.fulfill({
        status: 302,
        headers: {
          'Location': `${redirectUri}?code=mock-github-auth-code&state=${state}`
        }
      });
    });

    // Mock token exchange
    await page.route('**/api/auth/callback/github*', async (route) => {
      await route.fulfill({
        status: 302,
        headers: {
          'Location': '/dashboard',
          'Set-Cookie': 'session=mock-github-session; Path=/; HttpOnly; Secure'
        }
      });
    });

    // Click GitHub login button
    const githubLoginButton = page.getByTestId('login-with-github');
    await expect(githubLoginButton).toBeVisible();
    await expect(githubLoginButton).toContainText('Continue with GitHub');
    
    // Mock GitHub user data
    await page.route('**/api/auth/session', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          user: {
            id: 'github-user-123',
            email: 'developer@github.com',
            name: 'Test Developer',
            image: 'https://avatars.githubusercontent.com/u/123456',
            provider: 'github'
          }
        })
      });
    });

    // Click login button
    await githubLoginButton.click();
    
    // Verify redirect to dashboard
    await page.waitForURL('**/dashboard');
    await expect(dashboardPage.welcomeMessage).toBeVisible();
    
    // Verify GitHub user info
    await expect(page.getByTestId('user-provider-badge')).toContainText('GitHub');
    await expect(page.getByTestId('user-email')).toContainText('developer@github.com');
  });

  test('OAuth error handling - user denies consent', async ({ page }) => {
    // Mock OAuth error response
    await page.route('**/api/auth/callback/google*', async (route) => {
      const url = new URL(route.request().url());
      if (url.searchParams.get('error') === 'access_denied') {
        await route.fulfill({
          status: 302,
          headers: {
            'Location': '/login?error=oauth_denied'
          }
        });
      }
    });

    // Simulate OAuth denial
    await page.goto('/api/auth/callback/google?error=access_denied&error_description=User+denied+access');
    
    // Verify error message
    await expect(page).toHaveURL(/\/login\?error=oauth_denied/);
    await expect(page.getByTestId('auth-error')).toBeVisible();
    await expect(page.getByTestId('auth-error')).toContainText('Authentication was cancelled');
  });

  test('OAuth state mismatch protection', async ({ page }) => {
    // Try to access callback with invalid state
    await page.goto('/api/auth/callback/google?code=test&state=invalid-state');
    
    // Should redirect to login with error
    await expect(page).toHaveURL(/\/login\?error=invalid_state/);
    await expect(page.getByTestId('auth-error')).toContainText('Invalid authentication state');
  });

  test('first-time OAuth user registration flow', async ({ page, context }) => {
    // Mock Google OAuth with new user
    await context.route('https://accounts.google.com/o/oauth2/v2/auth*', async (route) => {
      const url = new URL(route.request().url());
      const redirectUri = url.searchParams.get('redirect_uri');
      const state = url.searchParams.get('state');
      
      await route.fulfill({
        status: 302,
        headers: {
          'Location': `${redirectUri}?code=new-user-auth-code&state=${state}`
        }
      });
    });

    // Mock new user creation
    await page.route('**/api/auth/callback/google*', async (route) => {
      await route.fulfill({
        status: 302,
        headers: {
          'Location': '/onboarding/welcome',
          'Set-Cookie': 'session=new-user-session; Path=/; HttpOnly; Secure'
        }
      });
    });

    // Click Google login
    await page.getByTestId('login-with-google').click();
    
    // Should redirect to onboarding for new users
    await page.waitForURL('**/onboarding/welcome');
    await expect(page.getByText('Welcome to Hexabase AI')).toBeVisible();
    await expect(page.getByTestId('onboarding-step-1')).toBeVisible();
    
    // Complete onboarding
    await page.getByTestId('organization-name-input').fill('My Company');
    await page.getByTestId('organization-type-select').selectOption('startup');
    await page.getByTestId('continue-button').click();
    
    // Verify workspace creation prompt
    await expect(page.getByText('Create your first workspace')).toBeVisible();
  });

  test('link existing account with OAuth provider', async ({ page }) => {
    // First login with email/password
    await loginPage.login('user@example.com', 'password123');
    await loginPage.isLoggedIn();
    
    // Navigate to account settings
    await page.getByTestId('user-menu-button').click();
    await page.getByTestId('user-menu-settings').click();
    
    // Go to connected accounts
    const connectedAccountsTab = page.getByRole('tab', { name: /connected accounts/i });
    await connectedAccountsTab.click();
    
    // Verify no OAuth providers linked
    await expect(page.getByTestId('google-account-status')).toContainText('Not connected');
    await expect(page.getByTestId('github-account-status')).toContainText('Not connected');
    
    // Link Google account
    const linkGoogleButton = page.getByTestId('link-google-account');
    await linkGoogleButton.click();
    
    // Mock successful linking
    await page.route('**/api/auth/link/google', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          provider: 'google',
          email: 'user@gmail.com'
        })
      });
    });
    
    // Complete OAuth flow
    await page.waitForTimeout(1000);
    
    // Verify account linked
    await expect(page.getByTestId('google-account-status')).toContainText('Connected');
    await expect(page.getByTestId('google-account-email')).toContainText('user@gmail.com');
    
    // Verify can now login with Google
    await loginPage.logout();
    await loginPage.goto();
    await page.getByTestId('login-with-google').click();
    await page.waitForURL('**/dashboard');
    await expect(dashboardPage.welcomeMessage).toBeVisible();
  });

  test('unlink OAuth provider from account', async ({ page }) => {
    // Login and navigate to settings
    await loginPage.login('user@example.com', 'password123');
    await page.getByTestId('user-menu-button').click();
    await page.getByTestId('user-menu-settings').click();
    
    const connectedAccountsTab = page.getByRole('tab', { name: /connected accounts/i });
    await connectedAccountsTab.click();
    
    // Mock account with linked providers
    await page.route('**/api/auth/linked-accounts', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          providers: [
            { provider: 'google', email: 'user@gmail.com', linked_at: '2025-01-01' },
            { provider: 'github', email: 'user@github.com', linked_at: '2025-01-02' }
          ]
        })
      });
    });
    
    await page.reload();
    
    // Unlink GitHub account
    const unlinkGithubButton = page.getByTestId('unlink-github-account');
    await unlinkGithubButton.click();
    
    // Confirm unlink
    const confirmDialog = page.getByRole('dialog');
    await expect(confirmDialog).toContainText('Unlink GitHub account?');
    await confirmDialog.getByTestId('confirm-unlink-button').click();
    
    // Mock successful unlink
    await page.route('**/api/auth/unlink/github', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true })
      });
    });
    
    // Verify account unlinked
    await expect(page.getByTestId('github-account-status')).toContainText('Not connected');
    await expect(page.getByTestId('success-message')).toContainText('GitHub account unlinked');
  });

  test('OAuth login with organization restrictions', async ({ page }) => {
    // Mock organization-restricted OAuth
    await page.route('**/api/auth/callback/google*', async (route) => {
      await route.fulfill({
        status: 403,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'organization_restriction',
          message: 'Your email domain is not allowed for this organization'
        })
      });
    });
    
    // Attempt Google login
    await page.getByTestId('login-with-google').click();
    
    // Verify organization restriction error
    await expect(page.getByTestId('auth-error')).toBeVisible();
    await expect(page.getByTestId('auth-error')).toContainText('email domain is not allowed');
    
    // Verify can still use regular login
    await loginPage.login('allowed@company.com', 'password123');
    await loginPage.isLoggedIn();
  });

  test('OAuth session refresh and expiry', async ({ page }) => {
    // Mock OAuth session
    await page.route('**/api/auth/session', async (route, request) => {
      const iteration = request.url().includes('refresh') ? 2 : 1;
      
      if (iteration === 1) {
        // First call - session expired
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'session_expired' })
        });
      } else {
        // After refresh - new session
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            user: {
              id: 'user-123',
              email: 'user@gmail.com',
              provider: 'google'
            }
          })
        });
      }
    });
    
    // Navigate to protected page
    await page.goto('/dashboard');
    
    // Should trigger session refresh
    await page.waitForTimeout(1000);
    
    // Verify still on dashboard (session refreshed)
    await expect(page).toHaveURL(/\/dashboard/);
    await expect(dashboardPage.welcomeMessage).toBeVisible();
  });

  test('multiple OAuth providers on login page @visual', async ({ page }) => {
    // Verify all OAuth options are visible
    await expect(page.getByTestId('login-with-google')).toBeVisible();
    await expect(page.getByTestId('login-with-github')).toBeVisible();
    
    // Verify traditional login is still available
    await expect(loginPage.emailInput).toBeVisible();
    await expect(loginPage.passwordInput).toBeVisible();
    await expect(loginPage.submitButton).toBeVisible();
    
    // Verify OAuth separator
    await expect(page.getByText('Or continue with')).toBeVisible();
    
    // Take screenshot for visual regression
    await page.screenshot({
      path: 'test-results/oauth-login-options.png',
      fullPage: true
    });
  });
});