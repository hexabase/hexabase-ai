-- Drop trigger
DROP TRIGGER IF EXISTS update_function_versions_updated_at_trigger ON function_versions;
DROP FUNCTION IF EXISTS update_function_versions_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_function_versions_app_active;
DROP INDEX IF EXISTS idx_function_invocations_app_time;
DROP INDEX IF EXISTS idx_function_invocations_duration;
DROP INDEX IF EXISTS idx_function_events_status;
DROP INDEX IF EXISTS idx_function_events_app_type;

-- Drop tables
DROP TABLE IF EXISTS function_events;
DROP TABLE IF EXISTS function_invocations;
DROP TABLE IF EXISTS function_versions;

-- Remove function-specific columns from applications table
ALTER TABLE applications 
DROP COLUMN IF EXISTS function_runtime,
DROP COLUMN IF EXISTS function_handler,
DROP COLUMN IF EXISTS function_timeout,
DROP COLUMN IF EXISTS function_memory,
DROP COLUMN IF EXISTS function_trigger_type,
DROP COLUMN IF EXISTS function_trigger_config,
DROP COLUMN IF EXISTS function_env_vars,
DROP COLUMN IF EXISTS function_secrets;