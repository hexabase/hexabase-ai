import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { OrganizationList } from '@/components/organizations/organization-list';
import { mockAuth } from '@/lib/auth-context';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';

// Auth context is already mocked in jest.setup.tsx
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));
jest.mock('@/lib/api-client', () => ({
  apiClient: {
    organizations: {
      list: jest.fn(),
      create: jest.fn(),
      update: jest.fn(),
      delete: jest.fn(),
    },
  },
}));

describe('OrganizationList', () => {
  const mockPush = jest.fn();
  const mockOrganizations = [
    { id: 'org-1', name: 'ACME Corp', role: 'admin', created_at: '2024-01-01', updated_at: '2024-01-01' },
    { id: 'org-2', name: 'Tech Startup', role: 'member', created_at: '2024-01-02', updated_at: '2024-01-02' },
  ];

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue({ push: mockPush });
    
    // Update the mock auth values
    mockAuth.activeOrganization = mockOrganizations[0];
    mockAuth.switchOrganization = jest.fn();
    
    (apiClient.organizations.list as jest.Mock).mockResolvedValue({ 
      data: { organizations: mockOrganizations, total: 2 } 
    });
  });

  it('should display list of organizations', async () => {
    render(<OrganizationList />);

    await waitFor(() => {
      expect(screen.getByText('ACME Corp')).toBeInTheDocument();
      expect(screen.getByText('Tech Startup')).toBeInTheDocument();
    });
  });

  it('should show role badges for each organization', async () => {
    render(<OrganizationList />);

    await waitFor(() => {
      expect(screen.getByText('admin')).toBeInTheDocument();
      expect(screen.getByText('member')).toBeInTheDocument();
    });
  });

  it('should highlight active organization', async () => {
    render(<OrganizationList />);

    await waitFor(() => {
      const activeOrgCard = screen.getByTestId('org-card-org-1');
      expect(activeOrgCard).toHaveClass('ring-2', 'ring-primary');
    });
  });

  it('should switch organization when clicked', async () => {
    const mockSwitchOrg = jest.fn();
    mockAuth.switchOrganization = mockSwitchOrg;

    render(<OrganizationList />);

    await waitFor(() => {
      const techStartupCard = screen.getByTestId('org-card-org-2');
      fireEvent.click(techStartupCard);
    });

    expect(mockSwitchOrg).toHaveBeenCalledWith('org-2');
    expect(mockPush).toHaveBeenCalledWith('/dashboard/organizations/org-2');
  });

  it('should open create organization dialog when clicking create button', async () => {
    render(<OrganizationList />);

    // Wait for organizations to load
    await waitFor(() => {
      expect(screen.getByText('ACME Corp')).toBeInTheDocument();
    });

    const createButton = screen.getByRole('button', { name: /create organization/i });
    fireEvent.click(createButton);

    expect(screen.getByTestId('create-organization-dialog')).toBeInTheDocument();
  });

  it('should create new organization', async () => {
    const newOrg = { id: 'org-3', name: 'New Org', role: 'admin', created_at: '2024-01-03', updated_at: '2024-01-03' };
    (apiClient.organizations.create as jest.Mock).mockResolvedValue({ data: newOrg });

    render(<OrganizationList />);

    // Wait for organizations to load
    await waitFor(() => {
      expect(screen.getByText('ACME Corp')).toBeInTheDocument();
    });

    const createButton = screen.getByRole('button', { name: /create organization/i });
    fireEvent.click(createButton);

    // Wait for dialog to open
    await waitFor(() => {
      expect(screen.getByTestId('org-name-input')).toBeInTheDocument();
    });

    const nameInput = screen.getByTestId('org-name-input');
    fireEvent.change(nameInput, { target: { value: 'New Org' } });

    const submitButton = screen.getByTestId('create-org-submit');
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(apiClient.organizations.create).toHaveBeenCalledWith({ name: 'New Org' });
    });
  });

  it('should show loading state while fetching organizations', () => {
    (apiClient.organizations.list as jest.Mock).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    render(<OrganizationList />);

    expect(screen.getByTestId('organizations-skeleton')).toBeInTheDocument();
  });

  it('should handle error when fetching organizations fails', async () => {
    (apiClient.organizations.list as jest.Mock).mockRejectedValue(
      new Error('Failed to fetch organizations')
    );

    render(<OrganizationList />);

    await waitFor(() => {
      expect(screen.getByText(/failed to load organizations/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
    });
  });

  it('should show empty state when no organizations', async () => {
    (apiClient.organizations.list as jest.Mock).mockResolvedValue({ 
      data: { organizations: [], total: 0 } 
    });

    render(<OrganizationList />);

    await waitFor(() => {
      expect(screen.getByText(/no organizations found/i)).toBeInTheDocument();
      expect(screen.getByText(/create your first organization/i)).toBeInTheDocument();
    });
  });

  it('should allow editing organization name for admin role', async () => {
    render(<OrganizationList />);

    await waitFor(() => {
      expect(screen.getByText('ACME Corp')).toBeInTheDocument();
    });

    const editButton = screen.getByTestId('edit-org-1');
    fireEvent.click(editButton);

    await waitFor(() => {
      expect(screen.getByTestId('edit-organization-dialog')).toBeInTheDocument();
    });
  });

  it('should not show edit button for member role', async () => {
    render(<OrganizationList />);

    await waitFor(() => {
      expect(screen.queryByTestId('edit-org-2')).not.toBeInTheDocument();
    });
  });

  it('should delete organization with confirmation', async () => {
    (apiClient.organizations.delete as jest.Mock).mockResolvedValue({});

    render(<OrganizationList />);

    await waitFor(() => {
      const deleteButton = screen.getByTestId('delete-org-1');
      fireEvent.click(deleteButton);
    });

    // Confirmation dialog
    expect(screen.getByText(/are you sure you want to delete/i)).toBeInTheDocument();
    
    const confirmButton = screen.getByRole('button', { name: /delete/i });
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(apiClient.organizations.delete).toHaveBeenCalledWith('org-1');
    });
  });
});