-- This file reverses the baseline migration.

-- Remove added columns from organizations table
ALTER TABLE organizations DROP CONSTRAINT IF EXISTS fk_organizations_plan;
ALTER TABLE organizations DROP COLUMN IF EXISTS status;
ALTER TABLE organizations DROP COLUMN IF EXISTS plan_id;
ALTER TABLE organizations DROP COLUMN IF EXISTS display_name;

DROP TABLE IF EXISTS auth_states;
DROP TABLE IF EXISTS cicd_credentials;
DROP TABLE IF EXISTS pipeline_runs;
DROP TABLE IF EXISTS pipelines;
DROP TABLE IF EXISTS pipeline_templates;
DROP TABLE IF EXISTS node_events;
DROP TABLE IF EXISTS dedicated_nodes;
DROP TABLE IF EXISTS workspace_node_allocations;
DROP TABLE IF EXISTS node_plans;
DROP TABLE IF EXISTS stripe_events;
DROP TABLE IF EXISTS vcluster_provisioning_tasks;
DROP TABLE IF EXISTS role_assignments;
DROP TABLE IF EXISTS group_memberships;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS organization_users;