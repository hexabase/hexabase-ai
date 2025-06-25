-- Add function-specific columns to applications table
ALTER TABLE applications 
ADD COLUMN IF NOT EXISTS function_runtime VARCHAR(50) CHECK (function_runtime IN ('python3.9', 'python3.10', 'python3.11', 'nodejs16', 'nodejs18', 'go1.20', 'go1.21')),
ADD COLUMN IF NOT EXISTS function_handler VARCHAR(255), -- e.g., "main.handler" for Python
ADD COLUMN IF NOT EXISTS function_timeout INTEGER DEFAULT 300, -- timeout in seconds
ADD COLUMN IF NOT EXISTS function_memory INTEGER DEFAULT 256, -- memory in MB
ADD COLUMN IF NOT EXISTS function_trigger_type VARCHAR(50) DEFAULT 'http' CHECK (function_trigger_type IN ('http', 'event', 'schedule')),
ADD COLUMN IF NOT EXISTS function_trigger_config JSONB, -- trigger-specific configuration
ADD COLUMN IF NOT EXISTS function_env_vars JSONB, -- function-specific environment variables
ADD COLUMN IF NOT EXISTS function_secrets JSONB; -- encrypted secret references

-- Note: Function-specific tables (function_versions, function_invocations, function_events) 
-- are now created in 009_create_functions_tables.up.sql with the new architecture
-- that uses independent functions table instead of extending applications table