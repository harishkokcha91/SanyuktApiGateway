-- 0003_add_status_to_events.up.sql
-- Add status column with default 'Pending' and new allowed values to events table

-- First, add the column if it doesn't exist (safe for existing deployments)
ALTER TABLE events ADD COLUMN IF NOT EXISTS status TEXT;

-- Update existing rows to have 'Pending' status
UPDATE events SET status = 'Pending' WHERE status IS NULL OR status = '';

-- Drop the old check constraint if it exists
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_status_check;

-- Add new check constraint with approval workflow values
ALTER TABLE events ADD CONSTRAINT events_status_check 
    CHECK (status IN ('Pending', 'Upcoming', 'Completed', 'Cancelled'));

-- Set default value for new rows
ALTER TABLE events ALTER COLUMN status SET DEFAULT 'Pending';

-- Make column NOT NULL
ALTER TABLE events ALTER COLUMN status SET NOT NULL;

-- Recreate index
DROP INDEX IF EXISTS idx_events_status;
CREATE INDEX idx_events_status ON events(status);