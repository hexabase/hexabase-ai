-- Add foreign key constraints to applications table
-- This must be done after workspaces and projects tables are created

ALTER TABLE applications ADD CONSTRAINT fk_applications_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
ALTER TABLE applications ADD CONSTRAINT fk_applications_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE applications ADD CONSTRAINT fk_applications_template FOREIGN KEY (template_app_id) REFERENCES applications(id); 