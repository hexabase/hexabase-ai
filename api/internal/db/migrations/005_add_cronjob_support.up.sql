-- Add CronJob support to applications table
ALTER TABLE applications 
ADD COLUMN IF NOT EXISTS type VARCHAR(20) DEFAULT 'stateless' CHECK (type IN ('stateless', 'cronjob', 'function')),
ADD COLUMN IF NOT EXISTS cron_schedule VARCHAR(100),
ADD COLUMN IF NOT EXISTS cron_command TEXT[],
ADD COLUMN IF NOT EXISTS cron_args TEXT[],
ADD COLUMN IF NOT EXISTS template_app_id UUID REFERENCES applications(id),
ADD COLUMN IF NOT EXISTS last_execution_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS next_execution_at TIMESTAMP;

-- Create index for template app lookups
CREATE INDEX IF NOT EXISTS idx_applications_template_app_id ON applications(template_app_id);

-- CronJob execution history table
CREATE TABLE IF NOT EXISTS cronjob_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID REFERENCES applications(id) ON DELETE CASCADE,
    job_name VARCHAR(255) NOT NULL,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status VARCHAR(20) CHECK (status IN ('running', 'succeeded', 'failed')),
    exit_code INT,
    logs TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_cronjob_executions_app_id ON cronjob_executions(application_id);
CREATE INDEX IF NOT EXISTS idx_cronjob_executions_started_at ON cronjob_executions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_cronjob_executions_status ON cronjob_executions(status);

-- Function for updating next_execution_at based on cron schedule
CREATE OR REPLACE FUNCTION update_next_execution_at()
RETURNS TRIGGER AS $$
BEGIN
    -- This is a placeholder - actual cron parsing would be done in application code
    -- For now, just set it to NULL when schedule changes
    IF NEW.cron_schedule IS DISTINCT FROM OLD.cron_schedule THEN
        NEW.next_execution_at = NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update next_execution_at when cron_schedule changes
CREATE TRIGGER update_next_execution_at_trigger
BEFORE UPDATE ON applications
FOR EACH ROW
WHEN (NEW.type = 'cronjob')
EXECUTE FUNCTION update_next_execution_at();