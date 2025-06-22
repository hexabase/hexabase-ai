-- Drop foreign key constraints from applications table

ALTER TABLE applications DROP CONSTRAINT IF EXISTS fk_applications_template;
ALTER TABLE applications DROP CONSTRAINT IF EXISTS fk_applications_project;
ALTER TABLE applications DROP CONSTRAINT IF EXISTS fk_applications_workspace; 