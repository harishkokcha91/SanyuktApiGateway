-- 0005_add_approval_audit_fields.up.sql
-- Add approval audit fields to all content tables

-- ============================
-- PROFILES table
-- ============================
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS approved_by BIGINT;
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS rejected_by BIGINT;
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ;

-- Foreign keys
ALTER TABLE profiles 
    ADD CONSTRAINT fk_profiles_approved_by 
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE profiles 
    ADD CONSTRAINT fk_profiles_rejected_by 
    FOREIGN KEY (rejected_by) REFERENCES users(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_profiles_approved_by ON profiles(approved_by);
CREATE INDEX IF NOT EXISTS idx_profiles_rejected_by ON profiles(rejected_by);

-- ============================
-- BUSINESSES table
-- ============================
ALTER TABLE businesses ADD COLUMN IF NOT EXISTS approved_by BIGINT;
ALTER TABLE businesses ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
ALTER TABLE businesses ADD COLUMN IF NOT EXISTS rejected_by BIGINT;
ALTER TABLE businesses ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ;

-- Foreign keys
ALTER TABLE businesses 
    ADD CONSTRAINT fk_businesses_approved_by 
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE businesses 
    ADD CONSTRAINT fk_businesses_rejected_by 
    FOREIGN KEY (rejected_by) REFERENCES users(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_businesses_approved_by ON businesses(approved_by);
CREATE INDEX IF NOT EXISTS idx_businesses_rejected_by ON businesses(rejected_by);

-- ============================
-- EVENTS table
-- ============================
ALTER TABLE events ADD COLUMN IF NOT EXISTS approved_by BIGINT;
ALTER TABLE events ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
ALTER TABLE events ADD COLUMN IF NOT EXISTS rejected_by BIGINT;
ALTER TABLE events ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ;

-- Foreign keys
ALTER TABLE events 
    ADD CONSTRAINT fk_events_approved_by 
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE events 
    ADD CONSTRAINT fk_events_rejected_by 
    FOREIGN KEY (rejected_by) REFERENCES users(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_events_approved_by ON events(approved_by);
CREATE INDEX IF NOT EXISTS idx_events_rejected_by ON events(rejected_by);

-- ============================
-- ACHIEVEMENTS table
-- ============================
ALTER TABLE achievements ADD COLUMN IF NOT EXISTS approved_by BIGINT;
ALTER TABLE achievements ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
ALTER TABLE achievements ADD COLUMN IF NOT EXISTS rejected_by BIGINT;
ALTER TABLE achievements ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ;

-- Foreign keys
ALTER TABLE achievements 
    ADD CONSTRAINT fk_achievements_approved_by 
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE achievements 
    ADD CONSTRAINT fk_achievements_rejected_by 
    FOREIGN KEY (rejected_by) REFERENCES users(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_achievements_approved_by ON achievements(approved_by);
CREATE INDEX IF NOT EXISTS idx_achievements_rejected_by ON achievements(rejected_by);