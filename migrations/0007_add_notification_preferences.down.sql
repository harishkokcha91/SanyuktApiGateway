-- Migration 0007: Add notification_preferences table
-- Created: 2026-09-26

-- DOWN Migration
DROP TRIGGER IF EXISTS trigger_notification_preferences_updated_at ON notification_preferences;
DROP FUNCTION IF EXISTS update_notification_preferences_updated_at();
DROP TABLE IF EXISTS notification_preferences;