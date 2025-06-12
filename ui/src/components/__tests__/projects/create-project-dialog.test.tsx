import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateProjectDialog } from '@/components/projects/create-project-dialog';

describe('CreateProjectDialog', () => {
  const mockOnOpenChange = jest.fn();
  const mockOnSubmit = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render dialog when open', () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText(/create new project/i)).toBeInTheDocument();
  });

  it('should not render dialog when closed', () => {
    render(
      <CreateProjectDialog
        open={false}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('should display form fields', () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    expect(screen.getByLabelText(/project name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/namespace/i)).toBeInTheDocument();
  });

  it('should display resource quota fields', () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    expect(screen.getByLabelText(/cpu limit/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/memory limit/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/storage limit/i)).toBeInTheDocument();
  });

  it('should validate required fields', async () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const submitButton = screen.getByRole('button', { name: /create project/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/project name is required/i)).toBeInTheDocument();
    });

    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it('should auto-generate namespace from project name', () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const nameInput = screen.getByLabelText(/project name/i);
    const namespaceInput = screen.getByLabelText(/namespace/i) as HTMLInputElement;

    fireEvent.change(nameInput, { target: { value: 'My Test Project' } });

    expect(namespaceInput.value).toBe('my-test-project');
  });

  it('should allow manual namespace override', () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const nameInput = screen.getByLabelText(/project name/i);
    const namespaceInput = screen.getByLabelText(/namespace/i) as HTMLInputElement;

    // First set name to auto-generate namespace
    fireEvent.change(nameInput, { target: { value: 'My Test Project' } });
    expect(namespaceInput.value).toBe('my-test-project');

    // Then manually change namespace
    fireEvent.change(namespaceInput, { target: { value: 'custom-namespace' } });
    expect(namespaceInput.value).toBe('custom-namespace');

    // Changing name again should not override manual namespace
    fireEvent.change(nameInput, { target: { value: 'Another Name' } });
    expect(namespaceInput.value).toBe('custom-namespace');
  });

  it('should submit form with valid data', async () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const nameInput = screen.getByLabelText(/project name/i);
    const descriptionInput = screen.getByLabelText(/description/i);
    const cpuInput = screen.getByLabelText(/cpu limit/i);
    const memoryInput = screen.getByLabelText(/memory limit/i);
    const storageInput = screen.getByLabelText(/storage limit/i);

    fireEvent.change(nameInput, { target: { value: 'Test Project' } });
    fireEvent.change(descriptionInput, { target: { value: 'A test project description' } });
    fireEvent.change(cpuInput, { target: { value: '4' } });
    fireEvent.change(memoryInput, { target: { value: '8' } });
    fireEvent.change(storageInput, { target: { value: '20' } });

    const submitButton = screen.getByRole('button', { name: /create project/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        name: 'Test Project',
        description: 'A test project description',
        namespace: 'test-project',
        resource_quota: {
          cpu: '4',
          memory: '8Gi',
          storage: '20Gi',
        },
      });
    });
  });

  it('should validate namespace format', async () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const namespaceInput = screen.getByLabelText(/namespace/i);
    fireEvent.change(namespaceInput, { target: { value: 'Invalid Namespace!' } });

    const submitButton = screen.getByRole('button', { name: /create project/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/namespace can only contain lowercase letters/i)).toBeInTheDocument();
    });
  });

  it('should close dialog on cancel', () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    fireEvent.click(cancelButton);

    expect(mockOnOpenChange).toHaveBeenCalledWith(false);
  });

  it('should reset form when dialog closes', () => {
    const { rerender } = render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const nameInput = screen.getByLabelText(/project name/i) as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'Test Project' } });
    expect(nameInput.value).toBe('Test Project');

    // Close dialog
    rerender(
      <CreateProjectDialog
        open={false}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    // Reopen dialog
    rerender(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const newNameInput = screen.getByLabelText(/project name/i) as HTMLInputElement;
    expect(newNameInput.value).toBe('');
  });

  it('should show loading state during submission', async () => {
    mockOnSubmit.mockImplementation(() => 
      new Promise(resolve => setTimeout(resolve, 100))
    );

    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const nameInput = screen.getByLabelText(/project name/i);
    fireEvent.change(nameInput, { target: { value: 'Test Project' } });

    const submitButton = screen.getByRole('button', { name: /create project/i });
    fireEvent.click(submitButton);

    expect(screen.getByText(/creating/i)).toBeInTheDocument();
    expect(submitButton).toBeDisabled();

    await waitFor(() => {
      expect(screen.queryByText(/creating/i)).not.toBeInTheDocument();
    });
  });

  it('should show preset resource templates', () => {
    render(
      <CreateProjectDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    expect(screen.getByText(/small/i)).toBeInTheDocument();
    expect(screen.getByText(/medium/i)).toBeInTheDocument();
    expect(screen.getByText(/large/i)).toBeInTheDocument();

    // Click on medium template
    const mediumTemplate = screen.getByText(/medium/i).closest('button');
    fireEvent.click(mediumTemplate!);

    const cpuInput = screen.getByLabelText(/cpu limit/i) as HTMLInputElement;
    const memoryInput = screen.getByLabelText(/memory limit/i) as HTMLInputElement;
    const storageInput = screen.getByLabelText(/storage limit/i) as HTMLInputElement;

    expect(cpuInput.value).toBe('4');
    expect(memoryInput.value).toBe('8');
    expect(storageInput.value).toBe('50');
  });
});