import { test, expect } from '@playwright/test';

test.describe('Project Management', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication
    await page.addInitScript(() => {
      document.cookie = 'hexabase_access_token=test-token; path=/';
      document.cookie = 'hexabase_refresh_token=test-refresh; path=/';
      localStorage.setItem('hexabase_user', JSON.stringify({
        id: 'test-user',
        email: 'test@hexabase.com',
        name: 'Test User',
        provider: 'google'
      }));
    });
  });

  test('should display projects navigation in organization dashboard', async ({ page }) => {
    // Wait for organization dashboard to load
    await page.waitForSelector('[data-testid="organization-name"]');
    
    // Should see projects tab in navigation
    await expect(page.locator('[data-testid="projects-tab"]')).toBeVisible();
    await expect(page.locator('text=Projects')).toBeVisible();
    
    // Click on projects tab
    await page.click('[data-testid="projects-tab"]');
    
    // Should see projects section
    await expect(page.locator('[data-testid="projects-section"]')).toBeVisible();
  });

  test('should display empty state when no projects exist', async ({ page }) => {
    // Navigate to projects section
    await page.waitForSelector('[data-testid="organization-name"]');
    await page.click('[data-testid="projects-tab"]');
    
    // Should see empty state
    await expect(page.locator('[data-testid="projects-empty-state"]')).toBeVisible();
    await expect(page.locator('text=No projects yet')).toBeVisible();
    await expect(page.locator('text=Create your first project')).toBeVisible();
    
    // Should see create project button
    await expect(page.locator('[data-testid="create-project-button"]')).toBeVisible();
  });

  test('should open create project dialog', async ({ page }) => {
    // Navigate to projects and click create
    await page.waitForSelector('[data-testid="organization-name"]');
    await page.click('[data-testid="projects-tab"]');
    await page.click('[data-testid="create-project-button"]');
    
    // Should see create project modal
    await expect(page.locator('[data-testid="create-project-modal"]')).toBeVisible();
    await expect(page.locator('text=Create New Project')).toBeVisible();
    
    // Should see form fields
    await expect(page.locator('[data-testid="project-name-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="project-description-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspace-selection"]')).toBeVisible();
    await expect(page.locator('[data-testid="submit-project"]')).toBeVisible();
  });

  test('should create new project with valid data', async ({ page }) => {
    // Navigate to projects and open create dialog
    await page.waitForSelector('[data-testid="organization-name"]');
    await page.click('[data-testid="projects-tab"]');
    await page.click('[data-testid="create-project-button"]');
    
    // Fill out project form
    await page.fill('[data-testid="project-name-input"]', 'Test Project');
    await page.fill('[data-testid="project-description-input"]', 'A test project for development');
    await page.selectOption('[data-testid="workspace-selection"]', 'workspace-1');
    
    // Submit form
    await page.click('[data-testid="submit-project"]');
    
    // Should see success message
    await expect(page.locator('text=Project created successfully')).toBeVisible();
    
    // Modal should close
    await expect(page.locator('[data-testid="create-project-modal"]')).not.toBeVisible();
    
    // Should see project in list
    await expect(page.locator('[data-testid="project-card-test-project"]')).toBeVisible();
  });

  test('should validate project form inputs', async ({ page }) => {
    // Navigate to create project dialog
    await page.waitForSelector('[data-testid="organization-name"]');
    await page.click('[data-testid="projects-tab"]');
    await page.click('[data-testid="create-project-button"]');
    
    // Try to submit empty form
    await page.click('[data-testid="submit-project"]');
    
    // Should see validation errors
    await expect(page.locator('[data-testid="name-error"]')).toBeVisible();
    await expect(page.locator('text=Project name is required')).toBeVisible();
    
    // Fill invalid name (too short)
    await page.fill('[data-testid="project-name-input"]', 'ab');
    await page.click('[data-testid="submit-project"]');
    
    // Should see length validation
    await expect(page.locator('text=Project name must be at least 3 characters')).toBeVisible();
    
    // Fill valid name but no workspace
    await page.fill('[data-testid="project-name-input"]', 'Valid Project Name');
    await page.click('[data-testid="submit-project"]');
    
    // Should see workspace validation
    await expect(page.locator('text=Please select a workspace')).toBeVisible();
  });

  test('should display project list with project cards', async ({ page }) => {
    // Navigate to projects section (assuming projects exist)
    await page.waitForSelector('[data-testid="organization-name"]');
    await page.click('[data-testid="projects-tab"]');
    
    // Should see projects list
    await expect(page.locator('[data-testid="projects-list"]')).toBeVisible();
    
    // Should see project cards
    await expect(page.locator('[data-testid="project-card"]').first()).toBeVisible();
    
    // Project cards should have required elements
    await expect(page.locator('[data-testid="project-name"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="project-description"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="project-workspace"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="project-status"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="project-namespace-count"]').first()).toBeVisible();
  });

  test('should navigate to project detail page', async ({ page }) => {
    // Navigate to projects section
    await page.waitForSelector('[data-testid="organization-name"]');
    await page.click('[data-testid="projects-tab"]');
    
    // Click on first project
    await page.click('[data-testid="project-card"]');
    
    // Should navigate to project detail page
    await expect(page).toHaveURL(/\/projects\/[^\/]+$/);
    await expect(page.locator('[data-testid="project-detail-page"]')).toBeVisible();
  });

  test('should display project detail page with namespace management', async ({ page }) => {
    // Go directly to project detail page
    await page.goto('/dashboard/organizations/org1/projects/project-123');
    
    // Should see project detail page
    await expect(page.locator('[data-testid="project-detail-page"]')).toBeVisible();
    await expect(page.locator('[data-testid="project-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="project-description"]')).toBeVisible();
    
    // Should see namespace management section
    await expect(page.locator('[data-testid="namespaces-section"]')).toBeVisible();
    await expect(page.locator('text=Namespaces')).toBeVisible();
    
    // Should see create namespace button
    await expect(page.locator('[data-testid="create-namespace-button"]')).toBeVisible();
  });

  test('should create namespace in project', async ({ page }) => {
    // Go to project detail page
    await page.goto('/dashboard/organizations/org1/projects/project-123');
    await page.waitForSelector('[data-testid="project-detail-page"]');
    
    // Click create namespace
    await page.click('[data-testid="create-namespace-button"]');
    
    // Should see create namespace modal
    await expect(page.locator('[data-testid="create-namespace-modal"]')).toBeVisible();
    await expect(page.locator('text=Create Namespace')).toBeVisible();
    
    // Fill namespace form
    await page.fill('[data-testid="namespace-name-input"]', 'development');
    await page.fill('[data-testid="namespace-description-input"]', 'Development environment');
    
    // Submit form
    await page.click('[data-testid="submit-namespace"]');
    
    // Should see success message
    await expect(page.locator('text=Namespace created successfully')).toBeVisible();
    
    // Should see namespace in list
    await expect(page.locator('[data-testid="namespace-card-development"]')).toBeVisible();
  });

  test('should display namespace resource quotas and usage', async ({ page }) => {
    // Go to project with namespaces
    await page.goto('/dashboard/organizations/org1/projects/project-123');
    await page.waitForSelector('[data-testid="project-detail-page"]');
    
    // Should see namespace cards with resource info
    const namespaceCard = page.locator('[data-testid="namespace-card"]').first();
    await expect(namespaceCard.locator('[data-testid="namespace-cpu-usage"]')).toBeVisible();
    await expect(namespaceCard.locator('[data-testid="namespace-memory-usage"]')).toBeVisible();
    await expect(namespaceCard.locator('[data-testid="namespace-pod-count"]')).toBeVisible();
    await expect(namespaceCard.locator('[data-testid="namespace-status"]')).toBeVisible();
  });

  test('should handle namespace actions (edit, delete)', async ({ page }) => {
    // Go to project with namespaces
    await page.goto('/dashboard/organizations/org1/projects/project-123');
    await page.waitForSelector('[data-testid="project-detail-page"]');
    
    // Should see namespace action menu
    await page.click('[data-testid="namespace-actions-menu"]');
    await expect(page.locator('[data-testid="edit-namespace"]')).toBeVisible();
    await expect(page.locator('[data-testid="delete-namespace"]')).toBeVisible();
    
    // Test edit namespace
    await page.click('[data-testid="edit-namespace"]');
    await expect(page.locator('[data-testid="edit-namespace-modal"]')).toBeVisible();
    
    // Close edit modal and test delete
    await page.click('[data-testid="cancel-edit"]');
    await page.click('[data-testid="namespace-actions-menu"]');
    await page.click('[data-testid="delete-namespace"]');
    
    // Should see confirmation dialog
    await expect(page.locator('[data-testid="delete-confirmation"]')).toBeVisible();
    await expect(page.locator('text=Are you sure you want to delete')).toBeVisible();
  });

  test('should display project statistics and metrics', async ({ page }) => {
    // Go to project detail page
    await page.goto('/dashboard/organizations/org1/projects/project-123');
    await page.waitForSelector('[data-testid="project-detail-page"]');
    
    // Should see project statistics cards
    await expect(page.locator('[data-testid="total-namespaces-stat"]')).toBeVisible();
    await expect(page.locator('[data-testid="total-pods-stat"]')).toBeVisible();
    await expect(page.locator('[data-testid="cpu-usage-stat"]')).toBeVisible();
    await expect(page.locator('[data-testid="memory-usage-stat"]')).toBeVisible();
    
    // Should see resource usage charts
    await expect(page.locator('[data-testid="resource-usage-chart"]')).toBeVisible();
  });

  test('should filter and search projects', async ({ page }) => {
    // Navigate to projects section
    await page.waitForSelector('[data-testid="organization-name"]');
    await page.click('[data-testid="projects-tab"]');
    
    // Should see search and filter controls
    await expect(page.locator('[data-testid="project-search-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="project-status-filter"]')).toBeVisible();
    await expect(page.locator('[data-testid="workspace-filter"]')).toBeVisible();
    
    // Test search functionality
    await page.fill('[data-testid="project-search-input"]', 'test');
    await expect(page.locator('[data-testid="project-card"]')).toHaveCount(1);
    
    // Test status filter
    await page.selectOption('[data-testid="project-status-filter"]', 'active');
    await expect(page.locator('[data-testid="project-card"][data-status="active"]')).toBeVisible();
  });

  test('should handle project settings and configuration', async ({ page }) => {
    // Go to project detail page
    await page.goto('/dashboard/organizations/org1/projects/project-123');
    await page.waitForSelector('[data-testid="project-detail-page"]');
    
    // Should see project settings button
    await expect(page.locator('[data-testid="project-settings-button"]')).toBeVisible();
    
    // Click project settings
    await page.click('[data-testid="project-settings-button"]');
    
    // Should see settings modal
    await expect(page.locator('[data-testid="project-settings-modal"]')).toBeVisible();
    await expect(page.locator('text=Project Settings')).toBeVisible();
    
    // Should see settings tabs
    await expect(page.locator('[data-testid="general-settings-tab"]')).toBeVisible();
    await expect(page.locator('[data-testid="resource-limits-tab"]')).toBeVisible();
    await expect(page.locator('[data-testid="permissions-tab"]')).toBeVisible();
  });
});