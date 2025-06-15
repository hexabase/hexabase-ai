-- Fix JSONB columns in workspaces table
ALTER TABLE workspaces 
  ALTER COLUMN cluster_info TYPE jsonb USING cluster_info::jsonb,
  ALTER COLUMN settings TYPE jsonb USING settings::jsonb,
  ALTER COLUMN metadata TYPE jsonb USING metadata::jsonb;

-- Fix JSONB columns in workspace_statuses table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'workspace_statuses') THEN
    ALTER TABLE workspace_statuses 
      ALTER COLUMN cluster_info TYPE jsonb USING cluster_info::jsonb;
  END IF;
END $$;

-- Fix JSONB columns in tasks table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tasks') THEN
    ALTER TABLE tasks 
      ALTER COLUMN payload TYPE jsonb USING payload::jsonb,
      ALTER COLUMN metadata TYPE jsonb USING metadata::jsonb;
  END IF;
END $$;

-- Fix JSONB columns in projects table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'projects') THEN
    ALTER TABLE projects 
      ALTER COLUMN settings TYPE jsonb USING settings::jsonb;
  END IF;
END $$;

-- Fix JSONB columns in applications table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'applications') THEN
    ALTER TABLE applications 
      ALTER COLUMN function_trigger_config TYPE jsonb USING function_trigger_config::jsonb;
  END IF;
END $$;

-- Fix JSONB columns in function_invocations table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'function_invocations') THEN
    ALTER TABLE function_invocations 
      ALTER COLUMN event_data TYPE jsonb USING event_data::jsonb;
  END IF;
END $$;

-- Fix JSONB columns in backup_executions table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'backup_executions') THEN
    ALTER TABLE backup_executions 
      ALTER COLUMN backup_manifest TYPE jsonb USING backup_manifest::jsonb,
      ALTER COLUMN metadata TYPE jsonb USING metadata::jsonb;
  END IF;
END $$;

-- Fix JSONB columns in backup_restores table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'backup_restores') THEN
    ALTER TABLE backup_restores 
      ALTER COLUMN restore_options TYPE jsonb USING restore_options::jsonb,
      ALTER COLUMN validation_results TYPE jsonb USING validation_results::jsonb;
  END IF;
END $$; 