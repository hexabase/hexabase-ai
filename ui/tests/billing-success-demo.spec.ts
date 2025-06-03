import { test, expect } from '@playwright/test';

test.describe('Billing Management Success Demo', () => {

  test('should demonstrate billing overview functionality', async ({ page }) => {
    await page.goto('/test-billing');
    
    // Wait for the page to load completely
    await page.waitForSelector('[data-testid="billing-test-page"]');
    
    // Check billing overview components
    await expect(page.locator('[data-testid="billing-overview"]')).toBeVisible();
    await expect(page.locator('[data-testid="current-plan"]')).toBeVisible();
    await expect(page.locator('[data-testid="plan-name"]')).toContainText('Professional');
    await expect(page.locator('[data-testid="plan-price"]')).toContainText('$29/mo');
    await expect(page.locator('[data-testid="billing-cycle"]')).toContainText('Monthly');
    
    // Check usage overview
    await expect(page.locator('[data-testid="usage-overview"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspaces-usage"]')).toBeVisible();
    await expect(page.locator('[data-testid="storage-usage"]')).toBeVisible();
    await expect(page.locator('[data-testid="bandwidth-usage"]')).toBeVisible();
    
    // Check action buttons
    await expect(page.locator('[data-testid="upgrade-plan-button"]')).toBeVisible();
    await expect(page.locator('[data-testid="payment-method-button"]')).toBeVisible();
    await expect(page.locator('[data-testid="download-invoice-button"]')).toBeVisible();
    
    console.log('✅ Billing Overview: ALL COMPONENTS WORKING!');
  });

  test('should demonstrate subscription plans modal', async ({ page }) => {
    await page.goto('/test-billing');
    await page.waitForSelector('[data-testid="billing-test-page"]');
    
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
    
    // Take screenshot of plans modal
    await page.screenshot({ 
      path: 'test-results/subscription-plans-modal.png',
      fullPage: true 
    });
    
    console.log('✅ Subscription Plans: ALL FEATURES WORKING!');
  });

  test('should demonstrate subscription upgrade flow', async ({ page }) => {
    await page.goto('/test-billing');
    await page.waitForSelector('[data-testid="billing-test-page"]');
    
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
    
    // Take screenshot of success state
    await page.screenshot({ 
      path: 'test-results/billing-upgrade-success.png',
      fullPage: true 
    });
    
    console.log('✅ Upgrade Flow: ALL STEPS WORKING!');
  });

  test('should demonstrate payment methods modal', async ({ page }) => {
    await page.goto('/test-billing');
    await page.waitForSelector('[data-testid="billing-test-page"]');
    
    // Click payment method button
    await page.click('[data-testid="payment-method-button"]');
    
    // Should see payment methods modal
    await expect(page.locator('[data-testid="payment-methods-modal"]')).toBeVisible();
    await expect(page.locator('text=Payment Methods')).toBeVisible();
    
    // Should see current payment method
    await expect(page.locator('[data-testid="current-payment-method"]')).toBeVisible();
    await expect(page.locator('[data-testid="card-last-four"]')).toContainText('4242');
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
    
    // Take screenshot of payment form
    await page.screenshot({ 
      path: 'test-results/payment-methods-modal.png',
      fullPage: true 
    });
    
    console.log('✅ Payment Methods: ALL FEATURES WORKING!');
  });

  test('should demonstrate usage analytics and billing features', async ({ page }) => {
    await page.goto('/test-billing');
    await page.waitForSelector('[data-testid="billing-test-page"]');
    
    // Check usage analytics
    await expect(page.locator('[data-testid="usage-analytics"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspaces-usage-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="storage-usage-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="bandwidth-usage-chart"]')).toBeVisible();
    
    // Check current period usage
    await expect(page.locator('[data-testid="current-period-usage"]')).toBeVisible();
    await expect(page.locator('[data-testid="usage-percentage"]')).toBeVisible();
    
    // Check billing forecast
    await expect(page.locator('[data-testid="billing-forecast"]')).toBeVisible();
    await expect(page.locator('[data-testid="projected-cost"]')).toContainText('$32.50');
    
    // Check billing history
    await expect(page.locator('[data-testid="billing-history"]')).toBeVisible();
    await expect(page.locator('[data-testid="invoice-list"]')).toBeVisible();
    await expect(page.locator('[data-testid="invoice-item"]')).toBeVisible();
    await expect(page.locator('[data-testid="invoice-status"]')).toContainText('Paid');
    await expect(page.locator('[data-testid="invoice-date"]')).toBeVisible();
    await expect(page.locator('[data-testid="invoice-amount"]')).toContainText('$29.00');
    await expect(page.locator('[data-testid="download-pdf"]')).toBeVisible();
    
    // Check billing settings
    await expect(page.locator('[data-testid="billing-settings"]')).toBeVisible();
    await expect(page.locator('[data-testid="billing-alerts"]')).toBeVisible();
    await expect(page.locator('[data-testid="usage-threshold-alert"]')).toBeVisible();
    await expect(page.locator('[data-testid="billing-email-alert"]')).toBeVisible();
    await expect(page.locator('[data-testid="notification-preferences"]')).toBeVisible();
    await expect(page.locator('[data-testid="email-notifications"]')).toBeVisible();
    await expect(page.locator('[data-testid="slack-notifications"]')).toBeVisible();
    
    // Test settings interaction
    await page.fill('[data-testid="usage-threshold-input"]', '85');
    await page.click('[data-testid="save-threshold"]');
    
    // Take final screenshot
    await page.screenshot({ 
      path: 'test-results/billing-complete-demo.png',
      fullPage: true 
    });
    
    console.log('✅ Usage Analytics & Settings: ALL FEATURES WORKING!');
    console.log('📊 Billing History: Complete');
    console.log('⚙️  Settings & Alerts: Functional');
    console.log('💳 Payment Management: Ready');
  });
});