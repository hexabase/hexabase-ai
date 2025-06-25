-- Create base tables required by other migrations
-- Foreign key constraints will be added in later migrations

-- 1. plans
CREATE TABLE plans (
    id text NOT NULL,
    name text NOT NULL,
    description text,
    price numeric NOT NULL,
    currency character varying(3) NOT NULL,
    stripe_price_id text NOT NULL,
    resource_limits jsonb,
    allows_dedicated_nodes boolean DEFAULT false,
    default_dedicated_node_config jsonb,
    max_projects_per_workspace integer,
    max_members_per_workspace integer,
    is_active boolean DEFAULT true,
    display_order integer DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_plans_currency CHECK ((length((currency)::text) = 3))
);
ALTER TABLE plans ADD CONSTRAINT plans_pkey PRIMARY KEY (id);
ALTER TABLE plans ADD CONSTRAINT uni_plans_stripe_price_id UNIQUE (stripe_price_id);

-- 2. roles
CREATE TABLE roles (
    id text NOT NULL,
    workspace_id text,
    project_id text,
    name text NOT NULL,
    description text,
    rules jsonb NOT NULL,
    scope text NOT NULL DEFAULT 'namespace',
    k8s_role_name text,
    is_custom boolean DEFAULT true,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_roles_scope CHECK ((scope = ANY (ARRAY['namespace'::text, 'cluster'::text])))
);
ALTER TABLE roles ADD CONSTRAINT roles_pkey PRIMARY KEY (id);

-- 3. organizations
CREATE TABLE organizations (
    id text NOT NULL,
    name text NOT NULL,
    stripe_customer_id text,
    stripe_subscription_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE organizations ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);
ALTER TABLE organizations ADD CONSTRAINT uni_organizations_name UNIQUE (name);
ALTER TABLE organizations ADD CONSTRAINT uni_organizations_stripe_customer_id UNIQUE (stripe_customer_id);
ALTER TABLE organizations ADD CONSTRAINT uni_organizations_stripe_subscription_id UNIQUE (stripe_subscription_id);

-- 4. users
CREATE TABLE users (
    id text NOT NULL,
    external_id text NOT NULL,
    provider text NOT NULL,
    email text NOT NULL,
    display_name text,
    avatar_url text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    last_login_at timestamp with time zone
);
ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE users ADD CONSTRAINT uni_users_external_id_provider UNIQUE (external_id, provider);
ALTER TABLE users ADD CONSTRAINT uni_users_email UNIQUE (email);

-- 5. workspaces
CREATE TABLE workspaces (
    id text NOT NULL,
    organization_id text NOT NULL,
    name text NOT NULL,
    plan_id text NOT NULL,
    vcluster_instance_name text,
    vcluster_status text DEFAULT 'PENDING_CREATION' NOT NULL,
    vcluster_config jsonb,
    dedicated_node_config jsonb,
    stripe_subscription_item_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_workspaces_vcluster_status CHECK ((vcluster_status = ANY (ARRAY['PENDING_CREATION'::text, 'CONFIGURING_HNC'::text, 'RUNNING'::text, 'UPDATING_PLAN'::text, 'UPDATING_NODES'::text, 'DELETING'::text, 'ERROR'::text, 'UNKNOWN'::text, 'STOPPED'::text, 'STARTING'::text, 'STOPPING'::text])))
);
ALTER TABLE workspaces ADD CONSTRAINT workspaces_pkey PRIMARY KEY (id);
ALTER TABLE workspaces ADD CONSTRAINT uni_workspaces_vcluster_instance_name UNIQUE (vcluster_instance_name);
ALTER TABLE workspaces ADD CONSTRAINT uni_workspaces_stripe_subscription_item_id UNIQUE (stripe_subscription_item_id);

-- 6. projects
CREATE TABLE projects (
    id text NOT NULL,
    workspace_id text NOT NULL,
    name text NOT NULL,
    description text,
    parent_project_id text,
    hnc_anchor_name text,
    namespace_status text DEFAULT 'PENDING_CREATION' NOT NULL,
    kubernetes_namespace text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_projects_namespace_status CHECK ((namespace_status = ANY (ARRAY['PENDING_CREATION'::text, 'ACTIVE'::text, 'DELETING'::text, 'ERROR'::text])))
);
ALTER TABLE projects ADD CONSTRAINT projects_pkey PRIMARY KEY (id);
ALTER TABLE projects ADD CONSTRAINT uni_projects_name_workspace_id UNIQUE (name, workspace_id); 