import { test, expect } from '@playwright/test';

test.describe('Billing Direct Test', () => {

  test('should demonstrate billing components working', async ({ page }) => {
    // Navigate directly to the test page
    await page.goto('http://localhost:3004/test-billing');
    
    // Wait for the billing overview to load
    await page.waitForSelector('[data-testid="billing-overview"]', { timeout: 10000 });
    
    // Take a screenshot of the billing overview
    await page.screenshot({ 
      path: 'test-results/billing-overview-working.png',
      fullPage: true 
    });
    
    // Test the upgrade plan modal
    await page.click('[data-testid="upgrade-plan-button"]');
    await page.waitForSelector('[data-testid="subscription-plans-modal"]');
    
    // Take a screenshot of the plans modal
    await page.screenshot({ 
      path: 'test-results/subscription-plans-working.png',
      fullPage: true 
    });
    
    // Close the modal and test payment methods
    await page.keyboard.press('Escape');
    await page.click('[data-testid="payment-method-button"]');
    await page.waitForSelector('[data-testid="payment-methods-modal"]');
    
    // Take a screenshot of the payment modal
    await page.screenshot({ 
      path: 'test-results/payment-methods-working.png',
      fullPage: true 
    });
    
    console.log('✅ All Billing Components Working Successfully!');
    console.log('📸 Screenshots saved:');
    console.log('   - billing-overview-working.png');
    console.log('   - subscription-plans-working.png'); 
    console.log('   - payment-methods-working.png');
  });
});