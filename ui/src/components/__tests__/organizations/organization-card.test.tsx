import { render, screen, fireEvent } from '@testing-library/react';
import { OrganizationCard } from '@/components/organizations/organization-card';

describe('OrganizationCard', () => {
  const mockOrganization = {
    id: 'org-1',
    name: 'ACME Corp',
    role: 'admin' as const,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };

  const mockOnClick = jest.fn();
  const mockOnEdit = jest.fn();
  const mockOnDelete = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should display organization information', () => {
    render(
      <OrganizationCard
        organization={mockOrganization}
        isActive={false}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('ACME Corp')).toBeInTheDocument();
    expect(screen.getByText('admin')).toBeInTheDocument();
  });

  it('should show active state when isActive is true', () => {
    render(
      <OrganizationCard
        organization={mockOrganization}
        isActive={true}
        onClick={mockOnClick}
      />
    );

    const card = screen.getByTestId(`org-card-${mockOrganization.id}`);
    expect(card).toHaveClass('ring-2', 'ring-primary');
    expect(screen.getByText('Active')).toBeInTheDocument();
  });

  it('should call onClick when card is clicked', () => {
    render(
      <OrganizationCard
        organization={mockOrganization}
        isActive={false}
        onClick={mockOnClick}
      />
    );

    const card = screen.getByTestId(`org-card-${mockOrganization.id}`);
    fireEvent.click(card);

    expect(mockOnClick).toHaveBeenCalledWith(mockOrganization.id);
  });

  it('should show edit and delete buttons for admin role', () => {
    render(
      <OrganizationCard
        organization={mockOrganization}
        isActive={false}
        onClick={mockOnClick}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    expect(screen.getByTestId(`edit-${mockOrganization.id}`)).toBeInTheDocument();
    expect(screen.getByTestId(`delete-${mockOrganization.id}`)).toBeInTheDocument();
  });

  it('should not show edit and delete buttons for member role', () => {
    const memberOrg = { ...mockOrganization, role: 'member' as const };
    
    render(
      <OrganizationCard
        organization={memberOrg}
        isActive={false}
        onClick={mockOnClick}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    expect(screen.queryByTestId(`edit-${memberOrg.id}`)).not.toBeInTheDocument();
    expect(screen.queryByTestId(`delete-${memberOrg.id}`)).not.toBeInTheDocument();
  });

  it('should call onEdit when edit button is clicked', () => {
    render(
      <OrganizationCard
        organization={mockOrganization}
        isActive={false}
        onClick={mockOnClick}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    const editButton = screen.getByTestId(`edit-${mockOrganization.id}`);
    fireEvent.click(editButton);

    expect(mockOnEdit).toHaveBeenCalledWith(mockOrganization);
    expect(mockOnClick).not.toHaveBeenCalled(); // Should not trigger card click
  });

  it('should call onDelete when delete button is clicked', () => {
    render(
      <OrganizationCard
        organization={mockOrganization}
        isActive={false}
        onClick={mockOnClick}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    const deleteButton = screen.getByTestId(`delete-${mockOrganization.id}`);
    fireEvent.click(deleteButton);

    expect(mockOnDelete).toHaveBeenCalledWith(mockOrganization.id);
    expect(mockOnClick).not.toHaveBeenCalled(); // Should not trigger card click
  });

  it('should show member count if provided', () => {
    const orgWithMembers = { ...mockOrganization, member_count: 15 };
    
    render(
      <OrganizationCard
        organization={orgWithMembers}
        isActive={false}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('15 members')).toBeInTheDocument();
  });

  it('should show workspace count if provided', () => {
    const orgWithWorkspaces = { ...mockOrganization, workspace_count: 3 };
    
    render(
      <OrganizationCard
        organization={orgWithWorkspaces}
        isActive={false}
        onClick={mockOnClick}
      />
    );

    expect(screen.getByText('3 workspaces')).toBeInTheDocument();
  });
});