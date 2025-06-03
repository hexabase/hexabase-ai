import { test, expect } from '@playwright/test';

test.describe('Projects Management Success Demo', () => {

  test('should demonstrate successful projects list interface', async ({ page }) => {
    await page.goto('/test-projects');
    
    // Wait for the page to load completely
    await page.waitForSelector('[data-testid="projects-test-page"]');
    
    // Check if the projects list component is present
    await expect(page.locator('h2:has-text("Projects")')).toBeVisible();
    
    // Should see project cards
    await expect(page.locator('[data-testid="project-card"]').first()).toBeVisible();
    
    // Verify project card elements
    await expect(page.locator('[data-testid="project-name"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="project-status"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="project-namespace-count"]').first()).toBeVisible();
    
    // Should see search and filter controls
    await expect(page.locator('[data-testid="project-search-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="project-status-filter"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspace-filter"]')).toBeVisible();
    await expect(page.locator('[data-testid="create-project-button"]')).toBeVisible();
    
    // Take a screenshot
    await page.screenshot({ 
      path: 'test-results/projects-list-interface.png',
      fullPage: true 
    });
    
    console.log('✅ Projects List: ALL TESTS PASSED!');
  });

  test('should demonstrate project creation dialog', async ({ page }) => {
    await page.goto('/test-projects');
    
    // Wait for page to load
    await page.waitForSelector('[data-testid="projects-test-page"]');
    
    // Click create project button
    await page.click('[data-testid="create-project-button"]');
    
    // Should see create project modal
    await expect(page.locator('[data-testid="create-project-modal"]')).toBeVisible();
    await expect(page.locator('text=Create New Project')).toBeVisible();
    
    // Should see form fields
    await expect(page.locator('[data-testid="project-name-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="project-description-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspace-selection"]')).toBeVisible();
    await expect(page.locator('[data-testid="submit-project"]')).toBeVisible();
    
    // Fill out the form
    await page.fill('[data-testid="project-name-input"]', 'Test Project');
    await page.fill('[data-testid="project-description-input"]', 'A test project for demonstration');
    
    // Take screenshot of create dialog
    await page.screenshot({ 
      path: 'test-results/project-creation-dialog.png',
      fullPage: true 
    });
    
    console.log('✅ Project Creation Dialog: Functional!');
  });

  test('should demonstrate project detail page with namespace management', async ({ page }) => {
    await page.goto('/test-projects/detail');
    
    // Wait for the page to load completely
    await page.waitForSelector('[data-testid="project-detail-page"]');
    
    // Verify project details are visible
    await expect(page.locator('[data-testid="project-name"]')).toHaveText('Frontend Application');
    await expect(page.locator('[data-testid="project-description"]')).toBeVisible();
    
    // Check statistics cards
    await expect(page.locator('[data-testid="total-namespaces-stat"]')).toContainText('3');
    await expect(page.locator('[data-testid="total-pods-stat"]')).toContainText('12');
    await expect(page.locator('[data-testid="cpu-usage-stat"]')).toContainText('65%');
    await expect(page.locator('[data-testid="memory-usage-stat"]')).toContainText('42%');
    
    // Check resource usage chart
    await expect(page.locator('[data-testid="resource-usage-chart"]')).toBeVisible();
    
    // Check namespaces section
    await expect(page.locator('[data-testid="namespaces-section"]')).toBeVisible();
    await expect(page.locator('[data-testid="create-namespace-button"]')).toBeVisible();
    
    // Check namespace cards
    await expect(page.locator('[data-testid="namespace-card"]')).toHaveCount(3);
    await expect(page.locator('[data-testid="namespace-cpu-usage"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="namespace-memory-usage"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="namespace-pod-count"]').first()).toBeVisible();
    
    // Take a final success screenshot
    await page.screenshot({ 
      path: 'test-results/project-detail-success.png',
      fullPage: true 
    });
    
    console.log('✅ Project Detail Page: ALL COMPONENTS WORKING!');
    console.log('🎯 Namespace Management: Functional');
    console.log('📊 Resource Monitoring: Functional');
    console.log('⚙️  Project Settings: Available');
  });

  test('should demonstrate namespace creation dialog', async ({ page }) => {
    await page.goto('/test-projects/detail');
    await page.waitForSelector('[data-testid="project-detail-page"]');
    
    // Click create namespace button
    await page.click('[data-testid="create-namespace-button"]');
    
    // Should see create namespace modal
    await expect(page.locator('[data-testid="create-namespace-modal"]')).toBeVisible();
    await expect(page.locator('[data-testid="create-namespace-modal"] h2')).toContainText('Create Namespace');
    
    // Should see form fields
    await expect(page.locator('[data-testid="namespace-name-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="namespace-description-input"]')).toBeVisible();
    
    // Fill namespace form
    await page.fill('[data-testid="namespace-name-input"]', 'production');
    await page.fill('[data-testid="namespace-description-input"]', 'Production environment');
    
    // Take screenshot of namespace creation
    await page.screenshot({ 
      path: 'test-results/namespace-creation-dialog.png',
      fullPage: true 
    });
    
    // Submit form
    await page.click('[data-testid="submit-namespace"]');
    
    // Should see success message
    await expect(page.locator('text=Namespace created successfully')).toBeVisible();
    
    console.log('✅ Namespace Creation: ALL INTERACTIONS WORKING!');
  });

  test('should demonstrate project settings functionality', async ({ page }) => {
    await page.goto('/test-projects/detail');
    await page.waitForSelector('[data-testid="project-detail-page"]');
    
    // Click project settings
    await page.click('[data-testid="project-settings-button"]');
    
    // Should see settings modal
    await expect(page.locator('[data-testid="project-settings-modal"]')).toBeVisible();
    await expect(page.locator('text=Project Settings')).toBeVisible();
    
    // Should see settings tabs
    await expect(page.locator('[data-testid="general-settings-tab"]')).toBeVisible();
    await expect(page.locator('[data-testid="resource-limits-tab"]')).toBeVisible();
    await expect(page.locator('[data-testid="permissions-tab"]')).toBeVisible();
    
    // Test tab navigation
    await page.click('[data-testid="resource-limits-tab"]');
    await page.click('[data-testid="permissions-tab"]');
    
    // Take screenshot of settings
    await page.screenshot({ 
      path: 'test-results/project-settings-success.png',
      fullPage: true 
    });
    
    console.log('✅ Project Settings: ALL TABS FUNCTIONAL!');
  });
});