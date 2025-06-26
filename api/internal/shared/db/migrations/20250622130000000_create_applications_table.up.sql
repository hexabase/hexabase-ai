-- Create applications table (base structure)
-- This is required by later migrations that add CronJob and Function support

CREATE TABLE applications (
    id text NOT NULL,
    workspace_id text NOT NULL,
    project_id text NOT NULL,
    name text NOT NULL,
    type text NOT NULL DEFAULT 'stateless',
    status text NOT NULL DEFAULT 'pending',
    source_type text NOT NULL,
    source_image text,
    source_git_url text,
    source_git_ref text,
    config jsonb,
    endpoints jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

ALTER TABLE applications ADD CONSTRAINT applications_pkey PRIMARY KEY (id);
CREATE INDEX idx_applications_workspace_id ON applications(workspace_id);
CREATE INDEX idx_applications_project_id ON applications(project_id); 