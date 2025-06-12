import { render, screen, fireEvent } from '@testing-library/react';
import { ProjectCard } from '@/components/projects/project-card';
import { Project } from '@/lib/api-client';

describe('ProjectCard', () => {
  const mockProject: Project = {
    id: 'proj-1',
    name: 'Web Application',
    workspace_id: 'ws-1',
    description: 'Main web application project',
    namespace: 'web-app',
    status: 'active',
    resource_quota: {
      cpu: '4',
      memory: '8Gi',
      storage: '20Gi',
    },
    resources: {
      deployments: 3,
      services: 5,
      pods: 12,
      configmaps: 8,
      secrets: 4,
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };

  const mockOnClick = jest.fn();
  const mockOnEdit = jest.fn();
  const mockOnDelete = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should display project information', () => {
    render(
      <ProjectCard
        project={mockProject}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('Web Application')).toBeInTheDocument();
    expect(screen.getByText('Main web application project')).toBeInTheDocument();
    expect(screen.getByText('web-app')).toBeInTheDocument();
    // The status is in a badge with an icon, so we need to check the parent element
    const statusBadge = screen.getByText((content, element) => {
      return element?.textContent === 'CheckCircleactive';
    });
    expect(statusBadge).toBeInTheDocument();
  });

  it('should display resource counts', () => {
    render(
      <ProjectCard
        project={mockProject}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('3')).toBeInTheDocument(); // deployments
    expect(screen.getByText('5')).toBeInTheDocument(); // services
    expect(screen.getByText('12')).toBeInTheDocument(); // pods
  });

  it('should display resource quotas', () => {
    render(
      <ProjectCard
        project={mockProject}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText(/4 CPU/)).toBeInTheDocument();
    expect(screen.getByText(/8Gi Memory/)).toBeInTheDocument();
    expect(screen.getByText(/20Gi Storage/)).toBeInTheDocument();
  });

  it('should call onClick when card is clicked', () => {
    render(
      <ProjectCard
        project={mockProject}
        onClick={mockOnClick}
      />
    );

    const card = screen.getByTestId(`project-card-${mockProject.id}`);
    fireEvent.click(card);

    expect(mockOnClick).toHaveBeenCalledWith(mockProject.id);
  });

  it('should show edit button when onEdit is provided', () => {
    render(
      <ProjectCard
        project={mockProject}
        onClick={mockOnClick}
        onEdit={mockOnEdit}
      />
    );

    const editButton = screen.getByTestId(`edit-${mockProject.id}`);
    expect(editButton).toBeInTheDocument();
    
    fireEvent.click(editButton);
    expect(mockOnEdit).toHaveBeenCalledWith(mockProject);
    expect(mockOnClick).not.toHaveBeenCalled();
  });

  it('should show delete button when onDelete is provided', () => {
    render(
      <ProjectCard
        project={mockProject}
        onClick={mockOnClick}
        onDelete={mockOnDelete}
      />
    );

    const deleteButton = screen.getByTestId(`delete-${mockProject.id}`);
    expect(deleteButton).toBeInTheDocument();
    
    fireEvent.click(deleteButton);
    expect(mockOnDelete).toHaveBeenCalledWith(mockProject.id);
    expect(mockOnClick).not.toHaveBeenCalled();
  });

  it('should show status badge with correct variant', () => {
    const testCases = [
      { status: 'active' },
      { status: 'creating' },
      { status: 'error' },
      { status: 'suspended' },
    ];

    testCases.forEach(({ status }) => {
      const { rerender } = render(
        <ProjectCard
          project={{ ...mockProject, status } as any}
          onClick={mockOnClick}
        />
      );

      // The status might be with an icon, so use a flexible matcher
      const badges = screen.getAllByText((content, element) => {
        return element?.textContent?.includes(status) || false;
      });
      expect(badges.length).toBeGreaterThan(0);

      rerender(<></>);
    });
  });

  it('should handle projects without description', () => {
    const projectWithoutDescription = {
      ...mockProject,
      description: undefined,
    };

    render(
      <ProjectCard
        project={projectWithoutDescription}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('No description')).toBeInTheDocument();
  });

  it('should handle projects without resource quota', () => {
    const projectWithoutQuota = {
      ...mockProject,
      resource_quota: undefined,
    };

    render(
      <ProjectCard
        project={projectWithoutQuota}
        onClick={mockOnClick}
      />
    );

    expect(screen.queryByText(/CPU/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Memory/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Storage/)).not.toBeInTheDocument();
  });

  it('should handle empty resource counts', () => {
    const projectWithNoResources = {
      ...mockProject,
      resources: {
        deployments: 0,
        services: 0,
        pods: 0,
        configmaps: 0,
        secrets: 0,
      },
    };

    render(
      <ProjectCard
        project={projectWithNoResources}
        onClick={mockOnClick}
      />
    );

    expect(screen.getAllByText('0')).toHaveLength(3); // deployments, services, pods shown
  });
});