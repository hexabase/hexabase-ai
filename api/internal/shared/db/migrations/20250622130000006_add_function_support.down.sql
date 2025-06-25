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

-- Note: Function-specific tables are dropped in 009_create_functions_tables.down.sql