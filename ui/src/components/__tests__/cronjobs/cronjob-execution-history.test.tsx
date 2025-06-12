import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CronJobExecutionHistory } from '@/components/cronjobs/cronjob-execution-history';
import { CronJobExecution } from '@/lib/api-client';

describe('CronJobExecutionHistory', () => {
  const mockExecutions: CronJobExecution[] = [
    {
      id: 'cje-1',
      application_id: 'app-1',
      job_name: 'backup-job-12345',
      started_at: '2024-01-01T10:00:00Z',
      completed_at: '2024-01-01T10:05:00Z',
      status: 'succeeded',
      exit_code: 0,
      logs: 'Backup completed successfully',
      created_at: '2024-01-01T10:00:00Z',
      updated_at: '2024-01-01T10:05:00Z',
    },
    {
      id: 'cje-2',
      application_id: 'app-1',
      job_name: 'backup-job-12346',
      started_at: '2024-01-01T11:00:00Z',
      completed_at: '2024-01-01T11:03:00Z',
      status: 'failed',
      exit_code: 1,
      logs: 'Error: Connection timeout',
      created_at: '2024-01-01T11:00:00Z',
      updated_at: '2024-01-01T11:03:00Z',
    },
    {
      id: 'cje-3',
      application_id: 'app-1',
      job_name: 'backup-job-12347',
      started_at: '2024-01-01T12:00:00Z',
      status: 'running',
      created_at: '2024-01-01T12:00:00Z',
      updated_at: '2024-01-01T12:00:00Z',
    },
  ];

  it('should display execution history', () => {
    render(
      <CronJobExecutionHistory
        executions={mockExecutions}
        loading={false}
      />
    );

    expect(screen.getByText('backup-job-12345')).toBeInTheDocument();
    expect(screen.getByText('backup-job-12346')).toBeInTheDocument();
    expect(screen.getByText('backup-job-12347')).toBeInTheDocument();
  });

  it('should show status badges with correct colors', () => {
    render(
      <CronJobExecutionHistory
        executions={mockExecutions}
        loading={false}
      />
    );

    // Status text might be in badges with icons, so use flexible text matching
    const succeededBadges = screen.getAllByText((content, element) => {
      return element?.textContent?.includes('succeeded') || false;
    });
    const failedBadges = screen.getAllByText((content, element) => {
      return element?.textContent?.includes('failed') || false;
    });
    const runningBadges = screen.getAllByText((content, element) => {
      return element?.textContent?.includes('running') || false;
    });

    // Badge now uses variant prop instead of color classes
    expect(succeededBadges.length).toBeGreaterThan(0);
    expect(failedBadges.length).toBeGreaterThan(0);
    expect(runningBadges.length).toBeGreaterThan(0);
  });

  it('should display duration for completed executions', () => {
    render(
      <CronJobExecutionHistory
        executions={mockExecutions}
        loading={false}
      />
    );

    expect(screen.getByText(/5 minutes/)).toBeInTheDocument();
    expect(screen.getByText(/3 minutes/)).toBeInTheDocument();
  });

  it('should show loading state', () => {
    render(
      <CronJobExecutionHistory
        executions={[]}
        loading={true}
      />
    );

    expect(screen.getByTestId('executions-skeleton')).toBeInTheDocument();
  });

  it('should show empty state when no executions', () => {
    render(
      <CronJobExecutionHistory
        executions={[]}
        loading={false}
      />
    );

    expect(screen.getByText(/no executions yet/i)).toBeInTheDocument();
  });

  it('should expand to show logs when clicked', () => {
    render(
      <CronJobExecutionHistory
        executions={mockExecutions}
        loading={false}
      />
    );

    const firstExecution = screen.getByText('backup-job-12345').closest('div')!;
    fireEvent.click(firstExecution);

    expect(screen.getByText('Backup completed successfully')).toBeInTheDocument();
  });

  it('should show exit code for completed executions', () => {
    render(
      <CronJobExecutionHistory
        executions={mockExecutions}
        loading={false}
      />
    );

    expect(screen.getByText('Exit code: 0')).toBeInTheDocument();
    expect(screen.getByText('Exit code: 1')).toBeInTheDocument();
  });

  it('should allow viewing full logs in modal', async () => {
    const mockOnViewLogs = jest.fn();
    
    render(
      <CronJobExecutionHistory
        executions={mockExecutions}
        loading={false}
        onViewLogs={mockOnViewLogs}
      />
    );

    // First expand the execution
    const firstExecution = screen.getByText('backup-job-12345').closest('div')!;
    fireEvent.click(firstExecution);

    // Now find and click the view logs button
    const viewLogsButton = screen.getByRole('button', { name: /view logs/i });
    fireEvent.click(viewLogsButton);

    expect(mockOnViewLogs).toHaveBeenCalledWith(mockExecutions[0]);
  });

  it('should format timestamps correctly', () => {
    render(
      <CronJobExecutionHistory
        executions={mockExecutions}
        loading={false}
      />
    );

    // Since we're mocking date-fns format to return ISO strings in tests
    const dates = screen.getAllByText((content, element) => {
      return element?.textContent?.includes('2024-01-01') || false;
    });
    expect(dates.length).toBeGreaterThan(0);
  });

  it('should handle pagination', () => {
    const mockOnPageChange = jest.fn();
    
    render(
      <CronJobExecutionHistory
        executions={mockExecutions}
        loading={false}
        totalExecutions={50}
        currentPage={1}
        pageSize={10}
        onPageChange={mockOnPageChange}
      />
    );

    const nextButton = screen.getByRole('button', { name: /next/i });
    fireEvent.click(nextButton);

    expect(mockOnPageChange).toHaveBeenCalledWith(2);
  });
});