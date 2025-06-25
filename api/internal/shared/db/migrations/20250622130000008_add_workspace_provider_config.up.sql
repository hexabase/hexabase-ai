-- Create workspace provider configurations table
CREATE TABLE IF NOT EXISTS workspace_provider_configs (
    workspace_id TEXT PRIMARY KEY,
    provider_type TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_workspace
        FOREIGN KEY (workspace_id) 
        REFERENCES workspaces(id) 
        ON DELETE CASCADE
);

-- Create index for faster lookups
CREATE INDEX idx_workspace_provider_configs_provider_type ON workspace_provider_configs(provider_type);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_workspace_provider_configs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_workspace_provider_configs_updated_at
    BEFORE UPDATE ON workspace_provider_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_workspace_provider_configs_updated_at();

-- Add comment
COMMENT ON TABLE workspace_provider_configs IS 'Stores FaaS provider configuration per workspace';
COMMENT ON COLUMN workspace_provider_configs.provider_type IS 'Type of FaaS provider (fission, knative, etc.)';
COMMENT ON COLUMN workspace_provider_configs.config IS 'Provider-specific configuration as JSON';