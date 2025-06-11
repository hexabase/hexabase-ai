import { render, screen, fireEvent } from '@testing-library/react';
import { ApplicationCard } from '@/components/applications/application-card';
import { Application } from '@/lib/api-client';

describe('ApplicationCard', () => {
  const mockOnClick = jest.fn();
  const mockOnEdit = jest.fn();
  const mockOnDelete = jest.fn();
  const mockOnStatusChange = jest.fn();

  const mockApplication: Application = {
    id: 'app-1',
    workspace_id: 'ws-1',
    project_id: 'proj-1',
    name: 'Frontend App',
    type: 'stateless',
    status: 'running',
    source_type: 'image',
    source_image: 'nginx:latest',
    endpoints: {
      external: 'https://frontend.example.com',
      internal: 'http://frontend.default.svc.cluster.local',
    },
    config: {
      replicas: 3,
      cpu: '100m',
      memory: '256Mi',
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should display application information', () => {
    render(
      <ApplicationCard
        application={mockApplication}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('Frontend App')).toBeInTheDocument();
    expect(screen.getByText('stateless')).toBeInTheDocument();
    expect(screen.getByText('nginx:latest')).toBeInTheDocument();
  });

  it('should display application status', () => {
    render(
      <ApplicationCard
        application={mockApplication}
        onClick={mockOnClick}
      />
    );

    const statusBadge = screen.getByText('running');
    expect(statusBadge).toBeInTheDocument();
    expect(statusBadge).toHaveClass('bg-success');
  });

  it('should display different statuses with correct colors', () => {
    const statuses = [
      { status: 'running', color: 'bg-success' },
      { status: 'pending', color: 'bg-warning' },
      { status: 'error', color: 'bg-destructive' },
      { status: 'terminating', color: 'bg-secondary' },
    ];

    statuses.forEach(({ status, color }) => {
      const { rerender } = render(
        <ApplicationCard
          application={{ ...mockApplication, status } as Application}
          onClick={mockOnClick}
        />
      );

      const statusBadge = screen.getByText(status);
      expect(statusBadge).toHaveClass(color);

      rerender(<></>);
    });
  });

  it('should display endpoints when available', () => {
    render(
      <ApplicationCard
        application={mockApplication}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText(/frontend.example.com/)).toBeInTheDocument();
  });

  it('should display configuration details', () => {
    render(
      <ApplicationCard
        application={mockApplication}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('3 replicas')).toBeInTheDocument();
    expect(screen.getByText('100m CPU')).toBeInTheDocument();
    expect(screen.getByText('256Mi Memory')).toBeInTheDocument();
  });

  it('should call onClick when card is clicked', () => {
    render(
      <ApplicationCard
        application={mockApplication}
        onClick={mockOnClick}
      />
    );

    const card = screen.getByTestId(`app-card-${mockApplication.id}`);
    fireEvent.click(card);

    expect(mockOnClick).toHaveBeenCalledWith(mockApplication.id);
  });

  it('should show action buttons when handlers are provided', () => {
    render(
      <ApplicationCard
        application={mockApplication}
        onClick={mockOnClick}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onStatusChange={mockOnStatusChange}
      />
    );

    expect(screen.getByTestId(`edit-${mockApplication.id}`)).toBeInTheDocument();
    expect(screen.getByTestId(`delete-${mockApplication.id}`)).toBeInTheDocument();
    expect(screen.getByTestId(`status-${mockApplication.id}`)).toBeInTheDocument();
  });

  it('should not trigger card click when action buttons are clicked', () => {
    render(
      <ApplicationCard
        application={mockApplication}
        onClick={mockOnClick}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    const editButton = screen.getByTestId(`edit-${mockApplication.id}`);
    fireEvent.click(editButton);

    expect(mockOnEdit).toHaveBeenCalledWith(mockApplication);
    expect(mockOnClick).not.toHaveBeenCalled();
  });

  it('should handle applications without endpoints', () => {
    const appWithoutEndpoints = {
      ...mockApplication,
      endpoints: undefined,
    };

    render(
      <ApplicationCard
        application={appWithoutEndpoints}
        onClick={mockOnClick}
      />
    );

    expect(screen.queryByText(/example.com/)).not.toBeInTheDocument();
  });

  it('should display source information based on type', () => {
    const gitApp = {
      ...mockApplication,
      source_type: 'git' as const,
      source_git_url: 'https://github.com/example/app',
      source_git_ref: 'main',
      source_image: undefined,
    };

    render(
      <ApplicationCard
        application={gitApp}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText(/github.com\/example\/app/)).toBeInTheDocument();
    expect(screen.getByText('main')).toBeInTheDocument();
  });

  it('should show stop/start button based on status', () => {
    render(
      <ApplicationCard
        application={mockApplication}
        onClick={mockOnClick}
        onStatusChange={mockOnStatusChange}
      />
    );

    const statusButton = screen.getByTestId(`status-${mockApplication.id}`);
    expect(statusButton).toHaveTextContent('Stop');

    fireEvent.click(statusButton);
    expect(mockOnStatusChange).toHaveBeenCalledWith(mockApplication.id, 'suspended');
  });
});