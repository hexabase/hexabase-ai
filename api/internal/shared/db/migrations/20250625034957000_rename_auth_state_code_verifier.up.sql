-- Rename code_verifier column to code_challenge to match RFC 7636 terminology
ALTER TABLE auth_states 
RENAME COLUMN code_verifier TO code_challenge;