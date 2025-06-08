-- Create chat_sessions table for storing AI chat conversations
CREATE TABLE IF NOT EXISTS chat_sessions (
    id VARCHAR(36) PRIMARY KEY,
    workspace_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    model VARCHAR(100) NOT NULL,
    messages JSONB NOT NULL DEFAULT '[]',
    context INTEGER[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for chat_sessions
CREATE INDEX idx_chat_sessions_workspace_id ON chat_sessions(workspace_id);
CREATE INDEX idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX idx_chat_sessions_created_at ON chat_sessions(created_at);
CREATE INDEX idx_chat_sessions_updated_at ON chat_sessions(updated_at DESC);

-- Create model_usage table for tracking LLM usage
CREATE TABLE IF NOT EXISTS model_usage (
    id VARCHAR(36) PRIMARY KEY,
    workspace_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(36),
    model_name VARCHAR(100) NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    request_duration_ms BIGINT NOT NULL DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for model_usage
CREATE INDEX idx_model_usage_workspace_id ON model_usage(workspace_id);
CREATE INDEX idx_model_usage_user_id ON model_usage(user_id);
CREATE INDEX idx_model_usage_session_id ON model_usage(session_id);
CREATE INDEX idx_model_usage_model_name ON model_usage(model_name);
CREATE INDEX idx_model_usage_timestamp ON model_usage(timestamp DESC);
CREATE INDEX idx_model_usage_workspace_model_time ON model_usage(workspace_id, model_name, timestamp);

-- Add foreign key constraints if workspace and user tables exist
-- These are commented out as they depend on the existence of other tables
-- ALTER TABLE chat_sessions ADD CONSTRAINT fk_chat_sessions_workspace
--     FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
-- ALTER TABLE model_usage ADD CONSTRAINT fk_model_usage_workspace
--     FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
-- ALTER TABLE model_usage ADD CONSTRAINT fk_model_usage_session
--     FOREIGN KEY (session_id) REFERENCES chat_sessions(id) ON DELETE SET NULL;