import { Page } from '@playwright/test';
import { 
  StartupScenario,
  EnterpriseScenario,
  ScenarioData,
} from '../fixtures/scenarios';
import { setupMockAPI } from './mock-api';

export type ScenarioType = 'startup' | 'enterprise' | 'development' | 'disaster' | 'migration' | 'performance';

/**
 * Test Data Manager
 * Manages test data generation and API mocking for E2E tests
 */
export class TestDataManager {
  private scenario: ScenarioData | null = null;
  private scenarioType: ScenarioType | null = null;
  
  constructor(private page: Page) {}
  
  /**
   * Load a predefined scenario
   */
  async loadScenario(type: ScenarioType, seed?: number): Promise<ScenarioData> {
    this.scenarioType = type;
    
    switch (type) {
      case 'startup':
        this.scenario = new StartupScenario(seed).generate();
        break;
        
      case 'enterprise':
        this.scenario = new EnterpriseScenario(seed).generate();
        break;
        
      // Add other scenarios as they're implemented
      default:
        throw new Error(`Scenario type '${type}' not implemented yet`);
    }
    
    // Set up mock API with scenario data
    await this.setupMockAPI();
    
    return this.scenario;
  }
  
  /**
   * Get current scenario data
   */
  getScenario(): ScenarioData {
    if (!this.scenario) {
      throw new Error('No scenario loaded. Call loadScenario() first.');
    }
    return this.scenario;
  }
  
  /**
   * Get specific entities from scenario
   */
  getOrganizations() {
    return this.getScenario().organizations;
  }
  
  getWorkspaces() {
    return this.getScenario().workspaces;
  }
  
  getProjects() {
    return this.getScenario().projects;
  }
  
  getApplications() {
    return this.getScenario().applications;
  }
  
  getUsers() {
    return this.getScenario().users;
  }
  
  /**
   * Find entities by criteria
   */
  findWorkspaceByName(name: string) {
    return this.getWorkspaces().find(w => w.name === name);
  }
  
  findProjectByName(name: string) {
    return this.getProjects().find(p => p.name === name);
  }
  
  findApplicationByName(name: string) {
    return this.getApplications().find(a => a.name === name);
  }
  
  findUserByEmail(email: string) {
    return this.getUsers().find(u => u.email === email);
  }
  
  /**
   * Get test credentials
   */
  getAdminCredentials() {
    const admin = this.getUsers().find(u => u.role === 'admin' || u.role === 'owner');
    if (!admin) {
      throw new Error('No admin user found in scenario');
    }
    return {
      email: admin.email,
      password: 'Test123!', // Default test password
    };
  }
  
  getDeveloperCredentials() {
    const developer = this.getUsers().find(u => u.role === 'developer');
    if (!developer) {
      throw new Error('No developer user found in scenario');
    }
    return {
      email: developer.email,
      password: 'Test123!',
    };
  }
  
  getViewerCredentials() {
    const viewer = this.getUsers().find(u => u.role === 'viewer');
    if (!viewer) {
      throw new Error('No viewer user found in scenario');
    }
    return {
      email: viewer.email,
      password: 'Test123!',
    };
  }
  
  /**
   * Set up mock API routes with scenario data
   */
  private async setupMockAPI() {
    if (!this.scenario) return;
    
    // Basic entity endpoints
    await this.page.route('**/api/organizations', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          organizations: this.scenario!.organizations,
          total: this.scenario!.organizations.length,
        }),
      });
    });
    
    await this.page.route('**/api/workspaces', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          workspaces: this.scenario!.workspaces,
          total: this.scenario!.workspaces.length,
        }),
      });
    });
    
    await this.page.route('**/api/projects', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          projects: this.scenario!.projects,
          total: this.scenario!.projects.length,
        }),
      });
    });
    
    await this.page.route('**/api/applications', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          applications: this.scenario!.applications,
          total: this.scenario!.applications.length,
        }),
      });
    });
    
    // Workspace-specific routes
    await this.page.route('**/api/organizations/*/workspaces/*', async (route) => {
      const url = new URL(route.request().url());
      const parts = url.pathname.split('/');
      const workspaceId = parts[parts.length - 1];
      
      const workspace = this.scenario!.workspaces.find(w => w.id === workspaceId);
      if (workspace) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(workspace),
        });
      } else {
        await route.fulfill({ status: 404 });
      }
    });
    
    // Project-specific applications
    await this.page.route('**/api/organizations/*/workspaces/*/projects/*/applications', async (route) => {
      const url = new URL(route.request().url());
      const parts = url.pathname.split('/');
      const projectId = parts[parts.indexOf('projects') + 1];
      
      const apps = this.scenario!.applications.filter(a => a.projectId === projectId);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          applications: apps,
          total: apps.length,
        }),
      });
    });
    
    // Metrics endpoints
    await this.page.route('**/api/metrics/**', async (route) => {
      const metrics = this.scenario!.monitoring.metrics[0] || {};
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(metrics),
      });
    });
    
    // Alerts endpoint
    await this.page.route('**/api/alerts', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          alerts: this.scenario!.monitoring.alerts,
          total: this.scenario!.monitoring.alerts.length,
        }),
      });
    });
  }
  
  /**
   * Create custom mock responses
   */
  async mockAPIResponse(pattern: string | RegExp, response: any) {
    await this.page.route(pattern, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(response),
      });
    });
  }
  
  /**
   * Simulate API errors
   */
  async simulateAPIError(pattern: string | RegExp, status: number = 500, message: string = 'Internal Server Error') {
    await this.page.route(pattern, async (route) => {
      await route.fulfill({
        status,
        contentType: 'application/json',
        body: JSON.stringify({
          error: message,
          code: status,
        }),
      });
    });
  }
  
  /**
   * Simulate network delays
   */
  async simulateNetworkDelay(pattern: string | RegExp, delayMs: number) {
    await this.page.route(pattern, async (route) => {
      await this.page.waitForTimeout(delayMs);
      await route.continue();
    });
  }
}