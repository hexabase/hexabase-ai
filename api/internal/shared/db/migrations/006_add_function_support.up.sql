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

-- Create function_versions table for version management
CREATE TABLE IF NOT EXISTS function_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    source_code TEXT, -- base64 encoded source code or S3 reference
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN ('inline', 's3', 'git')),
    source_url TEXT, -- for S3 or git sources
    build_logs TEXT,
    build_status VARCHAR(50) DEFAULT 'pending' CHECK (build_status IN ('pending', 'building', 'success', 'failed')),
    image_uri TEXT, -- container image URI after build
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deployed_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(application_id, version_number)
);

-- Create function_invocations table for tracking function calls
CREATE TABLE IF NOT EXISTS function_invocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    version_id UUID REFERENCES function_versions(id),
    invocation_id VARCHAR(255) NOT NULL UNIQUE, -- Knative request ID
    trigger_source VARCHAR(50) NOT NULL, -- http, event, schedule
    request_method VARCHAR(10), -- for HTTP triggers
    request_path TEXT, -- for HTTP triggers
    request_headers JSONB, -- for HTTP triggers
    request_body TEXT, -- base64 encoded
    response_status INTEGER,
    response_headers JSONB,
    response_body TEXT, -- base64 encoded
    error_message TEXT,
    duration_ms INTEGER, -- execution duration in milliseconds
    cold_start BOOLEAN DEFAULT false,
    memory_used INTEGER, -- in MB
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create function_events table for event-driven functions
CREATE TABLE IF NOT EXISTS function_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    event_type VARCHAR(255) NOT NULL, -- e.g., "s3.object.created", "kafka.message", "webhook.github"
    event_source VARCHAR(255) NOT NULL, -- source identifier
    event_data JSONB NOT NULL,
    processing_status VARCHAR(50) DEFAULT 'pending' CHECK (processing_status IN ('pending', 'processing', 'success', 'failed', 'retry')),
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    invocation_id UUID REFERENCES function_invocations(id),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_function_versions_app_active ON function_versions(application_id, is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_function_invocations_app_time ON function_invocations(application_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_function_invocations_duration ON function_invocations(duration_ms) WHERE duration_ms IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_function_events_status ON function_events(processing_status, created_at) WHERE processing_status IN ('pending', 'retry');
CREATE INDEX IF NOT EXISTS idx_function_events_app_type ON function_events(application_id, event_type);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_function_versions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_function_versions_updated_at_trigger
BEFORE UPDATE ON function_versions
FOR EACH ROW
EXECUTE FUNCTION update_function_versions_updated_at();