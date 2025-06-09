-- Drop triggers
DROP TRIGGER IF EXISTS update_backup_storages_updated_at ON backup_storages;
DROP TRIGGER IF EXISTS update_backup_policies_updated_at ON backup_policies;

-- Drop function
DROP FUNCTION IF EXISTS update_backup_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_backup_storages_workspace_id;
DROP INDEX IF EXISTS idx_backup_storages_status;
DROP INDEX IF EXISTS idx_backup_policies_application_id;
DROP INDEX IF EXISTS idx_backup_policies_storage_id;
DROP INDEX IF EXISTS idx_backup_policies_enabled;
DROP INDEX IF EXISTS idx_backup_executions_policy_id;
DROP INDEX IF EXISTS idx_backup_executions_status;
DROP INDEX IF EXISTS idx_backup_executions_started_at;
DROP INDEX IF EXISTS idx_backup_restores_backup_execution_id;
DROP INDEX IF EXISTS idx_backup_restores_application_id;
DROP INDEX IF EXISTS idx_backup_restores_status;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS backup_restores;
DROP TABLE IF EXISTS backup_executions;
DROP TABLE IF EXISTS backup_policies;
DROP TABLE IF EXISTS backup_storages;