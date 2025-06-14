import React from 'react';
import { render, screen, within, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
// Next.js uses its own routing, no need for BrowserRouter
import HomePage from '@/app/page';
import { AuthProvider } from '@/lib/auth-context';
import { mockApiClient } from '@/test-utils/mock-api-client';
import {
  loginUser,
  createOrganization,
  createWorkspace,
  createProject,
  deployApplication,
  setupBackupStorage,
  createBackupPolicy,
  verifyResourceStatus,
  expectNotification,
  expectNoErrors,
  generateProjectName,
  generateAppName,
  delay,
} from './test-helpers';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
  }),
  usePathname: () => '/',
}));

describe('Backup and Restore Scenario', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    
    // Reset all mocks
    jest.clearAllMocks();
  });

  const renderApp = () => {
    return render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <HomePage />
        </AuthProvider>
      </QueryClientProvider>
    );
  };

  it('should create backup storage, configure policies, and perform backup/restore', async () => {
    renderApp();

    // Initial setup
    await loginUser();
    await createOrganization('Backup Test Organization');
    await createWorkspace('Production Workspace', 'dedicated'); // Backups only available on dedicated
    const projectName = generateProjectName();
    await createProject(projectName);

    // Step 1: Deploy applications to backup
    console.log('Step 1: Deploying applications to backup...');
    const postgresApp = generateAppName();
    await deployApplication(postgresApp, 'stateful', {
      image: 'postgres:14',
      port: 5432,
      storage: '20Gi',
    });
    await verifyResourceStatus(postgresApp, 'running');

    const redisApp = generateAppName();
    await deployApplication(redisApp, 'stateful', {
      image: 'redis:7',
      port: 6379,
      storage: '5Gi',
    });
    await verifyResourceStatus(redisApp, 'running');

    // Step 2: Setup backup storage
    console.log('Step 2: Setting up backup storage...');
    await setupBackupStorage('Primary Backup Storage', 'proxmox');
    
    // Verify storage was created
    expect(mockApiClient.backup.createStorage).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        name: 'Primary Backup Storage',
        type: 'proxmox',
      })
    );
    await expectNotification(/backup storage created/i);

    // Step 3: Create backup policies
    console.log('Step 3: Creating backup policies...');
    
    // Daily backup for PostgreSQL
    await createBackupPolicy(postgresApp, '0 2 * * *', 30);
    expect(mockApiClient.backup.createPolicy).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        schedule: '0 2 * * *',
        retention_days: 30,
      })
    );

    // Weekly backup for Redis
    await createBackupPolicy(redisApp, '0 3 * * 0', 90);
    expect(mockApiClient.backup.createPolicy).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.objectContaining({
        schedule: '0 3 * * 0',
        retention_days: 90,
      })
    );

    // Step 4: Manually trigger backup
    console.log('Step 4: Manually triggering backup...');
    const backupTab = screen.getByRole('tab', { name: /backup/i });
    await userEvent.click(backupTab);

    // Find PostgreSQL backup policy and trigger it
    const policyRow = screen.getByText(postgresApp).closest('tr');
    if (!policyRow) throw new Error('Policy row not found');
    
    const triggerButton = within(policyRow).getByRole('button', { name: /trigger backup/i });
    await userEvent.click(triggerButton);

    // Confirm backup
    const confirmModal = screen.getByRole('dialog');
    const confirmButton = within(confirmModal).getByRole('button', { name: /start backup/i });
    await userEvent.click(confirmButton);

    await expectNotification(/backup started/i);

    // Step 5: Monitor backup progress
    console.log('Step 5: Monitoring backup progress...');
    const executionsTab = screen.getByRole('tab', { name: /executions/i });
    await userEvent.click(executionsTab);

    // Wait for backup to complete
    await waitFor(() => {
      const executionRow = screen.getByText(/be-manual/i).closest('tr');
      const status = within(executionRow!).getByTestId('execution-status');
      expect(status).toHaveTextContent(/completed/i);
    }, { timeout: 20000 });

    await expectNotification(/backup completed successfully/i);

    // Verify backup size and details
    const viewButton = screen.getByRole('button', { name: /view details/i });
    await userEvent.click(viewButton);

    const detailsModal = screen.getByRole('dialog');
    expect(within(detailsModal).getByText(/size.*500 mb/i)).toBeInTheDocument();
    expect(within(detailsModal).getByText(/includes.*applications/i)).toBeInTheDocument();
    expect(within(detailsModal).getByText(/includes.*configurations/i)).toBeInTheDocument();
    expect(within(detailsModal).getByText(/includes.*volumes/i)).toBeInTheDocument();

    const closeButton = within(detailsModal).getByRole('button', { name: /close/i });
    await userEvent.click(closeButton);

    // Step 6: Simulate data corruption and restore
    console.log('Step 6: Simulating restore scenario...');
    
    // Navigate to applications tab
    const appsTab = screen.getByRole('tab', { name: /applications/i });
    await userEvent.click(appsTab);

    // Simulate PostgreSQL failure
    const postgresRow = screen.getByText(postgresApp).closest('tr');
    expect(within(postgresRow!).getByTestId('resource-status')).toHaveTextContent(/running/i);

    // Go back to backup tab for restore
    await userEvent.click(backupTab);
    await userEvent.click(executionsTab);

    // Initiate restore
    const restoreButton = screen.getByRole('button', { name: /restore/i });
    await userEvent.click(restoreButton);

    const restoreModal = screen.getByRole('dialog');
    
    // Select restore options
    const restoreTypeSelect = within(restoreModal).getByLabelText(/restore type/i);
    await userEvent.selectOptions(restoreTypeSelect, 'full');

    const targetSelect = within(restoreModal).getByLabelText(/restore to/i);
    await userEvent.selectOptions(targetSelect, 'same-location');

    const confirmRestoreButton = within(restoreModal).getByRole('button', { name: /restore/i });
    await userEvent.click(confirmRestoreButton);

    // Confirm restore operation
    const warningModal = screen.getByRole('dialog', { name: /confirm restore/i });
    const warningText = within(warningModal).getByText(/this will overwrite existing data/i);
    expect(warningText).toBeInTheDocument();

    const finalConfirmButton = within(warningModal).getByRole('button', { name: /confirm/i });
    await userEvent.click(finalConfirmButton);

    await expectNotification(/restore initiated/i);

    // Monitor restore progress
    await waitFor(() => {
      const restoreStatus = screen.getByTestId('restore-status');
      expect(restoreStatus).toHaveTextContent(/in progress/i);
    });

    // Wait for restore completion
    await delay(3000);
    await waitFor(() => {
      const restoreStatus = screen.getByTestId('restore-status');
      expect(restoreStatus).toHaveTextContent(/completed/i);
    }, { timeout: 15000 });

    await expectNotification(/restore completed successfully/i);

    // Step 7: Verify restored applications
    console.log('Step 7: Verifying restored applications...');
    await userEvent.click(appsTab);

    await verifyResourceStatus(postgresApp, 'running');
    await verifyResourceStatus(redisApp, 'running');

    // Check restored data integrity
    const postgresRow2 = screen.getByText(postgresApp).closest('tr');
    const integrityBadge = within(postgresRow2!).getByTestId('data-integrity');
    expect(integrityBadge).toHaveTextContent(/verified/i);

    console.log('✅ Backup and restore scenario completed successfully!');
  });

  it('should handle backup policy scheduling and retention', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Policy Test Org');
    await createWorkspace('Backup Testing', 'dedicated');
    await createProject('backup-project');

    // Deploy test application
    const appName = generateAppName();
    await deployApplication(appName, 'stateful', {
      image: 'mysql:8',
      port: 3306,
      storage: '10Gi',
    });

    // Setup backup storage
    await setupBackupStorage('Test Storage', 'proxmox');

    // Create multiple backup policies with different schedules
    console.log('Creating backup policies with different schedules...');
    
    // Hourly backups with 1-day retention
    await createBackupPolicy(appName, '0 * * * *', 1);
    
    // Navigate to policies tab to create another
    const policyTab = screen.getByRole('tab', { name: /policies/i });
    await userEvent.click(policyTab);
    
    // Daily backups with 7-day retention  
    const createPolicyButton = screen.getByRole('button', { name: /create policy/i });
    await userEvent.click(createPolicyButton);
    
    const modal = screen.getByRole('dialog');
    const appSelect = within(modal).getByLabelText(/application/i);
    await userEvent.selectOptions(appSelect, appName);
    
    const scheduleInput = within(modal).getByLabelText(/schedule/i);
    await userEvent.type(scheduleInput, '0 0 * * *');
    
    const retentionInput = within(modal).getByLabelText(/retention days/i);
    await userEvent.clear(retentionInput);
    await userEvent.type(retentionInput, '7');
    
    const typeSelect = within(modal).getByLabelText(/backup type/i);
    await userEvent.selectOptions(typeSelect, 'incremental');
    
    const createButton = within(modal).getByRole('button', { name: /create/i });
    await userEvent.click(createButton);

    // Verify policies are listed
    await waitFor(() => {
      const policies = screen.getAllByTestId('backup-policy-row');
      expect(policies).toHaveLength(2);
    });

    // Test policy editing
    console.log('Testing policy editing...');
    const editButton = screen.getAllByRole('button', { name: /edit/i })[0];
    await userEvent.click(editButton);

    const editModal = screen.getByRole('dialog');
    const editRetentionInput = within(editModal).getByLabelText(/retention days/i);
    await userEvent.clear(editRetentionInput);
    await userEvent.type(editRetentionInput, '3');

    const saveButton = within(editModal).getByRole('button', { name: /save/i });
    await userEvent.click(saveButton);

    await expectNotification(/policy updated/i);

    // Test policy disable/enable
    console.log('Testing policy enable/disable...');
    const toggleSwitch = screen.getAllByRole('switch', { name: /enable policy/i })[0];
    await userEvent.click(toggleSwitch);

    await expectNotification(/policy disabled/i);

    // Re-enable
    await userEvent.click(toggleSwitch);
    await expectNotification(/policy enabled/i);

    // Verify next execution times
    const policyRows = screen.getAllByTestId('backup-policy-row');
    expect(within(policyRows[0]).getByText(/next run.*in 1 hour/i)).toBeInTheDocument();
    expect(within(policyRows[1]).getByText(/next run.*tomorrow/i)).toBeInTheDocument();
  });

  it('should support different backup types and storage options', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Storage Types Org');
    await createWorkspace('Multi-Storage', 'dedicated');
    await createProject('storage-test-project');

    // Step 1: Setup multiple storage backends
    console.log('Setting up multiple storage backends...');
    
    // Proxmox storage
    await setupBackupStorage('Proxmox Primary', 'proxmox');
    
    // S3-compatible storage
    const backupTab = screen.getByRole('tab', { name: /backup/i });
    await userEvent.click(backupTab);
    
    const addStorageButton = screen.getByRole('button', { name: /add storage/i });
    await userEvent.click(addStorageButton);
    
    const modal = screen.getByRole('dialog');
    const nameInput = within(modal).getByLabelText(/storage name/i);
    await userEvent.type(nameInput, 'S3 Compatible Storage');
    
    const typeSelect = within(modal).getByLabelText(/storage type/i);
    await userEvent.selectOptions(typeSelect, 's3');
    
    const endpointInput = within(modal).getByLabelText(/s3 endpoint/i);
    await userEvent.type(endpointInput, 'https://s3.example.com');
    
    const bucketInput = within(modal).getByLabelText(/bucket name/i);
    await userEvent.type(bucketInput, 'hexabase-backups');
    
    const accessKeyInput = within(modal).getByLabelText(/access key/i);
    await userEvent.type(accessKeyInput, 'test-access-key');
    
    const secretKeyInput = within(modal).getByLabelText(/secret key/i);
    await userEvent.type(secretKeyInput, 'test-secret-key');
    
    const createButton = within(modal).getByRole('button', { name: /create/i });
    await userEvent.click(createButton);

    // Verify both storages are listed
    await waitFor(() => {
      expect(screen.getByText('Proxmox Primary')).toBeInTheDocument();
      expect(screen.getByText('S3 Compatible Storage')).toBeInTheDocument();
    });

    // Step 2: Test different backup types
    console.log('Testing different backup types...');
    
    // Deploy applications
    const dbApp = generateAppName();
    await deployApplication(dbApp, 'stateful', {
      image: 'postgres:14',
      port: 5432,
      storage: '10Gi',
    });

    // Create full backup policy
    const policyTab = screen.getByRole('tab', { name: /policies/i });
    await userEvent.click(policyTab);
    
    const createPolicyButton = screen.getByRole('button', { name: /create policy/i });
    await userEvent.click(createPolicyButton);
    
    const policyModal = screen.getByRole('dialog');
    
    const appSelect = within(policyModal).getByLabelText(/application/i);
    await userEvent.selectOptions(appSelect, dbApp);
    
    const storageSelect = within(policyModal).getByLabelText(/backup storage/i);
    await userEvent.selectOptions(storageSelect, 'Proxmox Primary');
    
    const backupTypeSelect = within(policyModal).getByLabelText(/backup type/i);
    await userEvent.selectOptions(backupTypeSelect, 'full');
    
    const scheduleInput = within(policyModal).getByLabelText(/schedule/i);
    await userEvent.type(scheduleInput, '0 0 * * 0'); // Weekly
    
    const compressionCheckbox = within(policyModal).getByLabelText(/enable compression/i);
    await userEvent.click(compressionCheckbox);
    
    const encryptionCheckbox = within(policyModal).getByLabelText(/enable encryption/i);
    await userEvent.click(encryptionCheckbox);
    
    const createPolicyBtn = within(policyModal).getByRole('button', { name: /create/i });
    await userEvent.click(createPolicyBtn);

    // Create incremental backup policy
    await userEvent.click(createPolicyButton);
    
    const incrModal = screen.getByRole('dialog');
    
    const incrAppSelect = within(incrModal).getByLabelText(/application/i);
    await userEvent.selectOptions(incrAppSelect, dbApp);
    
    const incrStorageSelect = within(incrModal).getByLabelText(/backup storage/i);
    await userEvent.selectOptions(incrStorageSelect, 'S3 Compatible Storage');
    
    const incrTypeSelect = within(incrModal).getByLabelText(/backup type/i);
    await userEvent.selectOptions(incrTypeSelect, 'incremental');
    
    const incrScheduleInput = within(incrModal).getByLabelText(/schedule/i);
    await userEvent.type(incrScheduleInput, '0 */6 * * *'); // Every 6 hours
    
    const incrCreateBtn = within(incrModal).getByRole('button', { name: /create/i });
    await userEvent.click(incrCreateBtn);

    // Verify backup policies show correct types and features
    const policyRows = screen.getAllByTestId('backup-policy-row');
    expect(policyRows).toHaveLength(2);
    
    expect(within(policyRows[0]).getByText(/full backup/i)).toBeInTheDocument();
    expect(within(policyRows[0]).getByText(/compressed/i)).toBeInTheDocument();
    expect(within(policyRows[0]).getByText(/encrypted/i)).toBeInTheDocument();
    
    expect(within(policyRows[1]).getByText(/incremental backup/i)).toBeInTheDocument();

    console.log('✅ Multiple backup types and storage options tested successfully!');
  });

  it('should handle backup failures and retry logic', async () => {
    renderApp();

    await loginUser();
    await createOrganization('Failure Test Org');
    await createWorkspace('Failure Testing', 'dedicated');
    await createProject('failure-project');

    // Deploy application
    const appName = generateAppName();
    await deployApplication(appName, 'stateful', {
      image: 'mongodb:6',
      port: 27017,
      storage: '15Gi',
    });

    // Setup backup
    await setupBackupStorage('Test Storage', 'proxmox');
    await createBackupPolicy(appName, '*/30 * * * *', 7); // Every 30 minutes

    // Simulate backup failure
    console.log('Simulating backup failure...');
    mockApiClient.backup.triggerBackup.mockRejectedValueOnce(
      new Error('Storage quota exceeded')
    );

    const backupTab = screen.getByRole('tab', { name: /backup/i });
    await userEvent.click(backupTab);

    const policyRow = screen.getByText(appName).closest('tr');
    const triggerButton = within(policyRow!).getByRole('button', { name: /trigger backup/i });
    await userEvent.click(triggerButton);

    const confirmButton = screen.getByRole('button', { name: /start backup/i });
    await userEvent.click(confirmButton);

    await expectNotification(/backup failed.*storage quota exceeded/i);

    // Check failed execution in history
    const executionsTab = screen.getByRole('tab', { name: /executions/i });
    await userEvent.click(executionsTab);

    await waitFor(() => {
      const failedExecution = screen.getByText(/failed/i).closest('tr');
      expect(failedExecution).toBeInTheDocument();
      expect(within(failedExecution!).getByText(/storage quota exceeded/i)).toBeInTheDocument();
    });

    // Test retry functionality
    console.log('Testing retry functionality...');
    const retryButton = screen.getByRole('button', { name: /retry/i });
    await userEvent.click(retryButton);

    await expectNotification(/backup retry initiated/i);

    // This time it should succeed
    await waitFor(() => {
      const successExecution = screen.getAllByTestId('execution-status')[0];
      expect(successExecution).toHaveTextContent(/completed/i);
    }, { timeout: 10000 });

    console.log('✅ Backup failure and retry handling tested successfully!');
  });
});