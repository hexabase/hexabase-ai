import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DeployFunctionDialog } from '@/components/functions/deploy-function-dialog';
import { functionsApi } from '@/lib/api-client';

jest.mock('@/lib/api-client', () => ({
  functionsApi: {
    deploy: jest.fn(),
    getVersions: jest.fn(),
  },
}));

describe('DeployFunctionDialog', () => {
  const mockOnClose = jest.fn();
  const mockOnSuccess = jest.fn();
  const mockFunction = {
    id: 'func-1',
    name: 'image-processor',
    runtime: 'nodejs18',
    handler: 'index.handler',
    version: 'v1.2.0',
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (functionsApi.getVersions as jest.Mock).mockResolvedValue({
      data: {
        versions: [
          { version: 'v1.2.0', deployed_at: '2024-01-01T10:00:00Z', status: 'active' },
          { version: 'v1.1.0', deployed_at: '2023-12-01T10:00:00Z', status: 'inactive' },
        ],
      },
    });
  });

  it('should display function information', () => {
    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    expect(screen.getByText('Deploy Function')).toBeInTheDocument();
    expect(screen.getByText('image-processor')).toBeInTheDocument();
    expect(screen.getByText('Current version: v1.2.0')).toBeInTheDocument();
  });

  it('should show version history', async () => {
    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    await waitFor(() => {
      expect(screen.getByText('v1.2.0')).toBeInTheDocument();
      expect(screen.getByText('v1.1.0')).toBeInTheDocument();
    });
  });

  it('should allow source code upload', () => {
    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const fileInput = screen.getByLabelText(/source code/i);
    expect(fileInput).toBeInTheDocument();
    expect(fileInput).toHaveAttribute('accept', '.zip,.tar,.tar.gz');
  });

  it('should show environment variables editor', () => {
    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    expect(screen.getByText(/environment variables/i)).toBeInTheDocument();
    const addVarButton = screen.getByRole('button', { name: /add variable/i });
    expect(addVarButton).toBeInTheDocument();
  });

  it('should add environment variable', () => {
    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const addVarButton = screen.getByRole('button', { name: /add variable/i });
    fireEvent.click(addVarButton);

    const keyInputs = screen.getAllByPlaceholderText('Key');
    const valueInputs = screen.getAllByPlaceholderText('Value');
    
    expect(keyInputs.length).toBeGreaterThan(0);
    expect(valueInputs.length).toBeGreaterThan(0);
  });

  it('should validate deployment', async () => {
    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const deployButton = screen.getByRole('button', { name: /deploy/i });
    fireEvent.click(deployButton);

    await waitFor(() => {
      expect(screen.getByText(/please upload source code/i)).toBeInTheDocument();
    });
  });

  it('should deploy function with new version', async () => {
    (functionsApi.deploy as jest.Mock).mockResolvedValue({
      data: {
        version: 'v1.3.0',
        status: 'deploying',
      },
    });

    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    // Simulate file upload
    const fileInput = screen.getByLabelText(/source code/i);
    const file = new File(['function code'], 'function.zip', { type: 'application/zip' });
    Object.defineProperty(fileInput, 'files', {
      value: [file],
    });
    fireEvent.change(fileInput);

    const versionInput = screen.getByLabelText(/version tag/i);
    fireEvent.change(versionInput, { target: { value: 'v1.3.0' } });

    const deployButton = screen.getByRole('button', { name: /deploy/i });
    fireEvent.click(deployButton);

    await waitFor(() => {
      expect(functionsApi.deploy).toHaveBeenCalledWith(
        'org-1',
        'ws-1',
        'func-1',
        expect.objectContaining({
          version: 'v1.3.0',
          source: expect.any(File),
        })
      );
    });

    expect(mockOnSuccess).toHaveBeenCalled();
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should handle deployment errors', async () => {
    (functionsApi.deploy as jest.Mock).mockRejectedValue(new Error('Deployment failed'));

    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    // Simulate file upload
    const fileInput = screen.getByLabelText(/source code/i);
    const file = new File(['function code'], 'function.zip', { type: 'application/zip' });
    Object.defineProperty(fileInput, 'files', {
      value: [file],
    });
    fireEvent.change(fileInput);

    const deployButton = screen.getByRole('button', { name: /deploy/i });
    fireEvent.click(deployButton);

    await waitFor(() => {
      expect(screen.getByText(/deployment failed/i)).toBeInTheDocument();
    });
  });

  it('should close dialog on cancel', () => {
    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should rollback to previous version', async () => {
    (functionsApi.deploy as jest.Mock).mockResolvedValue({
      data: {
        version: 'v1.1.0',
        status: 'active',
      },
    });

    render(
      <DeployFunctionDialog
        open={true}
        onClose={mockOnClose}
        onSuccess={mockOnSuccess}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    await waitFor(() => {
      const rollbackButton = screen.getByTestId('rollback-v1.1.0');
      fireEvent.click(rollbackButton);
    });

    // Should show confirmation
    expect(screen.getByText(/rollback to v1.1.0/i)).toBeInTheDocument();
    
    const confirmButton = screen.getByRole('button', { name: /confirm/i });
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(functionsApi.deploy).toHaveBeenCalledWith(
        'org-1',
        'ws-1',
        'func-1',
        expect.objectContaining({
          rollback_to: 'v1.1.0',
        })
      );
    });
  });
});