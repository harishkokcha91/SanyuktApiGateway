-- 0004_fix_business_status_default.up.sql
-- Fix Business status default from 'Active' to 'Pending' and add approval workflow values

-- Update existing rows that were created with default 'Active' (unreviewed) to 'Pending'
UPDATE businesses SET status = 'Pending' WHERE status = 'Active';

-- Drop the old check constraint
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS businesses_status_check;

-- Add new check constraint with approval workflow values
ALTER TABLE businesses ADD CONSTRAINT businesses_status_check 
    CHECK (status IN ('Pending', 'Approved', 'Rejected'));

-- Set default value for new rows
ALTER TABLE businesses ALTER COLUMN status SET DEFAULT 'Pending';

-- Make column NOT NULL
ALTER TABLE businesses ALTER COLUMN status SET NOT NULL;

-- Recreate index
DROP INDEX IF EXISTS idx_businesses_status;
CREATE INDEX idx_businesses_status ON businesses(status);