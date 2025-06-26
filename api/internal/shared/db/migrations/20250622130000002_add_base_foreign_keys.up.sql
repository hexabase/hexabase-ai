-- Add foreign key constraints to base tables

ALTER TABLE workspaces ADD CONSTRAINT fk_workspaces_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE workspaces ADD CONSTRAINT fk_workspaces_plan FOREIGN KEY (plan_id) REFERENCES plans(id);
ALTER TABLE projects ADD CONSTRAINT fk_projects_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
ALTER TABLE projects ADD CONSTRAINT fk_projects_child_projects FOREIGN KEY (parent_project_id) REFERENCES projects(id);

-- Create indexes for projects
CREATE INDEX idx_projects_workspace_id ON projects(workspace_id);
CREATE INDEX idx_projects_parent_project_id ON projects(parent_project_id);

-- Add foreign key constraints for roles
ALTER TABLE roles ADD CONSTRAINT fk_roles_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
ALTER TABLE roles ADD CONSTRAINT fk_roles_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
CREATE INDEX idx_roles_workspace_id ON roles(workspace_id);
CREATE INDEX idx_roles_project_id ON roles(project_id); 