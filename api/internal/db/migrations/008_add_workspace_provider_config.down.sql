-- Drop trigger
DROP TRIGGER IF EXISTS update_workspace_provider_configs_updated_at ON workspace_provider_configs;

-- Drop function
DROP FUNCTION IF EXISTS update_workspace_provider_configs_updated_at();

-- Drop table
DROP TABLE IF EXISTS workspace_provider_configs;