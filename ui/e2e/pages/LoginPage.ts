import { Page, Locator } from '@playwright/test';

export class LoginPage {
  readonly page: Page;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly loginButton: Locator;
  readonly errorMessage: Locator;
  readonly forgotPasswordLink: Locator;
  readonly googleLoginButton: Locator;
  readonly githubLoginButton: Locator;
  readonly authError: Locator;
  readonly submitButton: Locator;
  readonly logoutButton: Locator;

  constructor(page: Page) {
    this.page = page;
    this.emailInput = page.getByTestId('email-input');
    this.passwordInput = page.getByTestId('password-input');
    this.loginButton = page.getByTestId('login-button');
    this.submitButton = page.getByTestId('login-button'); // alias for compatibility
    this.errorMessage = page.getByTestId('error-message');
    this.forgotPasswordLink = page.getByText('Forgot password?');
    this.googleLoginButton = page.getByTestId('login-with-google');
    this.githubLoginButton = page.getByTestId('login-with-github');
    this.authError = page.getByTestId('auth-error');
    this.logoutButton = page.getByTestId('logout-button');
  }

  async goto() {
    await this.page.goto('/');
  }

  async login(email: string, password: string) {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.loginButton.click();
  }

  async expectError(message: string) {
    await this.errorMessage.waitFor({ state: 'visible' });
    await this.page.waitForFunction(
      (msg) => {
        const element = document.querySelector('[data-testid="error-message"]');
        return element?.textContent?.includes(msg);
      },
      message
    );
  }

  async isLoggedIn() {
    // Check if redirected to dashboard
    await this.page.waitForURL('**/dashboard/**', { timeout: 5000 });
    return this.page.url().includes('/dashboard');
  }

  async logout() {
    // Open user menu and click logout
    await this.page.getByTestId('user-menu-button').click();
    await this.page.getByTestId('user-menu-logout').click();
    
    // Wait for redirect to login page
    await this.page.waitForURL('**/login', { timeout: 5000 });
  }

  async loginWithGoogle() {
    await this.googleLoginButton.click();
  }

  async loginWithGitHub() {
    await this.githubLoginButton.click();
  }
}