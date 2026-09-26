-- Migration 0007: Add notification_preferences table
-- Created: 2026-09-26

-- UP Migration
CREATE TABLE IF NOT EXISTS notification_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(20) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, channel)
);

-- Indexes for common queries
CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_notification_preferences_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_notification_preferences_updated_at ON notification_preferences;
CREATE TRIGGER trigger_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_notification_preferences_updated_at();

-- Insert default preferences for existing users
INSERT INTO notification_preferences (user_id, channel, enabled, created_at, updated_at)
SELECT u.id, 'in-app', TRUE, NOW(), NOW()
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM notification_preferences np WHERE np.user_id = u.id AND np.channel = 'in-app'
);

INSERT INTO notification_preferences (user_id, channel, enabled, created_at, updated_at)
SELECT u.id, 'email', TRUE, NOW(), NOW()
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM notification_preferences np WHERE np.user_id = u.id AND np.channel = 'email'
);

INSERT INTO notification_preferences (user_id, channel, enabled, created_at, updated_at)
SELECT u.id, 'push', TRUE, NOW(), NOW()
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM notification_preferences np WHERE np.user_id = u.id AND np.channel = 'push'
);