-- Remove refresh_token_selector column and its index

DROP INDEX IF EXISTS idx_sessions_refresh_token_selector;
ALTER TABLE sessions DROP COLUMN IF EXISTS refresh_token_selector;
