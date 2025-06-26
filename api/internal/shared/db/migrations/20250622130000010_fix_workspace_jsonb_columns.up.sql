-- Fix JSONB columns only for existing columns and tables

-- Fix JSONB columns in workspace_statuses table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'workspace_statuses') THEN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'workspace_statuses' AND column_name = 'cluster_info') THEN
      ALTER TABLE workspace_statuses 
        ALTER COLUMN cluster_info TYPE jsonb USING cluster_info::jsonb;
    END IF;
  END IF;
END $$;

-- Fix JSONB columns in tasks table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tasks') THEN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tasks' AND column_name = 'payload') THEN
      ALTER TABLE tasks 
        ALTER COLUMN payload TYPE jsonb USING payload::jsonb;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tasks' AND column_name = 'metadata') THEN
      ALTER TABLE tasks 
        ALTER COLUMN metadata TYPE jsonb USING metadata::jsonb;
    END IF;
  END IF;
END $$;

-- Fix JSONB columns in projects table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'projects') THEN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'projects' AND column_name = 'settings') THEN
      ALTER TABLE projects 
        ALTER COLUMN settings TYPE jsonb USING settings::jsonb;
    END IF;
  END IF;
END $$;

-- Fix JSONB columns in applications table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'applications') THEN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'function_trigger_config') THEN
      ALTER TABLE applications 
        ALTER COLUMN function_trigger_config TYPE jsonb USING function_trigger_config::jsonb;
    END IF;
  END IF;
END $$;

-- Fix JSONB columns in function_invocations table if it exists
-- Note: function_invocations table created by 009_create_functions_tables has JSONB columns already
-- No changes needed for this table

-- Fix JSONB columns in backup_executions table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'backup_executions') THEN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'backup_executions' AND column_name = 'backup_manifest') THEN
      ALTER TABLE backup_executions 
        ALTER COLUMN backup_manifest TYPE jsonb USING backup_manifest::jsonb;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'backup_executions' AND column_name = 'metadata') THEN
      ALTER TABLE backup_executions 
        ALTER COLUMN metadata TYPE jsonb USING metadata::jsonb;
    END IF;
  END IF;
END $$;

-- Fix JSONB columns in backup_restores table if it exists
DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'backup_restores') THEN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'backup_restores' AND column_name = 'restore_options') THEN
      ALTER TABLE backup_restores 
        ALTER COLUMN restore_options TYPE jsonb USING restore_options::jsonb;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'backup_restores' AND column_name = 'validation_results') THEN
      ALTER TABLE backup_restores 
        ALTER COLUMN validation_results TYPE jsonb USING validation_results::jsonb;
    END IF;
  END IF;
END $$; 