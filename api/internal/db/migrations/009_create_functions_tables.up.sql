-- Create functions table
CREATE TABLE IF NOT EXISTS functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL,
    runtime VARCHAR(50) NOT NULL,
    handler VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    active_version UUID,
    labels JSONB DEFAULT '{}',
    annotations JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_workspace
        FOREIGN KEY (workspace_id) 
        REFERENCES workspaces(id) 
        ON DELETE CASCADE,
        
    CONSTRAINT unique_function_name_per_project
        UNIQUE (workspace_id, project_id, name)
);

-- Create function_versions table
CREATE TABLE IF NOT EXISTS function_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id TEXT NOT NULL,
    function_id UUID NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    version INTEGER NOT NULL,
    runtime VARCHAR(50),
    handler VARCHAR(255),
    image TEXT,
    source_code TEXT,
    build_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    build_log TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT false,
    
    CONSTRAINT fk_function
        FOREIGN KEY (function_id) 
        REFERENCES functions(id) 
        ON DELETE CASCADE,
        
    CONSTRAINT unique_version_per_function
        UNIQUE (function_id, version)
);

-- Create function_triggers table
CREATE TABLE IF NOT EXISTS function_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id TEXT NOT NULL,
    function_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_function_trigger
        FOREIGN KEY (function_id) 
        REFERENCES functions(id) 
        ON DELETE CASCADE,
        
    CONSTRAINT unique_trigger_name_per_function
        UNIQUE (function_id, name)
);

-- Create function_invocations table
CREATE TABLE IF NOT EXISTS function_invocations (
    invocation_id VARCHAR(255) PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    function_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    result JSONB,
    error TEXT,
    
    CONSTRAINT fk_function_invocation
        FOREIGN KEY (function_id) 
        REFERENCES functions(id) 
        ON DELETE CASCADE
);

-- Create function_events table for audit trail
CREATE TABLE IF NOT EXISTS function_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id TEXT NOT NULL,
    function_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    user_id TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_function_event
        FOREIGN KEY (function_id) 
        REFERENCES functions(id) 
        ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_functions_workspace_project ON functions(workspace_id, project_id);
CREATE INDEX idx_functions_status ON functions(status);
CREATE INDEX idx_function_versions_function ON function_versions(function_id);
CREATE INDEX idx_function_versions_active ON function_versions(function_id, is_active) WHERE is_active = true;
CREATE INDEX idx_function_triggers_function ON function_triggers(function_id);
CREATE INDEX idx_function_triggers_enabled ON function_triggers(function_id, enabled) WHERE enabled = true;
CREATE INDEX idx_function_invocations_function ON function_invocations(function_id, started_at DESC);
CREATE INDEX idx_function_invocations_status ON function_invocations(status);
CREATE INDEX idx_function_events_function ON function_events(function_id, created_at DESC);

-- Add trigger to update updated_at timestamp for functions
CREATE OR REPLACE FUNCTION update_functions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_functions_updated_at
    BEFORE UPDATE ON functions
    FOR EACH ROW
    EXECUTE FUNCTION update_functions_updated_at();

-- Add trigger to update updated_at timestamp for function_triggers
CREATE OR REPLACE FUNCTION update_function_triggers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_function_triggers_updated_at
    BEFORE UPDATE ON function_triggers
    FOR EACH ROW
    EXECUTE FUNCTION update_function_triggers_updated_at();

-- Add comments
COMMENT ON TABLE functions IS 'Stores function definitions';
COMMENT ON TABLE function_versions IS 'Stores function version history';
COMMENT ON TABLE function_triggers IS 'Stores function triggers (HTTP, Event, Schedule, etc.)';
COMMENT ON TABLE function_invocations IS 'Tracks function invocations';
COMMENT ON TABLE function_events IS 'Audit trail for function-related events';