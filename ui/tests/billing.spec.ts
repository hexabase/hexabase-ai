import { test, expect } from '@playwright/test';

test.describe('Billing & Subscription Management', () => {
  
  test('should display billing overview with current subscription', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/billing');
    
    // Wait for billing page to load
    await page.waitForSelector('[data-testid="billing-overview"]');
    
    // Should see current subscription details
    await expect(page.locator('[data-testid="current-plan"]')).toBeVisible();
    await expect(page.locator('[data-testid="plan-name"]')).toContainText('Professional');
    await expect(page.locator('[data-testid="plan-price"]')).toBeVisible();
    await expect(page.locator('[data-testid="billing-cycle"]')).toContainText('Monthly');
    
    // Should see usage metrics
    await expect(page.locator('[data-testid="usage-overview"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspaces-usage"]')).toBeVisible();
    await expect(page.locator('[data-testid="storage-usage"]')).toBeVisible();
    await expect(page.locator('[data-testid="bandwidth-usage"]')).toBeVisible();
    
    // Should see billing actions
    await expect(page.locator('[data-testid="upgrade-plan-button"]')).toBeVisible();
    await expect(page.locator('[data-testid="download-invoice-button"]')).toBeVisible();
    await expect(page.locator('[data-testid="payment-method-button"]')).toBeVisible();
  });

  test('should show subscription plans comparison', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/billing');
    await page.waitForSelector('[data-testid="billing-overview"]');
    
    // Click upgrade plan button
    await page.click('[data-testid="upgrade-plan-button"]');
    
    // Should see plans modal
    await expect(page.locator('[data-testid="subscription-plans-modal"]')).toBeVisible();
    await expect(page.locator('text=Choose Your Plan')).toBeVisible();
    
    // Should see all plan options
    await expect(page.locator('[data-testid="starter-plan"]')).toBeVisible();
    await expect(page.locator('[data-testid="professional-plan"]')).toBeVisible();
    await expect(page.locator('[data-testid="enterprise-plan"]')).toBeVisible();
    
    // Should see plan features
    await expect(page.locator('[data-testid="plan-workspaces"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="plan-storage"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="plan-support"]').first()).toBeVisible();
    
    // Should see pricing toggle
    await expect(page.locator('[data-testid="billing-toggle"]')).toBeVisible();
    
    // Test monthly/yearly toggle
    await page.click('[data-testid="yearly-billing"]');
    await expect(page.locator('[data-testid="yearly-discount"]')).toBeVisible();
  });

  test('should manage payment methods', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/billing');
    await page.waitForSelector('[data-testid="billing-overview"]');
    
    // Click payment method button
    await page.click('[data-testid="payment-method-button"]');
    
    // Should see payment methods modal
    await expect(page.locator('[data-testid="payment-methods-modal"]')).toBeVisible();
    await expect(page.locator('text=Payment Methods')).toBeVisible();
    
    // Should see current payment method
    await expect(page.locator('[data-testid="current-payment-method"]')).toBeVisible();
    await expect(page.locator('[data-testid="card-last-four"]')).toBeVisible();
    await expect(page.locator('[data-testid="card-expiry"]')).toBeVisible();
    
    // Should see add payment method button
    await expect(page.locator('[data-testid="add-payment-method"]')).toBeVisible();
    
    // Click add payment method
    await page.click('[data-testid="add-payment-method"]');
    
    // Should see payment form
    await expect(page.locator('[data-testid="payment-form"]')).toBeVisible();
    await expect(page.locator('[data-testid="card-number-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="expiry-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="cvc-input"]')).toBeVisible();
  });

  test('should display billing history and invoices', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/billing/history');
    
    // Wait for billing history page
    await page.waitForSelector('[data-testid="billing-history"]');
    
    // Should see invoice list
    await expect(page.locator('[data-testid="invoice-list"]')).toBeVisible();
    await expect(page.locator('[data-testid="invoice-item"]').first()).toBeVisible();
    
    // Should see invoice details
    await expect(page.locator('[data-testid="invoice-date"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="invoice-amount"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="invoice-status"]').first()).toBeVisible();
    
    // Should see download buttons
    await expect(page.locator('[data-testid="download-pdf"]').first()).toBeVisible();
    
    // Should see pagination
    await expect(page.locator('[data-testid="invoice-pagination"]')).toBeVisible();
    
    // Test filtering
    await page.selectOption('[data-testid="invoice-filter"]', 'paid');
    await expect(page.locator('[data-testid="invoice-status"]').first()).toContainText('Paid');
  });

  test('should handle subscription upgrade flow', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/billing');
    await page.waitForSelector('[data-testid="billing-overview"]');
    
    // Start upgrade process
    await page.click('[data-testid="upgrade-plan-button"]');
    await page.waitForSelector('[data-testid="subscription-plans-modal"]');
    
    // Select enterprise plan
    await page.click('[data-testid="select-enterprise-plan"]');
    
    // Should see confirmation modal
    await expect(page.locator('[data-testid="upgrade-confirmation"]')).toBeVisible();
    await expect(page.locator('text=Confirm Upgrade')).toBeVisible();
    
    // Should see upgrade details
    await expect(page.locator('[data-testid="upgrade-from"]')).toContainText('Professional');
    await expect(page.locator('[data-testid="upgrade-to"]')).toContainText('Enterprise');
    await expect(page.locator('[data-testid="prorated-amount"]')).toBeVisible();
    
    // Should see payment method
    await expect(page.locator('[data-testid="payment-method-summary"]')).toBeVisible();
    
    // Confirm upgrade
    await page.click('[data-testid="confirm-upgrade"]');
    
    // Should see success message
    await expect(page.locator('[data-testid="upgrade-success"]')).toBeVisible();
    await expect(page.locator('text=Subscription upgraded successfully')).toBeVisible();
  });

  test('should display usage analytics and billing forecasts', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/billing/usage');
    
    // Wait for usage analytics page
    await page.waitForSelector('[data-testid="usage-analytics"]');
    
    // Should see usage charts
    await expect(page.locator('[data-testid="workspaces-usage-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="storage-usage-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="bandwidth-usage-chart"]')).toBeVisible();
    
    // Should see current period usage
    await expect(page.locator('[data-testid="current-period-usage"]')).toBeVisible();
    await expect(page.locator('[data-testid="usage-percentage"]')).toBeVisible();
    
    // Should see billing forecast
    await expect(page.locator('[data-testid="billing-forecast"]')).toBeVisible();
    await expect(page.locator('[data-testid="projected-cost"]')).toBeVisible();
    
    // Should see usage alerts
    await expect(page.locator('[data-testid="usage-alerts"]')).toBeVisible();
    
    // Test time period filter
    await page.selectOption('[data-testid="period-filter"]', '3months');
    await page.waitForTimeout(1000); // Wait for chart update
    
    // Take screenshot of usage analytics
    await page.screenshot({ 
      path: 'test-results/billing-usage-analytics.png',
      fullPage: true 
    });
  });

  test('should manage billing alerts and notifications', async ({ page }) => {
    await page.goto('/dashboard/organizations/org1/billing/settings');
    
    // Wait for billing settings page
    await page.waitForSelector('[data-testid="billing-settings"]');
    
    // Should see billing alerts section
    await expect(page.locator('[data-testid="billing-alerts"]')).toBeVisible();
    await expect(page.locator('[data-testid="usage-threshold-alert"]')).toBeVisible();
    await expect(page.locator('[data-testid="billing-email-alert"]')).toBeVisible();
    
    // Should see notification preferences
    await expect(page.locator('[data-testid="notification-preferences"]')).toBeVisible();
    await expect(page.locator('[data-testid="email-notifications"]')).toBeVisible();
    await expect(page.locator('[data-testid="slack-notifications"]')).toBeVisible();
    
    // Test setting usage alert threshold
    await page.fill('[data-testid="usage-threshold-input"]', '80');
    await page.click('[data-testid="save-threshold"]');
    
    // Should see success message
    await expect(page.locator('text=Alert threshold updated')).toBeVisible();
    
    // Take screenshot of billing settings
    await page.screenshot({ 
      path: 'test-results/billing-settings.png',
      fullPage: true 
    });
  });
});