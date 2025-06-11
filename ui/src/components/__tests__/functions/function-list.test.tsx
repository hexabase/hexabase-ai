import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { FunctionList } from '@/components/functions/function-list';
import { useRouter, useParams } from 'next/navigation';
import { apiClient } from '@/lib/api-client';

jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
  useParams: jest.fn(),
}));

jest.mock('@/lib/api-client', () => ({
  apiClient: {
    functions: {
      list: jest.fn(),
      create: jest.fn(),
      update: jest.fn(),
      delete: jest.fn(),
      deploy: jest.fn(),
      invoke: jest.fn(),
      getLogs: jest.fn(),
      getVersions: jest.fn(),
    },
  },
}));

describe('FunctionList', () => {
  const mockPush = jest.fn();
  const mockFunctions = [
    {
      id: 'func-1',
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: 'image-processor',
      description: 'Processes uploaded images',
      runtime: 'nodejs18',
      handler: 'index.handler',
      timeout: 30,
      memory: 256,
      environment_vars: {
        IMAGE_BUCKET: 'images',
      },
      triggers: ['http', 'event'],
      status: 'active',
      version: 'v1.2.0',
      last_deployed_at: '2024-01-01T10:00:00Z',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    },
    {
      id: 'func-2',
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: 'data-transformer',
      description: 'Transforms data between formats',
      runtime: 'python39',
      handler: 'main.handler',
      timeout: 60,
      memory: 512,
      environment_vars: {},
      triggers: ['schedule'],
      status: 'updating',
      version: 'v2.0.0',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    },
  ];

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue({ push: mockPush });
    (useParams as jest.Mock).mockReturnValue({
      orgId: 'org-1',
      workspaceId: 'ws-1',
      projectId: 'proj-1',
    });
    (apiClient.functions.list as jest.Mock).mockResolvedValue({
      data: {
        functions: mockFunctions,
        total: 2,
      },
    });
  });

  it('should display list of functions', async () => {
    render(<FunctionList />);

    await waitFor(() => {
      expect(screen.getByText('image-processor')).toBeInTheDocument();
      expect(screen.getByText('data-transformer')).toBeInTheDocument();
    });
  });

  it('should show function runtime and memory', async () => {
    render(<FunctionList />);

    await waitFor(() => {
      expect(screen.getByText('nodejs18')).toBeInTheDocument();
      expect(screen.getByText('256 MB')).toBeInTheDocument();
      expect(screen.getByText('python39')).toBeInTheDocument();
      expect(screen.getByText('512 MB')).toBeInTheDocument();
    });
  });

  it('should show function triggers', async () => {
    render(<FunctionList />);

    await waitFor(() => {
      expect(screen.getByText('http')).toBeInTheDocument();
      expect(screen.getByText('event')).toBeInTheDocument();
      expect(screen.getByText('schedule')).toBeInTheDocument();
    });
  });

  it('should show function version and status', async () => {
    render(<FunctionList />);

    await waitFor(() => {
      expect(screen.getByText('v1.2.0')).toBeInTheDocument();
      expect(screen.getByText('active')).toBeInTheDocument();
      expect(screen.getByText('v2.0.0')).toBeInTheDocument();
      expect(screen.getByText('updating')).toBeInTheDocument();
    });
  });

  it('should open function details on click', async () => {
    render(<FunctionList />);

    await waitFor(() => {
      const functionCard = screen.getByText('image-processor').closest('div[role="button"]');
      fireEvent.click(functionCard!);
    });

    expect(mockPush).toHaveBeenCalledWith(
      '/dashboard/organizations/org-1/workspaces/ws-1/projects/proj-1/functions/func-1'
    );
  });

  it('should invoke function', async () => {
    const mockInvokeResponse = {
      data: {
        invocation_id: 'inv-123',
        status: 'success',
        output: { result: 'ok' },
      },
    };
    (apiClient.functions.invoke as jest.Mock).mockResolvedValue(mockInvokeResponse);

    render(<FunctionList />);

    await waitFor(() => {
      const invokeButton = screen.getByTestId('invoke-func-1');
      fireEvent.click(invokeButton);
    });

    expect(apiClient.functions.invoke).toHaveBeenCalledWith('org-1', 'ws-1', 'func-1', {});
  });

  it('should deploy new version', async () => {
    const mockDeployResponse = {
      data: {
        version: 'v1.2.1',
        status: 'deploying',
      },
    };
    (apiClient.functions.deploy as jest.Mock).mockResolvedValue(mockDeployResponse);

    render(<FunctionList />);

    await waitFor(() => {
      const deployButton = screen.getByTestId('deploy-func-1');
      fireEvent.click(deployButton);
    });

    // Should open deploy dialog
    expect(screen.getByRole('heading', { name: /deploy function/i })).toBeInTheDocument();
  });

  it('should view function logs', async () => {
    render(<FunctionList />);

    await waitFor(() => {
      const logsButton = screen.getByTestId('logs-func-1');
      fireEvent.click(logsButton);
    });

    // Should open logs dialog
    expect(screen.getByRole('heading', { name: /function logs/i })).toBeInTheDocument();
  });

  it('should filter functions by runtime', async () => {
    render(<FunctionList />);

    await waitFor(() => {
      expect(screen.getByText('image-processor')).toBeInTheDocument();
    });

    const runtimeFilter = screen.getByTestId('runtime-filter');
    fireEvent.click(runtimeFilter);
    
    const nodejsOption = screen.getByText('Node.js');
    fireEvent.click(nodejsOption);

    await waitFor(() => {
      expect(apiClient.functions.list).toHaveBeenCalledWith(
        'org-1', 'ws-1', 'proj-1',
        expect.objectContaining({ runtime: 'nodejs' })
      );
    });
  });

  it('should show empty state', async () => {
    (apiClient.functions.list as jest.Mock).mockResolvedValue({
      data: {
        functions: [],
        total: 0,
      },
    });

    render(<FunctionList />);

    await waitFor(() => {
      expect(screen.getByText(/no functions found/i)).toBeInTheDocument();
    });
  });
});