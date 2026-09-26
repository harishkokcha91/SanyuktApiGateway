-- 0003_add_status_to_events.down.sql
-- Revert status column changes to events table

-- Drop the new check constraint
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_status_check;

-- Revert to original check constraint (if column existed before)
ALTER TABLE events ADD CONSTRAINT events_status_check 
    CHECK (status IN ('upcoming', 'ongoing', 'completed', 'cancelled'));

-- Remove default
ALTER TABLE events ALTER COLUMN status DROP DEFAULT;

-- If column was added by this migration, drop it (commented out for safety)
-- ALTER TABLE events DROP COLUMN IF EXISTS status;