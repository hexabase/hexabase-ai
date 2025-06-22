-- Baseline migration for tables previously managed by gorm.AutoMigrate
-- Note: Basic tables (plans, roles, organizations, users, workspaces, projects, applications) 
-- are created in earlier migrations (001, 000)

-- 5. organization_users
CREATE TABLE organization_users (
    organization_id text NOT NULL,
    user_id text NOT NULL,
    role text NOT NULL DEFAULT 'member',
    joined_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_organization_users_role CHECK ((role = ANY (ARRAY['admin'::text, 'member'::text])))
);
ALTER TABLE organization_users ADD CONSTRAINT organization_users_pkey PRIMARY KEY (organization_id, user_id);
ALTER TABLE organization_users ADD CONSTRAINT fk_organization_users_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE organization_users ADD CONSTRAINT fk_organization_users_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 6. sessions
CREATE TABLE sessions (
    id text NOT NULL,
    user_id text,
    refresh_token text,
    device_id text,
    ip_address text,
    user_agent text,
    expires_at timestamp with time zone,
    created_at timestamp with time zone,
    last_used_at timestamp with time zone,
    revoked boolean DEFAULT false
);
ALTER TABLE sessions ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);
ALTER TABLE sessions ADD CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- 9. groups
CREATE TABLE groups (
    id text NOT NULL,
    workspace_id text NOT NULL,
    name text NOT NULL,
    parent_group_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE groups ADD CONSTRAINT groups_pkey PRIMARY KEY (id);
ALTER TABLE groups ADD CONSTRAINT fk_groups_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
ALTER TABLE groups ADD CONSTRAINT fk_groups_parent_group FOREIGN KEY (parent_group_id) REFERENCES groups(id);
CREATE INDEX idx_groups_workspace_id ON groups(workspace_id);
CREATE INDEX idx_groups_parent_group_id ON groups(parent_group_id);

-- 10. group_memberships
CREATE TABLE group_memberships (
    group_id text NOT NULL,
    user_id text NOT NULL,
    role text,
    joined_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE group_memberships ADD CONSTRAINT group_memberships_pkey PRIMARY KEY (group_id, user_id);
ALTER TABLE group_memberships ADD CONSTRAINT fk_group_memberships_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;
ALTER TABLE group_memberships ADD CONSTRAINT fk_group_memberships_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 11. role_assignments
CREATE TABLE role_assignments (
    id text NOT NULL,
    group_id text NOT NULL,
    role_id text NOT NULL,
    created_at timestamp with time zone
);
ALTER TABLE role_assignments ADD CONSTRAINT role_assignments_pkey PRIMARY KEY (id);
ALTER TABLE role_assignments ADD CONSTRAINT fk_role_assignments_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;
ALTER TABLE role_assignments ADD CONSTRAINT fk_role_assignments_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;
CREATE INDEX idx_role_assignments_group_id ON role_assignments(group_id);
CREATE INDEX idx_role_assignments_role_id ON role_assignments(role_id);

-- 12. v_cluster_provisioning_tasks
CREATE TABLE v_cluster_provisioning_tasks (
    id text NOT NULL,
    workspace_id text NOT NULL,
    task_type text NOT NULL,
    status text NOT NULL DEFAULT 'PENDING',
    payload jsonb,
    error_message text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_v_cluster_provisioning_tasks_status CHECK ((status = ANY (ARRAY['PENDING'::text, 'RUNNING'::text, 'COMPLETED'::text, 'FAILED'::text]))),
    CONSTRAINT chk_v_cluster_provisioning_tasks_task_type CHECK ((task_type = ANY (ARRAY['CREATE'::text, 'DELETE'::text, 'UPDATE_PLAN'::text, 'UPDATE_DEDICATED_NODES'::text, 'SETUP_HNC'::text, 'START'::text, 'STOP'::text, 'UPGRADE'::text, 'BACKUP'::text, 'RESTORE'::text])))
);
ALTER TABLE v_cluster_provisioning_tasks ADD CONSTRAINT v_cluster_provisioning_tasks_pkey PRIMARY KEY (id);
ALTER TABLE v_cluster_provisioning_tasks ADD CONSTRAINT fk_v_cluster_provisioning_tasks_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
CREATE INDEX idx_v_cluster_provisioning_tasks_workspace_id ON v_cluster_provisioning_tasks(workspace_id);
CREATE INDEX idx_v_cluster_provisioning_tasks_status ON v_cluster_provisioning_tasks(status);

-- 13. stripe_events
CREATE TABLE stripe_events (
    event_id text NOT NULL,
    event_type text NOT NULL,
    data jsonb NOT NULL,
    status text NOT NULL DEFAULT 'PENDING',
    received_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    processed_at timestamp with time zone,
    CONSTRAINT chk_stripe_events_status CHECK ((status = ANY (ARRAY['PENDING'::text, 'PROCESSED'::text, 'FAILED'::text])))
);
ALTER TABLE stripe_events ADD CONSTRAINT stripe_events_pkey PRIMARY KEY (event_id);
CREATE INDEX idx_stripe_events_event_type ON stripe_events(event_type);
CREATE INDEX idx_stripe_events_status ON stripe_events(status);

-- 14. node_plans
CREATE TABLE node_plans (
    id text NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    description text,
    price_per_month numeric(10,2) NOT NULL,
    resources jsonb,
    features jsonb,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE node_plans ADD CONSTRAINT node_plans_pkey PRIMARY KEY (id);
CREATE INDEX idx_node_plans_type ON node_plans(type);

-- 15. workspace_node_allocations
CREATE TABLE workspace_node_allocations (
    id text NOT NULL,
    workspace_id text NOT NULL,
    plan_type text NOT NULL,
    shared_quota jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE workspace_node_allocations ADD CONSTRAINT workspace_node_allocations_pkey PRIMARY KEY (id);
ALTER TABLE workspace_node_allocations ADD CONSTRAINT uni_workspace_node_allocations_workspace_id UNIQUE (workspace_id);
ALTER TABLE workspace_node_allocations ADD CONSTRAINT fk_workspace_node_allocations_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;

-- 16. dedicated_nodes
CREATE TABLE dedicated_nodes (
    id text NOT NULL,
    workspace_id text NOT NULL,
    name text NOT NULL,
    status text NOT NULL,
    specification jsonb,
    proxmox_vmid integer,
    proxmox_node text,
    ip_address text,
    k3s_agent_version text,
    ssh_public_key text,
    labels jsonb,
    taints jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE dedicated_nodes ADD CONSTRAINT dedicated_nodes_pkey PRIMARY KEY (id);
ALTER TABLE dedicated_nodes ADD CONSTRAINT uni_dedicated_nodes_proxmox_vmid UNIQUE (proxmox_vmid);
ALTER TABLE dedicated_nodes ADD CONSTRAINT fk_dedicated_nodes_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
CREATE INDEX idx_dedicated_nodes_workspace_id ON dedicated_nodes(workspace_id);
CREATE INDEX idx_dedicated_nodes_status ON dedicated_nodes(status);
CREATE INDEX idx_dedicated_nodes_deleted_at ON dedicated_nodes(deleted_at);

-- 17. node_events
CREATE TABLE node_events (
    id text NOT NULL,
    node_id text NOT NULL,
    workspace_id text NOT NULL,
    type text NOT NULL,
    message text NOT NULL,
    details text,
    timestamp timestamp with time zone NOT NULL
);
ALTER TABLE node_events ADD CONSTRAINT node_events_pkey PRIMARY KEY (id);
ALTER TABLE node_events ADD CONSTRAINT fk_node_events_dedicated_node FOREIGN KEY (node_id) REFERENCES dedicated_nodes(id) ON DELETE CASCADE;
CREATE INDEX idx_node_events_node_id ON node_events(node_id);
CREATE INDEX idx_node_events_workspace_id ON node_events(workspace_id);
CREATE INDEX idx_node_events_timestamp ON node_events(timestamp);

-- 18. pipeline_templates
CREATE TABLE pipeline_templates (
    id text NOT NULL,
    name text NOT NULL,
    description text,
    workspace_id text,
    definition jsonb NOT NULL,
    is_public boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE pipeline_templates ADD CONSTRAINT pipeline_templates_pkey PRIMARY KEY (id);
ALTER TABLE pipeline_templates ADD CONSTRAINT fk_pipeline_templates_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;

-- 19. pipelines
CREATE TABLE pipelines (
    id text NOT NULL,
    project_id text NOT NULL,
    name text NOT NULL,
    template_id text,
    source_type text NOT NULL,
    source_url text,
    source_branch text,
    source_path text,
    definition jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE pipelines ADD CONSTRAINT pipelines_pkey PRIMARY KEY (id);
ALTER TABLE pipelines ADD CONSTRAINT fk_pipelines_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE pipelines ADD CONSTRAINT fk_pipelines_template FOREIGN KEY (template_id) REFERENCES pipeline_templates(id);

-- 20. pipeline_runs
CREATE TABLE pipeline_runs (
    id text NOT NULL,
    pipeline_id text NOT NULL,
    trigger_type text NOT NULL,
    triggered_by text,
    status text NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    logs_url text,
    commit_sha text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE pipeline_runs ADD CONSTRAINT pipeline_runs_pkey PRIMARY KEY (id);
ALTER TABLE pipeline_runs ADD CONSTRAINT fk_pipeline_runs_pipeline FOREIGN KEY (pipeline_id) REFERENCES pipelines(id) ON DELETE CASCADE;

-- 21. cicd_credentials
CREATE TABLE cicd_credentials (
    id text NOT NULL,
    project_id text NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    data jsonb NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_cicd_credentials_type CHECK ((type = ANY (ARRAY['git'::text, 'docker'::text, 'aws'::text])))
);
ALTER TABLE cicd_credentials ADD CONSTRAINT cicd_credentials_pkey PRIMARY KEY (id);
ALTER TABLE cicd_credentials ADD CONSTRAINT fk_cicd_credentials_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE cicd_credentials ADD CONSTRAINT uni_cicd_credentials_project_id_name UNIQUE (project_id, name);

-- 22. auth_states
CREATE TABLE auth_states (
    state text NOT NULL,
    provider text NOT NULL,
    redirect_url text,
    code_verifier text,
    client_ip text,
    user_agent text,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);
ALTER TABLE auth_states ADD CONSTRAINT auth_states_pkey PRIMARY KEY (state);
CREATE INDEX idx_auth_states_expires_at ON auth_states(expires_at);

-- 23. role_bindings
CREATE TABLE role_bindings (
    id text NOT NULL,
    workspace_id text NOT NULL,
    project_id text,
    role_id text NOT NULL,
    subject_type text NOT NULL,
    subject_id text NOT NULL,
    subject_name text NOT NULL,
    k8s_rolebinding_name text,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_role_bindings_subject_type CHECK ((subject_type = ANY (ARRAY['User'::text, 'Group'::text])))
);
ALTER TABLE role_bindings ADD CONSTRAINT role_bindings_pkey PRIMARY KEY (id);
ALTER TABLE role_bindings ADD CONSTRAINT fk_role_bindings_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
ALTER TABLE role_bindings ADD CONSTRAINT fk_role_bindings_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE role_bindings ADD CONSTRAINT fk_role_bindings_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;
CREATE INDEX idx_role_bindings_workspace_id ON role_bindings(workspace_id);
CREATE INDEX idx_role_bindings_project_id ON role_bindings(project_id);
CREATE INDEX idx_role_bindings_role_id ON role_bindings(role_id);
CREATE INDEX idx_role_bindings_subject_id ON role_bindings(subject_id);
