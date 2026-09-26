-- 0002_add_role_to_users.up.sql
-- Add role column to users table (added in Phase 2: Auth & RBAC)

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'));

COMMENT ON COLUMN users.role IS 'User role: user or admin';