-- Revert: Rename code_challenge column back to code_verifier
ALTER TABLE auth_states 
RENAME COLUMN code_challenge TO code_verifier;