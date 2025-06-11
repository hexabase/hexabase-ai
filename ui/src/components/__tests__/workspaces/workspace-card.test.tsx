import { render, screen, fireEvent } from '@testing-library/react';
import { WorkspaceCard } from '@/components/workspaces/workspace-card';
import { Workspace } from '@/lib/api-client';

describe('WorkspaceCard', () => {
  const mockWorkspace: Workspace = {
    id: 'ws-1',
    name: 'Production',
    plan_id: 'plan-dedicated',
    vcluster_status: 'active',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };

  const mockPlan = {
    id: 'plan-dedicated',
    name: 'Dedicated Plan',
    description: 'Dedicated resources',
    price: 299,
    currency: 'USD',
  };

  const mockOnClick = jest.fn();
  const mockOnDelete = jest.fn();
  const mockOnDownloadKubeconfig = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should display workspace information', () => {
    render(
      <WorkspaceCard
        workspace={mockWorkspace}
        plan={mockPlan}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('Production')).toBeInTheDocument();
    expect(screen.getByText('Dedicated Plan')).toBeInTheDocument();
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('should call onClick when card is clicked', () => {
    render(
      <WorkspaceCard
        workspace={mockWorkspace}
        plan={mockPlan}
        onClick={mockOnClick}
      />
    );

    const card = screen.getByTestId(`workspace-card-${mockWorkspace.id}`);
    fireEvent.click(card);

    expect(mockOnClick).toHaveBeenCalledWith(mockWorkspace.id);
  });

  it('should show status badge with correct variant', () => {
    const testCases = [
      { status: 'active', variant: 'success' },
      { status: 'provisioning', variant: 'warning' },
      { status: 'error', variant: 'destructive' },
      { status: 'suspended', variant: 'secondary' },
    ];

    testCases.forEach(({ status, variant }) => {
      const { rerender } = render(
        <WorkspaceCard
          workspace={{ ...mockWorkspace, vcluster_status: status }}
          plan={mockPlan}
          onClick={mockOnClick}
        />
      );

      const badge = screen.getByText(status);
      expect(badge).toHaveClass(`bg-${variant}`);

      rerender(<></>); // Clean up for next iteration
    });
  });

  it('should show download kubeconfig button when status is active', () => {
    render(
      <WorkspaceCard
        workspace={mockWorkspace}
        plan={mockPlan}
        onClick={mockOnClick}
        onDownloadKubeconfig={mockOnDownloadKubeconfig}
      />
    );

    const downloadButton = screen.getByTestId(`download-kubeconfig-${mockWorkspace.id}`);
    expect(downloadButton).toBeInTheDocument();
    expect(downloadButton).not.toBeDisabled();
  });

  it('should disable download kubeconfig button when status is not active', () => {
    render(
      <WorkspaceCard
        workspace={{ ...mockWorkspace, vcluster_status: 'provisioning' }}
        plan={mockPlan}
        onClick={mockOnClick}
        onDownloadKubeconfig={mockOnDownloadKubeconfig}
      />
    );

    const downloadButton = screen.getByTestId(`download-kubeconfig-${mockWorkspace.id}`);
    expect(downloadButton).toBeDisabled();
  });

  it('should call onDownloadKubeconfig when download button is clicked', () => {
    render(
      <WorkspaceCard
        workspace={mockWorkspace}
        plan={mockPlan}
        onClick={mockOnClick}
        onDownloadKubeconfig={mockOnDownloadKubeconfig}
      />
    );

    const downloadButton = screen.getByTestId(`download-kubeconfig-${mockWorkspace.id}`);
    fireEvent.click(downloadButton);

    expect(mockOnDownloadKubeconfig).toHaveBeenCalledWith(mockWorkspace.id);
    expect(mockOnClick).not.toHaveBeenCalled(); // Should not trigger card click
  });

  it('should show delete button and call onDelete', () => {
    render(
      <WorkspaceCard
        workspace={mockWorkspace}
        plan={mockPlan}
        onClick={mockOnClick}
        onDelete={mockOnDelete}
      />
    );

    const deleteButton = screen.getByTestId(`delete-${mockWorkspace.id}`);
    fireEvent.click(deleteButton);

    expect(mockOnDelete).toHaveBeenCalledWith(mockWorkspace.id);
    expect(mockOnClick).not.toHaveBeenCalled();
  });

  it('should show price information from plan', () => {
    render(
      <WorkspaceCard
        workspace={mockWorkspace}
        plan={mockPlan}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('$299/month')).toBeInTheDocument();
  });

  it('should show free badge for zero price plans', () => {
    const freePlan = { ...mockPlan, price: 0 };
    
    render(
      <WorkspaceCard
        workspace={mockWorkspace}
        plan={freePlan}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('Free')).toBeInTheDocument();
  });

  it('should show resource limits if provided', () => {
    const planWithLimits = {
      ...mockPlan,
      resource_limits: {
        cpu: '4 vCPU',
        memory: '16GB',
        storage: '100GB',
      },
    };

    render(
      <WorkspaceCard
        workspace={mockWorkspace}
        plan={planWithLimits}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText(/4 vCPU/)).toBeInTheDocument();
    expect(screen.getByText(/16GB/)).toBeInTheDocument();
    expect(screen.getByText(/100GB/)).toBeInTheDocument();
  });
});