# Test info

- Name: Billing Direct Test >> should demonstrate billing components working
- Location: /Users/hi/src/hexabase-kaas/ui/tests/billing-direct-test.spec.ts:5:7

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:3004/test-billing
Call log:
  - navigating to "http://localhost:3004/test-billing", waiting until "load"

    at /Users/hi/src/hexabase-kaas/ui/tests/billing-direct-test.spec.ts:7:16
```

# Test source

```ts
   1 | import { test, expect } from '@playwright/test';
   2 |
   3 | test.describe('Billing Direct Test', () => {
   4 |
   5 |   test('should demonstrate billing components working', async ({ page }) => {
   6 |     // Navigate directly to the test page
>  7 |     await page.goto('http://localhost:3004/test-billing');
     |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:3004/test-billing
   8 |     
   9 |     // Wait for the billing overview to load
  10 |     await page.waitForSelector('[data-testid="billing-overview"]', { timeout: 10000 });
  11 |     
  12 |     // Take a screenshot of the billing overview
  13 |     await page.screenshot({ 
  14 |       path: 'test-results/billing-overview-working.png',
  15 |       fullPage: true 
  16 |     });
  17 |     
  18 |     // Test the upgrade plan modal
  19 |     await page.click('[data-testid="upgrade-plan-button"]');
  20 |     await page.waitForSelector('[data-testid="subscription-plans-modal"]');
  21 |     
  22 |     // Take a screenshot of the plans modal
  23 |     await page.screenshot({ 
  24 |       path: 'test-results/subscription-plans-working.png',
  25 |       fullPage: true 
  26 |     });
  27 |     
  28 |     // Close the modal and test payment methods
  29 |     await page.keyboard.press('Escape');
  30 |     await page.click('[data-testid="payment-method-button"]');
  31 |     await page.waitForSelector('[data-testid="payment-methods-modal"]');
  32 |     
  33 |     // Take a screenshot of the payment modal
  34 |     await page.screenshot({ 
  35 |       path: 'test-results/payment-methods-working.png',
  36 |       fullPage: true 
  37 |     });
  38 |     
  39 |     console.log('✅ All Billing Components Working Successfully!');
  40 |     console.log('📸 Screenshots saved:');
  41 |     console.log('   - billing-overview-working.png');
  42 |     console.log('   - subscription-plans-working.png'); 
  43 |     console.log('   - payment-methods-working.png');
  44 |   });
  45 | });
```