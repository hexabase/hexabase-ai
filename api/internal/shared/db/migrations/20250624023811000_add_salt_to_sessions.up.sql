-- Add salt column to sessions table for secure refresh token hashing
-- This migration supports the implementation of salted SHA-256 hashing for refresh tokens
-- to eliminate the security risk of storing plain text tokens in the database

ALTER TABLE sessions ADD COLUMN salt VARCHAR(64) NOT NULL DEFAULT '';

-- Note: The default empty string allows for gradual migration of existing sessions
-- New sessions will have proper salt values generated by the application