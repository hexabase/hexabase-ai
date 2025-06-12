import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DeployApplicationDialog } from '@/components/applications/deploy-application-dialog';

// Mock UI components
jest.mock('@/components/ui/dialog', () => ({
  Dialog: ({ children, open }: any) => open ? <div role="dialog">{children}</div> : null,
  DialogContent: ({ children }: any) => <div>{children}</div>,
  DialogDescription: ({ children }: any) => <div>{children}</div>,
  DialogFooter: ({ children }: any) => <div>{children}</div>,
  DialogHeader: ({ children }: any) => <div>{children}</div>,
  DialogTitle: ({ children }: any) => <h2>{children}</h2>,
}));

jest.mock('@/components/ui/radio-group', () => ({
  RadioGroup: ({ children, value, onValueChange }: any) => <div data-value={value} onChange={(e: any) => onValueChange(e.target.value)}>{children}</div>,
  RadioGroupItem: ({ value, id }: any) => <input type="radio" value={value} id={id} name="radio-group" />,
}));

jest.mock('@/components/ui/select', () => ({
  Select: ({ children, value, onValueChange }: any) => {
    const childrenWithProps = React.Children.map(children, child => {
      if (child?.props?.id) {
        return <select id={child.props.id} value={value} onChange={(e) => onValueChange(e.target.value)}>
          {child.props.children}
        </select>;
      }
      return child;
    });
    return <>{childrenWithProps}</>;
  },
  SelectTrigger: ({ children, id }: any) => <div id={id}>{children}</div>,
  SelectContent: ({ children }: any) => <>{children}</>,
  SelectItem: ({ children, value }: any) => <option value={value}>{children}</option>,
  SelectValue: ({ placeholder }: any) => <span>{placeholder}</span>,
}));

jest.mock('@/components/ui/card', () => ({
  Card: ({ children, className, onClick }: any) => <div className={className} onClick={onClick}>{children}</div>,
  CardContent: ({ children }: any) => <div>{children}</div>,
}));

describe('DeployApplicationDialog', () => {
  const mockOnOpenChange = jest.fn();
  const mockOnSubmit = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render dialog when open', () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText(/deploy new application/i)).toBeInTheDocument();
  });

  it('should not render dialog when closed', () => {
    render(
      <DeployApplicationDialog
        open={false}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('should display application type options', () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    expect(screen.getByText('Stateless Application')).toBeInTheDocument();
    expect(screen.getByText('Stateful Application')).toBeInTheDocument();
    expect(screen.getByText('CronJob')).toBeInTheDocument();
    expect(screen.getByText('Serverless Function')).toBeInTheDocument();
  });

  it('should display source type options', () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    expect(screen.getByLabelText(/container image/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/git repository/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/buildpack/i)).toBeInTheDocument();
  });

  it('should show image fields when image source is selected', () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const imageRadio = screen.getByLabelText(/container image/i);
    fireEvent.click(imageRadio);

    expect(screen.getByLabelText(/image name/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/nginx:latest/)).toBeInTheDocument();
  });

  it('should show git fields when git source is selected', () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const gitRadio = screen.getByLabelText(/git repository/i);
    fireEvent.click(gitRadio);

    expect(screen.getByLabelText(/repository url/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/branch\/tag/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/https:\/\/github.com/)).toBeInTheDocument();
  });

  it('should validate required fields', async () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const submitButton = screen.getByRole('button', { name: /deploy/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/application name is required/i)).toBeInTheDocument();
    });

    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it('should show resource configuration for stateless apps', () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const statelessOption = screen.getByText('Stateless Application');
    fireEvent.click(statelessOption);

    expect(screen.getByLabelText(/replicas/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/cpu request/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/memory request/i)).toBeInTheDocument();
  });

  it('should show volume configuration for stateful apps', () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const statefulOption = screen.getByText('Stateful Application');
    fireEvent.click(statefulOption.closest('div')!);

    expect(screen.getByText(/storage size/i)).toBeInTheDocument();
    expect(screen.getByText(/storage class/i)).toBeInTheDocument();
  });

  it('should show schedule field for cronjobs', () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const cronjobOption = screen.getByText('CronJob');
    fireEvent.click(cronjobOption);

    expect(screen.getByLabelText(/cron schedule/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/0 \* \* \* \*/)).toBeInTheDocument();
  });

  it('should show function configuration for serverless', () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const functionOption = screen.getByText('Serverless Function');
    fireEvent.click(functionOption.closest('div')!);

    expect(screen.getByText(/runtime/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/handler/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/timeout/i)).toBeInTheDocument();
  });

  it('should submit form with valid data for stateless app', async () => {
    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    // Fill form
    const nameInput = screen.getByLabelText(/application name/i);
    fireEvent.change(nameInput, { target: { value: 'Test App' } });

    const imageRadio = screen.getByLabelText(/container image/i);
    fireEvent.click(imageRadio);

    const imageInput = screen.getByLabelText(/image name/i);
    fireEvent.change(imageInput, { target: { value: 'nginx:latest' } });

    const replicasInput = screen.getByLabelText(/replicas/i);
    fireEvent.change(replicasInput, { target: { value: '3' } });

    const submitButton = screen.getByRole('button', { name: /deploy/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        name: 'Test App',
        type: 'stateless',
        source_type: 'image',
        source_image: 'nginx:latest',
        config: {
          replicas: 3,
          cpu: '100m',
          memory: '128Mi',
        },
      });
    });
  });

  it('should reset form when dialog closes', () => {
    const { rerender } = render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const nameInput = screen.getByLabelText(/application name/i) as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'Test App' } });
    expect(nameInput.value).toBe('Test App');

    rerender(
      <DeployApplicationDialog
        open={false}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    rerender(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const newNameInput = screen.getByLabelText(/application name/i) as HTMLInputElement;
    expect(newNameInput.value).toBe('');
  });

  it('should show loading state during submission', async () => {
    mockOnSubmit.mockImplementation(
      () => new Promise(resolve => setTimeout(resolve, 100))
    );

    render(
      <DeployApplicationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
      />
    );

    const nameInput = screen.getByLabelText(/application name/i);
    fireEvent.change(nameInput, { target: { value: 'Test App' } });

    const submitButton = screen.getByRole('button', { name: /deploy/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/deploying/i)).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.queryByText(/deploying/i)).not.toBeInTheDocument();
    });
  });
});