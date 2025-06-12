import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { setupScenarioTest, createWorkspace, createOrganization, deployApplication } from '@/test-utils/scenario-helpers';
import { apiClient } from '@/lib/api-client';

// Import actual components for integration testing
import OrganizationsList from '@/components/organizations-list';
import { WorkspaceList } from '@/components/workspaces/workspace-list';
import { ProjectList } from '@/components/projects/project-list';
import { ApplicationList } from '@/components/applications/application-list';

jest.mock('next/navigation', () => ({
  useRouter: jest.fn(() => ({
    push: jest.fn(),
    replace: jest.fn(),
    back: jest.fn(),
  })),
  useParams: jest.fn(() => ({})),
  useSearchParams: jest.fn(() => ({
    get: jest.fn(),
  })),
}));

describe('Workspace Flow Integration Tests', () => {
  let user: any;

  beforeEach(() => {
    const setup = setupScenarioTest();
    user = setup.user;
  });

  it('should complete full workspace creation and resource deployment flow', async () => {
    // Step 1: Create Organization
    const mockOrganizations: any[] = [];
    const mockOnDelete = jest.fn();
    const mockOnUpdate = jest.fn();
    
    const { rerender } = render(
      <OrganizationsList 
        organizations={mockOrganizations}
        onDelete={mockOnDelete}
        onUpdate={mockOnUpdate}
      />
    );
    
    await createOrganization(user, 'Integration Test Org');
    
    // Verify organization was created
    await waitFor(() => {
      expect(apiClient.organizations.create).toHaveBeenCalledWith({
        name: 'Integration Test Org',
      });
    });
    
    // Step 2: Create Workspace
    const mockWorkspaces: any[] = [];
    const mockWorkspaceCreate = jest.fn();
    const mockWorkspaceDelete = jest.fn();
    
    rerender(
      <WorkspaceList 
        workspaces={mockWorkspaces}
        onCreateWorkspace={mockWorkspaceCreate}
        onDeleteWorkspace={mockWorkspaceDelete}
      />
    );
    
    await createWorkspace(user, 'org-test-1', 'Production Workspace', 'dedicated');
    
    // Verify workspace was created
    await waitFor(() => {
      expect(apiClient.workspaces.create).toHaveBeenCalledWith('org-test-1', {
        name: 'Production Workspace',
        plan: 'dedicated',
      });
    });
    
    // Step 3: Create Project
    const mockProjects: any[] = [];
    const mockProjectCreate = jest.fn();
    const mockProjectDelete = jest.fn();
    
    rerender(
      <ProjectList 
        projects={mockProjects}
        onCreate={mockProjectCreate}
        onDelete={mockProjectDelete}
      />
    );
    
    const createProjButton = screen.getByRole('button', { name: /create project/i });
    await user.click(createProjButton);
    
    const projNameInput = screen.getByLabelText(/project name/i);
    await user.type(projNameInput, 'Web Application');
    
    const submitProjButton = screen.getByRole('button', { name: /create project/i });
    await user.click(submitProjButton);
    
    await waitFor(() => {
      expect(apiClient.projects.create).toHaveBeenCalled();
    });
    
    // Step 4: Deploy Application
    rerender(<ApplicationList />);
    
    await deployApplication(user, 'org-test-1', 'ws-test-1', 'proj-test-1', {
      name: 'Frontend App',
      type: 'stateless',
      source_type: 'image',
      source_image: 'nginx:latest',
    });
    
    // Verify application was deployed
    await waitFor(() => {
      expect(apiClient.applications.create).toHaveBeenCalledWith(
        'org-test-1', 'ws-test-1', 'proj-test-1',
        expect.objectContaining({
          name: 'Frontend App',
          type: 'stateless',
        })
      );
    });
  });

  it('should handle errors gracefully during workspace creation', async () => {
    // Mock API failure
    (apiClient.workspaces.create as jest.Mock).mockRejectedValue(
      new Error('Insufficient quota')
    );
    
    render(<WorkspaceList />);
    
    // Try to create workspace
    const createButton = screen.getByRole('button', { name: /create workspace/i });
    await user.click(createButton);
    
    const nameInput = screen.getByLabelText(/workspace name/i);
    await user.type(nameInput, 'Failed Workspace');
    
    const submitButton = screen.getByRole('button', { name: /create/i });
    await user.click(submitButton);
    
    // Should show error message
    await waitFor(() => {
      expect(screen.getByText(/insufficient quota/i)).toBeInTheDocument();
    });
  });

  it('should update UI state correctly when resources are modified', async () => {
    // Setup initial state with applications
    (apiClient.applications.list as jest.Mock).mockResolvedValue({
      data: {
        applications: [
          {
            id: 'app-1',
            name: 'Test App',
            status: 'running',
            type: 'stateless',
          },
        ],
        total: 1,
      },
    });
    
    render(<ApplicationList />);
    
    await waitFor(() => {
      expect(screen.getByText('Test App')).toBeInTheDocument();
      expect(screen.getByText('running')).toBeInTheDocument();
    });
    
    // Update application status
    (apiClient.applications.updateStatus as jest.Mock).mockResolvedValue({
      data: {
        id: 'app-1',
        name: 'Test App',
        status: 'suspended',
        type: 'stateless',
      },
    });
    
    const statusButton = screen.getByTestId('status-app-1');
    await user.click(statusButton);
    
    // UI should update to show new status
    await waitFor(() => {
      expect(screen.getByText('suspended')).toBeInTheDocument();
    });
  });

  it('should handle concurrent operations correctly', async () => {
    render(<WorkspaceList />);
    
    // Start multiple create operations
    const promises = [];
    
    for (let i = 0; i < 3; i++) {
      promises.push(
        createWorkspace(user, 'org-test-1', `Workspace ${i}`, 'shared')
      );
    }
    
    // All operations should complete
    await Promise.all(promises);
    
    // Verify all API calls were made
    expect(apiClient.workspaces.create).toHaveBeenCalledTimes(3);
  });

  it('should maintain state consistency across navigation', async () => {
    const { rerender } = render(<OrganizationsList />);
    
    // Set up initial organizations
    (apiClient.organizations.list as jest.Mock).mockResolvedValue({
      organizations: [
        { id: 'org-1', name: 'Org 1' },
        { id: 'org-2', name: 'Org 2' },
      ],
      total: 2,
    });
    
    await waitFor(() => {
      expect(screen.getByText('Org 1')).toBeInTheDocument();
      expect(screen.getByText('Org 2')).toBeInTheDocument();
    });
    
    // Navigate to workspaces
    rerender(<WorkspaceList />);
    
    // Should maintain organization context
    await waitFor(() => {
      expect(apiClient.workspaces.list).toHaveBeenCalledWith(
        expect.any(String),
        expect.any(Object)
      );
    });
  });

  it('should handle real-time updates via websocket', async () => {
    render(<ApplicationList />);
    
    // Initial state
    (apiClient.applications.list as jest.Mock).mockResolvedValue({
      data: {
        applications: [
          {
            id: 'app-1',
            name: 'WebSocket Test App',
            status: 'pending',
          },
        ],
        total: 1,
      },
    });
    
    await waitFor(() => {
      expect(screen.getByText('WebSocket Test App')).toBeInTheDocument();
      expect(screen.getByText('pending')).toBeInTheDocument();
    });
    
    // Simulate websocket update
    // In real implementation, this would come from WebSocket
    const mockWebSocketUpdate = {
      type: 'application.status.changed',
      data: {
        id: 'app-1',
        status: 'running',
      },
    };
    
    // Update the mock to return new status
    (apiClient.applications.list as jest.Mock).mockResolvedValue({
      data: {
        applications: [
          {
            id: 'app-1',
            name: 'WebSocket Test App',
            status: 'running',
          },
        ],
        total: 1,
      },
    });
    
    // Trigger re-fetch (in real app, WebSocket would trigger this)
    const refreshButton = screen.getByRole('button', { name: /refresh/i });
    await user.click(refreshButton);
    
    // Should show updated status
    await waitFor(() => {
      expect(screen.getByText('running')).toBeInTheDocument();
    });
  });

  it('should validate resource quotas before creation', async () => {
    // Mock workspace with limited quota
    (apiClient.workspaces.get as jest.Mock).mockResolvedValue({
      id: 'ws-1',
      resource_limits: {
        cpu: '4',
        memory: '8Gi',
        storage: '100Gi',
      },
      resource_usage: {
        cpu: '3.5',
        memory: '7Gi',
        storage: '80Gi',
      },
    });
    
    render(<ApplicationList />);
    
    // Try to deploy resource-intensive application
    const deployButton = screen.getByRole('button', { name: /deploy application/i });
    await user.click(deployButton);
    
    // Set high resource requirements
    const cpuInput = screen.getByLabelText(/cpu request/i);
    await user.type(cpuInput, '2');
    
    const submitButton = screen.getByRole('button', { name: /deploy/i });
    await user.click(submitButton);
    
    // Should show quota exceeded error
    await waitFor(() => {
      expect(screen.getByText(/exceeds available cpu quota/i)).toBeInTheDocument();
    });
  });
});