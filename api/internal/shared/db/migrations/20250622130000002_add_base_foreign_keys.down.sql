-- Drop foreign key constraints from base tables

ALTER TABLE roles DROP CONSTRAINT IF EXISTS fk_roles_workspace;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS fk_roles_project;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS fk_projects_workspace;
ALTER TABLE workspaces DROP CONSTRAINT IF EXISTS fk_workspaces_plan;
ALTER TABLE workspaces DROP CONSTRAINT IF EXISTS fk_workspaces_organization; 