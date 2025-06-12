import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { FunctionInvocationDialog } from '@/components/functions/function-invocation-dialog';
import { functionsApi } from '@/lib/api-client';

jest.mock('@/lib/api-client', () => ({
  functionsApi: {
    invoke: jest.fn(),
  },
}));

describe('FunctionInvocationDialog', () => {
  const mockOnClose = jest.fn();
  const mockFunction = {
    id: 'func-1',
    name: 'image-processor',
    runtime: 'nodejs18',
    triggers: ['http', 'event'],
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should display function name', () => {
    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    expect(screen.getByText('Invoke Function')).toBeInTheDocument();
    expect(screen.getByText('image-processor')).toBeInTheDocument();
  });

  it('should show trigger type selector', () => {
    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    expect(screen.getByLabelText(/trigger type/i)).toBeInTheDocument();
    expect(screen.getByText('HTTP')).toBeInTheDocument();
    expect(screen.getByText('Event')).toBeInTheDocument();
  });

  it('should show JSON payload editor', () => {
    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const payloadEditor = screen.getByLabelText(/payload/i);
    expect(payloadEditor).toBeInTheDocument();
    expect(payloadEditor).toHaveAttribute('placeholder', expect.stringContaining('JSON'));
  });

  it('should validate JSON payload', async () => {
    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const payloadEditor = screen.getByLabelText(/payload/i);
    fireEvent.change(payloadEditor, { target: { value: 'invalid json' } });

    const invokeButton = screen.getByRole('button', { name: /invoke/i });
    fireEvent.click(invokeButton);

    await waitFor(() => {
      expect(screen.getByText(/invalid json format/i)).toBeInTheDocument();
    });
  });

  it('should invoke function with valid payload', async () => {
    const mockResponse = {
      data: {
        invocation_id: 'inv-123',
        status: 'success',
        output: { result: 'processed' },
        duration_ms: 150,
        logs: 'Function executed successfully',
      },
    };
    (functionsApi.invoke as jest.Mock).mockResolvedValue(mockResponse);

    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const payloadEditor = screen.getByLabelText(/payload/i);
    fireEvent.change(payloadEditor, { target: { value: '{"image": "test.jpg"}' } });

    const invokeButton = screen.getByRole('button', { name: /invoke/i });
    fireEvent.click(invokeButton);

    await waitFor(() => {
      expect(functionsApi.invoke).toHaveBeenCalledWith(
        'org-1',
        'ws-1',
        'func-1',
        {
          trigger_type: 'http',
          payload: { image: 'test.jpg' },
        }
      );
    });

    // Should show result
    expect(screen.getByText(/invocation result/i)).toBeInTheDocument();
    expect(screen.getByText(/success/i)).toBeInTheDocument();
    expect(screen.getByText(/150ms/i)).toBeInTheDocument();
  });

  it('should show function logs', async () => {
    const mockResponse = {
      data: {
        invocation_id: 'inv-123',
        status: 'success',
        output: { result: 'ok' },
        logs: 'Starting function\nProcessing image\nCompleted',
      },
    };
    (functionsApi.invoke as jest.Mock).mockResolvedValue(mockResponse);

    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const invokeButton = screen.getByRole('button', { name: /invoke/i });
    fireEvent.click(invokeButton);

    await waitFor(() => {
      expect(screen.getByText(/function logs/i)).toBeInTheDocument();
      expect(screen.getByText(/starting function/i)).toBeInTheDocument();
    });
  });

  it('should handle invocation errors', async () => {
    const mockResponse = {
      data: {
        invocation_id: 'inv-123',
        status: 'error',
        error: 'Function timeout',
        logs: 'Function started\nTimeout after 30s',
      },
    };
    (functionsApi.invoke as jest.Mock).mockResolvedValue(mockResponse);

    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const invokeButton = screen.getByRole('button', { name: /invoke/i });
    fireEvent.click(invokeButton);

    await waitFor(() => {
      expect(screen.getByText(/error/i)).toBeInTheDocument();
      expect(screen.getByText(/function timeout/i)).toBeInTheDocument();
    });
  });

  it('should allow invoking again', async () => {
    const mockResponse = {
      data: {
        invocation_id: 'inv-123',
        status: 'success',
        output: { result: 'ok' },
      },
    };
    (functionsApi.invoke as jest.Mock).mockResolvedValue(mockResponse);

    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const invokeButton = screen.getByRole('button', { name: /invoke/i });
    fireEvent.click(invokeButton);

    await waitFor(() => {
      const invokeAgainButton = screen.getByRole('button', { name: /invoke again/i });
      expect(invokeAgainButton).toBeInTheDocument();
    });
  });

  it('should handle HTTP trigger type', () => {
    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const triggerSelect = screen.getByLabelText(/trigger type/i);
    fireEvent.change(triggerSelect, { target: { value: 'http' } });

    // Should show HTTP-specific options
    expect(screen.getByLabelText(/http method/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/headers/i)).toBeInTheDocument();
  });

  it('should close dialog', () => {
    render(
      <FunctionInvocationDialog
        open={true}
        onClose={mockOnClose}
        functionData={mockFunction}
        orgId="org-1"
        workspaceId="ws-1"
      />
    );

    const closeButton = screen.getByRole('button', { name: /close/i });
    fireEvent.click(closeButton);

    expect(mockOnClose).toHaveBeenCalled();
  });
});