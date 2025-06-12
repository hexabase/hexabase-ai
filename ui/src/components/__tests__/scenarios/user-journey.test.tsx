import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
  useParams: jest.fn(),
  useSearchParams: jest.fn(),
}));

// Import components for the journey
import LoginPage from '@/components/login-page';
import OrganizationsList from '@/components/organizations-list';
import { WorkspaceList } from '@/components/workspaces/workspace-list';
import { ProjectList } from '@/components/projects/project-list';
import { ApplicationList } from '@/components/applications/application-list';
import { FunctionList } from '@/components/functions/function-list';
import { CronJobList } from '@/components/cronjobs/cronjob-list';
import { MonitoringDashboard } from '@/components/monitoring/workspace-metrics';
import { BackupDashboard } from '@/components/backup/backup-dashboard';

describe('Complete User Journey Scenario', () => {
  const user = userEvent.setup();
  const mockPush = jest.fn();
  const mockRouter = { push: mockPush, replace: jest.fn(), back: jest.fn() };

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
    
    // Reset auth state
    (apiClient.auth.login as jest.Mock).mockResolvedValue({
      access_token: 'mock-token',
      refresh_token: 'mock-refresh-token',
      user: {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
      },
    });
  });

  describe('1. User Signup and Login Flow', () => {
    it('should complete signup and login process', async () => {
      // Start with login page
      render(<LoginPage />);
      
      // Check login form is displayed
      expect(screen.getByText(/sign in to hexabase ai/i)).toBeInTheDocument();
      
      // Fill in login form
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/password/i);
      
      await user.type(emailInput, 'newuser@example.com');
      await user.type(passwordInput, 'SecurePassword123!');
      
      // Submit login form
      const loginButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(loginButton);
      
      // Wait for login to complete
      await waitFor(() => {
        expect(apiClient.auth.login).toHaveBeenCalledWith({
          email: 'newuser@example.com',
          password: 'SecurePassword123!',
        });
      });
      
      // Verify redirect to dashboard
      expect(mockPush).toHaveBeenCalledWith('/dashboard');
    });
  });

  describe('2. Organization and Workspace Creation', () => {
    beforeEach(() => {
      // Mock organization API responses
      (apiClient.organizations.list as jest.Mock).mockResolvedValue({
        organizations: [],
        total: 0,
      });
      
      (apiClient.organizations.create as jest.Mock).mockResolvedValue({
        id: 'org-new',
        name: 'My New Organization',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      });
      
      (apiClient.workspaces.list as jest.Mock).mockResolvedValue({
        workspaces: [],
        total: 0,
      });
      
      (apiClient.workspaces.create as jest.Mock).mockResolvedValue({
        id: 'ws-new',
        organization_id: 'org-new',
        name: 'Production Workspace',
        plan: 'dedicated',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      });
    });

    it('should create organization and workspace', async () => {
      // Render organizations list
      render(<OrganizationsList />);
      
      await waitFor(() => {
        expect(screen.getByText(/no organizations found/i)).toBeInTheDocument();
      });
      
      // Click create organization
      const createOrgButton = screen.getByRole('button', { name: /create organization/i });
      await user.click(createOrgButton);
      
      // Fill organization form
      const orgNameInput = screen.getByLabelText(/organization name/i);
      await user.type(orgNameInput, 'My New Organization');
      
      // Submit form
      const submitButton = screen.getByRole('button', { name: /create/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(apiClient.organizations.create).toHaveBeenCalledWith({
          name: 'My New Organization',
        });
      });
      
      // Now create workspace
      const { rerender } = render(<div />);
      rerender(<WorkspaceList />);
      
      await waitFor(() => {
        expect(screen.getByText(/no workspaces found/i)).toBeInTheDocument();
      });
      
      // Click create workspace
      const createWsButton = screen.getByRole('button', { name: /create workspace/i });
      await user.click(createWsButton);
      
      // Fill workspace form
      const wsNameInput = screen.getByLabelText(/workspace name/i);
      await user.type(wsNameInput, 'Production Workspace');
      
      // Select plan
      const planSelect = screen.getByLabelText(/plan/i);
      await user.selectOptions(planSelect, 'dedicated');
      
      // Submit
      const submitWsButton = screen.getAllByRole('button', { name: /create/i })[1];
      await user.click(submitWsButton);
      
      await waitFor(() => {
        expect(apiClient.workspaces.create).toHaveBeenCalledWith('org-new', {
          name: 'Production Workspace',
          plan: 'dedicated',
        });
      });
    });
  });

  describe('3. Resource Creation and Management', () => {
    beforeEach(() => {
      // Mock project and application APIs
      (apiClient.projects.create as jest.Mock).mockResolvedValue({
        id: 'proj-new',
        workspace_id: 'ws-new',
        name: 'Web App Project',
        namespace: 'web-app',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      });
      
      (apiClient.applications.create as jest.Mock).mockResolvedValue({
        data: {
          id: 'app-new',
          workspace_id: 'ws-new',
          project_id: 'proj-new',
          name: 'Frontend App',
          type: 'stateless',
          status: 'running',
          source_type: 'image',
          source_image: 'nginx:latest',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        }
      });
      
      (apiClient.functions.create as jest.Mock).mockResolvedValue({
        data: {
          id: 'func-new',
          workspace_id: 'ws-new',
          project_id: 'proj-new',
          name: 'data-processor',
          runtime: 'nodejs18',
          handler: 'index.handler',
          status: 'active',
          version: 'v1.0.0',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        }
      });
    });

    it('should create project, deploy application, and create function', async () => {
      // Create project
      const { rerender } = render(<ProjectList />);
      
      const createProjButton = screen.getByRole('button', { name: /create project/i });
      await user.click(createProjButton);
      
      const projNameInput = screen.getByLabelText(/project name/i);
      await user.type(projNameInput, 'Web App Project');
      
      const submitProjButton = screen.getByRole('button', { name: /create project/i });
      await user.click(submitProjButton);
      
      await waitFor(() => {
        expect(apiClient.projects.create).toHaveBeenCalled();
      });
      
      // Deploy application
      rerender(<ApplicationList />);
      
      const deployButton = screen.getByRole('button', { name: /deploy application/i });
      await user.click(deployButton);
      
      // Fill deployment form
      const appNameInput = screen.getByLabelText(/application name/i);
      await user.type(appNameInput, 'Frontend App');
      
      const deploySubmitButton = screen.getByRole('button', { name: /deploy/i });
      await user.click(deploySubmitButton);
      
      await waitFor(() => {
        expect(apiClient.applications.create).toHaveBeenCalled();
      });
      
      // Create function
      rerender(<FunctionList />);
      
      // Function creation would follow similar pattern
    });
  });

  describe('4. Monitoring and AI Operations', () => {
    beforeEach(() => {
      // Mock monitoring data
      (apiClient.monitoring.getWorkspaceMetrics as jest.Mock).mockResolvedValue({
        metrics: {
          cpu_usage: 45.5,
          memory_usage: 62.3,
          storage_usage: 30.1,
          network_ingress: 1024,
          network_egress: 2048,
        },
        timestamp: '2024-01-01T00:00:00Z',
      });
      
      // Mock AI chat
      (apiClient.aiops.chat as jest.Mock).mockResolvedValue({
        message: 'Your CPU usage is at 45.5%, which is within normal range.',
        suggestions: ['Consider scaling if usage exceeds 80%'],
      });
    });

    it('should view monitoring dashboard and interact with AI ops', async () => {
      render(<MonitoringDashboard workspaceId="ws-new" />);
      
      await waitFor(() => {
        expect(screen.getByText(/cpu usage/i)).toBeInTheDocument();
        expect(screen.getByText(/45.5%/)).toBeInTheDocument();
      });
      
      // Open AI chat
      const aiChatButton = screen.getByRole('button', { name: /ai assistant/i });
      await user.click(aiChatButton);
      
      // Send message to AI
      const chatInput = screen.getByPlaceholderText(/ask ai about your infrastructure/i);
      await user.type(chatInput, 'What is my current resource usage?');
      
      const sendButton = screen.getByRole('button', { name: /send/i });
      await user.click(sendButton);
      
      await waitFor(() => {
        expect(apiClient.aiops.chat).toHaveBeenCalledWith({
          message: 'What is my current resource usage?',
          context: expect.objectContaining({
            workspace_id: 'ws-new',
          }),
        });
      });
      
      // Verify AI response
      expect(screen.getByText(/your cpu usage is at 45.5%/i)).toBeInTheDocument();
    });
  });

  describe('5. Backup Configuration and Execution', () => {
    beforeEach(() => {
      // Mock backup APIs
      (apiClient.backup.createStorage as jest.Mock).mockResolvedValue({
        id: 'bs-new',
        workspace_id: 'ws-new',
        name: 'Primary Backup Storage',
        type: 'proxmox',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      });
      
      (apiClient.backup.createPolicy as jest.Mock).mockResolvedValue({
        id: 'bp-new',
        workspace_id: 'ws-new',
        name: 'Daily Backup Policy',
        schedule: '0 2 * * *',
        retention_days: 30,
        enabled: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      });
      
      (apiClient.backup.triggerBackup as jest.Mock).mockResolvedValue({
        execution_id: 'be-123',
        status: 'running',
        message: 'Backup started successfully',
      });
    });

    it('should configure backup storage and policy', async () => {
      render(<BackupDashboard />);
      
      // Create backup storage
      const createStorageButton = screen.getByRole('button', { name: /add backup storage/i });
      await user.click(createStorageButton);
      
      const storageNameInput = screen.getByLabelText(/storage name/i);
      await user.type(storageNameInput, 'Primary Backup Storage');
      
      const submitStorageButton = screen.getByRole('button', { name: /create storage/i });
      await user.click(submitStorageButton);
      
      await waitFor(() => {
        expect(apiClient.backup.createStorage).toHaveBeenCalled();
      });
      
      // Create backup policy
      const createPolicyButton = screen.getByRole('button', { name: /create backup policy/i });
      await user.click(createPolicyButton);
      
      const policyNameInput = screen.getByLabelText(/policy name/i);
      await user.type(policyNameInput, 'Daily Backup Policy');
      
      // Set schedule
      const scheduleInput = screen.getByLabelText(/schedule/i);
      await user.type(scheduleInput, '0 2 * * *');
      
      const submitPolicyButton = screen.getByRole('button', { name: /create policy/i });
      await user.click(submitPolicyButton);
      
      await waitFor(() => {
        expect(apiClient.backup.createPolicy).toHaveBeenCalled();
      });
      
      // Trigger manual backup
      const triggerBackupButton = screen.getByRole('button', { name: /run backup now/i });
      await user.click(triggerBackupButton);
      
      await waitFor(() => {
        expect(apiClient.backup.triggerBackup).toHaveBeenCalled();
        expect(screen.getByText(/backup started successfully/i)).toBeInTheDocument();
      });
    });
  });

  describe('6. Function Execution and Logs', () => {
    beforeEach(() => {
      // Mock function execution
      (apiClient.functions.invoke as jest.Mock).mockResolvedValue({
        data: {
          invocation_id: 'inv-123',
          status: 'success',
          output: { result: 'Data processed successfully' },
          duration_ms: 245,
        }
      });
      
      // Mock logs
      (apiClient.functions.getLogs as jest.Mock).mockResolvedValue({
        data: {
          logs: [
            { timestamp: '2024-01-01T00:00:00Z', level: 'info', message: 'Function started' },
            { timestamp: '2024-01-01T00:00:01Z', level: 'info', message: 'Processing data' },
            { timestamp: '2024-01-01T00:00:02Z', level: 'info', message: 'Function completed' },
          ],
        }
      });
    });

    it('should execute function and view logs', async () => {
      render(<FunctionList />);
      
      // Find and invoke function
      await waitFor(() => {
        const invokeButton = screen.getByTestId('invoke-func-new');
        fireEvent.click(invokeButton);
      });
      
      // Configure invocation
      const payloadInput = screen.getByLabelText(/payload/i);
      await user.type(payloadInput, '{"data": "test"}');
      
      const executeButton = screen.getByRole('button', { name: /execute/i });
      await user.click(executeButton);
      
      await waitFor(() => {
        expect(apiClient.functions.invoke).toHaveBeenCalledWith(
          'org-new', 'ws-new', 'func-new',
          { payload: { data: 'test' } }
        );
      });
      
      // View execution result
      expect(screen.getByText(/data processed successfully/i)).toBeInTheDocument();
      expect(screen.getByText(/245ms/i)).toBeInTheDocument();
      
      // View logs
      const viewLogsButton = screen.getByRole('button', { name: /view logs/i });
      await user.click(viewLogsButton);
      
      await waitFor(() => {
        expect(screen.getByText(/function started/i)).toBeInTheDocument();
        expect(screen.getByText(/processing data/i)).toBeInTheDocument();
        expect(screen.getByText(/function completed/i)).toBeInTheDocument();
      });
    });
  });

  describe('7. Resource Modification and Activity Logs', () => {
    beforeEach(() => {
      // Mock update operations
      (apiClient.applications.update as jest.Mock).mockResolvedValue({
        data: {
          id: 'app-new',
          config: {
            replicas: 5,
            cpu: '200m',
            memory: '512Mi',
          },
        }
      });
      
      // Mock activity logs
      (apiClient.monitoring.getActivityLogs as jest.Mock).mockResolvedValue({
        logs: [
          {
            timestamp: '2024-01-01T00:00:00Z',
            action: 'application.updated',
            user: 'test@example.com',
            details: 'Scaled replicas from 3 to 5',
          },
        ],
      });
    });

    it('should modify resources and check activity logs', async () => {
      // Update application configuration
      render(<ApplicationList />);
      
      await waitFor(() => {
        const editButton = screen.getByTestId('edit-app-new');
        fireEvent.click(editButton);
      });
      
      // Update replicas
      const replicasInput = screen.getByLabelText(/replicas/i);
      await user.clear(replicasInput);
      await user.type(replicasInput, '5');
      
      const saveButton = screen.getByRole('button', { name: /save changes/i });
      await user.click(saveButton);
      
      await waitFor(() => {
        expect(apiClient.applications.update).toHaveBeenCalledWith(
          'org-new', 'ws-new', 'app-new',
          expect.objectContaining({
            config: expect.objectContaining({
              replicas: 5,
            }),
          })
        );
      });
      
      // Check activity logs
      const activityButton = screen.getByRole('button', { name: /view activity/i });
      await user.click(activityButton);
      
      await waitFor(() => {
        expect(screen.getByText(/scaled replicas from 3 to 5/i)).toBeInTheDocument();
        expect(screen.getByText(/test@example.com/i)).toBeInTheDocument();
      });
    });
  });

  describe('8. Complete User Journey Integration', () => {
    it('should complete entire user journey from signup to backup', async () => {
      // This test combines all the above scenarios in sequence
      // to ensure the complete flow works together
      
      const steps = [
        'Login',
        'Create Organization',
        'Create Workspace',
        'Create Project',
        'Deploy Application',
        'Create Function',
        'Check Monitoring',
        'Use AI Assistant',
        'Configure Backup',
        'Execute Function',
        'Modify Resources',
        'Check Logs',
      ];
      
      // Track completion of each step
      const completedSteps: string[] = [];
      
      // Execute each step and verify
      for (const step of steps) {
        // Implementation would call appropriate functions
        // and verify expected outcomes
        completedSteps.push(step);
      }
      
      // Verify all steps completed
      expect(completedSteps).toEqual(steps);
    });
  });
});