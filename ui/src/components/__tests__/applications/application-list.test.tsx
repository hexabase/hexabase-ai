import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ApplicationList } from '@/components/applications/application-list';
import { useRouter, useParams } from 'next/navigation';
import { applicationsApi } from '@/lib/api-client';

jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
  useParams: jest.fn(),
}));

jest.mock('@/lib/api-client', () => ({
  applicationsApi: {
    list: jest.fn(),
    create: jest.fn(),
    update: jest.fn(),
    updateStatus: jest.fn(),
    delete: jest.fn(),
  },
}));

// Mock dialog components
jest.mock('@/components/ui/dialog', () => ({
  Dialog: ({ children, open }: any) => open ? <div role="dialog">{children}</div> : null,
  DialogContent: ({ children }: any) => <div>{children}</div>,
  DialogDescription: ({ children }: any) => <div>{children}</div>,
  DialogFooter: ({ children }: any) => <div>{children}</div>,
  DialogHeader: ({ children }: any) => <div>{children}</div>,
  DialogTitle: ({ children }: any) => <h2>{children}</h2>,
}));

jest.mock('@/components/ui/select', () => ({
  Select: ({ children, onValueChange, value }: any) => {
    return <div data-value={value}>{React.Children.map(children, child => 
      React.cloneElement(child, { onValueChange })
    )}</div>;
  },
  SelectTrigger: ({ children, ...props }: any) => <button {...props}>{children}</button>,
  SelectContent: ({ children, onValueChange }: any) => <div onClick={(e: any) => {
    const target = e.target as HTMLElement;
    const value = target.getAttribute('data-value');
    if (value !== null && onValueChange) {
      onValueChange(value);
    }
  }}>{children}</div>,
  SelectItem: ({ children, value }: any) => <div data-value={value}>{children}</div>,
  SelectValue: ({ placeholder }: any) => <span>{placeholder}</span>,
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

// Mock DeployApplicationDialog
jest.mock('@/components/applications/deploy-application-dialog', () => ({
  DeployApplicationDialog: ({ open, onOpenChange, onSubmit }: any) => {
    if (!open) return null;
    return (
      <div role="dialog">
        <h2>Deploy New Application</h2>
        <button onClick={() => {
          onSubmit({
            name: 'New App',
            type: 'stateless',
            source_type: 'image',
            source_image: 'node:latest',
          });
        }}>
          Deploy
        </button>
      </div>
    );
  },
}));

describe('ApplicationList', () => {
  const mockPush = jest.fn();
  const mockApplications = [
    {
      id: 'app-1',
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: 'Frontend App',
      type: 'stateless',
      status: 'running',
      source_type: 'image',
      source_image: 'nginx:latest',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    },
    {
      id: 'app-2',
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: 'Backend API',
      type: 'stateless',
      status: 'running',
      source_type: 'git',
      source_git_url: 'https://github.com/example/api',
      source_git_ref: 'main',
      created_at: '2024-01-02',
      updated_at: '2024-01-02',
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
    (applicationsApi.list as jest.Mock).mockResolvedValue({
      data: {
        applications: mockApplications,
        total: 2,
      },
    });
  });

  it('should display list of applications', async () => {
    render(<ApplicationList />);

    await waitFor(() => {
      expect(screen.getByText('Frontend App')).toBeInTheDocument();
      expect(screen.getByText('Backend API')).toBeInTheDocument();
    });
  });

  it('should show loading state', () => {
    (applicationsApi.list as jest.Mock).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    render(<ApplicationList />);

    expect(screen.getByTestId('applications-skeleton')).toBeInTheDocument();
  });

  it('should handle error state', async () => {
    (applicationsApi.list as jest.Mock).mockRejectedValue(
      new Error('Failed to fetch applications')
    );

    render(<ApplicationList />);

    await waitFor(() => {
      expect(screen.getByText(/failed to load applications/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
    });
  });

  it('should show empty state', async () => {
    (applicationsApi.list as jest.Mock).mockResolvedValue({
      data: {
        applications: [],
        total: 0,
      },
    });

    render(<ApplicationList />);

    await waitFor(() => {
      expect(screen.getByText(/no applications found/i)).toBeInTheDocument();
      expect(screen.getByText(/deploy your first application/i)).toBeInTheDocument();
    });
  });

  it('should filter applications by type', async () => {
    render(<ApplicationList />);

    await waitFor(() => {
      expect(screen.getByText('Frontend App')).toBeInTheDocument();
    });

    const typeFilter = screen.getByTestId('type-filter');
    fireEvent.click(typeFilter);
    
    const statelessOption = screen.getByText('Stateless');
    fireEvent.click(statelessOption.parentElement!);

    await waitFor(() => {
      expect(applicationsApi.list).toHaveBeenCalledWith(
        'org-1', 'ws-1', 'proj-1',
        expect.objectContaining({ type: 'stateless' })
      );
    });
  });

  it('should filter applications by status', async () => {
    render(<ApplicationList />);

    await waitFor(() => {
      expect(screen.getByText('Frontend App')).toBeInTheDocument();
    });

    const statusFilter = screen.getByTestId('status-filter');
    fireEvent.click(statusFilter);
    
    const runningOption = screen.getByText('Running');
    fireEvent.click(runningOption.parentElement!);

    await waitFor(() => {
      expect(applicationsApi.list).toHaveBeenCalledWith(
        'org-1', 'ws-1', 'proj-1',
        expect.objectContaining({ status: 'running' })
      );
    });
  });

  it('should open deploy dialog when clicking deploy button', async () => {
    render(<ApplicationList />);

    await waitFor(() => {
      expect(screen.getByText('Frontend App')).toBeInTheDocument();
    });

    const deployButton = screen.getByRole('button', { name: /deploy application/i });
    fireEvent.click(deployButton);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByText(/deploy new application/i)).toBeInTheDocument();
    });
  });

  it('should navigate to application details when clicked', async () => {
    render(<ApplicationList />);

    await waitFor(() => {
      const appCard = screen.getByTestId('app-card-app-1');
      fireEvent.click(appCard);
    });

    expect(mockPush).toHaveBeenCalledWith(
      '/dashboard/organizations/org-1/workspaces/ws-1/projects/proj-1/applications/app-1'
    );
  });

  it('should update application status', async () => {
    (applicationsApi.updateStatus as jest.Mock).mockResolvedValue({
      data: { ...mockApplications[0], status: 'suspended' },
    });

    render(<ApplicationList />);

    await waitFor(() => {
      const statusButton = screen.getByTestId('status-app-1');
      fireEvent.click(statusButton);
    });

    await waitFor(() => {
      expect(applicationsApi.updateStatus).toHaveBeenCalledWith(
        'org-1', 'ws-1', 'app-1',
        { status: 'suspended' }
      );
    });
  });

  it('should delete application with confirmation', async () => {
    (applicationsApi.delete as jest.Mock).mockResolvedValue({});

    render(<ApplicationList />);

    await waitFor(() => {
      const deleteButton = screen.getByTestId('delete-app-1');
      fireEvent.click(deleteButton);
    });

    expect(screen.getByText(/are you sure you want to delete/i)).toBeInTheDocument();

    const confirmButton = screen.getByRole('button', { name: /delete/i });
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(applicationsApi.delete).toHaveBeenCalledWith(
        'org-1', 'ws-1', 'proj-1', 'app-1'
      );
    });
  });

  it('should refresh list after creating application', async () => {
    const newApp = {
      id: 'app-3',
      name: 'New App',
      type: 'stateless',
      status: 'pending',
      source_type: 'image',
      source_image: 'node:latest',
    };
    (applicationsApi.create as jest.Mock).mockResolvedValue({ data: newApp });

    render(<ApplicationList />);

    await waitFor(() => {
      expect(screen.getByText('Frontend App')).toBeInTheDocument();
    });

    const deployButton = screen.getByRole('button', { name: /deploy application/i });
    fireEvent.click(deployButton);

    // Simulate form submission
    const submitButton = screen.getByRole('button', { name: /deploy/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(applicationsApi.list).toHaveBeenCalledTimes(2); // Initial load + refresh
    });
  });
});
