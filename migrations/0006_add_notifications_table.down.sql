-- Migration 0006: Add notifications table (DOWN)
-- Created: 2026-09-26

DROP TRIGGER IF EXISTS trigger_notifications_updated_at ON notifications;
DROP FUNCTION IF EXISTS update_notifications_updated_at();
DROP TABLE IF EXISTS notifications;