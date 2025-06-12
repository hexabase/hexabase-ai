import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { WorkspaceList } from '@/components/workspaces/workspace-list';
import { useRouter, useParams } from 'next/navigation';
import { workspacesApi, plansApi } from '@/lib/api-client';

jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
  useParams: jest.fn(),
}));
jest.mock('@/lib/api-client', () => ({
  workspacesApi: {
    list: jest.fn(),
    create: jest.fn(),
    update: jest.fn(),
    delete: jest.fn(),
    getKubeconfig: jest.fn(),
  },
  plansApi: {
    list: jest.fn(),
  },
}));

// Mock Dialog to avoid Radix UI act() warnings
jest.mock('@/components/ui/dialog', () => ({
  Dialog: ({ children, open }: any) => open ? <div role="dialog">{children}</div> : null,
  DialogContent: ({ children }: any) => <div>{children}</div>,
  DialogDescription: ({ children }: any) => <div>{children}</div>,
  DialogFooter: ({ children }: any) => <div>{children}</div>,
  DialogHeader: ({ children }: any) => <div>{children}</div>,
  DialogTitle: ({ children }: any) => <h2>{children}</h2>,
}));

jest.mock('@/components/ui/alert-dialog', () => ({
  AlertDialog: ({ children, open }: any) => open ? <div role="alertdialog">{children}</div> : null,
  AlertDialogContent: ({ children }: any) => <div>{children}</div>,
  AlertDialogDescription: ({ children }: any) => <div>{children}</div>,
  AlertDialogFooter: ({ children }: any) => <div>{children}</div>,
  AlertDialogHeader: ({ children }: any) => <div>{children}</div>,
  AlertDialogTitle: ({ children }: any) => <h2>{children}</h2>,
  AlertDialogAction: ({ children, onClick }: any) => <button onClick={onClick}>{children}</button>,
  AlertDialogCancel: ({ children }: any) => <button>{children}</button>,
}));

// Mock CreateWorkspaceDialog to simplify testing
jest.mock('@/components/workspaces/create-workspace-dialog', () => ({
  CreateWorkspaceDialog: ({ open, onOpenChange, plans, onSubmit }: any) => {
    if (!open) return null;
    return (
      <div role="dialog">
        <h2>Create New Workspace</h2>
        {plans?.map((plan: any) => (
          <div key={plan.id}>
            <span>{plan.name}</span>
            <span>${plan.price}/month</span>
          </div>
        ))}
        <input aria-label="Workspace Name" />
        <input type="radio" id="plan-shared" aria-label="Shared Plan" />
        <button onClick={() => onSubmit('Staging', 'plan-shared')}>
          Create Workspace
        </button>
      </div>
    );
  },
}));

