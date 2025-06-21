-- Revert JSONB columns back to text (not recommended)
-- This is just for completeness, you probably don't want to run this
ALTER TABLE workspaces 
  ALTER COLUMN cluster_info TYPE text USING cluster_info::text,
  ALTER COLUMN settings TYPE text USING settings::text,
  ALTER COLUMN metadata TYPE text USING metadata::text;

DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'workspace_statuses') THEN
    ALTER TABLE workspace_statuses 
      ALTER COLUMN cluster_info TYPE text USING cluster_info::text;
  END IF;
END $$;

DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tasks') THEN
    ALTER TABLE tasks 
      ALTER COLUMN payload TYPE text USING payload::text,
      ALTER COLUMN metadata TYPE text USING metadata::text;
  END IF;
END $$; 

DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'projects') THEN
    ALTER TABLE projects 
      ALTER COLUMN details TYPE text USING details::text,
      ALTER COLUMN metadata TYPE text USING metadata::text;
  END IF;
END $$;

DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'applications') THEN
    ALTER TABLE applications 
      ALTER COLUMN config TYPE text USING config::text,
      ALTER COLUMN metadata TYPE text USING metadata::text;
  END IF;
END $$;

DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'function_invocations') THEN
    ALTER TABLE function_invocations 
      ALTER COLUMN payload TYPE text USING payload::text,
      ALTER COLUMN metadata TYPE text USING metadata::text;
  END IF;
END $$;

DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'backup_executions') THEN
    ALTER TABLE backup_executions 
      ALTER COLUMN details TYPE text USING details::text,
      ALTER COLUMN metadata TYPE text USING metadata::text;
  END IF;
END $$;

DO $$ 
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'backup_restores') THEN
    ALTER TABLE backup_restores 
      ALTER COLUMN details TYPE text USING details::text,
      ALTER COLUMN metadata TYPE text USING metadata::text;
  END IF;
END $$;