-- Remove salt column from sessions table
-- This rollback migration removes the salt column added for refresh token hashing

ALTER TABLE sessions DROP COLUMN salt;