describe('WorkspaceList', () => {
  const mockPush = jest.fn();
  const mockWorkspaces = [
    {
      id: 'ws-1',
      name: 'Production',
      plan_id: 'plan-dedicated',
      vcluster_status: 'active',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    },
    {
      id: 'ws-2',
      name: 'Development',
      plan_id: 'plan-shared',
      vcluster_status: 'active',
      created_at: '2024-01-02',
      updated_at: '2024-01-02',
    },
  ];

  const mockPlans = [
    {
      id: 'plan-shared',
      name: 'Shared Plan',
      description: 'Shared resources',
      price: 0,
      currency: 'USD',
    },
    {
      id: 'plan-dedicated',
      name: 'Dedicated Plan',
      description: 'Dedicated resources',
      price: 299,
      currency: 'USD',
    },
  ];

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue({ push: mockPush });
    (useParams as jest.Mock).mockReturnValue({ orgId: 'org-1' });
    (workspacesApi.list as jest.Mock).mockResolvedValue({
      workspaces: mockWorkspaces,
      total: 2,
    });
    (plansApi.list as jest.Mock).mockResolvedValue({
      plans: mockPlans,
      total: 2,
    });
  });

  it('should display list of workspaces', async () => {
    render(<WorkspaceList />);

    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
      expect(screen.getByText('Development')).toBeInTheDocument();
    });
  });

  it('should show workspace status', async () => {
    render(<WorkspaceList />);

    await waitFor(() => {
      const statusBadges = screen.getAllByText('active');
      expect(statusBadges).toHaveLength(2);
    });
  });

  it('should show plan information', async () => {
    render(<WorkspaceList />);

    await waitFor(() => {
      expect(screen.getByText('Dedicated Plan')).toBeInTheDocument();
      expect(screen.getByText('Shared Plan')).toBeInTheDocument();
    });
  });

  it('should navigate to workspace details when clicked', async () => {
    render(<WorkspaceList />);

    await waitFor(() => {
      const productionCard = screen.getByTestId('workspace-card-ws-1');
      fireEvent.click(productionCard);
    });

    expect(mockPush).toHaveBeenCalledWith('/dashboard/organizations/org-1/workspaces/ws-1');
  });

  it('should open create workspace dialog', async () => {
    render(<WorkspaceList />);

    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    const createButton = screen.getByRole('button', { name: /create workspace/i });
    
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByText(/create new workspace/i)).toBeInTheDocument();
    });
  });

  it('should show available plans in create dialog', async () => {
    render(<WorkspaceList />);

    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    const createButton = screen.getByRole('button', { name: /create workspace/i });
    
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      // Check that plans are displayed in the dialog
      const dialogElement = screen.getByRole('dialog');
      expect(dialogElement).toHaveTextContent('Shared Plan');
      expect(dialogElement).toHaveTextContent('Dedicated Plan');
      // Our mock shows prices as $X/month format
      expect(dialogElement).toHaveTextContent('$0/month'); 
      expect(dialogElement).toHaveTextContent('$299/month');
    });
  });

  it('should create new workspace', async () => {
    const newWorkspace = {
      id: 'ws-3',
      name: 'Staging',
      plan_id: 'plan-shared',
      vcluster_status: 'provisioning',
      created_at: '2024-01-03',
      updated_at: '2024-01-03',
    };
    (workspacesApi.create as jest.Mock).mockResolvedValue(newWorkspace);

    render(<WorkspaceList />);

    await waitFor(() => {
      expect(screen.getByText('Production')).toBeInTheDocument();
    });

    const createButton = screen.getByRole('button', { name: /create workspace/i });
    
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    const submitButtons = screen.getAllByRole('button', { name: /create workspace/i });
    const dialogSubmitButton = submitButtons.find(btn => 
      btn.parentElement?.closest('[role="dialog"]')
    )!;
    fireEvent.click(dialogSubmitButton);

    await waitFor(() => {
      expect(workspacesApi.create).toHaveBeenCalledWith('org-1', {
        name: 'Staging',
        plan_id: 'plan-shared',
      });
    });
  });

  it('should show loading state while fetching workspaces', () => {
    (workspacesApi.list as jest.Mock).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    render(<WorkspaceList />);

    expect(screen.getByTestId('workspaces-skeleton')).toBeInTheDocument();
  });

  it('should handle error when fetching workspaces fails', async () => {
    (workspacesApi.list as jest.Mock).mockRejectedValue(
      new Error('Failed to fetch workspaces')
    );

    render(<WorkspaceList />);

    await waitFor(() => {
      expect(screen.getByText(/failed to load workspaces/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
    });
  });

  it('should show empty state when no workspaces', async () => {
    (workspacesApi.list as jest.Mock).mockResolvedValue({
      workspaces: [],
      total: 0,
    });

    render(<WorkspaceList />);

    await waitFor(() => {
      expect(screen.getByText(/no workspaces found/i)).toBeInTheDocument();
      expect(screen.getByText(/create your first workspace/i)).toBeInTheDocument();
    });
  });

  it('should download kubeconfig', async () => {
    const mockKubeconfig = 'apiVersion: v1\nkind: Config\n...';
    (workspacesApi.getKubeconfig as jest.Mock).mockResolvedValue({
      kubeconfig: mockKubeconfig,
      workspace: 'ws-1',
      status: 'active',
    });

    render(<WorkspaceList />);

    await waitFor(() => {
      const downloadButton = screen.getByTestId('download-kubeconfig-ws-1');
      fireEvent.click(downloadButton);
    });

    await waitFor(() => {
      expect(workspacesApi.getKubeconfig).toHaveBeenCalledWith('org-1', 'ws-1');
    });
  });

  it('should delete workspace with confirmation', async () => {
    (workspacesApi.delete as jest.Mock).mockResolvedValue({});

    render(<WorkspaceList />);

    await waitFor(() => {
      const deleteButton = screen.getByTestId('delete-ws-1');
      fireEvent.click(deleteButton);
    });

    expect(screen.getByText(/are you sure you want to delete/i)).toBeInTheDocument();

    const confirmButton = screen.getAllByRole('button', { name: /delete/i }).find(
      btn => btn.textContent === 'Delete' && !btn.hasAttribute('data-testid')
    )!;
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(workspacesApi.delete).toHaveBeenCalledWith('org-1', 'ws-1');
    });
  });

  it('should show different actions based on workspace status', async () => {
    const workspacesWithDifferentStatus = [
      { ...mockWorkspaces[0], vcluster_status: 'provisioning' },
      { ...mockWorkspaces[1], vcluster_status: 'error' },
    ];

    (workspacesApi.list as jest.Mock).mockResolvedValue({
      workspaces: workspacesWithDifferentStatus,
      total: 2,
    });

    render(<WorkspaceList />);

    await waitFor(() => {
      expect(screen.getByText('provisioning')).toBeInTheDocument();
      expect(screen.getByText('error')).toBeInTheDocument();
      
      // Kubeconfig download should be disabled for non-active workspaces
      const downloadButton = screen.getByTestId('download-kubeconfig-ws-1');
      expect(downloadButton).toBeDisabled();
    });
  });
});