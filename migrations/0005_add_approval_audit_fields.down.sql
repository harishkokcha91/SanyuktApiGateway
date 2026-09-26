-- 0005_add_approval_audit_fields.down.sql
-- Remove approval audit fields from all content tables

-- ============================
-- ACHIEVEMENTS table
-- ============================
ALTER TABLE achievements DROP CONSTRAINT IF EXISTS fk_achievements_approved_by;
ALTER TABLE achievements DROP CONSTRAINT IF EXISTS fk_achievements_rejected_by;
DROP INDEX IF EXISTS idx_achievements_approved_by;
DROP INDEX IF EXISTS idx_achievements_rejected_by;
-- ALTER TABLE achievements DROP COLUMN IF EXISTS approved_by;
-- ALTER TABLE achievements DROP COLUMN IF EXISTS approved_at;
-- ALTER TABLE achievements DROP COLUMN IF EXISTS rejected_by;
-- ALTER TABLE achievements DROP COLUMN IF EXISTS rejected_at;

-- ============================
-- EVENTS table
-- ============================
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_approved_by;
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_rejected_by;
DROP INDEX IF EXISTS idx_events_approved_by;
DROP INDEX IF EXISTS idx_events_rejected_by;
-- ALTER TABLE events DROP COLUMN IF EXISTS approved_by;
-- ALTER TABLE events DROP COLUMN IF EXISTS approved_at;
-- ALTER TABLE events DROP COLUMN IF EXISTS rejected_by;
-- ALTER TABLE events DROP COLUMN IF EXISTS rejected_at;

-- ============================
-- BUSINESSES table
-- ============================
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS fk_businesses_approved_by;
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS fk_businesses_rejected_by;
DROP INDEX IF EXISTS idx_businesses_approved_by;
DROP INDEX IF EXISTS idx_businesses_rejected_by;
-- ALTER TABLE businesses DROP COLUMN IF EXISTS approved_by;
-- ALTER TABLE businesses DROP COLUMN IF EXISTS approved_at;
-- ALTER TABLE businesses DROP COLUMN IF EXISTS rejected_by;
-- ALTER TABLE businesses DROP COLUMN IF EXISTS rejected_at;

-- ============================
-- PROFILES table
-- ============================
ALTER TABLE profiles DROP CONSTRAINT IF EXISTS fk_profiles_approved_by;
ALTER TABLE profiles DROP CONSTRAINT IF EXISTS fk_profiles_rejected_by;
DROP INDEX IF EXISTS idx_profiles_approved_by;
DROP INDEX IF EXISTS idx_profiles_rejected_by;
-- ALTER TABLE profiles DROP COLUMN IF EXISTS approved_by;
-- ALTER TABLE profiles DROP COLUMN IF EXISTS approved_at;
-- ALTER TABLE profiles DROP COLUMN IF EXISTS rejected_by;
-- ALTER TABLE profiles DROP COLUMN IF EXISTS rejected_at;