-- Backup storage configuration
CREATE TABLE IF NOT EXISTS backup_storages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('nfs', 'ceph', 'local', 'proxmox')),
    proxmox_storage_id VARCHAR(255) NOT NULL,
    proxmox_node_id VARCHAR(255) NOT NULL,
    capacity_gb INTEGER NOT NULL,
    used_gb INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'creating', 'active', 'failed', 'deleting')),
    connection_config JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workspace_id, name)
);

-- Backup policies for applications
CREATE TABLE IF NOT EXISTS backup_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    storage_id UUID NOT NULL REFERENCES backup_storages(id),
    enabled BOOLEAN DEFAULT true,
    schedule VARCHAR(100) NOT NULL, -- Cron expression
    retention_days INTEGER DEFAULT 30,
    backup_type VARCHAR(50) DEFAULT 'full' CHECK (backup_type IN ('full', 'incremental')),
    include_volumes BOOLEAN DEFAULT true,
    include_database BOOLEAN DEFAULT true,
    include_config BOOLEAN DEFAULT true,
    compression_enabled BOOLEAN DEFAULT true,
    encryption_enabled BOOLEAN DEFAULT true,
    encryption_key_ref VARCHAR(255), -- Reference to K8s secret
    pre_backup_hook TEXT,
    post_backup_hook TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(application_id)
);

-- Backup executions
CREATE TABLE IF NOT EXISTS backup_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES backup_policies(id) ON DELETE CASCADE,
    cronjob_execution_id UUID REFERENCES cronjob_executions(id),
    status VARCHAR(50) NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'succeeded', 'failed', 'cancelled')),
    size_bytes BIGINT,
    compressed_size_bytes BIGINT,
    backup_path TEXT,
    backup_manifest JSONB, -- Contains list of backed up resources
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Backup restore operations
CREATE TABLE IF NOT EXISTS backup_restores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    backup_execution_id UUID NOT NULL REFERENCES backup_executions(id),
    application_id TEXT NOT NULL REFERENCES applications(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'preparing', 'restoring', 'verifying', 'completed', 'failed')),
    restore_type VARCHAR(50) NOT NULL CHECK (restore_type IN ('full', 'selective')),
    restore_options JSONB, -- Options like target namespace, selective resources, etc.
    new_application_id TEXT REFERENCES applications(id), -- If restoring to new app
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    validation_results JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_backup_storages_workspace_id ON backup_storages(workspace_id);
CREATE INDEX idx_backup_storages_status ON backup_storages(status);
CREATE INDEX idx_backup_policies_application_id ON backup_policies(application_id);
CREATE INDEX idx_backup_policies_storage_id ON backup_policies(storage_id);
CREATE INDEX idx_backup_policies_enabled ON backup_policies(enabled);
CREATE INDEX idx_backup_executions_policy_id ON backup_executions(policy_id);
CREATE INDEX idx_backup_executions_status ON backup_executions(status);
CREATE INDEX idx_backup_executions_started_at ON backup_executions(started_at);
CREATE INDEX idx_backup_restores_backup_execution_id ON backup_restores(backup_execution_id);
CREATE INDEX idx_backup_restores_application_id ON backup_restores(application_id);
CREATE INDEX idx_backup_restores_status ON backup_restores(status);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_backup_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER update_backup_storages_updated_at
    BEFORE UPDATE ON backup_storages
    FOR EACH ROW
    EXECUTE FUNCTION update_backup_updated_at();

CREATE TRIGGER update_backup_policies_updated_at
    BEFORE UPDATE ON backup_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_backup_updated_at();