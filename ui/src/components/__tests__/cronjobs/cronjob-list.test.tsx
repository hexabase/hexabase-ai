import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CronJobList } from '@/components/cronjobs/cronjob-list';
import { useRouter, useParams } from 'next/navigation';
import { apiClient } from '@/lib/api-client';

jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
  useParams: jest.fn(),
}));

jest.mock('@/lib/api-client', () => ({
  apiClient: {
    applications: {
      list: jest.fn(),
      updateStatus: jest.fn(),
      triggerCronJob: jest.fn(),
      updateCronSchedule: jest.fn(),
      getCronJobExecutions: jest.fn(),
    },
  },
}));

describe('CronJobList', () => {
  const mockPush = jest.fn();
  const mockCronJobs = [
    {
      id: 'cron-1',
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: 'Daily Backup',
      type: 'cronjob',
      status: 'active',
      source_type: 'image',
      source_image: 'backup:latest',
      cron_schedule: '0 2 * * *',
      last_execution_at: '2024-01-01T02:00:00Z',
      next_execution_at: '2024-01-02T02:00:00Z',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    },
    {
      id: 'cron-2',
      workspace_id: 'ws-1',
      project_id: 'proj-1',
      name: 'Hourly Sync',
      type: 'cronjob',
      status: 'suspended',
      source_type: 'image',
      source_image: 'sync:latest',
      cron_schedule: '0 * * * *',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
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
    (apiClient.applications.list as jest.Mock).mockResolvedValue({
      data: {
        applications: mockCronJobs,
        total: 2,
      },
    });
    (apiClient.applications.getCronJobExecutions as jest.Mock).mockResolvedValue({
      data: {
        executions: [],
        total: 0,
      },
    });
  });

  it('should display list of cronjobs', async () => {
    render(<CronJobList />);

    await waitFor(() => {
      expect(screen.getByText('Daily Backup')).toBeInTheDocument();
      expect(screen.getByText('Hourly Sync')).toBeInTheDocument();
    });
  });

  it('should show cron schedules', async () => {
    render(<CronJobList />);

    await waitFor(() => {
      expect(screen.getByText('0 2 * * *')).toBeInTheDocument();
      expect(screen.getByText('0 * * * *')).toBeInTheDocument();
    });
  });

  it('should show last and next execution times', async () => {
    render(<CronJobList />);

    await waitFor(() => {
      expect(screen.getByText(/last run:/i)).toBeInTheDocument();
      expect(screen.getByText(/next run:/i)).toBeInTheDocument();
    });
  });

  it('should trigger manual execution', async () => {
    (apiClient.applications.triggerCronJob as jest.Mock).mockResolvedValue({
      data: {
        execution_id: 'cje-123',
        job_name: 'daily-backup-12345',
        status: 'running',
        message: 'CronJob triggered successfully',
      },
    });

    render(<CronJobList />);

    await waitFor(() => {
      const triggerButton = screen.getByTestId('trigger-cron-1');
      fireEvent.click(triggerButton);
    });

    await waitFor(() => {
      expect(apiClient.applications.triggerCronJob).toHaveBeenCalledWith(
        'org-1', 'ws-1', 'cron-1'
      );
    });
  });

  it('should toggle cronjob status', async () => {
    (apiClient.applications.updateStatus as jest.Mock).mockResolvedValue({
      data: { ...mockCronJobs[0], status: 'suspended' },
    });

    render(<CronJobList />);

    await waitFor(() => {
      const toggleButton = screen.getByTestId('toggle-cron-1');
      fireEvent.click(toggleButton);
    });

    await waitFor(() => {
      expect(apiClient.applications.updateStatus).toHaveBeenCalledWith(
        'org-1', 'ws-1', 'cron-1',
        { status: 'suspended' }
      );
    });
  });

  it('should open schedule editor', async () => {
    render(<CronJobList />);

    await waitFor(() => {
      const editScheduleButton = screen.getByTestId('edit-schedule-cron-1');
      fireEvent.click(editScheduleButton);
    });

    expect(screen.getByRole('heading', { name: /edit schedule/i })).toBeInTheDocument();
  });

  it('should update cron schedule', async () => {
    (apiClient.applications.updateCronSchedule as jest.Mock).mockResolvedValue({
      data: {
        id: 'cron-1',
        cron_schedule: '*/30 * * * *',
        next_execution_at: '2024-01-01T00:30:00Z',
      },
    });

    render(<CronJobList />);

    await waitFor(() => {
      const editScheduleButton = screen.getByTestId('edit-schedule-cron-1');
      fireEvent.click(editScheduleButton);
    });

    // Simulate schedule update
    const saveButton = screen.getByRole('button', { name: /save/i });
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(apiClient.applications.updateCronSchedule).toHaveBeenCalled();
    });
  });

  it('should show execution history', async () => {
    const mockExecutions = [
      {
        id: 'cje-1',
        application_id: 'cron-1',
        job_name: 'daily-backup-12345',
        status: 'succeeded',
        started_at: '2024-01-01T02:00:00Z',
        completed_at: '2024-01-01T02:05:00Z',
        created_at: '2024-01-01T02:00:00Z',
        updated_at: '2024-01-01T02:05:00Z',
      },
    ];

    (apiClient.applications.getCronJobExecutions as jest.Mock).mockResolvedValue({
      data: {
        executions: mockExecutions,
        total: 1,
        page: 1,
        page_size: 10,
      },
    });

    render(<CronJobList />);

    await waitFor(() => {
      const viewHistoryButton = screen.getByTestId('view-history-cron-1');
      fireEvent.click(viewHistoryButton);
    });

    await waitFor(() => {
      expect(screen.getByText('daily-backup-12345')).toBeInTheDocument();
    });
  });

  it('should filter cronjobs by status', async () => {
    render(<CronJobList />);

    await waitFor(() => {
      expect(screen.getByText('Daily Backup')).toBeInTheDocument();
    });

    // Clear the initial API calls
    (apiClient.applications.list as jest.Mock).mockClear();

    const statusFilter = screen.getByTestId('status-filter');
    fireEvent.click(statusFilter);
    
    const activeOption = screen.getByText('Active');
    fireEvent.click(activeOption);

    await waitFor(() => {
      expect(apiClient.applications.list).toHaveBeenCalledWith(
        'org-1', 'ws-1', 'proj-1',
        expect.objectContaining({ status: 'active', type: 'cronjob' })
      );
    });
  });

  it('should show empty state', async () => {
    (apiClient.applications.list as jest.Mock).mockResolvedValue({
      data: {
        applications: [],
        total: 0,
      },
    });

    render(<CronJobList />);

    await waitFor(() => {
      expect(screen.getByText(/no cronjobs found/i)).toBeInTheDocument();
    });
  });

  it('should disable actions for suspended cronjobs', async () => {
    render(<CronJobList />);

    await waitFor(() => {
      const triggerButton = screen.getByTestId('trigger-cron-2');
      expect(triggerButton).toBeDisabled();
    });
  });
});