-- Drop trigger and function
DROP TRIGGER IF EXISTS update_next_execution_at_trigger ON applications;
DROP FUNCTION IF EXISTS update_next_execution_at();

-- Drop cronjob executions table
DROP TABLE IF EXISTS cronjob_executions;

-- Remove cronjob-related columns from applications
ALTER TABLE applications 
DROP COLUMN IF EXISTS type,
DROP COLUMN IF EXISTS cron_schedule,
DROP COLUMN IF EXISTS cron_command,
DROP COLUMN IF EXISTS cron_args,
DROP COLUMN IF EXISTS template_app_id,
DROP COLUMN IF EXISTS last_execution_at,
DROP COLUMN IF EXISTS next_execution_at;