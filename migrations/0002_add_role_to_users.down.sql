-- 0002_add_role_to_users.down.sql
-- Rollback: Remove role column from users table

ALTER TABLE users DROP COLUMN IF EXISTS role;