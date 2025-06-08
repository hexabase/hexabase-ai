-- Drop indexes first
DROP INDEX IF EXISTS idx_model_usage_workspace_model_time;
DROP INDEX IF EXISTS idx_model_usage_timestamp;
DROP INDEX IF EXISTS idx_model_usage_model_name;
DROP INDEX IF EXISTS idx_model_usage_session_id;
DROP INDEX IF EXISTS idx_model_usage_user_id;
DROP INDEX IF EXISTS idx_model_usage_workspace_id;

DROP INDEX IF EXISTS idx_chat_sessions_updated_at;
DROP INDEX IF EXISTS idx_chat_sessions_created_at;
DROP INDEX IF EXISTS idx_chat_sessions_user_id;
DROP INDEX IF EXISTS idx_chat_sessions_workspace_id;

-- Drop tables
DROP TABLE IF EXISTS model_usage;
DROP TABLE IF EXISTS chat_sessions;