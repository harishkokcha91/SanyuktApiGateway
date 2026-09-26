-- 0004_fix_business_status_default.down.sql
-- Revert Business status default back to 'Active' and restore original CHECK constraint

-- Drop the new check constraint
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS businesses_status_check;

-- Revert to original check constraint
ALTER TABLE businesses ADD CONSTRAINT businesses_status_check 
    CHECK (status IN ('Active', 'Inactive', 'Pending'));

-- Remove default 'Pending' and restore 'Active'
ALTER TABLE businesses ALTER COLUMN status DROP DEFAULT;
ALTER TABLE businesses ALTER COLUMN status SET DEFAULT 'Active';

-- If column was modified by this migration only, drop it (commented for safety)
-- ALTER TABLE businesses DROP COLUMN IF EXISTS status;