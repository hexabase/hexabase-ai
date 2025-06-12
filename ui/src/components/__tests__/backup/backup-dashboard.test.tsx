import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BackupDashboard } from '@/components/backup/backup-dashboard';
import { backupApi } from '@/lib/api-client';
import { useParams } from 'next/navigation';

jest.mock('next/navigation', () => ({
  useParams: jest.fn(),
}));

jest.mock('@/lib/api-client', () => ({
  backupApi: {
    listBackupStorages: jest.fn(),
    createBackupStorage: jest.fn(),
    deleteBackupStorage: jest.fn(),
    listBackupPolicies: jest.fn(),
    createBackupPolicy: jest.fn(),
    updateBackupPolicy: jest.fn(),
    deleteBackupPolicy: jest.fn(),
    listBackupExecutions: jest.fn(),
    executeBackupPolicy: jest.fn(),
    validateBackup: jest.fn(),
    restoreBackup: jest.fn(),
  },
}));

describe('BackupDashboard', () => {
  const user = userEvent.setup();
  
  const mockStorages = [
    {
      id: 'bs-1',
      workspace_id: 'ws-123',
      name: 'Primary Backup Storage',
      type: 'proxmox',
      status: 'active',
      capacity_gb: 1000,
      used_gb: 250,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
  ];
  
  const mockPolicies = [
    {
      id: 'bp-1',
      workspace_id: 'ws-123',
      name: 'Daily Backup',
      storage_id: 'bs-1',
      schedule: '0 2 * * *',
      retention_days: 30,
      backup_type: 'full',
      enabled: true,
      last_execution: '2024-01-01T02:00:00Z',
      next_execution: '2024-01-02T02:00:00Z',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
  ];
  
  const mockExecutions = [
    {
      id: 'be-1',
      policy_id: 'bp-1',
      status: 'completed',
      started_at: '2024-01-01T02:00:00Z',
      completed_at: '2024-01-01T02:30:00Z',
      size_bytes: 1024 * 1024 * 500, // 500MB
      error: null,
    },
    {
      id: 'be-2',
      policy_id: 'bp-1',
      status: 'failed',
      started_at: '2024-01-02T02:00:00Z',
      completed_at: '2024-01-02T02:05:00Z',
      size_bytes: 0,
      error: 'Storage quota exceeded',
    },
  ];

  beforeEach(() => {
    jest.clearAllMocks();
    (useParams as jest.Mock).mockReturnValue({
      orgId: 'org-123',
      workspaceId: 'ws-123',
    });
    
    (backupApi.listBackupStorages as jest.Mock).mockResolvedValue({
      data: mockStorages,
    });
    
    (backupApi.listBackupPolicies as jest.Mock).mockResolvedValue({
      data: { policies: mockPolicies },
    });
    
    (backupApi.listBackupExecutions as jest.Mock).mockResolvedValue({
      data: { executions: mockExecutions },
    });
  });

  it('should display backup dashboard with all sections', async () => {
    render(<BackupDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText(/backup management/i)).toBeInTheDocument();
      expect(screen.getByText(/backup storage/i)).toBeInTheDocument();
      expect(screen.getByText(/backup policies/i)).toBeInTheDocument();
      expect(screen.getByText(/recent backups/i)).toBeInTheDocument();
    });
  });

  it('should show backup storage list', async () => {
    render(<BackupDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('Primary Backup Storage')).toBeInTheDocument();
      expect(screen.getByText(/proxmox/i)).toBeInTheDocument();
      expect(screen.getByText(/250 GB \/ 1000 GB/i)).toBeInTheDocument();
    });
  });

  it('should create new backup storage', async () => {
    (backupApi.createBackupStorage as jest.Mock).mockResolvedValue({
      id: 'bs-new',
      workspace_id: 'ws-123',
      name: 'Secondary Storage',
      type: 'proxmox',
      status: 'active',
      capacity_gb: 500,
      used_gb: 0,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    });
    
    render(<BackupDashboard />);
    
    const addStorageButton = screen.getByRole('button', { name: /add backup storage/i });
    await user.click(addStorageButton);
    
    // Fill form
    const nameInput = screen.getByLabelText(/storage name/i);
    await user.type(nameInput, 'Secondary Storage');
    
    const capacityInput = screen.getByLabelText(/capacity/i);
    await user.type(capacityInput, '500');
    
    const submitButton = screen.getByRole('button', { name: /create storage/i });
    await user.click(submitButton);
    
    await waitFor(() => {
      expect(backupApi.createBackupStorage).toHaveBeenCalledWith('org-123', 'ws-123', {
        name: 'Secondary Storage',
        type: 'proxmox',
        capacity_gb: 500,
      });
    });
  });

  it('should display backup policies', async () => {
    render(<BackupDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('Daily Backup')).toBeInTheDocument();
      expect(screen.getByText('0 2 * * *')).toBeInTheDocument();
      expect(screen.getByText(/30 days retention/i)).toBeInTheDocument();
    });
  });

  it('should create new backup policy', async () => {
    (backupApi.createBackupPolicy as jest.Mock).mockResolvedValue({
      id: 'bp-new',
      workspace_id: 'ws-123',
      name: 'Weekly Backup',
      storage_id: 'bs-1',
      schedule: '0 3 * * 0',
      retention_days: 90,
      backup_type: 'incremental',
      enabled: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    });
    
    render(<BackupDashboard />);
    
    const createPolicyButton = screen.getByRole('button', { name: /create backup policy/i });
    await user.click(createPolicyButton);
    
    // Fill form
    const nameInput = screen.getByLabelText(/policy name/i);
    await user.type(nameInput, 'Weekly Backup');
    
    const scheduleInput = screen.getByLabelText(/schedule/i);
    await user.type(scheduleInput, '0 3 * * 0');
    
    const retentionInput = screen.getByLabelText(/retention days/i);
    await user.clear(retentionInput);
    await user.type(retentionInput, '90');
    
    const typeSelect = screen.getByLabelText(/backup type/i);
    await user.selectOptions(typeSelect, 'incremental');
    
    const submitButton = screen.getByRole('button', { name: /create policy/i });
    await user.click(submitButton);
    
    await waitFor(() => {
      expect(backupApi.createBackupPolicy).toHaveBeenCalledWith('org-123', 'ws-123', 'app-1', {
        name: 'Weekly Backup',
        storage_id: 'bs-1',
        schedule: '0 3 * * 0',
        retention_days: 90,
        backup_type: 'incremental',
        enabled: true,
      });
    });
  });

  it('should show backup execution history', async () => {
    render(<BackupDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText(/completed/i)).toBeInTheDocument();
      expect(screen.getByText(/500 mb/i)).toBeInTheDocument();
      expect(screen.getByText(/failed/i)).toBeInTheDocument();
      expect(screen.getByText(/storage quota exceeded/i)).toBeInTheDocument();
    });
  });

  it('should trigger manual backup', async () => {
    (backupApi.executeBackupPolicy as jest.Mock).mockResolvedValue({
      execution_id: 'be-manual',
      status: 'running',
      message: 'Backup started successfully',
    });
    
    render(<BackupDashboard />);
    
    await waitFor(() => {
      const policyCard = screen.getByTestId('policy-bp-1');
      const runButton = within(policyCard).getByRole('button', { name: /run now/i });
      fireEvent.click(runButton);
    });
    
    await waitFor(() => {
      expect(backupApi.executeBackupPolicy).toHaveBeenCalledWith('org-123', 'ws-123', 'bp-1');
      expect(screen.getByText(/backup started successfully/i)).toBeInTheDocument();
    });
  });

  it('should restore from backup', async () => {
    (backupApi.validateBackup as jest.Mock).mockResolvedValue({
      backup: {
        id: 'be-1',
        policy_name: 'Daily Backup',
        size_bytes: 1024 * 1024 * 500,
        created_at: '2024-01-01T02:30:00Z',
        includes: ['applications', 'configurations', 'volumes'],
      },
    });
    
    (backupApi.restoreBackup as jest.Mock).mockResolvedValue({
      restore_id: 'res-123',
      status: 'in_progress',
      message: 'Restore initiated',
    });
    
    render(<BackupDashboard />);
    
    await waitFor(() => {
      const restoreButton = screen.getByTestId('restore-be-1');
      fireEvent.click(restoreButton);
    });
    
    // Confirm restore
    await waitFor(() => {
      expect(screen.getByText(/restore from backup/i)).toBeInTheDocument();
      expect(screen.getByText(/500 mb/i)).toBeInTheDocument();
    });
    
    const confirmButton = screen.getByRole('button', { name: /confirm restore/i });
    await user.click(confirmButton);
    
    await waitFor(() => {
      expect(backupApi.restoreBackup).toHaveBeenCalledWith('org-123', 'ws-123', 'app-1', {
        backup_execution_id: 'be-1',
        restore_type: 'in_place',
        restore_options: {
          restore_volumes: true,
          restore_database: true,
          restore_config: true,
        },
      });
      expect(screen.getByText(/restore initiated/i)).toBeInTheDocument();
    });
  });

  it('should handle backup policy toggle', async () => {
    (backupApi.updateBackupPolicy as jest.Mock).mockResolvedValue({
      ...mockPolicies[0],
      enabled: false,
    });
    
    render(<BackupDashboard />);
    
    await waitFor(() => {
      const toggleSwitch = screen.getByTestId('policy-toggle-bp-1');
      fireEvent.click(toggleSwitch);
    });
    
    await waitFor(() => {
      expect(backupApi.updateBackupPolicy).toHaveBeenCalledWith('org-123', 'ws-123', 'bp-1', {
        enabled: false,
      });
    });
  });

  it('should delete backup storage with confirmation', async () => {
    (backupApi.deleteBackupStorage as jest.Mock).mockResolvedValue({});
    
    render(<BackupDashboard />);
    
    await waitFor(() => {
      const deleteButton = screen.getByTestId('delete-storage-bs-1');
      fireEvent.click(deleteButton);
    });
    
    // Confirm deletion
    expect(screen.getByText(/are you sure you want to delete/i)).toBeInTheDocument();
    
    const confirmButton = screen.getByRole('button', { name: /delete storage/i });
    await user.click(confirmButton);
    
    await waitFor(() => {
      expect(backupApi.deleteBackupStorage).toHaveBeenCalledWith('org-123', 'ws-123', 'bs-1');
    });
  });

  it('should show storage usage chart', async () => {
    render(<BackupDashboard />);
    
    await waitFor(() => {
      expect(screen.getByTestId('storage-usage-chart')).toBeInTheDocument();
      expect(screen.getByText(/25% used/i)).toBeInTheDocument();
    });
  });

  it('should validate cron schedule', async () => {
    render(<BackupDashboard />);
    
    const createPolicyButton = screen.getByRole('button', { name: /create backup policy/i });
    await user.click(createPolicyButton);
    
    const scheduleInput = screen.getByLabelText(/schedule/i);
    await user.type(scheduleInput, 'invalid-cron');
    
    // Try to submit
    const submitButton = screen.getByRole('button', { name: /create policy/i });
    await user.click(submitButton);
    
    expect(screen.getByText(/invalid cron expression/i)).toBeInTheDocument();
  });
});
