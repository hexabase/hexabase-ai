import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { WorkspacePage } from '../pages/WorkspacePage';
import { ProjectPage } from '../pages/ProjectPage';
import { setupMockAPI } from '../utils/mock-api';
import { testUsers } from '../fixtures/mock-data';
import { generateAppName, generateProjectName, expectNotification } from '../utils/test-helpers';

test.describe('Backup and Restore (Dedicated Workspaces)', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let workspacePage: WorkspacePage;
  let projectPage: ProjectPage;

  test.beforeEach(async ({ page }) => {
    await setupMockAPI(page);
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    workspacePage = new WorkspacePage(page);
    projectPage = new ProjectPage(page);
    
    // Login as admin (dedicated workspace features require admin privileges)
    await loginPage.goto();
    await loginPage.login(testUsers.admin.email, testUsers.admin.password);
    await loginPage.isLoggedIn();
    
    // Mock dedicated workspace with backup capability
    await page.route('**/api/organizations/*/workspaces/dedicated-*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'ws-dedicated-123',
          name: 'Dedicated Production',
          plan: 'dedicated',
          features: {
            backup: true,
            restore: true,
            nodes: 3,
            storage: '1TB',
          },
        }),
      });
    });
  });

  test('configure backup storage for dedicated workspace', async ({ page }) => {
    // Navigate to dedicated workspace
    await dashboardPage.openWorkspace('Dedicated Production');
    
    // Go to workspace settings
    const settingsButton = page.getByTestId('workspace-settings-button');
    await settingsButton.click();
    
    // Navigate to backup configuration
    const backupTab = page.getByRole('tab', { name: /backup/i });
    await backupTab.click();
    
    // Verify backup feature is available (only for dedicated plans)
    await expect(page.getByTestId('backup-feature-badge')).toContainText('Dedicated Plan Feature');
    
    // Configure backup storage
    const configureStorageButton = page.getByTestId('configure-backup-storage-button');
    await configureStorageButton.click();
    
    const storageDialog = page.getByRole('dialog');
    
    // Select Proxmox storage type
    await storageDialog.getByTestId('storage-type-proxmox').click();
    
    // Enter Proxmox configuration
    await storageDialog.getByTestId('proxmox-host-input').fill('backup.proxmox.local');
    await storageDialog.getByTestId('proxmox-storage-input').fill('backup-pool');
    await storageDialog.getByTestId('proxmox-username-input').fill('backup-user@pve');
    await storageDialog.getByTestId('proxmox-password-input').fill('secure-backup-pass');
    
    // Set storage capacity
    await storageDialog.getByTestId('storage-capacity-input').fill('500');
    await storageDialog.getByTestId('capacity-unit-select').selectOption('GB');
    
    // Test connection
    const testConnectionButton = storageDialog.getByTestId('test-connection-button');
    await testConnectionButton.click();
    
    await expect(page.getByText('Connection successful')).toBeVisible();
    
    // Save configuration
    await storageDialog.getByTestId('save-storage-button').click();
    
    await expectNotification(page, 'Backup storage configured successfully');
    
    // Verify storage is configured
    await expect(page.getByTestId('backup-storage-status')).toContainText('Connected');
    await expect(page.getByTestId('storage-capacity')).toContainText('500 GB');
  });

  test('create backup policy with schedule', async ({ page }) => {
    // Navigate to dedicated workspace
    await dashboardPage.openWorkspace('Dedicated Production');
    
    // Go to backup settings
    const settingsButton = page.getByTestId('workspace-settings-button');
    await settingsButton.click();
    const backupTab = page.getByRole('tab', { name: /backup/i });
    await backupTab.click();
    
    // Create backup policy
    const createPolicyButton = page.getByTestId('create-backup-policy-button');
    await createPolicyButton.click();
    
    const policyDialog = page.getByRole('dialog');
    
    // Configure policy details
    await policyDialog.getByTestId('policy-name-input').fill('Daily Production Backup');
    await policyDialog.getByTestId('policy-description-input').fill('Daily backup of all production workloads');
    
    // Select backup type
    await policyDialog.getByTestId('backup-type-select').selectOption('full');
    
    // Configure schedule
    await policyDialog.getByTestId('schedule-type-select').selectOption('daily');
    await policyDialog.getByTestId('schedule-time-input').fill('02:00');
    await policyDialog.getByTestId('timezone-select').selectOption('UTC');
    
    // Set retention policy
    await policyDialog.getByTestId('retention-days-input').fill('30');
    await policyDialog.getByTestId('retention-count-input').fill('30');
    
    // Configure backup scope
    await policyDialog.getByTestId('backup-scope-all').check();
    
    // Enable encryption
    await policyDialog.getByTestId('enable-encryption-toggle').check();
    await policyDialog.getByTestId('encryption-key-input').fill('workspace-encryption-key-123');
    
    // Enable compression
    await policyDialog.getByTestId('enable-compression-toggle').check();
    await policyDialog.getByTestId('compression-level-select').selectOption('high');
    
    // Save policy
    await policyDialog.getByTestId('save-policy-button').click();
    
    await expectNotification(page, 'Backup policy created successfully');
    
    // Verify policy in list
    const policyCard = page.getByTestId('backup-policy-Daily Production Backup');
    await expect(policyCard).toBeVisible();
    await expect(policyCard).toContainText('Daily at 02:00 UTC');
    await expect(policyCard).toContainText('30 days retention');
  });

  test('trigger manual backup of workspace', async ({ page }) => {
    const projectName = generateProjectName();
    const appName = generateAppName();
    
    // Navigate to dedicated workspace and create test data
    await dashboardPage.openWorkspace('Dedicated Production');
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
    
    // Deploy test application
    await projectPage.deployApplication({
      name: appName,
      type: 'stateless',
      image: 'nginx:latest',
      replicas: 2,
      port: 80,
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Go back to workspace level
    await page.getByTestId('breadcrumb-workspace').click();
    
    // Navigate to backup section
    const backupButton = page.getByTestId('workspace-backup-button');
    await backupButton.click();
    
    // Trigger manual backup
    const manualBackupButton = page.getByTestId('trigger-manual-backup-button');
    await manualBackupButton.click();
    
    const backupDialog = page.getByRole('dialog');
    
    // Configure backup
    await backupDialog.getByTestId('backup-name-input').fill('Pre-upgrade Backup');
    await backupDialog.getByTestId('backup-description-input').fill('Manual backup before system upgrade');
    await backupDialog.getByTestId('backup-type-select').selectOption('full');
    
    // Select what to backup
    await backupDialog.getByTestId('include-applications-checkbox').check();
    await backupDialog.getByTestId('include-databases-checkbox').check();
    await backupDialog.getByTestId('include-volumes-checkbox').check();
    await backupDialog.getByTestId('include-secrets-checkbox').check();
    
    // Start backup
    await backupDialog.getByTestId('start-backup-button').click();
    
    // Monitor backup progress
    await expect(page.getByText('Backup started')).toBeVisible();
    
    const progressBar = page.getByTestId('backup-progress');
    await expect(progressBar).toBeVisible();
    
    // Mock backup stages
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('backup-stage')).toContainText('Preparing workspace snapshot');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('backup-stage')).toContainText('Backing up applications');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('backup-stage')).toContainText('Backing up persistent volumes');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('backup-stage')).toContainText('Finalizing backup');
    
    // Backup complete
    await expectNotification(page, 'Backup completed successfully');
    
    // Verify backup in history
    const backupHistory = page.getByTestId('backup-history');
    const latestBackup = backupHistory.locator('[data-testid^="backup-item-"]').first();
    
    await expect(latestBackup).toContainText('Pre-upgrade Backup');
    await expect(latestBackup).toContainText('Full Backup');
    await expect(latestBackup.getByTestId('backup-size')).toBeVisible();
    await expect(latestBackup.getByTestId('backup-status')).toContainText('Completed');
  });

  test('restore workspace from backup', async ({ page }) => {
    // Navigate to dedicated workspace
    await dashboardPage.openWorkspace('Dedicated Production');
    
    // Go to backup section
    const backupButton = page.getByTestId('workspace-backup-button');
    await backupButton.click();
    
    // Mock existing backups
    await page.route('**/api/organizations/*/workspaces/*/backups', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          backups: [
            {
              id: 'backup-123',
              name: 'Daily Backup - 2025-01-12',
              type: 'full',
              size: '15.7 GB',
              created_at: new Date(Date.now() - 86400000).toISOString(),
              status: 'completed',
              encrypted: true,
              compressed: true,
            },
            {
              id: 'backup-456',
              name: 'Weekly Backup - 2025-01-07',
              type: 'full',
              size: '14.2 GB',
              created_at: new Date(Date.now() - 5 * 86400000).toISOString(),
              status: 'completed',
              encrypted: true,
              compressed: true,
            },
          ],
        }),
      });
    });
    
    await page.reload();
    
    // Select backup to restore
    const backupItem = page.locator('[data-testid="backup-item-backup-123"]');
    const restoreButton = backupItem.getByTestId('restore-backup-button');
    await restoreButton.click();
    
    const restoreDialog = page.getByRole('dialog');
    await expect(restoreDialog).toContainText('Restore from Backup');
    await expect(restoreDialog).toContainText('Daily Backup - 2025-01-12');
    
    // Configure restore options
    await restoreDialog.getByTestId('restore-target-select').selectOption('new-workspace');
    await restoreDialog.getByTestId('new-workspace-name-input').fill('Production-Restored');
    
    // Select what to restore
    await restoreDialog.getByTestId('restore-applications-checkbox').check();
    await restoreDialog.getByTestId('restore-databases-checkbox').check();
    await restoreDialog.getByTestId('restore-volumes-checkbox').check();
    await restoreDialog.getByTestId('restore-configs-checkbox').check();
    
    // Verify backup warning
    await expect(restoreDialog).toContainText('This will create a new workspace with data from the backup');
    
    // Start restore
    await restoreDialog.getByTestId('start-restore-button').click();
    
    // Confirm restore
    const confirmDialog = page.getByRole('dialog');
    await confirmDialog.getByTestId('confirm-restore-input').fill('RESTORE');
    await confirmDialog.getByTestId('confirm-restore-button').click();
    
    // Monitor restore progress
    await expect(page.getByText('Restore started')).toBeVisible();
    
    // Mock restore stages
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('restore-stage')).toContainText('Creating new workspace');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('restore-stage')).toContainText('Restoring applications');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('restore-stage')).toContainText('Restoring persistent volumes');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('restore-stage')).toContainText('Restoring configurations');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('restore-stage')).toContainText('Finalizing restore');
    
    await expectNotification(page, 'Restore completed successfully');
    
    // Verify redirect to restored workspace
    await expect(page).toHaveURL(/.*\/workspaces\/Production-Restored/);
  });

  test('backup individual applications', async ({ page }) => {
    const projectName = generateProjectName();
    const appName = generateAppName();
    
    // Create and deploy application
    await dashboardPage.openWorkspace('Dedicated Production');
    await workspacePage.createProject(projectName);
    await workspacePage.openProject(projectName);
    
    await projectPage.deployApplication({
      name: appName,
      type: 'stateful',
      image: 'postgres:14',
      port: 5432,
      storage: '50Gi',
      env: {
        'POSTGRES_DB': 'production_db',
        'POSTGRES_USER': 'admin',
        'POSTGRES_PASSWORD': 'secure123',
      },
    });
    
    await projectPage.waitForApplicationStatus(appName, 'running');
    
    // Navigate to application
    const appCard = await projectPage.getApplicationCard(appName);
    await appCard.click();
    
    // Go to backup tab
    const backupTab = page.getByRole('tab', { name: /backup/i });
    await backupTab.click();
    
    // Create application backup
    const createBackupButton = page.getByTestId('create-app-backup-button');
    await createBackupButton.click();
    
    const backupDialog = page.getByRole('dialog');
    
    await backupDialog.getByTestId('backup-name-input').fill(`${appName}-backup-${Date.now()}`);
    await backupDialog.getByTestId('backup-type-select').selectOption('application');
    
    // Include application data
    await backupDialog.getByTestId('include-app-state-checkbox').check();
    await backupDialog.getByTestId('include-pvc-data-checkbox').check();
    await backupDialog.getByTestId('include-configs-checkbox').check();
    
    await backupDialog.getByTestId('start-backup-button').click();
    
    // Monitor backup
    await expect(page.getByText('Application backup started')).toBeVisible();
    
    await page.waitForTimeout(3000);
    await expectNotification(page, 'Application backup completed');
    
    // Verify backup in application history
    const backupList = page.getByTestId('app-backup-list');
    await expect(backupList.locator('[data-testid^="app-backup-"]').first()).toBeVisible();
  });

  test('schedule automated backups with CronJob integration', async ({ page }) => {
    // Navigate to dedicated workspace
    await dashboardPage.openWorkspace('Dedicated Production');
    
    // Go to backup settings
    const settingsButton = page.getByTestId('workspace-settings-button');
    await settingsButton.click();
    const backupTab = page.getByRole('tab', { name: /backup/i });
    await backupTab.click();
    
    // Create scheduled backup job
    const scheduleBackupButton = page.getByTestId('schedule-backup-button');
    await scheduleBackupButton.click();
    
    const scheduleDialog = page.getByRole('dialog');
    
    // Configure backup job
    await scheduleDialog.getByTestId('job-name-input').fill('nightly-backup-job');
    await scheduleDialog.getByTestId('job-description-input').fill('Automated nightly backup of production workspace');
    
    // Set schedule (daily at 3 AM)
    await scheduleDialog.getByTestId('schedule-input').fill('0 3 * * *');
    await scheduleDialog.getByTestId('schedule-helper-select').selectOption('daily-3am');
    
    // Configure backup settings
    await scheduleDialog.getByTestId('backup-type-select').selectOption('incremental');
    await scheduleDialog.getByTestId('full-backup-frequency-select').selectOption('weekly');
    
    // Set retention
    await scheduleDialog.getByTestId('keep-daily-input').fill('7');
    await scheduleDialog.getByTestId('keep-weekly-input').fill('4');
    await scheduleDialog.getByTestId('keep-monthly-input').fill('6');
    
    // Enable notifications
    await scheduleDialog.getByTestId('notify-on-success-checkbox').check();
    await scheduleDialog.getByTestId('notify-on-failure-checkbox').check();
    await scheduleDialog.getByTestId('notification-email-input').fill('ops-team@example.com');
    
    // Create scheduled job
    await scheduleDialog.getByTestId('create-schedule-button').click();
    
    await expectNotification(page, 'Backup schedule created successfully');
    
    // Verify CronJob created
    await page.getByTestId('breadcrumb-project').click();
    await projectPage.cronJobsTab.click();
    
    const cronJobCard = page.getByTestId('cronjob-nightly-backup-job');
    await expect(cronJobCard).toBeVisible();
    await expect(cronJobCard).toContainText('0 3 * * *');
    await expect(cronJobCard.getByTestId('cronjob-status')).toContainText('Active');
  });

  test('test backup integrity verification', async ({ page }) => {
    // Navigate to dedicated workspace
    await dashboardPage.openWorkspace('Dedicated Production');
    
    // Go to backup section
    const backupButton = page.getByTestId('workspace-backup-button');
    await backupButton.click();
    
    // Select a backup to verify
    const backupItem = page.locator('[data-testid^="backup-item-"]').first();
    const verifyButton = backupItem.getByTestId('verify-backup-button');
    await verifyButton.click();
    
    const verifyDialog = page.getByRole('dialog');
    await expect(verifyDialog).toContainText('Verify Backup Integrity');
    
    // Configure verification
    await verifyDialog.getByTestId('verify-checksum-checkbox').check();
    await verifyDialog.getByTestId('verify-encryption-checkbox').check();
    await verifyDialog.getByTestId('verify-contents-checkbox').check();
    
    // Start verification
    await verifyDialog.getByTestId('start-verification-button').click();
    
    // Monitor verification progress
    await expect(page.getByText('Verification started')).toBeVisible();
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('verify-stage')).toContainText('Verifying checksums');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('verify-stage')).toContainText('Verifying encryption');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('verify-stage')).toContainText('Verifying backup contents');
    
    // Verification complete
    await expectNotification(page, 'Backup verification passed');
    
    // Check verification report
    const reportButton = page.getByTestId('view-verification-report-button');
    await reportButton.click();
    
    const reportDialog = page.getByRole('dialog');
    await expect(reportDialog).toContainText('Verification Report');
    await expect(reportDialog).toContainText('✓ Checksum verification: PASSED');
    await expect(reportDialog).toContainText('✓ Encryption verification: PASSED');
    await expect(reportDialog).toContainText('✓ Content verification: PASSED');
  });

  test('handle backup storage quota and cleanup', async ({ page }) => {
    // Navigate to dedicated workspace
    await dashboardPage.openWorkspace('Dedicated Production');
    
    // Go to backup settings
    const settingsButton = page.getByTestId('workspace-settings-button');
    await settingsButton.click();
    const backupTab = page.getByRole('tab', { name: /backup/i });
    await backupTab.click();
    
    // Check storage usage
    const storageCard = page.getByTestId('backup-storage-usage-card');
    await expect(storageCard).toBeVisible();
    await expect(storageCard).toContainText('Storage Usage');
    
    // Mock high storage usage
    await page.route('**/api/organizations/*/workspaces/*/backup-storage/usage', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          total_capacity: '500GB',
          used: '475GB',
          available: '25GB',
          percentage: 95,
          backups: [
            { id: 'old-1', name: 'Old Backup 1', size: '50GB', age_days: 45 },
            { id: 'old-2', name: 'Old Backup 2', size: '75GB', age_days: 60 },
            { id: 'old-3', name: 'Old Backup 3', size: '100GB', age_days: 90 },
          ],
        }),
      });
    });
    
    await page.reload();
    
    // Verify storage warning
    await expect(page.getByTestId('storage-warning')).toContainText('Storage usage is above 90%');
    
    // Run cleanup
    const cleanupButton = page.getByTestId('run-backup-cleanup-button');
    await cleanupButton.click();
    
    const cleanupDialog = page.getByRole('dialog');
    await expect(cleanupDialog).toContainText('Backup Cleanup');
    
    // Configure cleanup policy
    await cleanupDialog.getByTestId('cleanup-older-than-input').fill('30');
    await cleanupDialog.getByTestId('keep-minimum-input').fill('5');
    await cleanupDialog.getByTestId('cleanup-type-select').selectOption('incremental-first');
    
    // Preview cleanup
    const previewButton = cleanupDialog.getByTestId('preview-cleanup-button');
    await previewButton.click();
    
    await expect(cleanupDialog).toContainText('3 backups will be removed');
    await expect(cleanupDialog).toContainText('225GB will be freed');
    
    // Execute cleanup
    await cleanupDialog.getByTestId('execute-cleanup-button').click();
    
    // Confirm cleanup
    const confirmDialog = page.getByRole('dialog');
    await confirmDialog.getByTestId('confirm-cleanup-input').fill('DELETE');
    await confirmDialog.getByTestId('confirm-cleanup-button').click();
    
    // Monitor cleanup
    await expect(page.getByText('Cleanup in progress')).toBeVisible();
    
    await page.waitForTimeout(3000);
    await expectNotification(page, 'Backup cleanup completed. 225GB freed');
    
    // Verify updated storage usage
    await expect(storageCard).toContainText('250GB / 500GB');
    await expect(storageCard).toContainText('50%');
  });

  test('backup with disaster recovery test', async ({ page }) => {
    // Navigate to dedicated workspace
    await dashboardPage.openWorkspace('Dedicated Production');
    
    // Go to backup section
    const backupButton = page.getByTestId('workspace-backup-button');
    await backupButton.click();
    
    // Start disaster recovery test
    const drTestButton = page.getByTestId('disaster-recovery-test-button');
    await drTestButton.click();
    
    const drDialog = page.getByRole('dialog');
    await expect(drDialog).toContainText('Disaster Recovery Test');
    
    // Select backup for DR test
    await drDialog.getByTestId('select-backup-dropdown').click();
    await page.getByTestId('backup-option-latest').click();
    
    // Configure test environment
    await drDialog.getByTestId('test-workspace-name-input').fill('DR-Test-Environment');
    await drDialog.getByTestId('test-duration-select').selectOption('2h');
    
    // Enable validation tests
    await drDialog.getByTestId('validate-applications-checkbox').check();
    await drDialog.getByTestId('validate-connectivity-checkbox').check();
    await drDialog.getByTestId('validate-data-integrity-checkbox').check();
    
    // Start DR test
    await drDialog.getByTestId('start-dr-test-button').click();
    
    // Monitor DR test progress
    await expect(page.getByText('Disaster recovery test started')).toBeVisible();
    
    const drProgress = page.getByTestId('dr-test-progress');
    await expect(drProgress).toBeVisible();
    
    // Mock DR test stages
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('dr-stage')).toContainText('Creating test environment');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('dr-stage')).toContainText('Restoring backup to test environment');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('dr-stage')).toContainText('Validating applications');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('dr-stage')).toContainText('Running connectivity tests');
    
    await page.waitForTimeout(2000);
    await expect(page.getByTestId('dr-stage')).toContainText('Verifying data integrity');
    
    // DR test complete
    await expectNotification(page, 'Disaster recovery test completed successfully');
    
    // View DR test report
    const viewReportButton = page.getByTestId('view-dr-report-button');
    await viewReportButton.click();
    
    const reportDialog = page.getByRole('dialog');
    await expect(reportDialog).toContainText('DR Test Report');
    await expect(reportDialog).toContainText('✓ All applications restored successfully');
    await expect(reportDialog).toContainText('✓ Network connectivity verified');
    await expect(reportDialog).toContainText('✓ Data integrity checks passed');
    await expect(reportDialog).toContainText('Recovery Time: 12 minutes');
    await expect(reportDialog).toContainText('Recovery Point: < 5 minutes');
  });
});