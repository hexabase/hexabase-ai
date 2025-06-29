-- Add refresh_token_selector column to sessions table for O(1) lookup
-- This supports the selector.verifier pattern implemented in the auth service

ALTER TABLE sessions ADD COLUMN refresh_token_selector VARCHAR(64);

-- Create index for fast lookup by selector
CREATE INDEX idx_sessions_refresh_token_selector ON sessions(refresh_token_selector);

-- Add comment for documentation
COMMENT ON COLUMN sessions.refresh_token_selector IS 'Selector part of refresh token for O(1) session lookup';
