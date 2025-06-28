-- Drop triggers
DROP TRIGGER IF EXISTS update_function_triggers_updated_at ON function_triggers;
DROP TRIGGER IF EXISTS update_functions_updated_at ON functions;

-- Drop functions
DROP FUNCTION IF EXISTS update_function_triggers_updated_at();
DROP FUNCTION IF EXISTS update_functions_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_functions_workspace_project;
DROP INDEX IF EXISTS idx_functions_status;
DROP INDEX IF EXISTS idx_function_versions_function;
DROP INDEX IF EXISTS idx_function_versions_active;
DROP INDEX IF EXISTS idx_function_triggers_function;
DROP INDEX IF EXISTS idx_function_triggers_enabled;
DROP INDEX IF EXISTS idx_function_invocations_function;
DROP INDEX IF EXISTS idx_function_invocations_status;
DROP INDEX IF EXISTS idx_function_events_function;

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS function_events;
DROP TABLE IF EXISTS function_invocations;
DROP TABLE IF EXISTS function_triggers;
DROP TABLE IF EXISTS function_versions;
DROP TABLE IF EXISTS functions; 