-- +migrate Up
-- Create refresh_token_blacklists table
CREATE TABLE refresh_token_blacklists (
    id TEXT NOT NULL,
    token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT refresh_token_blacklists_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX idx_refresh_token_blacklists_token ON refresh_token_blacklists (token);
CREATE INDEX idx_refresh_token_blacklists_expires_at ON refresh_token_blacklists (expires_at);

COMMENT ON TABLE refresh_token_blacklists IS 'Stores blacklisted refresh tokens to prevent reuse.';
COMMENT ON COLUMN refresh_token_blacklists.id IS 'Unique identifier for the blacklist entry.';
COMMENT ON COLUMN refresh_token_blacklists.token IS 'The blacklisted refresh token (hashed or original, depending on security implementation).';
COMMENT ON COLUMN refresh_token_blacklists.expires_at IS 'Timestamp when this blacklist entry (and the token) expires and can be cleaned up.';
COMMENT ON COLUMN refresh_token_blacklists.created_at IS 'Timestamp when the token was added to the blacklist.';
